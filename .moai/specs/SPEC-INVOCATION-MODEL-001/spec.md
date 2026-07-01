---
id: SPEC-INVOCATION-MODEL-001
title: "Native Invocation-Model Alignment doctrine + /moai feedback enhancement"
version: "0.1.0"
status: completed
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/config, .claude/rules/moai/workflow, .claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "invocation-model, doctrine, feedback, gh-cli, config, template-first"
tier: M
---

# SPEC-INVOCATION-MODEL-001 — Native Invocation-Model Alignment doctrine + /moai feedback enhancement

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-01 | 0.1.0 | manager-spec | Initial draft — two coupled deliverables: (1) native invocation-model alignment doctrine rule; (2) /moai feedback workflow enhancement as the doctrine's first concrete application. |

---

## §A. Context and Motivation (WHY)

A subcommand audit reframed the axis on which a MoAI `/moai` subcommand's existence is justified. The naive axis — "does a MoAI subcommand duplicate a Claude Code native command?" — is wrong. The correct axis, confirmed by official Claude Code documentation, is the **invocation model**:

- **Bundled skills** (`/code-review`, `/simplify`, `/loop`, `/deep-research`) carry a `[Skill]` / `[Workflow]` marker in the official `commands.md` and are **prompt-based** (per `skills.md`): an orchestrator/agent CAN auto-invoke them programmatically via the `Skill()` / `Workflow()` tool inside an automated pipeline (unless `disable-model-invocation: true`). This category is **PROGRAMMATIC**.
- **Built-in commands** (`/security-review`, `/goal`, `/review`, `/clear`, `/compact`) carry NO `[Skill]`/`[Workflow]` marker and **execute fixed CLI logic directly**: only a human typing the slash command triggers them. An orchestrator CANNOT trigger them via any tool call. This category is **HUMAN-ONLY**.

> The per-command classification above is not asserted from memory: it is verified against the official `commands.md` `[Skill]`/`[Workflow]` markers + `skills.md` prose (recorded per-command in the doctrine rule's citation matrix, REQ-IM-002) and re-confirmed at run-phase by a WebFetch of the official docs (AC-IM-022).

Official source (to be cited verbatim in the doctrine rule): Claude Code `skills.md` — "Unlike most built-in commands, which execute fixed logic directly, bundled skills are prompt-based: they give Claude detailed instructions and let it orchestrate the work using its tools." The commands reference (`commands.md`) uses a `[Skill]` / `[Workflow]` marker to distinguish skills/workflows from built-in commands.

**Core thesis.** A MoAI `/moai` subcommand's existence is justified by AUTOMATION — integrating a capability into the orchestrator's plan→run→sync pipeline WITHOUT human typing. Therefore:

- When the native equivalent is a **built-in command (HUMAN-ONLY)**, MoAI reimplementing it for automation is NOT redundant reinvention — it is the ONLY pipeline path (e.g. `/moai review --security` exists because `/security-review` is human-only and cannot run inside an automated workflow).
- When the native equivalent is a **bundled skill (PROGRAMMATIC)**, MoAI SHOULD prefer invoking the native skill via `Skill()` over reimplementing it (reuse over reinvention).

This SPEC (1) codifies this insight as a policy-layer doctrine rule and (2) applies it to `/moai feedback` as the first concrete case: `/moai feedback` has NO native equivalent — it targets the remote `modu-ai/moai-adk` tool repository (bug reports about the MoAI-ADK tool itself), NOT the user's own repository. It is therefore not replaceable by `gh issue create` (which targets the user's own repo) and warrants a first-class MoAI workflow. This SPEC hardens that workflow.

---

## §B. Scope (WHAT)

Two tightly-coupled deliverables.

### Deliverable 1 — Native Invocation-Model Alignment doctrine (new policy-layer rule)

A new rule file at the fixed path `.claude/rules/moai/workflow/native-invocation-model.md` (template mirror: `internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md`) that codifies the invocation-model taxonomy, a classification matrix for the overlapping native commands, the automation-justification thesis, per-subcommand justification notes, and the conditional-PROGRAMMATIC caveat. This is DOCTRINE (policy-layer), NOT a runtime mechanism.

### Deliverable 2 — /moai feedback enhancement

Four hardening items on the existing `.claude/skills/moai/workflows/feedback.md` workflow: (1) automatic tool-diagnostic attachment, (2) duplicate-issue detection, (3) `gh`-failure graceful fallback, (4) target-repo config-ization (default preserved as `modu-ai/moai-adk`).

> Execution-model note: `/moai feedback` runs **orchestrator-direct** — its frontmatter is `user-invocable: false`, its body states "the orchestrator executes directly" with "no subagent spawn", and it ALREADY legitimately uses orchestrator-level `AskUserQuestion` for feedback collection. There is NO subagent to constrain; REQ-IM-014 targets the SHAPE of the dedupe interaction (structured candidate-report, not a blind inline prompt), not a nonexistent subagent boundary.

---

## §C. GEARS Requirements

### Deliverable 1 — Doctrine rule

**REQ-IM-001** (Ubiquitous) — The invocation-model doctrine rule **shall** define the two-category taxonomy: PROGRAMMATIC (bundled skill / workflow, orchestrator-auto-invocable via `Skill()` / `Workflow()`) versus HUMAN-ONLY (built-in command, human-typed only), including the official `skills.md` / `commands.md` citation and the `[Skill]` / `[Workflow]`-marker heuristic for classification.

**REQ-IM-002** (Ubiquitous) — The doctrine rule **shall** include a classification matrix tagging each of the nine overlapping native commands — `/code-review`, `/simplify`, `/loop`, `/deep-research`, `/security-review`, `/goal`, `/review`, `/clear`, `/compact` — as PROGRAMMATIC or HUMAN-ONLY, and **shall** carry an inline citation anchor for EACH of the nine classifications — an official `skills.md`/`commands.md` quote OR the observed `[Skill]`/`[Workflow]` marker (present for bundled skills/workflows, absent for built-in commands). The matrix **shall** additionally record the `/loop` framing correction: the sibling rule `goal-directive.md` describes native `/loop` as "a fixed time interval scheduler", which is INCOMPLETE — the official `commands.md` shows native `/loop` is a bundled `[Skill]` that self-paces when the interval is omitted. The doctrine **shall** carry a cross-reference acknowledging + correcting the `goal-directive.md` framing (the doctrine does NOT edit `goal-directive.md` — that is out of scope).

**REQ-IM-003** (Ubiquitous) — The doctrine rule **shall** state the automation-justification thesis with two axes: Axis A (reuse native bundled skills via `Skill()` where the native equivalent is PROGRAMMATIC) and Axis B (reimplement native built-in commands where the native equivalent is HUMAN-ONLY).

**REQ-IM-004** (Ubiquitous) — The doctrine rule **shall** provide a per-subcommand justification note for each current MoAI subcommand that has a native counterpart, stating why that subcommand exists on the invocation-model axis (e.g. "feedback: no native equivalent — targets the remote `modu-ai/moai-adk` tool repo, not the user's own repo"; "review --security: `/security-review` is HUMAN-ONLY, MoAI automates it").

**REQ-IM-005** (Event-driven) — **When** a bundled skill carries `disable-model-invocation: true`, OR the settings-level `disableBundledSkills` is set, the doctrine rule **shall** document that the skill loses auto-invocability, so Axis-A reuse requires runtime verification of PROGRAMMATIC status before it is relied upon.

### Deliverable 2 — /moai feedback enhancement

**REQ-IM-006** (Event-driven) — **When** a user invokes `/moai feedback`, the feedback workflow **shall** auto-collect and append the two always-available tool-diagnostic items to the issue body — MoAI version (`moai version`) and operating system (`uname`) — the guaranteed diagnostic set and the testable contract of this REQ. The workflow **shall** additionally attempt Go toolchain version (`go version`) on a best-effort basis (its absence is NOT a failure — see the neutrality note below). Additionally, because `/moai feedback` runs orchestrator-direct, **where** the orchestrator has last-failed-command / error context in its own session it **may** pass that context into the feedback invocation and the workflow **shall** append it additively (best-effort — absence is NOT a failure).

> Neutrality note (CLAUDE.local.md §15): `go version` is a MoAI-ADK tool build-provenance diagnostic (the tool binary is implemented in Go) — it is NOT the user's project language, and most users run a prebuilt `moai` binary with no Go toolchain installed. It is therefore collected best-effort; the 16-language template neutrality is unaffected because the item describes the tool binary's provenance, not the user's project.

**REQ-IM-007** (Unwanted behavior) — The feedback workflow **shall not** attach arbitrary user file contents to the issue body; diagnostic attachment is restricted to tool-diagnostic information only.

**REQ-IM-008** (Event-driven) — **When** a user submits a drafted feedback title, the feedback workflow **shall** run a duplicate-issue search (`gh issue list --repo <target> --search "<title keywords>"`) and surface likely-duplicate candidates as a structured report for the orchestrator to present.

**REQ-IM-009** (Event-driven) — **When** `gh` is unauthenticated or rate-limited (detected via `gh auth status` failure or a rate-limit signal), the feedback workflow **shall** follow a graceful fallback path: guide the user to resolve the auth/rate-limit condition and offer to save the drafted issue body locally so no drafted content is lost.

**REQ-IM-010** (Ubiquitous) — The feedback target repository **shall** be a configuration value with a default of `modu-ai/moai-adk` and a maintainer/fork override.

**REQ-IM-011** (Capability gate) — **Where** a feedback target-repository override is configured, the feedback workflow **shall** target the override repository instead of the default `modu-ai/moai-adk`.

### Cross-cutting constraints

**REQ-IM-012** (Ubiquitous) — Every edit to the local `.claude/skills/moai/workflows/feedback.md` and to the new doctrine rule under `.claude/rules/moai/`, and every new `.moai/config/sections/` file, **shall** be mirrored to the corresponding `internal/template/templates/` path, and `make build` **shall** regenerate the embedded templates.

**REQ-IM-013** (Ubiquitous) — The new doctrine rule and the feedback workflow edits, as template-distributed assets, **shall** remain language-neutral across the 16 supported languages; the `modu-ai/moai-adk` default repository is a MoAI-ADK system identifier (allowed in the template), not internal-development leakage.

**REQ-IM-014** (Unwanted behavior) — The duplicate-detection step **shall not** hard-code an inline blind `AskUserQuestion` prompt inside the dedupe logic; instead it **shall** emit a structured candidate-report (the list of likely-duplicate issues) that the orchestrator surfaces. (`/moai feedback` runs orchestrator-direct and already uses orchestrator-level `AskUserQuestion` legitimately for feedback collection — this constraint is about not embedding a blind dedupe prompt inside the dedupe step, NOT a nonexistent subagent boundary.)

---

## §D. Acceptance Criteria Summary

Full testable AC enumeration lives in `acceptance.md`. Traceability: REQ-IM-001..005 → Deliverable 1 doctrine-presence checks; REQ-IM-006..011 → feedback workflow content + config-resolution unit test; REQ-IM-012..014 → local↔template parity + neutrality + boundary grep.

---

## §E. Exclusions (What NOT to Build)

This section is load-bearing: it defines what is explicitly out of scope so run-phase does not drift.

### Out of Scope — Axis-A bundled-skill reuse refactoring

- Refactoring `/moai clean` to invoke the native `/simplify` skill, or `/moai review` to invoke the native `/code-review` skill, is deferred to a follow-up SPEC. This SPEC establishes the doctrine and applies it ONLY to `/moai feedback`.
- No existing MoAI subcommand is retired or re-pointed at a native skill in this SPEC.

### Out of Scope — Legacy subcommand retirement

- Retirement of the legacy `design` / `brain` / `e2e` / `coverage` / `security` subcommands is owned by `SPEC-SUBCOMMAND-RETIRE-001` (a DIFFERENT set of subcommands from the 9 native commands classified here). There is no overlap; this SPEC does not touch that retirement work.

### Out of Scope — Runtime enforcement of the doctrine

- The doctrine rule is policy-layer (codification) ONLY. No hook, lint rule, or runtime mechanism is added to mechanically enforce the PROGRAMMATIC/HUMAN-ONLY classification. Runtime verification of a bundled skill's PROGRAMMATIC status (REQ-IM-005) is an agent obligation, not a mechanically-enforced gate.

### Out of Scope — Config schema beyond the feedback target repo

- The config-ization (REQ-IM-010/011) adds ONLY the feedback target-repository key. It does not introduce any other feedback configuration surface (issue templates, label overrides, auth config, etc.).

### Out of Scope — Feedback UX beyond the four hardening items

- No new feedback types, no priority-model change, no browser-automation for issue viewing, and no change to the existing issue-language policy are introduced.

---

## §F. Cross-References

- `.claude/rules/moai/workflow/archived-agent-rejection.md` — the migration table establishing that issue creation has no retained-agent owner (feedback runs orchestrator-direct).
- `.claude/rules/moai/development/coding-standards.md` § Footer Convention — rule-file footer policy for the new doctrine rule.
- `CLAUDE.local.md` §2 (Template-First Rule), §15 (16-language neutrality), §25 (Template Internal-Content Isolation) — the mirror + neutrality constraints.
- `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator–Subagent Boundary — the orchestrator owns the AskUserQuestion round that surfaces the REQ-IM-014 dedupe candidate-report (feedback runs orchestrator-direct; the constraint is about not embedding a blind dedupe prompt, not a subagent boundary).
- `SPEC-SUBCOMMAND-RETIRE-001` — sibling SPEC (legacy subcommand retirement); disjoint scope, noted for awareness only.
