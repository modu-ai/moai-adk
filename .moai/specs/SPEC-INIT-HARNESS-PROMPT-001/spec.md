---
id: SPEC-INIT-HARNESS-PROMPT-001
title: "Ask the agent-harness selection in the init wizard, keeping the --agent flag as the non-interactive path"
version: "0.1.3"
status: in-progress
created: 2026-09-01
updated: 2026-09-01
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: "internal/cli + internal/cli/wizard"
lifecycle: spec-anchored
tags: "init, wizard, agent-harness, codex, mcp-provision, flag-over-wizard-precedence"
related_specs: [SPEC-CODEX-WIRING-001, SPEC-CODEX-INIT-001, SPEC-INIT-WIZARD-REPAIR-001, SPEC-AUTONOMY-TIERS-001, SPEC-MCP-DEFAULT-ON-001]
tier: M
---

## HISTORY

- 2026-09-01 — v0.1.0 drafted (manager-spec, card t393 plan-phase). All ground truth re-measured at worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t393`, branch `WT-init-harness-prompt`, HEAD `2c18091d127cbc723074124e1015353e077300ca`, tree `a33625497c87246572f823cf427fe6335a79c825`. Two dispatch premises were corrected during measurement and are recorded in §2.
- 2026-09-01 — v0.1.1 audit-revision round (plan-auditor iteration 1: FAIL 0.70, MP-7 FAIL). Every coordinate below was re-opened at the same HEAD/tree and re-measured by the author; five figures drifted and are corrected in place. Substantive changes: the MCP consumer's observable is re-specified from file state to the stdout announcement (§4 D2, REQ-IHP-009) because `.mcp.json` is template-deployed and exists either way; REQ-IHP-007 now enumerates all **three** `codexwiring.Wire` artifacts; §2 now states that this change **narrows** SPEC-CODEX-WIRING-001 REQ-CW-001 rather than "extends and does not reverse" it; the flag's cobra default is corrected to the empty string; the 16-question figure is now an executed count.
- 2026-09-01 — v0.1.2 audit-revision round 2 (plan-auditor iteration 2: PASS-WITH-DEBT 0.85, all seven must-pass PASS; five textual defects). Corrections, each re-measured by the author at the same HEAD/tree: the AC-IHP-005 justification is narrowed — `runWizardFn` is the sole issuer of the HARNESS QUESTION, not of prompts (a second `huh` prompt at `internal/cli/init.go:598-613` is named as the measured counterexample); the non-interactive announcement observable is pinned ABSENT with the AC-CW-004 disambiguation (REQ-IHP-007, acceptance.md AC-IHP-006a); §8's REQ-CW-001 line no longer says "extended here" (it contradicted §2 C3's narrowing); REQ-IHP-005 now asks for what its AC verifies, with the AC-CI-004 deviation stated at the requirement level; and the `InitQuestions` coordinate is restored to `:296-303` — the v0.1.1 round over-corrected a coordinate that was already right.
- 2026-09-01 — v0.1.3 coordinate correction. Two errors introduced by the v0.1.2 R1 text, both re-measured by the author at HEAD `2c18091d1`: (1) the profile-setup block was cited as `internal/cli/init.go:598-612` and actually spans **`:598-613`** — three closing braces in a row (`:611` closes the `runProfileSetup` error check, `:612` the `confirmForm.Run()` check, `:613` the `if !nonInteractive && isatty.IsTerminal(...)` block itself), so the citation stopped one short; (2) "twelve lines above the wizard gate" was false — the gate is `:644`, which is **31 lines** from the block's closing brace at `:613` (46 from its start at `:598`); the acceptance.md sentence now states the distance with the endpoint it was measured from, and the load-bearing claim is the ordering (it runs BEFORE the gate), which needs no distance at all.

---

## §1 Problem

`moai init` runs one interactive wizard. The agent-harness selection — which harness MoAI wires for the project (`claude`, `codex`, or `both`) — is **not among its questions**. It is reachable only through the `--agent` flag, so a user who goes through the wizard is never asked, and silently receives the `claude` default.

Measured at HEAD `2c18091d1` (tree `a33625497`):

- The wizard's interactive init set is assembled by `InitQuestions` (`internal/cli/wizard/questions.go:296-303`) as `DefaultQuestions` (5 questions) + `Page3Questions` (11 questions) = **16 interactive questions**. None carries a harness or agent axis:

  ```
  $ grep -rn 'ID: *"agent\|ID: *"harness\|AgentWiring\|agent_wiring' internal/cli/wizard/ | wc -l
  0
  ```

- The flag is declared at `internal/cli/init.go:132` as `initCmd.Flags().String("agent", "", …)` — **its cobra default is the empty string**, not `claude`; `claude` is `resolveAgentWiring`'s fallback (`:151-158`) and the help string's prose. The distinction is load-bearing for §4 D1's helper signature and for `validateInitFlags`, which short-circuits on `agent != ""` (`:393-401`).
- `resolveAgentWiring` (`internal/cli/init.go:151-158`) reads **only** the cobra flag (`getStringFlag(cmd, "agent")`); an empty or unrecognized value falls back to `claude`.
- It has exactly two production consumers: `wireCodexUnlessClaude` (`internal/cli/init.go:167`, invoked at `:925`, delegating to `codexwiring.Wire`, which writes **three** artifacts — `.codex/hooks.json`, `.codex/config.toml`, and the trust sidecar `.moai/state/codex-wiring.json`) and the `.mcp.json` provisioning precedence switch (`internal/cli/init.go:911-916`).

The wizard's answers reach `opts` only inside the interactive block at `internal/cli/init.go:644`. Because `resolveAgentWiring` takes `*cobra.Command`, **a wizard answer cannot reach either consumer as the code stands** — this, not the question text, is the change's centre of gravity (§4 D1).

---

## §2 Corrections to the card's premises (measured, not assumed)

Two statements in the originating card do not survive measurement at this HEAD. They are recorded here because a SPEC written on either would misdescribe the change.

**C1 — "~23 wizard questions" counts a set the init wizard never renders.** `questions.go` declares 23 question IDs, but 7 of them are the Git block (`GitQuestions`), which `InitQuestions` deliberately excludes and only `ReconfigureQuestions` splices in (`questions.go:268-287`). The interactive `moai init` set is **16**.

Executed count (per-constructor `ID:` literals, this run at tree `a33625497`):

```
$ awk '/^func DefaultQuestions/{f=1} f&&/[[:space:]]ID:[[:space:]]/{c++} f&&/^}/{print c; exit}' internal/cli/wizard/questions.go
5
$ awk '/^func Page3Questions/{f=1} f&&/^\t\t\tID:/{c++} f&&/^}/{print "Page3Questions ID: literals = "c; exit}' internal/cli/wizard/questions.go
Page3Questions ID: literals = 11
$ awk '/^func GitQuestions/{f=1} f&&/ID:/{c++} f&&/^}/{print "GitQuestions ID: literals = "c; exit}' internal/cli/wizard/questions.go
GitQuestions ID: literals = 7
```

5 + 11 = 16, and 5 + 11 + 7 = 23 reconciles the whole file. **Stated scope of this measurement**: the `awk` commands count `ID:` literals inside each constructor body. That equals the returned slice length only under a premise the counts themselves do not establish — that each element carries exactly one `ID:` and neither constructor loops — and **that premise was established by reading the two constructor bodies, not by counting**. The distinction is not pedantic: a crude whole-file `grep -c` over these constructors returns different numbers depending on the pattern's indentation anchor, so a count taken without opening the file would have produced a false figure. This is a source-derived count, not a runtime `len()`. No test asserts the cardinality (`grep -rn 'len(InitQuestions' internal/cli/wizard/*_test.go` → no count assertion), so no guard would notice a question silently dropped.

**C2 — "only `mcp_provision` has a flag-over-wizard precedent" is false; three same-axis precedents exist,** and one of them is the pattern this SPEC copies:

| Axis | Shape | Location |
|---|---|---|
| `worktree_auto_create` | inline `if !cmd.Flags().Changed(...)` guard | `internal/cli/init.go:261` |
| `autonomy_tier` | named helper covering BOTH entry paths in one call | `internal/cli/init_autonomy_wizard.go:34`, called at `internal/cli/init.go:717` |
| `mcp_provision` × `--agent` | **cross-axis**: axis-A flag overriding an axis-B wizard answer | `internal/cli/init.go:911-916` |

The `autonomy_tier` helper is the correct precedent: it was landed by SPEC-INIT-WIZARD-REPAIR-001 REQ-005 under the explicit name "flag-over-wizard precedence", and one call covers the interactive and non-interactive paths alike. `mcp_provision` × `--agent` is a *different* shape and is the subject of §4 D2 rather than a template to copy.

**C3 — this change NARROWS SPEC-CODEX-WIRING-001 REQ-CW-001; it does not merely extend it.** Read verbatim this run at `.moai/specs/SPEC-CODEX-WIRING-001/spec.md:229-231`:

> **플래그 부재 하에서** `moai init`은 오늘의 동작과 동일하게 수행한다 — 동일한 템플릿 배포, 동일한 `.mcp.json` 기본 provisioning(`provisionMCPEntryUnlessDeclined` 경로), `.codex/hooks.json` 및 `.codex/config.toml` 미생성.

The clause is **unconditional over the flag-absent path**. It carries no `--non-interactive` qualifier; only its verifying criterion AC-CW-004 happens to use one. After this change, a flag-absent **interactive** run whose wizard answer is `codex` **does** create the Codex wiring artifacts — which the clause as written forbids.

Stated plainly, because the earlier draft of this paragraph said the opposite and would have misled a future reader:

- **What is superseded**: REQ-CW-001's flag-absent clause is narrowed from `flag-absent` to `flag-absent ∧ wizard-not-run` (equivalently `flag-absent ∧ (--non-interactive ∨ no TTY)`).
- **The sub-path it no longer covers**: `flag-absent ∧ interactive ∧ wizard-selects-codex|both`. There, Codex wiring is created — deliberately. That is the point of the card, not a side effect.
- **Where the narrowed boundary is stated**: REQ-IHP-007 below, which is exactly the narrowed clause and is the criterion run-phase preserves.
- **Amendment record**: carried by this SPEC. SPEC-CODEX-WIRING-001 is `status: completed` and is not edited here; a reader arriving at REQ-CW-001 is directed to this SPEC through the `related_specs` frontmatter entry. If sync-phase policy requires an in-place amendment note on the owning SPEC, that is a sync-phase action on SPEC-CODEX-WIRING-001, not a plan-phase edit made here.

---

## §3 Requirements (GEARS)

- **REQ-IHP-001** (Ubiquitous) — The init wizard's interactive question set shall carry a harness-selection question offering the same closed set the `--agent` flag accepts (`claude`, `codex`, `both`), with `claude` pre-selected as the recommended default.

- **REQ-IHP-002** (Event-driven) — **When** the interactive wizard completes and the `--agent` flag was not explicitly set, the init command shall resolve the harness wiring from the wizard's selection.

- **REQ-IHP-003** (Event-driven) — **When** the `--agent` flag is explicitly set, the init command shall resolve the harness wiring from the flag value and discard the wizard's selection, matching the `autonomy_tier` flag-over-wizard precedence landed by SPEC-INIT-WIZARD-REPAIR-001 REQ-005.

- **REQ-IHP-004** (Ubiquitous) — The resolved harness wiring shall be produced at a single resolution point and shall reach **both** consumers: the Codex wiring call (`wireCodexUnlessClaude`) and the `.mcp.json` provisioning precedence switch. A resolution reaching only one consumer does not satisfy this requirement.

- **REQ-IHP-005** (State-driven) — **While** the run is non-interactive — `--non-interactive` given, or stdin is not a TTY — the injectable wizard seam `runWizardFn` shall be invoked **zero** times. The obligation is on the invocation count, not on whether an answer is produced.

  **Deviation from SPEC-CODEX-INIT-001 AC-CI-004's letter, recorded at the requirement level.** That criterion counts *prompts*; this requirement counts *seam invocations*, because no prompt-counting instrument exists in this tree and building one is out of scope (§5). The substitution carries the intent on a narrow, measured claim: `runWizardFn` is the sole issuer of **the harness question** (it renders only through `Page3Questions` → `InitQuestions` → `runWizardFn`), so zero invocations entails zero harness prompts. It is **not** the sole prompt issuer on the init path — a second `huh` confirm at `internal/cli/init.go:598-613` (profile setup) reads `isatty.IsTerminal` directly and runs before the `:644` wizard gate — so this requirement makes no claim about prompt totals.

- **REQ-IHP-006** (State-driven, unwanted) — **While** the `--agent` flag is absent, the init command shall not derive a non-`claude` harness from anything other than an interactive wizard answer.

- **REQ-IHP-007** (State-driven, unwanted) — **While** the `--agent` flag is absent **and** the run is non-interactive, the init command shall not create any of the **three** artifacts `codexwiring.Wire` writes — `.codex/hooks.json`, `.codex/config.toml`, and the trust sidecar `.moai/state/codex-wiring.json` — and shall **not** emit the `.mcp.json` provisioning announcement (`internal/cli/init.go:216`) — the same absence as at HEAD `2c18091d1`, where the wizard is the only writer of `opts.MCPProvision` (`:286`, called only from `:704`) so a non-interactive run leaves it `false` and `mcpDeclined` `true`. This is the narrowed REQ-CW-001 clause of §2 C3.

  The neighbouring criterion SPEC-CODEX-WIRING-001 AC-CW-004 points the other way on this same path — it asserts the `.mcp.json` **entry is present** — and both hold without conflict: that one observes template-deployed file content, this one observes whether the ensure-entry call ran. See acceptance.md AC-IHP-006a for the full disambiguation.

  The three paths are the exported constants `HooksRelPath` / `ConfigRelPath` / `SidecarPath` (`internal/codexwiring/codexwiring.go:29, 31, 34`), the third written at `internal/codexwiring/wire.go:187`. The `.codex/` **directory** may exist regardless — the template tree ships `.codex/agents/**` — so absence is asserted per file, never on the directory.

- **REQ-IHP-008** (Where — capability gate) — **Where** the conversation locale is `ko`, `ja`, or `zh`, the wizard shall render the harness question's title, description, and every option label and description in that locale; **where** the locale is `en`, the question literal's own text is the rendered text (English is the source language and carries no translation table).

- **REQ-IHP-009** (Ubiquitous) — The harness selection, whatever its origin (flag or wizard), shall determine whether the MCP-entry provisioning call runs, by one rule: `codex` declines it, `both` forces it, and `claude` leaves the `mcp_provision` answer intact.

  **Observable.** "Declined" is NOT observable as an absent `.mcp.json`: that file is template-deployed, so it exists either way (§4 D2). The observable this requirement is verified against is the stdout line emitted only on the provisioning path — `Provisioned the moai MCP server entry in .mcp.json (default-on).` (`internal/cli/init.go:216`, reached only after the `if declined { return }` early exit at `:208-209`). Present ⇒ provisioned; absent ⇒ declined.

- **REQ-IHP-010** (Ubiquitous, documentation) — The comment governing the `.mcp.json` precedence switch shall state a rule that is true after the change. The current justification "the flag beats the wizard answer" (`internal/cli/init.go:908-910`) is falsified once harness is itself a wizard axis (§4 D2) and shall not survive verbatim.

- **REQ-IHP-011** (Ubiquitous) — The `--agent` flag's closed set and its fail-loud validation in `validateInitFlags` shall remain unchanged; the wizard adds an input source, not a new accepted value.

---

## §4 Decisions

### D1 — Where the harness selection is resolved (the resolution seam)

**Chosen: a pure precedence helper in a new `internal/cli/init_agent_wizard.go`, mirroring the shape of `applyAutonomyTierFromWizard`, whose result is held in a `runInit`-local variable read by both consumers.**

```
resolveAgentWiringWithWizard(flagChanged bool, flagValue string, res *wizard.WizardResult) agentWiring
```

- It calls the existing flag-reading logic for the flag branch, so the flag semantics stay defined in one place and the existing `resolveAgentWiring` unit tests (`internal/cli/init_agent_flag_test.go:120-145`) keep their subject.
- `runInit` computes the value **once**, immediately after the wizard block (adjacent to the existing `applyAutonomyTierFromWizard` call at `internal/cli/init.go:717`), into a local. Both consumers are in the same function's tail — the `.mcp.json` switch at `:911-916` and `wireCodexUnlessClaude` at `:925` — and `runInit` spans `:454` to the end of the file, so a local declared at `:717` is in scope at both (`grep -n '^func ' internal/cli/init.go` shows no function boundary between `:454` and `:940`).
- `wireCodexUnlessClaude`'s signature changes from `(cmd *cobra.Command, projectRoot string)` to take the already-resolved wiring, so it can no longer re-read the flag behind the resolution point. This is what makes REQ-IHP-004 structural rather than a convention.

**Rejected — (a) an `InitOptions.AgentWiring` field.** `InitOptions` is the persistence carrier: `applyAutonomyTierFromWizard` writes `opts.AutonomyTier` because a downstream writer persists it. The harness selection persists nothing — measured: every reference to `agentWiring` outside tests lives in `internal/cli/init.go` control flow — `grep -rn "agentWiring" internal --include="*.go" | grep -v "_test.go" | wc -l` → **15** this run (an earlier draft cited 14; the count was wrong, the claim it supports was and remains true — every one of the 15 is in `init.go`) — and `InitOptions` carries no harness field today (`sed -n '/^type InitOptions struct/,/^}/p' internal/core/project/initializer.go` field list contains no `AgentWiring`/`Harness`). Adding one would put a field in a cross-package struct that no writer reads.

  *Reversal condition, stated so run-phase does not re-litigate silently*: if run-phase finds that the harness selection must be persisted to project config, option (a) becomes correct — add the `InitOptions` field then, and record the persistence consumer that justifies it.

**Rejected — (b) keep `resolveAgentWiring(cmd)` and pass the wizard answer as a second argument.** It leaves two call sites each re-resolving, so a future edit can make them disagree; and it forces `wizardResult` down into `wireCodexUnlessClaude`, whose job is wiring, not precedence.

**`validateInitFlags` is unchanged.** It validates the flag's closed set (`internal/cli/init.go:393-401`); the wizard's options are a closed set by construction (they are `Option` literals), so no second validation path is introduced. REQ-IHP-011.

### D2 — (wizard harness = `codex`) × (wizard `mcp_provision` = yes)

**Chosen: the harness selection wins — `codex` declines `.mcp.json` provisioning even when the `mcp_provision` answer was yes.** Today's behaviour is preserved bit-for-bit; only the *justification* changes.

The current comment justifies the override with "the flag beats the wizard answer" (`internal/cli/init.go:908-910`). Once harness is also a wizard question that sentence describes a wizard answer overriding a wizard answer, and is false. The replacement rule is REQ-IHP-009: **the harness selection is the more specific declaration about the MCP surface, wherever it came from.**

**What "declines provisioning" does and does not mean.** `.mcp.json` is a **template-deployed** file: it exists in the embedded template tree (`internal/template/templates/.mcp.json`, 398 bytes, measured this run), the tree is embedded wholesale (`//go:embed all:templates`, `internal/template/embed.go:28`), and the template tests call it "the distributed `.mcp.json`" / "the template-managed `.mcp.json` surface" (`internal/template/mcp_template_neutrality_test.go:3-5`) while asserting it carries exactly the `moai` and `context7` entries. `mcpDeclined = true` therefore does **not** make the file absent — it only skips the ensure-entry call, whose sole user-visible effect is the announcement at `internal/cli/init.go:216`. Every requirement and criterion in this SPEC is written against that announcement, never against the file's existence.

Two measured facts carry the decision:

1. Under `codex`, the moai MCP server is already registered for the user through `.codex/config.toml`'s `mcp_servers.moai` entry written by `codexwiring.Wire` (SPEC-CODEX-WIRING-001 REQ-CW-002/004). `.mcp.json` is the *Claude-side* registration; declining it under a Codex-only harness withdraws nothing — and, given the paragraph above, withdraws even less than the earlier draft argued, since the template ships the entry regardless.
2. The alternative rule — "an explicit `mcp_provision: yes` beats the harness implication" — is **not implementable in scope**. `WizardResult.MCPProvision` is a plain `bool` (`internal/cli/wizard/types.go`), not a pointer, so an explicitly-chosen yes and an accepted default-yes are indistinguishable. Its Page-3 neighbours `TodoEnabled` and `FeedbackAutoSubmit` are `*bool` precisely to preserve that distinction; `MCPProvision` is not. Implementing the alternative would require changing `MCPProvision` to `*bool` and auditing every reader — out of scope (§5), and a behaviour change to the flag path as well.

Consequently the comment correction (REQ-IHP-010) is in scope and is not cosmetic: it is the only place the now-false justification is recorded.

### D3 — Where the question sits

The harness question joins **`Page3Questions`** ("Quality & Workflow"), adjacent to `mcp_provision`, so the two answers that jointly decide the MCP surface are read together. It does **not** join `DefaultQuestions`, because `ReconfigureQuestions` is built from `DefaultQuestions` and would then inherit it (see §5 on the reconfigure path).

### D4 — Locale coverage mechanism

`TestWizardQuestionTranslationCompleteness` (`internal/cli/wizard/translations_completeness_test.go:95-135`) iterates `InitQuestions(root)` and fails on any question ID lacking a `ko`/`ja`/`zh` entry, and on any Select question whose option-translation count does not match. A question added to `Page3Questions` is therefore covered **automatically**, with no registration step — provided it lands in `InitQuestions`, which `Page3Questions` does by construction.

`en` is not covered by that guard (English is the source language and has no translation table); it is covered by the question literal's own non-empty `Title`/`Description`/`Option` text, which run-phase asserts directly (AC-IHP-006b).

---

## §5 Exclusions

### Out of Scope — the `--agent` flag contract

- Changing the closed set `{claude, codex, both}` or adding a fourth harness value.
- Changing `validateInitFlags`' fail-loud rejection semantics or its error text.
- Changing the flag's `claude` default or its help string beyond what naming the new wizard question requires.

### Out of Scope — what Codex wiring writes

- Any change to `codexwiring.Wire`'s outputs (`.codex/hooks.json`, `.codex/config.toml`, the trust sidecar) or to its best-effort failure handling.
- Any change to `provisionMCPEntryUnlessDeclined`'s `.mcp.json` content.

### Out of Scope — persistence of the harness selection

- Writing the harness selection into `.moai/config/**`. Nothing persists it today (§4 D1 evidence), and adding persistence is a separate decision with its own template surface.
- Converting `WizardResult.MCPProvision` from `bool` to `*bool` (§4 D2).

### Out of Scope — the reconfigure path

- `moai update --reconfigure` (`ReconfigureQuestions`, `internal/cli/wizard/questions.go:268-287`) is deliberately built from `DefaultQuestions` **without** `Page3Questions`, so that the page-3 set does not leak into it (documented at `questions.go:294-295`, guarded by AC-WIZ-012a). Placing the harness question in `Page3Questions` (§4 D3) keeps the reconfigure set unchanged by construction.
- **Finding, stated either way as the dispatch required**: no evidence was found that the reconfigure path must change. `moai update` has its own Codex-wiring path (`internal/cli/update_codex_wiring_test.go` exists), which this SPEC does not touch. Run-phase asserts the reconfigure set is unchanged (AC-IHP-008) rather than assuming it.

### Out of Scope — run-phase and delivery concerns

- Any code change: this SPEC is plan-phase only.
- Documentation of the new question in docs-site or README.

---

## §6 Non-functional constraints

- **No new dependency.** The change uses `huh` question types already in the wizard.
- **Non-interactive cost is zero `runWizardFn` invocations and zero behaviour delta** (REQ-IHP-005, REQ-IHP-007).
- **Locale parity is mechanically guarded**, not asserted by review (§4 D4).

---

## §7 Template-First determination

**This change touches no file under `.claude/` or `.moai/`; it is Go-only, so the `internal/template/templates/` mirroring + `make build` obligation does NOT apply.** Evidence at HEAD `2c18091d1`:

```
$ grep -rniE 'agent_wiring|agent_harness|harness_wiring' internal/template/templates/.moai/config/ | wc -l
0
$ grep -rn "moai init" internal/template/templates/ | grep -i agent
internal/template/templates/.claude/agents/moai/manager-design.md:95: ... NOT scaffolded by `moai init` ...
internal/template/templates/.claude/rules/moai/core/settings-management.md:36: ... provisioned default-on by `moai init` ...
internal/template/templates/.claude/rules/moai/development/model-policy.md:144: ... `moai init --model-policy <tier>` ...
internal/template/templates/.codex/agents/moai/manager-design.toml:89: ... NOT scaffolded by `moai init` ...
```

None of the four hits documents `moai init --agent` or enumerates the wizard's questions; the 16 template hits for the literal `--agent` are all Claude Code's own `claude --agent` flag, a different flag on a different binary. No template config section carries a harness key.

Run-phase obligation: re-run both greps before closing. If either becomes non-zero because the change adds a template-visible surface, the Template-First rule re-attaches and `make build` becomes mandatory.

---

## §8 Cross-references

- `SPEC-CODEX-WIRING-001` — REQ-CW-001 (flag-absent clause, **narrowed here** to `flag-absent ∧ wizard-not-run`; see §2 C3 for the sub-path it no longer covers), AC-CW-004 (asserts the `.mcp.json` entry is present on that path — compatible, see REQ-IHP-007), REQ-CW-002/004 (what `codexwiring.Wire` writes, unchanged).
- `SPEC-CODEX-INIT-001` — plan.md:51 / AC-CI-004: judge the non-interactive path by a **count of zero** rather than by whether an answer arrives. REQ-IHP-005 adopts that discipline while deviating from its letter (seam invocations, not prompts) — the deviation is stated at REQ-IHP-005 itself.
- `SPEC-INIT-WIZARD-REPAIR-001` — REQ-005 "flag-over-wizard precedence", the pattern D1 copies.
- `SPEC-MCP-DEFAULT-ON-001` — the `mcp_provision` default-on semantics D2 reasons over.
