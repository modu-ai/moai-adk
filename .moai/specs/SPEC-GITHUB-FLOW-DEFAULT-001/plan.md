# SPEC-GITHUB-FLOW-DEFAULT-001 — 실행 계획

> 기준 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2`(카드 트리). 모든 좌표는 `acceptance.md` §B 증거 원장과 `research.md` 에서 온다. 시간 예측 없이 우선순위와 마일스톤 순서로만 서술한다.

## §A 전제

### A.1 두 계층과 병합 시점

| 계층 | 마일스톤 | 병합 경로 |
|---|---|---|
| PRE-CUTOVER-SAFE | M1, M2, M2a, M3, 스윕 가드의 픽스처 자가 시험(M4 의 일부) | 현 git-flow 경로로 develop 에 병합 — 이 경로를 마지막으로 쓴다 |
| CUTOVER-TIME | M4(가드 트리 단언 포함), M5 | 카드 브랜치에서 준비·검증하고 병합 가능 상태로 보류, 런북 2단계에서 한 묶음으로 develop 에 마지막으로 병합 |
| 절체 절차 | M6 | 문서·점검 스크립트·리허설 증거는 카드 산출물, 실행은 리더·운영자 |
| 퇴역 뒤 정리 | (이 카드 밖) | 워크플로 push 트리거의 `develop` 제거는 develop 삭제 뒤 리더가 정하는 정리 카드의 몫. 이 카드는 런북에 점검 문구만 남긴다(design D-13 단계 4) |

### A.2 이 SPEC 이 반드시 지키는 경계

- 기준 브랜치를 실제로 바꾸는 단계는 리더 확인 없이 실행하지 않는다(REQ-GFD-020, §E).
- 규칙 문서 편집은 t1448 이 develop 에 착지한 뒤에 시작한다(§G.3).
- 카드 브랜치에서 `develop` 을 주기적으로 흡수해 보류 묶음이 오래 낡지 않게 한다. 마지막 흡수는 절체 직전이며 스윕 가드의 무장 실행이 그 흡수의 최종 검사다.

### A.3 방법론

`quality.yaml` 의 개발 방법론은 TDD 다. 모든 코드 마일스톤은 RED 를 먼저 관측한다 — 위 인수 기준의 돌연변이 프로브가 그 RED 입력이다. 단일 작성자 원칙에 따라 한 트리에 한 writer 로 수행하며 병렬 write 는 하지 않는다.

## §B 결정 검토 순서 (바뀔 가능성이 큰 것부터)

plan-auditor 와 리더의 검토 시간은 아래 순서에 쓰는 것이 효율적이다. 데이터 모델·인터페이스·사용자에게 보이는 흐름이 앞이고 기계적 문서 갱신이 뒤다. 각 항목의 대안과 기각 근거는 `design.md`.

| 순서 | 결정 | 이유 | 위치 |
|---|---|---|---|
| 1 | D-4 카드 전달 상태 모델 | 카드 기록의 상태·전이가 바뀌는 데이터 모델 변경 | design §D-4 |
| 1a | D-30 투영 병합 목표의 인자 방식(개정 A-1) | 공개 함수 시그니처가 바뀌는 인터페이스 변경, 호출자가 없는 함수라 영향은 시험 호출 지점에 한정 | design §D-30 |
| 2 | D-3 병합 큐 보류 | 사용자(리더)에게 보이는 흐름 결정, 측정 부족 | design §D-3 |
| 3 | D-7·D-22·D-24 3-OS 게이트 위치와 스위치, D-6·D-16 rc 규칙 | 릴리스 흐름과 되돌릴 수 없는 태그에 닿는다 | design §D-7, §D-6, §D-16~D-24 |
| 4 | D-8·D-17·D-25 수렴 방식과 선행 조건, D-13 develop 퇴역 단계 | 외부 공유 시스템, 되돌리기 어려움 | design §D-8, §D-13, §D-17, §D-25 |
| 5 | D-5 착지 판정 세 층 | 기계적이지만 관측으로 반증된 가정 위에 있다 | design §D-5 |
| 6 | D-10 t810 흡수, D-11 t1452 좁힘 | 인접 카드와의 소유 경계 | design §D-10, §D-11 |
| 7 | D-9 스윕 가드 설계, D-12 식별자 동결 | 시험 설계 | design §D-9, §D-12 |
| 8 | 나머지 기계적 문서·미러 갱신 | 지침 반영 | M4, M5 |

## §C 마일스톤

각 마일스톤은 **착수 증거**(시작해도 되는 관측)와 **종료 증거**(끝났다고 말할 수 있는 관측)를 갖는다. 종료 증거는 `acceptance.md` 의 기준 ID 로 가리키며 판정서 `.moai/reports/t1453/verdict.md` 에 명령과 출력을 옮겨 적는다.

### M1 — PRE-CUTOVER-SAFE 코드·설정 재지정

**범위(RENAME 계열)**: 기준 브랜치 해석을 구성된 통합 목표로 옮긴다. 대상은 증거 원장 E-01(`landingBaseBranch`)·E-02(`sweep --base` 기본값)·E-28(카드 전달의 기준을 읽는 두 지점 `factory_card.go:1064`·`factory_merge.go:93`)·E-29(`codex_review_scope.go:55` 의 `cardBaseBranch`, `goal.go:326` 의 `MergeTarget`)와 입력 보고서의 RENAME 항목 중 읽어서 진입을 확인한 것(세션 정리의 통합 ref, `scripts/jev/triage.py` 기본값, `kanban` 착지 ref)이다. 이 목록은 단어 패턴 후보이며 항목마다 RED 를 쓰기 전에 읽어서 진입 여부를 정한다. **세션 종료 자동 병합(`session_worktree_automerge.go:176,183`)과 통합 창(`integration.go:353`)의 `DevelopBranch` 읽기는 옮기지 않는다** — github-flow 에서 휴면이 맞는 git-flow 전용 경로이고, 통합 목표로 옮기면 main 으로의 로컬 병합이 켜진다(design D-1). 와이어 식별자와 `contract/testdata/mission_surface_baseline.txt` 는 건드리지 않는다(design §D-12). `MergeTarget` 변경이 그 스냅숏을 바꾸면 그 사실을 판정서에 적고 해당 항목은 이번 범위에서 뺀다.

- 착수 증거: 카드 트리에서 E-01·E-02·E-28·E-29 가 기록과 같은 출력을 낸다(원장 재실행).
- 순서: (1) git-flow 특성화 시험으로 현 출력을 고정 → (2) 해석표 전 행에 대한 표 기반 RED → (3) 읽기 교체 → (4) 두 구성에서 GREEN.
- 종료 증거: AC-GFD-001·003. E-01·E-02·E-28·E-29 명령이 종료 코드 1 과 빈 출력(해석기 파일 제외). 세션 종료 자동 병합이 github-flow 에서 휴면임을 보이는 양성 시험이 초록(AC-GFD-001).
- 계층: PRE-CUTOVER-SAFE. git-flow 구성에서 출력이 같음이 병합 조건이다.

Exit: AC-GFD-001, AC-GFD-003
- 되돌리기: 커밋 되돌리기. 동작 변화가 github-flow 구성에만 있어 되돌려도 현행 흐름은 영향이 없다.

### M2 — github-flow 카드 전달 간선

**범위**: (a) 카드 상태 모델 결정(design §D-4, 소비자 세 지점 포함) 후 `moai factory complete` 의 PR 간선(push, 대상이 통합 목표인 PR 개설, 병합 완료 관측 후 기록), (b) PR 개설 전 병합 준비 점검, (c) 착지 판정 세 층을 `worktree done`·`sweep` 이 공유하고 세션 종료 정리는 1·2층만 쓰도록(design §D-5), (d) github-flow 에서 창이 선행 조건이 아니게. `push_serializer`·push 사전 점검은 이름만 main 으로 바꾸면 모든 main push 를 직렬화하므로 github-flow 에서 휴면으로 둔다(design §D-12). **계약 모드 에스컬레이션 분류기(`pushWhy` 류)는 이 마일스톤에서 바꾸지 않는다**(design D-23 — 1회차 계획의 (d)를 삭제했다).

- 착수 증거: M1 병합 완료(해석기가 develop 에 있음), 원장 E-14·E-15·E-16 재실행이 기록과 같다.
- 종료 증거: AC-GFD-002·004·005·006. (병합 큐 판정의 기록인 AC-GFD-022 는 M6 의 종료 증거이고 이 마일스톤의 것이 아니다.)
- 계층: PRE-CUTOVER-SAFE(github-flow 구성에서만 새 동작). 실제 gh 호출은 시험에서 대역으로 한다.

Exit: AC-GFD-002, AC-GFD-004, AC-GFD-005, AC-GFD-006
- 되돌리기: 커밋 되돌리기. 새 카드 상태 값이 기록 스키마에 들어가므로 되돌린 뒤 이미 기록된 새 상태가 남지 않았음을 확인한다(design §D-4).

### M2a — 계약에서 미션으로의 투영이 쓰는 병합 목표 (개정 A-1, M2 뒤 — M1 의 원장 밖에 있던 자리)

**범위(RENAME 계열)**: `internal/contract/projection_mission.go` 의 상수 `missionMergeTarget = "develop"`(원장 E-46)을 지우고, `ProjectToMission` 이 구성된 통합 목표를 인자로 받아 투영된 미션 계약의 `MergeTarget` 으로 쓴다(design D-30). 인자가 비었거나 공백뿐이면 필드 `merge_target` 을 이름으로 대는 `ErrNotProjectable` 거부이고 `develop` 대체값은 없다(M1 이 기록한 옵션 A 와 같은 결정). 같은 변경에 그 함수를 부르는 시험 파일 `internal/contract/projection_mission_test.go` 의 호출 지점 네 곳(원장 E-48)과 새 표 기반 시험 `TestProjectionMergeTargetFollowsIntegrationTarget` 이 든다. **건드리지 않는다**: 와이어 식별자(`local-merge-develop`·`push-develop`·`push_develop`·`--develop-worktree`·`local_develop_merge`), `contract/testdata/mission_surface_baseline.txt`, `internal/mission` 전체(`git_owner.go:201` 의 `branch != "develop"` 포함 — design D-30), 그리고 계약 코어가 `internal/config` 를 import 하는 일(원장 E-52 — 구성은 호출자가 읽는다).

- 착수 증거: M2 종료 증거가 관측됨(단일 작성자 직렬). 원장 E-46·E-47·E-48·E-51·E-52·E-54 를 재실행해 기록과 같음을 확인한다 — HEAD 가 `eda61419…` 와 다르면 재측정하고 다시 핀한다(원장 E-46 이후의 핀은 그 트리다).
- 순서: AC-GFD-024 의 녹색 경로 — (1) 변경 전 트리에서 git-flow 행의 봉인 해시를 특성화 시험으로 고정 → (2) 다섯 행의 표 기반 RED → (3) 시그니처와 상수 교체 → (4) 시험 호출 지점 네 곳 갱신 → (5) 두 시험 `TestProjectionMergeTargetFollowsIntegrationTarget`·`TestContractProjectsOntoMissionValidator` 초록.
- 종료 증거: AC-GFD-024. 스냅숏 파일과 `internal/mission` 이 편집되지 않았음은 같은 기준의 범위 가드(E-54 형태, 종료 트리 기준)가 판정한다.
- 계층: **PRE-CUTOVER-SAFE, 단 현 git-flow 구성에서 동작이 바이트 단위로 같을 때에 한한다.** 증명은 네 가지가 함께 서야 한다 — (a) 투영에 비테스트 호출자가 없다(E-48) 그래서 어느 바이너리의 운영 경로도 움직이지 않고, (b) git-flow 행(`develop`)의 봉인 해시가 변경 전 해시 리터럴과 같으며(AC-GFD-024 (3)), (c) 스냅숏 비교(limb 5)가 초록이고 스냅숏·`internal/mission` 이 무편집이며(E-51·E-54), (d) 동결 식별자 줄 문언이 같다(E-53). 넷 중 하나라도 어긋나면 PRE-CUTOVER-SAFE 로 병합하지 않고 마일스톤을 멈춰 리더에게 올린다 — 스냅숏을 손으로 다시 생성하는 길은 이 마일스톤의 선택지가 아니다.

Exit: AC-GFD-024
- 되돌리기: 커밋 되돌리기. 투영에 비테스트 호출자가 없으므로(E-48) 이 투영이 낸 계약이 운영에서 저장되는 경로가 코드에 없고, 되돌려도 남는 기록 상태가 없다. 되돌린 뒤에는 시험 호출 지점이 한 인자 시그니처로 돌아갔는지 `go vet ./internal/contract/` 로 확인한다.

### M3 — 릴리스 기구

**범위**: (a) 출처 검증 로직을 `scripts/verify-release-provenance.sh` 로 옮겨 `release.yml` 이 호출(design §D-6 — M3 은 `release.yml`·`.goreleaser.yml`·`scripts/` 를 바꿀 수 있다, 태그 push 에서만 도는 파일이고 접미사 없는 태그의 동작은 같다. `release.yml` 은 인라인 검사 1~7 의 본문을 지우고 `bash scripts/verify-release-provenance.sh "${TAG}"` 한 줄 호출로 바뀌며, 그 배선과 접미사 없는 태그의 동치는 AC-GFD-008 이 스크립트와 함께 판정한다), (b) rc 규칙 R-a(design §D-6 의 규칙문), (c) `.goreleaser.yml` 의 `release.prerelease: auto`, (d) `scripts/release.sh` 의 detached HEAD 허용과 rc 태그의 CHANGELOG 검증 건너뛰기, (e) 태그 전 매트릭스 확인(옵션 B, 운영자 결정 D-22)을 `--require-matrix-run` 옵션 뒤에 둔다(기본 꺼짐, design §D-24). 켜는 일은 M5 가 한다.

- 착수 증거: 원장 E-04·E-05·E-06·E-09·E-31 재실행이 기록과 같다. 입력: 옵션 B 는 운영자 결정으로 확정돼 있다(D-22).
- 종료 증거: AC-GFD-007·008·009·010.
- 계층: PRE-CUTOVER-SAFE. 단, **매트릭스 확인의 강제는 절체 시점에 켠다** — 현 git-flow 릴리스는 `release/*` PR 헤드에서 매트릭스를 돌리고 main 병합 SHA 에는 돌리지 않으므로, 강제를 지금 켜면 현행 정식 릴리스가 막힌다. 코드는 옵션 뒤에 두고(옵션이 없으면 현행 동작) 켜는 일은 M5 의 절체 묶음이 릴리스 하네스 본문에서 한다.

Exit: AC-GFD-007, AC-GFD-008, AC-GFD-009, AC-GFD-010
- 되돌리기: 커밋 되돌리기. 태그는 불변 규칙셋이므로 **이 카드는 어떤 rc 태그도 만들지 않는다**(acceptance §E).

### M4 — CUTOVER-TIME 문서·지침 묶음 (t1448 착지 뒤)

**착수 증거(전부 필요)**: (1) t1448 이 develop 에 착지 — 착지 여부는 `git log` 의 카드 id 병합 흔적으로 읽고 읽은 결과를 판정서에 적는다, (2) M2 병합 완료(문서가 새 간선을 서술하므로), (3) 상시로드 기준선 재측정(research §10).

**범위**: `AGENTS.md`·`CLAUDE.md`·`.claude/rules/moai/**`(kanban-dispatch*·main-checkout-branch-guard*·worktree-integration*·spec-workflow·agent-common-protocol*·contract-autonomy 등)·`.claude/agents/moai/*`(manager-git·manager-lead·manager-docs 등 develop 서술이 있는 것)·`.claude/skills/**`·README 4종·docs-site 4로케일·템플릿 사본, Frozen 조항 정규 개정, t810 의 REQ-LBR-001..006 흡수, 스윕 가드 본체. 템플릿 우선 순환으로 로컬과 템플릿을 함께 바꾼다. `make build` 와 임베드 축 확인은 저장소 절차에 따르고 이 카드의 관측이 아니다(acceptance §E).

- 순서: (1) 가드 픽스처 시험(사전 병합 가능) → (2) 편집 대상 파일마다 `moai constitution list --file <경로>` 로 `[ZONE:Frozen]` 줄의 등재 여부를 판정해 판정서에 적고(design D-26 보충: `spec-workflow.md` 는 등재 둘 — 027·028 — 과 미등재 표지 줄 셋), Frozen 개정 입력(`--before`/`--after`/`--evidence`)과 `--dry-run` 제안을 준비하고 **운영자가 `amend` 를 두 번 실행**(design D-26), 레인은 `validate` 종료 코드 0 과 문언 일치를 검증 → (3) 규칙·에이전트·스킬 갱신(`worktree-integration.md` 의 등재되지 않은 Frozen 표지 두 줄과 `spec-workflow.md` 의 미등재 표지 줄 둘은 일반 편집 — 개정된 등재 문언과 같은 방향, 로컬·템플릿 사본 동일) → (4) README·docs-site(oss-docs 하네스의 ko 정본→en→ja/zh 순) → (5) 가드 트리 단언 무장 → (6) 상시로드 전후 측정. 편집마다 AC-GFD-015 의 mirror guard 네 시험을 돌린다.
- 종료 증거: AC-GFD-012·013·014·017·021. **AC-GFD-011(가드 본체의 무장 실행)은 M5 의 종료 증거다** — 가드는 이 저장소의 `git-strategy.yaml` 값(M5 가 바꾼다)으로 무장하고 `AGENTS.local.md`·`.claude/rules/local/**`·`.moai/docs/*doctrine*`(M5 범위)까지 훑으므로, M4 와 M5 가 한 묶음으로 적용된 트리에서만 관측할 수 있다.
- 계층: CUTOVER-TIME(가드 픽스처 시험만 PRE-CUTOVER-SAFE).
- 되돌리기: 보류 묶음이라 미병합 시 폐기, 병합 후에는 커밋 되돌리기. Frozen 개정은 되돌리려면 역방향 amend 가 필요하다.

Exit: AC-GFD-012, AC-GFD-013, AC-GFD-014, AC-GFD-017, AC-GFD-021

### M5 — CUTOVER-TIME 설정·CI·로컬 지침 묶음

**범위**: 이 저장소의 `git-strategy.yaml`(워크플로·`develop_branch`·`worktree_base_branch`)·`workflow.yaml`(`deny_commits_on`·`push_develop` 의 값 처분)·`.coderabbit.yaml`·`spec-lint.yml` 의 develop 의존(가져오기와 릴리스 스냅숏 논리)·릴리스 하네스(`hns-release-specialist` 재작성과 `--require-matrix-run` 호출로 M3 의 매트릭스 확인 켜기)·`AGENTS.local.md`·`.claude/rules/local/**`·`.moai/docs/*doctrine*`. **워크플로 push 트리거 목록(ci·codeql·graph-freshness·judgment-first-consistency·lsel-leak-guard·template-neutrality-check·test-install·workflow-parse-guard·docs-i18n-check)은 이 묶음에 넣지 않는다** — develop 팁 CI 가 절체 내내 관측되도록 develop 이 퇴역할 때까지 남기고, 퇴역 뒤 정리 단계에서 뺀다(design D-17·D-13 단계 4).

- 착수 증거: M1·M2·M2a·M3 병합 완료, M4 준비 완료.
- 종료 증거: AC-GFD-016, **AC-GFD-011**(M4+M5 묶음에서 무장한 가드 본체가 녹색 — 가드 픽스처 시험 AC-GFD-021 은 M4 종료 증거), AC-GFD-015 의 mirror guard 네 시험 초록.
- 계층: CUTOVER-TIME. 이 묶음은 develop 이 main 으로 수렴하기 전에 develop 에 병합되는 마지막 CUTOVER-TIME 변경이며(런북 2단계) M4 와 함께 한 묶음으로 병합된다. M6 의 산출물(스크립트·시험·런북)은 기준 브랜치를 바꾸지 않는 새 파일이라 같은 병합이거나 그 앞 병합으로 develop 에 들어간다. 이 카드 자신의 브랜치는 이 병합과 push 뒤에 사전 점검에서 제외된다(design D-25).
- 순서 제약: `spec-lint.yml` 의 develop 가져오기 제거는 develop 브랜치 삭제보다 앞선다(design §D-13).

Exit: AC-GFD-011, AC-GFD-016, AC-GFD-015
- 되돌리기: 커밋 되돌리기. 설정 값은 되돌리면 git-flow 로 돌아가지만 이미 github-flow 기준으로 만든 카드 PR 은 영향 범위를 점검해야 한다.

### M6 — 절체 런북과 리허설

**범위**: 런북 문서(`.moai/specs/SPEC-GITHUB-FLOW-DEFAULT-001/cutover-runbook.md`, 수렴 병합 절차·트리 항등 검증·되돌리기·레인 재기동 절차·절체 후 확인 목록·develop 퇴역 단계와 퇴역 뒤 잔존 점검), 배치 경계 사전 점검 스크립트(`scripts/cutover-precheck.sh`), 수렴 리허설 스크립트(`scripts/cutover-rehearsal.sh`). 스크래치 클론에서 수렴과 되돌리기를 리허설한다. 리허설은 GitHub 를 건드리지 않는다. 병합 큐 판정과 t1452 조정의 기록(AC-GFD-022)과 권고 목록은 `progress.md` §G 에 있다.

- 착수 증거: M1~M5 준비 완료.
- 종료 증거: AC-GFD-018·019·022·023, AC-GFD-020 의 보호 상태 대조.
- 런북 구성은 design §D-8·§D-13. 각 단계에 실행 주체(카드 / 리더 / 운영자)와 외부 공유 시스템 여부를 표시한다.

Exit: AC-GFD-018, AC-GFD-019, AC-GFD-020, AC-GFD-022, AC-GFD-023

## §D 의존·순서

| 마일스톤 | 선행 | 병합 시점 | 외부 의존 |
|---|---|---|---|
| M1 | 없음 | 즉시 가능 | 없음 |
| M2 | M1 | M1 뒤 | 없음 |
| M2a | M2(단일 작성자 직렬. 기술적으로는 M1 과 같은 구성 해석을 쓰지만 투영에 운영 호출자가 없어 M2 와 독립) | M2 뒤, 즉시 가능 | 없음 |
| M3 | 없음(M1·M2·M2a 와 독립) | 즉시 가능 | 없음(옵션 B 는 운영자 결정으로 확정, D-22) |
| M4 | M2, t1448 착지 | 런북 2단계(가드 픽스처만 선행) | t1448 |
| M5 | M1, M2, M2a, M3 | 런북 2단계 | 없음 |
| M6 | M1~M5 준비 | 런북 2단계 이전(추가 파일) | 리더·운영자의 보류 단계 확인 |
| 절체 | M6 + 리더 확인 + origin/develop 팁 CI 녹색(D-17) | 배치 경계 | 운영자 |

단일 작성자 원칙에 따라 M1~M3(M2a 포함)은 직렬이다. 논리적 독립은 순서를 자유롭게 하는 것이지 병렬 write 를 허용하는 것이 아니다.

## §E 이 카드가 하는 일과 하지 않는 일

**하는 일**

- 코드·설정·스크립트·워크플로·문서·템플릿 변경을 카드 브랜치에서 작성하고 검증한다.
- PRE-CUTOVER-SAFE 변경을 현 경로로 develop 에 병합하도록 리더에게 인도한다.
- CUTOVER-TIME 묶음을 병합 가능 상태로 준비하고 보류한다.
- 런북 문서·점검 스크립트·스크래치 클론 리허설 증거를 만든다.
- 병합 큐 판정·t810 흡수·t1452 조정을 권고로 기록한다.

**하지 않는 일 (리더 확인 기록 없이는 금지)**

- main 보호 규칙·필수 체크 목록(`Release PR Multi-OS Gate` 제거 포함, **운영자가 직접 수행하는 단계**)·규칙셋·기본 브랜치·merge 설정의 변경.
- develop 의 보호 변경·삭제·이름 변경.
- 기준 브랜치 절체 실행(M5 묶음의 병합, develop push, 레인 재기동, `origin/main` 으로의 수렴 PR 개설과 병합).
- rc 또는 정식 태그 생성과 push.
- `moai constitution amend` 의 실제 실행(5단째 인간 승인이 대화형이므로 **운영자가 실행**, 레인은 입력과 `--dry-run` 제안을 준비하고 사후에 `validate` 로 검증한다).
- `SPEC-LATE-BRANCH-REDESIGN-001` 의 파일 수정과 상태 전환, t810·t1452·t1448 의 카드 본문 수정과 워크트리 처분.
- 휴면 git-flow 코드 삭제.

**확인 기록**: 위 금지 단계를 풀려면 리더의 확인이 `.moai/reports/t1453/cutover-confirmation.md` 에 단계 이름·시각이 아닌 **관측한 선행 조건**과 함께 남아야 한다. 선행 조건은 단계마다 다르다 — 런북 0·2단계(develop push)는 origin/develop 팁 CI 가 녹색이라는 관측(design D-17), 4단계 이후(수렴과 재기동)는 3단계 사전 점검의 출력(AC-GFD-019)이다. 이 파일이 없으면 해당 단계는 미실행으로 읽는다.

## §F 위험과 완화

| 위험 | 완화 |
|---|---|
| 보류 묶음이 develop 의 새 develop 기준 서술을 놓친다 | 절체 직전 마지막 흡수 뒤 스윕 가드 무장 실행(AC-GFD-011). 가드는 허용 목록 항목마다 이유를 요구해 조용한 허용을 막는다 |
| `strict: false` 에서 동시 카드의 의미 충돌이 main 을 붉게 만든다 | PR 개설 전 병합 준비 점검(AC-GFD-005), main push CI 가 사후 안전망, 병합 큐 판정은 측정 후속 카드(design §D-3) |
| 매트릭스 확인을 일찍 강제해 현행 릴리스를 막는다 | `--require-matrix-run` 옵션 뒤에 두고(기본 꺼짐) M5 가 하네스 본문에서 절체 시점에 켠다(M3, design D-24) |
| squash 병합 후 로컬 트리가 영원히 정리되지 않는다 | 착지 판정 세 층(AC-GFD-002) |
| 대형 수렴 PR 이 GitHub 한도에 걸린다 | 하나로 먼저 시도하고 폴백(분할 PR → 보호 일시 완화)을 런북에 두며 선택은 운영자·리더(design D-18) |
| develop 의 CI 가 적색인 채 절체한다 | 선행 조건으로 origin/develop 팁 CI 녹색을 요구한다(design D-17). 수리 카드는 리더 소관. M5 가 push 트리거를 빼지 않으므로 절체 내내 관측 가능하다 |
| develop 퇴역 뒤 push 트리거의 `develop` 이 조용히 죽은 설정으로 남는다 | 런북 9b 단계의 잔존 점검(`git grep -n -w develop -- .github/workflows`)과 CI 런의 전진 정지 확인(AC-GFD-023) |
| 불변 태그를 잘못 만든다 | 이 카드는 태그를 만들지 않는다. 게이트는 태그 전에 건다(design §D-7) |
| 레인이 절체 이전 기준으로 작업하다 낡은 기준에 착지한다 | 절체 전 레인 정지·`/clear`·정리, 절체 후 새 기준에서 재기동하는 절차(design §D-8) |
| 투영을 운영에 연결하는 후속 변경이 구성을 읽지 않고 리터럴 목표를 인자로 넘긴다 | 계약 코어가 `internal/config` 를 import 하지 않는 가드(원장 E-52)는 코어 쪽만 지키고 호출 쪽은 지키지 못한다. 연결하는 카드가 `config.LoadGitFlowIntegrationConfig(root).IntegrationTarget` 을 읽도록 design D-30 이 호출 규약을 적어 두었고, 그 연결의 시험은 그 카드의 몫이다(acceptance §E) |

## §G 인접 카드

### G.1 t810 — 흡수 권고

`SPEC-LATE-BRANCH-REDESIGN-001` 의 REQ-LBR-001..006 은 이 SPEC 의 M4 인수 기준(AC-GFD-013·017)과 일치한다. 같은 Frozen 조항(`CONST-V3R5-027`·`-028`)과 `[ZONE:Frozen]` 처분 줄을 `moai constitution amend` 로 개정하는 것이 공통 작업이다. 결정: **흡수한다** — 리더 결정(10-02 밤), design D-19. t810 의 카드·워크트리와 `SPEC-LATE-BRANCH-REDESIGN-001` 은 건드리지 않고, 큐 관계 기록이 서로를 흡수한다고 적혀 있던 모호함은 이 결정으로 방향이 정해졌다. t810 을 닫는 처분은 이 카드가 닫힐 때 운영자에게 올린다. 이 SPEC 이 그 SPEC 을 대체(`superseded`)로 표시하는 전환이 필요하면 그것은 manager-spec 이 별도 작업으로 하고 이 plan-phase 는 그 SPEC 의 파일을 수정하지 않았다. 또한 t810 카드 본문의 B2("primary feat/SPEC 산문 경로도 워크트리로")는 그 SPEC 의 REQ-LBR 여섯 어디에도 대응이 없다. M4 의 전달 경로 서술(REQ-GFD-012)이 해당 문장(`spec-workflow` Step 2/3, `delivery` Step 3.2)을 어차피 다시 쓰지만, t810 의 입력 보고서를 읽지 않아 범위 일치는 확인하지 못했다(design §D-10).

### G.2 t1452 — 좁힘 권고

| t1452 항목 | github-flow 에서 | 이유 |
|---|---|---|
| (a) 지명 없는 acquire | 무의미 | 로컬 병합 창이 카드 전달에서 은퇴한다 |
| (b) 창 획득 전 merge-tree 점검 | 무의미(창 기준) → PR 전 병합 준비 점검으로 이관 | 점검 자체는 AC-GFD-005 가 흡수 |
| (c) 동일 테스트 명령 재실행 억제 | 유효 | 카드 PR 체제에서도 같은 명령 재실행 낭비가 남는다 |

결정: 리더가 t1452 를 (c)와 "PR 전 병합 준비 점검"(sync 감사·merge-tree 충돌·트리 항등)으로 좁히기로 했다(design D-11). 카드 본문 수정은 리더의 몫이고 이 카드는 하지 않으며, 이 계획은 그 전제로 진행한다.

### G.3 t1448 — 순서 제약

t1448 은 `kanban-dispatch*.md`·`gtd.md`·`auto-semantics.md`·`manager-todo.md` 를 편집한다. M4 는 그 착지를 착수 증거로 요구하고 그 뒤에 흡수한다. 같은 파일을 동시에 편집하지 않는다.

## §H 후속 인벤토리 (범위 아님, 목록만)

휴면이 되는 항목 후보 — 읽기 결과가 입력 보고서에서 왔고 이 회차에 다시 읽지 않았다. 후속 카드가 확인한다.

- `internal/mission/integration.go` (비테스트 호출자 없음으로 보고됨)
- `lead_push_threshold`(타입·기본값·템플릿·인벤토리만 있고 소비자 없음으로 보고됨)
- `internal/spec/lint_movingref.go` 의 `origin/(main|develop|HEAD)` 정규식의 develop 대안
- `release-drafter-cleanup.yml`, `auto-merge.yml` 의 `release/*` 키
- 와이어 식별자 개명(`push_develop` 등) — 카드 발행은 리더 소관(design D-21)
- 계약 모드 에스컬레이션 분류기(`internal/escalation` 의 `pushWhy` 류)가 PR 전달의 카드 브랜치 push 를 허용하도록 하는 변경 — 새 행동 어휘와 git-flow 비영향 증명이 필요(design D-23)
- 워크플로 push 트리거의 `develop` 제거 — develop 삭제 뒤 정리 카드(design D-13 단계 4)

## §I 결정 이력 — 1회차의 미해소 질문 7건은 모두 결정으로 옮겼다

1회차 계획이 열어 둔 질문 일곱 항목은 plan-audit 의 MP-7 이 막았고, 모두 번호 붙은 결정으로 처분했다. 표지 문자열은 이 문서에 없다. 결정마다 출처(운영자·리더·계획 기본값)와 리더가 뒤집을 수 있는지가 `design.md` 의 같은 번호 행에 있다.

| 옛 번호 | 질문 | 결정 | 출처 | 막던 것의 현재 |
|---|---|---|---|---|
| NC-1 | rc 태그에서 검사 5·6 을 건너뛰는 대신 최소 대체 검사가 필요한가 | D-16 — 필요 없다(R-a: 5·6 건너뜀, 1~4·7 유지) | 계획 기본값(리더가 뒤집을 수 있다) | M3 의 rc 규칙이 확정됐다 |
| NC-2 | 절체 경계의 전제로 develop 팁 CI 녹색을 요구하는가 | D-17 — 요구한다. M5 는 push 트리거를 빼지 않는다 | 운영자 결정 | 절체 실행의 선행 조건(리더·운영자 소관) |
| NC-3 | 대형 수렴 PR 이 한도에 걸릴 때의 폴백 | D-18 — 하나로 먼저 시도, 분할 PR, 보호 일시 완화 순 | 계획 기본값(리더가 뒤집을 수 있다) | 폴백 선택은 절체 시점에 운영자·리더 |
| NC-4 | t810 과 t1453 의 흡수 방향 | D-19 — t1453 이 t810 을 흡수 | 리더 결정 | M4 의 AC-GFD-017 범위가 확정됐다 |
| NC-5 | 병합 큐 후속 측정의 채택 임계 | D-20 — 카드마다 PR, 병합 큐 보류. 측정은 후속 항목(Gaps) | 운영자 결정 | 이 SPEC 은 막히지 않는다 |
| NC-6 | 와이어 식별자 개명 후속 카드 | D-21 — 리더에게 하는 권고로 `progress.md` §G 에 기록, 카드 id 는 적지 않는다 | 계획 기본값 | 없음(스윕 가드의 별도 차선이 추적) |
| NC-7 | 필수 체크 `Release PR Multi-OS Gate` 의 처분 | D-22 — 필수 체크에서 뺀다(운영자가 수행), 태그 직전 `workflow_dispatch` | 운영자 결정 | 런북 9a 단계 |

---

🗿 MoAI
