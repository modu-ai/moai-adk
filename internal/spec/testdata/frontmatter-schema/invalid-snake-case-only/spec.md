---
id: SPEC-TST-INV-001
title: "Invalid snake_case only fixture"
version: "0.1.0"
status: draft
created_at: 2026-05-16
updated_at: 2026-05-16
author: Test Author
priority: P1
phase: "v3.0.0 test"
module: "internal/spec/testdata"
lifecycle: spec-anchored
labels: [test, fixture, invalid]
---

# SPEC-TST-INV-001: Invalid snake_case only fixture

이 fixture는 snake_case alias (`created_at:`, `updated_at:`, `labels:`)만 사용하고
canonical field (`created:`, `updated:`, `tags:`)를 누락하여
FrontmatterSchemaRule이 정확히 3개 FrontmatterInvalid finding을 생성함을 검증한다.

## 2. Scope

### 2.1 In Scope

- Snake_case alias 사용 시 lint.go FrontmatterSchemaRule이 3개 finding 생성 확인

### 2.2 Out of Scope

- 정상 케이스 검증은 valid-12-field fixture가 담당
