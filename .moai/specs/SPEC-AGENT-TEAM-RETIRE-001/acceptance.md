# Acceptance Criteria — SPEC-AGENT-TEAM-RETIRE-001

> SSOT for the AC matrix. 28 ACs covering all 22 REQs (100% AC→REQ coverage).
> Verification conventions: (1) every removal check is **anchored** (prefix or
> word boundary) — bare `team` substring greps are invalid evidence;
> (2) preservation ACs assert STILL-EXISTS, never absence; (3) counts follow the
> baseline-delta discipline — pre-flight measures the baseline, exit re-measures,
> and the AC asserts the delta, not a plan-time hardcoded figure; (4) `go test`
> ACs must show the test actually ran (`--- PASS` / `ok` line), not
> "no tests to run" (vacuous-green guard).

## §A. Given-When-Then Scenarios

### GWT-1 — Static layer removal keeps the repo green

- **Given** the Phase 0 migration has landed (lockfile + taskledger, whole repo
  green),
- **When** the team_spawn files and the TeamConfig type family are deleted and
  compile-coupled tests reconciled,
- **Then** `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and
  `go test ./...` all exit 0, and `TestClaimTaskAppend_Repro` still runs and
  passes from its migrated location.

### GWT-2 — sync-audit workflow refuses partial verdicts

- **Given** `sync-audit-4dim.js` runs with a judge agent returning null (e.g.
  rate-limited),
- **When** the Verdict phase executes,
- **Then** the workflow returns verdict `INCOMPLETE` naming the missing
  dimension(s), and no harmonic mean is computed over the remaining three
  scores.

### GWT-3 — Fresh deploy is resurrection-negative

- **Given** the post-removal binary built via `make build`,
- **When** `moai init <sandbox>` deploys a fresh project,
- **Then** the deployed tree contains no `team-protocol.md`, no
  `team-pattern-cookbook.md`, no `.claude/skills/moai/team/` directory, and no
  `workflow.yaml` `team:` block.

### GWT-4 — plan-research aborts on insufficient coverage

- **Given** `plan-research-fanout.js` runs and two of four lens agents return
  null,
- **When** the Explore phase completes,
- **Then** the Synthesize phase is skipped and the workflow returns
  `insufficient_coverage` naming the failed lenses.

## §B. AC ↔ REQ Mapping

| AC | REQ | Title |
|----|-----|-------|
| AC-ATR-001 | REQ-ATR-001 | lockfile package exists with build-tag pair |
| AC-ATR-002 | REQ-ATR-001 | Windows cross-build green |
| AC-ATR-003 | REQ-ATR-002 | CLIFIX repro test green from migrated symbols |
| AC-ATR-004 | REQ-ATR-002 | ClaimTask coverage tests migrated + green |
| AC-ATR-005 | REQ-ATR-003 | Migration-before-deletion commit ordering |
| AC-ATR-006 | REQ-ATR-004 | team_spawn file family absent |
| AC-ATR-007 | REQ-ATR-005 | TeamConfig type family absent |
| AC-ATR-008 | REQ-ATR-005 | WorkflowConfig.Team field + defaults block absent |
| AC-ATR-009 | REQ-ATR-006 | PRESERVE: config RoleProfile (Sandbox) exists |
| AC-ATR-010 | REQ-ATR-006 | PRESERVE: GitStrategy Team ModeProfile + tests |
| AC-ATR-011 | REQ-ATR-006 | PRESERVE: session merge + worktree team_launch |
| AC-ATR-012 | REQ-ATR-007 | PRESERVE: f.git_strategy.team.* i18n keys |
| AC-ATR-013 | REQ-ATR-008 | Web console team surfaces removed |
| AC-ATR-014 | REQ-ATR-009 | workflow lint repurposed, not no-op |
| AC-ATR-015 | REQ-ATR-010 | Team rules absent; Mode 3 tombstone; no dangling refs |
| AC-ATR-016 | REQ-ATR-011 | Team skills dir absent; workflow routing cleaned |
| AC-ATR-017 | REQ-ATR-012 | Template mirror + CI guards green |
| AC-ATR-018 | REQ-ATR-013 | moai init resurrection-negative |
| AC-ATR-019 | REQ-ATR-014 | Whole-repo green trio |
| AC-ATR-020 | REQ-ATR-015 | sync-audit structure + in-script harmonic mean |
| AC-ATR-021 | REQ-ATR-016 | sync-audit INCOMPLETE + prohibitions |
| AC-ATR-022 | REQ-ATR-017 | sync-audit Tier M/L gate |
| AC-ATR-023 | REQ-ATR-018 | plan-research structure (lenses, markdown, synthesize) |
| AC-ATR-024 | REQ-ATR-019 | plan-research insufficient_coverage abort |
| AC-ATR-025 | REQ-ATR-020 | plan-research prohibitions (≤4, no xhigh, no writes) |
| AC-ATR-026 | REQ-ATR-021 | House style + determinism both scripts |
| AC-ATR-027 | REQ-ATR-022 | CG Mode guidance migrated before glm.md deletion |
| AC-ATR-028 | REQ-ATR-010 | No dangling workflow.yaml auto_selection pointer (prose-only SSOT) |

## §C. Verification Commands (per AC)

### AC-ATR-001 (REQ-ATR-001)

```bash
ls internal/lockfile/            # expect: lock files incl. unix + windows variants + test
grep -l 'go:build !windows' internal/lockfile/*.go && grep -l 'go:build windows' internal/lockfile/*.go
grep -rn "cross-process" internal/lockfile/   # Windows limitation comment preserved
```

Expected: both build-tag variants present; limitation comment retained.

### AC-ATR-002 (REQ-ATR-001)

```bash
GOOS=windows GOARCH=amd64 go build ./... ; echo "exit=$?"   # expect exit=0
```

### AC-ATR-003 (REQ-ATR-002)

```bash
go test -v -run 'TestClaimTaskAppend_Repro' ./internal/cli/... 2>&1 | grep -E '^(=== RUN|--- PASS|ok )'
```

Expected: `=== RUN TestClaimTaskAppend_Repro` + `--- PASS` present (vacuous-green
guard: an empty run FAILS this AC).

### AC-ATR-004 (REQ-ATR-002)

```bash
go test -v -run 'TestClaimTask' ./internal/cli/taskledger/ 2>&1 | grep -c -- '--- PASS'
```

Expected: ≥ 2 (`TestClaimTask` + `TestClaimTaskConcurrent`).

### AC-ATR-005 (REQ-ATR-003)

```bash
git log --oneline --grep='SPEC-AGENT-TEAM-RETIRE-001' | tail -5
```

Expected: the M0 migration commit SHA precedes (is an ancestor of) the first
M1 deletion commit; progress.md §E.2 records `go test ./...` exit 0 at M0 exit.

### AC-ATR-006 (REQ-ATR-004)

```bash
ls internal/cli/team_spawn* 2>&1 ; echo "exit=$?"
```

Expected: "no matches" / non-zero ls exit (all five files gone).

### AC-ATR-007 (REQ-ATR-005)

```bash
grep -rn -E "TeamConfig|RoleProfileEntry|TeamAutoSelectionConfig" internal/config/ --include="*.go" | wc -l
```

Expected: `0`.

### AC-ATR-008 (REQ-ATR-005)

```bash
grep -n 'yaml:"team"' internal/config/types.go | wc -l     # expect 0 (WorkflowConfig.Team gone)
grep -n 'Team: TeamConfig' internal/config/defaults.go | wc -l   # expect 0
```

Note: the GitStrategy `Team ModeProfile` field uses the same yaml key inside a
DIFFERENT struct — if the first grep returns 1, verify by context that the
survivor is the GitStrategy field (PRESERVE) and record the disambiguation in
§E.2 evidence; the TeamConfig-typed field must be gone.

### AC-ATR-009 (REQ-ATR-006) — preservation

```bash
grep -c "type RoleProfile struct" internal/config/types.go   # expect 1
grep -n "Sandbox string" internal/config/types.go | wc -l    # expect >= 1
```

### AC-ATR-010 (REQ-ATR-006) — preservation

```bash
grep -n "Team     ModeProfile" internal/config/types.go | wc -l   # expect 1 (GitStrategy)
go test -run 'Defaults' ./internal/config/ 2>&1 | tail -2         # expect ok (AC-GSS-005 rows intact)
grep -c "Team.Automation.AutoPush" internal/config/defaults_test.go   # expect >= 1
```

### AC-ATR-011 (REQ-ATR-006) — preservation

```bash
grep -c "func (fs \*FileSessionStore) MergeTeamCheckpoints" internal/session/store.go  # expect 1
ls internal/cli/worktree/team_launch* internal/cli/worktree/swarm_registry.go internal/cli/worktree/handoff_guidance.go
grep -rn "teammateMode" internal/cli/glm.go internal/cli/launcher.go | wc -l   # expect >= 1
```

### AC-ATR-012 (REQ-ATR-007) — preservation

```bash
grep -c '"f\.git_strategy\.team\.' internal/web/assets/i18n.js   # expect == pre-flight baseline (unchanged)
grep -c '"f\.git_strategy\.mode\.opt\.team"' internal/web/assets/i18n.js   # expect >= 1
```

### AC-ATR-013 (REQ-ATR-008)

```bash
grep -c '"f\.workflow\.team\.' internal/web/assets/i18n.js       # expect 0
grep -rn "RoleProfileNames" internal/ --include="*.go" | wc -l   # expect 0
grep -in "team" internal/web/fieldsets.templ | wc -l             # expect 0 team fieldset markup (manual review of any survivor)
git diff --stat --exit-code internal/web/fieldsets_templ.go && echo REGEN-CLEAN   # after templ generate: no drift
```

### AC-ATR-014 (REQ-ATR-009)

```bash
grep -c "role_profiles" internal/cli/agentlint/workflow_lint.go        # expect 0
grep -c -i "ModelRouting" internal/cli/agentlint/workflow_lint.go      # expect >= 1
go test ./internal/cli/agentlint/... 2>&1 | tail -2                    # expect ok — incl. a violation-path test asserting non-zero lint result
```

### AC-ATR-015 (REQ-ATR-010)

```bash
for t in .claude internal/template/templates/.claude; do
  for f in rules/moai/workflow/team-protocol.md rules/moai/workflow/team-pattern-cookbook.md; do
    [ -e "$t/$f" ] && echo "RESIDUE: $t/$f"
  done
done   # expect: no RESIDUE lines (per-tree independent check)
grep -c -i "retire" .claude/rules/moai/workflow/orchestration-mode-selection.md   # expect >= 1 (Mode 3 tombstone)
grep -rn -E "team-protocol\.md|team-pattern-cookbook\.md|skills/moai/team/" .claude/ internal/template/templates/.claude/ --include="*.md" | wc -l   # expect 0 dangling refs (this SPEC's own artifacts exempt)
```

### AC-ATR-016 (REQ-ATR-011)

```bash
for t in .claude internal/template/templates/.claude; do
  [ -d "$t/skills/moai/team" ] && echo "RESIDUE: $t/skills/moai/team"
done   # expect: no RESIDUE lines
grep -c "team/run.md" .claude/skills/moai/workflows/run.md   # expect 0 (repeat per workflows file with anchored patterns measured at pre-flight)
```

### AC-ATR-017 (REQ-ATR-012)

```bash
make build ; echo "exit=$?"                                        # expect 0
go test ./internal/template/... 2>&1 | tail -2                     # expect ok (neutrality + leak guards)
grep -n "^    team:" internal/template/templates/.moai/config/sections/workflow.yaml | wc -l   # expect 0
grep -n "role_profiles:" .moai/config/sections/workflow.yaml | wc -l   # expect 0 (local tree too)
```

### AC-ATR-018 (REQ-ATR-013)

```bash
T=$(mktemp -d) && ./bin/moai init "$T/sandbox" >/dev/null 2>&1
[ -e "$T/sandbox/.claude/rules/moai/workflow/team-protocol.md" ] && echo RESIDUE
[ -d "$T/sandbox/.claude/skills/moai/team" ] && echo RESIDUE
grep -n "^    team:" "$T/sandbox/.moai/config/sections/workflow.yaml" | wc -l   # expect 0
```

Expected: no RESIDUE lines; count 0.

### AC-ATR-019 (REQ-ATR-014)

```bash
go build ./... ; echo "exit=$?"                          # expect 0
GOOS=windows GOARCH=amd64 go build ./... ; echo "exit=$?" # expect 0
go test ./... 2>&1 | tail -3                             # expect 0 failures
grep -rn "workflow_role_profiles_test" internal/ | wc -l  # expect 0 (file deleted)
```

### AC-ATR-020 (REQ-ATR-015)

```bash
node --check .claude/workflows/sync-audit-4dim.js ; echo "exit=$?"   # expect 0
grep -c "title:" .claude/workflows/sync-audit-4dim.js                # expect 3 (meta.phases: Context/Judge/Verdict)
grep -c -E "Functionality|Security|Craft|Consistency" .claude/workflows/sync-audit-4dim.js   # expect >= 4
grep -n "0.85" .claude/workflows/sync-audit-4dim.js | wc -l          # expect >= 1 (args threshold default)
grep -c "1 /" .claude/workflows/sync-audit-4dim.js                   # expect >= 1 (in-script reciprocal sum — harmonic mean)
```

Plus structural review: the harmonic mean and threshold comparison appear in
script JS between the Judge results and the return value, with a zero-score
guard branch; no agent call sits between judge collection and verdict return.

### AC-ATR-021 (REQ-ATR-016)

```bash
grep -c "INCOMPLETE" .claude/workflows/sync-audit-4dim.js   # expect >= 1
grep -c -i "meta-judge" .claude/workflows/sync-audit-4dim.js  # expect >= 1 (anti-pattern named in header, no meta-judge agent call)
grep -c "agentType: 'Explore'" .claude/workflows/sync-audit-4dim.js   # expect >= 1 (read-only enforcement noted for judges per design.md §C)
```

Plus structural review: the null-judge branch returns INCOMPLETE naming the
missing dimension(s) BEFORE any mean computation; judge agent options carry no
write-capable agentType.

### AC-ATR-022 (REQ-ATR-017)

```bash
grep -c "args.tier" .claude/workflows/sync-audit-4dim.js    # expect >= 1
grep -c -i "Tier S" .claude/workflows/sync-audit-4dim.js    # expect >= 1 (gate documented in header)
```

### AC-ATR-023 (REQ-ATR-018)

```bash
node --check .claude/workflows/plan-research-fanout.js ; echo "exit=$?"  # expect 0
grep -c -E "codebase-precedent|external-docs|constraints-risks|prior-SPEC-memory" .claude/workflows/plan-research-fanout.js  # expect >= 4
grep -c "confidence_and_gaps" .claude/workflows/plan-research-fanout.js  # expect >= 1
grep -c "effort: 'high'" .claude/workflows/plan-research-fanout.js       # expect >= 1 (synthesizer)
grep -c "effort: 'medium'" .claude/workflows/plan-research-fanout.js     # expect >= 1 (explorers)
grep -c "NONE found" .claude/workflows/plan-research-fanout.js           # expect >= 1
```

### AC-ATR-024 (REQ-ATR-019)

```bash
grep -c "insufficient_coverage" .claude/workflows/plan-research-fanout.js  # expect >= 1
grep -c -E ">= 2|>=2" .claude/workflows/plan-research-fanout.js            # expect >= 1 (null-lens abort threshold)
```

### AC-ATR-025 (REQ-ATR-020)

```bash
grep -c "slice(0, 4)" .claude/workflows/plan-research-fanout.js   # expect >= 1 (lens cap; or equivalent length guard — record the actual guard)
grep -c "xhigh" .claude/workflows/plan-research-fanout.js         # expect 0 on agent() opts (any hit must be a prose prohibition in the header, verified by context)
grep -c -i "research.md is written" .claude/workflows/plan-research-fanout.js   # expect >= 1 (no in-workflow writes doctrine in header)
```

### AC-ATR-026 (REQ-ATR-021)

```bash
grep -c "export const meta" .claude/workflows/sync-audit-4dim.js .claude/workflows/plan-research-fanout.js   # expect 1 each
grep -n -E "Date\.now\(|Math\.random\(" .claude/workflows/sync-audit-4dim.js .claude/workflows/plan-research-fanout.js | wc -l   # expect 0 call sites
grep -c "label: " .claude/workflows/sync-audit-4dim.js    # expect >= 2 ('<stage>:<item>' labels)
grep -c "label: " .claude/workflows/plan-research-fanout.js  # expect >= 2
```

### AC-ATR-027 (REQ-ATR-022)

```bash
grep -c "CG Mode (Claude + GLM" .claude/rules/moai/core/glm-web-tooling.md   # expect >= 1 (baseline 0 at plan time — non-vacuous)
grep -c "CG Mode (Claude + GLM" internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md   # expect >= 1 (mirror parity)
[ -e .claude/skills/moai/team/glm.md ] && echo RESIDUE   # expect: no RESIDUE (deleted AFTER migration)
```

Plus ordering evidence: the migration edit lands in the same M3 commit as (or an
earlier commit than) the team-skills deletion — recorded in progress.md §E.2
(migrate-then-delete per REQ-ATR-022; deleting first is a FAIL even if the
content is re-added later).

### AC-ATR-028 (REQ-ATR-010, D8 adopted)

```bash
grep -c "auto_selection" .claude/rules/moai/workflow/orchestration-mode-selection.md   # expect 0 (machine-readable pointer sentence removed)
grep -c "auto_selection" internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md   # expect 0 (mirror parity)
grep -c "≥ 3 domains" .claude/rules/moai/workflow/orchestration-mode-selection.md   # expect >= 1 (§B.1 prose thresholds RETAINED as sole SSOT)
```

Note: the `spec-workflow.md` "See workflow.yaml team.auto_selection" reference
is removed by the REQ-ATR-010 team-section cleanup and verified by the
AC-ATR-015 dangling-ref sweep; this AC binds the orchestration-mode-selection.md
surface specifically.

## §D. Edge Cases

- **E1 Windows lock semantics**: `internal/lockfile` Windows variant is
  in-process only (documented limitation) — the migration must not silently
  "upgrade" it to LockFileEx; behavior preservation is the contract.
- **E2 yaml:"team" key collision**: the GitStrategy ModeProfile `Team` field
  shares the yaml key name inside a different struct — AC-ATR-008 requires
  context disambiguation, not a blind zero-count.
- **E3 i18n locale partitions**: if i18n.js carries per-locale blocks, the
  `f.workflow.team.*` removal must cover every locale block symmetrically.
- **E4 sync-audit all-zero score**: a judge returning score 0 must trip the
  zero-score guard (harmonic mean divides by sᵢ) — verdict is a FAIL naming
  the zero-scored dimension, never a division error or Infinity.
- **E5 plan-research exactly-1 null lens**: synthesis proceeds with 3 lenses;
  the null lens is named in `per_lens_reports` as null and in the synthesis
  prompt as a coverage gap (only ≥2 nulls abort).
- **E6 templ regeneration drift**: `templ generate` must be run with the
  project-pinned templ version; a version-drift regen diff outside the team
  fieldset removal is a blocker, not a commit.

## §E. Quality Gates

- TRUST 5: Tested (repro + migrated tests green; new packages ≥85% coverage);
  Readable (migrated code keeps godoc); Unified (gofmt/golangci-lint NEW-issue
  free vs baseline); Secured (no new input surfaces; lock semantics preserved);
  Trackable (Conventional Commits per milestone, `🗿 MoAI` trailer).
- LSP/lint: run-phase zero NEW errors; sync-phase clean per spec-workflow.md.
- CI: spec-lint, golangci-lint, per-OS tests, template-neutrality-check all
  green on the final push.

## §F. Definition of Done

1. All 28 ACs PASS with verbatim command output recorded in progress.md §E.2
   (PASS-WITH-DEBT permitted only with named debt + follow-up owner).
2. Preservation ACs (009-012) confirmed AFTER the final removal commit.
3. Both workflow scripts pass `node --check` and structural review.
4. Whole-repo green trio + template CI guards green.
5. Sync-phase close by manager-docs per the Status Transition Ownership Matrix.
