package hook

// AC-AE-001 (SPEC-AUTONOMY-ESCALATION-001, REQ-AE-001): under every
// non-contract value of workflow.autonomy.mode the escalation detector is
// inert — the PreToolUse and PostToolUse outputs for a fixed fixture set stay
// byte-identical to a golden captured BEFORE any detector code existed, and
// no escalation directory appears anywhere.
//
// The golden file is regenerated only deliberately:
//
//	MOAI_ESCALATION_GOLDEN_UPDATE=1 go test ./internal/hook -run '^TestEscalationGuidedGolden$'
//
// Regenerating it after detector code lands defeats the criterion; the file's
// history is the evidence that it predates the detector.

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

const escalationGoldenPath = "testdata/escalation_guided_golden.jsonl"

// staticConfigProvider hands a loaded configuration to a handler.
type staticConfigProvider struct{ cfg *config.Config }

func (p staticConfigProvider) Get() *config.Config { return p.cfg }

// escalationGoldenEvent is one fixture hook invocation.
type escalationGoldenEvent struct {
	name  string
	event EventType
	input HookInput
}

func escalationGoldenEvents(root string) []escalationGoldenEvent {
	cwd := filepath.Join(root, "internal")
	raw := func(v any) json.RawMessage {
		b, _ := json.Marshal(v)
		return b
	}
	base := func(tool string, in any) HookInput {
		return HookInput{
			SessionID:      "golden-session",
			CWD:            cwd,
			PermissionMode: "default",
			ToolName:       tool,
			ToolInput:      raw(in),
		}
	}
	write := base("Write", map[string]string{"file_path": filepath.Join(root, "internal", "bar", "x.go"), "content": "package bar\n"})
	edit := base("Edit", map[string]string{"file_path": filepath.Join(root, "internal", "fixture", "y.go"), "old_string": "a", "new_string": "b"})
	outside := base("Write", map[string]string{"file_path": filepath.Join(filepath.Dir(root), "outside", "z.txt"), "content": "z\n"})
	bashLs := base("Bash", map[string]string{"command": "ls -la"})
	bashPush := base("Bash", map[string]string{"command": "git push origin main"})
	bashTag := base("Bash", map[string]string{"command": "git tag v1.0.0"})
	bashDeny := base("Bash", map[string]string{"command": "rm -rf /"})
	bashTest := base("Bash", map[string]string{"command": "go test ./internal/spec/..."})

	postWrite := write
	postWrite.ToolResponse = raw(map[string]any{"success": true})
	postEdit := edit
	postEdit.ToolResponse = raw(map[string]any{"success": true})
	postBash := bashTest
	postBash.ToolResponse = raw(map[string]any{"stdout": "FAIL\n", "stderr": "", "exit_code": 1})
	postLs := bashLs
	postLs.ToolResponse = raw(map[string]any{"stdout": "a\nb\n", "stderr": "", "exit_code": 0})

	return []escalationGoldenEvent{
		{"pre-write-outside-ownership", EventPreToolUse, write},
		{"pre-edit-inside-ownership", EventPreToolUse, edit},
		{"pre-write-outside-root", EventPreToolUse, outside},
		{"pre-bash-ls", EventPreToolUse, bashLs},
		{"pre-bash-push-main", EventPreToolUse, bashPush},
		{"pre-bash-tag", EventPreToolUse, bashTag},
		{"pre-bash-denylisted", EventPreToolUse, bashDeny},
		{"post-write", EventPostToolUse, postWrite},
		{"post-edit", EventPostToolUse, postEdit},
		{"post-bash-failing-test", EventPostToolUse, postBash},
		{"post-bash-ls", EventPostToolUse, postLs},
	}
}

// escalationGoldenRun processes every fixture event under one mode fixture
// and returns one JSON line per event: `{"name":..., "output":...}`.
func escalationGoldenRun(t *testing.T, mode string, modeAbsent bool) []string {
	t.Helper()
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteMode(mode, modeAbsent)
	// A signed, in-progress contract claiming this card: under a non-contract
	// mode it must still arm nothing.
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
	w.Write("internal/fixture/y.go", "package fixture\n")
	t.Setenv(config.EnvClaudeProjectDir, w.Root)

	cfg, err := config.NewLoader().Load(w.Path(".moai"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	provider := staticConfigProvider{cfg: cfg}
	pre := NewPreToolHandler(provider, DefaultSecurityPolicy())
	post := NewPostToolHandlerWithConfig(nil, nil, "", 0, provider)

	var lines []string
	for _, ev := range escalationGoldenEvents(w.Root) {
		in := ev.input
		var out *HookOutput
		switch ev.event {
		case EventPreToolUse:
			out, err = pre.Handle(context.Background(), &in)
		default:
			out, err = post.Handle(context.Background(), &in)
		}
		if err != nil {
			t.Fatalf("%s: Handle: %v", ev.name, err)
		}
		b, err := json.Marshal(struct {
			Name   string      `json:"name"`
			Output *HookOutput `json:"output"`
		}{ev.name, out})
		if err != nil {
			t.Fatalf("%s: marshal: %v", ev.name, err)
		}
		lines = append(lines, string(b))
	}
	assertNoEscalationDir(t, w.Root)
	assertNoEscalationDir(t, os.Getenv(config.EnvHome))
	return lines
}

// assertNoEscalationDir fails when a directory named "escalation" exists
// anywhere under root.
func assertNoEscalationDir(t *testing.T, root string) {
	t.Helper()
	if root == "" {
		return
	}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "escalation" {
			t.Errorf("escalation directory present: %s", p)
		}
		return nil
	})
}

func TestEscalationGuidedGolden(t *testing.T) {
	cases := []struct {
		name       string
		mode       string
		modeAbsent bool
	}{
		{"mode-absent", "", true},
		{"mode-guided", "guided", false},
		{"mode-empty", "", false},
		{"mode-bogus", "bogus", false},
	}

	if os.Getenv("MOAI_ESCALATION_GOLDEN_UPDATE") == "1" {
		lines := escalationGoldenRun(t, "", true)
		if err := os.MkdirAll(filepath.Dir(escalationGoldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(escalationGoldenPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("golden written: %s (%d events)", escalationGoldenPath, len(lines))
	}

	golden, err := os.ReadFile(escalationGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	want := strings.Split(strings.TrimSuffix(string(golden), "\n"), "\n")
	if len(want) != len(escalationGoldenEvents("")) {
		t.Fatalf("golden holds %d lines, fixture set has %d events", len(want), len(escalationGoldenEvents("")))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := escalationGoldenRun(t, tc.mode, tc.modeAbsent)
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("event %d differs from golden\n got: %s\nwant: %s", i, got[i], want[i])
				}
			}
		})
	}
}
