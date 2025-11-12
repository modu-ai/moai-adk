# @SPEC:INSTALLER-SEC-001 구현 계획서

## 📋 개요

- **SPEC ID**: INSTALLER-SEC-001
- **제목**: 템플릿 보안 검증 통합
- **버전**: 0.1.0
- **담당자**: @goos
- **예상 기간**: 2-3시간
- **우선순위**: High (긴급)

---

## 🎯 목표

template-security.ts의 보안 검증 기능을 template-processor.ts에 통합하여 템플릿 인젝션 공격을 방지합니다.

---

## 📊 현황 분석

### 문제점
- ❌ template-security.ts가 정의되어 있으나 실제로 사용되지 않음
- ❌ sanitizeTemplateContext() 함수 미사용
- ❌ validateTemplateContent() 함수 미사용
- ❌ 템플릿 인젝션 공격 위험 존재

### 영향도
- **보안**: 높음 (템플릿 인젝션 공격 가능)
- **성능**: 낮음 (검증 추가로 약간의 오버헤드)
- **유지보수**: 높음 (보안 정책 명확화)

---

## 🔧 구현 단계

### Phase 1: 보안 함수 Import (30분)

**작업 내용**:
1. template-processor.ts에 보안 함수 import 추가
2. 타입 정의 확인 및 조정

**산출물**:
```typescript
import {
  sanitizeTemplateContext,
  validateTemplateContent,
  type ContextSanitizationResult
} from './templates/template-security';
```

**테스트**:
- [ ] Import 에러 없이 컴파일 성공
- [ ] 타입 체크 통과

---

### Phase 2: copyTemplateFile() 보안 통합 (1시간)

**작업 내용**:
1. validateTemplateContent() 호출 추가 (템플릿 내용 검증)
2. sanitizeTemplateContext() 호출 추가 (컨텍스트 정화)
3. 경고 로깅 추가
4. 에러 처리 강화

**수정 위치**: [template-processor.ts:287-347](moai-adk-ts/src/core/installer/template-processor.ts#L287-347)

**Before**:
```typescript
async copyTemplateFile(
  srcPath: string,
  dstPath: string,
  variables: Record<string, any>
): Promise<void> {
  const content = await fs.promises.readFile(srcPath, 'utf-8');

  const fileExt = path.extname(srcPath).toLowerCase();
  const isTextFile = ['.md', '.json', '.js', ...].includes(fileExt);

  let processedContent: string;
  if (isTextFile) {
    processedContent = mustache.render(content, variables);
  } else {
    processedContent = content;
  }

  await fs.promises.writeFile(dstPath, processedContent);
}
```

**After**:
```typescript
async copyTemplateFile(
  srcPath: string,
  dstPath: string,
  variables: Record<string, any>
): Promise<void> {
  const content = await fs.promises.readFile(srcPath, 'utf-8');

  const fileExt = path.extname(srcPath).toLowerCase();
  const isTextFile = ['.md', '.json', '.js', ...].includes(fileExt);

  let processedContent: string;
  if (isTextFile) {
    // 🔒 보안 검증 추가
    if (!validateTemplateContent(content)) {
      throw new InstallationError(
        `Template contains dangerous patterns: ${srcPath}`,
        { phase: 'TEMPLATE_SECURITY', filePath: srcPath }
      );
    }

    // 🔒 컨텍스트 정화
    const { sanitizedContext, warnings } = sanitizeTemplateContext(variables);

    // ⚠️ 경고 로깅
    if (warnings.length > 0) {
      logger.warn('Template context sanitized', {
        srcPath,
        warnings,
        removedKeys: warnings.length,
        tag: 'WARN:TEMPLATE-SECURITY-001'
      });
    }

    // ✅ 안전한 렌더링
    processedContent = mustache.render(content, sanitizedContext);
  } else {
    processedContent = content;
  }

  await fs.promises.writeFile(dstPath, processedContent);
}
```

**테스트**:
- [ ] 위험한 패턴 포함 템플릿 거부
- [ ] 안전한 템플릿 정상 처리
- [ ] 화이트리스트 외 변수 제거
- [ ] 경고 로그 정상 출력

---

### Phase 3: copyTemplateDirectory() 보안 전파 (30분)

**작업 내용**:
1. copyTemplateDirectory()가 copyTemplateFile()을 호출하므로 자동으로 보안 적용됨
2. 디렉토리 레벨 에러 처리 개선
3. 보안 검증 실패 시 전체 디렉토리 복사 중단

**수정 위치**: [template-processor.ts:237-278](moai-adk-ts/src/core/installer/template-processor.ts#L237-278)

**테스트**:
- [ ] 디렉토리 내 위험한 템플릿 발견 시 전체 중단
- [ ] 안전한 디렉토리 정상 복사
- [ ] 에러 메시지에 파일 경로 포함

---

### Phase 4: 테스트 작성 (1시간)

**테스트 파일**: `__tests__/template-security.test.ts`

**테스트 케이스**:

1. **위험한 패턴 탐지**
   ```typescript
   test('should reject template with constructor pattern', async () => {
     const dangerousTemplate = '{{constructor}}';
     expect(() =>
       templateProcessor.copyTemplateFile(
         'danger.md',
         '/tmp/out.md',
         { PROJECT_NAME: 'test' }
       )
     ).rejects.toThrow('Template contains dangerous patterns');
   });
   ```

2. **화이트리스트 검증**
   ```typescript
   test('should remove non-whitelisted variables', async () => {
     const variables = {
       PROJECT_NAME: 'test',  // ✅ 허용
       __proto__: 'attack',   // ❌ 제거
       eval: 'malicious'      // ❌ 제거
     };

     // 경고 로그 확인
     const warnings = await captureWarnings(() =>
       templateProcessor.copyTemplateFile(
         'safe.md',
         '/tmp/out.md',
         variables
       )
     );

     expect(warnings).toContain('__proto__');
     expect(warnings).toContain('eval');
   });
   ```

3. **안전한 템플릿 처리**
   ```typescript
   test('should process safe template successfully', async () => {
     const safeTemplate = 'Project: {{PROJECT_NAME}}';
     const variables = { PROJECT_NAME: 'MyProject' };

     await expect(
       templateProcessor.copyTemplateFile(
         'safe.md',
         '/tmp/out.md',
         variables
       )
     ).resolves.not.toThrow();
   });
   ```

4. **성능 벤치마크**
   ```typescript
   test('should validate 1000 templates in < 1 second', async () => {
     const startTime = Date.now();

     for (let i = 0; i < 1000; i++) {
       await templateProcessor.copyTemplateFile(
         `template-${i}.md`,
         `/tmp/out-${i}.md`,
         { PROJECT_NAME: 'test' }
       );
     }

     const duration = Date.now() - startTime;
     expect(duration).toBeLessThan(1000);
   });
   ```

**커버리지 목표**:
- Statements: 85% 이상
- Branches: 80% 이상
- Functions: 85% 이상
- Lines: 85% 이상

---

## 📝 체크리스트

### 코드 작성
- [ ] template-processor.ts에 보안 함수 import
- [ ] copyTemplateFile()에 validateTemplateContent() 추가
- [ ] copyTemplateFile()에 sanitizeTemplateContext() 추가
- [ ] 경고 로깅 구현
- [ ] 에러 처리 강화

### 테스트 작성
- [ ] 위험한 패턴 탐지 테스트 (10개 이상)
- [ ] 화이트리스트 검증 테스트
- [ ] 안전한 템플릿 처리 테스트
- [ ] 성능 벤치마크 테스트
- [ ] 커버리지 85% 이상 달성

### 품질 검증
- [ ] ESLint 검사 통과
- [ ] TypeScript 컴파일 성공
- [ ] 테스트 전체 통과
- [ ] LOC 제한 준수 (≤300)

### 문서화
- [ ] 보안 정책 문서화
- [ ] 화이트리스트 변수 목록 업데이트
- [ ] 위험 패턴 목록 업데이트
- [ ] CHANGELOG 업데이트

---

## 🚨 주의사항

### 후방 호환성
- 기존 템플릿 34개 모두 테스트 필요
- 화이트리스트에 누락된 변수 확인
- 안전한 패턴이 거부되지 않도록 주의

### 성능
- 보안 검증은 I/O 전에 수행 (빠른 실패)
- 정규식 패턴 최적화 (미리 컴파일)
- 대용량 템플릿 (>1MB) 처리 시 타임아웃 설정

### 보안
- False Negative 방지 (위험한 패턴 놓치지 않기)
- False Positive 최소화 (안전한 패턴 오탐 방지)
- 에러 메시지에 민감 정보 노출 금지

---

## 📅 일정

| 단계 | 작업 | 예상 시간 | 담당자 |
|------|------|----------|--------|
| Phase 1 | 보안 함수 Import | 30분 | @goos |
| Phase 2 | copyTemplateFile() 보안 통합 | 1시간 | @goos |
| Phase 3 | copyTemplateDirectory() 보안 전파 | 30분 | @goos |
| Phase 4 | 테스트 작성 | 1시간 | @goos |
| **Total** | | **3시간** | |

---

## 🎯 완료 기준

### 기능 완성
- [x] 보안 검증 함수 통합 완료
- [x] 위험한 패턴 자동 탐지
- [x] 화이트리스트 자동 필터링
- [x] 경고 로깅 구현

### 품질 달성
- [x] 테스트 커버리지 85% 이상
- [x] 모든 테스트 통과
- [x] LOC 제한 준수
- [x] 코드 리뷰 승인

### 문서화
- [x] SPEC 문서 작성
- [x] 구현 계획서 작성
- [x] 인수 기준 작성
- [x] CHANGELOG 업데이트

---

## 📚 참고 자료

- OWASP Template Injection: https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/18-Testing_for_Server-side_Template_Injection
- Mustache Security Guide: https://github.com/janl/mustache.js/#security
- TRUST 5원칙: `.moai/memory/development-guide.md`
