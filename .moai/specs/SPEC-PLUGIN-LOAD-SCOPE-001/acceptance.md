# Acceptance — SPEC-PLUGIN-LOAD-SCOPE-001

All commands are plain, separately invocable, and run from the card worktree root. The exit code is
its own field. The pinned tree for every RED-now observation is `802a72235`. A path written
`evidence/...` or `verdict.md` without a directory prefix is relative to `.moai/reports/t1434/`.

## §D AC Matrix

| AC | Requirement | Description | Severity | Verification |
|----|-------------|-------------|----------|--------------|
| AC-001 | REQ-001, REQ-014 | Repository writes confined to the report directory; no product path touched | Must | Direct |
| AC-002 | REQ-002 | Isolation proven by observed scratch writes and an unchanged real manifest | Must | Direct |
| AC-003 | REQ-003 | Non-volatile protected-set rows identical before and after | Must | Direct |
| AC-004 | REQ-003 | The restoration diff is shown able to fail (negative control) | Must | Direct |
| AC-005 | REQ-004 | Every measured command is one scrubbed compound invocation | Must | Direct |
| AC-006 | REQ-005, REQ-013 | Verdict table has the 14 fixed rows with both columns | Must | Direct |
| AC-007 | REQ-012 | Every cell has a five-field evidence card; non-UNOBSERVED cells cite existing raw output | Must | Direct |
| AC-008 | REQ-012 | The verdict checker is shown able to fail (negative control) | Must | Direct |
| AC-009 | REQ-006, REQ-012 | Static measurement ran for Claude: validate, install, details evidence exists | Must | Direct |
| AC-010 | REQ-007, REQ-008, REQ-010 | Runtime cells carry a positive control, or are UNOBSERVED with a reason | Must | Direct |
| AC-011 | REQ-009 | Settings-hook equivalence evidence covers execution, expansion, and timeout | Must | Direct |
| AC-012 | REQ-011 | Codex measured under a scratch `CODEX_HOME` with a recorded command trail | Must | Direct |
| AC-013 | REQ-012, REQ-013 | Version stamps present and start/end versions agree; non-goal line present | Must | Direct |

Budget check (Tier M ceilings 16/16): 14 requirements, 13 acceptance criteria.

## RED-now Ledger

The deliverables do not exist at the pinned tree, so every criterion that reads a deliverable is red
for a stated reason: the work has not run. Each flips green at the milestone named in its row.

| Ledger | Command | Verbatim stdout | Exit code | Tree SHA | Why red |
|--------|---------|-----------------|-----------|----------|---------|
| L-1 | `test -d .moai/reports/t1434` | (empty) | 1 | 802a72235 | report directory not yet created; flips at M1 |
| L-2 | `test -f .moai/reports/t1434/check-verdict.sh` | (empty) | 1 | 802a72235 | checker not yet written; flips at M5 |
| L-3 | `test -f .moai/reports/t1434/verdict.md` | (empty) | 1 | 802a72235 | verdict not yet written; flips at M5 |

AC-001 is an invariant (no product-path change) and is green at arrival by construction; it is a
regression guard, not a release gate, and carries the positive control in its own text.

## §D.1 AC-001 — Repository writes confined (regression guard)

**Given** the card branch after the run phase, and the base printed by `git merge-base develop HEAD` (run as its own invocation; call its output BASE)
**When** `git diff --name-only BASE..HEAD -- internal cmd pkg .claude .mcp.json CLAUDE.md AGENTS.md` runs
**Then** stdout is empty and exit is 0; and the positive control `git diff --name-only BASE..HEAD` lists at least one path, every path beginning `.moai/reports/t1434/` or `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/`; an empty control means the measurement is reported as unmeasured, not as a pass.

## §D.2 AC-002 — Isolation proven by observation

**Given** `.moai/reports/t1434/evidence/00-isolation.txt` written in M1
**When** `grep -c "ISOLATION-OBSERVED" .moai/reports/t1434/evidence/00-isolation.txt` runs
**Then** the count is at least 2 (one line per tool, Claude and Codex), each line naming the scratch file that received the write and the real manifest diff result; where isolation failed for a tool the file instead carries `ISOLATION-FAILED <tool>` and the real-config protocol evidence (AC-003) covers that tool.

## §D.3 AC-003 — Restoration is byte-identical on the non-volatile rows

**Given** `evidence/protected-t0.sha256`, `evidence/protected-t0b.sha256` (the second pre-probe manifest), `evidence/protected-end.sha256`, and `evidence/volatile-rows.txt` (rows that differ between t0 and t0b)
**When** `diff -u evidence/protected-t0.nonvolatile.sha256 evidence/protected-end.nonvolatile.sha256` runs (both files are the manifests with the volatile rows removed, produced by the probe script)
**Then** stdout is empty and exit is 0, and `wc -l evidence/protected-t0.nonvolatile.sha256` reports at least 3 rows, so an empty manifest cannot pass vacuously; a run that left a non-volatile row changed fails the SPEC.

## §D.4 AC-004 — The restoration diff can fail (negative control)

**Given** a deliberately altered copy `evidence/protected-end.altered.sha256` (one hash changed by the probe script)
**When** `diff -q evidence/protected-t0.nonvolatile.sha256 evidence/protected-end.altered.sha256` runs
**Then** exit is 1 and stdout names the differing files, recorded verbatim in `evidence/negative-control-restore.txt`.

## §D.5 AC-005 — Every measured command is scrubbed and logged

**Given** `evidence/commands.log` where each measured command is one line beginning `CMD: unset ` and containing ` && `
**When** `grep -c "^CMD: " evidence/commands.log` and `grep -c "^CMD: unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED" evidence/commands.log` run
**Then** the two counts are equal and at least 40, so no command ran unscrubbed; any `MOAI_FACTORY_*` name listed in `evidence/env-scrub.txt` also appears in each `CMD:` line, checked by `grep -c` of one such name equalling the first count.

## §D.6 AC-006 — Verdict table has the fixed row set

**Given** `.moai/reports/t1434/verdict.md`
**When** `grep -c "^| R[0-9][0-9] " .moai/reports/t1434/verdict.md` and `grep -c "^| ID | Component | Claude plugin | Codex plugin | Recommended home | Consequence" .moai/reports/t1434/verdict.md` run
**Then** the first count is 14 and the second is 1, and no cell outside the vocabulary PLUGIN-OK, PROJECT-ONLY, PARTIAL(...), UNOBSERVED(...) appears (enforced by the checker, AC-008).

## §D.7 AC-007 — Every cell cites evidence; non-UNOBSERVED cells cite existing raw output

**Given** the 28 cards `evidence/cells/R01-claude.md` through `R14-codex.md`, each cell in `verdict.md` written `<VERDICT> [cell:<ID>-<tool>]`
**When** `sh .moai/reports/t1434/check-verdict.sh .moai/reports/t1434/verdict.md` runs (the script contains no git invocation)
**Then** it prints `CELLS=28 CARDLESS=0 MISSING-FIELDS=0 ORPHAN-RAW=0 EMPTY-REASON=0` and exits 0; `ORPHAN-RAW` counts non-UNOBSERVED cells whose Evidence field cites a file under `evidence/raw/` that does not exist or is empty; `MISSING-FIELDS` counts cards lacking any of Claim, Evidence, Baseline-attribution, Gaps, Residual-risk; `EMPTY-REASON` counts `UNOBSERVED()` cells with no reason.

## §D.8 AC-008 — The verdict checker can fail (negative control)

**Given** a copy `evidence/verdict.mutated.md` in which one non-UNOBSERVED cell references a card whose cited raw file was renamed
**When** `sh .moai/reports/t1434/check-verdict.sh .moai/reports/t1434/evidence/verdict.mutated.md` runs
**Then** exit is non-zero and the output reports `ORPHAN-RAW=1` (or the corresponding counter), recorded in `evidence/negative-control-checker.txt`; a checker that exits 0 on the mutant is a defect against this criterion.

## §D.9 AC-009 — Claude static measurement exists

**Given** the evidence directory after M2
**When** `ls evidence/raw` is read for the files matching `claude-validate-*`, `claude-install-*`, `claude-details-*`
**Then** each of the 14 component kinds has at least a `claude-validate-<ID>` file with an `exit=` line, the composite fixture has `claude-details-composite` containing a token-cost line, and the mod fixture has a `claude-test-R04` file or an `UNOBSERVED(plugin-test-absent)` reason recorded in its card.

## §D.10 AC-010 — Runtime cells carry a positive control or are UNOBSERVED

**Given** each runtime evidence file for R01–R13 (Claude) in `evidence/raw/claude-runtime-<ID>.txt`
**When** `grep -L "CONTROL-OBSERVED" evidence/raw/claude-runtime-R0*.txt evidence/raw/claude-runtime-R1*.txt` runs
**Then** every file it lists belongs to a cell whose verdict is `UNOBSERVED(...)`; a PROJECT-ONLY or PLUGIN-OK cell whose runtime file lacks `CONTROL-OBSERVED` fails this criterion, and the rules rows (R09, R10, R11) additionally record both `SENTINEL-PLUGIN-ABSENT|PRESENT` and `SENTINEL-PROJECT-PRESENT` lines.

## §D.11 AC-011 — Settings-hook equivalence is fully observed

**Given** `evidence/raw/claude-runtime-R08.txt`
**When** `grep -c "PLUGIN-HOOK-EXEC\|PROJECT-HOOK-EXEC\|PROJECT_DIR_EXPANDED=\|PLUGIN_ROOT_EXPANDED=\|TIMEOUT-CUTOFF=" evidence/raw/claude-runtime-R08.txt` runs
**Then** the count is at least 5 (execution of the plugin hook and of the project hook control, both expansions, and the timeout result with its observed seconds), or the R08 Claude cell is `UNOBSERVED(<reason>)` and the card's Gaps field names what blocked it.

## §D.12 AC-012 — Codex measured under a scratch home with a recorded trail

**Given** `evidence/commands.log`
**When** `grep -c "CODEX_HOME=" evidence/commands.log` and `grep -c "codex plugin add" evidence/commands.log` run
**Then** both counts are at least 1; no `codex plugin validate` line exists (the subcommand is absent, per the observed `codex plugin --help`); and every R01–R14 Codex cell is either backed by a raw file or `UNOBSERVED(<reason>)`.

## §D.13 AC-013 — Version agreement and non-goal

**Given** every evidence file begins with `# stamp: claude=<v> codex=<v> fixture_sha256=<h> date=<d>`
**When** `grep -L "^# stamp: claude=" evidence/raw/*` runs, then `grep -c "VERSION-START-END-MATCH" evidence/00-isolation.txt`, then `grep -c "Non-goal: user-global install mode removal" verdict.md`
**Then** the first prints nothing, the second is 1 (the start and end `claude --version` and `codex --version` readings were equal), and the third is 1.

## Edge Cases

- **EC-1 — Isolation partially works.** If `CLAUDE_CONFIG_DIR` redirects the registry but not the login, record both facts; runtime cells use the real-config `--plugin-dir` route under the snapshot protocol.
- **EC-2 — A fixture fails validate.** A failing validate is an observation (record verbatim), not a probe fault; the cell may still be measured at runtime if the plugin loads.
- **EC-3 — A command is refused (permission, guard).** The refusal is named in the card's Gaps with what was done instead; the cell is UNOBSERVED unless an observed channel remains (verification-claim-integrity §3.1).
- **EC-4 — Auto-update changes a CLI version mid-run.** AC-013 fails; the batch is re-run.
- **EC-5 — A model-dependent channel disagrees between its two runs.** The cell is PARTIAL(nondeterministic) or UNOBSERVED; never PLUGIN-OK.

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Restoration | empty diff, manifest rows >= 3, negative control exit 1 | AC-003, AC-004 |
| Verdict integrity | checker exit 0 and mutant exit non-zero | AC-007, AC-008 |
| Coverage | 14 rows x 2 tools, 28 cards | AC-006, AC-007 |
| Observation honesty | every non-UNOBSERVED runtime cell has CONTROL-OBSERVED | AC-010 |
| Product code untouched | empty diff over product paths with non-empty control | AC-001 |
| plan-auditor | PASS at the Tier M threshold | auditor report |

## Definition of Done

- [ ] All 13 acceptance criteria green with verbatim evidence under `.moai/reports/t1434/evidence/`
- [ ] `.moai/reports/t1434/verdict.md` carries the 14-row table, both columns, recommended-home, consequence, version stamps, and the non-goal line
- [ ] Both negative controls (restore diff, verdict checker) observed failing and recorded
- [ ] Real config homes' non-volatile protected rows byte-identical to the pre-run manifest
- [ ] No moai product code, template, `.mcp.json`, or settings change in the card diff
- [ ] Unobservable cells are UNOBSERVED with a reason, none inferred from documentation or from validate output alone
- [ ] plan-auditor PASS recorded before run-phase entry
