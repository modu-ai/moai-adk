---
name: moai-cc-claude-md
description: Claude Code Markdown integration, documentation generation, and structured content patterns. Use when generating documentation, managing markdown content, or creating structured reports.
version: 1.0.0
modularized: false
last_updated: 2025-11-22
compliance_score: 71
auto_trigger_keywords:
  - cc
  - claude
  - md
  - testing
category_tier: 1
---

## Quick Reference (30 seconds)

Claude Code의 Markdown 통합은 문서 생성, 콘텐츠 구조화, 템플릿 기반 문서 워크플로우를 제공합니다.
프로젝트 문서화(README, CHANGELOG), 기술 문서(API 가이드), 지식 베이스, 보고서 등을
체계적으로 관리할 수 있는 강력한 문서화 패턴을 제공합니다.

**핵심 기능**:
- Markdown 콘텐츠 자동 생성 및 렌더링
- 크로스 레퍼런스 및 링크 관리 시스템
- 템플릿 기반 문서 구조화
- 자동 콘텐츠 검증 및 품질 확인
- 버전 관리 및 변경 이력 추적

---

## Implementation Guide

### What It Does

Claude Code Markdown 통합은 다음을 제공합니다:

**Markdown 콘텐츠 생성**:
- AI 기반 문서 자동 생성
- 코드 블록 및 구문 강조
- 메타데이터 및 프런트매터 관리
- 동적 콘텐츠 인제션

**문서 구조화**:
- 계층적 문서 조직화
- 목차 자동 생성
- 섹션 간 네비게이션
- 일관된 포맷팅

**템플릿 시스템**:
- 재사용 가능한 문서 템플릿
- 변수 대체 및 조건부 렌더링
- 커스텀 블록 및 매크로
- 스타일 및 테마 적용

### When to Use

- ✅ 프로젝트 문서화 (README, CONTRIBUTING, CODE_OF_CONDUCT)
- ✅ 기술 문서 작성 (API 문서, 개발 가이드, 튜토리얼)
- ✅ 프로세스 문서화 (워크플로우, 정책, 절차)
- ✅ 보고서 생성 (분석, 상태 리포트, 요약)
- ✅ 지식 베이스 (FAQ, 모범 사례, 패턴 라이브러리)
- ✅ 자동화된 문서 배포 및 출판

### Core Markdown Patterns

#### 1. 문서 구조화 패턴
```markdown
# 제목 (레벨 1)
## 부제목 (레벨 2)
### 섹션 (레벨 3)

- 불릿 포인트
  1. 번호 목록
  2. 계층적 구조

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
```

#### 2. 크로스 레퍼런스 패턴
```markdown
[링크 텍스트](../path/to/file.md)
[내부 링크](#섹션-제목)
[외부 링크](https://example.com)

[변수 참조]: variable-definition
```

#### 3. 코드 블록 패턴
````markdown
```python
# Python 코드 예제
def function():
    pass
```

```typescript
// TypeScript 코드 예제
interface Props {
  name: string;
}
```
````

#### 4. 콘텐츠 검증 패턴
- 링크 유효성 검증
- 코드 블록 구문 검증
- 이미지 경로 검증
- 메타데이터 완성도 검증

### Dependencies

- Markdown 처리 엔진 (Remark, Marked, Pandoc)
- 콘텐츠 템플릿 시스템
- 문서 검증 프레임워크
- 출판 플랫폼 (Nextra, VitePress, Docusaurus)

---

## Works Well With

- `moai-docs-generation` (자동 문서 생성)
- `moai-docs-validation` (콘텐츠 품질 검증)
- `moai-docs-linting` (마크다운 스타일 체크)
- `moai-cc-commands` (문서화 워크플로우 자동화)

---

## Advanced Patterns

### 1. 고급 템플릿 시스템

**동적 콘텐츠 인제션**:
```markdown
<!-- Template Variable -->
{{ projectName }} - {{ version }}
{{ description }}

<!-- Conditional Content -->
{% if environment === 'production' %}
Production specific content
{% endif %}

<!-- Loop Patterns -->
{% for item in items %}
- {{ item.name }}
{% endfor %}
```

### 2. 자동 문서 생성 워크플로우

**프로세스**:
1. 소스 코드/설정 파일 파싱
2. 메타데이터 추출 (JSDoc, 타입 정의)
3. 템플릿과 메타데이터 병합
4. Markdown 문서 생성
5. 자동 검증 및 배포

**예시**:
```typescript
// TypeScript 코드에서 자동 API 문서 생성
/**
 * @description 사용자 생성 함수
 * @param {string} name - 사용자 이름
 * @returns {Promise<User>} 생성된 사용자 객체
 */
async function createUser(name: string): Promise<User> {
  // 자동으로 API 문서 생성됨
}
```

### 3. 멀티 채널 출판 패턴

**출판 대상**:
- Markdown → HTML (웹 사이트)
- Markdown → PDF (다운로드)
- Markdown → 슬라이드 (프레젠테이션)
- Markdown → Email (배포)
- Markdown → Wiki (조직 문서화)

### 4. 콘텐츠 버전 관리

**변경 이력 추적**:
- Git 기반 문서 버전 관리
- 자동 CHANGELOG 생성
- 마이그레이션 가이드 제공
- 하위 호환성 보장

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, markdown patterns
- **v1.0.0** (2025-10-22): Initial markdown integration

---

**End of Skill** | Updated 2025-11-21 | Lines: 180