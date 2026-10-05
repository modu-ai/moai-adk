package homestate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Acceptance coverage for the shipped hub list (SPEC-TODO-CARD-ISSUANCE-001
// REQ-TCI-020, AC-TCI-019 (d)(e)(f)). The loader is embedded DATA: it reads
// nothing but its own embedded file — no project tree, no SPEC directory, no
// reports path — so the product's hub chain cannot leak developer-machine
// state into a user project. The baseline comparison and the measurement
// re-run read the tracked baseline copy: they are TEST-side readers, the one
// place the `.moai` path appears.

// hubBaselineRelPath is the tracked baseline copy, relative to the repo root.
var hubBaselineRelPath = filepath.Join(".moai", "specs", "SPEC-TODO-CARD-ISSUANCE-001", "baseline", "hub-files.txt")

// repoRoot walks up from the current directory to the directory holding
// go.mod. Tests run with the package directory as cwd.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("walked past the repository root without finding go.mod")
		}
		dir = parent
	}
}

// hubBaselineEntries parses the tracked baseline copy into path→count pairs,
// order preserved. A missing baseline FAILS the test — the AC forbids
// counting a missing input as a skip (AC-TCI-019 (e)).
func hubBaselineEntries(t *testing.T, root string) []hubEntry {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, hubBaselineRelPath))
	if err != nil {
		t.Fatalf("read the tracked baseline copy: %v — the embedded list has nothing to answer to", err)
	}
	var out []hubEntry
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("baseline line %q is not <path> <count>", line)
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatalf("baseline count %q is not an integer", fields[1])
		}
		out = append(out, hubEntry{path: fields[0], count: count})
	}
	if len(out) == 0 {
		t.Fatal("the tracked baseline copy carries no entries")
	}
	return out
}

type hubEntry struct {
	path  string
	count int
}

// TestHubFileListFromBaseline — AC-TCI-019 (e): the embedded path set equals
// the tracked baseline copy's path set, and a missing baseline is a failure,
// never a skip.
func TestHubFileListFromBaseline(t *testing.T) {
	root := repoRoot(t)
	entries := hubBaselineEntries(t, root)

	want := make([]string, 0, len(entries))
	for _, e := range entries {
		want = append(want, e.path)
	}
	got := HubFiles()
	sort.Strings(want)
	sortedGot := append([]string(nil), got...)
	sort.Strings(sortedGot)
	if !reflect.DeepEqual(sortedGot, want) {
		t.Fatalf("embedded path set diverges from the tracked baseline copy:\n got: %v\nwant: %v", sortedGot, want)
	}
	_ = got
}

// TestHubFileLoaderReadsNoProjectPath — AC-TCI-019 (d), the static half: the
// loading file carries no `.moai` string literal, so the shipped loader
// cannot assemble a project-tree path without the literal the behavior test
// would catch anyway. Reads the loader source, not the data file — the data
// file's comments name no path either, and the check covers both.
func TestHubFileLoaderReadsNoProjectPath(t *testing.T) {
	root := repoRoot(t)
	for _, name := range []string{"hub_files.go", "hub_files.txt"} {
		raw, err := os.ReadFile(filepath.Join(root, "internal", "homestate", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(raw), ".moai") {
			t.Fatalf("%s carries the .moai literal — the loader must read only its embedded data", name)
		}
	}
}

// TestHubFileLoaderIgnoresProjectTree — AC-TCI-019 (d), the behavioral half:
// loading under an empty temp directory and under a temp directory seeded
// with marker files at the project-tree paths returns the same embedded list
// both times, with no marker path in it. A loader that assembled a project
// path from any input would surface the markers here.
func TestHubFileLoaderIgnoresProjectTree(t *testing.T) {
	embedded := HubFiles()
	if len(embedded) == 0 {
		t.Fatal("the embedded list is empty — the comparison would be vacuous")
	}

	empty := t.TempDir()
	markerRoot := t.TempDir()
	marker := func(rel string) {
		path := filepath.Join(markerRoot, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("marker\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	marker(filepath.Join(".moai", "specs", "SPEC-TODO-CARD-ISSUANCE-001", "baseline", "hub-files.txt"))
	marker(filepath.Join(".moai", "reports", "t999", "verdict.md"))

	var underEmpty, underMarker []string
	t.Chdir(empty)
	underEmpty = HubFiles()
	t.Chdir(markerRoot)
	underMarker = HubFiles()

	if !reflect.DeepEqual(underEmpty, embedded) {
		t.Fatalf("loading under an empty tree diverged:\n got: %v\nwant: %v", underEmpty, embedded)
	}
	if !reflect.DeepEqual(underMarker, embedded) {
		t.Fatalf("loading under a seeded tree diverged:\n got: %v\nwant: %v", underMarker, embedded)
	}
	for _, line := range underMarker {
		if strings.Contains(line, "t999") || strings.Contains(line, "marker") {
			t.Fatalf("the loaded list carried a project-tree marker: %s", line)
		}
	}
}

// The measurement constants are the tracked baseline header's recorded
// period and develop tip — the same values hub-files.txt names.
const (
	hubMeasureSince = "2026-08-13"
	hubMeasureUntil = "2026-10-05T12:00:00+09:00"
	hubMeasureTip   = "4964d0796"
)

// hubMeasureReporter is the reporting seam the measurement body writes
// through: skip, fail, and progress. *testing.T satisfies it; the policy
// test drives the body with a recording probe instead.
type hubMeasureReporter interface {
	Skip(args ...any)
	Fatal(args ...any)
	Logf(format string, args ...any)
}

// hubMeasureSkipDecision is the pure skip/fail/run decision (AC-TCI-019 (f)):
// a clone missing the recorded tip skips when CI is unset — a recorded gap,
// never a pass — and FAILS under CI, whose checkout carries full history
// (.github/workflows/ci.yml test job, fetch-depth: 0). A present tip always
// runs.
func hubMeasureSkipDecision(shaPresent, ciSet bool) string {
	switch {
	case shaPresent:
		return "run"
	case ciSet:
		return "fail"
	default:
		return "skip"
	}
}

// hubMeasureAll is the measurement body, one function receiving the seam so
// the skip policy test can drive the same three inputs through the same
// reporting (D32/D37): no outcome reaches the test output except through
// the seam.
func hubMeasureAll(rep hubMeasureReporter, root string, shaPresent, ciSet bool, measure func(path string) (int, error)) {
	switch hubMeasureSkipDecision(shaPresent, ciSet) {
	case "skip":
		rep.Skip("recorded develop tip", hubMeasureTip, "not in this clone and CI unset — the skip is a gap, not a pass")
		return
	case "fail":
		rep.Fatal("recorded develop tip", hubMeasureTip, "not in this clone under CI — the checkout carries full history (fetch-depth: 0)")
		return
	}
	for _, e := range hubBaselineEntriesReadonly(root) {
		got, err := measure(e.path)
		if err != nil {
			rep.Fatal("measure", e.path, ":", err.Error())
			return
		}
		if got != e.count {
			rep.Fatal(e.path, ": measured", strconv.Itoa(got), "commits, baseline records", strconv.Itoa(e.count))
			return
		}
		rep.Logf("%s: %d", e.path, got)
	}
	rep.Logf("all hub paths match their recorded counts")
}

// hubBaselineEntriesReadonly is hubBaselineEntries without the testing.T —
// the body reads the tracked baseline through it so the seam test can drive
// it against the real repo.
func hubBaselineEntriesReadonly(root string) []hubEntry {
	raw, err := os.ReadFile(filepath.Join(root, hubBaselineRelPath))
	if err != nil {
		return nil
	}
	var out []hubEntry
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		out = append(out, hubEntry{path: fields[0], count: count})
	}
	return out
}

// hubMeasureCommitsAt runs the recorded single-call measuring command —
// first-parent log over the recorded window at the recorded tip, one path —
// and counts the commits.
func hubMeasureCommitsAt(root, path string) (int, error) {
	out, err := exec.Command("git", "-C", root, "log", "--first-parent",
		"--since="+hubMeasureSince, "--until="+hubMeasureUntil,
		"--format=%H", hubMeasureTip, "--", path).Output()
	if err != nil {
		return 0, err
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return 0, nil
	}
	return strings.Count(text, "\n") + 1, nil
}

// hubTipPresent reports whether the clone carries the recorded tip object.
func hubTipPresent(root string) bool {
	err := exec.Command("git", "-C", root, "cat-file", "-t", hubMeasureTip).Run()
	return err == nil
}

// TestHubFileListMatchesMeasuringCommand — AC-TCI-019 (f): re-run the
// recorded single-call measuring command per path at the recorded develop
// tip and compare the counts. The body lives in hubMeasureAll — the skip
// policy test drives the same three inputs through the same seam and reads
// this wrapper's source to pin the routing (the wrapper calls the body
// exactly once, passes t, and skips/fails only through the seam).
func TestHubFileListMatchesMeasuringCommand(t *testing.T) {
	root := repoRoot(t)
	hubMeasureAll(t, root, hubTipPresent(root), os.Getenv("CI") != "", func(path string) (int, error) {
		return hubMeasureCommitsAt(root, path)
	})
}

// TestHubFileMeasurementSkipPolicy — AC-TCI-019 (f): the pure decision table
// AND the body's routing of the same three inputs through the seam, plus the
// wrapper source read ((i) one call carrying t, (ii) no direct skip/fail in
// the wrapper).
func TestHubFileMeasurementSkipPolicy(t *testing.T) {
	table := []struct {
		shaPresent, ciSet bool
		want              string
	}{
		{false, false, "skip"},
		{false, true, "fail"},
		{true, false, "run"},
		{true, true, "run"},
	}
	for _, row := range table {
		if got := hubMeasureSkipDecision(row.shaPresent, row.ciSet); got != row.want {
			t.Fatalf("decision(shaPresent=%v, ciSet=%v) = %q, want %q", row.shaPresent, row.ciSet, got, row.want)
		}
	}

	// The body routes the three inputs through the seam: a recording probe
	// stands in for *testing.T, so skip and fail are OBSERVED here without
	// ending this test.
	probe := &hubSeamProbe{}
	stub := func(string) (int, error) { return 1, nil }
	hubMeasureAll(probe, t.TempDir(), false, false, stub)
	if len(probe.skipped) != 1 || len(probe.failed) != 0 || len(probe.ran) != 0 {
		t.Fatalf("sha absent + CI unset: skip=%d fail=%d ran=%d, want skip=1 fail=0 ran=0", len(probe.skipped), len(probe.failed), len(probe.ran))
	}

	probe = &hubSeamProbe{}
	hubMeasureAll(probe, t.TempDir(), false, true, stub)
	if len(probe.failed) != 1 || len(probe.skipped) != 0 {
		t.Fatalf("sha absent + CI set: skip=%d fail=%d, want skip=0 fail=1", len(probe.skipped), len(probe.failed))
	}

	// sha present → the body runs. The baseline path is resolved against the
	// fixture root, so a throwaway baseline drives the comparison without
	// touching the tracked copy.
	fixture := t.TempDir()
	if err := os.MkdirAll(filepath.Join(fixture, filepath.Dir(hubBaselineRelPath)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, hubBaselineRelPath), []byte("some/path.txt 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	probe = &hubSeamProbe{}
	measured := 0
	hubMeasureAll(probe, fixture, true, false, func(string) (int, error) {
		measured++
		return 2, nil
	})
	if measured != 1 || len(probe.ran) == 0 || len(probe.failed) != 0 {
		t.Fatalf("sha present: measured=%d ran-lines=%d fail=%d, want 1 measure with run output", measured, len(probe.ran), len(probe.failed))
	}
	// The comparison is real: a count off by one fails through the seam.
	probe = &hubSeamProbe{}
	hubMeasureAll(probe, fixture, true, false, func(string) (int, error) { return 3, nil })
	if len(probe.failed) != 1 {
		t.Fatalf("count mismatch: fail=%d, want 1", len(probe.failed))
	}

	// Wrapper source read (D37): the wrapper calls the body exactly once,
	// passes t into it, and carries no direct t.Skip*/t.Fatal*/t.Error*.
	hubAssertWrapperRoutesThroughSeam(t)
}

// hubSeamProbe records the seam's outcomes instead of ending the test.
type hubSeamProbe struct {
	skipped, failed, ran []string
}

func (p *hubSeamProbe) Skip(args ...any)  { p.skipped = append(p.skipped, fmt.Sprint(args...)) }
func (p *hubSeamProbe) Fatal(args ...any) { p.failed = append(p.failed, fmt.Sprint(args...)) }
func (p *hubSeamProbe) Logf(format string, args ...any) {
	p.ran = append(p.ran, fmt.Sprintf(format, args...))
}

// hubAssertWrapperRoutesThroughSeam parses this test file and pins the
// wrapper's shape: exactly one hubMeasureAll call, carrying t, with zero
// direct t.Skip*/t.Fatal*/t.Error* selector calls in the wrapper body
// (iteration-6 audit D37; the adapter's own t-passing stays the accepted
// residual D42 named — this read covers the wrapper only).
func hubAssertWrapperRoutesThroughSeam(t *testing.T) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test file for the source read")
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, thisFile, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", thisFile, err)
	}
	var wrapper *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, isFn := decl.(*ast.FuncDecl)
		if isFn && fn.Name.Name == "TestHubFileListMatchesMeasuringCommand" {
			wrapper = fn
		}
	}
	if wrapper == nil {
		t.Fatal("wrapper TestHubFileListMatchesMeasuringCommand not found in source")
	}
	calls := 0
	passesT := false
	ast.Inspect(wrapper.Body, func(n ast.Node) bool {
		call, isCall := n.(*ast.CallExpr)
		if !isCall {
			return true
		}
		if ident, isIdent := call.Fun.(*ast.Ident); isIdent && ident.Name == "hubMeasureAll" {
			calls++
			for _, arg := range call.Args {
				if id, isID := arg.(*ast.Ident); isID && id.Name == "t" {
					passesT = true
				}
			}
		}
		return true
	})
	if calls != 1 {
		t.Fatalf("wrapper calls the measurement body %d times, want exactly 1", calls)
	}
	if !passesT {
		t.Fatal("wrapper does not pass t into the measurement body")
	}
	for _, bad := range []string{"Skip", "Skipf", "Fatal", "Fatalf", "Error", "Errorf"} {
		ast.Inspect(wrapper.Body, func(n ast.Node) bool {
			sel, isSel := n.(*ast.SelectorExpr)
			if !isSel {
				return true
			}
			if id, isID := sel.X.(*ast.Ident); isID && id.Name == "t" && sel.Sel.Name == bad {
				t.Fatalf("wrapper carries a direct t.%s — skip and fail must route through the seam", bad)
			}
			return true
		})
	}
}
