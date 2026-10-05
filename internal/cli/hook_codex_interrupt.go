package cli

// hook_codex_interrupt.go — SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e
// (REQ-HPR-012, design.md §D7). The Codex-only `moai hook interrupt`
// subcommand records a user cancellation. It has no Claude-side counterpart
// and nothing under internal/hook changes (Q6).

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// codexInterruptDir holds one JSONL cancellation log per session.
const codexInterruptDir = ".moai/state/codex-interrupt"

// interruptRecord is one recorded user cancellation. It binds the session and
// the goal run that was live when the interrupt arrived.
type interruptRecord struct {
	SessionID        string      `json:"session_id"`
	TranscriptPath   string      `json:"transcript_path,omitempty"`
	RecordedAt       string      `json:"recorded_at"`
	GoalStatusBefore goal.Status `json:"goal_status_before,omitempty"`
	GoalCreatedAt    string      `json:"goal_created_at,omitempty"`
}

func interruptRecordPath(root, sessionID string) string {
	return filepath.Join(root, codexInterruptDir, sessionID+".jsonl")
}

// readInterruptRecords returns the cancellation records for one session.
func readInterruptRecords(root, sessionID string) ([]interruptRecord, error) {
	f, err := os.Open(interruptRecordPath(root, sessionID))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open interrupt records: %w", err)
	}
	defer func() { _ = f.Close() }()
	var out []interruptRecord
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r interruptRecord
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, fmt.Errorf("parse interrupt record: %w", err)
		}
		out = append(out, r)
	}
	return out, sc.Err()
}

// runHookInterrupt is the RunE of `moai hook interrupt`. It records the
// cancellation first and then marks this session's armed goal `cancelled`, so
// the next Stop allows and the loop does not resume (REQ-HPR-015). It writes
// the no-opinion object: Interrupt is not decision-bearing.
func runHookInterrupt(cmd *cobra.Command, _ []string) error {
	if deps == nil || deps.HookProtocol == nil {
		return fmt.Errorf("hook system not initialized")
	}
	isCodex, err := harnessModeIsCodex(cmd)
	if err != nil {
		return err
	}
	if !isCodex {
		return fmt.Errorf("moai hook %s: the Interrupt event exists only on Codex; run it with --harness codex (there is no Claude-side counterpart)",
			codexadapter.CodexInterruptDispatcherArg)
	}
	input, err := deps.HookProtocol.ReadInput(os.Stdin)
	if err != nil {
		return fmt.Errorf("moai hook %s: read payload: %w", codexadapter.CodexInterruptDispatcherArg, err)
	}
	if input.HookEventName == "" {
		input.HookEventName = string(codexadapter.CodexEventInterrupt)
	}
	if err := validateCodexHarnessEvent(codexadapter.CodexEventInterrupt, input); err != nil {
		return err
	}
	if input.SessionID == "" {
		// Without a session the record cannot be bound, and goal state is
		// keyed by session; a guessed key could cancel another run.
		return fmt.Errorf("moai hook %s: payload carries no session_id; nothing recorded", codexadapter.CodexInterruptDispatcherArg)
	}
	if filepath.Base(input.SessionID) != input.SessionID || input.SessionID == ".." {
		// The session id names a file under .moai/state; a separator or a
		// parent reference would write outside it.
		return fmt.Errorf("moai hook %s: session_id %q is not a plain name; nothing recorded", codexadapter.CodexInterruptDispatcherArg, input.SessionID)
	}
	if err := recordCodexInterrupt(resolveCodexStopRoot(input), input, time.Now()); err != nil {
		return fmt.Errorf("moai hook %s: %w", codexadapter.CodexInterruptDispatcherArg, err)
	}
	return writeHookOutputCodex(codexadapter.CodexEventInterrupt, &hook.HookOutput{})
}

// @MX:NOTE: [AUTO] the Codex Interrupt producer of goal status `cancelled` (design.md §D6); only an armed goal is cancelled — a terminal status is recorded, never overwritten
// @MX:SPEC: SPEC-DUAL-HARNESS-HOOK-PARITY-001
// recordCodexInterrupt appends one cancellation record for the session and, if
// that session's goal is armed, sets it to goal.StatusCancelled.
func recordCodexInterrupt(root string, input *hook.HookInput, now time.Time) error {
	g, err := goal.LoadGoal(root, input.SessionID)
	if err != nil {
		return fmt.Errorf("load goal: %w", err)
	}
	rec := interruptRecord{
		SessionID:      input.SessionID,
		TranscriptPath: input.TranscriptPath,
		RecordedAt:     now.UTC().Format(time.RFC3339Nano),
	}
	if g != nil {
		rec.GoalStatusBefore = g.Status
		rec.GoalCreatedAt = g.CreatedAt
	}
	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal interrupt record: %w", err)
	}
	path := interruptRecordPath(root, input.SessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create interrupt record dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open interrupt record: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return fmt.Errorf("write interrupt record: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close interrupt record: %w", err)
	}
	if g == nil || g.Status != goal.StatusArmed {
		return nil
	}
	g.Status = goal.StatusCancelled
	if err := goal.SaveGoal(root, g); err != nil {
		return fmt.Errorf("save cancelled goal: %w", err)
	}
	return nil
}
