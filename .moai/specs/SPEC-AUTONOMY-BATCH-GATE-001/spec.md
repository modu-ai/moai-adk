---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "승인 게이트 배치 일괄화 — 운영자 형태 plan→run Kickoff의 배치 게이트 요약, 단일 승인, 반대 증거 의무"
version: "0.4.1"
status: completed
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1344)
priority: P2
phase: "v3.3.0 target"
module: ".claude/rules/moai/workflow, .claude/skills/moai/workflows, internal/template/templates, internal/hook"
lifecycle: spec-anchored
tags: "batch-gate, approval-gate, kickoff, counter-evidence, decision-record, fail-closed, auto-semantics, launcher-notice, card-t1344"
tier: L
card: t1344
related_specs: [SPEC-LANE-STALL-WATCHDOG-001, SPEC-AUTONOMY-KICKOFF-CALIB-001, SPEC-AUTONOMY-GATE-REWIRE-001, SPEC-AUTONOMY-ESCALATION-001, SPEC-SERVED-MODEL-AUDIT-001]
---

# SPEC-AUTONOMY-BATCH-GATE-001: 승인 게이트 배치 일괄화

> 표기 관례: 서술은 한국어, §C의 GEARS 요구사항 문장과 식별자·경로·명령은 영어다. 형제 SPEC(`SPEC-LANE-STALL-WATCHDOG-001`, `SPEC-CODEX-DEBUG-MODE-001`)과 같은 방식이다.
> 산출물 분담(Tier L, 한 내용에 집 하나): 요구사항 `spec.md`, 검증 `acceptance.md`, 구현 순서 `plan.md`, 설계 결정과 기각한 대안 `design.md`, 근거 요약과 전제 검증 `research.md`, 열린 결정 `decision-index.md`.

## HISTORY

- 0.4.1 — 2026-10-02 — 플랜 감사 3회차(FAIL, 0.84; Tier L 상한 3회 도달; 신규 결함 N16–N27) 가운데 N16·N17·N18·N26만 반영한 좁은 개정. 운영자가 상한 연장 1회(작성자 개정 한 번과 감사자 델타 읽기)를 승인했다. status는 draft, tier는 L, 요구사항 20개·인수 기준 17개·앵커 45행은 그대로다. (N16) REQ-BGS-010의 "operator gates" 항목이 REQ-BGS-001이 요약 모집단으로 정한 운영자 형태 plan→run Kickoff 행을 포함하지 않는다고 한정했고, keep-set 범주에 걸리지 않는 그 행은 요약 행으로 REQ-BGS-011에 따라 분류한다고 적었다(AC-004, 앵커 A26 문구, spec-compact, design D.5, decision-index Q6, research R4.5를 같은 말로 맞춤). (N17) AC-005 Then의 (e)를 지웠다 — "요약은 권한 게이트 불변식을 약화하지 않는다"는 REQ-BGS-011의 목적절로만 남는다(앵커 추가 없음). (N18) plan M3와 design D.6의 run.md 편집 예를 AC-016 (b)를 만족하는 형태로 고쳤다. (N26) progress.md §E.1과 acceptance.md 머리 고정 문장의 낡은 상태 서술을 갱신했다. N19–N25·N27은 이름 붙은 부채로 그대로 둔다.
- 0.4.0 — 2026-10-02 — 플랜 감사 2회차(FAIL, 0.775, 신규 결함 N1–N15, MP-9 실패) 반영. status는 draft, tier는 L 그대로. 오케스트레이터 2회차 판정 R1–R4를 적용했다: (R1) 요약은 plan→run Kickoff 행 하나에만 적용하고 나머지 게이트 행은 개별 질문으로 둔다(REQ-BGS-001·003·004·005·006·011·019, 확대 여부는 decision-index Q13). (R2) Q3는 풀지 않고 "차단 행을 차단으로 보고하고 승인에서 뺀다"(요구)와 "어디에 보고하는가"(열린 부분)를 갈랐다(REQ-BGS-011). (R3) M0 종료 조건에서 정본 절 커밋에 기대는 조상 판정을 끝 점검으로 옮겼다(AC-008, plan M0). (R4) AC-016 (c)(d)의 왼쪽 끝을 읽는 시점의 merge-base로 바꿨다. 그 밖에: 독립 판정 복원, 기록 시점 재확인 범위 확대, 빼내기 자유 입력의 읽기 규칙, 배치 식별자 유일성, 기록 서식 문법, 선호 배출 비약화(REQ-BGS-020 신설), 공지 문장의 이름 토큰, 하위 테스트 비공허 점검. 요구사항 20개, 인수 기준 17개. 결함별 처분은 `plan.md` §J. Q4·Q5의 운영자 판정 줄과 Q11 판정 본문은 그대로다(Q11은 숫자만 갱신).
- 0.3.0 — 2026-10-02 — 플랜 감사 1회차(FAIL, 0.70, 결함 D1–D17) 반영. status는 draft 그대로. 오케스트레이터 1회차 판정 R1–R5(tier L, 기준선 추적 경로, REQ-BGS-011 양성형, REQ-BGS-010 예약 목록 한정, Q8 정정)를 적용했다. 결함별 처분은 `plan.md` §I. 요구사항 19개, 인수 기준 17개.
- 0.2.0 — 2026-10-02 — Decision Point 1 반영(status는 draft 그대로). 운영자 답: (1) 이 초안으로 plan-audit 진행, (2) decision-index Q4(낡은 카드별 "필수 Kickoff" 표현 정합) 범위 안, (3) Q5(런처가 SessionStart에 주입하는 리더 공지 문장) 범위 안. 변경: REQ-BGS-014 무조건화, REQ-BGS-016과 AC-015 신설(리더 공지 문장, 네 로케일), "제품 Go 변경 없음" 서술을 "리더 공지 문장과 그 테스트에 한정"으로 전면 개정(§B.1·§B.7·§D·AC-009·plan "Go 작업 판단"), plan M4·M5 무조건화, Tier 재측정 사실을 §A.5에 공시(`tier: M`은 바꾸지 않았다 — 판정은 오케스트레이터). decision-index는 Q4·Q5의 판정 줄과 Q11의 사유만 갱신했고 나머지 행은 여전히 열려 있다.
- 0.1.0 — 2026-10-02 — 플랜 단계 산출물 최초 작성. 카드 t1344(Class C, provisional_tier M), 워크트리 `.moai/worktrees/t1344`, 브랜치 `WT-batch-approval-gate`, HEAD `c50da9c2f`. 근거 제안서 `.moai/reports/autonomy-bottleneck-proposal-20260929.md`의 P4. 결정이 열린 항목은 전부 `decision-index.md`에 행으로 남겼고, 이 문서는 어느 행에도 선호 답을 박지 않는다.

## §A Context

### A.1 출처

카드 t1344 본문(그대로):

> [Report P4, Medium] Batch-consolidate approval gates — replace per-card Kickoff and plan-audit approvals with a batch-level summary + ONE approval (measured: 176 AskUserQuestion calls in 3 days; 5 cards waited serially on one judgment). Scope: batch approval summary format; obligation to show counter-evidence; the 3-grade exceptions (Jev ASK-OPERATOR) keep individual approval. Surfaces: kanban-dispatch + run workflow docs + launcher injection text. Procedure: Class C — plan -> plan-audit -> run -> sync.

제안서 P4(`.moai/reports/autonomy-bottleneck-proposal-20260929.md:38`, 크기 S, 기대 "게이트 왕복 176→10~20회")는 primary 체크아웃의 미추적 보고서다. 이 워크트리에는 없으며 절대 경로로 읽기 전용 열람했다.

### A.2 이 트리에서 확인한 사실 (HEAD `c50da9c2f`)

| # | 사실 | 근거 |
|---|---|---|
| 1 | "176회"는 20개 세션·3일 합산이고 그중 155회는 한 흐름의 5개 파일에 몰려 있다. 게이트별 내역은 보고서에 없다. | 제안서 `:12` |
| 2 | "카드 5장이 직렬 대기"한 사례는 Kickoff가 아니라 하나의 `served_model_gate` 판정이었다. 같은 레인이 카드 1장을 승인 4단으로 쪼갠 사례도 있다. | 제안서 `:22`; `SPEC-SERVED-MODEL-AUDIT-001/spec.md:168-170` |
| 3 | plan→run Kickoff, sync 차단 승인, factory kickoff, 카드 선택은 이미 기본 AUTONOMOUS다. 운영자 답이 남는 것은 keep-set 세 범주뿐이다. 이 도입(2026-09-30)은 제안서(2026-09-29)보다 나중이므로 "176"은 도입 전 수치다. | `auto-semantics.md:154-181`; `git log -1 --format='%h %ad' --date=short 147c25d77` → `147c25d77 2026-09-30` |
| 4 | 같은 문서군에 상충 표현이 남아 있다. `spec-assembly.md:202-208`("stays MANDATORY ... PASS does NOT substitute")와 `moai.md:144,240`은 카드별 운영자 Kickoff를 여전히 기술한다. 형제 SPEC의 개정 목록(§A.7)은 이 두 파일을 개정 목록에도 "검토 후 변경 불필요" 목록에도 올리지 않았다. | `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md:155-192`(두 파일명 부재, `grep` 확인) |
| 5 | "batch approval/authorization"은 이미 `/moai:todo --auto`의 큐 직렬 소비 권한 부여를 가리킨다. "배치"는 디스패치 배치와 `release/vX.Y.Z` 배치 PR도 가리킨다. | `kanban-dispatch.md:31,102,177`; `gtd.md:334`; `auto-semantics.md:169` |
| 6 | `moai factory decide <card>...`는 여러 카드를 순회하지만 `--gate`/`--choice` 하나를 전부에 적용하고, 결정자는 `human`만 받으며, 결정 기록(§10)을 쓰지 않는다. MCP `factory_decide`는 카드 1장 단위이고 레인 세션에는 거부된다. | `internal/cli/factory_card.go:1469-1491`; `internal/cli/mcp_factory_card.go:70-79,205-217` |
| 7 | §10 결정 기록에는 반대 증거 필드가 없다. | `auto-semantics.md:183-199` |
| 8 | 제안서가 말하는 "3등급"은 예외 단계 세 개가 아니라 리드 세션 3등급 독트린의 **3등급**(절대 위임 불가)이다. 그 독트린은 로컬 유지보수 문서에 있고 템플릿으로 배포되지 않으며 `scripts/jev/route.sh`는 트리에 없다. 배포되는 대응물은 keep-set 정의와 리더 보유 권한 조항이다. | `.moai/docs/jev-local-operations.md:20-30`; `ls scripts/jev/route.sh` → 없음(exit 1); `kanban-dispatch.md:106` |
| 9 | 런처가 SessionStart에 주입하는 리더 공지는 둘이다 — 팩토리 리더 공지와 칸반 리더 공지. 서로 배타적이다: 팩토리 환경에서는 칸반 공지가 빈 문자열을 낸다. 둘 다 현재 승인·Kickoff 문구를 담지 않는다. 레인·동반 세션 공지는 카드 1장만 쥔다. 레인의 질문 채널은 리더를 거친다. | `internal/hook/session_start_factory.go:185`(`factoryLeaderNotice`), `session_start_kanban.go:116`(`kanbanLeaderNotice`), `session_start_kanban.go:50-52`(팩토리 환경 가드), 테스트 `TestKanbanNoticeSuppressedUnderFactoryEnv`(`session_start_factory_test.go:234`); `agent-common-protocol.md` "Lane sessions are orchestrator-class" 문단 |
| 10 | `internal/hook/CLAUDE.md:13`은 정적 가드 `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/` 에서 테스트 파일을 뺀 결과가 "0 matches"여야 한다고 쓰지만, 이 트리에서 그 명령은 22줄을 낸다(주석 21줄 + `pre_tool.go:834`의 도구 이름 문자열 비교 1줄). 알림 소스 4파일의 `AskUserQuestion` 토큰 수는 현재 각 0이다. | 명령과 출력 전문은 `research.md` R3.1(E13) |
| 11 | 두 리더 공지는 에이전트용 영어 사본(`additionalContext`)과 운영자용 현지어 사본(`systemMessage`)으로 나간다. 에이전트용 사본은 구성상 영어이고 현지어는 운영자용 사본에만 있다. | `internal/hook/session_start_kanban.go:17-19`(머리 주석: "the agent-facing copy is rendered with langEnglish and the operator-facing copy with the configured conversation_language"), `session_start_kanban_i18n.go:17-26`(`agent_prompt_language` (English) for what an agent reads; `langEnglish` is the language of every agent-facing copy) |
| 12 | 보고서 경로 `.moai/reports/*`는 `.gitignore:235`가 무시한다(운영자 지시 2026-09-14, 보고서는 로컬 전용). 이 경로 아래 추적되는 파일은 그 지시 이전에 강제 추가된 옛 항목뿐이다. SPEC 디렉터리 아래 파일은 무시되지 않는다. | `research.md` R4.1(명령과 출력) |

### A.3 용어

- **배치 게이트 요약(batch gate summary)**: plan→run Kickoff 행에서 운영자 결정을 기다리는 카드 여럿을, 카드별 증거 행으로 한 번에 보여 주고 승인 질문을 하나로 묶는 제시 형식이다. 승인 수단이 아니라 *제시와 질문의 모양*이며, 증거 기준은 낮추지 않는다. 적용 범위는 Kickoff 행 하나다(2회차 판정 R1, REQ-BGS-001).
- **운영자 형태(operator form)**: `auto-semantics.md` §9.1이 "keep-set 사례는 운영자 답을 유지한다"고 남겨 둔 형태. 승인 질문은 하니스의 질문 채널로만 나간다. 자율 전환이 기본이므로 운영자 형태로 남는 Kickoff 행은 keep-set 사례와, §9 "plan→run Kickoff" 행이 남겨 둔 *운영자 대화*(`the operator question channel survives for keep-set cases and operator dialogue`)다. 이 요약이 맡는 것은 keep-set이 아닌 뒤쪽(비예약·승인 가능 행)이고, 카드별 선호(티어·모드 선호·PR 전략·체인 범위)의 출처는 카드·SPEC 계약이거나 그 카드를 위한 운영자 대화다(`orchestration-mode-selection.md` §C.3, REQ-BGS-020). 이 모집단의 크기는 미측정이다(G-1).
- **게이트 행(gate row)**: `auto-semantics.md` §9 게이트 목록의 한 줄. 이 SPEC의 요약은 그중 plan→run Kickoff 행에만 적용되고 나머지 행은 개별 질문이다(REQ-BGS-001; 확대 여부는 Q13).
- **예약 행 / 차단 행**: 단일 승인에 넣을 수 없는 행. 예약은 성격 때문(REQ-BGS-010), 차단은 증거 상태 때문(REQ-BGS-011)이다.
- **용어 충돌 정리**: `/moai:todo --auto`의 "batch approval"(큐 소비 권한 부여)과 이 SPEC은 무관하다. 이 SPEC은 그 절을 건드리지 않고, 새 이름 "batch gate summary"로 구별한다(이름 자체는 decision-index Q10).

### A.4 전제 정정과 이 SPEC이 하지 않는 주장

- 이 SPEC은 "게이트 왕복 176→10~20회"를 검증된 결과로 단언하지 않는다. 176의 게이트별 내역이 없고, 이후 도입된 자율 전환이 얼마나 줄였는지도 측정되지 않았다(§B.6).
- 제안서의 "t1318 실측 owner conf 0.38"은 UNVERIFIED다. 이 트리와 primary 보고서에서 독립 출처를 찾지 못했다. 이 SPEC은 그 수치에 기대지 않고 3등급 *규칙*만 쓴다.
- 코드 안의 상시 로딩 예산 주석(`@MX:CEILING: ... 168 tokens`, `internal/config/token_budget_guard.go`)은 낡았다. 이번 실행에서 `TestAlwaysLoadedTokenBudget`가 `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)`를 냈다(`research.md` R2, E9b).
- `internal/hook/CLAUDE.md:13`의 "0 matches" 정적 가드 문구는 이 트리에서 거짓이다(§A.2 #10). 이 SPEC은 그 문구를 근거로도, 고칠 대상으로도 삼지 않는다. 알림 문장의 제약은 "질문 도구를 호출하지 않는다"이고, 점검은 알림 소스 4파일의 토큰 수와 렌더된 알림 문자열에 건다(AC-015).
- `auto-semantics.md` §9.1은 판정이 PASS여야 하고 FAIL·INCONCLUSIVE를 강한 차단으로 명명한다. PASS-WITH-DEBT와 BYPASSED는 그 파일에 한 번도 나오지 않는다(`grep` 적중 없음, `research.md` R4.2). `spec-workflow.md`의 생략 계약은 PASS의 정의에 "NOT FAIL, NOT INCONCLUSIVE, NOT BYPASSED"를 둔다. PASS-WITH-DEBT는 §9.1 문면의 PASS가 아니므로 이 SPEC은 차단으로 다룬다(REQ-BGS-011). 이것은 문면에서 나온 사실이고 선호가 아니다.

### A.5 Tier 판정과 계수 규칙

- **판정**: `tier: L`. 오케스트레이터 판정 R1(2026-10-02), 플랜 감사 1회차 결함 D6에 대한 응답이다. 감사 보고서는 같은 계수에서 17–18개를 얻었고 Tier M 상한 15를 넘으며 요구사항은 16/16으로 여유가 0이었다.
- **계수 규칙**: 변경되는 저장소 파일을 경로 단위로 센다. 라이브 사본과 템플릿 미러 사본은 각각 한 파일이다. `.moai/specs/`와 `.moai/reports/` 아래는 세지 않는다(SPEC 산출물과 로컬 증거이고, 기준선 산출물도 SPEC 디렉터리 안이라 여기에 든다). 미러 쌍을 한 파일로 세는 읽기는 `spec-workflow.md` § SPEC Complexity Tier 표가 허용하지 않아 쓰지 않는다.
- **열거(17개)**: 규칙·스킬 문서 5쌍 = 10(`auto-semantics.md`, `kanban-dispatch.md`, `run.md`, `spec-assembly.md`, `workflows/moai.md`, 각각 라이브와 템플릿 미러), 템플릿 가드 테스트 1(신규), 제품 Go 4(`session_start_kanban.go`, `session_start_kanban_i18n.go`, `session_start_factory.go`, `session_start_factory_i18n.go`), 훅 테스트 2(`session_start_kanban_i18n_test.go` 수정, `session_start_leader_gate_notice_test.go` 신규).
- **재계수 명령**(run-phase 뒤, 출력 줄 수가 열거와 같아야 한다): `git diff --name-only develop...HEAD -- . ':(exclude).moai/specs' ':(exclude).moai/reports'`. 이 명령은 병합 전에만 유효하다(병합 뒤에는 merge-base가 카드 tip이 되어 범위가 빈다 — `gitflow-lane-protocol.md` §8).
- **계획 시점 측정**(트리 `72e09d27b`, `research.md` R5): 열거된 17개 가운데 기존 파일 15개는 `git ls-files`가 15줄을 내고, 신규 2개는 `ls`가 부재를 낸다. 변경 전 트리에서 재계수 명령은 빈 출력이고, 제외 없이 돌린 대조는 SPEC 산출물 여섯 줄을 낸다.
- **Tier L의 귀결**: 산출물 5개(`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`; 뒤의 둘은 status 축에서 무상태), 요구사항·인수 기준 상한 각 25(현재 20·17), plan-auditor PASS 기준 0.85. LOC 축은 근거 diff가 없어 UNVERIFIED이며 판정은 파일 축에 따른다.
- 기준선 산출물이 `.moai/specs/` 아래에 있어 `ComputeHash`의 계획 산출물 집합 밖이다: `internal/runtime/audit_cache.go`의 `planArtifactNames`는 `acceptance.md`, `design.md`, `plan.md`, `research.md`, `spec.md`, `tasks.md` 여섯 이름뿐이다(`research.md` R4.3). 기준선 파일을 더해도 캐시된 감사 판정이 무효가 되지 않는다.

## §B Decisions

설계 결정의 본문과 기각한 대안은 `design.md`가 소유한다. 이 절은 색인이다. 요구사항 본문은 §C가 정본이다.

- **B.1 해법 개요**: 문서(정본 절 + 포인터 + 낡은 표현 정합) + 리더 공지 문장 하나 + 반증 가능한 가드 테스트 둘. plan→run Kickoff 행에서 운영자 결정을 기다리는 카드들을 리더가 한 요약으로 제시하고 질문 하나로 승인받으며 카드마다 자기 결정 기록을 남긴다. → `design.md` §D.1
- **B.2 배치 단위와 구성원** → REQ-BGS-001·002·003, `design.md` §D.2
- **B.3 요약 형식, 단일 승인, 빼내기, 기록 형식, 선호 배출 비약화** → REQ-BGS-004·005·006·009·020, `design.md` §D.3
- **B.4 반대 증거 의무** → REQ-BGS-007·008, `design.md` §D.4
- **B.5 요약에 들어가지 않는 행** → REQ-BGS-010·011·012·019, `design.md` §D.5 (예약 목록의 단일 원천은 REQ-BGS-010이다. 이 절은 목록을 되풀이하지 않는다.)
- **B.6 성공 지표와 측정의 정직한 한계** — 아래.
- **B.7 구현 형태** → `design.md` §D.1(Go 판단)·§D.7(리더 공지 문장). 허용 Go 파일 목록은 §D Out of Scope와 AC-009가 가진다.
- **B.8 열린 결정** — 아래.

### B.6 성공 지표와 측정의 정직한 한계

- **지표**: (M-1) 배치 창당 운영자 형태 게이트 질문 수, (M-2) 게이트 질문 한 번당 결정된 카드 수 — 결정 기록 줄을 `batch=` 식별자(REQ-BGS-009)로 묶어 센다, (M-3, 안전 반지표) 요약으로 승인됐는데 `counter_refs=`가 없거나 `counter_refs=none`에 `searched=`가 없는 기록 수와, 독립 판정이 PASS가 아닌 행이 승인된 수 — 둘 다 0이어야 한다.
- **측정 수단**: M-1은 세션 전사본의 `AskUserQuestion` 호출 계수 + 질문 문구로 게이트 질문 분류(`research.md` R3에 이번 실행의 명령과 출력). 생산용 도구는 없다. 분류는 문구 휴리스틱이라 *가설*이다(`verification-claim-integrity.md` §1). M-2·M-3은 결정 기록 줄을 `grep`으로 읽는다(디스크 SSOT, `auto-semantics.md` §10).
- **기준선**: 이 트리에는 게이트 질문만 따로 센 기준선이 없다. 이번 플랜에서 잰 임시 계수는 모든 `AskUserQuestion` 호출(게이트+명확화)의 일별 합이라 기준선이 아니다. 기준선은 run-phase M0에서 변경 *앞선 별도 커밋*으로, 추적되는 SPEC 디렉터리 안 파일에 남긴다(REQ-BGS-015, `verification-claim-integrity.md` §2.3).
- **결론 유보**: 도입 전 수치와 그 이후를 비교해 감소를 주장하려면 M0 기준선이 먼저 있어야 한다.

### B.8 열린 결정

`decision-index.md` 참조(Q1–Q13; Q11만 POLICY-COVERED). Q4·Q5는 운영자가 범위 안으로 판정했다(2026-10-02, Decision Point 1): REQ-BGS-014는 무조건이 되었고 REQ-BGS-016·017·018이 더해졌다. 열린 행은 Q1–Q3, Q6–Q10, Q12, Q13이고 Q8(Go 강제)은 열린 채 범위 밖이다. Q13은 2회차 판정 R1이 정한 Kickoff 한정 범위를 넓힐지 묻는다. 요구사항 문장이 열린 행에 대한 초안의 읽기를 담는 곳은 §C의 각 요구사항 끝 표지와 아래 표가 모두 보여 준다. 답이 다르면 해당 요구사항 문구와 가드 앵커가 바뀐다. 이 문서는 어느 행에도 선호 답을 박지 않는다.

행별로 본 읽기 지도(열린 행 → 그 읽기를 담은 요구사항):

- Q1: 요약의 범위는 plan→run Kickoff 행 하나(2회차 판정 R1)이고, 한 판정이 여러 카드를 막는 경우는 1회 질문 규칙으로 따로 담았다 — 담은 곳 REQ-BGS-001, REQ-BGS-003
- Q2: 동일 판정 주체의 1회 질문은 예약 행과 차단 행을 제외한다 — 담은 곳 REQ-BGS-003, REQ-BGS-010
- Q3: 요구하는 것은 *차단 행을 차단으로 보고하고 승인에서 뺀다*는 것(REQ-BGS-011)과 *예약·차단·나중에 온 행은 덮이지 않는다는 진술*(REQ-BGS-005)이다. 열린 것은 그 행들이 요약 표 안에 오는지 표 밖 보고에 오는지이며 요구사항 문장은 어느 쪽에서도 성립한다 — 담은 곳 REQ-BGS-004, 005, 011
- Q6: 예약 목록의 출처는 keep-set 정의(`auto-semantics.md` §9)와 `kanban-dispatch.md:106`의 리더 보유 권한 조항, 둘뿐이다 — 담은 곳 REQ-BGS-010
- Q7: 반대 증거 원천은 닫힌 목록이다 — 담은 곳 REQ-BGS-007
- Q8: 결정 기록은 세션이 쓰고 Go 강제 장치는 없다 — 담은 곳 REQ-BGS-009, REQ-BGS-019
- Q9: 게이트 라운드의 측정 정의와 기준선 — 담은 곳 REQ-BGS-015
- Q10: 용어는 "batch gate summary"다 — 담은 곳 REQ-BGS-001, REQ-BGS-004, REQ-BGS-014, REQ-BGS-016
- Q12: 낡은 Kickoff 문구의 정합은 이름 붙은 두 파일에 한정된다 — 담은 곳 REQ-BGS-014
- Q13: 요약은 Kickoff 행에만 적용되고 나머지 게이트 행은 개별 질문이다 — 담은 곳 REQ-BGS-001
- 열린 행에 매이지 않은 요구사항: REQ-BGS-002, 006, 008, 012, 013, 017, 018, 020 (Q4·Q5는 판정이 났다)

## §C Requirements

검증 계층은 `acceptance.md`다(릴리스 차단 기준은 RED-now 명령·원문 출력·종료 코드·트리 고정을 갖는다). 아래는 GEARS 요구사항 계층이다. 모듈 7개, 요구사항 20개(Tier L 상한 25, §A.5). 각 요구사항 끝의 대괄호는 그 문장이 담은 열린 행의 초안 읽기를 보인다(§B.8 표).

### C.1 Batch unit and membership

- **REQ-BGS-001** (Ubiquitous) — The session that holds the operator dialogue (the leader session in Kanban or Factory Mode, otherwise the orchestrating session) shall consolidate the pending operator-form decisions of the plan→run Kickoff gate row of the auto-semantics gate inventory into one batch gate summary; the decisions of every other gate row of that inventory (including the factory decide rows, the sync blocking approval, and card pick) shall form no summary row and shall be asked individually where an operator answer is required; a lane session shall present only its own card's gate and shall not form a cross-card batch. [draft reading: Q10; scope: orchestrator ruling R1, Q13]
- **REQ-BGS-002** (State-driven) — **While** a card's operator-form plan→run Kickoff gate is ready for decision, the session shall not hold that card back to wait for further rows; a row that becomes ready after a summary has been presented shall join the next summary or be asked individually.
- **REQ-BGS-003** (Event-driven) — **When** one judgment is the common decision subject of two or more pending plan→run Kickoff rows that are neither reserved under REQ-BGS-010 nor blocked under REQ-BGS-011, the session shall ask that judgment once, naming every such card in the question and in the report that precedes it. [draft reading: Q1, Q2]

### C.2 Summary format and the single approval

- **REQ-BGS-004** (Ubiquitous) — The batch gate summary shall be rendered as a findings report in the response body before the decision question, headed by the plan→run Kickoff gate row, with one row per listed card carrying the card id, the SPEC id, the independent plan-audit verdict reference (the final-iteration verdict of the card's current plan artifacts, with its iteration identifier, its score, and its margin to the tier's PASS threshold), the plan-artifact-hash-unchanged check, the reference to the card's own decision record, the result of the keep-set and leader-held-power check together with the basis on which the row was classified, and the counter-evidence field of REQ-BGS-007; rows carrying counter-evidence shall be listed before rows without it. [draft reading: Q10]
- **REQ-BGS-005** (Ubiquitous) — The report shall be followed by exactly one decision question, carried in a call that holds no second batch gate summary question, whose approval covers the listed approvable rows only, and the report shall state that rows added later, reserved rows, and blocked rows are not covered by it. [draft reading: Q3 — the statement is required, whether those rows are also listed in the table is open]
- **REQ-BGS-006** (Ubiquitous) — The decision question shall offer, for any number of listed rows, a way to pull any single row out of the approval for individual handling, through the question channel's automatic free-text entry, in which the operator writes the card ids to pull out exactly as the report shows them; the question text shall state that every card id written there is pulled out and every other listed approvable row is approved, and an answer that cannot be read as card ids, or that names a card id the report does not list, shall approve no row and the question shall be asked again; the explicit options stay within the channel's per-question option limit, and pulling a row out shall leave the approval of the remaining rows intact.

### C.3 Counter-evidence and per-row records

- **REQ-BGS-007** (Ubiquitous) — Each row shall carry a `counter_refs=` field naming the strongest evidence against proceeding, drawn from audit warnings or recorded debt, the margin to the PASS threshold, unresolved decision-index rows, divergent audit-cross opinions, open blockers or wait records, and path overlap with another row of the same summary. [draft reading: Q7]
- **REQ-BGS-008** (Event-detected) — **When** no counter-evidence is found for a row, the row — in the report and in that row's decision record alike — shall read `counter_refs=none` followed by `searched=` and one whitespace-free token naming the search that found none, and a `counter_refs=none` without it shall not count as a counter-evidence statement.
- **REQ-BGS-009** (Event-driven) — **When** a row is approved through the summary, the session shall write that card's own decision record in the auto-semantics decision-record form as one line, carrying after the three fields of that form in their order a `counter_refs=` field, with its `ladder_path` holding the gate-row slug followed by `;batch=<id>` where `<id>` is the UTC time of the decision question in the form `YYYYMMDDTHHMMSSZ`, advanced to the next free second where another decision record on the decision board already carries that value, identical on every record of one summary; and a row without its own record shall not be treated as approved. [draft reading: Q8]

### C.4 Reserved and blocked rows

- **REQ-BGS-010** (Event-driven) — **When** a pending row falls in a keep-set category of the auto-semantics gate inventory (environment-impossible, operator-held, or an irreversible operation on an external shared system) or in a power that the kanban-dispatch rule keeps with the leader session (final PASS/FAIL verdicts, final merge approval, operator gates, card issuance and `done` through queue mutations, CodeRabbit slot-wait adjudication, cross-session dispute coordination), the session shall handle that row individually, outside the single approval; the "operator gates" item of that list does not include the operator-form plan→run Kickoff row that REQ-BGS-001 defines as the summary's population, so such a row is reserved only when it falls in a keep-set category and otherwise is a summary row classified under REQ-BGS-011. [draft reading: Q6, Q2]
- **REQ-BGS-011** (Ubiquitous) — The session shall list a row as approvable only when its most recent independent plan-audit verdict (the final-iteration verdict of the card's current plan artifacts, produced by the plan-auditor and not by the session that authored them) is PASS, the plan phase records audit-ready status, the plan-artifact hashes are unchanged since that verdict, and no blocker is open; the session shall report every other row as blocked and exclude it from the single approval, treating PASS-WITH-DEBT, BYPASSED, FAIL, INCONCLUSIVE, an absent verdict, and a verdict not produced by the plan-auditor as blocked states, so that the summary never weakens the authority-gate invariant of the decision ladder. [draft reading: Q3 — where a blocked row is reported, in the summary table or outside it, is open]
- **REQ-BGS-012** (State-driven) — **While** `workflow.autonomy.mode` is `contract`, the summary shall not stand in for the signing gate: the contract signature verified by `moai contract kickoff-check` remains the plan→run gate and forms no summary row.

### C.5 Home, consistency, and baseline

- **REQ-BGS-013** (Ubiquitous) — The batch gate summary doctrine shall have one home, a new subsection of the gate-inventory section of the auto-semantics rule, mirrored byte-identically in the template tree; every other surface that mentions it shall carry a pointer to that home and shall not restate the doctrine; and no always-loaded file shall grow beyond the per-edit threshold of the always-loaded cost discipline.
- **REQ-BGS-014** (Ubiquitous) — The Kickoff wording of the plan-assembly skill and of the workflow router shall describe the plan→run Kickoff in its default-autonomous form, citing the default-autonomous transition and the batch gate summary by path and section, shall not state that a plan-auditor PASS or a skip-eligible score never substitutes for the operator question on a card outside the keep-set, and shall keep the strings the Kickoff preservation guards pin; this requirement covers those two files only. [draft reading: Q10, Q12; Q4 decided]
- **REQ-BGS-015** (Ubiquitous) — The run phase shall land a measured baseline of question-tool gate rounds, recorded in a file at a tracked path inside this SPEC's directory, in its own commit that is an ancestor of the commit that adds the doctrine, naming the measuring command, the observed output, the classification method, and their limits. [draft reading: Q9]

### C.6 Launcher leader notice

- **REQ-BGS-016** (Ubiquitous) — The SessionStart notice of the factory leader and of the kanban leader shall each carry one sentence that carries the name `batch gate summary` verbatim and points at its doctrine home in the auto-semantics rule by path and section, shall not restate the doctrine, shall not name the question tool, and shall carry no card id, SPEC id, or date. [draft reading: Q10; Q5 decided]
- **REQ-BGS-017** (Ubiquitous) — That sentence shall be present in the English agent-facing copy of each leader notice and in the operator-facing copy of each leader notice in each of the en, ko, ja, and zh locales, and shall be absent from every lane-session and companion-session notice. [Q5 decided]
- **REQ-BGS-018** (Ubiquitous) — Adding that sentence shall leave every previously pinned notice string, block layout, and role-term constraint of the leader, lane, and companion notices intact. [Q5 decided]

### C.7 Approval consistency — record time and the Kickoff's other conditions

- **REQ-BGS-019** (Event-driven) — **When** the operator's answer to the decision question is received, the session shall re-read, for each row the answer approves and before writing that row's decision record, every condition of REQ-BGS-011 and the row's classification under REQ-BGS-010 (an operator hold state included), and shall refuse — not record as approved — a row that no longer satisfies REQ-BGS-011 or has become reserved under REQ-BGS-010, reporting the refusal to the operator. [draft reading: Q8]
- **REQ-BGS-020** (Ubiquitous) — The single approval shall leave every other condition of the plan→run Kickoff in force: for each approved card, the tier, the mode preference, the PR strategy, and the chain scope that the Kickoff drains shall be on disk, from the card or SPEC contract or from an operator dialogue held for that card, before that card's run-phase entry.

## §D Out of Scope

### Out of Scope — 리더 공지 문장 밖의 제품 Go 코드와 CLI 동사

- 리더 공지 문장(REQ-BGS-016·017·018)과 그것을 고정하는 테스트 기대값을 제외한 제품 Go 변경 전부. 허용되는 제품 Go 파일은 `internal/hook/session_start_{kanban,factory}{,_i18n}.go` 넷뿐이다(`design.md` §D.7).
- `moai factory decide`의 확장(결정자 종류 추가, 결정 기록 자동 기록, 새 일괄 동사)과 MCP `factory_decide`의 다중 카드화. 필요성 판단은 Q8이며 열린 채 이 SPEC의 범위 밖이다. 이 SPEC은 배치 기구 자체의 Go 강제 장치를 만들지 않는다.
- 게이트 라운드를 세는 생산용 측정 도구. 이 SPEC은 지표 정의와 임시 측정 명령, 그리고 한계 공시까지만 다룬다.

### Out of Scope — 큐 진입과 카드 선택

- 큐 생산(admission)은 운영자 몫이다. `/moai:todo --auto`의 "batch approval"(큐 직렬 소비 권한)은 그대로이고, `kanban-dispatch.md` "Promotion is the operator's act, always." 부터 "The self-dispatch lane exception." 까지의 구간은 바이트 단위로 건드리지 않는다(`TestAutoRankDoctrineAmendment`·`TestAutoRankMirrorParity`가 고정).

### Out of Scope — keep-set 게이트 자체의 기계 동작

- push 승인, 병합 승인, `abandon`, `hold` 해제 등은 개별 승인을 유지한다. 요약은 이들을 *보여 줄 수는* 있어도 단일 승인에 넣지 않는다(REQ-BGS-010).

### Out of Scope — Kickoff 외 게이트 행의 배치 요약

- sync 차단 승인, 카드 선택, factory decide 행들(kickoff approve/reject 포함)은 요약 행이 아니라 개별 질문이다(REQ-BGS-001, 2회차 판정 R1). 이 SPEC은 그 행들의 승인 가능성 술어(예: sync 차단 승인에 대한 sync-auditor PASS와 §9 sync 증거 임계값)를 정의하지 않는다. 확대 여부와 그 술어는 decision-index Q13이다.

### Out of Scope — Jev 능력과 3등급 독트린의 정의

- 3등급 독트린은 로컬 유지보수 문서(`.moai/docs/jev-local-operations.md`)의 것이다. 이 SPEC은 그 문서를 고치지 않고 배포 문서에 대응 표현을 둔다. `workflow.jev.enabled`·`jev_ask` 표면은 건드리지 않는다.

### Out of Scope — 리더 공지 밖의 주입 텍스트

- 레인·동반 세션 공지(`factoryLaneNotice`, `kanbanCompanionNotice`), 레인 규칙 필드(`laneNextCardRule`·`laneOwnedCardRule`·`laneManualDispatchRule`), `laneSpawnAuthority`, 스테일 런 공지. 레인과 동반 세션은 카드 1장만 쥔다(REQ-BGS-001).
- SessionStart 방출 경로(`session_start.go`)와 기동 원천 게이트(`startup`만 방출). 리더 공지가 언제 나가는지는 바꾸지 않는다.
- `internal/hook/CLAUDE.md:13`의 정적 가드 문구 정정. 문구가 이 트리에서 거짓이라는 점은 사실로 공시했을 뿐(§A.2 #10) 이 SPEC에서 고치지 않는다.

### Out of Scope — Q4가 이름 붙인 두 파일 밖의 Kickoff 문구

- REQ-BGS-014는 `spec-assembly.md`와 `workflows/moai.md` 둘만 덮는다. 같은 종류의 낡은 문구(카드별 운영자 답을 단정하거나 "HUMAN GATE"/"mandatory human gate" 표지를 쓰는 문장)가 다른 파일에도 있음을 읽어서 확인했다. 파일별 분류는 `plan.md` 표면 목록과 `research.md` R4.4가 가진다. 이 SPEC은 그 파일들을 고치지 않고, 범위를 넓힐지는 decision-index Q12로 남긴다.

### Out of Scope — 결과 수치 주장

- "게이트 왕복 176→10~20회"를 달성했다는 단언. 지표 정의와 M0 기준선까지가 이 SPEC의 몫이고, 감소의 판정은 기준선 뒤 측정으로 따로 한다.

## §E Gaps and Residual Risks

- **G-1 — 모집단 크기 미측정.** 자율 전환 뒤 운영자 형태로 남는 비예약 행이 얼마인지 모른다. 예약 행(keep-set, 리더 보유 권한)과 차단 행이 단일 승인에서 빠지므로, 일괄화가 줄일 수 있는 라운드는 *남은 비예약 Kickoff 행과 동일 판정 주체 공유 행*으로 한정된다. 카드별 선호가 디스크에 없는 카드는 요약 승인 뒤에도 그 카드의 선호 질문이 따로 필요하다(REQ-BGS-020). 모집단이 작으면 효과도 작다. M0 기준선이 이 크기를 처음으로 드러낸다.
- **G-2 — 176의 재현 불가.** 이번 플랜에서 primary 프로젝트의 전사본 전체(1020개 파일)를 이벤트 시각으로 일별 계수했더니 09-26~09-29 합이 124였고(모든 `AskUserQuestion` 호출, 게이트 한정 아님), 09-30 17, 10-01 27이었다(`research.md` R3의 명령). 제안서의 176과 다르고, 분류·파일 범위·중복(세션 이어받기 사본) 한계 때문에 어느 쪽도 기준선이 아니다.
- **G-3 — 모델 매개 독트린.** 문서 규칙이라 기계 강제가 없다. 요약이 규칙을 어겼는지는 결정 기록 줄의 부재·불일치를 sync 감사가 재독해야 드러난다(탐지이지 예방이 아니다). 기록은 쓴 세션 자신이 증명하는 자기 진술이다(`auto-semantics.md` §10).
- **G-4 — 가드 테스트의 어휘적 한계.** 정본 절의 앵커 문구 점검은 변이로 반증할 수 있으나(모순 문장을 앵커 옆에 둔 변이는 통과시킬 수 있다), 의미 판정은 시나리오 AC(AC-011~013, AC-017)를 감사가 읽어야 한다. AC-007·AC-015는 하위 테스트가 변이 거부 줄을 실제로 내는지를 출력에서 센다(빈 하위 테스트는 줄을 못 낸다). 점검 없이 줄만 찍는 하위 테스트는 출력으로는 구별되지 않는다 — sync 감사가 테스트 소스를 읽는다.
- **G-5 — 요약 본문의 휘발성.** 렌더된 요약은 응답 본문에만 남고, 지속 흔적은 카드별 결정 기록이다. "승인이 정확히 나열된 행에만 미쳤다"는 기록의 존재와 공통 `batch=` 식별자로 사후 확인한다. 같은 `batch=` 값이 두 요약에 쓰이는 경우는 한 호출에 요약 질문 하나(REQ-BGS-005)와 결정 보드 점검(REQ-BGS-009)이 막는다. 점검 뒤에 쓰인 기록과의 충돌은 sync 감사가 기록을 읽어 찾는다.
- **G-6 — 질문 채널이 없는 하니스.** 요약 보고와 질문은 하니스의 질문 채널이 있어야 한다. 없으면 같은 요약이 blocker 보고서에 실려 나간다(`AGENTS.md` capability binding 행).
- **G-7 — 알림 문장 가드의 어휘적 한계.** AC-015의 점검은 포인터 토큰, 이름 토큰(`batch gate summary`를 네 로케일에 그대로 둔다 — 포인터 주소와 같은 취급이며 주변 문장의 자연스러움은 아래 sync 감사 판정이다), 금지 토큰, 로케일·공지 완비를 기계로 본다. 문장이 포인터를 둔 채 정본과 반대되는 뜻(예: 카드마다 묻는다)을 말하는 변이는 통과할 수 있다. 의미와 ko·ja·zh 문장의 자연스러움(`native-idiom-and-register.md`)은 sync 감사가 네 로케일 문장을 읽어 판정한다.
- **G-8 — "알림 문장에 한정"의 기계 점검 범위.** AC-009는 변경된 `.go` 파일 목록(일곱 줄)을 기계로 점검한다. 각 제품 파일의 변경이 알림 문장 필드와 합치는 한 줄에 한정되는지는 `git diff --numstat` 공시를 감사가 읽어 판정한다(기계 점검 없음). 이 읽기 단계가 없으면 허용 파일 안의 다른 제품 변경은 목록 점검을 통과한다.
- **G-9 — 포인터만 남김 점검의 어휘적 한계.** AC-016은 정본 전용 토큰(`counter_refs=`, `searched=`)이 정본 외 파일에 없는지, 포인터가 있는지, 바이트·줄 증가가 한도 안인지를 기계로 본다. 두 토큰을 피해 정본의 뜻을 풀어 쓴 문단은 통과할 수 있다. 편집된 헌크를 sync 감사가 읽어 판정한다.
- **G-10 — 기준선 점검의 어휘적 한계.** AC-008은 기준선 파일의 네 요소 레이블 줄이 비어 있지 않은 본문과 함께 있는지를 본다. 레이블마다 의미 없는 한 글자 본문은 통과할 수 있다. 기준선의 내용(명령이 실제로 그 값을 냈는가)은 sync 감사가 파일을 읽어 판정한다. 파일 존재·무시 여부·단독 커밋·단일 커밋 이력·순서는 기계 점검이다.
- **R-1 — 승인 품질 희석.** 행이 많아지면 훑어보기 승인이 된다. 완화: 반대 증거 행 우선 정렬(REQ-BGS-004), `none searched=` 의무(REQ-BGS-008), 행별 빼내기(REQ-BGS-006), keep-set 점검 필드(REQ-BGS-004). 한 번에 몇 행까지 허용할지의 상한은 정하지 않았다(측정 근거 없음).
- **R-2 — 템플릿 중립성.** 정본 절은 템플릿으로 배포되므로 카드 id·SPEC ID·날짜를 쓰지 않는다(`TestTemplateNoInternalContentLeak`). 리더 공지 문장도 같은 제약을 받는다(REQ-BGS-016). 배치 식별자 형식은 자리표시자만 쓴다(REQ-BGS-009).
- **R-3 — 비호출 세션의 비용과 의도된 중복.** 리더 공지 문장은 상시 로딩 파일이 아니다. 리더 세션의 `startup` SessionStart에서만 방출되고 resume·clear·compact·fork에서는 방출되지 않는다. 리더가 아닌 세션이 치르는 바이트는 0이다. 반면 리더는 `kanban-dispatch.md`를 매 턴 이미 읽으므로 이 문장은 같은 규칙을 한 번 더 싣는 *의도된 중복*이다(Q5 행이 공시했고 운영자가 범위 안으로 판정했다). 렌더된 리더 공지의 바이트 증가는 측정하지 않았다(UNVERIFIED).
- **R-4 — 포인터 부패.** 알림이 가리키는 정본 절의 번호나 이름이 바뀌면 알림이 허공을 가리킨다. 완화: AC-015가 포인터 토큰을 네 로케일·두 공지에 고정한다. Q10(명칭)이 바뀌면 문장과 가드 앵커가 함께 바뀐다.
- **R-5 — 열린 행의 후속 변경.** Q1–Q3, Q6, Q7, Q9, Q10, Q12, Q13 중 하나라도 초안과 다르게 판정되면 해당 요구사항 문구와 가드 앵커(`plan.md` M1 표)가 바뀐다. 요구사항 여유는 5개(20/25), 인수 기준 여유는 8개(17/25)다.
