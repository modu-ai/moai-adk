# plan.md — SPEC-PROJECT-NAVIGATOR-004

> Implementation plan. Ordered by decision-reversibility: the selection-logic decision (M1) is the most likely to be revisited and is presented first; mechanical Template-First + test steps follow.

## §A. Context

The Navigator regeneration script (`navigator-regen.sh`) ships as a distributed template and emits three markdown files under `.moai/project/navigator/`. The "Next task" line at `navigator.md` is the single most visible output: it tells the user which SPEC to advance next. Today the line is selected by `awk ... { print; exit }` over the alphabetically-sorted rows that are merely *not* in `{completed, superseded, archived, rejected}` — which lets an `implemented` SPEC (the most common "done but not yet sync-closed" state in older flows) win the recommendation while an active `in-progress` SPEC is missed.

Root cause has two layers (verified against worktree HEAD `adaff36e5`):

1. The candidate filter at line 212 / 218 excludes `{completed, superseded, archived, rejected}` but NOT `implemented`.
2. "Next task" is the alphabetically-first non-terminal SPEC — no status priority, so an `implemented` SPEC at `SPEC-AGENCY-ABSORB-001` beats everything.

Template/mirror byte-parity confirmed via `cmp` (both files byte-identical at HEAD).

## §B. Known Issues

- None blocking. The fix is localized to one shell-script block + one test extension.

## §C. Pre-flight (must hold before M1)

- [ ] CWD is the worktree (`.claude/worktrees/nav-nexttask`), NOT the primary checkout. Verify with `pwd` + `git rev-parse --show-toplevel` before every write.
- [ ] SPEC ID regex PASS: `SPEC-PROJECT-NAVIGATOR-004` validated.
- [ ] Branch rename pending pre-push: `plan/SPEC-NAVIGATOR-NEXTTASK-001` → `plan/SPEC-PROJECT-NAVIGATOR-004` (or the run-phase branch `fix/SPEC-PROJECT-NAVIGATOR-004`). The provisional branch name is retained through plan-phase; the orchestrator renames at run-phase entry.
- [ ] No foreign active session on this SPEC scope (Pre-Spawn Sync Check, run-phase only — plan-phase writes are orchestrator-direct under the worktree-isolation guarantee).

## §D. Constraints (binding)

1. Template-First (CLAUDE.local.md §2 [HARD]): edit template source + mirror in ONE commit, byte-identical. `internal/template/rule_template_mirror_test.go` enforces parity at CI.
2. §25 neutrality: no SPEC IDs / REQ tokens / dates / SHAs / internal paths in the script body. The selection predicate references only SPEC *status enum values* (`in-progress`, `draft`, `completed`, etc.) — these are generic lifecycle tokens, not moai-adk-internal identifiers. Run the 5-item pre-commit self-check (`template-internal-isolation-doctrine.md` §25.3).
3. Reproduction-First (CLAUDE.md §7 Rule 4): write the regression test, observe it red on the current logic, then apply the fix, observe it green.
4. `make build` recompiles the embedded FS after the template edit.
5. No `@MX` tags — the change is in a shell script, not Go.
6. LSP baseline: n/a (bash). The project's bash surface has no LSP; `go vet`/`golangci-lint` apply only to the Go test file, which is a standard `*_test.go`.

## §E. Self-Verification (manager-develop §E — run-phase)

Populated at run-phase by manager-develop. Plan-phase leaves the section headers as placeholders:

- §E.1 Plan-phase Audit-Ready Signal — populated by this plan-phase artifact set.
- §E.2 Run-phase Evidence — `_<pending run-phase>_`
- §E.3 Run-phase Audit-Ready Signal — `_<pending run-phase>_`
- §E.4 Sync-phase Audit-Ready Signal — `_<pending sync-phase>_`

## §F. Milestones

> Ordered by decision-reversibility: M1 is the selection-logic decision (highest change-likelihood — a reviewer may prefer "exclude implemented but keep alphabetical" over "status-tier preference"); M2 is the Template-First mechanical step; M3 is the test extension; M4 is the doc note + build.

### M1 — Selection-logic fix in `navigator-regen.sh` (Priority High)

**Decision to surface for review**: the fix adopts a *positive status-tier preference* — "Next task" = first `in-progress` SPEC by alphabetical sort, else first `draft` SPEC, else none — rather than merely extending the negative exclusion list with `implemented`. Reasons:

- A positive predicate defaults new statuses (any future enum addition) to excluded-from-Next-task, which is the safe default for a recommendation surface (R2 in spec.md).
- Pure negative-exclusion (`... && $3 != "implemented"`) would let an unknown future status win the recommendation silently.

**Alternative considered**: extend the negative exclusion list only (`$3 != "implemented"`) and keep pure alphabetical selection. Rejected because it does not fix the secondary symptom in the memory report — the *alphabetically-first* `draft` SPEC beating an *in-progress* SPEC that is further down the alphabet.

**Files** (both byte-identical in one commit):
- `internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` (template source, the SSOT)
- `.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` (local mirror)

**Approach**: in the "Next task" block (lines ~216–224), replace the single `awk ... $3 != "completed" && ... { print; exit }` with a two-stage selection:

1. First, try an `in-progress` candidate: `awk -F'\t' '$3 == "in-progress" { print; exit }' "$ROWS_FILE"`.
2. If empty, try a `draft` candidate: `awk -F'\t' '$3 == "draft" { print; exit }' "$ROWS_FILE"`.
3. If both empty, fall through to the existing "No active SPEC" branch.

Leave the "Current frontier" display filter (line 212) unchanged — REQ-PN-004-003 binds it to the inclusive semantics. Add a one-line inline comment at the "Next task" site noting the deliberate filter asymmetry (mitigates R1).

**No change** to `handle-session-start-navigator.sh` — REQ-PN-004-004.

### M2 — Template-First parity commit (Priority High)

- Verify `cmp` returns byte-identical AFTER the edit (re-run the parity check, not just trust the edit).
- Run the 5-item §25 neutrality pre-commit self-check.
- `make build` (recompiles embedded FS).
- This is one atomic commit together with M1 — the template source + mirror + catalog regeneration land together.

### M3 — Regression test extension (Priority High)

**File**: `internal/template/navigator_regen_test.go` (the existing black-box test driver that invokes the script as a subprocess against a `t.TempDir()` fixture project).

**New test**: `TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress` (or extend an existing test — the run-phase implementer names it).

**Fixture** (built under `t.TempDir()` via the existing `initFixtureRepo` helper): three SPECs with distinct statuses:

| SPEC-ID | title | status |
|---------|-------|--------|
| SPEC-A-001 | alpha | implemented |
| SPEC-B-001 | beta | in-progress |
| SPEC-C-001 | gamma | draft |

The alphabetical sort places SPEC-A-001 first; the pre-fix logic would therefore pick SPEC-A-001 (the misclassification). The new test asserts:

- `navigator.md` "Next task" line names SPEC-B-001 (in-progress), NOT SPEC-A-001.
- `navigator.md` "Current frontier" list still contains SPEC-A-001 (the implemented one) — REQ-PN-004-003.

**Reproduction-First discipline**: write the test, run `go test ./internal/template/ -run TestNavigatorRegen_NextTask -count=1`, observe RED (pre-fix logic picks SPEC-A-001), then apply M1, re-run, observe GREEN. Cite both outputs verbatim in the §E.2 evidence block.

### M4 — Optional doc note + final verification (Priority Medium)

- `references/navigator.md` §2 schema (both template + mirror sides, byte-identical): add a one-line note under the "Next task" example that the selection prefers `in-progress` SPECs over `draft` SPECs, with alphabetical sort as a tiebreaker within a tier. Assess: if the current schema description would otherwise mislead a reader into thinking pure-alphabetical selection, include the note; otherwise skip. The run-phase implementer decides based on the current prose at §2.
- Final verification batch (single-turn, read-only): `go test ./internal/template/ -count=1`, `golangci-lint run ./internal/template/`, `cmp` template vs mirror, `make build` clean.

## §G. Anti-Patterns (do NOT)

- **AP-004-001 — Unify the frontier and next-task filters**: do NOT refactor the "Current frontier" filter to also exclude `implemented`. That is a display regression (REQ-PN-004-003 forbids it).
- **AP-004-002 — Negative-only exclusion list for Next task**: do NOT fix REQ-PN-004-001 by merely adding `&& $3 != "implemented"` to the existing awk. Use the positive status-tier predicate per M1 (R2 mitigation).
- **AP-004-003 — Edit mirror without template source** (or vice versa): the parity test will fail CI.
- **AP-004-004 — Edit in two commits**: the template + mirror + catalog regeneration MUST land atomically (CLAUDE.local.md §2 [HARD]).
- **AP-004-005 — Touch `handle-session-start-navigator.sh`**: no hook change is needed (REQ-PN-004-004). A hook edit would be scope creep.
- **AP-004-006 — Cite SPEC IDs in the script body** (e.g. a comment mentioning SPEC-AGENCY-ABSORB-001): §25 neutrality violation. The script must reference only status enum tokens.

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-PROJECT-NAVIGATOR-004/spec.md`
- acceptance.md: `.moai/specs/SPEC-PROJECT-NAVIGATOR-004/acceptance.md`
- Bug memory: `feedback_navigator_next_task_misclassification.md`
- Template-First Rule: CLAUDE.local.md §2
- §25 neutrality: `.moai/docs/template-internal-isolation-doctrine.md`
- Existing test driver: `internal/template/navigator_regen_test.go`
