---
name: "SPEC-First TDD 개발 오케스트레이션"
description: "컨텍스트 엔지니어링, TRUST 원칙 및 @TAG 추적성으로 에이전트를 SPEC-First TDD 워크플로우 안내. /alfred:1-plan, /alfred:2-run, /alfred:3-sync 명령어에 필수적. EARS 요구사항, JIT 컨텍스트 로딩, TDD RED-GREEN-REFACTOR 사이클, TAG 체인 검증 포함."
allowed-tools: "Read, Bash(rg:*), Bash(grep:*)"
---

# Alfred 개발 가이드 스킬

## 핵심 워크플로우: SPEC → TEST → CODE → DOC

**스펙 없이 코드 없음. 테스트 없이 구현 없음.**

### 1. 단계 (`/alfred:1-plan`)
- 먼저 `@SPEC:ID` 태그로 상세한 명세 작성
- EARS 형식 사용 (5가지 패턴: 보편적, 이벤트 중심, 상태 중심, 선택적, 제약)
- `.moai/specs/SPEC-{ID}/spec.md`에 저장

### 2. TDD 단계 (`/alfred:2-run`)
- **RED**: `@TEST:ID` 태그로 실패하는 테스트 작성
- **GREEN**: `@CODE:ID` 태그로 최소 코드 구현
- **REFACTOR**: SPEC 준수 유지하며 코드 개선
- `@DOC:ID` 태그로 문서화

### 3. 동기화 단계 (`/alfred:3-sync`)
- @TAG 체인 무결성 검증 (SPEC→TEST→CODE→DOC)
- 구현과 문서 동기화
- 동기화 보고서 생성

## 주요 원칙

**컨텍스트 엔지니어링**: 각 단계에서 필요한 문서만 로드
- `/alfred:1-plan` → product.md, structure.md, tech.md
- `/alfred:2-run` → SPEC-{ID}/spec.md, development-guide.md
- `/alfred:3-sync` → sync-report.md, TAG 검증

**TRUST 5 기둥**:
1. **T** – 테스트 주도 (RED→GREEN→REFACTOR)
2. **R** – 가독성 (명확한 명명, 문서화)
3. **U** – 통일성 (일관된 패턴, 언어)
4. **S** – 보안 (OWASP 준수, 보안 검토)
5. **E** – 평가 (메트릭, 커버리지 ≥85%)

## 일반적인 패턴

| 시나리오 | 행동 |
|----------|--------|
| 테스트 먼저 작성 | RED 단계: @TEST:ID로 실패하는 테스트 |
| 기능 구현 | GREEN 단계: @CODE:ID로 최소 코드 |
| 안전하게 리팩토링 | REFACTOR 단계: 코드 구조 개선 |
| 변경사항 추적 | 항상 코드 + 문서에서 @TAG 시스템 사용 |
| 완성 검증 | `/alfred:3-sync`가 모든 링크 확인 |