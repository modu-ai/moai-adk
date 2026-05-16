---
id: SPEC-TST-VLD-001
title: "Valid 12-field canonical fixture"
version: "0.1.0"
status: draft
created: 2026-05-16
updated: 2026-05-16
author: Test Author
priority: P1
phase: "v3.0.0 test"
module: "internal/spec/testdata"
lifecycle: spec-anchored
tags: "test, fixture, valid"
---

# SPEC-TST-VLD-001: Valid 12-field canonical fixture

이 fixture는 lint.go FrontmatterSchemaRule이 요구하는 12개 canonical field를
모두 보유하여 FrontmatterInvalid finding이 0건임을 검증한다.

## 2. Scope

### 2.1 In Scope

- 12-field canonical frontmatter 사용 시 lint.go FrontmatterSchemaRule이 finding 0건 확인

### 2.2 Out of Scope

- snake_case alias 케이스는 invalid-snake-case-only fixture가 담당
