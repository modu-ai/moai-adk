# SPEC Review Report: SPEC-GITSTRAT-WORKFLOW-READER-001

Iteration: 1/2 (Tier M 상한 = 2, `harness.yaml` `plan_audit_tier_ceilings.M:77`)
Verdict: **FAIL**
Overall Score: **0.78** (네 차원의 조화평균 — Tier M PASS 기준 0.80 미달)

- 감사 대상 트리: 워크트리 `.claude/worktrees/t656`, 브랜치 `WT-git-flow-reader`, HEAD `cd48ec891` (base `b1bd81b23` = 로컬 develop). `git diff --stat b1bd81b23 HEAD` = SPEC 5종만 추가 — 내부 코드 무변환이 감사 전제로 유효.
- 입력: spec.md, plan.md, acceptance.md, research.md, progress.md (Tier M 3종 + research + progress).
- Reasoning context ignored per M1 Context Isolation. 판정 근거는 SPEC 5종 파일과 이 감사에서 직접 실행한 명령의 관측 출력뿐이다. 디스패치 전달 사항(작성자 완료 보고 포함)은 근거가 아니라 적대 주장으로 재측정했다.
- 교차 모델 감사: 실행하지 않았다. `grep -rn audit_model .moai/config/sections/` → 무출력(audit_model 미설정 → Claude 전용 감사; t748 선례와 동일). 기계 교차검증은 `moai spec lint` + `spec_audit`(project_root=워크트리)으로 수행했다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: REQ-GWS-001…008 연속 무갭·무중복(spec.md:44-62), AC-GWS-001…012 연속 무갭·무중복(acceptance.md:9-20). 3자리 0 채움 일관.
- [PASS] MP-2 GEARS 형식 (판정 계층: **요구 계층** `REQ-GWS-001..008`, spec.md §B만): 001·004·007·008 Ubiquitous `shall`(L44, 52, 60, 62), 002·003·006 `When … shall`(L46, 48, 56), 005 `While … shall`(L54). IF/THEN 없음. 003·006의 라벨 "(Event-detected)"은 비정식 라벨이지만 구조는 Event-driven과 동일(→ D7, 경미). AC 계층의 Given-When-Then(acceptance.md)은 검증 계층의 올바른 형식이라 여기서 판정하지 않았다(M3 § Scope).
- [PASS] MP-3 YAML frontmatter: spec.md:2-15에 12 정식 필드 전부(`id` 다중 세그먼트 ID, `title`, `version: "0.1.0"` 따옴표 semver, `status: draft`, `created/updated: 2026-09-14`, `author`, `priority: P2`, `phase: "v3.2.0 target"`(금지 단계명 아님), `module: "internal/config"`, `lifecycle: spec-anchored`, `tags` 쉼표 문자열) + 공인 옵션 `tier: M`·`related_specs`(SPEC-WORKTREE-BASEREF-001:15 선례 존재). 거부 별칭(`created_at` 등) 없음. 기계 확인: `moai spec lint` → **0 error(s)**, warning 2건은 추적성(→ D4), INFO 1건(`OwnershipTransitionUnmeasured` — plan 커밋에 트레일러 없음, t637 D13과 같은 무해 관측).
- [N/A] MP-4 언어 중립성: 단일 리포지토리 Go 내부 SPEC. 편집 대상에 프로그래밍 언어 도구 열거 없음(템플릿 편집은 git-strategy.yaml 값 아닌 주석뿐 — 실은 템플릿 무변경).
- [PASS] MP-5 D7 교차-SPEC: 본문+frontmatter SPEC-ID 추출 → `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` status: completed, `SPEC-WORKTREE-BASEREF-001` status: completed(실측 `grep '^status:'`), 자기 자신 draft. retired/superseded/archived 참조 없음, 미존재 참조 없음. BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: spec.md·plan.md에서 `grep syscall` → 0건. 자동 PASS.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → research.md:35 1건 적중. 그러나 그 줄은 미해결 마커(`[NEEDS CLARIFICATION: <topic>]` 형식)가 아니라 `count: 0` **부정 개수 선언문**이다. plan.md 0건, spec.md·acceptance.md 0건. 미해결 항목 없음 → PASS. 단, 리터럴 토큰이 기계 grep 게이트를 울리는 것은 t530 교훈(해소 선언문 안의 토큰도 게이트가 센다)이 경고한 형상 → D8(선택) 재표현 권고.

보조 기계 증거: `spec_audit`(project_root=워크트리, filter=본 SPEC) → `modern_era_clean: 1`, drift 0, `EraAutoDetected` INFO 1건(H-5, frontmatter `phase`/`created` 신호 정상).

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 요구·AC 자체는 단일 해석. 그러나 소비자 배선 서술이 모순된다 — spec.md:33(§A.2 "소비자가 흐름별 대상을 해석할 수 있게 된다")·plan.md:12(G2 "이 계획이 닫는 갭")·spec.md:91(D2 "소비자가 하나의 표를 공유")는 소비자 채택을 예고하는데, spec.md:62(REQ-GWS-008 관측 동작 불변)·acceptance.md:19(AC-GWS-011 github-flow 사전/사후 동일)·plan.md:56-57(M3는 경고+doctor만)은 채택하지 않는다고 고정한다(D1) |
| Completeness | 0.90 | 1.0 | 섹션 전부(§A-§H, Out of Scope H3 3개+bullet, Risks, Cross-refs), frontmatter 완전. 누격은 A.1:24이 "세 소비자 지점"이라고 서술한 세 번째 지점(slot-lease)이 실은 소비자가 아니어서(→ D5) 그 지점의 사후 처분 서술이 공허해진 것 |
| Testability | 0.75 | 0.75 | AC-GWS-010(acceptance.md:18)의 `git diff M1..HEAD`는 `M1`이 해석 불가한 유사-ref라 문자 그대로 실행 불가 + 같은 파일에 M2 테스트가 더해지면 diff가 비지 않아 오탈 FAIL(→ D3). AC-GWS-005..008은 판정 go test의 대상 함수명을 지명하지 않음. 나머지는 명령+관측값으로 이진 판정 가능 |
| Traceability | 0.75 | 0.75 | REQ↔AC 의미 대응은 완전(acceptance.md:29). 그러나 기계 가시성이 깨져 있다 — `moai spec lint` **CoverageIncomplete WARNING 2건**: AC-GWS-006/007/008의 요구 셀 슬래시 축약(`REQ-GWS-004/005`, `REQ-GWS-004/006`)이 완전 토큰을 품고 있지 않아 REQ-GWS-005·006이 무커버로 판독됨(→ D4). 더해 doctor 신규 검사 항목(spec.md:78, plan.md:57)이 AC 0개(→ D2) |

조화평균: 4 / (1/0.75 + 1/0.90 + 1/0.75 + 1/0.75) = 4 / 5.111 = **0.78** < 0.80.

## 디스패치 지정 검증 항목 판정 (전부 자기 실행 검증)

| 항목 | 판정 | 근거 |
|---|---|---|
| AC-GWS-009/010 기계 판정 가능성 | 009 PASS-able / 010 **FAIL** | 009: `git log --oneline`에서 M1(`test(config): characterize …`, plan.md:47)이 M2(`feat(config): …`, plan.md:53) 아래(과거)에 있음을 관측 — 명령+관측값 이진 판정 가능. 010: `git diff M1..HEAD -- …`은 `M1`이 ref가 아님(plan은 커밋 주제만 정의, SHA/tag 없음) + M1이 "기존 t637 파일 확장"(plan.md:42)이고 M2 테스트 파일 미지명(plan.md:51)이라 같은 파일에 더해지면 diff 비어있지 않음 → 올바른 구현도 FAIL 또는 판정자 재량 해석 필요(D3) |
| 소비자 커버리지 (acquire/automerge/slot-lease) | 부분 | invalid 처분: acquire 경고 AC-GWS-004 커버. automerge 게이트: 불변 명시(REQ-GWS-005, plan M3 "gate skip stays silent") — 무변경이므로 AC 불요는 정당. **slot-lease는 소비자가 아니다** — 비시험 Go 전수 grep에서 `LoadGitFlowDevelopBranch(` 프로덕션 호출 0건, `loader_slot_lease.go:6`은 "modelled on" 주석뿐이고 실제 키는 `workflow.slot_lease.default_max_duration`. SPEC의 "세 소비자 지점"(spec.md:24)·research.md:21은 거짓 관측(→ D5). valid-non-git-flow 대상 해석(github-flow→main 등)은 소비자 어디에도 배선되지 않음(→ D1) |
| D1 3-way disposition 타당성 | PASS | reader 불오류(t449 fail-open) 보존 + 구조적 처분 + 소비자 경고 + doctor 신규 검사 설계는 doctor_worktree_base.go(`checkWorktreeBaseBranch` → DiagnosticCheck) 선례와 일치. fail-open 계약 명시(spec.md:48, 109) 적정 |
| D2 precedence | PASS | 흐름 범위 키라 경쟁 없음 — `develop_branch`는 REQ-GWS-005 게이트로 비-git-flow에서 무시. 실측: `IsGitFlow() = Manual && GitFlowWorkflow`(loader_integration_branch.go:57-59), 슬롯릴리스 부재 확인으로 비-git-flow 하 develop_branch 무시 경로 소비자 2곳 모두 유지 |
| D3 위자드 연기 근거 | PASS (근거 성립) | 실측: `internal/cli/wizard_config_test.go:253-302` — 위자드는 git-strategy.yaml에 mode/provider만 기록, workflow 쓰기 경로 없음. 리더 정확성과 무관한 UX 표면이라는 연기 논리가 측정에 뒷받침됨 |
| D4 shipped_key_inventory 포함 | PASS | `internal/config/testdata/shipped_key_inventory.yaml` 500/557/623행에 `git_strategy.{manual,personal,team}.workflow` 실재 확인. M4의 evidence 갱신은 소규모·기계적 — 포함 정당 |
| 이력 귀속 (t449/t637/t655) | 핵심 확인·1행 부정확 | `git log -- internal/config/loader_integration_branch.go` → 정확히 2커밋: `4d58f407b`(card t449, 생성)·`ba0725be8`(card t637, struct+IsGitFlow+경고). 파일 헤더 주석 "card t449"(L3)·"card t637"(L39) 일치 — **t449/t637 귀속은 조작 아님, 실측 확인**. `git log -S LoadGitFlowIntegrationConfig -- internal/cli/integration.go` → ba0725be8(t637)만 — integration.go:283을 t655로 적은 research.md:15는 부정확. t655=`1d0f08f10`(SPEC-WORKTREE-KEY-WIRING-001, session_worktree_automerge.go 생성) — automerge 쪽 귀속은 정확(→ D5) |
| 커버리지 baseline 주장 | 확인 | 이번 실행·이 트리: `go test -coverprofile … -run 'TestLoadGitFlow'` → ok, `go tool cover -func` → LoadGitFlowDevelopBranch 100.0%·IsGitFlow 100.0%·LoadGitFlowIntegrationConfig 100.0%. research.md:14·spec.md:107의 "100%, 회귀 금지" 근거 유효 |
| 기타 in-tree 사실 | 전부 확인 | types.go Workflow/Environment/ReleaseBranchPrefix 필드, defaults.go 3프로필 `Workflow: "github-flow"`, 템플릿 `workflow: github-flow` 18/50/86행(=3), quality.yaml `development_mode: tdd`·`test_coverage_target: 85` |

## Defects Found

| ID | 심각도 | 분류 | 위치 | 결함 | 필요한 수정 |
|---|---|---|---|---|---|
| D1 | major | blocking | spec.md:33, 91; plan.md:7, 12 대 spec.md:62; acceptance.md:19; plan.md:56-57 | 소비자 배선 서술 모순. §A.2는 "acquire/automerge 소비자가 흐름별 대상을 해석할 수 있게"라고 목적을 걸고, plan G2는 "github-flow가 caller fallback으로 해석하는 갭"을 "이 계획이 닫는 갭"으로 적으며, D2는 "소비자가 하나의 표를 공유"라고 하나 — M2의 resolver를 호출하는 프로덕션 지점이 어디에도 없고(M3는 경고+doctor만), REQ-GWS-008·AC-GWS-011이 github-flow의 acquire/automerge 관측 동작을 사전과 동일로 고정한다. plan §F 대로 구현하면 G2는 리더 API 수준에서만 닫히고 §A.2의 목적절은 거짓이 되며, 해석 표(github-flow→main, gitlab-flow→environment, release-flow→prefix)는 태어난 날부터 소비자 0인 미사용 API가 된다 | 하나를 고른다. (a) 최소: spec.md:33을 "리더가 해석을 노출한다(소비자 채택은 후속 카드)"로, plan.md:12 G2를 "리더-API 수준에서 닫는다"로 고치고 plan §F M2에 "resolver의 프로덕션 호출 지점: 없음(단위 시험만)"을 명시. (b) 권장: M3 doctor 검사가 흐름별 해석 대상(환경값/접두어 포함)을 함께 보고하게 와이어해 표에 소비자를 하나 두고 G2를 부분 실현 — 어느 쪽이든 REQ-GWS-008·AC-GWS-011과 충돌 없음 |
| D2 | major | blocking | spec.md:74-78 (C.1 표·doctor 문단); plan.md:57; acceptance.md:39 (§D.4) | `moai doctor` 신규 검사 항목이 plan M3의 납품물이면서 AC 매트릭스에 행이 없다(AC-GWS-001..012 어디에도). §D.4가 "doctor 항목은 검사 함수 표 시험으로 검증"이라 선언하지만 그 시험을 요구하는 AC가 없어 — run §E 행렬에도, sync 감사 대조에도 걸리지 않는 미채택 기준이 된다 | AC-GWS-013 추가: "Given invalid-workflow 픽스처 / When doctor 검사 함수 실행 / Then 상태 WARN + 출력에 해당 값과 4개 허용값 명시; git-flow→OK, valid-non-git-flow→OK-무침묵 아닌 OK, 표 시험 3상태" + `go test ./internal/cli/ -run TestDoctorGitStrategyWorkflow -count=1` 형태의 판정 명령. §D.1 must-pass와 §D.2 추적표에도 편입 |
| D3 | major | blocking | acceptance.md:18; plan.md:42, 51 | AC-GWS-010 판정 명령 `git diff M1..HEAD -- internal/config/loader_integration_branch_test.go`에서 `M1`은 존재하지 않는 ref다(커밋 주제 예시만 있고 SHA/tag 없음). 게다가 M1이 그 파일을 "확장"(plan.md:42)하고 M2 단위 시험의 파일이 미지명(plan.md:51)이라, M2 시험이 같은 파일로 가면(자연스러운 선택) 아무 부정행위 없이도 diff가 비지 않아 AC가 거짓 FAIL을 내거나 "M1-test 편집만 아니다"라는 재량 해석을 강제한다 — 이진 판정 상실 | plan M2에 "M2 신규 disposition 시험은 별도 파일(예: `loader_workflow_disposition_test.go`)에 둔다"를 명시하고, AC-GWS-010을 `git diff <M1 커밋 SHA>..HEAD -- internal/config/loader_integration_branch_test.go` (SHA는 M1 커밋 시점에 핀) + 기대 관측 "빈 출력"으로 고정. AC-GWS-005..008도 대상 시험 함수명과 `go test -run` 선택자를 지명 |
| D4 | major | blocking | acceptance.md:14-16, 29 | AC-GWS-006/007/008의 요구 참조가 슬래시 축약(`REQ-GWS-004/005`, `REQ-GWS-004/006`)이라 완전 토큰 `REQ-GWS-005`·`REQ-GWS-006`이 문서 어디에도 리터럴로 없다 — 기계 판정으로 REQ 2개가 무커버. 실측: `moai spec lint` → `CoverageIncomplete WARNING ×2` ("REQ-GWS-005 is not referenced by any AC", "REQ-GWS-006 is not referenced by any AC"). lint 경고 2건은 run 전 0이어야 할 추적성 결함 | 세 셀의 요구 열을 완전 토큰 나열로 고친다: `REQ-GWS-004, REQ-GWS-005` / `REQ-GWS-004, REQ-GWS-006`. §D.2 추적표도 동일 형식으로. 수정 후 `moai spec lint` 재실행 → CoverageIncomplete 0 확인 |
| D5 | major | blocking | spec.md:24 (§A.1); research.md:15, 21 | 거짓 프로덕션 사실. research.md:21이 `internal/config/loader_slot_lease.go`를 "slot acquire via LoadGitFlowDevelopBranch" 소비자로 적고 spec.md:24가 이를 "세 소비자 지점" 중 하나로 상향했으나 — 비시험 Go 전수 grep에서 `LoadGitFlowDevelopBranch(` 프로덕션 호출은 0건이고, loader_slot_lease.go:6은 "modelled on LoadGitFlowDevelopBranch"(설계 선례 인용)일 뿐 실제 키는 `workflow.slot_lease.default_max_duration`이다. 작성자의 grep이 주석 행을 호출로 오독한 것. 아울러 research.md:15는 integration.go:283을 t655 귀속으로 적었으나 `git log -S` 실측상 t637(`ba0725be8`)이 착지했고 t655(`1d0f08f10`)는 automerge 쪽만 착지 | spec.md:24를 "두 소비자 지점(acquire, 세션-종료 auto-merge 게이트)"으로 정정하고 slot-lease는 "설계 선례 형제, 소비자 아님"으로 research §2에 기록. research §1 표의 t655 행에서 integration.go:283을 제거해 t637 행으로 옮긴다. 부수 효과: 디스패치가 물은 "slot-lease의 gitlab-flow/release-flow AC 필요 여부"는 소비자 부재로 불필요 — 소비자가 없으므로 새 처분이 닿을 표면이 없다 |
| D6 | minor | optional | spec.md:129 (R2); internal/config/defaults.go 763-788 | gitlab-flow 대상키 `environment`의 기본값은 "local"(manual)/"github"(personal/team) — 환경 라벨이지 브랜치명이 아니다. 기본값을 그대로 둔 gitlab-flow 채택자는 해석 대상으로 "local"·"github" 같은 비-브랜치 문자열을 얻는다(REQ-GWS-006은 빈 값만 다룸). 매핑 자체는 운영자 결정이라 재론하지 않는다 — 위험 서술만 누락됐다 | R2에 한 줄 추가: "기본 environment 값은 브랜치명이 아니므로 gitlab-flow 채택자는 environment를 브랜치명으로 재설정해야 한다; doctor 검사가 알려진 비-브랜치 기본값에 WARN을 고려(구현 재량)". |
| D7 | minor | optional | spec.md:48, 56 | REQ-GWS-003·006의 라벨 "(Event-detected)"는 정식 GEARS 라벨이 아니다(구조는 Event-driven과 동일 — MP-2에는 영향 없음, t748 D3 동일 지적) | 라벨을 "(Event-driven)"으로. |
| D8 | minor | optional | research.md:35 | 부정 개수 선언문이 리터럴 `[NEEDS CLARIFICATION]` 토큰을 품고 있어 MP-7의 기계 grep이 적중한다(본 감사는 문맥 판독으로 PASS). t530 교훈(게이트는 의도가 아니라 존재를 센다)의 재발형 | "open questions: none" 등 토큰 없는 표현으로 재표현. |
| D9 | minor | optional | spec.md:46, 60 | 요구 계층이 구체 함수명(IsGitFlow(), LoadGitFlowIntegrationConfig, LoadGitFlowDevelopBranch)을 지명한다 — 일반 RQ-4(구현 세부 배제)와 긴장. 다만 이 SPEC의 도메인 표면이 곧 그 함수들이라 특성화 SPEC 성격상 허용 판정(REQ-GWS-005의 types.go 계약 인용도 동일) | 수정 불요 — 관측 기록. 후속 SPEC에서는 도메인 개념어("판별자", "리더") 우선 사용 권고. |

## Regression Check

해당 없음(반복 1/2 — 선행 반복 보고서 없음). 단, `.moai/reports/t656/plan-audit-iter{1,2,3}.md`는 **폐기 전제**(0-readers) 시절의 다른 SPEC에 대한 구 보고서로, 본 보고서와 파일명이 겹치지 않게 본 보고서는 t637 명명 규약(`plan-audit-<SPEC-ID>-iter1.md`)을 따랐다.

## Recommendation

**FAIL — 최소 수리 목록(run 진입 전 manager-spec 경유 1회 수정으로 충분):**

1. **D5** — spec.md:24·research.md:15/21의 slot-lease 소비자 주장 정정 + t655/t637 귀속 정정. (사실 무결성 — 최우선)
2. **D4** — acceptance.md:14-16·29의 슬래시 축약 참조를 완전 토큰으로 → `moai spec lint` CoverageIncomplete 0 확인.
3. **D2** — AC-GWS-013(doctor 검사 3상태 표 시험) 추가 + §D.1/§D.2 편입.
4. **D3** — plan M2에 M2 시험 별도 파일 명시 + AC-GWS-010을 M1 커밋 SHA 핀 diff로 재작성 + AC-GWS-005..008에 판정 명령 지명.
5. **D1** — 소비자 배선 서술 단일화: (a) 서술을 리더-노출 한정으로 축소하거나 (b) doctor가 해석 표를 소비하게 와이어. REQ-GWS-008·AC-GWS-011과 충돌하지 않는 쪽으로.

선택(D6-D8)은 같은 편집에서 무료 수리 권장. 5건 모두 수리 후 재감사는 본 목록의 결함 delta 범위로 수행한다(Retry Loop Contract, Tier M 상한 2회).

— 보고서 작성: plan-auditor (반복 1/2). 판정 근거는 모두 이 워크트리에서 직접 실행한 명령의 관측 출력이다.
