# MoAI-ADK 개발 가이드라인 (로컬 프로젝트 전용)

**MoAI-ADK 프로젝트 개발 시 필수 규칙 및 설정**

---

## 🚀 패키지 동기화 규칙

**진실의 원천**: `src/moai_adk/templates/` (패키지 템플릿)
**동기화 방향**: 템플릿 변경 → 로컬 프로젝트 (즉시)
**읽기 전용**: 로컬 `.claude/` 파일 (정적 인프라)

**GitHub 릴리스 포맷**: `.claude/commands/moai/99-release.md` (Lines 255-283) 참조

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

## 🛠️ MoAI-ADK 개발 필수 규칙

### UV 전용 실행

**규칙**: 모든 스크립트는 `uv run` 사용 (❌ `python`, `pip` 직접 호출 금지)

**필수 설정**: `.moai/config/config.json` (없으면 statusline 버전 표시 불가)

**초기화**:
```bash
mkdir -p .moai/config
cp src/moai_adk/templates/.moai/config/config.json .moai/config/
```

**필수 필드**:
```json
{
  "moai": {"version": "0.26.0"},
  "language": {"conversation_language": "ko"},
  "project": {"name": "MoAI-ADK", "owner": "GoosLab"}
}
```

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

**범위**: MoAI-ADK 프로젝트 개발 가이드라인만 (로컬 전용)
**행 수**: 101줄 (최적화 전: 187줄, -46%)
**마지막 업데이트**: 2025-11-19
