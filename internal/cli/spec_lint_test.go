package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// SPEC-SPEC-LINT-ID-ARG-001 (card t518) — argument-shape contract for
// `moai spec lint`.
//
// Judgement discipline carried from acceptance.md §A:
//   - every assertion is made on BOTH the ID form and the path form (the path
//     form is this SPEC's control group; if it moves, nothing is attributable);
//   - the exit code is never the sole discriminator — it is read alongside
//     output content. In-process the code is read from *exitCodeError, never
//     through a shell pipe (a piped $? reads the last stage);
//   - fixtures live under t.TempDir(); .moai/specs/ is never mutated.

// prerepairPathFormCensus is the severity/code census the path form produced on
// this fixture BEFORE the resolver existed, captured at M1 (evidence:
// .moai/reports/t518/mb1-red.txt). It is the control group's frozen value.
const prerepairPathFormCensus = "INFO/OwnershipTransitionUnreachable," +
	"INFO/StatusGitUnreachable," +
	"WARNING/CoverageIncomplete," +
	"WARNING/FrontmatterInvalid,WARNING/FrontmatterInvalid,WARNING/FrontmatterInvalid," +
	"WARNING/FrontmatterInvalid,WARNING/FrontmatterInvalid,WARNING/FrontmatterInvalid," +
	"WARNING/MissingExclusions"

const fixtureSpecBody = `---
id: SPEC-FIX-001
title: "fixture spec"
version: "0.1.0"
status: draft
created: 2026-09-07
updated: 2026-09-07
---

## Requirements

- **REQ-FIX-001** (Ubiquitous) — The fixture shall exist.

## Out of Scope

- nothing
`

// newFixtureProject writes a project root holding one SPEC and returns it.
// The caller's cwd is moved into the root, and findProjectRootFn is pointed at
// it, so both the no-arg corpus scan (detectBaseDir(cwd)) and the ID resolver
// (findProjectRootFn) see the fixture and never the real repository.
func newFixtureProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-FIX-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(fixtureSpecBody), 0o644); err != nil {
		t.Fatalf("write fixture spec.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	t.Chdir(root)
	return root
}

// runSpecLint executes the subcommand in-process and returns its combined
// output plus the exit code carried by *exitCodeError (0 when the command
// returned nil). No pipe is involved, so the code is read from the command
// itself rather than from a shell's $?.
func runSpecLint(t *testing.T, args ...string) (string, int) {
	t.Helper()
	cmd := newSpecLintCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	out := buf.String()
	if err == nil {
		return out, 0
	}
	// The error's own message is deliberately NOT appended to the returned
	// output. Appending it would let a test assert on text the user never
	// sees: this command's exit-3 channel is silent at the terminal (fang
	// renders nothing for an ExitCoder), so a diagnostic that lives only
	// inside the error object is invisible in practice. Assertions are made
	// on what the command actually wrote.
	var ec *exitCodeError
	if errors.As(err, &ec) {
		return out, ec.ExitCode()
	}
	return out, -1
}

var findingLineRe = regexp.MustCompile(`^(ERROR|WARNING|INFO) (\S+) (\S+) (\d+) `)

type finding struct {
	severity string
	code     string
	file     string
	line     string
}

// parseFindings reads the printed table back into structured rows. The table
// is the user-visible surface the ACs speak about (FILE value, finding code,
// summary line), so the assertions are made on it rather than on internals.
func parseFindings(out string) []finding {
	var got []finding
	for _, ln := range strings.Split(out, "\n") {
		collapsed := strings.Join(strings.Fields(ln), " ") + " "
		m := findingLineRe.FindStringSubmatch(collapsed)
		if m == nil {
			continue
		}
		got = append(got, finding{severity: m[1], code: m[2], file: m[3], line: m[4]})
	}
	return got
}

func summaryLine(out string) string {
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "error(s),") && strings.Contains(ln, "warning(s)") {
			return strings.TrimSpace(ln)
		}
	}
	return ""
}

func countCode(out, code string) int {
	n := 0
	for _, f := range parseFindings(out) {
		if f.code == code {
			n++
		}
	}
	return n
}

func codeSeverityCensus(out string) []string {
	var census []string
	for _, f := range parseFindings(out) {
		census = append(census, f.severity+"/"+f.code)
	}
	sort.Strings(census)
	return census
}

// AC-SLI-001a (REQ-SLI-001) — an ID-shaped argument resolves to that SPEC's
// spec.md, and no ParseFailure is emitted.
// Mutant guard: removing the resolver makes ParseFailure reappear (RED).
func TestSpecLint_IDArg_ResolvesToSpecMD(t *testing.T) {
	root := newFixtureProject(t)
	out, rc := runSpecLint(t, "SPEC-FIX-001")

	want := filepath.Join(root, ".moai", "specs", "SPEC-FIX-001", "spec.md")
	if n := countCode(out, "ParseFailure"); n != 0 {
		t.Fatalf("ParseFailure count = %d, want 0 (rc=%d)\n%s", n, rc, out)
	}
	rows := parseFindings(out)
	if len(rows) == 0 {
		t.Fatalf("expected at least one finding row naming the resolved file\n%s", out)
	}
	for _, f := range rows {
		if f.file != want {
			t.Fatalf("FILE = %q, want resolved path %q\n%s", f.file, want, out)
		}
	}
}

// AC-SLI-001b (REQ-SLI-002) — control pair: the ID form and the path form
// produce the same findings, the same FILE values and the same summary line.
func TestSpecLint_IDArg_EqualsPathForm(t *testing.T) {
	root := newFixtureProject(t)
	pathArg := filepath.Join(root, ".moai", "specs", "SPEC-FIX-001", "spec.md")

	idOut, idRC := runSpecLint(t, "SPEC-FIX-001")
	pathOut, pathRC := runSpecLint(t, pathArg)

	if idRC != pathRC {
		t.Fatalf("exit codes differ: id=%d path=%d\n--- id ---\n%s\n--- path ---\n%s", idRC, pathRC, idOut, pathOut)
	}
	if got, want := summaryLine(idOut), summaryLine(pathOut); got != want {
		t.Fatalf("summary differs: id=%q path=%q", got, want)
	}
	idRows, pathRows := parseFindings(idOut), parseFindings(pathOut)
	if len(idRows) != len(pathRows) {
		t.Fatalf("finding count differs: id=%d path=%d\n--- id ---\n%s\n--- path ---\n%s", len(idRows), len(pathRows), idOut, pathOut)
	}
	for i := range idRows {
		if idRows[i] != pathRows[i] {
			t.Fatalf("finding %d differs: id=%+v path=%+v", i, idRows[i], pathRows[i])
		}
	}
}

// AC-SLI-001c (REQ-SLI-001) — resolution is anchored on findProjectRootFn (the
// sibling `view` idiom), so it still works from a subdirectory. Using
// detectBaseDir(cwd) instead would fail here.
// Mutant guard: swapping the base to detectBaseDir(cwd) makes this RED.
func TestSpecLint_IDArg_FromSubdirectory(t *testing.T) {
	root := newFixtureProject(t)
	t.Chdir(filepath.Join(root, "internal"))

	out, rc := runSpecLint(t, "SPEC-FIX-001")
	if rc == 3 {
		t.Fatalf("lint reported 'no such SPEC' from a subdirectory while view resolves it; rc=3\n%s", out)
	}
	if n := countCode(out, "ParseFailure"); n != 0 {
		t.Fatalf("ParseFailure count = %d, want 0 from a subdirectory\n%s", n, out)
	}
}

// AC-SLI-002 (REQ-SLI-003) — a SPEC directory resolves to the spec.md inside
// it, matching the path form.
// Mutant guard: removing the directory branch reproduces ParseFailure
// ("is a directory").
func TestSpecLint_DirectoryArg(t *testing.T) {
	root := newFixtureProject(t)
	dirArg := filepath.Join(".moai", "specs", "SPEC-FIX-001")
	fileArg := filepath.Join(root, ".moai", "specs", "SPEC-FIX-001", "spec.md")

	dirOut, dirRC := runSpecLint(t, dirArg)
	if n := countCode(dirOut, "ParseFailure"); n != 0 {
		t.Fatalf("ParseFailure count = %d for directory form, want 0 (rc=%d)\n%s", n, dirRC, dirOut)
	}
	pathOut, _ := runSpecLint(t, fileArg)
	if got, want := summaryLine(dirOut), summaryLine(pathOut); got != want {
		t.Fatalf("directory-form summary %q != path-form summary %q", got, want)
	}
}

// AC-SLI-003 (REQ-SLI-005) — an ID-shaped argument that resolves to nothing is
// an argument-type error (exit code 3) naming the attempted path, NOT a
// per-document ParseFailure finding.
// The ParseFailure==0 half is an absence assertion and is vacuous alone; it is
// read as a pair with AC-SLI-001a.
// Mutant guard: removing the diagnostic branch brings ParseFailure back.
func TestSpecLint_UnresolvableID_IsArgumentError(t *testing.T) {
	root := newFixtureProject(t)
	out, rc := runSpecLint(t, "SPEC-NOPE-999")

	if rc != 3 {
		t.Fatalf("exit code = %d, want 3\n%s", rc, out)
	}
	if n := countCode(out, "ParseFailure"); n != 0 {
		t.Fatalf("ParseFailure count = %d, want 0\n%s", n, out)
	}
	attempted := filepath.Join(root, ".moai", "specs", "SPEC-NOPE-999", "spec.md")
	if !strings.Contains(out, attempted) {
		t.Fatalf("message does not name the attempted path %q\n%s", attempted, out)
	}
}

// AC-SLI-004a (REQ-SLI-004) — the control group does not move. The census
// compared here is the PRE-REPAIR behaviour captured at M1 on this fixture,
// before the resolver existed.
// Mutant guard: dropping the path-signal pre-exclusion sends path-form
// arguments into the ID branch and the FILE value changes.
func TestSpecLint_PathForm_Unchanged(t *testing.T) {
	root := newFixtureProject(t)
	pathArg := filepath.Join(root, ".moai", "specs", "SPEC-FIX-001", "spec.md")

	out, rc := runSpecLint(t, pathArg)
	if rc != 0 && rc != 1 {
		t.Fatalf("path form exit code = %d, want the pre-repair 0 or 1\n%s", rc, out)
	}
	for _, f := range parseFindings(out) {
		if f.file != pathArg {
			t.Fatalf("path-form FILE = %q, want the argument itself %q\n%s", f.file, pathArg, out)
		}
	}
	got := strings.Join(codeSeverityCensus(out), ",")
	if got != prerepairPathFormCensus {
		t.Fatalf("path-form census moved:\n got: %s\nwant: %s\n%s", got, prerepairPathFormCensus, out)
	}
}

// AC-SLI-004b (REQ-SLI-004) — the other side of the discriminator: a
// path-shaped argument that does not exist still yields exactly one
// ParseFailure. Without this, AC-SLI-003 alone cannot be told apart from
// "every unresolvable argument was swallowed into exit code 3".
// Mutant guard: routing all unresolved arguments to exit 3 drops this to 0.
func TestSpecLint_MissingPathArg_StillParseFailure(t *testing.T) {
	newFixtureProject(t)
	out, rc := runSpecLint(t, filepath.Join(".", "missing", "spec.md"))

	if n := countCode(out, "ParseFailure"); n != 1 {
		t.Fatalf("ParseFailure count = %d, want 1 (rc=%d)\n%s", n, rc, out)
	}
	if rc != 1 {
		t.Fatalf("exit code = %d, want 1\n%s", rc, out)
	}
}

// AC-SLI-005 (REQ-SLI-006) — the anchor invariant.
//
// `SPEC-A-1` carries NO path signal, so the decision genuinely happens in the
// regexp: the anchored shape rejects it (3-digit rule) and it falls through to
// the path branch, while the unanchored same-named symbol in
// spec_status.go:18 would accept it. A fixture containing "/" cannot measure
// this — the path-signal pre-exclusion decides before the regexp is reached
// (acceptance.md §A.1).
// Mutant guard: swapping the discriminator for spec_status.go:18's unanchored
// specIDPattern makes SPEC-A-1 resolve, changing FILE or producing rc=3.
func TestSpecLint_AnchorInvariant_SPEC_A_1(t *testing.T) {
	newFixtureProject(t)
	out, rc := runSpecLint(t, "SPEC-A-1")

	if rc == 3 {
		t.Fatalf("SPEC-A-1 was treated as an ID (rc=3); the anchored shape must reject it\n%s", out)
	}
	if strings.Contains(out, filepath.Join(".moai", "specs")) {
		t.Fatalf("SPEC-A-1 was resolved under .moai/specs; the anchored shape must reject it\n%s", out)
	}
	rows := parseFindings(out)
	if len(rows) == 0 {
		t.Fatalf("expected the argument to be treated as a path and fail to parse\n%s", out)
	}
	for _, f := range rows {
		if f.file != "SPEC-A-1" {
			t.Fatalf("FILE = %q, want the argument verbatim %q\n%s", f.file, "SPEC-A-1", out)
		}
	}
}

// AC-SLI-006 (REQ-SLI-007) — the help text names all three accepted shapes.
// Mutant guard: reverting to `Use: "lint [spec.md...]"` drops at least one match.
func TestSpecLintHelp_NamesThreeArgumentShapes(t *testing.T) {
	cmd := newSpecLintCmd()
	text := cmd.Use + "\n" + cmd.Long

	for _, want := range []string{"SPEC-ID", "spec.md", "SPEC directory"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help does not mention %q\n--- Use ---\n%s\n--- Long ---\n%s", want, cmd.Use, cmd.Long)
		}
	}
}

// AC-SLI-007 (REQ-SLI-002, REQ-SLI-003) — mixed shapes in one invocation are
// each resolved.
// Mutant guard: resolving only the first argument shrinks the FILE set to 1.
func TestSpecLint_MixedArgs(t *testing.T) {
	root := newFixtureProject(t)
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-MIX-002")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := strings.Replace(fixtureSpecBody, "SPEC-FIX-001", "SPEC-MIX-002", 1)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	pathArg := filepath.Join(root, ".moai", "specs", "SPEC-FIX-001", "spec.md")
	wantResolved := filepath.Join(specDir, "spec.md")

	// BOTH orderings are exercised. With the ID first, an implementation that
	// resolves only the first argument still looks correct here, because the
	// second argument is a path and needs no resolution — the mutant M8 was a
	// no-op against a single ordering, and that is how this second case came
	// to exist.
	for _, order := range [][]string{
		{"SPEC-MIX-002", pathArg},
		{pathArg, "SPEC-MIX-002"},
	} {
		out, rc := runSpecLint(t, order...)
		if n := countCode(out, "ParseFailure"); n != 0 {
			t.Fatalf("order %v: ParseFailure count = %d, want 0 (rc=%d)\n%s", order, n, rc, out)
		}
		files := map[string]bool{}
		for _, f := range parseFindings(out) {
			files[f.file] = true
		}
		if !files[wantResolved] || !files[pathArg] {
			t.Fatalf("order %v: FILE set = %v, want both %q and %q\n%s", order, files, wantResolved, pathArg, out)
		}
	}
}

// AC-SLI-010 branch 2 (REQ-SLI-009, REQ-SLI-006) — literal freeze.
//
// The repository now holds TWO SPEC-ID regexps of different strictness, on
// purpose: the anchored shape check in internal/cli/specid (the argument
// discriminator) and this unanchored one, which is correct where it lives —
// it scrapes SPEC-IDs out of git commit messages (spec_status.go:248), and an
// anchor there would break that job. The disposition is recorded in
// spec.md §H: this card does NOT consolidate them.
//
// The residual risk has a name — same-name/different-meaning reuse drift: the
// next author looking for "the SPEC-ID regexp" may pick the loose one and
// reuse it as an argument discriminator, which misreads a path as an ID. This
// test freezes the literal so that this symbol silently widening, or quietly
// acquiring an anchor, is detected.
//
// What it does NOT catch: a change on the CALLING side that gives the symbol a
// different meaning. That axis is covered by AC-SLI-005's mutant.
func TestSpecStatusIDPattern_LiteralFrozen(t *testing.T) {
	const frozen = `SPEC-[A-Z0-9-]+-[0-9]+`
	if got := specIDPattern.String(); got != frozen {
		t.Fatalf("spec_status.go specIDPattern literal moved:\n got: %s\nwant: %s", got, frozen)
	}
}

// AC-SLI-009 branch 1 (REQ-SLI-009, REQ-SLI-006) — the shape function's home
// and name. It lives in the existing leaf package internal/cli/specid (option
// beta), not as a private copy inside spec_lint.go (option alpha), and its
// name differs from the same-named symbol specIDPattern in spec_status.go:18.
// Judged by reading the declaring file path and the name.
// Mutant guard: moving the declaration into spec_lint.go makes this RED.
//
// Branch 2 (the five-call-site no-change proof) is milestone M-B2 and is not
// asserted here.
func TestShapeFunction_HomeAndName(t *testing.T) {
	const decl = "func HasCanonicalSpecIDShape("

	specidSrc, err := os.ReadFile(filepath.Join("specid", "specid.go"))
	if err != nil {
		t.Fatalf("read internal/cli/specid/specid.go: %v", err)
	}
	if !strings.Contains(string(specidSrc), decl) {
		t.Fatalf("shape function is not declared in internal/cli/specid/specid.go")
	}

	lintSrc, err := os.ReadFile("spec_lint.go")
	if err != nil {
		t.Fatalf("read spec_lint.go: %v", err)
	}
	if strings.Contains(string(lintSrc), decl) {
		t.Fatalf("shape function is declared in spec_lint.go (option alpha); beta was ruled")
	}
	// The requirement is that the DECLARED symbol carries a different name, not
	// that the string never appears: the shape function's doc comment names the
	// loose symbol on purpose, to point the next reader at the distinction.
	if regexp.MustCompile(`(?m)^\s*(var|func|const)\s+specIDPattern\b`).Match(specidSrc) {
		t.Fatalf("the shape symbol reuses the name specIDPattern; REQ-SLI-006 requires a different name")
	}
}
