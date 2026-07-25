# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Implementation Plan

Tier: **L** (5 plan-phase artifacts + `classify.sh` + `progress.md`). Justification in §A.2.

Milestones are ordered by **decision reversibility**: the decisions most likely to change under review come first, mechanical work last.

---

## §A Context

### A.1 What this SPEC is

A triage-and-guard-refinement SPEC over the `S1-internal-date` strict-tier class of `internal/template/internal_content_leak_test.go`. 135 findings / 180 occurrence-class rows across 116 template files, six measured categories with opposite correct treatments.

### A.2 Tier classification — L

| Tier criterion | This SPEC | Verdict |
|---|---|---|
| Files affected | 116 template files + 1 Go guard file + up to 1 CI workflow | > 15 → **L** |
| LOC | Individual edits are small (mostly single-line deletions), aggregate crosses 300 and plausibly 1000 lines touched | ≥ M, ambiguous alone |
| Constitutional | Changes the enforcement posture of a CI-guarded isolation doctrine and adds a carve-out mechanism to a shared guard | → **L** |

The file-count criterion is met by a factor of ~7 on its own; the constitutional criterion is met independently. The provisional Tier M estimate offered at approval time is not supported by the measured scope.

### A.3 Worktree and branch

Authored in the isolated worktree `.claude/worktrees/debt-clear` on branch `spec/template-date-neutrality`, based on `origin/main` at `c7309aeb6`. The main checkout is held by a concurrent session and is never written.

---

## §B Settled Decisions (formerly open questions)

Four of the five iteration-1 clarification markers are resolved by user ruling and are now binding requirements. They are recorded here so no reader re-opens them:

| Question | Ruling | Binding requirement |
|---|---|---|
| Carve-out mechanism | Hybrid: structural gates for DC-1/DC-4, content-anchored allowlist for DC-3/DC-2b/DC-5-PRESERVE. The `design.md` §A admissibility boundary is accepted as stated. | REQ-TDN-010, REQ-TDN-010b |
| DC-1 frontmatter `updated:` (48 rows) | Preserve as authored. No schema change, no render-time neutralization. | REQ-TDN-009 |
| Mirror-capture stamps (11 rows) | Preserve. They are the reader's only staleness signal for a third-party document mirror. | REQ-TDN-011 |
| CI enforcement | Isolated target (`-run TestTemplateNoInternalContentLeak`) added to the existing neutrality workflow, only after the finding count reaches 0. Never a package-wide invocation. | REQ-TDN-013, REQ-TDN-015 |

### Scope correction carried into this plan

`DC-1` matches the frontmatter key **`updated:` only**. Measured: the tree contains **0** dated `created:` lines and **0** dated `version:` lines (`research.md` §F). Iteration-1 prose describing DC-1 as `updated:` / `created:` / `version:` overstated the rule and is corrected.

### Deferred question (NOT a pre-Kickoff gate)

- **Internal-incident prose disposition — deferred to M2.** Several `DC-5` rows anchor a policy or incident by date (`prevents the 2026-04-21 incident recurrence`, `Advisory — 2026-05-17 Policy`, `LANG-COMPLIANCE-001 plan-phase abandonment (2026-05-20)`). Iteration 1 posed this as a pre-Kickoff clarification, which was unanswerable as posed: it asked for a per-finding call over a set that had never been enumerated. It is re-posed as an **M2 step** operating on `triage.tsv` once that file exists — each `DC-5` row is adjudicated individually against its own line, and the resulting REMOVE/PRESERVE split is presented for confirmation at that point. This is not a blocker on Implementation Kickoff Approval.

---

## §C Pre-flight

Before any run-phase edit:

1. Re-run the strict guard and confirm the 135-finding / 180-row baseline still holds (the tree may have moved).
2. Confirm the narrow tier and `TestTemplateNeutralityAudit` are green.
3. Confirm branch and worktree: `git -C <worktree> rev-parse --show-toplevel` and `git -C <worktree> branch --show-current`.
4. Re-run `classify.sh` and confirm AC-TDN-021 (set equality with the guard's finding set) still passes.
5. Confirm none of the files about to be edited appears in the `rule_template_mirror_test.go` byte-parity allowlist; if one does, its local counterpart receives the identical edit in the same commit.

---

## §D Constraints

- Template edits under `internal/template/templates/**` only; `make build` after (Template-First, REQ-TDN-017).
- Go edits confined to `internal_content_leak_test.go`; CI edits confined to `.github/workflows/template-neutrality-check.yaml`.
- No local `.claude/` / `.moai/` copy is synchronized from the template tree or vice versa (REQ-TDN-018).
- No `DC-3`, `DC-4`, `DC-1`, or `DC-2b` value is altered.

---

## §E Self-Verification

Every milestone closes with its `acceptance.md` criteria executed and their verbatim output recorded in `progress.md` §E.2.

---

## §F Milestones

### M1 — Confirm the settled decisions against the tree

Deliverable: a recorded confirmation that the four §B rulings are implementable as written against the current tree — specifically, that the DC-1 and DC-4 structural gates are mechanically decidable from line syntax (they are: `research.md` §C.4 line-shape counts), and that the DC-2b directory scope selects exactly 11 files (AC-TDN-017's second command).

No code or template edit. This milestone exists so a ruling that turns out to be unimplementable surfaces before 80 deletions land, not after.

Closes: AC-TDN-001, AC-TDN-002.

### M2 — Produce the auditable triage inventory

Deliverable: `triage.tsv` in this SPEC directory — one header line plus one data row per occurrence-class row, seven columns (`file`, `date`, `line_shape`, `line_no`, `category`, `disposition`, `rationale`). Generated from `classify.sh`, then hand-adjudicated for the `DC-5` rows.

Constraints:
- Data-row count equals `classify.sh`'s row count (AC-TDN-003).
- No `disposition` cell is unset, `TODO`, `UNTRIAGED`, or `TBD` (AC-TDN-004).
- Each of the 22 `DC-5` rows is adjudicated against **its own line**, not the file's first matching line (REQ-TDN-003). This is where the deferred §B question is answered.
- `classify.sh` still reproduces the guard's finding set exactly (AC-TDN-021).

Closes: AC-TDN-003, AC-TDN-004, AC-TDN-021.

### M3 — Remediate the REMOVE set

Deliverable: template edits deleting the date-bearing construct for every REMOVE-dispositioned row, followed by `make build`.

Ordering within M3:
1. The 80 `DC-2a` rows — the largest mechanical cluster. Handle the 73 `Last Updated:` rows first, then the 7 bare-`Updated:` rows in `moai/workflows/` individually (they sit at the end of long changelog paragraphs, not on standalone footer lines).
2. The `DC-5` REMOVE subset, file by file.
3. The 12 dual-shape files (`research.md` §C.5) get special care: delete the prose row, preserve the frontmatter row, in the same file (REQ-TDN-019 / AC-TDN-023).

Watch: `research.md` gap G5 — a removal that leaves an orphaned or empty header block is a quality regression not caught by any AC; check the surrounding block, not just the line.

Closes: AC-TDN-005, 006, 007, 008, 016, 017, 018, 020, 022, 023.

### M4 — Implement the hybrid carve-out

Deliverable: the §B-ruled mechanism in `internal_content_leak_test.go` —

- a line-shape structural gate inside `collectLeakViolations` for `DC-1` and `DC-4`;
- a new `isDateAllowlisted` content-anchored allowlist for `DC-3`, `DC-2b`, and the `DC-5` PRESERVE subset, each entry traceable to a `triage.tsv` row.

Constraint: content-anchored, never line-number-anchored (REQ-TDN-010b / AC-TDN-009). The `pedagogicalAllowlist` precedent is already content-anchored — copy its enforcement shape, not the unused diagnostic line fields.

Note the real cost: `collectLeakViolations` currently scans whole-file text via `FindAllString`, so the structural gate requires the scan loop to become line-aware. This is a structural change to the scan, not a struct-field addition.

Exit condition: strict tier reports zero findings (AC-TDN-012).

### M5 — Report-cap change

Deliverable: `limit := 50` replaced per `design.md` §D so no finding is silently hidden; the truncation branch names a filesystem path holding the full listing.

Placed after M4 because with zero findings the cap is untestable against real data; AC-TDN-010's injection recipe supplies a synthetic finding instead.

Closes: AC-TDN-010.

### M6 — CI enforcement (conditional)

Deliverable: a strict-tier step in `.github/workflows/template-neutrality-check.yaml`, scoped by `-run TestTemplateNoInternalContentLeak`, **only if** preconditions P1-P3 (`design.md` §C) all hold — including P2's two future-date probes (DC-4 attribution and DC-1 frontmatter bump).

If any precondition fails, M6 closes as "not adopted" with a `precondition_failed: P<N>` line written to `progress.md`. That is a valid outcome, not a partial milestone, and AC-TDN-011's second command checks it mechanically.

Closes: AC-TDN-011, AC-TDN-015, AC-TDN-019.

### M7 — Non-regression sweep

Deliverable: narrow tier green, `TestTemplateNeutralityAudit` green, `go build ./...` green, and the full `acceptance.md` matrix executed with verbatim output in `progress.md` §E.2.

Closes: AC-TDN-002, AC-TDN-013, AC-TDN-014.

---

## §G Anti-Patterns

- **Sweeping all 135.** The category counts exist precisely because a uniform sweep would destroy 100 legitimate rows.
- **Expressing a REMOVE criterion as a hand-written grep.** Iteration 1 did this and produced an 18-line blind spot where the criterion reached PASS with DC-2-shaped dated lines still in the tree. Criteria that assert on a category must invoke the committed classifier.
- **Collapsing a dual-shape finding to one disposition.** 13 findings span two categories; a single disposition is wrong for one of the two lines.
- **Blind local↔template copy.** Prohibited in both directions; the two trees are intentionally divergent.
- **Line-number-anchored carve-outs.** Drift-prone, and unnecessary — the existing precedent is content-anchored.
- **Enforcing CI before both future-date probes pass.** A guard that blocks a legitimate future attribution entry or a routine frontmatter bump is a regression.
- **Replacing a removed date with a placeholder token.** Creates a new grep surface with no gain (REQ-TDN-007 / AC-TDN-018).
- **Citing "pre-existing package failures" as CI rationale.** Measured false — `go test ./internal/template/...` exits 0. Do not re-import that claim from the workflow file's stale comment.

---

## §H Cross-References

- `spec.md` — requirements REQ-TDN-001..020 and the Out of Scope boundary
- `design.md` — the five settled decisions, rejected alternatives, CI preconditions
- `research.md` — all measured baselines, the classifier equivalence proof, and seven recorded gaps
- `acceptance.md` — the 23-criterion matrix each milestone closes against
- `classify.sh` — the committed classifier (REQ-TDN-006 / REQ-TDN-020)
