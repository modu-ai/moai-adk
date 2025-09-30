# MoAI-ADK Project Structure

## 🏗️ 프로젝트 구조 분리

### 🛠️ 개발환경 (Development Environment)

**목적**: MoAI-ADK 패키지 자체 개발 및 유지보수

```
/Users/goos/MoAI/MoAI-ADK/
├── .claude/                    # Claude Code 개발 설정
│   ├── agents/                 # 개발용 에이전트
│   ├── commands/               # 개발용 명령어
│   ├── hooks/                  # 🆕 로컬 실행 훅들
│   │   ├── security/
│   │   ├── session/
│   │   └── workflow/
│   ├── output-styles/
│   └── settings.json           # 🆕 .claude/hooks 경로 참조
├── .moai/                      # MoAI 개발 메타데이터
│   ├── config.json             # 프로젝트 설정
│   # TAG는 소스코드에만 존재 (CODE-FIRST)
│   # 별도 폴더 불필요: rg '@TAG' 명령으로 직접 스캔
│   ├── project/
│   ├── specs/
│   └── reports/
├── CLAUDE.md                   # 개발용 Claude Code 지침
├── src/                        # Python 소스 코드 (레거시)
└── tests/                      # Python 테스트 (레거시)
```

### 📦 배포환경 (Distribution Package)

**목적**: 사용자가 설치하여 사용할 수 있는 TypeScript 패키지

```
/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/
├── package.json                # npm 배포 설정
├── dist/                       # 컴파일된 JavaScript 파일들
│   ├── cli/                    # CLI 진입점
│   ├── core/                   # 핵심 로직
│   ├── claude/                 # Claude Code 템플릿
│   │   └── hooks/              # 사용자용 훅 템플릿
│   └── scripts/                # 유틸리티 스크립트
├── src/                        # TypeScript 소스 코드
│   ├── cli/
│   ├── core/
│   ├── claude/
│   └── scripts/
└── README.md                   # 사용자 가이드
```

## 🔄 동기화 전략

### 동일해야 하는 요소들

1. **훅 (Hooks)**
   - 개발: `.claude/hooks/` (JavaScript 실행 파일)
   - 배포: `moai-adk-ts/src/claude/hooks/` (TypeScript 소스)
   - 배포: `moai-adk-ts/dist/claude/hooks/` (컴파일된 JavaScript)

2. **명령어 (Commands)**
   - 개발: `.claude/commands/moai/`
   - 배포: `moai-adk-ts/src/claude/commands/moai/`

3. **에이전트 (Agents)**
   - 개발: `.claude/agents/moai/`
   - 배포: `moai-adk-ts/src/claude/agents/moai/`

4. **스크립트 (Scripts)**
   - 개발: `.moai/scripts/`
   - 배포: `moai-adk-ts/src/scripts/`

### 차이점

- **개발환경**: TypeScript를 빌드하여 `.claude/hooks`에 JavaScript 파일 배치
- **배포환경**: 사용자가 `npm install moai-adk`로 설치하여 사용

## 🚀 배포 프로세스

1. TypeScript 소스 수정
2. `npm run build` 실행
3. 컴파일된 JavaScript를 개발환경 `.claude/hooks`에 복사
4. 테스트 및 검증
5. npm 패키지 배포