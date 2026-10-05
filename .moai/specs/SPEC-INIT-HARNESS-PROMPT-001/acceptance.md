# SPEC-INIT-HARNESS-PROMPT-001 — Acceptance Criteria

All evidence below was produced by commands **actually executed** at worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t393`, branch `WT-init-harness-prompt`,
HEAD `2c18091d127cbc723074124e1015353e077300ca`, tree `a33625497c87246572f823cf427fe6335a79c825`.
Every "RED now" cell carries the command run, its observed output, and why that output is a failure.
Where a criterion **cannot be made to fail today**, the cell says so explicitly instead of implying a failure.

---

## §A Baseline (measured this run, at this tree)

| Command | Observed output |
|---|---|
| `go test ./internal/cli/wizard/ -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	3.204s` |
| `go test ./internal/cli/ -run 'TestAgent\|TestInitAgent\|Agent' -count=1 -timeout 600s` | `ok  	github.com/modu-ai/moai-adk/internal/cli	2.999s` |

Both packages are green before the change. Any RED observed during run-phase is attributable to the change.

### §A.1 The MCP observable — why these ACs do not assert file state

`.mcp.json` is **template-deployed** (`internal/template/templates/.mcp.json`, 398 bytes; embedded via `//go:embed all:templates`, `internal/template/embed.go:28`; called "the distributed `.mcp.json`" by `internal/template/mcp_template_neutrality_test.go:3-5`), so it exists whether or not provisioning runs. `mcpDeclined = true` only skips the ensure-entry call:

```
$ sed -n '207,217p' internal/cli/init.go
func provisionMCPEntryUnlessDeclined(out, errOut io.Writer, projectRoot string, declined bool) {
	if declined {
		return
	}
	...
	_, _ = fmt.Fprintln(out, "Provisioned the moai MCP server entry in .mcp.json (default-on).")
}
```

**The discriminator used by every MCP-side AC below is therefore the stdout line
`Provisioned the moai MCP server entry in .mcp.json (default-on).` — present ⇒ provisioned, absent ⇒ declined.**

**Run-phase obligation (blocking, recorded here so it is not discovered late and answered by weakening an AC).** The existing helper builds an output buffer and then throws it away: `internal/cli/init_autonomy_wiring_test.go:48` declares `var out, errBuf bytes.Buffer`, binds them with `cmd.SetOut/SetErr`, and `:55` returns `projectDir` alone. Run-phase MUST first extend that helper (or add a sibling) to return the captured stdout, **before** writing AC-IHP-003/004/006a/009. Weakening any of those four ACs because the buffer was not available is the named failure this paragraph exists to prevent.

---

## §B Acceptance criteria

### AC-IHP-001 — the harness question exists in the interactive init set

**Given** a fresh project root
**When** `wizard.InitQuestions(root)` is built
**Then** it contains exactly one question whose ID is the harness axis, of `QuestionTypeSelect`, with option values `{claude, codex, both}` and `Default == "claude"`.

- **RED now** — command run: `grep -rn 'ID: *"agent\|ID: *"harness\|AgentWiring\|agent_wiring' internal/cli/wizard/ | wc -l` → observed output `0`. Zero matches across the whole wizard package means no harness question ID exists, so a test asserting `QuestionByID(InitQuestions(root), <harness-id>) != nil` fails on a nil return. This is a failure and not merely an absence-of-test: the production capability the AC names is measurably not present.
- **Green path** — add the question literal to `Page3Questions` (`internal/cli/wizard/questions.go`, adjacent to `mcp_provision` at `:463`); the same grep then returns ≥ 1 and the `QuestionByID` assertion returns non-nil.

### AC-IHP-002 — the wizard answer is captured into `WizardResult`

**Given** the harness question is answered `codex`
**When** the wizard's select-answer handler runs
**Then** the selection is stored on `WizardResult` and is readable after the wizard returns.

- **RED now** — command run: `grep -cin 'harness\|agentwiring' internal/cli/wizard/types.go` → observed output `0`. `WizardResult` carries no harness field, so there is nowhere for `saveAnswer` (`internal/cli/wizard/wizard.go:397-448`) to store the value; a test constructing a `WizardResult` and reading the field does not compile. A compile failure is the strongest available RED.
- **Green path** — add the field to `WizardResult` (`internal/cli/wizard/types.go`) and a capture branch to `saveAnswer`'s switch; the test then compiles and the round-trip assertion passes.

### AC-IHP-003 — the answer reaches BOTH consumers from one resolution point

**Given** `--agent` is not set and the wizard answered `codex`
**When** `runInit` reaches its tail
**Then** (a) all three Codex artifacts exist — `.codex/hooks.json`, `.codex/config.toml`, `.moai/state/codex-wiring.json` (the Codex-wiring consumer saw `codex`), **and** (b) the captured stdout does **not** contain `Provisioned the moai MCP server entry in .mcp.json (default-on).` (the MCP-precedence consumer saw `codex`) — asserted **with the wizard's `mcp_provision` answer set to yes**, so leg (b) cannot pass by way of the decline default. Both legs are in one test; a test asserting only (a) or only (b) does not satisfy this AC, because a half-wired implementation passes it.

- **RED now** — two commands run.
  1. `grep -rn 'applyAgentWiringFromWizard\|resolveAgentWiringWithWizard' internal/` → observed output: *no lines*; hit count `0`. No precedence helper exists.
  2. The resolution function's body, read at `internal/cli/init.go:151-158`:
     ```go
     func resolveAgentWiring(cmd *cobra.Command) agentWiring {
     	switch agentWiring(getStringFlag(cmd, "agent")) {
     	case agentWiringCodex, agentWiringBoth:
     		return agentWiring(getStringFlag(cmd, "agent"))
     	default:
     		return agentWiringClaude
     	}
     }
     ```
     Its only input is the cobra flag. Both consumers call it — `internal/cli/init.go:167` (inside `wireCodexUnlessClaude`) and `internal/cli/init.go:912` — so with `--agent` absent both resolve `claude` **regardless of any wizard answer**.
  3. The MCP consumer, re-read this run after the leg-(b) predicate was rewritten:
     ```
     $ sed -n '911,917p' internal/cli/init.go
     	mcpDeclined := !opts.MCPProvision
     	switch resolveAgentWiring(cmd) {
     	case agentWiringCodex:
     		mcpDeclined = true
     	case agentWiringBoth:
     		mcpDeclined = false
     	}
     ```
     With `--agent` absent the switch scrutinee is `claude`, so neither branch is taken and `mcpDeclined` stays `!opts.MCPProvision` = `false` for a wizard `mcp_provision: yes`. Provisioning therefore runs and the announcement at `:216` **is** emitted — while leg (b) requires it absent.

  Consequently a test driving init with a wizard answer of `codex` fails on leg (a) with three missing files and on leg (b) with a present announcement. Mechanical failure on both legs, not a projection.
- **Green path** — introduce `resolveAgentWiringWithWizard(flagChanged, flagValue, wizardResult)` (new `internal/cli/init_agent_wizard.go`), compute it once in `runInit` next to the `applyAutonomyTierFromWizard` call at `internal/cli/init.go:717`, change `wireCodexUnlessClaude` to accept the resolved wiring, and make the `:911-916` switch read the same local. The same test then finds both files in their expected states.

### AC-IHP-004 — flag beats wizard, on each non-default value

**Given** `--agent claude` **and** a wizard answer of `codex`
**When** init runs
**Then** none of the three Codex artifacts is written, and the provisioning announcement follows the `mcp_provision` answer.
**And given** `--agent both` **and** a wizard answer of `codex` **and** a wizard `mcp_provision` answer of **no**
**Then** the three Codex artifacts are written **and** the announcement **is** emitted — the `both` rule (force provisioning on) beating both the `codex` answer and the explicit decline. The `mcp_provision: no` setting is what makes this row non-vacuous: with a yes the announcement would be emitted anyway and the row would assert nothing.

- **RED now** — same evidence as AC-IHP-003 item 1: `grep -rn 'applyAgentWiringFromWizard\|resolveAgentWiringWithWizard' internal/` → no lines, count `0`. No precedence helper exists, so there is no function under test and no place where `cmd.Flags().Changed("agent")` is consulted (`grep -n 'Changed("agent")' internal/cli/init.go` → no lines). A precedence table test does not compile.
- **Honest scope note** — the *observable outcome* of the `--agent claude` row is already produced today, because the wizard answer is discarded unconditionally. What is RED today is the **rule**: nothing distinguishes "flag won" from "wizard was never read". The run-phase test must therefore assert the precedence helper's return directly (flag value in, flag value out, wizard value ignored) in addition to the end-to-end file assertions — a file-only test for the `claude` row would pass vacuously both before and after.
- **Green path** — the helper returns the flag value when `flagChanged && flagValue != ""`, else the wizard value, else `claude`; the table test asserts all six cells (flag ∈ {absent, claude, codex, both} × wizard ∈ {codex, both}).

### AC-IHP-005 — non-interactive never invokes the wizard seam

**Given** `--non-interactive`, and separately, an absent TTY
**When** init runs
**Then** the injectable wizard seam `runWizardFn` is invoked exactly **0** times. The test substitutes a counting stub for the seam and asserts the counter is zero.

- **Instrument, named and verified this run.** `runWizardFn` is a package-level `var` — declared `var runWizardFn = func(rootFlag, locale, userName string) (*wizard.WizardResult, error)` at `internal/cli/init_update_notice.go:69`, invoked once at `internal/cli/init.go:654` inside the `if !nonInteractive && isInteractiveStdin()` block that opens at `:644`. It is already swapped by existing tests (`internal/cli/init_autonomy_wiring_test.go:36-38`, `internal/cli/init_gitdetect_test.go:200-208`), so the seam and its swap idiom both exist; nothing new is instrumented.
- **Deliberate deviation from SPEC-CODEX-INIT-001 AC-CI-004's LETTER, stated rather than glossed.** That criterion counts *prompts*; this one counts *seam invocations*. The substitution is sound on a narrower claim than an earlier draft of this cell made: **`runWizardFn` is the sole issuer of the harness question** — not the sole issuer of prompts on the init path. The harness question renders only through `Page3Questions` → `InitQuestions` (`internal/cli/wizard/questions.go:296-303`) → `runWizardFn`, and there is no other path to it, so zero `runWizardFn` invocations entails zero harness prompts. That is the whole of what this AC needs.
- **Residual gap, with its measured counterexample named.** It does not count prompts, and `runInit` demonstrably issues prompts outside this seam: a **second, independent `huh` prompt** — profile setup — lives at `internal/cli/init.go:598-613`, and runs BEFORE the `:644` wizard gate (31 lines earlier, measured from the block's closing brace at `:613`):

  ```
  $ sed -n '598,599p' internal/cli/init.go
  	if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) && !profile.IsSetup(profileName) {
  		var wantSetup bool
  ```

  It builds its own `huh.NewConfirm()`, runs its own form, and calls `runProfileSetup` at `:609`. It reads `isatty.IsTerminal` **directly** — the only such call in the file (`grep -n 'isatty\.' internal/cli/init.go` → `598:` alone) — so it bypasses the injectable `isInteractiveStdin`/`runWizardFn` seams entirely, and it runs **before** the `:644` wizard gate. This AC therefore says nothing about prompt totals, and is question-agnostic: it asserts the wizard did not run, not that a given question did not prompt. It would not catch a harness prompt issued from outside `runWizardFn` — the profile-setup prompt is the standing proof that such a site can exist in this very function. Building a prompt-counting instrument is **out of scope** (spec.md §5); accepting this weaker-but-real observable, on the narrower claim above, is the deliberate trade.
- **Cannot be made to fail today, and this is stated rather than implied.** With no harness question, and with the seam already gated, the assertion passes vacuously; no command at this tree yields a RED for it.
- **Run-phase RED obligation (binding, and now attributable).** Mutate the gate at `internal/cli/init.go:644` so the wizard branch is entered under `--non-interactive`, observe the counter assertion **FAIL**, record the command and output, restore. Attribution is exact here in a way it was not under the earlier prompt-count phrasing: the assertion's subject *is* the seam invocation, so a gate mutation moves precisely the quantity being asserted — it is not diluted across the other 15 questions, because the assertion never mentions questions.

### AC-IHP-006a — flag-absent + non-interactive preserves REQ-CW-001 behaviour

**Given** no `--agent` flag and `--non-interactive`
**When** init completes
**Then** none of the three artifacts `codexwiring.Wire` writes exists — `.codex/hooks.json`, `.codex/config.toml`, `.moai/state/codex-wiring.json` — **and the captured stdout does NOT contain** `Provisioned the moai MCP server entry in .mcp.json (default-on).` Absence is asserted per file, never on the `.codex/` directory, which the template populates with `.codex/agents/**` regardless.

- **Direction of the announcement, pinned in the assertion and not left to prose.** On this path the announcement is **ABSENT**, and the assertion is `NotContains`. Reason, established by reading the assignment chain this run: `opts.MCPProvision` has exactly one writer, `opts.MCPProvision = result.MCPProvision` (`internal/cli/init.go:286`) inside `applyWizardPage3ToOpts`, which is called from exactly one site, `internal/cli/init.go:704`, inside the interactive block. There is no `--mcp-provision` flag (`grep -n 'mcp-provision' internal/cli/init.go` → no lines). So on a non-interactive run the field keeps its zero value `false`, `mcpDeclined := !opts.MCPProvision` (`:911`) is `true`, `provisionMCPEntryUnlessDeclined` returns at `:208-209`, and `:216` is never reached. An implementer who asserts PRESENCE here gets a RED that has nothing to do with their change — which is why the direction is in the assertion.
- **Disambiguation against the neighbour that points the other way.** SPEC-CODEX-WIRING-001 AC-CW-004 asserts, on this same path, that the `.mcp.json` **entry is present**. Both are true and they do not conflict, because they observe different things: AC-CW-004 observes the **file's content**, which the template deploys (`internal/template/templates/.mcp.json`, §A.1), while this AC observes **whether the ensure-entry call ran**, which the announcement reports. Template-deployed content present + provisioning call skipped is the actual state of a non-interactive flag-absent init. **Do not** reconcile the two by weakening either: a run-phase engineer who reads AC-CW-004 and "corrects" this AC to assert presence has swapped the observable back to the one iteration 1 was failed for.

- **Relationship to the existing guard, stated so this AC is not silently weaker than what it claims to preserve.** `TestRunInit_AgentAbsentLeavesNoCodexFiles` (`internal/cli/init_agent_flag_test.go:97-108`) already asserts all three paths on the **flag-absent, wizard-stubbed-absent** path, and its comment already records the `.codex/agents/**` caveat. This AC **extends that guard to the wizard path**: same three paths, same per-file discipline, now with the wizard seam active and answering something other than the harness axis. It adds the announcement leg, which the existing guard does not assert. It must not be written with fewer than three paths — that would be a preservation criterion weaker than the guard it preserves, which is worse than no criterion.
  - Measured asymmetry worth carrying into run-phase: the sibling `TestRunInit_AgentClaudeLeavesNoCodexFiles` (`internal/cli/init_agent_flag_test.go:109-117`) asserts only **two** paths — it omits the sidecar. Run-phase should assert three on both rows rather than copy the two-path sibling.
- **Cannot be made to fail today.** It is a preservation criterion: the behaviour it names is the current behaviour and is already guarded; the package is green (`go test ./internal/cli/ -run 'TestAgent|TestInitAgent|Agent' -count=1` → `ok ... 2.999s`, §A). No command at this tree yields a RED.
- **Run-phase RED obligation** — after the change, prove the guard still binds by mutating the resolution helper's fallback from `claude` to `codex` and observing this assertion FAIL on all three paths; record the command and output; restore.

### AC-IHP-006b — English source text is present on the question literal

**Given** the harness question literal
**When** it is read from `InitQuestions`
**Then** `Title`, `Description`, and every option's `Label` and `Desc` are non-empty.

- **RED now** — same as AC-IHP-001: the question does not exist (`grep ... | wc -l` → `0`), so the assertion fails on a nil question.
- **Green path** — the question literal carries English text; English has no translation table (it is the source language), so this literal *is* the `en` rendering.

### AC-IHP-007 — ko / ja / zh translations exist and match option arity

**Given** locales `ko`, `ja`, `zh`
**When** `TestWizardQuestionTranslationCompleteness` iterates `InitQuestions(root)`
**Then** the harness question has a non-empty title and description translation in each, and its option-translation count equals its option count.

- **Cannot be made to fail today — measured, not assumed.** Command run: `go test ./internal/cli/wizard/ -run 'TestWizardQuestionTranslationCompleteness' -count=1` → observed output `ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	0.377s`. The guard passes at this tree because it iterates `InitQuestions`, which contains no harness question yet. It therefore cannot serve as RED evidence today.
- **Mechanically determined coverage (the dispatch's question, answered)** — the guard requires **no registration step** for a new ID. Read at `internal/cli/wizard/translations_completeness_test.go:95-135`: it loops over `InitQuestions(root)`, and for each question errors with `question %q has NO translation entry` when `langTrans[q.ID]` is absent, plus arity and non-empty checks for Select options. A question added to `Page3Questions` reaches `InitQuestions` by construction (`internal/cli/wizard/questions.go:296-303`), so it is covered automatically. The only exemption list is `optionTranslationExemptIDs`, which contains `conversation_language` alone; the harness question must NOT be added to it.
- **Run-phase RED obligation (binding, and this one is cheap)** — add the question **before** the translations, run `go test ./internal/cli/wizard/ -run TestWizardQuestionTranslationCompleteness -count=1`, and record the FAIL output naming the new ID for all three locales. Only then add the three translation entries (`internal/cli/wizard/translations.go`, ko at `:32`, ja at `:183`, zh at `:333`) and re-run to green. Adding translations first would leave this AC verified only by a test that never failed.

### AC-IHP-008 — the reconfigure question set is unchanged

**Given** `wizard.ReconfigureQuestions(root)`
**When** it is built after the change
**Then** it contains no harness question, and its ID sequence is unchanged from HEAD `2c18091d1`.

- **Cannot be made to fail today** (the question does not exist). This is a leak guard for §4 D3.
- **Run-phase RED obligation** — prove non-vacuity by temporarily placing the harness question in `DefaultQuestions` instead of `Page3Questions` and observing this assertion FAIL (that is exactly the leak path, since `ReconfigureQuestions` splices `DefaultQuestions`, `internal/cli/wizard/questions.go:268-287`); record and restore.

### AC-IHP-009 — the MCP-precedence rule holds for the wizard × wizard combination

**Given** no `--agent` flag, a wizard harness answer of `codex`, and a wizard `mcp_provision` answer of yes
**When** init completes
**Then** the captured stdout does **not** contain `Provisioned the moai MCP server entry in .mcp.json (default-on).` (§4 D2: the harness selection wins over the `mcp_provision` answer), and `.codex/config.toml` carries the `mcp_servers.moai` entry. The AC does **not** assert `.mcp.json` absent — that file is template-deployed and present either way (§A.1).

- **RED now** — this combination is unreachable at this tree. With `--agent` absent, `resolveAgentWiring` returns `claude` (body quoted under AC-IHP-003), so in
  ```
  $ sed -n '911,917p' internal/cli/init.go
  	mcpDeclined := !opts.MCPProvision
  	switch resolveAgentWiring(cmd) {
  	case agentWiringCodex:
  		mcpDeclined = true
  	case agentWiringBoth:
  		mcpDeclined = false
  	}
  ```
  neither branch is taken and `mcpDeclined` stays `!opts.MCPProvision` = `false` for a wizard `mcp_provision: yes`. Provisioning runs, so the announcement at `internal/cli/init.go:216` **is** emitted and the assertion fails on a present line. Mechanical, and unaffected by the file-state correction that invalidated the earlier phrasing.
- **Green path** — the resolved wiring flows into the switch, whose `agentWiringCodex` branch sets `mcpDeclined = true`.

### AC-IHP-010 — the falsified justification is gone and the true rule is stated

**Given** `internal/cli/init.go` after the change
**When** the MCP-precedence comment is read
**Then** it does not contain the string `the flag beats the wizard answer`, and it states the D2 rule (the harness selection — from flag or wizard — decides provisioning).

- **RED now** — command run: `sed -n '895,940p' internal/cli/init.go`; observed output contains, at `:908-910`:
  ```
  // SPEC-CODEX-WIRING-001 D3 stacks on top: --agent codex treats the
  // provisioning as declined (the user declared their harness is Codex —
  // the flag beats the wizard answer), --agent both forces it on.
  ```
  A `grep -q 'the flag beats the wizard answer' internal/cli/init.go` succeeds today (exit 0), which is the failing condition for the post-change assertion `! grep -q ...`. The sentence is not merely stale — it is falsified by AC-IHP-009's combination, where a wizard answer overrides a wizard answer and no flag is present.
- **Green path** — rewrite the comment to state REQ-IHP-009's rule; the `!grep` assertion then passes. The doc assertion is a companion to AC-IHP-009, never a substitute: the behavioural test is the binding gate.

### AC-IHP-011 — the `--agent` flag contract is untouched

**Given** the post-change binary
**When** `--agent gemini` is passed
**Then** init fails with the existing message `invalid --agent value "gemini": must be one of: claude, codex, both`, and the existing flag tests still pass.

- **Cannot be made to fail today** — it is a no-regression criterion, and `internal/cli/init_agent_flag_test.go:53` already covers the rejection; §A shows the package green.
- **Green path** — leave `validateInitFlags` (`internal/cli/init.go:393-401`) unmodified; re-run `go test ./internal/cli/ -run 'TestAgent|TestInitAgent|Agent' -count=1` and record it green post-change.

### AC-IHP-012 — Template-First determination re-verified at close

**Given** the completed change
**When** the two §7 greps are re-run
**Then** both still report zero template-visible harness surface, or the Template-First obligation (`internal/template/templates/` mirror + `make build`) is discharged and evidenced.

- **RED now** — not applicable; this is a close-time gate. Baseline observed this run:
  `grep -rniE 'agent_wiring|agent_harness|harness_wiring' internal/template/templates/.moai/config/ | wc -l` → `0`.
- **Green path** — re-run at close and cite the output.

---

## §C Coverage matrix (REQ → AC)

| REQ | AC | RED available today? |
|---|---|---|
| REQ-IHP-001 | AC-IHP-001, AC-IHP-006b | yes (grep → 0) |
| REQ-IHP-002 | AC-IHP-002, AC-IHP-003 | yes (grep → 0 / compile failure / flag-only body) |
| REQ-IHP-003 | AC-IHP-004 | yes for the rule; the `claude` row's outcome is vacuous today (stated). The `both` row is non-vacuous only because it pins `mcp_provision: no` |
| REQ-IHP-004 | AC-IHP-003 | yes — both legs fail today: (a) three missing files, (b) announcement present |
| REQ-IHP-005 | AC-IHP-005 | no — `runWizardFn` invocation-count assertion; gate-mutation obligation recorded |
| REQ-IHP-006 | AC-IHP-004, AC-IHP-009 | yes (AC-IHP-009 combination unreachable today) |
| REQ-IHP-007 | AC-IHP-006a | no — preservation of all three artifacts; extends `init_agent_flag_test.go:97-108`; mutation obligation recorded |
| REQ-IHP-008 | AC-IHP-006b (en), AC-IHP-007 (ko/ja/zh) | en yes; ko/ja/zh no today — ordering obligation recorded |
| REQ-IHP-009 | AC-IHP-009 | yes |
| REQ-IHP-010 | AC-IHP-010 | yes (`grep -q` succeeds today) |
| REQ-IHP-011 | AC-IHP-011 | no — no-regression |
| — (D3 leak guard) | AC-IHP-008 | no — mutation obligation recorded |
| — (§7) | AC-IHP-012 | close-time gate |

**Counts, reconciled and executed** (`grep -c '^### AC-IHP-' acceptance.md` → `13`): there are **13** ACs — `001, 002, 003, 004, 005, 006a, 006b, 007, 008, 009, 010, 011, 012`. `006` is deliberately split into `006a` (behaviour preservation) and `006b` (English source text) because they verify different REQs; the split is why the highest numeral is `012` while the population is 13. Every REQ has at least one AC.

**Six** of the 13 cannot be made to fail at this tree — `005`, `006a`, `007`, `008`, `011`, `012` — each says so in its own cell and carries a binding run-phase mutation, ordering, or close-time obligation instead of prose implying a failure. The remaining seven carry an executed RED at this tree.

---

## §D Definition of Done

1. All **13** ACs pass, each with the command and output recorded in `progress.md` §E.2.
2. The four mutation / ordering obligations (AC-IHP-005, AC-IHP-006a, AC-IHP-007, AC-IHP-008) each carry a recorded observed FAIL and the restore.
2b. The §A.1 stdout-capture obligation is discharged **before** AC-IHP-003/004/006a/009 are written, and the extended helper is named in `progress.md` §E.2.
3. `go test ./internal/cli/wizard/ ./internal/cli/ -count=1` green, re-measured on the tree that will be committed.
4. `go vet ./internal/cli/... ` clean; `golangci-lint run` shows no new finding attributable to the change (per-rule delta, not a total).
5. AC-IHP-012 re-verified and its output cited.
6. CI green on a quiet head; the local run is an early signal, not the verdict.
