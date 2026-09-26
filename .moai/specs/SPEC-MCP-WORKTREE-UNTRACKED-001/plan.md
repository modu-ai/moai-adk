# SPEC-MCP-WORKTREE-UNTRACKED-001 — Plan

Card: t1202 | Issue: #1716 | Branch: WT-worktree-moai-root | Base: origin/develop `df526c9a9` | Tier: M

## §A Context

The `project_root` validator rejects a linked worktree of a repository that
keeps `.moai/` untracked, and omitting the input silently acts on the primary
checkout. Premise measured LIVE (`.moai/reports/t1202/verdict.md`). Design (a) —
validator-side acceptance plus per-class `.moai` routing — is recommended in
`spec.md` §2.3. plan-audit iter-1 (FAIL 0.62) findings D1–D16 are addressed in
spec v0.2.0; the defect→change map is in progress.md §E.1.

## §B Known Issues and measurement record

### B.1 Inventory evidence (spec.md §3)

Measured on base `df526c9a9` by reading the handlers and their callees:

- `grep -nE '"\.moai"|\.moai/|projectDirResolver\(|resolveProjectDir\('` over
  `internal/cli/mcp_{codex,glm,claude,audit_multi,convergence,server,code_tools,build_identity}.go`
  and `codex_audit_launch.go`.
- `internal/graph/codequery.go` `edgesArtifactPath` → `<root>/.moai/project/graph/edges.jsonl`.
- `internal/verify/store.go` `SnapshotDir = ".moai/state/verify/snapshots"`.
- `internal/spec/listdocs.go` and `internal/spec/audit.go` → `<base>/.moai/specs`.
- `internal/auditreceipt/store.go` → `.moai/state/audit-receipts`, and
  `CodexGateRequired` reads `<tree>/.moai/config/sections/workflow.yaml`.
- `internal/cli/mcp_convergence.go` `convergenceStateDirFor(root)` →
  `<root>/.moai/state/audit-multi/`.
- `internal/hook/audit_receipt_guard.go` → reader tree = `auditreceipt.TreeRootFromCWD`.
- `grep -rnE '"worktree", *"add"' internal cmd` (non-test) → 2 execution sites
  (`internal/core/git/worktree.go`, `internal/cli/session_worktree.go`); callers
  of `WorktreeManager.Add`: `internal/hook/worktree_create.go`,
  `internal/cli/worktree_branch_flag.go`; `internal/cli/worktree/new.go`
  delegates to the session materializer.
- Issue #1716 body (read by the plan auditor, 2026-09-26): reproduction step 1 is
  `moai worktree new …`; expected behavior lists creation-time provisioning as
  acceptable.

Gaps: `spec.Audit` and `verify` bodies were read at their join sites only; the
inventory is a lower bound. Re-verify at M1 and record any addition.

### B.2 Git helper constraint (REQ-MWU-005/006)

- `internal/core/git/checkout.go` `ResolveGitDirs` falls back to a second
  invocation on an unexpected output shape, and neither it nor
  `codexAuditGit` (`codex_audit_launch.go`) removes `GIT_*` from the child
  environment. **Do not call either as-is.** Any helper used must (1) run git
  with `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, and
  `GIT_CEILING_DIRECTORIES` removed from the child environment, and (2) reject
  an unexpected output shape instead of falling back. Wrapping an existing
  helper is acceptable only if both properties hold.

### B.3 Tier

Expected files: `mcp_project_root.go`, `mcp_server.go`, `mcp_codex.go`,
`mcp_claude.go`, `mcp_glm.go`, `mcp_convergence.go`, `mcp_audit_receipt.go`
(or `internal/auditreceipt`), two or three test files, two rule-doc copies —
about 12–14, inside Tier M. Recount after M1; tier up to L if the count passes 15.

## §C Operator decisions (Kickoff)

The requirements in spec.md §4 are written under the recommended default of
each decision. These are not resolved here; the orchestrator asks them at
Implementation Kickoff.

1. [NEEDS CLARIFICATION: configuration and catalogue root] Class-C reads
   (config sections, SPEC catalogue) go to the primary's `.moai` when the
   worktree has no MoAI configuration. **Recommended default: yes** (matches a
   tracked-`.moai` repository, keeps one configuration). Alternative: reject
   class-C tools for such worktrees and accept only class-T/G/S. Changes:
   REQ-MWU-008, REQ-MWU-012, AC-MWU-010, AC-MWU-011.
2. [NEEDS CLARIFICATION: state-write destination] Class-S writes
   (`verify_snapshot` record, audit receipts, `audit_multi` state) go to the
   worktree's `.moai/state`. **Recommended default: worktree** — their readers
   key on the worktree (receipt guard, convergence gate). Alternative: primary's
   `.moai/state`, which would require changing those readers. Changes:
   REQ-MWU-010, AC-MWU-013.
3. [NEEDS CLARIFICATION: registration requirement] A worktree not listed by
   `git worktree list --porcelain`, or marked prunable, is rejected.
   **Recommended default: reject.** Alternative: accept on common-dir match
   alone. Changes: REQ-MWU-002, REQ-MWU-003, AC-MWU-006.
4. Graph tools on a worktree without its own graph artifact return the existing
   "graph layer absent" error. **Recommended default: explicit error, no
   auto-build.** Alternative: build per worktree on demand. Changes: REQ-MWU-009,
   AC-MWU-012.
5. A worktree-local untracked `.moai/specs` is not merged into the catalogue
   read from the source root. **Recommended default: not merged.** Alternative:
   union of both catalogues. Changes: REQ-MWU-008, spec.md §6.
6. Design (b) — issue #1716 lists it as acceptable. **Recommended default: not
   built; record as a rejected alternative in spec.md §2.** Alternative: file a
   backlog card for (b). Changes: none in this SPEC.
7. `moai worktree new` created this card's tree from `main` rather than
   `develop` (same class as t1159). **Recommended default: separate card.**
   Changes: none in this SPEC.

## §D Constraints

- Reject-never-fall-back stays binding (spec.md §5).
- Scoped verification only: `go test ./internal/cli/ -run '<new tests>'`, the
  existing `mcp_project_root` tests, `./internal/auditreceipt/...` if touched,
  `./internal/template/...` for the mirror and neutrality guards. No local
  full-suite run.
- Tests that set `GIT_*` via `t.Setenv` are non-parallel.
- Template-first for the rule file: edit the template mirror and the local copy
  together, then `make build`.

## §E Self-Verification (run-phase exit)

- E1: AC-MWU-001..015 matrix with verbatim command output.
- E2: RED-first predicates of AC-MWU-001 with their outputs.
- E3: `diff` of the two rule copies exits 0.
- E4: `go vet ./internal/cli/` and `golangci-lint run ./internal/cli/...` clean.

## §F Milestones (ordered by decision-reversibility)

### M1 — Priority High — Fix the resolver result shape; re-verify the inventory

- Settle the result shape: tree root, source root, acceptance route, and the
  provenance fields (REQ-MWU-012). Most likely to change; settle before code.
- Re-verify spec.md §3 against the tree; record additions in progress.md §E.2
  and route each to its class.

### M2 — Priority High — RED reproduction, committed alone

- Add the AC-MWU-001 test (and AC-MWU-002) touching `_test.go` files only;
  observe failure on the unchanged validator; commit before any fix.

### M3 — Priority High — Validator

- REQ-MWU-002..006 on the no-configuration path: scrubbed git helper,
  primary-checkout predicate, registration check, fail-closed.

### M4 — Priority Medium — Per-class routing

- Route class-C reads to the source root; keep class-T, G, S on the tree root;
  extend `_root` provenance.

### M5 — Priority Medium — Descriptions and rule doc

- Update `projectRootDescCommon`; update § The `project_root` input in both
  copies of `moai-mcp-tools.md` (situation-table row + per-class paragraph),
  neutral wording; `make build`.

### M6 — Priority Low — Verification batch

- Run E1–E4 in one turn; record evidence in progress.md §E.2.

## §G Anti-Patterns

- Accepting any directory inside the repository — only a listed worktree top level.
- Falling back to the primary when a git inspection errors.
- Using "`.moai` directory exists" as the source-root predicate (flips on the
  first state write).
- Answering a graph query from the primary checkout's graph.
- Calling an unscrubbed git helper.

## §H Cross-References

- `internal/cli/mcp_project_root.go` — validator and resolvers
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input (+ template mirror)
- SPEC-MCP-WORKTREE-ROOT-001 — origin of the `project_root` input
- `.moai/reports/t1202/verdict.md` — premise measurement
- `.moai/reports/t1202/plan-audit.md` — iter-1 audit
