# acceptance.md — SPEC-GOAL-ENGINE-001

> Each AC is a single discriminating check. Router-registration ACs (P1 list vs
> Quick Reference) are pinned SEPARATELY from the workflow-file AC. Go ACs cite
> real test/coverage output. Template-mirror ACs use per-file checks.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Baseline → Post |
|----|-----|----------|-----------------|
| AC-GLE-001 | REQ-GLE-001 | goal.md workflow file exists + 4 verbs | absent → present |
| AC-GLE-002 | REQ-GLE-002 | Priority 1 block registers `**goal**` | 0 → ≥1 |
| AC-GLE-003 | REQ-GLE-003 | Quick Reference has `### goal` heading | 0 → ≥1 |
| AC-GLE-004 | REQ-GLE-004 | per-session state path (no shared file) | test PASS |
| AC-GLE-005 | REQ-GLE-005 | schema carries all fields incl ceiling=30 | test PASS |
| AC-GLE-006 | REQ-GLE-006 | atomic write (temp+rename) | test PASS |
| AC-GLE-007 | REQ-GLE-007 | orphan prune → consumed/ | test PASS |
| AC-GLE-008 | REQ-GLE-008 | writer_pid fallback | test PASS |
| AC-GLE-009 | REQ-GLE-009 | `moai hook stop-goal` verb exists | exit-0 smoke |
| AC-GLE-010 | REQ-GLE-010 | Tier-1 fail → exit0 block JSON | test PASS |
| AC-GLE-011 | REQ-GLE-011 | Tier-2 only after mechanical pass | test PASS |
| AC-GLE-012 | REQ-GLE-012 | all pass → no block | test PASS |
| AC-GLE-013 | REQ-GLE-013 | ceiling → 5-section verdict, no block | test PASS |
| AC-GLE-014 | REQ-GLE-014 | no AskUserQuestion in hook/goal code | grep 0 |
| AC-GLE-015 | REQ-GLE-015 | armed goal does not bypass Kickoff | doc + test |
| AC-GLE-016 | REQ-GLE-016 | native /goal active → yield | test PASS |
| AC-GLE-017 | REQ-GLE-017 | stagnation → stop + escalation note | test PASS |
| AC-GLE-018 | REQ-GLE-018 | handoff Block 5 goal variant documented | present |
| AC-GLE-019 | REQ-GLE-019 | goal-directive.md /moai goal row + Axis B | present |
| AC-GLE-020 | REQ-GLE-020 | native-invocation Axis B updated | present |
| AC-GLE-021 | REQ-GLE-021 | §2 stage ⑤ references goal evaluator | present |
| AC-GLE-022 | REQ-GLE-022 | distinctness guard stays green | test PASS |
| AC-GLE-023 | REQ-GLE-023 | internal/goal layout present | files exist |
| AC-GLE-024 | REQ-GLE-024 | internal/goal ≥85% coverage | cover output |
| AC-GLE-025 | REQ-GLE-025 | mirrors + neutral + make build | per-file + exit 0 |
| AC-GLE-026 | REQ-GLE-023 | Stop-hook COMPOSE (add-not-replace) | existing≥1 preserved + new≥1 |
| AC-GLE-027 | REQ-GLE-026 | goal.md documents progression-mode axis at kickoff | 0 → ≥1 |
| AC-GLE-028 | REQ-GLE-027 | autonomous mode continues, no checkpoint (schema+eval) | test PASS |
| AC-GLE-029 | REQ-GLE-028 | semi-autonomous checkpoint JSON emitted by hook | test PASS |
| AC-GLE-030 | REQ-GLE-028 | orchestrator confirm path documented (bridge) | 0 → ≥1 |
| AC-GLE-031 | REQ-GLE-029 | CLAUDE.md progression-mode axis documented | 0 → ≥1 |
| AC-GLE-032 | REQ-GLE-029 | run.md progression-mode axis documented | 0 → ≥1 |
| AC-GLE-033 | REQ-GLE-029 | orchestration-mode-selection.md axis documented | 0 → ≥1 |
| AC-GLE-034 | REQ-GLE-026,029 | kickoff mandatory in BOTH modes (grep + Go test) | 0→≥1 + test PASS |
| AC-GLE-035 | REQ-GLE-030,031 | `moai goal` CLI registered + lists arm/status/clear | unknown-cmd → exit-0 + 3 verbs |
| AC-GLE-036 | REQ-GLE-030,032,033 | E2E arm→eval linkage (parse cond; arm writes file; hook loads SAME id) | test PASS (make-or-break) |
| AC-GLE-037 | REQ-GLE-033 | session-id consistency (arm writes `<id>.json`, NOT `pid-*`) | test PASS |
| AC-GLE-038 | REQ-GLE-034 | PruneOrphans wired on session-start + fail-open | grep 0 → ≥1 + test PASS |
| AC-GLE-039 | REQ-GLE-031 | resume NOT delivered (out of scope) | CLI lacks resume + goal.md deferred |

### AC-GLE-001 — goal.md workflow file + 4 verbs

```bash
test -f .claude/skills/moai/workflows/goal.md && echo OK
grep -ci "register\|arm" .claude/skills/moai/workflows/goal.md   # the register/arm verb (bare "<condition>" form)
grep -c "status\|clear\|resume" .claude/skills/moai/workflows/goal.md   # the other 3 verbs
```
PASS when the file exists AND documents ALL 4 verbs: the register/arm form (the
bare `"<condition>"` argument form), `status`, `clear`, `resume`.

### AC-GLE-002 — P1 registration (anchored to the Priority 1 block)

```bash
# Anchor to the Priority 1 subcommand-matching block (a correct impl adds `**goal**` there).
awk '/^### Priority 1/,/^### Priority 2/' .claude/skills/moai/SKILL.md | grep -c '\*\*goal\*\*'   # expect ≥1
```
Baseline (verified this iteration): 0 inside the Priority 1 block. PASS when
`**goal**` appears in the Priority 1 explicit-subcommand block. (SEPARATE from
AC-GLE-003 — both surfaces must register, or the feature is inert per the
reachability lesson.)

### AC-GLE-003 — Quick Reference entry (anchored to the section, `### goal` heading)

```bash
# Anchor to the Workflow Quick Reference section; existing entries use `### <name>`
# headings (`### plan`, `### run`, `### sync`), so the new entry is `### goal`.
awk '/^## Workflow Quick Reference/,0' .claude/skills/moai/SKILL.md | grep -c '^### goal'   # expect ≥1
```
Baseline (verified this iteration): 0. PASS when the Quick Reference carries a
`### goal` heading entry. (SEPARATE from AC-GLE-002 — a `/moai goal` mention
anywhere is not enough; the actual section entry must exist.)

### AC-GLE-004 — per-session state path

```bash
go test ./internal/goal/ -run TestStatePathPerSession -v 2>&1 | tail -5
```
PASS when the test asserts the path is `.moai/state/goal/<session-id>.json` and
that NO single shared filename is used.

### AC-GLE-005 — schema completeness (ceiling default 30)

```bash
go test ./internal/goal/ -run TestSchemaFields -v 2>&1 | tail -5
```
PASS when the test asserts goal text, `conditions[]` (mechanical|model), `ceiling.max_turns==30` default, append-only `progress`, `session_id`, `created_at`, `status`.

### AC-GLE-006 — atomic write

```bash
go test ./internal/goal/ -run TestAtomicWrite -v 2>&1 | tail -5
```
PASS when the test asserts a temp file + rename sequence (no partial in-place write).

### AC-GLE-007 — orphan prune

```bash
go test ./internal/goal/ -run TestOrphanPrune -v 2>&1 | tail -5
```
PASS when a state file absent from `active-sessions.json` (or TTL-expired) is moved
to `.moai/state/goal/consumed/`.

### AC-GLE-008 — writer_pid fallback

```bash
go test ./internal/goal/ -run TestWriterPidFallback -v 2>&1 | tail -5
```
PASS when, given no session id, the engine keys state on `writer_pid`.

### AC-GLE-009 — hook verb smoke

```bash
echo '{}' | go run ./cmd/moai hook stop-goal ; echo "exit=$?"
```
PASS when the verb exists and exits 0 on empty/no-goal input (no crash).

### AC-GLE-010 — Tier-1 mechanical fail → block JSON

```bash
go test ./internal/goal/ -run TestTier1Block -v 2>&1 | tail -5
```
PASS when a failing mechanical condition yields exit 0 + stdout
`{"decision":"block","reason":...}` containing the failed condition + output tail.

### AC-GLE-011 — Tier-2 gated by mechanical pass

```bash
go test ./internal/goal/ -run TestTier2Gate -v 2>&1 | tail -5
```
PASS when Tier-2 model judgment runs ONLY after all mechanical conditions pass AND
at least one model condition exists.

### AC-GLE-012 — all-pass → no block

```bash
go test ./internal/goal/ -run TestAllPassNoBlock -v 2>&1 | tail -5
```
PASS when, with all conditions satisfied, no block decision is emitted and status
becomes `satisfied`.

### AC-GLE-013 — ceiling → 5-section verdict

```bash
go test ./internal/goal/ -run TestCeilingVerdict -v 2>&1 | tail -5
```
PASS when reaching `max_turns` emits a verdict with the 5 section names
(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) and stops blocking.

### AC-GLE-014 — hook boundary (no AskUserQuestion)

```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/goal/ internal/cli/hook_stop_goal.go .claude/hooks/moai/handle-stop-goal.sh \
  | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | wc -l   # expect 0
```
PASS when the count is 0.

### AC-GLE-015 — armed goal does not bypass Kickoff

```bash
grep -in "Implementation Kickoff Approval\|does not.*bypass\|never.*bypass" .claude/skills/moai/workflows/goal.md | head
go test ./internal/goal/ -run TestNoKickoffBypass -v 2>&1 | tail -5
```
PASS when goal.md states the armed goal never authorizes run-phase entry AND the
engine carries no auto-run-authorization path.

### AC-GLE-016 — native /goal active → yield

```bash
go test ./internal/goal/ -run TestNativeGoalYield -v 2>&1 | tail -5
```
PASS when, given an active-native-`/goal` signal, `stop-goal` yields (no double-block).

### AC-GLE-017 — stagnation → stop + escalation

```bash
go test ./internal/goal/ -run TestStagnationStop -v 2>&1 | tail -5
```
PASS when N no-progress iterations trigger a stop + an E1/E3 escalation note in the verdict.

### AC-GLE-018 — handoff Block 5 goal variant

```bash
grep -in "/moai goal" .claude/rules/moai/workflow/session-handoff.md | head
```
PASS when session-handoff.md documents the Block 5 `Run: /moai goal "<condition>"`
option and demotes the post-paste native-`/goal` follow-up to an optional variant.

### AC-GLE-019 — goal-directive.md row + Axis B

```bash
grep -in "/moai goal" .claude/rules/moai/workflow/goal-directive.md | head
grep -in "Axis B" .claude/rules/moai/workflow/goal-directive.md | head
```
PASS when goal-directive.md carries a `/moai goal` entry describing it as the
PROGRAMMATIC MoAI counterpart with an Axis B citation.

### AC-GLE-020 — native-invocation Axis B updated

```bash
grep -in "/moai goal" .claude/rules/moai/workflow/native-invocation-model.md | head
```
PASS when the Axis B worked illustration reflects `/moai goal` now reimplementing
the HUMAN-ONLY `/goal` (the "does not currently reimplement" sentence is updated).

### AC-GLE-021 — goal-evaluator reference + boundary doc (both clauses of REQ-GLE-021)

```bash
# clause (a): §2 stage ⑤ references the goal evaluator.
# awk-windowed into the §2 Request Processing Pipeline section (the lines between
# "## 2." and "## 3.") so a stray mention elsewhere in CLAUDE.md does not inflate
# the count. Anchored to the phrase "goal evaluator" (the precise term REQ-GLE-021
# mandates), NOT the generic "goal engine" — the latter already appears at
# CLAUDE.md:41 ("forthcoming goal engine") pre-implementation, which made the
# v0.2.0 combined grep `goal evaluator|goal engine` non-discriminating (baseline
# 1, not 0). Baseline (verified v0.2.1): 0.
awk '/^## 2\./,/^## 3\./' CLAUDE.md | grep -ic "goal evaluator"   # expect ≥1
# clause (b): the phase-vs-task boundary is documented in the Agentic Completion Loop file (baseline 0):
grep -ic "task-granular\|phase-granular\|goal engine" .claude/skills/moai/workflows/moai.md   # expect ≥1
```
Baseline (verified v0.2.1): clause (a) = 0 — the awk-windowed "goal evaluator"
phrase does NOT appear in §2 today; the stale v0.2.0 grep
`goal evaluator|goal engine` returned 1 due to CLAUDE.md:41's "forthcoming goal
engine" (added by `SPEC-ANALYZE-FIRST-ROUTING-001` commit `4d7ec04e4`), making
it non-discriminating. Clause (b) = 0. PATH CORRECTION — the Agentic Completion
Loop lives in **`.claude/skills/moai/workflows/moai.md`**, NOT
`.claude/output-styles/moai/moai.md` (the latter has no Agentic Completion Loop
section). **GATED**: clause (a) depends on `SPEC-ANALYZE-FIRST-ROUTING-001`
landing its §2 rewrite — this AC is evaluated only after ANALYZE-FIRST reaches
`completed` (the Depends_on pre-flight). Even with the gate, the check MUST be
discriminating when the gate clears, so the baseline-0 → post-≥1 property holds
(the v0.2.0 plan-auditor D2-2 defect is resolved by this re-anchor). PASS when
both clauses are ≥1.

### AC-GLE-022 — distinctness guard green

```bash
go test ./internal/config/ -run TestAgentic 2>&1 | tail -3   # expect ok/PASS
```
PASS when `agentic_loop_distinctness_test.go` still passes (max_iterations axes
remain distinct).

### AC-GLE-023 — internal/goal layout

```bash
ls internal/goal/schema.go internal/goal/state.go internal/goal/prune.go internal/goal/evaluate.go internal/cli/hook_stop_goal.go
```
PASS when the minimal-layout files exist (schema.go, state.go, prune.go,
evaluate.go, hook_stop_goal.go — the `prune.go` orphan-prune module included).

### AC-GLE-024 — coverage ≥85%

```bash
go test -cover ./internal/goal/... 2>&1 | tail -3
```
PASS when `coverage: N% of statements` with N ≥ 85.0.

### AC-GLE-025 — mirrors + neutrality + build

```bash
# per-file mirror existence for each changed .claude/ file + settings.json.tmpl
test -f internal/template/templates/.claude/skills/moai/workflows/goal.md && echo MIRROR_OK
grep -rn "SPEC-GOAL-ENGINE\|SPEC-ANALYZE-FIRST\|AGENTIC-CORE\|REQ-GLE" internal/template/templates/.claude/ | wc -l   # expect 0
make build ; echo "exit=$?"
```
PASS when every changed `.claude/` file has a mirror, the neutrality grep is 0,
and `make build` exits 0.

### AC-GLE-026 — Stop-hook COMPOSE (add-not-replace)

No AC previously pinned the Stop-hook compose-not-replace invariant. The
`settings.json.tmpl` already carries a `Stop` array with `handle-stop.sh`
(HARNESS-EVOLVE also adds a Stop hook). This AC proves the new entry is ADDED, not
substituted for the existing one:

```bash
# (a) existing entry preserved (baseline 2 — the two handle-stop.sh command lines):
grep -c 'handle-stop\.sh' internal/template/templates/.claude/settings.json.tmpl   # expect ≥1 (unchanged from baseline)
# (b) new entry added (baseline 0):
grep -c 'handle-stop-goal\.sh' internal/template/templates/.claude/settings.json.tmpl   # expect ≥1
```
Baseline (verified this iteration): (a) `handle-stop.sh` count = 2 (preserved),
(b) `handle-stop-goal.sh` count = 0. PASS when (a) is unchanged from its baseline
(the existing `handle-stop.sh` entry survives) AND (b) is ≥1 (the new
`handle-stop-goal.sh` entry is added) — proving add-not-replace composition. NOTE
`grep -c` on a `.tmpl` counts the two command-line variants (`bash "..."` +
`"..."`); the run-phase author records the exact pre-edit `handle-stop.sh` count
and asserts it is unchanged.

### AC-GLE-027 — goal.md documents progression-mode axis at kickoff (REQ-GLE-026)

```bash
# The /moai goal workflow file documents the autonomous/semi-autonomous axis offered at the kickoff gate.
grep -ci "semi-autonomous" .claude/skills/moai/workflows/goal.md   # expect ≥1
```
Baseline (verified this amendment): 0 (goal.md does not yet exist — it is a
run-phase deliverable of REQ-GLE-001). PASS when `goal.md` carries ≥1
`semi-autonomous` token documenting the progression-mode axis. This pins the
FEATURE-OWNER doc surface (goal.md is where `/moai goal` documents its own
progression semantics); the 3 cross-file doctrine surfaces are pinned separately
at AC-GLE-031..033. (SEPARATE from AC-GLE-030 which pins the orchestrator-bridge
flow — different token, same file, no vacuous-overlap risk.)

### AC-GLE-028 — autonomous mode continues without checkpoint (REQ-GLE-027)

```bash
go test ./internal/goal/ -run TestAutonomousModeNoCheckpoint -v 2>&1 | tail -5
```
PASS when the test asserts BOTH: (a) the state schema carries a
`progression_mode` field with value `"autonomous"` as the default, AND (b) given
`progression_mode == "autonomous"` and a goal NOT yet satisfied and ceiling NOT
reached, the evaluator emits the normal block JSON with NO `mode:
"semi-autonomous"` checkpoint signal (i.e., autonomous mode is the existing D3
behavior — no per-turn checkpoint). This verifies REQ-GLE-027's claim that
autonomous mode introduces no NEW behavioral surface beyond the state field.

### AC-GLE-029 — semi-autonomous checkpoint JSON emitted by hook (REQ-GLE-028a)

```bash
go test ./internal/goal/ -run TestSemiAutonomousCheckpointSignal -v 2>&1 | tail -5
```
PASS when the test asserts: given `progression_mode == "semi-autonomous"` AND
goal NOT satisfied AND ceiling NOT reached, `stop-goal` emits exit-0 stdout JSON
containing `"decision":"block"`, `"mode":"semi-autonomous"`, a `reason`
containing the literal `"semi-autonomous checkpoint"` prefix, AND — **when a
mechanical condition is failing** — a `failed_conditions` array whose entries
each carry `cmd`, `exit`, and `tail` (the failed-condition + output-tail detail
mandated by REQ-GLE-010, so the orchestrator's confirm AskUserQuestion surfaces
WHY the goal isn't satisfied; the generic `reason` label alone is insufficient —
v0.2.0 plan-auditor D2-1 defect resolved by this assertion). When no mechanical
condition is failing (e.g., the checkpoint fires because a model condition is
not yet satisfied), `failed_conditions` is empty `[]` or absent. This is the
hook-emits checkpoint half of REQ-GLE-028; the orchestrator-side confirm path is
pinned separately at AC-GLE-030.

### AC-GLE-030 — orchestrator confirm path documented (REQ-GLE-028b, the bridge)

```bash
# goal.md documents the orchestrator-bridge: reads the checkpoint JSON, runs AskUserQuestion.
grep -ci "checkpoint" .claude/skills/moai/workflows/goal.md   # expect ≥1
grep -ci "orchestrator" .claude/skills/moai/workflows/goal.md  # expect ≥1
```
Baseline (verified this amendment): 0 (goal.md does not yet exist). PASS when
goal.md carries ≥1 `checkpoint` token AND ≥1 `orchestrator` token documenting
that the ORCHESTRATOR (not the hook) reads the checkpoint-signal JSON and runs
the AskUserQuestion confirm round (the bridge pattern from
`agent-common-protocol.md` § User Interaction Boundary). This pins the
orchestrator-side reachability of the semi-autonomous confirm flow; the
hook-emits half is AC-GLE-029. (SEPARATE from AC-GLE-027 — `checkpoint` vs
`semi-autonomous` are different discriminating tokens on the same file.)

### AC-GLE-031 — CLAUDE.md progression-mode axis documented (REQ-GLE-029)

```bash
grep -ci "semi-autonomous" CLAUDE.md   # expect ≥1
```
Baseline (verified this amendment): 0. PASS when CLAUDE.md carries ≥1
`semi-autonomous` token in the kickoff / approval-gates context. NOTE: CLAUDE.md
is template-sensitive (§25) — the edit MUST be mirrored to
`internal/template/templates/CLAUDE.md` and MUST carry no internal SPEC IDs (the
progression-mode semantics are generic doctrine, not SPEC-specific). The mirror
is verified by REQ-GLE-025's build + the template-neutrality CI guard.

### AC-GLE-032 — run.md progression-mode axis documented (REQ-GLE-029)

```bash
# Co-located with the existing Run-phase Autonomy (/goal ac_converge) section.
grep -ci "semi-autonomous\|progression.mode" .claude/skills/moai/workflows/run.md   # expect ≥1
```
Baseline (verified this amendment): 0. PASS when run.md carries ≥1
`semi-autonomous` OR `progression mode` token co-located with the Run-phase
Autonomy section. (SEPARATE from AC-GLE-031 and AC-GLE-033 — 3 distinct doc
surfaces, each its own baseline-0 grep per the cross-file-reachability lesson.)

### AC-GLE-033 — orchestration-mode-selection.md progression-mode axis documented (REQ-GLE-029)

```bash
# Co-located with the Implementation Kickoff Approval mandatory-restoration policy.
grep -ci "semi-autonomous\|progression.mode" .claude/rules/moai/workflow/orchestration-mode-selection.md   # expect ≥1
```
Baseline (verified this amendment): 0. PASS when orchestration-mode-selection.md
carries ≥1 `semi-autonomous` OR `progression mode` token co-located with the
Kickoff mandatory-restoration header policy.

### AC-GLE-034 — kickoff mandatory in BOTH modes (safety-invariant NON-bypass)

```bash
# (a) doc grep: goal.md states the invariant for BOTH modes.
grep -ci "both.mode\|in both modes" .claude/skills/moai/workflows/goal.md   # expect ≥1
# (b) Go test: no auto-run-authorization path exists regardless of progression_mode.
go test ./internal/goal/ -run TestKickoffMandatoryBothModes -v 2>&1 | tail -5
```
Baseline (verified this amendment): (a) 0 (goal.md does not yet exist). PASS when
BOTH: (a) goal.md states that Implementation Kickoff Approval remains mandatory
in both autonomous AND semi-autonomous modes (the `both` / `in both modes`
anchor), AND (b) the Go test asserts NO code path in `internal/goal/` authorizes
run-phase entry — regardless of the `progression_mode` value (autonomous and
semi-autonomous both fail to bypass the gate). This is the safety-invariant
NON-bypass pin mandated by the amendment's binding AC-discipline. Distinct from
AC-GLE-015 (which verifies the general no-bypass property): AC-GLE-034
specifically pins that the NEW `progression_mode` field does not introduce a
mode-specific bypass.

### AC-GLE-035 — `moai goal` CLI registered (arm/status/clear) (REQ-GLE-030,031)

```bash
# Runtime --help smoke (preferred over a pure grep): the command must be registered under rootCmd.
go run ./cmd/moai goal --help 2>&1 | tail -20 ; echo "exit=$?"
# Each delivered verb must be present — three INDEPENDENT grep -q checks (NOT one OR-count,
# which the reachability lesson forbids: one token hitting N times must not pass).
H=$(go run ./cmd/moai goal --help 2>&1)
echo "$H" | grep -qw arm && echo "$H" | grep -qw status && echo "$H" | grep -qw clear && echo ALL3_PRESENT
```
Baseline (verified this amendment): `go run ./cmd/moai goal --help` → **non-zero** exit
(`unknown command "goal" for "moai"`) because no `goal` command is registered (verified
`grep goalCmd internal/cli/` → 0 command hits; only a `--goal` flag string in `handoff.go`).
PASS when `moai goal --help` exits 0 AND `ALL3_PRESENT` is printed (each of `arm`,
`status`, `clear` present, verified by an AND of three independent `grep -qw` — this
avoids the compound-OR-count vacuous-pass trap). Prefer this runtime smoke over a pure
`internal/cli/goal.go` symbol grep; a supplementary registration grep
(`grep -c 'rootCmd.AddCommand(goalCmd)' internal/cli/goal.go` → ≥1) MAY accompany it, but
the `--help` smoke is the load-bearing check — it proves the command is actually
reachable, not merely that a symbol exists.

### AC-GLE-036 — end-to-end arm → eval linkage (THE make-or-break AC) (REQ-GLE-030,032,033)

```bash
go test ./internal/cli/ -run TestGoalArmEvalLinkage -v 2>&1 | tail -8
```
PASS when the Go test asserts the full linkage in one flow, against a `t.TempDir()`
project root and a fixed session id `X`:
1. the arming path (the `moai goal arm` code path) PARSES the condition argument into a
   `conditions[]` entry (a bare shell string → `{type:mechanical}`; a transcript claim →
   `{type:model}` — REQ-GLE-032) with `ceiling.max_turns == 30`, and writes
   `.moai/state/goal/X.json` (assert the file exists at that EXACT path — NOT `pid-*.json`);
2. `moai hook stop-goal`, given the SAME session id `X` on its stdin, LOADS that goal
   (`LoadGoal(root, "X")` returns the armed goal) and emits the expected turn-end verdict
   — a block decision while a mechanical condition fails, or no-block once all conditions
   pass.
Baseline: no such linkage exists today — the arm path is absent (`grep goalCmd` → 0), so
arm and eval cannot share state. This AC is the make-or-break reachability pin: it cannot
pass today (the arm half cannot run) and passes only when the arm CLI AND the shared
session-id keying (AC-GLE-037) are both wired.

### AC-GLE-037 — session-id consistency (no silent pid fallback) (REQ-GLE-033)

```bash
go test ./internal/cli/ -run TestGoalArmResolvesSessionId -v 2>&1 | tail -8
```
PASS when the Go test asserts: given a resolvable real session id `X` (via the
`moai session current` / `resolveCurrentSessionID` path), the arming path writes
`.moai/state/goal/X.json` and does NOT write a `.moai/state/goal/pid-<n>.json` file
(i.e., it does not silently fall back to `WriterPidKey()` when a real session id is
available). The `WriterPidKey()` fallback remains valid ONLY when no real session id is
resolvable (REQ-GLE-008 unchanged). Baseline: no arming path exists, so no such assertion
is possible today. This pins the correctness property that makes AC-GLE-036's linkage
reachable — an arm CLI keyed on `pid-<n>` (a different PID than the hook) would write a
file the hook can never find.

### AC-GLE-038 — PruneOrphans wired on session-start (fail-open) (REQ-GLE-034)

```bash
# (a) real call site on the session-start path (baseline 0):
grep -c 'PruneOrphans' internal/hook/session_start.go   # expect ≥1 (was 0)
# (b) fail-open + orphan-moved behavior:
go test ./internal/hook/ -run TestSessionStartPrunesGoalOrphans -v 2>&1 | tail -8
```
Baseline (verified this amendment): (a) `grep -c 'PruneOrphans' internal/hook/session_start.go`
→ 0 (the only non-test occurrence of `PruneOrphans` in the repo is its DEFINITION at
`internal/goal/prune.go` — ZERO call sites). PASS when (a) ≥1 (a real call site exists on
the session-start path) AND (b) the Go test asserts BOTH: session-start moves an orphan
goal state file (a session id absent from `active-sessions.json`) to
`.moai/state/goal/consumed/`, AND a prune error does NOT block session start (fail-open —
the handler still returns its normal output on prune failure). NOTE: the grep is anchored
to `internal/hook/session_start.go` specifically (not any `_test.go`, not the
`internal/goal` definition), so it is discriminating — a call site elsewhere would not
satisfy it.

### AC-GLE-039 — resume NOT delivered (out of scope) (REQ-GLE-031)

```bash
# (a) the arm CLI does NOT register a runnable `resume` subcommand (out of scope §D.6):
go run ./cmd/moai goal --help 2>&1 | grep -qw resume && echo RESUME_PRESENT || echo RESUME_ABSENT
# (b) goal.md marks resume as deferred (run-phase doc edit):
grep -Eic 'resume[^.]*(defer|out of scope|follow-up)|(defer|out of scope|follow-up)[^.]*resume' .claude/skills/moai/workflows/goal.md   # expect ≥1
```
PASS when: (a) `moai goal --help` does NOT list `resume` as a subcommand (prints
`RESUME_ABSENT`) — the amendment delivers only `arm` / `status` / `clear` (REQ-GLE-031);
AND (b) `goal.md` carries ≥1 line marking `resume` as deferred / out of scope / follow-up
(the run-phase author annotates the existing `resume` section). Baseline (verified this
amendment): (b) = 0 today — `goal.md` documents `resume` as an active verb with no
deferral marker (line ~47-50: "Re-arm the most recently cleared goal ... best-effort
restore from the `consumed/` archive"). This pins that the amendment EXPLICITLY withholds
`resume` rather than silently leaving it half-built. (SEPARATE from AC-GLE-035, which
requires arm/status/clear PRESENT; this one requires resume ABSENT — the two together fix
the verb set to exactly the three delivered verbs.)

## §D.1 Definition of Done

- All 39 ACs PASS (34 original + 5 amendment-0.3.0 reachability ACs AC-GLE-035..039).
- `internal/goal/` ≥ 85% coverage; cross-platform build green
  (`GOOS=windows GOARCH=amd64 go build ./...`).
- Both router-registration surfaces (P1 list + Quick Reference) present.
- Stop-hook entry COMPOSED (existing entries preserved).
- `run.md ac_converge` section unmodified; `agentic_loop_distinctness_test.go` green.
- D8: progression-mode axis offered at kickoff (AC-GLE-027); `progression_mode`
  state field present (AC-GLE-028); semi-autonomous checkpoint JSON contract
  implemented (AC-GLE-029) + orchestrator-bridge documented (AC-GLE-030);
  kickoff-still-mandatory invariant pinned in BOTH modes (AC-GLE-034); 3 doc
  surfaces codified (AC-GLE-031..033).
- **Amendment 0.3.0 (arm CLI + prune wiring reachability)**: `moai goal` CLI
  registered under `rootCmd` with arm/status/clear verbs (AC-GLE-035); the
  make-or-break end-to-end arm→eval linkage passes (AC-GLE-036); the arming path
  resolves the session id via `moai session current` and does NOT silently
  pid-fallback (AC-GLE-037); `PruneOrphans` wired on the session-start path,
  fail-open (AC-GLE-038); `resume` explicitly NOT delivered (AC-GLE-039, §D.6).

## §D.2 Edge cases

- **No session id available**: `writer_pid` fallback (AC-GLE-008).
- **Goal with only mechanical conditions**: Tier 2 is skipped; all-pass → satisfied.
- **Goal with only model conditions**: Tier 1 is trivially satisfied; Tier 2 gates.
- **Concurrent sessions**: per-session file prevents cross-session clobber (AC-GLE-004).
- **Native /goal active**: MoAI `stop-goal` yields (AC-GLE-016) — no double-loop.
- **Depends_on unmet**: run-phase entry blocked unless ANALYZE-FIRST is `completed`
  or `--ignore-deps` + logged rationale (spec-workflow Depends_on pre-flight).
- **D8 — User declines the progression-mode choice at kickoff**: defaults to
  `autonomous` (existing D3 behavior); the kickoff gate still requires explicit
  approval (REQ-GLE-026).
- **D8 — User clears the goal mid-semi-autonomous-loop**: orchestrator runs
  `moai goal clear`; the checkpoint loop ends; no further AskUserQuestion rounds.
- **D8 — Semi-autonomous checkpoint with ceiling approaching**: the checkpoint
  JSON carries `turn` + `ceiling`; the orchestrator surfaces "N turns remaining"
  in the confirm AskUserQuestion so the user can decide with full information.
- **D8 — Mode switched autonomous→semi-autonomous mid-goal**: the orchestrator
  updates `progression_mode` in state; subsequent turns emit checkpoints.
- **D8 — Hook emits checkpoint but orchestrator context lost (compact/clear)**:
  the checkpoint JSON is the prior turn's block reason; on resume the orchestrator
  re-derives from goal state `progression_mode` (state-file-first detection,
  consistent with the session-handoff pattern).
- **Amendment 0.3.0 — no resolvable session id at arm time**: when `moai goal arm`
  runs and `moai session current` returns the fallback (runtime did not expose
  session.id), the arm path uses `WriterPidKey()` (`pid-<pid>`) per REQ-GLE-008 — but
  this yields a goal the hook (different PID) cannot find. The arm CLI SHOULD warn the
  user that the goal may be unreachable until a real session id is registered (surfaced,
  not silent). AC-GLE-037 pins that the pid fallback is NOT taken when a real id IS
  resolvable; the no-id case is the documented degrade.
- **Amendment 0.3.0 — `moai goal arm` with no active session state dir**: the arm path
  creates `.moai/state/goal/` if absent (`SaveGoal` atomic temp+rename already
  MkdirAll's the dir) — arming does not require a pre-existing goal dir.
- **Amendment 0.3.0 — session-start prune with no goal state dir**: `PruneOrphans` on a
  project with no `.moai/state/goal/` returns cleanly (no orphans, no error); fail-open
  means even a read error never blocks session start (AC-GLE-038b).
- **Amendment 0.3.0 — `moai goal clear` on an already-cleared / absent goal**: `ClearGoal`
  `os.Remove` on a missing file is tolerated (clear is idempotent); the CLI reports
  "no armed goal" rather than erroring.
