# Progress — SPEC-PROJECT-HARNESS-BRIDGE-001

> Lifecycle progress tracker. §E section skeleton is parser-load-bearing
> (`internal/spec/era.go` matches the literal `§E.2` / `§E.3` / `§E.4` headings).
> Plan-phase populates §E.1 only; §E.2-§E.4 are placeholder headings owned by
> manager-develop (run) and manager-docs (sync).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-11
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (+ progress.md skeleton)
- REQ count: 12 (REQ-PHB-001 … REQ-PHB-012)
- AC count: 14 (AC-PHB-001 … AC-PHB-014)
- open_clarifications: 0 (both resolved: interview.yaml additional_axes
  config-declared → AC-PHB-013 applies; harness-spec.yaml OVERWRITE re-run
  semantics)
- depends_on: (none — foundation SPEC of the Project-Harness Pipeline Epic)

## §E.2 Run-phase Evidence

Doc-only SPEC (markdown / yaml). Verification is grep / diff / `make build` / `moai init`
per acceptance.md §C. Test-coverage ACs are N/A (no Go code touched); stated honestly.

| AC | Status | Verification command (acceptance.md §C) | Actual output |
|----|--------|------------------------------------------|---------------|
| AC-PHB-001 | PASS | `grep -c clarity_threshold` + `Phase 0.3` in mode-detection.md; `clarity_threshold` + `Phase 1.5` in codebase-analysis.md | mode-detection: clarity_threshold=1, Phase 0.3=1; codebase-analysis: clarity_threshold=1, Phase 1.5=1 |
| AC-PHB-002 | PASS | `clarity\|max_rounds` + `sufficiency\|early.exit\|additional round\|abandon\|entry floor` in both hosts | mode-detection: 13 / 8; codebase-analysis: 13 / 8 |
| AC-PHB-003 | PASS | `grep -c clarity-interview` in both hosts | mode-detection: 2; codebase-analysis: 2 |
| AC-PHB-004 | PASS | four-axis greps (verification / ui.surface / external.system / team.shar) in both hosts | mode-detection: 4/2/7/2; codebase-analysis: 5/3/4/2 (all ≥ 1) |
| AC-PHB-005 | PASS | `grep -cE domain\|goal\|constraints\|scope\|verification\|external_systems\|ui_surface\|team_sharing doc-generation.md` (≥ 8) | 14 |
| AC-PHB-006 | PASS | 8-field schema + `harness-spec.yaml` present in spec.md §D | spec.md schema-field count ≥ 8, harness-spec.yaml ≥ 1 (schema authored plan-phase §D) |
| AC-PHB-007 | PASS | `.moai/project/harness-spec.yaml` ≥ 1 AND `.moai/specs/.*harness-spec` == 0 in doc-generation.md | 3 and 0 |
| AC-PHB-008 | PASS | `grep -c harness-spec.yaml meta-harness.md` (≥ 1); §5.1 + product/structure/tech + Phase 4.2 intent preserved (D3) | 1; compose line retains `product.md`/`structure.md`/`tech.md` + "user's stated intent from Phase 4.2" |
| AC-PHB-009 | PASS | `harness-spec.yaml` ≥ 1 AND pre-satisfy language ≥ 1 in harness-build-entry.md | 1 and 1 (PRE-SATISFY / already-answered / absent or ambiguous / only…absent) |
| AC-PHB-010 | PASS | guard prose preserved (mode-detection ≥ 1); no NEW `.moai/specs/` write path (excl. guard) == 0; `.moai/project/` ≥ 1 | guard prose = 2; new-write-path = 0; `.moai/project/` = 29 |
| AC-PHB-011 | PASS | `diff -q` all 5 skill file-pairs + interview.yaml | PARITY OK ×6 (all byte-identical) |
| AC-PHB-012 | PASS | `make build` exit 0; `go test ./internal/template/...` ok; `grep -rn SPEC-PROJECT-HARNESS-BRIDGE-001 internal/template/templates/` == 0 | make build exit 0; `ok internal/template 2.446s`; leak grep = 0 |
| AC-PHB-013 | PASS | `verification\|ui_surface\|external_systems\|team_sharing\|additional_axes` ≥ 1 in interview.yaml + byte-parity | 6; PARITY OK |
| AC-PHB-014 | PASS | `moai init` sandbox → clarity_threshold ≥ 1 (mode-detection.md), harness-spec.yaml ≥ 1 (doc-generation.md) | init exit 0; clarity_threshold=1; harness-spec.yaml=6; additional_axes(interview.yaml)=1 |

Invariant preservation (asserted STILL EXISTS at exit):
- NO-SPEC guard prose in mode-detection.md L30/L46/L50 — PRESERVED (guard grep = 2).
- `.moai/project/interview.md` human-readable output — PRESERVED (harness-spec.yaml is additive).
- `interview.yaml` `clarity_threshold: 4` + `project.max_rounds: 3` — PRESERVED (consumed, not changed).
- builder-harness specialist-generation internals — UNCHANGED (not touched).
- D3 "+ Phase 4.2 intent" clause in meta-harness §5.1 — PRESERVED (ADD, not replace).

## §E.3 Run-phase Audit-Ready Signal

- run_status: complete
- run_complete_at: 2026-07-11
- run_commit_sha: 9f1f0f53b2a65c3f91be09457d0f217de462354e (landed-on-main commit; run-phase implemented in a Claude Code L1 worktree and landed to main by the orchestrator — the pre-land worktree SHA 0ec4ee5f3 is superseded by this main SHA; backfilled per spec-frontmatter-schema.md SHA-placeholder-backfill exemption; all 14 ACs independently re-verified PASS on the main tree post-land)
- ac_pass_count: 14
- ac_fail_count: 0
- test_coverage_ac: N/A (doc-only SPEC — no Go code touched; grep/diff/make-build/moai-init verification per acceptance.md §E)
- preserve_list_post_run_count: 5 preserved invariants verified (NO-SPEC guard, interview.md, clarity_threshold/max_rounds, builder internals, Phase 4.2 intent)
- new_warnings_or_lints_introduced: 0 (doc-only; template neutrality + internal-content-leak guards green)
- cross_platform_build: native `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- whole_repo_test: `go test ./...` exit 1 — SOLE failure is a PRE-EXISTING cross-package parallel-contention flake (`TestDoctorPermission_NoMatchTrace` in internal/cli, shared `permissionCmd` cobra global-state); PASSES deterministically in isolation and when internal/cli runs alone (exit 0). NOT a regression: changeset is 12 markdown/yaml files, zero Go, internal/cli not in changeset. PASS-WITH-DEBT (pre-existing flake, unrelated to this SPEC).
- template_guards: `go test ./internal/template/...` ok (neutrality + internal_content_leak_test.go green)
- neutrality_grep: `grep -rn SPEC-PROJECT-HARNESS-BRIDGE-001 internal/template/templates/` = 0
- total_run_phase_files: 12 (6 local skill/config + 6 template mirrors; all byte-identical)
- catalog_cascade: none (catalog.yaml regenerated by make build but byte-identical — the `moai` unified skill catalog hash does not track workflow sub-files)
- m1_to_mN_commit_strategy: consolidated single run-phase commit covering M1-M4 (doc-only, interdependent mirror-parity edits) + one SHA-backfill follow-up commit

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

## §F Phase 0.95 Mode Selection

**Input parameters**
- tier: M
- scope (file count): 6 file-pairs (12 files) — mode-detection.md, codebase-analysis.md, doc-generation.md, meta-harness.md, harness-build-entry.md, interview.yaml + template mirrors
- domain count: 2 (project workflow skills + config yaml); single mechanical intent per file but files are inter-dependent via byte-parity mirror
- file language mix: 100% markdown/yaml (no Go source)
- concurrency benefit: LOW — Template-First mirror discipline + byte-parity + shared interview.yaml overwrite create sequential dependencies; not genuinely parallel
- Implementation Kickoff Approval: PASSED (user approved run-phase entry)

**Mode evaluation**
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-file semantic edit, not a typo |
| 2 background | no | write work (not read-only) |
| 3 agent-team | no | RETIRED (tombstone) |
| 4 parallel | no | not multi-domain research-heavy; edits are inter-dependent (mirror parity), coding-adjacent |
| 5 sub-agent | **YES** | Tier M doc-only, sequential milestones M1-M4 with mirror-parity dependencies; default fallback |
| 6 workflow | no | < 30 files, not a single uniform mechanical transform; multi-milestone dependent edits |

**Decision: sub-agent**

**Justification**: This is Tier M doc-only work across 6 interdependent file-pairs bound by Template-First byte-parity (edit template → mirror local → make build). The milestones M1-M4 have ordering dependencies (M1 axes feed M2 harness-spec.yaml schema; M2 feeds M3 consumption; M4 verifies parity across all). Per Anthropic's coding-task parallelism caveat, coding-adjacent interdependent edits belong to Mode 5 (sequential sub-agent), not the parallel/workflow modes. A single manager-develop spawn executes M1→M4 sequentially. Run-phase executed in a Claude Code runtime-materialized L1 worktree; commits landed to main by the orchestrator (worktree base 604dbf04a was an ancestor of origin/main — conflict-free, zero overlap with the 3 concurrent SPEC-AGENT-TEAM-RETIRE-001 commits).
