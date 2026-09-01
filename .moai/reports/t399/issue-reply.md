# Draft reply to modu-ai/moai-adk#1632

> Status: DRAFT — not posted. The lead posts this.
> Card t399 · SPEC-CODEX-REVIEW-TARGET-001 · branch `WT-codex-native-branch`.

---

Thank you for this report, and in particular for pasting the raw JSON-RPC error text.
That one line made the diagnosis a single step instead of an investigation — we were
able to reproduce your exact error locally and confirm we were looking at the same
failure rather than a lookalike.

Below is what is now closed, what is still open and why, and what we re-measured of
your five items.

## 1. Closed: the native `baseBranch` request now satisfies the schema (observed live)

Your `mode=native, target=baseBranch` call never reached codex as a review at all.
The request builder lifted the bare target string into `{"type": "baseBranch"}`, and
the codex `ReviewStartParams` schema requires `branch` for that variant. codex rejected
the request; our fail-open contract then absorbed the rejection as `inconclusive` — the
same value a healthy review that found nothing returns. A rejected request and a passed
review were being reported identically.

The part worth stating plainly: **nobody had ever run that round trip.** Our own
specification recorded the failure as a prediction derived from reading the schema, and
our plan-phase audit caught the specification overstating that prediction as an
observation. So we ran it, against codex-cli 0.150.1:

```
--> {"id":3,"jsonrpc":"2.0","method":"review/start","params":{"target":{"type":"baseBranch"},"threadId":"..."}}
<-- {"error":{"code":-32600,"message":"Invalid request: missing field `branch`"},"id":3}
```

That is your error, reproduced. After the fix the same probe reaches `turn/started`
with a real turn id.

What changed:

- The target builder is now variant-aware. It populates each variant's required fields,
  and returns an error rather than an incomplete object — or a substituted one.
- A request whose target cannot be assembled is **not sent at all**, and the cause is
  named in the result rather than being folded into a generic `inconclusive`.
- The branch name is resolved **server-side**, because the `target` parameter is a
  string enum with no companion branch parameter — you could not have supplied one.
  The resolver follows the same chain the GLM backend already uses (remote default head,
  then `main`) and confirms the name resolves as a ref before returning it, so both
  backends review the same change. The `codex_audit` tool description now says this.

We deliberately did **not** add `commit` or `custom` to the caller-facing `target` enum.
They are covered only by the safety rule that a half-built target object is never
serialized; exposing them is a feature addition, not part of this fix.

## 2. Still open: "a required native gate silently yields nothing"

The symptom you described is not fully closed, and we would rather say so than imply it is.

The schema-violation cause is gone. But the same shape survives through other causes,
each of which still returns `inconclusive`:

- codex binary absent from PATH,
- codex present but erroring,
- and one case this fix introduces: a tree where no base branch resolves (previously the
  code would have quietly reviewed something else; it now declines and names the cause).

All three are annotated by `applyGateUnmet`, so the gap is *visible* — but the verdict
is still `inconclusive`, and `inconclusive` is not yet visibly distinct from "the backend
ran and agreed". Making non-participation distinguishable from agreement — exposing the
on-target backend count, and refusing `disagreement_flag=false` when fewer than two
backends actually participated — is tracked separately on our side and touches the
convergence layer (`mcp_convergence.go`), which this change does not modify.

## 3. Your item #3 (`gates.codex: required` enforcement) is a policy question, not a bug we closed

You asked that `required` either be enforced or renamed. We have not done either, and
this change does not decide it.

What exists today: `applyGateUnmet` annotates a fail-open `inconclusive` with the unmet
gate, so an unmet `required` gate is recorded rather than erased. It does not block.
Whether `required` should block is a genuine policy decision about the fail-open contract
itself, and we would rather take it as an explicit decision than smuggle it into a bug fix.
Your argument for renaming it if it will not block is a fair one and is on the table.

## 4. What we re-measured of your five items

To be precise about which of your items this work touched:

| Your item | State |
|---|---|
| #1 structured findings | Already landed before this card (`4fe2c54c0`) |
| #2 verdict synthesis | Already landed before this card (`410da655f`) |
| #3 `gates.codex: required` enforcement | **Open policy question** — see above |
| #4 legacy `auth.json` shape | Already landed before this card (`dbca3f710`) |
| #5 immediate-timeout record | **Not measured.** No one has attempted to reproduce it |

On #1, #2 and #4: those landed in earlier work, not here. If you are still seeing any of
them on a current build, that is a live report and we would want to hear it — we have not
re-verified them against your environment.

On #4 specifically: we have not confirmed whether your environment still reports `unknown`.

On #5: we want to be explicit rather than reassuring. We have not reproduced it, and we
have not ruled it out. All we know is where that string is written from
(`internal/cli/codex_task.go`). Saying "not measured" is not the same as saying "not a
bug"; if you can reproduce it we will take that as the starting point.

## 5. One constraint on the above

Our measurements are on codex-cli **0.150.1**; you reported on **0.149.0**. If 0.149.0
differs in the required-field set, that difference is invisible from here. The schema
excerpt in your report is consistent with what we measured, so we do not expect a
divergence — but we have not observed 0.149.0 directly.

Thanks again for the detail in the original report.
