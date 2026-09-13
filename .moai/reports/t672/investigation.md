# t672 Investigation — receipt/history rejections on spawn and fork replays

Card: t672 · Branch: `WT-receipt-spawn-400` · Worktree: `.claude/worktrees/t672`
Date: 2026-09-13 · Base: develop `61a9bb57e` (includes t695's error-transparency fix)

## 0. Isolation statement

- No live gateway, session, or process was killed, restarted, or reconfigured.
- D4 probe children ran from this worktree's own build on ephemeral loopback
  ports with per-run session tokens; each child was stopped via
  `ChildProcess.Stop` (bounded 3 s) on every path. Post-run port check on the
  three probe ports: no listener remained.
- Isolated receipt stores were created under `/private/tmp/t672/receipts-*`
  (per-run UUID); nothing was written under
  `~/.moai/state/gateway-conversations/`, and no other family's directory was
  read or modified.
- GPT credentials were reused read-only through the gateway process from the
  real store (scrubbed env exactly as the production launcher does). No token
  material was printed, copied, or committed.
- Upstream cost discipline: every cell used a one-line prompt and
  `max_tokens <= 64`; 9 upstream calls per binary run (15 total across
  before/after).
- The probe driver is archived verbatim at `probe_driver_main.go.txt` (this
  directory); the untracked scratch copy was deleted from the worktree.

## 1. D1 — mechanism proof (code-pinned, file:line)

**One receipt root per gateway child, keyed to the launcher conversation.**
`internal/cli/gateway_factory.go:178` opens exactly one `receipt.Store` from
the private payload's `conversation.receipt_dir` under the launcher-authorized
`conversation.session_id`, and `gateway_factory.go:200-204` wires that single
store into BOTH `limits.History` (the replay validator) and
`limits.NativeReceiptAuthorize` (the metadata-session equality check). Every
request served by that child — parent turns, spawned subagents, teammates,
in-process summary forks — is validated against that ONE root, regardless of
which client session sent it. A client cannot select, seed, or switch a
receipt root through request content.

**Metadata session equality gates BEFORE history.**
`gateway_factory.go:69`: a request whose `metadata.user_id` JSON declares
`session_id != conversation.session_id` is rejected at
`internal/gateway/translate/request.go:105-109` with
"native receipt authorization required". Therefore any request observed to
fail with `HistoryReplayError` (the family-body class) PASSED this gate — its
declared session matched the root or it carried no native metadata.

**The history check itself.** `request.go:272-276` →
`receipt_history.go Check` → `checkObserved` → `receipt/core.go:99-126
Manifest.Check`: EVERY assistant message boundary in the replayed history must
match a published candidate on (Prefix, Previous, Opaque, Items), in the
target or the GPT-5.6-compatible domain (`receipt_history.go checkObserved`).
Prefix digests cover all canonical public content from the FIRST message of
the replay; Previous chains consecutive assistant boundaries; the opaque
digest pins the exact gateway-issued reasoning envelope at that boundary.

**Consequences (what can and cannot produce the observed errors):**

1. A fresh spawn history (system + task, no assistant messages) produces zero
   observations; `Manifest.Check` on an empty history is a branch, not a
   rejection. A spawn's FIRST turn can therefore never fail
   `HistoryReplayError` — confirmed live (matrix C2/C3: 200 with and without
   a model argument, pre- and post-fix).
2. A tail-truncated fork replay of published, pairing-valid boundaries keeps
   every retained boundary's (Prefix, Previous, Opaque, Items) identical —
   truncation-from-the-start is already tolerated. Confirmed live (C4b: 200
   pre- and post-fix).
3. A replay that INSERTS at the head, drops a middle boundary, strips an
   opaque envelope, or carries an unpublished assistant turn changes some
   boundary's digest and is rejected — by design; the prefix chain is the
   arrangement authorization ("syntax alone never authorizes history").
4. Hence the production HistoryReplayError spawn/fork incidents imply replayed
   histories that were NOT published-chain prefixes (client-side slice, strip,
   or head edit) or a lineage/domain mismatch — classes that were
   indistinguishable from the outside before this card. t695 could already
   exclude the model-id axis (catalog 404s first, its matrix §F) and the
   effort axis (its §H spawn cells green post-fix).

## 2. Design decision — why A and B were rejected, what was implemented

**Candidate A (item-level validation instead of chain equality) — REJECTED as
stated, security grounds.** The chain binding is what forbids re-splicing
gateway-issued reasoning into a different arrangement: Prefix hashes the whole
prior public content, so moving or recontextualizing a boundary changes it.
Item-level existence checks (or relaxing Prefix/Previous to tolerate
head/middle-cutting forks) would still reject foreign-session items, but would
accept intra-session transplant and reordering — a real weakening of the
layer's authorization property, not a tolerance of a legitimate shape. Stated
as a design tension, not resolved by loosening (per the card constraint).

**Candidate B (seed a new root from lineage when session has no receipts) —
REJECTED, violates a stated invariant.** Seeding would key a receipt root from
request-declared session data — exactly what the code forbids:
"NewReceiptHistory accepts only a launcher-authorized UUID and private store.
Request metadata never selects a receipt root or authorizes a conversation"
(`receipt_history.go:28-29` pre-existing comment; `conversation.Manager.Fork`
at `internal/gateway/conversation/family.go:257-300` is the sanctioned
launcher-driven path that seeds child roots). The lineage-miss case is instead
CLASSIFIED (CauseLineage) so it self-identifies.

**Candidate C (desync: rejection must not be an unexplainable loop) —
IMPLEMENTED, minimal form.** `HistoryReplayError` now carries a cause class
(`ReplayCause`: chain / lineage / reasoning) rendered as a fixed reason clause
after the historical guidance sentence. The message remains fixed recovery
guidance — no counts, digests, indexes, or session identifiers — so the
existing "exposes only fixed recovery guidance, never history data" contract
(`receipt_history.go`) is preserved, and past incident signatures keep
matching because the original sentence is kept verbatim as the prefix.

**Why classification is the instrumentation channel.** The production gateway
keeps no request log and client debug logs record bodies but not requests
(t695 investigation §6). The 400 body is the only surface that reliably
reaches operators, so the rejection class rides in the body. With this
deployed, the next production spawn/fork wedge self-identifies as exactly one
of:

- `reason: replayed history omits gateway-issued reasoning recorded at this
  position` — the stripped-reasoning fork signature (a fork that drops
  `redacted_thinking` carriers from a replayed boundary whose required receipt
  proves reasoning was issued there);
- `reason: no recorded history exists for this session` — the lineage-miss
  signature (a fork/summary session replaying parent history against a root
  with no seeded lineage; the sanctioned fix is launcher-driven `Fork`);
- `reason: replayed history does not match the recorded receipt chain` — the
  edited/sliced/unpublished-turn signature (includes the post-502 wedge shape).

**Desync (0efb66f7) disposition.** A safe gateway-side re-root does NOT exist:
the client transcript is the authority for history content, so any gateway
initiated re-root would have to trust a client-declared truncation — the same
loosening rejected above. Recovery remains: exact retry after a failed turn is
safe (nothing was published and nothing was appended — proven by the C5
precursor cells and the unit test), and a genuinely wedged conversation
(client recorded an unpublished turn) recovers only through a new conversation
or a launcher-driven fork that seeds a fresh validated root. The wedge class
now announces itself by name instead of looping unexplainably.

## 3. The masked-teammate axis (t688 advisor) — NOT fixed, instrumented

This card makes NO change to that axis. Disposition explicitly: the t695 D1
transparency (already merged in the base) unmasks every translate-path
rejection, and this card classifies the history-replay subclass. The next
named-teammate 400 will therefore carry one of: "native receipt authorization
required" (the `metadata.user_id` session ≠ launcher session hypothesis of
t695 investigation §7.3 — unchanged code, now visible), a classified
history-replay reason, or another concrete validation string. Any of the three
pins the mechanism without further speculation; guessing a fix before that
evidence was prohibited by the card and was not done.

## 4. Explicit gaps (not observed)

- The production spawn/fork incidents themselves were not reproduced from this
  lane (t695 §H cells were already green post-effort-fix); the slice/strip
  trigger class is inferred from code + the rejection-shape analysis in §1,
  and the classification makes the next occurrence conclusive.
- The upstream response behind any single 200 cell is not inspectable beyond
  the converted body; standard for this adapter.
- `internal/cli` was not touched, so the metadata-session mismatch path still
  returns its pre-existing message; distinguishing it further would require a
  launcher-scope change out of this card's envelope.
