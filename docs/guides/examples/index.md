# 실습 예제

MoAI-ADK의 SPEC-First TDD 방법론을 실전에서 학습할 수 있는 실습 예제들을 제공합니다.

---

## 📚 제공 예제

### 1. 풀스택 ToDo 앱 만들기

> **난이도**: 🟢 초급
> **소요시간**: 약 4-5시간
> **학습 목표**: MoAI-ADK 전체 워크플로우 체득

MoAI-ADK를 사용하여 학습용 ToDo 애플리케이션을 처음부터 끝까지 만드는 튜토리얼입니다.

**기술 스택**:

- Frontend: Vite + React + TypeScript + Tailwind CSS
- Backend: FastAPI + SQLAlchemy 2.0 + Alembic
- Database: SQLite
- Deploy: 로컬 개발 (Docker 선택사항)

**학습 내용**:

- ✅ 프로젝트 초기화 (`/alfred:0-project`)
- ✅ SPEC 작성 (`/alfred:1-spec`)
- ✅ TDD 구현 (`/alfred:2-build`)
- ✅ 문서 동기화 (`/alfred:3-sync`)
- ✅ EARS 요구사항 작성법
- ✅ @TAG 시스템 활용

**시작하기**: [ToDo 앱 튜토리얼](./todo-app/index.md)

---

## 🎯 학습 경로

### 초급 학습자

처음 MoAI-ADK를 사용하신다면 다음 순서로 학습하세요:

1. **[설치 가이드](../../getting-started/installation.md)** - MoAI-ADK 설치
2. **[빠른 시작](../../getting-started/quick-start.md)** - 5분 만에 시작하기
3. **[ToDo 앱 튜토리얼](./todo-app/index.md)** - 실전 프로젝트 완성
4. **[워크플로우 가이드](../workflow/overview.md)** - 심화 학습

### 중급 학습자

MoAI-ADK 기본 개념을 이해하셨다면:

1. **[SPEC-First TDD](../concepts/spec-first-tdd.md)** - 방법론 이해
2. **[EARS 가이드](../concepts/ears-guide.md)** - 요구사항 작성법
3. **[TAG 시스템](../concepts/tag-system.md)** - 추적성 관리
4. **[에이전트 활용](../agents/overview.md)** - 전문 에이전트 활용

---

## 💡 예제별 비교

| 예제 | 난이도 | 시간 | Frontend | Backend | 주요 학습 내용 |
|------|--------|------|----------|---------|--------------|
| **ToDo 앱** | 🟢 초급 | 4-5시간 | React | FastAPI | 전체 워크플로우, CRUD, TDD |

> 더 많은 예제가 추가될 예정입니다!

---

## 🛠️ 실습 전 준비사항

### 필수 도구

```bash
# MoAI-ADK CLI
npm install -g moai-adk

# Claude Code (권장)
# https://docs.anthropic.com/claude-code

# Git
git --version

# Node.js 18+
node --version

# Python 3.11+ (Backend 예제)
python --version
```

### 권장 도구

- **VS Code** + Claude Code Extension
- **Docker Desktop** (배포 실습용)
- **Postman** 또는 **Thunder Client** (API 테스트용)

---

## 📖 학습 지원

### 문서

- [CLAUDE.md](https://github.com/modu-ai/moai-adk/blob/main/CLAUDE.md) - Alfred SuperAgent 가이드
- [development-guide.md](https://github.com/modu-ai/moai-adk/blob/main/.moai/memory/development-guide.md) - 개발 가이드
- [spec-metadata.md](https://github.com/modu-ai/moai-adk/blob/main/.moai/memory/spec-metadata.md) - SPEC 메타데이터 표준

### 커뮤니티

- **GitHub Issues**: [문제 보고 및 질문](https://github.com/modu-ai/moai-adk/issues)
- **Discussions**: [커뮤니티 토론](https://github.com/modu-ai/moai-adk/discussions)

---

**시작하기**: [ToDo 앱 튜토리얼](./todo-app/index.md)로 첫 프로젝트를 만들어보세요! 🚀
