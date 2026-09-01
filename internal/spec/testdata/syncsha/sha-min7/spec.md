---
id: SPEC-SSFA-001
title: "sync_commit_sha slot fixture — a 7-character SHA at the inside edge of the band floor"
version: "0.1.0"
status: in-progress
created: 2026-08-29
updated: 2026-08-29
author: Test Author
priority: P2 Medium
phase: "v3.2.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "test, syncsha"
---

# SPEC-SSFA-001: sync_commit_sha slot fixture — a 7-character SHA at the inside edge of the band floor

Only the sibling progress.md's `sync_commit_sha` line varies across this fixture
set. `status: in-progress` is non-terminal and the progress.md below satisfies
era heuristic H-4, so nothing here is demoted to advisory — the fixture measures
the rule rather than the demotion path.

## 2. Scope

### 2.1 In Scope

- The single sync_commit_sha slot in the sibling progress.md

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-SSFA-001-001: The system SHALL validate the sync_commit_sha slot format.

## 6. Acceptance Criteria

- AC-SSFA-001-01: Given the slot, When lint runs, Then the rule decides. (maps REQ-SSFA-001-001)
