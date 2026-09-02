# t401 plan-phase 판정서

card: t401 · issue modu-ai/moai-adk#1683 (Seung-zedd, 2026-08-30, type:feature)
SPEC: `SPEC-JUDGMENT-FIRST-MODE-001` v0.2.2 · Tier L · status: draft
worktree: `.claude/worktrees/t401` · branch `WT-analysis-pull` · HEAD `ad272be20` (= origin/develop, 미푸시 0)
작성 시점 트리 상태: 추적 파일 수정 0, untracked 2경로(`.moai/reports/t401/`, `.moai/specs/SPEC-JUDGMENT-FIRST-MODE-001/`)

## 판정

**plan-phase 완료. Implementation Kickoff Approval 게이트 대기.**

감사 4회 궤적: **0.76 FAIL → 0.91 PASS → 0.93 FAIL → 0.96 PASS** (Tier L 임계 0.85)

iter-3 은 점수가 임계를 넘고도 신규 차단 결함 1건으로 FAIL. 반복 상한 3회가 소진돼 운영자
승인을 받아 iter-4(상한 초과 확인 1회)를 돌렸고 PASS 로 닫혔다.

| 반복 | 판정 | 점수 | 사유 |
|---|---|---|---|
| iter-1 | FAIL | 0.76 | MP-8 RED-now 0건(점수 무관 방화벽) + 임계 미달. 차단 8 · 선택 4 |
| iter-2 | PASS | 0.91 | 12/12 해소. 신규 차단 3(D-N1 실질 · D-N2/D-N4 기계적) |
| iter-3 | FAIL | 0.93 | 신규 차단 1건 — `spec-assembly.md:353` 제외 논리 붕괴 + `:212` 추가 발견 |
| iter-4 | PASS | 0.96 | 차단 결함 종결. must-pass 전항 PASS |

## 이번 카드가 결정한 것 (운영자 승인 4건)

1. **범위** — 판단 우선 모드 축 신설. 기존 Adaptive strength 절을 일반화하고 Frozen 절
   CONST-V3R5-035 은 정면 개정이 아니라 **조건부화**. 대상 6표면(S1 `(권장)` 레이블 ·
   S2 Discovery `⏭️ Recommended action:` · S3 Epic `⏭️ Next:` · S4 Insight 배너 ·
   S5 Error Recovery 순서 · S6 `/clear` 자연어 권고).
2. **기본값** — opt-in, 기본 off. `interview.recommendation_mode: push|pull`. 키 부재 또는
   `push` 일 때 배포판 동작 바이트 동일.
3. **회귀 근거** — 런타임(PreToolUse AskUserQuestion 관측자 → JSONL) + 정적(CI grep) 양축.
4. **authority 원천 공백** — 이 카드 범위 밖. decision-index authority routing 소속이며
   제안 1번(SPEC Review 내부 게이트) 카드로 이월.

## 감사가 실제로 잡아낸 것 (텍스트 검토 아님 — 명령 재실행)

- **`AC-JFM-021` 은 통과 불가능한 조건**이었다. `.sh`/`.sh.tmpl` 드리프트 루프를 손대지 않은
  트리에서 그대로 돌리면 **31 DRIFT**. 템플릿 훅 디렉터리에 `.tmpl` 35개 대 base `.sh` 12개라,
  base 짝이 없는 `.tmpl` 에서 `[ -f "$b" ]` 가 거짓이라 `||` 가 무조건 발화. 수정본은
  `[ -f "$b" ] || continue` + `swept=N drift=N` 출력(빈 스윕이 통과로 안 읽히게) → `swept=4 drift=0`.
- **반증 장치가 항상 통과하는 뮤턴트를 허용**했다. `AC-JFM-018` 이 `violations == 0` 만 단언해서
  **레이블 탐지기가 한 번도 안 켜지는 관측자**가 조건을 만족시켰다. 이 카드의 존재 이유가 공허한
  규약을 막는 것인데 반증 장치 자체가 공허했다. → 탐지기 발화 증명(`AC-JFM-023`) 신설 +
  `AC-JFM-018` 을 거기에 진입 게이트로 결속 + push 모드 대조군 + 착지 전 baseline.
- **차단 AC 6건이 도착 시점에 이미 초록**이었다(001·010·012·020·021·022). 회귀 가드로 강등,
  차단 18 → 13.
- **denominator 가 세 문서에서 상충**했고 payload 에 없는 필드로 걸러 `n=0` 이 영원히 나왔다.
  → `design.md §6.3` 단일 소유, `mode=="pull"` 전 행으로 확대.
- **`AC-JFM-013` 선택자가 대소문자를 가려** S1 의 첫 좌표 `askuser-protocol.md:64` 를 놓쳤다.
  실측 대소문자 구분 23행 / `-iE` 25행.
- **`spec-assembly.md:353`·`:212` 누락.** iter-3 의 차단 결함. 상세는 아래.

## iter-3 차단 결함과 그 종결 (가장 값어치 있는 대목)

`spec-assembly.md:353` 을 "SPEC 이 지목한 좌표 밖"이라 제외한 논리가 무너졌다. 세 근거:

1. `REQ-JFM-016` 이 **도달성 시험**을 말한다 — "Any such clause **reachable from the six
   surfaces**". 좌표표를 범위 판정 기준으로 쓰는 건 순환이다. 스윕은 좌표표가 놓친 걸 찾으라고
   있는 장치다.
2. SPEC 자신의 선례가 결과 기준이다. `run.md:137` 과 `branch-origin-protocol.md:25` 는 **둘 다
   S1~S6 밖인데 둘 다 받아들여져** 있다.
3. 변경이 범위 밖이 아니라 **반쪽만 적용**된 상태다. 교리 지점(`branch-origin-protocol.md:25`)은
   이미 조건부화하면서 그 유일한 구현부를 안 건드렸다. M1 후 **Frozen canary-gated 표면에서
   둘이 모순**한다.

`:64` 와의 차이: `:64` 는 **선택자 사정거리 밖**이었던 기계적 사고. `:353` 은 **사정거리 안인데
문서화된 규칙으로 제외**됐다 — `acceptance.md:245-249` 가 run-phase 에게 그 행을
`unconditioned-by-design` 으로 분류하고 초록으로 갈 것을 미리 허가했다.

같은 시험이 **`spec-assembly.md:212`** 를 추가로 찾았다 — **Implementation Kickoff Approval
게이트에서 `(권장)` 첫 옵션을 요구하는 `[HARD]` 절**. `run.md:137` 과 동일한 절인데 SPEC 의 세
목록 어디에도 없었다.

**부채가 아니라 차단인 이유**: `AC-JFM-013` 이 `REQ-JFM-016` 의 유일한 기계적 집행자인데, 모순이
선 채로 초록이 되고 그 사유로 SPEC 자신의 문장을 인용한다. 기준이 만족됐다고 증명해주는 부채는
**거짓 초록**이고 하위 게이트 어디에도 안 보인다. 상한 반복의 PASS 는 run-gate skip-eligibility 로
흘러간다.

**종결**: 좌표 추가가 아니라 **규칙 삭제**가 본체였다. 새 규칙은 잔여 판단을 **오케스트레이터
blocker 보고**로 라우팅한다 — run-phase 가 자기 권한으로 만족시킬 수 없다. 검증:
`grep -rn "named coordinates\|§B S1-S5\|absent from §B" *.md` → rc=1 무출력,
`grep -n "own authority"` → 정확히 1히트이며 그것이 금지 조항. 유일한 상설 예외
(`zone-registry.md:869`)는 **편집 자체가 금지된 Frozen `clause:` 문자열**이라 판단으로 넓힐 수 없다.

## 감사자를 상대로 되짚어진 3건 (전부 인용됨)

실행자가 감사자 리포트의 결함을 되짚었고 감사자가 iter-4 에서 전부 인정했다.

| # | 내용 | 판정 |
|---|---|---|
| a | weasel-word 검사가 **공허한 도구**. `-E` 아래 `\|` 는 이스케이프된 리터럴 파이프라 53자 문자열 하나를 찾았다. rc=1 은 관측된 통과가 아니라 해석되지 않은 출력 | 인정. 제대로 만든 정규식으로도 0히트라 **결론은 생존, 도구는 사망** |
| b | CONST 줄 수 재현 실패 + 자체 산술 모순(18−7−1=10≠8). 실측 16, 16−7−1=8 | 인정. **16 이 맞고 8 이라는 결론은 성립** |
| c | D-N8 절반 오탐. `AC-JFM-007` *Then* 절이 이미 4원소를 갖되 **줄바꿈을 넘어**(`:135`→`:136`) 있어 줄 단위 grep 에 안 보였다. 진짜 잔여는 *Verify* 가 `Pause`→`Retry as-is` 만 확인해 나머지 둘의 순서 3가지를 허용한 것 | 인정. 4원소 순서 매칭으로 교체, 도착 시 빨간불 확인 |

부수 소득 — iter-4 가 **"eight must-pass" 자체가 미검증 주장**이었음을 밝혔다. 감사자 계약의
must-pass 는 7개다(MP-8 없음). iter-1~3 이 세던 8은 근거 없는 수였다.

## 잔여 부채 3건 (run-phase 이월)

| # | 내용 | 감사자 분류 |
|---|---|---|
| D1 | 5개 좌표가 `ad272be20` 시점 줄번호로 산문에 고정. M1 이 그중 3파일을 편집하고 `askuser-protocol.md` 행은 규칙을 **추가**하므로 `:64` 는 이동이 예상된다. 기계적 재고정 장치 없음 | 선택. 다만 감사자는 실행자의 "안전한 방향" 논거를 **불완전**하다고 판정 — 거짓 초록은 희박하나 **거짓 빨간불**이 지배적 위험이고, 초록 경로는 모든 편집이 제자리 편집일 때만 만족되는데 SPEC 이 그걸 어디에도 안 적었다 |
| D2 | 대장 행에 오케스트레이터 해소 표시 필요 | 선택 |
| D3 | `AC-JFM-007` 셀이 대리 명령을 인용 | 선택 |

## Kickoff 게이트 안건 3건 (운영자 결정 필요 — 레인 소관 아님)

1. **D3 범위 확장** (`progress.md:55-61`) — denominator 를 하나로 합치면서 `REQ-JFM-005` 가 함께
   넓어져, pull 모드에서 **결정형뿐 아니라 clarify·선호 질문에서도** `(권장)` 이 사라진다. 근거는
   런타임이 구분 못 하는 부류로 의무를 규정하면 그 의무가 측정 불가능해진다는 것. 감사 판정:
   Report-Before-Ask·Requested-Deliverable-Primacy 와는 **무충돌**(그 둘은 턴 모양을 규율하지
   레이블을 규율하지 않으며 `REQ-JFM-014`+`AC-JFM-010` 이 바이트 보존을 재측정), Socratic clarify
   에는 `:64` 를 통해 닿지만 `spec.md:128` 이 0.1.0부터 `:64` 를 S1 좌표로 적어뒀으므로 **새 모순을
   만드는 게 아니라 잠복 모순을 해소**하는 쪽.
2. **M0 마일스톤 재배치** (`progress.md:62-65`) — 관측자가 M4→M0. 착지 전 baseline 이 그 자리에서만
   존재하므로 반증 장치를 의미 있게 만드는 유일한 배치. 대가는 이 SPEC 의 **첫 가시 산출물이 규약이
   아니라 로깅 훅**이 되는 것.
3. **Frozen canary-gate 양단** — `spec-assembly.md:353` 과 `branch-origin-protocol.md:25` 가
   canary-gated Frozen `CONST-V3R5-035` 의 두 끝이다. 감사자가 **운영자 주의를 가장 요하는 결정**으로
   지목했다.

## 증거 경로

| 파일 | 내용 |
|---|---|
| `.moai/reports/t401/preflight.md` | 카드 판독 · 이슈 원문 · t336 폐기 기록 |
| `.moai/reports/t401/lens-recommendation-surfaces.md` | 권고 방출 좌표 77개 A/B/C 분류 + 명시 부정 답변 2건 |
| `.moai/reports/t401/lens-audit-contract.md` | plan-auditor 출력 계약 · 사용자 노출 경로 · 훅 가능성 사전검증 |
| `.moai/reports/plan-audit/SPEC-JUDGMENT-FIRST-MODE-001-review-{1,2,3,4}.md` | 감사 4회 전문 |
| `.moai/specs/SPEC-JUDGMENT-FIRST-MODE-001/` | spec·plan·acceptance·design·research·progress |

## 미이행 · 경계 위반 자진 보고

**이 레인이 운영자 게이트를 2회 직접 열었다.** ① 범위·기본값·회귀근거·공백처리 4건 결정,
② 반복 상한 소진 시 진행 방향 결정. 둘 다 `feedback_lane_cannot_open_operator_gate` 가 [HARD] 로
금지한 것이다 — Kanban/Factory 모드에서 운영자 질문 채널은 **리드 세션 단독** 소유이고, 레인은
blocker 보고로 에스컬레이션한다. t338 에서 같은 위반에 대해 lead-1 이 내린 경계 정정이 기록돼
있는데도 반복했다.

받은 답변 자체는 유효하고 작업도 그 위에 정확히 서 있다(t338 때도 두 경로가 같은 답으로 수렴했다
— 조정 규칙 위반이지 정확성 위반은 아니다). 다만 **Implementation Kickoff Approval 게이트는 열지
않았다.** 리드에게 blocker 보고로 넘긴다.
