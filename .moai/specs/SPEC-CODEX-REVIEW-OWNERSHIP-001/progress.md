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

Run-phase implementer: manager-develop role, `cycle_type=tdd`, lane worker for card t1422. Scope of this run: M1 and M2 (first run), M3 (a later run on the same worktree), M4 (a fourth run on the same worktree). Every figure below is a command plus the observed output, in this run, against the tree named next to it. Raw logs live under `.moai/reports/t1422/red/` (gitignored, local evidence — not citable from another machine; the deciding lines are quoted here).

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

**M1 commit.** `fdb4932eb` (parent `984d64957`): the preserve test, the RED tests, the declaration-only stub, the `internal/config` field/constants/normaliser, the SPEC `status: draft → in-progress` flip, and this section. It is deliberately a RED commit (the policy tests fail at it); the commit repo has no git hooks (`git config core.hooksPath` → `/dev/null`) and nothing rejected it. Committing the tests before the implementation lets the commit graph witness the order (verification-claim-integrity §2.3).

### M2 — the policy (REQ-CRO-001..006)

**Change set (production).** New `internal/cli/codex_review_tree_scope.go` (reader `readCodexReviewGateTreeScope`, policy `treeScopeSkipApplies`, skip row `treeScopeSkipRow`/`logTreeScopeSkip`, seams `reviewGateTreeScopeReader` and `treeScopeSkipLogger`); wiring in `HandleCodexReviewGate` (step 3a, after the scope row, before the self-gate, root `reviewGateConfigRoot(projectDir)` resolved lazily and only for a tree-class session) and in `codexReviewMember` (after the scope row, before the self-gate, root `c.root`, outcome allow / not-applicable / reason `tree_scope=skip`, no receipt read); the handler header and the `projectDir` comment corrected (the "free of config I/O" and "CONFIG root only" statements); `internal/config` carries `CodexReviewGateConfig.TreeScope`, `CodexReviewGateTreeScopeReview/Skip`, `NormalizeCodexReviewGateTreeScope` and the default `review`. The stub was deleted (`git rm`). `HandleCodexReviewGate`'s signature is unchanged; `produceCodexReviewReceipt` and `moai verify codex-review` do not read the policy; the shell wrapper, the template and the shipped-key inventory are untouched (M4 owns the commented example). Size: 53 non-comment, non-blank lines in the new file + 6 wiring lines + about 15 config lines against plan.md §H's ≈70 for the policy — within the estimate.

**Deviation from plan.md §B.2 wording, recorded.** The plan names a helper `(scope, configRoot) → (skip bool, reason)`. The implementation is `treeScopeSkipApplies(scope, configRoot func() string) bool`: the root is lazy so a card session costs no root resolution, and the reason is not a return value because it has one value (`tree_scope=skip` on the chain, the logged row on the Claude path). Behaviour and the REQ-CRO-006 "one shared function" property are unchanged.

**GREEN (M2).** After deleting the stub and adding the policy file, the same 15 exact names as RED run 2 (`go test ./internal/cli/ -run '^(<15 names>)$' -count=1 -v`, env-scrubbed one compound call, log `m2-green1.txt`): 91 `=== RUN` lines (same count as RED run 2), all 15 top-level `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli  41.388s`. The final union run (50 regression/preserve names + the 14 new names, `-count=1 -v`, tree = HEAD `fdb4932eb` + the M2 working tree, log `m2-final-green.txt`): 64 top-level `--- PASS`, no FAIL/SKIP line, `ok  github.com/modu-ai/moai-adk/internal/cli  204.233s`. `go test ./internal/config/ ./internal/mcp/ ./internal/settings/ -count=1` → `ok` ×3 (config 12.270s, mcp 0.275s, settings 1.607s); `shasum -a 256 internal/config/testdata/shipped_key_inventory.yaml` → `a3d2c397827130d71047fad9aa59c6b690d64394b222dca56adc44db149fc409` (identical to the pre-change P5 value), `grep -c tree_scope` on it → `0`.

**Mutant probe (the criteria bite).** Three mutants were applied by edit and reverted by edit (`git status --short` afterwards shows only the intended M2 files; the probe helper file was deleted):

| Mutant | Result | Log |
|---|---|---|
| A: skip on the scope CLASS alone (drop the `cardScopeFromBranch` check) | `TestTreeScopeSkip_WTSessionsStillReviewed/WT-_session_without_a_develop_base` FAIL and `TestTreeScopePolicy_SameDecisionOnBothPaths/WT-_session_without_base,_skip` FAIL (the card-session leg and the other rows PASS) — REQ-CRO-004's swallow is caught | `m2-mutant-a-class-only.txt` |
| B: an environment read in a helper OUTSIDE the policy file, called from it (`os.Getenv(config.EnvMoaiKanbanLabel) != ""` in `codex_review_mutant_probe.go`) | `TestTreeScopePolicy_EnvMatrix` FAIL on rows `label leader` and `kanban companion` (the rows that set a label); the static guard stayed PASS — it cannot see this mutant, the behaviour matrix is its counterpart | `m2-mutant-b-env-helper.txt` |
| C: `os.Getenv` directly inside the policy file | `TestTreeScopePolicy_SourceReadsNoEnvironment` FAIL: `the policy file must read no environment, found [os.Getenv config.EnvMoaiKanbanLabel]` | `m2-mutant-c-env-in-policy.txt` |

**Build, vet, lint, coverage, boundary.** `go build ./...` → `build-exit=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild-exit=0`; `go vet ./internal/cli/ ./internal/config/` clean; `gofmt -l` on the changed Go files → no output; `golangci-lint run --timeout=5m ./internal/cli/... ./internal/config/...` (v2.1.6) → `0 issues.` (baseline was `0 issues.`, so NEW = 0). Coverage, `go test ./internal/cli/ -run '^(<the 15 names>)$' -coverprofile=…` then per-function `go tool cover -func`: `readCodexReviewGateTreeScope 100.0%`, `treeScopeSkipApplies 100.0%`, `treeScopeSkipRow 100.0%`, `logTreeScopeSkip 75.0%` (the unreachable `json.Marshal` error branch), whole new file 19 of 20 statements = **95.0%** (target 85); `NormalizeCodexReviewGateTreeScope 100.0%` (`internal/config/codex_review_gate_tree_scope_test.go`). `HandleCodexReviewGate` 78.3% and `codexReviewMember` 63.3% are the whole functions under this 15-test subset — the added branches are covered, the rest of those functions is exercised by the wider suites. Boundary: `grep -rn --include='*.go' 'AskUserQuestion' internal/cli internal/config | grep -v _test.go | grep -v '// ' | wc -l` → `18` pre-existing matches (help-text strings); `git diff -U0 984d64957 HEAD -- internal | grep '^+' | grep -c AskUserQuestion` → `0`, and the same grep on the new policy file, `codex_review_gate.go`, `codex_stop_chain.go`, `defaults.go`, `types.go` → no output (exit 1): the change adds none.

**AC matrix for the ACs M1/M2 own** (command = `go test ./internal/cli/ -run '^<test>$' -count=1 -v`, env-scrubbed; result = `--- PASS` in `m2-final-green.txt` on the tree named above):

| AC | Observed by | Result |
|---|---|---|
| AC-001 | `TestTreeScopeReader_TruthTable` (18 fixtures + empty root), `TestTreeScopeReader_AgreesWithConfigLoader`, `TestReviewGateReaders_AgreeWithConfigLoader` (+1 fixture), `TestNormalizeCodexReviewGateTreeScope` | PASS |
| AC-002 | `TestCodexReviewGate_TreeScopeSkip`, `TestCodexStopChain_TreeScopeSkip` (variants `main`, `develop`, `detached`, `non-git`; lookups/detects/wire requests all 0, one skip row, review control), `TestTreeScopeSkipRow` | PASS |
| AC-003 | `TestCodexReviewGate_TreeScopeRequestShapeUnchanged`, `TestTreeScope_NonSkipValuesKeepTreeRequest` (absent, review, unknown, misplaced flat skip) | PASS (regression line, green before and after) |
| AC-004 | `TestCodexReviewGate_CardScopeRequestIsCardDiff`, `TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree`, `TestTreeScopeSkip_WTSessionsStillReviewed` (both paths, card and WT- without base, skip configured) | PASS ((b) counted as regression only after its own pre-change PASS above) |
| AC-005 | `TestTreeScopePolicy_EnvMatrix` (8 rows × 2 trees × 2 paths), `TestTreeScopePolicy_SourceReadsNoEnvironment` | PASS (mutants B, C killed) |
| AC-006 | `TestTreeScopePolicy_SameDecisionOnBothPaths`, `TestTreeScopePolicy_EachPathReadsItsOwnEnabledRoot`, `TestTreeScopePolicy_ConfigOrphanedWorktree`, `TestTreeScopePolicy_ExplicitProducerIgnoresPolicy` | PASS |

**Gaps (M2).** (1) `internal/web`, `internal/template` and `internal/hook` tests were not run: no file there was touched, and `tree_scope` is not a shipped key (the inventory hash and the `grep -c` above are the evidence); the leader's batch CI covers them. (2) The package-wide `./internal/cli` suite was not run (AGENTS.md §4, gitflow-lane-protocol §8): only the 64 named tests, chosen by `go test -list` over the gate, scope, receipt, stop-chain and catalogue-guard names — `grep -rn "HandleCodexReviewGate(" internal --include='*_test.go'` finds callers in `codex_review_gate_test.go` (10), `codex_review_scope_test.go` (9), `codex_stop_chain_golden_test.go` (2), `codex_rpc_error_test.go` (1) — all of which are in the 64-test run — plus `codex_review_gate_live_test.go` (1, a live-codex test, not run: it needs a real codex binary) and the two new files. A caller outside those files would not have been seen; the grep lists none. (3) The policy tests are heavier than the plan's ≈350–420 test LOC estimate for the whole SPEC (`codex_review_ownership_test.go` is 733 lines per `wc -l`, `TestTreeScopePolicy_EnvMatrix` ran 9.60 s in the final GREEN run): the env matrix multiplies eight rows by two trees by two paths as plan-audit I3-1 asked; the cost is reported, not trimmed. (4) A first `jq` snapshot attempt wrote two JSON documents and was redone (see M1) — only the final hash is the snapshot. (5) Nothing was pushed; the lane never pushes.

**Residual risk (M2).** The policy is only as good as the resolver's `WT-` evidence: a card worktree on a detached HEAD (rebase in progress) resolves `Branch == ""` and is skipped under `tree_scope: skip` (plan.md §G risk 11, a design choice named there). The reader is hand-rolled like its `enabled` sibling, so a future change of the key's place in `workflow.yaml` must move both readers and the loader-agreement pins together. `treeScopeSkipApplies` reads the config root once per gate evaluation of a tree-class session; with the key absent that is one extra file read per Stop.

### M3 — the self-review tools and the parity-enforced edits (REQ-CRO-007..010, REQ-CRO-014 M3 slice)

Run by a later manager-develop worker (`cycle_type=tdd`) on the same worktree (the run was cut by a usage limit once and resumed; nothing finished was redone). Pre-change tree: `git rev-parse HEAD` → `7919da7315acc2f8071b6dede9bcad961da5566f` (branch `WT-codex-review-lane-scope`, `git status --short` clean). All tests below ran as one env-scrubbed compound call (`unset MOAI_KANBAN … MOAI_PROFILE_LEASE_TOKEN && go test … -count=1 -v`); raw logs are under `.moai/reports/t1422/red/` (gitignored).

**Pre-flight (pre-change tree `7919da731`).** `go build ./...` → `build-exit=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild-exit=0`; `golangci-lint --version` → `v2.1.6` (CI version); `golangci-lint run --timeout=5m ./internal/cli/... ./internal/mcp/...` → `0 issues.` (`m3-lint-baseline.txt`; nothing pre-existing to separate from new).

**RED.** Two new test files, `internal/cli/mcp_selfreview_test.go` (23 test functions) and `internal/cli/mcp_selfreview_fixture_test.go` (real git fixtures, an in-process MCP client, stub reviewers over the existing `withCodexSession` / `withGLMSeams` seams), were written first and address the tools by name over an in-process client, so the package compiled without a stub: `go test ./internal/cli/ -run '^$'` → `ok … [no tests to run]` (`m3-red1-compile.txt`) — there was no compile-failure RED to lower. Run (`m3-red2-runtime.txt`, pre-implementation tree = HEAD `7919da731` + the two test files): 92 `=== RUN` lines, exit 1, 21 FAIL / 2 PASS. **EXPECTED_RED (21)**, each at its intended assertion: `invalid params: tool 'codex_review' / 'glm_review' not found: tool not found` (the call-by-name tests), `glm_review is not registered`, `open mcp_selfreview.go: no such file or directory`, and the catalogue / docs-site / `project_root` inventory rows absent. **PASS (2)** are preserve lines, green before any change: `TestSelfReview_AuditToolSchemasUnchanged` (AC-007 (e)(f) against the M1 snapshot, sha256 `6910a467…134ce6`) and `TestSelfReview_AuditToolHoldersUnchanged` (P1). TOOL_FAILURE: none. Semantic record: `m3-red-semantic.txt`. One earlier draft of the heartbeat test skipped (the in-process client does not surface server notifications); it was reworked to drive `MCPServer.HandleMessage` over a counting session before the recorded RED run.

**Which parity guards go red when the two tools are registered, before any doc / figure / i18n edit** (plan.md §I M3 operating note; `m3-red-guard-names.txt`, logs `m3-guards-*-before-docs.txt`; tree = HEAD `7919da731` + `mcp_selfreview.go`, the two registrations, the two catalog lines, the new tests):

```text
internal/mcp   RED   TestMoaiMCPTools_CatalogSize
internal/cli   RED   TestMCPToolCatalogueFiguresMatchRegistry, TestProjectRootDocMatchesServer,
                     TestDocsSiteProjectRootMatchesServer (4 locale subtests), TestCodexAuditMCPTool
               green TestMCPToolCatalogueDocsStayMirrorIdentical (reds only on a ONE-SIDED edit), TestMoaiMCPServer_RegistrationMatchesCatalog,
                     TestMoaiMCPServer_AnnotationsMatchCatalog, TestAC_C_005_PerToolFieldsInSchema
internal/web   RED   TestDataI18nKeysSubsetOfDictionary, TestI18nKeySetParity
               green TestI18nKeyCoverageForward, TestI18nUntranslatedValues, TestMCPConsoleRendersAllTools (+5 more)
internal/settings    ok
```

plan.md §F.2 predicted `TestI18nKeyCoverageForward`, `TestI18nUntranslatedValues` and `TestMCPConsoleRendersAllTools` would go red ("inference from test source, red state not run"). Observed: they stay green; the web guards that actually fail are `TestDataI18nKeysSubsetOfDictionary` and `TestI18nKeySetParity`. Also red but not named by the plan: `TestCodexAuditMCPTool` (it reads the "45 tools exposed" sentence).

**GREEN.** Production: new `internal/cli/mcp_selfreview.go` (two handlers over one core `runSelfReview`, `SelfReviewOutput`, `selfReviewMaterial`, `codexSelfReview`, `glmSelfReview`); two `add(...)` registrations in `internal/cli/mcp_server.go` (beside the audit tools, `projectRootOption()`, `WithOutputSchema[SelfReviewOutput]()`, `ReadOnlyHint(true)`); two lines in `internal/mcp/catalog.go` (`WriteCapable: false`). Parity edits: `internal/mcp/catalog_test.go` (`wantCatalogSize` 45→47, comment 25→27 read-only), `internal/cli/mcp_project_root_doc_test.go` (`docCountWords` + `Twenty-one`/`Twenty-two`, `projectRootDocSentence` `(\w+)`→`([\w-]+)`, four locale count phrases → 22), both copies of `moai-mcp-tools.md` and of the catalogue companion (every figure 45→47, family header `43 of the 47`, a new "On-demand self-review" table + paragraph, a family-table row, `Twenty-two`, `Four of the twenty-two`), the four `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md` (project_root sentence + a new "on-demand self-review" table section in each, native prose per locale), and 8 new `i18n.js` entries per `f.mcp.tools.{codex,glm}_review.enabled.{title,desc}` × 4 locales = 16 `grep -c` hits. `cmp` of each rule pair → identical; `wc -c` of the always-loaded `moai-mcp-tools.md` 4306 → 4344 (+38 bytes, under the 1,000-byte duty).

GREEN run (`m3-green2.txt`: 23 new tests + the 10 cli parity guards, `-count=1 -v`): 121 `=== RUN`, exit 0, all top-level `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 70.454s`. `go test ./internal/mcp/ -count=1 -v` → 4 PASS incl. `TestMoaiMCPTools_CatalogSize`, `ok … 0.274s`. `go test ./internal/web/... ./internal/settings/... -count=1` → `ok internal/web 53.325s`, `ok internal/settings 1.357s`, `ok internal/settings/yamlpatch 0.342s` (whole packages). Final union (`m3-final-green.txt`, run on the tree before the SplitSeq tidy below; names enumerated with `go test -list`, 109 selected): 308 `=== RUN`, 109 top-level `--- PASS`, 0 FAIL, 0 SKIP, `ok … 84.434s`. 35 of the 144 listed names were deliberately not selected (`TestFindProjectRoot*`, `TestNearestProjectRoot*`, `TestRunCC_*` and other helpers whose code this milestone does not touch, plus the live-codex gate test). `internal/template` guards that read the edited rule docs: `TestClaudeAuditTemplateSurfacesAndCatalogHash`, `TestContractMode*` (6), `TestMCPNeutralityTemplateShape`, `TestDeclaredRuleMirrorForks`, `TestRuleTemplateMirrorDrift`, `TestTemplateNeutralityAudit*`, `TestLanguageNeutrality` → all PASS, one SKIP (`TestContractModeAlwaysLoadedBudget`, skipped by the test itself; unobserved, not a pass). The C2 diff carries no `SPEC-`/`REQ-`/card/date token (`grep -n -i 't1422\|SPEC-\|REQ-\|2026-'` → no output, exit 1).

**AC matrix for the ACs M3 owns** (command = `go test ./internal/cli/ -run '^<test>$' -count=1 -v`, env-scrubbed; result = `--- PASS` in `m3-final-green.txt` / `m3-green2.txt`):

| AC | Observed by | Result |
|---|---|---|
| AC-007 | `TestSelfReview_ToolSurface` (scope enum exactly `{card, uncommitted}`, `required`, no default, `project_root`/`model` declared with `projectRootDesc`, no `focus`, output schema carries the flattened verdict fields plus the six added ones, read-only hint, catalog `WriteCapable:false`, "advisory" in the description), `TestSelfReview_AuditToolSchemasUnchanged` (4 audit tools equal the M1 snapshot by value), `TestSelfReview_AuditToolHoldersUnchanged`, `TestSelfReview_ScopeIsRequiredAndValidated` (missing/invalid/empty/`baseBranch` scope, nonexistent and non-moai root → tool error, reviewer reached 0 times) | PASS |
| AC-008 | `…CodexCardScopeRequestIsTheGateRequest` (cwd = canonical card tree, `branch` = recomputed merge base ≠ develop tip, DeepEqual to `reviewRequestParams(reviewScopeResolver(tree))`, primary tree absent from the wire), `…GLMCardMaterialIsTheCardDiff` (posted prompt == independently computed `git diff <base> -- . :(exclude)…`, F/C/D absent), `…CardBaseMovesWhenTheCardAbsorbsDevelop` (both backends move to the new base), `…ResolverIsTheOnlyDiscriminator` (sentinel MergeBase reaches both wires and both results), `…NonCardTreeReturnsInconclusiveWithoutReviewing` (develop / detached / non-git / `WT-` without base × 2 backends: `inconclusive`, cause named, reviewer 0 calls) | PASS |
| AC-009 | `…CodexUncommittedRequestShapeIsTheGatesTreeRequest` (byte-equal to the gate's own wire for a plain tree; on a `WT-` card tree still `uncommittedChanges`, no `branch`), `…GLMUncommittedMaterialIncludesStagedOnlyChanges`, `…GLMExcludedUntrackedForBothScopes`, `…GLMTruncatedBoundary` (cap−1 false, == cap false, cap+1 true), `…CodexNeverTruncatesAndListsNoUntracked` | PASS |
| AC-010 | `…AdvisoryMetadataOnEveryPath` (2 backends × pass/fail/inconclusive × 6 paths = 36 subtests: `advisory`, `scope`, `backend`, `base` == independently computed merge-base SHA / empty, `tree` == `EvalSymlinks` result ≠ the symlinked spelling passed, `truncated` bool, `excluded_untracked` array, no `gate_unmet`), `…EmptyMaterial`, `…ReviewerErrorsFailOpen` | PASS |
| AC-011 | `…ModelIsCallerInputOrBackendDefaultNeverAnAuditPin` (workflow.audit.{codex,glm} pins present: no model → codex sends none, GLM sends `glmTaskDefaultModel`; `model: X` → X on both) | PASS |
| AC-016 | `…NeverMintsReceiptsAndIgnoresTheRequiredGate` (positive control: `codex_audit` under `required` → `fail` + non-empty `gate_unmet` + a receipt; then 12 self-review calls: receipt dir identical, no `gate_unmet`, unavailable reviewer stays `inconclusive`), `…SourceReferencesNoAuditMachinery` (0 hits of `recordAuditReceipt` / `applyGateUnmet` / `auditreceipt.` / progress notifier in `mcp_selfreview.go`; controls ≥2 / ≥5 / ≥2 in `mcp_codex.go`; package-wide `recordAuditReceipt(` non-comment lines == 4) | PASS |
| AC-015 (M3 slice) | `TestMCPToolCatalogueDocsStayMirrorIdentical`, `TestMCPToolCatalogueFiguresMatchRegistry`, `TestProjectRootDocMatchesServer`, `TestDocsSiteProjectRootMatchesServer`, `TestMoaiMCPServer_RegistrationMatchesCatalog`, `TestMoaiMCPServer_AnnotationsMatchCatalog`, `TestAC_C_005_PerToolFieldsInSchema`, `TestMoaiMCPTools_CatalogSize`, whole `internal/web` (incl. `TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`, `TestI18nKeyCoverageForward`, `TestI18nUntranslatedValues`, `TestMCPConsoleRendersAllTools`), `TestSelfReview_CatalogueAndGuideCarryTheNewRows` (a table row for each tool in both catalogue copies and the four guides; both tools named in the `project_root` inventory; `Four of the twenty-two`) | PASS. The template (M4) half of AC-015 is not M3 |
| AC-012, AC-013, AC-014 | M4 | done in M4 (§E.2 M4, below) |

**Mutant probe (the criteria bite).** Applied by edit, reverted by edit; the final file was re-read for each mutation site and `gofmt -l` is clean:

| Mutant | Result | Log |
|---|---|---|
| A: call `resolveReviewScope` directly instead of the `reviewScopeResolver` seam | `TestSelfReview_ResolverIsTheOnlyDiscriminator` FAIL: codex `branch`/result `base` ≠ sentinel, GLM material is not `git diff <sentinel>`. (First attempt survived on the GLM side because c0 and c1 produced the same non-runtime diff — the fixture was strengthened with a non-runtime file on c1 and re-run) | `m3-mutant-a-second-discriminator.txt` |
| B: `runCodexAuditReviewRPC` (audit pin) in place of the pin-free driver | `…ModelIs…NeverAnAuditPin/no-model` FAIL: "codex thread/start carries model gpt-pinned-audit" | `m3-mutant-b-audit-pin.txt` |
| C: GLM call on the caller's server-bearing context | `…SendsNoProgressNotifications` FAIL: "glm_review sent 2 notification(s)" | `m3-mutant-c-heartbeat.txt` |
| D: `tree` reported as the spelling the caller passed | `…AdvisoryMetadataOnEveryPath` FAIL (30 subtests; `tree = …/lnk, want /private/var/…`) | `m3-mutant-d-tree-spelling.txt` |
| E: no empty-material short-circuit | `…EmptyMaterial` FAIL ×3 (GLM clean, codex clean, GLM untracked-only reach the backend) | `m3-mutant-e-no-empty-check.txt` |
| F: `>=` instead of `>` on the truncation cap | `…GLMTruncatedBoundary/exactly-the-cap` FAIL | `m3-mutant-f-truncation-boundary.txt` |
| G: `scope` defaults to `uncommitted` | `…ScopeIsRequiredAndValidated/{codex,glm}_review/missing` FAIL (reviewer reached) | `m3-mutant-g-scope-default.txt` |
| H: plain `git diff` (no `HEAD`) for the uncommitted scope | `…GLMUncommittedMaterialIncludesStagedOnlyChanges` FAIL | `m3-mutant-h-staged-dropped.txt` |

The plan-audit carry-over mutants: I3-3 (iii) (figures bumped without a catalogue row) is killed by `TestSelfReview_CatalogueAndGuideCarryTheNewRows` — RED before the doc edits (`m3-red2-runtime.txt`, the "no table row" lines); I3-3 (iv) / I3-7 (ii) (`project_root` through the pass-through variant) is killed by `TestSelfReview_ToolSurface` (the declared description must equal `projectRootDesc`) and by `…ProjectRootOmittedFallsBackLikeTheOtherTools` (an omitted root resolves through `CLAUDE_PROJECT_DIR`); I3-3 (ii) / I3-2 (a) (GLM, untracked-only tree) is `TestSelfReview_EmptyMaterial/glm/untracked-only`. The tests were written against the pre-implementation tree and observed red there; mutants A–H were applied afterwards to the finished implementation.

**Carry-over items — disposition in M3.**

| Item | Disposition |
|---|---|
| I3-2 (a) GLM empty material | Settled as plan-audit proposed; it departs from the literal wording of spec.md REQ-CRO-010's definition ("no untracked non-runtime files either"), so it is recorded here for the leader to read: for **GLM** a tree whose diff is empty is empty material whether or not untracked files exist, because the request cannot carry them — verdict `inconclusive`, HTTP 0, and `excluded_untracked` still names the files. For **codex** the definition is unchanged (it reads the tree itself, so untracked-only is still reviewed — `…EmptyMaterial/codex/untracked-only` asserts the reviewer is reached). |
| I3-2 (b) `base` on empty material | Empty string, as AC-010 (b) says; interpretation recorded in the code (`out.Base = ""` before the empty return) and in the test. |
| I3-3 (iii)(iv), I3-7 (i)(ii) | Done as listed above. |
| I3-6 | Not touched (spec.md body, sync-phase item). |

**Deviations and decisions to read.** (1) `SelfReviewOutput` embeds `ReviewOutput`; plan.md §B.1 said to verify the schema flattens the embedded struct. Observed: `tools/list` shows `verdict`, `summary`, `findings`, `next_steps` plus the six added fields in the output schema (asserted in `TestSelfReview_ToolSurface`), so the fields were not unrolled. (2) `glmSelfReview` calls `callGLMAudit` on a context detached from the MCP server (cancellation forwarded with `context.AfterFunc`), because `callGLMAudit` narrates through the progress notifier whenever its context carries the server and Q13 says the new tools carry no heartbeat; the alternative (a parameter on `callGLMAudit`) would have touched the audit path. (3) The timeout for both legs is `config.DefaultCodexReviewGateTimeout`. (4) The codex leg uses `runCodexReviewRPC` (the pin-free driver, plan.md §B.1 / Q6); on a missing binary the summary is `codex unavailable: codex binary not found in PATH` (the existing factory), on a non-card tree it names the cause. (5) The runtime-managed prefix list is `reviewGateRuntimePrefixes` itself (single source), not a copy. (6) Three loops in the new files use `strings.SplitSeq` (the in-IDE diagnostics' stringsseq suggestion); `go vet ./internal/cli/` → clean afterwards.

**Live probe (plan.md §G risk 5).** `codex-cli 0.160.0` is on PATH. One bounded round trip, cut at `turn/started` like `TestCodexLive_ReviewStartBaseBranchIsNotRejected`: `review/start {target:{type:"baseBranch", branch:"<40-hex merge-base SHA>"}}` on a throwaway repo (command in `.moai/reports/t1422/live-probe/2026-10-02.md`). **Observed:** accepted — no JSON-RPC error, `enteredReviewMode` echoed `changes against '595f3ac7…'`, `turn/started` arrived (`--- PASS 3.51s`; `.moai/reports/t1422/live-probe/basebranch-sha-roundtrip.ndjson`, 44 lines). **Not observed (Gap):** which working-tree files codex reads under a `baseBranch` target — the session was cut before any review text, the three marker strings appear nowhere in the transcript. The temporary probe test file was deleted; nothing was committed from it.

**Build, vet, lint, coverage, boundary, diagnostics.** Measured on the tree before the SplitSeq tidy and re-measured after it (see the end of this paragraph). `go build ./...` → `build-exit=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild-exit=0`; `go vet ./internal/cli/... ./internal/mcp/... ./internal/web/...` → `vet-exit=0`; `gofmt -l` over the changed files → no output; `golangci-lint run --timeout=5m ./internal/cli/... ./internal/mcp/... ./internal/web/...` (v2.1.6) → `0 issues.` (baseline `0 issues.`, so NEW = 0). Coverage (19 self-review tests with `-coverprofile`, then `go tool cover -func`): `handleCodexReview 100.0%`, `handleGLMReview 100.0%`, `runSelfReview 94.1%`, `selfReviewInconclusive 100.0%`, `selfReviewMaterial 84.6%`, `codexSelfReview 100.0%`, `glmSelfReview 100.0%`; profile statement count for the file → `statements=70 covered=66 pct=94.3` (target 85). Boundary: `grep -rn 'AskUserQuestion' internal/cli/mcp_selfreview.go internal/mcp/catalog.go` → no output (exit 1); `git diff -U0 | grep '^+' | grep -c AskUserQuestion` → `0`. Size: `internal/cli/mcp_selfreview.go` 221 lines, 129 non-comment non-blank; the registration block adds 18 code lines to `mcp_server.go`, the catalog 2 → ≈149 production lines against plan.md §H's ≈165 for the tool part. Files: 3 production (1 new), 4 test (2 new: 880 + 352 lines, heavier than the plan's whole-SPEC ≈350–420 estimate because AC-010 multiplies 2 backends × 3 verdicts × 6 paths), 4 rule docs, 4 guides, 1 `i18n.js`, plus this file. Diagnostics raised after the run was interrupted: (a) `readFile redeclared` in `tool_policy_test.go` — `internal/cli/tool_policy_test.go:276 func readFile(name string)` is pre-existing (`git diff 7919da731 --stat -- internal/cli/tool_policy_test.go` → empty, exit 0) and none of the new files declares it (`grep -n 'readFile\b' internal/cli/mcp_selfreview*.go` → no output); an intermediate draft of the fixture helper had that name and was renamed (`srReadFile`) before the first compile, so the report reads a stale state. (b) The medium path-traversal hint on lines near 111/145/146/850/851/854 of `mcp_selfreview_test.go`: those are the constant relative literals `"../../.moai/reports/…"`, `"../../.claude/agents/moai"`, `"../../internal/template/…"`, `"../../docs-site/content/"+loc+…` (loc ∈ ko/en/ja/zh) — repository-relative test inputs, the same shape as `projectRootDocFiles` in `mcp_project_root_doc_test.go`; nothing user-controlled reaches them, so no change.

**Gaps (M3).** (1) No whole-package `./internal/cli` run (AGENTS.md §4, gitflow-lane-protocol §8): the selected names plus the web / mcp / settings whole packages are the measurement; the other `internal/cli` tests and the `internal/template` / `internal/hook` / `internal/config` whole packages were not run (only the named template guards above). (2) `TestContractModeAlwaysLoadedBudget` SKIPPED on its own condition. (3) No Hugo build was run (docs-site parity is evidenced by the Go tests, which read the markdown, not by a rendered site). (4) Exit codes were taken from the tool result and the `ok`/`FAIL` summary lines or from `echo "exit=$?"` after a redirected run. (5) The running MCP server of this session is the older `v3.2.0-rc.23` build and does not list the new tools (plan.md §G risk 7); everything here ran against binaries built from this tree. (6) The live probe did not observe which files codex reads. (7) Nothing was pushed.

**Residual risk (M3).** `scope: uncommitted` in the primary checkout reviews the shared working tree, which may hold another session's work (named in the tool description and the guides; no path restriction by design). Both tools are synchronous: a host tool timeout shorter than the 900 s review budget loses the call with no retry and no record (plan.md §G risk 6; there is no GLM fallback). A server process started before this milestone does not list the new tools. `excluded_untracked` lists every untracked non-runtime path with no cap, so a tree with thousands of untracked files returns a long array. Absorbing `develop` into a card moves the merge base, so two calls can review different slices — by design, covered by `…CardBaseMovesWhenTheCardAbsorbsDevelop`.

### M4 — agent tool grants, the card-review doctrine, the template comment (REQ-CRO-011..013, REQ-CRO-014 M4 slice)

Run by a fourth manager-develop worker (`cycle_type=tdd`) on the same worktree. Numbering note: this run was dispatched with "AC-012, AC-013, AC-014, AC-015 template half"; acceptance.md numbers the tool holders AC-012, the stage order AC-013, the leader-side rules AC-014, and the template/inventory single-live-key clause AC-015 (iii)(iv). The matrix below follows acceptance.md. All tests ran as one env-scrubbed compound call (`unset MOAI_KANBAN … MOAI_PROFILE_LEASE_TOKEN && go test … -count=1 -v`); raw logs under `.moai/reports/t1422/red/m4-*.txt` (gitignored).

**Pre-flight (pre-change tree).** `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1422`; `git rev-parse HEAD` → `5c006bf9b2f8b8de5fced834041fcb3dc9c6543e`, branch `WT-codex-review-lane-scope`, `git status --short` clean. `go build ./...` → `build-exit=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild-exit=0`. Baseline, green before any change: the M1 holder/mirror guards `TestSelfReview_AuditToolHoldersUnchanged`, `TestSelfReview_AuditToolSchemasUnchanged`, `TestMCPToolCatalogueDocsStayMirrorIdentical`, `TestMCPToolCatalogueFiguresMatchRegistry` (4 `=== RUN`, 4 PASS, `ok … internal/cli 0.980s`, `m4-cli-baseline.txt`); `TestAutoRankDoctrineAmendment` and `TestAutoRankAgentDoctrine` (they read `kanban-dispatch.md`, PASS); `TestGoldenCommittedArtifactsMatchEmission` PASS; 32 `internal/template` tests that read the files M4 touches (agent frontmatter/catalog/structure, rule and workflow mirror, neutrality, leak, provenance, contract-mode, review-gate template): 228 `=== RUN`, 29 top-level PASS, 3 SKIP (`TestContractModeInheritedDivergence`, `TestContractModeAlwaysLoadedBudget`, `TestContractModeEmitterSites` skip on their own conditions — unobserved, not passes), 0 FAIL, `ok … internal/template 4.180s` (`m4-template-baseline.txt`).

**RED.** New file `internal/template/codex_review_ownership_m4_test.go` (7 tests, 28 `=== RUN` including subtests). Every checker takes document text and returns problems, so one function judges the real files and a table of deliberately broken variants. Pre-implementation tree = HEAD `5c006bf9b` + the uncommitted test file (sha256 `3623b766461a52e627d4238f62335cfcbab33def979b64dc9a49681ee61c7af6` at the recorded run). Recorded run `m4-red2.txt`, classification `m4-red-semantic.txt`: exit 1, **EXPECTED_RED (4)**: `TestReviewOwnership_ToolHolders` (`mcp__moai__codex_review holders = [], want exactly [manager-develop manager-docs manager-lead]`, ×4 for both tools and both trees), `TestReviewOwnership_CardReviewDoctrine` (`the lane stage section does not carry card-review`, `section ## The card-review stage is missing`, local and mirror), `TestReviewOwnership_DoctrineMutants` (11 subtests: `mutant did not apply — the text it edits is absent`, the doctrine they mutate does not exist yet), `TestReviewOwnership_TemplateWorkflowKeyIsCommentOnly` (`no commented example of tree_scope in the template`; its inventory-hash and no-`tree_scope`-elsewhere assertions already passed). **PASS in RED (3, regression/preserve and checker-works lines)**: `TestReviewOwnership_ToolHoldersCodexEmission`, `TestReviewOwnership_ToolHolderMutants` (6 subtests), `TestReviewOwnership_TemplateWorkflowMutants` (4 subtests). TOOL_FAILURE: none. Two draft defects were fixed before GREEN and are disclosed: the first draft's neutrality assertion tripped on an older card-number token already inside the stage section's pre-existing paragraph (`m4-red1.txt`, superseded — the assertion now judges only paragraphs that talk about card-review), and after the first GREEN run two mutant cases carried a stale anchor string / expected-message and were corrected (`m4-green1.txt` → `m4-green2.txt`); a later lint fix replaced one `switch` with an `if` (`m4-green3.txt`).

**GREEN — the changes.** (1) `tools:` lines, appended names only (`mcp__moai__codex_review, mcp__moai__glm_review`), line numbers unchanged: C1 `manager-develop.md:9`, `manager-docs.md:9`, `manager-lead.md:10`; C2 `manager-develop.md:10`, `manager-docs.md:9`, `manager-lead.md:10`. For card t1424's later edit of the same `manager-develop.md` line (C1 `:9`, C2 `:10`): this change appends at the end of that line, so a later merge conflicts on that one line only. (2) `kanban-dispatch.md` (C1 and C2): one `[HARD]` paragraph in `### The lane's task list carries the card's stages` (the `card-review` stage between run-exit verification and integration, the evidence path, advisory, and the conditional leader sentence "With `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools.") and one sentence in `## Completion is read, never trusted` (the evidence list carries `card-review.md`; a card whose progress record neither cites a readable `card-review.md` nor records a reason for its absence is a gap and stays in its column). (3) `kanban-dispatch-detail.md` (C1 and C2, byte-identical): new `## The card-review stage` with the ordered list `[run-exit]` → `[card-review]` → `[integration]` → `[report]`, the call, the fallback, the evidence fields, the re-review ceiling (at most 2, the same number as the run-exit verify gate's "at most two verify re-entries" in `workflows/run/mode-orchestration.md` § Verify Exit Gate), advisory, the continued-firing answer, one-review-per-diff, and the leader's `scope: uncommitted` use. (4) Template `internal/template/templates/.moai/config/sections/workflow.yaml`: five comment lines and one commented example `# tree_scope: review  # review | skip` under `codex.review_gate`; this repository's tracked `.moai/config/sections/workflow.yaml` and the primary checkout's file are untouched.

**Regeneration.** `make agents-emit` → exit 0, `git status --short` afterwards lists no `.codex/agents/moai/*.toml` change: the emitter reads each definition's body and maps only the `mcp__moai__*` family to one `[mcp_servers.moai]` table, so the two added tool names change no emitted byte (the three TOMLs already carry that table). `make agents-emit-check` → exit 0 (`m4-agents-emit-check.txt`). `make build` → exit 0 (`agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`, `templ-generate`, `gen-catalog-hashes --all`, `go build` into the ignored `bin/moai`; not installed anywhere): `internal/template/catalog.yaml` changed in exactly three hash lines (the three edited agents), `git diff --stat`: `3 insertions(+), 3 deletions(-)`.

**Holder sets (printed, this tree).** `grep -l mcp__moai__codex_review` and `grep -l mcp__moai__glm_review` over `.claude/agents/moai/*.md` (C1) and `internal/template/templates/.claude/agents/moai/*.md` (C2) → `manager-develop`, `manager-docs`, `manager-lead` for each tool in each tree; `grep -l mcp__moai__codex_audit` / `mcp__moai__glm_audit` over both trees → `plan-auditor`, `sync-auditor` only (P1, unchanged). C3 `internal/template/templates/.codex/agents/moai/*.toml`: the review tools appear in no emitted definition (`grep -l` exit 1; the emitter carries no tool names), the audit tools appear in `plan-auditor.toml` and `sync-auditor.toml` only (unchanged).

**Inventory and the other single-live-key surfaces.** `shasum -a 256 internal/config/testdata/shipped_key_inventory.yaml` → `a3d2c397827130d71047fad9aa59c6b690d64394b222dca56adc44db149fc409`, identical to the M1 value (P5) recorded above; `grep -c tree_scope` on the inventory → `0`, on `internal/settings/schema_sections.go`, `internal/web/fieldsets.templ`, `internal/web/assets/i18n.js` → `0` each. `git diff -U0 HEAD -- internal/template/templates/.moai/config/sections/workflow.yaml`: two hunks, 6 added lines, `grep -vc '^+ *#'` over the added lines → `0` non-comment additions, 0 removed lines (`m4-workflow-yaml-diff.txt`); `grep -c tree_scope` on the template → `2` (L8 was `0`).

**GREEN run.** The 7 new tests: `m4-green3.txt` (final): 28 `=== RUN`, 7 top-level PASS, 0 FAIL, `ok … internal/template 0.471s`. Whole `./internal/template/...` under the `go-test-template` slot lease (`moai slot status` → free; acquired, released): `ok internal/template 205.830s`, `ok …/agentemit 0.624s`, `ok …/commandemit 0.114s` (`m4-template-pkg.txt`). `internal/cli` guards after the change (19 exact names, 119 `=== RUN`, 19 top-level PASS, 0 FAIL/SKIP, `m4-cli-after.txt`): `TestSelfReview_AuditToolHoldersUnchanged`, `…AuditToolSchemasUnchanged`, `…CatalogueAndGuideCarryTheNewRows`, `TestMCPToolCatalogueDocsStayMirrorIdentical`, `…FiguresMatchRegistry`, `TestAutoRankDoctrineAmendment`, `TestAutoRankAgentDoctrine`, the two `TestReviewGateReaders_*`, the two `TestTreeScopeReader_*`, the init/workflow-yaml readers and the managed-redeploy pair. `go test ./internal/config/... ./internal/mcp/ ./internal/settings/... -count=1` → all `ok` (config 15.628s, mcp 0.288s, settings 1.852s). Build/vet/lint: `go build ./...` `build-exit=0`; `GOOS=windows GOARCH=amd64 go build ./...` `winbuild-exit=0`; `go vet ./internal/template/` `vet-exit=0`; `gofmt -l` clean; `golangci-lint run --timeout=5m ./internal/template/...` (v2.1.6) → `0 issues.` after one fix (the first run reported one `QF1002` in the new test; baseline for the package was clean). Coverage: no non-test Go file was added or changed — N/A. Boundary: `git diff -U0 HEAD | grep '^+' | grep -c AskUserQuestion` over the tracked changes → `0`, and the new test file → `0`.

**AC matrix (ACs M4 owns).**

| AC | Observed by | Result |
|---|---|---|
| AC-012 | `TestReviewOwnership_ToolHolders` (a: holders of the two review tools = the three, of the two audit tools = the two auditors, in C1 and C2; b: per-agent C1 tool set == C2 tool set), `…ToolHoldersCodexEmission` (C3), `make agents-emit-check` exit 0 (c); mutants (grant to a fourth agent, review tool to an auditor, audit tool to a lane agent, grant withheld, one-sided copy edit) all flagged by `…ToolHolderMutants` | PASS |
| AC-013 | `TestReviewOwnership_CardReviewDoctrine` (stub names `card-review`, the evidence path, "advisory"; detail ordered list `run-exit>card-review>integration>report`; no negating phrase; ceiling "at most 2"; C2 text equals C1 text for the card-review sections and carries no SPEC id/card number/date token; detail files byte-identical); mutants (stage after integration, after the report, removed, negating sentence, ceiling dropped, path dropped) flagged by `…DoctrineMutants` | PASS |
| AC-014 | same test: the completion-read section lists the evidence path and carries the gap sentence; the stage section carries one conditional sentence with `tree_scope: skip` before `turn-end codex review gate` and "directly with the same tools"; mutants (unconditional sentence, condition after the gate, gap rule negated, gap rule moved out of the section, evidence path missing from the stage section) flagged | PASS |
| AC-015 (iii)(iv) (template half) | `TestReviewOwnership_TemplateWorkflowKeyIsCommentOnly` (parsed `workflow.codex.review_gate` keys == `[enabled]`, value false; every `tree_scope` line is a comment, ≥1; inventory sha256 equal; no `tree_scope` in inventory/schema/console/i18n) + the `git diff` evidence above; mutants (live key, live key beside a comment, example removed, `enabled` flipped) flagged by `…TemplateWorkflowMutants` | PASS |

**Always-loaded growth statement (rule-authoring.md duty).** `.claude/rules/moai/workflow/kanban-dispatch.md` is always-loaded: `wc -c` 26807 → 27672 bytes (+865) locally, 26485 → 27350 (+865) in the distributed mirror — under the 1,000-byte threshold, stated anyway. A session that never dispatches a card (anything outside Kanban Mode) pays those 865 bytes on every turn and again after every `/clear`; the content cannot live under a `paths:` scope because the leader and lane sessions learn their role from the session context rather than from a file path (the file's own loading note), and the detail (stage list, call, fallback, ceiling, continued-firing answer) went to `kanban-dispatch-detail.md`, which is path-scoped and costs the always-loaded surface nothing.

**Gaps (M4).** (1) `internal/cli`, `internal/web` and `internal/hook` were not run as whole packages (AGENTS.md §4): `internal/cli` ran the 19 exact names above (the tests that read the three kinds of file M4 changed); `internal/web`/`internal/hook` tests do not read any M4-changed file. (2) The mutant tables are checker-level probes against derived text; no mutated file was committed or run through the full suite. (3) `go test` exit codes were taken from `echo "exit=$?"` after a redirected run (single-command redirect form) or the tool result. (4) The running MCP server of the session is the older build (plan.md §G risk 7) and does not list the new tools; nothing in M4 calls them — the doctrine describes the fallback for that state. (5) `TestContractModeAlwaysLoadedBudget` and two sibling tests SKIP on their own conditions in the template package (unobserved, as at baseline). (6) Nothing was pushed; `bin/moai` was built into the ignored `bin/` and installed nowhere. (7) The pre-existing paragraph in `### The lane's task list carries the card's stages` names a card number in the distributed mirror; it predates M4 and was left alone (scope).

**Residual risk (M4).** The doctrine is a lane duty without a hook: a lane that stops running the stage is caught only if the leader's completion read lists `card-review.md` — by design (no hook is out of scope) and covered by the gap sentence, but it depends on the leader reading. `tree_scope` appears in the template only as a comment, so a user who wants `skip` must uncomment it by hand (and the key is parsed but not offered by the console). The `card-review` stage text assumes the new MCP tools exist in the running server; the documented fallback covers only the codex leg.

## §E.3 Run-phase Audit-Ready Signal

**Scope of this signal: M1 through M5 — the run phase is complete.** M1-M4 are the implementation milestones below; M5 (CHANGELOG, optional docs, hand-off) closed in the single sync commit, whose evidence is §E.4. This paragraph was updated by the sync-phase worker at the leader's dispatch (the earlier text said a later run-phase worker would replace it when M5 closed; M5 is documentation and hand-off only, so no further run-phase code worker exists). The claims below are unchanged from the M4 run. What was ready to be audited at M4 and remains the audited subject:

- M4 (a later commit on the same branch, parent = M3's commit, not pushed): AC-012, AC-013, AC-014 and the template half of AC-015 PASS on the evidence in §E.2 M4 (7 new tests green, whole `internal/template/...` green, 19 `internal/cli` guards green, lint 0 issues, `make build` exit 0).
- `run_milestones_done: M1, M2, M3, M4, M5` (M5 = CHANGELOG + hand-off, closed in the sync commit) · cycle: tdd (RED → GREEN → REFACTOR; the REFACTOR step found nothing to simplify — one policy file of 53 code lines, a one-line wiring on each path, and the single config normaliser; no second discriminator, no new dependency).
- Commits on `WT-codex-review-lane-scope` (parent `984d64957`): M1 `fdb4932eb` (a deliberately RED commit — tests, declaration stub, config field, SPEC `status: in-progress`); M2 = the commit that follows it in `git log` (policy file, the two wirings, config test, this evidence). Not pushed.
- M3 (a later commit on the same branch, parent = M2's commit, not pushed): AC-007, AC-008, AC-009, AC-010, AC-011, AC-016 and the M3 slice of AC-015 PASS on the evidence in §E.2 M3 (109-test union run: 0 FAIL, 0 SKIP; whole `internal/web`, `internal/mcp`, `internal/settings` packages green; eight mutants killed; coverage of the new file 94.3%). AC-012, AC-013, AC-014 and the template half of AC-015 were done in M4 (first bullet above).
- ACs owned by M1/M2 — AC-001, AC-002, AC-003, AC-004, AC-005, AC-006: PASS on the evidence in §E.2 (final union run: 64 top-level PASS, no FAIL/SKIP; three mutants killed). AC-004 (b) became a regression line only after its own pre-change PASS was observed (M1).
- Carry-over items from plan-audit round 3: I3-1 done (M2-relevant env rows, helper-outside-policy mutant, static guard); I3-5 recorded in the orphan-worktree test comment; I3-2, I3-3 (ii)–(iv), I3-7 done in M3 (§E.2 M3; the GLM untracked-only empty-material decision departs from the literal REQ-CRO-010 wording and is recorded there for the leader to read); I3-6 is a sync-phase spec.md item; I3-4 needed nothing.
- The audit-tool schema snapshot for AC-007 (f) exists: `.moai/reports/t1422/red/audit-tools-schema-pre.json`, sha256 `6910a4678ee5c82966a8e12af5e6762e216d109577a606b9ad0165b76b134ce6` (local evidence, gitignored; the hash is the citable part).
- Hand-off items from §G of plan.md are unchanged and still the leader's: `tree_scope: skip` in the primary's `workflow.yaml` after landing (only needed if the gate is re-enabled there), t1404 residual edit, card t1426, and the sibling SPEC.
- Open for the leader to read: the one deviation from plan.md §B.2's helper signature (`treeScopeSkipApplies(scope, func() string) bool`, recorded in §E.2 M2) and the size of the policy tests (733 lines in `codex_review_ownership_test.go`; the env matrix alone took 9.60 s in the final GREEN run).

## §E.4 Sync-phase Audit-Ready Signal

Sync base `f586579c3` (M4 commit) on branch `WT-codex-review-lane-scope`, card worktree `.moai/worktrees/t1422`, nothing pushed. Sync scope (manager-docs role): the `CHANGELOG.md` `[Unreleased]` `### Added` entry (first bullet, newest-first), this signal, the M5 hand-off list, the §E.3 closing sentence, and the `spec.md` frontmatter `status` and `updated` fields on the single sync commit. No Go, agent, rule or template file is touched by this commit.

```yaml
sync_status: audit-ready
sync_complete_at: 2026-10-02
card: t1422
sync_commit_sha: 630888ad74208d13d94bf556d1446ae36c471bc5
changelog_path: CHANGELOG.md
changelog_entry_position: "[Unreleased] > ### Added, first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged, single sync commit); updated: 2026-10-02"
  plan_md_acceptance_md: "stateless on the status axis per spec-frontmatter-schema.md (no status field to transition)"
  progress_md: "no frontmatter; this section is the sync signal"
b12_self_test_a: "PASS — `grep -c 'SPEC-CODEX-REVIEW-OWNERSHIP-001' CHANGELOG.md` = 0 before the entry landed; = 1 after"
b12_self_test_b: "PASS — ac_source=.moai/specs/SPEC-CODEX-REVIEW-OWNERSHIP-001/acceptance.md (tier M, resolved, non-empty). The MOAI-AC-COUNTER (default AC prefix) printed 16 on stdout and `live=16 excluded=0 ambiguous=0` on stderr, exit 0; `grep -c '^### AC-' acceptance.md` = 16; the CHANGELOG entry states 16 acceptance criteria AC-001..016"
b12_self_test_c: "PASS — every path named in the CHANGELOG entry was checked present with ls (the two new Go files and their tests, the changed Go files, the catalogue file, i18n.js, the four docs-site guides, the three agent definitions and mirrors, the two rule-file pairs and mirrors, the template workflow.yaml and catalog.yaml, the new template test); the diffstat 984d64957..f586579c3 lists defaults.go, types.go, catalog.yaml, i18n.js, codex_review_gate.go, codex_stop_chain.go and mcp_selfreview_fixture_test.go"
public_docs_judgment: "docs-site: the four guides/mcp-server.md pages were updated in M3 (tool table, project_root inventory) and no page in README*.md or docs-site/content carries a stale 45-tool count (`grep -rn -E '\\b45 (MCP )?tools|45-tool|forty-five'` over README*.md and docs-site/content: no match). No further page is owed: tree_scope is a commented template example, not a console field, and the card-review stage is lane doctrine."
m5_optional_items:
  readme_x4: "skipped — the README files name codex_audit but carry no tool-count or review-tool inventory to correct; a new line would need the four-locale heading/emoji/parity checks for no stale claim removed"
  multi_model_audit_x4: "skipped — advanced/multi-model-audit pages describe audit convergence, which this SPEC did not change; adding a self-review paragraph would touch four locales and their parity guards for a non-audit surface"
  moai_ref_cross_model_audit_skill_one_liner: "skipped — touching a skill body changes internal/template/catalog.yaml skill hashes and the skill mirror pair, and card t1423 edits the auditor/cross-model documents; recorded, not dropped"
mx_validation: "read-only; no tag owed in sync. The run phase already placed @MX:NOTE on runSelfReview, glmSelfReview and treeScopeSkipApplies, and @MX:SPEC on the two new Go files; this commit edits no Go file"
read_during_sync: "spec.md, plan.md (§F, §G, §I M5, §J), acceptance.md (AC-014..016, DoD), progress.md §E.1-§E.3 and the Decision Log, manager-docs.md §B12, internal/cli/mcp_selfreview.go, internal/cli/codex_review_tree_scope.go, the template workflow.yaml diff, the mcp_server.go and catalog.go diffs, CHANGELOG.md [Unreleased] head"
not_run_in_sync: "no whole-package ./internal/cli or ./internal/template run (AGENTS.md section 4; moai slot status showed a held lease); no CI; no sync-audit; no make build"
```

### Sync verification batch (this run, this tree, HEAD `f586579c3` plus the uncommitted CHANGELOG.md and progress.md edits)

Every Go test ran as one env-scrubbed compound call (`unset MOAI_KANBAN … MOAI_PROFILE_LEASE_TOKEN && go test … -count=1 -v`); logs under `.moai/reports/t1422/sync/` (gitignored, local evidence — the lines below are the citable part).

| Check | Command (abbreviated) | Observed |
|---|---|---|
| SPEC lint | `moai spec lint SPEC-CODEX-REVIEW-OWNERSHIP-001` | exit 0, `✓ No findings — all SPEC documents are valid` |
| `internal/cli` guard set | `go test ./internal/cli/ -run '^(<73 exact names>)$' -count=1 -v` — names from `go test ./internal/cli/ -list` filtered to the TreeScope, CodexReviewGate, CodexStopChain_TreeScopeSkip, CodexReviewScope, SelfReview, MCPToolCatalogue, ProjectRootDoc, DocsSiteProjectRoot, ReviewGateReaders, required-gate and registration/annotation guards | exit 0; 73 names, 73 top-level `--- PASS`, 0 FAIL/SKIP, 259 `=== RUN` (subtests included); `ok github.com/modu-ai/moai-adk/internal/cli 179.401s` |
| `internal/template` M4 set | `go test ./internal/template/ -run '^(TestReviewOwnership_…7 names)$' -count=1 -v` | exit 0; 7 listed, 7 top-level `--- PASS`, 0 FAIL/SKIP, 28 `=== RUN`; `ok …/internal/template 0.296s` |
| `internal/mcp`, `internal/web` | `go test ./internal/mcp/ ./internal/web/ -count=1` (whole packages) | exit 0; `ok …/internal/mcp 0.275s`, `ok …/internal/web 75.289s` |
| `internal/config` | `go test ./internal/config/ -run '^(TestNormalizeCodexReviewGateTreeScope|TestDefaultConfig_CodexReviewGateTreeScopeIsReview|TestTemplateWorkflowYAML_JevShipsOff|TestShippedConfigKeysHaveReaders)$' -count=1 -v` | exit 0; 4 names, 4 top-level `--- PASS`, 8 `=== RUN`; `ok …/internal/config 5.295s` |
| Build | `go build ./...` ; `GOOS=windows GOARCH=amd64 go build ./...` | `build-exit=0` ; `winbuild-exit=0` |
| Lint | `golangci-lint run --timeout=5m ./internal/cli/... ./internal/template/... ./internal/mcp/... ./internal/web/... ./internal/config/...` (golangci-lint v2.1.6) | `lint-exit=0`, `0 issues.` |

### Hand-off list (the leader's, outside this SPEC's code — none of these is a satisfied requirement)

1. **`tree_scope: skip` in the primary checkout.** After landing, the leader decides whether the primary `workflow.yaml` (tracked file, local modification) needs `workflow.codex.review_gate.tree_scope: skip`. As of the 2026-10-02 16:39 KST observation that file read `codex.review_gate.enabled: false` (an operator-approved temporary measure, modified 15:32), so the gate was already off and `skip` is needed only if the gate is turned back on there. It is a moving operator state: re-read the file at landing. Persistence across a release sync is unresolved (plan.md O-3; first release sync is when to check).
2. **t1404** — edit it down to its residual items (T2, T4, T5, T8) after this SPEC lands; do not close it beforehand (plan.md §D, Q4).
3. **Card t1426** — the audit tools' `baseBranch` resolution (`resolveReviewBaseBranchName`, `resolveReviewMergeBase`); this SPEC leaves it unchanged by design. Issuing the card is the leader's.
4. **Sibling SPEC-CODEX-REVIEW-ASYNC-001** — lands after this one; it needs one amendment and a final re-audit first (see its own progress.md). Its insertion points in `HandleCodexReviewGate` do not overlap this SPEC's (spec.md §E).
5. **Decision to read — GLM empty material (M3, I3-2 (a)).** A `glm_review` call on a tree whose diff is empty returns `inconclusive` with zero HTTP calls even when untracked files exist (they are named in `excluded_untracked`). That departs from the literal wording of spec.md REQ-CRO-010's empty-material definition ("no untracked non-runtime files either"); Q14 delegated the case to the first run milestone. codex keeps the literal definition (untracked-only is still reviewed). Recorded in §E.2 M3; reverting is one condition in `runSelfReview`.
6. **Merge-conflict risk on card t1424.** t1424 edits the `tools:` line of `manager-develop.md`. M4 only appended `mcp__moai__codex_review, mcp__moai__glm_review` at the end of that line without moving any line number: local (C1) `.claude/agents/moai/manager-develop.md:9`, distributed mirror (C2) `internal/template/templates/.claude/agents/moai/manager-develop.md:10` (manager-docs C1 `:9`/C2 `:9`, manager-lead C1 `:10`/C2 `:10`; line numbers per §E.2 M4). Whichever card lands second merges that one line in both copies; the emitted `.codex/agents/moai/*.toml` files did not change, and `internal/template/catalog.yaml` agent-hash lines will need regenerating (`make build`).
7. **Risk with card t1399.** `reviewGateEnvContext` in `internal/cli/codex_review_scope.go` references `config.EnvMoaiFactoryWorker`; if t1399 deletes that constant, the file stops compiling (not a change of this SPEC). `TestTreeScopePolicy_EnvMatrix` derives its key list from `internal/config/envkeys.go`, so a deleted key moves the matrix rather than failing it silently.
8. **The session MCP server runs an older build.** It reports `v3.2.0-rc.23 (commit d194083fb)` and does not list `codex_review` / `glm_review`. After the leader rebuilds and installs the rc, the MCP server must be reconnected before a lane can call the new tools; until then the card-review stage's codex leg uses `moai verify codex-review --project-root <tree>` and the GLM leg is unavailable.

### Gaps

- No whole-package `./internal/cli` or `./internal/template` run in this sync (`moai slot status` showed a held lease; AGENTS.md section 4); the 73 + 7 named tests are the measurement, and the leader's batch CI covers the rest. `internal/hook` was not run (no file there changed in this SPEC).
- Which working-tree files codex reads for a `baseBranch` target is unobserved (the live probe was cut at `turn/started`); the CHANGELOG makes no claim about it.
- Hugo build of the docs-site was not run; docs-site parity rests on the Go guards that read the markdown.
- The `go test` exit codes were taken from the tool result and from `echo "exit=$?"` after redirected runs; the per-run logs are local scratch.
- `moai spec audit` result: see the Close-out section, appended by the backfill commit.

### Residual risk

- The card-review stage has no hook: a lane that stops running it is caught only if the leader's completion read lists `card-review.md` (by design; the gap sentence covers it).
- Both tools are synchronous; a host tool timeout shorter than the 900 s review budget loses a call, and `glm_review` has no fallback.
- Under `tree_scope: skip`, a card worktree on a detached HEAD resolves as having no `WT-` branch and is not reviewed by the turn-end gate (plan.md §G risk 11).
- `scope: uncommitted` in the primary checkout reviews the shared working tree, which may hold another session's work.
- The persistence of a primary-checkout `tree_scope: skip` across a release sync is unresolved (hand-off item 1).

### Close-out

Recorded by the backfill commit that replaces the `pending-backfill` placeholder with the sync commit's full SHA, `630888ad74208d13d94bf556d1446ae36c471bc5` (subject `docs(SPEC-CODEX-REVIEW-OWNERSHIP-001): sync-phase artifacts`, parent `f586579c3`).

- `moai spec lint SPEC-CODEX-REVIEW-OWNERSHIP-001` on the tree with the backfilled SHA: `✓ No findings — all SPEC documents are valid`.
- `moai spec audit --filter-spec SPEC-CODEX-REVIEW-OWNERSHIP-001 --json`: `total_specs 1`, `grandfathered 0`, `modern_era_clean 1`, one finding only — `EraAutoDetected`, severity `INFO`, era `V3R6`, heuristic `H-4 (§E.2 + §E.4 + sync_commit_sha)`; no drift finding.
- Tool provenance (verification-claim-integrity §2.2): the installed `moai` reports `v3.2.0-rc.25`, commit `802a72235`, built 2026-10-02T08:00:14Z, and `git merge-base --is-ancestor 802a72235 HEAD` exits 1 (not an ancestor of this tree's HEAD), so the same lint and audit were repeated with a binary built from this tree (`go build ./cmd/moai`, no version ldflags, so it reports `v3.1.3 none`): same results. Neither build was compared against the other beyond those outputs.
- A sync-audit has not been run; the leader owns that dispatch.

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
