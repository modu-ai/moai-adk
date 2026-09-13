---
id: SPEC-GATEWAY-WEDGE-REROOT-001
title: "Wedge-safe re-rooting policy — client-side single-shot re-rooting with a locked receipt-chain authorization determination"
version: "0.1.1"
status: in-progress
created: 2026-09-13
updated: 2026-09-14
author: manager-spec (card t700)
priority: P1
phase: "v3.2.0 target"
module: "internal/gateway/translate, internal/gateway/receipt, internal/gateway/conversation, internal/cli"
lifecycle: spec-anchored
tags: "gateway, receipt-chain, wedge, re-rooting, recovery, security-determination, gpt"
tier: M
related_specs:
  - SPEC-MOAI-GATEWAY-001
---

# SPEC: Wedge-safe re-rooting policy — the security-gated recovery path card t672 named

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-13 | manager-spec | Initial plan-phase emission (card t700, Class C). Evidence base: `.moai/reports/t672/{verdict,investigation,matrix}.md` (lane-6, 2026-09-13) — mechanism proof, 9-cell real-request matrix, and the t672 disposition that names this card. Symbol-level feasibility pinned on tree `WT-wedge-reroot-policy` (base local develop `44e56d017`) — plan.md records each research command and its observed output. |
| 0.1.1 | 2026-09-13 | manager-spec | Plan-audit iteration-2 fix pass (card t700; iter-1 COND-FAIL 0.6875 < 0.80, report `.moai/reports/t700/plan-audit.md`). D1: REQ-WRR-008/009 gain AC-WRR-014/015; combined token "REQ-WRR-003/006" split to full IDs. D2: AC-WRR-012 tightened to file-level zero-diff incl. `request.go`; `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory` cited as the wire-level behavioral net. D3: REQ-WRR-007 consequence reworded to the §3.2 prefix-property form with client-side indistinguishability of trailing forged/stripped stated; §3.4 row added; AC-WRR-016 added. D4: durable attempt marker + fresh-process single-shot bound; aside-before-replace explicit. D5: absorb target pinned to local `develop` + t697 ancestry gate; broken `go doc` pre-flight replaced with declaration greps; AC-WRR-013 downgraded MINOR. D6: lineage test named. D7: structural no-store-handle assertion on AC-WRR-010. Optional O1-O4 recorded open (orchestrator discretion). |

## 1. Problem — measured shape

### 1.1 The wedge mechanism (proven, card t672)

A transient upstream failure (HTTP 502, the masked effort-400) during a GPT-gateway turn can leave the **client transcript holding an assistant turn the gateway never published** (for example, a stream that delivered partial content before aborting — the gateway publishes the receipt candidate only on a completed turn, `History.Publish` via `translate/response.go`; a failed turn publishes nothing). Every later replay of that conversation history fails receipt-chain validation — `HistoryReplayError` with `CauseChain` — and is rejected with 400. The rejection itself is correct: it is the layer's authorization function (blocking foreign, reordered, stripped, and never-published history). But until t672 the failure reason was indistinguishable from every other chain mismatch, and the only recovery guidance was "start a new conversation" — so **long-lived conversations died permanently** (the lane-8 incident). t672's matrix cell C5 locks this shape as rejected-chain-classified; its precursor cells prove the contrasting fact that an **exact retry after a failed turn is safe** (nothing was published and nothing was appended).

Two facts bound the problem precisely:

1. **The wedge shape is a suffix divergence.** The client transcript differs from the published receipt chain by a *trailing* never-published assistant boundary (possibly with dependent tool results after it). Everything before it is a true published-chain prefix.
2. **Tail-truncated replays of the published chain are already accepted.** A client replay that drops the trailing unpublished turn passes the **unchanged** validator — proven live (t672 matrix C4b: 200 before and after the fix) and locked by `TestReceiptHistoryAcceptsTruncatedForkReplay`. This card does not need to loosen anything to make that shape recoverable; it needs a **policy** that makes the recovery deliberate, bounded, and provably non-bypassing.

### 1.2 What t672 shipped and what it left open

t672 shipped **cause classification only** (`ReplayCause`: chain / lineage / reasoning, rendered as a fixed reason clause after the historical guidance sentence) — no recovery mechanism. Its disposition states the residual risk verbatim: the wedge UX still forces "start a new conversation" on users, long-lived conversations keep dying, and "안전한 재뿌리 정책은 보안 판단이 필요한 별도 카드 후보" (a safe re-rooting policy is a separate card candidate requiring a security judgment). This card is that card.

### 1.3 Constraints inherited from t672 (policy-space bounds — violating any is an automatic design reject)

1. **Item-level validation instead of chain equality — REJECTED** in t672 on security grounds (it would accept intra-session transplant and reordering). Not reintroduced here.
2. **Request-declared seeding or selection of receipt roots from client-supplied metadata — REJECTED** ("request metadata never selects a receipt root"). Not reintroduced here.
3. **The fixed-guidance-only error contract is frozen**: the historical sentence (`historyReplayGuidance`) stays byte-identical as the prefix of every classified message, and t672's three reason clauses stay byte-identical — past incident signatures keep matching.
4. **Tail-truncated replays are already accepted** (C4b / `TestReceiptHistoryAcceptsTruncatedForkReplay`). The security determination (§3.2) must state precisely what this does and does not imply: it makes the *client-side* recovery shape already-valid against the unchanged checker; it implies nothing about any gateway-state-changing path, which remains governed by §3.3.
5. **No background load.** Any live gateway probe in run phase follows t672's isolation discipline: ephemeral loopback ports, per-run session tokens, isolated receipt stores under `/tmp` (per-run UUID), cleanup-guaranteed child shutdown (`ChildProcess.Stop`, bounded), bounded upstream cost (`max_tokens <= 64`).

### 1.4 Sibling-card sequencing (t697 → t700)

Card **t697** is concurrently modifying the same gateway area and **lands before this card's run phase**. This SPEC therefore pins **symbol names, never line numbers** (§2 anchors), and its run phase MUST first absorb `develop` (which by then carries t697) per the git-flow lane protocol and **re-verify every symbol pin in §2 and plan.md §A against the absorbed tree before the first code commit**. Merge order: **t697 → t700**. If t697 renamed or moved any pinned symbol, the SPEC artifacts are corrected first (via the orchestrator's D-NEW-1 mid-run re-delegation path) before implementation proceeds.

## 2. Requirements (GEARS)

### 2.1 User story

As a user of a long-lived GPT-gateway conversation, when a transient upstream failure wedges my conversation (every retry rejected because my transcript carries an assistant turn the gateway never published), I want a safe, bounded, explicitly-invoked way to re-root my transcript at the last published boundary so the conversation survives — without the gateway accepting one byte of history it would not have accepted before, and without any path that lets request content select, seed, or rewrite gateway receipt state.

### 2.2 Requirements

- **REQ-WRR-001** (Ubiquitous — authorization-invariant lock): **The receipt-chain validator's accept/reject behavior shall remain unchanged by this SPEC** — `receiptHistory.Check` → `Manifest.Check` accepts exactly the replayed histories it accepts today (published-chain prefixes, including tail-truncated fork replays) and rejects exactly the histories it rejects today, with `ReplayCause` classification unchanged. This card adds a recovery *policy around* the validator; it does not touch the validator.
  - Anchors: `receiptHistory.Check`, `checkObserved`, `replayCause` (`internal/gateway/translate/receipt_history.go`); `Manifest.Check` (`internal/gateway/receipt/core.go`).

- **REQ-WRR-002** (Ubiquitous — error contract frozen): **The fixed recovery guidance sentence (`historyReplayGuidance`) and t672's three classified reason clauses shall remain byte-identical** on every rejection body. Recovery guidance is documented for operators (REQ-WRR-009), never added to the wire body.
  - Anchor: `HistoryReplayError.Error()` (`internal/gateway/translate/receipt_history.go`).

- **REQ-WRR-003** (Capability gate — recovery admissibility): **Where** a wedge recovery procedure re-roots a conversation transcript, **the procedure** shall satisfy ALL of:
  1. *Trailing-boundary-bounded removal* — it removes only the client-recorded assistant boundary that the gateway never published, plus the tool-result turns dependent on that boundary; it never removes, edits, or reorders any other message (plain user turns after the unpublished boundary are preserved);
  2. *Single-shot* — at most one re-rooting attempt per wedge incident; no trim-until-accepted iteration. The attempt is recorded as a **durable attempt marker on the launcher conversation record** (attempted flag plus the preimage digest of the removed boundary), so the bound holds across process restarts — a fresh process reading the record performs no further removal;
  3. *Non-destructive* — removed content is preserved aside **before the transcript is replaced** (aside-before-replace) and remains recoverable, never hard-deleted;
  4. *No gateway write* — the procedure never creates, seeds, selects, or mutates any receipt store, receipt root, or lineage.

- **REQ-WRR-004** (Event-driven — user-invoked recovery): **When** a conversation is wedged (a replay rejected with the chain cause while the transcript's trailing assistant boundary was never published), **the client-side recovery path** (launcher-side, operating on the conversation record's transcript — see `internal/cli/gateway_session.go`'s launcher conversation flow) shall, **on explicit user invocation**, re-root the transcript per REQ-WRR-003 and allow an exact retry of the conversation.

- **REQ-WRR-005** (Unwanted — no automatic surgery): **The recovery path shall not run automatically on any 400 rejection.** A chain-cause classification alone shall never modify a transcript without explicit user action (the chain class does not uniquely identify the wedge shape; a human confirms).

- **REQ-WRR-006** (Unwanted — bounded termination): **When** the retry after the single re-rooting attempt is still rejected, **the recovery path** shall terminate recovery, preserve its non-destructive state, and surface the gateway's classified guidance unchanged — it shall not retry again, trim further, or reclassify. The bound is read from the durable attempt marker (REQ-WRR-003-2), so termination survives process restarts.

- **REQ-WRR-007** (Ubiquitous — recovery-class boundary): **Only the trailing-unpublished-turn wedge shape is recoverable under this policy.** Mid-history boundary drops, head or middle edits, foreign or forged items, stripped reasoning envelopes, and lineage misses shall continue to be rejected with their existing classes (chain / reasoning / lineage) wherever a replay carries them. The governing property is the §3.2 prefix property — **no removed content ever validates against any replay, and acceptance after recovery implies the remaining history is a published-chain prefix** — not a claim about converting shapes. **The procedure cannot distinguish, client-side, a trailing never-published boundary from a trailing forged or stripped one** (the chain class alone does not identify the shape, and receipt-store reads are forbidden by REQ-WRR-003-4); REQ-WRR-005's explicit user confirmation and REQ-WRR-003-3's non-destructive preservation are the safeguards for that indistinguishability, and a trailing forged or stripped boundary removed by the procedure remains content that no replay will ever accept.

- **REQ-WRR-008** (Capability gate — gateway-state-changing recovery): **Where** any recovery path changes gateway state (re-root or lineage seeding), **it** shall be admissible only when ALL of the following hold: (a) the truncation point is chosen by the gateway from its own receipt store — never declared by request metadata; (b) invocation is launcher-driven, not request-driven; (c) every replayed boundary is validated by the unchanged chain check — prefix-equality, no relaxed matching; (d) no reasoning envelope is re-issued, synthesized, or transplanted; (e) no-weakening — the candidate set visible to the check after recovery is a subset of (or identical to) the candidate set before, so no replay gains acceptance power the check did not already grant. The existing launcher-driven fork (`conversation.Manager.Fork`, invoked via the launcher's fork path in `internal/cli/gateway_session.go`) is the **only sanctioned instance** of this class today; until a new path demonstrates all five properties, it shall not exist.
  - Anchors: `conversation.Manager.Fork` (`internal/gateway/conversation/family.go`), `Manifest.Fork` (`internal/gateway/receipt/core.go`), `authorizeGatewayNativeReceipt` / `newGatewayHandlerFactory` (`internal/cli/gateway_factory.go`).

- **REQ-WRR-009** (Ubiquitous — operator documentation): **The wedge class, the recovery procedure, its single-shot and non-destructive bounds, and the non-recoverable shapes shall be documented for operators** (the error body keeps pointing at fixed guidance; operator docs carry the recovery path and its limits).

## 3. Security determination — why the chosen recovery path cannot bypass receipt-chain authorization

### 3.1 The property being protected

The receipt chain is the gateway's history-authorization primitive: `Manifest.Check` requires every assistant boundary in a replayed history to match a published candidate on the four-way binding (Prefix — digest over all prior canonical public content; Previous — chain of consecutive assistant boundaries; Opaque — digest of the exact gateway-issued reasoning envelope; Items), and an empty-envelope boundary is accepted only where no required receipt exists at that position (anti-downgrade). The property: **a replay is accepted if and only if it is a prefix of what the gateway itself published** (tail truncation included). Item-level checking, relaxed prefix matching, request-driven root selection, and client-declared truncation are all ways of weakening this property, and all are rejected (§1.3).

### 3.2 Path (a) — chosen: client-side single-shot re-rooting, gateway byte-unchanged

The chosen recovery is a client-side transcript operation with **zero gateway changes**. Its safety argument has four legs:

1. **Acceptance predicate invariance.** The validator, the receipt store, and the publish flow are untouched (REQ-WRR-001), so a replay accepted after recovery is accepted by exactly the same predicate that rejected it before. The accepted shape is the already-locked tail-truncation tolerance (C4b, `TestReceiptHistoryAcceptsTruncatedForkReplay`) — the recovery creates no new acceptance, it makes a deliberate, bounded, user-confirmed path to a shape the checker already grants.
2. **Truncation is client-chosen from client-owned content.** The removal is derived from the client's own transcript (its own trailing boundary), not from request-declared metadata selecting gateway state; the gateway store is never read for the decision and never written (REQ-WRR-003-4). The "request metadata never selects a receipt root" invariant is untouched by construction.
3. **Deletion has no acceptance power.** The removal only ever deletes client content. Any deletion before a published boundary breaks that boundary's Prefix digest and stays rejected; deletion after the last published boundary is invisible to the checker (it produces no observations). There is no arrangement of removed content that makes a foreign, reordered, stripped, or never-published boundary validate — Prefix covers all prior public content and Opaque pins the gateway-issued envelope.
4. **Bounded and confirmed.** Single-shot (REQ-WRR-003-2) and user-invoked (REQ-WRR-004, REQ-WRR-005): the procedure cannot iterate into a trim-until-accepted loop, cannot fire silently on an unrelated 400, and its one attempt either recovers the wedge shape or terminates with the classified guidance unchanged (REQ-WRR-006).

**Consequence (the no-bypass statement):** for every transcript T and every gateway state S, if T-after-recovery is accepted, then T-after-recovery is a published-chain prefix of S — i.e., the unchanged `Manifest.Check` would accept it. The recovery path and the unchanged check have identical acceptance sets; the policy adds a *procedure*, not an *authorization*.

### 3.3 Path (b) — gateway-state-changing recovery: admissibility conditions

Any recovery that changes gateway state (re-rooting the receipt root, seeding lineage) would create a *new* root against which the chain check runs, and is admissible only under the five conditions of REQ-WRR-008. The reasoning: (a)+(b) keep root selection in launcher-gateway hands (the request channel stays non-authoritative); (c) keeps the acceptance predicate unchanged on the new root; (d) prevents manufactured reasoning envelopes (the Opaque digest must always reference something the gateway itself issued); (e) is the no-weakening core — the candidate set after recovery must grant no acceptance the pre-recovery set did not. `conversation.Manager.Fork` satisfies all five today: the truncation set is the parent store's own snapshot (gateway-chosen), invocation is launcher-driven, the child root's candidates are a verbatim copy of the parent's (no-weakening is structural), the unchanged check runs on the child as on any root, and no envelope is synthesized. **This card introduces no new instance of path (b)**; REQ-WRR-008 stands as the gate for any future one.

### 3.4 Recoverable vs non-recoverable shapes

| Shape | Under this policy | Replay outcome |
|---|---|---|
| Trailing-unpublished-turn wedge (assistant boundary never published, possibly + dependent tool results) | **Recoverable** — user-invoked, single-shot, non-destructive re-root per REQ-WRR-003/004 | After recovery: accepted by the unchanged check (published-chain prefix); conversation survives |
| Exact retry after a failed turn (client recorded nothing) | Already safe today (t672 C5 precursors) — no recovery needed | Accepted |
| Mid-history boundary drop, head/middle edit | **Not recoverable** — removal cannot restore a broken prefix chain | Still rejected, chain-classified (t672 reason clause, byte-identical) |
| Foreign / forged item; stripped reasoning envelope | **Not recoverable** — the four-way binding rejects them regardless of trailing content | Still rejected, chain- or reasoning-classified |
| Lineage miss (replay against a root with no recorded history) | **Not recoverable by this policy** — the sanctioned fix remains the launcher-driven fork | Still rejected, lineage-classified; `Manager.Fork` per REQ-WRR-008 |
| Trailing forged or stripped boundary (client-side indistinguishable from the wedge) | **Removable by the same single-shot procedure, never recoverable as content** — safeguards: user confirmation (REQ-WRR-005) + non-destructive preservation (REQ-WRR-003-3) | After recovery: the remaining genuine published prefix is accepted by the unchanged check; the removed content is accepted by no replay |

## Out of Scope

### Out of Scope — validator weakening

- No item-level validation, relaxed Prefix/Previous matching, or tolerance for head/middle-cut forks (t672 candidate A, rejected on security grounds).
- No change to `Manifest.Check`, `checkObserved`, `replayCause`, `Publish`, or the observation/domain binding.

### Out of Scope — request-driven gateway state

- No request field, header, or metadata key that selects, seeds, re-roots, or switches a receipt root (t672 candidate B, rejected).
- No gateway-initiated re-root that trusts a client-declared truncation point.

### Out of Scope — wire-body changes

- No new reason clause, recovery wording, or guidance text in the 400 body; the historical sentence and the three t672 reason clauses stay byte-identical (REQ-WRR-002).

### Out of Scope — automatic recovery

- No automatic transcript modification on any rejection; recovery is explicit-user-invoked only (REQ-WRR-005).

### Out of Scope — the masked-teammate 400 axis

- The masked named-teammate 400 (t695/t672 instrumented, root cause unspecific) is untouched; the next occurrence's body answers it.

### Out of Scope — fork-path redesign

- `conversation.Manager.Fork` behavior is unchanged; this card documents it as the sanctioned path-(b) instance and changes nothing about it.
