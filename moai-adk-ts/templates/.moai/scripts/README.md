# MoAI 핵심 스크립트

이 디렉토리는 MoAI-ADK의 핵심 기능을 담당하는 TypeScript 스크립트들을 포함합니다.

## 📋 스크립트 목록

### 1. 프로젝트 관리 스크립트

#### `project-init.ts` - 프로젝트 초기화
```bash
tsx .moai/scripts/project-init.ts --name "my-project" --type personal
```
- MoAI 프로젝트 구조 생성
- 기본 설정 파일 생성
- TAG 시스템 초기화

#### `detect-language.ts` - 언어 감지
```bash
tsx .moai/scripts/detect-language.ts --path . --verbose
```
- 프로젝트 주 언어 자동 감지
- 프레임워크 및 도구 추천
- 패키지 매니저 감지

### 2. SPEC 관리 스크립트

#### `spec-builder.ts` - SPEC 문서 생성
```bash
tsx .moai/scripts/spec-builder.ts --interactive
tsx .moai/scripts/spec-builder.ts --title "새 기능" --type feature
```
- EARS 방식 SPEC 문서 생성
- 대화형 SPEC 작성 지원
- 메타데이터 자동 관리

#### `spec-validator.ts` - SPEC 검증
```bash
tsx .moai/scripts/spec-validator.ts --all --fix
tsx .moai/scripts/spec-validator.ts --spec SPEC-001 --strict
```
- SPEC 문서 유효성 검증
- @TAG 형식 검사
- 자동 수정 기능

### 3. TDD 구현 스크립트

#### `tdd-runner.ts` - TDD 사이클 실행
```bash
tsx .moai/scripts/tdd-runner.ts --phase all --coverage
tsx .moai/scripts/tdd-runner.ts --phase red --language typescript
```
- Red-Green-Refactor 사이클 자동화
- 다중 언어 지원 (Python, TypeScript, Java, Go, Rust)
- 커버리지 측정 및 품질 검증

#### `test-analyzer.ts` - 테스트 분석
```bash
tsx .moai/scripts/test-analyzer.ts --coverage --format markdown
tsx .moai/scripts/test-analyzer.ts --path src --save
```
- 테스트 파일 자동 스캔
- 커버리지 분석
- 품질 점수 계산

### 4. 동기화 스크립트

#### `doc-syncer.ts` - 문서 동기화
```bash
tsx .moai/scripts/doc-syncer.ts --target all
tsx .moai/scripts/doc-syncer.ts --target readme
```
- README.md 자동 업데이트
- API 문서 생성
- 릴리스 노트 생성

#### `tag-updater.ts` - TAG 인덱스 업데이트
```bash
tsx .moai/scripts/tag-updater.ts --scan --repair --backup
tsx .moai/scripts/tag-updater.ts --validate --cleanup
```
- 프로젝트 전체 TAG 스캔
- TAG 데이터베이스 수리
- 고아 TAG 및 끊어진 참조 감지

### 5. 품질 검증 스크립트

#### `trust-checker.ts` - TRUST 원칙 검증
```bash
tsx .moai/scripts/trust-checker.ts --principle all --report
tsx .moai/scripts/trust-checker.ts --principle test --fix
```
- TRUST 5원칙 자동 검증
- 코드 품질 점수 계산
- 자동 수정 제안

#### `debug-analyzer.ts` - 디버깅 분석
```bash
tsx .moai/scripts/debug-analyzer.ts --error "Error message"
tsx .moai/scripts/debug-analyzer.ts --system --performance --dependencies
```
- 에러 메시지 패턴 분석
- 시스템 진단 및 성능 분석
- 의존성 이슈 감지

## 🔧 사용법

### 기본 실행
```bash
# tsx로 직접 실행
tsx .moai/scripts/[script-name].ts [options]

# 또는 Node.js로 실행 (사전 컴파일 필요)
node .moai/scripts/[script-name].js [options]
```

### 도움말 확인
```bash
tsx .moai/scripts/[script-name].ts --help
```

### 전역 설치 (선택사항)
```bash
# TypeScript를 전역으로 실행할 수 있도록 설정
npm install -g tsx

# 이후 간단하게 실행 가능
cd your-project
tsx .moai/scripts/project-init.ts
```

## 📊 워크플로우 통합

### 1. 새 프로젝트 시작
```bash
# 1. 프로젝트 초기화
tsx .moai/scripts/project-init.ts --name "my-project"

# 2. 언어 감지 및 설정
tsx .moai/scripts/detect-language.ts --verbose

# 3. 첫 번째 SPEC 생성
tsx .moai/scripts/spec-builder.ts --interactive
```

### 2. 개발 사이클
```bash
# 1. SPEC 검증
tsx .moai/scripts/spec-validator.ts --spec SPEC-001 --fix

# 2. TDD 구현
tsx .moai/scripts/tdd-runner.ts --phase all --coverage

# 3. 품질 검증
tsx .moai/scripts/trust-checker.ts --principle all

# 4. 문서 동기화
tsx .moai/scripts/doc-syncer.ts --target all
```

### 3. 유지보수
```bash
# 1. TAG 시스템 업데이트
tsx .moai/scripts/tag-updater.ts --scan --repair

# 2. 테스트 분석
tsx .moai/scripts/test-analyzer.ts --coverage --save

# 3. 시스템 진단
tsx .moai/scripts/debug-analyzer.ts --system --performance
```

## 🎯 출력 형식

모든 스크립트는 다음 형식으로 일관된 출력을 제공합니다:

```json
{
  "success": true,
  "result": { ... },
  "nextSteps": [
    "다음 단계 안내"
  ]
}
```

## 🔗 .claude/ 지침과의 연동

이 스크립트들은 `.claude/` 디렉토리의 에이전트 지침에서 참조됩니다:

- **spec-builder**: `/moai:1-spec` 명령어에서 사용
- **tdd-runner**: `/moai:2-build` 명령어에서 사용
- **doc-syncer**: `/moai:3-sync` 명령어에서 사용
- **debug-analyzer**: `@agent-debug-helper`에서 사용
- **trust-checker**: 품질 검증 시 사용

## 🛠️ 개발자 가이드

### 새 스크립트 추가 시 준수사항

1. **TypeScript 기반**: 모든 스크립트는 TypeScript로 작성
2. **Commander.js 사용**: CLI 인터페이스는 Commander.js 패턴 준수
3. **JSON 출력**: 구조화된 결과를 JSON으로 출력
4. **에러 처리**: 명확한 에러 메시지와 종료 코드 제공
5. **도움말**: `--help` 옵션으로 사용법 안내

### 공통 인터페이스
```typescript
interface ScriptResult {
  success: boolean;
  message: string;
  data?: any;
  nextSteps?: string[];
}
```

### 종료 코드
- `0`: 성공
- `1`: 실패 또는 오류

---

이 스크립트들은 MoAI-ADK의 핵심 기능을 담당하며, `.claude/` 에이전트 시스템과 완전히 통합되어 동작합니다.