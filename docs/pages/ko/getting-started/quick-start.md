---
title: "5분 빠른 시작"
description: "5분 만에 MoAI-ADK로 첫 프로젝트를 시작하는 방법 - 설치부터 첫 기능 개발까지"
---

# 5분 빠른 시작

5분 안에 MoAI-ADK로 첫 프로젝트를 생성하고 실행해보세요.

## ⚡ 5단계 빠른 시작

### 1단계: 설치 (1분)

```bash
# uv tool로 MoAI-ADK 전역 설치
uv tool install moai-adk

# 설치 확인
moai-adk --version
```

### 2단계: 프로젝트 생성 (30초)

```bash
# 새 프로젝트 생성
moai-adk init my-first-project
cd my-first-project
```

### 3단계: Claude Code 실행 (30초)

```bash
# Claude Code로 프로젝트 열기
claude-code .
```

### 4단계: 프로젝트 설정 (1분)

Claude Code에서 다음 명령을 실행하세요:

```bash
/alfred:0-project
```

Alfred가 자동으로:
- ✅ 프로젝트 메타데이터 설정
- ✅ 개발 언어 감지
- ✅ Git 전략 구성
- ✅ 다국어 시스템 활성화

### 5단계: 첫 기능 개발 (2분)

```bash
# SPEC 작성
/alfred:1-plan "간단한 계산기 기능"

# 자동화된 TDD로 구현
/alfred:2-run CALC-001

# 문서 동기화
/alfred:3-sync
```

## 🎉 결과

5분 후 다음을 얻게 됩니다:

- ✅ **명확한 SPEC 문서**: 구조화된 요구사항
- ✅ **종합적인 테스트**: 87.84%+ 테스트 커버리지
- ✅ **구현 코드**: 모범 사례를 따른 코드
- ✅ **업데이트된 문서**: 자동으로 동기화된 문서
- ✅ **Git 히스토리**: @TAG 참조가 포함된 커밋

## 실제 실행 예시

### 프로젝트 생성 출력

```bash
$ moai-adk init my-first-project

🗿 MoAI-ADK v0.23.0
Initializing MoAI-ADK project...

✅ Initialization Completed Successfully!
────────────────────────────────────────────────────────────
📊 Summary:
  📁 Location:   /Users/user/my-first-project
  🌐 Language:   Auto-detect (use /alfred:0-project)
  🔧 Mode:       personal
  🌍 Locale:     ko
  📄 Files:      47 created
  ⏱️  Duration:   1,234ms
────────────────────────────────────────────────────────────

🚀 Next Steps:
  1. Run /alfred:0-project in Claude Code for full setup
     (Configure: mode, language, report generation, etc.)
  2. Start developing with MoAI-ADK!
```

### Claude Code 설정

```bash
Claude Code> /alfred:0-project

📋 Configuration Health Check:
✅ Project configuration complete
✅ Recent setup: Just now
✅ Version match: 0.23.0
✅ Multi-language system: Active
✅ Expert delegation: Ready

All systems are healthy! 🎉
```

### 첫 기능 개발

```bash
Claude Code> /alfred:1-plan "간단한 계산기 기능"

🎯 SPEC 계획 생성 완료:
- SPEC-CALC-001: 간단한 계산기 기능
- 요구사항: 사칙연산, 입력 유효성 검사, 오류 처리
- 테스트 케이스: 15개 포함
- 예상 개발 시간: 30분

Claude Code> /alfred:2-run CALC-001

🔄 TDD 사이클 실행:
1️⃣ RED: 테스트 작성 완료 (15개 테스트)
2️⃣ GREEN: 최소 구현 완료
3️⃣ REFACTOR: 코드 품질 개선 완료
4️⃣ SYNC: 문서 자동 동기화 완료

✅ 기능 개발 완료!
- 테스트 커버리지: 92.3%
- 코드 품질: TRUST 5 준수
- 생성된 파일: 5개
```

## 확인 단계

### 프로젝트 구조 확인

```bash
# 프로젝트 구조
tree my-first-project -I '__pycache__|node_modules'

my-first-project/
├── .claude/
│   ├── agents/
│   ├── commands/
│   ├── skills/
│   └── hooks/
├── .moai/
│   ├── config.json
│   ├── specs/
│   │   └── SPEC-CALC-001/
│   │       ├── spec.md
│   │       ├── plan.md
│   │       └── acceptance.md
│   └── reports/
├── src/
│   └── calculator.py
├── tests/
│   └── test_calculator.py
├── docs/
│   └── api/
├── README.md
├── CHANGELOG.md
└── .git/
```

### 테스트 실행

```bash
# 프로젝트에서 테스트 실행
python -m pytest tests/

# 결과 예시
========== test session starts ==========
collected 15 items

tests/test_calculator.py .............
15 passed in 0.123s

92.3% coverage
```

### Git 히스토리 확인

```bash
# Git 커밋 확인
git log --oneline -5

feat(calculator): Add basic arithmetic operations with TDD
test(calculator): Add comprehensive test suite for calculator
refactor(calculator): Improve error handling and input validation
docs(calculator): Auto-sync documentation with implementation
feat(SPEC-CALC-001): Complete calculator feature with full coverage
```

## 🆕 v0.23.1 최신 기능 활용하기

### BaaS 플랫폼 빠른 통합

MoAI-ADK v0.23.1은 **12개 BaaS 플랫폼**을 완전 지원합니다:

```bash
# Supabase 통합 예제
/alfred:1-plan "Supabase를 활용한 실시간 채팅 기능"
/alfred:2-run CHAT-001

# Firebase 통합 예제
/alfred:1-plan "Firebase Auth를 활용한 소셜 로그인"
/alfred:2-run AUTH-002
```

**지원 플랫폼**: Supabase, Firebase, Vercel, Cloudflare, Auth0, Convex, Railway, Neon, Clerk, PocketBase, Appwrite, Parse

### Expert Delegation System 활용

```bash
# 자동 전문가 할당 (v0.23.1)
/alfred:0-project  # project-manager 자동 할당
/alfred:1-plan "복잡한 요구사항"  # spec-builder 자동 할당
/alfred:2-run SPEC-001  # tdd-implementer 자동 할당
```

**60% 상호작용 감소**: Alfred가 자동으로 적절한 전문가를 선택합니다.

### 292 Skills 활용

```bash
# Skills 목록 확인
moai-adk skills list

# 특정 Skill 정보 확인
moai-adk skills info moai-baas-supabase
```

## 다음 단계

빠른 시작을 완료했습니다! 이제 다음을 할 수 있습니다:

### 실전 학습 자료

1. **[Tutorial 1: REST API 개발](/ko/tutorials/tutorial-01-rest-api)** - 30분, 초보자 추천
2. **[Tutorial 2: JWT 인증 구현](/ko/tutorials/tutorial-02-jwt-auth)** - 1시간, 실전 보안
3. **[Tutorial 4: Supabase 통합](/ko/tutorials/tutorial-04-baas-supabase)** - 1시간, BaaS 활용

### 코드 예제 라이브러리

- **[REST API 예제](/ko/examples/rest-api)**: CRUD, 인증, 에러 처리
- **[인증 예제](/ko/examples/authentication)**: JWT, OAuth, Session
- **[BaaS 예제](/ko/examples/baas)**: Supabase, Firebase 통합

### 심화 학습

- **[초보자 가이드](/ko/guides/beginner)**: 체계적인 학습 경로
- **[중급자 가이드](/ko/guides/intermediate)**: 고급 패턴과 실전 활용
- **[Skills 생태계](/ko/skills/ecosystem-upgrade-v4)**: 292 Skills 완전 가이드

## 빠른 참조

### 유용한 Alfred 명령어

| 명령어 | 목적 | 사용법 |
|--------|------|--------|
| `/alfred:0-project` | 프로젝트 설정 | 초기화 또는 재설정 |
| `/alfred:1-plan` | SPEC 작성 | 기능 계획 및 요구사항 |
| `/alfred:2-run` | TDD 구현 | 자동화된 개발 |
| `/alfred:3-sync` | 동기화 | 문서 및 리포트 생성 |

### 일반적인 작업

```bash
# 새 기능 추가
/alfred:1-plan "새 기능 이름"
/alfred:2-run FEATURE-ID

# 버그 수정
/alfred:1-plan "버그 수정: 설명"
/alfred:2-run BUG-ID

# 프로젝트 상태 확인
/alfred:status

# 리포트 생성
/alfred:report
```

## 도움말

문제가 있나요?

- 📖 [설치 가이드](./installation): 자세한 설치 지침
- 🔧 [문제 해결](../troubleshooting): 일반적인 문제 해결
- 💬 [커뮤니티](https://github.com/modu-ai/moai-adk/discussions): 도움 요청

---

🎉 **축하합니다!** 이제 MoAI-ADK로 생산적인 개발을 시작할 준비가 되었습니다.