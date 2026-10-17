# t1854 run-phase evidence

Card t1854 (Class B): `auditReadyRecorded` ignored the bulleted `- audit_ready: true` form in progress.md §E.1.

## Claim

The fix at tree commit `343a33746` makes the kickoff decider accept `- ` / `* ` bulleted `audit_ready` lines and name an ignored look-alike key in its refusal, and the `internal/homestate` and `internal/template` packages pass in full.

## Evidence

Run inside the `go-test-heavy` slot (acquired by lane-23 at 2026-10-17T15:40:15Z, released after the run; `moai slot status` read `free` afterwards), in the card worktree:

```
go test ./internal/homestate ./internal/template -count=1
ok  	github.com/modu-ai/moai-adk/internal/homestate	242.363s
ok  	github.com/modu-ai/moai-adk/internal/template	325.363s
rc=0
```

Pre-fix failure (RED, 74f2864c8): `.moai/reports/t1854/cause-evidence.md` — 5 of 10 `TestAuditReadyBulletForms` subtests failed against the old parser.

Post-fix targeted run, same tree: `go test ./internal/homestate -run 'TestAuditReadyBulletForms|TestFDA_AuditDecider' -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/homestate	48.441s`.

## Baseline-attribution

Tree: `.moai/worktrees/t1854`, HEAD `343a33746` (RED `74f2864c8` precedes the fix on the commit graph). Tests were built from this tree with the toolchain's `go test`; the installed `moai` binary was used only for slot acquire/release/status and never judged the tree.

## Gaps

- The landing re-measure (changed packages plus direct importers) has not run; the leader has scheduled it after t1856 lands.
- The test log was machine-local scratch; the three result lines above are the carried-over deciding output.
- `gofmt -l` and `go vet` on the touched packages were reported by the implementing agent, not re-run by the lane.

## Residual-risk

- Importers of the two packages other than those in the test batch were not exercised.
- A prose line in §E.1 whose key normalises to a prefix of `audit_ready` now adds a hint to the refusal text; the hint only appears when the signal is not accepted.
