# SPEC Review Report: SPEC-INIT-HARNESS-PROMPT-001

Iteration: 2/2 (Tier M ceiling, `harness.plan_audit_tier_ceilings.M = 2`)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85** (harmonic mean; Tier M PASS threshold 0.80)
Delta from iteration 1: **0.70 → 0.85, monotonic** — every dimension non-decreasing. No score
regression, so no STOP signal and no scope-reduction proposal.

Tree state re-verified before starting: `git rev-parse --short HEAD` → `2c18091d1`;
`git branch --show-current` → `WT-init-harness-prompt`; `git rev-parse HEAD^{tree}` → `a33625497…`
— identical to iteration 1. `git status --short` → only `?? .moai/specs/SPEC-INIT-HARNESS-PROMPT-001/`
and `?? .moai/reports/t393/`. No foreign writer.

Artifacts re-read at `spec.md` v0.1.1 (23,203 B), `plan.md` (11,710 B), `acceptance.md` (23,624 B),
`progress.md` (5,237 B). The author's iteration-2 revision table (`progress.md:41-52`) was treated
as a set of claims; every row was re-measured here. Two of them do not survive.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-IHP-001…011`, unchanged, sequential, no duplicates.
- **[PASS] MP-2 GEARS format compliance** — requirement layer only; all 11 REQs match a GEARS
  pattern. `moai spec lint <spec.md>` → `✓ No findings — all SPEC documents are valid`.
- **[PASS] MP-3 YAML frontmatter validity** — 12 canonical fields present, types correct;
  `version: "0.1.1"` bumped with a matching HISTORY row (`spec.md:21`).
- **[N/A] MP-4 language neutrality** — single-language (Go) SPEC.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the five referenced SPECs are unchanged and all
  `completed`; no `retired`/`superseded`/`archived` reference. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — no `syscall` token in the SPEC body.
- **[PASS] MP-7 clarification gate — CLOSED** — `grep -rn 'NEEDS CLARIFICATION'
  .moai/specs/SPEC-INIT-HARNESS-PROMPT-001/` → **no matches (rc=1)**, over the whole directory, not
  only `plan.md`. The marker is genuinely gone, not reworded into a deferral: `plan.md:25-44`
  Decision B1 states the outcome ("asked UNCONDITIONALLY … no `Condition` func"), its accepted cost
  ("a user who selects `codex` is asked a question whose answer is subsequently overridden … accepted,
  not hidden"), its reversal condition (if `MCPProvision` becomes `*bool`), and — the part that makes
  it reach run-phase — a **consequence clause** binding M3: "An implementer who adds one is
  contradicting a settled decision, not exercising discretion." `plan.md` M3:138 repeats the
  constraint at the point of use. This is recorded where run-phase actually reads.

**M5 firewall: all seven must-pass criteria satisfied.** The iteration-1 FAIL driver is cleared.

---

## Category Scores

| Dimension | iter1 | iter2 | Band | Evidence |
|-----------|------:|------:|------|----------|
| Clarity | 0.65 | **0.80** | 0.75 | The MCP predicate is now defined rather than assumed: `REQ-IHP-009`'s "Observable" paragraph (`spec.md:112`) and `acceptance.md` §A.1 both name the stdout line and explicitly say absence-of-file is NOT the observable. §2 C3 states the REQ-CW-001 narrowing precisely. Three residual interpretation hazards remain (R2, R3, R4 below), each localized and cheap. |
| Completeness | 0.85 | **0.90** | 0.75→1.0 | MP-7 closed with cost + reversal recorded; §A.1 added; M2.0 added as a blocking prerequisite with DoD item 2b tracking it; the "cannot fail today" population corrected from five to six with `012` included (`acceptance.md:237`, `plan.md:77-79`). Deductions: the M2.0 blast radius was declared as a gap rather than measured (one grep — I measured it below), and AC-IHP-006a's announcement leg has no pinned expected value (R2). |
| Testability | 0.55 | **0.85** | 0.75→1.0 | The largest mover, and it is earned rather than asserted — see the measurements below. AC-IHP-003 now fails today on **both** legs; the AC-IHP-004 `both` row pin is verified non-vacuous by measurement; AC-IHP-005 has a real pre-existing instrument with exact mutation attribution; M2.0 is concrete, not aspirational. Deductions: R2 (an implementer following AC-IHP-006a literally writes a wrong assertion) and R1 (the AC-IHP-005 substitution works, but its stated justification is measurably false). |
| Traceability | 0.85 | **0.85** | 0.75 | Matrix updated with the pins and a reconciled, executed count (13). Deductions: REQ-IHP-005 ↔ AC-IHP-005 is now only *partial* coverage — the REQ still demands a prompt count the AC deliberately does not measure, and the deviation is recorded on the AC side only (R4); and `spec.md:229` contradicts §2 C3 (R3). |

Aggregate 0.85 = harmonic mean(0.80, 0.90, 0.85, 0.85). Arithmetic would read 0.85 as well.

---

## Re-measurement of each iteration-1 defect

Every claim below was re-run in this worktree. Where a probe was needed, it was a transient
in-package test file, deleted immediately; `git status --short` afterwards shows only the two
untracked directories (quoted above).

### D1 (was critical) — **RESOLVED for AC-003/004/009; PARTIALLY ADDRESSED for AC-006a**

The decisive question — *does the AC set now force the answer to reach BOTH consumers?* — is
**yes**, and it is verified rather than reasoned.

*Leg (b) fails today.* AC-IHP-003 now requires the captured stdout to **not** contain the
announcement, **with `mcp_provision` pinned to yes**. Measured (iteration-1 probe, same tree):
`wizard-yes_no-flag → announced=true`. So today, with `--agent` absent and a wizard `codex` answer,
the announcement **is** emitted while the AC requires it absent — a mechanical RED, exactly as
`acceptance.md:93` claims. The pin does what is claimed: without it, a `mcp_provision: no` default
would make the announcement absent for the wrong reason and leg (b) would pass vacuously.

*Half-wiring now fails.* Routing the wizard answer only to `wireCodexUnlessClaude` leaves the
`:911-916` switch resolving `claude`, so with `mcp_provision: yes` the announcement is emitted and
leg (b) fails. Routing it only to the switch leaves the three Codex artifacts absent and leg (a)
fails. Neither half passes. **REQ-IHP-004's coverage is now real, not nominal.**

*The `both` row pin is verified non-vacuous.* AC-IHP-004 pins `mcp_provision: no` and asserts the
announcement **is** emitted. Measured this run:

```
PROBE wizard-mcp-no_flag-absent    announced=false
PROBE wizard-mcp-no_agent-both     announced=true
PROBE wizard-mcp-yes_agent-codex   announced=false
```

The announcement appears on the `both` row **only** because `agentWiringBoth` flips `mcpDeclined`
back to false. Without the pin the row would assert nothing. The claim at `acceptance.md:104` holds.

*M2.0 is concrete, not the aspiration shape rejected in iteration 1.* Verified line by line:
`internal/cli/init_autonomy_wiring_test.go:48` is `var out, errBuf bytes.Buffer`, `:55` is
`return projectDir` — both coordinates exact. M2.0 names the file, both coordinates, the action
("return the captured stdout"), the alternative ("or add a sibling"), the ordering ("Do M2.0 first"),
the failure it prevents, and it is tracked by DoD item 2b. Contrast with the rejected D3 mutation,
which named no instrument at all. **This is executable and checkable.**

*The one thing not carried through.* See **R2** below: AC-IHP-006a and REQ-IHP-007 swapped the
observable to the announcement but never pinned which way it goes at HEAD.

### D3 (was major) — **SUPERSEDED by operator decision, and the instrument is real — but its soundness argument is measurably false**

*The instrument exists and is already in use.* Verified: `runWizardFn` is declared
`var runWizardFn = func(rootFlag, locale, userName string) (*wizard.WizardResult, error)` at
`internal/cli/init_update_notice.go:69`, and has exactly **one** non-test call site,
`internal/cli/init.go:654`, inside the block opening at `:644`
(`grep -rn 'runWizardFn' internal/cli/*.go | grep -v _test` → decl `:66`/`:69`, comment `:651`, call
`:654`). The swap idiom is already exercised at `internal/cli/init_autonomy_wiring_test.go:36-38`.
Nothing new is instrumented — that part of `acceptance.md:116` is accurate.

*The mutation-attribution claim holds.* The assertion's subject is the seam invocation; the `:644`
gate directly guards the `:654` call, so mutating the gate moves the counter from 0 to 1 and nothing
else. It is **not** diluted across the other 15 questions, because the assertion never mentions
questions. This is a genuine improvement over the iteration-1 formulation, and the author's claim
about it is correct.

*The load-bearing premise is false as written* — see **R1**. `runInit` contains a **second,
independent `huh` prompt** outside `runWizardFn`.

### D4 (was major) — **RESOLVED, and the author's additional finding is verified real**

REQ-IHP-007 (`spec.md:104-106`) and AC-IHP-006a (`acceptance.md:126-129`) now enumerate all three
artifacts and cite the constants. Verified: `HooksRelPath` / `ConfigRelPath` / `SidecarPath` at
`internal/codexwiring/codexwiring.go:29, 31, 34`; sidecar written at `internal/codexwiring/wire.go:187`.
The `.codex/agents/**` caveat is carried, and the per-file (never per-directory) discipline is stated.

*The asymmetry the author reported independently is real.* Read this run:
`TestRunInit_AgentAbsentLeavesNoCodexFiles` (`internal/cli/init_agent_flag_test.go:97-108`) iterates
**three** paths including the sidecar; its sibling `TestRunInit_AgentClaudeLeavesNoCodexFiles`
(`:109-117`) iterates a two-element slice `{".codex/hooks.json", ".codex/config.toml"}` — the sidecar
is omitted. This is a pre-existing weakness in a landed guard for a `completed` SPEC, not something
this change introduces.

*Is "record it as a warning" the right disposition?* **Half right.** Keeping the warning inside
AC-IHP-006a is correct and useful: it is exactly where an implementer would otherwise copy the weak
two-path pattern, and it names the corrective ("assert three on both rows rather than copy the
two-path sibling"). But a note in one SPEC's AC body has no owner and no criterion binding the
*sibling test itself*, so the weak guard stays weak after this card closes. **Recommendation**: keep
the warning, and additionally file it as its own queue card (`moai todo add`) so it is acted on
rather than only observed. Strengthening someone else's landed test is out of this SPEC's scope; that
is a reason to route it, not a reason to leave it as a footnote.

### D5 (was major) — **PARTIALLY ADDRESSED**

§2 C3 (`spec.md:75-86`) is now correct and precise. Verified against the source clause read this run
at `.moai/specs/SPEC-CODEX-WIRING-001/spec.md:230-231` (the quote is accurate; it begins on 230, not
229 — a one-line-off citation, noted, not scored). C3 states what is superseded
(`flag-absent` → `flag-absent ∧ wizard-not-run`), names the sub-path no longer covered
(`flag-absent ∧ interactive ∧ wizard-selects-codex|both`), points at REQ-IHP-007 as the narrowed
clause, and records the amendment carrier.

*Is the amendment mechanism legitimate?* **Yes at plan-phase, with one gap.** The heavier machinery —
`completed → in-progress` plus `amendment_of:` plus a HISTORY `## Amendments` sub-section
(`spec-frontmatter-schema.md`) — is designed for amending a SPEC's own content, and invoking it to
record that a *different* SPEC narrowed a clause would reopen a closed SPEC's lifecycle at plan time
for a documentation edit. C3's disposition (carry the record here; defer any in-place note to
sync-phase policy) is proportionate. **The gap is directionality**: `related_specs` on *this* SPEC
points outward, so a reader arriving at REQ-CW-001 from SPEC-CODEX-WIRING-001 finds nothing pointing
forward to the narrowing. Recommendation: record a sync-phase obligation now — when this SPEC
completes, add a one-line forward pointer at REQ-CW-001. Until then the carrier is real but
one-directional, and C3 slightly overstates it by saying "a reader arriving at REQ-CW-001 is directed
to this SPEC".

Residual contradiction: **R3** below.

### D6 / D7 / D8 / D9 (minor) — **RESOLVED, with one over-correction introduced**

Spot-checked rather than re-derived, as instructed:

| Figure | Claimed | Measured this run | |
|---|---|---|---|
| `agentWiring` non-test refs | 15 | 15 | ✓ |
| `saveBoolAnswer` span | 459-476 | `awk` → `459-476` | ✓ |
| `saveAnswer` span | 397-448 | `awk` → `397-448` | ✓ |
| completeness guard span | 95-135 | `awk` → `95-135` | ✓ |
| AC count | 13 | `grep -c '^### AC-IHP-'` → `13` | ✓ |
| cobra flag default | `""`, `claude` is the fallback; `flagChanged && flagValue != ""` authoritative | `init.go:132` verbatim; `plan.md` M1:104-110 states both conjuncts with the `--agent ""` and attribution reasons | ✓ |
| `InitQuestions` span | `296-305` | `awk` → **296-303**; `sed -n '294,308p'` confirms `func` at 296, closing `}` at 303 | ✗ **R5** |

### Gap 2 (the 16-question figure) — **CLOSED, arithmetic and premise both verified**

Arithmetic: 5 + 11 = 16 interactive, + 7 Git = 23 total, which reconciles against my own direct count
of 23 `ID:` literals in the file. The **non-looping premise is true**: inspecting every `for`/`append(`
hit inside the three constructor bodies shows they are all the English word "for" inside comments and
`Description:` strings (e.g. "recommended default for solo developers", "Gate for the codex
reviewer") — there is **no `for` statement and no `append(` in `DefaultQuestions`, `Page3Questions`,
or `GitQuestions`**. A crude `grep -c` returns 2/6/9 and would have produced a false defect claim;
the hits were opened before judging. The stated scope ("counts `ID:` literals … not a runtime
`len()`") is therefore honest and the count equals the returned slice length.

### The author's new gap 1 (HOME-pinned probe blocked) — **claim is TRUE; the explanatory text is accurate**

The author could not run a real `moai init` because the worktree guard refuses a `HOME`-setting
invocation. I hit the identical refusal this round and routed around it the same way as in iteration 1
(a transient in-package test using `t.Setenv`, which the guard permits). Judging both halves:

- *Does the D1 fix depend on the unmeasured fact?* **No.** The announcement predicate observes which
  branch ran, and my four measured rows confirm it discriminates correctly in every combination
  regardless of the file's existence. The fix stands on its own.
- *Is the explanatory text accurate?* **Yes**, and its citations check out: `.mcp.json` is 398 bytes;
  `//go:embed all:templates` is at `internal/template/embed.go:28`; `mcp_template_neutrality_test.go`
  says "the template-managed `.mcp.json` surface" (`:3`) and "The distributed `.mcp.json`" (`:5`).
  I separately measured in iteration 1 that `.mcp.json` exists with the `moai` entry in both the
  provisioned and declined cases (`statErr=<nil>` in both rows), which is precisely what §A.1 asserts.
  The author confined an unmeasurable claim to explanation and built no criterion on it — correct
  handling of a blocked measurement.

### The author's new gap 6 (M2.0 helper blast radius unmeasured) — **acceptable as a plan-phase gap**

I measured what the author declared: `runInitForAutonomyAtHome` has **5** call sites outside its
definition (`init_audit_wiring_test.go:39`, `init_autonomy_wiring_test.go:67, 184, 187`,
`init_workflow_wiring_test.go:30`); its wrapper `runInitForAutonomy` has **6**
(`init_agent_flag_test.go:71, 83, 98, 110`, `init_autonomy_wiring_test.go:162, 222`). Four files,
≤11 sites, and in Go a changed return arity fails to compile at every one of them — the blast radius
is bounded, mechanical, and cannot fail silently. The plan's own "(or add a sibling)" reduces it to
zero. **Not under-specified**; the gap is honest but small. Recommendation: default to the sibling.

---

## Residual defects (carried into run-phase as debt)

**R1. The AC-IHP-005 soundness argument rests on a measurably false premise — `acceptance.md:117`, `plan.md:151` — Severity: major — Class: blocking — NEWLY INTRODUCED this round**

Both artifacts state: "the wizard is the only prompt issuer on the init path — `runWizardFn` is the
sole entry to `huh` from `runInit` — so zero invocations entails zero harness prompts." Measured, that
is false. `runInit` issues a **second, independent `huh` prompt** at `internal/cli/init.go:598-612`:

```go
if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) && !profile.IsSetup(profileName) {
        var wantSetup bool
        confirm := huh.NewConfirm().
                Title("No profile found. Set up profile preferences now?").
        ...
        confirmForm := huh.NewForm(huh.NewGroup(confirm)).WithTheme(moaiHuhTheme())
        if err := confirmForm.Run(); err == nil && wantSetup {
                if err := runProfileSetup(cmd, nil); err != nil {
```

and on acceptance it calls `runProfileSetup`, which opens further `huh` forms
(`internal/cli/profile_setup.go:129-133, 321-327`). Two further details: this surface sits **before**
the `:644` gate, and it reads `isatty.IsTerminal(os.Stdin.Fd())` **directly** rather than through the
injectable `isInteractiveStdin` seam, so it is neither swappable by tests nor touched by the
AC-IHP-005 gate mutation.

The AC's *conclusion* survives — the harness question lives in `Page3Questions` → `InitQuestions`,
which nothing but `runWizardFn` renders, so zero seam invocations does entail zero **harness**
prompts. But the sentence offered as the justification for deviating from the cited discipline is
wrong about the code, and the residual-gap paragraph ("It would not catch a harness prompt issued
from outside `runWizardFn`") reads as hypothetical when a second prompt surface demonstrably exists
in the same function.

**Required fix**: replace the entailment with the narrower one that is true — "the harness question
can only be rendered inside `runWizardFn`, because it lives in `Page3Questions` → `InitQuestions` and
nothing else renders that set" — and add the measured caveat that `runInit` carries a second,
unrelated `huh` prompt at `:598-612` (profile setup) which cannot render a wizard question and is
gated separately on `nonInteractive` + a direct `isatty` call. Same edit in `plan.md` M4.

**R2. AC-IHP-006a / REQ-IHP-007 swapped the observable but never pinned its direction — `acceptance.md:126`, `spec.md:104` — Severity: major — Class: blocking — PARTIALLY ADDRESSED carry-over of D1**

Both now say the provisioning announcement is "emitted exactly as at HEAD `2c18091d1`" without saying
which way HEAD goes. Measured this run:

```
PROBE noninteractive_flag-absent    announced=false
```

On the flag-absent, non-interactive path the announcement is **absent** — because `opts.MCPProvision`
is assigned only inside the interactive block (`internal/cli/init.go:286`, reached from `:704`), so it
holds its `bool` zero value and `mcpDeclined = !false` short-circuits `provisionMCPEntryUnlessDeclined`
at `:208-209`. An implementer reading "the announcement is emitted exactly as at HEAD" will most
naturally assert **presence** and get a RED unrelated to this change. The hazard is sharpened by
SPEC-CODEX-WIRING-001 AC-CW-004, which asserts the `.mcp.json` **entry** is present on that same path
— true, from the template — so the two observables point opposite ways and the SPEC swapped to the
announcement without re-measuring it on this row.

**Required fix**: pin the value. "…and the provisioning announcement is **absent** on this path, as at
HEAD `2c18091d1` (measured: `opts.MCPProvision` is set only on the interactive path, `init.go:286`, so
`mcpDeclined` is true and the `:216` line is not reached)". Do not leave a self-referential "as at
HEAD" on the one AC whose whole job is preservation.

**R3. `spec.md:229` still says REQ-CW-001 is "extended here", contradicting §2 C3 — Severity: minor — Class: blocking — PARTIALLY ADDRESSED carry-over of D5**

§8 Cross-references: "`SPEC-CODEX-WIRING-001` — REQ-CW-001 (flag-absent regression clause, **extended
here**)". §2 C3 says the opposite in bold, and the HISTORY row records the change away from that exact
word. A reader who consults only §8 gets the superseded framing.

**Required fix**: "flag-absent regression clause, **narrowed here** — see §2 C3".

**R4. REQ-IHP-005 still demands a prompt count its AC deliberately does not verify — `spec.md:100`, `spec.md:230` — Severity: minor — Class: blocking — NEWLY INTRODUCED this round**

The requirement layer is unchanged from v0.1.0: "The obligation is on the prompt-issuance count, not
on whether an answer is produced (SPEC-CODEX-INIT-001 AC-CI-004 discipline)", echoed at `spec.md:230`
("the discipline REQ-IHP-005 adopts"). The verification layer now measures seam invocations and
declares the deviation — but only on the AC side. The result is a REQ whose stated obligation its only
AC does not meet, which makes that matrix row nominal coverage.

**Required fix**: one sentence in REQ-IHP-005 recording that the verifying criterion measures wizard-seam
invocations rather than prompts, with the reason and the residual gap pointed at AC-IHP-005; adjust
`spec.md:230` to say the discipline is adopted *in spirit, with a stated deviation*.

**R5. `InitQuestions` span over-corrected from a value that was already right — `spec.md:31`, `acceptance.md:149`, `progress.md:49` — Severity: minor — Class: blocking — NEWLY INTRODUCED this round**

v0.1.0 said `questions.go:296-303`; the revision changed it to `:296-305` and `progress.md:49` records
that change as one of the D7 corrections. Measured: `awk '/^func InitQuestions/{s=NR} s&&/^}/{print
s"-"NR; exit}'` → `296-303`, confirmed by reading `sed -n '294,308p'` (func at 296, closing brace at
303). The coordinate was correct before and is wrong now.

**Required fix**: revert to `296-303` in all three places, and drop that row from the
`progress.md` correction table.

---

## Gaps (what this audit did NOT observe)

- Full-catalog `moai spec lint` was again not run (the 300 s / 722-dir cost is unchanged). Only this
  SPEC's `spec.md` was linted; cross-SPEC lint effects remain unverified across both iterations.
- `golangci-lint` was not run; darwin only.
- The `--agent codex` rows were measured through the **flag** path, since the wizard axis does not
  exist yet. Flag and wizard converge on the same `:911-916` switch, so the measurement transfers —
  an inference, stated rather than buried.
- I did not re-verify the unchanged D2 type premise or the D1 seam evidence from iteration 1; both
  were measured then and neither artifact section changed materially.
- Cross-model convergence (`mcp__moai__audit_multi`) was again not run: no `audit_model` key exists in
  `.moai/config/`, and the artifacts are untracked, so the diff-collecting backends would return
  `inconclusive` on an empty diff.

## Residual risk

R2 is the one that can still cost a run-phase cycle: an implementer who asserts announcement
*presence* on the flag-absent non-interactive row gets a RED that looks like a regression in the
change and is not. If run-phase hits an unexplained RED on AC-IHP-006a, the first thing to check is
the direction of that predicate, not the resolution helper.

Second-order: M2.0 is correctly ordered but is a *test-infrastructure* change gating four behavioural
ACs. If it is deferred "just to see the tests compile", the four ACs will be written against file
state and the iteration-1 defect returns wholesale. DoD item 2b is the guard; it should be recorded in
`progress.md` §E.2 with the extended helper named, exactly as written.

---

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria are satisfied, the aggregate (0.85) clears the Tier M
threshold (0.80), the delta is monotonic across every dimension, and the two defects that made
iteration 1 a FAIL — the unresolved clarification and the unobservable MCP predicate — are closed with
evidence I re-measured rather than accepted. The SPEC is implementable as written and its gates now
bind: AC-IHP-003 provably fails a half-wired implementation, and that was the failure this SPEC was
most likely to ship.

Five residual corrections are owed. All are text edits; none blocks starting run-phase, and R3/R4/R5
are one-line each:

1. **R1** — replace the false "only prompt issuer" entailment in `acceptance.md:117` and `plan.md:151`
   with the narrower true one, naming the second `huh` surface at `internal/cli/init.go:598-612`.
2. **R2** — pin the announcement's direction on the flag-absent non-interactive row (`absent`) in
   `acceptance.md:126` and `spec.md:104`.
3. **R3** — `spec.md:229`: "extended here" → "narrowed here — see §2 C3".
4. **R4** — one sentence in `REQ-IHP-005` recording the seam-invocation substitution and its residual
   gap; adjust `spec.md:230`.
5. **R5** — revert `InitQuestions` to `:296-303` in `spec.md:31`, `acceptance.md:149`, and drop the row
   from `progress.md:49`.

Two routing recommendations that are not SPEC edits:

- File the `TestRunInit_AgentClaudeLeavesNoCodexFiles` two-path weakness
  (`internal/cli/init_agent_flag_test.go:109-117`) as its own queue card. The warning in AC-IHP-006a
  correctly stops the pattern being copied, but nothing in this card will strengthen the sibling.
- Record a sync-phase obligation to add a forward pointer at SPEC-CODEX-WIRING-001 REQ-CW-001 when
  this SPEC completes, so the narrowing is discoverable from the narrowed clause and not only from the
  narrowing SPEC.

Iteration ceiling for Tier M is 2 and this is iteration 2: the residual list is the debt record, not a
third-iteration trigger. Verdict authority for the repairs above stays with this agent only if the
orchestrator elects to re-audit them; otherwise they travel as recorded debt into run-phase.
