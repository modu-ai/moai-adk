---
id: SPEC-20261120-secret-project-keys
title: "tool-preflight refuses the project and service keys of OpenAI and Anthropic in written content"
status: draft
revision: 1
created: 2026-11-20
---

# tool-preflight refuses the project and service keys of OpenAI and Anthropic in written content

## Intent

첫 조각(SPEC-20261120-secret-token-shapes)은 OpenAI 형 키를 "글자 s, k, 하이픈" 뒤에 영문·숫자 32자 이상이 이어지는 모양으로만 막았어요. 요즘 키는 그 접두 뒤에 이름과 하이픈이 더 붙어요: OpenAI 프로젝트 키(접두 뒤 `proj-`), 서비스 계정 키(`svcacct-`), 관리 키(`admin-`)와 Anthropic 키(`ant-api03-`, 관리 키 `ant-admin01-`)예요. 이런 키는 영문·숫자 연속이 4~5자에서 하이픈으로 끊겨서 지금 규칙(A4)이 그대로 통과시켜요. 첫 조각의 독립 감사(리더 노트 3979 N1)가 이 틈을 짚었고 리더가 두 번째 조각 범위에 넣기로 정했어요.

이 SPEC 은 그 네 접두를 같은 거부 규칙에 더해요. 접두로 맞추고 길이는 보수적 하한을 써요. 근거(조사, 2026-11-20): OpenAI 공식 문서는 새 키의 길이와 문자 집합을 게시하지 않아요. 커뮤니티 글은 프로젝트 키가 접두 포함 약 130~164자이고 본문에 밑줄과 하이픈이 들어간다고 해요. GitGuardian 탐지기 문서는 2024년 9월 이후 키가 길어졌다고 적어요. Anthropic 공식 문서는 `ant-api03-` 꼴의 예시만 보여 주고 길이는 없어요. 그래서 접두 뒤 영문·숫자·하이픈·밑줄이 40자 이상 이어질 때만 거부해요. 40 은 제 선택이에요: 실제 키보다 훨씬 짧아서 놓치지 않고, 하이픈이 많은 일반 식별자의 오탐을 줄여요. 되돌리려면 표의 숫자 하나예요.

"프로젝트 키 접두"는 글자 s, k, 하이픈 다음에 아래 표의 이름이 이어진 글자열이에요(이 문서에는 비밀 모양 검사 때문에 풀어 써요).

| 이름(글자 s, k, 하이픈 다음) | 접두 뒤 허용 글자 | 접두 뒤 최소 길이 | 낱말 시작 표지 |
| --- | --- | --- | --- |
| `proj-` | 영문·숫자·`-`·`_` | 40 | 있음 |
| `svcacct-` | 영문·숫자·`-`·`_` | 40 | 있음 |
| `admin-` | 영문·숫자·`-`·`_` | 40 | 있음 |
| `ant-` | 영문·숫자·`-`·`_` | 40 | 있음 |

`ant-` 는 `ant-api03-`·`ant-admin01-` 처럼 뒤따르는 이름까지 한 접두로 묶은 거예요(접두 뒤 40자 안에 `api03-` 가 들어가요). 낱말 시작 표지는 첫 조각과 같아요: 접두 바로 앞이 ASCII 영문·숫자가 아닐 때만 접두로 세요. 거부 문구는 첫 조각의 고정 문구를 그대로 써서 값을 되풀이하지 않아요. 직접 짠 매처를 쓰는 결정(리더 노트 3802)은 그대로예요: 표 한 줄씩 더하는 변경이라 새 맥락 패턴은 없어요.

## Non-goals

- 글자 s, k, 하이픈 다음에 `None-` 이 오는 접두 같은 다른 접두: 한 커뮤니티 글에서만 봤고 확인하지 못했어요. 확인되면 표에 한 줄 더해요.
- 길이를 실제 키 길이(예: 100자 이상)로 올리는 일: 공식 문서가 길이를 게시하지 않아서 하한만 보수적으로 써요.
- 알림 전용 검사와 ast-grep: 둘째 조각(security-scan)과 뒤 조각의 일이에요.
- 셸 리다이렉션으로 쓰는 경우, 인증서 블록: 기존 A4 의 알려진 틈 그대로예요.
- 이 규칙을 끄는 스위치: A 계열 거부는 모드가 없어요(`src/hook_guards.rs` 머리말).

## Requirements

- REQ-1: WHEN a Write or Edit or MultiEdit or NotebookEdit call or a Codex apply_patch added line writes text that holds a project-key prefix of the table at a word start followed by at least 40 characters of the accepted set, tool-preflight shall refuse the call with the fixed refusal text of the access-token rule.
- REQ-2 (REQ-1): IF the run after the prefix is 39 characters or shorter or a character outside the accepted set ends it or a letter or digit stands before the prefix or the prefix is written in another letter case, THEN tool-preflight shall allow the write.
- REQ-3: The refusal text shall stay the fixed sentence of the access-token rule and never repeat the written value.
- REQ-4: Every shape of the first slice and the private-key refusal and every other tool-preflight answer shall stay as it was.
- REQ-5: The time to judge written content shall grow in proportion to its length.

## Acceptance

- AC-1 (REQ-1, REQ-3): each of the four prefixes with exactly 40 accepted characters and with hyphens and underscores inside the run and a realistic Anthropic and a realistic OpenAI project key shape is refused through Write and Edit and MultiEdit and NotebookEdit and a Codex apply_patch added line with the fixed text and the text never holds the value · verify: `cargo test --locked --offline --test hook_secret_project_keys_cli refused`
- AC-2 (REQ-2): 39 characters and a run ended by a period or a space and a letter or digit before the prefix and another letter case and the bare prefix words in prose are each allowed and the 39 and 40 edge of every prefix is pinned · verify: `cargo test --locked --offline --test hook_secret_project_keys_cli boundary`
- AC-3 (REQ-4): the first slice's tests and the existing neighbour answers pass unchanged · verify: `cargo test --locked --offline --test hook_secret_tokens_cli`
- AC-4 (REQ-5): 3 MiB of near-miss text for each new prefix is judged within a generous bound and the measured time is printed · verify: `cargo test --locked --offline --test hook_secret_project_keys_cli linear`

## Scope

- write: `src/hook_guards.rs` (four entries in `TOKEN_SHAPES` and the header comment line), `tests/hook_secret_project_keys_cli.rs`, `.moai/specs/SPEC-20261120-secret-project-keys/**`
- read-only: `src/hook_engine.rs`, `tests/hook_secret_tokens_cli.rs`, `tests/common/**`, `tools/hook_latency.py`
- forbidden: `skills/**`, `docs/hooks-generation.md`, `schemas/**`, other projects' hook files, the real HOME

## Plan

Tests first. The branch starts at the head of the first slice after its audit fix (`01d5bc0743`), in its own worktree, and goes in after the first slice.

1. RED: `tests/hook_secret_project_keys_cli.rs` follows the model of `tests/hook_secret_tokens_cli.rs` (real binary, temporary project and HOME, every secret-shaped word assembled at run time): refused through every write tool, the 39/40 edge of every prefix, characters that end the run, a word-start and case check, the fixed text, and the linear-time check.
2. GREEN: four `TokenShape` entries (prefix, 40, hyphen-underscore set, word start) and the header comment.
3. Mutants (at most 5): one minimum, one accepted set, one word-start flag, one prefix dropped, one case fold.
4. Neighbours: the first slice's tests, the whole `hook_` test set and `--lib`, as in the READY rule.

Research notes: formats from the OpenAI community threads on project key length and GitGuardian's OpenAI project key detector page, and the `ant-api03-` example of Anthropic's authentication page (all read 2026-11-20); none states a length or a character set officially.

## Risks

- A false refusal of a long hyphenated identifier that starts with one of the prefixes at a word start: the prefixes are specific and the minimum is 40, so the case is rare; a deny rule has no off switch, which is why only the four distinctive prefixes are added.
- A key shorter than 40 characters after the prefix would be allowed: real keys are far longer, and the number is one table cell.
- Vendors may change their formats again: a table line per format, as with the first slice.
