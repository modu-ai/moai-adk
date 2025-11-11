# CompanyAnnouncements 자동 번역 구현 가이드

## 개요

MoAI-ADK v0.20.1부터 `.claude/settings.json`의 `companyAnnouncements` 항목이 사용자 선택 언어에 따라 자동으로 번역됩니다.

## 주요 기능

### 1. 자동 번역 시스템

**위치**: `.claude/hooks/alfred/shared/utils/announcement_translator.py`

**지원 언어**:
- **ko** (한국어): 하드코딩된 번역 (22개 항목)
- **en** (영어): 기준 버전 (22개 항목)
- **ja** (일본어): 하드코딩된 번역 (22개 항목)
- **기타 모든 언어**: 영어 폴백 (향후 Claude API 동적 번역 지원 예정)

### 2. 자동 실행 트리거

`/alfred:0-project` 명령어의 **4가지 모드 모두**에서 자동 실행됩니다:

#### 1) INITIALIZATION MODE (신규 프로젝트 초기화)
- **트리거**: 언어 선택 후 `.moai/config.json` 생성 완료 시점
- **동작**: 선택된 언어로 announcements 자동 번역 및 settings.json 업데이트

#### 2) AUTO-DETECT MODE (기존 프로젝트 감지)
- **트리거**: 언어 확인 완료 시점
- **동작**: 현재 설정 언어로 announcements 재번역 (일관성 보장)

#### 3) SETTINGS MODE (설정 수정)
- **트리거**: 사용자가 언어 변경 후
- **동작**: 새 언어로 announcements 즉시 업데이트

#### 4) UPDATE MODE (템플릿 업데이트)
- **트리거**: 템플릿 최적화 완료 후
- **동작**: 현재 언어 유지하며 announcements 재적용

## 사용법

### 자동 실행 (권장)

```bash
# /alfred:0-project 실행 시 자동으로 호출됨
# 별도 실행 불필요
```

### 수동 실행 (테스트/개발용)

```bash
# 현재 config.json의 언어로 자동 번역
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py

# 특정 언어로 번역
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py ko  # 한국어
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py ja  # 일본어
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py en  # 영어
```

## 구현 세부사항

### 번역 데이터 구조

```python
# 하드코딩된 번역 딕셔너리
HARDCODED_TRANSLATIONS = {
    "en": REFERENCE_ANNOUNCEMENTS_EN,  # 22개 항목
    "ko": ANNOUNCEMENTS_KO,             # 22개 항목
    "ja": ANNOUNCEMENTS_JA              # 22개 항목
}
```

### 주요 함수

#### `auto_translate_and_update()`
- **목적**: 전체 번역 워크플로우 실행
- **단계**:
  1. `.moai/config.json`에서 `language.conversation_language` 읽기
  2. 해당 언어로 announcements 번역
  3. `.claude/settings.json` 업데이트

#### `translate_announcements(language_code)`
- **입력**: 언어 코드 (예: "ko", "en", "ja")
- **출력**: 번역된 22개 announcement 문자열 리스트
- **로직**:
  - 하드코딩된 언어라면 해당 번역 반환
  - 지원되지 않는 언어라면 영어 폴백

#### `update_settings_json(announcements)`
- **목적**: `.claude/settings.json` 파일의 `companyAnnouncements` 필드 업데이트
- **안전성**: JSON 파싱 에러 처리 포함

## 22개 기준 Announcement (영어)

1. Start with a plan: Write down what you want to build first to avoid confusion (/alfred:1-plan)
2. ✅ 5 promises: Test-first + Easy-to-read code + Clean organization + Secure + Trackable
3. Task list: Continuous progress tracking ensures nothing gets missed
4. Language separation: We communicate in your language, computers understand in English
5. Everything connected: Plan→Test→Code→Docs are all linked together
6. ⚡ Parallel processing: Independent tasks can be handled simultaneously
7. Tools first: Find the right tools before starting any work
8. Step by step: What you want→Plan→Execute→Report results
9. Auto-generated lists: Planning automatically creates task lists
10. ❓ Ask when confused: If something isn't clear, just ask right away
11. 🧪 Automatic quality checks: Code automatically verified against 5 core principles
12. Multi-language support: Automatic validation for Python, JavaScript, and more
13. ⚡ Never stops: Can continue even when tools are unavailable
14. Flexible approach: Choose between team collaboration or individual work as needed
15. 🧹 Auto cleanup: Automatically removes unnecessary items when work is complete
16. ⚡ Quick updates: New versions detected in 3 seconds, only fetch what's needed
17. On-demand loading: Only loads current tools to save memory
18. Complete history: All steps from planning to code are recorded for easy reference
19. Bug reporting: File bug reports to GitHub in 30 seconds
20. 🩺 Health check: Use 'moai-adk doctor' to instantly check current status
21. Safe updates: Use 'moai-adk update' to safely add new features
22. 🧹 When work is done: Use '/clear' to clean up conversation for the next task

## 새 언어 추가 방법

### 1. 하드코딩 번역 추가 (권장)

```python
# announcement_translator.py에 추가

ANNOUNCEMENTS_ES = [
    "Comienza con un plan: Escribe lo que quieres construir primero para evitar confusiones (/alfred:1-plan)",
    "✅ 5 promesas: Pruebas primero + Código fácil de leer + Organización limpia + Seguridad + Rastreable",
    # ... 나머지 20개 항목
]

HARDCODED_TRANSLATIONS = {
    "en": REFERENCE_ANNOUNCEMENTS_EN,
    "ko": ANNOUNCEMENTS_KO,
    "ja": ANNOUNCEMENTS_JA,
    "es": ANNOUNCEMENTS_ES  # 새 언어 추가
}
```

### 2. 번역 요구사항

- **이모지 보존**: ✅, ⚡, 🧪, 🧹, 🩺, ❓ 등 모든 이모지 그대로 유지
- **명령어 참조 유지**: `/alfred:1-plan`, `moai-adk doctor`, `/clear` 등은 그대로
- **특수 문자 유지**: →, + 등
- **톤**: 격려적이고 행동 지향적이며 사용자 친화적인 톤 유지

### 3. 테스트

```bash
# 새 언어로 번역 테스트
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py es

# settings.json 확인
cat .claude/settings.json | jq '.companyAnnouncements[0:3]'
```

## 파일 동기화

**중요**: 패키지 템플릿이 source of truth

```bash
# 로컬 변경 → 패키지 템플릿 동기화
cp .claude/hooks/alfred/shared/utils/announcement_translator.py \
   src/moai_adk/templates/.claude/hooks/alfred/shared/utils/announcement_translator.py
```

## 문제 해결

### 번역이 업데이트되지 않음

```bash
# 수동으로 재번역 실행
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py

# settings.json 권한 확인
ls -la .claude/settings.json
```

### 잘못된 언어로 번역됨

```bash
# config.json의 언어 설정 확인
cat .moai/config.json | jq '.language.conversation_language'

# 올바른 언어로 강제 번역
uv run .claude/hooks/alfred/shared/utils/announcement_translator.py ko
```

### JSON 파싱 에러

```bash
# settings.json 유효성 검사
cat .claude/settings.json | jq .

# 백업에서 복원 (필요시)
cp .moai-backups/[TIMESTAMP]/.claude/settings.json .claude/
```

## 향후 개선 계획

### Phase 1 (완료)
- ✅ 한국어, 영어, 일본어 하드코딩 번역
- ✅ 4가지 모드 자동 실행 통합
- ✅ 영어 폴백 메커니즘

### Phase 2 (예정)
- [ ] Claude API를 통한 동적 번역 (지원되지 않는 언어용)
- [ ] 번역 캐싱 시스템 (API 호출 최소화)
- [ ] 사용자 커스텀 announcement 지원

### Phase 3 (예정)
- [ ] 커뮤니티 번역 기여 시스템
- [ ] 번역 품질 검증 자동화
- [ ] 다국어 A/B 테스트 지원

## 버전 히스토리

- **v0.20.1**: 초기 자동 번역 시스템 구현
  - 한국어, 영어, 일본어 지원
  - /alfred:0-project 4가지 모드 통합
  - 영어 폴백 메커니즘

## 관련 파일

- **구현**: `.claude/hooks/alfred/shared/utils/announcement_translator.py`
- **템플릿**: `src/moai_adk/templates/.claude/hooks/alfred/shared/utils/announcement_translator.py`
- **명령어**: `src/moai_adk/templates/.claude/commands/alfred/0-project.md` (Lines 538-663)
- **설정**: `.claude/settings.json` (`companyAnnouncements` field)
- **언어 설정**: `.moai/config.json` (`language.conversation_language`)

## 참고 자료

- 0-project.md 섹션: "🌍 Language-Specific CompanyAnnouncements"
- CLAUDE.md 섹션: "🌍 Alfred's Language Boundary Rule"
- MoAI-ADK 언어 지원 전략 문서
