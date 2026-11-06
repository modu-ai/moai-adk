---
card: t2118
reviewed_sha: 734f8ac819f7a1706d3214405a583b2a9d4cf54f
reviewer: lane-28 self-review (Claude)
codex_review: not run — stage codex budget used (rule d-20261106T080039Z-305c)
---

# t2118 card review

## Why this is a self-review

Standing rule `d-20261106T080039Z-305c` allows at most 2 codex calls per audit stage, made by the auditor. The sync stage used both:
- `rcpt-c403c403d97e3c1d034d1dc7` (iter-1, pass)
- `rcpt-e273b2e393454f11fea57e3e` (iter-2 delta from f7f536ade9, pass)

A third iter-2 attempt was rejected before any review ran, because the `delta_base` was mistyped. It produced no receipt. To stay inside the rule, the lane did not call `codex_review` for this card review. The two codex receipts above are this card's codex evidence. This file is advisory, as is any card review; the binding verdict is `sync-audit-iter2.md`.

## Diff reviewed (04d8c27a7e..734f8ac819)

- `internal/sqliteuri/sqliteuri.go` (new): `File(path, query)`, a pure string builder.
  - Drive paths become `file:///C:/…`; backslashes become `/`.
  - Relative paths are made absolute.
  - Query is encoded by `url.Values.Encode`.
  - The coverage figure (100%) is the auditor's.
- Production switch sites, 9 hand-built and 2 `url.URL` forms:
  - ledger `ledgerDSN`
  - homestate `runtime_census`
  - discovery `factory_discovery`
  - factorymsg `run_state` ×3 and `store` ×2
  - cli `factory_card` and `migrate_home_state` ×2
- Test helper `qasReadOnlyDSN`, from `internal/cli/factory_quota_lanes_test.go`.
- Guard test: forbids `"file:"+` in production code.

## Lane observations

- RED-before-fix holds for both fixes: `e267f5f4a6` precedes `9deb412a48`, and `dfefa51452` precedes `95dfec79b5`.
- Lane-run evidence for this card:
  - `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
  - `go test ./internal/sqliteuri/ ./internal/agenticfactory/ledger/` passes.
  - `TestQAS_AC018` and `TestQAS_AC021` pass (10.9s and 14.5s).
- Security: the path-traversal flag at `schema.go:85` is a false positive. The lane and both audits agree that the path is operator-owned and reached the driver before the change. The switch also closes URI-parameter injection through `?` in a path at the former concatenation sites.

## Debt carried (non-blocking, from the audits)

- F3: the guard misses the `url.URL{Path: …}` form.
- F4: 7 equivalent builders remain.
- F5: edge inputs that no caller reaches.
- F6: test helpers that still use the `"file:"+…` form may surface once Windows CI gets past `ledger.Open`. The Windows race matrix is the decider.
