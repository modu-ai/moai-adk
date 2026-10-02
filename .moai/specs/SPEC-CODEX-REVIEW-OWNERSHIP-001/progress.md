# SPEC-CODEX-REVIEW-OWNERSHIP-001 — progress

## Decision Log

원문 출처: `.moai/reports/t1422/jev-decisions.md`(Jev 판정 기록, 모델 `jev-1.13.0`, 트리 HEAD `c50da9c2f`). Jev 응답은 **표시 전용 판단 신호**이며 결정 권한은 레인, 0.5 미만 신뢰도는 리더 소관이다. 아래 표는 원문 그대로다.

| # | Question | Jev choice | Confidence | Probabilities | Disposition |
|---|----------|-----------|------------|---------------|-------------|
| D1 | How the leader stops being reviewed by the turn-end gate | `config_axis` | 1.00 | config_axis 1.00 · hard_skip 0.00 · role_detect 0.00 | Adopted. A new `review_gate` key decides what a non-card (tree-scope) session does; default keeps today's whole-tree review so product users are unchanged; this repo's primary config sets skip. No kanban env is read (t1399 deletes it; REQ-CGS-004 forbids env as a discriminator). |
| D2 | On-demand self-review surface | `new_review_tools` | 0.44 (**below 0.5 — reported to leader**) | new_review_tools 0.63 · extend_audit_tools 0.37 · cli_only 0.00 | Provisional: dedicated review tools. Plan-audit must test this against `extend_audit_tools`; reversible at plan stage. |
| D3 | Relation of t1404 to t1422 | `absorb_scoping_only` | 0.82 | absorb_scoping_only 0.88 · keep_separate 0.11 · absorb_all 0.01 | Adopted. Stale-main and non-card scoping cases are absorbed; the sync-gate `WCI_EXCLUDES` gap and doc-drift card t1406 stay separate. Absorption must be re-verified against code in the SPEC (premise check, `verification-claim-integrity.md` §1.1 surface 4). |

### Q1-Q8 해결 행 (2026-10-02, 운영자 위임 Jev `jev-1.13.0` + 리더 승인)

위 D1-D3 표는 `.moai/reports/t1422/jev-decisions.md` 와 바이트 동일하게 둔다. plan 단계에서 표면화된 오픈 결정 Q1-Q8 의 해결은 별도 표로 이어 적는다. 답변 원문은 레인 스크래치패드에 기록돼 있다(경로는 휘발성이라 인용하지 않음). **Q6 은 신뢰도 0.5 미만 — 잠정(리더 수용). Q7 은 2차 배치(아래)가 대체했다.**

| # | Question | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| Q1 | `tree_scope` 배포 형태 | `comment_example` | 0.74 | 템플릿에 주석 예시로만. 구조체는 파싱. 인벤토리·스키마·콘솔·i18n 불변(AC-CRO-015) |
| Q2 | 키 이름·값 | `workflow.codex.review_gate.tree_scope` = review\|skip | 1.00 | 최종 |
| Q3 | D2 자기 리뷰 표면 | 전용 도구 | (리더 조건부 승인) | 조건(비교 표·결함 (b) 수리·advisory) 충족: plan.md §B.1, REQ-CRO-008/010. Jev D2 intake 0.44 |
| Q4 | t1404 처분 | `edit_to_residual` | 0.92 | 착지 뒤 리더가 잔여(T2·T4·T5·T8)로 편집, 착지 전 닫지 않음. sync 인계 항목(코드 변경 아님) |
| Q5 | 추적 `workflow.yaml` 의 `enabled: true` | `do_not_commit` | 0.97 | 커밋하지 않음. 리더가 **착지 뒤** primary 의 `workflow.yaml`(추적 파일의 로컬 수정본)에 `tree_scope: skip` (운영자 소유, SPEC 은 편집 안 함). 레인 수단은 카드 리뷰 단계뿐 |
| Q6 | 자기 리뷰 모델 해상 | `no_pins` | 0.22 (< 0.5, **잠정**) | 감사 핀 미적용, 모델 = 선택 입력 또는 백엔드 기본. 되돌림 비용: 해상기 호출 한 번 |
| Q7 | CLI 거울 | `include` → **대체됨(2차 배치: drop)** | 0.38 (< 0.5, 잠정) → 0.83 | 2차 배치가 삭제로 대체 — 아래 B2-1 |
| Q8 | 감사 도구 `baseBranch` 정렬 | `split_followup` | 1.00 | 감사 도구 불변. 별도 카드 **t1426** 이 소유(리더 소관). 결함 (b)는 자기 리뷰 한정으로 닫힘 |

### 2차 결정 배치와 plan-audit 1회차 (2026-10-02, Jev `jev-1.13.0` + 리더 판정)

앞 표들 뒤에 이어 붙인 행이다(앞 표들은 수정하지 않았고 Q7 행만 "대체됨"으로 표기).

| # | Question / event | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| B2-1 | CLI 거울 포함 여부 (Q7 대체) | `drop` | 0.83 | 이 SPEC 에서 삭제: 구 REQ-CRO-011·CLI AC·M4 제거, 요구 재번호(14건). 결과: GLM 쪽 호스트 타임아웃·오래된 MCP 서버 우회로 없음, codex 쪽은 기존 `moai verify codex-review --project-root <tree>` 가 대체 — plan.md §G 위험 6·7, spec.md §E |
| B2-2 | 구조 — Claude Stop 게이트 asyncRewake 전환 | `sibling_spec` | 0.93 | 같은 카드의 형제 SPEC 으로 분리(오케스트레이터가 이 개정 뒤 작성). 이 SPEC 은 spec.md §E 한 단락과 plan.md §G 위험 1(같은 핸들러·시그니처·설정 루트 배선 충돌, 이 SPEC 이 먼저 착지·선취 금지)만 둔다 |
| B2-3 | 리더 판정 | Q6 `no_pins` 수용, Q7 수용(이후 B2-1 로 삭제), 키 이름 `workflow.codex.review_gate.tree_scope` = review\|skip 최종 | 리더 | Q6 은 신뢰도 0.22 의 잠정 결정이 리더 수용으로 굳어졌다 |
| B2-4 | 감사 도구 `baseBranch` 문제의 소유 | 카드 **t1426** 으로 분리 | 리더 | spec.md §E·plan.md 가 "후속 카드 필요" 문구를 t1426 인용으로 교체 |
| B2-5 | `tree_scope: skip` 기록 시점·주체 | 착지 **뒤** 리더가 primary 에 기록 | 리더 | SPEC 요구가 아니라 인계 항목(아래) |
| B2-6 | t1399 충돌 | plan.md 위험으로 명시 | 리더 | `reviewGateEnvContext`(`codex_review_scope.go:328-334`)가 참조하는 `config.EnvMoaiFactoryWorker` 가 t1399 의 `MOAI_KANBAN*` 삭제로 사라지면 컴파일 불가 — plan.md §G 위험 2 |
| PA-1 | **plan-audit 1회차 = FAIL 0.76** (Tier M 임계 0.80; must-fix D1-D3, should-fix D4-D14, notes D15-D17) | 개정 | — | 보고서 `.moai/reports/t1422/plan-audit.md`(로컬). 반영: D1 파리티 편집 M3 로 이동·§F 마일스톤 열; D2 AC-004 회귀 칸 전환·RED 이유 열; D3 AC-010 두 백엔드×세 verdict; D4 REQ-CRO-004·AC-004; D5 추적 파일의 로컬 수정본·인계 항목; D6 AC 15·REQ 14·파일 수 재추정·O-1; D7 REQ-CRO-008 축소·라이브 프로브 기록 단계·AC-008 픽스처 확장; D8 REQ-CRO-010·AC-009; D9 AC-005; D10 AC-001·AC-008; D11 plan.md §B.1 (다)·권한 범위 행·§H 기준선; D12 위험 6·7; D13 REQ-CRO-012·013·AC-013·014; D14 `focus` 삭제·`model` 입력 AC-011·i18n 강제 테스트 식별; D15 핸들러 배선 결정(시그니처 불변·트리 클래스일 때 루트 재해상); D16 좌표 정정; D17 양성 대조 추가/문구 완화 |

### 3차 결정 배치 — O-1~O-3 (2026-10-02, Jev `jev-1.13.0` + 리더 판정)

| # | Question | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| B3-1 | O-1 Tier | `keep_tier_m` (keep_tier_m 0.84 / retier_l 0.16) | 0.68 | Tier M 유지 (plan.md §G) |
| B3-2 | O-2 크기 3배 점검 | `accept_current_size` | 0.45 (< 0.5, Jev 수준 잠정) | **리더가 2026-10-02 에 현 크기(≈235 LOC, Tier M) 수용** — 크기 질문은 더 이상 잠정이 아니다. 근거: plan.md §H「3배 초과 사유」(좁은 기준선은 GLM 리뷰·어느 세션에서든의 MCP 도달·advisory 표식을 충족하지 못함). 되돌림 경로: 새 codex 리뷰 도구를 뺀다(≈165 LOC, 2.4배) |
| B3-3 | O-3 리더 제외의 지속성 | 코드로 미해결 | — | 명시적 운영자/리더 인계 항목로 유지. **리더가 primary `workflow.yaml` 로컬 수정과 동기화 지속성 인계를 수용**했고 첫 릴리스 동기화 때 확인한다. 이 SPEC 이 만족시킨 요구가 아니다 |

### 인계 항목 (이 SPEC 의 코드 밖 — 만족된 요구가 아니다)

1. **리더 제외의 전달(D5).** 착지 **뒤** 리더가 primary 체크아웃의 `/Users/goos/MoAI/moai-adk-go/.moai/config/sections/workflow.yaml`(추적 파일의 로컬 수정본; main 커밋본과 429 줄 차이, primary HEAD 는 `ref: refs/heads/main`)에 `workflow.codex.review_gate.tree_scope: skip` 을 쓴다. 운영자 소유 파일이며 이 SPEC·run 은 편집하지 않는다. **지속성은 미해결(plan.md O-3)** — `main` 이 릴리스 PR 로만 전진하므로 릴리스가 이 파일을 바꿀 때 로컬 수정과 부딪힐 수 있다. `moai update` 가 old-only 키를 유지한다는 것은 문서 주석으로만 관측했고(`internal/cli/update/backup/merge.go:20-22`) 병합/풀 흐름은 시도하지 않았다.
2. **t1404 편집.** 착지 뒤 리더가 t1404 를 잔여(T2·T4·T5·T8)로 편집한다. 착지 전에는 닫지 않는다(Q4). sync 단계 인계 항목.
3. **t1426.** 감사 도구 `baseBranch` 정렬 카드는 리더가 소유한다(Q8). 이 레인·run 은 카드를 발행하지 않는다.
4. **형제 SPEC.** Claude Stop 게이트 asyncRewake 전환 SPEC 은 오케스트레이터가 이 개정 뒤 작성하며 이 SPEC 은 그 결정을 선취하지 않는다.

### 결정별 질문·답·신뢰도

- **D1 — 질문**: 리더가 턴 종료 게이트의 리뷰 대상에서 빠지는 방법. **답**: `config_axis`(설정 축). **신뢰도**: 1.00. **처분**: 채택. 비카드(트리 스코프) 세션의 동작을 새 키가 정한다. 키 이름·값은 이 SPEC이 `workflow.codex.review_gate.tree_scope` = `review`|`skip` 으로 정했다(plan.md §B.2, decision-index Q1·Q2). REQ-CGS-003 을 REQ-CRO-003 으로 **명시적으로 개정**했다(spec.md HISTORY › Amendments).
- **D2 — 질문**: 온디맨드 자기 리뷰 표면. **답**: `new_review_tools`(전용 도구). **신뢰도**: 0.44 — 임계 0.5 미만이라 리더에 보고됨. **처분**: 잠정. 코디네이터가 조건부로 구속했다(아래 코디네이터 지시). plan.md §B.1 이 비교 표와 권고를 담았다.
- **D3 — 질문**: t1404 와 t1422 의 관계. **답**: `absorb_scoping_only`. **신뢰도**: 0.82. **처분**: 채택, 코드 판독으로 항목별 재검증했다(plan.md §D, T1·T3 흡수 / T2·T4·T5·T6·T7·T8 분리).

### 코디네이터 지시 (D2 조건부 구속 — Jev 판정 아님)

- D2: 전용 도구는 plan.md 가 "감사 도구에 카드 diff 대상을 추가"하는 재사용 대안과의 비교 표를 싣고, 새 도구 열이 **측정 가능한 근거**(권한 범위·판정/영수증 결속 분리·도구 목록 보유·미러/패리티 테스트 영향)로, 각각 file:line 인용과 함께 이길 때만 채택한다. 근거가 약하면 재사용 대안을 권고한다. 어느 쪽이든 SPEC 은 결함 (b) — `baseBranch` 대상의 서버 측 해상 — 를 고쳐야 한다: 호출자가 게이트와 같은 해상기가 정한 develop merge-base 기준 카드 diff 를 요청할 수 있어야 한다. 자기 리뷰 결과는 도구 출력/스키마와 교리에서 **advisory(non-binding)** 로 표시한다. D1·D3 은 확정이다.
- 반영: REQ-CRO-008(결함 (b) 수리), REQ-CRO-010(advisory·영수증 비결속), plan.md §B.1(비교 표·권고 (가)·(나) 근거 보존). Q3 은 리더 조건부 승인으로 해결됐다(위 Q1-Q8 표).

### 이 레인의 전제 점검 기록

- 설정 루트와 출처: spec.md §A.3 — 추적 확인(`ls-files --error-unmatch`), main 커밋본에 `review_gate` 0, primary HEAD `refs/heads/main`, primary 작업 사본은 main 커밋본과 429 줄 차이(이 레인이 읽기 전용 명령으로 실행; 이전 Gap 은 닫힘). 레인 Stop 게이트는 이 저장소에서 설정상 off 로 읽힌다는 귀결은 코드 판독이며 라이브 관측이 아니다.
- 밀어붙임: 리더의 당면 문제는 primary 로컬 `enabled: false` 로도 풀린다. `tree_scope` 의 가치는 repo 전체 `enabled: true` 아래의 스코프 분리다(plan.md §G).
- 미관측: 카드 본문의 9~12분·282파일(리더 제공), t1395 처분 보고서 5건(경로 부재), 메인 세션의 도구 목록, codex 가 `baseBranch` 의 `branch` 필드에 SHA 를 받는지(라이브), primary 의 로컬 수정이 릴리스 병합을 견디는지(O-3), 웹 i18n 테스트의 빨간 상태(추론만).

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` (Tier M 산출 집합 + progress + decision gate `interview.decision_gate: on`).
- Tier: **M**(O-1 로 올림). 개수 규칙: 요구 = `### REQ-` 제목 수 14, 수락 기준 = `### AC-` 제목 수 15(하위 ID 없음) — 둘 다 상한 16 이하. 파일: 코드+테스트 ≈15(M 대역 상단), 미러·문서 ≈25, 합 ≈40(plan.md §H).
- plan_status: **audit-ready (개정 2판)** — plan_complete_at: 2026-10-02. plan-audit 1회차 = FAIL 0.76(PA-1, 보고서 `.moai/reports/t1422/plan-audit.md`, 감사 트리 `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`). 이 개정이 D1-D17 을 반영했고 **2회차 감사는 아직 수행되지 않았다**(이 줄은 작성 레인의 준비 신호이지 감사 판정이 아니다).
- 측정 원천: 본 트리(`.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 2026-10-02) 코드 좌표 직접 판독 — spec.md §A, plan.md §B·§D·§F. 리더 제공(미독립 재현): 9~12분·282파일(카드 본문), t1404 본문 3건 오탐.
- 결정: Q1-Q9 해결(plan.md §G, decision-index.md, 위 Decision Log). 잠정 1건 — Q6(0.22, 리더 수용). O-1(Tier M 유지)·O-2(현 크기 리더 수용)는 해결, O-3(리더 제외의 지속성)은 코드로 미해결인 채 리더가 인계를 수용(B3-1~3, decision-index Q10–Q12). 남은 점검 항목: `baseBranch` 의 SHA 수용·미커밋 파일 포함 여부(미관측 — M3 기록 관측 단계).
- 충돌 사전 검사: SPEC-CODEX-GATE-SCOPE-001(completed) REQ-CGS-003 만 개정(후속 개정, `amendment_of`). 인접 진행 카드: t1424(manager-develop `tools:` 줄 편집 — 병합 충돌 위험), t1399(`MOAI_KANBAN*` 삭제), t1423(감사관 문서) — plan.md §G.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
