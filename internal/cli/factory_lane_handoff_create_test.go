package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// TestLaneHandoffPreparationRefusals pins the admission refusals of
// prepareLaneHandoff that the named AC tests do not reach: an unknown lane
// activity, an invalid branch slug, a project with no active factory run
// (with the default materializer and output), a missing local develop pin,
// and a lane already sitting in the card's target with no earlier handoff.
func TestLaneHandoffPreparationRefusals(t *testing.T) {
	ctx := context.Background()

	t.Run("unknown_activity", func(t *testing.T) {
		_, err := prepareLaneHandoff(ctx, laneHandoffRequest{Activity: "compacting", Slug: handoffTestSlug}, laneHandoffDeps{})
		requireHandoffNack(t, err, factorymsg.NackInvalidRequest)
	})

	t.Run("invalid_slug", func(t *testing.T) {
		for _, slug := range []string{"Bad_Slug", "one-two-three-four", strings.Repeat("a", 25)} {
			_, err := prepareLaneHandoff(ctx, laneHandoffRequest{Activity: laneActivityIdle, Slug: slug}, laneHandoffDeps{})
			requireHandoffNack(t, err, factorymsg.NackInvalidRequest)
		}
	})

	t.Run("no_active_run_with_default_deps", func(t *testing.T) {
		t.Setenv("MOAI_HOME", t.TempDir())
		_, err := prepareLaneHandoff(ctx, laneHandoffRequest{
			ProjectRoot: t.TempDir(), Slot: handoffTestSlot, CardID: handoffTestCard, SpecID: handoffTestSpec,
			Slug: handoffTestSlug, Mode: factorymsg.HandoffModeInteractive, Activity: laneActivityIdle,
		}, laneHandoffDeps{})
		if err == nil {
			t.Fatal("preparation without an active factory run succeeded")
		}
		if _, isNack := factorymsg.HandoffNackReason(err); isNack {
			t.Fatalf("err=%v, want the run-resolution error, not a handoff NACK", err)
		}
	})

	t.Run("missing_develop_pin", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		handoffGit(t, f.primary, "branch", "-D", "develop")
		_, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), f.deps())
		if err == nil || !strings.Contains(err.Error(), "read local develop pin") {
			t.Fatalf("err=%v, want the develop-pin read error", err)
		}
		if n := f.count(t, `SELECT count(*) FROM lane_handoffs`); n != 0 {
			t.Fatalf("handoff rows = %d, want 0", n)
		}
	})

	t.Run("cwd_is_target_without_earlier_handoff", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		target := f.target(handoffTestCard)
		if err := os.MkdirAll(target, 0o700); err != nil {
			t.Fatal(err)
		}
		req := f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive)
		req.SourceCwd = target
		_, err := prepareLaneHandoff(ctx, req, f.deps())
		requireHandoffNack(t, err, factorymsg.NackTargetPathConflict)
		if f.materializeCalls != 0 {
			t.Fatalf("materializer calls = %d, want 0", f.materializeCalls)
		}
	})
}

// TestLaneHandoffCreateTargetProvenance pins every provenance refusal of
// createHandoffTarget with a materializer that produces exactly one defect.
// Each refusal NACKs the reserved handoff with its reason and never removes
// what the materializer made.
func TestLaneHandoffCreateTargetProvenance(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		make func(t *testing.T, f *laneHandoffFixture, target string) (string, error)
		want string
	}{
		{"materializer_fails", func(*testing.T, *laneHandoffFixture, string) (string, error) {
			return "", errors.New("materializer refused")
		}, factorymsg.NackTargetPathConflict},
		{"materializer_returns_another_path", func(t *testing.T, _ *laneHandoffFixture, _ string) (string, error) {
			return t.TempDir(), nil
		}, factorymsg.NackTargetPathConflict},
		{"card_branch_missing", func(t *testing.T, _ *laneHandoffFixture, target string) (string, error) {
			return target, os.MkdirAll(target, 0o700)
		}, factorymsg.NackBranchCollision},
		{"develop_moved_after_creation", func(t *testing.T, f *laneHandoffFixture, target string) (string, error) {
			handoffGit(t, f.primary, "worktree", "add", "-q", "-b", handoffTestCard, target, "develop")
			moved := handoffGit(t, f.primary, "commit-tree", "-p", "develop", "-m", "moved", "develop^{tree}")
			handoffGit(t, f.primary, "update-ref", "refs/heads/develop", moved)
			return target, nil
		}, factorymsg.NackBaseDrift},
		{"detached_head", func(t *testing.T, f *laneHandoffFixture, target string) (string, error) {
			handoffGit(t, f.primary, "worktree", "add", "-q", "--detach", target, "develop")
			handoffGit(t, f.primary, "branch", handoffTestCard, "develop")
			return target, nil
		}, factorymsg.NackBranchCollision},
		{"branch_checked_out_twice", func(t *testing.T, f *laneHandoffFixture, target string) (string, error) {
			handoffGit(t, f.primary, "worktree", "add", "-q", "-b", handoffTestCard, target, "develop")
			handoffGit(t, f.primary, "worktree", "add", "-q", "-f", target+"-dup", handoffTestCard)
			return target, nil
		}, factorymsg.NackBranchCollision},
		{"target_dirty", func(t *testing.T, f *laneHandoffFixture, target string) (string, error) {
			handoffGit(t, f.primary, "worktree", "add", "-q", "-b", handoffTestCard, target, "develop")
			return target, os.WriteFile(filepath.Join(target, "stray.txt"), []byte("stray\n"), 0o600)
		}, factorymsg.NackTargetDirty},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			target := f.target(handoffTestCard)
			var out bytes.Buffer
			deps := laneHandoffDeps{
				Materialize: func(string, io.Writer) (string, error) {
					f.materializeCalls++
					return tc.make(t, f, target)
				},
				AppServer: f.appServer, Out: &out,
			}
			_, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), deps)
			requireHandoffNack(t, err, tc.want)
			if f.materializeCalls != 1 {
				t.Fatalf("materializer calls = %d, want 1", f.materializeCalls)
			}
			hs, herr := f.store.HandoffsForLane(ctx, handoffTestSlot)
			if herr != nil || len(hs) != 1 || hs[0].State != factorymsg.HandoffNack || hs[0].Reason != tc.want {
				t.Fatalf("handoffs = %+v err=%v, want one NACK/%s", hs, herr, tc.want)
			}
			if tc.name == "materializer_fails" && !strings.Contains(out.String(), "handoff worktree creation failed") {
				t.Fatalf("creation failure not reported: %q", out.String())
			}
			if f.appServer.calls != 0 {
				t.Fatalf("app-server requests = %d, want 0", f.appServer.calls)
			}
		})
	}
}
