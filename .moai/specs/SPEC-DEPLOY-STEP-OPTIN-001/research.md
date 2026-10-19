# SPEC-DEPLOY-STEP-OPTIN-001 — 연구

이 문서는 plan 단계의 측정 기록이다. 입력 메모(`.moai/reports/t1895/OPTIN-DESIGN.md`, ADK 읽기 전용 조사, main `74ad197380` 기준)는 로컬 입력이며 gitignored 라서 이 저장소의 어떤 클론에도 없다. 메모의 사실은 **가설**로만 읽고, 이 SPEC 이 기대는 모든 사실은 아래 증거 원장(§8)에서 이 트리로 다시 측정했다.

## §1 방법과 기준

- **트리.** 모든 측정의 트리 SHA 는 `8249bff8c3933d647b4f99e59048595f9808e83a`(브랜치 `WT-opt-in-moai`)다(L01). 작성을 시작하기 전에 `git status --short` 를 한 번 실행했고 출력은 비어 있었다(종료 코드 0). 원장 L02 는 원장을 마지막으로 다시 만든 시점의 같은 명령이며, 그때는 SPEC 디렉터리만 추적되지 않은 항목으로 나온다.
- **명령 형식.** 원장의 모든 항목은 읽기 전용 단일 호출이다. 파이프·리다이렉트·`&&`·`;` 연쇄·서브셸이 없다. 각 항목이 명령, 원문 stdout, 종료 코드, 트리 SHA 를 가진다(verification-completeness §2.1). 항목은 스크립트(`python3 -I`, 셸 미사용)가 한 번에 실행해 기록했고 stdout 은 가공하지 않았다.
- **빈 결과의 해석.** 빈 stdout 과 종료 코드 1 을 부재의 증거로 쓰는 항목에는 같은 트리·같은 도구가 존재하는 토큰을 찾는 양성 대조를 둔다. 아래 표가 짝을 보인다.

| 빈 결과 항목 | 양성 대조 | 대조가 보이는 것 |
|---|---|---|
| L04 (`--draft` 없음) | L03 | 같은 파일에서 같은 도구가 호출 지점 네 줄을 찾는다 |
| L07 (이력에 t1882 없음) | L06 | 같은 형식의 제목(`card t1710`)이 이력에 있다. 이 짝은 추론이다(Gap 5) |
| L08 (정책 토큰 없음) | L03, L54 | 같은 파일에 호출 지점과 `verifyLease` 가 있다 |
| L29, L30 (웹·설정에 revision·감사 없음) | L28, L31 | 같은 디렉터리에서 같은 도구가 `saveMu`·읽기 함수를 찾는다 |
| L47, L48 (단계 키 없음) | L14, L15, L22 | 같은 범위에서 기존 토큰을 찾는다 |
| L57 (`--auto-source` 없음) | L58 | 같은 디렉터리에서 `add` 선언을 찾는다 |
| L59 (policy revision 없음) | L54 | 같은 범위에서 호출 지점을 찾는다 |
| L35 (지역 지침 템플릿 없음) | L36 | 같은 경로 형식에서 `AGENTS.md.tmpl` 이 추적 파일로 나온다 |

- **go test·go build 미실행.** 자원 규칙(호스트 부하)에 따라 plan 단계에서 실행하지 않았다. 모든 녹색 경로는 약속이다.
- **판정 빌드의 출처.** `moai spec lint` 를 판정한 바이너리는 `v3.1.3` 이라고만 출력하고 커밋과 빌드 시각은 `none` / `unknown` 이다(L38). 이 트리 HEAD 와의 조상 관계는 확인하지 못했다. verification-claim-integrity §2.2 에 따라 lint 결과는 "트리만 인용되고 판정 빌드의 커밋은 모름"인 측정이며 Gap 으로 적는다(§5 Gap 1).

## §2 전제 대 측정

판정 열: **유지** 는 메모의 사실이 이 트리에서도 참이다. **좌표 변경** 은 사실은 참이나 줄·파일이 다르다. **반증** 은 메모의 주장이 이 트리에서 거짓이다. **신규** 는 메모에 없던 사실이다.

| # | 전제 | 메모의 좌표 | 이 트리의 측정 | 원장 | 판정 |
|---|---|---|---|---|---|
| 1 | 전달 본체가 lease·version·SHA 를 재확인한 뒤 push → PR → auto-merge 를 잇는다 | `factory_card_pr.go:351`, :395, :410, :444 | 호출 지점 395, 410, 444, pr-open 기록 449. 재확인은 네 지점 모두 lease·version·만료뿐 | L03, L54 | 유지 |
| 2 | 읽은 main 의 PR 생성 인자에 `--draft` 가 없다 | (같은 곳) | `--draft` 없음, 병합 요청 `"--auto"` 그대로 | L04, L05 | 유지 |
| 3 | t1882 의 변경이 이 트리에 있는지 | "먼저 대조할 지점" | **없다.** 이력에 t1882 없음, 파일의 마지막 변경은 t1710·t1589 | L04–L07 | 확정: 미착지 |
| 4 | 정책을 읽는 지점이 없다 | (암시) | 전달 파일에 단계 정책 토큰 없음. 유일한 게이트는 워크플로 종류 | L08, L09 | 신규(확정) |
| 5 | `--no-auto-dispatch` 가 없으면 FactoryDispatchAuto | `factory.go:316` | `laneDispatchSelection` 이 :319, 반환이 :323 | L11 | 좌표 변경 |
| 6 | 배차 기본값의 성격 | (없음) | 설정 키가 아니라 실행마다 찍는 환경 변수 `MOAI_FACTORY_AUTO_DISPATCH`. 읽는 곳은 안내문 선택 둘. 인자 없는 `factory next` 는 우선순위 순서로 집는다 | L12, L45, L46 | 신규 |
| 7 | `project_continuation.go` 는 누락·잘못된 값에서 card | `:56` | 56–70 줄이 그대로 | L13 | 유지 |
| 8 | "project.continuation 키 누락" (카드 별도 항목) | "근거 메모 준비 중" | **키는 누락이 아니다.** 템플릿에 `continuation: card`, 웹 스키마 필드, Go 기본값이 모두 있다. 문제는 누락이 card 로 풀려 카드 발급 쪽으로 열린다는 것 | L13–L16 | 반증(카드 문언), 신규(실제 문제) |
| 9 | `todo.enabled` 는 안내 표시용, 부재는 켜짐 | `todo_enabled.go:12,29,46` | 29–34 줄이 부재·nil 에서 true | L17 | 유지 |
| 10 | git 기본 모드 team, init 이 remote 유무로 manual/personal | `defaults.go:1333`, `init_gitdetect.go:58`, `init.go:598` | `defaults.go:1333` team. `detectGitConfig` 가 manual(원격 없음)/personal(있음) | L60, L21 | 유지 |
| 11 | 템플릿 git-strategy 의 automation 값 | `.tmpl:39,72,111` | `auto_commit` 41·74·113, `auto_pr` 42·75·114, `auto_push` 43·76·115(team 만 true), `draft_pr` 100 | L18 | 좌표 변경 |
| 12 | automation 키가 동작과 이어진다 | (암시) | `AutoPush`/`AutoPR` 은 구조체와 기본값에만 있고 읽는 비테스트 코드가 없다 | L19, L20 | 신규 |
| 13 | 이미 수동인 항목 | worktree 토글, local_main_integration, candidate_ci, jev | 모두 false. review gate 도 템플릿에서 false | L22–L24 | 유지 |
| 14 | 승인된 로컬 factory 설정 | (일반 언급) | 이 저장소 배포본이 `auto_create`·`auto_cleanup` true, review gate true, git-strategy personal/team 의 `auto_pr`·`auto_push` true | L25, L53 | 신규(구체화) |
| 15 | `/settings`, `POST /save` 연결 | `app.go:194` | 207, 210 | L26 | 좌표 변경 |
| 16 | Host/same-origin 검사 | `app.go:255` | 257, 268 | L27 | 좌표 변경 |
| 17 | 검증 실패는 쓰기를 멈춘다 | `handlers.go:515` | 배너가 520, 저장은 `saveMu` 로 직렬화(381) | L28 | 좌표 변경 |
| 18 | 파일 단위 temp+rename | `settings/yamlpatch.go:518` | `yamlpatch` 는 **디렉터리**다. `yamlpatch/yamlpatch.go` 의 `atomicWriteMode`(518)·rename(554) | L32 | 좌표 변경 |
| 19 | 읽기 모델이 값·출처·override·revision 을 함께 돌려주지 못한다 | `sectionvalues.go:54` | `internal/settings/sectionvalues.go:54` 는 문자열 맵. revision 개념, 감사 기록 쓰기 코드 모두 없음 | L29–L31 | 유지(확정) |
| 20 | update 3-way 병합이 old == base 면 새 템플릿 값을 쓴다 | `node_merge.go:445` | :449 | L33 | 좌표 변경 |
| 21 | 템플릿에 없는 사용자 키 | (암시) | 모두 보존한다 | L34 | 신규(확정) |
| 22 | `AGENTS.local.md`·`CLAUDE.local.md` 는 배포 템플릿이 아니다 | tracked_template_files=0 | 세 경로 모두 추적되지 않음. 양성 대조 통과. `AGENTS.md.tmpl` 은 산문으로 6줄에서 언급 | L35, L36, L61 | 유지 |
| 23 | 시험 이음매 `factoryGHCall` | AC-07 | **이음매는 `factoryGH` 변수**다. `factoryGHCall` 은 일반 함수. 기존 시험은 `factoryGH` 를 교체하고 실제 `git push` 를 임시 bare 원격에 한다 | L62, L63 | 반증(이름) |
| 24 | 기존 설정 해석기의 local 계층 | (없음) | 8단 해석기는 있으나 호출자는 doctor 하나, local 경로를 CWD 상대로 읽는다. 이 트리에 `.moai/config/local` 없음 | L39–L42 | 신규 |
| 25 | 배포본에 게시(release/tag/deploy) 호출이 있다 | (release/deploy 단계 언급) | 비테스트 Go 코드에 없다(주석 둘) | L51 | 신규 |
| 26 | 템플릿-배포본 diff 0 | 카드 완료 판정 | 스킬·규칙 파일은 사본과 일치(L55). `.moai/config/sections/*.yaml` 은 운영자 자신의 파일이라 템플릿과 다르다(L56) | L55, L56 | 신규(범위 정정) |
| 27 | 리뷰 지적 수리 요청의 실행 경로 | "review repair" | 턴을 잇는 차단 `hook.DecisionBlock`(`codex_review_gate.go:426`) 하나 | L52 | 신규 |
| 28 | SPEC 도구에 마일스톤 순서 검증 동사 | (과업 지시) | 없다(`spec`·`verify` 도움말 전수) | L43, L44 | 신규 |

## §3 메모와 달라진 전제 요약

1. **t1882 미착지**(전제 3): 이 트리는 t1882 이전이다. M3 는 착지 뒤 재측정을 요구한다.
2. **"키 누락"은 반증**(전제 8): 카드 별도 항목은 키의 부재가 아니라 부재가 열린 쪽으로 푸는 해석기 방향의 문제다. 이웃 키 `todo.enabled` 도 같은 방향이다(전제 9).
3. **시험 이음매 이름**(전제 23): `factoryGHCall` 이 아니라 `factoryGH` 이고 push 는 가짜가 아니다. AC 의 "push 0회"는 호출 기록이 아니라 bare 원격의 ref 부재로 확인한다.
4. **`yamlpatch` 는 디렉터리**(전제 18), 여러 좌표가 이동했다(전제 5, 11, 15–17, 20).
5. **배차 기본값은 설정이 아니라 환경 스탬프**(전제 6): 메모의 "factory.go 의 기본 auto-dispatch" 는 맞으나, 정책을 도입하려면 키가 먼저 없다.
6. **automation 키는 장식**(전제 12): 메모가 보지 못한 사실이며 Q7 의 근거다.
7. **config 섹션 파일은 템플릿-사본 비교의 예외**(전제 26): 카드의 "템플릿-배포본 diff 0"은 `.claude/` 계열에 맞고 `.moai/config/sections/*.yaml` 에는 맞지 않는다. REQ-DSO-018 이 범위를 정정해 적었다.
8. **웹 쪽 메모 주장 가운데 이 plan 이 다시 읽지 않은 것**: `schemaform.go:383`(미제출 필드 보존·중복 제출 거부)은 다시 읽지 않았다(Gap 3).

## §4 관계 카드 본문의 관측 (원장 밖)

큐 본문은 `moai todo show <id>` 의 출력 읽기이며 이 plan 이 인용한 줄은 아래다. 원장 형식(단일 호출 원문 전체)을 따르지 않으므로 증거가 아니라 관계 기록으로만 쓴다.

- t1882: 상태 `picked`. 본문이 `internal/cli/factory_card_pr.go:395-460` 을 대상으로 draft 게시, auto-merge 요청 제거 또는 설정 키로 끄기, pr-open 은 PR 생성·확인만으로 기록, 리뷰 수정 복귀, 부분 게시 복구를 요구한다.
- t1752: 상태 `picked`. 본문이 `git_strategy.done_on` 신설과 `internal/web/projectconfig.go`·`schemaform.go` 에 선택 항목 추가를 요구한다.
- t1644: 상태 `hold`. 본문이 `/todo`·`/settings` 2페이지와 라우트 6개 삭제(`/profile/*` `/factory` `/monitor` `/kanban` `/specs` `/events`)를 요구한다.
- t1884: 상태 `dropped`, t1900 으로 재발급(AGENTS 전수 검토). 본문에 "배포 기본 수동·단계별 독립 opt-in … t1895 — 미구현은 구현 과제/한계로만 서술"이 있다.
- t1894: 상태 `picked`. 카드 순차/병렬 판단 지침. 본문이 `factory-dispatch-cards.md` 와 카드 발급 절차 문서를 대상으로 하고 코드는 t1737 몫이라고 한다.

## §5 Gaps (미관측)

1. **판정 빌드의 출처.** lint 를 판정한 바이너리의 커밋을 확인하지 못했다(L38: `v3.1.3`, `none`, `unknown`). lint 가 이 트리의 규칙을 반영하는지는 모른다.
2. **go test·go build 미실행.** 위 §1. 모든 녹색 경로와 변이체 탐침은 plan 시점에 실행되지 않았다.
3. **`schemaform.go` 미독.** 미제출 필드 보존과 중복 제출 거부(AC-DSO-017 (b) 의 기존 동작 근거)를 이 plan 이 다시 읽지 않았다. M4 사전 점검에서 읽는다.
4. **`internal/homestate` 레코드의 영속 형태 미독.** 전환 기록에 revision 필드를 더하는 비용(스키마 이전 여부)을 모른다.
5. **L07 의 양성 대조는 추론이다.** `--grep=t1882` 가 "있다면 찾는다"는 것을 같은 형식 문자열로 직접 시험하지 않았다. L04·L05·L06 이 독립적으로 미착지를 보인다.
6. **t1894 와의 파일 겹침**은 큐 본문 읽기 수준이다. 실제 파일 목록은 착지 전에는 알 수 없다.
7. **`yamlpatch` 키 제거 연산 미독.** 웹의 "기본값으로 되돌리기"가 쓰는 제거 연산이 이 트리에서 어떻게 동작하는지(`remove_test.go` 파일 존재만 확인)를 읽지 않았다.
8. **local 계층 파일 이름.** `loadLocalTier` 가 `.moai/config/local/*.yaml` 을 섹션 이름으로 읽는 것까지 읽었고(L40 주변), 실제 `workflow.yaml` 파일로 동작을 시험하지 않았다.
9. **`moai doctor` 부모 연결.** `doctor_config.go` 의 `Use: "config"` 만 봤다. Q12 의 (a) 가 성립하는지는 M2 에서 확인한다.
10. **사용자 역할별 ACL, 실행 중 레인에 OFF 전달**: 메모도 미검증이라 했고 이 plan 도 측정하지 않았다. 설계는 경계 재판정으로 대체한다.
11. **설치 바이너리, 새 init/update 의 실제 출력, 브라우저 저장, 실제 외부 효과**: 모두 미관측이다(메모와 같다).
12. **거절된 명령.** 한 번의 복합 Bash 호출(git 과 python 을 한 줄에 묶음)이 워크트리 가드에 거절되었다. 같은 명령을 재시도하지 않고 python 실행과 git 호출을 별개의 호출로 나눴다. 대체한 호출의 결과는 원장 L01–L63 이며 그 항목들이 측정이다.

## §6 잔여 위험

- 줄 번호는 이동한다. t1882 착지 뒤 L03–L10 을 반드시 다시 측정한다.
- 템플릿에 주석 블록만 싣는 설계는 update 의 주석 병합이 블록을 복제하거나 잃을 수 있다. 주석은 적재량이 아니므로 동작 위험은 아니나 시험(AC-DSO-014)이 관측해야 한다.
- AC-DSO-005, 014, 015, 021 은 회귀 가드로 오늘 GREEN 이거나 빈 스윕이다. 변이체 탐침과 양성 대조 시험이 없으면 아무것도 단언하지 못한다.
- 14개 결정 행 가운데 열린 11개가 닫히기 전에는 AC-DSO-002, 008, 009, 010, 011 의 기대값이 확정되지 않는다.
- 이 plan 이 작성한 RED-now 셀은 이 트리의 부재·결함을 보인다. 구현이 다른 위치(예: 다른 파일)에 판정을 둔다면 L08 같은 파일 단위 명령의 "flip" 이 의미를 잃으므로 flip 은 시험 PASS 로 판정한다(원장 명령 flip 은 보조 증거).

## §7 결정 색인과의 연결

14개 결정의 근거 측정은 위 §2 의 전제 번호가 가리키는 원장 항목이다. 결정 Q1(전제 11–13, 19), Q2(전제 9, 7), Q3(전제 27, 13, 14), Q4(전제 25), Q5(전제 9), Q6(전제 7–8), Q7(전제 11–12), Q8(전제 14), Q9(전제 17), Q11(전제 3), Q14(전제 5–6).

## §8 증거 원장

항목마다 명령, 원문 stdout, 종료 코드, 트리 SHA 가 있다. `stderr bytes` 는 참고다. 모든 항목의 트리는 `8249bff8c3933d647b4f99e59048595f9808e83a` 다.

#### L01 — tree identity

```text
$ git rev-parse HEAD
8249bff8c3933d647b4f99e59048595f9808e83a
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L02 — tree status at ledger generation: only the SPEC directory is new (the same command printed nothing before authoring began)

```text
$ git status --short
?? .moai/specs/SPEC-DEPLOY-STEP-OPTIN-001/
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L03 — delivery edge: push, PR create, auto-merge, pr-open call sites (memo said :395 :410 :444)

```text
$ grep -n -E '"push", "origin"|"pr", "create"|"pr", "merge"|To: homestate.CardPROpen' internal/cli/factory_card_pr.go
395:	if _, err := factoryGitRead(wt, "push", "origin", checkedTip+":refs/heads/"+cardBranch); err != nil {
410:		if _, err := factoryGHCall(wt, "pr", "create", "--base", target, "--head", cardBranch, "--title", title, "--body", body); err != nil {
444:		if _, err := factoryGHCall(wt, "pr", "merge", strconv.Itoa(pr.Number), "--auto", mergeFlag, "--match-head-commit", checkedTip); err != nil {
449:		RunID: runID, CardID: card.CardID, To: homestate.CardPROpen,
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L04 — t1882 draft-flag change is NOT in this tree: no --draft in the delivery edge

```text
$ grep -n -- --draft internal/cli/factory_card_pr.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L05 — t1882 auto-merge removal is NOT in this tree: --auto merge request still present

```text
$ grep -n -- '"--auto"' internal/cli/factory_card_pr.go
444:		if _, err := factoryGHCall(wt, "pr", "merge", strconv.Itoa(pr.Number), "--auto", mergeFlag, "--match-head-commit", checkedTip); err != nil {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L06 — last commits touching the delivery edge (no t1882 commit)

```text
$ git log --oneline -5 -- internal/cli/factory_card_pr.go
34ecf2418c fix(cli): run the delivery gh in the judged root and fold the GH_REPO drop on Windows (card t1710, sync-audit F1/F2)
94e2d838ee fix(factory,cli): scrub the inherited git env and drop GH_REPO in factory delivery paths (card t1710)
65a1706b55 fix(factory): gh runs with the git environment scrubbed (card t1589)
1ce50c5fd2 test(factory): RED — the gh environment must not carry an inherited GIT_DIR (card t1589)
4faf5ff818 fix(factory): scrub the git environment on the delivery edge's reads (card t1589)
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L07 — no commit in history mentions t1882

```text
$ git log --oneline -3 --grep=t1882

```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L08 — the delivery edge has no stage-policy read (no auto_push/auto_pr/stage/opt-in token)

```text
$ grep -n -i -E 'auto_push|auto_pr|AutoPush|AutoPR|opt-in|optin|StagePolicy' internal/cli/factory_card_pr.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L09 — the only gate on the delivery edge is the workflow kind

```text
$ sed -n 106,111p internal/cli/factory_card_pr.go
// factoryGitHubFlow reports whether the project's active git strategy is
// github-flow — the only configuration the delivery edge is live under. An
// absent or unreadable git-strategy.yaml, and every other workflow, is not.
func factoryGitHubFlow(root string) bool {
	return config.LoadGitFlowIntegrationConfig(root).Workflow == config.WorkflowGitHubFlow
}
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L10 — who calls the delivery body

```text
$ grep -n factoryDeliverByPR( internal/cli/factory_card_pr.go
231:		return factoryDeliverByPR(ctx, out, db, root, runID, card, cardBranch, target, cfg.MergeMethod, lane)
272:func factoryDeliverByPR(ctx context.Context, out io.Writer, db *homestate.FactoryDB, root, runID string, card homestate.Card, cardBranch, target, method, lane string) error {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L11 — factory.go dispatch default (memo said :316 :323)

```text
$ grep -n -E 'FactoryDispatchAuto|noAutoDispatchFlag +=|func laneDispatchSelection' internal/cli/factory.go
82:	noAutoDispatchFlag = "--no-auto-dispatch"
319:func laneDispatchSelection(entry launcherEntryParse) string {
323:	return config.FactoryDispatchAuto
844:// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011): config.FactoryDispatchAuto
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L12 — auto-dispatch is a per-launch env stamp, not a config key

```text
$ sed -n 429,446p internal/config/envkeys.go
	// EnvFactoryAutoDispatch carries the auto-dispatch selection a factory
	// lane launch made (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011): the
	// launcher stamps it into the lane session's environment and the
	// SessionStart lane rule reads it. The DEFAULT is auto-dispatch and is
	// recorded in code, not in this carrier — the stamp always overwrites
	// (the clear-policy discipline) so an outer session's value cannot leak
	// into a lane launched without one. Any value other than
	// FactoryDispatchManual reads as the default.
	EnvFactoryAutoDispatch = "MOAI_FACTORY_AUTO_DISPATCH"

	// FactoryDispatchAuto is the stamped default: the lane enters its
	// factory-next self-dispatch loop without a per-card lead routing step.
	FactoryDispatchAuto = "auto"

	// FactoryDispatchManual is the --no-auto-dispatch opt-out stamp: the lane
	// receives a manual-mode rule instead of the next-card rule.
	FactoryDispatchManual = "manual"
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L13 — project.continuation resolver: absent or unmatched resolves to card

```text
$ sed -n 56,70p internal/config/project_continuation.go
func (c *Config) ProjectContinuation() (value string, unmatched string) {
	if c == nil {
		return ProjectContinuationCard, ""
	}
	configured := c.Workflow.Project.Continuation
	if configured == "" {
		return ProjectContinuationCard, ""
	}
	for _, valid := range ValidProjectContinuations() {
		if configured == valid {
			return configured, ""
		}
	}
	return ProjectContinuationCard, configured
}
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L14 — project.continuation Go default is card

```text
$ grep -n ProjectContinuationCard internal/config/defaults.go
1511:		Project: WorkflowProjectConfig{Continuation: ProjectContinuationCard},
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L15 — project.continuation ships in the template with value card

```text
$ grep -n -B1 'continuation: card' internal/template/templates/.moai/config/sections/workflow.yaml
54-    project:
55:        continuation: card
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L16 — project.continuation is a web schema field

```text
$ sed -n 390,392p internal/settings/schema_sections.go
		withOptionDesc(closedSeam(SectionWorkflow, "workflow", "f.workflow.project.continuation.opt.",
			config.ValidProjectContinuations(), "", "", "workflow", "project", "continuation"),
			"f.workflow.project.continuation.option."),
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L17 — todo.enabled: absent resolves to enabled, load error resolves to enabled

```text
$ sed -n 29,34p internal/config/todo_enabled.go
func (c *Config) TodoEnabled() bool {
	if c == nil || c.Workflow.Todo.Enabled == nil {
		return true
	}
	return *c.Workflow.Todo.Enabled
}
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L18 — git-strategy template automation defaults and draft_pr

```text
$ grep -n -E 'auto_(commit|pr|push):|draft_pr:' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
41:      auto_commit: true
42:      auto_pr: false
43:      auto_push: false
74:      auto_commit: true
75:      auto_pr: false              # Manual PR for review (default: disabled)
76:      auto_push: false            # Do not auto-push (default: disabled)
100:    draft_pr: true
113:      auto_commit: true
114:      auto_pr: false
115:      auto_push: true
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L19 — Go defaults for the three profiles

```text
$ grep -n -E AutoPush: internal/config/defaults.go
1354:			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
1367:			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
1383:			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: true},
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L20 — AutoPush/AutoPR have no non-test Go reader outside the struct and defaults

```text
$ grep -rn -E '\.AutoPush|\.AutoPR|AutoPush:|AutoPR:' internal --include=*.go --exclude=*_test.go
internal/config/defaults.go:1354:			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
internal/config/defaults.go:1367:			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
internal/config/defaults.go:1383:			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: true},
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L21 — init picks manual (no remote) or personal (remote) from remote presence

```text
$ sed -n 50,66p internal/cli/init_gitdetect.go
//
// Mode: a directory with no configured remote — including a non-git directory —
// yields "manual"; at least one remote yields "personal".
//
// Provider: derived from the origin remote's host. A github.com host yields
// "github"; any other host yields "gitlab" (self-hosted GitLab instances use
// arbitrary hostnames). When origin is absent or its host cannot be parsed,
// the provider falls back to the "github" default.
func detectGitConfig(dir string) (mode, provider string) {
	provider = gitProviderGitHub

	out, err := gitRemoteListFunc(dir)
	if err != nil || strings.TrimSpace(out) == "" {
		return gitModeManual, provider
	}
	mode = gitModePersonal
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L22 — template workflow.yaml worktree toggles (already manual)

```text
$ grep -n -E 'auto_create:|auto_merge:|auto_cleanup:|execution_mode:' internal/template/templates/.moai/config/sections/workflow.yaml
20:    execution_mode: auto
59:    # auto_create: read once (internal/cli/worktree_advisory.go) only to select
63:    # auto_cleanup: read by the two auto-cleanup paths (session-exit disposal
65:    # auto_merge: read by the session-exit auto-merge path — when true, a clean
71:        auto_create: false
72:        auto_merge: false
73:        auto_cleanup: false
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L23 — template workflow.yaml codex.review_gate ships off

```text
$ sed -n 160,163p internal/template/templates/.moai/config/sections/workflow.yaml
    codex:
        review_gate:
            enabled: false
            # tree_scope: review  # review | skip
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L24 — template local_main_integration, candidate_ci, jev blocks

```text
$ grep -n -A3 -E '^    (local_main_integration|candidate_ci|jev):' internal/template/templates/.moai/config/sections/workflow.yaml
248:    local_main_integration:
249-        enabled: false
250-    # commit_identity_guard: refuse a shell command that would create a commit
251-    # under a known test-fixture identity (an email your test suites use for
--
299:    jev:
300-        enabled: false
301-    # autonomy: how a SPEC's autonomous-execution contract is signed and bounded.
302-    # An absent key takes the default shown here; an unrecognized value for
--
348:    candidate_ci:
349-        enabled: false
350-        guard_bundle_required: true
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L25 — deployed (this repo) local overrides differ from the template defaults

```text
$ grep -n -E 'auto_create:|auto_cleanup:|auto_pr:|auto_push:|draft_pr:' .moai/config/sections/workflow.yaml .moai/config/sections/git-strategy.yaml
.moai/config/sections/workflow.yaml:167:        # auto_cleanup: true → false (SPEC-WORKTREE-REAPER-001 M1 착지 조건).
.moai/config/sections/workflow.yaml:174:        auto_cleanup: true
.moai/config/sections/workflow.yaml:175:        auto_create: true
.moai/config/sections/git-strategy.yaml:19:        draft_pr: false
.moai/config/sections/git-strategy.yaml:39:            auto_pr: false
.moai/config/sections/git-strategy.yaml:40:            auto_push: false
.moai/config/sections/git-strategy.yaml:59:        draft_pr: false
.moai/config/sections/git-strategy.yaml:66:            auto_pr: true
.moai/config/sections/git-strategy.yaml:67:            auto_push: true
.moai/config/sections/git-strategy.yaml:86:        draft_pr: false
.moai/config/sections/git-strategy.yaml:93:            auto_pr: true
.moai/config/sections/git-strategy.yaml:94:            auto_push: true
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L26 — web routes: /settings and /save

```text
$ grep -n -E 'HandleFunc\("/(settings|save)"' internal/web/app.go
207:	mux.HandleFunc("/settings", a.handleIndex)
210:	mux.HandleFunc("/save", a.handleSave)
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L27 — web gate: loopback Host and Sec-Fetch-Site same-origin

```text
$ grep -n -E '!isLoopbackHost|Sec-Fetch-Site\"\) != ' internal/web/app.go
257:		if !isLoopbackHost(r.Host) {
268:			if r.Header.Get("Sec-Fetch-Site") != "same-origin" {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L28 — web save is serialized and validation failure writes nothing

```text
$ grep -n -E 'saveMu\.Lock|Validation failed .{1,3} no changes were saved' internal/web/handlers.go
381:	a.saveMu.Lock()
520:		view.Banner = "Validation failed — no changes were saved."
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L29 — web has no revision concept (zero hits)

```text
$ grep -rn -i revision internal/web internal/settings --include=*.go --exclude=*_test.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L30 — web has no audit-record writer for saves (zero hits)

```text
$ grep -rn -i -E 'auditlog|audit_log|AuditEntry|AuditRecord|audit\.jsonl' internal/web internal/settings --include=*.go --exclude=*_test.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L31 — web read model returns scalars only (no source or override)

```text
$ grep -n -E '^func (SchemaCurrentValues|RawBlockValues)' internal/settings/sectionvalues.go
54:func SchemaCurrentValues(projectRoot string) (map[string]string, error) {
92:func RawBlockValues(projectRoot string) (map[string]string, error) {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L32 — yamlpatch is a directory in this tree (memo cited a file path)

```text
$ grep -n -E 'os\.Rename|func .*[Aa]tomic' internal/settings/yamlpatch/yamlpatch.go
512:func atomicWrite(path string, data []byte) error {
518:func atomicWriteMode(path string, data []byte, perm os.FileMode) error {
554:	if err := os.Rename(tmpName, path); err != nil {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L33 — update merge: user-unchanged-vs-base takes the new template (memo said :445)

```text
$ grep -n -B1 -A2 'nodeValuesEqual(oldV, baseV)' internal/cli/update/backup/node_merge.go
448-		chosen := oldV
449:		if baseExists && nodeValuesEqual(oldV, baseV) {
450-			chosen = newV
451-		}
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L34 — update merge: old-only keys are all retained

```text
$ grep -n REQ-UYP-006 internal/cli/update/backup/node_merge.go
390:// appends old-only keys at the end (REQ-UYP-006). Key order comes from the
455:	// Second pass: old-only keys (absent from the new template). REQ-UYP-006
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L35 — no AGENTS.local.md / CLAUDE.local.md / CLAUDE.md shipped template (empty = untracked)

```text
$ git ls-files -- internal/template/templates/AGENTS.local.md internal/template/templates/CLAUDE.local.md internal/template/templates/CLAUDE.md

```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L36 — positive control for L35: the shipped AGENTS.md.tmpl is tracked

```text
$ git ls-files -- internal/template/templates/AGENTS.md.tmpl
internal/template/templates/AGENTS.md.tmpl
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L37 — existing tests that the candidate AC set extends (all exist)

```text
$ ls internal/cli/factory_autodispatch_test.go internal/cli/factory_card_pr_test.go internal/config/project_continuation_test.go internal/config/todo_enabled_test.go internal/web/host_gate_test.go internal/settings/write_safety_test.go internal/cli/init_quiet_wizard_test.go internal/cli/update/backup/node_merge_test.go
internal/cli/factory_autodispatch_test.go
internal/cli/factory_card_pr_test.go
internal/cli/init_quiet_wizard_test.go
internal/cli/update/backup/node_merge_test.go
internal/config/project_continuation_test.go
internal/config/todo_enabled_test.go
internal/settings/write_safety_test.go
internal/web/host_gate_test.go
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L38 — SPEC lint judging binary: build identity

```text
$ <judging-binary> version
╭───────────────────╮
│                   │
│  moai-adk v3.1.3  │
│                   │
│                   │
╰───────────────────╯
 v3.1.3   none   built unknown 
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L39 — existing 8-tier resolver: interface carries Key(section, field) with a Source

```text
$ sed -n 27,34p internal/config/resolver.go
type SettingsResolver interface {
	// Load reads all 8 tier sources and produces merged settings.
	Load() (*MergedSettings, error)

	// Key retrieves a single key's value from the merged settings.
	Key(section, field string) (Value[any], bool)

	// Dump writes the merged settings to the provided writer.
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L40 — the 8-tier resolver reads CWD-relative local paths, not a project root

```text
$ sed -n 391,394p internal/config/resolver.go
func (r *resolver) loadLocalTier() (map[string]any, string, error) {
	settingsPath := ".claude/settings.local.json"
	localPath := ".moai/config/local"
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L41 — the 8-tier resolver has one production caller (doctor), none in factory or web

```text
$ grep -rn config.NewResolver() internal --include=*.go --exclude=*_test.go
internal/cli/doctor_config.go:50:	resolver := config.NewResolver()
internal/cli/doctor_config.go:102:	resolver := config.NewResolver()
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L42 — this tree has no .moai/config/local directory

```text
$ ls .moai/config/local

```
exit code: 1  
stderr bytes: 50  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L43 — SPEC tooling surface: no milestone-ordering verb under moai spec

```text
$ <judging-binary> spec --help

  Manage SPEC documents in .moai/specs/ directory.                                                                      
         
  USAGE  
         
    moai spec [command] [--flags]  
            
  COMMANDS  
            
    status <SPEC-ID> <new-status> [--flags]                          Update or list SPEC document statuses
    drift [--flags]                                                  Detect SPEC status drift between frontmatter and git log
    view <SPEC-ID> [--flags]                                         View acceptance criteria in tree structure
    lint [SPEC-ID | path/to/spec.md | SPEC directory ...] [--flags]  Lint SPEC documents for EARS compliance and structural validity
    close SPEC-ID [--flags]                                          Atomic 4-phase close for a SPEC (spec.md status: completed + progress.md backfill)
    audit [--flags]                                                  Audit SPEC era classification and detect modern-era status drift
    archive [--flags]                                                Archive finished SPECs out of .moai/specs/
    ceiling <SPEC-ID> [--evidence <dir>]... [--record] [--flags]     Evaluate a SPEC's plan-audit iteration ceiling
         
  FLAGS  
         
    -h --help                                                        Help for spec
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L44 — SPEC tooling surface: moai verify has audit-plan only (no ordering verb)

```text
$ <judging-binary> verify --help

  Shared diagnostic snapshot contract.                                                                                  
                                                                                                                        
  Quality-check results (tests, lint, coverage, ...) recorded under the current                                         
  working-tree key can be reused by sibling verification layers instead of                                              
  re-executing, under a strict freshness rule: the stored key must equal the key                                        
  recomputed from the current tree AND the recorded timestamp must be within the                                        
  TTL (default 10 minutes). A stale snapshot is never citable evidence — on any                                         
  mismatch the consumer re-executes the check.                                                                          
                                                                                                                        
  Verbs:                                                                                                                
    verify record   record one executed check result under the current-tree key                                         
    verify check    freshness query — exit 0 fresh / exit 1 stale                                                       
    verify run      run a command, or reuse a passing result recorded for the same tree state                           
    verify sync-gate     run the sync-phase gate checks, record a receipt (Codex Stop chain)                            
    verify codex-review  run the codex review of uncommitted changes, record a receipt (Codex Stop chain)               
         
  USAGE  
         
    moai verify [command] [--flags]  
            
  COMMANDS  
            
    record [--flags]                Record one executed check result under the current-tree key
    check [--flags]                 Freshness query for the current-tree snapshot (exit 0 fresh / exit 1 stale)
    run  -- <command...> [--flags]  Run a command, or reuse a passing result recorded for the same tree state
    audit-plan [--flags]            Print the resolved audit backend plan of one tree (read-only); --result-file checks an audit_multi result file against it
    codex-review                    Run the codex review of uncommitted changes and record a receipt (Codex Stop chain)
    sync-gate                       Run the sync-phase quality gate checks and record a receipt (Codex Stop chain)
         
  FLAGS  
         
    -h --help                       Help for verify
    --project-root                  Project root (default: $CLAUDE_PROJECT_DIR or cwd)
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L45 — factory next: bare takes the priority-order choice (self-dispatch), --card leases a nominee; no stage gate

```text
$ sed -n 1702,1708p internal/cli/factory_card.go
// newFactoryNextCommand — `moai factory next [--card <id>] [--wait]
// [--wait-bound <d>]` (REQ-SD-008/-009/-010): a lane session's self-dispatch
// verb, run from the parent checkout. Bare, it takes the CLI's priority-order
// choice; with --card it leases exactly the nominated card or refuses it
// (SPEC-TODO-AUTO-PICK-001: exit status 4, one line
// `factory next: refused <token>: <detail>` on the error stream).
func newFactoryNextCommand() *cobra.Command {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L46 — the two readers of the auto-dispatch stamp are guidance selectors, not lease gates

```text
$ grep -rn EnvFactoryAutoDispatch internal/cli/factory_card.go internal/hook/session_start_factory.go
internal/cli/factory_card.go:2766:	if os.Getenv(config.EnvFactoryAutoDispatch) == config.FactoryDispatchManual {
internal/hook/session_start_factory.go:154:		if os.Getenv(config.EnvFactoryAutoDispatch) == config.FactoryDispatchManual {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L47 — no stage-policy resolver or stage key exists anywhere in non-test Go (zero hits)

```text
$ grep -rn -i -E 'StagePolicy|stage_policy|review_repair|release_deploy|todo_creation|pr_creation' internal --include=*.go --exclude=*_test.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L48 — no stage keys in the shipped workflow.yaml (zero hits)

```text
$ grep -n -i -E 'review_repair|release_deploy|todo_creation|pr_creation|automation_stages' internal/template/templates/.moai/config/sections/workflow.yaml

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L49 — ignore rules: local tier directory and logs are untracked in the repo

```text
$ grep -n -E '^\.moai/config/local/|^\.moai/logs/' .gitignore
282:.moai/config/local/
400:.moai/logs/
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L50 — ignore rules in the shipped project template: logs and state are untracked

```text
$ grep -n -E '^\.moai/logs/\*|^\.moai/state/' internal/template/templates/.gitignore
242:.moai/state/
253:.moai/logs/*
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L51 — no distributed Go path publishes a release or tag (the only hits are two comments)

```text
$ grep -rn -E '"release", "(create|upload|edit)"|"tag", "-[a-z]|gh release|goreleaser' internal --include=*.go --exclude=*_test.go
internal/update/checker.go:188:	// Find the platform-specific archive URL matching goreleaser format.
internal/cli/doctor.go:516:// across CI runners (gh release lag differs between ubuntu-latest and macOS).
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L52 — the codex review gate's turn-keeping block (the review-repair side effect)

```text
$ grep -n hook.DecisionBlock internal/cli/codex_review_gate.go
426:		Decision: hook.DecisionBlock,
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L53 — this repo's deployed workflow.yaml turns the review gate on (a local operator setting)

```text
$ sed -n 317,319p .moai/config/sections/workflow.yaml
    codex:
        review_gate:
            enabled: true
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L54 — the delivery edge re-verifies lease, version, expiry before each remote mutation; none of the four re-reads policy

```text
$ grep -n verifyLease( internal/cli/factory_card_pr.go
377:	if err := verifyLease("the push"); err != nil {
398:	if err := verifyLease("before the pull request"); err != nil {
406:		if err := verifyLease("the pull request"); err != nil {
439:	if err := verifyLease("the auto-merge request"); err != nil {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L55 — mirror parity: a shipped skill file equals its project mirror byte for byte

```text
$ diff -q internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md .claude/skills/moai/workflows/project/doc-generation.md

```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L56 — config parity exemption: the deployed workflow.yaml is the operator's own file and differs from the template

```text
$ diff -q internal/template/templates/.moai/config/sections/workflow.yaml .moai/config/sections/workflow.yaml
Files internal/template/templates/.moai/config/sections/workflow.yaml and .moai/config/sections/workflow.yaml differ
```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L57 — todo add has no automatic-producer marker (zero hits)

```text
$ grep -rn auto-source internal/cli --include=*.go --exclude=*_test.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L58 — todo add verb declaration

```text
$ sed -n 846,850p internal/cli/todo.go
	var scan *todoAddScan
	cmd := &cobra.Command{
		Use:                "add <text>",
		Short:              "Append a card to the backlog queue",
		DisableFlagParsing: true,
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L59 — no transition or record carries a policy revision (zero hits)

```text
$ grep -rn -i -E 'policy_revision|PolicyRevision' internal --include=*.go --exclude=*_test.go

```
exit code: 1  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L60 — Go default git mode is team (memo: defaults.go:1333)

```text
$ grep -n 'Mode:           "team"' internal/config/defaults.go
1333:		Mode:           "team",
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L61 — AGENTS.md.tmpl mentions AGENTS.local.md in prose (count of mentioning lines)

```text
$ grep -c AGENTS.local.md internal/template/templates/AGENTS.md.tmpl
6
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L62 — the gh seam is the package variable factoryGH (memo said factoryGHCall; that is a plain function)

```text
$ grep -n -E '^var factoryGH =|^func factoryGHCall|^func factoryGitRead' internal/cli/factory_card_pr.go
50:var factoryGH = func(ctx context.Context, dir string, args ...string) ([]byte, error) {
130:func factoryGHCall(dir string, args ...string) ([]byte, error) {
170:func factoryGitRead(dir string, args ...string) (string, error) {
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a

#### L63 — existing tests swap factoryGH and run a real push against a temp bare origin

```text
$ grep -n -E 'factoryGH = d\.run|init -q|--bare|remote.{1,6}add.{1,6}origin' internal/cli/factory_card_pr_test.go
81:		fcGit(t, filepath.Dir(f.bare), "init", "-q", "--bare", "-b", "main", f.bare)
82:		fcGit(t, root, "remote", "add", "origin", f.bare)
190:	factoryGH = d.run
```
exit code: 0  
stderr bytes: 0  
tree: 8249bff8c3933d647b4f99e59048595f9808e83a
