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
- REQ-2: The smoke runner shall build the environment of every step from nothing but HOME, a PATH of its own folder plus `/usr/bin` and `/bin`, a TMPDIR inside the temporary root, locale, terminal and git settings, never from the caller's environment.
- REQ-3 (REQ-2): IF the temporary folder base lies inside the caller's real home folder, named by HOME or by the account entry, in any spelling of it that names the same folder, THEN the smoke runner shall refuse before it creates anything and exit with 2.
- REQ-4: The smoke runner shall print one trace line per step with its name, its outcome and its time, and end its output with one line `smoke: <area> pass|fail steps=<n> …`, exit 0 on pass.
- REQ-5 (REQ-4): IF a step fails or raises an error the runner did not expect, THEN the smoke runner shall print the failing step's command, exit code and output and stop with exit 1.
- REQ-6: The install area shall stage the binary with its skill package and agent sources in a temporary prefix and check the version signature and the install inspection envelope.
- REQ-7: The update area shall run the preview that writes nothing, the confirm with the preview's digest that writes the user layer, and a second preview that reports nothing to apply.
- REQ-8: The project area shall run the detection and the plan that write nothing, then the apply with the plan's digest, and check that the project profile exists.
- REQ-9: The launcher area shall run the dry run and a pty run with a fake host first on PATH, and check that the host got its arguments, a terminal and the temporary HOME.
- REQ-10 (REQ-1): IF the area is unknown or the binary is missing or not executable, THEN the smoke runner shall exit with 2 and say which.
- REQ-11: The smoke runner shall remove its temporary root when it ends, after a pass, a failure, an unexpected error or SIGTERM, unless it is given `--keep`.
- REQ-12 (REQ-11): IF the temporary root cannot be removed, THEN the smoke runner shall name it with the reason and exit with 1.

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-11): `install`, `update`, `project`, `launcher` 영역이 각각 빌드한 바이너리로 pass이고, 단계마다 추적 줄이 하나씩 찍히고, 마지막 줄이 `smoke: <영역> pass steps=<n>`이고 종료 코드가 0이에요.
  verify: `python3 -m unittest tests.smoke_harness_test.AreasTest`
- AC-2 (REQ-4, REQ-5): 실패하는 바이너리로 돌리면 실패한 단계 이름·명령·종료 코드·출력이 찍히고 마지막 줄이 `smoke: … fail … failed=<단계>`이고 종료 코드가 1이에요. HOME에 파일뿐 아니라 폴더나 링크가 생겨도 쓴 것으로 세고, harness가 예상하지 못한 오류(읽을 수 없는 파일 복사 등)도 실패한 단계와 마지막 줄로 나와요(Traceback 없음).
  verify: `python3 -m unittest tests.smoke_harness_test.FailureTest`
- AC-3 (REQ-2, REQ-3): 호출자의 HOME이 다른 폴더(미끼)일 때 미끼는 한 바이트도 바뀌지 않고, 단계가 본 HOME은 미끼가 아니고, 단계가 받은 PATH는 `<root>/fakebin:/usr/bin:/bin`이고 TMPDIR은 `<root>/tmp`이며 호출자의 환경 변수는 닿지 않고, 설치 점검이 말한 상태 폴더가 임시 HOME 밖이면 실패해요. 임시 폴더 기반이 실제 HOME 안이면(HOME과 계정 항목 둘 다, 다른 글자 크기·`/System/Volumes/Data` 철자 포함 파일 동일성으로 비교) 아무것도 만들지 않고 2로 거절해요.
  verify: `python3 -m unittest tests.smoke_harness_test.IsolationTest`; `python3 -m unittest tests.smoke_harness_test.IdentityTest`
- AC-4 (REQ-10, REQ-11, REQ-12): 모르는 영역과 없는 바이너리는 2이고, 끝나면 임시 폴더가 없어지고(단계가 권한 000 폴더를 남겨도, SIGTERM으로 끝나도 — 마지막 줄 `failed=interrupted`) `--keep`이면 남아요. 지울 수 없으면 `smoke: could not remove root=<경로>: <이유>`를 찍고 `failed=cleanup`으로 1이에요.
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
- 임시 HOME에서 `git`을 부르면 macOS의 `/usr/bin/git` shim이 호출마다 라이선스 확인을 해서 느려져요 → git의 절대 경로를 harness가 한 번 찾아 harness 폴더에 링크하고(호출자의 PATH 폴더는 단계에 주지 않아요), 있으면 `DEVELOPER_DIR`을 Command Line Tools 폴더로 줘요.
- 파일 동일성 비교는 실제 HOME과 그 위 폴더들을 `stat`해요(읽기·열기는 하지 않아요). 같은 폴더가 두 철자로 보이는 길은 대소문자 무시 볼륨과 firmlink(`/System/Volumes/Data`)뿐이라고 보고 시험이 둘 다 시험해요; Linux에서는 돌려 보지 않았어요.
- PATH의 `moai`는 `$HOME/.moai-cli/path`를 따라가는 래퍼라 임시 HOME에서 깨져요 → harness는 PATH를 쓰지 않고 `--bin`(또는 빌드 폴더의 `moai-cli`)을 직접 불러요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 환경은 비어 있는 데서 쌓아요(상속하지 않아요) | 호출자의 `ANTHROPIC_API_KEY`나 `CLAUDE_*`가 단계에 닿으면 실제 키·호스트를 쓸 수 있어요 | 없음 |
| D-2 | 임시 폴더 기반이 실제 HOME 안이면 거절해요; 비교는 철자가 아니라 파일 동일성이에요 | `TMPDIR`이 `~/tmp` 같은 곳이면 "임시"가 실제 HOME 안이 돼요; 철자 비교는 `/users/<이름>`(다른 글자 크기)과 `/System/Volumes/Data/Users/<이름>`을 놓쳐요(감사 B1) | `--keep`과 무관하게 거절해요; 기반을 바꾸면 돼요 |
| D-3 | `install`은 릴리스를 받지 않고 설치된 모양을 꾸려 확인해요 | 내려받기는 네트워크예요; `update`가 바이너리 옆 `skills/moai`·`agents`를 요구해요(`src/update.rs`) | 설치 스크립트 영역은 후속 SPEC |
| D-4 | 런처의 호스트는 가짜 실행 파일이에요 | 실제 호스트는 키와 네트워크가 필요해요 | 없음 |
| D-5 | 파이썬(표준 라이브러리만)과 얇은 셸 래퍼 | 새 의존성을 두지 않고 `tools/`의 다른 도구와 같아요 | 없음 |
| D-6 | 시험의 미끼 폴더 훑기는 `os.scandir` 재귀예요 | `tests/repo_scans_test.py`가 `REPO_ROOT`와 `.rglob(`가 함께 있는 시험 파일을 "저장소를 훑는 시험"으로 표시해요; 이 시험은 임시 미끼 폴더만 훑지만 예외 목록은 이 SPEC의 쓰기 범위 밖이에요 | 예외 목록에 올리면 `rglob`으로 되돌려도 돼요 |
| D-7 | 단계에는 호출자의 PATH도, 호출자의 git이 있는 폴더도 주지 않아요; `git`은 harness 폴더에 링크해요 | 버전 관리자의 shim 폴더가 PATH에 있으면 설치된 `moai`가 단계에 보여요(감사 N1) | 없음 |
| D-8 | 지우기 전에 단계가 남긴 폴더의 권한을 되돌리고, 그래도 못 지우면 이름과 이유를 말하고 실패로 처리해요 | 말없이 남기면 임시 HOME이 쌓여요(감사 B2) | 없음 |
| D-9 | SIGTERM은 실행 중인 명령을 멈추고 뿌리를 지운 뒤 `failed=interrupted`로 끝나요 | 신호로 끝나면 뿌리와 자식이 남았어요(감사 N2) | 없음 |

## Result

감사 FAIL(B1·B2·B3)을 같은 가지에서 고친 뒤의 결과예요. 아래는 모두 이 가지의 현재 코드로, 빌드 폴더의 `moai-cli`(head `e005f8b17b` 기반 빌드)로 쟀어요.

- AC-1: met — `python3 -m unittest tests.smoke_harness_test.AreasTest` 5개 ok(install·update·project·launcher·all). 추적 줄이 단계마다 하나고 마지막 줄이 `smoke: <영역> pass steps=<n>`이에요. 이번 `tools/smoke/run.sh all` 실측은 `pass steps=13 time=10.0s`예요(예전 측정 22.5초와 다른 실행이라 둘 다 5분 안이에요).
- AC-2: met — `…FailureTest` 9개 ok(예전 7개 + 새 2개): HOME에 폴더·링크가 생기면 쓴 것으로 세는 시험(`holds 2 entries`)과, 읽을 수 없는 파일을 복사하다 harness가 예상하지 못한 오류가 나도 `[FAIL] stage-package … unexpected`와 마지막 줄 `failed=stage-package`로 나오고 Traceback이 없고 뿌리가 지워지는 시험이에요(감사 N3·N4).
- AC-3: met — `…IsolationTest` 5개 ok(예전 3개 + 새 2개)와 `…IdentityTest` 3개 ok. 새 시험: 단계가 받은 PATH가 `<root>/fakebin:/usr/bin:/bin`이고 TMPDIR이 `<root>/tmp`이라는 것(호출자의 PATH 폴더와 TMPDIR을 일부러 줬어요), 설치 점검이 말한 상태 폴더가 임시 HOME 밖이면 `not inside the temporary HOME`으로 실패하는 것, 임시 기반의 모든 철자 — 보통 철자, 다른 글자 크기, `/System/Volumes/Data` 철자 — 가 파일 동일성으로 실제 HOME 안이라고 인정되어 거절되는 것(이 컴퓨터에서 세 철자가 모두 존재함을 확인했고 시험이 건너뛰지 않았어요), HOME과 계정 항목 둘 다 실제 HOME으로 세는 것(감사 B1·B3).
- AC-4: met — `…UsageTest` 6개 ok(예전 3개 + 새 3개): 단계가 권한 000 폴더를 남겨도 뿌리가 지워지는 것, 지우기가 실패하면(`shutil.rmtree`를 실패시켜서) `smoke: could not remove root=<경로>: /the/locked/folder: Permission denied`가 한 줄 나오고 마지막 줄이 `failed=cleanup`, 종료 1이고 뿌리를 찾을 수 있는 것, SIGTERM을 보내면 `failed=interrupted`로 끝나고 뿌리가 지워지는 것(감사 B2·N2).
- AC-5: met — `python3 -m unittest tests.smoke_harness_test` 28 ok(예전 18); `cargo test --locked --offline --test ac_command_lint` 5, `home_isolation_scan` 18, `spec_cli` 17, `git_env_scrub` 3 passed; `python3 -m unittest tests.repo_scans_test` 25·`ci_scope_test` 19·`local_check_test` 36·`source_manifest_test` 43 OK; `spec check` errors 0·warnings 0. 시험 먼저: 예전 harness(`f15e3a074d`의 `smoke.py`)에 새 시험 파일을 돌리면 11개가 실패해요(실패 7·오류 4) — 새 시험 9개와, 문구가 `files`에서 `entries`로 바뀐 기존 시험 2개. 새 시험 하나(상태 폴더가 HOME 밖이면 실패)는 예전 코드도 통과해요: 그 검사는 있었지만 시험이 없었어요(변이로 확인).
- 변이: 이번에 24개를 새 코드에 만들어 돌렸고 24/24 KILLED예요. 격리 10개(환경 상속, 기반 검사 끄기, 호출자 HOME, 호출자 PATH 덧붙이기, 호출자 TMPDIR, 계정 홈 빼기, 상태 폴더 검사 끄기, 철자 비교로 되돌리기, git 폴더를 PATH에 넣기, 기반=HOME 허용), 지우기 4개(지우기 없음, 실패를 말없이 넘김, 권한 되돌리기 없음, 실패해도 통과), 그 밖 10개(예상 밖 오류, SIGTERM, 폴더·링크 안 셈, `--keep` 뒤집기, 실패 종료 0, 실패해도 계속, 버전 검사 느슨, HOME 비었는지·미리보기 쓰기·터미널 검사 끄기). 감사가 찾은 살아남은 4개(PATH·TMPDIR·계정 홈·상태 폴더 검사)가 여기에 들어 있고, 예전 Result의 "12/12 KILLED"는 제 목록에만 해당했어요(고쳤어요). 매번 `git checkout --`로 되돌렸고 끝난 뒤 `git status`가 깨끗했어요.
- 남은 것: (1) VERIFY 워크플로 글(`skills/moai/workflows/verify.md`)에 이 harness를 첫 점검으로 적는 일과 `source-manifest.json` 줄은 다른 카드의 예약과 겹쳐 이 SPEC의 쓰기 범위 밖이에요(카드의 두 번째 기준의 그 문장은 열려 있어요). (2) CI가 install·update 영역을 PR에서 돌리게 하는 일(카드의 세 번째 기준)은 `.github/workflows/**`가 범위 밖이라 열려 있어요. (3) 설치 스크립트의 내려받기는 하지 않았어요(D-3). (4) Linux에서는 돌려 보지 않았어요(감사 N6); 두 번째 철자 시험(대소문자·firmlink)은 macOS 전용이라 Linux에서는 철자가 하나뿐이면 그것만 시험해요. (5) `tests/` 전체를 찾아 도는 Python 단위 시험(`discover`)은 돌리지 않았고 위의 모듈만 돌렸어요. (6) 이 가지의 앞선 커밋 6개에는 `Moai-Verified` 줄이 없어요(그 규칙이 나중에 들어왔어요, 감사 N5); 이번 수리 커밋부터 있어요. SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.
