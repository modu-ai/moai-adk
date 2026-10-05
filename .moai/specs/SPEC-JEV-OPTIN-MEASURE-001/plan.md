# SPEC-JEV-OPTIN-MEASURE-001 — Implementation Plan

> Sections are ordered by decision-reversibility: §B (user-facing flow — the init-only placement and its consequence) and §C (measurement design) carry the choices most likely to change on review; mechanical work is last.

---

## §A. Context

Second SPEC of a four-part chain split from `SPEC-JEV-INTEGRATION-001` (card t1020). Depends on `SPEC-JEV-CORE-001` for the package, the `workflow.jev.enabled` key, and the credential reader.

Verified tree facts (measured in this worktree at `fd75cf692`):

| Fact | Source | Observed |
|---|---|---|
| Wizard question shape | `internal/cli/wizard/types.go:78-99` | `Question{ID, Type, Title, Description, Options, Default, Required, Condition, Group}` |
| Init set size today | `internal/cli/wizard/questions.go:290-322` | four: `conversation_language`, `user_name` (from `DefaultQuestions`) + `agent_wiring`, `autonomy_tier` (from `Page3Questions`) |
| Reconfigure excludes page 3 | `questions.go:269-288` | deliberate, to preserve pre-restructure membership |
| Locales present | `translations.go` | `ko`, `ja`, `zh`, `en` |
| Console routes | `internal/web/app.go:154-201` | 11 routes, **no init route** |
| One-writer rule | `internal/web/projectconfig.go:194-214, 257-276` | AP-2: the load-modify-write seam lives in `internal/settings` so console and wizard share one writer |

## §B. User-facing flow decisions (review these first)

**B1 — The question is init-only, and that has a cost the SPEC must pay for.** Operator decision: the Jev question joins `InitQuestions`, bringing it to five. It does not join `DefaultQuestions` and does not appear in `ReconfigureQuestions`.

The consequence is not incidental. `ReconfigureQuestions` deliberately excludes the page-3 set to preserve its pre-restructure membership, so a question placed there is genuinely unreachable from `moai update --reconfigure`. A user who initializes from the terminal, declines Jev, and never opens the console has no path back to the switch unless we build one — and the thing we build is a *sentence*, not a route: REQ-JEVO-007 requires the wizard text or the init completion output to name `moai web`.

That is the reversible part of this decision and the part most worth reviewing: which of the two surfaces carries the pointer, and in what words, in four locales. The placement itself is settled.

**B2 — One web section, no new route.** The Jev section renders inside the existing `/settings` screen and persists through the existing `/save` handler. The console has no init route and gains none. A key-reveal route is *not* proposed: the GLM precedent has one (`glmKeyRevealPath`), but reveal is a separable decision (open question Q4, owned by `SPEC-JEV-CORE-001`) and the disclosure requirement is satisfied without it.

**B3 — One persistence path.** Both entrances call the neutral `internal/settings` seam, per the AP-2 rule recorded in the console's own source. The nested-isolation crux applies: `SetSection` replaces the whole section struct and `Save()` serializes it, so the seam copies the entire struct and mutates only the targeted field, and every sibling field in `workflow.yaml` — a large file — rides through byte-identical.

**B4 — Privacy disclosure at the point of choice.** Both the wizard question's `Description` and the web section's body state that enabling sends card text or request text to a third-party server. A string in four locales, not a link.

## §C. Measurement design decisions

**C1 — The baseline is the constant answer, not raw accuracy.** Each consumer's gate computes the accuracy of always answering the majority label, and that is the number to beat. Raw accuracy is unreadable without the base rate, and the repository has already been misled by one: on the rejected premise-death task the headline moved 31.5% → 47.6% under a measurement-design repair, a sixteen-point gain that looked like success while the constant sat at 75.0%, unbeaten. The margin over the constant is the only figure that changes a decision.

**C2 — Two language arms, decided by neither.** Korean original and English translation, with the delta recorded. The model card notes reduced non-English and CJK accuracy and this repository's cards are Korean, so the effect is real — previously measured at roughly ten points on a 40-card subsample, not the ~50 points an earlier 16-card comparison suggested. This SPEC measures and does not decide; REQ-JEVO-011 forbids it changing any card.

**C3 — Thresholds are fitted, per consumer, per question shape.** Not copied from vendor documentation, and not transferred between a Noul and a Choice. The rejected task's gate sweep was flat across 0.30-0.80, and the two high-gate cells that looked better rested on 5 and 4 samples — which is why a fitted threshold needs its sample size recorded alongside it.

**C4 — Every absence carries a positive control.** A zero-hit and a broken measuring apparatus are indistinguishable on their own, and the zero-direction error is the one that does not prompt a re-measurement.

## §D. Constraints

- **Init-only placement is settled**; the pointer sentence is the reviewable part.
- **One writer.** No parallel persistence path, and sibling config fields byte-identical after a write.
- **Four locales** for every user-facing string.
- **Measurement is binding.** A consumer that does not beat its baseline is not shipped, and its absence is recorded as a decision.
- **Pinned model id** in every measurement citation; `jev-latest` nowhere.
- **No card change.** Schema, issuance path, and card language all untouched.
- **Template-First.** Every `.claude/` or `.moai/` file lands in `internal/template/templates/` first; template content carries no SPEC ID, card id, date, price, or measurement figure.
- **Verification scope.** Affected packages only; full-suite verdict from CI.

## §E. Self-Verification

Report per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). The AC matrix in `acceptance.md` is the E1 subject.

## §F. Milestones

**M2a — Wizard question.** `Question` value with the Jev id joined to `InitQuestions` (five questions); `QuestionTranslation` entries in all four locales; the `moai web` pointer in the question text or the init completion output.

**M2b — Web section.** Jev section on `/settings` with enable toggle and credential field; privacy disclosure in the section body, four locales; no new route.

**M2c — Shared seam.** The `internal/settings` load-modify-write path both entrances call; the byte-identity test for sibling fields; the reconfigure-absence test.

**M3a — Harness skeleton.** Labelled-answer-set runner, two language arms, per-arm result and delta recording, pinned-model-id citation.

**M3b — Baseline and thresholds.** Constant-answer baseline computation per question shape; threshold fitting with sample size recorded; the positive-control requirement for absence-shaped results.

**M3c — Gate demonstration.** Run the harness end to end for one trial consumer so the gate is demonstrated before `SPEC-JEV-CONSUMERS-001` depends on it.

## §G. Anti-patterns

- **Reporting raw accuracy without its constant baseline.** The number becomes unreadable and has already misled once.
- **Copying a vendor threshold.** `0.6` and `0.85` are documentation examples; an unfitted threshold is an unmeasured claim.
- **Transferring a threshold between Noul and Choice.** Different question shapes, different calibration.
- **Reading a zero-hit as an absence.** Without a positive control it is a gap, not a finding.
- **Adding the question to `DefaultQuestions` "so reconfigure can reach it".** That reverses the operator decision; the escape hatch is the `moai web` pointer, not a second placement.
- **Building an init flow in the console.** No route exists and none is in scope.
- **Treating a shipped-but-unmeasured consumer as gated.** The gate is the committed measurement, not the intention to take one.

## §H. Cross-references

- `SPEC-JEV-CORE-001` — predecessor; owns the package, config key, credential, fail-open, and doctor check.
- `SPEC-JEV-CONSUMERS-001` — successor; each of its three consumers passes the gate this SPEC builds.
- `internal/web/projectconfig.go` §AP-2 — the one-persistence-path rule B3 follows.
- `internal/cli/wizard/questions.go:269-322` — the constructor split B1 depends on.
