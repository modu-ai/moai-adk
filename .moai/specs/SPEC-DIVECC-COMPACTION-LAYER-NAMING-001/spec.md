---
id: SPEC-DIVECC-COMPACTION-LAYER-NAMING-001
title: "5-Layer Compaction Naming Cross-Reference (Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact)"
version: "0.1.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "compaction, context-window, runtime-recovery, doc-alignment, provenance, dogfooding, divecc"
era: V3R6
tier: S
---

# SPEC-DIVECC-COMPACTION-LAYER-NAMING-001 — 5-Layer Compaction Naming Cross-Reference

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial plan-phase draft. Candidate N5 of Epic Dive-into-CC. VERIFIED-by-citation premise (paper names the 5 layers explicitly). DOC-ALIGNMENT only — no Go, no behavior change. |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

This SPEC is candidate **N5** of the **Epic Dive-into-CC** (domain token `DIVECC`), a dogfooding exercise applying findings from a reverse-engineering analysis of Claude Code internals to moai-adk's own harness doctrine. The Epic roadmap lives at `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` (§N5 candidate detail).

Source body of work (one publication, two surfaces):

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE).
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository + "Build Your Own AI Agent: A Design Guide".

### A.2 The artifact this SPEC records — the 5 layer names (verbatim from the paper)

The paper names Claude Code's graduated-compaction mechanism as five distinct layers, in escalation order:

```
Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact
```

This SPEC records those five names exactly, as a provenance-enriching **cross-reference** in two existing workflow rule files. It records the *names*; it asserts no behavior about the moai-adk tree.

### A.3 Why this is worth documenting in moai-adk — convergent provenance

`runtime-recovery-doctrine.md` §1 ALREADY names a graduated-compaction layer sequence grounded in `github.com/wquguru/harness-books` book1 ch03:

```
memory prefetch → snip → microcompact → context-collapse → autocompact
```

The VILA-Lab paper is a **convergent SECOND source** describing the same graduated-compaction concept under capitalized names. Recording the paper's names alongside the existing book1 sequence enriches provenance (two independent reverse-engineering sources naming the same mechanism) without changing any behavior. `context-window-management.md`, by contrast, currently has ZERO compaction-layer mention; N5 adds a new cross-reference there naming the five layers as the graduated-compaction vocabulary the orchestrator's `/clear` discipline interacts with.

### A.4 Critical framing boundary — consume, not implement

[ZONE:Frozen-spirit] **moai-adk CONSUMES Claude Code's graduated compaction; it does NOT implement these five layers.** Budget Reduction / Snip / Microcompact / Context Collapse / Auto-Compact are Claude Code runtime internals. moai-adk is a harness ON TOP of Claude Code and cannot modify the native compaction loop (the same constraint `runtime-recovery-doctrine.md` §Policy-layer-only already states for the query loop). Every added cross-reference MUST frame the five layers as "Claude Code's graduated-compaction layers (which moai-adk consumes)" — NEVER imply moai-adk implements them. This is a binary acceptance criterion (AC-CLN-003).

---

## §B. Problem Statement & Grounding

### B.1 Problem

Two adjacent doctrine surfaces describe context-window pressure and recovery without a shared, named vocabulary for Claude Code's graduated-compaction layers:

- `context-window-management.md` describes `/clear` thresholds and stall risk but never names the compaction layers the runtime applies before the ceiling.
- `runtime-recovery-doctrine.md` §1 names a lowercase book1-derived sequence but does not record the convergent VILA-Lab capitalized naming, so a reader cannot see that two independent sources name the same mechanism.

Recording the paper's five layer names in both surfaces gives the doctrine a shared, externally-grounded compaction vocabulary, and makes the book1 ↔ VILA-Lab convergence explicit.

### B.2 Grounding (VERIFIED-by-citation + moai-tree observations this plan-phase)

The premise is VERIFIED-by-citation: the paper names the five layers explicitly. The SPEC records the names; it asserts no behavior about the moai-adk tree, so no moai-tree re-grounding of the *premise* is required (per the Epic Origin note in the ROADMAP, only per-SPEC premises that assert facts about moai-adk's own tree are subject to the verification-claim-integrity invariant; N5 is not such a premise).

The following moai-tree observations were nonetheless reproduced this plan-phase (Read/grep, 2026-06-22), to scope the run-phase edit accurately:

**Observed — `context-window-management.md` (local):**
- File exists at `.claude/rules/moai/workflow/context-window-management.md` (4328 bytes).
- Template mirror EXISTS at `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` (4328 bytes) and is currently **byte-identical** to local (`diff` exit 0).
- The file currently has **ZERO** compaction-layer mention: `grep -niE 'snip|microcompact|compaction|budget reduction|context collapse|auto-compact|book1|VILA|2604.14228|arxiv'` → 0 matches.
- The file is NOT in the byte-parity mirror-test allowlist of `internal/template/rule_template_mirror_test.go` (so `TestRuleTemplateMirrorDrift` does not mechanically gate it — but the mirror IS scanned by the neutrality tests, and AC-CLN-004 still requires the two trees stay identical via `diff`).

**Observed — `runtime-recovery-doctrine.md` (local):**
- File exists at `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` (16985 bytes).
- Template mirror **DOES NOT EXIST**: `find internal/template/templates -name runtime-recovery-doctrine.md` → empty. This file carries internal SPEC IDs (e.g. `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001`, `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001`) and is intentionally local-only / NOT template-distributed. → local-only edit, no mirror, no neutrality constraint.
- §1 line ~20 already names `memory prefetch → snip → microcompact → context-collapse → autocompact`, grounded in book1 ch03.
- `AP-RR-004` mandates the book1 named principles (`withheld-recoverable`, `cheapest-first`, `death-spiral`, `narrative consistency`, `hasAttemptedReactiveCompact`, `truncateHeadForPTLRetry`, `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES`) be preserved verbatim. All seven terms confirmed present this plan-phase.

### B.3 Run-phase file scope (Tier S confirmed)

Three files total:

1. `.claude/rules/moai/workflow/context-window-management.md` (local) — NEW cross-reference naming the 5 layers (consume-not-implement framing + paper citation). Template-distributed → neutrality constraint applies.
2. `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` (mirror) — byte-identical mirror of (1).
3. `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` (local) — ADDITIVE cross-reference adding the VILA-Lab paper as a convergent second source and mapping its capitalized names onto the existing book1 lowercase sequence. Local-only (no mirror), no neutrality constraint, but the AP-RR-004 verbatim terms MUST remain present (additive-only).

3 files, doc-only, < 300 LOC of prose → Tier S confirmed.

---

## §C. Requirements (GEARS)

> Notation: GEARS (current). `<subject>` is generalized.

**REQ-CLN-001 (Ubiquitous)** — The added cross-reference(s) **shall** record the five compaction-layer names verbatim: `Budget Reduction`, `Snip`, `Microcompact`, `Context Collapse`, `Auto-Compact`.

**REQ-CLN-002 (Ubiquitous)** — The added cross-reference(s) **shall** cite the source paper (arXiv:2604.14228 and/or github.com/VILA-Lab/Dive-into-Claude-Code) as the provenance of the five-layer naming.

**REQ-CLN-003 (Ubiquitous)** — The added cross-reference(s) **shall** frame the five layers as Claude Code's graduated-compaction layers that moai-adk **consumes** — explicitly NOT as layers moai-adk implements (consume-not-implement framing).

**REQ-CLN-004 (Ubiquitous)** — After the run-phase edit, the local `context-window-management.md` and its template mirror **shall** be byte-identical (`diff` exits 0).

**REQ-CLN-005 (Where capability gate / neutrality)** — **Where** the edit lands in the template-distributed `context-window-management.md` (local + mirror), the added cross-reference **shall** cite only the PUBLIC paper (arXiv / VILA-Lab) and **shall not** introduce any forbidden internal-content class (no internal `SPEC-DIVECC` ID, no internal date, no commit SHA, no internal-only path) into that file or its mirror.

**REQ-CLN-006 (Ubiquitous)** — The `runtime-recovery-doctrine.md` cross-reference **shall** explicitly note that the VILA-Lab paper is a CONVERGENT second source to book1 ch03 (both naming the same graduated-compaction concept).

**REQ-CLN-007 (Unwanted behavior)** — The run-phase edit **shall not** remove, paraphrase, or weaken any of the AP-RR-004 book1 named principles in `runtime-recovery-doctrine.md` (`withheld-recoverable`, `cheapest-first`, `death-spiral`, `narrative consistency`, `hasAttemptedReactiveCompact`, `truncateHeadForPTLRetry`, `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES`); the cross-reference is ADD-only.

**REQ-CLN-008 (Unwanted behavior)** — The run-phase edit **shall not** modify any file other than the three named in §B.3; **shall not** create a template mirror for `runtime-recovery-doctrine.md`; and **shall not** change any compaction, recovery, or context-window runtime behavior (moai-adk cannot — it consumes Claude Code compaction).

---

## §D. Acceptance Criteria (inline — Tier S)

> Tier S: AC inline in spec.md §D (no separate acceptance.md). Each AC is observable/mechanical (grep / file-existence / diff / test). These ACs bind the **run-phase** outcome; they are NOT satisfied at plan-phase.

**AC-CLN-001 — Five layer names present verbatim** (binds REQ-CLN-001)
- GIVEN the run-phase edits to the two target files
- WHEN `grep -F -e "Budget Reduction" -e "Snip" -e "Microcompact" -e "Context Collapse" -e "Auto-Compact" <file>` is run on the added cross-reference region
- THEN all five literal layer names appear in the added cross-reference of `context-window-management.md` AND the same five capitalized layer names appear in the `runtime-recovery-doctrine.md` mapping (both arms binary-gated by a dedicated for-loop).
- Verification (CWM arm): `for n in "Budget Reduction" "Snip" "Microcompact" "Context Collapse" "Auto-Compact"; do grep -qF "$n" .claude/rules/moai/workflow/context-window-management.md || echo "MISSING: $n"; done` → no output.
- Verification (rrd arm): `for n in "Budget Reduction" "Snip" "Microcompact" "Context Collapse" "Auto-Compact"; do grep -qF "$n" .claude/rules/moai/workflow/runtime-recovery-doctrine.md || echo "MISSING: $n"; done` → no output. (Mirrors the CWM arm so the runtime-recovery-doctrine.md capitalized-name mapping is also binary-gated, not asserted by prose alone.)

**AC-CLN-002 — Paper citation present** (binds REQ-CLN-002)
- GIVEN the added cross-reference(s)
- WHEN `grep -E "2604.14228|VILA-Lab|Dive into Claude Code|Dive-into-Claude-Code" <file>` is run
- THEN the paper is cited in BOTH edited files (`context-window-management.md` AND `runtime-recovery-doctrine.md`).
- Verification: `grep -lE "2604.14228|VILA-Lab" .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/workflow/runtime-recovery-doctrine.md` lists BOTH files.

**AC-CLN-003 — Consume-not-implement framing present** (binds REQ-CLN-003, REQ-CLN-008)
- GIVEN the added cross-reference(s)
- WHEN the added prose is read
- THEN it explicitly frames the five layers as Claude Code's graduated-compaction layers that moai-adk **consumes** (NOT implements), with the consume / not-implement language **co-located with a layer name** (`Budget Reduction` or the phrase `graduated-compaction`).
- Verification (non-vacuous co-location anchor): `grep -niE 'consume[sd]?.{0,80}(Budget Reduction|graduated[- ]compaction)|(Budget Reduction|graduated[- ]compaction).{0,80}(consume[sd]?|does not implement|not implement)' .claude/rules/moai/workflow/context-window-management.md` returns ≥ 1 match in the added region.
- Non-vacuity note: this anchor pattern was confirmed to return **0 matches** against the current (pre-edit) `context-window-management.md` (whose only pre-existing `consume` token — `Trigger #1 consumes the model-specific threshold table` on line 73 — is NOT co-located with any layer name), so the AC can pass ONLY after the run-phase adds the framed cross-reference. The whole-file `grep "consume"` of the prior AC version was vacuous because line 73 already matched it pre-edit; this co-location anchor (mirroring the contrastive pattern of AC-CLN-007) re-anchors the AC onto the SPEC's central consume-not-implement hazard.

**AC-CLN-004 — CWM local ↔ mirror byte-identical** (binds REQ-CLN-004)
- GIVEN `context-window-management.md` was edited locally
- WHEN `diff .claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` is run
- THEN the diff is empty (exit 0) — the mirror received the identical edit.
- Verification: `diff <local> <mirror>; echo "exit=$?"` → `exit=0`.

**AC-CLN-005 — Template neutrality preserved for the mirrored file** (binds REQ-CLN-005, template isolation doctrine)
- GIVEN the `context-window-management.md` edit was mirrored
- WHEN the neutrality gate runs
- THEN the mirrored file contains no forbidden internal-content class (no internal `SPEC-DIVECC` token, no internal date, no commit SHA, no internal-only path) — the five layer names + public paper citation + consume framing are all acceptable content class.
- Verification: `go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak'` passes (run-phase gate) AND `grep -E "SPEC-DIVECC|2026-06-22" internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` returns NO match.
- Coverage note: the `go test` arm is the **authoritative** neutrality gate — it covers the full forbidden-content class set REQ-CLN-005 enumerates (internal `SPEC-DIVECC` ID, internal date, **commit SHA, internal-only path**). The inline `grep -E "SPEC-DIVECC|2026-06-22"` arm is an **illustrative spot-check** of only 2 of those 4 classes (it omits the commit-SHA and internal-only-path classes); it is a convenience signal, NOT the SSOT. When the inline grep and the `go test` arm disagree, the `go test` arm wins.

**AC-CLN-006 — book1 named principles preserved verbatim (additive-only proof)** (binds REQ-CLN-007)
- GIVEN the `runtime-recovery-doctrine.md` edit
- WHEN `grep -F` is run for each AP-RR-004 named term after the edit
- THEN all seven terms (`withheld-recoverable`, `cheapest-first`, `death-spiral`, `narrative consistency`, `hasAttemptedReactiveCompact`, `truncateHeadForPTLRetry`, `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES`) remain present.
- Verification: `for t in "withheld-recoverable" "cheapest-first" "death-spiral" "narrative consistency" "hasAttemptedReactiveCompact" "truncateHeadForPTLRetry" "MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES"; do grep -qF "$t" .claude/rules/moai/workflow/runtime-recovery-doctrine.md || echo "REGRESSED: $t"; done` → no output.

**AC-CLN-007 — Convergent-second-source note present in runtime-recovery-doctrine.md** (binds REQ-CLN-006)
- GIVEN the added `runtime-recovery-doctrine.md` cross-reference
- WHEN the added prose is read
- THEN it explicitly states the VILA-Lab paper is a convergent / second source to book1 ch03 naming the same graduated-compaction concept.
- Verification: `grep -niE "convergent|second source|same (graduated[- ])?compaction" .claude/rules/moai/workflow/runtime-recovery-doctrine.md` returns ≥ 1 match in the added region, AND that region also references both `book1` and the VILA-Lab paper.

**AC-CLN-008 — No out-of-scope file change; no rrd mirror created** (binds REQ-CLN-008)
- GIVEN the run-phase commit diff
- WHEN `git show --stat <run-commit>` is inspected
- THEN the only changed files are the three named in §B.3 (CWM local + CWM mirror + runtime-recovery-doctrine.md local, plus the spec.md frontmatter `status` transition); no `internal/template/templates/.claude/rules/moai/workflow/runtime-recovery-doctrine.md` is created; no hook script, skill body, plugin, MCP config, or Go source is touched.
- Verification: `git show --stat <run-commit>` file list ⊆ {CWM local, CWM mirror, runtime-recovery-doctrine.md local, spec.md} AND `ls internal/template/templates/.claude/rules/moai/workflow/runtime-recovery-doctrine.md` returns "No such file".

---

## §E. Lifecycle Progress Markers

> Plan-phase emits the §E section skeleton (placeholder headings only). §E.2–§E.4 are populated by manager-develop (run) and manager-docs (sync), not by this plan-phase author. See progress.md for the canonical skeleton.

- **§E.1 Plan-phase Audit-Ready Signal** — see progress.md §E.1.
- §E.2 Run-phase Evidence — _pending run-phase_.
- §E.3 Run-phase Audit-Ready Signal — _pending run-phase_.
- §E.4 Sync-phase Audit-Ready Signal — _pending sync-phase_.

---

## §F. Out of Scope

This section bounds what this SPEC does NOT cover (satisfies `OutOfScopeRule` / avoids `MissingExclusions`).

### Out of Scope — run-phase implementation

- Authoring the cross-reference text itself into the two target files. This SPEC is plan-phase only; the rule edits happen at run-phase against the AC matrix in §D.
- Any Go code change. There is none — this is a doc-alignment SPEC.

### Out of Scope — compaction / recovery behavior change

- Changing any compaction, recovery, or context-window runtime behavior. moai-adk CONSUMES Claude Code's graduated compaction (Budget Reduction / Snip / Microcompact / Context Collapse / Auto-Compact); it cannot implement or alter those layers (REQ-CLN-008). The cross-reference is a *provenance record*, not a mechanism change.

### Out of Scope — runtime-recovery-doctrine.md template mirroring

- Creating a template mirror for `runtime-recovery-doctrine.md`. That file is intentionally NOT template-distributed (it carries internal SPEC IDs and is local-only / dev-only). The run-phase edit to it is local-only; producing a mirror is explicitly out of scope (REQ-CLN-008).

### Out of Scope — editing any rule file other than the two named

- Editing any rule, agent, skill, hook, or config file other than `context-window-management.md` (+ its mirror) and `runtime-recovery-doctrine.md`. No other surface is in scope.

### Out of Scope — layer-name verification / behavior assertion

- Independently verifying, benchmarking, or asserting any behavior of the five compaction layers in the moai-adk tree. The five names are recorded as the paper's naming (arXiv:2604.14228), a provenance citation — not a moai-adk measurement or behavioral claim (REQ-CLN-003 + the verification-claim-integrity invariant).

### Out of Scope — other Epic Dive-into-CC candidates

- N1 (hook failure-mode audit, closed), N2 (extension-cost ladder, closed), N3 (delegation token-cost, closed), N4 (observability loop), N6 (unified inventory), N7 (paper archival). Each is its own SPEC; this SPEC touches none of their surfaces.

---

## §G. Cross-References

- **arXiv:2604.14228** — "Dive into Claude Code" (the source of the five-layer compaction naming).
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository.
- `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` — Epic Dive-into-CC roadmap (§N5 candidate detail).
- `.claude/rules/moai/workflow/context-window-management.md` — run-phase target #1 (+ its template mirror) — NEW compaction-layer cross-reference.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` — run-phase target #2 (local-only) — convergent-second-source cross-reference; §1 already names the book1 lowercase sequence; AP-RR-004 verbatim-term preservation is a run-phase constraint.
- `github.com/wquguru/harness-books` book1 ch03 — the EXISTING first source for the graduated-compaction sequence (the VILA-Lab paper is the convergent second source N5 records).
- `.moai/docs/template-internal-isolation-doctrine.md` — template neutrality gate for the CWM mirror edit (AC-CLN-005).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — provenance discipline (paper-naming citation vs moai-tree behavioral claim).
