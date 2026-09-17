package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// t696: agents consuming `moai todo list --json` and `moai model profile --json`
// guessed a top-level ARRAY and failed with a jq type error (exit 5). Both
// commands emit a single top-level OBJECT whose arrays live under a named key
// (`items` / `agents`). These guards pin the DOCUMENTED shape to the ACTUAL
// output on both doc surfaces (the deployed template copy and the local
// dogfood copy): if a producer changes its top-level shape without updating
// the doc examples — or an example drifts away from what the command emits —
// the mismatch goes red here instead of at some agent's jq invocation.

// docJSONFenceKeys returns the top-level keys of the first ```json fenced
// block in the file at repoRoot/rel that parses as a JSON object (the file may
// carry earlier non-object or non-JSON fences). A missing fence is an error:
// the shape contract these guards pin is expressed through a fenced example.
func docJSONFenceKeys(t *testing.T, repoRoot, rel string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	s := string(data)
	for start := strings.Index(s, "```json"); start >= 0; {
		body := s[start+len("```json"):]
		end := strings.Index(body, "```")
		if end < 0 {
			t.Fatalf("%s: unterminated ```json example", rel)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(body[:end]), &obj); err == nil {
			keys := make(map[string]bool, len(obj))
			for k := range obj {
				keys[k] = true
			}
			return keys
		}
		s = body[end+3:]
		start = strings.Index(s, "```json")
	}
	t.Fatalf("%s carries no ```json object example — the shape contract is undocumented", rel)
	return nil
}

// assertDocKeysCovered fails when any key the doc example names is absent from
// the actual output — a documented field that does not exist is the drift the
// guard exists to catch.
func assertDocKeysCovered(t *testing.T, docRel string, docKeys map[string]bool, actual map[string]json.RawMessage) {
	t.Helper()
	for k := range docKeys {
		if _, ok := actual[k]; !ok {
			t.Errorf("%s documents top-level key %q but the actual output has no such key — doc/output shape drift", docRel, k)
		}
	}
}

func docSurfaces(t *testing.T, repoRoot string, localRel, templateRel string) []string {
	t.Helper()
	paths := []string{filepath.Join(repoRoot, localRel), filepath.Join(repoRoot, templateRel)}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("doc surface missing: %v", err)
		}
	}
	return []string{localRel, templateRel}
}

// TestTodoListJSONShapeMatchesDoc pins `moai todo list --json` to the object
// shape documented in todo.md (both the deployed template copy and the local
// dogfood copy): an object, never an array, and every key the example names
// exists in the actual output.
func TestTodoListJSONShapeMatchesDoc(t *testing.T) {
	todoFixture(t) // isolate the queue root; content is irrelevant to the shape
	out, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("todo list --json: %v", err)
	}
	var actual map[string]json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &actual); err != nil {
		t.Fatalf("output is not a JSON object (consumers piping it to jq would fail): %v", err)
	}
	// The array the consumer wants is the `items` KEY, not the top level.
	for _, want := range []string{"version", "last_seq", "items", "findings"} {
		if _, ok := actual[want]; !ok {
			t.Errorf("actual output missing documented key %q; top-level keys: %v", want, topLevelKeysOf(actual))
		}
	}

	root := repoRootForTest(t)
	for _, rel := range docSurfaces(t, root,
		".claude/skills/moai/workflows/gtd.md",
		"internal/template/templates/.claude/skills/moai/workflows/gtd.md") {
		assertDocKeysCovered(t, rel, docJSONFenceKeys(t, root, rel), actual)
	}
}

// TestModelProfileJSONShapeMatchesDoc pins `moai model profile --json` to the
// object shape documented in model-policy.md: `{profile, backend, agents}` —
// the per-agent cells under the `agents` array, never a top-level array. The
// marshalled report is exactly what runModelProfile emits through
// json.NewEncoder on the --json path, so asserting the marshalled shape
// asserts the emitted bytes.
func TestModelProfileJSONShapeMatchesDoc(t *testing.T) {
	rpt := resolveModelProfileReport(config.LLMConfig{Profile: "high"})
	data, err := json.Marshal(rpt)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var actual map[string]json.RawMessage
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("report is not a JSON object: %v", err)
	}
	for _, want := range []string{"profile", "backend", "agents"} {
		if _, ok := actual[want]; !ok {
			t.Errorf("actual output missing documented key %q; top-level keys: %v", want, topLevelKeysOf(actual))
		}
	}

	root := repoRootForTest(t)
	for _, rel := range docSurfaces(t, root,
		".claude/rules/moai/development/model-policy.md",
		"internal/template/templates/.claude/rules/moai/development/model-policy.md") {
		assertDocKeysCovered(t, rel, docJSONFenceKeys(t, root, rel), actual)
	}
}

func topLevelKeysOf(m map[string]json.RawMessage) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

// repoRootForTest resolves the repository root the way the internal/cli test
// suite does (package dir two levels below the root).
func repoRootForTest(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Join(cwd, "..", "..")
}
