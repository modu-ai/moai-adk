# acceptance.md — SPEC-GRAPH-CARD-SQUASH-EDGE-001

## §A Verification Principles

- Every criterion is binary-testable and judged by the exit code or count of a named command;
  outputs are recorded verbatim in progress.md §E.2.
- All commands are plain one-git-per-line form (worktree-guard safe); no `git` inside `$()`.
- "Unchanged" and scope checks resolve against the pinned plan base **5fb7baf88** (the lane's
  origin/main absorb — plan §C.1), never a moving ref; the scope check is pathspec-restricted
  to product paths (`-- internal/`).
- The RED arm is re-witnessed at run phase (plan M1 step 2) with verbatim output — the
  plan-phase baseline (spec.md §1) establishes the mechanism; M1 establishes the behavior.

## §B AC Matrix

This matrix maps REQ-GCSE-001 through REQ-GCSE-012 to AC-GCSE-001 through AC-GCSE-011 (some
criteria cover a REQ pair; every REQ has at least one owning AC).

| AC | REQ | Criterion |
|----|-----|-----------|
| AC-GCSE-001 | REQ-GCSE-001, REQ-GCSE-002, REQ-GCSE-009 | RED-first squash-landing reproduction |
| AC-GCSE-002 | REQ-GCSE-004, REQ-GCSE-007 | Over-attribution control arm |
| AC-GCSE-003 | REQ-GCSE-003 | Fingerprint input convergence |
| AC-GCSE-004 | REQ-GCSE-005 | Single attribution point |
| AC-GCSE-005 | REQ-GCSE-006, REQ-GCSE-007 | Breadth + absorb refusal preserved |
| AC-GCSE-006 | REQ-GCSE-008 | Determinism carried over |
| AC-GCSE-007 | REQ-GCSE-010 | Six existing tests unmodified in intent |
| AC-GCSE-008 | REQ-GCSE-011 | Doc-comment truthfulness + rename completeness |
| AC-GCSE-009 | REQ-GCSE-012 | Package gates |
| AC-GCSE-010 | REQ-GCSE-011 | Scope guard |
| AC-GCSE-011 | REQ-GCSE-009 | RED→GREEN ordering witnessed by the commit graph |

### AC-GCSE-001 — RED-first squash-landing reproduction

- **Given** a fixture repo under `t.TempDir()` whose `main` history carries a base commit and
  then a `git merge --squash cardbr` + `git commit` landing `squash.txt` with subject
  `fix(graph): repair the card-file squash blind spot (card t1560) (#1999)` — a single-parent
  commit, reachable from HEAD, attributed through `factory.AttributeSubject`.
- **When** `TestGraphCardFileEdgesSeeSquashLanding` runs `CardFileEdges` and
  `CardAttributedMergeSHAs` over the fixture.
- **Then** exactly one edge `{Card: t1560, File: squash.txt, SHA: <9-char abbreviated SHA>}` is
  returned, the SHA list contains the landing's full SHA, and `CardMergeFingerprint` over this
  fixture differs from the same fixture without the landing.
- **Command**: `go test ./internal/graph/ -run '^TestGraphCardFileEdgesSeeSquashLanding$' -v -count=1`
- **Judgement**: fails on the pre-fix tree (0 edges, want 1 — RED, plan M1 step 2); passes
  after M2 (GREEN).

### AC-GCSE-002 — Over-attribution control arms

- **Given** the same fixture's base arm: the card branch's intermediate commit is NOT reachable
  from `main` (squash flattens it away), and the reachable base commits carry no card token.
- **When** `CardFileEdges` and `CardAttributedMergeSHAs` run over the pre-landing base history.
- **Then** zero edges and an empty SHA list — the broadened walk attributes nothing unearned.
- **Given (root contrast group)** a second fixture whose ROOT (parentless) commit itself
  carries a card token — subject `fix(t1561): seeded root (card t1561)` — so the attribution
  gate returns non-empty for a parentless commit.
- **When** `CardFileEdges` runs over that fixture.
- **Then** it returns nil error with zero edges — the first-parent-diff failure guard
  (card_file.go:95-98 `continue`) is exercised for real, because the token-less base arm above
  skips at the attribution gate (:86-89) before the diff ever runs; a mutant turning the guard
  into a hard error return flips this assertion while leaving the base arm green.

### AC-GCSE-003 — Fingerprint input convergence

- **Given** two fixtures identical except for the squash landing (with / without).
- **When** `CardAttributedMergeSHAs` and `CardMergeFingerprint` run on both.
- **Then** the with-landing SHA list contains exactly the landing SHA, the without-landing list
  is empty, and the two fingerprint values differ.

### AC-GCSE-004 — Single attribution point

- **Given** the post-fix `internal/graph/card_file.go`.
- **When** `grep -cE 'regexp\.|MatchString' internal/graph/card_file.go` runs.
- **Then** exit 1 with 0 matches — the layer still holds no subject-matching logic; attribution
  flows only through the `CardFileAttributor` parameter. (Baseline measured 2026-10-07 pre-fix:
  0 matches, exit 1 — plan §C C5.)

### AC-GCSE-005 — Breadth + absorb refusal preserved

- **Given** the unchanged fixtures of the existing absorb tests.
- **When** `go test ./internal/graph/ -run '^(TestGraphCardFileEdgesSeeAbsorbedMerge|TestGraphCardFileEdgesIgnoreAbsorbMerge)$' -v -count=1`
- **Then** both pass unmodified — the t7 second-parent edge still appears (MU-86) and the
  absorb-direction merge still yields no edge and no fingerprint input (MU-87).

### AC-GCSE-006 — Determinism carried over

- **When** `go test ./internal/graph/ -run '^TestGraphCardFileEdgesDeterministic$' -v -count=1`
- **Then** passes — two runs byte-identical for edges and for the fingerprint.

### AC-GCSE-007 — Six existing tests unmodified in intent

- **When** `go test ./internal/graph/ -run '^TestGraph(CardFile|CheckNoticesCardFileSource)' -v -count=1`
  (leading anchor only — prefix-open selector; the full-wrap form would match only
  `TestGraphCheckNoticesCardFileSource`, card_file_test.go:146 — 1 of 7 selected, the six
  `TestGraphCardFile*` names silently dropped) and
  `git diff 5fb7baf88 HEAD -- internal/graph/card_file_test.go`.
- **Then** all six existing tests pass (each named on its own `-v` PASS line) and the diff
  shows only the added `TestGraphCardFileEdgesSeeSquashLanding` — no existing fixture pattern
  edited.

### AC-GCSE-008 — Doc-comment truthfulness + rename completeness

- **When** `grep -c 'walkCardMerges' internal/graph/card_file.go` (baseline: 4, plan §C C4)
  and `grep -rn 'walkCardMerges' internal/` run after M2.
- **Then** the file-local count is 0 and the repo-wide scan finds 0 hits; the file-top comment
  and the walk/edge/fingerprint doc comments state commit-level semantics (no residual
  "merge commits whose subject" claim).

### AC-GCSE-009 — Package gates

- **When** `go vet ./internal/graph/...` · `golangci-lint run ./internal/graph/...` ·
  `go test -timeout 30m ./internal/graph/...`
- **Then** all three exit 0.

### AC-GCSE-010 — Scope guard

- **When** `git diff --name-only 5fb7baf88 HEAD -- internal/` (product paths only — the SPEC
  documents themselves are commits in this range and are excluded by the pathspec)
- **Then** the output is exactly two lines — `internal/graph/card_file.go` and
  `internal/graph/card_file_test.go` — no other path. (Base re-pinned to the absorbed merge
  5fb7baf88; the authoring base df0c417e9 predates the SPEC-document commits, so an unpinned
  whole-tree diff can never yield two lines.)

### AC-GCSE-011 — RED→GREEN ordering witnessed by the commit graph

- **When** `git log --oneline 5fb7baf88..HEAD`
- **Then** the newest commit is the fix commit and the next-older commit is the test-only
  commit — the test commit precedes the fix commit, so the RED-first claim is re-witnessable
  from git history alone (SPEC-V3R6-GRAPH-FRESHNESS-001 §2.3 ordering attribution).

## §C Quality Gate Criteria

- **Tested**: the new reproduction test plus the six existing tests; the owning package
  re-measured (`go test -timeout 30m ./internal/graph/...` exit 0).
- **Readable / Unified**: comments state commit-level semantics in English; gofmt clean via
  `golangci-lint run ./internal/graph/...`.
- **Secured**: no new input surface (same `git` subprocess, one flag removed); the
  NUL-separated parsing discipline kept (security lens, dispatch `--security`).
- **Trackable**: commits carry the card id (`card t1559`); evidence paths under
  `.moai/specs/SPEC-GRAPH-CARD-SQUASH-EDGE-001/`.

## §D Definition of Done

- Every §B criterion judged PASS with verbatim command output recorded in progress.md §E.2.
- Zero MUST-FIX findings from plan-audit before run-phase entry; sync-phase re-verifies
  AC-GCSE-009/010 on the landing tree.
