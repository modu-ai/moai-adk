# MoAI-ADK TypeScript 개발자 가이드

## 🚀 현대적 개발 스택

MoAI-ADK TypeScript 프로젝트는 최신 도구를 사용하여 최고 성능을 달성합니다.

### 성능 지표 (SPEC-012 달성)

- **Bun**: 98% 성능 향상 (npm 대비)
- **Vitest**: 92.9% 테스트 성공률, 빠른 실행
- **Biome**: 94.8% 린팅 성능, ESLint + Prettier 대체
- **TypeScript 5.9.2+**: 최신 언어 기능, 엄격한 타입 검사

### 도구 체인

```bash
# 패키지 관리자
bun install              # 의존성 설치 (npm 대신)
bun add <package>        # 패키지 추가
bun remove <package>     # 패키지 제거

# 개발 서버
bun run dev             # 개발 모드 (tsx 기반)
bun run build           # 프로덕션 빌드 (tsup)
bun run preview         # 빌드 결과 미리보기

# 테스트
bun test                # Vitest 테스트 실행
bun run test:watch      # 감시 모드 테스트
bun run test:coverage   # 커버리지 포함 테스트

# 코드 품질
bun run lint            # Biome 린팅
bun run format          # Biome 포맷팅
bun run type-check      # TypeScript 타입 검사
```

## 📁 프로젝트 구조

```
moai-adk-ts/
├── src/
│   ├── cli/                    # CLI 명령어
│   │   ├── commands/           # 개별 명령어 구현
│   │   └── index.ts           # Commander.js 진입점
│   ├── core/                   # 핵심 로직
│   │   ├── installer/          # 설치 시스템
│   │   ├── git/               # Git 관리
│   │   └── project/           # 프로젝트 관리
│   └── claude/                # Claude Code 통합
│       ├── agents/            # 에이전트 시스템
│       └── hooks/             # 이벤트 훅
├── resources/templates/        # 사용자 프로젝트 템플릿
├── __tests__/                 # Vitest 테스트
├── dist/                      # 빌드 결과 (ESM/CJS)
└── docs/                      # 개발 문서
```

## 🔧 개발 워크플로우

### 1. 환경 설정

```bash
# Bun 설치 (macOS/Linux)
curl -fsSL https://bun.sh/install | bash

# Windows
powershell -c "irm bun.sh/install.ps1 | iex"

# 프로젝트 설정
cd moai-adk-ts
bun install
bun run build
```

### 2. TDD 개발 사이클

```bash
# 1. RED: 실패하는 테스트 작성
bun test --watch                # 테스트 감시 모드

# 2. GREEN: 최소 구현
bun run dev                     # 개발 모드로 구현

# 3. REFACTOR: 코드 개선
bun run lint && bun run format  # 자동 린팅/포맷팅
bun run type-check             # 타입 안전성 검증
```

### 3. 품질 검증

```bash
# 전체 품질 체크
bun run build && bun test && bun run lint && bun run type-check

# 커버리지 확인
bun run test:coverage          # 95%+ 목표
```

## 🚦 성능 목표

### 빌드 성능
- **개발 빌드**: < 200ms (tsx 기반)
- **프로덕션 빌드**: < 1초 (tsup 기반)
- **테스트 실행**: < 5초 (Vitest)

### 런타임 성능
- **CLI 명령어**: < 100ms 응답
- **파일 처리**: < 50ms/파일
- **템플릿 렌더링**: < 10ms

## 🔄 CI/CD 파이프라인

### GitHub Actions (.github/workflows/)

```yaml
# 품질 게이트
- Bun 의존성 설치
- TypeScript 타입 검사
- Biome 린팅 검사
- Vitest 테스트 실행 (95%+ 커버리지)
- 빌드 검증 (ESM/CJS)

# 배포 자동화
- npm 패키지 배포
- GitHub Release 생성
- 태그 기반 버전 관리
```

## 📊 사용자 프로젝트 vs MoAI-ADK 개발

| 구분 | 사용자 프로젝트 | MoAI-ADK 개발 |
|------|----------------|---------------|
| **목적** | 범용성, 호환성 | 최고 성능 |
| **도구** | Jest, ESLint, npm | Vitest, Biome, Bun |
| **타겟** | 안정성 우선 | 혁신 우선 |
| **성능** | 표준 | 극한 최적화 |

### 설계 원칙

1. **이원화 전략**: 개발팀은 최신 도구, 사용자는 선택의 자유
2. **호환성 유지**: 사용자 템플릿은 널리 사용되는 도구 사용
3. **성능 최우선**: MoAI-ADK 자체는 속도와 효율성 극대화
4. **점진적 채택**: 사용자가 원할 때 최신 도구로 업그레이드 가능

## 🐛 디버깅 도구

### 개발 도구

```bash
# 실시간 로그 확인
DEBUG=moai:* bun run dev

# 테스트 디버깅
bun test --inspect-brk         # Node.js 디버거
bun test --reporter=verbose    # 상세 리포트

# 빌드 분석
bun run build --analyze        # 번들 크기 분석
```

### 성능 프로파일링

```bash
# 성능 측정
hyperfine "bun run build"      # 빌드 시간 측정
time bun test                  # 테스트 시간 측정

# 메모리 사용량
bun --bun run --heap-snapshot
```

## 📈 기여 가이드

### 코드 스타일

```typescript
// ✅ 권장: 명시적 타입, 간결한 로직
interface ProjectConfig {
  readonly name: string;
  readonly version: string;
}

const createProject = (config: ProjectConfig): Project => {
  return new Project(config);
};

// ❌ 지양: any 타입, 복잡한 로직
const createProject = (config: any) => {
  // 복잡한 로직...
};
```

### 커밋 규칙

```bash
# 타입: 간결한 설명 (50자 이내)
feat: Add TypeScript template validation
fix: Resolve Bun compatibility issue
perf: Optimize build pipeline by 30%
docs: Update development guide

# 본문 (선택사항): 상세 설명
# 푸터: 관련 이슈 번호
Closes #123
```

## 🎯 로드맵

### Phase 1 (완료) ✅
- [x] TypeScript 5.9.2+ 마이그레이션
- [x] Bun + Vitest + Biome 통합
- [x] 686ms 빌드 성능 달성

### Phase 2 (진행중) 🚧
- [ ] 범용 언어 지원 확대
- [ ] CLI 성능 300ms 목표
- [ ] 웹 대시보드 개발

### Phase 3 (계획) 📅
- [ ] Rust 백엔드 통합
- [ ] AI 코드 생성 도구
- [ ] 클라우드 배포 시스템

---

**현대적 개발 스택으로 최고 성능을 달성하는 MoAI-ADK!** 🚀