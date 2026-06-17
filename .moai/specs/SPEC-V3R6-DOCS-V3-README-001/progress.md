# progress.md — SPEC-V3R6-DOCS-V3-README-001

> Run-phase / sync-phase / Mx-phase evidence skeleton. Plan-phase에서는 §E.1만 채우고 §E.2-§E.5는 placeholder heading만 둔다.

---

## §A. SPEC Status

- **ID**: SPEC-V3R6-DOCS-V3-README-001
- **Tier**: M (standard)
- **Status**: in-progress (run-phase M1-M4 complete, M5-M6 in progress)
- **Created**: 2026-06-17
- **Plan-phase artifacts commit**: 03ff915ed (M1 commit carried plan-phase artifacts)
- **Run-phase base**: origin/main 109c1c0d0 (synced 0 0; 1 unrelated docs-site book-migration commit after 4a6f4b4d3)

---

## §B. Milestone Tracker

| Milestone | Scope | Status | Commit |
|-----------|-------|--------|--------|
| M1 | Agent catalog + tier-table rewrite (en) | complete | 03ff915ed |
| M2 | GLM tier-model 정정 (en) | complete | 3ed266f0f |
| M3 | `/moai` 17-command + "47 Skills" 헤더 제거 (en) | complete | 50f22e261 |
| M4 | README.ko.md 동기화 (ko) | complete | 8a3108e41 |
| M5 | statusline 보존 + scope boundary 확인 | complete | (this commit) |
| M6 | en/ko cross-check + 최종 AC 검증 | complete | (this commit) |

---

## §C. AC Tracker

| AC | Severity | Status | Evidence |
|----|----------|--------|----------|
| AC-1 (agent catalog en) | MUST | PASS | grep "27 agents"=0; archived names only in migration-ref prose (Edge-1 allowed); tier table rows = retained 8 only; "8 retained"×3 |
| AC-2 (agent catalog ko) | MUST | PASS | grep "24개/26개 에이전트"=0; "52개/47개 스킬"=0; "Agency \| 6" row=0; ko tier rows = retained 8 only; "retained 에이전트"×6 |
| AC-3 (command set en+ko) | MUST | PASS | en "47 Skills"=0; ko "47개 스킬"=0; en 17-command mention=1; ko 17-command mention=4; en /moai plan\|run\|sync=17 |
| AC-4 (GLM tier en) | MUST | PASS | glm-5.2[1m]×3; GLM-5.1=0; glm-4.7×2; glm-4.5-air×2 |
| AC-5 (GLM tier ko) | MUST | PASS | glm-5.2[1m]×3; GLM-5.1=0; glm-4.7+glm-4.5-air present |
| AC-6 (en/ko sync) | MUST | PASS | both "8 retained"; both glm-5.2[1m]; both 17 commands |
| AC-7 (statusline 보존) | MUST | PASS | en preset-retire L1271 + multi-line L1231; ko preset-폐기 L1340 + 멀티라인 L1300 |
| AC-8 (scope boundary) | MUST | PASS | my commits (109c1c0d0..HEAD): .go=0, docs-site=0, CLAUDE.md=0, template=0, README×2 |

---

## §D. Pre-flight Verification (plan-phase)

- [x] docs-truth.md 존재 (122L, commit 4a6f4b4d3)
- [x] §1 1차 소스 재검증: `ls .claude/agents/moai/*.md` → 7 files ✓
- [x] §2 1차 소스 재검증: `internal/spec/status.go` ValidStatuses → 8 values ✓
- [x] §3 1차 소스 재검증: `internal/spec/lint.go` required slice → 12 entries ✓
- [x] §4.2 1차 소스 재검증: `ls .claude/commands/moai/*.md` → 17 files ✓
- [x] §5 1차 소스 재검증: `internal/config/defaults.go` DefaultGLMHigh → `glm-5.2[1m]` ✓
- [x] README.md (1370L) + README.ko.md (1418L) 존재 ✓
- [x] drift inventory 17항 완료 (rewrite 13 + info-only 4축)

---

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts**: 4 files authored (spec.md + plan.md + acceptance.md + progress.md)
- **Tier**: M (standard)
- **Drift inventory**: 17 items across 7 axes (§1 agent catalog, §4.2 command set, §5 GLM tier = rewrite; §2/§3/§6/§7 = info-only/preservation)
- **1차 소스 재검증**: 전수 PASS at commit 4a6f4b4d3 — blocker 없음
- **Milestones**: M1..M6 (6 milestones, 파일-단위)
- **AC**: 8 AC (AC-1..AC-8), 전수 grep/diff 기반 기계적 검증 가능
- **Scope boundary**: README.md + README.ko.md 2개 파일만; docs-site / Go / CLAUDE.md / template EXCLUDE
- **Anti-overengineering**: 사실 정정(reconciliation)만; 추상화/설정 시스템/미래 확장 hook 금지
- **Plan-phase readiness**: audit-ready (plan-auditor verdict: PASS-WITH-DEBT 0.86, zero BLOCKING)

---

## §E.2 Run-phase Evidence

### M1 — Agent catalog + tier-table rewrite (en) — commit 03ff915ed

| AC | Verification Command | Actual Output | Status |
|----|---------------------|---------------|--------|
| AC-1(a) | `grep -c "27 agents" README.md` | 0 (exit 1) | PASS |
| AC-1(b) | archived-name grep (categories context) | 4 matches, all in migration-reference prose (Edge-1 allowed) | PASS |
| AC-1(b2) | `grep -cF 'Design System** \| 4 (+ evaluator)'` | 0 | PASS |
| AC-1(c) | `grep -iEc "8 retained\|8 agents\|retained agents"` | 3 | PASS |
| AC-1(d) | tier-table ROWS archived grep | 0 archived table rows (retained 8 only) | PASS |

Also fixed: README.md L326 "24 agents" → "8 retained agents" (stale count surface in Model Policy intro).

### M2 — GLM tier-model 정정 (en) — commit 3ed266f0f

| AC | Verification Command | Actual Output | Status |
|----|---------------------|---------------|--------|
| AC-4(a) | `grep -c 'glm-5\.2\[1m\]' README.md` | 3 | PASS |
| AC-4(b) | `grep -nE 'GLM-5\.1' README.md` | exit 1 (no match) | PASS |
| AC-4(c) | `grep -c 'glm-4\.7' README.md` | 2 | PASS |
| AC-4(d) | `grep -ci 'glm-4\.5-air' README.md` | 2 | PASS |

Source: `internal/config/defaults.go` `DefaultGLMHigh = "glm-5.2[1m]"`.

### M3 — /moai 17-command set + "47 Skills" 헤더 제거 (en) — commit 50f22e261

| AC | Verification Command | Actual Output | Status |
|----|---------------------|---------------|--------|
| AC-3(a) | `grep -c "47 Skills" README.md` | 0 (exit 1) | PASS |
| AC-3(c) | `grep -iEc "17 (commands\|slash\|/moai)\|/moai.*17"` | 1 | PASS |
| AC-3(e) | `grep -cE "/moai plan\|/moai run\|/moai sync"` | 17 | PASS |

Replaced "### 47 Skills" header + 13-row stale count table with "### /moai Slash Commands (17)" listing. Progressive Disclosure 3-level system description preserved.

### M4 — README.ko.md 동기화 (ko) — commit 8a3108e41

| AC | Verification Command | Actual Output | Status |
|----|---------------------|---------------|--------|
| AC-2(a) | `grep -nE "24개.*에이전트\|26개.*에이전트"` | exit 1 (no match) | PASS |
| AC-2(b) | `grep -nE "52개 스킬\|47개 스킬"` | exit 1 (no match) | PASS |
| AC-2(b2) | `grep -cF '**Agency** \| 6'` | 0 | PASS |
| AC-2(c) | `grep -cE "8개.*retained\|8 retained\|retained 에이전트"` | 6 | PASS |
| AC-2(d) | ko tier ROWS archived grep | 0 archived table rows (retained 8 only, expert-testing removed) | PASS |
| AC-3(b) | `grep -c "47개 스킬"` | 0 (exit 1) | PASS |
| AC-3(d) | `grep -cE "17.*(명령\|commands\|/moai)\|/moai.*17"` | 4 | PASS |
| AC-5(a) | `grep -c 'glm-5\.2\[1m\]'` | 3 | PASS |
| AC-5(b) | `grep -nE 'GLM-5\.1'` | exit 1 (no match) | PASS |

5 stale ko count surfaces fixed: L40 ("24개+52개"), L110 ("26개+47개"), L308 ("24개 전문 에이전트"), L372 ("24개 에이전트"), L334 category table. ko GLM L709/L715 → glm-5.2[1m] + [1m] suffix note.

### M5 — statusline 보존 + scope boundary (verification-only) — this commit

| AC | Verification Command | Actual Output | Status |
|----|---------------------|---------------|--------|
| AC-7(a) | en preset retire grep | L1271 match | PASS |
| AC-7(b) | ko preset retire grep | L1340 match | PASS |
| AC-7(c) | en statusline v3 multi-line grep | L1231 match | PASS |
| AC-7(d) | ko statusline v3 grep | L1300 match | PASS |
| AC-8(a) | `git diff --stat 109c1c0d0..HEAD -- '*.go'` | 0 lines | PASS |
| AC-8(b) | `git diff --stat 109c1c0d0..HEAD -- 'docs-site/'` | 0 lines | PASS |
| AC-8(c) | `git diff --stat 109c1c0d0..HEAD -- 'CLAUDE.md'` | 0 lines | PASS |
| AC-8(d) | `git diff --stat 109c1c0d0..HEAD -- 'internal/template/templates/'` | 0 lines | PASS |
| AC-8(e) | `git diff --stat 109c1c0d0..HEAD -- 'README.md' 'README.ko.md'` | 2 files, 83 insertions, 105 deletions | PASS |

**AC-8 baseline note**: SPEC acceptance.md hardcodes `4a6f4b4d3..HEAD`. Measured against that baseline, `docs-site/` shows 11 changed lines — but those originate from the unrelated intermediate commit `109c1c0d0` (docs-site book → book.mo.ai.kr migration), NOT from this SPEC's commits. Measured against `109c1c0d0..HEAD` (this SPEC's 4 commits + this M5/M6 progress commit), `docs-site/` = 0. The substantive scope boundary holds: this SPEC touched only README.md + README.ko.md + SPEC artifacts.

### M6 — en/ko cross-check + 최종 AC 검증 — this commit

| Fact | en value | ko value | Parity |
|------|----------|----------|--------|
| Retained agent count | 8 retained | 8 retained (8개 retained) | PASS |
| GLM Opus model | glm-5.2[1m] | glm-5.2[1m] | PASS |
| /moai command count | 17 | 17 (17개) | PASS |
| "47 Skills"/"47개 스킬" header | absent | absent | PASS |

All 8 AC PASS. See §C AC Tracker above.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-17
run_commit_sha: 8a3108e41  # M4 final source-edit commit; M5/M6 are progress-only
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 2  # README.md + README.ko.md only
l44_pre_commit_fetch: "origin/main synced 0 0 at 109c1c0d0 pre-run"
l44_post_push_fetch: "pending push"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_na: true  # doc-only SPEC, Go LOC change = 0
total_run_phase_files: 2  # README.md + README.ko.md
m1_to_mN_commit_strategy: "6 commits (M1 03ff915ed + M2 3ed266f0f + M3 50f22e261 + M4 8a3108e41 + M5/M6 progress)"
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-17
sync_commit_sha: 5d0055adb0
sync_status: complete
changelog_entry: true  # CHANGELOG.md [Unreleased] -> Added
frontmatter_status: implemented  # in-progress -> implemented
readme_files_synced: 2  # README.md + README.ko.md
en_ko_fact_parity: true  # 8 retained / glm-5.2[1m] / 17 commands
sync_method: orchestrator-direct  # GLM manager-docs spawn context-limit fallback
spec_lint_post_sync: "No findings"
```

---

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: 2026-06-17
mx_commit_sha: <pending-backfill>
mx_status: complete
final_status: completed  # implemented -> completed
spec_lint_final: "No findings"
ac_final: "8/8 PASS"
go_change_loc: 0
harness_level_effective: standard
mx_method: orchestrator-direct  # GLM sync-auditor spawn context-limit fallback
self_assessment:
  functionality: PASS  # 8/8 AC, README 7-axis reconciliation complete
  security: N/A  # doc-only, no Go change, no input boundary
  craft: PASS  # en/ko fact-for-fact parity, surgical edits, no over-engineering
  consistency: PASS  # docs-truth baseline adhered, CODEMAPS cohort successor consistent
era: V3R6
four_phase_close: true  # plan + run + sync + Mx
```

---

## §F. Gaps (forward-looking findings, NOT blockers)

1. **Mermaid architecture diagram stale count nodes** — en README.md L268-272 (`Manager (8)`, `Expert (8)`, `Builder (3)`, `Evaluator (2)`, `Design System (4+1)`) and ko README.ko.md L318 (`Agency (6)`) are architecture-diagram node labels showing the old 27-agent / Agency-6 structure. These are NOT category-table rows or tier-mapping rows (the SPEC's declared drift inventory §C.1 scope), so they were NOT fixed in this SPEC per the anti-overengineering directive. Symmetric across en/ko (both stale), so AC-6 en/ko parity is not violated. Candidate for a follow-up SPEC.

2. **Design-workflow prose references to archived agents** — en README.md L921 (`manager-quality`), L924 (`20 agents`), L942/L944/L972 (`expert-frontend`) describe archived agents as active participants in the `/moai design` pipeline. These are descriptive prose, not category/tier rows, and fall outside the SPEC's declared drift inventory scope (§C.1 enumerates categories table + tier table only). NOT fixed in this SPEC. Candidate for a follow-up SPEC.

3. **ko free-model note case** — README.ko.md L721 `GLM-4.7-Flash, GLM-4.5-Flash` retain uppercase. These are z.ai free-tier product names (distinct from the `glm-4.5-air` tier-model), so they legitimately stay as-is. NOT a drift.

---

## HISTORY

- 2026-06-17: plan-phase artifacts authored (4 files). §E.1 채움, §E.2-§E.5 placeholder heading만. drift inventory 17항, 1차 소스 전수 재검증 PASS.
- 2026-06-17 (run-phase): M1-M4 complete (commits 03ff915ed, 3ed266f0f, 50f22e261, 8a3108e41). §E.2 Run-phase Evidence + §E.3 Run-phase Audit-Ready Signal 채움. 8/8 AC PASS. 2 gaps recorded (Mermaid diagram nodes + design-workflow prose — both outside SPEC declared scope, symmetric en/ko). M5 (statusline preservation + scope boundary) + M6 (en/ko cross-check + final AC verification) verification recorded.
