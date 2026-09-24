// ac_baseline_commit_guard_test.go — repository-local tests for the
// commit-time AC-snapshot guard (SPEC-ACSNAPSHOT-COMMIT-GUARD-001).
//
// The guard is a git config-defined pre-commit hook backed by two local-only
// scripts under scripts/ac-baseline/. Every scenario runs in a throwaway repo
// under t.TempDir() with an isolated git environment; nothing here writes the
// repository's own git config, hookdir, or tracked files.
package spec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// guardEnv returns an environment for git subprocesses that cannot see the
// caller's git state: every inherited GIT_* variable is dropped (a test run
// from inside a hook would otherwise inherit GIT_INDEX_FILE / GIT_DIR), global
// and system config are disabled, and HOME points at a per-test directory.
// It is set per exec.Cmd, never through t.Setenv, so parallel tests are safe.
func guardEnv(home string, extra ...string) []string {
	env := []string{}
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GIT_") || strings.HasPrefix(kv, "HOME=") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"HOME="+home,
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=guard-test",
		"GIT_AUTHOR_EMAIL=guard-test@example.invalid",
		"GIT_COMMITTER_NAME=guard-test",
		"GIT_COMMITTER_EMAIL=guard-test@example.invalid",
	)
	return append(env, extra...)
}

// guardRepo is one throwaway repository.
type guardRepo struct {
	t    *testing.T
	dir  string
	home string
}

func newGuardRepo(t *testing.T) *guardRepo {
	t.Helper()
	base := t.TempDir()
	r := &guardRepo{t: t, dir: filepath.Join(base, "repo"), home: filepath.Join(base, "home")}
	for _, d := range []string{r.dir, r.home} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	r.git("init", "-q", "-b", "main")
	return r
}

// run executes name+args in dir with the isolated env and returns stdout,
// stderr, and the exit code. A non-exit failure (binary missing) is fatal.
func (r *guardRepo) runIn(dir string, extraEnv []string, name string, args ...string) (string, string, int) {
	r.t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = guardEnv(r.home, extraEnv...)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if !asExitError(err, &ee) {
			r.t.Fatalf("%s %v: %v", name, args, err)
		}
		code = ee.ExitCode()
	}
	return out.String(), errb.String(), code
}

// git runs a git command that must succeed and returns its trimmed stdout.
func (r *guardRepo) git(args ...string) string {
	r.t.Helper()
	out, errb, code := r.runIn(r.dir, nil, "git", args...)
	if code != 0 {
		r.t.Fatalf("git %v: exit %d: %s", args, code, errb)
	}
	return strings.TrimSpace(out)
}

func (r *guardRepo) write(rel, content string) {
	r.t.Helper()
	p := filepath.Join(r.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatalf("write %s: %v", rel, err)
	}
}

func (r *guardRepo) head() string {
	r.t.Helper()
	return r.git("rev-parse", "HEAD")
}

// requireRealCommitGit skips real-commit subtests where the premise cannot
// hold: config-defined hooks need git >= 2.54, and hook-command shell
// semantics under Git for Windows are unmeasured (plan.md R7).
func requireRealCommitGit(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("config-defined hook shell semantics under Git for Windows are unmeasured (plan.md R7)")
	}
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) < 3 {
		t.Skipf("unparseable git version %q", out)
	}
	parts := strings.SplitN(fields[2], ".", 3)
	major, _ := strconv.Atoi(parts[0])
	minor := 0
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if major < 2 || (major == 2 && minor < 54) {
		t.Skipf("git %s < 2.54: config-defined hooks unavailable", fields[2])
	}
}

// TestACBaselineCommitGuard is the single selector for every criterion of
// SPEC-ACSNAPSHOT-COMMIT-GUARD-001.
func TestACBaselineCommitGuard(t *testing.T) {
	// M0 premise probe: a config-defined pre-commit hook fires and aborts the
	// commit while core.hooksPath=/dev/null, coexists with a hookdir file,
	// and is inherited by a linked worktree. If this fails, nothing else in
	// the design holds.
	t.Run("commit/premise_probe", func(t *testing.T) {
		requireRealCommitGit(t)
		r := newGuardRepo(t)
		r.write("seed.txt", "seed\n")
		r.git("add", "seed.txt")
		r.git("commit", "-q", "-m", "seed")
		r.git("config", "hook.probe.event", "pre-commit")
		r.git("config", "hook.probe.command", "echo PROBE-FIRED >&2; exit 1")

		attempt := func(dir string, extra ...string) (string, int) {
			t.Helper()
			before := strings.TrimSpace(mustOut(t, r, dir, "rev-parse", "HEAD"))
			if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed "+strconv.Itoa(len(extra))+" "+dir+"\n"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
			args := append([]string{"commit", "-a", "-m", "probe"}, extra...)
			_, errb, code := r.runIn(dir, nil, "git", args...)
			after := strings.TrimSpace(mustOut(t, r, dir, "rev-parse", "HEAD"))
			if before != after {
				t.Errorf("HEAD moved in %s despite failing config hook: %s -> %s", dir, before, after)
			}
			return errb, code
		}

		// Leg 1: core.hooksPath=/dev/null.
		r.git("config", "core.hooksPath", "/dev/null")
		errb, code := attempt(r.dir)
		t.Logf("leg1 hooksPath=/dev/null: exit=%d stderr=%q", code, errb)
		if code == 0 || !strings.Contains(errb, "PROBE-FIRED") {
			t.Fatalf("PREMISE FAILED: config hook did not fire/abort under core.hooksPath=/dev/null (exit=%d stderr=%q)", code, errb)
		}

		// Leg 2: hookdir pre-commit present (hooksPath unset) — both coexist.
		r.git("config", "--unset", "core.hooksPath")
		marker := filepath.Join(r.home, "hookdir-ran")
		hook := filepath.Join(r.dir, ".git", "hooks", "pre-commit")
		if err := os.WriteFile(hook, []byte("#!/bin/sh\ntouch '"+marker+"'\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write hookdir hook: %v", err)
		}
		errb, code = attempt(r.dir)
		_, statErr := os.Stat(marker)
		t.Logf("leg2 hookdir present: exit=%d hookdir_marker_written=%v stderr=%q", code, statErr == nil, errb)
		if code == 0 || !strings.Contains(errb, "PROBE-FIRED") {
			t.Fatalf("PREMISE FAILED: config hook did not abort with a hookdir pre-commit present (exit=%d)", code)
		}

		// Leg 3: linked worktree inherits the config hook.
		wt := filepath.Join(filepath.Dir(r.dir), "wt")
		r.git("worktree", "add", "-q", "-b", "wt", wt)
		errb, code = attempt(wt)
		t.Logf("leg3 linked worktree: exit=%d stderr=%q", code, errb)
		if code == 0 || !strings.Contains(errb, "PROBE-FIRED") {
			t.Fatalf("PREMISE FAILED: config hook did not fire/abort in a linked worktree (exit=%d)", code)
		}
	})

	guardCheckerSubtests(t)
	guardCommitSubtests(t)
}

func mustOut(t *testing.T, r *guardRepo, dir string, args ...string) string {
	t.Helper()
	out, errb, code := r.runIn(dir, nil, "git", args...)
	if code != 0 {
		t.Fatalf("git %v in %s: exit %d: %s", args, dir, code, errb)
	}
	return out
}

const (
	guardSpecX       = ".moai/specs/SPEC-X-001/acceptance.md"
	guardCheckerRel  = "scripts/ac-baseline/check-staged.sh"
	guardInstallRel  = "scripts/ac-baseline/install-hook.sh"
	guardFixtureDir  = "testdata/ac_baseline_guard"
	guardRejectLine  = "ac-baseline-guard: REJECT"
	guardCheckedLine = "ac-baseline-guard: checked"
	guardNotChecked  = "ac-baseline-guard: NOT CHECKED"
)

func guardFixture(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "internal/spec", guardFixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(raw)
}

// guardCarrier returns the tracked counter carrier, read at runtime so the
// tests never hold a second copy of the counter (REQ-ABG-007, plan.md R4).
func guardCarrier(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), acLocalClausePath))
	if err != nil {
		t.Fatalf("read counter carrier: %v", err)
	}
	return string(raw)
}

func guardCountRecord(path string, live, excluded int) string {
	return fmt.Sprintf("%s  COUNT %d  live=%d excluded=%d ambiguous=0", path, live, live, excluded)
}

func guardHaltRecord(path, ids string) string {
	return fmt.Sprintf("%s  HALT %s  owner=t1150 reason=fixture", path, ids)
}

func guardBaseline(records ...string) string {
	return "# fixture baseline\n# lifecycle cascade: .moai/docs/ac-count-baseline-refresh.md\n" + strings.Join(records, "\n") + "\n"
}

// copyGuardScripts copies the scripts under test into the temp repo when they
// exist (before GREEN they do not, and the checker invocation fails RED).
func (r *guardRepo) copyGuardScripts() {
	r.t.Helper()
	for _, rel := range []string{guardCheckerRel, guardInstallRel} {
		raw, err := os.ReadFile(filepath.Join(repoRoot(r.t), rel))
		if err != nil {
			continue
		}
		r.write(rel, string(raw))
	}
}

// seedCommon builds the common Given: HEAD holds the counter carrier, the
// scripts, SPEC-X-001/acceptance.md with two live criteria, and a baseline
// recording it as COUNT 2.
func seedCommon(t *testing.T, extraFiles map[string]string, records ...string) *guardRepo {
	t.Helper()
	r := newGuardRepo(t)
	r.write(acLocalClausePath, guardCarrier(t))
	r.copyGuardScripts()
	r.write(guardSpecX, guardFixture(t, "two.md"))
	for rel, body := range extraFiles {
		r.write(rel, body)
	}
	recs := append([]string{guardCountRecord(guardSpecX, 2, 0)}, records...)
	r.write(acBaselineSnapshotPath, guardBaseline(recs...))
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "seed")
	return r
}

func (r *guardRepo) stage(rel, content string) {
	r.t.Helper()
	r.write(rel, content)
	r.git("add", "--", rel)
}

func (r *guardRepo) check() (string, string, int) {
	r.t.Helper()
	return r.runIn(r.dir, nil, "sh", guardCheckerRel)
}

func nonEmptyLines(s string) []string {
	out := []string{}
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			out = append(out, ln)
		}
	}
	return out
}

func guardCheckerSubtests(t *testing.T) {
	two, twoProse, three, halt := guardFixture(t, "two.md"), guardFixture(t, "two_prose.md"), guardFixture(t, "three.md"), guardFixture(t, "halt.md")

	t.Run("reject_count_move", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(guardSpecX, three)
		_, errb, code := r.check()
		if code == 0 {
			t.Fatalf("count-moving amendment without baseline must be rejected; exit 0 stderr=%q", errb)
		}
		for _, want := range []string{guardRejectLine, guardSpecX, "live=2", "live=3", acRegenerateCommand, ".moai/docs/ac-count-baseline-refresh.md"} {
			if !strings.Contains(errb, want) {
				t.Errorf("reject stderr lacks %q:\n%s", want, errb)
			}
		}
	})

	t.Run("pass_count_unchanged", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(guardSpecX, twoProse)
		_, errb, code := r.check()
		lines := nonEmptyLines(errb)
		if code != 0 || len(lines) != 1 || !strings.HasPrefix(lines[0], guardCheckedLine+" 1") {
			t.Fatalf("want exit 0 and exactly one %q line; exit=%d stderr=%q", guardCheckedLine+" 1", code, errb)
		}
	})

	t.Run("pass_new_file", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(".moai/specs/SPEC-Y-001/acceptance.md", three)
		out, errb, code := r.check()
		if code != 0 || out != "" || errb != "" {
			t.Fatalf("new file (status A) must be silent; exit=%d stdout=%q stderr=%q", code, out, errb)
		}
	})

	unrecorded := ".moai/specs/SPEC-Z-001/acceptance.md"
	t.Run("pass_unrecorded_counts", func(t *testing.T) {
		r := seedCommon(t, map[string]string{unrecorded: two})
		r.stage(unrecorded, three)
		_, errb, code := r.check()
		if code != 0 || !strings.Contains(errb, guardCheckedLine+" 1") || !strings.Contains(errb, unrecorded) || !strings.Contains(errb, "COUNT 3") {
			t.Fatalf("unrecorded counting file must pass with a COUNT 3 report; exit=%d stderr=%q", code, errb)
		}
	})

	t.Run("pass_unrecorded_halts", func(t *testing.T) {
		r := seedCommon(t, map[string]string{unrecorded: two})
		r.stage(unrecorded, halt)
		_, errb, code := r.check()
		if code != 0 || !strings.Contains(errb, guardCheckedLine+" 1") || !strings.Contains(errb, "HALT AC-HLT-001") {
			t.Fatalf("unrecorded halting file must pass with a HALT report; exit=%d stderr=%q", code, errb)
		}
	})

	t.Run("pass_with_staged_baseline", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(guardSpecX, three)
		r.stage(acBaselineSnapshotPath, guardBaseline(guardCountRecord(guardSpecX, 3, 0)))
		_, errb, code := r.check()
		if code != 0 {
			t.Fatalf("amendment + staged regenerated baseline must pass; exit=%d stderr=%q", code, errb)
		}
	})

	t.Run("reject_unstaged_baseline", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(guardSpecX, three)
		r.write(acBaselineSnapshotPath, guardBaseline(guardCountRecord(guardSpecX, 3, 0)))
		_, errb, code := r.check()
		if code == 0 || !strings.Contains(errb, guardRejectLine) {
			t.Fatalf("baseline edited but not staged must still reject; exit=%d stderr=%q", code, errb)
		}
	})

	brokenCarrier := func(t *testing.T) string {
		out := []string{}
		for _, ln := range strings.Split(guardCarrier(t), "\n") {
			if s := strings.TrimSpace(ln); s == acCounterBeginSentinel || s == acCounterEndSentinel {
				continue
			}
			out = append(out, ln)
		}
		return strings.Join(out, "\n")
	}
	noop := func(t *testing.T, r *guardRepo) {
		t.Helper()
		out, errb, code := r.check()
		if code != 0 || out != "" || errb != "" {
			t.Fatalf("no qualifying file: want silent exit 0; exit=%d stdout=%q stderr=%q", code, out, errb)
		}
	}
	t.Run("noop_unrelated", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(acLocalClausePath, brokenCarrier(t))
		r.stage("README.txt", "unrelated\n")
		noop(t, r)
	})
	t.Run("noop_archive", func(t *testing.T) {
		arch := ".moai/specs/_archive/SPEC-OLD-001/acceptance.md"
		r := seedCommon(t, map[string]string{arch: two})
		r.stage(acLocalClausePath, brokenCarrier(t))
		r.stage(arch, three)
		noop(t, r)
	})
	t.Run("noop_depth2", func(t *testing.T) {
		deep := ".moai/specs/a/b/acceptance.md"
		r := seedCommon(t, map[string]string{deep: two})
		r.stage(acLocalClausePath, brokenCarrier(t))
		r.stage(deep, three)
		noop(t, r)
	})

	// Parity: every row of the corpus-test transition table, plus one row
	// whose ambiguous identifiers first appear in reverse sorted order.
	type parityRow struct {
		name      string
		tc        acComparisonTransitionCase
		firstSeen []string // identifier order as written into the file
	}
	rows := []parityRow{}
	for _, tc := range acComparisonTransitionCases {
		rows = append(rows, parityRow{name: tc.name, tc: tc, firstSeen: strings.Fields(tc.ids)})
	}
	rows = append(rows, parityRow{
		name:      "halt_ids_unsorted",
		tc:        acComparisonTransitionCase{want: acBaselineEntry{halt: true, haltIDs: "AC-SYN-002 AC-SYN-004"}, known: true, halted: true, ids: "AC-SYN-002 AC-SYN-004"},
		firstSeen: []string{"AC-SYN-004", "AC-SYN-002"},
	})
	for _, row := range rows {
		row := row
		t.Run("parity/"+row.name, func(t *testing.T) {
			tc := row.tc
			var b strings.Builder
			b.WriteString("# parity fixture\n")
			for i := 1; i <= tc.live; i++ {
				fmt.Fprintf(&b, "## AC-LIV-%03d live\n", i)
			}
			for i := 1; i <= tc.exc; i++ {
				fmt.Fprintf(&b, "## AC-EXC-%03d [RETIRED] excluded\n", i)
			}
			if tc.halted {
				for _, id := range row.firstSeen {
					fmt.Fprintf(&b, "- %s [RETIRED] marked occurrence\n", id)
				}
				for _, id := range row.firstSeen {
					fmt.Fprintf(&b, "- %s unmarked occurrence\n", id)
				}
			}
			records := []string{guardCountRecord(".moai/specs/SPEC-OTHER-001/acceptance.md", 1, 0)}
			if tc.known {
				if tc.want.halt {
					records = append(records, guardHaltRecord(guardSpecX, tc.want.haltIDs))
				} else {
					records = append(records, guardCountRecord(guardSpecX, tc.want.live, tc.want.excluded))
				}
			}
			r := newGuardRepo(t)
			r.write(acLocalClausePath, guardCarrier(t))
			r.copyGuardScripts()
			r.write(guardSpecX, "# seed\n")
			r.write(acBaselineSnapshotPath, guardBaseline(records...))
			r.git("add", "-A")
			r.git("commit", "-q", "-m", "seed")
			r.stage(guardSpecX, b.String())

			ids := strings.Fields(tc.ids)
			sort.Strings(ids)
			problem, _ := acComparison(tc.want, tc.known, tc.halted, strings.Join(ids, " "), tc.live, tc.exc)
			_, errb, code := r.check()
			if (code != 0) != (problem != "") {
				t.Fatalf("parity broken: checker exit=%d, acComparison problem=%q\nstderr=%s", code, problem, errb)
			}
			if code != 0 && !strings.Contains(errb, guardRejectLine) {
				t.Errorf("non-zero exit must be a measured rejection, got stderr=%q", errb)
			}
		})
	}

	// Tool faults fail open, loudly.
	carrier := guardCarrier(t)
	replaceLine := func(src, from, to string) string {
		lines := strings.Split(src, "\n")
		for i, ln := range lines {
			if strings.TrimSpace(ln) == from {
				lines[i] = to
			}
		}
		return strings.Join(lines, "\n")
	}
	faults := []struct {
		kind  string
		apply func(t *testing.T, r *guardRepo)
	}{
		{"sentinel_absent", func(t *testing.T, r *guardRepo) { r.stage(acLocalClausePath, brokenCarrier(t)) }},
		{"sentinel_duplicated", func(t *testing.T, r *guardRepo) {
			r.stage(acLocalClausePath, carrier+"\n   "+acCounterBeginSentinel+"\n   true\n   "+acCounterEndSentinel+"\n")
		}},
		{"end_before_begin", func(t *testing.T, r *guardRepo) {
			s := replaceLine(carrier, acCounterBeginSentinel, "@@PLACEHOLDER@@")
			s = replaceLine(s, acCounterEndSentinel, "   "+acCounterBeginSentinel)
			s = replaceLine(s, "@@PLACEHOLDER@@", "   "+acCounterEndSentinel)
			r.stage(acLocalClausePath, s)
		}},
		{"empty_body", func(t *testing.T, r *guardRepo) {
			out := []string{}
			inside := false
			for _, ln := range strings.Split(carrier, "\n") {
				s := strings.TrimSpace(ln)
				if s == acCounterBeginSentinel {
					inside = true
					out = append(out, ln)
					continue
				}
				if s == acCounterEndSentinel {
					inside = false
				}
				if !inside {
					out = append(out, ln)
				}
			}
			r.stage(acLocalClausePath, strings.Join(out, "\n"))
		}},
		{"baseline_absent", func(t *testing.T, r *guardRepo) {
			r.git("rm", "-q", "--cached", "--", acBaselineSnapshotPath)
		}},
		{"baseline_line_malformed", func(t *testing.T, r *guardRepo) {
			r.stage(acBaselineSnapshotPath, guardBaseline(guardSpecX+"  COUNT two  live=2 excluded=0 ambiguous=0"))
		}},
		{"counter_exit_2", func(t *testing.T, r *guardRepo) {
			r.stage(acLocalClausePath, replaceLine(carrier, acCounterEndSentinel, "   exit 2\n   "+acCounterEndSentinel))
		}},
	}
	for _, f := range faults {
		f := f
		t.Run("fault/"+f.kind, func(t *testing.T) {
			r := seedCommon(t, nil)
			r.stage(guardSpecX, three)
			f.apply(t, r)
			_, errb, code := r.check()
			if code != 0 {
				t.Fatalf("tool fault must fail open (exit 0); exit=%d stderr=%q", code, errb)
			}
			var hit string
			for _, ln := range nonEmptyLines(errb) {
				if strings.HasPrefix(ln, guardNotChecked) && strings.Contains(ln, guardSpecX) {
					hit = ln
				}
			}
			if hit == "" {
				t.Fatalf("fault must print a %q line naming the file; stderr=%q", guardNotChecked, errb)
			}
			t.Logf("fault line: %s", hit)
		})
	}

	t.Run("mixed/mismatch_plus_fault", func(t *testing.T) {
		specB := ".moai/specs/SPEC-B-001/acceptance.md"
		r := seedCommon(t, map[string]string{specB: two}, specB+"  COUNT two  live=2 excluded=0 ambiguous=0")
		r.stage(guardSpecX, three)
		r.stage(specB, twoProse)
		_, errb, code := r.check()
		if code == 0 || !strings.Contains(errb, guardRejectLine) {
			t.Fatalf("mismatch on A must reject even with a fault on B; exit=%d stderr=%q", code, errb)
		}
		var rejectA, faultB bool
		for _, ln := range nonEmptyLines(errb) {
			if strings.HasPrefix(ln, guardRejectLine) && strings.Contains(ln, guardSpecX) {
				rejectA = true
			}
			if strings.HasPrefix(ln, guardNotChecked) && strings.Contains(ln, specB) {
				faultB = true
			}
		}
		if !rejectA || !faultB {
			t.Fatalf("want REJECT line for A and NOT CHECKED line for B; stderr=%q", errb)
		}
	})

	t.Run("index_only/mm_unstaged_criterion", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(guardSpecX, twoProse)
		r.write(guardSpecX, three)
		_, errb, code := r.check()
		if code != 0 || !strings.Contains(errb, guardCheckedLine+" 1") {
			t.Fatalf("staged blob is count-unchanged; want pass; exit=%d stderr=%q", code, errb)
		}
	})
	t.Run("index_only/staged_criterion_reverted_tree", func(t *testing.T) {
		r := seedCommon(t, nil)
		r.stage(guardSpecX, three)
		r.write(guardSpecX, two)
		_, errb, code := r.check()
		if code == 0 || !strings.Contains(errb, guardRejectLine) {
			t.Fatalf("staged blob moves the count; working-tree revert must not hide it; exit=%d stderr=%q", code, errb)
		}
	})
	t.Run("index_only/counter_from_index", func(t *testing.T) {
		r := seedCommon(t, nil)
		mutated := strings.Replace(carrier, `BEGIN { prefixes = "AC" }`, `BEGIN { prefixes = "ZZ" }`, 1)
		if mutated == carrier {
			t.Fatalf("carrier no longer carries the default-prefix line; mutation would be a no-op")
		}
		r.stage(acLocalClausePath, mutated)
		r.write(acLocalClausePath, carrier) // working tree == HEAD, unmutated
		r.stage(guardSpecX, twoProse)
		_, errb, code := r.check()
		if code == 0 || !strings.Contains(errb, guardRejectLine) || !strings.Contains(errb, "live=0") {
			t.Fatalf("measured count must follow the STAGED carrier (live=0); exit=%d stderr=%q", code, errb)
		}
	})

	t.Run("hostile_path", func(t *testing.T) {
		seg := "SPEC-$(touch${IFS}pwned).[x-001"
		decoy := strings.Replace(seg, ".", "Q", 1)
		hostile := ".moai/specs/" + seg + "/acceptance.md"
		decoyPath := ".moai/specs/" + decoy + "/acceptance.md"
		// The decoy record comes LAST and records 3: a regex lookup ('.'
		// matching 'Q') would pick it and wrongly pass the 2 -> 3 move.
		r := seedCommon(t, map[string]string{hostile: two},
			guardCountRecord(hostile, 2, 0), guardCountRecord(decoyPath, 3, 0))
		r.stage(hostile, three)
		_, errb, code := r.check()
		var pwned []string
		_ = filepath.WalkDir(filepath.Dir(r.dir), func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.Name() == "pwned" {
				pwned = append(pwned, p)
			}
			return nil
		})
		if len(pwned) != 0 {
			t.Fatalf("hostile path text reached a shell: %v", pwned)
		}
		if code == 0 {
			t.Fatalf("must reject the hostile-path amendment against the RIGHT record; stderr=%q", errb)
		}
		if !strings.Contains(errb, hostile) || !strings.Contains(errb, "live=2") {
			t.Fatalf("reject must name the exact path and the right record's value (live=2); stderr=%q", errb)
		}
	})
}

// writeTempFile writes content at an absolute path that the caller has built
// under t.TempDir() (a temp repo, its .git, or a per-test HOME), creating the
// parent directory. Every guard-test write outside guardRepo.write goes
// through here.
func writeTempFile(t *testing.T, absPath string, content []byte, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", absPath, err)
	}
	if err := os.WriteFile(absPath, content, perm); err != nil {
		t.Fatalf("write %s: %v", absPath, err)
	}
}

func fileSHA(t *testing.T, p string) string {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func guardCommitSubtests(t *testing.T) {
	three := guardFixture(t, "three.md")

	installed := func(t *testing.T) *guardRepo {
		t.Helper()
		requireRealCommitGit(t)
		r := seedCommon(t, nil)
		_, errb, code := r.runIn(r.dir, nil, "sh", guardInstallRel)
		if code != 0 {
			t.Fatalf("installer failed: exit=%d stderr=%q", code, errb)
		}
		return r
	}
	commitRejected := func(t *testing.T, r *guardRepo, dir string, args ...string) {
		t.Helper()
		before := strings.TrimSpace(mustOut(t, r, dir, "rev-parse", "HEAD"))
		_, errb, code := r.runIn(dir, nil, "git", args...)
		after := strings.TrimSpace(mustOut(t, r, dir, "rev-parse", "HEAD"))
		t.Logf("git %v in %s: exit=%d stderr=%q", args, filepath.Base(dir), code, errb)
		if code == 0 || before != after {
			t.Fatalf("commit must be rejected with HEAD unchanged; exit=%d HEAD %s -> %s", code, before, after)
		}
		if !strings.Contains(errb, guardRejectLine) {
			t.Fatalf("rejection must come from the guard; stderr=%q", errb)
		}
	}
	hookdirMarker := func(t *testing.T, r *guardRepo) string {
		t.Helper()
		marker := filepath.Join(r.home, "hookdir-ran")
		writeTempFile(t, filepath.Join(r.dir, ".git", "hooks", "pre-commit"), []byte("#!/bin/sh\ntouch '"+marker+"'\nexit 0\n"), 0o755)
		return marker
	}

	t.Run("commit/hookdir_present", func(t *testing.T) {
		r := installed(t)
		marker := hookdirMarker(t, r)
		r.stage(guardSpecX, three)
		commitRejected(t, r, r.dir, "commit", "-m", "x")
		_, err := os.Stat(marker)
		t.Logf("reject case: hookdir marker written=%v", err == nil)
		_ = os.Remove(marker)
		r.stage(acBaselineSnapshotPath, guardBaseline(guardCountRecord(guardSpecX, 3, 0)))
		before := r.head()
		_, errb, code := r.runIn(r.dir, nil, "git", "commit", "-m", "x")
		_, err = os.Stat(marker)
		t.Logf("pass case: exit=%d hookdir marker written=%v stderr=%q", code, err == nil, errb)
		if code != 0 || r.head() == before {
			t.Fatalf("amendment + staged baseline must commit; exit=%d stderr=%q", code, errb)
		}
	})

	t.Run("commit/hookspath_devnull", func(t *testing.T) {
		r := installed(t)
		r.git("config", "core.hooksPath", "/dev/null")
		r.stage(guardSpecX, three)
		commitRejected(t, r, r.dir, "commit", "-m", "x")
	})

	t.Run("commit/linked_worktree", func(t *testing.T) {
		r := installed(t)
		wt := filepath.Join(filepath.Dir(r.dir), "wt")
		r.git("worktree", "add", "-q", "-b", "wt", wt)
		writeTempFile(t, filepath.Join(wt, filepath.FromSlash(guardSpecX)), []byte(three), 0o644)
		mustOut(t, r, wt, "add", "--", guardSpecX)
		commitRejected(t, r, wt, "commit", "-m", "x")
	})

	t.Run("commit/all_flag", func(t *testing.T) {
		r := installed(t)
		r.write(guardSpecX, three) // left unstaged
		commitRejected(t, r, r.dir, "commit", "-a", "-m", "x")
	})

	t.Run("commit/pathspec_only", func(t *testing.T) {
		r := installed(t)
		r.stage(guardSpecX, three)
		r.stage(acBaselineSnapshotPath, guardBaseline(guardCountRecord(guardSpecX, 3, 0)))
		// The temporary index holds HEAD's baseline plus only the named path.
		commitRejected(t, r, r.dir, "commit", "-m", "x", guardSpecX)
	})

	t.Run("install/idempotent", func(t *testing.T) {
		r := installed(t)
		_, errb, code := r.runIn(r.dir, nil, "sh", guardInstallRel)
		if code != 0 {
			t.Fatalf("second install failed: exit=%d stderr=%q", code, errb)
		}
		events := nonEmptyLines(r.git("config", "--get-all", "hook.ac-baseline-guard.event"))
		cmds := nonEmptyLines(r.git("config", "--get-all", "hook.ac-baseline-guard.command"))
		if len(events) != 1 || events[0] != "pre-commit" || len(cmds) != 1 {
			t.Fatalf("want exactly one event=pre-commit and one command; events=%q commands=%q", events, cmds)
		}
		t.Logf("installed command: %s", cmds[0])
	})

	t.Run("install/old_git", func(t *testing.T) {
		requireRealCommitGit(t)
		r := seedCommon(t, nil)
		stub := filepath.Join(r.home, "stubbin")
		writeTempFile(t, filepath.Join(stub, "git"), []byte("#!/bin/sh\necho 'git version 2.53.0'\n"), 0o755)
		_, errb, code := r.runIn(r.dir, []string{"PATH=" + stub + string(os.PathListSeparator) + os.Getenv("PATH")}, "sh", guardInstallRel)
		if code == 0 || !strings.Contains(errb, "2.53.0") || !strings.Contains(errb, "2.54") {
			t.Fatalf("old git must be refused naming found and required versions; exit=%d stderr=%q", code, errb)
		}
		out, _, gcode := r.runIn(r.dir, nil, "git", "config", "--get-regexp", `^hook\.`)
		if gcode != 1 || out != "" {
			t.Fatalf("installer must write nothing on old git; exit=%d out=%q", gcode, out)
		}
	})

	t.Run("install/missing_script", func(t *testing.T) {
		requireRealCommitGit(t)
		r := newGuardRepo(t)
		r.write("seed.txt", "seed\n")
		r.git("add", "seed.txt")
		r.git("commit", "-q", "-m", "seed")
		_, errb, code := r.runIn(r.dir, nil, "sh", filepath.Join(repoRoot(t), guardInstallRel))
		if code != 0 {
			t.Fatalf("installer failed: exit=%d stderr=%q", code, errb)
		}
		r.stage("seed.txt", "seed 2\n")
		_, errb, code = r.runIn(r.dir, nil, "git", "commit", "-m", "x")
		if code != 0 || !strings.Contains(errb, guardNotChecked) {
			t.Fatalf("tree without the checker: want exit 0 with a %q line; exit=%d stderr=%q", guardNotChecked, code, errb)
		}
		t.Logf("missing-script line: %s", strings.TrimSpace(errb))
	})

	t.Run("install/managed_untouched", func(t *testing.T) {
		requireRealCommitGit(t)
		r := seedCommon(t, nil)
		hookdirMarker(t, r)
		shaFile := filepath.Join(r.dir, ".git", "hooks", ".moai-pre-commit.sha256")
		writeTempFile(t, shaFile, []byte("0000 fixture provenance\n"), 0o644)
		r.git("config", "core.hooksPath", "/dev/null")
		hook := filepath.Join(r.dir, ".git", "hooks", "pre-commit")
		hookBefore, shaBefore, pathBefore := fileSHA(t, hook), fileSHA(t, shaFile), r.git("config", "--get", "core.hooksPath")

		if _, errb, code := r.runIn(r.dir, nil, "sh", guardInstallRel); code != 0 {
			t.Fatalf("installer failed: %q", errb)
		}
		baselineAbs := filepath.Join(r.dir, filepath.FromSlash(acBaselineSnapshotPath))
		r.stage(guardSpecX, three)
		baseIdx, baseTree := r.git("rev-parse", ":"+acBaselineSnapshotPath), fileSHA(t, baselineAbs)
		commitRejected(t, r, r.dir, "commit", "-m", "x")
		if r.git("rev-parse", ":"+acBaselineSnapshotPath) != baseIdx || fileSHA(t, baselineAbs) != baseTree {
			t.Fatalf("checker must not write the baseline (index or working tree)")
		}
		r.stage(acBaselineSnapshotPath, guardBaseline(guardCountRecord(guardSpecX, 3, 0)))
		baseIdx, baseTree = r.git("rev-parse", ":"+acBaselineSnapshotPath), fileSHA(t, baselineAbs)
		before := r.head()
		if _, errb, code := r.runIn(r.dir, nil, "git", "commit", "-m", "x"); code != 0 || r.head() == before {
			t.Fatalf("passing commit failed: %q", errb)
		}
		if r.git("rev-parse", ":"+acBaselineSnapshotPath) != baseIdx || fileSHA(t, baselineAbs) != baseTree {
			t.Fatalf("checker must not write the baseline on the passing path")
		}
		if fileSHA(t, hook) != hookBefore || fileSHA(t, shaFile) != shaBefore || r.git("config", "--get", "core.hooksPath") != pathBefore {
			t.Fatalf("managed hook, provenance file, or core.hooksPath changed")
		}
	})
}
