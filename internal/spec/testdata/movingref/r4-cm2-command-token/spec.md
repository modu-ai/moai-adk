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

CM-2 (command-token bypass) of AC-MRG-013 (REQ-MRG-010, card t353). A fetch
chained to a divergence read, carrying a claim. The row names command tokens
(`git fetch`) but carries NO imperative measuring directive and NO demoted
dated reference — an exclusion keyed on a command token would silence this and,
measured at 43329ec8b, 76 of 117 real unpinned divergence lines with it.

| AC | Command | Expected |
|---|---|---|
| AC-X | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0` (no divergence) | 1 finding |

Nothing else in this fixture directory names a moving ref.

## 2. Scope

### 2.1 In Scope

- The single CM-2 row above

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-MGA-001-001: The system SHALL flag an unpinned moving-ref invariant claim.

## 6. Acceptance Criteria

- AC-MGA-001-01: Given the CM-2 row, When lint runs, Then one finding. (maps REQ-MGA-001-001)
