---
id: SPEC-SKILL-BODY-NEUTRALITY-001
title: "Skill-Body Neutrality — run-phase progress"
version: "0.1.1"
status: in-progress
created: 2026-06-04
updated: 2026-06-23
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills"
lifecycle: spec-anchored
tags: "template-system, neutrality, skills, ci-guard, distribution"
tier: M
---

# Run-phase Progress — SPEC-SKILL-BODY-NEUTRALITY-001

## §E.1 Run-phase context (re-run on current main base)

- Worktree base: this run-phase is a RE-RUN on the current local main HEAD `9c5f0e2b1` (the prior run-phase, ~640 commits ago, is STALE and discarded).
- L1 worktree materialized by runtime at `.claude/worktrees/agent-a6a80c9a8d9750c04`, initially based on `origin/main` (`3629ed232`) which LACKED the SPEC files. The 4 intervening commits between `3629ed232` and local main `9c5f0e2b1` touched ONLY SPEC artifacts + CLAUDE.local.md — NOT skill bodies or the leak test — so the edited files are byte-identical at both bases. The worktree-base divergence is therefore inert for this SPEC's scope; final commits FF/cherry-pick cleanly onto local main.
- cycle_type: tdd (Part B RED guard first → Part A purge → Part B GREEN).
- Plan inventory was authored ~640 commits ago and DRIFTED. Ground-truth re-grep was performed fresh on the current tree (see §E.2 RED evidence). Key divergences from stale plan:
  - CLASS4 `99-release`: plan claimed 6 hits; current tree has **2** (`INDEX.md:143`, `moai/references/reference.md:241`). `commands-reference.md` no longer carries `99-release`.
  - CLASS4 `97-release-update` / `release-update` / "NOT distributed": **0** hits (already cleaned in the 640-commit interval; `moai/SKILL.md` carries none).
  - CLASS4 maintainer-doctrine date `2026-05-26`: **0** hits. Residue is `catch-up SPEC` + `maintainer doctrine` in `moai-meta-harness/SKILL.md` (lines 232, 248).
  - CLASS3 SPEC-V3R: 37 files (matches plan). CLASS3 REQ-token: 27 files (broader than plan's stated 13 — default-tier regex). CLASS2 Go-path: 9 files (incl. `internal/design/dtcg/frozen_guard_test.go`).

## §E.2 Run-phase Evidence

### M1 — Part B RED (extend the neutrality guard) — COMPLETE

Guard file edited: `internal/template/internal_content_leak_test.go`. Additions (all skill-body-scoped via new `skillBodyScoped` field + `skillBodyPrefix = ".claude/skills/"`, so they do NOT fire on agents/rules/hooks/config per EXCL-SBN-002):

- `C1b-spec-id-skill-v3r` — broadens SPEC-ID detection to `SPEC-V3R[0-9]-*` / `CONST-V3R[0-9]-*` + named `SPEC-WF-AUDIT-GATE-001` / `SPEC-MX-001` (REQ-SBN-014 / REQ-SBN-006).
- `C6-agentless-test-ref` — matches the literal `agentless_audit_test.go` (REQ-SBN-012).
- `C7-internal-go-path` — package-restricted `internal/(spec|cli|hook|ciwatch|design)/[a-z0-9_/]*\.go` (REQ-SBN-013 HARD; does NOT match `internal/auth|api|core` illustrative paths).
- `S3-req-ac-token-any-prefix` — PROMOTED from former strict tier into the default tier (skill-body-scoped), as a strict-superset regex `(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+` (catches both `REQ-BRAIN-001` and the `REQ-WF003-010` form; the original narrow `[A-Z]{2,}` form missed the WF-NNN form). The former `strictLeakClasses` S3 sibling was REMOVED so there is exactly ONE REQ-token regex entry (AC-SBN-018(b) partition guard). The superset still matches the original S3 probe → remains a strict superset.

Belt-and-suspenders allowlist (REQ-SBN-013 / AC-SBN-020(c)): the 3 illustrative Go paths (`internal/auth/login.go`, `internal/api/handler.go` in `pr-review-multi-agent.md`; `internal/core/handler.go` in `mx.md`) added to `pedagogicalAllowlist`.

New structural unit tests added: `TestLeakClassReqTokenPartition` (AC-SBN-018(b)), `TestLeakClassNoDateShaInDefaultTier` (AC-SBN-018(a)), `TestSkillBodyLeakClassRecurrenceBackstop` (AC-SBN-017), `TestC7PackageRestriction` (AC-SBN-020(a)+(b)).

#### M1 RED checkpoint — AC-SBN-012 evidence (test FAILS on current leaks)

Command: `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$'`
Result: **FAIL** — `template internal-content leak detected (185 occurrences, mode=narrow)`.

Per-class uncapped counts (grep ground-truth against `internal/template/templates/.claude/skills/`):

| Class | Occurrences | Files | Requirement |
|-------|-------------|-------|-------------|
| C6-agentless-test-ref | 6 | 6 | AC-SBN-012 (≥6 file findings) — PASS |
| C1b-spec-id-skill-v3r | 101 | — | REQ-SBN-006/014 |
| C7-internal-go-path | 12 | 9 | REQ-SBN-005/013 |
| S3-req-ac-token-any-prefix | 131 | — | REQ-SBN-007 |

The scan correctly does NOT flag the 3 illustrative paths (allowlisted) nor `SPEC-AUTH-001`/`SPEC-001` placeholders.

Structural unit tests at M1 (all PASS — they validate the guard structure, not the leaks):
- `TestLeakClassReqTokenPartition` PASS
- `TestLeakClassNoDateShaInDefaultTier` PASS
- `TestSkillBodyLeakClassRecurrenceBackstop` PASS
- `TestC7PackageRestriction` PASS

### M2 — Part A CLASS1 + CLASS2 build/Go-path purge (workflow skills) — COMPLETE

CLASS1 (REQ-SBN-001/002/003): rewrote the `agentless_audit_test.go` prose in all 6 files (`loop.md`, `design.md`, `run.md`, `sync.md`, `plan.md`, `run/context-loading.md`). Kept every sentinel keyword VALUE; dropped the test-file name + REQ tokens. The `TestLoopAliasCrossReference` audit still finds the literal `/moai run --mode loop` cross-reference (preserved in loop.md).
- AC-SBN-001: `grep -rln 'agentless_audit_test' $SK/` → **0** (PASS).
- AC-SBN-002: sentinels retained — `MODE_UNKNOWN` 3 files, `MODE_PIPELINE_ONLY_UTILITY` 7, `MODE_FLAG_IGNORED_FOR_UTILITY` 5, `MODE_TEAM_UNAVAILABLE` 3. Sentinel tests `go test ./internal/template/ -run 'Sentinel|AgentlessControlFlow|LoopAlias'` → **ok** (PASS).
- AC-SBN-003: no line with a `MODE_*` sentinel AND a `REQ-WF` token → **0** (PASS).

CLASS2 release build (REQ-SBN-004, `sync/delivery.md`): generic-ized the 5-platform `GOOS=... ./cmd/moai/` block to a `<your-module>` placeholder pattern; dropped the pinned `golangci-lint@v2.1.6` version (both occurrences).
- AC-SBN-004(a) `GOOS=...cmd/moai/` → 0; (b) `golangci-lint.*@v2.1.6` → 0; (c) `<your-binary>|<your-module>|<target>` → 6 (PASS).

CLASS2 Go-impl paths in workflow skills (REQ-SBN-005): generic-ized `internal/spec/lint.go FrontmatterSchemaRule` → "the SPEC frontmatter lint rule" (`plan/spec-assembly.md`, `team/plan.md`); `internal/cli/harness.go` → "the harness CLI verb path" (`harness.md`); `internal/hook/dbsync/db_schema_sync.go` + `internal/cli/hook.go` → "the DB-schema-sync hook handler" (`sync/quality-gates-context.md`). The 6 non-workflow Go-paths are deferred to M3.

Local mirror synced byte-identical for all 12 edited files (REQ-SBN-011 / AC-SBN-011 verified via `diff`). `make build` → exit 0 (catalog.yaml unchanged — workflow sub-files do not feed the SKILL.md-level catalog hash).

### M3 — Part A CLASS2 Go-path purge (non-workflow skills) — COMPLETE

Generic-ized the 6 remaining real Go-impl paths in `moai-workflow-spec/SKILL.md` + `references/reference.md` (`internal/spec/lint.go` → the SPEC frontmatter lint rule), `moai-workflow-worktree/SKILL.md` (`internal/cli/worktree/team_launch.go` → the team-launch entry point), `moai-workflow-ci-loop/SKILL.md` (`internal/ciwatch/*.go`, `internal/cli/pr/watch.go` → the CI-watch helpers), `moai-workflow-design/SKILL.md` (`internal/design/dtcg/frozen_guard_test.go` → the DTCG frozen-guard CI test). AC-SBN-005 → 0. Mirror synced. (M2 and M3 committed together as f36eec41b.)

### M4 — Part A CLASS3 purge: internal SPEC IDs + REQ tokens + 4-locale — COMPLETE

- 4-locale (REQ-SBN-019): dropped the maintainer-facing "4-locale" annotation at all 4 §B.5 sites (`spec-ears-format.md`, `moai-workflow-spec/SKILL.md` ×2, `references/reference.md`), keeping the `adk.mo.ai.kr` URL. AC-SBN-019(a)=0, (b)=3 files preserved.
- SPEC-V3R IDs (REQ-SBN-006): generic-ized all `SPEC-V3R[0-9]-*` / `CONST-V3R[0-9]-*` + named `SPEC-WF-AUDIT-GATE-001` / `SPEC-MX-001` to plain-language policy descriptions (e.g. `SPEC-V3R5-LATE-BRANCH-001` → "the late-branch opt-in policy", `SPEC-WF-AUDIT-GATE-001` → "the plan audit gate contract", `SPEC-MX-001` → "the MX tag protocol", harness SPEC IDs → "the harness foundation/learning/lifecycle policy"). AC-SBN-006(a)=0, (b)=0. Placeholders preserved (SPEC-AUTH-001 ×14).
- REQ/AC tokens (REQ-SBN-007): generic-ized all `REQ-<PREFIX>-NNN` / `REQ-WF<NNN>-NNN` / `AC-<PREFIX>-NNN` in prose (parenthetical citations removed, range citations → "the relevant requirements", `Verifies REQ-X:` comments → `Verifies:`, `REQ coverage:` footers → "(internal provenance omitted)"). AC-SBN-007=0.
- Recovered one perl byte-mode UTF-8 corruption: a `($1)` note-preservation substitution mangled an em-dash in `design.md:80` (`(— idempotency)` → `(idempotency)`); the file was re-validated as clean UTF-8 and the residual `## SPEC Reference` block REQ tokens (lines 5-6, 124) were then fixed with surgical Edits. All 50 edited template files verified valid UTF-8.

### M5 — Part A CLASS4 purge: dev-only self-ref + maintainer doctrine — COMPLETE

- Dev-only `99-release` (REQ-SBN-008): removed the `/moai:99-release` row from `moai-foundation-core/modules/INDEX.md` and the `/moai:99-release` note line from `moai/references/reference.md` (the 2 baseline hits on current tree — the plan's claimed 6 was 640-commit-stale; `commands-reference.md` already carried none, and `97-release-update` / "NOT distributed" were already 0). AC-SBN-008: all 3 greps → 0.
- Maintainer doctrine (REQ-SBN-009): replaced the "catch-up SPEC" + "maintainer doctrine" + "doctrine-code drift" residue in `moai-meta-harness/SKILL.md` (lines 232, 248) with a generic namespace-separation statement. AC-SBN-009(a)=0, 2026-05-26=0, (b) `harness-*`/namespace policy still present (21 matches). The earlier `moai-meta-harness/SKILL.md:168` doctrine date the stale plan cited was already absent (640-commit drift).
- `team/glm.md` deferred-site note: this file carries 0 leaks of any class on the current tree and was never modified by this run (the GLM WIP lives only in the main checkout, not in this worktree), so it is NOT a deferred site — no purge needed and no clobber risk.

### M6 — Part B GREEN + finalize — COMPLETE

- AC-SBN-016 [MUST-PASS]: `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$'` → **ok** (GREEN). Full leak re-scan: C1(agentless)=0, C1b(SPEC-V3R)=0, C7(Go-path)=0, S3(REQ/AC)=0, 4-locale=0.
- AC-SBN-017 recurrence backstop: `TestSkillBodyLeakClassRecurrenceBackstop` PASS (each new class fires on a re-leak, not on a clean replacement).
- AC-SBN-018 partition guards: `TestLeakClassReqTokenPartition` + `TestLeakClassNoDateShaInDefaultTier` PASS.
- AC-SBN-020 C7 restriction: `TestC7PackageRestriction` PASS.
- CI wiring: `.github/workflows/template-neutrality-check.yaml` path filter already covers `internal/template/templates/.claude/skills/**` (via the `internal/template/templates/**` glob); added a `Run TestTemplateNoInternalContentLeak (isolated)` step + the leak-test file to the path-trigger list so the guard runs in CI.
- Full verification batch: `go test ./internal/template/...` → ok; `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `go vet ./internal/template/...` → exit 0; `make build` → exit 0. Mirror byte-identical for all 50 edited files (AC-SBN-011).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-23
run_commit_sha: bc350cf04
run_status: implemented
ac_pass_count: 20
ac_fail_count: 0
m1_red_evidence: captured (185 occurrences, C6=6 files, C1b=101, C7=12, S3=131)
m6_green_evidence: TestTemplateNoInternalContentLeak PASS; all leak classes 0
cross_platform_build: { host: exit-0, windows_amd64: exit-0 }
new_warnings_or_lints_introduced: false
total_run_phase_files: 52 (50 skill template files + 50 mirror files share content; leak test + CI workflow + progress.md)
m1_to_mN_commit_strategy: M1 (42499880a) / M2-M3 (f36eec41b) / M4-M6 (this commit)
```

## §E.4 per-AC PASS evidence matrix

| AC | Status | Verification | Result |
|----|--------|--------------|--------|
| AC-SBN-001 | PASS | `grep -rln agentless_audit_test $SK/` | 0 files |
| AC-SBN-002 [MUST-PASS] | PASS | `go test -run 'Sentinel\|AgentlessControlFlow\|LoopAlias'` | ok; sentinels retained |
| AC-SBN-003 | PASS | `grep MODE_* lines with REQ-WF` | 0 |
| AC-SBN-004 | PASS | GOOS+cmd/moai=0; golangci@v2.1.6=0; placeholder=6 | PASS |
| AC-SBN-005 | PASS | `grep internal/(spec\|cli\|hook\|ciwatch\|design)/...go` | 0 |
| AC-SBN-006 | PASS | SPEC-V3R/CONST=0; named IDs=0 | PASS |
| AC-SBN-007 | PASS | `grep REQ/AC tokens` | 0 |
| AC-SBN-008 | PASS | 99-release=0; 97-release-update=0; NOT-distributed=0 | PASS |
| AC-SBN-009 | PASS | doctrine=0; 2026-05-26=0; namespace present=21 | PASS |
| AC-SBN-010 | PASS | SPEC-AUTH-001=14; illustrative paths present | preserved |
| AC-SBN-011 | PASS | `diff` template vs mirror, 50 files | byte-identical |
| AC-SBN-012 (RED) | PASS | M1 RED: test FAILED, C6≥6 files | 185 occurrences |
| AC-SBN-013 | PASS | C7 class pkg-restricted in test file | present |
| AC-SBN-014 | PASS | C1b matches V3R[0-9] in test file | present |
| AC-SBN-015 | PASS | allowlist + GREEN despite placeholders present | PASS |
| AC-SBN-016 [MUST-PASS] | PASS | `go test -run TestTemplateNoInternalContentLeak` | ok (GREEN) |
| AC-SBN-017 | PASS | `TestSkillBodyLeakClassRecurrenceBackstop` | PASS |
| AC-SBN-018 | PASS | `TestLeakClassReqTokenPartition` + NoDateSha | PASS |
| AC-SBN-019 | PASS | 4-locale annotation=0; adk URL=3 files | PASS |
| AC-SBN-020 | PASS | `TestC7PackageRestriction` | PASS |

## §E.5 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-23
sync_commit_sha: 2e7925e3b
sync_status: completed
changelog_entry: added (Unreleased / Changed)
mirror_byte_identical: true (50 files; foundation-cc/core diff = pre-existing drift, out of scope)
guard_green: true (TestTemplateNoInternalContentLeak PASS, ok 0.863s)
leak_classes_zero: { agentless: 0, spec_v3r: 0, req_token: 0, go_path: 0, four_locale: 0 }
keep_list_preserved: { spec_auth_001: 14, adk_url: 3, go_install_moai: 2 }
three_phase_close: true (in-progress -> completed merged into this sync commit)
```
