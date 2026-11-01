// ci_race_scope_test.go: the PR race-shard scope of ci.yml (card t2012).
//
// scripts/ci/race-scope.sh decides which of the six advisory Race Test shards a
// pull request needs: cli-1..cli-4 (internal/cli), rosterguard, remainder. Its
// evidence is the import graph plus a textual scan, not a proof: a shard is left
// out only when no changed package is among its (test-inclusive) dependencies AND
// nothing in its module closure reads files, runs commands, embeds files or
// touches testdata (a shard that does is selected for any Go change). Anything
// it cannot judge (a non-Go file, go.mod, a workflow, testdata, a deleted
// package, an empty list, a failing `go list`) runs all six.
//
// Every row runs the real script inside a scratch Go module that mirrors the
// shard layout (internal/cli, internal/harness/rosterguard, shared and leaf
// packages, a test-less cmd/tool).
package template_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	rsAll       = ",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"
	rsCLI       = ",cli-1,cli-2,cli-3,cli-4,"
	rsScriptRel = "scripts/ci/race-scope.sh"
)

// rsFixture writes the scratch module and returns its directory.
func rsFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":                             "module example.com/fx\n\ngo 1.21\n",
		"internal/shared/shared.go":          "package shared\n\nfunc S() int { return 1 }\n",
		"internal/shared/shared_test.go":     "package shared\n\nimport \"testing\"\n\nfunc TestS(t *testing.T) { _ = S() }\n",
		"internal/cli/cli.go":                "package cli\n\nimport \"example.com/fx/internal/shared\"\n\nfunc C() int { return shared.S() }\n",
		"internal/cli/cli_test.go":           "package cli\n\nimport \"testing\"\n\nfunc TestC(t *testing.T) { _ = C() }\n",
		"internal/harness/rosterguard/rg.go": "package rosterguard\n\nfunc R() int { return 1 }\n",
		// rosterguard scans the repository's Go sources, an edge the import graph
		// cannot see (the real package does this with filepath.WalkDir).
		"internal/harness/rosterguard/rg_test.go": "package rosterguard\n\nimport (\n\t\"io/fs\"\n\t\"path/filepath\"\n\t\"testing\"\n\n\t\"example.com/fx/internal/tmpl\"\n)\n\nfunc TestR(t *testing.T) {\n\t_ = filepath.WalkDir(\"../../..\", func(string, fs.DirEntry, error) error { return nil })\n\t_ = R() + tmpl.T()\n}\n",
		"internal/tmpl/tmpl.go":                   "package tmpl\n\nfunc T() int { return 1 }\n",
		"internal/tmpl/tmpl_test.go":              "package tmpl\n\nimport \"testing\"\n\nfunc TestT(t *testing.T) { _ = T() }\n",
		"internal/leaf/leaf.go":                   "package leaf\n\nfunc L() int { return 1 }\n",
		"internal/leaf/leaf_test.go":              "package leaf\n\nimport \"testing\"\n\nfunc TestL(t *testing.T) { _ = L() }\n",
		"internal/other/other.go":                 "package other\n\nimport \"example.com/fx/internal/shared\"\n\nfunc O() int { return shared.S() }\n",
		"internal/other/other_test.go":            "package other\n\nimport \"testing\"\n\nfunc TestO(t *testing.T) { _ = O() }\n",
		"cmd/tool/main.go":                        "package main\n\nimport \"example.com/fx/internal/cli\"\n\nfunc main() { _ = cli.C() }\n",
		"internal/leaf/testdata/fixture/x.go":     "package fixture\n",
	}
	for rel, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return dir
}

// rsRun runs the real script in dir with the given changed-file list and
// returns its stdout line, stderr and exit code.
func rsRun(t *testing.T, dir string, changed []string, mutate func(dir string)) (shards, stderr string, exit int) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
	root := findProjectRootForMirrorTest(t)
	script := filepath.Join(root, filepath.FromSlash(rsScriptRel))
	list := filepath.Join(t.TempDir(), "changed.txt")
	if err := os.WriteFile(list, []byte(strings.Join(changed, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write changed list: %v", err)
	}
	if mutate != nil {
		mutate(dir)
	}
	cmd := exec.Command("bash", script, "--changed", list)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off", "GO111MODULE=on")
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	exit = 0
	if ee, ok := err.(*exec.ExitError); ok {
		exit = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("start race-scope.sh: %v", err)
	}
	return strings.TrimSpace(out.String()), errb.String(), exit
}

// TestRaceScopeNarrowsOnlyWhatItCanProve: the scoped rows. Each changed
// package selects exactly the shards whose tested packages depend on it.
func TestRaceScopeNarrowsOnlyWhatItCanProve(t *testing.T) {
	rows := []struct {
		name    string
		changed []string
		want    string
	}{
		// The rosterguard shard is selected for EVERY .go change: rosterguard scans
		// the whole repository's Go sources with WalkDir, a file-reading edge the
		// import graph cannot see (audit round 2, codex receipt
		// rcpt-3fdb738b943e3a856025b27e: changing internal/cli/agentlint/agent_lint.go
		// dropped it).
		{"leaf_package_selects_remainder_and_rosterguard", []string{"internal/leaf/leaf.go"}, ",rosterguard,remainder,"},
		{"leaf_test_file_selects_remainder_and_rosterguard", []string{"internal/leaf/leaf_test.go"}, ",rosterguard,remainder,"},
		{"cli_package_selects_cli_and_rosterguard", []string{"internal/cli/cli.go"}, ",cli-1,cli-2,cli-3,cli-4,rosterguard,"},
		{"rosterguard_package_only_rosterguard", []string{"internal/harness/rosterguard/rg.go"}, ",rosterguard,"},
		{"tmpl_is_a_test_import_of_rosterguard", []string{"internal/tmpl/tmpl.go"}, ",rosterguard,remainder,"},
		{"shared_package_reaches_cli_remainder_and_rosterguard", []string{"internal/shared/shared.go"}, ",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"},
		{"two_packages_union", []string{"internal/leaf/leaf.go", "internal/harness/rosterguard/rg.go"}, ",rosterguard,remainder,"},
		{"file_read_only_by_rosterguard_still_selects_it", []string{"internal/other/other.go"}, ",rosterguard,remainder,"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			shards, stderr, exit := rsRun(t, rsFixture(t), row.changed, nil)
			if exit != 0 {
				t.Fatalf("exit = %d\n%s", exit, stderr)
			}
			if shards != "shards="+row.want {
				t.Errorf("shards = %q, want %q\nstderr:\n%s", shards, "shards="+row.want, stderr)
			}
		})
	}
}

// TestRaceScopeSelectsEveryShardItCannotProveIndependent (audit round 2, F1): a
// shard whose packages, tests or dependencies may read files, run commands, embed
// files or touch testdata is selected for ANY Go change, because the import graph
// cannot show what such code reads. Each row plants one such primitive and changes
// a package the shard does not import; the shard must still be selected. Where no
// primitive exists (the leaf-only row of the narrowing test) the shard stays out.
func TestRaceScopeSelectsEveryShardItCannotProveIndependent(t *testing.T) {
	prims := []struct{ name, imports, stmt string }{
		{"os_readfile", `"os"`, `_, _ = os.ReadFile("x")`},
		{"os_readdir", `"os"`, `_, _ = os.ReadDir(".")`},
		{"os_stat", `"os"`, `_, _ = os.Stat("x")`},
		{"os_getwd", `"os"`, `_, _ = os.Getwd()`},
		{"os_open", `"os"`, `_, _ = os.Open("x")`},
		{"filepath_walk", `"path/filepath"`, `_ = filepath.Walk(".", nil)`},
		{"filepath_walkdir", `"path/filepath"`, `_ = filepath.WalkDir(".", nil)`},
		{"exec_command", `"os/exec"`, `_ = exec.Command("true")`},
		{"runtime_caller", `"runtime"`, `_, _, _, _ = runtime.Caller(0)`},
		{"testdata_literal", ``, `_ = "testdata/golden.txt"`},
	}
	// The primitive sits in a TEST file of the shard's own package.
	for _, p := range prims {
		for _, c := range []struct {
			shard, pkg, name, pkgDir string
			changed                  string
			want                     string
		}{
			{"cli", "cli", "cli_test_file", "internal/cli", "internal/leaf/leaf.go", ",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"},
			{"remainder", "leaf", "remainder_test_file", "internal/leaf", "internal/cli/cli.go", ",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"},
		} {
			t.Run(c.name+"/"+p.name, func(t *testing.T) {
				imp := ""
				if p.imports != "" {
					imp = "import (\n\t\"testing\"\n\t" + p.imports + "\n)\n"
				} else {
					imp = "import \"testing\"\n"
				}
				body := "package " + c.pkg + "\n\n" + imp + "\nfunc TestZZReader(t *testing.T) { " + p.stmt + " }\n"
				shards, stderr, exit := rsRun(t, rsFixture(t), []string{c.changed}, func(dir string) {
					if err := os.WriteFile(filepath.Join(dir, c.pkgDir, "zz_reader_test.go"), []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				})
				if exit != 0 || shards != "shards="+c.want {
					t.Errorf("shards = %q (exit %d), want %q\n%s", shards, exit, "shards="+c.want, stderr)
				}
			})
		}
	}

	// The primitive sits in NON-test code a shard depends on: shared.go is imported
	// by internal/cli and internal/other, so a file read there makes both groups
	// readers although the changed package (leaf) is imported by neither.
	t.Run("dependency_source_reads_files", func(t *testing.T) {
		shards, stderr, exit := rsRun(t, rsFixture(t), []string{"internal/leaf/leaf.go"}, func(dir string) {
			body := "package shared\n\nimport \"os\"\n\nfunc S() int { _, _ = os.ReadFile(\"x\"); return 1 }\n"
			if err := os.WriteFile(filepath.Join(dir, "internal/shared/shared.go"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		})
		want := ",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"
		if exit != 0 || shards != "shards="+want {
			t.Errorf("shards = %q (exit %d), want %q\n%s", shards, exit, "shards="+want, stderr)
		}
	})

	// go:embed in a dependency (a .go file or any file is pulled in without an import).
	t.Run("dependency_uses_go_embed", func(t *testing.T) {
		shards, stderr, exit := rsRun(t, rsFixture(t), []string{"internal/leaf/leaf.go"}, func(dir string) {
			if err := os.WriteFile(filepath.Join(dir, "internal/shared/data.txt"), []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			body := "package shared\n\nimport _ \"embed\"\n\n//go:embed data.txt\nvar data string\n\nfunc S() int { return len(data) }\n"
			if err := os.WriteFile(filepath.Join(dir, "internal/shared/shared.go"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		})
		want := ",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"
		if exit != 0 || shards != "shards="+want {
			t.Errorf("shards = %q (exit %d), want %q\n%s", shards, exit, "shards="+want, stderr)
		}
	})

	// Control: with no primitive in the cli group (cli.go and shared.go are pure),
	// a leaf change does not reach the cli shards. The narrowing still exists where
	// independence is provable.
	t.Run("control_pure_cli_group_stays_out", func(t *testing.T) {
		shards, stderr, exit := rsRun(t, rsFixture(t), []string{"internal/leaf/leaf.go"}, nil)
		want := ",rosterguard,remainder,"
		if exit != 0 || shards != "shards="+want {
			t.Errorf("shards = %q (exit %d), want %q\n%s", shards, exit, "shards="+want, stderr)
		}
	})
}

// TestRaceScopeClosesToAllSixWhenItCannotJudge: every row the script must not
// narrow ends with all six shards. A narrowed answer on any of these rows is the
// bug the filter must never have.
func TestRaceScopeClosesToAllSixWhenItCannotJudge(t *testing.T) {
	rows := []struct {
		name    string
		changed []string
		mutate  func(dir string)
	}{
		{name: "go_mod", changed: []string{"go.mod"}},
		{name: "go_sum", changed: []string{"go.sum"}},
		{name: "workflow", changed: []string{".github/workflows/ci.yml"}},
		{name: "filter_file", changed: []string{".github/test-input-filters.yml"}},
		{name: "makefile", changed: []string{"Makefile"}},
		{name: "ci_scripts", changed: []string{"scripts/ci/retry-flaky.sh"}},
		{name: "non_go_file_beside_a_go_file", changed: []string{"internal/leaf/leaf.go", "internal/leaf/data.json"}},
		{name: "markdown_beside_a_go_file", changed: []string{"internal/leaf/leaf.go", "README.md"}},
		{name: "template_tree_file", changed: []string{"internal/template/templates/x.md"}},
		{name: "testdata_go_file", changed: []string{"internal/leaf/testdata/fixture/x.go"}},
		{name: "deleted_package_dir", changed: []string{"internal/gone/gone.go"}},
		{name: "path_outside_every_package", changed: []string{"tools/gen.go"}},
		// Tests build and run the command binary (internal/cli/binary_fixture_test.go),
		// an edge the import graph does not show: a main package is never narrowed.
		{name: "main_package_is_built_by_tests", changed: []string{"cmd/tool/main.go"}},
		{name: "empty_list", changed: nil},
		{name: "blank_lines_only", changed: []string{"", "  "}},
		{name: "go_list_fails", changed: []string{"internal/leaf/leaf.go"}, mutate: func(dir string) {
			if err := os.WriteFile(filepath.Join(dir, "internal/leaf/leaf.go"), []byte("this is not a go file\n"), 0o644); err != nil {
				t.Fatalf("break package: %v", err)
			}
		}},
		{name: "no_go_mod_at_all", changed: []string{"internal/leaf/leaf.go"}, mutate: func(dir string) {
			if err := os.Remove(filepath.Join(dir, "go.mod")); err != nil {
				t.Fatalf("remove go.mod: %v", err)
			}
		}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			shards, stderr, exit := rsRun(t, rsFixture(t), row.changed, row.mutate)
			if exit != 0 {
				t.Fatalf("exit = %d, want 0 (the script closes to all six, it does not fail)\n%s", exit, stderr)
			}
			if shards != "shards="+rsAll {
				t.Errorf("shards = %q, want all six %q\nstderr:\n%s", shards, "shards="+rsAll, stderr)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Errorf("a closed-to-all answer must name its reason on stderr")
			}
		})
	}
}

// TestRaceScopeMissingChangedFileClosesToAll: an unreadable list is "cannot
// judge", not "nothing changed".
func TestRaceScopeMissingChangedFileClosesToAll(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
	root := findProjectRootForMirrorTest(t)
	script := filepath.Join(root, filepath.FromSlash(rsScriptRel))
	for _, args := range [][]string{{"--changed", filepath.Join(t.TempDir(), "absent.txt")}, {}} {
		cmd := exec.Command("bash", append([]string{script}, args...)...)
		cmd.Dir = rsFixture(t)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("args %v: %v", args, err)
		}
		if got := strings.TrimSpace(string(out)); got != "shards="+rsAll {
			t.Errorf("args %v: shards = %q, want all six", args, got)
		}
	}
}

type rsWorkflow struct {
	Jobs map[string]struct {
		Name  string `yaml:"name"`
		Needs any    `yaml:"needs"`
		If    string `yaml:"if"`
	} `yaml:"jobs"`
}

func rsLoadCI(t *testing.T) (rsWorkflow, string) {
	t.Helper()
	root := findProjectRootForMirrorTest(t)
	raw, err := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var wf rsWorkflow
	if err := yaml.Unmarshal(raw, &wf); err != nil {
		t.Fatal(err)
	}
	return wf, string(raw)
}

func rsNeeds(v any) []string {
	switch n := v.(type) {
	case string:
		return []string{n}
	case []any:
		var out []string
		for _, e := range n {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// TestCIRaceShardJobsConsumeTheScope: each of the six race jobs waits for
// race-scope, runs when the scope is empty (fail-closed) or names its own
// shard token, and keeps the release/ exclusion and the go_code gate.
func TestCIRaceShardJobsConsumeTheScope(t *testing.T) {
	wf, _ := rsLoadCI(t)
	shardJobs := map[string]string{
		"test-race-cli-1":       ",cli-1,",
		"test-race-cli-2":       ",cli-2,",
		"test-race-cli-3":       ",cli-3,",
		"test-race-cli-4":       ",cli-4,",
		"test-race-rosterguard": ",rosterguard,",
		"test-race-remainder":   ",remainder,",
	}
	scope, ok := wf.Jobs["race-scope"]
	if !ok {
		t.Fatal("ci.yml has no race-scope job")
	}
	if !strings.Contains(scope.If, "github.event_name == 'pull_request'") {
		t.Errorf("race-scope must narrow pull requests only; if = %q", scope.If)
	}
	for id, token := range shardJobs {
		job, ok := wf.Jobs[id]
		if !ok {
			t.Errorf("job %s missing", id)
			continue
		}
		needs := rsNeeds(job.Needs)
		has := map[string]bool{}
		for _, n := range needs {
			has[n] = true
		}
		if !has["detect"] || !has["race-scope"] {
			t.Errorf("%s needs %v, want detect and race-scope", id, needs)
		}
		for _, want := range []string{
			"needs.detect.result == 'success'",
			"needs.detect.outputs.go_code == 'true'",
			"!startsWith(github.head_ref, 'release/')",
			"needs.race-scope.outputs.shards == ''",
			"contains(needs.race-scope.outputs.shards, '" + token + "')",
			"!cancelled()",
		} {
			if !strings.Contains(job.If, want) {
				t.Errorf("%s: if lacks %q\n  if = %s", id, want, job.If)
			}
		}
	}
}

// TestCIRequiredAndWindowsJobsIgnoreTheScope: the scope reaches the six
// advisory race jobs and nothing else. Lint, Build, Test, Analyze, the Windows
// legs and every other job neither need race-scope nor mention it.
func TestCIRequiredAndWindowsJobsIgnoreTheScope(t *testing.T) {
	wf, raw := rsLoadCI(t)
	raceJobs := map[string]bool{
		"test-race-cli-1": true, "test-race-cli-2": true, "test-race-cli-3": true,
		"test-race-cli-4": true, "test-race-rosterguard": true, "test-race-remainder": true,
		"race-scope": true,
	}
	for id, job := range wf.Jobs {
		if raceJobs[id] {
			continue
		}
		for _, n := range rsNeeds(job.Needs) {
			if n == "race-scope" {
				t.Errorf("job %s needs race-scope; only the six race shard jobs may", id)
			}
		}
		if strings.Contains(job.If, "race-scope") {
			t.Errorf("job %s gates on race-scope; only the six race shard jobs may", id)
		}
	}
	// The six shard selectors themselves are untouched: still six, still -race.
	if got := strings.Count(raw, "Race Test ("); got != 6 {
		t.Errorf("ci.yml defines %d 'Race Test (' jobs, want 6", got)
	}
}

// rsGit runs git in dir with a fixed identity and fails the test on error.
func rsGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestRaceScopeSeesBothPathsOfAMovedFile: a file moved out of a package must
// still select the shards of the package it LEFT. The workflow's own diff line
// (extracted from ci.yml, so removing --no-renames from it turns this red) is
// run in a scratch repository where a rosterguard file moves into the leaf
// package; with rename detection the diff lists only the new path and the
// rosterguard shard is lost.
func TestRaceScopeSeesBothPathsOfAMovedFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := findProjectRootForMirrorTest(t)
	raw, err := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`git diff --name-only[^\n;]*`).FindString(string(raw))
	if m == "" {
		t.Fatal("ci.yml has no `git diff --name-only` line in the Race scope step")
	}

	dir := rsFixture(t)
	// A body large enough for git to pair the old and new path as a rename.
	body := "package rosterguard\n\nfunc R() int { return 1 }\n" + strings.Repeat("// filler line so the moved file is recognised as a rename\n", 30)
	rg := filepath.Join(dir, "internal/harness/rosterguard/rg.go")
	if err := os.WriteFile(rg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	rsGit(t, dir, "init", "-q", "-b", "main")
	rsGit(t, dir, "add", "-A")
	rsGit(t, dir, "commit", "-q", "-m", "base")
	base := strings.TrimSpace(rsOut(t, dir, "git", "rev-parse", "HEAD"))

	rsGit(t, dir, "mv", "internal/harness/rosterguard/rg.go", "internal/leaf/zz_moved.go")
	moved := strings.Replace(body, "package rosterguard", "package leaf", 1)
	moved = strings.Replace(moved, "func R()", "func Moved()", 1)
	if err := os.WriteFile(filepath.Join(dir, "internal/leaf/zz_moved.go"), []byte(moved), 0o644); err != nil {
		t.Fatal(err)
	}
	rsGit(t, dir, "add", "-A")
	rsGit(t, dir, "commit", "-q", "-m", "move")

	list := filepath.Join(t.TempDir(), "changed-files.txt")
	cmd := exec.Command("bash", "-c", m+` > "$OUT"`)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "BASE_SHA="+base, "OUT="+list)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("workflow diff line %q: %v\n%s", m, err, out)
	}
	changed, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"internal/harness/rosterguard/rg.go", "internal/leaf/zz_moved.go"} {
		if !strings.Contains(string(changed), want) {
			t.Errorf("the workflow's diff omits %s; got:\n%s", want, changed)
		}
	}

	shards, stderr, exit := rsRun(t, dir, strings.Split(strings.TrimSpace(string(changed)), "\n"), nil)
	if exit != 0 {
		t.Fatalf("race-scope exit = %d\n%s", exit, stderr)
	}
	if !strings.Contains(shards, ",rosterguard,") || !strings.Contains(shards, ",remainder,") {
		t.Errorf("a file moved out of rosterguard into leaf selected %q, want rosterguard and remainder", shards)
	}
}

// rsOut runs a command in dir and returns its stdout.
func rsOut(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v: %v", name, args, err)
	}
	return string(out)
}
