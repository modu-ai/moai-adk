---
id: SPEC-GLM-JEV-KEY-001
title: "Decision index — glm --key flag and moai jev command"
created: 2026-10-09
---

# SPEC-GLM-JEV-KEY-001 — decision-index.md

interview.decision_gate is `on` (`.moai/config/sections/interview.yaml:6`). One row per decision surfaced during assembly that the operator did not settle verbatim. The operator's card settled the deliverable set (two commands, mirror-the-GLM-approach, setup unchanged) and the scope boundary (no t1612 pull-in); the rows below are what assembly still had to resolve.

### Q1: Where does the new `moai jev --key` command store the credential?

Label: DECIDED
Authority anchor: SPEC-JEV-CORE-001 §C REQ-JEVC-018 (status: completed; `.moai/specs/SPEC-JEV-CORE-001/spec.md`)
Why unresolved: the card said "mirror the GLM key storage approach" and left the concrete file/env-var/format pin to this SPEC; assembly found the question already decided by the completed predecessor — `internal/jevcred` IS the mirror, mandated as the single writer with a fixed location.
Pinned citation (verbatim from the committed SPEC):

```
REQ-JEVC-018 (Ubiquitous) The API credential shall live at `~/.moai/.env.typesafe` at file mode 0600, outside the repository, read and written through one package modelled on `internal/glmcred` — including that package's tightening of a pre-existing wider mode on write.
```

Operator verdict: RESOLVED-BY-AUTHORITY SPEC-JEV-CORE-001 REQ-JEVC-018 (completed) — jev storage = `~/.moai/.env.typesafe`, dotenv key `TYPESAFE_API_KEY` (package-owned), mode 0600; see plan.md §D.4. Consequence: none of the existing guidance lines (`mcp_jev.go:87`, `doctor_jev.go:82/88`, web fieldsets) changes — they already name this file.

### Q2: What may the jev save confirmation disclose about the credential?

Label: DECIDED
Authority anchor: SPEC-JEV-CORE-001 §C REQ-JEVC-020 (status: completed; same file)
Why unresolved: the card said "mask like saveGLMKey does"; assembly had to choose between glm's `maskAPIKey` (first4+last4) and the jev domain's own asserted disclosure contract.
Pinned citation (verbatim from the committed SPEC):

```
REQ-JEVC-020 (Ubiquitous) A view of the credential shall disclose only a `configured` boolean and, for a credential longer than four characters, its final four characters — never the credential itself, and never any part of a credential of four characters or fewer.
```

Operator verdict: RESOLVED-BY-AUTHORITY SPEC-JEV-CORE-001 REQ-JEVC-020 (completed) — the confirmation uses `jevcred.View()`'s bounded disclosure (plan.md §D.3). The glm `--key` path reuses the existing `maskAPIKey` string unchanged (its own domain's contract).

### Q3: What does bare `moai jev` (no `--key`) do?

Label: FOUNDER
Class: implementation-level
Why unresolved: the command is new, so no prior SPEC or operator setting decides its bare-invocation default; the card ordered only the `--key` form and t1612 owns the richer guidance surface.
Default: print help and exit 0 (rule: smaller user-visible surface — a help render touches nothing, and the undo is a single revert of this SPEC's commits; it also leaves t1612 the cleanest seam).
Alternate: exit non-zero with a one-line pointer to `--key`.
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:05:06Z t1613-spec/manager-spec

### Q4: What happens when `--key` appears together with other glm arguments (launch flags or a subcommand token)?

Label: FOUNDER
Class: implementation-level
Why unresolved: the card did not address the combined spelling; with `DisableFlagParsing` the combined input would otherwise fall into launch parsing with a flag token in it — silent misbehavior.
Default: refuse with a usage error naming the conflict and store nothing (rule: preserves current no-`--key` behavior for every existing invocation shape, and the repo's fail-loud precedent — an ambiguous save must not half-happen).
Alternate: `--key` wins and remaining tokens are ignored (silently swallows a launch flag — rejected).
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:05:06Z t1613-spec/manager-spec

---

Routing note: no product-level FOUNDER rows exist — both deliverable surfaces were ordered by the operator in the card itself, and neither changes an existing shipped command's default behavior (`moai glm` without `--key` and `moai glm setup <key>` are untouched per AC-GJK-006/008).
