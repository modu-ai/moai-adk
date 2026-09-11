package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
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

// ---------------------------------------------------------------------------
// Fixtures
//
// The corpus is synthetic and lives in t.TempDir(); the real .moai/specs/ tree
// is never read, so no assertion here depends on the repository's own standing
// inventory size (REQ-SLGS-003 — no warning-count literal is frozen anywhere).
//
// Two injection vectors, both confirmed by measurement before use:
//   - a SPEC-*/ directory carrying no spec.md yields exactly one NON-ADVISORY
//     warning, SpecsDirMissingSpecFile (a directory-level finding, so the era
//     demotion that marks per-SPEC findings advisory never reaches it);
//   - a modern-era spec.md with its "Out of Scope" section removed yields one
//     ERROR-severity MissingExclusions.
// ---------------------------------------------------------------------------

const slOutOfScopeHeading = "## 5. Scope and Out of Scope"

const slCleanSpecMD = `---
id: SPEC-PROBE-001
title: "Probe"
version: "0.1.0"
status: in-progress
created: 2026-09-01
updated: 2026-09-01
author: probe
priority: P1
phase: "v3.1.0"
module: "internal/probe"
lifecycle: spec-anchored
tags: "probe"
tier: S
---

# SPEC: Probe

## 2. Requirements

- **REQ-PRB-001** (Ubiquitous): The system shall do the thing.

## 3. Acceptance

### AC-PRB-001 — thing happens (maps REQ-PRB-001)

- Given a thing
- When it runs
- Then it happens

## 5. Scope and Out of Scope

### 5.2 Out of Scope

- nothing else
`

const slProgressMD = `# progress

## §E.2 Run-phase Evidence

none

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: "a1b2c3d4e5f6"
`

// slNewCorpus creates a project root holding .moai/specs/ with one baseline SPEC.
func slNewCorpus(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-PROBE-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir corpus: %v", err)
	}
	slWriteFile(t, filepath.Join(specDir, "spec.md"), slCleanSpecMD)
	slWriteFile(t, filepath.Join(specDir, "progress.md"), slProgressMD)
	return root
}

func slWriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// slInjectWarning adds one SpecsDirMissingSpecFile finding.
func slInjectWarning(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("inject warning: %v", err)
	}
	// A file inside keeps the directory non-empty and realistic.
	slWriteFile(t, filepath.Join(dir, "research.md"), "# umbrella doc\n")
	return dir
}

// slInjectError removes the Out of Scope section, producing MissingExclusions.
func slInjectError(t *testing.T, root string) {
	t.Helper()
	idx := strings.Index(slCleanSpecMD, slOutOfScopeHeading)
	if idx < 0 {
		t.Fatalf("fixture drift: %q not found in the clean SPEC body", slOutOfScopeHeading)
	}
	slWriteFile(t, filepath.Join(root, ".moai", "specs", "SPEC-PROBE-001", "spec.md"), slCleanSpecMD[:idx])
}

func slRestoreClean(t *testing.T, root string) {
	t.Helper()
	slWriteFile(t, filepath.Join(root, ".moai", "specs", "SPEC-PROBE-001", "spec.md"), slCleanSpecMD)
}

func slFileSHA256(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("hash %s: %v", path, err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// slRunLint executes the real cobra command in `dir`, returning combined
// stdout and stderr as the user would see them, plus the exit code the process
// would have used.
func slRunLint(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	stdout, stderr, code := slRunLintStreams(t, dir, args...)
	return stdout + stderr, code
}

// slRunLintStreams is slRunLint with the two streams kept apart, for tests that
// must tell where a message was written and how many times.
//
// The error's own message is appended to stderr ONLY where the real binary
// would render it. An *exitCodeError is rendered as nothing (moaiErrorHandler
// stays silent for ExitCoder carriers), so its text reaches the user only if
// the command wrote it itself; appending it here would let a test assert on
// text the user never sees — the same rule runSpecLint documents. Any other
// error is rendered by fang's default handler, so it is appended to mirror that.
func slRunLintStreams(t *testing.T, dir string, args ...string) (string, string, int) {
	t.Helper()
	t.Chdir(dir)

	cmd := newSpecLintCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		return out.String(), errBuf.String(), 0
	}
	var ec *exitCodeError
	if errors.As(err, &ec) {
		return out.String(), errBuf.String(), ec.ExitCode()
	}
	return out.String(), errBuf.String() + err.Error() + "\n", 1
}

// slCaptureBaseline captures the corpus's current state as the baseline file.
func slCaptureBaseline(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "baseline.json")
	out, code := slRunLint(t, root, "--baseline", path, "--update-baseline", "--reason", "initial capture")
	if code != 0 {
		t.Fatalf("baseline capture failed (rc=%d):\n%s", code, out)
	}
	fi, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("baseline file not written: %v", statErr)
	}
	if fi.Size() == 0 {
		t.Fatalf("baseline file is empty")
	}
	return path
}

// ---------------------------------------------------------------------------
// AC-SLGS-005 — a new non-advisory warning turns the gate red and names the
// rule; removing it returns the gate to green.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_InjectedWarningTurnsRedAndNamesTheRule(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	// Precondition: the unchanged corpus is green against its own baseline.
	// Without this the red below could come from a broken capture rather than
	// from the injection.
	if preOut, preCode := slRunLint(t, root, "--baseline", baseline); preCode != 0 {
		t.Fatalf("precondition: unchanged corpus must be green, rc=%d\n%s", preCode, preOut)
	}

	dir := slInjectWarning(t, root, "SPEC-INJECTED-001")

	redOut, redCode := slRunLint(t, root, "--baseline", baseline)
	if redCode != 1 {
		t.Fatalf("injected violation must exit 1, got %d\n%s", redCode, redOut)
	}
	if !strings.Contains(redOut, "SpecsDirMissingSpecFile") {
		t.Errorf("output must name the rule that grew:\n%s", redOut)
	}
	if !strings.Contains(redOut, "+1") {
		t.Errorf("output must show the per-rule delta (+1):\n%s", redOut)
	}
	if !strings.Contains(strings.ToUpper(redOut), "EXCEEDED") {
		t.Errorf("output must state the baseline was exceeded:\n%s", redOut)
	}

	// Remove the violation — the gate returns to green.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove injected dir: %v", err)
	}
	greenOut, greenCode := slRunLint(t, root, "--baseline", baseline)
	if greenCode != 0 {
		t.Fatalf("removing the violation must return rc=0, got %d\n%s", greenCode, greenOut)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-006 — the standing inventory passes and stays observable.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_UnchangedCorpusPassesAndPrintsWarningTotal(t *testing.T) {
	root := slNewCorpus(t)
	// Give the corpus a standing non-advisory inventory so the pass is not
	// vacuous: something must actually be absorbed by the baseline.
	slInjectWarning(t, root, "SPEC-STANDING-001")
	slInjectWarning(t, root, "SPEC-STANDING-002")

	baseline := slCaptureBaseline(t, root)

	// The captured baseline must actually record the standing inventory —
	// a baseline of zero rules would pass for the wrong reason.
	b := slReadBaseline(t, baseline)
	if b.Rules["SpecsDirMissingSpecFile"] != 2 {
		t.Fatalf("precondition: baseline must record the standing inventory, got %v", b.Rules)
	}

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 0 {
		t.Fatalf("standing inventory must pass, rc=%d\n%s", code, out)
	}
	if !strings.Contains(out, "warning(s)") {
		t.Errorf("the warning total must remain visible while green:\n%s", out)
	}
	if !strings.Contains(strings.ToUpper(out), "OK") {
		t.Errorf("output must state the baseline verdict:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-007 — a decrease passes, is reported, and does NOT rewrite the file.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_DecreasePassesAndLeavesFileByteIdentical(t *testing.T) {
	root := slNewCorpus(t)
	dir := slInjectWarning(t, root, "SPEC-STANDING-001")
	slInjectWarning(t, root, "SPEC-STANDING-002")
	baseline := slCaptureBaseline(t, root)

	before := slFileSHA256(t, baseline)

	// Fix one violation: the count drops from 2 to 1.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove: %v", err)
	}

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 0 {
		t.Fatalf("a decrease must pass, rc=%d\n%s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "decreas") {
		t.Errorf("the improvement must be reported:\n%s", out)
	}

	after := slFileSHA256(t, baseline)
	if before != after {
		t.Errorf("the baseline file must not shrink implicitly — a silent rewrite bypasses the rebaseline audit.\nbefore=%s\nafter =%s", before, after)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-008 — rebaseline is explicit, requires a non-empty reason, and
// records tree SHA + date + reason.
// ---------------------------------------------------------------------------

func TestSpecLintUpdateBaseline_RejectsMissingOrEmptyReason(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)
	before := slFileSHA256(t, baseline)

	cases := []struct {
		name string
		args []string
	}{
		{"reason omitted", []string{"--baseline", baseline, "--update-baseline"}},
		{"reason empty", []string{"--baseline", baseline, "--update-baseline", "--reason", ""}},
		{"reason whitespace only", []string{"--baseline", baseline, "--update-baseline", "--reason", "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, code := slRunLint(t, root, tc.args...)
			if code != 3 {
				t.Errorf("must be rejected with exit 3, got %d\n%s", code, out)
			}
			if after := slFileSHA256(t, baseline); after != before {
				t.Errorf("a rejected rebaseline must leave the file byte-unchanged")
			}
		})
	}
}

func TestSpecLintUpdateBaseline_RecordsTreeSHADateAndReason(t *testing.T) {
	root := slNewCorpus(t)
	slInjectWarning(t, root, "SPEC-STANDING-001")

	path := filepath.Join(t.TempDir(), "baseline.json")
	out, code := slRunLint(t, root, "--baseline", path, "--update-baseline", "--reason", "absorb t518 population shift")
	if code != 0 {
		t.Fatalf("rebaseline must succeed, rc=%d\n%s", code, out)
	}

	b := slReadBaseline(t, path)
	if strings.TrimSpace(b.Reason) != "absorb t518 population shift" {
		t.Errorf("reason not recorded: %q", b.Reason)
	}
	if strings.TrimSpace(b.TreeSHA) == "" {
		t.Errorf("tree_sha not recorded: %+v", b)
	}
	if len(b.UpdatedAt) != len("2026-09-08") {
		t.Errorf("updated_at must be an ISO date, got %q", b.UpdatedAt)
	}
	if b.Rules["SpecsDirMissingSpecFile"] != 1 {
		t.Errorf("recomputed counts missing: %v", b.Rules)
	}
	// The same three fields must be reported to the operator, not only written.
	for _, want := range []string{"absorb t518 population shift", b.TreeSHA, b.UpdatedAt} {
		if !strings.Contains(out, want) {
			t.Errorf("rebaseline output must report %q:\n%s", want, out)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-009 — an error fails regardless of the baseline, and the reported
// cause is the error, not a baseline increase. The cause distinction is what
// makes the error-BEFORE-baseline ordering observable: both exit 1.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_ErrorFailsRegardlessOfBaseline(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	if preOut, preCode := slRunLint(t, root, "--baseline", baseline); preCode != 0 {
		t.Fatalf("precondition: corpus must be green before the error injection, rc=%d\n%s", preCode, preOut)
	}

	slInjectError(t, root)

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("an error must fail the gate even with a passing baseline, rc=%d\n%s", code, out)
	}
	if !strings.Contains(out, "MissingExclusions") {
		t.Errorf("the error must be visible in the output:\n%s", out)
	}
	if !strings.Contains(strings.ToUpper(out), "ERROR-GATED") {
		t.Errorf("the reported cause must be the error, not a baseline increase:\n%s", out)
	}
	if strings.Contains(strings.ToUpper(out), "EXCEEDED") {
		t.Errorf("an error-gated failure must not be reported as a baseline exceedance:\n%s", out)
	}

	slRestoreClean(t, root)
	if postOut, postCode := slRunLint(t, root, "--baseline", baseline); postCode != 0 {
		t.Fatalf("removing the error must return rc=0, got %d\n%s", postCode, postOut)
	}
}

// The error-BEFORE-baseline ORDERING is only observable when both causes are
// present at once: with an error alone, an implementation that consults the
// baseline first still falls through to the error branch and reports the same
// thing. Added after a mutant that swapped the two branches survived the whole
// suite — this is the case that kills it.
func TestSpecLintBaseline_ErrorOutranksASimultaneousIncrease(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	slInjectWarning(t, root, "SPEC-BOTH-001") // a baseline increase
	slInjectError(t, root)                    // and an error, at the same time

	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("both causes present must exit 1, rc=%d\n%s", code, out)
	}
	// Precondition: the increase really is present, or this asserts nothing.
	if !strings.Contains(out, "SpecsDirMissingSpecFile") {
		t.Fatalf("precondition: the injected increase must be in the report:\n%s", out)
	}
	if !strings.Contains(out, "MissingExclusions") {
		t.Fatalf("precondition: the injected error must be in the report:\n%s", out)
	}

	if !strings.Contains(strings.ToUpper(out), "ERROR-GATED") {
		t.Errorf("the error must be reported as the cause, ahead of the ratchet:\n%s", out)
	}
	if strings.Contains(strings.ToUpper(out), "EXCEEDED") {
		t.Errorf("the ratchet verdict must not preempt the error cause:\n%s", out)
	}
}

// An error also gates the rebaseline path — the baseline records no errors, so
// a rebaseline must not read as a clean run while one stands.
func TestSpecLintUpdateBaseline_ErrorStillExitsOne(t *testing.T) {
	root := slNewCorpus(t)
	slInjectError(t, root)

	path := filepath.Join(t.TempDir(), "baseline.json")
	out, code := slRunLint(t, root, "--baseline", path, "--update-baseline", "--reason", "capture")
	if code != 1 {
		t.Fatalf("a standing error must exit 1 even on rebaseline, rc=%d\n%s", code, out)
	}
	// Guard against a wrong-reason pass: an argument error also exits non-zero.
	// Observed during RED — this test went green purely on "unknown flag".
	if strings.Contains(out, "unknown flag") {
		t.Fatalf("rc=1 came from an argument error, not from the standing error:\n%s", out)
	}
	if !strings.Contains(out, "MissingExclusions") {
		t.Errorf("the standing error must be visible in the output:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Flag contract + boundary cases — no silent fallback anywhere.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_FlagContractRejections(t *testing.T) {
	root := slNewCorpus(t)
	existing := slCaptureBaseline(t, root)
	missing := filepath.Join(t.TempDir(), "absent.json")

	cases := []struct {
		name     string
		args     []string
		wantWord string
	}{
		{"missing baseline file is an argument error, never a silent fallback",
			[]string{"--baseline", missing}, "absent.json"},
		{"--reason without --update-baseline",
			[]string{"--baseline", existing, "--reason", "x"}, "--update-baseline"},
		{"--update-baseline without --baseline",
			[]string{"--update-baseline", "--reason", "x"}, "--baseline"},
		{"--baseline with --json",
			[]string{"--baseline", existing, "--json"}, "--json"},
		{"--baseline with --sarif",
			[]string{"--baseline", existing, "--sarif"}, "--sarif"},
		{"--baseline with --strict",
			[]string{"--baseline", existing, "--strict"}, "--strict"},
		{"--update-baseline with a whitespace-only --reason",
			[]string{"--baseline", existing, "--update-baseline", "--reason", "   "}, "non-empty --reason"},
		// Not a baseline flag, but the same silent-exit-3 shape: the older
		// output-format rejection must be just as visible.
		{"--json with --sarif",
			[]string{"--json", "--sarif"}, "cannot use --json and --sarif together"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := slRunLintStreams(t, root, tc.args...)
			if code != 3 {
				t.Fatalf("want exit 3, got %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
			}
			// The rejection happens before any lint runs, so nothing belongs on
			// stdout; the diagnostic belongs on stderr, written exactly once.
			// An exit code alone names nothing — the real binary renders an
			// exit-3 error as silence, so stderr is the only place the cause
			// can reach the user.
			if stdout != "" {
				t.Errorf("a flag rejection must write nothing to stdout, got:\n%s", stdout)
			}
			if stderr == "" {
				t.Fatalf("the rejection must be written to stderr; stderr is empty (the user sees nothing)")
			}
			lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
			if len(lines) != 1 {
				t.Fatalf("the rejection must be written to stderr exactly once, got %d line(s):\n%s", len(lines), stderr)
			}
			if !strings.Contains(lines[0], tc.wantWord) {
				t.Errorf("the stderr message must name %q:\n%s", tc.wantWord, stderr)
			}
		})
	}
}

// Boundary: a rule that fires for the first time (absent from the baseline)
// reads as recorded 0 and is an increase.
func TestSpecLintBaseline_RuleAbsentFromBaselineIsAnIncrease(t *testing.T) {
	root := slNewCorpus(t)
	baseline := slCaptureBaseline(t, root)

	b := slReadBaseline(t, baseline)
	if b.Rules["SpecsDirMissingSpecFile"] != 0 {
		t.Fatalf("precondition: the rule must be ABSENT from this baseline, got %v", b.Rules)
	}

	slInjectWarning(t, root, "SPEC-FIRSTFIRE-001")
	out, code := slRunLint(t, root, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("a rule firing for the first time must be red, rc=%d\n%s", code, out)
	}
	if !strings.Contains(out, "recorded 0") {
		t.Errorf("an absent rule must read as recorded 0:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Backward compatibility — with no new flag, behaviour is unchanged.
// ---------------------------------------------------------------------------

func TestSpecLint_DefaultAndStrictBehaviourUnchanged(t *testing.T) {
	root := slNewCorpus(t)
	slInjectWarning(t, root, "SPEC-STANDING-001")

	defOut, defCode := slRunLint(t, root)
	if defCode != 0 {
		t.Errorf("default mode must not gate on warnings, rc=%d\n%s", defCode, defOut)
	}
	if !strings.Contains(defOut, "warning(s)") {
		t.Errorf("the default summary line must be preserved:\n%s", defOut)
	}

	strictOut, strictCode := slRunLint(t, root, "--strict")
	if strictCode != 1 {
		t.Errorf("--strict must still escalate a non-advisory warning, rc=%d\n%s", strictCode, strictOut)
	}

	jsonOut, jsonCode := slRunLint(t, root, "--json")
	if jsonCode != 0 {
		t.Errorf("--json must still exit 0 on a warning-only corpus, got %d\n%s", jsonCode, jsonOut)
	}
}

// ---------------------------------------------------------------------------
// AC-SLGS-005 cross-process — the in-process tests above exercise the cobra
// command, but only a real binary invocation can catch a wiring defect between
// the process entry point, the flag, and the exit-code mapping.
// ---------------------------------------------------------------------------

func TestSpecLintBaseline_CrossProcessBothDirections(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-process test builds the CLI binary; skipped under -short")
	}
	bin := slBuildMoaiBinary(t)
	root := slNewCorpus(t)

	baselinePath := filepath.Join(t.TempDir(), "baseline.json")
	if capOut, capCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath, "--update-baseline", "--reason", "cross-process capture"); capCode != 0 {
		t.Fatalf("capture failed rc=%d\n%s", capCode, capOut)
	}

	if preOut, preCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath); preCode != 0 {
		t.Fatalf("precondition: unchanged corpus must be green, rc=%d\n%s", preCode, preOut)
	}

	dir := slInjectWarning(t, root, "SPEC-XPROC-001")
	redOut, redCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath)
	if redCode != 1 {
		t.Fatalf("cross-process: injected violation must exit 1, got %d\n%s", redCode, redOut)
	}
	if !strings.Contains(redOut, "SpecsDirMissingSpecFile") {
		t.Errorf("cross-process output must name the rule:\n%s", redOut)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove: %v", err)
	}
	cleanOut, cleanCode := slRunBinary(t, bin, root, "spec", "lint", "--baseline", baselinePath)
	if cleanCode != 0 {
		t.Fatalf("cross-process: removal must return rc=0, got %d\n%s", cleanCode, cleanOut)
	}
}

func slBuildMoaiBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "moai-test-bin")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/modu-ai/moai-adk/cmd/moai")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build moai binary: %v\n%s", err, out)
	}
	return bin
}

func slRunBinary(t *testing.T, bin, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run %s: %v\n%s", bin, err, out)
	return "", -1
}

// slBaselineFile mirrors the on-disk record so the test decodes it without
// importing the producer's own marshaller — a serialization defect cannot then
// hide behind a symmetric bug in the reader.
type slBaselineFile struct {
	Version   int            `json:"version"`
	UpdatedAt string         `json:"updated_at"`
	TreeSHA   string         `json:"tree_sha"`
	Reason    string         `json:"reason"`
	Rules     map[string]int `json:"rules"`
}

func slReadBaseline(t *testing.T, path string) slBaselineFile {
	t.Helper()
	var b slBaselineFile
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("decode baseline: %v\n%s", err, data)
	}
	return b
}
