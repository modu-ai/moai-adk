# Plan — SPEC-REVIEW-SECRET-SCAN-REFS-001

Card: t629 (Class C) · Worktree: `.claude/worktrees/t629` · Branch: `WT-secret-scan-refs`
Card base: `feeecc980` · Evidence: `.moai/reports/t629/reproduction.md`, `.moai/reports/t629/cost-baseline.md`

## §A Context

The review workflow's secrets scan anchors its incremental range on HEAD and claims equivalence
with a full-history scan. The fixture reproduction refutes that claim (`spec.md` §1.1). The fix
surface is two byte-identical document copies. The operator has decided the coverage-versus-cost
question: Option 2, a per-ref tip-set checkpoint, adopted on the condition that it is measured
first, with an exact example-value allowlist and no path exclusions (`spec.md` §3, §3.4).

## §B Known issues carried in

- The defect is live on develop; the secret-scan section is unchanged since main.
- On this repository, 14 of 15 full-scan matches are reachable only through refs other than HEAD;
  their content was not examined (`cost-baseline.md` § Gaps). The lead later classified all
  matches by file path only as examples or fixtures — 0 real-leak candidates by path and shape,
  with the gap that a path-only judgement cannot tell a real value inside a documentation or
  example folder apart (`spec.md` §3.3; lead-reported, not re-measured by the lane).
- The only cost measurement was taken under heavy contention.
- Option 2 has no fixture or timing measurement yet; the run phase measures it before any document
  edit (`spec.md` §3.4).
- The public documentation example access key matches the scan regex (orchestrator measurement,
  2026-09-10), so the allowlist cannot list it literally in a committed file (`spec.md` §3.4,
  REQ-012).

## §C Pre-flight (before any document edit)

1. **Decision recorded.** `spec.md` §3 carries the operator's decision — Option 2, conditional;
   exact example-value allowlist, no path exclusions — committed alone as
   `docs(t629): record the operator's Option 2 and exact-allowlist decision in the SPEC (card t629)`.
   Its adoption conditions are `spec.md` §3.4. No `review.md` edit precedes the decision commit
   (REQ-008, AC-007).
2. Re-read both `review.md` copies in full, then run `diff -q` on them. If they have diverged since
   plan time, stop and report; do not reconcile by copying one over the other.
3. Before the first gate command runs, pin verbatim in `progress.md` §E.2 the Option 2 procedure
   under measurement: how and when the tip set is recorded, the scan command, and the handling of a
   recorded tip that no longer exists (`spec.md` §3.4). The pin sits under a `### Pinned procedure`
   heading, with how and when the tip set is recorded on one line beginning `Tip recording: `, the
   scan command on one line beginning `Scan command: `, and the missing-tip handling on one line
   beginning `Missing-tip handling: `, each written in the form the document will carry, so AC-016
   can find all three verbatim in the edited section (REQ-013). Pin in the same
   place, before the document edit, the sentence that will name the commits the per-review step does
   not cover (REQ-004), so AC-006 checks a sentence fixed in advance rather than one chosen after the
   fact.
4. Obtain and record the lead's approval before gate cell ③ takes its first run on this repository.
   That run reaches the full ref-reachable history, and the full `--all` scan on this repository is
   not re-run without lead approval. The approval record, quoting the lead's message, is committed
   alone, before the commit that records cell ③'s evidence, so the commit graph witnesses the order
   (AC-010).

## §D Constraints

- [HARD] The lane does not run `make build`. The lead runs one build and embed check at batch close.
  The lane verifies only the source axis: distributed copy versus local copy.
- [HARD] Template-First order: edit `internal/template/templates/.claude/skills/moai/workflows/review.md`
  first, then `.claude/skills/moai/workflows/review.md`. Read both first; never overwrite either
  copy wholesale from the other.
- [HARD] The fixture lives outside this repository (session scratchpad). Any fixture script or
  evidence file persisted under `.moai/reports/t629/` assembles the marker line from fragments at
  runtime, so no line matching the scan regex is ever committed (REQ-007, AC-004).
- [HARD] No SPEC artifact or evidence file writes a listed allowlist value, whole or in fragments,
  and neither `review.md` copy carries text matching the scan regex (REQ-007, REQ-012, AC-015).
- [HARD] Gate cell ③ keeps scan output and its recorded tip set in files outside this repository
  and records counts only — no SHA, path, or matched value of any matching commit (`spec.md` §3.4,
  §5).
- [HARD] The distributed copy carries no SPEC ID, card ID, or cost figure from this repository, and
  names no programming language as primary (REQ-006, AC-005).
- Cost: gate cell ③ records load averages before and after each run (`spec.md` §3.4). Never cite
  the 76.084 s figure as a clean benchmark.
- Exercising other unreachable ref kinds — remote-only branches, tags, stash entries, branches
  deleted after push — is inferred rather than measured. It is recorded as a gap (§G) and as a
  candidate criterion in `acceptance.md` §D.17, not as a Definition-of-Done item.
- Do not examine or record the SHAs of the 15 commits matched on this repository.

## §E Self-verification

Every acceptance criterion carries its command in `acceptance.md`. Run-phase evidence goes to
`progress.md` §E.2 with command, verbatim output, and the tree it was measured on. Fixture output is
redirected to files, exit codes are read without a pipe, and each file is measured with `wc -c` and
counted per label with `grep -c`, exactly as in `reproduction.md`.

## §F Milestones (ordered by decision reversibility)

### M0 — Operator decision recorded (Priority High) — done

Recorded in `spec.md` §3 by the commit
`docs(t629): record the operator's Option 2 and exact-allowlist decision in the SPEC (card t629)`,
committed alone. The adoption conditions follow in `spec.md` §3.4.

### M1 — Option 2 measure-first gate (Priority High)

The decision most likely to change everything downstream: whether Option 2 holds up at all. Pin the
procedure (§C item 3), then take gate cells ① and ② on the fixture and, after the lead's approval
(§C item 4), cell ③ on this repository (`spec.md` §3.4; AC-008, AC-009, AC-010). Record every cell
with its verdict in `progress.md` §E.2, and commit that evidence in its own commit before any
`review.md` edit (AC-011). **If any cell is untrustworthy, stop:** no further milestone runs, the
stop is recorded in `progress.md`, and the result goes back to the lead so the question returns to
the operator (REQ-010, AC-012).

### M2 — Checkpoint, coverage, and allowlist wording (Priority High)

Word what the checkpoint records (the tip of every ref at the last completed scan), what each scan
step covers and does not cover (REQ-003, REQ-004), and the replacement for the equivalence sentence.
Word the allowlist rule (REQ-011) and choose its representation from the candidates in `spec.md`
§3.4 — a digest of each listed value, or fragment assembly at scan time — under the constraint that
no committed text matches the scan regex (REQ-012). This is the data-model decision of the change.
A representation that cannot keep exact-value matching without committing regex-matching text is a
blocker to report, not a trade-off to make.

### M3 — Fixture proof of the worded procedure (Priority High)

Build the throwaway fixture and run the worded procedure through the review sequence of
`acceptance.md` §D.1 and the allowlist controls of §D.13 and §D.14. Record every cell in
`progress.md` §E.2. If a cell fails, no document is edited. A fix that changes only the wording,
with the pinned procedure unchanged, returns to M2. A fix that changes the procedure itself is a new
gate round and re-takes the gate before any edit (REQ-013, AC-016): one commit moves the standing
round's pinned procedure, lead approval, and gate evidence out of `progress.md` verbatim into
`.moai/reports/t629/`; the changed procedure is pinned; gate cells ①, ②, and ③ are taken again —
cell ③ only after a fresh lead approval committed alone as in §C item 4 — and committed as in M1;
then M2 and M3 run again.

### M4 — Edit the two copies (Priority Medium)

Edit the distributed copy, then the local copy. Verify AC-002, AC-003, AC-005, AC-006, AC-015.

### M5 — Mechanical closure checks (Priority Low)

Record `K`, the card branch tip, in `progress.md` §E.2 before the develop absorb; the closure checks
run at `K` (`acceptance.md` § Closure-check anchor). AC-004 (no credential-shaped added line since
the card base), AC-007 (the decision commit precedes the first document-edit commit), AC-011 (the
gate evidence commit precedes it too), and AC-016 (the edited copies prescribe the pinned procedure,
unchanged since the gate evidence). Populate `progress.md` §E.2 and §E.3.

## §G Risks and gaps

- **Option 2 is unmeasured until the gate runs** — no timing and no fixture run exist yet; its cost
  claim is an expectation. M1 measures it (`spec.md` §3.4 cells ① and ③), and an untrustworthy cell
  stops the run phase.
- **Option 2, vanished tips:** a recorded tip that has been garbage-collected may make `--not <tip>`
  fail (inferred). M1 measures this as gate cell ② (`spec.md` §3.4), which defines the acceptable
  behaviours; any other outcome stops the run phase.
- **Option 2, when the tip set is recorded (inferred):** a tip set recorded after the scan rather
  than before it could exclude commits that landed while the scan ran. The pinned procedure states
  when tips are recorded (§C item 3).
- **Literal allowlist conflict:** the public documentation example access key matches the scan
  regex (orchestrator measurement, 2026-09-10). Listed literally, it would commit a regex-matching
  line (REQ-007) and make `review.md` match its own working-tree step. M2 must pick a representation
  that keeps exact-value matching without committing regex-matching text (REQ-012).
- **Option 2 (inferred):** `--all` follows refs only; commits reachable only from a reflog, or from
  no ref, are outside it, and REQ-004 requires the document to say so.
- **Ref kinds not exercised:** only a local branch was built. Remote-only branches, tags, stash
  entries, and branches deleted after push are expected to behave the same but are not measured.
- **Merge commits and stash entries (inferred):** `git log -p` shows no patch for merge commits
  unless a `--diff-merges` variant is given. A stash's working-tree change, and a credential
  introduced while resolving a merge, may therefore be missed even by `--all`. Not measured; see
  `acceptance.md` §D.17 and §D.18.
- **Regex alternatives:** only the PEM-header alternative was exercised.
- **Cost is repository-specific and contended:** the ratio on a project with few branches is not
  measured. Gate cell ③ is also taken on a shared machine; its predicate checks that the record is
  complete and consistent, not that the time is clean.
- **Full-history steps keep surfacing this repository's own fixture values (inferred):** Option 2's
  first completed scan reaches the full ref-reachable history and reports the path-classified
  example and fixture matches (`spec.md` §3.3). The operator's allowlist holds public example values
  only, so this repository's non-public fixture and test values are not suppressed and keep being
  reported by full-history steps. Their count is not measured, because the matched content was not
  opened (`spec.md` §5).
- **No enforcement:** the checkpoint is prose; whether any agent runs this section is not observed.

## §H Anti-patterns to avoid

- Phrasing any option more favourably than its measurements support, or calling Option 2 "cheap"
  without a measurement.
- Citing the 76.084 s figure as a clean benchmark.
- Softening, widening, or re-opening the operator's decision inside the lane — adding an option, a
  path exclusion, or a non-public value to the allowlist.
- Editing either `review.md` copy before the gate evidence commit, or after an untrustworthy gate
  cell.
- Copying one `review.md` copy over the other.
- Writing a literal credential-shaped line into any committed file, including evidence scripts, or
  writing a listed allowlist value, whole or in fragments, into any SPEC artifact or evidence file.
- Accepting a procedure that reports the HEAD-line marker but never the side-branch marker.

## §I Cross-references

- `.moai/reports/t629/reproduction.md` — defect measurement and controls
- `.moai/reports/t629/cost-baseline.md` — cost measurement and contention caveat
- `.claude/skills/moai/workflows/review.md` § Secrets Scan (Incremental with Checkpoint)
- `.claude/rules/moai/core/verification-claim-integrity.md` — evidence and attribution discipline
