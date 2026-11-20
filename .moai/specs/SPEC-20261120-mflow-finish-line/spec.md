---
id: SPEC-20261120-mflow-finish-line
title: "M-Flow guidance: the verified local integration branch is the finish line; a remote pull request only for a formal release"
status: draft
revision: 1
created: 2026-11-20
---

# M-Flow guidance: the verified local integration branch is the finish line; a remote pull request only for a formal release

## Intent

사용자 지시(2026-11-20, 리더 대화, "m-flow 지침도 모두 변경해라. 이렇게 하기로 하고선 왜 따르지 않냐?")예요. 사용자가 정한 M-Flow 의 끝선은 "검증된 로컬 통합 브랜치"인데, 지침과 리더의 메모는 아직 "통합마다 릴리스 PR 하나"라고 말해서 리더가 통합할 때마다 원격 묶음 PR 을 열었어요. 이 SPEC 은 지침의 끝선 문장을 한 가지 글로 정하고, 지침 파일마다 한 번씩만 적고, 카드·묶음·통합마다 PR 을 열라는 문장이 어디에도 남지 않게 스캔 테스트로 고정해요.

## Non-goals

- 정식 릴리스 PR 의 절차(브랜치 이름, CI, 감사)를 바꾸는 일: 그 절차는 그대로이고, 언제 여는지만 정해요.
- 원격 PR 을 코드로 막는 가드: 이 SPEC 은 지침 글만 바꾸고, 가드는 따로 다뤄요.
- 다른 카드가 쥔 파일을 이 SPEC 이 직접 고치는 일: 그 파일들은 쥔 카드가 같은 글을 자기 변경에 넣고, 이 SPEC 의 스캔 테스트가 마지막에 확인해요.

## Requirements

- REQ-1: The `AGENTS.md` workflow shall state the M-Flow wording once, in step 6, which opens with `M-Flow (user decisions 2026-11-20; one canonical text, put once per file)`. · failure: none — a test counts that marker in the file.
- REQ-2: The M-Flow wording shall say that a lane commits on its own card branch, pushes it and gets the independent audit, and after PASS makes the merge of its head onto the current local integration branch tip in its own worktree, runs `moai check affected` there and posts `MERGE-READY` on the card. · failure: none — a test reads the two sentences.
- REQ-3: The M-Flow wording shall say that the primary checkout is fast-forwarded one at a time under a lock with `git merge --ff-only`, by the leader's supervision round where the host isolates a lane to its worktree, and that the lane redoes the merge when the base moved. · failure: none — a test reads the three sentences.
- REQ-4: The M-Flow wording shall say that only for a real deployment the leader creates a release branch `release/v{head}.{yearweek}.{build}` from the local integration branch and opens one pull request to the remote integration branch, with the full CI on Linux, macOS and Windows and an independent audit by a lane with no commits in the range, and that no pull request is opened per card, per bundle or per integration. · failure: none — a test reads the sentence.
- REQ-9: The M-Flow wording shall say that builds for local testing and installation are made from the verified local integration branch tip, that the leader checks that tip as a whole from time to time and repairs a broken tip with a revert commit, and that the statusline shows commits ahead and behind. · failure: none — a test reads the two sentences.
- REQ-10: The M-Flow wording shall say that a card that landed locally is closed with `card done` on the local landing evidence and not a pull request. · failure: none — a test reads the sentence.
- REQ-5: The `AGENTS.md` workflow shall give a lane no step that opens a pull request per Issue or per card. · failure: none — a test reads the file for the old step sentence.
- REQ-6: While `AGENTS.md` carries the finish line, it shall stay under the 32 KiB budget that Codex applies by default and keep every sentence the repository tests pin. · failure: none — `repo_agents_md` and a size test read the file.
- REQ-11: The size test of the repository's `AGENTS.md` test shall use the same 32 KiB limit, with the Codex source in a comment and the Claude Code advice of fewer than 200 lines per file as a note. · failure: none — `repo_agents_md` runs with the new limit.
- REQ-7: The guidance files that tell a lane or the leader how a change ends shall each state the finish line once, namely `AGENTS.md`, `AGENTS.local.md` and the mode-agent, mode-leader, sync and verify workflows. · failure: none — a test counts the marker in each file.
- REQ-8: The shipped guidance shall hold no instruction to open a pull request per card, per bundle or per integration under M-Flow, in `AGENTS.md`, `AGENTS.local.md`, `agents/` or `skills/`. · failure: none — a test scans those files for the old phrases.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-9, REQ-10): `AGENTS.md` 6단계에 M-Flow 표지 문장이 한 번 있고, 여섯 항목(레인 커밋·푸시·감사, 레인 병합과 MERGE-READY, 잠금 아래 ff-only, 로컬 빌드·설치·tip 검증, 실제 배포 때만 릴리스 PR, 로컬 착지로 card done)의 문장이 모두 있어요.
  verify: `cargo test --locked --offline --test mflow_finish_line agents_md_states_the_finish_line_once`
- AC-2 (REQ-5): `AGENTS.md` 에 이슈마다 PR 을 연다는 옛 단계 문장이 없어요.
  verify: `cargo test --locked --offline --test mflow_finish_line agents_md_has_no_pull_request_per_issue_step`
- AC-3 (REQ-6, REQ-11): `AGENTS.md` 가 32 KiB 보다 작고 기존 시험이 고정한 문장이 그대로 있어요.
  verify: `cargo test --locked --offline --test mflow_finish_line agents_md_stays_inside_its_budget`; `cargo test --locked --offline --test repo_agents_md`
- AC-4 (REQ-7): 이름이 붙은 여섯 파일이 끝선 표지 문장을 한 번씩만 가져요(다른 카드가 쥔 파일은 그 카드가 들어온 뒤 통과해요).
  verify: `cargo test --locked --offline --test mflow_finish_line every_named_file_states_the_finish_line_once`
- AC-5 (REQ-8): 지침 파일 어디에도 묶음 PR·통합마다 PR 을 열라는 문장이 없어요.
  verify: `cargo test --locked --offline --test mflow_finish_line no_guidance_opens_a_pull_request_per_bundle`

## Scope

- write: `.moai/specs/SPEC-20261120-mflow-finish-line/spec.md`, `tests/mflow_finish_line.rs`, `AGENTS.md`, `AGENTS.local.md`, and one constant in `tests/repo_agents_md.rs` (the size limit)
- read-only: the workflow files held by other cards
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기

## Plan

1. 시험 먼저: `tests/mflow_finish_line.rs` 시험 다섯 개를 쓰고 실패(RED)를 읽어요.
2. `AGENTS.md` 의 6단계(이슈마다 PR)를 끝선 문장으로 바꾸고, 7단계는 정식 릴리스 PR 의 병합 조건으로 두어요. 크기 상한은 Codex 기본 32 KiB(출처 주석)로 올려요.
3. 다른 카드가 쥔 파일은 리더가 보낸 같은 규칙 글을 쥔 카드가 넣어요. 그 머리가 들어온 뒤 마지막 스캔 테스트를 돌려요.
4. `AGENTS.local.md` 는 쥔 카드가 통합된 뒤에 같은 글로 바꿔요.

## Risks

- `AGENTS.md` 여유가 40 바이트라 글을 더하면 9 KiB 시험이 깨져요 → 사용자 결정(2026-11-20 10:52Z)으로 상한을 Codex 의 `project_doc_max_bytes` 기본값 32 KiB 로 올려요. Claude Code 문서의 파일당 200줄 이하 권장은 강제가 아니라서 시험 주석과 실패 메시지에만 둬요.
- 지침 파일 여럿이 같은 글을 넣으면 표현이 어긋나요 → 표지 문장 한 줄과 핵심 문장 셋을 시험이 글자 그대로 고정해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | M-Flow 글은 리더의 최종 영어 문구(2026-11-20 10:45Z, 앞서 준 두 문구를 바꾼 것) 여섯 항목을 그대로 쓰고, 한 줄 400자 제한 때문에 항목 3·4 의 이어지는 문장만 다음 줄로 내려요 | 사용자 지시 2026-11-20 '지침을 모두 업데이트'; 파일마다 같은 글이어야 스캔이 쉬워요 | 글을 지워요 |
| D-2 | 이슈마다 PR 을 여는 6단계는 끝선 문장으로 바꾸고 7단계의 PR 병합 문장은 정식 릴리스 PR 에만 걸어요 | 7단계 문장은 `repo_agents_md` 가 고정해서 글자는 그대로 둬요 | 옛 6단계 문장을 되돌려요 |
| D-3 | 스캔 테스트는 이름 붙은 파일 모두를 보고, 남의 파일은 그 카드가 들어온 뒤에 통과해요 | 카드가 쥔 파일은 이 카드가 쓸 수 없어요 | 시험의 파일 목록을 줄여요 |

## Result

(READY 전에 채워요.)
