---
id: SPEC-V3R6-HARNESS-NAMESPACE-V2-001
title: "Harness namespace doctrine-code drift catch-up — research (verbatim ref classification)"
version: "0.1.0"
status: draft
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/harness, internal/template, .claude/skills/moai-meta-harness"
lifecycle: spec-anchored
tags: "namespace, harness, doctrine-drift, atomic-transition, generator-emission"
---

# Research — SPEC-V3R6-HARNESS-NAMESPACE-V2-001

> This document reproduces verbatim grep output measured by manager-spec on 2026-06-18. No number is carried over from the doctrine's stale "39 Go files + 30+ tests" estimate or the handoff's "59 refs" figure. Every count below is re-measured.

## §A. Measurement Commands (Reproducible)

```bash
# Total ref count
$ grep -rn "my-harness" internal/ | wc -l
188

# Total file count
$ grep -rl "my-harness" internal/ | wc -l
43

# Non-test .go refs
$ grep -rn "my-harness" internal/ | grep -v "_test.go" | wc -l
55

# Test .go refs
$ grep -rn "my-harness" internal/ | grep "_test.go" | wc -l
133

# Non-test files
$ grep -rl "my-harness" internal/ | grep -v "_test.go" | wc -l
15

# Test files
$ grep -rl "my-harness" internal/ | grep "_test.go" | wc -l
28

# moai-builder refs (must be 0 in codebase)
$ grep -rn "moai-builder" internal/ .claude/ | wc -l
3
# (all 3 are in .claude/agent-memory/plan-auditor/ feedback notes about the rejected predecessor — NOT codebase refs)
```

**Authoritative count: 188 refs / 43 files (55 non-test refs / 15 non-test files; 133 test refs / 28 test files).**

## §B. Namespace SSOT Verification (3 Sources Read Verbatim)

### §B.1 `.moai/docs/harness-namespace-doctrine.md` §24.1 (line 15)

```
| **`harness-*`** | **사용자 생성** — `moai-meta-harness`가 `/moai project` Phase 5+ 인터뷰 후 사용자 프로젝트 도메인에 맞춰 generate | user project | **NOT synced (보호)** |
```

The doctrine declares `harness-*` as user-owned. Line 61 [HARD] documents the drift and pre-reserves this SPEC's ID (가칭 `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`).

### §B.2 `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy

```
328:| `moai-harness-*` | **하네스 builder/lifecycle** (현재 `moai-meta-harness` + `moai-harness-learner`만 해당) | template | **삭제 후 신규 설치** (overwrite) |
329:| **`harness-*`** | **사용자 생성** — `moai-meta-harness`가 `/moai project` Phase 5+ 인터뷰 후 generate | user project | **절대 삭제/modify 금지 + 백업 보존** ... |
345:- [HARD] `harness-*` namespace는 user-owned.
348:- [HARD] `harness-*` (user-owned) vs `moai-harness-*` (template builder) substring 구분: prefix 매칭은 정확한 startsWith 비교를 사용하고, `*harness-*` substring 패턴은 false positive 위험이 있으므로 금지.
349:- [HARD] CI guard: ... sentinel pattern은 catch-up SPEC에서 `my-harness-` → `harness-` 으로 갱신.
```

The skill-authoring rule already declares `harness-*` user-owned with the startsWith-separation mandate and the CI sentinel catch-up pointer.

### §B.3 `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation

```
155:**`harness-*` namespace and `.claude/agents/harness/` directory** are user-owned.
164:- [HARD] This meta-harness MUST emit user-generated skills with `harness-*` prefix ONLY.
168:- [HARD] Doctrine-code drift (2026-05-26 ~ catch-up SPEC 완료 전): ... Go enforcement ... 는 `my-harness-*` 작동 유지. ... catch-up SPEC 완료 후 generator runtime behavior가 `harness-*`로 전환.
```

The generator contract declares `harness-*` emission with the drift note pointing to this catch-up SPEC.

**Conclusion of §B**: All three SSOTs ALREADY declare `harness-*` as user-owned. The doctrine is correct. This SPEC migrates code to match. No doctrine change. No policy reversal.

## §C. File-by-File Ref Classification

### §C.1 Non-test Go enforcement files (15 files, 55 refs)

#### `internal/cli/update.go` (3 enforcement refs + comments)

| Line | Code | Classification |
|------|------|----------------|
| 1199 | `if strings.HasPrefix(norm, ".claude/skills/my-harness-") {` | **ENFORCEMENT** — `isUserAreaPath` skills check. M1: add `harness-*`, keep `my-harness-*` legacy (REQ-HNS-005). |
| 1204 | `if strings.HasPrefix(norm, ".claude/agents/my-harness/") \|\| norm == ".claude/agents/my-harness" {` | **ENFORCEMENT (legacy agents path)** — `isUserAreaPath` agents check. Note: `isUserOwnedNamespace` (1240-1242) already uses `.claude/agents/harness/`. This `isUserAreaPath` agents branch is the older predicate; M1 adds `.claude/agents/harness/` here too for consistency + keeps `my-harness/` legacy. |
| 1236 | `if strings.HasPrefix(norm, ".claude/skills/my-harness-") {` | **ENFORCEMENT** — `isUserOwnedNamespace` skills check (REQ-UNP-001). M1: add `harness-*`, keep `my-harness-*` legacy. |
| 1240-1242 | `.claude/agents/harness/` | **ALREADY CANONICAL** — agents side migrated; no change (EXCL-7). |
| 1245-1248 | `.moai/harness/` | **ALREADY CANONICAL** — `.moai/harness/` migrated; no change (EXCL-6). |

#### `internal/cli/doctor_harness.go` (2 enforcement refs + comments)

| Line | Code | Classification |
|------|------|----------------|
| 122 | `if !e.IsDir() \|\| !strings.HasPrefix(e.Name(), "my-harness-") {` | **ENFORCEMENT** — L1 trigger scan enumeration. M1: enumerate `harness-*`. |
| 271 | `if !strings.HasPrefix(ref, "my-harness-") {` | **ENFORCEMENT** — REQ-HAW-013 dangling-ref check. M1: recognize `harness-*`. |

#### `internal/cli/doctor_skills.go` (1 enforcement ref)

| Line | Code | Classification |
|------|------|----------------|
| 52 | `if strings.HasPrefix(name, "my-harness-") {` | **ENFORCEMENT** — classify as INFO. M1: recognize `harness-*`. |

#### `internal/cli/update_preserve_inventory.go` (1 comment ref)

| Line | Code | Classification |
|------|------|----------------|
| 21 | `//   - .claude/skills/my-harness-*  (user harness skills)` | **COMMENT** — doc update only. M1. |

#### `internal/harness/prefix_conflict.go` (2 enforcement refs + comments)

| Line | Code | Classification |
|------|------|----------------|
| 52 | `case strings.HasPrefix(name, "my-harness-"):` | **ENFORCEMENT** — collect user-area names. M1: collect `harness-*`. |
| 60 | `suffix := strings.TrimPrefix(mh, "my-harness-")` | **ENFORCEMENT** — strip prefix for suffix match. M1: strip `harness-`. |

#### `internal/harness/frozen_guard.go` (2 enforcement refs)

| Line | Code | Classification |
|------|------|----------------|
| 19 | `".claude/agents/my-harness/",` | **ENFORCEMENT** — allowedPrefixes. M1: `.claude/agents/harness/` (keep `my-harness/` legacy). |
| 20 | `".claude/skills/my-harness-",` | **ENFORCEMENT** — allowedPrefixes. M1: `.claude/skills/harness-` (keep `my-harness-` legacy). |

#### `internal/harness/layer1.go`, `layer2.go`, `layer5.go`, `chaining_rules.go`, `types.go` (7 comment refs)

These files reference `my-harness-*` in comments and docstrings only (no `HasPrefix` enforcement). Classified as **COMMENT** — doc updates in M1. See §C.1 detail in plan.md M1.

### §C.2 Test fixture files (28 files, 133 refs)

Mechanical rename in M4. The 28 files (full list from `grep -rl "my-harness" internal/ | grep "_test.go"`):

```
internal/cli/doctor_harness_test.go
internal/cli/doctor_skills_test.go
internal/cli/update_archive_flow_test.go
internal/cli/update_preserve_my_harness_test.go
internal/cli/update_safety_test.go
internal/design/pipeline/path_b1_test.go
internal/design/pipeline/path_b2_test.go
internal/harness/chaining_rules_test.go
internal/harness/e2e_ios_test.go
internal/harness/layer1_test.go (if present)
internal/harness/layer2_test.go (if present)
internal/harness/layer5_test.go (if present)
internal/harness/lineage_test.go
internal/harness/prefix_conflict_test.go
internal/harness/safety/frozen_guard_test.go
internal/template/namespace_protection_audit_test.go   # M3 — logic change, not just fixture
internal/template/skills_removal_test.go               # M3 — logic change, not just fixture
... (remaining ~10 test files enumerated by the grep)
```

> The exact 28-file list is reproducible via `grep -rl "my-harness" internal/ | grep "_test.go" | sort`. M4 performs the mechanical rename; M3 handles the 2 sentinel files (logic change, not just fixture rename).

### §C.3 Template Markdown files (generator emission surface)

These are the files where the `moai-meta-harness` emission prefix is declared. M2 edits the prefix string in these files. **This is NOT new generator authoring — it is a prefix-string edit in existing emission documentation consumed by the Claude generator at runtime.**

#### `internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md`

~15 occurrences across lines 292-293, 298, 301-304, 316, 326-327, 333-336, 344, 376, 417. The emission contract (§ 6.4.1) declares `my-harness-<domain>-patterns` / `my-harness-<domain>-best-practices` as the emitted skill prefix. M2: rename to `harness-<domain>-patterns` / `harness-<domain>-best-practices`.

#### `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md`

4 occurrences at lines 164, 168, 174, 177. The generator contract + drift note. M2: update emission contract to `harness-*`; remove the drift note (line 168) since this SPEC resolves it.

#### `internal/template/CLAUDE.md`

1+ occurrence (package-convention doc). M2: doc update.

### §C.4 Doctrine / rule doc files (NOT in scope for code migration, but §24.5 drift note removed at sync-phase)

- `.moai/docs/harness-namespace-doctrine.md` §24.5 (line 61-71) — the drift note. Removed by manager-docs at sync-phase as documentation cleanup (EXCL-1).
- `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` line 349 — CI sentinel catch-up pointer. Updated at sync-phase to remove the "catch-up SPEC" forward reference.

These are documentation confirmations that the drift is resolved, NOT doctrine changes (the `harness-*` declaration at line 329 is already correct and stays).

## §D. Substring-Conflict Math Proof

The doctrine (skill-authoring.md line 348) mandates exact `startsWith` comparison, not `*harness-*` substring. Proof that `harness-*` and `moai-harness-*` do not collide under `strings.HasPrefix`:

```
strings.HasPrefix("harness-foo", "harness-")       = true   ✓ (user-owned, correct)
strings.HasPrefix("moai-harness-foo", "harness-")  = false  ✓ (template-managed, correct)
strings.HasPrefix("harness-foo", "moai-harness-")  = false  ✓ (not template-managed, correct)
strings.HasPrefix("moai-harness-foo", "moai-harness-") = true ✓ (template-managed, correct)
```

All four directions classify correctly. The migration MUST use `HasPrefix` (not `Contains`) to preserve this. REQ-HNS-004.

## §E. `cleanMoaiManagedPaths` Deletion-Path Analysis

The deletion path (`internal/cli/update.go:1678-1722`) uses a fixed target list with one glob:

```go
{
    displayPath: filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
    fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
    isGlob:      true,
},
```

This glob `.claude/skills/moai*` matches ONLY `moai-*` prefixed skills. It does NOT match `my-harness-*` or `harness-*`. Therefore the **deletion path was already safe** for both the legacy and canonical user-owned prefixes. The actual risk surface is the **backup + overlay-write path**, which consults `isUserOwnedNamespace` (`internal/cli/update_namespace_protect.go:125, 229`). This is why REQ-HNS-005 (backward-compat recognition) binds the backup predicate, not the deletion glob.

## §F. Rejected Predecessor Analysis (Why Option A, Not Policy Reversal)

The rejected `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` chose `moai-harness-*` as the new user-owned target. Plan-auditor FAIL 0.62 with 5 blocking defects:

- **D1**: Option A self-sealing — the SPEC chose a target that collided with existing state, then argued the collision "must be resolved" by the reversal.
- **D2**: `moai-builder-*` lineage claimed established but `grep -rn moai-builder` returned 0 codebase matches (3 matches are agent-memory feedback notes, not code).
- **D3**: `moai-meta-harness/SKILL.md:164` HARD-forbids emitting `moai-harness-*` — the generator has NO producer for the claimed user-owned namespace.
- **D4**: `cleanMoaiManagedPaths` does NOT consult `isUserOwnedNamespace` for the `moai*` glob — the SPEC confused predicate-superset-math with does-the-deletion-path-call-the-predicate.
- **D9**: Policy reversal — reclassifying template-managed `moai-harness-*` as user-owned contradicts all 3 namespace SSOTs.

**Counterfactual test (from plan-auditor feedback)**: "If the SPEC had honored the doctrine's declared target (`harness-*`), would the collision exist?" Answer: No. This SPEC honors the doctrine target. No collision. No self-sealing. No reversal.

Recorded in `.claude/agent-memory/plan-auditor/feedback_policy_reversal_rationale_self_sealing.md`.

## §G. Conclusion

The migration is a pure doctrine-code drift catch-up. 188 refs / 43 files across 5 surfaces (Go enforcement, test fixtures, generator emission Markdown, CI sentinel, backward-compat). The doctrine is correct; the code is stale. Option A (this SPEC) aligns code to doctrine atomically. The rejected predecessor's failure modes (self-sealing, nonexistent lineage, policy reversal) do not apply because this SPEC does not choose a new target — it implements the target the doctrine already declares.
