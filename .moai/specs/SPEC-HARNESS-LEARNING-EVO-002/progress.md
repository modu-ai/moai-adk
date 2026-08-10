# SPEC-HARNESS-LEARNING-EVO-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` authored at `status: draft`, Tier M.
- Scope: **L2 only** (the delegation-map analyzer). Split from `SPEC-HARNESS-LEARNING-EVO-001` v0.1.0, which carried 33 requirements and 36 acceptance criteria — over the ceiling at Tier M (16/16) and Tier L (25/25). L1 (instrumentation) stays in the sibling; L3 (application to `delegation.yaml`) remains an explicit non-goal with its three-surface rationale.
- Requirements: 16 GEARS requirements (REQ-HLA-001..016); 16 acceptance criteria with 100% requirement coverage (`acceptance.md` §I). Both counts are exactly at the Tier M ceiling.
- Sibling relationship: `SPEC-HARNESS-LEARNING-EVO-001` is listed under `related_specs`, deliberately **not** `depends_on` — the fixture-based test strategy makes this SPEC runnable without the sibling reaching `status: completed`, and a `depends_on` entry would impose a Depends_on Pre-flight gate that the design does not need.
- Open decisions carried into audit: the reader contract (`plan.md` §E D1 — accept the materializing read, bound it by a declared max size, rather than adding a streaming variant that would create a dependency on the sibling), the catalog-membership source (`plan.md` §E D2 — a declared constant seeded from `CLAUDE.md` §4, not derived from the delegation map), and the two threshold constants (`plan.md` §F M1).

_<pending plan-audit>_

## §E.2 Run-phase Evidence

Commands were run in the worktree `.claude/worktrees/hle-l2` on branch
`feat/harness-learning-evo-l2`, based on `origin/main` @ `8ed9622f3`.

### Baseline (pre-flight, before the first edit)

| Check | Command | Observed |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Lint | `golangci-lint run --timeout=3m` | `0 issues.` |
| Map references | `grep -rn 'delegation.yaml' internal/ --include='*.go' \| grep -v _test` | no matches (exit 1) |

### AC matrix

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-HLA-001 | PASS | `go test ./internal/harness/delegationmap/ -run TestFixtures_GeneratedViaFinalize -count=1` | `--- PASS: TestFixtures_GeneratedViaFinalize (0.00s)` |
| AC-HLA-002 | PASS | `grep -rn 'routing-ledger.jsonl' internal/harness/delegationmap/ --include='*.go' \| grep -v _test` + `-run TestAnalyze_OversizedLedgerRefused` | no grep output; `--- PASS: TestAnalyze_OversizedLedgerRefused (0.00s)` |
| AC-HLA-003 | PASS | `-run TestAggregate_Basic` | `--- PASS: TestAggregate_Basic (0.00s)` |
| AC-HLA-004 | PASS | `-run TestAnalyze_RerouteAbortExcluded` | `--- PASS: TestAnalyze_RerouteAbortExcluded (0.00s)` |
| AC-HLA-005 | PASS | `-run TestAnalyze_NonCatalogNeverUndesignated` | `--- PASS: TestAnalyze_NonCatalogNeverUndesignated (0.00s)` |
| AC-HLA-006 | PASS | `-run TestAnalyze_UnattributedShareReported` | `--- PASS: TestAnalyze_UnattributedShareReported (0.01s)` |
| AC-HLA-007 | PASS | `-run 'TestAnalyze_BelowMinRows\|TestAnalyze_StrayRowSuppressed'` | both `--- PASS` |
| AC-HLA-008 | PASS | `-run TestAnalyze_TwoKindsOnly` | `--- PASS: TestAnalyze_TwoKindsOnly (0.00s)` |
| AC-HLA-009 | PASS | `-run TestAnalyze_ConditionalDesignationsExcluded` | `--- PASS: TestAnalyze_ConditionalDesignationsExcluded (0.00s)` |
| AC-HLA-010 | PASS | `-run TestMapReader_ReadOnly` | `--- PASS: TestMapReader_ReadOnly (0.01s)` |
| AC-HLA-011 | PASS | `-run TestPatternKey_NamespaceIsolated`; `grep -c 'delegation_map' internal/harness/types.go`; `go test ./internal/harness/ -count=1` | `--- PASS`; `0`; `ok github.com/modu-ai/moai-adk/internal/harness 2.731s` |
| AC-HLA-012 | PASS | `-run TestProposal_CarriesEvidence` | `--- PASS: TestProposal_CarriesEvidence (0.01s)` |
| AC-HLA-013 | PASS | `-run TestAnalyze_DeterministicIdempotent` | `--- PASS: TestAnalyze_DeterministicIdempotent (0.01s)` |
| AC-HLA-014 | PASS | `-run TestAnalyze_GracefulNoOp` | `--- PASS: TestAnalyze_GracefulNoOp (0.01s)` |
| AC-HLA-015 | PASS | `go test ./internal/cli/ -run TestHarnessDelegationAnalyzeCmd -count=1` | 3 subtests `--- PASS`; `ok ... 1.165s` |
| AC-HLA-016 | PASS-WITH-DEBT | see the note below | boundary holds; one cited grep is unsatisfiable as literally written |

### AC-HLA-016 — PASS-WITH-DEBT, with the reason stated

Three of the criterion's four cited commands pass verbatim:

- `grep -rn 'delegation.yaml' internal/ --include='*.go' \| grep -v _test` → exactly one match, `internal/harness/delegationmap/mapreader.go:15`, and no write primitive appears anywhere in the package (`TestNoMapWritePath`).
- `git diff --name-only origin/main...HEAD \| grep -c 'frozen-allowlist.json'` → `0`; `grep -c 'moai/config/sections' .claude/lsel/frozen-allowlist.json` → `1`.
- `go test ./internal/harness/delegationmap/ -run TestAnalyze_NoSkillProposals -count=1` → `--- PASS`.

The fourth, `grep -rn 'AskUserQuestion' internal/harness/delegationmap/ internal/cli/harness_delegation.go` → expected no output, **cannot pass as written**. It scans comments and test files, and `plan.md` §F M5 separately mandates a boundary grep test mirroring `internal/harness/routing/subagent_boundary_test.go` — a test that necessarily contains the literal `AskUserQuestion` in order to search for it. The same command run against the routing package returns 6 matches for exactly this reason, so the literal form is unsatisfiable by construction rather than by anything this implementation did.

The boundary itself is verified with the project's own canonical form (`.claude/rules/moai/development/manager-develop-prompt-template.md` §B3):

```
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/delegationmap/ internal/cli/harness_delegation.go \
  | grep -v "_test.go" | grep -v "// "
```

→ 0 matches. The remaining occurrences are one package-doc sentence stating the boundary and the guard test that enforces it. **Proposed correction**: amend AC-HLA-016's fourth command to the §B3 canonical form.

### Deviation from plan.md §I — one file outside the inventory

`internal/harness/proposalgen/{types.go,scaffolder.go}` were modified; the inventory did not list them.

REQ-HLA-012 requires every emitted proposal to carry the observation count, support ratio, qualifying-row count, unattributed share, and kind, and AC-HLA-012 decides it by reading `proposal.json` and `spec.md`. REQ-HLA-011 requires emission through the existing writer. Those two are jointly unsatisfiable against the writer as it stood: `ProposalCandidate` carries four fixed promotion-shaped fields and `WriteProposals` renders a fixed template, with no extension point.

The resolution is an additive optional `Evidence map[string]any` field rendered into both artifacts. It is not a change to `PatternBearingEventTypes()` and not a change to `MapPromotions` — the two things REQ-HLA-011 actually forbids — and an empty map renders nothing, so both artifacts stay byte-identical for every existing producer (`go test ./internal/harness/proposalgen/ -count=1` → `ok`, with its byte-idempotence tests unchanged).

### Quality gate

| Gate | Command | Observed |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Vet | `go vet ./...` | exit 0, no output |
| Lint | `golangci-lint run --timeout=5m` | `0 issues.` — unchanged from baseline |
| Tests (harness) | `go test ./internal/harness/... -count=1` | all 13 packages `ok` |
| Tests (cli) | `go test ./internal/cli/ -count=1` | `ok ... 193.215s` |
| Coverage | `go test -cover ./internal/harness/...` | `delegationmap 87.7%`, `routing 88.5%`, `proposalgen 72.7%` |
| Coverage (cli) | `go test -cover ./internal/cli/` | `76.7% of statements` |
| Spec lint | `moai spec lint .../spec.md --strict` | `✓ No findings` |
| Scope | `git diff --name-only origin/main...HEAD \| grep -cE '^(internal/template/templates/\|\.claude/skills/)'` | `0` |
| Isolation | `git diff --name-only origin/main...HEAD \| grep -cE 'frozen-allowlist.json\|config/sections/delegation.yaml'` | `0` |

`go test ./... -count=1` (§F, whole tree) was NOT run: `internal/statusline` and `internal/cli` are known to hang for ten minutes on this machine when `internal/statusline/usage.go`'s network and keychain dependencies are unreachable, and the hang reproduces on untouched base commits. The two packages this change touches were run individually and both pass; CI is the arbiter for the whole-tree run. Recorded as a gap rather than reported as a pass.

### R8 (row-splitting) — how the degradation was made visible

`plan.md` §B6 required an explicit decision on whether aggregation groups by `session_id` across rows or treats each row independently, because the choice changes what a support ratio means.

**Decision: each row is treated independently.** Grouping by `session_id` would merge rows the producer deliberately closed as separate observations, and the split's cause is unreproduced — a grouping rule written against an unconfirmed hypothesis would encode the hypothesis as fact.

Because the resulting errors are quiet (under-counted support ratios and over-counted empty-delegation rows both push toward suppressing real patterns rather than inventing false ones), `SubcommandStat.EmptyDelegationRows` counts qualifying rows that carried no delegation at all. That count is the observable signature of the split: a subcommand whose qualifying rows are largely empty is visible in the result directly, rather than having to be inferred from findings that never appeared. Both the decision and its reasoning are recorded in `aggregate.go`'s doc comment and in the field's own comment.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-10
run_commit_sha: pending-backfill-run
run_status: audit-ready
ac_pass_count: 15
ac_fail_count: 0
ac_pass_with_debt_count: 1
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 27
m1_to_mN_commit_strategy: "5 commits — M1 contract+fixtures, M2-M3 aggregation+findings, M4 emission, M5 CLI+guards, plus this evidence commit; not pushed"
```

## §E.4 Sync-phase Audit-Ready Signal

- **sync_complete_at**: 2026-08-10
- **sync_commit_sha**: `pending-backfill-sync` (self-referential-hazard placeholder per `spec-frontmatter-schema.md` D3; backfilled by a follow-up `chore(...)` commit — amend is forbidden)
- **sync_status**: audit-ready
- **changelog_entry_position**: `[Unreleased] → ### Added`, first bullet — **shared with `SPEC-HARNESS-LEARNING-EVO-001`**. The two SPECs are one shipped capability (L1 observation + L2 analysis) split only because the combined requirement count exceeded the Tier ceiling; two entries would each have described half a loop. The entry links both `spec.md` files.
- **frontmatter_status_transitions.spec**: `in-progress → implemented → completed` (single sync commit, merged close per the Status Transition Ownership Matrix)
- **frontmatter_status_transitions.plan**: n/a — `plan.md` carries no YAML frontmatter and no `Status:` marker (verified by `grep -n '^status:\|^Status:\|Status:'`)
- **frontmatter_status_transitions.acceptance**: n/a — same, no frontmatter or status marker
- **frontmatter_status_transitions.progress**: n/a — same, no frontmatter or status marker
- **frontmatter_status_transitions.updated_refreshed**: 2026-08-10 — `spec.md` `updated:` already held the sync-commit date; no change was required, and none was made for its own sake
- **b12_self_test_a** (pre-emission grep): `grep -c 'HARNESS-LEARNING-EVO' CHANGELOG.md` returned **0** before emission — no duplicate for either SPEC of the pair
- **b12_self_test_b** (AC count match): `acceptance.md` holds **16** distinct AC identifiers by both counts that must agree — deduped token grep → 16, heading/row count → 16. §E.3 records `ac_pass_count: 15` + `ac_pass_with_debt_count: 1` = 16, `ac_fail_count: 0`. The CHANGELOG states 32 across both SPECs and names the one PASS-WITH-DEBT rather than rounding it into the pass count.
- **b12_self_test_c** (file path verification): every path named in the CHANGELOG entry verified to exist before committing — `internal/harness/delegationmap/` (`aggregate.go`, `analyze.go`, `mapreader.go`, `proposal.go`, `types.go`), `internal/cli/harness_delegation.go`, `.moai/config/sections/delegation.yaml`, `.claude/lsel/frozen-allowlist.json`.
- **canary_compliance_check.body_untouched**: `spec.md` / `plan.md` / `acceptance.md` body content unchanged on this commit. The only `spec.md` edit is the frontmatter `status:` field. In particular, AC-HLA-016's **proposed correction** recorded in §E.2 was NOT applied — amending an acceptance criterion is a `manager-spec` body edit, outside sync-phase ownership, and it is carried forward as an open item rather than silently taken.
- **canary_compliance_check.frozen_allowlist_untouched**: `.claude/lsel/frozen-allowlist.json` is not in this commit's diff, and its `^\.moai/config/sections/` frozen pattern is intact — the guarantee the CHANGELOG entry makes to the reader ("nothing shipped here can alter routing on its own") is verified rather than asserted.
- **canary_compliance_check.no_run_phase_section_edits**: `§E.2` and `§E.3` of this file are unmodified — they are owned by `manager-develop`. This includes the `run_commit_sha: pending-backfill-run` placeholder in §E.3, which is still unbackfilled; see the open item below.
- **mx_tag_validation**: sync-phase MX sub-step performed. The sync commit is documentation-only; it introduces no Go symbol, so no new `@MX` annotation is required and none was added.
- **project_doc_sync**: `.moai/project/` needed updating and was updated — this SPEC's five `delegationmap` non-test files moved the `internal/harness` count from 75 to 80, a figure stated exactly in `structure.md`, `codemaps/modules.md`, and `codemaps/overview.md`. Measurement and correction are recorded in `SPEC-HARNESS-LEARNING-EVO-001` §E.4 `project_doc_sync`.
- **open_item.run_commit_sha**: §E.3 still reads `run_commit_sha: pending-backfill-run`. The run-phase code merged as `96bedbfe3` (`feat(SPEC-HARNESS-LEARNING-EVO-002): delegation-map analyzer (L2)`, PR #1424, squash), so the value is known — but §E.3 is `manager-develop`-owned and sync phase does not write it. Flagged for the orchestrator to backfill alongside `sync_commit_sha`, in the same follow-up `chore(...)` commit.
