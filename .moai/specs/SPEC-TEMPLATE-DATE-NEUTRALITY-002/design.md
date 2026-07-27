# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Design

Design decisions, the alternatives rejected, and the hazards the plan is built around. Measurements referenced here are recorded in `research.md`.

---

## §A The coupled-change problem

Three changes must ship together, and the coupling is not stylistic — each ordering of the pair produces a distinct failure:

| Order | Failure |
|---|---|
| Widen the year class first | The strict tier reports 48 findings immediately. The tier is a CI step, so the repository goes red on the next push touching the template tree, and every subsequent commit in the SPEC is made against a red baseline. |
| Remediate first, never widen | The tree is clean at the moment of the commit, but the guard remains structurally blind to the `2025` shape. Nothing prevents the next author from re-introducing it. This is the predecessor's own stated reason for deferring the pair as a unit. |
| Remediate, carve out, then widen | Each step is verifiable against a green baseline, and the widening is the step that converts a manual cleanup into an enforced invariant. |

The third ordering is adopted. The consequence for the milestone sequence is that the guard edit — normally the "core" change — is deliberately the *last* substantive step, and everything before it runs against a guard that cannot yet see the rows being fixed.

### The blind interval

Between the start of remediation and the widening, the guard cannot verify the work in progress. This interval is unavoidable under the chosen ordering, and it is why the plan carries file-scoped verification commands that do not depend on the guard (`acceptance.md` uses direct greps for the remediation criteria, and reserves the guard for the post-widening criteria).

The interval is bounded by the SPEC, not by a release: all steps land in one change set, so no intermediate state is published.

---

## §B Category taxonomy — why no seventh category

The `2025` set contains a construct the predecessor's six categories do not *name*: the documentation-example value (13 rows — 10 example frontmatter blocks, 3 structured-data or code-sample values). The obvious response is a seventh category.

**Rejected.** The classifier's value comes from a property that a seventh category would break: every category is assigned by a **mechanically decidable, first-match-wins rule** over `(path, line-text, fence-state)`. A category whose membership requires reading the surrounding prose to decide "is this teaching a format?" is not mechanically decidable, and adding it would make the classifier's output depend on the classifier author's judgment rather than on the file.

Two concrete mechanical rules were considered and both fail:

- **Fence state.** Rejected on measurement: `DC-5` splits 18 unfenced / 15 fenced, `DC-2a` and `DC-2b` each contain fenced rows, and the JSON example lives in a file with no fence markers at all (`research.md` §H).
- **Path prefix.** The example rows sit under `reference/`, `references/`, and `schemas/` directories — but so do many `DC-2a` prose stamps and the entire `DC-2b` mirror set. The prefix does not separate them.

**Adopted instead:** the construct is named as a *rationale code* inside `DC-5`, which is already defined as per-row adjudicated residue. Six codes (`EX-FM`, `EX-DATA`, `HIST`, `CREATED`, `DEADLINE`, `COMPOSITE`) are recorded in the `rationale` column of the existing 7-column triage schema.

This preserves three properties simultaneously:

1. The classifier stays mechanical — only its year range changes.
2. The triage record stays comparable to the predecessor's, same seven columns.
3. Rows sharing a construct receive a consistent disposition, because the code is assigned before the disposition rather than derived from it.

### The `DEADLINE` sub-shape and category DC-3

One row (`Next Review: 2025-12-25 or standards update`) is semantically a functional deadline — the predecessor's `DC-3`. It does not land in `DC-3` because that category's decision rule pins a specific date literal rather than describing a line shape.

Generalizing `DC-3` to a shape rule (`Next Review:` / `Expires:` / `Deadline:`) was considered and rejected for a single row: it would change a category rule inherited from a completed SPEC, for no measured benefit, and the predecessor's own `DC-3` rows would then be classified by a different rule than the one their triage record cites. The row is adjudicated inside `DC-5` with the `DEADLINE` code and a PRESERVE default.

---

## §C The allowlist-masking hazard

The carve-out allowlist is keyed on `(file, date)` and matched by an **exact** templates-root-relative path plus the literal date — the guard compares `entry.File == relPath`, not a suffix. It therefore suppresses **every** occurrence of that date in that file.

The exactness matters in two directions. An entry written as a path *suffix* never matches, so the row it was meant to preserve stays a finding and the strict tier reports it. An entry that does match preserves the whole `(file, date)` pair, not the single row the author had in mind — which is the masking hazard below.

For the predecessor's set this was benign, because its dual-shape findings paired a PRESERVE frontmatter row with a REMOVE prose row and the frontmatter row was covered by the *structural gate* rather than the allowlist — the gate matches per-line, so it could not mask the prose row.

For this set the situation inverts. All three structurally-gated categories are empty (`research.md` §C), so **every** preserved row needs an allowlist entry, and two of the four dual-category findings pair a `DC-2a` REMOVE row with a `DC-5` row in the same file under the same date:

```
.claude/skills/moai-workflow-spec/references/examples.md      2025-12-07   DC-2a + DC-5
.claude/skills/moai-workflow-project/references/examples.md   2025-12-06   DC-2a + DC-5
```

If the `DC-5` row is adjudicated PRESERVE, its allowlist entry masks the `DC-2a` row too. A guard-clean strict tier would then be consistent with two different worlds: one where the prose stamp was deleted, and one where it was not.

**Consequence for verification.** A green guard is necessary but not sufficient evidence that the remediation happened. `spec.md` REQ-TDN2-015 requires the removal to be confirmed by a file-scoped check, and `acceptance.md` carries that check as a criterion independent of the guard. This is the specific mechanism behind the general rule that a criterion must be able to fail — an allowlist entry is exactly the kind of change that can make a criterion vacuously pass.

**Rejected alternative: narrowing the allowlist key to `(file, date, line-shape)`**, so a `DC-2a` row and a `DC-5` row under the same date could be distinguished and the masking could be *prevented* rather than merely detected.

The decisive objection is **classifier duplication and drift**. The guard matches on `entry.File == relPath && entry.Date == matched` and has no line context at all — it never inspects the line's shape. Adding shape-awareness to the key would require the Go guard to classify line shapes itself: re-implementing `classify.sh`'s `LS-FM` / `LS-FM-FENCED` / `LS-PROSE-STAMP` / `LS-OTHER` rules, including its fence tracking and its column-0-versus-indented distinction, in a second language. Two implementations of one classification, maintained apart, drift — and a drift between the guard's shape rules and the triage record's shape rules is a worse and quieter hazard than the masking it would fix, because it would make the allowlist silently stop matching rows the triage believes are covered.

Two weaker objections are recorded and explicitly **not** relied on. "It is a Go change to a mechanism inherited from a completed SPEC" is undercut by the fact that this SPEC already edits that same file three ways (widens `S1`, updates a doc comment, appends allowlist entries). "It serves only two rows" understates the case — masking is a *class* hazard, and a future dual-category finding would get no equivalent protection.

**What the accepted mitigation actually buys, stated precisely.** AC-016 provides sufficient *detection* for this SPEC's two measured cases. It does not *prevent* masking, it is scoped to two named files and one line shape, and a future dual-category finding outside that scope would need its own criterion. That is a narrower claim than "achieves the same assurance", which an earlier revision of this section asserted and which is not accurate.

---

## §D Acceptance-criterion coupling

The predecessor's acceptance record documents three of its own criterion defects. This SPEC treats a fourth as a first-class design constraint, because the task that motivated it is in scope here.

### The observed defect

The predecessor's criterion counting workflow test-target invocations returned its expected value **only because** the three invocations were quoted inconsistently — two as `-run 'Name'` and one as `-run Name`. Normalizing the quoting in either direction would have flipped the criterion without changing any behavior. The criterion was measuring the quoting, not the property.

### The rule adopted

`spec.md` REQ-TDN2-020 states it as a requirement: no criterion may depend on quoting style, line numbers, or whitespace. Operationally this splits into two criterion shapes:

- **Property criteria** are written quoting-agnostically (`-run '?Name'?`), so they hold both before and after the normalization. These verify that the step exists and targets the right test.
- **Normalization criteria** assert the formatting change directly and explicitly (count of unquoted forms → 0, count of quoted forms → 3), and are labelled as formatting criteria so no future reader mistakes them for behavioral ones.

Keeping the two shapes separate is the point. Collapsing them reproduces the original defect in the opposite direction.

### Line numbers

The same principle rules out line-number anchors throughout. `spec.md` §4 and this file cite the guard's constructs by name and by pattern, never by line number, because line numbers drift on any edit to the file — including the edit this SPEC makes.

---

## §E Why the empty categories are stated rather than omitted

`DC-1`, `DC-3`, and `DC-4` each carry zero rows. A shorter SPEC would simply not mention them.

They are stated explicitly, with their emptiness reasons, because each represents a carve-out that a reader would reasonably *assume* applies:

- `DC-1` looks like it should cover the 10 `updated:` lines — the task brief made exactly this inference, and it is wrong for a non-obvious reason (column-0 versus indented).
- `DC-4` looks like it should cover any attribution record — and it would, but the attribution file happens to carry no `2025` date.
- `DC-3` looks like it should cover the `Next Review:` row — and semantically it does, but its decision rule is a date literal.

In all three cases the assumption produces the same error: believing a row is already carved when it is not, and therefore omitting its allowlist entry. That error surfaces only at the widening step, as an unexplained red tier. Recording the emptiness converts a latent trap into a documented fact.

---

## §F Scope boundary against the predecessor

SPEC-TEMPLATE-DATE-NEUTRALITY-001 is `completed`. Its acceptance criteria cannot be revived, re-scored, or amended, and its triage record is a historical artifact.

This SPEC therefore:

- **Adds** allowlist entries; it does not rewrite or re-order existing ones.
- **Changes** one year range in one class pattern; it does not restructure the carve-out mechanism.
- **Re-uses** the six-category vocabulary so the two triage records can be read side by side, without importing the predecessor's row counts as if they were this set's.
- **Normalizes** the workflow quoting that the predecessor's criterion accidentally depended on — a change the predecessor could not make while its own criterion was live.

The last point is the reason the quoting normalization belongs here rather than in a separate cleanup: it is only safe to make once the criterion coupled to it is closed.
