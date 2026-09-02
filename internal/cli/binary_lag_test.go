package cli

import (
	"bufio"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/binlag"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// lagBaselineSHA is the commit this SPEC's first commit sits on top of. The
// AC-BLV-009 judgment is a BEFORE/AFTER delta, so it needs a fixed point to
// measure from; a moving reference like a branch name would read an upstream
// commit as this SPEC's own change.
const lagBaselineSHA = "22f90b1c7"

// --- AC-BLV-005: one comparison, two surfaces -------------------------------

// A stub installed in the single seam must be observed by BOTH the doctor item
// and the session-start advisory. A copy of the ancestry logic living on one
// side would still pass the behaviour tests, but would go on calling real git
// here and so miss the stub.
func TestBinaryLag_OneSeamServesBothSurfaces(t *testing.T) {
	const stubBinary = "1234567890abcdef1234567890abcdef12345678"
	const stubHead = "fedcba0987654321fedcba0987654321fedcba09"

	orig := binlag.Comparer
	t.Cleanup(func() { binlag.Comparer = orig })
	binlag.Comparer = func(context.Context, binlag.Request) binlag.Verdict {
		return binlag.Verdict{
			Status:       binlag.StatusBehind,
			BinaryCommit: stubBinary,
			SourceHead:   stubHead,
		}
	}

	// Surface 1 — the doctor check item.
	check := checkBinaryFreshness(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("doctor item status = %v, want Warn under a behind verdict", check.Status)
	}
	if !strings.Contains(check.Message, binlag.Short(stubHead)) {
		t.Errorf("doctor item did not reflect the stub verdict: %q", check.Message)
	}

	// Surface 2 — the unprompted session-start advisory.
	projectDir := t.TempDir()
	// WithSynchronousDeferredScans: this test owns projectDir and t.TempDir
	// deletes it the moment the test body returns. Without the option Handle's
	// deferred MX cold-start scan writes .moai/state/mx-index.json from a
	// goroutine that can outrun the join bound, and the write races that
	// deletion ("unlinkat ... .moai/state: directory not empty").
	// SPEC-TEMPDIR-CLEANUP-RACE-001 REQ-TCR-003.
	out, err := hook.NewSessionStartHandler(nil, hook.WithSynchronousDeferredScans()).Handle(context.Background(), &hook.HookInput{
		SessionID:     "sess-seam",
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	})
	if err != nil {
		t.Fatalf("session start Handle: %v", err)
	}
	if out.HookSpecificOutput == nil ||
		!strings.Contains(out.HookSpecificOutput.AdditionalContext, binlag.Short(stubHead)) {
		t.Errorf("session-start advisory did not reflect the stub verdict; it is not reading the same seam")
	}
}

// --- AC-BLV-003 (c): the not-applicable path leaves the exit status alone ----

// A deployed user project has no repository to compare against. Reporting that
// as a Fail would promote every downstream `moai doctor` run to exit 1, because
// doctorExitStatus escalates on a single Fail.
func TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	check := checkBinaryFreshness(false)
	if check.Status == uikit.CheckFail {
		t.Fatalf("non-git directory reported Fail: %q", check.Message)
	}
	failCount := 0
	if check.Status == uikit.CheckFail {
		failCount = 1
	}
	if err := doctorExitStatus(failCount); err != nil {
		t.Errorf("a run carrying only this check exits non-zero: %v", err)
	}
}

// --- AC-BLV-009: the doctor check-name set gains nothing --------------------

// checkNamesFromSource extracts the name expression of every entry in the three
// check registries.
//
// The unit is the ENTRY's name expression, not a quoted string: an entry may
// name its check with a constant identifier (mcpServerVersionCheckName already
// does), so an extraction that only saw string literals would miss a new entry
// added the same way and report an empty delta while the requirement was being
// broken. The whole slice literal is read, not a line window, because the
// natural place to append a new check is the end.
func checkNamesFromSource(t *testing.T, src []byte) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "doctor.go", src, 0)
	if err != nil {
		t.Fatalf("parse doctor.go: %v", err)
	}

	registries := map[string]bool{"systemChecks": true, "moaiChecks": true, "workspaceChecks": true}
	names := map[string]bool{}

	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			ident, ok := lhs.(*ast.Ident)
			if !ok || !registries[ident.Name] || i >= len(assign.Rhs) {
				continue
			}
			lit, ok := assign.Rhs[i].(*ast.CompositeLit)
			if !ok {
				continue
			}
			for _, elt := range lit.Elts {
				entry, ok := elt.(*ast.CompositeLit)
				if !ok || len(entry.Elts) == 0 {
					continue
				}
				names[exprSource(fset, src, entry.Elts[0])] = true
			}
		}
		return true
	})

	if len(names) == 0 {
		t.Fatal("extracted no check names; the extraction itself is broken, so any delta it reports is meaningless")
	}
	return names
}

// exprSource renders an expression as the source text that produced it, so a
// string literal and a constant identifier are both usable set elements.
func exprSource(fset *token.FileSet, src []byte, e ast.Expr) string {
	start := fset.Position(e.Pos()).Offset
	end := fset.Position(e.End()).Offset
	return string(src[start:end])
}

func TestBinaryLag_DoctorCheckNameSetIsUnchanged(t *testing.T) {
	before, err := exec.Command("git", "show", lagBaselineSHA+":internal/cli/doctor.go").Output()
	if err != nil {
		t.Skipf("baseline blob %s unavailable (shallow clone?): %v", lagBaselineSHA, err)
	}
	after, err := os.ReadFile("doctor.go")
	if err != nil {
		t.Fatal(err)
	}

	beforeNames := checkNamesFromSource(t, before)
	afterNames := checkNamesFromSource(t, after)

	for name := range afterNames {
		if !beforeNames[name] {
			t.Errorf("this SPEC added doctor check name %s; REQ-BLV-009 rewires the existing "+
				"\"Binary Freshness\" item and registers no new name", name)
		}
	}
	for name := range beforeNames {
		if !afterNames[name] {
			t.Errorf("this SPEC removed doctor check name %s", name)
		}
	}
	if !afterNames[`"Binary Freshness"`] {
		t.Error(`"Binary Freshness" is not registered after the change`)
	}
}

// --- AC-BLV-004: monotone build identity, VERSION untouched -----------------

// versionLineExpected is the Makefile's VERSION derivation as it stood at
// lagBaselineSHA. Operator decision (b): this SPEC introduces the monotone
// identity in a separate BUILD_ID and leaves VERSION byte-identical, so the
// release artifact name, version.json, and the update path stay outside this
// card's blast radius.
const versionLineExpected = `VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "dev")`

func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// makefileAssignment returns the full line and the right-hand side of the first
// assignment to name in the Makefile.
func makefileAssignment(t *testing.T, name string) (line string, rhs string) {
	t.Helper()
	f, err := os.Open(filepath.Join(repoRoot(t), "Makefile"))
	if err != nil {
		t.Fatalf("open Makefile: %v", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		for _, op := range []string{" ?= ", " := ", " = "} {
			if strings.HasPrefix(text, name+op) {
				return text, strings.TrimPrefix(text, name+op)
			}
		}
	}
	return "", ""
}

// (c) — the release version derivation is byte-identical to the baseline.
func TestBuildIdentity_VersionDerivationUnchanged(t *testing.T) {
	line, _ := makefileAssignment(t, "VERSION")
	if line != versionLineExpected {
		t.Errorf("the Makefile VERSION line changed.\n got: %s\nwant: %s\n"+
			"The monotone build identity belongs in BUILD_ID; changing VERSION reaches "+
			"RELEASE_BINARY, version.json, and internal/update/local.go.", line, versionLineExpected)
	}
}

// primaryGitDerivation pulls the first alternative out of a Makefile
// `$(shell A || B || C)` value and returns it as argv.
//
// The argv is executed directly rather than handed to a shell: the test is
// judging what the build derives, not exercising shell parsing, and running the
// tokens as a program keeps the test free of a shell-interpolation surface.
func primaryGitDerivation(t *testing.T, rhs string) []string {
	t.Helper()
	inner := strings.TrimSuffix(strings.TrimPrefix(rhs, "$(shell "), ")")
	if inner == rhs {
		t.Fatalf("BUILD_ID is not a $(shell ...) derivation: %q", rhs)
	}
	first := strings.TrimSpace(strings.Split(inner, "||")[0])
	var argv []string
	for _, field := range strings.Fields(first) {
		if strings.HasPrefix(field, "2>") || strings.HasPrefix(field, ">") {
			continue
		}
		argv = append(argv, field)
	}
	if len(argv) == 0 || argv[0] != "git" {
		t.Fatalf("BUILD_ID's primary derivation is not a git invocation: %q", first)
	}
	return argv
}

// (a) and (b) — BUILD_ID separates two commits standing in an ancestor
// relation, and the descendant's identity carries a component identifying that
// commit. The derivation under test is read out of the Makefile rather than
// restated here, so the test judges what the build actually does.
func TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation(t *testing.T) {
	_, rhs := makefileAssignment(t, "BUILD_ID")
	if rhs == "" {
		t.Fatal("the Makefile defines no BUILD_ID; there is no monotone build identity to derive")
	}
	argv := primaryGitDerivation(t, rhs)

	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	commit := func(name string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
		git("add", name)
		git("commit", "-q", "-m", name)
	}
	derive := func() string {
		t.Helper()
		out, err := exec.Command(argv[0], append([]string{"-C", dir}, argv[1:]...)...).Output()
		if err != nil {
			t.Fatalf("run BUILD_ID derivation %v: %v", argv, err)
		}
		return strings.TrimSpace(string(out))
	}

	git("init", "-q", "-b", "main")
	commit("a.txt")
	git("tag", "v9.9.9")
	ancestorID := derive()

	commit("b.txt")
	descendantID := derive()
	descendantSHA := revParseShort(t, dir)

	// (a) the descendant's identity differs from the ancestor's
	if descendantID == ancestorID {
		t.Errorf("BUILD_ID collapses two commits in an ancestor relation onto %q; "+
			"a build identity that cannot separate them cannot report lag", ancestorID)
	}
	// (b) it carries a component identifying the descendant commit. A build
	// timestamp would satisfy (a) alone while proving nothing about order.
	if !strings.Contains(descendantID, descendantSHA) {
		t.Errorf("BUILD_ID %q carries no component identifying commit %s", descendantID, descendantSHA)
	}
}

func revParseShort(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}
