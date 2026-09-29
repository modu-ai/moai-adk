# SPEC-CODEX-ROLE-AUDIT-ROOT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-29
- plan_audit: iter1 FAIL 0.875 (MP-7 unresolved marker; .moai/reports/t1324/plan-audit-iter1.md) → repairs D1-D6 at 91c522dfe → iter2 PASS 1.0, no regression (.moai/reports/t1324/plan-audit-iter2.md); auditor-model: glm-5.3-flash[1m] both rounds
- Implementation Kickoff Approval: decision_kickoff: approved (2026-09-29) — basis: card t1324 operator directive ("원인을 찾아서 카드 발행해서 해결", operator instruction 2026-09-29) + AGENTS.local.md §31 autonomous-kickoff policy (the card names no operator gate); progression mode: autonomous; the lane reports this decision to the factory leader

## §E.2 Run-phase Evidence

### M0 pre-flight (tree 91c522dfe, binary `/tmp/t1324-run-moai` built from this tree)

- Pre-flight deviation note: plan.md §C's characterization anchor `-run '^(CodexAudit|CodexTask)$'` (exact-name form) swept ZERO tests twice (`[no tests to run]`) — empty-sweep green is not a baseline (verification-completeness §1.1). Corrected prefix anchor `-run '(Test)?(CodexAudit|CodexTask)'`: `ok github.com/modu-ai/moai-adk/internal/cli 32.028s` (green baseline, this tree). Baseline builds: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Lint baseline: `golangci-lint v2.1.6 run ./internal/cli/...` → `0 issues.` (CI version per lane lesson).
- AC-001 RED re-observation, piped-EOF one-shot (server cwd = primary checkout, worktree_root = this worktree): tool result `isError: true`, verbatim text `codex_role_audit: the server did not start inside a git worktree` — the lower-layer artifact plan.md §C predicted; session form switched, not diagnosis.
- AC-001 RED re-observation, LIVE stdio session (stdin held open across tools/call; same wiring): `isError: true`, verbatim text `codex_role_audit: codex audit plan-auditor: working root rejected: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1324 is not the caller's own worktree (/Users/goos/MoAI/moai-adk-go)` — the target equality-clause refusal, matching acceptance.md AC-001 RED-now.
- AC-002 RED re-observation: `/tmp/t1324-run-moai codex role-audit` → stdout/stderr verbatim `unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app`, exit code 1 (clean re-measure; a first piped reading showed `head`'s exit 0 and was discarded).
- AC-003 RED re-observation: `tools/list` properties `codex_task -> ['background', 'prompt', 'resume_last', 'thread_id', 'work_key', 'write']` — no `project_root`, matching acceptance.md AC-003 RED-now.

### M1 — root-verification repair (flips AC-001)

- **Design decision (recorded deviation from plan.md §B/§F mechanics)**: the presented root is NOT passed as ProjectRoot. The serving-checkout anchor (server start dir's toplevel) is KEPT as projectRoot for the git-common-dir comparison, and ONLY the equality clause (old lines 308-310) plus the CallerDir coupling are dropped. Rationale: plan.md §G forbids widening acceptance to "any registered worktree of any repository — the same-repo boundary is the security boundary"; self-anchoring (root compared against itself) would vacate that boundary and make AC-003's gating arm (`belongs to a different repository` observable via codex_task) unsatisfiable. Acceptance.md/REQ layer governs over plan mechanics. `CallerDir` is fully removed from `codexAuditRequest` and the validator signature — no hidden second gate (plan §G).
- RED evidence (verbatim, before GREEN): `go test ./internal/cli/ -run 'TestCodexAuditRootAcceptsRegisteredServingRepoWorktrees' -count=1` → 6 failing arms, all through the equality clause — `sibling registered worktree accepted`, `primary checkout accepted`, `server primary anchor, lane root`, `symlinked lane root resolves` each with `refused: ... is not the caller's own worktree (.../A1)`; `unregistered_directory_refused` and `other_repository_refused` failing on premature equality refusal instead of the required fragments. Full transcript: /tmp red capture retained in this session; the failing fragments are quoted verbatim above and below.
- GREEN: `codexAuditValidateRoot(ctx, projectRoot, root)` — empty-root and empty-project-root refusals name the missing argument; symlink canonicalization, git-common-dir comparison, and registered `worktree list --porcelain` scan unchanged (REQ-003); handler passes `ProjectRoot: top, Root: root`; tool description and file-header rewritten (no server-started-in promise); residue grep `caller's own worktree|server started in` over internal/cli non-test → only the unrelated project_root-family comments remain.
- Affected-suite GREEN: `go test -timeout 30m ./internal/cli/ -run '(Test)?(CodexAudit|CodexTask)' -count=1` → `ok ... 34.051s` (post-fix); `go vet ./internal/cli/...` clean; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run ./internal/cli/...` → `0 issues.` (v2.1.6).
- Test-contract updates (SPEC-required, old-contract tables): `TestCodexAuditMCPTool` — "primary checkout"/"sibling worktree" moved from rejected to explicit accept arms (job_id returned); `TestCodexAuditLaunchDestinationConfinement` — legal combinations now A1/A/A2 with an in-tree destination (accepted==3); new `TestCodexAuditRootAcceptsRegisteredServingRepoWorktrees` pins the presented-root contract table (8 arms).

### M2 — codex_task project_root + super-advisor caller update (flips AC-003)

- RED evidence (verbatim, before GREEN): `go test ./internal/cli/ -run 'TestCodexTaskProjectRootGate' -count=1` → all 5 arms failed: `absent_project_root_refused` and `empty_project_root_refused` got `IsError:false` (the tool ran to a session instead of refusing — `Error:codex initialize rejected: codex stdout closed...`); `foreign_repository_refused` and `unregistered_directory_refused` got `IsError:false` (argument ignored); `registered_worktree_accepted_and_acted_on` timed out at the 10s bound (no gate, wrong cwd). Full transcript captured this session (M2 RED).
- GREEN: `project_root` added to the codex_task schema (Required) and gated in `handleCodexTask` through the SAME `codexAuditValidateRoot(servingTop, raw)` as codex_role_audit; absent/empty refused naming the argument (`project_root is required (pass your own git rev-parse --show-toplevel); codex_task does not default to any tree`); accepted root becomes the turn's `cwd`. **Scope note**: the job registry and the write opt-in (allow_write) stay bound to the serving project (`projectDirResolver()`) — REQ-005 requires the gate and the acted-on tree only; records stay server-scoped wherever the caller sits (matches the role-audit job model). The gate's git verification runs on the background context: it is a precondition decided before any session work, so a request cancelled mid-handshake surfaces cancellation at the session, not as a misleading refusal (found live by `TestCodexTask_ForegroundSessionStillBoundToRequest`'s 50ms-cancel race).
- M2 coupling (landed in the SAME milestone as the flip, per plan §F): super-advisor `codex_task` bullet updated in C1 `.claude/agents/moai/super-advisor.md` AND C2 `internal/template/templates/.claude/agents/moai/super-advisor.md` (project_root REQUIRED, own toplevel, never defaulted); C3 `internal/template/templates/.codex/agents/moai/super-advisor.toml` regenerated via `make agents-emit` (`AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission` → ok); `make build` re-embedded templates + regenerated `catalog.yaml` (13408 bytes), exit 0.
- Test fallout (SPEC-required harness adaptation, classified): every pre-existing codex_task test called the handler without `project_root` → now refused. `callCodexTask` harness injects this repository's own toplevel when the key is absent (documented; gate tests needing an ABSENT value call the handler raw); raw-handler sites in failcause / process_context / session_ctx tests inject `thisRepoRoot(t)` explicitly. No assertion weakened.
- Affected-suite GREEN: `go test -timeout 30m ./internal/cli/ -run '(Test)?(CodexAudit|CodexTask)' -count=1` → `ok ... 52.056s`; `go vet ./internal/cli/...` clean; `golangci-lint v2.1.6 run ./internal/cli/...` → `0 issues.`; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

### M3 — CLI role-audit verb (flips AC-002)

- RED evidence: live CLI at M0 (`unknown verb - usage: moai codex [cli] ...`, exit 1, tree 91c522dfe) + unit RED `go test ./internal/cli/ -run 'TestCodexRoleAuditVerb'` → compile-level RED `undefined: newCodexRoleAuditCmd` (the verb and its constructor did not exist).
- GREEN: `newCodexRoleAuditCmd()` registered under the codex cobra namespace (`role-audit <role> [--out <path>]`; no Use-prefix collision — no existing command starts with `role`); the RunE body of both verbs extracted into one shared `runCodexAuditVerb` so both travel through `runCodexAudit` → `prepareCodexAudit` → the repaired validator (no bypass shape, plan §G). Unit arms: `TestCodexRoleAuditVerbRegistered` + `TestCodexRoleAuditVerbRunsInCallerWorktree` (route shell, verdict verbatim, launch record) green.
- Live: `/tmp/t1324-run-moai codex role-audit --help` exit 0 printing the CLI-twin long text; `codex --help` lists `role-audit` (grep count 1).

### M4 — re-measure and residue sweep

- Binary rebuilt from the run tree (`make build`, commit `2f1631763` embedded): live re-observations —
  - AC-001 GREEN (live stdio session, server cwd = primary, worktree_root = this worktree, probe role unknown so no codex process spawns): refusal now comes from the ROLE loader — verbatim `codex_role_audit: codex audit no-such-role: role "no-such-role" is not a read-only contract role` — i.e. root validation PASSED; the AC-001 RED fragment `is not the caller's own worktree` is absent from the real wiring. Full-launch green (isError:false + job_id) is proven at unit level (`TestCodexAuditMCPTool/accepts/sibling worktree|primary checkout`) against the fake codex.
  - AC-003 GREEN live: `tools/list` → `codex_task -> ['background', 'project_root', 'prompt', 'resume_last', 'thread_id', 'work_key', 'write']`; live `tools/call` codex_task without the argument → verbatim `codex_task: project_root is required (pass your own git rev-parse --show-toplevel); codex_task does not default to any tree`, `isError: true`.
  - Piped-EOF artifact reproduced again (lower-layer refusal) — confirms plan §C's environment note; live-session form used for all role-audit MCP observations.
- E3 coverage (scoped `-run '(Test)?(CodexAudit|CodexTask|ProjectRoot)' -coverprofile`): touched-file function mean 90.8% (44 funcs across codex_audit_launch.go / codex_audit_mcp.go / codex_task.go); key funcs — `codexAuditValidateRoot` 88.9%, `handleCodexRoleAudit` 87.5%, `handleCodexTask` 88.8%, `runCodexAuditVerb` 66.7% (shared runner; the uncovered arms are the shared verb's already-covered-by-twin error paths), status/result handlers 87.5-100%. Above the 85% target.
- E4 residue greps (this tree): `callerTop|caller's own worktree` in internal/cli non-test → 0; `must be the worktree this server started in` / `server-started-in` across internal/cli + the super-advisor mirrors → 0; `AskUserQuestion` in the four touched non-test files → 3 matches, all inside comments (the no-AskUserQuestion documentation itself).
- Final batch: scoped suite `ok ... 48.867s`; `go vet ./internal/cli/...` clean; `golangci-lint v2.1.6 run ./internal/cli/...` → `0 issues.`; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go build ./...` exit 0.

### Known-red classification (NEW vs pre-existing)

- NEW test failures OUTSIDE the scoped suites, introduced by M2 and scheduled for sync by the SPEC itself: `TestProjectRootDocMatchesServer` and `TestDocsSiteProjectRootMatchesServer` (4 locale arms) — the doc-parity tests key `moai-mcp-tools.md` (+ docs-site locale lists) against the set of tools declaring `project_root`; M2 added `codex_task` to that family. Repair = editing `.claude/rules/moai/core/moai-mcp-tools.md` (+ mirror + docs-site), which plan.md §G forbids in run phase and acceptance.md §D.5 schedules for sync. Disposition: EXPECTED-SYNC-DEBT, not a code defect; the lane's scoped suites are green, and sync (same card, pre-merge per lane protocol) repairs the docs before any push.

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-09-29
- run_commit_sha: 793249821
- run_status: complete (M1-M4 landed; AC matrix in §E.2; one EXPECTED-SYNC-DEBT row pending the sync-phase doc repair, see §E.2 Known-red)
- ac_pass_count: 7
- ac_fail_count: 0
- preserve_list_post_run_count: 0
- l44_pre_commit_fetch: not-run (lane worktree; no origin fetch required for scoped run phase)
- l44_post_push_fetch: not-run (lane does not push; factory leader batch-pushes develop)
- new_warnings_or_lints_introduced: 0
- cross_platform_build.windows: pass
- cross_platform_build.darwin: pass
- total_run_phase_files: 20
- m1_to_mN_commit_strategy: one commit per milestone (M1 validator repair + status flip; M2 codex_task gate + interlocked super-advisor caller update; M3 CLI verb; M4 evidence; SHA backfill follow-up)

## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-09-29
- sync_commit_sha: pending-backfill-sync
- sync_status: complete
- expected_sync_debt: cleared — `TestProjectRootDocMatchesServer` + `TestDocsSiteProjectRootMatchesServer` (4 locale arms) now `ok`; repair = the project_root-family documentation updated to the post-M2 server schema (`codex_task` joins, 14 tools, codex_task's input required-never-defaulted): `.claude/rules/moai/core/moai-mcp-tools.md` + byte-identical template mirror (`cmp` → identical; `make build` exit 0, catalog.yaml 13408 bytes) + `docs-site/content/{en,ko,ja,zh}/guides/mcp-server.md`. The doc-conformance test's own count vocabulary (`docCountWords` map, 4 locale count phrases in `internal/cli/mcp_project_root_doc_test.go`) extended to fourteen/14 per lead authorization (coordinator message 2026-09-29) — the alternative of keeping the Thirteen phrase while listing 14 names was refused as a false published count.
- b12_self_test_a: pass — `grep -c 'SPEC-CODEX-ROLE-AUDIT-ROOT-001' CHANGELOG.md` = 0 pre-emission
- b12_self_test_b: pass — 7 distinct live AC identifiers in acceptance.md (AC-001..AC-007); the CHANGELOG entry references the same 7
- b12_self_test_c: pass — every file path cited in the CHANGELOG entry verified with `ls`
- changelog_entry_position: [Unreleased] → `### Added`, first entry
- frontmatter_status_transitions.spec_md: in-progress → implemented → completed (merged into the single sync commit; `updated:` refreshed 2026-09-29)
- canary_compliance_check.mx_tags: no MX tag surface touched — sync-phase repair is documentation-only (rule file + mirror + docs-site + doc-conformance test vocabulary)

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~6 files (internal/cli Go + tests, super-advisor frontmatter C1+C2, agents-emit regen); domain count 2 (Go source, agent/template mirrors); language mix Go + markdown; concurrency benefit LOW (coding-heavy); Agent Teams prereqs: not requested
- Mode evaluation: direct — not selected (semantic multi-file change); serial — **selected**; fanout — not selected (coding-heavy, Anthropic parallelism caveat); sweep — not selected (semantic, not mechanical)
- Decision: serial
- Justification: single-package Go implementation with an interlocked M2 coupling (refusal flip never lands before the caller update) — sequential per-milestone delegation to one manager-develop writer keeps the tree single-writer and matches the coding-heavy default (Anthropic: most coding tasks involve fewer truly parallelizable tasks than research).
