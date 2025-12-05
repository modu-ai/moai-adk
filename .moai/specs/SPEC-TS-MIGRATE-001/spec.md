---
id: SPEC-TS-MIGRATE-001
version: "1.0.0"
status: draft
created: "2025-12-05"
updated: "2025-12-05"
author: "MoAI-ADK Team"
priority: high
tags: [migration, typescript, bun, refactoring, architecture]
---

# SPEC-TS-MIGRATE-001: MoAI-ADK TypeScript/Bun 마이그레이션

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2025-12-05 | workflow-spec | 초기 SPEC 생성 |

---

## 1. 개요

### 1.1 배경

MoAI-ADK는 현재 Python 기반의 Claude Code 개발 도구로, 60,600+ LOC 규모의 코드베이스를 보유하고 있습니다. Claude Code의 핵심 런타임인 TypeScript/Bun 환경과의 긴밀한 통합 및 성능 개선을 위해 전체 코드베이스의 TypeScript 마이그레이션이 필요합니다.

### 1.2 목적

Python 코드베이스를 TypeScript/Bun으로 완전 마이그레이션하여:
- Claude Code와의 네이티브 통합 강화
- 타입 안전성 확보 (현재 `Dict[str, Any]` 남용 문제 해결)
- 런타임 성능 개선
- God Object 패턴 제거 및 아키텍처 개선

### 1.3 범위

**포함 범위:**
- core/ 모듈 (36,078 LOC, 82 파일)
- cli/ 모듈 (6,960 LOC, 22 파일)
- foundation/ 모듈 (11,436 LOC, 13 파일)
- statusline/ 모듈 (2,673 LOC, 10 파일)
- project/ 모듈 (2,097 LOC, 3 파일)
- utils/ 모듈 (1,372 LOC, 8 파일)
- templates/ 정적 자산 (278 파일)

**제외 범위:**
- 외부 MCP 서버 연동 (별도 SPEC 필요)
- Claude Desktop 앱 통합 (별도 SPEC 필요)

---

## 2. 환경 (Environment)

### 2.1 현재 환경

- **런타임:** Python 3.13+
- **패키지 관리:** uv, pip
- **CLI 프레임워크:** Typer, Click (혼재)
- **테스트:** pytest
- **린팅:** ruff, pylint
- **타입 체크:** mypy

### 2.2 목표 환경

- **런타임:** Bun 1.1+
- **언어:** TypeScript 5.9 strict mode
- **패키지 관리:** Bun (npm 호환)
- **CLI 프레임워크:** Commander.js 또는 oclif
- **테스트:** Vitest
- **린팅:** Biome
- **타입 체크:** TypeScript 내장 (strict)

### 2.3 라이브러리 매핑

| Python | TypeScript | 버전 |
|--------|------------|------|
| Typer/Click | Commander.js | 12+ |
| Rich | chalk + ora | 5+ / 8+ |
| InquirerPy | @inquirer/prompts | 7+ |
| GitPython | simple-git | 3.20+ |
| PyYAML | js-yaml | 4+ |
| asyncio | Native async/await | - |
| subprocess | execa | 9+ |
| dataclasses | Zod schemas | 3.23+ |
| pydantic | Zod | 3.23+ |

### 2.4 필수 의존성 (TypeScript)

| 패키지 | 버전 | 용도 |
|--------|------|------|
| bun | 1.1+ | 런타임 |
| typescript | 5.9+ | 언어 |
| zod | 3.23+ | 타입 검증 |
| commander | 12+ | CLI 프레임워크 |
| simple-git | 3.20+ | Git 작업 |
| chalk | 5+ | 터미널 색상 |
| ora | 8+ | 스피너 |
| js-yaml | 4+ | YAML 파싱 |
| execa | 9+ | 프로세스 실행 |
| @inquirer/prompts | 7+ | 대화형 프롬프트 |

### 2.5 개발 의존성

| 패키지 | 버전 | 용도 |
|--------|------|------|
| @biomejs/biome | 1.9+ | 린팅/포맷팅 |
| vitest | 2+ | 테스트 |
| @types/node | 22+ | Node.js 타입 |

---

## 3. 가정 (Assumptions)

### 3.1 기술적 가정

- [ASM-001] Bun 1.1+는 프로덕션 안정성을 갖추고 있음
- [ASM-002] TypeScript 5.9 strict mode는 모든 Python 패턴을 표현 가능
- [ASM-003] Commander.js/oclif는 Typer의 모든 기능을 대체 가능
- [ASM-004] simple-git은 GitPython의 주요 기능을 지원

### 3.2 프로젝트 가정

- [ASM-005] 마이그레이션 중 기존 Python 버전 유지보수 최소화
- [ASM-006] 단계별 마이그레이션으로 점진적 전환 가능
- [ASM-007] 기존 테스트 커버리지 90%+ 유지
- [ASM-008] God Object 리팩토링과 마이그레이션 병행 가능

### 3.3 리소스 가정

- [ASM-009] 4-5개월 마이그레이션 기간 확보
- [ASM-010] TypeScript/Bun 전문성 보유

---

## 4. 요구사항 (Requirements)

### 4.1 유비쿼터스 요구사항 (Ubiquitous)

시스템이 항상 만족해야 하는 조건입니다.

- [REQ-U001] 시스템은 모든 모듈에서 TypeScript strict mode를 적용해야 한다.
- [REQ-U002] 시스템은 90% 이상의 테스트 커버리지를 유지해야 한다.
- [REQ-U003] 시스템은 기존 CLI 명령어와 100% 호환성을 유지해야 한다.
- [REQ-U004] 시스템은 모든 공개 API에 타입 정의를 제공해야 한다.

### 4.2 이벤트 구동 요구사항 (Event-Driven)

특정 이벤트 발생 시 시스템의 반응을 정의합니다.

- [REQ-E001] **WHEN** 사용자가 `moai-adk` CLI 명령을 실행하면, **THEN** 시스템은 100ms 이내에 응답을 시작해야 한다.
- [REQ-E002] **WHEN** God Object가 감지되면, **THEN** 시스템은 단일 책임 원칙에 따라 모듈을 분리해야 한다.
- [REQ-E003] **WHEN** 순환 의존성이 감지되면, **THEN** 시스템은 의존성 주입 패턴으로 해결해야 한다.
- [REQ-E004] **WHEN** 타입 오류가 발생하면, **THEN** 시스템은 명확한 에러 메시지와 수정 가이드를 제공해야 한다.

### 4.3 원치 않는 동작 요구사항 (Unwanted Behavior)

시스템이 피해야 하는 동작을 정의합니다.

- [REQ-W001] 시스템은 `any` 타입 사용을 금지해야 한다 (명시적 예외 처리 제외).
- [REQ-W002] 시스템은 순환 의존성을 허용하지 않아야 한다.
- [REQ-W003] 시스템은 500 LOC 이상의 단일 파일을 허용하지 않아야 한다.
- [REQ-W004] 시스템은 하드코딩된 설정값을 포함하지 않아야 한다.
- [REQ-W005] 시스템은 동기 파일 I/O 작업을 허용하지 않아야 한다.

### 4.4 상태 구동 요구사항 (State-Driven)

시스템의 특정 상태에서의 동작을 정의합니다.

- [REQ-S001] **WHILE** 마이그레이션이 진행 중인 동안, 시스템은 Python과 TypeScript 버전을 병행 운영해야 한다.
- [REQ-S002] **WHILE** 테스트가 실행 중인 동안, 시스템은 격리된 테스트 환경을 유지해야 한다.
- [REQ-S003] **WHILE** 빌드가 진행 중인 동안, 시스템은 증분 컴파일을 지원해야 한다.

### 4.5 선택적 요구사항 (Optional)

구현이 선택적인 기능을 정의합니다.

- [REQ-O001] **WHERE** 성능이 중요한 경우, 시스템은 WebAssembly 모듈을 사용할 수 있다.
- [REQ-O002] **WHERE** 필요한 경우, 시스템은 Deno 런타임 호환성을 제공할 수 있다.
- [REQ-O003] **IF** 필요하다면, 시스템은 플러그인 시스템을 지원할 수 있다.

---

## 5. 명세 (Specifications)

### 5.1 아키텍처 설계

#### 5.1.1 모듈 구조

```
moai-adk/
├── src/
│   ├── core/                    # 핵심 기능 (리팩토링 필수)
│   │   ├── hooks/               # jit_enhanced_hook_manager 분리
│   │   ├── errors/              # error_recovery_system 분리
│   │   ├── monitoring/          # realtime_monitoring_dashboard 분리
│   │   ├── enterprise/          # enterprise_features 분리
│   │   └── events/              # event_driven_hook_system 분리
│   ├── cli/                     # CLI 명령어
│   │   ├── commands/            # 개별 명령어 모듈
│   │   ├── worktree/            # Worktree 관리
│   │   └── update/              # update.py 분리 (57 함수)
│   ├── foundation/              # 도메인 클래스
│   ├── statusline/              # 상태 표시줄
│   ├── project/                 # 프로젝트 설정
│   └── utils/                   # 유틸리티
├── templates/                   # 정적 자산 (변경 최소화)
├── tests/                       # 테스트
├── package.json                 # 의존성
├── tsconfig.json                # TypeScript 설정
└── biome.json                   # 린터 설정
```

#### 5.1.2 God Object 분리 계획

| 파일명 | 현재 LOC | 분리 모듈 수 | 분리 후 예상 LOC |
|--------|----------|--------------|------------------|
| jit_enhanced_hook_manager.py | 1,987 | 5 | 400 이하/모듈 |
| error_recovery_system.py | 1,902 | 4 | 475 이하/모듈 |
| realtime_monitoring_dashboard.py | 1,724 | 4 | 430 이하/모듈 |
| enterprise_features.py | 1,404 | 3 | 470 이하/모듈 |
| event_driven_hook_system.py | 1,371 | 3 | 460 이하/모듈 |
| update.py | 2,686 | 6 | 450 이하/모듈 |
| configuration.py | 1,084 | 4 | 270 이하/모듈 |

#### 5.1.3 캐싱 전략 통합

현재 5가지 캐싱 전략을 단일 캐싱 레이어로 통합:

```typescript
// src/core/cache/unified-cache.ts
export class UnifiedCache {
  private strategies: Map<CacheType, CacheStrategy>;

  constructor() {
    this.strategies = new Map([
      [CacheType.Memory, new MemoryCacheStrategy()],
      [CacheType.File, new FileCacheStrategy()],
      [CacheType.TTL, new TTLCacheStrategy()],
      [CacheType.LRU, new LRUCacheStrategy()],
      [CacheType.Hybrid, new HybridCacheStrategy()]
    ]);
  }
}
```

### 5.2 타입 시스템 설계

#### 5.2.1 핵심 타입 정의

```typescript
// src/types/core.ts
import { z } from 'zod';

export const ConfigSchema = z.object({
  moai: z.object({
    version: z.string(),
    versionCheck: z.object({
      enabled: z.boolean(),
      cacheTtlHours: z.number()
    })
  }),
  constitution: z.object({
    enforceTdd: z.boolean(),
    testCoverageTarget: z.number().min(0).max(100)
  }),
  gitStrategy: GitStrategySchema,
  language: LanguageSchema,
  project: ProjectSchema
});

export type Config = z.infer<typeof ConfigSchema>;
```

#### 5.2.2 타입 가드 패턴

```typescript
// src/utils/type-guards.ts
export function isConfig(value: unknown): value is Config {
  return ConfigSchema.safeParse(value).success;
}

export function assertConfig(value: unknown): asserts value is Config {
  ConfigSchema.parse(value);
}
```

### 5.3 테스트 전략

#### 5.3.1 테스트 구조

```
tests/
├── unit/                        # 단위 테스트
│   ├── core/
│   ├── cli/
│   └── utils/
├── integration/                 # 통합 테스트
│   ├── cli-commands/
│   └── workflows/
├── e2e/                         # E2E 테스트
│   └── scenarios/
└── fixtures/                    # 테스트 픽스처
```

#### 5.3.2 커버리지 기준

- 단위 테스트: 각 모듈 90% 이상
- 통합 테스트: 주요 워크플로우 100%
- E2E 테스트: 핵심 사용자 시나리오 100%

---

## 6. 제약사항 (Constraints)

### 6.1 기술적 제약

- [CON-001] TypeScript 5.9+ strict mode 필수
- [CON-002] Bun 1.1+ 런타임 사용
- [CON-003] ESM 모듈 시스템 사용 (CommonJS 미지원)
- [CON-004] Node.js 호환 API 사용 (Bun 전용 API 최소화)

### 6.2 품질 제약

- [CON-005] 테스트 커버리지 90% 이상 유지
- [CON-006] 모든 파일 500 LOC 이하
- [CON-007] 순환 의존성 0건
- [CON-008] `any` 타입 사용 금지 (명시적 예외 제외)

### 6.3 일정 제약

- [CON-009] 마이그레이션 기간 4-5개월 이내 완료
- [CON-010] 각 단계별 품질 게이트 통과 필수

---

## 7. 추적성 (Traceability)

### 7.1 관련 문서

- 프로젝트 분석 결과: 7개 Explore 에이전트 보고서
- 기술 스택 명세: tech.md (생성 예정)
- 아키텍처 문서: structure.md (생성 예정)

### 7.2 관련 SPEC

- (없음 - 첫 번째 마이그레이션 SPEC)

### 7.3 TAG 연결

```
SPEC-TS-MIGRATE-001 → plan.md (구현 계획)
SPEC-TS-MIGRATE-001 → acceptance.md (인수 기준)
```

---

## 8. 위험 요소 (Risks)

### 8.1 기술적 위험

| ID | 위험 | 영향도 | 발생확률 | 완화 전략 |
|----|------|--------|----------|-----------|
| RSK-001 | Bun 런타임 불안정성 | 높음 | 낮음 | Node.js 폴백 준비 |
| RSK-002 | 라이브러리 호환성 문제 | 중간 | 중간 | 대체 라이브러리 목록 준비 |
| RSK-003 | 성능 저하 | 중간 | 낮음 | 벤치마크 테스트 수립 |

### 8.2 프로젝트 위험

| ID | 위험 | 영향도 | 발생확률 | 완화 전략 |
|----|------|--------|----------|-----------|
| RSK-004 | 일정 지연 | 높음 | 중간 | 우선순위 기반 마일스톤 |
| RSK-005 | 기능 누락 | 중간 | 낮음 | 기능 체크리스트 관리 |
| RSK-006 | 테스트 커버리지 미달 | 중간 | 낮음 | CI 파이프라인 게이트 |

---

## 9. 승인 이력

| 역할 | 승인자 | 날짜 | 서명 |
|------|--------|------|------|
| SPEC 작성자 | workflow-spec | 2025-12-05 | - |
| 기술 검토자 | (대기) | - | - |
| 최종 승인자 | (대기) | - | - |

---

*본 SPEC은 EARS (Easy Approach to Requirements Syntax) 방법론을 따릅니다.*
