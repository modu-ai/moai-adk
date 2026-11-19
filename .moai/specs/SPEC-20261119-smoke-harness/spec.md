---
id: SPEC-20261119-smoke-harness
title: "One smoke harness the lanes and auditors share: the built binary in an empty temporary HOME"
status: draft
revision: 1
created: 2026-11-19
---

# One smoke harness the lanes and auditors share: the built binary in an empty temporary HOME

## Intent

레인과 감사자는 "시험이 통과했다"만으로 끝내지 않고, 빌드한 바이너리를 **빈 임시 HOME과 임시 저장소에서 실제로 돌려** 설치→업데이트→project→런처가 이어지는지 봐야 해요(Addy Osmani, "How the Claude Code team thinks about verification", claude.dev 2026-11-16: 끝에서 끝까지 돌려 보는 일이 시험보다 중요하다는 점). 지금은 사람마다 손으로 임시 HOME을 만들고 명령을 이어 붙여서, 어떤 단계가 어떻게 실패했는지 보고가 제각각이에요(`tools/local-rc1-isolation.sh`는 옛 RC1 명령 묶음이고, `tools/host_live_harness.py`는 실제 호스트를 재는 도구예요). 이 SPEC은 `tools/smoke/run.sh <area>` 하나를 만들어요: 같은 임시 환경, 같은 단계 기록, 실패하면 실패한 단계와 그 출력, 실제 HOME은 읽지도 쓰지도 않아요.

## Non-goals

- VERIFY 워크플로 글(`skills/moai/workflows/verify.md`)에 이 harness를 첫 점검으로 적는 일과 `skills/moai/references/source-manifest.json` 줄: 그 두 파일은 다른 카드의 예약과 겹쳐서 이 SPEC의 쓰기 범위 밖이에요(리더 결정으로 나중 조각).
- CI 워크플로(`.github/workflows/**`)에 이 harness를 거는 일: 범위 밖이에요.
- 설치 스크립트(`install/install.sh`)가 릴리스를 내려받는 길: 네트워크가 필요해서 하지 않아요. `install` 영역은 "설치된 모양"(바이너리와 옆의 묶음)을 꾸려 확인해요.
- 실제 호스트(Claude Code, Codex), 실제 키, 실제 endpoint, 네트워크: 쓰지 않아요. 호스트는 이 harness가 만든 가짜 실행 파일이에요.
- 단계 안의 명령이 쓰는 파일이 맞는지(내용)를 깊이 검사하는 일: 각 영역의 시험이 하고, 이 harness는 "끝까지 이어지는가"를 봐요.
- Windows: pty와 셸 스크립트를 쓰므로 macOS와 Linux만 대상이에요.

## Requirements

- REQ-1: The smoke runner shall run the steps of one area (`install`, `update`, `project`, `launcher`, or `all`) with the binary under test inside a fresh temporary HOME and a fresh temporary git project.
- REQ-2: The smoke runner shall build the environment of every step from nothing but HOME, PATH, TMPDIR, locale, terminal and git settings, never from the caller's environment.
- REQ-3 (REQ-2): IF the temporary folder base lies inside the caller's real home folder, THEN the smoke runner shall refuse before it creates anything and exit with 2.
- REQ-4: The smoke runner shall print one trace line per step with its name, its outcome and its time.
- REQ-5 (REQ-4): IF a step fails, THEN the smoke runner shall print the failing step's command, exit code and output and stop with exit 1.
- REQ-6: The install area shall stage the binary with its skill package and agent sources in a temporary prefix and check the version signature and the install inspection envelope.
- REQ-7: The update area shall run the preview that writes nothing, the confirm with the preview's digest that writes the user layer, and a second preview that reports nothing to apply.
- REQ-8: The project area shall run the detection and the plan that write nothing, then the apply with the plan's digest, and check that the project profile exists.
- REQ-9: The launcher area shall run the dry run and a pty run with a fake host first on PATH, and check that the host got its arguments, a terminal and the temporary HOME.
- REQ-10: The smoke runner shall end its output with one line `smoke: <area> pass|fail steps=<n> …` and exit 0 on pass.
- REQ-11 (REQ-1): IF the area is unknown or the binary is missing or not executable, THEN the smoke runner shall exit with 2 and say which.
- REQ-12: The smoke runner shall remove its temporary root when it ends unless it is given `--keep`.

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-10, REQ-12): `install`, `update`, `project`, `launcher` 영역이 각각 빌드한 바이너리로 pass이고, 단계마다 추적 줄이 하나씩 찍히고, 마지막 줄이 `smoke: <영역> pass steps=<n>`이고 종료 코드가 0이에요.
  verify: `python3 -m unittest tests.smoke_harness_test.AreasTest`
- AC-2 (REQ-5, REQ-10): 실패하는 바이너리로 돌리면 실패한 단계 이름·명령·종료 코드·출력이 찍히고 마지막 줄이 `smoke: … fail … failed=<단계>`이고 종료 코드가 1이에요.
  verify: `python3 -m unittest tests.smoke_harness_test.FailureTest`
- AC-3 (REQ-2, REQ-3, REQ-12): 호출자의 HOME이 다른 폴더(미끼)일 때 미끼는 한 바이트도 바뀌지 않고, 단계가 본 HOME은 미끼가 아니고, 호출자의 환경 변수는 단계에 닿지 않고, 임시 폴더 기반이 미끼 안이면 아무것도 만들지 않고 2로 거절해요.
  verify: `python3 -m unittest tests.smoke_harness_test.IsolationTest`
- AC-4 (REQ-11, REQ-12): 모르는 영역과 없는 바이너리는 2이고, 끝나면 임시 폴더가 없어지고 `--keep`이면 남아요.
  verify: `python3 -m unittest tests.smoke_harness_test.UsageTest`
- AC-5 (REQ-1…REQ-12): 이웃 검사.
  verify: `python3 -m unittest tests.smoke_harness_test`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test home_isolation_scan`

## Scope

- write: `.moai/specs/SPEC-20261119-smoke-harness/spec.md`, `tools/smoke/run.sh`, `tools/smoke/smoke.py`, `tests/smoke_harness_test.py`
- read-only: `src/update.rs`, `src/main.rs`, `tools/local-rc1-isolation.sh`, `tools/host_live_harness.py`, `skills/moai/`, `agents/`
- forbidden: 실제 HOME(읽기·쓰기), 네트워크, 실제 키·endpoint, 새 의존성, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`, `.github/workflows/**`, `src/**`

## Plan

1. 시험 먼저: `tests/smoke_harness_test.py`를 쓰고, harness 파일이 없는 상태에서 실패하는 것(RED)을 읽어요.
2. `tools/smoke/smoke.py`와 `tools/smoke/run.sh`를 구현해 GREEN으로 만들어요.
3. 영역별 시간과 이웃 검사를 재고, 변이 시험(보호 장치를 하나씩 끄기)을 돌려요.

## Risks

- 단계가 바이너리의 출력 모양(JSON 필드)에 기대서 바이너리가 바뀌면 harness가 깨져요 → 필요한 필드만 읽고(`ok`, `plan_digest`, `nothing_to_apply`, `home_state_root`), 모양이 다르면 실패한 단계와 받은 글을 그대로 보여 줘요.
- 임시 HOME에서 `git`을 부르면 macOS의 `/usr/bin/git` shim이 호출마다 라이선스 확인을 해서 느려져요 → git의 절대 경로를 harness가 한 번 찾고, 있으면 `DEVELOPER_DIR`을 Command Line Tools 폴더로 줘요.
- PATH의 `moai`는 `$HOME/.moai-cli/path`를 따라가는 래퍼라 임시 HOME에서 깨져요 → harness는 PATH를 쓰지 않고 `--bin`(또는 빌드 폴더의 `moai-cli`)을 직접 불러요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 환경은 비어 있는 데서 쌓아요(상속하지 않아요) | 호출자의 `ANTHROPIC_API_KEY`나 `CLAUDE_*`가 단계에 닿으면 실제 키·호스트를 쓸 수 있어요 | 없음 |
| D-2 | 임시 폴더 기반이 실제 HOME 안이면 거절해요 | `TMPDIR`이 `~/tmp` 같은 곳이면 "임시"가 실제 HOME 안이 돼요 | `--keep`과 무관하게 거절해요; 기반을 바꾸면 돼요 |
| D-3 | `install`은 릴리스를 받지 않고 설치된 모양을 꾸려 확인해요 | 내려받기는 네트워크예요; `update`가 바이너리 옆 `skills/moai`·`agents`를 요구해요(`src/update.rs`) | 설치 스크립트 영역은 후속 SPEC |
| D-4 | 런처의 호스트는 가짜 실행 파일이에요 | 실제 호스트는 키와 네트워크가 필요해요 | 없음 |
| D-5 | 파이썬(표준 라이브러리만)과 얇은 셸 래퍼 | 새 의존성을 두지 않고 `tools/`의 다른 도구와 같아요 | 없음 |
| D-6 | 시험의 미끼 폴더 훑기는 `os.scandir` 재귀예요 | `tests/repo_scans_test.py`가 `REPO_ROOT`와 `.rglob(`가 함께 있는 시험 파일을 "저장소를 훑는 시험"으로 표시해요; 이 시험은 임시 미끼 폴더만 훑지만 예외 목록은 이 SPEC의 쓰기 범위 밖이에요 | 예외 목록에 올리면 `rglob`으로 되돌려도 돼요 |

## Result

- AC-1: met — `python3 -m unittest tests.smoke_harness_test.AreasTest` 5개 ok(install·update·project·launcher·all), tip 빌드(`moai-cli 4.2647.0 (rust)`, head `e005f8b17b` 위의 이 가지)로 돌렸어요. 추적 줄이 단계마다 하나고 마지막 줄이 `smoke: <영역> pass steps=<n>`이에요(`all`은 13단계). 한 번의 `tools/smoke/run.sh all` 실측은 22.5초(install ≈3초, update ≈16초, project <1초, launcher ≈2초)이고 미끼 HOME은 비어 있었어요. 시험 먼저: 구현 전 RED에서 13개 모두 실패(`run.sh`가 없어서 exit 127).
- AC-2: met — `…FailureTest` 7개 ok: 항상 실패하는 바이너리(첫 단계 `version`에서 멈추고 `exit: 3`·출력·`failed=version`), JSON이 아닌 답(`failed=install-inspect`), Go 세대 바이너리(`version`), HOME에 쓰는 설치 점검(`home-untouched`)과 업데이트 미리보기(`preview`), 터미널이 없는 호스트와 다른 HOME을 보는 호스트(`host-in-pty`).
- AC-3: met — `…IsolationTest` 3개 ok: 호출자 환경(`PROBE_CALLER_VALUE`, `CLAUDE_CODE_SESSION_ID`)과 미끼 HOME 경로가 가짜 바이너리가 남긴 기록에 없고, 단계가 본 HOME 이름은 `home`이고 미끼는 한 바이트도 안 바뀌었어요; 임시 기반이 미끼 안이거나 미끼 그 자체이면 아무것도 만들지 않고 exit 2(`real home` 메시지); `--home` 같은 옵션은 없어요(exit 2).
- AC-4: met — `…UsageTest` 3개 ok: 모르는 영역·없는 바이너리·실행권 없는 바이너리는 exit 2이고, 끝나면 임시 폴더가 없어지고 `--keep`이면 남아요.
- AC-5: met — `python3 -m unittest tests.smoke_harness_test` 18 ok; `cargo test --locked --offline --test ac_command_lint` 5, `--test home_isolation_scan` 18, `--test spec_cli` 17, `--test git_env_scrub` 3 passed; `python3 -m unittest repo_scans_test`·`ci_scope_test`·`local_check_test`·`source_manifest_test` OK(처음에 `repo_scans_test`가 새 시험 파일을 걸어서 D-6으로 풀었어요); `spec check` errors 0·warnings 0; `trailer-check` 커밋 6개 문제 없음. 변이는 harness에 12개(환경 상속, 기반 검사 끄기, 호출자 HOME 쓰기, 정리 없음, `--keep` 뒤집기, 실패 종료 0, 실패해도 계속, 버전 검사 느슨, HOME 비었는지 검사 끄기, 미리보기 쓰기 검사 끄기, 기반=HOME 허용, 터미널 검사 끄기): 처음 10개 KILLED, 2개(HOME 비었는지·터미널)가 SURVIVED여서 가짜 바이너리 시험을 더했고 다시 돌려 12/12 KILLED예요.
- 남은 것: (1) VERIFY 워크플로 글(`skills/moai/workflows/verify.md`)에 이 harness를 첫 점검으로 적는 일과 `source-manifest.json` 줄은 다른 카드의 예약과 겹쳐 이 SPEC의 쓰기 범위 밖이에요(카드의 두 번째 기준의 그 문장은 열려 있어요). (2) CI가 install·update 영역을 PR에서 돌리게 하는 일(카드의 세 번째 기준)은 `.github/workflows/**`가 범위 밖이라 열려 있어요; 5분 이내는 실측 22.5초로 충족해요. (3) 설치 스크립트의 내려받기는 하지 않았어요(D-3). SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.
