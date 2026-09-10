# SPEC-INIT-HARNESS-PROMPT-001 — Implementation Plan

Ordered by decision-reversibility: the seam and the precedence rule come first (hardest to change once
landed and the only places a reviewer's disagreement is expensive), the user-facing question next, and the
mechanical work — translations, comment correction — last.

Measurement anchor for every file:line below: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t393`,
HEAD `2c18091d127cbc723074124e1015353e077300ca`, tree `a33625497c87246572f823cf427fe6335a79c825`.

---

## §A Context

`moai init` asks 16 interactive questions; none of them is the agent-harness axis, which is reachable only
through `--agent`. The card asks to add the question while keeping the flag for the non-interactive path.
The real work is not the question — it is that `resolveAgentWiring` takes `*cobra.Command` and reads the
flag, so **no wizard answer can reach either of its two consumers as the code stands**. spec.md §4 D1
(the resolution seam) and §4 D2 (the cross-axis justification that breaks once harness is a wizard axis)
carry the two decisions; this plan sequences them.

---

## §B Decisions taken, and known issues

**Decision B1 (settled by the operator, 2026-09-01) — `mcp_provision` is asked UNCONDITIONALLY.**
The harness question carries **no** `Condition` func. Every user is asked `mcp_provision` whatever they
answered on the harness axis; a `codex` answer then overrides it per spec.md §4 D2. This was previously
an open clarification; it is now closed, and the marker is removed.

- **Accepted cost, stated plainly**: a user who selects `codex` is asked a question whose answer is
  subsequently overridden. That is a real UX wart and it is accepted, not hidden.
- **Why it is the right trade** (measured this run): the alternative — hiding the question with a
  `Condition` func (`internal/cli/wizard/types.go:108`, applied by `FilteredQuestions`,
  `internal/cli/wizard/questions.go:307-316`) — means the confirm never reaches `saveBoolAnswer`
  (`internal/cli/wizard/wizard.go:459-476`), leaving `WizardResult.MCPProvision` at the `bool` zero value
  `false`. Because that field is a plain `bool` while its Page-3 neighbours `TodoEnabled` and
  `FeedbackAutoSubmit` are `*bool` precisely to keep "unasked" distinguishable, hiding makes "not asked"
  and "declined" indistinguishable for this one axis. Asking unconditionally changes no existing field
  semantics.
- **Reversal condition**: if `WizardResult.MCPProvision` ever becomes `*bool`, the zero-value hazard
  disappears and conditional hiding becomes safe — revisit this decision then. Converting it is out of
  scope here (spec.md §5).
- **Consequence for M3**: the question is added with no `Condition` field. An implementer who adds one is
  contradicting a settled decision, not exercising discretion.

**Known issue, not blocking**: no test asserts the cardinality of `InitQuestions`
(`grep -rn 'len(InitQuestions' internal/cli/wizard/*_test.go` → no count assertion), so adding a question
cannot break a count guard — and equally, nothing would notice a question silently dropped. Out of scope.

---

## §C Pre-flight

1. Re-read HEAD and branch immediately before any commit (`git rev-parse --short HEAD`, `git branch --show-current`).
2. Re-run the §A baseline of acceptance.md on the tree about to change:
   `go test ./internal/cli/wizard/ ./internal/cli/ -count=1`.
3. Confirm the SPEC directory is the only `.moai/` path this card touches (`git status --short`, explicit pathspec staging only).
4. Re-run the spec.md §7 Template-First greps; a non-zero result re-attaches the mirror + `make build` obligation.

---

## §D Constraints

- **Scope is two packages**: `internal/cli` (seam, precedence, call sites, comment) and `internal/cli/wizard`
  (question, result field, capture branch, translations), plus their tests.
- **`codexwiring` and `provisionMCPEntryUnlessDeclined` are untouched** — what they write is out of scope.
- **No persistence**: the harness selection reaches no config file (spec.md §5).
- **Verification is lane-local**: run the two affected packages, then let CI judge the full suite. Do not run
  `go test ./...` locally.
- **Template-First determined not to apply** (spec.md §7), re-verified at close (AC-IHP-012).

---

## §E Self-verification

Each milestone below closes only when its named ACs are recorded in `progress.md` §E.2 with the command run
and its verbatim output. The six ACs that cannot fail at this tree (AC-IHP-005, AC-IHP-006a, AC-IHP-007, AC-IHP-008,
AC-IHP-011, AC-IHP-012) close only with their **observed mutation FAIL**, ordering record, or close-time output recorded — a guard never seen failing is not
evidence.

---

## §F Milestones

### M1 — The resolution seam (highest reversibility cost; do first)

Introduce `internal/cli/init_agent_wizard.go` with the pure helper

```
resolveAgentWiringWithWizard(flagChanged bool, flagValue string, res *wizard.WizardResult) agentWiring
```

mirroring `applyAutonomyTierFromWizard` (`internal/cli/init_autonomy_wizard.go:34`) minus the `opts` write
(spec.md §4 D1 records why no `InitOptions` field is added, and the condition that would reverse that).

- Compute it **once** in `runInit`, adjacent to the existing `applyAutonomyTierFromWizard` call
  (`internal/cli/init.go:717`), into a local.
- Change `wireCodexUnlessClaude` (`internal/cli/init.go:166`, called at `:925`) to take the resolved wiring
  instead of `*cobra.Command`, so it can no longer re-read the flag behind the resolution point.
- Change the `.mcp.json` precedence switch (`internal/cli/init.go:911-916`) to read the same local.
- Keep `resolveAgentWiring(cmd)` as the flag-reading primitive the helper calls, so
  `internal/cli/init_agent_flag_test.go:120-145` keeps its subject and stays green.

**Which input is authoritative, stated because the flag's cobra default is `""` and not `claude`**
(`initCmd.Flags().String("agent", "", …)`, `internal/cli/init.go:132`): the helper takes the flag branch
when **`flagChanged && flagValue != ""`** — both conjuncts, matching `applyAutonomyTierFromWizard`
(`internal/cli/init_autonomy_wizard.go:34-40`) and consistent with `validateInitFlags`, which
short-circuits on `agent != ""` (`internal/cli/init.go:395`). `flagChanged` alone is not sufficient
(`--agent ""` is explicitly set and empty); a non-empty `flagValue` alone is not sufficient either, since
the resolution fallback must stay attributable to absence rather than to a value.

ACs: AC-IHP-003, AC-IHP-004, AC-IHP-011.

### M2 — The precedence rule and the MCP interaction (D2)

**M2.0 (blocking prerequisite)** — extend `runInitForAutonomyAtHome` (or add a sibling) to **return the
captured stdout**. Today it declares `var out, errBuf bytes.Buffer` at
`internal/cli/init_autonomy_wiring_test.go:48`, binds them with `cmd.SetOut/SetErr`, and returns
`projectDir` alone at `:55`. Without this, the MCP-side observable that AC-IHP-003/004/006a/009 assert
(the `Provisioned the moai MCP server entry in .mcp.json (default-on).` line,
`internal/cli/init.go:216`) is unreachable from a test — and the predictable failure is that someone
weakens those ACs to file-existence assertions, which are true in every branch because `.mcp.json` is
template-deployed. Do M2.0 first.

**Default to adding a sibling rather than changing the signature.** The helper is shared, and the coupling is compile-enforced — a signature change breaks every caller at build time, which is loud but wide. Measured this run: `grep -rn 'runInitForAutonomy' internal/cli --include="*_test.go"` returns **16** matches across **4** files, of which 2 are the two `func` declarations and 3 are comment mentions, leaving **11 real invocation sites**. A sibling (`runInitForAutonomyCapturingOut`, or an `AtHome` variant returning the buffer) touches none of them. Change the signature only if the sibling would duplicate meaningful setup; record which was chosen and why in `progress.md` §E.2.

Then wire the `codex` / `both` branches to the resolved value and add the behavioural test for the
wizard × wizard combination that is unreachable today, pinning `mcp_provision: no` on the `both` row so it
is not vacuous.

ACs: AC-IHP-009, and the flag-vs-wizard table completing AC-IHP-004.

### M3 — The question and its capture

- Add the harness question to `Page3Questions` (`internal/cli/wizard/questions.go`), placed **immediately
  before** `mcp_provision` (`:463`) so the overriding answer is given first. Select type, options
  `claude` (recommended) / `codex` / `both`, default `claude`, `Group: "Quality & Workflow"`.
- Add the field to `WizardResult` (`internal/cli/wizard/types.go`) and the capture branch to `saveAnswer`
  (`internal/cli/wizard/wizard.go:397-448`).
- Per Decision B1, the question carries **no** `Condition` func.

ACs: AC-IHP-001, AC-IHP-002, AC-IHP-006b, AC-IHP-008 (leak guard + its mutation).

### M4 — Non-interactive guard, proven non-vacuous

Assert that the injectable wizard seam **`runWizardFn` is invoked zero times** under `--non-interactive`
and under an absent TTY, by swapping in a counting stub — the same swap idiom the tree already uses
(`internal/cli/init_autonomy_wiring_test.go:36-38`). The seam is declared at
`internal/cli/init_update_notice.go:69` and invoked once at `internal/cli/init.go:654`, inside the
`if !nonInteractive && isInteractiveStdin()` block opening at `:644`.

This is a **deliberate deviation from SPEC-CODEX-INIT-001 AC-CI-004's letter** (which counts prompts).
The claim it rests on is narrow and measured: `runWizardFn` is the sole issuer of **the harness
question** — the question renders only through `Page3Questions` → `InitQuestions` → `runWizardFn` — so
zero invocations entails zero harness prompts. It is **not** the sole prompt issuer on the init path: a
second `huh` confirm (profile setup, `internal/cli/init.go:598-613`) reads `isatty.IsTerminal` directly,
bypasses both injectable seams, and runs before the `:644` gate. The residual gap is stated in
acceptance.md AC-IHP-005 with that counterexample named, and accepted. **Do not build a prompt-counting
instrument**; it is out of scope.

Then prove the assertion binds by mutating the `:644` gate and observing the counter FAIL, and re-assert
the REQ-CW-001 preservation row on all **three** Codex artifacts with its own mutation (AC-IHP-006a).

ACs: AC-IHP-005, AC-IHP-006a.

### M5 — Locale coverage, in the order that produces a real RED

Add the question first, run
`go test ./internal/cli/wizard/ -run TestWizardQuestionTranslationCompleteness -count=1`, record the
**FAIL** naming the new ID for ko/ja/zh, then add the three entries to `internal/cli/wizard/translations.go`
(ko `:32`, ja `:183`, zh `:333`) and re-run to green. Do not add the question to
`optionTranslationExemptIDs`.

AC: AC-IHP-007.

### M6 — Comment correction (mechanical, last)

Replace the falsified justification at `internal/cli/init.go:908-910` ("the flag beats the wizard answer")
with the D2 rule, and add the `! grep -q` doc assertion alongside — never in place of — the behavioural
test from M2.

AC: AC-IHP-010.

### M7 — Close

Re-measure both packages on the tree that will be committed; re-run the Template-First greps; record every
AC's command and output in `progress.md` §E.2/§E.3.

AC: AC-IHP-012.

---

## §G Anti-patterns to avoid

- **Half-wiring**: routing the wizard answer to `wireCodexUnlessClaude` only, or to the `.mcp.json` switch
  only. AC-IHP-003 asserts both in one test precisely because a half-wired implementation passes a
  single-consumer test.
- **Translations before the question**: makes AC-IHP-007's guard pass without ever having failed.
- **A file-only precedence test for the `--agent claude` row**: it passes identically before and after,
  because the wizard answer is discarded today. Assert the helper's return directly.
- **Persisting the harness selection** "while we are here" — out of scope, and it drags in a template surface.
- **Changing `MCPProvision` to `*bool`** to implement the rejected D2 alternative.
- **Running `go test ./...` locally.**
- **Treating the comment grep as the D2 gate**: the behavioural test is the gate.

---

## §G.1 Dispositions recorded, not acted on

- **The sibling-guard asymmetry is OUT OF SCOPE for this SPEC.** `TestRunInit_AgentClaudeLeavesNoCodexFiles` (`internal/cli/init_agent_flag_test.go:109-117`) asserts only two of the three artifacts, omitting the sidecar. acceptance.md AC-IHP-006a keeps the warning so run-phase does not copy the two-path form — but **strengthening that landed test is not this SPEC's work** and belongs in a separate card. Do not widen scope to fix it here.
- **REQ-CW-001's back-link is one-directional; closing it is a SYNC-PHASE obligation.** This SPEC narrows REQ-CW-001 (spec.md §2 C3) and points at it, but a reader arriving at SPEC-CODEX-WIRING-001 sees nothing pointing back here. Sync-phase MUST add a forward pointer from that SPEC's REQ-CW-001 to SPEC-INIT-HARNESS-PROMPT-001 recording the narrowing. It is deliberately not done at plan-phase: the narrowing is not in effect until the change lands, and SPEC-CODEX-WIRING-001 is `status: completed`.

## §H Cross-references

- spec.md §4 D1 / D2 — the two decisions this plan sequences, with their rejection records and D1's reversal condition.
- `internal/cli/init_autonomy_wizard.go` — the precedence-helper shape M1 copies (SPEC-INIT-WIZARD-REPAIR-001 REQ-005).
- `internal/cli/wizard/translations_completeness_test.go:95-135` — the guard that covers a new ID automatically.
- SPEC-CODEX-INIT-001 `plan.md:51` — prompt-issuance-count-zero discipline adopted by AC-IHP-005.
