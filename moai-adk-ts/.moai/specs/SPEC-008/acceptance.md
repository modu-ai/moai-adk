# SPEC-008 인수 기준

@TEST:REFACTOR-008 | Chain: @CODE:REFACTOR-008 -> @TEST:REFACTOR-008

## 테스트 체크리스트

### 1. API 호환성 검증
- [ ] `ConfigManager.createClaudeSettings()` 동일한 결과 반환
- [ ] `ConfigManager.createMoAIConfig()` 동일한 결과 반환
- [ ] `ConfigManager.createPackageJson()` 동일한 결과 반환
- [ ] `ConfigManager.validateConfigFile()` 동일한 결과 반환
- [ ] `ConfigManager.backupConfigFile()` 동일한 결과 반환
- [ ] `ConfigManager.setupFullProjectConfig()` 동일한 결과 반환

### 2. 파일 크기 제약
- [ ] config-manager.ts ≤ 300 LOC
- [ ] claude-settings-builder.ts ≤ 150 LOC
- [ ] moai-config-builder.ts ≤ 150 LOC
- [ ] package-json-builder.ts ≤ 150 LOC
- [ ] config-file-utils.ts ≤ 150 LOC
- [ ] config-helpers.ts ≤ 150 LOC

### 3. 기존 테스트 통과
```bash
npm test -- config-manager.test.ts
# 모든 테스트 PASS
```

### 4. 새로운 유닛 테스트
- [ ] claude-settings-builder.test.ts 작성
- [ ] moai-config-builder.test.ts 작성
- [ ] package-json-builder.test.ts 작성
- [ ] config-file-utils.test.ts 작성
- [ ] 커버리지 ≥ 85%

### 5. 타입 안전성
```bash
npm run type-check
# 0 errors
```

### 6. 순환 의존성 검사
```bash
npx madge --circular src/core/config/
# No circular dependencies found
```

## 수동 검증 시나리오

### 시나리오 1: Solo 모드 초기화
```typescript
const manager = new ConfigManager();
const result = await manager.setupFullProjectConfig('/test/path', {
  projectName: 'test-project',
  mode: 'solo',
  runtime: 'node',
  techStack: ['typescript'],
  shouldCreatePackageJson: true
});

// 검증: result.success === true
// 검증: 3개 파일 생성 (claude, moai, package.json)
```

### 시나리오 2: 백업 생성
```typescript
const backupResult = await manager.backupConfigFile('/path/to/config.json');
// 검증: backupResult.success === true
// 검증: 백업 파일이 타임스탬프와 함께 생성됨
```

## 성능 기준
- setupFullProjectConfig() 실행 시간: < 500ms
- 개별 설정 생성: < 100ms each
