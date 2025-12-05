# SPEC-TS-MIGRATION-001 구현 계획

## 1. 개요

### 1.1 프로젝트 배경
Python 기반 MoAI-ADK는 90개 이상의 모듈로 구성되어 있으며, 다음과 같은 구조적 문제점이 있다:
- God Object 패턴 (phase_executor.py, template/processor.py)
- 3개의 분리된 Config 관리자
- 인터페이스 및 DI 부재
- GitPython에 강결합
- 레이어 분리 미흡

### 1.2 목표
- TypeScript로 완전 재작성
- Clean Architecture 적용
- 웹 대시보드 추가
- 테스트 커버리지 80% 이상

### 1.3 기대 효과
- 타입 안전성 확보
- IDE 지원 향상
- 웹 기반 관리 UI
- 유지보수성 개선

---

## 2. 기술 스택

### 2.1 런타임 및 빌드

| 기술 | 버전 | 용도 |
|------|------|------|
| Bun | v1.3.3+ | 런타임, 패키지 관리, 테스트 |
| Turborepo | latest | 모노레포 빌드 오케스트레이션 |
| TypeScript | v5.3+ | 언어 |
| Biome | v2.x | 린터 + 포매터 (ESLint + Prettier 대체) |

### 2.2 Dashboard 앱

| 기술 | 버전 | 용도 |
|------|------|------|
| Next.js | v15 | App Router 기반 프레임워크 |
| HeroUI | latest | UI 컴포넌트 라이브러리 |
| TanStack Query | v5+ | 서버 상태 관리 |
| Zustand | v4+ | 클라이언트 상태 관리 |
| Recharts | latest | 차트 시각화 |

### 2.3 CLI 앱

| 기술 | 버전 | 용도 |
|------|------|------|
| Commander.js | v12+ | CLI 프레임워크 |
| @clack/prompts | latest | 대화형 프롬프트 |
| chalk | v5+ | 터미널 스타일링 |
| ora | v8+ | 스피너 |

### 2.4 공유 패키지

| 기술 | 버전 | 용도 |
|------|------|------|
| Zod | v3+ | 스키마 검증 및 타입 추론 |
| simple-git | v3+ | Git 작업 추상화 |
| yaml | v2+ | YAML 파싱 |
| tsyringe | latest | 의존성 주입 |

### 2.5 테스트

| 기술 | 버전 | 용도 |
|------|------|------|
| Vitest | latest | 단위/통합 테스트 |
| Playwright | latest | E2E 테스트 |
| MSW | v2+ | API 모킹 |

---

## 3. 아키텍처 설계

### 3.1 모노레포 구조

```
moai-adk-ts/
├── apps/
│   ├── dashboard/              # Next.js 15 웹 대시보드
│   │   ├── app/
│   │   │   ├── dashboard/      # 프로젝트 개요
│   │   │   ├── specs/          # SPEC 관리
│   │   │   ├── tests/          # 테스트 대시보드
│   │   │   ├── git/            # Git 시각화
│   │   │   ├── hooks/          # 훅 모니터링
│   │   │   ├── settings/       # 설정 관리
│   │   │   └── api/            # API Routes
│   │   └── components/
│   │
│   └── cli/                    # Bun + Commander CLI
│       └── src/
│           ├── commands/       # init, doctor, status, update, worktree
│           └── ui/             # 대화형 프롬프트
│
├── packages/
│   ├── core/                   # 핵심 비즈니스 로직
│   │   └── src/
│   │       ├── domain/         # 엔티티 (Spec, Project, Session)
│   │       ├── services/       # 애플리케이션 서비스
│   │       └── repositories/   # 리포지토리 인터페이스
│   │
│   ├── hooks/                  # Claude Code Hooks
│   │   └── src/
│   │       ├── handlers/       # 훅 타입별 핸들러
│   │       └── lifecycle/      # 라이프사이클 관리
│   │
│   ├── config/                 # 설정 스키마 및 관리
│   │   └── src/
│   │       ├── schemas/        # Zod 스키마
│   │       ├── loader.ts       # 설정 로딩
│   │       └── defaults.ts     # 기본값
│   │
│   ├── git/                    # Git 작업 추상화
│   │   └── src/
│   │       ├── interfaces/     # IGitRepository
│   │       └── adapters/       # SimpleGitAdapter
│   │
│   ├── ui/                     # 공유 UI 컴포넌트
│   │
│   └── types/                  # 공유 타입 정의
│
├── templates/                  # 배포 템플릿 (기존 유지)
│
├── turbo.json
├── biome.json
├── package.json
└── tsconfig.base.json
```

### 3.2 Clean Architecture 레이어

```
┌─────────────────────────────────────────────────────────────────┐
│                      PRESENTATION LAYER                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Dashboard  │  │     CLI      │  │    Hooks     │          │
│  │  (Next.js)   │  │ (Commander)  │  │  (Scripts)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                          │
│  ProjectService, SpecService, SessionService, GitService        │
│  ConfigService, HookService                                     │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                        DOMAIN LAYER                             │
│  Entities: Spec, Project, Session                               │
│  Value Objects: SpecId, SpecStatus, Config                      │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE LAYER                         │
│  FileSpecRepository, SimpleGitAdapter, FileSystem, Logger       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 구현 단계

### Phase 1: 기반 구축 (2-3주)

**목표:** 모노레포 설정 및 기본 구조 구축

**작업 항목:**
- [ ] Turborepo + Bun 워크스페이스 초기화
- [ ] TypeScript 공유 설정 (tsconfig.base.json)
- [ ] Biome 설정 (biome.json)
- [ ] packages/types 패키지 생성
- [ ] CI/CD 파이프라인 기본 구축

**산출물:**
- turbo.json, package.json, biome.json
- packages/types/src/index.ts

**검증 기준:**
- `bun install` 성공
- `bun run build` 성공
- `bun run lint` 성공

---

### Phase 2: Config 패키지 (2주)

**목표:** 설정 관리 시스템 구현

**작업 항목:**
- [ ] Zod 스키마 정의 (project, language, git-strategy)
- [ ] ConfigLoader 구현 (YAML/JSON 지원)
- [ ] ConfigService 싱글톤 구현
- [ ] 기존 Python config.yaml 호환성 테스트
- [ ] 단위 테스트 작성

**산출물:**
- packages/config/src/schemas/*.ts
- packages/config/src/loader.ts
- packages/config/src/service.ts

**검증 기준:**
- 기존 config.yaml 파싱 성공
- 테스트 커버리지 90%

---

### Phase 3: Core + Git 패키지 (3-4주)

**목표:** 핵심 비즈니스 로직 및 Git 통합

**작업 항목:**
- [ ] 도메인 엔티티 구현 (Spec, Project, Session)
- [ ] 서비스 레이어 구현 (ProjectService, SpecService)
- [ ] IGitRepository 인터페이스 정의
- [ ] SimpleGitAdapter 구현
- [ ] WorktreeService 구현
- [ ] DI 컨테이너 설정 (tsyringe)

**산출물:**
- packages/core/src/domain/*.ts
- packages/core/src/services/*.ts
- packages/git/src/*.ts

**검증 기준:**
- 단위 테스트 통과
- Git 작업 통합 테스트 통과

---

### Phase 4: CLI 앱 (2-3주)

**목표:** CLI 도구 구현

**작업 항목:**
- [ ] Commander.js 기반 CLI 구조
- [ ] init 명령어 구현
- [ ] doctor 명령어 구현
- [ ] status 명령어 구현
- [ ] update 명령어 구현
- [ ] worktree 명령어 구현
- [ ] @clack/prompts 대화형 UI

**산출물:**
- apps/cli/src/commands/*.ts
- apps/cli/src/ui/*.ts

**검증 기준:**
- Python CLI와 동일 UX
- E2E 테스트 통과

---

### Phase 5: Hooks 패키지 (2주)

**목표:** Claude Code Hooks 통합

**작업 항목:**
- [ ] 훅 핸들러 구현 (PreTool, PostTool, SessionStart, SessionEnd)
- [ ] 훅 라이프사이클 관리
- [ ] 로그 출력 (.jsonl 형식)
- [ ] settings.json 통합

**산출물:**
- packages/hooks/src/handlers/*.ts
- packages/hooks/src/lifecycle/*.ts

**검증 기준:**
- Claude Code 통합 테스트 통과
- 훅 실행 시간 15초 이내

---

### Phase 6: Dashboard 앱 (4-5주)

**목표:** 웹 대시보드 구현

**작업 항목:**
- [ ] Next.js 15 App Router 설정
- [ ] HeroUI 통합
- [ ] 레이아웃 및 네비게이션
- [ ] 페이지 구현:
  - [ ] dashboard (프로젝트 개요)
  - [ ] specs (SPEC 관리)
  - [ ] tests (테스트 대시보드)
  - [ ] git (Git 시각화)
  - [ ] hooks (훅 모니터링)
  - [ ] settings (설정 관리)
- [ ] API Routes 구현
- [ ] SSE 실시간 훅 모니터링
- [ ] TanStack Query + Zustand 상태 관리

**산출물:**
- apps/dashboard/app/**/*.tsx
- apps/dashboard/components/**/*.tsx

**검증 기준:**
- 페이지 로드 3초 이내
- Lighthouse 성능 80+
- 반응형 디자인

---

### Phase 7: 통합 및 안정화 (2주)

**목표:** 전체 시스템 통합 및 안정화

**작업 항목:**
- [ ] CLI + Dashboard + Hooks 통합 테스트
- [ ] 성능 최적화 (캐싱, 번들 크기)
- [ ] API 문서화 (TSDoc)
- [ ] 마이그레이션 가이드 작성
- [ ] npm 패키지 발행 준비

**산출물:**
- 통합 테스트 스위트
- API 문서
- 마이그레이션 가이드

**검증 기준:**
- 전체 테스트 커버리지 80%
- 프로덕션 빌드 성공
- npm 발행 준비 완료

---

## 5. 위험 요소 및 완화 전략

| 위험 | 영향도 | 확률 | 완화 전략 |
|------|--------|------|----------|
| Next.js 15 불안정성 | 높음 | 낮음 | Next.js 14 폴백 준비 |
| HeroUI 기능 부족 | 중간 | 중간 | Tailwind 직접 사용 가능 |
| Bun 호환성 이슈 | 중간 | 낮음 | Node.js 폴백 지원 |
| Python 사용자 혼란 | 중간 | 높음 | 명확한 마이그레이션 가이드 |
| 기능 누락 | 높음 | 중간 | 체크리스트 기반 검증 |
| 일정 지연 | 높음 | 중간 | 버퍼 2주 포함 |

---

## 6. 품질 게이트

| 게이트 | 기준 | 측정 방법 |
|--------|------|----------|
| 코드 품질 | Biome 오류 0개 | `bun run lint` |
| 테스트 커버리지 | 80% 이상 | Vitest coverage |
| 타입 안전성 | strict 모드 통과 | `tsc --noEmit` |
| 빌드 성능 | 30초 이내 | Turborepo 캐시 |
| Dashboard 성능 | Lighthouse 80+ | lighthouse-ci |
| CLI 응답 시간 | 500ms 이내 | 벤치마크 |

---

## 7. 일정 요약

| Phase | 기간 | 누적 | 마일스톤 |
|-------|------|------|----------|
| Phase 1 | 2-3주 | 2-3주 | 모노레포 설정 완료 |
| Phase 2 | 2주 | 4-5주 | Config 시스템 완료 |
| Phase 3 | 3-4주 | 8-9주 | Core/Git 완료 |
| Phase 4 | 2-3주 | 10-12주 | CLI 완료 |
| Phase 5 | 2주 | 12-14주 | Hooks 완료 |
| Phase 6 | 4-5주 | 16-19주 | Dashboard 완료 |
| Phase 7 | 2주 | **18-21주** | 프로덕션 준비 |

**총 예상 기간: 18-21주 (약 4-5개월)**
