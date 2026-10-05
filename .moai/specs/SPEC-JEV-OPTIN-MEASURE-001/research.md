# SPEC-JEV-OPTIN-MEASURE-001 — Research

Read-only findings from this worktree at `fd75cf692`, plus the inherited measurement record. Anything inherited rather than measured here is labelled as such.

---

## §1. Wizard surface

`Question` (`internal/cli/wizard/types.go:78-92`) carries `ID`, `Type`, `Title`, `Description`, `Options`, `Default`, `Required`, `Condition`, `Group`. `Option` (`:94-99`) carries `Label`, `Value`, `Desc`.

Translations map language code → question ID → `QuestionTranslation{Title, Description, Options}`, with all four locales present (`ko` at line 30, `ja` at 117, `zh` at 203, `en` at 293).

**The constructor split is the load-bearing fact for the operator's decision.** From the file's own header comment and `questions.go:290-322`:

- `DefaultQuestions` returns five (conversation_language, user_name, project_name, model_policy, report_format) and is shared with the reconfigure path.
- `Page3Questions` returns the init-only two (`agent_wiring`, `autonomy_tier`).
- `InitQuestions` picks `conversation_language` and `user_name` from `DefaultQuestions` by id (`initSharedQuestionIDs`, line 293) and appends `Page3Questions` — **four questions today**.
- `ReconfigureQuestions` is `DefaultQuestions` with `GitQuestions` spliced in, and **deliberately does NOT include the page-3 questions**, to keep its pre-restructure membership.

So the operator's init-only decision makes `InitQuestions` five, and makes the setting genuinely unreachable from `moai update --reconfigure` — not merely inconvenient. That is why REQ-JEVO-006 states the consequence and REQ-JEVO-007 requires a pointer to `moai web`.

---

## §2. Web console surface

`internal/web/app.go:154-201` registers 11 routes: `/`, `/kanban`, `/monitor`, `/todo`, `/settings`, `/events`, `/save`, `/specs`, `/profile/create`, `/profile/delete`, plus the profile-rename and GLM-key-reveal paths and `/__shutdown__`.

**There is no init route**, which confirms the operator's out-of-scope call on designing an init flow inside the console.

All routes sit behind `hostCheckMiddleware`: a loopback Host check on every method and route (including `/static/`), plus a `Sec-Fetch-Site` same-origin check on state-changing methods. There is no per-process CSRF token and the source states that is deliberate.

---

## §3. One persistence path

`internal/web/projectconfig.go:194-214` and `:257-276` record the AP-2 rule: the load-modify-write seam was relocated to the neutral `internal/settings` package so the web console and the TUI wizard drive one writer.

The nested-isolation crux is spelled out at `:264-273` — `SetSection` replaces the WHOLE section struct and `Save()` serializes the whole struct, so the seam copies the ENTIRE section struct returned by `LoadRaw` and mutates ONLY the targeted nested field; every sibling nested field rides through byte-identical, and each `*Set` flag gates a per-field mutation so an unsubmitted field keeps its persisted value.

`workflow.yaml` is large (the model-routing profiles alone are 36 entries), which is what makes AC-JEVO-003 worth asserting rather than assuming.

---

## §4. Credential disclosure precedent

`internal/web/glmkey.go` is the pattern for the credential field: a hand-built parse / validate / view path, deliberately OUT of the schema `FieldDef` set so no generic schema-walking loop can pick it up, with `TestGLMKeyField_AbsentFromSchema` as the structural guard.

`computeGLMKeyHint` returns `Configured: true` with an empty hint for a key of four characters or fewer, and the source notes a naive "last four or the whole key" fallback would disclose a short key entirely.

The GLM precedent also carries a reveal route (`glmKeyRevealPath`, POST-only, loopback-gated). Whether Jev wants one is open question Q4, owned by `SPEC-JEV-CORE-001`; nothing in this SPEC requires it.

---

## §5. The inherited negative result (measurement-design lessons)

**Labelled as inherited, not re-measured.** The figures originate in `.moai/reports/t943/verdict.md`, a gitignored local evidence file in the primary checkout that may be absent from any given tree. Nothing in this research pass re-ran the experiment.

Using Jev to judge whether a stale card's premise is dead was measured on 124 Korean-original cards and rejected:

| Measure | Result |
|---|---|
| `premise_dead` call precision | 14/48 = **29.2%** against a **25%** base rate |
| 2-class accuracy | **58.9%** |
| Constant "always alive" answer | **75.0%** |
| English control (40 cards) | 3-class 60.0%, 2-class 67.5% — still below the constant |
| Threshold sweep | flat across 0.30-0.80; higher gates cut adoption without lifting accuracy |
| Errors vs confidence | 67 of 85 errors (79%) above 0.30, peaking at 0.93 |

Four measurement-design lessons, each of which shapes a requirement in this SPEC:

1. **The constant baseline is the number to beat** (REQ-JEVO-008/009). A 16-point headline gain (31.5% → 47.6% on 3-class) looked like success while the constant sat at 75.0% untouched.
2. **A threshold sweep is not a rescue** (REQ-JEVO-012/013). Flat across the whole range; the two high-gate cells that looked better rested on 5 and 4 samples.
3. **High confidence does not imply correctness** (REQ-JEVO-012/020). The assumption that errors cluster at low confidence was measured false.
4. **The English arm is worth measuring and does not change a verdict by itself** (REQ-JEVO-010/011). A ~10-point effect on a 40-card subsample translated by the same model family — which is that control's own weak point, and is recorded as such rather than smoothed over.

A distinction the record also draws, which this SPEC's gate preserves: the same tool used as a **marker** rather than a judgment narrowed 124 cards to 48 to inspect, yielding 14 — a 2.6× narrowing. That usefulness exists only while "it only marks" holds; read as a judgment it skips 34 live cards wrongly.

---

## §6. Known model weaknesses (vendor-stated, applied)

From the published jev-1.13 jaggedness notes, each mapped to a design response in `design.md` §5: literal reading (weak on negation, scoping words, implied conditions); unreliable counting, arithmetic, numeric proximity, and date ordering; weaker multi-hop indirection; accuracy degrading as irrelevant state grows; vulnerability to instructions injected in state; and `P(yes)` not guaranteed to equal `1 − P(no)`. Request limits: 64k total, 32k for state plus longest question (enforced by `SPEC-JEV-CORE-001`).

Vendor claims, not measurements taken here. They are used as design constraints — reasons to compute values in Go, keep state small, and add a no-match option — which is safe even if a given claim is conservative. No requirement asserts any of them as a measured property of this repository's data.

---

## §7. Open items

| # | Item | Why it is not settled here |
|---|---|---|
| Q3 | Labelled-set size per consumer | Depends on each consumer's base rate, which M3 measures. The rejected task used 124 cards — a precedent, not a target. |
| R1 | Which surface carries the `moai web` pointer | REQ-JEVO-007 permits the question `Description` or the init completion output. The second reaches a user who has already declined, which is arguably the user who later wants the switch back. Implementation-time choice. |
