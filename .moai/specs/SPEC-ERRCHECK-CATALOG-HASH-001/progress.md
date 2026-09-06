# SPEC-ERRCHECK-CATALOG-HASH-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

SPEC artifact set authored 2026-09-06 by manager-spec (card t489, lane-15) at Tier M compact
scope (`tier: M` in frontmatter): spec.md (6 GEARS requirements) + plan.md (single milestone M1) +
acceptance.md (5 ACs, all must-pass) + this progress skeleton. Plan-audit iteration 1 returned
FAIL with 5 findings (D1-D5), all applied 2026-09-06 — lifecycle enum corrected to
`spec-anchored`, `tier: M` added, AC-SCOPE extended with a diff-shape predicate, AC-MUTANT revert
mechanism corrected to a valid tree-ish form, AC-GREEN bound additionally to the package summary
line. SPEC ID regex pre-write check: PASS. Not committed — the lane session commits with explicit
pathspecs.

## §E.2 Run-phase Evidence

REQ-002 fix applied 2026-09-06 (manager-develop spawn, card t489 lane-15) — single edit,
`internal/template/catalog_tree_hash.go` inside `ComputeDirTreeHash`:

- The bare call `fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)` (former line 60) is now prefixed with
  the explicit blank discards `_, _ =` and carries a two-line English comment citing the reason:
  the write target is a `hash.Hash` (from `sha256.New()`) whose `Write` never returns an error per
  the stdlib contract, so the error is structurally always nil (REQ-003 — error propagation would
  be a dead path). Two discards, not one: `fmt.Fprintf` returns `(n int, err error)`, so a
  single-value discard fails to compile ("assignment mismatch: 1 variable but fmt.Fprintf returns
  2 values" — observed by the lane's `go build`); the landed form is
  `_, _ = fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)`.
- No other file changed; no `//nolint` directive introduced anywhere (REQ-003 + AC-SCOPE
  clause (c)(ii)). `.golangci.yml` pins `check-blank: false`, so the discard passes errcheck.
- No lint/test/build was run by this agent — the lane (single-measurer rule) performs all
  measurements and owns all commits (explicit pathspecs, no agent commit).

Intended verification (lane-executed, per acceptance.md §A — the single judge for all ACs):

```bash
golangci-lint run ./internal/template/... --timeout=2m
```

- AC-GREEN: zero errcheck matches (file-scoped `catalog_tree_hash.go` findings AND package
  summary `* errcheck: 0`).
- AC-MUTANT: revert via `git checkout 615d18c1f -- internal/template/catalog_tree_hash.go` (or the
  recorded pre-fix blob `git show 735b7d7b4856088dc7702a56fd450c0800327666 > …`), re-run, expect
  exactly 1 errcheck finding at `catalog_tree_hash.go:60:14`, then re-apply and re-confirm green.
- AC-SCOPE closure check: `git diff 615d18c1f..HEAD -- '*.go'` contains NO `//nolint`.

Lane-measured results (single-measurer rule — all commands below were run by the lane session,
not by this agent; verbatim outcomes relayed by the lane):

| Run | Tree | Command | Observed output |
|-----|------|---------|-----------------|
| B — GREEN | `9f341f371` (fix committed) | `golangci-lint run ./internal/template/... --timeout=2m` | `0 issues.` |
| C — MUTANT | file checked out at base blob `735b7d7b`, HEAD `9f341f371` | same command | `internal/template/catalog_tree_hash.go:60:14: Error return value of ... (errcheck)` · `* errcheck: 1` |
| D — re-confirm | file restored to HEAD | same command | `0 issues.` |

Post-correction compile check (lane-run): `go build ./internal/template/` — silent, exit 0.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: PASS
measured_at_tree: 175619586
verdict_artifact: .moai/reports/t489/verdict.md
measurer: lane-15 session (all commands run and measured by the lane; this agent implemented the edit only)
ac_pass_count: 5
ac_fail_count: 0
new_warnings_or_lints_introduced: 0
```

AC verdicts (all PASS; each pointer resolves inside `.moai/reports/t489/verdict.md` at tree
`175619586`):

- AC-GREEN — PASS → verdict.md (Run B: `0 issues.`)
- AC-MUTANT — PASS → verdict.md (Run C: exactly 1 errcheck finding at `catalog_tree_hash.go:60:14`; Run D re-confirmed `0 issues.`)
- AC-ORDERING — PASS → verdict.md (RED baseline commit precedes the fix commit on the card branch)
- AC-SCOPE — PASS → verdict.md (linter-only verification; diff shape `_, _ =` + comment; no `//nolint` in `git diff 615d18c1f..HEAD -- '*.go'`)
- AC-EVIDENCE — PASS → verdict.md (evidence committed under `.moai/reports/t489/`; every card-branch commit carries `t489`)

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: 5ad9c0b51
sync_status: completed
```

Close summary (card t489, sync commit carries the `draft → completed` transition — the run phase
never committed an intermediate `in-progress` state; the single close commit shape is intended):

- Fix shipped: `_, _ =` discard of the structurally-nil `fmt.Fprintf` error in
  `internal/template/catalog_tree_hash.go` (one line + comment; no behavior change).
- Evidence: `.moai/reports/t489/verdict.md` (primary), §E.2/§E.3 above (all AC PASS).
- SHA placeholder backfill (D3): a commit cannot cite its own SHA; `pending-backfill-sync` is the
  canonical placeholder and the real SHA is backfilled in a follow-up commit by the lane.

CHANGELOG skip rationale: the source change is a single-line internal lint repair with no
user-facing behavior change, so no `CHANGELOG.md` entry is emitted (precedent: t487 / t472).
This note is the record of that decision.
