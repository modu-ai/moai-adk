# Implementation Plan — SPEC-RETRY-IDEMPOTENCY-001

## §A. Context

Augment `.claude/rules/moai/core/agent-common-protocol.md` § Error Recovery Pattern with a
retry-safety-asymmetry principle (observe-before-retry gate for side-effecting tool calls),
mirror the edit into the template source, and run `make build`. This is a doctrine-layer
rule augmentation with no Go code changes.

## §B. Known Issues / Preconditions

- The deployed § Error Recovery Pattern is a 4-step numbered list at
  `.claude/rules/moai/core/agent-common-protocol.md` (heading `### Error Recovery Pattern`).
- The template mirror exists at
  `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` — the
  **augmentation block only** (the § Error Recovery Pattern insertion) must be region-scoped
  byte-identical between the two after the edit. **Whole-file byte-parity does NOT hold** for
  this file: it is NOT in `workflowOptMirroredPaths` (`internal/template/rule_template_mirror_test.go`)
  — it is leak-test-covered, so the deployed copy legitimately carries internal SPEC-IDs that
  the template mirror strips (~33-line baseline divergence). Do NOT run a whole-file diff.
- The CI internal-content leak guard (`internal/template/internal_content_leak_test.go`)
  forbids SPEC-ID / REQ-AC token / audit-citation / internal-date / commit-SHA / archive-path
  patterns in template files. The augmentation prose must be neutral of all of these.
- The template-neutrality CI workflow (`.github/workflows/template-neutrality-check.yaml`)
  fires on `internal/template/templates/**` path change.

## §C. Pre-flight Checklist

- [ ] Read the current § Error Recovery Pattern block (both deployed + template) before editing.
- [ ] Confirm the 4-step list and the constitution "Maximum 3 retries" line are the exact
      anchors, unchanged by the edit.
- [ ] Draft the augmentation as an appended block after step 4 (not a rewrite of steps 1-4).

## §D. Constraints (HARD)

1. **Augment, do not replace.** The existing 4-step list and the constitution 3-retry line
   remain verbatim. The augmentation is an appended block under § Error Recovery Pattern.
2. **Template-first mirror.** Edit deployed + template mirror; run `make build`; the
   **augmentation block** (§ Error Recovery Pattern insertion) remains region-scoped
   byte-identical between the two. Whole-file byte-parity does NOT hold for this file (it is
   leak-test-covered, not in `workflowOptMirroredPaths` — see §B).
3. **Deployed-rule neutrality.** Generic prose only. No SPEC-ID, no REQ/AC token, no
   external-commentary citation, no author name, no internal date, no commit SHA in the
   deployed rule body. Provenance stays in `.moai/specs/` only.
4. **Anti-duplication.** Reference the existing step 3 ("do not retry the identical call")
   as the rule being extended; do not restate the ledger-closure invariant (§ Ledger
   Closure) or the Pre-Spawn Sync Check (both are different concerns).
5. **Scope discipline.** Touch only `agent-common-protocol.md` (deployed + template) and run
   `make build`. Do NOT edit `moai-constitution.md`.

## §E. Self-Verification Deliverables

The run-phase self-verification (E1 AC matrix, E5 lint status, template-parity grep) is
recorded in `progress.md` §E.2 / §E.3 at run-phase by manager-develop. This plan section is
a pointer only.

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Augment the deployed rule (Priority: High)

Append the following block immediately after step 4 of § Error Recovery Pattern in
`.claude/rules/moai/core/agent-common-protocol.md`. **Draft augmentation wording** (final
wording may be tightened at run-phase, keeping the required keywords):

> **Retry safety is asymmetric with respect to a call's side effects.** Before applying
> the 3-retry ceiling above, classify the failed call by its side-effect profile:
>
> - **Idempotent / read-only calls** (re-reading a file, re-running a search or query,
>   re-running an initializer, fetching a URL) may be retried up to the ceiling — repeating
>   them produces the same observable result, so a transient failure (a file lock, a
>   network blip) is legitimately recovered by a retry.
> - **Side-effecting calls** (writing/editing a file, committing, pushing, opening a pull
>   request, deploying, mutating external-API state) carry a duplicate-effect hazard. When
>   a side-effecting call fails *ambiguously* — the failure signal is present but whether
>   the effect already landed is uncertain — first **observe the current state** to
>   determine whether the effect already occurred, and retry only when the effect is
>   confirmed absent. Retrying a side-effecting call without first observing state is the
>   duplicate-effect hazard: a blind retry after an uncertain failure risks a duplicate
>   commit, a duplicate pull request, or a double deploy. The absence of a success signal
>   is not evidence the effect did not land.
>
> This refines step 3 above ("do not retry the identical call") along the side-effect axis:
> for a side-effecting call, "try an alternative approach" begins with observing whether the
> effect already occurred.

### M2 — Mirror into the template + build (Priority: High)

- Apply the same augmentation block, byte-identical, to the corresponding § Error Recovery
  Pattern region of
  `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`.
- Run `make build` to regenerate the embedded template.
- Verify the **augmentation block region** is byte-identical between deployed + template
  (region-scoped diff of the § Error Recovery Pattern insertion — NOT a whole-file diff, which
  false-fails on the ~33-line internal-SPEC-ID baseline divergence per §B). Additionally grep
  the augmentation keywords ("side-effect", "idempotent", "observe") in the regenerated
  `internal/template/embedded.go` to confirm the mirror landed.

### M3 — Verify guards + parity (Priority: Medium)

- `go test ./internal/template/...` — internal-content leak guard + neutrality pass.
- `golangci-lint run --timeout=2m` — baseline clean.
- Grep parity: the idempotency-asymmetry keywords present in both deployed + template;
  the existing 4-step list + constitution 3-retry line unchanged.

## §G. Anti-Patterns

- **AP-RI-001** — Rewriting steps 1-4 instead of appending. The augmentation is additive.
- **AP-RI-002** — Restating the ledger-closure invariant or the Pre-Spawn Sync Check inside
  the augmentation (both are different concerns; reference by concept, do not duplicate).
- **AP-RI-003** — Leaking a SPEC-ID / date / citation into the deployed rule text.
- **AP-RI-004** — Editing only the deployed file and forgetting the template mirror + build
  (or vice versa) — leaves the two out of sync and fails parity.
- **AP-RI-005** — Changing the "Maximum 3 retries" count or the retry interval — out of scope.

## §H. Cross-References

- `.claude/rules/moai/core/agent-common-protocol.md` § Error Recovery Pattern (the edit target).
- `.claude/rules/moai/core/moai-constitution.md` § Error Handling Protocol ("Maximum 3
  retries per operation" — left unchanged).
- `.claude/rules/moai/core/verification-claim-integrity.md` (the "no unobserved success
  claim" invariant — the completion-declaration promotion is out of scope, see spec.md §C).
- `CLAUDE.local.md` §2 Template-First Rule + §25 Template Internal-Content Isolation.

## Tier Classification

**Tier S** (minimal). Rationale: a single deployed rule file + its one template mirror, an
append of ~1 paragraph (~3-4 sentences) with no Go code, no schema change, and no behavior
change to any binary. Scope is 2 file edits + `make build` + a read-only verification batch.
This matches the Tier S envelope (single rule file + 1 template mirror, doc-only, low risk).
