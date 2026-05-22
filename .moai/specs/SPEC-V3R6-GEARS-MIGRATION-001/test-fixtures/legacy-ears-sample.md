---
id: SPEC-FIXTURE-LEGACY-001
title: "Legacy EARS Lint Fixture"
version: "0.0.1"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: GEARS-MIGRATION-001
priority: P3
phase: "v3.0.0"
module: "test-fixtures"
lifecycle: spec-anchored
tags: "fixture, ears, legacy"
---

# SPEC-FIXTURE-LEGACY-001: Legacy EARS Sample Fixture

This fixture exercises the 5 legacy EARS patterns. The IF/THEN REQ (REQ-LEG-001-005)
triggers exactly one `LegacyEARSKeyword` warning per AC-GM-001/AC-GM-002.

## 2. Scope

### 3.1 Out of Scope

- GEARS canonical notation (covered by gears-sample.md)
- Real-world REQ semantics (synthetic test data only)

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-LEG-001-001: The system SHALL always emit structured findings.

### 5.2 Event-driven (WHEN)

- REQ-LEG-001-002: WHEN the user runs `moai spec lint`, the system SHALL parse all SPEC files.

### 5.3 State-driven (WHILE)

- REQ-LEG-001-003: WHILE the lint engine is active, the system SHALL maintain rule registry consistency.

### 5.4 Optional precondition (WHERE)

- REQ-LEG-001-004: WHERE a registry path is configured, the system SHALL load zone references.

### 5.5 Legacy conditional (IF/THEN)

- REQ-LEG-001-005: IF a deprecated keyword is detected THEN the system SHALL emit a migration warning.

## 6. Acceptance Criteria

- AC-LEG-001-01: Given the fixture, When linted in non-strict, Then exit 0. (maps REQ-LEG-001-001)
- AC-LEG-001-02: Given the fixture, When linted, Then 1 LegacyEARSKeyword finding. (maps REQ-LEG-001-002)
- AC-LEG-001-03: Given the fixture, When linted, Then no ModalityMalformed errors. (maps REQ-LEG-001-003)
- AC-LEG-001-04: Given the fixture, When linted, Then registry loaded. (maps REQ-LEG-001-004)
- AC-LEG-001-05: Given IF/THEN REQ, When linted, Then warning message contains GEARS migration URL. (maps REQ-LEG-001-005)
