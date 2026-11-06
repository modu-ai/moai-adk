auditor-model: claude-opus-5-5

# t2087 sync 감사 — 3회차(N1 수리 델타) 판정

verdict: FAIL
audited_sha: cbf7e3808fd34df391499f04eb35ad561174e491
required_backend_fail: codex

- 카드: t2087 (Class B, SPEC 없음)
- 점수: 코드 판단 93/100. 판정은 FAIL — 코드 결함 때문이 아니라 필수 Codex 게이트가 채워지지 않아서다.
- 감사 범위: `a6663e9417..cbf7e3808f` (`117238b08b` 2회차 판정 기록 파일 추가, `cbf7e3808f` N1 시험 수리). 파서는 바뀌지 않았다.
- Codex 영수증: 없음. 운영자가 허용한 3회차 호출 1회(결정 `d-20261106T071854Z-643f`)를 보냈으나 서버가 실행 전에 거부했다(`stale_server`). 재시도하지 않았다.

## 요약

N1은 고쳐졌다. 새 시험은 파일이나 디렉터리를 만들지 않고, 손으로 쓴 `git log -M -p` 스트림을 `parseOwnershipBatch`에 넣는다. `"`가 든 경로는 스트림 문자열 안에만 있다. Windows 제외도 없다.

F3 회귀 방어력도 그대로다. 스크래치 사본에서 파서를 F3 이전 조건(`CutPrefix(line, "diff --git a/")`)으로 되돌리자 새 시험이 실패했고, 현재 파서에서는 통과했다. 같은 스크래치 사본에서 이 기기의 실제 git이 낸 출력(따옴표 든 디렉터리가 있는 스크래치 저장소)을 넣는 감사용 시험도 돌렸는데, 결과가 합성 스트림과 똑같았다(F3 이전 파서에서 실패, 현재 파서에서 통과).

빌드, vet(네이티브와 `GOOS=windows`), gofmt, 수리된 시험은 모두 통과했다.

판정을 FAIL로 둔 까닭은 하나다. 감사 계획이 codex를 `required`로 표시하는데, 허용된 단 한 번의 Codex 호출이 서버 쪽 `stale_server` 거부로 실행되지 않았다. 필수 게이트가 답하지 않으면 PASS를 낼 수 없다. 코드 쪽 차단 결함은 찾지 못했다. 운영자 결정 본문에 따라 이 뒤로는 호출하지 않으며, 다음 조치는 리더가 운영자에게 물을 일이다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 내용 | 필요한 수정 |
|---|---|---|---|---|---|
| G1 | 게이트 | blocking(게이트), 코드 결함 아님 | MCP 서버 `moai` rc.42 `646955327e` | 필수 codex 게이트 미충족: `codex_audit` 호출이 `stale_server`로 거부되어 영수증이 없다. | 리더가 운영자에게 결정을 받는다. 서버를 재연결(호스트 세션 재시작)한 뒤 Codex 1회를 새로 허용할지, 다른 처분을 할지. 감사자는 더 호출하지 않는다. |
| O2 | Info | optional, in-diff, 측정 | `internal/spec/lint_gitbatch_test.go:510` | 합성 스트림의 커밋 머리 줄은 `%b` 끝의 세 번째 US를 같은 줄에 둔다. 실제 git은 본문 끝 줄바꿈 때문에 세 번째 US를 다음 줄에 단독으로 낸다(아래 증거 2). 그래서 이 시험은 머리 줄 한 줄 경로(`hdrSeps >= 3` 즉시 종료)를 타고, 실제 출력은 여러 줄 경로를 탄다. F3 판단과는 무관하고, 여러 줄 경로는 실제 git을 쓰는 다른 배치 시험이 탄다. | 차단 아님. 원하면 머리 줄을 `"...manager-docs", us,` 두 줄로 나눠 실제 모양과 맞춘다. |
| N1 | (해소) | 2회차 차단 결함 | `lint_gitbatch_test.go` | 확인 결과 해소(아래 증거 1~3). | 없음 |

이월 부채(이번 판정 대상 아님, 그대로 유지):
- F4 (Low): `batchedOwnershipTransition` root별 pathspec — `_archive` 경로 이름 변경에서 Info 출력이 다르다.
- F5 (Info): 집합 단위 이력 단순화와 `--follow`의 차이(잔여 위험).
- F6 (Low): 배치 실패 시 per-SPEC 경로로 조용히 대체된다(결과는 정확, 느림).
- O1 (Low, optional): 같은 패키지의 다른 시험 파일 7개가 `gc.auto=0`/`maintenance.auto=false` 없이 git 저장소를 만든다.

베이스라인 위반 `OwnershipTransitionInvalid` recorded 0 -> current 1은 이번 델타가 건드리지 않은 기존 결함이고 t2105가 소유한다(파서·베이스라인 JSON·워크플로 모두 델타 밖, 증거 0).

반복 수리 중단 규칙: 이번 수리(`cbf7e3808f`)에서 생긴 차단 결함은 찾지 못했다. 연속 2회 한도에 닿지 않았다.

## 차원별 판단

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 새 시험이 F3 이전 파서를 잡아낸다(돌연변이 실패/현재 통과). 실제 git 출력으로도 같은 결과. |
| Security (25%) | 95 | PASS | 시험 파일만 바뀌었다. 새 입력 경로, 비밀값, 외부 호출이 없다. 파서는 바이트 단위로 같다. |
| Craft (20%) | 88 | PASS | 파일·디렉터리 생성 없음, Windows 제외 없음, build/vet/gofmt/시험 통과. O2(머리 줄 모양 차이)와 원래 시험의 오라클 대조가 빠진 점은 감점. |
| Consistency (15%) | 92 | PASS | 같은 파일의 시험 이름·헬퍼 관례를 따른다. 스트림 주석이 왜 합성 입력인지 밝힌다. |

가중합: 0.4×95 + 0.25×95 + 0.2×88 + 0.15×92 = 93.15 → 93. 판정은 필수 codex 게이트 미충족(G1) 때문에 FAIL이다.

## 증거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### 0. 트리와 범위
- 주장: 맞는 트리에서, 지시된 델타만 감사했다. 파서는 바뀌지 않았다.
- 증거: `git rev-parse --show-toplevel` → `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`; `git branch --show-current` → `WT-spec-lint-baseline-2`; `git rev-parse HEAD` → `cbf7e3808fd34df391499f04eb35ad561174e491`; `git status --short` → 빈 출력. Codex 호출 직전 다시 확인: 같은 toplevel, 같은 HEAD, 빈 status.
- `git diff --stat a6663e9417 cbf7e3808f` → `.moai/reports/t2087/sync-audit-verdict-iter2.md | 121 +++...` / `internal/spec/lint_gitbatch_test.go | 70 +++++++++-----` / `2 files changed, 168 insertions(+), 23 deletions(-)` (exit 0).
- `git diff --stat a6663e9417 cbf7e3808f -- internal/spec/lint_ownership_batch.go` → 빈 출력 (exit 0).
- 기준 귀속: 이 실행, 이 트리, HEAD `cbf7e3808f`.

### 1. N1 해소 — 따옴표 경로는 문자열 안에만
- 증거: `grep -n 'SPEC-Z' internal/spec/lint_gitbatch_test.go` →
  - `520:		` + "`diff --git \"a/.moai/specs/SPEC-Z\\\"Q-001/spec.md\" \"b/.moai/specs/SPEC-Z\\\"Q-001/spec.md\"`,"
  - `522:` `--- "a/.moai/specs/SPEC-Z\"Q-001/spec.md"`, `523:` `+++ "b/.moai/specs/SPEC-Z\"Q-001/spec.md"` — 세 줄 모두 `strings.Join` 인자 안의 스트림 텍스트다.
- 시험 본문(495~553행)에서 `os.|MkdirAll|WriteFile|Create|TempDir|newGBRepo|Skip|GOOS|runtime`을 찾은 결과: 일치 없음(grep exit 1). 파일·디렉터리 생성도, Windows 제외도 없다. 시험은 `parseOwnershipBatch(strings.NewReader(stream))`와 `b.resolve(...)`만 부른다.
- 수리된 시험 실행: `go test -count=1 -run 'QuotedPathHeader' -v ./internal/spec/` → `--- PASS: TestLintGitBatch_QuotedPathHeaderClosesSegment (0.00s)` / `ok  	github.com/modu-ai/moai-adk/internal/spec	0.682s` (exit 0).

### 2. 합성 스트림과 실제 git 출력 비교
- 방법: 저장소 밖 스크래치 저장소(`/private/tmp/claude-501/t2087-iter3-audit/repo`)에 `SPEC-A-001`과 `SPEC-Z"Q-001` 디렉터리를 만들고, 같은 두 커밋(생성 / A 본문 수정 + Z 상태 변경)을 만든 뒤 `readOwnershipBatch`와 같은 인자로 `git -c core.quotepath=false log -M --no-color --no-ext-diff --no-relative --src-prefix=a/ --dst-prefix=b/ '--format=%x1e%H%x1f%s%x1f%b%x1f' -p -- ':(top,glob).moai/specs/*/spec.md'`를 실행했다(`log-exit=0`, `git version 2.54.0 (Apple Git-157)`). 셸 스크립트 파일로 실행한 까닭은 Gaps에 적었다.
- 실제 출력(`cat -v`, 발췌):
  ```
  ^^1b566f2b8f...^_fix: M1^_Authored-By-Agent: manager-docs
  ^_

  diff --git a/.moai/specs/SPEC-A-001/spec.md b/.moai/specs/SPEC-A-001/spec.md
  ...
  diff --git "a/.moai/specs/SPEC-Z\"Q-001/spec.md" "b/.moai/specs/SPEC-Z\"Q-001/spec.md"
  index f70786f..0603dcd 100644
  --- "a/.moai/specs/SPEC-Z\"Q-001/spec.md"
  +++ "b/.moai/specs/SPEC-Z\"Q-001/spec.md"
  @@ -1,3 +1,3 @@
   id: SPEC-Z
  -status: draft
  +status: in-progress
  ```
- 판단: 따옴표 헤더 줄, `---`/`+++` 줄, `core.quotepath=false`에서도 `"`가 인용된다는 점, 커밋 순서(최신 먼저)는 합성 스트림과 같은 모양이다. 다른 점은 (a) 해시·index 값(파서는 해시를 그대로 옮길 뿐), (b) 세 번째 US 위치(O2)다. 파서가 기대하는 레이아웃(`ownershipBatchFormat = "--format=%x1e%H%x1f%s%x1f%b%x1f"`, RS로 커밋 시작, US 3개로 머리 끝)에 둘 다 맞는다.

### 3. F3 방어력 — 돌연변이
- 방법: `git archive -o .../src.tar cbf7e3808f...`(exit 0)로 감사 커밋을 스크래치에 풀었다(`tar-exit=0`; `cmp` 파서·시험 파일 모두 `0`). 저장소는 건드리지 않았다. 감사용 시험 `zz_audit_realstream_test.go`(스크래치에만 있음)는 증거 2의 실제 출력 `real.txt`를 파서에 넣고 같은 단언을 한다.
- 현재 파서: `go test -count=1 -run 'QuotedPathHeader|AuditRealStream' -v ./internal/spec/` → 두 시험 `PASS`, `ok ... 0.583s` (exit 0).
- 돌연변이(스크래치 사본 283행을 `if rest0, ok := strings.CutPrefix(line, "diff --git a/"); ok { fileHdr := "a/" + rest0` — F3 이전 조건과 같은 동작): 같은 명령 exit 1:
  - `lint_gitbatch_test.go:551: SPEC-A-001 must resolve to its creation record, got &{PreviousStatus:draft CurrentStatus:in-progress CommitSubject:fix: M1 CommitSHA:bbbb AuthoredByAgent:manager-docs}`
  - `zz_audit_realstream_test.go:22: SPEC-A-001 must resolve to its creation record, got &{PreviousStatus:draft CurrentStatus:in-progress CommitSubject:fix: M1 CommitSHA:1b566f2b8f2836bd2c6db60d8536e92d16b49b68 AuthoredByAgent:manager-docs}`
- 기준 대조: `git show 140c1ca406:internal/spec/lint_ownership_batch.go` 280행 `if rest, ok := strings.CutPrefix(line, "diff --git a/"); ok {` — 돌연변이가 이 조건을 재현한다.
- 판단: 새 시험은 F3 회귀를 잡고, 합성 입력과 실제 git 입력이 같은 방식으로 실패한다. 손으로 쓴 스트림 덕에만 통과하는 시험이 아니다.

### 4. 기존 검사 (이 트리, HEAD `cbf7e3808f`)
- `go build ./...` → exit 0, 출력 없음.
- `go vet ./internal/spec/...` → exit 0.
- `GOOS=windows go vet ./internal/spec/` → exit 0. Windows 컴파일만 확인했고 Windows 실행은 관찰하지 않았다.
- `gofmt -l internal/spec` → 출력 없음, exit 0.
- `go test -count=1 -run 'QuotedPathHeader' ./internal/spec/` → exit 0 (증거 1).
- `go test -count=1 -run 'TestLintGitBatch|Ownership|ImpliedStatus|StatusGitConsistency|ConstantGitLog|Unreachable' ./internal/spec/` → `ok  	github.com/modu-ai/moai-adk/internal/spec	558.041s` (exit 0, 부하 140~190에서 실행)
- 판단 빌드: 시험과 vet은 이 트리 소스로 컴파일한 것이다(§2.2 충족). `moai verify audit-plan`·`moai decision read`는 설치된 바이너리로 실행한 절차 확인용이다.

### 5. Windows 금지 문자 경로 — 패키지 전체
- 증거: `internal/spec/*_test.go`에서 `MkdirAll|WriteFile|os.Create|Mkdir(|filepath.Join` 인자 리터럴에 `* : < > ? |`가 든 경우를 찾았다. 나온 것은 모두 `filepath.Glob` 패턴(`*/acceptance.md`, `SPEC-*/spec.md`, `*.pack`)이라 파일을 만들지 않는다. 백틱 안 `"`가 든 줄 중 경로는 증거 1의 스트림 세 줄뿐이다. `lint_gitbatch_test.go`의 `r.spec(...)` 이름은 모두 `SPEC-[A-Z0-9-]+` 모양이다(56~202행, 385행 `SPEC-CNT-%03d`).
- 판단: 이 델타가 새로 들인 것도, 기존 것도 찾지 못했다. grep 기반이라 런타임에 조립되는 이름은 다 보지 못했다(Gaps).

## Codex 감사 (3회차, 수리 델타 한정)
- 승인: 결정 `d-20261106T071854Z-643f`(scope `card:t2087`, 운영자 원문 "수리 범위만 3회차 허용 (권장)"). 앞선 두 번은 `rcpt-265f061edb178a80d58392da`, `rcpt-4bf94c4b99b843a5d0c8b2e4`.
- 감사 계획: `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2087` (exit 0) → `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude·glm `"gate": "off"`, `"enforced_required": ["codex"]`. 계획에는 target 항목이 없어 delta를 쓸 수 있는 유일한 대상 `baseBranch`를 골랐다.
- 넘긴 인자(정확히): `project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`, `target=baseBranch`, `delta_base=a6663e9417718fcd3481c91aa573aefec4617d4d`, `mode=native`. `model`·`focus`는 넘기지 않았다.
- 결과(그대로 인용): `{"_server":{"commit":"646955327e","version":"v3.2.0-rc.42"},"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves","stale_server":true}`
- 영수증: 없음. 재시도하지 않았고 다른 감사 도구도 부르지 않았다. 이 거부가 운영자가 허용한 1회를 쓴 것으로 칠지는 리더·운영자가 정할 일이다(호출은 보냈지만 리뷰는 실행되지 않았다).
- 해결되지 않은 Codex 차단 지적: 없음(리뷰가 실행되지 않았으므로 "지적 없음"이 아니라 "관찰 없음"이다).

## 반복 이력
- 1회차(`140c1ca406`): FAIL 77. F1(Windows PATH 심), F2(`diff.relative`) 차단, F3 수리 권고, F4~F6 부채.
- 2회차(`a6663e9417`): FAIL 84. F1·F2·F3·플레이크 수리 확인. 수리에서 생긴 N1(따옴표 디렉터리, Windows 금지 문자).
- 수리: `cbf7e3808f` — 시험을 합성 스트림 입력으로 바꿈(+47/-23, 시험 파일만).
- 3회차(`cbf7e3808f`): N1 해소와 F3 방어력 확인, 코드 점수 93. 필수 Codex 게이트가 `stale_server` 거부로 채워지지 않아 FAIL.

## Gaps (관찰하지 않은 것)
- 도구 거부 1: 스크래치 저장소를 만드는 복합 셸 명령(`git init`/`commit`/`log`가 한 줄에 든 것)을 worktree 격리 가드가 거부했다. 같은 단계를 `/private/tmp/claude-501/t2087-iter3-audit/mkrepo.sh` 스크립트 파일(`GIT_DIR`/`GIT_WORK_TREE`를 스크래치 저장소로 고정)로 실행했다. 결과는 실제 실행이고 추론으로 대체하지 않았다.
- 도구 거부 2: `rm -rf`가 든 스크래치 준비 명령이 권한 거부됐다. `rm` 없이 새 디렉터리를 만들어 다시 실행했다.
- 도구 거부 3: `codex_audit`가 서버 쪽 `stale_server`로 거부됐다(G1). Codex 리뷰 결과는 관찰하지 않았다.
- 감사 계획이 `cross_model_active: true`라서 에이전트 정의상 `audit_multi` 경로가 기본이지만, 운영자 결정 범위(Codex 1회)와 지시에 따라 `audit_multi`는 부르지 않았다.
- Windows 실행: 관찰하지 않았다. 컴파일(`GOOS=windows go vet`)과 정적 판단뿐이다.
- 실행하지 않은 검사: `./internal/spec/...` 전체(지시로 제외), golangci-lint, 커버리지, CI.
- 원래 시험이 하던 배치 경로 대 per-SPEC 오라클 대조는 새 시험에서 빠졌다. 따옴표 경로가 있는 실제 저장소에서 `lookupOwnershipTransitionFromGit` 전체 경로를 도는 시험은 이제 없다(이번 감사의 실제 출력 시험은 스크래치에만 있다).
- 금지 문자 검색은 grep 기반이라 런타임에 조립되는 이름은 모두 보지 못했다.

## Residual-risk (관찰했는데도 남는 위험)
- 합성 스트림은 `ownershipBatchFormat`을 손으로 흉내 낸다. 형식 문자열이 바뀌면 시험과 파서를 함께 고쳐야 한다(레인도 같은 위험을 적었다).
- 감사 시스템 머신 부하가 140~190이어서 시간 민감한 시험은 CI에서 다르게 보일 수 있다.
- 스크래치 증거(`/private/tmp/claude-501/t2087-iter3-audit/` 아래 `real.txt`, `mut-baseline.txt`, `mut-pref3.txt`, `t-*.txt`, `src/`)는 이 기기에만 있다. 판정에 쓴 줄은 위에 옮겼다.

## 운영 메모 (미검증)
- `measured`: 서버는 rc.42 `646955327e`로 떠 있고 설치본과 다르다고 스스로 보고했다(위 Codex 결과 원문). 재연결 뒤 호출이 성공할지는 측정하지 않았다 — 확인 명령: `moai doctor`.
- `inferred`: 코드 쪽 판단만으로는 PASS 조건(N1 해소, F3 방어 유지, 기존 검사 통과)을 채웠다. 남은 것은 게이트 하나다. 근거 규칙: 필수 백엔드가 답하지 않으면 PASS를 낼 수 없다(sync-auditor 정의 § Read the plan before scoring).
