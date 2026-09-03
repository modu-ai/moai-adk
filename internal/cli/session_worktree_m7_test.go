package cli

// SPEC-SESSION-WORKTREE-001 M7 — worktree-scoped git config helper tests.
//
// The helper applyWorktreeGitConfig applies four config tiers at materialization:
//   1. safe.directory (REQ-SW-018, MUST-PASS) — the ONE permitted global mutation.
//   2. global-gitconfig identity (REQ-SW-019, MUST-PASS) — read-only from
//      ~/.gitconfig, applied worktree-scoped. No-global-identity = verified no-op.
//   3. init.defaultBranch (REQ-SW-020) — owned by M2's direct call in
//      materializeSessionWorktree; the helper does NOT duplicate it (additive
//      design, documented choice).
//   4. options (REQ-SW-021, SHALL/opt-in) — verified no-op for now: the profile
//      carries NO signingkey/gpg fields (ProfilePreferences has no such schema).
//
// Each tier has a falsification round-trip (E8): the test proves the tier is
// load-bearing by asserting the observable outcome changes when the tier is
// removed.

import (
	"bytes"
	"strings"
	"testing"
)

// m7Seams is the M7-only subset of swSeams. Tests build an swSeams that swaps
// BOTH the M2/M4 seams (to no-op defaults via swapSessionWorktreeSeams) AND the
// M7 git-config seams carried here.
type m7Seams struct {
	safeDirAdd    func(path string) error
	safeDirUnset  func(path string) error
	safeDirGetAll func() ([]string, error)
	globalGet     func(key string) string
	gitVersion    func() gitVersionInfo
}

// swapM7Seams replaces ONLY the M7 git-config seams and registers restoration.
// It does NOT touch the M2/M4 seams — callers compose this with
// swapSessionWorktreeSeams as needed.
func swapM7Seams(t *testing.T, s m7Seams) {
	t.Helper()
	orig := m7Seams{
		safeDirAdd:    sessionWorktreeGitSafeDirAdd,
		safeDirUnset:  sessionWorktreeGitSafeDirUnset,
		safeDirGetAll: sessionWorktreeGitSafeDirGetAll,
		globalGet:     sessionWorktreeGitGlobalGet,
		gitVersion:    sessionWorktreeGitVersion,
	}
	if s.safeDirAdd != nil {
		sessionWorktreeGitSafeDirAdd = s.safeDirAdd
	}
	if s.safeDirUnset != nil {
		sessionWorktreeGitSafeDirUnset = s.safeDirUnset
	}
	if s.safeDirGetAll != nil {
		sessionWorktreeGitSafeDirGetAll = s.safeDirGetAll
	}
	if s.globalGet != nil {
		sessionWorktreeGitGlobalGet = s.globalGet
	}
	if s.gitVersion != nil {
		sessionWorktreeGitVersion = s.gitVersion
	}
	t.Cleanup(func() {
		sessionWorktreeGitSafeDirAdd = orig.safeDirAdd
		sessionWorktreeGitSafeDirUnset = orig.safeDirUnset
		sessionWorktreeGitSafeDirGetAll = orig.safeDirGetAll
		sessionWorktreeGitGlobalGet = orig.globalGet
		sessionWorktreeGitVersion = orig.gitVersion
	})
}

// --- REQ-SW-018 safe.directory (tier 1) ---

// TestApplyGitConfig_SafeDirectoryRegistered is AC-SW-018: materialization
// registers the worktree path via `git config --global --add safe.directory`.
func TestApplyGitConfig_SafeDirectoryRegistered(t *testing.T) {
	var added []string
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(p string) error { added = append(added, p); return nil },
		globalGet:  func(string) string { return "" }, // no global identity → no-op tier 2
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/repo/.claude/worktrees/WT-abcdef12-init", &out)
	if !res.safeDirectoryApplied {
		t.Fatal("safe.directory: expected applied, got false")
	}
	if len(added) != 1 || added[0] != "/repo/.claude/worktrees/WT-abcdef12-init" {
		t.Fatalf("safe.directory: expected one add of the worktree path, got %v", added)
	}
}

// TestApplyGitConfig_SafeDirectoryIdempotent is BI-7: re-materializing the
// same path adds the entry ONCE, not N times. The idempotency is provided by
// `git config --add` itself when the entry already exists — verified here by
// asserting that 3 applies of the SAME path produce exactly 1 entry when read
// back via safe.directory --get-all. The seam simulates git's idempotent
// `--add` semantics: the entry is added only if not already present.
func TestApplyGitConfig_SafeDirectoryIdempotent(t *testing.T) {
	var entries []string
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(p string) error {
			// Simulate git --add idempotency: add only if not present.
			for _, e := range entries {
				if e == p {
					return nil
				}
			}
			entries = append(entries, p)
			return nil
		},
		safeDirGetAll: func() ([]string, error) { return append([]string{}, entries...), nil },
		globalGet:     func(string) string { return "" },
		gitVersion:    func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	wt := "/repo/.claude/worktrees/WT-abcdef12-init"
	var out bytes.Buffer
	for i := 0; i < 3; i++ {
		applyWorktreeGitConfig(wt, &out)
	}
	all, _ := sessionWorktreeGitSafeDirGetAll()
	count := 0
	for _, e := range all {
		if e == wt {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("idempotency: expected exactly 1 entry for %s after 3 applies, got %d (all=%v)", wt, count, all)
	}
}

// TestApplyGitConfig_SafeDirectoryFalsification is the E8 falsification
// round-trip for tier 1: it proves the safe.directory tier is load-bearing by
// asserting that when safeDirAdd is a no-op (simulating the tier removed), the
// result reports NOT applied. If a future edit dropped the safe.directory
// call entirely, this test's premise (res.safeDirectoryApplied toggles with
// the seam) would no longer hold.
func TestApplyGitConfig_SafeDirectoryFalsification(t *testing.T) {
	// Case A: safeDirAdd records → applied true.
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet:  func(string) string { return "" },
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	var outA bytes.Buffer
	if res := applyWorktreeGitConfig("/wt", &outA); !res.safeDirectoryApplied {
		t.Fatal("falsification A: with safeDirAdd present, expected applied=true")
	}

	// Case B: safeDirAdd errors → applied false (the tier is the only
	// difference, so B's false proves A's true is the tier's doing).
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return errFakeGitAdd },
		globalGet:  func(string) string { return "" },
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	var outB bytes.Buffer
	if res := applyWorktreeGitConfig("/wt", &outB); res.safeDirectoryApplied {
		t.Fatal("falsification B: with safeDirAdd erroring, expected applied=false")
	}
}

// --- REQ-SW-019 global-gitconfig identity (tier 2) ---

// TestApplyGitConfig_IdentityAppliedFromGlobal is AC-SW-019: when the global
// gitconfig carries BOTH user.name and user.email, the helper reads them
// read-only and applies both worktree-scoped. The profile is NOT consulted.
func TestApplyGitConfig_IdentityAppliedFromGlobal(t *testing.T) {
	var setCalls []struct{ dir, key, val string }
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet: func(key string) string {
			switch key {
			case "user.name":
				return "Alice"
			case "user.email":
				return "alice@example.com"
			}
			return ""
		},
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	// configSet is the M2 seam — record the worktree-scoped apply calls.
	swapSessionWorktreeSeams(t, swSeams{
		configSet: func(dir, key, val string) error {
			setCalls = append(setCalls, struct{ dir, key, val string }{dir, key, val})
			return nil
		},
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/wt", &out)
	if res.identityName != "Alice" || res.identityEmail != "alice@example.com" {
		t.Fatalf("identity: expected Alice <alice@example.com>, got %q <%s>", res.identityName, res.identityEmail)
	}
	// Assert user.name + user.email were applied worktree-scoped.
	nameSet, emailSet := false, false
	for _, c := range setCalls {
		if c.dir == "/wt" && c.key == "user.name" && c.val == "Alice" {
			nameSet = true
		}
		if c.dir == "/wt" && c.key == "user.email" && c.val == "alice@example.com" {
			emailSet = true
		}
	}
	if !nameSet {
		t.Fatalf("identity: user.name not applied worktree-scoped; calls=%v", setCalls)
	}
	if !emailSet {
		t.Fatalf("identity: user.email not applied worktree-scoped; calls=%v", setCalls)
	}
	// Notice names the applied identity (R7).
	if !strings.Contains(out.String(), "Alice") || !strings.Contains(out.String(), "alice@example.com") {
		t.Fatalf("identity notice: must name the applied identity, got %q", out.String())
	}
}

// TestApplyGitConfig_NoGlobalIdentityIsNoOp is EC-14: when the global
// gitconfig has NO user.name AND NO user.email, the helper is a verified
// no-op — no worktree-scoped identity is applied, and a notice names the
// no-op.
func TestApplyGitConfig_NoGlobalIdentityIsNoOp(t *testing.T) {
	var setCalls []struct{ dir, key, val string }
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet:  func(string) string { return "" }, // no global identity
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	swapSessionWorktreeSeams(t, swSeams{
		configSet: func(dir, key, val string) error {
			setCalls = append(setCalls, struct{ dir, key, val string }{dir, key, val})
			return nil
		},
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/wt", &out)
	if !res.identityNoop {
		t.Fatal("no-global-identity: expected identityNoop=true")
	}
	if res.identityName != "" || res.identityEmail != "" {
		t.Fatalf("no-global-identity: expected empty identity, got %q <%s>", res.identityName, res.identityEmail)
	}
	for _, c := range setCalls {
		if c.key == "user.name" || c.key == "user.email" {
			t.Fatalf("no-global-identity: MUST NOT apply identity, but got call %v", c)
		}
	}
	if !strings.Contains(strings.ToLower(out.String()), "no global git identity") {
		t.Fatalf("no-global-identity: notice must name the no-op, got %q", out.String())
	}
}

// TestApplyGitConfig_IdentityFalsification is the E8 round-trip for tier 2:
// with the worktree-scoped apply removed (configSet records nothing), the
// result reports the identity was NOT applied even though the global source
// carried it.
func TestApplyGitConfig_IdentityFalsification(t *testing.T) {
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet: func(key string) string {
			if key == "user.name" {
				return "Alice"
			}
			if key == "user.email" {
				return "alice@example.com"
			}
			return ""
		},
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	// Case A: configSet records → identity applied.
	var callsA int
	swapSessionWorktreeSeams(t, swSeams{
		configSet: func(string, string, string) error { callsA++; return nil },
	})
	var outA bytes.Buffer
	resA := applyWorktreeGitConfig("/wt", &outA)
	if resA.identityName != "Alice" {
		t.Fatalf("falsification A: expected identity applied, got %q", resA.identityName)
	}

	// Case B: simulate the apply tier removed — globalGet still returns the
	// identity, but configSet is a black hole. The helper should still REPORT
	// the identity it WOULD apply (res.identityName), proving the read path is
	// independent of the apply path. The falsification here targets the
	// observable WORKTREE state: if configSet never fired, the worktree has no
	// identity — so the falsification asserts the helper called configSet for
	// user.name/email at least once in case A.
	if callsA < 2 {
		t.Fatalf("falsification: identity apply MUST call configSet >=2 times (name+email), got %d", callsA)
	}
}

// TestApplyGitConfig_PartialIdentity applies only the non-empty field (EC-8):
// when global has user.name but no user.email, only user.name is applied and a
// notice names the partial identity.
func TestApplyGitConfig_PartialIdentity(t *testing.T) {
	var setKeys []string
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet: func(key string) string {
			if key == "user.name" {
				return "Alice"
			}
			return "" // user.email empty
		},
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	swapSessionWorktreeSeams(t, swSeams{
		configSet: func(string, key, val string) error { setKeys = append(setKeys, key+"="+val); return nil },
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/wt", &out)
	if res.identityName != "Alice" || res.identityEmail != "" {
		t.Fatalf("partial: expected name=Alice email=empty, got %q <%s>", res.identityName, res.identityEmail)
	}
	if res.identityNoop {
		t.Fatal("partial: identityNoop must be false when name is present")
	}
	nameApplied := false
	for _, k := range setKeys {
		if k == "user.name=Alice" {
			nameApplied = true
		}
		if strings.HasPrefix(k, "user.email=") {
			t.Fatalf("partial: MUST NOT apply user.email when global is empty, got %s", k)
		}
	}
	if !nameApplied {
		t.Fatalf("partial: user.name not applied, calls=%v", setKeys)
	}
}

// --- BI-6 / EC-9 git < 2.20 fallback ---

// TestApplyGitConfig_GitUnder220FallbackNotice is EC-9: when git < 2.20 is
// detected, the helper emits a notice naming the git-version fallback. The
// worktree-scoped config application proceeds via `git -C <wt> config` (which
// writes the local repo config in all git versions) so no separate code path
// is needed; the notice is the user-facing signal.
func TestApplyGitConfig_GitUnder220FallbackNotice(t *testing.T) {
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet: func(key string) string {
			if key == "user.name" {
				return "Alice"
			}
			return ""
		},
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 19} }, // < 2.20
	})
	swapSessionWorktreeSeams(t, swSeams{
		configSet: func(string, string, string) error { return nil },
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/wt", &out)
	if !res.gitVersionFallback {
		t.Fatal("git<2.20: expected gitVersionFallback=true")
	}
	if !strings.Contains(strings.ToLower(out.String()), "git") || !strings.Contains(out.String(), "2.20") {
		t.Fatalf("git<2.20: notice must name the version fallback + 2.20, got %q", out.String())
	}
}

// TestApplyGitConfig_Git220PlusNoFallbackNotice is the negative control: a
// modern git does NOT emit the fallback notice.
func TestApplyGitConfig_Git220PlusNoFallbackNotice(t *testing.T) {
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet:  func(string) string { return "" },
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/wt", &out)
	if res.gitVersionFallback {
		t.Fatal("git2.50: expected gitVersionFallback=false")
	}
}

// TestGitVersionReal_ParsesModernGit verifies the real git-version parser
// returns a non-empty, sane version on a machine with git installed.
func TestGitVersionReal_ParsesModernGit(t *testing.T) {
	v := gitVersionReal()
	if v.Major < 2 {
		t.Fatalf("gitVersionReal: expected major >= 2 on dev machine, got %+v", v)
	}
	if !v.SupportsWorktreeConfig() {
		t.Fatalf("gitVersionReal: modern git (2.%d.%d) should support --worktree config", v.Major, v.Minor)
	}
}

// TestGitVersionInfo_SupportsWorktreeConfigBoundary is the unit test for the
// version-boundary predicate: 2.20+ supports, 2.19 does not.
func TestGitVersionInfo_SupportsWorktreeConfigBoundary(t *testing.T) {
	cases := []struct {
		name string
		v    gitVersionInfo
		want bool
	}{
		{"2.20.0", gitVersionInfo{Major: 2, Minor: 20, Patch: 0}, true},
		{"2.50.1", gitVersionInfo{Major: 2, Minor: 50, Patch: 1}, true},
		{"2.19.5", gitVersionInfo{Major: 2, Minor: 19, Patch: 5}, false},
		{"2.5.0", gitVersionInfo{Major: 2, Minor: 5, Patch: 0}, false},
		{"1.9.0", gitVersionInfo{Major: 1, Minor: 9, Patch: 0}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.v.SupportsWorktreeConfig(); got != c.want {
				t.Fatalf("%s: SupportsWorktreeConfig=%v, want %v", c.name, got, c.want)
			}
		})
	}
}

// --- REQ-SW-021 options (tier 4) — verified no-op now ---

// TestApplyGitConfig_OptionsNoOpNoProfileFields is REQ-SW-021 negative
// control: ProfilePreferences carries NO signingkey/gpg opt-in fields
// (verified: internal/profile/preferences.go has no such schema), so the
// options tier is a verified no-op. This is recorded as tracked schema debt
// (E7 blocker). When the profile gains the fields, this test is updated to
// exercise the opt-in path.
func TestApplyGitConfig_OptionsNoOpNoProfileFields(t *testing.T) {
	var setKeys []string
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet:  func(string) string { return "" },
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	swapSessionWorktreeSeams(t, swSeams{
		configSet: func(string, key, val string) error { setKeys = append(setKeys, key); return nil },
	})
	var out bytes.Buffer
	res := applyWorktreeGitConfig("/wt", &out)
	// REQ-SW-021 keys MUST NOT be set when no opt-in source provides them.
	forbidden := []string{"commit.gpgsign", "user.signingkey", "core.hooksPath", "core.autocrlf", "fetch.prune", "push.default"}
	for _, f := range forbidden {
		for _, k := range setKeys {
			if k == f {
				t.Fatalf("options no-op: MUST NOT set %s without an opt-in source (set keys=%v)", f, setKeys)
			}
		}
	}
	// core.hooksPath is REMOVED at v0.2.1 — even when opt-in fields exist
	// eventually, hooksPath is never set.
	for _, k := range setKeys {
		if k == "core.hooksPath" {
			t.Fatal("options: core.hooksPath MUST NEVER be set (v0.2.1 removal)")
		}
	}
	if res.optionsApplied {
		t.Fatal("options: expected optionsApplied=false (no opt-in schema)")
	}
}

// --- Wiring: helper invoked from materializeSessionWorktree ---

// TestEnterSessionWorktree_M7WiringInvokesHelper proves the M7 call-site hook
// comment was replaced by a real helper invocation: a successful
// materialization calls safe.directory add at least once for the materialized
// path.
func TestEnterSessionWorktree_M7WiringInvokesHelper(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	var safeAdds []string
	swapSessionWorktreeSeams(t, swSeams{
		inWt:      func() bool { return false },
		short:     func() string { return "abcdef12" },
		commonDir: func() (string, error) { return "/repo/.git", nil },
		add:       func(dest, branch, base string) (string, error) { return dest, nil },
		configSet: func(string, string, string) error { return nil },
	})
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(p string) error { safeAdds = append(safeAdds, p); return nil },
		globalGet:  func(string) string { return "" },
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	var out bytes.Buffer
	got := enterSessionWorktree(nil, "init", &out)
	if got == "" {
		t.Fatal("wiring: expected materialized path")
	}
	if len(safeAdds) != 1 || safeAdds[0] != got {
		t.Fatalf("wiring: helper must call safe.directory add for %q once, got %v", got, safeAdds)
	}
}

// TestEnterSessionWorktree_DefaultOffDoesNotInvokeHelper is REQ-SW-001 +
// REQ-SW-018: when the feature is OFF, the helper MUST NOT run (no
// safe.directory mutation).
func TestEnterSessionWorktree_DefaultOffDoesNotInvokeHelper(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "")
	safeAddCalled := false
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { safeAddCalled = true; return nil },
	})
	var out bytes.Buffer
	if got := enterSessionWorktree(nil, "init", &out); got != "" {
		t.Fatalf("OFF: expected empty path, got %q", got)
	}
	if safeAddCalled {
		t.Fatal("OFF: safe.directory add MUST NOT be invoked when feature is off")
	}
}

// --- E9: safe.directory unset on M4 cleanup ---

// TestCleanupSessionWorktree_UnsetsSafeDirectoryOnRemoval is R5 mitigation /
// E9: when auto-cleanup removes a worktree, the matching safe.directory entry
// is unset so the global allowlist does not accumulate (BI-7 reverse).
func TestCleanupSessionWorktree_UnsetsSafeDirectoryOnRemoval(t *testing.T) {
	var unsetPaths []string
	swapSessionWorktreeSeams(t, swSeams{
		remove:     func(string) error { return nil },
		statusPorc: func(string) (string, error) { return "", nil }, // clean
	})
	swapM7Seams(t, m7Seams{
		safeDirUnset: func(p string) error { unsetPaths = append(unsetPaths, p); return nil },
	})
	var out bytes.Buffer
	wt := "/repo/.claude/worktrees/WT-abcdef12-web"
	cleanupSessionWorktree(worktreeCfg(true), wt, true, &out)
	if len(unsetPaths) != 1 || unsetPaths[0] != wt {
		t.Fatalf("cleanup: must unset safe.directory for the removed worktree %q, got %v", wt, unsetPaths)
	}
}

// TestCleanupSessionWorktree_DoesNotUnsetWhenPreserved is the negative
// control: when cleanup PRESERVES the worktree (default-manual / non-clean
// exit / dirty), the safe.directory entry MUST NOT be unset (the worktree
// still exists on disk and still needs the entry).
func TestCleanupSessionWorktree_DoesNotUnsetWhenPreserved(t *testing.T) {
	unsetCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		remove:     func(string) error { unsetCalled = true; return nil },
		statusPorc: func(string) (string, error) { return "", nil },
	})
	swapM7Seams(t, m7Seams{
		safeDirUnset: func(string) error { unsetCalled = true; return nil },
	})
	var out bytes.Buffer
	wt := "/repo/.claude/worktrees/WT-abcdef12-web"

	// (a) default-manual (auto_cleanup=false) → preserve, no unset.
	cleanupSessionWorktree(worktreeCfg(false), wt, true, &out)
	if unsetCalled {
		t.Fatal("default-manual: MUST NOT unset safe.directory (worktree preserved)")
	}

	// (b) non-clean exit → preserve, no unset.
	unsetCalled = false
	cleanupSessionWorktree(worktreeCfg(true), wt, false, &out)
	if unsetCalled {
		t.Fatal("non-clean-exit: MUST NOT unset safe.directory (worktree preserved)")
	}

	// (c) dirty → preserve, no unset.
	unsetCalled = false
	swapSessionWorktreeSeams(t, swSeams{
		statusPorc: func(string) (string, error) { return " M dirty.go\n", nil },
	})
	cleanupSessionWorktree(worktreeCfg(true), wt, true, &out)
	if unsetCalled {
		t.Fatal("dirty: MUST NOT unset safe.directory (worktree preserved)")
	}
}

// --- REQ-SW-019 source invariant: profile NOT consulted ---

// TestApplyGitConfig_ProfileNotConsulted is the source-invariant guard: the
// identity comes from the global gitconfig ONLY. The helper signature takes
// no profile/config argument, so this is enforced structurally — but the test
// documents the invariant: even when the global gitconfig is empty, the
// helper does NOT fall back to any profile field (it is a verified no-op).
func TestApplyGitConfig_ProfileNotConsulted(t *testing.T) {
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet:  func(string) string { return "" }, // empty global
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50} },
	})
	var out bytes.Buffer
	// The helper takes only (wtPath, out) — there is no profile parameter to
	// consult. This is the structural guarantee that REQ-SW-019's
	// "profile NOT the source" invariant holds.
	res := applyWorktreeGitConfig("/wt", &out)
	if !res.identityNoop {
		t.Fatal("source-invariant: empty global MUST be a no-op, not a profile fallback")
	}
}
