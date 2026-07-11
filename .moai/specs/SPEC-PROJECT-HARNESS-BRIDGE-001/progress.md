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

### M5 (v0.2.0 amendment) — two-stage interview / Round 4 reachability

Delivered M5.1-M5.7. All 20 ACs re-verified on the post-M5 tree: AC-PHB-001 … AC-PHB-014
(v0.1.1 regression guard) STILL PASS, and AC-PHB-015 … AC-PHB-020 (the REACHABILITY ACs,
the amendment's exit criteria) newly PASS.

| AC | Status | Verification command (acceptance.md §C) | Actual output (post-M5) |
|----|--------|------------------------------------------|-------------------------|
| AC-PHB-001 | PASS | `clarity_threshold` + `Phase 0.3` (mode-detection); `clarity_threshold` + `Phase 1.5` (codebase-analysis) | 2 / 1 ; 2 / 1 (all ≥ 1) |
| AC-PHB-002 | PASS | `clarity\|max_rounds` + `sufficiency\|early.exit\|additional round\|abandon\|entry floor`, both hosts | mode-detection 23 / 16; codebase-analysis 23 / 16 |
| AC-PHB-003 | PASS | `grep -c clarity-interview`, both hosts | 2 ; 2 |
| AC-PHB-004 | PASS | four-axis greps, both hosts | mode-detection 5/2/7/2; codebase-analysis 6/3/4/2 (all ≥ 1) |
| AC-PHB-005 | PASS | 8 schema-field grep in doc-generation.md (≥ 8) | 16 |
| AC-PHB-006 | PASS | 8-field schema + `harness-spec.yaml` in spec.md §D | 40 ; 33 |
| AC-PHB-007 | PASS | `.moai/project/harness-spec.yaml` ≥ 1 AND `.moai/specs/.*harness-spec` == 0 | 3 and 0 |
| AC-PHB-008 | PASS | `grep -c harness-spec.yaml meta-harness.md` (≥ 1) | 1 (untouched by M5 — v0.1.x wiring stands) |
| AC-PHB-009 | PASS | `harness-spec.yaml` ≥ 1 AND pre-satisfy language ≥ 1 in harness-build-entry.md | 1 and 1 (untouched by M5) |
| AC-PHB-010 | PASS | guard prose ≥ 1; NEW `.moai/specs/` write path (excl. guard) == 0; `.moai/project/` ≥ 1 | 2 ; 0 ; 29 |
| AC-PHB-011 | PASS | `diff -q` all 7 file-pairs (incl. the new `project.md` pair + interview.yaml) | PARITY OK ×7 (all byte-identical) |
| AC-PHB-012 | PASS | `make build` exit 0; `go test ./internal/template/...`; SPEC-ID leak grep == 0 | make build exit 0; `ok internal/template 1.320s`; leak grep = 0 |
| AC-PHB-013 | PASS | axes grep ≥ 1 in interview.yaml + byte-parity | 7 ; PARITY OK |
| AC-PHB-014 | PASS | `moai init` sandbox → clarity_threshold ≥ 1; harness-spec.yaml ≥ 1 | init exit 0; clarity_threshold=2; harness-spec.yaml=6; (M5 extras: Stage A/B=33 & 35; interview.yaml Stage A=4; `3-round`=0) |
| **AC-PHB-015** | **PASS** | (a) exemption contract; (b) anchored to Round 4 / Stage B (`grep -A 12`); (c) `Stage A\|Stage B` ≥ 2 — both hosts | mode-detection 5 / 7 / 33; codebase-analysis 5 / 7 / 35 |
| **AC-PHB-016** | **PASS** | (a) required-base-fields precondition in early-exit prose (`grep -A 6`); (b) exit scoped to Stage A — both hosts | mode-detection 3 / 15; codebase-analysis 3 / 15 |
| **AC-PHB-017** | **PASS** | (a) Stage-A scoping in interview.yaml; (b) exempt round named within `grep -A 4 max_rounds`; (c) byte-parity | 5 ; 4 ; PARITY OK |
| **AC-PHB-018** | **PASS** | (a) explicit base-field coverage mapping; (b) mapping names all four fields (`grep -A 10`, ≥ 4) — both hosts | mode-detection 3 / 6; codebase-analysis 3 / 4 |
| **AC-PHB-019** | **PASS** | (a) `REQUIRED\|EXTENDED` in spec.md ≥ 2; (b) partition + extended fields in doc-generation.md ≥ 1 | 20 ; 14 and 6 |
| **AC-PHB-020** | **PASS** | `grep -rn "3-round"` across both trees `\| wc -l` == 0 (baseline 6) | 0 (case-insensitive sweep incl. `3-Round` also 0) |

M5 defect closed: the extended-axes Round 4 was UNREACHABLE on both interview paths
(`project.max_rounds: 3` capped the interview AND the clarity early-exit skipped "the
remaining rounds"). Both hosts are now restructured into Stage A (clarity-scored,
`max_rounds`-capped) + Stage B (mandatory Round 4, EXEMPT from the cap, the early-exit
skip, the abandon path, and clarity scoring). The exemption prose — not the mere presence
of a Round 4 heading — is the reachability contract, and AC-PHB-015(b) anchors the grep to
that section.

M5 invariant preservation (re-asserted post-M5):
- NO-SPEC guard prose in mode-detection.md — PRESERVED (guard grep = 2; NEW `.moai/specs/` write paths = 0).
- `.moai/project/interview.md` human-readable output — PRESERVED (Stage A/B section names updated; the artifact is not replaced).
- `interview.yaml` `clarity_threshold: 4` + `project.max_rounds: 3` VALUES — PRESERVED (only in-place documentation added; `max_rounds` was NOT raised to 4).
- The v0.1.1 early exit — AMENDED (gated on required base fields + rescoped to Stage A), never removed (AC-PHB-002 still PASS).
- builder-harness internals, meta-harness.md, harness-build-entry.md — UNTOUCHED by M5.
- D3 "+ Phase 4.2 intent" clause in meta-harness §5.1 — PRESERVED (file not modified).

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

### M5 (v0.2.0 amendment) run-phase audit-ready signal

- run_status: complete (M5 — two-stage interview / Round 4 reachability)
- run_complete_at: 2026-07-11
- run_commit_sha: 3af01567fdd56dae3afb485fcf0b1f735fcd22e6 (backfilled per spec-frontmatter-schema.md SHA-placeholder-backfill exemption — a commit cannot reference its own hash; all 20 ACs verified on this tree)
- ac_pass_count: 20 (AC-PHB-001 … AC-PHB-020)
- ac_fail_count: 0
- ac_regression_guard: AC-PHB-001 … AC-PHB-014 re-verified PASS on the post-M5 tree (v0.1.1 behavior not weakened)
- ac_amendment_exit_criteria: AC-PHB-015 … AC-PHB-020 (REACHABILITY) newly PASS
- test_coverage_ac: N/A (doc-only SPEC — markdown/yaml only, zero Go; verification is grep / diff / `make build` / `moai init` per acceptance.md §E)
- subagent_boundary_grep: N/A (no Go code touched; no `internal/` package in the changeset)
- preserve_list_post_run_count: 6 preserved invariants verified (NO-SPEC guard, interview.md, clarity_threshold/max_rounds VALUES, the v0.1.1 early exit amended-not-removed, builder/meta-harness/harness-build-entry untouched, Phase 4.2 intent clause)
- new_warnings_or_lints_introduced: 0 (doc-only; template neutrality + internal-content-leak guards green)
- cross_platform_build: native `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- whole_repo_test: `go test ./...` exit 0 — FULLY GREEN this run (zero FAIL lines). Note: a PRE-EXISTING wall-clock flake exists at `TestBudget_FullRepoScanWithin35Sec` (scripts/i18n-validator, 35s budget) which can fail under parallel contention; it did NOT trigger in this run (`ok scripts/i18n-validator 3.992s`). This corrects the v0.1.1 §E.3 entry, which mis-attributed the flake to `TestDoctorPermission_NoMatchTrace`.
- template_guards: `go test ./internal/template/...` → `ok internal/template 1.320s`
- neutrality_grep: `grep -rn SPEC-PROJECT-HARNESS-BRIDGE-001 internal/template/templates/` = 0
- total_run_phase_files: 10 (5 local skill/config + 5 template mirrors; all byte-identical) + progress.md
- file_pair_set: 7 pairs total; M5 touched 5 (mode-detection, codebase-analysis, doc-generation, project.md [NEW 7th pair], interview.yaml); meta-harness.md + harness-build-entry.md untouched (v0.1.x wiring stands)
- catalog_cascade: none (catalog.yaml regenerated by `make build` but byte-identical — workflow sub-files are not catalog-hash subjects)
- m1_to_mN_commit_strategy: single consolidated M5 commit (doc-only, interdependent mirror-parity edits across 5 file-pairs) + one SHA-backfill follow-up commit
- plan_artifact_note: plan.md §F M5 (M5.1-M5.7) was authored at plan-phase but left uncommitted in the working tree by the v0.2.0 amendment commit (a2b66c01c); persisted verbatim in commit 09e83b1e4 before M5 run-phase began. No body content authored or altered by manager-develop.

## §E.4 Sync-phase Audit-Ready Signal

### v0.2.0 amendment close (supersedes the v0.1.1 §E.4 record below)

- sync_status: complete
- sync_complete_at: 2026-07-11
- sync_commit_sha: pending-backfill-v0.2.0-amendment-close (self-referential hazard — a commit cannot reference its own hash; backfilled with the real full SHA in an immediate follow-up commit per spec-frontmatter-schema.md SHA-placeholder-backfill exemption; verify via `git show --stat <sha>`)
- ac_pass_count: 22 (AC-PHB-001 … AC-PHB-022; verified `grep -cE '^### AC-PHB-[0-9]{3}' acceptance.md` == 22 before emission)
- doc_sync_summary: CHANGELOG.md `[Unreleased]` § Added entry for SPEC-PROJECT-HARNESS-BRIDGE-001 was **REWRITTEN IN PLACE**, NOT appended. Rationale: the existing CHANGELOG.md:12 entry (from the prior v0.1.1 close, sync commit `0c9871b46b6719325427dc0126e4eb65d7b0f2d8`) described the pre-amendment BROKEN semantics as if they were shipped-and-working — it claimed a delivered "new Round 4" and "14/14 AC PASS", when in fact Round 4 (the four extended axes) was UNREACHABLE on both interview paths (the `project.max_rounds: 3` cap AND the REQ-PHB-002 early-exit "skip remaining rounds" both blocked it). Because v0.1.1 and this v0.2.0 amendment land in the SAME `[Unreleased]` cycle (nothing has shipped to users yet), the entry now describes the FINAL two-stage-interview behavior (Stage A adaptive discovery + Stage B mandatory extended-axes round, exempt from the cap/early-exit/scoring) and the corrected **22/22 AC PASS** count, rather than carrying a broken-then-fixed narrative forward. The true historical framing that this SPEC replaces the prior STATIC fixed-3-round interview was preserved (that statement remains accurate); only the description of the BROKEN Round-4-unreachable semantics and the stale AC count were corrected. README.md and docs-site 4-locale sync remain explicitly OUT OF SCOPE per spec.md § Out of Scope — deferred follow-up, no user-facing CLI surface change.
- mx_tag_validation: N/A — doc-only SPEC (markdown/yaml), no Go code touched; no @MX annotation targets
- frontmatter_transition: spec.md `status: in-progress → completed` (single sync commit carries the v0.2.0 amendment's 3-phase close); `updated: 2026-07-11` (unchanged — plan-phase, run-phase (M5), and sync-phase all land same day); plan.md/acceptance.md have no `status:` frontmatter field (untouched)
- amendment_close_confirmation: the v0.2.0 in-place amendment (`completed → in-progress → completed`) is closed. F-1 (Round 4 unreachable — cap + early-exit double-block), F-2 (early exit could strand REQUIRED base fields), F-3 (asymmetric Round-topic coverage across the two hosts), F-3r (residual asymmetry re-check), and F-5 (stale `3-round` fixed-interview claim in codebase-analysis.md frontmatter) are all closed per AC-PHB-015 … AC-PHB-022. Foundation SPEC of the 3-SPEC Project-Harness Pipeline Epic — downstream SPEC-HARNESS-MCP-PROVISION-001 and SPEC-HARNESS-VERIFY-PROMOTE-001 (both `depends_on` this SPEC) may now build on the `harness-spec.yaml` contract as amended.

### v0.1.1 close (superseded — retained for audit trail only; see amendment note above)

- sync_status: complete
- sync_complete_at: 2026-07-11
- sync_commit_sha: 0c9871b46b6719325427dc0126e4eb65d7b0f2d8 (backfilled per spec-frontmatter-schema.md SHA-placeholder-backfill exemption)
- doc_sync_summary: CHANGELOG.md `[Unreleased]` § Added entry appended (verified `grep -c "PROJECT-HARNESS-BRIDGE-001" CHANGELOG.md` == 0 before emission); README.md and docs-site 4-locale sync explicitly OUT OF SCOPE per spec.md § Out of Scope — CHANGELOG / README / docs-site (deferred follow-up, no user-facing CLI surface change)
- mx_tag_validation: N/A — doc-only SPEC (markdown/yaml), no Go code touched; no @MX annotation targets
- frontmatter_transition: spec.md `status: in-progress → completed` (single sync commit, 3-phase close); `updated: 2026-07-11` (unchanged — plan-phase and sync-phase land same day); plan.md/acceptance.md have no `status:` frontmatter field (untouched)
- close_confirmation: SPEC-PROJECT-HARNESS-BRIDGE-001 sync-phase complete; foundation SPEC of the 3-SPEC Project-Harness Pipeline Epic — downstream SPEC-HARNESS-MCP-PROVISION-001 and SPEC-HARNESS-VERIFY-PROMOTE-001 (both `depends_on` this SPEC) may now proceed
- **NOTE (added at v0.2.0 amendment close)**: this v0.1.1 close was subsequently found by an adversarial sync-audit to have shipped a BROKEN headline feature (Round 4 / extended axes unreachable). The SPEC was reopened as the v0.2.0 in-place amendment above. This record is retained verbatim for audit-trail purposes; it is NOT the current close signal.

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
