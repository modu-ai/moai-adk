# @SPEC:SKILL-REFACTOR-001 인수 기준 (Acceptance Criteria)

> Claude Code Skills 표준화 작업 인수 기준

---

## 📋 AC-001: 파일명 표준화

### Given (전제 조건)
- 50개 스킬에 skill.md (소문자) 파일이 존재
- 1개 스킬에 SKILL.md (대문자) 파일이 존재

### When (작업 실행)
- 파일명 변경 스크립트를 실행
- `.claude/skills/`와 `src/moai_adk/templates/.claude/skills/` 양쪽에 적용

### Then (예상 결과)
- **모든 스킬이 SKILL.md (대문자) 파일을 가져야 함**
  - `.claude/skills/*/SKILL.md`: 51개
  - `src/moai_adk/templates/.claude/skills/*/SKILL.md`: 51개
- **skill.md (소문자) 파일이 0개여야 함**
  - `find .claude/skills/ -name "skill.md" | wc -l` → 0
- **파일 내용은 변경되지 않았어야 함**
  - 파일명만 변경, 내용은 동일

### 검증 명령어
```bash
# skill.md 파일 없음 확인
find .claude/skills/ -name "skill.md" | wc -l  # 0

# SKILL.md 파일 51개 확인
find .claude/skills/ -name "SKILL.md" | wc -l  # 51

# 템플릿 디렉토리도 동일
find src/moai_adk/templates/.claude/skills/ -name "SKILL.md" | wc -l  # 51
```

---

## 📋 AC-002: 중복 템플릿 삭제

### Given (전제 조건)
- moai-cc-agent-template 디렉토리 존재
- moai-cc-command-template 디렉토리 존재
- moai-cc-skill-template 디렉토리 존재
- moai-cc-plugin-template 디렉토리 존재
- moai-cc-settings-template 디렉토리 존재
- moai-claude-code 디렉토리 존재 (통합 템플릿)

### When (작업 실행)
- 중복 템플릿 삭제 스크립트를 실행
- `.claude/skills/`와 `src/moai_adk/templates/.claude/skills/` 양쪽에 적용

### Then (예상 결과)
- **moai-cc-*-template 디렉토리가 0개여야 함**
  - `.claude/skills/moai-cc-*-template`: 삭제됨
  - `src/moai_adk/templates/.claude/skills/moai-cc-*-template`: 삭제됨
- **moai-claude-code 디렉토리는 남아있어야 함**
  - `.claude/skills/moai-claude-code`: 존재
  - `src/moai_adk/templates/.claude/skills/moai-claude-code`: 존재

### 검증 명령어
```bash
# 중복 템플릿 없음 확인
ls .claude/skills/ | grep -c "moai-cc-.*-template"  # 0

# moai-claude-code 존재 확인
test -d .claude/skills/moai-claude-code && echo "EXISTS" || echo "MISSING"  # EXISTS

# 템플릿 디렉토리도 동일
test -d src/moai_adk/templates/.claude/skills/moai-claude-code && echo "EXISTS" || echo "MISSING"  # EXISTS
```

---

## 📋 AC-003: YAML 필드 정리

### Given (전제 조건)
- 50개 스킬에 version, author, license, tags, model 필드 존재
- 총 174개 불필요한 필드 존재 (50개 × 평균 3.5개)

### When (작업 실행)
- YAML 필드 정리 스크립트를 실행 (`clean_yaml_fields.py`)
- version, author, license, tags, model 필드 제거
- name, description, allowed-tools만 유지

### Then (예상 결과)
- **version 필드가 0개여야 함**
  - `rg "^version:" .claude/skills/*/SKILL.md | wc -l` → 0
- **author 필드가 0개여야 함**
  - `rg "^author:" .claude/skills/*/SKILL.md | wc -l` → 0
- **license 필드가 0개여야 함**
  - `rg "^license:" .claude/skills/*/SKILL.md | wc -l` → 0
- **tags 필드가 0개여야 함**
  - `rg "^tags:" .claude/skills/*/SKILL.md | wc -l` → 0
- **model 필드가 0개여야 함** (Skills는 model 필드 불필요)
  - `rg "^model:" .claude/skills/*/SKILL.md | wc -l` → 0
- **name, description 필드는 유지되어야 함**
  - `rg "^name:" .claude/skills/*/SKILL.md | wc -l` → 51
  - `rg "^description:" .claude/skills/*/SKILL.md | wc -l` → 51

### 검증 명령어
```bash
# 불필요한 필드 제거 확인
rg "^version:" .claude/skills/*/SKILL.md | wc -l  # 0
rg "^author:" .claude/skills/*/SKILL.md | wc -l  # 0
rg "^license:" .claude/skills/*/SKILL.md | wc -l  # 0
rg "^tags:" .claude/skills/*/SKILL.md | wc -l  # 0
rg "^model:" .claude/skills/*/SKILL.md | wc -l  # 0

# 필수 필드 유지 확인
rg "^name:" .claude/skills/*/SKILL.md | wc -l  # 51
rg "^description:" .claude/skills/*/SKILL.md | wc -l  # 51
```

---

## 📋 AC-004: allowed-tools 필드 추가

### Given (전제 조건)
- 25개 스킬에 allowed-tools 필드 누락
- 26개 스킬에 allowed-tools 필드 이미 존재

### When (작업 실행)
- allowed-tools 추가 스크립트를 실행 (`add_allowed_tools.py`)
- 스킬 유형별 도구 권한 추가:
  - **Alfred 에이전트**: Read, Write, Edit, Bash, TodoWrite
  - **Lang 스킬**: Read, Bash
  - **Domain 스킬**: Read, Bash

### Then (예상 결과)
- **모든 스킬(51개)이 allowed-tools 필드를 가져야 함**
  - `rg "^allowed-tools:" .claude/skills/*/SKILL.md | wc -l` → 51
- **Alfred 에이전트 스킬은 5개 도구를 가져야 함**
  - debugger-pro: Read, Write, Edit, Bash, TodoWrite
  - code-reviewer: Read, Write, Edit, Bash, TodoWrite
- **Lang 스킬은 2개 도구를 가져야 함**
  - python: Read, Bash
  - typescript: Read, Bash
- **Domain 스킬은 2개 도구를 가져야 함**
  - backend: Read, Bash
  - frontend: Read, Bash
- **기존에 allowed-tools가 있던 스킬은 변경되지 않았어야 함**

### 검증 명령어
```bash
# 모든 스킬에 allowed-tools 필드 존재 확인
rg "^allowed-tools:" .claude/skills/*/SKILL.md | wc -l  # 51

# Alfred 에이전트 예시 확인
rg -A 5 "^allowed-tools:" .claude/skills/moai-alfred-debugger-pro/SKILL.md

# Lang 스킬 예시 확인
rg -A 3 "^allowed-tools:" .claude/skills/moai-lang-python/SKILL.md

# Domain 스킬 예시 확인
rg -A 3 "^allowed-tools:" .claude/skills/moai-domain-backend/SKILL.md
```

---

## ✅ 종합 검증

### 통합 체크리스트

모든 인수 기준을 통합하여 한 번에 검증:

```bash
#!/bin/bash

echo "=== SPEC-SKILL-REFACTOR-001 종합 검증 ==="

# AC-001: 파일명 표준화
skill_md_count=$(find .claude/skills/ -name "skill.md" 2>/dev/null | wc -l | tr -d ' ')
SKILL_md_count=$(find .claude/skills/ -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')

echo "AC-001: 파일명 표준화"
echo "  ✓ skill.md (비표준): $skill_md_count (0이어야 함)"
echo "  ✓ SKILL.md (표준): $SKILL_md_count (51이어야 함)"

# AC-002: 중복 템플릿 삭제
duplicate_count=$(ls .claude/skills/ 2>/dev/null | grep -c "moai-cc-.*-template" || echo 0)

echo "AC-002: 중복 템플릿 삭제"
echo "  ✓ moai-cc-*-template: $duplicate_count (0이어야 함)"

# AC-003: YAML 필드 정리
version_count=$(rg "^version:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
model_count=$(rg "^model:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
name_count=$(rg "^name:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
desc_count=$(rg "^description:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')

echo "AC-003: YAML 필드 정리"
echo "  ✓ version 필드 제거: $version_count (0이어야 함)"
echo "  ✓ model 필드 제거: $model_count (0이어야 함)"
echo "  ✓ name 필드 유지: $name_count (51이어야 함)"
echo "  ✓ description 필드 유지: $desc_count (51이어야 함)"

# AC-004: allowed-tools 추가
allowed_tools_count=$(rg "^allowed-tools:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')

echo "AC-004: allowed-tools 추가"
echo "  ✓ allowed-tools 필드: $allowed_tools_count (51이어야 함)"

# 종합 판정
if [ "$skill_md_count" -eq 0 ] && \
   [ "$SKILL_md_count" -eq 51 ] && \
   [ "$duplicate_count" -eq 0 ] && \
   [ "$version_count" -eq 0 ] && \
   [ "$model_count" -eq 0 ] && \
   [ "$name_count" -eq 51 ] && \
   [ "$desc_count" -eq 51 ] && \
   [ "$allowed_tools_count" -eq 51 ]; then
    echo ""
    echo "✅ 모든 AC 통과!"
    echo ""
    echo "다음 단계:"
    echo "  1. Git 커밋 (4개 커밋 생성)"
    echo "  2. /alfred:3-sync 실행 (문서 동기화)"
    exit 0
else
    echo ""
    echo "❌ 일부 AC 실패. 위 항목을 확인하세요."
    exit 1
fi
```

### 실행 결과 예시

**성공 시**:
```
=== SPEC-SKILL-REFACTOR-001 종합 검증 ===
AC-001: 파일명 표준화
  ✓ skill.md (비표준): 0 (0이어야 함)
  ✓ SKILL.md (표준): 51 (51이어야 함)
AC-002: 중복 템플릿 삭제
  ✓ moai-cc-*-template: 0 (0이어야 함)
AC-003: YAML 필드 정리
  ✓ version 필드 제거: 0 (0이어야 함)
  ✓ model 필드 제거: 0 (0이어야 함)
  ✓ name 필드 유지: 51 (51이어야 함)
  ✓ description 필드 유지: 51 (51이어야 함)
AC-004: allowed-tools 추가
  ✓ allowed-tools 필드: 51 (51이어야 함)

✅ 모든 AC 통과!

다음 단계:
  1. Git 커밋 (4개 커밋 생성)
  2. /alfred:3-sync 실행 (문서 동기화)
```

---

## 📦 산출물 확인

### 변경된 파일 목록

**삭제** (5개 디렉토리):
- .claude/skills/moai-cc-agent-template/
- .claude/skills/moai-cc-command-template/
- .claude/skills/moai-cc-skill-template/
- .claude/skills/moai-cc-plugin-template/
- .claude/skills/moai-cc-settings-template/

**파일명 변경** (100개):
- .claude/skills/*/skill.md → SKILL.md (50개)
- src/moai_adk/templates/.claude/skills/*/skill.md → SKILL.md (50개)

**파일 수정** (100개):
- YAML 필드 정리 (51개 × 2 = 102개)
- allowed-tools 추가 (25개 × 2 = 50개)

### Git 상태 확인

```bash
# 변경 파일 확인
git status --short

# 예상 출력:
# D  .claude/skills/moai-cc-agent-template/
# D  .claude/skills/moai-cc-command-template/
# D  .claude/skills/moai-cc-skill-template/
# D  .claude/skills/moai-cc-plugin-template/
# D  .claude/skills/moai-cc-settings-template/
# R  .claude/skills/moai-alfred-code-reviewer/skill.md -> SKILL.md
# M  .claude/skills/moai-alfred-code-reviewer/SKILL.md
# (반복 50회)
```

---

**최종 업데이트**: 2025-10-19
**작성자**: @Goos
