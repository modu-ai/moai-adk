# SPEC-JEV-INTEGRATION-001 — Design

Design-level decisions and their alternatives. Implementation-level naming (function signatures, field names) is deliberately absent: it belongs to the run phase.

---

## §1. Layering

```
                      ┌──────────────────────────────────┐
  moai init wizard ──▶│  internal/settings  (one writer) │◀── moai web /settings
                      └───────────────┬──────────────────┘
                                      │  workflow.jev.enabled
                                      ▼
  consumer C ─┐            ┌─────────────────────┐
  consumer A ─┼───────────▶│     internal/jev    │──▶ POST /v1/systemone
  consumer B ─┤            │  gate · creds · HTTP│
  MCP tool   ─┘            │  bounds · screening │
                           │  accounting · pin   │
                           └─────────────────────┘
```

One implementation, four callers. The MCP tool is a caller, not a layer — it holds no transport code of its own, which is what keeps a second set of bounds, a second fail-open policy, and a second model pin from coming into existence.

The dependency direction is forced by existing structure: `internal/cli` imports `internal/web` one-way, so `internal/jev` and its credential sibling depend only on the standard library and on stdlib-only leaves (`internal/paths`, `internal/defs`), exactly as `internal/glmcred` does. Any richer dependency would close a cycle.

---

## §2. Why the gate sits below the consumers

The config gate could live at each call site or inside the package. It lives inside the package, checked before request construction, for one reason: REQ-JEV-017 requires that a disabled capability construct no request at all. A gate at the call sites is N places to forget; a gate in the package is one place, and the "no request constructed" property is then testable once rather than N times.

The cost is that a caller cannot tell "disabled" from "unavailable" without reading the typed result. That is why `Availability` names the condition rather than being a boolean — the doctor check (REQ-JEV-026) needs exactly that distinction.

---

## §3. Why an unavailable result is a value

An error return would propagate, and a propagated error becomes a non-zero exit somewhere — which is precisely the failure REQ-JEV-007 forbids. Returning a value makes the degradation the default path rather than the exceptional one: a caller that ignores availability gets "no answer", and "no answer" is already a state every consumer must handle, because a disabled capability produces it.

The symmetry matters. Absence of a Jev answer is not evidence of anything — not that the pair is unrelated, not that the question is lead-owned, not that no skill is needed. Every consumer treats an unavailable result as "no signal", never as a negative signal.

---

## §4. The third finding source

### The problem

`BacklogFinding.Source` is a free string with two constants. They partition the finding space by *who observed it*: `mechanical` means the text analyser measured it, `agent` means a reader who understands what the cards mean judged it. `HasAgentFindingForPair` selects on `agent` and drives the `machine-only` mark, whose documented meaning is "nothing agent-sourced was recorded here" — an honest claim precisely because the CLI cannot know who called it.

A Jev answer is neither. It is not a measurement of text similarity, and it is not a reader who understands the cards. Filing it under either existing constant makes the queue assert something false: under `mechanical` it would be rendered with a `score` that reads as a measured similarity; under `agent` it would erase the `machine-only` mark from pairs no one reviewed, claiming a review that did not happen.

### The decision

A third constant, `jev`. Three consequences, each a design choice rather than a mechanical follow-on:

**Precedence.** A Jev finding is appended only when no finding of any source already names that unordered pair with the same relation; a later mechanical or agent finding is appended alongside it rather than replacing it. The asymmetry is deliberate: a model signal must never suppress a measurement or a judgement, and must never be silently suppressed by one, because both suppressions would make the queue quieter than the evidence warrants.

**Render.** The existing rule prints a score only for `mechanical`, on the stated reasoning that an agent judgement carries no measurement and `0.00` would read as measured dissimilarity. A Jev probability is a third thing again — a calibrated model confidence. It renders with its own label, so the three are distinguishable at a glance.

**Mark semantics.** The `machine-only` mark's meaning is unchanged. It continues to mean "no agent-sourced record for this pair", and a Jev finding does not satisfy it. AC-JEV-034 pins this.

### Alternative rejected

Reusing `agent` with a note saying "written by Jev" was considered and rejected: the note is prose, `HasAgentFindingForPair` reads the constant, and no reader of the predicate would see the note. A distinction that only exists in a free-text field is not a distinction the code makes.

---

## §5. Question design against the model's known jaggedness

The published weaknesses shape the design rather than being mitigated afterwards.

| Weakness | Design response |
|---|---|
| Literal reading; weak on negation, scoping words, implied conditions | Questions make conditions explicit; no question relies on the model inferring a scope from context |
| Unreliable at counting, arithmetic, numeric proximity, date ordering | Go computes every such value; it reaches the model as a named JSON field |
| Weaker on multi-hop indirection | Indirection is resolved in Go before the question is asked |
| Accuracy falls as irrelevant state grows | A request's state carries only the fields its questions read |
| Vulnerable to instructions injected in state | State is untrusted data; no code path treats its text as instruction |
| `P(yes) ≠ 1 − P(no)` | Both are read where both matter; neither is derived from the other |

The batching rule (one state, N questions, one request) is both an economy and a correctness property: N separate requests over the same state would each pay the state's token cost and could return mutually inconsistent judgments over identical input.

The no-match option on every Choice exists because a forced choice over an option set that excludes the true answer produces a confident wrong answer — a failure mode this repository has recorded from its own dispatch practice, independently of any model.

---

## §6. The measurement gate

Each consumer's gate has the same shape and different contents:

1. Draw a labelled set from this repository's own data.
2. Compute the **constant-answer baseline** — the accuracy of always answering the majority label. This is the number to beat, and it is often high: a constant answer scored 75.0% on the rejected premise-death task, which is why that task's 58.9% model accuracy was a rejection rather than a modest result.
3. Run both language arms, Korean original and English translation, and record the delta.
4. Fit the threshold on the measured distribution.
5. Cite the pinned model id.

The gate is binding. A consumer that does not beat its baseline is not shipped, and its absence is recorded as a decision so the next reader does not rebuild it.

**Why the baseline rather than raw accuracy.** Raw accuracy is unreadable without the base rate. The rejected task's headline moved from 31.5% to 47.6% under a measurement-design repair — a 16-point gain that looks like success and was not, because the constant baseline sat at 75.0% the whole time. The number that changes a decision is the margin over the constant, and only the constant makes it visible.

**Why two language arms.** The model card notes reduced non-English and CJK accuracy, and this repository's cards are Korean. Measuring both arms records the size of that effect (previously measured at roughly 10 points on a 40-card subsample, not the ~50 points an earlier 16-card comparison suggested) so a later card can decide whether English cards are worth their cost. This SPEC measures and does not decide — REQ-JEV-030.

---

## §7. Opt-in surfaces

Two entrances, one writer. The console and the wizard both reach the neutral `internal/settings` seam, following the rule already recorded in the console's own source: the load-modify-write body was relocated there specifically so no parallel writer exists.

The nested-isolation property carries over unchanged: section writes replace the whole section struct on save, so the seam copies the entire struct and mutates only the targeted field. Every sibling field in `workflow.yaml` — and that file is large — rides through byte-identical. AC-JEV-019 pins it.

The credential stays outside the settings schema entirely, on the precedent the GLM key set: absence from the schema field set is a *structural* guarantee that no generic schema-walking loop (bulk value read, form-state dump, diagnostics view) can pick it up. A regression test asserts the absence, because the guarantee is only as good as the thing that notices when it breaks.

The four-character disclosure floor is inherited verbatim and is not an arbitrary rounding: a naive "last four, or the whole key if shorter" fallback discloses a short key entirely, which is the exact inverse of the requirement.

---

## §8. `/moai goal --auto` seats

**Seat (i) — routing inside the sealed loop.** The loop's existing contract is that after the single approval it asks no further user questions, and that anything exceeding sealed scope becomes a persisted blocked result. Consumer A's routing does not change either property; it changes which side of that line an item lands on. Lead-owned and cheap-to-reverse items were already the lead's to answer — the routing only classifies them. Operator-owned items reach the same persisted blocked result they reach today. An operator answer or a below-threshold confidence both route to operator: uncertainty escalates, never downgrades.

**Seat (ii) — auxiliary input to the governor.** The governor reads a sealed snapshot and returns one bounded decision object; a deterministic executor validates it and performs any state change. A Jev Noul is one more piece of evidence in the snapshot, recorded as a separate item in the governance receipt so that a reader can see what the governor was shown. Three properties stay untouched: the governor's read-only scope, its output shape, and the receipt's existing binding fields. And the hard line — REQ-JEV-059 — is that a Jev answer never becomes an element of a completion predicate, nor of landed-ancestry or authoritative-readback evidence. Those are the irreversible judgments, and irreversible judgments do not take model answers even as inputs.

---

## §9. Disposition of `scripts/jev/`

The scripts exist only as an uncommitted working copy in the primary checkout; they are absent from `develop` and have no template mirror, so they reach no user project. The Go package supersedes them: after M1 the scripts' behaviour is available through the binary, and after M8 through the MCP tool as well. The recommended disposition is therefore **supersede without porting** — the scripts are not committed, not distributed, and not maintained; the Go package is the canonical implementation and the only one this SPEC's acceptance criteria bind. Nothing in this SPEC deletes anything from anyone's working tree.

---

## §10. Cost shape

Input tokens are charged and output tokens are not, so cost tracks state size rather than answer count — which is the economic reason batching several questions over one state is the right shape, independent of the correctness reason in §5. The per-call record of input-token count and pinned model id (REQ-JEV-004) is what makes a later cost question answerable from the record.

No price figure appears in this SPEC's template-bound artifacts, per the neutrality constraint; the figures belong in evidence files.
