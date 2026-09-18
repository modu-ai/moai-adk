# Acceptance Criteria — SPEC-OUTPUT-STYLE-SWITCH-PATH-001

Every criterion below names the command whose output decides it. All criteria are
**source-level**: none is decided by a built binary, and none requires
`make build`, `moai doctor`, or a `moai` invocation of any kind. The embed
refresh the change owes is the batch-close step (`spec.md` REQ-OSP-010), not a
gate here.

`<W>` abbreviates `.claude/output-styles/moai`.
`<T>` abbreviates `internal/template/templates/.claude/output-styles/moai`.

All commands run from the repository root of the run-phase worktree.

## §A. The addition

- AC-OSP-001 (maps REQ-OSP-001, 003): Given the shipped M1 working copies, When

  ```
  for f in moai.md moai-easy.md moai-learn.md; do \
    printf '%s %s\n' "$f" "$(grep -cE '/output-style([^-a-zA-Z]|$)' <W>/$f)"; done
  ```

  runs, Then it prints `moai.md 0`, `moai-easy.md 4`, `moai-learn.md 1`. Any other triple fails this criterion.

  - RED-now (base `9dcbc3dbe`, same command): `moai.md 0`, `moai-easy.md 0`, `moai-learn.md 0`. The whole six-file anchored sweep returns zero hits.
  - Non-empty-sweep control (C7): the `moai.md 0` row is the one value satisfiable by a sweep that never ran. It is witnessed by the two non-zero rows in the **same command invocation** — a mistyped path or wrong directory yields zero on all three, which fails this criterion outright. The `0` is asserted only alongside a `4` and a `1` produced by the same loop.

- AC-OSP-002 (maps REQ-OSP-001, 007): Given the shipped M1 template mirrors, When the same loop runs against `<T>` instead of `<W>`, Then it prints the identical triple: `moai.md 0`, `moai-easy.md 4`, `moai-learn.md 1`. A mirror count differing from its working-copy count is a drift finding, not a rounding difference.

- AC-OSP-011 (maps REQ-OSP-004, 005): Given the shipped M1 code, When the sentence added at `<W>/moai-learn.md` line 35's site and any one of the four sentences added at `<W>/moai-easy.md`'s sites are compared byte-for-byte, Then they are **not identical**. The evidence report quotes all five added sentences side by side.

  This is the mechanical half of the register requirement and it is deliberately
  weak: it catches the paste failure **between** the two files and cannot see a
  single parenthetical repeated at all four `moai-easy.md` sites. The semantic
  property — that each persona sounds like itself — is carried by review of the
  quoted sentences in the M2 evidence, and is named here rather than left implied.
  A criterion that claimed to mechanically verify register would be overclaiming.

## §B. Additivity and boundaries

- AC-OSP-004 (maps REQ-OSP-002): Given the shipped M1 code, When

  ```
  for f in moai.md moai-easy.md moai-learn.md; do \
    printf '%s %s\n' "$f" "$(grep -c '/config' <W>/$f)"; done
  ```

  runs, Then it prints `7` for each of the three files — unchanged from the base measurement. A count below 7 means a `/config` reference was replaced rather than joined, which fails REQ-OSP-002. A count above 7 means a `/config` site was added, which is outside this SPEC's scope and also fails.

  - Base value (`9dcbc3dbe`, same command): `moai.md 7`, `moai-easy.md 7`, `moai-learn.md 7`.
  - This criterion is self-witnessing: it asserts a non-zero count, and an empty sweep yields `0` and fails it.

- AC-OSP-005 (maps REQ-OSP-006): Given the shipped M1 code, When

  ```
  for f in moai.md moai-easy.md moai-learn.md; do \
    printf '%s %s\n' "$f" "$(grep -c '\.moai/config/sections/language\.yaml' <W>/$f)"; done
  ```

  runs, Then it prints `moai.md 7`, `moai-easy.md 3`, `moai-learn.md 6` — unchanged from the base measurement. This is the criterion that separates the configuration-path occurrences of `/config` from the switch-guidance ones: it must hold while AC-OSP-004 also holds, and together they pin that all five edits landed on switch sites and none on a `language.yaml` reference.

- AC-OSP-006 (maps REQ-OSP-007): Given the run-phase diff, When

  ```
  git diff <base>..HEAD --stat -- <W>/moai.md <T>/moai.md
  git diff <base>..HEAD --stat -- <W>/moai-easy.md <T>/moai-easy.md
  ```

  runs — `<base>` pinned to the literal run-phase base SHA, an anchor at which the measurement is taken rather than a moving ref — Then the **first** command prints nothing and exits 0, and the **second** prints a non-empty stat naming both files.

  The second command is the non-empty-sweep control (C7) and is not decoration: an empty first diff is exactly what a wrong `<base>`, a wrong path spelling, or an unstaged tree produces, and in every one of those cases the second diff is empty too. The pair distinguishes "`moai.md` was not touched" from "the diff command saw nothing at all".

- AC-OSP-003 (maps REQ-OSP-007): Given the shipped M1 code, When

  ```
  for f in moai.md moai-easy.md moai-learn.md; do \
    cmp <W>/$f <T>/$f && echo "identical $f" || echo "DIFFER $f"; done
  ```

  runs, Then it prints `identical` for all three files. `cmp` exits 0 on identity, so a silent run is not a pass here — the echoed token is what makes the result readable and the criterion decidable.

- AC-OSP-009 (maps REQ-OSP-009): Given the run-phase diff, When `git diff <base>..HEAD --name-only` runs, Then the changed set is exactly the four files named in `spec.md` §C plus this SPEC's own directory under `.moai/specs/SPEC-OUTPUT-STYLE-SWITCH-PATH-001/`. No file under `.claude/agents/`, `.claude/skills/`, `.claude/commands/`, `.claude/rules/`, `.claude/hooks/`, `internal/`, `pkg/`, `cmd/`, or `docs-site/` appears. Any additional path is a finding to report and adjudicate, not a stray to absorb.

## §C. Template neutrality

- AC-OSP-007 (maps REQ-OSP-008): Given the shipped M1 template mirrors, When

  ```
  go test ./internal/template/... -run 'TestTemplateNeutralityAudit' -v
  go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak' -v
  ```

  runs, Then each exits 0 and its output names the target test as run. The `-v` flag and the named-test check are load-bearing: `go test -run` against a name matching nothing exits **0** and prints `ok … [no tests to run]`, so the exit code alone cannot decide this criterion and a mistyped target would pass it vacuously.

  These two targets are run **in isolation**, matching the CI guard's own
  invocation. A package-wide green is explicitly NOT the bar — the guard
  workflow records that `internal/template` carries failures unrelated to this
  work.

- AC-OSP-008 (maps REQ-OSP-008): Given the shipped M1 template mirrors, When

  ```
  grep -nE '/Users/|CLAUDE\.local\.md|PR #[0-9]|SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z0-9-]+' \
    <T>/moai-easy.md <T>/moai-learn.md
  ```

  runs over **the added lines only** (the lines appearing in this card's diff for those two files), Then it matches nothing on those lines. This is a targeted tripwire on the classes this card could plausibly introduce, not a re-implementation of AC-OSP-007's guards — those own the authoritative verdict. The restriction to added lines is deliberate: the criterion must not be decided by pre-existing content this card did not write.

## §D. Quality gate

- AC-OSP-GATE-001: `go test ./internal/template/...` is run and its result compared against the pre-flight baseline recorded in `plan.md` §C, measured in the same run-phase before any edit. Passing means **no failure that is new against that baseline**. A pre-existing failure quoted on both sides does not fail this criterion; a failure present only on the after side does. A baseline asserted from memory rather than measured before the edit does not discharge this criterion.
- AC-OSP-GATE-002: No `make build`, `go build`, or `moai` invocation appears anywhere in the run-phase evidence. The embed refresh is the batch-close step; a criterion here decided by a built binary would contradict `spec.md` REQ-OSP-010.
- AC-OSP-GATE-003: The clarification at `plan.md` §B.2 is resolved before Implementation Kickoff Approval, and the resolution is recorded in `progress.md`. An unresolved marker blocks run-phase entry; a marker resolved silently at run time, without the recorded answer, fails this criterion. Two commands decide it, and both must hold:

  ```
  grep -rcE '\[NEEDS[ ]CLARIFICATION:' .moai/specs/SPEC-OUTPUT-STYLE-SWITCH-PATH-001/
  grep -c 'open_clarifications: 0' .moai/specs/SPEC-OUTPUT-STYLE-SWITCH-PATH-001/progress.md
  ```

  Then the first prints `0` for every file in the SPEC directory, and the second prints `1`.

  The first pattern is deliberately **self-avoiding**. The marker's syntax is an open bracket, the two words, a colon, the topic, a close bracket — spelled out here rather than quoted, because quoting it would plant the very string the sweep hunts for. Writing the separating space as the bracket expression `[ ]` does the same job for the command line above: the regex matches a literal space, and the text of the command does not contain one at that position, so the criterion cannot match itself. A naive sweep for the two bare words matches this section and can therefore never print `0` — a criterion no correct work could satisfy, and the shape this note exists to keep out.

  - **Resolution of record (plan-phase, 2026-09-18)**: the lead ruled **option (가)** — four files changed, both `moai.md` copies byte-unchanged — on the predicate that a file is not opened to fix a defect it does not have. The sibling-asymmetry question was split to card **t936**. Recorded at `plan.md` §B.2 and `progress.md` §E.1.
  - **Non-empty-sweep control (C7)**: the first command's `0` is the value a mistyped path or a wrong working directory also produces. The second command is its witness — it asserts a **non-zero** count against a file inside the same directory, in the same run, so a sweep that reached nothing yields `0` there too and fails the criterion. Neither command alone decides this AC.
  - This criterion does NOT re-adjudicate the ruling. It asserts only that the question was answered before the gate and that the answer is on record where a later reader will find it.

## §E. The corrected-premise control

- AC-OSP-010 (maps REQ-OSP-011): Given the shipped M1 code, When

  ```
  grep -c '/output-style' <W>/moai.md
  grep -cE '/output-style([^-a-zA-Z]|$)' <W>/moai.md
  ```

  runs, Then the first prints `2` and the second prints `0`, unchanged from the base measurement. The two hits are the file-path substring `output-style-localization-catalogue.md` at lines 245 and 268 — not slash-command references.

  This criterion exists to keep a **corrected false positive** measured rather
  than remembered. An earlier reading of the unanchored count concluded that
  `moai.md` already mentioned the command; it does not. Carrying the pair
  forward means the next reader who reaches for the unanchored pattern meets the
  discrepancy as a recorded fact instead of rediscovering it as a conclusion.

  The `0` here is the anchored assertion and is witnessed by the `2` beside it
  (C7): both figures come from the same file in the same run, so a sweep that
  reached nothing yields `0` and `0` and fails the criterion.

## §F. Definition of Done

All of §A-§E pass; the M2 evidence report exists under `.moai/reports/t906/`
and has been **exported to the primary checkout** before being cited, carrying:
the anchored counts before and after on both trees, the three `cmp` results, the
`/config` and `language.yaml` counts before and after, the paired empty and
non-empty diffs, the two isolated neutrality-guard runs with their exit codes,
the unanchored-vs-anchored `moai.md` pair, and the five added sentences quoted
side by side for register review.

Explicitly **not** in the Definition of Done, by the lead's ruling: a `make build`
embed refresh, a `moai doctor --check "Agent Emit Embed"` result, and any
docs-site change. The first two are discharged once at batch close; the third is
a proposed follow-up card (`spec.md` §D).
