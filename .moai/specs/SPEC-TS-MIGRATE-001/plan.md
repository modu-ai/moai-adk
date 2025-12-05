# SPEC-TS-MIGRATE-001: 구현 계획서

---
spec_id: SPEC-TS-MIGRATE-001
version: "1.0.0"
created: "2025-12-05"
updated: "2025-12-05"
author: "MoAI-ADK Team"
---

## 1. 개요

본 문서는 MoAI-ADK의 Python에서 TypeScript/Bun으로의 마이그레이션 구현 계획을 정의합니다.

### 1.1 전체 마이그레이션 전략

마이그레이션은 **점진적 병행 운영** 전략을 채택합니다:

1. **Foundation First**: 의존성이 적은 foundation, utils 모듈부터 시작
2. **Bottom-Up**: 하위 모듈에서 상위 모듈로 진행
3. **God Object 분리 선행**: 마이그레이션 전 대형 파일 리팩토링
4. **Feature Parity 보장**: 각 단계별 기능 동등성 검증

---

## 2. 마일스톤 구조

### 2.1 우선순위 기반 마일스톤

#### Primary Goal: 프로젝트 인프라 구축 (Phase 0-1)

Phase 0와 Phase 1 완료를 통해 TypeScript 프로젝트 기반 확립

#### Secondary Goal: 핵심 모듈 마이그레이션 (Phase 2-4)

foundation, statusline, project, utils, core 모듈의 TypeScript 전환

#### Final Goal: CLI 및 통합 완료 (Phase 5-6)

CLI 마이그레이션, templates 자산 이전, 전체 시스템 통합

#### Optional Goal: 최적화 및 확장 (Phase 7+)

성능 최적화, WebAssembly 통합, 플러그인 시스템

---

## 3. 단계별 구현 계획

### Phase 0: God Object 리팩토링 (Python)

**목표:** TypeScript 마이그레이션 전 코드베이스 정리

**우선순위:** 최상위 (Primary Goal)

**의존성:** 없음

**작업 항목:**

1. core/ 모듈 God Object 분리
   - jit_enhanced_hook_manager.py (1,987 LOC) 분리
     - hooks/manager.py - 훅 관리 핵심
     - hooks/jit_compiler.py - JIT 컴파일 로직
     - hooks/cache.py - 캐시 관리
     - hooks/executor.py - 실행 엔진
     - hooks/types.py - 타입 정의
   - error_recovery_system.py (1,902 LOC) 분리
     - errors/recovery.py - 복구 로직
     - errors/handlers.py - 에러 핸들러
     - errors/strategies.py - 복구 전략
     - errors/types.py - 타입 정의
   - realtime_monitoring_dashboard.py (1,724 LOC) 분리
     - monitoring/dashboard.py - 대시보드 UI
     - monitoring/metrics.py - 메트릭 수집
     - monitoring/alerts.py - 알림 시스템
     - monitoring/types.py - 타입 정의
   - enterprise_features.py (1,404 LOC) 분리
     - enterprise/features.py - 기능 관리
     - enterprise/licensing.py - 라이선스
     - enterprise/types.py - 타입 정의
   - event_driven_hook_system.py (1,371 LOC) 분리
     - events/system.py - 이벤트 시스템
     - events/handlers.py - 핸들러
     - events/types.py - 타입 정의

2. cli/ 모듈 리팩토링
   - update.py (2,686 LOC, 57 함수) 분리
     - update/checker.py - 버전 체크
     - update/downloader.py - 다운로드
     - update/installer.py - 설치
     - update/rollback.py - 롤백
     - update/config.py - 설정
     - update/types.py - 타입 정의

3. project/ 모듈 리팩토링
   - configuration.py (1,084 LOC, 8 클래스) 분리
     - config/loader.py - 설정 로더
     - config/validator.py - 검증
     - config/schema.py - 스키마
     - config/types.py - 타입 정의

4. 캐싱 전략 통합
   - 5가지 캐싱 전략을 unified_cache.py로 통합

**품질 게이트:**
- 모든 파일 500 LOC 이하
- 순환 의존성 0건
- 기존 테스트 100% 통과
- 테스트 커버리지 90% 이상 유지

**산출물:**
- 리팩토링된 Python 코드베이스
- 모듈 분리 문서
- 업데이트된 테스트

---

### Phase 1: TypeScript 프로젝트 인프라 구축

**목표:** TypeScript/Bun 프로젝트 기반 구조 확립

**우선순위:** 최상위 (Primary Goal)

**의존성:** Phase 0 완료

**작업 항목:**

1. 프로젝트 초기화
   ```
   bun init
   ```

2. package.json 구성
   ```json
   {
     "name": "moai-adk",
     "version": "1.0.0",
     "type": "module",
     "main": "dist/index.js",
     "types": "dist/index.d.ts",
     "bin": {
       "moai-adk": "dist/cli/index.js"
     },
     "scripts": {
       "build": "bun build src/index.ts --outdir dist",
       "test": "vitest",
       "lint": "biome check .",
       "typecheck": "tsc --noEmit"
     }
   }
   ```

3. tsconfig.json 구성
   ```json
   {
     "compilerOptions": {
       "target": "ES2024",
       "module": "ESNext",
       "moduleResolution": "bundler",
       "strict": true,
       "noImplicitAny": true,
       "strictNullChecks": true,
       "noUncheckedIndexedAccess": true,
       "esModuleInterop": true,
       "declaration": true,
       "outDir": "dist",
       "rootDir": "src"
     },
     "include": ["src/**/*"],
     "exclude": ["node_modules", "dist"]
   }
   ```

4. Biome 설정 (biome.json)
   ```json
   {
     "linter": {
       "enabled": true,
       "rules": {
         "recommended": true,
         "suspicious": {
           "noExplicitAny": "error"
         }
       }
     },
     "formatter": {
       "enabled": true,
       "indentStyle": "space",
       "indentWidth": 2
     }
   }
   ```

5. Vitest 설정 (vitest.config.ts)
   ```typescript
   import { defineConfig } from 'vitest/config';

   export default defineConfig({
     test: {
       coverage: {
         provider: 'v8',
         reporter: ['text', 'lcov'],
         thresholds: {
           statements: 90,
           branches: 90,
           functions: 90,
           lines: 90
         }
       }
     }
   });
   ```

6. 디렉토리 구조 생성
   ```
   src/
   ├── core/
   ├── cli/
   ├── foundation/
   ├── statusline/
   ├── project/
   ├── utils/
   ├── types/
   └── index.ts
   ```

7. CI/CD 파이프라인 구성
   - GitHub Actions 워크플로우
   - 테스트 자동화
   - 린팅 검사
   - 타입 체크

**품질 게이트:**
- `bun build` 성공
- `bun test` 설정 완료
- `biome check` 통과
- `tsc --noEmit` 통과

**산출물:**
- TypeScript 프로젝트 구조
- CI/CD 파이프라인
- 개발 환경 문서

---

### Phase 2: Foundation 모듈 마이그레이션

**목표:** foundation/ 모듈 TypeScript 전환 (11,436 LOC, 13 파일)

**우선순위:** 높음 (Secondary Goal)

**의존성:** Phase 1 완료

**작업 항목:**

1. 타입 정의 생성
   ```typescript
   // src/types/foundation.ts
   import { z } from 'zod';

   export const SpecSchema = z.object({
     id: z.string(),
     version: z.string(),
     status: z.enum(['Draft', 'Review', 'Approved', 'Implemented']),
     created: z.string().datetime(),
     updated: z.string().datetime(),
     requirements: z.array(RequirementSchema)
   });

   export type Spec = z.infer<typeof SpecSchema>;
   ```

2. 도메인 클래스 마이그레이션
   - SpecBuilder 클래스
   - RequirementParser 클래스
   - ValidationEngine 클래스

3. Mock 구현을 실제 구현으로 전환
   - Python의 mock 구현 식별
   - TypeScript에서 실제 구현

4. 단위 테스트 작성
   - 각 클래스별 테스트 파일
   - 엣지 케이스 커버리지

**품질 게이트:**
- 테스트 커버리지 90% 이상
- 타입 오류 0건
- 린팅 오류 0건
- Python 버전과 기능 동등성 검증

**산출물:**
- TypeScript foundation/ 모듈
- 타입 정의 파일
- 테스트 스위트

---

### Phase 3: Utils 및 Project 모듈 마이그레이션

**목표:** utils/ (1,372 LOC) 및 project/ (2,097 LOC) 마이그레이션

**우선순위:** 높음 (Secondary Goal)

**의존성:** Phase 2 완료

**작업 항목:**

1. utils/ 모듈 마이그레이션
   - common.py 분리 및 마이그레이션 (SRP 위반 해결)
     - utils/string.ts - 문자열 유틸리티
     - utils/file.ts - 파일 유틸리티
     - utils/async.ts - 비동기 유틸리티
     - utils/validation.ts - 검증 유틸리티
   - toon_utils.py 완전 구현 (JSON 래퍼 이상)

2. project/ 모듈 마이그레이션
   - configuration 모듈 (Phase 0에서 분리됨)
   - project manager
   - 설정 스키마 (Zod)

3. 통합 캐시 레이어 구현
   ```typescript
   // src/core/cache/unified-cache.ts
   export class UnifiedCache {
     private memoryCache: Map<string, CacheEntry>;
     private fileCache: FileCache;
     private ttl: number;

     async get<T>(key: string): Promise<T | null>;
     async set<T>(key: string, value: T, options?: CacheOptions): Promise<void>;
     async invalidate(pattern: string): Promise<void>;
   }
   ```

**품질 게이트:**
- 테스트 커버리지 90% 이상
- 타입 오류 0건
- 기존 기능 100% 호환

**산출물:**
- TypeScript utils/ 모듈
- TypeScript project/ 모듈
- 통합 캐시 시스템

---

### Phase 4: Statusline 및 Core 모듈 마이그레이션

**목표:** statusline/ (2,673 LOC) 및 core/ (36,078 LOC) 마이그레이션

**우선순위:** 높음 (Secondary Goal)

**의존성:** Phase 3 완료

**작업 항목:**

1. statusline/ 모듈 마이그레이션
   - version_reader.ts (741 LOC 최적화)
   - 캐싱 전략 통합 적용
   - CLI 출력 컴포넌트

2. core/ 모듈 마이그레이션 (Phase 0에서 분리된 구조)
   - hooks/ 모듈
     - HookManager
     - JITCompiler
     - HookCache
     - HookExecutor
   - errors/ 모듈
     - RecoverySystem
     - ErrorHandlers
     - RecoveryStrategies
   - monitoring/ 모듈
     - Dashboard
     - MetricsCollector
     - AlertSystem
   - enterprise/ 모듈
     - FeatureManager
     - LicenseManager
   - events/ 모듈
     - EventSystem
     - EventHandlers

3. 의존성 주입 패턴 구현
   ```typescript
   // src/core/di/container.ts
   export class DIContainer {
     private services: Map<string, any>;

     register<T>(token: string, factory: () => T): void;
     resolve<T>(token: string): T;
   }
   ```

4. 순환 의존성 제거 검증

**품질 게이트:**
- 테스트 커버리지 90% 이상
- 순환 의존성 0건
- 모든 파일 500 LOC 이하
- 성능 벤치마크 통과

**산출물:**
- TypeScript statusline/ 모듈
- TypeScript core/ 모듈
- DI 컨테이너

---

### Phase 5: CLI 모듈 마이그레이션

**목표:** cli/ 모듈 TypeScript 전환 (6,960 LOC, 22 파일)

**우선순위:** 높음 (Final Goal)

**의존성:** Phase 4 완료

**작업 항목:**

1. CLI 프레임워크 선정 및 구성
   - Commander.js 또는 oclif 평가
   - 최종 프레임워크 선정
   - 기본 구조 설정

2. 명령어 마이그레이션
   ```typescript
   // src/cli/commands/init.ts
   import { Command } from 'commander';

   export const initCommand = new Command('init')
     .description('Initialize a new MoAI-ADK project')
     .option('-l, --language <lang>', 'Language setting')
     .action(async (options) => {
       // Implementation
     });
   ```

3. 대화형 프롬프트 구현
   - @inquirer/prompts 통합
   - 기존 InquirerPy 기능 매핑

4. 출력 스타일링
   - chalk 통합
   - ora (스피너)
   - cli-progress (진행바)

5. worktree/ 서브모듈 마이그레이션 (잘 구조화됨)

6. update/ 서브모듈 마이그레이션 (Phase 0에서 분리됨)

**품질 게이트:**
- 모든 CLI 명령어 100% 호환
- 테스트 커버리지 90% 이상
- 응답 시간 100ms 이내

**산출물:**
- TypeScript cli/ 모듈
- CLI 테스트 스위트
- 사용자 가이드

---

### Phase 6: Templates 및 통합 완료

**목표:** templates/ 자산 이전 및 전체 시스템 통합

**우선순위:** 높음 (Final Goal)

**의존성:** Phase 5 완료

**작업 항목:**

1. templates/ 자산 이전 (278 파일)
   - 정적 파일 복사
   - 변수 치환 시스템 마이그레이션
   - 템플릿 검증 도구

2. 전체 통합 테스트
   ```typescript
   // tests/e2e/full-workflow.test.ts
   describe('Full Workflow', () => {
     it('should initialize project', async () => {
       // E2E test
     });

     it('should run SPEC workflow', async () => {
       // E2E test
     });
   });
   ```

3. Python 버전 기능 동등성 검증
   - 기능 체크리스트 검증
   - 회귀 테스트

4. 문서화
   - API 문서
   - 마이그레이션 가이드
   - 릴리스 노트

5. Python 버전 deprecation 계획

**품질 게이트:**
- E2E 테스트 100% 통과
- 기능 동등성 100%
- 문서화 완료
- Python 버전과 병행 운영 가능

**산출물:**
- 완전한 TypeScript 버전
- 통합 테스트 스위트
- 마이그레이션 문서
- 릴리스 패키지

---

## 4. 기술 결정 사항

### 4.1 CLI 프레임워크 비교

| 기준 | Commander.js | oclif |
|------|-------------|-------|
| 학습 곡선 | 낮음 | 중간 |
| 기능 완성도 | 높음 | 매우 높음 |
| TypeScript 지원 | 좋음 | 우수 |
| 플러그인 시스템 | 없음 | 내장 |
| 번들 크기 | 작음 | 중간 |

**결정:** Commander.js (단순성, 낮은 학습 곡선)

### 4.2 타입 검증 라이브러리

| 기준 | Zod | TypeBox | io-ts |
|------|-----|---------|-------|
| 타입 추론 | 우수 | 우수 | 좋음 |
| 번들 크기 | 13KB | 작음 | 중간 |
| 생태계 | 넓음 | 중간 | 넓음 |
| 문서화 | 우수 | 좋음 | 좋음 |

**결정:** Zod (타입 추론, 생태계, 문서화)

### 4.3 테스트 프레임워크

| 기준 | Vitest | Jest | Bun Test |
|------|--------|------|----------|
| 속도 | 빠름 | 중간 | 매우 빠름 |
| ESM 지원 | 네이티브 | 설정 필요 | 네이티브 |
| 생태계 | 성장 중 | 성숙 | 제한적 |
| 호환성 | Jest 호환 | 기준 | 제한적 |

**결정:** Vitest (속도, ESM 지원, Jest 호환성)

---

## 5. 위험 완화 전략

### 5.1 Bun 런타임 불안정성 (RSK-001)

**완화 전략:**
- Node.js 호환 API 우선 사용
- Bun 전용 기능 추상화 레이어
- Node.js 폴백 구현 준비

### 5.2 라이브러리 호환성 (RSK-002)

**완화 전략:**
- 각 라이브러리 대체재 목록 준비
- 추상화 레이어를 통한 라이브러리 교체 용이성
- 호환성 테스트 스위트

### 5.3 일정 지연 (RSK-004)

**완화 전략:**
- 우선순위 기반 마일스톤 (시간 예측 대신)
- 각 Phase 독립적 배포 가능
- 기능 스코프 조정 유연성

---

## 6. 의존성 다이어그램

```
Phase 0 (God Object 리팩토링)
    │
    ▼
Phase 1 (인프라 구축)
    │
    ▼
Phase 2 (Foundation)
    │
    ▼
Phase 3 (Utils, Project)
    │
    ▼
Phase 4 (Statusline, Core)
    │
    ▼
Phase 5 (CLI)
    │
    ▼
Phase 6 (Templates, 통합)
```

---

## 7. 추적성

### 7.1 요구사항 매핑

| 요구사항 ID | 구현 Phase | 상태 |
|-------------|------------|------|
| REQ-U001 | Phase 1 | 대기 |
| REQ-U002 | Phase 2-6 | 대기 |
| REQ-U003 | Phase 5-6 | 대기 |
| REQ-U004 | Phase 2-5 | 대기 |
| REQ-E001 | Phase 5 | 대기 |
| REQ-E002 | Phase 0 | 대기 |
| REQ-E003 | Phase 0, 4 | 대기 |
| REQ-W001 | Phase 1-6 | 대기 |
| REQ-W002 | Phase 0, 4 | 대기 |
| REQ-W003 | Phase 0 | 대기 |

### 7.2 관련 문서

- spec.md: SPEC-TS-MIGRATE-001 명세서
- acceptance.md: 인수 기준

---

*본 계획서는 SPEC-TS-MIGRATE-001의 구현 로드맵을 정의합니다.*
