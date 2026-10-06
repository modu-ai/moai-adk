# Implementation Plan — SPEC-GRAPH-CARD-SQUASH-EDGE-001

## §A Context

- Card t1559, lane-3 self-dispatch run tmhxo0. Tier M. cycle_type **tdd** (RED → GREEN → verify).
- Working tree: this card worktree, branch `WT-graph-card-merges-edge`, plan base pinned at
  **df0c417e9** (all "unchanged" and scope checks resolve against this SHA, never a moving ref).
- Defect in one sentence: the `--merges` filter at internal/graph/card_file.go:49 predates the
  2026-10-05 squash-landing flow, and attribution (`factory.AttributeSubject` forms 1/2/2b/2c)
  already reads squash subjects — removing the filter is the whole fix.
- Provenance: discovered by lane-6 card-review + lane-12 gate double convergence; sync-gate
  lenses `--security --deep`.

## §B Known Issues

- The freshness source name `card-merges` (meta.go:28) now covers squash landings too, but its
  spelling is unchanged — a naming imprecision this SPEC records and does not fix (meta.go is
  excluded by the scope guard).
- After this fix lands, any graph build stamped before it reads **stale once** — previously
  missed squash landings now enter the fingerprint. That is the desired convergence, not a
  regression; one rebuild cycle follows the fix landing.
- `CardAttributedMergeSHAs` and `CardMergeFingerprint` names become broader than their content
  ("Merge" → any attributed landing commit). The doc comments state the commit-level semantics;
  renaming the exported identifiers is out of scope.
- In a card worktree mid-flight, the broadened walk attributes intermediate card-branch commits
  (the traceability mandate puts the card id in every commit subject), so edges and the
  fingerprint pick up unlanded SHAs while the card runs. On `main` only the landed squash is
  reachable — the CI-facing surface this SPEC targets. Accepted behavior, no carve-out (spec.md
  §3).

## §C Pre-flight (measured, this tree, 2026-10-07)

| # | Fact | Command | Observed |
|---|------|---------|----------|
| C1 | HEAD is single-parent (squash) | `git rev-list --parents -1 HEAD` | `df0c417e971… cb2a011d03…` — exactly one parent |
| C2 | Attribution reads the squash shape | HEAD subject vs form-2b ERE | 1 match |
| C3 | Walk breadth delta | `git rev-list --count HEAD` / `--merges` | 13127 / 2664 |
| C4 | Rename baseline | `grep -c walkCardMerges internal/graph/card_file.go` | 4 (target after M2: 0) |
| C5 | Layer holds no matching logic | `grep -cE 'regexp\.\|MatchString' internal/graph/card_file.go` | 0, exit 1 (stays 0 after M2) |
| C6 | Test helper | `gitFix` exists in package graph tests (used by all six existing tests) | available |

## §D Constraints

- [HARD] Scope: only `internal/graph/card_file.go` and `internal/graph/card_file_test.go`
  change. `meta.go`, `graph.go`, and `internal/factory/**` are read-only for this SPEC.
- [HARD] No exported identifier is renamed (`CardFileEdges`, `CardAttributedMergeSHAs`,
  `CardMergeFingerprint`, `CardFileAttributor`, `CardFileEdge` keep their names). File-local
  renames sanctioned: `walkCardMerges` → `walkCardCommits`, `mergeInfo` → `commitInfo`.
- Attribution stays the caller-supplied `CardFileAttributor`; no second matcher
  (REQ-GCSE-005, MU-87). Walk breadth stays any-parent-path (MU-86).
- Keep the NUL-separated parsing discipline (`%H%x00%s` subject split, `-z` diff split) — no
  format change, no new input surface (security lens).
- Comments in English (`code_comments: en`); error string `card_file: log merges:` becomes
  `card_file: log:`.
- Commits carry the card id; the test commit precedes the fix commit (ordering witnessed by the
  commit graph, SPEC-V3R6-GRAPH-FRESHNESS-001 §2.3 discipline).

## §E Self-Verification

- E1 AC matrix — every acceptance.md §B criterion judged with verbatim command output.
- E2 package re-measure — `go test -timeout 30m ./internal/graph/...` exit 0.
- E3 lint — `go vet ./internal/graph/...` and `golangci-lint run ./internal/graph/...` exit 0.
- E4 scope — `git diff --name-only df0c417e9 HEAD` lists exactly the two sanctioned paths.
- E5 ordering — `git log --oneline df0c417e9..HEAD` shows the fix commit newest and the
  test-only commit immediately before it.

## §F Milestones (decision-reversibility order)

### M1 — RED (Priority High)

1. Add `TestGraphCardFileEdgesSeeSquashLanding` to `internal/graph/card_file_test.go`:
   fixture = base history on `main`, a card branch committing `squash.txt`, then
   `git merge --squash cardbr` + `git commit -m "fix(graph): repair the card-file squash blind
   spot (card t1560) (#1999)"` on `main` — a single-parent landing; assert exactly one edge
   `{t1560, squash.txt, <9-char SHA>}`, the landing SHA in `CardAttributedMergeSHAs`, and a
   `CardMergeFingerprint` difference against the same fixture without the landing (control arm).
2. Run `go test ./internal/graph/ -run '^TestGraphCardFileEdgesSeeSquashLanding$' -v -count=1` —
   observe FAIL (0 edges, want 1); record verbatim output in progress.md.
3. Commit test only: `test(graph): reproduce the squash-landing blind spot in the card-file walk (card t1559)`.

Files: `internal/graph/card_file_test.go`.

### M2 — GREEN (Priority High)

1. `internal/graph/card_file.go`: remove `--merges` from the walk argv; rename
   `walkCardMerges` → `walkCardCommits` and `mergeInfo` → `commitInfo`; update the file-top
   comment, the walk doc comment, the `CardFileEdges` / `CardAttributedMergeSHAs` /
   `CardMergeFingerprint` doc comments, the per-commit diff comments, and the error string to
   commit-level semantics (keep the root-commit guard and its rationale, broadened to "commit").
2. Run the new test GREEN, then
   `go test ./internal/graph/ -run '^TestGraph(CardFile|CheckNoticesCardFileSource)' -count=1`
   — all six existing tests pass unmodified. Leading anchor only, deliberately no `$`: the
   selector is prefix-open (`TestGraphCardFile` names no test exactly — the linter's full-wrap
   form would select nothing), and none of these tests uses subtests.
3. Commit: `fix(graph): walk every commit so squash landings produce card edges (card t1559)`.

Files: `internal/graph/card_file.go`.

### M3 — Gates and evidence (Priority Medium)

1. `go vet ./internal/graph/...` · `golangci-lint run ./internal/graph/...` ·
   `go test -timeout 30m ./internal/graph/...` — all exit 0.
2. `git diff --name-only df0c417e9 HEAD` — exactly the two sanctioned paths.
3. `moai spec lint` on the SPEC artifacts — 0 errors.
4. Record §E evidence in progress.md; hand to plan-audit.

No further file changes unless a gate fails; any repair re-runs the whole affected-package
family of every touched function.

## §G Anti-Patterns

- Filtering commits inside the walk by subject or shape — attribution is the only card filter
  (REQ-GCSE-001); a walk-side filter reintroduces the blind spot one layer down.
- A second subject matcher in the layer (REQ-GCSE-005, MU-87).
- Renaming exported identifiers or touching `meta.go` / `graph.go` / `internal/factory/**` (§D).
- Re-implementing the NUL parsing or changing the `%H%x00%s` / `-z` formats (t1454 finding 16
  discipline).
- Absorbing `origin/main` mid-run without re-running §C measurements — an absorb invalidates
  pinned counts; re-measure C1–C5 after any absorb.

## §H Cross-References

- `spec.md` / `acceptance.md` / `progress.md` — same SPEC directory.
- `SPEC-TODO-CARD-ISSUANCE-001` — origin of the card→file edge layer (REQ-TCI-016).
- `SPEC-WORKTREE-SQUASH-MERGE-001` — sibling squash-landing detection in worktree cleanup.
- `SPEC-V3R6-GRAPH-FRESHNESS-001` — freshness/baseline-attribution discipline (§2.3 ordering).
