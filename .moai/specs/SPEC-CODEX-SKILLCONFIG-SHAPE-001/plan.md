# plan.md — SPEC-CODEX-SKILLCONFIG-SHAPE-001

> Tier S, MEASUREMENT SPEC (no product code changes), card t504. Harness minimal.
> Tree: worktree `.claude/worktrees/t504`, branch `WT-codex-skillpath-shape`, base `ace1c5440`.
> Evidence deliverable: `.moai/reports/t504/skills-config-path-shape.md` (card t504's declared evidence path).

## §A. Context

### §A.1 What is being measured and why

`moai doctor`'s Codex stale-skill check (t451) parses `[[skills.config]]` from the user's config.toml and advises removal of entries whose path is missing. The premise that Codex loads this key at all is unmeasured — `internal/template/agentemit/agents-codex.yaml` says so explicitly, and `internal/codexwiring/skills.go`'s "Codex neither prunes it nor complains" claim is unverified. Card t502 (skills.config work) is blocked on this measurement. This SPEC runs the measurement; it edits no code.

### §A.2 Method decision — CODEX_HOME isolation, backup/restore only as fallback

**Primary method: CODEX_HOME isolation.** Every isolated cell runs under a scratch `CODEX_HOME` created under one `mktemp -d` root (e.g. `/tmp/t504-<random>/`), one scratch home per variant — no cross-contamination. The user's real `~/.codex/config.toml` is never written; the only cell that reads it is R, read-only.

**Why stronger than backup/restore:** a method that never touches the real home cannot corrupt it; a restore-based method is one interrupted run away from exactly that. **Fallback trigger:** if the isolation-validity control (cell IV) fails — the `T504IVMARKER` in the scratch home's AGENTS.md does not appear in prompt-input output — `CODEX_HOME` isolation is invalid on this codex-cli version, and the dispatch's backup/restore plan applies instead (real-home backup before any cell, hash-verified restore after the last), with the fallback recorded in the evidence file.

**Precedent:** t91 (README §4/§9) measured skills exposure under isolated `CODEX_HOME` + `codex debug prompt-input "hi"` (0 model calls) on 0.147-era codex; the 0.152.1 probe in agents-codex.yaml used the same harness shape in an isolated /tmp project. This is a precedented harness, not a novel one.

### §A.3 Facts carried from plan-phase authoring (context — run-phase re-measures, never substitutes)

- codex-cli `0.153.4` on this machine (re-stamp at run time).
- Real config sha256 `c45741c1...8638e01a`, 629 lines, 49 `[[skills.config]]` entries, all `path = "/Users/goos/.codex/skills/moai-*/SKILL.md"`, all 49 targets nonexistent, zero duplicates.
- t91-confirmed conventions: skills expose from `<repo>/.agents/skills/` and `$CODEX_HOME/skills/`; `debug prompt-input` renders the skill roots table.

## §B. Known Issues (filtered, Tier S — measurement-shaped)

- **Vacuous-green hazard (D1 family)**: every cell verdict needs a swept-count census. `grep -c` printing 0 on an EMPTY capture is indistinguishable from a true negative — each cell's evidence row names the capture file's byte size alongside the count, so a 0-count on a 0-byte file is reported as "not measured", never as a verdict.
- **Empty-output masquerade (t485 lesson)**: `grep -c` exits 1 when the count is 0 — expected for V2/R. Cells record the COUNT, not the exit code; a cell script that gates on grep's exit code would invert the verdict.
- **B8 working-tree hygiene**: the only repo writes are the evidence file under `.moai/reports/t504/`; stage by explicit pathspec.
- **No local full suite**: no code changes; no `go test` in this SPEC's verification.
- **zsh PIPESTATUS void**: cell commands that pipe must be structured so the recorded count comes from the file on disk (`grep -c <token> <capture-file>`), never from an inline pipe verdict.

## §C. Pre-flight (run before any cell)

```bash
codex --version                                        # re-stamp the version (do NOT carry 0.153.4 as a baseline)
codex debug prompt-input --help 2>&1 | head -20        # confirm the subcommand exists and takes [PROMPT] on THIS binary
shasum -a 256 ~/.codex/config.toml                     # baseline hash (recorded verbatim in the evidence file)
EXP_ROOT=$(mktemp -d /tmp/t504-XXXXXXXX)               # one tmp root for the whole experiment; echoed into the evidence file
```

Cell-layout plan under `$EXP_ROOT`:

```
$EXP_ROOT/
  home-IV/  home-N/  home-V1/  home-V2/  home-V3/  home-V4/   # one scratch CODEX_HOME per cell
  proj/                                                       # neutral cwd for ALL codex invocations (no .agents/ anywhere above it)
  neutral/t504probe/SKILL.md                                  # V1/V2/V3/V4 fixture target — outside every known skill root
  cells/<cell>.out  cells/<cell>.err                          # per-cell captures, stdout and stderr separate
  backup-config.toml                                          # pre-R belt-and-suspenders copy of the real config
```

Fixture discipline: every skill fixture is named `t504probe` and carries the unique token `T504MARKER` in BOTH the SKILL.md frontmatter description and the body. The IV home's AGENTS.md carries `T504IVMARKER` on its own line. The neutral fixture path is derived from `$EXP_ROOT` so it can never accidentally coincide with `$CODEX_HOME/skills/` or a cwd `.agents/skills/`.

## §D. Constraints (DO NOT VIOLATE)

- **No real-home writes.** Only cell R reads `~/.codex/config.toml`; NOTHING writes to `~/.codex` at any point of the isolated matrix (REQ-CSS-006). The pre-R backup copy is a read of the file into `$EXP_ROOT`.
- **Serial, no background, no model calls.** Cells run one at a time; `codex debug prompt-input` only; no background processes (NFC-3).
- **Stop conditions, recorded not guessed** (REQ-CSS-001/-002/-004): IV negative → stop the isolated matrix, fall back to backup/restore, record. N negative → the isolated verdict is `MEASUREMENT VOID`. N negative AND V2 positive (same direction) → `MEASUREMENT VOID`, no cell interpretation.
- **Evidence format** (REQ-CSS-008): per-cell command + verbatim bounded output + grep counts (with capture byte size); 5-section evidence-bearing report at close (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk); codex-cli version stamped at run time.
- **No product code edits** (NFC-1) — the run touches `$EXP_ROOT` and `.moai/reports/t504/` only.

## §E. Self-Verification (manager-develop deliverable)

- E-item matrix over spec.md §D AC-CSS-001-01 .. -08, each reported in the 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) with the command + verbatim output + tree SHA.
- Every cell row carries: the command, the capture file path, its byte size, the `grep -c` count, and the count's interpretation gate (which controls it depends on).
- The version stamp line: `codex --version` output, verbatim, dated this run.
- The no-touch proof: both `shasum -a 256 ~/.codex/config.toml` lines verbatim, and the equality statement.
- Gaps section names explicitly what was NOT observed (e.g. if IV failed and fallback ran, the isolated-matrix cells' non-observations are the Gaps).

## §F. Milestones (measurement order — decision-reversibility first: the controls that can void everything come before any verdict cell)

- **M1 — Harness preparation.** Pre-flight (§C): version stamp, subcommand existence, baseline hash, tmp root, all six scratch homes, fixtures (IV AGENTS.md marker; N's skill at `$CODEX_HOME/skills/t504probe/SKILL.md`; the neutral fixture file and directory for V1-V4), V2's deliberately-absent path recorded. Nothing runs yet.
- **M2 — Isolated matrix, controls first.** In this exact order, each cell one `codex debug prompt-input "hi"` invocation from `$EXP_ROOT/proj` with `CODEX_HOME` set to that cell's home, stdout/stderr captured to `cells/<cell>.{out,err}`: **IV** → grep `T504IVMARKER` (stop conditions apply). **N** → grep `T504MARKER` (void condition applies). **V2** → grep `T504MARKER` (must be 0) + preserve any complaint. Then the verdict cells: **V1** (RQ1: does the config key drive exposure — the neutral path means only the config key can explain a marker), **V3** (directory shape), **V4** (`enabled = false`). All admissibility gates evaluated against the recorded counts before any verdict sentence is written.
- **M3 — Real-environment observation + close.** Copy the real config to `$EXP_ROOT/backup-config.toml` (belt-and-suspenders). Cell **R**: NO `CODEX_HOME` override, cwd `$EXP_ROOT/proj`, `codex debug prompt-input "hi"` → grep for `T504MARKER` and moai skill names (expected 0), capture any complaint about the 49 dead paths. Post-run hash check. Assemble `.moai/reports/t504/skills-config-path-shape.md`: per-cell rows, RQ1/RQ2/RQ3 verdicts (or the VOID verdict), 5-section evidence-bearing report. Commit evidence by explicit pathspec.

## §G. Anti-Patterns

- Reading a 0 grep count on a 0-byte capture as a negative verdict — that is "not measured", not "absent" (the empty-sweep rule).
- Gating cell logic on grep's exit code (inverts exactly the expected-negative cells V2/R).
- Interpreting V1/V3/V4 after a failed control "because the result looks informative anyway" — the VOID verdict exists precisely to forbid this.
- Writing to the real `~/.codex` "just to test one thing" — HARD violation of REQ-CSS-006.
- Running cells concurrently or in the background to save wall-time — serial is a requirement (NFC-3), not a preference.
- Substituting the session's measured facts (§A.3) for run-time re-measurement in the evidence file (carry-over is not a baseline).
- Editing doctor/codexwiring code while "in the neighborhood" — the measurement may reveal the doctor's semantics are wrong; that finding goes in the evidence file as t502 input, it is not fixed here (scope discipline).

## §H. Cross-References

- spec.md §B.2 measured facts; §C REQ-CSS-001..008; §D AC-CSS-001-01..08.
- Predecessor: SPEC-CODEX-SKILL-PATH-001 (t451/t468 — the doctor semantics this SPEC grounds); SPEC-CODEX-WIRING-001 (doctor surface owner).
- Downstream: card t502 (skills.config work — unblocked by RQ1/RQ2), card t506 (the 49 real stale entries — unblocked by RQ3's observation shape).
- Method precedent: `.moai/reports/t91/README.md` §4 (skill-root conventions) and §9 (isolated `CODEX_HOME` + `debug prompt-input` harness) — primary checkout.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1/§2 — verdicts attributable to run-time commands only.
