# SPEC-UPDATE-HOOK-DELIVERY-001 — Design (OPEN decision axes)

> **Status: the resolution option is an OPEN decision.** This document frames the axes with evidence and trade-offs; it concludes NOTHING. The operator selects the option at the Implementation Kickoff Approval gate; the verdict is recorded in progress.md at run-phase M1 and (body edit re-delegated per D-NEW-1) appended to §G here.

## §A The decision question

When the shipped template gains a hook entry inside a hook event key the user already carries, what should `moai update` do?

- **Option A — deliver**: merge new hook entries into the user's settings.json.
- **Option B — detect + guide only**: leave update behavior unchanged; `moai doctor` (or the update output) detects template hook entries missing from the user's file and reports them with guidance.
- **Option C — explicit no-op + docs**: change nothing in behavior; document the limitation in user-facing docs (and optionally a one-line update-output notice).

Shared evidence: today the gap is silent and undetected (research.md §A facts 1-3). Even Option C converts a silent drop into a documented one — the floor any option must clear is REQ-UHD-004's "resolve the gap by exactly one operator-selected mechanism" mandate; the shorthand "no silent drop remains undetected AND unreported" is this document's paraphrase of that mandate, not its wording.

## §B Option A — deliver

**What it does**: the merge path becomes identity-aware at the array-entry level for hook event keys: entries present in the template but absent from the user's file are added; entries the user removed stay removed.

**Resurrection-avoidance is the hard half.** Three identity-scheme axes (choose ONE at M1):

| Scheme | Mechanism | Resurrection avoidance | Cost / failure mode |
|---|---|---|---|
| A1 — stable entry-identity key | Each template hook entry carries an identity field (e.g. an `id` marker or convention-derived identity from matcher+command). Base derivation recurses into hook arrays keyed by identity: entry in base+user, absent in user = user-deleted (keep deleted); entry in template, absent in base+user = template-new (add). | Strong: deletion is observed as identity present-in-base, absent-in-user. Requires base to actually contain the entry — but today's base omits array granularity entirely, so base derivation itself must become array-aware (the core code change). | Adds a visible field to shipped settings.json (template-content decision; neutrality rules apply); entries lacking identity (user-authored, legacy) need a fallback rule (treat as user-owned, never touch). |
| A2 — tombstone in user file | A deleted entry leaves a marker (e.g. entry re-serialized under a `_deleted` side-list or comment sentinel in settings.json). Update skips tombstoned identities. | Strong once a tombstone exists — but the FIRST deletion after the scheme ships must be detectable as deletion, not absence. If the user deleted entries before the scheme existed, update cannot distinguish "deleted" from "never seen" and A2 degenerates toward A3. | Pollutes the user's settings.json with moai bookkeeping (Claude Code may rewrite/stripe the file — unknown durability, a real risk given settings.json is user-facing config Claude Code itself edits). |
| A3 — seen-registry sidecar | State file under `.moai/` (e.g. `.moai/state/hook-delivery.json`) records template entry identities already delivered/seen. Update adds only entries whose identity is not in the registry; user deletions are honored because a seen entry absent from the user file is never re-added. | Strong: "seen once, never force-again" — deletion is respected by construction. No pollution of settings.json itself. | State lives outside settings.json; `CleanMoaiManagedPaths` wipes `.moai/state/` on update (CLAUDE.local.md §2.3) unless the registry path is placed outside wiped roots or the wipe is reconciled — a deployment-coupling hazard that must be resolved before A3 is selectable. |

**Option A failure modes**: identity-scheme complexity (three variants above, each with a distinct coupling); wrong-identity collisions re-add or skip wrongly; template renames of an entry read as delete+add; behavior divergence from what Claude Code itself does to settings.json.

**Option A payoff**: users get shipped hook capabilities automatically — the only option that closes the capability gap without user action. This is why the defect matters: hook entries ARE capabilities (SessionStart wiring, drain triggers), and a silent non-delivery means moai-adk upgrades ship features that never activate for existing projects.

## §C Option B — detect + guide only

**What it does**: `moai update` behavior unchanged; a doctor check (and/or update-output line) diffs the template hook set against the user's file and lists missing entries with per-event-key guidance for manual adoption.

**Failure modes**: drift between shipped template and user file keeps growing (nothing converges it — every release with new hooks re-raises the report); users on automated/CI update flows never read the report (detection without delivery is silent in headless contexts); the doctor check needs the same entry-identity reasoning to name entries precisely, so it pays part of Option A's identity cost without A's payoff; "reported" ≠ "resolved" — the card's underlying complaint (capability never activates) survives as a manual chore.

**Option B payoff**: near-zero risk — no write path, no resurrection hazard, no settings.json coupling. Smallest safe increment; composable with C and can precede A in a later SPEC.

## §D Option C — explicit no-op + docs

**What it does**: nothing behavioral; user-facing documentation states which hook entry classes are not delivered on update and how to adopt them manually.

**Failure modes**: the drift still grows; documentation rot (the statement must track every future template hook change to stay true); the weakest reading of REQ-UHD-004 — defensible only if the evidence shows delivery is too risky for the release window and B's detection surface is not wanted either.

**Option C payoff**: zero code risk, zero identity machinery; honest instead of silent.

## §E Cross-cutting constraints on ANY option

- REQ-UHD-001/002/003 hold: template-new event keys keep delivering; user deletions keep being honored; user-modified entries keep surviving. Any option's mechanism must pass AC-UHD-001/002/011 (guards).
- Claude Code itself rewrites settings.json — any in-file bookkeeping (A2) must survive that rewrite or fail visibly; this is an unmeasured environment property (see §F).
- The `.moai/state/` wipe coupling (CLAUDE.local.md §2.3) binds any sidecar state (A3).

## §F Open questions feeding the gate (NOT concluded here)

1. Which identity scheme (A1/A2/A3) if Option A? [NEEDS CLARIFICATION: identity scheme — gated at Implementation Kickoff Approval]
2. Is settings.json in-file bookkeeping durable under Claude Code's own rewrites? Unmeasured — bears directly on A2.
3. Is one-shot detection (B) acceptable as a first increment with A deferred, or does the capability gap require A now?
4. Does the release window tolerate the base-derivation change (array-aware pruning) that A requires?

## §G Resolution (populated at run-phase M1)

_pending — the operator's option selection and, for Option A, the chosen identity scheme are recorded here at M1._
