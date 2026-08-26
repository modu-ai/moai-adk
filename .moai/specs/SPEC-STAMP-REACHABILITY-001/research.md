# research.md — SPEC-STAMP-REACHABILITY-001

All code readings were performed in this worktree (`.claude/worktrees/t291`, branch `WT-stamp-orphan-guard`, base `da791eb0a`). Premises carried from the dispatch were independently re-measured where load-bearing; each measurement cites its command below. Version 0.1.0 · manager-spec · 2026-08-27.

## 1. Incident record (read, not re-derived)

Source: `.moai/reports/t279/triage-table.md` §F5 (tracked, on base). Chain: PR #1648 squash merge (`6786c3fa4`) landed without carrying t250's stamp commit `0d15864ae90b` into main history → stamp unreachable from main → `graph check` not-comparable → red inherited by every subsequent PR (#1662 measured). Emergency repair shipped by SPEC-V3R6-GRAPH-FRESHNESS-002 M0 (`52f7ba135`, restamp vs `c9eed8ac6`), with the measured regeneration ordering `mx scan --quiet → graph build → graph stamp codemaps → graph build (again; stamp rewrites provenance.json and stales the edges fingerprint) → graph check`.

## 2. Workflow job structure (measured)

`.github/workflows/graph-freshness.yml` (50 lines, read in full):

| Aspect | Measured value |
|---|---|
| Triggers | `pull_request` (all branches) + `push` → `branches: [main]` |
| Permissions | `contents: read` |
| Concurrency | group per-ref, cancel-in-progress |
| Job steps | checkout@v7 `fetch-depth: 0` → setup-go@v7 → `go build -o ./bin/moai ./cmd/moai` → bootstrap (`mx scan --quiet`; `graph build`) → `./bin/moai graph check` |
| Existing comment | documents that depth-0 fetch exists because "the codemaps provenance stamp names an ancestor commit that a depth-1 shallow checkout does not hold" (run 32775609689 cited) |

Guard placement decision: immediately after `actions/checkout`, BEFORE setup-go/build. Rationale: the guard needs only git + jq (both preinstalled on ubuntu-latest runners) and the tracked JSON — an orphan-bound stamp should fail in seconds, not after paid Go-setup minutes. The `fetch-depth: 0` guarantee the workflow already documents makes the history-complete check possible in this checkout design. On `push` events, ancestry-versus-base is vacuous (`base == head`); the step's object-existence clause is the meaningful check there and pins the generic exit 2 with a defect-class name.

GHA plumbing facts used: `github.base_ref` carries the target branch name for pull_request events (resolved as `origin/${{ github.base_ref }}`; valid because fetch-depth 0 materializes remote refs); `jq` selection `.commit_sha // empty` handles the omitempty field; on dirty-anchor provenance the clause prints a skip-with-reason and exits 0 (REQ-SR-003).

## 3. Stamp path code reading

- `internal/cli/graph_stamp.go`: single subcommand `stamp codemaps`; `RunE` → `resolveGraphRoot(rootArg)` → `mx.StampCodemaps(projectRoot)` → marshal → tmp-file write → `os.Rename` install (atomicity contract) → deferred temp cleanup with path-free reporting. Only flag: `--root` (line 106). The `--commit` insertion point is exactly the `StampCodemaps` call — everything downstream (marshal/install/reporting) is shared.
- `internal/mx/provenance.go`: `baseProvenance(projectRoot, generatedBy, describedRoots)` decides the anchor exclusively: `treeDirty()` true → `Dirty:true` + aggregate fingerprint; else `CommitSHA = GitHead(projectRoot)` (lines 185-203). There is **no way today** to record any commit other than detected HEAD. `GitHead` uses `gitOut(root,"rev-parse","HEAD")` (fail-open "").
- Extension shape (delegated-execution latitude left to run phase): thread an optional explicit commit through `StampCodemaps` — e.g. a variadic/options param or a sibling `StampCodemapsWithCommit(root, rev)` wrapper that resolves once via `git rev-parse --verify <rev>^{commit}` then delegates; the CLI validates resolvability (non-zero + path-free stderr on failure) BEFORE any filesystem write, honoring the tmp/rename contract. `described_roots` (SSOT: `DefaultDescribedRoots`, commented as deliberately non-configurable) are unaffected.
- Dirty interaction: when `treeDirty` holds and `--commit` is supplied, the invocation must fail pre-write (REQ-SR-006) — the two honesty claims cannot coexist in the schema (one `commit_sha`, one `content_fingerprint`; `omitempty` fields would silently blank whichever loses).
- Check-side coupling worth documenting: `CheckFreshness` diffs described-source working-tree content against the CONTENT AT THE NAMED COMMIT. An explicitly named revision therefore must be content-faithful, not merely reachable: naming `origin/main` after main has moved past the branch point counts other PRs' merged churn as this tree's drift. The correct ergonomic recipe is the merge-base: `--commit "$(git merge-base HEAD origin/main)"` — an ancestor of origin/main (squash-surviving) AND content-equal to the branch's described sources (drift-free). REQ-SR-010 encodes exactly this recipe.

## 4. Check path code reading

- `internal/graph/check.go:118-121`: a stamped commit not comparable in the checkout aborts with system-error semantics; `check.go:209-210` renders `"stamped commit %s not comparable in this checkout"` and returns the error. `internal/cli/graph_check.go:19-25` maps it to exit 2 (`exitSystemError` constant family) — this is the exact CI surface that went red after the #1648 squash. The guard's attributable annotation (names the sha + base ref) complements this generic verdict; replacing or suppressing it is out of scope.
- Exit-code contract 0/1/2 is consumed by `moai gate` and CI (ANCHOR-tagged, fan-in ≥3) — REQ-SR-009 freezes it; no tests may weaken it.

## 5. Live fixture and premise measurements (this run, this tree)

| Claim | Command | Result |
|---|---|---|
| Worktree identity | `pwd; git rev-parse --show-toplevel; git branch --show-current; git rev-parse --short HEAD` | `.../.claude/worktrees/t291` · `WT-stamp-orphan-guard` · `da791eb0a` |
| Orphan premise (dispatch premise re-measured) | `git merge-base --is-ancestor 0d15864ae90b origin/main` | rc≠0 → `NOT_ANCESTOR` — the historic orphan is still a usable negative fixture |
| Committed provenance anchor | `jq -r '.commit_sha' .moai/project/codemaps/provenance.json` | `410da655f39d1d3731abf2f247e8e5353a9d0de5` |
| False-positive premise | `git merge-base --is-ancestor 410da655f… origin/main` | rc=0 → `CURRENT_STAMP_ANCESTOR` — today's stamp passes the planned guard (no immediate red on landing) |
| Schema fidelity | read `provenance.json` | `{schema_version:1, tree_root:<abs>, commit_sha:"410da…", dirty:false, described_roots:[internal,cmd,pkg], generated_by:"codemaps-gen", generated_at}` — matches the dispatched schema; NOTE the committed `tree_root` names the t279 worktree path (informational-only for tracked layers per CR-R2 — expected state everywhere but the stamper's machine, NOT a defect to fix here) |
| SPEC ID collision probe | `ls .moai/specs \| grep -c STAMP-REACHABILITY` | `0` |
| Template mirroring | `ls internal/template/templates/.github/workflows/` | only `label-sync.yml` — graph-freshness.yml is dev-only infra, no Template-First cycle |
| Docs surface | `ls docs-site/content/*/cli-reference/graph.md` | en/ja/ko/zh all present, `moai graph stamp codemaps` section exists in each (4-locale parity obligation applies) |

Fixture caveat for run-phase AC-SP/AC-SR execution: the `0d15864ae90b` object exists LOCALLY (it was fetched — that is precisely why locals saw "fresh"); exercising the object-absence branch needs a history-free context, produced by fetching ONLY `main` into an empty throwaway repo under `/tmp` (t.TempDir()/mktemp discipline). The ancestry branch (negative case) is exercisable directly here.

## 6. Countermeasure adjudication (options as dispatched)

1. **CI pre-merge reachability guard** — SELECTED (primary). Sole pre-orphaning enforcement point; cheap (shell + jq); complementary attribution to the existing exit-2 verdict; zero runtime-API change.
2. **Explicit stamp-commit mode** — SELECTED (ergonomic aid). Closes the gap that made REQ-GFR-014 violable by accident (only HEAD was recordable). Validation split: resolvability at stamp time, survivability at the CI gate (local origin/main staleness makes any local survivability promise dishonest — see §3 recipe analysis).
3. Post-merge chore restamp; merge-convention switch — REJECTED, recorded in spec.md §E (not planned; nothing here depends on them).

## 7. Open items explicitly OUT of this card's scope

- **PR #1666 (t278) recurrence classification** — dispatched as unadjudicated and remains so; orphan-class vs staleness-class does not change this SPEC's mechanism (the guard fires identically either way). Not investigated further here.
- **Predecessor-series relations** — related_specs carries both GRAPH-FRESHNESS SPEC IDs; strict `depends_on` is avoided since the second predecessor is `completed` and would add a lifecycle predicate with no value (its audit trail is frozen; this SPEC starts fresh).
