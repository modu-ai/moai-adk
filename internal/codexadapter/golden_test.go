package codexadapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// goldenDirRel is the tracked payload directory.
//
// It is deliberately NOT .moai/reports/t91/hook-payloads/, where these dumps
// were originally captured: that directory is untracked and exists only in the
// primary checkout, so a test reading it passes on one machine and fails in a
// worktree and in CI.
const goldenDirRel = ".moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads"

// goldenPayload is the field set common to every captured Codex hook payload.
type goldenPayload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
}

func goldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), goldenDirRel, name)
}

func readGolden(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(goldenPath(t, name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	return raw
}

// TestGoldenPayloadsParse — AC-REQ-6.
//
// The six adapted events each parse from their captured payload, and the
// hook_event_name each carries round-trips through the event table. This is
// what backs the SPEC's claim that internal/hook's parsing layer is reusable:
// the field names are snake_case-identical across the two harnesses.
func TestGoldenPayloadsParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		file  string
		event hook.EventType
	}{
		{"PreToolUse.json", hook.EventPreToolUse},
		{"PostToolUse.json", hook.EventPostToolUse},
		{"SessionStart.json", hook.EventSessionStart},
		{"SessionEnd.json", hook.EventSessionEnd},
		{"Stop.json", hook.EventStop},
		{"UserPromptSubmit.json", hook.EventUserPromptSubmit},
	}

	const wantCovered = 6
	if len(cases) != wantCovered {
		t.Fatalf("golden coverage = %d events, want %d", len(cases), wantCovered)
	}

	for _, tc := range cases {
		t.Run(string(tc.event), func(t *testing.T) {
			t.Parallel()

			var p goldenPayload
			if err := json.Unmarshal(readGolden(t, tc.file), &p); err != nil {
				t.Fatalf("unmarshal %s: %v", tc.file, err)
			}

			if p.HookEventName != string(tc.event) {
				t.Errorf("hook_event_name = %q, want %q", p.HookEventName, tc.event)
			}
			for name, got := range map[string]string{
				"session_id":      p.SessionID,
				"transcript_path": p.TranscriptPath,
				"cwd":             p.CWD,
			} {
				if got == "" {
					t.Errorf("%s is empty in %s", name, tc.file)
				}
			}

			arg, err := Resolve(p.HookEventName)
			if err != nil {
				t.Fatalf("Resolve(%s) error = %v — a golden event must be adapted", p.HookEventName, err)
			}
			if arg == "" {
				t.Error("resolved dispatcher arg is empty")
			}
		})
	}
}

// TestGoldenToolNameNormalized records the observation that Codex reports its
// own shell tool under Claude Code's name, which is why MoAI's pre-tool
// pattern guards match without any renaming.
func TestGoldenToolNameNormalized(t *testing.T) {
	t.Parallel()

	var p struct {
		ToolName string `json:"tool_name"`
	}
	if err := json.Unmarshal(readGolden(t, "PreToolUse.json"), &p); err != nil {
		t.Fatalf("unmarshal PreToolUse.json: %v", err)
	}
	if p.ToolName != "Bash" {
		t.Fatalf("tool_name = %q, want Bash", p.ToolName)
	}
}

// TestGoldenStopCarriesStopHookActive pins the Stop-specific field the
// adapter's callers rely on to avoid re-blocking their own continuation.
func TestGoldenStopCarriesStopHookActive(t *testing.T) {
	t.Parallel()

	var p map[string]any
	if err := json.Unmarshal(readGolden(t, "Stop.json"), &p); err != nil {
		t.Fatalf("unmarshal Stop.json: %v", err)
	}
	if _, ok := p["stop_hook_active"]; !ok {
		t.Fatal("Stop payload has no stop_hook_active field")
	}
}

// TestGoldenSessionEndOmitsModelFields records a real asymmetry: SessionEnd
// carries neither model nor permission_mode, so a parser requiring them on
// every event would break on this one.
func TestGoldenSessionEndOmitsModelFields(t *testing.T) {
	t.Parallel()

	var p map[string]any
	if err := json.Unmarshal(readGolden(t, "SessionEnd.json"), &p); err != nil {
		t.Fatalf("unmarshal SessionEnd.json: %v", err)
	}
	for _, absent := range []string{"model", "permission_mode"} {
		if _, ok := p[absent]; ok {
			t.Errorf("SessionEnd carries %q; the capture shows it absent", absent)
		}
	}
}

// TestGoldensAreTracked — AC-REQ-6.
//
// The point of vendoring was reachability from a branch and in CI, so the test
// asserts the files resolve from the module root rather than from a developer's
// primary checkout.
func TestGoldensAreTracked(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(repoRoot(t), goldenDirRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("golden directory unreachable at %s: %v", goldenDirRel, err)
	}

	const wantMin = 6
	count := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			count++
		}
	}
	if count < wantMin {
		t.Fatalf("golden payloads = %d, want at least %d", count, wantMin)
	}
}
