---
id: SPEC-20261120-secret-token-shapes
title: "tool-preflight refuses six more access-token shapes in written content"
status: draft
revision: 3
created: 2026-11-20
---

# tool-preflight refuses six more access-token shapes in written content

## Intent

Go 세대의 PreToolUse 는 쓰기 내용에서 비밀 모양 9가지를 막았어요(`internal/hook/pre_tool.go:263-273`). Rust `tool-preflight` 의 A4 는 개인 키 블록과 GitHub 토큰 두 형태만 막아서(`src/hook_guards.rs` 의 `carries_secret`), Go 를 끄면 나머지 접근 토큰이 그대로 파일에 써져요. 실행으로도 확인했어요: 빈 임시 폴더에서 `hook-dispatch --host claude --event tool-preflight` 에 AWS 키 ID 형과 OpenAI 형 키를 쓰는 입력을 보내면 둘 다 조용히 통과해요(보안 가드 대조표의 프로브 P6·P7, 2026-11-20).

이 SPEC 은 접두가 뚜렷한 여섯 가지(OpenAI 형 키, GitHub OAuth 토큰, GitLab 토큰, Slack 토큰, AWS 접근 키 ID, Google OAuth 토큰)를 같은 거부 규칙(A4)에 더해요. 사용자 결정(2026-11-20, 리더 노트 3802): 보안 가드 이식은 조각 순서의 첫 조각이고, 직접 짠 매처로 가요(맥락 패턴이 6개를 넘으면 `regex` 크레이트 도입을 다시 결정).

### Go 와 이 SPEC 의 차이 (조사 결과, 모양마다 이유가 있어요)

| 모양 | Go 정규식 | 이 SPEC | 이유 |
| --- | --- | --- | --- |
| OpenAI 형 | 접두(글자 s, k, 하이픈) 뒤 영문·숫자 32자 이상 (낱말 어디서나) | 같은 길이·글자, 낱말 시작에서만 | `task-based` 같은 평범한 낱말 오탐을 막아요(기존 규칙 `secret_shaped` 와 같은 낱말 시작 정의) |
| GitHub OAuth | 접두(글자 g, h, o, 밑줄) 뒤 영문·숫자 36자 | 같음 | 기존 GitHub 일반 토큰 규칙(`has_token`)과 같은 방식 |
| GitLab | `glpat-[a-zA-Z0-9\-]{20}` | 글자에 밑줄 추가, 20자 이상 | 실제 토큰에 밑줄이 있어요. Go 의 글자 집합보다 넓어서 거부가 조금 더 많아져요 |
| Slack | `xox[baprs]-[a-zA-Z0-9\-]+` (길이 제한 없음) | 접두 뒤 10자 이상, 낱말 시작 | Go 는 `xoxb-a` 같은 산문도 거부해요. 기존 `redact_outgoing` 의 15자 규칙과 맞춰요 |
| AWS 키 ID | `AKIA[0-9A-Z]{16}` | `AKIA` 와 임시 키 `ASIA`, 16자 이상 대문자·숫자, 낱말 시작 | 임시 키도 접근 키 ID 이고 `redact_outgoing` 이 이미 둘을 같이 봐요 |
| Google OAuth | `ya29\.[a-zA-Z0-9_\-]+` (길이 제한 없음) | 접두 뒤 20자 이상, 낱말 시작 | 접두만 있는 산문을 막아요 |
| 인증서 블록 | 거부 | **허용 그대로** | 공개물이라 일부러 뺀 이전 결정(#2391 정정 2번) |

낱말 시작의 정의: 내용의 맨 앞이거나, 바로 앞 글자가 ASCII 영문자나 숫자가 아닌 자리(`src/lib.rs` 의 `secret_shaped` 와 같은 규칙).

기준선(측정, 2026-11-20T01:05Z, 설치된 `moai-cli 4.2647.9 (rust)`, 빈 임시 폴더): 작은 쓰기 입력 `tool-preflight` 중앙값 50 ms, 1 MiB 평문 한 번에 116 ms. 사용자가 정한 성능 예산(리더 노트 3802): 새 검사의 추가 비용은 중앙값 10 ms 이하, 최악 50 ms.

## Non-goals

- 인증서 블록과 확장자(`.pem`, `.key`)만으로 막는 규칙: 지금 차이 그대로예요.
- `secrets.json` 류 파일 이름 규칙과 확인(ask) 등급 파일 패턴: 지금 차이 그대로예요.
- 접두 뒤에 하이픈이 더 오는 새 OpenAI 프로젝트 키 형태: Go 와 같은 범위만 가져와요. 넓히는 건 오탐 시험과 함께 따로 해요.
- 알림 전용 검사(L1·L2)와 ast-grep 규칙: 같은 이식의 뒤 조각이에요.
- 셸 리다이렉션으로 쓰는 경우: 기존 A4 의 알려진 틈 그대로예요.
- Codex 생성 설정에 쓰기 가드를 연결하는 일: 지금 생성되는 Codex 훅 파일은 이 가드를 부르지 않아요(`src/hooks.rs:45`, `:287-300`). 연결은 계획서의 조각 5 예요. 이 조각의 Codex 시험은 `--host codex` 로 엔진을 직접 불러요.
- 이 규칙을 끄는 스위치: A 계열 거부는 모드가 없어요(`src/hook_guards.rs` 머리말). 접두가 뚜렷한 토큰만 거부로 올리는 이유예요.

## Requirements

낱말 시작은 위 정의를 써요. 모든 모양은 대소문자를 구분해요.

- REQ-1: WHEN a Write, Edit, MultiEdit or NotebookEdit input, or a line added by a Codex apply_patch, carries an OpenAI-style key — the three-character prefix of the letters s and k and a hyphen, at a word start, followed by 32 or more ASCII letters or digits, tool-preflight shall refuse the call with the secret-content reason.
- REQ-2: WHEN the written content carries a GitHub OAuth token — the prefix of the letters g, h and o and an underscore, anywhere, followed by 36 or more ASCII letters or digits, tool-preflight shall refuse the call with the secret-content reason.
- REQ-3: WHEN the written content carries a GitLab personal access token — the prefix `glpat-` anywhere followed by 20 or more ASCII letters, digits, hyphens or underscores, tool-preflight shall refuse the call with the secret-content reason.
- REQ-4: WHEN the written content carries a Slack token — at a word start the prefix `xox`, one of the letters b, a, p, r or s, a hyphen, then 10 or more ASCII letters, digits or hyphens, tool-preflight shall refuse the call with the secret-content reason.
- REQ-5: WHEN the written content carries an AWS access key id — at a word start the prefix `AKIA` or `ASIA` followed by 16 or more upper-case ASCII letters or digits, tool-preflight shall refuse the call with the secret-content reason.
- REQ-6: WHEN the written content carries a Google OAuth token — at a word start the prefix `ya29.` followed by 20 or more ASCII letters, digits, hyphens or underscores, tool-preflight shall refuse the call with the secret-content reason.
- REQ-7 (REQ-1…REQ-6): IF a candidate value is one character shorter than the minimum of its shape, or its prefix stands inside a longer word for a shape that needs a word start, or its letters differ in case, or the prefix is only named in prose, THEN tool-preflight shall allow the call.
- REQ-8: IF the written content is a certificate block or a public key block, THEN tool-preflight shall allow the call as it does today.
- REQ-9 (REQ-1…REQ-6): The refusal reason shall name private keys and access tokens in fixed words that never echo the written content or the matched value.
- REQ-10: The existing refusals (private-key block; the two GitHub token forms; protected paths; A5; A6) and their allowed look-alikes shall keep their answers, a Codex apply_patch being still scanned on its added lines only.
- REQ-11: WHILE tool-preflight scans written content of any size up to the stdin limit, the scan time shall grow in proportion to the content length: a candidate run is counted only up to its minimum, and every shape is found in one forward pass per prefix.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6): each of the six shapes (AWS in both prefixes), built at run time at exactly its minimum length, is refused through `hook-dispatch --event tool-preflight` as Write content, Edit new_string, MultiEdit edits new_string and NotebookEdit new_source on `--host claude`, and as the added line of an apply_patch on `--host codex` · verify: `cargo test --locked --offline --test hook_secret_tokens_cli refused`
- AC-2 (REQ-7): for each shape the value one character shorter than its minimum is allowed, and so are a prefix named in prose, `task-based` and `risk-free` style words for the OpenAI-style shape, a prefix inside a longer word for every word-start shape, and the lower-case look-alike of `AKIA` and `ASIA` · verify: `cargo test --locked --offline --test hook_secret_tokens_cli boundary`
- AC-3 (REQ-8, REQ-10): a certificate block, a public key block and the existing look-alikes are allowed, the existing refusals answer as before, and a token on a removed or context line of an apply_patch is allowed · verify: `cargo test --locked --offline --test hook_secret_tokens_cli unchanged`
- AC-4 (REQ-9): the refusal output of every shape is the documented deny object whose reason is the fixed secret-content text, and the output holds none of the matched value · verify: `cargo test --locked --offline --test hook_secret_tokens_cli reason`
- AC-5 (REQ-11): a 4 MiB content made only of a prefix followed by one character short of its minimum, repeated, and a 1 MiB plain text both finish and give the right verdict within the test's generous bound; the measured time is printed, not asserted · verify: `cargo test --locked --offline --test hook_secret_tokens_cli linear`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): the unit tests of the matcher pass, and the text of the two `capabilities` guard descriptions says access-token content · verify: `cargo test --locked --offline --lib hook_guards::`
- AC-7 (REQ-11): the added cost against the baseline is measured with `tools/hook_latency.py` and reported in the Result: median 10 ms or less on a small input, 50 ms or less on 1 MiB (the user's budget, not a test assertion) · verify: `python3 tools/hook_latency.py --help`

## Scope

- write: `src/hook_guards.rs`, `src/hook_engine.rs` (the refusal wording is a constant in `hook_guards.rs`; the two guard descriptions in `capabilities` are the only lines of `hook_engine.rs`), `tests/hook_secret_tokens_cli.rs`, `.moai/specs/SPEC-20261120-secret-token-shapes/spec.md`
- read-only: `src/redact.rs`, `src/lib.rs`, `tests/common/**`, `tests/green_test_lock.rs`, `tools/hook_latency.py`
- forbidden: `skills/**`, `docs/hooks-generation.md`, `schemas/**`, other projects' hook files, the real HOME
- Pinned wording (checked with `git grep` on the base commit, 2026-11-20): the refusal text is used only through the constant `SECRET_CONTENT` (6 uses in `src/hook_guards.rs`) and no test or document repeats its words; the two `capabilities` strings are at `src/hook_engine.rs:435` and `:461`.

## Plan

Tests first (RED), then code, then mutants and neighbours, then the Result with one line per AC. The code step waits for the two source files: another lane's reservation holds `src/hook_guards.rs` and `src/hook_engine.rs` (the leader asked for their return, note 3804); the SPEC and the RED test file do not touch them.

1. RED: `tests/hook_secret_tokens_cli.rs` runs the real binary in a temporary folder with a temporary HOME (the `preflight` helper of `tests/green_test_lock.rs` is the model). Every fixture is assembled at run time with `format!`/`concat!`, so no source line holds a whole token. The six shapes at their minimum length fail first.
2. GREEN: `carries_secret` gains a small shape table (prefix, minimum length, allowed characters, word-start flag, case-sensitive) that `has_token` already half expresses; `has_token` counts a run only up to its minimum (`take(min)`), which also keeps the scan linear. The refusal text and the module header say access token, not only GitHub token.
3. `capabilities` in `src/hook_engine.rs` says "access-token content" in its two descriptions (Claude write tools, Codex apply_patch).
4. Mutants (at most five): one change per shape's minimum length, caught by AC-1 or AC-2.

Research notes: the Go table (`pre_tool.go:263-273`) is the source of the six shapes and their lengths and the first column of the difference table above; the Slack and AWS rules follow `redact_outgoing` in `src/redact.rs`; the word-start rule follows `secret_shaped` in `src/lib.rs`. Measured while writing this SPEC: a long run of repeated prefixes is refused at the first match, and a failing candidate is always shorter than its minimum, so the existing scan is not quadratic; REQ-11 pins that property for the new shapes.

## Risks

- A short vendor prefix inside a longer word could refuse ordinary text: the short-prefix shapes count only at a word start, and AC-2 pins `task-based`/`risk-free` style words.
- A real test fixture in a user's repository that carries a well-formed fake token is refused: the same cost the existing GitHub token rule has. The reason names private keys and access tokens, so the writer can change the fixture.
- No off switch exists for A4 (REQ scope note). If a shape turns out to refuse real work, the way back is a code change in the next pull request, which is why only prefixes with a minimum length are added.
- GitLab accepts underscores where Go does not, so it may refuse slightly more than Go did; AC-2 pins the one-short boundary.
- Another lane holds `src/hook_guards.rs` and `src/hook_engine.rs`: the code step waits for the leader's answer; nothing in those files is edited before the reservation is mine.
