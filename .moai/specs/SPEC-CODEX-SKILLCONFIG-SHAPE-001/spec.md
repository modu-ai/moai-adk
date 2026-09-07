---
id: SPEC-CODEX-SKILLCONFIG-SHAPE-001
title: "Codex [[skills.config]] surface measurement — load behavior, path value shapes, and real-environment stale-entry observation on codex-cli 0.153.4"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/codexwiring
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "codex, skills-config, measurement, codex-home-isolation, doctor, prompt-input, t504"
related_specs: [SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-WIRING-001]
---

# SPEC-CODEX-SKILLCONFIG-SHAPE-001 — Codex [[skills.config]] surface measurement

## HISTORY

- 2026-09-07 (plan-phase, v0.1.0) Initial authoring. Card t504 (MEASUREMENT SPEC — no product code changes, Tier S, harness minimal). Card t502 (skills.config work) cannot start safely until the surface is measured; t451's doctor semantics (SPEC-CODEX-SKILL-PATH-001) are built on the same unmeasured premise. This SPEC is authored after SPEC-CODEX-SKILL-PATH-001 landed its path-shape resolution; it does not modify that code — it measures the external Codex behavior the code reasons about.

## §A. User Story

**As a** MoAI-ADK maintainer whose `moai doctor` Codex stale-skill check advises users to "remove the stale entries" based on `[[skills.config]]` parsing, **I want** the actual Codex-side behavior of that config key measured on the shipping binary — whether it loads entries, which `path` value shapes result in skill exposure, and what a dead path does — **so that** the doctor's semantics and card t502's skills.config work rest on observed behavior rather than on an assumption nobody has verified.

**Outcome hypotheses (each settled by measurement, never assumed):**
- RQ1 — Whether codex-cli 0.153.4 loads `[[skills.config]]` entries (path / enabled) from config.toml at all: the surface t451's doctor is built on is either live or dead, and the verdict is one of those two words.
- RQ2 — Which `path` value shapes result in actual skill exposure (existing FILE vs existing DIRECTORY), whether `enabled = false` suppresses exposure, and whether Codex stays silent about a nonexistent path (t451's docstring claims "neither prunes it nor complains" — unverified until now).
- RQ3 — With the real `~/.codex` (49 stale entries) and no isolation, whether prompt-input surfaces any moai skill content and whether Codex complains about the 49 dead paths.

## §B. Context and Background

### §B.1 The motivating defect

The repository's `moai doctor` Codex Wiring check (t451, SPEC-CODEX-SKILL-PATH-001) parses `[[skills.config]]` array-of-tables entries from the user's Codex config.toml (`internal/codexwiring/skills.go` `ParseSkillEntries`) and reports entries whose declared path no longer exists, advising "remove the stale entries or restore the skill files" (`internal/cli/doctor_codex.go` `codexSkillPathShape` + `codexStaleSkillFinding`). The ENTIRE premise — that Codex actually reads and loads this config key — has never been measured. `internal/template/agentemit/agents-codex.yaml` states it outright: "skills.config value set is unmeasured and M1 owns skills".

`internal/codexwiring/skills.go`'s header comment carries the unmeasured claim verbatim: "Codex neither prunes it nor complains". The tri-state `SkillEnabled` parser deliberately refuses to guess Codex's default for an absent key — this SPEC is where that guessing could end.

### §B.2 Measured facts established this session (embedded as given — run-phase re-verifies, never re-derives)

- codex-cli version on this machine: `codex-cli 0.153.4` (self-measured, `codex --version`). Run-phase MUST re-stamp the version at measurement time — a session-carried figure is a carry-over, not a baseline (`verification-claim-integrity.md` §2).
- `codex debug prompt-input` EXISTS: "Render the model-visible prompt input list as JSON", takes [PROMPT], honors `-c key=value`, 0 model calls. Per the 0.152.1 probe recorded in `internal/template/agentemit/agents-codex.yaml`, it "renders the skill roots table".
- User's real `~/.codex/config.toml` (sha256 `c45741c10c3e844ae2fe768787338cfdc659dc41b2e2b75db83f5ebf8638e01a`, 629 lines): exactly 49 `[[skills.config]]` entries, every one `path = "/Users/goos/.codex/skills/moai-*/SKILL.md"`, ZERO duplicates, and ALL 49 target paths are nonexistent (existing=0, missing=49). `~/.codex/skills/` contains only `.system` and `hatch-pet` — no moai-* dirs.
- Precedent methodology (t91 README §4/§9, codex 0.147-era): skills are exposed from `<repo>/.agents/skills/` and `$CODEX_HOME/skills/` (directory convention, confirmed), NOT from `<repo>/.claude/skills/`; symlinked skills expose fine; `CODEX_HOME=<isolated scratch>` + `codex debug prompt-input "hi"` is the established 0-model-call observation harness. The 0.152.1 probe (agents-codex.yaml skill-loader rationale) also used an isolated CODEX_HOME in an isolated /tmp project.
- The `[[skills.config]]` key's WRITER is not in this repository (grep: no writer on main or develop; develop only reads it). The 49 entries are believed to be legacy from a pre-M1 moai skill installer whose files were later cleaned (t91 noted 46 old moai skills in `~/.codex/skills`, dated 2026-06-07, cleanup targeted) — provenance is NOT established and is out of scope (§F).

### §B.3 Method decision (recorded here, honored by the plan)

CODEX_HOME isolation is the primary method: the experiment NEVER writes the user's real `~/.codex/config.toml`. This is deliberately stronger than a backup/restore fallback; the backup/restore plan applies ONLY if the isolation-validity control (cell IV, REQ-CSS-001) fails on 0.153.4. Rationale: a method that never touches the real home cannot corrupt it, whereas a restore-based method is one interrupted run away from a corrupted user config; t91 (§9) and the 0.152.1 probe both measured under isolated CODEX_HOME, so the harness is precedented, not novel.

## §C. Requirements (GEARS)

- **REQ-CSS-001 (event-driven, isolation-validity control)** — **When** the experiment runs its first isolated cell, the harness shall first prove that `CODEX_HOME` isolation is valid on the measured codex-cli version by running control cell IV: a scratch home whose AGENTS.md carries the unique marker `T504IVMARKER` with NO skills config, then `codex debug prompt-input "hi"` with stdout captured; the marker MUST appear in the captured output (grep count ≥ 1). **When** the marker does NOT appear, `CODEX_HOME` isolation is invalid on this version — the harness shall stop, record the failure verbatim, and fall back to the user-home backup/restore method (real-home backup taken before any cell, restored and hash-verified after the last), recording the fallback in the evidence file.
- **REQ-CSS-002 (event-driven, detector baseline / catch-all control)** — **When** control cell N runs (scratch home, NO `[[skills.config]]`, skill fixture named `t504probe` carrying token `T504MARKER` in its SKILL.md frontmatter description and body, placed at `$CODEX_HOME/skills/t504probe/SKILL.md` — the t91-confirmed convention), the harness shall record whether `T504MARKER` appears in the captured output. The V1/V3/V4 verdicts are admissible ONLY when cell N is positive; **When** cell N is negative, the judgment means itself is broken on this version and the verdict of the entire isolated matrix is "measurement void" — reported as void, never interpreted.
- **REQ-CSS-003 (event-driven, RQ1 — load behavior)** — **When** cell V1 runs (scratch home whose config.toml declares exactly one `[[skills.config]]` entry whose `path` points to an EXISTING FILE at a neutral location — under the experiment's tmp root but NOT under `$CODEX_HOME/skills/`, NOT under any cwd `.agents/skills/`, so only the config key can explain exposure), the harness shall record whether `T504MARKER` appears in the captured output; this cell alone answers RQ1 — marker present means the surface is live, marker absent (admissibility per REQ-CSS-002) means the surface is dead on this version.
- **REQ-CSS-004 (event-driven, RQ2 — invalid control and complaint capture)** — **When** cell V2 runs (identical to V1 except the declared `path` does NOT exist), the harness shall record that `T504MARKER` does NOT appear (grep count = 0) AND shall capture any stderr/stdout complaint about the dead path verbatim, bounded; the V1/V3/V4 verdicts are admissible ONLY when cell V2 is negative. **When** both cell N and cell V2 fail in the same direction (N negative AND V2 positive), the verdict is "measurement void" — never an interpretation.
- **REQ-CSS-005 (event-driven, RQ2 — directory and enabled shapes)** — **When** cell V3 runs (declared `path` points to an EXISTING DIRECTORY — the skill directory, not the SKILL.md file) and cell V4 runs (V1's config plus `enabled = false` on the entry), the harness shall record marker presence for each cell independently, discriminating the FILE vs DIRECTORY shape semantics and whether the `enabled = false` declaration suppresses exposure.
- **REQ-CSS-006 (unwanted, real-home invariance)** — The harness shall not write to the user's real `~/.codex` at any point of the isolated matrix; only cell R (REQ-CSS-007) may read it, read-only. `sha256` of `~/.codex/config.toml` recorded before the first cell and after the last cell MUST be identical; a pre-run working copy of the file into the tmp root before cell R is a belt-and-suspenders backup whose restore check must be a no-op (post-run hash equal to the pre-run hash means nothing was restored).
- **REQ-CSS-007 (event-driven, RQ3 — real-environment observation)** — **When** cell R runs (NO `CODEX_HOME` override, the real config with its 49 stale entries, neutral cwd, `codex debug prompt-input "hi"`), the harness shall grep the captured output for `T504MARKER` and any moai skill marker (expected absence: grep count = 0) and shall capture any Codex complaint about the 49 dead paths verbatim, bounded. Nothing in this cell writes to the user home.
- **REQ-CSS-008 (ubiquitous, evidence discipline)** — The harness shall run every cell serially (no concurrency), make zero model calls (`debug prompt-input` only), spawn no background processes, capture stdout and stderr into separate per-cell files under the experiment tmp root, and record into the evidence file `.moai/reports/t504/skills-config-path-shape.md` for every cell: the full command, the verbatim bounded output, the grep counts (including the swept-count census, so an empty sweep is visible), and the codex-cli version stamp taken at run time; the evidence file closes with the 5-section evidence-bearing report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

## §D. Acceptance Criteria (inline, Tier S — mechanically verifiable)

Evidence file (all ACs cite it): `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t504/.moai/reports/t504/skills-config-path-shape.md`. All commands are run-phase actuations of plan.md §C; this section fixes what each AC's observed output must be.

- **AC-CSS-001-01 (isolation-validity)** — Given the cell-IV capture file, When the harness runs `grep -c T504IVMARKER <iv.stdout>`, Then the count is ≥ 1 and the evidence file records the count alongside the command and output. **When** the count is 0, Then the evidence file instead records the IV failure verbatim, the words `ISOLATION INVALID — FALLBACK`, and the backup/restore-method results under the same AC (the fallback path, not a silent stop).
- **AC-CSS-001-02 (detector baseline positive)** — Given the cell-N capture file, When the harness runs `grep -c T504MARKER <n.stdout>`, Then the count is ≥ 1 and the evidence file records the count with its command. The evidence file also names the swept convention explicitly (`$CODEX_HOME/skills/t504probe/SKILL.md`) so the positive is attributable to the convention it tested.
- **AC-CSS-001-03 (invalid control negative + complaint capture)** — Given the cell-V2 captures, When the harness runs `grep -c T504MARKER <v2.stdout>`, Then the count is 0 AND `<v2.stderr>` (and stdout) are preserved in the evidence file whether or not a complaint exists — the absence of a complaint is itself a recorded observation, never an empty slot. Admissibility: V1/V3/V4 verdicts cite AC-01 PASS and this AC as preconditions; if both controls fail in the same direction, the evidence file's verdict line reads `MEASUREMENT VOID` and NO cell verdict is recorded.
- **AC-CSS-001-04 (RQ1 verdict recorded with provenance)** — Given the admissible matrix (AC-01 and AC-02 PASS, AC-03 PASS), When the harness reads the cell-V1 count, Then the evidence file carries an explicit RQ1 verdict — one of `LIVE` (count ≥ 1) or `DEAD` (count = 0) — each attributed to the V1 command, its verbatim bounded output, and the codex-cli version stamp taken at run time.
- **AC-CSS-001-05 (RQ2 shape verdicts recorded)** — Given the admissible matrix, When the harness reads the cell-V3 and cell-V4 counts, Then the evidence file carries, for each: the command, verbatim bounded output, grep count, and a one-line verdict — DIRECTORY loads / does not load, and `enabled = false` suppresses / does not suppress exposure — with the complaint-capture slots for both cells filled the same way as AC-03.
- **AC-CSS-001-06 (no-touch proof)** — Given the two hash lines in the evidence file, When the harness compares `shasum -a 256 ~/.codex/config.toml` captured before the first cell and after the last cell, Then the two digests are byte-identical and both command lines are recorded verbatim; the belt-and-suspenders pre-R backup copy is recorded with the observation that its restore check was a no-op.
- **AC-CSS-001-07 (RQ3 real-env observation recorded)** — Given the cell-R captures, When the harness runs `grep -c T504MARKER <r.stdout>` and greps the output for any moai skill name, Then the expected-absence result (count = 0) or any unexpected presence is recorded either way, together with the verbatim bounded complaint capture for the 49 dead paths; the evidence file states explicitly that no cell other than R touched the real home.
- **AC-CSS-001-08 (evidence report completeness)** — Given the finished evidence file, When the harness greps it for the five section headers (`Claim`, `Evidence`, `Baseline-attribution`, `Gaps`, `Residual-risk`) and for per-cell command lines, Then every one of the 7 cells (IV, N, V1, V2, V3, V4, R) has a command + verbatim bounded output + grep count, the 5 sections are all present, and the Gaps section names anything not observed (e.g. an empty Gaps asserts nothing was left unobserved — it must be true when written).

## §E. Non-Functional Constraints

- **NFC-1 (measurement-only)** — This SPEC changes no product code. No file under `internal/`, `pkg/`, `cmd/` is modified; the deliverables are the measurement verdicts and the evidence file.
- **NFC-2 (read-only posture)** — The real `~/.codex` stays byte-invariant (REQ-CSS-006); every write in the experiment lands under the experiment's own `mktemp -d` tmp root or in the evidence file.
- **NFC-3 (no model calls, no background)** — `codex debug prompt-input` only (0 model calls); cells run serially; no background processes are spawned at any point.
- **NFC-4 (evidence attributability)** — Every verdict names its command, its verbatim output, and the run-time codex-cli version; figures carried from this plan-phase authoring session (e.g. `0.153.4`, the config sha256) are context, never substituted for a run-time re-measurement.

## §F. Exclusions

### Out of Scope — the 49 real stale entries themselves

- Deleting, modifying, pruning, or re-writing any of the 49 real `[[skills.config]]` entries in the user's `~/.codex/config.toml` — that is card t506's scope.
- Repairing or restoring the missing moai-* skill files the 49 entries point at.

### Out of Scope — provenance and history

- Establishing WHO or WHAT wrote the 49 entries (believed legacy pre-M1 installer; grep shows no writer in this repository on main or develop) — the entries' provenance is not settled by this SPEC and does not gate its verdicts.

### Out of Scope — product code and adjacent surfaces

- Any change to `internal/codexwiring/skills.go`, `internal/cli/doctor_codex.go`, or any other product file (NFC-1) — the doctor's semantics are RE-MEASURED, not edited here; a semantic change to the doctor based on this SPEC's verdicts is follow-up card scope (t502 and successors).
- Re-verifying the t91 skill-root conventions (`.agents/skills/`, `$CODEX_HOME/skills/`, `.claude/skills/` non-exposure, symlink behavior) beyond the single convention cell N needs as its detector baseline.
- Codex versions other than the binary measured at run time (this session: 0.153.4) — no cross-version compatibility sweep.
