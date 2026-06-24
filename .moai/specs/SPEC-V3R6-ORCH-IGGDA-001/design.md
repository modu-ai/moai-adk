# Design — SPEC-V3R6-ORCH-IGGDA-001

> **Tier L design artifact**. This document carries the rigorous justification for amending a FROZEN invariant (Implementation Kickoff Approval). §F is the load-bearing section.

---

## §A — The IGGDA 4-phase architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 0 — Intent (human-in-the-loop, Socratic)                     │
│  - Socratic interview to 100% clarity (CLAUDE.md §7 Rule 5)        │
│  - Drain ALL preferences (Tier, mode, PR strategy, domain)         │
│  - ABSORBS the gating weight of Implementation Kickoff Approval   │
│    under Path B (the human checkpoint moves HERE, except in        │
│    dangerous domains where it remains at the plan→run boundary)    │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ auto-advance (D4 Stop hook driver)
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 1 — Plan (autonomous)                                        │
│  - manager-spec: intent → SPEC artifacts                            │
│  - plan-auditor: INDEPENDENT audit (fresh context, bias prevention)│
│  - PASS → auto-advance to Phase 2                                   │
│  - FAIL/INCONCLUSIVE → halt, surface to user (REQ-IGGDA-023)       │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ auto-advance + Implementation Kickoff Approval
                           │   (safe-condition predicate — §F)
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 2 — Run (autonomous + recursive self-diagnosis)              │
│  - manager-develop: implementation (DDD/TDD/autofix cycle_type)    │
│  - /goal ac_converge autonomy (inherited from AUTONOMY-RUN-GOAL)   │
│  - Bounded recursive self-diagnosis loop (D3):                     │
│      mechanical failure → DIAGNOSE-PATCH-VERIFY (max 3 iterations) │
│      semantic failure  → IMMEDIATE escalate (never auto-fix)       │
│  - all blocking ACs PASS + go test exit 0 + no out-of-scope        │
│    modification → auto-advance to Phase 3                          │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ auto-advance
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 3 — Sync + final independent audit (autonomous)              │
│  - manager-docs: CHANGELOG, README, frontmatter transitions        │
│  - sync-auditor: INDEPENDENT 4-dimension score (fresh context)     │
│  - moai spec audit: deterministic SPEC-compliance final check      │
│      (0 MUST-FIX drift required — REQ-IGGDA-028)                   │
│  - sync-auditor ≥ threshold + 0 MUST-FIX + git clean → IGGDA-done  │
│  - PR/main-direct → goal met → /goal clear (loop ends)             │
└─────────────────────────────────────────────────────────────────────┘
```

**Auto-advance mechanism**: the D4 Stop hook driver (`iggda-phase-driver.sh`) fires at turn-end, reads `progress.md` + invokes `moai spec audit`, and emits a `/goal`-style auto-advance signal when the phase's safe-transition predicate holds. The user does NOT write a `/goal` condition string; the goal is derived from Socratic intent and auto-converted to the condition by the driver.

---

## §B — Path B Implementation Kickoff Approval handling

### §B.1 The original invariant (Path A — what AUTONOMY-RUN-GOAL-001 preserved)

`SPEC-AUTONOMY-RUN-GOAL-001` C1 mandates: Implementation Kickoff Approval is a mandatory `AskUserQuestion` human gate at the plan→run boundary; plan-auditor PASS or score ≥ 0.90 NEVER auto-bypasses it. This invariant is grounded in:
- `run.md:122` — "[HARD] Before any run-phase autonomy ..., the orchestrator MUST have already obtained explicit Implementation Kickoff Approval approval."
- `run.md:124` — "[HARD] Implementation Kickoff Approval is score-independent."
- `orchestration-mode-selection.md:14` — "[ZONE:Frozen] [HARD] All Phase 0.95 execution modes are strictly downstream of Implementation Kickoff Approval."
- `CLAUDE.local.md §19.1` — REQ-ATR-015 mandatory-restoration.

### §B.2 The Path B amendment (what this SPEC introduces)

Path B does NOT remove the gate. It transforms the gate from a **per-run-blocking `AskUserQuestion`** into a **safe-condition predicate** with two branches:

| Branch | Trigger | Behavior |
|--------|---------|----------|
| **Auto-proceed** (REQ-IGGDA-004) | All 4 safe-conditions hold | `AskUserQuestion` STILL ISSUED (lightweight confirmation); auto-proceeds after bounded timeout if user does not veto |
| **Explicit-gate** (REQ-IGGDA-005) | Any safe-condition fails | `AskUserQuestion` REMAINS mandatory; blocks until user responds |

The 4 safe-conditions (REQ-IGGDA-004):
- **(a) Intent clarity 100%** — Socratic interview complete (CLAUDE.md §7 Rule 5).
- **(b) plan-auditor PASS** — independent audit verdict obtained.
- **(c) Tier S or Tier M** — NOT Tier L (Tier L is the complexity cutoff for "dangerous").
- **(d) No dangerous keywords AND no destructive scope** — security/payment/critical keywords absent (design.md §F.3); `--pr` flag absent; scope not marked destructive.

### §B.3 Why Path B preserves the protection (the FROZEN-spirit argument)

The Implementation Kickoff Approval invariant protects **intent verification before code is written**. The rationale: a plan-auditor PASS score is not sufficient evidence that the user's intent is faithfully captured; a human must confirm before the run-phase commits tokens to implementation.

Path B preserves this protection because **Phase 0 Socratic interview IS intent verification**:
1. **Condition (a)** requires Socratic clarity to 100% — the interview is complete, every assumption surfaced, every ambiguity resolved. This is STRONGER intent verification than a single plan→run boundary confirmation.
2. **The `AskUserQuestion` round is STILL ISSUED** in the auto-proceed branch (AC-IGGDA-004). The user retains veto authority; the gate is REDUCED (lightweight confirmation with timeout), not REMOVED.
3. **Dangerous domains are carved out** (condition (d) + Tier L). Where the stakes are high, the gate remains fully mandatory.

The change moves the human checkpoint's WEIGHT from a per-run-blocking gate to an intent-collection-time gate (Phase 0), EXCEPT in dangerous domains where the per-run-blocking gate remains.

---

## §C — The driver mechanism (moai-aware Stop hook)

### §C.1 Why a Stop hook (not a `/goal` condition string)

Claude Code's `/goal` primitive (v2.1.139+) takes a condition string the user authors. The IGGDA vision is that the user does NOT author a condition string; the goal is derived from Socratic intent. The moai-aware Stop hook (`iggda-phase-driver.sh`) is the mechanism that converts the Socratic intent + the SPEC's `progress.md` + `moai spec audit` output into a phase-transition signal.

### §C.2 The driver's evaluation loop

```
[Stop hook fires at turn-end]
  │
  ├── Is this a recovery turn? (stopReason indicates PTL/max_output_tokens/
  │   media_size/compact-failure — runtime-recovery-doctrine.md §1)
  │   ├── YES → exit 0 (Recovery-Signal Carve-Out, REQ-IGGDA-011)
  │   └── NO  → continue
  │
  ├── Read .moai/specs/<SPEC-ID>/progress.md
  │   - Extract §E.2 run-phase evidence markers
  │   - Extract §E.3 run-phase audit-ready signal
  │   - Extract §E.4 sync-phase audit-ready signal
  │
  ├── Invoke moai spec audit --json --filter-spec=<SPEC-ID>
  │   - Parse drift_findings[] (MUST-FIX count)
  │
  ├── Evaluate the current phase's safe-transition predicate:
  │   - Phase 1 → Phase 2: plan-auditor PASS + Implementation Kickoff Approval cleared
  │   - Phase 2 → Phase 3: all blocking ACs PASS + go test exit 0 + no out-of-scope mod
  │   - Phase 3 → done: sync-auditor ≥ threshold + 0 MUST-FIX + git clean
  │
  ├── Predicate holds?
  │   ├── YES → emit JSON {continue: true, phase_transition: "<from>→<to>"}
  │   │         → exit 0 (orchestrator reads JSON, advances phase)
  │   └── NO  → emit JSON {continue: false, stopReason: "<reason>",
  │                         ledger_note: "<human-readable block reason>"}
  │             → exit 2 (orchestrator translates to AskUserQuestion)
  │
  └── (never invoke AskUserQuestion directly — REQ-IGGDA-013, B11)
```

### §C.3 Why the driver reads `moai spec audit` (not frontmatter inference)

Per `verification-claim-integrity.md` §1.1 surface 3: a defect/success claim is valid only when the domain's dedicated tool confirms it. The domain here is SPEC lifecycle compliance; `moai spec audit` is the dedicated tool. Inferring phase completion from frontmatter text (`status: implemented`) alone is an unobserved claim — the audit tool's output is the evidence (REQ-IGGDA-010).

---

## §D — The bounded recursive self-diagnosis loop

### §D.1 Mechanical vs semantic failure classification

| Failure type | Examples | Loop action |
|--------------|----------|-------------|
| **Mechanical** | lint rule violation, type error, build error, missing import, formatting drift | DIAGNOSE-PATCH-VERIFY (max 3 iterations) |
| **Semantic** | data race, deadlock, panic, test assertion failure, concurrency hazard | IMMEDIATE escalate (NEVER auto-fix) |

The classification is grounded in `run.md:152` (semantic-failure escalation) + `ci-autofix-protocol.md` (mechanical-autofix bound) + runtime-recovery-doctrine.md §3 (the 5 circuit-breaker invariants).

### §D.2 The loop's compliance with the 5 circuit-breaker invariants

1. **Max-3 same-rung failures → escalate rung** (invariant 1): the loop's max-3-iteration bound (REQ-IGGDA-014) IS this invariant's projection. After 3 failed DIAGNOSE-PATCH-VERIFY iterations, the loop escalates (does NOT attempt a 4th).
2. **`hasAttemptedReactiveCompact` no-self-loop** (invariant 2): within a single turn, the loop does NOT re-attempt the same DIAGNOSE-PATCH-VERIFY if it already failed this turn. Each iteration is a new turn.
3. **Compact-can-PTL last-resort escape** (invariant 3): if the loop itself hits PTL (the recovery triggers the error it's recovering from), the loop falls to rung-4 abort + preserve (persist to `progress.md`), NOT another patch.
4. **Abort-closes-ledger** (invariant 4): on halt (iteration 3 OR semantic failure OR PTL), the loop persists its state to `progress.md §E Recursive Self-Diagnosis Log` before the session ends.
5. **Narrative-consistency** (invariant 5): across compact/recovery boundaries, the loop's state is reported via the 5-Section Evidence-Bearing Report format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per `verification-claim-integrity.md` §3).

### §D.3 Why max-3 (not max-5 or max-10)

The max-3 bound is inherited from two sources:
- `ci-autofix-protocol.md` — the autofix loop's per-PR-push max-3-iteration contract.
- runtime-recovery-doctrine.md §3 invariant 1 — `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3` (book1 ch06 §5).

Three iterations is the empirically-validated threshold where a mechanical failure is either patched (iterations 1–2 typically succeed) or is actually semantic-misclassified (iteration 3 fails → escalate, and the human re-classifies). Raising the bound risks the death-spiral; lowering it under-fixes easily-patched mechanical failures.

---

## §E — Independent-audit preservation (why "self-audit" ≠ "independent audit" removal)

### §E.1 The bias-prevention rationale

The independent auditors (plan-auditor, sync-auditor) exist because a first-party verification (the implementer checking their own work) is biased. The implementer's context carries the assumptions that produced the implementation; the auditor's fresh context catches what the implementer rationalized away. This is the "skeptical evaluation stance" in `agent-common-protocol.md` § Skeptical Evaluation Stance.

### §E.2 What IGGDA preserves

IGGDA's autonomous execution does NOT collapse the independent auditor into the implementer:
- **Phase 1**: plan-auditor is spawned via `Agent(subagent_type: "plan-auditor")` in a FRESH context — NOT a continuation of manager-spec's turn (REQ-IGGDA-021, AC-IGGDA-027).
- **Phase 3**: sync-auditor is spawned via `Agent(subagent_type: "sync-auditor")` in a FRESH context — NOT a continuation of manager-docs's turn (AC-IGGDA-028).

### §E.3 The disambiguation

- **"Self-audit"** in IGGDA = D3's recursive self-diagnosis loop performing first-party verification of MECHANICAL failures during Phase 2. It is a code-quality loop, NOT a SPEC-quality audit.
- **"Independent audit"** in IGGDA = plan-auditor (Phase 1) + sync-auditor (Phase 3) in fresh contexts. These are the SPEC-quality audits with bias-prevention guarantees.

The two are COMPLEMENTARY: self-audit handles mechanical code failures fast (bounded loop, no human in the loop for easy cases); independent audit handles SPEC-quality assurance (human-grade skeptical evaluation). IGGDA does NOT trade one for the other (REQ-IGGDA-022, AC-IGGDA-030).

---

## §F — FROZEN invariant analysis (LOAD-BEARING — the FROZEN-amend justification)

> This section is the rigorous justification for amending a `[ZONE:Frozen] [HARD]` invariant. The plan-auditor MUST scrutinize this section. If the argument here is flawed, the SPEC must be rejected.

### §F.1 What the invariant protects

The Implementation Kickoff Approval invariant (run.md:122,124; orchestration-mode-selection.md:14; CLAUDE.local.md §19.1 REQ-ATR-015) protects **intent verification before code is written**. The protected value is:

1. **The user's intent is faithfully captured** before the run-phase commits tokens to implementation.
2. **A plan-auditor PASS score is NOT sufficient evidence** of intent fidelity — a human must confirm.
3. **The gate is score-independent** — even a high skip-eligible (≥ 0.90) plan-auditor score does NOT auto-bypass the gate, because score measures plan QUALITY, not intent FIDELITY.

The invariant's historical motivation: a prior incident where a plan-auditor PASS score was used to auto-bypass the plan→run gate, and the implementation proceeded on a mis-captured intent, wasting significant tokens before the mis-capture was discovered. The gate was hardened to FROZEN to prevent recurrence.

### §F.2 Why Path B preserves the protection

Path B's claim: the Socratic interview at Phase 0 IS intent verification, and is STRONGER than a single plan→run boundary confirmation. The argument:

1. **Condition (a) — intent clarity 100%** — the Socratic interview (CLAUDE.md §7 Rule 5 + askuser-protocol.md § Socratic Interview Structure) drains intent to 100% clarity. This is a MULTI-ROUND interview (max 4 questions per round, multiple rounds until clarity), far more rigorous than a single plan→run confirmation.
2. **The `AskUserQuestion` round is STILL ISSUED in the auto-proceed branch** (AC-IGGDA-004). The user sees the confirmation; the user CAN veto within the timeout window. The gate is REDUCED (lightweight confirmation + timeout auto-proceed), not REMOVED. This preserves the user's veto authority — the FROZEN-spirit guarantee.
3. **Dangerous domains are carved out** (conditions (c) Tier L + (d) security/payment/critical keywords + destructive scope). Where the stakes are high, the gate remains fully mandatory (REQ-IGGDA-005). The amendment is NARROW — it applies only where the Socratic interview has drained intent AND the domain is safe.
4. **The decision is auditable** (REQ-IGGDA-006). Every auto-proceed is logged to `progress.md §E IGGDA Kickoff Predicate` with all 4 conditions + the verdict. Post-hoc verification is possible; a miscaptured intent that slipped through auto-proceed is detectable and reviewable.

### §F.3 The safe-condition keyword list (security / payment / critical)

The keyword list is the determinator of condition (d). It is enumerable at design time and extensible via rule-file edit (not code change). Initial list:

**Security domain** (any match → explicit-gate):
- `auth`, `authentication`, `authorization`, `acl`
- `secret`, `credential`, `password`, `token`, `api_key`, `apikey`
- `crypto`, `encryption`, `decrypt`, `hash`, `salt`
- `session`, `cookie`, `jwt`, `oauth`, `saml`, `sso`
- `injection`, `xss`, `csrf`, `sqli`, `rce`
- `vulnerability`, `cve`, `owasp`

**Payment domain** (any match → explicit-gate):
- `payment`, `billing`, `invoice`, `charge`
- `stripe`, `paypal`, `portone`, `iamport`, `toss`, `kakopay`, `naverpay`
- `card`, `pan`, `pci`, `dss`
- `refund`, `settlement`, `pgp` (in payment context)

**Critical domain** (any match → explicit-gate):
- `production`, `prod`, `live`
- `migration`, `rollback`, `drop_table`, `drop table`
- `force_push`, `force-push`
- `rm_rf`, `rm -rf`
- `database`, `schema`, `migration` (when paired with `drop` / `alter`)

**Destructive scope markers** (any match → explicit-gate):
- `--pr` flag supplied
- SPEC scope marked `destructive: true` (a frontmatter field this SPEC does NOT introduce — out of scope; detected via PR-strategy preference)

**Maintenance**: the list is maintained in `orchestration-mode-selection.md §F.3` and extended via SPEC amendment. The list is intentionally OVER-INCLUSIVE in v0.1.0 (better to force explicit-gate on a false-positive than auto-proceed on a false-negative).

### §F.4 The auto-proceed timeout

REQ-IGGDA-004 says the lightweight confirmation "auto-proceeds after a bounded timeout". Proposed timeout: **30 seconds**.

- **Long enough** for a human to read the confirmation and veto (a typical `AskUserQuestion` round response time is 5–15 seconds; 30 seconds gives 2x margin).
- **Short enough** to not stall autonomy (the goal of IGGDA is to keep moving; a 30-second pause is acceptable, a 5-minute pause defeats the purpose).

The timeout is configurable via `.moai/config/sections/workflow.yaml` (a NEW field `iggda.kickoff_timeout_seconds: 30`). Deferred to plan.md M1.

### §F.5 Domains NEVER auto-proceeded (summary)

| Domain | Why never auto-proceeded |
|--------|--------------------------|
| **Tier L SPECs** | Complexity cutoff — Tier L work is high-stakes by definition |
| **Security** (auth, secrets, crypto, injection) | Mis-captured intent in security code produces vulnerabilities |
| **Payment** (billing, card, PCI) | Mis-captured intent in payment code produces financial loss |
| **Critical infra** (prod, migration, drop table, force-push) | Mis-captured intent in infra code produces outages |
| **Destructive scope** (`--pr`, `rm -rf`) | Destructive operations are irreversible |
| **Incomplete Socratic intent** (clarity < 100%) | The interview must complete first — no gate bypass on incomplete intent |
| **plan-auditor FAIL/INCONCLUSIVE** | Independent audit is the final guarantee — FAIL halts regardless |

### §F.6 Relationship to Anthropic's plan-editor mandate

Claude Code's Ctrl+G plan editor mandate (the platform-level plan-to-implement human gate) is the Anthropic-platform analogue of Implementation Kickoff Approval. The mandate requires that before the model enters implementation, the user has reviewed and approved the plan.

**Path B's relationship**:
- Path B does NOT claim to bypass the Anthropic plan-editor mandate. The mandate is a platform-level mechanism that fires regardless of MoAI's rule layer.
- Path B AMENDS MoAI's rule-layer Implementation Kickoff Approval (the moai-native analogue), reducing its per-run-blocking weight in safe domains. The Anthropic mandate (if it fires) still applies on top.
- In practice, when the safe-condition predicate auto-proceeds, the user has ALREADY reviewed the plan via the Socratic interview + plan-auditor verdict surfaced in Phase 0/1. The Anthropic mandate's intent (user reviewed before implementation) is satisfied by the Socratic interview, which is upstream of the predicate.

**Honest flag**: the Anthropic plan-editor mandate's exact firing conditions are not fully documented; if the mandate fires AFTER MoAI's auto-proceed, there is no conflict (the user gets a second chance via the platform gate). If the mandate does NOT fire (some surfaces), the MoAI auto-proceed stands on its own — and the Phase 0 Socratic interview is the sole intent verification. This is acceptable because condition (a) requires 100% clarity, which IS the intent verification.

**Reactive vs preventive prevention (D10 resolution)**: the plan-editor honest-flag mitigation above is **reactive** (FL-1 post-hoc measurement on the auto-proceed/explicit-gate ratio) — it measures drift after the fact, it does not prevent a misfired auto-proceed on a surface where the Anthropic mandate is known NOT to fire. The **preventive complement** is the Phase 0 Socratic multi-round intent-clarification itself: condition (a) of the compound safe-condition predicate (REQ-IGGDA-004) requires 100% intent clarity via a multi-round Socratic interview (max 4 questions per round, multiple rounds until clarity per `askuser-protocol.md` § Socratic Interview Structure), which IS the intent-verification the plan-editor mandate exists to enforce. The Socratic interview runs upstream of the predicate on ALL runtime surfaces (TUI, headless `-p`, SDK) because it is a MoAI rule-layer mechanism, not a platform-layer mechanism — so the preventive intent-verification guarantee holds even where the platform-layer mandate does not fire.

**Known reactive-only zone**: headless `claude -p` without `--permission-mode acceptEdits` AND the SDK surface are documented (per `.claude/rules/moai/workflow/dynamic-workflows.md` § How a Workflow Runs) as surfaces where the Anthropic plan-editor mandate's firing behavior is not fully documented. On these surfaces, the plan-editor honest-flag is reactive-only (no platform-layer preventive gate is guaranteed to fire alongside MoAI's auto-proceed). The explicit mitigation on these surfaces is that the compound safe-condition predicate REQ-IGGDA-004 STILL binds in full — conditions (a) intent clarity 100%, (b) plan-auditor PASS, (c) Tier S/M, (d) no dangerous keywords/destructive scope must ALL hold before auto-proceed, and the `AskUserQuestion` round is STILL ISSUED (AC-IGGDA-004) with the user's veto authority preserved. The destructive-domain carve-out (REQ-IGGDA-005) forces explicit-gate regardless of surface. A follow-up SPEC (`SPEC-V3R6-IGGDA-PREFLIGHT-001` candidate, EX-1) may add runtime-surface detection that force-disables auto-proceed on surfaces where the mandate is known not to fire; until then, the compound predicate + Socratic preventive complement is the documented mitigation.

---

## §G — Independent-audit preservation architecture (diagram)

```
┌─────────────────┐         ┌─────────────────────┐
│  manager-spec   │  spawn  │   plan-auditor      │
│  (Phase 1)      │ ───────▶│   (FRESH CONTEXT)   │
│  implementer    │         │   independent audit │
└─────────────────┘         └─────────────────────┘
        NOT a continuation          bias prevention
        of the same turn            guaranteed

┌─────────────────┐         ┌─────────────────────┐
│  manager-docs   │  spawn  │   sync-auditor      │
│  (Phase 3)      │ ───────▶│   (FRESH CONTEXT)   │
│  implementer    │         │   independent audit │
└─────────────────┘         └─────────────────────┘
        NOT a continuation          bias prevention
        of the same turn            guaranteed

┌─────────────────────────────────────────────────┐
│  manager-develop (Phase 2)                      │
│  + recursive self-diagnosis sub-agent           │
│  = "self-audit" (first-party, MECHANICAL only)  │
│  NOT a replacement for independent audit        │
└─────────────────────────────────────────────────┘
```

The two audit types are COMPLEMENTARY:
- **Self-audit** (Phase 2): fast, bounded, handles mechanical code failures. No bias prevention (it's the implementer checking their own mechanical work — acceptable for lint/type/import errors).
- **Independent audit** (Phase 1 + Phase 3): slow, fresh-context, handles SPEC-quality assurance. Bias prevention guaranteed.

---

## §H — Design alternatives considered (and rejected)

### §H.1 Alternative A — Path A (no amendment; preserve Implementation Kickoff Approval verbatim)

**Rejected by user**. The user explicitly chose Path B. Path A is what `SPEC-AUTONOMY-RUN-GOAL-001` delivered; it preserves the invariant but does not deliver the IGGDA vision (autonomy with no human intervention after intent collection).

### §H.2 Alternative B — Remove Implementation Kickoff Approval entirely

**Rejected**. This would violate the FROZEN invariant without justification. The protection the invariant provides (intent verification before code) is valuable; removing it entirely would regress to the prior incident that motivated the invariant. Path B's safe-condition predicate is the NARROW alternative.

### §H.3 Alternative C — Stop hook driver as a `/goal` condition string the user authors

**Rejected**. The IGGDA vision is that the user does NOT author a condition string. The driver auto-converts Socratic intent + `progress.md` + `moai spec audit` into the phase-transition signal. Requiring the user to author a condition string would re-introduce a human-in-the-loop step that defeats the autonomy vision.

### §H.4 Alternative D — Recursive loop unbounded (max ∞ iterations)

**Rejected**. An unbounded loop risks the death-spiral (runtime-recovery-doctrine.md §3 invariant 1). The max-3 bound is the empirically-validated threshold; raising it requires a SPEC amendment with evidence.

### §H.5 Alternative E — Collapse independent auditor into implementer (single agent does implementation + audit)

**Rejected**. This destroys bias prevention (`agent-common-protocol.md` § Skeptical Evaluation Stance). The independent auditor's fresh context is the bias-prevention mechanism; collapsing it defeats the purpose.

---

## §I — Risks and mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Safe-condition predicate is too permissive (auto-proceeds on mis-captured intent) | HIGH | (1) Condition (a) requires 100% Socratic clarity; (2) `AskUserQuestion` still issued with veto; (3) FL-1 post-implementation measurement; (4) narrow dangerous-domain carve-out |
| Recursive loop mis-classifies semantic failure as mechanical | MEDIUM | M2 regression test for the classifier; EC-2 edge case; test assertion failure explicitly enumerated as semantic (REQ-IGGDA-016) |
| Stop hook driver fires on a compact-recovery turn and blocks | MEDIUM | Recovery-Signal Carve-Out (AC-IGGDA-014); driver exits 0 on recovery-turn `stopReason` |
| `moai spec audit --filter-spec` flag does not exist | LOW | M5 adds it (minor Go); graceful degradation in M3 if M5 not yet done |
| Template neutrality CI guard fails (internal SPEC IDs leak) | LOW | §D.2 constraint; AC-IGGDA-038; keyword list is generic prose |
| FROZEN-invariant amendment rejected by plan-auditor | HIGH | This design.md §F is the rigorous justification; if plan-auditor rejects, the SPEC returns to draft for §F revision (the amendment is reversible — restore Path A) |

---

## §J — Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — defect/success claims require the dedicated tool (moai spec audit compliance)
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §2,§3,§4 — cheapest-first ladder + 5 circuit-breaker invariants + Recovery-Signal Carve-Out
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` semantics
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.3 — Mode 6 capability gate
- `.claude/skills/moai/workflows/run.md:116-171` — existing run-phase autonomy (sibling)
- `.moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/design.md` — sibling design (run-phase autonomy)
- `CLAUDE.local.md §19.1` — REQ-ATR-015 mandatory-restoration (the FROZEN invariant owner)

---

Version: 0.1.0 (plan-phase)
Status: Active — awaits plan-auditor independent audit. The plan-auditor's primary scrutiny target is §F (FROZEN-invariant analysis).
