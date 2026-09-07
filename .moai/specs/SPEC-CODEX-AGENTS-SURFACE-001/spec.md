---
id: SPEC-CODEX-AGENTS-SURFACE-001
title: "Codex agent surface judgment record — model omission retained, per-agent skills.config dropped, global [agents] table not wired (codex-cli 0.153.4 evidence)"
version: "1.0.0"
status: in-progress
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0"
module: internal/template/agentemit
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "codex, agents, agentemit, manifest, judgment-record, t505"
related_specs: [SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-SKILL-LOADER-001, SPEC-CODEX-SKILLCONFIG-SHAPE-001]
---

# SPEC-CODEX-AGENTS-SURFACE-001 — Codex Agent Surface Judgment Record

## HISTORY

- 2026-09-07 (plan-phase, v1.0.0) Initial authoring. Card t505. Documentation-of-judgment SPEC: zero Go behavior changes, zero template content changes (the 11 committed TOMLs stay byte-identical). The only source edit is the mapping manifest `internal/template/agentemit/agents-codex.yaml` — rationale refresh plus one new judgment record. Evidence measured this session on worktree `.claude/worktrees/t505` @ `0b1e27877`: `.moai/reports/t505/probe-01534.md` (17-cell matrix, `[agents]` type map), `.moai/reports/t505/research-fanout-md.md` (4-lens synthesis, contradictions C1–C7), lineage SPECs (SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-SKILL-LOADER-001, SPEC-CODEX-SKILLCONFIG-SHAPE-001). The card closes the t494 candidate ledger: A1 is overturned with evidence (§B.4), A3 is already discharged by the manifest itself, A2/A4 remain held. `priority: P2` follows the measurement/judgment-card series precedent (SPEC-CODEX-SKILLCONFIG-SHAPE-001, card t504).

## §A User Story

**As a** MoAI-ADK maintainer whose agent emitter's Codex-behavior facts were measured against codex-cli 0.147.0 while the shipping binary is 0.153.4, **I want** the three open agent-surface questions — per-agent `model`, per-agent `skills.config`, and the global `[agents]` config.toml table — judged from measured behavior and recorded in the mapping manifest, **so that** a future codex upgrade re-opens the right question with its evidence in hand instead of re-deriving it, or worse, silently re-deciding it.

**The card's [HARD] rule is judge first, wire only what is needed.** Wiring all three surfaces is explicitly NOT the goal. The three dispositions this SPEC records:

1. **Per-agent `model` — KEEP OMITTED** (`fields.model.emit: false`).
2. **Per-agent `skills.config` — KEEP DROPPED** (skill-loader `documented-drop`).
3. **Global `[agents]` table — DO NOT WIRE**, recorded as a per-key judgment that overturns the t494 candidate A1 explicitly.

**Deliverable shape**: manifest-comment-only change. The 11 committed TOMLs under `internal/template/templates/.codex/agents/moai/` must remain byte-identical after regeneration (the `TestCodexAgentsDeployFixture` guard and REQ-CSL-008 prove it). Editing the TOMLs directly is forbidden — they are emitter outputs (regenerate-not-edit).

## §B Evidence Base

### §B.1 Per-agent `model` — omission is a measured decision, and 0.153.4 strengthens it

- **Measured (0.147.0, P-03)** — `SPEC-CODEX-DUAL-AGENTS-001` progress.md:198-201: an omitted `model` key registers and delegates (inherits the subagent default); `model = "t89-bogus-model-string"` registers **silently** at parse. Emitting a Claude alias would be accepted-but-wrong. R-011 omit-model (spec.md:126-131) was confirmed as the only safe choice and is mechanically enforced (AC-009: `golden_test.go:170,205-207`; `agentemit_test.go:264-279`).
- **Docs (0.153.4, tag-pinned)** — the agent-file `model` is precedence-winning over **both** the parent and the `[agents]` default: exactly the two-sources conflict that makes an emitted pin a dual-source hazard.
- **Value churn** — valid model ids are never enumerated; the bundled default changed within a hotfix line (0.153.4 adds "Astra", #42874). A hardcoded moai pin would be fragile.
- **Open precedence bug** — openai/codex#32587 (desktop app ignores the agent-file model) — the documented semantics are themselves unstable.
- **Flipping today is inert** — `renderTOML` has no `model` branch (writer.go:87-155; the only conditional field branches are `model_reasoning_effort` :111 and `sandbox_mode` :122) and the validator ignores unknown `fields` keys (manifest.go:126-137). A real flip = new writer branch + validator work + AC-009 inversion + regeneration — unjustified by any moai-side source.
- **Model-policy axis** — `.claude/rules/moai/development/model-policy.md` has no codex model axis; the profile matrix (33 cells) resolves per-agent effort, not a static model pin, and the profile matrix cannot be baked into committed byte-guarded TOMLs without breaking R-006 determinism (writer.go:1-7 header: no environment-derived values).

### §B.2 Per-agent `skills.config` — an override, not a grant, and unobservable on any measured surface

- **Source (0.153.4, tag-pinned `skills_config.rs`)** — `SkillConfig = {path?, name?, enabled}` is an enable/disable **override**, not a skill grant; `enabled` is required (t504 measured at global scope: a missing `enabled` hard-fails every codex invocation, rc=1).
- **Measured (0.152.1, t452)** — `codex exec` does not read the project `.codex/agents/` directory at all on the non-interactive path (m2-verdict.md:142): flat placement included; a known-bad sandbox control exited 0 silently. Any emitted value is unverifiable.
- **Measured (0.153.4, t505 probes P4–P7)** — agent-TOML `model`, bogus-model, and `skills.config` with and without `enabled` are **all silent** on `debug prompt-input` + `doctor --json` (probe-01534.md F4). The 0.152.1 non-read finding holds. The global-scope hard-fail (t504 F1) did NOT reproduce at agent scope on this surface — but the surface cannot distinguish "not read" from "read leniently", so the ship-omitted rule still applies to both readings.

### §B.3 Global `[agents]` table — measured per-key type map (codex-cli 0.153.4)

Measured this session via `probe-01534.md` (harness.sh + harness2.sh, 17 cells, CODEX_HOME isolation, zero model calls) and confirmed against tag-pinned `config_toml.rs:444` (`pub agents: Option<AgentsToml>`):

| Key | Measured type (0.153.4) | Disposition | Grounds |
|---|---|---|---|
| `enabled` | bool, default true | no-wire | emitting true is a no-op; false would un-register the 11 moai agents |
| `max_threads` | usize (legacy alias of `max_concurrent_threads_per_session`) | no-wire | no moai codex resource policy; limits written into a user-editable config.toml are invasive |
| `max_concurrent_threads_per_session` | usize | no-wire | same as above |
| `max_depth` | source-only (V1-unstable) | no-wire | unstable surface; no moai policy to carry |
| `default_subagent_model` | string | no-wire | a second precedence-winning model knob — the same dual-source hazard as §B.1, session-global |
| `default_subagent_reasoning_effort` | string | no-wire | no-op for moai: all 11 agents already carry per-agent `model_reasoning_effort` via the P-02 identity map, which overrides the global default |
| `interrupt_message` | **bool**, default true | no-wire | docs naming implies a message string — MEASURED mismatch (string value → rc=1 "expected a boolean"; bool value → rc=0); no moai use |
| `job_max_runtime_seconds` | no-op per source | no-wire | the source marks it a no-op |
| unknown flat keys | parsed as role entries (`struct AgentRoleToml`) | hazard note | any wrong-typed known key **bricks codex startup rc=1** (P2/B3b/B6 verbatim errors); unknown flat keys are role slots, not settings |
| `[agents.<name>]` `{config_file, description}` | role registration | no-wire | redundant with `.codex/agents/` auto-discovery (t91 P-04 measured; t494 A4 hold) |

Schema facts: the table is real and **strictly parsed** — the schema check is alive (the bad-typed controls went red by design, P2/B3b/B6), and `interrupt_message` is a documented-schema mismatch (F2). `default_subagent_model`/`default_subagent_reasoning_effort` parse but leave zero prompt-input trace (F3) — their effect is only observable at delegation time, which is out of scope for a no-model-call measurement discipline.

### §B.4 The t494 A1 overturn — explicit, evidence-carrying

t494 candidate A1 (`.moai/reports/t494/codex-doc-survey.md`, §5(d)) recommended **adopting** the `[agents]` table (`default_subagent_model` / `default_subagent_reasoning_effort` / `max_concurrent_threads_per_session`) on a documentation read, before the measured type map and the dual-source semantics were known. This SPEC overturns A1 with evidence: every key A1 named fails the wire test — `default_subagent_model` is a second precedence-winning model source (dual-source hazard, session-global), `default_subagent_reasoning_effort` is a no-op against moai's per-agent keys, and the thread-cap keys would write limits into a user-editable file for which moai has no policy. The overturn is recorded here and in the manifest judgment record (REQ-CAS-003) — never silently. A3 ("is the model omission intentional?") was already discharged by the manifest itself (class-`model` `omit` rationale, agents-codex.yaml:190-194); A2/A4 remain held.

### §B.5 Wiring surface reality

moai's only writers to `.codex/config.toml` are `EnsureMCPTable` (internal/codexwiring/configtoml.go:85) and `EnsureStatusLine` (:106), existence-gated in internal/cli/update_codex_wiring.go. No SPEC or Go code references `[agents]` keys (repository grep, research finding 3). Wiring the table would add a third merge surface into a user-owned file with new idempotence obligations — rejected here on top of the per-key dispositions.

## §C Requirements (GEARS)

### REQ-CAS-001 — Per-agent model omission retained (Ubiquitous) `[AC-CAS-003, AC-CAS-006]`

The mapping manifest shall retain `fields.model.emit: false` (agents-codex.yaml:28-34), the class-`model` `omit` disposition (:190-194), and the `model-pin-manager-git` documented drop (:214-218); their rationales shall carry the 0.153.4 evidence set of §B.1 — agent-file `model` precedence over parent and `[agents]` default, open-bug #32587, per-release model-id churn (#42874), and the P-03 measured accepted-but-wrong decision (0.147.0).

### REQ-CAS-002 — Per-agent skills.config drop retained (Ubiquitous) `[AC-CAS-003, AC-CAS-006]`

The mapping manifest shall retain the skill-loader `documented-drop` disposition (agents-codex.yaml:117-139) and its rationale shall carry the 0.153.4 evidence of §B.2: the tag-pinned `SkillConfig = {path?, name?, enabled}` override semantics (not a grant; `enabled` required), the t505 P4–P7 measured silence of agent-TOML `model` / bogus-model / `skills.config` with-and-without-`enabled` on `debug prompt-input` + `doctor --json`, and the retained 0.152.1 measured non-read (t452 m2-verdict.md:142). The existing re-probe clause ("Re-probe when an observable agent-role surface exists") shall survive the refresh.

### REQ-CAS-003 — Global `[agents]` no-wire judgment record (Ubiquitous) `[AC-CAS-001, AC-CAS-002, AC-CAS-004]`

The mapping manifest shall carry a judgment record for the global `[agents]` config.toml table: a `documented_drops` entry with `id: codex-global-agents-table` plus an adjacent YAML comment block reproducing the measured per-key type map of §B.3 (keyed line-per-key, stamped codex-cli 0.153.4). The entry's rationale shall state the no-wire disposition, the dual-source hazard (`default_subagent_model`), the rc=1 strict-parse hazard, the auto-discovery redundancy, and the **explicit t494 A1 overturn** with its grounds (§B.4). The record shall require no Go change: `documented_drops` entries parse into the existing `DocumentedDrop` struct (manifest.go:73) and nothing consumes them at emission time.

### REQ-CAS-004 — Zero emission delta on manifest change (Event-driven) `[AC-CAS-005, AC-CAS-006]`

**When** the mapping manifest's Codex-behavior rationales change, the regeneration gate shall prove a zero emission delta before commit: `make agents-emit` regenerated all 11 TOMLs byte-identically (`git diff --stat -- internal/template/templates/.codex` empty), and `make agents-emit-check` exited 0 (REQ-CSL-008 obligation — SPEC-CODEX-SKILL-LOADER-001 progress.md:117). The 11-count and byte-identity are mechanically enforced by `TestCodexAgentsDeployFixture` (internal/template/codex_agents_deploy_test.go:60-62,74-88), which stays green.

### REQ-CAS-005 — Measured-version stamp retention (State-driven) `[AC-CAS-004]`

**While** the top-level `codex_measured_version` remains `"0.147.0"` (agents-codex.yaml:14), the manifest shall record axis-wise measured versions inside the rationales — the skill-loader rationale keeps its 0.152.1 stamp and gains a 0.153.4 stamp, the `[agents]` type map carries a 0.153.4 stamp — and shall **not** raise the top-level stamp. Raising it would claim coverage of axes not re-measured this session: sandbox_mode, mcp_servers, layout, and the effort map (AC-CSL-009 recorded decision — SPEC-CODEX-SKILL-LOADER-001 progress.md:118).

## §D Acceptance Criteria (inline — Tier S)

Classification: AC-CAS-001/002/003 are release-blocking adoption flips (two-cell, RED-now re-executable by a single grep). AC-CAS-004 is a preservation guard whose meaning is co-observed with AC-CAS-001/003 (they prove the file changed between the two measurements). AC-CAS-005/006 are regression/proof guards whose red direction is defined by the named mutants — they are not adopted on a starting green alone.

### AC-CAS-001 — `[agents]` judgment record exists

RED-now ledger:

```text
AC: AC-CAS-001 (RED-now)
command: grep -c 'codex-global-agents-table' internal/template/agentemit/agents-codex.yaml
stdout: 0
exit_code: 1
tree: 0b1e27877 (worktree .claude/worktrees/t505, branch WT-codex-agents-table)
measured: 2026-09-07, plan-phase session
why_red: the id is a token this SPEC introduces; count 0 means the judgment record does not
         exist (a positive-scoped selector, not a vacuous one)
green_path: M1 adds the documented_drops entry id codex-global-agents-table → the same
            command returns >= 1, rc 0
```

### AC-CAS-002 — the judgment record carries the measured type map (mutant depth guard)

RED-now ledger:

```text
AC: AC-CAS-002 (RED-now)
command: grep -cE 'interrupt_message|default_subagent_model|max_concurrent_threads_per_session' internal/template/agentemit/agents-codex.yaml
stdout: 0
exit_code: 1
tree: 0b1e27877 (worktree .claude/worktrees/t505, branch WT-codex-agents-table)
measured: 2026-09-07, plan-phase session
why_red: none of the three measured keys appears anywhere in the manifest today; count 0 is
         the exact absence the record fills
green_path: M1's type-map comment block names the keys line-per-key → the same command
            returns >= 3 (three distinct keys on three distinct lines — a per-key map, not a
            one-line mention)
```

### AC-CAS-003 — the 0.153.4 evidence refresh landed at all three sites

RED-now ledger:

```text
AC: AC-CAS-003 (RED-now)
command: grep -c '0\.153\.4' internal/template/agentemit/agents-codex.yaml
stdout: 0
exit_code: 1
tree: 0b1e27877 (worktree .claude/worktrees/t505, branch WT-codex-agents-table)
measured: 2026-09-07, plan-phase session
why_red: the file cites 0.147.0 (:14,:17,:36,:46,:177,:245), 0.152.1 (:120) and 0.150.1
         (:164) but never 0.153.4 — the evidence refresh has not landed
site_lines_recheck: grep -n '0\.147\.0\|0\.152\.1\|0\.150\.1' internal/template/agentemit/agents-codex.yaml
         → hits at 14 / 17 / 36 / 46 / 120 / 164 / 177 / 245 (verbatim, measured 2026-09-07
         on this same tree 0b1e27877; replaces mis-transcribed line numbers per plan-audit F-1)
green_path: M1 cites 0.153.4 in the model rationale, the skill-loader rationale, and the
            [agents] judgment record → the same command returns >= 3; site placement is
            evidenced per-site in the M4 verdict record — the >= 3 total alone does not
            force placement (corrected threshold note below)
```

### AC-CAS-004 — measured-version stamp preserved

```text
AC: AC-CAS-004 (preservation guard)
command: grep -c 'codex_measured_version: "0.147.0"' internal/template/agentemit/agents-codex.yaml
observed (pre-M1): 1, rc 0 — tree 0b1e27877, measured 2026-09-07
expected (post-M1): 1, rc 0 — unchanged
vacuous-green note: this guard alone is satisfiable by never touching the file; it is
          non-vacuous only co-observed with AC-CAS-001 and AC-CAS-003, which prove the same
          file changed (0 -> >=1) between the two measurements
```

### AC-CAS-005 — zero emission delta (REQ-CSL-008 proof, regression guard)

```text
AC: AC-CAS-005 (regression guard — red direction defined by mutants)
commands (M2, in order):
  1. make agents-emit
  2. git diff --stat -- internal/template/templates/.codex   -> empty stdout, rc 0
  3. control: git diff --stat -- internal/template/agentemit/agents-codex.yaml
       -> NON-empty after M1 (proves the diff command sees changes; an empty-output guard
          without this catch-all control asserts nothing)
expected: step 2 empty + step 3 non-empty + TestCodexAgentsDeployFixture green
mutant C: append one comment line to a committed TOML -> step 2 non-empty and the deploy
          fixture goes red (codex_agents_deploy_test.go:82-84) -> revert -> byte-identical
mutant D: change a PARSED manifest value (not a comment, e.g. sandbox_mode value) ->
          regeneration produces changed TOMLs -> step 2 non-empty; proves the proof is not
          "green regardless of what was edited"
```

### AC-CAS-006 — package tests stay green (regression guard)

```text
AC: AC-CAS-006 (regression guard)
commands:
  1. go test ./internal/template/... -count=1   -> ok lines, rc 0
  2. make agents-emit-check                     -> rc 0 (drift aborts the build: Makefile:34,47-49)
expected: both green after M1-M3; the inherited guards keep their pre-existing meaning:
  manifest parse (AC-013 fail-closed, manifest.go:112-120), golden AC-009 no-model
  (golden_test.go:170,205-207), TestEmitAllOmitsModel (agentemit_test.go:264-279),
  TestCodexAgentsDeployFixture 11-count + byte identity (codex_agents_deploy_test.go:60-62,74-88)
```

### Mutant-probe record (adoption discipline)

- **Mutant A** — a one-line `[agents]` record with no type map ("we don't wire [agents]"): passes AC-CAS-001, **fails AC-CAS-002** (0 named keys < 3). AC-CAS-002 is the depth guard.
- **Mutant B** — delete the whole `[agents]` block: fails AC-CAS-001 and AC-CAS-002 together.
- **Mutant C/D** — see AC-CAS-005.
- **Threshold note (corrected per plan-audit F-2)** — the ≥3 threshold does NOT force site placement: a refresh that double-cites 0.153.4 in only two sites also passes it. Site placement is enforced procedurally instead — the M4 verdict record must carry a per-site count for each of the three required sites (model rationale, skill-loader rationale, [agents] judgment record).

## Out of Scope

### Out of Scope — emitter field additions

- No `fields.model` flip, no `skills.config` emission, no `[agents]` emission. Zero Go changes: no writer branch, no validator change.

### Out of Scope — codexwiring changes

- `EnsureMCPTable` / `EnsureStatusLine` (internal/codexwiring/configtoml.go:85,106) stay untouched; no new merge surface into `.codex/config.toml`.

### Out of Scope — full-manifest 0.153.4 re-measurement of other axes

- sandbox_mode, mcp_servers, layout, and the effort map are NOT re-measured this SPEC — that is a separate candidate card (t494 S4 style). This exclusion is the direct reason the top-level `codex_measured_version` stays `"0.147.0"` (REQ-CAS-005).

### Out of Scope — doctor enhancements

- Any `moai doctor` Codex-check changes are t504-lineage downstream work, separate ownership.

### Out of Scope — skills provisioning on skills.config

- The t504 recorded downstream directive stands: do NOT build skill provisioning on `skills.config`; the live channels are `$CODEX_HOME/skills/`, `.agents/skills/`, and plugin roots.

### Out of Scope — direct TOML edits

- The 11 committed TOMLs are emitter outputs (regenerate-not-edit); the byte guard enforces identity. Any future emission change goes through the manifest + regeneration, never hand edits.

## §E Evidence Ledger

- **Claim**: the three dispositions in §A rest on measured behavior (t505 probe 0.153.4; t452 0.152.1; P-01/P-02/P-03 0.147.0; t504 global-layer measurement), tag-pinned source reads, and official docs — and the RED-now cells of AC-CAS-001/002/003 were measured on this tree this session.
- **Evidence**: `.moai/reports/t505/probe-01534.md` (17-cell matrices, verbatim errors, type map, Gaps/Residual-risk); `.moai/reports/t505/research-fanout-md.md` (4-lens synthesis, C1 resolved by direct code read — writer.go:87-155, manifest.go:126-137); t452 m2-verdict.md:142; `SPEC-CODEX-DUAL-AGENTS-001` spec.md:126-131, progress.md:198-201, acceptance.md:177; `SPEC-CODEX-SKILL-LOADER-001` progress.md:117-118; t504 skills-config-path-shape.md (primary checkout); t494 codex-doc-survey.md §5(d); Makefile:34,38,47-49. New measurements this session: the four grep cells in §D (command, stdout, exit code, tree each) and the SPEC-ID uniqueness check (0 matches, rc=1).
- **Baseline-attribution**: worktree `.claude/worktrees/t505` @ `0b1e27877`, branch WT-codex-agents-table, 2026-09-07.
- **Gaps**: delegation-time behavior (model-call-bearing paths) unmeasured — the probe's Gaps 1–4 are inherited here; the type map is 0.153.4-single-version; the `interrupt_message` bool judgment rests on error-message-based observation. This plan phase wrote no files outside `.moai/specs/SPEC-CODEX-AGENTS-SURFACE-001/`.
- **Residual-risk**: the codex agent schema changes per release — the `[agents]` type map is 0.153.4-bounded and any future wiring review MUST re-measure at the deciding version (the judgment record states this premise). AC-CAS-004's vacuous-green hazard is contained by co-observation with AC-CAS-001/003, which the artifact authoring keeps in the same document and measurement round.
