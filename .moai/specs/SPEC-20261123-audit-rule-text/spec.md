---
id: SPEC-20261123-audit-rule-text
title: "One audit rule text: the audit agent, the RUN Auditor row and the lander's audit brief carry the same four risk tiers word for word"
status: draft
revision: 1
created: 2026-11-23
parent: null
---

# One audit rule text

## Intent

Four texts told an auditor four different things on 2026-11-22:
- the installed audit agent: its own worktree and a home build folder;
- the RUN "Check-run rules" Auditor row: fresh-clone reproduction and the full check set;
- the "Heavy checks are serialised" sentence: one full build per machine, though the machine has three to five slots;
- the leader's audit briefs: four different build settings in four hours.

So audits rebuilt what the lane had built, cloned the repository again, or failed on a setting clash. A round took 50 to 70 minutes with 5 to 11 builds.

This SPEC writes the risk tiers once, as four fixed lines, and makes three texts carry them word for word:
- the audit agent definition (`agents/audit.md`, installed as `moai-audit.md`);
- the RUN Auditor row and the sentences around it;
- the lander's audit brief in the AUDIT workflow.

A scan test pins the four lines in all three. The audit agent loses `isolation: worktree`: the lander gives it an audit worktree.

### Sources (read whole, 2026-11-22T19:2xZ, files under `~/.moai/reports/moai-adk/`)

- `agentic-redesign-2026-11-22.md`, the user-approved redesign:
  - §3 D: tiers 0 docs, 1 tests only, 2 code, 3 high-risk.
  - "규칙 글 하나로: `moai-audit.md`(isolation 제거), `verify.md` Auditor 줄, 리더 지시를 같은 등급표로" (one rule text: `moai-audit.md` without isolation, the `verify.md` Auditor row and the leader's brief on the same tier table).
  - §5 card ⑪.
- `audit-landing-process-review-2026-11-22.md`:
  - A1: the isolated worktree rebuilds the lane's head, 15 to 25 minutes.
  - A2: 5 to 11 builds a round; the build counts per tier.
  - A3: the four texts that disagree, quoted above.
  - A7: git commands refused inside the isolated worktree.
  - B1: a merged-tree run belongs to the lander only.
  - C1: a READY without a receipt reached an audit.
  - Plan "카드가 필요한 것" 2 (cards needed): drop `isolation: worktree`; the Auditor row's "full check set, fresh-clone reproduction" becomes the tier rule; "at most one full build" becomes the slot rule.
- Leader decision 6924 (the ledger card's note): use today's landed audit findings as input:
  - who audits each tier;
  - the light audit by the lander;
  - a receipt with every step 0 (a non-zero step is not READY);
  - the merged-tree run before landing when the tip moved.
- Code read at tip 895df943e1:
  - `src/gate.rs` `tier0_safe`: tier 0 may hold `docs/**`, `tests/**` and Markdown outside `skills/` other than `AGENTS.md` and `CLAUDE.md`.
  - `TIER2_PREFIXES`: the tier 2 floor.
  - `tier_of`: a stated tier above 2 is not read, and paths only raise it.
  - `src/bundle.rs` `stated_tier`: a digit from 0 to 2.
- Text slice C of SPEC-20261122-next-flow-skills-agents (head 44f16c6ab3, in audit): AUDIT step 9 and the audit contract's "Audit tiers" section. This SPEC rewrites both; its file edits start after that slice lands.

## Non-goals

- No code change. The tiers keep the code's numbers, 0 to 2. A fourth number (tier 3) would change `tier_of` and `stated_tier`; it is not done here.
- No `moai audit-prep` command and no CoW copy of the lane's build. Those are another card. Until it lands, the lander makes the audit worktree with plain git.
- The implementation agent keeps `isolation: worktree`.
- No change to the merge gate, `audit-gate` or the bundle check.

## Requirements

- REQ-1: The audit contract `skills/moai/references/audit.md` shall hold the four tier lines in its "Audit tiers" section as the one source text.
- REQ-2: The audit agent definition `agents/audit.md` shall carry the four tier lines word for word.
- REQ-3: The RUN workflow `skills/moai/workflows/run.md` shall carry the four tier lines word for word under its "Check-run rules" table.
- REQ-4: The AUDIT workflow `skills/moai/workflows/audit.md` shall carry the four tier lines word for word in the lander's audit brief.
- REQ-5: The audit agent definition shall not set `isolation: worktree`.
- REQ-6: When the lander starts the audit agent, the audit agent shall confirm that the HEAD of the audit worktree named in the brief's `scope` equals `head` before it reads anything.
- REQ-7: The RUN Auditor row shall name the tier lines in place of fresh-clone reproduction and the full check set.
- REQ-8: The RUN heavy-check sentence shall state the slot rule of `moai heavy-slot-acquire` in place of one full build per machine.
- REQ-9: The lander's audit brief shall name the receipt rule that a ready note whose receipt has a non-zero step is not ready.
- REQ-10: The lander's audit brief shall name the rule that only the lander runs the merged tree and only when the tip moved and the changed files overlap.
- REQ-12 (REQ-6): If the audit worktree named in `scope` is missing or its HEAD differs from `head`, then the audit agent shall end with `RESULT audit blocked` without reading the parent's checkout.
- REQ-11: A scan test in `tests/next_flow_skill_scan.rs` shall fail when any of the four tier lines differs between the audit contract and one of the three texts.
- REQ-13: The AUDIT tier step shall separate what the tools judge from what the sessions keep as a rule.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-11): the four lines that start with `- docs-only (tier`, `- test-only (tier`, `- code (tier` and `- high-risk (tier` in the contract's "Audit tiers" section appear byte for byte in `agents/audit.md`, `workflows/run.md` and `workflows/audit.md`; on the slice C texts the test fails first.
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_tier_lines_match_in_every_text`
- AC-2 (REQ-5): the front matter of `agents/audit.md` has no `isolation: worktree`, and the `agent_defs_package` tests that read the front matter stay green.
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_agent_sets_no_isolation`
- AC-3 (REQ-6, REQ-12): the agent's first step names the worktree in `scope`, `git -C <worktree> rev-parse HEAD` and a stop with `blocked` when it differs, and no longer says a separate worktree may start from the default branch.
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_agent_reads_the_given_worktree`
- AC-4 (REQ-7, REQ-8): the Auditor row holds neither "fresh-clone reproduction" nor "the full check set", the heavy-check paragraph holds no "at most one full build" and names `moai heavy-slot-acquire`, and the scrub word list gains "fresh-clone reproduction".
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_run_auditor_row_names_the_tiers`
- AC-5 (REQ-9, REQ-10): the brief holds the receipt sentence (a non-zero step is not ready) and the merged-tree sentence (lander only, tip moved, files overlap).
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_brief_names_receipt_and_merged_tree`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): every existing test that reads a rewritten phrase stays green, including `product_cli` (package paths in a workflow, the shared base-agent line, size limits), `agent_defs_package`, `lane_rules_2401` and `check_pr_self_audit_test.py`.
  verify: `cargo test --locked --offline --test product_cli --test agent_defs_package --test lane_rules_2401`
- AC-7 (REQ-13): the AUDIT tier step says that the tools judge the floor of the changed paths (`tier_of`) and the tier a PASS note states (`moai bundle check`, `audit_tier_low`), that who audits each class is a rule for the sessions, and that `moai card land` reads no verdict.
  verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_tier_step_names_tool_and_session_parts`

### The four tier lines (the text REQ-1 to REQ-4 carry)

- docs-only (tier 0): only documents or SPEC text change; the lander reads the change and lands it, and nothing is built.
- test-only (tier 0, tier 1 when a tool changes): only tests or test tools change; the lander gives it a light audit: it reads the change and runs the changed test targets once.
- code (tier 1): product code or skill text changes; a session that wrote none of it audits the head: it cites the lane's receipt instead of running the head again, and tries at most three mutants on changed lines, each against one or two targets.
- high-risk (tier 2): a changed path starts with a prefix of the floor list below; a session that wrote none of it gives the full audit: the code steps plus the boundary, security and data-loss checks of this contract.

The tier a ready note states (`audit-tier: 0|1|2 <reason>`) never goes below the floor of its changed paths (`src/gate.rs` `tier_of`). A change that mixes classes takes the highest one.

## Scope

- SPEC slice (now): `.moai/specs/SPEC-20261123-audit-rule-text/**`.
- Text slice (slice C of SPEC-20261122-next-flow-skills-agents has landed):
  - `agents/audit.md`;
  - `skills/moai/references/audit.md`;
  - `skills/moai/workflows/audit.md`;
  - `skills/moai/workflows/run.md`;
  - `skills/moai/references/source-manifest.json`;
  - `tests/next_flow_skill_scan.rs`;
  - every test that reads a rewritten sentence (AC-6).

## Plan

1. (this slice) The SPEC.
2. After slice C lands:
   - re-read the four files at the new tip;
   - write AC-1 to AC-5 and AC-7 as red tests;
   - edit the texts;
   - rewrite the source manifest;
   - run the changed targets and every reader of the changed phrases;
   - run `check affected`; READY with `audit-tier: 2` (the floor: `agents/audit.md` and `skills/moai/workflows/audit.md` are tier 2 prefixes).

## Risks

- R-1: `references/audit.md` is at 5,413 of 6,144 bytes. Four lines of about 600 bytes fit only if step 9's duplicate wording is cut and the floor list stays the only long part. If they do not fit, the lines are shortened in all three texts together.
- R-2: Without `isolation: worktree`, an audit agent started without a worktree in `scope` would read the parent's checkout. REQ-6 makes it stop `blocked` instead.
- R-3: The installed agent and skill change only with the next `moai update`. Until then the lander's brief carries the lines itself, which is why the brief is one of the three texts.

## Decisions

| id | decision | reason | rejected |
| --- | --- | --- | --- |
| D-1 | Four classes on the code's three numbers: docs-only 0, test-only 0 (1 when a tool changes), code 1, high-risk 2. | `tier0_safe` lets `tests/**` be tier 0 and `tools/` is not tier-0 safe; `stated_tier` reads 0 to 2 only; the card asks for four tiers, not four numbers. | The redesign's numbers 0 to 3 (a code change in `tier_of` and `stated_tier`). |
| D-2 | High-risk gets one full independent audit, as the gate counts it. | The gate and the bundle check count one PASS. Two audits would be a rule no tool checks. | Two independent audits (redesign §3 D), open as U-1. |
| D-3 | Skill text and agent definitions are the code class (tier 1), except the paths on the floor list. | `tier0_safe`: "Skill Markdown is product instructions, so it is tier 1 or more". | Treat skill text as documents. |
| D-4 | The audit worktree is passed as the brief's `scope`, not as a new block field. | `product_cli` pins the shared base-agent line ("one block, `task`, `scope`, `base` and `refs`, plus `head` …") identical in five workflows. | A new `worktree` field. |
| D-5 | The lander makes the audit worktree with `git worktree add --detach .claude/worktrees/audit-<name> <head>` and removes it after the verdict note, until a prepare command lands. | Review A1, A7 and the cleanup section; no command exists yet. | Keep `isolation: worktree` until the command lands. |
| D-6 | This SPEC's tier lines replace text slice C's step 9 rule, recorded there as D-19: tier 1 is audited by the lander for documents, tests or tools only. Under them, tools-only is test-only and skill text is code. | One rule text; the leader confirmed D-19 in the slice C verdict note (6928). | Keep both wordings. |
| D-7 | Two follow-ups of that verdict note belong here, not to the next-flow card. F1: the AUDIT tier step says which part the tools judge and which is a rule for the sessions (REQ-13). F5: the RUN heavy-check paragraph wording (REQ-8). | This SPEC rewrites exactly those two passages; two cards editing one sentence would conflict. | Fix them in the next-flow card's next slice. |

### Unsettled

- U-1: whether high-risk needs two independent audits (redesign §3 D) or one (D-2). This is a leader decision at SPEC review.
- U-2: whether the leader's own dispatch template lives in a tracked file. None was found by `git grep -i brief` over `skills/moai` and `agents`. If it does, it is a fourth text for the scan.

## Result

Status: SPEC revision 1; the text slice is written on the branch of this SPEC, on the tip that held slice C of SPEC-20261122-next-flow-skills-agents.

Text slice, as built:
- `skills/moai/references/audit.md` holds the four tier lines in "Audit tiers" (6,106 of 6,144 bytes: the intro, the findings bullet and three side rules were shortened, none of the pinned phrases).
- `agents/audit.md`, `workflows/run.md` (under "Check-run rules") and `workflows/audit.md` ("Base agents", the lander's brief) carry the lines word for word. The scan test `refs_audit_tier_lines_match_in_every_text` pins it, with a planted change, a missing line and a doubled line.
- `agents/audit.md` sets `isolation: none`; the line stays because the agent definition parser requires the key (`none` or `worktree`), and `tests/agent_defs_package.rs` now expects `none` for `moai-audit` and `worktree` for `moai-impl`. Its first step reads the audit worktree named in `scope` and ends `RESULT audit blocked` when it is missing or its HEAD differs (REQ-6, REQ-12).
- The RUN Auditor row names the tier lines; the heavy-check paragraph states the slot rule; "fresh-clone reproduction" is in the scrub word list and left every shipped text.
- The lander's brief carries the receipt rule and the merged-tree rule (REQ-9, REQ-10). The AUDIT tier step separates what the tools judge from what is a rule for the sessions (REQ-13); the old "who audits tier 1" wording is replaced, as D-6 says, and `refs_audit_tiers_match_the_code_floor` follows it.
- AC-1 to AC-5 and AC-7 are the six `refs_` tests of `tests/next_flow_skill_scan.rs`, red first (8 of 51 failed before the texts changed). AC-6 holds: `agent_defs_package`, `lane_rules_2401` and the `product_cli` skill package and base-agent tests pass.

Not done here:
- `skills/moai/workflows/update.md` and `src/worktree_include.rs` still say `moai-audit` runs in an isolated worktree; they are outside this SPEC's scope.
- The brief makes the audit worktree with plain git (D-5); it names a prepare command once one lands.
- U-1 (one or two independent audits for high-risk) stays open for the user; the text carries one audit, as D-2 says.
