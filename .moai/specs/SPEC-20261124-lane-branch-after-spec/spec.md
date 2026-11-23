---
id: SPEC-20261124-lane-branch-after-spec
title: "레인 브랜치 이름이 SPEC 슬러그와 달라도, SPEC 이 그 브랜치를 가리키면 코드 쓰기가 열려요"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7875 (2026-11-23T20:46:12Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7874 에 대한 답): (A) 판 1 승인, 길 (a) 앞머리 branch: 줄, D-1~D-4, REQ-1~8, AC-1~8, 조건 둘(훅 tier 2 감사와 guard-mutant 줄, 스캔은 한 단계만 읽고 링크를 따라가지 않는 테스트 하나)을 AC-9 와 D-5 로 옮겼어요; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요"
  digest: "sha256:0570b6475097f9b35980b879a7aaeebe9885924c488506fc0d61d7b798b4157b"
---

# 레인 브랜치 이름이 SPEC 슬러그와 달라도, SPEC 이 그 브랜치를 가리키면 코드 쓰기가 열려요

## Intent

재발 방지예요. 리더는 레인이 SPEC 이름을 고르기 전에 워크트리와 브랜치 `spec/<slug>-<yyyymmdd>` 를 만들어요. SPEC-first 훅은 그 브랜치 이름에서 `SPEC-<yyyymmdd>-<slug>` 를 읽고 그 폴더의 `spec.md` 가 있어야만 `.moai/specs/` 밖의 쓰기를 열어요. 그래서 레인이 다른 슬러그나 다른 날짜로 SPEC 을 쓰면 코드 쓰기가 막히고, 레인은 브랜치 이름을 바꿔야 했어요. 같은 일이 두 번 있었어요: 한 번은 날짜가 달라서(기준 없는 카드를 건너뛰는 SPEC 을 다시 승인받음), 한 번은 슬러그가 달라서(`moai-leader-route` SPEC 이 `spec/next-flow-skill-set-20261124` 브랜치 위에서 막힘).

고치는 곳은 훅이 브랜치를 읽는 단 하나의 해석기 `spec_branch::resolve` 예요. SPEC 앞머리에 `branch: <브랜치 이름>` 줄을 두면, 이름으로는 SPEC 을 못 찾는 브랜치에서 그 SPEC 을 브랜치의 SPEC 으로 읽어요. 코드는 여전히 SPEC 이 있어야만 열려요: 규칙은 약해지지 않아요.

읽는 쪽은 SPEC 을 쓰는 레인, 레인 브랜치를 이름 짓는 리더, 이 변경을 감사하는 세션이에요.

## Interview

- Q1: Which of the card's two fixes, (a) a `branch:` link read by the hook or (b) a dispatch helper that names the branch after the SPEC? -> A: (a) (D-1). It is data inside the SPEC, so it needs no new command and no reservation rewrite, and one resolver gives it to the hook, `moai tdd` and the Stop check together.
- Q2: When does the resolver look at `branch:` lines? -> A: Only when the name-based reading does not end in a SPEC file that exists (D-2), so the common path reads no extra file and a branch whose own SPEC exists is judged exactly as before.
- Q3: Which SPECs count as a link? -> A: One whose folder name equals its `id:`, whose status is `draft` or `approved`, whose folder is a real folder and whose `branch:` value equals the branch name exactly (D-3).
- Q4: Does `moai spec new` write the line? -> A: No (D-4). A lane writes the line by hand when its SPEC id differs from the branch; a command for it is a follow-up if the hand step fails a third time.

## Non-goals

- Weakening the rule that code waits for a SPEC: a branch with no SPEC and no link is refused as today.
- A dispatch helper that renames branches or rewrites the card reservation (the card's option (b)).
- A new `moai spec new --branch` option.
- Reading `branch:` anywhere outside `spec_branch::resolve`.
- Changing `CLAUDE.md` or the SPEC approval rules.

## Requirements

- REQ-1: When a SPEC under `.moai/specs/` has a `branch:` frontmatter line equal to the current branch and the branch name alone reads no existing SPEC, the SPEC-first hook shall read that SPEC as the branch's SPEC and allow writes outside `.moai/specs/`.
- REQ-2 (REQ-1): IF no SPEC reads for the branch, neither by its name nor by a `branch:` line, THEN the SPEC-first hook shall refuse writes outside `.moai/specs/` as it does today.
- REQ-3 (REQ-1): IF a `branch:` line sits in a SPEC whose folder name differs from its `id:`, whose status is neither `draft` nor `approved` or whose folder is a link, THEN the resolver shall not read that SPEC as the branch's SPEC.
- REQ-4 (REQ-1): IF two SPECs have a `branch:` line equal to the branch, THEN the SPEC-first hook shall refuse writes outside `.moai/specs/` with the two-ids reason.
- REQ-5: When the branch's SPEC is read through a `branch:` line, `moai tdd` shall use that SPEC for its cycle. · failure: none — an unreadable link is REQ-3 and a name that reads two ways is REQ-4.
- REQ-6: The SPEC check shall accept `branch` as a frontmatter key without an `fm_unknown_key` warning.
- REQ-7: The SPEC-first refusal text shall name the `branch:` line as the way to link a branch to a SPEC whose id differs.
- REQ-8: The leader rules shall say that the leader names a lane branch `spec/<slug>-<yyyymmdd>` and that the lane adds `branch:` to its SPEC when the SPEC id differs.

## Acceptance

- AC-1 (REQ-1): in a temporary MoAI repository on branch `spec/lane-x-20261124`, a code write is refused while only `SPEC-20261124-other` exists without a link, and allowed once its frontmatter has `branch: spec/lane-x-20261124` · verify: `cargo nextest run --locked --offline --test spec_before_code_hook a_spec_that_names_the_branch_opens_code_writes`
- AC-2 (REQ-2): a branch with no SPEC at all, and a branch whose only linking SPEC names another branch, are still refused · verify: `cargo nextest run --locked --offline --test spec_before_code_hook a_branch_with_no_spec_and_no_link_stays_refused`
- AC-3 (REQ-3): a `branch:` line in a SPEC whose folder differs from its `id:`, whose status is `completed`, or whose folder is a symbolic link opens nothing · verify: `cargo nextest run --locked --offline --test spec_before_code_hook a_link_that_is_not_a_live_spec_of_its_own_folder_opens_nothing`
- AC-4 (REQ-4): two SPECs naming one branch are refused with the two-ids reason, and removing one makes the write pass · verify: `cargo nextest run --locked --offline --test spec_before_code_hook two_specs_naming_one_branch_are_refused`
- AC-5 (REQ-5): `moai tdd` red on a linked branch records the cycle under the linking SPEC · verify: `cargo nextest run --locked --offline --test spec_branch_name_first a_linked_branch_reaches_tdd`
- AC-6 (REQ-6): `branch` raises no `fm_unknown_key` warning, and another unknown key still does · verify: `cargo nextest run --locked --offline --test spec_doc branch_is_a_known_spec_key`
- AC-7 (REQ-7): the refusal text names `branch:` · verify: `cargo nextest run --locked --offline --test spec_before_code_hook a_spec_that_names_the_branch_opens_code_writes`
- AC-8 (REQ-8): `docs/leader-rules.md` has the lane-branch naming paragraph · verify: `rg -c "Naming a lane branch" docs/leader-rules.md`
- AC-9 (REQ-3): the scan reads only the direct children of `.moai/specs/` and follows no link: a `spec.md` one folder deeper and a `spec.md` that is a symbolic link open nothing · verify: `cargo nextest run --locked --offline --test spec_before_code_hook the_scan_reads_one_level_and_follows_no_link`

## Scope

- write: `.moai/specs/SPEC-20261124-lane-branch-after-spec/spec.md`, `src/spec_branch.rs`, `src/spec_doc.rs`, `src/hook_guards.rs`, `tests/spec_before_code_hook.rs`, `tests/spec_branch_name_first.rs`, `tests/spec_doc.rs`, `docs/leader-rules.md`
- read-only: `src/tdd.rs`, `src/spec_take.rs`, `CLAUDE.md`
- forbidden: `CLAUDE.md`, `tools/repo_scans.json`

`src/tdd.rs` and the Stop check (`ac_work` in `src/hook_guards.rs`) call `spec_branch::resolve` already, so REQ-5 needs no edit there.

## Plan

1. RED: the five hook tests (the fifth is `the_scan_reads_one_level_and_follows_no_link`), the `moai tdd` test and the `spec_doc` test; each fails on the current code (the link is not read; `branch` warns as unknown).
2. GREEN: in `src/spec_branch.rs` add the link scan to `resolve` (only after the name-based reading finds no existing SPEC file); add `branch` to `KNOWN_KEYS` in `src/spec_doc.rs`; extend `SPEC_FIRST` in `src/hook_guards.rs`.
3. `docs/leader-rules.md`: one paragraph "Naming a lane branch".
4. Guard mutants: remove the scan, the folder-equals-id check, the status check, the symlink check and the two-ids arm one at a time and see the named test fail.
5. SYNC: Result with each AC met or not met, then the READY note with the `check affected` receipt.

## Risks

- The scan reads every SPEC folder of `.moai/specs/` (hundreds) and runs on the refusal path only; it reads at most the first 16 KiB of each file and skips a name that is not a SPEC id, so a write that is allowed pays nothing.
- A SPEC that links a branch it should not (a typo) unlocks code writes for that branch: the same strength as a SPEC with the right name, because the SPEC must exist, be `draft` or `approved` and name itself.
- The approval digest may cover the frontmatter: the lane adds `branch:` while the SPEC is still a draft, before approval, so no approved SPEC changes.
- `docs/leader-rules.md` has no Rust pin; AC-8 is a text match, a hypothesis only (core rule 10) that the audit reads whole.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A `branch:` frontmatter line read by the resolver, not a dispatch helper | data in the SPEC; no new command, no reservation rewrite; one resolver serves the hook, `moai tdd` and the Stop check | remove the scan and the key |
| D-2 | The link is read only when the name-based reading finds no existing SPEC file | the common path stays as it was and costs nothing; a branch whose own SPEC exists is judged as before | drop the early return |
| D-3 | A link counts only for a `draft` or `approved` SPEC whose folder equals its `id:` and is not a link | a landed or archived SPEC must not unlock a reused branch name; the folder rule is the existing SPEC id rule | none |
| D-4 | `moai spec new` does not write the line | keeps this change to the reader; the lane writes one line | add `--branch` to `moai spec new` by a later SPEC |
| D-5 | The scan reads one level of `.moai/specs/` and follows no link; the hook change is audited at tier 2 | the leader's conditions at approval (note 7875): the hook is a security gate; the READY note carries one guard-mutant line for "a name-read SPEC is never scanned past" and one for each REQ-3 refusal | none |

## Result

Measured 2026-11-23 (UTC) on macOS in the worktree `.claude/worktrees/lane-branch-after-spec`, code head `3e672308deb92d1a801fafd3744bc3135761b10e`, base `6054432304ef81c384cbb179a561be9f03180d21` (the `NeXT` tip when the worktree was made; the checks did not follow the tip). Commits: draft `2ca8154ab3`, approval `5edc569811`, RED tests `1b9ac09ac7`, code and document `cf7f18b1b9`, repair `3e672308de`.
- Repair: the first `moai check affected` at the head `a0df03f176` answered `result=FAIL` with `test=100`; one test of 3959 failed, `state_folder_literal_scan the_repository_tests_stay_within_the_state_folder_baseline`, because the new tests spelled the state folder name 36 times in `tests/spec_before_code_hook.rs` (baseline 26) and 7 times in `tests/spec_branch_name_first.rs` (baseline 6). `3e672308de` makes the new test code find the SPEC folder through `moai_cli::state_root::path` and leaves `tests/fixtures/state_folder_literals/allow.tsv` as it was; the scan then passed with the touched tests (34 tests, exit 0). The guard mutants below ran at `cf7f18b1b9`, before this test-only change, and were not repeated.
- AC-1 met: `a_spec_that_names_the_branch_opens_code_writes` refuses a code write on `spec/lane-x-20261124` while `SPEC-20261124-other` has no `branch:` line, and allows it once the line names the branch, for `draft` and `approved`; RED at `1b9ac09ac7` (exit 100), GREEN at `cf7f18b1b9`.
- AC-2 met: `a_branch_with_no_spec_and_no_link_stays_refused` refuses with no SPEC and with a link to `spec/lane-y-20261124`, to the longer `spec/lane-x-20261124-r2` and to a name without the `spec/` prefix; it passed before and after the change, as a guard for the rule that code waits for a SPEC.
- AC-3 met: `a_link_that_is_not_a_live_spec_of_its_own_folder_opens_nothing` refuses a link in a SPEC whose folder differs from its `id:` and in a `completed` SPEC, and beside one good link neither is counted.
- AC-4 met: `two_specs_naming_one_branch_are_refused` refuses two links with the reason `reads as two SPEC ids`, allows the write when the branch name reads an existing SPEC file although two links stand beside it, and allows it again after the second link is removed.
- AC-5 met: `a_linked_branch_reaches_tdd` records the cycle of `moai tdd red` under `SPEC-20261124-other`; `src/tdd.rs` is unchanged because it calls the same resolver.
- AC-6 met: `branch_is_a_known_spec_key` shows no warning for `branch` and `fm_unknown_key` for `owner`.
- AC-7 met: the first test above asserts the refusal text contains `branch:` in backticks.
- AC-8 met: `docs/leader-rules.md` has the section "Naming a lane branch"; `rg -c "Naming a lane branch" docs/leader-rules.md` printed 1. A text match only (Risks).
- AC-9 met: `the_scan_reads_one_level_and_follows_no_link` shows that a SPEC folder that is a link, a `spec.md` that is a link and a SPEC one folder deeper open nothing, each beside one good link.
- Targeted run on the clean head `cf7f18b1b9`: `cargo nextest run --locked --offline` over `spec_before_code_hook`, `spec_branch_name_first`, `spec_doc` and `repo_claude_md` ran 90 tests, 90 passed (exit 0, in a heavy slot with `RUSTC_WRAPPER=sccache`). This is the targeted run, not the `moai check affected` receipt, which the READY note quotes.
- Guard mutants, each applied on the committed head `cf7f18b1b9` and restored (`git status --short` and `git diff --stat` empty after the last): name-read SPEC scanned past (`&& false` on the file-exists arm of `resolve`) fails `two_specs_naming_one_branch_are_refused`; `id:` equal to the folder not required fails `a_link_that_is_not_a_live_spec_of_its_own_folder_opens_nothing`; status not required fails the same test; a folder link followed (`entry.path().is_dir()`) fails `the_scan_reads_one_level_and_follows_no_link`; a `spec.md` link followed (`metadata` for `symlink_metadata`) fails the same test; the `branch:` value not compared fails `a_branch_with_no_spec_and_no_link_stays_refused`; the scan removed fails five tests. All seven ended exit 100 and every one was applied.
- Gap: "one folder deeper" has no mutant, because the scan has no recursion to remove; the test pins the behavior and would fail only if a recursion were added.
- Gap: the refusal text (AC-7) and the `branch` key (AC-6) have RED runs at `1b9ac09ac7` and no separate mutant run.
- Gap: `moai spec new` does not write the `branch:` line (D-4); a lane writes it by hand, and a third failure of the hand step would need a command.
- Gap: the scan reads at most 16 KiB of each SPEC file; a `branch:` line past that point is not read, and a SPEC frontmatter is far shorter.
