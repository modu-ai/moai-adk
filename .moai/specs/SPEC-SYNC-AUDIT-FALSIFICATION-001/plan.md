# plan.md — SPEC-SYNC-AUDIT-FALSIFICATION-001

> Tier M plan-phase artifact. Implementation plan for the three auditor-body improvements (IMP-1 / IMP-3 / IMP-6) defined in spec.md §C.

## §A Context

### A.1 Mission

Add three falsification obligations to the sync-auditor agent body (`.claude/agents/moai/sync-auditor.md`) and its template mirror. Doc-only — no Go runtime code changes.

### A.2 Files in scope

| File | Role | Edit shape |
|------|------|-----------|
| `.claude/agents/moai/sync-auditor.md` | live retained agent | insert three obligation clauses + a cross-reference to VCI §1.1 surface 3; ~30-60 net new lines |
| `internal/template/templates/.claude/agents/moai/sync-auditor.md` | template mirror (distributed to all users) | byte-identical to the live file after edit |

### A.3 Primary edit surfaces inside sync-auditor.md

- **§ Per-Dimension Mechanical Verification (L55-66 today)** — IMP-1 obligation slots in here (Functionality row gains a "falsify one high-blast-radius AC mechanism" obligation, language-neutral). IMP-6 (AC-class coverage minimum) also slots in here as a sampling-minimum note above the dimension table.
- **§ Output Format → `### Findings` (L83-94 today)** — IMP-3 obligation slots in here (each blocking finding structured per VCI §3 5-section format; `unverified-premise` marker for findings lacking tool output; cross-reference to `verification-claim-integrity.md` §1.1 surface 3).
- **§ Evaluation Contract (L120-129 today)** — carries the normative statement that falsification + VCI-binding + class-coverage-minimum are auditor obligations, not optional.

### A.4 Approach options (decision-record)

| Option | Description | Verdict |
|--------|-------------|---------|
| A — Obligation clauses inline in existing sections | IMP-1 in Per-Dimension Mechanical Verification, IMP-3 in Output Format / Findings, IMP-6 as a sampling note above the dimension table | **ADOPTED** — closest to existing structure, minimal new headings, lowest drift risk |
| B — New top-level section "§ Falsification Obligations" | Single new section consolidating all three IMPs | Rejected — fragments the obligations away from the dimension / output sections they bind to, increasing the chance a future editor misses the binding |
| C — Inline + worked example per IMP | Option A + a worked falsification example for each IMP | Rejected for the template mirror (worked examples referencing SPEC IDs / internal dates break template neutrality); a single generic worked example for IMP-1 is acceptable |

### A.5 PRESERVE / EXTEND (what run-phase must not break)

- **PRESERVE** the existing 4-dimension enum (Functionality / Security / Craft / Consistency) — FROZEN per design-constitution §12 Mechanism 3 (cf. SPEC-AUDIT-GATE-INTEGRITY-001 REQ-AGI-004). IMP-6 adds a sampling minimum ABOVE the dimension table; it does NOT add a fifth dimension.
- **PRESERVE** the existing flat / hierarchical scoring-model selection rule (sync-auditor.md L44-53). The three IMPs are orthogonal to scoring mode.
- **PRESERVE** `permissionMode: plan`, the `tools:` list (no Write tool), and the `Stop` hook (`evaluator-completion`). The IMPs add no new tool permissions.
- **PRESERVE** byte-identity between the live file and the template mirror — run-phase MUST edit both in the same commit, then `diff` to verify.
- **EXTEND** the Per-Dimension Mechanical Verification area with the IMP-1 falsification obligation (REQ-SAF-001) and the IMP-6 sampling minimum (REQ-SAF-003).
- **EXTEND** the Output Format / Findings area with the IMP-3 VCI §1.1 surface-3 binding (REQ-SAF-002).
- **CROSS-REFERENCE** (do NOT duplicate) `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §3 5-section Evidence format. The agent body names the SSOT; it does not copy its content.

### A.6 Template-mirror + neutrality constraints (CLAUDE.local.md §2 + §25)

- The obligation prose in the template mirror is GENERIC — no SPEC IDs, no REQ tokens (`REQ-SAF-*`), no internal dates, no commit SHAs, no audit citations ("Audit N Finding AX"). The obligation is named by what it does (falsify / bind-to-VCI / sample-across-classes), not by this SPEC's identifiers.
- Pre-commit self-check (CLAUDE.local.md §25.3, 5 items) runs before the PR opens.
- CI guard `template-neutrality-check.yaml` is the safety net.

### A.7 Phase 4 mode selection placeholder (run-phase)

Run-phase is a single-agent doc-only edit (1 file live + 1 file mirror, both byte-identical). Phase 4 mode: **solo-sequential** (default) — no fan-out warranted. If the run-phase discovers mid-execution that the IMP-3 cross-reference requires touching `verification-claim-integrity.md` too (it should NOT — this SPEC cross-references only), escalate to the orchestrator before expanding scope.

## §B Known Issues

(none — plan-phase artifacts are self-contained; any open question surfaces as a `[NEEDS CLARIFICATION]` marker below.)

## §C Pre-flight (before M1)

- [ ] Re-read `.claude/agents/moai/sync-auditor.md` end-to-end to confirm the L55-66 / L83-94 / L120-129 anchors still match (line numbers drift; the §-headings are the stable anchor).
- [ ] Re-read `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §3 to confirm the cross-reference target.
- [ ] Confirm `internal/template/templates/.claude/agents/moai/sync-auditor.md` is currently byte-identical to the live file (baseline the mirror).

## §D Constraints (carried from spec.md §D)

- Doc-only.
- Template mirror byte-identical.
- Template Content Neutrality (no SPEC IDs / REQ tokens / dates / SHAs in the mirror).
- VCI cross-reference, not duplication.
- Language-neutrality preserved.

## §E Self-Verification (run-phase manager-develop §E alignment)

Run-phase self-verification MUST include (in addition to the standard E1-E7):

- **E-extra-1** — `diff .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` exits 0 with no output (AC-SAF-004).
- **E-extra-2** — `grep -c 'REQ-SAF\|SPEC-SYNC-AUDIT-FALSIFICATION' internal/template/templates/.claude/agents/moai/sync-auditor.md` returns 0 (AC-SAF-005, template neutrality — the live file MAY reference this SPEC in a provenance comment, but the mirror MUST NOT).
- **E-extra-3** — `moai spec lint .moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/spec.md` exits 0 (or only advisory warnings).
- **E-extra-4** — VCI cross-reference: `grep -c 'verification-claim-integrity.md' .claude/agents/moai/sync-auditor.md` returns ≥1 (the cross-reference is present) AND the agent body does NOT duplicate §1.1 surface 3 / §3 content verbatim (visual check).

## §F Milestones (priority-ordered, no time estimates)

### M1 — IMP-1 AC-mechanism falsification clause (priority a, highest leverage)

Edit the live `sync-auditor.md` § Per-Dimension Mechanical Verification area. Add a normative clause satisfying REQ-SAF-001:

- Scope: at least one high-blast-radius AC per SPEC.
- Action: construct/invoke a runtime probe that falsifies the stated mechanism (negative probe preferred where feasible).
- On observation that the mechanism does not produce the asserted outcome: Functionality row = FAIL + blocking finding naming the AC ID, stated mechanism, probe input, observed outcome.
- Applies on the happy path (cold sync-auditor spawned, regardless of green/red test suite).

Phrasing: language-neutral (project-language auto-detection). A single generic worked example is acceptable.

**Exit gate**: `grep -c 'falsif' .claude/agents/moai/sync-auditor.md` returns ≥ a threshold (TBD at run-phase; the clause is present) AND spec-lint PASS.

### M2 — IMP-3 VCI §1.1 surface-3 binding for Findings (priority c)

Edit the live `sync-auditor.md` § Output Format → `### Findings` area (L83-94 today). Add a normative clause satisfying REQ-SAF-002:

- Each blocking finding cites verbatim tool output as Evidence (not frontmatter text / grep match alone).
- Findings lacking tool output are downgraded to `optional` severity with an `unverified-premise` marker.
- Each blocking finding is structured per VCI §3 5-section Evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
- Cross-reference `verification-claim-integrity.md` §1.1 surface 3 + §3 (do NOT duplicate).

**Exit gate**: `grep -c 'verification-claim-integrity.md' .claude/agents/moai/sync-auditor.md` returns ≥1 AND the agent body does NOT carry a verbatim copy of §1.1 surface 3 / §3 (visual check).

### M3 — IMP-6 AC-class coverage minimums (defense-in-depth)

Edit the live `sync-auditor.md` § Per-Dimension Mechanical Verification area, ABOVE the dimension table. Add a normative clause satisfying REQ-SAF-003:

- Sample ≥1 AC per AC class present in the SPEC's acceptance.md §D AC matrix.
- High-blast-radius AC is mandatory regardless of class.
- Absent classes are skipped (minimum is "per present class", not "per canonical class").

**Exit gate**: `grep -c 'class' .claude/agents/moai/sync-auditor.md` returns ≥ the new-clause threshold (TDD fixture TBD at run-phase).

### M4 — Template mirror + neutrality (HARD gate before M5)

- Apply the same edits byte-identically to `internal/template/templates/.claude/agents/moai/sync-auditor.md`.
- Strip any SPEC IDs (`SPEC-SYNC-AUDIT-FALSIFICATION-001`), REQ tokens (`REQ-SAF-*`), internal dates, commit SHAs, audit citations from the mirror.
- Run the 5-item pre-commit self-check (CLAUDE.local.md §25.3).
- Run `diff .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` — MUST exit 0 with no output (AC-SAF-004).
- Run `grep -c 'REQ-SAF\|SPEC-SYNC-AUDIT-FALSIFICATION' internal/template/templates/.claude/agents/moai/sync-auditor.md` — MUST return 0 (AC-SAF-005).
- `make build` to regenerate the embedded template FS (CLAUDE.local.md §2 — templates are embedded via `//go:embed all:templates`).

**Exit gate**: AC-SAF-004 + AC-SAF-005 PASS.

### M5 — sync-phase (manager-docs)

- CHANGELOG entry.
- Frontmatter `status: draft → in-progress → implemented → completed` per the canonical transition ownership (this plan-phase emits `draft`; run-phase emits `draft → in-progress`; sync-phase emits the rest).
- `progress.md` §E.2 / §E.3 / §E.4 populated by run/sync actors (NOT by this plan-phase agent).

## §G Anti-Patterns (carried from spec.md §F)

- Do NOT author the IMP-1 obligation as "verify the mechanism" (asymmetric — confirmation is satisfied by a vacuous pass). Use "falsify".
- Do NOT duplicate VCI §1.1 / §3 into the agent body (cross-reference instead).
- Do NOT name a specific language's toolchain in the obligation prose (breaks 16-language template neutrality).
- Do NOT leave the template mirror non-byte-identical to the live file.

## §H Cross-References

- spec.md §C (REQ-SAF-001/002/003), §D (constraints), §H (AC summary).
- acceptance.md §D (AC matrix, GWT).
- `CLAUDE.local.md` §2 (Template-First), §25 (Template Internal-Content Isolation).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §3.
- `SPEC-FOURDIM-PHANTOM-001` (IMP-5 owner — out of scope here).
