# SPEC Review Report: SPEC-INIT-HARNESS-PROMPT-001

Iteration: 1/2 (Tier M ceiling, `harness.plan_audit_tier_ceilings.M = 2`)
Verdict: **FAIL**
Overall Score: **0.70** (harmonic mean of the four dimensions; Tier M PASS threshold is 0.80)

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t393`, branch `WT-init-harness-prompt`,
HEAD `2c18091d1`, tree `a33625497` (`git rev-parse HEAD^{tree}` re-measured this run — matches the
anchor the artifacts declare).

Reasoning context ignored per M1 Context Isolation. The dispatch's leads (H1, H2, the author's six
declared Gaps) were treated as hypotheses and each was re-measured against the tree; two of them
produced findings the author's own framing did not.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-IHP-001` … `REQ-IHP-011`, sequential, no gaps, no
  duplicates, uniform 3-digit padding (`spec.md:67-87`). AC IDs are `AC-IHP-001…012` with `006`
  split into `006a`/`006b` — a documented split, not a gap (see D8 for the resulting miscount).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer only**
  (`spec.md` §3 `REQ-XXX` entries); the Given/When/Then entries in `acceptance.md` are verification
  layer and were graded under Group 4, not here. All 11 REQs match a GEARS pattern: Ubiquitous
  (001, 004, 009, 010, 011), Event-driven (002, 003), State-driven (005), State-driven unwanted with
  canonical `shall not` (006, 007), Where (008). `moai spec lint <spec.md>` → `✓ No findings — all
  SPEC documents are valid`. One borderline noted as an optional finding (D10).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-15`): `id`, `title`, `version: "0.1.0"` (quoted), `status: draft` (enum member),
  `created`/`updated` ISO dates, `author`, `priority: P2`, `phase`, `module`, `lifecycle:
  spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at`/`updated_at`/`labels`/`spec_id`) is present. Extras `related_specs`, `tier: M` are
  permitted optional fields.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go only; `internal/cli`,
  `internal/cli/wizard`). No multi-language tooling surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — five referenced SPECs extracted and each read:
  `SPEC-AUTONOMY-TIERS-001`, `SPEC-CODEX-INIT-001`, `SPEC-CODEX-WIRING-001`,
  `SPEC-INIT-WIZARD-REPAIR-001`, `SPEC-MCP-DEFAULT-ON-001` — all exist under `.moai/specs/` and all
  carry `status: completed`. None is `retired`/`superseded`/`archived`, so no reconciliation clause
  is required. No BLOCKING finding. (A *substantive* reconciliation defect against
  SPEC-CODEX-WIRING-001 REQ-CW-001 is recorded as D5 — it is a wording/scope defect, not a D7
  lifecycle-status BLOCKING.)
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-PASS.
- **[FAIL] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-INIT-HARNESS-PROMPT-001/`
  returns:
  ```
  plan.md:25:**[NEEDS CLARIFICATION: should `mcp_provision` be conditionally hidden when the harness answer is `codex`?]**
  progress.md:22:- **Open marker**: one `[NEEDS CLARIFICATION: ...
  ```
  One unresolved marker in `plan.md` §B at audit time. `research.md` does not exist (Tier M), so the
  N/A carve-out does not apply — `plan.md` alone triggers the gate. **Score-independent must-pass
  failure.** The SPEC's own mitigation ("must be settled before Implementation Kickoff Approval",
  `plan.md:39`) points at a *downstream* gate; it does not satisfy this one. Resolution is one
  orchestrator `AskUserQuestion` round.

**M5 firewall**: MP-7 fails ⇒ Verdict FAIL regardless of aggregate score. The aggregate (0.70) is
independently below the Tier M threshold (0.80), driven by D1.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.65 | between 0.50 and 0.75 | Requirement prose is unusually precise and citation-dense (`spec.md:30-41`, §2 C1/C2 corrections). But the central MCP predicate — "`codex` declines `.mcp.json` provisioning" (`REQ-IHP-009`, `spec.md:83`; `D2`, `spec.md:115`) — names an internal branch as if it were a user-observable file state, and the ACs then assert the file state. Measured, the file state is identical either way (D1). A reasonable engineer implements `AC-IHP-009` as written and gets a permanently red test. |
| Completeness | 0.85 | 0.75 band | All required sections present: HISTORY (`spec.md:18`), WHY (§1), WHAT (§2/§3), REQUIREMENTS (§3), ACCEPTANCE CRITERIA (`acceptance.md` §B), Out of Scope (§5, four `### Out of Scope — <topic>` H3 headings each with specific `-` bullets, `spec.md:140-164`). Frontmatter complete. Deductions: the unresolved clarification, and the absent measured baseline for the one observable the change hinges on. |
| Testability | 0.55 | 0.50 band | 7 of 13 ACs are binary-testable as written (001, 002, 006b, 007, 008, 010, 011). Four carry a predicate that is false or vacuous against measured behaviour (003b, 004 `both` row, 006a, 009 — D1). One (005) states a count with no named instrument and no counting seam anywhere in `internal/` (`grep -rn 'promptCount\|PromptCount\|issuePrompt' internal/ --include='*.go'` → no matches) — D3. |
| Traceability | 0.85 | 0.75 band | §C matrix maps every REQ to ≥1 AC and every AC to a REQ or to a named guard; re-derived by hand and it holds. Deductions: REQ-IHP-004 ("reaches **both** consumers") maps only to AC-IHP-003, whose second leg is the broken one — so its coverage is nominal, not real (D1); and the AC count is stated as 12 where 13 IDs exist (D8). |

Aggregate 0.70 = harmonic mean(0.65, 0.85, 0.55, 0.85), per the skeptical-evaluation stance
(harmonic, not arithmetic — arithmetic would read 0.725).

---

## Defects Found (structured defect-list)

**D1. AC-IHP-009 / AC-IHP-003(b) / AC-IHP-004(`both` row) / AC-IHP-006a — `acceptance.md:46, 70, 90, 127` — the `.mcp.json` predicate is not a real observable; measured — Severity: critical — Class: blocking**

The ACs assert file state: "`.mcp.json` is **not** provisioned" (009), "`.mcp.json` provisioning is
declined" (003b), "`.mcp.json` is provisioned" (004 `both`, 006a). Measured at this tree, `.mcp.json`
**exists with the `moai` entry in every one of those cases**, because it is a *template-deployed*
file (`internal/template/templates/.mcp.json`, 398 bytes) — `provisionMCPEntryUnlessDeclined` only
*ensures* the entry, and `mcpDeclined = true` merely skips that call.

Evidence — a transient probe test in `internal/cli` (created, run, deleted; `git status --short`
after: only the untracked SPEC dir):

```
PROBE wizard-yes_no-flag:      .mcp.json statErr=<nil>  announced=true
PROBE wizard-yes_agent-codex:  .mcp.json statErr=<nil>  announced=false
```

(`announced` = stdout contains `Provisioned the moai MCP server entry`.) A second probe on the
flag-absent non-interactive path also found `.mcp.json` present, carrying
`mcpServers.moai.command = "moai"` plus `context7`.

Consequences, in order of cost:

1. **AC-IHP-003 does not in fact force both consumers.** Its whole purpose (`acceptance.md:46`,
   restated as anti-pattern §G "Half-wiring", `plan.md:156-158`) is to make a single test fail on a
   half-wired implementation. Leg (b) cannot be asserted as written; whoever weakens it to make the
   test compile-and-pass produces exactly the single-consumer test the AC was written to forbid.
   This is the failure mode this SPEC most plausibly ships.
2. **AC-IHP-004's `both` row and AC-IHP-006a are vacuous** on the `.mcp.json` clause — true before
   and after the change, in every branch.
3. **`REQ-IHP-009`'s rule text** ("`codex` declines provisioning, `both` forces it") describes a
   branch, not an outcome. §4 D2's supporting fact 1 (`spec.md:121`) is *more* true than argued —
   declining withdraws nothing because the entry arrives from the template anyway — which makes the
   decision defensible but its stated observable wrong.

**Required fix**: re-specify the MCP consumer's observable to something the system actually
produces, and say so in `REQ-IHP-009` as well as in the four ACs. The measured discriminator is
available and cheap: the stdout line `Provisioned the moai MCP server entry in .mcp.json
(default-on).` is emitted on the provisioning path and absent on the declining path (probe above,
`announced=true` vs `false`). Add a run-phase helper that returns the captured stdout buffer
(`runInitForAutonomyAtHome` currently discards it, `internal/cli/init_autonomy_wiring_test.go:48-55`),
and assert on that line — optionally alongside a direct assertion on the resolved-wiring value that
reaches the `:911-916` switch. Also state, in `spec.md` §4 D2, that `.mcp.json` is template-deployed
so "declined" never means "absent".

**D2. Unresolved `[NEEDS CLARIFICATION]` marker — `plan.md:25` — Severity: critical — Class: blocking**

MP-7 failure (evidence quoted above). Judged on the merits, as the dispatch asked:

- *Is it genuinely a user decision?* Yes. Both options are implementable and neither is forced by
  the code; the axis is UX (asking a question whose answer is then overridden) against an
  interpretation burden ("not asked" reads as "declined" for one field).
- *Is the stated consequence correct?* Verified. `Condition func(*WizardResult) bool` exists
  (`internal/cli/wizard/types.go:108`) and `FilteredQuestions` drops conditioned-out questions
  (`internal/cli/wizard/questions.go:307-316`), so a hidden confirm never reaches `saveBoolAnswer`
  (`internal/cli/wizard/wizard.go:459-476`), leaving `MCPProvision` at its `bool` zero value —
  while its Page-3 neighbours `TodoEnabled`/`FeedbackAutoSubmit` are `*bool`
  (`internal/cli/wizard/types.go:62, 68, 84`) precisely to keep "unasked" distinguishable.
- *Is leaving it open legitimate?* As a **question**, yes. As a **plan-phase exit state**, no — the
  clarification gate is score-independent, and this SPEC additionally cannot answer it well until D1
  is settled (under `codex` the two options are observationally identical on a fresh init, since the
  template writes `.mcp.json` regardless).

**Required fix**: orchestrator runs one `AskUserQuestion` round (default (i) "ask unconditionally"
is the sound recommendation — it changes no existing field semantics), records the answer in
`plan.md` §B replacing the marker, and updates `plan.md` M3's "Resolve §B's clarification before
writing the question's `Condition`" to state the decision.

**D3. AC-IHP-005 names a count with no instrument — `acceptance.md:80-84` — Severity: major — Class: blocking**

"the count of harness prompts issued is exactly **0**. The criterion is the count, not whether an
answer arrives." No prompt-counting seam exists: `grep -rn 'promptCount\|PromptCount\|issuePrompt'
internal/ --include='*.go'` → no matches. The cited precedent, SPEC-CODEX-INIT-001 AC-CI-004, could
judge a count because that SPEC's M2 *introduced* a proposal seam (`acceptance.md:344` of that SPEC:
"M2 — 제안 seam + 생성기 인자 포착 seam"); this SPEC neither inherits one nor scopes one — `plan.md`
M4 says "Assert prompt-issuance **count zero** … then prove the assertion binds by mutation" without
naming what does the counting. The mutation step compounds it: "deliberately remove the
non-interactive gate" removes the entry condition for the **whole** 16-question wizard
(`internal/cli/init.go:644`), so an observed FAIL is not attributable to the harness question, and
in a non-TTY test process `huh` is as likely to error as to prompt.

**Required fix**: either (a) scope the seam — state in `plan.md` M4 that a counting seam is added
(e.g. a package-level hook the wizard field-builder increments, injected in tests) and name it; or
(b) re-state AC-IHP-005 against an observable that exists, e.g. "`runWizardFn` is not invoked" /
"`WizardResult` is the zero value at the resolution point" — and if (b) is chosen, delete the
"count, not whether an answer is produced" sentence rather than leaving a criterion the test does not
measure. Then re-specify the mutation as one that isolates the harness question (e.g. force *only*
that question into the non-interactive path), not one that removes the shared gate.

**D4. AC-IHP-006a and REQ-IHP-007 omit the trust sidecar — `acceptance.md:90`, `spec.md:79` — Severity: major — Class: blocking**

Both name exactly two artifacts (`.codex/hooks.json`, `.codex/config.toml`). `codexwiring.Wire`
writes **three**: `HooksRelPath = ".codex/hooks.json"`, `ConfigRelPath = ".codex/config.toml"`,
`SidecarPath = ".moai/state/codex-wiring.json"` (`internal/codexwiring/codexwiring.go:29-35`,
written at `internal/codexwiring/wire.go:187`). The tree's own landed regression guard already
asserts all three and documents why `.codex/` itself may exist (template ships `.codex/agents/**`):
`TestRunInit_AgentAbsentLeavesNoCodexFiles`, `internal/cli/init_agent_flag_test.go:97-108`. This
SPEC's preservation AC is therefore **weaker than the guard it claims to preserve** — the direct
consequence of the author's declared Gap 5 (the writer was never read). The dependency the dispatch
asked about is load-bearing: a REQ-CW-001 preservation criterion that does not enumerate what the
writer writes cannot detect a partial wiring.

**Required fix**: add `.moai/state/codex-wiring.json` to REQ-IHP-007 and AC-IHP-006a, and cite
`internal/cli/init_agent_flag_test.go:97-108` as the existing three-path guard the new assertion
mirrors. State the `.codex/agents/**` caveat so run-phase does not assert `.codex/` absent.

**D5. "extends it and does not reverse it" mischaracterises REQ-CW-001 — `spec.md:61` — Severity: major — Class: blocking**

REQ-CW-001's clause is unconditional over the flag-absent path, read verbatim at
`.moai/specs/SPEC-CODEX-WIRING-001/spec.md:229-231`: "**플래그 부재 하에서** `moai init`은 오늘의
동작과 동일하게 수행한다 — … `.codex/hooks.json` 및 `.codex/config.toml` 미생성." It is not scoped
to non-interactive; only its verifying criterion AC-CW-004 happens to use `--non-interactive`
(`.moai/specs/SPEC-CODEX-WIRING-001/acceptance.md:51-54`). After this change, a flag-absent
**interactive** run that selects `codex` creates both files — which the clause as written forbids.
The change therefore **narrows** a `completed` SPEC's requirement to `flag-absent ∧
non-interactive`; REQ-IHP-007 states that narrowed boundary precisely and correctly, so the SPEC is
internally coherent — but §2 tells a future reader the opposite of what happened, and no amendment
record exists on the SPEC that owns the clause.

**Required fix**: rewrite the §2 paragraph to say the change **narrows** REQ-CW-001's flag-absent
clause to the flag-absent ∧ non-interactive sub-path, name the sub-path it no longer covers
(flag-absent ∧ interactive ∧ wizard-selects-codex), and record whether the owning SPEC needs an
amendment note or whether the narrowing is carried by this SPEC alone.

**D6. Cited count does not reproduce — `spec.md:105` — Severity: minor — Class: blocking**

The D1 rejection cites `grep -rn "agentWiring" internal --include="*.go" | grep -v "_test.go"` → "14
lines". Re-running that exact command at the declared tree returns **15** (lines
`internal/cli/init.go:135,139,142,143,144,151,152,153,154,156,167,397,398,913,915`). The claim it
supports — every reference lives in `init.go` control flow — is **true and verified**; only the
number is wrong. (The rest of the D1 rejection also verifies: `InitOptions` carries no harness field
(`internal/core/project/initializer.go` field list read in full), and no writer anywhere consumes an
agent-harness selection — the `harness` keys in `internal/config/` are the *quality* harness, a
different concept.)

**Required fix**: replace `14` with `15`, or re-word to the claim the command actually supports
("every non-test reference is in `internal/cli/init.go`").

**D7. Mis-cited line range for the clarification's supporting evidence — `plan.md:31` — Severity: minor — Class: blocking**

Cites `saveBoolAnswer (internal/cli/wizard/wizard.go:475-490)`. Measured: `func saveBoolAnswer` is at
`:459` and its body ends at `:476`; `475-490` lands mostly in `buildConfirmField`. The substantive
claim is correct (verified above under D2); the coordinate is not. Same class of drift, smaller:
`acceptance.md:39` cites `saveAnswer` at `:410-448` where the function starts at `:397`, and
`spec.md:132` cites the completeness guard at `translations_completeness_test.go:93-131` where the
function spans `:95-135` — both land inside their subject, so those two are acceptable.

**Required fix**: correct to `:459-476`.

**D8. AC count is stated as 12; there are 13 — `acceptance.md:192` (§D.1), and §C prose — Severity: minor — Class: blocking**

IDs present: 001, 002, 003, 004, 005, 006a, 006b, 007, 008, 009, 010, 011, 012 = **13**. §D
Definition of Done item 1 reads "All 12 ACs pass", which is a gate that under-counts its own subject
by one. (Both counts stay inside the Tier M ceiling of 16, so no tier consequence.)

**Required fix**: state 13, or renumber `006a`/`006b` to `006`/`007` and shift the tail.

**D9. Flag default described imprecisely — `spec.md:37` — Severity: minor — Class: optional**

"The flag is declared at `internal/cli/init.go:132` over the closed set `{claude, codex, both}`,
defaulting to `claude`." The cobra default is the **empty string** (`initCmd.Flags().String("agent",
"", …)`, `init.go:132`); `claude` is `resolveAgentWiring`'s fallback (`:151-158`) and the help
string's prose. The distinction is not cosmetic for this change: D1's helper takes `(flagChanged
bool, flagValue string, …)`, and an implementer who believes the flag's default is `"claude"` may
reasonably test `flagValue != ""` and `flagChanged` inconsistently. `validateInitFlags` also
short-circuits on `agent != ""` (`:395`), which is the same distinction.

**Required fix**: say "declared with an empty default; `claude` is the resolution fallback", and make
`plan.md` M1 state which of `flagChanged` / non-empty `flagValue` is authoritative.

**D10. `Where` used for a runtime locale condition — `spec.md:81` (REQ-IHP-008) — Severity: minor — Class: optional**

GEARS reframes `Where` as capability gate / feature flag / static config. `conversation_language` is
persisted configuration, so this reads as static config and passes MP-2 — but a `While` (state-driven)
phrasing would be unambiguous. No action required; recorded so a later reader does not re-litigate it.

**D11. No AC covers the wizard composition end to end — `acceptance.md` §B — Severity: minor — Class: optional**

The behavioural ACs (003, 004, 009) will run through `runInitForAutonomy`, which **stubs**
`runWizardFn` (`internal/cli/init_autonomy_wiring_test.go:36-38`), so the real path
question → `saveAnswer` → `WizardResult` is never exercised in the same test as the resolution.
AC-IHP-001 (question exists) and AC-IHP-002 (field round-trips) cover the two halves separately,
which is defensible — a composition test would need a real `huh` driver. Recorded as a known seam,
not a required fix.

---

## Verification of the dispatch's named hazards

**H1 — vacuous acceptance criteria.** Checked per-cell. The seven "RED now" cells that claim a
command and an output were each re-run and each reproduces: AC-IHP-001/006b
(`grep … internal/cli/wizard/ | wc -l` → `0`), AC-IHP-002 (`grep -cin 'harness\|agentwiring'
internal/cli/wizard/types.go` → `0`), AC-IHP-003/004 (`grep -rn
'applyAgentWiringFromWizard\|resolveAgentWiringWithWizard' internal/` → no lines; `grep -n
'Changed("agent")' internal/cli/init.go` → no lines; `resolveAgentWiring`'s body quoted at
`acceptance.md:52-59` is byte-accurate against `internal/cli/init.go:151-158`), AC-IHP-010 (`grep -q
'the flag beats the wizard answer'` succeeds; the comment is at `:908-910` verbatim). **No
counterfactual or future-tense RED cell was found** — the t343 defect class is absent, and each of
the five non-failable ACs does say so in its own cell rather than implying a failure. That part of
the authoring is sound.

The substitute obligations were judged individually for executability, not accepted as a set:

| AC | Mutation obligation | Executable & checkable? |
|---|---|---|
| AC-IHP-007 | add question before translations; run the named test; record the FAIL naming the ID for ko/ja/zh; then add entries | **Yes** — verified mechanically: the guard iterates `InitQuestions(root)` and errors `question %q has NO translation entry` (`translations_completeness_test.go:95-135`), the only exemption is `conversation_language`, and `Page3Questions` reaches `InitQuestions` by construction (`questions.go:296-305`). Cheapest and strongest of the four. |
| AC-IHP-008 | temporarily place the question in `DefaultQuestions`; observe the leak-guard FAIL | **Yes** — `ReconfigureQuestions` splices `DefaultQuestions` (`questions.go:268-287`), so that is genuinely the leak path. Weakness: the AC's expected ID sequence is given by reference ("unchanged from HEAD") rather than enumerated (the 12 IDs are re-derivable: 5 Default + 7 Git). |
| AC-IHP-006a | mutate the resolution helper's default `claude` → `codex`; observe FAIL; restore | **Yes** — mechanically sound, but the assertion it proves is under-specified (D4). |
| AC-IHP-005 | remove the non-interactive gate; observe FAIL | **No** — aspiration, not procedure (D3). |

**H2 — the two design decisions.**

*D1 (resolution seam).* The rejection of the `InitOptions` field is **verified independently and
holds**: `InitOptions` carries no harness field; every non-test `agentWiring` reference is
`init.go` control flow; no writer anywhere consumes an agent-harness selection (the `harness` keys in
`internal/config/` belong to the quality-harness subsystem). The chosen shape's structural claims
also hold: `runInit` spans `:454`→`:940` with no intervening `func` (`grep -n '^func '
internal/cli/init.go`), so a local at `:717` is in scope at both `:911-916` and `:925`; and
`applyAutonomyTierFromWizard` (`internal/cli/init_autonomy_wizard.go:34`) is a real, matching
precedent. The reversal condition is recorded in `spec.md:107` **and** pointed at from `plan.md:89`
(M1), so run-phase will see it. **D1 is sound.**

*D2 (harness wins).* The type premise is **verified**: `MCPProvision bool`
(`internal/cli/wizard/types.go:84`) against `TodoEnabled *bool` / `FeedbackAutoSubmit *bool` (`:62`,
`:68`), and the question's `Default: "true"` (`questions.go:463-469`) makes explicit-yes and
default-yes identical. The conclusion follows: the alternative rule is unimplementable without a
`*bool` conversion, which §5 excludes. The alternative was assessed on implementability rather than
dismissed. **The decision is right; its stated observable is wrong (D1 above)** — the SPEC should
record that "declines provisioning" means "does not run the ensure-entry call", not "leaves the
project without `.mcp.json`".

**Both consumers.** As specified, the AC set does **not** force it — see D1, consequence 1.

**The author's six declared Gaps, treated as leads.** Gap 1 (full-catalog lint) — this spec.md lints
clean; cross-SPEC effects still unmeasured, and this audit did not close that. Gap 2 (the "16") —
re-derived structurally: `DefaultQuestions` 5 + `Page3Questions` 11 = 16 for `InitQuestions`, 23 IDs
total, `GitQuestions` 7 excluded; no conditional/dynamic append exists in either constructor. No REQ
or AC leans on the number, so it is background, not an unattributed load-bearing figure. Gap 3
(`golangci-lint`) — not run here either; `acceptance.md` §D.4 correctly demands a per-rule delta.
Gap 4 (darwin only) — this change touches no platform-conditional code; low risk. Gap 5
(`codexwiring.Wire`'s write targets) — **read this run, and it produced D4**; the two-file claim in
`init.go`'s comments is incomplete. Gap 6 (`moai update --reconfigure`) — the out-of-scope finding
survives: `ReconfigureQuestions` is built from `DefaultQuestions` only, so a question placed in
`Page3Questions` cannot reach it by construction, and AC-IHP-008 asserts rather than assumes it.

**Other dispatch questions.** *Template-First*: verified — both §7 greps reproduce exactly (`0`, and
the same four unrelated hits). The determination stands; worth adding one line, since `.mcp.json`
**is** template-managed (`internal/template/templates/.mcp.json`) even though this change does not
edit it. *Tier M vs S*: justified — two packages, a new file, ~6 source files plus tests, a
cross-package seam change; Tier S is `<5 files` / `<300 LOC`. REQ 11 and AC 13 are both inside the
Tier M ceiling of 16.

---

## Gaps (what this audit did NOT observe)

- Cross-model convergence (`mcp__moai__audit_multi`) was **not** run. Reason: no `audit_model` key
  exists in `.moai/config/` (grep → no matches), and the SPEC artifacts are untracked
  (`git status --short` → `?? .moai/specs/SPEC-INIT-HARNESS-PROMPT-001/`), so the diff-collecting
  backends would return `inconclusive` on an empty diff. The verdict rests on mechanical
  measurement, not on a second opinion.
- Full-catalog `moai spec lint` was not run (same 300 s / 722-dir cost the author hit). Cross-SPEC
  lint effects remain unverified.
- `golangci-lint` was not run; no platform other than darwin was exercised.
- The transient probe measured the **flag** path (`--agent codex`), not a wizard-selected `codex`,
  because the wizard axis does not exist yet. The two share the same `:911-916` switch, so the
  measurement transfers — but that transfer is an inference, stated here rather than buried.

## Residual risk

Fixing D1 by weakening the `.mcp.json` assertions instead of replacing the observable would leave
the SPEC nominally green and structurally half-wired — the exact outcome D1 exists to prevent. The
re-audit should check that AC-IHP-003 still fails a deliberately half-wired implementation, by
mutation, before accepting the repair.

---

## Recommendation

FAIL. Iteration 2 is available (Tier M ceiling 2). Fix in this order — the first two are the ones
that change what run-phase builds:

1. **D1** — replace the `.mcp.json` file-state predicate in `REQ-IHP-009`, `AC-IHP-003(b)`,
   `AC-IHP-004` (`both` row) and `AC-IHP-006a` with the measured discriminator (stdout
   `Provisioned the moai MCP server entry in .mcp.json (default-on).`, plus/or a direct assertion on
   the resolved wiring reaching `internal/cli/init.go:911-916`). Add a `spec.md` §4 D2 note that
   `.mcp.json` is template-deployed, so "declined" never means "absent". Add to `plan.md` M2 the
   helper change that surfaces the captured stdout buffer
   (`internal/cli/init_autonomy_wiring_test.go:48-55` currently discards it).
2. **D2** — orchestrator resolves the clarification via one `AskUserQuestion` round; replace the
   marker in `plan.md:25` with the decision and its rationale; update `plan.md` M3 accordingly.
   Recommended default: (i) ask unconditionally.
3. **D3** — either scope a prompt-counting seam in `plan.md` M4 and name it, or restate AC-IHP-005
   against an observable that exists; re-specify the mutation so it isolates the harness question
   rather than removing the shared gate at `internal/cli/init.go:644`.
4. **D4** — add `.moai/state/codex-wiring.json` to `REQ-IHP-007` and `AC-IHP-006a`; cite
   `internal/cli/init_agent_flag_test.go:97-108` as the mirrored guard; note the `.codex/agents/**`
   caveat.
5. **D5** — rewrite `spec.md:61` as a *narrowing* of REQ-CW-001, naming the sub-path no longer
   covered.
6. **D6 / D7 / D8** — correct `14`→`15` (`spec.md:105`), `:475-490`→`:459-476` (`plan.md:31`),
   `12`→`13` ACs (`acceptance.md:192`).
7. **D9** — optional but cheap: state the flag's real default and which of `flagChanged` /
   non-empty `flagValue` is authoritative.

D10 and D11 are optional-class: surface them, do not route them into a revision unless the author
wants them.

The re-audit is scoped to this enumerated delta plus a regression check over it, per the Retry Loop
Contract — not a from-scratch re-audit.
