package spec

// lint_vacuous_assertion_test.go — VacuousTestAssertionRule acceptance suite
// (SPEC-SPEC-LINT-VACUOUS-ASSERT-001, card t1269).
//
// Every detection axis carries two arms: a detection arm whose inputs must fire
// and a conformant arm whose inputs must stay silent. A rule that fires on
// nothing and a rule that fires on everything each fail one arm, so the pair is
// what makes the criterion falsifiable (acceptance.md, AC-VTA-013).
//
// The tests live in package spec (not spec_test) because AC-VTA-008 counts the
// entries of the unexported rule slice built by NewLinter.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const vtaCode = "VacuousTestAssertion"

// vtaFrontmatter renders a schema-valid modern-era frontmatter. The era pin is
// load-bearing: without it a fixture with no progress.md classifies as a
// grandfathered era and every warning is demoted to advisory for a reason that
// has nothing to do with the cutoff under test.
func vtaFrontmatter(created, extra string) string {
	createdLine := ""
	if created != "" {
		createdLine = "created: " + created + "\n"
	}
	return "---\n" +
		"id: SPEC-VTAT-001\n" +
		"title: \"Vacuous assertion fixture\"\n" +
		"version: \"0.1.0\"\n" +
		"status: draft\n" +
		createdLine +
		"updated: 2026-09-27\n" +
		"author: Test Author\n" +
		"priority: P2 Medium\n" +
		"phase: \"v3.2.0\"\n" +
		"module: \"internal/spec\"\n" +
		"era: V3R6\n" +
		"dependencies: []\n" +
		"bc_id: []\n" +
		"lifecycle: spec-anchored\n" +
		"tags: \"test\"\n" +
		"breaking: false\n" +
		"related_rule: []\n" +
		extra +
		"---\n"
}

// vtaWriteDir writes the given artifacts into dir and returns a SPECDoc
// pointing at dir/spec.md. spec.md defaults to a bare frontmatter.
func vtaWriteDir(t *testing.T, dir, created string, files map[string]string) *SPECDoc {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, ok := files["spec.md"]; !ok {
		files["spec.md"] = vtaFrontmatter(created, "") + "\n# Fixture\n"
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &SPECDoc{
		Path:        filepath.Join(dir, "spec.md"),
		Frontmatter: SPECFrontmatter{Created: created},
	}
}

// vtaCheckLine runs the rule over a spec.md whose body is the single line.
func vtaCheckLine(t *testing.T, line string) []Finding {
	t.Helper()
	body := vtaFrontmatter(vacuousGateCutoff, "") + "\n" + line + "\n"
	doc := vtaWriteDir(t, t.TempDir(), vacuousGateCutoff, map[string]string{"spec.md": body})
	return (&VacuousTestAssertionRule{}).Check(doc, nil)
}

func vtaAxisCount(findings []Finding, axis string) int {
	n := 0
	for _, f := range findings {
		if f.Code == vtaCode && strings.HasPrefix(f.Message, axis+":") {
			n++
		}
	}
	return n
}

// TestVacuousAssertionRule_RunPatternAxis — AC-VTA-001.
func TestVacuousAssertionRule_RunPatternAxis(t *testing.T) {
	detect := map[string]string{
		"missing trailing dollar":        `go test ./x -run '^TestA' -v`,
		"missing leading caret":          `go test ./x -run 'TestA$' -v`,
		"bare unquoted name":             `go test ./x -run TestA -v`,
		"equals form unanchored":         `go test ./x -run=TestA -v`,
		"equals quoted unanchored":       `go test ./x -run='^TestA' -v`,
		"first branch unanchored":        `go test ./x -run 'TestA|^TestB$' -v`,
		"last branch unanchored":         `go test ./x -run '^TestA$|^TestB' -v`,
		"unanchored subtest level":       `go test ./x -run '^TestA$/sub' -v`,
		"sibling groups top-level pipe":  `go test ./x -run '^(TestA)|(TestB)$' -v`,
		"double quoted unanchored":       `go test ./x -run "^TestA" -v`,
		"escaped dollar in single quote": `go test ./x -run '^TestA\$' -v`,
	}
	for name, line := range detect {
		t.Run("detect/"+name, func(t *testing.T) {
			got := vtaCheckLine(t, line)
			if n := vtaAxisCount(got, "run-pattern"); n != 1 || len(got) != 1 {
				t.Errorf("line %q: run-pattern findings = %d (all = %+v), want exactly 1", line, n, got)
			}
		})
	}

	conform := map[string]string{
		"single anchored name":           `go test ./x -run '^TestA$' -v`,
		"grouped alternation":            `go test ./x -run '^(TestA|TestB)$' -v`,
		"per-branch alternation":         `go test ./x -run '^TestA$|^TestB$' -v`,
		"per-level anchored subtests":    `go test ./x -run '^TestA$/^sub$' -v`,
		"grouped subtest level":          `go test ./x -run '^(TestA)$/^(x|y)$' -v`,
		"empty-match form":               `go test ./x -run '^$' -v`,
		"leading flag group":             `go test ./x -run '(?i)^testa$' -v`,
		"flag group on subtest level":    `go test ./x -run '^TestA$/(?i)^sub$' -v`,
		"double quote escaped dollar":    `go test ./x -run "^TestA\$" -v`,
		"double quote plain dollar":      `go test ./x -run "^TestA$" -v`,
		"equals form anchored":           `go test ./x -run=^TestA$ -v`,
		"inner group alternation":        `go test ./x -run '^TestV_(A|B)$' -v`,
		"class holding pipe and slash":   `go test ./x -run '^Test[|/]A$' -v`,
		"regex-escaped pipe":             `go test ./x -run '^Test\|A$' -v`,
		"no go test token is not judged": `run it with -run 'TestA' later`,
	}
	for name, line := range conform {
		t.Run("conform/"+name, func(t *testing.T) {
			if got := vtaCheckLine(t, line); len(got) != 0 {
				t.Errorf("line %q: findings = %+v, want none", line, got)
			}
		})
	}
}

// TestVacuousAssertionRule_OutcomeAssertionAxis — AC-VTA-002 (with the N1 and
// N3 plan-audit debts).
func TestVacuousAssertionRule_OutcomeAssertionAxis(t *testing.T) {
	prefixes := []string{"--- PASS: ", "--- FAIL: ", "--- (PASS|FAIL): ", "--- (FAIL|PASS): "}
	for _, p := range prefixes {
		endings := map[string]string{
			"single quote": `grep -E -- '` + p + `TestX'`,
			"double quote": `grep -E -- "` + p + `TestX"`,
			"backtick":     "the output contains `" + p + "TestX`",
			"end of line":  "the output contains " + p + "TestX",
			"dollar":       `grep -E -- '` + p + `TestX$'`,
		}
		for name, line := range endings {
			t.Run("detect/"+p+name, func(t *testing.T) {
				got := vtaCheckLine(t, line)
				if n := vtaAxisCount(got, "outcome-assertion"); n != 1 || len(got) != 1 {
					t.Errorf("line %q: outcome findings = %d (all = %+v), want exactly 1", line, n, got)
				}
			})
		}
	}

	detect := map[string]string{
		"word boundary is not a delimiter": `grep -E -- '--- PASS: TestX\b'`,
		"paren directly after name":        `grep -F -- '--- PASS: TestX('`,
		"on a go test line":                `go test ./x -run '^TestX$' -v | grep -- '--- PASS: TestX'`,
		"escaped pipe prefix on table row": "| row | `--- (PASS\\|FAIL): TestX` |",
		"escaped dot then quote":           `grep -E -- '--- PASS: TestX/spec-workflow\.md'`,
	}
	for name, line := range detect {
		t.Run("detect/"+name, func(t *testing.T) {
			got := vtaCheckLine(t, line)
			if n := vtaAxisCount(got, "outcome-assertion"); n != 1 || len(got) != 1 {
				t.Errorf("line %q: outcome findings = %d (all = %+v), want exactly 1", line, n, got)
			}
		})
	}

	t.Run("detect/hyphenated subtest read whole", func(t *testing.T) {
		got := vtaCheckLine(t, `grep -- '--- PASS: TestX/fail-open_path'`)
		if len(got) != 1 {
			t.Fatalf("findings = %+v, want exactly 1", got)
		}
		if !strings.Contains(got[0].Message, "TestX/fail-open_path") {
			t.Errorf("message %q does not name the whole subtest name", got[0].Message)
		}
	})

	conform := map[string]string{
		"space":                      `grep -F -- '--- PASS: TestX '`,
		"tab":                        "grep -P -- '--- PASS: TestX\t'",
		"backslash s":                `grep -E -- '--- PASS: TestX\s'`,
		"posix space class":          `grep -E -- '--- PASS: TestX[[:space:]]'`,
		"hyphen subtest then space":  "the output contains `--- PASS: TestX/fail-open_path (`",
		"dot subtest then space":     "the output contains `--- PASS: TestX/spec-workflow.md (`",
		"subtest then posix class":   `grep -E -- '--- PASS: TestX/case[[:space:]]'`,
		"escaped dot then space":     `grep -E -- '--- PASS: TestX/spec-workflow\.md '`,
		"subtest then backslash s":   `grep -E -- '--- PASS: TestX/case\s'`,
		"prefix with no name":        "the four prefixes start with `--- PASS: ` and friends",
		"grouped prefix delimited":   `grep -E -- '--- (PASS|FAIL): TestX '`,
		"escaped pipe prefix spaced": "| row | `--- (PASS\\|FAIL): TestX (` |",
	}
	for name, line := range conform {
		t.Run("conform/"+name, func(t *testing.T) {
			if got := vtaCheckLine(t, line); len(got) != 0 {
				t.Errorf("line %q: findings = %+v, want none", line, got)
			}
		})
	}
}

// TestVacuousAssertionRule_MarkdownContexts — AC-VTA-003. A rule that skips
// blockquotes (mutant M3) fails the blockquote rows here.
func TestVacuousAssertionRule_MarkdownContexts(t *testing.T) {
	type pair struct{ defective, conformant string }
	axes := map[string]pair{
		"run":     {`go test ./x -run 'TestA' -v`, `go test ./x -run '^TestA$' -v`},
		"outcome": {`grep -- '--- PASS: TestA'`, `grep -- '--- PASS: TestA '`},
	}
	wrap := map[string]func(string) []string{
		"fenced":            func(s string) []string { return []string{"```bash", s, "```"} },
		"inline":            func(s string) []string { return []string{"Run `" + s + "` now."} },
		"table":             func(s string) []string { return []string{"| step | `" + s + "` |"} },
		"prose":             func(s string) []string { return []string{"Then " + s + " must hold."} },
		"blockquote":        func(s string) []string { return []string{"> " + s} },
		"nested blockquote": func(s string) []string { return []string{"> > " + s} },
	}
	for ctx, w := range wrap {
		for axis, p := range axes {
			t.Run(ctx+"/"+axis, func(t *testing.T) {
				if got := vtaCheckLine(t, strings.Join(w(p.defective), "\n")); len(got) != 1 {
					t.Errorf("defective %s in %s: findings = %+v, want exactly 1", axis, ctx, got)
				}
				if got := vtaCheckLine(t, strings.Join(w(p.conformant), "\n")); len(got) != 0 {
					t.Errorf("conformant %s in %s: findings = %+v, want none", axis, ctx, got)
				}
			})
		}
	}

	// A markdown-escaped pipe on a table row is an alternation in both arms.
	t.Run("table escaped pipe/detect", func(t *testing.T) {
		line := "| a | `go test ./x -run '^TestA$\\|TestB$' -v` |"
		if got := vtaCheckLine(t, line); len(got) != 1 {
			t.Errorf("findings = %+v, want exactly 1", got)
		}
	})
	t.Run("table escaped pipe/conform", func(t *testing.T) {
		line := "| a | `go test ./x -run '^TestA$\\|^TestB$' -v` |"
		if got := vtaCheckLine(t, line); len(got) != 0 {
			t.Errorf("findings = %+v, want none", got)
		}
	})
	t.Run("blockquoted table escaped pipe/detect", func(t *testing.T) {
		line := "> | a | `go test ./x -run '^TestA$\\|TestB$' -v` |"
		if got := vtaCheckLine(t, line); len(got) != 1 {
			t.Errorf("findings = %+v, want exactly 1", got)
		}
	})
}

// TestVacuousAssertionRule_D15Regression — AC-VTA-004. The t1243 iter-3 probe
// shapes: the dollar-quote sequence elsewhere on a line never makes a pattern
// conformant (mutant M4 fails here).
func TestVacuousAssertionRule_D15Regression(t *testing.T) {
	cases := []struct {
		line string
		want int
	}{
		{`go test ./x -run 'TestA' -v 2>&1 | grep -vE "$'" | wc -l`, 1},
		{`go test ./x -run 'TestA|^TestB$' -v`, 1},
		{`go test ./x -run '^TestA$' -v`, 0},
	}
	for _, c := range cases {
		if got := vtaCheckLine(t, c.line); len(got) != c.want {
			t.Errorf("line %q: findings = %+v, want %d", c.line, got, c.want)
		}
	}
}

// TestVacuousAssertionRule_ArtifactScope — AC-VTA-005.
func TestVacuousAssertionRule_ArtifactScope(t *testing.T) {
	root := t.TempDir()
	bad := `go test ./x -run 'TestA' -v`
	specBody := vtaFrontmatter(vacuousGateCutoff, "") + "\n# Fixture\n\n" + bad + "\n"
	doc := vtaWriteDir(t, filepath.Join(root, "SPEC-VTAT-001"), vacuousGateCutoff, map[string]string{
		"spec.md":         specBody,
		"plan.md":         "# plan\n" + bad + "\n",
		"acceptance.md":   "# acc\n\nintro\n" + bad + "\n",
		"progress.md":     bad + "\n",
		"spec-compact.md": bad + "\n",
		"research.md":     bad + "\n",
		"design.md":       bad + "\n",
	})
	vtaWriteDir(t, filepath.Join(root, "SPEC-VTAT-002"), vacuousGateCutoff, map[string]string{
		"spec.md":       vtaFrontmatter(vacuousGateCutoff, "") + bad + "\n",
		"acceptance.md": bad + "\n",
	})

	got := (&VacuousTestAssertionRule{}).Check(doc, nil)
	want := map[string]int{
		filepath.Join(root, "SPEC-VTAT-001", "spec.md"):       strings.Count(strings.Split(specBody, bad)[0], "\n") + 1,
		filepath.Join(root, "SPEC-VTAT-001", "plan.md"):       2,
		filepath.Join(root, "SPEC-VTAT-001", "acceptance.md"): 4,
	}
	if len(got) != 3 {
		t.Fatalf("findings = %+v, want exactly 3", got)
	}
	for _, f := range got {
		line, ok := want[f.File]
		if !ok {
			t.Errorf("finding in unscanned artifact %s", f.File)
			continue
		}
		if f.Line != line {
			t.Errorf("%s: line = %d, want %d", filepath.Base(f.File), f.Line, line)
		}
	}

	for _, line := range []string{
		`go test ./x -run "$PATTERN" -v`,
		`go test ./x -run "^${NAME}" -v`,
		`go test ./x -run "$(cat names)" -v`,
		`go test ./x -run $_x -v`,
	} {
		if got := vtaCheckLine(t, line); len(got) != 0 {
			t.Errorf("shell expansion %q: findings = %+v, want none (stated limit)", line, got)
		}
	}
}

// TestVacuousAssertionRule_GateCutoff — AC-VTA-006, with the N4 plan-audit debt
// (a present-but-malformed created fails closed).
func TestVacuousAssertionRule_GateCutoff(t *testing.T) {
	// Pinning: moving the cutoff requires editing this literal (REQ-VTA-010).
	if vacuousGateCutoff != "2026-09-27" {
		t.Fatalf("vacuousGateCutoff = %q, want the pinned 2026-09-27", vacuousGateCutoff)
	}

	cases := []struct {
		name, created string
		advisory      bool
	}{
		{"day before", "2026-09-26", true},
		{"equal", "2026-09-27", false},
		{"after", "2026-10-15", false},
		{"missing", "", true},
		{"quoted equal", `"2026-09-27"`, false},
		{"malformed month", "2026-9-30", false},
		{"not a date", "tomorrow", false},
	}
	bad := `go test ./x -run 'TestA' -v`
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			created := strings.Trim(c.created, `"`)
			doc := vtaWriteDir(t, t.TempDir(), created, map[string]string{
				"plan.md": bad + "\n",
			})
			got := (&VacuousTestAssertionRule{}).Check(doc, nil)
			if len(got) != 1 {
				t.Fatalf("findings = %+v, want exactly 1", got)
			}
			if got[0].Severity != SeverityWarning {
				t.Errorf("severity = %q, want warning", got[0].Severity)
			}
			if got[0].Advisory != c.advisory {
				t.Errorf("created %q: advisory = %v, want %v", c.created, got[0].Advisory, c.advisory)
			}
		})
	}

	// The committed red fixture must stay gated under the pinned cutoff, so a
	// cutoff move cannot silently turn the end-to-end proof advisory.
	data, err := os.ReadFile(filepath.Join("testdata", "vacuous_assert_e2e", "red", ".moai", "specs",
		"SPEC-FIXTURE-VTA-001", "spec.md"))
	if err != nil {
		t.Fatalf("read red fixture: %v", err)
	}
	m := regexp.MustCompile(`(?m)^created:\s*"?([0-9-]+)"?\s*$`).FindStringSubmatch(string(data))
	if m == nil {
		t.Fatal("red fixture carries no created value")
	}
	if m[1] < vacuousGateCutoff {
		t.Errorf("red fixture created %s is before the cutoff %s", m[1], vacuousGateCutoff)
	}
}

// vtaLint runs the full Linter over one SPEC directory holding spec.md.
func vtaLint(t *testing.T, specBody string) []Finding {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "SPEC-VTAT-001")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(path, []byte(specBody), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := NewLinter(LinterOptions{BaseDir: filepath.Dir(dir)}).Lint([]string{path})
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	var out []Finding
	for _, f := range report.Findings {
		if f.Code == vtaCode {
			out = append(out, f)
		}
	}
	return out
}

// TestVacuousAssertionRule_LintSkip — AC-VTA-007.
func TestVacuousAssertionRule_LintSkip(t *testing.T) {
	bad := `go test ./x -run 'TestA' -v`
	skip := "lint:\n  skip: [VacuousTestAssertion]\n"

	if got := vtaLint(t, vtaFrontmatter(vacuousGateCutoff, skip)+"\n"+bad+"\n"); len(got) != 0 {
		t.Errorf("with lint.skip: findings = %+v, want none", got)
	}
	if got := vtaLint(t, vtaFrontmatter(vacuousGateCutoff, "")+"\n"+bad+"\n"); len(got) < 1 {
		t.Error("without lint.skip: no findings, want at least 1")
	}
	commented := bad + " <!-- vacuous-assertion-ok: documented example -->"
	if got := vtaLint(t, vtaFrontmatter(vacuousGateCutoff, "")+"\n"+commented+"\n"); len(got) != 1 {
		t.Errorf("commented line: findings = %+v, want exactly 1", got)
	}
}

// TestVacuousAssertionRule_Registered — AC-VTA-008.
func TestVacuousAssertionRule_Registered(t *testing.T) {
	n := 0
	for _, r := range NewLinter(LinterOptions{}).rules {
		if r.Code() == vtaCode {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("rules with code %s = %d, want exactly 1", vtaCode, n)
	}

	got := vtaLint(t, vtaFrontmatter(vacuousGateCutoff, "")+
		"\n`go test ./x -run 'TestA' -v`\n\n`grep -- '--- PASS: TestB'`\n")
	if len(got) != 2 {
		t.Fatalf("findings = %+v, want 2", got)
	}
	for _, f := range got {
		if f.Severity != SeverityWarning {
			t.Errorf("severity = %q, want warning", f.Severity)
		}
		if f.Advisory {
			t.Errorf("finding on a post-cutoff V3R6 SPEC is advisory: %+v", f)
		}
	}
	run, out := got[0].Message, got[1].Message
	if !strings.HasPrefix(run, "run-pattern:") || !strings.Contains(run, "'TestA'") || !strings.Contains(run, "'^TestA$'") {
		t.Errorf("run-pattern message %q must name the axis, quote the text, and state '^TestA$'", run)
	}
	if !strings.HasPrefix(out, "outcome-assertion:") || !strings.Contains(out, "--- PASS: TestB") ||
		!strings.Contains(out, "'--- PASS: TestB '") {
		t.Errorf("outcome message %q must name the axis, quote the text, and state the delimited form", out)
	}
}
