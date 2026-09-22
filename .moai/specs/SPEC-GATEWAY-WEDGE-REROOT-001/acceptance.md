---
id: SPEC-GATEWAY-WEDGE-REROOT-001
title: "Wedge-safe re-rooting policy — acceptance criteria"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t700)
tier: M
---

# Acceptance — SPEC-GATEWAY-WEDGE-REROOT-001

## §A Scope

Acceptance proves two things and refuses to conflate them: (1) the receipt-chain authorization property is **unchanged** (the validator's acceptance set is byte-for-byte today's), and (2) the recovery path is **bounded, user-invoked, non-destructive, and gateway-write-free**, and it recovers exactly the trailing-unpublished-turn wedge shape.

Severity legend: **BLOCKER** = must pass before integration; **MAJOR** = must pass before sync; **MINOR** = tracked, non-gating.

## §B AC Matrix

| AC | Requirement | Statement (binary) | Severity | Verification |
|----|-------------|--------------------|----------|--------------|
| AC-WRR-001 | REQ-WRR-001/007 | Wedge-shaped replay (assistant boundary never published) is rejected, chain-classified, on the merged tree | BLOCKER | `go test -run TestReceiptHistoryClassifiesDesyncWedgeAfterUnpublishedTurn ./internal/gateway/translate/ -count=1` → `ok` |
| AC-WRR-002 | REQ-WRR-001 | Tail-re-rooted replay (trailing unpublished boundary removed) is accepted by the unchanged check | BLOCKER | new characterization test (see §C) green |
| AC-WRR-003 | REQ-WRR-007 | Mid-history boundary drop stays rejected, chain-classified, even after tail removal | BLOCKER | new characterization test green |
| AC-WRR-004 | REQ-WRR-007 | Foreign/forged item replay stays rejected with today's class | BLOCKER | `TestReceiptHistoryRejectsForeignItemReplay` still green |
| AC-WRR-005 | REQ-WRR-007 | Stripped-reasoning replay stays rejected, reasoning-classified | BLOCKER | `TestReceiptHistoryKeepsStrippedReasoningRejection` still green |
| AC-WRR-006 | REQ-WRR-007 | Lineage miss (nonempty replay vs empty root) stays rejected, lineage-classified | BLOCKER | `TestReceiptHistoryRejectsForeignItemReplay` still green — its `CauseLineage` assertion section (receipt_history_cause_test.go:187–190) is the lineage-classification check; no standalone lineage test exists |
| AC-WRR-007 | REQ-WRR-002 | `HistoryReplayError.Error()` carries the historical sentence as a byte-identical prefix and the three t672 reason clauses byte-identical | BLOCKER | golden string comparison in the characterization suite |
| AC-WRR-008 | REQ-WRR-003, REQ-WRR-006 | After one re-rooting attempt — recorded as a durable attempt marker (attempted flag + preimage digest) on the launcher conversation record — a further invocation, including from a fresh process reading that record, performs no additional removal and surfaces the classified guidance | BLOCKER | recovery-path unit test covering both same-process and fresh-process re-invocation (bounded termination, restart-durable) |
| AC-WRR-009 | REQ-WRR-003 | Recovery removes only the unpublished trailing assistant boundary + its dependent tool results; all user-authored turns remain; removed content is preserved aside and recoverable | BLOCKER | recovery-path unit test (content assertions) |
| AC-WRR-010 | REQ-WRR-003-4 | Recovery structurally cannot touch gateway state — the recovery implementation takes no receipt-store argument and does not open one (no receipt-store handle / `OpenStore` reference in the recovery path's files) — complemented by before/after store digest equality | BLOCKER | structural assertion on the recovery path's code + before/after store digest comparison as the complement |
| AC-WRR-011 | REQ-WRR-005 | A chain-classified 400 with no user action modifies no transcript file | BLOCKER | negative test (no invocation → no mutation) |
| AC-WRR-012 | REQ-WRR-001 (mechanical lock) | Card diff (merge-base develop..HEAD, re-derived at measurement time) contains file-level zero changes to `internal/gateway/receipt/core.go`, `internal/gateway/translate/receipt_history.go`, `internal/gateway/translate/request.go`, `internal/gateway/conversation/family.go`, `internal/cli/gateway_factory.go`, and zero paths under `internal/gateway/receipt/`; the behavioral net over the frozen call ordering is `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory` (full `Request` entrypoint) | BLOCKER | scripted diff-scope assertion (`git merge-base develop HEAD` recorded, then `git diff --name-only <recorded-base>..HEAD`), output recorded in progress.md §E.2 |
| AC-WRR-013 | REQ-WRR-004 (integration) | In an isolated gateway child (t672 discipline): wedge a turn via upstream failure, re-root the client transcript, retry → accepted; a subsequent turn publishes normally | MINOR | live probe per plan.md §D discipline; non-gating — run at M5 (Priority Low) or record the skip as a gap in progress.md |
| AC-WRR-014 | REQ-WRR-008 | The card introduces no new gateway-state-changing recovery path: `Manager.Fork` remains the sole `.Fork(` / receipt-re-root call site in production code (dual check with AC-WRR-012's scope) | BLOCKER | `grep -rn "\.Fork(" internal/ --include="*.go" \| grep -v "_test"` → exactly the pre-existing call-site set (the `Manager.Fork` definition + the launcher caller in `internal/cli/gateway_session.go`); count unchanged from the pre-card baseline recorded in progress.md §E.2 |
| AC-WRR-015 | REQ-WRR-009 | Operator document exists at `.moai/docs/gateway-wedge-recovery.md` naming all five elements: wedge class, recovery procedure, single-shot bound, non-destructive bound, non-recoverable shapes (+ the sanctioned fork path for lineage misses) | MINOR | file exists; one grep per element (keywords fixed in plan.md M6) returns ≥1 hit each |
| AC-WRR-016 | REQ-WRR-007 | Recovery on a trailing forged/stripped tail (client-side indistinguishable from the wedge): removed content is preserved aside, is never accepted by any replay, and the remaining genuine prefix is accepted by the unchanged check | BLOCKER | recovery-path + validator test: reintroducing the removed forged/stripped boundary into any replay is rejected; the re-rooted remainder passes |

All ACs are GREEN-by-test on the merged tree; AC-WRR-002/003/008..011/014..016 are new tests that follow TDD (E8 RED evidence for the recovery-path ones; AC-WRR-014's grep and AC-WRR-015's doc check are mechanical, not TDD).

## §C Given-When-Then scenarios

- **AC-WRR-001** — Given a transcript whose trailing assistant boundary was never published (the C5 fixture shape), When the history is replayed against the validator, Then the check fails with `CauseChain` and the classified body whose chain reason clause is byte-identical to t672's.
- **AC-WRR-002** — Given the same transcript after recovery removed the trailing unpublished boundary (dependent tool results removed with it), When replayed, Then `Manifest.Check` accepts: every remaining assistant boundary matches its published candidate on (Prefix, Previous, Opaque, Items).
- **AC-WRR-003** — Given a transcript missing a mid-history published boundary (tail removed or not), When replayed, Then rejected with `CauseChain` — recovery cannot and does not repair mid-history divergence.
- **AC-WRR-004 / 005 / 006** — Given a forged-item / stripped-reasoning / lineage-miss replay respectively, When replayed, Then rejected with exactly today's class and today's message; no shape migrated between classes. AC-WRR-006's lineage assertion is verified inside `TestReceiptHistoryRejectsForeignItemReplay` (its `CauseLineage` section, receipt_history_cause_test.go:187–190) — no standalone lineage test exists to name.
- **AC-WRR-007** — Given any `HistoryReplayError` value of any cause, When `Error()` is rendered, Then the string starts with the historical guidance sentence byte-identically and the cause clause is one of the three t672 clauses byte-identically.
- **AC-WRR-008** — Given a wedge incident where one re-rooting attempt completed (durable attempt marker + preimage digest written to the launcher conversation record) and the retry was still rejected, When the recovery path is invoked again — once in the same process and once from a fresh process that reads only the conversation record — Then neither invocation removes content (the incident is exhausted by the durable marker) and the classified guidance is surfaced unchanged.
- **AC-WRR-009** — Given a wedge transcript containing [published prefix, user turn N, unpublished assistant turn with tool_use, dependent tool results, plain user turn N+1], When recovery runs, Then the transcript retains both user turns and the full published prefix, loses only the unpublished boundary + dependent tool results, and the removed content is retrievable from the preserved-aside copy.
- **AC-WRR-010** — Given a prepared receipt store, When recovery executes end-to-end, Then the recovery implementation structurally holds no receipt-store handle (takes no store argument, calls no `OpenStore` — asserted on the recovery path's code), and the store's files hash identically before and after (the digest is the complement, not the sole proof: it cannot see write-then-restore, which the structural assertion covers).
- **AC-WRR-011** — Given a chain-classified rejection delivered to the client, When no user recovery action is taken, Then the transcript file's mtime and content are unchanged (no automatic surgery).
- **AC-WRR-012** — Given the card's final diff, When scoped via `git merge-base develop HEAD` recorded then `git diff --name-only <recorded-base>..HEAD` (merge-base re-derived at measurement time, per gitflow-lane-protocol §8), Then none of the five preserved files appears (file-level zero-diff: `core.go`, `receipt_history.go`, `request.go`, `family.go`, `gateway_factory.go`) and no path under `internal/gateway/receipt/` appears (recorded verbatim in progress.md §E.2); the behavioral net over the frozen authorization ordering is `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory` driving a rejection through the full `Request` entrypoint.
- **AC-WRR-013** — Given an isolated gateway child per t672 §0 (ephemeral loopback port, per-run session token, isolated `/tmp` receipt store, cleanup-guaranteed stop), When a turn fails upstream leaving an unpublished assistant turn, and the client re-roots per the policy and retries, Then the retry is accepted and the next turn both publishes a candidate and receives a normal response. (MINOR, non-gating — M5 or skip-with-gap.)
- **AC-WRR-014** — Given the card's final tree, When production call sites for receipt fork/re-root are enumerated (`grep -rn "\.Fork(" internal/ --include="*.go"`, tests excluded), Then exactly the pre-existing set is found — the `conversation.Manager.Fork` definition and the launcher caller in `internal/cli/gateway_session.go` — and no new gateway-state-changing recovery path exists (baseline count recorded in progress.md §E.2 pre-card).
- **AC-WRR-015** — Given the merged tree, When `.moai/docs/gateway-wedge-recovery.md` is checked, Then the file exists and names all five elements (wedge class; recovery procedure; single-shot bound; non-destructive bound; non-recoverable shapes) plus the sanctioned fork path for lineage misses — one grep hit per element.
- **AC-WRR-016** — Given a transcript whose trailing assistant boundary carries forged or stripped content (never published; client-side indistinguishable from the wedge shape), When the recovery procedure runs, Then the removed content is preserved aside, any replay that reintroduces it is rejected by the unchanged check, and the re-rooted remainder — a genuine published-chain prefix — is accepted.

## §D Edge cases

- Wedge with the unpublished turn as the transcript's **only** assistant boundary (first-turn wedge): re-rooting yields a fresh-history replay (spawn-shaped) — accepted with zero observations (existing behavior, `TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding` family).
- Unpublished turn containing tool_use with **no** dependent tool results yet (stream broke mid tool-call): removal takes the boundary only; transcript stays structurally valid.
- Recovery invoked on a **non-wedge** chain rejection (mid-edit): the single attempt does not fix it; bounded termination per AC-WRR-008; no data lost beyond the preserved-aside copy (restorable).
- **Trailing forged/stripped tail** (client-side indistinguishable from the wedge, REQ-WRR-007): recovery removes it under the same single-shot bound; the removed content never validates against any replay and the remaining genuine prefix is accepted — AC-WRR-016.
- Conversation already closed/expired at the launcher layer when recovery is invoked: the path refuses cleanly (nothing to re-root) without touching any file.
- Compatible-domain (GPT-5.6/astra) replays: the `checkObserved` alternate-domain pairing is untouched; characterization covers one compatible-domain shape per the existing suite's pattern.

## §E Quality gates

- TRUST 5: Tested (new tests green + characterization suite green), Readable (comments English per `code_comments`), Unified (`gofmt`/`golangci-lint` clean on new code), Secured (this SPEC's own subject — §B/§C are the security gate), Trackable (Conventional Commits + card id).
- Scoped verification only: `go vet` + `golangci-lint --new-from-rev=HEAD` + `go test` on affected packages (`internal/gateway/translate/...`, `internal/cli/...` as touched); full suite is CI's.
- LSP gates per phase (plan baseline captured; run zero errors; sync clean).

## §F Definition of Done

1. Every BLOCKER AC PASS with attributed evidence in progress.md §E.2/§E.3; AC-WRR-013 (MINOR, non-gating) either PASS or its skip recorded as a gap; AC-WRR-015 (MINOR) PASS at sync.
2. AC-WRR-012's diff-scope assertion recorded verbatim (the no-weakening lock is evidence, not assertion).
3. Operator documentation merged (REQ-WRR-009).
4. Sync-phase close (implemented → completed) done **before** the local develop merge (t342 lesson), then integration per gitflow-lane-protocol with the completion report fields (card id, branch + HEAD, local merge SHA, unpushed count, evidence path).
