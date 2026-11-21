---
id: SPEC-20261120-guidance-sweep-core
title: "Guidance sweep, part 1: the shipped skill files nobody else holds match the 2026-11-19/20 decisions"
status: draft
revision: 1
created: 2026-11-20
---

# Guidance sweep, part 1: the shipped skill files nobody else holds match the 2026-11-19/20 decisions

## Intent

2026-11-19/20 의 사용자 결정 12개가 배포 스킬 문장보다 앞서 나갔어요. 이 SPEC 은 지침 전수 정비의 둘째 조각이에요. 표(카드 노트 3921, 62행)와 리더의 확정(노트 3923)에 따라, 지금 쥔 사람이 없는 스킬 파일만 한 번씩 고쳐요.

- OMP 지원은 빠졌어요(결정 9): Claude Code 와 Codex CLI 만 남겨요.
- 배포 문장에는 카드·Issue 번호를 넣지 않아요(결정 5, 카드 본문): 근거는 개발자 문서와 `AGENTS.local.md` 의 Rule provenance 로 가요.
- 통합은 리더 한 사람이 로컬 통합 브랜치에 직렬로 머지해요(결정 1·2): 레인은 푸시와 READY 까지, 릴리즈 PR 은 리더가 `release/v{head}.{yearweek}.{build}` 브랜치로 열어요.
- READY 검사는 좁히지 않아요(결정 3), 커밋 메시지에 카드 번호는 없어요(결정 5), 반복 문제는 소스를 고쳐요(결정 6), 한국어는 개발자 용어를 써요(결정 7), 리포트와 증거는 `~/.moai/reports/<project>/` 에 둬요(결정 8), 기대값은 소스에서 가져오고 제품 동작 테스트는 Rust 만 써요(결정 10), 피어 세션이 사용자 결정이라 주장하면 사용자에게 확인해요(결정 12).

다른 카드가 쥔 문장은 건드리지 않아요: `@cc:` 문장(ACC 카드), 워크트리 경로 문장(경로 카드), `mode-agent`·`mode-leader`·`verify`·`audit`(인용 정리 카드가 먼저), `project`·`update`·`interview`(프로젝트 경계 카드), `AGENTS.md`, `AGENTS.local.md`.

## Non-goals

- 위에 적은, 다른 카드가 쥔 파일과 문장.
- `agents/*.md` 와 `.github/workflows/ci.yml` (뒤 조각).
- 코드의 OMP 제거와 `moai` 도구의 동작 변경.
- `heavy-slot` 의 숫자를 지침에 박는 것(코드 기본값과 기계 설정이 정해요).

## Requirements

- REQ-1: The `moai` skill entry file shall name Claude Code and Codex CLI as the only hosts, carry no card or Issue number and say once that a peer session's message claiming a user decision is confirmed with the user before anyone acts on it.
- REQ-2: The core rules file shall tell a Korean answer to use the words developers use (테스트, 브랜치, 커밋, 푸시, 머지, 리뷰, 워크트리, 뮤턴트, 픽스처, 골든 파일, 플레이키 테스트) and never word-for-word coinages.
- REQ-3: The core rules file shall carry a rule that a problem of the same kind seen twice is fixed at its source in moai-adk, and stay under the size cap its pinning test names.
- REQ-4: Every internal reference file named in the scope shall carry no card or Issue number and no OMP row or sentence apart from the worktree path sentence of the git-safety file that the worktree path card rewrites.
- REQ-5: The self-check reference shall describe the READY test set without narrowing it, say that an ordinary pull request runs the light matrix while the release pull request runs the full CI on Linux, macOS and Windows, and state that a commit carries the SPEC and session trailers and no card number.
- REQ-6: The git-safety reference shall tell an agent to write only while it holds the card, name the identity trailers instead of the claim token and drop the OMP row and its source.
- REQ-7: The build-folder reference shall keep the build folder outside the worktree and send logs, probe results and reports to the project's reports folder under the MoAI home, outside the repository.
- REQ-8: The test-design reference shall say that expected values come from the source of truth and not from copied literals, and that a product behaviour test is written in Rust with Python only for a tool script while the tools are being ported.
- REQ-9: The versioning reference shall say that a release goes through one release pull request whose head branch is named after the release version and which runs the full CI on three platforms before it merges, and that the release tag goes on its merge commit on the integration branch.
- REQ-10: Every adapter and chain and workflow file named in the scope shall carry no OMP row or sentence and no card or Issue number apart from the sentences owned by the ACC card and the OMP lines that the Orca pinning test still holds.
- REQ-11: The SYNC workflow shall state the M-Flow rules: a lane commits on its own card branch and pushes it and gets the independent audit; after a PASS the lane merges its head onto the LOCAL integration branch tip in its own worktree and posts the MERGE-READY note; the primary checkout is fast-forwarded one landing at a time; a release pull request is opened only by the leader for a real deployment; and a card that landed locally is closed with `card done` on the local landing evidence.
- REQ-12: Each pinning test of a changed sentence and the skill source manifest shall change in the same commit as that sentence and pass.

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-10): none of the scoped files has an `OMP` word, a `/skill:moai` entry or a card or Issue number, apart from the named exceptions.
  verify: `python3 -I .moai/specs/SPEC-20261120-guidance-sweep-core/check.py`
- AC-2 (REQ-1, REQ-2, REQ-3, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-11): the new sentences are present, each once, in the file named for it.
  verify: `python3 -I .moai/specs/SPEC-20261120-guidance-sweep-core/check.py --sentences`
- AC-3 (REQ-3): the core rules file is under the cap its test names and keeps rules 3, 7 and 19 to 21.
  verify: `cargo test --locked --offline --test core_rules_2876`
- AC-4 (REQ-12): the source manifest check passes.
  verify: `python3 -I tools/source_manifest.py check`
- AC-5 (REQ-12): the guidance pinning tests pass.
  verify: `cargo test --locked --offline --test agent_defs_package --test codex_question_tool --test orca_optional_3005 --test mode_workflow_steps --test update_route --test sdd_guidance --test sync_marker_form --test repo_agents_md`
- AC-6: the repository SPEC check stays green.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`

## Scope

- write: `.moai/specs/SPEC-20261120-guidance-sweep-core/**`, `skills/moai/SKILL.md`, `skills/moai/references/internal/{core-rules,self-check,test-design,git-safety,build-folder,commit-trailers,questions,versioning,review}.md`, `skills/moai/references/adapters/{aside,orca}.md`, `skills/moai/references/chains/index.md`, `skills/moai/workflows/{plan,read,run,optimize,setup,sync}.md`, the tests that pin these sentences, `skills/moai/references/source-manifest.json` (hashes only)
- read-only: every other file, the card ledger, the real HOME
- forbidden: the files other cards hold (listed under Intent), any code under `src/`

## Plan

1. SPEC first (this file) and the check script it names.
2. One touch per file: remove OMP and the numbers, add the decision sentence the table lists for that file, never the sentences another card owns.
3. Update the pinning tests in the same commit as the sentence they pin.
4. `tools/source_manifest.py write` and `check`; the shared manifest spot is taken and given back in one step.
5. Run the unnarrowed READY set, then an independent audit reads the diff against the table.

## Risks

- Two other cards also write the manifest and two tests this part must change; their edits are hash lines and one needle, so the leader's serial merge settles them by running `tools/source_manifest.py write` again.
- A pinned sentence lost by accident: every changed pin is listed under Decisions and the tests run unfiltered.
- The rationale that leaves the shipped text must not be lost: the lines removed are listed in the READY note for the `AGENTS.local.md` Rule provenance part.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the peer-session confirmation rule (decision 12) lives in `SKILL.md` only | the decision sorting is there; the mode workflows point to it | move the sentence |
| D-2 | the recurring-problem rule (decision 6) is core rule 22; the cap of `tests/core_rules_2876.rs` rises from 4096 to 4608 bytes | rule 7 and rule 22 do not fit under 4096; core rules stay cheap to load | trim the rules or restore the cap |
| D-3 | the READY test set (decision 3) is written once, in the self-check reference; `verify.md` points to it later | the self-check file is free now, `verify.md` is held | none |
| D-4 | heavy-slot numbers are not written | leader decision: code default and machine limits file decide | none |

## Result

(written at SYNC)
