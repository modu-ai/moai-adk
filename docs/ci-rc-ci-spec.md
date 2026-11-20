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
- `moai heavy-slot-acquire`가 있으면 test·doctest·clippy를 각각 기계 공용 슬롯 안에서 돌린다(`MOAI_HEAVY_SLOT_MAX` 기본 5, `MOAI_HEAVY_SLOT_MIN_FREE_GB` 기본 50). 없으면(CI 포함) 바로 돌린다.
- 파이썬 전체 실행(unittest 단계)은 cargo 슬롯과 다른 폴더의 파이썬 슬롯 안에서 돌린다: `moai heavy-slot-acquire --dir <슬롯 폴더> --max <한도> --min-free-gb <GB> --name <이름>-py -- …`. 폴더는 `MOAI_PYTHON_SLOT_DIR`(기본 `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/python-slots`), 동시에 도는 수는 `MOAI_PYTHON_SLOT_MAX`(기본 2)다. 비어 있는 값은 기본값으로 읽는다. 한도는 `heavy-slot-acquire` 의 규칙대로 기계의 `limits.json`(기본 5) 아래로만 낮출 수 있어서 `MOAI_PYTHON_SLOT_MAX=5` 는 그대로 5 이 된다. 폴더는 절대 경로로 준다(상대 경로는 실행한 폴더 기준이라 후보 모드에서는 슬롯이 공유되지 않는다). `MOAI_PYTHON_SLOT_DIR`, `XDG_CACHE_HOME`, `HOME` 이 모두 비어 있으면 폴더를 정할 수 없으므로 cargo 슬롯 폴더로 되돌아가지 않고, 표준 오류에 한 줄을 알린 뒤 슬롯 없이 바로 돌린다. 38분 걸리던 전체 discover 가 cargo 슬롯 5칸 중 하나를 쥐지 않고, 파이썬끼리는 2개까지만 겹친다. 슬롯 명령이 없으면 바로 돌린다. `tools/self_check.sh --full-python` 도 같은 파이썬 슬롯을 쓰고, 슬롯이 거절하면(exit 75) 그 항목은 `slot-refused (not run)` 으로 not_run 이다.
- in-place 모드의 diff-check 범위는 `LOCAL_CHECK_BASE`(기본 `origin/v2026.11.0`, 없으면 `HEAD~1`)`...LOCAL_CHECK_DIFF_HEAD`(기본 `HEAD`)다. CI는 PR base·head SHA를 넘긴다. candidate 모드의 범위는 고정한 `<base>...<head>`다.

병합 조건은 이 단계에서 바뀌지 않는다. 지금 게이트(head의 GitHub CI 성공 + 감사 PASS)가 그대로이고, `local-check:` 줄을 병합 조건으로 읽는 일은 S5b의 범위다.

## READY 전에 돌리는 전체 트리 점검 `tools/self_check.sh --pre-ready`

레인은 READY 를 알리기 전에 이 목록을 위에서 아래로 돌린다(빠른 순서). 한 줄에 한 항목, `self-check: <항목> exit=<코드> …` 로 찍히고 마지막 줄이 요약이다. 이 목록은 한 묶음이 몇 시간 READY 에 못 올라가게 한 실패 6가지(fmt, SPEC Result 없음, source-manifest 항목 없음, repo_scans 표 항목 없음, 워크플로 파일 속 금지 문자열, windows 게이트)를 PR 전에 잡으려는 것이고, 병합 증거가 아니다. 병합 조건은 위와 같다.

| 항목 | 명령 | 슬롯 |
| --- | --- | --- |
| `diff-check` | `git diff --check <merge base>` | 없음 |
| `fmt` | `cargo fmt --check` | 없음 |
| `fmt-ops` | `rustfmt --check --edition 2021 src/ops/*.rs` | 없음 |
| `manifest` | `python3 -I tools/source_manifest.py check` | 없음 |
| `spec-check` | 바뀐 SPEC 마다 빌드한 `moai-cli spec check <SPEC-ID> --closure --project .` (초안 SPEC 은 승인이 사용자 몫이라 항상 나오는 `closure_status` 하나만 통과로 읽고, `## Result` 없음(`closure_result`)이나 읽히지 않는 요구사항은 실패) | 없음 |
| `py-source-manifest` | `python3 -m unittest tests.source_manifest_test` | 파이썬 슬롯 |
| `py-repo-scans` | `python3 -m unittest tests.repo_scans_test` | 파이썬 슬롯 |
| `skill-package` | `cargo test --offline --locked --test product_cli skill_package` | heavy 슬롯 |

- `spec-check` 와 `py-repo-scans` 는 빌드한 `moai-cli` 를 읽는다(`MOAI_CLI_BIN`, 없으면 `${CARGO_TARGET_DIR:-target}/debug/moai-cli`). 없거나 `src/`·`build.rs`·`Cargo.toml`·`Cargo.lock` 의 어느 파일보다 오래됐으면 `not_run (no built moai-cli: …)` 또는 `not_run (stale moai-cli: …)` 이고 통과로 세지 않는다. 먼저 `cargo build --offline --locked --bin moai-cli` 를 돌린다(heavy 슬롯 안에서). 빌드가 없으면 `tests.repo_scans_test` 는 25개 중 19개가 `moai-cli not built` 로 실패한다(2026-11-19 직접 실행).
- 바뀐 SPEC 이 없으면 `spec-check` 는 `skip (no changed SPEC)` 이다. merge base 가 없으면 `diff-check` 와 `spec-check` 는 `not_run (no base: …)` 이다(`SELF_CHECK_BASE`).
- windows 게이트는 항목이 아니다. 플랫폼 코드(unix 전용 호출, `std::os::unix`)를 건드린 변경만 직접 `cargo check --offline --locked --tests --target x86_64-pc-windows-msvc` 로 형을 확인하고(맥에서는 `libsqlite3-sys` 빌드가 C 컴파일러를 요구해 흉내 낸 `CC_`/`AR_` 환경이 있어야 하고, 링크와 실행은 하지 않는다), 확정은 병합 뒤 `windows-build` run 이다.
- `--pre-ready` 는 `--full-python` 과 함께 줄 수 없다(exit 2). 일반 실행(`tools/self_check.sh`)의 항목과 순서는 바뀌지 않는다.

## SYNC가 CI 결과를 읽는 방법

후보 SHA마다 아래 명령으로 조회해 integration receipt의 `ci_checks[]`에 넣는다. 결과가 없으면 `not_run`이다.

```sh
gh api repos/modu-ai/moai-adk/commits/<candidate_sha>/check-runs \
  --jq '.check_runs[] | {name,status,conclusion,head_sha}'
```

## 범위 밖

- `rc-ci.yml` 수정, 필수 검사 추가, ruleset 변경(08-D #2015, 저장소 설정은 사용자 권한)
- merge queue(`merge_group`): RC1은 직렬 통합이라 도입하지 않는다(#2012 본문의 비채택 기록).
