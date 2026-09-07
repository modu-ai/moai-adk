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

AC-MRG-013's retained fixture line (REQ-MRG-010, card t353). Both R4 conjuncts
are present: the imperative measuring directive ("run `<cmd>` at read time") and
the demoted dated reference carrying the value the command produces
("(reference reading 2026-08-28: empty)"). Measured flaggable absent the
exclusion: moving-ref token 1, hex SHA 0, claim marker 1, git-command context 1.

| AC | Command | Expected |
|---|---|---|
| AC-X | verify `internal/hook` is unchanged by this work: run `git diff --name-only origin/develop -- internal/hook` at read time (reference reading 2026-08-28: empty) | 0 findings |

Nothing else in this fixture directory names a moving ref.

## 2. Scope

### 2.1 In Scope

- The single R4-form row above

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-MGA-001-001: The system SHALL flag an unpinned moving-ref invariant claim.

## 6. Acceptance Criteria

- AC-MGA-001-01: Given the R4-form row, When lint runs, Then zero findings. (maps REQ-MGA-001-001)
