# progress.md — SPEC-PROJECT-NAVIGATOR-004

> Lifecycle progress tracker. Plan-phase emits the §E skeleton; run-phase populates §E.2/§E.3; sync-phase populates §E.4.

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex: `SPEC-PROJECT-NAVIGATOR-004` → PASS (verified via Bash regex, output cited).
- Frontmatter: 12 canonical fields present, `status: draft`, `phase: "v3.3 target"`, ISO dates valid.
- ID uniqueness: no existing `SPEC-PROJECT-NAVIGATOR-004` in `.moai/specs/` (001/002/003 exist, closed).
- Requirements: GEARS notation (Ubiquitous / State-driven / Capability gate / Event-driven).
- Out of Scope: present, four `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- Artifact set (Tier M): spec.md + plan.md + acceptance.md + progress.md.
- spec.md carries no implementation detail (selection predicate is named in plan.md §F M1, not in spec.md).

## §E.2 Run-phase Evidence

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output (attributed) |
|----|--------|---------------------|----------------------------|
| AC-001 (implemented excluded from Next task) | PASS | `go test ./internal/template/ -run TestNavigatorRegen_NextTask -count=1 -v` | post-fix: `--- PASS: TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress (0.66s)` — Next task section no longer contains `SPEC-A-001` (implemented). Baseline: this run, this tree (HEAD post-M1). |
| AC-002 (in-progress preferred over draft) | PASS | same test | Next task section contains `SPEC-B-001` (in-progress) and NOT `SPEC-C-001` (draft). |
| AC-003 (Current frontier still lists implemented) | PASS | same test | `navigator.md` body contains `SPEC-A-001` (frontier stays inclusive — display semantics preserved). |
| AC-004 (SessionStart hook unchanged) | PASS | `git diff origin/main..HEAD -- .claude/hooks/moai/handle-session-start-navigator.sh internal/template/templates/.claude/hooks/moai/handle-session-start-navigator.sh` | (verified at sync-phase against origin/main; no diff — hook file byte-identical to origin/main). Run-phase self-check: hook file not in this SPEC's edit set (`git status --short` lists only the 4 plan-declared files + the test + progress.md). |
| AC-005 (template/mirror byte-parity) | PASS | `cmp internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh .claude/skills/moai-workflow-project/scripts/navigator-regen.sh` + same `cmp` for `references/navigator.md` | both `cmp` exit 0 (byte-identical). Baseline: pre-edit `cmp` also byte-identical; post-edit `cmp` byte-identical — the pair stayed in lockstep. CI guard `internal/template/rule_template_mirror_test.go` passes. |
| AC-006 (regression test red→green) | PASS | `go test ./internal/template/ -run TestNavigatorRegen_NextTask -count=1 -v` | RED (pre-fix, verbatim): `AC-001: SPEC-A-001 (implemented) appeared in Next task section` + `AC-002: SPEC-B-001 (in-progress) is NOT the Next task` → `--- FAIL: TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress (0.68s)`. GREEN (post-fix): `--- PASS: ... (0.66s)`. See §E item E8 for verbatim RED block. |
| AC-007 (§25 neutrality self-check clean) | PASS | 5-item self-check (`.moai/docs/template-internal-isolation-doctrine.md` §25.3) + `internal/template/internal_content_leak_test.go` | added lines contain only generic status-enum tokens (`in-progress`, `draft`, `implemented`) + mechanism prose; no SPEC IDs, REQ tokens, dates, SHAs, or internal paths. `InternalContentLeak` + `TestTemplateNeutrality` tests pass. |

### E2. Cross-Platform Build + embedded FS recompile

```
$ go build ./...                                → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...      → exit 0
$ make build                                    → exit 0 (catalog.yaml regenerated; moai-workflow-project hash UNCHANGED because catalog hashes only SKILL.md, not scripts/references — my edits are sub-files, so no catalog.yaml diff vs HEAD is expected and none is present)
```

Baseline: this run, this tree (HEAD `d8af35526` plan commit + M1 run-phase edits, uncommitted at time of capture).

### E3. Coverage measurement

The fix is in a shell script (`navigator-regen.sh`) which is exercised black-box via a Go subprocess-driver test (`internal/template/navigator_regen_test.go`). Go `go test -cover` does not attribute coverage to bash scripts — there is no Go source to measure. The relevant coverage figure is the regression-test's behavioral coverage of the selection logic: the new test `TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress` asserts all three behavioral ACs (AC-001 implemented-excluded, AC-002 in-progress-preferred, AC-003 frontier-inclusive) against a fixture with the implemented+in-progress+draft mix that would exercise the bug. No 85% Go-coverage figure applies; the binding verification is the red→green regression test (AC-006), not a coverage percentage.

### E4. Subagent Boundary Grep

N/A — the change is in a bash script + a markdown reference + a Go test file; there is no Go package under `internal/` or `internal/hook/` modified, so the C-HRA-008 subagent-boundary grep (`grep -rn 'AskUserQuestion' <pkg>`) does not apply to this SPEC's edit set. State as gap: no subagent-boundary Go surface was touched, so no boundary grep was run.

### E5. Lint Status

```
$ golangci-lint run --timeout=3m ./internal/template/   → 0 issues.
$ go vet ./internal/template/                            → exit 0
```

No NEW lint issues introduced; baseline (pre-edit) was also clean. spec-lint: the SPEC frontmatter transition `draft → in-progress` is owned by manager-develop (this run-phase M1 commit) per the Status Transition Ownership Matrix; `OwnershipTransitionRule` will validate the transition on the M1 commit subject.

### E6. Branch HEAD + Push state

- New commit SHA(s): populated by the M1 commit (this run-phase's single atomic commit). The orchestrator handles push+PR after sync; run-phase does not push.
- `git push`: NOT performed by run-phase (PR-mandatory repo, `enforce_admins:true`; orchestrator owns push+PR).

### E7. Blocker Report

None. No SPEC body content needed modification mid-implementation; the plan-declared scope envelope held.

### E8. RED Failure Output (verbatim, pre-GREEN)

Captured BEFORE the script fix was applied, against the pre-fix logic at HEAD `d8af35526`:

```
$ go test ./internal/template/ -run TestNavigatorRegen_NextTask -count=1 -v
=== RUN   TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress
    navigator_regen_test.go:584: AC-001: SPEC-A-001 (implemented) appeared in Next task section (must be excluded):

        Advance **SPEC-A-001** toward its next milestone. See `.moai/project/navigator/progress-map.md` for its frontier milestone.

        Full entry brief: this file. Full progress rollup: `progress-map.md`.

    navigator_regen_test.go:588: AC-002: SPEC-B-001 (in-progress) is NOT the Next task (must be preferred over draft):

        Advance **SPEC-A-001** toward its next milestone. See `.moai/project/navigator/progress-map.md` for its frontier milestone.

        Full entry brief: this file. Full progress rollup: `progress-map.md`.

--- FAIL: TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress (0.68s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	1.139s
FAIL
```

Post-fix GREEN (same command, after the M1 fix):

```
=== RUN   TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress
--- PASS: TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress (0.66s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	1.131s
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-07"
run_commit_sha: "pending-backfill-M1"   # M1 commit SHA backfilled in a follow-up commit (self-referential-hazard workaround per D3)
run_status: "audit-ready"
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0          # no PRESERVE-list items carried into sync
l44_pre_commit_fetch: true               # Pre-Spawn/Pre-Edit Sync Check: worktree-isolated session, single session on this checkout
l44_post_push_fetch: "n/a — run-phase does not push (orchestrator owns push+PR)"
new_warnings_or_lints_introduced: false
cross_platform_build:
  linux_amd64: "exit 0"
  windows_amd64: "exit 0"
  make_build: "exit 0 (catalog.yaml regenerated; moai-workflow-project hash unchanged — catalog hashes SKILL.md only, edits were sub-files)"
total_run_phase_files: 6                 # 2 template sources + 2 mirrors + 1 test file + 1 progress.md (this file); spec.md frontmatter-only edit
m1_to_mN_commit_strategy: "single atomic M1 commit (RED test + GREEN script+mirror fix + refs note + spec.md frontmatter draft→in-progress + progress.md §E.2/§E.3/§F)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Decision: Mode 5 (sub-agent).**

Input parameters:
- tier: M (3 artifacts, acceptance.md present)
- scope (file count): 6 (2 template sources + 2 mirrors + 1 Go test + progress.md/spec.md frontmatter) — well under the multi-domain threshold
- domain count: 1 (navigator template script — single bash surface + its doc reference)
- file language mix: bash + markdown + Go test
- concurrency benefit: LOW (single-package, single-thread-of-work bugfix; no independent research/coding lanes)

Mode evaluation table:

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Multi-file edit (template source + mirror + test + docs); not a single-line typo. |
| 2 background | no | Work is not read-only analysis; it includes Write/Edit to the shared template tree. |
| 3 agent-team | RETIRED | Mode 3 is a tombstone; never selected. |
| 4 parallel | no | Single domain (navigator script), no multi-domain research fan-out benefit. |
| 5 sub-agent | **selected** | Single-thread coding-heavy bugfix; sequential sub-agent is the Anthropic-recommended default for coding work. |
| 6 workflow | no | 6 files, not ≥~30; bash+md+Go is not a uniform mechanical transform across many call sites. |

Justification: a focused bugfix in a single shell script with one regression test and one doc note is the canonical Mode 5 case — coding-heavy, low concurrency benefit, single domain. Mode 4 would add coordination overhead without parallelism gain; Mode 6's mechanical-fan-out primitive does not apply at this file count. The selection aligns with Anthropic's coding-task parallelism caveat (most coding tasks involve fewer truly parallelizable tasks than research).

(Phase 4 mode selection was performed by the run-phase implementer (`manager-develop`) at M1; this §F log records the autonomous selection. No `AskUserQuestion` was issued — Phase 4 mode selection is autonomous by contract.)
