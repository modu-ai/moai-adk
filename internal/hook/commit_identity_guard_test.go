package hook

// commit_identity_guard_test.go — SPEC-COMMIT-IDENTITY-GUARD-001 run-phase
// tests (AC-CIG-001..009, AC-CIG-012, AC-CIG-013, AC-CIG-014).
//
// This file matches the commit_identity_guard*_test.go exclusion glob of the
// fixture-email enumeration predicate on purpose: it carries OUT-OF-LIST
// control emails (dev@real-host.invalid, ops-bot@corp.invalid) that must
// never be promoted into builtinCommitIdentityDenyEmails.
//
// Fixture discipline (plan.md §D/§G): no test writes an identity into any
// git config file — the incident's channel B. Temporary-repository identity
// arrives only through the probe seam's environment list (cmd.Env); commits
// needed for worktree fixtures use `git -c` command-level identity, never
// `git config`. Tests that override the guard's package-level seams are NOT
// parallel.
//
// TDD note: this file was authored BEFORE the guard core compiled; the
// verbatim RED output of the first run (core refusal) is the §E E8 evidence.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// ─── out-of-list control values (excluded from enumeration; never in the
// built-in list) ───
const (
	ciControlEmail   = "dev@real-host.invalid"
	ciControlEmail2  = "ops-bot@corp.invalid"
	ciFixtureEmail   = "t@t.t"
	ciFixtureEmailB  = "t@t.test"
	ciProbeTimeoutVM = "probe failed (test-injected var error)"
)

// ─── helpers ───

// ciInput builds a shell tool input carrying command.
func ciInput(command string) *HookInput {
	return &HookInput{
		SessionID:     "sess-cig",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(`{"command": ` + strconv.Quote(command) + `}`),
	}
}

// ciCounts records the probe seams' invocations.
type ciCounts struct {
	scope, ident int
}

// ciOverrideProbes swaps in fake probes: the scope probe always reports
// sameDir (both sides equal — the "this repository" premise of AC-CIG-001
// through 007), the var probe returns the given idents or varErr.
func ciOverrideProbes(t *testing.T, sameDir string, authorIdent, committerIdent string, varErr error) *ciCounts {
	t.Helper()
	counts := &ciCounts{}
	prevScope, prevVar, prevEnv := commitIdentityScopeProbe, commitIdentityVarProbe, commitIdentityProbeEnv
	t.Cleanup(func() {
		commitIdentityScopeProbe, commitIdentityVarProbe, commitIdentityProbeEnv = prevScope, prevVar, prevEnv
	})
	commitIdentityScopeProbe = func(dir string) (string, error) {
		counts.scope++
		return sameDir, nil
	}
	commitIdentityVarProbe = func(dir string, env []string) (string, string, error) {
		counts.ident++
		if varErr != nil {
			return "", "", varErr
		}
		return authorIdent, committerIdent, nil
	}
	commitIdentityProbeEnv = func() []string { return []string{"PATH=" + os.Getenv("PATH")} }
	return counts
}

// ciIdent formats a probe ident line.
func ciIdent(name, email string) string {
	return fmt.Sprintf("%s <%s> 1760000000 +0900", name, email)
}

// ciAuditLines counts the guard audit log's lines under projectDir.
func ciAuditLines(t *testing.T, projectDir string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, commitIdentityAuditRelPath))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	text := strings.TrimRight(string(data), "\n")
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}

// ciInitRepo creates an empty git repository (no identity writes — channel B
// discipline) and returns its path.
func ciInitRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	repo := t.TempDir()
	out, err := exec.Command("git", "-C", repo, "init").CombinedOutput()
	if err != nil {
		t.Fatalf("git init %s: %v\n%s", repo, err, out)
	}
	return repo
}

// ciCountingRealProbes wraps the DEFAULT probes with invocation counters and
// returns the counts (used by the real-git integration tests).
func ciCountingRealProbes(t *testing.T) *ciCounts {
	t.Helper()
	counts := &ciCounts{}
	prevScope, prevVar, prevEnv := commitIdentityScopeProbe, commitIdentityVarProbe, commitIdentityProbeEnv
	t.Cleanup(func() {
		commitIdentityScopeProbe, commitIdentityVarProbe, commitIdentityProbeEnv = prevScope, prevVar, prevEnv
	})
	commitIdentityScopeProbe = func(dir string) (string, error) {
		counts.scope++
		return prevScope(dir)
	}
	commitIdentityVarProbe = func(dir string, env []string) (string, string, error) {
		counts.ident++
		return prevVar(dir, env)
	}
	return counts
}

// ciHandler builds a pre-tool handler over cfg with projectDir.
func ciHandler(cfg *config.Config, projectDir string) *preToolHandler {
	return &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}
}

// ─── AC-CIG-001 — a config-layer fixture identity commit is denied ───

func TestAC_CIG_001_FixtureIdentityCommitDenied(t *testing.T) {
	projectDir := t.TempDir()
	cases := []struct {
		name      string
		author    string
		committer string
		wantRole  string
	}{
		{"author_only", ciIdent("T", ciFixtureEmail), ciIdent("Dev", ciControlEmail), "author"},
		{"committer_only", ciIdent("Dev", ciControlEmail), ciIdent("T", ciFixtureEmail), "committer"},
		{"both", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), "author"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ciOverrideProbes(t, "/ci-common", tc.author, tc.committer, nil)
			decision, reason := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effectiveDenyEmails(nil))
			if decision != DecisionDeny {
				t.Fatalf("decision = %q, want deny", decision)
			}
			if !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") {
				t.Fatalf("reason %q does not start with the sentinel", reason)
			}
			if !strings.Contains(reason, ciFixtureEmail) {
				t.Fatalf("reason %q does not name the matched email", reason)
			}
			if !strings.Contains(reason, tc.wantRole) {
				t.Fatalf("reason %q does not name the role %q", reason, tc.wantRole)
			}
			if !strings.Contains(strings.ToLower(reason), "fix") {
				t.Fatalf("reason %q carries no remedy", reason)
			}
		})
	}
}

// ─── AC-CIG-002 — all eight trigger verbs are examined ───

func TestAC_CIG_002_AllEightTriggerVerbsDenied(t *testing.T) {
	projectDir := t.TempDir()
	verbs := []string{"commit", "merge", "cherry-pick", "revert", "rebase", "am", "commit-tree", "pull"}
	forms := []struct {
		name string
		wrap func(verb string) string
	}{
		{"plain", func(v string) string { return "git " + v + " arg1" }},
		{"exe", func(v string) string { return "git.exe " + v + " arg1" }},
		{"chained", func(v string) string { return "echo hi && git " + v + " arg1" }},
		{"globalC", func(v string) string { return "git -C /tmp/somewhere " + v + " arg1" }},
	}
	for _, verb := range verbs {
		for _, form := range forms {
			verb, form := verb, form
			t.Run(verb+"_"+form.name, func(t *testing.T) {
				ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)
				decision, reason := checkCommitIdentity(ciInput(form.wrap(verb)), projectDir, effectiveDenyEmails(nil))
				if decision != DecisionDeny || !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") {
					t.Fatalf("verb %q form %q: decision=%q reason=%q, want deny with sentinel", verb, form.name, decision, reason)
				}
			})
		}
	}
}

// ─── AC-CIG-003 — non-trigger commands run zero probes ───

func TestAC_CIG_003_NonTriggerRunsNoProbes(t *testing.T) {
	projectDir := t.TempDir()
	commands := []string{
		"git status",
		"git log --oneline",
		"git diff",
		"echo 'git commit -m x'",
		"# git commit",
		"cat <<EOF\ngit commit -m x\nEOF",
	}
	for _, cmd := range commands {
		cmd := cmd
		t.Run(strconv.Itoa(len(cmd))+"_"+strings.ReplaceAll(cmd, "\n", "\\n")[:minInt(20, len(strings.ReplaceAll(cmd, "\n", "\\n")))], func(t *testing.T) {
			counts := ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)
			decision, reason := checkCommitIdentity(ciInput(cmd), projectDir, effectiveDenyEmails(nil))
			if decision == DecisionDeny {
				t.Fatalf("command %q denied: %s", cmd, reason)
			}
			if counts.scope != 0 || counts.ident != 0 {
				t.Fatalf("command %q: scope=%d ident=%d probe calls, want 0/0", cmd, counts.scope, counts.ident)
			}
		})
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── AC-CIG-004 — command-level overrides are judged by git precedence ───

func TestAC_CIG_004_CommandLevelOverrides(t *testing.T) {
	projectDir := t.TempDir()
	// The fake probe resolves BOTH roles to the control email — the seam-level
	// model of a repository whose user.email config resolves outside the list.
	deny := []string{
		"git_author_env:GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " git commit -m x",
		"git_committer_env:GIT_COMMITTER_EMAIL=" + ciFixtureEmailB + " git commit -m x",
		"export_author:export GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " && git commit -m x",
		"c_user_email:git -c user.email=" + ciFixtureEmail + " commit -m x",
		"author_option:git commit --author=\"t <" + ciFixtureEmail + ">\" -m x",
	}
	allow := []string{
		"git_author_env:GIT_AUTHOR_EMAIL=" + ciControlEmail + " git commit -m x",
		"git_committer_env:GIT_COMMITTER_EMAIL=" + ciControlEmail + " git commit -m x",
		"export_author:export GIT_AUTHOR_EMAIL=" + ciControlEmail + " && git commit -m x",
		"c_user_email:git -c user.email=" + ciControlEmail + " commit -m x",
		"author_option:git commit --author=\"dev <" + ciControlEmail + ">\" -m x",
		"both_env_overrides:GIT_AUTHOR_EMAIL=" + ciControlEmail + " GIT_COMMITTER_EMAIL=" + ciControlEmail + " git commit -m x",
		"c_overrides_both:git -c user.email=" + ciControlEmail + " commit -m x",
		"email_shadowed_by_config:EMAIL=" + ciFixtureEmail + " git commit -m x",
		"author_option_replaces_default:git commit --author=\"dev <" + ciControlEmail + ">\" -m x",
	}
	for _, tc := range deny {
		tc := tc
		t.Run("deny_"+tc[:strings.Index(tc, ":")], func(t *testing.T) {
			ciOverrideProbes(t, "/ci-common", ciIdent("Dev", ciControlEmail), ciIdent("Dev", ciControlEmail), nil)
			cmd := tc[strings.Index(tc, ":")+1:]
			decision, reason := checkCommitIdentity(ciInput(cmd), projectDir, effectiveDenyEmails(nil))
			if decision != DecisionDeny || !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") {
				t.Fatalf("%s: decision=%q reason=%q, want deny", cmd, decision, reason)
			}
		})
	}
	for _, tc := range allow {
		tc := tc
		t.Run("allow_"+tc[:strings.Index(tc, ":")], func(t *testing.T) {
			ciOverrideProbes(t, "/ci-common", ciIdent("Dev", ciControlEmail), ciIdent("Dev", ciControlEmail), nil)
			cmd := tc[strings.Index(tc, ":")+1:]
			decision, reason := checkCommitIdentity(ciInput(cmd), projectDir, effectiveDenyEmails(nil))
			if decision == DecisionDeny {
				t.Fatalf("%s: denied (%s), want allow", cmd, reason)
			}
		})
	}

	t.Run("author_option_replaces_fixture_default", func(t *testing.T) {
		// Default author resolves to the fixture identity; --author replaces
		// only the author, and the committer stays outside the list → allow.
		ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("Dev", ciControlEmail), nil)
		decision, reason := checkCommitIdentity(ciInput("git commit --author=\"dev <"+ciControlEmail+">\" -m x"), projectDir, effectiveDenyEmails(nil))
		if decision == DecisionDeny {
			t.Fatalf("denied (%s); --author should replace the fixture author while the committer is outside the list", reason)
		}
	})

	t.Run("email_env_denies_when_no_config_email_resolves", func(t *testing.T) {
		// The probe FAILS (no resolvable identity — the model of a repository
		// without user.email) and the command carries no higher override, so
		// the EMAIL question is undecidable: allow with one audit line
		// (REQ-CIG-007), never a deny.
		ciOverrideProbes(t, "/ci-common", "", "", fmt.Errorf("%s", ciProbeTimeoutVM))
		decision, _ := checkCommitIdentity(ciInput("EMAIL="+ciFixtureEmail+" git commit -m x"), projectDir, effectiveDenyEmails(nil))
		if decision == DecisionDeny {
			t.Fatal("denied; EMAIL alone is not positive evidence when the probe cannot resolve the config layer")
		}
	})
}

// ─── AC-CIG-005 — a real identity passes through the real probe (positive
// control), with its fixture pair denied ───

func TestAC_CIG_005_RealProbeRealIdentityAllowed(t *testing.T) {
	repo := ciInitRepo(t)
	counts := ciCountingRealProbes(t)

	identityEnv := func(email string) []string {
		return append(gitenv.Scrub(os.Environ()),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=Real Dev",
			"GIT_AUTHOR_EMAIL="+email,
			"GIT_COMMITTER_NAME=Real Dev",
			"GIT_COMMITTER_EMAIL="+email,
		)
	}
	prevEnv := commitIdentityProbeEnv
	t.Cleanup(func() { commitIdentityProbeEnv = prevEnv })

	t.Run("real_identity_allows_with_zero_audit_lines", func(t *testing.T) {
		commitIdentityProbeEnv = func() []string { return identityEnv(ciControlEmail) }
		input := ciInput("git commit -m x")
		input.CWD = repo // AC-CIG-005 Given: the temp repo is both project dir and command cwd
		decision, reason := checkCommitIdentity(input, repo, effectiveDenyEmails(nil))
		if decision == DecisionDeny {
			t.Fatalf("denied: %s", reason)
		}
		// The real probe returned exactly the injected identity.
		author, committer, err := commitIdentityVarProbe(repo, identityEnv(ciControlEmail))
		if err != nil {
			t.Fatalf("real var probe: %v", err)
		}
		for name, ident := range map[string]string{"author": author, "committer": committer} {
			got, perr := emailFromIdent(ident)
			if perr != nil || got != ciControlEmail {
				t.Fatalf("%s ident %q resolved %q (%v), want %s", name, ident, got, perr, ciControlEmail)
			}
		}
		if n := ciAuditLines(t, repo); n != 0 {
			t.Fatalf("audit log gained %d lines, want 0", n)
		}
		if counts.ident < 1 {
			t.Fatal("the real probe seam never ran; the allow is not carried by the resolution path")
		}
	})

	t.Run("fixture_pair_denies", func(t *testing.T) {
		commitIdentityProbeEnv = func() []string { return identityEnv(ciFixtureEmail) }
		input := ciInput("git commit -m x")
		input.CWD = repo // same Given: without it the scope stage allows before identity resolution runs
		decision, reason := checkCommitIdentity(input, repo, effectiveDenyEmails(nil))
		if decision != DecisionDeny || !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") {
			t.Fatalf("decision=%q reason=%q, want deny", decision, reason)
		}
	})
}

// ─── AC-CIG-006 — exact match only ───

func TestAC_CIG_006_ExactMatchOnly(t *testing.T) {
	projectDir := t.TempDir()
	cases := []struct {
		email string
		want  string
	}{
		{"T@T.T", DecisionDeny},
		{" t@t.t ", DecisionDeny},
		{"t@t.tt", ""},
		{"xt@t.t", ""},
		{"t@t.t.example.org", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.email, func(t *testing.T) {
			ciOverrideProbes(t, "/ci-common", ciIdent("P", tc.email), ciIdent("P", ciControlEmail), nil)
			decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effectiveDenyEmails(nil))
			if tc.want == DecisionDeny && decision != DecisionDeny {
				t.Fatalf("probe email %q: decision=%q, want deny", tc.email, decision)
			}
			if tc.want == "" && decision == DecisionDeny {
				t.Fatalf("probe email %q denied; only exact matches deny", tc.email)
			}
		})
	}
}

// ─── AC-CIG-007 — every resolution failure allows with exactly one audit
// line; a decisive deny-listed override denies without an audit line ───

func TestAC_CIG_007_ResolutionFailureAllowsWithAudit(t *testing.T) {

	t.Run("var_probe_error", func(t *testing.T) {
		projectDir := t.TempDir()
		ciOverrideProbes(t, "/ci-common", "", "", fmt.Errorf("%s", ciProbeTimeoutVM))
		if decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effectiveDenyEmails(nil)); decision == DecisionDeny {
			t.Fatal("denied on probe error; fail-open required")
		}
		if n := ciAuditLines(t, projectDir); n != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", n)
		}
	})

	t.Run("unparseable_output", func(t *testing.T) {
		projectDir := t.TempDir()
		ciOverrideProbes(t, "/ci-common", "no brackets here", "no brackets here", nil)
		if decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effectiveDenyEmails(nil)); decision == DecisionDeny {
			t.Fatal("denied on unparseable output; fail-open required")
		}
		if n := ciAuditLines(t, projectDir); n != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", n)
		}
	})

	t.Run("missing_cwd", func(t *testing.T) {
		projectDir := t.TempDir()
		// REAL scope probe (no override): the nonexistent cwd fails
		// resolution on the target side.
		input := ciInput("git commit -m x")
		input.CWD = filepath.Join(projectDir, "does-not-exist-t1289")
		if decision, _ := checkCommitIdentity(input, projectDir, effectiveDenyEmails(nil)); decision == DecisionDeny {
			t.Fatal("denied on missing cwd; fail-open required")
		}
		if n := ciAuditLines(t, projectDir); n != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", n)
		}
	})

	t.Run("target_scope_failure", func(t *testing.T) {
		projectDir := t.TempDir()
		counts := ciOverrideProbes(t, "/ci-common", "", "", nil)
		commitIdentityScopeProbe = func(string) (string, error) {
			return "", fmt.Errorf("target side git rev-parse failed (test-injected)")
		}
		if decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effectiveDenyEmails(nil)); decision == DecisionDeny {
			t.Fatal("denied on target scope failure; fail-open required")
		}
		if n := ciAuditLines(t, projectDir); n != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", n)
		}
		if counts.ident != 0 {
			t.Fatalf("identity probes = %d, want 0 (the scope question comes first)", counts.ident)
		}
	})

	t.Run("project_scope_failure", func(t *testing.T) {
		projectDir := t.TempDir()
		ciOverrideProbes(t, "/ci-common", "", "", nil)
		commitIdentityScopeProbe = func(dir string) (string, error) {
			if dir == projectDir {
				return "", fmt.Errorf("project side git rev-parse failed (test-injected)")
			}
			return "/ci-common", nil
		}
		if decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effectiveDenyEmails(nil)); decision == DecisionDeny {
			t.Fatal("denied on project scope failure; fail-open required")
		}
		if n := ciAuditLines(t, projectDir); n != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", n)
		}
	})

	t.Run("decisive_override_denies_despite_probe_failure", func(t *testing.T) {
		projectDir := t.TempDir()
		ciOverrideProbes(t, "/ci-common", "", "", fmt.Errorf("%s", ciProbeTimeoutVM))
		before := ciAuditLines(t, projectDir)
		decision, reason := checkCommitIdentity(ciInput("GIT_AUTHOR_EMAIL="+ciFixtureEmail+" git commit -m x"), projectDir, effectiveDenyEmails(nil))
		if decision != DecisionDeny || !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") {
			t.Fatalf("decision=%q reason=%q; the decisive deny-listed override is positive evidence", decision, reason)
		}
		if n := ciAuditLines(t, projectDir); n != before {
			t.Fatalf("audit lines %d -> %d; a deny adds no fail-open line", before, n)
		}
	})
}

// TestAC_CIG_007b_ProbeTimeoutRealPath exercises the REAL probe's time bound:
// a git shim that sleeps past the (shrunken) budget makes the probe fail, and
// the guard fails open with one audit line.
func TestAC_CIG_007b_ProbeTimeoutRealPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script git shim needs a POSIX shell")
	}
	requireGit(t)
	repo := ciInitRepo(t)
	projectDir := t.TempDir()

	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git not found: %v", err)
	}
	shimDir := t.TempDir()
	if err := os.Symlink(realGit, filepath.Join(shimDir, "git.real")); err != nil {
		t.Fatalf("symlink real git: %v", err)
	}
	// `$3 == "var"` — `git -C <dir> var GIT_*_IDENT`: only the identity probe
	// sleeps; the scope probe (rev-parse) execs the real git immediately.
	shim := "#!/bin/sh\nif [ \"$3\" = \"var\" ]; then sleep 5; fi\nexec \"" + filepath.Join(shimDir, "git.real") + "\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(shim), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	prevTimeout := commitIdentityProbeTimeout
	commitIdentityProbeTimeout = 150 * time.Millisecond
	t.Cleanup(func() { commitIdentityProbeTimeout = prevTimeout })

	input := ciInput("git commit -m x")
	input.CWD = repo
	if decision, _ := checkCommitIdentity(input, projectDir, effectiveDenyEmails(nil)); decision == DecisionDeny {
		t.Fatal("denied on probe timeout; fail-open required")
	}
	if n := ciAuditLines(t, projectDir); n != 1 {
		t.Fatalf("audit lines = %d, want exactly 1 (timeout is a REQ-CIG-007 failure)", n)
	}
}

// ─── AC-CIG-008 — the disabled guard is never invoked ───

func TestAC_CIG_008_DisabledGuardRunsNoProbes(t *testing.T) {
	repo := t.TempDir() // any directory; the probes must never run at all
	for _, tc := range []struct {
		name string
		cfg  *config.Config
	}{
		{"engine_default", config.NewDefaultConfig()},
		{"explicit_false", func() *config.Config {
			c := config.NewDefaultConfig()
			c.Workflow.CommitIdentityGuard.Enabled = false
			return c
		}()},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			counts := ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)
			out, err := ciHandler(tc.cfg, repo).Handle(context.Background(), ciInput("git commit -m x"))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if decisionOf(out) == DecisionDeny {
				t.Fatalf("disabled guard denied: %s", reasonOf(out))
			}
			if counts.scope != 0 || counts.ident != 0 {
				t.Fatalf("disabled guard probed: scope=%d ident=%d, want 0/0 (REQ-CIG-006)", counts.scope, counts.ident)
			}
		})
	}

	t.Run("config_default_false", func(t *testing.T) {
		if config.NewDefaultConfig().Workflow.CommitIdentityGuard.Enabled {
			t.Fatal("Workflow.CommitIdentityGuard.Enabled = true in the default config, want false")
		}
	})
}

// TestCommitIdentityGuard_EnabledFlipsDecision is the non-vacuity half of the
// wiring gate: with everything else identical, ONLY the flag flips the
// decision.
func TestCommitIdentityGuard_EnabledFlipsDecision(t *testing.T) {
	repo := t.TempDir()
	counts := ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)
	cfgOn := config.NewDefaultConfig()
	cfgOn.Workflow.CommitIdentityGuard.Enabled = true
	out, err := ciHandler(cfgOn, repo).Handle(context.Background(), ciInput("git commit -m x"))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if decisionOf(out) != DecisionDeny || !strings.HasPrefix(reasonOf(out), "TEST_IDENTITY_VIOLATION:") {
		t.Fatalf("enabled guard: decision=%q reason=%q, want deny with sentinel", decisionOf(out), reasonOf(out))
	}
	if counts.ident == 0 {
		t.Fatal("the enabled guard never probed; the deny is not carried by the identity path")
	}
}

// ─── AC-CIG-009 — the config list adds to the built-in list; the yaml files
// carry the key with the right values ───

func TestAC_CIG_009_ConfigListAddsToBuiltin(t *testing.T) {
	projectDir := t.TempDir()
	effective := effectiveDenyEmails([]string{ciControlEmail2})

	t.Run("config_email_denies", func(t *testing.T) {
		ciOverrideProbes(t, "/ci-common", ciIdent("Ops", ciControlEmail2), ciIdent("Ops", ciControlEmail2), nil)
		decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effective)
		if decision != DecisionDeny {
			t.Fatal("deny_emails entry did not deny")
		}
	})
	t.Run("builtin_email_still_denies", func(t *testing.T) {
		ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)
		decision, _ := checkCommitIdentity(ciInput("git commit -m x"), projectDir, effective)
		if decision != DecisionDeny {
			t.Fatal("config list replaced the built-in list; the union is required (REQ-CIG-008)")
		}
	})
}

func TestAC_CIG_009_WorkflowYamlCarriesKey(t *testing.T) {
	root := moduleRootFromCaller(t)
	cases := []struct {
		path   string
		wantOn bool
	}{
		{filepath.Join(root, "internal/template/templates/.moai/config/sections/workflow.yaml"), false},
		{filepath.Join(root, ".moai/config/sections/workflow.yaml"), true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path[len(root)+1:], func(t *testing.T) {
			data, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			var hit bool
			inBlock := false
			for _, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "commit_identity_guard:") {
					hit = true
					inBlock = true
					continue
				}
				if inBlock {
					if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
						inBlock = false
						continue
					}
					if strings.HasPrefix(trimmed, "enabled:") {
						got := strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:"))
						if tc.wantOn && got != "true" {
							t.Fatalf("local workflow.yaml commit_identity_guard.enabled = %s, want true", got)
						}
						if !tc.wantOn && got != "false" {
							t.Fatalf("template workflow.yaml commit_identity_guard.enabled = %s, want false", got)
						}
					}
				}
			}
			if !hit {
				t.Fatal("commit_identity_guard key missing from workflow.yaml")
			}
		})
	}

	t.Run("template_block_neutral", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(root, "internal/template/templates/.moai/config/sections/workflow.yaml"))
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var block []string
		inBlock := false
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "commit_identity_guard:") {
				inBlock = true
				block = append(block, line)
				continue
			}
			if inBlock {
				if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
					inBlock = false
					continue
				}
				block = append(block, line)
			}
		}
		text := strings.Join(block, "\n")
		if len(block) == 0 {
			t.Fatal("template block not found")
		}
		for _, forbidden := range []string{"SPEC-", "t1289", "t12", "2026-"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("template block carries internal state %q (template neutrality §25):\n%s", forbidden, text)
			}
		}
	})
}

// ─── AC-CIG-012 — wiring order, PowerShell unclassified, sentinel ───

func TestAC_CIG_012_WiringPreservesEarlierDeny(t *testing.T) {
	ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)

	t.Run("destructive_command_check_wins", func(t *testing.T) {
		repo := t.TempDir()
		cfg := config.NewDefaultConfig()
		cfg.Workflow.CommitIdentityGuard.Enabled = true
		out, err := ciHandler(cfg, repo).Handle(context.Background(), ciInput("git commit -m x; rm -rf /"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if decisionOf(out) != DecisionDeny {
			t.Fatal("want deny from the destructive-command check")
		}
		if strings.Contains(reasonOf(out), "TEST_IDENTITY_VIOLATION") {
			t.Fatalf("the identity sentinel displaced the earlier deny: %s", reasonOf(out))
		}
	})

	t.Run("no_verify_defense_wins", func(t *testing.T) {
		repo := t.TempDir()
		cfg := config.NewDefaultConfig()
		cfg.Workflow.CommitIdentityGuard.Enabled = true
		out, err := ciHandler(cfg, repo).Handle(context.Background(), ciInput("git commit --no-verify -m x"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if decisionOf(out) != DecisionDeny || strings.Contains(reasonOf(out), "TEST_IDENTITY_VIOLATION") {
			t.Fatalf("decision=%q reason=%q; the earlier guard must win", decisionOf(out), reasonOf(out))
		}
	})

	t.Run("branch_guard_wins", func(t *testing.T) {
		repo := newBranchGuardRepoFixture(t)
		t.Setenv(branchGuardExemptEnv, "")
		cfg := config.NewDefaultConfig()
		cfg.Workflow.BranchGuard.Enabled = true
		cfg.Workflow.CommitIdentityGuard.Enabled = true
		input := ciInput("git switch -c cig-bypass-branch && git commit -m x")
		input.CWD = repo
		out, err := ciHandler(cfg, repo).Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if decisionOf(out) != DecisionDeny || !strings.Contains(reasonOf(out), branchGuardViolationPrefix) {
			t.Fatalf("decision=%q reason=%q; want the branch guard's deny", decisionOf(out), reasonOf(out))
		}
		if strings.Contains(reasonOf(out), "TEST_IDENTITY_VIOLATION") {
			t.Fatalf("the identity sentinel displaced the branch guard: %s", reasonOf(out))
		}
	})

	t.Run("push_readiness_wins", func(t *testing.T) {
		// The full Handle path under contract mode runs the escalation
		// detector (pre_tool.go observeEscalationWith), which records an
		// observation under paths.MoaiHome(). Redirect MoaiHome to this
		// test's own tempdir so the write stays out of the SHARED package
		// test home — TestEscalationGuidedGolden walks that home and would
		// see the stray db/primary-*/contract/escalation directory.
		t.Setenv("MOAI_HOME", t.TempDir())
		f := closuretest.New(t)
		queueC1(t, f)
		writeGitFlowConfig(t, f)
		f.Git(f.Root, "branch", "main", "HEAD")
		f.Git(f.Root, "worktree", "add", filepath.Join(f.Parent, "cig-main-tree"), "-b", "cig-main-tree", "main")
		mainTree := filepath.Join(f.Parent, "cig-main-tree")
		f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
		closureReportPair(t, f)

		cfg := autonomyProvider("contract").c
		cfg.Workflow.CommitIdentityGuard.Enabled = true
		input := ciInput("git commit -m x && git push origin develop")
		input.CWD = mainTree
		out, err := ciHandler(cfg, f.Root).Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if decisionOf(out) != DecisionDeny || !strings.Contains(reasonOf(out), ClosurePushStopPrefix) {
			t.Fatalf("decision=%q reason=%q; want the push-readiness deny", decisionOf(out), reasonOf(out))
		}
		if !strings.Contains(reasonOf(out), "second_review_not_performed") {
			t.Fatalf("reason %q does not name second_review_not_performed", reasonOf(out))
		}
		if strings.Contains(reasonOf(out), "TEST_IDENTITY_VIOLATION") {
			t.Fatalf("the identity sentinel displaced the push-readiness deny: %s", reasonOf(out))
		}
	})
}

func TestAC_CIG_012_PowerShellUnclassified(t *testing.T) {
	projectDir := t.TempDir()
	counts := ciOverrideProbes(t, "/ci-common", ciIdent("T", ciFixtureEmail), ciIdent("T", ciFixtureEmail), nil)
	input := ciInput("Invoke-Expression 'git commit -m x'")
	input.ToolName = "PowerShell"
	decision, reason := checkCommitIdentity(input, projectDir, effectiveDenyEmails(nil))
	if decision == DecisionDeny {
		t.Fatalf("unclassifiable PowerShell denied: %s", reason)
	}
	if counts.scope != 0 || counts.ident != 0 {
		t.Fatalf("probes ran on an unclassifiable call: scope=%d ident=%d, want 0/0", counts.scope, counts.ident)
	}
	if n := ciAuditLines(t, projectDir); n != 1 {
		t.Fatalf("audit lines = %d, want exactly 1 unclassified line", n)
	}
	data, _ := os.ReadFile(filepath.Join(projectDir, commitIdentityAuditRelPath))
	if !strings.Contains(string(data), "unclassifiable") {
		t.Fatalf("audit line does not carry the unclassifiable reason:\n%s", data)
	}
}

func TestAC_CIG_012_SentinelDefinedOnce(t *testing.T) {
	root := moduleRootFromCaller(t)
	dir := filepath.Join(root, "internal", "hook")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	hits := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, rerr := os.ReadFile(filepath.Join(dir, name))
		if rerr != nil {
			t.Fatalf("read %s: %v", name, rerr)
		}
		hits += strings.Count(string(data), `"TEST_IDENTITY_VIOLATION`)
	}
	if hits != 1 {
		t.Fatalf(`'"TEST_IDENTITY_VIOLATION"' appears %d times in non-test internal/hook files, want exactly 1 (the definition)`, hits)
	}
}

// ─── AC-CIG-013 — another repository's fixture commit is allowed (real git) ───

func TestAC_CIG_013_OtherRepositoryAllowed(t *testing.T) {
	project := ciInitRepo(t) // P
	fixture := ciInitRepo(t) // F
	counts := ciCountingRealProbes(t)

	cases := []struct {
		name    string
		cwd     string
		command string
		want    string // "" allow, "deny"
	}{
		{"cwd_is_fixture_repo", fixture, "GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " GIT_COMMITTER_EMAIL=" + ciFixtureEmail + " git commit -m x", ""},
		{"c_into_fixture_repo", project, "git -C " + fixture + " -c user.email=" + ciFixtureEmail + " commit -m x", ""},
		{"cd_into_fixture_repo", project, "cd " + fixture + " && GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " git commit -m x", ""},
		{"pair_cwd_is_project", project, "GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " GIT_COMMITTER_EMAIL=" + ciFixtureEmail + " git commit -m x", "deny"},
		{"pair_c_into_project", project, "git -C " + project + " -c user.email=" + ciFixtureEmail + " commit -m x", "deny"},
		{"pair_cd_into_project", project, "cd " + project + " && GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " git commit -m x", "deny"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			input := ciInput(tc.command)
			input.CWD = tc.cwd
			before := counts.ident
			decision, reason := checkCommitIdentity(input, project, effectiveDenyEmails(nil))
			if tc.want == "deny" {
				if decision != DecisionDeny || !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") {
					t.Fatalf("decision=%q reason=%q, want deny with sentinel", decision, reason)
				}
				return
			}
			if decision == DecisionDeny {
				t.Fatalf("other-repository command denied: %s", reason)
			}
			if counts.ident != before {
				t.Fatalf("identity probes %d -> %d; a scope mismatch must not probe identity (REQ-CIG-012)", before, counts.ident)
			}
		})
	}
}

// ─── AC-CIG-014 — this repository's linked worktree commit is denied
// (real git, real scope probe) ───

func TestAC_CIG_014_LinkedWorktreeDenied(t *testing.T) {
	requireGit(t)
	project := t.TempDir() // P
	run := func(dir string, args ...string) {
		full := append([]string{"-C", dir}, args...)
		out, err := exec.Command("git", full...).CombinedOutput()
		if err != nil {
			t.Fatalf("git -C %s %v: %v\n%s", dir, args, err, out)
		}
	}
	run(project, "init")
	// Seed commit with COMMAND-LEVEL identity (`git -c`): no config file write.
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=Seed", "GIT_AUTHOR_EMAIL=seed@seed.invalid",
		"GIT_COMMITTER_NAME=Seed", "GIT_COMMITTER_EMAIL=seed@seed.invalid",
	)
	cmd := exec.Command("git", "-C", project, "commit", "--allow-empty", "-m", "seed")
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("seed commit: %v\n%s", err, out)
	}
	worktree := filepath.Join(t.TempDir(), "linked-wt")
	run(project, "worktree", "add", worktree, "-b", "cig-linked-wt")

	ciCountingRealProbes(t) // count, but keep the real behavior

	if project == worktree || project == filepath.Clean(worktree) {
		t.Fatal("P and W paths are equal; this is not a linked worktree measurement")
	}

	cases := []struct {
		name    string
		cwd     string
		command string
	}{
		{"cwd_is_worktree", worktree, "GIT_AUTHOR_EMAIL=" + ciFixtureEmail + " git commit -m x"},
		{"c_into_worktree", project, "git -C " + worktree + " -c user.email=" + ciFixtureEmail + " commit -m x"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			input := ciInput(tc.command)
			input.CWD = tc.cwd
			decision, reason := checkCommitIdentity(input, project, effectiveDenyEmails(nil))
			if decision != DecisionDeny || !strings.HasPrefix(reason, "TEST_IDENTITY_VIOLATION:") || !strings.Contains(reason, ciFixtureEmail) {
				t.Fatalf("decision=%q reason=%q; a linked worktree shares this repository's common dir and must deny", decision, reason)
			}
		})
	}
}

// ─── helper unit coverage — the pure parsing/normalization helpers the AC
// tests exercise only incidentally. Table-driven; no probes, no subprocess. ───

func TestCommitIdentityGuard_HelperUnits(t *testing.T) {
	t.Run("shellFields_keeps_quoted_groups_attached", func(t *testing.T) {
		cases := []struct {
			in   string
			want []string
		}{
			{"git commit -m x", []string{"git", "commit", "-m", "x"}},
			{`--author="t <t@t.t>"`, []string{`--author="t <t@t.t>"`}},
			{`-C "my path"`, []string{"-C", `"my path"`}},
			{`a\ b c`, []string{`a\ b`, "c"}},
			{"  spaced\tout  ", []string{"spaced", "out"}},
		}
		for _, tc := range cases {
			got := shellFields(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("shellFields(%q) = %q, want %q", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("shellFields(%q)[%d] = %q, want %q", tc.in, i, got[i], tc.want[i])
				}
			}
		}
	})

	t.Run("splitShellSegments_excludes_heredoc_and_comment_text", func(t *testing.T) {
		cases := []struct {
			name    string
			command string
			want    []string // segment texts in order
		}{
			{"plain", "a && b; c", []string{"a", "b", "c"}},
			{"comment carries no command", "# git commit -m x", []string{}},
			{"trailing comment trimmed", "git status # git commit", []string{"git status"}},
			{"heredoc body excluded", "cat <<EOF\nit commits\nEOF\ngit status", []string{"cat <<EOF", "git status"}},
		}
		for _, tc := range cases {
			segs := splitShellSegments(tc.command)
			if len(segs) != len(tc.want) {
				t.Fatalf("%s: segments = %q, want %q", tc.name, segs, tc.want)
			}
			for i := range segs {
				if segs[i].text != tc.want[i] {
					t.Fatalf("%s: segment[%d] = %q, want %q", tc.name, i, segs[i].text, tc.want[i])
				}
			}
		}
	})

	t.Run("splitShellSegment_connectors", func(t *testing.T) {
		segs := splitShellSegments("cd /tmp && export A=1 && git commit -m x")
		if len(segs) != 3 {
			t.Fatalf("segments = %d, want 3", len(segs))
		}
		if segs[0].conn != "" || segs[1].conn != "&&" || segs[2].conn != "&&" {
			t.Fatalf("connectors = %q %q %q, want \"\" && &&", segs[0].conn, segs[1].conn, segs[2].conn)
		}
		segs = splitShellSegments("a; b | c || d")
		if len(segs) != 4 || segs[1].conn != ";" || segs[2].conn != "|" || segs[3].conn != "||" {
			t.Fatalf("connector chain wrong: %+v", segs)
		}
	})

	t.Run("gitGlobalOptionSpan", func(t *testing.T) {
		next := []string{"commit", "-m"}
		cases := []struct {
			tok  string
			span int
			ok   bool
		}{
			{"-C", 2, true},
			{"-c", 2, true},
			{"-C/tmp/repo", 1, true},
			{"-cuser.email=x@y.z", 1, true},
			{"--git-dir", 2, true},
			{"--git-dir=/x", 1, true},
			{"--bare", 1, true},
			{"commit", 0, false},
		}
		for _, tc := range cases {
			nx := next
			if tc.ok {
				_ = nx
			}
			span, ok := gitGlobalOptionSpan(tc.tok, next)
			if span != tc.span || ok != tc.ok {
				t.Fatalf("gitGlobalOptionSpan(%q) = %d,%v; want %d,%v", tc.tok, span, ok, tc.span, tc.ok)
			}
		}
		// The no-following-token form: a two-token global option at the end is
		// not a spanning option.
		if span, ok := gitGlobalOptionSpan("-C", nil); span != 0 || ok {
			t.Fatalf("gitGlobalOptionSpan(-C, nil) = %d,%v; want 0,false", span, ok)
		}
	})

	t.Run("emailFromAuthorValue", func(t *testing.T) {
		cases := []struct {
			in   string
			want string
		}{
			{`"t <t@t.t>"`, "t@t.t"},
			{"t@t.t", "t@t.t"},
			{"plain-name", ""},
			{"no angle t@t.t", "no angle t@t.t"}, // a bare @-carrying value returns verbatim
		}
		for _, tc := range cases {
			if got := emailFromAuthorValue(tc.in); got != tc.want {
				t.Fatalf("emailFromAuthorValue(%q) = %q, want %q", tc.in, got, tc.want)
			}
		}
	})

	t.Run("resolveAgainst", func(t *testing.T) {
		if got := resolveAgainst("/base", "/abs/p"); got != "/abs/p" {
			t.Fatalf("absolute passthrough = %q", got)
		}
		if got := resolveAgainst("/base", "rel/p"); got != "/base/rel/p" {
			t.Fatalf("relative join = %q", got)
		}
		if got := resolveAgainst("", "rel"); got != "" {
			t.Fatalf("empty base = %q, want empty", got)
		}
		if got := resolveAgainst("/base", "  "); got != "" {
			t.Fatalf("blank operand = %q, want empty", got)
		}
	})

	t.Run("sameCommonDir_and_normalizeRepoPath", func(t *testing.T) {
		if sameCommonDir("", "/x") || sameCommonDir("/x", "") {
			t.Fatal("empty operand must compare unequal")
		}
		if !sameCommonDir("/Repo/Common", "/repo/common") {
			t.Fatal("comparison must be case-insensitive")
		}
		if normalizeRepoPath("") != "" {
			t.Fatal("empty input must normalize to empty")
		}
		if got := normalizeRepoPath("/a/../b"); got != "/b" {
			t.Fatalf("clean = %q, want /b", got)
		}
	})

	t.Run("emailFromIdent_rejects_unparseable", func(t *testing.T) {
		if _, err := emailFromIdent("no brackets"); err == nil {
			t.Fatal("output without angle brackets must be a resolution failure")
		}
		if _, err := emailFromIdent("a <b"); err == nil {
			t.Fatal("unclosed bracket must be a resolution failure")
		}
	})

	t.Run("fail_open_audit_write_failures_stay_silent", func(t *testing.T) {
		// A file used as a directory: MkdirAll fails, the guard still allows
		// (the log write never blocks the decision) and only notes stderr.
		notADir := filepath.Join(t.TempDir(), "blocker")
		if err := os.WriteFile(notADir, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		appendCommitIdentityFailOpen(ciInput("git commit -m x"), notADir, "cmd", "/d", "cause")
	})

	t.Run("shellFields_escape_forms", func(t *testing.T) {
		if got := shellFields(`"a\"b" c`); len(got) != 2 || got[0] != `"a\"b"` {
			t.Fatalf("double-quote escape = %q", got)
		}
		if got := shellFields("trail\\"); len(got) != 1 || got[0] != "trail\\" {
			t.Fatalf("trailing backslash = %q", got)
		}
	})

	t.Run("gitVarIdent_fails_on_missing_dir", func(t *testing.T) {
		if _, err := gitVarIdent(filepath.Join(t.TempDir(), "missing"), []string{"PATH=" + os.Getenv("PATH")}, "GIT_AUTHOR_IDENT"); err == nil {
			t.Fatal("a probe against a missing directory must fail (fail-open input)")
		}
	})

	t.Run("applySegmentFacts_attached_forms", func(t *testing.T) {
		facts := identityFacts{}
		applySegmentFacts(`git --author "T <a@b.c>" commit -m x`, &facts)
		if facts.authorEmail != "a@b.c" {
			t.Fatalf("space-separated --author = %q", facts.authorEmail)
		}
		facts = identityFacts{}
		applySegmentFacts("git -C/tmp/r -cuser.email=d@e.f -c core.bare=true commit", &facts)
		if facts.targetDir != "/tmp/r" {
			t.Fatalf("attached -C = %q", facts.targetDir)
		}
		if facts.configEmail != "d@e.f" {
			t.Fatalf("attached -c user.email = %q", facts.configEmail)
		}
	})
}
