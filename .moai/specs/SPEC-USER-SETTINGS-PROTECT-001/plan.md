# SPEC-USER-SETTINGS-PROTECT-001 — Implementation Plan

Scope of this document: plan phase only. It describes what the run phase must do and records what was observed while planning. It makes no claim that any change has been implemented. Every factual claim about current code carries its command and verbatim output in §A.2. Every unobserved item is listed in §A.4. Excerpts are marked where a line was truncated for width.

## §A Context and Evidence

Tree and measurement identity: worktree HEAD `2aab5f797` (short), branch `WT-3-2-0`, measured 2026-10-10 between 06:51 and 06:55 UTC. Go toolchain go1.26.8 darwin/arm64.

### §A.1 Claim

- C-1: `moai init` rewrites the USER-scope settings `permissions` object and removes the user's `allow` and `deny` lists on the default (semi-auto) path. Basis: E-3, E-4, E-5, E-6.
- C-2: The root cause is that the USER-scope writer builds a permissions block without the three lists, and the splice step takes only the unmodelled keys from the existing file. Basis: E-7.
- C-3: The live review-gate Codex test sets neither CODEX_HOME nor HOME. The live audit fixture sets CODEX_HOME. Basis: E-10.
- C-4: 36 of the 38 `TestMain(m *testing.M)` entry points redirect neither MOAI_HOME nor HOME. Basis: E-8.
- C-5: Five packages with a TestMain (homestate, escalation, factory, factorymsg, web) and one package without one (contract/receipt) reach home-resolving functions without a MOAI_HOME sandbox. Basis: E-9.
- C-6: `clean --home` scans five categories (projects, debug, releases, logs, backups) and neither run nor db. Basis: E-11.
- C-7: The settings template ships no `permissions.defaultMode`, deliberately. Basis: E-12, E-13.
- C-8: Three ordering commits named by the dispatch are not ancestors of the tree. Basis: E-14.
- C-9: The decision gate is on, so a decision index is required. Basis: E-17.
- C-10: The S-1 behaviour contradicts a committed contract. SPEC-INIT-WIZARD-REPAIR-001 §4 and its v0.1.1 HISTORY row require the distributed-default write to change only `permissions.defaultMode` and to preserve allow, deny, and ask verbatim. SPEC-AUT-PERMMODES-001 (REQ-007, re-scoped) requires exactly one JSON key in exactly one file. Basis: E-23, and the probes in E-3 and E-4.

### §A.2 Evidence

E-1 — worktree identity.

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630
$ git branch --show-current
WT-3-2-0
```

E-2 — SPEC ID self-check and uniqueness.

```
$ ID="SPEC-USER-SETTINGS-PROTECT-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
$ ls -d .moai/specs/SPEC-USER-SETTINGS-PROTECT-001 2>&1
ls: .moai/specs/SPEC-USER-SETTINGS-PROTECT-001: No such file or directory
```

E-3 — probe 1, the USER-scope writer `toolpolicy.WriteUserDefaultMode`. Re-run at tree 3975fe3cc as `go -C <worktree> test -overlay .moai/reports/t1630/evidence/overlay-e3.json -count=1 -v -run '^TestProbeUserScopeWriteUserDefaultModeKeepsAllow$' ./internal/config/toolpolicy/`. Inputs: `.moai/reports/t1630/evidence/overlay-e3.json` maps the package path `zz_e3_probe_test.go` to `.moai/reports/t1630/evidence/probe-e3_test.go.txt`. Exit 0; the logs match the original observation. The probe asserts nothing; the RED is the AFTER body (acceptance.md AC-001). The probe file is injected through `-overlay`; the repository was not written (Appendix A-1).

```
=== RUN   TestProbeUserScopeWriteUserDefaultModeKeepsAllow
    zz_probe_test.go:24: PROBE BEFORE:
        {
          "permissions": {
            "defaultMode": "plan",
            "allow": ["Bash(make:*)"],
            "ask": ["Bash(rm:*)"],
            "deny": ["Read(./secret)"],
            "additionalDirectories": ["/tmp/x"]
          },
          "env": {"FOO": "1"}
        }
    zz_probe_test.go:25: PROBE AFTER:
        {
          "permissions": {
            "defaultMode": "acceptEdits",
            "additionalDirectories": ["/tmp/x"]
          },
          "env": {"FOO": "1"}
        }
--- PASS: TestProbeUserScopeWriteUserDefaultModeKeepsAllow (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.283s
```

The probe passes because it only logs. The red is the AFTER body, which has no `allow`, `ask`, or `deny`. A criterion asserting that those lists survive fails on this tree.

E-4 — probe 2, the function `init` calls on the default tier path: `project.ApplyAutonomyTierBundle` with an empty persisted tier. Re-run at tree 3975fe3cc as `go -C <worktree> test -overlay .moai/reports/t1630/evidence/overlay-e4.json -count=1 -v -run '^TestProbeInitDefaultTierKeepsUserAllowList$' ./internal/core/project/`. Inputs: `.moai/reports/t1630/evidence/overlay-e4.json` maps `zz_e4_probe_test.go` to `.moai/reports/t1630/evidence/probe-e4_test.go.txt`. Exit 0; the logs match the original observation (Appendix A-2).

```
=== RUN   TestProbeInitDefaultTierKeepsUserAllowList
    zz_probe2_test.go:30: PROBE2 BEFORE:
        {
          "permissions": {
            "allow": ["Bash(git status:*)"],
            "deny": ["Read(./.env)"]
          },
          "env": {"A": "1"}
        }
    zz_probe2_test.go:31: PROBE2 AFTER:
        {
          "permissions": {
            "defaultMode": "acceptEdits"
          },
          "env": {"A": "1"}
        }
--- PASS: TestProbeInitDefaultTierKeepsUserAllowList (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/core/project	0.370s
```

E-5 — the init call chain.

```
$ grep -n -E 'applyAutonomyTierBundleFn|ensureUserAssetsLocked' internal/cli/init.go
946:		if err := ensureUserAssetsLocked(homeDir, selection, cmd.OutOrStdout()); err != nil {
971:		if tierErr := applyAutonomyTierBundleFn(
$ grep -rn -E 'applyAutonomyTierBundleFn\s*=' internal/cli --include='*.go' | head -5
internal/cli/update_settings_snapshot.go:32:var applyAutonomyTierBundleFn = project.ApplyAutonomyTierBundle
$ sed -n '962,973p' internal/cli/init.go
	if homeDir, homeErr := userHomeDirFn(); homeErr == nil {
		// SPEC-INIT-HARNESS-001 (REQ-IH-005): a codex-only project carries no
		// .claude/ surface, so the bundle gets an empty projectSettingsPath —
		// its contract is USER-scope-only writes in that case, never a
		// project-root .claude/settings.json.
		projectSettingsPath := filepath.Join(opts.ProjectRoot, ".claude", "settings.json")
		if agentWiringSelection == agentWiringGPT {
			projectSettingsPath = ""
		}
		if tierErr := applyAutonomyTierBundleFn(
			opts.ProjectRoot,
			filepath.Join(homeDir, ".claude", "settings.json"),
```

The bundle call sits under `if homeDir, homeErr := userHomeDirFn(); homeErr == nil {` and has no condition on the autonomy flag. The USER-scope path is `<home>/.claude/settings.json`, built from the home directory and not from CLAUDE_CONFIG_DIR.

E-6 — the empty tier resolves to semi-auto, and the semi-auto branch writes only the USER-scope defaultMode.

```
$ sed -n '60,70p' internal/config/autonomy_tiers.go
// they are NOT re-validated here (the selector validated at write time, and the
// env-key wins per STOPCHAIN-TRIM's canonical-source rule).
func ResolveEffectiveTier(persistedTier string) string {
	normalized := strings.ToLower(strings.TrimSpace(persistedTier))
	if normalized == "" {
		return AutonomyTierSemiAuto
	}
	return normalized
}

// TierDefaultMode maps an autonomy tier to its Claude Code permissions
$ sed -n '72,78p' internal/core/project/autonomy_bundle.go
	effective := config.ResolveEffectiveTier(persistedTier)
	if effective == config.AutonomyTierSemiAuto {
		// REQ-004 (re-scoped): bounded delta — the USER-scope acceptEdits
		// record is the ONLY sanctioned write. The gates never bind semi-auto
		// (EffectiveTierWithGates passes lower tiers through) and the
		// PROJECT-scope deny/ask are tier-invariant, so neither is touched.
		if err := toolpolicy.WriteUserDefaultMode(userSettingsPath, config.TierDefaultMode(effective)); err != nil {
```

E-7 — the root cause.

```
$ sed -n '106,113p' internal/config/toolpolicy/tier_render.go
func WriteUserDefaultMode(userPath, defaultMode string) error {
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		return fmt.Errorf("create user settings dir: %w", err)
	}
	block := &PermissionsBlock{
		DefaultMode: defaultMode,
		Raw:         map[string]json.RawMessage{},
	}
$ sed -n '138,142p' internal/config/toolpolicy/tier_render.go
	// Preserve any extra Raw keys from the existing block (additionalDirectories
	// etc.) so the codegen does not drop non-list permission settings.
	existing, extrErr := extractPermissions(body)
	if extrErr == nil && existing != nil {
		block.Raw = existing.Raw
```

The settings-region renderer (Read tool, `internal/config/toolpolicy/settings_region.go` lines 198-209) writes each list only when it is non-empty:

```
	if block.DefaultMode != "" {
		writeKey("defaultMode", quoteJSON(block.DefaultMode))
	}
	if len(block.Allow) > 0 {
		writeKey("allow", renderStringList(block.Allow))
	}
	if len(block.Ask) > 0 {
		writeKey("ask", renderStringList(block.Ask))
	}
	if len(block.Deny) > 0 {
		writeKey("deny", renderStringList(block.Deny))
	}
```

`renderIntoFile` restores only the `Raw` extras from the existing file (tier_render.go:138-142). `Allow`, `Ask`, and `Deny` come from the block. For the USER scope the block has none, so the lists are dropped. The doc comment on `WriteUserDefaultMode` says it preserves "allow, deny, ask" (tier_render.go:93-94). The code does not.

E-8 — TestMain redirection counts and the cli and hook sandboxes.

```
$ grep -rl -E 'func TestMain\(m \*testing\.M\)' internal cmd --include='*_test.go' | wc -l
      38
$ grep -rL -E 'EnvHome|MOAI_HOME|Setenv\("HOME"' $(grep -rl -E 'func TestMain\(m \*testing\.M\)' internal cmd --include='*_test.go') | wc -l
      36
$ grep -n -E 'homeSandboxEnv|codexHomeEnvVar|config\.EnvHome|moaiHomeSandboxEnv' internal/cli/main_test.go | head -8
222:// homeSandboxEnv and realHomeEnv carry the sandbox path and the captured real
228:	homeSandboxEnv = "MOAI_CLI_TEST_HOME_SANDBOX"
283:	if inherited := os.Getenv(homeSandboxEnv); inherited != "" {
314:	origCodexHome, hadCodexHome := os.LookupEnv(codexHomeEnvVar)
315:	_ = os.Unsetenv(codexHomeEnvVar)
316:	_ = os.Setenv(homeSandboxEnv, dir)
322:			_ = os.Setenv(codexHomeEnvVar, origCodexHome)
324:		_ = os.Unsetenv(homeSandboxEnv)
$ grep -n '"HOME"' internal/cli/main_test.go
720:		t.Setenv("HOME", tmp)
$ grep -n -E 'EnvHome|"HOME"' internal/hook/main_test.go
101:		_ = os.Setenv(config.EnvHome, dir)
110:	origHome, origUserProfile := os.Getenv("HOME"), os.Getenv("USERPROFILE")
111:	_ = os.Setenv("HOME", sandboxRoot)
127:			_ = os.Setenv("HOME", origHome)
129:			_ = os.Unsetenv("HOME")
```

The cli package sandboxes MOAI_HOME (its TestMain calls `sandboxMoaiHome`, which sets `config.EnvHome`) and unsets CODEX_HOME for the whole process inside `sandboxUserHomeDir` (line 315). It does not set the process HOME variable. Its only HOME assignment is the `t.Setenv` inside one test (line 720). The hook package sets both MOAI_HOME (line 101) and HOME (line 111) for its TestMain sandbox.

E-9 — reach of home-resolving functions in test code (name-based measure; see G-9).

```
$ grep -rl -E 'EnsureProjectLayout|EnsureHomeLayout|RunProjectDir|paths\.MoaiHome|StoreDir\(|SearchDBPath|FactoryDBPath' internal cmd --include='*_test.go' | sed 's#/[^/]*$##' | sort -u
internal/cli
internal/contract/receipt
internal/escalation
internal/factory
internal/factorymsg
internal/homestate
internal/hook
internal/web
$ grep -L -E 'EnvHome|MOAI_HOME|sandbox|Setenv\("HOME"' internal/homestate/main_test.go internal/escalation/main_test.go internal/factory/main_test.go internal/factorymsg/main_test.go internal/web/main_test.go internal/cli/main_test.go internal/hook/main_test.go
internal/homestate/main_test.go
internal/escalation/main_test.go
internal/factorymsg/main_test.go
internal/factory/main_test.go
$ ls internal/contract/receipt/main_test.go
ls: internal/contract/receipt/main_test.go: No such file or directory
$ grep -n -E 'EnvHome|MOAI_HOME|sandbox|Setenv\("HOME"|UserHomeDir' internal/web/main_test.go | head -12
29:// sandboxProfileBaseDir points profile.GetBaseDir at a throwaway directory for
35:func sandboxProfileBaseDir() func() {
56:	restore := sandboxProfileBaseDir()
62:// TestProfileBaseDirIsSandboxed is the guard for sandboxProfileBaseDir. It fails
74:			"sandboxProfileBaseDir() before m.Run(). Without it, any test " +
79:	home, err := os.UserHomeDir()
81:		t.Skip("cannot determine home directory; sandbox comparison unavailable")
$ grep -L -E 'EnvHome|MOAI_HOME' internal/homestate/*_test.go | wc -l
      47
$ ls internal/homestate/*_test.go | wc -l
      57
```

The `grep -L` with the sandbox token matches the four files listed. The web package is in the reach set but carries only a profile-base-directory sandbox, which is why the MOAI_HOME-token grep for `web` is listed separately above (no `EnvHome` or `MOAI_HOME` line in its output). Of the 57 `internal/homestate` test files, 47 contain neither `EnvHome` nor `MOAI_HOME`.

E-10 — Codex live tests.

```
$ grep -n -E 'Setenv|exec\.Command|CommandContext|HOME|CODEX|TempDir|Skip' internal/cli/codex_review_gate_live_test.go
27:// Skip conditions (CI without codex still passes):
30://   - MOAI_SKIP_LIVE_CODEX set (manual opt-out)
36:	if os.Getenv("MOAI_SKIP_LIVE_CODEX") == "1" {
37:		t.Skip("MOAI_SKIP_LIVE_CODEX=1")
41:		t.Skipf("codex binary not on PATH: %v", err)
46:	if ver, vErr := exec.Command(bin, "--version").Output(); vErr != nil || !strings.Contains(string(ver), "codex") {
47:		t.Skipf("codex --version non-functional (ver=%q err=%v) — environment lacks a working codex", strings.TrimSpace(string(ver)), vErr)
57:	repo := t.TempDir()
60:		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
138:		t.Skipf("codex review turn did not complete — the producer recorded inconclusive with the error surfaced (correct behavior)")
158:		"func runQuery(q string) { exec.Command(\"sh\", \"-c\", q).Run() }\n\n" +
$ grep -n 'CODEX_HOME' internal/cli/codex_review_gate_live_test.go; echo "codex-home-matches-exit=$?"
codex-home-matches-exit=1
$ grep -n -E 'func TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey' internal/cli/codex_review_gate_live_test.go
35:func TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey(t *testing.T) {
$ grep -n -E 'envCodexRoleLive|t\.Setenv\(config\.EnvHome|t\.Setenv\(codexHomeEnvVar' internal/cli/codex_audit_live_test.go | head -6
53:	if os.Getenv(envCodexRoleLive) != "1" {
54:		t.Skip("NOT_RUN " + envCodexRoleLive + " is not 1: the live Codex audit run is gated off")
72:	t.Setenv(config.EnvHome, t.TempDir())
78:	t.Setenv(codexHomeEnvVar, f.codexHome)
```

The review-gate test starts the codex binary (`codex --version` at line 46, then a review turn) and its file contains no CODEX_HOME, HOME, or Setenv call. Whether that turn writes under the operator's real `~/.codex` was not executed (G-10). The audit fixture sets MOAI_HOME (line 72) and CODEX_HOME (line 78) to temporary directories, which is the pattern REQ-009 asks the review-gate test to follow.

E-11 — `clean --home` categories and carve-outs.

```
$ grep -n -E 'add\(|Category|scanReleaseCandidates|func runCleanHome' internal/cli/clean_home.go
128:	Category string // debug | releases | logs | backups
182:			Category: category,
187:	// Category 0 — per-profile projects/ entries. Age-based expiry is 180
236:					if add(entry.abs, "projects", entry.size) {
249:				if add(entry.abs, "projects", entry.size) {
257:	// Category 1 — per-profile debug/ entries older than retention.
286:				add(abs, "debug", size)
291:	// Category 2 — releases/ beyond current + keep newest.
293:	candidates = append(candidates, scanReleaseCandidates(root, releasesDir, releaseKeep, currentVersion)...)
295:	// Category 3 — root logs/ files older than retention.
306:			add(filepath.Join(logsDir, e.Name()), "logs", info.Size())
310:	// Category 4 — backups/removed-* directories older than retention.
329:			add(abs, "backups", size)
404:// scanReleaseCandidates returns the deletable release binaries (plus their
408:func scanReleaseCandidates(root, releasesDir string, releaseKeep int, currentVersion string) []homeCleanCandidate {
458:			Category: "releases",
530:func runCleanHome(p printer.Printer, force bool) error {
570:				p.Info("Deleted [%s] %s (%s)", c.Category, c.RelPath, formatDiskBytes(c.Size))
573:			p.Info("[dry-run] Would delete [%s] %s (%s)", c.Category, c.RelPath, formatDiskBytes(c.Size))
$ grep -n -A10 'var carveOutDirNames' internal/cli/clean_home.go
32:var carveOutDirNames = map[string]bool{
33-	"projects":  true,
34-	"config":    true,
35-	"state":     true,
36-	"worktrees": true,
37-	"mcp":       true,
38-	"bin":       true,
39-	"search":    true,
40-	"studio":    true,
41-	"plugins":   true,
42-}
```

The categories in the scan are projects (lines 236, 249), debug (286), releases (293, 458), logs (306), and backups (329). Neither `run` nor `db` appears in the category set or in `carveOutDirNames`. Both directories are created by `EnsureHomeLayout` (`internal/homestate/paths.go`, directory list entries for `db` and `run`) and are never candidates.

E-12 — template.

```
$ grep -n -E 'permissions|defaultMode|"allow"|"ask"|"deny"' internal/template/templates/.claude/settings.json.tmpl | head -12
436:  "permissions": {
437:    "allow": [
557:    "deny": [
$ grep -c '"defaultMode"' internal/template/templates/.claude/settings.json.tmpl; echo "template-defaultMode-count-exit=$?"
0
template-defaultMode-count-exit=1
```

E-13 — the template default was removed on purpose; two comments say so.

```
$ grep -n 'stopped shipping a defaultMode default' internal/cli/launcher.go
738:		// The template settings.json stopped shipping a defaultMode default
1121:// the template settings.json stopped shipping a defaultMode default
```

These comments state the decision without citing a SPEC. Decided Q2 (decision-index.md) settles the disposition.

E-14 — ordering commits named by the dispatch.

```
$ git log --oneline --all -i --grep=t1619 -n 3
8108eb256 docs(SPEC-UPDATE-REPAIR-RC30-001): U-01 blocked-on token, count note, stale Q2 wording, retention text follows the implemented rule (card t1619)
569a3fe5f test(update): RED for profile-mapped projects keeping shared copies (D1, U-01 (c), card t1619)
d16765672 docs(SPEC-UPDATE-REPAIR-RC30-001): record operator decision Q2 as option (c) with two conditions (d-20261010T060848Z-dc3b, card t1619)
$ git log --oneline --all -i --grep=t1578 -n 3
cbb6c7940 Merge branch 'WT-user-asset-bundle' into WT-moai-update-repair (card t1619)
e73a7cbf5 docs(SPEC-UPDATE-MIGRATION-FIX-001): sync-phase artifacts and 3-phase close (card t1578)
32035f28b docs(SPEC-UPDATE-MIGRATION-FIX-001): state the applicable members in the DoD read-surface wording (card t1578)
$ git log --oneline --all -i --grep=t1591 -n 3
a372a984c docs(SPEC-USERASSET-DEPLOY-GUARD-001): backfill the sync commit sha (card t1591)
19a7380f2 docs(SPEC-USERASSET-DEPLOY-GUARD-001): align the round-3 gap wording with decision d-20261010T060606Z-db59 (card t1591)
543baddc9 docs(SPEC-USERASSET-DEPLOY-GUARD-001): sync-phase artifacts for the round-3 repair (card t1591)
$ git merge-base --is-ancestor 8108eb256 HEAD; echo "t1619 ancestor-of-HEAD exit=$?"
t1619 ancestor-of-HEAD exit=1
$ git merge-base --is-ancestor e73a7cbf5 HEAD; echo "t1578 ancestor-of-HEAD exit=$?"
t1578 ancestor-of-HEAD exit=1
$ git merge-base --is-ancestor a372a984c HEAD; echo "t1591 ancestor-of-HEAD exit=$?"
t1591 ancestor-of-HEAD exit=1
```

E-15 — decision-record IDs named in the dispatch (1e9b, 3db2, d37b, 2b7b, ca0c).

```
$ grep -rn -w -E '1e9b|3db2|d37b|2b7b|ca0c' .moai --include='*.md' -c | grep -v ':0$'
.moai/reports/t1630/progress.md:2
```

The only file matching the word-bounded IDs is the lane's own progress record. An unbounded search matches substrings of unrelated SHAs (for example `01872b7bc...`), which is why the bounded form is used.

E-16 — the install-design reference.

```
$ grep -rn -E '설치 설계|install design|설치설계' .moai .claude --include='*.md' -l
.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/spec.md
.moai/reports/t1630/progress.md
$ sed -n '186p' .moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/spec.md
**REQ-VVCR-019** (Unwanted): The reinstall phase **shall not** reinstall design-domain assets (REQ-VVCR-011 removal targets); design assets remain absent from the v3 baseline and are opt-in via a future separate flow.
```

The clean-reinstall match is a false positive: the phrase "install design" occurs inside "reinstall design". The other match is the lane's own note that the reference was not located. No committed document of this tree contains the reference.

E-17 — decision gate.

```
$ grep -rn -E 'decision_gate|recommendation_mode' .moai/config/sections/ | head -6
.moai/config/sections/interview.yaml:6:  decision_gate: on
.moai/config/sections/interview.yaml:9:  recommendation_mode: pull
```

E-18 — covering-SPEC status and tier, read from frontmatter.

```
$ grep -n -E '^status:|^tier:' .moai/specs/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001/spec.md .moai/specs/SPEC-SETTINGS-ORIGIN-001/spec.md .moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/spec.md .moai/specs/SPEC-USER-ASSET-INSTALL-001/spec.md
.moai/specs/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001/spec.md:5:status: implemented
.moai/specs/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001/spec.md:14:tier: M
.moai/specs/SPEC-SETTINGS-ORIGIN-001/spec.md:5:status: completed
.moai/specs/SPEC-SETTINGS-ORIGIN-001/spec.md:14:tier: S
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/spec.md:5:status: completed
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/spec.md:14:tier: M
.moai/specs/SPEC-USER-ASSET-INSTALL-001/spec.md:5:status: completed
.moai/specs/SPEC-USER-ASSET-INSTALL-001/spec.md:14:tier: L
```

E-19 — close markers in the covering-SPEC progress records (lines truncated at 200 characters for width).

```
$ grep -n -E 'sync_commit_sha|plan_status|^## §E' .moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md .moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md .moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md | cut -c1-200
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:3:## §E.1 Plan-phase Audit-Ready Signal
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:5:plan_status: audit-ready
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:10:## §E.2 Run-phase Evidence
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:32:## §E.3 Run-phase Audit-Ready Signal
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:44:## §E.4 Sync-phase Audit-Ready Signal
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:47:sync_commit_sha: "c7da0d522"
.moai/specs/SPEC-SETTINGS-ORIGIN-001/progress.md:53:**Sync summary** — Sync scope: spec.md frontmatter transition only (`status: in-progress → implemented → completed` merged close + `updated: 2026-09
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md:3:## §E.1 Plan-phase Audit-Ready Signal
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md:5:plan_status: audit-ready
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md:131:## §E.2 Run-phase Evidence
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md:276:## §E.3 Run-phase Audit-Ready Signal
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md:299:## §E.4 Sync-phase Audit-Ready Signal
.moai/specs/SPEC-PREMERGE-SETTINGS-DRIFT-001/progress.md:303:sync_commit_sha: 955e5190c86c88605efcaaf41bd002005516ed51   # re-close commit. Superseded prior value: 199d2777be085033a96d89cc45466d11ef4a
.moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md:11:## §E.1 Plan-phase Audit-Ready Signal
.moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md:13:- plan_status: audit-ready
.moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md:297:## §E.2 Run-phase Evidence
.moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md:660:## §E.3 Run-phase Audit-Ready Signal
.moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md:699:## §E.4 Sync-phase Audit-Ready Signal
.moai/specs/SPEC-USER-ASSET-INSTALL-001/progress.md:702:- sync_commit_sha: 33127926c178499bd972b433e31e13e0e57f6e38
```

The update-namespace progress record is not in this grep. A separate count (`grep -c -E '^## |sync_commit_sha|plan_status'`) returns 12, all `##` headings, and no line of its own matches the close-marker pattern.

E-20 — live check of the update-namespace mechanism.

```
$ grep -rln 'my-harness' internal --include='*.go'
internal/template/namespace_protection_audit_test.go
internal/cli/update_namespace_harness_v2_test.go
internal/cli/update_preserve_inventory.go
internal/cli/doctor_harness.go
internal/cli/update_namespace_hns_test.go
internal/cli/doctor_skills.go
internal/cli/update/plan/plan_test.go
internal/cli/update/plan/plan.go
internal/harness/types.go
internal/harness/chaining_rules.go
internal/harness/hns_prefix_test.go
internal/harness/layer2.go
internal/harness/layer1.go
internal/harness/prefix_conflict.go
internal/harness/frozen_guard.go
$ grep -n 'my-harness-' internal/cli/update.go
(no output; exit status 1)
```

E-21 — existing test names cited by the acceptance criteria (each checked with a repository grep).

```
$ grep -rn -E 'func Test[A-Za-z0-9_]*CleanHome[A-Za-z0-9_]*' internal/cli --include='*_test.go' | head -12
internal/cli/clean_home_test.go:100:func TestCleanHome_DryRunMutatesNothing(t *testing.T) {
internal/cli/clean_home_test.go:126:func TestCleanHome_ForceDeletesOnlyAllowlistedCategories(t *testing.T) {
internal/cli/clean_home_test.go:173:func TestCleanHome_RetentionFromHomeTier(t *testing.T) {
internal/cli/clean_home_test.go:374:func TestCleanHome_NoHomeIsNoop(t *testing.T) {
internal/cli/clean_home_test.go:388:func TestCleanHome_HomeFlagWiring(t *testing.T) {
internal/cli/clean_home_carveout_test.go:30:func TestCleanHomeCarveOut_PathPredicate(t *testing.T) {
internal/cli/clean_home_carveout_test.go:82:func TestCleanHomeCarveOut_ForcePreservesCarvedSegments(t *testing.T) {
internal/cli/clean_home_carveout_test.go:154:func TestCleanHomeCarveOut_ReleasesKeepCurrentPlusNewest(t *testing.T) {
internal/cli/clean_home_carveout_test.go:200:func TestCleanHomeCarveOut_MOAIHomeRedirect(t *testing.T) {
internal/cli/profile_lease_integration_test.go:42:func TestCleanHomeReportsProtectedProfileReasonInDryRunAndForce(t *testing.T) {
internal/cli/profile_lease_integration_test.go:83:func TestCleanHomeSkipsLiveAndIndeterminateProfiles(t *testing.T) {
$ grep -rn -E 'func TestStoreDirUsesQueueProjectKey|func TestStoreDirMatchesEscalation' internal --include='*_test.go'
internal/escalation/store_test.go:35:func TestStoreDirUsesQueueProjectKey(t *testing.T) {
internal/contract/receipt/dir_test.go:14:func TestStoreDirMatchesEscalation(t *testing.T) {
$ grep -rn 'func TestMainSandboxesProfileLeaseEnv' internal/cli
internal/cli/moai_home_sandbox_test.go:13:func TestMainSandboxesProfileLeaseEnv(t *testing.T) {
```

acceptance.md names only these existing functions, plus `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` (E-10). New test names are not invented here; run-phase names are recorded in progress.md when the tests are authored.

E-22 — the worktree was not written by the probes.

```
$ git rev-parse --short HEAD
2aab5f797
$ git status --short
(no output)
```

Both probes ran through `-overlay`, so the repository was not written. The probe sources are in the session scratch directory and are reproduced in Appendix A.

E-23 — the committed contract for the USER-scope write (completed SPECs, tracked files).

```
$ sed -n '21p' .moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md | cut -c1-1500
- 2026-08-22 — v0.1.1 audit revision round (iteration-1 FAIL → fix, lead ruling): both plan.md §B markers RESOLVED — wire both — with the conditions pinned in SPEC text (§4 key-scoped USER-write constraint + REQ-003 splice clause; TTY gate + default-preservation binding for the update-wizard step); SPEC-WT-DOC-001 archive reconciliation added to §6.
$ sed -n '96p' .moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md | cut -c1-1200
- **USER-scope write is key-scoped (lead ruling 2026-08-22)**: `~/.claude/settings.json` is user-owned territory. Chain ①'s USER-scope write is a read-modify-write splice limited to the `permissions` block — distributed-default path: exactly the `permissions.defaultMode` key via `toolpolicy.WriteUserDefaultMode`, which preserves every other region (PATH, hooks, env, allow, deny, ask) verbatim. (The `RenderTierPermissions` full-bundle path, reachable only when the initialized project already ships a tool-policy.yaml — the distributed template does not — regenerates deny/ask within the same block per SPEC-AUTONOMY-TIERS-001 REQ-003, still as a region splice.) Whole-file overwrite is prohibited and MUST be asserted by an M1 preservation test.
$ sed -n '83p' .moai/specs/SPEC-AUT-PERMMODES-001/spec.md | cut -c1-900
REQ-007 of SPEC-AUTONOMY-TIERS-001 ("unset / semi-auto → zero behavior delta, no file written") is RE-SCOPED: the unset / `semi-auto` selection SHALL write ONLY the USER-scope `defaultMode: "acceptEdits"` record and NOTHING else — the PROJECT-scope `allow`/`ask`/`deny` arrays, the deployed template files, and every other file MUST remain byte-identical. The sanctioned delta is exactly one JSON key in exactly one file. Deny/ask tier-invariance (REQ-004 of the owning SPEC) is unchanged.
$ git ls-files .moai/specs/SPEC-AUT-PERMMODES-001/spec.md .moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md
.moai/specs/SPEC-AUT-PERMMODES-001/spec.md
.moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md
$ grep -n -i 'M1 preservation\|preservation test' .moai/specs/SPEC-INIT-WIZARD-REPAIR-001/spec.md | cut -c1-260 | head -5
96:- **USER-scope write is key-scoped (lead ruling 2026-08-22)**: `~/.claude/settings.json` is user-owned territory. Chain ①'s USER-scope write is a read-modify-write splice limited to the `permissions` block — distributed-default path: exactly the `permissions.defaultMode` key via `toolpolicy.WriteUserDefaultMode`, which preserves every other region (PATH, hooks, env, allow, deny, ask) verbatim. (The `RenderTierPermissions` full-bundle path, reachable only when the initialized project already ships a tool-policy.yaml — the distributed template does not — regenerates deny/ask within the same block per SPEC-AUTONOMY-TIERS-001 REQ-003, still as a region splice.) Whole-file overwrite is prohibited and MUST be asserted by an M1 preservation test.
$ grep -rn -i -E 'preserv' internal/config/toolpolicy internal/core/project internal/cli --include='*_test.go' | grep -i -E 'allow|deny|\bask\b|WriteUserDefaultMode|M1' | cut -c1-220 | head -10; echo "preservation-grep-exit=$?"
internal/cli/codex_review_gate_wtnobase_test.go:3:// SPEC-CODEX-REVIEW-OWNERSHIP-001 M1 — the gate-level preservation line for a
internal/cli/launcher_test.go:652:					t.Error("permissions.allow should be preserved")
internal/cli/codex_event_adaptation_test.go:153:// TestCodexPermissionRequestDenyPreserved is the AC-HPR-010 golden leg: the
internal/cli/codex_event_adaptation_test.go:158:func TestCodexPermissionRequestDenyPreserved(t *testing.T) {
internal/cli/update_hygiene_characterization_test.go:3:// This file is the M1 PRESERVE safety net for SPEC-CLIFIX-HYGIENE-001.
internal/cli/fang_characterization_test.go:16:// SPEC-CLI-TUX-V3-001 M1c must preserve across the charm.land/fang/v2 swap
internal/cli/update_deny_migration_test.go:108:	// Surviving v3 entries + user-custom entry preserved, in order.
internal/cli/update_deny_migration_test.go:131:		t.Errorf("outputStyle not preserved: %v", m["outputStyle"])
internal/cli/update_deny_migration_test.go:138:		t.Errorf("env.MOAI_CONFIG_SOURCE not preserved: %v", env)
internal/cli/update_deny_migration_test.go:143:		t.Errorf("permissions.allow not preserved: %v", allow)
preservation-grep-exit=0
```

Reading E-23: the pinned condition (line 96) names the USER-scope write and the lists the probe shows removed. The M1 preservation test that the condition requires is not in `internal/config/toolpolicy` or `internal/core/project` (the grep output above has no matching test in those packages). The `permissions.allow` assertions at `internal/cli/launcher_test.go:652` and `internal/cli/update_deny_migration_test.go:143` test other paths (the launcher and the deny migration), not the init writer, and were not analysed further.

### §A.3 Baseline-attribution

- Tree: `2aab5f797` on `WT-3-2-0`, the same tree for every item in §A.2.
- Toolchain: go1.26.8 darwin/arm64, the locally installed Go toolchain. Probe commands ran from the worktree through `go -C`.
- Tool provenance (verification-claim-integrity §2.2): no moai CLI measurement is cited. The moai MCP server reported build v3.2.0-rc.29 (commit 4f8aba061). The server produced no evidence here, so §2.2 does not bind this plan.
- Attribution limit: the probes exercise the writer functions in isolation and do not run the full init executor (R-1).

### §A.4 Gaps (not observed; not claimed)

- G-1 — The t1567 card text is not in the tree. Its scope is taken from the dispatch text ("permissions.defaultMode template"). Decided Q2 records the disposition; the card text itself remains absent from the tree.
- G-2 — The t1594 card text is not in the tree, so its residue inventory is unknown. Scope item S-7 is blocked (§B B4).
- G-3 — The "3.2 disposition table" named in the dispatch is not in the worktree. The absorbed-card scope rests on the dispatch.
- G-4 — The install-design reference "설치 설계 §9" is not in any committed document of the tree (E-16). Nothing from it is cited.
- G-5 — Operator decision records 1e9b, 3db2, d37b, and 2b7b are not on disk (E-15). Where the dispatch attributes a statement to them, that statement is carried as dispatch text. No decision row relies on them.
- G-6 — The landing-order decision ca0c is not located in the worktree; location not observed. It is not cited as authority. REQ-013 rests on the dispatch text and the commits in E-14.
- G-7 — Whether the init writer produces the dirty `.claude/settings.json` shape recorded in SPEC-SETTINGS-ORIGIN-001 §1 was not measured. The two are not linked by this plan.
- G-8 — The entry count under the operator's `~/.moai/run` was not measured. Reading the operator's home is outside the worktree boundary. It is a run-phase measurement (AC-007).
- G-9 — The reach measure (E-9) is name-based: it finds direct references to home-resolving names in test files. Transitive reach through production helpers was not measured, so the file estimate in §A.4 Basis may be low.
- G-10 — The live review-gate test was not executed. Running it would start `codex` against the operator's real home, which this plan must not do. Its effect on `~/.codex` is unobserved.
- G-11 — `moai init --force` on an existing project's `.claude/settings.json` was not measured (Q6). Reading the deployer shows a force-update mode that overwrites existing files without a manifest check (`internal/template/deployer.go:80` and `:254`), but that is the update path, not a measured init run.
- G-12 — The PROJECT-scope regeneration under a tool-policy document (`internal/config/toolpolicy/tier_render.go:76-81`, which sets `Allow: full.Allow`) was read in code but not probed. Decided Q5 records it.
- G-13 — The landing SHAs for t1619, t1578, and t1591 on the run base are not known. E-14 names representative commits. The leader names the landing SHAs at dispatch.
- G-14 — Resolved in the repair round: the RED-now cells are observed on 3975fe3cc (acceptance.md), with the probes in `.moai/reports/t1630/evidence/`. The run-phase re-observation at P-3 and M2 remains owed, and AC-011 clause (iii) is a regression-guard with no RED cell at plan time (G-21).
- G-15 — Whether a `CLAUDE_CONFIG_DIR` profile session reads a different settings file from `<home>/.claude/settings.json` was not measured (Q4).
- G-16 — The dispatch lists the decision IDs and the landing order as leader instructions. They are carried as dispatch text, not as committed authority.
- G-17 — Pinned by the lane (progress.md §G): the template value is `"default"`. The board record decides that the template carries the default (Q2 (b)) but does not name the literal. Basis: `.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md:89` uses `"defaultMode": "default"`; `internal/cli/launcher.go:737-741` states that the template no longer ships a default, so under CC 2.1.283+ the CC built-in default wins (auto; Manual under a GLM backend). `acceptEdits` is not chosen: it would auto-approve edits, a behaviour change beyond ruling d-20261010T073547Z-07ae. Ruling d-20261010T081810Z-49e5 (item 2) places this choice in the lane's ladder. (REQ-004, AC-004, M3.)
- G-18 — The approved digest-prefix (`9c93e47809f9`) covers only the header copy `.moai/reports/t1630/ruling-d-20261010T073547Z-07ae.txt`. The decisions sit in the body line, which each DECIDED row copies verbatim and which was matched to the board listing as a whole line (`moai decision read --all`, grep -x -F, one match). The 12-hex digest does not cover the body line. Extending the digest to the body changes the approved value and needs a new approval.
- G-19 — Resolved by ruling d-20261010T081810Z-49e5 (item 2): the record's `--yes` is the existing CLI confirmation flag, `--force` (`internal/cli/clean.go:106`). No new flag is added. REQ-011 and AC-011 clause (iii) use `--force`.
- G-20 — Pinned by the lane (progress.md §G): the managed-block detection is diff-based against the sidecar record `.moai/state/tool-policy/managed-allow.json` (`last_generated`). user_added = existing allow entries minus `last_generated`; result = regenerated ∪ user_added; with no record, every existing entry is kept. The record's allowance is "marker- or diff-based" (board:d-20261010T073547Z-07ae). Probes: `.moai/reports/t1630/evidence/probe-ac005_test.go.txt` (case i) and `.moai/reports/t1630/evidence/probe-ac005b_test.go.txt` (case ii). (REQ-005, AC-005, M3.)
- G-21 — AC-011 clause (iii) (a run item referenced by a live record is not a candidate; a db item is deleted only with `--force`, the record's `--yes`) is a regression-guard, not release-blocking. No RED cell can be observed at plan time: the scan has no run or db deletion path yet, so an exclusion assertion would pass vacuously. The run phase authors its RED, with a mutant probe, before its GREEN.
- G-22 — The record states that `tier_render.go:106` "writes only when the value is absent". AC-003 part B observes the overwrite on this tree (3975fe3cc: `plan` replaced by `acceptEdits` on the default path and by `auto` on the automatic path). The statement describes the decided rule, not the current code, and it is not used as evidence that the current code is correct.

#### §A.4 Basis — affected-file estimate for the tier

Production and test files expected to change under the default design (§D, §F):

1. `internal/config/toolpolicy/tier_render.go` (writer fix)
2. `internal/config/toolpolicy/tier_render_test.go` (preservation assertions)
3. `internal/core/project/autonomy_bundle_test.go` (default-path assertions)
4. `internal/cli/clean_home.go` (run and db categories)
5. `internal/cli/clean_home_test.go` (run and db cases)
6. `internal/cli/main_test.go` (process HOME and CODEX_HOME redirection)
7. `internal/cli/codex_review_gate_live_test.go` (CODEX_HOME temporary root)
8. `internal/homestate/main_test.go` (MOAI_HOME sandbox)
9. `internal/escalation/main_test.go` (MOAI_HOME sandbox)
10. `internal/factory/main_test.go` (MOAI_HOME sandbox)
11. `internal/factorymsg/main_test.go` (MOAI_HOME sandbox)
12. `internal/web/main_test.go` (MOAI_HOME sandbox)
13. `internal/contract/receipt/main_test.go` (new file, MOAI_HOME sandbox)
14. `internal/testhome/testhome.go` (new shared sandbox helper)
15. `internal/testhome/guard_test.go` (new guard, REQ-010)

The count is 15 without the template change. The decided Q2 verdict adds the template file, making 16; the 16th file is a one-line template edit. The orchestrator's Tier M decision (lane-18, ladder ⑤) overrides the tier-up trigger in spec.md §0, and ruling d-20261010T081810Z-49e5 confirms option (b). Any transitive reach found under G-9 above 15 is reported in progress.md.

### §A.5 Residual-risk (could still be wrong despite the evidence)

- R-1 — The probes call the writer functions directly. The full `moai init` executor also stages and restores settings snapshots (`internal/cli/init.go`, around the `StageDeployedSettingsSnapshot` and `SettleSettingsSnapshot` calls). The end-to-end effect may differ. AC-001 must be observed through the init command in a sandboxed home in the run phase.
- R-2 — Redirecting HOME in test binaries may expose tests that silently depended on the real home. Some green tests may fail or change meaning. The run must run each reaching package under the sandbox before and after the change.
- R-3 — The reach measure (G-9) can undercount. A package left out of §A.4 Basis may still write to the real home. AC-007 is the backstop: a before-and-after manifest of the operator home across the full suite.
- R-4 — Redirecting the process HOME does not stop code that resolves the home through a path other than `os.UserHomeDir` or `paths.Home`. The run must search for such resolvers.
- R-5 — Adding `db` as a clean-home candidate could delete evidence a later run needs. Decided Q3 lists db items and deletes them only with the confirmation flag `--force`, which is the record's `--yes` (G-19, resolved by ruling d-20261010T081810Z-49e5).
- R-6 — The decided Q1 disposition changes what `moai init` does with an existing differing defaultMode: it keeps the value, where the current writer overwrites it. The change is user-visible and is recorded in progress.md when the run lands it.

## §B Known Issues

- B1 — The default init overwrites a user-set USER-scope defaultMode. E-3 shows `plan` replaced by `acceptEdits`, and AC-003 part B shows the same overwrite on the automatic path (`auto`). Decided Q1 (keep the existing value) settles the disposition on both paths; the run implements it.
- B2 — The writer loses the user's `allow`, `ask`, and `deny` on every default init (E-4). This is the card's item (a) and is the first fix. It contradicts the condition pinned by SPEC-INIT-WIZARD-REPAIR-001 §4 (E-23, decision-index Q8). The M1 preservation test that the condition requires is absent, so M2 adds it.
- B3 — The doc comment on `WriteUserDefaultMode` (tier_render.go:93-94) describes preservation that the code does not perform. The comment is corrected with the fix.
- B4 — Scope item S-7 (t1594) is blocked. Its residue inventory and card text are not in the tree (G-2). The run cannot start the t1594 item until the card text is supplied to the leader.
- B5 — Scope item S-6 (t1567) is blocked on Q2 and G-1.
- B6 — The run cannot start before the three ordering landings (REQ-013, G-13).
- B7 — The live review-gate test may write to the operator's `~/.codex` today (E-10, G-10). The run must measure this in a sandbox before it changes the test.

## §C Pre-flight (run phase, before any change)

- P-1 — Record the tree identity and `git status --short` (must be empty).
- P-2 — Confirm the ordering landings (REQ-013): `git merge-base --is-ancestor <landing-sha> HEAD` for each of the three SHAs the leader names, expecting exit 0.
- P-3 — Re-observe every RED cell of acceptance.md on the post-landing run base, with the same four-element commands. Record that observation beside the pinned-tree observation (3975fe3cc). P-3 and M2 both record their observations in progress.md. Probes write only under `.moai/reports/t1630/evidence/`.
- P-4 — The operator-home manifest is taken as a read-only listing by the operator, not by an agent (AC-007 baseline, G-8).

## §D Constraints

See spec.md §4 (C1 to C6). In addition: the run must not run any unsandboxed test binary to observe a leak (G-10). Leak observation uses a throwaway account or a copy of the home.

## §E Self-Verification (plan phase)

- E-PL-1 — SPEC ID regex check: PASS (E-2).
- E-PL-2 — Frontmatter carries the 12 canonical fields. Check: `grep -c -E '^(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags):' .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md` must print 12.
- E-PL-3 — Counts within the Tier M ceilings: 13 requirements of 16, and 13 criteria of 16 (checks listed in acceptance.md).
- E-PL-4 — No `[NEEDS CLARIFICATION` marker in spec.md or acceptance.md. Blocked items are recorded as B-items and in decision-index.md.
- E-PL-5 — spec.md carries the `### Out of Scope —` H3 sub-headings with `-` bullets.

## §F Milestones (ordered by decision reversibility; decisions first)

Priority labels only, no time estimates.

- M1 (Priority High, blocks run) — Decision verdicts: Q1, Q2, Q3, and Q5 are DECIDED by the pinned board record `board:d-20261010T073547Z-07ae#9c93e47809f9` (decided_by 영실이 판단, operator-delegated). Q4 and Q6 are closed by the evidence produced at M2 and do not block run. Q7 stays open with no applied default (decision-index.md); it blocks the autonomous Kickoff until an operator verdict is recorded.
- M7 (gate, cross-cutting) — Ordering (REQ-013) checked before M2 starts and again before the final sync. Covers AC-013.
- M2 (Priority High) — Init USER-scope writer (REQ-001, REQ-002, REQ-003): RED probes on the current tree, a GREEN writer that preserves the three lists and every unmodelled key, a no-op path when nothing changes, and the decided Q1 disposition (an existing defaultMode is kept, written only when absent). Covers AC-001 to AC-003.
- M3 (Priority High) — Policy-path and template dispositions (REQ-004, REQ-005): the template carries `permissions.defaultMode` = `"default"`, written by a fresh init only where absent and never by update (decided Q2; value pinned in progress.md §G, G-17); PROJECT-scope user-added allow entries are kept by set union with the regenerated managed block, using diff-based detection against `.moai/state/tool-policy/managed-allow.json` (decided Q5; pinned in progress.md §G, G-20); with no record, every existing entry is kept. Covers AC-004, AC-005.
- M4 (Priority Medium) — `clean --home` run and db candidates (REQ-011, REQ-012) under the decided Q3 rule (a run item is a candidate only when no live record references it; a db item is listed without `--force` and deleted only with `--force`, the record's `--yes`; G-19 is resolved by ruling d-20261010T081810Z-49e5); the run and db categories join the `--force` allowlist (progress.md §G, G-23); dry-run and allowlist regressions kept. Covers AC-011 clauses (i) and (ii) (release-blocking), AC-011 clause (iii) (regression-guard, G-21), and AC-012; the run-phase deletion criterion named in acceptance.md's Definition of Done is authored here.
- M5 (Priority Medium) — Codex live isolation (REQ-009): a CODEX_HOME temporary root in the review-gate live test, measured before and after on a sandboxed home. Covers AC-009.
- M6 (Priority Low, mechanical) — Sandboxing for the reach set (REQ-006, REQ-007, REQ-008, REQ-010): a MOAI_HOME sandbox in TestMain for the six packages that lack one (homestate, escalation, factory, factorymsg, web, and contract/receipt, which has no TestMain); the process HOME redirected to a test-owned directory for every reaching package, including `internal/cli` (its TestMain sets MOAI_HOME but not HOME); `internal/hook` already sets both; the shared helper and the guard in `internal/testhome`. Covers AC-006, AC-007, AC-008, AC-010.
- M8 (blocked) — The t1594 residue (S-7) waits for card text. No requirement is authored for it in this revision.

## §G Anti-patterns

- AP-1 — Rewriting the USER-scope `permissions` object as a whole. Lists must be merged, not replaced.
- AP-2 — Running an unsandboxed test binary to "check" a leak. The leak is observed on a throwaway home only.
- AP-3 — Citing `status: completed` or `implemented` as proof that a behaviour is live. Each claim needs a grep or a test on this tree.
- AP-4 — Adding a requirement for t1567 or t1594 without their card text.
- AP-5 — Treating a name-based reach count as complete. The count is a lower bound (G-9).

## §H Cross-References

- `.claude/rules/moai/development/verification-completeness.md` — §1.1 observed failure, §1.2 three-part check spec, §2 two-cell adoption (RED-now plus green path), §3 cross-layer sweep.
- `.claude/rules/moai/core/verification-claim-integrity.md` — §2 attribution, §2.2 tool provenance (not triggered here, §A.3), §3 five-section report.
- `.claude/rules/moai/workflow/spec-workflow.md` — § SPEC Complexity Tier (Tier M band, REQ and AC ceilings).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — the canonical 12 fields, and `phase` as a release label.
- `.moai/config/sections/interview.yaml` — `decision_gate: on` (decision-index.md required).

## Appendix A — Probe sources (reproduction)

A-1 (`internal/config/toolpolicy`, probe 1; source `.moai/reports/t1630/evidence/probe-e3_test.go.txt`; overlay `.moai/reports/t1630/evidence/overlay-e3.json`; measured at 3975fe3cc). Package `toolpolicy`. The probe writes a temporary settings file with `defaultMode` "plan" and the three lists, calls `WriteUserDefaultMode(path, "acceptEdits")`, and logs the file before and after. It uses `t.TempDir()` only.

A-2 (`internal/core/project`, probe 2; source `.moai/reports/t1630/evidence/probe-e4_test.go.txt`; overlay `.moai/reports/t1630/evidence/overlay-e4.json`; measured at 3975fe3cc). Package `project`. The probe writes a temporary USER-scope settings file with `allow` and `deny` and no `defaultMode`, calls `ApplyAutonomyTierBundle(dir, userPath, "", "")`, and logs the file before and after. It uses `t.TempDir()` only.

A-3 (`internal/core/project`, AC-003 part B; source `.moai/reports/t1630/evidence/probe-ac003b_test.go.txt`; overlay `.moai/reports/t1630/evidence/overlay-ac003b.json`; measured at 3975fe3cc). Package `project`. The probe writes a USER-scope settings file with `defaultMode` "plan" and calls `ApplyAutonomyTierBundle` with an empty persisted tier and with the `automatic` tier. It sets HOME and MOAI_HOME inside its test process to t.TempDir paths.

A-4 (`internal/cli`, AC-011 clause (ii); source `.moai/reports/t1630/evidence/probe-ac011_test.go.txt`, function `TestAC011DBEntryListed`; overlay `.moai/reports/t1630/evidence/overlay-ac011.json`; measured at 3975fe3cc). It places one aged `db/` entry under a temporary root and calls `scanHomeCleanable`.

A-5 (`internal/core/project`, AC-005 case ii; source `.moai/reports/t1630/evidence/probe-ac005b_test.go.txt`, function `TestAC005WithRecordKeepsUserAddedEntry`; overlay `.moai/reports/t1630/evidence/overlay-ac005b.json`; measured at 3975fe3cc). It writes the sidecar record `{"last_generated":["Bash(old-managed:*)"]}` and a project allow list holding that entry and `Bash(user-added:*)`, then calls `ApplyAutonomyTierBundle` with the automatic tier. It sets HOME and MOAI_HOME inside its test process to t.TempDir paths.

All probes are injected through a JSON overlay in `.moai/reports/t1630/evidence/` (`overlay-e3.json`, `overlay-e4.json`, and the overlays named in acceptance.md). Each overlay maps a repository path inside the package directory to the evidence source. The overlay never touches the repository. The plan-phase scratch copies are superseded by these evidence files.

Evidence binding (D23; option c, ruling d-20261010T084134Z-612e). The D23 wording is: "결속은 매니페스트 해시로, 원본은 로컬 .moai/reports/t1630/" (binding is by manifest hash; the originals stay local under `.moai/reports/t1630/`). The originals are the probe sources, overlays, scripts, fixtures, recorded listings, and the names-only record in `.moai/reports/t1630/evidence/`. `.gitignore:235` excludes that directory from git, so none of it is tracked (operator directive 2026-09-14 keeps evidence on disk and off the remote). The one tracked artifact is `.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence-manifest.json`. For each of its 37 files it records the worktree-relative `path`, the `sha256`, the `size` in bytes, and the `created_at` birth time in UTC, with a top-level `generated_at`. It holds hashes, paths, sizes, and times only: no values and no file bodies. Verify it from the worktree root with `jq -r '.files[] | "\(.sha256)  \(.path)"' .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence-manifest.json | shasum -a 256 -c -`; all 37 lines must read OK. The originals moved from `.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence/` to `.moai/reports/t1630/evidence/` after the pre-move manifest re-verified 37 of 37 OK in the same run. Fourteen files then differed from their pre-move bytes by the directory prefix alone: eleven overlays, the AC-007 leak probe fixture, and the two recorded AC-007 listings, whose absolute paths named the old directory. For each of the fourteen, reversing that one substitution reproduces its pre-move bytes exactly; the other twenty-three are byte-identical. The fourteen rewritten files carry their rewrite time as `created_at`. The recorded RED outputs in acceptance.md are not edited. Residual risk, disclosed: a fresh clone cannot re-run the RED cells without the local evidence directory.
