
## 1. Phase 1: Critical Files (Priority 1)

### 1.1 installer-core.ts
- **파일 경로**: `moai-adk-ts/src/core/installer/installer-core.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/installer-core.test.ts`
- **주요 테스트 케이스**:
  - InstallerCore 초기화 및 의존성 주입
  - install() 메서드 정상 동작
  - update() 메서드 정상 동작
  - Phase 실행 순서 검증
  - 에러 발생 시 처리

### 1.2 phase-executor.ts
- **파일 경로**: `moai-adk-ts/src/core/installer/phase-executor.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/phase-executor.test.ts`
- **주요 테스트 케이스**:
  - executePhase() 각 Phase별 실행
  - 의존성 객체 모킹 (DependencyInstaller, TemplateInstaller 등)
  - 에러 전파 검증
  - 로깅 검증

### 1.3 update-executor.ts
- **파일 경로**: `moai-adk-ts/src/core/installer/update-executor.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/update-executor.test.ts`
- **주요 테스트 케이스**:
  - executeUpdate() 정상 동작
  - 백업 생성 검증
  - 업데이트 Phase 순서 검증
  - 실패 시 롤백 트리거

## 2. Phase 2: High Priority Files

### 2.1 dependency-installer.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/dependency-installer.test.ts`
- **주요 테스트 케이스**:
  - installDependencies() 다양한 패키지 매니저 (npm, pnpm, yarn)
  - 패키지 설치 실패 시 에러 처리
  - 의존성 버전 검증

### 2.2 package-manager.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/package-manager.test.ts`
- **주요 테스트 케이스**:
  - detectPackageManager() 각 패키지 매니저 감지
  - installPackages() 실행 검증
  - which 명령 모킹

### 2.3 template-installer.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/template-installer.test.ts`
- **주요 테스트 케이스**:
  - installTemplates() 템플릿 복사
  - 파일 시스템 작업 모킹
  - 템플릿 처리 검증

## 3. Phase 3: Medium Priority Files

### 3.1 context-manager.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/context-manager.test.ts`
- **주요 테스트 케이스**:
  - createContext() 컨텍스트 생성
  - 컨텍스트 데이터 접근 및 수정

### 3.2 fallback-builder.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/fallback-builder.test.ts`
- **주요 테스트 케이스**:
  - buildFallbackStructure() 폴백 디렉토리 생성
  - 디렉토리 구조 검증

### 3.3 typescript-setup.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/typescript-setup.test.ts`
- **주요 테스트 케이스**:
  - setupTypeScript() tsconfig.json 생성
  - TypeScript 설정 검증

## 4. Phase 4: Low Priority Files

### 4.1 pre-install.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/pre-install.test.ts`
- **주요 테스트 케이스**:
  - runPreInstall() 사전 검증 로직

### 4.2 post-install.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/post-install.test.ts`
- **주요 테스트 케이스**:
  - runPostInstall() 후처리 로직

### 4.3 update-manager.ts
- **테스트 파일**: `moai-adk-ts/tests/core/installer/update-manager.test.ts`
- **주요 테스트 케이스**:
  - manageUpdate() 업데이트 관리

## 5. TDD 워크플로우

### 5.1 Red Phase
```bash
# 실패하는 테스트 작성
pnpm test:watch moai-adk-ts/tests/core/installer/installer-core.test.ts
```

### 5.2 Green Phase
```bash
# 최소한의 코드로 테스트 통과
pnpm test moai-adk-ts/tests/core/installer/installer-core.test.ts
```

### 5.3 Refactor Phase
```bash
# 코드 개선 및 커버리지 확인
pnpm test:coverage
```

## 6. 커버리지 모니터링

### 6.1 전체 커버리지 확인
```bash
pnpm test:coverage --reporter=html
open coverage/index.html
```

### 6.2 파일별 커버리지 확인
```bash
pnpm test:coverage --reporter=json-summary
cat coverage/coverage-summary.json | jq '.["path/to/file.ts"]'
```

## 7. 일정

- **Phase 1**: 2일 (Critical Files)
- **Phase 2**: 2일 (High Priority Files)
- **Phase 3**: 1.5일 (Medium Priority Files)
- **Phase 4**: 0.5일 (Low Priority Files)
- **Total**: 6일

## 8. 체크리스트

- [ ] 각 파일 커버리지 ≥ 85% 달성
- [ ] 전체 커버리지 ≥ 85% 달성
- [ ] CI/CD에 커버리지 게이트 추가
- [ ] 테스트 문서화 (README.md 업데이트)
