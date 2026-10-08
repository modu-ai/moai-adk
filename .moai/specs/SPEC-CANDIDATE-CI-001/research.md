# research.md — SPEC-CANDIDATE-CI-001

Every existing-behavior claim in spec.md/plan.md, with the file:line evidence actually
read on tree db0c514d3 (branch WT-10-03-tier), plus the live measurements taken during
plan-phase research (2026-10-09).

---

## R1 — The integration verb family (the surface this SPEC extends)

| Claim | Evidence |
|---|---|
| `moai integration` registers seven subcommands: status, acquire, release, policy, remeasure, merge, preflight | internal/cli/integration.go:269 (`cmd.AddCommand(...)` line read directly) |
| The lock record lives in the PRIMARY checkout's state, shared by all worktrees | internal/cli/integration.go:12-15, :33-47 (`integrationLockRoot`) |
| `resolveIntegrationTarget` order: explicit --branch flag → configured develop branch → caller's tree | internal/cli/integration.go:152-183 |
| The configured integration branch resolves through the flow-scoped loader (git-flow→develop, github-flow→main), not a literal | internal/cli/integration.go:273-282 (`configuredIntegrationBranch` → `config.LoadGitFlowIntegrationConfig(...).IntegrationTarget`) |
| acquire records the OWNING SESSION's pid and refuses-with-record semantics | internal/cli/integration.go:443-466 |
| `--wait` queues behind a held window (REQ-MWQ-002/012) | internal/cli/integration.go:468-484 |
| remeasure verb runs a command and stores a record keyed by the candidate tree; verdict errors are returned AFTER the record prints | internal/cli/integration_remeasure.go:12-51, :40-47 |
| Live RED: `go run ./cmd/moai integration candidate --help` on db0c514d3 exits 0 and prints the PARENT help; the only "candidate" text is remeasure's Short line | measured 2026-10-09 in this worktree (exit 0; `grep -c candidate` == 1, the remeasure Short) |

## R2 — The merge step and its pre-allocated seams (the gate this SPEC plugs into)

| Claim | Evidence |
|---|---|
| The merge step is a 13-cause ordered gate (codes 1-15 + complete's 20) | internal/factory/integration_merge_step.go:33-59 |
| Gate 5 is the landing check, driven by `MergeStepSeams.LandingCheck`; nil or key-disabled = absent no-op seam | internal/factory/integration_merge_step.go:103-106, :303-309 |
| The seam is DOCUMENTED as this SPEC's: "LandingCheck is SPEC-CANDIDATE-CI-001's REQ-CCI-011 shared check" | internal/factory/integration_merge_step.go:104-105 |
| Card branch resolution is documented as this SPEC's REQ-CCI-004 contract ("the resolver behind is REQ-CCI-004's contract"; "the REQ-CCI-004 contract resolves the tree") | internal/factory/integration_merge_step.go:251-252, :675-677 |
| `candidateCIEnabled` is a constant-false placeholder owning the key name `workflow.candidate_ci.enabled`, @MX:UPGRADE names t1478's landing | internal/cli/integration_merge.go:175-189 |
| The merge verb currently wires LandingCheck to a LOUD refusal placeholder while t1478 is unlanded | internal/cli/integration_merge.go:90-99 |
| The merge target branch is read from the window RECORD first, config fallback, literal `develop` last resort | internal/cli/integration_merge.go:47-61 |
| The merge worktree resolver refuses the primary checkout | internal/cli/integration_merge.go:120-141 |
| Post-merge holds use a policy record naming the card and merge SHA | internal/factory/integration_merge_step.go:570-582 (`CompletePostMergeHold`) |
| Merge gate reads the pinned SHA from `refs/heads/<cardBranch>` and the tree from `<pinned>^{tree}` | internal/factory/integration_merge_step.go:251-266 |

## R3 — merge-tree prior art (the candidate construction primitive)

| Claim | Evidence |
|---|---|
| `git merge-tree --write-tree` is the established dry-run probe; exit 1 reports conflicts as a RESULT, not a tool failure | internal/factorylane/merge.go:36-46, :53-61, :272-277 |
| Tree identity: a clean merge commit's tree equals the merge-tree result (`HEAD^{tree} == HEAD^2^{tree}` on the real merge) | internal/factorylane/merge.go:40-46 |

## R4 — Card state machine (what moves, what is reserved)

| Claim | Evidence |
|---|---|
| The 21-state F1 machine includes merged-local → pushed → ci-green → done and the github-flow pair pr-open/merged-pr | internal/homestate/card_record.go:17-64, :76-78 |
| `pushed → ci-green` and `ci-green → done` are RESERVED edges; the transition API refuses them: "the CI verdict reader that admits it is owned by F3" | internal/homestate/card_transition.go:183-185, :294-295 |
| T17 merged-local → pushed is `guardPush` | internal/homestate/card_transition.go:137, :587 |

## R5 — CI workflow current shape (the workflow this card restructures)

| Claim | Evidence |
|---|---|
| Triggers: push→[main], pull_request→[main], workflow_dispatch; NO ci/** pattern | .github/workflows/ci.yml:16-25 |
| `concurrency: cancel-in-progress: true` for EVERY ref | .github/workflows/ci.yml:35-37 |
| Race detector already split 2-way by test-name prefix: `-run '^(TestFactory\|TestInstall\|TestTodo\|TestRun)'` (826 tests / 834.5s) vs the `-skip` leg (4264 / 947s), measured on develop run 37185766779 | .github/workflows/ci.yml:251-257, :298, :357 |
| internal/cli -race measured ~379s ≈ 70% of race runtime (in-tree corroboration of the card's 369s family figure) | .github/workflows/ci.yml:279-281 |
| Per-binary -timeout raised to 35m with the +43% same-class variance rationale; ceiling headroom discipline documented in-step | .github/workflows/ci.yml:288-298 |
| Both race jobs run `scripts/ci-census/test-census.sh` on the -json stream and upload the compressed stream | .github/workflows/ci.yml:299-300, :358-359 |
| Build job vets cross-platform (`go vet ./...` under GOOS/GOARCH matrix incl. windows/amd64) because `go build` does not compile _test.go | .github/workflows/ci.yml:570-582 |
| Required-check NAME parity contract (test-skip-marker must emit the same names) | .github/workflows/ci.yml:376-384 |
| detect job uses the shared filter .github/test-input-filters.yml | .github/workflows/ci.yml:52-62 |
| Census scripts: per-test census from a go test -json stream; single-predicate skip detection; build-failure rows; fixture check census-check.sh | scripts/ci-census/test-census.sh:1-45, scripts/ci-census/census-check.sh:1-30 |

## R6 — Config surfaces (where the gate key lives)

| Claim | Evidence |
|---|---|
| workflow.yaml is the `workflow:` section loaded via Loader.Load() into cfg.Workflow | .claude/rules/moai/core/settings-management.md § MoAI Configuration → Section Loaders (loader table row `workflow.yaml`); .moai/config/sections/workflow.yaml:1 |
| Gate-flag families live as workflow.yaml sub-keys with Go structs: `integration_lock` (types.go:515), `settings_drift_gate` (types.go:526) — `candidate_ci` follows the same pattern | internal/config/types.go:510-526 |
| The template mirror of workflow.yaml exists and must change with the live file (Template-First) | internal/template/templates/.moai/config/sections/workflow.yaml (file present, 18,367 bytes) |
| YAML-symmetry and loader-completeness audits run on ./internal/config/ | .claude/rules/moai/core/settings-management.md § CI Guards |
| Active git-strategy profile: `workflow: git-flow`, `develop_branch: develop`; flip to github-flow deferred to SPEC-GITHUB-FLOW-DEFAULT-001 M5; `lead_push_threshold: 20` with the count command in comments | .moai/config/sections/git-strategy.yaml:9-16, :23-35 |

## R7 — Local doctrine (the rule this card amends)

| Claim | Evidence |
|---|---|
| Lanes never push develop; the leader batch-pushes (2026-09-02 operator directive); the delivery.md Step 3.2 push stage is EXCLUDED repo-locally | .claude/rules/local/gitflow-lane-protocol.md:39 (§2 EXCLUDED clause), :77-89 (§4) |
| Green-conditional batch push: a red last-push holds the next push; the merge-window re-measure gate caught the last 3 reds pre-push | .claude/rules/local/gitflow-lane-protocol.md:87 (§4 초록 조건부) |
| The file is local-only, never mirrored to templates | .claude/rules/local/gitflow-lane-protocol.md:10, :188 |
| Remote CI on the integration branch is the authoritative verdict surface | .claude/rules/local/gitflow-lane-protocol.md:83 |
| Conflict resolution is the owning lane's duty; unsolvable conflicts are a leader blocker | .claude/rules/local/gitflow-lane-protocol.md:93 (§5) |
| The card text's "§4.1" names the WT-push/CI-request prohibition; CLAUDE.local.md is the untracked sibling (.gitignore line 276) carrying §4.1 | card text (verbatim in the dispatch); /CLAUDE.local.md gitignore entry (worktree .gitignore:276); CLAUDE.local.md absent from the worktree (untracked) |

## R8 — Guard families (the bundle's members)

| Claim | Evidence |
|---|---|
| Source-scan guard precedent: the destructive-target registry check statically parses Go source rather than reading its own table ("That independence is the whole point") | internal/cli/update_destructive_registry.go:13-21 |
| Line-key guard precedent: AST guards record 1-based line numbers; the vocabulary guard records line numbers without keying drift to them | internal/cli/codex_launcher_guards_test.go:301; internal/cli/vocabulary_guard_test.go:20 |
| Anti-precedent the registry design cites AGAINST line keys: the destructive registry is deliberately (File, Function, Sites)-keyed because "a line-keyed registry would report drift after every unrelated change" — the drift class investigation C measured | internal/cli/update_destructive_registry.go:23-27 |
| Census guard machinery: test-census.sh + census-check.sh fixture check | scripts/ci-census/ (both files read; see R5) |

## R9 — Flaky prior art (the registry's evidence discipline)

| Claim | Evidence |
|---|---|
| Flaky failures admitted only from verbatim GitHub Actions logs, "추정이 아니다"; branch protection strict:true makes one flaky red a merge delay | .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/spec.md §A (frontmatter + §A table read) |
| CI retry-policy adjustment was explicitly ruled OUT of that SPEC's scope as symptom-hiding — the registry in THIS SPEC records instead of hides, and root-cause repair stays with the series | .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/spec.md Out of Scope bullet ("CI 워크플로 재시도 정책(`retry` 횟수) 조정 — 증상 은폐이므로…") |
| No gotestsum / rerun mechanism exists in ci.yml or scripts/ci-census today | grep over .github/workflows/ci.yml + scripts/ci-census/*.sh for `gotestsum\|rerun\|retry` → 0 rows (measured 2026-10-09) |

## R10 — REQ-CCI mnemonic collision (disclosed, not silently absorbed)

| Claim | Evidence |
|---|---|
| SPEC-V3R6-CLI-CONFIG-INTEGRITY-001 owns REQ-CCI-001/002 and REQ-CCI-006/007 in the CLI-config-integrity domain | internal/cli/update.go:113; internal/cli/launcher.go:1131; internal/cli/profile_setup.go:26, :35-36, :361 |
| The factory code anchors cite REQ-CCI-004 + REQ-CCI-011 for THIS SPEC | internal/factory/integration_merge_step.go:104, :252, :675 (R2) |

## R11 — Operator-provided figures NOT verifiable in-tree (carried as card facts, re-measured at run phase)

- "develop CI 24회 연속 비초록" and "가드 드리프트 12종·플랫폼 드리프트, 순수 회귀 0건" — investigation C
  (card text). No in-tree artifact records the 24-run series or the 12-kind taxonomy.
  The SPEC treats them as the operator's evidence for WHY, and plan.md M5/M6 re-measure
  what they motivate.
- "369s 테스트" — investigation C figure for the dominant internal/cli race test. Nearest
  in-tree corroboration is "~379s/70%" (ci.yml:281). M6 re-measures per-test before
  repairing (plan.md §B B2).
