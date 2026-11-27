---
id: SPEC-20261127-doctor-moai-layer
title: "moai doctor: moai layer diagnostics"
status: approved
revision: 1
created: 2026-11-27
approval:
  by: user
  at: 2026-11-27
  basis: "리더 결정 노트 10830 (USER-APPROVAL C-338 SPEC-20261127-doctor-moai-layer revision 1, 2026-11-27): 사용자가 직접 골랐어요(리더 세션의 질문 도구), \"기본값으로 승인 (Recommended)\"; 선택지 설명 \"세 기본값(D-2, D-3, D-5)으로 승인해요. 요구가 15개라 크기 경고가 있지만 한 SPEC으로 가고 점검 항목 순서대로 구현해요. kenan이 테스트부터 시작해요.\"; D-8 은 리더 판단으로 한 SPEC; 승인 범위는 이 SPEC revision 1 의 요구·기준·결정 전부이고, tools/repo_scans.json 을 고쳐 tier 2 가 되어도 이 승인이 그 SPEC 승인이며 전체 리뷰를 받아요; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요(레인은 자기 승인을 기록하지 않았어요)"
  digest: "sha256:e367fe850f08dfe193d41cf5c3260ecf8af4cccc3acedc82c8163fd9ded4e3ec"
---

# moai doctor: moai layer diagnostics

## Interview

- Q1: 로그인 만료 확인에 `claude auth status` 를 실행해도 되나요? → A: "기본값으로 승인 (Recommended)"(사용자의 선택, 2026-11-27, 리더 세션의 질문 도구; 리더 결정 노트 10830). 선택지의 설명은 `claude auth status` 를 `--auth-status` 옵션을 줄 때만 실행한다는 기본값(D-2)이에요. 이유: 공식 문서는 이 명령을 읽기 전용이라고 쓰지 않고, 변경 기록에는 토큰 갱신·잠금·로그아웃 수정이 있어요. 다른 선택지(항상 확인, 릴리즈 뒤로 미룸)는 고르지 않았어요.
- Q2: 스킬 로드 확인과 상태줄 가짜 입력 실행을 이번 조각에 넣을까요? → A: 같은 선택(노트 10830)의 기본값 D-3: 스킬 로드·상태줄 시험·업데이트 안내는 뒤 조각이에요. 둘 다 사용자의 명령이나 실제 세션을 실행하는 일이라 기본 점검에서 빼요.
- Q3: 최소 버전 표에서 카드 본문과 달라진 값(Windows 세션 간 메시지, `--discard-unpushed`)을 어떻게 쓸까요? → A: 같은 선택(노트 10830)의 기본값 D-5: 조사에서 읽은 값을 쓰고 확인하지 못한 칸은 `unverified` 로 둬요.
- Q4: 요구가 15개라 크기 경고가 있어요, 나눌까요? → A: 리더 판단(노트 10830, D-8): 한 SPEC 으로 가고 점검 항목 순서대로 구현해요. 크기 경고는 받아들여요.

## Intent

moai 가 도는 환경에서 "왜 moai 가 안 도는지"를 사용자가 스스로 알아볼 수 있게 해요. `moai doctor` 는 **moai 층만** 진단해요: Claude Code 최소 버전, 로그인·API 키 경고, 관리 잠금, 같은 컴퓨터 메시지 수신 값, 호스트 능력 표. Claude Code 자체의 점검(`claude doctor`, `/status`, `/context`)은 대신 하지 않고 어디서 보는지만 알려요. 읽기 전용이고, 비밀 값은 변수 이름으로만 나와요.

이 SPEC 은 카드의 첫 조각(SPEC)이에요. 읽어서 확인한 사실(2026-11-27, NeXT `aadb78141a`):

- `moai doctor` 는 없어요(`unrecognized subcommand`). 비슷한 `claude-doctor <PROJECT>`(`src/claude.rs` 의 `doctor`)는 프로젝트의 AGENTS.md 를 Claude Code 가 읽는지만 예측하고, `claude --version` 을 5초 상한으로 한 번 실행해요. 이 SPEC 은 그 프로브를 그대로 쓰고 `claude-doctor` 는 바꾸지 않아요.
- Claude 설정 폴더는 `src/host_home.rs`(`Host::Claude`: `CLAUDE_CONFIG_DIR` 은 계정 홈일 때만 인정, 상대·빈 값은 변수 이름만 말하고 거절)가 정해요. doctor 는 이것을 부르고 폴더 찾기를 새로 만들지 않아요.
- `crossSessionInbound: accept` 는 헤드리스 실행(`src/headless_run.rs`)과 `SPEC-20261122-session-messages` D-8 이 다뤄요. 권한 우회로 도는 레인(`-l`·`-a`, 2026-11-25 복귀)은 값이 accept 가 아니면 같은 컴퓨터 메시지를 보류해요.
- 관리 잠금 네 키를 읽는 코드는 저장소에 없어요.
- 카드가 쓰인 뒤 바뀐 것: Codex·OMP 철수(능력 표에 칸이 없어요), `moai-cli` 이름 철수(문장은 이름 대신 파일 경로를 말해요), 런처 `-l`/`-a` 복귀.

읽는 사람: 승인하는 사용자, 구현하는 레인, 독립 감사자.

## Non-goals

- Claude Code 자체의 점검을 대신하지 않아요(`claude doctor`, `/status`, `/doctor prompt-audit`, `/skill-doctor`, `/usage`, `/context all`). 알려 주기만 하고 실행하지 않아요(REQ-13).
- 쓰기가 없어요: 설정·원장·HOME·프로젝트 어디에도 쓰지 않고, 네트워크를 쓰지 않아요.
- 이번 조각에 넣지 않아요(뒤 조각 후보, D-3): `system/init` 로 스킬 로드 확인(실제 `claude -p` 세션이 필요하고 비용이 들어요), 상태줄 가짜 입력 실행(사용자가 설정한 명령을 실행해요), `FUNCTION_HOOKS` 잔여 점검(저장소 어디에도 이 이름이 없고 공식 문서에서 읽은 출처도 없어요: unverified), 업데이트 명령(Homebrew·WinGet) 안내(조사에서 읽지 않았어요: unverified).
- `claude-doctor` 를 합치거나 바꾸지 않아요(D-4).
- Codex·OMP 칸은 없어요.
- 정책 도우미(`policyHelper`) 실행, 서버 관리 설정 가져오기, macOS MDM 도메인·Windows 레지스트리 읽기는 하지 않아요(REQ-10: 읽지 않았다고 말해요).

## Requirements

<!-- This SPEC adds no check that blocks: `moai doctor` only reports, so there is no observe stage. -->
- REQ-1: When `moai doctor` runs, it shall print one JSON answer with `operation` `doctor`, `writes` false, a `checks` list (each with `id`, `state` one of `ok`, `warn`, `fail`, `unknown`, `not_applicable`, a one-line `summary`, `evidence` and `next`) and a `pointers` list. · failure: none — a check it cannot judge is `unknown` with its reason, never absent.
- REQ-2: While `moai doctor` runs, the command shall write no file or ledger entry, open no network connection and start no process other than the `claude --version` probe of REQ-4 and, only when asked, the `claude auth status` of REQ-7. · failure: none — a test compares the trees and the stand-in `claude`'s call record.
- REQ-3 (REQ-1): IF the Claude config folder cannot be resolved by `host_home` (an empty or relative `CLAUDE_CONFIG_DIR`, or a link where a folder is read), THEN `moai doctor` shall answer the `config_folder` check `fail` naming the variable and never its value. · failure: none — every check that needs the folder is `unknown` with that reason and the other checks still run.
- REQ-4: When `moai doctor` runs, it shall run `claude --version` once with the existing five second bound, read the leading `major.minor.patch` token, and for each row of the minimum-version table (the evidence table below, `verified` rows only) answer `ok` when the installed version is at or above the row's version and `below` with the row's version and feature name otherwise. · failure: none — an unparsable output is `unknown`.
- REQ-5 (REQ-4): IF `claude` is not found or the probe times out, THEN `moai doctor` shall answer the `claude_version` check `unknown` with the reason `not_found` or `timeout` and judge no row. · failure: none — the rest of the report is unchanged.
- REQ-6: When `moai doctor` runs, it shall report the login source from names only: which of `ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN`, `CLAUDE_CODE_OAUTH_TOKEN`, `ANTHROPIC_PROFILE`, `CLAUDE_CODE_USE_BEDROCK`, `CLAUDE_CODE_USE_VERTEX` and `CLAUDE_CODE_USE_FOUNDRY` are set, whether `.credentials.json` exists in the config folder (existence only, never read), and the credential the documented precedence would pick. · failure: none — on macOS the credential may be in the Keychain, so a missing file is not a finding.
- REQ-7: When `--auth-status` is given, `moai doctor` shall run `claude auth status` once with a ten second bound and keep only `authMethod`, `configDirectory`, `loginExpiresAt` and the exit code. · failure: none — a missing or failing command is `unknown`; without the option the login expiry is `unknown` and `next` names `claude auth status --text`.
- REQ-8: When `ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN` or an `apiKeyHelper` entry of the user or project settings is present, `moai doctor` shall answer the `api_key` check `warn`, name each present variable or entry, and say in its `summary` that the key is used instead of a subscription login (always in `-p`, after one approval otherwise) and that Remote Control, `/schedule` and claude.ai connectors are off while it is set. · failure: none — only names are printed.
- REQ-9: When `moai doctor` runs, it shall read the managed `managed-settings.json` and the `managed-settings.d/*.json` files of the platform folder, the user `settings.json` of the config folder and the project `.claude/settings.json` and `.claude/settings.local.json` as files, and for each of `allowManagedHooksOnly`, `strictPluginOnlyCustomization`, `allowManagedModsOnly` (as `pluginConfigs["cc-plugin-sec-default@builtin"].options.allowManagedModsOnly`) and `disableAllHooks` answer `effective` with the source file and the surfaces it locks, `ignored` where a scope the docs say has no effect carries it, or `unset`. · failure: none — the scope rules are the documented ones, written next to each key in the evidence table.
- REQ-10 (REQ-9): IF a settings source cannot be read as a file (a file that is unreadable or not valid JSON, the macOS managed preferences domain, the Windows registry, server-managed settings, `policyHelper`, a session's `--settings`), THEN `moai doctor` shall answer that key `unknown` with the source it did not read named. · failure: none — `policyHelper` is never run and nothing is guessed.
- REQ-11: When `moai doctor` runs, it shall answer the `cross_session_inbound` check with the value the documented precedence gives (managed settings, then user settings; a project or local value only when stricter, or when no trusted source sets one) and the file that set it, and `warn` when the value is unset, `hold` or `refuse` because a lane started with `-l` or `-a` runs with bypassed permissions and holds messages then. · failure: none — a value the check cannot see (`--settings`, server-managed) is named as not read.
- REQ-12: When `moai doctor` runs, it shall print the capability row of the host it detects (`gateway` when `ANTHROPIC_BASE_URL` names a host other than `api.anthropic.com`, otherwise `cli`) and, with `--table`, the whole table for Claude login, GLM gateway, desktop, VS Code and cloud, each cell `yes`, `no`, `partial` or `unverified` with the docs page and line it came from and the read time. · failure: none — an `unverified` cell is never printed as `yes`.
- REQ-13: When `moai doctor` runs, it shall list in `pointers`, for what it does not examine, the Claude Code check that does (`claude doctor`, `/status`, `/context`, and the others the card names, each marked `verified` or `unverified` against the docs). · failure: none — it runs none of them.
- REQ-14: When `moai doctor` prints a check, it shall print variable names and states only, never a value, a key prefix, a length, the output of `apiKeyHelper` or the contents of `.credentials.json`. · failure: none — a test plants distinct values and fails if one appears.
- REQ-15: While the tests of `moai doctor` run, the tests shall use a temporary HOME and a stand-in `claude` that drives each check to pass and to fail. · failure: none — they read neither the real Claude config folder nor the real home state folder.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-14): 임시 HOME 과 가짜 `claude` 로 돌리면 `writes` 가 false 이고, 실행 전후 임시 HOME·프로젝트 트리가 같고, 심어 둔 비밀 값(`ANTHROPIC_API_KEY` 등)이 출력 어디에도 없어요 · verify: `cargo test --locked --offline --test moai_doctor_cli read_only_`
- AC-2 (REQ-3): `CLAUDE_CONFIG_DIR` 가 상대 경로나 빈 값이면 `config_folder` 가 fail 이고 변수 이름만 말하며 그 폴더가 필요한 점검은 unknown 이에요 · verify: `cargo test --locked --offline --test moai_doctor_cli config_folder_`
- AC-3 (REQ-4, REQ-5): 가짜 `claude` 가 표의 어떤 행보다 낮은 버전, 모두 이상인 버전, 알 수 없는 출력, 없음, 5초 넘는 지연을 내면 각각 `below`(행 이름과 버전), 모두 `ok`, `unknown`, `not_found`, `timeout` 이에요 · verify: `cargo test --locked --offline --test moai_doctor_cli versions_`
- AC-4 (REQ-6, REQ-7): 변수 일곱 개의 설정 여부, `.credentials.json` 의 존재(읽지 않음), 우선순위가 고르는 자격 증명이 이름으로만 나오고, `--auth-status` 없이는 만료가 `unknown` 이며 있으면 가짜 `claude auth status` 의 세 필드와 종료 코드만 남아요 · verify: `cargo test --locked --offline --test moai_doctor_cli login_`
- AC-5 (REQ-8): 키 변수나 `apiKeyHelper` 가 있으면 warn 이고 이름과 문장이 나오며, 없으면 ok 예요 · verify: `cargo test --locked --offline --test moai_doctor_cli api_key_`
- AC-6 (REQ-9, REQ-10): 네 키를 각 범위 파일에 심으면 문서의 범위 규칙대로 `effective`·`ignored`·`unset` 이 나오고(사용자·프로젝트의 `allowManagedHooksOnly` 는 ignored, `allowManagedModsOnly` 는 관리 파일의 `pluginConfigs` 아래만), 읽을 수 없는 파일은 그 소스 이름과 함께 `unknown` 이에요 · verify: `cargo test --locked --offline --test moai_doctor_cli locks_`
- AC-7 (REQ-11): 관리·사용자·프로젝트 값 조합마다 우선순위대로 유효 값과 파일이 나오고 unset·hold·refuse 면 warn, accept 면 ok 예요 · verify: `cargo test --locked --offline --test moai_doctor_cli inbound_`
- AC-8 (REQ-12): 능력 표가 Claude 로그인·GLM·데스크톱·VS Code·클라우드 다섯 열과 여덟 행을 갖고 각 칸에 문서 줄과 읽은 때가 있으며 `unverified` 칸이 `yes` 로 나오지 않아요; `ANTHROPIC_BASE_URL` 이 다른 호스트면 `gateway` 열이 나와요 · verify: `cargo test --locked --offline --test moai_doctor_cli table_`
- AC-9 (REQ-13, REQ-15): `pointers` 의 항목 각각에 `verified`·`unverified` 표시가 있고 `moai doctor` 가 그 명령을 실행하지 않아요(가짜 `claude` 의 호출 기록에 `--version` 과 허용된 `auth status` 만 있어요) · verify: `cargo test --locked --offline --test moai_doctor_cli pointers_`

## Scope

- write: `src/moai_doctor.rs` (new), `src/ops/doctor.rs` (new), `tests/moai_doctor_cli.rs` (new), this SPEC's folder
- write (shared files, by line lease): `src/lib.rs`, `src/help.rs`, `tests/areas.toml`, `tests/fixtures/help/*.txt`, `docs/design-review/appendix/operations/` one new file
- read-only: `src/host_home.rs`, `src/claude.rs` (the `claude --version` probe is reused through a small public function, no change to its behaviour)
- forbidden: any settings file, the user's Claude config folder and the home state folder (tests use a temporary HOME), the real `claude` and its credentials in tests, `src/hook*`, `src/gate*`, `tools/repo_scans.json` unless a scan test names it

## Plan

1. RED: `tests/moai_doctor_cli.rs` with the stand-in `claude` and the temporary HOME, one test group per AC.
2. `src/moai_doctor.rs`: the checks as pure functions over a read environment (variables, files, probe results), so each is unit-testable without a process.
3. `src/ops/doctor.rs` and the registration; help text and golden files; the operation document.
4. The minimum-version table and the capability table as data in `src/moai_doctor.rs`, each row with its docs page, quote id and read time.
5. Receipt, mutants for each new guard (one per REQ that refuses or hides), READY.

## Risks

- A docs cell is wrong or goes stale: each cell carries its page and read time, and `unverified` is a state of its own; a re-read on a new Claude Code version changes data, not code.
- `claude --version` and `claude auth status` start the real `claude`: only the version probe runs by default, and a stand-in replaces it in every test.
- A user-writable settings value is read as a lock: the scope rules are tested per key (AC-6); a value in a scope the docs say is ignored is reported as `ignored`, never `effective`.
- `tools/repo_scans.json` or another floor-2 path turns out to need a line when the command is registered: the tier is then 2 and the implementation slice needs this SPEC approved first (it is anyway).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `moai doctor` is one new read-only command; each check is a separate `id` so a later slice adds checks without changing the others | 카드 본문(moai 층만, 나머지는 Claude Code 점검으로 안내) | 명령을 지우면 돼요 |
| D-2 | (사용자 선택 2026-11-27, 노트 10830: 기본값 승인 — 실행하지 않음) `claude auth status` 는 `--auth-status` 옵션을 줄 때만 한 번 실행해요 | 공식 문서는 읽기 전용이라 쓰지 않았고, 변경 기록에 토큰 갱신·잠금·로그아웃 수정이 있어요(2.1.318 의 `loginExpiresAt` 도 이 명령에서만 나와요) | 기본 실행으로 바꿀 수 있어요 |
| D-3 | (사용자 선택 2026-11-27, 노트 10830: 기본값 승인) 스킬 로드 확인(`system/init`)과 상태줄 가짜 입력 실행은 이번 조각 밖(뒤 조각 후보)이고, `FUNCTION_HOOKS` 와 업데이트 명령은 출처를 읽기 전에는 넣지 않아요 | 실행과 비용이 있고, 출처가 `unverified` 예요 | 뒤 조각으로 더해요 |
| D-4 | `claude-doctor`(AGENTS.md 예측)는 그대로 두고 `moai doctor` 의 `pointers` 에서 이름만 알려요 | 두 명령의 역할이 달라요(프로젝트 파일 예측 / 환경 진단) | 합치려면 뒤 SPEC 이 필요해요 |
| D-5 | (사용자 선택 2026-11-27, 노트 10830: 기본값 승인) 최소 버전 표는 조사에서 `verified` 로 읽은 행만 판정에 쓰고, 달라진 값은 읽은 값을 써요: Windows 세션 간 메시지는 변경 기록 2.1.239(공식 가용성 문서는 2.1.234 라고 써서 둘이 달라요 — 높은 쪽을 쓰고 두 줄을 근거에 남겨요), `--discard-unpushed` 는 변경 기록에서 못 찾아 `unverified`(판정 안 함) | 조사 4번(2026-11-27, 변경 기록 2.1.332 까지) | 행 값은 데이터라 고치면 돼요 |
| D-6 | `allowManagedModsOnly` 는 최상위 키가 아니라 관리 설정의 `pluginConfigs["cc-plugin-sec-default@builtin"].options` 아래 옵션으로 읽어요 | `plugins/mods/admin.md`(조사 2번) 가 그렇게 말하고, 카드 본문의 이름과 달라요 | 문서가 바뀌면 데이터로 고쳐요 |
| D-7 | 읽을 수 없는 소스(MDM 도메인, 레지스트리, 서버 관리, `policyHelper`, `--settings`)는 추측하지 않고 `unknown` 으로 이름을 밝혀요 | 문서가 읽기 방법을 주지 않거나, 읽으려면 코드를 실행해야 해요 | 읽는 방법이 생기면 더해요 |
| D-8 | (리더 판단, 노트 10830: 한 SPEC) 요구가 15개라 `spec check` 가 12개 넘음 경고를 내요; 한 SPEC 으로 가되 점검마다 `id` 가 따로라 순서대로 구현하고, 나누려면 A(REQ-1~8: 골격·버전·로그인·API 키)와 B(REQ-9~15: 관리 잠금·수신 값·능력 표·pointers·비밀·테스트)를 `parent` 로 이어요 | 경고가 말하는 크기 기준이에요 | `parent` 로 나눠요 |

## Evidence

읽은 때 2026-11-27T12:15Z~12:18Z(UTC), 조사 에이전트 넷(`moai-research`), 출처는 `https://code.claude.com/docs/en/<page>.md`. 방법의 한계: 쪽은 WebFetch 가 요약 모델을 거쳐 준 글이고(`settings-reference.md` 는 462 KB 라 raw `curl` 로 해당 항목만 읽음), 변경 기록(1.4 MB)은 `curl`+`grep` 으로 찾은 줄만 읽었어요. `verified` 는 그 문장이 도구 출력에 있었다는 뜻이에요. 사용한 Claude Code: 지역 2.1.332, 변경 기록 최신 2.1.332(2026-11-25).

### 최소 버전 표 (moai 가 쓰는 기능)

| 기능 | 최소 | 근거 줄(변경 기록, 인용) | 상태 |
| --- | --- | --- | --- |
| `worktree.baseRef` | 2.1.133 | "Added `worktree.baseRef` setting (`fresh` \| `head`) …" | verified |
| `sessionTitle`(SessionStart 훅) | 2.1.152 | "`SessionStart` hooks can now set the session title via `hookSpecificOutput.sessionTitle` …" (UserPromptSubmit 형은 2.1.94) | verified |
| 에이전트 보기(`claude agents`) | 2.1.139 | "Added agent view (Research Preview) …" | verified |
| `claude --bg` | 2.1.140 이 첫 언급 | "Fixed `claude --bg` failing with …" ("Added" 줄을 찾지 못함) | 도입은 unverified |
| `--discard-unpushed` | — | 변경 기록에서 이름을 찾지 못함(2.1.260 은 삭제 동작만 말함) | not found |
| `notify_when_idle` | 2.1.236 | "Added `notify_when_idle` to cross-session `SendMessage` …" | verified |
| `onFailure`(command·http 훅) | 2.1.295 | "Added `onFailure: "block"` for command and HTTP hooks …" | verified |
| `--permission-prompts none` | 2.1.259 | "Added `--permission-prompts none` for unattended headless hosts …" | verified |
| `permissions.denyInsteadOfPrompting` | 2.1.319 | "Added `permissions.denyInsteadOfPrompting` setting …" | verified |
| mods | 2.1.287(데스크톱 2.1.286) | "Added Claude Mods: plugins may now modify deeper behavior" | verified |
| 플러그인 `engines` | 2.1.323 | "Added `engines["claude-code"]` to `plugin.json` …" | verified |
| 세션 간 메시지, 같은 컴퓨터 | 2.1.224(macOS·Linux·WSL), Windows 2.1.234(`cross-session-messaging.md` 가용성) / 2.1.239(변경 기록 줄) | "requires Claude Code v2.1.224 or later on macOS, Linux, and WSL 2, and v2.1.234 or later on native Windows" · "Windows: cross-session messaging is now available …" | verified, 두 값이 달라요(D-5) |
| `crossSessionInbound` | 2.1.224 | "Added `crossSessionInbound` and `dialogExpiry` settings …" (`/config` 행은 2.1.232) | verified |
| AGENTS.md 직접 읽기 | 2.1.277 | "Added AGENTS.md support …" (2.1.281 이 다른 제공자까지 넓힘) | verified |
| `claude auth status` | 2.1.41, `configDirectory` 2.1.268, `loginExpiresAt` 2.1.318 | "Added `claude auth login`, `claude auth status`, and `claude auth logout` …" 외 둘 | verified |
| `strictPluginOnlyCustomization` 의 잘못된 값 → true | 2.1.282 | `managed-settings.md` 의 fail-closed 표 | verified |
| `allowManagedHooksOnly`·`strictPluginOnlyCustomization`·`allowManagedModsOnly` 도입 | — | "Added" 줄을 찾지 못함(고침 줄만 있음) | unverified |

### 관리 잠금 (범위 규칙)

| 키 | 효과 | 값이 먹는 범위 | 상태 |
| --- | --- | --- | --- |
| `allowManagedHooksOnly` | "Restrict hook execution to hooks your organization deploys." 사용자·프로젝트·로컬 훅과 다른 플러그인의 훅·mods 가 막혀요(`settings-reference.md`) | 관리 소스만("Claude Code reads the following keys only from a managed source; placing them in user or project settings files has no effect." `managed-settings.md`) | verified |
| `strictPluginOnlyCustomization` | `true` 또는 `["skills","agents","hooks","mcp"]` 배열로 해당 표면을 막아요 | 관리 소스만; 틀린 형은 네 표면 모두 잠금(v2.1.282~) | verified |
| `allowManagedModsOnly` | 관리 설정의 `pluginConfigs["cc-plugin-sec-default@builtin"].options.allowManagedModsOnly: true` 이면 조직 mods 와 내장 mods 만 돌아요(`plugins/mods/admin.md`) | 그 자리의 관리 값만("Claude Code ignores the option in other settings files") | verified |
| `disableAllHooks` | 훅·상태줄·파일 제안 명령 끔 | 아무 파일; 관리 파일이면 관리 훅까지, 다른 파일이면 관리 훅은 계속 돌아요; 잘못된 관리 값은 경고와 함께 버려져요(`settings-reference.md`, `hooks.md`, `managed-settings.md`) | verified |

관리 설정 파일: macOS `/Library/Application Support/ClaudeCode/managed-settings.json`, Linux·WSL `/etc/claude-code/managed-settings.json`, Windows `C:\Program Files\ClaudeCode\managed-settings.json`, 그리고 같은 폴더의 `managed-settings.d/*.json`(알파벳 순 병합). MDM 도메인 `com.anthropic.claudecode`, 레지스트리 `HKLM\SOFTWARE\Policies\ClaudeCode` 의 `Settings`, 서버 관리, `policyHelper` 는 파일로 읽지 않아요(REQ-8). `CLAUDE_CONFIG_DIR` 가 관리 경로를 옮기는지: unverified(문서가 말하지 않음; 고정 시스템 경로라는 것은 추론). 유효 설정을 한 번에 출력하는 공식 명령: 찾지 못함(unverified), 그래서 doctor 가 문서의 우선순위대로 파일을 직접 합쳐요.

### 로그인·API 키

- 자격 증명 우선순위(`authentication.md`, "Authentication precedence", verified): 클라우드 제공자 변수 → `ANTHROPIC_AUTH_TOKEN` → `ANTHROPIC_API_KEY`("In interactive mode, you are prompted once to approve or decline the key … In non-interactive mode (`-p`), the key is always used when present.") → `apiKeyHelper` → `CLAUDE_CODE_OAUTH_TOKEN` → Anthropic 프로필 → `/login` 구독.
- 저장 위치(verified): macOS Keychain(실패 시 `~/.claude/.credentials.json`), Linux `~/.claude/.credentials.json`, Windows `%USERPROFILE%\.claude\.credentials.json`; `CLAUDE_CONFIG_DIR` 가 있으면 파일과 Keychain 항목 모두 그 폴더 기준이에요. Keychain 서비스 이름은 문서에 없어요(unverified), 그래서 macOS 에서 파일이 없다는 것은 "로그인 없음"의 증거가 아니에요.
- `claude auth status`(`cli-reference.md`, verified): "Exits with code 0 if logged in, 1 if not"; JSON 의 `authMethod`(`none`·`claude.ai`·`oauth_token`·`api_key`·`api_key_helper`·`third_party`), `configDirectory`, `loginExpiresAt`(2.1.318~). 읽기 전용이라는 문장은 없고, 변경 기록에 토큰 갱신·잠금·로그아웃 수정 줄이 여럿 있어요(D-2). 이 JSON 의 나머지 필드는 문서에 없어요(unverified), 그래서 세 필드만 남겨요.
- 키가 구독을 이기는 경고: "If you have an active Claude subscription but also have `ANTHROPIC_API_KEY` set … Claude Code uses the API key once you approve it."; 키가 있으면 Remote Control, `/schedule`, claude.ai 커넥터가 꺼져요(2.1.139 줄). 경고 문구와 도입 버전은 unverified.
- `claude doctor`(verified): "Print read-only installation and settings diagnostics … without starting a session"; 로그인 점검을 한다는 말은 없어요(unverified).

### 능력 표 (호스트 × 기능)

CLI = 터미널 Claude 로그인, GW = `ANTHROPIC_BASE_URL` 이 다른 호스트(GLM), 데스크톱, VS Code, 클라우드. 칸은 yes/no/partial/unverified, 줄은 문서 쪽과 제목(읽은 때 위와 같아요).

| 기능 | CLI | GW | 데스크톱 | VS Code | 클라우드 |
| --- | --- | --- | --- | --- | --- |
| Remote Control | yes | no | yes | yes | no |
| 세션 간 메시지, 같은 컴퓨터 | yes | yes(GW 줄은 unverified: "available on every provider") | partial("local, SSH, and WSL sessions"만 봐요) | yes | no(추론) |
| 세션 간 메시지, 다른 컴퓨터 | yes(Remote Control 필요) | no | unverified | unverified | partial(받기만) |
| mods | yes | yes("whichever provider serves the model") | yes(2.1.286~) | yes | partial |
| 플러그인 | yes | yes | yes | yes | partial |
| 출력 스타일 | yes | unverified | partial | yes(2.1.257~) | unverified |
| 5시간·7일 한도(상태줄) | yes(Pro·Max·Ultra) | no(제외로 읽음) | unverified | unverified | unverified |
| 아티팩트 | yes | no | yes | unverified | unverified |

쪽: `feature-availability.md`("Features that require a Claude subscription"), `remote-control.md`, `llm-gateway-connect.md`, `cross-session-messaging.md`("Availability"), `plugins/mods/overview.md`, `output-styles.md`, `statusline.md`("Rate limit usage"), `artifacts.md`(Availability), `desktop.md`, `vs-code.md`. 데스크톱 열의 일부 칸이 `unverified` 인 까닭은 그 열을 직접 가르는 줄을 읽지 못해서예요.

### `crossSessionInbound`

문자열 `"accept"`·`"hold"`·`"refuse"`("`refuse` is stricter than `hold`, and `hold` is stricter than `accept`"), 기본값 unset(메시지마다 판단), 2.1.224~. 읽는 순서(`cross-session-messaging.md`, "Control inbound messages", verified): 관리 설정 → `--settings` → 사용자 설정에서 처음 나온 값, 신뢰 소스에 없으면 프로젝트·로컬의 hold·refuse 가 적용돼요(일반 설정 우선순위와 달라요). 잘못된 값은 사용자·프로젝트에서는 hold, 관리에서는 refuse. 유효 값을 출력하는 공식 명령: 찾지 못함(unverified).

### 상태줄 입력(뒤 조각용 기록)

`statusLine` 의 `command` 는 셸에서 돌고 stdin 으로 세션 JSON 을 받아요(`model.display_name`, `workspace.current_dir`, `context_window.used_percentage`, `session_id` 등); 종료 코드 0 일 때만 출력을 쓰고, 신뢰 대화상자 전에는 비어요. 시간 상한 값은 문서에 없어요(unverified). `system/init` 의 `skills: string[]` 필드는 있지만 그 안에 무엇이 들어가는지는 문서에 없어요(unverified).
