# SPEC-JEV-CORE-001 — Research

Read-only findings from this worktree at `fd75cf692`. Every claim names how it was observed; anything inherited rather than measured is labelled as such.

---

## §1. What exists today

| Question | Command run | Observed |
|---|---|---|
| Does any Go code reach TypeSafe? | `grep -ril "typesafe\|jev" internal/ pkg/ cmd/` | 0 files |
| — positive control | `grep -ril "glmcred" internal/ cmd/` | 5 files (`internal/config/envkeys.go`, `internal/config/settings_axis_test.go`, `internal/glmcred/glmcred.go`, `internal/glmcred/glmcred_test.go`, `internal/web/glmkey.go`) |
| Is `scripts/jev/` on develop? | `ls scripts/` | 16 entries, no `jev` |

The zero-hit is paired with a positive control because a zero-hit and a broken search are indistinguishable on their own — the control shows the search fires on a comparable symbol in the same tree.

**Consequence.** This is greenfield in Go. Nothing is being ported, adapted, or kept compatible. The local script set is not committed and therefore cannot be cited as canonical; its disposition is recorded by `SPEC-JEV-GOAL-DIST-001`.

---

## §2. Credential and disclosure precedent

`internal/glmcred/glmcred.go` is the model to follow, and its header states why the package exists at all: exactly one writer implementation, because two would mean two file-mode policies and two escaping rules. It depends only on the standard library and stdlib-only leaves, which is what lets both `internal/cli` and `internal/web` import it while the one-way `cli → web` dependency stays acyclic.

Two details are load-bearing:

- **Mode tightening on write.** `os.WriteFile`'s perm argument applies only at creation, so an existing 0644 file stays 0644 without an explicit `Chmod`. The GLM package closes this; a new package inherits the same latent defect if it forgets.
- **The four-character disclosure floor.** `computeGLMKeyHint` returns `Configured: true` with an *empty* hint for a key of four characters or fewer, and the source notes that a naive "last four or the whole key" fallback would disclose a short key entirely.

`internal/web/glmkey.go` records the structural guarantee: the credential is deliberately outside `settings.AllFields()`, so no generic schema-walking loop can read or render it, and `TestGLMKeyField_AbsentFromSchema` asserts the absence.

---

## §3. Configuration precedent

The template `workflow.yaml` carries three switches in the shape this SPEC needs — `codex.review_gate.enabled: false`, `multi.review_gate.enabled: false`, `slot_lease.enabled: false` — each with a comment stating it ships off and costs nothing while off.

One switch ships *on* (`drift_cache_fill.enabled: true`) and its comment explicitly calls that an accepted cost rather than a neutral one, which is the pattern for justifying a non-inert default. Jev ships off, so it follows the first group.

The local `workflow.yaml` additionally carries `branch_guard.enabled: true` and `agent_stop_guard.enabled: true` — dogfood opt-ins whose template defaults are false. The same local-enables-what-template-ships-off pattern is available for Jev.

---

## §4. Doctor and verification surfaces

Checks are `DiagnosticCheck{Name: "..."}` values (`doctor_codex.go:163`, `doctor_disk.go:72`, `doctor_harness.go:21`, `doctor_hook_delivery.go:55`), addressable by `moai doctor --check "<Name>"`.

The queue-hash method AC-JEVC-001 reuses: `internal/cli/todo_triage_test.go` imports `crypto/sha256` (line 21) and computes `sha256.Sum256(data)` (line 649).

---

## §5. Known model weaknesses (vendor-stated)

From the published jev-1.13 jaggedness notes: literal reading (weak on negation, scoping words, implied conditions); unreliable counting, arithmetic, numeric proximity, and date ordering; weaker multi-hop indirection; accuracy degrading as irrelevant state grows; vulnerability to instructions injected in state; and `P(yes)` not guaranteed to equal `1 − P(no)`. Request limits: 64k tokens total, 32k for state plus the longest question.

These are vendor claims, not measurements taken here. Two of them bind at this SPEC's layer — the state-size sensitivity and the injection vulnerability both argue for bounding and distrusting the state payload, which is what REQ-JEVC-009 and REQ-JEVC-021 do. The rest bind at the question-authoring layer and are owned by `SPEC-JEV-OPTIN-MEASURE-001`.

Using them as design constraints is safe even if a given claim is conservative. No requirement in this SPEC asserts any of them as a measured property of this repository's data.

---

## §6. Open items

| # | Item | Why it is not settled here |
|---|---|---|
| Q2 | Is the pinned model id compiled or operator-visible config? | Compiled is simpler and harder to drift; operator-visible lets a pin move without a release. Affects the `workflow.jev` block's shape. |
| Q4 | Is a credential-reveal route wanted? | The GLM precedent has one; the disclosure requirement is satisfied without it. |
