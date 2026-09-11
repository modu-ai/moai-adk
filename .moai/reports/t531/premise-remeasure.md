# t531 — 카드 전제 재측정 (plan 착수 전)

- 카드: t531 · 트리 `.claude/worktrees/t531` · 브랜치 `WT-claudelocal-push-model`
- base: `bce6d7e08` (= `origin/develop` tip, 진입 시점 `git fetch` 후 실측)
- 측정일: 2026-09-08

## Claim

카드는 **두 사본**(main / develop)이 반대 지시를 한다고 서술한다. 실제로는 **세 변종**이고,
그중 문제의 지시를 담은 것은 **어느 브랜치에도 없는 primary 체크아웃의 미커밋 편집본**이다.
그리고 카드의 **축 2(흡수 대상)는 갈라짐이 아니라 두 사본 공통의 결함**이다.

## Evidence

### 세 변종

| # | 좌표 | §4.1 제목 | `push origin develop` 출현 | 모델 |
|---|---|---|---|---|
| 1 | `origin/main` = `main` = `7ad9f8534` (커밋된 블롭) | `### §4.1 로컬 통합 레인 (develop)` | **0** | **폐기된 모델** — 「develop은 원격에 올리지 않는다」 |
| 2 | primary 워킹트리 (미커밋) | `### §4.1 GitFlow 통합 체인 (develop)` | **3** | 레인이 창 안에서 `git push origin develop` |
| 3 | `origin/develop` 블롭 | `### §4.1 GitFlow 통합 체인 (develop)` | **2** (둘 다 리드 일괄 절차 안) | **리드 일괄 push** — 레인은 push 안 함 |

측정 명령과 출력:

```
$ git rev-parse --short origin/main ; git rev-parse --short main
7ad9f8534
7ad9f8534

$ git show origin/main:CLAUDE.local.md | /usr/bin/grep -c 'push origin develop'
(exit 1 — 0건)

$ git show origin/main:CLAUDE.local.md | sed -n '262p'
### §4.1 로컬 통합 레인 (develop)

$ git show origin/main:CLAUDE.local.md | /usr/bin/grep -n '§4.1\|push'
278:1. **`develop`은 원격에 올리지 않는다.** push 금지, upstream 설정 금지. …
296:# 통과하면 카드 브랜치를 push하고 main PR — develop은 push하지 않는다

$ /usr/bin/grep -c 'push origin develop' /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md
3
$ /usr/bin/grep -n '^### §4.1' /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md
262:### §4.1 GitFlow 통합 체인 (develop)

$ git show origin/develop:CLAUDE.local.md | /usr/bin/grep -c 'push origin develop'
2

$ git show origin/develop:CLAUDE.local.md > /tmp/dev-claudelocal.md
$ diff -q /tmp/dev-claudelocal.md /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md
Files … differ    (rc=1)

$ git diff --stat origin/main origin/develop -- CLAUDE.local.md
 CLAUDE.local.md | 261 +++++++++++++++++++--------- 196 insertions(+), 65 deletions(-)
```

세션 시작 시 기록된 `git status` 에 ` M CLAUDE.local.md` 가 있다. 그 파일은 추적 대상이고
main 블롭과 내용이 다르므로(위 0 vs 3), **변종 2 는 미커밋 워킹트리 편집본**이다.

### 축 1 — 카드의 귀속이 틀렸다

카드는 「primary(main) 사본 `:305` 가 창 절차에 레인의 `git push origin develop` 을 넣는다」고
적는다. **`main` 의 커밋된 사본에는 그 문장이 없다** — 그 사본의 §4.1 은 제목부터 다르고
(`로컬 통합 레인`), 내용은 **폐기된 모델**이다(`develop` 을 원격에 올리지 않고 카드별로 main PR).

그 문장이 실재하는 곳은 **변종 2**, 즉 primary 체크아웃의 미커밋 편집본이다. 갈라짐은
「main 대 develop」이 아니라 **「미커밋 워킹 사본 대 develop」**이고, 그 위에 **「main 커밋본은
폐기된 제3의 모델」**이라는 층이 하나 더 있다.

### 축 2 — 갈라짐이 아니라 공통 결함

`git merge origin/develop` 흡수는 **변종 2 와 변종 3 둘 다**에 있다:

```
$ diff /tmp/s41-primary.md /tmp/s41-dev.md | /usr/bin/grep 'merge origin'
< - 창을 받으면: … 본인 워크트리에서 `git merge origin/develop` 흡수 → …
> - 창을 받으면: … 본인 워크트리에서 `git merge origin/develop` 흡수 → …
```

양쪽이 같으므로 두 사본을 일치시켜도 이 결함은 남는다. 리드 일괄 push 모델에서 로컬
`develop` 은 `origin/develop` 보다 앞설 수 있으므로(레인들이 로컬 병합했고 리드가 아직
push 하지 않은 구간), `origin/develop` 을 흡수하면 **다른 레인의 착지분을 빠뜨린 베이스에서
재측정**하게 된다. 수리 형태가 「사본 일치」가 아니라 **「정본 한 곳의 내용 정정」**이다.

## Baseline-attribution

전부 이 트리(`.claude/worktrees/t531`, base `bce6d7e08`)에서 2026-09-08 에 실행했다. 카드
본문의 행번호(`:305` / `:338` / `:376` / `:379`)는 사용하지 않았다 — 리드 권고대로 절
구분자(`sed -n '/^### §4.1/,/^## 5\./p'`)로 범위를 잡았다.

## Gaps (관측하지 않은 것)

1. **develop 워크트리의 워킹트리 사본을 읽지 않았다.** 변종 3 은 `origin/develop` **블롭**이다.
   `.claude/worktrees/develop` 의 워킹 파일이 그 블롭과 같은지는 재지 않았다 — 같다는 보장이
   없다(primary 가 바로 그 반례다). 네 번째 변종이 있을 수 있다.
2. **다른 워크트리 사본을 세지 않았다.** 워크트리가 40개 이상이고 각각 `CLAUDE.local.md` 워킹
   사본을 갖는다. 「세 변종」은 **내가 읽은 세 좌표에 대한 진술**이지 전수가 아니다.
3. 미커밋 편집본의 **저자와 시점**을 모른다. `git log` 로는 안 나온다(커밋되지 않았다).
4. 2026-09-07 피해 사건(레인 12곳 배포, lane-3·lane-4 가 멈춤)은 카드 본문의 서술이고
   내가 재현하지 않았다.

## [추가 2026-09-08] 집행된 처분 — 내가 독립 검증한 것

운영자 결재로 리드가 primary 워킹본을 develop 정본으로 교체했다. **리드 보고를 옮기지 않고
직접 쟀다.**

```
$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/reports/t531/
-rw-r--r--  40545  CLAUDE.local.md.primary-uncommitted-backup-20260908

$ wc -l < …/CLAUDE.local.md.primary-uncommitted-backup-20260908
660
$ /usr/bin/grep -c 'push origin develop' …backup-20260908
3
$ shasum -a 256 …backup-20260908
23f8427589b739c705c8b17def0409c187af92ce5f27ee8bb87037a8376c9a7a
```

**백업은 무결하다** — 660줄, 서명 마커 3건(변종 2 의 지문), sha256 이 리드 보고와 일치.
편집본이 이미 덮인 뒤이므로 이 파일이 유일본이고, 되돌림이 가능하다는 것이 여기서 성립한다.

교체 결과:

```
$ /usr/bin/grep -c 'push origin develop' /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md
2
$ wc -l < /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md
734
$ git show origin/develop:CLAUDE.local.md > /tmp/dev2.md
$ diff -q /tmp/dev2.md /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md
(무출력, rc=0 — 바이트 동일)
```

**변종 2 는 소멸했다.** primary 워킹본 = `origin/develop` 블롭.

### `M CLAUDE.local.md` 는 이제 미해결이 아니라 의도된 상태다

교체 뒤에도 `git status` 는 계속 ` M CLAUDE.local.md` 를 낸다. primary 가 `main` 에 체크아웃돼
있고 **main 의 커밋본이 폐기된 제3 모델**(변종 1)이므로, 워킹본이 develop 판인 이상 main 대비로는
영원히 modified 다.

[HARD] **이 표식을 「또 갈라졌다」로 읽고 되돌리면 폐기 모델로 회귀한다.** 레인이 분기하는
트리가 develop 이므로 develop 판이 지배한다. 이 사실은 SPEC 본문에 들어가야 한다 — 다음 사람이
`git status` 한 줄을 보고 판단하지 않도록.

## [추가 2026-09-08] lint 판정의 도구 출처 — 설치본이 뒤처져 있었다

SPEC 작성자가 「lint 를 돌린 바이너리가 트리 HEAD 보다 앞선다는 것을 확인하지 않았다」고
잔여 위험으로 신고했다. **닫았다 — 위험은 실현돼 있었고, 재측정으로 해소했다.**

```
$ moai version | /usr/bin/grep -io '[0-9a-f]\{7,40\}'
e79c010b8

$ git merge-base --is-ancestor e79c010b8 HEAD
rc=0        ← 빌드 커밋이 트리 HEAD 의 조상 = 설치본이 뒤처짐

$ git diff --name-only e79c010b8..HEAD -- internal/spec/
internal/spec/lint_movingref.go
internal/spec/lint_movingref_test.go
internal/spec/era.go  internal/spec/status.go  internal/spec/closer.go  … (20건)
```

`lint_movingref.go` 를 포함한 **lint 규칙이 그 빌드 이후 실제로 바뀌었다.** 뒤처진 빌드는
그 뒤 착지한 규칙을 아예 돌리지 않고도 초록을 낸다 — **초록과 침묵이 같은 출력**이라
구별되지 않는다. 설치본의 `✓ No findings` 만으로는 근거가 서지 않는다.

트리에서 빌드해 다시 쟀다:

```
$ go build -o /tmp/moai-t531 ./cmd/moai          rc=0
$ /tmp/moai-t531 spec lint …/spec.md
✓ No findings — all SPEC documents are valid     EXIT=0
```

**두 좌표가 모두 선다** — 잰 트리 `bce6d7e08`, 판정한 빌드 = 그 트리에서 만든 것.

## Residual-risk

**이 카드가 건드리는 파일은 세션 시작 시 모든 에이전트 컨텍스트에 자동 로드된다**
(`CLAUDE.md` 의 `@CLAUDE.local.md`). primary 의 미커밋 편집본을 커밋하든 되돌리든, 그 순간부터
primary 에서 시작하는 **모든 세션이 믿는 내용이 바뀐다.** 그리고 그 편집본은 커밋되지 않은
사용자 작업이라 `git restore` 류로 지우면 복구 수단이 없다 — 처분은 운영자 판정 사안이다.
