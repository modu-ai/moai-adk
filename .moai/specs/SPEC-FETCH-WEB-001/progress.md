# SPEC-FETCH-WEB-001 — Progress

Card: t1968 · Tier M · plan-phase artifacts authored 2026-10-26. Scope source: leader analysis `.moai/reports/leader-plan-20261023/jevfetch-analysis-20261026.md` (applications 1-4).

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed Bash, output cited verbatim) |
| SPEC ID collision | none (`ls .moai/specs/ | grep -c '^SPEC-FETCH'` → `0`) |
| Tier | M — REQ 16 / ceiling 16; AC 14 / ceiling 16 |
| Artifact set | spec.md · plan.md · acceptance.md · progress.md |
| Requirements | 16 (REQ-FW-001 … REQ-FW-016; numbering stable across revision 0.2.0) |
| Acceptance criteria | 14 (AC-FW-001 … AC-FW-014; numbering stable) + RED-now evidence ledger (10 rows — R1-R7 + behavior probes R8-R10, tree `4f0e58197d`, all re-observed post-repair) |
| Plan audit | iter1 FAIL 0.80 (3 blocking D1-D3, 4 should-fix D4-D7, 2 minor D8-D9; report `.moai/reports/t1968/plan-audit.md`) → revision 0.2.0 applied with source-verified premises; B1/B5/B6 lane-adjudicated and resolved |
| Dependency | `SPEC-JEV-CORE-001` |
| Boundary | t1762 (`SPEC-JEV-LEDGER-001`) owns `internal/jevledger` (new) + its config/template surfaces; `internal/jev` is exclusively this SPEC's client surface (REQ-FW-016, revised per D7) |
| Follow-up linkage | headless-render fallback deferred (t1965 linkage) |
| Status transition | (none) → draft (no transition in revision 0.2.0) |

plan_status: audit-ready
plan_complete_at: 2026-10-26

## §E.2 Run-phase Evidence

Milestones M1–M7 executed TDD (RED captured before each implementation; the M1–M6 RED outputs are the pre-GREEN `build failed: undefined` runs quoted in the §E.8 ledger below; M7's RED basis is the plan-phase ledger rows R6/R7 at pin `4f0e58197d`). All commands below ran in the card worktree `.moai/worktrees/t1968` on the final run-phase tree `b81bf4d1a1` (2026-10-26).

| AC | Status | Verification command | Actual output (verbatim, abridged only where marked) |
|----|--------|---------------------|---------------|
| AC-FW-001 | PASS | `go test -run 'TestSplitBlocks\|TestBlockTokens\|TestDocTokens' -v ./internal/fetchweb/` | 12 `--- PASS` lines (kinds, heading paths, pop-on-same-level, nested lists, table header association, headerless table, code newlines, chrome zero-blocks, aside separation, document order, whitespace collapse, estimator seam); `ok github.com/modu-ai/moai-adk/internal/fetchweb 0.151s` |
| AC-FW-002 | PASS | `go test -run 'TestStage1\|TestSelection\|TestProcess\|TestNeighborContext' ./internal/fetchweb/` | `ok github.com/modu-ai/moai-adk/internal/fetchweb 0.152s` (candidate threshold, window bound, stage-2 keep + same-heading rescue, rescue-needs-qualifying-neighbor, no-match-first, first-block keep, backfill under 300 + score order, table-header accompaniment, document-order emission, aside separate scoring, unavailable→fetch_error) |
| AC-FW-003 | PASS | `go test -run 'TestProcess_Passthrough\|TestProcess_Over2000\|TestProcess_Empty' ./internal/fetchweb/` | covered by the AC-FW-002 run: passthrough ≤2,000 tokens with zero scoring calls (fail-on-call answerer), over-bound is scored, empty precedes passthrough |
| AC-FW-004 | PASS | `go test -run 'TestFetchWeb_StatusExitContract' -v ./internal/cli/` | 7 sub-tests PASS: filtered→0, no_match branch (a)→1 (no first-block keep), branch (b)→filtered 0 with first block present, passthrough→0 (3/3 blocks), empty→1, fetch_error→2, --json parseable with same status/exit_code |
| AC-FW-005 | PASS | `go test -run 'TestFetchWebCommand_Surface' -v ./internal/cli/` | `--- PASS: TestFetchWebCommand_Surface` (help documents `web <url>`, `--json`; parent name `fetch`) |
| AC-FW-006 | PASS | `go test -run 'TestFetchWebTool' -v ./internal/cli/` | 4 PASS: InCatalogue, GateOffZeroNetwork, DispatchPathParity (handler result == direct RunFetchPipeline result), ShapeMatchesCLIJSON |
| AC-FW-007 | PASS | `go test -run 'TestFetchWeb_OffGateZeroNetwork' -v ./internal/cli/` | `--- PASS: TestFetchWeb_OffGateZeroNetwork` (0 network calls + 0 Jev calls through tripwires, stderr names WebFetch, exit 2, empty stdout) |
| AC-FW-008 | PASS | `go test -run 'TestFetchWeb_E2EFakeServers' -v ./internal/cli/` | `--- PASS: TestFetchWeb_E2EFakeServers` (fake page server + fake Jev server, filtered contract, verbatim block text, `## Doc` heading lines) |
| AC-FW-009 | PASS | `go test -run 'TestCache' -v ./internal/jev/` | `ok ... internal/jev` — hit without network + restart-offline hit, model-id-in-key two files, explicit-override miss → network → own entry (no cross-model stale serve), atomic writes (no tmp residue, 5/5 parse) |
| AC-FW-010 | PASS | `go test -run 'TestRetryAfter\|TestSemaphore' -v ./internal/jev/` | `ok ... internal/jev` — Retry-After:1 waited ≥900ms then retried (2 calls), delay capped at 30s, bounded budget (2 attempts), non-repeatable NOT retried (1 call), cache hit issues no retry (0 calls), semaphore width 3 observed max ≤3 with 8 concurrent |
| AC-FW-011 | PASS | `go test -run 'TestUsage_FromCacheSeparation\|TestCache_StaleServeKeepAndWarn' -v ./internal/jev/` | `ok ... internal/jev` — usage records [network FromCache=false, cache FromCache=true]; alias-moved stale serve names cached model id in the warning + entry kept + zero network |
| AC-FW-012 | PASS | `go test -run 'TestLoadBenchCases\|TestRunBench\|TestCompareRun\|TestBaseline' ./internal/jevmeasure/` + `go test -run 'TestJevmeasureBench' ./internal/cli/` | both `ok` — containment + savings rate + summary; baseline write → deterministic PASS re-run → degraded run FAIL exit 1; command gated off = zero network exit 2 |
| AC-FW-013 | PASS | `go test -run 'TestFetchWebSkillShips\|TestFetchWebPermissionRuleShips\|TestFetchWebGLMRoutingRowShips\|TestFetchWebArtifactsAreNeutral' -v ./internal/template/` | 4 PASS (skill covers WebFetch/--json/no_match/network_access/gate; `Bash(moai fetch web:*)` rule; `jev active → fetch_web first` row; neutrality of the shipped lines) |
| AC-FW-014 | PASS | `grep -rn "tiktoken" internal/fetchweb/` → 0 hits; `grep -rn 'AskUserQuestion' internal/fetchweb internal/cli/fetch_web.go internal/cli/mcp_fetch_web.go internal/cli/jevmeasure_cmd.go internal/jev/cache.go` → 0 hits; `git diff cb30e05b3e..HEAD -- go.mod` → 0 lines; `go test -run 'TestFetchWeb_NoLiveEndpointInTests' ./internal/cli/` → PASS | all outputs verbatim-zero as stated; no test dials a real endpoint |

Invariants (REQ-JEVC-008 composition / REQ-JEVC-002 stdlib-only / REQ-FW-016 boundary):

| Invariant | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| Existing jev retry suite unchanged and green | PASS | `go test ./internal/jev/` | `ok github.com/modu-ai/moai-adk/internal/jev 1.548s` |
| internal/jev stdlib-only imports | PASS | `go test -run TestPackageImports ./internal/jev/` | covered by the package run (display_only_test green) |
| t1762 boundary: internal/jevledger untouched | PASS | `git diff --name-only cb30e05b3e..HEAD -- internal/jevledger` | (empty) — positive control `-- internal/jev` returned 3 files |
| Template neutrality guard | PASS | `go test -run TestTemplateNeutralityAudit ./internal/template/` + full package `go test ./internal/template/` | `ok ... internal/template 258.997s` (full suite, exit 0) |
| `make build` (template parity) | PASS | `make build` | catalog regenerated, embed manifest enrolled, binary built, all checks green |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-26
run_commit_sha: b81bf4d1a1
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0   # internal/jevledger diff vs merge-base cb30e05b3e: empty (positive control internal/jev: 3 files)
l44_pre_commit_fetch: n/a (card worktree, lane landing — no origin/main push performed)
l44_post_push_fetch: not-pushed (the lane owns landing; no push, no PR, per B9)
new_warnings_or_lints_introduced: 0   # golangci-lint over fetchweb/jev/jevmeasure/cli/mcp/template: "0 issues." (repo-wide baseline staticcheck finding in internal/discovery untouched)
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 39   # git diff --name-only 62583ecf57..HEAD
m1_to_mN_commit_strategy: per-milestone Conventional Commits on WT-jevfetch-a2sys-ai (M1 edca0736c0, M2 61bf2da4ff, M3 b7b5938f5f, M4 9123657813, M5 1558cdc13e, M6 02482a9323, M7 1dd4e19d8e, lint chore b81bf4d1a1) + this evidence commit
```

### §E.8 RED evidence (TDD, per milestone)

Every Go milestone's first failing run was captured before its implementation existed. Verbatim pre-GREEN outputs (abridged to the discriminating lines; full runs in the session record):

- **M1** — `go test ./internal/fetchweb/...` → `internal/fetchweb/blocks_test.go:106:12: undefined: SplitBlocks` … `FAIL github.com/modu-ai/moai-adk/internal/fetchweb [build failed]`
- **M2** — `go vet ./internal/fetchweb/` → `internal/fetchweb/scoring_test.go:63:12: undefined: stage1Scores` / `undefined: Stage1Threshold` / `undefined: Process` / `undefined: StatusFiltered` … `FAIL [build failed]`
- **M3** — `go test ./internal/fetchweb/` → `undefined: FetchPage` / `undefined: MaxResponseBytes` / `undefined: RenderMarkdown` / `undefined: RenderJSON`; `go vet ./internal/cli/` → `undefined: fetchWebPageClientFn` / `undefined: fetchWebAskerFn`
- **M4** — `go test -run 'TestFetchWebTool' ./internal/cli/` → `undefined: handleFetchWeb` (3 sites) `FAIL [build failed]`
- **M5** — `go test ./internal/jev/` → `c.CacheDir undefined (type *Client has no field or method CacheDir)` / `ModelOverride undefined` … `FAIL [build failed]`
- **M6** — `go test ./internal/jevmeasure/` → `undefined: BenchCase` / `undefined: LoadBenchCases` / `undefined: RunBench` / `undefined: CompareRun`; CLI leg → `undefined: newJevmeasureCmd`
- **M7** — markdown artifact milestone; RED basis is the plan-phase ledger rows R6 (`ls internal/template/templates/.claude/skills/moai-fetch-web` → `No such file or directory`, exit 1) and R7 (permission rule zero hits in settings.json.tmpl) at pin `4f0e58197d`, plus the first packaging-test run failing on the routing-row literal before the template text carried it.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-27
sync_commit_sha: pending-backfill-sync   # D3 placeholder — a commit cannot cite its own hash; backfilled in the next commit
sync_status: complete
ac_source: .moai/specs/SPEC-FETCH-WEB-001/acceptance.md   # tier M
b12_self_test_a: pass   # pre-emission grep -c 'SPEC-FETCH-WEB-001' CHANGELOG.md -> 0 (exit 1); no duplicate entry
b12_self_test_b: pass   # AC counter on acceptance.md: live=14 excluded=0 ambiguous=0; CHANGELOG entry cites 14 (AC-FW-001..014)
b12_self_test_c: pass   # every path named in the entry verified present by ls/grep (internal/fetchweb/{blocks,scoring,fetch,output}.go, internal/cli/{fetch_web,mcp_fetch_web}.go, internal/cli/mcp_server.go, internal/jev/cache.go, internal/jevmeasure/bench.go, internal/cli/jevmeasure_cmd.go, skills/moai-fetch-web/SKILL.md, settings.json.tmpl:569, glm-web-tooling.md:40)
changelog_entry_position: CHANGELOG.md [Unreleased] > Added — first entry (directly under the '### Added' heading)
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed   # single sync commit; owner manager-docs (Authored-By-Agent trailer)
  plan_md: no status field and no frontmatter block — stateless artifact, nothing to refresh
  acceptance_md: no status field and no frontmatter block — stateless artifact, nothing to refresh
mx_tag_validation: 0 tags added, 4 existing verified accurate (cli gate-path WARN x2 with REASON, RunFetchPipeline ANCHOR with REASON, jev package ANCHOR with REASON); no doctrine trigger fires — zero 'go func' sites in the new surfaces, the semaphore and both retry sleeps carry and honor context.Context, no exported function reaches fan_in >= 3
```

## §F Phase 4 Mode Selection

Mode Selection — orchestrator: lane-7 (factory card t1968), 2026-10-26. Decision: serial (rationale and mode table below).

| Input | Value |
|---|---|
| tier | M (SPEC-FETCH-WEB-001) |
| scope (files) | ~10-14 (internal/fetchweb new pkg, internal/jev client surface, internal/jevmeasure, internal/cli, mcp_server.go, templates skill/settings/rule) |
| domain count | 3 (Go code, templates, rules/docs) — coding-heavy |
| file language mix | Go + markdown templates |
| concurrency benefit | LOW (Anthropic coding-task caveat: new-code sequential work) |
| agent-team prereqs | not requested (no --team) |

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | New package + multi-file implementation, not a one-liner |
| serial | **yes** | Coding-heavy new-feature work; milestones M1-M7 sequential in one package |
| fanout | no | Write fan-out on one card prohibited; research fan-out not needed (design source complete) |
| sweep | no | Semantic new code, not mechanical bulk; <30 files |

Decision: serial
Justification: the milestone chain (M1-M7) builds one new package with intra-package dependencies; Anthropic's coding-task parallelism caveat selects sequential single-agent delegation; write-capable concurrency on one card is prohibited regardless. manager-develop (cycle_type=tdd) carries M1-M7 serially.

