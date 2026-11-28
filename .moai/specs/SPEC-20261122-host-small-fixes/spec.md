---
id: SPEC-20261122-host-small-fixes
title: "Three small fixes from the Claude Code docs: claude.exe on Windows, the cloud session trailer, and the README's relation to Anthropic"
status: draft
revision: 1
created: 2026-11-22
---

# Three small fixes from the Claude Code docs: claude.exe on Windows, the cloud session trailer, and the README's relation to Anthropic

## Intent

Three facts from the official Claude Code pages reach moai: Windows runs a `.cmd` shim through `cmd.exe`, which cannot escape arguments safely; commits made in a cloud session carry a `Claude-Session:` trailer that `moai trailer-check` does not know; and a product that runs Claude Code may say so in plain text but must not suggest Anthropic built or endorses it.

## Non-goals

- The audit gate (`src/audit_gate.rs`) keeps reading authors from `Moai-Session:` only; whether it also reads `Claude-Session:` as an author is a later decision for the leader.
- No change to the trailers moai writes (`moai trailers`), and no setting of `attribution.sessionUrl`.
- No new logo, badge or image in either README; no change to LICENSE.
- No install or update of Claude Code by moai on any platform.

## Requirements

- REQ-1: When moai starts Claude Code on Windows, it shall start a native `claude.exe` and never a `.cmd` or `.bat` file.
- REQ-2 (REQ-1): IF the only Claude Code moai finds on Windows is a `.cmd` or `.bat` file, THEN moai shall refuse to start it and name the native install command from the official page.
- REQ-3: When `moai trailer-check` reads a commit whose trailer block holds a `Claude-Session:` line, the check shall add no failing finding because of that line.
- REQ-4: When a commit carries no `Moai-Session:` line and exactly one `Claude-Session:` value of the shape `https://claude.ai/code/session_<id>`, `moai trailer-check` shall report a non-failing `cloud_session` finding that names the link instead of `missing_session`.
- REQ-5 (REQ-4): IF a commit's `Claude-Session:` value has another shape or appears more than once, THEN `moai trailer-check` shall report a non-failing `malformed_cloud_session` finding beside `missing_session`.
- REQ-6: The README and its Korean translation shall each hold one section that states in plain text how MoAI ADK relates to Anthropic and Claude Code.
- REQ-7 (REQ-6): IF either README lacks that section or one of its five statements, THEN a Rust test shall fail and name the file and the missing statement.
- REQ-8 (REQ-3): IF a `Claude-Session:` line makes `moai trailer-check` report a failing finding, THEN a Rust test shall fail and name that finding.

## Acceptance

- AC-1 (REQ-1, REQ-2): on Windows, a `claude_bin` that resolves to `claude.cmd` or `claude.bat` is refused with the install command `irm https://claude.ai/install.ps1 | iex`, and a `claude.exe` is started as it is · verify: `cargo test --offline --locked --test host_small_fixes_cli windows_` (slice 2)
- AC-2 (REQ-3, REQ-8): a commit with `Moai-Spec`, `Moai-Session`, `Moai-Verified` and a `Claude-Session: https://claude.ai/code/session_<id>` line passes `trailer-check` with no new finding, and the card-number rule finds nothing in that line · verify: `cargo test --offline --locked --test host_small_fixes_cli cloud_trailer_`
- AC-3 (REQ-4): a commit with `Moai-Spec` and one well-formed `Claude-Session` line and no `Moai-Session` gives `cloud_session` (`fails: false`) and no `missing_session` · verify: `cargo test --offline --locked --test host_small_fixes_cli cloud_trailer_`
- AC-4 (REQ-5): a `Claude-Session` value of another shape and a doubled one each give `malformed_cloud_session` (`fails: false`) beside `missing_session`, and `ok` stays true on a non-SPEC branch · verify: `cargo test --offline --locked --test host_small_fixes_cli cloud_trailer_`
- AC-5 (REQ-6, REQ-7): `README.md` and `README.ko.md` each hold the section with five statements — moai runs Claude Code; it is not built, endorsed by or partnered with Anthropic; it uses no Claude Code or Anthropic name or logo as its own name or logo; it does not modify the Claude Code executable; it does not ask for, store or relay Claude account credentials or session tokens · verify: `cargo test --offline --locked --test host_small_fixes_cli readme_`

## Scope

- write (slice 1): this SPEC folder, `src/trailers.rs`, `README.md`, `README.ko.md`, `tests/host_small_fixes_cli.rs`
- write (slice 2, after the files are free): the program choice where moai starts Claude Code — `src/session.rs` (`launch`, reserved by another card) and `src/claude.rs` (`claude_version`, reserved by another card); a new module needs `src/lib.rs` (reserved by another card)

## Plan

1. Slice 1, REQ-3 to REQ-5: red tests in `tests/host_small_fixes_cli.rs` run `moai trailer-check` on commits in a temporary repository; then `log_trailers` reads `Claude-Session` as a fourth trailer field and `findings` adds the two non-failing kinds.
2. Slice 1, REQ-6 and REQ-7: a red test reads both READMEs and names each missing statement; then one section is added before "License" in each.
3. Slice 2, REQ-1 and REQ-2: one function chooses the program on Windows (a `.cmd`/`.bat` path or a bare name that resolves only to one is refused; `claude.exe` is kept) and both start places call it; tests run on every platform through a function that takes the file list, and the Windows-only start is pinned under `cfg(windows)`.

## Decisions

| Id | Decision | Source |
|---|---|---|
| D-1 | `Claude-Session` never fails a check: it is the host's trailer, not moai's, and `attribution.sessionUrl` can remove it. | cloud-environments.md, settings-reference.md |
| D-2 | A cloud commit without `Moai-Session` names its link as `cloud_session` so the reader sees who made it; the audit gate's author rule stays as it is (Non-goals). | report section 16 item 25 |
| D-3 | The README states only what the code shows: no code reads Claude sign-in files (`credentials.json` appears only in refusal and skip lists, `src/worktree_include.rs`). | `git grep` in the tree at the integration tip, 2026-11-22 |
| D-4 | Slice 2 waits for the three reserved files; the leader decides the order. | `card take` refusals, 2026-11-22 |

### Research

| Id | Fact | Read |
|---|---|---|
| R-1 | "Windows runs .bat/.cmd files via cmd.exe, which can execute commands injected through CLI arguments, and no reliable escaping for cmd.exe exists." The fix it gives: "Use a native claude executable instead: install Claude Code natively (irm https://claude.ai/install.ps1 \| iex), point ClaudeAgentOptions(cli_path=...) at a claude.exe". | https://code.claude.com/docs/en/agent-sdk/troubleshooting.md, read whole 2026-11-22 |
| R-2 | "Commits that Claude creates in a cloud session include a `Claude-Session: <url>` git trailer, and PR bodies include the session URL on its own line. To omit the trailer and the PR-body link, set `attribution.sessionUrl` to `false`." | https://code.claude.com/docs/en/cloud-environments.md, section "Link output back to the session", 2026-11-22 |
| R-3 | `attribution.sessionUrl`: "Omit the claude.ai session link from cloud and Remote Control commits". | https://code.claude.com/docs/en/settings-reference.md, settings index, 2026-11-22 |
| R-4 | "You can accurately say, in plain text, that your product has Claude Code preinstalled or that it runs Claude Code. But you can't use the Claude Code or Anthropic names or logos as part of your own product, feature, or company name, in your own logo, or in a way that suggests Anthropic built, endorses, or is partnered with your product." "The Claude Code binary must not be modified." "developers may not collect, store, or intermediate Claude.ai credentials or session tokens". | https://code.claude.com/docs/en/legal-and-compliance.md, read whole 2026-11-22 |
| R-5 | `trailer-check` reads only `Moai-Spec`, `Moai-Session` and `Moai-Verified` (`log_trailers`); a commit without `Moai-Session` gets `missing_session` with `fails: false`. moai starts Claude Code with `Command::new` on the configured `claude_bin` (default `claude`) in `src/session.rs` `launch` and with `Command::new("claude")` in `src/claude.rs` `claude_version`. | the tree at the integration tip, 2026-11-22 |

| R-6 | Slice 2 reading, 2026-11-28 (`curl -sL`, sha256 of the same fetch). STATED: native file and default place `%USERPROFILE%\.local\bin\claude.exe` (troubleshoot-install.md, "Verify your PATH"); PowerShell command `irm https://claude.ai/install.ps1 \| iex` and CMD command `curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd` (setup.md, "Install Claude Code"); the SDK's refusal text and reason: "Windows runs batch scripts by rewriting the spawn into a `cmd.exe /c` invocation, and `cmd.exe` re-parses the whole command line at execution time, so an argument value can execute injected commands", and "the SDK prefers the bundled CLI, then any native `claude.exe` it can discover, before falling back to a batch shim" (agent-sdk/troubleshooting.md, "Refusing to execute batch script"; written about the Python SDK). NOT STATED: PATHEXT on any of the four pages; any sentence that a bare `claude` resolves to `claude.exe` or `claude.cmd`; any sentence that `claude.exe` can be started directly without a shell (only "point `cli_path` at a `claude.exe`"). | agent-sdk/troubleshooting.md 02:05:51Z sha256 `a0a0d2c7b21bebc0eb4d7d47cda94dec1252f976c6b82fae8c07b0fee4746ec6` (whole); setup.md 02:05:59Z `78a1f2c0b67c30ea4ae126a6e4e84747418eeb0e409ce21079b2829d75cc5779` (whole); troubleshoot-install.md 02:06:07Z `49e94c86f96d7ea5519a4d034327ad3967c3b561526a6e2a5ed641f8e0858787` (read by grep and sed, not whole), all under https://code.claude.com/docs/en/ |

## Risks

- A cloud session id with a `c-<digit>` part would trip the card-number rule on the trailer line; AC-2 pins a real-shaped link and the rule's boundary keeps `session_…` ids out.
- The README section must stay true as the product changes; REQ-7's test holds the five statements, not their wording.

## Result

Slice 1 (REQ-3 to REQ-8). Status draft; only the lines below are claimed. Slice 2 (REQ-1, REQ-2) is described in the last paragraph of this section.

- AC-2: met. `cloud_trailer_beside_moai_trailers_adds_no_finding` passes in `tests/host_small_fixes_cli.rs`: a commit with the three moai trailers and a `Claude-Session` link gives no finding at all, the card-number rule included; `trailer-check` reads the link as its own field (`CommitTrailers::cloud_sessions`), so a `Claude-Session` line never fails.
- AC-3: met. `cloud_trailer_without_moai_session_names_the_cloud_session` passes: with no `Moai-Session` and one well-formed link, the only finding is `cloud_session` (`fails: false`) whose detail names the link, and `missing_session` is not reported.
- AC-4: met. `cloud_trailer_of_another_shape_or_doubled_is_named_without_failing` passes: a link of another host, a doubled link, the link's start with no id and an id holding a `-` each give `malformed_cloud_session` beside `missing_session`, none failing, and `ok` stays true on `main`. The two mutants that first survived (an empty id accepted, any character accepted in `is_cloud_session_link`) are killed by the last two cases (run by hand on head 0472c57f89).
- AC-5: met. `readme_states_how_moai_relates_to_anthropic` passes: `README.md` ("Relation to Anthropic and Claude Code") and `README.ko.md` ("Anthropic·Claude Code와의 관계") each hold one section with the five statements, before the license section.
- AC-1: met. The eight `windows_` tests pass on a Mac through `choose_program`, which takes the target and the file list (see the slice 2 paragraph at the end); the real start on Windows is `not_run`.

Slice 2, documentation part (the four non-blocking lines N1 to N4 of the slice 1 review, re-applied on the integration tip on 2026-11-27): both READMEs say that Claude Code connects to a Claude account or to another provider such as GLM and that a provider key you choose is kept in moai's own key vault (`moai key-set glm`, which exists in this build); the Korean text reads naturally; the `parse_log` comment names seven fields, as the format string in `src/trailers.rs` has. `cargo test --locked --offline --test host_small_fixes_cli` passes with 4 tests (the lines above stay true). The Windows part is the paragraph below.

슬라이스 2, Windows 시작(REQ-1, REQ-2; 2026-11-28, 브랜치 `spec/host-small-fixes-20261122-r3`, 시작점 NeXT 끝 `4d849223e0`): 시작하는 두 곳(`src/session.rs` 의 `launch`, `src/claude.rs` 의 `claude_version`)이 `crate::claude::start_program` 한 함수를 불러요. 고르는 일은 `choose_program(name, windows, files)` 가 해요. Windows 에서 `.cmd`·`.bat` 로 끝나는 이름(대소문자 무관, 폴더 부분 무관)은 파일 목록에 `.exe` 가 있어도 거절하고 메시지에 파일 이름과 설치 명령 `irm https://claude.ai/install.ps1 | iex` 를 넣어요. `.exe` 로 끝나는 이름은 그대로 시작해요. 맨 이름 `claude` 는 `candidates_in`(PATH 폴더마다 `.exe`, `.cmd`, `.bat` 순서로 실제 있는 파일)이 돌려준 목록에서 첫 `.exe` 를 쓰고, 셸 파일만 있으면 거절하고, 아무 파일도 없으면 이름을 그대로 두어 시작 오류가 이름을 말하게 해요. 다른 대상에서는 모든 이름이 그대로예요. 시험 아홉 개를 `tests/host_small_fixes_cli.rs` 에 썼어요(`windows_` 로 시작; 하나는 `cfg(windows)`, 하나는 `cfg(not(windows))` 라 한 플랫폼에서는 여덟 개가 돌아요). RED 는 `3e5d4510c1`(`moai tdd red`, 검사 exit 100: 여덟 개 중 네 개가 단언 실패), 구현 `8e8da5f13f`, `moai tdd green` 통과. 가드 뮤턴트 여덟 개(거절 줄, 목록의 `.exe` 우선, 셸만 있는 목록 거절, 대상 무시, `.exe`/`.cmd` 순서, 폴더를 파일로 셈, `start_program` 의 플랫폼 가드, 폴더 부분이 있는 이름)를 하나씩 넣고 `windows_` 를 돌렸어요. 다섯째(`.exe`/`.cmd` 순서)는 처음에 살아남았어요(exit 0, 한 폴더 안에 둘이 같이 있는 경우를 시험이 보지 않았어요): `moai tdd reopen` 으로 잠금을 풀고 시험의 한 폴더에 `claude.cmd` 를 더한 뒤 같은 뮤턴트가 exit 100 으로 잡혔고, 나머지 일곱은 처음부터 exit 100 이었어요. 이 문서 읽기의 결과는 R-6 이에요. 문서에 없는 것(PATHEXT, `claude.exe` 를 셸 없이 직접 시작한다는 문장)은 이 구현의 근거로 쓰지 않았어요: 확장자 순서는 moai 자신의 선택이고 `.exe` 를 먼저 보는 쪽이에요. 한계: `start_program` 은 현재 폴더(Windows 가 먼저 찾는 곳)를 읽지 않아요. `cfg(windows)` 시험과 실제 Windows 시작은 이 Mac 에서 돌리지 않았고 `not_run` 이에요.
