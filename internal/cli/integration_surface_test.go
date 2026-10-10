package cli

// integration_surface_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1
// RED tests for the integration surface (REQ-LMF-002; plan §B2). The surface is
// resolved by integrationMergeWorktree, the one resolver the merge verb and
// factory complete share, so the cases are asserted on it directly. The
// case-5 guidance sentence is the plan §B2 wording.

import (
	"path/filepath"
	"strings"
	"testing"
)

// lmfCaseFiveGuidance is the plan §B2 case-5 sentence that a refusal with no
// tree holding the integration branch must carry.
const lmfCaseFiveGuidance = "The tool does not switch branches"

func TestIntegrationSurfaceSelectsPrimaryWhenEnabled(t *testing.T) {
	root, _ := lmfRepo(t, true, "")
	resolved, err := integrationMergeWorktree(root, lmfBranch)
	if err != nil {
		t.Fatalf("with the gate on and the primary holding %s, the surface is the primary's local main: %v", lmfBranch, err)
	}
	if !factorySameTree(resolved, root) {
		t.Fatalf("the surface must be the primary checkout %s, got %s", root, resolved)
	}
}

func TestIntegrationSurfaceRefusesPrimaryWhenDisabled(t *testing.T) {
	root, _ := lmfRepo(t, false, "")
	resolved, err := integrationMergeWorktree(root, lmfBranch)
	if err == nil {
		t.Fatalf("with the gate off, the primary holding %s must stay refused, got %q", lmfBranch, resolved)
	}
	if !strings.Contains(err.Error(), "primary checkout") {
		t.Fatalf("the refusal must name the primary checkout: %v", err)
	}
}

func TestIntegrationSurfaceRefusesPrimaryOffBranch(t *testing.T) {
	// The primary checks out another branch, and no tree holds main.
	root, _ := lmfRepo(t, true, "")
	fcGit(t, root, "branch", "feature")
	fcGit(t, root, "symbolic-ref", "HEAD", "refs/heads/feature")
	resolved, err := integrationMergeWorktree(root, lmfBranch)
	if err == nil {
		t.Fatalf("the primary on another branch, with no tree holding %s, must refuse; got %q", lmfBranch, resolved)
	}
	if !strings.Contains(err.Error(), lmfCaseFiveGuidance) {
		t.Fatalf("the refusal must carry the case-5 guidance (plan §B2): %v", err)
	}
	if head := fcGit(t, root, "rev-parse", "--symbolic-full-name", "HEAD"); head != "refs/heads/feature" {
		t.Fatalf("the refusal must switch no branch: HEAD now names %q", head)
	}
}

func TestIntegrationSurfaceRefusesNoHolder(t *testing.T) {
	// The primary is detached, so its HEAD names no branch and no tree holds main.
	root, _ := lmfRepo(t, true, "")
	fcGit(t, root, "checkout", "-q", "--detach")
	resolved, err := integrationMergeWorktree(root, lmfBranch)
	if err == nil {
		t.Fatalf("a detached primary with no tree holding %s must refuse; got %q", lmfBranch, resolved)
	}
	if !strings.Contains(err.Error(), lmfCaseFiveGuidance) {
		t.Fatalf("the refusal must carry the case-5 guidance (plan §B2): %v", err)
	}
	if head := fcGit(t, root, "rev-parse", "--symbolic-full-name", "HEAD"); head != "HEAD" {
		t.Fatalf("the refusal must leave HEAD detached: it now names %q", head)
	}
}

func TestIntegrationSurfaceSeparateWorktreeUnchanged(t *testing.T) {
	// A separate integration worktree holds main. Its resolution must not depend on the gate.
	for _, gateOn := range []bool{false, true} {
		gateOn := gateOn
		name := "gate-off"
		if gateOn {
			name = "gate-on"
		}
		t.Run(name, func(t *testing.T) {
			root, _ := lmfRepo(t, gateOn, "")
			fcGit(t, root, "branch", "feature")
			fcGit(t, root, "symbolic-ref", "HEAD", "refs/heads/feature")
			integ := filepath.Join(t.TempDir(), "integ")
			fcGit(t, root, "worktree", "add", "-q", integ, lmfBranch)
			resolved, err := integrationMergeWorktree(root, lmfBranch)
			if err != nil {
				t.Fatalf("a separate integration worktree holding %s must resolve whatever the gate: %v", lmfBranch, err)
			}
			if !factorySameTree(resolved, integ) {
				t.Fatalf("the surface must be the separate worktree %s, got %s", integ, resolved)
			}
		})
	}
}
