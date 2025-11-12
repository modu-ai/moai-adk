---
id: INSTALLER-SEC-001
version: 0.2.0
status: completed
created: 2025-10-06
updated: 2025-10-18
labels:
  - security
  - template
  - installer
  - archived-typescript
priority: high
---


## HISTORY

### v0.2.0 (2025-10-18)
- **CHANGED**: deprecated → completed (TypeScript 프로젝트 아카이브)
- **REASON**: TypeScript 프로젝트에서 구현 완료된 기능, Python 전환으로 deprecated 처리했으나 실제로는 완료된 것으로 간주

### v0.1.0 (2025-10-16)
- **DEPRECATED**: TypeScript 프로젝트용 SPEC, Python 프로젝트에는 적용 불가
- **REASON**: MoAI-ADK가 Python 프로젝트로 전환됨에 따라 TypeScript template 보안 SPEC 불필요
- **ALTERNATIVE**: Python 프로젝트는 Jinja2 템플릿 사용, 별도 보안 메커니즘 적용

### v0.0.1 (2025-10-06)
- **INITIAL**: 템플릿 보안 검증 통합 명세 작성 (TypeScript용)
- **SCOPE**: TypeScript template-security.ts 통합

---

## Environment (환경)

### 시스템 환경
- **Language**: TypeScript 5.x
- **Runtime**: Node.js 18+
- **Package Manager**: npm/bun
- **Template Engine**: Mustache
- **Security**: template-security.ts 모듈

### 관련 파일
- `moai-adk-ts/src/core/installer/template-processor.ts` (371 LOC)
- `moai-adk-ts/src/core/installer/templates/template-security.ts` (298 LOC)
- `moai-adk-ts/templates/` (34개 템플릿 파일)

### 전제 조건
- Mustache 템플릿 엔진 사용
- TRUST 5원칙 중 Secured(보안) 준수 필요

---

## Assumptions (가정)

1. **보안 우선**: 템플릿 처리 시 보안 검증은 필수이며 성능보다 우선한다
2. **화이트리스트 방식**: 허용된 템플릿 변수만 사용 (블랙리스트 방식 금지)
3. **실패 시 중단**: 위험한 패턴 발견 시 즉시 설치 중단 (부분 성공 허용 안 함)
4. **Zero Trust**: 모든 템플릿 입력은 신뢰할 수 없다고 가정
5. **후방 호환성**: 기존 안전한 템플릿은 그대로 작동해야 함

---

## Requirements (요구사항)

### Ubiquitous Requirements (필수 기능)

1. **보안 검증 통합**
   - 시스템은 모든 템플릿 파일 처리 시 보안 검증을 수행해야 한다
   - 시스템은 sanitizeTemplateContext() 함수를 통해 컨텍스트를 정화해야 한다
   - 시스템은 validateTemplateContent() 함수를 통해 템플릿 내용을 검증해야 한다

2. **위험 패턴 탐지**
   - 시스템은 다음 패턴을 위험으로 분류해야 한다:
     - `constructor`, `prototype`, `__proto__` 접근
     - `eval()`, `Function()` 호출
     - `process`, `global`, `require` 접근
     - JavaScript 실행 URI (`javascript:`)
     - 템플릿 인젝션 시도 (`{{constructor}}`, `{{__proto__}}`)

3. **화이트리스트 검증**
   - 시스템은 ALLOWED_CONTEXT_KEYS에 정의된 변수만 허용해야 한다
   - 시스템은 허용되지 않은 변수를 자동으로 제거해야 한다

### Event-driven Requirements (이벤트 기반)

1. **템플릿 처리 시작 시**
   - WHEN 템플릿 파일 복사가 시작되면, 시스템은 보안 검증을 먼저 실행해야 한다
   - WHEN copyTemplateFile() 함수가 호출되면, 시스템은 validateTemplateContent()를 실행해야 한다

2. **위험 패턴 발견 시**
   - WHEN 템플릿에서 위험한 패턴이 발견되면, 시스템은 InstallationError를 발생시켜야 한다
   - WHEN 컨텍스트에 위험한 속성이 포함되면, 시스템은 해당 속성을 제거하고 경고를 로깅해야 한다

3. **검증 실패 시**
   - WHEN 보안 검증이 실패하면, 시스템은 템플릿 렌더링을 중단해야 한다
   - WHEN 보안 검증이 실패하면, 시스템은 상세한 에러 메시지를 제공해야 한다

### State-driven Requirements (상태 기반)

1. **검증 중 상태**
   - WHILE 템플릿 검증이 진행 중일 때, 시스템은 다른 템플릿 처리를 차단해야 한다
   - WHILE 보안 검증이 활성화된 상태일 때, 시스템은 모든 변수를 화이트리스트로 필터링해야 한다

2. **안전한 상태**
   - WHILE 템플릿이 검증을 통과한 상태일 때, 시스템은 Mustache 렌더링을 허용해야 한다
   - WHILE 컨텍스트가 정화된 상태일 때, 시스템은 읽기 전용 객체로 동결해야 한다

### Optional Features (선택 기능)

1. **보안 로깅**
   - WHERE 위험한 패턴이 탐지되면, 시스템은 상세한 보안 로그를 기록할 수 있다
   - WHERE 변수가 화이트리스트에서 제거되면, 시스템은 경고 메시지를 출력할 수 있다

2. **성능 최적화**
   - WHERE 동일한 템플릿을 반복 처리하면, 시스템은 검증 결과를 캐싱할 수 있다
   - WHERE 텍스트가 아닌 파일(이미지 등)을 처리하면, 시스템은 보안 검증을 건너뛸 수 있다

### Constraints (제약사항)

1. **성능 제약**
   - IF 템플릿 파일 크기가 1MB를 초과하면, 시스템은 경고를 발생시켜야 한다
   - IF 보안 검증이 5초를 초과하면, 시스템은 타임아웃 에러를 발생시켜야 한다

2. **보안 제약**
   - IF 위험한 패턴이 발견되면, 시스템은 즉시 중단해야 하며 부분 렌더링을 허용하지 않아야 한다
   - IF DANGEROUS_PROPERTIES에 포함된 속성이 사용되면, 시스템은 무조건 거부해야 한다

3. **코드 품질 제약**
   - template-processor.ts는 보안 통합 후에도 300 LOC를 초과하지 않아야 한다
   - 보안 검증 함수는 50 LOC를 초과하지 않아야 한다
   - 테스트 커버리지는 85% 이상이어야 한다

---

## Technical Design (기술 설계)

### 통합 방식

```typescript

import {
  sanitizeTemplateContext,
  validateTemplateContent
} from './templates/template-security';

export class TemplateProcessor {
  async copyTemplateFile(
    srcPath: string,
    dstPath: string,
    variables: Record<string, any>
  ): Promise<void> {
    // 1. 템플릿 내용 읽기
    const content = await fs.promises.readFile(srcPath, 'utf-8');

    // 2. 보안 검증 (NEW)
    if (!validateTemplateContent(content)) {
      throw new InstallationError(
        `Template contains dangerous patterns: ${srcPath}`,
        { phase: 'TEMPLATE_SECURITY', filePath: srcPath }
      );
    }

    // 3. 컨텍스트 정화 (NEW)
    const { sanitizedContext, warnings } = sanitizeTemplateContext(variables);

    // 4. 경고 로깅 (NEW)
    if (warnings.length > 0) {
      logger.warn('Template context sanitized', {
        srcPath,
        warnings,
        tag: 'WARN:TEMPLATE-SECURITY-001'
      });
    }

    // 5. 안전한 렌더링
    const processedContent = mustache.render(content, sanitizedContext);

    // 6. 파일 쓰기
    await fs.promises.writeFile(dstPath, processedContent);
  }
}
```

### 에러 처리

```typescript
// 보안 검증 실패 시 상세 에러 메시지
throw new InstallationError(
  'Template security validation failed',
  {
    phase: 'TEMPLATE_SECURITY',
    filePath: srcPath,
    violations: [
      'Dangerous pattern detected: {{constructor}}',
      'Forbidden property: __proto__'
    ],
    recommendation: 'Please review template file for security issues'
  }
);
```

---


- **TEST**: `moai-adk-ts/src/core/installer/__tests__/template-security.test.ts`
- **CODE**:
  - `moai-adk-ts/src/core/installer/template-processor.ts` (보안 통합)
  - `moai-adk-ts/src/core/installer/templates/template-security.ts` (보안 함수)
- **DOC**: `.moai/specs/SPEC-INSTALLER-SEC-001/`

---

## Success Criteria (성공 기준)

### 기능 완성도
- [ ] sanitizeTemplateContext() 함수가 template-processor.ts에서 사용됨
- [ ] validateTemplateContent() 함수가 template-processor.ts에서 사용됨
- [ ] 위험한 패턴 발견 시 InstallationError 발생
- [ ] 화이트리스트에 없는 변수 자동 제거
- [ ] 보안 경고 로깅 구현

### 품질 기준
- [ ] 테스트 커버리지 85% 이상
- [ ] template-processor.ts LOC ≤ 300
- [ ] 모든 보안 함수 LOC ≤ 50
- [ ] TRUST 5원칙 준수 (특히 Secured)

### 성능 기준
- [ ] 보안 검증 시간 < 100ms (평균)
- [ ] 보안 검증 타임아웃 5초 이하
- [ ] 메모리 사용량 증가 < 10MB

### 문서화
- [ ] 보안 정책 문서화
- [ ] 화이트리스트 변수 목록 문서화
- [ ] 위험 패턴 목록 문서화
- [ ] 에러 처리 가이드 문서화

---

## Dependencies (의존성)

### 기술 의존성
- Mustache 템플릿 엔진
- Winston Logger
- Node.js fs/promises API

### SPEC 의존성
- **독립적**: 다른 SPEC과 의존성 없음 (우선 진행 가능)

### 선행 조건
- template-security.ts 모듈 존재
- InstallationError 클래스 정의

---

## Risk Analysis (리스크 분석)

### 높은 리스크
1. **후방 호환성 깨짐**: 기존 템플릿이 보안 검증 실패 가능
   - **완화**: 기존 템플릿 전수 검사 및 화이트리스트 조정

2. **성능 저하**: 보안 검증으로 설치 시간 증가
   - **완화**: 경량 패턴 매칭, 캐싱 적용

### 중간 리스크
1. **False Positive**: 안전한 템플릿이 위험으로 분류
   - **완화**: 테스트 케이스 충분히 작성, 화이트리스트 확장

2. **보안 우회**: 새로운 공격 패턴 미탐지
   - **완화**: 정기적인 보안 패턴 업데이트

---

## Implementation Notes (구현 참고사항)

### 단계별 구현
1. **Phase 1**: sanitizeTemplateContext() 통합
2. **Phase 2**: validateTemplateContent() 통합
3. **Phase 3**: 에러 처리 및 로깅
4. **Phase 4**: 테스트 작성 (85% 커버리지)
5. **Phase 5**: 문서화 및 검토

### 테스트 전략
- 위험한 패턴 샘플 30개 이상 테스트
- 안전한 템플릿 정상 작동 검증
- 화이트리스트 경계값 테스트
- 성능 벤치마크 (1000개 템플릿 처리)

### 코드 리뷰 포인트
- 보안 검증 누락 케이스 확인
- 에러 메시지 명확성 검토
- 성능 병목 구간 확인
- 로깅 레벨 적절성 검토
