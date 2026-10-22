# SPEC-JEV-CORE-001 — Implementation Plan

> Sections are ordered by decision-reversibility: §B (data model and config shape) and §C (interfaces) carry the choices most likely to change on review; the mechanical work is last.

---

## §A. Context

First SPEC of a four-part chain split from `SPEC-JEV-INTEGRATION-001` (card t1020). Ships the Jev capability with no consumers.

Verified tree facts (measured in this worktree at `fd75cf692`):

| Fact | Command | Observed |
|---|---|---|
| No Go reference to the model today | `grep -ril "typesafe\|jev" internal/ pkg/ cmd/` | 0 files (positive control: `grep -ril glmcred` → 5 files) |
| `scripts/jev/` absent on develop | `ls scripts/` | no `jev` entry among 16 |
| Opt-in switch shape | template `workflow.yaml` | `codex.review_gate.enabled: false`, `multi.review_gate.enabled: false`, `slot_lease.enabled: false` |
| Queue-hash method | `internal/cli/todo_triage_test.go:21,649` | `crypto/sha256` + `sha256.Sum256(data)` |
| Doctor check shape | `doctor_codex.go:163`, `doctor_disk.go:72` | `DiagnosticCheck{Name: "..."}` |

## §B. Data-model and config decisions (review these first)

**B1 — Config key shape.** `workflow.jev.enabled` inside the existing `workflow.yaml`, following the three opt-in switches already there: shipped `false`, compiled default in `internal/config/defaults.go`, template block documenting it. An alternative shape (`workflow.jev: {enabled, model, …}`) is available if the pinned model id should be operator-visible rather than compiled — that is open question Q2, not a settled decision.

**B2 — The credential is out of the settings schema.** Modelled on `internal/glmcred` and `internal/web/glmkey.go`: a dedicated package, mode 0600 with the existing wider-mode tightening, a hand-built parse/validate/view path, and absence from `settings.AllFields()` enforced by a regression test in the shape of `TestGLMKeyField_AbsentFromSchema`. The four-character disclosure floor is inherited verbatim: a key of four characters or fewer discloses nothing but `Configured`.

**B3 — Package boundary.** `internal/jev` depends only on the standard library and stdlib-only internal leaf packages (`internal/paths`, `internal/defs`), exactly as `internal/glmcred` does, so it can be imported by both `internal/cli` and `internal/web` without creating a cycle. The credential reader lives in a sibling `internal/jevcred`, closer to the existing precedent than folding it into `internal/jev`.

## §C. Interface decisions

**C1 — The unavailable result is a value, not an error.** The proposed surface returns `(Answer, Availability)` where `Availability` names the observed condition (disabled / no-credential / unauthorized / rate-limited / overloaded / unreachable / oversize). A caller that ignores `Availability` degrades to "no answer", never to a failure. An error return would propagate, and a propagated error becomes a non-zero exit somewhere — the failure REQ-JEVC-007 forbids.

**C2 — Batching is in the type.** A request carries one state and a *list* of questions. A single-question convenience wrapper is fine; a path that sends two requests over one state is not.

**C3 — Size refusal precedes transport.** The 32k and 64k bounds are checked before the HTTP call and refuse rather than truncate, so an oversize state is a visible condition rather than a silently shortened one.

**C4 — The gate sits below the callers.** Checked inside the package, before request construction, because REQ-JEVC-017 requires that a disabled capability construct no request at all. A gate at the call sites is N places to forget. The cost is that a caller cannot tell "disabled" from "unavailable" without reading the typed result — which is why `Availability` names the condition rather than being a boolean, since the doctor check needs exactly that distinction.

## §D. Constraints

- **Display-only** is the load-bearing invariant, verified against a queue-file hash taken before and after.
- **Default off.** Template default `false`, compiled default in `defaults.go`.
- **Fail-open.** No absence becomes a failure; retry only where repetition is provably side-effect-free.
- **Pinned model id.** `jev-latest` appears in no request.
- **Template-First.** The `workflow.yaml` template block lands in `internal/template/templates/` first; template content carries no SPEC ID, card id, date, price, or measurement figure.
- **Verification scope.** Affected packages only; full-suite verdict from CI. No full local `go test ./...`.

## §E. Self-Verification

Report per the 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). The AC matrix in `acceptance.md` is the E1 subject.

## §F. Milestones

**M1a — Package skeleton.** Request/response types, `Availability`, pinned model id, size bounds, batching shape. No transport yet.

**M1b — Credential.** `internal/jevcred`: path resolution through `internal/paths`, 0600 write with wider-mode tightening, read, the four-character disclosure helper, absence from `settings.AllFields()` plus its regression test.

**M1c — Config gate.** `workflow.jev.enabled` in `defaults.go`, local `workflow.yaml`, template block.

**M1d — Transport, screening, accounting.** HTTP call, secret screening ahead of send, fail-open mapping of 401/429/529/unreachable, input-token and model-id recording, retry policy.

**M1e — Doctor check.** `DiagnosticCheck{Name: "Jev"}` reporting enabled-state, credential presence, and reachability without a judgment request.

## §G. Anti-patterns

- **Letting an absence become an error return.** The whole fail-open contract turns on the unavailable result being a value.
- **Citing `jev-latest`.** A request or measurement under a moving alias is unattributable.
- **A second client.** Any transport code outside `internal/jev` defeats REQ-JEVC-002 and doubles the bounds, the fail-open policy, and the model pin.
- **Truncating an oversize state.** Refusal is visible; truncation is a silently different question.
- **Shipping a consumer in this SPEC.** The capability ships unwired on purpose, so the display-only invariant is testable before anything can violate it.

## §H. Cross-references

- `SPEC-JEV-OPTIN-MEASURE-001` — the successor; owns the wizard and web surfaces and the measurement harness.
- `internal/glmcred/glmcred.go`, `internal/web/glmkey.go` — the credential and disclosure precedent.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the evidence format §E reports in.
