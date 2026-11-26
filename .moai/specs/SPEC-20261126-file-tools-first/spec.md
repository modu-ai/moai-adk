---
id: SPEC-20261126-file-tools-first
title: "Lanes and sub-agents write tracked and in-scope files with the host file tools"
status: draft
revision: 2
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
- REQ-3: The instruction texts shall say in each of `agents/impl.md`, `skills/moai/references/internal/shell-guards.md` and `docs/lane-rules.md` that each mutant is run without a script, by changing the code with Edit, running the test with one command and changing the code back with Edit, never with a script that holds `git apply`, `git checkout` or `sed -i`.
- REQ-4 (REQ-3): IF one of those three lines is missing, THEN the repository text test shall fail and name the file that lacks it.
- REQ-5: The compound git rows of the shell guard reference and the git safety reference shall say to split such work into plain, separate commands as the refusal names, one git command per line, and never to put git into a script file.

## Acceptance

- AC-1 (REQ-1): `agents/impl.md`, `agents/docs.md`, `skills/moai/references/internal/shell-guards.md`, `docs/lane-rules.md` 에 같은 문장이 있고 테스트가 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan file_tools_first`
- AC-2 (REQ-2): 네 줄을 하나씩 지우면 같은 테스트가 매번 실패하고 그 파일 이름을 말해요. 되돌리면 다시 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan file_tools_first`
- AC-3 (REQ-1): 네 글을 읽는 이웃 테스트와 패키지 manifest 검사가 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan --test base_agent_contract --test lane_rules_2836 --test source_manifest_facts`
- AC-4 (REQ-3): `agents/impl.md`, `skills/moai/references/internal/shell-guards.md`, `docs/lane-rules.md` 에 뮤턴트를 스크립트 없이 돌린다는 같은 문장이 있고 테스트가 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan mutants_without_a_script`
- AC-5 (REQ-4): 세 줄을 하나씩 지우면 같은 테스트가 매번 실패하고 그 파일 이름을 말해요. 되돌리면 다시 통과해요 · verify: `cargo test --locked --offline --test next_flow_skill_scan mutants_without_a_script`

- AC-6 (REQ-5): `shell-guards.md` 의 그 행은 "Do what the refusal names: split the line into plain, separate commands, one `git` command per line, run from the worktree. Never put git into a script file: a script runs git where the guard cannot read it." 라고 말해요. `git-safety.md` 의 Claude Code 행 끝은 "split such work into plain, separate commands as the refusal names, one git command per line, and never put git into a script file." 이에요. 스크립트 파일을 권하는 옛 문구는 두 글에 없고, `tests/lane_rules_2836.rs` 가 새 문구와 옛 문구 없음을 단언해요 · verify: `cargo test --locked --offline --test lane_rules_2836`

## Scope

- write: `skills/moai/references/internal/git-safety.md`
- write: `tests/lane_rules_2836.rs`
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
| D-3 | 뮤턴트를 스크립트 없이 돌린다는 문장을 이 카드에 더해요(revision 2). `agents/docs.md` 는 뮤턴트를 돌리지 않아 넣지 않아요. | 리더 메시지 2026-11-26(같은 일이 하루에 두 번 났어요), 결정 노트 10314. | REQ-3, REQ-4 와 그 문장을 되돌려요. |
| D-4 | git 을 스크립트 파일에 넣으라는 두 글의 행을 고쳐요. `shell-guards.md` 의 행은 리더 결정이고, 같은 말을 하는 `git-safety.md` 의 행은 같은 원인이라 함께 넣자는 구현 레인의 제안이에요(이 revision 의 승인이 포함 여부를 정해요). | 리더 결정 노트 10428, 결정 노트 10314. | REQ-5 와 두 행을 되돌려요. |

## Result

2026-11-27, 기준 `f48ea9d815e7f11787dea9b47b5dc5fa5fb2688e`. 실행 환경은 Mac 한 대예요.

- AC-1: met — `cargo test --locked --offline --test next_flow_skill_scan file_tools_first` 1개 통과(`4342c319e7`), 빨강 `4b15648142` 에서는 exit 101 로 실패
- AC-2: met — `agents/impl.md`, `agents/docs.md`, `shell-guards.md`, `docs/lane-rules.md` 의 줄을 하나씩 Edit 로 지우고 같은 테스트를 돌리니 네 번 모두 exit 101 이고 지운 파일 이름을 말했어요. 하나씩 Edit 로 되돌린 뒤 통과
- AC-3: met — `cargo test --locked --offline --test next_flow_skill_scan --test base_agent_contract --test lane_rules_2836 --test source_manifest_facts` 가 78, 8, 2, 8개 통과(`1dce7dad56`). `--lib the_impl_agent_names_the_tools_folder_for_project_scripts` 1개 통과
- AC-4: met — `cargo test --locked --offline --test next_flow_skill_scan mutants_without_a_script` 1개 통과(`a5b36f3da6`), 빨강 `ebe00c966e` 에서는 exit 101 로 실패
- AC-5: met — `agents/impl.md`, `shell-guards.md`, `docs/lane-rules.md` 의 줄을 하나씩 Edit 로 지우니 세 번 모두 exit 101 이고 파일 이름을 말했어요. Edit 로 되돌린 뒤 통과
- AC-6: met — `cargo test --locked --offline --test lane_rules_2836` 2개 통과(`3a18b214bf`), 빨강 `f03d20b95a` 에서는 exit 101 로 실패. 두 행의 새 문구를 옛 문구로 하나씩 되돌리니 두 번 모두 exit 101 이고 파일 이름을 말했어요

- 커밋: SPEC `c1643cb82c`, 빨강 `4b15648142`, 초록 `4342c319e7`, SPEC 개정 `e843bf8060`, 빨강 `ebe00c966e`, 초록 `a5b36f3da6`, SPEC `3e76d55489`, 빨강 `f03d20b95a`, 초록 `3a18b214bf`, manifest `1dce7dad56`.
- manifest 의 sha256 두 개(`git-safety.md`, `shell-guards.md`)는 `moai check fix` 가 `card line-lease` 안에서 썼고 `shasum -a 256` 으로 두 파일과 맞는 것을 읽었어요.
- 줄 제거 확인 아홉 번 가운데 한 번(`shell-guards.md` 의 새 행)은 되돌릴 때 Edit 대신 평범한 `git checkout -- <그 파일>` 한 줄을 썼어요. 거절은 없었고 되돌린 파일의 sha256 이 manifest 값과 같았어요. 이 SPEC 이 넣은 규칙("Edit 로 되돌려요")과 어긋나서 적어요.
- 이 SPEC 은 거절 뒤 우회 예외를 넣지 않았어요(D-1). `CLAUDE.md` 는 바꾸지 않았어요.
