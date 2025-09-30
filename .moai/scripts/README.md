# MoAI 스크립트 디렉토리

이 디렉토리는 MoAI-ADK 핵심 스크립트와 프로젝트별 자동화 스크립트를 포함한다.

## 📦 MoAI-ADK 핵심 Scripts (10개)

### 1차 호출 Scripts (에이전트가 직접 실행)

| Script | TAG | 역할 | 사용처 |
|--------|-----|------|--------|
| **debug-analyzer.ts** | DEBUG-ANALYZER-001 | 오류 분석 및 진단 | `@agent-debug-helper` |
| **detect-language.ts** | LANGUAGE-DETECT-001 | 프로젝트 언어 감지 | `steering-guard.js`, `file-monitor.js` |
| **doc-syncer.ts** | DOC-SYNCER-001 | 문서 동기화 | `@agent-doc-syncer`, `/moai:3-sync` |
| **spec-builder.ts** | SPEC-BUILDER-001 | SPEC 문서 생성 | `@agent-spec-builder`, `/moai:1-spec` |
| **tdd-runner.ts** | TDD-RUNNER-001 | TDD 실행 | `@agent-code-builder`, `/moai:2-build` |
| **trust-checker.ts** | TRUST-CHECKER-001 | TRUST 5원칙 검증 | `@agent-trust-checker` |
| **tag-updater.ts** | TAG-UPDATER-001 | TAG 시스템 관리 | `@agent-tag-agent` |

### 내부 지원 Scripts (다른 scripts가 import)

| Script | TAG | 역할 | 사용처 |
|--------|-----|------|--------|
| **project-init.ts** | PROJECT-INIT-001 | 프로젝트 초기화 | CLI `init` 명령 내부 |
| **spec-validator.ts** | SPEC-VALIDATOR-001 | SPEC 검증 | `spec-builder.ts` 내부 |
| **test-analyzer.ts** | TEST-ANALYZER-001 | 테스트 분석 | `tdd-runner.ts`, `trust-checker.ts` 내부 |

## 권장 스크립트 구조

### 1. 프로젝트 관리
- `project-setup.ts` - 프로젝트 초기 설정
- `detect-tools.ts` - 개발 도구 자동 감지
- `language-detector.ts` - 프로젝트 주 언어 식별

### 2. TAG 시스템 관리 (코드 스캔 방식)

**핵심 철학**: TAG의 진실은 코드 자체에만 존재한다. 중간 인덱스 없이 rg/grep으로 실시간 스캔한다.

```bash
# TAG 전체 검증 (코드 직접 스캔)
rg '@TAG' -n src/ tests/ docs/

# 8-Core @TAG 검증

# Primary Chain (4 Core) 검증
rg '@REQ:[A-Z]+-[0-9]{3}' -n src/
rg '@DESIGN:[A-Z]+-[0-9]{3}' -n src/
rg '@TASK:[A-Z]+-[0-9]{3}' -n src/
rg '@TEST:[A-Z]+-[0-9]{3}' -n tests/

# Implementation (4 Core) 검증
rg '@FEATURE:[A-Z]+-[0-9]{3}' -n src/
rg '@API:[A-Z]+-[0-9]{3}' -n src/
rg '@UI:[A-Z]+-[0-9]{3}' -n src/
rg '@DATA:[A-Z]+-[0-9]{3}' -n src/

# 고아 TAG 감지
rg '@TAG:DEPRECATED' -n

# 특정 도메인 TAG 검색
rg '@TAG:[A-Z]+-AUTH' -n
```

**스크립트 예시** (`tag-validator.ts`):
```typescript
// @FEATURE:TAG-VALIDATOR-001 | Chain: @REQ:TAG-001 -> @DESIGN:TAG-001 -> @TASK:TAG-001 -> @TEST:TAG-001
import { execSync } from 'child_process';

interface TagValidationResult {
  total: number;
  broken: string[];
  orphaned: string[];
}

export function validateTags(): TagValidationResult {
  // 코드 직접 스캔 - 중간 인덱스 없음
  const output = execSync('rg "@TAG" -n src/ tests/', {
    encoding: 'utf-8',
  });

  // TAG 체인 검증 로직
  // ...
}
```

### 3. 품질 검증
- `quality-check.ts` - TRUST 원칙 검증
- `coverage-report.ts` - 테스트 커버리지 리포트
- `complexity-analyzer.ts` - 복잡도 분석

### 4. 문서 관리
- `doc-generator.ts` - API 문서 자동 생성
- `readme-updater.ts` - README 동기화
- `changelog-builder.ts` - 변경 이력 생성

## 사용 가이드

### 기본 실행
```bash
# TypeScript 직접 실행 (tsx 사용)
tsx .moai/scripts/[script-name].ts [options]

# 컴파일 후 실행
node .moai/scripts/[script-name].js [options]
```

### 도움말 확인
```bash
tsx .moai/scripts/[script-name].ts --help
```

## 스크립트 작성 가이드

스크립트를 추가할 때 다음 규칙을 준수한다:

1. **TypeScript로 작성**: 모든 스크립트는 TypeScript로 작성
2. **Commander.js 패턴 사용**: CLI 인터페이스 표준화
3. **JSON 출력 형식 준수**: 구조화된 결과 제공
4. **`--help` 옵션 제공**: 사용법 안내 필수
5. **에러 처리**: 명확한 에러 메시지와 종료 코드

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

## TAG 시스템 통합

**중요**: TAG INDEX 파일을 생성하거나 관리하는 스크립트는 작성하지 않는다. TAG의 진실은 오직 코드 자체에만 존재한다.

**권장 방식**:
- ✅ `rg`/`grep`으로 코드 직접 스캔
- ✅ 실시간 TAG 검증
- ✅ TAG 체인 무결성 확인
- ❌ TAG INDEX 파일 생성/관리
- ❌ 중간 캐시 사용
- ❌ 별도 TAG 데이터베이스

## .claude/ 에이전트 통합

이 디렉토리의 스크립트들은 `.claude/agents/` 에이전트 지침에서 참조될 수 있다:

- `doc-syncer.ts` → `@agent-doc-syncer`에서 사용
- `quality-check.ts` → `@agent-trust-checker`에서 사용
- `tag-validator.ts` → `/moai:3-sync`에서 사용

---

이 디렉토리는 프로젝트별 자동화 스크립트를 위한 **템플릿 공간**이다. 필요에 따라 스크립트를 추가하고 커스터마이징한다.