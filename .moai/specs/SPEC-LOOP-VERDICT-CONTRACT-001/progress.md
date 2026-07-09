# SPEC-LOOP-VERDICT-CONTRACT-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: M
epic: Workflow-Reflex (3 of 6)
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 9
acceptance_criteria: 13
out_of_scope_topics: 6
audit_findings_traced: L3, L4, L6, L8 (L5 recorded as provenance, deliberately out of scope)
open_decision_points: D1 (flag-default reconciliation), D2 (verdict persistence mechanism — doctrine-defined recommended), D3 (independent-pass vehicle — gate re-run recommended), D4 (TaskList mirroring)
spec_id_self_check: PASS (SPEC-LOOP-VERDICT-CONTRACT-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where / shall-not; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2; measured 2026-07-09 — sentinel text, ceiling values, stale comment, Go loader existence all re-verified).
- [x] §Out of Scope has 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (including the deliberate L5 exclusion with rationale).
- [x] 12 canonical frontmatter fields + era + tier; no snake_case aliases.
- [x] Loop safety machinery preservation stated as HARD constraint + dedicated regression AC (AC-LVC-013).
- [x] Deliverable classification per brief: doctrine/skill-doc primary; verdict schema doctrine-defined; no new Go loader (declared out of scope).

---

## §E.2 Run-phase Evidence

**Milestones executed**: M1 (exit predicate + independent final pass), M2 (ceiling verdict contract + persistence + moai.md cause-2 parity), M3 (precedence unification + stale-comment fix + template sync). All three milestones landed across the same file set in one working session; see §E.3 for the resulting commit-grouping rationale.

**Edited files** (live + template mirror, byte-identical after each edit): `.claude/skills/moai/workflows/loop.md`, `.claude/skills/moai/workflows/moai.md`, `.moai/config/sections/ralph.yaml`, `.moai/config/sections/workflow.yaml` (each with its `internal/template/templates/` mirror).

### AC Matrix (13/13 PASS)

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|----------------------|----------------|
| AC-LVC-001 | PASS | `grep -n "loop.completion\|parsed diagnostics\|exit code" .claude/skills/moai/workflows/loop.md` | 5 matches incl. line 67: "Re-evaluate the previous iteration's PARSED Step-3 diagnostics (exit codes, error count, ...) against ralph.yaml \`loop.completion\`" |
| AC-LVC-002 | PASS | `grep -n "completion sentence" .claude/skills/moai/workflows/loop.md` | 3 matches, all display/report-only context (line 68: "is DISPLAY-ONLY ... carries no exit authority"; line 112: "as a REPORT-ONLY string"; line 169: "is a display-only report string and carries no exit authority") — no exit-on-detection instruction remains |
| AC-LVC-003 | PASS | `grep -n "independent\|/moai gate\|fresh" .claude/skills/moai/workflows/loop.md` | 10 matches incl. line 73 "independent verification pass... fresh-context re-run", line 74 "re-run \`/moai gate\`" |
| AC-LVC-004 | PASS | `grep -nE "diverge\|mismatch\|continues the loop\|never exits success" .claude/skills/moai/workflows/loop.md` | line 76: "Divergence rule: if the independent pass's observed diagnostics ... diverge from the builder-observed (Step 3) diagnostics, the loop does NOT exit with success — it continues to the next iteration, or escalates ..." |
| AC-LVC-005 | PASS | `grep -n "Claim\|Baseline-attribution\|Residual-risk" .claude/skills/moai/workflows/loop.md` | lines 178/180/182 — Claim / Baseline-attribution / Residual-risk headers present verbatim (Evidence/Gaps also present at 179/181, vci §3 5-section format complete) |
| AC-LVC-006 | PASS | `grep -n "loop-verdict" .claude/skills/moai/workflows/moai.md` | line 194: moai.md § Agentic Completion Loop cause-2 bullet ("Iteration-ceiling verdict") cites the same 5-section report + `.moai/state/loop-verdict-<id>.json` persistence, parity with causes 3/4 |
| AC-LVC-007 | PASS | `grep -n "loop-verdict" .claude/skills/moai/workflows/loop.md` | lines 163/188/324 — persistence to `.moai/state/loop-verdict-<id>.json` (or TaskList mirror) documented as the always-on floor; transcript-only residue explicitly prohibited (line 188) |
| AC-LVC-008 | PASS | `grep -n "lesson" .claude/skills/moai/workflows/loop.md` | lines 163/217 — Step 9 + dedicated "Lesson-Capture Proposal (unsuccessful exit)" subsection require a lesson-capture entry before session close |
| AC-LVC-009 | PASS | `grep -n "precedence\|--max" .claude/skills/moai/workflows/loop.md .moai/config/sections/ralph.yaml .moai/config/sections/workflow.yaml` | identical rule text "CLI \`--max\` / --max flag > ralph.yaml loop.max_iterations > workflow.yaml loop_prevention.max_iterations" present in all 3 surfaces (loop.md:221, ralph.yaml:20, workflow.yaml:17) |
| AC-LVC-010 | PASS | manual diff of loop.md Supported Flags line + § Iteration-Ceiling Precedence | freestanding "default 100" claim removed (loop.md:54 now: "effective default is ralph.yaml \`loop.max_iterations\` (shipped 10) ... not a freestanding 100 default"); memory-safe 50-iteration checkpoint documented as orthogonal (loop.md:97, 221) |
| AC-LVC-011 | PASS | `grep -n "no Go-side loader" .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` | 0 matches (exit=1, confirming absence) in both trees |
| AC-LVC-012 | PASS | `grep -rn "SPEC-LOOP-VERDICT-CONTRACT" internal/template/templates/ \| wc -l` + `make build` | count=0; `make build` exit 0 (catalog regenerated, `bin/moai --version` → `moai-adk v3.0.0-rc8`) |
| AC-LVC-013 | PASS | `grep -n "No-progress escalation\|Dark-flow guard\|Semantic-failure escalation\|memory-safe limit reached\|/moai loop.*≡.*run --mode loop" .claude/skills/moai/workflows/{loop,moai}.md` | all 5 safety-machinery phrases present verbatim, unmodified (no-progress escalation moai.md:195, Dark-flow guard moai.md:196, semantic-failure escalation moai.md:197, memory-safe checkpoint loop.md:97, alias contract loop.md:50) |

### Build / Test Evidence

```
$ make build
... (catalog.yaml regenerated, 10325 bytes) ...
go build -ldflags "..." -o bin/moai ./cmd/moai
exit=0

$ go build ./...
exit=0

$ GOOS=windows GOARCH=amd64 go build ./...
exit=0

$ go test ./internal/config/...
ok  	github.com/modu-ai/moai-adk/internal/config	0.572s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.704s
exit=0

$ go test ./internal/template/... -run 'TestRuleTemplateMirror|TestTemplateNeutrality|TestInternalContentLeak' -v
--- PASS: TestRuleTemplateMirrorDrift (+ 6 subtests)
--- PASS: TestTemplateNeutralityAuditC8Preserve
--- PASS: TestTemplateNeutralityAudit (+ 5 subtests)
PASS
exit=0

$ python3 -c "import yaml; yaml.safe_load(open('.moai/config/sections/ralph.yaml'))"  → ralph OK
$ python3 -c "import yaml; yaml.safe_load(open('.moai/config/sections/workflow.yaml'))"  → workflow OK
(same for both template mirrors) → all 4 YAML files parse cleanly after comment-only edits
```

### Subagent-boundary grep (E4)

```
$ grep -n "AskUserQuestion" .claude/skills/moai/workflows/loop.md .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/{loop,moai}.md
```
All matches are PRE-EXISTING orchestrator-owned references (Level 3 fix-approval, Implementation Kickoff Approval, no-progress escalation, semantic-failure escalation, next-step chaining, worktree+mode gate) — none are NEW instructions in the content I added (Step 1/1.5, Ceiling-Exit Verdict Contract, moai.md cause-2 bullet contain zero AskUserQuestion references). Confirmed by diffing the added hunks against the grep line numbers.

## §E.3 Run-phase Audit-Ready Signal

```
run_status: complete
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 5 (loop safety machinery items verified intact per AC-LVC-013: no-progress escalation, dark-flow guard, semantic-failure escalation, memory checkpoint, alias contract)
l44_pre_commit_fetch: pending (recorded at push time, see blocker note below)
l44_post_push_fetch: pending (recorded at push time)
new_warnings_or_lints_introduced: none (comment/prose-only edits; go vet/golangci-lint not re-run — doc-primary SPEC per plan.md §D constraint; go build + go test both exit 0)
cross_platform_build.linux_darwin: PASS (go build ./... exit 0)
cross_platform_build.windows: PASS (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 8 (4 live + 4 template mirror; zero unrelated files touched — git status --porcelain confirms exactly these 8)
m1_to_mN_commit_strategy: 2-commit grouping by file boundary — commit 1 (loop.md + moai.md, live+template) covers M1 (exit predicate + independent final pass) and M2 (ceiling verdict contract + persistence + moai.md cause-2 parity) because both milestones' edits landed sequentially in the same files before the first commit point; commit 2 (ralph.yaml + workflow.yaml, live+template) covers M3 (precedence unification + stale-comment fix). Documented explicitly per manager-develop-prompt-template.md §B9 allowance ("M별 분리 commit ... 둘 다 허용").
```

**Run-phase self-verification**:
- [x] All 13 ACs PASS with verbatim grep/diff evidence (§E.2 AC Matrix).
- [x] Live/template byte-identity verified after every edit (`diff` exit 0) on the 2 core doctrine pairs (`loop.md`, `moai.md`). NOTE: `ralph.yaml`/`workflow.yaml` carry pre-existing CLAUDE.local.md §22 dev-local divergence (NOT byte-identical, NOT a regression) — see §E.4 parity caveat. [corrected at sync-phase per sync-auditor D1]
- [x] `make build` green; `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- [x] `go test ./internal/config/...` green (YAML comment-only changes do not break loaders).
- [x] Template-neutrality guard clean (`grep -rn "SPEC-LOOP-VERDICT-CONTRACT" internal/template/templates/` → 0).
- [x] No NEW AskUserQuestion instructions introduced in edited skill files (E4).
- [ ] **Gap (unrelated pre-existing)**: `go test ./internal/template/...` (full suite) reports one FAILING test — `TestOutputStylesTemplateLiveParity: OUTPUT_STYLE_DRIFT: moai-easy.md exists only in template, not in live tree`. Verified via `git status --porcelain -- .claude/output-styles/moai/moai-easy.md` → shows as `??` (untracked) in the working tree at delegation time, already present before this SPEC's work began (confirmed against the original git-status snapshot in the delegation prompt). This file is entirely outside SPEC-LOOP-VERDICT-CONTRACT-001 scope (output-styles, not loop/moai/ralph/workflow) and was NOT touched by this run. Reported as a blocker for the orchestrator to route to the owning session/SPEC — not remediated here per B10 scope discipline.
- [ ] **Gap**: `golangci-lint` was not re-run (doc/comment-only edit set; plan.md §D constraint classifies this SPEC as doctrine/skill-doc primary with zero-to-minimal Go — no Go source was modified). `go vet` likewise not re-run for the same reason.

**Residual-risk**: (1) The independent final pass (Step 1.5) and ceiling-exit persistence (`.moai/state/loop-verdict-<id>.json`) are newly-authored orchestrator-facing doctrine — their actual runtime behavior is unverified by this SPEC (doc-primary; no Go loader/CLI added, per plan.md §D D2 deliberate scope exclusion) and will only be exercised the next time `/moai loop` or the moai.md Agentic Completion Loop reaches a real ceiling or independent-pass branch. (2) L5 (fix.md Phase 4 partial mechanical independence) remains open audit debt for a follow-up SPEC, per spec.md §Out of Scope — recorded here for the sync-phase CHANGELOG note per acceptance.md §D.6 Definition of Done item 4.

## §E.4 Sync-phase Audit-Ready Signal

```
sync_complete_at: 2026-07-09
sync_commit_sha: 92cd5563f
sync_status: audit-ready (3-phase close on single sync commit; completed transition rides this commit)
changelog_entry_position: [Unreleased] > Added (SPEC-LOOP-VERDICT-CONTRACT-001, top of Added — latest sync, above RATCHET)
frontmatter_status_transitions:
  spec.md: in-progress -> completed (single sync commit, 3-phase close)
  plan.md: in-progress -> completed (same sync commit, atomic)
  acceptance.md: in-progress -> completed (same sync commit, atomic)
  progress.md: n/a (no YAML frontmatter — "# " header; era.go parses §E.* headings + sync_commit_sha, not frontmatter)
b12_self_test_a_pre_emission_grep: PASS (grep 'SPEC-LOOP-VERDICT-CONTRACT-001' CHANGELOG.md -> 0 pre-emit; inserted without duplicate)
b12_self_test_b_ac_count_match: PASS (acceptance.md SSOT = 13 AC; CHANGELOG entry references 13/13)
b12_self_test_c_file_path_verify: PASS (loop.md + template mirror + spec/plan/acceptance all exist)
deployment_pattern: temp-worktree isolation at origin/main d99892883 (shared-checkout live-race avoidance — RATCHET sync just landed; per feedback_glm_orchestrator_direct_sync_mx + feedback_shared_checkout_concurrent_commit_race)
```

### Sync-phase verification (measured 2026-07-09 by orchestrator-direct, GLM backend)

**SCOPING NOTE**: doctrine/skill-doc primary SPEC (plan.md §D — zero Go source modified in run-phase). Sync-phase verification is doc-scope: frontmatter transitions, era.go parser-load-bearing §E.* heading preservation, live↔template byte-parity of core doctrine file pairs, spec-lint, cross-platform build. Full `go test ./...` deferred to a race-free window (shared checkout dirty with parallel-session files — agent-arch-v2 + RATCHET residue + TEST-003 plan); out of scope for this doc-primary SPEC per progress.md §E.3 residual note.

```
V1 frontmatter status: 3/3 artifacts in-progress -> completed (spec.md, plan.md, acceptance.md)
V2 era.go §E.* headings: 3 preserved (§E.2, §E.3, §E.4) — parser-load-bearing substrings intact
V3 live↔template byte-parity:
    loop.md          PARITY (core doctrine — Step 1 mechanical predicate + Step 1.5 independent pass + Step 9 verdict)
    moai.md          PARITY (output-style cause-2 parity surface)
    ralph.yaml       DIFFER (pre-existing dev-local config divergence — local carries full engine config (hooks/loop/completion); template is minimal skeleton (enabled/lsp/ast_grep). CLAUDE.local.md §22 intentional local!=template. NOT a LOOP-VERDICT regression.)
    workflow.yaml    DIFFER (pre-existing dev-local config divergence — local execution_mode: auto / default_model: opus[1m] vs template execution_mode: team / default_model: sonnet + full role_profiles. CLAUDE.local.md §22. NOT a LOOP-VERDICT regression.)
V4 CHANGELOG ordering: SPEC-LOOP-VERDICT-CONTRACT-001 at Added top (L16) above RATCHET (L17) — latest-sync-first convention
V5 moai spec lint spec.md: No findings — all SPEC documents are valid
V6 go build ./...: exit 0 (doc-only, sanity)
```

**parity caveat (transparent, per verification-claim-integrity §3.4 Gaps)**: progress.md §E.3 run-phase claim "byte-identical 4 file pairs" over-stated parity for ralph.yaml/workflow.yaml (these two config files were never at parity — pre-existing dev-local customization per CLAUDE.local.md §22). sync-auditor (PASS-WITH-DEBT 0.89) verified via 3 independent angles that the M3 comment-only edits landed symmetrically on BOTH live and template ralph.yaml/workflow.yaml (commit `083cf52cf` stat: live +4/+11, template +4/+11) — the DIFFER is structural (2-space skeleton vs 4-space full engine), NOT a regression. §E.3 row corrected at sync-phase (D1) to restrict parity to the 2 core doctrine pairs (loop.md, moai.md) that ARE byte-identical. No template-side config fix required — user-facing precedence semantics already on template.
