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
| Q1 | `tree_scope` 배포 형태 | `comment_example` | 0.74 | 템플릿에 주석 예시로만. 구조체는 파싱. 인벤토리·스키마·콘솔·i18n 불변(AC-015) |
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
| PA-2 | **plan-audit 2회차 = FAIL 0.79** (Tier M 임계 0.80, 마지막 허용 반복; 필수 R1, 권고 R2·N1-N5, 노트 N6-N11; D3·D9 부분 해결) | 개정 | — | 보고서 `.moai/reports/t1422/plan-audit-iter2.md`(로컬, 감사 트리 `a0d801409`, 점수 0.78/0.80/0.70/0.88 평균). 반영: R1 AC-010 을 모든 경로(`card`·`uncommitted`·비카드 조기 반환·빈 자료)·세 verdict·두 백엔드에서 `advisory`/`scope`/`base`/`backend`/`tree` 값으로 단정하고 `base == git merge-base develop HEAD`·`tree ==` 정규 루트를 값 비교, 빈 자료 픽스처 추가, AC-016 으로 분할(REQ-CRO-010 에 정의 문단); R2 AC-005 에 런처의 실제 값(`leader`·`worker-1`·`lane-1`·`1`·빈 값·부재) 행과 정책 파일 환경 미참조 정적 가드(`codex_review_tree_scope.go` 신규 파일); N1 두 MCP 규칙 사본 쌍은 바이트 동일·두 가드 시험을 목록과 AC-015 에·"Four of the twenty-two" grep; N2 게이트 수준 보존 시험을 M1 에서 먼저 작성·AC-004 (b) 를 그때까지 미측정으로; N3 primary 설정은 이동하는 상태 — 시각 표기 관측으로 낮춤; N4 감사 도구 스키마 스냅숏을 M1·DoD·AC-007 (f) 에; N5 리더 문장 조건부; N6 번호 단정 수 공개; N7 AC id 정정; N8 detached HEAD 위험 11; N9 heartbeat 미사용 명시; N10 P6-P8 verbatim; N11 형제 SPEC 존재·경계 교차 확인 |

### 3차 결정 배치 — O-1~O-3 (2026-10-02, Jev `jev-1.13.0` + 리더 판정)

| # | Question | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| B3-1 | O-1 Tier | `keep_tier_m` (keep_tier_m 0.84 / retier_l 0.16) | 0.68 | Tier M 유지 (plan.md §G) |
| B3-2 | O-2 크기 3배 점검 | `accept_current_size` | 0.45 (< 0.5, Jev 수준 잠정) | **리더가 2026-10-02 에 현 크기(≈235 LOC, Tier M) 수용** — 크기 질문은 더 이상 잠정이 아니다. 근거: plan.md §H「3배 초과 사유」(좁은 기준선은 GLM 리뷰·어느 세션에서든의 MCP 도달·advisory 표식을 충족하지 못함). 되돌림 경로: 새 codex 리뷰 도구를 뺀다(≈165 LOC, 2.4배) |
| B3-3 | O-3 리더 제외의 지속성 | 코드로 미해결 | — | 명시적 운영자/리더 인계 항목로 유지. **리더가 primary `workflow.yaml` 로컬 수정과 동기화 지속성 인계를 수용**했고 첫 릴리스 동기화 때 확인한다. 이 SPEC 이 만족시킨 요구가 아니다 |

### 4차 결정 배치 — Q13·Q14 (2026-10-02, plan-audit 3회차 PASS 뒤, Jev `jev-1.13.0` 운영자 위임)

| # | Question | Resolution | Confidence | Disposition |
|---|----------|-----------|------------|-------------|
| B4-1 | Q13 새 자기 리뷰 도구의 진행 알림(`notifyMCPProgress`) | `no_heartbeat_residual_risk` | 1.00 | 새 도구는 알림을 호출하지 않는다. 호스트 도구 타임아웃 노출은 명시된 잔여 위험(plan.md §G 위험 6, spec.md §E). REQ·AC 증가 없음 |
| B4-2 | Q14 빈 자료의 처분 | `skip_both_backends_inconclusive` | 1.00 | 빈 자료(런타임 접두 제외 뒤 카드 diff 가 비었거나 미커밋 diff 가 빈 경우)는 두 백엔드를 호출하지 않고 원인을 밝힌 `inconclusive` 를 돌려준다. **GLM 의 비추적 파일만 있는 경우는 첫 run 마일스톤에서 정한다**(아래 I3-2) |

Jev 판정은 표시 전용 신호이며 결정 권한은 레인·리더 소관이다. 이 두 행은 `decision-index.md` Q13·Q14 의 `Operator verdict:` 와 같다.

### 인계 항목 (이 SPEC 의 코드 밖 — 만족된 요구가 아니다)

1. **리더 제외의 전달(D5, N3 갱신).** primary 체크아웃의 `/Users/goos/MoAI/moai-adk-go/.moai/config/sections/workflow.yaml`(추적 파일의 로컬 수정본; primary HEAD 는 `ref: refs/heads/main`)은 **이동하는 운영자 상태**다. 2026-10-02 16:39 KST 관측: `codex.review_gate.enabled: false`(수정 15:32, 누가 바꿨는지는 관측하지 않음) — 이 시각에는 게이트가 이미 꺼져 있어 리더 제외에 `tree_scope: skip` 이 필요하지 않다. **`enabled: true` 로 다시 켤 때** 착지 뒤 리더가 `workflow.codex.review_gate.tree_scope: skip` 을 함께 쓴다(착지 시점에 파일을 다시 읽는다). 이 SPEC 은 그 값에 의존하지 않는다. 운영자 소유 파일이며 이 SPEC·run 은 편집하지 않는다. **지속성은 미해결(plan.md O-3)** — `main` 이 릴리스 PR 로만 전진하므로 릴리스가 이 파일을 바꿀 때 로컬 수정과 부딪힐 수 있다. `moai update` 가 old-only 키를 유지한다는 것은 문서 주석으로만 관측했고(`internal/cli/update/backup/merge.go:20-22`) 병합/풀 흐름은 시도하지 않았다.
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

- 설정 루트와 출처: spec.md §A.3 — 추적 확인(`ls-files --error-unmatch`), main 커밋본에 `review_gate` 0, primary HEAD `refs/heads/main`, primary 작업 사본은 main 커밋본과 429 줄 차이(최초 측정). **primary 로컬 값은 이동하는 상태**라 2026-10-02 16:39 KST 관측(`enabled: false`, 수정 15:32)으로만 인용하며 이 SPEC 은 그 값에 의존하지 않는다(N3). 레인 Stop 게이트는 이 저장소에서 설정상 off 로 읽힌다는 귀결은 코드 판독이며 라이브 관측이 아니다.
- 밀어붙임: 리더의 당면 문제는 primary 로컬 `enabled: false` 로 이미 풀려 있다(위 관측). `tree_scope` 의 가치는 repo 전체 `enabled: true` 아래의 스코프 분리다(plan.md §G).
- 미관측: 카드 본문의 9~12분·282파일(리더 제공), t1395 처분 보고서 5건(경로 부재), 메인 세션의 도구 목록, codex 가 `baseBranch` 의 `branch` 필드에 SHA 를 받는지(라이브), primary 의 로컬 수정이 릴리스 병합을 견디는지(O-3), 웹 i18n 테스트의 빨간 상태(추론만).

## plan-audit 이력 (2026-10-02)

| 회차 | 판정 | 점수 | 보고서 (로컬 증거) | 감사 커밋 |
|---|---|---|---|---|
| 1 | FAIL | 0.76 | `.moai/reports/t1422/plan-audit.md` | `3ae43ed8e78ffa673ca238227df6ca7202c1ce70` |
| 2 | FAIL | 0.79 | `.moai/reports/t1422/plan-audit-iter2.md` | `61befc5f90e36801f54a84c02b8fa4ff3dbced57` |
| 3 (마지막 허용) | **PASS** | **0.835** (Tier M 임계 0.80) | `.moai/reports/t1422/plan-audit-iter3.md` | `d242e0ea933ea0f9d237a166d78e6a1f78b419be` |

3회차: 필수 통과 기준 아홉 개 모두 통과, 필수 수정 0, 권고 2(I3-1, I3-2 — 비차단), 노트 5(I3-3 ~ I3-7), 이전 회차 미해결 0(D3·D9 가 해결로 바뀜). 감사 도구 환경에서 Grep 도구는 쓰이지 않았고 형식 verb 는 Write 도구로 만든 스크립트를 `sh <script>` 로 돌렸다.

## 해시 기준선 — Kickoff "산출물 불변" 점검용 (HEAD `0eb437eec67d03db2e077511341481bf7f6ae346`, 이 레인이 `shasum -a 256` 으로 계산)

```text
a1c230e77c07fc4cd4e9ade67e6b054398527a9cdc7cd78ed338d469ca1f4c4e  .moai/specs/SPEC-CODEX-REVIEW-OWNERSHIP-001/spec.md
b8895b339d943cae470e1236e3190797563332ced61cf0686d17c0a58a352c50  .moai/specs/SPEC-CODEX-REVIEW-OWNERSHIP-001/plan.md
5bb90833426902455c9220f54103ab239b69602338fe198c64bee33c0bd68767  .moai/specs/SPEC-CODEX-REVIEW-OWNERSHIP-001/acceptance.md
```

세 값은 감사 3회차가 읽은 값과 같다(감사 보고서 첫머리). 이 SPEC 의 `ComputeHash` 대상(`spec.md`·`plan.md`·`acceptance.md`)이다. **PASS 는 이 해시에 걸려 있으므로 세 파일을 고치지 않는다**; `progress.md`·`decision-index.md` 는 해시 대상이 아니다.

## M1 테스트 작성으로 이월 (SPEC 본문 수정 없음) — plan-audit 3회차 I3-1 ~ I3-7

PASS 를 유지하려고 spec/plan/acceptance 를 고치지 않는다. 아래는 M1(RED 확립·테스트 작성) 단계가 AC 문구를 넘어서 시험에 담을 항목이다. 감사 보고서가 "run 단계에서 값싸게 닫을 수 있다"고 적었다.

- **I3-1 (AC-005 env 행 — 권고).** 감사 관측: 런처는 `MOAI_KANBAN_LABEL=leader` 를 설정하지 않는다 — 칸 리더는 `MOAI_KANBAN=1`·id·리더 소켓 주소만 쓰고 라벨은 없으며(`kanban.go:193-206`), 라벨은 동반 세션이 자기 역할 라벨로 설정하고(`kanban.go:333-343`, `cc.go:319`), 팩토리 실행은 `MOAI_KANBAN`/라벨을 일부러 설정하지 않으며(`factory.go:665-668`) 팩토리 레인은 `MOAI_FACTORY_WORKER=<라벨>`·`MOAI_FACTORY_WORKERS=<n>` 을 쓴다(`factory.go:727-737`). 그러므로 AC-005 의 "런처가 실제로 쓰는 값" 문구와 R-leader 행은 정확하지 않다. **M1 시험이 할 일:** (1) 실제 칸 리더 행(`MOAI_KANBAN=1`, id, 주소, 라벨 없음), 팩토리 리더 행(`MOAI_FACTORY_WORKERS=8` 만), 동반 세션 행(`leader` 아닌 비어 있지 않은 라벨, 예 `run`)을 행렬에 더한다; (2) 변이 "다른 파일(`codex_review_scope.go`, 환경 읽기 2건 — P10)에 `isCompanion()` 도우미를 두고 정책 파일에서 호출"을 죽이는 행동 행(동반 라벨 행의 T2 결정이 부재 행과 같아야 함)을 포함한다 — 정적 가드는 이 변이를 못 본다.
- **I3-2 (빈 자료 정의 — 권고).** (a) 정의가 "비추적 비런타임 파일도 없을 것"을 요구하는데 `glm_review` 의 자료에는 비추적 파일이 들어가지 않으므로(`excluded_untracked` 로만 보고) 비추적 파일만 있는 카드/`develop` 트리는 정의상 "비어 있지 않음"이 되어 GLM 이 빈 diff 로 호출될 수 있다(`callGLMAudit` 는 받은 문자열을 그대로 보낸다, `mcp_glm.go:302-308`). Q14 verdict 가 이 경우를 첫 run 마일스톤에 맡겼다. **M1 시험이 할 일:** GLM 백엔드의 빈 자료를 "비추적 파일 유무와 무관하게 diff 가 빈 경우"로 시험에 고정하고(`excluded_untracked` 는 계속 보고), 비추적 파일만 있는 픽스처 한 건(GLM: HTTP 0 회·`inconclusive`·`excluded_untracked` 에 그 파일; codex: 현 정의)을 p5 에 더한다. 이 결정은 SPEC 본문의 정의와 다르므로 M1 에서 리더 확인을 받고 progress.md 에 기록한다. (b) `scope: card` 호출이 빈 자료에 닿을 때의 `base`: REQ-CRO-010 첫 절은 SHA, 마지막 절은 빈 문자열이다 — AC-010 (b) 는 p5 에서 빈 문자열이다. **M1 시험이 할 일:** AC 문구대로 빈 자료 경로의 `base` 를 빈 문자열로 단정하고 이 해석을 시험 주석에 적는다.
- **I3-3 (변이 점검의 "열림: 없음"은 너무 강하다 — 노트).** 아래 변이 점검 결과 절의 "죽임 67 / 열림 0"은 **작성자 판독 기준**이다. I3-1, I3-2, I3-7 이 표에 없는 생존 변이를 설명하므로 이 셋은 **생존 후보로서 M1 에서 시험**한다: (i) 정책 파일 밖 `isCompanion()` 도우미, (ii) GLM 비추적 파일만 있는 트리에서 빈 diff 호출, (iii) 카탈로그 수치만 올리고 도구 행을 쓰지 않는 구현, (iv) `project_root` 를 `resolveOptionalToolProjectRoot` 로 읽는 구현.
- **I3-4 (결정 인덱스 빈 칸 — 노트).** Q13·Q14 verdict 를 채웠다(위 4차 결정 배치). 남은 열린 칸 없음.
- **I3-5 (REQ-CRO-006 문구 대 AC-006 (b) — 노트).** REQ 는 두 경로가 "서로 어긋날 수 없다"고 하나 AC-006 (b) 는 설정 고아 워크트리에서 Claude 경로(primary 키)와 Codex 체인(`c.root`)이 갈라지는 기존 비대칭을 고정한다. AC 가 우선한다; M1 시험 주석에 이 비대칭이 의도임을 적는다.
- **I3-6 (`related_specs` 누락 — 노트).** spec.md 프론트매터가 `SPEC-CODEX-REVIEW-ASYNC-001`·`SPEC-WORKTREE-STATE-ROOT-001` 을 `related_specs` 에 담지 않는다. 스키마 위반이 아니고 spec.md 편집은 해시를 바꾸므로 **고치지 않는다** — 이 SPEC 이 sync 단계에서 열릴 때 함께 정리한다.
- **I3-7 (정합 변이 두 건 — 노트).** (i) 카탈로그 수치만 47/43 으로 올리고 새 도구 표 행(`moai-mcp-tools-catalogue.md` 의 감사 행 옆)을 쓰지 않는 구현은 `TestMCPToolCatalogueFiguresMatchRegistry`(수치만 읽음)와 docs-site 표 행(미검사; `project_root` 문장만 검사)을 모두 통과한다 — **M3 시험이 카탈로그 행 grep**(`grep -c 'codex_review' …/moai-mcp-tools-catalogue.md` ≥ 1, 두 사본)과 docs-site 4개 표 행 존재 단정을 더한다. (ii) `project_root` 를 `resolveToolProjectRoot`(부재 시 `CLAUDE_PROJECT_DIR`/서버 cwd)가 아니라 `resolveOptionalToolProjectRoot`(부재 시 루트 없음)로 읽는 구현은 REQ-CRO-007 의 "기존 도구들이 공유하는 계약"에 모호성이 있어(`mcp_project_root.go:92,153` 두 변형) 통과한다 — **M3 시험이 `project_root` 를 생략한 호출 픽스처 한 건**(어느 변형이 계약인지 정해 기록)을 더한다.
- 작은 어긋남(추정치): plan.md §H 가 새 테스트 파일을 "신규 ≈2"로 적었으나 §F.2 에는 세 곳(소유권 시험, `mcp_selfreview_test.go`, 에이전트 `tools:` 검사)이 나온다 — 추정(`≈`)이므로 결함이 아니다.

## 변이 점검 결과 (이 개정의 자체 스윕 — 작성 시점 판독, 마지막 허용 감사 반복 전)

> **정정(감사 3회차 I3-3).** 아래 "열림 0"은 작성자 판독 기준이다. 위「M1 테스트 작성으로 이월」의 I3-1·I3-2·I3-7 이 생존 후보를 추가하므로 "열림 0"을 닫힘으로 읽지 않는다.

acceptance.md §C 표가 16개 AC 모두에 대해 요구를 어기면서 기준을 만족하는 가장 값싼 변이와 그것을 죽이는 단정 칸을 적는다. 행 수: AC-001 5 · AC-002 5 · AC-003 2 · AC-004 3 · AC-005 5 · AC-006 3 · AC-007 6 · AC-008 6 · AC-009 5 · AC-010 6 · AC-011 2 · AC-012 3 · AC-013 5 · AC-014 3 · AC-015 5 · AC-016 3 = **67 개 변이 행, 죽임 67 / 열림 0**(감사 보고서 2회차의 10개에서 확장; 생존 7개 — AC-010 두 건·AC-005 `MOAI_KANBAN_LABEL == "leader"`·AC-015 두 건·AC-007 스키마·REQ-CRO-010 빈 자료 — 는 각각 AC-010 (a)-(f)·AC-005 (a) R-leader 행·AC-015 (ii) L12 및 가드 두 시험·AC-007 (f)·AC-010 (e) 로 죽는다). "죽임"은 구현이 없는 plan 단계의 판독이며 실행 관측이 아니다. 변이 하나가 다른 칸으로 옮겨 가는 경우(예: AC-005 의 환경 읽기를 다른 파일 도우미로 옮김)는 행동 행렬과 정적 가드가 서로의 대조로 짝지어져 있다. 한 가지 정직한 한계: AC-004 (b) 는 M1 의 게이트 수준 시험이 변경 전 트리에서 초록으로 관측되기 전까지 회귀 칸이 아니라 미측정이다(L11).

**검증 Gap(감사 보고서 인용).** 감사 도구 환경에서 Grep 도구가 쓸 수 없었고 verb 스크립트를 셸 here-document 로 쓰는 시도가 격리 가드에 거부되어 Write 도구로 만든 스크립트를 평범한 `sh <script>` 로 돌렸다. CN-4 verb 는 한국어 순서어를 보지 못하고 plan 에 `Exit:` 줄이 없어 MP-9 는 손 판독으로 판정했다고 감사가 적었다. 이 레인은 같은 verb 를 돌리지 않았고 헤딩 개수와 `moai spec lint` 로만 점검했다.

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` (Tier M 산출 집합 + progress + decision gate `interview.decision_gate: on`).
- Tier: **M**(O-1 로 올림). 개수 규칙: 요구 = `### REQ-` 제목 수 14, 수락 기준 = `### AC-` 제목 수 16(하위 ID 없음; R1 분할로 AC-016 추가) — 둘 다 상한 16 이하이고 AC 는 상한에 닿았다. 번호 단정 62개(acceptance.md §E). 파일: 코드+테스트 ≈15(M 대역 상단), 미러·문서 ≈25, 합 ≈40(plan.md §H).
- plan_status: **audit-ready** — plan_complete_at: 2026-10-02. plan-audit 1회차 = FAIL 0.76(PA-1), 2회차 = FAIL 0.79(PA-2), **3회차 = PASS 0.835**(보고서 `.moai/reports/t1422/plan-audit-iter3.md`, 감사 커밋 `d242e0ea933ea0f9d237a166d78e6a1f78b419be` — 위「plan-audit 이력」). 해시 기준선은 위「해시 기준선」절의 세 값이다. 감사 3회차의 비차단 항목은「M1 테스트 작성으로 이월」에 있다(SPEC 본문은 수정하지 않는다).
- 측정 원천: 본 트리(`.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 2026-10-02) 코드 좌표 직접 판독 — spec.md §A, plan.md §B·§D·§F. 리더 제공(미독립 재현): 9~12분·282파일(카드 본문), t1404 본문 3건 오탐.
- 결정: Q1-Q9 해결(plan.md §G, decision-index.md, 위 Decision Log). 잠정 1건 — Q6(0.22, 리더 수용). O-1(Tier M 유지)·O-2(현 크기 리더 수용)는 해결, O-3(리더 제외의 지속성)은 코드로 미해결인 채 리더가 인계를 수용(B3-1~3, decision-index Q10–Q12). 남은 점검 항목: `baseBranch` 의 SHA 수용·미커밋 파일 포함 여부(미관측 — M3 기록 관측 단계).
- 충돌 사전 검사: SPEC-CODEX-GATE-SCOPE-001(completed) REQ-CGS-003 만 개정(후속 개정, `amendment_of`). 인접 진행 카드: t1424(manager-develop `tools:` 줄 편집 — 병합 충돌 위험), t1399(`MOAI_KANBAN*` 삭제), t1423(감사관 문서) — plan.md §G.

## §E.2 Run-phase Evidence

Run-phase implementer: manager-develop role, `cycle_type=tdd`, lane worker for card t1422. Scope of this run: M1 and M2 only. Every figure below is a command plus the observed output, in this run, against the tree named next to it. Raw logs live under `.moai/reports/t1422/red/` (gitignored, local evidence — not citable from another machine; the deciding lines are quoted here).

### M1 — regression lines, preserve test, schema snapshot, RED

**Pre-change tree.** `git rev-parse HEAD` → `984d649577362d9c10e9a47e21326b6e6241f708`, branch `WT-codex-review-lane-scope`, `git status --short` clean at start. Toolchain: `golangci-lint --version` → `v2.1.6` (the CI version); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.` (baseline: nothing pre-existing to separate from new).

**Regression lines, observed GREEN before any production change** (one env-scrubbed compound call each: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_AUTONOMY_TIER MOAI_PROFILE_LEASE_TOKEN && go test … -count=1 -v`; swept set counted: `-list` named 10 tests, the run printed 31 `=== RUN` lines (10 top level + subtests) and 10 top-level `--- PASS`; log `m1-regression-pre.txt`):

```text
--- PASS: TestCodexAudit_RequiredGateBlocksWhenBinaryAbsent (0.01s)        # codex_audit required-gate control (AC-016 (d))
--- PASS: TestReviewGate_FailOpenOnMissingCodex (0.05s)                      # fail-open
--- PASS: TestCodexReviewGate_TreeScopeRequestShapeUnchanged (1.37s)         # AC-003 / P6
--- PASS: TestCodexReviewGate_CardScopeRequestIsCardDiff (2.22s)             # AC-004 (a) / P7
--- PASS: TestCodexReviewScope_UnidentifiedFallsToTree (3.25s)               # AC-004 (b) resolver half / P8
--- PASS: TestCodexReviewGate_CardScopeFailOpenOnMissingReviewer (1.89s)     # fail-open, card scope
--- PASS: TestMCPToolCatalogueDocsStayMirrorIdentical (0.01s)
--- PASS: TestMCPToolCatalogueFiguresMatchRegistry (0.00s)
--- PASS: TestReviewGateReaders_HonourNestedWorkflowKeyPath (0.05s)
--- PASS: TestReviewGateReaders_AgreeWithConfigLoader (0.05s)
ok  	github.com/modu-ai/moai-adk/internal/cli	10.225s
```

`go test ./internal/mcp/ -count=1 -v` (log `m1-mcp-pre.txt`): `TestMoaiMCPTools_CatalogSize`, `…WriteCapableSet`, `…NoDuplicateNames`, `TestMoaiMCPToolNames_MatchesCatalog` all PASS, `ok  github.com/modu-ai/moai-adk/internal/mcp  0.292s` (the `wantCatalogSize = 45` invariant is green). `go test ./internal/config/ -run '^(TestShippedConfigKeysHaveReaders|TestTemplateWorkflowYAML_JevShipsOff)$' -count=1 -v` → exit 0 (log `m1-config-pre.txt`).

Source-level preserve readings on the same tree: `grep -l mcp__moai__codex_audit .claude/agents/moai/manager-develop.md …manager-docs.md …manager-lead.md …plan-auditor.md …sync-auditor.md` → `.claude/agents/moai/plan-auditor.md` and `.claude/agents/moai/sync-auditor.md` only (P1); `shasum -a 256 internal/config/testdata/shipped_key_inventory.yaml` → `a3d2c397827130d71047fad9aa59c6b690d64394b222dca56adc44db149fc409` (P5); `grep -c tree_scope internal/config/testdata/shipped_key_inventory.yaml` → `0` (P3).

**New gate-level preserve test (plan.md §I M1, N2).** `TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree` in `internal/cli/codex_review_gate_wtnobase_test.go`, authored WITHOUT the skip key, run before any production change on HEAD `984d64957` (log `m1-wtnobase-preserve-pre.txt`):

```text
=== RUN   TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree
--- PASS: TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree (1.56s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.013s
```

It asserts: one reviewer lookup, request `{target: uncommittedChanges, cwd: <tree>}`, one scope row of class tree carrying branch `WT-orphan` with a basis naming "merge base unavailable". AC-004 (b) now counts as a regression line.

**Audit-tool schema snapshot (plan.md §I M1, N4; AC-007 (f)).** A server was built from this tree (`go build -o <scratch>/moai-pre ./cmd/moai` exit 0, HEAD `984d64957`, built from the working tree that carries no production change at snapshot time) and spoken to over stdio (`initialize`, `notifications/initialized`, `tools/list`); `tools/list` returned 45 tools (`jq … | length` → `45`). The four audit tools were extracted with `jq -S -s '[ .[] | select(.id==2) | .result.tools[] | select(.name=="codex_audit" or .name=="glm_audit" or .name=="claude_audit" or .name=="audit_multi") | {name, inputSchema, outputSchema}] | sort_by(.name)'` into `.moai/reports/t1422/red/audit-tools-schema-pre.json` (20520 bytes, tools `audit_multi`, `claude_audit`, `codex_audit`, `glm_audit`; `audit_multi` declares no output schema; the `target` enum is `["uncommittedChanges","baseBranch"]` on the other three).

```text
sha256  6910a4678ee5c82966a8e12af5e6762e216d109577a606b9ad0165b76b134ce6  .moai/reports/t1422/red/audit-tools-schema-pre.json
```

(A first attempt without `-s` wrote two JSON documents because the response stream has two lines; it was discarded and redone with `-s` — only the hash above is the snapshot.)

**RED (M1).** Verbatim logs and the semantic classification are in `.moai/reports/t1422/red/m1-red1-compile-failure.txt`, `m1-red2-runtime.txt`, `m1-red-semantic.txt`. Pre-implementation tree: HEAD `984d64957` + uncommitted tests + declaration-only stub (no policy code; neither `HandleCodexReviewGate` nor `codexReviewMember` reads `tree_scope`).

- Run 1, `go test ./internal/cli/ -run '^TestTreeScopeReader_TruthTable$' -count=1 -v` → exit 1, `FAIL … [build failed]`, `undefined: treeScopeSkipLogger`, `undefined: reviewGateTreeScopeReader`, `undefined: readCodexReviewGateTreeScope`, … → **TOOL_FAILURE** (never counted as RED; reclassified by taking the stub step below).
- Run 2 (15 exact names, `-count=1 -v`, 91 `=== RUN` lines) → exit 1, 11 FAIL / 4 PASS. **EXPECTED_RED** (11, each at its intended assertion): `TestTreeScopeReader_TruthTable` (`readCodexReviewGateTreeScope = "review", want "skip"`), `TestTreeScopeReader_AgreesWithConfigLoader` and `TestReviewGateReaders_AgreeWithConfigLoader` (`reader = "review", config loader = "skip" (schema drift)`), `TestCodexReviewGate_TreeScopeSkip` (`a skip must precede the detector, the lookup and the review (lookups=1 detects=1 reviewed=true)`), `TestCodexStopChain_TreeScopeSkip` (member 6 `Decision:deny … ReceiptRead:true`), `TestTreeScopePolicy_EnvMatrix` (`T1 develop tree gate path: skipped=false, want true` in all 8 rows), `TestTreeScopePolicy_SourceReadsNoEnvironment` (`open codex_review_tree_scope.go: no such file or directory`), `TestTreeScopePolicy_SameDecisionOnBothPaths`, `TestTreeScopePolicy_EachPathReadsItsOwnEnabledRoot`, `TestTreeScopePolicy_ConfigOrphanedWorktree`, `TestTreeScopeSkipRow`. **PASS (4)**: the preserve lines `TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree`, `TestTreeScope_NonSkipValuesKeepTreeRequest`, `TestTreeScopeSkip_WTSessionsStillReviewed`, `TestTreeScopePolicy_ExplicitProducerIgnoresPolicy` (they must stay green after M2; the mutant probe in the M2 section shows they bite).

**Carried-over items (plan-audit 3rd round) — disposition in M1.**

| Item | Disposition |
|---|---|
| I3-1 env rows | Done in `TestTreeScopePolicy_EnvMatrix`: rows are the values the launchers really set — kanban leader without a label, the acceptance.md "label leader" row, factory leader (`MOAI_FACTORY_WORKERS` only), kanban companion (`MOAI_KANBAN_LABEL=run`, no `MOAI_KANBAN`), factory worker, factory lane, all keys set-but-empty, all absent. The key list is derived from `internal/config/envkeys.go` by parsing it, so a key added or deleted by t1399 moves the matrix. A helper-outside-policy mutant is killed by the behaviour rows (the static guard cannot see it; it is the counterpart). Static guard: `TestTreeScopePolicy_SourceReadsNoEnvironment` (go/parser over `codex_review_tree_scope.go`, positive control: the same scan finds environment reads in `codex_review_scope.go`). |
| I3-2 empty-material definition (GLM) | M3 — untouched. |
| I3-3 survivor candidates | (i) helper outside the policy file: covered in M1/M2 as above; (ii)–(iv): M3 — untouched. |
| I3-4 | Nothing to do (decision-index filled). |
| I3-5 asymmetry comment | Recorded in the doc comment of `TestTreeScopePolicy_ConfigOrphanedWorktree`. |
| I3-6 `related_specs` | Not touched (spec.md body stays out of scope of this role; sync-phase item). |
| I3-7 catalogue/`project_root` mutants | M3 — untouched. |

**Gaps (M1).** The raw logs under `.moai/reports/t1422/red/` are gitignored local evidence; this section quotes the deciding lines. `go test` exit codes were not captured as `$?` (the isolation guard refuses `echo $?` in a compound call); the process status was observed through the tool result (non-zero reported as `Exit code 1` for the RED runs, no error for the green ones) and through the `ok`/`FAIL` summary lines. The `tools/list` snapshot was taken over stdio from a binary built from this tree, not from the long-running MCP server of the session (that one is `v3.2.0-rc.23`, an older build — plan.md §G risk 7).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- 입력: tier M · 범위 ≈40 파일(코드+테스트 ≈15, 미러·문서 ≈25) · 도메인 5(Go 소스, 훅·설정 템플릿, 에이전트 정의, 규칙, docs-site) · 파일 언어 혼합(Go + markdown/yaml) · 동시성 이득 LOW(코딩 중심, 마일스톤 간 의존이 강함).
- 평가: `direct` 미선택(사소하지 않음) · `serial` **선택** · `fanout` 미선택(코딩 중심, 병렬 읽기 이득이 작음) · `sweep` 미선택(균일한 기계적 변환 아님).
- Decision: serial
- 근거: 마일스톤 M1→M5 가 서로의 산출(RED 로그, 정책 함수, 등록 도구)에 의존하고 한 작업 트리에 쓰기 에이전트는 하나여야 한다. 마일스톤마다 구현 에이전트를 순서대로 띄운다.

### Kickoff 결정 (자율 형태, `.claude/rules/moai/workflow/auto-semantics.md` §9.1)

- 충족 조건: ① 독립 plan-audit 판정 PASS(3회차 0.835, 합격선 0.80) ② plan 단계 audit-ready 기록(§E.1) ③ plan-artifact 해시 불변 — 위「해시 기준선」세 값을 2026-10-02 에 `shasum -a 256` 으로 다시 계산해 일치를 관측 ④ 열린 차단 없음(잠정 결정은 리더가 수용, O-3 은 리더 인계 항목).
- keep-set 해당 없음: 환경상 불가능한 작업 없음, 운영자 보유 작업 없음, 외부 공유 시스템에 대한 되돌릴 수 없는 작업 없음(push·PR·develop 통합은 리더 일괄 소관으로 남는다).
- 참고: 이 카드는 리더 직접 배차라 팩토리 기록에 없고 `factory_decide` 는 레인에 거부되므로 결정 기록은 여기에 둔다.

decision record: decided_by=lane-7 (claude main session, orchestrator role) evidence_refs=.moai/reports/t1422/plan-audit-iter3.md (PASS 0.835, audited commit d242e0ea9), .moai/specs/SPEC-CODEX-REVIEW-OWNERSHIP-001/progress.md §E.1 + 해시 기준선 ladder_path=gate row plan→run Kickoff AUTONOMOUS (auto-semantics.md §9.1)
