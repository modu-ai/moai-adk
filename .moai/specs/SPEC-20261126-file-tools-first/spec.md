---
id: SPEC-20261126-file-tools-first
title: "Lanes and sub-agents write tracked and in-scope files with the host file tools"
status: draft
revision: 1
created: 2026-11-26
---

# Lanes and sub-agents write tracked and in-scope files with the host file tools

## Intent

레인과 하위 에이전트가 소스와 문서를 셸(heredoc, `sed -i`, 리다이렉트)로 쓰면 호스트의 워크트리 격리 가드가 그 명령을 거절하고, 레인은 거절된 일을 다른 길로 다시 하지 않아서 멈춰요. 네 지침 글에 파일은 처음부터 호스트의 파일 도구로 쓴다는 한 줄을 넣고, 그 줄이 사라지면 저장소 글 테스트가 실패하게 해요.

## Non-goals

- 거절된 셸 쓰기를 파일 도구로 바꿔 다시 하는 예외는 넣지 않아요. 거절된 일을 다른 길로 다시 하지 않는 규칙은 그대로예요.
- `CLAUDE.md`, 훅, 권한 설정, 호스트 가드는 바꾸지 않아요.
- 셸에 거는 검사나 차단을 새로 만들지 않아요.

## Requirements

- REQ-1: The instruction texts shall say in each of `agents/impl.md`, `agents/docs.md`, `skills/moai/references/internal/shell-guards.md` and `docs/lane-rules.md` that tracked files and in-scope files are written with the host's file tools (Write, Edit) from the first write, never with a shell heredoc, `sed -i` or a redirect.
- REQ-2 (REQ-1): IF one of those four lines is missing, THEN the repository text test shall fail and name the file that lacks it.

## Acceptance

- AC-1 (REQ-1): `agents/impl.md`, `agents/docs.md`, `skills/moai/references/internal/shell-guards.md`, `docs/lane-rules.md` 에 같은 문장이 있고 테스트가 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan file_tools_first`
- AC-2 (REQ-2): 네 줄을 하나씩 지우면 같은 테스트가 매번 실패하고 그 파일 이름을 말해요. 되돌리면 다시 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan file_tools_first`
- AC-3 (REQ-1): 네 글을 읽는 이웃 테스트와 패키지 manifest 검사가 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan --test base_agent_contract --test lane_rules_2836 --test source_manifest_facts`

## Scope

- write: `agents/impl.md`
- write: `agents/docs.md`
- write: `skills/moai/references/internal/shell-guards.md`
- write: `docs/lane-rules.md`
- write: `tests/next_flow_skill_scan.rs`
- write (공유 파일, `card line-lease` 로 잠깐 빌려 씀): `skills/moai/references/source-manifest.json`
- forbidden: `CLAUDE.md`

## Plan

1. 테스트를 먼저 더해 지금 글에서 실패하는 것을 확인해요.
2. 네 글에 같은 문장을 한 줄씩 넣어요.
3. 바뀐 `shell-guards.md` 의 sha256 을 manifest 에 맞춰요.
4. 네 줄을 하나씩 지워 테스트가 실패하는지 확인하고 되돌려요.
5. `cargo fmt --all`, `moai check affected --card <base>..<head>`, 제출.

## Risks

- 새 문장이 거절 뒤 우회를 허락하는 말로 읽히면 안 돼요. "처음부터"라는 말을 넣고, `shell-guards.md` 의 거절 규칙 줄은 바꾸지 않아요.
- 셸에서 `moai card note` 에 heredoc 으로 본문을 넘기는 것은 파일 쓰기가 아니에요. 문장은 트래킹되는 파일과 범위 안 파일만 말해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 네 글에는 처음부터 파일 도구로 쓴다는 문장만 넣고, 거절 뒤 예외는 넣지 않아요. | 리더 결정 노트 10294. 거절 규칙과 부딪히고, 사용자의 허용(노트 10287)은 그 한 건에 대한 것이었어요. | 문장 변경을 되돌려요. |
| D-2 | 네 줄의 고정 테스트는 이미 그 글들을 읽는 `tests/next_flow_skill_scan.rs` 에 하나로 둬요. | 새 테스트 파일이나 검사 표 변경이 필요 없어요. | 테스트를 옮겨요. |
