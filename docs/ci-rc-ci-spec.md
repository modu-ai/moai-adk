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
- 마지막 줄은 기계가 읽는 한 줄이다: `local-check: head=<40자 HEAD> result=pass|fail [dirty=1 ]cmd=tools/local_check.sh at=<UTC>`. 시작이나 끝에 커밋하지 않은 변경(무시 파일 제외)이 있으면 확인한 것이 HEAD가 아니므로 `result=fail dirty=1`이다. exit 0은 `result=pass`일 때뿐이다.
- `LOCAL_CHECK_DRY_RUN=1`은 돌릴 단계와 마지막 줄만 출력하고 아무것도 돌리지 않는다. 확인한 것이 없으므로 늘 `result=fail`이다.
- `moai heavy-slot-acquire`가 있으면 test·doctest·clippy를 각각 기계 공용 슬롯 안에서 돌린다(`MOAI_HEAVY_SLOT_MAX` 기본 3, `MOAI_HEAVY_SLOT_MIN_FREE_GB` 기본 50). 없으면(CI 포함) 바로 돌린다.
- diff-check 범위는 `LOCAL_CHECK_BASE`(기본 `origin/v2026.11.0`, 없으면 `HEAD~1`)`...LOCAL_CHECK_DIFF_HEAD`(기본 `HEAD`)다. CI는 PR base·head SHA를 넘긴다.

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
