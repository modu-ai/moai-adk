---
id: SPEC-MRGA-001
title: "Moving-ref fixture — unpinned anchor"
version: "0.1.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: Test Author
priority: P2 Medium
phase: "v3.2.0"
module: "internal/spec"
era: V3R6
dependencies: []
bc_id: []
lifecycle: spec-anchored
tags: "test, movingref"
breaking: false
related_rule: []
---

# SPEC-MRGA-001: Moving-ref fixture — unpinned anchor

The flagged row lives in THIS file, not in a sibling artifact, and the placement is
load-bearing rather than incidental. `SPECDoc.Body` carries `spec.md` alone, so a
row placed here survives AC-MRG-009's body-only mutant — which is what lets that
mutant take AC-MRG-009 red while leaving AC-MRG-001 green. A row placed in
`acceptance.md` instead takes both criteria down together, and the separation the
two criteria assert becomes unobservable.

| AC | Command | Expected |
|---|---|---|
| AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) |

Nothing else in this fixture directory names a moving ref, so the only finding the
linter reports is the one `MovingRefUnpinned` above.

## 2. Scope

### 2.1 In Scope

- The single flagged row in this spec.md body

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-MGA-001-001: The system SHALL flag an unpinned moving-ref invariant claim.

## 6. Acceptance Criteria

- AC-MGA-001-01: Given the flagged row, When lint runs, Then one finding. (maps REQ-MGA-001-001)
