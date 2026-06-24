# Acceptance — SPEC-V3R6-SKILL-DECISION-HEURISTICS-001

> Acceptance criteria, Given-When-Then scenarios, edge cases, and Definition of Done.
> Companion to `spec.md` (REQ-SDH-001..009) and `plan.md` (M1-M3).

---

## §A. AC Matrix

| AC ID | REQ trace | Severity | Description |
|---|---|---|---|
| AC-SDH-001 | REQ-SDH-001 | MUST | All 4 target SKILL.md files have a `## Decision Heuristics` section |
| AC-SDH-002 | REQ-SDH-002 | MUST | Each Decision Heuristics section contains 3-5 "if X, default to Y" heuristics |
| AC-SDH-003 | REQ-SDH-003 | MUST | Each Decision Heuristics section is ≤ 13 lines PASS; ≥ 14 lines VIOLATION (binary) |
| AC-SDH-004 | REQ-SDH-004, REQ-SDH-008 | MUST | No new doctrine invented; bodies preserved (additive-only diff) |
| AC-SDH-005 | REQ-SDH-005 | SHOULD | Existing evolvable rationalization/red-flag rows in the **3 evolvable-bearing skills** (foundation-core, workflow-spec, foundation-cc) with memory-mapped anti-pattern families gain a one-line provenance binding. N/A for moai-meta-harness (0 evolvable markers — §F A-3). |
| AC-SDH-006 | REQ-SDH-006 | MUST | No fabricated dates; pending-provenance form used where memory lacks the fact |
| AC-SDH-007 | REQ-SDH-007 | MUST | All 4 skills still pass the skill frontmatter lint (YAML valid, no snake_case aliases introduced) |
| AC-SDH-008 | REQ-SDH-009 | MUST | plan.md flags the template-source mirror sync as a follow-up (REQ-SDH-009); verified by grep of plan.md §I/FU-1, scoped OUT of run-phase |

### §A.1 Severity Definitions

- **MUST** — blocks acceptance if failed. Verified by mechanical grep / line-count / lint.
- **SHOULD** — non-blocking but expected; a justified omission is acceptable (e.g. a skill's evolvable rows have no memory-mapped anti-pattern family — then 0 provenance bindings is honest, not a failure).

### §A.2 Traceability

Every REQ in `spec.md §C` maps to ≥1 AC here. Every AC maps to ≥1 milestone verification step in `plan.md §F` (M3 gate).

| REQ | AC(s) | Verification (M3 step) |
|---|---|---|
| REQ-SDH-001 (Decision Heuristics section present) | AC-SDH-001 | M3.2 (grep ×4 hits) |
| REQ-SDH-002 (3-5 heuristics) | AC-SDH-002 | M3.2 (manual count per section) |
| REQ-SDH-003 (≤12 lines) | AC-SDH-003 | M3.3 (line-count per section) |
| REQ-SDH-004 (no new doctrine) | AC-SDH-004 | M3.5 (body-preservation audit + heuristic-citation check) |
| REQ-SDH-005 (provenance binding, 3 evolvable skills only) | AC-SDH-005 | M3.4 (provenance honesty audit on 3 evolvable-bearing skills) + M3.6 (moai-meta-harness N/A verification) |
| REQ-SDH-006 (no fabricated dates) | AC-SDH-006 | M3.4 (every cited date traced to memory OR pending form) |
| REQ-SDH-007 (frontmatter lint) | AC-SDH-007 | M3.1 (lint run) |
| REQ-SDH-008 (bodies preserved) | AC-SDH-004 | M3.5 (additive-only diff) |
| REQ-SDH-009 (template follow-up flag) | AC-SDH-008 | M3.7 (grep plan.md §I for FU-1 / template-mirror flag, scoped OUT of run-phase) |

### §A.3 Indirect Verification (where direct grep is insufficient)

- **Heuristic faithfulness** (REQ-SDH-004) — indirect. Each heuristic MUST end with a `(<- §<section>)` pointer to the body section it compresses. The pointer is the verification artifact; a reviewer follows the pointer and confirms the heuristic is a faithful compression, not new doctrine. Where the pointer is absent or dangling, the heuristic is suspect.
- **Provenance honesty** (REQ-SDH-006) — indirect. Each cited date/SHA is cross-checked against the memory system files listed in `spec.md §B Memory Provenance Sources`. A date that appears in NO memory file is a fabrication and fails AC-SDH-006.

### §A.4 Closure Gates

- **Gate-1 (functional)** — AC-SDH-001 + AC-SDH-002 + AC-SDH-003 MUST pass. Without all three, the Decision Heuristics device is not honestly delivered.
- **Gate-2 (craft)** — AC-SDH-004 + AC-SDH-007 MUST pass. Bodies preserved (additive-only diff), frontmatter clean, no new doctrine. These protect the existing skill contracts. (AC-SDH-004 traces to REQ-SDH-004 + REQ-SDH-008 jointly.)
- **Gate-3 (honesty)** — AC-SDH-006 MUST pass. No fabricated dates. The no-unobserved-claim invariant binds.
- **Gate-4 (provenance coverage, SHOULD)** — AC-SDH-005 SHOULD pass for every memory-mapped anti-pattern family present in the **3 evolvable-bearing skills'** evolvable rows. N/A for moai-meta-harness (0 markers, honest skip per §F A-3 + M3.6). Non-blocking if a skill's rows have no memory-mapped family (honest 0).
- **Gate-5 (follow-up flagging)** — AC-SDH-008 MUST pass. The template-source mirror sync is flagged as a follow-up in plan.md §I, scoped OUT of run-phase (REQ-SDH-009). This gate is a documentation-existence check, not a run-phase deliverable.

### §A.5 Forward-Looking Checks (anti-regression)

- **FL-1** — `grep -rc "## Decision Heuristics" .claude/skills/moai-*/SKILL.md` must not regress below 4 in subsequent commits touching these files.
- **FL-2** — No evolvable-block delimiter (`<!-- moai:evolvable-start -->` / `<!-- moai:evolvable-end -->`) is removed or unbalanced by this SPEC's edits. `grep -c "moai:evolvable-start" SKILL.md` == `grep -c "moai:evolvable-end" SKILL.md` per file (before == after). For moai-meta-harness this remains 0/0 (no evolvable blocks to preserve — the baseline is honestly zero).
- **FL-3** — Template mirror sync follow-up (FU-1) remains flagged in plan.md §I until the chore/follow-up SPEC lands. This acceptance does NOT require the template sync to be complete (C-4 scope protection). Verified by AC-SDH-008.

### §A.6 Anti-Sycophancy Check

The acceptance verifier MUST NOT pass AC-SDH-004 (no new doctrine) on the basis of "the heuristics look reasonable." Each heuristic MUST carry a `(<- §<section>)` pointer, and the verifier MUST follow ≥1 pointer per skill to confirm the heuristic is a faithful compression. A reasonable-looking heuristic without a traceable source section is a FAIL.

### §A.7 Definition of Done

This SPEC is Definition-of-Done when ALL of the following hold:

- [ ] Gate-1 passes (AC-SDH-001/002/003).
- [ ] Gate-2 passes (AC-SDH-004/007).
- [ ] Gate-3 passes (AC-SDH-006).
- [ ] Gate-4 SHOULD pass (AC-SDH-005), with honest justification for any 0-coverage skill — including the explicit N/A for moai-meta-harness (0 evolvable markers).
- [ ] Gate-5 passes (AC-SDH-008) — template follow-up flag present in plan.md §I.
- [ ] M3.1-M3.7 verification commands all run with observed output recorded (no unobserved claims per `verification-claim-integrity.md`).
- [ ] FL-1/FL-2 anti-regression checks pass.
- [ ] Template mirror sync (FU-1) remains flagged as follow-up — NOT silently dropped.
- [ ] Implementation Kickoff Approval (CLAUDE.local.md §19.1) was obtained before run-phase began.

---

## §B. Given-When-Then Scenarios

### AC-SDH-001 — Decision Heuristics section present in all 4 skills

**Given** the 4 target SKILL.md files at `.claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md`
**When** the verifier runs `grep -l "## Decision Heuristics"` across the 4 files
**Then** all 4 file paths are returned (4 hits, 1 per file).

### AC-SDH-002 — 3-5 heuristics per section

**Given** a Decision Heuristics section in one of the 4 target skills
**When** the verifier counts the "if X, default to Y" rule lines (lines beginning with `- ` or matching the `**If** ... **default** ...` pattern) within that section
**Then** the count is between 3 and 5 inclusive (3 ≤ N ≤ 5).

### AC-SDH-003 — Section ≤ 13 lines (binary threshold)

**Given** a Decision Heuristics section in one of the 4 target skills
**When** the verifier counts the lines from the `## Decision Heuristics` heading to the next `## ` heading (or EOF)
**Then** the line count is **≤ 13 → PASS; ≥ 14 → VIOLATION** (binary, per spec.md §D C-2). No ambiguous 14-line gap.

### AC-SDH-004 — No new doctrine; bodies preserved

**Given** the `git diff` of the 4 target SKILL.md files after the SPEC's edits
**When** the verifier inspects the diff
**Then** (a) the only additions are the closing Decision Heuristics section and the provenance one-liners; (b) no existing line is rewritten or deleted (whitespace-only `-` lines acceptable); (c) each heuristic carries a `(<- §<section>)` pointer to an existing body section.

### AC-SDH-005 — Provenance binding on memory-mapped rows (SHOULD, 3 evolvable-bearing skills only)

**Given** an evolvable rationalization or red-flag row in one of the **3 evolvable-bearing skills** (moai-foundation-core, moai-workflow-spec, moai-foundation-cc) whose conceptual anti-pattern maps to a memory-documented family (AP-V-004, AP-SRN-004, AP-D-001..005, AP-V-001/002/003, AP-GWT-001..006, etc.)
**When** the verifier inspects that row
**Then** a one-line provenance binding is appended (`AP-X — recurred on <DATE> in <SPEC-ID>` or the pending-provenance form). Where a skill's rows have NO memory-mapped family, 0 bindings is honest and acceptable. **moai-meta-harness is explicitly out of scope for this AC** — it has 0 evolvable markers (§F A-3 baseline, verified by M3.6), so there is no evolvable row to bind to; deliverable (b) is N/A for that skill and no content is fabricated.

### AC-SDH-006 — No fabricated dates

**Given** a provenance one-liner citing a `<DATE>` or `<SPEC-ID>` or commit `<SHA>`
**When** the verifier greps the memory system at `~/.claude/projects/{hash}/memory/*.md` for that literal date/SPEC-ID/SHA
**Then** at least one memory file returns a match. A date/SPEC/SHA that matches NO memory file is a fabrication and FAILS this AC. The pending-provenance form (`observed recurrence, provenance pending in memory`) is always acceptable.

### AC-SDH-007 — Frontmatter lint clean

**Given** the 4 edited SKILL.md files
**When** the verifier runs the skill frontmatter lint (or equivalent YAML schema check enforcing the 12 canonical fields + no snake_case aliases)
**Then** 0 findings are emitted. The frontmatter block is unchanged from the pre-SPEC state (additive body edits only).

### AC-SDH-008 — Template-source mirror sync flagged as follow-up (REQ-SDH-009)

**Given** plan.md §I (Follow-Up Flags) and the SPEC's run-phase scope
**When** the verifier runs `grep -nE 'FU-1|Template mirror sync|template-source mirror' plan.md`
**Then** ≥1 hit is returned in plan.md, confirming the template-source mirror sync (`internal/template/templates/.claude/skills/`) is flagged as a follow-up OUT of this SPEC's run-phase scope (per CLAUDE.local.md §2 Template-First Rule + spec.md §D C-4). This AC is a plan-phase documentation-existence check — it does NOT require the template sync to be performed during run-phase. REQ-SDH-009 is satisfied by the flag's presence, not by the sync's completion.

---

## §C. Edge Cases

- **EC-1 — A skill has 0 memory-mapped evolvable rows.** AC-SDH-005 returns 0 bindings for that skill. This is honest (SHOULD, not MUST) and acceptable with a one-line justification in the M3 verification log.
- **EC-2 — A heuristic cannot fit one line without becoming ambiguous.** Drop to 4 heuristics rather than cross the ≤13-line threshold (AC-SDH-003 binary). AC-SDH-002 permits 3-5; 4 is always safe.
- **EC-3 — The memory system carries a date but the SPEC-ID is ambiguous (multiple SPECs same date).** Cite the date + the SPEC-ID that the memory file itself attributes. Do not pick arbitrarily.
- **EC-4 — An evolvable block is shared across multiple anti-pattern families.** Append one provenance one-liner per family on separate lines inside the block. Each line must independently satisfy AC-SDH-006.
- **EC-5 — moai-foundation-cc/SKILL.md lacks a clean insertion anchor.** Per K-2, M2.1 does a structural read first. If no clean anchor exists, append the Decision Heuristics section at the literal end of body (after the last existing section).
- **EC-6 — The template source (`internal/template/templates/`) is accidentally edited during this SPEC.** This violates C-4. The verifier MUST confirm `git status internal/template/templates/.claude/skills/` shows no changes attributable to this SPEC. If changes are found, they MUST be reverted and the FU-1 follow-up flag reaffirmed.
- **EC-7 — A reviewer argues the heuristics are "too obvious to need a pointer."** Reject the argument. REQ-SDH-004 is non-negotiable; every heuristic carries a pointer. Obviousness is not a substitute for traceability.

---

## §D. Test Commands (M3 gate, verifiable)

```bash
# M3.2 — Decision Heuristics section present in all 4 skills
grep -rc "## Decision Heuristics" \
  .claude/skills/moai-foundation-core/SKILL.md \
  .claude/skills/moai-workflow-spec/SKILL.md \
  .claude/skills/moai-foundation-cc/SKILL.md \
  .claude/skills/moai-meta-harness/SKILL.md
# Expected: 4 hits total (1 per file)

# M3.3 — Line-count per section (binary: ≤13 PASS, ≥14 VIOLATION)
for f in .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md; do
  awk '/^## Decision Heuristics/{flag=1;next} /^## /{flag=0} flag' "$f" | wc -l
done
# Expected: each ≤ 13

# M3.4 — Provenance honesty audit (3 evolvable-bearing skills only; moai-meta-harness excluded — 0 markers)
grep -hoE "AP-[A-Z]+-[0-9]+ — recurred on [0-9]{4}-[0-9]{2}-[0-9]{2}" \
  .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc}/SKILL.md \
  | sort -u
# Each returned date MUST appear in ~/.claude/projects/{hash}/memory/*.md

# M3.5 — Additive-only diff (no - lines except whitespace)
git diff .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md \
  | grep -E '^-' | grep -vE '^---|^-+$|^-\s*$'
# Expected: empty (no content deletions)

# M3.6 — Honest-evolvable N/A for moai-meta-harness (deliverable b skipped, not fabricated)
grep -c 'moai:evolvable-start' .claude/skills/moai-meta-harness/SKILL.md
# Expected: 0 (unchanged from §F A-3 baseline)
grep -c 'moai:evolvable-start' .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc}/SKILL.md
# Expected: 3 / 3 / 3 (unchanged baseline — provenance appended INSIDE existing blocks)

# M3.7 — Template follow-up flag present in plan.md (REQ-SDH-009 / AC-SDH-008)
grep -nE 'FU-1|Template mirror sync|template-source mirror' plan.md
# Expected: ≥1 hit in plan.md §I (documentation-existence check, NOT a run-phase deliverable)

# FL-2 — Evolvable-block balance preserved
for f in .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md; do
  s=$(grep -c "moai:evolvable-start" "$f"); e=$(grep -c "moai:evolvable-end" "$f")
  echo "$f start=$s end=$e"
done
# Expected: start == end per file (unchanged from baseline; moai-meta-harness 0/0)

# EC-6 — Template source untouched
git status internal/template/templates/.claude/skills/
# Expected: no changes attributable to this SPEC
```

---

## §E. Quality Gate Criteria

- **QG-1 (Functional)**: Gate-1 (AC-001/002/003) MUST pass — the Decision Heuristics device is functionally delivered across all 4 skills.
- **QG-2 (Craft)**: Gate-2 (AC-004/007) MUST pass — existing contracts (frontmatter, bodies, evolvable blocks) are preserved.
- **QG-3 (Honesty)**: Gate-3 (AC-006) MUST pass — no fabricated dates; the no-unobserved-claim invariant holds.
- **QG-4 (Provenance coverage)**: Gate-4 (AC-005) SHOULD pass — every memory-mapped row in the 3 evolvable-bearing skills gains a binding, with honest justification for any 0 and explicit N/A for moai-meta-harness.
- **QG-5 (Follow-up flagging)**: Gate-5 (AC-008) MUST pass — template mirror sync flagged as follow-up in plan.md §I, scoped OUT of run-phase.

---

## §F. Residual Risk

- **RR-1** — A future skill-body edit rewrites a section that a Decision Heuristics rule points to, leaving a dangling pointer. Mitigation: FL-1 + the `(<- §<section>)` pointer convention makes the dependency visible at edit time.
- **RR-2** — The memory system's incident history grows; the provenance one-liners become stale. Mitigation: FU-3 periodic refresh (out of scope here); pending-provenance form is honest about coverage gaps.
- **RR-3** — The template mirror sync (FU-1) is delayed, creating a local-vs-template drift window. Mitigation: the drift is additive-only and documented; no user-facing behavior depends on the Decision Heuristics section being in the template at the same time.
- **RR-4** — A heuristic, despite the pointer, is misread by a future agent as the full policy. Mitigation: each section opens with a one-liner stating "Fast defaults — always confirm against the cited body section for non-trivial decisions."

---

Version: 0.2.0 (plan-phase, iter-2 — 5 plan-auditor defects resolved: D1 §H Out of Scope h3; D2 §F A-3 honest 3-of-4 evolvable + REQ-SDH-005/AC-SDH-005 N/A for moai-meta-harness; D3 AC-SDH-008 added for REQ-SDH-009; D4 binary ≤13/≥14 threshold; D5 REQ-SDH-009 system-actor subject)
Status: draft
Last Updated: 2026-06-18
