---
id: SPEC-20261121-security-scan-turn
title: "A turn-end security scan in the engine: the security-turn event reads the turn's changes, reports new critical and high findings once, and never blocks"
status: draft
revision: 2
created: 2026-11-21
---

# A turn-end security scan in the engine: the security-turn event reads the turn's changes, reports new critical and high findings once, and never blocks

## Intent

Go 세대의 보안 3층 중 L2(턴이 끝날 때 이번 턴에 바뀐 줄을 훑어 알리는 층)를 Rust 엔진에 옮겨요. L1(`security-scan`)은 쓰는 순간의 한 번의 쓰기만 보지만, L2 는 한 턴 동안 여러 번 고친 결과(파일 몇 개에 걸친 변경)와 새로 만든 파일까지 한 번에 봐요. 새 이벤트 이름은 `security-turn`, 호스트 이벤트는 Stop 이에요. 이번 조각은 **엔진만**이에요(S2 와 같은 순서): 생성 설정 연결과 capabilities 목록은 이어지는 SPEC-20261121-security-scan-turn-wire 가 맡아요. 그래서 이 조각만으로는 어떤 호스트도 이 이벤트를 부르지 않고, 호출하면 위 동작대로 답해요. Codex 의 Stop 연결은 그 다음 조각이에요.

Go 구현(`internal/hook/security/guardian.go`의 `HandleSecurityTurn`, `diff.go`의 `ScanDiff`)에서 가져오는 것과 바꾸는 것:
- 가져와요: Stop 에서 `git diff HEAD` 의 **추가된 줄만** 훑고(예전부터 있던 줄이 문맥으로 보이는 것은 다시 알리지 않아요), critical·high 만 알리고, 기본은 알리기만 하며 어떤 경우에도 차단하지 않아요. git 이 없거나 저장소가 아니면 조용히 넘어가요.
- 바꿔요 1: Go 는 `git diff HEAD` 에 잡히지 않는 **추적되지 않는 새 파일**을 놓쳐요. 사용자 결정(C-173 카드 본문)은 새 파일도 훑으라고 했으니, 무시 규칙에 걸리지 않는 새 파일의 전체 텍스트를 같이 봐요.
- 바꿔요 2: Go 는 추가된 줄을 한 덩어리로 이어 붙여 훑어서 파일을 가로지르는 가짜 일치가 생길 수 있고 어느 파일인지 알려 주지 못해요. 여기서는 파일마다 따로 훑고 알림에 프로젝트 안 상대 경로를 적어요(S2 와 같은 경로 표기 규칙).
- 바꿔요 3: Go 는 미커밋 변경이 남아 있는 동안 **매 턴마다 같은 발견을 다시 알려요.** 여기서는 같은 세션에서 같은 발견 집합을 이미 알렸으면 다시 알리지 않아요(지문으로 판단).
- 바꿔요 4: 모양 판정은 새로 만들지 않고 S2 의 스캐너(여섯 분류, 한 줄 단위)를 그대로 써요. 그래서 L1 과 L2 가 서로 어긋나지 않아요.

모드는 S2 처럼 `~/.moai-cli/hook-config.toml` 의 `[hooks.modes]` 에서 규칙 `security_turn` 으로 정해요: 기본 `observe`(로그 한 줄, 출력 없음), `enforce`(알림을 `systemMessage` 로 한 번 출력), `off`(아무것도 안 해요). 로그는 S3 가 만든 `security-scan.jsonl` 을 같이 써요(deny 줄이 쌓이는 `hook-engine.jsonl` 과 한도를 나누지 않아요).

공식 문서로 확인한 Stop 의 사실(https://code.claude.com/docs/en/hooks, 읽은 날 2026-11-21, 원문 인용은 Research notes):
- `systemMessage` 는 "Warning message shown to the user" 로 정의돼요(모델이 아니라 사용자에게 보임). 그런데 Stop 절은 `systemMessage` 를 언급하지 않아요: Stop 에서 화면에 나오는지는 **문서에 적혀 있지 않아요**(버리는 이벤트 목록에 Stop 이 있다는 말도 없어요).
- Stop 의 `hookSpecificOutput.additionalContext` 는 대화를 **이어 가게** 해요("The conversation continues so Claude can act on it"). 리더 결정은 "알리기만"이라서 이번 조각은 쓰지 않아요. 모델이 직접 받게 하는 길은 이 값이고, 그 대가가 턴 하나라는 점을 후속 결정 거리로 남겨요.
- Stop 은 matcher 를 지원하지 않고 항상 실행돼요. `stop_hook_active` 는 "Claude Code is already continuing as a result of a stop hook" 일 때 true 예요.
- Stop 의 기본 시간 제한은 문서에 적혀 있지 않아요(command 일반 기본은 600초). 생성되는 훅은 다른 훅과 같은 `HOOK_TIMEOUT_SECS` 를 쓰고, 이 이벤트는 git 호출마다 스스로 3초 제한을 둬요.

## Non-goals

- 생성 설정 연결·capabilities 목록(SPEC-20261121-security-scan-turn-wire), Codex 의 Stop 연결(계획서 S5), L3 `security-commit`(퇴역 안내만, 이미 정해진 일), ast-grep 규칙(S6), 전환 도구와 74곳(S7).
- 차단(`decision: block`)과 `additionalContext` 로 모델을 깨우는 일. 사용자 결정은 "알리기만"이에요. `MOAI_SECURITY_BLOCKING` 같은 환경변수 옵션은 옮기지 않아요.
- 모델을 불러 교차 파일 검토를 하는 에스컬레이션(Go 의 `MOAI_SECURITY_TURN_REVIEW`).
- `hooks-status` 요약에 새 로그를 넣는 일, 문서 4곳(`docs/hooks-generation.md` 등)의 갱신: 둘 다 S3 감사가 남긴 후속과 함께 다뤄요.
- S3 의 후속(N2·N3·N4·N5)과 S2 의 후속(경로 판정 함수 합치기 N6 등).

## Requirements

- REQ-1: WHEN the host claude sends the event security-turn, the duty shall read the added lines of the tracked changes against HEAD and the text of every untracked file that the project does not ignore.
- REQ-2: The duty shall scan the text of each file on its own with the scanner of security-scan and attribute each finding to that file.
- REQ-3: WHEN the duty finds a critical or high class and the rule security_turn has no entry, is observe or holds an unknown word, the duty shall append one would_notify line to the log `security-scan.jsonl` and print nothing. · failure: none
- REQ-4: WHEN the duty finds a critical or high class and the rule security_turn is enforce, the duty shall print one JSON object whose only member is systemMessage and exit 0. · failure: none
- REQ-5 (REQ-3, REQ-4): IF the rule security_turn is off, THEN the duty shall run no git command, read no file and print nothing.
- REQ-6: WHEN the input says stop_hook_active is true, the duty shall run no git command, write no line and print nothing. · failure: none
- REQ-7: WHEN the findings of the call equal the findings this duty already reported to the same session, the duty shall write no line and print nothing. · failure: none
- REQ-8 (REQ-1): IF git cannot be run, the folder is not a repository, HEAD names no commit, a git call fails or a git call runs longer than three seconds, THEN the duty shall print nothing, write no line and exit 0.
- REQ-9: The would_notify line and the notified line shall hold the rule, the class names, the count, the host, the session, the call, the time and a digest of the findings, and no path and no text.
- REQ-10: The notification shall be rendered by the renderer of security-scan, with its limit of five findings and 800 bytes and its rule for a path that holds a control character.
- REQ-11: The duty shall read the tracked changes only when the diff output is at most 4 MiB, read at most 200 untracked files and at most 1 MiB of text per file, and skip a file that is binary, not UTF-8, a link or outside the project.
- REQ-12: WHEN the event security-turn arrives for the host codex or omp, the engine shall answer with its usage error and exit 1. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-9, REQ-10): in a temporary git repository with one commit, a tracked file edited to add a line of each of the six classes, and a new untracked file with one more, the duty in `enforce` prints one object with only `systemMessage` that lists the first five findings in the order tracked files by path, then untracked files by path, each by line, as class, severity, relative path and line, and counts the sixth; the `observe` line names all six classes; a line that already stood in the file before the turn is not reported; a file that `.gitignore` names is not read; in no mode does any output hold a `decision`, `continue`, `stopReason` or `hookSpecificOutput` member · verify: `cargo test --locked --offline --test security_scan_turn_cli notify`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-9): without an entry, with `observe`, with an unknown word and with a differently cased `Enforce` the duty prints nothing and appends one `would_notify` line of rule `security_turn` (class names, count, digest, host, session, call, time; no path, no text) to `security-scan.jsonl`; with `off` it prints nothing, writes nothing and, measured by a git wrapper that records its calls, runs no git command; with `enforce` it writes a `notified` line next to the printed object · verify: `cargo test --locked --offline --test security_scan_turn_cli modes`
- AC-3 (REQ-6, REQ-7): a second Stop of the same session over the same working tree prints nothing and writes no line; a Stop of another session notifies again; a new finding after a fix notifies again; `stop_hook_active: true` prints nothing, writes nothing and runs no git command · verify: `cargo test --locked --offline --test security_scan_turn_cli once`
- AC-4 (REQ-8, REQ-11, REQ-12): a folder that is no repository, a repository without a commit, a `git` that exits non-zero, and a `git` that sleeps past three seconds each give no output, no line and exit 0 (the last one within five seconds); a binary file, a file that is not UTF-8, a symbolic link, a file of exactly 1 MiB and one of 1 MiB plus one byte, 200 and 201 untracked files, and a diff of exactly 4 MiB and one of 4 MiB plus one byte (the tracked part is then not read at all, the untracked files still are) each behave as the limits say, one boundary pair per limit; the host codex and the host omp get the usage error with exit 1 and no line · verify: `cargo test --locked --offline --test security_scan_turn_cli skips`
- AC-5 (REQ-1…REQ-12): every test file that names the list of event words of the usage error or the Stop answer passes after its expectation follows and for no other reason, and the corpus tool test still passes · verify: the test files named in the Result, run unfiltered, and `python3 -m unittest tests.security_scan_corpus_test`

## Scope

- write: `src/hook_engine.rs` (the event, its dispatch gate, its answer), `src/hook_engine/turn.rs` (new: the git reading, the limits, the fingerprint and the once-per-session check), `src/hook_engine/scan.rs` (only to expose what the new module needs, for example the renderer), `tests/security_scan_turn_cli.rs` (new), the neighbour tests whose expectation must follow the new event (named in the Result and added to the reservation one by one by `card scope`), `.moai/specs/SPEC-20261121-security-scan-turn/**`
- read-only: `src/hook_legacy.rs`, `src/claude.rs`, `src/go_hooks.rs`, `src/ops/go_hooks_plan.rs`, `src/identity.rs` (the git environment scrub), `tests/fixtures/hooks-pre-p4/**`, `docs/**`, the Go repository
- forbidden: `skills/**`, `schemas/**`, other projects' hook files, the real HOME and ledger

## Plan

Tests first (RED), then code, then neighbours and a grep of every reader of what changes, then mutants, then the Result with one line per AC. The worktree is `~/.moai/worktrees/moai-adk/secturn` on a branch made from the head of the S3 branch (S2 and S3 must land first; this slice adds to the same two files they changed).

1. RED: `tests/security_scan_turn_cli.rs` with the fixture model of `tests/hook_security_scan_cli.rs` (a temporary project that is a git repository, a temporary HOME, the real binary, secret-shaped words assembled at run time). A fake `git` earlier on PATH records its calls and can fail or sleep (the binary must still find the real `git` for the real cases: the wrapper is used only in the cases that need it).
2. GREEN: `Event::SecurityTurn` (host event Stop), the dispatch gate (Claude only), the answer arm `{"systemMessage": text}`, `turn.rs` (one `git diff HEAD -U0 --no-ext-diff --no-color --no-renames --no-textconv --relative` bounded read, one `git ls-files --others --exclude-standard -z`, per-file scan with the S2 scanner, the fingerprint, the tail check of the log). No hook set, no `hooks.rs` and no capabilities entry change in this slice.
3. Neighbours: a grep with a path argument of tests/, tools/ and docs/ for every reader of the list of event words in the usage error and of the Stop answer, then the test files it finds, run unfiltered (this was the gap of the previous slice).
4. Mutants (at most 5): the untracked files not read, the dedupe check removed, `decision: block` added to the enforce answer, the added-lines filter reading context lines, the stop_hook_active early return removed.

Research notes (read 2026-11-21): Go `internal/hook/security/guardian.go` (`HandleSecurityTurn`: `git diff HEAD`, `highSeverity`, `emitStopFinding` with `systemMessage`, fail-open), `diff.go` (`ScanDiff`: only `+` lines without the `+++` header, joined and scanned as one buffer), `internal/cli/hook.go` (`security-turn`: advisory, blocking opt-in), `internal/template/templates/.claude/hooks/moai/handle-security-turn.sh.tmpl` (thin forwarder), `.moai/specs/SPEC-SEC-GUARDIAN-001/design.md` (L2 row: Stop, 5 s, working-tree diff, `systemMessage`, ON); Rust `src/hook_engine.rs` (`Event`, `turn_validation_with` and its `stop_hook_active` rule, `head_commit` with `scrub_git_env`, `security_scan`, the per-verdict answer, `capabilities`), `src/hook_config.rs` (`mode_of` takes any rule id). Official page https://code.claude.com/docs/en/hooks, read by a fetch agent on 2026-11-21: "### JSON output" table, `systemMessage`: "Warning message shown to the user. In Agent SDK and `--output-format stream-json` output, it can arrive as an `SDKInformationalMessage`"; "## Hook input and output": "To surface a message to the user on any platform, return `systemMessage` in JSON output. Some events discard it or deliver it elsewhere, and each event's section says so."; "#### Stop decision control": "`hookSpecificOutput.additionalContext`: Non-error feedback for Claude. The conversation continues so Claude can act on it"; "#### Stop input": `stop_hook_active` "is `true` when Claude Code is already continuing as a result of a stop hook"; the matcher table lists Stop under "no matcher support | always fires on every occurrence".

## Risks

- `systemMessage` on Stop is defined by the page only in general and the Stop sections do not mention it: whether the host shows it at turn end is not verified in a real session. Until then `enforce` may have no visible effect while the log line still records it; observe, the default, does not depend on it.
- A Stop hook runs at the end of every turn: the cost is two git calls and the scan of the changed text, bounded by the limits of REQ-11 and a three-second limit per git call; a very large untracked tree is cut at 200 files and the rest is not read.
- The once-per-session check reads the tail of `security-scan.jsonl`; when that log is full no line can be written, so `enforce` notifies again on every turn until the log is trimmed (a full log is the case of the log limit of the wiring slice: observe stays silent).
- Until the wiring SPEC lands no host calls this event, so the duty is reachable only by a hand-run `hook-dispatch`; the generated hook sets and `hooks-plan` do not change in this slice.
- The tracked part compares against HEAD, so a repository without a commit is skipped quietly; staged-only new files count as tracked changes of an unborn HEAD and are not read in that case.
- The usage error text lists the known event words; tests that quote it must follow the new word and for no other reason.

## Result

All five criteria are met, on the head named in the READY note (a branch made from the head of the wiring slice, `706c14d7ba`, so the slices before it must land first). Commits on top of it: SPEC, RED tests (`Moai-Phase: red`), a test-expectation fix, then the engine code. The event is in the engine only; no generated hook set calls it yet (the wiring SPEC is the next slice).

- AC-1: met — `cargo test --locked --offline --test security_scan_turn_cli notify`: exit 0, 2 passed. In a temporary repository with an old finding in the commit, a tracked file edited to hold three classes, an untracked file with three more and an ignored file, `enforce` prints one object whose only member is `systemMessage`, listing `code-injection-eval src/a.py:2`, `dom-injection-xss src/a.py:3`, `hardcoded-secret src/a.py:4`, `unsafe-deserialization src/new.py:1`, `command-injection src/new.py:2` and "and 1 more findings not listed"; no matched text, no `ignored.py`, no `src/old.py`; the `observe` line names all six classes and counts six; a staged new file is read as a tracked change.
- AC-2: met — filter `modes`: exit 0, 2 passed. With no entry, `observe`, an unknown word and `Enforce` (capital E) nothing is printed and one `would_notify` line of rule `security_turn` lands in `security-scan.jsonl` (classes, count 6, a 16-hex digest, host, session, null call, time; none of `src/`, `pickle`, `hunter2`, `eval(`, `ignored` appears in it; the engine log stays empty); `enforce` prints the object and writes one `notified` line; `off` prints and writes nothing and the recording fake git saw no call.
- AC-3: met — filter `once`: exit 0, 2 passed. A second Stop of the same session is silent and adds no line, another session is told again, a changed set is told once and then silent, observe writes one line for two Stops; `stop_hook_active: true` prints nothing, writes nothing and the fake git saw no call.
- AC-4: met — filter `skips`: exit 0, 7 passed. No repository, no commit, a git that fails on the diff, a git that fails on the untracked list and a git that sleeps past the limit (cut within five seconds) each give no output, no line and exit 0; a binary file, a file that is not UTF-8 and two links (outside and inside the project) are skipped; 1 MiB is read and 1 MiB plus one byte is not (tracked and untracked); 200 untracked files give 200 findings and 201 give 200; a diff tuned to exactly 4 MiB gives its five tracked findings and one of 4 MiB plus one byte gives none while the untracked file still counts; the hosts codex and omp get exit 1 with "security-turn is Claude Code only".
- AC-5: met — the unknown-event answer lists the new word (`usage_error`: exit 0, 2 passed, one of them the host test above); the neighbour test files, each run unfiltered, exit 0: binary_state_cli 11, cli_name_moai 5, edit_diagnostics_cli 14 (1 ignored as before), go_hooks_migrate 17, hook_dispatch_cli 56, hook_legacy_cli 10, hook_mode_cli 5, hook_security_scan_cli 16, hook_status_cli 10, hooks_cli 62, hooks_conflict_choices_cli 24, hooks_pre_p4_fixtures 6, host_home_cli 12, lane_stop_hook_cli 5, mode_workflow_steps 43, no_moai_on_path 2, project_update_hooks_cli 23, status_line_cli 16, update_hooks_cli 11, security_scan_wire_cli 8; `--lib hook_engine` 21 (3 in the new module, 7 in the scanner module), `--lib hooks::` 16; `python3 -m unittest tests.security_scan_corpus_test`: 3 OK. `cli_name_moai` lists the Go hook names and already held `security-turn`; no expectation needed to change.
- Design choices the SPEC text did not fix (each can be reversed in a later slice): the tracked part scans the whole working-tree file and keeps the findings on added lines, so a shape that needs its neighbours still sees them and the line numbers are real; git runs with `--relative`, so a session in a subfolder judges that subfolder only; entries under `.moai-cli/` and `.git/` are never read; the digest hashes path, class and the trimmed line text, so a line that only moves is not a new finding; the once-per-session check reads the whole log (at most 4 MiB) once per turn; the notification has its own opening words (`TURN_HEADER`) over the same entries and bounds.
- Found while building: the first GREEN run failed four cases because the `systemMessage` arm sat in the Codex renderer instead of the Claude one (the engine answered "policy returned a verdict this duty cannot carry"); moved, then 14 of 14. One RED expectation sorted the class names alphabetically while the log keeps the table order of the write-time duty; the test was wrong, the lock was lifted with `tdd reopen` and a reason, and the fix is its own commit.
- Mutants (python3 of a script that edits the sources, runs the new test file and restores them; 5 of 5 killed, none a compile error, `restored: True`, rebuilt and 14 passed, `git status` empty afterwards): untracked files not read (7 tests fail), once-per-session check removed (1), `decision: block` added to the answer (4), context lines counted as added (2), a continuing turn scanned (1).
- Not run: a real Claude Code session (whether the host shows a `systemMessage` at turn end is not stated in the Stop sections of the hooks page and is unverified); the generated hook set and `hooks-plan` (the wiring slice); `docs/**` (follow-up with the earlier slices' doc items); Linux and Windows builds; a git older than the one on this machine (the calls use only `rev-parse`, `diff` with `--relative` and `ls-files --others`).
