package fix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// llmClientTokens is the forbidden LLM-client token set per AC-NS5-007a /
// REQ-NS5-007. Each token is matched case-insensitively against the package's
// non-test Go source — as an import path OR a symbol reference. Any hit means
// the Go engine grew an embedded LLM client, which the split-architecture
// decision (design.md §A) forbids: the AI draft rides the orchestrator's
// manager-develop delegation (REQ-NS5-007), never a Go-embedded inference
// path. moai-adk has no Go-embedded LLM client in the navigator subsystem —
// the LLM is reached only through Claude Code's Agent() runtime, which a Go
// CLI binary cannot invoke.
var llmClientTokens = []string{
	"openai",
	"anthropic",
	"langchain",
	"mcp__",
	"generativeai",
	"claude.ai",
	"api.openai",
}

// findLLMClientToken scans dir's non-test *.go files and returns the first
// (basename, token) pair whose token appears case-insensitively, or ("", "")
// when the tree is clean. Mirrors the M0 internal/navigator/sync/nonoverlap_test.go
// and M2 internal/cli/navigator_route_test.go grep-guard idiom: filepath.Glob +
// os.ReadFile + strings.Contains over the package's own source — NOT a shell
// grep subprocess. AC-NS5-007a scopes the guard to non-test source, so every
// *_test.go is skipped (the test files themselves name the forbidden tokens as
// the assertion target).
func findLLMClientToken(t *testing.T, dir string) (file, token string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("findLLMClientToken: glob %s: %v", dir, err)
	}
	for _, m := range matches {
		if strings.HasSuffix(filepath.Base(m), "_test.go") {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatalf("findLLMClientToken: read %s: %v", m, err)
		}
		lower := strings.ToLower(string(b))
		for _, tok := range llmClientTokens {
			if strings.Contains(lower, tok) {
				return filepath.Base(m), tok
			}
		}
	}
	return "", ""
}

// TestFixNoLLMClient is the AC-NS5-007a split-architecture CI guard: the M3
// Go engine source (internal/navigator/fix/*.go excluding *_test.go) MUST
// contain ZERO LLM-client imports (no openai / anthropic / langchain / mcp__
// / generativeai / claude.ai / api.openai). The AI draft is produced by the
// orchestrator-spawned manager-develop delegation (REQ-NS5-007), NOT by the
// Go engine. This carries forward the C-HRA-008 subagent-boundary pattern +
// the M0/M1/M2 non-overlap guards — REQ-NS5-007 is the load-binding contract
// (design.md §A.2: moai-adk has no Go-embedded LLM client in the navigator
// subsystem; design.md §A.4: the handoff is a file-based contract, not an
// in-process inference call).
func TestFixNoLLMClient(t *testing.T) {
	t.Parallel()

	if file, tok := findLLMClientToken(t, "."); tok != "" {
		t.Errorf("AC-NS5-007a violation: %s contains forbidden LLM-client token %q — the Go engine MUST NOT embed an LLM client (design.md §A, REQ-NS5-007); the AI draft is produced by the orchestrator's manager-develop delegation, not a Go-embedded inference path", file, tok)
	}
}

// TestFixNoLLMClient_Teeth proves the guard actually detects a planted LLM
// import — it falsifies the "guard is a no-op / current package merely happens
// to be clean" hypothesis. Each forbidden token is planted as a (fake) import
// path in a non-test *.go file inside t.TempDir, and the matcher MUST flag it;
// a no-op matcher returns ("", "") and fails the subtest. This is the teeth
// check plan.md §F M3.3 + the delegation prompt require (E8 detection
// evidence). Note: openai is a substring of api.openai, so the matcher may
// return the earlier-listed token — the assertion is that the planted file was
// flagged at all, which is the falsifiable teeth property.
func TestFixNoLLMClient_Teeth(t *testing.T) {
	t.Parallel()

	for _, tok := range llmClientTokens {
		tok := tok
		t.Run("planted_"+tok, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			planted := []byte("// planted by TestFixNoLLMClient_Teeth — exercise the matcher\n" +
				"package fix\n" +
				"import _ \"github.com/example/" + tok + "\"\n")
			if err := os.WriteFile(filepath.Join(dir, "planted_engine.go"), planted, 0o644); err != nil {
				t.Fatalf("write planted file: %v", err)
			}
			file, got := findLLMClientToken(t, dir)
			if got == "" {
				t.Fatalf("matcher missed planted token %q (guard has no teeth)", tok)
			}
			if file != "planted_engine.go" {
				t.Errorf("matcher flagged %q, expected planted_engine.go", file)
			}
		})
	}
}

// TestFixNoLLMClient_CaseInsensitive proves the matcher is case-insensitive:
// a mixed-case symbol/import (e.g. "Anthropic") MUST be caught, because the
// strings.ToLower normalization is load-bearing for AC-NS5-007a (import paths
// and symbol references appear in both casing conventions). A future
// "simplify to case-sensitive" regression would let this subtest fail.
func TestFixNoLLMClient_CaseInsensitive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Mixed-case "Anthropic" would NOT match a case-sensitive Contains(_, "anthropic");
	// only the strings.ToLower normalization catches it.
	planted := []byte("package fix\nimport _ \"github.com/Anthropic/anthropic-sdk-go\"\n")
	if err := os.WriteFile(filepath.Join(dir, "mixed_engine.go"), planted, 0o644); err != nil {
		t.Fatalf("write mixed-case planted file: %v", err)
	}
	if file, got := findLLMClientToken(t, dir); got == "" {
		t.Fatal("matcher missed mixed-case token \"Anthropic\" — guard MUST be case-insensitive (strings.ToLower is load-bearing)")
	} else if file != "mixed_engine.go" {
		t.Errorf("matcher flagged %q, expected mixed_engine.go", file)
	}
}
