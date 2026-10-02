# SPEC-CODEX-REVIEW-OWNERSHIP-001 — progress

## Decision Log

원문 출처: `.moai/reports/t1422/jev-decisions.md`(Jev 판정 기록, 모델 `jev-1.13.0`, 트리 HEAD `c50da9c2f`). Jev 응답은 **표시 전용 판단 신호**이며 결정 권한은 레인, 0.5 미만 신뢰도는 리더 소관이다. 아래 표는 원문 그대로다.

| # | Question | Jev choice | Confidence | Probabilities | Disposition |
|---|----------|-----------|------------|---------------|-------------|
| D1 | How the leader stops being reviewed by the turn-end gate | `config_axis` | 1.00 | config_axis 1.00 · hard_skip 0.00 · role_detect 0.00 | Adopted. A new `review_gate` key decides what a non-card (tree-scope) session does; default keeps today's whole-tree review so product users are unchanged; this repo's primary config sets skip. No kanban env is read (t1399 deletes it; REQ-CGS-004 forbids env as a discriminator). |
| D2 | On-demand self-review surface | `new_review_tools` | 0.44 (**below 0.5 — reported to leader**) | new_review_tools 0.63 · extend_audit_tools 0.37 · cli_only 0.00 | Provisional: dedicated review tools. Plan-audit must test this against `extend_audit_tools`; reversible at plan stage. |
| D3 | Relation of t1404 to t1422 | `absorb_scoping_only` | 0.82 | absorb_scoping_only 0.88 · keep_separate 0.11 · absorb_all 0.01 | Adopted. Stale-main and non-card scoping cases are absorbed; the sync-gate `WCI_EXCLUDES` gap and doc-drift card t1406 stay separate. Absorption must be re-verified against code in the SPEC (premise check, `verification-claim-integrity.md` §1.1 surface 4). |

### Q1-Q8 해결 행 (2026-10-02, 운영자 위임 Jev `jev-1.13.0` + 리더 승인)

위 D1-D3 표는 `.moai/reports/t1422/jev-decisions.md` 와 바이트 동일하게 둔다. plan 단계에서 표면화된 오픈 결정 Q1-Q8 의 해결은 별도 표로 이어 적는다. 답변 원문은 레인 스크래치패드에 기록돼 있다(경로는 휘발성이라 인용하지 않음). **Q6·Q7 은 신뢰도 0.5 미만 — 잠정, 리더에 통지됨.**

| # | Question | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| Q1 | `tree_scope` 배포 형태 | `comment_example` | 0.74 | 템플릿에 주석 예시로만. 구조체는 파싱. 인벤토리·스키마·콘솔·i18n 불변(AC-CRO-001b) |
| Q2 | 키 이름·값 | `workflow.codex.review_gate.tree_scope` = review\|skip | 1.00 | 최종 |
| Q3 | D2 자기 리뷰 표면 | 전용 도구 | (리더 조건부 승인) | 조건(비교 표·결함 (b) 수리·advisory) 충족: plan.md §B.1, REQ-CRO-008/010. Jev D2 intake 0.44 |
| Q4 | t1404 처분 | `edit_to_residual` | 0.92 | 착지 뒤 리더가 잔여(T2·T4·T5·T8)로 편집, 착지 전 닫지 않음. sync 인계 항목(코드 변경 아님) |
| Q5 | 추적 `workflow.yaml` 의 `enabled: true` | `do_not_commit` | 0.97 | 커밋하지 않음. 리더가 primary 비추적 로컬 설정에 `tree_scope: skip` (운영자 소유, SPEC 은 편집 안 함). 레인 수단은 카드 리뷰 단계뿐 |
| Q6 | 자기 리뷰 모델 해상 | `no_pins` | 0.22 (< 0.5, **잠정**) | 감사 핀 미적용, 모델 = 선택 입력 또는 백엔드 기본. 되돌림 비용: 해상기 호출 한 번 |
| Q7 | CLI 거울 | `include` | 0.38 (< 0.5, **잠정**) | 가장 낮은 우선순위의 폐기 가능 마일스톤(M4). 폐기 규칙 plan.md §G |
| Q8 | 감사 도구 `baseBranch` 정렬 | `split_followup` | 1.00 | 감사 도구 불변. 후속 카드 필요(리더 발행). 결함 (b)는 자기 리뷰 한정으로 닫힘 |

### 결정별 질문·답·신뢰도

- **D1 — 질문**: 리더가 턴 종료 게이트의 리뷰 대상에서 빠지는 방법. **답**: `config_axis`(설정 축). **신뢰도**: 1.00. **처분**: 채택. 비카드(트리 스코프) 세션의 동작을 새 키가 정한다. 키 이름·값은 이 SPEC이 `workflow.codex.review_gate.tree_scope` = `review`|`skip` 으로 정했다(plan.md §B.2, decision-index Q1·Q2). REQ-CGS-003 을 REQ-CRO-003 으로 **명시적으로 개정**했다(spec.md HISTORY › Amendments).
- **D2 — 질문**: 온디맨드 자기 리뷰 표면. **답**: `new_review_tools`(전용 도구). **신뢰도**: 0.44 — 임계 0.5 미만이라 리더에 보고됨. **처분**: 잠정. 코디네이터가 조건부로 구속했다(아래 코디네이터 지시). plan.md §B.1 이 비교 표와 권고를 담았다.
- **D3 — 질문**: t1404 와 t1422 의 관계. **답**: `absorb_scoping_only`. **신뢰도**: 0.82. **처분**: 채택, 코드 판독으로 항목별 재검증했다(plan.md §D, T1·T3 흡수 / T2·T4·T5·T6·T7·T8 분리).

### 코디네이터 지시 (D2 조건부 구속 — Jev 판정 아님)

- D2: 전용 도구는 plan.md 가 "감사 도구에 카드 diff 대상을 추가"하는 재사용 대안과의 비교 표를 싣고, 새 도구 열이 **측정 가능한 근거**(권한 범위·판정/영수증 결속 분리·도구 목록 보유·미러/패리티 테스트 영향)로, 각각 file:line 인용과 함께 이길 때만 채택한다. 근거가 약하면 재사용 대안을 권고한다. 어느 쪽이든 SPEC 은 결함 (b) — `baseBranch` 대상의 서버 측 해상 — 를 고쳐야 한다: 호출자가 게이트와 같은 해상기가 정한 develop merge-base 기준 카드 diff 를 요청할 수 있어야 한다. 자기 리뷰 결과는 도구 출력/스키마와 교리에서 **advisory(non-binding)** 로 표시한다. D1·D3 은 확정이다.
- 반영: REQ-CRO-008(결함 (b) 수리), REQ-CRO-010(advisory·영수증 비결속), plan.md §B.1(비교 표·권고 (가)·(나) 근거 보존). Q3 은 리더 조건부 승인으로 해결됐다(위 Q1-Q8 표).

### 이 레인의 전제 점검 기록

- 설정 루트: §A.3 — 추적된 `workflow.yaml` 에 `review_gate` 키 0, primary 사본에 `enabled: true`. 레인 Stop 게이트는 이 저장소에서 설정상 off 로 읽힌다는 귀결은 코드 판독이며 라이브 관측이 아니다.
- 밀어붙임: 리더의 당면 문제는 primary 로컬 `enabled: false` 로도 풀린다. `tree_scope` 의 가치는 repo 전체 `enabled: true` 아래의 스코프 분리다(plan.md §G).
- 미관측: 카드 본문의 9~12분·282파일(리더 제공), t1395 처분 보고서 5건(경로 부재), 메인 세션의 도구 목록, codex 가 `baseBranch` 의 `branch` 필드에 SHA 를 받는지(라이브), primary `workflow.yaml` 이 추적 파일의 미커밋 수정인지.

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` (Tier M 산출 집합 + progress + decision gate `interview.decision_gate: on`).
- Tier: **M**. 근거와 한계는 plan.md §A — REQ 14·AC 16 으로 Tier M 상한(각 16) 안쪽이지만 AC 는 상한에 닿았다. 파일 수는 기계적 정합 때문에 Tier M 범위(5-15)를 넘는 추정(plan.md §H)이라 plan-auditor 가 tier 판정을 점검 대상으로 삼아도 좋다.
- plan_status: **audit-ready** — plan_complete_at: 2026-10-02. **독립 plan-audit 는 아직 수행되지 않았다**(이 줄은 작성 레인의 준비 신호이지 감사 판정이 아니다). 감사 트리: `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`(아티팩트 작성 시점 HEAD).
- 측정 원천: 본 트리(`.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 2026-10-02) 코드 좌표 직접 판독 — spec.md §A, plan.md §B·§D·§F. 리더 제공(미독립 재현): 9~12분·282파일(카드 본문), t1404 본문 3건 오탐.
- 결정: Q1-Q8 전부 해결(plan.md §G, decision-index.md, 위 Decision Log). 잠정 2건 — Q6(0.22)·Q7(0.38), 리더 통지됨. 남은 점검 항목: Tier 판정, `baseBranch` 의 SHA 수용 여부(미관측).
- 충돌 사전 검사: SPEC-CODEX-GATE-SCOPE-001(completed) REQ-CGS-003 만 개정(후속 개정, `amendment_of`). 인접 진행 카드: t1424(manager-develop `tools:` 줄 편집 — 병합 충돌 위험), t1399(`MOAI_KANBAN*` 삭제), t1423(감사관 문서) — plan.md §G.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
