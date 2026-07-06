# SPEC-HOOK-PREEDIT-INVESTIGATE-001 — Progress

> Living document tracking plan/run/sync phase evidence. Section structure per `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map. The `§E.*` namespace is parser-load-bearing for era.go classification — see the Section Map before renaming any `§E.N` heading.

## §A. Plan-Phase Entry

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-HOOK-PREEDIT-INVESTIGATE-001 |
| Tier | S |
| Lifecycle | spec-anchored |
| Created | 2026-07-07 |
| Author (plan) | manager-spec |
| Plan-phase commit | (pending — plan-phase artifacts not yet committed) |
| plan-auditor iter-1 verdict | (pending — orchestrator has not yet invoked plan-auditor) |
| Implementation Kickoff Approval | (pending — orchestrator has not yet run the AskUserQuestion gate) |

## §B. Run-Phase Entry

| Field | Value |
|-------|-------|
| Mode (Phase 0.95) | (pending — populated by orchestrator before first run-phase Agent() spawn) |
| cycle_type | (pending — tdd / ddd / autofix selected at run-phase entry) |
| Run-phase commit (M1) | (pending) |
| Run-phase start timestamp | (pending) |

## §C. Sync-Phase Entry

| Field | Value |
|-------|-------|
| Sync-phase commit | (pending) |
| Sync-phase start timestamp | (pending) |
| CHANGELOG entry added | (pending) |

## §D. Open Blockers

(none — plan-phase artifact creation is the current step; no blockers surfaced during authoring)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-07T00:00:00Z
plan_artifacts:
  - spec.md (12 canonical frontmatter fields, 12 REQ-FF requirements in GEARS notation, 7 Out-of-Scope entries)
  - plan.md (single M1 milestone, file-by-file change list with Template-First ordering, 12 known-issues-filtered constraints)
  - acceptance.md (12 AC-FF ACs in Given-When-Then format, 10 MUST + 2 SHOULD, 10 edge cases, quality gate criteria, Definition of Done)
  - progress.md (this file)
plan_tier: S
plan_loc_estimate: ~175 LOC (shell script ~80, settings.json edits ~7 each surface, hook-independence.md update ~5)
plan_files_affected: 4 (template hook + template settings.json + local hook + local settings.json) + 1 doc (hook-independence.md)
plan_runtime_changes:
  go_changes: false
  template_first_ordering: true
  additive_only: true
  fail_open_default: true
plan_audit_preconditions:
  spec_lint_clean: pending
  ears_gears_compliant: true
  exclusions_section_present: true
  frontmatter_12_fields: true
plan_auditor_iter: 0
```

## §E.2 Run-phase Evidence

Run-phase commit: M1 (pending orchestrator commit; orchestrator owns git per Hybrid Trunk Tier S main-direct).

### Files changed (Template-First order)

| # | Path | Action |
|---|------|--------|
| 1 | `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` | NEW (template, ~95 LOC shell, executable, fail-open, 5s self-terminate) |
| 2 | `internal/template/templates/.claude/settings.json.tmpl` | MODIFY — added NEW PreToolUse matcher group `Edit\|Write\|MultiEdit` → `gateguard-fact-force.sh` (5s timeout); existing `Write\|Edit\|Bash` group UNCHANGED |
| 3 | (run `make build` — regenerated embedded assets; catalog.yaml unchanged at 10325 bytes) | BUILD |
| 4 | `.claude/hooks/moai/gateguard-fact-force.sh` | NEW (local mirror, byte-identical to #1 per AC-FF-011) |
| 5 | `.claude/settings.json` | MODIFY (local mirror of #2) |
| 6 | `.claude/rules/moai/development/hook-independence.md` § 3 | MODIFY — appended Mode G row (fact-force state-file dir; single-consumer → acceptable-by-design) |
| 6m | `internal/template/templates/.claude/rules/moai/development/hook-independence.md` § 3 | MODIFY (template mirror of #6; byte-identical; SPEC-ID stripped per §25 neutrality) |

### E1. AC PASS/FAIL Matrix (12 AC-FF)

Self-contained TDD test script: `/tmp/fact-force-test.sh` (27 assertions across AC-FF-001..012).

| AC | Severity | Status | Evidence |
|----|----------|--------|----------|
| AC-FF-001 first-edit block | MUST | PASS | exit=2; stdout contains FACT-FORCE GATE / IMPORTERS / DATA SCHEMAS / USER INSTRUCTION; 1 state file written |
| AC-FF-002 second-edit allow | MUST | PASS | exit=0; no stdout; state count unchanged (1) |
| AC-FF-003 composite session+path key | MUST | PASS | different session_id → exit=2; state count 2 |
| AC-FF-004 MOAI_FACT_FORCE=off opt-out | MUST | PASS | exit=0; no stdout; 1 line in fact-force-skip.log; no state file |
| AC-FF-005 no user-prompting mechanism | MUST | PASS | canonical grep (excluding comments) returns empty on both hook + template |
| AC-FF-006 self-termination < 5s | MUST | PASS | 5-sample max wall-clock < 1000ms (target < 100ms; generous CI threshold) |
| AC-FF-007 existing group preserved | MUST | PASS | `python3 -c json.load(...)` confirms PreToolUse has 2 groups: [0]matcher=`Write\|Edit\|Bash`, [1]matcher=`Edit\|Write\|MultiEdit`; original UNCHANGED |
| AC-FF-008 no PostToolUse registration | MUST | PASS | only PreToolUse array edited; PostToolUse array untouched (additive-only) |
| AC-FF-009 subagent context (agent_id) | SHOULD | PASS | step1 parent first-edit → exit=2; step2 subagent same session → exit=0; 1 state file (parent-session keying) |
| AC-FF-010 self-loop prevention (Read) | MUST | PASS | tool_name=Read → exit=0; 0 state files |
| AC-FF-011 template/local byte-identity | MUST | PASS | `diff -q` empty; both executable (-rwxr-xr-x) |
| AC-FF-012 state file hygiene | SHOULD | PASS | stat -f '%Lp' = 600; ≤2 lines (single JSON line + trailing newline) |

**MUST: 10/10 PASS. SHOULD: 2/2 PASS. Total: 12/12 PASS.**

### E2. Cross-platform build result

```
$ make build
  ... (catalog.yaml updated successfully (10325 bytes))
  go build -ldflags "..." -o bin/moai ./cmd/moai
$ echo $?
0
```
Shell-only SPEC; no `GOOS=windows` build-tag concern (no Go source touched).

### E3. Coverage

Shell-only SPEC — Go coverage N/A. `golangci-lint` scope: no `.go` files changed. shellcheck not installed on this host (Gap — see §E.7).

### E4. Subagent-boundary grep (C-HRA-008 family)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh | grep -v '^[^:]*:[0-9]*:[ 	]*#'
(empty — exit 1)
```
Canonical comment-excluding form per `agent-common-protocol.md` § Hook Invocation Surface.

### E5. Lint status

No Go changes → no NEW golangci-lint findings possible from this SPEC. Full `go test ./...` regression: 1 pre-existing unrelated failure (`internal/statusline` `TestBuild_WritesContextUsageWithSessionID` — context_window_size=1000000 vs expected 256000; a model-band memory-state mismatch from a different SPEC; this SPEC touches zero `.go` files and `internal/statusline` has no dependency on `.claude/hooks/` or `.claude/settings.json`). All other packages `ok`, including `internal/template` and `internal/spec`.

### E6. Branch HEAD + push state

Not committed / not pushed — orchestrator owns git per Hybrid Trunk 1-person OSS Tier S main-direct. Worktree branch: `worktree-agent-a382eaf19f0bc0600`.

### E7. Gaps / Residual risk

- **shellcheck not installed** on this host (`command not found: shellcheck`) — the hook script was not statically analyzed. Residual risk: low (script is short, uses only POSIX-adjacent coreutils; tested functionally via 27 assertions). A CI environment with shellcheck SHOULD run it as a defense-in-depth layer.
- **`internal/statusline` test failure** is pre-existing and environment-specific (model-band memory state); NOT caused by this SPEC. Flagged for the orchestrator's awareness but out of this SPEC's scope.
- **State-file cleanup / GC** is explicitly out of scope (spec.md §E; deferred to a follow-up SPEC). State files accumulate per (session_id, file_path) tuple; bounded by project activity.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-07T18:15:00Z
run_commit_sha: (pending — orchestrator owns the M1 commit per Hybrid Trunk Tier S main-direct)
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 7   # handle-pre-tool.sh, handle-harness-observe.sh, handle-post-tool.sh, status-transition-ownership.sh, sync-phase-quality-gate.sh, team-ac-verify.sh, internal/hook/pre_tool.go — all untouched
l44_pre_commit_fetch: n/a (no orchestrator commit issued from this run-phase agent; orchestrator handles commit)
l44_post_push_fetch: n/a (orchestrator owns push)
new_warnings_or_lints_introduced: 0   # no Go changes; shell hook not covered by golangci-lint
cross_platform_build:
  go_changes: false
  template_first_ordering: true
  additive_only: true
  fail_open_default: true
total_run_phase_files: 6   # 2 new hook + 2 settings edits + 2 hook-independence.md doc edits
m1_to_mN_commit_strategy: single-milestone (Tier S); orchestrator commits via Route A main-direct
shellcheck_available: false   # host gap; functional TDD coverage substitutes (27 assertions)
pre_existing_unrelated_test_failure: "internal/statusline TestBuild_WritesContextUsageWithSessionID (context_window_size band mismatch; not caused by this SPEC — zero .go files touched)"
```

## §E.4 Sync-phase Audit-Ready Signal

(populated by manager-docs at sync-phase — carries the `sync_commit_sha:` field)

## §F Phase 0.95 Mode Selection

(populated by orchestrator before first run-phase `Agent()` spawn — preserves the `Mode Selection` token for the grep AC per the canonical progress.md Section Map)

## §G IGGDA Kickoff Predicate

(populated by orchestrator at the plan→run boundary IF IGGDA Path B applies — preserves the 4-condition predicate evaluation)

## §H Recursive Self-Diagnosis Log

(populated by manager-develop / orchestrator at run-phase IF a bounded recursive self-diagnosis loop fires)

## §I Notes

### §I.1 Plan-phase authoring notes (2026-07-07)

- **SPEC ID chosen**: `SPEC-HOOK-PREEDIT-INVESTIGATE-001` (user's proposal; decomposition self-check PASS — all 4 segments match `[A-Z][A-Z0-9]*`, final segment `001` matches `\d{3}`, no alpha suffix).
- **Tier classification**: Tier S (single milestone, ≤ 5 files, shell-only, additive, fail-open). User requested 4-artifact set despite Tier S convention being 2 artifacts; honored for AC traceability + V3R6 era-classification readiness.
- **Doctrinal alignment**: this SPEC is the preventive layer for CLAUDE.md §7 Rule 1 + Rule 4 + `verification-claim-integrity.md` §1.1 surfaces 1 + 2.
- **Audit-confirmed facts (re-verified at plan time)**:
  - PreToolUse broad `Edit|Write|MultiEdit` slot IS empty (`.claude/settings.json:38-49` — only one matcher group `Write|Edit|Bash` exists).
  - PostToolUse `"*"` IS occupied by `handle-harness-observe.sh` (`.claude/settings.json:67-76`).
  - The existing `agent-common-protocol.md` "Read before Edit" guideline IS a non-mechanical recommendation (no hook backs it today).
  - The ECC source pattern IS verbatim as quoted in spec.md §A.1.
- **Open questions for plan-auditor** (spec.md §F):
  - F.3 — subagent `agent_type` exemption (no — kept in scope per REQ-FF-009).
  - F.4 — `shasum` vs `sha1sum` portability (implementing agent SHOULD prefer whichever requires no external dependency beyond coreutils; alternative is `/` → `_` path substitution with no hashing).
- **Risk surface**: false-positive friction on bulk-edit scenarios (mitigated by MOAI_FACT_FORCE=off opt-out); state-file path collision on shared hosts (mitigated by 0o600 permissions + SHA-1 keying).
- **Not entered run-phase**: this is plan-phase only. The orchestrator will run plan-auditor and obtain Implementation Kickoff Approval before any run-phase entry.

### §I.2 Verification checklist (plan-phase)

- [x] SPEC ID Pre-Write Self-Check Protocol decomposition printed with `→ PASS` marker (manager-spec response body)
- [x] Directory format: `.moai/specs/SPEC-HOOK-PREEDIT-INVESTIGATE-001/`
- [x] ID uniqueness verified (no prior SPEC-HOOK-PREEDIT-* or SPEC-*-FACT-FORCE-* in `.moai/specs/`)
- [x] 4 artifacts created (spec.md + plan.md + acceptance.md + progress.md)
- [x] GEARS notation used for all 12 REQ-FF requirements
- [x] Exclusions section present with 7 `### Out of Scope — <topic>` H3 sub-headings
- [x] No implementation details (code blocks) in spec.md beyond the high-level mechanism
- [x] Frontmatter 12-canonical-field schema validated
- [x] `created` / `updated` used (NOT `created_at` / `updated_at`)
- [x] `tags` comma-separated string present (NOT `labels` YAML array)
- [x] No Go code changes specified (shell-only implementation per Tier S scope)
- [x] Template-First ordering documented in plan.md §A.3
- [x] hook-independence.md § 3 update listed in plan.md §A.3 file #6

### §I.3 Run-phase pre-flight (to be executed by manager-develop)

See plan.md §C for the 7-step pre-flight check list. The implementing agent MUST re-verify the audit baseline (`.claude/settings.json` PreToolUse / PostToolUse structure) at run time because the audit was performed at plan time and the baseline may have drifted.

### §I.4 Sync-phase notes (placeholder)

The sync-phase (manager-docs) for this SPEC SHOULD:
- Update `CHANGELOG.md` with a `feat` entry referencing the new hook
- Update `hook-independence.md` § 3 catalogue if the run-phase did not already do so
- Verify template-parity (AC-FF-011) one final time after the sync commit
- Populate `sync_commit_sha` in §E.4 above
- Transition frontmatter `status: in-progress → implemented → completed` on the single sync commit (per the 3-phase close convention)

## §F. Residual / Deferred Debt (sync-phase record, 2026-07-07)

- **spec.md §C.1 latency prose** ("<100ms" 선언 vs 실측 114-122ms) — AC-FF-006은 "<5s timeout"으로 PASS (40× 여유). prose 현실화는 후속.
- **path canonicalization** — `/tmp/x.go` vs `/tmp/./x.go`가 다른 hash key (spec edge E5/E9/E10이 명시적 연기). 후속 SPEC 소관.
- **state-file GC/cleanup** — spec.md §E 연기 항목.
- **Go-side `moai hook fact-force` handler** — spec.md §E 연기 (Tier S는 shell-only).
- **shellcheck in CI** — defense-in-depth 권장 (현재 host 미설치, TDD 27 assertions이 대체).
