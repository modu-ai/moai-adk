# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Implementation Plan

Tier: **L** (5 plan-phase artifacts). Justification in §A.2.

Milestones are ordered by **decision reversibility**: the decisions most likely to change under review come first, mechanical work last.

---

## §A Context

### A.1 What this SPEC is

A triage-and-guard-refinement SPEC over the `S1-internal-date` strict-tier class of `internal/template/internal_content_leak_test.go`. 135 findings across 116 template files, four measured kinds with opposite correct treatments.

### A.2 Tier classification — L

| Tier criterion | This SPEC | Verdict |
|---|---|---|
| Files affected | 116 template files (≥ 1 edited per REMOVE finding) + 1 Go guard file + up to 1 CI workflow | > 15 → **L** |
| LOC | Individual edits are small (mostly single-line deletions), but the aggregate crosses 300 and plausibly 1000 lines touched | ≥ M, ambiguous alone |
| Constitutional | Changes the enforcement posture of a CI-guarded isolation doctrine and adds a new carve-out mechanism to a shared guard | → **L** |

The file-count criterion (`> 15 files` → L) is met by a factor of ~7 on its own; the constitutional criterion is met independently. The provisional Tier M estimate offered at approval time is not supported by the measured scope. Classified **L**, so `design.md` and `research.md` are authored.

### A.3 Worktree and branch

Authored in the isolated worktree `.claude/worktrees/debt-clear` on branch `spec/template-date-neutrality`, based on `origin/main` at `c7309aeb6`. The main checkout is held by a concurrent session and is never written.

---

## §B Known Issues and Open Questions

Each marker below must be resolved via the orchestrator's user-question channel before Implementation Kickoff Approval.

- **[NEEDS CLARIFICATION: mirror-capture stamps]** — `moai-foundation-cc/reference/*-official.md` mirror third-party documentation and carry `Updated: 2026-01-06` as the mirror's capture date (11 findings). Preserving it keeps the reader's staleness signal; removing it treats it as internal authoring metadata. `design.md` §A/§B records the arguments; the call is the user's.
- **[NEEDS CLARIFICATION: DC-1 frontmatter disposition]** — 48 findings are schema-required `updated:` / `created:` / `version:` frontmatter fields. Preserve-as-is (default), neutralize-value-at-render, or schema change. Preserve-as-is is the only option that avoids a schema or render-path change; it also means the distributed template continues to carry per-skill authoring dates. See `design.md` §B.
- **[NEEDS CLARIFICATION: internal-incident prose]** — several DC-5 findings anchor a policy or incident by date (`prevents the 2026-04-21 incident recurrence`, `Advisory — 2026-05-17 Policy`, `LANG-COMPLIANCE-001 plan-phase abandonment (2026-05-20)`). Removing the date may orphan the reference it anchors; keeping it publishes an internal work date. Per-finding call.
- **[NEEDS CLARIFICATION: carve-out mechanism]** — `design.md` §A recommends Option 4 (structural gate for recurring shapes + small content-anchored allowlist for judgement calls). Options 1-3 remain live. This decision determines how much of the guard's scanning loop changes.
- **[NEEDS CLARIFICATION: CI enforcement scope]** — whether strict mode joins the existing neutrality workflow as a second isolated target, or becomes a separate workflow, or remains opt-in indefinitely after remediation.

---

## §C Pre-flight

Before any run-phase edit:

1. Re-run the strict guard and confirm the 135 baseline still holds (the tree may have moved).
2. Confirm the narrow tier is green.
3. Confirm the branch and worktree with `git -C <worktree> rev-parse --show-toplevel` and `git -C <worktree> branch --show-current`.
4. Confirm none of the files about to be edited appears in the `rule_template_mirror_test.go` byte-parity allowlist; if one does, its local counterpart receives the identical edit in the same commit.

---

## §D Constraints

- Template edits under `internal/template/templates/**` only; `make build` after (Template-First).
- Go edits confined to `internal_content_leak_test.go` (+ the workflow YAML if M6 proceeds).
- No local `.claude/` / `.moai/` copy is synchronized from the template tree or vice versa.
- No DC-3 or DC-4 date literal is altered.

---

## §E Self-Verification

Every milestone below closes with the corresponding acceptance criteria in `acceptance.md` executed and their verbatim output recorded in `progress.md` §E.2.

---

## §F Milestones

### M1 — Resolve the open questions (highest change likelihood)

Deliverable: every `[NEEDS CLARIFICATION]` marker in §B resolved and the resolution recorded in this file. No code or template edit.

This milestone is first because M2's category dispositions and M4's mechanism choice both depend on its outcome; running them first would mean redoing them.

### M2 — Produce the auditable triage inventory

Deliverable: `triage.md` (or `triage.tsv`) in this SPEC directory, one row per finding: file, date literal, category, disposition, rationale. Plus the enumeration recipe (`research.md` §B) committed as the regeneration mechanism.

Constraints:
- Row count equals the guard's reported finding count.
- Category counts partition the total with no residual bucket.
- Every DC-5 row carries an adjudicated disposition, not a placeholder.
- Per REQ-TDN-003 and gap G4, each DC-5 finding is confirmed against its own line, not the file's first matching line.

### M3 — Remediate the REMOVE set

Deliverable: template edits deleting the date-bearing construct for every REMOVE-dispositioned finding, followed by `make build`.

Ordering within M3: start with the largest single mechanical cluster (the `Last Updated:` header/footer stamps) so the remaining set shrinks to reviewable size, then handle the per-file DC-5 removals individually.

Watch: `research.md` gap G5 — a removal that leaves an orphaned or empty header block is a quality regression; check the surrounding block, not just the line.

### M4 — Implement the carve-out mechanism

Deliverable: the mechanism chosen in M1, implemented in `internal_content_leak_test.go`, with each carve-out carrying a rationale traceable to a `triage.md` row.

Constraint: content-anchored, never line-number-anchored (REQ-TDN-010). The `pedagogicalAllowlist` precedent is already content-anchored — copy its enforcement shape, not the unused diagnostic line fields.

Exit condition: strict tier reports zero findings.

### M5 — Report-cap change

Deliverable: `limit := 50` replaced per `design.md` §D so no finding is silently hidden.

Placed after M4 because with zero findings the cap is untestable against real data; it is verified against a synthetic injected finding instead.

### M6 — CI enforcement (conditional)

Deliverable: strict tier wired into CI, **only if** preconditions P1-P3 in `design.md` §C all hold, including the P2 future-legitimate-date probe.

If any precondition fails, M6 closes as "not adopted" with the failing precondition recorded — that is a valid outcome, not a partial milestone.

### M7 — Non-regression sweep

Deliverable: narrow tier green, `TestTemplateNeutralityAudit` green, `go build ./...` green, and the full `acceptance.md` matrix executed with verbatim output in `progress.md`.

---

## §G Anti-Patterns

- **Sweeping all 135.** The category counts exist precisely because a uniform sweep would destroy 60+ legitimate dates.
- **Blind local↔template copy.** Prohibited in both directions; the two trees are intentionally divergent.
- **Line-number-anchored carve-outs.** Drift-prone, and unnecessary — the existing precedent is content-anchored.
- **Enforcing CI before the future-date probe.** A guard that blocks a legitimate future attribution entry is a regression.
- **Replacing a removed date with a placeholder token.** Creates a new grep surface with no gain (REQ-TDN-007).
- **Trusting the §C.2 first-pass partition as final.** Gap G4 makes it a first pass, not an adjudication.

---

## §H Cross-References

- `spec.md` — requirements REQ-TDN-001..018 and the Out of Scope boundary
- `design.md` — carve-out options, DC-1 disposition, CI preconditions, report cap, mirror handling
- `research.md` — all measured baselines and the six recorded gaps
- `acceptance.md` — the AC matrix each milestone closes against
