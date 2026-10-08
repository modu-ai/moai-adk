package outbox

// modelcalls.go — the persisted model-call budget (review-gate finding 3,
// P2): the rolling daily cap on summary calls lives in the user-scoped
// store, not in a Sender instance. The in-process budget reset with every
// new flush — N flushes could spend N x the daily cap. The judgment and the
// record are ONE queue-locked mutation, so two concurrent flushes cannot
// both pass the cap (counted atomically cross-process, like every ledger
// write).

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// ModelCallsFileName is the persisted budget file under
// <moai home>/state/bugreport/.
const ModelCallsFileName = "modelcalls.json"

// modelCallLog is the on-disk spend: the attempt timestamps inside the
// rolling 24-hour window (older entries are pruned on each judgment).
type modelCallLog struct {
	At []string `json:"at"`
}

// loadModelCalls reads the budget file; an absent or malformed file reads as
// empty spend — failing closed here would suppress every summary forever
// until a manual purge, and the cap is a spend bound, not a safety gate.
// The read is BOUNDED (review gate finding, P2): the judgment runs INSIDE
// the queue-lock mutation, so a non-regular budget file is refused without
// opening and the open+read runs under DefaultBugreportModelCallsReadTimeBox
// with the config.DefaultBugreportModelCallsMaxBytes size cap — a FIFO swapped in
// at the budget path used to park the read while HOLDING the queue lock,
// stalling every other queue operation. On deadline the helper goroutine is
// left parked on the blocked handle; it exits when the blocking writer
// closes, and the caller never waits for it.
func loadModelCalls() *modelCallLog {
	path, err := StorePath(ModelCallsFileName)
	if err != nil {
		return &modelCallLog{}
	}
	if info, serr := os.Stat(path); serr == nil && !info.Mode().IsRegular() {
		return &modelCallLog{}
	}
	type readResult struct {
		raw []byte
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		f, err := os.Open(path)
		if err != nil {
			done <- readResult{}
			return
		}
		defer func() { _ = f.Close() }()
		raw, err := io.ReadAll(io.LimitReader(f, config.DefaultBugreportModelCallsMaxBytes+1))
		done <- readResult{raw: raw, err: err}
	}()
	var raw []byte
	select {
	case r := <-done:
		if r.err != nil || len(r.raw) == 0 || len(r.raw) > config.DefaultBugreportModelCallsMaxBytes {
			return &modelCallLog{}
		}
		raw = r.raw
	case <-time.After(config.DefaultBugreportModelCallsReadTimeBox):
		return &modelCallLog{}
	}
	var l modelCallLog
	if json.Unmarshal(raw, &l) != nil {
		return &modelCallLog{}
	}
	return &l
}

// saveModelCalls persists the budget atomically (temp + rename in the same
// directory).
func saveModelCalls(l *modelCallLog) error {
	path, err := StorePath(ModelCallsFileName)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".modelcalls-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(encoded, '\n')); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(outboxFilePerm); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// AllowAndRecordModelCall judges the rolling 24-hour cap and records the
// attempt in one queue-locked mutation. A lock or save failure refuses the
// call (false) — an uncounted call is an unbounded one, and the deterministic
// template covers the refused path. The caller's context bounds the lock
// wait.
func AllowAndRecordModelCall(ctx context.Context, now time.Time) bool {
	allowed := false
	store := BugreportQueueStore()
	_ = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		l := loadModelCalls()
		stamp := now.UTC().Format(time.RFC3339)
		kept := l.At[:0]
		for _, ts := range l.At {
			at, err := time.Parse(time.RFC3339, ts)
			if err != nil {
				continue
			}
			if now.Sub(at) < 24*time.Hour {
				kept = append(kept, ts)
			}
		}
		l.At = kept
		if len(l.At) >= config.DefaultBugreportModelCallsPerDay {
			_ = saveModelCalls(l) // persist the pruning even on a refusal
			return nil
		}
		l.At = append(l.At, stamp)
		if err := saveModelCalls(l); err != nil {
			return nil // the record did not land: refuse, the template covers it
		}
		allowed = true
		return nil
	})
	return allowed
}
