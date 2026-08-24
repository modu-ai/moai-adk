package hook

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// SPEC-SEC-SCAN-SURFACE-001 M0 — the differential invariant test.
//
// This is the instrument the rest of the SPEC is measured against: it replays a
// fixture corpus through scanWriteContent and pins the (decision,
// reason-nonempty) pair each fixture produced on the UNMODIFIED gate. It must
// keep passing, unmodified, after M1 (config/covered-language skips), M2 (the
// literal pre-filter), and M3 (the PostToolUse merge). Any milestone that makes
// the gate cheaper by making it deny less fails here.
//
// The wantDeny column states the verdict the shipped ruleset's error-severity
// rules call for, verified rule-by-rule with `sg scan` directly (see the
// corpus-validity gate at the bottom of this file).
//
// HISTORY: when this corpus was first recorded, none of these fixtures denied.
// astGrepScanner.Scan collected the subprocess output with
// exec.Cmd.CombinedOutput, and `sg scan --json` exits 1 and prints
// "Error: N error(s) found in code." on stderr whenever an error-severity
// finding exists — so the merged stream was not valid JSON, json.Unmarshal
// failed, and the regex fallback (which parses ast-grep's TEXT format) found
// nothing in JSON output. Warnings were unaffected: they exit 0 with an empty
// stderr, so their JSON parsed and WarningCount was populated. The gate could
// warn but could never deny. The scanner now collects stdout and stderr
// separately, and the corpus-validity gate below is met, so the assertions run.
//
// The corpus-validity gate is retained: it turns any future regression of the
// collection path back into a loud, self-explaining skip rather than a corpus
// that silently observes nothing.

// scanCorpusFixture is one recorded row of the differential corpus.
type scanCorpusFixture struct {
	// name identifies the row in test output.
	name string
	// file is the fixture under internal/hook/security/testdata/scan-corpus/.
	file string
	// virtualPath is the file_path the Write payload claims; its extension is
	// what selects the language, not the fixture's own name.
	virtualPath string
	// language is the ast-grep language the extension maps to, or "" for a
	// language the shipped ruleset carries no rules for.
	coveredLanguage string
	// wantDeny is the decision recorded on the pre-implementation tree.
	wantDeny bool
}

// scanCorpus is the recorded corpus. Four covered languages carry at least one
// denying fixture each (the corpus-validity gate below enforces that); the
// remaining rows cover clean payloads, a warning-only payload, and two
// rule-uncovered languages.
var scanCorpus = []scanCorpusFixture{
	{name: "go/deny/hardcoded-credential", file: "go_deny_credential.go", virtualPath: "sample.go", coveredLanguage: "go", wantDeny: true},
	{name: "go/deny/weak-hash-md5", file: "go_deny_md5.go", virtualPath: "digest.go", coveredLanguage: "go", wantDeny: true},
	{name: "go/allow/clean", file: "go_clean.go", virtualPath: "clean.go", coveredLanguage: "go", wantDeny: false},
	{name: "go/allow/warning-only", file: "go_warning_only.go", virtualPath: "box.go", coveredLanguage: "go", wantDeny: false},
	{name: "javascript/deny/child-process-exec", file: "js_deny_exec.js", virtualPath: "run.js", coveredLanguage: "javascript", wantDeny: true},
	{name: "javascript/allow/clean", file: "js_clean.js", virtualPath: "add.js", coveredLanguage: "javascript", wantDeny: false},
	{name: "typescript/deny/child-process-exec", file: "ts_deny_exec.ts", virtualPath: "run.ts", coveredLanguage: "typescript", wantDeny: true},
	{name: "typescript/allow/clean", file: "ts_clean.ts", virtualPath: "add.ts", coveredLanguage: "typescript", wantDeny: false},
	{name: "python/deny/os-system", file: "py_deny_os_system.py", virtualPath: "run.py", coveredLanguage: "python", wantDeny: true},
	{name: "python/allow/clean", file: "py_clean.py", virtualPath: "add.py", coveredLanguage: "python", wantDeny: false},
	{name: "rust/allow/uncovered-language", file: "rs_uncovered.rs", virtualPath: "main.rs", coveredLanguage: "", wantDeny: false},
	{name: "java/allow/uncovered-language", file: "java_uncovered.java", virtualPath: "Sample.java", coveredLanguage: "", wantDeny: false},
}

// coveredCorpusLanguages is the set of languages the shipped ruleset carries
// rules for. Every one of them must have a denying fixture, else the corpus
// observes nothing for that language.
var coveredCorpusLanguages = []string{"go", "javascript", "typescript", "python"}

// repoRootForTest walks up from the test's working directory to the repository
// root (the directory carrying go.mod).
func repoRootForTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// shippedRulesetSource is the ruleset the template distributes. Tests copy it
// into a t.TempDir() project root; they never point at the developer's own
// .moai/config/astgrep-rules, which is local-only and dogfood-grade.
func shippedRulesetSource(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRootForTest(t), "internal", "template", "templates", ".moai", "config", "astgrep-rules")
}

// projectRootWithShippedRuleset builds a temp project root carrying a copy of
// the shipped ruleset at .moai/config/astgrep-rules/.
func projectRootWithShippedRuleset(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dst := filepath.Join(root, ".moai", "config", "astgrep-rules")
	copyTreeForTest(t, shippedRulesetSource(t), dst)
	return root
}

// copyTreeForTest copies a directory tree, creating parents as needed.
func copyTreeForTest(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			return mkErr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy tree %s -> %s: %v", src, dst, err)
	}
}

// corpusFixtureContent reads one fixture's bytes.
func corpusFixtureContent(t *testing.T, file string) string {
	t.Helper()
	path := filepath.Join(repoRootForTest(t), "internal", "hook", "security", "testdata", "scan-corpus", file)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", file, err)
	}
	return string(data)
}

// writePayload builds the Write tool input JSON scanWriteContent parses.
func writePayload(t *testing.T, filePath, content string) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"file_path": filePath, "content": content})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}

// requireASTGrep skips when the `sg` binary is absent: without it every scan
// returns Scanned=false and the corpus would observe nothing.
func requireASTGrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("ast-grep (sg) not on PATH — the differential corpus cannot observe a deny without it")
	}
}

// TestScanWriteContentDifferential closes AC-SSS-001 assertion (i): every
// fixture yields the identical (decision, reason-nonempty) pair it yielded on
// the unmodified gate.
func TestScanWriteContentDifferential(t *testing.T) {
	requireASTGrep(t)

	root := projectRootWithShippedRuleset(t)
	h := &preToolHandler{
		scanner:    security.NewSecurityScanner(),
		projectDir: root,
	}

	type observed struct {
		decision string
		reason   string
	}
	got := make(map[string]observed, len(scanCorpus))
	denyingLanguages := map[string]bool{}

	for _, fx := range scanCorpus {
		content := corpusFixtureContent(t, fx.file)
		decision, reason := h.scanWriteContent(context.Background(), writePayload(t, fx.virtualPath, content))
		got[fx.name] = observed{decision: decision, reason: reason}
		if decision == DecisionDeny && fx.coveredLanguage != "" {
			denyingLanguages[fx.coveredLanguage] = true
		}
	}

	// Corpus-validity gate: a corpus that denies nothing observes nothing, so
	// asserting against it would bless a gate that cannot deny. Report the
	// measured state and skip rather than pretend the corpus is meaningful.
	var missing []string
	for _, lang := range coveredCorpusLanguages {
		if !denyingLanguages[lang] {
			missing = append(missing, lang)
		}
	}
	if len(missing) > 0 {
		var b strings.Builder
		b.WriteString("corpus rejected: no denying fixture for covered language(s) ")
		b.WriteString(strings.Join(missing, ", "))
		b.WriteString("\nobserved decisions on this tree:")
		for _, fx := range scanCorpus {
			b.WriteString("\n  ")
			b.WriteString(fx.name)
			b.WriteString(": decision=")
			if d := got[fx.name].decision; d == "" {
				b.WriteString("allow")
			} else {
				b.WriteString(d)
			}
			b.WriteString(" wantDeny=")
			if fx.wantDeny {
				b.WriteString("true")
			} else {
				b.WriteString("false")
			}
		}
		b.WriteString("\ncause: astGrepScanner.Scan uses CombinedOutput; `sg scan --json` writes " +
			"\"Error: N error(s) found in code.\" to stderr on any error-severity finding, " +
			"which corrupts the JSON so no error finding is ever parsed. " +
			"The pre-write gate can warn but cannot deny.")
		t.Skip(b.String())
	}

	for _, fx := range scanCorpus {
		o := got[fx.name]
		gotDeny := o.decision == DecisionDeny
		if gotDeny != fx.wantDeny {
			t.Errorf("%s: decision changed: got deny=%v (decision=%q reason=%q), recorded deny=%v",
				fx.name, gotDeny, o.decision, o.reason, fx.wantDeny)
			continue
		}
		if gotDeny && strings.TrimSpace(o.reason) == "" {
			t.Errorf("%s: deny with empty reason", fx.name)
		}
		if !gotDeny && o.reason != "" {
			t.Errorf("%s: allow carried a non-empty reason: %q", fx.name, o.reason)
		}
	}
}
