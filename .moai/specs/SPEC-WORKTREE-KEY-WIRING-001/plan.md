# SPEC-WORKTREE-KEY-WIRING-001 — Plan

## §A Context

Card t655 (Class C, Tier M+, Factory lane-4). `workflow.worktree.auto_merge`
is declared config plumbing with zero production readers;
`workflow.worktree.auto_create` is read but its `true` wording claims an
automation nothing performs. This SPEC wires `auto_merge` to a real, safe,
local-only merge reader beside the `auto_cleanup` consumers, and re-scopes
`auto_create` honestly. Baseline: branch `WT-worktree-keys-wiring`, base
`1d150a27d` (= origin/develop at plan time).

Design decisions are settled in design.md (window = Option A acquire;
target = reuse `develop_branch`; auto_create = explicit scope). Cycle type:
**ddd** — behavior wiring onto existing code with characterization coverage
of the OFF baseline (REQ-WKW-001 is a characterization invariant).

## §B Known issues carried into this SPEC

- The template `workflow.yaml` comment is doubly stale: it calls BOTH
  auto_merge and auto_cleanup "declared but not read", while auto_cleanup
  has had two readers since SPEC-SESSION-WORKTREE-001. Fixed in M4 because
  this card edits exactly that comment block.
- The inventory row for `auto_cleanup` says class P (prose) although
  production readers exist (session_worktree.go:630,
  session_worktree_prmerge.go:150). Out of card scope to re-classify — the
  reader-classification test (`shipped_key_reader_test.go`) classifies
  mechanically from code and is the live arbiter; this card touches only
  the `auto_merge` row and the classification expectations that name it.
- `internal/github/pr_merger.go` AutoMerge fields are the PR-merger CLI
  option (`--auto-merge`) — UNRELATED. Do not "unify" them.

## §C Pre-flight (measured baseline, this tree, 2026-09-12)

| # | Fact | Anchor |
|---|---|---|
| P1 | auto_merge: zero production readers; comment says so | internal/config/types.go:629 |
| P2 | auto_create: read once, advisory wording only, does not gate creation | internal/cli/worktree_advisory.go:16-17,51-61; types.go:624-625 |
| P3 | auto_cleanup: two live consumers (behavioral precedent) | internal/cli/session_worktree.go:630; session_worktree_prmerge.go:150 |
| P4 | Lock Go API importable; force flag exists; stale = reclaimable by human | internal/cli/integration.go:284,334,361; kanban lock package |
| P5 | Integration target resolution already reads develop_branch, git-flow+manual gated, "" on every failure | internal/config/loader_integration_branch.go:34-83; integration.go:121-136 |
| P6 | Template ships no develop_branch; workflow.yaml carries the stale comment | internal/template/templates/.moai/config/sections/{git-strategy.yaml.tmpl, workflow.yaml:49-60} |
| P7 | Inventory: auto_merge = D with deprecate_after v3.1.0; auto_create = W; session_name_pattern/tmux_preferred = D (untouched) | internal/config/testdata/shipped_key_inventory.yaml:2949-2961 |
| P8 | Reader-classification collision assertions exist for both keys | internal/config/shipped_key_reader_test.go:801-814 |
| P9 | Session-worktree activation gate has its own key + env override | internal/config/session_worktree.go:31-44; session_worktree.go:158 |
| P10 | This repo's local .moai workflow.yaml sets auto_merge: true (currently a no-op) — the dogfood path that goes live with this card | .moai/config/sections/workflow.yaml:151 |

## §D Constraints

- [HARD] NEVER push — the auto-merge path executes no remote-mutating git
  command (REQ-WKW-005). Enforced by seam-log assertion (AC-WKW-006).
- [HARD] No `--force` on any lock call; auto path never displaces any hold,
  live or stale (design.md §1.2).
- [HARD] Template neutrality: `develop_branch` is NOT added to
  `git-strategy.yaml.tmpl` (nothing to add — reuse). The `workflow.yaml`
  comment edit is template-managed and ships to all 16 programming-language
  distributions — keep it programming-language and locale neutral (no SPEC
  IDs beyond the existing convention, no internal dates).
- [HARD] `make build` after the template edit (//go:embed), before tests
  that assert template content.
- Default OFF everywhere: `defaults.go` stays `AutoMerge: false`; the local
  dev repo's `auto_merge: true` (P10) is the operator's own opt-in.
- Match the session_worktree family style: function-variable seams, English
  comments, notice-prefix constants, non-blocking failure paths,
  cross-platform `filepath` semantics (GOOS=windows build check in M5).
- Max 3 retries per operation; after that, blocker report.

## §E Self-verification

E1. AC matrix PASS/FAIL from acceptance.md §B, each with its judging command
    output.
E2. Cross-platform build: `GOOS=windows GOARCH=amd64 go build ./...` and
    darwin build.
E3. Coverage of the new merge path ≥ package target (85%); characterization
    test for the OFF baseline recorded.
E4. Seam-log zero-push assertion green (AC-WKW-006).
E5. `golangci-lint run` clean on touched packages.
E6. `make build` + `go test ./internal/config/... ./internal/cli/...`
    (affected packages; NOT the full suite locally — lane discipline).
E7. Evidence paths recorded in progress.md §E.2.

## §F Milestones

Ordered by decision reversibility: behavior first (most likely to change
under review), mechanical honesty sweep last.

### M1 (High) — Auto-merge engine + window ceremony

New file in the session_worktree family implementing the core: trigger
predicate evaluation, ahead-of check, target resolution via
`config.LoadGitFlowIntegrationConfig` + `worktreeForBranch`, window
acquire(`force=false`) → guards (source dirty, target dirty/absent) →
`git merge --no-ff` with `cmd.Dir` at the target tree → conflict abort →
deferred release → notice family with a distinct prefix constant.
Function-variable seams for every git invocation and both lock calls.
Tests: table-driven, seam-injected — OFF baseline byte-identical
(REQ-WKW-001), each guard's skip+notice, busy-window skip (live AND stale),
conflict abort leaves no MERGE_HEAD, no-op silent skip (REQ-WKW-010),
zero-push seam-log assertion.

### M2 (High) — Session-exit wiring

Hook the engine into the session-exit disposal surface BEFORE
`cleanupSessionWorktree`, gated on clean exit only (REQ-WKW-002) and
independent of `auto_cleanup` (REQ-WKW-013 — all four toggle combinations
tested). Tests: exit-code matrix × toggle matrix; ordering (merge observed
before any removal seam call); the merge fires exactly once per exit.

### M3 (Medium) — auto_create scope truth-fix

Reword the `true` branch of `emitWorktreeAdvisory` to claim nothing false:
keep the AC-WBG-009 regex satisfaction, state the preference + point at
`workflow.session_worktree.enabled` for real automatic isolation. Tests:
regex positive (observability kept) + negative (no auto-creation claim);
both wordings; config-load failure degrades to the `false` wording
(existing behavior preserved).

### M4 (Medium) — Honesty sweep (mechanical)

- `internal/config/types.go`: WorkflowWorktreeConfig reader-status comment —
  AutoMerge sentence becomes the reader description; AutoCreate sentence
  states the declared advisory scope.
- `internal/config/defaults.go`: update the worktree block comment (the
  true→false mutation note gains the reader fact).
- `internal/template/templates/.moai/config/sections/workflow.yaml`: comment
  rewritten — auto_merge read (by the session-exit merge path), auto_cleanup
  read (two consumers — fixes the stale half), auto_create scope declared.
  Then `make build`.
- `internal/config/testdata/shipped_key_inventory.yaml`: auto_merge row
  D→W, evidence names the reader, `deprecate_after` removed.
- `internal/config/shipped_key_reader_test.go`: the auto_merge collision
  assertion updated to expect direct-live classification via the new
  reader; auto_create assertion unchanged.
- Types comment / inventory / template must agree with each other (one
  grep-able reader story).

### M5 (Low) — Verification and evidence

§E batch executed and recorded: affected-package tests, lint, cross-build,
`make build`, coverage, seam-log push assertion. Evidence under
`.moai/state/verify/` per the evidence-persistence contract; progress.md
§E.2 populated by run phase.

## §G Anti-patterns

- Do NOT add a fetch/absorb step "for correctness" — the deviation is
  deliberate and recorded (design.md §1.3).
- Do NOT make the lock force flag a parameter "for flexibility" — it is a
  literal false on this path.
- Do NOT sweep WT-* branches at the on-touch surface — merging other
  sessions' in-flight work is the prohibited failure shape.
- Do NOT "unify" with `internal/github` MergeOptions.AutoMerge — different
  key domain (P8 collision assertion exists precisely because of this).
- Do NOT touch session_name_pattern / tmux_preferred / WorktreeRoot.
- Do NOT mirror develop_branch into the template.
- Do NOT run the full test suite locally (lane discipline — affected
  packages only; CI owns the full suite).

## §H Cross-references

- SPEC-SESSION-WORKTREE-001 — the M4/M8 surfaces, REQ-SW-004/008/009/010,
  notice-prefix discipline this SPEC inherits.
- SPEC-CONFIG-KEY-HONESTY-001 — the reader-status comment and inventory
  contract this card updates.
- SPEC-WORKTREE-BASEREF-001 — the single-key loader pattern
  (`LoadWorktreeBaseBranch`) that `LoadGitFlowIntegrationConfig` follows.
- CLAUDE.local.md §4.1 — git-flow integration chain; the lead batch-push
  discipline that owns pushing.
- kanban-dispatch.md § Integration into the release branch is self-served —
  the window doctrine the auto path participates in.
