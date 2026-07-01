---
id: SPEC-INVOCATION-MODEL-002
title: "Invocation-model divergence reconciliation + Axis-A alignment (review compose; clean scoped out)"
version: "0.1.0"
status: in-progress
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows, .claude/rules/moai/workflow, .moai/specs"
lifecycle: spec-anchored
tags: "invocation-model, doctrine, axis-a, review, errata, template-first"
tier: S
---

# SPEC-INVOCATION-MODEL-002 — Invocation-model divergence reconciliation + Axis-A alignment

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-01 | 0.1.0 | manager-spec | Initial draft — successor to SPEC-INVOCATION-MODEL-001. Two deliverables: (1) reconcile the closed SPEC's outdated §A classification via an errata pointer + authoritative record; (2) the deferred Axis-A alignment — review↔/code-review compose; clean↔/simplify merit-checked and scoped out. |

---

## §A. Context and Motivation (WHY)

SPEC-INVOCATION-MODEL-001 (completed, era V3R6) established the policy-layer doctrine `.claude/rules/moai/workflow/native-invocation-model.md`, which classifies native Claude Code commands on the **invocation-model** axis: PROGRAMMATIC (a bundled skill/workflow, OR a built-in exposed via the `Skill` tool — orchestrator-auto-invocable via `Skill()` / `Workflow()`) vs HUMAN-ONLY (a built-in command with no `Skill`-tool bridge — human-typed only).

During SPEC-INVOCATION-MODEL-001's run-phase, a classification divergence was discovered and **the rule file was corrected**, but the closed `spec.md` §A examples were NOT updated (SPEC bodies are immutable post-completion). Two residues remain:

1. **Divergence residue.** The closed `SPEC-INVOCATION-MODEL-001/spec.md` §A (lines ~31-41) still lists `/security-review` and `/review` as HUMAN-ONLY. This is OUTDATED. The official `skills.md` states verbatim: *"A few built-in commands are also available through the Skill tool, including `/init`, `/review`, and `/security-review`."* Both are therefore PROGRAMMATIC (built-in-exposed-via-`Skill`-tool sub-case). The rule file `native-invocation-model.md` ALREADY carries the corrected matrix and a "Classification-divergence note"; only the closed SPEC body is stale.
2. **Deferred Axis-A refactoring.** The closed SPEC deferred the Axis-A bundled-skill reuse refactoring (pointing `/moai clean` at native `/simplify`, `/moai review` at native `/code-review`) to a follow-up SPEC (its §E Exclusions "Out of Scope — Axis-A bundled-skill reuse refactoring").

This SPEC resolves BOTH residues, under a confirmed user decision (2026-07-01, via AskUserQuestion) that divergence-handling = **Option A** (errata pointer + authoritative record; immutable body preserved).

### Authoritative classification record (corrected 9-command matrix)

This SPEC records the corrected classification as the **authoritative record** (superseding the closed SPEC's §A example prose). The rule file matrix is the operational SSOT; this table restates it for traceability.

| Native command | Category | Basis |
|----------------|----------|-------|
| `/code-review` | PROGRAMMATIC | `[Skill]` marker (commands reference) |
| `/simplify` | PROGRAMMATIC | `[Skill]` marker (commands reference) |
| `/loop` | PROGRAMMATIC | `[Skill]` marker; self-paces when interval omitted |
| `/deep-research` | PROGRAMMATIC | `[Workflow]` marker (commands reference) |
| `/security-review` | PROGRAMMATIC | built-in exposed via the `Skill` tool (`skills.md`) |
| `/review` | PROGRAMMATIC | built-in exposed via the `Skill` tool (`skills.md`) |
| `/goal` | HUMAN-ONLY | no marker, not `Skill`-tool-exposed |
| `/clear` | HUMAN-ONLY | no marker, not `Skill`-tool-exposed |
| `/compact` | HUMAN-ONLY | no marker, not `Skill`-tool-exposed |

**Count: 6 PROGRAMMATIC / 3 HUMAN-ONLY.** The only genuinely HUMAN-ONLY native commands are `/goal`, `/clear`, `/compact`.

### Axis-A merit findings (grounded in the actual skill bodies)

Axis A of the doctrine prefers **reusing a native PROGRAMMATIC command via `Skill()` over reimplementing it**. The doctrine named `/moai clean` and `/moai review` as the ONLY two deferred Axis-A candidates. This SPEC merit-checks each — the candidacy is NOT assumed valid:

- **`/moai clean` ↔ native `/simplify` — capability mismatch, SCOPED OUT.** `/moai clean` (`.claude/skills/moai/workflows/clean.md`) is whole-tree **dead-code detection** (static analysis + usage-graph traversal) followed by **safe deletion** with test-verified rollback and `@MX:ANCHOR` protection. Native `/simplify` reviews the **changed** code for reuse / simplification / efficiency / altitude cleanups and **applies quality fixes** — quality only, no dead-code usage-graph model, no whole-tree deletion. These are DIFFERENT capabilities: `/moai clean` does not reinvent `/simplify`; they are complementary (parallel to how `/moai loop` is distinct from native `/loop`). Native `/simplify`'s actual capability maps closer to the EXISTING `/moai review --lean` over-engineering audit (`delete:`/`stdlib:`/`native:`/`yagni:`/`shrink:` tags) than to `/moai clean`. Forcing a swap would replace a safe dead-code-removal pipeline with a quality-refactor pass that does not detect dead code at all. **Verdict: scoped out; no `clean.md` edit (local or template).**
- **`/moai review` ↔ native `/code-review` — COMPOSE (not swap).** Native `/code-review` reviews the current diff for **correctness bugs + reuse/simplification/efficiency cleanups**. `/moai review` (`.claude/skills/moai/workflows/review.md`) is a COMPOSITION the native single-purpose command does not provide: Security (Perspective 1: OWASP + dependency-vuln scan + full-git-history secrets scan + data-isolation) + Performance (P2) + Quality (P3) + UX (P4) + `@MX` tag-compliance (Phase 3) + design review (Phase 4.5) + sync-auditor skeptical 4-dimension synthesis. Native `/code-review` overlaps the correctness/quality-cleanup portion (P2+P3). **Verdict: compose — native `/code-review` becomes ONE Skill()-invoked component covering the correctness-bug + reuse/simplification/efficiency portion, while the Security / `@MX` / UX / design composition is PRESERVED (compose-not-swap).**

---

## §B. Scope (WHAT)

Two deliverables. Both are policy-alignment / documentation edits; NO Go runtime, NO hook, NO lint rule (the doctrine is codification-only).

### Deliverable 1 — Divergence reconciliation (errata pointer + authoritative record)

The closed `SPEC-INVOCATION-MODEL-001/spec.md` §A body stays **immutable**; the corrected classification is recorded as the authoritative record in THIS SPEC's §A + §C. Exactly ONE line — an errata pointer — is appended to the very bottom of the closed `spec.md` (after §F). This single line is the ONLY permitted edit to the immutable closed body — a documented, intentional exception, NOT a content correction.

### Deliverable 2 — Axis-A alignment (review compose; clean scoped out)

- `/moai review`: native `/code-review` is composed into the Phase 2 pipeline via `Skill()` as ONE correctness/quality-cleanup component, with the Security / `@MX` / UX / design composition preserved and a conditional-PROGRAMMATIC runtime verification + fallback path. Local + template mirror both edited; `make build` recompiles.
- `/moai clean`: merit-checked, found a capability mismatch, and recorded as scoped-OUT with rationale. No `clean.md` edit.

The rule file `native-invocation-model.md` needs no classification edit (already correct); an optional SPEC-ID cross-ref into it is scoped OUT (see §E — it would breach template neutrality or Template-First parity).

---

## §C. GEARS Requirements

The generalized `<subject>` (any noun: SPEC body, skill, orchestrator, workflow) is used per GEARS.

### Deliverable 1 — Divergence reconciliation

**REQ-IM2-001** (Ubiquitous) — This SPEC's `spec.md` §A + §C **shall** record the corrected 9-command classification as the authoritative record: 6 PROGRAMMATIC (`/code-review`, `/simplify`, `/loop`, `/deep-research`, `/security-review`, `/review`) + 3 HUMAN-ONLY (`/goal`, `/clear`, `/compact`), with `/security-review` and `/review` classified PROGRAMMATIC (built-in exposed via the `Skill` tool).

**REQ-IM2-002** (Event-driven) — **When** this SPEC is run-phase implemented, exactly one errata-pointer line **shall** be appended to the bottom of the closed `.moai/specs/SPEC-INVOCATION-MODEL-001/spec.md`, naming `/security-review` and `/review` as PROGRAMMATIC and pointing to `native-invocation-model.md` + this follow-up SPEC. This single line **shall** be the ONLY edit to the closed body.

**REQ-IM2-003** (Unwanted behavior) — The reconciliation **shall not** rewrite, correct, or otherwise modify the closed `spec.md` §A classification prose; the immutable body's HUMAN-ONLY classification for `/security-review` and `/review` is superseded ONLY via the appended errata pointer, never by an in-place edit.

**REQ-IM2-004** (Capability gate) — **Where** a cross-reference to this follow-up SPEC would be added to `native-invocation-model.md`, the rule file's classification matrix **shall not** be edited (it is already correct); the SPEC-ID cross-ref is scoped OUT (§E) to preserve template neutrality and Template-First parity.

### Deliverable 2 — Axis-A alignment

**REQ-IM2-005** (Ubiquitous) — The `clean` ↔ `/simplify` Axis-A mapping, having been merit-checked and found a capability mismatch (dead-code removal ≠ changed-code quality refactoring), **shall** be recorded as scoped-OUT with rationale in §E; no edit to `clean.md` (local or template) **shall** be made.

**REQ-IM2-006** (Event-driven) — **When** the `review` ↔ `/code-review` compose is implemented, native `/code-review` **shall** be invoked via `Skill()` as ONE component of `/moai review`'s Phase 2 pipeline for the correctness-bug + reuse/simplification/efficiency portion, while the Security pass (Perspective 1), `@MX` tag-compliance (Phase 3), UX (Perspective 4), design review (Phase 4.5), and sync-auditor skeptical synthesis composition **shall** be PRESERVED (compose-not-swap).

**REQ-IM2-007** (Capability gate) — **Where** native `/code-review` is not auto-invocable (a bundled skill with `disable-model-invocation: true`, a session with `disableBundledSkills`, or a denied `Skill` tool per the conditional-PROGRAMMATIC caveat), the review workflow **shall** verify auto-invocability at runtime before relying on `Skill()` invocation and **shall** fall back to the existing sync-auditor Phase 2 path.

**REQ-IM2-008** (Unwanted behavior) — The review compose **shall not** remove or weaken the Security, `@MX`, UX, or design-review composition; native `/code-review` augments, never replaces, the MoAI-specific perspectives.

### Cross-cutting constraints

**REQ-IM2-009** (Ubiquitous) — Every edit to the local `.claude/skills/moai/workflows/review.md` **shall** be mirrored to `internal/template/templates/.claude/skills/moai/workflows/review.md`, and `make build` **shall** regenerate the embedded templates.

**REQ-IM2-010** (Unwanted behavior) — This SPEC **shall not** introduce any hook, lint rule, or runtime mechanism to enforce the invocation-model classification; the doctrine remains policy-layer codification only.

**REQ-IM2-011** (Ubiquitous) — The `review.md` compose edits, as template-distributed assets, **shall** remain language-neutral across the 16 supported languages and free of internal-development leakage (no SPEC IDs, REQ tokens, ISO dates, or commit SHAs per CLAUDE.local.md §25). Native command names (`/code-review`) and permanent rule citations (`native-invocation-model.md`) are allowed content classes; this SPEC's own ID **shall not** appear in the template `review.md`.

---

## §D. Acceptance Criteria Summary

Full testable AC enumeration lives in `acceptance.md`. Traceability: REQ-IM2-001..004 → divergence reconciliation (authoritative-record grep + single-errata-line + immutability diff + rule-file-untouched); REQ-IM2-005..008 → Axis-A (clean-untouched + review-compose presence + conditional-caveat/fallback + composition-preserved); REQ-IM2-009..011 → local↔template parity + no-runtime-mechanism + neutrality grep.

---

## §E. Exclusions (What NOT to Build)

This section is load-bearing: it defines what is explicitly out of scope so run-phase does not drift.

### Out of Scope — clean ↔ /simplify Axis-A alignment (capability mismatch)

- `/moai clean` is NOT re-pointed at native `/simplify`. Merit-check finding: `/moai clean` is whole-tree dead-code detection + usage-graph + safe deletion + test-verified rollback + `@MX:ANCHOR` protection; native `/simplify` is changed-code quality refactoring (reuse/simplification/efficiency/altitude) that applies fixes without dead-code usage-graph analysis. These are different capabilities — `/moai clean` does not reinvent `/simplify`. `/simplify`'s capability maps closer to the existing `/moai review --lean` over-engineering audit than to `/moai clean`.
- No edit to `clean.md` (local `.claude/skills/moai/workflows/clean.md` or template mirror) is made by this SPEC.

### Out of Scope — native-invocation-model.md SPEC-ID cross-reference

- The optional §Cross-References entry naming this SPEC ID inside `native-invocation-model.md` is scoped out. The rule file is template-mirrored; a SPEC-ID cross-ref would either leak an internal SPEC ID into the distributed template (CLAUDE.local.md §25 neutrality violation) OR create local↔template drift (Template-First parity violation). The errata pointer in the closed `SPEC-INVOCATION-MODEL-001/spec.md` (a local, non-template `.moai/specs/` artifact) plus this SPEC's own body serve as the authoritative pointer without touching the neutral rule file.

### Out of Scope — closed SPEC body content correction

- The closed `SPEC-INVOCATION-MODEL-001/spec.md` §A classification prose is NOT rewritten. The immutable body is preserved verbatim; only a single errata-pointer line is appended at the file end.

### Out of Scope — runtime enforcement of the doctrine

- No hook, lint rule, or runtime mechanism is added. The invocation-model classification and the conditional-PROGRAMMATIC verification remain agent obligations (policy-layer codification), not mechanically-enforced gates.

### Out of Scope — other subcommand refactoring

- Only `/moai clean` and `/moai review` are considered (the doctrine's two named Axis-A candidates). No other MoAI subcommand is retired, re-pointed, or refactored. `/moai loop` remains distinct from native `/loop` (complementary, not a duplicate).

### Out of Scope — review workflow feature changes

- No new review flags, no change to the `--lean` / `--design` / `--critique` / `--team` / `--security` behavior, no change to the sync-auditor scoring model beyond composing native `/code-review` as one correctness/quality component in Phase 2.

---

## §F. Cross-References

- `.claude/rules/moai/workflow/native-invocation-model.md` — the invocation-model doctrine (already correct); the SSOT for the classification matrix, Axis A/B thesis, and conditional-PROGRAMMATIC caveat.
- `.moai/specs/SPEC-INVOCATION-MODEL-001/spec.md` — the predecessor SPEC whose §A carries the outdated HUMAN-ONLY classification; receives the single errata-pointer line (Deliverable 1).
- `.claude/skills/moai/workflows/review.md` (+ `internal/template/templates/` mirror) — the `/moai review` workflow receiving the native `/code-review` compose (Deliverable 2).
- `.claude/skills/moai/workflows/clean.md` — the `/moai clean` workflow; UNTOUCHED (clean↔/simplify scoped out).
- `CLAUDE.local.md` §2 (Template-First Rule), §15 (16-language neutrality), §25 (Template Internal-Content Isolation) — the mirror + neutrality constraints binding the `review.md` template edit.
