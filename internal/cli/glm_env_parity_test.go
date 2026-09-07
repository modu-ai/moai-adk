package cli

// glm_env_parity_test.go — SPEC-CLIFIX-HYGIENE-001 REQ-HYG-001-003 (M3)
//
// Asserts the GLM env-var inject↔clear parity invariant: every GLM env-var
// NAME lives exactly once in internal/config/envkeys.go, and no bare
// CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS / CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC
// / CLAUDE_CODE_TEAMMATE_DISPLAY literal survives in the production source
// under internal/cli/. Test files (_test.go) are exempt (CLAUDE.local.md §14
// hardcoding-allowed zone), so this scan restricts to non-test .go files.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMEnvSetParity asserts the helper returns the canonical 3-key set in
// the documented order. Both inject and clear paths derive from this set, so
// pinning its contents pins the contract.
func TestGLMEnvSetParity(t *testing.T) {
	t.Parallel()

	got := config.GLMEnvVarSet()
	want := []string{
		config.EnvClaudeCodeDisableExperimentalBetas,
		config.EnvClaudeCodeDisableNonessentialTraffic,
		config.EnvClaudeCodeTeammateDisplay,
	}
	if len(got) != len(want) {
		t.Fatalf("GLMEnvVarSet returned %d keys, want %d", len(got), len(want))
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("GLMEnvVarSet()[%d] = %q, want %q", i, g, want[i])
		}
	}

	// The set MUST be exactly the 3 canonical keys — no extras, no missing.
	wantSet := map[string]bool{
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":   true,
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": true,
		"CLAUDE_CODE_TEAMMATE_DISPLAY":             true,
	}
	for _, g := range got {
		if !wantSet[g] {
			t.Errorf("GLMEnvVarSet returned unexpected key %q", g)
		}
	}
}

// TestNoBareGLMEnvVarLiteralsInCLIProduction scans every non-test .go file
// under internal/cli for bare GLM env-var string literals and fails if any
// survive. After M3, every reference MUST go through the config constants.
// This is the mechanical gate that prevents the inject↔clear drift from
// silently re-entering the codebase.
//
// Excludes: _test.go files (allowed hardcoding zone per CLAUDE.local.md §14).
func TestNoBareGLMEnvVarLiteralsInCLIProduction(t *testing.T) {
	t.Parallel()

	banned := []string{
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC",
		"CLAUDE_CODE_TEAMMATE_DISPLAY",
	}
	bannedSet := make(map[string]bool, len(banned))
	for _, b := range banned {
		bannedSet[b] = true
	}

	cliDir := "."
	wd, err := os.Getwd()
	if err == nil {
		cliDir = wd
	}

	var offenders []string
	_ = filepath.WalkDir(cliDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		node, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return nil // skip unparseable files; not the concern of this test
		}
		ast.Inspect(node, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			val := strings.Trim(lit.Value, `"`)
			if bannedSet[val] {
				pos := fset.Position(lit.Pos())
				offenders = append(offenders, pos.String())
			}
			return true
		})
		return nil
	})

	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Fatalf("bare GLM env-var string literal(s) found in production CLI source "+
			"(must reference config.EnvClaudeCode* constants instead):\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
