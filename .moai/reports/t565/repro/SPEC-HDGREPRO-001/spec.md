---
id: SPEC-HDGREPRO-001
title: "t565 repro L - out-of-scope heading naming acceptance.md before the real AC section"
version: "0.1.0"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: Test Author
priority: P2
phase: "v3.2.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "t565, fixture"
---

# SPEC-HDGREPRO-001: out-of-scope heading naming acceptance.md before the real AC section

## 2. Scope

### Out of Scope — sibling acceptance.md inline parsing

- The sibling file is read by another path.

## 5. Requirements (EARS)

- REQ-HDGREPRO-001-001: The system SHALL read the real acceptance criteria section.

## 6. Acceptance Criteria

- AC-HDGREPRO-001-01: Given a SPEC, When it is parsed, Then the real section is read (maps REQ-HDGREPRO-001-001)
