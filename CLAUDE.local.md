# 로컬 프로젝트 설정 & 특화 규칙

**MoAI-ADK 로컬 프로젝트 전용 구성 및 규칙**

> **참고**: 언어 규칙 및 보안 정책은 CLAUDE.md의 "언어 아키텍처"와 "보안 및 모범 사례" 섹션을 참조하세요.

---

## 🚀 GitHub 릴리스 & 패키지 동기화

### 릴리스 작성자 형식

- **작성자**: GoosLab (not @goos)
- **공동 작성자**: `🎩 Alfred@MoAI` (이모지 + 이름, 이메일 없음)
- **푸터**: `🤖 Generated with [Claude Code](https://claude.com/claude-code)\n\nCo-Authored-By: 🎩 Alfred@MoAI`

### 멘션 방지

마크다운에서 `@TAG`, `@SPEC` 회피 → `TAG`, `SPEC` 또는 `` `@TAG` `` 사용

### 패키지 동기화 규칙

**진실의 원천**: `src/moai_adk/templates/` (패키지 템플릿)
**동기화 방향**: 템플릿 변경 → 로컬 프로젝트 (즉시)
**읽기 전용**: 로컬 `.claude/` 파일 (정적 인프라)

**이유**:
- 패키지: 변경 사항이 모든 사용자에게 즉시 배포
- 로컬: Claude Code 실행을 위한 구체적 경로

---

## 🔧 템플릿 변수 치환 규칙

**규칙**: 패키지 템플릿의 `{{PROJECT_DIR}}` 변수는 **절대 치환하지 않습니다**.

### 패턴 비교

| 컨텍스트 | 경로 형식 | 예시 |
|---------|---------|------|
| **패키지 템플릿** | `{{PROJECT_DIR}}/...` | `uv run {{PROJECT_DIR}}/.moai/scripts/statusline.py` |
| **로컬 프로젝트** | 상대 경로 `./...` | `uv run .moai/scripts/statusline.py` |

### 왜?

- **패키지**: 변수 유지 → 사용자 환경에서 자동 치환
- **로컬**: 직접 경로 → Claude Code 즉시 실행
- **SSOT**: 패키지 템플릿이 단일 진실의 원천

---

## 🧙 Yoda 시스템

**상태**: 로컬 전용 강의 자료 생성기 (MoAI-ADK 패키지와 함께 배포되지 않음)

### 핵심 원칙

"바퀴를 재발명하지 말고, 기존의 도구를 현명하게 재사용하자"

- 기존 Skills 재사용 (`moai-document-processing`, `moai-yoda-content-generator`)
- MCP 도구 직접 활용 (Notion, Context7)
- 3가지 표준 템플릿 활용 (education/presentation/workshop)

### 빠른 사용법

```bash
# 대화형 모드 (권장)
/yoda:generate

# 명령줄 모드
/yoda:generate --topic "React Hooks" --format education --output pdf,pptx
```

**출력**: `.moai/yoda/output/{topic}.{md,pdf,pptx,docx}`

### 아키텍처 (Skill 기반)

Commands → Agents → Skills 패턴:

- **Command**: `/yoda:generate` (인자 파싱, 사용자 인터페이스)
- **Agent**: `yoda-master` (워크플로우 오케스트레이션)
- **Skills**:
  - `moai-yoda-content-generator` (콘텐츠 생성, 3,356+ 줄)
  - `moai-document-processing` (형식 변환)
  - `moai-yoda-system` (템플릿 + 메타데이터 관리)

### 성능

- 소형 (10페이지): < 1분
- 중형 (50페이지): < 3분
- 대형 (100페이지): < 5분
- 토큰 효율성: 모놀리식 대비 56% 감소

### 문서

- **Command**: `.claude/commands/yoda/generate.md` (350줄)
- **Agent**: `.claude/agents/yoda-master.md` (600줄)
- **Skill**: `.claude/skills/moai-yoda-content-generator/SKILL.md` (3,356+ 줄)

### 파일 구성

- **템플릿**: `.claude/skills/moai-yoda-system/templates/` (education.md, presentation.md, workshop.md)
- **출력**: `.moai/yoda/output/` (gitignored, 로컬 전용)

---

## 🛠️ MoAI-ADK: UV 전용 실행

**규칙 1**: moai-adk 프로젝트는 **uv만 사용**
**규칙 2**: `.moai/config/config.json` **필수** (없으면 statusline 버전 표시 불가)

### 명령어 규칙

| 작업 | ❌ 금지 | ✅ 필수 |
|------|--------|--------|
| **스크립트 실행** | `python script.py` | `uv run script.py` |
| **모듈 실행** | `python -m module` | `uv run -m module` |
| **패키지 관리** | `pip install/remove` | `uv add/remove` |
| **의존성 동기화** | `poetry install` | `uv sync` |
| **빌드** | `poetry build` | `uv build` |
| **테스트/린팅** | `pytest`, `mypy` | `uv run pytest`, `uv run mypy` |

### UV 전용 이유

- **일관성**: 전체 moai-adk 생태계가 uv 사용
- **격리**: 시스템 Python/pip과 분리
- **재현성**: 잠금 파일 기반 (uv.lock) 정확한 의존성

### 필수 설정 파일

**경로**: `.moai/config/config.json`

**필수 필드**:
```json
{
  "moai": {"version": "0.26.0"},
  "language": {"conversation_language": "ko"},
  "project": {"name": "MoAI-ADK", "owner": "GoosLab"}
}
```

**초기화**: `mkdir -p .moai/config && cp src/moai_adk/templates/.moai/config/config.json .moai/config/`

**누락 시 발생**:
- Statusline에서 "Ver unknown" 표시
- VersionReader 실패
- SessionStart 훅 불완전

---

## 📋 로컬 환경 체크리스트

### 필수 설정

- [ ] `.moai/config/config.json` 존재 및 올바른 형식
- [ ] `uv sync` 완료 (의존성 설치)
- [ ] `.gitignore`에 `.env*`, `.vercel/` 등 보안 파일 포함
- [ ] Git 사용자 정보 설정 (`git config user.name` / `user.email`)

### 개발 시작

```bash
# 1. 의존성 동기화
uv sync

# 2. 프로젝트 상태 확인
uv run .moai/scripts/statusline.py

# 3. 테스트 실행
uv run pytest

# 4. SPEC 생성 (새 기능)
/moai:1-plan "기능 설명"

# 5. 토큰 절약을 위해 세션 초기화
/clear

# 6. TDD 구현
/moai:2-run SPEC-XXX
```

---

**범위**: MoAI-ADK 프로젝트 개발 전용
**위치**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.local.md`
**Git**: 커밋되지 않음 (gitignore)
**마지막 업데이트**: 2025-11-19
