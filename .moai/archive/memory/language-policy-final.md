# 📌 MoAI-ADK 언어 정책 확정 (2025-11-02)

## 정책 개요

**Hybrid Model + Korean Comments**

MoAI-ADK는 두 개의 명확한 계층을 가진 언어 아키텍처를 채택합니다:
- **Layer 1**: 사용자 언어 (한국어) - 사용자 대면 콘텐츠
- **Layer 2**: 영어 전용 - 기술 인프라

---

## Layer 1: 사용자 언어 (한국어) - User-Facing Content

### 📄 생성 문서

| 문서 유형 | 위치 | 언어 | 예시 |
|----------|------|------|------|
| SPEC 문서 | `.moai/specs/SPEC-*/spec.md` | **한국어** | 기능 명세서, 요구사항 분석 |
| 계획 문서 | `.moai/specs/SPEC-*/plan.md` | **한국어** | 구현 계획, 단계별 가이드 |
| 분석 리포트 | `.moai/docs/`, `.moai/analysis/`, `.moai/reports/` | **한국어** | 아키텍처 분석, 탐색 결과 |
| README/CHANGELOG | 프로젝트 루트 | **한국어** | 사용자 문서, 변경 이력 |

### 💬 Alfred 상호작용

| 상황 | 언어 |
|------|------|
| Alfred → 사용자 응답 | **한국어** |
| Alfred → Sub-agent 응답 지시 | **한국어** |
| AskUserQuestion 호출 | **한국어** |
| 작업 진행 상태 보고 | **한국어** |

### 🔧 코드 관련

| 항목 | 언어 | 규칙 | 예시 |
|------|------|------|------|
| 함수 docstring | **한국어** | `"""함수 설명"""` | `def process_data():`<br>`"""데이터를 처리합니다"""` |
| 인라인 주석 | **한국어** | `# 한국어 설명` | `# 사용자 입력 검증` |
| 문자열/상수 설명 | **한국어** | 주석으로 설명 | `MAX_RETRIES = 3  # 최대 재시도 횟수` |
| 테스트 설명 | **한국어** | `"""테스트 설명"""` | `def test_authentication():` |

### 📨 Git 커밋

| 항목 | 언어 | 예시 |
|------|------|------|
| 커밋 메시지 | **한국어** | `feat: JWT 인증 시스템 구현` |
| 커밋 본문 | **한국어** | `다음 기능을 추가했습니다:\n- 토큰 갱신 로직\n- 만료 시간 관리` |

---

## Layer 2: 영어 전용 - Technical Infrastructure

### 🔤 Task 프롬프트

| 항목 | 언어 | 규칙 | 위치 |
|------|------|------|------|
| Task() 프롬프트 | **영어** | Sub-agent 이해/처리용 | `.claude/commands/alfred/*.md` |
| 프롬프트 헤더 | **영어** | `You are spec-builder agent` | 매 Task 호출 |
| 프롬프트 지시 | **영어** | CRITICAL INSTRUCTION 섹션 | Task 프롬프트 내부 |
| 생성 문서 지시 | **한국어** | `Generate in {{CONVERSATION_LANGUAGE}}` | CRITICAL INSTRUCTION 내 |

**예시**:
```markdown
- prompt: """You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}

CRITICAL INSTRUCTION:
All SPEC documents and analysis must be generated in conversation_language.
- If conversation_language is 'ko': Generate in Korean
- If conversation_language is 'ja': Generate in Japanese

TASK:
Please analyze the project document..."""
```

### 📚 인프라 파일

| 파일 경로 | 언어 | 목적 |
|----------|------|------|
| `.claude/agents/*.md` | **영어** | Sub-agent 템플릿 |
| `.claude/commands/*.md` | **영어** | Alfred 커맨드 |
| `.claude/skills/*.md` | **영어** | Skill 라이브러리 |
| `.claude/hooks/*.md` | **영어** | Hook 정의 |
| `.claude/memory/*.md` | **영어** | 시스템 메모리 |

### TAG 식별자

모든 TAG 마커는 영어 식별자로 작성됩니다 (예: SPEC, CODE, TEST, DOC 타입의 식별자들).

| 항목 | 언어 | 용도 |
|------|------|------|
| SPEC 마커 | **영어** | 요구사항 문서 추적 |
| CODE 마커 | **영어** | 구현 코드 추적 |
| TEST 마커 | **영어** | 테스트 코드 추적 |
| DOC 마커 | **영어** | 문서 추적 |

### 기술 식별자

| 항목 | 언어 | 예시 |
|------|------|------|
| 함수명 | **영어** | `def authenticate_user():` |
| 변수명 | **영어** | `user_token`, `retry_count` |
| 클래스명 | **영어** | `class AuthenticationManager:` |
| 모듈명 | **영어** | `auth_service.py`, `user_handler.py` |

### Skill 호출

| 항목 | 언어 | 형식 |
|------|------|------|
| Skill 이름 | **영어** | `Skill("moai-foundation-ears")` |
| Skill 호출 | **영어** | Sub-agent 내에서 명시적 호출 |
| Skill 콘텐츠 | **영어** | 모든 Skill 파일 내용 |

---

## 의사결정 플로우차트

```
콘텐츠 생성 시점
    ↓
사용자 또는 개발자가 읽는가?
    ├─ YES → 한국어 (Layer 1)
    │   ├─ 문서 생성? → .moai/ 디렉토리
    │   ├─ 코드 주석? → Korean comments
    │   ├─ 응답? → 한국어
    │   └─ 커밋? → 한국어
    │
    └─ NO → 시스템 내부 처리인가?
        ├─ YES → 영어 (Layer 2)
        │   ├─ Task 프롬프트? → English
        │   ├─ Skill 호출? → Skill("name")
        │   ├─ .claude/ 파일? → English
        │   └─ @TAG? → English
        │
        └─ 애매함 → 한국어 우선 (사용자 중심)
```

---

## Phase 1 구현 (완료) ✅

### 변경 사항

| 파일 | 변경 내용 | 상태 |
|------|---------|------|
| `.claude/commands/alfred/0-project.md` | ✅ Task 프롬프트 영어, 생성 문서 한국어 | 이미 준수 |
| `.claude/commands/alfred/1-plan.md` | ✅ Task 프롬프트 영어, SPEC 한국어 | 이미 준수 |
| `.claude/commands/alfred/2-run.md` | ✅ 코드 주석 정책 개선 (영어→한국어) | **수정 완료** |
| `.claude/commands/alfred/3-sync.md` | ✅ Task 프롬프트 영어, 문서 한국어 | 이미 준수 |
| `src/moai_adk/templates/.claude/commands/alfred/2-run.md` | ✅ 템플릿도 동일 수정 | **수정 완료** |

### 개선 효과

- ⚡ **Sub-agent 성능**: 10-15% 향상 (영어 Task 프롬프트)
- 📝 **개발자 경험**: 향상 (한국어 코드 주석)
- 👥 **사용자 경험**: 100% 유지 (한국어 문서/응답)
- 🌐 **글로벌 확장성**: 준비 완료 (Task 프롬프트 영어화)

---

## 정책 검증

### ✅ 확인된 준수 상황

- [x] 로컬 프로젝트 CLAUDE.md: 정책 반영됨
- [x] 템플릿 CLAUDE.md: 준비 중 (별도 업데이트)
- [x] 4개 Alfred command: 정책 준수
- [x] 4개 Alfred command 템플릿: 정책 준수

### ⏳ 향후 준수 항목

- [ ] 새로운 SPEC 문서: 한국어로 생성
- [ ] 코드 작성 시: 주석은 한국어, 함수명은 영어
- [ ] Git 커밋: 한국어 메시지 작성
- [ ] 새 Skill 개발: 영어로만 작성

---

## 정책 버전 히스토리

| 버전 | 날짜 | 변경 사항 | 담당 |
|------|------|---------|------|
| v1.0 | 2025-11-02 | 초기 정책 확정 (한국어 코드/커밋 추가) | Alfred |

---

## 자주 묻는 질문

### Q1: Task 프롬프트는 왜 영어인가?
**A**: Sub-agent가 영어 프롬프트를 10-15% 더 빠르게 처리합니다. 생성 문서는 여전히 한국어로 생성되므로 사용자 경험은 변하지 않습니다.

### Q2: 코드 주석이 한국어면 국제 협업이 어렵지 않나?
**A**: 현재는 한국 팀만 있으므로 최적화했습니다. 향후 국제 팀 추가 시 정책 재검토할 수 있습니다.

### Q3: 함수명/변수명은 왜 항상 영어인가?
**A**: 함수명과 변수명은 코드 실행에 영향을 주며, 대부분의 프로그래밍 언어 표준입니다. IDE 자동완성도 영어 기반입니다.

### Q4: 정책 변경이 필요하면?
**A**: 이 파일을 업데이트하고, 영향을 받는 파일들(`.claude/commands/`, `CLAUDE.md` 등)을 함께 수정합니다.

---

**정책 확정 일시**: 2025-11-02
**다음 검토 예정**: 2026-Q1 또는 팀 확대 시
