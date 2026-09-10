# Plan — SPEC-REVIEW-SECRET-SCAN-REFS-001

Card: t629 (Class C) · Worktree: `.claude/worktrees/t629` · Branch: `WT-secret-scan-refs`
Card base: `feeecc980` · Evidence: `.moai/reports/t629/reproduction.md`, `.moai/reports/t629/cost-baseline.md`

## §A Context

The review workflow's secrets scan anchors its incremental range on HEAD and claims equivalence
with a full-history scan. The fixture reproduction refutes that claim (`spec.md` §1.1). The fix
surface is two byte-identical document copies. The coverage-versus-cost choice is open and belongs
to the operator (`spec.md` §3).

## §B Known issues carried in

- The defect is live on develop; the secret-scan section is unchanged since main.
- On this repository, 14 of 15 full-scan matches are reachable only through refs other than HEAD;
  their content was not examined (`cost-baseline.md` § Gaps). The lead later classified all
  matches by file path only as examples or fixtures — 0 real-leak candidates by path and shape,
  with the gap that a path-only judgement cannot tell a real value inside a documentation or
  example folder apart (`spec.md` §3.3; lead-reported, not re-measured by the lane).
- The only cost measurement was taken under heavy contention.

## §C Pre-flight (before any document edit)

1. **[NEEDS CLARIFICATION: coverage versus cost — which of Options 1, 2, 3 in `spec.md` §3 is
   adopted, and, for Option 3, the period of the full scan]** The orchestrator surfaces this to the
   operator. The answer is recorded in `spec.md` §3 as a `**Decision:** Option N` line in its own
   commit before the run phase touches either `review.md` copy (REQ-008, AC-007).
2. Re-read both `review.md` copies in full, then run `diff -q` on them. If they have diverged since
   plan time, stop and report; do not reconcile by copying one over the other.
3. For Option 3 only: pin the exact sentence that will name the uncovered refs in `progress.md`
   §E.2 **before** the edit, so AC-006 checks a sentence fixed in advance rather than one chosen
   after the fact.

## §D Constraints

- [HARD] The lane does not run `make build`. The lead runs one build and embed check at batch close.
  The lane verifies only the source axis: distributed copy versus local copy.
- [HARD] Template-First order: edit `internal/template/templates/.claude/skills/moai/workflows/review.md`
  first, then `.claude/skills/moai/workflows/review.md`. Read both first; never overwrite either
  copy wholesale from the other.
- [HARD] The fixture lives outside this repository (session scratchpad). Any fixture script or
  evidence file persisted under `.moai/reports/t629/` assembles the marker line from fragments at
  runtime, so no line matching the scan regex is ever committed (REQ-007, AC-004).
- [HARD] The distributed copy carries no SPEC ID, card ID, or cost figure from this repository, and
  names no programming language as primary (REQ-006, AC-005).
- Cost: if the adopted option's cost is re-measured, re-measure on a quieter machine or state the
  contention again (load averages before and after). Never cite the 76.084 s figure as a clean
  benchmark.
- Exercising other unreachable ref kinds — remote-only branches, tags, stash entries, branches
  deleted after push — is inferred rather than measured. It is recorded as a gap (§G) and as a
  candidate criterion in `acceptance.md` §D.8, not as a Definition-of-Done item.
- Do not examine or record the SHAs of the 15 commits matched on this repository.

## §E Self-verification

Every acceptance criterion carries its command in `acceptance.md`. Run-phase evidence goes to
`progress.md` §E.2 with command, verbatim output, and the tree it was measured on. Fixture output is
redirected to files, exit codes are read without a pipe, and each file is measured with `wc -c` and
counted per label with `grep -c`, exactly as in `reproduction.md`.

## §F Milestones (ordered by decision reversibility)

### M0 — Operator decision recorded (Priority High)

The choice most likely to change everything downstream. Record `**Decision:** Option N` in
`spec.md` §3 with attribution and date, committed alone. No document edit precedes it.

### M1 — Checkpoint content and coverage statement (Priority High)

Decide the wording of what the checkpoint records and what each scan step covers, per the adopted
option — a single HEAD SHA (Option 3), a set of per-ref tips (Option 2), or no coverage role
(Option 1). Draft the replacement for the equivalence sentence. This is the data-model decision of
the change.

### M2 — Fixture proof of the adopted procedure (Priority High)

Build the throwaway fixture and run the adopted procedure through the review sequence of
`acceptance.md` §D.1: baseline, HEAD-line control, side-branch cell, over-time cell, and — Option 2
only — the no-movement cell. Record every cell in `progress.md` §E.2. If the procedure fails a cell,
return to M1 before editing any document.

### M3 — Edit the two copies (Priority Medium)

Edit the distributed copy, then the local copy. Verify AC-002, AC-003, AC-005, AC-006.

### M4 — Mechanical closure checks (Priority Low)

AC-004 (no credential-shaped added line since the card base) and AC-007 (decision commit precedes
the first document-edit commit). Populate `progress.md` §E.2 and §E.3.

## §G Risks and gaps

- **Option 2 is unmeasured** — no timing, no fixture run. Its cost claim is an expectation.
- **Option 2, vanished tips (inferred):** a previous tip that has been garbage-collected may make
  `--not <tip>` fail. Whether the procedure should then fall back to a full scan is unexamined.
- **Options 1 and 2 (inferred):** `--all` follows refs only; commits reachable only from a reflog,
  or from no ref, are outside both.
- **Ref kinds not exercised:** only a local branch was built. Remote-only branches, tags, stash
  entries, and branches deleted after push are expected to behave the same but are not measured.
- **Merge commits and stash entries (inferred, all options):** `git log -p` shows no patch for merge
  commits unless a `--diff-merges` variant is given. A stash's working-tree change, and a credential
  introduced while resolving a merge, may therefore be missed even by `--all`. Not measured; see
  `acceptance.md` §D.8 and §D.9.
- **Regex alternatives:** only the PEM-header alternative was exercised.
- **Cost is repository-specific and contended:** the ratio on a project with few branches is not
  measured.
- **Full-history steps surface known example matches:** on this repository, full-history steps
  (Option 1 on every review, Option 3's periodic scan, Option 2's first scan — the last inferred)
  report the 15 path-classified example/fixture matches (`spec.md` §3.3). Handling them — an
  example-value allowlist or path exclusions — is a follow-on operator decision, and path
  exclusions narrow scan coverage (inferred, not measured).
- **No enforcement:** the checkpoint is prose; whether any agent runs this section is not observed.

## §H Anti-patterns to avoid

- Phrasing any option more favourably than its measurements support, or calling Option 2 "cheap"
  without a measurement.
- Citing the 76.084 s figure as a clean benchmark.
- Resolving the open decision inside the lane.
- Copying one `review.md` copy over the other.
- Writing a literal credential-shaped line into any committed file, including evidence scripts.
- Accepting a procedure that reports the HEAD-line marker but never the side-branch marker.

## §I Cross-references

- `.moai/reports/t629/reproduction.md` — defect measurement and controls
- `.moai/reports/t629/cost-baseline.md` — cost measurement and contention caveat
- `.claude/skills/moai/workflows/review.md` § Secrets Scan (Incremental with Checkpoint)
- `.claude/rules/moai/core/verification-claim-integrity.md` — evidence and attribution discipline
