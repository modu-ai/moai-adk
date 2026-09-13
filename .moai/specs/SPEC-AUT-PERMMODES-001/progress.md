# SPEC-AUT-PERMMODES-001 — Progress

> Card t584. Evidence path for the dispatch: `.moai/specs/SPEC-AUT-PERMMODES-001/progress.md`.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-AUT-PERMMODES-001
card: t584
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
req_count: 10
ac_count: 12
id_regex_check: PASS
owning_spec_of_amended_reqs: SPEC-AUTONOMY-TIERS-001 (REQ-006, REQ-007)
defaultmode_verdict: "auto is a VALID CC defaultMode value (6-value enum, official docs fetched 2026-09-13); bundled IAM reference stale -> refreshed in-scope (REQ-007)"
open_blockers: none
```

Plan-phase research notes (attributable):

- SPEC ID regex check `SPEC-AUT-PERMMODES-001` → PASS (verbatim output in plan-phase session).
- Official permission-modes table fetched 2026-09-13 from `https://code.claude.com/docs/en/permissions` (§ Permission modes): `default` (alias `manual`, CC v2.1.200+) / `acceptEdits` / `plan` / `auto` / `dontAsk` / `bypassPermissions`; kill switches `permissions.disableBypassPermissionsMode` + `permissions.disableAutoMode`.
- Stale-reference evidence: `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md:30-40` ("exactly four values", denies `auto`).
- Downgrade path on this tree: `internal/config/autonomy_tiers.go` (`EffectiveTierWithGates`, `AppendDowngradeAdvisory`) + `internal/core/project/autonomy_bundle.go:75` (log sink `.moai/logs/autonomy-downgrade.log`).
- Regression tests located: `internal/core/project/autonomy_bundle_test.go:89,115,138,158`; `internal/config/autonomy_tiers_toggle_test.go:57`; `internal/cli/init_autonomy_wiring_test.go:237`; `internal/cli/wizard/autonomy_test.go:49`.

## §E.2 Run-phase Evidence

### M1 — Token verification + version floor + scope finding (2026-09-13)

**Scope finding (M1 EXIT GATE — verdict: NO blocker, proceed).** The moai autonomy bundle writes `defaultMode` to USER scope ONLY:

- `internal/cli/init.go:863` calls `applyAutonomyTierBundleFn(opts.ProjectRoot, filepath.Join(homeDir, ".claude", "settings.json"), filepath.Join(opts.ProjectRoot, ".claude", "settings.json"), opts.AutonomyTier)` — `userSettingsPath` IS the USER-scope `~/.claude/settings.json`.
- `internal/config/toolpolicy/tier_render.go:106` `WriteUserDefaultMode(userPath, defaultMode)` writes ONLY the USER file (fallback path).
- `internal/config/toolpolicy/tier_render.go:55` `RenderTierPermissions` splits: defaultMode → USER scope; deny/ask → PROJECT scope with PROJECT defaultMode reset to "" (the AP-6 comment already records the PROJECT-scope silent-fail rule).

Since the silent-ignore caveat below binds PROJECT scope (`.claude/settings.json` / `.claude/settings.local.json`) only, and `acceptEdits` applies from any scope, all three wizard values are effective at the scope moai actually writes. No blocker; M2 proceeds.

**Official docs verification (fetched 2026-09-13, this run, `curl -sL https://code.claude.com/docs/en/permission-modes`, exit 0, 669472 bytes — quotes verbatim from the fetched page):**

1. Scope caveat (load-bearing, verbatim): *"If you set `auto` in `.claude/settings.json` or `.claude/settings.local.json`, the value doesn't take effect, and Claude Code then uses the built-in default rather than a `defaultMode` from `~/.claude/settings.json`. If you set `bypassPermissions` in those two files, it doesn't take effect either, and the session starts in Manual mode. The other values apply from any settings file."*
2. Version floors (verbatim): *"The built-in auto default requires Claude Code v2.1.228 or later on macOS, Linux, and WSL, and v2.1.233 or later on native Windows. On earlier versions, the built-in default is Manual."*
3. `manual` alias (verbatim): *"The CLI accepts `manual` as an alias wherever you type the value ... The Manual label and the `manual` alias require Claude Code v2.1.200 or later."*
4. Kill switch (verbatim): *"To remove auto mode so nobody can select it, set `permissions.disableAutoMode` to `\"disable\"` instead"* — settable in managed settings; recorded as Out-of-Scope context (no `EffectiveTierWithGates` wiring this card, per spec §D).
5. Six-value enum confirmed on the page's own section list: default (Manual) / acceptEdits / plan / auto / dontAsk / bypassPermissions. Cloud-web note (context): *"Claude Code on the web does not honor `defaultMode: \"bypassPermissions\"` or `\"dontAsk\"` from your settings files ... ignored silently"* — terminal/CLI scope is unaffected; moai writes for terminal sessions.

**Local CC build round-trip (M1 plan §F.1):** `claude --version` → `2.1.270 (Claude Code)` — ≥ both floors (v2.1.228/v2.1.200), so the local build reads `auto` from settings.

**REQ-007 edit list (plan §F.M1.3) — `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md` § Permission Modes:**

- `:30` "accepts exactly four values" → six values (add `auto`, `dontAsk` rows).
- `:40-41` "These are the only valid values. Earlier revisions listed `dontAsk` and `ignore` — those are not real Claude Code permission modes." → remove the `dontAsk` denial (now false), keep the note only for retired/illusory tokens.
- Add kill-switch pair note (`disableBypassPermissionsMode` + `disableAutoMode`) + version floors (`manual` alias v2.1.200+; built-in auto default v2.1.228+/v2.1.233+ Windows) + the project-scope silent-ignore caveat.

**Pre-flight baselines (this run, HEAD 162b6ef92):** `go build ./...` → BUILD_OK; `GOOS=windows GOARCH=amd64 go build ./...` → WIN_BUILD_OK; `golangci-lint run --timeout=2m ./internal/...` → 328 pre-existing issues (errcheck 296 / staticcheck 30 / unused 2), exit 0; affected-package tests (`./internal/cli/wizard/... ./internal/config/... ./internal/core/project/...`) all `ok`.

## §F Phase 4 Mode Selection

- Input parameters: tier M · scope ~8-10 files (wizard questions/translations, config tiers, core/project bundle, cli init flag help, IAM reference doc) · domains 4 (Go cli, Go config, Go core, template docs) · language mix Go + markdown · concurrency benefit LOW (coding-heavy) · Agent Teams prereqs: not requested.
- Mode evaluation: direct — not selected (multi-file, semantic change); fanout — not selected (coding-heavy, Anthropic caveat); sweep — not selected (not mechanical-uniform); agent-team — not requested.
- Decision: **serial** (one manager-develop milestone spawn at a time).
- Justification: coding-heavy work on interdependent surfaces (question options → bundle mapping → tests); sequential milestone spawns preserve ownership boundaries and keep verification evidence attributable per milestone. Kickoff approval: PASSED (operator approval 2026-09-13, relayed by lead). M1 (CC `auto` version-floor verification) is a hard predecessor of M2 per plan §F; a contradicting M1 finding returns as a blocker, not a workaround.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
