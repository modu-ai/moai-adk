# companyAnnouncements Update Process Fix Summary

## 문제 정의 (Problem Definition)

`/alfred:0-project update` 프로세스에서 `companyAnnouncements` 섹션이 누락되는 문제:

1. **STEP 0-UPDATE.2** (템플릿 비교)에서 companyAnnouncements 비교 로직 누락
2. **STEP 0-UPDATE.5** (.claude/settings.json 병합)에서 companyAnnouncements 병합 로직 누락
3. 런타임 번역 로직(2300-2340행)만 존재하고 실제 update 프로세스와 연동 안 됨

## 해결책 (Solution)

### 1. STEP 0-UPDATE.2 수정 (Template Comparison)

**변경사항**:
- `companyAnnouncements` 섹션 언어 감지 로직 추가
- 현재 `conversation_language`를 `.moai/config.json`에서 읽기
- 백업된 announcements가 번역된 것인지 감지 (영어 외 언어)

```diff
4. **Compare settings.json**:
   - Check: Custom environment variables in old backup
   - Check: Custom permissions in `permissions.allow` array
   - Check: Custom hooks in `hooks` section
+  - Check: `companyAnnouncements` section and detect language
+    - Read current conversation_language from `.moai/config.json`
+    - Check if announcements in backup are translated (non-English)
   - Identify: User-added configurations
   - Store: Settings that need preservation
```

### 2. STEP 0-UPDATE.5 수정 (Settings Merge)

**변경사항**:
- `companyAnnouncements` 처리 로직 추가
- 언어에 따른 보존/번역 전략 구현

**처리 로직**:
```diff
2. **Merge .claude/settings.json**:
   - Read: New template from `src/moai_adk/templates/.claude/settings.json`
   - Read: User's custom permissions from backup
   - Read: User's custom environment variables from backup
+  - Read: Current conversation_language from `.moai/config.json`
+  - Read: User's companyAnnouncements from backup
+  - **Handle companyAnnouncements**:
+    - **If user's announcements are already translated** (non-English):
+      - Keep user's translated announcements (preserve existing translations)
+    - **If user's announcements are English** AND conversation_language != "en":
+      - Translate English template announcements to user's conversation_language
+      - Use runtime translation process (documented in lines 2300-2340)
+    - **If conversation_language == "en"**:
+      - Use English template announcements as-is
   - Merge strategy:
     ```
     {
+      "companyAnnouncements": [preserve/translate based on language logic above],
       "hooks": [merge user's custom hooks with new defaults],
       "permissions": {
         "allow": [merge user's + new defaults, remove duplicates],
         "ask": [keep new defaults],
         "deny": [keep new defaults]
       },
       "environmentVariables": [merge user's custom vars with new defaults]
     }
     ```
```

### 3. 비교 및 완료 보고서 업데이트

**변경사항**:
- 비교 보고서에 companyAnnouncements 상태 표시
- 완료 보고서에 번역 보존 상태 표시

```diff
[IF settings.json has customizations]
   ✓ .claude/settings.json:
     - Custom permissions: [list count] items
     - Custom environment variables: [list count] items
     - Custom hooks: [list if any]
+    - companyAnnouncements: [detected language] translations preserved

✓ .claude/settings.json
  - New default settings applied
  - Your custom permissions preserved
  - Your environment variables preserved
+  - Your companyAnnouncements translations preserved
```

### 4. 템플릿 파일 업데이트

**src/moai_adk/templates/.claude/settings.json 변경사항**:

1. **Hook 경로 변수 수정**:
   - `{{HOOK_PROJECT_DIR}}` → `$CLAUDE_PROJECT_DIR`
   - Claude Code 환경 변수 표준에 맞춤

2. **PostToolUse 훅 추가**:
   ```json
   {
     "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/post_tool__enable_streaming_ui.py",
     "type": "command",
     "description": "Ensure streaming UI indicators and progress displays are enabled"
   },
   {
     "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/core/post_tool__multilingual_linting.py",
     "type": "command",
     "description": "Run multilingual linting checks on modified files (Python, JavaScript, TypeScript, Go, Rust, Java, Ruby, PHP)"
   },
   {
     "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/core/post_tool__multilingual_formatting.py",
     "type": "command",
     "description": "Run multilingual code formatting on modified files"
   }
   ```

3. **UI 설정 추가**:
   ```json
   "spinnerTipsEnabled": true,
   "outputStyle": "streaming"
   ```

## 동작 방식 (How It Works)

### 언어 감지 및 처리 흐름

```
1. /alfred:0-project update 실행
   ↓
2. STEP 0-UPDATE.2: 백업된 settings.json 분석
   - conversation_language 읽기 (ko, en, ja, 등)
   - companyAnnouncements 언어 감지
   - 번역된 announcements 식별
   ↓
3. STEP 0-UPDATE.5: 스마트 병합
   - 이미 번역된 announcements → 보존
   - 영어 announcements + 비영어 언어 → 번역
   - 영어 announcements + 영어 언어 → 그대로 사용
   ↓
4. 런타임 번역 프로세스 연동
   - lines 2300-2340에 문서화된 번역 로직 활용
   - Alfred의 언어 정책 준수
```

### 예제 시나리오

**시나리오 1: 한국어 프로젝트**
- `conversation_language`: "ko"
- 백업 announcements: 한국어 (번역됨)
- 결과: 한국어 announcements 보존

**시나리오 2: 새로운 일본어 프로젝트**
- `conversation_language`: "ja"
- 백업 announcements: 영어 (미번역)
- 결과: 영어 → 일본어로 번역

**시나리오 3: 영어 프로젝트**
- `conversation_language`: "en"
- 백업 announcements: 영어
- 결과: 영어 announcements 그대로 사용

## Alfred 언어 정책 준수

### 규칙 준수
- ✅ **사용자 대면 콘텐츠**: conversation_language 사용 (한국어)
- ✅ **인프라**: 시스템 인프라는 영어 유지
- ✅ **@TAG 식별자**: 영어 유지
- ✅ **기술 함수명**: 영어 유지

### 번역 전략
- **단일 소스**: 영어 announcements만 유지 (템플릿)
- **런타임 번역**: 사용자 언어로 즉시 번역
- **제로 중복**: 번역된 사본 유지 안 함
- **미래 보장**: 새로운 언어 자동 지원

## 테스트 시나리오

### 테스트 케이스 1: 기존 번역 보존
```bash
# 현재 상태 (한국어 announcements)
/alfred:0-project update
# 예상 결과: 한국어 announcements 그대로 보존
```

### 테스트 케이스 2: 새로운 번역
```bash
# conversation_language를 "ja"로 변경
# .moai/config.json 수정
/alfred:0-project update
# 예상 결과: 영어 → 일본어로 번역
```

### 테스트 케이스 3: 영어 유지
```bash
# conversation_language를 "en"으로 설정
/alfred:0-project update
# 예상 결과: 영어 announcements 그대로 사용
```

## 영향 범위

### 변경된 파일
1. `.claude/commands/alfred/0-project.md`
   - STEP 0-UPDATE.2: companyAnnouncements 비교 로직 추가
   - STEP 0-UPDATE.5: companyAnnouncements 병합 로직 추가
   - 보고서 템플릿 업데이트

2. `src/moai_adk/templates/.claude/settings.json`
   - Hook 경로 변수 표준화
   - 추가 훅 포함
   - UI 설정 추가

### 영향받는 프로세스
- `/alfred:0-project update` 서브커맨드
- 템플릿 최적화 프로세스
- 런타임 announcements 번역

### 하위 호환성
- ✅ 기존 번역된 announcements 자동 보존
- ✅ 새로운 언어 설정 즉시 지원
- ✅ 영어 환경에서 변경 없음

## 결론

이 수정으로 `/alfred:0-project update` 프로세스는:

1. **companyAnnouncements 자동 보존**: 기존 번역된 announcements를 그대로 유지
2. **언어 인식 번역**: 필요한 경우에만 자동 번역 실행
3. **Alfred 언어 정책 준수**: 사용자 대면은 현지어, 인프라는 영어
4. **단일 소스 유지**: 템플릿은 영어로, 번역은 런타임에 처리
5. **제로 중복**: 미리 번역된 사본 관리 불필요

사용자는 더 이상 update 후 announcements가 영어로 덮어쓰여지는 문제를 겪지 않을 것입니다.