// Package receipt is the moai-owned contract store: two append-only,
// hash-chained JSON Lines files under $MOAI_HOME/db/<project-key>/contract/.
//
//	receipts.jsonl  kickoff receipts moai issued (`moai contract decide`)
//	events.jsonl    every signing event — sign-human, sign-receipt, reseal,
//	                decide, revoke
//
// Each line carries the hash of the line before it, so an edited middle line
// breaks the chain and Verify names the line. The chain leaves a trace; it
// does not prevent tampering: an actor with write access can rewrite the
// whole file, and deleting the last line is invisible to the chain itself
// (the consumers cover that case — a deleted signing event fails the Kickoff
// check as unrecorded, a deleted revoke event leaves the revoke escalation
// record as a second witness).
//
// The store lives outside every worktree and is shared by all worktrees of
// the repository through the project key (StoreDir), so an agent
// editing its tree cannot reach it with an ordinary write. Tests point
// MOAI_HOME at t.TempDir() so they never touch the real home.
package receipt

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// Store file names.
const (
	ReceiptsFile = "receipts.jsonl"
	EventsFile   = "events.jsonl"
	lockName     = "store.lock"
)

// Line kinds. KindReceipt is the only kind in receipts.jsonl; the other five
// are events.jsonl kinds.
const (
	KindReceipt     = "receipt"
	KindSignHuman   = "sign-human"
	KindSignReceipt = "sign-receipt"
	KindReseal      = "reseal"
	KindDecide      = "decide"
	KindRevoke      = "revoke"
)

// Line is one chained record. Data is kept as the exact bytes written, so the
// hash recomputes byte for byte.
type Line struct {
	Prev string          `json:"prev"`
	At   string          `json:"at"`
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
	Hash string          `json:"hash"`
}

// ChainError reports the first broken line of a store file (1-based).
type ChainError struct {
	File string
	Line int
	Why  string
}

func (e *ChainError) Error() string {
	return fmt.Sprintf("contract store integrity: %s line %d: %s", e.File, e.Line, e.Why)
}

// ErrIntegrity is wrapped by every chain failure.
var ErrIntegrity = errors.New("contract store: integrity check failed")

func (e *ChainError) Unwrap() error { return ErrIntegrity }

// Store is one project's contract store.
type Store struct {
	// Dir is the store directory.
	Dir string
	// Now is the clock stamped on appended lines. Default: time.Now.
	Now func() time.Time
}

// Open returns the store of the project the worktree belongs to:
// $MOAI_HOME/db/<project-key>/contract, the directory the escalation
// detector's arming state already uses. The directory is not created until
// the first append.
func Open(worktreeRoot string) (*Store, error) {
	dir, err := StoreDir(worktreeRoot)
	if err != nil {
		return nil, err
	}
	return &Store{Dir: dir}, nil
}

func (s *Store) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// hashLine is the chain hash of one line's content.
func hashLine(prev, at, kind string, data []byte) string {
	h := sha256.New()
	for _, part := range [][]byte{[]byte(prev), []byte(at), []byte(kind), data} {
		h.Write(part)
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// readChain reads and verifies one store file. An absent file is an empty,
// valid chain.
func (s *Store) readChain(name string) ([]Line, error) {
	f, err := os.Open(filepath.Join(s.Dir, name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("contract store: open %s: %w", name, err)
	}
	defer func() { _ = f.Close() }()
	var (
		lines []Line
		prev  string
		n     int
	)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		n++
		raw := sc.Bytes()
		if len(bytes.TrimSpace(raw)) == 0 {
			return nil, &ChainError{File: name, Line: n, Why: "empty line"}
		}
		var l Line
		if err := json.Unmarshal(raw, &l); err != nil {
			return nil, &ChainError{File: name, Line: n, Why: "not a JSON line: " + err.Error()}
		}
		if l.Prev != prev {
			return nil, &ChainError{File: name, Line: n, Why: "prev does not match the preceding line"}
		}
		if hashLine(l.Prev, l.At, l.Kind, l.Data) != l.Hash {
			return nil, &ChainError{File: name, Line: n, Why: "hash does not match the line content"}
		}
		prev = l.Hash
		lines = append(lines, l)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("contract store: read %s: %w", name, err)
	}
	return lines, nil
}

// Verify checks both chains. It returns a *ChainError (wrapping ErrIntegrity)
// for a broken chain, another error for an I/O failure, nil otherwise.
func (s *Store) Verify() error {
	for _, name := range []string{ReceiptsFile, EventsFile} {
		if _, err := s.readChain(name); err != nil {
			return err
		}
	}
	return nil
}

// Events returns the verified events.jsonl lines.
func (s *Store) Events() ([]Line, error) { return s.readChain(EventsFile) }

// Receipts returns the verified receipts.jsonl lines.
func (s *Store) Receipts() ([]Line, error) { return s.readChain(ReceiptsFile) }

// AppendEvent appends one event line. kind must be an events.jsonl kind.
func (s *Store) AppendEvent(kind string, data any) (Line, error) {
	switch kind {
	case KindSignHuman, KindSignReceipt, KindReseal, KindDecide, KindRevoke:
	default:
		return Line{}, fmt.Errorf("contract store: unknown event kind %q", kind)
	}
	return s.append(EventsFile, kind, data)
}

// AppendReceipt appends one issued receipt line.
func (s *Store) AppendReceipt(data ReceiptData) (Line, error) {
	return s.append(ReceiptsFile, KindReceipt, data)
}

// append adds one line under the store lock, after verifying the chain it
// extends. Nothing is written when the chain is already broken.
func (s *Store) append(name, kind string, data any) (Line, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return Line{}, fmt.Errorf("contract store: encode %s: %w", kind, err)
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return Line{}, fmt.Errorf("contract store: create %s: %w", s.Dir, err)
	}
	lock, err := os.OpenFile(filepath.Join(s.Dir, lockName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return Line{}, fmt.Errorf("contract store: open lock: %w", err)
	}
	defer func() { _ = lock.Close() }()
	if err := lockfile.Lock(lock); err != nil {
		return Line{}, fmt.Errorf("contract store: lock: %w", err)
	}
	defer func() { _ = lockfile.Unlock(lock) }()

	lines, err := s.readChain(name)
	if err != nil {
		return Line{}, err
	}
	prev := ""
	if len(lines) > 0 {
		prev = lines[len(lines)-1].Hash
	}
	at := s.now().UTC().Format(time.RFC3339Nano)
	l := Line{Prev: prev, At: at, Kind: kind, Data: payload, Hash: hashLine(prev, at, kind, payload)}
	enc, err := json.Marshal(l)
	if err != nil {
		return Line{}, fmt.Errorf("contract store: encode line: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(s.Dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return Line{}, fmt.Errorf("contract store: open %s: %w", name, err)
	}
	if _, err := f.Write(append(enc, '\n')); err != nil {
		_ = f.Close()
		return Line{}, fmt.Errorf("contract store: append %s: %w", name, err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return Line{}, fmt.Errorf("contract store: sync %s: %w", name, err)
	}
	if err := f.Close(); err != nil {
		return Line{}, fmt.Errorf("contract store: close %s: %w", name, err)
	}
	return l, nil
}
