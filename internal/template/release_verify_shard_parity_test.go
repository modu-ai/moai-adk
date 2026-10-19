// release_verify_shard_parity_test.go: drift and exact-cover guard for the
// Release Verify shard matrix (SPEC-CI-RELEASE-VERIFY-SHARD-001 M3, card t1892).
//
// release-pr-multi-os.yml runs the race suite as the same six shards ci.yml's
// Race Test legs use (internal/cli 1-4 of 4, the rosterguard package, and every
// other package once), on all three OS values. The selectors are copies, so
// nothing but this test keeps them equal, and nothing else notices when a new
// top-level internal/cli test name falls between the four first-letter selectors
// and is silently run by neither workflow.
//
// TestReleaseVerifyShardParity asserts, over the repository files as they are:
//
//   - REQ-RVS-003: every shard's selector and package, and the remainder
//     exclusion list, are equal between ci.yml and the release workflow, and the
//     exclusion list is exactly the packages the other shards name.
//   - REQ-RVS-015: every top-level internal/cli test name matches exactly one of
//     the four selectors (zero or two is a failure naming the test), and every
//     rosterguard name matches the rosterguard selector.
//   - REQ-RVS-016: the go tool lists no fuzz target or example for either package
//     (no ^Test selector would run it).
//   - AC-RVS-001/007/013: the matrix has no exclude key, exactly one include entry
//     per shard keyed by shard alone, expands to exactly os x shard legs, gives
//     every leg its own artifact name, and runs go vet on exactly one leg per OS;
//     the shard and OS lists equal the constants in scripts/release.sh.
//
// The names come from the go tool, not from parsing Go source:
// `go test -race -list '.*'` (the shards' own build conditions; -race also sets
// the race build tag, so a name that exists only in a //go:build race file is
// listed). No Go discovery rule is re-implemented here. A name that exists only
// under another OS's build tags is not listed on the running OS; the three-OS
// release matrix covers that limit.
//
// A check that has never been seen failing has proven nothing, so the test also
// runs a mutation table over in-memory copies of its inputs (mutants i-xi, each
// must be detected, controls C1 and C2 must pass) and the paired plain/-race
// fixture run for mutant (x). The ci.yml go-code filter already lists both
// workflow files, so this test runs whenever either file changes.
//
// Sentinel on failure: RELEASE_VERIFY_SHARD_DRIFT
package template_test

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	rvsCIWorkflowRel      = ".github/workflows/ci.yml"
	rvsReleaseWorkflowRel = ".github/workflows/release-pr-multi-os.yml"
	rvsReleaseScriptRel   = "scripts/release.sh"
	rvsMatrixJob          = "full-matrix-test"
	rvsTestStep           = "Run tests with race detector"
	rvsUploadStep         = "Upload test event stream"
	rvsCLIPkg             = "./internal/cli/"
	rvsRosterPkg          = "./internal/harness/rosterguard/"
	rvsRemainderShard     = "remainder"
	rvsRosterShard        = "rosterguard"
	rvsWantShards         = 6
	rvsWantOSes           = 3
	rvsWantCLISelectors   = 4
	// rvsListTimeout bounds one nested listing from outside; it sits under the
	// default 10m test timeout so a stuck listing reports itself.
	rvsListTimeout = 9 * time.Minute
)

// Fixture for mutant (x): one clean-base package plus a race-only file. The base
// has TestReal, BenchmarkB and Fuzzhelper only (no bare Test/Fuzz/Example, no
// Example function, so go vet's tests analyzer passes with no -vet=off); the
// race-only file defines FuzzRaceOnly. A plain listing prints TestReal and
// BenchmarkB; `-race` adds FuzzRaceOnly. Mirrors
// .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/repro/listing-fixture/race-pair/.
//
// The fixture source is built from interpreted strings, not raw strings, so no
// line of this file begins with `func Test...`/`func Fuzz...` and a line-based
// scan of this package cannot mistake the fixture for a real test.
const (
	rvsFixtureGoMod = "module gl\n\ngo 1.22\n"

	rvsFixtureBase = "package gl\n\n" +
		"import \"testing\"\n\n" +
		"func TestReal(t *testing.T)      {}\n" +
		"func BenchmarkB(b *testing.B)    {}\n" +
		"func Fuzzhelper(s string) string { return s }\n"

	rvsFixtureRaceOnly = "//go:build race\n\n" +
		"package gl\n\n" +
		"import \"testing\"\n\n" +
		"func FuzzRaceOnly(f *testing.F) { f.Add(1); f.Fuzz(func(t *testing.T, n int) {}) }\n"
)

var (
	// rvsRunSelRe reads `-run '<selector>' <package>` from a ci.yml race step.
	rvsRunSelRe = regexp.MustCompile(`-run '([^']*)' (\./[^\s>]+)`)
	// rvsExclRe reads the -e '<package>' arguments of one grep line.
	rvsExclRe = regexp.MustCompile(`-e '([^']*)'`)
	// rvsIdentRe is the shape of a name `go test -list` prints; the trailing
	// `ok  <pkg>  <time>` line has spaces and is not one.
	rvsIdentRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	// rvsMatrixRefRe matches ${{ matrix.<key> }} in an Actions expression.
	rvsMatrixRefRe = regexp.MustCompile(`\$\{\{\s*matrix\.([A-Za-z_][A-Za-z0-9_-]*)\s*\}\}`)
	rvsModuleRe    = regexp.MustCompile(`(?m)^module\s+(\S+)`)
)

// ---------------------------------------------------------------------------
// Workflow models
// ---------------------------------------------------------------------------

type rvsStepDoc struct {
	Name string         `yaml:"name"`
	Run  string         `yaml:"run"`
	With map[string]any `yaml:"with"`
}

// rvsCIDoc deliberately omits `strategy`: other ci.yml jobs may carry a matrix
// that is an expression string, which would not decode into a map.
type rvsCIDoc struct {
	Jobs map[string]struct {
		Steps []rvsStepDoc `yaml:"steps"`
	} `yaml:"jobs"`
}

type rvsReleaseDoc struct {
	Jobs map[string]struct {
		Strategy struct {
			Matrix map[string]any `yaml:"matrix"`
		} `yaml:"strategy"`
		Steps []rvsStepDoc `yaml:"steps"`
	} `yaml:"jobs"`
}

type rvsShardCfg struct {
	Pkg string
	Run string
}

// rvsCI is what ci.yml says about the six race shards.
type rvsCI struct {
	Shards        map[string]rvsShardCfg
	RemainderExcl []string
}

// rvsRelease is what the release workflow's matrix job says.
type rvsRelease struct {
	Matrix       map[string]any
	TestRun      string
	ArtifactName string
}

func (r rvsRelease) osValues() []string    { return rvsStrings(r.Matrix["os"]) }
func (r rvsRelease) shardValues() []string { return rvsStrings(r.Matrix["shard"]) }

// shardCfgs reads pkg and run from the include entries, first entry per shard.
func (r rvsRelease) shardCfgs() map[string]rvsShardCfg {
	out := map[string]rvsShardCfg{}
	for _, inc := range rvsMaps(r.Matrix["include"]) {
		shard, ok := inc["shard"].(string)
		if !ok {
			continue
		}
		if _, dup := out[shard]; dup {
			continue
		}
		cfg := rvsShardCfg{}
		if v, ok := inc["pkg"].(string); ok {
			cfg.Pkg = v
		}
		if v, ok := inc["run"].(string); ok {
			cfg.Run = v
		}
		out[shard] = cfg
	}
	return out
}

func rvsStrings(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, fmt.Sprint(it))
	}
	return out
}

func rvsMaps(v any) []map[string]any {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// rvsGrepVExclusions returns the -e '<package>' arguments of the first
// `grep -v -x` line of a step body: the remainder shard's exclusion list. Only
// that line is read, not the whole body, so a later guard line cannot stand in.
func rvsGrepVExclusions(run string) []string {
	for _, line := range strings.Split(run, "\n") {
		if !strings.Contains(line, "grep -v -x") {
			continue
		}
		var out []string
		for _, m := range rvsExclRe.FindAllStringSubmatch(line, -1) {
			out = append(out, m[1])
		}
		return out
	}
	return nil
}

func rvsParseCI(t *testing.T, text string) rvsCI {
	t.Helper()
	var doc rvsCIDoc
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		t.Fatalf("%s: parse: %v", rvsCIWorkflowRel, err)
	}
	out := rvsCI{Shards: map[string]rvsShardCfg{}}
	for id, job := range doc.Jobs {
		shard, ok := strings.CutPrefix(id, "test-race-")
		if !ok {
			continue
		}
		for _, st := range job.Steps {
			if !strings.Contains(st.Run, "go test") || !strings.Contains(st.Run, "-race") {
				continue
			}
			if shard == rvsRemainderShard {
				out.Shards[shard] = rvsShardCfg{Pkg: rvsRemainderShard}
				out.RemainderExcl = rvsGrepVExclusions(st.Run)
				continue
			}
			if m := rvsRunSelRe.FindStringSubmatch(st.Run); m != nil {
				out.Shards[shard] = rvsShardCfg{Run: m[1], Pkg: m[2]}
			}
		}
	}
	return out
}

func rvsParseRelease(t *testing.T, text string) rvsRelease {
	t.Helper()
	var doc rvsReleaseDoc
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		t.Fatalf("%s: parse: %v", rvsReleaseWorkflowRel, err)
	}
	job, ok := doc.Jobs[rvsMatrixJob]
	if !ok {
		t.Fatalf("%s: job %q not found (the gate reads it through needs.%s.result)", rvsReleaseWorkflowRel, rvsMatrixJob, rvsMatrixJob)
	}
	rel := rvsRelease{Matrix: job.Strategy.Matrix}
	for _, st := range job.Steps {
		switch st.Name {
		case rvsTestStep:
			rel.TestRun = st.Run
		case rvsUploadStep:
			if n, ok := st.With["name"].(string); ok {
				rel.ArtifactName = n
			}
		}
	}
	return rel
}

// rvsShellArray reads NAME=(a b c) from a shell script.
func rvsShellArray(t *testing.T, script, name string) []string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(name) + `=\(([^)]*)\)`)
	m := re.FindStringSubmatch(script)
	if m == nil {
		t.Fatalf("%s: no %s=(...) array found", rvsReleaseScriptRel, name)
	}
	return strings.Fields(m[1])
}

// ---------------------------------------------------------------------------
// Matrix expansion (GitHub Actions include/exclude merge rule)
// ---------------------------------------------------------------------------

// rvsExpand returns the final leg list of a matrix with include and exclude
// applied, not the plain cross product. Every non-include, non-exclude key is an
// original dimension. Exclude entries remove the combinations they match
// entirely. An include entry is merged into every combination whose original
// dimension values it does not overwrite; one that fits no combination becomes
// a leg of its own.
func rvsExpand(matrix map[string]any) []map[string]string {
	dims := map[string][]string{}
	for k, v := range matrix {
		if k == "include" || k == "exclude" {
			continue
		}
		dims[k] = rvsStrings(v)
	}
	var combos []map[string]string
	if len(dims) > 0 {
		combos = []map[string]string{{}}
		for _, k := range slices.Sorted(maps.Keys(dims)) {
			var next []map[string]string
			for _, c := range combos {
				for _, v := range dims[k] {
					n := maps.Clone(c)
					n[k] = v
					next = append(next, n)
				}
			}
			combos = next
		}
	}
	for _, ex := range rvsMaps(matrix["exclude"]) {
		var kept []map[string]string
		for _, c := range combos {
			match := true
			for k, v := range ex {
				if c[k] != fmt.Sprint(v) {
					match = false
					break
				}
			}
			if !match {
				kept = append(kept, c)
			}
		}
		combos = kept
	}
	for _, inc := range rvsMaps(matrix["include"]) {
		added := false
		for _, c := range combos {
			fits := true
			for k, v := range inc {
				if _, isDim := dims[k]; isDim && c[k] != fmt.Sprint(v) {
					fits = false
					break
				}
			}
			if !fits {
				continue
			}
			for k, v := range inc {
				c[k] = fmt.Sprint(v)
			}
			added = true
		}
		if !added {
			leg := map[string]string{}
			for k, v := range inc {
				leg[k] = fmt.Sprint(v)
			}
			combos = append(combos, leg)
		}
	}
	return combos
}

// rvsSubstMatrix replaces ${{ matrix.<key> }} with the leg's value.
func rvsSubstMatrix(tmpl string, leg map[string]string) string {
	return rvsMatrixRefRe.ReplaceAllStringFunc(tmpl, func(ref string) string {
		return leg[rvsMatrixRefRe.FindStringSubmatch(ref)[1]]
	})
}

// ---------------------------------------------------------------------------
// Test-name listing (the go tool is the source, never Go source parsing)
// ---------------------------------------------------------------------------

// rvsNames is a classified `go test -list` output. Benchmark names are dropped
// (benchmarks do not run without -bench); the trailing ok/FAIL line is not a
// name. Test feeds the cover check, Other (Fuzz..., Example..., anything else
// the tool enumerates) feeds the discovery guard.
type rvsNames struct {
	Test  []string
	Other []string
}

func rvsClassify(lines []string) rvsNames {
	var n rvsNames
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || l == "ok" || l == "FAIL" || l == "PASS" || !rvsIdentRe.MatchString(l) {
			continue
		}
		switch {
		case strings.HasPrefix(l, "Benchmark"):
		case strings.HasPrefix(l, "Test"):
			n.Test = append(n.Test, l)
		default:
			n.Other = append(n.Other, l)
		}
	}
	return n
}

// rvsListRaw runs `go test [-race] -list '.*' <pkg>` with workDir as the working
// directory and returns the stdout lines. A non-zero exit is a test failure: a
// failed listing is not evidence.
func rvsListRaw(t *testing.T, workDir, pkg string, race bool, env ...string) []string {
	t.Helper()
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	args = append(args, "-list", ".*", pkg)
	ctx, cancel := context.WithTimeout(context.Background(), rvsListTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Run()
	t.Logf("rvs: go %s (dir %s) took %s", strings.Join(args, " "), workDir, time.Since(start).Round(time.Millisecond))
	if err != nil {
		t.Fatalf("RELEASE_VERIFY_SHARD_DRIFT: go %s failed: %v\nstdout tail: %s\nstderr tail: %s",
			strings.Join(args, " "), err, rvsTail(stdout.String()), rvsTail(stderr.String()))
	}
	return strings.Split(stdout.String(), "\n")
}

// rvsShardBuildRaw lists names under the shards' own build conditions: the
// shards run with -race, which also sets the race build tag, so a name that
// exists only in a //go:build race file runs in the shards and is listed only
// under this flag. This is the single place that decision is made.
func rvsShardBuildRaw(t *testing.T, workDir, pkg string, env ...string) []string {
	t.Helper()
	return rvsListRaw(t, workDir, pkg, true, env...)
}

func rvsTail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 800 {
		return "..." + s[len(s)-800:]
	}
	return s
}

// ---------------------------------------------------------------------------
// The checks
// ---------------------------------------------------------------------------

type rvsSelector struct {
	Shard string
	Expr  string
	Re    *regexp.Regexp
}

// rvsCoverFailures: every name must match exactly one selector.
func rvsCoverFailures(label string, tests []string, sels []rvsSelector) []string {
	var out []string
	for _, name := range tests {
		var hit []string
		for _, s := range sels {
			if s.Re.MatchString(name) {
				hit = append(hit, fmt.Sprintf("%s %s", s.Shard, s.Expr))
			}
		}
		if len(hit) != 1 {
			out = append(out, fmt.Sprintf("%s test %q matches %d selectors [%s], want exactly 1", label, name, len(hit), strings.Join(hit, "; ")))
		}
	}
	return out
}

// rvsDiscoveryFailures: no fuzz target or example may be listed.
func rvsDiscoveryFailures(label string, other []string) []string {
	var out []string
	for _, name := range other {
		out = append(out, fmt.Sprintf("discovery guard: go test -race -list lists %q for %s, and no ^Test selector would run it", name, label))
	}
	return out
}

// rvsInputs is every input the check reads, already parsed. A mutant edits a
// fresh copy (the sources are re-parsed for each) and must be detected.
type rvsInputs struct {
	ci        rvsCI
	rel       rvsRelease
	shOS      []string
	shShards  []string
	module    string
	cliRaw    []string
	rosterRaw []string
}

func (in *rvsInputs) cliSelectors(t *testing.T) []rvsSelector {
	t.Helper()
	cfgs := in.rel.shardCfgs()
	var out []rvsSelector
	for _, shard := range slices.Sorted(maps.Keys(cfgs)) {
		cfg := cfgs[shard]
		if cfg.Pkg != rvsCLIPkg {
			continue
		}
		re, err := regexp.Compile(cfg.Run)
		if err != nil {
			t.Fatalf("shard %s: selector %q does not compile: %v", shard, cfg.Run, err)
		}
		out = append(out, rvsSelector{Shard: shard, Expr: cfg.Run, Re: re})
	}
	return out
}

// failures returns every divergence found; empty means the matrix, the
// selectors and the listed names agree.
func (in *rvsInputs) failures(t *testing.T) []string {
	t.Helper()
	var out []string
	add := func(format string, a ...any) { out = append(out, fmt.Sprintf(format, a...)) }

	osVals, shardVals := in.rel.osValues(), in.rel.shardValues()
	incs := rvsMaps(in.rel.Matrix["include"])
	cfgs := in.rel.shardCfgs()

	// --- matrix shape (AC-RVS-001) ---
	if !slices.Equal(rvsSortedSet(osVals), rvsSortedSet(in.shOS)) || len(osVals) != len(rvsSortedSet(osVals)) {
		add("matrix os %v differs from MATRIX_OSES %v in %s", osVals, in.shOS, rvsReleaseScriptRel)
	}
	if !slices.Equal(rvsSortedSet(shardVals), rvsSortedSet(in.shShards)) || len(shardVals) != len(rvsSortedSet(shardVals)) {
		add("matrix shard ids %v differ from MATRIX_SHARDS %v in %s", shardVals, in.shShards, rvsReleaseScriptRel)
	}
	if len(osVals) != rvsWantOSes {
		add("matrix has %d os values, want %d", len(osVals), rvsWantOSes)
	}
	if len(shardVals) != rvsWantShards {
		add("matrix has %d shard ids, want %d (the count ci.yml uses)", len(shardVals), rvsWantShards)
	}
	if _, has := in.rel.Matrix["exclude"]; has {
		add("matrix carries an exclude key; the final leg set must be the full os x shard product")
	}
	if len(incs) != len(shardVals) {
		add("include has %d entries, want %d (exactly one per shard id)", len(incs), len(shardVals))
	}
	seen := map[string]int{}
	for i, inc := range incs {
		shard, ok := inc["shard"].(string)
		if !ok {
			add("include entry %d has no shard key", i+1)
			continue
		}
		seen[shard]++
		if _, hasOS := inc["os"]; hasOS {
			add("include entry %d (shard %s) names os; entries must be keyed by shard alone", i+1, shard)
		}
	}
	for _, s := range slices.Sorted(maps.Keys(seen)) {
		if seen[s] > 1 {
			add("include lists shard %s %d times, want once", s, seen[s])
		}
		if !slices.Contains(shardVals, s) {
			add("include entry for shard %s, which the matrix does not declare", s)
		}
	}
	for _, s := range shardVals {
		if seen[s] == 0 {
			add("include has no entry for shard %s", s)
		}
	}

	// --- expanded legs, artifact names, vet legs (AC-RVS-001/007/013) ---
	legs := rvsExpand(in.rel.Matrix)
	if want := len(osVals) * len(shardVals); len(legs) != want {
		add("matrix expands to %d legs, want %d (%d os x %d shards)", len(legs), want, len(osVals), len(shardVals))
	}
	pairs := map[string]bool{}
	for _, leg := range legs {
		pairs[leg["os"]+"\x00"+leg["shard"]] = true
	}
	if len(pairs) != len(legs) {
		add("matrix expands to %d legs but only %d distinct (os, shard) pairs", len(legs), len(pairs))
	}
	for _, o := range osVals {
		for _, s := range shardVals {
			if !pairs[o+"\x00"+s] {
				add("leg (%s, %s) is missing from the expanded matrix", o, s)
			}
		}
	}
	if in.rel.ArtifactName == "" {
		add("no %q step with a with.name was found in job %s", rvsUploadStep, rvsMatrixJob)
	}
	byName := map[string]int{}
	for _, leg := range legs {
		name := rvsSubstMatrix(in.rel.ArtifactName, leg)
		byName[name]++
		if !strings.Contains(name, leg["os"]) || !strings.Contains(name, leg["shard"]) {
			add("artifact name %q for leg (%s, %s) does not carry both its os and its shard", name, leg["os"], leg["shard"])
		}
	}
	for _, name := range slices.Sorted(maps.Keys(byName)) {
		if byName[name] > 1 {
			add("artifact name %q is shared by %d legs, so uploads would collide", name, byName[name])
		}
	}
	vet := 0
	for _, leg := range legs {
		if leg["vet"] == "true" {
			vet++
		}
	}
	if vet != len(osVals) {
		add("%d legs run go vet, want %d (once per os, not once per shard)", vet, len(osVals))
	}

	// --- selector and package parity with ci.yml (REQ-RVS-003) ---
	all := map[string]bool{}
	for s := range in.ci.Shards {
		all[s] = true
	}
	for s := range cfgs {
		all[s] = true
	}
	for _, s := range in.shShards {
		all[s] = true
	}
	for _, shard := range slices.Sorted(maps.Keys(all)) {
		c, cOK := in.ci.Shards[shard]
		r, rOK := cfgs[shard]
		switch {
		case !cOK:
			add("shard %s: no race step found for it in %s", shard, rvsCIWorkflowRel)
		case !rOK:
			add("shard %s: no include entry for it in %s", shard, rvsReleaseWorkflowRel)
		default:
			if c.Run != r.Run {
				add("shard %s: selector diverges: %s has %q, %s has %q", shard, rvsCIWorkflowRel, c.Run, rvsReleaseWorkflowRel, r.Run)
			}
			if c.Pkg != r.Pkg {
				add("shard %s: package diverges: %s has %q, %s has %q", shard, rvsCIWorkflowRel, c.Pkg, rvsReleaseWorkflowRel, r.Pkg)
			}
		}
		if rOK && strings.Contains(r.Run, "/") {
			add("shard %s: selector %q contains '/', which go test -run splits into subtest levels", shard, r.Run)
		}
	}
	relExcl := rvsGrepVExclusions(in.rel.TestRun)
	switch {
	case len(in.ci.RemainderExcl) == 0:
		add("remainder exclusion list not found in %s", rvsCIWorkflowRel)
	case len(relExcl) == 0:
		add("remainder exclusion list not found in %s (step %q)", rvsReleaseWorkflowRel, rvsTestStep)
	case !slices.Equal(rvsSortedSet(in.ci.RemainderExcl), rvsSortedSet(relExcl)):
		add("remainder exclusion list diverges: %s has %v, %s has %v", rvsCIWorkflowRel, in.ci.RemainderExcl, rvsReleaseWorkflowRel, relExcl)
	}
	// Partition identity: the exclusion list is exactly the packages the other
	// shards name, so each package runs once and none is dropped.
	var named []string
	for _, shard := range slices.Sorted(maps.Keys(cfgs)) {
		pkg := cfgs[shard].Pkg
		if shard == rvsRemainderShard || pkg == "" || pkg == rvsRemainderShard {
			continue
		}
		named = append(named, in.module+"/"+strings.Trim(strings.TrimPrefix(pkg, "./"), "/"))
	}
	if len(relExcl) > 0 && !slices.Equal(rvsSortedSet(named), rvsSortedSet(relExcl)) {
		add("remainder exclusion list %v is not the package set the other shards name %v; a package would run twice or never", relExcl, rvsSortedSet(named))
	}

	// --- exact cover and discovery guard over the real names (REQ-RVS-015/016) ---
	cli, roster := rvsClassify(in.cliRaw), rvsClassify(in.rosterRaw)
	if len(cli.Test) == 0 {
		add("go test -race -list printed no Test names for internal/cli; an empty list asserts nothing")
	}
	if len(roster.Test) == 0 {
		add("go test -race -list printed no Test names for the rosterguard package; an empty list asserts nothing")
	}
	cliSels := in.cliSelectors(t)
	if len(cliSels) != rvsWantCLISelectors {
		add("found %d internal/cli selectors in the release matrix, want %d", len(cliSels), rvsWantCLISelectors)
	}
	out = append(out, rvsCoverFailures("internal/cli", cli.Test, cliSels)...)
	if rc, ok := cfgs[rvsRosterShard]; ok {
		re, err := regexp.Compile(rc.Run)
		if err != nil {
			add("shard %s: selector %q does not compile: %v", rvsRosterShard, rc.Run, err)
		} else {
			out = append(out, rvsCoverFailures("rosterguard", roster.Test, []rvsSelector{{Shard: rvsRosterShard, Expr: rc.Run, Re: re}})...)
		}
	}
	out = append(out, rvsDiscoveryFailures("internal/cli", cli.Other)...)
	out = append(out, rvsDiscoveryFailures("the rosterguard package", roster.Other)...)
	return out
}

func rvsSortedSet(ss []string) []string {
	c := slices.Clone(ss)
	slices.Sort(c)
	return slices.Compact(c)
}

// ---------------------------------------------------------------------------
// Sources
// ---------------------------------------------------------------------------

type rvsSources struct {
	ciText, relText, shText, modText string
	cliRaw, rosterRaw                []string
}

func (s rvsSources) inputs(t *testing.T) *rvsInputs {
	t.Helper()
	m := rvsModuleRe.FindStringSubmatch(s.modText)
	if m == nil {
		t.Fatalf("go.mod: no module line")
	}
	return &rvsInputs{
		ci:        rvsParseCI(t, s.ciText),
		rel:       rvsParseRelease(t, s.relText),
		shOS:      rvsShellArray(t, s.shText, "MATRIX_OSES"),
		shShards:  rvsShellArray(t, s.shText, "MATRIX_SHARDS"),
		module:    m[1],
		cliRaw:    slices.Clone(s.cliRaw),
		rosterRaw: slices.Clone(s.rosterRaw),
	}
}

func rvsReadFile(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("%s: read: %v", rel, err)
	}
	return string(raw)
}

// ---------------------------------------------------------------------------
// Mutation table
// ---------------------------------------------------------------------------

func rvsFindInclude(t *testing.T, in *rvsInputs, shard string) map[string]any {
	t.Helper()
	for _, inc := range rvsMaps(in.rel.Matrix["include"]) {
		if s, _ := inc["shard"].(string); s == shard {
			return inc
		}
	}
	t.Fatalf("mutation anchor: no include entry for shard %s", shard)
	return nil
}

type rvsMutant struct {
	id string
	// editSrc, when set, edits the workflow text before it is parsed, so the
	// extractors are proven to read the file and not only the parsed model.
	editSrc func(t *testing.T, s *rvsSources)
	apply   func(t *testing.T, in *rvsInputs)
	// want: every substring must appear in at least one reported failure.
	want []string
}

// rvsReplaceOnce replaces the single occurrence of old and fails when the
// anchor is missing or ambiguous, so a mutation can never apply to nothing.
func rvsReplaceOnce(t *testing.T, text, old, repl string) string {
	t.Helper()
	if n := strings.Count(text, old); n != 1 {
		t.Fatalf("mutation anchor %q occurs %d times, want exactly 1", old, n)
	}
	return strings.Replace(text, old, repl, 1)
}

func rvsMutants() []rvsMutant {
	injectCLI := func(names ...string) func(*testing.T, *rvsInputs) {
		return func(_ *testing.T, in *rvsInputs) { in.cliRaw = append(in.cliRaw, names...) }
	}
	noop := func(*testing.T, *rvsInputs) {}
	return []rvsMutant{
		{
			// The release workflow text itself, read through the YAML parser.
			id: "i-text_release_workflow_selector_J-R_to_J-Q",
			editSrc: func(t *testing.T, s *rvsSources) {
				s.relText = rvsReplaceOnce(t, s.relText, "run: '^Test[J-R]'", "run: '^Test[J-Q]'")
			},
			apply: noop,
			want:  []string{"shard cli-3: selector diverges"},
		},
		{
			// The other side: ci.yml's text, read through the shard-step regex.
			id: "i-text_ci_workflow_selector_J-R_to_J-Q",
			editSrc: func(t *testing.T, s *rvsSources) {
				s.ciText = rvsReplaceOnce(t, s.ciText, "-run '^Test[J-R]' ./internal/cli/", "-run '^Test[J-Q]' ./internal/cli/")
			},
			apply: noop,
			want:  []string{"shard cli-3: selector diverges"},
		},
		{
			id: "i_selector_J-R_to_J-Q",
			apply: func(t *testing.T, in *rvsInputs) {
				inc := rvsFindInclude(t, in, "cli-3")
				old := inc["run"].(string)
				inc["run"] = strings.Replace(old, "[J-R]", "[J-Q]", 1)
				if inc["run"] == old {
					t.Fatalf("mutation did not apply to %q", old)
				}
			},
			want: []string{"shard cli-3: selector diverges"},
		},
		{
			id: "ii_remainder_exclusion_removed",
			apply: func(t *testing.T, in *rvsInputs) {
				drop := " -e 'github.com/modu-ai/moai-adk/internal/harness/rosterguard'"
				lines := strings.Split(in.rel.TestRun, "\n")
				changed := false
				for i, l := range lines {
					if strings.Contains(l, "grep -v -x") && strings.Contains(l, drop) {
						lines[i] = strings.Replace(l, drop, "", 1)
						changed = true
						break
					}
				}
				if !changed {
					t.Fatalf("mutation anchor: no grep -v -x line carrying the rosterguard exclusion")
				}
				in.rel.TestRun = strings.Join(lines, "\n")
			},
			want: []string{"remainder exclusion list diverges"},
		},
		{
			id:    "iii_injected_bare_Test_matches_zero",
			apply: injectCLI("Test"),
			want:  []string{`test "Test" matches 0 selectors`},
		},
		{
			id: "iv_shard2_widened_to_F-I_matches_two",
			apply: func(t *testing.T, in *rvsInputs) {
				inc := rvsFindInclude(t, in, "cli-2")
				inc["run"] = "^Test[F-I]"
				injectCLI("TestFaInjected")(t, in)
			},
			want: []string{`test "TestFaInjected" matches 2 selectors`, "shard cli-2: selector diverges"},
		},
		{id: "v_injected_FuzzX", apply: injectCLI("FuzzX"), want: []string{`lists "FuzzX"`}},
		{id: "v_injected_ExampleX", apply: injectCLI("ExampleX"), want: []string{`lists "ExampleX"`}},
		{id: "v_injected_bare_Fuzz", apply: injectCLI("Fuzz"), want: []string{`lists "Fuzz"`}},
		{id: "v_injected_bare_Example", apply: injectCLI("Example"), want: []string{`lists "Example"`}},
		{
			id:    "v_injected_FuzzX_in_rosterguard",
			apply: func(_ *testing.T, in *rvsInputs) { in.rosterRaw = append(in.rosterRaw, "FuzzR") },
			want:  []string{`lists "FuzzR"`},
		},
		{
			id: "vi_artifact_name_without_shard",
			apply: func(t *testing.T, in *rvsInputs) {
				old := in.rel.ArtifactName
				in.rel.ArtifactName = regexp.MustCompile(`-?\$\{\{\s*matrix\.shard\s*\}\}`).ReplaceAllString(old, "")
				if in.rel.ArtifactName == old {
					t.Fatalf("mutation did not apply to %q", old)
				}
			},
			want: []string{"artifact name", "shared by"},
		},
		{
			id: "vii_second_vet_include",
			apply: func(t *testing.T, in *rvsInputs) {
				rvsFindInclude(t, in, "cli-3")["vet"] = true
			},
			want: []string{"legs run go vet"},
		},
		{
			id: "viii_exclude_ubuntu_remainder",
			apply: func(_ *testing.T, in *rvsInputs) {
				in.rel.Matrix["exclude"] = []any{map[string]any{"os": "ubuntu-latest", "shard": "remainder"}}
			},
			want: []string{"carries an exclude key", "17 legs", "leg (ubuntu-latest, remainder) is missing"},
		},
		{
			id: "ix_duplicated_include_entry",
			apply: func(t *testing.T, in *rvsInputs) {
				incs, ok := in.rel.Matrix["include"].([]any)
				if !ok || len(incs) == 0 {
					t.Fatalf("mutation anchor: no include list")
				}
				in.rel.Matrix["include"] = append(incs, incs[0])
				// The SPEC's prediction: the copy merges into the same
				// combinations, so the leg count stays 18 and only the
				// include-entry assertions can catch it.
				if got := len(rvsExpand(in.rel.Matrix)); got != rvsWantOSes*rvsWantShards {
					t.Fatalf("duplicated include changed the leg count to %d; the structural-catch premise is wrong", got)
				}
			},
			want: []string{"include has 7 entries, want 6", "lists shard cli-1 2 times"},
		},
		{
			id: "xi_five_shards_remainder_dropped",
			apply: func(_ *testing.T, in *rvsInputs) {
				var shards, incs []any
				for _, s := range in.rel.Matrix["shard"].([]any) {
					if s != rvsRemainderShard {
						shards = append(shards, s)
					}
				}
				for _, inc := range in.rel.Matrix["include"].([]any) {
					if inc.(map[string]any)["shard"] != rvsRemainderShard {
						incs = append(incs, inc)
					}
				}
				in.rel.Matrix["shard"], in.rel.Matrix["include"] = shards, incs
			},
			want: []string{"MATRIX_SHARDS", "matrix has 5 shard ids"},
		},
	}
}

// ---------------------------------------------------------------------------
// The test
// ---------------------------------------------------------------------------

func TestReleaseVerifyShardParity(t *testing.T) {
	root := findProjectRootForMirrorTest(t)

	src := rvsSources{
		ciText:  rvsReadFile(t, root, rvsCIWorkflowRel),
		relText: rvsReadFile(t, root, rvsReleaseWorkflowRel),
		shText:  rvsReadFile(t, root, rvsReleaseScriptRel),
		modText: rvsReadFile(t, root, "go.mod"),
	}
	// Two nested race listings, each compiling a race test binary. Taken once
	// here and shared by the repository check and the mutation table below.
	src.cliRaw = rvsShardBuildRaw(t, root, rvsCLIPkg)
	src.rosterRaw = rvsShardBuildRaw(t, root, rvsRosterPkg)
	cli, roster := rvsClassify(src.cliRaw), rvsClassify(src.rosterRaw)
	t.Logf("rvs: internal/cli lists %d Test names (+%d other), rosterguard %d (+%d other)", len(cli.Test), len(cli.Other), len(roster.Test), len(roster.Other))

	t.Run("repository_state", func(t *testing.T) {
		in := src.inputs(t)
		for _, f := range in.failures(t) {
			t.Errorf("RELEASE_VERIFY_SHARD_DRIFT: %s", f)
		}
	})

	t.Run("controls_must_pass", func(t *testing.T) {
		controls := []struct {
			id    string
			names []string
		}{
			// C1: a benchmark never runs without -bench, so it must not fail.
			{"C1_injected_BenchmarkX", []string{"BenchmarkX"}},
			// C2: a name matching exactly one selector is covered.
			{"C2_injected_TestZ", []string{"TestZebraInjected"}},
		}
		for _, c := range controls {
			t.Run(c.id, func(t *testing.T) {
				in := src.inputs(t)
				in.cliRaw = append(in.cliRaw, c.names...)
				if got := in.failures(t); len(got) != 0 {
					t.Fatalf("control %s must pass but reported: %v", c.id, got)
				}
			})
		}
	})

	t.Run("mutants_must_fail", func(t *testing.T) {
		for _, m := range rvsMutants() {
			t.Run(m.id, func(t *testing.T) {
				s := src
				if m.editSrc != nil {
					m.editSrc(t, &s)
				}
				in := s.inputs(t)
				m.apply(t, in)
				got := in.failures(t)
				if len(got) == 0 {
					t.Fatalf("mutant %s was NOT detected: the check cannot fail on this input", m.id)
				}
				for _, w := range m.want {
					found := false
					for _, f := range got {
						if strings.Contains(f, w) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("mutant %s: detected, but no failure contains %q; got %v", m.id, w, got)
					}
				}
				t.Logf("mutant %s detected: %s", m.id, got[0])
			})
		}
	})

	// Mutant (x): a name that exists only in a //go:build race file. Evidenced by
	// a PAIR of runs on one clean-base package: the plain listing exits 0 and
	// passes both guards; the -race listing exits 0 and fails the discovery guard
	// naming FuzzRaceOnly. A collector that dropped -race would give the plain
	// result twice and fail the second half.
	t.Run("x_race_only_name_pair", func(t *testing.T) {
		dir := t.TempDir()
		for name, body := range map[string]string{
			"go.mod":            rvsFixtureGoMod,
			"base_test.go":      rvsFixtureBase,
			"race_only_test.go": rvsFixtureRaceOnly,
		} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatalf("write fixture %s: %v", name, err)
			}
		}
		env := []string{"GOFLAGS=", "GOWORK=off"}
		sels := src.inputs(t).cliSelectors(t)

		plain := rvsClassify(rvsListRaw(t, dir, "./...", false, env...))
		if !slices.Equal(plain.Test, []string{"TestReal"}) || len(plain.Other) != 0 {
			t.Fatalf("run 1 (plain listing): want Test=[TestReal] and no other names, got Test=%v Other=%v", plain.Test, plain.Other)
		}
		if f := append(rvsCoverFailures("fixture", plain.Test, sels), rvsDiscoveryFailures("fixture", plain.Other)...); len(f) != 0 {
			t.Fatalf("run 1 (plain listing) must pass both guards, got %v", f)
		}

		race := rvsClassify(rvsShardBuildRaw(t, dir, "./...", env...))
		if !slices.Equal(race.Other, []string{"FuzzRaceOnly"}) {
			t.Fatalf("run 2 (-race listing): want Other=[FuzzRaceOnly], got Test=%v Other=%v (a collector without -race gives run 1's result)", race.Test, race.Other)
		}
		if f := rvsCoverFailures("fixture", race.Test, sels); len(f) != 0 {
			t.Fatalf("run 2 (-race listing): the cover check must still pass, got %v", f)
		}
		got := rvsDiscoveryFailures("fixture", race.Other)
		if len(got) != 1 || !strings.Contains(got[0], `lists "FuzzRaceOnly"`) {
			t.Fatalf("run 2 (-race listing): the discovery guard must fail naming FuzzRaceOnly, got %v", got)
		}
		t.Logf("mutant x detected only under -race: %s", got[0])
	})
}
