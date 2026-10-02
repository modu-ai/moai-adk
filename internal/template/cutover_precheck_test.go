// cutover_precheck_test.go: AC-GFD-019 (SPEC-GITHUB-FLOW-DEFAULT-001 M6, design
// D-8 step 3 and D-25) for scripts/cutover-precheck.sh, the read-only batch
// boundary precheck.
//
// Every external reader the script consults (integration window, slot leases,
// session registry, pid liveness, the queue, the card-to-branch map) is an
// injectable command; the fixtures replace each with a stub and put a poisoned
// `moai`/`gh` first on PATH, so a script that bypasses a seam fails the test.
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	cvoWindowFree   = "release-integration window: free\n"
	cvoWindowHeld   = "release-integration window: held\n  holder:   lane-2 (abc, pid 4242)\n  branch:   release/v9.9.9\n  worktree: /x/.moai/worktrees/release\n  since:    2026-10-02T22:00:00Z\n"
	cvoWindowStale  = "release-integration window: held by a session that is gone (reclaimable)\n  holder:   lane-2 (abc, pid 4242)\n"
	cvoSlotsNone    = "no slot leases recorded\n"
	cvoSlotsFree    = "slot go-test-cli-shared: free\n"
	cvoSlotHeld     = "slot go-test-cli-shared: held\n  session: s1 (pid 4243)\n  name:    (not given)\n  command: (not given)\n  since:   2026-10-02T22:39:25Z\n  bound:   30m0s, ends 2026-10-02T23:09:25Z\n"
	cvoSlotExpired  = "slot go-test-cli-shared: past its declared bound (reclaimable)\n  session: s1 (pid 4243)\n"
	cvoSlotGone     = "slot go-test-cli-shared: held by a session that is gone (reclaimable)\n  session: s1 (pid 4243)\n"
	cvoNoSessions   = "[]\n"
	cvoGtdHeader    = "todo: answered by the SQLite backlog store; the backlog.json beside it is NOT the queue\n"
	cvoQueuedOnly   = cvoGtdHeader + "t1348\tqueued\tfixture queued card\n"
	cvoCardBranchOK = "wt-merged"
)

func cvoSessionsJSON(entries ...string) string {
	return "[\n" + strings.Join(entries, ",\n") + "\n]\n"
}

func cvoSession(pid int, cwd string) string {
	return fmt.Sprintf("  {\n    \"session_id\": \"sess-%d\",\n    \"spec_id\": \"(none)\",\n    \"phase\": \"(none)\",\n    \"started_at\": \"2026-10-02T04:00:00Z\",\n    \"last_heartbeat\": \"2026-10-02T04:05:00Z\",\n    \"pid\": %d,\n    \"host\": \"fixture.local\",\n    \"cwd\": %q\n  }", pid, pid, cwd)
}

// cvoPre is a scratch origin + clone on `develop` plus one stub per reader.
type cvoPre struct {
	t      *testing.T
	repo   *rlsRepo
	dir    string
	poison *cvoPoison

	integration, slot, sessions, alive, gtd, branch string
}

// cvoNewPre builds a fully clear fixture: develop pushed, no holders, no lane
// sessions, no picked cards. Individual fixtures then break exactly one thing.
func cvoNewPre(t *testing.T) *cvoPre {
	t.Helper()
	rlsRequireTools(t)
	r := rlsNewRepo(t)
	r.git("checkout", "-q", "-b", "develop")
	r.commit(map[string]string{"README.md": "one\n"}, "one")
	r.git("push", "-q", "origin", "develop")
	p := &cvoPre{t: t, repo: r, dir: t.TempDir(), poison: cvoNewPoison(t)}
	p.setIntegration(cvoWindowFree, 0)
	p.setSlot(cvoSlotsNone, 0)
	p.setSessions(cvoNoSessions, 0)
	p.setAlive(true)
	p.setQueue(cvoQueuedOnly, map[string]string{})
	return p
}

func (p *cvoPre) setIntegration(out string, code int) {
	p.integration = cvoStubCmd(p.t, p.dir, "integration-status", out, code)
}

func (p *cvoPre) setSlot(out string, code int) {
	p.slot = cvoStubCmd(p.t, p.dir, "slot-status", out, code)
}

func (p *cvoPre) setSessions(out string, code int) {
	p.sessions = cvoStubCmd(p.t, p.dir, "session-list", out, code)
}

// setAlive makes every pid probe answer alive (true) or gone (false).
func (p *cvoPre) setAlive(alive bool) {
	code := 1
	if alive {
		code = 0
	}
	p.alive = rlsWriteScript(p.t, p.dir, "pid-alive", "#!/bin/sh\nexit "+fmt.Sprint(code)+"\n")
}

// setQueue installs the gtd list output and the card -> branch map. A card
// absent from the map prints nothing (an unresolvable branch).
func (p *cvoPre) setQueue(gtd string, branches map[string]string) {
	p.gtd = cvoStubCmd(p.t, p.dir, "gtd-list", gtd, 0)
	var b strings.Builder
	b.WriteString("#!/bin/sh\ncase \"$1\" in\n")
	for card, br := range branches {
		fmt.Fprintf(&b, "  %s) echo %s ;;\n", card, br)
	}
	b.WriteString("  *) : ;;\nesac\nexit 0\n")
	p.branch = rlsWriteScript(p.t, p.dir, "card-branch", b.String())
}

// run executes the precheck inside the clone and asserts the poison is untouched.
func (p *cvoPre) run(args ...string) rlsResult {
	p.t.Helper()
	env := p.poison.env(
		"CUTOVER_INTEGRATION_STATUS_CMD="+p.integration,
		"CUTOVER_SLOT_STATUS_CMD="+p.slot,
		"CUTOVER_SESSION_LIST_CMD="+p.sessions,
		"CUTOVER_PID_ALIVE_CMD="+p.alive,
		"CUTOVER_GTD_LIST_CMD="+p.gtd,
		"CUTOVER_CARD_BRANCH_CMD="+p.branch,
	)
	res := rlsRun(p.t, p.repo.work, env, "bash", append([]string{cvoScript(p.t, cvoPrecheckRel)}, args...)...)
	p.poison.assertUntouched(p.t)
	p.t.Logf("exit=%d\n%s", res.exit, rlsNorm(res.out))
	return res
}

// branchAhead creates a branch with one commit that origin/develop lacks.
func (p *cvoPre) branchAhead(name string) {
	p.t.Helper()
	p.repo.git("checkout", "-q", "-b", name, "develop")
	p.repo.commit(map[string]string{name + ".txt": name + "\n"}, name)
	p.repo.git("checkout", "-q", "develop")
}

// branchMerged creates a branch whose tip is already in origin/develop.
func (p *cvoPre) branchMerged(name string) {
	p.t.Helper()
	p.repo.git("branch", name, "develop")
}

func cvoWantFail(t *testing.T, res rlsResult, token string, also ...string) {
	t.Helper()
	if res.exit == 0 {
		t.Errorf("exit code = 0, want non-zero naming %q", token)
	}
	out := rlsNorm(res.out)
	rlsMustContain(t, out, append([]string{"FAIL " + token}, also...)...)
}

func cvoWantClear(t *testing.T, res rlsResult, also ...string) {
	t.Helper()
	if res.exit != 0 {
		t.Errorf("exit code = %d, want 0", res.exit)
	}
	out := rlsNorm(res.out)
	if strings.Contains(out, "FAIL ") {
		t.Errorf("a clear fixture printed a FAIL line:\n%s", out)
	}
	rlsMustContain(t, out, append([]string{"PRECHECK_CLEAR"}, also...)...)
}

// TestCutoverPrecheck is the AC-GFD-019 matrix: four negative fixtures that each
// name their violated condition, two positive fixtures (a fully clear tree, and
// the card's own branch exempted by an explicit --exclude-card, design D-25).
func TestCutoverPrecheck(t *testing.T) {
	t.Run("negative_1_unpushed_develop_commit", func(t *testing.T) {
		p := cvoNewPre(t)
		p.repo.commit(map[string]string{"README.md": "two\n"}, "unpushed")
		cvoWantFail(t, p.run(), "unpushed-develop-commits", "1")
	})

	t.Run("negative_2_live_window_or_slot_holder", func(t *testing.T) {
		w := cvoNewPre(t)
		w.setIntegration(cvoWindowHeld, 0)
		cvoWantFail(t, w.run(), "live-integration-window", "lane-2")

		s := cvoNewPre(t)
		s.setSlot(cvoSlotHeld, 0)
		cvoWantFail(t, s.run(), "live-slot-holder", "go-test-cli-shared")
	})

	t.Run("negative_3_picked_card_unmerged_or_unpushed", func(t *testing.T) {
		// Tip not in origin/develop at all.
		a := cvoNewPre(t)
		a.branchAhead("wt-unmerged")
		a.setQueue(cvoGtdHeader+"t9001\tpicked\tfixture picked card\n", map[string]string{"t9001": "wt-unmerged"})
		cvoWantFail(t, a.run(), "unmerged-picked-card", "t9001")

		// Merged into local develop but never pushed: the tip is not an ancestor
		// of origin/develop, so the card still counts as not landed.
		b := cvoNewPre(t)
		b.repo.git("checkout", "-q", "-b", "wt-merged-local", "develop")
		b.repo.commit(map[string]string{"b.txt": "b\n"}, "card work")
		b.repo.git("checkout", "-q", "develop")
		b.repo.git("merge", "-q", "--no-ff", "-m", "merge card locally", "wt-merged-local")
		b.setQueue(cvoGtdHeader+"t9002\tpicked\tfixture picked card\n", map[string]string{"t9002": "wt-merged-local"})
		res := b.run()
		cvoWantFail(t, res, "unmerged-picked-card", "t9002")
		// ... and the unpushed merge is named as well.
		rlsMustContain(t, rlsNorm(res.out), "FAIL unpushed-develop-commits")
	})

	t.Run("negative_4_active_lane_session", func(t *testing.T) {
		p := cvoNewPre(t)
		p.setSessions(cvoSessionsJSON(cvoSession(4242, "/x/moai/.moai/worktrees/t9003")), 0)
		p.setAlive(true)
		cvoWantFail(t, p.run(), "active-lane-session", "t9003")
	})

	t.Run("positive_5_every_condition_met", func(t *testing.T) {
		p := cvoNewPre(t)
		p.branchMerged(cvoCardBranchOK)
		p.setQueue(cvoGtdHeader+"t9004\tpicked\tmerged and pushed\nt1348\tqueued\tnot picked\n", map[string]string{"t9004": cvoCardBranchOK})
		p.setIntegration(cvoWindowFree, 0)
		p.setSlot(cvoSlotsFree, 0)
		cvoWantClear(t, p.run())
	})

	t.Run("positive_6_own_card_exempted_by_explicit_exclusion", func(t *testing.T) {
		p := cvoNewPre(t)
		p.branchAhead("WT-github-flow-default") // this card's branch: a post-merge commit, not an ancestor
		p.setQueue(cvoGtdHeader+"t1453\tpicked\tthe cutover card itself\n", map[string]string{"t1453": "WT-github-flow-default"})

		// With the exclusion the precheck is clear and the exclusion is printed by name.
		cvoWantClear(t, p.run("--exclude-card", "t1453"), "EXCLUDED card t1453")

		// Without it the same fixture fails on the picked-card condition: the
		// exemption is never implicit.
		cvoWantFail(t, p.run(), "unmerged-picked-card", "t1453")
	})
}

// TestCutoverPrecheckGuards pins the edges the six fixtures do not: what the
// exclusion may not mask, fail-closed readers, stale holders and stale sessions
// that are not violations, and the read-only property.
func TestCutoverPrecheckGuards(t *testing.T) {
	t.Run("the_exclusion_is_one_card_only", func(t *testing.T) {
		p := cvoNewPre(t)
		res := p.run("--exclude-card", "t1453", "--exclude-card", "t9999")
		if res.exit != 2 {
			t.Errorf("a second --exclude-card must be a usage error (exit 2), got %d", res.exit)
		}
		rlsMustContain(t, rlsNorm(res.out), "--exclude-card")
	})

	t.Run("the_exclusion_does_not_mask_another_picked_card", func(t *testing.T) {
		p := cvoNewPre(t)
		p.branchAhead("wt-own")
		p.branchAhead("wt-other")
		p.setQueue(cvoGtdHeader+"t1453\tpicked\town\nt9005\tpicked\tother\n",
			map[string]string{"t1453": "wt-own", "t9005": "wt-other"})
		res := p.run("--exclude-card", "t1453")
		cvoWantFail(t, res, "unmerged-picked-card", "t9005")
		if strings.Contains(rlsNorm(res.out), "FAIL unmerged-picked-card: t1453") {
			t.Errorf("the excluded card must not be reported as a violation:\n%s", rlsNorm(res.out))
		}
	})

	t.Run("the_exclusion_does_not_mask_unpushed_commits_or_holders", func(t *testing.T) {
		p := cvoNewPre(t)
		p.repo.commit(map[string]string{"README.md": "x\n"}, "unpushed")
		p.setIntegration(cvoWindowHeld, 0)
		res := p.run("--exclude-card", "t1453")
		cvoWantFail(t, res, "unpushed-develop-commits")
		cvoWantFail(t, res, "live-integration-window")
	})

	t.Run("an_unresolvable_picked_card_branch_fails_closed", func(t *testing.T) {
		p := cvoNewPre(t)
		p.setQueue(cvoGtdHeader+"t9006\tpicked\tno branch known\n", map[string]string{})
		cvoWantFail(t, p.run(), "unmerged-picked-card", "t9006", "no resolvable branch")
	})

	t.Run("an_unreadable_reader_fails_closed_by_name", func(t *testing.T) {
		rows := []struct {
			name  string
			set   func(*cvoPre)
			token string
		}{
			{"integration", func(p *cvoPre) { p.setIntegration("boom\n", 1) }, "reader-unavailable"},
			{"slot", func(p *cvoPre) { p.setSlot("boom\n", 1) }, "reader-unavailable"},
			{"sessions", func(p *cvoPre) { p.setSessions("boom\n", 1) }, "reader-unavailable"},
			{"queue", func(p *cvoPre) { p.setQueue("boom\n", nil); p.gtd = cvoStubCmd(t, p.dir, "gtd-fail", "boom\n", 1) }, "reader-unavailable"},
		}
		for _, row := range rows {
			t.Run(row.name, func(t *testing.T) {
				p := cvoNewPre(t)
				row.set(p)
				res := p.run()
				cvoWantFail(t, res, row.token, row.name)
			})
		}
	})

	t.Run("a_window_or_slot_that_is_reclaimable_is_noted_not_failed", func(t *testing.T) {
		p := cvoNewPre(t)
		p.setIntegration(cvoWindowStale, 0)
		p.setSlot(cvoSlotExpired+cvoSlotGone, 0)
		cvoWantClear(t, p.run(), "reclaimable")
	})

	t.Run("a_dead_or_leader_session_is_not_an_active_lane", func(t *testing.T) {
		p := cvoNewPre(t)
		p.setSessions(cvoSessionsJSON(
			cvoSession(4242, "/x/moai/.moai/worktrees/t9007"), // lane cwd, but the pid is gone
			cvoSession(4243, "/x/moai"),                       // a leader on the primary checkout
		), 0)
		p.setAlive(false)
		cvoWantClear(t, p.run())

		p.setAlive(true) // the lane entry is now live; the leader entry never counts
		res := p.run()
		cvoWantFail(t, res, "active-lane-session", "t9007")
		if strings.Contains(rlsNorm(res.out), "FAIL active-lane-session: /x/moai ") {
			t.Errorf("a session on the primary checkout is not a lane:\n%s", rlsNorm(res.out))
		}
	})

	t.Run("the_excluded_card_own_session_is_exempted_and_printed", func(t *testing.T) {
		p := cvoNewPre(t)
		p.setSessions(cvoSessionsJSON(cvoSession(4242, "/x/moai/.moai/worktrees/t1453")), 0)
		cvoWantFail(t, p.run(), "active-lane-session", "t1453")
		cvoWantClear(t, p.run("--exclude-card", "t1453"), "EXCLUDED card t1453", "session")
	})

	t.Run("usage_errors_are_exit_2", func(t *testing.T) {
		p := cvoNewPre(t)
		if res := p.run("--no-such-flag"); res.exit != 2 {
			t.Errorf("unknown flag exit = %d, want 2", res.exit)
		}
		if res := p.run("--exclude-card"); res.exit != 2 {
			t.Errorf("--exclude-card without a value exit = %d, want 2", res.exit)
		}
	})

	t.Run("the_precheck_is_read_only", func(t *testing.T) {
		p := cvoNewPre(t)
		before := p.repo.git("for-each-ref") + "|" + p.repo.git("status", "--porcelain")
		// The script must actually run (a missing script would leave the repository
		// untouched too, which proves nothing).
		if res := p.run(); res.exit != 0 && res.exit != 1 {
			t.Fatalf("exit code = %d, want 0 or 1 (the script did not run)", res.exit)
		}
		after := p.repo.git("for-each-ref") + "|" + p.repo.git("status", "--porcelain")
		if before != after {
			t.Errorf("the precheck changed repository state:\nbefore=%s\nafter=%s", before, after)
		}
		entries, err := os.ReadDir(filepath.Join(p.repo.work, ".git"))
		if err != nil {
			t.Fatalf("read .git: %v", err)
		}
		for _, e := range entries {
			if e.Name() == "FETCH_HEAD" {
				t.Errorf("the precheck fetched (FETCH_HEAD exists); it must read origin/develop as it stands")
			}
		}
	})
}
