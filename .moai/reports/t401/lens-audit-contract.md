# 렌즈 2 — plan-phase 감사 체인 출력 계약

조사 트리: `.claude/worktrees/t401` @ `ad272be20`
읽은 문서: `agents/moai/plan-auditor.md`, `skills/moai/workflows/plan.md` +
`plan/spec-assembly.md`, `rules/moai/workflow/spec-workflow.md`,
`rules/moai/workflow/orchestration-mode-selection.md`

## plan-auditor 리포트 필수 절 (`plan-auditor.md:398-436`)

헤더(`Iteration`/`Verdict`/`Overall Score`) → `## Must-Pass Results` → `## Category Scores`
→ `## Defects Found` → `## Regression Check`(iter2+) → `## Recommendation`

판정 어휘는 **`PASS | FAIL` 두 토큰뿐**(`:401`). `UNVERIFIED` 는 차원별 상태이며 must-pass 에서는
FAIL 로 계산(`:25`, `:131`). `STOP` 은 Verdict 블록 안의 신호(`:456`). **`PASS-WITH-DEBT` 는
감사자 토큰이 아니다** — 오케스트레이터가 AskUserQuestion 으로 제시하는 사용자 선택지(`:459`, `:468`).

## 처방(remediation)은 의무다 — 두 겹으로

1. 결함 레코드 형식 자체에 박혀 있음(`:423`):
   `D1. {id} — {file}:L{N} — {desc} — Severity: … — Class: … — Required fix: {concrete, actionable fix instruction}`
   **`Required fix:` 없는 결함 형식은 존재하지 않는다.**
2. `## Recommendation` 절(`:433-435`) — FAIL 이면 번호 매긴 실행가능 수정지시, PASS 면 근거 서술.
   **생략 분기가 없다.**

추가 처방 방출: STOP 시 scope-reduction 제안(`:456`), iter-3 FAIL 시 "recommends user
intervention"(`:448`).

반대 방향 가드도 있음(M6, `:188`): "A long list of optional findings does not by itself justify a
FAIL, and it must not be used to manufacture one."

## MUST-PASS 는 4개가 아니라 **8개** (MP-1…MP-8, `:135-162`)

REQ 번호 일관성 / EARS·GEARS 형식 / frontmatter 12필드 스키마 / 22절 언어중립성 /
D7 cross-SPEC / D8 cross-platform / `[NEEDS CLARIFICATION]` 미해결 0 / RED-now 재실행.
"ANY single must-pass failure = overall FAIL." N/A 는 auto-pass.

티어 임계는 SSOT 위임(`spec-workflow.md:138-142`): **S 0.75 / M 0.80 / L 0.85**, `tier:` 부재 → L.

## [핵심] 사용자가 감사 출력을 처음 보는 지점

| 경로 | 사용자 노출 | 권고 부착 |
|---|---|---|
| iter-1 PASS (happy path) | 리포트 본문 **안 보임**. 로그 1행뿐(`spec-assembly.md:196`). 첫 실질 노출은 Implementation Kickoff Approval **과 같은 턴**에 딸려오는 HTML 포인터(`:198-206`) | HTML 이 verdict·score·must-pass·**defects** 를 렌더(`:204`) → **권고가 첫 노출 순간에 이미 붙어 있음** |
| iter-1~2 FAIL | 사용자 노출 **0**. 리포트가 manager-spec 으로 직행(`:220-226`) | 에이전트 간 소비 |
| 3× FAIL 에스컬레이션 | review-1~3 전문 + 3옵션 AskUserQuestion(`:234-245`) | 전량 노출 |
| run-gate FAIL/INCONCLUSIVE | "Block Phase 1, surface report, AskUserQuestion"(`spec-workflow.md:401,403`) | 전량 노출 |

**어느 경로에서도 결함이 `Required fix` 없이 노출되는 일은 없다** — 필드가 레코드 형식의 일부이므로.

### 카드 전제 정정

카드 본문은 "plan-auditor·오케스트레이터가 관측과 권고를 항상 함께 낸다"고 적었다. 측정 결과
**오케스트레이터 축은 참, plan-auditor 축은 부분적으로만 참**이다. plan-auditor 의 권고는
대부분(iter-1 PASS, iter-1~2 FAIL) **에이전트 간 소비**이고 사용자에게 도달하지 않는다.
사용자에게 실제 도달하는 것은 ① Kickoff 턴의 HTML ② 3×FAIL ③ run-gate 차단 세 경로다.

## Implementation Kickoff Approval 의 위치

순서: plan Phase 11 판정 → (plan) HTML 방출 → `/moai run` Phase 1 Plan Audit Gate(스킵 가능)
→ **Implementation Kickoff Approval** → Phase 4 모드 선택.

`orchestration-mode-selection.md:18` `[ZONE:Frozen][HARD]`: "Implementation Kickoff Approval is
mandatory and score-independent (a plan-auditor PASS or a high skip-eligible score never
auto-bypasses it)". 게이트 자체가 `run.md:137` `[HARD]` 로 **"first option marked
'(Recommended)'"** 을 요구한다 — 즉 **현행 유일한 human gate 가 그 자체로 anchoring 표면**이다.

## zero-flag ≠ pass 는 현행에 어떻게 서 있나

- "결함 0 = PASS" 라는 절은 **없다**. `(If no defects found: "No defects found.")`(`:426`)는 렌더링
  지시일 뿐 판정 주장이 아니다.
- PASS 정의 = must-pass 전부 PASS/N-A **AND** 총점 ≥ 티어 임계 **AND** 모든 PASS 에 증거 인용
  (`:24`, `:131`, `:464`). 결함 0 은 그 어느 조건도 충족시키지 못한다.
- 즉 제보의 "zero-flag ≠ pass" 우려는 현행 계약에서 **이미 구조적으로 성립**한다. 다만 그 취지를
  명시한 절은 없고, 반대 방향 가드(M6 `:188`)만 있다.

## "judgment point" 는 first-class 로 존재하지 않는다

리포트 템플릿(`:398-436`)에 미해결 질문·불확실성·인간 위임 결정을 담는 절이 **없다**. 모든 발견은
severity + class + `Required fix` 를 갖춘 결함으로 `## Defects Found` 에 착지해야 한다.

근접 구성물 전부 결함으로 접히거나 오케스트레이터로 라우팅된다:

- **MP-7** `[NEEDS CLARIFICATION]` 마커(`:149`) — 가장 가깝지만 **severity=critical 결함으로 접힌다**.
  에스컬레이션은 오케스트레이터 몫("The orchestrator MUST resolve each marked topic via
  `AskUserQuestion` … before Implementation Kickoff Approval"). **감사자는 탐지만 하고 묻지 않는다.**
- **UNVERIFIED** — 검증 불가 상태이나 FAIL 로 계산된다. 인간에게 올라가는 질문이 아니다.
- **Class `optional`**(`:184-186`) — "orchestrator's discretion" 으로 유예되지만 여전히 결함 형태이고,
  유예 대상이 사용자가 아니라 오케스트레이터다.
- **구조적 차단**: plan-auditor 의 `tools:` 에 `AskUserQuestion` 이 없다(`:7`).

→ 제보가 제안한 Stage 2 "judgment point" 는 현행 체인에 **대응물이 없다**. 다만 이는 리드가
별건으로 분리한 1번(SPEC Review 내부 게이트) 범위다.

## 훅 기반 런타임 관측 가능성 (회귀 근거 후보 사전 검증)

- `.claude/settings.json:42-71` — `PreToolUse` 에 이미 matcher 3블록 등록
  (`Write|Edit|Bash`, `Agent|Task`, `SendMessage|TaskStop`) 전부 `handle-pre-tool.sh` 로 라우팅.
  matcher 는 도구명 정규식이므로 `AskUserQuestion` 추가는 **기존 `SendMessage|TaskStop` 블록과
  동일한 형태**다.
- payload 접근 선례 확인: `internal/hook/agent_stop_guard.go:307`, `:368`, `:448` 이
  `input.ToolInput` 에서 TaskStop 대상 / SendMessage 수신자 / Agent spawn 을 파싱한다.

→ AskUserQuestion payload 를 PreToolUse 에서 관측해 `(권장)`/`(Recommended)` 레이블 유무를
JSONL 로 기록하는 것은 **기존 선례와 같은 기계다**. 정적 grep 이 못 잼는 "에이전트가 규약을 실제로
따르는가"를 이 경로로 잴 수 있다.
