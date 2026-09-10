---
id: SPEC-UPDATE-MERGE-CONFLICT-BLIND-001
title: "moai update merge: an unreachable conflict detector and no signal when a shared key is preserved"
version: "0.3.1"
status: in-progress
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
- 2026-09-10 — **third framing revision**, authored after M1 was executed. M1's `untouched_shared` cell measured the merge writing the user's value for a key whose template value had changed, which establishes a fact wider than §A's plan-phase premise: for a top-level JSON key the user's file already carries, `moai update` cannot change that key's value at all. Recorded as §A.6. The two earlier framings are **retained** — the withdrawn one in §A.2's opening record, the plan-phase two-defect reading in §A.2 proper — rather than erased; §A.6 supersedes the plan-phase reading in breadth only.
- 2026-09-10 — **fourth framing revision**, authored after M2.0 measured the recursive and the YAML breadths. The measurement forced a **granularity correction** to §A.6, not a defect discovery: the post-M1 wording — that the merge cannot change the value of a key the user's file already carries — is true at **leaf** granularity and false at **container** granularity, because a container both sides carry gains the template's new leaves. §A.6 is restated around the canonical sentence; its post-M1 wording is **superseded and retained here as the record**, alongside the withdrawn first framing and §A.2's plan-phase reading. §A.3 moves the recursive and YAML breadths from unmeasured to measured and carries M2.0's own Gaps list forward.
- 2026-09-10 — v0.3.1, **annotation only**: `REQ-UMC-008` and `REQ-UMC-009` still say "shared key", which predates §A.6's leaf/container split. Their text is deliberately left unchanged — restating them is entangled with M2.1's choice of conflict-determination granularity — and each now carries an annotation marking the wording stale, backed by an entry gate at M2.1 in `plan.md` §F. No REQ text changed and no acceptance criterion was added or removed.

## §A Context

### §A.1 What the merge does, and why it does it

`moai update` deploys the embedded template over the user's file, then runs a per-file 3-way merge to carry the user's customizations back in (`internal/cli/update/merge/merge.go` `MergeUserFiles`). A 3-way merge needs a base — the template content the user's file was originally deployed from — and that content is not stored anywhere. `internal/cli/update/merge/base.go` therefore **derives** a base: the freshly deployed template, narrowed to the keys the user's file also carries (`pruneToShared`).

The derivation states its own intent verbatim at `internal/cli/update/merge/base.go:109-111`:

> Values always come from updated. A key present on both sides therefore enters the base carrying the template's value, which is what makes a user's edit to that key read as their change during the merge.

So for any key both sides carry, `baseVal == updVal` by construction (`base.go:128`, `pruned[key] = updatedVal`), and a user's value winning on a shared key is the **designed** outcome. The same file's header comment already names the one thing the derived base cannot express — a template value change to a key the user never touched — and names its closure: snapshotting the deployed template at deploy time, which is the surface of `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`.

### §A.2 The two defects, and what they are not

> **Superseded in breadth by §A.6, and retained as the record of the first two framings.** The plan-phase reading below — two defects sitting in the gap between the code's stated intent and its instrumentation — remains correct as far as it goes. M1 measured something wider and M2.0 fixed its granularity, which §A.6 states. Nothing here is deleted: the withdrawn framing is recorded in HISTORY, and this section is the plan-phase framing it replaced.

[HARD] A user's value winning on a shared key is **designed behaviour, not a defect**. A SPEC that calls the code's stated intent a defect collapses on one reading of the source. The defects sit in the **gap between that intent and its instrumentation**.

**Defect 1 — an unreachable conflict detector.** `internal/merge/strategies.go:419-420` computes:

```go
baseChanged := !valuesEqual(baseVal, curVal)
updChanged  := !valuesEqual(baseVal, updVal)
```

Because the derived base set `baseVal` **to** `updVal` for every shared key, `updChanged` is always false there. The `default:` both-changed arm at `strategies.go:435-436` — the arm that appends a `Conflict` and raises `MergeResult.HasConflict` — therefore cannot execute for any shared key. Only the `baseChanged && !updChanged` arm ("only user changed", `strategies.go:427-429`, `result[key] = curVal`) can fire.

A reader of `mergeJSON` / `mergeYAML` sees a populated `Conflicts` slice and a `HasConflict` flag and believes `moai update` detects template-versus-user conflicts. It does not, for any shared key, and nothing in the code or in the user-facing output says so. This repository already names this shape: **an unreachable check is an instrument defect.**

**Defect 2 — no path for the template to re-assert, and no signal when it matters.** Whatever the user's file holds for a shared key wins, whatever wrote it. For most keys that is correct and desirable. For a **security-relevant key emptied by an unidentified writer**, that state is preserved indefinitely and the user is never told.

### §A.3 Scope is wider than the card's subject — measured breadth, and the breadth still open

[HARD] The scope is wider than `permissions.ask`, which is the **instance that exposed the defects**, not the extent of them. But the measured breadth and the asserted breadth must not be confused, so this section states them separately.

**Measured — three breadths, 20 predictions, 0 failed** (`progress.md` §E.2):

| Breadth | Milestone | Fixture | Codec |
|---|---|---|---|
| Top-level JSON keys | M1 | flat document, four cells | `json_merge` |
| Recursive path — a nested leaf under a container both sides carry | M2.0 | four cells one level down, plus a case where the whole container is absent on the user's side | `json_merge` |
| YAML, flat and nested | M2.0 | the four cells driven through the YAML strategy, then the nested intersection that `.moai/config/sections/*.yaml` actually has | `yaml_deep`, asserted per cell |

Each breadth carried **its own discriminator cell** — the one cell whose prediction inverts, the template's value landing rather than the user's, so a harness that merely copied the user's document fails it — and **its own conflict-surface control**: the same engine over the same key, given a base agreeing with neither side, reporting `HasConflict=true` / `len(Conflicts)=1`. A control inside one codec says nothing about the other, so the YAML control is measured rather than inherited from the JSON one.

**The YAML path is the identical call, not a parallel strategy.** `mergeYAML` (`internal/merge/strategies.go:343`) and `mergeJSON` (`:314`) both call `deepMergeMap`; `three_way.go:56-59` dispatches the two from the extension map at `strategies.go:79-82`, and `deriveTemplateBase` routes YAML through `yaml.Unmarshal` / `yaml.Marshal` into the same `pruneToShared` (`base.go:75-78`). The arm structure of `strategies.go:418-462` is therefore shared, and the two paths differ only in the codec that produces the three maps and re-serializes the result. Because every YAML cell asserts `MergeResult.Strategy == yaml_deep`, a YAML reading cannot be a JSON reading wearing a `.yaml` file name.

**Not measured — the breadth still open.** Carried faithfully from M2.0's own Gaps list; none of the following may be asserted:

- Which **arm** of `strategies.go:422-462` fires for a shared container key. `MergeResult` does not distinguish them, so the arm is not inferable from the result.
- **Non-string YAML keys**, and value shapes beyond scalars, string arrays, and nested maps.
- Any **end-to-end `moai update`** run, and any real checkout — `permissions.ask` has been measured nowhere outside a fixture.
- Cross-platform (`GOOS=windows`) behaviour, and any coverage delta.

[HARD] So the three breadths above are established and the list above is a gap. What now scopes `§A.6` is **granularity** — leaf versus container — rather than codec; `§A.2`'s defects are likewise established across all three breadths at leaf granularity. Citing either beyond that scope is an unobserved claim.

### §A.4 The healing boundary — omitted heals, empty does not

Measurable in the same code, and load-bearing for the reproduction:

- A key **absent** from the user's file is not shared, so `pruneToShared` leaves it out of the base (`base.go:116-120`); the merge reads the template as *introducing* it and the template's value lands. The key **heals**.
- A key present but holding an **empty array** is shared, so the user's `[]` wins. The key does **not** heal.

This is why the in-repo `toolpolicy` writers — which omit `ask` entirely when it is empty (`internal/config/toolpolicy/settings_region.go:204`) — would have self-healed on the next update, while the literal `"ask": []` produced by a writer **outside this repository** would not.

### §A.5 What was unexecuted at plan-phase, and what M1 and M2.0 have since measured

The plan-phase statement, retained as the record it was: §A.1-§A.4 were derived from reading `base.go` and `strategies.go`; no test had been run against a live file; Defect 1 and Defect 2 were **hypotheses** until the §C M1 reproduction was observed.

**M1 has since been executed**, and **M2.0 after it**. The evidence for both is `progress.md` §E.2. Together they discharge the hypothesis marker at the three breadths §A.3 names — top-level JSON, nested JSON, and flat and nested YAML — and at **leaf granularity** (§A.6). What is established, and what is not:

- **Established.** §A.4's healing boundary (an omitted key or container heals, an emptied one does not); the shared-leaf conflict surface reading `HasConflict=false` / `len(Conflicts)=0` on a genuinely divergent shared key, against a control on the same engine and the same key that fired `true` / `1` under a base differing from both sides — so the reading is a property of the derived base, not of the harness, and the YAML control is measured on its own codec rather than inherited. The §A.6 canonical statement's three clauses.
- **Not established.** Everything in §A.3's gap list: which arm fires for a shared container key, non-string YAML keys, value shapes beyond scalars / string arrays / nested maps, any end-to-end `moai update` run, any real checkout, cross-platform behaviour, coverage delta. No production code changed in either milestone, so no `REQ-UMC-008` / `009` / `011` obligation is discharged and M2.1's design choice stays open.

### §A.6 Fourth framing revision (post-M2.0) — a shared LEAF's value cannot change

[HARD] This is the **current** framing. It supersedes this section's post-M1 wording — which stated the invariance of "a key the user's file already carries" — in **granularity**, not in correctness of the mechanism it described, and it erases nothing: §A.2 stands as the record of the plan-phase framing, the withdrawn first framing stands in HISTORY, and the superseded post-M1 wording is recorded there too, dated.

**Why the revision happened is worth stating plainly: this is a granularity correction, not a defect discovery.** M2.0 measured the recursive and the YAML breadths and every prediction held (20 of 20). What it also showed, in the same runs, is that the earlier sentence was pitched one level too high in the document tree.

**The canonical statement.**

> A shared leaf's value cannot change. A leaf absent on the user's side lands. A container's value changes only by gaining such leaves.

Each of the three clauses is measured, at all three breadths of §A.3:

- **A shared leaf's value cannot change.** M1's `untouched_shared` cell held `["template-old"]` on the user's side while the template shipped `["template-new"]`; the merge wrote `["template-old"]`. The nested-JSON and the flat- and nested-YAML runs reproduced it on their own fixtures and codecs.
- **A leaf absent on the user's side lands.** The discriminator cell of every breadth: an omitted key or leaf is not shared, so it stays out of the base and the template reads as introducing it. This is §A.4's healing boundary, and it is what separates a merge that ran from a document left alone.
- **A container's value changes only by gaining such leaves.** In both nested runs the container `container` was carried by both sides, and the written container read `{"changed_shared":["user-choice"],"emptied_shared":[],"omitted_leaf":["template-only"],"untouched_shared":["template-old"],"user_only":["user-addition"]}` — the user's leaves **plus** the template's new leaf. That value is neither side's, which is precisely why the invariance cannot be stated of "a key". The recursion is `base.go:105-111` working as it describes; it is not a defect.

**The mechanism.** `pruneToShared` sets the base to the template's value for every shared key (`base.go:128`), so at `strategies.go:419-420` `updChanged` is always false there, while `baseChanged` is true whenever the user's value differs from the incoming template's — the ordinary case for a user sitting on an older template. The `baseChanged && !updChanged` arm ("only user changed", `strategies.go:427-429`) returns the user's value. Of the four arms of the shared-key switch (`strategies.go:422-462`), only two are reachable for a shared **leaf**: **"Only template changed" (`strategies.go:431-433`) and "Both changed" (`strategies.go:435-436`) cannot execute there.**

[HARD] **What is not observable, and therefore not claimed.** Which arm of `strategies.go:422-462` fires for a shared **container** key. `MergeResult` carries the written value, the conflict flag, and the strategy — it does not distinguish the arms — so the arm may not be inferred from the result. The reachability sentence above is stated of leaves for that reason, and a reading of it that covers containers is an unobserved claim (§A.3).

**The load-bearing point, unchanged by the correction.** `pruneToShared` states its intent at `base.go:109-111`: a key on both sides enters the base carrying the template's value, "which is what makes a user's edit to that key read as their change during the merge." The mechanism **cannot distinguish "the user edited this leaf" from "the user is on an older template version"** — both present as a difference between the user's value and the incoming template's, and both therefore read as the user's change. The stated intent is satisfiable only under an assumption that is false for the ordinary updating user, which is precisely the population `moai update` exists to serve.

This is why §A.2's Defect 2 needs a remedy of its own and why `REQ-UMC-010` is not in tension with it: the resolution is designed and stays, but a subsystem that cannot tell an edit from a stale deployment must not present a surface implying it can.

**The breadth of each sentence above.** The three clauses of the canonical statement, the healing boundary, and the leaf-level reachability reading are **measured** across top-level JSON, nested JSON, and flat and nested YAML (§A.3). The container-arm question is **not observable** and is claimed nowhere. Everything in §A.3's gap list — end-to-end `moai update`, a real checkout, non-string YAML keys, value shapes beyond scalars / string arrays / nested maps, cross-platform behaviour — remains unmeasured, and citing this section as though it covered them is an unobserved claim.

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

> **Granularity annotation (2026-09-10).** This requirement's "shared key" wording predates the leaf/container split of §A.6 and is stale relative to the measured mechanism. It must be re-adjudicated as leaf, container, or both before M2.1 begins — see the entry gate at M2.1 in `plan.md` §F. The requirement sentence above is unchanged.

**REQ-UMC-009** — The merge subsystem shall not present a conflict-detection surface — a `Conflicts` slice, a `HasConflict` flag, or user-facing output derived from either — that a shared key cannot reach.

> **Granularity annotation (2026-09-10).** This requirement's "shared key" wording predates the leaf/container split of §A.6 and is stale relative to the measured mechanism. It must be re-adjudicated as leaf, container, or both before M2.1 begins — see the entry gate at M2.1 in `plan.md` §F. The requirement sentence above is unchanged.

**REQ-UMC-010** — The merge subsystem shall continue to resolve a shared key in favour of the user's value, preserving the behaviour `base.go:109-111` states as its intent; no remedy for REQ-UMC-008 or REQ-UMC-009 shall change which value the merge writes for a shared key.

### §B.4 Signal on preserved-and-security-relevant state

**REQ-UMC-011** — **When** the merge preserves a user-side value for a key the template ships with a non-empty security-relevant value, the update subsystem shall emit a signal identifying the key and both values.

**REQ-UMC-012** — **Where** no such divergence exists, the update subsystem shall emit no signal, so that the signal of REQ-UMC-011 remains readable rather than becoming routine output the user learns to skip.

**REQ-UMC-013** — The update subsystem shall not modify, restore, or delete any user-side permission value as a consequence of REQ-UMC-011; the obligation is to report, not to repair.

### §B.5 Template-First coupling

**REQ-UMC-014** — **Where** a change in this repository touches `.claude/settings.json`, the same change shall touch `internal/template/templates/.claude/settings.json.tmpl`, because `moai update` redeploys the file wholesale (`CLAUDE.local.md` §2.3), and shall be followed by `make build`.

**REQ-UMC-015** — A template-only change shall not be treated as reaching a user whose on-disk value is already empty: per §A.4 an empty array is shared and the user's side wins, so the asymmetry between REQ-UMC-014 and REQ-UMC-011 is why Defect 2 needs a remedy of its own rather than a template edit.

## §C Milestone shape

Decision-reversibility order, most-likely-to-change first. Full plan in `plan.md`.

- **M1 — Reproduction (no code change).** Observe the four cells. **Executed** — evidence in `progress.md` §E.2; it produced the §A.6 revision. Everything downstream is conditional on what M1 records.
- **M2 — Instrument honesty.** REQ-UMC-008/009/010. Entered by a **precondition measurement** (M2.0): the recursive `pruneToShared` path and the `mergeYAML` path are measured **before** any repair (§A.3), because the fixed scope and the believed scope diverge silently otherwise. **M2.0 is executed** — evidence in `progress.md` §E.2; it produced the §A.6 granularity correction. The design decision of M2.1 (make the arm reachable, or declare the surface's limit) remains open: it was deliberately not taken in the measurement milestone.
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
- [HARD] Every source line cited in this SPEC was re-read and re-confirmed against tree **`715078afb`** (worktree `.claude/worktrees/t576`, branch `WT-permissions-ask-empty`). A line citation decays like a HEAD reading, so the SHA travels with it; a reader on a later tree re-measures rather than trusting the number.
- `internal/cli/update/merge/base.go:109-111` (the stated intent), `:116-120` (the not-shared `continue`), `:124-126` (the nested-map recursion), `:128` (`pruned[key] = updatedVal`) — the derived base.
- `internal/merge/strategies.go:418` (the shared-key case), `:419-420` (`baseChanged` / `updChanged`), `:422-462` (the four arms), `:427-429` (only user changed), `:431-433` (only template changed — unreachable), `:435-436` (both changed — unreachable), `:452-458` (the `Conflict` append) — the arms of the shared-key switch.
- `internal/config/toolpolicy/settings_region.go:204` — the in-repo key-omission behaviour that makes the §A.4 boundary observable.
- `CLAUDE.local.md` §2.3 — the wholesale-redeploy behaviour REQ-UMC-014 derives from.
- Cards **t598**, **t599** — the sibling instrumentation gaps, excluded above.

🗿 MoAI
