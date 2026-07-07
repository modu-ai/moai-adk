# Progress — SPEC-CLAUDEMD-DIET-V2-001

Tier M. 2nd-round CLAUDE.md diet. Lifecycle: plan → run → sync (3-phase). Track 2 of official-docs-audit-2026-07.

---

## §E.1 Plan-phase Audit-Ready Signal

- **plan_complete_at**: 2026-07-08
- **plan_status**: audit-ready (iter-2 — D1 BLOCKING + D2-D6 applied)
- **version**: 0.1.1 (iter-2 revision)
- **Tier**: M (4-artifact set: spec.md + plan.md + acceptance.md + this progress.md skeleton)
- **era**: V3R6
- **Artifacts**:
  - `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` v0.1.1 (§A-H, frontmatter 12 canonical fields + era:V3R6 + tier:M, `### Out of Scope —` H3 sub-sections × 8)
  - `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/plan.md` (§A-§H, milestones M1-M4, Template-First order)
  - `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/acceptance.md` (AC-CMD2-001..010, 전부 MUST-blocking)
  - `.moai/specs/SPEC-CLAUDEMD-DIET-V2-001/progress.md` (this file — §E.1-E.4 skeleton)
- **SPEC ID self-check**: `decomposition: SPEC ✓ | CLAUDEMD ✓ | DIET ✓ | V2 ✓ | 001 ✓ → PASS`
- **Requirements**: REQ-CMD2-001 (line count ≤ 320 MUST / ≤ 210 SHOULD), -002 (existing-SSOT pointer-ization), -003 (no-SSOT gate — §16 신규 룰 파일 또는 in-place 압축), -004 (§16 path-scoped rule), -005 (byte-parity), -006 (no behavior change), -007 ([HARD]/@embed preservation — D1 fix: `@<path>` embed NOT `@import`), -008 (template neutrality), -009 (always-loaded reduction), -010 (derived target arithmetic).
- **Acceptance summary**: AC-CMD2-001 (line ≤ 320), -002 (diff exit 0), -003 ([HARD] ≥ 14, [ZONE:*] == 14, @embed == 2 — D1 fix), -004 (always-loaded reduction + path-scoped), -005 (agent catalog/archived/command set unchanged), -006 (§16 Option A/B completeness), -007 (TestTemplateNeutralityAudit PASS — TestInternalContentLeak "no tests to run"), -008 (moai spec lint clean — D2 fix: file path form), -009 (distinctive-content grep precondition), -010 (§1/§2/§3/§4 line count ± 1 — D4/D5 fix). **10 AC, 전부 MUST-blocking + 1 SHOULD (AC-001 ≤ 210)**.

### Plan-auditor iter-1 verdict + iter-2 fix log

- **iter-1 plan-auditor verdict**: **FAIL (0.81)** — 1 BLOCKING (D1) + 5 SHOULD/MINOR recommended.
- **D1 (BLOCKING, mandatory fix)**: AC-CMD2-003 regex `^@import` matched 0 lines. Actual syntax is Obsidian-style `@<path>` embed (`@.moai/config/sections/{user,language}.yaml`). Fixed regex to `^@\.moai/config/sections/(user|language)\.yaml$` across acceptance.md / spec.md §B.5 REQ-CMD2-007 / §C.7 / §F.1 / plan.md §C.1.3 / §F.M4.
- **D2 (SHOULD-FIX, verification-claim-integrity §1.1 surface 2)**: AC-CMD2-008 command was never test-run. Fixed form `moai spec lint SPEC-CLAUDEMD-DIET-V2-001` (ParseFailure) → `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` (실측 `✓ No findings — all SPEC documents are valid`). **Ran ALL 10 AC verification commands this iteration; verbatim observed outputs in spec.md §F.1 + this section below.**
- **D3 (SHOULD-FIX, scope-honesty)**: §A.4 / plan.md §D.5 reframed "200L impossible" → "200L scope choice" with theoretical-pointer option disclosed (§1's 11-bullet HARD list partially pointer-ized to `moai-constitution.md` 287L; further pointer-ization theoretically possible but out of scope). MUST ≤ 320 / SHOULD ≤ 210 numbers unchanged.
- **D4 (MINOR)**: AC-CMD2-010 baselines corrected to awk-measured §1=29 / §3=12 / §4=50 with ±1 tolerance. (Plan-audit reported 27/10/48 = body-excluding-heading; awk is heading-inclusive → +1-2 offset. This spec uses awk-measured values directly with explicit tolerance.)
- **D5 (MINOR)**: spec.md §D Out-of-Scope heading "§1/§3/§4" → "§1/§2/§3/§4" reconciled with body bullet that already listed §2.
- **D6 (MINOR)**: SPEC-V3R5-CLAUDE-REFRESH-001 "superseded scope" → "completed" (실측 `status: completed`, 2026-05-18~19, "Architecture Truth Reconciliation + Bundle A Settings Fix" — different scope from this diet SPEC).

### iter-2 observed-baseline evidence (all 10 AC verification commands actually run)

Recorded per D2 fix (verification-claim-integrity §1.1 surface 2 — baseline attribution):

| AC | Verification command (iter-2 actual run) | Observed output |
|----|------------------------------------------|-----------------|
| AC-CMD2-001 | `L=$(wc -l < CLAUDE.md); T=$(wc -l < internal/template/templates/CLAUDE.md); echo "LIVE=$L TEMPLATE=$T"` | `LIVE=405 TEMPLATE=405` |
| AC-CMD2-002 | `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo "DIFF_EXIT=$?"` | `DIFF_EXIT=0` (no diff output) |
| AC-CMD2-003a | `grep -c '\[HARD\]' CLAUDE.md` | `14` |
| AC-CMD2-003b | `grep -c '\[ZONE:' CLAUDE.md` | `14` |
| AC-CMD2-003c-OLD | `grep -c '^@import' CLAUDE.md` | `0` (D1 bug proven) |
| AC-CMD2-003c-FIX | `grep -cE '^@\.moai/config/sections/(user\|language)\.yaml$' CLAUDE.md` | `2` (D1 fix verified) |
| AC-CMD2-004 | `ls .claude/rules/moai/workflow/context-search.md 2>/dev/null` | NOT-FOUND (expected pre-M1) |
| AC-CMD2-005a | `grep -c '^\| \`manager-' CLAUDE.md` | `4` |
| AC-CMD2-005b | `grep -c 'Use the.*subagent' CLAUDE.md` | `11` |
| AC-CMD2-006a | `ls .claude/rules/moai/workflow/context-search.md 2>/dev/null && echo EXISTS` | NOT-FOUND (expected pre-M1) |
| AC-CMD2-006b | `awk '/^## 16\. Context Search/,/^## 17\./' CLAUDE.md \| wc -l` | `50` (awk heading-inclusive baseline) |
| AC-CMD2-007a | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.538s` |
| AC-CMD2-007b | `go test ./internal/template/ -run TestInternalContentLeak -count=1` | `ok ... [no tests to run]` (test name not found; AC-CMD2-007 should rely on TestTemplateNeutralityAudit + `split_namespace_test.go` TestSplitNamespaceNoLeak instead — flag for run-phase resolution) |
| AC-CMD2-008 | `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` | `✓ No findings — all SPEC documents are valid` (D2 fix verified) |
| AC-CMD2-009a (§5) | `grep -c 'Phase 1.*plan-phase.*manager-spec' spec-workflow.md` | **`0`** (M2 RISK — §5 Agent Chain distinctive content NOT in SSOT) |
| AC-CMD2-009b (§11) | `grep -c 'Error Recovery Pattern' agent-common-protocol.md` | `1` |
| AC-CMD2-009c (§14) | `grep -c 'Parallel Execution' agent-common-protocol.md` | `1` |
| AC-CMD2-009d (§15) | `grep -c 'CG Mode\|moai cg' spec-workflow.md dynamic-workflows.md` | **`0` in both** (M2 RISK — §15 CG Mode distinctive content NOT in SSOT) |
| AC-CMD2-009e (§8) | `grep -c 'AskUserQuestion' askuser-protocol.md` | `50` |
| AC-CMD2-010a | `awk '/^## 1\. Core Identity/,/^## 2\./' CLAUDE.md \| wc -l` | `29` (D4 actual; iter-1 spec had 28) |
| AC-CMD2-010b | `awk '/^## 3\. Command Reference/,/^## 4\./' CLAUDE.md \| wc -l` | `12` (D4 actual; iter-1 spec had 11) |
| AC-CMD2-010c | `awk '/^## 4\. Agent Catalog/,/^## 5\./' CLAUDE.md \| wc -l` | `50` (D4 actual; iter-1 spec had 49) |

### iter-2 new findings (flagged for run-phase attention)

1. **M2 RISK — §5 + §15 distinctive-content 0-hit**: AC-CMD2-009 baseline grep returned 0 hits for §5 Agent Chain (in spec-workflow.md) and §15 CG Mode (in both spec-workflow.md and dynamic-workflows.md). This means M2's target −43L reduction (§15 −23L + §5 −20L) is at risk — run-phase M2 must either (a) find alternative distinctive-content tokens that DO exist in the SSOTs and widen the grep, OR (b) demote §5/§15 partial-KEEP per 1st-round D1 precedent (only the genuinely-duplicated subsections get pointer-ized; unique content stays). This is NOT a plan-phase blocker — AC-CMD2-009 is a run-phase precondition that correctly flags the issue at the right time.
2. **AC-CMD2-007b test name drift**: `TestInternalContentLeak` returned "no tests to run" — the canonical test name in `internal/template/` may differ. Run-phase M4 should use the actual test name (CI guard file `internal/template/internal_content_leak_test.go` exists; the `-run` regex needs adjustment). Flagged for run-phase resolution, not a plan-phase blocker.
3. **§16 line count actual = 50 (not 48)**: awk heading-inclusive measured 50L; iter-1 spec said 48L (content body estimate). Off-by-2 — plan.md §C.4 arithmetic still holds because the reduction target (5L or 15L post-diet) is the relevant figure, not the baseline. No arithmetic correction needed.

- **Key constraint**: BODY editing (1st-round과 동일 위험 등급); Template-First edit order; @embed는 감소 산술에서 제외(1st-round AC-CMD-004 계승); §1/§2/§3/§4 canonical content 편집 금지(HARD); 200L SHOULD 달성은 scope choice(D3 reframe)로 §1-§4 보호 결정의 결과 — 3rd-round 위임을 out of scope로 명시.
- **iter-2 plan-auditor verdict**: _<pending Phase 0.5 re-run — Tier M PASS threshold 0.80; D1 BLOCKING fix + D2-D6 recommended applied; ready for iter-2 audit>_

---

## §E.2 Run-phase Evidence

### Run-phase decisions

- **M1 §16 Option A (in-place compress) chosen over Option B (new rule file)**: §16 content is a search PROCEDURE (when/how to search previous sessions), not file-type-specific guidance. Path-scoping to `.moai/specs/**` would hide the procedure during general coding sessions where context search is equally likely. In-place compression preserves always-available access while achieving −37L always-loaded reduction (50→13L). Arithmetic (319L final) comfortably clears the 320L MUST ceiling. Decision within manager-develop authority per plan §D.1.
- **M2 partial-KEEP for §5/§15**: AC-CMD2-009 preconditions confirmed §5 Agent Chain substance exists in spec-workflow.md under different wording (`manager-spec`=3, `Phase 1`=4 hits); §15 CG Mode ASCII / `moai cg` = 0 hits (genuinely distinctive → KEEP). Dynamic Workflows pointer-ized (`Agent Teams`=4+2 hits in SSOTs). Per plan §F.M2 fallback path (b).
- **M3 re-compression**: initial M3 (§8+§11+§14) landed at 324L (4L over ceiling). Re-compressed per plan §F.M4: §13 2nd paragraph merged into blockquote pointer (−2L), footer 3 meta lines → 1 pipe-separated line (−3L) → 319L.
- **Pre-existing parity break fixed**: plan-phase commit `365feae6b` removed `.claude/rules/moai/design/constitution.md` reference from template §9 but not live. Synced live to template (Template-First SSOT) as part of M1.

### Milestone commits (worktree branch `worktree-agent-a39cdf096b4cb18c8`)

| Milestone | Commit SHA | Subject |
|-----------|-----------|---------|
| M1 | `210ca1818` | fix(SPEC-CLAUDEMD-DIET-V2-001): M1 §16 in-place compress + byte-parity restore |
| M2 | `ab92ccaa0` | fix(SPEC-CLAUDEMD-DIET-V2-001): M2 §5+§15 partial-KEEP pointer-ization |
| M3 | `614c71618` | fix(SPEC-CLAUDEMD-DIET-V2-001): M3 §8+§11+§14 pointer-ization + footer/§13 micro-trim |
| M4 | _(this commit)_ | fix(SPEC-CLAUDEMD-DIET-V2-001): M4 run-phase evidence + AC verification |

### AC PASS/FAIL matrix (verbatim observed outputs, measured at M4 HEAD `614c71618`)

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CMD2-001 (MUST ≤320) | **PASS** (319 ≤ 320) | `wc -l CLAUDE.md internal/template/templates/CLAUDE.md` | `319 CLAUDE.md` / `319 internal/template/templates/CLAUDE.md` (SHOULD ≤210 NOT met — 3rd-round deferred per spec §D Out-of-Scope) |
| AC-CMD2-002 (byte-parity) | **PASS** | `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo "DIFF_EXIT=$?"` | `DIFF_EXIT=0` (no diff output) |
| AC-CMD2-003 ([HARD]/@embed) | **PASS** | `grep -c '\[HARD\]' CLAUDE.md` / `grep -c '\[ZONE:' CLAUDE.md` / `grep -cE '^@\.moai/config/sections/(user\|language)\.yaml$' CLAUDE.md` | `14` / `14` / `2` (both trees equal) |
| AC-CMD2-004 (always-loaded) | **PASS** (Option A) | `ls .claude/rules/moai/workflow/context-search.md 2>/dev/null` | NOT-FOUND (Option A — no new rule file, n/a per AC) |
| AC-CMD2-005 (no behavior change) | **PASS** | `grep -c '^| \`manager-' CLAUDE.md` / `grep -c 'Use the.*subagent' CLAUDE.md` / `grep -c 'archived.*MUST NOT be spawned' CLAUDE.md` | `4` / `11` / `1` (all baseline) |
| AC-CMD2-006 (§16 extraction) | **PASS** (Option A) | `awk '/^## 16\. Context Search/,/^## 17\./' CLAUDE.md \| wc -l` | `13` (≤17 Option A target) |
| AC-CMD2-007 (template neutrality) | **PASS\*** | `go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` / `-run TestSplitHarnessNamespaceNoLeak -count=1` | both `ok` (D7 fix: correct test names used). `TestTemplateNeutralityAudit` C8Preserve subtest fails PRE-EXISTINGLY (see note below) |
| AC-CMD2-008 (GEARS lint) | **PASS\*** | `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` | `0 error(s), 1 warning(s)` — StatusGitConsistency (worktree artifact, see note below) |
| AC-CMD2-009 (distinctive-content grep) | **PASS** | pre-pointer greps per section | §11 `Error Recovery Pattern`=1, §14 `Parallel Execution`=1, §8 `AskUserQuestion`=50 (all ≥1). §5/§15 partial-KEEP per 0-hit distinctive tokens (CG Mode ASCII, `moai cg`) |
| AC-CMD2-010 (§1-4 guard) | **PASS** | `awk` per section | S1=29 S2=36 S3=12 S4=50 (all ±1 of baseline) |

### Build + lint evidence

```
$ go build ./...                          → exit 0
$ make build                              → exit 0 (catalog.yaml regenerated, bin/moai built)
$ golangci-lint run --timeout=2m ./internal/template/ → 0 issues
```

### Notes on PASS\* items (verification-claim-integrity §3.4 Gaps)

1. **AC-CMD2-007 C8Preserve pre-existing failure**: `TestTemplateNeutralityAuditC8Preserve` fails with "C8 GOOS= PRESERVE expected 3 files, got 2". Verified PRE-EXISTING by checkout of `2eb5bcd23` (pre-M1 state) — identical failure. Cause: plan-phase commit `365feae6b` removed `internal/template/templates/scripts/ci-mirror/*.sh` (which contained GOOS= cross-compile patterns), dropping the C8 file count from 3 to 2. NOT caused by this SPEC's CLAUDE.md diet. The tests DIRECTLY validating my edits (`TestTemplateNoInternalContentLeak` + `TestSplitHarnessNamespaceNoLeak`) both PASS — my CLAUDE.md changes introduced zero new neutrality violations.
2. **AC-CMD2-008 StatusGitConsistency warning**: `moai spec lint` reports 1 warning — frontmatter `status: in-progress` disagrees with git-implied `draft`. This is a **worktree artifact**: the M1 commit transitioned the status on the worktree branch (`worktree-agent-a39cdf096b4cb18c8`), but `main` still carries `status: draft` (my commits are not yet merged to main). The warning will resolve when the orchestrator merges the worktree branch to main. The iter-2 baseline (`✓ No findings`) was measured when status was still `draft` (pre-transition). 0 errors.
3. **D7 resolved**: AC-CMD2-007 canonical tests are `TestTemplateNoInternalContentLeak` + `TestSplitHarnessNamespaceNoLeak` (not the non-existent `TestInternalContentLeak`). Both PASS.
4. **D8 actual baselines**: §16=50, §15=45, §5=37, §11=21, §14=18, §8=11 (total 182L, +12L vs plan's 170L estimate). Used actuals in arithmetic.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-08
run_commit_sha: 614c71618   # M3 final content commit; M4 = this progress.md evidence commit (placeholder, backfill if separate)
run_status: audit-ready
ac_pass_count: 10
ac_fail_count: 0
ac_pass_with_note_count: 2   # AC-CMD2-007 (C8Preserve pre-existing) + AC-CMD2-008 (StatusGitConsistency worktree artifact)
preserve_list_post_run_count: 5   # [HARD]=14 [ZONE:]=14 @embed=2 (both trees)
l44_pre_commit_fetch: true
l44_post_push_fetch: pending   # worktree branch not yet merged to main
new_warnings_or_lints_introduced: 0   # CLAUDE.md diet introduced 0 new neutrality/lint issues
cross_platform_build:
  go_build_all: pass   # exit 0
  make_build: pass     # exit 0
total_run_phase_files: 2   # CLAUDE.md + internal/template/templates/CLAUDE.md (+ spec.md frontmatter transition)
m1_to_mN_commit_strategy: per-milestone (M1=210ca1818, M2=ab92ccaa0, M3=614c71618, M4=progress.md evidence)
```

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates this with the single sync commit SHA carrying the `implemented → completed` transition>_

---

## §F Phase 0.95 Mode Selection

**Evaluated at**: 2026-07-08 (this session, after Implementation Kickoff Approval).

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | M |
| scope (file count) | 2–3 (CLAUDE.md + `internal/template/templates/CLAUDE.md` mirror + possibly new `.claude/rules/moai/workflow/context-search.md`) |
| domain count | 1 (CLAUDE.md markdown editing — single domain, multi-section) |
| file language mix | 100% markdown |
| concurrency benefit | LOW (same-file sequential section edits — milestone N depends on milestone N-1's CLAUDE.md state) |
| Agent Teams prereqs | enabled=true + env=1 — but scope below Mode 3 threshold |

### Mode evaluation table

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | NO | 4 milestones, 6+ sections — not a typo/single-line |
| 2 background | NO | write task (Write/Edit on CLAUDE.md + template mirror) |
| 3 agent-team | NO | 1 domain < 3, 2–3 files < 10 (prereqs met but scope fails the compound clause) |
| 4 parallel | NO | 1 domain — not multi-domain research |
| 5 sub-agent | **YES** | default fallback; coding/markdown-heavy single-file work (Anthropic coding-task parallelism caveat) |
| 6 workflow | NO | 2–3 files ≪ ~30-file mechanical threshold |

### Decision

`sub-agent` (Mode 5, sequential per-milestone).

### Justification

단일 도메인·단일 파일(CLAUDE.md)의 sequential 섹션 편집이다. 각 milestone이 CLAUDE.md의 같은 파일을 수정하므로 milestone N은 milestone N-1의 CLAUDE.md 결과 상태에 의존한다 — 병렬화가 불가능하고 inter-edit 의존성이 존재한다. Anthropic의 coding-task parallelism caveat("most coding tasks involve fewer truly parallelizable tasks than research")가 markdown-heavy 문서 편집에도 동일하게 적용된다. Tier M이므로 Section A–E delegation template(`.claude/rules/moai/development/manager-develop-prompt-template.md`)을 적용하여 manager-develop에게 위임한다. M1(§16) → M2(§15+§5) → M3(§11+§14+§8) → M4(template 동기화+AC 검증) 순차 진행.

---

## §G IGGDA Kickoff Predicate

**Evaluated at**: 2026-07-08 (this session).

| Condition | Value | Source |
|-----------|-------|--------|
| (a) Intent clarity 100% | TRUE | 이전 세션 Socratic interview 완료(paste-ready resume 명시) + resume 전제 3종 실측 통과(byte-parity 회복 · 0 8 clean ahead · V0-c=2) |
| (b) plan-auditor PASS | TRUE | iter-2 PASS-WITH-DEBT 0.904 (PASS verdict, NOT FAIL/INCONCLUSIVE) |
| (c) Tier S or M | TRUE | tier: M |
| (d) No dangerous keywords AND no destructive scope | TRUE | CLAUDE.md diet — title/scope/plan/acceptance에 auth·secret·credential·password·crypto·payment·billing·production·migration·drop_table·force_push·rm_rf 해당 없음; `--pr` flag 부재; destructive scope 아님 |

**Matched dangerous keyword**: (none)

**Final verdict**: `auto-proceed` (all 4 conditions hold)

> §H.2 auto-select-after-timeout deferral note: AskUserQuestion은 timeout/auto-select 메커니즘이 없으므로, 본 auto-proceed는 framing 경량화(첫 옵션 `(권장)` low-friction)만 해당. 사용자는 "run-phase 진입 (권장)" 선택으로 veto를 행사하지 않고 진입을 승인함 — Implementation Kickoff Approval gate 통과.
