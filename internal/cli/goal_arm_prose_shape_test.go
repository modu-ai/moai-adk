package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// TestProseShapedCommand pins the predicate itself (card t556).
//
// The t436 runnability gate refuses a mechanical condition whose FIRST word
// resolves to no command. That misses the shape this predicate exists for:
// prose whose first word happens to BE a command name — `make`, `test`, `go`
// all resolve, so the gate waves the sentence through and the goal blocks every
// turn-end to the ceiling.
//
// The controls matter as much as the catches: a predicate that also flags real
// commands would refuse legitimate goals at the door.
func TestProseShapedCommand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want bool
	}{
		// --- catches: prose whose first word resolves as a command ---
		{"make-prose", "make sure every AC row is marked PASS", true},
		{"test-prose", "test coverage reaches 85 percent for every package", true},
		{"go-prose", "go through each SPEC and confirm the fix landed", true},
		{"trailing-period", "make sure every AC row is marked PASS.", true},
		{"prose-with-one-path", "Every blocking acceptance criterion in .moai/specs/SPEC-X/spec.md has PASS evidence", true},
		{"non-english", "모든 AC 가 PASS 로 표시 된다 확인", true},

		// --- controls: real commands must NOT be flagged ---
		{"short-cmd", "go test ./...", false},
		{"builtin", "true", false},
		{"two-word", "make build", false},
		{"flags", "gh pr checks --json state --jq .[0].state", false},
		{"chained", "go build ./... && go vet ./... && go test ./...", false},
		{"npm", "npm run test -- --coverage --watchAll=false", false},
		{"compose", "docker compose -f docker-compose.yml up -d --wait", false},
		{"pytest", "pytest tests/unit -q --maxfail=1 --disable-warnings", false},
		{"empty", "", false},
	} {
		if got := proseShapedCommand(tc.in); got != tc.want {
			t.Errorf("proseShapedCommand(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestGoalArm_ProseShapedConditionRefused is the arm-surface half: the CLI must
// refuse, and MUST NOT write a state file. This is the failure the card reports
// — a sentence armed as a shell command that can never exit 0.
func TestGoalArm_ProseShapedConditionRefused(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "make sure every AC row is marked PASS", "--session", "T556PROSE"})
	err := rc.Execute()
	if err == nil {
		t.Fatalf("prose condition was armed (out=%s)", buf.String())
	}
	if !strings.Contains(err.Error(), "model:") {
		t.Errorf("refusal does not name the model: remedy: %v", err)
	}
	if g, loadErr := goal.LoadGoal(root, "T556PROSE"); loadErr == nil && g != nil {
		t.Errorf("a state file was written on refusal: %+v", g)
	}
}

// TestGoalArm_ProseShapeExemptions pins that the escape hatches still arm. A
// refusal the author cannot override would be worse than the defect it closes.
func TestGoalArm_ProseShapeExemptions(t *testing.T) {
	for _, tc := range []struct{ name, condition, session string }{
		{"cmd-prefix", "cmd: make sure every AC row is marked PASS", "T556CMD"},
		{"model-prefix", "model: make sure every AC row is marked PASS", "T556MODEL"},
		{"real-command", "true", "T556REAL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("CLAUDE_PROJECT_DIR", root)

			rc, buf := newGoalTestRoot()
			rc.SetArgs([]string{"goal", "arm", tc.condition, "--session", tc.session})
			if err := rc.Execute(); err != nil {
				t.Fatalf("exempt condition was refused: %v (out=%s)", err, buf.String())
			}
		})
	}
}

// TestMCPGoalArm_ProseShapedConditionRefused pins the MCP wrapper's wiring.
//
// Issue #1660 was reported against the MCP surface specifically, and the two
// arm surfaces silently disagreeing is the defect class. A CLI-only test would
// stay green while the wrapper armed exactly what the CLI refuses, so the gate
// is measured at BOTH doors — and the control (a real command) proves the
// wrapper still arms what it should.
func TestMCPGoalArm_ProseShapedConditionRefused(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	ctx := context.Background()
	c, err := client.NewInProcessClient(newMoaiMCPServer())
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	call := func(condition string) *mcp.CallToolResult {
		t.Helper()
		res, callErr := c.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
			Name:      "goal_arm",
			Arguments: map[string]any{"session_id": "T556MCP", "condition": condition},
		}})
		if callErr != nil {
			t.Fatalf("tools/call goal_arm: %v", callErr)
		}
		return res
	}

	if r := call("make sure every AC row is marked PASS"); !r.IsError {
		t.Errorf("MCP goal_arm armed a prose condition: %v", r.Content)
	}
	if g, loadErr := goal.LoadGoal(tmp, "T556MCP"); loadErr == nil && g != nil {
		t.Errorf("MCP refusal wrote a state file: %+v", g)
	}
	// Control: the wrapper still arms a real command.
	if r := call("go test ./... exits 0"); r.IsError {
		t.Errorf("MCP goal_arm refused a real command: %v", r.Content)
	}
}
