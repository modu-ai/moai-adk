---
id: SPEC-LSEL-LOCAL-EVOLUTION-001
title: "Local Self-Evolution Loop — close the broken PROPOSE→APPLY seam in user-owned surfaces without touching distributed doctrine"
version: "0.1.0"
status: in-progress
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P1
phase: "v3.x target"
module: lsel
lifecycle: spec-anchored
tags: "lsel, self-evolution, harness-learning, propose-apply-seam, user-owned, frozen-doctrine, bypass-applier"
tier: M
related_specs: [SPEC-HARNESS-EVOLVE-001, SPEC-HARNESS-EVOLVE-002]
---

# SPEC-LSEL-LOCAL-EVOLUTION-001 — Local Self-Evolution Loop (user-owned PROPOSE→APPLY closure)

## HISTORY

- 2026-08-04 — Initial draft. Codifies the design report `.moai/reports/moai-local-self-evolution-design-20260804.html` (v2, ~626 lines, ultracode workflow: 5 parallel research → 1 design → 2 parallel critique → 1 synthesize). The report is the canonical source; this SPEC is its plan-phase artifact translation. Two rounds of adversarial critique on the v1 design surfaced the self-amending-allowlist defect (allowlist inside an evolvable skill file); the synthesized v2 design (this SPEC's basis) relocates the allowlist to a frozen meta file outside the 6 evolvable surfaces and adds execution-meta files to the CSA forced-gate list.

## §A. User Story

**As a** MoAI-ADK maintainer (GOOS, single-developer OSS) whose local dogfood repository accumulates large-scale observation (110,403 usage-log events, 569 lessons-inbox stubs over 25 days, 159 tier-promotion rows, 1583 `@MX` annotation lines) but ZERO mechanical drain or apply,
**I want** a local self-evolution loop that closes the broken PROPOSE→APPLY seam — `Applier.Apply` never ran (`manifest.jsonl` absent, `snapshots/` empty), `CuratorDispatch` has 0 production callers, `drain` is a constitution paragraph with 0 Go code — entirely within user-owned surfaces,
**so that** observation data finally flows into refined `feedback_*.md` lesson topics and approved low-risk edits to the 6 evolvable surfaces, while distributed moai doctrine (`.claude/rules/moai/**`, `CLAUDE.md`, `internal/template/templates/**`, retained agents, `moai-*` skills) stays byte-for-byte untouched.

**Outcome hypotheses:**
- The 569-stub lessons-inbox backlog is drained for the first time (companion-offset marked), producing candidate `feedback_*.md` topics.
- The frozen Go applier stays frozen (`enableTriggerInjectionWrites=false` verified at `internal/harness/applier.go:22`); the seam closes via BYPASS (a parallel user-owned applier), not unfreeze.
- Every successful apply is a `lsel-<proposal-id>`-tagged git commit on a feature branch, rollback is `git revert <tag>` (one-liner), and the distributed template leak is blocked by `split_namespace_test.go` + an extended `internal_content_leak_test.go`.

## §B. Context and Background

### §B.1 The audit's central finding (observation alive, apply dead)

The design report §1 stat grid + §2 audit findings, verified against this repo's ground truth on 2026-08-04, establish that half the declared self-learning loop is dead at the identical seam — the PROPOSE→APPLY boundary:

| Signal (report §1 stat grid) | Live value | Audit finding (report §2) |
|------------------------------|-----------|---------------------------|
| `usage-log.jsonl` events | 110,403 | Live (passive collection) |
| `lessons-inbox.jsonl` stubs / 25 days | 569 | ZERO drained; writer-only non-test Go code is `internal/hook/failure_observer.go:129` |
| `tier-promotions.jsonl` rows | 159 | 72 abandoned drafts route nowhere |
| `Applier.Apply` runs | 0 | `manifest.jsonl` absent; `enableTriggerInjectionWrites=false` at `internal/harness/applier.go:22` (verified) |
| `CuratorDispatch` production callers | 0 | Type exists at `internal/harness/curator_dispatch.go`; grep finds self + test only |
| `drain`-marking Go code | 0 | `drain-marking` literal appears only in `moai-constitution.md:147` (a doctrine paragraph), nowhere in `internal/` |
| Noise share of inbox | ~63% | `tool_failure:Bash:UnknownFailure` (timeout / sandbox noise) dominates |

**Root cause** (report §2): observation-first build; apply closure was explicitly deferred (`applier.go:22` + the `@MX:TODO "Phase 4"` annotation). Real improvement today is **model-mediated** — agents read doctrine during SPEC work and hand-carry `feedback_*.md` / `@MX` edits — not a mechanical feedback signal.

### §B.2 The closing principle (report §4)

The seam is closed by a **parallel user-owned applier** that writes ONLY to the 6 evolvable surfaces, validated against a **frozen allowlist** stored OUTSIDE those surfaces. The frozen Go applier is NEVER unfrozen — unfreezing it would inject `@MX` annotations into distributed agent surfaces (doctrine corruption). The loop's write-permission handcuffs (the allowlist) live in a frozen meta file the loop cannot amend.

### §B.3 Surface classification (report §5)

The 6 evolvable surfaces (where the loop MAY write, subject to the allowlist):

1. `CLAUDE.local.md` (git-tracked local dev guide, NOT distributed)
2. `.claude/settings.local.json` (runtime-managed personal settings, NEVER templated)
3. `memory/` (`feedback_*.md` + `MEMORY.md` auto-memory, `memory/_archive/` cold tier)
4. `.claude/agents/harness/**` + `hns-*` skills (user-owned harness namespace, §24 protected)
5. `.moai/lessons-inbox.jsonl` (Tier-1 raw trace, drain target)
6. `.moai/state/` + new `.moai/state/lsel/` (session/handoff/goal + loop self-state)

Everything else is FROZEN: `.claude/rules/moai/**`, `CLAUDE.md`, `internal/template/templates/**`, retained agents under `.claude/agents/moai/**`, `moai*` / `moai-*` skills, the frozen Go `applier.go` / `curator_dispatch.go`, `[ZONE:Frozen]` markers, AND the new frozen allowlist meta file itself.

### §B.4 v2 corrections (do NOT regress to v1)

- `.moai/config/sections/**` is merge-managed at the framework layer (3-way merge preserves user keys via `backup.go` + `MergeYAML3Way` + `node_merge.go`), NOT a loop-writable surface. The term "evolvable" is reserved exclusively for the six loop-writable surfaces enumerated in §B.3; config sections are framework-managed (updated via Template-First PRs and the `moai update` merge path), and the LSEL loop does not write to them. BUT a NEW section file (e.g. `lsel.yaml`) is wiped + restored from template on `moai update` (`deploy.go:113`) — therefore this SPEC puts loop state under `.moai/state/lsel/` and creates NO new config section file.
- Mechanism skills follow the `hns-*`→`moai-*` prototype→graduate path. M1-M3 artifacts are `hns-lsel-*` (this repo's dogfood, GOOS-owned, NOT distributed). Graduation to `moai-lsel-*` + `internal/template/templates/` (16-language distribution) is a SEPARATE SPEC and explicitly out of scope (§G).

**Glossary note (D4 de-overload):** "evolvable" in this SPEC denotes a surface the LSEL loop MAY write to (the six surfaces in §B.3). It does NOT denote the framework-layer merge-management property of `.moai/config/sections/**`. The two concepts are distinct: framework-managed surfaces are outside the loop's write scope and are updated by humans through the Template-First cycle; loop-writable surfaces are the six enumerated in §B.3 where the LSEL loop may propose edits. REQ-LSEL-001's hard-reject of `.moai/config/sections/**` is therefore consistent with §B.4's framing — config sections are framework-managed, and the loop does not write to them.

## §C. Requirements (GEARS)

### §C.1 Cross-cutting safety invariants (bind EVERY milestone)

#### REQ-LSEL-001 (Ubiquitous, cross-cutting)
The LSEL loop shall confine every write operation to the 6 evolvable surfaces listed in §B.3, validated by a positive-allowlist regex that hard-rejects any other path (including `.claude/rules/moai/**`, `CLAUDE.md`, `internal/template/templates/**`, retained agents, `moai-*` skills, and all of `.moai/config/sections/**`) and records each rejection in `.moai/logs/lsel-reject.log`.

#### REQ-LSEL-002 (Ubiquitous, cross-cutting — the heaviest fix from critique)
The frozen-surface allowlist itself shall live in a frozen meta file OUTSIDE the 6 evolvable surfaces, such that the loop cannot amend its own write-permission handcuffs; any proposal touching the allowlist, the applier/curator skill bodies, the `lsel-apply.sh` hook script, or the `settings.local.json` hook-registration subblock SHALL be a CSA forced-gate target (synchronous `AskUserQuestion`) regardless of proposer confidence. The forced-gate on these four execution-meta categories is mechanically enforced by the applier per REQ-LSEL-005; REQ-LSEL-002 names the doctrine, REQ-LSEL-005 names the mechanism.

#### REQ-LSEL-003 (Ubiquitous, cross-cutting)
The frozen Go applier (`internal/harness/applier.go`) shall remain frozen — `enableTriggerInjectionWrites` stays `false` — and the LSEL loop shall close the PROPOSE→APPLY seam by BYPASS (a parallel user-owned applier writing only to the 6 evolvable surfaces), never by unfreezing the distributed applier.

#### REQ-LSEL-004 (Ubiquitous, cross-cutting)
Every successful LSEL apply shall be a single git commit tagged `lsel-<proposal-id>` on a feature branch (PR-mandatory per repo-local-pr-policy.md; Tier S self-merge allowed), such that rollback is the one-liner `git revert <lsel-tag>` and the apply is traceable in `git log --grep "lsel-"`.

#### REQ-LSEL-005 (Ubiquitous, cross-cutting — CSA escalation, report §7)
The LSEL APPROVE stage shall force a synchronous `AskUserQuestion` — regardless of proposer confidence — for any proposal whose blast radius touches: the INVARANTS kernel at the top of `CLAUDE.local.md`, security/validation exception bands, HIGH-fan-in references, Bash risk paths, `permissions.allow` additions (explicit security-exception band, per-line synchronous approval), OR any execution-meta file named in REQ-LSEL-002; and forced gates SHALL be bother-cost-exempt (the bother-cost gating rule applies only to routine-tier proposals).

**Enforcement clause (mechanical trigger, D3).** The forced-gate above is mechanically enforced, not merely declared in doctrine. The applier (`hns-lsel-applier` skill driving the `lsel-apply.sh` hook script) intercepts every proposal whose `target_surface` OR `blast_radius` matches any of the four execution-meta categories — namely (i) the frozen allowlist meta file at `.claude/lsel/frozen-allowlist.json`, (ii) the body of an applier or curator skill (any file under `hns-lsel-applier/` or `hns-lsel-curator/`), (iii) the apply hook script itself (`lsel-apply.sh` and its wrappers), or (iv) the `settings.local.json` hook-registration subblock — and REFUSES to write unless the proposal's decision record carries an explicit synchronous-approval marker (an `AskUserQuestion`-produced approval artifact recorded in `decision.json`). The interception is a positive-match gate keyed on those four execution-meta surfaces; the applier performs the path/category match before any write primitive is invoked, and a match with no approval marker aborts the apply, appends a rejection row to `.moai/logs/lsel-reject.log` naming the matched category, and writes nothing. Proposals that match none of the four categories proceed through the routine bother-cost path. This mechanical interception is what makes the self-amending-handcuffs defense (option-b of author-flag-1) defensible without resting on the regex paradox alone: the loop cannot write to the files that define its own permission surface, because the applier refuses the write before it happens.

#### REQ-LSEL-006 (Ubiquitous, cross-cutting — namespace separation)
All new LSEL artifacts shall live under `hns-lsel-*` skills, `.claude/agents/harness/`, or `.moai/state/lsel/`, and the repo's `internal/template/split_namespace_test.go` plus an extended `internal_content_leak_test.go` shall assert ZERO leak of `lsel` / `hns-lsel` / SPEC-ID / internal-SHA tokens into `internal/template/templates/`.

#### REQ-LSEL-007 (State-driven — mechanical trigger, report §11 mustFix B#1/B#2)
**While** no mechanical trigger fires, the LSEL loop SHALL NOT be considered closed; the loop shall be mechanically fired by (a) a SessionStart hook that checks `wc -l .moai/lessons-inbox.jsonl` against `drain-offset.json + N` and emits a system-reminder when the backlog exceeds threshold, AND (b) a default-on scheduled `/loop` read-only recipe, such that the loop does not depend on the orchestrator remembering to invoke it (the audit's exact failure mode).

#### REQ-LSEL-008 (State-driven — queue-only apply, report §11 mustFix A#5)
**While** `lsel-apply.sh` is the apply playback path, it SHALL consume ONLY already-approved `decision.json` records and SHALL NOT create new apply operations; the script is a playback device, not a second apply path.

### §C.2 Per-milestone behavioral requirements

#### REQ-LSEL-009 (Event-driven — M1 Drain closure)
**When** the `hns-lsel-curator` skill is invoked (manually via `/moai:harness lsel drain` OR mechanically via the REQ-LSEL-007 SessionStart/`/loop` trigger), the curator shall read `.moai/lessons-inbox.jsonl` starting from the `drain-offset.json` companion offset, cluster stubs by `event_key`, discard single-occurrence noise, apply a Generative-Agents-style 1-10 importance gate (informed by a write-time coarse severity hint so that Bash-timeout / sandbox-violation noise is filtered before clustering), and mark consumed stubs in `drain-offset.json` — all WITHOUT writing to `memory/` (M1 produces candidate topics only; no APPLY yet).

#### REQ-LSEL-010 (Event-driven — M2 PROPOSE shadow + retrieval-before-propose)
**When** the curator extends to shadow proposals at `.moai/state/lsel/proposals/<proposal-id>/`, it SHALL first retrieve relevant `feedback_*.md` files (Reflexion-style retrieval-before-propose), then emit a proposal payload `{proposal_id, target_surface ∈ 6 evolvable, diff.patch, rationale, WHY-not-just-WHAT, prediction (falsifiable expected effect), verify_command, blast_radius, memory_type: semantic|procedural|episodic}` accompanied by a self-critique against frozen doctrine, and the proposal SHALL be blocked until the self-critique is clean; M2 SHALL also VERIFY (via grep / behavioral test) that the existing `moai-harness-learner` Tier-4 AskUserQuestion flow actually fires before depending on it, and SHALL add `CLAUDE.local.md §28` (LSEL operating instructions) plus the INVARANTS kernel block at the top of the file.

#### REQ-LSEL-011 (Event-driven — M3 APPLY closure via bypass, report §11 mustFix A#1)
**When** an APPROVE stage produces an approved `decision.json`, the `hns-lsel-applier` skill (driving `lsel-apply.sh` registered in `settings.local.json`) SHALL validate the target path against the frozen allowlist (REQ-LSEL-001/002), write the diff via Edit/Write/git, emit a `lsel-<proposal-id>`-tagged commit (REQ-LSEL-004) on a feature branch, and append a row to `.moai/state/lsel/apply-ledger.jsonl` (the manifest the frozen applier never produced, finally real in user-owned space); the frozen Go applier remains at `enableTriggerInjectionWrites=false`.

#### REQ-LSEL-012 (Event-detected — M3 rollback rehearsal, report §11 mustFix A#8)
**When** a mixed history of `lsel-*` commits and manual edits to `CLAUDE.local.md` is reverted via `git revert <lsel-tag>`, the revert SHALL land without conflict on at least one representative fixture history (the rollback-rehearsal AC — a precondition for M3 ship).

#### REQ-LSEL-013 (Event-detected — M4 VERIFY independence + flaky policy, report §11 mustFix B#6)
**When** an apply completes, the VERIFY stage SHALL run BOTH the proposal's author-supplied `verify_command` AND a mandatory `/moai gate` superset (lint+format+type+test), because proposer-authored verify alone is circular logic; on a timeout-class failure the VERIFY SHALL retry exactly once, and on a second non-timeout failure the VERIFY SHALL auto-`git revert lsel-<proposal-id>` and mark the proposal's `feedback_*.md` `verified:false` (a Reflexion lesson the proposer retrieves before the next proposal).

#### REQ-LSEL-014 (State-driven — M4 periodic reflection, report §6 stage 7)
**While** reflection-threshold accumulation (NOT a wall clock) reaches the configured ceiling, a periodic REFLECTION pass SHALL synthesize an abstract-principle `feedback_*.md` from several concrete topic files, move the originals to `memory/_archive/` (cold tier, NOT deleted), label each new topic with a CoALA `memory_type` (procedural routes to `hns-*` skills, semantic to `feedback`), and weight retrieval by a decay function so older lessons fall out of the top-5.

#### REQ-LSEL-015 (Capability-gate — M5 personalization is CONDITIONAL)
**Where** the LSEL personalization layer (M5) ships, it SHALL first prove in a NAMED simulation harness that Thompson-sampling posterior auto-tune of evolvable config (effort / model / harness level) reduces override-rate versus the static template default, with a hysteresis margin (anti-thrash) and a bother-cost gate; **Where** that proof is absent, the personalization layer SHALL NOT ship (M5 may never land — report §13).

### §C.3 Honesty / scope-boundary requirements

#### REQ-LSEL-016 (Ubiquitous — scope honesty, report §13 caveat 1)
The LSEL loop shall close ONLY the GOOS-local PROPOSE→APPLY seam; the distributed dead apply buttons (`moai harness apply` / `execute` / `propose` CLI shipped to all 16-language users) SHALL remain out of scope and SHALL be addressed by a SEQUENTIAL follow-up companion SPEC (placeholder ID `SPEC-LSEL-DEAD-VERB-RETIRE-001`, explicitly NOT yet authored, deferred per §G.1).

## §D. Acceptance Criteria Matrix (summary — full Given-When-Then in acceptance.md)

| AC ID | REQ | Severity | Summary | Milestone |
|-------|-----|----------|---------|-----------|
| AC-LSEL-001 | REQ-LSEL-001 | MUST | allowlist hard-rejects FROZEN paths; reject log non-empty on fixture | M3 |
| AC-LSEL-002 | REQ-LSEL-002 | MUST | allowlist stored OUTSIDE 6 evolvable surfaces; proposal-to-amend-allowlist triggers forced gate | M3 |
| AC-LSEL-003 | REQ-LSEL-003 | MUST | `enableTriggerInjectionWrites` stays `false` post-M3; grep confirms no LSEL code mutates it | M3 |
| AC-LSEL-004 | REQ-LSEL-004 | MUST | every apply = one `lsel-*` commit on feature branch; `git log --grep "lsel-"` finds it | M3 |
| AC-LSEL-005 | REQ-LSEL-005 | MUST | CSA forced-gate categories enumerated; bother-cost exemption explicit; applier mechanically refuses execution-meta proposals lacking synchronous-approval marker (D3) | M2 |
| AC-LSEL-006 | REQ-LSEL-006 | MUST | `split_namespace_test.go` + extended `internal_content_leak_test.go` green; zero `lsel`/`hns-lsel`/SPEC-ID leak into template | M2 |
| AC-LSEL-007 | REQ-LSEL-007 | MUST | SessionStart backlog check fires system-reminder on fixture overflow; default `/loop` recipe registered | M2 |
| AC-LSEL-008 | REQ-LSEL-008 | MUST | `lsel-apply.sh` is queue-playback-only (grep: no new-apply creation primitive) | M3 |
| AC-LSEL-009 | REQ-LSEL-009 | MUST | 569-stub backlog drain produces ≥1 candidate topic; `drain-offset.json` advances; zero `memory/` writes | M1 |
| AC-LSEL-010 | REQ-LSEL-009 | MUST | drain severity filter excludes Bash-timeout/sandbox-violation noise BEFORE clustering | M1 |
| AC-LSEL-011 | REQ-LSEL-010 | MUST | shadow proposal payload schema present; self-critique blocks until clean; retrieval-before-propose evidenced | M2 |
| AC-LSEL-012 | REQ-LSEL-010 | MUST | `moai-harness-learner` Tier-4 AskUserQuestion flow verified-firing before wiring (M2 gate) | M2 |
| AC-LSEL-013 | REQ-LSEL-011 | MUST | first `lsel-*` tagged commit lands; apply-ledger row appended; allowlist validates path | M3 |
| AC-LSEL-014 | REQ-LSEL-012 | MUST | rollback-rehearsal: `git revert <lsel-tag>` lands clean on mixed-history fixture | M3 |
| AC-LSEL-015 | REQ-LSEL-013 | MUST | VERIFY runs `/moai gate` superset; timeout retry-once; auto-revert on second fail | M4 |
| AC-LSEL-016 | REQ-LSEL-014 | MUST | reflection pass synthesizes principle file; originals archived (not deleted); decay-weighted retrieval verified | M4 |

(M5 carries no MUST AC — it is conditional per REQ-LSEL-015.)

## §E. Constraints (non-functional)

- **Template-First (CLAUDE.local.md §2):** the LSEL loop writes ONLY to user-owned surfaces; it touches NO file under `internal/template/templates/`. The Template-First cycle (edit template → `make build` → commit) remains the SSOT for distributed doctrine changes, and is the mandatory path for any doctrine-level change the loop's reflection concludes is warranted.
- **PR-mandatory (CLAUDE.local.md §23):** every `lsel-*` commit lands via feature-branch + PR (Tier S self-merge allowed; 0 approvals + green CI). Direct-to-`main` is disabled by `enforce_admins: true`.
- **Namespace separation (CLAUDE.local.md §24):** `hns-*` / `.claude/agents/harness/` / `.moai/state/lsel/` are user-owned; `moai update` MUST NOT delete or modify them.
- **Verification-claim integrity (.claude/rules/moai/core/verification-claim-integrity.md):** every drain / cluster / apply / verify claim MUST be attributable to a run command + observed output (the `.moai/state/verify/<session>/` verbatim-output contract). Summaries are not evidence.
- **No new `.moai/config/sections/` file** (§B.4 v2 correction): `lsel.yaml` or any new section file would be wiped on `moai update` (`deploy.go:113`); loop state lives under `.moai/state/lsel/`.
- **No model parameter updates (report §13):** all self-improvement is prompt-time / in-context retrieved text (ExpeL constraint); no fine-tuning.
- **No subagent `AskUserQuestion` (CLAUDE.md §8):** LSEL mechanism skills and specialists are subagents; on a missing input they return a blocker report, and the orchestrator runs the `AskUserQuestion` round.

## §F. Technical Approach (summary — full in plan.md)

The loop is seven stages — OBSERVE → CLUSTER → PROPOSE → APPROVE → APPLY → REMEMBER → VERIFY — plus a periodic REFLECTION pass. OBSERVE reuses the existing live hooks unchanged (plus a new `signals.jsonl` for intentional hints). CLUSTER/PROPOSE are model-mediated skills over JSONL (no Go code added; the frozen `curator_dispatch.go` stays inert and read-only-referenced). APPLY is the parallel user-owned applier. REMEMBER reuses Claude Code auto-memory. VERIFY reuses `/moai gate` + the verification-batch pattern. The mechanism skills (`hns-lsel-curator`, `hns-lsel-applier`) are `hns-*` prototypes; graduation to `moai-lsel-*` + template distribution is a separate SPEC.

## §G. Out of Scope

### Out of Scope — Distributed dead apply buttons (report §13 caveat 1)

- The `moai harness apply` / `execute` / `propose` CLI shipped to all 16-language template users remains DEAD after this SPEC. Closing the GOOS-local PROPOSE→APPLY seam does NOT close the distributed surface, and this SPEC makes no such claim.
- The distributed dead-verb retirement is deferred to a SEQUENTIAL follow-up companion SPEC, placeholder ID `SPEC-LSEL-DEAD-VERB-RETIRE-001` (explicitly NOT yet authored — a future follow-up). Boundary rationale (per design-report §13 caveat 1): the local bypass has independent value and ships first (it unblocks the 569-stub backlog drain in user-owned surfaces without waiting on the distributed-CLI retirement decision); the distributed dead-verb retirement is a separate concern whose scope (drive the existing opt-in `execute` path vs. explicitly retire the verbs) warrants its own design pass.
- Shipping the local bypass while the distributed 69 proposals stagnate on the CLI surface is a known slow doctrine-erosion hazard this SPEC does NOT solve; the companion retirement SPEC owns that hazard.

### Out of Scope — Graduation to `moai-lsel-*` + 16-language distribution

- M1-M3 produce `hns-lsel-*` mechanism skills (dogfood, GOOS-owned, NOT distributed). Promoting them to `moai-lsel-*` + `internal/template/templates/` (16-language distribution = doctrine) is a separate SPEC gated on M3 verification evidence.

### Out of Scope — Model-performed self-critique as a safety floor

- PROPOSE self-critique is model-performed, NOT a mechanical doctrine checker (report §13 caveat 3). The model can rationalize; the frozen allowlist only stops it from WRITING doctrine, not from MISREADING doctrine. The actual safety weight rests on the mechanical allowlist + `/moai gate` verification floor; self-critique is a light inner door.

### Out of Scope — Fine-tuning / parameter updates

- The loop performs prompt-time / in-context self-improvement only (ExpeL / Voyager constraint). No fine-tuning of the orchestrator model.

### Out of Scope — Local-state durability across machine loss

- `apply-ledger.jsonl` and `.moai/state/lsel/` are gitignored machine-local. The immutable audit ledger IS finally real (in user-owned space), but its durability holds only while local state lives. `moai clean` may erase state; `lsel-*` commits survive on the branch.

### Out of Scope — P5 personalization guarantees

- M5 personalization is CONDITIONAL (REQ-LSEL-015). The named simulation harness is not yet specified in this SPEC (the report under-specifies it — flagged [DECISION] in plan.md §B). M5 may never ship.

## §H. Cross-References

- **Design report (SSOT):** `.moai/reports/moai-local-self-evolution-design-20260804.html` — §1 stat grid, §2 audit, §3 five research lenses, §6 seven-stage loop, §7 hardened guardrails, §8 artifact table, §10 five-stage roadmap (P1-P5 → M1-M5), §11 two-lens critique + mustFix, §12 risks, §13 honest caveats.
- **Frozen applier (reference only):** `internal/harness/applier.go` (`enableTriggerInjectionWrites=false` at line 22), `internal/harness/curator_dispatch.go`.
- **3-way merge (why `.moai/config/sections/**` is evolvable, why a new section file is NOT the loop-state location):** `internal/config/backup.go`, `internal/config/node_merge.go`, `internal/config/deploy.go:113`.
- **Template-First / namespace / PR doctrine:** `CLAUDE.local.md` §2 / §23 / §24 / §25.
- **Namespace guard:** `internal/template/split_namespace_test.go`, `internal/template/internal_content_leak_test.go` (to be extended).
- **Sibling SPECs:** `SPEC-HARNESS-EVOLVE-001`, `SPEC-HARNESS-EVOLVE-002` (write-layer-only Curator Editable Surfaces — LSEL is the user-owned BYPASS parallel to their frozen Go write layer). (The formerly referenced `SPEC-HARNESS-EVOLUTION-EPIC` does not exist on disk and has been removed from this list; no live umbrella SPEC currently groups the EVOLVE family.)
- **Verification-claim integrity:** `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (defect/budget claims require mechanical tool confirmation — the audit's own grep is the evidence baseline).
- **Constitution drain paragraph (the "0 Go code" doctrine stub):** `.claude/rules/moai/core/moai-constitution.md:147`.
