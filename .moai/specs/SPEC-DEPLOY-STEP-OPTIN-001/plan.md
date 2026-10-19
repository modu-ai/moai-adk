# SPEC-DEPLOY-STEP-OPTIN-001 — 구현 계획

## §A 맥락

- **Tier L 판정(측정).** 변경 표면이 여섯 서브시스템에 걸친다: 설정 해석(`internal/config`), 카드 전달·배차(`internal/cli/factory*.go`), 세션 시작 훅(`internal/hook`), 리뷰 게이트(`internal/cli/codex_review_gate.go`), 웹 설정(`internal/settings`, `internal/web`), update 병합 시험. §C 의 파일 지도는 신규 약 16개와 수정 약 12개로 28개 안팎이다(Tier L 기준 15개 초과). 요구 18개, 인수 기준 22개(각각 상한 25). 새 설정 표면과 외부 변경 게이트라는 헌법적 성격도 있다. 리서치(research.md)와 설계(design.md)가 필요하므로 다섯 산출물을 둔다. 여기에 progress.md 와 decision-index.md 를 더한다(decision_gate 가 `on`).
- **현재 상태.** plan 단계 산출물만 있다. 구현 코드는 건드리지 않았다. 제품 수준 결정 행(Q1–Q9, Q14)과 EVIDENCE-NEEDED 행(Q11)은 열려 있다. 입력 메모는 가설 원천으로만 썼고 모든 전제를 이 트리(`8249bff8c3`)에서 다시 측정했다(research.md).
- **이 계획이 하지 않는 것.** 구현, 템플릿 변경, 이 저장소 자신의 로컬 설정 변경은 run 단계의 몫이다. t1882·t1752·t1644·t1900 의 범위는 건드리지 않는다.

## §B 되돌리기 어려운 결정 (구현 전 확정)

가장 바뀌기 쉬운 결정부터 적는다. 각 항목은 decision-index.md 의 같은 번호와 짝이다. 모두 열려 있다(Q10·Q12·Q13 만 규칙 적용으로 채움).

1. **Q1·Q2 키 이름·위치·값 형태.** 사용자가 쓰고 웹이 편집하는 공개 계약이다. 정해진 뒤에는 이름을 바꾸는 일이 호환 파괴다. 설계는 작업 이름으로 진행하고 확정되면 이름만 바꾼다.
2. **Q5 잘못된 값의 영향 범위.** fail-closed 판정의 폭이다. 시험 기대값(AC-DSO-002)이 이 결정에 묶인다.
3. **Q3 review_repair 기본값.** 템플릿에는 값이 없으므로(설계 §6) 기본값은 코드 상수다. (b) 를 고르면 해석기가 다른 키(`codex.review_gate.enabled`)를 읽게 되어 출처 `derived` 가 생긴다.
4. **Q6·Q7 이웃 키와의 관계.** `project.continuation`, team 프로필 `auto_push` 의 처리는 템플릿 기본값 변경 여부를 가른다.
5. **Q14 dispatch 범위.** 자가 배차만 막는지 지명까지 막는지.
6. **Q9 즉시 중지의 의미**, **Q4 release/deploy 첫 조각 포함**, **Q8 이 저장소 자신의 설정 이전**, **Q11 t1882 와의 관계.**

## §C 기술 접근 — 파일 지도

줄 번호는 기준 트리(`8249bff8c3`)의 값이다. 파일 목록은 이 §C 가 정한다. run 단계가 목록 밖 파일을 바꿔야 하면 멈추고 보고한다(마일스톤 종료 시 트리 변화 점검). 목록이 Tier L 의 한도를 넘으면 Tier 판정을 다시 한다. `progress.md` 는 모든 마일스톤의 공통 승인 파일이다.

| 파일 | 역할 / 변경 | 마일스톤 |
|---|---|---|
| `internal/config/stage_policy.go` (신규) | 해석기: 일곱 단계, 판정 표(설계 §4), 출처, revision, 기본값 상수 | M2 |
| `internal/config/stage_policy_test.go` (신규) | AC-DSO-001–005, 015 의 시험 | M2, M3 |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | 단계 블록을 **주석으로만** 추가(값 0개, L48) | M2 |
| `.moai/config/sections/workflow.yaml` | 같은 주석 블록. 저장소 자신의 값 이전(Q8)은 M6 의 별도 커밋 | M2, M6 |
| `internal/template/stage_policy_template_test.go` (신규) | AC-DSO-015 (a) | M2 |
| `internal/cli/doctor_stages.go` + 시험 (신규, Q12) | 산문 소비자용 읽기 동사 | M2 |
| `internal/cli/factory_card_pr.go` | 전달 본체 네 지점에 단계 판정(L03, L54). t1882 착지 뒤 기준 | M3 |
| `internal/cli/factory_card_pr_stage_test.go` (신규) | AC-DSO-006–008, 012, 013 | M3 |
| `internal/cli/factory_card_pr_test.go` | 기존 시험 설비에 "정책 auto" 준비 추가(기대값 변경, 승인) | M3 |
| `internal/cli/factory.go` (`laneDispatchSelection`, :319) | 스탬프를 정책에서 도출 | M3 |
| `internal/cli/factory_card.go` (`next`, :1702; 안내문 :2766) | 인자 없는 `next` 게이트 | M3 |
| `internal/hook/session_start_factory.go` (:154) + 시험 (신규) | 규칙 선택을 정책에서 도출 | M3 |
| `internal/cli/factory_stage_dispatch_test.go` (신규) | AC-DSO-009 | M3 |
| `internal/cli/todo.go` (`add`, :848) + `todo_stage_test.go` (신규) | `--auto-source` 표지와 거절(수동 플래그 파싱) | M3 |
| `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md` + 사본 `.claude/skills/moai/workflows/project/doc-generation.md` | Phase 14 가 읽기 동사와 `--auto-source` 를 쓰도록 | M3, M6 |
| `internal/template/project_phase14_stage_test.go` (신규) | AC-DSO-010 의 산문 시험 | M3 |
| `internal/cli/codex_review_gate.go` (:426) + `codex_review_gate_stage_test.go` (신규) | 턴 연장 차단 게이트 | M3 |
| `internal/homestate/card_transition.go` + 시험 (신규) (+ 레코드가 영속이면 스키마 이전) | 전환 기록에 revision 필드 | M3 |
| `internal/settings/schema_sections.go` (:384 부근) + `stage_fields_test.go` (신규) | 일곱 단계 필드 | M4 |
| `internal/web/handlers.go`, `internal/web/app.go` (즉시 중지 라우트), `internal/web/fieldsets.templ` (+생성물), `internal/web/stage_policy.go` (신규: 뷰·감사), i18n 자산, `internal/web/stage_policy_handler_test.go` (신규) | AC-DSO-016–020 | M4 |
| `internal/cli/stage_policy_e2e_test.go` (신규) | AC-DSO-022 | M4 |
| `internal/cli/update/backup/stage_policy_merge_test.go` (신규) | AC-DSO-014 (시험만. 소스 변경 없음을 예상) | M5 |
| `internal/config/stage_policy_publish_scan_test.go` (신규) | AC-DSO-021 (Q4 가 (a) 일 때만) | M7 |
| `.moai/specs/SPEC-DEPLOY-STEP-OPTIN-001/progress.md` (공통) | 마일스톤별 증거 | M1–M7 |

레코드 스키마(`internal/homestate`)의 영속 형태는 plan 단계에서 읽지 않았다. M3 사전 점검에서 읽고, 마이그레이션이 필요하면 멈추고 보고한다(research.md §5 Gap 4).

## §D 마일스톤 (우선순위 순)

가장 바뀌기 쉬운 결정(데이터 모델, 사용자 흐름)을 먼저, 기계적 단계를 뒤에 둔다. 의존 관계는 각 마일스톤에 적는다.

**마일스톤 순서 조항(커밋 그래프로 검증).** RED 단계가 있는 마일스톤은 커밋 둘로 나눈다. RED 커밋 R 은 시험과, 새 API 가 필요하면 시험이 컴파일되도록 하는 최소 골격(닫히지 않은 쪽의 값을 돌려주는 자리표시 구현)을 담는다. GREEN 커밋 G 는 구현을 담고 R 의 직계 자식이다. 순서는 커밋 그래프만으로 검증한다: (1) `git merge-base --is-ancestor R G` 가 종료 코드 0, (2) `git rev-list --count R..G` 가 1 이상, (3) R 의 트리에서 마일스톤 종료 명령이 시험 단언에서 `EXPECTED_RED` 로 실패하고 컴파일 실패(TOOL_FAILURE)가 아니다, (4) G 의 트리에서 같은 명령이 기대 `--- PASS` 개수로 통과한다. 이미 존재하는 심볼만 쓰는 시험은 R 이 골격 없이 시험만 담는다. SPEC 도구에는 마일스톤 순서를 검증하는 동사가 없으므로(research.md L43, L44) 위 git 명령이 순서 검사다. 그 명령, 종료 코드, 원문 출력은 progress.md §E.2 에 남긴다. 검사 대상은 RED 단계가 있는 M2, M3, M4, M7 이다. M1·M5·M6 은 RED 단계가 없고 사유를 해당 항목에 적었다. 같은 커밋에 시험과 구현이 함께 있으면 순서를 증명할 수 없으므로(verification-claim-integrity §2.3) 허용하지 않는다.

### M1 — 결정 확인 (우선순위 High, 차단 게이트)

- 내용: 결정 행의 `Operator verdict` 가 채워졌는지 읽어 확인한다. M2 를 막는 행은 Q1·Q2·Q5, M3 는 Q3·Q6·Q11·Q14, M4 는 Q9, M5·M6 은 Q7·Q8, M7 은 Q4 다.
- 종료 조건: 해당 마일스톤의 막는 행이 모두 판정을 가진다. 비어 있는 행이 있으면 그 마일스톤은 시작하지 않고 보고한다.
- 또 하나의 전제: t1882 착지. 착지한 트리에서 research.md L03–L08 의 명령을 다시 실행해 AC-DSO-006–008 의 RED-now 셀을 새 SHA 로 다시 고정한다(verification-completeness §4: 리베이스는 다시 측정하고 다시 고정). Q11 이 이 확인의 결과로 채워진다.
- RED 단계 없음(문서 확인 마일스톤).

### M2 — 해석기와 데이터 모델 (우선순위 High, 새 타입과 새 설정 표면)

- 요구: REQ-DSO-001–006, 018(코드 기본값 부분). 인수 기준: AC-DSO-001–004, 015(a)(b).
- RED(커밋 R): 시험과, 일곱 단계 타입·해석기 함수의 골격을 담는다. 골격은 모든 단계에 `auto` 를 돌려주는 자리표시 구현이라 기본값·fail-closed 단언이 단언 실패로 빨개진다(컴파일은 된다).
- GREEN: 설계 §3·§4 의 판정 표를 구현한다. 템플릿 워크플로 파일에 주석 블록을 더한다. 읽기 동사를 더한다(Q12).
- 종료 명령(단일 호출): `go test -count=1 -v -run '^(TestStagePolicy_DefaultsAllManual|TestStagePolicy_FailClosedMatrix|TestStagePolicy_EachStageIndependent|TestStagePolicy_SourceAndRevision|TestStagePolicy_CodeDefaultsAllManual)$' ./internal/config` 가 최상위 `--- PASS` 5개(서브테스트 합계 5+15+7+6+7 = 40개), `ok`. 템플릿 시험은 `./internal/template` 에서 따로 `-run '^TestTemplateWorkflowShipsNoStageValues$'` 로 PASS 1.
- 순서 검사: M2 의 RED 커밋이 GREEN 커밋의 조상이다(위 조항).

### M3 — 외부 변경 경계의 단계 게이트 (우선순위 High, 사용자 흐름 변경)

- 선행: M2, 결정 Q3·Q6·Q11·Q14, t1882 착지.
- 요구: REQ-DSO-007–013. 인수 기준: AC-DSO-005–013.
- RED(커밋 R): 각 게이트 시험은 R 의 트리에서 `EXPECTED_RED`. 기존 `factory_card_pr_test.go` 설비가 기본 수동에서 막히지 않도록 "정책 auto" 준비 코드를 더하는 기대값 변경도 R 에 둔다. 이 변경만으로는 R 에서 기존 시험이 계속 통과해야 한다(준비 코드가 현재 코드에 무해).
- GREEN: 전달 본체의 네 지점에 판정을 넣는다(`push` 는 전이 앞에서 한 번 더). `laneDispatchSelection`·세션 시작 규칙·인자 없는 `next`, `todo add --auto-source`, 리뷰 게이트의 수리 요청, 전환 기록의 revision 을 구현한다. 문서 Phase 14 를 읽기 동사로 연결한다.
- 종료 명령(패키지별 단일 호출 다섯 개, 자원 규칙 §G). 모든 `-run` 선택자는 이름을 모두 적고 양 끝을 고정한다. 기대값은 최상위 `--- PASS` 줄 수다.
  1. `./internal/config`: `-run '^(TestStagePolicy_NoDirectReaders|TestStagePolicy_NoDirectReaders_FlagsSeededReader)$'` — 2.
  2. `./internal/cli`: `-run '^(TestFactoryDeliverByPR_PushManualStopsBeforePush|TestFactoryDeliverByPR_PushAutoStillPushes|TestFactoryDeliverByPR_PRManualStopsAfterPush|TestFactoryDeliverByPR_PRAutoWithPushManualRefuses|TestFactoryDeliverByPR_MergeManualRequestsNoAutoMerge|TestFactoryDeliverByPR_MergeAutoPinsVerifiedTip|TestFactoryDeliverByPR_FlipToManualBetweenStepsStopsAtNextBoundary|TestFactoryDeliverByPR_FlipToAutoDoesNotReplayCompletedSteps|TestLaneDispatchSelection_FollowsStagePolicy|TestFactoryNext_BareRefusedWhenDispatchManual|TestFactoryNext_NominatedAllowedWhenDispatchManual|TestTodoAdd_AutoSourceRefusedWhenStageManual|TestTodoAdd_AutoSourceAllowedWhenStageAuto|TestTodoAdd_OperatorAddIsNeverGated|TestReviewGate_RepairManualAllowsWithStateMessage|TestReviewGate_RepairAutoBlocksAsBefore)$'` — 16.
  3. `./internal/homestate`: `-run '^TestFactoryTransition_RecordsAuthorizingPolicyRevision$'` — 1.
  4. `./internal/hook`: `-run '^TestSessionStartLaneRule_ManualWhenStageManual$'` — 1.
  5. `./internal/template`: `-run '^TestProjectPhase14_DocGatesCardOnTodoCreationStage$'` — 1.
  기존 시험(`TestFactoryCompleteGitHubFlow*` 계열)은 준비 코드 추가 외 기대값 변경 없이 통과해야 한다(별도 호출, 전체 파일 범위).
- 순서 검사: RED 커밋이 GREEN 커밋의 조상.

### M4 — 웹 설정 (우선순위 High, 사용자 흐름·새 화면 모델)

- 선행: M2, 결정 Q9. M3 와는 코드 의존이 없다(AC-DSO-022 만 둘 다 필요).
- 요구: REQ-DSO-015, 016. 인수 기준: AC-DSO-016–020, 022.
- RED(커밋 R): 웹 시험은 R 에서 `EXPECTED_RED`. 시험이 새 뷰·라우트 심볼을 쓰므로 R 은 최소 골격과 템플릿 생성물(`*_templ.go`)을 함께 담아 컴파일되게 한다.
- GREEN: 스키마 필드, 뷰 모델, 검증, 응답의 재조회, 감사 기록, 즉시 중지 라우트.
- 종료 명령(단일 호출 셋). 선택자는 이름을 모두 적고 양 끝을 고정한다. 기대값은 최상위 `--- PASS` 줄 수다.
  1. `./internal/settings`: `-run '^TestStageFields_SchemaCarriesSevenStages$'` — 1.
  2. `./internal/web`: `-run '^(TestSettingsPage_ShowsStageEffectiveSourceOverrideImpact|TestSave_RejectsInvalidStageValueAtomically|TestSave_UnknownOrDuplicateStageRejected|TestSave_ResponseCarriesReReadPolicyAndRevision|TestSave_WritesAuditRecord|TestSave_RejectedSaveIsAudited|TestStopAll_SetsAllManualAndAudits|TestStopAll_LocalOverrideStillWinsAndIsDisclosed|TestSave_NonLoopbackHostDeniedForStageForm|TestSave_CrossSiteDeniedForStageForm|TestSave_ProjectWriteUnderLocalOverrideIsDisclosed|TestSave_ConcurrentStageSavesSerialized|TestSave_CannotWriteLocalTier)$'` — 13.
  3. `./internal/cli`: `-run '^TestStagePolicy_WebSaveToRuntimeEndToEnd$'` — 1 (M3 가 먼저 끝나 있어야 한다).
- 순서 검사: RED 커밋이 GREEN 커밋의 조상.

### M5 — update 보존 확인 (우선순위 Medium)

- 요구: REQ-DSO-014. 인수 기준: AC-DSO-014 (회귀 가드).
- 시험만 더한다. 현재 트리에서 이미 GREEN 이어야 하며(L34) RED 단계가 없다. 시험이 실패하면 소스 변경이 필요하다는 뜻이므로 멈추고 보고한다(blocker).
- 종료 명령: `go test -count=1 -v -run '^TestUpdate_(KeepsExplicitStageValuesCommentsUnknownKeys|AbsentStageBlockStaysManual|LocalTierUntouched)$' ./internal/cli/update/backup` 가 PASS 3.

### M6 — 템플릿 일치와 이 저장소 설정 이전 (우선순위 Medium, 기계적)

- 요구: REQ-DSO-018. 인수 기준: AC-DSO-015(c).
- 내용: `make build` 로 내장 템플릿을 재생성하고, 변경한 `.claude/` 아래 파일의 사본을 일치시킨다. 이 저장소 자신의 승인된 로컬 factory 설정(Q8)은 템플릿 변경과 **별도 커밋**으로, 운영자가 정한 계층에 명시 `auto` 를 적는다. 이 커밋은 운영자 확인 없이 만들지 않는다.
- 종료 명령: 변경 템플릿 파일 목록에 대한 `diff -q <template> <mirror>` 단일 호출들이 종료 코드 0. config 섹션 파일은 주석 블록 헤더 `grep -c` 가 양쪽 1.
- RED 단계 없음(회귀 가드).

### M7 — release/deploy 선언과 정적 검사 (우선순위 Low, Q4 가 (a) 일 때만)

- 요구: REQ-DSO-017. 인수 기준: AC-DSO-021 (회귀 가드, 양성 대조 포함).
- 종료 명령: `go test -count=1 -v -run '^(TestNoUngatedPublishCallSites|TestNoUngatedPublishCallSites_FlagsSeededCallSite)$' ./internal/config` 가 최상위 PASS 2.
- 순서 검사: 양성 대조 시험이 검사기가 실패하는 것을 관측한 커밋이 GREEN 커밋의 조상.

## §E 다른 카드와의 겹침과 순서

| 카드 | 겹치는 파일 | 순서 결과 |
|---|---|---|
| t1882 (큐 상태 `picked`, 이 트리에 변경 없음 L04–L07) | `internal/cli/factory_card_pr.go` 전체(M3 가 같은 함수의 같은 지점) | **이 SPEC 의 run 은 t1882 착지 뒤에 시작한다.** 착지 후 RED-now 를 다시 측정한다. t1882 가 auto-merge 요청을 없애면 AC-DSO-008 의 대상이 바뀐다(Q11). 이 SPEC 의 plan 은 t1882 와 병행해도 되나 착지 전에 Kickoff 를 열지 않는다. |
| t1752 | `internal/settings/schema_sections.go`, 웹 폼 파일(M4). 큐 본문이 `internal/web/projectconfig.go`·`schemaform.go` 를 언급한다. `internal/homestate/card_transition.go` 는 M3 의 revision 필드 한 줄과 겹칠 수 있다 | 의존은 없다. 먼저 착지한 쪽이 기준이고 나중 쪽이 다시 맞춘다. 두 카드의 M3·M4 에 해당하는 구현을 같은 레인이 동시에 하지 않는다(리더 순번). |
| t1644 (보류) | `/settings` 페이지 구조, 라우트 6개 삭제 목록에 `/events`·`/profile/*` | 이 SPEC 은 설정 섹션만 더한다. t1644 가 착지하면 단계 섹션은 새 /settings 구조에 맞춰 옮긴다. 이 SPEC 이 먼저 착지하면 t1644 는 단계 섹션을 보존해야 한다(그쪽 plan 의 입력). |
| t1884 → 폐기, 재발급 t1900 | AGENTS 계열 문서. 이 SPEC 은 건드리지 않는다 | 결과 문서화는 t1900 몫이다. t1900 은 "미구현은 구현 과제/한계로만 서술"을 지킨다. |
| t1894 | 카드 발급 절차 문서(스킬 gtd 계열). 이 SPEC 의 `doc-generation.md` 와 다른 파일 | 큐 본문상 코드는 t1737 몫이라 파일 겹침은 없을 것으로 읽었다. run 시작 때 `moai todo show` 로 다시 확인한다(research.md §5 Gap 6). |

## §F 위험

1. **t1882 의 형태가 정해지기 전에 M3 를 설계한 위험.** AC-DSO-008 의 대상이 바뀐다. M1 이 착지 뒤 재측정을 요구한다.
2. **이 저장소 자신의 레인 전달 중단.** 키가 없으면 수동이다. M6 의 별도 커밋 전까지 이 저장소의 `moai factory complete` 는 push 앞에서 멈춘다. M3 와 M6 의 착지 사이에 레인이 걸린다. Q8 판정과 M6 를 M3 와 같은 착지 묶음에 두는 것이 리더의 선택지다.
3. **웹 템플릿 생성물.** `.templ` 변경은 생성 파일을 함께 갱신해야 한다. R 이 컴파일되도록 생성물을 R 에 함께 둔다.
4. **local 계층 경로 해석.** 기존 해석기의 local 경로는 CWD 상대다(L40). 새 해석기는 프로젝트 루트 기준이다. 두 읽기가 어긋나면 doctor 가 보여 주는 값과 달라질 수 있다. doctor 는 단계 판정에 쓰지 않는다.
5. **감사 기록 쓰기 실패.** 저장은 막지 않되 경고를 싣는다(설계 §7). 경고가 조용히 사라지지 않는지 시험(`TestSave_WritesAuditRecord` 의 실패 변형)이 필요하다. 이 변형은 같은 시험 함수의 서브테스트로 둔다.
6. **감사에서 `dispatch` 해석 이견.** 자가 배차만 막는 해석(Q14 (a))이 감사에서 뒤집히면 AC-DSO-009 의 (d) 와 REQ-DSO-010 이 바뀐다.

## §G 자원·검증 규칙

- 호스트 부하가 높은 동안 패키지 전체 `go test` 는 쓰지 않는다. 변경 패키지로 범위를 좁히고 `-run` 선택자로 시험을 고르며, 실행된 시험 수를 센다. 무거운 패키지(`internal/cli`)는 슬롯 임대 아래서 실행한다(`moai slot acquire`).
- 이 plan 단계에서는 `go test`·`go build`·패키지 단위 도구를 실행하지 않았다. 읽기 전용 `grep`·`sed`·`git` 과 SPEC lint 만 썼다.
- 모든 마일스톤 종료 때 `git status --short` 가 승인 파일 목록 밖 파일을 보이지 않아야 한다.

## §H 교차 참조

- verification-completeness.md §2.1, §4: RED-now 네 요소와 트리 고정.
- SPEC-REVIEW-GATE-REPAIR-001: 턴 연장 차단의 상한(C = 1)과 상태 메시지 어휘. 이 SPEC 은 상태 이름 하나(`수리 요청 꺼짐`)를 더하며 그 SPEC 의 어휘 표를 고치지 않는다.
- SPEC-PROJECT-CONTINUATION-KEY-001: `continuation` 은 표현 키(Q6).
- SPEC-UPDATE-YAML-PRESERVE-001: 템플릿에 없는 키 보존(L34).
