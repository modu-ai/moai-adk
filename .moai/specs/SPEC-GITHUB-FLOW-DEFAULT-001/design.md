# SPEC-GITHUB-FLOW-DEFAULT-001 — 설계

> 이 문서는 결정(D-n)마다 **채택안, 근거, 기각된 대안, 기각 근거**를 적어 plan-auditor 가 반박할 수 있게 한다. 근거 수치는 `research.md` 와 `acceptance.md` §B 증거 원장에 있다. 구현 세부(함수명·파일 분할)는 run-phase 의 몫이며 여기서는 구현이 만족할 구조만 정한다.

## §A github-flow 에서 카드가 흐르는 길

| 단계 | 일어나는 일 | 주체 |
|---|---|---|
| 1 | 카드 워크트리를 `origin/main` 에서 만든다(`worktree_base_branch: main`), 브랜치는 `WT-<slug>` | 레인 |
| 2 | 카드 브랜치에 커밋(주 체크아웃에는 커밋하지 않는다 — 브랜치 가드 유지) | 레인 |
| 3 | `moai factory complete` 가 PR 개설 전 병합 준비 점검(sync 감사 PASS·통합 목표 대비 병합 충돌 없음·트리 항등)을 돌린다 | 레인 |
| 4 | 카드 브랜치 push, 대상 `main` 의 PR 개설, 자동 병합 요청 | 레인(CLI) |
| 5 | 필수 체크(`Test`·`Lint`·`Build`·`Analyze`·게이트)와 CodeRabbit 가 PR 헤드에서 재측정 | GitHub |
| 6 | 체크가 통과하면 GitHub 이 squash 병합하고 원격 카드 브랜치를 지운다(`delete_branch_on_merge: true`) | GitHub |
| 7 | 레인이 병합 완료를 관측해 카드를 기록하고, 착지 판정(조상 → 누적 patch-id → PR 병합 상태)이 맞으면 `worktree sweep` 이 트리를 처분한다 | 레인·CLI |
| 8 | main push CI 가 사후 안전망으로 한 번 더 돈다 | GitHub |

## §D-1 git-flow 는 기본값에서 은퇴하고 기능으로 남는다

**채택**: 배포 템플릿은 이미 github-flow 가 기본이다(원장 E-08 의 같은 파일에 있는 다른 두 프로필 값). 이 저장소만 manual 프로필에 git-flow 값을 든다. 코드가 `DevelopBranch` 를 읽는 지점(원장 E-03)은 manual+git-flow 일 때만 값이 채워지는 필드라 github-flow 구성에서 빈 기준을 본다. 통합 목표 해석기는 이미 해석표(`github-flow`→고정 `main`, `git-flow`→`develop_branch`, `gitlab-flow`→환경, `release-flow`→릴리스 접두)를 갖고 있으므로 읽기를 그 해석기로 옮기면 git-flow 구성의 동작은 같고 github-flow 구성은 `main` 이 된다.

**기각**
- *git-flow 코드를 모두 삭제*: 사용자 프로젝트가 고를 수 있는 워크플로를 없애는 범위 확장이다. 이 카드의 요구도 아니다.
- *설정 값만 바꾸고 코드는 그대로*: `DevelopBranch` 가 비는 순간 통합 창 획득은 호출자 브랜치로 되돌아가고, 세션 종료 병합은 건너뛰고, 병합 준비 점검과 `factory complete` 는 거부한다(입력 보고서의 읽기 결과). 조용한 오동작이다.
- *github-flow 를 환경변수로 추가 기능화*: 이미 있는 구성 값이 스위치이므로 새 스위치는 중복이다(단순화 사다리 2번).

**후속 인벤토리(범위 아님)**: `internal/mission/integration.go`, `lead_push_threshold`, `lint_movingref.go` 의 develop 정규식 대안, `release-drafter-cleanup.yml`·`auto-merge.yml` 의 `release/*` 키. 입력 보고서의 읽기 결과이며 이 회차에 다시 읽지 않았다.

## §D-2 두 변경 계층

**채택**: PRE-CUTOVER-SAFE 는 구성이 git-flow 일 때 출력이 같다는 특성화 시험이 병합 조건이다. CUTOVER-TIME 은 카드 브랜치에서 보류한다.

**기각**
- *한 번의 큰 절체 PR*: 코드 결함과 문서 결함이 한 번에 섞여 되돌리기 단위가 없다. 코드 변경은 먼저 검증된 채로 develop 에 있어야 한다.
- *런타임 스위치로 두 흐름을 병행*: 두 흐름의 문서와 설정이 동시에 참일 수 없으므로 문서가 거짓이 되는 구간이 길어진다.
- *문서를 먼저 바꾸고 코드를 나중에*: 일찍 병합한 서술이 거짓이다. 규칙 편집은 전 세션 프롬프트 캐시를 무효화한다는 카드의 제약과도 어긋난다.

## §D-3 병합 큐는 보류한다

**증거 표**

| 입력 | 측정 여부 | 값 |
|---|---|---|
| main 필수 체크 `strict` | 측정 | `false` (`research.md` §2) |
| 필수 체크 | 측정 | 5개 — Test(ubuntu)·Lint·Build(linux/amd64)·Analyze(Go)·Release PR Multi-OS Gate |
| CI 벽시계 | 측정(표본 한정) | develop push 최근 12회 중 failure 7회 1046~1551 초·cancelled 5회·success 0회. PR 이벤트 표본(2026-04-09~07-06) success 중앙값 173 초. main push success 평균 594 초(n=21) |
| 3-OS 통합 시험 | 측정(파일 읽기) | `ci.yml:392-400` `test-integration` 이 ubuntu·macos·windows 행렬 |
| 3-OS race 매트릭스 | 측정(파일 읽기) | `release-pr-multi-os.yml` 에서만, `release/*` 헤드 또는 수동 실행 |
| 호스티드 러너 청구 분 | 미측정 | 저장소는 공개(`private: false`), 청구 정책은 미확인 |
| CodeRabbit 한도 이력 | 미측정(입력 보고서 인용 3회 표본 "Review completed") | |
| 의미 충돌 빈도(녹색 PR 뒤 적색 main) | 미측정 | PR 체제 이전 값이 없다 |

**옵션**: Q-a 보류(현행 `strict: false` 유지, 사후 main push CI 가 안전망), Q-b `strict: true`(최신 main 위에서만 병합), Q-c 병합 큐.

**채택**: Q-a, 후속 측정 카드로 넘긴다. 근거: (1) `strict: false` 는 "PR 헤드를 병합 결과 위에서 시험하라"는 요구를 하지 않으므로, 병합 큐의 핵심 이익(병합 결과를 시험)은 현 보호 설정이 이미 포기한 바로 그 속성이다. 큐를 켜려면 보호 설정과 `merge_group` 트리거 변경이 함께 필요하다(운영자 소관). (2) 큐는 항목마다 CI 를 다시 돌린다. 카드 PR 체제는 이미 지금의 "develop push CI 한 번에 약 `lead_push_threshold`(20)장" 대비 카드마다 PR CI 한 번과 병합 뒤 main push CI 한 번을 요구한다. 큐는 거기에 더 얹는다. (3) 측정 입력 셋이 비어 있다.

**기각**: Q-b — 레인이 여럿 동시에 PR 을 열면 병합마다 나머지 PR 이 낡아 다시 돌려야 하므로 지금 비용을 측정하지 못한 채 도입하면 재시도 비용이 가장 크다. Q-c — 위 (1)(2)(3).

**리더가 뒤집을 수 있다**: 후속 카드가 위 미측정 입력을 채우고 임계(NC-5)를 정한 뒤 다시 판정한다. 이 SPEC 은 병합 큐 없이도 동작하는 PR 간선을 만든다.

## §D-4 카드 전달 상태 모델

**현황**: `merged-local` 이 카드 기록 모델에 있고(비테스트 Go 에서 `factory_card.go` 3·`mcp_factory_card.go` 1·`homestate/card_*.go` 6 지점의 참조), 전이 가드가 로컬 병합을 가정한다(입력 보고서의 읽기 결과).

**옵션**
- **S-a**: `pr-open` 과 병합 완료 상태 하나를 추가하고 `merged-local` 은 git-flow 용으로 둔다.
- **S-b**: `merging` 을 "PR 열림"으로, `merged-local` 을 양쪽 병합 완료로 겸용하고 병합 방식 필드를 둔다.
- **S-c**: `merged-local` 을 `merged` 로 개명한다.

**채택**: S-a. 근거: S-b 는 이름이 거짓이 된다("local" 이 PR 경로에서도 쓰임) — 큐 `done` 판정과 착지 점검이 그 이름을 읽는 곳에서 오독이 생긴다. S-c 는 이미 기록된 카드 레코드와 MCP 도구 스키마를 이주해야 한다. S-a 는 새 값을 추가하고 전이표와 MCP 표면을 늘리되 기존 기록은 읽힌다.

**수용 기준과의 연결**: 상태 이름은 run-phase 결정이며 AC-GFD-004 는 "병합 완료는 PR 병합이 관측된 뒤에만 기록"이라는 동작으로 판정한다. 시험 대역의 병합 관측 방법은 `gh pr view` 계열의 JSON 읽기다(대역 포함). 자동 병합은 `gh pr merge --auto` 로 요청한다(`allow_auto_merge: true`, 병합 방식은 `git_strategy.manual.merge_method: squash` — 저장소는 병합 커밋도 허용하지만 구성된 값을 따른다).

**병합 방식 기각 대안**: 병합 커밋 방식은 착지 판정을 조상 관계 하나로 만들지만 main 에 카드의 모든 WT 커밋이 들어와 이력이 길어진다. 구성된 값이 squash 이므로 이 SPEC 은 squash 를 안전하게 처리한다(§D-5).

## §D-5 착지 판정은 세 층이다

**관측**: 격리 저장소 실험(`research.md` §5)에서 커밋 2개 카드를 squash 병합한 뒤 `git merge-base --is-ancestor` 는 종료 코드 1, `git cherry` 는 두 커밋 모두 `+` 였고 병합 기준점에서의 누적 diff patch-id 는 squash 커밋의 patch-id 와 같았다. 따라서 `session_worktree.go` 의 `git cherry` 보조는 커밋 1개 카드만 구한다.

**채택(위에서 아래로 시도, 확정하면 멈춘다)**
1. **조상 관계** — `git merge-base --is-ancestor <카드 팁> <통합 ref>`. 병합 커밋 방식과 빨리감기를 구한다.
2. **누적 patch-id** — 카드 브랜치와 병합 기준점의 누적 diff 를 patch-id 로 만들어, 기준점 이후 통합 ref 의 커밋들의 patch-id 와 비교한다. 비교할 커밋 수에 상한을 두고 넘으면 다음 층으로 간다.
3. **PR 병합 상태** — 카드 브랜치를 헤드로 하는 병합된 PR 의 병합 커밋이 통합 ref 의 조상인지로 확인한다(`gh`). 네트워크나 토큰이 없으면 답하지 못한 것으로 읽는다.

어느 층도 확정하지 못하면 보존한다. 이는 `sweep` 이 이미 가진 세 갈래 계약(착지·미착지·답할 수 없음→보존)을 유지한다.

**기각**
- *병합 방식을 병합 커밋으로 고정*: 구성된 `merge_method: squash` 를 뒤집는 별도 결정이고 main 이력을 길게 만든다.
- *`git cherry` 만*: 위 관측에서 반증됐다.
- *PR 상태만*: 오프라인에서 영원히 보존하게 되고, 토큰 만료가 조용한 보존으로 보인다.

## §D-6 rc 태그의 출처 검증

**관측**: `release.yml` 의 검증 5(CHANGELOG 절)·6(`system.yaml` 버전)은 접미사를 구분하지 않고(원장 E-06), `scripts/release.sh` 의 검증 8 도 모든 버전에 CHANGELOG 절을 요구한다. 현재 rc 는 로컬·태그 없이 만들어 이 검사들을 지나지 않는다(합성 보고서 인용).

**옵션**
- **R-a**: 접미사(`-rc.N`) 태그는 검사 5·6 을 건너뛰고 1~4·7 을 유지한다.
- **R-b**: rc 에도 `[Unreleased]` 비어 있지 않음과 `system.yaml` 의 목표 기본 버전 일치를 요구한다.
- **R-c**: rc 마다 CHANGELOG 절과 `system.yaml` 을 바꾸는 커밋을 main 에 만든다.

**채택**: R-a. 근거: 검사 1~4(주석 태그·트레일러·트레일러 버전 일치·커밋 결속)와 7(main 조상)이 "누가 어느 커밋을 어떤 경로로 태그했는가"를 보증하고, 5·6 은 정식 버전 문서의 정합이라 rc 에는 의미가 없다. R-b 는 `[Unreleased]` 내용이 있는지 같은 약한 신호를 새 관문으로 만들어 규칙만 늘린다. R-c 는 rc 하나에 main 커밋과 PR 하나를 요구한다(main 은 PR 필수) — 가장 비싸다.

**테스트 가능성**: 인라인 워크플로 셸은 시험할 수 없으므로 출처 검증을 `scripts/` 의 독립 스크립트로 옮기고 워크플로가 호출하게 한다. 기각: 인라인 시험(불가능), `act` 같은 외부 실행기(새 의존).

**위험**: 이 저장소에는 원격에 없는 로컬 태그(`v3.1.0-rc.0`~`rc.2`, 합성 보고서 인용)가 있어 전체 태그 push 는 불변 태그를 만든다. 런북은 이름 붙인 태그 하나만 push 한다.

## §D-7 3-OS 매트릭스 게이트

**관측**: 3-OS race 매트릭스는 `release/*` 헤드나 수동 실행에서만 돈다. 필수 체크 `Release PR Multi-OS Gate` 는 `if: always()` 이고 detect 나 매트릭스가 failure 일 때만 실패하므로 비릴리스 PR 에서는 성공한다. `release/*` PR 이 없어지면 이 게이트는 영구 녹색 무동작이다. `v*` 태그는 만든 뒤 지우거나 옮길 수 없으므로(`research.md` §2) 검증은 태그 **앞**에 있어야 한다.

| 옵션 | 내용 | 비용·특성 |
|---|---|---|
| A | main 대상 모든 PR(Go 변경 시)에서 3-OS 매트릭스 | 가장 강하다. 카드마다 3-OS race(레그당 제한 시간 30)를 더해 카드 수에 비례. 필수 체크가 의미를 얻는다 |
| **B** | 태그 전에 대상 SHA 로 `workflow_dispatch` 매트릭스를 돌리고 `release.sh` 가 그 SHA 의 통과 기록이 없으면 태그를 거부 | 릴리스 시도당 1회. 수동 단계가 하나 늘고 릴리스 사이 main 의 mac/windows race 적색은 보이지 않을 수 있다(3-OS 통합 시험은 CI 마다 돈다) |
| C | `release.yml` 안, GoReleaser 앞에서 매트릭스 | 자동이지만 태그가 이미 존재한 뒤라 실패하면 태그가 소각된다 — 기각 |
| D | main push 마다 매트릭스 + `release.sh` 가 태그 대상 SHA 의 녹색을 확인 | 지속 가시성. 병합 카드마다 매트릭스(동시 병합은 ref 단위 취소로 합쳐진다 — main push CI 30회 중 8회가 cancelled) |

**채택**: B. 근거: 되돌릴 수 없는 행위(태그) 직전에만 비싼 검증을 둔다. D 는 업그레이드 경로로 남긴다(B 와 같은 `release.sh` 확인을 그대로 쓴다). **강제는 절체 시점에 켠다** — 현재 git-flow 릴리스는 main 병합 SHA 에 매트릭스를 돌리지 않으므로 지금 켜면 현행 정식 릴리스가 막힌다(M3). 옵션 B 이후 필수 체크 `Release PR Multi-OS Gate` 의 의미는 "비릴리스 PR 에서 무동작 성공"이다. 필수 목록에서 뺄지는 보호 변경이라 운영자 소관(NC-7).

**리더가 바꿀 수 있다**: A·D 로 바꿔도 `release.sh` 의 확인 로직은 같다.

## §D-8 main 과 develop 의 수렴

**채택**: (1) develop 이 `origin/main` 을 흡수(병합 커밋 — 이 두 팁에서 충돌 없음, 원장 E-23), (2) develop→main PR 을 **병합 커밋**으로 병합, (3) 병합 후 main 트리와 병합 직전 develop 팁 트리가 같음과 develop 팁이 main 의 조상임을 확인, (4) 되돌리기는 main 에 병합 커밋의 되돌리기 PR 을 올리는 것이다(force-push 불가). develop 은 1단계 동안 지우지 않으므로 되돌려도 다시 시도할 원본이 남는다.

**기각**
- *main force-push*: 보호 규칙이 막는다(`allow_force_pushes: false`, `enforce_admins: true`, `research.md` §2).
- *main 삭제 후 develop 을 main 으로 이름 변경*: main 은 삭제도 막혀 있고 보호 구성을 잃는다.
- *squash PR*: 트리는 같아지지만 조상 관계가 끊겨 이후 두 브랜치 사이의 병합이 영구히 충돌한다. 이 이유로 AC-GFD-018 이 조상 확인을 함께 요구한다.
- *API 를 통한 빨리감기 ref 갱신*: 보호 설정(PR 필수)이 막는 경로로 읽히지만 시도하지 않았다.

**폴백**: GitHub 가 7.7천 파일 규모의 PR 을 거부하면 develop 의 조상 커밋 구간을 순서대로 main 에 병합하는 분할 PR, 또는 운영자가 보호를 일시 완화하는 방법이 있다. 선택은 운영자(NC-3).

### 절체 순서 (런북 골격)

| # | 단계 | 주체 | 외부 공유 시스템 |
|---|---|---|---|
| 0 | 배치 경계 사전 점검(AC-GFD-019): 미푸시 0, 창·슬롯 보유자 없음, 미병합 picked 카드 없음, 활성 레인 정지. 점검 출력이 확인 기록의 일부 | 카드 스크립트, 리더 실행 | 아니오 |
| 1 | 레인 정지: 새 배차 중단, 각 레인이 진행 카드를 끝내고 증거를 hoist 하고 `/clear`, 세션 종료 | 리더 | 아니오 |
| 2 | M1~M3 와 M4·M5 묶음을 마지막으로 현 경로로 develop 에 병합하고 push. develop 팁 CI 판정 확인(NC-2) | 리더 | 예(develop push) |
| 3 | develop 이 `origin/main` 을 흡수 | 리더 | 아니오 |
| 4 | develop→main PR 개설·병합(병합 커밋) | **운영자·리더** | **예** |
| 5 | 트리 항등과 조상 확인 | 리더 | 아니오 |
| 6 | 레인 재기동: `moai cc -f N` 로 새 세션, 모든 새 카드 트리는 `origin/main` 에서, 이전 기준에서 만든 미착지 카드 브랜치는 main 이 develop 팁을 조상으로 가지므로 PR 대상만 바꿔 이어간다 | 리더 | 아니오 |
| 7 | 첫 카드가 PR→CI→병합→착지 판정→sweep 까지 가는지 관측 | 리더 | 예(PR) |
| 8 | develop 퇴역 단계화(§D-13) | 운영자 | 예 |

4 와 8 은 이 카드가 실행하지 않는 단계다(REQ-GFD-020).

## §D-9 스윕 가드의 설계

**위치**: `internal/template` 패키지. 이미 템플릿·규칙 가드(`evidence_citation_guard_test.go`·`rule_provenance_audit_test.go`)가 있고, 저장소 사본과 템플릿 사본을 모두 순회하는 구조(`allowEntry{File, Literal, Why}`, 방문 수 바닥값, 하위 트리 집합 동등 단언)를 재사용한다.

**범위 표면**: `AGENTS.md`·`AGENTS.local.md`·`CLAUDE.md`·`.claude/rules/**`·`.claude/agents/**`·`.claude/skills/**`·`.claude/output-styles/**`·README 4종·docs-site 4로케일·`.moai/docs/**`·템플릿 사본. 설정 값과 워크플로 트리거는 prose 스윕이 아니라 별도 키 단언(AC-GFD-016)으로 본다.

**패턴(RE2, 뒷보기 없음)**: 왼쪽 경계 `(?:^|[^A-Za-z0-9_./-])`, 오른쪽 경계 `(?:$|[^A-Za-z0-9_-])` 로 `develop` 을 둘러싼다. 하이픈·밑줄·점·슬래시 이웃은 식별자나 디렉터리명으로 보아 제외하고(`manager-develop`, `.claude/worktrees/develop`, `push-develop`, `--develop-worktree`), 비 ASCII(CJK) 이웃은 경계로 읽어 `develop에서` 를 잡는다. `\b` 는 쓰지 않는다 — 하이픈 옆에서도 경계가 서므로 `manager-develop` 을 잡아 버린다.
- **P-A(강)**: 백틱 `develop`·`origin/develop`, 또는 펜스 셸 줄에서의 `develop`. 위반.
- **P-B(강)**: `develop` 이 branch·브랜치·worktree·워크트리·base·기준·통합·integration·merge·병합·push 와 같은 줄에서 이웃. 위반.
- **P-C(약)**: 토큰 없는 살아 있는 문장 — `integration (worktree|window|branch)`·`통합 (워크트리|브랜치)`·`일괄 push`·`batch.?push`·`commit-dead`·`lead_push_threshold`·`\bWT-`. 위반이 아니라 **기준선 래칫**: 절체 묶음이 기록한 개수보다 늘면 실패하고 줄면 보고만 한다(`moai spec lint --baseline` 의 구조를 따른다).

**허용**
- 줄 표지: 같은 줄이나 둘러싼 인용 블록·제목의 `RETIRED`·`SUPERSEDED`·`폐기`·`은퇴`·`historical`·`legacy`.
- 스탬프: `develop @ <7~40자 hex>` 형태.
- git-flow 옵션 서술은 **줄 표지가 아니라 파일·절 단위 허용 항목**(`File`, `Literal`, `Why`)으로만 허용한다. `git-flow` 라는 단어가 있는 줄을 일괄 허용하면 "git-flow 라서 develop 에 병합한다"는 낡은 서술이 통과한다.
- 구조 식별자(`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree`)는 위 경계 때문에 본 패턴에서 이미 빠지고, **별도 차선**의 기대 파일 목록으로 추적해 새 파일로 번지면 실패한다(§D-12).

**무장과 연속 발화**: 가드 본체는 이 저장소의 `git-strategy.yaml` 의 활성 워크플로 값이 `github-flow` 일 때 무장한다. 절체 묶음이 그 값을 바꾸므로 값 변경과 가드 무장이 한 변경이다. 무장 전에는 `t.Skip` 이 아니라 "disarmed" 를 출력하는 별도 시험이 있고, AC-GFD-011·021 은 `--- SKIP` 이 없음을 요구한다. 방문 파일 수 바닥값과 하위 트리 집합 동등 단언은 순회가 조용히 줄어드는 것을 잡는다.

**기각**
- *로케일별 개수 동등 단언*: 현재 트리에서 이미 실패한다 — `moai-sync.md` 의 `develop` 단어 적중이 en 6·ja 5·ko 5·zh 5 이고 분류한 살아 있는 문장은 3/2/2/4 로 보고됐다. 의도를 단언한다.
- *내장 템플릿만 순회*: 저장소 사본(예: 원장 E-07 의 `AGENTS.md`)을 놓친다.
- *`git-flow` 줄 일괄 허용*: 위 이유.
- *Python 식 유니코드 `\w`*: CJK 가 단어 문자로 합쳐져 `develop에서` 를 놓친다. RE2 의 ASCII 클래스를 명시한다.

## §D-10 t810 은 흡수한다

**채택**: `SPEC-LATE-BRANCH-REDESIGN-001` 의 REQ-LBR-001..006 은 이 SPEC 의 AC-GFD-013·017 로 들어온다. 같은 Frozen 조항과 `[ZONE:Frozen]` 처분 줄을 amend 로 개정하는 일이 공통이므로 두 번 개정하지 않는다.

**기각**: *t810 을 먼저 별도로 끝내고 이 카드가 다시 개정*: Frozen 조항을 두 번 바꾸고 상시로드 규칙의 캐시를 두 번 무효화하며, t810 이 쓴 Route B 서술을 github-flow 가 곧 다시 쓴다.

**미해소**: t810 카드 본문에는 B2 로 "primary feat/SPEC 산문 경로도 워크트리로(spec-workflow Step 2/3·delivery Step 3.2 github-flow)"가 있지만 `SPEC-LATE-BRANCH-REDESIGN-001` 의 REQ 여섯에는 대응이 없다. M4 의 전달 경로 서술(REQ-GFD-012)이 그 문장들을 어차피 다시 쓰므로 덮이지만, t810 의 입력 보고서(`.moai/reports/t622/`)를 읽지 않아 범위 일치는 확인하지 못했다(Gap). 상태 전환과 방향은 리더 소관(NC-4).

## §D-11 t1452 는 좁힌다

**채택(권고)**: (a) 지명 없는 acquire 와 (b) 창 획득 전 merge-tree 점검은 로컬 창의 은퇴로 무의미해진다. (b) 의 점검 자체는 AC-GFD-005 의 PR 전 병합 준비 점검이 흡수한다. (c) 동일 테스트 명령 재실행 억제는 PR 체제에서도 남는다. t1452 를 (c)와 "PR 전 병합 준비 점검"으로 좁힌다.

**기각**: *t1452 를 그대로 진행*: 은퇴할 창을 더 빨리 만드는 작업이다. *t1452 를 닫고 (c)를 버림*: 같은 테스트 반복의 낭비는 PR 체제에서 카드마다 CI 를 돌릴 때 더 크다.

카드 본문 수정은 하지 않는다. 결정은 리더 소관이다.

## §D-12 구조 식별자는 이번에 개명하지 않는다

`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree` 는 설정 키·액션 어휘·CLI 플래그·기준선 스냅숏(`contract/testdata/mission_surface_baseline.txt`)에 걸쳐 호환 표면이다. 개명은 호환 이주를 요구하므로 후속 카드(NC-6)다. 같은 이유로 `push_serializer`·push 사전 점검은 이름을 `main` 으로 바꾸지 않는다 — 바꾸면 모든 main push 가 직렬화 대상이 된다. github-flow 에서 이 경로는 휴면이다. 스윕 가드는 이 식별자들을 별도 차선에서 기대 파일 목록으로 추적한다.

## §D-13 develop 퇴역은 단계로 한다

| 단계 | 상태 | 조건 |
|---|---|---|
| 0 | develop 이 기준 | 지금 |
| 1 | 수렴과 절체 후. develop 은 남아 있고 더 이상 커밋을 받지 않는다 | 되돌리기 원본 유지 |
| 2 | develop 을 읽기 전용으로 만들거나 보호(운영자) | 열린 PR 이 develop 을 대상으로 하지 않음(최근 병합 PR 20건 중 17건이 develop 대상이었으므로 재대상 지정이 필요하다) |
| 3 | develop 삭제(운영자) | `spec-lint.yml` 의 develop 의존 제거 확인(M5), 첫 카드 PR 이 main 까지 완주, main push CI 녹색, 옵션 B 의 `--dry-run` 관측 |

단계 순서는 시간이 아니라 조건이다. 삭제 시점은 리더 결정이다. `spec-lint.yml:70` 의 `git fetch origin main:… develop:…` 는 develop 이 사라지면 잡 전체를 실패시키므로 삭제보다 앞서 제거한다.

## §D-14 4로케일은 의도로 갱신한다

oss-docs 하네스 규칙에 따라 ko 정본을 먼저 쓰고 en, 그다음 ja·zh 를 같은 변경에서 갱신한다. zh 의 추가 문장(입력 보고서가 `moai-sync.md` 의 :105·:309 로 보고)은 M4 의 zh 담당이 살아 있는 서술인지 판정해 지우거나 폐기 표지를 붙인다. en 의 "Check out develop, pull, delete the local branch" 문장은 ko·ja 에 대응이 없어 의도 판정 대상이다. 가드는 긍정 앵커와 부정 스윕만 단언하고 로케일 간 개수는 단언하지 않는다(§D-9).

## §D-15 상시로드 예산

`AGENTS.md` 의 5줄(원장 E-07)과 규칙 서술 교체는 대체로 같은 길이의 교체라 순증가가 작을 것으로 예상하지만 이는 예측이지 측정이 아니다. 기준선(16개·285,543 바이트)을 M4 착수 시 다시 재고, 1,000 바이트를 넘는 단일 편집은 `rule-authoring.md` 의 진술(파일 바이트, 비호출 세션이 치르는 비용, `paths:` 로 옮길 수 있는지)을 변경 설명에 담는다. 새 절이 필요하면 `paths:` 범위의 동반 파일에 둔다.

---

🗿 MoAI
