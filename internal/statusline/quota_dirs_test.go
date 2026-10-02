// Package statusline tests for SPEC-QUOTA-RECORD-WORKTREES-001 M1 (card t1442):
// the multi-directory quota reading. A fresh reading that exists only in a
// linked worktree's record directory is seen (AC-QWR-001), the freshest record
// wins across directories (AC-QWR-002), the single-directory reading is
// unchanged (AC-QWR-003), an unusable worktree entry contributes nothing and
// breaks nothing (AC-QWR-004), the entry bound counts every examined entry
// (AC-QWR-005), the new code is offline, spawn-free, read-only, and
// non-recursive (AC-QWR-006), and a root without git metadata reads its own
// directory alone (AC-QWR-010, statusline half).
//
// Every fixture is built under t.TempDir(); the metadata entries are written by
// the test as <root>/.git/worktrees/<entry>/gitdir files, except the one
// subtest that runs the installed git (real_git_worktree_layout). Time is the
// injected qasT0; record files are fixtures with their modification time set per
// case, so no test sleeps or reads the wall clock for a verdict.
package statusline

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
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// qwrEnvBenchRoot names the repository root BenchmarkQWR_RealRoot measures; the
// benchmark is skipped while it is unset (the M5 timing data point, not a gate).
const qwrEnvBenchRoot = "MOAI_QWR_BENCH_ROOT"

// The signatures the SPEC fixes (AC-QWR-003 signature_unchanged): a change to
// any of them fails to compile here.
var (
	_ func(string, time.Time, time.Duration) QuotaAggregate   = AggregateQuota
	_ func([]string, time.Time, time.Duration) QuotaAggregate = AggregateQuotaDirs
	_ func(string, int) []string                              = QuotaStateDirs
)

// qwrRoot returns a temporary primary root whose .git is a directory with no
// worktrees directory yet.
func qwrRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// qwrPrimaryState is the root's own state directory.
func qwrPrimaryState(root string) string { return filepath.Join(root, ".moai", "state") }

// qwrEntry writes the metadata entry <root>/.git/worktrees/<entry>/gitdir with
// the given content and returns the entry directory.
func qwrEntry(t *testing.T, root, entry, content string) string {
	t.Helper()
	meta := filepath.Join(root, ".git", "worktrees", entry)
	if err := os.MkdirAll(meta, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(meta, "gitdir"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return meta
}

// qwrLink registers a linked worktree rooted at wt under the metadata entry
// entry: the worktree's .git file, and the entry's gitdir naming it. It returns
// the worktree's state directory.
func qwrLink(t *testing.T, root, entry, wt string) string {
	t.Helper()
	meta := filepath.Join(root, ".git", "worktrees", entry)
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: "+meta+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	qwrEntry(t, root, entry, filepath.Join(wt, ".git")+"\n")
	return filepath.Join(wt, ".moai", "state")
}

// qwrWorktreeDir is the conventional location of a worktree under the root.
func qwrWorktreeDir(root, name string) string { return filepath.Join(root, ".moai", "worktrees", name) }

// qwrRead is the evaluation's multi-directory reading: the enumeration, then the
// cross-directory aggregate, at the injected clock.
func qwrRead(root string, maxDirs int, now time.Time) QuotaAggregate {
	return AggregateQuotaDirs(QuotaStateDirs(root, maxDirs), now, qasMaxAge)
}

// qwrFixtureMtime writes a schema-3 record whose file modification time is
// chosen independently of its capture time.
func qwrFixtureMtime(t *testing.T, stateDir, id string, captured, mtime time.Time, five, seven *QuotaWindowRecord) {
	t.Helper()
	rec := fmt.Sprintf(`{"schema_version": %d, "session_id": %q, "writer_pid": 1, "captured_at": %q, "context_window_size": 200000, "tokens_used": 50000, "raw_pct": 25, "stage": "none", "band": "standard"`,
		contextUsageSchemaVersion, id, qasStamp(captured))
	for _, w := range []struct {
		key string
		win *QuotaWindowRecord
	}{{"five_hour", five}, {"seven_day", seven}} {
		if w.win != nil {
			rec += fmt.Sprintf(`, %q: {"used_percentage": %v, "resets_at": %d}`, w.key, w.win.UsedPercentage, w.win.ResetsAt)
		}
	}
	rec += "}\n"
	qasFixtureBytes(t, stateDir, id, mtime, []byte(rec))
}

// AC-QWR-001 — a reading that exists only in a linked worktree's directory is
// seen, in every layout git and `moai cc -w` produce.
func TestQWR_AC001_WorktreeOnlyRecordIsSeen(t *testing.T) {
	now := qasT0

	// seen builds a root whose primary directory holds only a stale record and
	// registers one worktree holding the one fresh record, then reads.
	seen := func(t *testing.T, root, entry, wt string) {
		t.Helper()
		qasFixture(t, qwrPrimaryState(root), "sess-primary-stale", now.Add(-2*time.Hour), qasWin(99, time.Hour), nil)
		state := qwrLink(t, root, entry, wt)
		qasFixture(t, state, "sess-worktree", now.Add(-time.Minute), qasWin(62.5, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 62.5)
	}

	t.Run("moai_worktrees_layout", func(t *testing.T) {
		root := qwrRoot(t)
		seen(t, root, "t1442", filepath.Join(root, ".moai", "worktrees", "t1442"))
	})
	t.Run("claude_worktrees_layout", func(t *testing.T) {
		root := qwrRoot(t)
		seen(t, root, "t1442", filepath.Join(root, ".claude", "worktrees", "t1442"))
	})
	t.Run("outside_root_layout", func(t *testing.T) {
		root := qwrRoot(t)
		seen(t, root, "t1442", filepath.Join(t.TempDir(), "elsewhere", "t1442"))
	})
	t.Run("entry_name_differs_from_directory", func(t *testing.T) {
		root := qwrRoot(t)
		seen(t, root, "a1", qwrWorktreeDir(root, "b1"))
	})
	t.Run("real_git_worktree_layout", func(t *testing.T) {
		gitPath, err := exec.LookPath("git")
		if err != nil {
			t.Skip("git is absent from PATH: the on-disk layout guard cannot run (a Gap, not a pass)")
		}
		// Clear every GIT_* variable the host may carry so the installed git
		// reads only the temporary repository below.
		for _, kv := range os.Environ() {
			if k, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(k, "GIT_") {
				t.Setenv(k, "")
				_ = os.Unsetenv(k)
			}
		}
		t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
		t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

		repo := t.TempDir()
		run := func(args ...string) {
			t.Helper()
			full := append([]string{"-C", repo, "-c", "user.name=qwr", "-c", "user.email=qwr@example.invalid"}, args...)
			if out, err := exec.Command(gitPath, full...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", full, err, out)
			}
		}
		run("init", "-q")
		run("commit", "--allow-empty", "-q", "-m", "init")
		wt := filepath.Join(t.TempDir(), "real-wt")
		run("worktree", "add", "-q", "-b", "qwr-real", wt)

		qasFixture(t, filepath.Join(wt, ".moai", "state"), "sess-worktree", now.Add(-time.Minute), qasWin(62.5, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(repo, 128, now).FiveHour, QuotaFresh, 62.5)
	})
}

// AC-QWR-002 — the freshest record wins across directories: not the maximum,
// not the minimum, not the first or last directory, per window; skew and age
// apply to a worktree record exactly as to a primary one.
func TestQWR_AC002_FreshestAcrossDirectories(t *testing.T) {
	now := qasT0

	// two builds a root with one registered worktree and returns both state dirs.
	two := func(t *testing.T) (root, primary, worktree string) {
		t.Helper()
		root = qwrRoot(t)
		return root, qwrPrimaryState(root), qwrLink(t, root, "wt1", qwrWorktreeDir(root, "wt1"))
	}

	t.Run("stale_primary_copy_fresh_worktree_copy", func(t *testing.T) {
		root, primary, worktree := two(t)
		qasFixture(t, primary, "sid", now.Add(-20*time.Minute), qasWin(95, time.Hour), nil)
		qasFixture(t, worktree, "sid", now.Add(-2*time.Minute), qasWin(60, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 60)
	})
	t.Run("newer_primary_copy_wins", func(t *testing.T) {
		root, primary, worktree := two(t)
		qasFixture(t, primary, "sid", now.Add(-2*time.Minute), qasWin(60, time.Hour), nil)
		qasFixture(t, worktree, "sid", now.Add(-20*time.Minute), qasWin(95, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 60)
	})
	t.Run("not_max_across_dirs", func(t *testing.T) {
		root, primary, worktree := two(t)
		qasFixture(t, primary, "older", now.Add(-20*time.Minute), qasWin(97, time.Hour), nil)
		qasFixture(t, worktree, "newer", now.Add(-5*time.Minute), qasWin(70, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 70)
	})
	t.Run("not_min_across_dirs", func(t *testing.T) {
		root, primary, worktree := two(t)
		qasFixture(t, primary, "older", now.Add(-20*time.Minute), qasWin(40, time.Hour), nil)
		qasFixture(t, worktree, "newer", now.Add(-5*time.Minute), qasWin(80, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 80)
	})
	t.Run("per_window_across_dirs", func(t *testing.T) {
		// A newer record carrying only the seven-day window must not hide an
		// older fresh record's five-hour window held in the other directory.
		root, primary, worktree := two(t)
		qasFixture(t, primary, "newer", now.Add(-time.Minute), nil, qasWin(30, 24*time.Hour))
		qasFixture(t, worktree, "older", now.Add(-20*time.Minute), qasWin(55, time.Hour), nil)
		got := qwrRead(root, 128, now)
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 55)
		qasExpect(t, "seven_day", got.SevenDay, QuotaFresh, 30)
	})
	t.Run("tie_keeps_primary", func(t *testing.T) {
		// Equal capture time, different percentages: the primary directory is
		// listed first and keeps the tie. The worktree record also carries the
		// seven-day window alone, which only a reading that actually enumerated
		// the worktree can show.
		root, primary, worktree := two(t)
		tie := now.Add(-3 * time.Minute)
		qasFixture(t, primary, "sid-primary", tie, qasWin(70, time.Hour), nil)
		qasFixture(t, worktree, "sid-worktree", tie, qasWin(80, time.Hour), qasWin(33, 24*time.Hour))
		got := qwrRead(root, 128, now)
		qasExpect(t, "five_hour", got.FiveHour, QuotaFresh, 70)
		qasExpect(t, "seven_day", got.SevenDay, QuotaFresh, 33)
	})
	t.Run("skew_and_age_apply_to_worktree_records", func(t *testing.T) {
		skew := config.QuotaClockSkewTolerance + time.Second
		t.Run("skew", func(t *testing.T) {
			// A worktree record captured beyond the skew tolerance contributes
			// nothing: the valid control record beside it (older) must be the
			// reading, not the "newest" skewed one.
			root, _, worktree := two(t)
			qasFixture(t, worktree, "skewed", now.Add(skew), qasWin(99, time.Hour), nil)
			qasFixture(t, worktree, "control", now.Add(-time.Minute), qasWin(61, time.Hour), nil)
			qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 61)
		})
		t.Run("age", func(t *testing.T) {
			// A worktree record captured 30m1s before now (its file modified now,
			// so only the capture-time rule excludes it) contributes nothing; the
			// fresh seven-day-only control proves the worktree was read at all.
			root, _, worktree := two(t)
			qwrFixtureMtime(t, worktree, "aged", now.Add(-qasMaxAge-time.Second), now, qasWin(98, time.Hour), nil)
			qasFixture(t, worktree, "control", now.Add(-time.Minute), nil, qasWin(30, 24*time.Hour))
			got := qwrRead(root, 128, now)
			qasExpect(t, "seven_day", got.SevenDay, QuotaFresh, 30)
			qasExpect(t, "five_hour", got.FiveHour, QuotaUnknown, 0)
		})
	})
}

// AC-QWR-003 — the single-directory entry point and its result are unchanged.
func TestQWR_AC003_SingleDirectoryUnchanged(t *testing.T) {
	now := qasT0

	t.Run("single_dir_equals_multi_dir_of_one", func(t *testing.T) {
		scenarios := []struct {
			name string
			fill func(t *testing.T, s string)
		}{
			{"freshest_not_max", func(t *testing.T, s string) {
				qasFixture(t, s, "older", now.Add(-20*time.Minute), qasWin(95, time.Hour), nil)
				qasFixture(t, s, "newer", now.Add(-5*time.Minute), qasWin(60, time.Hour), nil)
			}},
			{"per_window", func(t *testing.T, s string) {
				qasFixture(t, s, "older", now.Add(-20*time.Minute), qasWin(55, time.Hour), nil)
				qasFixture(t, s, "newer", now.Add(-time.Minute), nil, qasWin(30, 24*time.Hour))
			}},
			{"only_stale", func(t *testing.T, s string) {
				qasFixture(t, s, "a", now.Add(-time.Hour), qasWin(99, 2*time.Hour), qasWin(99, 48*time.Hour))
			}},
			{"reset_window", func(t *testing.T, s string) {
				qasFixture(t, s, "rolled", now.Add(-time.Minute), qasWin(99, -time.Second), nil)
			}},
			{"source_times_and_exhausted", func(t *testing.T, s string) {
				win := qasWin(100, time.Hour)
				win.ExhaustedAt = qasStamp(now.Add(-10 * time.Minute))
				qasFixture(t, s, "full", now.Add(-4*time.Minute), win, nil)
			}},
			{"future_beyond_tolerance", func(t *testing.T, s string) {
				qasFixture(t, s, "skewed", now.Add(config.QuotaClockSkewTolerance+time.Second), qasWin(99, time.Hour), nil)
			}},
			{"unparseable_beside_valid", func(t *testing.T, s string) {
				qasFixtureBytes(t, s, "broken", now.Add(-time.Minute), []byte("{not json"))
				qasFixture(t, s, "good", now.Add(-2*time.Minute), qasWin(64, time.Hour), nil)
			}},
			{"dir_missing", func(t *testing.T, s string) {}},
		}
		fresh := 0
		for _, sc := range scenarios {
			dir := filepath.Join(t.TempDir(), ".moai", "state")
			sc.fill(t, dir)
			want := AggregateQuota(dir, now, qasMaxAge)
			got := AggregateQuotaDirs([]string{dir}, now, qasMaxAge)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: AggregateQuotaDirs([dir]) = %+v, want the single-directory reading %+v", sc.name, got, want)
			}
			if want.FiveHour.State == QuotaFresh || want.SevenDay.State == QuotaFresh {
				fresh++
			}
		}
		if fresh < 4 {
			t.Fatalf("only %d scenarios read a fresh window; the comparison would be vacuous", fresh)
		}
	})

	t.Run("signature_unchanged", func(t *testing.T) {
		// The compile-time assignments at the top of this file fix the three
		// signatures; this subtest states them at run time as well.
		for name, c := range map[string]struct {
			fn   any
			want string
		}{
			"AggregateQuota":     {AggregateQuota, "func(string, time.Time, time.Duration) statusline.QuotaAggregate"},
			"AggregateQuotaDirs": {AggregateQuotaDirs, "func([]string, time.Time, time.Duration) statusline.QuotaAggregate"},
			"QuotaStateDirs":     {QuotaStateDirs, "func(string, int) []string"},
		} {
			if got := reflect.TypeOf(c.fn).String(); got != c.want {
				t.Errorf("%s has type %s, want %s", name, got, c.want)
			}
		}
	})

	t.Run("capture_age_independent_of_mtime", func(t *testing.T) {
		// File mtime at now, capture time 30m1s earlier: unknown. The control —
		// exactly the max age — is fresh, so the unknown is the capture rule, not
		// a broken fixture. Pins the line the scan-loop refactor moves.
		for _, read := range []struct {
			name string
			fn   func(dir string) QuotaAggregate
		}{
			{"single", func(dir string) QuotaAggregate { return AggregateQuota(dir, now, qasMaxAge) }},
			{"multi", func(dir string) QuotaAggregate { return AggregateQuotaDirs([]string{dir}, now, qasMaxAge) }},
		} {
			stale := filepath.Join(t.TempDir(), "state")
			qwrFixtureMtime(t, stale, "aged", now.Add(-qasMaxAge-time.Second), now, qasWin(98, time.Hour), nil)
			qasExpect(t, read.name+"/aged_30m1s", read.fn(stale).FiveHour, QuotaUnknown, 0)

			edge := filepath.Join(t.TempDir(), "state")
			qwrFixtureMtime(t, edge, "edge", now.Add(-qasMaxAge), now, qasWin(98, time.Hour), nil)
			qasExpect(t, read.name+"/aged_exactly_max", read.fn(edge).FiveHour, QuotaFresh, 98)
		}
	})
}

// AC-QWR-004 — an unusable worktree entry contributes nothing and breaks
// nothing: every subtest carries a valid entry with a fresh record (the positive
// control) beside one broken entry and expects the valid entry's reading.
func TestQWR_AC004_UnusableEntriesContributeNothing(t *testing.T) {
	now := qasT0
	const controlPct = 60.0

	// build makes the root with the valid entry "ok", lets setup add the broken
	// entry, reads, and asserts the five-hour reading is want.
	build := func(t *testing.T, setup func(t *testing.T, root string), want float64) {
		t.Helper()
		root := qwrRoot(t)
		ok := qwrLink(t, root, "ok", qwrWorktreeDir(root, "ok"))
		qasFixture(t, ok, "sess-ok", now.Add(-time.Minute), qasWin(controlPct, time.Hour), nil)
		if setup != nil {
			setup(t, root)
		}
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, want)
	}
	// validGitdir is an absolute gitdir first line for a worktree directory that
	// has no usable record, so a lenient parse of a broken entry stays harmless.
	validGitdir := func(root, name string) string {
		return filepath.Join(qwrWorktreeDir(root, name), ".git") + "\n"
	}

	t.Run("gitdir_unreadable", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("file modes do not make a file unreadable on Windows (a Gap there, not a pass)")
		}
		if os.Geteuid() == 0 {
			t.Skip("running as root: mode 0 does not make a file unreadable (a Gap here, not a pass)")
		}
		build(t, func(t *testing.T, root string) {
			meta := qwrEntry(t, root, "bad", validGitdir(root, "bad"))
			if err := os.Chmod(filepath.Join(meta, "gitdir"), 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(filepath.Join(meta, "gitdir"), 0o644) })
		}, controlPct)
	})
	t.Run("gitdir_empty", func(t *testing.T) {
		build(t, func(t *testing.T, root string) { qwrEntry(t, root, "bad", "") }, controlPct)
	})
	t.Run("gitdir_garbage_not_a_path", func(t *testing.T) {
		build(t, func(t *testing.T, root string) { qwrEntry(t, root, "bad", "%%% not a path %%%\n") }, controlPct)
	})
	t.Run("gitdir_relative_path", func(t *testing.T) {
		build(t, func(t *testing.T, root string) { qwrEntry(t, root, "bad", "relative/worktree/.git\n") }, controlPct)
	})
	t.Run("gitdir_oversized", func(t *testing.T) {
		build(t, func(t *testing.T, root string) {
			qwrEntry(t, root, "bad", strings.Repeat("x", 1<<20))
		}, controlPct)
	})
	t.Run("gitdir_first_line_is_the_path", func(t *testing.T) {
		// The first line names worktree B (fresher than the control, 77%); the
		// second line names worktree C (fresher still, 10%) and must be ignored.
		build(t, func(t *testing.T, root string) {
			b := qwrLink(t, root, "line1", qwrWorktreeDir(root, "b"))
			qasFixture(t, b, "sess-b", now.Add(-30*time.Second), qasWin(77, time.Hour), nil)
			c := filepath.Join(qwrWorktreeDir(root, "c"), ".moai", "state")
			qasFixture(t, c, "sess-c", now.Add(-10*time.Second), qasWin(10, time.Hour), nil)
			qwrEntry(t, root, "line1", filepath.Join(qwrWorktreeDir(root, "b"), ".git")+"\n"+filepath.Join(qwrWorktreeDir(root, "c"), ".git")+"\n")
		}, 77)
	})
	t.Run("target_missing_pruned", func(t *testing.T) {
		build(t, func(t *testing.T, root string) { qwrEntry(t, root, "bad", validGitdir(root, "pruned")) }, controlPct)
	})
	t.Run("record_dir_absent", func(t *testing.T) {
		build(t, func(t *testing.T, root string) { qwrLink(t, root, "bad", qwrWorktreeDir(root, "bad")) }, controlPct)
	})
	t.Run("record_dir_is_a_file", func(t *testing.T) {
		build(t, func(t *testing.T, root string) {
			state := qwrLink(t, root, "bad", qwrWorktreeDir(root, "bad"))
			if err := os.MkdirAll(state, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(state, contextUsageDirName), []byte("not a directory"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, controlPct)
	})
	t.Run("record_dir_symlink_loop", func(t *testing.T) {
		build(t, func(t *testing.T, root string) {
			state := qwrLink(t, root, "bad", qwrWorktreeDir(root, "bad"))
			if err := os.MkdirAll(state, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(contextUsageDirName, filepath.Join(state, contextUsageDirName)); err != nil {
				t.Skipf("cannot create a symbolic link here (a Gap, not a pass): %v", err)
			}
		}, controlPct)
	})
	t.Run("locked_entry_is_still_read", func(t *testing.T) {
		// The locked worktree's record (77%, fresher than the control) IS read.
		build(t, func(t *testing.T, root string) {
			state := qwrLink(t, root, "locked", qwrWorktreeDir(root, "locked"))
			if err := os.WriteFile(filepath.Join(root, ".git", "worktrees", "locked", "locked"), []byte("agent tree\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			qasFixture(t, state, "sess-locked", now.Add(-30*time.Second), qasWin(77, time.Hour), nil)
		}, 77)
	})
	t.Run("worktrees_dir_unreadable", func(t *testing.T) {
		// <root>/.git/worktrees is a file: only the primary directory is read.
		root := qwrRoot(t)
		if err := os.WriteFile(filepath.Join(root, ".git", "worktrees"), []byte("not a directory"), 0o644); err != nil {
			t.Fatal(err)
		}
		qasFixture(t, qwrPrimaryState(root), "sess-primary", now.Add(-time.Minute), qasWin(55, time.Hour), nil)
		qasExpect(t, "five_hour", qwrRead(root, 128, now).FiveHour, QuotaFresh, 55)
	})
}

// qwrManyEntries registers n worktrees wt-000..wt-<n-1>; entry i holds one
// fresh record at i+1 percent whose capture time grows with i, so the freshest
// record among the examined entries is the highest-index one.
func qwrManyEntries(t *testing.T, root string, n int, now time.Time) {
	t.Helper()
	for i := range n {
		name := fmt.Sprintf("wt-%03d", i)
		state := qwrLink(t, root, name, qwrWorktreeDir(root, name))
		qasFixture(t, state, "sess-"+name, now.Add(-25*time.Minute+time.Duration(i)*10*time.Second), qasWin(float64(i+1), time.Hour), nil)
	}
}

// AC-QWR-005 — at most the given number of entries are examined, unusable ones
// included, and a gitdir file is read for at most 4 KiB.
func TestQWR_AC005_EnumerationBound(t *testing.T) {
	now := qasT0

	root := qwrRoot(t)
	qwrManyEntries(t, root, 130, now)

	t.Run("bound_128_entry_128_visible", func(t *testing.T) {
		qasExpect(t, "bound 128", qwrRead(root, 128, now).FiveHour, QuotaFresh, 128) // wt-127
		qasExpect(t, "bound 127", qwrRead(root, 127, now).FiveHour, QuotaFresh, 127) // wt-126: the 128th entry needs a bound of 128
	})
	t.Run("bound_128_entry_129_and_130_not_read", func(t *testing.T) {
		got := qwrRead(root, 128, now).FiveHour
		qasExpect(t, "bound 128", got, QuotaFresh, 128)
		if got.State == QuotaFresh && got.UsedPercentage > 128 {
			t.Errorf("the reading %v comes from an entry beyond the bound of 128", got.UsedPercentage)
		}
	})
	t.Run("bound_3_reads_three", func(t *testing.T) {
		qasExpect(t, "bound 3", qwrRead(root, 3, now).FiveHour, QuotaFresh, 3) // wt-002
	})
	t.Run("bound_1_reads_one", func(t *testing.T) {
		qasExpect(t, "bound 1", qwrRead(root, 1, now).FiveHour, QuotaFresh, 1) // wt-000
	})
	t.Run("unusable_entries_consume_the_bound", func(t *testing.T) {
		r := qwrRoot(t)
		qwrEntry(t, r, "wt-000", filepath.Join(qwrWorktreeDir(r, "gone"), ".git")+"\n") // target missing
		qwrEntry(t, r, "wt-001", "%%% not a path %%%\n")                                // bad gitdir
		for i, pct := range []float64{3, 4} {
			name := fmt.Sprintf("wt-%03d", i+2)
			state := qwrLink(t, r, name, qwrWorktreeDir(r, name))
			qasFixture(t, state, "sess-"+name, now.Add(-5*time.Minute+time.Duration(i)*4*time.Minute), qasWin(pct, time.Hour), nil)
		}
		qasExpect(t, "bound 3", qwrRead(r, 3, now).FiveHour, QuotaFresh, 3) // only wt-002 is examined and usable
		qasExpect(t, "bound 4", qwrRead(r, 4, now).FiveHour, QuotaFresh, 4) // control: wt-003 is usable once the bound allows
	})
	t.Run("gitdir_read_is_bounded", func(t *testing.T) {
		r := qwrRoot(t)
		// accepted: a valid first line followed by 1 MiB of filler.
		okState := qwrLink(t, r, "a-accepted", qwrWorktreeDir(r, "accepted"))
		qasFixture(t, okState, "sess-ok", now.Add(-5*time.Minute), qasWin(70, time.Hour), nil)
		qwrEntry(t, r, "a-accepted", filepath.Join(qwrWorktreeDir(r, "accepted"), ".git")+"\n"+strings.Repeat("x", 1<<20))
		// late: 4096 bytes of leading blanks, then the valid path; its record is
		// the freshest and the newest, so a read beyond 4 KiB would let it win.
		lateState := qwrLink(t, r, "b-late", qwrWorktreeDir(r, "late"))
		qasFixture(t, lateState, "sess-late", now.Add(-time.Minute), qasWin(99, time.Hour), nil)
		qwrEntry(t, r, "b-late", strings.Repeat(" ", 4096)+filepath.Join(qwrWorktreeDir(r, "late"), ".git")+"\n")
		qasExpect(t, "five_hour", qwrRead(r, 128, now).FiveHour, QuotaFresh, 70)
	})
}

// qwrForbiddenRefs returns the forbidden "importpath.Name" selectors a Go source
// references, resolving import aliases (src nil reads filename).
func qwrForbiddenRefs(t *testing.T, filename string, src any) []string {
	t.Helper()
	forbidden := map[string]bool{}
	for _, name := range []string{"Walk", "WalkDir"} {
		forbidden["path/filepath."+name] = true
	}
	for _, name := range []string{"WriteFile", "Create", "OpenFile", "Remove", "RemoveAll", "Rename", "Symlink", "Chmod", "Truncate", "Mkdir", "MkdirAll"} {
		forbidden["os."+name] = true
	}
	f, err := parser.ParseFile(token.NewFileSet(), filename, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	pkgs := map[string]string{} // local name -> import path
	for _, imp := range f.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		name := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		pkgs[name] = path
	}
	var hits []string
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok {
			if path, ok := pkgs[id.Name]; ok && forbidden[path+"."+sel.Sel.Name] {
				hits = append(hits, path+"."+sel.Sel.Name)
			}
		}
		return true
	})
	return hits
}

// qwrUnresolvedCalls returns the unqualified calls in a Go source that name a
// package-level function declared outside the allowed files (other than the one
// pre-existing reader), so a helper placed in a differently named file cannot be
// reached from the quota files unseen. Method values and interface calls are
// not followed — the stated residual of AC-QWR-006.
func qwrUnresolvedCalls(t *testing.T, filename string, src any, decls map[string][]string, allowedFiles map[string]bool) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), filename, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	var bad []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || id.Name == "ReadSessionTelemetry" {
			return true
		}
		files, declared := decls[id.Name]
		if !declared {
			return true // a builtin, a conversion, or a local value
		}
		for _, file := range files {
			if !allowedFiles[file] {
				bad = append(bad, id.Name+" (declared in "+file+")")
			}
		}
		return true
	})
	return bad
}

// AC-QWR-006 — the new code is offline, spawn-free, read-only, and
// non-recursive, and cannot hide outside the swept files.
func TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive(t *testing.T) {
	now := qasT0

	matches, err := filepath.Glob("quota*.go")
	if err != nil {
		t.Fatal(err)
	}
	var swept []string
	for _, m := range matches {
		if !strings.HasSuffix(m, "_test.go") {
			swept = append(swept, m)
		}
	}
	sweptSet := map[string]bool{}
	for _, f := range swept {
		sweptSet[filepath.Base(f)] = true
	}

	t.Run("swept_set_names_both_files", func(t *testing.T) {
		for _, name := range []string{"quota.go", "quota_dirs.go"} {
			if !sweptSet[name] {
				t.Errorf("the quota*.go sweep %v does not contain %s", swept, name)
			}
		}
	})

	t.Run("no_network_or_process_imports", func(t *testing.T) {
		forbidden := map[string]bool{"net": true, "net/http": true, "os/exec": true}
		for _, file := range swept {
			parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse %s: %v", file, err)
			}
			for _, imp := range parsed.Imports {
				if path, _ := strconv.Unquote(imp.Path.Value); forbidden[path] {
					t.Errorf("%s imports %q: the quota path must be offline and spawn-free", file, path)
				}
			}
		}
	})

	t.Run("no_write_walk_or_mutation_selectors", func(t *testing.T) {
		const control = `package control
import (
	"os"
	o "os"
	"path/filepath"
)
func f() {
	_ = filepath.Walk
	_ = filepath.WalkDir
	_ = os.WriteFile
	_ = os.Create
	_ = os.OpenFile
	_ = os.Remove
	_ = os.RemoveAll
	_ = os.Rename
	_ = os.Symlink
	_ = os.Chmod
	_ = os.Truncate
	_ = os.Mkdir
	_ = o.MkdirAll
}
`
		if hits := qwrForbiddenRefs(t, "control.go", control); len(hits) != 13 {
			t.Fatalf("positive control: the scan flagged %d of 13 forbidden selectors: %v", len(hits), hits)
		}
		for _, file := range swept {
			if hits := qwrForbiddenRefs(t, file, nil); len(hits) != 0 {
				t.Errorf("%s references forbidden selectors %v", file, hits)
			}
		}
	})

	t.Run("new_functions_declared_in_swept_files_only", func(t *testing.T) {
		decls := qwrPackageFuncs(t)
		for _, fn := range []string{"QuotaStateDirs", "AggregateQuotaDirs"} {
			files := decls[fn]
			if len(files) != 1 || !sweptSet[files[0]] {
				t.Errorf("%s is declared in %v, want exactly one declaration, in a swept quota*.go file", fn, files)
			}
		}
	})

	t.Run("calls_resolve_to_the_two_files_or_the_reader", func(t *testing.T) {
		const control = `package control
func g() { helper(); AggregateQuota(); ReadSessionTelemetry() }
`
		fake := map[string][]string{"helper": {"other.go"}, "AggregateQuota": {"quota.go"}, "ReadSessionTelemetry": {"context_usage.go"}}
		allowed := map[string]bool{"quota.go": true, "quota_dirs.go": true}
		if bad := qwrUnresolvedCalls(t, "control.go", control, fake, allowed); len(bad) != 1 {
			t.Fatalf("positive control: flagged %v, want exactly the call of the helper in other.go", bad)
		}
		decls := qwrPackageFuncs(t)
		for _, name := range []string{"quota.go", "quota_dirs.go"} {
			if bad := qwrUnresolvedCalls(t, name, nil, decls, allowed); len(bad) != 0 {
				t.Errorf("%s calls package-level functions declared elsewhere: %v", name, bad)
			}
		}
	})

	t.Run("run_leaves_every_tree_byte_identical", func(t *testing.T) {
		root := qwrRoot(t)
		other := t.TempDir()
		qasFixture(t, qwrPrimaryState(root), "sess-primary", now.Add(-20*time.Minute), qasWin(40, time.Hour), nil)
		wtA := qwrLink(t, root, "wt-a", qwrWorktreeDir(root, "a"))
		qasFixture(t, wtA, "sess-a", now.Add(-time.Minute), qasWin(62, time.Hour), qasWin(20, 24*time.Hour))
		wtB := qwrLink(t, root, "wt-b", filepath.Join(other, "b"))
		qasFixture(t, wtB, "sess-b", now.Add(-40*time.Minute), qasWin(10, time.Hour), nil)
		qasFixtureBytes(t, wtB, "broken", now.Add(-time.Minute), []byte("{nope"))
		if err := os.WriteFile(filepath.Join(root, ".git", "worktrees", "wt-b", "locked"), []byte("lock\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		beforeRoot, beforeOther := qasTreeDigest(t, root), qasTreeDigest(t, other)
		got := qwrRead(root, 128, now)
		// The positive control: the run exercised the worktree directories.
		qasExpect(t, "five_hour (the run must reach the worktree directories)", got.FiveHour, QuotaFresh, 62)
		if after := qasTreeDigest(t, root); !reflect.DeepEqual(beforeRoot, after) {
			t.Errorf("the run changed the primary root tree\nbefore: %v\n after: %v", beforeRoot, after)
		}
		if after := qasTreeDigest(t, other); !reflect.DeepEqual(beforeOther, after) {
			t.Errorf("the run changed the second tree\nbefore: %v\n after: %v", beforeOther, after)
		}
	})
}

// qwrPackageFuncs maps every package-level function declared in the package's
// non-test files to the base names of the files that declare it.
func qwrPackageFuncs(t *testing.T) map[string][]string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	decls := map[string][]string{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				decls[fd.Name.Name] = append(decls[fd.Name.Name], filepath.Base(file))
			}
		}
	}
	for name := range decls {
		sort.Strings(decls[name])
	}
	return decls
}

// AC-QWR-010 — a root without a git metadata directory reads its own directory
// alone, exactly as before the change (statusline half).
func TestQWR_AC010_NoGitMetadataReadsRootOnly(t *testing.T) {
	now := qasT0

	// check puts a fresh 55% record in the root's own directory and a fresher
	// 99% record in a would-be worktree directory the git metadata (when it
	// exists at all) would name, then expects the single-directory reading.
	check := func(t *testing.T, root string) {
		t.Helper()
		qasFixture(t, qwrPrimaryState(root), "sess-own", now.Add(-5*time.Minute), qasWin(55, time.Hour), nil)
		qasFixture(t, filepath.Join(qwrWorktreeDir(root, "x"), ".moai", "state"), "sess-wt", now.Add(-time.Minute), qasWin(99, time.Hour), nil)
		want := AggregateQuota(qwrPrimaryState(root), now, qasMaxAge)
		qasExpect(t, "the single-directory reading", want.FiveHour, QuotaFresh, 55)
		got := qwrRead(root, 128, now)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("reading = %+v, want the root's own single-directory reading %+v", got, want)
		}
	}

	t.Run("no_git_entry", func(t *testing.T) {
		check(t, t.TempDir())
	})
	t.Run("git_is_a_file", func(t *testing.T) {
		root := t.TempDir()
		common := filepath.Join(t.TempDir(), "common")
		wt := qwrWorktreeDir(root, "x")
		meta := filepath.Join(common, "worktrees", "x")
		if err := os.MkdirAll(meta, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(meta, "gitdir"), []byte(filepath.Join(wt, ".git")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: "+common+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		check(t, root)
	})
	t.Run("no_worktrees_dir", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		check(t, root)
	})
}

// BenchmarkQWR_RealRoot is the M5 timing data point (not a gate): one call of
// the multi-directory reading against a real repository root at the default
// bound, beside the single-directory call on the primary directory. It is
// skipped unless MOAI_QWR_BENCH_ROOT names a repository root.
func BenchmarkQWR_RealRoot(b *testing.B) {
	root := os.Getenv(qwrEnvBenchRoot)
	if root == "" {
		b.Skipf("%s is unset: no repository root to measure", qwrEnvBenchRoot)
	}
	now := time.Now()
	b.Run("multi_dir_bound_128", func(b *testing.B) {
		for b.Loop() {
			AggregateQuotaDirs(QuotaStateDirs(root, 128), now, qasMaxAge)
		}
	})
	b.Run("single_primary_dir", func(b *testing.B) {
		primary := qwrPrimaryState(root)
		for b.Loop() {
			AggregateQuota(primary, now, qasMaxAge)
		}
	})
}
