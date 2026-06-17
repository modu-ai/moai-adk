---
id: SPEC-V3R6-HARNESS-NAMESPACE-V2-001
title: "Harness namespace doctrine-code drift catch-up (my-harness-* → harness-*)"
version: "0.1.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/harness, internal/template, .claude/skills/moai-meta-harness"
lifecycle: spec-anchored
tags: "namespace, harness, doctrine-drift, atomic-transition, generator-emission"
supersedes: SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001
depends_on: []
---

# SPEC-V3R6-HARNESS-NAMESPACE-V2-001 — Harness namespace doctrine-code drift catch-up

## §A. Problem Statement

On 2026-05-26, a chore commit declared `harness-*` as the user-owned skill namespace in three doctrine SSOTs:

1. `.moai/docs/harness-namespace-doctrine.md` §24.1 (line 15: `harness-*` = user-owned)
2. `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy (line 329)
3. `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation (line 155)

The Go enforcement layer, however, still recognizes `my-harness-*` as the user-owned skill prefix. This is a **documented intentional doctrine-code drift** (`.moai/docs/harness-namespace-doctrine.md` §24.5 line 61 [HARD]). The drift is operationally safe ONLY because §24.5 imposes a freeze: `moai-meta-harness` must keep emitting `my-harness-*` (not `harness-*`) until this SPEC lands, so no user skill is generated into the unprotected `harness-*` namespace.

This SPEC is the **Phase 2 catch-up**: align Go enforcement + test fixtures + generator emission prefix to the `harness-*` namespace the doctrine already declares. **No doctrine change. No policy reversal.** The doctrine target is correct and fixed; the code is stale.

### Why Option A (doctrine catch-up), not the rejected alternative

A previous SPEC (`SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`, status → `superseded` by this SPEC) chose `moai-harness-*` as the new user-owned target. That was a **policy reversal** — reclassifying a template-managed builder namespace (`moai-harness-*` hosts `moai-meta-harness` + `moai-harness-learner`) as user-owned. It was rejected by plan-auditor (FAIL 0.62, 5 blocking defects D1/D2/D3/D4/D9) and by orchestrator ground-truth verification of the three namespace SSOTs.

This SPEC follows **Option A**: the doctrine already says `harness-*`; we migrate code to match. The `moai-harness-*` classification (template-managed builder) is untouched.

## §B. Background & Cross-References

### §B.1 Doctrine SSOTs (already declare `harness-*` — this SPEC changes code, not doctrine)

| SSOT | Line | Declaration |
|------|------|-------------|
| `.moai/docs/harness-namespace-doctrine.md` §24.1 | 15 | `harness-*` = user-owned, NOT synced |
| `.moai/docs/harness-namespace-doctrine.md` §24.5 | 61 | [HARD] Go code still enforces `my-harness-*`; catch-up SPEC = `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` (가칭) |
| `skill-authoring.md` § Skills Namespace Policy | 329 | `harness-*` = 사용자 생성 |
| `skill-authoring.md` § Skills Namespace Policy | 345-349 | [HARD] `harness-*` vs `moai-harness-*` startsWith separation; CI sentinel pattern updates in catch-up SPEC |
| `.claude/skills/moai-meta-harness/SKILL.md` | 155, 164-168 | `harness-*` user-owned; generator runtime behavior deferred to catch-up SPEC |

### §B.2 Rejected predecessor

`SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` is marked `status: superseded` with a frontmatter `superseded_by: SPEC-V3R6-HARNESS-NAMESPACE-V2-001` pointer. Its plan-auditor FAIL verdict (0.62) and the orchestrator's ground-truth re-verification are recorded in `.claude/agent-memory/plan-auditor/feedback_policy_reversal_rationale_self_sealing.md`.

## §C. Scope (WHAT this SPEC does)

### §C.1 In scope — Go enforcement layer migration

Migrate every `strings.HasPrefix(..., "my-harness-")` and equivalent literal `"my-harness-"` / `"my-harness"` string in the Go enforcement surface to `harness-*` / `harness`. Authoritative ref classification in `research.md` §C; verbatim grep evidence reproduced there.

### §C.2 In scope — Test fixture migration

Migrate every test fixture referencing `my-harness-*` to `harness-*`. Test fixtures are not user data; they are developer-owned test inputs.

### §C.3 In scope — Generator emission prefix migration

Migrate the `moai-meta-harness` emission prefix from `my-harness-*` to `harness-*`. **This is a prefix-string edit in existing emission logic, NOT new generator authorship.** The generator's 7-Phase workflow body, layer architecture, and skill-skeleton templates are unchanged; only the prefix constant/string changes. The emission prefix lives in template Markdown (`.claude/skills/moai/workflows/project/meta-harness.md` + `.claude/skills/moai-meta-harness/SKILL.md` § 6.4.1), consumed by the Claude generator at runtime — NOT a Go code constant. This boundary is load-bearing for the run-phase constraint "생성 코드 새로 작성 안 함 — namespace rename만".

### §C.4 In scope — CI sentinel pattern update

Update `internal/template/namespace_protection_audit_test.go` `TestNamespaceLeakMyHarnessSkills` and `internal/template/skills_removal_test.go` from `my-harness-` pattern to `harness-` pattern (doctrine §24.5 line 68).

### §C.5 In scope — Backward-compat recognition for existing user skill dirs

Existing user projects may have `my-harness-*` skill directories generated before this SPEC lands. The atomic transition MUST NOT leave them unprotected. `isUserAreaPath` / `isUserOwnedNamespace` MUST recognize BOTH `my-harness-*` (legacy) AND `harness-*` (canonical) during a deprecation window, so an `moai update` does not delete a pre-migration user skill. See `design.md` §D for the dual-recognition window and its sunset.

## §D. Requirements (GEARS)

### REQ-HNS-001 — Go enforcement recognizes `harness-*` as user-owned skill namespace

The `moai update` preservation logic shall treat any path matching `.claude/skills/harness-*` as user-owned and shall not delete, modify, or sync it.

**Ubiquitous.** Subject: `moai update` preservation logic.

### REQ-HNS-002 — Go enforcement recognizes `harness/` as user-owned agent directory

The `moai update` preservation logic shall treat any path matching `.claude/agents/harness/` as user-owned and shall not delete, modify, or sync it.

**Ubiquitous.** Subject: `moai update` preservation logic.

> Note: `isUserOwnedNamespace` at `internal/cli/update.go:1240-1242` ALREADY protects `.claude/agents/harness/` (REQ-UNP-002). The agents-dir side is already migrated; only the skills side (`my-harness-*`) remains stale. This REQ is stated for completeness and to lock the already-migrated state.

### REQ-HNS-003 — Generator emits `harness-*` prefix only

When `moai-meta-harness` generates a project-specific domain skill, the generator shall emit the skill under the `harness-*` prefix and shall not emit any `my-harness-*` prefixed artifact.

**Ubiquitous.** Subject: `moai-meta-harness` generator.

### REQ-HNS-004 — `harness-*` vs `moai-harness-*` substring separation

Where the preservation logic tests whether a skill path belongs to the user-owned namespace, the test shall use an exact `strings.HasPrefix(name, "harness-")` comparison and shall not use a `*harness-*` substring pattern, so that a `moai-harness-*` template-managed skill is not misclassified as user-owned.

**Where** the preservation logic classifies a skill path. **Ubiquitous** (the constraint binds whenever classification runs). Subject: classification predicate.

### REQ-HNS-005 — Backward-compat recognition of legacy `my-harness-*` during deprecation window

While the deprecation window is active, the preservation logic shall additionally recognize `.claude/skills/my-harness-*` as user-owned, so that a user skill generated before this SPEC is not deleted by `moai update`.

**While** the deprecation window is active. Subject: preservation logic. See `design.md` §D for window definition and sunset trigger.

### REQ-HNS-006 — Atomic transition (no window where `harness-*` is unprotected)

When the Go enforcement switches to recognize `harness-*`, the switch shall land in the same commit (or tightly coupled commit group within one PR) as the generator emission switch, the CI sentinel switch, and the test fixture switch, so that at no revision does a generated `harness-*` skill lack enforcement protection or an enforced `harness-*` pattern lack a generator producing it.

**Ubiquitous.** Subject: the migration commit group. This is the atomic-transition HARD AC.

### REQ-HNS-007 — CI sentinel detects `harness-*` leak into template

When the CI template-neutrality audit scans `internal/template/templates/.claude/skills/`, the audit shall fail if any directory named `harness-*` is present, indicating a user-owned namespace leaked into the template tree.

**When** the CI template-neutrality audit scans the skills tree. Subject: CI sentinel (`TestNamespaceLeakHarnessSkills`, renamed from `TestNamespaceLeakMyHarnessSkills`).

### REQ-HNS-008 — Doctor command recognizes `harness-*`

When `moai doctor harness` scans user skill directories, the command shall enumerate `harness-*` prefixed directories and shall report any missing L1 trigger keys against the `harness-*` enumeration.

**When** `moai doctor harness` enumerates skills. Subject: `moai doctor harness`.

### REQ-HNS-009 — Prefix-conflict detector scans `harness-*`

When the prefix-conflict detector scans the skills directory, the detector shall collect `harness-*` directory names (stripping `harness-` → suffix) and shall flag semantic collisions against `moai-*` skills using the same suffix.

**When** the prefix-conflict detector scans the skills directory. Subject: `internal/harness/prefix_conflict.go` `DetectPrefixConflicts`.

## §E. Constraints (Non-Functional)

### NFR-HNS-001 — No policy reversal

`moai-harness-*` shall remain template-managed (builder namespace: `moai-meta-harness`, `moai-harness-learner`). This SPEC shall not reclassify `moai-harness-*` as user-owned.

### NFR-HNS-002 — No `moai-builder-*` invention

The namespace `moai-builder-*` does not exist in the codebase (verified: `grep -rn moai-builder internal/ .claude/` returns 0 real refs; the 3 matches in `.claude/agent-memory/plan-auditor/` are feedback notes about the rejected predecessor SPEC, not codebase refs). This SPEC shall not introduce it.

### NFR-HNS-003 — Generator constraint (emission-prefix edit only)

The run-phase implementation shall edit the emission prefix in existing emission logic (template Markdown strings + any Go-side prefix constants) and shall not author new generator code, new generator phases, or new skill-skeleton templates.

### NFR-HNS-004 — Atomic transition

The migration shall land as one atomic commit group within one PR. Partial migration (Go switches, generator doesn't) is a HARD FAILURE because `harness-*` skills would be generated without protection.

### NFR-HNS-005 — Cross-platform path normalization preserved

Every modified prefix comparison shall preserve the existing `strings.ReplaceAll(rel, "\\", "/")` slash normalization (NFR-UNP-003 lineage) so Windows backslash paths classify correctly.

### NFR-HNS-006 — No user-data loss

No existing user `my-harness-*` skill directory shall be left unprotected at any revision in the migration commit group. REQ-HNS-005 backward-compat recognition is the mechanism.

## §F. Out of Scope (What NOT to Build)

### Out of Scope — Doctrine SSOTs unchanged

- This SPEC does NOT change the doctrine SSOTs. `.moai/docs/harness-namespace-doctrine.md`, `skill-authoring.md` § Skills Namespace Policy, and `moai-meta-harness/SKILL.md` § Namespace Separation already declare `harness-*`; they are correct. The §24.5 drift note (line 61) will be removed by manager-docs at sync-phase as a documentation cleanup, NOT as a doctrine change.

### Out of Scope — `moai-harness-*` classification untouched

- This SPEC does NOT touch the `moai-harness-*` classification. `moai-meta-harness` and `moai-harness-learner` remain template-managed. Reclassifying them is the rejected predecessor's failure mode (NFR-HNS-001).

### Out of Scope — No `moai-builder-*` invention

- This SPEC does NOT introduce a `moai-builder-*` namespace. That lineage does not exist (`grep -rn moai-builder internal/ .claude/` = 0 codebase refs; 3 matches are agent-memory feedback notes about the rejected predecessor).

### Out of Scope — No new generator authoring

- This SPEC does NOT author new generator code. The 7-Phase workflow body, layer1/layer2/layer5 architecture, and skill-skeleton templates are unchanged. Only the emission prefix string changes (NFR-HNS-003).

### Out of Scope — No rename of `moai-meta-harness` skill

- This SPEC does NOT rename the `moai-meta-harness` skill itself. The meta-harness skill name stays `moai-meta-harness` (it is the generator, template-managed); only the prefix of what it EMITS changes.

### Out of Scope — `.moai/harness/` already canonical

- This SPEC does NOT migrate `.moai/harness/` paths. `isUserOwnedNamespace` at `update.go:1245-1248` already protects `.moai/harness/` (REQ-UNP-003) with the canonical `harness/` name; no drift exists there.

### Out of Scope — `.claude/agents/harness/` already canonical

- This SPEC does NOT migrate `.claude/agents/harness/` paths. `isUserOwnedNamespace` at `update.go:1240-1242` already protects `.claude/agents/harness/` (REQ-UNP-002) with the canonical `harness/` name; only the skills side drifted.

## §G. Acceptance Criteria Summary

Acceptance criteria (Given-When-Then) live in `acceptance.md`. The HARD ACs are:

- **AC-HNS-001** (atomic transition): all 5 surfaces (Go enforcement, test fixtures, generator emission, CI sentinel, backward-compat) switch in one PR.
- **AC-HNS-002** (substring separation): `harness-*` and `moai-harness-*` do not collide under `HasPrefix`.
- **AC-HNS-003** (generator boundary): run-phase produces zero new generator code (emission-prefix edit only).
- **AC-HNS-004** (no user-data loss): legacy `my-harness-*` recognized during deprecation window.
- **AC-HNS-005** (no policy reversal): `moai-harness-*` stays template-managed.

## §H. History

- **2026-06-18**: plan-phase artifacts authored (manager-spec). Supersedes rejected `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` (iter-1 plan-auditor FAIL 0.62, Option A self-sealing + nonexistent `moai-builder-*` lineage + generator HARD contract violation). This SPEC follows Option A (doctrine catch-up): the doctrine already declares `harness-*`; code migrates to match. No policy reversal.
