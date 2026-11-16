# rc-ci 명세 (08-A #2012)

`v2026.11.0` 대상 PR의 CI가 SYNC 절차(`skills/moai/workflows/sync.md`)에 넘겨야 하는 것을 정한다. workflow 파일 `.github/workflows/rc-ci.yml`의 소유 Issue는 08-D #2015이고(2026-11-08 사용자 확정, `docs/decisions-2026-11-08.md`), 첫 판은 00-E #2020이 만들었다. 이 문서는 요구만 적는다. 파일을 바꿀 일이 생기면 08-D에 요청한다.

아래 "현재 상태"는 2026-11-08에 `rc-ci.yml`을 읽고 PR #2058의 실행(run 42478275200)을 조회해 확인한 내용이다.

## 요구와 현재 상태

| ID | 요구 | 현재 상태 |
| --- | --- | --- |
| CI-1 | `pull_request`의 대상 브랜치가 `v2026.11.0`일 때 실행한다. | 충족. `on.pull_request.branches: [v2026.11.0]`. PR #2058(`issue/2054-…` → `v2026.11.0`)에서 `rc-ci`가 실행되어 pass로 끝났다. 기본 브랜치에 이 파일이 없어도 PR 쪽 ref의 파일로 실행된다는 Issue의 미확인 항목은 이 실행으로 관측했다. |
| CI-2 | 권한은 읽기만 쓴다(`contents: read`). checkout은 자격 증명을 남기지 않는다. | 충족. workflow·job 모두 `contents: read`, `persist-credentials: false`. |
| CI-3 | 로컬 검사 묶음과 같은 명령을 돌린다: `cargo test --offline --locked`, `cargo clippy --offline --locked --all-targets -- -D warnings`, `cargo fmt --check`, Python unittest, `git diff --check`. | 충족. 의존성은 `cargo fetch --locked`로 한 번 받은 뒤 오프라인으로 실행한다. `git diff --check`는 PR base와 head 사이를 본다. |
| CI-4 | 한 단계가 실패해도 나머지를 끝까지 돌리고, 단계마다 명령과 exit code를 남긴다. | 충족. step summary 표에 `step · command · exit`를 적고, 하나라도 실패하면 job이 실패한다. |
| CI-5 | 후보 SHA(`github.event.pull_request.head.sha`)와 실제로 빌드한 merge ref SHA(`GITHUB_SHA`)를 함께 기록한다. | 충족. summary 첫 줄에 `tree: <head> (head) · merge ref <GITHUB_SHA>`를 적는다. |
| CI-6 | 같은 PR의 이전 실행은 취소해 마지막 head의 결과만 남긴다. | 충족. `concurrency` 그룹을 PR 번호로 묶고 `cancel-in-progress: true`. |
| CI-7 | 외부 action은 커밋 SHA로 고정한다. | 충족. `actions/checkout`, `actions/setup-python`을 40자 SHA로 고정했다. |
| CI-8 | merge conflict로 실행되지 않은 경우를 PASS로 읽지 않는다. | CI 쪽에서 할 일은 없다. GitHub은 충돌이 있는 PR에서 `pull_request` workflow를 실행하지 않는다. SYNC 절차 6단계가 이 경우를 `not_run`으로 적는다. |
| CI-9 | `rc-ci`를 필수 검사로 둘지 정한다. | 미결. 현재 ruleset `RC1 integration gate (v2026.11.0)`(id 26791615)의 필수 status는 `moai/audit` 하나다. `rc-ci`를 추가하는 일은 08-D #2015의 범위다. |

## 로컬 검사 정본 `tools/local_check.sh` (#2906 S5a)

`ci.yml`의 `check` job이 돌리는 검사 단계(fetch, test(cargo-nextest), doctest, clippy, fmt, unittest, diff-check)는 `tools/local_check.sh` 한 파일에 같은 순서·인자로 있고, CI의 "Check set" 단계는 이 파일을 부른다. 그래서 로컬 검사와 CI 검사가 갈라지지 않는다. `tests/local_check_test.py`가 단계 목록을 고정한다(AC20). 체크아웃·도구 설치·캐시·docs-only 범위 판정 같은 CI 준비 단계는 `ci.yml`에 남는다.

- 저장소 루트가 아닌 곳에서 불러도 스스로 루트로 옮겨 돈다. 기본은 첫 실패에서 멈춘다. CI는 `LOCAL_CHECK_KEEP_GOING=1`로 모든 단계를 돌리고 단계마다 exit을 summary 표에 남긴다(CI-4 그대로).
- 모드가 둘이다(#3066 P2). 기본인 `candidate`(후보) 모드는 head를 base 가지의 최신 tip과 합친 커밋을 검사한다: `LOCAL_CHECK_BASE_BRANCH`(기본 `v2026.11.0`, `[A-Za-z0-9._/-]`만, 앞 `-`·`..` 거절)를 `origin`에서 `--quiet --no-write-fetch-head --no-tags +refs/heads/<b>:refs/remotes/origin/<b>`로 받아 SHA로 고정하고, `git merge-tree --write-tree`와 `git commit-tree`(부모 base, head)로 병합 커밋을 만들고, `mktemp -d` 아래에 `git worktree add --detach`로 꺼내 같은 단계 목록을 거기서 돌린 뒤 `git worktree remove`로만 지운다(`rm -rf` 없음). 작업 폴더·index·가지·stash는 건드리지 않고 바뀌는 ref는 `refs/remotes/origin/<b>` 하나뿐이다. 커밋한 HEAD만 검사하므로 커밋하지 않은 변경이 있으면 아무것도 돌리지 않고 fail이다. fetch가 실패하거나 가지 이름이 거절되면 낡은 base로 통과 줄을 내지 않고 `result=fail base=none tree=none`이다. 충돌(merge-tree 종료 1)이나 병합 오류는 `base=<40hex> tree=none`인 fail이고 후보 폴더를 만들지 않는다. 이 모드에서 `LOCAL_CHECK_BASE`·`LOCAL_CHECK_DIFF_HEAD`는 무시하고(stderr로 알린다) diff-check는 `<base>...<head>`를 본다. 후보 폴더는 매번 새로라 cargo가 거기서 처음부터 빌드한다(`CARGO_TARGET_DIR`로 바꿀 수 있다). 강제 종료로 남은 `local-check-cand.*` 폴더는 다음 실행이 알리기만 하고 지우지 않는다. `LOCAL_CHECK_MODE=in-place`(CI 전용)는 이 작업 폴더에서 그대로 돌고 아래 짧은 꼴을 찍는다. CI의 checkout이 이미 병합 ref이므로 `ci.yml`의 "Check set"이 이 값을 넘긴다. 그 밖의 값은 줄 없이 종료 2다.
- 마지막 줄은 기계가 읽는 한 줄이다. in-place 모드의 짧은 꼴: `local-check: head=<40자 HEAD> result=pass|fail [dirty=1 ]cmd=tools/local_check.sh at=<UTC>`. 시작이나 끝에 커밋하지 않은 변경(무시 파일 제외)이 있으면 확인한 것이 HEAD가 아니므로 `result=fail dirty=1`이다. candidate 모드의 긴 꼴(아래)은 `dirty=1` 필드를 찍지 않고(게이트가 읽지 못한다) 같은 경우를 `result=fail`로 닫는다. exit 0은 `result=pass`일 때뿐이다.
- 이 줄에는 두 가지 꼴이 있다(#3066 P1·P2). 짧은 꼴은 위의 줄이고 in-place 모드가 찍는다. 긴 꼴은 candidate 모드(기본)가 찍으며 `base`는 고정한 base tip, `tree`는 그 병합의 트리, `script`는 이 스크립트의 `git hash-object`, `rustc`는 후보 안의 `rustc --version`의 버전 낱말(읽지 못하면 `unknown`과 fail), `run`은 `head base tree at pid`의 sha1 앞 12자리다. 모양은 `local-check: head=<40hex> result=<pass|fail> base=<40hex|none> tree=<40hex|none> script=<40hex> rustc=<토큰> run=<12hex> cmd=tools/local_check.sh at=<UTC>`이며(필드 순서·한 칸 띄움·0열·뒤에 글자 없음, hex는 모두 소문자, `rustc`는 `[A-Za-z0-9._+-]` 1~32자), `pass` 줄에는 실제 base와 tree가 있고 `base=none`·`tree=none`은 `result=fail`일 때뿐이며 `base=none`이면 `tree=none`이다. 게이트(`gate::parse_local_check`)는 두 꼴을 모두 읽는다. 어느 꼴이든 `dirty=1` 필드가 있으면 게이트는 줄이 없는 것으로 본다.
- `LOCAL_CHECK_DRY_RUN=1`은 돌릴 단계와 마지막 줄만 출력하고 아무것도 돌리지 않는다(candidate 모드에서는 fetch도 하지 않아 diff-check가 `<base>...<head>`로 찍힌다). 확인한 것이 없으므로 늘 `result=fail`(긴 꼴은 `base=none tree=none`)이다.
- `moai heavy-slot-acquire`가 있으면 test·doctest·clippy를 각각 기계 공용 슬롯 안에서 돌린다(`MOAI_HEAVY_SLOT_MAX` 기본 3, `MOAI_HEAVY_SLOT_MIN_FREE_GB` 기본 50). 없으면(CI 포함) 바로 돌린다.
- in-place 모드의 diff-check 범위는 `LOCAL_CHECK_BASE`(기본 `origin/v2026.11.0`, 없으면 `HEAD~1`)`...LOCAL_CHECK_DIFF_HEAD`(기본 `HEAD`)다. CI는 PR base·head SHA를 넘긴다. candidate 모드의 범위는 고정한 `<base>...<head>`다.

병합 조건은 이 단계에서 바뀌지 않는다. 지금 게이트(head의 GitHub CI 성공 + 감사 PASS)가 그대로이고, `local-check:` 줄을 병합 조건으로 읽는 일은 S5b의 범위다.

## SYNC가 CI 결과를 읽는 방법

후보 SHA마다 아래 명령으로 조회해 integration receipt의 `ci_checks[]`에 넣는다. 결과가 없으면 `not_run`이다.

```sh
gh api repos/modu-ai/moai-adk/commits/<candidate_sha>/check-runs \
  --jq '.check_runs[] | {name,status,conclusion,head_sha}'
```

## 범위 밖

- `rc-ci.yml` 수정, 필수 검사 추가, ruleset 변경(08-D #2015, 저장소 설정은 사용자 권한)
- merge queue(`merge_group`): RC1은 직렬 통합이라 도입하지 않는다(#2012 본문의 비채택 기록).
