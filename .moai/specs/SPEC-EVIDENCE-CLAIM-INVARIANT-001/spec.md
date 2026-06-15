---
id: SPEC-EVIDENCE-CLAIM-INVARIANT-001
title: "검증 주장 무결성 doctrine — no unobserved-verification-claim invariant + baseline 귀속 + 5-섹션 증거 리포트 포맷"
version: "0.1.0"
status: implemented
created: 2026-06-15
updated: 2026-06-15
author: GOOS행님
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "doctrine, verification, evidence, integrity, trust"
tier: S
---

# SPEC-EVIDENCE-CLAIM-INVARIANT-001 — 검증 주장 무결성 doctrine

## A. 의도 / Provenance (WHAT & WHY)

### A.1 한 줄 요약

순수 doctrine 규칙 파일 1개(+ template mirror)를 신설하여, **"관측하지 않은 검증을 주장하지 말 것(no unobserved-verification-claim)" invariant**, **baseline 무결성 귀속(baseline-integrity attribution)**, **fable-ish 5-섹션 증거-부착 리포트 포맷**을 정책 계층(policy layer)으로 codify한다. 코드/CI/런타임 훅 없음 — markdown 규칙 파일만.

### A.2 Provenance — IMP-06 (fable-ish 13-에이전트 분석 로드맵의 최종 채택 항목)

본 SPEC은 MoAI의 "Verify, Don't Assume" 자세를 점검한 fable-ish(chrisryugj) 13-에이전트 분석 로드맵의 **IMP-06**(최종 채택 항목)이다. 선행 항목 두 건은 모두 완료되었다:

- **SPEC-HOOK-DISCIPLINE-WIRING-001** (IMP-01) — dormant enforcement 훅 배선
- **SPEC-STOP-EVIDENCE-GATE-001** (IMP-02/03) — Stop 검증-증거 게이트 + 세션 ledger

IMP-06은 STOP-EVIDENCE-GATE-001의 `spec.md` §B.2(line 185)에서 명시적으로 OUT OF SCOPE로 연기되었다. 해당 SPEC은 "baseline-integrity attribution + 5-section report format + 'no unobserved-verification-claim' invariant"를 별도 후속 SPEC(IMP-06)으로 넘긴다고 적었다.

### A.3 계층 구분 — detection(기계) vs codification(정책)

| 계층 | 담당 SPEC | 성격 |
|------|-----------|------|
| 기계적 advisory **detection** (1개 형태: code-session false-success) | SPEC-STOP-EVIDENCE-GATE-001 | 런타임 코드(advisory, warn-first, fail-open) |
| 정책/doctrine **codification** (invariant + baseline 귀속 + 리포트 포맷) | **본 SPEC (IMP-06)** | 순수 markdown doctrine |

본 SPEC은 invariant의 정책 codification만 제공한다. STOP-EVIDENCE-GATE-001은 그 invariant가 위반되는 하나의 shape를 advisory로 탐지할 뿐이며, 정책 자체를 codify하지 않는다. 두 계층은 상호보완(complementary)이며, doctrine 파일은 STOP-EVIDENCE-GATE-001을 보완적 기계-탐지 계층으로 cross-reference한다.

### A.4 동기 결함 클래스 — `L_manager_docs_false_backfill_report`

본 doctrine이 표적으로 삼는 일반적 결함 클래스: **어떤 actor가 실제로 관측하지 않은 검증/완료를 주장하는 것**. 동기 사례는 sync-phase에서 `sync_commit_sha`가 "backfilled"되었다고 보고했으나 실제로는 placeholder였던 사건(`L_manager_docs_false_backfill_report`)이다. 핵심 명제: **"증거 부재는 성공의 증거가 아니다(Evidence absent ≠ evidence of success)."**

본 doctrine은 그 정확한 sync-phase 사건을 런타임에서 탐지한다고 **주장하지 않는다**(런타임 탐지는 §X.1 OUT OF SCOPE — STOP-EVIDENCE-WRITER-001 등 후속 소관). doctrine은 모든 actor에게 적용되는 정책 규범을 명문화할 뿐이다.

---

## B. 런타임(run-phase) 산출물 명세 — 신설 doctrine 규칙 파일 1개(+ mirror)

run-phase는 **단일 신규 doctrine 규칙 파일**을 생성한다. 파일은 다음 **세 구성요소**를 정확히 codify해야 한다.

### B.1 구성요소 1 — "no unobserved-verification-claim" invariant (핵심 정책)

규칙 파일은 다음 invariant를 명문화한다:

> 어떤 actor도 실제로 관측하지 않은 검증(verification) 또는 완료(completion)를 주장(assert)해서는 안 된다(MUST NOT). **증거 부재는 성공의 증거가 아니다(Evidence absent ≠ evidence of success).**

이것은 정책 계층(policy layer)이다. 규칙 파일은 STOP-EVIDENCE-GATE-001을 보완적 기계-탐지(mechanical-detection) 계층으로 cross-reference한다.

#### B.1.1 결속 범위(binding scope) — 두 표면(BOTH)

invariant는 다음 **두 표면 모두**를 결속한다. doctrine 파일은 두 표면을 명시적으로 명명(name)해야 한다:

1. **오케스트레이터 자가-보고(orchestrator self-report)** — output-style `.claude/output-styles/moai/moai.md` §8의 Completion Report / Verification Matrix 배너, 그리고 Trust-but-verify 배치(batch).
2. **manager 에이전트 완료 리포트(manager-agent completion report)** — manager-develop / manager-docs의 §E 자가검증(self-verification, E1-E7).

### B.2 구성요소 2 — baseline 무결성 귀속(baseline-integrity attribution)

규칙 파일은 다음 요건을 명문화한다:

> 모든 검증 주장(예: "tests pass", "coverage 87%", "lint clean", "0 0 sync")은 **실제로 측정된 baseline**(실행한 명령 + 관측된 출력)에 귀속(attribute)되어야 하며, 가정하거나 무관한 이전 측정에서 이월(carry over)해서는 안 된다.

### B.3 구성요소 3 — fable-ish 5-섹션 증거-부착 리포트 포맷 (사용자 확정 구조)

규칙 파일은 다음 **5개 섹션을 정확히(verbatim)** 정의한다:

1. **주장(Claim)** — 무엇을 단언하는가
2. **증거(Evidence)** — 실제로 실행한 명령 + 그 verbatim 출력(요약이 아님)
3. **baseline 귀속(attribution)** — 주장이 어떤 baseline에 대해 측정되었는가
4. **미검증(Gaps)** — 무엇을 명시적으로 관측하지 **않았는가**(negative space — 이것이 침묵하는 미관측 주장을 방지한다)
5. **잔여 위험(Residual risk)** — 남은 불확실성 / 연기된 검증

§4 미검증(Gaps)이 본 포맷의 핵심 방어선이다. 이 섹션이 actor로 하여금 관측하지 않은 것을 명시적으로 표기하도록 강제함으로써, 침묵 속에서 미관측 주장이 성공처럼 보이는 것을 차단한다.

---

## C. 런타임 산출물 — 파일 경로 결정

| 산출물 | 경로 | 중립성 |
|--------|------|--------|
| Local doctrine 규칙 | `.claude/rules/moai/core/verification-claim-integrity.md` | SPEC provenance 허용 |
| Template mirror | `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` | **내부-중립 필수** (SPEC ID / REQ / 날짜 / SHA / 메모리 ref 금지) |

`core/` 서브디렉터리 선택 근거(plan.md §D.2에 상술): invariant가 오케스트레이터(`agent-common-protocol.md`, `moai-constitution.md` 동거)와 manager 에이전트를 동시에 결속하는 **cross-cutting** 정책이므로, 품질 게이트 전용인 `quality/`보다 `core/`가 정합한다.

---

## 3. Acceptance Criteria (Tier S inline AC — 전부 grep/파일-존재 검증 가능)

각 AC는 run-phase가 기계적으로 검사 가능하도록 falsifiable하게 작성한다.

| AC ID | 검증 명제 | 검증 방법 (mechanical) |
|-------|-----------|------------------------|
| AC-ECI-001 | doctrine 규칙 파일이 선택 경로에 존재 | `test -f .claude/rules/moai/core/verification-claim-integrity.md` |
| AC-ECI-002 | template mirror가 병렬 경로에 존재 | `test -f internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` |
| AC-ECI-003 | "no unobserved-verification-claim" invariant 명문 + 두 결속 표면(오케스트레이터 self-report + manager-agent §E) 모두 명명 | §3.2 C3 — concrete token grep (D2 정정: "또는 동치" escape hatch 제거) |
| AC-ECI-004 | 5-섹션 증거-부착 리포트 포맷 정의 (5개 섹션 전부) | `grep`: `주장`, `증거`, `baseline 귀속`, `미검증`, `잔여 위험` 모두 존재 |
| AC-ECI-005 | baseline 무결성 귀속 요건 존재 | `grep`: `baseline` + `귀속`(또는 `attribution`) + `측정`(또는 `measured`) |
| AC-ECI-006 | template mirror 중립성 — 내부 누출 0 | §3.2 C6 — real-alternation grep (D1 정정: 코드블록 이동, vacuous `\|`=literal-pipe 회피) |
| AC-ECI-007 | 기존 규칙으로의 cross-reference ≥4개 (내용 복제 없음) | §3.2 C7 — 단일 `grep -cE ... -ge 4` (D3 정정: human-tally 제거) |

AC 전부 MUST(gating).

### 3.1 cross-reference 대상 (내용 복제 금지 — SSOT 준수)

규칙 파일은 다음 기존 표면을 **복사가 아닌 cross-reference**한다:

- `.claude/rules/moai/core/agent-common-protocol.md` § Skeptical Evaluation Stance (line 113)
- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #6 "Verify, Don't Assume" (line 262)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § E (E1-E7 자가검증 — 5-섹션 포맷이 이를 일반화/관계함, line 167)
- `.claude/rules/moai/workflow/verification-batch-pattern.md`
- `.claude/output-styles/moai/moai.md` §8 (Verification Matrix line 368 / Completion Report line 574)

### 3.2 실행 가능 검증 명령 (executable — plan-audit D1-D3 정정)

AC 검증 명령은 markdown 테이블 셀이 아닌 fenced 코드블록에 둔다. 이유: 테이블 셀 내부의 `|`는 markdown escape(`\|`)가 필요하나, `\|`는 ERE(`grep -E`)에서 alternation이 아닌 **literal pipe**로 해석되어 grep이 vacuous(항상 0-match)해진다 (plan-audit D1, 경험적 입증: `SPEC-EVIDENCE` 누출이 old form에서 0-match). 본 SPEC invariant("관측하지 않은 검증을 주장 금지")의 self-apply — 검증 명령은 실제 실행 가능(observable)해야 한다. run-phase는 verbatim 이 명령을 §E에서 실행하고 그 출력을 증거로 보고한다.

```bash
RULE=.claude/rules/moai/core/verification-claim-integrity.md
MIRROR=internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md

# C1: local doctrine 파일 존재
test -f "$RULE" && echo C1-PASS
# C2: template mirror 존재
test -f "$MIRROR" && echo C2-PASS
# C3 (D2 정정): invariant + 두 결속 표면을 concrete token으로 검증 ("또는 동치" 제거)
grep -q 'unobserved' "$RULE" && grep -q 'Evidence absent' "$RULE" \
  && grep -q 'output-styles/moai/moai.md' "$RULE" \
  && grep -qE 'manager-develop.*§E|E1[^0-9].*E7' "$RULE" && echo C3-PASS
# C4: 5-섹션 포맷 (5개 전부 존재)
for s in '주장' '증거' 'baseline 귀속' '미검증' '잔여 위험'; do grep -q "$s" "$RULE" || echo "C4-FAIL: $s"; done
# C5: baseline 무결성 귀속 요건
grep -qE 'baseline' "$RULE" && grep -qE '귀속|attribution' "$RULE" && grep -qE '측정|measured' "$RULE" && echo C5-PASS
# C6 (D1 정정): mirror 중립성 — REAL alternation, dot escape. 매치 0 기대.
grep -nE 'SPEC-[A-Z]|feedback_|/Users/|CLAUDE\.local|[0-9a-f]{7,40} commit' "$MIRROR" || echo C6-PASS-0-leaks
# C7 (D3 정정): cross-ref >=4 — 단일 mechanical 명령
test "$(grep -cE 'agent-common-protocol\.md|moai-constitution\.md|manager-develop-prompt-template\.md|verification-batch-pattern\.md|output-styles/moai/moai\.md' "$RULE")" -ge 4 && echo C7-PASS
```

C6/C7의 real-alternation 형태는 테이블 셀이 아닌 코드블록이므로 markdown escape가 불필요하다 (`|`가 literal로 보존됨 → ERE alternation 정상 작동).

---

## X. 범위 경계 (Scope Boundaries)

### X.1 Out of Scope — 본 SPEC이 빌드하지 않는 것

- 기계적 CI 가드 / Go 존재-테스트(presence-test) — 사용자가 순수 doctrine을 선택했으므로 코드 산출물 없음.
- sync-phase 원천 사건 형태(`L_manager_docs_false_backfill_report`)의 **런타임 탐지** — 후속 STOP-EVIDENCE-WRITER-001 또는 별도 런타임 SPEC으로 연기.
- coverage-relation correlation — STOP-EVIDENCE-GATE-001의 연기 항목이며 IMP-06 소관이 아님.
- output-style 본문 / manager 에이전트 본문의 대규모 재작성 — run-phase는 신규 규칙 파일 작성 + 최소 cross-ref 포인터만 추가(plan.md §F M3 참조).

---

## HISTORY

- 2026-06-15 — manager-spec — plan-phase artifacts 작성 (status: draft, Tier S, 2 artifacts: spec.md + plan.md, + progress.md §F.1 signal). IMP-06 (fable-ish 13-에이전트 로드맵 최종 채택 항목). 선행: SPEC-HOOK-DISCIPLINE-WIRING-001(IMP-01), SPEC-STOP-EVIDENCE-GATE-001(IMP-02/03, §B.2 line 185에서 OUT OF SCOPE 연기).
