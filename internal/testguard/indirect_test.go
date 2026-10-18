package testguard

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// indirectProd is a faithful miniature of the production shape that hid the
// t1816 blind spot: seam variables default to functions that exec
// `git config --global`, a materializer reaches them through an apply helper,
// a command struct reaches the materializer by function reference (the cobra
// RunE shape), and a variable of another package is wired to the materializer
// by an assignment (the worktree.WorktreeCreator shape).
const indirectProd = `package fx

import (
	"io"
	"os/exec"

	wtcmd "example.invalid/fx/worktree"
)

var (
	safeDirAdd  = safeDirAddReal
	globalGet   = globalGetReal
	worktreeAdd = worktreeAddReal
)

func safeDirAddReal(path string) error {
	return exec.Command("git", "config", "--global", "--add", "safe.directory", path).Run()
}

func globalGetReal(key string) string {
	out, _ := exec.Command("git", "config", "--global", "--get", key).Output()
	return string(out)
}

func worktreeAddReal(dest string) error {
	return exec.Command("git", "worktree", "add", dest).Run()
}

func applyConfig(wt string) {
	_ = safeDirAdd(wt)
	_ = globalGet("user.name")
}

func materialize(name string, out io.Writer) (string, error) {
	_ = worktreeAdd(name)
	applyConfig(name)
	return name, nil
}

type command struct{ RunE func() error }

func newWebCmd() *command { return &command{RunE: runWeb} }

func runWeb() error {
	_, err := materialize("web", nil)
	return err
}

func init() { wtcmd.WorktreeCreator = materialize }
`

// indirectScan writes prod and the named test files into one package directory
// and returns the violations of the new kind that ScanTree reports for it.
func indirectScan(t *testing.T, prod string, tests map[string]string) []Violation {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "internal", "fx", "prod.go"), prod)
	for name, src := range tests {
		writeFile(t, filepath.Join(root, "internal", "fx", name), src)
	}
	res, err := ScanTree(root, []string{"internal"})
	if err != nil {
		t.Fatalf("ScanTree: %v", err)
	}
	var out []Violation
	for _, v := range res.Violations {
		if v.Kind == KindIndirectGlobalGitConfig {
			out = append(out, v)
		}
	}
	return out
}

// wantFlagged fails unless vs holds exactly one violation, on the declaration
// line of the named test function in src.
func wantFlagged(t *testing.T, vs []Violation, src, fn string) {
	t.Helper()
	want := lineOf(t, src, "func "+fn+"(")
	if len(vs) != 1 || vs[0].Line != want || !strings.Contains(vs[0].Text, fn) {
		t.Fatalf("violations = %+v; want exactly one %q at line %d (func %s)", vs, KindIndirectGlobalGitConfig, want, fn)
	}
}

const runWebTestSrc = `package fx

import "testing"

func TestRunWeb_AutoEntryOn(t *testing.T) {
	cmd := newWebCmd()
	if err := cmd.RunE(); err != nil {
		t.Fatal(err)
	}
}
`

const creatorTestSrc = `package fx

import (
	"testing"

	wtcmd "example.invalid/fx/worktree"
)

func TestWorktreeNew_WiresTheSharedMaterializer(t *testing.T) {
	if _, err := wtcmd.WorktreeCreator("probe-card", nil); err != nil {
		t.Fatal(err)
	}
}
`

// swapHelperSrc only assigns the seams, which is what the scanner reads.
const swapHelperSrc = `package fx

import "testing"

func swapGitSeams(t *testing.T) {
	t.Helper()
	safeDirAdd = func(string) error { return nil }
	globalGet = func(string) string { return "" }
}
`

// TestIndirect_FlagsRunWebShape: a test that reaches the sinks only through a
// function reference stored in a struct literal is reported.
func TestIndirect_FlagsRunWebShape(t *testing.T) {
	vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": runWebTestSrc})
	wantFlagged(t, vs, runWebTestSrc, "TestRunWeb_AutoEntryOn")
}

// TestIndirect_FlagsCrossPackageCreatorShape: a test that calls a variable of
// another package, wired to the materializer by an assignment, is reported.
func TestIndirect_FlagsCrossPackageCreatorShape(t *testing.T) {
	vs := indirectScan(t, indirectProd, map[string]string{"creator_test.go": creatorTestSrc})
	wantFlagged(t, vs, creatorTestSrc, "TestWorktreeNew_WiresTheSharedMaterializer")
}

// TestIndirect_NotFlaggedWhenBothSeamsSwappedInTheTest: swapping both global
// seams inside the test body isolates every path to the sinks.
func TestIndirect_NotFlaggedWhenBothSeamsSwappedInTheTest(t *testing.T) {
	src := strings.Replace(runWebTestSrc, "cmd := newWebCmd()",
		"safeDirAdd = func(string) error { return nil }\n\tglobalGet = func(string) string { return \"\" }\n\tcmd := newWebCmd()", 1)
	if vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": src}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: both seams are swapped in the test", vs)
	}
}

// TestIndirect_NotFlaggedWhenAReachableHelperSwapsBothSeams: the swap may sit in
// a helper in another test file, as swapSessionWorktreeSeams does.
func TestIndirect_NotFlaggedWhenAReachableHelperSwapsBothSeams(t *testing.T) {
	src := strings.Replace(creatorTestSrc, "if _, err", "swapGitSeams(t)\n\tif _, err", 1)
	vs := indirectScan(t, indirectProd, map[string]string{"creator_test.go": src, "helper_test.go": swapHelperSrc})
	if len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: a reachable helper swaps both seams", vs)
	}
}

// TestIndirect_StillFlaggedWhenOnlyOneSeamIsSwapped: the global-get path stays
// open when only the safe.directory add is swapped.
func TestIndirect_StillFlaggedWhenOnlyOneSeamIsSwapped(t *testing.T) {
	src := strings.Replace(runWebTestSrc, "cmd := newWebCmd()",
		"safeDirAdd = func(string) error { return nil }\n\tcmd := newWebCmd()", 1)
	vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": src})
	wantFlagged(t, vs, src, "TestRunWeb_AutoEntryOn")
}

// TestIndirect_HelperInAnotherTestFileDoesNotCoverAnUnrelatedTest: a swap
// helper the test never reaches does not isolate it.
func TestIndirect_HelperInAnotherTestFileDoesNotCoverAnUnrelatedTest(t *testing.T) {
	vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": runWebTestSrc, "helper_test.go": swapHelperSrc})
	wantFlagged(t, vs, runWebTestSrc, "TestRunWeb_AutoEntryOn")
}

// TestIndirect_AllowMarkerExemptsAndNeedsAReason: the marker on the line above
// the declaration exempts the test; a bare marker exempts nothing.
func TestIndirect_AllowMarkerExemptsAndNeedsAReason(t *testing.T) {
	ok := strings.Replace(runWebTestSrc, "func Test",
		"// home-guard:allow the pinned TestMain config is the subject under test\nfunc Test", 1)
	if vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": ok}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want the marked test exempt", vs)
	}
	bare := strings.Replace(runWebTestSrc, "func Test", "// home-guard:allow\nfunc Test", 1)
	vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": bare})
	wantFlagged(t, vs, bare, "TestRunWeb_AutoEntryOn")
}

// TestIndirect_RedirectHomeAndGitConfigEnvIsolate: redirecting the global git
// config (RedirectHome, or setting GIT_CONFIG_GLOBAL) isolates the test.
func TestIndirect_RedirectHomeAndGitConfigEnvIsolate(t *testing.T) {
	for name, line := range map[string]string{
		"redirect_home": "testguard.RedirectHome(t)",
		"config_env":    "t.Setenv(\"GIT_CONFIG_GLOBAL\", t.TempDir())",
	} {
		t.Run(name, func(t *testing.T) {
			src := strings.Replace(runWebTestSrc, "cmd := newWebCmd()", line+"\n\tcmd := newWebCmd()", 1)
			src = strings.Replace(src, "import \"testing\"", "import (\n\t\"testing\"\n\n\t\"example.invalid/fx/testguard\"\n)\n\nvar _ = testguard.RedirectHome", 1)
			if vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": src}); len(vs) != 0 {
				t.Fatalf("violations = %+v; want none: %s isolates the global config", vs, line)
			}
		})
	}
}

// TestIndirect_CommentAndStringMentionsAreNotSinks: the sink spelled in a
// comment or inside a string literal is not an exec call.
func TestIndirect_CommentAndStringMentionsAreNotSinks(t *testing.T) {
	prod := `package fx

import "os/exec"

// safeDirAddReal would run exec.Command("git", "config", "--global", "--add").
func safeDirAddReal(path string) error {
	_ = "exec.Command(\"git\", \"config\", \"--global\")"
	return exec.Command("git", "status").Run()
}

var safeDirAdd = safeDirAddReal
`
	test := "package fx\n\nimport \"testing\"\n\nfunc TestUsesIt(t *testing.T) { _ = safeDirAdd(\"x\") }\n"
	if vs := indirectScan(t, prod, map[string]string{"x_test.go": test}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: the only mentions are a comment and a string", vs)
	}
}

// TestIndirect_LocalNamesAreNotReferences: a local variable that reuses a seam
// name is not a reference to the seam.
func TestIndirect_LocalNamesAreNotReferences(t *testing.T) {
	test := "package fx\n\nimport \"testing\"\n\n" +
		"func TestLocalShadow(t *testing.T) {\n\tglobalGet := func(string) string { return \"\" }\n\t_ = globalGet(\"k\")\n}\n"
	if vs := indirectScan(t, indirectProd, map[string]string{"x_test.go": test}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: globalGet here is a local value", vs)
	}
}

// TestIndirect_KeyHasNoLineNumber: the key of the new kind follows the Keys
// format, so it survives unrelated edits above the test.
func TestIndirect_KeyHasNoLineNumber(t *testing.T) {
	vs := indirectScan(t, indirectProd, map[string]string{"web_test.go": runWebTestSrc})
	shifted := indirectScan(t, indirectProd, map[string]string{"web_test.go": strings.Replace(runWebTestSrc, "func Test", "var _ = 1\n\nfunc Test", 1)})
	if len(vs) != 1 || len(shifted) != 1 {
		t.Fatalf("violations = %+v / %+v; want one each", vs, shifted)
	}
	a, b := Keys(vs)[0], Keys(shifted)[0]
	want := "internal/fx/web_test.go\t" + KindIndirectGlobalGitConfig + "\tfunc TestRunWeb_AutoEntryOn(t *testing.T) {\t0"
	if a != want || a != b {
		t.Fatalf("keys: got %q / %q; want both %q", a, b, want)
	}
}

// swapArgsHelperSrc has the swapM7Seams shape: a seam is replaced only when its
// field is passed, and the cleanup restores every seam.
const swapArgsHelperSrc = `package fx

import "testing"

type seamArgs struct {
	safeDirAdd func(string) error
	globalGet  func(string) string
}

func swapSeams(t *testing.T, s seamArgs) {
	t.Helper()
	orig := seamArgs{safeDirAdd: safeDirAdd, globalGet: globalGet}
	if s.safeDirAdd != nil {
		safeDirAdd = s.safeDirAdd
	}
	if s.globalGet != nil {
		globalGet = s.globalGet
	}
	t.Cleanup(func() {
		safeDirAdd = orig.safeDirAdd
		globalGet = orig.globalGet
	})
}
`

const swapArgsOneTestSrc = `package fx

import "testing"

func TestPassesOneField(t *testing.T) {
	swapSeams(t, seamArgs{safeDirAdd: func(string) error { return nil }})
	_, _ = materialize("x", nil)
}
`

// TestIndirect_ParameterizedSwapCountsOnlyThePassedFields: a helper that swaps a
// seam only when its field is passed isolates just the fields the test passes,
// and its cleanup restoring every seam isolates nothing.
func TestIndirect_ParameterizedSwapCountsOnlyThePassedFields(t *testing.T) {
	vs := indirectScan(t, indirectProd, map[string]string{"x_test.go": swapArgsOneTestSrc, "helper_test.go": swapArgsHelperSrc})
	wantFlagged(t, vs, swapArgsOneTestSrc, "TestPassesOneField")

	both := strings.Replace(swapArgsOneTestSrc, "seamArgs{safeDirAdd: func(string) error { return nil }}",
		"seamArgs{safeDirAdd: func(string) error { return nil }, globalGet: func(string) string { return \"\" }}", 1)
	if vs := indirectScan(t, indirectProd, map[string]string{"x_test.go": both, "helper_test.go": swapArgsHelperSrc}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: both fields are passed", vs)
	}
}

// TestIndirect_RestoreInCleanupIsNotASwap: a test whose only assignments sit in
// a Cleanup closure swaps nothing.
func TestIndirect_RestoreInCleanupIsNotASwap(t *testing.T) {
	src := "package fx\n\nimport \"testing\"\n\nfunc TestOnlyRestores(t *testing.T) {\n" +
		"\torig := safeDirAdd\n\tt.Cleanup(func() { safeDirAdd = orig; globalGet = globalGet })\n\t_, _ = materialize(\"x\", nil)\n}\n"
	wantFlagged(t, indirectScan(t, indirectProd, map[string]string{"x_test.go": src}), src, "TestOnlyRestores")
}

// TestIndirect_RawStringGlobalFlagIsASink: a package whose only "--global" is a
// raw string literal is still scanned.
func TestIndirect_RawStringGlobalFlagIsASink(t *testing.T) {
	prod := "package fx\n\nimport \"os/exec\"\n\nfunc run() error { return exec.Command(\"git\", \"config\", `--global`, \"--list\").Run() }\n"
	test := "package fx\n\nimport \"testing\"\n\nfunc TestRuns(t *testing.T) { _ = run() }\n"
	wantFlagged(t, indirectScan(t, prod, map[string]string{"x_test.go": test}), test, "TestRuns")
}

// restoreFixture runs body inside a test that reaches materialize, which
// reaches both global seams of indirectProd.
func restoreFixture(body string) string {
	return "package fx\n\nimport \"testing\"\n\nfunc TestRestoreShape(t *testing.T) {\n" + body + "\n\t_, _ = materialize(\"x\", nil)\n}\n"
}

// TestIndirect_RestoresOutsideCleanupAreNotSwaps (F1): a deferred restore, a
// self-assignment, and a restore of a captured original replace nothing.
func TestIndirect_RestoresOutsideCleanupAreNotSwaps(t *testing.T) {
	cases := map[string]string{
		"defer_restore": "\toldAdd, oldGet := safeDirAdd, globalGet\n\tdefer func() { safeDirAdd = oldAdd; globalGet = oldGet }()",
		"self_assign":   "\tsafeDirAdd = safeDirAdd\n\tglobalGet = globalGet",
		"captured":      "\toldAdd := safeDirAdd\n\toldGet := globalGet\n\tsafeDirAdd = oldAdd\n\tglobalGet = oldGet",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			src := restoreFixture(body)
			wantFlagged(t, indirectScan(t, indirectProd, map[string]string{"x_test.go": src}), src, "TestRestoreShape")
		})
	}
	// A real replacement with a deferred restore is still a swap.
	ok := restoreFixture("\toldAdd, oldGet := safeDirAdd, globalGet\n\tdefer func() { safeDirAdd = oldAdd; globalGet = oldGet }()\n" +
		"\tsafeDirAdd = func(string) error { return nil }\n\tglobalGet = func(string) string { return \"\" }")
	if vs := indirectScan(t, indirectProd, map[string]string{"x_test.go": ok}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: both seams are replaced", vs)
	}
}

// TestIndirect_FieldMustReachTheHelperCall (F2): a literal that sets a field is
// credited only when it is the helper call's argument, directly or through a
// local variable. An unrelated literal elsewhere in the test is not.
func TestIndirect_FieldMustReachTheHelperCall(t *testing.T) {
	stub := "func(string) error { return nil }"
	get := "func(string) string { return \"\" }"
	both := "seamArgs{safeDirAdd: " + stub + ", globalGet: " + get + "}"
	scan := func(body string) []Violation {
		return indirectScan(t, indirectProd, map[string]string{"x_test.go": restoreFixture(body), "helper_test.go": swapArgsHelperSrc})
	}
	unrelated := scan("\tswapSeams(t, seamArgs{})\n\t_ = seamArgs{safeDirAdd: " + stub + ", globalGet: " + get + "}")
	wantFlagged(t, unrelated, restoreFixture("\tswapSeams(t, seamArgs{})\n\t_ = seamArgs{safeDirAdd: "+stub+", globalGet: "+get+"}"), "TestRestoreShape")
	if vs := scan("\tswapSeams(t, " + both + ")"); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: the literal is the call argument", vs)
	}
	if vs := scan("\targs := " + both + "\n\tswapSeams(t, args)"); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: the literal reaches the call through a local", vs)
	}
}

// TestIndirect_IsolationWalkHonorsTheDispatchBound (F3): a helper that only a
// dispatch table references does not isolate a test that never calls it.
func TestIndirect_IsolationWalkHonorsTheDispatchBound(t *testing.T) {
	var b strings.Builder
	b.WriteString("package fx\n\nimport \"testing\"\n\nfunc isolateAll(t *testing.T) {\n\tsafeDirAdd = func(string) error { return nil }\n\tglobalGet = func(string) string { return \"\" }\n}\n\n")
	b.WriteString("var table = []func(*testing.T){\n\tisolateAll,\n")
	for i := 0; i < dispatchFanOut+1; i++ {
		fmt.Fprintf(&b, "\tfiller%d,\n", i)
	}
	b.WriteString("}\n\n")
	for i := 0; i < dispatchFanOut+1; i++ {
		fmt.Fprintf(&b, "func filler%d(*testing.T) {}\n", i)
	}
	b.WriteString("\nfunc TestUsesTable(t *testing.T) {\n\t_ = table\n\t_, _ = materialize(\"x\", nil)\n}\n")
	src := b.String()
	wantFlagged(t, indirectScan(t, indirectProd, map[string]string{"x_test.go": src}), src, "TestUsesTable")
}

// TestIndirect_EscapedGlobalFlagIsASink (F4): the flag spelled with escapes
// still holds the value --global, so the package is scanned.
func TestIndirect_EscapedGlobalFlagIsASink(t *testing.T) {
	test := "package fx\n\nimport \"testing\"\n\nfunc TestRuns(t *testing.T) { _ = run() }\n"
	for name, lit := range map[string]string{"hex": `"\x2d-global"`, "all_escaped": `"\x2d\x2d\x67lobal"`} {
		t.Run(name, func(t *testing.T) {
			prod := "package fx\n\nimport \"os/exec\"\n\nfunc run() error { return exec.Command(\"git\", \"config\", " + lit + ", \"--list\").Run() }\n"
			wantFlagged(t, indirectScan(t, prod, map[string]string{"x_test.go": test}), test, "TestRuns")
		})
	}
}

// TestIndirect_OnlyGitIsASink (F5): another program given the same words is not
// a git config write; a path whose base is git is.
func TestIndirect_OnlyGitIsASink(t *testing.T) {
	test := "package fx\n\nimport \"testing\"\n\nfunc TestRuns(t *testing.T) { _ = run() }\n"
	prodFor := func(call string) string {
		return "package fx\n\nimport (\n\t\"context\"\n\t\"os/exec\"\n)\n\nvar _ = context.Background\n\nfunc run() error { return " + call + ".Run() }\n"
	}
	if vs := indirectScan(t, prodFor("exec.Command(\"echo\", \"config\", \"--global\")"), map[string]string{"x_test.go": test}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: echo is not git", vs)
	}
	if vs := indirectScan(t, prodFor("exec.CommandContext(context.Background(), \"echo\", \"config\", \"--global\")"), map[string]string{"x_test.go": test}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: echo is not git (CommandContext)", vs)
	}
	wantFlagged(t, indirectScan(t, prodFor("exec.Command(\"/usr/bin/git\", \"config\", \"--global\")"), map[string]string{"x_test.go": test}), test, "TestRuns")
	wantFlagged(t, indirectScan(t, prodFor("exec.CommandContext(context.Background(), \"git\", \"config\", \"--global\")"), map[string]string{"x_test.go": test}), test, "TestRuns")
}

// dispatchProd builds a command tree of n handlers; the first one reaches the
// global git config sink and the others do nothing.
func dispatchProd(n int) string {
	var b strings.Builder
	b.WriteString("package fx\n\nimport \"os/exec\"\n\n")
	b.WriteString("func handler0() error { return exec.Command(\"git\", \"config\", \"--global\", \"--list\").Run() }\n")
	for i := 1; i < n; i++ {
		fmt.Fprintf(&b, "func handler%d() error { return nil }\n", i)
	}
	b.WriteString("\nfunc newTree() []func() error {\n\treturn []func() error{\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "\t\thandler%d,\n", i)
	}
	b.WriteString("\t}\n}\n")
	return b.String()
}

const dispatchTestSrc = "package fx\n\nimport \"testing\"\n\nfunc TestRunsOneBranch(t *testing.T) { _ = newTree()[3]() }\n"

// TestIndirect_DispatchTableIsNotFollowed: a node that registers more than
// dispatchFanOut handlers is a dispatch table, and a test that runs it runs one
// branch. A tree at the bound is still followed.
func TestIndirect_DispatchTableIsNotFollowed(t *testing.T) {
	small := indirectScan(t, dispatchProd(dispatchFanOut), map[string]string{"x_test.go": dispatchTestSrc})
	wantFlagged(t, small, dispatchTestSrc, "TestRunsOneBranch")
	if vs := indirectScan(t, dispatchProd(dispatchFanOut+1), map[string]string{"x_test.go": dispatchTestSrc}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: newTree references %d handlers, a dispatch table", vs, dispatchFanOut+1)
	}
}

// TestIndirect_UnqualifiedRedirectHomeIsolates: this package's own tests call
// RedirectHome without a qualifier.
func TestIndirect_UnqualifiedRedirectHomeIsolates(t *testing.T) {
	prod := "package fx\n\nimport \"os/exec\"\n\nfunc run() error { return exec.Command(\"git\", \"config\", \"--global\", \"--list\").Run() }\n"
	test := "package fx\n\nimport \"testing\"\n\nfunc TestRuns(t *testing.T) {\n\tRedirectHome(t)\n\t_ = run()\n}\n"
	if vs := indirectScan(t, prod, map[string]string{"x_test.go": test}); len(vs) != 0 {
		t.Fatalf("violations = %+v; want none: RedirectHome precedes the call", vs)
	}
	bare := strings.Replace(test, "\tRedirectHome(t)\n", "", 1)
	wantFlagged(t, indirectScan(t, prod, map[string]string{"x_test.go": bare}), bare, "TestRuns")
}
