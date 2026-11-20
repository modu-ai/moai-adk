---
id: SPEC-20261120-security-scan-notify
title: "a security-scan hook event that notifies the model about critical and high patterns in written text, observe first"
status: draft
revision: 3
created: 2026-11-20
---

# a security-scan hook event that notifies the model about critical and high patterns in written text, observe first

## Intent

Go 세대의 보안 3층 중 L1(PostToolUse 때 방금 쓴 내용을 훑어 모델에게 알리는 층)은 10개 분류 30개 패턴이었고, 막지 않고 알리기만 했어요(`internal/hook/security/patterns.go`). Rust 훅 세트에는 이 층이 없어서 Go 를 끄면 위험한 호출 모양(`eval(`, 셸을 여는 호출, 안전하지 않은 역직렬화 등)이 조용히 지나가요. 계획서(`~/.moai/reports/moai-adk/rust-hooks-plan-2026-11-20.md` 3.2·3.4·4장)와 사용자 결정(2026-11-20, 리더 노트 3802)에 따라 이 조각(S2)은 **엔진만** 만들어요: 새 이벤트 `security-scan`, "알림" 응답, 규칙별 모드(`observe` 기본), 그리고 오탐을 미리 보는 코퍼스 보고 도구예요. 생성되는 설정에는 연결하지 않아서(연결은 S3) 기존 프로젝트의 동작은 하나도 안 바뀌어요. `tool-observe` 가 같은 방식으로 먼저 들어온 전례가 있어요.

설계 결정(계획서 3장과 리더 결정에서):
- 직접 짠 매처(결정 5). 맥락이 필요한 패턴은 다섯 개(하드코딩 비밀 대입 둘, SQL 이어 붙이기 둘, 따옴표 속 `sh -c`)라서 여섯 개 기준 안이에요.
- 알림 전용: 어떤 경우에도 `deny`/`block` 을 내지 않아요(결정 3의 정신). 모드 `observe`(기본) 는 알림 대신 `would_notify` 로그 한 줄, `enforce` 는 모델에게 알림, `off` 는 아무것도 안 해요(계획서 3.4, 결정 2).
- 이번 첫 걸음은 critical·high 분류만이에요. 아래 표의 여섯 분류예요(Go 의 medium 넷은 뒤 조각).
- 개인 키 블록과 AWS 키 ID 모양은 이미 쓰기 전에 거부돼요(S1 의 A4 와 기존 규칙), 그래서 알림 표에는 넣지 않아요. 알림 쪽의 하드코딩 비밀은 "이름 + 대입 + 따옴표 값" 맥락 패턴이에요.

| 분류 | 심각도 | 모양(줄 안, 대소문자 구분. 다르게 적은 것만 무시) |
| --- | --- | --- |
| unsafe-deserialization | high | `yaml.load`, `pickle.load`, `pickle.loads`, `torch.load`, `Marshal.load` 뒤에 공백 선택 후 `(`; `new` 공백 하나 이상 `ObjectInputStream` 공백 선택 `(`; 낱말 경계가 앞에 있는 `unserialize` 공백 선택 `(` |
| dom-injection-xss | high | `.innerHTML` 공백 선택 단일 `=`(바로 뒤가 또 `=` 인 `==`·`===` 비교는 제외, Go 와 다른 점); `dangerouslySetInnerHTML`; `document.write` 공백 선택 `(`; `v-html` 공백 선택 `=` |
| hardcoded-secret | critical | 대소문자 무시 `api` 선택 `_`/`-` `key` 또는 `password` 뒤에 공백 선택, `=`·`:=`·`<-` 중 하나, 공백 선택, 여는 따옴표(`"` 또는 `'`), 따옴표가 아닌 글자 한 개 이상, 닫는 따옴표(`"` 또는 `'`) |
| code-injection-eval | high | 낱말 경계가 앞에 있는 `eval` 공백 선택 `(`; `new` 공백 하나 이상 `Function` 공백 선택 `(` |
| sql-injection | high | (대소문자 무시) 따옴표(`"`, `'`, 백틱) 공백 선택 SELECT·INSERT·UPDATE·DELETE·DROP 낱말 경계, 따옴표가 아닌 글자 0개 이상, 따옴표, 공백 선택, `+`; 또는 (대소문자 무시) `execute`·`query` 공백 선택 `(` 뒤에 `)` 가 나오기 전에 `+` 가 있고 `)` 가 있는 모양 |
| command-injection | high | `shell` 공백 선택 `=` 공백 선택 `True`; `os.system` 공백 선택 `(`; `child_process.exec` 공백 선택 `(`; 따옴표(`"`, `'`, 백틱) 바로 뒤 `sh` 또는 `bash`, 공백 하나 이상, `-c`, 낱말 경계 |

낱말 경계는 ASCII 글자·숫자·`_` 기준이에요(Go 의 `\b` 와 같아요). 공백은 스페이스와 탭이에요.

Go 와 이 SPEC 의 차이(조사 결과, 이유가 있어요): (1) 정규식 `[^)]*` 과 `[^"'\x60]*` 은 줄을 넘어 걸리지만 이 SPEC 은 한 줄 안만 봐요. 줄을 넘는 SQL 이어 붙이기는 놓치지만, 줄 단위 선형 스캔과 오탐 관리가 단순해져요. (2) 뒤를 내다보는 거리(`(` 뒤, 키워드 뒤, 여는 따옴표 뒤)는 512 바이트로 막아요. 시간이 입력 길이에 비례하게 하려는 거예요. (3) `.innerHTML ==` 비교는 알리지 않아요(Go 는 `=` 하나만 보므로 `==` 도 걸렸어요). (4) Go 는 파일 경로와 언어별 제외가 따로 있었지만, 여기서는 쓰기 도구 입력에 담긴 텍스트만 봐요. 같은 점도 있어요: SQL 첫 모양은 따옴표 안에 따옴표가 없을 때만 걸려요(따옴표가 섞이면 둘째 모양 `query(… + …)` 이 잡아요).

## Non-goals

- Go 의 medium 분류 넷(path-traversal, ssrf, weak-crypto, insecure-random)과 L3 의 IDOR 휴리스틱: 뒤 조각이거나 퇴역이에요(계획서 3.2).
- 생성 설정에 연결(PostToolUse 항목 추가, `go-hooks-plan` 의 Go 항목 교체, 골든 파일 갱신): 조각 S3 이에요. 이 조각은 `hooks-plan`·`hook_set` 의 출력을 바꾸지 않아요.
- L2 Stop 알림, Codex 쓰기 가드 연결, ast-grep: 계획서의 S4·S5·S6 이에요.
- 디스크에서 파일을 다시 읽는 일: 쓰기 도구 입력의 텍스트만 봐요(Claude 의 `content`·`new_string`·`new_source`·`edits[].new_string`, Codex 패치의 추가된 줄). 그래서 읽기 실패나 경쟁 상태가 없어요.
- OMP: 이 이벤트는 Claude Code 와 Codex 만이에요(OMP 에는 PostToolUse 연결이 없어요).
- `src/main.rs` 의 `--event` 도움말 줄에 새 이벤트 이름을 더하는 일: 그 파일은 다른 카드가 예약 중이라서 이 조각에서는 못 고쳐요. `capabilities` 응답과 이 SPEC 이 새 이벤트를 문서화하고, 도움말 한 줄은 그 파일이 풀리는 뒤 조각에서 더해요.
- 사용자가 직접 켜는 경로 규칙(`[hooks.outside_project_writes]`)과의 합치기: 별개예요.

## Requirements

- REQ-1: WHEN a Write, Edit, MultiEdit or NotebookEdit call on Claude Code or an apply_patch call on Codex has written text that holds a shape of the six classes in the table, and the mode of the rule security_scan is enforce, security-scan shall answer with one notification in the host's documented PostToolUse additionalContext shape and exit 0.
- REQ-2: WHEN written text holds a shape of the six classes and the mode of security_scan is observe, security-scan shall print nothing, exit 0 and append one would_notify line that names the rule, the class names, the finding count, the host, the session and the call, and never a path, a command or written text.
- REQ-3: WHEN the mode of security_scan is off, security-scan shall print nothing, write nothing and exit 0.
- REQ-4: The scanner shall judge each written line on its own against exactly the shapes of the table, case-sensitive where the table says so, with a look-ahead of at most 512 bytes beyond the end of the fixed lead-in of a shape inside the line and a scan time that grows in proportion to the text length up to 1 MiB, reporting a class at most once per line.
- REQ-5 (REQ-1…REQ-3): IF the written text holds a NUL character, or is longer than 1 MiB, or every target path lies outside the project, THEN security-scan shall print nothing, write nothing and exit 0.
- REQ-6 (REQ-1…REQ-3): IF a line only names a class in prose without its code shape, such as the word eval without a call parenthesis, a hash name without a call, or a comparison of innerHTML with ==, THEN security-scan shall report nothing for that line.
- REQ-7: The notification shall open with the fixed words MoAI security-scan (advisory, nothing was blocked) and name each finding as one line of the form dash, class, severity in parentheses, relative path, colon and line number inside the written text, listing at most 5 findings within at most 800 bytes, counting the rest in a line of the form dash and N more, never repeating the matched text and writing the word (unnamed) for a path that holds a control character.
- REQ-8: IF the hook input is not a JSON object or was cut at the stdin limit, THEN security-scan shall print nothing and exit 0, the way tool-observe does.
- REQ-9: IF security-scan is asked for on OMP, or the input names a hook event other than PostToolUse, THEN the engine shall answer with its engine error and exit 1.
- REQ-10: Every existing event and every generated hook file and the output of hooks-plan shall keep its answer exactly, because security-scan is not wired into any generated setting in this change.
- REQ-11: The capabilities answer shall list security-scan for claude and codex with the keys event, tools, mode_key, default_mode, answer and blocks, and for omp as not supported.
- REQ-12: The corpus tool shall feed the tracked text files of a repository to security-scan in a temporary HOME, count findings per class and per file, and print that report without writing into the repository or the real HOME and without deciding pass or fail on a number.

## Acceptance

- AC-1 (REQ-1, REQ-7): in enforce mode every class of the table, built at run time, is notified through Write, Edit, MultiEdit and NotebookEdit on Claude and through an apply_patch added line on Codex, as the documented PostToolUse additionalContext object with the fixed words, the class, the severity, the relative path and the line, no matched text and at most 5 findings and 800 bytes · verify: `cargo test --locked --offline --test hook_security_scan_cli notify`
- AC-2 (REQ-2, REQ-3): with no entry, with an unknown word and with observe, nothing is printed and one would_notify line is written that holds no path, command or text; with off nothing is printed and no line is written; a text with no finding writes no line in any mode · verify: `cargo test --locked --offline --test hook_security_scan_cli mode`
- AC-3 (REQ-4, REQ-6): each shape of the table is found at its edge and a one-character change that leaves the shape is not, including prose that names a class, a comparison with ==, a word that merely ends in eval, a class reported once per line and the 512-byte look-ahead edge of the call, quote, keyword and blank-run shapes · verify: `cargo test --locked --offline --test hook_security_scan_cli shapes`
- AC-4 (REQ-5): a NUL character, a text of 1 MiB and one byte, and a target outside the project each give no output and no line, while a text of exactly 1 MiB is scanned · verify: `cargo test --locked --offline --test hook_security_scan_cli skips`
- AC-5 (REQ-8, REQ-9, REQ-10): malformed and cut input exit 0 with no output, OMP and a wrong event name give the engine error with exit 1, and the answers of tool-preflight, tool-observe, session-check and turn-validation are unchanged for the same inputs · verify: `cargo test --locked --offline --test hook_security_scan_cli contract`
- AC-6 (REQ-11): the hooks-plan answer for claude and codex lists security-scan with event PostToolUse, the tools, mode_key security_scan, default_mode observe, the additionalContext answer and blocks false, omp lists it as not supported, and the unit tests of the scanner pass · verify: `cargo test --locked --offline --test hook_security_scan_cli contract`
- AC-7 (REQ-4): a 1 MiB text of near-miss lines and a 1 MiB text of one long line of repeated prefixes both finish with the right verdict within a generous bound; the measured time is printed, not asserted · verify: `cargo test --locked --offline --test hook_security_scan_cli linear`
- AC-8 (REQ-12): the corpus tool prints a per-class and per-file report for a small fixture repository, writes nothing into it, and the report of this repository is kept in the Result · verify: `python3 -m unittest tests.security_scan_corpus_test`
- AC-9 (REQ-4): the added cost against a quiet PostToolUse is measured with the built binary and reported in the Result: median 10 ms or less on a small input and 50 ms or less on 1 MiB (the user's budget, not a test assertion) · verify: `python3 -I tools/hook_latency.py --help`

## Scope

- write: `src/hook_engine.rs` (the new event, the notification answer, the mode check, the capabilities text), `src/hook_engine/scan.rs` (the scanner, a child module of the engine so `src/lib.rs` stays untouched), `tests/hook_security_scan_cli.rs`, `tools/security_scan_corpus.py`, `tests/security_scan_corpus_test.py`, `.moai/specs/SPEC-20261120-security-scan-notify/**`
- read-only: `src/hook_config.rs` (`RuleMode`, `mode_of`), `src/hook_guards.rs` (`write_targets` reads the same tool inputs), `src/lib.rs`, `src/main.rs`, `tests/common/**`, `tests/edit_diagnostics_cli.rs` (the fixture model), `tools/hook_latency.py`
- forbidden: `skills/**`, `docs/hooks-generation.md`, `schemas/**`, generated hook files, other projects' hook files, the real HOME, the Go repository (read-only reference)

## Plan

Tests first (RED), then code, then mutants and neighbours, then the Result with one line per AC. The worktree is `~/.moai/worktrees/moai-adk/secscan` on a branch made from the head of the first slice (the same card, so no other card's branch is built on).

1. RED: `tests/hook_security_scan_cli.rs` uses the fixture model of `tests/edit_diagnostics_cli.rs` (a project with `.moai-cli/`, a temporary HOME with `.moai-cli/hook-config.toml`) and runs the real binary with `hook-dispatch --event security-scan`. Every code shape is assembled at run time. The engine log is read from `.moai-cli/receipts/hook-engine.jsonl`.
2. GREEN part 1: `src/hook_engine/scan.rs` holds `scan(text) -> Vec<Finding>` with a small table of line matchers (hand-written, no regex crate), the class and severity names, and the 512-character look-ahead. Unit tests in that file pin each shape's edges.
3. GREEN part 2: `Event::SecurityScan` (word `security-scan`, host event PostToolUse), the quiet skip of malformed or cut input, the OMP and event-name errors, collecting the written text per target from the tool input (the same tools `write_targets` reads), the skip rules (NUL, over 1 MiB, outside the project), the mode check (`config.mode_of("security_scan")`), the `would_notify` line, the notification text, the render arms for Claude and Codex, and the `capabilities` entry.
4. The corpus tool and its unit test, then a run over this repository whose report goes into the Result.
5. Mutants (at most 5): one per decision that a test must pin (the 1 MiB edge, the per-line class dedupe, the observe-versus-enforce branch, one shape's look-ahead bound, the `==` exclusion).
6. Latency: the new event against `tool-observe` on the same binary and inputs, reported in the Result.

Research notes: the shapes and severities are from `internal/hook/security/patterns.go` (Go repository, read 2026-11-20); the PostToolUse `additionalContext` support of both hosts and the Codex apply_patch input are from the plan's section 3.5 (official pages read 2026-11-20); the mode and log machinery is `apply_rule_mode` and `ENGINE_LOG` in `src/hook_engine.rs`.

## Risks

- Notification noise once S3 wires it in: the default is observe, and decision 2 puts a corpus report before the first enforce. The corpus report of this repository in the Result is the first data point.
- Codex PostToolUse `additionalContext` is documented but only the plan's reading of the official page backs it here; no real Codex session is run. Not wiring (S3) means no user is affected yet.
- The mode setting lives in the user's HOME and covers every project of the machine, not one project; this is the existing design of `[hooks.modes]` and is not changed.
- A hand-written matcher can drift from the Go regexes: the SPEC fixes each shape in words and every shape has an edge test; a miss costs one missing advisory, never a refusal.
- A long line of repeated prefixes could make a look-ahead quadratic: the 512-character bound and REQ-4 with its linear-time test pin that.
- Another card holds `src/main.rs` and `src/lib.rs`: this change avoids both, at the cost of the missing help line (a non-goal above).
