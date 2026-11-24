---
id: SPEC-20261120-release-pr-three-os
title: "릴리스 PR 은 머지 전에 Linux·macOS·Windows 에서 검사 세트 전부를 돌려요"
status: draft
revision: 2
created: 2026-11-20
parent: null
---

# 릴리스 PR 은 머지 전에 Linux·macOS·Windows 에서 검사 세트 전부를 돌려요

## Intent

사용자 결정(2026-11-20, 선택 "세 OS 모두, 테스트 전부"; 2026-11-24 카드 정리에서 RC1 에 남기기로 다시 선택):
원격 `v2026.11.0` 으로 올리는 릴리스 PR 이 릴리스 단위라, 머지 전에 Linux x86_64·macOS arm64·Windows x86_64 세 OS 에서
검사 세트 전부를 돌리고, Windows 에서 실패하면 머지가 막혀요.

판 2 는 2026-11-20 의 판 1 을 통합 가지의 지금 모양에 맞춘 거예요. 판 1 은 옛 가지에만 있고 통합 가지에는
들어오지 않았어요. 판 1 이 쓴 뒤에 세 가지가 달라졌어요.

- 릴리스 가지 이름은 `release/v…` 가 아니라 `release/<40자 hex>` 와 `release/rc/<40자 hex>` 예요
  (`.github/workflows/release.yml` 의 plan 단계).
- `.github/workflows/candidate.yml` 이 `.github/workflows/candidate-target.yml` 을 불러서 Windows 에서도 검사 세트를
  돌려요. Windows 용 처리(고정 버전 ripgrep 설치, `.exe` 멤버, `cygpath -u`)가 그 파일에 이미 있어요.
- 검증 테스트는 Rust 만이에요. 판 1 의 verify 명령(Python unittest)은 그대로 쓸 수 없어요.

읽어서 확인한 사실(코드, 통합 가지 `24dceb4e2b0`, 2026-11-24):

- `.github/workflows/ci.yml` 의 `targets` 잡 decide 단계는 `full=`(linux-x86_64·macos-arm64, 다리 둘씩) 아니면
  `linux=` 를 `matrix` 출력으로 내보내요. `pull_request` 일 때만 `tools/ci_scope.py` 분류기가 답하고 나머지는
  `full` 이에요. `windows-build` 는 `if: ${{ github.event_name != 'pull_request' }}` 이고 테스트를 실행하지 않아요.
  워크플로 어디에도 Windows 테스트 다리가 없어요.
- `ci` 집계 잡은 `needs: [check, ci-verdict]` 이고 단계는 `[ "${{ needs.check.result }}" = "success" ]` 만 봐요.
  `ci.yml` 과 `candidate-target.yml` 의 머리 주석에 따르면 필수 검사는 `ci` 와 `moai/audit` 이고 후보 검증은 필수가
  아니에요(ruleset 은 오늘 다시 읽지 않았어요). 그래서 오늘은 Windows 실패가 릴리스 PR 의 머지를 막지 못해요.
- `candidate.yml` 의 `windows-x86_64` 잡이 `./.github/workflows/candidate-target.yml` 을 여섯 입력(`sha`,
  `target: windows-x86_64`, `runner: windows-2025`, `timeout: 60`, `nextest_asset: x86_64-pc-windows-msvc`,
  `nextest_sha256: e51bdb4b5883df48f335381b287378edb244318fac65ae0adaf6b9e69298246b`)으로 불러요.
- `release.yml` 의 plan 단계는 `^release/rc/([0-9a-f]{40})$` 와 `^release/([0-9a-f]{40})$` 두 모양만 받아요.
- `tests/github_assets_test.py` 가 이 변경이 바꾸는 세 곳을 글자 그대로 고정해요: decide 스크립트의
  `pull_request` 가드 한 줄, `windows-build` 의 `if`, 집계 잡의 `needs` 목록.

## Non-goals

- 일반 PR 의 매트릭스, push·`workflow_dispatch` 의 매트릭스, 분류기와 그 실패 시 전체 매트릭스 규칙. 일반 PR 은 오늘의
  규칙 그대로예요(REQ-2).
- 필수 검사 이름(`ci`, `moai/audit`), ruleset, `release.yml`, `audit-gate.yml`, `check-guard.yml`, `candidate.yml` 의
  잡 구성.
- Windows 에서만 실패하는 테스트와 Windows 대상 clippy 경고 고치기. 첫 릴리스 PR 실행이 그것을 보여주고, 고치는 일은
  열린 카드 "Windows legs green for the release PR" 의 몫이에요. 이 SPEC 은 그 실패가 릴리스를 막는 것만 보장해요.
- Windows 에서 smoke harness 돌리기. `candidate-target.yml` 은 smoke 단계를 실행하지 않아요.
- Windows 실행을 nextest·rest 두 다리로 나누기(D-2).
- `check` 잡의 Scope(문서 전용 분류) 단계 바꾸기. 코드가 하나도 없는 릴리스 PR 은 릴리스가 아니라서 그 단계는 글자
  그대로 남아요(Risks).

## Requirements

- REQ-1: WHEN a pull request run's head branch name is `release/<40 lowercase hex>` or `release/rc/<40 lowercase hex>`, the decide step shall emit the full matrix and report a release pull request without calling the platform-sensitive classifier. · failure: none — decide 단계가 실패하면 `targets` 잡이 실패하고 `check` 를 건너뛴 `ci` 집계가 빨개지는 오늘의 실패 안전 경로 그대로예요
- REQ-2: WHEN a pull request run's head branch name has any other shape, the decide step shall answer the classifier's rule exactly as before this change. · failure: none — 이 경로의 줄은 바뀌지 않고 AC-3 이 그대로임을 고정해요
- REQ-3: The head branch name shall reach every script of the `ci` workflow through an environment variable, never through an expression inside a `run:` block.
- REQ-4: WHEN the decide step reports a release pull request, the `ci` workflow shall run the full check set on windows-x86_64 by calling the reusable target workflow of this repository with the target, runner, timeout and nextest inputs of the candidate verification's Windows job, on the pull request's head commit.
- REQ-5 (REQ-4): IF the decide step reports no release pull request or reports nothing, THEN the Windows job shall be skipped.
- REQ-6: WHEN the decide step reports a release pull request, the `ci` aggregate shall succeed only when the matrix result and the Windows check set both succeeded.
- REQ-7 (REQ-6): IF the decide step reports a release pull request and the Windows check set did not succeed, THEN the `ci` aggregate shall fail.
- REQ-8: WHEN the decide step reports no release pull request, the `ci` aggregate shall judge the matrix result as before this change and accept a skipped Windows job. · failure: none — 집계의 남은 조건은 오늘 그대로예요
- REQ-9: WHEN a pull request's head branch name begins with `release/`, the `windows-build` job shall run as it does after a merge. · failure: none — 이 잡은 필수 검사가 아니고 결과만 보여줘요
- REQ-10: The release names the decide step accepts shall be exactly the names the release workflow's plan step numbers.
- REQ-11: Each support document that states which CI job runs on Windows shall name the Windows check set of a release pull request.

실패 경로: REQ-4 → REQ-5, REQ-6 → REQ-7. 분류기와 집계의 실패 안전은 REQ-1·REQ-2·REQ-8 의 `failure: none` 근거예요.
"검사 세트"는 `tools/local_check.sh` 가 도는 단계(fetch, test, doctest, clippy, fmt, fmt-ops, unittest, diff-check)예요.
요구 문장의 "40 lowercase hex"는 `0-9a-f` 40자예요.

## Acceptance

- AC-1 (REQ-1, REQ-3): decide 단계의 진짜 스크립트를 임시 폴더에서 돌려요(가짜 `gh`·`python3` 는 호출을 기록해요).
  head 가지가 `release/<40자 hex>` 나 `release/rc/<40자 hex>` 인 `pull_request` 는 `matrix=` 에 `full=` 목록을,
  `release=true` 를 내고 분류기와 files API 를 부르지 않아요. 이름은 `env` 로만 들어가요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(a_release_pull_request_gets_the_full_matrix_and_the_windows_flag)'`
- AC-2 (REQ-10): 같은 이름 목록으로 decide 스크립트와 `release.yml` plan 스크립트를 둘 다 돌려요. plan 이 받는
  이름이면 decide 가 `release=true` 를 내고, plan 이 거절하는 이름(`release/` 뒤가 39자·41자 hex 인 이름, 대문자 hex 인
  이름, `release/v…`, `candidate/<40자 hex>`, `Release/<40자 hex>`, `release/<40자 hex>/x`)이면 decide 가
  `release=false` 를 내요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(the_decide_step_and_the_release_workflow_accept_the_same_branch_names)'`
- AC-3 (REQ-2): 일반 PR 과 push·`workflow_dispatch` 의 decide 결과가 오늘과 같아요. 분류기가 "필요 없음" 이면 `linux=`,
  "필요" 나 실패면 `full=`, push·`workflow_dispatch` 는 `full=` 이고 모두 `release=false` 예요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(an_ordinary_pull_request_keeps_the_classifier_rule) + test(push_and_dispatch_keep_the_full_matrix)'`
- AC-4 (REQ-4, REQ-5): `ci.yml` 의 Windows 잡은 `./.github/workflows/candidate-target.yml` 을 부르고, `target`·`runner`·
  `timeout`·`nextest_asset`·`nextest_sha256` 다섯 입력이 `candidate.yml` 의 Windows 잡과 글자 그대로 같고, `sha` 는
  PR head 커밋이며, `if` 는 decide 의 `release` 출력이 `true` 일 때만이에요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(the_windows_job_calls_the_candidate_target_with_the_candidates_windows_inputs)'`
- AC-5 (REQ-6, REQ-7, REQ-8): 집계 잡의 진짜 단계 스크립트를 결과 표로 돌려요. 매트릭스 성공에 Windows 성공은 통과,
  `release=false` 에 Windows 건너뜀은 통과, `release=true` 에 Windows 실패·취소·건너뜀·결과 없음은 실패, 매트릭스 실패는
  어떤 경우에도 실패, decide 출력이 없고 둘 다 건너뛴 경우(`targets` 실패)는 실패예요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(the_aggregate_fails_when_a_release_pull_requests_windows_check_did_not_succeed)'`
- AC-6 (REQ-9): `windows-build` 의 `if` 는 오늘의 식 뒤에 `|| startsWith(github.head_ref, 'release/')` 를 더한 글자예요. 앞의
  식이 참인 push 와 `workflow_dispatch` 는 오늘처럼 실행하고, pull request 는 `release/` 로 시작하는 head 에서만
  실행해요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(the_build_job_runs_on_a_release_pull_request_and_as_before_after_a_merge)'`
- AC-7 (REQ-3): `github.head_ref` 는 `ci.yml` 의 `run:` 블록 어디에도 없고 `env:` 와 `if:` 에만 있어요.
  · verify: `cargo nextest run --offline --locked --test release_pr_three_os -E 'test(the_head_branch_name_reaches_scripts_through_env_only)'`
- AC-8 (REQ-11): `docs/support-matrix.md` 의 Windows 행이 릴리스 PR 에서 검사 세트를 돌린다고 적어요.
  · verify: `grep -n 'release/' docs/support-matrix.md`
- AC-9 (REQ-1..REQ-9): 이웃 — `check affected` 영수증이 PASS 고(바뀐 Python 고정 세 곳을 포함), `actionlint` 가 0 이고,
  SPEC 검사가 오류 0 이에요.
  · verify: `moai check affected <base>..<head>`; `actionlint .github/workflows/ci.yml .github/workflows/candidate-target.yml`; `moai spec check SPEC-20261120-release-pr-three-os`
- AC-10 (REQ-1, REQ-4, REQ-6): 관측 — 첫 릴리스 PR 의 `ci` 실행에 매트릭스 다리와 Windows 잡이 모두 있고 결과가
  초록이며, run id 가 카드에 적혀요. 레인은 push 도 PR 도 열지 않으므로 릴리스는 사용자의 선택을 기다려요. 그 전에는
  `not_run` 이에요. 카드에 적은 run id 로 다시 읽어요.
  · verify: `gh run view RUN_ID --json jobs`

## Scope

- write: `.github/workflows/ci.yml`, `.github/workflows/candidate-target.yml`(머리 주석 한 줄만),
  `tests/release_pr_three_os.rs`(새 파일), `tests/github_assets_test.py`(바뀐 줄을 적은 고정 세 곳),
  `docs/support-matrix.md`(Windows 행의 문장), 이 SPEC 파일
- 범위 확장 요청(새 테스트 파일의 읽기 등록 한 줄, `test_tables` 단계가 가리키는 쪽 하나만, 줄 임대):
  `tools/repo_scans.json` 또는 `tests/repo_scans_table.rs`
- read-only: `.github/workflows/candidate.yml`, `.github/workflows/release.yml`, `.github/workflows/audit-gate.yml`,
  `.github/workflows/check-guard.yml`, `tools/ci_scope.py`, `tools/ci_superseded.py`, `tools/local_check.sh`,
  `tools/candidate_summary.py`, `tests/release_workflow_test.py`, `src/flow.rs`
- forbidden: 필수 검사 이름, ruleset, 일반 PR 의 매트릭스와 분류기, `release.yml`, 실제 HOME, push·PR·tag, 다른 카드의 가지

## Plan

1. 이 SPEC(판 2)을 커밋하고 카드에 연결해요. 승인 노트를 받은 뒤에만 다음 걸음을 해요.
2. RED: `tests/release_pr_three_os.rs`(임시 폴더, 가짜 `gh`·`python3`, 진짜 단계 스크립트)와 새 파일의 읽기 등록을
   커밋해요. `tests/github_assets_test.py` 의 고정 세 곳을 새 줄에 맞춰요. 바꾸기 전에 일곱 테스트가 실패하는지 봐요.
3. GREEN: `ci.yml` — decide 단계의 `env` 한 줄, 릴리스 판별과 `release` 출력, 분류기 가드 한 줄; `targets` 잡의
   `outputs` 한 줄; `candidate-target.yml` 을 부르는 Windows 잡; 집계 잡의 `needs` 와 단계; `windows-build` 의 `if`
   한 줄; 머리 주석. `candidate-target.yml` 머리 주석 한 줄.
4. 문서: `docs/support-matrix.md` 의 Windows 행.
5. 같은 head 의 `moai check affected` 영수증, 새 가드 테스트마다 뮤턴트 하나, `actionlint`, `moai trailer-check`,
   READY(AC-10 은 `not_run` 으로 적어요).
6. 착지 뒤(리더): 첫 릴리스 PR 전에 "Windows legs green for the release PR" 카드를 끝내고, 첫 실행의 run id 를
   카드에 적어요.

## Risks

- 첫 릴리스 PR 의 Windows 잡은 빨강일 수 있어요(Windows 전용 테스트 실패, Windows 대상 clippy 경고). 그게 이 SPEC 의
  목적이에요: 실패가 머지를 막아요. 고치는 일은 이 SPEC 밖이고, 착지 순서는 Plan 6 과 같아요.
- `pull_request` 실행은 머지 커밋의 맥락에서 돌아요(공식 문서: `pull_request_target` 이 "the default branch … rather
  than in the context of the merge commit, as the `pull_request` event does" 에서 돈다고 대조). 그래서 첫 릴리스 PR 은
  새 `ci.yml` 과 `candidate-target.yml` 을 스스로 가져와요. 어느 쪽 워크플로 파일이 실행되는지는 문서의 함축이고
  관측은 아니에요. 기반 가지의 옛 `ci.yml` 이 돌면 Windows 잡이 없고 AC-10 이 그것을 잡아요.
- `release/<40자 hex>` 처럼 SHA 처럼 보이는 가지 이름의 PR 이 `pull_request` 실행을 시작하는지 오늘은 관측되지
  않았어요. 문서는 그런 이름이 `pull_request_target` 실행을 시작하지 않을 수 있다고만 적어요(Sources 2). 릴리스
  워크플로의 Result 도 실제 실행을 하지 않았다고 적어요. AC-10 이 관측해요.
- Windows 잡은 PR head 커밋을 검사하고 linux·macOS 다리는 머지 커밋을 검사해요(D-3). 원격 기반 가지에 head 에 없는
  커밋이 있으면 둘이 달라요. RC1 에서는 통합이 로컬에서 끝나고 릴리스 PR 하나만 열려서 그럴 일이 없다고 보지만
  관측은 아니에요.
- `candidate-target.yml` 의 Windows 잡이 오늘 초록인지는 알지 못해요(Windows 실행 증거 없음). Windows 러너에 `python3`
  명령이 있다는 관측은 2026-11-20 의 러너 실행 하나예요. 오늘 읽은 이미지 목록은 `Python 3.12.10` 과 `Bash
  5.3.15(2)-release` 를 적고 ripgrep 은 적지 않아요. 첫 실행이 확인해요.
- `windows-build` 의 `if` 는 접두사 `release/` 만 봐요(식은 글자 수를 세지 못해요). `release/notes` 처럼 생긴 PR 도
  빌드 잡을 돌려요. 필수 검사가 아니고 빌드만 돌므로 해롭지 않아요. 테스트 다리는 decide 의 정확한 모양으로만
  정해져요.
- Scope(문서 전용 분류) 단계는 바꾸지 않아요. 문서 뿐인 릴리스 PR 은 무거운 단계를 건너뛰어요. 그런 PR 은 배포할 코드가
  없어서 릴리스가 아니에요.
- `ci.yml` 은 공유 파일이에요. 일반 경로의 줄은 바꾸지 않고 decide 의 가드 한 줄만 늘어나요. 고정 세 곳은 같은 변경에서
  맞춰요.
- `candidate-target.yml` 의 입력·단계가 바뀌면 릴리스 PR 의 Windows 실행도 바뀌어요. AC-4 가 다섯 입력이 `candidate.yml`
  과 같은지 고정해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 릴리스 PR 은 head 가지 이름으로 판별하고, 이름은 `env` 로만 스크립트에 들어가요 | `github.head_ref` 는 PR 올린 사람이 정하는 값이에요. 공식 문서는 신뢰할 수 없는 입력을 "an intermediate environment variable" 로 넘기기를 권해요(Sources 1). 이름은 `release.yml` plan 단계와 같은 모양이에요(REQ-10, AC-2) | 라벨로 판별하기(라벨은 쓰기 권한이면 누구나 바꿔요) |
| D-2 | Windows 는 `candidate-target.yml` 을 부르는 잡 하나로 검사 세트 전부를 돌려요. `check` 매트릭스에 Windows 두 다리를 넣지 않아요 | 새 코드가 가장 적어요. Windows 용 처리가 한 곳에만 있어요. smoke harness 를 Windows 에서 돌릴 필요가 없어요(그건 열린 카드의 몫). 판 1 의 두 다리 설계는 smoke 단계가 생기기 전의 설계예요 | `full=` 목록에 Windows 두 항목을 더하고 `check` 잡에 Windows 용 처리를 넣어요 |
| D-3 | Windows 잡은 PR head 커밋을 검사해요. 머지 커밋이 아니에요 | 릴리스 워크플로가 빌드하고 tag 가 가리키는 커밋이 head 예요(`release.yml`: "A push runs this file from the pushed commit") | `sha` 입력을 머지 커밋으로 바꿔요 |
| D-4 | Windows 결과는 집계 잡이 잡 결과로 판정해요. `release=true` 면 성공만 통과, 아니면 건너뜀도 통과예요 | 필수 검사 이름은 그대로 두고(변경 없음) Windows 실패가 머지를 막게 하는 가장 작은 변경이에요 | Windows 잡을 필수 검사 목록에 더하기(ruleset 변경, 사용자 선택) |
| D-5 | `windows-build` 는 `needs` 를 두지 않고 `if` 한 줄의 접두사 검사만 늘려요 | `needs: [targets]` 를 두면 push 실행의 빌드 잡이 decide 잡을 기다려요. 일반 경로의 시간을 바꾸지 않아요 | `needs` 로 바꾸고 decide 의 `release` 출력을 읽어요 |
| D-6 | 새 테스트는 Rust 고, 워크플로의 진짜 단계 스크립트를 임시 폴더에서 돌려요. 글자만 맞추지 않아요 | 판 1 의 감사가 "문자열만 맞추는 테스트 대신 단계 스크립트를 실제로 도는 테스트" 를 요구했어요. 테스트는 Rust 만이에요 | 없음(프로젝트 규칙) |
| D-7 | 첫 릴리스 PR 의 관측(AC-10)은 레인 증거에 `not_run` 으로 남아요 | push 와 PR 은 사용자의 선택이에요. 카드는 착지 뒤에도 그 기준이 열린 채로 있을 수 있어요(`card finish` 는 열린 기준을 적어요) | 관측을 이 SPEC 밖의 카드로 옮기기 |
| D-8 | 새 SPEC 이 아니라 2026-11-20 SPEC 의 판 2 로 써요 | 결과와 카드가 같고, 판 1 의 감사 기록이 그 아이디에 붙어 있어요. 판 1 은 통합 가지에 들어오지 않아서 승인을 다시 받아요 | 새 아이디로 옮기고 카드 연결을 바꿔요 |

## Sources

읽은 시각은 2026-11-24 UTC 이고, 모두 본문을 직접 내려받아 읽었어요. sha256 은 내려받은 본문 전체의 값이에요.

1. GitHub Docs, Secure use reference — `https://docs.github.com/en/actions/reference/security/secure-use`
   (06:46:25Z 에 읽고 06:53:53Z 에 sha256 `46385e58650cb066ced1c31c4cfeed1ff97776d9992255fe938d48176b3ba8c1`): "For inline scripts, the preferred approach to handling
   untrusted input is to set the value of the expression to an intermediate environment variable."
2. GitHub Docs, Events that trigger workflows — `https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows`
   (06:47:31Z, sha256 `f24a7813b65939cdbeb6b4f0d594242c04648e285b2b15f5129d684e50f4eac8`): `pull_request` 의 `GITHUB_SHA` 는 "Last merge commit on the
   `GITHUB_REF` branch"; `pull_request_target` 은 "runs in the context of the default branch of the base repository,
   rather than in the context of the merge commit, as the `pull_request` event does."; "branches with names that
   match certain patterns (such as those which look similar to SHAs) may not trigger workflows with the
   `pull_request_target` event."
3. GitHub Docs, Reusing workflows — `https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows`
   (06:52:56Z, sha256 `275f8bb02de6ed3f481872c87be49504cd5414a1ad53ae4591c4d6b5155442bc`): `./.github/workflows/{filename}` 은 같은 저장소의 재사용 워크플로이고 "the called
   workflow is from the same commit as the caller workflow".
4. GitHub Docs, Workflow syntax for GitHub Actions — `https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax`
   (06:53:53Z, sha256 `5d56229db925f0bb8ea99e3b77cd27713868de8bf7073ebe4c39f9d4e8b48017`): Windows 에서 `shell: bash` 는 "the bash shell included with Git for Windows";
   러너 라벨 `windows-2025`; "A matrix will generate a maximum of 256 jobs per workflow run."
5. GitHub Docs, Variables reference — `https://docs.github.com/en/actions/reference/workflows-and-actions/variables`
   (06:53:53Z, sha256 `7f874dc0ec584e3c07838c8b6a85e6db732e6cba34b7c322ceef0544731fdbc2`): `GITHUB_HEAD_REF` 는 "The head ref or source branch of the pull request in a
   workflow run. This property is only set when the event that triggers a workflow run is either `pull_request` or
   `pull_request_target`."
6. GitHub Docs, Contexts reference — `https://docs.github.com/en/actions/reference/workflows-and-actions/contexts`
   (06:53:53Z, sha256 `64fe9abcfb6ae5eabb2b7f2cc79127215eeec5f403b1668b5bccaa7aefcf8c2d`): `github.head_ref` 와
   `github.event_name` 의 뜻.
7. actions/runner-images, `images/windows/Windows2025-Readme.md`(main, 06:46:06Z, sha256
   `af7621a27ad87d8a61c96f7751a336d23802a2311c877738ef361733d4561e74`): Image Version
   `20261116.335.1`, `Bash 5.3.15(2)-release`, `Python 3.12.10`; ripgrep 항목 없음.
8. nextest 릴리스 `cargo-nextest-0.9.151` 의 자산 목록 — `https://api.github.com/repos/nextest-rs/nextest/releases/tags/cargo-nextest-0.9.151`
   (06:46:05Z 에 읽고 06:53:53Z 에 sha256 `0f6d5ee500f902a09fd8c1596c021ba5ec6f452178d35ec863c42b74db50322a`): `cargo-nextest-0.9.151-x86_64-pc-windows-msvc.tar.gz` 의 digest
   `sha256:e51bdb4b5883df48f335381b287378edb244318fac65ae0adaf6b9e69298246b` — `candidate.yml` 의 값과 같아요.
9. BurntSushi/ripgrep 14.1.1 `ripgrep-14.1.1-x86_64-pc-windows-msvc.zip` 을 내려받아 계산(06:45:47Z):
   `d0f534024c42afd6cb4d38907c25cd2b249b79bbe6cc1dbee8e3e37c2b6e25a1` — `candidate-target.yml` 의 값과 같아요.

채택: 1(env 으로만), 3(같은 커밋의 재사용 워크플로), 4·7·8·9(`candidate-target.yml` 의 Windows 입력과 고정값을 그대로).
거절: 없음. 오늘 관측하지 않은 것: Windows 러너의 실제 실행, ruleset 의 필수 검사 목록, SHA 처럼 보이는 가지 이름의
`pull_request` 실행.

## Result

구현 에이전트가 채워요. 이 커밋은 SPEC 판 2 까지예요.
