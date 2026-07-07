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

_<pending run-phase — manager-develop populates this section with verbatim command outputs, M1-M4 milestone evidence, and the AC PASS/FAIL matrix>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop populates this on M4 completion>_

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
