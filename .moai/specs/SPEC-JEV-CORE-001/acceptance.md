# SPEC-JEV-CORE-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here.

## §A. Display-only invariant

**AC-JEVC-001** — Given the backlog queue file with a recorded SHA-256, When any code path in this SPEC's scope runs to completion with the capability enabled and a credential present, Then the queue file's SHA-256 is unchanged. Method: the `sha256.Sum256` before/after comparison already used in `internal/cli/todo_triage_test.go`.

**AC-JEVC-002** — Given any card, When the package returns any answer, Then no field of any `BacklogItem` differs from its pre-call value.

**AC-JEVC-003** — Given the capability enabled, When a completion verdict, a merge approval, a `moai todo` or `moai gtd` mutation, an operator gate, a user-surface behaviour change, or a CodeRabbit slot-wait adjudication is reached, Then no call to `internal/jev` occurs. Method: a guard test asserting the call path is unreachable from those code paths, plus a grep over those packages for the client symbol with a positive control on a path that does call it.

**AC-JEVC-004** — Given output containing a Jev answer, When a reader inspects it, Then the answer is labelled as a model-produced signal and is textually distinguishable from both a mechanical measurement and an agent judgement.

## §B. Default-off and fail-open

**AC-JEVC-005** — Given a freshly initialized project from the shipped template, When `workflow.jev.enabled` is read, Then it is `false`, and the compiled default in `internal/config/defaults.go` agrees.

**AC-JEVC-006** — Given `workflow.jev.enabled: false`, When each affected command runs, Then its stdout and stderr are byte-identical to the pre-SPEC baseline and no HTTP request to the TypeSafe host is constructed. Method: a transport stub asserting zero calls, plus a golden-output comparison.

**AC-JEVC-007** — Given the capability enabled and `~/.moai/.env.typesafe` absent, When each affected command runs, Then it exits 0 and emits at most one notice line.

**AC-JEVC-008** — Given the capability enabled, When the transport returns 401, 429, or 529, or the host is unreachable, Then the command exits 0, emits at most one notice line, and the typed unavailable result names the observed condition. One sub-case per condition.

**AC-JEVC-009** — Given a call that failed ambiguously, When the retry policy is exercised, Then no retry is issued for a non-idempotent call, and at most the configured number for an idempotent one.

## §C. Credential handling

**AC-JEVC-010** — Given a fresh credential write, When the file mode of `~/.moai/.env.typesafe` is read, Then it is 0600.

**AC-JEVC-011** — Given a pre-existing credential file at mode 0644, When a new credential is written, Then the resulting mode is 0600.

**AC-JEVC-012** — Given `settings.AllFields()`, When it is enumerated, Then no entry names the Jev credential. Method: a regression test in the shape of `TestGLMKeyField_AbsentFromSchema`.

**AC-JEVC-013** — Given a stored credential longer than four characters, When the view model is computed, Then `Configured` is true and the hint is exactly the final four characters.

**AC-JEVC-014** — Given a stored credential of four characters or fewer, When the view model is computed, Then `Configured` is true and the hint is empty.

**AC-JEVC-015** — Given a request payload containing a credential-shaped token, When the send is attempted, Then the client refuses to send and reports the refusal. A negative control asserts an ordinary payload sends.

## §D. Request shape and readiness

**AC-JEVC-016** — Given `moai doctor`, When the Jev check runs, Then it reports enabled-state, credential presence, and endpoint reachability, and sends no judgment request. Method: a transport stub asserting zero judgment calls.

---

## §E. Quality gates

- `go test ./internal/jev/... ./internal/jevcred/... ./internal/config/...` passes; full-suite verdict from CI.
- `golangci-lint run` clean on changed packages.
- `make build` succeeds.
- Coverage on `internal/jev` at or above the project's 85% package-level target.

## §F. Definition of Done

1. Every AC above passes.
2. The default-off byte-identity check (AC-JEVC-006) passes.
3. The credential anti-leak check (AC-JEVC-012) passes.
4. No consumer of the package exists in this SPEC's diff — the capability ships unwired.
