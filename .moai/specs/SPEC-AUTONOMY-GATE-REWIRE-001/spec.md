---
id: SPEC-AUTONOMY-GATE-REWIRE-001
title: "계약 기반 자율 하네스 A3 — contract 모드 게이트 재배선, Kickoff 자율 승인, contract decide·kickoff-check·revoke"
version: "0.3.6"
status: in-progress
created: 2026-09-26
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.md, .claude/rules/moai/core, .claude/rules/moai/workflow, .claude/skills/moai, .moai/config/sections, internal/template/templates, internal/template, internal/contract, internal/cli, .moai/specs/SPEC-JEV-CORE-001"
lifecycle: spec-anchored
tags: "autonomy, contract-mode, gate-rewire, kickoff, autonomous-kickoff, contract-decide, kickoff-check, contract-revoke, receipt-store, tamper-evidence, jev-doctrine-amendment, guided-preservation, template-mirror, lifecycle, llm-jev-cross-check, jev-llm-fallback"
tier: L
depends_on: [SPEC-AUTONOMY-CONTRACT-001, SPEC-AUTONOMY-ESCALATION-001, SPEC-ALWAYS-LOADED-DIET-002]
related_specs: [SPEC-JEV-CORE-001, SPEC-AUTONOMY-TIERS-001]
---

# SPEC-AUTONOMY-GATE-REWIRE-001 — contract 모드 게이트 재배선 + Kickoff 자율 승인 (카드 t1236, AUTONOMY-A3)

## HISTORY

| 날짜 | 버전 | 변경 | 작성 |
|---|---|---|---|
| 2026-09-27 | 0.3.6 | **plan-audit iter-6(최종) FAIL 0.868 → 리드 결정 PASS-WITH-DEBT** (보고서 `.moai/reports/t1236/plan-audit-6.md`, 감사 대상 `a02e7c02b`, 7차 감사 없음). 처분은 `plan.md §L` 추가 행. REQ·AC 수 25·25 그대로. (D56) `constitution.Validate` 는 `ZONE_UNREGISTERED`·`ANCHOR_NOT_FOUND` 를 내지 않으므로 AC-GR-003 의 미등록 `[HARD]` 반증을 테스트 자체의 `[HARD]` 집합 비교(always-loaded 대상 여섯 사본, 현재 ⊆ BASE)로 바꿈 — 셸 측정 가능성 탐침 결과(GREEN/RED)는 `.moai/reports/t1236/verdict.md`; (D57) plan.md §H 의 낡은 A2 표지; (D58) AC-GR-009 선택 확인 증거의 명령 귀속; (D59) design.md §12 판독기 행의 기대 열 개수 | manager-spec |
| 2026-09-27 | 0.3.5 | **plan-audit iter-5(델타) FAIL 0.841 수리** (보고서 `.moai/reports/t1236/plan-audit-5.md`, 감사 대상 `f40824c82`). 처분은 `plan.md §L`. REQ·AC 수 25·25 그대로. (D47) AC-GR-003 을 비-OK 전 범주 `(sentinel, id)` 비교·세 개수 이하로 넓히고 건너뛰기 변수 제거·비건너뜀·BASE=EV-6 전제 단언, 미등록 `[HARD]` 반증 추가, 판정 명령을 하나로; (D48) 판독기 소비 경로 픽스처 — AC-GR-016 (17)·AC-GR-018 (e) 두 변형; (D49) 허용 목록 24행에 `internal/cli/contract.go`(516행·389~414행); (D50) EV-6 원문을 원시 바이트로 복원; (D51) plan·acceptance 의 옛 A2 표지 정리; (D52) `doc.go` 허용 범위에 135~141행; (D53) REQ-GR-009 추적에 AC-GR-023; (D54) 에스컬레이션 디렉터리 부재 = 기록 0건(해제, REQ-GR-022·AC-GR-023 (r0)); (D55) AC 명령의 `-run` 정규식을 `^…$` 로 고정 | manager-spec |
| 2026-09-27 | 0.3.4 | **M0 재앵커 blocker 처분(리드 결정 B1~B5, 2026-09-27)** — run M0(`progress.md §E.2`, BASE `7fe658815`)이 멈춘 다섯 차이를 제자리에서 고쳤다. REQ·AC 수는 25·25 그대로. (B1) t1175 가 `askuser-protocol.md` 의 「The Five Exceptions」 본문을 `askuser-protocol-reference.md:229` 로 옮기고 `## Ambiguity Triggers and Exceptions` 스텁(206행)만 남겼으므로 `contract-ambiguity` 블록 위치를 그 스텁 문단 뒤로 옮김(`design.md §2` 3행, M3). (B2) `moai-mcp-tools.md` 에는 더 이상 `jev_ask` 행이 없으므로 개정 위치를 `moai-mcp-tools-catalogue.md` 의 두 행(138·216)으로 바꾸고 `moai-mcp-tools.md` +300자 예산을 삭제, 카탈로그는 always-loaded 가 아니므로 조건부 로드 예산(행당 300자)을 따로 둠(REQ-GR-013·025, D-6, AC-GR-010·022). (B3) A2 는 revoke 기록을 `status: resolved` 로 두고 needs-decision 으로 세지 않는다(A2 0.4.3 spec §I.2, `escalation.NeedsDecision`) — A2 는 바꾸지 않고, A3 가 자기 판독기로 `kind: revoke` 기록을 재개 차단 신호로 읽는다(REQ-GR-009 (e)·007·012·020·022, AC-GR-023 이 판독 규칙을 고정). (B4) A1 서명기 단계 (1) 은 `internal/contract/receipt.go` `ReceiptOutcome`(208~232행)에 있으므로 그 자리에서 조건화하도록 허용 목록 23행에 추가(`design.md §2`·`§7.1`). (B5) `moai constitution validate` 는 BASE 에서 이미 exit 1·DRIFT 9건(이 SPEC 범위 밖, 리드가 별도 카드 발행) — AC-GR-003 의 기대를 「BASE 대비 DRIFT 가 늘지 않음」으로 바꾸고 BASE 측정을 증거 원장 EV-6 에 고정 | manager-spec |
| 2026-09-26 | 0.3.3 | 운영자 승인 4차 예외 (D43~D46 한정), 2026-09-26, 리드 전달. 보고서 `.moai/reports/t1236/plan-audit-3.md`, 감사 대상 `710530d67`. (D43) kickoff-check 사유가 여럿이면 `--json` 에 모두 담고 대표 사유는 정해진 우선순위를 따름(REQ-GR-007, AC-GR-016); (D44) A1 서명기 단계 (1) 은 doctrine 플래그를 주입받고 대체 테스트가 매 head 에서 두 상태를 주입해 검사 — 「매 커밋 CI」 전제 삭제(`design.md §2`·`§7.1`, AC-GR-017); (D45) `acceptance.md` 의 `AC-CONTRACT-016` 교차 참조 두 곳에 `[REF]`; (D46) 원천 없는 「moai 측 세션 식별자」를 REQ-GR-012·`design.md §9` 에서 삭제. 함께: 운영자가 `CLAUDE.local.md §29` 한 줄 문안을 그대로 승인(리드 전달) — 연동 커밋 보류 조건 해소 | manager-spec |
| 2026-09-26 | 0.3.2 | **plan-audit iter-2 FAIL 0.77 수리** (보고서 `.moai/reports/t1236/plan-audit-2.md`, 감사 대상 `1b071a573`; must-pass 전부 통과). 결함 D26~D42 처분은 `plan.md §I`. 주요 변경: (D26) kickoff-check 에 `not-signed-valid`·`receipt-not-approved` 사유와 `effective_decider` ≠ `signer_kind` 검사를 넣고 픽스처 (12)~(16) 추가; (D27) 「`reject`·`human` 은 둘 다 서명 거절·사람 결정 필요」를 REQ-GR-004 의 문장으로; (D28) 작성자 배제의 데이터 원천을 `internal/spec` 의 `Authored-By-Agent:` 판독기로 정하고(git 트레일러 파서는 이 저장소 커밋에서 값을 보지 못함 — 실측), 세션 식별자 조건은 원천이 없어 삭제, 트레일러가 없으면 `author-check-unmeasured` → 사람; (D29) 결정하지 않는 decide 경로는 A1 영수증을 쓰지 않고 사건만 남김; (D30) plan-audit 판정을 `plan_artifact_hash:` 로 현재 plan 산출물에 묶고 `PASS-WITH-DEBT` 는 문턱 이상일 때만 통과; (D31) Jev 원칙 개정 위치에 `SPEC-JEV-CORE-001` REQ-JEVC-011 과 `Out of Scope — authority` 두 항목 추가; (D34) R3 와 A1 임시 규칙을 한 상수 `jevDoctrineAmended` 로 조건화하고 A1 AC-CONTRACT-016 (t) 의 대체 테스트를 AC-GR-017 에 둠; D35~D42 수리. **리드 결정 2026-09-26 (D32, D33)**: REQ-GR-025 의 「한 커밋」 유지 — M7b 는 manager-spec(SPEC 본문) → manager-develop(코드·규칙) → 레인 오케스트레이터가 명시 경로로 한 커밋, 동시 작성 없음; `CLAUDE.local.md §29` 는 연동 묶음 안이고 운영자 확인이 없으면 연동 커밋 전체 보류(문안은 `design.md §11.1`). A1 기준선을 **A1 0.5.2 `25283ebf8`** 로 갱신 — §C.8·REQ-CONTRACT-019·024·AC-CONTRACT-016 (t) 가 조건부 문언이 되어 A3 가 A1 SPEC 문언을 고칠 필요가 없어졌으므로 `design.md §2` 27행과 M7b 보류를 없앴다. reject/human 소비자 실측을 `research.md §10.4` 에 기록 | manager-spec |
| 2026-09-26 | 0.3.1 | **운영자 재결정(2026-09-26, 리드 경유) 반영 — Jev 단독 없음, `llm+jev` 교차 확인만.** 같은 날 잠시 지시된 「단일 결정자 `human`·`llm`·`jev`」안은 이 재결정으로 대체됐다. contract 모드의 자율 결정자는 `llm`(기본)과 `llm+jev`, guided 는 `human`. **A3 가 결정 규칙을 소유**(리드 소유 조정): decide 가 영수증 `outcome ∈ {approve, reject, human}` 을 R1 교차 확인(모두 approve → 시작, 모두 reject → reject, 불일치·escalate → 사람), R2 Jev 쪽 실패 5종 → `llm` 단독 대체와 영수증·저장소 기록, R3 원칙 개정 연동 전 Jev 가 답한 결정 → 사람, R4 `llm` 단독(approve/reject/escalate), R5 Jev 단독 거절로 도출한다(REQ-GR-010). `reject` 와 `human` 은 모두 Kickoff 를 사람에게 보낸다. `on_disagree` 와 「A5 전까지 불허」는 삭제, kickoff-check 의 `decider-not-permitted` 는 `jev` 에만 남김(REQ-GR-007). 규칙·갈래마다 RED 먼저의 Go 테스트(AC-GR-019·025). Jev 원칙 개정은 「`llm+jev` 교차 확인의 두 번째 신호」 한 곳으로 좁혔고(REQ-GR-013), **개정·R3 해제·A1 임시 규칙 해제를 같은 커밋으로 묶는 연동 요구**를 신설했다(REQ-GR-025, AC-GR-017) — 이를 위해 revoke 의 두 요구(옛 022·023)를 REQ-GR-022 로 합치고 옛 024·025 를 023·024 로 옮겼다. `moai contract decide`(가칭)의 이름·입력(카드 id, 계약 `card` 필드와 일치)·출력(`$MOAI_HOME/db/<project-key>/contract/` 의 `receipts.jsonl`·`events.jsonl` 추가 + A1 고정 경로 영수증)을 명시하고 AC-GR-020 에서 기계 검사한다(REQ-GR-011·012). 리드 결정 R10(2026-09-26): 저장소 디렉터리 `$MOAI_HOME/db/<project-key>/contract/`. A1 기준선을 **A1 0.5.1 `65e0a9167`** 로 갱신 — 0.5.1 이 결정자 값 집합(`human`·`llm`·`llm+jev`)·파생 기본값·영수증 필드·구조 검증·`card` 필드를 확정해 v0.3.1 초안의 「A1 0.5.1 대기」 목록은 모두 해소됐다(`research.md §10.3`, 이력 보존). 남은 A1↔A3 확인 항목은 `research.md §10.4`. 리드 결정: push 직렬화와 에이전트발 `sign`·`decide` 차단은 A2b(t1245) 소유 — run 전제에 t1245 병합을 더하고 소유 상충을 해소로 옮겼다(`plan.md §C`, `research.md §10.2`). plan-audit 판정 파일의 보수적 선택 규칙은 리드 승인, 명명 불일치는 후속 카드 후보로 범위 밖. 맥락(범위 확장 아님): 자율 모드의 로컬 병합은 자동(F-Q2) | manager-spec |
| 2026-09-26 | 0.3.0 | **plan-audit iter-1 FAIL 0.71 수리** (보고서 `.moai/reports/t1236/plan-audit-1.md`, 감사 대상 `62a65f709`). 결함 D1~D25 처분은 `plan.md §I`. 주요 변경: (1) 미해결 질문 표지 0건 — NC-1·3·5·6·7·8·9 는 본문 결정, NC-2·NC-4 는 리드가 승인한 기본값(운영자가 바꿀 수 있음)으로 기록(D1). (2) REQ 를 001~025 로 연속 재번호하고 모든 참조 갱신(D2·D17). (3) **Kickoff 게이트 판정을 A3 가 Go 로 소유** — `moai contract kickoff-check`(가칭)가 사람 서명 또는 저장소에 있는 `approve` 기록과 일치하는 결정자 쌍 서명만 통과시키고, `llm`·`jev` 단독 서명은 A5 전까지 거절(D3·D6). (4) 등가 조항을 **사람 서명에 한정**; 비인간 서명의 등가는 Frozen 태그 문단 개정을 별도 활성 조건으로 둠(D4). (5) 두 번째 리뷰 정지는 A4(t1237) 소유로 이동(D7). (6) 합의 규칙은 A1 을 정본으로 참조(D8). (7) revoke 대상을 「서명된 계약이 있는 카드」로 정의(D9). (8) guided 보존 주장을 「블록 밖 텍스트 보존」으로 좁히고 guided 문맥 증가량을 수치로 명시, SSOT `paths:` 를 계약 파일로 좁힘(D11). (9) 모든 AC 를 워크트리 세션에서 실행 가능한 Go 테스트 또는 단순 명령으로 재작성(D12). **리드 결정 2026-09-26** 반영: L1 Jev 게이트 입력 개정을 이 카드 범위로 편입(contract 모드 Kickoff 교차 확인 한정, REQ-013), L2/D10 에스컬레이션 기록은 A2 최종 형식 `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` + 기계 판독 YAML 머리 + `revoke` 종류(「A2 개정본으로 재확인」), L3 주 LLM 결정자는 리드 세션 또는 새 컨텍스트 판단 역할(`mission-governor` 아님). A1 기준선을 **A1 0.4.1 `6d98ca466`** 로 갱신하고, 0.4.1 이 확정한 항목은 재확인 표지 대신 「A1 0.4.1 (6d98ca466)」 인용으로 바꿨다(남은 표지는 A1 이 A3 로 남긴 항목과 A2·A4 항목뿐). A1 0.4.1 과의 상충 1건(push 직렬화·`sign` 차단의 소유 카드)은 선택하지 않고 `research.md §10.2` 에 보고. **추가(리드, A1 plan-audit iter-4 발견 V1)**: 저장소가 decide 영수증뿐 아니라 사람·`llm`·`jev`·`llm+jev` 서명, 재봉인, revoke 까지 **모든 서명 사건**을 기록하고(REQ-012), 저장소 기록이 없는 서명은 kickoff-check 가 거절(REQ-007) — A1 `supersedes` 는 위조 가능해 증거가 아니기 때문. A1 0.4.1 §C.6 인용 | manager-spec |
| 2026-09-26 | 0.2.0 | 리드 범위 추가: Kickoff 자율 승인, `moai contract revoke`, `moai contract decide`(가칭)와 moai 소유 영수증 저장소, t1235(A2) 선행 | manager-spec |
| 2026-09-26 | 0.1.0 | 최초 작성 (plan phase). 기준 트리 develop `ca1d5dc43`. A1 초안 `8f77d9a33` 기준 | manager-spec |

## §A. 배경

| 카드 | 담당 | 이 SPEC 과의 관계 |
|---|---|---|
| t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`, 0.5.2 `25283ebf8`) | 스키마 — `contract.yaml`(`card` 필드 포함), `moai contract sign/show/verify`, 영수증 필드와 구조·일관성 검증기, 결정자 값 집합·파생 기본값, 설정 `workflow.autonomy.*` | **선행.** A1 은 영수증을 파일로만 검증하고 출처를 `file` 로 기록하며, 저장소 대조로 **거절하는 일은 A3 에 맡겼다**(A1 §C.6). 결정 규칙은 A3 소유이고 A1 은 기록된 `outcome` 의 일관성만 본다(A1 §C.8). Jev 가 답한 `llm+jev` 영수증을 사람으로 보내는 A1 임시 규칙은 A3 가 원칙을 개정할 때 푼다(REQ-GR-025) |
| t1235 (A2, `SPEC-AUTONOMY-ESCALATION-001`) | 에스컬레이션 감지기와 기록 | **선행(리드 결정).** needs-decision 은 큐 상태가 아니라 A2 기록의 존재다 |
| t1245 (A2b) | push 직렬화 강제, 에이전트발 대화형 `contract sign` 차단, `MOAI_FACTORY_ROLE=agent` 세션의 `moai contract decide` 거절 | **선행(리드 결정 2026-09-26, A1 0.5.2 §C.1·§C.2 와 일치).** SPEC 이 아직 없을 수 있어 내부는 여기서 정하지 않는다 **[A2b SPEC 으로 재확인]** |
| t1237 (A4) | 종료 보고, 두 번째 리뷰 수행 기록, **리뷰 미수행 시 Push 전 정지** | 이 SPEC 의 Push 단계가 기대는 정지의 소유자(A1 §C.1). run 전제가 아니라 Push 단계 활성의 전제 |
| A5 | Jev 보정 | 이 SPEC 은 A5 를 기다리지 않는다. `jev_min_confidence` 값의 근거 측정만 A5 몫이다 |

운영자 결정(2026-09-26): A-Q1 계약 서명이 Implementation Kickoff Approval 을 대체 · A-Q2 develop push 는 계약 안에서 허용하되 직렬화 · A-Q3 두 번째 리뷰 필수, 미수행 시 Push 전 정지 · A-Q4 배포 기본값 `guided` · 「Kickoff 는 LLM·Jev 도 할 수 있게」(Kickoff 자율 승인) · 재결정(2026-09-26): Jev 단독 결정은 없고, contract 모드의 자율 결정자는 `llm`(기본)과 `llm+jev` 교차 확인이다. 맥락(이 SPEC 의 범위가 아님): F-Q2 자율 모드의 로컬 병합은 자동.

게이트 처분(조항 위치와 등록 여부는 `research.md §2`·`§3`):

| 게이트 | contract 모드 처분 |
|---|---|
| G1 Implementation Kickoff Approval | 계약 서명으로 전환. 사람 서명은 A-Q1 로 등가. `llm`·`llm+jev` 결정자 서명은 활성 조건(`design.md §7.1`)이 모두 참일 때만 |
| G2 소크라테스 인터뷰 | **서명 이후 작업에만** 계약 서명으로 흡수(plan phase 인터뷰는 그대로 — 기본값 D-2) |
| G3 접근 승인 | 계약의 `approach` 필드에서 1회 |
| G4 가정 확인 대기 | 기록하고 진행, 계약과 모순되면 에스컬레이션 |
| G7/G8 plan-audit FAIL | 상한까지 자동 수리 후 에스컬레이션 |
| G11 sync 확인 질문 | 제거 |
| G5·G14·G20·헌법 | 건드리지 않는다. 자율 승인과 revoke 는 결정을 기록할 뿐 묻지 않는다 |

## §B. 범위

1. **문서 층(추가형)** — 기존 텍스트를 바꾸지 않고 contract 모드 지시를 마커 블록으로 덧붙인다. 예외는 Jev 원칙 개정(REQ-013) 한 건으로, 그것은 원칙 문장 자체의 개정이다. 편집 대상은 `design.md §2`.
2. **Go 코드 층** — `moai contract kickoff-check`(가칭), `moai contract decide`(가칭), `moai contract revoke`, 영수증 저장소, 결정 규칙 R1~R5 와 원칙 개정 뒤 R3 해제. `design.md §7`~`§10`.
3. **검증 층** — 블록 가드, guided 보존·동등·변경 집합 검사, 활성 순서 검사를 Go 테스트로 둔다(`design.md §5`).

### 본문 결정 (NC 해소 기록 — `plan.md §B`)

- **D-1 모드 판독**: 오케스트레이터는 `moai contract kickoff-check --json`(A3)의 `mode`·`decider` 필드로 판독한다. kickoff-check 는 A1 설정 판독기를 거치므로 무효값 → `guided`·`human` 처리를 A1 코드 한 곳이 적용한다. A1 0.5.2 (25283ebf8) `show --json` 은 `mode` 만 내고 결정자 필드가 없어 kickoff-check 가 그 자리를 채운다.
- **D-2 plan phase 인터뷰(기본값, 운영자가 바꿀 수 있음)**: contract 모드는 서명 이후 작업에만 적용된다. 계약은 plan-audit 뒤에 서명되므로 plan phase 인터뷰는 계약 내용을 만드는 입력이다. `clarity-interview.md` 는 편집하지 않는다.
- **D-3 G4 원문**: `moai-constitution.md` 는 고치지 않는다(헌법 개정은 범위 밖). G4 처분은 SSOT 등가 조항으로만 표현한다.
- **D-4 서명 후 goal(기본값, 운영자가 바꿀 수 있음)**: 서명 검증 뒤 goal 을 자동으로 무장하지 않는다. goal 무장은 현행 절차대로 오케스트레이터가 필요할 때 한다.
- **D-5 초안 contract.yaml 지시**: 에이전트 파일을 고치지 않으므로 `spec-assembly.md` 의 위임문에 싣는다.
- **D-6 always-loaded 예산**: 세 always-loaded 블록 합계 1,500자 이하, 블록당 조건문 + 포인터만(REQ-002). Jev 원칙 개정이 고치는 `moai-mcp-tools-catalogue.md` 의 두 `jev_ask` 행은 always-loaded 가 아니다(`paths:` 가 맞을 때만 로드되는 상세 짝) — 그래서 상시 예산에 넣지 않고, 행당 증가 300자 이하의 조건부 로드 예산을 따로 둔다(`design.md §4`). t1175 이후 `moai-mcp-tools.md` 에는 `jev_ask` 행이 없어 편집 대상이 아니다. t1175 와의 조율은 리드 몫이며, 이 상한을 넘기지 않는다.
- **D-7 Jev 게이트 입력**: 리드 결정 L1 을 운영자 재결정으로 좁힌다 — contract 모드 Kickoff 의 `llm+jev` 교차 확인에서 두 번째 신호로 쓰일 때만 예외(REQ-013).
- **D-8 주 LLM 결정자**: 리드 결정 L3 — 리드 세션 또는 새 컨텍스트의 판단 역할. `mission-governor` 는 쓰지 않는다(그 정의가 승인을 배제한다).
- **D-9 revoke 정지 시점**: 다음 단계 경계에서 멈춘다. 프로세스를 죽이지 않는 안전한 방향이며, 경계 사이 작업이 끝까지 가는 지연을 잔여 위험으로 둔다.
- **D-10 결정자 값과 사람 경로**: contract 모드의 자율 결정자는 `llm`(결정자 키가 없을 때의 파생 기본값)과 `llm+jev` 둘이고 `jev` 는 거절된다(R5). 사람 서명 경로는 결정자 설정과 무관하게 언제나 열려 있다 — `reject`·`human` 결과의 도착점이다. 값 집합과 파생 기본값은 A1 스키마다 — A1 0.5.2 (25283ebf8) REQ-CONTRACT-015, design § Configuration
- **D-11 루트 해석**: 세 동사(`kickoff-check`·`decide`·`revoke`)는 SPEC 디렉터리와 `.moai/reports/<card>/` 를 호출한 작업 트리의 최상위(`git rev-parse --show-toplevel`)에서 찾는다 — 계약과 보고서는 카드 브랜치에만 있을 수 있기 때문이다. 저장소는 `homestate.ProjectKey` 로 모든 워크트리가 공유한다.
- **D-12 `card-mismatch` 의 exit**: kickoff-check 는 판정 명령이라 불일치를 판정 실패(exit 1)로, decide·revoke 는 쓰기 명령이라 쓰기 전의 입력 오류(exit 2)로 낸다.

## §C. 요구사항 (GEARS)

Tier L 상한 25개, 번호는 문서 순서대로 연속이다.

### 모드 분기와 guided 보존

- **REQ-GR-001** (Ubiquitous) — The 편집 대상 문서 shall contract 모드 지시를 `<!-- moai:contract-mode-start id="<slug>" -->` 와 `<!-- moai:contract-mode-end -->` 가 각각 한 줄을 차지하는 블록 안에만 두며, 그 블록은 어떤 `moai:evolvable-start` ~ `moai:evolvable-end` 구간 안에도 놓이지 않는다.
- **REQ-GR-002** (Ubiquitous) — The 편집 대상 문서 shall, contract-mode 블록(마커 줄 포함)을 제거하면 기준 ref `BASE` 의 같은 파일과 바이트 동일하다 — 이것은 **블록 밖 텍스트의 보존**이며, guided 세션도 블록을 읽는다. 그 증가를 줄이기 위해 always-loaded 파일의 블록은 적용 조건 한 문장과 SSOT 포인터만 담고 블록당 500자 이하·세 파일 합계 1,500자 이하이며, 스킬 파일의 블록은 블록당 900자 이하다.
- **REQ-GR-003** (Where) — **Where** `workflow.autonomy.mode` 가 `guided` 이거나, 키가 없거나, 무효한 값인 경우, the 오케스트레이터 지시문 shall 현행 guided 게이트를 그대로 적용하고, 모든 contract 블록은 첫 문장에서 적용 조건을 `workflow.autonomy.mode: contract` 로 선언한다. — A1 0.5.2 (25283ebf8) REQ-CONTRACT-015

### G1 — Kickoff 를 계약 서명으로

- **REQ-GR-004** (Where) — **Where** `workflow.autonomy.mode: contract` 인 경우, the plan→run 게이트 shall `moai contract kickoff-check <SPEC-ID> --card <card>` 의 exit 0 이다 — 오케스트레이터는 Implementation Kickoff Approval `AskUserQuestion` 을 내지 않는다. 설정 결정자가 `human` 이거나 자율 Kickoff 가 비활성이면(REQ-GR-008) 오케스트레이터는 서명 명령과 요약을 보고하고 턴을 닫으며, 운영자가 대화형 터미널에서 서명한다. **When** decide 의 `outcome` 이 `reject` 또는 `human` 이면 오케스트레이터는 영수증 서명 경로를 부르지 않고 설정 결정자 `human` 일 때와 같은 사람 서명 절차를 따르며 두 값을 구별해 다루지 않는다 — 둘 다 「서명 거절, 사람 결정 필요」다. kickoff-check 가 exit 0 이 아니면 run phase 에 들어가지 않고 그 사유를 담아 에스컬레이션하며, guided 로 조용히 떨어지지 않는다. — A1 0.5.2 (25283ebf8) REQ-CONTRACT-010·015
- **REQ-GR-005** (Ubiquitous) — The 신규 SSOT 규칙 shall 「Implementation Kickoff Approval 이 통과했다/얻었다」를 전제로 하는 MoAI 규칙·스킬·에이전트·출력 스타일의 참조가 contract 모드에서 **사람 서명**(`signer_kind: human`, `method: interactive-tty`)으로 충족된다는 등가 조항을 담는다. 비인간 서명의 등가는 이 조항에 들지 않고 REQ-GR-008 의 활성 조건으로만 성립한다. G1 발화 지점(`research.md §1.2` 의 E)은 각각 contract-mode 블록을 갖고, 참조 지점(R)은 이 조항으로 덮여 편집되지 않는다. — A1 0.5.2 (25283ebf8) REQ-CONTRACT-011, design § Signature Seal
- **REQ-GR-006** (Where) — **Where** contract 모드인 경우, the plan 워크플로 shall manager-spec 위임 지시에 서명 전 초안 `contract.yaml`(`signature` 블록 없음)을 SPEC 산출물과 함께 내라는 지시를 포함한다. — A1 0.5.2 (25283ebf8) design § Contract Schema

### Kickoff 자율 승인 — 판정의 소유자는 A3

- **REQ-GR-007** (When) — **When** `moai contract kickoff-check <SPEC-ID> --card <card>` 가 실행될 때, the 명령 shall 다음이 모두 참일 때만 exit 0 을 낸다: 계약의 `card` 필드가 `<card>` 와 같고, A1 verify 가 `signed-valid` 이고, 현재 서명(봉인 `seal` 기준)과 일치하는 서명 사건이 `events.jsonl` 에 있으며, 다음 둘 중 하나 — (a) `signer_kind: human` 이고 `method: interactive-tty`, (b) `signer_kind` 가 `llm` 또는 `llm+jev` 이고 `method: receipt` 이며, 서명의 `receipt.sha256` 이 `receipts.jsonl` 의 같은 카드·SPEC 영수증 해시와 같고, 그 영수증의 `outcome` 이 `approve` 이고, `effective_decider`(요청 `llm+jev` 에서 대체된 `llm` 포함)가 `signer_kind` 와 같고 `requested_decider` 가 현재 설정 결정자와 같으며, 그 발급 뒤에 같은 카드의 revoke 사건이 없으며 A3 revoke 판독기(REQ-GR-022)가 현재 서명에 대해 차단을 보고하지 않고, REQ-GR-008 의 활성 조건이 모두 참. 실패 사유는 모두 exit 1 이며 `card-mismatch`, `not-signed-valid`(A1 verify 상태 `unsigned`·`signed-invalid` 를 함께 보고), `signature-not-recorded`, `decider-not-permitted`(`signer_kind: jev`, R5), `receipt-not-issued`(저장소에 없는 영수증 — 에이전트가 쓴 파일 포함), `receipt-not-approved`(대응 영수증의 `outcome` 이 `reject`·`human` — 그 영수증과 일치하게 위조한 서명 포함), `decider-mismatch`(`effective_decider` ≠ `signer_kind` 또는 `requested_decider` ≠ 설정 결정자), `revoked`, `autonomous-kickoff-inactive` 다. 실패 사유가 둘 이상 성립하면 `--json` 은 성립하는 사유를 모두(정렬·중복 제거) 담고, 함께 내는 대표 사유 하나는 다음 우선순위를 따른다: `card-mismatch` > `decider-not-permitted` > `not-signed-valid` > `signature-not-recorded` > `receipt-not-issued` > `receipt-not-approved` > `decider-mismatch` > `revoked` > `autonomous-kickoff-inactive`. 사용법·I/O·저장소 무결성 오류는 exit 2 다. `--json` 은 판정과 사유, `mode`·설정 결정자·`autonomous_kickoff_enabled`·`jev_doctrine_amended`·A1 verify 상태를 낸다. 이 명령은 아무것도 쓰지 않는다(`card-mismatch` 의 exit 는 본문 결정 D-12). — A1 0.5.2 (25283ebf8) REQ-CONTRACT-008·011·024, design § Contract Schema·§ Card Field (A1 은 출처를 `file` 로만 기록하고 저장소 대조 거절을 A3 에 맡김, §C.6)
- **REQ-GR-008** (Where) — **Where** 설정 결정자가 `llm` 또는 `llm+jev` 인 경우, the 자율 Kickoff shall `design.md §7.1` 의 활성 조건 표(이 SPEC 의 유일한 정본)의 활성 조건이 모두 참일 때만 활성이며, 그 조건은 revoke·decide·저장소의 존재, A1 의 영수증 서명 경로, A2b 의 push 직렬화 강제와 에이전트발 대화형 `sign`·`decide` 차단, 그리고 `orchestration-mode-selection.md` 의 Frozen 태그 문단에 비인간 서명 경로를 명시하는 **별도 개정**의 착지다. 비활성 동안 kickoff-check 는 `llm`·`llm+jev` 서명을 사유 `autonomous-kickoff-inactive` 로 거절한다. REQ-GR-025 의 연동 커밋은 전체 활성의 조건이 아니라 Jev 가 답한 교차 확인이 사람 대신 결정할 수 있게 되는 조건이다. — A1 0.5.2 (25283ebf8) REQ-CONTRACT-024, §C.1·§C.6 **[A2b SPEC 으로 재확인]** (두 가드와 `decide` 거절, t1245)
- **REQ-GR-009** (When) — **When** `moai contract decide <card> --spec <SPEC-ID> --judgement <file|->` 가 실행될 때, the 명령 shall 판단을 열기 전에 여섯 전제조건을 기계적으로 평가하고, 하나라도 성립하지 않으면 Jev 를 호출하지 않고 `outcome: human` 을 기록한다: (a) plan-audit 판정 — decide 가 카드 증거 경로 `.moai/reports/<card>/` 에서 `plan-audit-<N>.md`·`plan-audit-iter<N>.md` 중 가장 높은 N 인 파일을 스스로 골라 그 경로와 해시를 영수증 `inputs.plan_audit_report` 에 기록한다 — 의 `Verdict:` 줄 값이 정확히 `PASS` 또는 `PASS-WITH-DEBT` 이고, `Overall Score:` 줄의 값(부동소수로 파싱 가능해야 함)이 Tier 문턱(S 0.75 / M 0.80 / L 0.85) 이상이며(`PASS-WITH-DEBT` 도 문턱 이상일 때만 통과), 그 파일의 `plan_artifact_hash:` 줄 값이 현재 SPEC 디렉터리의 plan 산출물 해시(`internal/runtime` 의 plan-artifact 해시 — spec·plan·acceptance·design·research, 있으면 tasks)와 같다 — 줄이 없거나 다르면 판정이 현재 산출물에 묶이지 않은 것이므로 실패, (b) SPEC 디렉터리의 `plan.md`·`research.md` 에 미해결 질문 표지(대괄호로 시작하는 `NEEDS CLARIFICATION` 표지) 0건, (c) A1 verify 가 서명 부재 외 사유 없음, (d) 계약 `actions` 에 A1 금지 행동 토큰 없음, (e) 해당 카드에 A2 판독(`escalation.NeedsDecision`)으로 열린 기록이 없고 A3 revoke 판독기(REQ-GR-022)가 현재 서명에 대해 차단을 보고하지 않음, (f) `ownership.write` 와 A1 `frozen-files` 불변식 집합의 교집합이 공집합. — A1 0.5.2 (25283ebf8) design § Kickoff Receipt · § Verify Reason Codes · § Action Vocabulary · § Frozen Files; 열린 기록 판독은 A2 0.4.3 (`internal/escalation` `NeedsDecision`, BASE `7fe658815`)
- **REQ-GR-010** (Ubiquitous) — The `moai contract decide` shall 영수증의 `outcome ∈ {approve, reject, human}` 을 다음 규칙으로 도출하며, 이 규칙은 이 SPEC 이 소유한다(운영자 재결정 2026-09-26, 리드 소유 조정 — A1 은 기록된 `outcome` 이 답과 모순되지 않는지만 본다): **R1** 설정 결정자 `llm+jev` 에서 Jev 가 답했으면 두 답이 모두 approve → `approve`, 모두 reject → `reject`, 불일치나 escalate → `human`; **R2** `llm+jev` 에서 Jev 쪽이 실패하면 — `jev_call_failed`, `jev_key_missing`, `jev_disabled`, `jev_malformed_response`, `jev_low_confidence`(신뢰도 < `jev_min_confidence`) — `llm` 단독으로 대체해 R4 로 도출하고 `requested_decider: llm+jev`·`effective_decider: llm`·`fallback{applied: true, reason}` 을 영수증과 저장소에 기록; **R3** REQ-GR-025 의 연동 커밋 전에는 Jev 가 답한 `llm+jev` 결정이 `human`(`jev-doctrine-not-amended`); **R4** `llm` 단독(설정 또는 대체)에서 approve → `approve`, reject → `reject`, escalate → `human`; **R5** 설정 결정자 `jev` 는 거절 — exit 2(`decider-jev-refused`), 쓰기 없음, Jev 미호출. `reject` 와 `human` 은 모두 자율 시작을 막고 Kickoff 를 사람에게 보낸다 — 운영자 결정 「`llm` 거절 → 사람」은 이 효과로 충족되며 둘의 차이는 기록뿐이다. 또한 **SPEC 작성 에이전트는 결정자가 될 수 없다**: decide 는 SPEC 디렉터리를 건드린 plan-phase 커밋들의 본문을 `internal/spec` 의 `Authored-By-Agent:` 판독기(`lint_ownership.go` `parseAuthoredByAgent`)로 읽어 작성 에이전트 집합을 만들고 — git 의 트레일러 파서는 이 저장소 커밋 끝의 서명 줄 때문에 그 값을 보지 못한다 — 판단 파일의 결정자 신고값(`agent`)이 `manager-spec` 이거나 그 집합에 들면 `outcome: human`(`author-decider-conflict`), 그 커밋들 중 트레일러를 가진 커밋이 하나도 없으면 `outcome: human`(`author-check-unmeasured`)이다. 결정자 신원은 판단 파일의 자기 신고라 moai 가 검증하지 못한다 — 저장소는 신고값과 판독한 트레일러 집합을 함께 기록하며, 이것이 남는 위험이다. — 필드와 대체 사유 이름은 A1 0.5.2 (25283ebf8) design § Kickoff Receipt(필드 규칙 2~9); 규칙은 같은 곳 § Cross-check Rules(A3 소유, 참고)와 일치
- **REQ-GR-011** (When) — **When** `moai contract decide <card> --spec <SPEC-ID> --judgement <file|-> [--json]`(동사 이름은 가칭 — A1 동사 `show`·`verify`·`sign` 에 없고 A3 가 만든다)이 실행될 때, the 명령 shall `<card>` 가 계약의 `card` 필드와 같을 때만 진행하고(다르면 exit 2 `card-mismatch`), 결정하는 경로(R1~R4 로 `outcome` 을 도출)에서는 moai 가 발급한 영수증을 `$MOAI_HOME/db/<project-key>/contract/receipts.jsonl` 에 한 줄로, 그 발급 사건을 같은 디렉터리의 `events.jsonl` 에 `decide` 한 줄로 추가하고 같은 영수증 본문을 A1 서명 경로의 입력인 `.moai/specs/<SPEC-ID>/kickoff-receipt.json` 에 쓰며, 결정하지 않는 경로(설정 결정자 `human`, 전제조건 실패, 작성자 배제)에서는 `events.jsonl` 의 `decide` 사건 한 줄만 쓰고 영수증 줄과 파일은 쓰지 않는다 — A1 영수증은 `requested_decider ∈ {llm, llm+jev}` 와 `llm_answer` 를 요구하고 모르는 필드를 거절하므로(필드 규칙 1·2·6) 그런 경로를 표현할 수 없다. 사유 코드(`precondition:<x>`·`author-decider-conflict`·`author-check-unmeasured`·`cross-check-disagree`·`jev-doctrine-not-amended`·`decider-human`)는 영수증 본문에 넣지 않고 사건 줄과 `--json` 에만 둔다. 결과와 무관하게 기록에 성공하면 exit 0, 저장소 무결성 실패로 아무것도 쓰지 않으면 exit 1, 사용법·입력 형식·I/O 오류로 아무것도 쓰지 않으면 exit 2 이고, 결과는 `--json` 의 `outcome` 으로 전달한다. 계약에 서명하지 않고, `contract.yaml`·SPEC 문서·큐(`backlog.db`)·git 상태를 바꾸지 않으며, LLM 을 호출하지 않는다 — 판단은 호출한 세션의 LLM 이 `--judgement` 로 넘기고, 정상 경로는 리드 세션의 도구 호출이다. `MOAI_FACTORY_ROLE=agent` 세션에서의 거절은 A2b 의 가드다 **[A2b SPEC 으로 재확인]**. — A1 0.5.2 (25283ebf8) design § Kickoff Receipt(고정 경로, 필드 규칙 1·2·6)·§ Card Field, spec §C.6(b)
- **REQ-GR-012** (Ubiquitous) — The moai 소유 저장소 shall 워크트리 밖 디렉터리 `$MOAI_HOME/db/<project-key>/contract/`(리드 결정 R10, 2026-09-26 — `<project-key>` 는 `internal/homestate` 의 프로젝트 키)의 두 추가 전용 파일 — 발급 영수증 `receipts.jsonl` 과 서명 사건 `events.jsonl` — 로 이루어지며, **모든 서명 사건** — 사람 서명, `llm`·`llm+jev` 영수증 서명(대체 영수증 포함), `--resign` 재봉인, decide 영수증 발급, revoke — 이 `events.jsonl` 에 각각 정확히 한 줄을 남기고, 두 파일의 각 줄은 직전 줄의 해시를 담아 체인을 이룬다. 서명·재봉인 사건은 SPEC·카드·`signer_kind`·`method`·봉인 `seal`·`contract_sha256`·`acceptance_sha256`·`supersedes` 를, `decide` 사건은 대응 영수증 줄의 해시·적용한 규칙(R1~R5)·결정자 신고 신원·SPEC 작성자 트레일러 집합·전제조건 평가를 담고, 영수증 줄은 A1 영수증 본문(`requested_decider`·`effective_decider`·`fallback`·`llm_answer`·`jev_answer`·`outcome`·입력 해시)과 Jev 를 호출했으면 그 원시 요청·응답을 담는다. A1 의 `supersedes` 는 재봉인 때 위조자가 다시 쓸 수 있으므로 증거가 아니며, 변조 흔적의 근거는 이 사건 기록뿐이다 — 이를 위해 A3 는 A1 서명기(사람·영수증 경로와 `--resign`)가 서명 직후 `events.jsonl` 에 사건을 추가하도록 개정하고, 추가에 실패하면 서명 파일도 쓰지 않는다. 목표는 변조 **흔적**이며 방지가 아니다. 체인은 마지막 줄 삭제를 스스로 검출하지 못한다: 서명 사건이 삭제되면 kickoff-check 가 `signature-not-recorded` 로 거절하고, revoke 사건이 삭제되면 revoke 가 함께 쓴 `kind: revoke` 에스컬레이션 기록이 두 번째 증인으로 남아 A3 revoke 판독기(REQ-GR-022)가 그것을 차단 신호로 읽으므로 전제조건 (e)·kickoff-check·단계 경계가 재개를 막는다. A2 는 revoke 기록을 `status: resolved` 로 두고 needs-decision 으로 세지 않는다(A2 0.4.3 spec §I.2) — 이 두 번째 증인의 차단 효과는 A3 판독기의 책임이다. — A1 0.5.2 (25283ebf8) §C.6·§H 가 이 요구를 A3 의 전방 요구로 넘겼다(「that store shall record every signing event, including human signatures」, 위치 `$MOAI_HOME/db/<project-key>/contract/`). 서명기 안의 추가 지점은 A1 이 정하지 않았다 — A3 소관
- **REQ-GR-013** (Ubiquitous) — The Jev 표시 전용 원칙 shall **contract 모드 Kickoff 의 `llm+jev` 교차 확인에서 두 번째 신호로 쓰일 때** 한 곳에서만 예외를 갖도록 개정되며, 개정 위치는 `SPEC-JEV-CORE-001` 의 세 곳 — REQ-JEVC-011(표시 전용 문장: 그 답을 moai 가 Kickoff 영수증과 저장소에 기록하는 일을 같은 한 곳의 예외로), REQ-JEVC-012(개정 표지), `### Out of Scope — authority` 의 두 항목(같은 예외를 명시해 그 SPEC 이 스스로 모순되지 않게) — 과 그 HISTORY 행(운영자 결정 「Kickoff 는 LLM·Jev 도 할 수 있게」와 2026-09-26 재결정 — Jev 단독 없음, 교차 확인만 — 인용), `moai-mcp-tools-catalogue.md` 의 두 `jev_ask` 행(도구 표의 `mcp__moai__jev_ask` 행과 계열 요약 표의 `Judgment (gated)` 행 — BASE `7fe658815` 기준 138·216행), `workflow.yaml` 의 `jev:` 주석(각각 로컬·템플릿 사본), 그리고 `CLAUDE.local.md §29` 3등급 문장의 한 줄 예외(문안은 `design.md §11.1`)다. `moai-mcp-tools.md` 는 t1175 이후 `jev_ask` 행을 담지 않아 개정 위치가 아니다. 다른 모든 게이트와 Jev 단독 결정에 대한 금지는 그대로다. 개정과 규칙 R3·A1 임시 규칙의 해제는 REQ-GR-025 에 따라 한 커밋으로 착지한다. — A1 0.5.2 (25283ebf8) §C.8 (세 원칙 위치 목록)

### G2 / G3 / G4 — 인터뷰·접근 승인·가정 대기

- **REQ-GR-014** (While) — **While** 서명된 계약 범위 안의 작업이 진행 중인 동안, the 오케스트레이터 shall 소크라테스 인터뷰와 접근 승인을 계약 서명으로 충족된 것으로 취급하고(접근은 `approach` 필드가 1회 승인), 가정은 `progress.md` 에 적고 기다리지 않고 진행하되, 가정이나 새 모호함이 계약의 `acceptance`·`invariants`·`ownership`·`approach` 와 모순되면 멈추고 `escalate_on` 의 해당 종류로 에스컬레이션한다. plan phase 인터뷰는 이 요구의 대상이 아니다(본문 결정 D-2). — A1 0.5.2 (25283ebf8) design § Contract Schema

### G7 / G8 — plan-audit FAIL

- **REQ-GR-015** (When) — **When** contract 모드에서 plan-auditor 가 FAIL 을 내거나 SPEC 품질 게이트(Phase 15)가 WARNING/FAIL 을 낼 때, the 오케스트레이터 shall 수리 위임과 재감사를 상한까지 자동으로 반복하고, 상한에 이르면 선택 질문 대신 에스컬레이션 보고를 낸다. 상한은 초안 `contract.yaml` 의 `budget.audit_retries`, 없으면 `workflow.autonomy.escalation.budget_default.audit_retries`(기본 2)다. — A1 0.5.2 (25283ebf8) design § Contract Schema (`budget.audit_retries` 는 plan-audit·sync-audit 공용 상한), § Configuration

### G11 — sync 확인 질문

- **REQ-GR-016** (Where) — **Where** contract 모드인 경우, the sync 워크플로 shall 문서 범위 승인(`gate-sync-2`), Context-Aware Next Steps 질문, 「현재 브랜치에서 동기화할까」 질문을 내지 않고, sync 워크플로 안의 실패 결정점 — Phase 1 게이트 실패, Phase 3·7 테스트 실패, Phase 6 호환성 파괴, Phase 8 보안 critical, Phase 13 로컬 CI 미러 실패(`sync/delivery.md` Step 3.1.5) — 을 질문 대신 에스컬레이션 보고로 라우팅한다. PR 뒤의 CI autofix 루프(Frozen)는 이 요구의 대상이 아니다.

### 유지되는 게이트와 에스컬레이션

- **REQ-GR-017** (Ubiquitous) — The contract-mode 블록과 이 SPEC 의 CLI 동사 shall Frozen 게이트(질문 채널 독점, CI autofix 3회 후 질문, 컨텍스트 한도 `/clear`)와 헌법 조항을 완화·재해석·재서술하지 않으며 — 자율 승인과 revoke 는 결정을 기록할 뿐 사용자에게 묻지 않는다 — SSOT 는 유지 게이트를 명시한다: sync-auditor must-pass, `main`/release 동작은 계약으로 허용될 수 없음, 카드 선택은 운영자, goal 상한, 파괴적 명령 확인. — A1 0.5.2 (25283ebf8) design § Action Vocabulary
- **REQ-GR-018** (Ubiquitous) — The 에스컬레이션 보고 shall Report-Before-Ask 보고 형식을 재사용하며, 보고를 남기고 멈추는 것이지 `AskUserQuestion` 이 아니다. needs-decision 은 큐 상태가 아니라 A2 형식의 열린 에스컬레이션 기록(`.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`, 기계 판독 YAML 머리)의 존재이며, 기록 형식은 A2 소관이다. revoke 기록은 A2 에서 해결된 기록이라 needs-decision 이 아니며, 그 재개 차단은 A3 revoke 판독기(REQ-GR-022)가 맡는다. — A2 0.4.3 spec §I·§I.1·§I.2 (BASE `7fe658815`, M0 대조)

### 한 호흡 생명주기

- **REQ-GR-019** (Ubiquitous) — The run 워크플로 문서 shall contract 모드 생명주기의 Discovery → RED → GREEN → Qualification 을, the sync 워크플로 문서 shall Closure → Integration → Push 를, the SSOT 의 생명주기 절 shall 일곱 단계 전부를 이 순서로 담는다.
- **REQ-GR-020** (Ubiquitous) — The 각 단계 shall 자신의 증거가 기록되고, 해당 카드에 A2 판독(`escalation.NeedsDecision`)으로 열린 기록이 없으며, A3 revoke 판독기(REQ-GR-022)가 현재 서명에 대해 차단을 보고하지 않을 때만 다음 단계로 넘어가며 — revoke 기록은 A2 에서 `status: resolved` 라 A2 판독에 잡히지 않으므로, 오케스트레이터는 단계 경계마다 `moai contract kickoff-check` 를 불러 사유 `revoked` 로 이를 읽는다 — SSOT 는 단계별 증거를 명시한다 — Discovery 는 계약 `reobserve` 목록 재관측과 재현 명령·출력, RED 는 구현 전에 커밋된 실패 테스트의 커밋 SHA 와 원문 실패 출력, GREEN 은 같은 테스트의 통과 출력, Qualification 은 범위 테스트·lint·커버리지·변이 검사·두 번째 모델 리뷰(`audit_multi`, 모델은 `review.second_model`)의 명령과 출력, Closure 는 상태 전이와 `.moai/reports/<card>/verdict.md`, Integration 은 develop 흡수 후 병합 트리 재측정 출력. — A1 0.5.2 (25283ebf8) design § Contract Schema; 열린 기록 판독은 A2 0.4.3 (`internal/escalation` `NeedsDecision`, BASE `7fe658815`)
- **REQ-GR-021** (While) — **While** Push 단계에 있는 동안, the develop push shall 계약 `actions` 에 `push-develop` 이 있고 `workflow.autonomy.contract.push_develop` 가 참이며 A1 이 정한 `push-develop` 리스(A2b 가 강제)를 잡은 상태에서만 수행되고, 두 번째 리뷰 미수행 시의 Push 전 정지는 A4(t1237)가 소유하므로 A4 가 착지하기 전에는 contract 모드 Push 단계가 비활성이며 SSOT 는 그 자리를 A4 로 가리킨다. — A1 0.5.2 (25283ebf8) REQ-CONTRACT-018 (`push_requires_lease`, 강제는 A2b) **[A4 감사 통과본으로 재확인]** (정지 형태) **[A2b SPEC 으로 재확인]** (리스 강제)

### moai contract revoke

- **REQ-GR-022** (When) — **When** `moai contract revoke <card> --spec <SPEC-ID>` 가 실행될 때, the 명령 shall `<card>` 가 계약의 `card` 필드와 같을 때(다르면 exit 2) 그 SPEC 의 계약이 서명되어 있으면(서명 종류·출처와 무관) 저장소에 revoke 기록을 추가하고 A2 형식의 에스컬레이션 기록을 정확히 1건 쓰며 exit 0 을 낸다 — 그 기록은 A2 의 쓰기 함수 모양 그대로 경로 `escalation.RecordPath(<워크트리 최상위>, <card>, revoke-operator, <지문>, 0)`(= `.moai/reports/<card>/escalation/revoke-operator-<지문>.md`)에 두고, YAML 머리는 `kind: revoke`·`class: revoke-operator`·`status: resolved`·비어 있지 않은 `decider`(A2 0.4.3 spec §I.2 가 revoke 기록에 요구하는 값)·`card`·`spec` 이며, 지문은 `escalation.Fingerprint("revoke-operator", <폐기한 서명의 seal>)` 다. 이 기록을 재개 차단 신호로 읽는 것은 **A3 revoke 판독기**다: `<워크트리 최상위>/.moai/reports/<card>/escalation/*.md` 를 A2 의 `escalation.ParseRecord` 로 읽어, `kind` 가 `revoke` 이고 `card`·`spec` 이 대상과 같으며 `fingerprint` 가 `escalation.Fingerprint(<그 기록의 class>, <현재 서명의 seal>)` 와 같은 기록이 하나라도 있으면 차단으로 판정하고, `status` 값은 보지 않으며, 디렉터리가 없으면 기록 0건(해제, 오류 아님 — A2 `NeedsDecision` 의 `filepath.Glob` 과 같은 의미)이고, 있는 디렉터리의 목록이나 있는 기록 파일을 읽거나 파싱하지 못하면 차단으로 판정한다(오류는 차단 쪽). 새 서명은 seal 이 달라 차단을 풀고, 사람 서명 절차로 되돌아가는 것이 revoke 의 효과다. **경계: A2 는 revoke 를 해결된 기록으로 적는다 — A3 의 재개 차단은 A3 판독기의 책임이며 A2(`internal/escalation` 의 `NeedsDecision`, A2 spec §I.2)는 바꾸지 않는다.** 이미 revoke 되어 그 뒤 새 서명이 없으면(판독기가 현재 seal 에 대해 차단을 보고하면) 아무것도 쓰지 않고 exit 0(멱등); 계약이 서명되지 않았거나 없으면 exit 1; 사용법·I/O·저장소 무결성 오류는 exit 2. 진행 중인 run 은 다음 단계 경계에서 멈춘다(본문 결정 D-9). 또한 the 명령 shall not 워크트리를 지우거나, 브랜치를 지우거나 이름을 바꾸거나, push 하거나, 큐를 바꾸거나, `contract.yaml`·서명·SPEC 문서를 고치거나, 프로세스를 종료하거나, git 쓰기 명령을 실행한다 — 이 금지는 멱등 경로와 오류 경로에도 적용된다. — 기록 경로·YAML 머리·`revoke` 종류는 A2 0.4.3 spec §I·§I.1·§I.2 와 `internal/escalation/record.go`(`RecordPath`·`Fingerprint`·`ParseRecord`·`NeedsDecision`, BASE `7fe658815`, M0 대조); A1 0.5.2 (25283ebf8) REQ-CONTRACT-008·014, design § Card Field

### 로컬·템플릿 동등, 중립성, 범위

- **REQ-GR-023** (Ubiquitous) — The 각 contract-mode 블록 shall 로컬 사본과 템플릿 사본에서 바이트 동일하고(신규 SSOT 는 통째로 동일), 블록 밖의 기존 로컬·템플릿 분기는 `BASE` 대비 변하지 않으며, 블록과 SSOT 는 SPEC ID·REQ/AC 토큰·카드 id·내부 날짜·커밋 SHA 를 담지 않고, 템플릿 테스트 패키지는 짝 맞춤·비중첩·evolvable 구간 밖 배치·금지 클래스 부재·블록 문자 상한을 상시 검사하며 검사 대상 블록이 0개면 실패한다.
- **REQ-GR-024** (Unwanted) — The 이 SPEC 의 구현 shall not `design.md §2` 의 편집 허용 목록(템플릿 사본과 이 SPEC 디렉터리 포함) 밖의 파일을 바꾸며, 특히 `moai-constitution.md`·`zone-registry.md`·`.claude/agents/**`·`.claude/output-styles/**`·`ci-autofix-protocol.md`·`context-window-management.md`·`agent-common-protocol.md`·`spec-workflow.md`·`session-handoff-examples.md`·큐 스키마를 바꾸지 않는다.


### 활성 순서 — Jev 원칙 개정과 해제의 연동

- **REQ-GR-025** (Ubiquitous) — The Jev 원칙 개정(REQ-GR-013 의 모든 위치 — `SPEC-JEV-CORE-001` 세 곳과 HISTORY, `moai-mcp-tools-catalogue.md` 의 두 `jev_ask` 행, `workflow.yaml` `jev:` 주석(각각 로컬·템플릿 사본), `CLAUDE.local.md §29` 한 줄), 규칙 R3 의 해제, A1 임시 규칙의 해제 shall **같은 커밋에서 함께** 착지하며, 셋 중 일부만 담은 커밋이나 트리는 없어야 한다 — 어느 하나만으로는 Jev 가 답한 교차 확인이 자율 Kickoff 를 열지 못하게 하고, 원칙이 금지하는 동안 코드가 Jev 답을 결정에 쓰지 않게 하기 위해서다. R3 와 A1 서명기 단계 (1) — `internal/contract/receipt.go` 의 `ReceiptOutcome` 첫 분기, 그 자리에서 조건화하며 `sign/` 에 규칙을 다시 두지 않는다 — 은 한 상수 `jevDoctrineAmended` 로 함께 조건화되며(거짓이면 둘 다 적용), 그 상수를 참으로 바꾸는 것이 두 해제다. `CLAUDE.local.md §29` 는 이 묶음 안에 있다 — 운영자가 그 한 줄(`design.md §11.1`)을 문안 그대로 승인했고(2026-09-26, 리드 전달), 실제 편집은 카드 워크트리의 develop 사본 `CLAUDE.local.md` 에서만 하며 develop 병합으로 착지한다. 커밋의 조립은 리드 결정 2026-09-26(D32)을 따른다 — SPEC 본문은 manager-spec 이, 코드와 규칙·설정은 manager-develop 이 차례로(동시에 아님) 작성하고, 레인 오케스트레이터가 두 결과를 명시 경로로 스테이징해 한 커밋을 만든다. 이 커밋은 자율 Kickoff 전체 활성(REQ-GR-008)의 조건이 아니다. — A1 0.5.2 (25283ebf8) spec §C.8 「A3 lifts this A1 rule, and replaces its test」, REQ-CONTRACT-024, design § Kickoff Receipt 서명기 단계 (1), acceptance AC-CONTRACT-016 (t) **[A1 개정본으로 재확인]** (상수 조건화가 A1 의 「A3 removes this refusal」 문언의 구현으로 받아들여지는지 — `25283ebf8` 기준)

## §D. 범위 밖

### Out of Scope — Frozen 게이트와 헌법

- 질문 채널 독점(G5), CI autofix 3회 후 질문(G14), 컨텍스트 한도 `/clear`(G20), 헌법 조항과 그 개정 절차, `moai-constitution.md` 본문, zone-registry 등록.
- `orchestration-mode-selection.md` Frozen 태그 문단의 **비인간 서명** 확장 개정 자체. 이 SPEC 은 그 개정을 활성 조건으로 둘 뿐이며, 개정은 운영자의 명시적 개정 결정으로 별도 수행한다(`design.md §7.1` 6행).

### Out of Scope — A1·A2·A2b·A4·A5 소관

- `contract.yaml` 스키마, `sign/show/verify`, 영수증 형식과 검증기, 결정자 값 집합과 파생 기본값, `card` 필드, 설정 키 구현 — t1234(A1).
- 에스컬레이션 감지와 기록 형식 — t1235(A2).
- push 직렬화 강제, 에이전트발 대화형 `sign` 차단, 팩토리 에이전트 세션의 `decide` 거절 — t1245(A2b, 리드 결정 2026-09-26).
- 두 번째 리뷰 수행 기록과 미수행 시 Push 전 정지, 종료 보고 — t1237(A4).
- Jev 보정(`jev_min_confidence` 값의 근거 측정) — A5.

### Out of Scope — plan-audit 보고서 명명 통일

- `internal/runtime` 이 읽는 `<SPEC-ID>-review-<N>.md` 와 카드 증거 경로의 `plan-audit-<N>.md` 명명 불일치 해소 — 후속 카드 후보(`research.md §9.6`). 이 SPEC 은 보수적 선택 규칙만 둔다(REQ-GR-009 (a)).

### Out of Scope — 큐 상태 추가

- `needs-decision` 을 `backlog.db` 의 네 번째 상태로 추가하는 일. 리드 결정으로 needs-decision 은 A2 기록이다.

### Out of Scope — 변조 방지

- 단일 사용자 로컬 환경의 영수증 위조를 막는 암호학적 보장. 흔적만 제공한다.

### Out of Scope — 동음이의 게이트

- `e2e.md` 의 `--autofix` 「Kickoff Approval (one-time gate)」와 `harness-build-entry.md` 의 Builder 승인 게이트.

### Out of Scope — 에이전트·출력 스타일 편집

- `.claude/agents/**` 와 출력 스타일. Kickoff 언급은 참조 지점이며 REQ-GR-005 가 덮는다. `mission-governor` 는 결정자로 쓰지 않는다(D-8).

### Out of Scope — 로컬 전용 하네스와 한도 초과 파일

- `.claude/agents/harness/workflow-specialist.md`(템플릿 미러 없음), `spec-workflow.md`(40,799자)·`session-handoff-examples.md`(41,616자) — 참조 지점, 편집하지 않음.
