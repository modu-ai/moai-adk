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
# clause (a): §2 stage ⑤ references the goal evaluator (baseline 0):
grep -ic "goal evaluator\|goal engine" CLAUDE.md   # expect ≥1
# clause (b): the phase-vs-task boundary is documented in the Agentic Completion Loop file (baseline 0):
grep -ic "task-granular\|phase-granular\|goal engine" .claude/skills/moai/workflows/moai.md   # expect ≥1
```
Baseline (verified this iteration): both 0. PATH CORRECTION — the Agentic
Completion Loop lives in **`.claude/skills/moai/workflows/moai.md`**, NOT
`.claude/output-styles/moai/moai.md` (the latter has no Agentic Completion Loop
section). **GATED**: clause (a) depends on `SPEC-ANALYZE-FIRST-ROUTING-001`
landing its §2 rewrite — this AC is evaluated only after ANALYZE-FIRST reaches
`completed` (the Depends_on pre-flight). PASS when both clauses are ≥1.

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

## §D.1 Definition of Done

- All 26 ACs PASS.
- `internal/goal/` ≥ 85% coverage; cross-platform build green
  (`GOOS=windows GOARCH=amd64 go build ./...`).
- Both router-registration surfaces (P1 list + Quick Reference) present.
- Stop-hook entry COMPOSED (existing entries preserved).
- `run.md ac_converge` section unmodified; `agentic_loop_distinctness_test.go` green.

## §D.2 Edge cases

- **No session id available**: `writer_pid` fallback (AC-GLE-008).
- **Goal with only mechanical conditions**: Tier 2 is skipped; all-pass → satisfied.
- **Goal with only model conditions**: Tier 1 is trivially satisfied; Tier 2 gates.
- **Concurrent sessions**: per-session file prevents cross-session clobber (AC-GLE-004).
- **Native /goal active**: MoAI `stop-goal` yields (AC-GLE-016) — no double-loop.
- **Depends_on unmet**: run-phase entry blocked unless ANALYZE-FIRST is `completed`
  or `--ignore-deps` + logged rationale (spec-workflow Depends_on pre-flight).
