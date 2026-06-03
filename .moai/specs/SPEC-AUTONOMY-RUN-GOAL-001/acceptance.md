# Acceptance Criteria — SPEC-AUTONOMY-RUN-GOAL-001

> Each AC is a Given-When-Then scenario traceable to a REQ. Blocking ACs gate completion. Verification is grep/rule/test-based (low Go-code volume, doctrine-focused).

## §A — Traceability Matrix

| AC | REQ | Deliverable | Severity |
|----|-----|-------------|----------|
| AC-ARG-001 | REQ-ARG-001 | D1 | Blocking |
| AC-ARG-002 | REQ-ARG-002, -003 | D1 | Blocking |
| AC-ARG-003 | REQ-ARG-004, -005 | D1 | Blocking |
| AC-ARG-004 | REQ-ARG-006 | D2 | Blocking |
| AC-ARG-005 | REQ-ARG-007 | D2 | Blocking |
| AC-ARG-006 | REQ-ARG-008 | D2 | Blocking |
| AC-ARG-007 | REQ-ARG-009, -010 | D2 | Blocking |
| AC-ARG-008a | REQ-ARG-011 | D3 | Blocking |
| AC-ARG-008b | REQ-ARG-012, -013 | D3 | Blocking |
| AC-ARG-009 | REQ-ARG-014 | cross-cutting | Blocking |
| AC-ARG-010 | REQ-ARG-015 | cross-cutting | Non-blocking |
| AC-ARG-011 | REQ-ARG-016 | mirror | Blocking |
| AC-ARG-012 | EX-1..EX-8 | scope guard | Blocking |

> Note: AC-ARG-008a / AC-ARG-008b are two sub-criteria of one logical GATE-2-preservation AC group (the lowercase suffix is an acceptance-criteria sub-ID convention, scoped to this file only — SPEC IDs never carry an alpha suffix).

## §B — Acceptance Criteria

### AC-ARG-001 — Mode 6 appended without disturbing Modes 1–5
**Given** the canonical mode catalog `.claude/rules/moai/workflow/orchestration-mode-selection.md` previously defined 5 modes (trivial / background / agent-team / parallel / sub-agent),
**When** D1 is applied,
**Then** the §A Mode Catalog table contains a 6th row `workflow` (Mode 6), AND Modes 1–5 retain their original numbers and labels.
**Verify**:
```bash
grep -cE '^\| 6 \| `workflow`' .claude/rules/moai/workflow/orchestration-mode-selection.md   # == 1
grep -cE '^\| [1-5] \| `(trivial|background|agent-team|parallel|sub-agent)`' .claude/rules/moai/workflow/orchestration-mode-selection.md   # == 5
```

### AC-ARG-002 — Mode 6 entry conditions are strict (≥30 files AND mechanical AND parallel; coding-heavy → Mode 5)
**Given** Mode 6 is added,
**When** a reader inspects the §B decision tree and §C capability gate,
**Then** the Mode 6 candidate branch states all three entry conditions (scope ≥ ~30 files, mechanical transformation, genuinely parallel / no inter-file dependency), AND a fall-through clause prefers Mode 5 for coding-heavy or multi-domain work citing the Finding A4 caveat.
**Verify**:
```bash
grep -niE 'mode 6|workflow.*(30 files|mechanical|genuinely parallel)' .claude/rules/moai/workflow/orchestration-mode-selection.md | head
grep -niE 'Finding A4|coding-heavy.*Mode 5|fewer truly parallelizable' .claude/rules/moai/workflow/orchestration-mode-selection.md   # ≥1
```

### AC-ARG-003 — Mode 6 selection requires GATE-2-passed + preferences-collected, logged to progress.md; pre-GATE-2 launch rejected
**Given** Mode 6 is a candidate,
**When** the §C gate and §E anti-patterns are inspected,
**Then** the gate states Mode 6 is selectable ONLY after GATE-2 passed AND all preferences collected AND the selection is logged to `progress.md` § Mode Selection; AND an §E anti-pattern entry rejects a Mode 6 launch attempted before GATE-2 passes.
**Verify**:
```bash
grep -niE 'GATE-2.*pass|preferences.*collect|progress\.md.*Mode Selection' .claude/rules/moai/workflow/orchestration-mode-selection.md | head
grep -niE 'Mode 6.*before GATE-2|launch.*before GATE-2' .claude/rules/moai/workflow/orchestration-mode-selection.md   # ≥1 (anti-pattern)
```

### AC-ARG-004 — `/goal` `ac_converge` set only after GATE-2 approval
**Given** the NEW self-contained `## Run-phase Autonomy (/goal ac_converge)` section added by D2 to the router file `.claude/skills/moai/workflows/run.md` (which had ZERO GATE-2/`/goal` markers before this SPEC), co-locating BOTH the GATE-2 `AskUserQuestion` ordering reference AND the `/goal ac_converge` set,
**When** that section is inspected,
**Then** the section states the `ac_converge` goal is set ONLY after GATE-2 approval is obtained, AND the GATE-2 `AskUserQuestion` reference textually precedes the first `/goal` token in the file.
**Verify**:
```bash
# FIRST GATE-2 marker line precedes FIRST /goal token (first-match guard on both — the
# GATE-2 human-gate ordering anchor; later §19.1 cross-reference GATE-2 mentions are not the anchor)
awk '/GATE-2/{if(!g)g=NR} /\/goal/{if(!gl)gl=NR} END{print "gate2_line="g" goal_line="gl; exit !(g>0 && gl>0 && g<gl)}' .claude/skills/moai/workflows/run.md
```

### AC-ARG-005 — `ac_converge` predicates are transcript-measurable (no file-path predicate)
**Given** the `ac_converge` condition text,
**When** the condition is inspected,
**Then** every predicate references an orchestrator-surfaced transcript line (e.g., "PASS evidence surfaced", "`go test ./...` exit 0 surfaced", "git status surfaced"), AND the condition does NOT instruct the evaluator to read a file at a path.
**Verify**:
```bash
grep -niE 'surfaced|exit 0|AC-id: PASS|git status' .claude/skills/moai/workflows/run.md | head
# Negative: no "read .moai/specs/.../acceptance.md" file-read predicate inside the condition block
! grep -niE '/goal.*read .*\.md|evaluator.*read.*file' .claude/skills/moai/workflows/run.md
```

### AC-ARG-006 — `ac_converge` carries `max 20 turns` bound
**Given** the `ac_converge` condition,
**When** inspected,
**Then** it contains an explicit bound `max 20 turns`.
**Verify**:
```bash
grep -niE 'max 20 turns' .claude/skills/moai/workflows/run.md   # ≥1
```

### AC-ARG-007 — Semantic-failure escape + non-substitution clause present
**Given** the run autonomy section,
**When** inspected,
**Then** it states that on semantic failure (data race / deadlock / panic / test assertion) the orchestrator clears the goal and escalates via `AskUserQuestion`, AND it states the goal never bypasses GATE-2, PR creation, or destructive operations.
**Verify**:
```bash
grep -niE 'data race|deadlock|panic|assertion.*clear.*goal|escalat' .claude/skills/moai/workflows/run.md | head
grep -niE "goal.*not.*bypass|never.*substitut|PR creation.*separate|destructive" .claude/skills/moai/workflows/run.md   # ≥1
```

### AC-ARG-008a — GATE-2 human gate is emitted before any `/goal`/Mode-6 launch (regression)
**Given** a run-phase entry, where D2 has introduced BOTH the GATE-2 `AskUserQuestion` ordering marker AND the `/goal ac_converge` set into the ONE self-contained `## Run-phase Autonomy (/goal ac_converge)` section of `.claude/skills/moai/workflows/run.md` (the guard relies on both markers being co-located in this single file, not in the EX-5-excluded `phase-execution.md`),
**When** the GATE-2 preservation regression guard runs (Go test in `internal/template/` reading the `run.md` body, OR an audit assertion),
**Then** the guard PASSES only when `run.md` emits a GATE-2 `AskUserQuestion` human-approval gate textually before the first `/goal` set AND before any Mode-6 launch reference.
**Verify**:
```bash
go test ./internal/template/... -run 'TestGate2PreservedBeforeGoal' -v 2>&1 | tail -5   # PASS
# (guard implementation introduced by this SPEC at M3)
```

### AC-ARG-008b — GATE-2 is score-independent + doctrine cross-reference present (regression)
**Given** a plan-auditor verdict of PASS with score ≥ 0.90 (skip-eligible),
**When** the regression guard's score-independence check runs,
**Then** the run skill body contains a statement that GATE-2 is emitted regardless of plan-auditor score (skip-eligibility applies only to Phase 0.5 verdict re-execution, not GATE-2), AND contains a cross-reference to CLAUDE.local.md §19.1 / REQ-ATR-015.
**Verify**:
```bash
grep -niE 'regardless of.*plan-auditor score|skip-eligib.*Phase 0\.5.*not GATE-2|score-independent' .claude/skills/moai/workflows/run.md   # ≥1
grep -niE '§19\.1|REQ-ATR-015' .claude/skills/moai/workflows/run.md   # ≥1
```

### AC-ARG-009 — No typed/named Workflow script API asserted; 6 safety conditions preserved
**Given** all delivered text (rule + skill body + template mirrors),
**When** scanned,
**Then** no typed/named Workflow script API signature is asserted (`agent(`/`parallel(`/`pipeline(`/`phase(` as documented functions), AND the conceptual coordinate-agents model is used instead; AND the 6 safety-condition concepts (GATE-2 score-independent, preferences-before-launch, no-nested-spawn, background-read-only, transcript-measurable+bounded goal, workflow-parallel-only/loop-diagnostic) are each represented.
**Verify**:
```bash
# Negative: no asserted named-script API function signatures
! grep -nE '\b(agent|parallel|pipeline|phase)\s*\(' .claude/rules/moai/workflow/orchestration-mode-selection.md
grep -niE 'coordinate.*agent|script variable|final synthesis|no documented.*API' .claude/rules/moai/workflow/orchestration-mode-selection.md | head
```

### AC-ARG-010 — Mode-6/`/goal` agents return blocker report, never prompt user (non-blocking)
**Given** a Mode-6 Workflow agent or `/goal`-turn agent missing a required input,
**When** the delivered doctrine describes the failure path,
**Then** the text states the agent returns a structured blocker report and the orchestrator runs `AskUserQuestion` + re-delegates (agents never prompt the user — asymmetric boundary).
**Verify**:
```bash
grep -niE 'blocker report|never prompt|asymmetric boundary|re-delegat' .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/skills/moai/workflows/run.md | head
```

### AC-ARG-011 — Template mirror + build green
**Given** the `.claude/` edits (D1 + D2 + D3 guard if applicable),
**When** the run-phase mirror step completes,
**Then** the corresponding `internal/template/templates/` files are updated, `make build` exits 0, and mirror-drift + neutrality tests pass.
**Verify**:
```bash
make build   # exit 0
go test ./internal/template/... -run 'TestRuleTemplateMirror|TestTemplateNeutralityAudit' 2>&1 | tail -5   # PASS
```

### AC-ARG-012 — Scope guard: no sibling-SPEC scope pulled in
**Given** this SPEC's exclusions (EX-1..EX-8),
**When** the diff is inspected,
**Then** no `workflow.yaml` autonomy-profile Go struct is added (EX-1), no full Workflow pattern catalog is registered in `dynamic-workflows.md` (EX-2), no full `/goal` condition registry beyond `run`'s `ac_converge` is added (EX-3), the legacy run sub-skill agent chain is NOT rewritten (EX-5), and `autonomy.enabled` is not flipped on by default (EX-7).
**Verify**:
```bash
# No new autonomy nested struct / profile accessor in internal/config (EX-1).
# D6 broadening: also reject MaxTurns field + autonomy: yaml key + max_turns token,
# so a partial config-schema leak (struct field OR yaml key) is caught, not only the
# `Autonomy ... struct` / `goal_condition_template` forms.
! grep -rniE 'Autonomy[A-Za-z]*[[:space:]]+struct|goal_condition_template|MaxTurns|max_turns|^[[:space:]]*autonomy:' internal/config/ 2>/dev/null
# dynamic-workflows.md not expanded with a 5-pattern catalog (EX-2) — file unchanged or only cross-ref added
git diff --stat .claude/rules/moai/workflow/dynamic-workflows.md | tail -1
# run condition is the only /goal condition added (EX-3)
grep -cE 'ac_converge' .claude/skills/moai/workflows/run.md   # ≥1; no plan/sync/loop condition templates added in this SPEC's diff
```
> D6 note: the EX-1 negative grep is intentionally broadened to `MaxTurns` / `max_turns` / `autonomy:` (in addition to the original `Autonomy ... struct` / `goal_condition_template`) so that a partial leak of the sibling SPEC-AUTONOMY-CONFIG schema (either the Go struct field or the yaml key) is caught. The `internal/config/` path scoping keeps this from false-positiving on the inline `max_turns: 20` design-intent reference inside the SPEC artifacts themselves (which live under `.moai/specs/`, not `internal/config/`).

## §C — Edge Cases

- **EC-1 — `/goal` unavailable (version < v2.1.139 or hooks disabled)**: the run autonomy section MUST note graceful degradation (autonomy off, manual flow) rather than failing. Verify the section references the preflight/version-gate concept (defer the actual preflight implementation to `SPEC-AUTONOMY-CONFIG` / EX-1, but the degradation note is in-scope).
- **EC-2 — Mode 6 candidate but Workflows disabled (`CLAUDE_CODE_DISABLE_WORKFLOWS=1`)**: the catalog text MUST note Mode 6 falls through to Mode 5 when Workflows are disabled (cannot assume availability).
- **EC-3 — Exactly 30 files (boundary)**: the `≥ ~30 files` threshold uses the tilde to signal a soft boundary; at the boundary, the existing §B.2 tie-breaker ("default to the simpler mode") resolves toward Mode 5.
- **EC-4 — GATE-2 declined by user**: when the user does not approve at GATE-2, no `/goal` is set and no Mode-6 Workflow launches — run-phase entry halts (consistent with the existing Decision-Point exit behavior).

## §D — Definition of Done

- [ ] All Blocking ACs (AC-ARG-001..009, 011, 012; plus 008a/008b) PASS with surfaced evidence.
- [ ] AC-ARG-010 (non-blocking) addressed or explicitly deferred with rationale.
- [ ] `moai spec lint .moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/` clean (no MissingExclusions, no FrontmatterInvalid).
- [ ] `go test ./...` exit 0.
- [ ] `golangci-lint run` no NEW findings.
- [ ] Template mirror + `make build` green (AC-ARG-011).
- [ ] GATE-2 preservation regression guard committed and green (AC-ARG-008a/008b).
- [ ] No sibling-SPEC scope creep (AC-ARG-012).
- [ ] Quality gate (TRUST 5) satisfied: Tested (regression guard), Readable, Unified, Secured (no destructive autonomy), Trackable (Conventional Commits).
