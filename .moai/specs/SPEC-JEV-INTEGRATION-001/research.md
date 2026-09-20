# SPEC-JEV-INTEGRATION-001 — Research

Read-only findings from this worktree at `fd75cf692`, plus the inherited measurement record. Every claim below names how it was observed; anything inherited rather than measured is labelled as such.

---

## §1. What exists today

| Question | Command run | Observed |
|---|---|---|
| Does any Go code reach TypeSafe? | `grep -ril "typesafe\|jev" internal/ pkg/ cmd/` | 0 files |
| — positive control | `grep -ril "glmcred" internal/ cmd/` | 5 files (`internal/config/envkeys.go`, `internal/config/settings_axis_test.go`, `internal/glmcred/glmcred.go`, `internal/glmcred/glmcred_test.go`, `internal/web/glmkey.go`) |
| Is `scripts/jev/` on develop? | `ls scripts/` | 16 entries, no `jev` |
| Is the SPEC ID free? | `ls -d .moai/specs/SPEC-JEV-INTEGRATION-001` | absent |
| — positive control | `ls -d .moai/specs/SPEC-AGENT-001` | resolves |
| SPEC ID conforms? | `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` | `PASS` |

The zero-hit on the first row is paired with a positive control because a zero-hit and a broken search are indistinguishable on their own — the control shows the search fires on a comparable symbol in the same tree.

**Consequence.** This is greenfield in Go. Nothing is being ported, adapted, or kept compatible; the only inheritance is the measurement record in §5 and the local script behaviour, which is not committed and therefore cannot be cited as canonical.

---

## §2. The backlog finding contract

`internal/kanban/backlog_store.go` carries the pieces Consumer C extends.

**Two source constants (lines 117-120)**, and they partition by observer: `mechanical` = the text analyser measured it; `agent` = a reader who understands the cards judged it, written through `todo relate`.

**`Source` has no validator.** It is a plain `string` field. `AppendFindingOnce` dedupes on the tuple and does not check membership. But no CLI path accepts an arbitrary value either — `todo_relate.go:75` hardcodes `BacklogSourceAgent`, and `todo_analysis.go:62,71` hardcode `BacklogSourceMechanical`. So the set is closed in practice by its write paths, not by a check, and a third source needs a new write path rather than a new flag.

**`HasAgentFindingForPair` (lines 408-416)** selects on `Source == BacklogSourceAgent`. Its doc comment states it is the predicate behind the `machine-only` mark and that the mark records the *absence* of an agent-sourced record for a pair, "never that any review took place". This is the constraint that forces a third constant: a Jev finding written as `agent` would silently clear the mark from pairs nobody reviewed.

**`todoFindingLine` (todo_analysis.go ~215-240)** is source-conditioned twice: the `machine-only` mark renders only for `mechanical` findings lacking an agent counterpart, and the score renders only for `mechanical` — with the reason stated in the source, that an agent judgement carries no measurement and `0.00` would read as measured dissimilarity. A third source therefore needs a third render form, or it inherits the "no score" branch and loses its probability.

**Findings are records and nothing else.** The `BacklogFinding` doc comment notes this is structural rather than conventional: folding, reordering, dropping, or editing a card in response to a finding would each need code that does not exist. Consumer C inherits that property for free and must not be the thing that breaks it.

**Tests enumerating the source set** (so likely to need declaration-only updates): `internal/kanban/backlog_findings_test.go`, `backlog_archive_test.go`, `todo_merge_procedure_test.go`, `todo_queue_merge_test.go`, `internal/cli/todo_analysis_test.go`, `todo_analysis_add_test.go`, `todo_relate_test.go`. Inherited from the orchestrator's measurement; not independently re-run here.

---

## §3. Credential and disclosure precedent

`internal/glmcred/glmcred.go` is the model to follow, and its header states why the package exists at all: exactly one writer implementation, because two would mean two file-mode policies and two escaping rules. It depends only on the standard library and stdlib-only leaves so it cannot participate in an import cycle — which is what lets both `internal/cli` and `internal/web` import it while the one-way `cli → web` dependency stays acyclic.

Two details are load-bearing and easy to lose:

- **Mode tightening on write.** `os.WriteFile`'s perm argument applies only at creation, so an existing 0644 file stays 0644 without an explicit `Chmod`. The GLM package closes this; the Jev package inherits the same latent defect if it forgets.
- **The four-character disclosure floor.** `computeGLMKeyHint` returns `Configured: true` with an *empty* hint for a key of four characters or fewer, and the source notes that a naive "last four or the whole key" fallback would disclose a short key entirely — the exact inverse of the requirement.

`internal/web/glmkey.go` also records the structural guarantee: the credential is deliberately outside `settings.AllFields()`, so no generic schema-walking loop can read or render it, and a regression test (`TestGLMKeyField_AbsentFromSchema`) asserts the absence.

---

## §4. Configuration and surfaces

**`workflow.yaml` opt-in shape.** The template file carries three switches in the shape this SPEC needs — `codex.review_gate.enabled: false`, `multi.review_gate.enabled: false`, `slot_lease.enabled: false` — each with a comment stating it ships off and costs nothing while off. One switch ships *on* (`drift_cache_fill.enabled: true`) and its comment explicitly calls that an accepted cost rather than a neutral one, which is the pattern for justifying a non-inert default. Jev ships off, so it follows the first group.

The local `workflow.yaml` additionally carries `branch_guard.enabled: true` and `agent_stop_guard.enabled: true` — dogfood opt-ins whose template defaults are false. The same local-enables-what-template-ships-off pattern is available for Jev.

**Wizard.** `Question` (`internal/cli/wizard/types.go:78-92`) carries `ID`, `Type`, `Title`, `Description`, `Options`, `Default`, `Required`, `Condition`, `Group`. Translations map language code → question ID → `QuestionTranslation`, with all four locales present (`ko`, `ja`, `zh`, `en`). The constructor split matters: `InitQuestions` asks four (`conversation_language`, `user_name` from `DefaultQuestions`, plus `agent_wiring`, `autonomy_tier` from `Page3Questions`), and `ReconfigureQuestions` deliberately excludes the page-3 set to preserve its pre-restructure membership. Adding one question to `Page3Questions` makes init ask five and leaves reconfigure unable to change it — which is why §D1 of the plan flags the constructor choice as a decision rather than an assumption.

**Web console.** `internal/web/app.go:154-201` registers 11 routes: `/`, `/kanban`, `/monitor`, `/todo`, `/settings`, `/events`, `/save`, `/specs`, `/profile/create`, `/profile/delete`, plus the profile-rename and GLM-key-reveal paths and `/__shutdown__`. There is **no init route**, confirming the operator's out-of-scope call. All routes sit behind `hostCheckMiddleware`: a loopback Host check on every method and route, plus a `Sec-Fetch-Site` same-origin check on state-changing methods; there is no per-process CSRF token and the source says so deliberately.

**One persistence path.** `internal/web/projectconfig.go:194-214` and `:257-276` record the AP-2 rule: the load-modify-write seam was relocated to the neutral `internal/settings` package so the console and the TUI wizard share one writer. The nested-isolation crux is spelled out there — `SetSection` replaces the whole section struct and `Save()` serializes it, so the seam copies the entire struct and mutates only the targeted nested field, and each `*Set` flag gates a per-field mutation so an unsubmitted field keeps its persisted value.

**Doctor.** Checks are `DiagnosticCheck{Name: "..."}` values (`doctor_codex.go:163`, `doctor_disk.go:72`, `doctor_harness.go:21`, `doctor_hook_delivery.go:55`), addressable by `moai doctor --check "<Name>"`.

**Queue-hash method.** `internal/cli/todo_triage_test.go` imports `crypto/sha256` (line 21) and computes `sha256.Sum256(data)` (line 649) — the before/after comparison AC-JEV-001 reuses.

---

## §5. The inherited negative result

**Labelled as inherited, not re-measured.** The figures below originate in `.moai/reports/t943/verdict.md`, a gitignored local evidence file in the primary checkout that may be absent from any given tree. Nothing in this research pass re-ran the experiment, and no claim here asserts otherwise.

Using Jev to judge whether a stale card's premise is dead was measured on 124 Korean-original cards and rejected:

| Measure | Result |
|---|---|
| `premise_dead` call precision | 14/48 = **29.2%** against a **25%** base rate |
| 2-class accuracy | **58.9%** |
| Constant "always alive" answer | **75.0%** |
| English control (40 cards) | 3-class 60.0%, 2-class 67.5% — still below the constant |
| Threshold sweep | flat across 0.30-0.80; higher gates cut adoption without lifting accuracy |
| Errors vs confidence | 67 of 85 errors (79%) sat above 0.30, peaking at 0.93 |

Four things this record establishes, each of which shapes a requirement in this SPEC:

1. **The constant baseline is the number to beat** (REQ-JEV-027/028). A 16-point headline gain (31.5% → 47.6% on 3-class) looked like success while the constant sat at 75.0% untouched.
2. **A threshold sweep is not a rescue** (REQ-JEV-031). Flat across the whole range; the two high-gate cells that looked better rested on 5 and 4 samples.
3. **High confidence does not imply correctness** (REQ-JEV-031/039). The assumption that errors cluster at low confidence was measured false.
4. **The English arm is worth measuring and does not change a verdict by itself** (REQ-JEV-029/030). A ~10-point effect, on a 40-card subsample translated by the same model family — which is that control's own weak point.

A distinction the record also draws, and which this SPEC preserves: the same tool used as a **marker** rather than a judgment narrowed 124 cards to 48 to inspect, yielding 14 — a 2.6× narrowing. That usefulness exists only while "it only marks" holds. Read as a judgment it skips 34 live cards wrongly. This is the general shape of the display-only invariant, arrived at by measurement on one task before it was written as a rule for all three.

---

## §6. Known model weaknesses (vendor-stated, applied)

From the published jev-1.13 jaggedness notes, each mapped to the design response in `design.md` §5: literal reading (weak on negation, scoping words, implied conditions); unreliable counting, arithmetic, numeric proximity, and date ordering; weaker multi-hop indirection; accuracy degrading as irrelevant state grows; vulnerability to instructions injected in state; and `P(yes)` not guaranteed to equal `1 − P(no)`. Request limits: 64k tokens total, 32k for state plus the longest question.

These are vendor claims, not measurements taken here. They are used as *design constraints* — reasons to compute values in Go, keep state small, and add a no-match option — which is safe even if a given claim is conservative. No requirement in this SPEC asserts any of them as a measured property of this repository's data.

---

## §7. Open items

| # | Item | Why it is not settled here |
|---|---|---|
| R1 | Which wizard constructor the Jev question joins | `Page3Questions` makes init ask five and leaves reconfigure unable to change it; `DefaultQuestions` reaches reconfigure but changes the shared set. A product decision (see `plan.md` §D1). |
| R2 | Whether the pinned model id is compiled or operator-visible config | Compiled is simpler and harder to drift; operator-visible lets a pin move without a release. Affects the `workflow.jev` block's shape (`plan.md` §B2). |
| R3 | Whether a credential-reveal route is wanted | The GLM precedent has one; the disclosure requirement is satisfied without it. |
| R4 | Labelled-set size per consumer | The rejected task used 124 cards. The right size for each new consumer depends on its base rate, which is not yet known — M3 determines it. |
