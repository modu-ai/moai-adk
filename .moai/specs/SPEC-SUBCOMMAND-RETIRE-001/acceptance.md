# Acceptance — SPEC-SUBCOMMAND-RETIRE-001

> Binding acceptance criteria for the retirement of 5 `/moai` subcommands + 7 dependent
> skills. Verification is evidence-bearing: every PASS row corresponds to an actually-run
> command (verification-claim-integrity). The single most load-bearing gate is AC-SCR-004
> (`go test ./internal/template/...` GREEN).

## §A. Acceptance Criteria Matrix

| AC | Requirement | Verification command / evidence | Pass condition |
|----|-------------|----------------------------------|----------------|
| AC-SCR-001 | REQ-SCR-001 | `ls .claude/commands/moai/ internal/template/templates/.claude/commands/moai/`; `ls .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/` | No `brain`/`design`/`e2e`/`coverage`/`security` command or workflow file in either tree |
| AC-SCR-002 | REQ-SCR-002 | `ls -d .claude/skills/<7names>/ internal/template/templates/.claude/skills/<7names>/ 2>&1` | All 7 skill dirs absent in both trees |
| AC-SCR-003a | REQ-SCR-001/003 | `grep -c 'name: moai-domain-ideation\|name: moai-domain-research\|name: moai-domain-brand-design\|name: moai-domain-copywriting\|name: moai-domain-design-handoff\|name: moai-workflow-design\|name: moai-workflow-gan-loop' internal/template/catalog.yaml` | `0` (all 7 entries removed); `moai-domain-humanize` entry still present |
| AC-SCR-003b | REQ-SCR-003 | `make build` then `git diff --exit-code internal/template/embedded.go` after a 2nd `make build` | build succeeds; idempotent (2nd build no diff); `embedded.go` regenerated |
| AC-SCR-004 | REQ-SCR-004 | `go test ./internal/template/...` | Zero failures. Binds: count constants reconciled (35→28, 42→35, 42→35); `TestBrainCommandThinPattern` deleted; **all THREE `agentless_audit_test.go` path-list edits applied** — `coverage.md` removed from `utilitySkillPaths` (else `TestAgentlessUtilityNoLLMControlFlow` + `TestUtilitySkillsContainModeFlagIgnoredSentinel` RED), `design.md` removed from `implementationSkillPaths` (else `TestImplementationSkillsContainPipelineRejectionSentinel` RED), `design.md` removed from `runDesignSkillPaths` (`TestRunDesignSkillsContainModeUnknownSentinel`) |
| AC-SCR-004b | REQ-SCR-004 | `go test ./...` | Zero failures (no cascading breakage outside `internal/template`) |
| AC-SCR-005 | REQ-SCR-005 | `grep -nE '^\s*-?\s*\*?\*?(brain|design|e2e|coverage|security)\b' .claude/skills/moai/SKILL.md` (both trees) | Router header list + Priority-1 routing + Quick-Reference blocks for the 5 are gone; remaining matches are retained-context only (e.g. `review --security` note) |
| AC-SCR-006 | REQ-SCR-006 | per-basename absence loop over the 7 removed names across both `.claude/skills/` + `internal/template/templates/.claude/skills/` (see plan.md M4 script); `ls` both trees | The seven removed skill basenames are absent from BOTH trees; `commands/moai/` + `workflows/` free of the 5 in both trees. Skill counts are **asymmetric by design: template = 28, local = 30** (28 + 2 user-owned `harness-*`). A raw full-tree `diff` expecting identity is NOT used (would false-FAIL or motivate a §24 violation) |
| AC-SCR-006b | REQ-SCR-006 (hardening) | `moai init` (or `moai update`) into a clean temp project, then `ls <proj>/.claude/skills/` + `ls <proj>/.claude/commands/moai/` | Resurrection-negative: a freshly-deployed project contains NONE of the 7 retired skills and NONE of the 5 retired command/workflow files — confirms the embedded.go regen (the real Template-First binding mechanism) propagated the removal to distributed users |
| AC-SCR-012 | REQ-SCR-012 | `ls -d .claude/skills/harness-*/`; `ls -d internal/template/templates/.claude/skills/harness-*/ 2>&1` | The two user-owned `harness-moaiadk-best-practices` + `harness-moaiadk-patterns` skills are PRESENT in the local tree and ABSENT from the template tree (unchanged by the removal); no `harness-*` skill was deleted (§24) |
| AC-SCR-007 | REQ-SCR-007 | Read SKILL.md security-routing note; confirm `moai-ref-owasp-checklist`/`moai-ref-secops`/`moai-ref-supply-chain`/`moai-ref-llm-security` still present in catalog | Security audit reroutes to `/moai review --security` + `Agent(general-purpose)` loading retained ref skills; the 4 ref skills NOT removed |
| AC-SCR-008 | REQ-SCR-008 | repo-wide grep (scratch/`.moai/specs/`/backup excluded; FROZEN `design/constitution.md` + `zone-registry.md` exempted) for the 5 subcommands + 7 skills across moai.md, CLAUDE.md, spec-workflow.md, glm-web-tooling.md, moai-domain-humanize | 0 dangling references in the in-scope surfaces |
| AC-SCR-009 | REQ-SCR-009 | `go test ./internal/template/ -run 'InternalContentLeak\|Neutrality'`; template-neutrality CI guard | Pass — no forbidden content class introduced |
| AC-SCR-010 | REQ-SCR-010 | `grep -n 'moai-domain-copywriting' .claude/skills/moai-domain-humanize/SKILL.md` (both trees) | `0` — humanize cross-ref rewritten to drop the retired-skill dependency |
| AC-SCR-011 | REQ-SCR-011 | `ls docs-site/content/{en,ko,ja,zh}/quality-commands/`; `grep -n 'coverage\|e2e' docs-site/data/menu/main.yaml`; `hugo --cleanDestinationDir` | No `moai-coverage.md`/`moai-e2e.md` in any locale; no menu entries; build succeeds; en/ko/ja/zh symmetric |

## §B. Given-When-Then Scenarios

### Scenario 1 — A distributed user runs `moai init` after the retirement (core outcome)
- **Given** the template source has had the 5 subcommands + 7 skills removed and `make build` regenerated `embedded.go`,
- **When** a fresh user runs `moai init myproject`,
- **Then** the deployed project contains no `brain`/`design`/`e2e`/`coverage`/`security` command or workflow file, the skills catalog deploys 28 skills (not 35), and `/moai` natural-language routing presents no path to the 5 retired subcommands.

### Scenario 2 — CI runs the template guard suite after removal (highest-risk gate)
- **Given** the 7 skill dirs + catalog entries are removed and the count constants are reconciled (35→28 skills, 42→35 total),
- **When** CI runs `go test ./internal/template/...`,
- **Then** the suite is GREEN: `catalog_tier_audit_test`/`catalog_loader_test`/`embed_catalog_test` count assertions pass, `TestBrainCommandThinPattern` is absent, and `TestRunDesignSkillsContainModeUnknownSentinel` reads only `run.md`.

### Scenario 3 — A user requests a security audit after `/moai security` is gone (no capability gap)
- **Given** `/moai security` is removed but the four `moai-ref-*` security reference skills are retained,
- **When** a user types "run a security audit on the auth module",
- **Then** the orchestrator routes to `Agent(general-purpose)` with security scope loading the retained OWASP/secops/supply-chain/llm-security reference skills (or the user uses `/moai review --security`), and the OWASP audit capability is preserved.

### Scenario 4 — The docs-site is rebuilt across all four locales
- **Given** the coverage+e2e pages and menu entries are removed symmetrically across en/ko/ja/zh,
- **When** `hugo --cleanDestinationDir` builds the site,
- **Then** no menu entry or content page for `coverage`/`e2e` renders in any locale, the build succeeds, and 4-locale parity holds (§17).

## §C. Edge Cases

- **EC-1 (RED-between-commits)**: removing a skill dir without updating the count constant must not be committed. M2 is atomic — verified by `go test ./internal/template/...` GREEN at the milestone boundary, not mid-edit.
- **EC-2 (hash error on missing SKILL.md)**: a catalog entry left behind while its dir is removed makes `gen-catalog-hashes --all` error. Verified by AC-SCR-003a (0 stale entries) preceding AC-SCR-003b (`make build` succeeds).
- **EC-3 (FROZEN false-positive)**: the 0-dangling-ref grep MUST exempt `design/constitution.md` + `zone-registry.md` (FROZEN); a PASS that required editing them is a FAIL of design.md §A.
- **EC-4 (humanize orphan)**: `moai-domain-humanize` is retained but referenced `moai-domain-copywriting`; AC-SCR-010 confirms the rewrite.
- **EC-5 (local/template drift)**: a removal applied to only one tree is caught by AC-SCR-006's per-basename absence check across both trees (NOT a full-tree `diff` — the trees are legitimately asymmetric per EC-6).
- **EC-6 (harness-* asymmetry, §24)**: the local tree carries 2 user-owned `harness-*` skills absent from the template tree, so post-removal counts are 28 template / 30 local. A verifier that asserts "28 each" or runs a raw `diff` expecting identity would either false-FAIL or be tempted to delete the `harness-*` skills (a §24 violation + data loss). AC-SCR-006 + AC-SCR-012 guard against this.

## §D. Quality Gate / Definition of Done

A run-phase milestone is DONE only when ALL hold (evidence-bearing — paste real command output):
- [ ] AC-SCR-001..012 all PASS with cited command output (incl. AC-SCR-006b hardening + AC-SCR-012 harness-* preservation).
- [ ] `go test ./internal/template/...` and `go test ./...` GREEN.
- [ ] `golangci-lint run` clean.
- [ ] `make build` succeeds and is idempotent; `embedded.go` regenerated.
- [ ] Template-neutrality CI guard passes.
- [ ] Repo-wide in-scope 0-dangling-ref grep PASS (FROZEN surfaces exempted).
- [ ] Seven removed basenames absent from BOTH trees; counts are 28 template / 30 local (asymmetric by design); the 2 user-owned `harness-*` skills preserved (§24).
- [ ] docs-site builds; coverage/e2e gone in all 4 locales.
- [ ] No FROZEN `design/constitution.md` / zone-registry edit (negative check).

## §E. Out-of-scope reminders (not acceptance-tested here)
- design rule subsystem, brand config, `internal/design/` Go package — NOT removed (spec §E).
- CHANGELOG/README entries — owned by manager-docs (sync phase).
- Replacement-subcommand re-implementation — none built.
