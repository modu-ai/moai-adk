# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Implementation Plan

Milestones are ordered by decision-reversibility: the adjudication decisions most likely to change on review come first, and the mechanical edits that follow from them come last.

---

## §A Context

- **SPEC**: `SPEC-TEMPLATE-DATE-NEUTRALITY-002`, Tier L.
- **Predecessor**: `SPEC-TEMPLATE-DATE-NEUTRALITY-001` (`status: completed`). Its `classify.sh` and `triage.tsv` are the reusable instruments; its artifacts are read-only here.
- **Baseline**: HEAD `760f09f73` (== `origin/main`), branch `spec/template-date-2025`, clean tree.
- **Scope surface**:

| Surface | Files | Nature |
|---|---:|---|
| Template tree — REMOVE rows | 22 | line deletions |
| Guard — `S1-internal-date` year class + allowlist entries | 1 | pattern edit + additive entries |
| CI workflow — quoting normalization | 1 | formatting |
| SPEC directory — `triage.tsv`, `classify.sh` | 2 | new artifacts |

- **Measurement instrument**: a year-widened replica of the predecessor's classifier, validated against the predecessor's own disposition arithmetic (`research.md` §A.3).

### §A.1 Tier justification

Tier L, on the file-count threshold. The edit surface is 22 template files plus the guard plus the CI workflow — 24 files minimum, before any `DC-5` row adjudicated REMOVE adds more. That exceeds the Tier M ceiling of 15 files, and claiming Tier M would require asserting a file count the measurement contradicts.

Two secondary factors support the same answer rather than driving it: the coupled-ordering hazard (`design.md` §A) is a design-level risk that warrants a design artifact, and the higher Tier L audit threshold is proportionate to a change that can turn repository CI red if mis-sequenced.

The counter-argument — that this SPEC is smaller and better-precedented than its predecessor, with the mechanism already designed — is real but does not move the file count.

### §A.2 PRESERVE list

Do not modify:

- `.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/**` — completed SPEC artifacts.
- Existing entries in the guard's date allowlist and the existing structural gate — additive changes only.
- The narrow-tier class set and the separate neutrality-audit test.
- Any `202[6-9]` date literal in the template tree — the predecessor's remediated state is a fixed baseline.
- `internal/template/catalog.yaml` — build-generated.
- The main checkout at `/Users/goos/MoAI/moai-adk-go` and the sibling worktree `.claude/worktrees/debt-clear` — owned by parallel sessions.

---

## §B Known Issues

| # | Risk | Mitigation |
|---|---|---|
| B1 | Assuming a category carves a row when it does not. All three structurally-gated categories are empty for this set. | `spec.md` REQ-TDN2-001 states each emptiness and its reason; M4 derives allowlist entries from the triage record, not from assumption. |
| B2 | Allowlist masking hides an un-performed deletion (`design.md` §C). | REQ-TDN2-015; `acceptance.md` verifies removals with file-scoped greps independent of the guard. |
| B3 | Widening before the tree is clean turns CI red on 48 findings. | REQ-TDN2-017 fixes the ordering; M5 is gated on M3 and M4 both being verified. |
| B4 | Mechanical `DC-2a` sweep edits a fenced documentation sample. | REQ-TDN2-012; the single fenced row is named in `research.md` §H. |
| B5 | A criterion that a formatting-only change can flip. | REQ-TDN2-020; property and normalization criteria are kept separate (`design.md` §D). |
| B6 | Criterion runs from the wrong directory and passes vacuously. | Every criterion in `acceptance.md` runs from the worktree root with repo-root-relative paths; none uses a bare filename or `go test .`. |
| B7 | An `awk`/`grep` that errors on a missing file returns the same value as a clean pass. | Each criterion pairs its measurement with an existence check, and `acceptance.md` records the pre-change baseline so a criterion that cannot move is visible. |
| B8 | Template edit not reflected in the embedded binary. | REQ-TDN2-022; the build step runs after the last template edit. |
| B9 | Cross-tree copying between the template tree and local working trees. | REQ-TDN2-023; edits are confined to `internal/template/templates/**`. |
| B10 | Parallel session commits into the shared checkout. | All git operations use `git -C <worktree>`; commits stage by explicit pathspec, never `git add -A`. |

---

## §C Pre-flight

Run from the worktree root before the first edit:

```bash
git -C <worktree> rev-parse --short HEAD          # expect 760f09f73
git -C <worktree> branch --show-current           # expect spec/template-date-2025
git -C <worktree> status --porcelain              # expect empty
go build ./...                                     # expect exit 0
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... \
  -run TestTemplateNoInternalContentLeak -v        # expect PASS (green baseline)
go test ./internal/template/... -run TestTemplateNeutralityAudit -v   # expect PASS
```

A red strict tier at pre-flight means the working tree is not at the recorded baseline; stop and reconcile before proceeding.

---

## §D Constraints

- **C-1** Go edits confined to `internal/template/internal_content_leak_test.go`.
- **C-2** CI edits confined to `.github/workflows/template-neutrality-check.yaml`.
- **C-3** Template edits confined to `internal/template/templates/**`.
- **C-4** No line-number anchors in any allowlist entry or acceptance criterion.
- **C-5** No placeholder substitution for a removed date (REQ-TDN2-009).
- **C-6** Additive-only against the predecessor's carve-out entries.

---

## §E Self-Verification

On completion, report:

1. AC PASS/FAIL matrix with the command and its observed output per row.
2. Strict tier result after widening, with the finding count.
3. Narrow tier and neutrality-audit results (non-regression).
4. `go build ./...` result.
5. Per-file removal verification for the two dual-category files (B2).
6. Triage record completeness: zero `DC-5` rows carrying a placeholder disposition.
7. Any row whose adjudication differs from its sub-shape default, with the reason.

---

## §F Milestones

### M1 — Triage instrument and record

Commit the year-widened classifier to the SPEC directory and generate the full occurrence-class record.

- Copy the predecessor's `classify.sh`, change only the `DATE_RE` year range to `202[5-9]`, and commit it as this SPEC's `classify.sh`.
- Generate `triage.tsv` with the seven inherited columns, one row per occurrence-class row (74 rows expected).
- Assign each `DC-5` row its sub-shape code in the `rationale` column (REQ-TDN2-003). The plan-phase partition is `EX-FM` 10 / `EX-DATA` 3 / `HIST` 14 / `CREATED` 3 / `DEADLINE` 1 / `COMPOSITE` 2 = 33; the per-row derivation is enumerated in `research.md` §K.
- Record the literal marker `fenced` in the `rationale` column of the one fenced `DC-2a` row (REQ-TDN2-012), so AC-012 can anchor on content rather than on a line number.
- Verify the classifier's `(file, date)` output set has 48 members.

No tree edits. Exit condition: `triage.tsv` exists with 74 data rows and every `DC-5` row carries a sub-shape code.

### M2 — Adjudication of the 33 `DC-5` rows

The highest-change-likelihood step: these dispositions are the ones a reviewer is most likely to want changed, and every later milestone is derived from them.

Resolve the three open questions recorded in `research.md` §J:

- `HIST` (14 rows) — version-history records. Decide whether a released-version date is a factual record worth keeping or internal project history that should not ship. **The predecessor already set a precedent here and it must be on the table before deciding**: its `spec.md` §5 "Known cosmetic residue" note records that it removed the *2026* halves of this same version-history list in `moai-foundation-cc/SKILL.md`, leaving `v3.0.0 (2025-12-06)` / `v2.0.0 (2025-11-26)` (the two `HIST` rows at lines 244-245) stranded with dates while their newer siblings have none. Preserving them perpetuates a residue the predecessor explicitly flagged; removing them resolves it. See `research.md` §J item 1.
- `CREATED` (3 rows) — `Created:` stamps. Decide whether they follow `DC-2a` (REMOVE) or are distinguishable from an authoring stamp.
- `COMPOSITE` (2 rows) — mid-line stamps in a composite footer. Removal is a line edit rather than a deletion; confirm the edit does not become a placeholder substitution under REQ-TDN2-009.

Confirm the defaults for `EX-FM` (10), `EX-DATA` (3), and `DEADLINE` (1) as PRESERVE.

Separately adjudicate the fenced `DC-2a` row (REQ-TDN2-012) and the four dual-category findings (REQ-TDN2-004), recording each determination in the row's `rationale` column.

Exit condition: every one of the 74 rows carries `REMOVE` or `PRESERVE`; no placeholders.

### M3 — Remediation of REMOVE rows

Apply the deletions the triage records. The 28 `DC-2a` rows across 22 files, plus any `DC-5` row M2 adjudicated REMOVE.

- Delete the date-bearing construct; no placeholder (REQ-TDN2-009).
- Do not touch `DC-2b` rows (REQ-TDN2-010) or documentation-example values (REQ-TDN2-011).
- Run the build step after the last template edit (REQ-TDN2-022).

Exit condition: the REMOVE-scoped prose-stamp grep returns 0; the PRESERVE-scoped grep still returns 13; `go build ./...` exit 0.

### M4 — Carve-out entries for PRESERVE rows

Add one content-anchored allowlist entry per preserved `(file, date)` finding. Every preserved row needs one — no structural gate covers this set (`design.md` §E).

- Anchor by the **exact** templates-root-relative path plus the literal date; never by line number (REQ-TDN2-014). The guard compares `entry.File == relPath` — an entry written as a path *suffix* silently fails to match and the row stays a finding.
- Entries added at this point are inert, because the year class has not widened yet and no `2025` literal is a match. This is expected; M5 is what makes them live.

Exit condition: entry count equals the preserved-finding count from `triage.tsv`; the strict tier is still green (the entries changed nothing yet).

### M5 — Widen the year class

The single-line flip that converts the cleanup into an enforced invariant.

- Change the `S1-internal-date` class year range from `202[6-9]` to `202[5-9]` (REQ-TDN2-016).
- Update the descriptive comment that names the class's range in prose, so it does not drift from the implemented pattern (REQ-TDN2-024). The guard holds exactly three `202[6-9]` occurrences — the archive class, the `S1` pattern, and this comment — and two of the three move.
- **Cross-SPEC coupling on that comment.** It is the doc comment of `TestLeakClassNoDateShaInDefaultTier`, a different test enforcing a different SPEC's acceptance criterion (`AC-SBN-018(a)`). Change only the prose parenthetical naming the year range; leave its probe values, assertions, and class membership untouched. That test scans the *default*-tier class set while this SPEC widens a *strict*-tier class, so there is no behavioural coupling — but AC-032 runs it explicitly, because an implementer who declines to touch another SPEC's test would otherwise fail AC-018 with no stated reason.
- Leave the attribution matcher and the narrow-tier archive class unchanged (REQ-TDN2-018).
- Run the strict tier immediately. A non-zero finding count here means M3 or M4 is incomplete — read the reported findings rather than adding allowlist entries to silence them.

Exit condition: strict tier reports zero findings with the widened class.

### M6 — CI quoting normalization and non-regression

- Normalize the three test-target invocations to the quoted form (REQ-TDN2-019).
- Run the full non-regression set: narrow tier, neutrality audit, strict tier, `go build ./...` (REQ-TDN2-021).
- Confirm the property criteria hold both before and after the normalization, and the normalization criteria moved (`design.md` §D).

Exit condition: all four checks green; unquoted invocation count 0, quoted 3.

---

## §G Anti-Patterns

- **Widening first "to see what breaks".** Produces 48 findings against a CI-enforced tier and destroys the green baseline every later step is verified against.
- **Adding an allowlist entry to silence an unexpected M5 finding.** An unexpected finding is evidence that a REMOVE row was missed; silencing it converts a visible gap into a permanent carve-out.
- **Treating a green strict tier as proof of removal.** For the two dual-category files, the allowlist entry masks the REMOVE row (`design.md` §C). Verify per-file.
- **Sweeping all 41 prose stamps.** 13 of them are `DC-2b` PRESERVE. The REMOVE count is 28.
- **Carrying the "41 stamps / 29 files" figure forward.** It measures only one line shape; the triage surface is 74 rows / 48 findings / 34 files (`research.md` §G).
- **Assuming the frontmatter gate carves the `updated:` lines.** It does not — they sit at column 0 and the gate requires indentation (`research.md` §D).
- **Deleting a date from a documentation example.** The construct exists to teach a format; removing the date degrades the example rather than neutralizing a leak.
- **Anchoring an allowlist entry or a criterion to a line number.** Line numbers drift on the very edits this SPEC makes.
- **Rewriting the predecessor's allowlist entries while adding new ones.** Additive only.

---

## §H Cross-References

- `spec.md` — requirements REQ-TDN2-001 … REQ-TDN2-023.
- `acceptance.md` — criteria with their pre-change baselines.
- `design.md` — coupled-ordering design, allowlist-masking hazard, taxonomy rationale, criterion-coupling rule.
- `research.md` — full measurement record and the refuted frontmatter hypothesis.
- `.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/` — predecessor (completed, read-only): `spec.md` §5 records the deferral, `classify.sh` is the instrument, `triage.tsv` is the format precedent, `acceptance.md` §D records the criterion defects this SPEC avoids.
