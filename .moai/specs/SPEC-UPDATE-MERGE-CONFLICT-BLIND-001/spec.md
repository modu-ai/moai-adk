---
id: SPEC-UPDATE-MERGE-CONFLICT-BLIND-001
title: "moai update merge: an unreachable conflict detector and no signal when a shared key is preserved"
version: "0.1.0"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/merge"
lifecycle: spec-anchored
tags: "update, merge, instrumentation, conflict-detection, settings"
era: V3R6
related_specs: ["SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001", "SPEC-UPDATE-YAML-PRESERVE-001", "SPEC-PREMERGE-SETTINGS-DRIFT-001"]
tier: M
---

# SPEC — moai update merge: an unreachable conflict detector and no signal when a shared key is preserved

## HISTORY

- 2026-09-10 — plan-phase artifacts authored (Tier M: spec.md + plan.md + acceptance.md + progress.md). Origin: card **t576**, whose investigation is committed verbatim at `.moai/reports/t576/verdict.md`. The card arrived describing a `permissions.ask` symptom in `.claude/settings.json`; the investigation established the symptom's writer as external to this repository and surfaced this merge finding as a **latent second finding**, static-derived and explicitly **not executed**. This SPEC's first milestone is therefore the reproduction, not a repair.
- 2026-09-10 — a framing used earlier in the t576 investigation — "the merge is broken because the both-changed branch cannot fire" — is **withdrawn** and MUST NOT reappear in this SPEC or its siblings. See §A.2.

## §A Context

### §A.1 What the merge does, and why it does it

`moai update` deploys the embedded template over the user's file, then runs a per-file 3-way merge to carry the user's customizations back in (`internal/cli/update/merge/merge.go` `MergeUserFiles`). A 3-way merge needs a base — the template content the user's file was originally deployed from — and that content is not stored anywhere. `internal/cli/update/merge/base.go` therefore **derives** a base: the freshly deployed template, narrowed to the keys the user's file also carries (`pruneToShared`).

The derivation states its own intent verbatim at `internal/cli/update/merge/base.go:105-107`:

> Values always come from updated. A key present on both sides therefore enters the base carrying the template's value, which is what makes a user's edit to that key read as their change during the merge.

So for any key both sides carry, `baseVal == updVal` by construction (`base.go:126`, `pruned[key] = updatedVal`), and a user's value winning on a shared key is the **designed** outcome. The same file's header comment already names the one thing the derived base cannot express — a template value change to a key the user never touched — and names its closure: snapshotting the deployed template at deploy time, which is the surface of `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`.

### §A.2 The two defects, and what they are not

[HARD] A user's value winning on a shared key is **designed behaviour, not a defect**. A SPEC that calls the code's stated intent a defect collapses on one reading of the source. The defects sit in the **gap between that intent and its instrumentation**.

**Defect 1 — an unreachable conflict detector.** `internal/merge/strategies.go:419-420` computes:

```go
baseChanged := !valuesEqual(baseVal, curVal)
updChanged  := !valuesEqual(baseVal, updVal)
```

Because the derived base set `baseVal` **to** `updVal` for every shared key, `updChanged` is always false there. The `default:` both-changed arm at `strategies.go:434` — the arm that appends a `Conflict` and raises `MergeResult.HasConflict` — therefore cannot execute for any shared key. Only the `baseChanged && !updChanged` arm ("only user changed", `strategies.go:426`, `result[key] = curVal`) can fire.

A reader of `mergeJSON` / `mergeYAML` sees a populated `Conflicts` slice and a `HasConflict` flag and believes `moai update` detects template-versus-user conflicts. It does not, for any shared key, and nothing in the code or in the user-facing output says so. This repository already names this shape: **an unreachable check is an instrument defect.**

**Defect 2 — no path for the template to re-assert, and no signal when it matters.** Whatever the user's file holds for a shared key wins, whatever wrote it. For most keys that is correct and desirable. For a **security-relevant key emptied by an unidentified writer**, that state is preserved indefinitely and the user is never told.

### §A.3 Scope is wider than the card's subject

[HARD] Both defects bind **every shared key** in every JSON and YAML file that passes through this merge — not `permissions.ask` alone. `permissions.ask` is the **instance that exposed them**, not the extent of them.

### §A.4 The healing boundary — omitted heals, empty does not

Measurable in the same code, and load-bearing for the reproduction:

- A key **absent** from the user's file is not shared, so `pruneToShared` leaves it out of the base (`base.go:117-121`); the merge reads the template as *introducing* it and the template's value lands. The key **heals**.
- A key present but holding an **empty array** is shared, so the user's `[]` wins. The key does **not** heal.

This is why the in-repo `toolpolicy` writers — which omit `ask` entirely when it is empty (`internal/config/toolpolicy/settings_region.go:204`) — would have self-healed on the next update, while the literal `"ask": []` produced by a writer **outside this repository** would not.

### §A.5 Everything above is unexecuted

[HARD] §A.1-§A.4 are derived from reading `base.go` and `strategies.go`. No test was run against a live file. Until the reproduction in §C M1 is observed, Defect 1 and Defect 2 are **hypotheses**, and no downstream milestone may cite them as established.

## §B Requirements (GEARS)

### §B.1 Observation

**REQ-UMC-001** — The reproduction harness shall execute the update merge path against an isolated temporary repository created for the run, and shall not read from or write to any real project checkout.

**REQ-UMC-002** — The reproduction harness shall not invoke the `moai update` command against a real checkout; it shall exercise the merge subsystem directly, or a temporary repository it created itself.

**REQ-UMC-003** — When the reproduction runs, the harness shall measure every control cell of §B.2 in the same run, against the same fixture, and shall record each cell's observed outcome verbatim.

**REQ-UMC-004** — **When** the merge subsystem processes a shared key whose current and updated values disagree, the reproduction harness shall record the resulting `MergeResult.HasConflict` value and the length of `MergeResult.Conflicts`, so that the reachability of the both-changed arm is measured rather than inferred.

### §B.2 The control cells

**REQ-UMC-005** — The fixture shall carry four keys, distinguished by how the user's side treats them: (i) a shared key the user did not touch, (ii) a shared key the user emptied to `[]`, (iii) a shared key whose value the user changed to a different non-empty value, and (iv) a key the user's file omits entirely.

**REQ-UMC-006** — **Where** the four cells of REQ-UMC-005 are measured in one run, cell (iv) shall serve as the counter-cell: cells (i), (ii), and (iii) are all predicted to end with the user's side standing, so a harness that merely copied the user's file would pass all three. Cell (iv), predicted to end with the template's value landing, is the cell that distinguishes a merge that ran from a file that was left alone.

**REQ-UMC-007** — The reproduction shall not be reported as observed while any one of the four cells is unmeasured, and an unmeasured cell shall be recorded as a gap rather than inferred from its siblings.

### §B.3 Instrument honesty

**REQ-UMC-008** — **When** the merge subsystem cannot determine whether a shared key's divergence is a template change, a user change, or both, the merge subsystem shall not report the outcome as an unconflicted merge.

**REQ-UMC-009** — The merge subsystem shall not present a conflict-detection surface — a `Conflicts` slice, a `HasConflict` flag, or user-facing output derived from either — that a shared key cannot reach.

**REQ-UMC-010** — The merge subsystem shall continue to resolve a shared key in favour of the user's value, preserving the behaviour `base.go:105-107` states as its intent; no remedy for REQ-UMC-008 or REQ-UMC-009 shall change which value the merge writes for a shared key.

### §B.4 Signal on preserved-and-security-relevant state

**REQ-UMC-011** — **When** the merge preserves a user-side value for a key the template ships with a non-empty security-relevant value, the update subsystem shall emit a signal identifying the key and both values.

**REQ-UMC-012** — **Where** no such divergence exists, the update subsystem shall emit no signal, so that the signal of REQ-UMC-011 remains readable rather than becoming routine output the user learns to skip.

**REQ-UMC-013** — The update subsystem shall not modify, restore, or delete any user-side permission value as a consequence of REQ-UMC-011; the obligation is to report, not to repair.

### §B.5 Template-First coupling

**REQ-UMC-014** — **Where** a change in this repository touches `.claude/settings.json`, the same change shall touch `internal/template/templates/.claude/settings.json.tmpl`, because `moai update` redeploys the file wholesale (`CLAUDE.local.md` §2.3), and shall be followed by `make build`.

**REQ-UMC-015** — A template-only change shall not be treated as reaching a user whose on-disk value is already empty: per §A.4 an empty array is shared and the user's side wins, so the asymmetry between REQ-UMC-014 and REQ-UMC-011 is why Defect 2 needs a remedy of its own rather than a template edit.

## §C Milestone shape

Decision-reversibility order, most-likely-to-change first. Full plan in `plan.md`.

- **M1 — Reproduction (no code change).** Observe the four cells. Everything downstream is conditional on what M1 records.
- **M2 — Instrument honesty.** REQ-UMC-008/009/010. The design decision (make the arm reachable, or declare the surface's limit) is deliberately left open until M1 is read.
- **M3 — The signal.** REQ-UMC-011/012/013.

## §D Exclusions

### Out of Scope — restoring or modifying permission values

- Restoring, editing, or deleting any real `permissions` value in this or any other checkout. Card t576 forbids it, and the investigation supports the prohibition: the writer that emptied `ask` is outside this codebase, so no change here prevents recurrence.
- Treating a restored `ask` list as a live defence. `settings.local.json` sets `permissions.defaultMode: "bypassPermissions"`, which makes the project `ask` layer inert regardless of its contents.

### Out of Scope — the identity of the external writer

- Establishing which action of the external tool rewrote `.claude/settings.json`. The verdict establishes the writer as outside this repository and records the triggering action as **unestablished** (`verdict.md` § Gaps).
- Establishing intent behind the emptying. Nothing observed distinguishes a deliberate removal from an incidental loss.

### Out of Scope — the drift-ledger instrumentation gaps

- A restoration leaving no trace in the drift ledger — card **t598**.
- `preserved_path` recorded under a non-canonical path case (`/Users/goos/moai/…` vs `/Users/goos/MoAI/…`) — card **t599**.
- Both are cross-referenced, not absorbed. They live in the `integration acquire` drift-preservation path, not in the merge.

### Out of Scope — the deploy-time template snapshot

- Persisting the rendered template at deploy time so a genuine 3-way base exists on the next update. That is the surface of `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` (status `completed`, scoped to `.moai/config/sections/*.yaml`). This SPEC may cite it as the structural closure of §A.1's inherited limitation, and may not re-specify or widen it.

### Out of Scope — widening beyond the named surface

- Any file, package, or subsystem outside `internal/merge/` and `internal/cli/update/merge/`, except the template-side coupling REQ-UMC-014 names explicitly.
- Changing which value the merge writes for a shared key (REQ-UMC-010 forbids it).

## §E Cross-references

- `.moai/reports/t576/verdict.md` — the authoritative investigation input; §A restates it, and the verdict's Gaps section is binding on what this SPEC may claim.
- `internal/cli/update/merge/base.go:105-107`, `:117-121`, `:126` — the derived base and its stated intent.
- `internal/merge/strategies.go:419-420`, `:426`, `:434` — the arms of the shared-key switch.
- `internal/config/toolpolicy/settings_region.go:204` — the in-repo key-omission behaviour that makes the §A.4 boundary observable.
- `CLAUDE.local.md` §2.3 — the wholesale-redeploy behaviour REQ-UMC-014 derives from.
- Cards **t598**, **t599** — the sibling instrumentation gaps, excluded above.

🗿 MoAI
