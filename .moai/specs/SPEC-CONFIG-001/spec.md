---
id: CONFIG-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @goos
---

# @SPEC:CONFIG-001: config.json 스키마 통합 및 표준화

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: config.json 템플릿과 TypeScript 인터페이스 스키마 통합
- **AUTHOR**: @goos
- **REVIEW**: Alfred (자동 검증)
- **STATUS**: Ready (구현 완료)

## 개요

MoAI-ADK의 config.json 템플릿과 TypeScript 인터페이스 간의 스키마 불일치 문제를 해결하고, MoAI-ADK 핵심 철학(TRUST 원칙, CODE-FIRST)을 config 구조에 명시적으로 반영합니다.

## 문제 정의

### 발견된 문제

1. **스키마 불일치 (Critical)**
   - 3가지 다른 MoAIConfig 정의 공존:
     - `templates/.moai/config.json` (템플릿)
     - `src/core/config/types.ts` (TypeScript 인터페이스)
     - `src/cli/config/config-builder.ts` (대화형 빌더)

2. **locale 필드 누락 (High)**
   - CLI (`src/cli/index.ts:60`)에서 `config.locale` 읽으려 시도
   - 템플릿에는 필드 없음

3. **MoAI-ADK 철학 묵시적 (Medium)**
   - TRUST 원칙, CODE-FIRST 철학이 코드에만 존재
   - config 자체가 자기 문서화되지 않음

## EARS 요구사항

### UR-001: 스키마 통합
**Ubiquitous Requirement**

시스템은 config.json 템플릿과 TypeScript 인터페이스가 100% 일치하는 단일 스키마를 제공해야 한다.

**Acceptance Criteria**:
- [ ] 템플릿 JSON 구조 ≡ TypeScript MoAIConfig 인터페이스
- [ ] 모든 필드 타입 일치
- [ ] TypeScript 컴파일 에러 0개

### UR-002: locale 필드 지원
**Ubiquitous Requirement**

시스템은 config.json에 locale 필드를 포함하여 다국어 CLI 지원을 표준화해야 한다.

**Acceptance Criteria**:
- [ ] `project.locale` 필드 존재 (값: 'ko' | 'en')
- [ ] CLI에서 config.locale 읽기 성공
- [ ] 기본값: 'ko'

### UR-003: MoAI-ADK 철학 명시
**Ubiquitous Requirement**

시스템은 config.json 구조 내에 MoAI-ADK 핵심 철학을 명시적으로 포함해야 한다.

**Acceptance Criteria**:
- [ ] `constitution` 섹션: TRUST 원칙 반영
- [ ] `tags.code_scan_policy.philosophy`: CODE-FIRST 원칙 명시
- [ ] `git_strategy`: Personal/Team 모드 전략 명시

### ER-001: 템플릿 중심 통합
**Event-Driven Requirement**

WHEN 스키마 통합 전략을 선택할 때,
시스템은 템플릿 JSON 구조를 유지하고 TypeScript 인터페이스를 템플릿에 맞춰야 한다.

**Rationale**:
- 템플릿이 실제 운영 코드 (`template-processor.ts`)와 일치
- MoAI-ADK 철학이 템플릿에 더 풍부하게 반영됨
- 변경 영향도 최소화

**Acceptance Criteria**:
- [ ] `src/core/config/types.ts` 수정 (템플릿 구조 반영)
- [ ] `src/core/config/builders/moai-config-builder.ts` 수정
- [ ] `src/core/project/template-processor.ts` 수정
- [ ] 템플릿 파일 최소 변경 (locale 필드만 추가)

### ER-002: 자기 문서화 config
**Event-Driven Requirement**

WHEN config.json이 생성될 때,
시스템은 MoAI-ADK 철학과 원칙을 config 내부에 설명하는 필드를 포함해야 한다.

**Acceptance Criteria**:
- [ ] `constitution.principles.simplicity.notes`: 설명 포함
- [ ] `tags.code_scan_policy.philosophy`: 철학 문장 포함
- [ ] 사용자가 config만 읽어도 MoAI-ADK 이해 가능

## 기술 설계

### 통합 스키마 구조

```typescript
// src/core/config/types.ts
export interface MoAIConfig {
  _meta?: {
    '@CODE:CONFIG-STRUCTURE-001'?: string;
    '@SPEC:PROJECT-CONFIG-001'?: string;
  };

  project: {
    name: string;
    version: string;
    mode: 'personal' | 'team';
    description?: string;
    initialized: boolean;
    created_at: string;
    locale?: 'ko' | 'en';  // ← 추가
  };

  constitution: {
    enforce_tdd: boolean;
    require_tags: boolean;
    test_coverage_target: number;
    simplicity_threshold: number;
    principles: {
      simplicity: {
        max_projects: number;
        notes: string;  // ← 자기 문서화
      };
    };
  };

  git_strategy: {
    personal: { ... };
    team: { ... };
  };

  tags: {
    auto_sync: boolean;
    storage_type: 'code_scan';
    categories: string[];
    code_scan_policy: {
      no_intermediate_cache: boolean;
      realtime_validation: boolean;
      scan_tools: string[];
      scan_command: string;
      philosophy: string;  // ← CODE-FIRST 철학
    };
  };

  pipeline: {
    available_commands: string[];
    current_stage: string;
  };
}
```

### 구현 계획

#### Phase 1: locale 필드 추가
**파일**: `templates/.moai/config.json`
```json
"project": {
  ...
  "locale": "ko"
}
```

#### Phase 2: types.ts 통합
**파일**: `src/core/config/types.ts`
- MoAIConfig 인터페이스 전면 재정의
- 템플릿 구조와 100% 일치

#### Phase 3: 빌더 통합
**파일**:
- `src/core/config/builders/moai-config-builder.ts`
- `src/core/project/template-processor.ts`

빌더 로직을 새로운 인터페이스에 맞춰 수정

#### Phase 4: 테스트 수정
**파일**:
- `src/__tests__/core/project/template-processor.test.ts`
- `src/core/config/__tests__/config-manager.test.ts`

새로운 구조에 맞춰 테스트 업데이트

## 검증 기준

### 자동 검증
- [ ] `npm run type-check` 통과 (에러 0개)
- [ ] `npm run lint` 통과 (경고 0개)
- [ ] 모든 테스트 통과

### 수동 검증
- [ ] 템플릿 JSON ≡ TypeScript 인터페이스 (필드별 비교)
- [ ] `moai init` 실행 시 locale 필드 포함된 config 생성
- [ ] CLI 실행 시 config.locale 정상 읽기

### 문서화
- [ ] CHANGELOG.md 업데이트
- [ ] 분석 보고서 작성 (`.moai/reports/config-template-analysis.md`)
- [ ] 완료 요약 보고서 작성

## 영향 범위

### 수정 파일 (6개)
1. `templates/.moai/config.json`
2. `src/core/config/types.ts`
3. `src/core/config/builders/moai-config-builder.ts`
4. `src/core/project/template-processor.ts`
5. `src/__tests__/core/project/template-processor.test.ts`
6. `src/core/config/__tests__/config-manager.test.ts`

### 하위 호환성
- ✅ **유지**: 기존 config 파일도 정상 동작
- ✅ **선택적 필드**: locale은 옵션 (없어도 동작)
- ✅ **점진적 적용**: 새 프로젝트부터 자동 적용

## 위험 요소 및 대응

### R1: 기존 사용자 config 호환성
**위험**: 구버전 config와 신버전 인터페이스 불일치

**대응**:
- locale 필드는 옵션 (`locale?: 'ko' | 'en'`)
- 누락 시 기본값 'ko' 사용
- 런타임 에러 없음

### R2: 테스트 커버리지 저하
**위험**: 새 구조로 인한 테스트 실패

**대응**:
- Phase 4에서 모든 테스트 수정
- 검증 단계에서 100% 통과 확인

## 성공 지표

| 지표 | 목표 | 실제 |
|------|------|------|
| 스키마 통합 | 100% | 100% ✅ |
| 타입 안전성 | 에러 0개 | 0개 ✅ |
| 테스트 통과 | 100% | 100% ✅ |
| 하위 호환성 | 유지 | 유지 ✅ |

## 참고 자료

- `.moai/reports/config-template-analysis.md` - 상세 분석 보고서
- `.moai/reports/config-schema-integration-summary.md` - 완료 요약
- `CHANGELOG.md` - v0.0.3 릴리스 노트

## Related SPECs

- None (신규 작업)

## Tags

- @SPEC:CONFIG-001
- @CODE:CONFIG-STRUCTURE-001
- @DOC:CONFIG-001
