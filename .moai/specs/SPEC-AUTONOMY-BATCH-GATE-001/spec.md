---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "승인 게이트 배치 일괄화 — 운영자 형태 게이트의 배치 게이트 요약, 단일 승인, 반대 증거 의무"
version: "0.2.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1344)
priority: P2
phase: "v3.3.0 target"
module: ".claude/rules/moai/workflow, .claude/skills/moai/workflows, internal/template/templates, internal/hook"
lifecycle: spec-anchored
tags: "batch-gate, approval-gate, kickoff, counter-evidence, decision-record, fail-closed, auto-semantics, launcher-notice, card-t1344"
tier: M
card: t1344
related_specs: [SPEC-LANE-STALL-WATCHDOG-001, SPEC-AUTONOMY-KICKOFF-CALIB-001, SPEC-AUTONOMY-GATE-REWIRE-001, SPEC-AUTONOMY-ESCALATION-001, SPEC-SERVED-MODEL-AUDIT-001]
---

# SPEC-AUTONOMY-BATCH-GATE-001: 승인 게이트 배치 일괄화

> 표기 관례: 서술은 한국어, §C의 GEARS 요구사항 문장과 식별자·경로·명령은 영어다. 형제 SPEC(`SPEC-LANE-STALL-WATCHDOG-001`, `SPEC-CODEX-DEBUG-MODE-001`)과 같은 방식이다.

## HISTORY

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
| 10 | `internal/hook/CLAUDE.md:13`은 정적 가드 `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/` 에서 테스트 파일을 뺀 결과가 "0 matches"여야 한다고 쓰지만, 이 트리에서 그 명령은 22줄을 낸다(주석 21줄 + `pre_tool.go:834`의 도구 이름 문자열 비교 1줄). 알림 소스 4파일의 `AskUserQuestion` 토큰 수는 현재 각 0이다. | 명령과 출력 전문은 `plan.md` E13 |

### A.3 용어

- **배치 게이트 요약(batch gate summary)**: 같은 게이트 행에서 운영자 결정을 기다리는 카드 여럿을, 카드별 증거 행으로 한 번에 보여 주고 승인 질문을 하나로 묶는 제시 형식이다. 승인 수단이 아니라 *제시와 질문의 모양*이며, 증거 기준은 낮추지 않는다.
- **운영자 형태(operator form)**: `auto-semantics.md` §9.1이 "keep-set 사례는 운영자 답을 유지한다"고 남겨 둔 형태. 승인 질문은 하니스의 질문 채널로만 나간다.
- **게이트 행(gate row)**: `auto-semantics.md` §9 게이트 목록의 한 줄.
- **예약 행 / 차단 행**: 단일 승인에 넣을 수 없는 행. 예약은 성격 때문(§B.5), 차단은 증거 상태 때문(§B.5)이다.
- **용어 충돌 정리**: `/moai:todo --auto`의 "batch approval"(큐 소비 권한 부여)과 이 SPEC은 무관하다. 이 SPEC은 그 절을 건드리지 않고, 새 이름 "batch gate summary"로 구별한다(이름 자체는 decision-index Q10).

### A.4 전제 정정과 이 SPEC이 하지 않는 주장

- 이 SPEC은 "게이트 왕복 176→10~20회"를 검증된 결과로 단언하지 않는다. 176의 게이트별 내역이 없고, 이후 도입된 자율 전환이 얼마나 줄였는지도 측정되지 않았다(§B.6).
- 제안서의 "t1318 실측 owner conf 0.38"은 UNVERIFIED다. 이 트리와 primary 보고서에서 독립 출처를 찾지 못했다. 이 SPEC은 그 수치에 기대지 않고 3등급 *규칙*만 쓴다.
- 코드 안의 상시 로딩 예산 주석(`@MX:CEILING: ... 168 tokens`, `internal/config/token_budget_guard.go`)은 낡았다. 이번 실행에서 `TestAlwaysLoadedTokenBudget`가 `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)`를 냈다(`plan.md` 근거 절).
- `internal/hook/CLAUDE.md:13`의 "0 matches" 정적 가드 문구는 이 트리에서 거짓이다(§A.2 #10). 이 SPEC은 그 문구를 근거로도, 고칠 대상으로도 삼지 않는다. 알림 문장의 제약은 "질문 도구를 호출하지 않는다"이고, 점검은 알림 소스 4파일의 토큰 수와 렌더된 알림 문자열에 건다(AC-015).

### A.5 Tier 재측정 (사실, 0.2.0 개정 시점)

- **계획된 변경 파일은 18개**다(라이브와 템플릿 미러를 따로 센 수). 구성: 규칙·스킬 문서 5쌍 = 10(`auto-semantics.md`, `kanban-dispatch.md`, `run.md`, `spec-assembly.md`, `workflows/moai.md`), 템플릿 가드 테스트 1(신규), 기준선 산출물 1(신규), 제품 Go 4(`session_start_kanban.go`, `session_start_kanban_i18n.go`, `session_start_factory.go`, `session_start_factory_i18n.go`), 훅 테스트 2(`session_start_kanban_i18n_test.go` 수정, `session_start_leader_gate_notice_test.go` 신규). 미러 쌍을 한 파일로 세면 13개다.
- `spec-workflow.md` § SPEC Complexity Tier 표: Tier M은 5–15개, Tier L은 15개 초과다. 표는 미러 쌍을 어떻게 세는지 말하지 않는다. 따로 세면 M 범위를 넘고(18 > 15) 한 파일로 세면 안이다(13 ≤ 15). LOC 축은 근거 diff가 없어 UNVERIFIED이며 M 범위(300–1000)로 추정한다.
- 요구사항 16개(Tier M 상한 16과 같다), 인수 기준 15개(상한 16). Tier L이면 상한은 둘 다 25이고, 산출물에 `design.md`와 `research.md`가 더해지며, plan-auditor PASS 기준은 0.80에서 0.85로 오른다.
- 이 개정은 `tier:`를 바꾸지 않았다. 판정은 오케스트레이터 소관이다(`decision-index.md` Q11).
- 위 목록은 계획 시점의 열거다. run-phase 뒤 `git diff --name-only develop...HEAD`(`.moai/specs/`와 `.moai/reports/` 제외)로 실제 수를 다시 잰다.

## §B Decisions

### B.1 해법 개요 — 문서 + 리더 공지 문장 하나, 가드 테스트 둘

운영자 형태로 남는 게이트 라운드를 *제시 단위*로 묶는다. 리더(운영자 대화를 쥔 세션)가 같은 게이트 행에서 대기 중인 카드들을 읽어 요약 보고서를 쓰고, 질문 하나로 승인받고, 카드마다 자기 결정 기록을 남긴다. 바꾸는 것은 문서(정본 절, 포인터, 낡은 표현 정합)와, 런처가 SessionStart에 주입하는 *리더 공지의 문장 하나*다. 제품 Go 변경은 그 공지 문자열과 그것을 고정하는 테스트 기대값에 한정된다(§B.7). 다중 카드 적용 수단(`factory decide <card>...`)은 이미 있고, 결정 기록은 형제 SPEC과 같은 방식으로 오케스트레이터가 디스크에 줄로 쓴다. 안전 불변식(fail-closed, keep-set 제외)이 문서에서 조용히 사라지지 않게 하는 *테스트 전용* Go 파일(템플릿 쪽 정본 절 가드)과, 두 리더 공지가 문장을 네 로케일 모두에 싣는지 점검하는 *테스트 전용* Go 파일(훅 쪽)을 추가한다.

### B.2 배치 단위와 구성원

- 키는 **같은 게이트 행**이다. Kickoff 행과 sync 승인 행은 한 질문에 섞이지 않는다.
- **대기 중인 행 전부를 스냅숏으로 제시**한다. 요약을 채우려고 준비된 카드를 붙잡지 않는다(REQ-BGS-002). 늦게 준비된 행은 다음 요약이나 개별 질문이다.
- 하나의 판정이 여러 카드의 공통 결정 주체일 때(§A.2 #2)는, 같은 질문을 카드 수만큼 묻지 않고 한 번 묻고 영향받는 카드를 전부 적는다(REQ-BGS-003). 예약 행에 이 규칙을 확장할지는 Q2다.
- 묶는 주체는 리더 세션뿐이다. 레인은 카드 1장만 쥐므로 교차 카드 배치를 만들지 않는다(REQ-BGS-001).

### B.3 요약 형식과 단일 승인

보고 → 질문 순서는 `askuser-protocol.md`의 Report-Before-Ask를 따른다. 행마다 카드 id, 게이트 행, SPEC id, 감사 판정 참조와 점수·PASS 기준 대비 여유, 산출물 해시 불변 확인, 카드 자신의 결정 기록 참조, `counter_refs=`를 싣는다. 반대 증거가 있는 행을 먼저 놓는다. 질문은 하나이고, 승인은 *나열된 승인 가능 행에만* 미친다. 행마다 빼내기(pull-out)가 항상 열려 있고, 빼낸 행은 개별 처리로 돌아간다. 권고 라벨은 `interview.recommendation_mode`의 현행 규약을 그대로 따른다(`pull` 동안 라벨 없음).

### B.4 반대 증거 의무

행마다 "진행하지 말아야 할 가장 강한 증거"를 `counter_refs=`로 적는다. 후보 원천은 닫힌 목록이다: 감사 경고·기록된 부채, PASS 기준 대비 여유, 미해결 decision-index 행, 감사 교차의 의견 불일치, 열린 차단·대기 기록, 같은 요약 안의 다른 행과의 경로 겹침. 못 찾았으면 `counter_refs=none searched=<검색 내용>`이다. 검색 내용 없는 `none`은 반대 증거 진술로 치지 않는다. 목록을 닫을지 열어 둘지는 Q7이다.

### B.5 요약에 들어가지 않는 행

- **예약 행**: keep-set 세 범주(환경상 불가능, 운영자 보유, 외부 공유 시스템에 대한 되돌릴 수 없는 조작)와, 권한 게이트가 운영자·리더에게 남겨 둔 결정(최종 판정, 병합 승인, 큐 변경, 운영자 게이트, 사용자 표면 동작 변경, CodeRabbit 슬롯 대기 판정)은 개별 승인을 유지한다. 배포 문서에서 이 목록의 대응 표현은 `auto-semantics.md` §9의 keep-set과 `kanban-dispatch.md:106`의 리더 보유 권한 조항이다(위치 결정은 Q6).
- **차단 행**: 감사 판정이 FAIL·INCONCLUSIVE·부재이거나, 판정 이후 계획 산출물 해시가 바뀌었거나, 차단 요인이 열려 있으면 *차단 행으로 표시*하고 승인에서 뺀다. `auto-semantics.md` §7의 권한 게이트 불변식(부정·불확정 판정은 fail-closed)을 요약이 약화하지 못한다.
- **contract 모드**: `workflow.autonomy.mode: contract`에서 plan→run 게이트는 계약 서명(`moai contract kickoff-check`)이며 요약 행이 생기지 않는다(`contract-autonomy.md` § The signing gate).

### B.6 성공 지표와 측정의 정직한 한계

- **지표**: (M-1) 배치 창당 운영자 형태 게이트 질문 수, (M-2) 게이트 질문 한 번당 결정된 카드 수, (M-3, 안전 반지표) 요약으로 승인됐는데 `counter_refs=`가 없는 기록 수와 FAIL·INCONCLUSIVE 행이 승인된 수 — 둘 다 0이어야 한다.
- **측정 수단**: M-1은 세션 전사본의 `AskUserQuestion` 호출 계수 + 질문 문구로 게이트 질문 분류(`plan.md` 근거 절에 이번 실행의 명령과 출력). 생산용 도구는 없다. 분류는 문구 휴리스틱이라 *가설*이다(`verification-claim-integrity.md` §1). M-3은 결정 기록 줄을 `grep`으로 읽는다(디스크 SSOT, §10).
- **기준선**: 이 트리에는 게이트 질문만 따로 센 기준선이 없다. 이번 플랜에서 잰 임시 계수는 모든 `AskUserQuestion` 호출(게이트+명확화)의 일별 합이라 기준선이 아니다. 기준선은 run-phase M0에서 변경 *앞선 별도 커밋*으로 남긴다(REQ-BGS-015, `verification-claim-integrity.md` §2.3).
- **결론 유보**: 도입 전 수치와 그 이후를 비교해 감소를 주장하려면 M0 기준선이 먼저 있어야 한다.

### B.7 구현 형태 결정 (Core Behavior 4)

- 제품 Go 변경은 **리더 공지 문자열에 한정**된다(운영자 판정 Q5, 2026-10-02). 파일은 정확히 넷이다: `internal/hook/session_start_kanban.go`(칸반 리더 공지 조립), `session_start_kanban_i18n.go`(칸반 로케일 표), `session_start_factory.go`(팩토리 리더 공지 조립), `session_start_factory_i18n.go`(팩토리 로케일 표). 변경 내용은 로케일 표에 문장 필드 하나와 네 로케일 값을 더하고, 조립부에서 그 필드를 *기존 블록 안의 한 줄*로 합치는 것이다. 새 블록·새 분기·새 환경 변수·새 호출은 없다. 레인·동반 세션 공지, `laneSpawnAuthority`, 기동 원천 게이트, 방출 경로(`session_start.go`)는 그대로다.
- 두 리더 공지에 **모두** 싣는다. 근거: 두 공지는 배타적이라 한쪽에만 두면 다른 쪽 리더가 문장을 받지 못하고, 두 리더 모두 운영자 대화를 쥔다(§A.2 #9). 레인·동반 세션은 카드 1장만 쥐므로(REQ-BGS-001) 받지 않는다.
- 추가·수정하는 테스트 Go 파일은 셋이다. (a) 기존 `session_start_kanban_i18n_test.go`에는 새 필드를 필드 빈값 점검 표에 더하는 덧붙임만 한다. (b) 새 `session_start_leader_gate_notice_test.go`는 두 리더 공지 × 네 로케일을 렌더해 문장 규칙을 점검하고 변이 본문이 거부됨을 보인다. (c) 새 `internal/template/batch_gate_summary_doctrine_test.go`는 정본 절의 안전 불변식 앵커를 점검하고, 앵커를 하나씩 뺀 변이 본문이 각각 거부됨을 보여 준다(관측된 실패, `verification-completeness.md` §1.1·§2). 셋 다 런타임 동작을 바꾸지 않는다.
- 다중 카드 적용은 `factory decide <card>...`가 이미 하고, 결정 기록은 줄 기록이며, 형제 SPEC도 같은 선(문서 중심)을 지켰다. 강제 장치(결정 기록 자동 기록, 요약 검증기)를 Go로 만들지는 Q8이며 열린 채 범위 밖이다.

### B.8 열린 결정

`decision-index.md` 참조(Q1–Q11, Q11만 POLICY-COVERED). Q4·Q5는 운영자가 범위 안으로 판정했다(2026-10-02, Decision Point 1): REQ-BGS-014는 무조건이 되었고 REQ-BGS-016이 더해졌다. 열린 행은 Q1–Q3, Q6–Q10이다. 현 초안의 요구사항은 그 행들에 대한 초안의 읽기를 따르며(Q1·Q2·Q3·Q10 행이 초안이 택한 읽기를 적었고, Q6·Q7은 초안이 닫힌 목록과 배포 대응 표현을 쓴다는 점이 §B.4·§B.5에 있다), 답이 다르면 해당 REQ 문구와 가드 앵커가 바뀐다. Q8(Go 강제)은 열린 채 범위 밖이다. Q11은 정책이 덮는 행이지만 측정된 사실이 바뀌었다(§A.5).

## §C Requirements

검증 계층은 `acceptance.md`다(릴리스 차단 기준은 RED-now 명령·원문 출력·종료 코드·트리 고정을 갖는다). 아래는 GEARS 요구사항 계층이다. 모듈 6개, 요구사항 16개(Tier M 상한 16과 같다 — Tier L 상한은 25, §A.5).

### C.1 Batch unit and membership

- **REQ-BGS-001** (Ubiquitous) — The session that holds the operator dialogue (the leader session in Kanban or Factory Mode, otherwise the orchestrating session) shall consolidate pending operator-form decisions into one batch gate summary only when the rows share the same gate row of the auto-semantics gate inventory; a lane session shall present only its own card's gate and shall not form a cross-card batch.
- **REQ-BGS-002** (State-driven) — **While** a card's operator-form gate is ready for decision, the session shall not hold that card back to wait for further rows; a row that becomes ready after a summary has been presented shall join the next summary or be asked individually.
- **REQ-BGS-003** (Event-driven) — **When** one judgment is the common decision subject of two or more pending rows that are not reserved under REQ-BGS-010, the session shall ask that judgment once, naming every affected card in the question and in the report that precedes it.

### C.2 Summary format and the single approval

- **REQ-BGS-004** (Ubiquitous) — The batch gate summary shall be rendered as a findings report in the response body before the decision question, with one row per card carrying the card id, the gate row, the SPEC id, the audit verdict reference with its score and its margin to the tier's PASS threshold, the plan-artifact-hash-unchanged check, the reference to the card's own decision record, and the counter-evidence field of REQ-BGS-007; rows carrying counter-evidence shall be listed before rows without it.
- **REQ-BGS-005** (Ubiquitous) — The report shall be followed by exactly one decision question whose approval covers the listed approvable rows only, and the report shall state that rows added later and reserved or blocked rows are not covered by it.
- **REQ-BGS-006** (Ubiquitous) — The decision question shall always offer a way to pull any single row out of the approval for individual handling, and pulling a row out shall leave the approval of the remaining rows intact.

### C.3 Counter-evidence and per-row records

- **REQ-BGS-007** (Ubiquitous) — Each row shall carry a `counter_refs=` field naming the strongest evidence against proceeding, drawn from audit warnings or recorded debt, the margin to the PASS threshold, unresolved decision-index rows, divergent audit-cross opinions, open blockers or wait records, and path overlap with another row of the same summary.
- **REQ-BGS-008** (Event-detected) — **When** no counter-evidence is found for a row, the row shall read `counter_refs=none` followed by `searched=` naming the search that found none, and a bare `none` shall not count as a counter-evidence statement.
- **REQ-BGS-009** (Event-driven) — **When** a row is approved through the summary, the session shall write that card's own decision record in the auto-semantics decision-record form, additively carrying `counter_refs=` and naming the batch in its ladder path, and a row without its own record shall not be treated as approved.

### C.4 Reserved and blocked rows

- **REQ-BGS-010** (Unwanted) — The single approval shall not cover a row in a keep-set category (environment-impossible, operator-held, or an irreversible operation on an external shared system) or a row whose decision the authority gates reserve to the operator or the leader (final verdicts, merge approval, queue mutations, operator gates, user-facing behavior changes, CodeRabbit slot-wait judgment); such rows shall keep their individual approval.
- **REQ-BGS-011** (Event-driven) — **When** a row's audit verdict is FAIL, INCONCLUSIVE, or absent, its plan-artifact hashes changed since the verdict, or a blocker is open, the session shall show the row as blocked and exclude it from the approval, so that the summary never weakens the authority-gate invariant of the decision ladder.
- **REQ-BGS-012** (State-driven) — **While** `workflow.autonomy.mode` is `contract`, the summary shall not stand in for the signing gate: the contract signature verified by `moai contract kickoff-check` remains the plan→run gate and forms no summary row.

### C.5 Home, consistency, and baseline

- **REQ-BGS-013** (Ubiquitous) — The doctrine shall have one home, a new subsection of the gate-inventory section of `auto-semantics.md`, byte-identical in the template tree; every other surface shall carry a pointer only, an always-loaded file shall not grow by more than 1,000 bytes, and the run workflow skill shall gain no line.
- **REQ-BGS-014** (Ubiquitous) — The Kickoff wording of the plan-assembly skill and of the workflow router shall describe the plan→run Kickoff in its default-autonomous form, citing the default-autonomous transition and the batch gate summary by path and section, shall not state that a plan-auditor PASS or a skip-eligible score never substitutes for the operator question on a card outside the keep-set, and shall keep the strings the Kickoff preservation guards pin.
- **REQ-BGS-015** (Ubiquitous) — The run phase shall land a measured baseline of question-tool gate rounds in its own commit, ahead of the doctrine commit, naming the measuring command, the observed output, the classification method, and their limits.

### C.6 Launcher leader notice

- **REQ-BGS-016** (Ubiquitous) — The SessionStart notice of a leader session — the factory leader and the kanban leader, each in its agent-facing and its operator-facing copy and in each of the en, ko, ja, and zh locales — shall carry one sentence that names the batch gate summary and points at its doctrine home in `auto-semantics.md` by path and section, shall not restate the doctrine, shall not name the question tool, shall carry no card id, SPEC id, or date, shall not appear in a lane or companion notice, and shall leave every previously pinned notice string intact.

## §D Out of Scope

### Out of Scope — 리더 공지 문장 밖의 제품 Go 코드와 CLI 동사

- 리더 공지 문장(REQ-BGS-016)과 그것을 고정하는 테스트 기대값을 제외한 제품 Go 변경 전부. 허용되는 제품 Go 파일은 `internal/hook/session_start_{kanban,factory}{,_i18n}.go` 넷뿐이다(§B.7).
- `moai factory decide`의 확장(결정자 종류 추가, 결정 기록 자동 기록, 새 일괄 동사)과 MCP `factory_decide`의 다중 카드화. 필요성 판단은 Q8이며 열린 채 이 SPEC의 범위 밖이다. 이 SPEC은 배치 기구의 Go 강제 장치를 만들지 않는다.
- 게이트 라운드를 세는 생산용 측정 도구. 이 SPEC은 지표 정의와 임시 측정 명령, 그리고 한계 공시까지만 다룬다.

### Out of Scope — 큐 진입과 카드 선택

- 큐 생산(admission)은 운영자 몫이다. `/moai:todo --auto`의 "batch approval"(큐 직렬 소비 권한)은 그대로이고, `kanban-dispatch.md` "Promotion is the operator's act, always." 부터 "The self-dispatch lane exception." 까지의 구간은 바이트 단위로 건드리지 않는다(`TestAutoRankDoctrineAmendment`·`TestAutoRankMirrorParity`가 고정).

### Out of Scope — keep-set 게이트 자체의 기계 동작

- push 승인, 병합 승인, `abandon`, `hold` 해제 등은 개별 승인을 유지한다. 요약은 이들을 *보여 줄 수는* 있어도 단일 승인에 넣지 않는다(REQ-BGS-010).

### Out of Scope — Jev 능력과 3등급 독트린의 정의

- 3등급 독트린은 로컬 유지보수 문서(`.moai/docs/jev-local-operations.md`)의 것이다. 이 SPEC은 그 문서를 고치지 않고 배포 문서에 대응 표현을 둔다. `workflow.jev.enabled`·`jev_ask` 표면은 건드리지 않는다.

### Out of Scope — 리더 공지 밖의 주입 텍스트

- 레인·동반 세션 공지(`factoryLaneNotice`, `kanbanCompanionNotice`), 레인 규칙 필드(`laneNextCardRule`·`laneOwnedCardRule`·`laneManualDispatchRule`), `laneSpawnAuthority`, 스테일 런 공지. 레인과 동반 세션은 카드 1장만 쥔다(REQ-BGS-001).
- SessionStart 방출 경로(`session_start.go`)와 기동 원천 게이트(`startup`만 방출). 리더 공지가 언제 나가는지는 바꾸지 않는다.
- `internal/hook/CLAUDE.md:13`의 정적 가드 문구 정정. 문구가 이 트리에서 거짓이라는 점은 사실로 공시했을 뿐(§A.2 #10) 이 SPEC에서 고치지 않는다.

### Out of Scope — 결과 수치 주장

- "게이트 왕복 176→10~20회"를 달성했다는 단언. 지표 정의와 M0 기준선까지가 이 SPEC의 몫이고, 감소의 판정은 기준선 뒤 측정으로 따로 한다.

## §E Gaps and Residual Risks

- **G-1 — 모집단 크기 미측정.** 자율 전환 뒤 운영자 형태로 남는 비예약 행이 얼마인지 모른다. 예약 행(keep-set, 권한 예약)과 차단 행이 단일 승인에서 빠지므로, 일괄화가 줄일 수 있는 라운드는 *남은 비예약 행과 동일 판정 주체 공유 행*으로 한정된다. 모집단이 작으면 효과도 작다. M0 기준선이 이 크기를 처음으로 드러낸다.
- **G-2 — 176의 재현 불가.** 이번 플랜에서 primary 프로젝트의 전사본 전체(1020개 파일)를 이벤트 시각으로 일별 계수했더니 09-26~09-29 합이 124였고(모든 `AskUserQuestion` 호출, 게이트 한정 아님), 09-30 17, 10-01 27이었다(`plan.md` 근거 절의 명령). 제안서의 176과 다르고, 분류·파일 범위·중복(세션 이어받기 사본) 한계 때문에 어느 쪽도 기준선이 아니다.
- **G-3 — 모델 매개 독트린.** 문서 규칙이라 기계 강제가 없다. 요약이 규칙을 어겼는지는 결정 기록 줄의 부재·불일치를 sync 감사가 재독해야 드러난다(탐지이지 예방이 아니다).
- **G-4 — 가드 테스트의 어휘적 한계.** 정본 절의 앵커 문구 점검은 변이로 반증할 수 있으나(모순 문장을 앵커 옆에 둔 변이는 통과시킬 수 있다), 의미 판정은 시나리오 AC(AC-011~013)를 감사가 읽어야 한다.
- **G-5 — 요약 본문의 휘발성.** 렌더된 요약은 응답 본문에만 남고, 지속 흔적은 카드별 결정 기록이다. "승인이 정확히 나열된 행에만 미쳤다"는 기록의 존재로 사후 확인한다.
- **G-6 — 질문 채널이 없는 하니스.** 요약 보고와 질문은 하니스의 질문 채널이 있어야 한다. 없으면 같은 요약이 blocker 보고서에 실려 나간다(`AGENTS.md` capability binding 행).
- **G-7 — 알림 문장 가드의 어휘적 한계.** AC-015의 점검은 포인터 토큰, 금지 토큰, 로케일·공지 완비를 기계로 본다. 문장이 포인터를 둔 채 정본과 반대되는 뜻(예: 카드마다 묻는다)을 말하는 변이는 통과할 수 있다. 의미와 ko·ja·zh 문장의 자연스러움(`native-idiom-and-register.md`)은 sync 감사가 네 로케일 문장을 읽어 판정한다.
- **G-8 — "알림 문장에 한정"의 기계 점검 범위.** AC-009는 변경된 `.go` 파일 목록(일곱 줄)을 기계로 점검한다. 각 제품 파일의 변경이 알림 문장 필드와 합치는 한 줄에 한정되는지는 `git diff --numstat` 공시를 감사가 읽어 판정한다(기계 점검 없음). 이 읽기 단계가 없으면 허용 파일 안의 다른 제품 변경은 목록 점검을 통과한다.
- **R-1 — 승인 품질 희석.** 행이 많아지면 훑어보기 승인이 된다. 완화: 반대 증거 행 우선 정렬(REQ-BGS-004), `none searched=` 의무(REQ-BGS-008), 행별 pull-out(REQ-BGS-006). 한 번에 몇 행까지 허용할지의 상한은 정하지 않았다(측정 근거 없음).
- **R-2 — 템플릿 중립성.** 정본 절은 템플릿으로 배포되므로 카드 id·SPEC ID·날짜를 쓰지 않는다(`TestTemplateNoInternalContentLeak`). 리더 공지 문장도 같은 제약을 받는다(REQ-BGS-016).
- **R-3 — 비호출 세션의 비용과 의도된 중복.** 리더 공지 문장은 상시 로딩 파일이 아니다. 리더 세션의 `startup` SessionStart에서만 방출되고 resume·clear·compact·fork에서는 방출되지 않는다. 리더가 아닌 세션이 치르는 바이트는 0이다. 반면 리더는 `kanban-dispatch.md`를 매 턴 이미 읽으므로 이 문장은 같은 규칙을 한 번 더 싣는 *의도된 중복*이다(Q5 행이 공시했고 운영자가 범위 안으로 판정했다). 렌더된 리더 공지의 바이트 증가는 측정하지 않았다(UNVERIFIED).
- **R-4 — 포인터 부패.** 알림이 가리키는 정본 절의 번호나 이름이 바뀌면 알림이 허공을 가리킨다. 완화: AC-015가 포인터 토큰을 네 로케일·두 공지에 고정한다. Q10(명칭)이 바뀌면 문장과 가드 앵커가 함께 바뀐다.
