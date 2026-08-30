---
id: SPEC-ASTF-001
title: "Artifact-statelessness fixture — plan.md carries status"
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
tags: "test, artifactstatus"
breaking: false
related_rule: []
---

# SPEC-ASTF-001: Artifact-statelessness fixture — plan.md carries status

The rule under test reads this SPEC's SIBLING artifacts, never this file. This
body is deliberately clean so the only findings reported are the ones the case
is about.

## 2. Scope

### 2.1 In Scope

- The sibling artifacts of this fixture directory

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-ASF-001-001: The system SHALL reject a status field in a non-spec.md artifact.

## 6. Acceptance Criteria

- AC-ASF-001-01: Given the fixture, When lint runs, Then the expected findings. (maps REQ-ASF-001-001)
