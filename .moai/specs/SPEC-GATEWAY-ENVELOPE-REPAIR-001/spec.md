---
id: SPEC-GATEWAY-ENVELOPE-REPAIR-001
title: "Reasoning-envelope repair — launcher-side verbatim re-injection of gateway-issued envelopes for stripped-replay recovery"
version: "0.2.2"
status: completed
created: 2026-09-13
updated: 2026-09-14
author: manager-spec (card t708)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/gateway/conversation, internal/gateway/translate, internal/gateway/receipt, internal/gateway/opaque"
lifecycle: spec-anchored
tags: "gateway, receipt-chain, reasoning-envelope, opaque, repair, security-determination, gpt"
tier: M
related_specs:
  - SPEC-GATEWAY-WEDGE-REROOT-001
  - SPEC-MOAI-GATEWAY-001
---

# SPEC: Reasoning-envelope repair — the launcher-side recovery path for the t703 stripped-replay shape

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-13 | manager-spec | Initial plan-phase emission (card t708, Class C). Evidence base: `.moai/reports/t708/research.md` (4-lens synthesis + orchestrator addendum), `.moai/reports/t703/verdict.md` (incident ground truth), `.moai/reports/t672/{verdict,investigation,matrix}.md` (binding invariant). Sibling SPEC `SPEC-GATEWAY-WEDGE-REROOT-001` (card t700, draft, sibling worktree) cross-referenced by path + requirement IDs; the interpretive seam between the two SPECs is surfaced in §4, not resolved silently. Symbol pins verified on tree `WT-envelope-persist` (base local develop `7a7a08f20`). |
| 0.2.0 | 2026-09-13 | manager-spec | Plan-audit iteration-1 repair pass (card t708; iter-1 COND-FAIL 0.8375, report `.moai/reports/t708/plan-audit.md`; the §3 security determination was verified SOUND against code by the auditor). D1: AC-EVR-003 added to the acceptance.md §D.1 traceability matrix (was an orphan row). D2: plan.md §C pre-flight `refreshNative` grep corrected to its method form (rc=1 as written). D4: REQ-EVR-003-1 aligned with the all-or-nothing refusal semantics of REQ-EVR-007/AC-EVR-005 — a non-self-attesting boundary aborts the entire repair. D3: §3.4 relocated before §4 (H2 scan order). D5: carrier count restated per the research addendum measurement (36 main transcript, 37 incl. a subagent-row boundary). New lead design input folded in: the serial-turn `CauseChain` 400 class (production family `1f14d174`, card t707 reproducing) declared out of scope with a §3.4 row; BerriAI/litellm #40288 recorded as external byte-stability prior art in §4.1; plan.md gains a third bounded clarification entry and M0's adjudication gate extends to t707's verdict as a second input. |
| 0.2.1 | 2026-09-13 | manager-spec | Clarification-resolution pass per lead conditional-Kickoff directive (2026-09-13): plan.md §H converted from three open clarification entries to a zero-entry Resolution Record (topic / DECISION / basis / route per row; dispositions ARE the acceptance.md §D.4-3 clause). No design change — every disposition follows the audited bounded paths (REQ-EVR-007 refusals, M0 two-input gate). Cross-references updated (spec.md §3.3/§4.1, plan.md §A/§I, acceptance.md §D.4-3). |
| 0.2.2 | 2026-09-14 | manager-spec | Composite-refusal semantics correction per sync-audit F7 (`.moai/reports/t708/sync-audit.md`; no design change). §3.3 restated honestly: the repair-layer refusal triggers are ONLY source-gone / digest-mismatch / already-attempted / incomplete-transcript; the missing-marker shape (client-side indistinguishable from a text-only turn) and the Prefix-class public-content-mismatch shape (undetectable without the receipt reads REQ-EVR-003-4 forbids) have COMPOSITE semantics — the repair injects (or skips) on marker-attested boundaries only and the unchanged Check delivers the final classified refusal at request time. AC-EVR-007's letter aligned to the same split. §3.4 wording touched minimally for the same honesty. |

## 1. Problem — measured shape

### 1.1 The stripped-replay mechanism (proven, card t703)

On a mid-conversation model switch within a GPT history family (sol → luna) with tool-bearing history, the external client (Claude Code, outside this repository) re-encodes the replayed history at request-encode time: the gateway-issued `redacted_thinking` envelope blocks are **stripped from the outgoing request** while the bound `toolu_moai_v1_*` tool markers **survive**. The next replay fails with 400 at the opaque-marker-without-envelope site in `receiptHistory.observations()` (`internal/gateway/translate/receipt_history.go`, the `strings.HasPrefix(id, opaque.ToolPrefix)` branch), classified `CauseReasoning` by t703's fix. The client's transcript-record path keeps the exact issued bytes — measured directly on the t703 incident family: 36 `moai_opaque_v2_*` carriers in the main transcript (37 including a subagent-row boundary) retained post-switch, per the research addendum measurement (`.moai/reports/t708/research.md`). The conversation dies on the first request of the new model, and the only in-repo remedy today is the classified rejection body.

Two facts bound the problem precisely:

1. **The gateway cannot make the stripped replay succeed.** Accepting an envelope-less replay would require relaxing the empty fallback against Required candidates or introducing item-level existence checks — both rejected by t672 on security grounds (`.moai/reports/t672/investigation.md` §2 Candidate A/Desync). It would also be untranslatable: the forward path needs the envelope to restore tool ids and re-inject upstream reasoning (`internal/gateway/translate/request.go`, the tool-marker restore and reasoning re-injection splices). The t672 invariant — an envelope-less replay of a Required boundary stays rejected — is locked and not reopened here.
2. **The bytes that would satisfy the unchanged validator still exist, client-side.** The envelope is fully client-visible at issuance (non-stream `content[0]` and stream terminal `content_block_start` both carry the full `moai_opaque_*` data carrier — `internal/gateway/translate/response.go` / `internal/gateway/translate/stream.go` terminal re-emit), and the family's native transcript retains it verbatim after the switch turn (research addendum, measured). The launcher already reads that transcript family today (`refreshNative`, `internal/gateway/conversation/native.go`).

### 1.2 What t703 and t672 shipped and what they left open

t672 shipped the four-way binding analysis and cause classification (`ReplayCause`: chain / lineage / reasoning) — no recovery. t703 shipped classification of the observations-level rejection sites so the stripped-replay 400 self-identifies (`CauseReasoning` with recovery guidance) — also no recovery. Both verdicts name the residual: a switch turn with tool-bearing history still kills the conversation on the new model's first request. This card is the recovery candidate for exactly that shape.

### 1.3 Constraints inherited from t672/t703 (policy-space bounds — violating any is an automatic design reject)

1. **No acceptance relaxation.** The validator (`receiptHistory.Check` → `Manifest.Check`), the publish flow, and the error contract stay byte-identical (t672 invariant; t700 REQ-WRR-001/002 pattern). Verification of any repair is done by the UNCHANGED Check at request time — never by a repair-side reimplementation of the binding.
2. **No item-level validation, no request-driven receipt-root selection/seeding, no envelope re-issue or synthesis** (t672 §2; t700 forbidden directions). Reappearance of any of these in any route is an automatic reject.
3. **The fixed-guidance error contract is frozen**: `historyReplayGuidance` and the t672/t703 reason clauses stay byte-identical on every rejection body.
4. **Launcher-driven, never request-driven**: `NewReceiptHistory` accepts only a launcher-authorized UUID and private store; request metadata never selects a receipt root or authorizes a conversation. Any repair path lives on the launcher surface, never in the request channel's authority.
5. **No background load** in run phase: t672's isolation discipline (ephemeral loopback ports, per-run session tokens, isolated stores under `/tmp`, bounded upstream cost, cleanup-guaranteed child shutdown) applies to any live probe.

### 1.4 Sibling-card sequencing (t700 → t708)

Sibling card **t700** (`SPEC-GATEWAY-WEDGE-REROOT-001`) is concurrently authored in a parallel worktree and is **unlanded** (status: draft) at this SPEC's authoring time. This SPEC pins **symbol names, never line numbers** (§2 anchors, plan.md §A), cross-references t700 by path and requirement IDs, and its run phase MUST first absorb `develop` per the git-flow lane protocol and **re-verify every symbol pin against the absorbed tree before the first code commit** — including whether t700's artifacts have landed and whether its requirement IDs moved. Merge order: **t700 → t708** is preferred but not load-bearing for correctness; the load-bearing gate is the pin re-verification.

## 2. Requirements (GEARS)

### 2.1 User story

As a user of a long-lived GPT-gateway conversation, when a mid-conversation model switch wedges my conversation (the client stripped the gateway-issued reasoning envelopes from its replay and every request is rejected at the reasoning cause), I want an explicitly-invoked, bounded, non-destructive launcher-side repair that re-injects the exact envelopes the gateway itself issued at their original boundaries, so the conversation survives — without the gateway accepting one replay it would not have accepted before, without any manufactured reasoning content, and without any path that lets request content touch gateway receipt state.

### 2.2 Requirements

- **REQ-EVR-001** (Ubiquitous — authorization-invariant lock): **The receipt-chain validator's accept/reject behavior shall remain unchanged by this SPEC** — `receiptHistory.Check` → `Manifest.Check` accepts exactly the replayed histories it accepts today and rejects exactly the histories it rejects today, with `ReplayCause` classification unchanged; the publish flow is untouched. This card adds a repair *procedure around* the validator; it does not touch the validator. A repaired replay is verified by the unchanged Check at request time, never by repair-side logic.
  - Anchors: `receiptHistory.Check`, `checkObserved`, `replayCause`, `observations`, `Publish` (`internal/gateway/translate/receipt_history.go`); `Manifest.Check` (`internal/gateway/receipt/core.go`).

- **REQ-EVR-002** (Ubiquitous — byte-exactness): **The repair path shall inject only envelope bytes previously issued by the gateway** — the `moai_opaque_*` data carrier read verbatim from the family's native transcript source — and shall never mint, re-encode, re-serialize, wrap, or synthesize an envelope. For every injected envelope, `sha256(raw)` shall equal the digest embedded in the orphaned tool marker at that boundary (`BindToolID`-embedded `opaque_sha256`, `internal/gateway/opaque/codec.go`), and the carrier's decoded canonical bytes shall be byte-identical to a form the gateway's own `opaque.Decode` accepts without alternate-spelling tolerance (`bytes.Equal(e.raw, raw)`).
  - Anchors: `opaque.Decode`, `BindToolID`, `RestoreToolID` (`internal/gateway/opaque/codec.go`); `outputEnvelopeWithRaw`, `envelope.Data()` (`internal/gateway/translate/reasoning.go`, `internal/gateway/translate/response.go`).

- **REQ-EVR-003** (Capability gate — position exactness and marker self-attestation): **Where** the repair path injects an envelope at an assistant boundary, **the repair path** shall satisfy ALL of:
  1. the boundary carries a surviving `toolu_moai_v1_*` marker whose embedded digest matches the injected envelope's digest (marker self-attestation) — a boundary without a self-attesting surviving marker aborts the entire repair with zero modification (all-or-nothing per attempt, per REQ-EVR-007), never a partial repair of the remaining boundaries;
  2. injection happens at the envelope's original boundary position only — no reordering, no relocation across boundaries, no injection at a boundary the transcript does not prove was envelope-bearing;
  3. no public content is edited in the same operation — the repair changes thinking-carried opaque content only, and never modifies any non-thinking block (user turns, tool results, text, tool-use blocks);
  4. the repair decision reads no receipt store — admissibility is decided by marker self-attestation plus the gateway's unchanged Check at request time, so the repair path gains no dependency on hash-only manifests (`internal/gateway/receipt/store.go`) and no read handle on gateway state.

- **REQ-EVR-004** (Event-driven — user-invoked repair): **When** a conversation replay is rejected with the reasoning-envelope cause (a 400 classified `CauseReasoning`) and the user explicitly invokes the repair, **the launcher-side repair path** (operating from the launcher conversation flow in `internal/cli/gateway_session.go`, with the family's native transcript — already read by `refreshNative`, `internal/gateway/conversation/native.go` — as the capture source) shall re-inject the matching gateway-issued envelope bytes at their original boundaries per REQ-EVR-002/003 and permit a single retry of the conversation.
  - Anchor: the classified rejection shape proven by t703 (`.moai/reports/t703/verdict.md` — `CauseReasoning` at the marker-without-envelope site).

- **REQ-EVR-005** (Unwanted — no automatic surgery): **The repair path shall not run automatically on any 400 rejection.** A `CauseReasoning` classification alone shall never modify any replayed history, transcript, or payload without explicit user action (the classification identifies the shape; a human authorizes the repair).

- **REQ-EVR-006** (Capability + State + Event compound — bounded termination): **Where** the repair capability is present **While** a repair attempt is durably recorded on the launcher conversation record (attempted flag plus the per-boundary provenance of REQ-EVR-008), **When** the retry after that single repair attempt is still rejected, **the repair path** shall terminate recovery, preserve its non-destructive state, and surface the gateway's classified guidance unchanged — it shall not retry again, inject further, or fall back to any other history modification. The bound is read from the durable record, so termination survives process restarts; there is no inject-until-accepted iteration.

- **REQ-EVR-007** (Event-detected — refusal on non-repairable preconditions): **When** any repair precondition fails — the transcript source no longer retains a verbatim envelope for a stripped boundary (client compaction, `--continue` slicing, or record loss), the marker self-attestation mismatches the candidate envelope's digest, a stripped boundary lacks a surviving self-attesting marker, or the history exhibits a Prefix-class mismatch indicating public content changed on the same replay — **the repair path** shall refuse with zero modification of any history, payload, or transcript, leave the non-destructive state intact, and surface the gateway's classified guidance unchanged.
  - Rationale: the public-content-unchanged precondition is enforced by refusal + Check adjudication (the unchanged Check fails at the Prefix site regardless), so the unmeasured question of whether the t703 switch turn also altered public content is bounded either way — the answer changes only refusal frequency, never the design (the two bounded measurement questions are dispositioned in plan.md §H Resolution Record).

- **REQ-EVR-008** (Capability gate — non-destructive provenance): **Where** the repair path performs an injection, **it** shall (a) preserve the unmodified history aside **before** any repaired form is used or persisted (aside-before-replace), keeping it recoverable and never hard-deleted; (b) durably record what was repaired — per-boundary digest and boundary position — on the launcher conversation record; and (c) leave every public content byte identical to the pre-repair history, so the repair's entire diff against the prior replay is confined to thinking-carried opaque content that the gateway itself issued.

- **REQ-EVR-009** (Ubiquitous — gateway-state and channel invariance): **The repair path shall never create, seed, select, mutate, or read any receipt store, receipt root, receipt candidate, or lineage, and shall never be triggered by, keyed from, or authorized by request metadata.** The repair lives on the launcher surface only; the request channel remains non-authoritative; the sanctioned fork path (`conversation.Manager.Fork`) remains the only gateway-state-changing recovery and is untouched by this card.

- **REQ-EVR-010** (Ubiquitous — operator documentation): **The stripped-replay shape, the repair procedure, its bounds (explicit invocation, single-shot, byte-exact, non-destructive), its refusal conditions, the non-repairable shapes, and the sanctioned fork path for lineage misses shall be documented for operators.** The error body keeps pointing at the frozen fixed guidance; operator docs carry the repair path and its limits.

## 3. Security determination — why verbatim re-injection satisfies the t672 binding rather than bypassing it

### 3.1 The property being protected

The receipt chain is the gateway's history-authorization primitive: `Manifest.Check` requires every replayed assistant boundary to match a published candidate on the four-way binding (Prefix — digest over prior canonical public content; Previous — chain of consecutive assistant boundaries; Opaque — digest of the exact gateway-issued reasoning envelope; Items), and an envelope-less boundary is accepted only where no Required candidate exists at that position (anti-downgrade). The property: **a replay is accepted if and only if it presents the observations the gateway itself published**. Item-level checking, relaxed matching, request-driven root selection, and manufactured envelopes are all ways of weakening this property, and all remain rejected (§1.3).

### 3.2 The byte-exact re-injection argument

The repair is safe because of two structural facts of the binding, both pinned in code:

1. **Prefix deliberately excludes thinking content.** `CanonicalPrefixes` skips `thinking` / `redacted_thinking` blocks when hashing (`internal/gateway/receipt/projection.go`; doc line: "It does not authenticate thinking blocks"). Re-injecting an envelope at its original boundary therefore changes neither Prefix nor Previous — it restores exactly the Opaque/Items observation, and cannot launder any public-content change.
2. **The opaque digest is canonical and content-addressed.** `Digest()` is SHA-256 over the canonical raw envelope bytes, and `opaque.Decode` rejects any alternate spelling of the same envelope (`bytes.Equal(e.raw, raw)`, `internal/gateway/opaque/codec.go`) — exactly one byte sequence per envelope can ever pass. A repaired replay passes Check if and only if the injected bytes are byte-identical to what the gateway issued at that boundary.

On top of these, the **surviving tool markers make the repair provable before injection**: `BindToolID` embeds `{call_id, opaque_sha256}` and `RestoreToolID` refuses any marker whose embedded digest differs from the presented envelope's (codec.go). Each orphaned marker discloses the exact digest required, so the repair verifies `sha256(candidate_raw) == marker.opaque_sha256` without reading the receipt store (REQ-EVR-003-4); the unchanged Check then adjudicates the full binding at request time, including the Prefix/Previous legs the repair must not touch.

**Why this is not the forbidden direction.** Accepting an envelope-less replay would require relaxing the empty fallback against Required candidates or item-level existence checks — t672 rejected both. Verbatim re-injection does neither: it reconstructs exactly the observation tuple the gateway already authorized at that boundary. The worst-case abuse is bounded by the binding itself: a cross-boundary transplant fails Opaque equality (the injected bytes' digest does not match the marker at the wrong boundary), any reordering or public-content edit fails Prefix/Previous, and a manufactured envelope fails both the marker self-attestation and Decode canonicality. Verbatim re-injection's worst outcome is exact reconstruction of an already-published boundary — the t672 invariant is **satisfied, not bypassed**: the envelope-less replay remains rejected; only the replay that carries the gateway's own bytes back is accepted, by the same predicate that always accepted it.

**Consequence (the no-bypass statement):** for every history H and gateway state S, if the repaired H is accepted, then the unchanged `Manifest.Check` accepts it on H's own observations — the repair path and the unchanged check have identical acceptance sets. The repair adds a *procedure*, not an *authorization*.

### 3.3 Refusal semantics (repair-layer vs composite) and residual risks (bounded)

The repair layer's OWN refusal triggers are exactly four: (a) **source-gone** — no verbatim carrier in the transcript for an attested digest; (b) **digest-mismatch** — candidate carrier bytes fail the marker's self-attested sha256; (c) **already-attempted** — the durable single-shot record; (d) **incomplete transcript** — the capture source is not a completed record. Each refuses with zero modification and preserves the non-destructive state (REQ-EVR-007/008).

Two further shapes have **composite** refusal semantics, and this SPEC does not claim repair-layer refusal for them:

- **Missing marker**: a boundary stripped of both envelope and marker is client-side indistinguishable from an ordinary text-only assistant turn — the indistinguishability t700's REQ-WRR-007 codifies — so the repair cannot see it and cannot abort on it. For the CauseReasoning target class the marker provably survived (the classification fires at the marker-without-envelope site, t703 verdict §1), so the shape sits outside the repair's trigger class: the repair finds nothing attested at that boundary and injects nothing there.
- **Public-content changed (Prefix-class mismatch)**: the repair path is forbidden from reading the receipt store (REQ-EVR-003-4) and holds no public-content reference, so it cannot detect a Prefix mismatch. The refusal is COMPOSITE: the repair performs (or skips) injection on marker-attested boundaries only, and the UNCHANGED Check delivers the final refusal at request time with the classified guidance. The operator-visible outcome is still a classified rejection, never an accepted bad replay. REQ-EVR-007's refusal obligation for this shape is read through these composite semantics: its zero-modification bound holds (the repair modifies only attested boundaries; the aside preserves the preimage), while the Check adjudicates the binding.

Residual risks (bounded):

- Whether the t703 switch turn also altered public content is unmeasured (t703 verdict Gaps). Under composite semantics the outcome is the Check's chain/reasoning-classified rejection either way — the conversation is never made worse by the attempt.
- Transcript retention under client compaction / `--continue` slicing is unmeasured; source-gone is a repair-layer refusal, so degradation is bounded by design.
- The transcript-retention measurement sampled carrier well-formedness (6 of 37 decoded) rather than proving all carriers; the repair's own digest self-attestation is the per-use gate, so an unverified carrier is refused at repair time, not silently injected.

## 3.4 Recoverable vs non-recoverable shapes

| Shape | Under this SPEC | Replay outcome |
|---|---|---|
| Stripped-envelope replay, markers survive, transcript source retains verbatim envelopes, public content unchanged | **Repairable** — explicit-user-invoked, single-shot, byte-exact re-injection per REQ-EVR-004/002/003 | After repair: accepted by the unchanged Check; conversation survives |
| Stripped-envelope replay, transcript source gone (compaction / `--continue` slicing / record loss) | **Not repairable** — refusal per REQ-EVR-007 | Still rejected, reasoning-classified (frozen body) |
| Stripped-envelope replay with public content also changed (Prefix mismatch) | **Composite** — repair cannot detect a Prefix mismatch (no receipt reads, REQ-EVR-003-4); it injects marker-attested boundaries only, never edits public content, and the unchanged Check delivers the final refusal | Still rejected (chain- or reasoning-classified per the changed legs) |
| Stripped boundary without a surviving self-attesting marker, or digest mismatch | Digest mismatch: **repair-layer refusal** (no admissible digest proof). Marker-less boundary: client-side indistinguishable from a text-only turn — outside the repair's trigger class; the repair injects nothing there and the Check adjudicates | Still rejected, reasoning-classified |
| Serial-turn chain-class 400 (fresh session, no model switch, all prior receipts complete — production family `1f14d174`, model luna, 12th request) | **Not covered by this SPEC — under determination by card t707**; this SPEC's repair path (REQ-EVR-004) is CauseReasoning-scoped and fires only on the stripped-envelope shape | Per t707's verdict |
| Trailing-unpublished-turn wedge (t700's shape) | **Out of scope here** — governed by `SPEC-GATEWAY-WEDGE-REROOT-001`'s client-side re-rooting policy | Per t700's policy |
| Lineage miss | **Not repairable here** — the sanctioned path remains the launcher-driven fork (`conversation.Manager.Fork`), untouched by this card (research F6: fork cannot repair mid-history stripping) | Still rejected, lineage-classified; fork per t700 REQ-WRR-008 |
| Envelope-less replay of a Required boundary, generally | **Remains rejected — t672 invariant locked**; this card never relaxes it | Rejected, always |

## 4. Sibling-SPEC interpretive seam — REQ-WRR-007 vs REQ-WRR-008(d) (surfaced, not silently resolved)

Cross-reference: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t700/.moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001/spec.md` (unlanded draft; card t700). Its REQ-WRR-007 declares stripped reasoning envelopes non-recoverable under its client-side re-rooting policy, with rejection mandated "wherever a replay carries them"; REQ-WRR-008(d) bans reasoning-envelope "re-issue, synthesis, or transplant" for gateway-state-changing recoveries, with rationale that "the Opaque digest must always reference something the gateway itself issued"; its §3.4 marks the stripped-envelope row "Not recoverable".

**Reading A (letter-based — this repair is inadmissible without t700 reconciliation).** REQ-WRR-008(d)'s letter bans "transplant" of an envelope, and re-injecting a previously-issued envelope into a replay is a transplant of that envelope into a request that no longer carries it; REQ-WRR-007's "wherever a replay carries them" mandate binds the rejection behavior this repair circumvents by making the carried replay unmodified. Under this reading the repair needs an explicit amendment or carve-out in t700's SPEC before its repair requirements may enter run phase.

**Reading B (rationale- and scope-based — this repair is admissible).** REQ-WRR-008(d)'s rationale bans *manufactured* envelopes — digests referencing something the gateway never issued; a verbatim re-injection references exactly what the gateway issued (the §3.2 argument). Further, REQ-WRR-008 governs *gateway-state-changing* recoveries ("Where any recovery path changes gateway state"), and this repair changes no gateway state (REQ-EVR-009); REQ-WRR-007 declares the shape non-recoverable *under its own client-side re-rooting (removal) policy* — a removal procedure cannot restore content, so the declaration is true of that policy without bearing on an injection procedure.

**This SPEC's position (explicit):** it adopts Reading B for its own design — and, regardless of reading, binds itself to the common core both readings forbid (no manufactured, re-encoded, or synthesized envelope bytes; only the gateway's own issued bytes, verbatim, per REQ-EVR-002). It does NOT amend t700's SPEC and does not treat the seam as resolved by silence. The seam, both readings, and this position are recorded here for the t700 owner and the orchestrator; the run-phase entry gate for the repair milestones (plan.md §F M3+) includes explicit confirmation that the t700 interplay has been adjudicated (or that t700 remains unlanded such that no conflict materializes in the absorbed tree). If adjudicated under Reading A, M3+ is blocked and the card reduces to its characterization and documentation milestones pending re-delegation.

### 4.1 External prior art — BerriAI/litellm #40288 (recorded; applicability unestablished)

Named by the lead as prior art: the litellm bridge preserves OpenAI reasoning `encrypted_content` as a signature and must restore it on inbound translation to keep byte stability across turns. It is recorded here as external prior art for byte-stability design — the same two stability surfaces this SPEC's binding depends on (envelope-digest stability and public-content byte stability under client re-serialization). No applicability claim is made: whether the observed serial-turn chain-class failure (§3.4 row, family `1f14d174`) is caused by client-side split-record re-serialization — storing `redacted_thinking` and `tool_use` as separate records and re-encoding them — changing the envelope digest or public-content bytes is precisely the question card t707 is reproducing, and this SPEC asserts nothing about it until t707's verdict lands (dispositioned in plan.md §H Resolution Record, row 3).

## Out of Scope

### Out of Scope — validator weakening

- No change to `receiptHistory.Check`, `observations`, `checkObserved`, `replayCause`, `Publish`, or `Manifest.Check`; no item-level validation, relaxed Prefix/Previous matching, or empty-fallback relaxation against Required candidates (t672 Candidate A).
- No repair-side reimplementation or pre-adjudication of the binding: the unchanged Check is the only accept/reject authority at request time.

### Out of Scope — request-driven gateway state

- No request field, header, or metadata key that triggers, shapes, or authorizes a repair; no gateway-side repair hook in the request path.
- No receipt-store read or write from the repair path; no new receipt-root, candidate, or lineage creation.

### Out of Scope — envelope manufacture

- No envelope re-issue, synthesis, re-encoding, or re-serialization; no new capture store minting envelope bytes. Only the gateway's own previously-issued bytes, read verbatim from the transcript source, are ever injected (REQ-EVR-002).

### Out of Scope — automatic recovery

- No automatic injection or history modification on any 400; repair is explicit-user-invoked only (REQ-EVR-005), single-shot with durable termination (REQ-EVR-006).

### Out of Scope — the external client fix and the fork path

- The genuine client-side continuation fix (Claude Code not stripping gateway-issued reasoning on re-encode) lives outside this repository and is not built here.
- `conversation.Manager.Fork` behavior is unchanged; this card documents it as the sanctioned lineage-miss path and changes nothing about it. No new fork instance.

### Out of Scope — the serial-turn chain-class shape

- A second confirmed production failure class — a `CauseChain`-classified 400 on a fresh session and current build, serial turns with no model switch (family `1f14d174`, model luna, 12th request, all 11 prior receipts complete) — is NOT covered by this SPEC: the repair path (REQ-EVR-004) is CauseReasoning-scoped and fires only on the stripped-envelope shape.
- Card **t707** (lane-5) is actively reproducing the shape and owns the determination; this card neither presupposes nor forecloses its result.
- The hypothesis context (client-side split-record re-serialization breaking envelope-digest or public-content byte stability) and the external prior art (BerriAI/litellm #40288) are recorded in §4.1 as input to t707 — not as an assertion of this card.

### Out of Scope — unrelated rejection axes

- The masked named-teammate 400 axis (t695/t672 instrumented) and the chain-cause wedge class (t700's policy) are untouched; the wire error contract stays byte-identical (REQ-EVR-001, §1.3-3).
