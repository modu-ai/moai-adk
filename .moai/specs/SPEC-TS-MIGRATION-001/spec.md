---
id: SPEC-TS-MIGRATION-001
version: "1.0.0"
status: "draft"
created: "2025-12-05"
updated: "2025-12-05"
author: "MoAI-ADK Team"
priority: "high"
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|----------|
| 1.0.0 | 2025-12-05 | MoAI-ADK Team | 초기 SPEC 작성 |

---

# SPEC-TS-MIGRATION-001: MoAI-ADK TypeScript 마이그레이션

## 1. 개요

### 1.1 목적
Python 기반 MoAI-ADK를 TypeScript로 완전히 재작성하여 웹 대시보드와 CLI를 통합된 모노레포 구조로 개발한다.

### 1.2 범위

**포함 범위:**
- Turborepo 모노레포 구조 설정
- 6개 핵심 패키지 구현 (types, config, core, git, hooks, ui)
- 2개 앱 구현 (cli, dashboard)
- 웹 대시보드 (Next.js 15 + HeroUI)
- CLI 앱 (Bun + Commander)
- Claude Code Hooks 통합
- Biome 기반 코드 품질 관리

**제외 범위:**
- Python 버전 유지보수 (별도 SPEC)
- MCP 서버 구현 (선택적 확장)
- 모바일 앱
- 백엔드 API 서버 (별도)

### 1.3 기술 스택

| 기술 | 버전 | 용도 |
|------|------|------|
| Bun | v1.3+ | 런타임 + 패키지 관리 |
| Next.js | v15 | 웹 대시보드 프레임워크 |
| HeroUI | latest | UI 컴포넌트 라이브러리 |
| Biome | v2.x | 린터 + 포매터 |
| TanStack Query | v5+ | 서버 상태 관리 |
| Zustand | v4+ | 클라이언트 상태 관리 |
| Zod | v3+ | 타입 검증 |
| Vitest | latest | 테스트 프레임워크 |
| Playwright | latest | E2E 테스트 |

---

## 2. EARS 요구사항

### 2.1 Ubiquitous Requirements (항상 적용)

| ID | 요구사항 |
|----|----------|
| U-001 | 시스템은 TypeScript strict 모드로 컴파일되어야 한다 |
| U-002 | 모든 패키지는 ESM 형식으로 빌드되어야 한다 |
| U-003 | 코드 스타일은 Biome 설정을 준수해야 한다 |
| U-004 | 모든 public API는 JSDoc 주석을 포함해야 한다 |
| U-005 | Node.js 20+ 및 Bun 1.3+와 호환되어야 한다 |

### 2.2 Event-Driven Requirements (이벤트 발생 시)

| ID | 트리거 | 요구사항 |
|----|--------|----------|
| E-001 | CLI 명령어 실행 시 | 해당 핸들러가 호출되어야 한다 |
| E-002 | SPEC 파일 변경 시 | Dashboard가 실시간으로 업데이트되어야 한다 |
| E-003 | Git 이벤트 발생 시 | Hooks가 적절히 처리해야 한다 |
| E-004 | 설정 파일 변경 시 | Config 서비스가 자동으로 반영해야 한다 |

### 2.3 State-Driven Requirements (상태 조건)

| ID | 상태 조건 | 요구사항 |
|----|----------|----------|
| S-001 | Dashboard가 활성 상태일 때 | 프로젝트 상태가 5초마다 폴링되어야 한다 |
| S-002 | SPEC이 in_progress 상태일 때 | 관련 파일 변경이 추적되어야 한다 |
| S-003 | Git 작업 중일 때 | 충돌 가능성이 감지되어야 한다 |

### 2.4 Optional Features (선택적 기능)

| ID | 요구사항 |
|----|----------|
| O-001 | 시스템은 다크 모드 UI를 지원할 수 있다 |
| O-002 | 시스템은 원격 SPEC 동기화를 지원할 수 있다 |
| O-003 | 시스템은 Context7 MCP 통합을 지원할 수 있다 |

### 2.5 Complex Requirements (복합 조건)

| ID | 조건 | 요구사항 |
|----|------|----------|
| C-001 | SPEC 생성 중 Git 충돌이 발생하면 | 시스템은 충돌 해결 워크플로우를 제공해야 한다 |
| C-002 | TDD 테스트 실패 시 | 시스템은 자동 수정 제안을 제공해야 한다 |

---

## 3. 아키텍처 개요

### 3.1 모노레포 구조

```
moai-adk-ts/
├── apps/
│   ├── dashboard/          # Next.js 15 웹 대시보드
│   └── cli/                # Bun + Commander CLI
├── packages/
│   ├── core/               # 핵심 비즈니스 로직
│   ├── hooks/              # Claude Code Hooks
│   ├── config/             # Zod 스키마 기반 설정
│   ├── git/                # Git 작업 추상화
│   ├── ui/                 # 공유 UI 컴포넌트
│   └── types/              # 공유 타입 정의
├── templates/              # 배포 템플릿
├── turbo.json
├── biome.json
└── package.json
```

### 3.2 레이어 아키텍처

```
Presentation Layer (Dashboard, CLI, Hooks)
         ↓
Application Layer (Services)
         ↓
Domain Layer (Entities, Value Objects)
         ↓
Infrastructure Layer (Git, FileSystem, HTTP)
```

---

## 4. 의존성

### 4.1 외부 의존성
- Bun 런타임 (v1.3+)
- Git (v2.0+)
- Node.js (v20+ - 호환성)

### 4.2 내부 의존성
- 없음 (첫 번째 TypeScript SPEC)

---

## 5. 제약사항

### 5.1 기술적 제약
- Python 버전과의 하위 호환성 불필요 (완전 재작성)
- 설정 파일 형식은 YAML 유지 (기존 호환)
- Claude Code hooks는 Bun 또는 Node.js로 실행

### 5.2 비기술적 제약
- 개발 기간: 18-21주
- 단계별 마일스톤 검증 필요

---

## 6. 품질 속성

| 속성 | 목표 | 측정 방법 |
|------|------|----------|
| 테스트 커버리지 | 80% 이상 | Vitest coverage |
| 빌드 시간 | 30초 이내 | Turborepo 캐시 |
| Dashboard 로드 | 3초 이내 | Lighthouse |
| CLI 응답 시간 | 500ms 이내 | 벤치마크 |
