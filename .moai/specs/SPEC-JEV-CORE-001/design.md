# SPEC-JEV-CORE-001 — Design

Design-level decisions and their alternatives. Implementation-level naming is deliberately absent: it belongs to the run phase.

---

## §1. Layering

```
  (no consumers in this SPEC)
            │
            ▼
   ┌─────────────────────┐
   │     internal/jev    │──▶ POST /v1/systemone
   │  gate · bounds      │
   │  screening · pin    │
   │  accounting         │
   └──────────┬──────────┘
              │
   ┌──────────▼──────────┐
   │  internal/jevcred   │──▶ ~/.moai/.env.typesafe (0600)
   └─────────────────────┘
```

One implementation, zero callers. The zero is the point: every later SPEC adds a caller, and each addition is a place the display-only invariant could be violated. Shipping the boundary first means the invariant has a test before it has anything to defend against.

The dependency direction is forced by existing structure. `internal/cli` imports `internal/web` one-way, so both packages here depend only on the standard library and on stdlib-only leaves (`internal/paths`, `internal/defs`), exactly as `internal/glmcred` does. A richer dependency would close a cycle.

## §2. Why the gate sits below the callers

The config gate could live at each call site or inside the package. It lives inside the package, checked before request construction, because REQ-JEVC-017 requires a disabled capability to construct no request at all. A gate at the call sites is N places to forget; a gate in the package is one place, and the "no request constructed" property is then testable once rather than N times.

The cost is that a caller cannot tell "disabled" from "unavailable" without reading the typed result. That is why `Availability` names the condition rather than being a boolean — the doctor check (REQ-JEVC-022) needs exactly that distinction, and it is the only consumer in this SPEC.

## §3. Why an unavailable result is a value

An error return would propagate, and a propagated error becomes a non-zero exit somewhere — precisely the failure REQ-JEVC-007 forbids. Returning a value makes the degradation the default path rather than the exceptional one: a caller that ignores availability gets "no answer", and "no answer" is already a state every future consumer must handle, because a disabled capability produces it.

The symmetry matters, and it is the clause every later SPEC will lean on. Absence of a Jev answer is not evidence of anything — not that a pair is unrelated, not that a question is lead-owned, not that no skill is needed. Every consumer treats an unavailable result as "no signal", never as a negative signal.

## §4. The display-only boundary, shipped with the capability

REQ-JEVC-011 through REQ-JEVC-014 could plausibly have been deferred to the SPEC that adds the first consumer. They are here instead, for a reason that is about ordering rather than tidiness: a boundary written after the thing it bounds is written against code that already exists, and tends to describe that code rather than constrain it. Written first, with no consumer in view, it constrains every consumer equally.

REQ-JEVC-012's "not as an input" is the load-bearing half. A rule forbidding Jev from *deciding* an irreversible judgment is easy to satisfy and easy to route around: the model answers, a human reads the answer, the human decides, and the model's confidence is in the verdict's reasoning with nothing recording how it got there. Forbidding it as an input closes that route, and makes the acceptance criterion a reachability question a test can answer.

## §5. Credential handling

`internal/glmcred` is the model, and its header states why the package exists at all: exactly one writer implementation, because two would mean two file-mode policies and two escaping rules.

Two details are load-bearing and easy to lose:

- **Mode tightening on write.** `os.WriteFile`'s perm argument applies only at creation, so an existing 0644 file stays 0644 without an explicit `Chmod`. The GLM package closes this; a new package inherits the same latent defect if it forgets.
- **The four-character disclosure floor.** The GLM view helper returns `Configured: true` with an *empty* hint for a key of four characters or fewer, and its source notes that a naive "last four or the whole key" fallback would disclose a short key entirely — the exact inverse of the requirement.

The credential stays outside `settings.AllFields()` as a *structural* guarantee: no generic schema-walking loop (bulk value read, form-state dump, diagnostics view) can pick it up. A regression test asserts the absence, because the guarantee is only as good as the thing that notices when it breaks.

## §6. Request shape under the model's known jaggedness

The published weaknesses shape the package's contract even though no question is authored here. The two that bind at this layer:

- **Accuracy falls as irrelevant state grows**, and **the model is vulnerable to instructions injected in state.** Both make the state payload a thing to bound and to distrust. The 32k/64k limits are enforced here, and refusal rather than truncation is what keeps an oversize state a visible condition instead of a silently different question.
- **Batching is both an economy and a correctness property.** Input tokens are charged and output tokens are not, so cost tracks state size rather than answer count. Independently of cost, N separate requests over the same state could return mutually inconsistent judgments over identical input. Putting the question *list* in the type makes the single-request shape the default rather than a discipline.

The remaining weaknesses — literal reading, unreliable counting and date ordering, multi-hop indirection, and `P(yes) ≠ 1 − P(no)` — bind at the question-authoring layer and are owned by `SPEC-JEV-OPTIN-MEASURE-001`.

## §7. Cost shape

Input tokens are charged; output tokens are not. The per-call record of input-token count and pinned model id (REQ-JEVC-004) is what makes a later cost or provenance question answerable from the record rather than from recollection. No price figure appears in this SPEC's template-bound artifacts, per the neutrality constraint; figures belong in evidence files.
