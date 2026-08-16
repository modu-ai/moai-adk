package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// newKanbanProjectDir builds the minimal on-disk shape the SessionStart handler
// expects, so these cases exercise the handler rather than a mkdir failure.
func newKanbanProjectDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return dir
}

// TestSessionStartKanbanNoticeReachesOperator is the regression case for the
// defect this dual emission exists to prevent: the lead notice carries three
// launch commands the OPERATOR must type by hand into new terminals, but it was
// emitted only on hookSpecificOutput.additionalContext — a model-facing channel.
// The operator's terminal stayed empty, so the run appeared not to start at all.
//
// systemMessage is the operator-facing channel ("Warning message shown to the
// user"), and SessionStart does not discard it. Both channels must carry the
// notice: the orchestrator needs the labels to address companions later, and the
// operator needs the launch lines to open those sessions in the first place.
func TestSessionStartKanbanNoticeReachesOperator(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "tjpyre")

	projectDir := newKanbanProjectDir(t)
	out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
		SessionID:  "uuid-kanban-surface-001",
		CWD:        projectDir,
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

	// The operator-facing channel is the one the defect emptied.
	if out.SystemMessage == "" {
		t.Fatal("SystemMessage is empty — the kanban notice never reaches the operator's terminal")
	}
	for _, want := range []string{
		"Kanban Mode: run tjpyre",
		"moai cc -k --name plan",
		"moai cc -k --name run",
		"moai cc -k --name sync",
	} {
		if !strings.Contains(out.SystemMessage, want) {
			t.Errorf("SystemMessage missing %q.\nGot: %q", want, out.SystemMessage)
		}
	}

	// The model-facing channel must keep carrying it too — the orchestrator
	// addresses the companions by these labels.
	if out.HookSpecificOutput == nil || !strings.Contains(out.HookSpecificOutput.AdditionalContext, "Kanban Mode: run tjpyre") {
		t.Error("AdditionalContext lost the kanban notice — the orchestrator can no longer address companions")
	}
}

// TestSessionStartKanbanCompanionNoticeReachesOperator covers the companion
// branch, which carries a single join line rather than the launch block.
func TestSessionStartKanbanCompanionNoticeReachesOperator(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "plan")

	projectDir := newKanbanProjectDir(t)
	out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
		SessionID:  "uuid-kanban-surface-002",
		CWD:        projectDir,
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.SystemMessage, "Kanban Mode: joined the kanban run as plan") {
		t.Errorf("companion join notice missing from SystemMessage.\nGot: %q", out.SystemMessage)
	}
}

// TestSessionStartNonKanbanSystemMessageUnaffected is the blast-radius case: an
// ordinary session must gain no operator-facing message from this path.
func TestSessionStartNonKanbanSystemMessageUnaffected(t *testing.T) {
	clearKanbanEnv(t)

	projectDir := newKanbanProjectDir(t)
	out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
		SessionID:  "uuid-kanban-surface-003",
		CWD:        projectDir,
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if strings.Contains(out.SystemMessage, "Kanban Mode") {
		t.Errorf("non-kanban session got a kanban SystemMessage: %q", out.SystemMessage)
	}
}

// TestSessionStartKanbanNoticeOnlyOnNewSession pins the bootstrap notice to a
// genuinely new session.
//
// SessionStart fires on five documented sources, and the kanban environment
// survives all of them, so an ungated notice re-announced the bootstrap every
// time a lead session came back: the operator was told to open three companion
// terminals that were already open, for a run already under way. The
// instruction is only actionable at startup.
//
// The table carries all five, plus an undocumented value. The last row is the
// one that matters for the future: `fork` did not exist before Claude Code
// v2.1.214 (a forked session reported `resume`), so a denylist of the four
// older sources silently resumed emitting when the fifth appeared. Asserting an
// unknown value is suppressed pins the allowlist shape, not just today's set.
//
// The empty source is asserted to still emit. Claude Code always populates the
// field, so an empty value means a caller that predates it — including the other
// cases in this file, which construct HookInput without one.
func TestSessionStartKanbanNoticeOnlyOnNewSession(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{
		{source: "startup", want: true},
		{source: "", want: true},
		{source: "resume", want: false},
		{source: "clear", want: false},
		{source: "compact", want: false},
		{source: "fork", want: false},
		{source: "some-future-source", want: false},
	} {
		name := tc.source
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "tjpyre")

			projectDir := newKanbanProjectDir(t)
			out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
				SessionID:  "uuid-kanban-source-" + name,
				CWD:        projectDir,
				ProjectDir: projectDir,
				Source:     tc.source,
			})
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			got := strings.Contains(out.SystemMessage, "Kanban Mode")
			if got != tc.want {
				t.Errorf("source %q: notice emitted = %v, want %v.\nSystemMessage: %q",
					tc.source, got, tc.want, out.SystemMessage)
			}

			// The model-facing channel is gated by the same predicate; a
			// suppressed notice must not survive on either surface.
			ac := ""
			if out.HookSpecificOutput != nil {
				ac = out.HookSpecificOutput.AdditionalContext
			}
			if gotAC := strings.Contains(ac, "Kanban Mode"); gotAC != tc.want {
				t.Errorf("source %q: AdditionalContext carried notice = %v, want %v",
					tc.source, gotAC, tc.want)
			}
		})
	}
}
