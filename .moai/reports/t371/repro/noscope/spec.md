---
id: SPEC-REPRO-001
title: "Reproduction fixture"
version: "0.1.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: t371
priority: P1
phase: "v3.1.4 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "repro, fixture"
---

# SPEC: Reproduction fixture

## 1. Overview

Fixture used to reproduce the zero-finding short-circuit claim (D1).

## 2. Requirements

- REQ-RP-001 (Ubiquitous): The linter shall report findings.

## 3. Acceptance Criteria

- AC-RP-001: Given the fixture, when lint runs, then the output is observed.
