---
id: SPEC-20261121-update-always-asks
title: "The update workflow always asks whether to apply right after the preview and never applies without an answer"
status: draft
revision: 1
created: 2026-11-21
---

# The update workflow always asks whether to apply right after the preview and never applies without an answer

## Intent

`/moai update` in Claude Code and `$moai update` in Codex CLI run `moai project-update-plan` and `moai update --json` to show a preview. Today the workflow text asks about each step in its later steps, but nothing says that the session asks after every preview, whatever it holds: a preview with nothing to apply or with a blocked step can end without a question, and a session that read the user's earlier "continue" could run `moai update --confirm` without an answer to this preview.

The user decided (2026-11-21, "맞다, 카드로 진행"): right after the preview, always ask whether to proceed, through the host's question tool with the recommended option first. When nothing is to apply or a step is blocked, still state the status and what the choice means ("nothing to apply — no change"; "blocked: <reason>, apply the rest?") and record the answer. Never run `--confirm` without an answer in this session.

This change adds one rule to the shipped workflow `skills/moai/workflows/update.md` and a scan test that pins its place before any line that names `--confirm`. It links, and does not repeat, the update's guided migration and the model step: those change what the preview holds and what runs after the yes, not whether the question is asked.

## Non-goals

- The terminal question of the `moai update` command itself (`[y/N]` on a person's own terminal) and the model pick that command asks after an apply: they are Rust behaviour and stay as they are.
- What the preview holds, the order of the steps, the digests and the undo order: unchanged.
- The `/moai project` workflow: it describes the project and ends with one line that points at update; it applies nothing of update's.
- A new command or a new JSON field: the rule is text for the session, which already holds the host's question tool.

## Requirements

- REQ-1: The update workflow shall tell the model to ask the user one selection through the host's question tool right after it shows the preview.
- REQ-2: The update workflow shall state the rule for `/moai update` and for `$moai update` in the same sentence.
- REQ-3: When nothing is to apply, the update workflow shall tell the model to state `nothing to apply — no change` and still ask. · failure: none — an unanswered question is REQ-6
- REQ-4: When a step reads `blocked`, the update workflow shall tell the model to state its reason and ask whether to apply the rest. · failure: none — an unanswered question is REQ-6
- REQ-5: The update workflow shall forbid any applying command, `moai update --confirm` included, in a turn where the question has no answer.
- REQ-6: The update workflow shall tell the model to report a cancelled, empty or unanswered question as open and never as a yes.
- REQ-7: The update workflow shall state the rule before the first line that names `--confirm`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): `skills/moai/workflows/update.md` holds one paragraph that starts with `**Always ask after the preview.**` and names the host's question tool, `/moai update` and `$moai update`, the words `nothing to apply — no change` and the blocked case with its reason and the rest · verify: `cargo test --test update_always_asks the_rule`
- AC-2 (REQ-5, REQ-6): the same paragraph forbids `moai update --confirm`, `--confirm-cleanup`, `--confirm-home` and an apply of a step before an answer, and says that a cancelled, empty or unanswered question is open · verify: `cargo test --test update_always_asks no_apply`
- AC-3 (REQ-7): the paragraph stands before the first line of the file that names `--confirm`, and no line before it runs an applying command · verify: `cargo test --test update_always_asks order`

## Scope

- write: `tests/update_always_asks.rs` (new), `.moai/specs/SPEC-20261121-update-always-asks/**`
- paired (written by the card that holds the file, from a pair line, merged in one chain with this change): `skills/moai/workflows/update.md` (one paragraph), `skills/moai/references/source-manifest.json` (its sha256 entry)
- read-only: `skills/moai/references/internal/questions.md`, `src/update.rs`
- forbidden: the Go repository, the real HOME, any card number in the shipped text

## Plan

Tests first (RED). `tests/update_always_asks.rs` reads the workflow as text, like `tests/project_preview_order.rs`: no command, host or HOME runs. On the tip alone the three groups fail on assertions (there is no such paragraph).

The paragraph (from the holder of the file, placed right before the paragraph that starts `The user layer —`, so that it comes before any line naming `--confirm`):

`**Always ask after the preview.** Right after the preview is shown, in \`/moai update\` and in \`$moai update\` alike, ask the user one selection through the host's question tool, the recommended option first, and record the answer for the report. When something is ready the options are "Apply what is ready (recommended)" and "Leave everything as it is"; when nothing is to apply, say \`nothing to apply — no change\` and still ask, with "Nothing to do, finish (recommended)" first; when a step reads \`blocked\`, say \`blocked: <its reason>\` and ask whether to apply the rest, "Apply the rest (recommended)" first and "Stop here" second. Never run an applying command — \`moai update --confirm\`, \`--confirm-cleanup\`, \`--confirm-home\` or the apply of any step — in a turn where this question has no answer; a cancelled, empty or unanswered question is reported as open, never as a yes, and an earlier "continue" does not answer a new preview.`

## Risks

- The tests of the update route pin other sentences of the file (`tests/update_route.rs`, `tests/project_preview_order.rs`); the paragraph only adds text, so they must stay green when the holder merges it, and the merged chain runs them.
- The shipped text must not gain a maintainer token (a card number, an issue number, a repository link or the local-rules file name): the paragraph has none, and the shipped-text ratchet must still pass with it.
- The file is being rewritten by its holder; the pair line names the anchor by the first words of the paragraph it precedes, not by a line number, so it survives the rewrite.
- A scan of text cannot prove that a live session asks: host use is verified only after the installed skill carries the paragraph (not run until then), as the file's Completion Check already says.
- The question adds one selection to a run that had none for an empty preview: that is the intent of the user's decision.

## Result

(not yet written)
