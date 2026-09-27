package cli

// SPEC-MCP-WORKTREE-UNTRACKED-001 — AC-MWU-014..016: on a config-orphaned
// worktree the primary's explicit audit gate still binds, the gate read fails
// closed only with positive worktree evidence, and catalogue/state answers
// carry a worktree warning. Non-parallel throughout (t.Setenv, t.Chdir).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// assertGateFail checks the codex_audit fail + gate_unmet shape of AC-MWU-014.
func assertGateFail(t *testing.T, label string, out ReviewOutput, wantReceipt bool) {
	t.Helper()
	if out.Verdict != "fail" || out.GateUnmet == "" {
		t.Fatalf("%s: want verdict fail with gate_unmet, got verdict=%q gate_unmet=%q summary=%q",
			label, out.Verdict, out.GateUnmet, out.Summary)
	}
	if wantReceipt && out.AuditReceipt == "" {
		t.Fatalf("%s: want a non-empty audit_receipt (the receipt-id read must see the primary's gate)", label)
	}
}

// multiAuditOn runs audit_multi's convergence on the root resolved for
// project_root = root, with Claude passing and codex returning no verdict.
func multiAuditOn(t *testing.T, root string) ConvergenceResult {
	t.Helper()
	resolved, err := resolveOptionalToolProjectRoot(newToolRequest(map[string]any{"project_root": root}))
	if err != nil {
		t.Fatalf("resolve %s: %v", root, err)
	}
	withBackendCall(t, gateBlockCaller(map[string]string{BackendCodex: VerdictInconclusive}))
	return runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		ProjectRoot: resolved,
	}, nil)
}

// AC-MWU-014: the primary's `codex: required` is enforced on W — before and
// after W/.moai/state/ exists — and on a hand-moved worktree and a
// relative-path worktree evaluated from a working directory where the relative
// gitdir does not resolve.
func TestConfigOrphanedWorktree_PrimaryGateEnforced(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowCodexRequired)

	for round, label := range []string{"first call", "after W/.moai/state exists"} {
		if round == 1 {
			// SPEC-WORKTREE-STATE-ROOT-001 AC-WSR-004 (i): the audits no longer
			// write under W, so the second round creates W/.moai/state itself.
			mustMkdir(t, filepath.Join(fx.W, ".moai", "state"))
		}
		out := codexAuditOn(t, fx.W)
		assertGateFail(t, "codex_audit "+label, out, true)
		r := multiAuditOn(t, fx.W)
		if r.OverallVerdict != overallVerdictFail || !strings.Contains(r.GateUnmet, BackendCodex) {
			t.Fatalf("audit_multi %s: want overall fail with gate_unmet naming codex, got overall=%q gate_unmet=%q",
				label, r.OverallVerdict, r.GateUnmet)
		}
		if round == 0 {
			if _, err := os.Stat(filepath.Join(fx.W, ".moai")); !os.IsNotExist(err) {
				t.Fatalf("premise: W/.moai must not exist after the first call (stat err=%v)", err)
			}
		}
	}

	t.Run("hand-moved worktree", func(t *testing.T) {
		base := filepath.Dir(fx.W)
		orig := filepath.Join(base, "Wmove0")
		wtGit(t, fx.env, fx.P, "worktree", "add", "-q", "-b", "wmove", orig)
		moved := filepath.Join(base, "Wmoved")
		if err := os.Rename(orig, moved); err != nil {
			t.Fatal(err)
		}
		mustMkdir(t, filepath.Join(moved, ".moai"))
		backRef, err := os.ReadFile(filepath.Join(fx.P, ".git", "worktrees", "Wmove0", "gitdir"))
		if err != nil || !strings.Contains(string(backRef), "Wmove0") {
			t.Fatalf("premise: the admin gitdir file must still name the old path; got %q err=%v", backRef, err)
		}
		assertGateFail(t, "hand-moved W′", codexAuditOn(t, moved), true)
	})

	t.Run("relative-path worktree from an unrelated cwd", func(t *testing.T) {
		wr := filepath.Join(filepath.Dir(fx.W), "WR")
		wtGit(t, fx.env, fx.P, "worktree", "add", "-q", "--relative-paths", "-b", "wrel", wr)
		mustMkdir(t, filepath.Join(wr, ".moai"))
		dotGit, err := os.ReadFile(filepath.Join(wr, ".git"))
		if err != nil {
			t.Fatal(err)
		}
		rel := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(dotGit)), "gitdir:"))
		if filepath.IsAbs(rel) {
			t.Fatalf("premise: WR's .git must carry a relative gitdir, got %q", rel)
		}
		cwd := filepath.Join(wtCanonTempDir(t), "a", "b")
		mustMkdir(t, cwd)
		t.Chdir(cwd)
		if _, err := os.Stat(filepath.Join(cwd, rel)); err == nil {
			t.Fatalf("premise: relative gitdir %q must NOT resolve from cwd %s", rel, cwd)
		}
		assertGateFail(t, "relative-path W_R", codexAuditOn(t, wr), true)
	})
}

// AC-MWU-015 (i): with positive worktree evidence and a primary that cannot be
// identified, the codex gate is assumed required. P declares NO codex gate, so
// a fail here can only come from the fail-closed path.
func TestConfigOrphanedWorktree_FailsClosedWhenPrimaryUnidentified(t *testing.T) {
	wantNote := func(t *testing.T, label string, out ReviewOutput) {
		t.Helper()
		assertGateFail(t, label, out, false)
		if !strings.Contains(out.GateUnmet, "assumed `required`") || !strings.Contains(out.GateUnmet, "could not be identified") {
			t.Fatalf("%s: gate_unmet must say the gate was assumed required because the primary could not be identified; got %q", label, out.GateUnmet)
		}
		t.Logf("%s: gate_unmet=%q", label, out.GateUnmet)
	}

	t.Run("separate-git-dir worktree (ambiguous layout)", func(t *testing.T) {
		_, w3 := newSeparateGitDirRepo(t)
		mustMkdir(t, filepath.Join(w3, ".moai"))
		wantNote(t, "W3", codexAuditOn(t, w3))
	})

	t.Run("git unavailable", func(t *testing.T) {
		fx := newUntrackedFixture(t, wtWorkflowNoGate)
		mustMkdir(t, filepath.Join(fx.W, ".moai"))
		noGitPath(t)
		wantNote(t, "W without git", codexAuditOn(t, fx.W))
	})

	t.Run("admin HEAD deleted (git exits non-zero)", func(t *testing.T) {
		fx := newUntrackedFixture(t, wtWorkflowNoGate)
		mustMkdir(t, filepath.Join(fx.W, ".moai"))
		admin := filepath.Join(fx.P, ".git", "worktrees", "W")
		if err := os.Remove(filepath.Join(admin, "HEAD")); err != nil {
			t.Fatal(err)
		}
		for _, keep := range []string{"commondir", "gitdir"} {
			if _, err := os.Stat(filepath.Join(admin, keep)); err != nil {
				t.Fatalf("premise: %s must stay intact: %v", keep, err)
			}
		}
		// Premise observation: the scrubbed inspection exits non-zero.
		if _, err := runScrubbedGit(fx.W, "rev-parse", "--path-format=absolute", "--git-common-dir"); err == nil {
			t.Fatal("premise: git rev-parse must fail once the admin HEAD is gone")
		} else {
			t.Logf("premise observed: git rev-parse after admin HEAD deletion: %v", err)
		}
		wantNote(t, "W with admin HEAD removed", codexAuditOn(t, fx.W))
	})
}

// AC-MWU-015 (ii): every root without worktree evidence keeps today's fail-open
// inconclusive and gets no gate_unmet.
func TestConfigOrphanedWorktree_OtherRootsKeepFailOpen(t *testing.T) {
	wantInconclusive := func(t *testing.T, label, root string) {
		t.Helper()
		out := codexAuditOn(t, root)
		if out.Verdict != VerdictInconclusive || out.GateUnmet != "" {
			t.Fatalf("%s: want fail-open inconclusive with no gate_unmet, got verdict=%q gate_unmet=%q", label, out.Verdict, out.GateUnmet)
		}
	}
	env := wtGitEnv(t)

	nonGit := wtCanonTempDir(t)
	mustMkdir(t, filepath.Join(nonGit, ".moai"))
	wantInconclusive(t, "non-git directory", nonGit)

	primary := filepath.Join(wtCanonTempDir(t), "R")
	mustMkdir(t, filepath.Join(primary, ".moai"))
	wtGit(t, env, primary, "init", "-q", "-b", "main")
	wantInconclusive(t, "repository primary", primary)

	s, _ := newSeparateGitDirRepo(t)
	wantInconclusive(t, "separate-git-dir primary", s)

	// Submodules: an ordinary one, and one checked out at a path whose parent
	// component is named "worktrees".
	libBase := wtCanonTempDir(t)
	lib := filepath.Join(libBase, "lib")
	mustMkdir(t, lib)
	wtGit(t, env, lib, "init", "-q", "-b", "main")
	wtWriteFile(t, filepath.Join(lib, "lib.txt"), "lib\n")
	wtGit(t, env, lib, "add", "lib.txt")
	wtGit(t, env, lib, "commit", "-q", "-m", "lib")
	super := filepath.Join(wtCanonTempDir(t), "super")
	mustMkdir(t, super)
	wtGit(t, env, super, "init", "-q", "-b", "main")
	wtWriteFile(t, filepath.Join(super, "s.txt"), "s\n")
	wtGit(t, env, super, "add", "s.txt")
	wtGit(t, env, super, "commit", "-q", "-m", "s")
	for _, sub := range []string{"sub", "vendor/worktrees/lib"} {
		wtGit(t, env, super, "-c", "protocol.file.allow=always", "submodule", "add", "-q", lib, sub)
		subRoot := filepath.Join(super, filepath.FromSlash(sub))
		modGitDir := filepath.Join(super, ".git", "modules", filepath.FromSlash(sub))
		if _, err := os.Stat(filepath.Join(modGitDir, "commondir")); !os.IsNotExist(err) {
			t.Fatalf("premise: submodule git dir %s must have no commondir (stat err=%v)", modGitDir, err)
		}
		t.Logf("premise observed: %s has no commondir", modGitDir)
		mustMkdir(t, filepath.Join(subRoot, ".moai"))
		wantInconclusive(t, "submodule "+sub, subRoot)
	}

	// Last, because it empties PATH for the rest of the test.
	noGitPath(t)
	wantInconclusive(t, "repository primary without git", primary)
}

// AC-MWU-016: catalogue/state answers read from W carry _root.worktree_warning;
// P's do not; spec_audit always carries _root; under a CLAUDE_PROJECT_DIR
// fallback naming W both warnings appear.
func TestConfigOrphanedWorktree_CatalogueWarning(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	key := "t0000000000000:wt-probe"
	calls := map[string]func(root string) string{
		"spec_progress": func(root string) string {
			return callSpecToolJSON(t, handleSpecProgress, map[string]any{"project_root": root})
		},
		"spec_audit": func(root string) string {
			return callSpecToolJSON(t, handleSpecAudit, map[string]any{"project_root": root})
		},
		"spec_drift": func(root string) string {
			return callSpecToolJSON(t, handleSpecDrift, map[string]any{"project_root": root})
		},
		"verify_snapshot": func(root string) string {
			return callSpecToolJSON(t, handleVerifySnapshot, map[string]any{"project_root": root, "key": key})
		},
		"verify_trend": func(root string) string {
			return callSpecToolJSON(t, handleVerifyTrend, map[string]any{"project_root": root, "key": key})
		},
	}
	for _, round := range []string{"first", "after W/.moai/state exists"} {
		if round != "first" {
			mustMkdir(t, filepath.Join(fx.W, ".moai", "state"))
		}
		for name, call := range calls {
			if body := call(fx.W); !strings.Contains(body, `worktree_warning`) {
				t.Errorf("%s on W (%s): missing _root.worktree_warning; body=%s", name, round, body)
			}
		}
	}
	for name, call := range calls {
		body := call(fx.P)
		if strings.Contains(body, "worktree_warning") {
			t.Errorf("%s on P: unexpected worktree_warning; body=%s", name, body)
		}
		if name == "spec_audit" && !strings.Contains(body, `_root`) {
			t.Errorf("spec_audit on P: missing _root; body=%s", body)
		}
	}

	t.Setenv("CLAUDE_PROJECT_DIR", fx.W)
	res, err := handleSpecProgress(context.Background(), newToolRequest(map[string]any{}))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Root map[string]any `json:"_root"`
	}
	blob, _ := json.Marshal(res.StructuredContent)
	if err := json.Unmarshal(blob, &decoded); err != nil {
		t.Fatalf("decode: %v (%s)", err, blob)
	}
	if decoded.Root["warning"] == nil || decoded.Root["worktree_warning"] == nil {
		t.Fatalf("fallback on W: want both warning and worktree_warning in _root, got %v", decoded.Root)
	}
	want := rootProvenanceMap(fx.W, "env:CLAUDE_PROJECT_DIR")["warning"]
	if decoded.Root["warning"] != want {
		t.Fatalf("existing warning text changed: got %v want %v", decoded.Root["warning"], want)
	}
}

// specAuditDirect renders spec.Audit's own result as a JSON object — the
// fields spec_audit returned before _root was attached.
func specAuditDirect(root string) (map[string]any, error) {
	r, err := spec.Audit(spec.AuditOptions{BaseDir: root})
	if err != nil {
		return nil, err
	}
	blob, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	return m, json.Unmarshal(blob, &m)
}

// spec_audit keeps its existing result fields with the same values once _root
// is attached (REQ-MWU-013).
func TestSpecAudit_RootBlockKeepsExistingFields(t *testing.T) {
	root := newProbeProject(t, "SPEC-WTAUDITKEEP-001")
	res, err := handleSpecAudit(context.Background(), newToolRequest(map[string]any{"project_root": root}))
	if err != nil {
		t.Fatal(err)
	}
	blob, _ := json.Marshal(res.StructuredContent)
	var got map[string]any
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatal(err)
	}
	if got["_root"] == nil {
		t.Fatalf("spec_audit lacks _root: %s", blob)
	}
	direct, err := specAuditDirect(root)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range direct {
		if k == "audited_at" {
			if _, ok := got[k]; !ok {
				t.Errorf("field %q dropped", k)
			}
			continue // a per-call timestamp: present, value naturally differs
		}
		gb, _ := json.Marshal(got[k])
		db, _ := json.Marshal(v)
		if string(gb) != string(db) {
			t.Errorf("field %q changed: got %s want %s", k, gb, db)
		}
	}
}
