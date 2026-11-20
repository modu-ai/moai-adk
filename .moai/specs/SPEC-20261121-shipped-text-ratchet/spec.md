---
id: SPEC-20261121-shipped-text-ratchet
title: "A repository scan fails when the shipped text gains one more maintainer-only token than its baseline"
status: draft
revision: 1
created: 2026-11-21
---

# A repository scan fails when the shipped text gains one more maintainer-only token than its baseline

## Intent

The text MoAI ships to every user (`skills/moai/**`, `agents/*.md`, the help texts in `src/help.rs`) must not carry what only this repository's maintainers need: card numbers, issue numbers, links to this repository's tracker, the name of the maintainers' private rules file. Today that text already holds many of them, so a rule "none allowed" would fail at once. This slice adds the first half of the separation the self-improvement report asks for (section 4.3): a count that may fall but may not rise. A change that adds one more token of any kind to the shipped text, or moves one into a file that had none, fails the scan; removing them is a separate, later job.

The scan is repository-only. It reads two files kept under `tools/improve/` (the token list with the shipped paths, and the baseline counts), writes nothing, and ships in no package: `tools/` is outside `skills/moai/` and `agents/`.

## Non-goals

- Removing the tokens that exist now. The baseline records them; lowering it is a card of its own.
- Rule names from `AGENTS.local.md` as tokens: they are prose, not a shape a pattern can hold; only the file name is a token.
- Judging text for meaning. A token is a pattern match, nothing more.
- Reading git history (a base commit): the baseline is a checked-in file, so a branch that raises it shows a diff in `tools/improve/` for the audit to read.
- Wiring the `moai improve` engine, signals, or the evaluation cases (other slices of the same card).

## Requirements

- REQ-1: The scan shall count every match of every token of the token file in every regular file under the shipped paths of the token file.
- REQ-2: WHEN no per-file count is above the baseline and no token total is above its baseline total, the scan shall exit 0. · failure: none
- REQ-3 (REQ-2): IF a token total is above its baseline total, THEN the scan shall exit 1 and name the token with both numbers.
- REQ-4 (REQ-2): IF a file holds a token more often than the baseline records for that file, THEN the scan shall exit 1 and name the file and the token with both numbers.
- REQ-5: The scan shall report every count below its baseline as headroom while it exits 0 for it.
- REQ-6: IF the token file or the baseline is missing, is not JSON or has another shape than the documented one, THEN the scan shall exit 2 with the reason on stderr and nothing on stdout.
- REQ-7: IF a token pattern is not a valid regular expression or can match the empty string, THEN the scan shall exit 2 with the token named on stderr.
- REQ-8: The scan shall write no file.
- REQ-9: The scan shall skip a file that is not valid UTF-8 and list it as skipped.
- REQ-10: The scan of the checked-in tree shall exit 0 with the baseline file of this change.
- REQ-11: The tool and its two data files and its test shall lie outside `skills/moai/` and `agents/` and outside the package source manifest.
- REQ-12: The table `tools/repo_scans.json` shall list the test of the scan with the shipped paths as its scope.

## Acceptance

- AC-1 (REQ-1, REQ-9): on a temporary tree with known matches per token in a nested file, a top-level file and a file outside the shipped paths, `baseline` prints the counts of the files under the shipped paths only (the exact baseline file shape, nothing else on stdout); `check` lists a file holding a byte that is not UTF-8 under `skipped` and does not count it · verify: `python3 -m unittest tests.shipped_text_ratchet_test.CountTest`
- AC-2 (REQ-2, REQ-3, REQ-4): with the baseline taken from a temporary tree, the same tree exits 0; one more match of one token in a file that already holds it exits 1 and names the file token and numbers; one more match in a file the baseline does not list exits 1; a match moved from one file to another with an equal total exits 1 on the per-file rule; a total one above with every file at or below cannot arise, so the total rule is shown by a baseline whose total is lower than its files add up to · verify: `python3 -m unittest tests.shipped_text_ratchet_test.BoundaryTest`
- AC-3 (REQ-5): one match fewer than the baseline exits 0 and lists the file and the token under `headroom` with both numbers · verify: `python3 -m unittest tests.shipped_text_ratchet_test.HeadroomTest`
- AC-4 (REQ-6, REQ-7): a missing token file, a missing baseline, a file that is not JSON, an extra top-level key, a negative or non-integer count, a baseline naming a token the token file lacks, an invalid pattern and a pattern that matches the empty string each exit 2 with stderr text and empty stdout · verify: `python3 -m unittest tests.shipped_text_ratchet_test.RefusalTest`
- AC-5 (REQ-8): a hash of every file of the temporary tree and of the repository checkout paths the scan reads is equal before and after `check` and `baseline` · verify: `python3 -m unittest tests.shipped_text_ratchet_test.ReadOnlyTest`
- AC-6 (REQ-10, REQ-11): `check` on the repository root exits 0 with the checked-in token file and baseline; none of the four files is under the shipped paths or named in `skills/moai/references/source-manifest.json` · verify: `python3 -m unittest tests.shipped_text_ratchet_test.RepoTreeTest`
- AC-7 (REQ-12): the table lists the scan with its test, its check command and a scope that reaches `skills/moai/`, `agents/` and `src/help.rs`, and the table's own completeness test passes · verify: `python3 -m unittest tests.repo_scans_test`

## Scope

- write: `tools/shipped_text_ratchet.py` (new), `tools/improve/shipped-text-tokens.json` (new), `tools/improve/shipped-text-baseline.json` (new), `tests/shipped_text_ratchet_test.py` (new), `tools/repo_scans.json` (one entry; reserved by another card until it releases, so this edit is the last step), `.moai/specs/SPEC-20261121-shipped-text-ratchet/**`
- read-only: `skills/moai/**`, `agents/*.md`, `src/help.rs`, `skills/moai/references/source-manifest.json`, `tests/repo_scans_test.py`
- forbidden: any edit of the shipped paths (the baseline records them as they are), `AGENTS.local.md`, the Go repository, the real HOME

## Plan

Tests first (RED), then the tool and the data, then the table entry.

1. RED: `tests/shipped_text_ratchet_test.py` builds temporary trees and runs the tool through `subprocess` (`python3 -I tools/shipped_text_ratchet.py check|baseline --root <tmp> --tokens <f> --baseline <f>`), so the test exercises the command a maintainer runs. Classes `CountTest`, `BoundaryTest`, `HeadroomTest`, `RefusalTest`, `ReadOnlyTest`, `RepoTreeTest`.
2. GREEN: the tool (standard library only; written for Python 3.9 because the first `python3` on the maintainers' PATH can be the system 3.9), then the token file with four tokens and the shipped paths, then the baseline generated by `baseline` on the base tree and read back.
3. The table entry once `tools/repo_scans.json` is free; until then the SPEC stays open on AC-7.
4. Mutants (at most 6): `>` changed to `>=` in the per-file rule, the same in the total rule, a file missing from the baseline counted as its own count instead of zero, the empty-match refusal removed, the UTF-8 skip turned into a crash, the `skills/moai` start path dropped from the walk.

Tokens (first list, each a pattern and a reason): `card-number` `\bC-[0-9]+\b`; `issue-number` `(?<![\w&/])#[0-9]{3,5}\b`; `repo-link` `modu-ai/moai-adk`; `local-rules-file` `AGENTS\.local\.md`.

## Risks

- Headroom regrows. A file that drops from five to three leaves room for two again, because the baseline is only lowered by hand. `check` lists the headroom so the next change can tighten it; a strict mode that fails on it would make every reduction conflict with every other branch's baseline edit, so it is not built here.
- A branch can raise the baseline instead of fixing its text. The baseline is a checked-in file, so the raise is a diff the audit reads; the scan alone cannot tell a reasoned raise from a hidden one.
- A match is a shape, not a meaning. a card-number shape used as an example in a user-facing help line counts the same as a real card number; the baseline absorbs the existing ones and the scan will flag a new one, which the author then rewrites.
- A new line in the shipped text now fails this scan on every branch that touches it (the scan is in the repo-scans table). That is the aim, and it means a branch written before this scan landed can fail at its merge.
- `tools/repo_scans.json` is held by another card at the time of writing; the entry (REQ-12) waits for it.

## Result

(not yet written)
