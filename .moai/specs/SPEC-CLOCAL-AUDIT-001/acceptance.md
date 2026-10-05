# SPEC-CLOCAL-AUDIT-001 — Acceptance Criteria

## D. AC Matrix

| AC | Requirement | Severity | Verification method |
|----|-------------|----------|---------------------|
| AC-CLOCAL-001 | REQ-CLOCAL-002 | Must | Transcript inspection: every DEFECT-CONFIRMED row has command + verbatim output + exit code |
| AC-CLOCAL-002 | REQ-CLOCAL-003 | Must | No UNVERIFIED-HYPOTHESIS item appears under the DEFECT-CONFIRMED label |
| AC-CLOCAL-003 | REQ-CLOCAL-004 | Must | Zero findings whose defect text lies inside the content-defined §4.1 zone (frozen-basis L273–337 advisory); re-anchor log present in progress notes; pointer-validation exceptions allowed |
| AC-CLOCAL-004 | REQ-CLOCAL-005 | Should | KNOWN entries exactly once each AND zero adjudication CHK entries for those topics in the transcript |
| AC-CLOCAL-005 | REQ-CLOCAL-006 | Must | update.go:513 citation check exists WITHOUT re-adjudication text |
| AC-CLOCAL-006 | REQ-CLOCAL-007 | Must | Every INV item carries exactly one verdict class (closure count == 76) |
| AC-CLOCAL-007 | REQ-CLOCAL-008 | Must | Fix diff contains only minimal corrections + dated correction markers |
| AC-CLOCAL-008 | REQ-CLOCAL-009 | Must | `git status --short` shows only the three write-capable paths touched |
| AC-CLOCAL-009 | REQ-CLOCAL-010 | Must | Evidence pack: 3 files present, verdict draft has all five sections |
| AC-CLOCAL-010 | REQ-CLOCAL-011 | Must | Inventory closure report: N confirmed / N attested / N historical / N known / N open |
| AC-CLOCAL-011 | REQ-CLOCAL-001 | Must | HEAD re-read log entries before first edit and before commit |
| AC-CLOCAL-012 | REQ-CLOCAL-001 (§28 surface naming obligation) | Should | Each LSEL finding states which surface (worktree vs primary) was measured |

## D.1 Given-When-Then Scenarios

### AC-CLOCAL-001 — Mechanical evidence for every confirmed defect
**Given** the claims inventory at `.moai/reports/t308/claims-inventory.md`
**When** an item is marked `DEFECT-CONFIRMED`
**Then** `checks-transcript.md` contains the executing command, its verbatim output, and its exit code for that item ID.

### AC-CLOCAL-002 — Hypotheses never masquerade as defects
**Given** the audit is complete
**When** the verifier scans verdict classes
**Then** no item carries `DEFECT-CONFIRMED` without a corresponding transcript entry (empty transcript entry ⇒ misclassification must be fixed before close).

### AC-CLOCAL-003 — Exclusion zone respected across drift
**Given** the content-defined §4.1 zone (frozen-basis coordinates L273–337 at d29b8942e; live boundaries re-derived from the `### §4.1 ...` heading and the horizontal rule before `## 5.`)
**When** any edit lands above the zone during run-phase, and again when the final diff is inspected against the frozen basis
**Then** no correction lands on defect text inside the CURRENT content-derived interval, both boundary re-derivations are logged in the progress notes with their live line numbers, and only §4.1-pointing-pointer validation notes appear in the report.

### AC-CLOCAL-004 — Known-unresolved carried, not re-found (artifact predicate)
**Given** the known-unresolved register (BranchGuard over-match; three §2.3 code defects)
**When** the report and transcript are reviewed together
**Then** each registered topic appears exactly once as a KNOWN entry with sibling-card references, and `checks-transcript.md` contains no CHK entry whose output adjudicates whether those four topics' defects exist — the sole permitted exception is the single citation-only check for update.go:513 (AC-CLOCAL-005).

### AC-CLOCAL-005 — Citation check without re-adjudication
**Given** the INV-073 register entry for `update.go:513` (archiveLegacySkills-after-wipe)
**When** `checks-transcript.md` and the report are read together
**Then** exactly one CHK entry exists for INV-073 and its recorded evidence is limited to symbol-and-line presence at d29b8942e, while the report carries NO text asserting whether the registered code defect still exists.

### AC-CLOCAL-006 — Inventory closure
**Given** the 76 frozen INV items
**When** the run ends
**Then** `claims-inventory.md` shows exactly one REQ-CLOCAL-007 verdict class per item and zero blanks; the summary line counts all 76 across classes summing to 76.

### AC-CLOCAL-007 — Minimal-fix discipline
**Given** every DEFECT-CONFIRMED item
**When** the final CLAUDE.local.md diff is reviewed
**Then** each hunk is the smallest text correction preserving tone/format (or adds a dated correction marker for recalibrated history), and each hunk references its INV id in the report.

### AC-CLOCAL-008 — Scope boundary holds
**Given** the run-phase session
**When** `git status --short` runs before close
**Then** modified/untracked paths fall only within `.moai/specs/SPEC-CLOCAL-AUDIT-001/**`, `.moai/reports/t308/**`, `CLAUDE.local.md`.

### AC-CLOCAL-009 — Five-section verdict draft
**Given** `verdict-draft.md`
**When** read at close
**Then** sections Claim / Evidence / Baseline-attribution / Gaps / Residual-risk are all present, Gaps names what was not observed, and Baseline-attribution pins every figure to a command run in this audit against d29b8942e.

### AC-CLOCAL-011 — Basis re-read at boundaries
**Given** pinned HEAD d29b8942e on WT-clocal-audit
**When** run-phase starts editing and again right before committing
**Then** `git rev-parse --short HEAD` + `git branch --show-current` outputs are recorded in the transcript for both moments.

### AC-CLOCAL-012 — Surface naming for mixed-surface findings
**Given** any finding whose measurement touched a PRIMARY-checkout surface (the authorized §28 reads) or any tracked-vs-untracked split claim
**When** the report row is read
**Then** it names the measured surface explicitly — worktree, PRIMARY, or BOTH — so a reader can tell which filesystem instance the evidence came from.

## D.2 Edge cases

- A line-range citation whose SYMBOL moved to a nearby line (±10) — classify CITATION-DRIFT (minor defect), correct to current line, note the symbol held.
- An ambiguous-path referent found BOTH in-repo and in the global memory dir — cite the one a reader can reproduce from repo root; mark the other.
- Untracked-runtime-vs-tracked mixing in one §28 finding — split per surface.
- git-strategy.yaml holding template values at HEAD — the recipe finding stands even though §4.1 (its neighboring topic) is excluded, because the recipe text sits OUTSIDE §4.1.

## D.3 Quality gates

Documentation-audit Tier M: `go vet ./...` unnecessary (no Go change); required gates = E1–E5 self-verification of plan.md, inventory closure (AC-CLOCAL-006), and evidence-attribution spot re-run of any three transcript checks by the lead session.

## D.4 Definition of Done

All Must ACs green with recorded evidence; Open Questions subsection lists remaining hypotheses; progress.md updated; evidence pack complete at `.moai/reports/t308/`.
