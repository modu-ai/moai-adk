---
id: SPEC-V3R6-HARNESS-NAMESPACE-V2-001
title: "Harness namespace doctrine-code drift catch-up — acceptance criteria"
version: "0.1.0"
status: completed
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/harness, internal/template, .claude/skills/moai-meta-harness"
lifecycle: spec-anchored
tags: "namespace, harness, doctrine-drift, atomic-transition, generator-emission"
---

# Acceptance — SPEC-V3R6-HARNESS-NAMESPACE-V2-001

## §D. Acceptance Criteria Matrix

### AC-HNS-001 — Atomic transition (HARD)

**Given** the migration commit group is opened as a single PR
**When** all 5 surfaces (Go enforcement, test fixtures, generator emission, CI sentinel, backward-compat recognition) are edited
**Then** the PR shall not merge until all 5 surfaces pass their tests, AND no intermediate commit in the PR shall leave `harness-*` generated without enforcement protection.

**Severity**: MUST-PASS. **Traceability**: REQ-HNS-006, NFR-HNS-004.

**Indirect verification**: `git log --oneline <PR-base>..<PR-head>` shows the surface edits; `go test ./...` passes on the PR head; the PR is merged as one squash-merge (not staged-merge across revisions).

### AC-HNS-002 — `harness-*` vs `moai-harness-*` substring separation (HARD)

**Given** a skills directory containing `harness-foo` (user-owned) and `moai-harness-bar` (template-managed)
**When** `isUserOwnedNamespace(".claude/skills/harness-foo/SKILL.md")` and `isUserOwnedNamespace(".claude/skills/moai-harness-bar/SKILL.md")` are evaluated
**Then** the first shall return `true` and the second shall return `false`.

**Severity**: MUST-PASS. **Traceability**: REQ-HNS-004.

**Indirect verification**: table-driven test in `internal/cli/update_test.go` covering both directions of the prefix comparison.

### AC-HNS-003 — Generator emission boundary (HARD)

**Given** the run-phase diff is produced
**When** the diff for `internal/harness/layer*.go`, `internal/harness/chaining_rules.go`, `internal/harness/types.go`, and `internal/harness/frozen_guard.go` is inspected
**Then** the diff shall contain ONLY prefix-string edits (`my-harness` → `harness`) and comment updates, NOT new functions, new phases, or new skill-skeleton templates.

**Severity**: MUST-PASS. **Traceability**: NFR-HNS-003.

**Indirect verification**: `git diff --stat <base>..<head> -- internal/harness/` shows no new files; `git diff <base>..<head> -- internal/harness/` shows no `+func ` additions (only modifications to existing functions / string literals / comments).

### AC-HNS-004 — No user-data loss (HARD)

**Given** a fixture project with `.claude/skills/my-harness-legacy-skill/SKILL.md` (pre-migration user skill)
**When** `isUserOwnedNamespace(".claude/skills/my-harness-legacy-skill/SKILL.md")` is evaluated during the deprecation window
**Then** it shall return `true` (the legacy prefix is still recognized as user-owned).

**Severity**: MUST-PASS. **Traceability**: REQ-HNS-005, NFR-HNS-006.

**Indirect verification**: table-driven test in `internal/cli/update_test.go` with a `my-harness-*` legacy path.

### AC-HNS-005 — No policy reversal (HARD)

**Given** the migration is complete
**When** `isUserOwnedNamespace(".claude/skills/moai-harness-learner/SKILL.md")` is evaluated
**Then** it shall return `false` (template-managed builder namespace stays unprotected-as-user-owned, i.e. remains template-managed).

**Severity**: MUST-PASS. **Traceability**: NFR-HNS-001.

**Indirect verification**: table-driven test confirming `moai-harness-*` is NOT in the user-owned set.

### AC-HNS-006 — Generator emits `harness-*` only

**Given** the migrated `moai-meta-harness` emission contract
**When** the template Markdown at `internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md` and `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` is grepped for `my-harness`
**Then** zero matches shall remain (the emission prefix is `harness-*` throughout).

**Severity**: MUST-PASS. **Traceability**: REQ-HNS-003.

**Indirect verification**: `grep -rn "my-harness" internal/template/templates/.claude/skills/moai-meta-harness/ internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md` returns 0.

### AC-HNS-007 — CI sentinel detects `harness-*` leak

**Given** a hypothetical template tree with `.claude/skills/harness-leaked/SKILL.md`
**When** `go test -run TestNamespaceLeakHarnessSkills ./internal/template/` runs
**Then** the test shall fail with `NAMESPACE_LEAK_HARNESS_SKILL`.

**Severity**: MUST-PASS. **Traceability**: REQ-HNS-007.

### AC-HNS-008 — Doctor recognizes `harness-*`

**Given** a fixture project with `.claude/skills/harness-ios-patterns/SKILL.md` (no triggers section)
**When** `moai doctor harness` runs
**Then** the L1 check shall flag `harness-ios-patterns` as missing required trigger keys.

**Severity**: SHOULD-PASS. **Traceability**: REQ-HNS-008.

### AC-HNS-009 — Prefix-conflict detector scans `harness-*`

**Given** a fixture skills dir with `harness-foundation-core` and `moai-foundation-core`
**When** `DetectPrefixConflicts(skillsDir)` is called
**Then** the returned conflicts shall include `harness-foundation-core` (suffix `foundation-core` collides with `moai-foundation-core`).

**Severity**: SHOULD-PASS. **Traceability**: REQ-HNS-009.

### AC-HNS-010 — No `moai-builder-*` introduced

**Given** the migration is complete
**When** `grep -rn "moai-builder" internal/ .claude/` runs (excluding `.claude/agent-memory/`)
**Then** zero matches shall remain (the prefix was never introduced).

**Severity**: MUST-PASS. **Traceability**: NFR-HNS-002.

### AC-HNS-011 — Full test suite green

**Given** the migration commit group is complete
**When** `go test ./...` runs
**Then** all tests pass with zero failures.

**Severity**: MUST-PASS. **Traceability**: NFR-HNS-004.

### AC-HNS-012 — Lint clean

**Given** the migration commit group is complete
**When** `golangci-lint run --timeout=2m` runs
**Then** zero findings.

**Severity**: MUST-PASS.

## §D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-HNS-001 | MUST-PASS | Atomic transition — partial migration = user-data loss risk |
| AC-HNS-002 | MUST-PASS | Substring collision = `moai-harness-*` misclassification |
| AC-HNS-003 | MUST-PASS | Generator boundary — run-phase over-reach risk |
| AC-HNS-004 | MUST-PASS | Legacy user-data protection |
| AC-HNS-005 | MUST-PASS | Policy-reversal guard (rejected predecessor's failure mode) |
| AC-HNS-006 | MUST-PASS | Generator emission contract |
| AC-HNS-007 | MUST-PASS | CI leak detection |
| AC-HNS-008 | SHOULD-PASS | Doctor UX completeness |
| AC-HNS-009 | SHOULD-PASS | Diagnostic completeness |
| AC-HNS-010 | MUST-PASS | No-invention guard |
| AC-HNS-011 | MUST-PASS | Regression guard |
| AC-HNS-012 | MUST-PASS | Code-quality guard |

## §D.2 Edge Cases

- **EC-1**: A user project has BOTH `my-harness-legacy` and `harness-newstyle` skill dirs. Both must be recognized during the deprecation window.
- **EC-2**: A path `.claude/skills/harness/SKILL.md` (bare `harness`, no suffix). The `HasPrefix(name, "harness-")` check does NOT match this — it is treated as a custom non-moai skill, protected by the REQ-UNP-009 fallback. This is correct: the user-owned skill namespace is `harness-*` (with suffix), not bare `harness`.
- **EC-3**: A path `.claude/agents/harness/agent.md`. Already protected by `isUserOwnedNamespace` REQ-UNP-002 (update.go:1240-1242). Unchanged by this SPEC.
- **EC-4**: A Windows path `.claude\skills\harness-foo\SKILL.md`. After `ReplaceAll(rel, "\\", "/")` normalization, `HasPrefix(norm, ".claude/skills/harness-")` returns true. NFR-HNS-005.
- **EC-5**: The `cleanMoaiManagedPaths` deletion glob is `.claude/skills/moai*` (update.go:1706-1708). Neither `my-harness-*` nor `harness-*` matches this glob, so the deletion path was already safe for both prefixes. The risk surface is the backup + overlay-write path, which consults `isUserOwnedNamespace`.

## §D.3 Closure Gates

- [ ] All MUST-PASS ACs (001-007, 010-012) verified with verbatim command output
- [ ] SHOULD-PASS ACs (008, 009) verified or debt explicitly recorded
- [ ] Atomic PR merged as single squash-merge
- [ ] `moai spec lint` clean on the SPEC after frontmatter `status` transition
- [ ] Deprecation-window sunset trigger documented (design.md §D) — sunset is a follow-up chore, NOT this SPEC's closure gate

## §D.4 Forward-Looking Checks (Post-Close)

- The `my-harness-*` backward-compat recognition (REQ-HNS-005) should be sunset in a follow-up chore SPEC after a grace period (design.md §D proposes 2 release cycles). That follow-up is out of scope here.
- The doctrine §24.5 drift note (`.moai/docs/harness-namespace-doctrine.md` line 61) is removed by manager-docs at sync-phase of THIS SPEC (documentation cleanup confirming the drift is resolved).

## §D.5 Definition of Done

- 5 surfaces migrated atomically in one PR
- Zero `my-harness` refs remain in `internal/` (except the intentional backward-compat recognition in `isUserAreaPath` / `isUserOwnedNamespace`, documented with an `@MX:NOTE` citing REQ-HNS-005)
- `go test ./...` green; `golangci-lint` clean
- `moai-meta-harness` emission prefix is `harness-*` throughout template Markdown
- CI sentinel renamed and passing
- SPEC frontmatter `status` transitioned `draft → in-progress → implemented → completed` per the canonical 4-phase lifecycle
