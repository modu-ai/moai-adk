# SPEC-TPL-AST-GUARD-001 — Implementation Plan

## §A Context

Card t1377 (operator-directed 2026-10-03, v3.2.0 mission): adopt an AST-based guard mechanism for the `workflow.worktree.*` template key-honesty contract. Evidence basis: PR #1707 (CLOSED, unmerged) and its review thread. Worktree `WT-ast-template-guard` @ `7c7c84b5c` (= local develop tip).

The deliverable is one Go test file in `internal/template` plus a small testdata fixture package. No production code changes. Tier M: spec.md + plan.md + acceptance.md + progress.md, plus the operator-directed research.md (reader-map + approach evidence).

Prior art this plan builds on (all verified this session at baseline `7c7c84b5c`):

- `internal/config/shipped_key_reader_test.go` — the in-repo AST scanner precedent: `packages.Config{Mode: NeedName|NeedTypes|NeedSyntax|NeedTypesInfo|NeedFiles, Dir: repoRoot}`, `packages.Load(cfg, "./...")`, walk `ast.Inspect` for `*ast.SelectorExpr`, resolve via `pkg.TypesInfo`, skip `_test.go`. Measured runtime this session: `ok github.com/modu-ai/moai-adk/internal/config 23.997s`.
- `internal/config/workflow_key_honesty_test.go` — the text-based guard being superseded in role (retained, not modified): its hardcoded expectation-map style (`workflowKeyClaims`, `workflowKeyAccessors`) and its negative-control test (`TestWorkflowWorktreeKeyDocGuardDetectsHistoricalDrift`) are the repo-convention evidence for decision D1.

## §B Known Issues

- **B-1 Alias blindness (the core defect)**: text/line scans of `.Worktree.<Field>` miss struct-copy alias reads (`w := cfg.Workflow.Worktree; w.AutoCleanup`). No such alias read exists on the current tree (grepped: zero non-test assignments of `.Workflow.Worktree` to a local), so the defect is latent — which is precisely why the guard must be type-resolved before the first alias lands.
- **B-2 One-directionality of the t682 guard**: it asserts `readers["auto_cleanup"] >= 1` plus template phrasing; it cannot detect an unnamed additional reader, a dropped expected reader, or a stale reader map.
- **B-3 The PR's test is absent on develop** — nothing on the current tree enforces the bidirectional file contract. The guard this SPEC plans is net-new on this branch.

## §C Pre-flight (run-phase M0, before any code)

1. Re-verify the reader map at the then-current HEAD (the map in spec.md §A.1 was measured at `7c7c84b5c`):
   - `grep -rn "Worktree\.AutoCreate\|Worktree\.AutoMerge\|Worktree\.AutoCleanup\|Worktree\.TmuxPreferred\|Worktree\.SessionNamePattern" internal cmd --include="*.go" | grep -v _test.go`
   - `grep -rnE "(:=|=)\s+\w+(\.)?Workflow\.Worktree\b" internal --include="*.go" | grep -v _test.go` (alias-copy probe)
   - If the map moved, update the expectation table in the same commit that lands the map change (the table is the documented reader map; drift is a maintenance action, not a failure).
2. Confirm clean baseline: `go test ./internal/template/ -count=1` green before adding files.
3. Confirm `git rev-parse --short HEAD` matches the branch the lane expects (staleness rule).

## §D Constraints

- From spec.md §C: zero new dependencies (x/tools already direct); no template edits (→ **`make build` regeneration is NOT required** — the binary embeds templates, but no template file changes under contract A); no production config changes; t682 + shipped-key guards untouched; TDD with mutation-demonstrated RED.
- File budget: 2 new files — `internal/template/workflow_worktree_key_honesty_test.go` (guard + unexported scanner helpers) and `internal/template/testdata/worktreekeyaliasprobe/aliasprobe.go` (fixture). Estimated ~300-450 LOC total (Tier M band).
- The fixture lives under `testdata/`, so the go tool's `./...` wildcard ignores it; the guard loads it by explicit pattern (same mechanism the production scan uses, just a narrower pattern), which keeps it out of normal builds while keeping it compilable in-module (it imports `github.com/modu-ai/moai-adk/internal/config`).

## §E Self-Verification

Every AC in acceptance.md is a runnable command or a mutation-observe-revert cycle. Mutation evidence (AC-002..AC-004, AC-006..AC-008, AC-012): apply the mutation, run the guard, capture the failing output, revert, re-run clean. Evidence lands under `.moai/reports/t1377/` (primary, citation-target) per the card's evidence convention; scratch under `.moai/state/verify/` is never cited. Re-measurement commands run against the owning package (`./internal/template/`), not a narrower `-run` selector, when claiming package-level gates.

## §F Milestones

Decision-reversibility ordering: the contract-shape decisions (table entries, matching predicate) are the highest-change-likelihood items and lead; mechanical gates close the plan.

### §F.0 Decision record (leader may override at plan review; default = recommendation)

- **D1 — Contract: A (template-as-is expectation table) vs B (template clauses name files, guard parses them).** **Recommendation: A.** Reasoning: (1) repo convention — the t682 guard this SPEC succeeds already uses hardcoded package-level expectation maps, and `agent_frontmatter_audit_test.go` + the shipped-key inventory follow the same guard-carries-expectations shape; (2) the template wording already landed on develop — option B re-edits shipped template content, forcing `make build` regeneration plus a wording review, for zero added enforcement power (the enforcement lives in the AST scan either way); (3) staleness is structurally mitigated under A: REQ-006 ties the table to the live struct (a new field without a table entry fails), and set equality (REQ-003) catches both dropped and unnamed readers — table drift reduces to a one-line test edit, which is exactly the documented maintenance action. Full tradeoff in research.md §3.
- **D2 — Scope: (a) minimal single guard vs (b) reusable scanner helper extraction.** **Recommendation: (a).** The scanner stays as unexported helpers inside the one `_test.go` file. No second consumer exists today; the in-repo precedent keeps an 895-line scanner test-local; extraction later is mechanical cut-and-paste. Option (c) (house-wide migration of text guards) is excluded (spec.md §C Exclusions).

### M1 (Priority High) — AST scanner + expectation table + production honesty test

Files: `internal/template/workflow_worktree_key_honesty_test.go`.

1. Scanner (unexported, same file): `packages.Load` with the precedent's Mode and `Dir:` repo root, pattern `./...`; walk selector expressions; resolve the selected object through `pkg.TypesInfo` — **primary route `Selections[sel].Obj()`** (the precedent's route; alias-proof by construction). Note: `Uses[sel.Sel]` is NOT the working route for field selections — a field selection's Sel resolves through `Selections`, so a Uses-keyed index comes back empty (a defect AC-001's executed-test assertion and REQ-003 set equality surface immediately); match when the object is the named field of `internal/config`'s `WorkflowWorktreeConfig`, counting only value-consuming positions (REQ-002's read/write exclusion — a selection in an assignment-target position is a pure write, not a reader); record field → repo-relative file paths; skip `_test.go`.
2. Expectation table (package-level map, contract A): `auto_create → [internal/cli/worktree_advisory.go]`; `auto_merge → [internal/cli/session_worktree_automerge.go]`; `auto_cleanup → [internal/cli/session_worktree.go, internal/cli/session_worktree_prmerge.go]`; `tmux_preferred → []`; `session_name_pattern → []`.
3. Assertions: table completeness against the live struct (REQ-006); set equality per key with named-file findings (REQ-003); **independent table-content assertion for REQ-004** — the `auto_cleanup` expectation entry must contain BOTH `internal/cli/session_worktree.go` and `internal/cli/session_worktree_prmerge.go`, evaluated against the table itself without consulting the computed scan, so a same-commit removal of a read and its table entry still fails (plan §C's endorsed maintenance path must not be able to silence REQ-004); reserved-key rules (REQ-005); type-error failure (REQ-007); test-file exclusion (REQ-008).
4. Gate: AC-001 (clean-tree pass).

### M2 (Priority High) — Alias characterization fixture

Files: `internal/template/testdata/worktreekeyaliasprobe/aliasprobe.go` + `internal/template/testdata/worktreekeyaliasprobe/writeonly.go`.

Fixture package importing `internal/config` with two files: `aliasprobe.go` reads `AutoCleanup` through a local alias copy (`w := cfg.Workflow.Worktree; _ = w.AutoCleanup`); `writeonly.go` touches the same field ONLY as a pure write (`w.AutoCleanup = false`). The guard's fixture-mode test loads the package by explicit pattern with the SAME matcher and asserts: detection of `aliasprobe.go` (AC-005a), non-attribution of `writeonly.go` (AC-005b — REQ-002's write exclusion), and the negative characterization that the legacy text accessor string matches zero times in `aliasprobe.go` (AC-005c) — the retained t682 text scan's blind spot, recorded as evidence.

### M3 (Priority High) — Mutation evidence matrix

Execute AC-002, AC-003, AC-004 (three arms — the third being the combined read+table-entry removal), AC-006, AC-007a, AC-007b, AC-008, AC-012 as mutation-observe-revert cycles; capture verbatim outputs to `.moai/reports/t1377/`. Every mutation is reverted and the clean-tree run re-observed before the next.

### M4 (Priority Medium) — Package-level quality gates

`go test ./internal/template/ -count=1` green (AC-009); `golangci-lint run ./internal/template/...` 0 issues (AC-010); `gofmt -l` on the two new files empty (AC-011). Sync-phase docs per the standard flow.

## §G Anti-Patterns

- Do NOT use `strings.Contains` / line matching anywhere in the new guard's reader computation — that is the defect being removed.
- Do NOT edit the template yaml, the t682 guard, or `shipped_key_reader_test.go` (spec.md §C Exclusions).
- Do NOT extract a shared scanner helper package "while we're here" (D2; mechanical later if a second consumer appears).
- Do NOT gate the fixture by build tags or move it out of `testdata/` — it must stay invisible to `./...` yet explicitly loadable.
- Do NOT record mutation evidence from memory: each AC's evidence is the verbatim command output captured in this run (verification-claim-integrity §1/§2).
- Do NOT put the reader map's provenance only in commit messages — the table entries and the types.go doc comment are the documentation SSOT; when the table changes, update `internal/config/types.go`'s `WorkflowWorktreeConfig` doc comment in the same commit (maintenance convention, not a lint gate).

## §H Cross-References

- PR #1707 (CLOSED, unmerged) — the bidirectional contract + review findings this SPEC rebuilds on AST.
- GH #1705 / card t682 — the original drift (auto_cleanup called reserved) and the text guard.
- SPEC-CONFIG-KEY-HONESTY-001 — the shipped-key liveness guard and the types.go reader-status doc comment (related).
- SPEC-WORKTREE-KEY-WIRING-001 — REQ-WKW-012, the auto_merge wiring documented in the struct doc (related).
- Card t1377 evidence path: `.moai/reports/t1377/`.
