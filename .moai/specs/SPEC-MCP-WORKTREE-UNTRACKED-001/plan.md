# SPEC-MCP-WORKTREE-UNTRACKED-001 — Plan

Card: t1202 | Issue: #1716 | Branch: WT-worktree-moai-root | Base: origin/develop `df526c9a9` | Tier: M

## §A Context

The `project_root` validator rejects a linked worktree of a repository that
keeps `.moai/` untracked, and omitting the input silently acts on the primary
checkout. Premise measured LIVE (`.moai/reports/t1202/verdict.md`). After
plan-audit iter-2 (FAIL 0.78) the lead chose scope reduction: this SPEC keeps
validator acceptance plus tree operations; configuration/catalogue routing and
state writes move to card t1213 (spec.md §6).

## §B Known Issues and measurement record

### B.1 Evidence (base `df526c9a9`)

- Validator today: `internal/cli/mcp_project_root.go` accepts on `.moai`
  directory existence only; fixtures relying on that include `newProbeProject`
  (`.moai/specs/<id>` only) in `mcp_project_root_test.go` and bare `.moai`
  directories in `mcp_glm_test.go`, `mcp_build_identity_test.go`,
  `mcp_audit_write_capability_test.go`, `mcp_shortest_path_test.go` — the
  REQ-MWU-001 branch must keep all of them green (AC-MWU-013/014).
- Tree operations: `mcp_codex.go` (`params["cwd"] = root`),
  `collectReviewDiff` in `mcp_review_material.go`, `mcp_glm.go`,
  `mcp_audit_multi.go` / `mcp_convergence.go`, `auditBuildIdentity` in
  `mcp_build_identity.go`, and `internal/graph/codequery.go`
  `edgesArtifactPath`. All take the resolved root, so REQ-MWU-008 needs no
  per-handler change beyond the validator returning the worktree path; M4
  verifies this rather than assuming it.
- Codex test seam: `codexLookPath` and `codexRunner` in `mcp_codex.go`.
- Creation paths (for spec.md §2.2): `grep -rnE '"worktree", *"add"' internal cmd`
  → `internal/core/git/worktree.go`, `internal/cli/session_worktree.go`;
  `grep -rn 'materializeSessionWorktree\b' internal` → `session_worktree.go:177`,
  `root.go:137`, `factory_lane_handoff.go:130`,
  `factory_lane_handoff_recover.go:84`; `WorktreeManager.Add` callers →
  `worktree_branch_flag.go:162`, `internal/hook/worktree_create.go:134`.
- Issue #1716 body (read by the plan auditor): reproduction step 1 is
  `moai worktree new …`; expected behavior lists creation-time provisioning as
  acceptable.

### B.2 Git helper constraint (REQ-MWU-005/006)

`internal/core/git/checkout.go` `ResolveGitDirs` falls back to a second
invocation on an unexpected output shape, and neither it nor `codexAuditGit`
(`codex_audit_launch.go`) removes `GIT_*` from the child environment. Do not
call either as-is. Any helper used must run git with `GIT_DIR`,
`GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, and
`GIT_CEILING_DIRECTORIES` removed, and must reject an unexpected output shape
instead of falling back.

### B.3 Tier

Expected files: `mcp_project_root.go` (or a sibling file for the git helper),
one or two test files, the two rule-doc copies — 4–6 files, roughly 250–400 LOC
with tests. Tier M stays (upper end of S by LOC, but five-plus files and a
security-relevant boundary). Recount after M1.

## §C Operator decisions (Kickoff)

Resolved inside the reduced scope (no operator input needed):

- Registration requirement: an unlisted or prunable worktree is rejected
  (REQ-MWU-002/003, AC-MWU-006).
- Graph tools on a worktree without its own artifact: existing "graph layer
  absent" error, no auto-build (REQ-MWU-008, AC-MWU-012).

Still for the operator:

1. **Interim behavior until t1213 lands** (spec.md §6, first Out of Scope
   block): catalogue tools on an accepted worktree read its empty catalogue,
   configuration reads see defaults (a primary-declared `required` codex gate is
   not seen), and a state write creates `.moai` under the worktree. Recommended
   default: **accept the interim and schedule t1213 next**. Alternative: pull a
   `_root` warning for linked-worktree acceptance into this SPEC — adds one
   requirement (REQ-MWU-011) and one criterion (AC-MWU-015).
2. **Design (b)** — issue #1716 lists it as acceptable. Recommended default:
   **not built; recorded as a rejected alternative in spec.md §2**.
   Alternative: file a backlog card for (b). Changes nothing in this SPEC.
3. **Worktree base branch** — `moai worktree new` created this card's tree from
   `main` rather than `develop` (same class as t1159). Recommended default:
   **separate card**. Changes nothing in this SPEC.

## §D Constraints

- Reject-never-fall-back stays binding (spec.md §5).
- Scoped verification only: `go test ./internal/cli/ -run '<new tests>'`, the
  existing `project_root`, graph, and codex tests in `./internal/cli/`, and
  `./internal/template/...` for the mirror and neutrality guards. No local
  full-suite run.
- Tests that set `GIT_*` via `t.Setenv` are non-parallel.
- Template-first for the rule file: edit the template mirror and the local copy
  together, then `make build`.

## §E Self-Verification (run-phase exit)

- E1: AC-MWU-001..014 matrix with verbatim command output.
- E2: RED-first predicates of AC-MWU-001 with their outputs.
- E3: `diff` of the two rule copies exits 0.
- E4: `go vet ./internal/cli/` and `golangci-lint run ./internal/cli/...` clean.

## §F Milestones (ordered by decision-reversibility)

### M1 — Priority High — Fix the branch structure

- Keep the existing `.moai` test as the first branch, untouched. Add the
  linked-worktree branch behind it. Decide where the scrubbed git helper lives.

### M2 — Priority High — RED reproduction, committed alone

- Add the AC-MWU-001 test using only `validateProjectRoot(W)` (asserting
  `err == nil` and the canonical `W`), so it compiles on the pre-fix tree and
  fails with the rejection error. The commit touches `_test.go` files only.

### M3 — Priority High — Validator branch

- REQ-MWU-002..007: scrubbed git helper, primary-checkout predicate,
  registration check on the matching entry only, fail-closed.

### M4 — Priority Medium — Confirm tree operations

- Verify each §3 access receives the worktree path (AC-MWU-002, AC-MWU-012);
  change a handler only where it does not.

### M5 — Priority Medium — Descriptions and rule doc

- Update `projectRootDescCommon`; update § The `project_root` input in both
  copies of `moai-mcp-tools.md` (linked-worktree row, and a sentence that
  configuration, catalogue, and state are still read from the accepted tree),
  neutral wording; `make build`.

### M6 — Priority Low — Verification batch

- Run E1–E4 in one turn; record evidence in progress.md §E.2.

## §G Anti-Patterns

- Changing the existing `.moai` branch (regresses today's accepted set).
- Accepting any directory inside the repository — only a listed worktree top level.
- Rejecting a valid worktree because some other listed entry is stale.
- Falling back to the primary when a git inspection errors.
- Answering a graph query from the primary checkout's graph.
- Calling an unscrubbed git helper.

## §H Cross-References

- `internal/cli/mcp_project_root.go` — validator and resolvers
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input (+ template mirror)
- SPEC-MCP-WORKTREE-ROOT-001 — origin of the `project_root` input
- Card t1213 — deferred configuration/catalogue/state routing
- `.moai/reports/t1202/verdict.md`, `plan-audit.md`, `plan-audit-iter2.md`
