# SPEC-TPL-AST-GUARD-001 — Research (operator-directed Tier-M addition)

All evidence measured in this session against worktree `WT-ast-template-guard` @ `7c7c84b5c` (2026-10-03). Paths relative to repo root.

## 1. Reader-map evidence (the map the expectation table encodes)

Command: `grep -rn "Worktree\.AutoCreate\|Worktree\.AutoMerge\|Worktree\.AutoCleanup\|Worktree\.TmuxPreferred\|Worktree\.SessionNamePattern" internal cmd --include="*.go" | grep -v _test.go` — verbatim production hits (exactly four):

```
internal/cli/session_worktree.go:656:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
internal/cli/session_worktree_prmerge.go:150:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
internal/cli/session_worktree_automerge.go:162:	if cfg == nil || !cfg.Workflow.Worktree.AutoMerge {
internal/cli/worktree_advisory.go:66:	return cfg.Workflow.Worktree.AutoCreate
```

Enclosing functions read directly: `cleanupSessionWorktree` (session_worktree.go), `prMergeCleanup` (session_worktree_prmerge.go), `sessionExitAutoMerge` (session_worktree_automerge.go), `readWorktreeAutoCreate` (worktree_advisory.go).

- `TmuxPreferred`: zero production reads — only `internal/config` `_test.go` files and the struct/defaults declarations (`types.go:696`, `defaults.go:1193`).
- `SessionNamePattern`: zero reads of any kind outside the declaration.
- Alias-copy probe `grep -rnE "(:=|=)\s+.*\.Workflow\.Worktree\b" ... | grep -v _test.go`: zero true assignments (the three hits were the `if cfg == nil ||` lines matching `=` inside `==`) — the alias defect is latent, not live, on this tree.
- Cross-check, struct doc comment `internal/config/types.go:676-690` (SPEC-CONFIG-KEY-HONESTY-001 M5 / SPEC-WORKTREE-KEY-WIRING-001 REQ-WKW-012): names the same four reader sites and declares `SessionNamePattern` readerless. It does not mention `TmuxPreferred` — consistent with zero readers.
- Template block verified at `internal/template/templates/.moai/config/sections/workflow.yaml` (~lines 54-69): prose clauses for `auto_create` (names worktree_advisory.go), `auto_cleanup`, `auto_merge`; `session_name_pattern` marked reserved; `tmux_preferred: true` shipped active with no clause; `sparse_paths` a commented-out example with **no struct field** behind it.
- PR #1707: `gh pr view 1707 --json state,title` → `{"state":"CLOSED","title":"fix(template): say that auto_cleanup gates worktree removal"}`. Its test file absent on develop and on HEAD (`git cat-file -e` → absent, both). The PR's **final** diff is AST-based (plan-audit iteration 1, `gh pr diff 1707`: `ast.Inspect` ×3, `TestFieldReadersIn_FollowsCopiesAndIgnoresOtherStructs`); the review's alias Major was fixed inside the PR at d102e1e — the alias-blind text scan never shipped in the PR's final state and the PR never landed on develop.

## 2. The two existing guards (what to keep, what supersedes what)

- `internal/config/workflow_key_honesty_test.go` (171 lines, t682/GH #1705): text-based, `strings.Contains(src, ".Workflow.Worktree.AutoCleanup")`; one-directional (template phrase vs reader existence, `readers["auto_cleanup"] >= 1`); carries a negative-control test (`TestWorkflowWorktreeKeyDocGuardDetectsHistoricalDrift`) — a convention worth mirroring in assertion style. **Retained unchanged**: under contract A the new guard does not parse template comments, so the t682 template-phrase check remains the only guard on phrasing drift.
- `internal/config/shipped_key_reader_test.go` (895 lines, SPEC-CONFIG-KEY-HONESTY-001 REQ-CKH-008): the in-repo AST scanner precedent. `packages.Config{Mode: NeedName|NeedTypes|NeedSyntax|NeedTypesInfo|NeedFiles, Dir: findRepoRoot}`, `packages.Load(cfg, "./...")`, `ast.Inspect` over `*ast.SelectorExpr`, `pkg.TypesInfo.Selections[sel]`, `_test.go` skipped, `types.go` accessor reads excluded. Different contract (key liveness vs reader-set honesty) — untouched.

## 3. Decision D1 — Contract A (recommended) vs B

| axis | A: template-as-is + test-local expectation table | B: template clauses name files, guard parses them |
|---|---|---|
| template churn | none (wording landed on develop) | re-edits shipped template → `make build` regen + wording review |
| enforcement power | set equality + live-struct completeness + reserved-key rules | identical checks, but the expectation source is prose parsing |
| parser fragility | none (Go map literal) | comment-format parser becomes part of the contract surface |
| convention fit | matches t682's `workflowKeyClaims`/`workflowKeyAccessors` maps and the audit-test family | no in-repo precedent of guards parsing template comment clauses |
| staleness | table can drift — mitigated structurally: REQ-006 (live-struct completeness) + REQ-003 (set equality) make drift a loud one-line fix | self-documenting template; drift still requires a template edit |
| doc SSOT | table mirrors the types.go reader-status doc comment (update both together — plan.md §G) | template comment becomes a second copy of the same map |

Recommendation: **A**. The deciding argument is not effort — it is that B moves the contract into a prose parser for zero added enforcement, and re-opens already-shipped template wording the card explicitly declines to re-edit. A's staleness risk is bounded by two structural requirements, and the table sits next to a doc-comment SSOT that already carries the same map with file names.

## 4. Decision D2 — Scope (a) minimal (recommended) vs (b) helper extraction vs (c) house migration

Simplicity ladder (constitution § Core Behaviors 4): (1) needed at all — yes, the card mandates the guard; (2) existing helper to reuse — the precedent's scanner is test-local in a different package; no shared helper exists, and extracting one from `shipped_key_reader_test.go` would touch an out-of-scope guard; (3) smallest thing that works — scanner as unexported funcs in the one `_test.go`. No second consumer exists for a shared AST-readers helper; the precedent itself keeps an 895-line scanner test-local. Extraction later is mechanical. (c) is excluded per card + spec.md §C Exclusions.

## 5. go/packages vs stdlib go/parser + go/importer

- **Decision: `golang.org/x/tools/go/packages`.** Zero new dependencies — x/tools v0.49.0 is already a direct require (`go.mod:32`).
- Why not stdlib-only: `go/parser` alone repeats the text-scan's failure mode (syntax without binding); adding correct `go/types` resolution via `go/importer`'s source importer means re-implementing module-context package loading (module root discovery, export-data vs source fallback, cross-package `internal/` imports) that `packages.Load` provides and that the precedent already exercises at 24.0s measured (`go test ./internal/config/ -run '^TestShippedConfigKeysHaveReaders$' -count=1` → `ok ... 23.997s`, this session, warm cache).
- Alias-proofness mechanism: with `NeedTypesInfo`, every field selection's identifier resolves through `pkg.TypesInfo` (`Uses[sel.Sel]`, or `Selections[sel]` kind FieldVal as in the precedent) to the same `*types.Var` field object regardless of receiver shape — local alias copies, method-internal reads, and multi-line expressions included. This is the property the text scan lacks; it is a checker guarantee, not a pattern heuristic.
- Known limitation (documented, acceptable): reads that reach the field through untyped paths — `reflect`, `interface{}` round-trips — do not resolve to the field object. None exist for this struct on the tree (grep-verified); REQ-007's type-error failure prevents a degraded index from reading as a pass.

## 6. PR #1707 review findings → requirement mapping

| review finding | where it lands |
|---|---|
| MAJOR: alias-blind reader scan (raised on the PR's initial text scan; fixed inside the PR at d102e1e — the final diff is AST-based; the retained alias-blind exposure on develop is the t682 text guard) | REQ-001/002 (type-resolved attribution), AC-005 (characterization fixture) |
| both named reader sites required for auto_cleanup (be8b2dc) | REQ-004, AC-004 (two mutation arms) |
| Minor: empty-readers branch must also require named files empty | REQ-005, AC-007b (table-invalid when a reserved key names files) |
| bidirectional file comparison | REQ-003, AC-002/003 |

## 7. Gaps (explicitly unobserved this session)

- Provenance correction (plan-audit iteration 1, audit D2): the card's verified-context item 5 described the PR's review-time scan as text-based; the PR's final diff is AST-based (fixed at d102e1e before close). spec.md §A.2/§A.3 and this file were corrected in 0.1.1 accordingly. The PR body was not read end-to-end (the diff and the deciding review comments were, per the audit verdict).
- The 24.0s runtime figure is from the precedent guard in `internal/config`, not from the new guard (which does not exist at plan phase); the new guard's actual cost is an M1 observation.
- Reader-map greps cover `internal/` and `cmd/` non-test `.go` in this module; a reader living outside those roots (none plausible — config consumers are in-module) would only be caught by the guard's own `./...` scan at run time.
