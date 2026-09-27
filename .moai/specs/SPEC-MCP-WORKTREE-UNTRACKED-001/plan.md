# SPEC-MCP-WORKTREE-UNTRACKED-001 — Plan

Card: t1202 | Issue: #1716 | Branch: WT-worktree-moai-root | Base: origin/develop `df526c9a9` | Tier: M

## §A Context

The `project_root` validator rejects a linked worktree of a repository that
keeps `.moai/` untracked, and omitting the input silently acts on the primary
checkout. Premise measured LIVE (`.moai/reports/t1202/verdict.md`). After
plan-audit iter-2 (FAIL 0.78) the lead chose scope reduction: this SPEC keeps
validator acceptance plus tree operations; configuration/catalogue routing and
state writes move to card t1213 (spec.md §6). After the delta audit (FAIL 0.86,
D27) this SPEC also enforces the primary's audit gate on a config-orphaned
worktree, fail-closed, and — lead decision D30 = B — warns on catalogue/state
answers read from such a worktree (spec.md §4.4).

## §B Known Issues and measurement record

### B.1 Evidence (base `df526c9a9`)

- Validator today: `internal/cli/mcp_project_root.go` accepts on `.moai`
  directory existence only; fixtures relying on that include `newProbeProject`
  (`.moai/specs/<id>` only) in `mcp_project_root_test.go` and bare `.moai`
  directories in `mcp_glm_test.go`, `mcp_build_identity_test.go`,
  `mcp_audit_write_capability_test.go`, `mcp_shortest_path_test.go` — the
  REQ-MWU-001 branch must keep all of them green (AC-MWU-012/013).
- Audit-gate readers (spec.md §3.1): `applyGateUnmet` (`mcp_codex.go`),
  `workflowAuditGates` → `enforceRequiredGateUnmet` (`mcp_convergence.go`),
  `recordAuditReceipt` → `auditreceipt.CodexGateRequired`
  (`mcp_audit_receipt.go`, `internal/auditreceipt/store.go`); all read the root's
  raw `workflow.yaml` through `loadWorkflowAuditSection` (`audit_pin.go`), which
  returns zero gates when the file is absent. Receipt writes create
  `<root>/.moai/state/audit-receipts/` (`auditreceipt` `StateDir`), which is why
  every new condition keys on the config-orphaned property rather than the
  acceptance branch (plan-audit delta D27/D29).
- `_root` provenance: `rootProvenanceMap` (`mcp_project_root.go`) is attached by
  `spec_progress`, `spec_drift`, `verify_snapshot`, `verify_trend`; `spec_audit`
  returns the audit result without it (REQ-MWU-013 adds it).
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

Expected files: `mcp_project_root.go` (plus a sibling for the scrubbed git
helper and config-orphan predicate), `mcp_codex.go`, `mcp_convergence.go`,
`mcp_audit_receipt.go` (the receipt-id gate read is routed at this MCP call
site; `auditreceipt.CodexGateRequired` is NOT changed, so the hook-side receipt
guard in `internal/hook/audit_receipt_guard.go` keeps today's behaviour and
stays with card t1213), `audit_pin.go` (shared gate seam for the other two
reads),
`mcp_server.go` (warning on catalogue/state responses, `_root` on
`spec_audit`), two or three test files, the two rule-doc copies — about 9–11
files, roughly 400–700 LOC with tests. Tier M. REQs 13, ACs 16 (at the Tier M
ceiling). Recount after M1.

## §C Operator decisions (Kickoff)

Resolved inside the reduced scope (no operator input needed):

- Registration requirement: an unlisted or prunable worktree is rejected
  (REQ-MWU-002/003, AC-MWU-005).
- Graph tools on a worktree without its own artifact: existing "graph layer
  absent" error, no auto-build (REQ-MWU-008, AC-MWU-011).

Decided by the lead (recorded, not open):

- **Catalogue/state warning (delta D30) — lead decision B.** Catalogue and state
  tools on a config-orphaned root carry a `_root` warning so an empty catalogue
  is distinguishable from "no SPECs" (REQ-MWU-013, AC-MWU-016). The former
  default ("defer to t1213, no warning in this SPEC", under which catalogue tools
  returned zero SPECs on the worktree with no signal) is withdrawn.
- **Audit gate (delta D27).** Not an operator option: the primary's
  `workflow.audit.gates` is enforced on a config-orphaned worktree and the read
  fails closed (REQ-MWU-011/012). Routing of the remaining configuration,
  catalogue, and state stays with card t1213.

Still for the operator:

1. *(resolved — see above)*
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

- E1: AC-MWU-001..016 matrix with verbatim command output.
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

- Verify each §3 access receives the worktree path (AC-MWU-002, AC-MWU-011);
  change a handler only where it does not.

### M4b — Priority High — Config-orphan predicate, gate read, warning

- Shared config-orphan predicate (reads `<root>/.git`, then the admin
  directory's `commondir`, resolving a relative `gitdir:` against the directory
  holding the `.git` file; no back-reference check, no subprocess) used
  by the three §3.1 gate reads (REQ-MWU-011/012) and by the catalogue/state
  `_root.worktree_warning` (REQ-MWU-013). Primary identification (scrubbed git,
  `LC_ALL=C`, exit status and output shape only) runs only for config-orphaned
  roots and fails closed there; every other root keeps today's gate path. Only
  `workflow.audit.gates` is routed; no write destination changes; the existing
  `_root.warning` key is untouched. AC-MWU-014, AC-MWU-015, AC-MWU-016, including the second-call case
  after `W/.moai/state/` exists.

### M5 — Priority Medium — Descriptions and rule doc

- Update the `codex_audit` tool description in `internal/cli/mcp_server.go`
  (today: gate read from "the reviewed tree") and any other tool description
  naming the gate source.
- Update `projectRootDescCommon`; update § The `project_root` input in both
  copies of `moai-mcp-tools.md` (linked-worktree row; sentences that the audit gate of a
  worktree without its own workflow config is read from the primary checkout,
  and that other configuration, the catalogue, and state are still read from
  the accepted tree),
  neutral wording; `make build`.

### M6 — Priority Low — Verification batch

- Run E1–E4 in one turn; record evidence in progress.md §E.2.

## §G Anti-Patterns

- Changing the existing `.moai` branch (regresses today's accepted set).
- Accepting any directory inside the repository — only a listed worktree top level.
- Rejecting a valid worktree because some other listed entry is stale.
- Falling back to the primary when a git inspection errors.
- Answering a graph query from the primary checkout's graph.
- Keying the gate read or the warning on "accepted through REQ-MWU-002" — the
  first receipt write flips later calls to the REQ-MWU-001 branch.
- Failing closed on a root without worktree evidence (non-repository, primary
  checkout, subdirectory, submodule) — those keep today's behaviour.
- Classifying "not a repository" from git's (localized) message text.
- Deciding worktree evidence from the `gitdir:` path string alone (a submodule
  at `…/worktrees/<x>` matches); require `commondir`.
- Adding an admin-dir `gitdir` back-reference check: it can only disqualify a
  worktree (e.g. one moved by hand) and so only weakens the gate.
- Resolving a relative `gitdir:` against the process working directory.
- Writing the worktree warning into the existing `_root.warning` key.
- Calling an unscrubbed git helper.

## §H Cross-References

- `internal/cli/mcp_project_root.go` — validator and resolvers
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input (+ template mirror)
- SPEC-MCP-WORKTREE-ROOT-001 — origin of the `project_root` input
- Card t1213 — deferred configuration/catalogue/state routing
- `.moai/reports/t1202/verdict.md`, `plan-audit.md`, `plan-audit-iter2.md`
