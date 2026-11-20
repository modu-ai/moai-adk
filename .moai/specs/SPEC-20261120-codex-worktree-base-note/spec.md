---
id: SPEC-20261120-codex-worktree-base-note
title: "/moai project tells a Codex user where Codex worktrees start, and writes no Codex key"
status: draft
revision: 1
created: 2026-11-20
---

# /moai project tells a Codex user where Codex worktrees start, and writes no Codex key

## Intent

m-flow 는 새 worktree 를 통합 가지의 HEAD 에서 시작해야 해요. Claude Code 는 설정 `worktree.baseRef: head` 로 그렇게 맞추고, `/moai project` 는 m-flow 를 고르고 호스트에 claude 가 있을 때 그 설정을 미리 보여 주고 적용해요. 사용자가 Codex 를 호스트로 골랐을 때 같은 일이 가능한지 공식 문서와 소스로 조사했어요(조사 기록은 이 SPEC 의 근거 자료로 카드에 있어요).

조사 결과: Codex 에는 worktree 의 시작점을 정하는 설정이 없어요. config 키, 프로필 키, 환경 변수, 명령줄 플래그 모두 없고, 사용자 설정은 `[desktop]` 의 worktree 폴더·자동 정리·보관 개수 셋뿐이에요. Codex CLI 의 `--worktree` 는 값을 받지 않고 지금의 HEAD 에서 시작하므로 m-flow 가 바라는 동작과 같아요. 앱은 worktree 를 만들 때 작성 칸 아래에서 기반 Git 가지를 매번 골라요.

그래서 Codex 에는 쓸 것이 없어요(키를 지어내지 않아요). `/moai project` 는 m-flow 를 고르고 호스트에 codex 가 있을 때, 아무것도 쓰지 않고 이 사실을 한 문장으로 알려요. 이 SPEC 은 그 문장과 문장을 고정하는 시험 하나예요. 문장이 들어갈 파일 `skills/moai/workflows/project.md` 는 다른 카드가 쥐고 있어서, 문장은 그 카드의 다음 커밋에 들어가고 이 시험은 그 커밋과 함께 병합돼요.

## Non-goals

- Codex 설정 파일을 쓰는 plan/apply 단계. 쓸 키가 없으니 만들지 않아요; Codex 가 설정을 더하면 그때 다시 해요.
- 사용자의 실제 HOME 의 Codex 설정 읽기와 쓰기.
- Claude 쪽 `worktree.baseRef` 안내의 변경.

## Requirements

- REQ-1: When the pick is m-flow and `codex` is one of the hosts, the PROJECT workflow shall tell the user that Codex has no setting for the start point of its worktrees, that the Codex CLI's `--worktree` starts from the current HEAD, and that in the Codex app the user picks the integration branch below the composer when creating a worktree. · failure: none — the workflow only tells the user a fact and writes nothing, so no step can fail.
- REQ-2: The PROJECT workflow shall not plan or apply any Codex setting for the worktree start point.

## Acceptance

- AC-1 (REQ-1): the PROJECT workflow carries the Codex sentence of REQ-1 in the m-flow step.
  verify: `cargo test --locked --offline --test project_codex_worktree_base_note the_project_workflow_says_where_codex_worktrees_start`
- AC-2 (REQ-2): no line of the PROJECT workflow gives Codex a key of its own for the worktree start point (no Codex `baseRef`, no Codex settings plan).
  verify: `cargo test --locked --offline --test project_codex_worktree_base_note the_project_workflow_says_where_codex_worktrees_start`
- AC-3 (REQ-1, REQ-2): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261120-codex-worktree-base-note`

## Scope

- write: `.moai/specs/SPEC-20261120-codex-worktree-base-note/**`, `tests/project_codex_worktree_base_note.rs` (new).
- the sentence itself goes into `skills/moai/workflows/project.md` in the next commit of the card that holds that file; this SPEC does not write it.
- read-only: every other file, the real HOME.
- tier 1: an audit of the exact head together with the commit that adds the sentence.

## Plan

1. SPEC and the red test: the test reads the PROJECT workflow and fails until the sentence is there.
2. Hand the exact sentence to the card that holds `skills/moai/workflows/project.md`; it adds the sentence; the two heads merge in one chain.
3. READY with the pair named, then an independent audit.

## Risks

- The test is red alone: it needs the sentence from the other card. The merge chain must carry both, and the READY note names the pair.
- The wording of the sentence may be changed by the card that holds the file; the test pins three short phrases, not the whole sentence.

## Result

(written at SYNC)
