# SPEC-INIT-HARNESS-PROMPT-001 — Progress

Card: t393 · Worktree: `.claude/worktrees/t393` · Branch: `WT-init-harness-prompt`
Plan-phase anchor: HEAD `2c18091d127cbc723074124e1015353e077300ca`, tree `a33625497c87246572f823cf427fe6335a79c825`

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` — Tier M set, `status: draft`.
- **Tier**: M. Two packages (`internal/cli`, `internal/cli/wizard`), a resolution-seam change with two
  consumers, a cross-axis precedence correction, a new question with 3-locale translations, and a
  documentation correction. Larger than a single-file Tier S change; no new subsystem or external
  contract, so not Tier L.
- **SPEC ID check** — command run:
  `ID="SPEC-INIT-HARNESS-PROMPT-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
  → observed output: `PASS`.
- **ID uniqueness** — `ls .moai/specs/ | grep -i -E 'INIT|HARNESS|PROMPT'` at this HEAD lists 47 neighbouring
  directories; `SPEC-INIT-HARNESS-PROMPT-001` is not among them.
- **Baseline measured this run**: `go test ./internal/cli/wizard/ -count=1` → `ok ... 3.204s`;
  `go test ./internal/cli/ -run 'TestAgent|TestInitAgent|Agent' -count=1 -timeout 600s` → `ok ... 2.999s`.
- **Template-First**: determined NOT to apply (spec.md §7, two greps recorded there); re-verified at close
  by AC-IHP-012.
- **Open markers: ZERO.** The single clarification (whether `mcp_provision` should be conditionally
  hidden under a `codex` harness answer) was settled by the operator on 2026-09-01 — **ask it
  unconditionally, no `Condition` func** — and is recorded as `plan.md` §B Decision B1 with its accepted
  cost and reversal condition. Verified: the clarification-marker grep over this directory returns no matches.
- **Phases skipped**: Phase 2 (exploration), Phase 4 (deep interview), Phase 6 (deep research) were
  SKIPPED. Rationale: the dispatch fixed the facts at file:line granularity and the orchestrator
  re-verified them at this HEAD, so a research sweep would re-measure what is already attributed. All
  cited file:line coordinates were nonetheless re-read in this worktree rather than carried over, and two
  dispatch premises were corrected as a result (spec.md §2 C1, C2). **Iteration-2 correction**: the skip
  was too broad in one place — `internal/codexwiring/` was never opened in iteration 1, which is what
  made the two-artifact error (D4) possible. It was read in iteration 2; the skip rationale holds for
  the paths the dispatch actually named, not for paths it merely referred to.

### Iteration 2 — audit-revision record (2026-09-01)

plan-auditor iteration 1 returned FAIL 0.70 (Clarity 0.65 / Completeness 0.85 / Testability 0.55 /
Traceability 0.85), plus a score-independent MP-7 failure. Every figure below was **re-measured by the
author** at HEAD `2c18091d1` / tree `a33625497`; the auditor's findings were treated as leads.

| Defect | Change made | Author's own re-measurement |
|---|---|---|
| MP-7 clarification | Marker deleted; `plan.md` §B Decision B1 records "ask unconditionally", cost, reversal condition | the clarification-marker grep over this directory returns no matches |
| D1 MCP predicate | Four ACs + REQ-IHP-009 re-specified onto the stdout announcement; `acceptance.md` §A.1 added | `ls -l internal/template/templates/.mcp.json` → `398` bytes; `grep -n 'Provisioned the moai MCP server entry' internal/cli/init.go` → `216:` |
| D3 AC-IHP-005 | Re-stated against `runWizardFn` invocation count; deviation + residual gap stated; old mutation replaced | `grep -rn 'runWizardFn' internal/cli/*.go \| grep -v _test` → decl `init_update_notice.go:69`, call `init.go:654` |
| D4 third artifact | REQ-IHP-007 + AC-IHP-006a now name the sidecar; relationship to the landed guard stated | `grep -n 'HooksRelPath =\|ConfigRelPath =\|SidecarPath =' internal/codexwiring/codexwiring.go` → `29 / 31 / 34`; write at `wire.go:187`; guard `init_agent_flag_test.go:97` |
| D5 REQ-CW-001 | §2 C3 rewritten as a **narrowing** with the precise new condition | Clause read at `.moai/specs/SPEC-CODEX-WIRING-001/spec.md:229-231` — no `--non-interactive` qualifier present |
| D6 count | `14` → `15` | `grep -rn "agentWiring" internal --include="*.go" \| grep -v "_test.go" \| wc -l` → `15` |
| D7 coordinate | `wizard.go:475-490` → `:459-476`; `saveAnswer` `:410-448` → `:397-448`; guard `:93-131` → `:95-135`; `questions.go:296-303` → `:296-305` | `awk '/^func saveBoolAnswer/{s=NR} s&&/^}/{print s"-"NR; exit}'` → `459-476` |
| D8 AC count | `12` → `13`, with the `006a`/`006b` split explained | `grep -c '^### AC-IHP-' acceptance.md` → `13` |
| D9 flag default | §1 now states the cobra default is `""`, `claude` is the resolution fallback; `plan.md` M1 states `flagChanged && flagValue != ""` | `init.go:132` read verbatim |
| Gap 2 (16 questions) | Now an executed per-constructor count with its scope stated | `awk` counts → Default `5`, Page3 `11`, Git `7`; 5+11=16, +7=23 |

### Iteration 3 — audit-revision record (2026-09-01, final authoring round)

plan-auditor iteration 2: **PASS-WITH-DEBT 0.85** (Tier M threshold 0.80), all seven must-pass PASS, delta
0.70 → 0.85 monotonic. Five remaining defects were textual; all five are fixed below. Every coordinate was
re-opened at HEAD `2c18091d1` / tree `a33625497` by the author.

| Defect | Change | Author's own re-measurement |
|---|---|---|
| R1 — false justification | AC-IHP-005 + plan.md M4 + REQ-IHP-005 now claim only that `runWizardFn` is the sole issuer of the **harness question**; the profile-setup prompt is named as the measured counterexample | `sed -n '598,599p' internal/cli/init.go` → `if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) && !profile.IsSetup(profileName) {`; `grep -n 'isatty\.' internal/cli/init.go` → `598:` only (the sole direct isatty use, bypassing both seams, 46 lines before the `:644` wizard gate) |
| R2 — observable had no direction | AC-IHP-006a + REQ-IHP-007 pin the announcement **ABSENT** in the assertion, with the AC-CW-004 disambiguation | `grep -n 'MCPProvision' internal/cli/init.go` → sole writer `:286` (`opts.MCPProvision = result.MCPProvision`), called only from `:704` inside the interactive block; `grep -n 'mcp-provision' internal/cli/init.go` → no lines (no flag source). Non-interactive ⇒ field stays `false` ⇒ `mcpDeclined` true ⇒ `:216` unreached. **Established by reading the assignment chain, not by executing init** (see Gaps) |
| R3 — contradictory line | §8's REQ-CW-001 entry rewritten from "extended here" to "**narrowed here** to `flag-absent ∧ wizard-not-run`", with AC-CW-004 named as compatible | `grep -n 'extended here' spec.md` → no lines |
| R4 — REQ/AC mismatch | REQ-IHP-005 now asks for zero `runWizardFn` invocations; the AC-CI-004 deviation is stated at the requirement level; §6 and §8 wording aligned | requirement text re-read after edit |
| R5 — over-correction | `InitQuestions` coordinate restored `:296-305` → `:296-303` | `awk '/^func InitQuestions/{s=NR} s&&/^}/{print s"-"NR; exit}' internal/cli/wizard/questions.go` → `296-303`; `sed -n '296,303p'` shows the closing brace on 303. The v0.1.1 round changed a coordinate that was already right — a correction round introduces coordinate errors as readily as it removes them |
| **R5 recurrence (v0.1.3)** | The R1 text written in this same round carried two fresh coordinate errors: the profile-setup block cited `:598-612` (actual `:598-613`) and "twelve lines above the wizard gate" (actual 31 from `:613`, 46 from `:598`). Both corrected in v0.1.3 | `awk 'NR>=596 && NR<=616 {printf "%d:%s\n", NR, $0}' internal/cli/init.go` → braces close at `611` / `612` / `613`, `:614` blank, `:615` `prefs, err := profile.ReadPreferences(profileName)`; `grep -n 'if !nonInteractive && isInteractiveStdin()' internal/cli/init.go` → `644:`. **The lesson recorded one row above was re-instantiated in the round that wrote it** — not a new lesson, a recurrence: the correcting text is exactly as prone to coordinate drift as the text it corrects, and neither round's coordinates were re-opened after being written. The durable fix is to re-measure citations *after* drafting the sentence that carries them, not only before |
| Gap 6 (now measured) | plan.md M2.0 carries sibling-by-default guidance with the measurement | `grep -rn 'runInitForAutonomy' internal/cli --include="*_test.go"` → 16 matches / 4 files; minus 2 `func` declarations and 3 comment mentions = **11 invocation sites**; coupling compile-enforced |
| Gap 2 (provenance) | spec.md §2 C1 now states the non-looping premise was established **by reading**, not by counting | the `awk` counts remain (5 / 11 / 7); only the premise's provenance is corrected |

**Dispositions recorded, not acted on** (plan.md §G.1): the sibling-guard asymmetry in
`init_agent_flag_test.go:109-117` is out of scope and belongs to a separate card; the one-directional
REQ-CW-001 back-link is recorded as a sync-phase obligation.

### Gaps carried into run-phase (current as of iteration 3)

1. **Full-catalog `moai spec lint` not run** — >300s at 722 SPEC dirs; only the changed files were linted. Cross-SPEC dependency-DAG and zone-registry effects unverified.
2. **`golangci-lint` not run** — the baseline is `go test` on the two affected packages only.
3. **darwin only** — no cross-platform measurement of the init path.
4. **`moai update --reconfigure`'s own Codex path** (`internal/cli/update_codex_wiring_test.go`) not read; the out-of-scope finding rests on `ReconfigureQuestions`' composition, which was read.
5. **The `.mcp.json`-is-deployed and announcement-absent facts were established by reading, not by executing `moai init`.** A real init needs `HOME` pinned to a temp dir to stay out of the user's `~/.claude`, and the worktree guard refuses a `HOME`-setting invocation. This does not weaken the ACs: the announcement predicate observes the branch taken and holds whether or not the file exists. Run-phase confirms it inside the Go harness, where `t.Setenv` pins `HOME` legitimately.
6. ~~Helper blast radius unmeasured~~ — **CLOSED** at iteration 3: 11 invocation sites across 4 files, compile-enforced; sibling-by-default recorded in M2.0.

**Residual risk**: the seam change alters `wireCodexUnlessClaude`'s signature, and the `-run 'Agent'`
baseline does not cover all of `internal/cli`; run-phase runs the package unfiltered before claiming green.

## §E.2 Run-phase Evidence

Run-phase anchor: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t393`, branch
`WT-init-harness-prompt`, base HEAD `2c18091d1` (= `origin/develop` at run-phase entry). Card t393.
Every command below was executed in this worktree, against this tree, in this run.

### E.2.0 M2.0 route taken (the §A.1 blocking prerequisite), and why

**Route: SIBLING, with the existing helper delegating to it.**

`runInitForAutonomyAtHomeCapturingOut(t, homeDir, wizResult, flags) (projectDir, stdout string)` was
added to `internal/cli/init_autonomy_wiring_test.go`; the pre-existing
`runInitForAutonomyAtHome` now delegates to it and discards the second return.

The plan's M2.0 default was "sibling unless the sibling would duplicate meaningful setup". Delegation
satisfies **both** halves at once rather than trading one against the other: the 11 pre-existing
invocation sites across 4 files are untouched (the sibling property), and the ~30 lines of seam-swap +
flag + buffer setup exist in exactly one place (no duplication). A literal copy-paste sibling would
have satisfied the letter of the default while duplicating the setup the default's escape clause
exists to avoid.

This was done BEFORE AC-IHP-003 / 004 / 006a / 009 were written, per the §A.1 obligation. No MCP-side
AC was weakened to a file-existence assertion; every MCP-side assertion in this run observes the
stdout announcement.

### E.2.1 RED evidence (verbatim, captured before GREEN)

**RED-1a — AC-IHP-002 (compile failure).** `internal/cli/wizard/agent_wiring_question_test.go` written
first, against a `WizardResult` with no harness field:

```
$ go test ./internal/cli/wizard/ -count=1
# github.com/modu-ai/moai-adk/internal/cli/wizard [github.com/modu-ai/moai-adk/internal/cli/wizard.test]
internal/cli/wizard/agent_wiring_question_test.go:126:13: result.AgentWiring undefined (type *WizardResult has no field or method AgentWiring)
internal/cli/wizard/agent_wiring_question_test.go:127:105: result.AgentWiring undefined (type *WizardResult has no field or method AgentWiring)
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard [build failed]
FAIL
```

**RED-1b — AC-IHP-001 / 002 / 006b (behavioural).** After adding ONLY the `WizardResult.AgentWiring`
field (the minimum to compile), with no question and no capture branch:

```
$ go test ./internal/cli/wizard/ -run 'AgentWiring|ReconfigureQuestions_NoHarnessLeak|SaveAnswer_CapturesAgentWiring' -count=1
--- FAIL: TestSaveAnswer_CapturesAgentWiring (0.00s)
    agent_wiring_question_test.go:127: after saveAnswer(agent_wiring, "claude"): WizardResult.AgentWiring = "", want "claude"
    agent_wiring_question_test.go:127: after saveAnswer(agent_wiring, "codex"): WizardResult.AgentWiring = "", want "codex"
    agent_wiring_question_test.go:127: after saveAnswer(agent_wiring, "both"): WizardResult.AgentWiring = "", want "both"
--- FAIL: TestAgentWiringQuestion_InInitSetWithClosedOptionSet (0.00s)
    agent_wiring_question_test.go:33: InitQuestions contains 0 questions with ID "agent_wiring", want exactly 1 (AC-IHP-001)
--- FAIL: TestAgentWiringQuestion_PrecedesMCPProvision (0.00s)
    agent_wiring_question_test.go:79: agent_wiring is not in Page3Questions (spec.md §4 D3)
--- FAIL: TestAgentWiringQuestion_EnglishSourceTextPresent (0.00s)
    agent_wiring_question_test.go:99: agent_wiring question absent (AC-IHP-006b)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard	0.456s
FAIL
```

**RED-2a — AC-IHP-003 / 004 / 009 (behavioural, against the flag-only resolution).** The end-to-end
tests were written before the resolution seam existed. Three failed; the other three rows passed
today for exactly the reasons acceptance.md records (vacuous / preservation).

```
$ go test ./internal/cli/ -run 'TestRunInit_Wizard|TestRunInit_Flag|TestRunWizardFn_Zero' -count=1 -timeout 900s
--- FAIL: TestRunInit_WizardCodexReachesBothConsumers (0.33s)
    init_agent_wizard_test.go:74: .codex/hooks.json missing, want present: stat .../tier-proj/.codex/hooks.json: no such file or directory
    init_agent_wizard_test.go:74: .codex/config.toml missing, want present: stat .../tier-proj/.codex/config.toml: no such file or directory
    init_agent_wizard_test.go:74: .moai/state/codex-wiring.json missing, want present: stat .../tier-proj/.moai/state/codex-wiring.json: no such file or directory
    init_agent_wizard_test.go:78: stdout contains "Provisioned the moai MCP server entry in .mcp.json (default-on)." — the MCP precedence consumer did not see the wizard's codex answer (AC-IHP-003 leg b)
--- FAIL: TestRunInit_WizardCodexDeclinesMCPProvisioning (0.24s)
--- FAIL: TestRunInit_WizardBothForcesProvisioningOverDecline (0.31s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.966s
FAIL
```

Both legs of AC-IHP-003 failed, as acceptance.md predicted: leg (a) on three missing files, leg (b)
on a present announcement.

**RED-2b — AC-IHP-004 rule / AC-IHP-010 (compile failure).**

```
$ go test ./internal/cli/ -run 'TestResolveAgentWiringWithWizard|TestMCPPrecedenceComment' -count=1 -timeout 600s
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/init_agent_wizard_precedence_test.go:70:11: undefined: resolveAgentWiringWithWizard
internal/cli/init_agent_wizard_precedence_test.go:84:12: undefined: resolveAgentWiringWithWizard
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

### E.2.2 The four mutation / ordering obligations (each observed FAIL, then restored)

**AC-IHP-007 — ordering RED (M5).** The question was added BEFORE the translations, exactly so the
completeness guard could be seen failing rather than passing without ever having failed:

```
$ go test ./internal/cli/wizard/ -run TestWizardQuestionTranslationCompleteness -count=1
--- FAIL: TestWizardQuestionTranslationCompleteness (0.00s)
    translations_completeness_test.go:108: locale "ko": question "agent_wiring" has NO translation entry
    translations_completeness_test.go:108: locale "ja": question "agent_wiring" has NO translation entry
    translations_completeness_test.go:108: locale "zh": question "agent_wiring" has NO translation entry
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard	0.387s
FAIL
```

The three entries were then added to `internal/cli/wizard/translations.go` and the guard re-run green.
`agent_wiring` was NOT added to `optionTranslationExemptIDs`.

**AC-IHP-008 — leak-guard mutation.** The harness question was temporarily relocated from
`Page3Questions` into `DefaultQuestions` (the exact leak path, since `ReconfigureQuestions` splices
`DefaultQuestions`):

```
$ go test ./internal/cli/wizard/ -run TestReconfigureQuestions_NoHarnessLeak -count=1
--- FAIL: TestReconfigureQuestions_NoHarnessLeak (0.00s)
    agent_wiring_question_test.go:141: agent_wiring leaked into ReconfigureQuestions — page-3 questions must not reach `moai update --reconfigure` (AC-WIZ-012a, AC-IHP-008)
    agent_wiring_question_test.go:164: ReconfigureQuestions ID sequence = [conversation_language user_name project_name agent_wiring model_policy report_format git_mode git_provider gitlab_instance_url github_username github_token gitlab_username gitlab_token], want [conversation_language user_name project_name model_policy report_format git_mode git_provider gitlab_instance_url github_username github_token gitlab_username gitlab_token] (unchanged from HEAD 2c18091d1)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard	0.388s
FAIL
```

Both legs failed — the presence check and the pinned ID sequence. Restored; package re-run green
(`ok github.com/modu-ai/moai-adk/internal/cli/wizard 2.934s`).

**AC-IHP-005 — gate mutation.** The `!nonInteractive` conjunct was dropped from the wizard gate in
`internal/cli/init.go` (`if !nonInteractive && isInteractiveStdin()` → `if isInteractiveStdin()`), so
the wizard branch is entered under `--non-interactive`:

```
$ go test ./internal/cli/ -run TestRunWizardFn_ZeroInvocationsNonInteractive -count=1 -timeout 600s
--- FAIL: TestRunWizardFn_ZeroInvocationsNonInteractive (0.71s)
    --- FAIL: TestRunWizardFn_ZeroInvocationsNonInteractive/--non-interactive_with_a_TTY_present (0.45s)
        init_agent_wizard_test.go:265: runWizardFn was invoked 1 times, want 0 (AC-IHP-005)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.955s
FAIL
```

Attribution is exact, as acceptance.md notes: the assertion's subject IS the seam invocation, so the
gate mutation moves precisely the asserted quantity. Restored (`grep -n 'if !nonInteractive && isInteractiveStdin()' internal/cli/init.go` → `669:`); re-run green.

**AC-IHP-006a — fallback mutation.** The resolution helper's claude fallback was flipped to codex:

```
$ go test ./internal/cli/ -run TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence -count=1 -timeout 600s
--- FAIL: TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence (0.34s)
    init_agent_wizard_test.go:200: .codex/hooks.json present, want absent
    init_agent_wizard_test.go:200: .codex/config.toml present, want absent
    init_agent_wizard_test.go:200: .moai/state/codex-wiring.json present, want absent
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.330s
FAIL
```

FAIL on all THREE paths, which is what this AC requires (it must not be weaker than the landed guard
it extends). Restored; `grep -c MUTATION internal/cli/init_agent_wizard.go internal/cli/init.go` → `0, 0`.

Note, stated rather than glossed: the announcement leg of this AC did NOT move under this mutation,
because a codex resolution sets `mcpDeclined = true` and the announcement stays absent either way.
The mutation exercised the three-artifact leg only. The announcement leg's non-vacuity is carried by
AC-IHP-003 leg (b) and AC-IHP-009, where it was observed failing (RED-2a above).

### E.2.3 AC PASS/FAIL matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-IHP-001 | PASS | `go test ./internal/cli/wizard/ -run 'TestAgentWiringQuestion_InInitSetWithClosedOptionSet' -count=1` and `-run 'TestAgentWiringQuestion_PrecedesMCPProvision'` | `ok  	github.com/modu-ai/moai-adk/internal/cli/wizard` — one `agent_wiring` question in `InitQuestions`, Select type, values `[claude codex both]`, `Default == "claude"`, no `Condition`, immediately before `mcp_provision` in the `Quality & Workflow` group. RED at E.2.1 (`0` questions with that ID). |
| AC-IHP-002 | PASS | `go test ./internal/cli/wizard/ -run TestSaveAnswer_CapturesAgentWiring -count=1` | green; round-trip `saveAnswer("agent_wiring", v)` → `WizardResult.AgentWiring == v` for all three values. RED at E.2.1 (compile failure, then `""` for all three). |
| AC-IHP-003 | PASS | `go test ./internal/cli/ -run TestRunInit_WizardCodexReachesBothConsumers -count=1 -timeout 600s` | green. BOTH legs in ONE test: (a) all three Codex artifacts present, (b) stdout does NOT contain the announcement — with `mcp_provision: true` pinned, so leg (b) cannot pass via the decline default. RED at E.2.1 RED-2a: three missing files AND a present announcement. |
| AC-IHP-004 | PASS | `go test ./internal/cli/ -run 'TestResolveAgentWiringWithWizard_PrecedenceTable' -count=1` plus the two end-to-end flag rows | green. The RULE is asserted directly on the helper's return across 11 rows (flag ∈ {absent, set-empty, claude, codex, both, unrecognized} × wizard ∈ {claude, codex, both, silent}) — a file-only test for the `claude` row would pass vacuously. The `both` row pins `mcp_provision: no`, so the force-on is non-vacuous. RED at E.2.1 RED-2b (undefined helper). |
| AC-IHP-005 | PASS | `go test ./internal/cli/ -run TestRunWizardFn_ZeroInvocationsNonInteractive -count=1 -timeout 600s` | green — `runWizardFn` invoked 0 times under `--non-interactive` and under an absent TTY. Non-vacuity proven by the gate mutation at E.2.2 (`invoked 1 times, want 0`). No prompt-counting instrument was built (out of scope). |
| AC-IHP-006a | PASS | `go test ./internal/cli/ -run TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence -count=1 -timeout 600s` | green — all THREE Codex artifacts absent (not the two-path sibling form) AND the announcement absent (`NotContains`, the pinned direction). Non-vacuity proven by the fallback mutation at E.2.2 (FAIL on all three paths). The AC-CW-004 neighbour was NOT reconciled by weakening either side. |
| AC-IHP-006b | PASS | `go test ./internal/cli/wizard/ -run TestAgentWiringQuestion_EnglishSourceTextPresent -count=1` | green — `Title`, `Description`, and every option `Label` / `Desc` non-empty on the question literal. RED at E.2.1 RED-1b (`agent_wiring question absent`). |
| AC-IHP-007 | PASS | `go test ./internal/cli/wizard/ -run TestWizardQuestionTranslationCompleteness -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/cli/wizard`. Ordering obligation discharged: the FAIL naming `agent_wiring` for ko/ja/zh was observed FIRST (E.2.2), then the three entries were added. Not in `optionTranslationExemptIDs`. |
| AC-IHP-008 | PASS | `go test ./internal/cli/wizard/ -run TestReconfigureQuestions_NoHarnessLeak -count=1` | green — no `agent_wiring` in `ReconfigureQuestions`, and the 12-ID sequence is byte-identical to the sequence pinned at HEAD `2c18091d1`. Non-vacuity proven by the relocation mutation at E.2.2. |
| AC-IHP-009 | PASS | `go test ./internal/cli/ -run 'TestRunInit_WizardCodexDeclinesMCPProvisioning' -count=1` plus the wizard-`both` row | green — wizard harness `codex` + wizard `mcp_provision: yes` ⇒ announcement absent, AND `.codex/config.toml` carries `mcp_servers.moai`. The `both`-from-wizard row (forcing provisioning over an explicit decline) is asserted too, so the rule is shown independent of the selection's origin. RED at E.2.1 RED-2a. |
| AC-IHP-010 | PASS | `go test ./internal/cli/ -run TestMCPPrecedenceComment -count=1` | green — `init.go` no longer contains the falsified justification string, and states the harness-selection rule. RED: the `grep -q` succeeded at HEAD. The behavioural test (AC-IHP-009) is the binding gate; this is its companion. |
| AC-IHP-011 | PASS | `go test ./internal/cli/ -run 'TestInitAgentFlag\|TestValidateInitFlags_Agent\|TestRunInit_Agent\|TestRunInit_Codex\|TestRunInit_CallsCodexWiring' -count=1 -timeout 900s` | green — the closed set, the fail-loud `invalid --agent value "gemini"` rejection, and all landed `--agent` behaviour tests pass unmodified. `validateInitFlags` was not touched. |
| AC-IHP-012 | PASS | two §7 greps, re-run at close | `grep -rniE 'agent_wiring\|agent_harness\|harness_wiring' internal/template/templates/.moai/config/ \| wc -l` → `0`; `grep -rn "moai init" internal/template/templates/ \| grep -i agent \| wc -l` → `4` — identical to the spec.md §7 baseline, and the 4 hits are the same pre-existing ones (none documents `--agent` or enumerates the wizard's questions). Additionally `grep -rn 'agent_wiring' internal/template/templates/ \| wc -l` → `0`. Template-First determination holds; no mirror, no `make build`. |

**13 / 13 PASS. 0 FAIL. 0 AC weakened.**

### E.2.4 Deviations from the plan, recorded rather than absorbed silently

1. **`WizardResult.AgentWiring` was added in the seam cycle, not in M3.** plan.md M1 gives the helper
   the signature `(..., res *wizard.WizardResult) agentWiring`, which cannot compile against a struct
   with no harness field, while M3 owns "add the field to `WizardResult`". The field was therefore
   added first, as the minimum needed to compile the RED test; the question literal and the
   `saveAnswer` capture branch stayed in M3. The milestone ORDERING (seam before question before
   translations before comment) is unchanged.

2. **The `agentWiring` type doc in `internal/cli/init.go` was corrected as well.** REQ-IHP-010 names
   the comment governing the `.mcp.json` switch. The type doc one screen above carried the same
   now-false claim (flag-beats-wizard, "resolved in ONE place — `resolveAgentWiring`"), and the seam
   change is exactly what falsifies it. Leaving it would have left the SPEC's own correction
   contradicted by the adjacent comment. This is a same-SPEC cascade within the scope envelope, not a
   drive-by edit.

3. **The replacement comment initially quoted the falsified sentence verbatim** to record what was
   falsified — which tripped the AC-IHP-010 `NotContains` assertion on my own test. The comment was
   reworded to describe the former justification without reproducing the literal. The test was NOT
   weakened; the AC forbids the literal, and the AC won.

4. **Five landed wizard test assertions were updated** because they enumerate or count the Page-3 set,
   and a new question legitimately changes both: `expansion_test.go` (ID/type/order table + the
   visible-count assertions 16→17 and 10→11), `restructure_test.go` (page membership list),
   `wizard_test.go` (stepper denominator 17→18). These are cardinality/ordering guards doing their
   job, not obstacles; each was updated with the reason stated inline.

### E.2.5 Process incident, recorded because it was mine

While probing whether a `gofmt` deviation in `internal/cli/wizard/mcp_audit_test.go` predated my
change, I issued a compound diagnostic command that contained a bare `git stash`. It executed, and it
stashed all 7 modified files. `git stash list` showed the entry at `stash@{0}`
(`WIP on WT-init-harness-prompt: 2c18091d1`), and `git stash pop` restored all 7 and dropped the
entry; the pre-existing stash stack (18 entries, unrelated sessions) was unchanged before and after.
No work was lost, and the wizard package was re-run green after the restore to confirm it.

`git stash` is repository-global and is forbidden in the shared checkout for exactly this reason; a
worktree does not exempt it, because the stash stack is shared. The correct probe — used afterwards —
is `git show HEAD:<path> > /tmp/x.go; gofmt -l /tmp/x.go`, which touches no repository state.

**Finding from that probe**: `internal/cli/wizard/mcp_audit_test.go` is unformatted at HEAD
`2c18091d1` (`gofmt -l` on the HEAD blob reports it), and `git diff --stat` on it is empty — I have
not modified it. Pre-existing, not attributable to this change, and out of scope.

### E.2.6 Verification, full and unfiltered

```
$ go test ./internal/cli/wizard/ ./internal/cli/ -count=1 -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	3.144s
ok  	github.com/modu-ai/moai-adk/internal/cli	384.838s
```

`internal/cli` was run **unfiltered**, which closes the residual risk recorded at the end of §E.1: the
`wireCodexUnlessClaude` signature change is not covered by a `-run 'Agent'` baseline.

```
$ go vet ./internal/cli/...
(no output; exit 0)

$ golangci-lint run ./internal/cli/...
0 issues.
```

Lint is **0 issues total** on the changed scope, so the per-rule delta is necessarily zero — no new
finding is attributable to this change. (Plan-phase §E.1 gap 2 recorded that no golangci-lint
baseline existed; a zero total makes the baseline unnecessary for the "no NEW finding" claim, since
the count cannot fall below zero.)

`gofmt -l` is clean on every file this change authored or modified; the single reported file
(`mcp_audit_test.go`) is the pre-existing HEAD deviation described in §E.2.5.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: pending-backfill-run-commit
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
ac_weakened_count: 0
mutation_obligations_discharged: 4   # AC-IHP-005, 006a, 007 (ordering), 008
preserve_list_post_run_count: 0      # no PRESERVE-listed file modified
l44_pre_commit_fetch: performed      # git status --short + rev-parse HEAD + branch re-read at each commit
l44_post_push_fetch: n/a             # no push performed — integration is the lead's window
new_warnings_or_lints_introduced: 0  # golangci-lint ./internal/cli/... => 0 issues; go vet clean
cross_platform_build:
  darwin_arm64: pass                 # go build ./internal/cli/... exit 0 (native)
  windows_amd64: not_measured        # see Gaps
total_run_phase_files: 13            # 9 modified + 4 added (Go only; no .claude/ or .moai/ outside this SPEC dir)
m1_to_mN_commit_strategy: two commits — wizard cycle (M3+M5), then seam cycle (M1+M2+M6+M7)
template_first_reattached: false     # AC-IHP-012 re-verified at close: both greps unchanged
```

### Gaps — what was explicitly NOT observed

1. **Cross-platform build not measured.** `GOOS=windows GOARCH=amd64 go build ./...` was not run. The
   change adds no build tags, no syscall use, and no path handling, so the risk is low — but low risk
   is not an observation, and this is recorded as a gap rather than asserted as a pass. CI's matrix is
   the measuring surface.
2. **Full suite not run** (`go test ./...`), deliberately, per the lane-local verification rule and the
   load-413 incident. Only the two affected packages were measured. CI on `origin/develop` is the
   full-suite verdict.
3. **CI not observed.** Nothing was pushed and no PR exists, so there is no CI signal for this work at
   all. The local green is an early signal, never the verdict.
4. **`moai spec lint` not run** on this SPEC directory (plan-phase gap 1 carries forward — the
   full-catalog run exceeds 300s at 722 SPEC dirs).
5. **No real `moai init` executed outside the Go harness.** Plan-phase gap 5 is now partly closed: the
   announcement's presence/absence was observed through `runInit` inside the test harness with `HOME`
   pinned by `t.Setenv`, on every branch. What remains unobserved is a genuine terminal `moai init`
   with an interactive TTY — the wizard's rendered appearance (option labels, on-screen ordering,
   locale rendering) was never seen by a human in this run. The question's *data* is asserted; its
   *presentation* is not.
6. **`moai update --reconfigure` was not executed.** AC-IHP-008 asserts the reconfigure question set
   structurally (`ReconfigureQuestions`' ID sequence), not by running the command. Plan-phase gap 4
   (that path's own Codex wiring, `internal/cli/update_codex_wiring_test.go`) remains unread.
7. **The sibling-guard asymmetry is untouched**, as plan.md §G.1 directs:
   `TestRunInit_AgentClaudeLeavesNoCodexFiles` still asserts only two of the three artifacts. Card
   t405 carries it. My own tests assert three on every row.

### Residual risk — what could still be wrong despite what was observed

- **The `agent_wiring` question ID is new vocabulary.** Nothing outside this change consumes it yet,
  but a future config key, template surface, or docs reference choosing a different name (`harness`,
  `agent_harness`) would fork the vocabulary. The ID is asserted in four tests, so a rename would be
  loud — but the *choice* was mine, not the SPEC's, which names only "the harness axis".
- **The accepted UX wart is now live** (plan.md §B B1): a user selecting `codex` is still asked
  `mcp_provision`, and the answer is then overridden. This was settled by the operator and is
  deliberate, but it is the first time it is user-visible rather than flag-only.
- **`normalizeAgentWiring` silently maps an unrecognized value to claude.** For the flag path this is
  unchanged behaviour (invalid values are rejected earlier, fail-loud, by `validateInitFlags`); for
  the wizard path the options are a closed set by construction, so no unrecognized value can arise
  today. If a future edit adds a wizard option without a matching `agentWiring` constant, it will
  silently resolve to claude rather than fail loudly.
- **The unfiltered `internal/cli` run took 384s**, long enough that a flaky test elsewhere in the
  package could have passed by luck in this single run. `-count=1` was used, but the run was not
  repeated.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

- plan_complete_at: 2026-09-01T03:46:10Z
- plan_status: audit-ready
- plan_audit: iter-2 PASS-WITH-DEBT 0.85 (Tier M threshold 0.80), 7/7 must-pass; evidence .moai/reports/t393/plan-audit-iter2.md
- kickoff_approved: operator, 2026-09-01, autonomous progression mode
