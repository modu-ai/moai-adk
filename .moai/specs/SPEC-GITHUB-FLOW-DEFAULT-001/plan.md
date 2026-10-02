# SPEC-GITHUB-FLOW-DEFAULT-001 — 실행 계획

> 기준 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2`(카드 트리). 모든 좌표는 `acceptance.md` §B 증거 원장과 `research.md` 에서 온다. 시간 예측 없이 우선순위와 마일스톤 순서로만 서술한다.

## §A 전제

### A.1 두 계층과 병합 시점

| 계층 | 마일스톤 | 병합 경로 |
|---|---|---|
| PRE-CUTOVER-SAFE | M1, M2, M3, 스윕 가드의 픽스처 자가 시험(M4 의 일부) | 현 git-flow 경로로 develop 에 병합 — 이 경로를 마지막으로 쓴다 |
| CUTOVER-TIME | M4(가드 트리 단언 포함), M5 | 카드 브랜치에서 준비·검증하고 병합 가능 상태로 보류, 절체 경계에서 한 묶음으로 적용 |
| 절체 절차 | M6 | 문서·점검 스크립트·리허설 증거는 카드 산출물, 실행은 리더·운영자 |

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
| 2 | D-3 병합 큐 보류 | 사용자(리더)에게 보이는 흐름 결정, 측정 부족 | design §D-3 |
| 3 | D-7 3-OS 게이트 위치, D-6 rc 규칙 | 릴리스 흐름과 되돌릴 수 없는 태그에 닿는다 | design §D-7, §D-6 |
| 4 | D-8 수렴 방식, D-13 develop 퇴역 단계 | 외부 공유 시스템, 되돌리기 어려움 | design §D-8, §D-13 |
| 5 | D-5 착지 판정 세 층 | 기계적이지만 관측으로 반증된 가정 위에 있다 | design §D-5 |
| 6 | D-10 t810 흡수, D-11 t1452 좁힘 | 인접 카드와의 소유 경계 | design §D-10, §D-11 |
| 7 | D-9 스윕 가드 설계, D-12 식별자 동결 | 시험 설계 | design §D-9, §D-12 |
| 8 | 나머지 기계적 문서·미러 갱신 | 지침 반영 | M4, M5 |

## §C 마일스톤

각 마일스톤은 **착수 증거**(시작해도 되는 관측)와 **종료 증거**(끝났다고 말할 수 있는 관측)를 갖는다. 종료 증거는 `acceptance.md` 의 기준 ID 로 가리키며 판정서 `.moai/reports/t1453/verdict.md` 에 명령과 출력을 옮겨 적는다.

### M1 — PRE-CUTOVER-SAFE 코드·설정 재지정

**범위(RENAME 계열)**: 기준 브랜치 해석을 구성된 통합 목표로 옮긴다. 대상 후보는 증거 원장 E-01·E-02·E-03 의 지점과 입력 보고서의 RENAME 항목(`codex_review_scope` 의 카드 기준, 세션 정리의 통합 ref, `scripts/jev/triage.py` 기본값, `kanban` 착지 ref, 병합 목표 리터럴)이다. 이 목록은 단어 패턴 후보이며 항목마다 RED 를 쓰기 전에 읽어서 진입 여부를 정한다. 와이어 식별자는 건드리지 않는다(design §D-12).

- 착수 증거: 카드 트리에서 E-01·E-02·E-03 이 기록과 같은 출력을 낸다(원장 재실행).
- 순서: (1) git-flow 특성화 시험으로 현 출력을 고정 → (2) 해석표 전 행에 대한 표 기반 RED → (3) 읽기 교체 → (4) 두 구성에서 GREEN.
- 종료 증거: AC-GFD-001·003. E-01·E-02·E-03 명령이 종료 코드 1 과 빈 출력(해석기 파일 제외).
- 계층: PRE-CUTOVER-SAFE. git-flow 구성에서 출력이 같음이 병합 조건이다.
- 되돌리기: 커밋 되돌리기. 동작 변화가 github-flow 구성에만 있어 되돌려도 현행 흐름은 영향이 없다.

### M2 — github-flow 카드 전달 간선

**범위**: (a) 카드 상태 모델 결정(design §D-4) 후 `moai factory complete` 의 PR 간선(push, 대상이 통합 목표인 PR 개설, 병합 완료 관측 후 기록), (b) PR 개설 전 병합 준비 점검, (c) 착지 판정 세 층을 `worktree done`·`sweep`·세션 종료 정리가 공유하도록, (d) 카드 브랜치 push 가 `pushWhy` 류 에스컬레이션에 걸리지 않도록 허용, (e) github-flow 에서 창이 선행 조건이 아니게. `push_serializer`·push 사전 점검은 이름만 main 으로 바꾸면 모든 main push 를 직렬화하므로 github-flow 에서 휴면으로 둔다(design §D-12).

- 착수 증거: M1 병합 완료(해석기가 develop 에 있음), 원장 E-14·E-15·E-16 재실행이 기록과 같다.
- 종료 증거: AC-GFD-002·004·005·006, AC-GFD-022 의 후속 측정 카드 id 기록.
- 계층: PRE-CUTOVER-SAFE(github-flow 구성에서만 새 동작). 실제 gh 호출은 시험에서 대역으로 한다.
- 되돌리기: 커밋 되돌리기. 새 카드 상태 값이 기록 스키마에 들어가므로 되돌린 뒤 이미 기록된 새 상태가 남지 않았음을 확인한다(design §D-4).

### M3 — 릴리스 기구

**범위**: (a) 출처 검증 로직을 시험 가능한 스크립트로 옮겨 워크플로가 호출(design §D-6), (b) rc 규칙, (c) `.goreleaser.yml` 의 `release.prerelease: auto`, (d) `scripts/release.sh` 의 detached HEAD 허용과 rc 태그의 CHANGELOG 검증 완화, (e) 태그 전 매트릭스 확인(옵션 B)과 그 확인을 켜는 스위치(design §D-7), (f) `ci.yml` 의 `Race Test` 건너뛰기 조건 재검토 기록.

- 착수 증거: 원장 E-04·E-05·E-06·E-09 재실행이 기록과 같다. 입력: 리더의 옵션 B 확인(NC 아님 — 옵션 B 가 권고이며 리더가 바꿀 수 있음).
- 종료 증거: AC-GFD-007·008·009·010.
- 계층: PRE-CUTOVER-SAFE. 단, **매트릭스 확인의 강제는 절체 시점에 켠다** — 현 git-flow 릴리스는 `release/*` PR 헤드에서 매트릭스를 돌리고 main 병합 SHA 에는 돌리지 않으므로, 강제를 지금 켜면 현행 정식 릴리스가 막힌다. 코드는 스위치 뒤에 두고 켜는 일은 M5 의 절체 묶음이 한다.
- 되돌리기: 커밋 되돌리기. 태그는 불변 규칙셋이므로 **이 카드는 어떤 rc 태그도 만들지 않는다**(acceptance §E).

### M4 — CUTOVER-TIME 문서·지침 묶음 (t1448 착지 뒤)

**착수 증거(전부 필요)**: (1) t1448 이 develop 에 착지 — 착지 여부는 `git log` 의 카드 id 병합 흔적으로 읽고 읽은 결과를 판정서에 적는다, (2) M2 병합 완료(문서가 새 간선을 서술하므로), (3) 상시로드 기준선 재측정(research §10).

**범위**: `AGENTS.md`·`CLAUDE.md`·`.claude/rules/moai/**`(kanban-dispatch*·main-checkout-branch-guard*·worktree-integration*·spec-workflow·agent-common-protocol*·contract-autonomy 등)·`.claude/agents/moai/*`(manager-git·manager-lead·manager-docs 등 develop 서술이 있는 것)·`.claude/skills/**`·README 4종·docs-site 4로케일·템플릿 사본, Frozen 조항 정규 개정, t810 의 REQ-LBR-001..006 흡수, 스윕 가드 본체. 템플릿 우선 순환으로 로컬과 템플릿을 함께 바꾼다. `make build` 와 임베드 축 확인은 저장소 절차에 따르고 이 카드의 관측이 아니다(acceptance §E).

- 순서: (1) 가드 픽스처 시험(사전 병합 가능) → (2) Frozen 개정(amend 5단) → (3) 규칙·에이전트·스킬 갱신 → (4) README·docs-site(oss-docs 하네스의 ko 정본→en→ja/zh 순) → (5) 가드 트리 단언 무장 → (6) 상시로드 전후 측정.
- 종료 증거: AC-GFD-011·012·013·014·017·021. AC-GFD-015 의 두 축 초록.
- 계층: CUTOVER-TIME(가드 픽스처 시험만 PRE-CUTOVER-SAFE).
- 되돌리기: 보류 묶음이라 미병합 시 폐기, 병합 후에는 커밋 되돌리기. Frozen 개정은 되돌리려면 역방향 amend 가 필요하다.

### M5 — CUTOVER-TIME 설정·CI·로컬 지침 묶음

**범위**: 이 저장소의 `git-strategy.yaml`(워크플로·`develop_branch`·`worktree_base_branch`)·`workflow.yaml`(`deny_commits_on`·`push_develop` 의 값 처분)·워크플로 push 트리거 목록(ci·codeql·graph-freshness·judgment-first-consistency·lsel-leak-guard·template-neutrality-check·test-install·workflow-parse-guard·docs-i18n-check·spec-lint)·`.coderabbit.yaml`·`spec-lint.yml` 의 develop 의존(가져오기와 릴리스 스냅숏 논리)·릴리스 하네스(`hns-release-specialist` 재작성)·`AGENTS.local.md`·`.claude/rules/local/**`·`.moai/docs/*doctrine*`·M3 의 매트릭스 확인 스위치 켜기.

- 착수 증거: M1·M2·M3 병합 완료, M4 준비 완료.
- 종료 증거: AC-GFD-016, AC-GFD-011 의 가드가 무장 상태에서 녹색.
- 계층: CUTOVER-TIME. 이 묶음은 develop 이 main 으로 수렴하기 전에 develop 에 병합되는 마지막 변경이며 절체 경계에서 M4 와 함께 적용된다.
- 순서 제약: `spec-lint.yml` 의 develop 가져오기 제거는 develop 브랜치 삭제보다 앞선다(design §D-13).
- 되돌리기: 커밋 되돌리기. 설정 값은 되돌리면 git-flow 로 돌아가지만 이미 github-flow 기준으로 만든 카드 PR 은 영향 범위를 점검해야 한다.

### M6 — 절체 런북과 리허설

**범위**: 수렴 병합 절차, 트리 항등 검증, 되돌리기, 레인 재기동 절차, 배치 경계 사전 점검 스크립트, 절체 후 확인 목록, develop 퇴역 단계. 스크래치 클론에서 수렴과 되돌리기를 리허설한다. 리허설은 GitHub 를 건드리지 않는다.

- 착수 증거: M1~M5 준비 완료.
- 종료 증거: AC-GFD-018·019, AC-GFD-020 의 보호 상태 대조.
- 런북 구성은 design §D-8·§D-13. 각 단계에 실행 주체(카드 / 리더 / 운영자)와 외부 공유 시스템 여부를 표시한다.

## §D 의존·순서

| 마일스톤 | 선행 | 병합 시점 | 외부 의존 |
|---|---|---|---|
| M1 | 없음 | 즉시 가능 | 없음 |
| M2 | M1 | M1 뒤 | 없음 |
| M3 | 없음(M1·M2 와 독립) | 즉시 가능 | 리더의 옵션 B 확인 |
| M4 | M2, t1448 착지 | 절체 경계(가드 픽스처만 선행) | t1448 |
| M5 | M1, M2, M3 | 절체 경계 | 없음 |
| M6 | M1~M5 준비 | 절체 전 | 리더·운영자의 보류 단계 확인 |
| 절체 | M6 + 리더 확인 | 배치 경계 | 운영자 |

단일 작성자 원칙에 따라 M1~M3 은 직렬이다. 논리적 독립은 순서를 자유롭게 하는 것이지 병렬 write 를 허용하는 것이 아니다.

## §E 이 카드가 하는 일과 하지 않는 일

**하는 일**

- 코드·설정·스크립트·워크플로·문서·템플릿 변경을 카드 브랜치에서 작성하고 검증한다.
- PRE-CUTOVER-SAFE 변경을 현 경로로 develop 에 병합하도록 리더에게 인도한다.
- CUTOVER-TIME 묶음을 병합 가능 상태로 준비하고 보류한다.
- 런북 문서·점검 스크립트·스크래치 클론 리허설 증거를 만든다.
- 병합 큐 판정·t810 흡수·t1452 조정을 권고로 기록한다.

**하지 않는 일 (리더 확인 기록 없이는 금지)**

- main 보호 규칙·필수 체크 목록·규칙셋·기본 브랜치·merge 설정의 변경.
- develop 의 보호 변경·삭제·이름 변경.
- 기준 브랜치 절체 실행(M5 묶음의 병합, 레인 재기동, `origin/main` 으로의 수렴 PR 개설과 병합).
- rc 또는 정식 태그 생성과 push.
- `SPEC-LATE-BRANCH-REDESIGN-001` 의 파일 수정과 상태 전환, t1452·t1448 의 카드 본문 수정.
- 휴면 git-flow 코드 삭제.

**확인 기록**: 위 금지 단계를 풀려면 리더의 확인이 `.moai/reports/t1453/cutover-confirmation.md` 에 단계 이름·시각이 아닌 **관측한 선행 조건**(AC-GFD-019 의 점검 출력)과 함께 남아야 한다. 이 파일이 없으면 해당 단계는 미실행으로 읽는다.

## §F 위험과 완화

| 위험 | 완화 |
|---|---|
| 보류 묶음이 develop 의 새 develop 기준 서술을 놓친다 | 절체 직전 마지막 흡수 뒤 스윕 가드 무장 실행(AC-GFD-011). 가드는 허용 목록 항목마다 이유를 요구해 조용한 허용을 막는다 |
| `strict: false` 에서 동시 카드의 의미 충돌이 main 을 붉게 만든다 | PR 개설 전 병합 준비 점검(AC-GFD-005), main push CI 가 사후 안전망, 병합 큐 판정은 측정 후속 카드(design §D-3) |
| 매트릭스 확인을 일찍 강제해 현행 릴리스를 막는다 | 스위치 뒤에 두고 M5 가 절체 시점에 켠다(M3) |
| squash 병합 후 로컬 트리가 영원히 정리되지 않는다 | 착지 판정 세 층(AC-GFD-002) |
| 대형 수렴 PR 이 GitHub 한도에 걸린다 | 폴백(분할 PR, 보호 일시 완화)을 런북에 두고 선택은 운영자(NC-3) |
| 불변 태그를 잘못 만든다 | 이 카드는 태그를 만들지 않는다. 게이트는 태그 전에 건다(design §D-7) |
| 레인이 절체 이전 기준으로 작업하다 낡은 기준에 착지한다 | 절체 전 레인 정지·`/clear`·정리, 절체 후 새 기준에서 재기동하는 절차(design §D-8) |

## §G 인접 카드

### G.1 t810 — 흡수 권고

`SPEC-LATE-BRANCH-REDESIGN-001` 의 REQ-LBR-001..006 은 이 SPEC 의 M4 인수 기준(AC-GFD-013·017)과 일치한다. 같은 Frozen 조항(`CONST-V3R5-027`·`-028`)과 `[ZONE:Frozen]` 처분 줄을 `moai constitution amend` 로 개정하는 것이 공통 작업이다. 권고: **흡수한다.** 이 SPEC 이 그 SPEC 을 대체(`superseded`)하는 전환은 새 SPEC 작성 시 manager-spec 소관이지만, t810 에는 보존된 워크트리와 `picked` 카드가 있고 큐 관계 기록이 서로를 흡수한다고 적혀 있어 **방향 결정은 리더 소관**이다(NC-4). 이 plan-phase 는 그 SPEC 의 파일을 수정하지 않았다. 또한 t810 카드 본문의 B2("primary feat/SPEC 산문 경로도 워크트리로")는 그 SPEC 의 REQ-LBR 여섯 어디에도 대응이 없다. M4 의 전달 경로 서술(REQ-GFD-012)이 해당 문장(`spec-workflow` Step 2/3, `delivery` Step 3.2)을 어차피 다시 쓰지만, t810 의 입력 보고서를 읽지 않아 범위 일치는 확인하지 못했다(design §D-10).

### G.2 t1452 — 좁힘 권고

| t1452 항목 | github-flow 에서 | 이유 |
|---|---|---|
| (a) 지명 없는 acquire | 무의미 | 로컬 병합 창이 카드 전달에서 은퇴한다 |
| (b) 창 획득 전 merge-tree 점검 | 무의미(창 기준) → PR 전 병합 준비 점검으로 이관 | 점검 자체는 AC-GFD-005 가 흡수 |
| (c) 동일 테스트 명령 재실행 억제 | 유효 | 카드 PR 체제에서도 같은 명령 재실행 낭비가 남는다 |

권고: t1452 를 (c)와 "PR 전 병합 준비 점검"(sync 감사·merge-tree 충돌·트리 항등)으로 좁힌다. 카드 본문 수정은 하지 않는다. 결정은 리더 소관이다.

### G.3 t1448 — 순서 제약

t1448 은 `kanban-dispatch*.md`·`gtd.md`·`auto-semantics.md`·`manager-todo.md` 를 편집한다. M4 는 그 착지를 착수 증거로 요구하고 그 뒤에 흡수한다. 같은 파일을 동시에 편집하지 않는다.

## §H 후속 인벤토리 (범위 아님, 목록만)

휴면이 되는 항목 후보 — 읽기 결과가 입력 보고서에서 왔고 이 회차에 다시 읽지 않았다. 후속 카드가 확인한다.

- `internal/mission/integration.go` (비테스트 호출자 없음으로 보고됨)
- `lead_push_threshold`(타입·기본값·템플릿·인벤토리만 있고 소비자 없음으로 보고됨)
- `internal/spec/lint_movingref.go` 의 `origin/(main|develop|HEAD)` 정규식의 develop 대안
- `release-drafter-cleanup.yml`, `auto-merge.yml` 의 `release/*` 키
- 와이어 식별자 개명(`push_develop` 등)

## §I 미해소 질문

아래 항목은 모두 권고 기본값이 있고, 어느 것도 M1~M3 의 착수나 plan-audit 를 막지 않는다. 각 항목이 **무엇을 막는지**를 함께 적는다.

- [NEEDS CLARIFICATION: NC-1 — rc 태그에서 검사 5·6 을 건너뛰는 대신 넣을 최소 대체 검사가 필요한가. 권고 기본값: 필요 없음. 막는 것: M3 의 rc 규칙 확정(리더 확인 전 R-a 로 진행 가능)]
- [NEEDS CLARIFICATION: NC-2 — 절체 경계의 전제로 develop 팁 CI 녹색을 요구할 것인가(최근 12회 success 없음, research §4). 막는 것: 절체 실행(리더 소관)]
- [NEEDS CLARIFICATION: NC-3 — 대형 수렴 PR 이 GitHub 한도에 걸릴 때 분할 PR 폴백과 보호 일시 완화 중 무엇을 쓸 것인가(운영자). 막는 것: 절체 실행]
- [NEEDS CLARIFICATION: NC-4 — t810 과 t1453 의 상호 흡수 기록 정리와 `SPEC-LATE-BRANCH-REDESIGN-001` 의 `superseded` 전환 방향·소유(리더). 막는 것: M4 의 AC-GFD-017 범위 확정. 흡수하지 않기로 하면 AC-GFD-017 만 빠진다]
- [NEEDS CLARIFICATION: NC-5 — 병합 큐 후속 측정 카드의 채택 임계(리더). 막는 것: 후속 카드의 판정(이 SPEC 은 막히지 않음)]
- [NEEDS CLARIFICATION: NC-6 — 와이어 식별자 개명 후속 카드 발행 여부(리더). 막는 것: 없음(스윕 가드의 별도 차선이 그때까지 추적)]
- [NEEDS CLARIFICATION: NC-7 — 필수 체크 `Release PR Multi-OS Gate` 를 옵션 B 채택 뒤에도 필수 목록에 둘 것인가(보호 변경이라 운영자). 막는 것: develop 퇴역 단계 2 이후의 정리]

---

🗿 MoAI
