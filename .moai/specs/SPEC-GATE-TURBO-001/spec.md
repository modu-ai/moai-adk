---
id: SPEC-GATE-TURBO-001
title: "Turbo-safe Node test command in quality gate"
status: draft
priority: P1
created: "2026-05-29"
version: 0.1.0
author: eren-cupist
module: "quality"
tags: "quality-gate, turborepo, node, bugfix"
---

# SPEC-GATE-TURBO-001: Turbo-safe Node test command in quality gate

## Overview

품질 게이트는 `package.json`이 있는 모든 프로젝트에서 `npm test -- --passWithNoTests`를 실행한다. `--passWithNoTests`는 vitest/jest 전용 플래그로, Turborepo 모노레포에서 루트 `test` 스크립트가 보통 `turbo run test`이기 때문에 npm이 이 플래그를 turbo에 전달하면 turbo가 `ERROR unexpected argument '--passWithNoTests' found`로 거부한다. 그 결과 turbo 모노레포에서는 모든 `git commit`이 게이트에 의해 차단되며, 테스트 스텝 명령 자체가 무효이므로 우회 수단도 없다.

이 SPEC은 이미 존재하는 `resolveDartFlutter` 패턴(Dart vs Flutter 변형)을 그대로 따라, Node.js 툴체인이 매칭된 프로젝트 디렉터리에 `turbo.json`이 있을 때 turbo-safe한 테스트 명령(`npm test`, 플래그 없음)을 사용하는 변형 툴체인을 반환하도록 한다. 패키지 레벨 `toolchains` 슬라이스는 변형하지 않는다.

## Requirements (EARS Format)

### REQ-GATE-TURBO-001 (State-Driven)
When a project has both `package.json` and `turbo.json` in the project dir, the quality gate test step SHALL invoke a turbo-safe command that does NOT pass `--passWithNoTests` (i.e. `npm test` with `args: []string{"test"}`).

### REQ-GATE-TURBO-002 (State-Driven)
When a Node.js project has `package.json` but NO `turbo.json` in the project dir, the test step SHALL remain `npm test -- --passWithNoTests` (unchanged, no regression).

### REQ-GATE-TURBO-003 (Ubiquitous)
The turbo detection SHALL NOT mutate the package-level `toolchains` slice; it SHALL return a new toolchain variant, mirroring `resolveDartFlutter`.

### REQ-GATE-TURBO-004 (Ubiquitous)
The Node.js variant SHALL keep the existing eslint `lintSteps` unchanged.

### REQ-GATE-TURBO-005 (Ubiquitous)
The fix SHALL leave all other language toolchains (Go, Python, Rust, etc.) unchanged.

### REQ-GATE-TURBO-006 (Ubiquitous)
Unit tests SHALL cover both the turbo-present path and the turbo-absent path.

## Architecture

```
detectToolchain() matches Node.js toolchain by package.json marker
  → resolveNodeTurbo(&toolchains[i], dir)
    → turbo.json present in dir?
       YES → return Node variant: testStep args = []string{"test"} (turbo-safe), lintSteps unchanged
       NO  → return tc unchanged (npm test -- --passWithNoTests)
```

`resolveNodeTurbo` is the Node analogue of `resolveDartFlutter` (gate.go:325-340): both match a toolchain by marker, inspect a project-specific file, and return a variant without mutating the shared `toolchains` slice.

## Implementation Scope

### Modified Files
- `internal/hook/quality/gate.go` — add `resolveNodeTurbo(tc, dir)`; call it from `detectToolchain()` on the Node.js toolchain match (analogous to the existing `resolveDartFlutter` call).
- `internal/hook/quality/gate_test.go` — add tests: (1) `turbo.json` present → Node `testStep` args contain no `--passWithNoTests`; (2) `turbo.json` absent → `testStep` unchanged with the flag.

### Out of Scope
- Modifying any non-Node toolchain entry.
- Changing the eslint lint step.

## Non-Goals
- Replacing the test command with `turbo run test --affected`. The minimal, behavior-preserving fix is to drop `--passWithNoTests` and let turbo decide what to run via the project's own root `test` script. `--affected` is noted here as a possible future enhancement, not part of this SPEC.
- Generalized per-monorepo-tool detection (nx, lerna, pnpm workspaces). Only Turborepo (`turbo.json`) is in scope.
- Dynamically rewriting arbitrary user `test` scripts.
