---
id: SPEC-RULE-DIET-002
title: "Rule diet: scope 6 reference-doctrine rules out of the always-loaded context surface"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai + internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "rule-diet, always-loaded, context-budget, paths-scoping, template-first, steering-align"
tier: M
era: V3R6
---

## HISTORY

- 2026-07-02 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier M; spec.md + plan.md + acceptance.md + progress.md §E.1). Continuation of SPEC-STEERING-ALIGN-RULE-SCOPING-001 (which reduced LIVE always-loaded rules 15→11). This SPEC re-audits the remaining 11 against a stricter per-turn-load-bearing test and scopes 6 reference-doctrine rules out of the always-loaded surface (LIVE 11→5, TEMPLATE 11→5). `status: draft`.

---

## A. Context / Background

### A.1 Problem

The always-loaded context surface — the bytes Claude Code re-injects into the orchestrator's context on **every** turn — includes all `.claude/rules/moai/**/*.md` files whose YAML frontmatter carries **no `paths:` restriction**. A rule with `paths:` loads only when Claude touches a matching file; a rule without `paths:` is always-loaded.

Ground-truth measurement (re-verified live 2026-07-02, commands + output in acceptance.md AC-RD2-001):

| Tree | Total rule files | Always-loaded (no `paths:`) |
|------|------------------|------------------------------|
| LIVE (`.claude/rules/moai/`) | 61 | **11** |
| TEMPLATE (`internal/template/templates/.claude/rules/moai/`) | 61 | **11** |

The 11 always-loaded rule files sum to **176,405 bytes ≈ 172.3 KB** (≈ 44,101 tokens by the `char/4` estimate the in-repo guard uses). This is a per-turn tax paid on turns that never need most of it.

Anthropic's official "Steering Claude Code" guidance is the governing principle (also cited by the predecessor SPEC): *"If a rule only applies to `src/api/**`, scoping it with `paths:` keeps it out of context during unrelated work"* and the best-practices per-line test *"Would removing this cause Claude to make mistakes? If not, cut it."*

### A.2 The make-or-break nuance (LOAD-BEARING vs reference doctrine)

The predecessor SPEC (STEERING-ALIGN-RULE-SCOPING-001) reduced LIVE always-loaded rules 15→11 by scoping rules that had a **clean file-touch trigger** (Class A) and excluding a legal-attribution file (Class D). It deliberately left 11 rules always-loaded, split into Class B (env/intent-triggered, "scoping infeasible") and Class C ("genuinely cross-cutting"). One Class-B rule (`glm-web-tooling.md`) was nonetheless later scoped using a **self-referential `paths:` glob** + on-demand hook delivery — establishing the precedent this SPEC generalizes.

This SPEC re-audits the 11 against a stricter test than "has a file-touch trigger":

> **The per-turn-load-bearing test**: Does the orchestrator (or an agent it spawns) need this rule's content on turns where the rule's own subject is NOT the active concern? Concretely — is the rule's actionable content (a) referenced by CLAUDE.md / the output-style as an **every-turn obligation**, or (b) depended upon by another KEEP rule for **inline values it does not itself carry**? If YES → **KEEP always-loaded**. If the rule is consulted only in a **specific context** (recovery event, `/goal` usage, workflow fan-out, subcommand audit, a verification action) and its retrieval survives via a pointer in a KEEP surface → **SCOPE**.

Over-scoping a load-bearing rule silently drops it from the orchestrator's context and breaks per-turn behavior (AskUserQuestion discipline, session-handoff triggers, agent-spawn boundaries, verification-claim truthfulness, `/clear` thresholds). **The anti-goal is over-scoping into a load-bearing rule.** The value delivered is the SAFE subset only.

### A.3 Per-file classification (evidence-based)

Evidence columns: **CC** = reference count in `CLAUDE.md`; **OS** = reference count in `.claude/output-styles/moai/moai.md`; **Tree** = present in both trees (MIRRORED) or live-only. **Nature of reference** distinguishes an SSOT/every-turn obligation from a "See X" pointer.

| # | File | KB | CC | OS | Tree | Self-declared loading scope | Verdict |
|---|------|----|----|----|------|-----------------------------|---------|
| 1 | `workflow/session-handoff.md` | 32.0 | 2 | 10 | MIRRORED | "**Intentionally always-loaded** (no `paths:` restriction) because Trigger #3 (user explicit session-end) can fire from any session context" | **KEEP** |
| 2 | `core/agent-common-protocol.md` | 28.0 | 2 (SSOT) | 0 | MIRRORED | "Shared protocol for all MoAI agent definitions … automatically loaded for all agents" | **KEEP** |
| 3 | `core/askuser-protocol.md` | 25.3 | 2 (SSOT) | 4 | MIRRORED | SSOT for the AskUserQuestion channel monopoly (CLAUDE.md §8 [Frozen][HARD]) | **KEEP** |
| 4 | `core/verification-claim-integrity.md` | 9.7 | 0 | 0 | MIRRORED | "automatically loaded for the orchestrator and all agents" — governs every completion/verification claim | **KEEP** |
| 5 | `workflow/context-window-management.md` | 5.4 | 1 (SSOT) | 1 | MIRRORED | SSOT for the numeric `/clear` thresholds that session-handoff.md (KEEP) delegates to it | **KEEP** |
| 6 | `workflow/runtime-recovery-doctrine.md` | 18.4 | 0 | 0 | MIRRORED | "consult this doctrine … mid-turn" **only on a withheld-recoverable error event** | **SCOPE** (mitigated) |
| 7 | `workflow/dynamic-workflows.md` | 17.4 | 3 (pointers) | 0 | MIRRORED | "Read **when deciding how to fan out** a large task … or when a user asks for a 'workflow'" | **SCOPE** |
| 8 | `workflow/native-invocation-model.md` | 13.0 | 0 | 0 | MIRRORED | "Read **when deciding whether a `/moai` subcommand is justified** … codification only — no runtime mechanism enforces" | **SCOPE** |
| 9 | `development/sprint-round-naming.md` | 11.6 | 0 | 1 | MIRRORED | Epic naming taxonomy SSOT (legacy-alias table + anti-patterns + migration) | **SCOPE** (mitigated) |
| 10 | `workflow/goal-directive.md` | 7.6 | 1 (pointer) | 0 | MIRRORED | "Read **when a user sets a `/goal`**" | **SCOPE** |
| 11 | `workflow/verification-batch-pattern.md` | 3.9 | 0 | 0 | MIRRORED | grouping rationale; the HARD batching obligation + 7-item example live in `agent-common-protocol.md` (KEEP) | **SCOPE** |

**KEEP total: 5 files, 102,939 bytes ≈ 100.5 KB.** **SCOPE total: 6 files, 73,466 bytes ≈ 71.7 KB.**

#### A.3.1 KEEP evidence (why each stays always-loaded)

- **session-handoff.md** — self-declares always-loaded because Trigger #3 (session end) fires from any context. It carries the per-turn handoff triggers (context threshold, PR success, phase completion) the orchestrator evaluates continuously, and it is a bidirectional render-surface partner of the output-style (10 OS refs). Removing it would drop the handoff triggers from every turn.
- **agent-common-protocol.md** — governs **every `Agent()` spawn**: the [HARD] Pre-Spawn Sync Check (multi-session race mitigation), the [HARD] read-only verification batching obligation + the canonical 7-item batch, the User Interaction Boundary (subagent prohibitions), Ledger Closure, and the Error Recovery pattern. CLAUDE.md cites it as SSOT for error recovery and the subagent boundary. These are per-turn obligations on any turn that spawns an agent or verifies completion.
- **askuser-protocol.md** — the SSOT for the AskUserQuestion channel monopoly, a Frozen HARD rule (CLAUDE.md §8) governing **every** user-facing question and the deferred-tool preload sequence. Absence would let the orchestrator drift to free-form prose questions.
- **verification-claim-integrity.md** — self-declared "automatically loaded for the orchestrator and all agents"; binds the truthfulness of **every** orchestrator Completion Report / Verification Matrix banner (the "no unobserved-claim" invariant, Behavior #6). 0 refs, but load-bearing by its self-declaration + moai-constitution Behavior #6.
- **context-window-management.md** — **hard dependency**: session-handoff.md (KEEP) Trigger #1 states verbatim *"this file carries no inline model-class numbers to avoid label drift"* and delegates the numeric `/clear` thresholds (1M=50%, 200K=90%) to context-window-management.md as the authoritative SSOT. Scoping it would strip those numbers from the always-loaded surface entirely; the orchestrator monitors context usage every turn and must know the threshold. This is the classic over-scoping trap and the reason it is KEEP despite the task-provided list omitting it.

#### A.3.2 SCOPE evidence (why each is safe to remove from always-load)

- **dynamic-workflows.md** — reference doctrine consulted only when fanning out. CLAUDE.md §15 already carries the summary + three "See" pointers; the trigger is intent-based ("user asks for a workflow"), not per-turn.
- **native-invocation-model.md** — the clearest case: 0 refs anywhere, self-declared narrow trigger ("when justifying/adding/auditing a subcommand"), and self-declared "codification only — no runtime mechanism enforces."
- **goal-directive.md** — consulted only for `/goal`; the per-turn actionable content (the Block 1 conditional `/goal` re-set line) already lives inline in session-handoff.md (KEEP); this file is the detailed backup.
- **verification-batch-pattern.md** — the [HARD] batching obligation and the canonical 7-item batch live in agent-common-protocol.md (KEEP); this file self-declares it "owns only the *why* (grouping rationale + class taxonomy)". Consulted only at verification time, and even then the actionable batch is in the KEEP file.
- **runtime-recovery-doctrine.md** (mitigated) — 0 refs; consulted only on a withheld-recoverable error event, not per-turn. The recovery **entry-point pointer already exists in an always-loaded KEEP surface**: agent-common-protocol.md § Recovery-Signal Carve-Out names it as SSOT (`runtime-recovery-doctrine.md §4`). The **actionable** recovery is also preserved in KEEP files: the ladder rungs 2-3 (paste-ready resume + `/clear`, worktree restart) live in session-handoff.md, and the `/clear` discipline in context-window-management.md. The full doctrine (circuit-breaker invariants, book1 provenance, 4-rung ladder codification) is on-demand reference. Mitigation requirement REQ-RD2-004 makes the pointer survival a hard precondition.
- **sprint-round-naming.md** (mitigated) — the naming taxonomy applies to SPEC bodies, memory, and orchestrator output, but the **per-turn** need is only the canonical 4-term set (Epic / SPEC / Milestone / Constitution) + the retired-alias warning; the full 12 KB doctrine (legacy-alias table, anti-patterns AP-SRN-001..005, migration checklist, localization table) is on-demand reference. The output-style (KEEP) already uses Epic / SPEC / Milestone inline in the Epic Stats / Epic Status banners and § Banner Localization cross-references this file. Mitigation requirement REQ-RD2-005 makes a compact 4-term + retired-alias summary in an always-loaded surface a hard precondition.

### A.4 Epic Steering-Align context

This SPEC is a follow-up in the same rule-diet lineage as Epic Steering-Align (RULE-SCOPING-001, CLAUDEMD-DIET-001, GUARDRAIL-HOOK-001, OUTPUT-STYLE-SLIM-001, LOCAL-DIET-001). Those SPECs dieted CLAUDE.md, the output-style, the guardrail hooks, and the maintainer-local files. This SPEC continues the `.claude/rules/moai/**` rule-surface diet that RULE-SCOPING-001 began (15→11) by taking it 11→5. It reuses RULE-SCOPING-001's mechanisms verbatim: frontmatter-only `paths:` additions, MIRRORED template-first + `make build` + byte-identical parity.

---

## B. Requirements (GEARS notation)

### B.1 Scope the 6 reference-doctrine rules

- **REQ-RD2-001 (Ubiquitous)**: The rule-diet change SHALL add a `paths:` frontmatter field to each of the 6 SCOPE-classified rules (`runtime-recovery-doctrine.md`, `dynamic-workflows.md`, `native-invocation-model.md`, `sprint-round-naming.md`, `goal-directive.md`, `verification-batch-pattern.md`), using the comma-separated quoted-glob syntax already established in-repo, so that each is excluded from the always-loaded context surface.

- **REQ-RD2-002 (Ubiquitous)**: Each scoped rule SHALL use a **self-referential** `paths:` glob of the form `"**/<filename>.md"` (loading the rule only when the rule file itself is the edit target), matching the `glm-web-tooling.md` (`paths: "**/glm-web-tooling.md"`) precedent. This uniform mechanism preserves every existing cross-reference (which points to the `.claude/rules/moai/...` path) without rewriting any pointer.

- **REQ-RD2-003 (Event-driven)**: **When** Claude touches a file matching a scoped rule's `paths` glob (i.e. the rule file itself, during rule maintenance), the rule SHALL load into context — behavior is preserved for maintenance of the rule; only its always-on residency is removed.

### B.2 Preserve retrieval for the two consult-obligation rules (mitigation)

- **REQ-RD2-004 (Where runtime-recovery-doctrine)**: **Where** the scoped rule is `runtime-recovery-doctrine.md`, an always-loaded KEEP surface SHALL retain (a) the recovery entry-point pointer naming the doctrine as SSOT, and (b) the actionable recovery ladder rungs and `/clear` discipline. The change MUST verify that `agent-common-protocol.md § Recovery-Signal Carve-Out` (pointer), `session-handoff.md` (ladder rungs 2-3), and `context-window-management.md` (`/clear` discipline) already carry this content before scoping; if any is absent it MUST be added as a one-line pointer, never left to the scoped doctrine alone.

- **REQ-RD2-005 (Where sprint-round-naming)**: **Where** the scoped rule is `sprint-round-naming.md`, an always-loaded KEEP surface SHALL retain the canonical 4-term set (Epic / SPEC / Milestone / Constitution) and the retired-alias note (Sprint / Round / Wave / cohort retired) so orchestrator banner output does not drift to legacy terms. The change MUST verify this compact summary is present in `.claude/output-styles/moai/moai.md § Banner Localization`; if absent it MUST be added as a one-line note before scoping.

### B.3 Do NOT scope the 5 KEEP rules (anti-goal guard)

- **REQ-RD2-006 (Unwanted)**: The change SHALL NOT add a `paths:` field to any of the 5 KEEP rules (`session-handoff.md`, `agent-common-protocol.md`, `askuser-protocol.md`, `verification-claim-integrity.md`, `context-window-management.md`). These govern per-turn orchestrator behavior and MUST remain always-loaded.

- **REQ-RD2-007 (Ubiquitous)**: The `context-window-management.md` rule SHALL remain always-loaded specifically because `session-handoff.md` (KEEP) delegates its numeric `/clear` thresholds to it and carries no inline model-class numbers — scoping it would strip the thresholds from the always-loaded surface.

### B.4 Frontmatter-only change + template parity

- **REQ-RD2-008 (Ubiquitous)**: The scoped-rule change SHALL be frontmatter-only. No SCOPE-rule BODY content SHALL be modified — only the `paths:` field is added. (The additive one-line retrieval summaries mandated by REQ-RD2-004/005 target always-loaded KEEP surfaces, not the scoped rule bodies, and are the only body-touching edits permitted.)

- **REQ-RD2-009 (Where MIRRORED)**: **Where** a scoped rule is MIRRORED (all 6 are present in both trees), the `paths:` field SHALL be applied to BOTH the template SSOT tree and the live deployed tree with identical frontmatter — template-first per CLAUDE.local.md §2, then re-embedded via `make build`, with byte-identical template/live parity verified.

### B.5 Measurable delta + guard non-regression

- **REQ-RD2-010 (State-driven)**: **While** the change is in effect, the always-loaded rule count SHALL drop per-tree as LIVE 11→5 and TEMPLATE 11→5, each measurable by a re-runnable `find`/`grep` command.

- **REQ-RD2-011 (Ubiquitous)**: The always-loaded token-budget guard (`internal/config/token_budget_guard.go` + `_test.go`) SHALL continue to pass after the change (measured always-loaded total ≤ `AlwaysLoadedTokenBudget`), and its enumeration test SHALL pass with the reduced no-`paths:` count — providing mechanical evidence the diet is real and non-regressing. The guard source and constant SHALL NOT be modified by this SPEC.

---

## C. Exclusions

The following are explicitly **out of scope** for SPEC-RULE-DIET-002. Each excluded topic is routed to its correct owner so scope creep does not bleed into a load-bearing surface.

### Out of Scope — KEEP-class rules

- No `paths:` field SHALL be added to any of the 5 KEEP rules. Their always-loaded residency is a correctness requirement, not debt to be reclaimed (see REQ-RD2-006 / REQ-RD2-007).
- No content trimming, splitting, or relocation of the KEEP rules — this SPEC is a scoping diet of reference doctrine, not a rewrite of the load-bearing surface.

### Out of Scope — rule-body relocation to .moai/docs/

- Mechanism (b) — moving a scoped rule's body to `.moai/docs/` and leaving a 1-line pointer — is REJECTED for this SPEC in favor of self-referential `paths:` (mechanism (a)). Rationale: mechanism (a) is frontmatter-only (lowest risk, matches the RULE-SCOPING-001 precedent), keeps the files discoverable as rules, and preserves every existing cross-reference path without rewrites. Relocation would require rewriting all inbound `.claude/rules/moai/...` cross-references.

### Out of Scope — token-budget ratchet

- This SPEC SHALL NOT lower the `AlwaysLoadedTokenBudget` constant to "lock in" the reduction. The guard is a regression tripwire with intentional headroom, not a ratchet; tightening it is a separate concern that would fight normal future rule edits. The guard merely validates the diet is real (REQ-RD2-011).

### Out of Scope — CLAUDE.md / output-style / maintainer-local diet

- Dieting `CLAUDE.md`, the output-style body, the guardrail hooks, and `CLAUDE.local.md` is owned by the sibling Epic Steering-Align SPECs (CLAUDEMD-DIET-001, OUTPUT-STYLE-SLIM-001, GUARDRAIL-HOOK-001, LOCAL-DIET-001), all completed. This SPEC touches only the `.claude/rules/moai/**` rule surface (plus the two additive one-line retrieval summaries REQ-RD2-004/005 mandate).

### Out of Scope — natural file-touch triggers for scoped rules

- Assigning a broader natural-file-touch `paths:` glob (e.g. `dynamic-workflows.md` → `**/.claude/workflows/**`) instead of the self-referential glob is deliberately NOT done. A uniform self-referential policy (REQ-RD2-002) is chosen for consistency and to avoid over-triggering; a future SPEC MAY refine specific globs if a clean trigger proves valuable.

---

## D. Constraints

- **Frontmatter syntax**: comma-separated quoted glob string with the in-repo `**/` precedent prefix, e.g. `paths: "**/runtime-recovery-doctrine.md"`.
- **Template-First (CLAUDE.local.md §2 [HARD])**: edit the template SSOT tree first, then `make build` to re-embed, then verify live-tree byte-identical parity. All 6 scoped rules are MIRRORED, so all 6 follow this path.
- **Language neutrality (CLAUDE.local.md §15)**: frontmatter `paths:` values are language-neutral globs; no language bias is introduced into `internal/template/templates/**`.
- **Guard immutability**: `internal/config/token_budget_guard.go` and its test are NOT edited; they are consumed as the verification oracle.

---

## E. Self-Verification (plan-phase)

- SPEC ID `SPEC-RULE-DIET-002` passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition: SPEC ✓ | RULE ✓ | DIET ✓ | 002 ✓ → PASS).
- 12 canonical frontmatter fields present + `tier: M` + `era: V3R6`.
- Every REQ is testable by a re-runnable command (see acceptance.md).
- Exclusions section carries ≥1 `### Out of Scope — <topic>` H3 sub-heading with `-` bullets.
- KEEP/SCOPE boundary is evidence-based per §A.3; the anti-goal (over-scoping a load-bearing rule) is guarded by REQ-RD2-006/007 and the mitigation requirements REQ-RD2-004/005.

---

## F. Cross-References

- `.moai/specs/SPEC-STEERING-ALIGN-RULE-SCOPING-001/` — predecessor (15→11); MIRRORED vs LIVE-ONLY handling, frontmatter-only precedent, the `glm-web-tooling.md` self-referential-paths precedent.
- `internal/config/token_budget_guard.go` + `token_budget_guard_test.go` — the always-loaded token-budget guard (verification oracle for REQ-RD2-011).
- `.claude/rules/moai/core/agent-common-protocol.md § Recovery-Signal Carve-Out` — the recovery entry-point pointer preserving runtime-recovery-doctrine retrieval (REQ-RD2-004).
- `.claude/rules/moai/workflow/session-handoff.md` Trigger #1 — the SSOT-delegation that makes context-window-management.md KEEP (REQ-RD2-007).
- `.claude/output-styles/moai/moai.md § Banner Localization` — the always-loaded surface preserving the 4-term taxonomy (REQ-RD2-005).
- Anthropic "Steering Claude Code" + "Write an effective CLAUDE.md" — the governing scoping principle (§A.1).
