---
id: SPEC-SEC-DEEPSCAN-001
title: "On-demand multi-agent deep vulnerability scan (/moai review --deep) — Epic SECURITY-ABSORB SPEC-1"
version: "0.1.0"
status: in-progress
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/template/templates/.claude (skills/moai/workflows/review.md, commands/moai/review.md.tmpl) + local .claude siblings + .moai/config/sections/delegation.yaml"
lifecycle: spec-anchored
tags: "security, deepscan, security-absorb, review, adversarial-verification, owasp, dynamic-workflow, absorb, template"
tier: L
era: V3R6
related_specs: [SPEC-SUBCOMMAND-RETIRE-001, SPEC-V3R6-SEC-SKILL-INTEGRATION-001]
---

# SPEC-SEC-DEEPSCAN-001 — On-demand multi-agent deep vulnerability scan

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Plan-phase artifact set authored (Tier L, 5 artifacts: spec / plan / acceptance / design / research + progress skeleton). Epic SECURITY-ABSORB, SPEC-1 of the cohort. **Absorption source**: Anthropic's official `claude-security` Claude Code plugin architecture — reimplemented NATIVELY into MoAI's own harness (NOT installed as a third-party plugin). **Candidate A** = the on-demand DEEP multi-agent vulnerability scan. **Surface decision (grounded)**: extends the existing `/moai review` with a new `--deep` mode rather than a new `/moai security` subcommand — `/moai security` was RETIRED by SPEC-SUBCOMMAND-RETIRE-001, and reviving it would contradict that completed decision. **Layering**: this is the ON-DEMAND heavy scan; the lighter always-on in-session guardian is deferred to Epic SECURITY-ABSORB SPEC-2 (see §C Out of Scope). |

---

## §A Context & Problem

### A.1 Motivation

Anthropic ships an official `claude-security` Claude Code plugin: a single-entry command that opens a menu of security-scan jobs and drives a six-phase multi-agent vulnerability scan with independent adversarial verification and reviewer-vouched patch drafting. MoAI already carries most of the raw capability — a `Workflow()` dynamic-workflow primitive with a built-in adversarial-verify pattern, four production-grade security reference skills, worktree-isolated agents, and a read-only `/moai review` lens with a Security perspective — but it does NOT compose them into a rigorous, adversarially-verified, patch-drafting deep scan.

The gap this SPEC closes: today's `/moai review --security` is a single-pass, report-only lens (OWASP checklist + dependency scan + incremental secrets scan). It surfaces candidate findings but performs NO independent adversarial verification (no reachability / impact / defenses cross-examination), produces NO machine-readable finding artifact, and drafts NO reviewer-vouched patches. A security-conscious user who wants the rigor of the absorbed plugin has no MoAI-native path to it.

### A.2 What is absorbed (source architecture, restated in MoAI terms)

The absorbed source architecture, reproduced natively:

1. **Single entry, menu of jobs**: scan the whole repo / scan a diff (branch diff, PR diff, single commit) / suggest patches.
2. **Six-phase multi-agent scan**: (1) architecture map → (2) threat model → (3) vulnerability hunt → (4) independent adversarial verification → (5) report → (6) patch.
3. **Adversarial verification**: every candidate finding passes a 3-voter panel — REACHABILITY, IMPACT, DEFENSES — with a 2-of-3 quorum; a non-unanimous panel caps the finding's stated confidence at "medium". Findings enter the report ONLY after passing.
4. **Reviewer-vouched patches**: drafted in an isolated scratch copy, then reviewed by an agent independent of the drafter, which vouches for THREE claims — (a) addresses only the one finding, (b) introduces no new vulnerability, (c) leaves behavior otherwise unchanged. When it cannot vouch for all three, it emits a short note instead of a patch. Patches are NEVER auto-applied — one finding = one patch = one PR, applied by the user via `git apply`.
5. **Timestamped results directory**: a human report (`.md`, finding IDs `F1..`, impact, exploit scenario, severity, confidence, recommendation), a machine-readable `.jsonl` (one finding per line), and a revision stamp (which commit was scanned + effort tier + whether the working tree was included). The directory ships its own `.gitignore` so a stray `git add` never sweeps it into a commit.

### A.3 Prerequisite & layering context

- **Dynamic Workflows prerequisite**: the deep scan's primary execution path uses the `Workflow()` dynamic-workflow primitive (Claude Code v2.1.154+). When Workflows are unavailable, the scan degrades gracefully (see §B REQ-SDS-050..052).
- **Layering (composes with, does not replace)**: this is an ON-DEMAND deep-scan layer, invoked explicitly by a user who wants rigor. It is complementary to (a) the existing single-pass `/moai review --security` lens (which remains unchanged), and (b) the lighter always-on in-session guardian that Epic SECURITY-ABSORB SPEC-2 will cover. The three layers compose without overlap (§C Out of Scope enumerates the SPEC-2 boundary).

### A.4 MoAI assets reused (NOT reinvented)

| Absorbed concept | MoAI asset reused |
|------------------|-------------------|
| 6-phase orchestration | `Workflow()` dynamic-workflow primitive — `pipeline()`/`parallel()` stages (`.claude/rules/moai/workflow/dynamic-workflows.md`) |
| 3-voter adversarial panel (2-of-3 quorum) | The dynamic-workflow built-in adversarial-verify pattern (independent REFUTE-skewed voters, perspective-diverse verify) — maps 1:1 |
| Vulnerability-hunt domain knowledge | Existing security reference skills: `moai-ref-owasp-checklist`, `moai-ref-llm-security`, `moai-ref-secops`, `moai-ref-supply-chain` |
| Scratch-clone patch generation | `Agent(isolation: "worktree")` (isolated write copy) |
| Read-only batched verification + evidence contract | `agent-common-protocol.md` § Parallel Execution + `verification-batch-pattern.md` |
| Entry surface + skeptical synthesis | The existing `/moai review` subcommand + sync-auditor's skeptical stance |

### A.5 Non-goals (see §C for the formal exclusions)

This SPEC does NOT add a new agent, does NOT add new Go runtime logic, does NOT ship a static `.js` workflow script into the template tree, and does NOT revive the retired `/moai security` subcommand. The shipped deliverable is a template-distributable, 16-language-neutral orchestrator playbook (the `--deep` mode of the `review` workflow skill) plus the command argument-hint surface.

---

## §B Requirements (GEARS)

### Surface & entry

**REQ-SDS-001 — Deep-scan entry via `/moai review --deep` (Ubiquitous)**
The `review` workflow skill **shall** expose an on-demand deep vulnerability scan through a new `--deep` flag on `/moai review`, composing with the existing `--security` focus (a bare `--deep` is treated as security-focused).

**REQ-SDS-002 — No revival of the retired `/moai security` subcommand (Unwanted behavior)**
The deliverables **shall not** introduce, register, or reference a `/moai security` subcommand, a `security` intent-router row, or any `security`/`audit`/`sec` alias that resurrects the subcommand retired by SPEC-SUBCOMMAND-RETIRE-001; the deep scan **shall** live entirely under the retained `/moai review` surface.

**REQ-SDS-003 — Job menu maps onto review scope flags (Where — capability gate)**
**Where** a user invokes `/moai review --deep`, the workflow **shall** offer the absorbed job menu mapped onto scope selection: whole-repo scan (`--repo`), diff scan (`--staged` / `--branch <B>` / a new `--commit <SHA>` single-commit scope), and patch drafting (a distinct opt-in `--patch` flag), reusing the existing Phase 1 scope-selection mechanism.

**REQ-SDS-004 — Patch drafting is opt-in and off by default (State-driven)**
**While** the `--patch` flag is absent, the deep scan **shall** stop after the report phase and **shall not** draft, write, or apply any patch; patch drafting occurs only when `--patch` is explicitly present.

### Six-phase pipeline

**REQ-SDS-010 — Six-phase pipeline (Ubiquitous)**
The deep scan **shall** execute the six absorbed phases in order: (1) architecture map, (2) threat model, (3) vulnerability hunt, (4) independent adversarial verification, (5) report, (6) patch (patch phase gated by REQ-SDS-004).

**REQ-SDS-011 — Hunt-phase domain-knowledge loading (Where — capability gate)**
**Where** the vulnerability-hunt phase (3) runs, each hunt agent **shall** load the relevant security reference skill(s) — `moai-ref-owasp-checklist`, `moai-ref-llm-security`, `moai-ref-secops`, `moai-ref-supply-chain` — via on-demand `Skill()` injection (per `skill-routing.md` §1), NOT via static frontmatter preload.

**REQ-SDS-012 — Read-only hunt/verify phases (Unwanted behavior)**
The architecture-map, threat-model, vulnerability-hunt, and adversarial-verification phases **shall not** modify any file in the working tree; these phases are read-only reconnaissance and analysis.

### Adversarial verification panel

**REQ-SDS-020 — 3-voter adversarial panel with 2-of-3 quorum (Event-driven)**
**When** a candidate finding is produced by the hunt phase, the adversarial-verification phase **shall** subject it to a 3-voter panel — REACHABILITY, IMPACT, DEFENSES — and **shall** admit the finding to the report ONLY when at least 2 of the 3 voters affirm it (2-of-3 quorum).

**REQ-SDS-021 — Non-unanimous confidence cap (State-driven)**
**While** a finding's panel verdict is non-unanimous (2-of-3, not 3-of-3), the report **shall** cap that finding's stated confidence at "medium" and **shall not** state "high" confidence for it.

**REQ-SDS-022 — Voter independence (Ubiquitous)**
The three panel voters **shall** be independent, perspective-diverse (REFUTE-skewed) evaluators — no voter reuses the hunt agent's own reasoning as its sole basis — reusing the dynamic-workflow built-in adversarial-verify pattern.

**REQ-SDS-023 — Rejected candidates excluded from the report body (Unwanted behavior)**
A candidate finding that fails the 2-of-3 quorum **shall not** appear as a confirmed finding in the human report body; failed candidates MAY be recorded only in a clearly-separated "unconfirmed candidates" appendix (never mixed with confirmed findings).

### Patch drafting & reviewer vouch

**REQ-SDS-030 — Scratch-clone patch drafting (Where — capability gate)**
**Where** `--patch` is present and at least one confirmed finding exists, each patch **shall** be drafted in an isolated scratch copy of the repository via `Agent(isolation: "worktree")` — never in the user's live working tree.

**REQ-SDS-031 — Independent reviewer 3-claim vouch (Event-driven)**
**When** a patch is drafted, an agent independent of the one that authored the patch **shall** review it and vouch for THREE claims: (a) it addresses only the one finding, (b) it introduces no new vulnerability, (c) it leaves behavior otherwise unchanged.

**REQ-SDS-032 — Vouch-failure emits a note, not a patch (State-driven)**
**While** the independent reviewer cannot vouch for all three claims of REQ-SDS-031, the pipeline **shall** emit a short explanatory note for that finding instead of a patch.

**REQ-SDS-033 — Patches are never auto-applied (Unwanted behavior)**
The pipeline **shall not** apply, stage, commit, or push any drafted patch to the user's repository; each confirmed-and-vouched finding produces exactly one patch artifact (one finding = one patch = one PR) that the user applies manually via `git apply`.

### Results artifacts

**REQ-SDS-040 — Timestamped results directory as a report (Ubiquitous)**
The deep scan **shall** write its outputs to a timestamped results directory classified as a REPORT (analysis of existing code) under `.moai/reports/` — NOT under `.moai/specs/` — per the SPEC-vs-report classification rules.

**REQ-SDS-041 — Human report with finding schema (Ubiquitous)**
The results directory **shall** contain a human-readable Markdown report in which each confirmed finding carries a stable finding ID (`F1`, `F2`, …), impact, exploit scenario, severity, confidence, and recommendation.

**REQ-SDS-042 — Machine-readable JSONL (Ubiquitous)**
The results directory **shall** contain a machine-readable findings file with exactly one finding per line (JSONL), each line carrying the finding ID plus its structured fields.

**REQ-SDS-043 — Revision stamp (Ubiquitous)**
The results directory **shall** contain a revision stamp recording which commit was scanned, the effort tier used, and whether the working tree was included in the scan.

**REQ-SDS-044 — Self-contained `.gitignore` (Where — capability gate)**
**Where** the results directory is created, it **shall** contain its own `.gitignore` (ignoring its entire contents) so a stray `git add` can never sweep the results into a commit.

### Cross-cutting constraints

**REQ-SDS-050 — Dynamic Workflows prerequisite documented (Ubiquitous)**
The playbook **shall** document that the deep scan's primary path requires Dynamic Workflows (Claude Code v2.1.154+) and **shall** name the observable signal(s) that determine availability.

**REQ-SDS-051 — Graceful degradation ladder (State-driven)**
**While** Dynamic Workflows are unavailable (disabled, or runtime version below v2.1.154), the deep scan **shall** degrade gracefully to a bounded parallel (Mode-4, 3-5 concurrent) fan-out of the same phases, and — when even that is not viable — to a single-pass `/moai review --security` plus the native `/security-review` fallback, rather than failing.

**REQ-SDS-052 — Degradation preserves the adversarial contract at the Mode-4 rung; labels the single-pass rung as rigor-reduced (Ubiquitous)**
The **Mode-4 fallback rung** (the middle rung of REQ-SDS-051) **shall** preserve the 2-of-3 adversarial quorum and the non-unanimous confidence cap — at that rung, degradation reduces concurrency/scale, NOT verification rigor. The **single-pass last-resort rung** (the final rung of REQ-SDS-051: `/moai review --security` + native `/security-review`) is explicitly a **rigor-reduced** fallback (no per-finding 3-voter panel) and **shall** be labeled as such in the report, so a reduced-rigor result is never mistaken for a full adversarially-verified scan (aligns with design.md §F, which already labels the rungs).

**REQ-SDS-060 — Template-First distribution (Ubiquitous)**
Every shipped artifact **shall** be authored in `internal/template/templates/` first and synced to the local `.claude/` copy, with template and local copies kept lockstep (byte-identical for the workflow skill; rendered for the command).

**REQ-SDS-061 — 16-language neutrality of shipped content (Unwanted behavior)**
The shipped playbook and any shipped pattern/prompt content **shall not** hardcode a single-language toolchain, standard-library name, or platform feature, and **shall not** embed internal SPEC IDs, internal dates, or commit SHAs; the scan's reasoning **shall** apply equally across all 16 supported languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift).

**REQ-SDS-062 — No new agent, no Go runtime, no shipped `.js` (Unwanted behavior)**
The deliverables **shall not** add a new agent to the catalog, **shall not** add new Go runtime logic, and **shall not** ship a static `.js` workflow script into the template tree; the deep scan is expressed as an orchestrator playbook that constructs the Workflow at runtime and reuses existing agents.

**REQ-SDS-063 — AskUserQuestion boundary preserved (Unwanted behavior)**
The playbook **shall not** direct any subagent or workflow agent to prompt the user; all user decisions (scope selection, patch opt-in, degradation confirmation) **shall** be collected by the orchestrator via `AskUserQuestion` before the pipeline launches, and agents that lack input **shall** return blocker reports.

---

## §C Out of Scope

This section prevents scope creep and records the layering boundaries. Items below are explicitly NOT part of SPEC-SEC-DEEPSCAN-001.

### Out of Scope — Always-on in-session guardian (Epic SECURITY-ABSORB SPEC-2)
- The lighter, always-on, in-session security guardian (a per-turn / PostToolUse-style lightweight lens that flags obvious issues as the user codes) is deferred to Epic SECURITY-ABSORB SPEC-2. This SPEC-1 covers ONLY the heavy, explicitly-invoked, on-demand deep scan. The two layers compose without overlap: SPEC-1 = deep + on-demand + report/patch; SPEC-2 = light + always-on + inline.

### Out of Scope — New `/moai security` subcommand
- Reviving the `/moai security` subcommand (retired by SPEC-SUBCOMMAND-RETIRE-001) is explicitly out of scope. The deep scan lives under `/moai review --deep`. No `security`/`audit`/`sec` alias is added.

### Out of Scope — New Go runtime / new agent / shipped workflow script
- No new agent is added to the catalog (`catalog.yaml` agent counts stay unchanged). No new Go package or CLI subcommand is added. No static `.js` dynamic-workflow script is shipped into `internal/template/templates/` (the `.claude/workflows/` directory is user-owned and not template-managed); the Workflow is constructed at runtime from the playbook.

### Out of Scope — Automatic patch application & PR creation
- The deep scan never applies patches, never commits, never opens PRs. It emits patch artifacts the user applies manually via `git apply`. PR creation for an applied patch remains the user's own `/moai sync` / `manager-git` flow, out of this SPEC.

### Out of Scope — Changes to the existing single-pass `/moai review --security` lens
- The behavior of the existing `--security` single-pass lens (OWASP checklist, dependency scan, incremental secrets scan) is unchanged. `--deep` is additive; it does not modify or remove the single-pass path.

### Out of Scope — Results-directory retention / auto-prune
- Retention/auto-prune of old `.moai/reports/security-deepscan-*/` directories is deferred to SPEC-2; the first cut leaves retention to the user (no auto-prune).

### Out of Scope — Offensive security / exploit execution
- The scan is defensive-only: it identifies and explains vulnerabilities and drafts fixes. Authoring working exploits, running attack payloads, or any offensive tooling is out of scope (consistent with the defensive posture of the reused reference skills).

---

## §D Acceptance Criteria

Full Given-When-Then scenarios, the AC-to-REQ matrix, edge cases, and the Definition of Done live in `acceptance.md`. Summary of the must-pass gates:

- Deep scan reachable via `/moai review --deep` (argument-hint + playbook section present in BOTH trees, byte-identical workflow skill).
- Six phases documented in order; adversarial 2-of-3 quorum + non-unanimous "medium" cap present.
- Patch gate: scratch-clone drafting + independent 3-claim vouch + never-auto-apply, all documented; `--patch` opt-in default-off.
- Results-dir schema: human `.md` (F-IDs), machine `.jsonl`, revision stamp, self-`.gitignore` — all specified.
- Graceful degradation ladder documented with availability signal; degraded path preserves the adversarial quorum.
- Template-First: no template/local drift; 16-language neutrality grep clean; no internal SPEC-ID/date/SHA leak in template content.
- No `/moai security` revival; no new agent; no Go runtime; no shipped `.js`.
