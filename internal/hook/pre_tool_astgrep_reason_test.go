package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// TestGateRun_AstGrepSkipReasonSurfaces drives the ast-grep skip reason across
// the two frames that carry it — RunAstGrepGateV2 → QualityGate.Run — and
// asserts it survives to Run's pass-path output.
//
// Repairing astgrep_gate.go alone leaves the reason inert: Run discarded the
// string on the pass path. This test is what makes the propagation, not merely
// the final assignment, load-bearing.
//
// The consumer of that output is `moai gate` (internal/cli/gate.go runGate),
// which prints a non-empty pass-path output to stderr so the git pre-commit
// hook surfaces it. There used to be a third frame — preToolHandler.Handle,
// which ran the whole gate inline on a git commit and forwarded the notice on
// HookOutput.SystemMessage. That frame is gone: the gate no longer runs inside
// the PreToolUse hook's 10s budget (see pre_tool_commit_gate_test.go for why),
// so the assertions below stop at Run. Nothing about the propagation itself
// changed; only the surface that consumed it did.
//
// Project fixture, gate by gate:
//   - build.zig selects the Zig toolchain, whose entry has no vet or lint
//     steps, so Run reaches the ast-grep step immediately.
//   - no ast-grep-ignore anywhere, so the suppression sweep cannot deny first.
//   - SkipTests, so the test step after the ast-grep step never runs.
//   - t.Setenv PATH strips sg, which is the condition under test.
//
// Non-parallel: t.Setenv is incompatible with t.Parallel().
func TestGateRun_AstGrepSkipReasonSurfaces(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "build.zig"), []byte("// zig build marker\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "main.ts"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	// Constructed explicitly rather than inherited from ambient project config,
	// so the test's outcome does not depend on the host's gate.yaml.
	gcfg := quality.DefaultGateConfig()
	gcfg.Enabled = true
	gcfg.SkipTests = true
	gcfg.ProjectDir = projectDir
	gcfg.AstGrepGate = &quality.AstGrepGateConfig{
		Enabled:  true,
		RulesDir: ".moai/config/astgrep-rules",
	}

	passed, got := quality.NewQualityGate(gcfg).Run(context.Background())

	// 1. an absent optional scanner must never block.
	if !passed {
		t.Fatalf("the gate failed when sg is merely absent; an optional scanner's absence is a skip, not a failure. output: %s", got)
	}

	// 2. the reason reaches Run's output at all.
	if got == "" {
		t.Fatal("Run's output is empty: the ast-grep skip reason was dropped between the gate step and Run")
	}

	// 3. the string that arrived is the one the gate step produced — the same
	//    reason traversed both frames rather than being re-invented at Run.
	//    Recomputed against the same fixture under the same stripped PATH, so
	//    the comparison is against the live gate, not a literal.
	_, want := quality.RunAstGrepGateV2(context.Background(), projectDir, &quality.AstGrepGateConfig{
		Enabled:  true,
		RulesDir: ".moai/config/astgrep-rules",
	})
	if want == "" {
		t.Fatal("the gate step produced an empty reason; the end-to-end comparison has nothing to compare against")
	}
	//    Containment rather than equality: Run concatenates the notices of every
	//    step that skipped, so the ast-grep reason arrives alongside others
	//    (the typecheck step's, for a language with no default command).
	if !strings.Contains(got, want) {
		t.Errorf("Run output does not carry the gate step's reason.\nwant substring: %q\ngot: %q", want, got)
	}
}
