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
| CI 벽시계 — develop push | 측정, **실행 id 열두 개로 고정**(2026-10-02, 열두 id 를 `gh run view <id>` 로 하나씩 다시 읽었다 — 목록 호출 `gh run list` 는 재현되지 않아 쓰지 않는다, D-28) | 열두 실행: failure 7회(1046·1223·1413·1484·1517·1551·1443 초, 평균 1382 초)·cancelled 5회·success 0회. `createdAt` 2026-10-01T13:38:32Z ~ 2026-10-02T09:27:17Z (원장 M-1, id 표) |
| CI 벽시계 — main push | 미측정(재현되지 않아 표에서 뺐다) | 창을 고정한 200건 목록 호출(원장 M-2 의 1회차 기록)은 감사 회차에 다른 결과와 HTTP 502 를 돌려주었고 200개 실행을 id 로 고정하지도 않았다. 이 SPEC 의 어떤 결정도 이 수에 기대지 않는다. `research.md` §9 Gaps 에 올린다 |
| 병합 PR 의 대상 분포 | 측정, **PR 번호 스무 개로 고정**(2026-10-02, 스무 PR 을 번호별로 다시 읽었다) | 스무 PR: develop 17·main 3(1695·1702·1740). 병합된 PR 의 `baseRefName` 은 병합 뒤 바뀌지 않는다 (원장 M-3, 번호 표) |
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
2. **누적 patch-id** — 카드 브랜치와 병합 기준점의 누적 diff 를 patch-id 로 만들어, 기준점 이후 통합 ref 의 커밋들의 patch-id 와 비교한다. **비교할 커밋 수의 상한은 500 이다.** 근거: `git rev-list --count --first-parent --since=2026-09-25T00:00:00Z --until=2026-10-02T00:00:00Z 284e09c44023598affe486f17701717ca173e6ca` 가 `248` 이었고(원장 M-6, 팁 SHA 로 고정한 형태, 1주 동안 착지한 카드 병합 수), 500 은 그 약 2주분이다. 넘으면 이 층은 "답하지 못함"이고 다음 층으로 간다.
3. **PR 병합 상태** — `gh` 로 카드 브랜치를 헤드로 하는 PR 을 읽는다. 다음 셋이 모두 참일 때만 착지다: PR 상태가 MERGED, **PR 의 `headRefOid` 가 로컬 카드 팁과 같다**, 병합 커밋(`mergeCommit.oid`)이 통합 ref 의 조상이다. 로컬 팁이 `headRefOid` 와 다르거나 PR 병합 뒤에 로컬 커밋이 더 있으면 착지가 아니라 **보존**이다(병합된 PR 뒤에 연장된 브랜치를 지우면 미병합 작업이 사라진다). 병합 커밋이 아직 로컬 통합 ref 에서 도달되지 않는 경우(가져오기 지연)도 **보존**이다 — 이 세 번째 조건만 빠진 층 3 은 AC-GFD-002 의 F9 가 잡는다. `gh` 는 한 번 호출에 **제한 시간 10 초**, 재시도 없이, 시간 초과·비 0 종료·토큰 부재는 모두 "답하지 못함 → 보존"(fail-closed)이다. `worktree sweep` 은 가능하면 병합된 PR 을 `gh pr list --state merged` 한 번으로 일괄 읽어 같은 제한을 적용한다.

어느 층도 확정하지 못하면 보존한다. 이는 `sweep` 이 이미 가진 세 갈래 계약(착지·미착지·답할 수 없음→보존)을 유지한다.

**어느 경로가 어느 층까지 도는가**(REQ-GFD-002 의 두 문장이 이 구분을 그대로 적는다): `moai worktree done` 과 `moai worktree sweep` 은 세 층을 모두 돈다. **세션 종료 정리는 1·2층까지만 돌고 3층(네트워크)을 돌지 않는다** — 세션 종료 경로는 REQ-WSS-304 가 "No network runs ... shared exit path stays cheap" 로 정한 대로 네트워크를 쓰지 않으며 이 SPEC 이 그 성질을 바꾸지 않는다. 1·2층이 확정하지 못하면 보존하고 그 트리는 다음 `sweep` 의 3층이 판정한다. 이 성질은 양성 시험으로 고정한다(AC-GFD-002: 세션 종료 경로에서 `gh` 호출 0).

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

**배선 계약(AC-GFD-008 이 판정하는 것은 스크립트만이 아니라 출하되는 관문이다)**: 스크립트의 인터페이스는 위치 인자 하나 `scripts/verify-release-provenance.sh <tag>` 이고 저장소 루트에서 도는 읽기 전용 검사 1~7 이다. 네트워크가 필요한 두 `git fetch` 줄(태그와 `origin/main` 가져오기)과 `env: TAG: ${{ github.ref_name }}` 는 워크플로에 남고, 인라인 검사 1~7 의 본문은 지워져 그 자리에 `bash scripts/verify-release-provenance.sh "${TAG}"` 한 줄이 선다. 그래서 (a) 워크플로에 `check 5 (`·`check 6 (` 가 없어야 하고(`git grep` 종료 코드 1), (b) 스크립트 호출이 있어야 하며(`git grep -F` 종료 코드 0), (c) 시험이 워크플로 단계의 `run` 본문에서 호출 줄을 뽑아 시험이 직접 부르는 인자(`<tag>` 하나)와 같은 인자로 같은 픽스처에서 재생해 같은 종료 코드와 같은 판정 줄을 얻어야 한다. **접미사 없는 태그의 동치**는 정식 태그 픽스처 셋((4) 5 없음·(5) 6 틀림·(6) 모두 올바름)이 각자 변경 전 인라인 단계가 찍던 문구(`RELEASE_PROVENANCE_GATE: check N (<이름>)` 의 N·이름과 성공 줄 `all 7 checks passed for <tag> (<commit>)`)와 같은 판정 줄을 내는 것으로 본다 — 기준 문구는 원장 E-06 과 변경 전 파일(`.github/workflows/release.yml` 의 `all 7 checks passed` 줄, 원장 E-44)에서 오고, 변경 전 인라인 단계를 같은 모양의 스크래치 픽스처 셋에서 실행한 판정 줄과 종료 코드는 원장 M-10 에 있다.

**위험**: 이 저장소에는 원격에 없는 로컬 태그(`v3.1.0-rc.0`~`rc.2`, 합성 보고서 인용)가 있어 전체 태그 push 는 불변 태그를 만든다. 런북은 이름 붙인 태그 하나만 push 한다.

## §D-7 3-OS 매트릭스 게이트

**관측**: 3-OS race 매트릭스는 `release/*` 헤드나 수동 실행에서만 돈다. 필수 체크 `Release PR Multi-OS Gate` 는 `if: always()` 이고 detect 나 매트릭스가 failure 일 때만 실패하므로 비릴리스 PR 에서는 성공한다. `release/*` PR 이 없어지면 이 게이트는 영구 녹색 무동작이다. `v*` 태그는 만든 뒤 지우거나 옮길 수 없으므로(`research.md` §2) 검증은 태그 **앞**에 있어야 한다.

| 옵션 | 내용 | 비용·특성 |
|---|---|---|
| A | main 대상 모든 PR(Go 변경 시)에서 3-OS 매트릭스 | 가장 강하다. 카드마다 3-OS race(레그당 제한 시간 30)를 더해 카드 수에 비례. 필수 체크가 의미를 얻는다 |
| **B** | 태그 전에 대상 SHA 로 `workflow_dispatch` 매트릭스를 돌리고 `release.sh` 가 그 SHA 의 통과 기록이 없으면 태그를 거부 | 릴리스 시도당 1회. 수동 단계가 하나 늘고 릴리스 사이 main 의 mac/windows race 적색은 보이지 않을 수 있다(3-OS 통합 시험은 CI 마다 돈다) |
| C | `release.yml` 안, GoReleaser 앞에서 매트릭스 | 자동이지만 태그가 이미 존재한 뒤라 실패하면 태그가 소각된다 — 기각 |
| D | main push 마다 매트릭스 + `release.sh` 가 태그 대상 SHA 의 녹색을 확인 | 지속 가시성. 병합 카드마다 매트릭스(동시 병합은 `ci.yml` 의 ref 단위 `cancel-in-progress: true` 로 합쳐질 수 있다 — 구성 읽기이고 취소 빈도는 측정하지 않았다, `research.md` §9) |

**채택**: B. 출처: 운영자 결정(10-02 밤, 리더 경유). 근거: 되돌릴 수 없는 행위(태그) 직전에만 비싼 검증을 둔다. 태그 불변 규칙셋(`Release tag immutability (v*)`)이 있으므로 게이트는 태그 **앞**에 있다. D 는 업그레이드 경로로 남긴다(B 와 같은 `release.sh` 확인을 그대로 쓴다). **강제는 절체 시점에 켠다** — 현재 git-flow 릴리스는 main 병합 SHA 에 매트릭스를 돌리지 않으므로 지금 켜면 현행 정식 릴리스가 막힌다(M3). 켜는 방법은 D-24 의 `--require-matrix-run` 옵션이다.

**필수 체크 `Release PR Multi-OS Gate`(NC-7 의 해소, D-22)**: 옵션 B 아래서 이 체크는 비릴리스 PR 에서 무동작 성공이므로 main 의 필수 체크 목록에서 **뺀다**(운영자 결정). 이 변경은 외부 공유 시스템의 되돌리기 어려운 변경이라 **운영자가 직접 수행하는 런북 단계**이며 이 카드의 레인은 실행하지 않는다(REQ-GFD-020). 순서: M5 의 스위치 켜기가 병합되고 AC-GFD-010 의 두 스위치 상태 시험이 통과한 뒤(런북 9a 단계). 체크를 빼면 `release/*` 헤드 PR 경로의 자동 매트릭스는 더 이상 병합을 막지 않는다 — 그 자리를 태그 직전 확인이 대신한다.

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

선행 조건은 운영자 결정(D-17)이다 — 리더가 미푸시 로컬 develop 커밋을 push 하고, 적색 CI 는 리더가 정하는 수리로 녹색이 되고(수리 카드는 이 카드의 범위가 아니다, 측정된 적색 집합은 `research.md` §4.1 에 실행 id 와 함께 옮겨 두었다 — `.moai/reports/t1453/m0-develop-ci-red.md` 는 gitignored 로컬 보조 자료라 인용 대상이 아니다), 그 뒤에야 절체가 진행된다. develop 의 push CI 는 절체 내내 계속 돈다(M5 는 push 트리거에서 `develop` 을 빼지 않는다).

| # | 단계 | 주체 | 외부 공유 시스템 |
|---|---|---|---|
| 0 | 선행 조건 확인(D-17): 리더가 미푸시 로컬 develop 커밋을 일괄 push 하고 적색 CI 수리가 끝나 origin/develop 팁의 CI 가 success 인지 `gh run list --workflow=CI --branch develop --limit 3` 로 관측해 확인 기록에 적는다 | 리더 | 예(develop push) |
| 1 | 레인 정지: 새 배차 중단, 각 레인이 진행 카드를 끝내고 증거를 hoist 하고 `/clear`, 세션 종료, 워크트리 정리(착지 전 폐기 금지) | 리더(레인 각자) | 아니오 |
| 2 | M4·M5 묶음을 현 경로로 develop 에 마지막으로 병합(M1~M3 는 이미 병합돼 있다)하고 push. develop 팁 CI 를 관측한다(push 트리거가 살아 있으므로 관측 가능) | 리더 | 예(develop push) |
| 3 | 배치 경계 사전 점검(AC-GFD-019, 2단계 병합·push 뒤): 미푸시 0, 창·슬롯 보유자 없음, 병합·push 되지 않은 picked 카드 없음(이 카드는 2단계 병합 뒤라 제외 — D-25), 활성 레인 세션 없음. 점검 출력이 확인 기록의 일부 | 카드 스크립트, 리더 실행 | 아니오 |
| 4 | develop 이 `origin/main` 을 흡수하고 push | 리더 | 예(develop push) |
| 5a | develop→main PR 개설(병합 커밋 방식 지정, 운영자의 확인 기록을 받은 뒤). 이 PR 의 필수 체크(`pull_request: branches: [main]`, `ci.yml:20`)가 두 번째 CI 증거다 | 리더 | 예(PR) |
| 5b | 필수 체크 통과를 읽고 그 PR 을 병합 커밋으로 병합 | 운영자 | 예(main) |
| 6 | 트리 항등과 조상 확인 | 리더 | 아니오 |
| 7 | 레인 재기동: `moai cc -f N` 로 새 세션, 모든 새 카드 트리는 `origin/main` 에서. 3단계가 병합·push 되지 않은 picked 카드 0 을 보증했으므로 이전 기준에서 만든 미착지 카드 브랜치는 없다 | 리더 | 아니오 |
| 8 | 첫 카드가 PR→CI→병합→착지 판정→sweep 까지 가는지 관측 | 리더 | 예(PR) |
| 9a | develop 퇴역 단계화(§D-13)의 운영자 몫: 단계 2 에서 `Release PR Multi-OS Gate` 를 main 필수 체크에서 빼고(D-22), develop 을 보호하고, 조건이 서면 develop 을 삭제 | 운영자 | 예 |
| 9b | 삭제 뒤 리더의 몫: 워크플로 push 트리거의 `develop` 을 정리하는 카드(main 대상 PR)와 잔존 점검(`git grep -n -w develop -- .github/workflows` 종료 코드 1, `gh run list --workflow=CI --branch develop --limit 3` 의 가장 새 `createdAt` 이 전진하지 않음) | 리더 | 예(PR) |

**행마다 주체는 하나다**(카드·리더·운영자 중 하나 — D-29). 한 단계에 두 주체가 필요하면 접미 번호(`5a`·`5b`)로 행을 나눈다. 5b 와 9a 의 운영자 단계, 0·2·4 의 develop push 는 이 카드가 실행하지 않는 단계다(REQ-GFD-020). 이 표가 런북의 골격이고 M6 의 런북 문서는 같은 순서·같은 두 열(주체, 외부 공유 시스템 여부)과 같은 접미 번호를 가진다(AC-GFD-023).

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

**핵심 패턴(RE2, 뒷보기 없음 — 개정 0.1.2 회차 스크래치 프로브로 픽스처 25행과 표면 1523개 파일을 돌려 관측했다, 원장 M-9. 0.1.1 의 프로브 M-5 는 21행 설계의 기록이다)**

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
4. 그 밖의 맨 적중은 **약한 적중**이다 — **허용 목록에 줄 전체로 오르지 않았으면 위반이다**(D-27). 0.1.1 설계는 약한 적중을 "위반이 아니라 개수로만 보고"했고 그 결과 `Work starts from develop.`·`Lane PRs go to develop.`·`Compare with the develop tip.` 같이 분기 말 토큰이 없는 살아 있는 문장이 가드를 통과했다(원장 M-9 의 회피 프로브). 동사 `develop` 의 거짓 양성은 이제 가드가 아니라 허용 항목(`Why` 에 "동사/키워드, 브랜치 아님")이 처리한다.

**P-C(약, 기준선 래칫)**: 토큰 없는 살아 있는 문장 — `integration (worktree|window|branch)`·`통합 (워크트리|브랜치)`·`일괄 push`·`batch.?push`·`commit-dead`·`lead_push_threshold`. 위반이 아니라 절체 묶음이 기록한 개수보다 늘면 실패하고 줄면 보고만 한다(`moai spec lint --baseline` 의 구조를 따른다). **`\bWT-` 는 뺐다** — 카드 브랜치 `WT-<slug>` 는 github-flow 에서도 살아 있는 이름이다(§A 1단계).

**허용(조용한 허용을 막기 위해 모두 좁혔다)**
- 줄 표지: 같은 줄이 `RETIRED`·`SUPERSEDED`·`폐기`·`은퇴` 중 하나와 **날짜(`20YY-MM-DD`)나 카드·SPEC 식별자(`t1234`·`SPEC-…`)를 함께** 가질 때. `historical`·`legacy` 는 표지로 인정하지 않는다(실측 프로브: `(legacy)` 를 붙인 살아 있는 문장은 위반으로 남는다). 제목 줄이 표지를 가지면 그 효력은 **그 절 자신**, 곧 하위 제목을 포함해 같은 수준이거나 더 상위 수준인 다음 제목 전까지에 한정된다(프로브 F-red-10·F-green-8). 인용 블록은 그 블록 하나에만 효력이 있다.
- 스탬프: `develop @ <7~40자 hex>` 형태.
- git-flow 옵션 서술은 **줄 표지가 아니라 파일·줄 단위 허용 항목** `{File, Literal, Why}` 으로만 허용한다. `Literal` 은 **그 파일의 한 줄 전체(앞뒤 공백 제거 뒤 동일)** 여야 하고, 파일에 그런 줄이 없으면 낡은 항목으로 실패한다. 맨 토큰(`develop`)은 줄 전체가 아니므로 거부된다(1회차 설계는 `{File: "AGENTS.md", Literal: "develop"}` 한 항목이 파일 전체를 침묵시키는 변이를 막지 못했다). **허용 항목은 전체 40개 이하**이다. 근거: 입력 보고서 §5 가 문서화된 git-flow 옵션(class N)으로 든 곳은 `delivery.md`·`moai-ref-git-workflow`·`manager-git.md`·`spec-workflow.md:60` 과 CI 샘플 줄 몇 곳이고, 40 은 그 줄 수의 두 배를 넘겨 잡은 여유다(이 회차에 줄 수를 다시 세지는 않았다 — Gap). 상향은 이 문서의 결정 행이 필요하다. **이 40개 상한은 약한 적중을 허용하는 항목까지 합친 전체 수다**(D-27 — 상한은 올리지 않았다). 이 회차의 측정(원장 M-9)은 약한 적중 31줄(25개 파일, 로컬·템플릿 사본 쌍 포함)이고 그중 살아 있는 기준 서술(`goal.md` "the leased local develop" 와 사본, `scripts/jev/triage.py` "ancestor of develop" 와 시험 줄, `scripts/ac-baseline/*.sh` "until it absorbs develop")은 허용이 아니라 **문장 교체** 대상이며 동사·키워드·중괄호 목록(`Create and develop foundation SPEC`, `[…, "develop", …]`, `manager-{spec,develop,docs}`) 쪽은 사본 쌍까지 열 줄 안팎이라 git-flow 옵션 서술 줄과 합쳐도 상한 안에 든다는 것은 **예측이지 측정이 아니다**. 합이 40 을 넘으면 허용을 늘리지 않고 문장을 먼저 바꾸며, 그래도 안 되면 결정 행으로 상향한다. `git-flow` 라는 단어가 있는 줄을 일괄 허용하지 않는다.
- 구조 식별자(`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree`)는 위 경계 때문에 본 패턴에서 이미 빠지고, **별도 차선**의 기대 파일 목록으로 추적해 새 파일로 번지면 실패한다(§D-12).
- 표지가 줄 단위로 면제한 줄 수는 파일별로 방문 보고에 센다(표지 추가로 살아 있는 문장을 덮는 변이는 막지 못하고 개수 증가를 눈에 띄게 할 뿐이다 — Residual-risk, 리뷰 소관).

**무장과 연속 발화**: 가드 본체는 이 저장소의 `git-strategy.yaml` 의 활성 워크플로 값이 `github-flow` 일 때 무장한다. 절체 묶음이 그 값을 바꾸므로 값 변경과 가드 무장이 한 변경이다. 무장 전에는 `t.Skip` 이 아니라 "disarmed" 를 출력하는 별도 시험이 있고, AC-GFD-011·021 은 `--- SKIP` 이 없음을 요구한다. 방문 파일 수 바닥값(전체 1372, 하위 트리마다 위 표의 값)과 하위 트리 집합 동등 단언은 순회가 조용히 줄어드는 것을 잡는다. 가드는 `visited=<N>`, 하위 트리별 방문 수, **허용 목록으로 통과한 약한 적중 수와 위반으로 남은 약한 적중 수**, 표지 면제 줄 수를 출력한다. 허용 목록 사용 수가 항목 수보다 적으면(낡은 항목) 가드가 실패한다.

**남는 한계(Residual-risk, 리뷰 소관)**: 핵심 패턴은 대소문자를 구분한다(`develop` 만). `Develop branch is the integration line.`·`DEVELOP` 는 일치하지 않는다(프로브 M-9: nomatch). 이 트리의 표면에서 `git grep -n -w "Develop"` 가 낸 16줄은 모두 동사·제목·다이어그램 라벨이었다(원장 E-45) — 대소문자 무시 변형을 더하면 `Develop in the Worktree` 같은 동사 줄이 P-B 토큰(`worktree`)에 걸려 거짓 양성이 늘어 허용 항목을 소모하므로 더하지 않고 한계로 기록한다. 또한 H1 제목의 표지는 그 절(= 문서 전체)을 침묵시키고, 날짜를 곁들인 표지를 살아 있는 줄에 붙이는 변이는 막지 못한다(표지 면제 줄 수 보고가 증가를 눈에 띄게 할 뿐이다).

**퇴역 뒤의 연속 발화(별도 점검)**: develop 이 퇴역한 뒤에도 워크플로 push 트리거에 `develop` 이 남으면 그 목록은 조용히 죽은 설정이다(실행되지 않는 점검은 성공과 구별되지 않는다). 런북 9b 단계의 점검이 이를 잡는다 — `git grep -n -w develop -- .github/workflows` 가 종료 코드 1 이어야 하고(원장 M-7 은 지금 10개 파일이 적중함을 보인다), `gh run list --workflow=CI --branch develop --limit 3` 의 가장 새 `createdAt` 이 더 이상 전진하지 않아야 한다. 이 점검 문구는 런북 문서의 필수 내용이다(AC-GFD-023).

**기각**
- *로케일별 개수 동등 단언*: 현재 트리에서 이미 실패한다 — `moai-sync.md` 의 `develop` 단어 적중이 en 6·ja 5·ko 5·zh 5 이고 분류한 살아 있는 문장은 3/2/2/4 로 보고됐다. 의도를 단언한다.
- *내장 템플릿만 순회*: 저장소 사본(예: 원장 E-07 의 `AGENTS.md`)을 놓친다.
- *`git-flow` 줄 일괄 허용*: 위 이유.
- *Python 식 유니코드 `\w`*: CJK 가 단어 문자로 합쳐져 `develop에서` 를 놓친다. RE2 의 ASCII 클래스를 명시한다.
- *줄 표지에 `historical`·`legacy` 인정, 날짜·식별자 없는 표지*: 프로브에서 `(legacy)` 한 단어가 살아 있는 문장을 면제하는 변이를 막지 못했다.
- *`Literal` 이 맨 토큰이어도 허용*: 항목 하나가 파일 전체를 침묵시킨다.
- *약한 적중을 개수로만 보고(0.1.1 설계) 또는 기준선 래칫으로 묶기*: 보고만 하면 토큰 없는 살아 있는 문장이 영원히 녹색이고, 래칫은 지금의 31줄(`the leased local develop`·`ancestor of develop` 같은 살아 있는 기준 서술 포함)을 기준선으로 굳혀 "늘지만 않으면 통과"로 만든다. 줄마다 허용 사유를 요구하는 쪽이 REQ-GFD-011("살아 있는 문장이 남지 않아야 한다")과 같은 방향이다(D-27).
- *핵심 패턴의 P-B 토큰에 `from|against|ancestor|absorb|lease|go(es)? to|on` 를 더하기*: `on`·`from` 은 일반 영어 동사구와 겹쳐 거짓 양성이 크고 목록은 다음 회피 문장이 나올 때마다 늘어난다. 토큰을 늘리는 대신 약한 적중 전체를 위반으로 승격해 목록을 닫는다.

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

## §D-16 ~ D-30 개정 0.1.1~0.1.3 에서 기록하는 결정

1회차 plan-audit 의 MP-7 은 미해소 질문 표지 7건(NC-1~NC-7)을 막았다. 아래는 그 7건을 번호 붙은 결정으로 옮기고(D-16~D-22) 감사가 지적한 설계 결함에 대한 결정 넷(D-23~D-26)을 더한 것이다. 개정 0.1.2 는 2회차 감사가 지적한 결함에 대한 결정 셋(D-27~D-29)과 D-26 의 보충(Frozen 줄 전수)을 더한다. 개정 0.1.3 은 run 단계 중 리더가 정한 in-run 개정 A-1 의 결정 하나(D-30)와 그 보충을 더한다. 각 행은 **출처**(누가 정했나)와 **리더가 뒤집을 수 있는가**를 적는다. 표지 문자열은 `plan.md`·`research.md` 에서 지웠다.

| ID | 결정 | 출처 | 뒤집기 |
|---|---|---|---|
| D-16 | **rc 태그 규칙은 R-a** — 접미사 태그는 `release.yml` 검사 5·6 을 건너뛰고 1~4·7 을 유지한다, 대체 검사 없음. 규칙문은 D-6 한 곳에만 있다(REQ-GFD-008·P6 이 가리킨다) | 계획 기본값(권고) | 리더가 뒤집을 수 있다 |
| D-17 | **절체 선행 조건은 origin/develop 팁의 CI 녹색이다.** 순서: 리더가 미푸시 로컬 develop 커밋을 push → 적색 CI 수리(수리 카드는 리더의 판단이고 이 카드의 범위가 아니다. 측정된 적색 집합은 `research.md` §4.1 에 있다 — 로컬 보조 자료 `.moai/reports/t1453/m0-develop-ci-red.md` 는 인용 대상이 아니다 — gofmt 3파일, Windows vet `parseLsofCWDs`, `run.md` 263줄이 200줄 상한 초과, codemaps 접힘 가드, 그리고 ubuntu 에서만 나는 실패) → 그 뒤에야 절체. **M5 는 워크플로 push 트리거에서 `develop` 을 빼지 않는다** — develop 팁 CI 가 절체 내내 관측되도록, 빼는 일은 develop 퇴역 뒤의 정리 단계다(D-13 단계 4). 두 번째 증거는 develop→main PR 의 필수 체크(`ci.yml:20`)다 | 운영자 결정(10-02 밤, 리더 경유) | 운영자 소관 |
| D-18 | **수렴 PR 은 하나로 먼저 시도한다.** 거부되면 분할 PR → 보호 일시 완화 순. 폴백 선택은 절체 시점의 운영자·리더 판단 | 계획 기본값 | 리더가 뒤집을 수 있다 |
| D-19 | **t1453 이 t810 을 흡수한다.** t810 의 카드·워크트리·`SPEC-LATE-BRANCH-REDESIGN-001` 은 건드리지 않고, t810 의 닫는 처분은 t1453 이 닫힐 때 운영자에게 올린다 | 리더 결정(10-02 밤) | 리더 소관 |
| D-20 | **카드는 카드마다 PR 로 전달하고 병합 큐는 보류한다.** 청구 분·CodeRabbit 한도 이력·의미 충돌 빈도의 측정은 후속 측정 항목이며 `research.md` §9 Gaps 에 남는다. 그 항목의 카드 발행은 리더에게 하는 권고이고 이 SPEC 에는 카드 id 를 적지 않는다 | 운영자 결정(10-02 밤, 리더 경유) | 운영자 소관 |
| D-21 | **와이어 식별자(`push_develop` 등) 개명은 후속 카드를 리더에게 권고로 남긴다.** 이 카드는 식별자·baseline 파일을 건드리지 않고 카드 id 를 만들지 않는다. 권고는 `progress.md` §G 에 있다 | 계획 기본값 | 리더 소관 |
| D-22 | **옵션 B 로 `Release PR Multi-OS Gate` 는 main 의 필수 체크에서 뺀다.** 3-OS 매트릭스는 태그 직전 main SHA 에 `workflow_dispatch` 로 돌리고 `scripts/release.sh` 가 녹색 실행을 강제한다. 필수 체크 제거는 외부 공유 시스템의 되돌리기 어려운 변경이라 **운영자가 직접 수행하는 런북 단계**이며 레인은 하지 않는다. `Release tag immutability (v*)` 규칙셋 때문에 게이트는 태그 **앞**이다 | 운영자 결정(10-02 밤, 리더 경유) | 운영자 소관 |
| D-23 | **계약 모드 에스컬레이션 분류기(`internal/escalation` 의 `pushWhy` 류)는 이 카드가 바꾸지 않는다.** 1회차의 M2(d)를 삭제한다. 그 변경은 새 행동 어휘(기존 `push-develop` 와 같은 와이어 식별자)와 git-flow 비영향 증명이 필요한 별도 변경이라 이 SPEC 에 REQ·AC 가 없다. 1회차 감사가 읽은 `internal/escalation/operational.go:108-126` 에 따르면 `pushWhy` 는 승인된 `develop` 이 아닌 push 에 "pushes <target>" 사유를 돌려주므로, 계약 모드 레인은 PR 전달의 카드 브랜치 push 에서 에스컬레이션을 만나는 것으로 읽힌다(이 회차에 다시 읽지 않았다 — Gap). 후속은 리더에게 권고로 남긴다 | 감사 F-12 반영(제 판단) | 리더 소관 |
| D-24 | **매트릭스 강제의 스위치는 `scripts/release.sh --require-matrix-run` 옵션이다(기본 꺼짐, 새 설정 키 없음).** 근거: 새 설정 키는 사용자에게 하는 약속이라 정직성 가드가 필요하고(`.moai/docs/config-key-triage-rule.md`) 이 저장소의 `workflow.yaml` 값은 AC-GFD-016 의 사전 방향 목록에 있어 M3 에서 바꿀 수 없다. `release.sh` 의 인자 해석은 `case` 문이고 모르는 `-*` 플래그는 `die` 이므로 새 옵션은 한 분기로 들어간다. 옵션이 없으면 현행 동작이다. 켜는 일은 릴리스 하네스 본문(`hns-release-specialist.md`, 절체 시점 파일)이 옵션을 넘기도록 바꾸는 M5 의 편집이다 | 감사 F-21 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-25 | **절체 카드(이 카드) 자신의 면제와 사전 점검의 순서.** 이 카드의 M4·M5 묶음은 런북 2단계에서 현 경로로 마지막에 병합되고 push 된다. 사전 점검(AC-GFD-019)은 그 **뒤**, 수렴 병합(4단계) **앞**인 3단계에서 돌며, 이때 이 카드의 브랜치 팁이 origin/develop 의 조상이면 picked 카드 조건을 자연히 만족한다. 병합 뒤 커밋(예: 증거 hoist)이 있어 조상이 아니면 `--exclude-card <id>` 인자로 명시하고 점검 출력이 그 제외를 이름으로 적는다. 다른 picked 카드는 제외되지 않는다 | 감사 F-03 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-26 | **Frozen 조항 개정의 실행 모형.** 등재된 조항은 `CONST-V3R5-027`·`-028` 둘뿐이다(`moai constitution list --file …spec-workflow.md` 가 세 항목 중 이 둘을 낸다. `…worktree-integration.md` 는 `No entries.`). `moai constitution amend` 는 `--rule`·`--before`·`--after`(필수), `--evidence`(Frozen 에서 필수), `--dry-run` 을 받고 5단째 인간 승인은 대화형 Y/N 이다. 그러므로 레인은 개정 입력(`--before`/`--after` 문언과 `--evidence`)과 `--dry-run` 제안을 준비하고, 실제 개정은 **운영자가 두 번 실행**하며, 레인은 그 뒤에 `moai constitution validate` 종료 코드 0 과 등재·본문 문언 일치를 검증한다. 등재 여부는 줄마다 `moai constitution list --file <경로>` 로 판정한다 — 이 SPEC 이 건드리는 `[ZONE:Frozen]` 줄의 전수와 분류는 아래 표(D-26 보충)에 있다. 등재되지 않은 줄(`worktree-integration.md:390,401`, `spec-workflow.md` 의 Route A/B 문단 표지 줄과 plan 단계 표지 줄)은 `amend` 를 쓸 수 없는(`--rule` 은 등재 ID 만 받는다) **일반 편집**이며 개정된 등재 문언과 같은 방향이어야 하고 로컬과 템플릿 사본이 같아야 한다 | 감사 F-06 반영(제 판단), 2회차 N-03 로 보강 | 리더가 뒤집을 수 있다 |
| D-27 | **스윕 가드의 약한 적중은 허용 목록에 줄 전체로 오르지 않으면 위반이다**(기준선 래칫이 아니다). 허용 목록 전체 상한 40 은 올리지 않는다 — 약한 적중 허용도 같은 40 안에서 센다. 근거: 감사가 같은 알고리즘을 다시 구현해 이 트리 1523개 파일에서 위반 286줄·약한 적중 31줄·표지 면제 9줄을 측정했고(원장 M-9 가 같은 수를 재현했다), 약한 적중에는 `the leased local develop`(`goal.md`)·`ancestor of develop`(`scripts/jev/triage.py`) 같은 살아 있는 기준 서술이 들어 있었으며 `Work starts from develop.`·`Lane PRs go to develop.`·`Compare with the develop tip.` 가 모두 녹색이었다(0.1.1 설계의 구멍). 픽스처는 21→25행(F-red-12·13·14, F-green-10 추가) | 2회차 감사 N-05 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-28 | **CI 증거는 변하지 않는 식별자로 고정하거나 Gaps 로 옮긴다.** `gh run list`·`gh pr list` 같은 목록 호출은 같은 명령이 다른 창을 돌려주는 것이 두 회차 연속 관측됐다(감사: 세 번 연속 호출이 2026-09-07·2026-09-29~30·2026-10-01~02 창을 돌려주었다. 이번 회차: 세 번 연속 호출이 2026-09-07 창 한 번과 2026-10-01~02 창 두 번을 돌려주었다 — 같은 명령이 서로 다른 결과를 냈다). 그래서 인용하는 CI 행은 실행 id(`gh run view <id>`)나 PR 번호로 고정해 다시 읽은 값만 쓰고, 고정하지 못한 행(main push 200건 표본, PR 이벤트 표본)은 표에서 빼 `research.md` §9 Gaps 에 올린다. "같은 명령 N회 동일" 문구는 쓰지 않는다 | 2회차 감사 N-07 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-29 | **런북의 행마다 실행 주체는 하나다.** 한 단계에 두 주체가 있으면 접미 번호 행으로 나눈다(5→5a 리더·5b 운영자, 9→9a 운영자·9b 리더). 근거: AC-GFD-023 의 형태 시험이 주체 열을 한 값(카드·리더·운영자)으로 파싱하고, 외부 공유 시스템을 바꾸는 운영자 단계가 두 값 셀 안에 묻히면 REQ-GFD-020 의 확인 경계를 읽을 수 없다 | 2회차 감사 N-02 반영(제 판단) | 리더가 뒤집을 수 있다 |
| D-30 | **계약에서 미션으로의 투영이 쓰는 병합 목표는 호출자가 넘기는 구성된 통합 목표다**(개정 A-1, REQ-GFD-023). `ProjectToMission(c, mergeTarget)` — 빈·공백 목표는 `merge_target` 을 이름으로 대는 `ErrNotProjectable` 거부이고 `develop` 대체값은 없다(M1 이 기록한 옵션 A 와 같은 규칙: 목표 없음 = 답 없음). 와이어 식별자·`mission_surface_baseline.txt`·`internal/mission/git_owner.go:201` 은 이번에 건드리지 않는다. 근거·기각된 대안은 아래 D-30 보충 | 리더 결정(2026-10-03, 카드 t1453 run 단계 — 후속 권고로 남기지 않고 카드에 포함) | 리더 소관 |

### D-30 보충 — 투영의 병합 목표 (개정 A-1, 측정: 트리 `eda61419564296892b41d0e7608c8b136e52c04c`, 원장 E-46~E-55·M-11)

**D-1·D-12 와의 위치.** D-1 은 develop 에 고정된 읽기를 구성된 통합 목표 해석기로 옮겨 git-flow 동작은 같고 github-flow 는 `main` 이 되게 한다. D-30 은 같은 방향의 한 자리 더다 — M1 의 원장 밖에 있던 투영의 상수(원장 E-46). D-12 는 식별자와 스냅숏을 동결하지 값을 동결하지 않는다: 투영의 병합 목표는 **값**이고, 동결된 행동 식별자(`local-merge-develop`·`push-develop`·`local_develop_merge`)와 스냅숏은 이름이므로 값을 옮겨도 D-12 를 어기지 않는다. 고정된 스냅숏에서 `develop` 이 든 줄은 `ActionLocalMerge … "local_develop_merge"` 한 줄뿐이고 `MergeTarget` 은 필드 선언으로만 있다(원장 E-51). 그래서 이 변경이 스냅숏을 바꾸지 않는다는 것이 읽기 결과이고, 구현이 그 예측을 `go doc` 비교(limb 5)로 확인한다(M-11 이 도착 시점의 초록을 기록했다).

**같은 필드, 별개 생산 지점.** `mission.MissionContract.MergeTarget` 을 채우는 곳은 둘이다. `internal/cli/goal.go` 의 approve 는 M1 이 `LoadGitFlowIntegrationConfig(root).IntegrationTarget` 으로 옮겼고(원장 E-29), `internal/contract/projection_mission.go` 의 투영은 상수를 쓴다(E-46). 두 곳은 서로의 값을 받지 않는다. 투영에는 비테스트 호출자가 없고(E-48), 필드를 읽는 비테스트 코드는 봉인 전 비어 있지 않음 검사 한 줄뿐이며(E-49) 값은 계약 해시에 들어가지만 어떤 브랜치와도 비교되지 않는다.

**인자 방식을 고른 이유.** 계약 코어는 표준 라이브러리·`yaml.v3`·`internal/mission` 만 import 하고 `internal/config` 는 import 하지 않는다(원장 E-52) — 정책 값은 호출자가 넘기는 층위다. 투영이 구성을 직접 읽으면 순수 검증 코어에 파일 입출력이 들어와 훅에서 안전하게 부르던 성질이 깨진다. 호출자가 구성을 읽는 규약은 이미 있다: approve 가 `LoadGitFlowIntegrationConfig(root).IntegrationTarget` 을 읽어 비면 봉인을 거부한다. 투영을 운영에 연결하는 호출자는 같은 읽기를 인자로 넘긴다. 현재 호출자가 없으므로(E-48) 시그니처 변경의 파급은 시험 파일의 호출 지점 네 곳이다. 단순화 사다리의 첫 두 칸(만들 필요가 있나, 이미 있는 해석기를 재사용하나)이 답을 준다 — 해석기는 이미 있고 투영은 값만 받으면 된다.

**`internal/mission/git_owner.go:201` 은 범위 밖이다.** 그 줄은 `ActionLocalMerge`(동결된 식별자 `local_develop_merge`)의 로컬 병합 경계이며 현재 브랜치가 리터럴 `develop` 이 아니면 병합을 거부한다(E-50, 읽기만 했다). 이유 셋. (i) 이 경로는 `develop` 통합 워크트리로 병합하는 git-flow 전용 경로이고 github-flow 에서는 로컬 통합 창이 전달의 선행 조건이 아니므로(REQ-GFD-006) 휴면이 맞다 — D-1 이 세션 종료 자동 병합에 대해 낸 결론과 같다. 이 줄의 `develop` 을 구성된 목표로 옮기면 github-flow 에서 `main` 으로의 로컬 `git merge --no-ff` 가 켜져 REQ-GFD-004·006 과 주 체크아웃 규칙을 어긴다. (ii) 이 경계는 `MergeTarget` 값과 비교되지 않으므로(E-49) 투영의 값이 `main` 이 돼도 이 경계가 따라 움직일 필요가 없다. (iii) `internal/mission` 은 봉인된 검증기로 취급되고 스냅숏 시험이 그 노출 표면을 지키므로 편집 범위를 넓히면 감사 표면이 넓어진다. 가드는 M2a 종료 트리에서 `internal/mission` 이 무편집임을 보이는 범위 가드(AC-GFD-024 의 돌연변이 프로브 8·9, E-54·E-55)다. github-flow 구성에서 `local-merge-develop` 행동을 가진 미션 계약을 어떻게 다룰지(투영에서 거르기와 허용하되 휴면 두기)는 이 개정이 정하지 않는다 — 읽기 결과이고 실행해 관측하지 않았으며 `progress.md` §G 에 권고로 남긴다.

**기각된 대안.**

| 대안 | 기각 이유 |
|---|---|
| 리터럴을 두고 후속 카드로 권고만 남긴다 | 이 카드의 목적이 develop 에 고정된 표면을 모두 찾아 옮기는 일이다. 절체 뒤 github-flow 구성의 미션 계약이 틀린 병합 목표를 봉인하는 자리를 알면서 남기게 된다. 리더가 카드에 포함하기로 정했다 |
| 동결된 행동 식별자를 함께 개명한다(`local_develop_merge`·`local-merge-develop`) | D-12·D-21 이 호환 이주를 요구하는 후속 일로 정했고, 개명은 스냅숏 45 줄의 편집을 부른다. 값을 옮기는 일은 이름을 바꾸지 않고 된다 |
| 리터럴을 `"main"` 으로 바꾼다 | git-flow 구성(이 저장소의 현재 값)에서 투영 결과가 달라져 PRE-CUTOVER-SAFE 가 아니고 REQ-GFD-023 의 git-flow 바이트 동일을 어긴다 |
| 계약 코어가 `internal/config` 를 직접 읽는다 | 순수 검증 코어의 층위 위반(원장 E-52). 호출 쪽이 이미 구성을 읽는 규약이 있다 |
| 인자가 비면 `develop` 이나 `main` 으로 대체한다 | M1 옵션 A 와 REQ-GFD-003 의 "목표 없음 = 답 없음"을 어긴다. 틀린 브랜치를 봉인하는 조용한 오동작이다 |
| 패키지 수준 변수나 전역 설정으로 주입한다 | 전역 상태라 병행 호출과 시험 격리를 해친다. 인자 하나로 충분하다 |
| 옵션 구조체·함수형 옵션 | 인자가 하나뿐이라 과설계다(단순화 사다리 6번 — 한 줄이면 된다) |
| 스냅숏을 다시 생성해 갱신한다 | 값이 스냅숏에 들어 있지 않아 필요가 없다(E-51). 편집이 필요해 보이면 마일스톤을 멈추고 리더에게 올린다 — 손으로 다시 생성하는 길은 열지 않는다 |

### D-26 보충 — `[ZONE:Frozen]` 줄의 전수와 분류 (측정: 트리 `6c2277295d9ddeaa92e83c0225910445f83cc1c2`, 원장 E-38~E-41)

`git grep -n "ZONE:Frozen" -- <spec-workflow.md>` 는 로컬과 템플릿 사본에서 같은 세 줄(23·49·164)을 냈고, `moai constitution list --file …/spec-workflow.md`(이 트리에서 빌드한 `moai`)는 `CONST-V3R2-001`·`CONST-V3R5-027`·`CONST-V3R5-028` 셋을, 템플릿 사본 경로로는 `No entries.` 를 냈다(레지스트리는 로컬 경로만 안다 — 템플릿 사본은 AC-GFD-015 의 사본 정합 시험이 로컬과 같게 지킨다). 줄 번호는 이 트리의 값이고 편집이 줄 번호를 옮기므로 인용은 문언으로 한다.

| 줄(이 트리) | 문언 | 등재 | 취급 | 이 SPEC 에서 |
|---|---|---|---|---|
| 23 | `[ZONE:Frozen] [HARD] Every MoAI SPEC follows the three-phase lifecycle …` (Route A/B 문단, 줄 26 의 불릿 `opens a PR per phase` 가 이 문단 아래에 있다) | **미등재** | 일반 편집 | 두 경로(Route A/B) 서술을 github-flow 경로에 맞춘다(AC-GFD-017: `opens a PR per phase` 0) |
| 49 | `[ZONE:Frozen] [HARD] Step ordering rules:` (머리줄) | **미등재**(머리줄 자체) | 일반 편집 | 머리줄은 바꿀 필요가 없다 — 그 아래 두 불릿이 등재 clause 다 |
| 50 | `Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step …` | **등재 `CONST-V3R5-027`** | `amend`(운영자) | plan 단계의 워크트리 진입을 허용하도록 개정(AC-GFD-013) |
| 53 | `Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged …` | **등재 `CONST-V3R5-028`** | `amend`(운영자) | 단일 PR 폐기 조건으로 개정(AC-GFD-013) |
| 164 | `[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step. See § SPEC Phase Discipline (Step 1).` | **미등재** | 일반 편집 | plan 단계 금지 문장을 개정된 027 과 같은 방향으로 바꾼다. **`NO worktree at this step` 가 로컬·템플릿 두 사본에서 0**(AC-GFD-017) |
| 166 | `Create comprehensive specification using EARS format.` | **등재 `CONST-V3R2-001`**(`zone-registry.md` 의 clause, anchor `#plan-phase`) | 편집하지 않는다 | 164 줄의 편집이 166 줄을 건드리면 `amend` 대상이 된다 — 건드리지 않는다 |

`worktree-integration.md` 는 `moai constitution list --file` 이 `No entries.`(원장 E-33)이므로 그 파일의 `[ZONE:Frozen]` 줄(390·401 포함, 그 밖의 표지 줄은 이 SPEC 이 편집하지 않는다)은 모두 일반 편집이다. **일반 규칙**: M4 가 편집하는 파일마다 같은 `list --file` 를 돌려 등재 여부를 판정서에 적고, 등재되지 않은 `[ZONE:Frozen]` 줄은 `amend` 없이 일반 편집으로 처리하되 개정된 등재 문언과 어긋나지 않게 한다. 미등재 표지 줄을 지켜 주는 장치는 `amend` 가 아니라 AC-GFD-017 의 0-grep 과 AC-GFD-015 의 사본 정합 시험이다.

---

🗿 MoAI
