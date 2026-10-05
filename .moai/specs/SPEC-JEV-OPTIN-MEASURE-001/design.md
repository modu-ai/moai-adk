# SPEC-JEV-OPTIN-MEASURE-001 — Design

---

## §1. Two entrances, one writer

```
  moai init wizard ──┐
                     ├──▶ internal/settings (one seam) ──▶ workflow.yaml
  moai web /settings ┘
```

The rule is already recorded in the console's own source: the load-modify-write body was relocated to the neutral `internal/settings` package specifically so no parallel writer exists. Two writers would mean two escaping rules, two nested-isolation implementations, and two places to get the byte-identity property wrong.

The nested-isolation property carries over unchanged. Section writes replace the whole section struct on save, so the seam copies the entire struct and mutates only the targeted field. `workflow.yaml` is large — model-routing profiles alone run to 36 entries — and every one of those fields rides through byte-identical or the write is a defect. AC-JEVO-003 pins it.

## §2. Init-only, and the sentence that pays for it

The operator chose the init-only set. The design consequence is sharper than it looks, because `ReconfigureQuestions` deliberately excludes the page-3 set to preserve its pre-restructure membership — so this is not "hard to reach from reconfigure", it is *unreachable* from reconfigure by construction.

That leaves exactly one post-init path: the console. And a console-only escape hatch has an obvious failure mode — a user who works entirely from the terminal never learns it exists. The capability would then be, for that user, permanently whatever they answered once during `moai init`, with no visible way back.

The fix is deliberately the cheapest one available: a sentence. REQ-JEVO-007 requires the wizard question text or the init completion output to name `moai web`. No second route, no reconfigure entry, no new command — the placement decision stands and the discoverability cost is paid in four locales of prose.

This is the part of the design most worth arguing with on review. Which surface carries the pointer (question `Description` versus completion output) changes who sees it: the `Description` reaches a user deciding, the completion output reaches a user who has already decided. The second is arguably better — a user who declines is exactly the user who later wants the switch back — and the requirement permits either, so the choice is open at implementation time rather than settled here.

## §3. Why the credential stays outside the schema

Inherited from the GLM precedent and restated because the web surface is where it would break: the credential is deliberately outside `settings.AllFields()`, so no generic schema-walking loop — bulk value read, form-state dump, diagnostics view — can pick it up and render it. The guarantee is structural rather than conventional, and a regression test asserts it, because a structural guarantee is only as good as the thing that notices when it stops holding.

The four-character disclosure floor is not an arbitrary rounding. A naive "last four, or the whole key if it is shorter" fallback discloses a short key entirely, which is the exact inverse of the requirement.

## §4. The measurement gate

Each consumer's gate has the same shape and different contents:

1. Draw a labelled set from this repository's own data.
2. Compute the **constant-answer baseline** — the accuracy of always answering the majority label. This is the number to beat, and it is often high.
3. Run both language arms, Korean original and English translation, and record the delta.
4. Fit the threshold on the measured distribution, recording its sample size.
5. Cite the pinned model id.

**Why the baseline rather than raw accuracy.** Raw accuracy is unreadable without the base rate. On the rejected premise-death task the headline moved from 31.5% to 47.6% under a measurement-design repair — a sixteen-point gain that reads as success and was not, because the constant baseline sat at 75.0% the whole time. The number that changes a decision is the margin over the constant, and only the constant makes it visible.

**Why a threshold sweep is not a rescue.** On that same task the sweep was flat across 0.30-0.80; raising the gate cut adoption without lifting accuracy. Two high-gate cells did look better, at 60% and 75% — on 5 and 4 samples respectively, where a single flip moves 60% to 40%. Any fitted threshold therefore carries its sample size, or it is a number with no error bar presented as a decision rule.

**Why high confidence is not a proxy for correctness.** The assumption that errors cluster at low confidence was measured false on that task: 67 of 85 errors sat above 0.30, peaking at 0.93. A threshold is a cost/coverage lever, not a correctness filter.

**Why two language arms, deciding nothing.** The model card notes reduced non-English and CJK accuracy, and this repository's cards are Korean. Measuring both arms records the size of that effect — roughly ten points on a 40-card subsample, not the ~50 points an earlier 16-card comparison suggested — so a later card can decide whether English cards are worth their cost. That subsample was also translated by the same model family, which is its own weak point and is recorded as such.

## §5. Question design against the model's known jaggedness

The published weaknesses shape the questions rather than being mitigated afterwards.

| Weakness | Design response |
|---|---|
| Literal reading; weak on negation, scoping words, implied conditions | Questions make conditions explicit; no question relies on the model inferring a scope from context |
| Unreliable at counting, arithmetic, numeric proximity, date ordering | Go computes every such value; it reaches the model as a named JSON field |
| Weaker on multi-hop indirection | Indirection is resolved in Go before the question is asked |
| Accuracy falls as irrelevant state grows | A request's state carries only the fields its questions read |
| Vulnerable to instructions injected in state | State is untrusted data; no code path treats its text as instruction |
| `P(yes) ≠ 1 − P(no)` | Both are read where both matter; neither is derived from the other |

The no-match option on every Choice exists because a forced choice over an option set that excludes the true answer produces a confident wrong answer. That failure mode is not model-specific — this repository has recorded it from its own dispatch practice, where a question offering two branches got a confident answer and the true answer was a third thing.

## §6. What this SPEC deliberately does not do

It asks the model no question in production. The harness asks questions against a labelled set; nothing here wires a question into a workflow. That separation is what makes the gate meaningful: if measurement and consumption shipped together, a consumer's measurement would be taken on the same commit that made the consumer reachable, and a failing measurement would arrive as a revert rather than as a decision not to ship.
