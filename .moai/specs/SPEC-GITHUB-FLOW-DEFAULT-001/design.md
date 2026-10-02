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

**채택**: 배포 템플릿은 이미 github-flow 가 기본이다(원장 E-30: `git-strategy.yaml.tmpl` 의 세 프로필이 모두 `workflow: github-flow`, `worktree_base_branch: ""`). 이 저장소만 manual 프로필에 git-flow 값을 든다(원장 E-08, 이 저장소의 파일). 코드가 `DevelopBranch` 를 읽는 지점은 manual+git-flow 일 때만 값이 채워지는 필드라 github-flow 구성에서 빈 기준을 본다. 카드 전달의 기준을 정하는 두 지점이 원장 E-28 이고, 나머지 세 지점(`integration.go:353`, `session_worktree_automerge.go:176,183`)은 github-flow 에서 휴면이 맞는 git-flow 전용 경로라 읽기를 옮기지 않는다(원장 E-03 은 맥락 행이다). 통합 목표 해석기는 이미 해석표(`github-flow`→고정 `main`, `git-flow`→`develop_branch`, `gitlab-flow`→환경, `release-flow`→릴리스 접두)를 갖고 있으므로 읽기를 그 해석기로 옮기면 git-flow 구성의 동작은 같고 github-flow 구성은 `main` 이 된다.

**기각**
- *git-flow 코드를 모두 삭제*: 사용자 프로젝트가 고를 수 있는 워크플로를 없애는 범위 확장이다. 이 카드의 요구도 아니다.
- *설정 값만 바꾸고 코드는 그대로*: `DevelopBranch` 가 비는 순간 통합 창 획득은 호출자 브랜치로 되돌아가고, 세션 종료 병합은 건너뛰고, 병합 준비 점검과 `factory complete` 는 거부한다(입력 보고서의 읽기 결과). 조용한 오동작이다.
- *세션 종료 병합을 통합 목표 해석기로 옮겨 E-03 을 비우기*: `session_worktree_automerge.go` 는 git-flow 에서만 도는 로컬 `git merge --no-ff` 경로이고(자기 주석: "inert when the project is not manual git-flow") 읽기를 구성된 통합 목표로 옮기면 github-flow 에서 main 으로의 로컬 병합이 켜져 REQ-GFD-004·006 과 주 체크아웃 규칙을 어긴다. 그래서 이 경로는 github-flow 에서 휴면이어야 하고 그 사실을 양성 시험으로 고정한다(AC-GFD-001).
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
| CI 벽시계 — develop push | 측정(2026-10-02, 같은 명령 3회 동일) | 최근 12회: failure 7회(1046·1223·1413·1484·1517·1551·1443 초, 평균 1382 초)·cancelled 5회·success 0회. `createdAt` 2026-10-01T13:38:32Z ~ 2026-10-02T09:27:17Z (원장 M-1) |
| CI 벽시계 — main push | 측정(창을 고정, 같은 명령 2회 동일) | `--created '2026-07-01..2026-10-02' --limit 200` 의 200회: success 146·cancelled 43·failure 10·startup_failure 1, success 평균 438 초. `createdAt` 2026-07-16T21:58:31Z ~ 2026-09-10T01:46:49Z (원장 M-2) |
| 병합 PR 의 대상 분포 | 측정(2회 동일) | 최근 병합 PR 20건: develop 17·main 3. `mergedAt` 2026-09-09T20:10:46Z ~ 2026-10-01T16:45:43Z (원장 M-3) |
| PR 이벤트 CI 표본 | 재현되지 않아 표에서 뺐다 | 1회차 기록(success 29·failure 28·cancelled 3, 평균 190 초)을 감사가 같은 명령으로 다시 돌리자 다른 값(46/8/6, 평균 470 초)이 나왔다. 창을 고정하지 않은 `--limit` 표본은 움직이는 데이터이며 `research.md` §9 Gaps 에 올린다 |
| 3-OS 통합 시험 | 측정(파일 읽기) | `ci.yml:392-400` `test-integration` 이 ubuntu·macos·windows 행렬 |
| 3-OS race 매트릭스 | 측정(파일 읽기) | `release-pr-multi-os.yml` 에서만, `release/*` 헤드 PR 또는 `workflow_dispatch` |
| 호스티드 러너 청구 분 | 미측정 | 저장소는 공개(`private: false`), 청구 정책은 미확인 — 후속 측정 항목(`research.md` §9) |
| CodeRabbit 한도 이력 | 미측정(입력 보고서 인용 3회 표본 "Review completed") | 후속 측정 항목 |
| 의미 충돌 빈도(녹색 PR 뒤 적색 main) | 미측정 | PR 체제 이전 값이 없다 — 후속 측정 항목 |

**옵션**: Q-a 보류(현행 `strict: false` 유지, 사후 main push CI 가 안전망), Q-b `strict: true`(최신 main 위에서만 병합), Q-c 병합 큐.

**채택**: Q-a, 후속 측정 항목으로 넘긴다. 출처: 운영자 결정(10-02 밤, 리더 경유) — 카드마다 PR 로 전달하고 병합 큐는 보류한다(D-20). 근거: (1) `strict: false` 는 "PR 헤드를 병합 결과 위에서 시험하라"는 요구를 하지 않으므로, 병합 큐의 핵심 이익(병합 결과를 시험)은 현 보호 설정이 이미 포기한 바로 그 속성이다. 큐를 켜려면 보호 설정과 `merge_group` 트리거 변경이 함께 필요하다(운영자 소관). (2) 큐는 항목마다 CI 를 다시 돌린다. 카드 PR 체제는 이미 지금의 "develop push CI 한 번에 약 `lead_push_threshold`(20)장" 대비 카드마다 PR CI 한 번과 병합 뒤 main push CI 한 번을 요구한다. 큐는 거기에 더 얹는다. (3) 측정 입력 셋이 비어 있다.

**기각**: Q-b — 레인이 여럿 동시에 PR 을 열면 병합마다 나머지 PR 이 낡아 다시 돌려야 하므로 지금 비용을 측정하지 못한 채 도입하면 재시도 비용이 가장 크다. Q-c — 위 (1)(2)(3).

**리더가 뒤집을 수 있다**: 후속 측정이 위 미측정 입력을 채우고 채택 임계를 정한 뒤 다시 판정한다(D-20). 이 SPEC 은 병합 큐 없이도 동작하는 PR 간선을 만든다.

## §D-4 카드 전달 상태 모델

**현황**: `merged-local` 이 카드 기록 모델에 있고(비테스트 Go 에서 `factory_card.go` 3·`mcp_factory_card.go` 1·`homestate/card_*.go` 6 지점의 참조), 전이 가드가 로컬 병합을 가정한다(입력 보고서의 읽기 결과).

**소비자 목록(S-a 의 새 상태가 만족해야 하는 지점, 이 회차에 읽음)**: ① `internal/homestate/card_picked.go:195` — 선행 카드가 `merged-local`·`pushed`·`ci-green`·`done` 중 하나여야 통과시키는 `case` 목록. 새 병합 완료 상태가 이 목록에 들어가야 PR 로 병합된 선행 카드가 뒤 카드를 막지 않는다. ② `card_picked.go:205` — 그렇지 않으면 `ErrPredecessorUnmerged`("has not reached merged-local"). 문구가 새 상태를 가리키도록 갱신돼야 한다. ③ `internal/homestate/card_transition.go:532-537` — 원격이 설정돼 있으면 `merged-local → done` 을 거부한다. 새 병합 완료 상태는 이 거부를 만나지 않는 경로로 `done` 에 닿아야 한다. 그 밖에 입력 보고서가 센 `factory_card.go` 3·`mcp_factory_card.go` 1·`homestate/card_*.go` 6 지점.

**옵션**
- **S-a**: `pr-open` 과 병합 완료 상태 하나를 추가하고 `merged-local` 은 git-flow 용으로 둔다.
- **S-b**: `merging` 을 "PR 열림"으로, `merged-local` 을 양쪽 병합 완료로 겸용하고 병합 방식 필드를 둔다.
- **S-c**: `merged-local` 을 `merged` 로 개명한다.

**채택**: S-a. 근거: S-b 는 이름이 거짓이 된다("local" 이 PR 경로에서도 쓰임) — 큐 `done` 판정과 착지 점검이 그 이름을 읽는 곳에서 오독이 생긴다. S-c 는 이미 기록된 카드 레코드와 MCP 도구 스키마를 이주해야 한다. S-a 는 새 값을 추가하고 전이표와 MCP 표면을 늘리되 기존 기록은 읽힌다.

**수용 기준과의 연결**: 상태 이름은 run-phase 결정이며 AC-GFD-004 는 "병합 완료는 PR 병합이 관측된 뒤에만 기록"이라는 동작으로 판정하고, 위 소비자 세 지점은 AC-GFD-004 의 하위 시험(선행 카드 통과·`done` 도달)으로 덮는다. 시험 대역의 병합 관측 방법은 `gh pr view` 계열의 JSON 읽기다(대역 포함). 자동 병합은 `gh pr merge --auto` 로 요청한다(`allow_auto_merge: true`, 병합 방식은 `git_strategy.manual.merge_method: squash` — 저장소는 병합 커밋도 허용하지만 구성된 값을 따른다).

**병합 방식 기각 대안**: 병합 커밋 방식은 착지 판정을 조상 관계 하나로 만들지만 main 에 카드의 모든 WT 커밋이 들어와 이력이 길어진다. 구성된 값이 squash 이므로 이 SPEC 은 squash 를 안전하게 처리한다(§D-5).

## §D-5 착지 판정은 세 층이다

**관측**: 격리 저장소 실험(`research.md` §5)에서 커밋 2개 카드를 squash 병합한 뒤 `git merge-base --is-ancestor` 는 종료 코드 1, `git cherry` 는 두 커밋 모두 `+` 였고 병합 기준점에서의 누적 diff patch-id 는 squash 커밋의 patch-id 와 같았다. 따라서 `session_worktree.go` 의 `git cherry` 보조는 커밋 1개 카드만 구한다.

**채택(위에서 아래로 시도, 확정하면 멈춘다)**
1. **조상 관계** — `git merge-base --is-ancestor <카드 팁> <통합 ref>`. 병합 커밋 방식과 빨리감기를 구한다.
2. **누적 patch-id** — 카드 브랜치와 병합 기준점의 누적 diff 를 patch-id 로 만들어, 기준점 이후 통합 ref 의 커밋들의 patch-id 와 비교한다. **비교할 커밋 수의 상한은 500 이다.** 근거: `git rev-list --count --first-parent --since=2026-09-25T00:00:00Z --until=2026-10-02T00:00:00Z origin/develop` 가 `248` 이었고(원장 M-6, 1주 동안 착지한 카드 병합 수), 500 은 그 약 2주분이다. 넘으면 이 층은 "답하지 못함"이고 다음 층으로 간다.
3. **PR 병합 상태** — `gh` 로 카드 브랜치를 헤드로 하는 PR 을 읽는다. 다음 셋이 모두 참일 때만 착지다: PR 상태가 MERGED, **PR 의 `headRefOid` 가 로컬 카드 팁과 같다**, 병합 커밋(`mergeCommit.oid`)이 통합 ref 의 조상이다. 로컬 팁이 `headRefOid` 와 다르거나 PR 병합 뒤에 로컬 커밋이 더 있으면 착지가 아니라 **보존**이다(병합된 PR 뒤에 연장된 브랜치를 지우면 미병합 작업이 사라진다). `gh` 는 한 번 호출에 **제한 시간 10 초**, 재시도 없이, 시간 초과·비 0 종료·토큰 부재는 모두 "답하지 못함 → 보존"(fail-closed)이다. `worktree sweep` 은 가능하면 병합된 PR 을 `gh pr list --state merged` 한 번으로 일괄 읽어 같은 제한을 적용한다.

어느 층도 확정하지 못하면 보존한다. 이는 `sweep` 이 이미 가진 세 갈래 계약(착지·미착지·답할 수 없음→보존)을 유지한다.

**어느 경로가 어느 층까지 도는가**: `moai worktree done` 과 `moai worktree sweep` 은 세 층을 모두 돈다. **세션 종료 정리는 1·2층까지만 돌고 3층(네트워크)을 돌지 않는다** — 세션 종료 경로는 REQ-WSS-304 가 "No network runs ... shared exit path stays cheap" 로 정한 대로 네트워크를 쓰지 않으며 이 SPEC 이 그 성질을 바꾸지 않는다. 1·2층이 확정하지 못하면 보존하고 그 트리는 다음 `sweep` 의 3층이 판정한다. 이 성질은 양성 시험으로 고정한다(AC-GFD-002: 세션 종료 경로에서 `gh` 호출 0).

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

**규칙문(R-a — 이 문장이 유일한 정의이고 REQ-GFD-008·P6·D-16 은 이를 가리킨다)**: 접미사(`-rc.N`)를 가진 태그에서 릴리스 출처 검증은 검사 5(CHANGELOG 절)와 검사 6(`system.yaml` 버전)을 **건너뛰고**, 검사 1~4 와 7 은 **유지**한다. 건너뛴 자리를 채우는 대체 검사는 두지 않는다. 접미사 없는 태그의 7개 검사는 바뀌지 않는다. `scripts/release.sh` 의 검증 8(CHANGELOG 절 존재)도 같은 접미사 조건으로 건너뛴다(접미사 없는 태그에는 그대로).

**채택**: R-a. 출처: 계획 기본값(NC-1 에 대한 권고, 리더가 뒤집을 수 있음 — D-16). 근거: 검사 1~4(주석 태그·트레일러·트레일러 버전 일치·커밋 결속)와 7(main 조상)이 "누가 어느 커밋을 어떤 경로로 태그했는가"를 보증하고, 5·6 은 정식 버전 문서의 정합이라 rc 에는 의미가 없다. R-b 는 `[Unreleased]` 내용이 있는지 같은 약한 신호를 새 관문으로 만들어 규칙만 늘린다. R-c 는 rc 하나에 main 커밋과 PR 하나를 요구한다(main 은 PR 필수) — 가장 비싸다.

**테스트 가능성**: 인라인 워크플로 셸은 시험할 수 없으므로 출처 검증을 `scripts/verify-release-provenance.sh`(새 파일) 로 옮기고 `release.yml` 이 그것을 호출하게 한다. 이 변경은 `release.yml` 을 건드리지만 그 워크플로는 태그 push 에서만 도는 파일이라 접미사 없는 태그의 동작이 같으므로 PRE-CUTOVER-SAFE 다(AC-GFD-016 의 사전 방향 목록에 `release.yml` 은 없다). 기각: 인라인 시험(불가능), `act` 같은 외부 실행기(새 의존). 시험은 `internal/template` 패키지에서 스크립트를 격리 저장소 픽스처에 대해 실행한다(`release_workflow_pipefail_test.go` 가 같은 패키지에서 `release.yml` 을 읽고 `bash` 를 실행하는 선례).

**위험**: 이 저장소에는 원격에 없는 로컬 태그(`v3.1.0-rc.0`~`rc.2`, 합성 보고서 인용)가 있어 전체 태그 push 는 불변 태그를 만든다. 런북은 이름 붙인 태그 하나만 push 한다.

## §D-7 3-OS 매트릭스 게이트

**관측**: 3-OS race 매트릭스는 `release/*` 헤드나 수동 실행에서만 돈다. 필수 체크 `Release PR Multi-OS Gate` 는 `if: always()` 이고 detect 나 매트릭스가 failure 일 때만 실패하므로 비릴리스 PR 에서는 성공한다. `release/*` PR 이 없어지면 이 게이트는 영구 녹색 무동작이다. `v*` 태그는 만든 뒤 지우거나 옮길 수 없으므로(`research.md` §2) 검증은 태그 **앞**에 있어야 한다.

| 옵션 | 내용 | 비용·특성 |
|---|---|---|
| A | main 대상 모든 PR(Go 변경 시)에서 3-OS 매트릭스 | 가장 강하다. 카드마다 3-OS race(레그당 제한 시간 30)를 더해 카드 수에 비례. 필수 체크가 의미를 얻는다 |
| **B** | 태그 전에 대상 SHA 로 `workflow_dispatch` 매트릭스를 돌리고 `release.sh` 가 그 SHA 의 통과 기록이 없으면 태그를 거부 | 릴리스 시도당 1회. 수동 단계가 하나 늘고 릴리스 사이 main 의 mac/windows race 적색은 보이지 않을 수 있다(3-OS 통합 시험은 CI 마다 돈다) |
| C | `release.yml` 안, GoReleaser 앞에서 매트릭스 | 자동이지만 태그가 이미 존재한 뒤라 실패하면 태그가 소각된다 — 기각 |
| D | main push 마다 매트릭스 + `release.sh` 가 태그 대상 SHA 의 녹색을 확인 | 지속 가시성. 병합 카드마다 매트릭스(동시 병합은 ref 단위 취소로 합쳐진다 — main push CI 30회 중 8회가 cancelled) |

**채택**: B. 출처: 운영자 결정(10-02 밤, 리더 경유). 근거: 되돌릴 수 없는 행위(태그) 직전에만 비싼 검증을 둔다. 태그 불변 규칙셋(`Release tag immutability (v*)`)이 있으므로 게이트는 태그 **앞**에 있다. D 는 업그레이드 경로로 남긴다(B 와 같은 `release.sh` 확인을 그대로 쓴다). **강제는 절체 시점에 켠다** — 현재 git-flow 릴리스는 main 병합 SHA 에 매트릭스를 돌리지 않으므로 지금 켜면 현행 정식 릴리스가 막힌다(M3). 켜는 방법은 D-24 의 `--require-matrix-run` 옵션이다.

**필수 체크 `Release PR Multi-OS Gate`(NC-7 의 해소, D-22)**: 옵션 B 아래서 이 체크는 비릴리스 PR 에서 무동작 성공이므로 main 의 필수 체크 목록에서 **뺀다**(운영자 결정). 이 변경은 외부 공유 시스템의 되돌리기 어려운 변경이라 **운영자가 직접 수행하는 런북 단계**이며 이 카드의 레인은 실행하지 않는다(REQ-GFD-020). 순서: M5 의 스위치 켜기가 병합되고 AC-GFD-010 의 두 스위치 상태 시험이 통과한 뒤(런북 9단계). 체크를 빼면 `release/*` 헤드 PR 경로의 자동 매트릭스는 더 이상 병합을 막지 않는다 — 그 자리를 태그 직전 확인이 대신한다.

**리더가 바꿀 수 있다**: A·D 로 바꿔도 `release.sh` 의 확인 로직은 같다.

## §D-8 main 과 develop 의 수렴

**채택**: (1) develop 이 `origin/main` 을 흡수(병합 커밋 — 이 두 팁에서 충돌 없음, 원장 E-23), (2) develop→main PR 을 **병합 커밋**으로 병합, (3) 병합 후 main 트리와 병합 직전 develop 팁 트리가 같음과 develop 팁이 main 의 조상임을 확인, (4) 되돌리기는 main 에 병합 커밋의 되돌리기 PR 을 올리는 것이다(force-push 불가). develop 은 1단계 동안 지우지 않으므로 되돌려도 다시 시도할 원본이 남는다.

**기각**
- *main force-push*: 보호 규칙이 막는다(`allow_force_pushes: false`, `enforce_admins: true`, `research.md` §2).
- *main 삭제 후 develop 을 main 으로 이름 변경*: main 은 삭제도 막혀 있고 보호 구성을 잃는다.
- *squash PR*: 트리는 같아지지만 조상 관계가 끊겨 이후 두 브랜치 사이의 병합이 영구히 충돌한다. 이 이유로 AC-GFD-018 이 조상 확인을 함께 요구한다.
- *API 를 통한 빨리감기 ref 갱신*: 보호 설정(PR 필수)이 막는 경로로 읽히지만 시도하지 않았다.

**폴백(NC-3 의 해소, D-18)**: 수렴 PR 은 **하나로 먼저 시도한다**. GitHub 가 7.7천 파일 규모의 PR 을 거부하거나 병합하지 못하면 (a) develop 의 조상 커밋 구간을 순서대로 main 에 병합하는 분할 PR, (b) 운영자가 보호를 일시 완화하는 방법 순으로 간다. 어느 폴백을 쓸지는 절체 시점의 운영자·리더 판단이며 이 카드의 레인은 고르지 않는다. 출처: 계획 기본값(리더가 뒤집을 수 있음).

### 절체 순서 (런북 골격)

선행 조건은 운영자 결정(D-17)이다 — 리더가 미푸시 로컬 develop 커밋을 push 하고, 적색 CI 는 리더가 정하는 수리로 녹색이 되고(수리 카드는 이 카드의 범위가 아니다, 측정된 적색 집합은 `.moai/reports/t1453/m0-develop-ci-red.md`), 그 뒤에야 절체가 진행된다. develop 의 push CI 는 절체 내내 계속 돈다(M5 는 push 트리거에서 `develop` 을 빼지 않는다).

| # | 단계 | 주체 | 외부 공유 시스템 |
|---|---|---|---|
| 0 | 선행 조건 확인(D-17): 리더가 미푸시 로컬 develop 커밋을 일괄 push 하고 적색 CI 수리가 끝나 origin/develop 팁의 CI 가 success 인지 `gh run list --workflow=CI --branch develop --limit 3` 로 관측해 확인 기록에 적는다 | 리더 | 예(develop push) |
| 1 | 레인 정지: 새 배차 중단, 각 레인이 진행 카드를 끝내고 증거를 hoist 하고 `/clear`, 세션 종료, 워크트리 정리(착지 전 폐기 금지) | 리더(레인 각자) | 아니오 |
| 2 | M4·M5 묶음을 현 경로로 develop 에 마지막으로 병합(M1~M3 는 이미 병합돼 있다)하고 push. develop 팁 CI 를 관측한다(push 트리거가 살아 있으므로 관측 가능) | 리더 | 예(develop push) |
| 3 | 배치 경계 사전 점검(AC-GFD-019, 2단계 병합·push 뒤): 미푸시 0, 창·슬롯 보유자 없음, 병합·push 되지 않은 picked 카드 없음(이 카드는 2단계 병합 뒤라 제외 — D-25), 활성 레인 세션 없음. 점검 출력이 확인 기록의 일부 | 카드 스크립트, 리더 실행 | 아니오 |
| 4 | develop 이 `origin/main` 을 흡수하고 push | 리더 | 예(develop push) |
| 5 | develop→main PR 개설·병합(병합 커밋). 이 PR 의 필수 체크(`pull_request: branches: [main]`, `ci.yml:20`)가 두 번째 CI 증거다 | **운영자·리더** | **예** |
| 6 | 트리 항등과 조상 확인 | 리더 | 아니오 |
| 7 | 레인 재기동: `moai cc -f N` 로 새 세션, 모든 새 카드 트리는 `origin/main` 에서. 3단계가 병합·push 되지 않은 picked 카드 0 을 보증했으므로 이전 기준에서 만든 미착지 카드 브랜치는 없다 | 리더 | 아니오 |
| 8 | 첫 카드가 PR→CI→병합→착지 판정→sweep 까지 가는지 관측 | 리더 | 예(PR) |
| 9 | develop 퇴역 단계화(§D-13): 단계 2 에서 `Release PR Multi-OS Gate` 를 필수 체크에서 빼고(D-22), 삭제 뒤에 워크플로 push 트리거의 `develop` 을 정리하고 잔존 점검을 돌린다 | 운영자(체크 제거·develop 보호·삭제), 리더(정리 카드·점검) | 예 |

5 와 9 의 운영자 단계, 0·2·4 의 develop push 는 이 카드가 실행하지 않는 단계다(REQ-GFD-020). 이 표가 런북의 골격이고 M6 의 런북 문서는 같은 순서·같은 두 열(주체, 외부 공유 시스템 여부)을 가진다(AC-GFD-023).

## §D-9 스윕 가드의 설계

**위치**: `internal/template` 패키지. 이미 템플릿·규칙 가드(`evidence_citation_guard_test.go`·`rule_provenance_audit_test.go`)가 있고, 저장소 사본과 템플릿 사본을 모두 순회하는 구조(`allowEntry{File, Literal, Why}`, 방문 수 바닥값, 하위 트리 집합 동등 단언)를 재사용한다.

**범위 표면(하위 트리 열 곳, 이 트리 `855563dba…` 에서 `git ls-files` 로 센 파일 수 — 원장 M-4)**

| # | 하위 트리(추적 파일) | 파일 수 | 바닥값(⌊0.9×수⌋) |
|---|---|---|---|
| 1 | `AGENTS.md`·`AGENTS.local.md`·`CLAUDE.md`·README 4종 | 7 | 6 |
| 2 | `.claude/rules/**/*.md` | 113 | 101 |
| 3 | `.claude/agents/**/*.md` | 22 | 19 |
| 4 | `.claude/skills/**/*.md` | 255 | 229 |
| 5 | `.claude/output-styles/**/*.md` | 3 | 2 |
| 6 | `docs-site/content/**/*.md` (4로케일) | 620 | 558 |
| 7 | `.moai/docs/**/*.md` | 37 | 33 |
| 8 | `internal/template/templates/**/*.md`·`*.md.tmpl` (템플릿 사본) | 408 | 367 |
| 9 | `.moai/project/codemaps/*.md` | 6 | 5 |
| 10 | `scripts/**/*.sh`·`scripts/**/*.py` | 54 | 48 |
| | 전체(중복 제거) | **1525** | **1372** |

하위 트리 9번(codemaps 6파일 중 5파일에 `develop` 이 있다)과 10번(`scripts` 54파일 중 4파일에 `develop` 이 있다)은 1회차 감사가 빠진 면으로 지적한 곳이며 이제 범위에 든다. 가드의 방문 수 단언은 **전체 ≥ 1372 이고 하위 트리마다 위 바닥값 이상**이다. 바닥값 숫자는 이 계획 시점의 측정에서 ⌊0.9×수⌋ 로 도출한 참조값이며, 가드 시험 파일의 리터럴은 M4 착수 시 같은 명령으로 다시 센 값으로 갱신한다(그 값과 이 표가 다르면 판정서에 차이를 적는다). 설정 값과 워크플로 트리거는 prose 스윕이 아니라 별도 키 단언(AC-GFD-016)으로 본다. `.moai/reports/**`·`.moai/specs/**` 는 범위 밖이다(역사 기록과 SPEC 본문).

**핵심 패턴(RE2, 뒷보기 없음 — 이 회차 스크래치 프로브로 픽스처 21행과 `AGENTS.md` 를 돌려 관측했다, 원장 M-5)**

```
(?:^|[^A-Za-z0-9_./-])((?:refs/(?:remotes|heads)/)?(?:origin/)?develop)(?:$|[^A-Za-z0-9_-]|-(?:based|branch|tip|line)\b)
```

- 왼쪽 경계는 `-`·`_`·`.`·`/`·영숫자 이웃을 식별자나 디렉터리명으로 보아 제외한다(`manager-develop`, `.claude/worktrees/develop`, `push-develop`, `--develop-worktree`). 1회차 설계는 `/` 이웃을 제외하는 바람에 `git switch -c x origin/develop` 같은 맨 `origin/develop` 산문을 놓쳤다. 이제 `origin/` 과 `refs/remotes/`·`refs/heads/` 접두를 패턴 **안**에 두고, 접두 앞의 경계만 검사한다.
- 오른쪽 경계는 영숫자·`_`·`-` 이웃을 식별자로 보아 제외하되, 뒤의 `-based`·`-branch`·`-tip`·`-line` 은 서술이므로 잡는다. 1회차 설계는 `develop-based worktrees`(`AGENTS.md:76`, 원장 E-07 의 한 줄)를 놓쳤다 — 프로브가 `AGENTS.md` 에서 E-07 의 5줄을 모두 잡는다.
- 비 ASCII(CJK) 이웃은 경계로 읽어 `develop에서`·`develop을` 을 잡는다. `\b` 는 쓰지 않는다 — 하이픈 옆에서도 경계가 서므로 `manager-develop` 을 잡아 버린다.

**위반 판정(핵심 패턴이 적중한 뒤, 위에서 아래로 첫 일치)**

0. **제외** — 스탬프(`develop @ <7~40자 hex>`), 동사 용례(적중 앞이 `to` 이고 뒤가 관사·한정사 `a|an|the|your|our|their|this|that|new|more|it|them|and|software|feature(s)|code`)는 위반이 아니다. 동사 제외는 **`origin/` 접두가 붙은 적중에는 적용하지 않고**, `push to develop` 처럼 뒤가 한정사가 아닌 `to develop` 은 제외하지 않는다.
1. **P-A(강)** — 적중이 `origin/`·`refs/` 접두를 가졌거나, 백틱으로 둘러싸였거나, 펜스 코드 블록 안 줄이다. 위반.
2. **P-D(강)** — 적중에 CJK 문자(한글·한자·가나)가 바로 이웃한다. 영어 동사 `develop` 은 CJK 와 붙어 쓰이지 않으므로 이웃한 `develop` 은 브랜치 이름이다. 위반. (1회차 설계는 이를 P-B 토큰 목록에만 맡겼고 `분기` 가 목록에 없어 `develop에서 분기한다` 를 `to develop a feature` 와 구별하지 못했다.)
3. **P-B(강)** — 같은 줄에 브랜치 말 토큰 `branch|브랜치|worktree|워크트리|base|기준|통합|integration|merge|병합|push|분기|fork|check ?out|체크아웃|rebase|리베이스|pull|rev-list` 가 있다. 위반. (`분기`·`fork`·`check out`·`pull` 은 입력 보고서 §5 의 증거 — `develop에서` 와 "Check out develop, pull, delete the local branch" — 에서 더했다.)
4. 그 밖의 맨 적중은 **약한 적중**이다 — 위반이 아니라 방문 보고에 개수로 남긴다(동사 `develop` 의 거짓 양성을 피한다).

**P-C(약, 기준선 래칫)**: 토큰 없는 살아 있는 문장 — `integration (worktree|window|branch)`·`통합 (워크트리|브랜치)`·`일괄 push`·`batch.?push`·`commit-dead`·`lead_push_threshold`. 위반이 아니라 절체 묶음이 기록한 개수보다 늘면 실패하고 줄면 보고만 한다(`moai spec lint --baseline` 의 구조를 따른다). **`\bWT-` 는 뺐다** — 카드 브랜치 `WT-<slug>` 는 github-flow 에서도 살아 있는 이름이다(§A 1단계).

**허용(조용한 허용을 막기 위해 모두 좁혔다)**
- 줄 표지: 같은 줄이 `RETIRED`·`SUPERSEDED`·`폐기`·`은퇴` 중 하나와 **날짜(`20YY-MM-DD`)나 카드·SPEC 식별자(`t1234`·`SPEC-…`)를 함께** 가질 때. `historical`·`legacy` 는 표지로 인정하지 않는다(실측 프로브: `(legacy)` 를 붙인 살아 있는 문장은 위반으로 남는다). 제목 줄이 표지를 가지면 그 효력은 **그 절 자신**, 곧 하위 제목을 포함해 같은 수준이거나 더 상위 수준인 다음 제목 전까지에 한정된다(프로브 F-red-10·F-green-8). 인용 블록은 그 블록 하나에만 효력이 있다.
- 스탬프: `develop @ <7~40자 hex>` 형태.
- git-flow 옵션 서술은 **줄 표지가 아니라 파일·줄 단위 허용 항목** `{File, Literal, Why}` 으로만 허용한다. `Literal` 은 **그 파일의 한 줄 전체(앞뒤 공백 제거 뒤 동일)** 여야 하고, 파일에 그런 줄이 없으면 낡은 항목으로 실패한다. 맨 토큰(`develop`)은 줄 전체가 아니므로 거부된다(1회차 설계는 `{File: "AGENTS.md", Literal: "develop"}` 한 항목이 파일 전체를 침묵시키는 변이를 막지 못했다). **허용 항목은 전체 40개 이하**이다. 근거: 입력 보고서 §5 가 문서화된 git-flow 옵션(class N)으로 든 곳은 `delivery.md`·`moai-ref-git-workflow`·`manager-git.md`·`spec-workflow.md:60` 과 CI 샘플 줄 몇 곳이고, 40 은 그 줄 수의 두 배를 넘겨 잡은 여유다(이 회차에 줄 수를 다시 세지는 않았다 — Gap). 상향은 이 문서의 결정 행이 필요하다. `git-flow` 라는 단어가 있는 줄을 일괄 허용하지 않는다.
- 구조 식별자(`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree`)는 위 경계 때문에 본 패턴에서 이미 빠지고, **별도 차선**의 기대 파일 목록으로 추적해 새 파일로 번지면 실패한다(§D-12).
- 표지가 줄 단위로 면제한 줄 수는 파일별로 방문 보고에 센다(표지 추가로 살아 있는 문장을 덮는 변이는 막지 못하고 개수 증가를 눈에 띄게 할 뿐이다 — Residual-risk, 리뷰 소관).

**무장과 연속 발화**: 가드 본체는 이 저장소의 `git-strategy.yaml` 의 활성 워크플로 값이 `github-flow` 일 때 무장한다. 절체 묶음이 그 값을 바꾸므로 값 변경과 가드 무장이 한 변경이다. 무장 전에는 `t.Skip` 이 아니라 "disarmed" 를 출력하는 별도 시험이 있고, AC-GFD-011·021 은 `--- SKIP` 이 없음을 요구한다. 방문 파일 수 바닥값(전체 1372, 하위 트리마다 위 표의 값)과 하위 트리 집합 동등 단언은 순회가 조용히 줄어드는 것을 잡는다. 가드는 `visited=<N>`, 하위 트리별 방문 수, 약한 적중 수, 표지 면제 줄 수를 출력한다.

**퇴역 뒤의 연속 발화(별도 점검)**: develop 이 퇴역한 뒤에도 워크플로 push 트리거에 `develop` 이 남으면 그 목록은 조용히 죽은 설정이다(실행되지 않는 점검은 성공과 구별되지 않는다). 런북 9단계의 점검이 이를 잡는다 — `git grep -n -w develop -- .github/workflows` 가 종료 코드 1 이어야 하고(원장 M-7 은 지금 10개 파일이 적중함을 보인다), `gh run list --workflow=CI --branch develop --limit 3` 의 가장 새 `createdAt` 이 더 이상 전진하지 않아야 한다. 이 점검 문구는 런북 문서의 필수 내용이다(AC-GFD-023).

**기각**
- *로케일별 개수 동등 단언*: 현재 트리에서 이미 실패한다 — `moai-sync.md` 의 `develop` 단어 적중이 en 6·ja 5·ko 5·zh 5 이고 분류한 살아 있는 문장은 3/2/2/4 로 보고됐다. 의도를 단언한다.
- *내장 템플릿만 순회*: 저장소 사본(예: 원장 E-07 의 `AGENTS.md`)을 놓친다.
- *`git-flow` 줄 일괄 허용*: 위 이유.
- *Python 식 유니코드 `\w`*: CJK 가 단어 문자로 합쳐져 `develop에서` 를 놓친다. RE2 의 ASCII 클래스를 명시한다.
- *줄 표지에 `historical`·`legacy` 인정, 날짜·식별자 없는 표지*: 프로브에서 `(legacy)` 한 단어가 살아 있는 문장을 면제하는 변이를 막지 못했다.
- *`Literal` 이 맨 토큰이어도 허용*: 항목 하나가 파일 전체를 침묵시킨다.

## §D-10 t810 은 흡수한다

**채택**: `SPEC-LATE-BRANCH-REDESIGN-001` 의 REQ-LBR-001..006 은 이 SPEC 의 AC-GFD-013·017 로 들어온다. 같은 Frozen 조항(등재된 `CONST-V3R5-027`·`-028`)을 amend 로 개정하는 일이 공통이므로 두 번 개정하지 않는다. 출처: 리더 결정(10-02 밤) — t1453 이 t810 을 흡수한다(D-19).

**기각**: *t810 을 먼저 별도로 끝내고 이 카드가 다시 개정*: Frozen 조항을 두 번 바꾸고 상시로드 규칙의 캐시를 두 번 무효화하며, t810 이 쓴 Route B 서술을 github-flow 가 곧 다시 쓴다.

**남은 Gap**: t810 카드 본문에는 B2 로 "primary feat/SPEC 산문 경로도 워크트리로(spec-workflow Step 2/3·delivery Step 3.2 github-flow)"가 있지만 `SPEC-LATE-BRANCH-REDESIGN-001` 의 REQ 여섯에는 대응이 없다. M4 의 전달 경로 서술(REQ-GFD-012)이 그 문장들을 어차피 다시 쓰므로 덮이지만, t810 의 입력 보고서(`.moai/reports/t622/`)를 읽지 않아 범위 일치는 확인하지 못했다. t810 의 카드·워크트리·SPEC 파일은 건드리지 않고, 닫는 처분은 이 카드가 닫힐 때 운영자에게 올린다.

## §D-11 t1452 는 좁힌다

**채택**: 출처: 리더 결정(10-02 밤) — t1452 를 (c) 와 PR 전 병합 준비 점검으로 좁힌다. 카드 본문 수정은 리더의 몫이고 이 SPEC 은 그 전제로 진행한다. (a) 지명 없는 acquire 와 (b) 창 획득 전 merge-tree 점검은 로컬 창의 은퇴로 무의미해진다. (b) 의 점검 자체는 AC-GFD-005 의 PR 전 병합 준비 점검이 흡수한다. (c) 동일 테스트 명령 재실행 억제는 PR 체제에서도 남는다. t1452 를 (c)와 "PR 전 병합 준비 점검"으로 좁힌다.

**기각**: *t1452 를 그대로 진행*: 은퇴할 창을 더 빨리 만드는 작업이다. *t1452 를 닫고 (c)를 버림*: 같은 테스트 반복의 낭비는 PR 체제에서 카드마다 CI 를 돌릴 때 더 크다.

카드 본문 수정은 이 카드가 하지 않는다.

## §D-12 구조 식별자는 이번에 개명하지 않는다

`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree` 는 설정 키·액션 어휘·CLI 플래그·기준선 스냅숏(`contract/testdata/mission_surface_baseline.txt`)에 걸쳐 호환 표면이다. 개명은 호환 이주를 요구하므로 후속 카드의 몫이다 — 카드 발행은 리더에게 하는 권고이며(D-21) 이 카드는 baseline 파일과 위 식별자를 건드리지 않는다. 같은 이유로 `push_serializer`·push 사전 점검은 이름을 `main` 으로 바꾸지 않는다 — 바꾸면 모든 main push 가 직렬화 대상이 된다. github-flow 에서 이 경로는 휴면이다. 스윕 가드는 이 식별자들을 별도 차선에서 기대 파일 목록으로 추적한다.

## §D-13 develop 퇴역은 단계로 한다

| 단계 | 상태 | 조건 |
|---|---|---|
| 0 | develop 이 기준 | 지금 |
| 1 | 수렴과 절체 후. develop 은 남아 있고 더 이상 커밋을 받지 않는다. push 트리거의 `develop` 은 그대로라 develop 팁 CI 는 계속 관측된다 | 되돌리기 원본 유지 |
| 2 | develop 을 읽기 전용으로 만들거나 보호(운영자). `Release PR Multi-OS Gate` 를 main 필수 체크에서 뺀다(운영자, D-22) | 열린 PR 이 develop 을 대상으로 하지 않음(최근 병합 PR 20건 중 17건이 develop 대상이었으므로 재대상 지정이 필요하다). 체크 제거는 M5 의 스위치 켜기가 병합되고 AC-GFD-010 의 두 상태 시험이 통과한 뒤 |
| 3 | develop 삭제(운영자) | `spec-lint.yml` 의 develop 의존 제거 확인(M5), 첫 카드 PR 이 main 까지 완주, main push CI 녹색, 옵션 B 의 `--dry-run` 관측 |
| 4 | 삭제 뒤 정리: 워크플로 push 트리거 목록의 `develop` 제거(리더가 정하는 정리 카드, main 대상 PR) | 잔존 점검 통과(`git grep -n -w develop -- .github/workflows` 종료 코드 1) — 연속 발화(verification-completeness §1.3) |

단계 순서는 시간이 아니라 조건이다. 삭제 시점은 리더 결정이다. `spec-lint.yml:70` 의 `git fetch origin main:… develop:…` 는 develop 이 사라지면 잡 전체를 실패시키므로 삭제보다 앞서 제거한다(M5). 반면 push 트리거의 `develop` 은 삭제 **뒤** 단계 4 에서 정리한다 — 절체 내내 develop 팁 CI 를 관측하기 위해서다(D-17).

## §D-14 4로케일은 의도로 갱신한다

oss-docs 하네스 규칙에 따라 ko 정본을 먼저 쓰고 en, 그다음 ja·zh 를 같은 변경에서 갱신한다. zh 의 추가 문장(입력 보고서가 `moai-sync.md` 의 :105·:309 로 보고)은 M4 의 zh 담당이 살아 있는 서술인지 판정해 지우거나 폐기 표지를 붙인다. en 의 "Check out develop, pull, delete the local branch" 문장은 ko·ja 에 대응이 없어 의도 판정 대상이다. 가드는 긍정 앵커와 부정 스윕만 단언하고 로케일 간 개수는 단언하지 않는다(§D-9).

## §D-15 상시로드 예산

`AGENTS.md` 의 5줄(원장 E-07)과 규칙 서술 교체는 대체로 같은 길이의 교체라 순증가가 작을 것으로 예상하지만 이는 예측이지 측정이 아니다. 기준선(16개·285,543 바이트)을 M4 착수 시 다시 재고, 1,000 바이트를 넘는 단일 편집은 `rule-authoring.md` 의 진술(파일 바이트, 비호출 세션이 치르는 비용, `paths:` 로 옮길 수 있는지)을 변경 설명에 담는다. 새 절이 필요하면 `paths:` 범위의 동반 파일에 둔다.

## §D-16 ~ D-26 개정 0.1.1 에서 기록하는 결정

1회차 plan-audit 의 MP-7 은 미해소 질문 표지 7건(NC-1~NC-7)을 막았다. 아래는 그 7건을 번호 붙은 결정으로 옮기고(D-16~D-22) 감사가 지적한 설계 결함에 대한 결정 넷(D-23~D-26)을 더한 것이다. 각 행은 **출처**(누가 정했나)와 **리더가 뒤집을 수 있는가**를 적는다. 표지 문자열은 `plan.md`·`research.md` 에서 지웠다.

| ID | 결정 | 출처 | 뒤집기 |
|---|---|---|---|
| D-16 | **rc 태그 규칙은 R-a** — 접미사 태그는 `release.yml` 검사 5·6 을 건너뛰고 1~4·7 을 유지한다, 대체 검사 없음. 규칙문은 D-6 한 곳에만 있다(REQ-GFD-008·P6 이 가리킨다) | 계획 기본값(권고) | 리더가 뒤집을 수 있다 |
| D-17 | **절체 선행 조건은 origin/develop 팁의 CI 녹색이다.** 순서: 리더가 미푸시 로컬 develop 커밋을 push → 적색 CI 수리(수리 카드는 리더의 판단이고 이 카드의 범위가 아니다. 측정된 적색 집합: `.moai/reports/t1453/m0-develop-ci-red.md` — gofmt 3파일, Windows vet `parseLsofCWDs`, `run.md` 263줄이 200줄 상한 초과, codemaps 접힘 가드, 그리고 ubuntu 에서만 나는 실패) → 그 뒤에야 절체. **M5 는 워크플로 push 트리거에서 `develop` 을 빼지 않는다** — develop 팁 CI 가 절체 내내 관측되도록, 빼는 일은 develop 퇴역 뒤의 정리 단계다(D-13 단계 4). 두 번째 증거는 develop→main PR 의 필수 체크(`ci.yml:20`)다 | 운영자 결정(10-02 밤, 리더 경유) | 운영자 소관 |
| D-18 | **수렴 PR 은 하나로 먼저 시도한다.** 거부되면 분할 PR → 보호 일시 완화 순. 폴백 선택은 절체 시점의 운영자·리더 판단 | 계획 기본값 | 리더가 뒤집을 수 있다 |
| D-19 | **t1453 이 t810 을 흡수한다.** t810 의 카드·워크트리·`SPEC-LATE-BRANCH-REDESIGN-001` 은 건드리지 않고, t810 의 닫는 처분은 t1453 이 닫힐 때 운영자에게 올린다 | 리더 결정(10-02 밤) | 리더 소관 |
| D-20 | **카드는 카드마다 PR 로 전달하고 병합 큐는 보류한다.** 청구 분·CodeRabbit 한도 이력·의미 충돌 빈도의 측정은 후속 측정 항목이며 `research.md` §9 Gaps 에 남는다. 그 항목의 카드 발행은 리더에게 하는 권고이고 이 SPEC 에는 카드 id 를 적지 않는다 | 운영자 결정(10-02 밤, 리더 경유) | 운영자 소관 |
| D-21 | **와이어 식별자(`push_develop` 등) 개명은 후속 카드를 리더에게 권고로 남긴다.** 이 카드는 식별자·baseline 파일을 건드리지 않고 카드 id 를 만들지 않는다. 권고는 `progress.md` §G 에 있다 | 계획 기본값 | 리더 소관 |
| D-22 | **옵션 B 로 `Release PR Multi-OS Gate` 는 main 의 필수 체크에서 뺀다.** 3-OS 매트릭스는 태그 직전 main SHA 에 `workflow_dispatch` 로 돌리고 `scripts/release.sh` 가 녹색 실행을 강제한다. 필수 체크 제거는 외부 공유 시스템의 되돌리기 어려운 변경이라 **운영자가 직접 수행하는 런북 단계**이며 레인은 하지 않는다. `Release tag immutability (v*)` 규칙셋 때문에 게이트는 태그 **앞**이다 | 운영자 결정(10-02 밤, 리더 경유) | 운영자 소관 |
| D-23 | **계약 모드 에스컬레이션 분류기(`internal/escalation` 의 `pushWhy` 류)는 이 카드가 바꾸지 않는다.** 1회차의 M2(d)를 삭제한다. 그 변경은 새 행동 어휘(기존 `push-develop` 와 같은 와이어 식별자)와 git-flow 비영향 증명이 필요한 별도 변경이라 이 SPEC 에 REQ·AC 가 없다. 1회차 감사가 읽은 `internal/escalation/operational.go:108-126` 에 따르면 `pushWhy` 는 승인된 `develop` 이 아닌 push 에 "pushes <target>" 사유를 돌려주므로, 계약 모드 레인은 PR 전달의 카드 브랜치 push 에서 에스컬레이션을 만나는 것으로 읽힌다(이 회차에 다시 읽지 않았다 — Gap). 후속은 리더에게 권고로 남긴다 | 감사 F-12 반영(제 판단) | 리더 소관 |
| D-24 | **매트릭스 강제의 스위치는 `scripts/release.sh --require-matrix-run` 옵션이다(기본 꺼짐, 새 설정 키 없음).** 근거: 새 설정 키는 사용자에게 하는 약속이라 정직성 가드가 필요하고(`.moai/docs/config-key-triage-rule.md`) 이 저장소의 `workflow.yaml` 값은 AC-GFD-016 의 사전 방향 목록에 있어 M3 에서 바꿀 수 없다. `release.sh` 의 인자 해석은 `case` 문이고 모르는 `-*` 플래그는 `die` 이므로 새 옵션은 한 분기로 들어간다. 옵션이 없으면 현행 동작이다. 켜는 일은 릴리스 하네스 본문(`hns-release-specialist.md`, 절체 시점 파일)이 옵션을 넘기도록 바꾸는 M5 의 편집이다 | 감사 F-21 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-25 | **절체 카드(이 카드) 자신의 면제와 사전 점검의 순서.** 이 카드의 M4·M5 묶음은 런북 2단계에서 현 경로로 마지막에 병합되고 push 된다. 사전 점검(AC-GFD-019)은 그 **뒤**, 수렴 병합(4단계) **앞**인 3단계에서 돌며, 이때 이 카드의 브랜치 팁이 origin/develop 의 조상이면 picked 카드 조건을 자연히 만족한다. 병합 뒤 커밋(예: 증거 hoist)이 있어 조상이 아니면 `--exclude-card <id>` 인자로 명시하고 점검 출력이 그 제외를 이름으로 적는다. 다른 picked 카드는 제외되지 않는다 | 감사 F-03 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-26 | **Frozen 조항 개정의 실행 모형.** 등재된 조항은 `CONST-V3R5-027`·`-028` 둘뿐이다(`moai constitution list --file …spec-workflow.md` 가 세 항목 중 이 둘을 낸다. `…worktree-integration.md` 는 `No entries.`). `moai constitution amend` 는 `--rule`·`--before`·`--after`(필수), `--evidence`(Frozen 에서 필수), `--dry-run` 을 받고 5단째 인간 승인은 대화형 Y/N 이다. 그러므로 레인은 개정 입력(`--before`/`--after` 문언과 `--evidence`)과 `--dry-run` 제안을 준비하고, 실제 개정은 **운영자가 두 번 실행**하며, 레인은 그 뒤에 `moai constitution validate` 종료 코드 0 과 등재·본문 문언 일치를 검증한다. `worktree-integration.md:390,401` 의 Frozen 표지 두 줄은 등재되지 않았으므로 일반 편집이다 | 감사 F-06 반영(제 판단) | 리더가 뒤집을 수 있다 |

---

🗿 MoAI
