---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Design — Page-3 answer persistence architecture"
version: "0.2.1"
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: L
---

# Design — SPEC-CLI-WIZARD-RESTRUCTURE-001

Added at v0.2.0 with the Tier M → L re-tier. Scope: the **D4 persistence
architecture** decision surfaced by plan-audit review-2 N1. The UX decisions
(D1 3-page grouping, D2 default recalibration, D3 question removal) were
settled at v0.1.0-v0.1.1 and are not re-opened here; they live in plan.md §A.1.

All evidence cited below is measured in `research.md`. No open questions
remain — every option was evaluated to a decision.

## §D1 — Where should the Page-3 writes run?

The Page-3 answers must reach disk on the path a real `moai init` takes. Today
they reach `WritePhase1Configs`, whose only caller sits inside
`generateConfigsFallback` — unreachable from the CLI (research.md §R1).

| Option | Shape | Verdict |
|---|---|---|
| **A — Step 3d sibling of Step 3c (CHOSEN)** | Move the `WritePhase1Configs(opts, result)` call out of `generateConfigsFallback` and place it immediately after the existing Step 3c `writeReportConfig` call at `initializer.go:195`, i.e. **outside** the `if i.deployer != nil { … } else { … }` block, with the same non-fatal `result.Warnings` error handling. | **CHOSEN.** It is the codebase's own documented answer to precisely this problem — Step 3c exists because `report.yaml` had the identical "template writes a default, wizard must override it" shape, and its comment says so verbatim. Zero new concepts; one call relocation. |
| B — Call it from both branches | Leave the call in `generateConfigsFallback` and add a second call inside the `i.deployer != nil` branch. | **REJECTED.** Two call sites for one obligation; a future edit to one branch silently diverges. Step 3c deliberately avoided this. |
| C — Move the writes into the deployer | Have the template deployer render the Page-3 answers into the `.tmpl` files via `TemplateContext`. | **REJECTED.** Requires new `TemplateContext` fields (spec.md §C forbids new config fields), only works for the two `.tmpl` targets (`harness.yaml`/`design.yaml` are static, research.md §R2), and couples a CLI-answer concern to template rendering. Larger blast radius than the whole rest of the SPEC. |
| D — Post-init `moai config set` pass | Persist via the existing config-write API after `Initialize` returns. | **REJECTED.** Adds a second write pass over files just deployed, needs an error/rollback story for a partially-initialized project, and leaves `WritePhase1Configs` as dead code the retirement would then also have to remove. |

**Consequence of A:** `WritePhase1Configs` becomes unconditional (Gate 2, the
`if !opts.StandardMode` early return, is removed — plan.md C31) and the
`InitOptions.StandardMode` field becomes dead (C33). Both are required anyway
by the REQ-WIZ-018 retirement.

## §D2 — How should the writers avoid destroying deployed content?

Making the writers reachable is the easy half. Three of the five do a wholesale
`os.WriteFile` of a 2-4 line document; reaching them without changing them
destroys 22,338 bytes of deployed config per `moai init` (research.md §R2).

The audit prescribed "convert them to read-patch (`patchYAMLKey`)". That
prescription is **unsafe as stated** — `patchYAMLKey` matches on
whitespace-stripped lines and rewrites at a hardcoded 2-space indent, so
against `lsp.yaml` it also rewrites L323 and against `design.yaml` it rewrites
L25/44/55/76, flattening all of them to depth 2 (research.md §R5).

| Option | Shape | Verdict |
|---|---|---|
| **A — Additive depth-aware patch helper (CHOSEN)** | Add a new helper alongside `patchYAMLKey` that targets a key by full path (`design.claude_design.enabled`) and preserves the original indentation of the line it rewrites. Convert `writeLSPYAML` + `writeDesignYAML` to use it. Leave `patchYAMLKey` and its two existing callers untouched. | **CHOSEN.** Correct for nested keys, additive (no regression surface for the two working callers), and testable in isolation — AC-WIZ-017's indentation-multiset assertion distinguishes it from the naive implementation. |
| B — Fix `patchYAMLKey` in place | Make the existing helper depth-aware. | **REJECTED.** Its two current callers (`writeProjectModeYAML`, `writeQualityExpansionYAML`) work today and are out of scope (spec.md §C). Changing shared behaviour to serve two new callers risks the one Page-3 answer whose write path already functions (`enforce_quality`). |
| C — Adopt a real YAML library round-trip | Parse to a node tree, set the key, re-serialize. | **REJECTED for this SPEC.** Comment- and format-preserving round-trips are the hard case; a non-preserving marshal would strip every comment from an 11 KB annotated config — a different flavour of the same destruction. Introducing a new parsing dependency across five writers is a larger change than the SPEC it is embedded in. A future SPEC may revisit. |
| D — Accept the clobber, document it | Add a REQ stating deployed config is replaced by init. | **REJECTED.** `lsp.yaml` is the 16-language LSP configuration that CLAUDE.local.md §15 language-neutrality exists to protect; silently deleting it on every init is a defect, not a documentable trade-off. |

**Per-writer disposition:**

| Writer | Disposition | Rationale |
|---|---|---|
| `writeProjectModeYAML` | unchanged | already read-patch; `project.mode` is unique in its file |
| `writeQualityExpansionYAML` | unchanged | already read-patch (research.md §R4 corrects the audit); `enforce_quality` unique in its file |
| `writeLSPYAML` | → depth-aware read-patch (C35) | live Page-3 answer; 11,306 B at risk; one nested collision |
| `writeDesignYAML` | → depth-aware read-patch (C35) | two live Page-3 answers, one of them nested (`claude_design.enabled`); four nested collisions |
| `writeHarnessProfileYAML` | **removed from the set** (C36) | its question is deleted by REQ-WIZ-012 and the deployed `harness.yaml` already ships `default_profile: "default"` — converting it would be work to restate a correct value while risking 8,165 B |

**Ordering is load-bearing.** C35/C36 must land before C32. Making the writers
reachable (C32) before making them non-destructive (C35/C36) produces a window
in which every `moai init` deletes deployed config. plan.md §F M5 fixes the
intra-milestone order and §G names the inversion as an anti-pattern.

## §D3 — Flag versus wizard precedence

Removing Gate 1 (C20) silently inverts an existing behaviour: today the gate
discards the wizard result in Quick mode so an explicit `--enable-lsp` wins;
after removal the wizard answer overwrites it unconditionally (plan.md §B
B-precedence).

| Option | Verdict |
|---|---|
| **A — Flag wins when explicitly supplied (CHOSEN)** | **CHOSEN.** Consistent with the rule already documented one screen above the gate for `--profile` (`init.go:332-334`: *"the wizard fills opts.Profile only when the flag is absent, so the flag takes precedence over the wizard answer"*), and an explicitly typed flag is unambiguous intent. Encoded as REQ-WIZ-020, verified by AC-WIZ-016. |
| B — Wizard wins | **REJECTED.** Contradicts the neighbouring `--profile` rule, producing two opposite precedence rules in one function. |
| C — Leave it unstated | **REJECTED.** This is what review-2 N8 flagged: a run-phase implementer has no rule to follow when the two disagree, so the behaviour becomes an accident of edit order. |

**Implementation constraint carried into plan.md C30:**
`getBoolFlagWithDefault(cmd, "enforce-quality", true)` and
`getBoolFlag(cmd, "enable-lsp")` return a value that cannot distinguish "flag
absent" from "flag explicitly false". A value-only implementation of option A
is therefore wrong for the `false` case. `cmd.Flags().Changed(<name>)` is
required, and AC-WIZ-016 exercises `--enforce-quality=false` /
`--enable-lsp=false` specifically so a value-only implementation fails.

## §D4 — Tier re-evaluation: M → L

Review-1 called Tier M "defensible but boundary" at ~15-16 files. Review-2's N1
added a package.

**File count: 22** (12 production + 7 existing test + 3 new test files, all
enumerated in research.md §R8; the 3 new files are plan.md C39/C40/C41, made
explicit at v0.2.1 per review-3 D2). Tier M's envelope is 5-15 files.

Count alone is the weaker argument. The decisive factors are qualitative:

1. **Architectural relocation, not a mechanical edit.** C32 moves a
   config-write obligation from one execution path to another across a package
   boundary (`internal/cli` → `internal/core/project`). It changes which code
   runs on the primary user path.
2. **A data-loss hazard with an ordering constraint.** 22,338 bytes of
   deployed configuration are exposed if C32 lands before C35/C36. Milestone-
   internal ordering becoming load-bearing is a Tier L signal.
3. **A new primitive.** C34 introduces a YAML-patch helper that did not exist,
   with its own correctness AC (AC-WIZ-017) and its own failure mode
   (depth-flattening) that a naive implementation would exhibit silently.
4. **Two falsified audit claims** (research.md §R4, §R5) — the second of which
   inverted a prescribed fix from "safe" to "unsafe". A scope where the audit's
   own remediation needed correction is not a standard-envelope scope.
5. **Cross-package test surface.** `internal/core/project` tests must be
   rewritten from "asserts a no-op" to "asserts non-destructive patching" —
   a semantic rewrite, not an assertion touch-up.

**Decision: Tier L.** Artifact set extended to 6 files (spec / plan /
acceptance / progress / design / research) to match.

**Split option considered and rejected.** Review-2's STOP escalation offered
option 1: land M1-M4 + M6 under this SPEC with REQ-WIZ-015/AC-WIZ-010 removed,
and route the persistence wiring to a follow-up SPEC. Rejected because it ships
exactly the outcome plan.md §G names as its first anti-pattern — a cosmetic
page restructure whose answers are discarded — and leaves a released version in
which `moai init` presents five settings that do nothing. The re-tier keeps the
SPEC honest about its own size instead of shrinking the SPEC to fit the tier.

## §D5 — Design invariants for run-phase

1. Persistence correctness is measured **on disk after a deployer-path init**,
   never on `project.InitOptions` (AC-WIZ-010).
2. A patch touches **one key**; every other byte of the file survives
   (AC-WIZ-010a).
3. A patch never changes the indentation of a line it did not target
   (AC-WIZ-017).
4. Where a setting's correct value is already the shipped template default and
   its question has been removed, the design answer is **no write path**, not a
   redundant one (`harness.default_profile`, `coverage_exemptions.enabled`).
5. Any new grep-based acceptance check states its expected match count against
   the pre-change tree (acceptance.md §D.3) — the standing rule after review-1
   D1 and review-2 N4.
