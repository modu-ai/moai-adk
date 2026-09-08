---
id: SPEC-T528-DUPFIX-001
title: "t528 positive control — deliberate duplicate AC id"
version: "0.1.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-develop
priority: P2
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "t528, fixture, positive-control"
---

# Fixture — deliberate duplicate AC id

This file exists so the corpus-wide "hard errors == 0" measurement means
something. Without a case where the measurement DOES fire, a zero is equally
consistent with "the corpus holds no material" and "the measurement never ran",
and AC-ACA-001-015 requires the detector be shown firing FIRST.

It lives under `.moai/reports/`, not `.moai/specs/`, so it is outside the
measured corpus and cannot move the figures it exists to validate.

## Acceptance Criteria

- AC-DUPFIX-001: first declaration carrying this id.
- AC-DUPFIX-001: SECOND declaration carrying the SAME id — the deliberate duplicate.
