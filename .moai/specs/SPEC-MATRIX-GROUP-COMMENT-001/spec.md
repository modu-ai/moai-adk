---
id: SPEC-MATRIX-GROUP-COMMENT-001
title: "Correct the profile-matrix group-layer comment — AgentGroup has a validation-gate consumer, not only a display one"
version: "0.1.0"
status: in-progress
created: 2026-09-20
updated: 2026-09-21
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/template/profile_matrix.go"
lifecycle: spec-first
tags: "comment-accuracy, profile-matrix, agent-group, doc-comment, card-t1055"
tier: S
era: V3R6
related_specs: [SPEC-MODEL-MATRIX-CORE-001, SPEC-MODEL-PROFILE-MATRIX-001]
---

# SPEC-MATRIX-GROUP-COMMENT-001 — Correct the profile-matrix group-layer comment

Card: **t1055**. Every commit on this branch names it.

## HISTORY

| Date | Change |
|---|---|
| 2026-09-20 | Initial draft (plan phase). Coordinates measured in worktree `.claude/worktrees/t1055` at HEAD `d8304b49a`. |

## §A Context — what is wrong

`internal/template/profile_matrix.go:481-483` carries a Go doc comment on
`ResolveAgentModelEffort` that asserts something the code does not do:

```go
// Lookup is by agent NAME, not by group: per-agent cells split two of the former
// groups, so the group layer no longer carries routing information and survives
// only as a display classification (see AgentGroup).
```

**"survives only as a display classification" is false.** `template.AgentGroup`
(declared at `profile_matrix.go:464`) has exactly **two** production consumers in
this tree, and they differ in kind:

| # | Site | What it uses | Kind |
|---|---|---|---|
| 1 | `internal/cli/model.go:101` | the group **string** (`if g, ok := …; ok { group = g }`), defaulting to `"-"`, stored in `modelProfileEntry.Group` (field at `model.go:38`) and printed as a table column (`model.go:149`, `model.go:154`) | display — matches the comment |
| 2 | `internal/web/agentfm.go:491` | the **bool only** (`if _, ok := …; !ok { continue }`) — the string is discarded; a non-matrix agent's frontmatter-override submission is skipped | **validation gate** — the comment denies this exists |

The second site's own surrounding comment (`agentfm.go:488-490`) states the
intent explicitly: "Overrides are only valid for profile-matrix member agents …
a valid but non-matrix submission is ignored (no override, no frontmatter
write)."

### Why it matters

Someone removing the group layer who reads only the comment at 481-483 concludes
that moving a display column suffices, and silently breaks the `agentfm.go:491`
gate — a submission that should be rejected would be accepted and written. The
defect class is the one card **t1033** fixed: a comment asserting what the code
does not do.

The corrected comment is an **input** to the group-layer removal decision owned
by `SPEC-MODEL-MATRIX-CORE-001`. That is precisely why the two consumers' kinds
must be stated separately rather than summarised.

### Sibling instance in the same file

`internal/template/profile_matrix.go:261` carries the **same false claim**:

```go
// key: retained agent NAME (not a group — the group layer is display-only now,
```

Fixing 481-483 and leaving 261 saying "display-only" reproduces the defect at a
second site in the same file, so it is in scope for "align the comment with the
code".

> **Operator note:** the card text named **only** 481-483. Line 261 was found by
> the lane's measurement pass and folded in on the "same claim, same file"
> rationale. Trim it if you disagree — the requirement for it (REQ-MGC-002) is
> separable from REQ-MGC-001.

## §B Evidence basis

All coordinates re-measured in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1055`,
branch `WT-retention-comment`, HEAD `d8304b49a`.

**Measurement command.**

```
grep -rn --include='*.go' 'AgentGroup' .
```

Returns exactly: `internal/web/agentfm.go:491`, `internal/cli/model.go:101`, the
declaration and its doc comment at `internal/template/profile_matrix.go:462,464`,
the false claim at `profile_matrix.go:483`, and four
`internal/web/agentfm_ordering_test.go` rows (31, 35, 63, 67).

**Negative control / counting trap.** The four `agentfm_ordering_test.go` rows are
`TestAgentGroupRankCoversEveryCatalogClass` and `TestAgentGroupLabelMatchesRank` —
**name lookalikes**. None carries a `template.` qualifier, so none calls
`template.AgentGroup`. They MUST NOT be counted as consumers; an inflated
consumer count would change the prescription. The discriminating command:

```
grep -rn --include='*.go' 'template\.AgentGroup' .
```

returns exactly the two production sites and nothing from the test file.

**Coordinate caveat.** The card cited 481-483 from HEAD `159dd30df`; this tree is
`d8304b49a` and those lines happen to be identical. That is a coincidence, not a
guarantee — do not infer that any other coordinate matches across the two trees.

## §C Requirements (GEARS)

**REQ-MGC-001** (Ubiquitous) — The doc comment on `ResolveAgentModelEffort` in
`internal/template/profile_matrix.go` shall name **both** `template.AgentGroup`
consumers by file path, and shall state that one of them uses the membership
result as a validation gate rather than as a display classification.

**REQ-MGC-002** (Ubiquitous) — The `defaultProfileMatrix` doc comment in
`internal/template/profile_matrix.go` shall not assert that the group layer is
display-only.

**REQ-MGC-003** (Unwanted) — The change shall not modify any non-comment line of
`internal/template/profile_matrix.go`, and shall not modify any other Go file.

**REQ-MGC-004** (Event-detected) — When the package is built and its tests are
run after the change, the Go toolchain shall report the same success it reported
before the change.

## §D Out of Scope

### Out of Scope — group-layer removal

- Whether the group layer should be removed, retained, or restructured. That
  question is owned by `SPEC-MODEL-MATRIX-CORE-001`; this SPEC only makes the
  comment a trustworthy input to it.
- Any deletion or restructuring of `AgentGroup`, `agentGroupMembership`, the
  `modelProfileEntry.Group` field, or the `GROUP` table column.

### Out of Scope — behaviour and tests

- Any production code-behaviour change. The deliverable is comment text only.
- Any test file change, including the four `agentfm_ordering_test.go` lookalikes.

### Out of Scope — the profile-axis token at `:8`

- `internal/template/profile_matrix.go:8` MUST NOT be edited under this SPEC.
  See §E; inclusion is pending an operator decision.

## §E Report-only observation — the profile-axis token at `:8`

Recorded as an observation with its evidence. **No requirement is written for it
and the line is left unedited.**

The same file names the profile axis inconsistently:

| Line | Text |
|---|---|
| `:8` | "a single 3-column profile axis (**max**/medium/low)" |
| `:260` | "Outer key: profile {**high**, medium, low}" |
| `:266` | "each row is monotone: **high** >= medium >= low" |

Which is correct: the outer keys of `defaultProfileMatrix` are the constants
`PerformanceTierHigh` / `PerformanceTierMedium` / `PerformanceTierLow`, defined
at `internal/template/model_policy.go:182,184,186` as `"high"`, `"medium"`,
`"low"`. So `:8`'s "max" is the wrong token for the live axis.

Refinement found by `manager-spec` while authoring this SPEC (**not** carried by
the card brief, which resolved only the live
`PerformanceTierHigh`/`Medium`/`Low` constants): `"max"` is not a typo but the
**superseded name**, still accepted at read time. That makes `:8` a stale-token
instance rather than a spelling slip, which may change how the operator wants it
fixed.

The live mechanism that accepts it is in `internal/config`, not in
`internal/template`:

| Coordinate | Role |
|---|---|
| `internal/config/profile.go:29` | `LegacyProfileMax = "max"` — the declaration |
| `internal/config/profile.go:48-53` | `NormalizeProfile` — `if name == LegacyProfileMax { return ProfileHigh }`, the read-time alias itself |
| `internal/config/profile.go:92-100` | `LLMConfig.EffectiveProfile` — reaches `NormalizeProfile` on both the `llm.profile` and the legacy `llm.performance_tier` paths |
| `internal/config/profile.go:59-61` | `IsValidProfile` — accepts the alias |
| `internal/config/profile.go:66-68` | `ValidProfiles` — deliberately omits it from UI option lists ("readable but never offered") |

### Sub-observation — `template.LegacyPerformanceTierMax` is unused

Smaller, separate, and deliberately not inflated.
`internal/template/model_policy.go:189` declares a second
`"max"` constant, `LegacyPerformanceTierMax`, which **nothing reads**:

```
grep -rn --include='*.go' 'LegacyPerformanceTierMax' internal/
```

returns exactly two rows — its own declaration at `:189` and its doc comment at
`:187`.

Positive control that the probe is not blind — the same grep shape on the live
constant:

```
grep -rn --include='*.go' 'LegacyProfileMax' internal/ | grep -v _test
```

returns five rows, including two genuine reads (`internal/config/profile.go:49`
inside `NormalizeProfile`, and `internal/config/model_routing.go:150`). The probe
finds consumers where they exist, so the zero for the template-package constant
is absence rather than probe failure.

**This is a dead-constant / duplicate-declaration observation, NOT another
instance of this card's false-comment defect class.** Its doc comment ("accepted
as a read-time alias and never written back") is not false about the token — the
token genuinely is accepted, just by `config.LegacyProfileMax` rather than by
this declaration. Report-only, no requirement, like the rest of §E.

## §F Adjacent observation — not a defect claim

`internal/cli/model.go:100` binds the second return of `ResolveAgentModelEffort`
(named `mapped` at its declaration) to a local called `hasGroup`. The name
suggests a group check where the value is matrix membership. This is noted only
because it sits one line above consumer #1; card **t1037** already touched the
`hasGroup` naming axis, so it is neither claimed as a new defect here nor placed
in scope.
