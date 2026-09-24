package cli

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// noWriteCounters are the AC-FLH-011 counters. Every one must stay 0 until
// BOUND; messages are counted end to end.
type noWriteCounters struct {
	CodeWrites, WrongCwdWrites, PrimarySwitches, PreBoundCommits, TaskACKs int
	Sent, Received, Lost                                                   int
	Observations                                                           []string
}

// gitArgvRecorder instruments the handoff Git adapter: every subprocess the
// controller starts goes through handoffCommand, and its argv is recorded.
type gitArgvRecorder struct {
	mu   sync.Mutex
	argv [][]string
}

func (r *gitArgvRecorder) install(t *testing.T) {
	t.Helper()
	prev := handoffCommand
	handoffCommand = func(name string, args ...string) *exec.Cmd {
		r.mu.Lock()
		r.argv = append(r.argv, append([]string{name}, args...))
		r.mu.Unlock()
		return prev(name, args...)
	}
	t.Cleanup(func() { handoffCommand = prev })
}

// primarySwitchesAndCommits counts recorded Git calls that would switch the
// primary checkout's branch or create a commit anywhere.
func (r *gitArgvRecorder) primarySwitchesAndCommits(primary string) (switches, commits int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.argv {
		dir, verb, rest := "", "", []string(nil)
		for i := 1; i < len(a); i++ {
			if a[i] == "-C" && i+1 < len(a) {
				dir = a[i+1]
				i++
				continue
			}
			verb, rest = a[i], a[i+1:]
			break
		}
		if verb == "commit" || verb == "merge" || verb == "cherry-pick" || verb == "rebase" {
			commits++
		}
		if dir != primary {
			continue
		}
		switch verb {
		case "checkout", "switch", "reset", "stash":
			switches++
		case "branch":
			for _, x := range rest {
				if x == "-m" || x == "-M" || x == "-c" || x == "-d" || x == "-D" || x == "-f" {
					switches++
					break
				}
			}
		}
	}
	return switches, commits
}

// primaryTreeListing is the primary checkout's worktree files, excluding Git
// metadata, the L1 worktree root, and the ignored MoAI state directory.
func primaryTreeListing(t *testing.T, primary string) string {
	t.Helper()
	listing := dirListing(t, primary)
	var keep []string
	for _, e := range strings.Split(listing, ";") {
		rel := strings.TrimPrefix(e, primary+string(filepath.Separator))
		if strings.HasPrefix(rel, ".claude"+string(filepath.Separator)+"worktrees") || strings.HasPrefix(rel, ".moai"+string(filepath.Separator)) {
			continue
		}
		keep = append(keep, rel)
	}
	return strings.Join(keep, ";")
}

// TestFactoryLaneHandoffNoPreBoundWrites is AC-FLH-011 (REQ-FLH-008/009/013):
// through every state up to BOUND the handoff makes no code write, no
// wrong-cwd write, no primary branch switch, no commit, and no task ACK; after
// BOUND and one dispatch round every message sent is received, none lost.
func TestFactoryLaneHandoffNoPreBoundWrites(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
	rec := &gitArgvRecorder{}
	rec.install(t)
	lead := abandonTestLead(t, f)
	ctx := context.Background()
	target := f.target(handoffTestCard)

	primaryBranch := handoffGit(t, f.primary, "symbolic-ref", "--short", "HEAD")
	primaryHead := handoffGit(t, f.primary, "rev-parse", "HEAD")
	primaryStatus := handoffGit(t, f.primary, "status", "--porcelain")
	primaryFiles := primaryTreeListing(t, f.primary)

	var c noWriteCounters
	sent := map[string]bool{}
	send := func(key string) {
		env, err := f.store.Send(ctx, factorymsg.SendRequest{From: lead, To: f.source, Kind: factorymsg.KindDispatchNotice, IdempotencyKey: key, TaskRef: "t1082", CorrelationID: "c-" + key, TTL: time.Hour, Payload: []byte("body-" + key)})
		if err != nil {
			t.Fatalf("send %s: %v", key, err)
		}
		sent[env.ID] = true
		c.Sent++
	}
	// observe measures every counter in the current state. It runs at each
	// controller crash point (used here as an observation hook) and between
	// controller steps.
	observe := func(state string) {
		c.Observations = append(c.Observations, state)
		if b := handoffGit(t, f.primary, "symbolic-ref", "--short", "HEAD"); b != primaryBranch {
			c.PrimarySwitches++
		}
		if h := handoffGit(t, f.primary, "rev-parse", "HEAD"); h != primaryHead {
			c.PrimarySwitches++
		}
		if handoffGit(t, f.primary, "status", "--porcelain") != primaryStatus || primaryTreeListing(t, f.primary) != primaryFiles {
			c.WrongCwdWrites++
		}
		if pathExists(target) {
			if out := handoffGit(t, target, "status", "--porcelain"); out != "" {
				c.CodeWrites++
			}
			if n := handoffGit(t, target, "rev-list", "--count", f.developPin+"..HEAD"); n != "0" {
				c.PreBoundCommits++
			}
		}
		// The broker never authorizes a code write or a commit before BOUND.
		if err := f.store.AuthorizeCardWrite(ctx, f.source, handoffTestCard); err == nil {
			c.CodeWrites++
		}
		// A task ACK needs a claim; before BOUND neither is granted.
		if claims, err := f.store.Claim(ctx, f.source, factorymsg.MaxBatch, time.Minute); err == nil {
			for _, cl := range claims {
				if f.store.Receipt(ctx, f.source, cl.ID, cl.ClaimToken) == nil {
					c.TaskACKs++
				}
			}
		}
	}

	send("d1-before-reservation")
	prev := laneHandoffFailpoint
	laneHandoffFailpoint = func(point string) {
		if point != handoffPointBound {
			observe("point:" + point)
		}
	}
	t.Cleanup(func() { laneHandoffFailpoint = prev })

	h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeHeadless), f.deps())
	if err != nil || h.State != factorymsg.HandoffWTReady {
		t.Fatalf("prepare = %+v err=%v", h, err)
	}
	observe("WT_READY")
	h, err = switchLaneHandoffHeadless(ctx, h, f.switchRequest(laneActivityIdle, "thr-source"), laneHandoffDeps{})
	if err != nil || h.State != factorymsg.HandoffSwitchPendingHeadless {
		t.Fatalf("switch = %+v err=%v", h, err)
	}
	send("d2-during-switch-pending")
	observe("SWITCH_PENDING_HEADLESS")
	switches, commits := rec.primarySwitchesAndCommits(f.primary)
	c.PrimarySwitches += switches
	c.PreBoundCommits += commits

	bind := headlessOwner(t)
	bind.ProjectRoot = f.primary
	b, err := bindLaneHandoffHeadless(ctx, h, bind)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}

	// One dispatch round on the bound endpoint.
	bound := factorymsg.Peer{ProjectKey: "project", RunID: f.run, Backend: f.source.Backend, Role: f.source.Role, Slot: handoffTestSlot,
		SessionUUID: b.New.SessionUUID, Generation: b.New.Generation, PID: bind.OwnerPID, ProcessStart: bind.OwnerProcessStart}
	if err := f.store.AuthorizeCardWrite(ctx, bound, handoffTestCard); err != nil {
		t.Fatalf("bound endpoint not authorized after BOUND (positive control): %v", err)
	}
	claims, err := f.store.Claim(ctx, bound, factorymsg.MaxBatch, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	for _, cl := range claims {
		if !sent[cl.ID] {
			t.Fatalf("claimed %s that was never sent", cl.ID)
		}
		if _, err := f.store.ReadBody(ctx, bound, cl.ID, cl.ClaimToken); err != nil {
			t.Fatal(err)
		}
		if err := f.store.RecordDisposition(ctx, bound, cl.ID, cl.ClaimToken, factorymsg.DispositionAccepted); err != nil {
			t.Fatal(err)
		}
		if err := f.store.Receipt(ctx, bound, cl.ID, cl.ClaimToken); err != nil {
			t.Fatal(err)
		}
		c.Received++
	}
	c.Lost = c.Sent - f.count(t, `SELECT count(*) FROM messages WHERE state='acknowledged'`)

	wantObs := []string{"point:" + handoffPointReserved, "point:" + handoffPointCreated, "point:" + handoffPointRenamed, "WT_READY", "point:" + handoffPointSwitchPending, "point:" + handoffPointRelocated, "SWITCH_PENDING_HEADLESS"}
	if got := strings.Join(c.Observations, ","); got != strings.Join(wantObs, ",") {
		t.Fatalf("states observed = %s, want %s", got, strings.Join(wantObs, ","))
	}
	t.Logf("AC_FLH_011_COUNTERS code_writes=%d wrong_cwd_writes=%d primary_switches=%d pre_bound_commits=%d task_acks=%d messages=%d/%d/%d git_calls=%d",
		c.CodeWrites, c.WrongCwdWrites, c.PrimarySwitches, c.PreBoundCommits, c.TaskACKs, c.Sent, c.Received, c.Lost, len(rec.argv))
	if c.CodeWrites != 0 || c.WrongCwdWrites != 0 || c.PrimarySwitches != 0 || c.PreBoundCommits != 0 || c.TaskACKs != 0 {
		t.Fatalf("pre-BOUND counters = %+v, want all 0", c)
	}
	if c.Sent != 2 || c.Received != c.Sent || c.Lost != 0 {
		t.Fatalf("messages sent/received/lost = %d/%d/%d, want 2/2/0", c.Sent, c.Received, c.Lost)
	}
	if len(rec.argv) == 0 {
		t.Fatal("Git adapter recorded no calls: the instrumentation saw nothing")
	}
}
