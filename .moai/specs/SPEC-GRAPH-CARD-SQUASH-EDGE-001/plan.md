# Implementation Plan — SPEC-GRAPH-CARD-SQUASH-EDGE-001

## §A Context

- Card t1559, lane-3 self-dispatch run tmhxo0. Tier M. cycle_type **tdd** (RED → GREEN → verify).
- Working tree: this card worktree, branch `WT-graph-card-merges-edge`. Plan base re-pinned at
  **5fb7baf88** (the lane's origin/main absorb after plan-audit iter 1 — plan §C.1); all
  "unchanged" and scope checks resolve against this SHA, never a moving ref. Authoring-time
  baselines were measured at df0c417e9 (pre-absorb) and stay in §C.2 as historical anchors.
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
- Intermediate-commit attribution is accepted (lane decision, spec.md §3): every intermediate
  branch commit carries the card id by the traceability mandate, and pre-cutover `--no-ff`
  merges make those commits reachable from `main` — so edges and the fingerprint now span a
  card's intermediate commits, on `main` as well as in in-flight worktrees. Semantically
  consistent landed evidence; a one-time historical enrichment with no ongoing churn under
  squash landing (each new landing attributes exactly one new commit). No carve-out — a
  walk-side subject filter would contradict REQ-GCSE-001 and plan §G.

## §C Pre-flight

### §C.1 Absorbed-state precondition (re-pinned 2026-10-07 — D4)

Plan-audit iter 1 found the verification commands failing with `[setup failed]`:
`internal/template/embed_manifest_gen.go:71` still embeds `handle-agent-hook.sh`, which card
t1540 (d9b1334d2, PR #1779) had deleted from the tree. The heal landed on origin/main —
f97edcc55 (PR #1762), plus #1783/#1784 — and the lane absorbed origin/main into the card
branch. Post-absorb HEAD is the merge commit **5fb7baf88**, the new plan base. Measured
post-absorb on this tree:

| # | Fact | Command | Observed |
|---|------|---------|----------|
| C7 | Template package compiles | `go build ./internal/template/` | exit 0 |
| C8 | Graph suite green | `go test -count=1 ./internal/graph/` | ok 107.741s |

### §C.2 Authoring-time baselines (measured at df0c417e9, pre-absorb — historical anchors)

| # | Fact | Command | Observed |
|---|------|---------|----------|
| C1 | HEAD was single-parent (squash) | `git rev-list --parents -1 HEAD` | `df0c417e971… cb2a011d03…` — exactly one parent |
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
- E4 scope — `git diff --name-only 5fb7baf88 HEAD -- internal/` prints exactly two lines (the
  two sanctioned paths; product paths only — the SPEC documents themselves are commits in this
  range and are excluded by the pathspec).
- E5 ordering — `git log --oneline 5fb7baf88..HEAD` shows the fix commit newest and the
  test-only commit immediately before it.

## §F Milestones (decision-reversibility order)

### M1 — RED (Priority High)

1. Add `TestGraphCardFileEdgesSeeSquashLanding` to `internal/graph/card_file_test.go`:
   fixture = base history on `main`, a card branch committing `squash.txt`, then
   `git merge --squash cardbr` + `git commit -m "fix(graph): repair the card-file squash blind
   spot (card t1560) (#1999)"` on `main` — a single-parent landing; assert exactly one edge
   `{t1560, squash.txt, <9-char SHA>}`, the landing SHA in `CardAttributedMergeSHAs`, and a
   `CardMergeFingerprint` difference against the same fixture without the landing (control arm).
   The test also carries AC-GCSE-002's root contrast group: a second fixture whose ROOT
   (parentless) commit subject carries a card token (`fix(t1561): seeded root (card t1561)`),
   asserting nil error + zero edges so the first-parent-diff guard path (card_file.go:95-98) is
   actually exercised through the attribution gate (:86-89).
2. Run `go test ./internal/graph/ -run '^TestGraphCardFileEdgesSeeSquashLanding$' -v -count=1` —
   observe FAIL (0 edges, want 1); record verbatim output in progress.md. The expected RED is
   an assertion failure — a `[setup failed]` compile error is a defect to fix first, not the
   RED being sought (the post-absorb tree compiles: plan §C.1).
3. Commit test only: `test(graph): reproduce the squash-landing blind spot in the card-file walk (card t1559)`.

Files: `internal/graph/card_file_test.go`.

### M2 — GREEN (Priority High)

1. `internal/graph/card_file.go`: remove `--merges` from the walk argv; rename
   `walkCardMerges` → `walkCardCommits` and `mergeInfo` → `commitInfo`; update the file-top
   comment, the walk doc comment, the `CardFileEdges` / `CardAttributedMergeSHAs` /
   `CardMergeFingerprint` doc comments, the per-commit diff comments, and the error string to
   commit-level semantics (keep the root-commit guard and its rationale, broadened to "commit").
2. Run the new test GREEN, then
   `go test ./internal/graph/ -run '^TestGraph(CardFile|CheckNoticesCardFileSource)' -v -count=1`
   — all six existing tests pass unmodified (seven `-v` PASS lines at M2 time: six existing +
   the new one). Leading anchor only, deliberately no `$`: the selector is prefix-open — the
   full-wrap form the linter suggests would match only `TestGraphCheckNoticesCardFileSource`
   (card_file_test.go:146) and silently drop the six `TestGraphCardFile*` tests (1 of 7
   selected); none of these tests uses subtests.
3. Commit: `fix(graph): walk every commit so squash landings produce card edges (card t1559)`.

Files: `internal/graph/card_file.go`.

### M3 — Gates and evidence (Priority Medium)

1. `go vet ./internal/graph/...` · `golangci-lint run ./internal/graph/...` ·
   `go test -timeout 30m ./internal/graph/...` — all exit 0; the package run's wall-clock is
   recorded once in progress.md (the measured cost of the broadened walk — D2 residual).
2. `git diff --name-only 5fb7baf88 HEAD -- internal/` — exactly two lines: the two sanctioned
   paths (product paths only; the SPEC documents are commits in this range).
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
