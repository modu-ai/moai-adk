# t1921 — lease-path queue spec_id propagation repair

Card: t1921 (Class B, serial, run→sync, no SPEC) · Run: tnbn7e · Lane: lane-7
Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1921 (WT-spec-less-lane @ c4784db4c2, clean)
Dispatch: leader (2026-10-26 08:5x KST, operator approval: card-tree reuse lifted; ② landing-record only; F3 adjacent records kept — repair forbidden)

## 1. Card intake

- 2026-10-26 ~09:2x KST — Re-dispatch received from leader. Tree verified: toplevel `.moai/worktrees/t1921`, branch WT-spec-less-lane, HEAD c4784db4c2, clean.
- 8859 lease recovery, step 1 observed: `factory_stage --run tnbn7e` → refused `card lease expired: card t1921 lease held by "lane-7" expired at 2026-10-25T22:44:18Z` → card reset to `assigned` (factory status: t1921 state=assigned v5 owner=lane-7 lease=none). Confirmed.
- 8859 step 2 attempted: `factory next --card t1921 --run tnbn7e` (parent checkout) → **refused serial-slot**: another serial card is in flight.
- Slot holder read (factory status, read-only): t1961 state=leased owner=lane-9, lease live until 2026-10-26T00:34:21Z — the only serial card with a live lease. Mechanical gate; no lane action can lift it.

## 2. Wait record (slot)

- **Reason**: serial slot held by t1961 (lane-9, live lease until 00:34:21Z ≈ 09:34 KST).
- **Whom**: no one — mechanical serial-slot gate in `factory next`; the leader holds serial-slot policy but the gate resolves itself when t1961's lease lapses or lane-9 lands.
- **Recheck point**: standing cron (every 5 min) wakes the lane; on each wake run the watchdog pass first, then re-read `moai factory status --run tnbn7e`; when t1921 is leasable, re-run `factory next --card t1921 --run tnbn7e` (parent checkout), then `factory_stage` to `run`.
- No polling cron added for this wait beyond the standing recheck (mechanics § Lease renewal — only the card being executed).

## 3. Class B cause-establishing evidence — source verified against the leader's prescription

(lesson: leader prescription is implemented only after source comparison — t1958)

Verified against `internal/cli/factory_card.go` @ c4784db4c2 (read in this tree):

- **Mechanism** — `homestate.CardFields.SpecID *string` (`internal/homestate/card_picked.go:64`); on a fresh claim `recordPickedTx` inserts the new row with `fields.apply(&c)`, so a SpecID present in fields rides record creation (`card_picked.go:162-170`). For an EXISTING row, non-empty fields update it only while `picked` (`card_picked.go:184`).
- **Queue source** — `factory.BacklogItem.SpecID *string` (`internal/factory/backlog_store.go:89`), set at pick time as-is (`internal/cli/todo.go:1673` — "the store is not a SPEC registry; normalization is out of scope"), also via `set-spec` on a picked item (`internal/cli/factory_card_metadata.go:394`).
- **Defect sites (all three drop the queue item's SpecID at record creation)**:
  1. **Generic unnominated arm (b2)** — `factory_card.go:1335`: `factoryNextRecordAndClaim(..., it.ID, lane, hubFields(it.ID))`; `hubFields` carries only the hub-chain after-hint. Queue item `it` is in scope.
  2. **Arm (c)** — `factory_card.go:1417-1423`: `fields := factoryGeneratedHubFields(hubFields(promoted), promotedRow)` → `factoryNextRecordAndClaim(..., fields)` at 1423. The promoted queue item's SpecID is dropped; `queueRec` (loaded at 1065) can answer the lookup via `queueItemIn(queueRec, promoted)`. The `fields = homestate.CardFields{}` reset at 1421 (assigned-row reuse) is CORRECT as-is — an assigned row is an existing row, not a creation.
  3. **Nominated path** — `factory_card.go:2174`: `hubHint := factoryGeneratedHubFields(factoryHubChainFields(...), nom.row)` passed at 2178 → `factoryNextNominatedClaim` → `factoryNextRecordAndClaim` at 2206 (the `nom.row == nil` creation path). `nom.item` carries SpecID; dropped.
- **Contrast (already correct, not in scope)** — `factory assign --spec` populates `fields.SpecID` (`factory_card.go:3851`); the assign path is not the lease path.
- **Regression hazard the repair must respect** — the queue stores spec as-is with no format validation, while `RecordPicked` runs `fields.validate()` which rejects a non-empty SpecID not matching `specIDPattern` (`card_picked.go:88`). Naive as-is propagation would turn a previously-working lease into a hard refusal for a malformed queue spec. Repair should propagate a well-formed spec and leave a malformed one nil (today's behavior for that input) — same predicate the row validation itself uses.
- **Creation-only discipline** — mirror `factoryGeneratedHubFields`'s t1533 rule: a creation input is written only when the card has NO row at all; an existing row keeps whatever it carries. SpecID follows the same shape (set only on the `row == nil` paths).
- **Out of scope per dispatch**: defect ② = landing record only (this file records it; no code change); F3 = adjacent records kept, repair forbidden. The mirror sites (`factory_mirror.go:65`, `gtd_operation.go:402` — `CardFields{}` with actor "dispatch") are NOT in the dispatch's three-site list and stay untouched.

## 4. Plan

1. ⏸ serial slot → lease t1921 (8859 steps 2-3), `factory_stage` → run
2. RED: regression test asserting the record row's SpecID after lease through each of the three creation paths (expected to fail on @ c4784db4c2) — observe RED verbatim before any fix
3. Fix: populate `CardFields.SpecID` from the queue item at the three creation sites (well-formed only, creation-only)
4. GREEN + build/vet/lint 0 (lane-local, env-scrubbed, compound invocation)
5. sync stage → card-review (`codex_review` scope=card → `card-review.md`)
6. Landing: `moai factory complete` → draft PR (title carries t1921) → head-bound independent audit; merge is the leader's (no force push; commit messages carry t1921)

## 5. Repair design (leader-directed waiting work — design only, NO commits)

Leader fact notice 2026-10-26: t1921 is LAST in the serial order (t1962 → t1944 → t1932 → t1921); the slot refusal is not a defect; waiting-period work = this design, recorded here, nothing committed.

### 5.1 Shared helper (new)

`internal/homestate/card_picked.go` — beside `specIDPattern` (line 15), following the `ValidCardID`/`ValidSetSpecID` export precedent (one grammar, one owner):

```go
// ValidSpecID reports whether s matches the card-record SPEC grammar — the
// same pattern CardFields.validate applies. Lease-path propagation gates on
// it so a queue's as-is spec (the store is not a SPEC registry) never turns
// a working lease into a validation refusal.
func ValidSpecID(s string) bool { return specIDPattern.MatchString(s) }
```

`internal/cli/factory_card.go` — near the other creation-fields helpers:

```go
// factoryCreationSpecField returns the record-creation SpecID for a queue
// item: the item's spec when well-formed (the row-level validation would
// accept it), nil otherwise — a malformed queue spec keeps today's behavior
// (no propagation, no new claim refusal). Creation-input only: callers gate
// on "no row exists", mirroring factoryGeneratedHubFields's t1533 rule.
func factoryCreationSpecField(item factory.BacklogItem) *string {
	if item.SpecID == nil || !homestate.ValidSpecID(*item.SpecID) {
		return nil
	}
	return item.SpecID
}
```

Rejected alternative: cli-local copy of the regexp — two grammars drift; homestate owns the grammar `CardFields.validate` enforces.

### 5.2 Site 1 — arm (b2), factory_card.go:1335 (creation guaranteed: arm fires only for `recorded[it.ID] == false`)

```go
fields := hubFields(it.ID)
fields.SpecID = factoryCreationSpecField(it)
return factoryNextRecordAndClaim(ctx, db, root, runID, it.ID, lane, fields)
```

### 5.3 Site 2 — arm (c), factory_card.go:1417-1423 (creation = `promotedRow == nil`)

After the existing `fields := factoryGeneratedHubFields(...)` / assigned-row reset block, add:

```go
if promotedRow == nil {
	if it, ok := queueItemIn(queueRec, promoted); ok {
		fields.SpecID = factoryCreationSpecField(it)
	}
}
```

- Lookup source: `queueRec` — the pass's one pure read (line 1065); the mutate promoted that same item and SpecID is immutable across the promote. `queueItemIn(record *factory.BacklogRecord, cardID string) (factory.BacklogItem, bool)` verified at line 3740.
- The assigned-row reuse reset (`fields = homestate.CardFields{}`, line 1421) stays untouched — an existing row takes no generated creation input.

### 5.4 Site 3 — nominated path, factory_card.go:2174-2178 (creation = `nom.row == nil`; verified `factoryNominee.item` is a value, `row` nil when no row — lines 1933-1937)

After the `hubHint := factoryGeneratedHubFields(...)` line:

```go
if nom.row == nil {
	hubHint.SpecID = factoryCreationSpecField(nom.item)
}
```

- Assigned nominees never reach RecordPicked (routed to `factoryNextClaim` at 2203-2205); picked existing rows keep their recorded spec — creation-only holds on both.

### 5.5 RED test draft — `internal/cli/factory_card_spec_propagation_test.go` (new file)

Fixtures mirror `factory_dispatch_ordering_test.go` / `factory_card_eligible_pull_test.go` (`fcFixture`, `fcQueue`, `sdRegisterLane`, `runFactory`). Table over five cases; each positive case asserts lease success AND the row field, each negative case guards the no-regression shape:

| case | setup | expected after lease |
|---|---|---|
| b2-creation | queue item `picked` + `spec_id`, no factory row (arm b2 shape) | row created at picked, `SpecID == "SPEC-X-001"` |
| c-creation | queue item `queued` + spec_id (arm c promotes) | row created, SpecID set |
| nominated-creation | queued + spec_id via `factoryNextNominate` | row created, SpecID set |
| malformed-no-refusal | queue item spec `"not-a-spec"` | lease SUCCEEDS, row.SpecID == "" (no new refusal mode) |
| existing-row-untouched | picked+unowned row already present (pre-fix spec-less row) | lease updates row, SpecID stays "" (creation-only) |

RED observation obligation (tdd-result-contract): run the three positive cases on @ c4784db4c2 BEFORE any fix — expect `row.SpecID == ""` failures recorded verbatim as EXPECTED_RED in this file; then fix; then GREEN + regression set.

Test skeleton (accumulated per leader request 03:04Z; fixtures per `factory_dispatch_ordering_test.go` — `fcFixture`, `fcQueue`, `sdRegisterLane`, `runFactory`; manager-develop verifies exact helper signatures at implementation):

```go
func TestFactoryNextLeasePropagatesQueueSpecID(t *testing.T) {
	cases := []struct {
		name    string
		spec    *string // queue item's spec_id; nil = absent, malformed = "not-a-spec"
		path    string  // "b2" | "c" | "nominated"
		wantRow bool    // creation expected
		wantSpec string // expected row.SpecID after the lease
	}{ /* five cases from the table above */ }
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, store := fcFixture(t)
			// seed the queue item with spec per tc.spec, in the state the path needs
			// (b2: picked + no row · c: queued · nominated: queued + factoryNextNominate)
			// run the path; then:
			row, err := db.LoadCard(ctx, "run-1", cardID)
			// assert row.SpecID == tc.wantSpec (creation-only) and lease success for the
			// malformed case (no new refusal) — RED today on the three positive cases.
		})
	}
}
```

Existing-row case rides the same harness with a pre-seeded picked+unowned row (fcPlace) asserting SpecID stays empty after the lease.

### 5.6 Verification plan (lease-held only)

Env-scrubbed compound, lane-local scope: `unset GOFLAGS && go build ./... && go vet ./internal/cli/... ./internal/homestate/... && go test ./internal/cli/ -run 'FactoryNext.*' -count=1 && golangci-lint run internal/cli/... internal/homestate/...` — bound per the file-redirect contract; the deciding lines land in this record and the verdict.

## Verification log

(appended as stages complete)

## Watchdog passes

- 2026-10-26T00:32:34Z — pass 1 (first observation, fail-open). decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status(t1961 leased until 00:34:21Z, slot held; t1921 assigned v5 lease=none) ladder_path=§3.2-wait-confirmed — explicit wait stands, recheck = next cron fire (00:35Z).
- 2026-10-26T00:36:25Z — pass 2. t1961 lease now EXPIRED but slot still refused on `factory next --card t1921` retry → lane-9 holder is ALIVE (t1828: expired lease of a live holder keeps holding). decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status + factory-next refusal verbatim ladder_path=§3.2-wait-confirmed — t1961 is ahead of t1921 in serial order; wait until it advances past the slot. Recheck = next cron fire.
- 2026-10-26T00:42:06Z — pass 3. Serial order holding as the leader stated: t1962 (lane-10) now holds the slot with a live lease until 00:49:38Z; t1961 (lane-9) still alive on an expired lease; t1944/t1932/t1921 assigned and waiting. Repair design recorded in §5 this period (evidence mtime moved — no stall). decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
- 2026-10-26T00:51:37Z — pass 5. Queue advanced one slot: t1944 (lane-8) now holds the serial slot (live lease until 01:04:39Z); t1961 reset to assigned (v5, lease reclaimed); t1962 still on an expired lane-10 lease. Two serial cards still ahead (t1944 → t1932 → t1921). decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
- 2026-10-26T01:05:58Z — pass 8. t1944 lease EXPIRED but `factory next --card t1921` refused again → lane-8 holder alive (t1828). Order unchanged: t1944 → t1932 → t1921. decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status + factory-next refusal ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
- 2026-10-26T01:10:41Z — pass 9. Corrected read: t1962 (lane-10) ADVANCED to sync-audit (v9, live lease until 01:22:59Z) — it re-leased between reads and holds the slot; refusal is explained. t1944/t1961 reset to assigned; t1932 still assigned (lane-3). t1962 near landing → slot close to freeing. decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
- 2026-10-26T01:58:22Z — pass 19. t1962 reached pr-open/merge-ready (v22, lease released) — its slot hold ended. t1944 (lane-8) took the slot: sync-audit v9, live lease until 02:12:39Z (run stage already passed). Order: t1944 (in sync) → t1932 → t1921. Harness note: factory messaging degraded (inbox context deadline exceeded) surfaced at this wake — queue-on-disk remains the channel; no retry. decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
- 2026-10-26T02:36:18Z — pass 27. t1944 reached pr-open/merge-ready (v16, slot freed). t1932 (lane-3) took the slot: leased, live until 02:49:06Z — one serial card ahead (t1932 → t1921). decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
- 2026-10-26T03:04:55Z — pass 33 + leader coordination. t1944 reached merged-pr (v17 — PR merged). t1932 (lane-3) still holds the slot (lease renewed until 03:19:55Z). LEADER REQUEST accepted (voluntary yield, serial-slot policy): lease t1921 only AFTER t1961 (lane-9, one complete step remaining, starved) reaches pr-open; until then keep accumulating design work in this record. decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status + leader notice (socket 50218) ladder_path=§3.2-wait + leader slot-policy coordination — modified wait: t1932 lands → t1961 pr-open → then t1921 lease; recheck = next cron fire.
- 2026-10-26T08:14:08Z — pass 98. t1961 RECLASSIFIED to parallelizable (sync stage v8, lane-9 active, live lease) — it left the serial order; the starvation the leader cited is resolved and the yield condition is overtaken by the reclassification. Remaining serial ahead of t1921: only t1932 (lane-3, in run, live lease until 08:27:28Z). t1887 (after=t1944, merged) is also now serial-leasable — first come when the slot frees. Lease attempt still gated by t1932; wait continues. decision record: decided_by=claude+lane-7-watchdog evidence_refs=factory-status ladder_path=§3.2-wait-confirmed — wait; recheck = next cron fire.
## 6. Run-phase delegation
- 2026-10-26T10:13:12Z — LEASE ACQUIRED: `factory next --card t1921` succeeded (lane-3 holder gone; t1932 flight ended); 8859 steps 2-3 complete. `factory_stage` → run (v7, lease renewed). Card body read: defect ② detail matches the dispatch scope (factory_card_pr.go:253-255, landing-record only).
- 2026-10-26T10:1xZ — DELEGATED: manager-develop (sonnet, effort medium, Class B), background spawn, sole writer while live. Delegation: §5 design + verified source facts; RED-first obligation (verbatim output required before fix); 5-case regression test; one commit `fix(factory): propagate the queue item's spec_id into lease-path record creation (card t1921)` + 🗿 MoAI; no push; scope excludes defect ② and mirror sites. Evidence: the agent's structured report → appended to §7 on return.
- 2026-10-26T10:33Z — delegate interim notice: RED→fix→GREEN portion done (moved on); regression set (FactoryNext selector) running in its own background, >10min under load; delegate will resume + re-notify on completion. Awaited-delegate wait continues (§3.5: delegate not ended, evidence moving).

## 7. Run-phase verification (lane-observed)

- Delegate report (manager-develop, completed): RED observed verbatim on @c4784db4c2 (3 positive cases FAIL with `row SpecID = ""`); GREEN post-fix; regression `go build ./...` OK · `go vet` OK · `go test ./internal/cli/ -run 'FactoryNext' -count=1 -timeout 45m` ok 659.761s exit 0 (first attempt hit the default 10m timeout under load — environmental per its stack trace + isolation pass) · `go test ./internal/homestate/ -count=1` ok 180.847s exit 0 · gofmt clean · golangci-lint `0 issues.` exit 0. Deviations: helper relocation (comment-only); malformed-no-refusal case rides arm (b2) with self_dispatch ON — within §5 latitude. Gaps: no full suite (lane-local discipline); RED run's EXIT line captured tee's status (FAIL verdict taken from verbatim `--- FAIL` lines); hang diagnosis rests on stack trace + isolation pass.
- Lane independent re-observation (this run, tree @ 75e37d0f0b, env-scrubbed `unset GOFLAGS &&` compound): `go build ./...` OK; `go vet ./internal/cli/... ./internal/homestate/...` OK; `go test ./internal/cli/ -run 'TestFactoryNextLeasePropagatesQueueSpecID' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 16.507s`; `gofmt -l` on the 3 files → empty; `golangci-lint run internal/cli/... internal/homestate/...` → `0 issues.` exit=0.
- Commit verified by lane read: 75e37d0f0b (parent c4784db4c2) — subject `fix(factory): propagate the queue item's spec_id into lease-path record creation (card t1921)`, body names the three sites, ends `🗿 MoAI`; diffstat exactly 3 files (factory_card.go +24/-1, new test +117, card_picked.go +6); tree clean; NOT pushed.
- Verdict: run-phase criteria met (regression ok + build/vet/lint 0, RED-first observed). Card leaves run.

## 8. Landing record — defect ② (dispatch: landing-record only, no repair)

Defect ② (card body; NOT repaired per leader dispatch): `internal/cli/factory_card_pr.go:253-255` — the SPEC-less branch's checkSyncAudit reads progress.md via a process-cwd-relative path, so under the MCP server (cwd=primary checkout) the file reads absent and only a worktree-cwd CLI run passes; an empty specDir should anchor to RepoDir. Recorded here as the landing record the dispatch requires; F3 adjacent records kept untouched.
