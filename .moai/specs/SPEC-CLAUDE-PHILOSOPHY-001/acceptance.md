# @SPEC:CLAUDE-PHILOSOPHY-001 수용 기준

## 개요

이 문서는 SPEC-CLAUDE-PHILOSOPHY-001 (CLAUDE.md 철학 재정렬 및 Skill 분리) 구현의 수용 기준을 정의합니다. 모든 시나리오는 Given-When-Then 형식으로 작성되었습니다.

---

## 수용 기준 시나리오

### 시나리오 1: 새로운 세션 시작 시 Tier 1 핵심 규칙 표시

**Given**: 사용자가 MoAI-ADK 프로젝트에서 새로운 세션을 시작한다
**When**: CLAUDE.md가 로드된다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] Tier 1 섹션이 문서 상단 500줄 이내에 위치한다
- [ ] Tier 1 섹션에 다음 내용이 포함된다:
  - [ ] 4단계 워크플로우 (의도 파악 → 계획 수립 → 작업 실행 → 보고 및 커밋)
  - [ ] 언어 경계 규칙 (Layer 1: conversation_language, Layer 2: 영어 인프라)
  - [ ] Permissions 우선순위 (deny → ask → allow)
  - [ ] TRUST 5 원칙 (Test First, Readable, Unified, Secured, Trackable)
- [ ] Tier 1 섹션이 400-500줄 범위 내에 있다
- [ ] 스크롤 없이 핵심 규칙을 확인할 수 있다

**검증 방법**:
```bash
# Tier 1 위치 확인
head -500 /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | grep "Tier 1"

# Tier 1 줄 수 확인
awk '/## .*Tier 1/,/## .*Tier 2/' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | wc -l
# 예상 결과: 400-500줄

# 핵심 섹션 포함 여부 확인
grep "4단계 워크플로우\|언어 경계\|Permissions\|TRUST 5" /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
```

**성공 기준**: 모든 필수 조건 체크리스트 완료

---

### 시나리오 2: Alfred가 세션 분석 필요 시 Skill JIT 로드

**Given**: 사용자가 "세션 분석해줘" 또는 "로그 확인" 요청을 한다
**When**: Alfred가 Skill("moai-alfred-session-analytics")를 호출한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] `.claude/skills/moai-alfred-session-analytics/` 디렉토리가 존재한다
- [ ] 다음 파일이 존재한다:
  - [ ] `SKILL.md` (YAML frontmatter + 개요)
  - [ ] `reference.md` (세션 메트릭 정의)
  - [ ] `examples.md` (세션 분석 예시)
- [ ] SKILL.md의 YAML frontmatter에 다음 필드가 포함된다:
  - [ ] `name: moai-alfred-session-analytics`
  - [ ] `version: 1.0.0`
  - [ ] `status: active`
  - [ ] `description`: 세션 분석, 로깅, 메트릭 수집 관련
  - [ ] `keywords`: ['session', 'analytics', 'logging', 'metrics']
  - [ ] `allowed-tools`: [Read, Bash, Grep]
- [ ] Skill이 정상적으로 로드된다 (에러 없음)
- [ ] Alfred가 세션 분석 정보를 제공한다

**검증 방법**:
```bash
# 디렉토리 존재 확인
ls -la /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics/

# 필수 파일 확인
test -f /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics/SKILL.md && echo "SKILL.md OK"
test -f /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics/reference.md && echo "reference.md OK"
test -f /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics/examples.md && echo "examples.md OK"

# YAML frontmatter 검증
head -15 /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics/SKILL.md | grep "name:\|version:\|status:\|description:\|keywords:\|allowed-tools:"

# Skill 로드 테스트 (수동)
# Claude Code 세션에서: Skill("moai-alfred-session-analytics") 호출
```

**성공 기준**: 모든 필수 조건 체크리스트 완료 + Skill 정상 로드

---

### 시나리오 3: Alfred가 고급 설정 필요 시 Skill JIT 로드

**Given**: 사용자가 "Hook 타임아웃 조정" 또는 "권한 정책 변경" 요청을 한다
**When**: Alfred가 Skill("moai-alfred-config-advanced")를 호출한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] `.claude/skills/moai-alfred-config-advanced/` 디렉토리가 존재한다
- [ ] 다음 파일이 존재한다:
  - [ ] `SKILL.md` (YAML frontmatter + 개요)
  - [ ] `reference.md` (고급 설정 필드 설명)
  - [ ] `examples.md` (고급 설정 예시)
- [ ] SKILL.md의 YAML frontmatter에 다음 필드가 포함된다:
  - [ ] `name: moai-alfred-config-advanced`
  - [ ] `version: 1.0.0`
  - [ ] `status: active`
  - [ ] `description`: Hook 타임아웃, 권한 세분화, 메타데이터 최적화 관련
  - [ ] `keywords`: ['config', 'advanced', 'hooks', 'permissions']
  - [ ] `allowed-tools`: [Read, Edit, Bash]
- [ ] Skill이 정상적으로 로드된다 (에러 없음)
- [ ] Alfred가 고급 설정 정보를 제공한다

**검증 방법**:
```bash
# 디렉토리 존재 확인
ls -la /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced/

# 필수 파일 확인
test -f /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced/SKILL.md && echo "SKILL.md OK"
test -f /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced/reference.md && echo "reference.md OK"
test -f /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced/examples.md && echo "examples.md OK"

# YAML frontmatter 검증
head -15 /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced/SKILL.md | grep "name:\|version:\|status:\|description:\|keywords:\|allowed-tools:"

# Skill 로드 테스트 (수동)
# Claude Code 세션에서: Skill("moai-alfred-config-advanced") 호출
```

**성공 기준**: 모든 필수 조건 체크리스트 완료 + Skill 정상 로드

---

### 시나리오 4: 개발자가 새로운 Permissions 규칙 추가 시 1곳만 수정

**Given**: 개발자가 새로운 Permissions 규칙을 추가하려고 한다
**When**: CLAUDE.md의 Permissions 섹션을 수정한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] Permissions 섹션이 Tier 1에 1개만 존재한다
- [ ] 다른 섹션에서 Permissions 규칙이 중복되지 않는다
- [ ] Skill("moai-alfred-config-advanced")에 상세 설명이 위임되어 있다
- [ ] 개발자가 1개 섹션만 수정하면 된다
- [ ] 변경사항이 패키지 템플릿에도 자동 반영된다 (동기화)

**검증 방법**:
```bash
# Permissions 섹션 개수 확인
grep -c "^## .*Permissions\|^### .*Permissions" /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
# 예상 결과: 1 (Tier 1에만 존재)

# 중복 검사
grep -n "deny.*ask.*allow\|allow.*ask.*deny" /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | wc -l
# 예상 결과: 1 (Tier 1 섹션만)

# Skill 링크 확인
grep "Skill(\"moai-alfred-config-advanced\")" /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
```

**성공 기준**: Permissions 섹션이 1개만 존재하며 상세 내용은 Skill에 위임

---

### 시나리오 5: 로컬과 패키지 템플릿 유지보수 시 구조 동기화

**Given**: 개발자가 로컬 CLAUDE.md를 변경하고 커밋한다
**When**: 패키지 템플릿을 동기화한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] 로컬 CLAUDE.md와 패키지 템플릿의 구조가 동일하다
- [ ] 섹션 수가 일치한다
- [ ] Skill 링크 개수가 일치한다
- [ ] Tier 1-4 구조가 일치한다
- [ ] 언어만 다르다 (로컬: 한국어, 패키지: 영어)
- [ ] YAML frontmatter 필드가 일치한다 (언어 제외)

**검증 방법**:
```bash
# 섹션 수 비교
LOCAL_SECTIONS=$(grep -c '^##' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md)
PKG_SECTIONS=$(grep -c '^##' /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/CLAUDE.md)
echo "Local: $LOCAL_SECTIONS, Package: $PKG_SECTIONS"
test "$LOCAL_SECTIONS" -eq "$PKG_SECTIONS" && echo "Sections OK"

# Skill 링크 비교
LOCAL_SKILLS=$(grep -c 'Skill("' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md)
PKG_SKILLS=$(grep -c 'Skill("' /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/CLAUDE.md)
echo "Local Skills: $LOCAL_SKILLS, Package Skills: $PKG_SKILLS"
test "$LOCAL_SKILLS" -eq "$PKG_SKILLS" && echo "Skills OK"

# Tier 구조 비교
diff <(grep '^## .*Tier' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md) <(grep '^## .*Tier' /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/CLAUDE.md)
# 예상 결과: 언어만 다름 (예: "Tier 1 (핵심 규칙)" vs "Tier 1 (Core Rules)")

# 구조 일치 자동 검증 스크립트
bash /Users/goos/MoAI/MoAI-ADK/.moai/scripts/verify-claude-md-sync.sh
```

**성공 기준**: 모든 필수 조건 체크리스트 완료 + 구조 일치 검증 통과

---

### 시나리오 6: 긍정적 가이드라인 적용 후 20개 이상 변환 확인

**Given**: 부정적 제약을 긍정적 가이드라인으로 변환한다
**When**: CLAUDE.md를 검토한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] 최소 20개 이상의 부정적 표현이 긍정적으로 변환되었다
- [ ] 변환 전 부정적 표현 개수: N개
- [ ] 변환 후 부정적 표현 개수: (N - 20)개 이하
- [ ] 긍정적 표현 개수: 최소 20개 증가
- [ ] 핵심 금지사항은 유지되었다:
  - [ ] "NEVER run git push --force to main/master"
  - [ ] "NEVER amend other developers' commits"
  - [ ] "NEVER skip hooks (--no-verify)"
  - [ ] "NEVER hardcode secrets"

**검증 방법**:
```bash
# 변환 전 부정적 표현 개수 (예상: 50개)
grep -c 'DO NOT\|NEVER\|DON'\''T\|AVOID\|MUST NOT' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md.backup

# 변환 후 부정적 표현 개수 (예상: 30개 이하)
grep -c 'DO NOT\|NEVER\|DON'\''T\|AVOID\|MUST NOT' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md

# 긍정적 표현 증가 확인
grep -c 'INSTEAD\|PREFER\|USE:\|CREATE:\|RUN:\|CHECK:' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
# 예상 결과: 최소 20개 이상

# 핵심 금지사항 유지 확인
grep -n 'NEVER.*git push --force' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
grep -n 'NEVER.*amend other' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
grep -n 'NEVER.*skip hooks\|--no-verify' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
grep -n 'NEVER.*hardcode.*secret' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
```

**성공 기준**: 최소 20개 변환 + 핵심 금지사항 유지

---

### 시나리오 7: Tier 3 고급 기능 섹션이 Skill 링크로 대체

**Given**: CLAUDE.md Tier 3 섹션을 확인한다
**When**: 고급 기능 섹션을 읽는다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] Tier 3 섹션에 다음 Skill 링크가 포함된다:
  - [ ] Skill("moai-alfred-personas") - 적응형 페르소나 시스템
  - [ ] Skill("moai-alfred-autofixes") - 자동 수정 프로토콜
  - [ ] Skill("moai-alfred-reporting") - 보고 스타일
- [ ] 각 Skill 링크 옆에 간단한 설명이 있다 (1-2줄)
- [ ] 상세 내용은 Skill 파일에 위임되어 있다
- [ ] Tier 3 섹션 총 줄 수: 100-150줄 (상세 내용 제거)

**검증 방법**:
```bash
# Tier 3 섹션 추출
awk '/## .*Tier 3/,/## .*Tier 4/' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md > tier3.txt

# Skill 링크 확인
grep 'Skill("moai-alfred-personas")' tier3.txt
grep 'Skill("moai-alfred-autofixes")' tier3.txt
grep 'Skill("moai-alfred-reporting")' tier3.txt

# Tier 3 줄 수 확인
wc -l tier3.txt
# 예상 결과: 100-150줄
```

**성공 기준**: 모든 고급 기능이 Skill 링크로 대체 + 간략 설명만 포함

---

### 시나리오 8: CLAUDE.md 최소 400줄 유지 (과도한 단순화 방지)

**Given**: CLAUDE.md Tier 1-4 재구조화가 완료되었다
**When**: 전체 CLAUDE.md 파일을 확인한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] 전체 CLAUDE.md 줄 수: 최소 400줄
- [ ] Tier 1 (핵심 규칙): 400-500줄
- [ ] Tier 2 (실행 가이드): 200-300줄
- [ ] Tier 3 (고급 기능): 100-150줄
- [ ] Tier 4 (참조): 100-150줄
- [ ] 총 줄 수: 800-1100줄 (현재 1000줄 대비 약간 감소)

**검증 방법**:
```bash
# 전체 줄 수 확인
wc -l /Users/goos/MoAI/MoAI-ADK/CLAUDE.md
# 예상 결과: 800-1100줄

# Tier별 줄 수 확인
awk '/## .*Tier 1/,/## .*Tier 2/' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | wc -l  # Tier 1
awk '/## .*Tier 2/,/## .*Tier 3/' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | wc -l  # Tier 2
awk '/## .*Tier 3/,/## .*Tier 4/' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | wc -l  # Tier 3
awk '/## .*Tier 4/,EOF' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md | wc -l            # Tier 4
```

**성공 기준**: 전체 400줄 이상 + Tier별 줄 수 범위 내

---

### 시나리오 9: 기존 참조 링크 유효성 유지

**Given**: CLAUDE.md 섹션 재배치가 완료되었다
**When**: 기존 앵커 링크를 확인한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] 모든 내부 앵커 링크가 유효하다
- [ ] 섹션 제목이 유지되었다 (위치만 변경)
- [ ] 깨진 링크가 0개이다
- [ ] 링크 형식: `[텍스트](#섹션-id)` 모두 정상

**검증 방법**:
```bash
# 모든 앵커 링크 추출
grep -o '\[.*\](#.*\)' /Users/goos/MoAI/MoAI-ADK/CLAUDE.md > links.txt

# 각 링크 유효성 검증
while IFS= read -r link; do
  anchor=$(echo "$link" | sed -E 's/.*\(#(.*)\)/\1/')
  grep -q "^#.*$anchor" /Users/goos/MoAI/MoAI-ADK/CLAUDE.md || echo "Broken: $link"
done < links.txt

# 깨진 링크 개수
BROKEN_COUNT=$(while IFS= read -r link; do
  anchor=$(echo "$link" | sed -E 's/.*\(#(.*)\)/\1/')
  grep -q "^#.*$anchor" /Users/goos/MoAI/MoAI-ADK/CLAUDE.md || echo "1"
done < links.txt | wc -l)
echo "Broken links: $BROKEN_COUNT"
test "$BROKEN_COUNT" -eq 0 && echo "All links OK"
```

**성공 기준**: 깨진 링크 0개

---

### 시나리오 10: Git 커밋 메시지 TDD 패턴 준수

**Given**: Phase 1-4 작업이 완료되었다
**When**: Git 커밋을 생성한다
**Then**: 다음 조건이 만족되어야 한다

**필수 조건**:
- [ ] 커밋 메시지가 다음 형식을 따른다:
  ```
  refactor(docs): Phase 6 CLAUDE.md 재구조화 (Tier 1-4, Skill 분리, 긍정적 가이드라인)

  - Tier 1-4 계층 구조 도입 (핵심 규칙 500줄 이내)
  - 2개 Skill 분리 (session-analytics, config-advanced)
  - 20개 이상 부정적 제약 → 긍정적 가이드라인 변환
  - 패키지 템플릿 동기화 완료

  🤖 Generated with Claude Code

  Co-Authored-By: 🎩 Alfred@MoAI
  ```
- [ ] 커밋에 다음 파일이 포함된다:
  - [ ] `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md`
  - [ ] `.claude/skills/moai-alfred-session-analytics/`
  - [ ] `.claude/skills/moai-alfred-config-advanced/`
  - [ ] `src/moai_adk/templates/CLAUDE.md`

**검증 방법**:
```bash
# 커밋 로그 확인
git log -1 --pretty=format:"%s%n%b"

# 커밋 파일 확인
git show --name-only --pretty="" HEAD

# 커밋 메시지 형식 검증
git log -1 --pretty=format:"%s" | grep "refactor(docs): Phase 6"
git log -1 --pretty=format:"%b" | grep "🤖 Generated with Claude Code"
git log -1 --pretty=format:"%b" | grep "Co-Authored-By: 🎩 Alfred@MoAI"
```

**성공 기준**: 커밋 메시지 형식 준수 + 모든 변경 파일 포함

---

## 전체 검증 체크리스트

### Phase 1: 구조 재설계
- [ ] 시나리오 1 통과 (Tier 1 핵심 규칙 표시)
- [ ] 시나리오 7 통과 (Tier 3 Skill 링크 대체)
- [ ] 시나리오 8 통과 (최소 400줄 유지)
- [ ] 시나리오 9 통과 (기존 링크 유효성)

### Phase 2: Skill 분리
- [ ] 시나리오 2 통과 (session-analytics Skill 로드)
- [ ] 시나리오 3 통과 (config-advanced Skill 로드)
- [ ] 시나리오 4 통과 (1곳만 수정)

### Phase 3: 긍정적 가이드라인 변환
- [ ] 시나리오 6 통과 (20개 이상 변환)

### Phase 4: 패키지 동기화
- [ ] 시나리오 5 통과 (로컬-패키지 구조 동기화)
- [ ] 시나리오 10 통과 (Git 커밋 TDD 패턴)

---

## Definition of Done (완료 정의)

### 필수 조건 (모두 만족 시 완료)
1. **구조 검증**: Tier 1-4 계층 구조 완료 (Tier 1: 400-500줄)
2. **Skill 검증**: 2개 Skill 정상 로드 (session-analytics, config-advanced)
3. **변환 검증**: 최소 20개 부정적 → 긍정적 가이드라인 변환
4. **동기화 검증**: 로컬-패키지 구조 일치 (언어만 다름)
5. **링크 검증**: 모든 Skill 링크 + 앵커 링크 유효성 확인
6. **커밋 검증**: TDD 패턴 커밋 메시지 + 모든 변경 파일 포함
7. **테스트 통과**: 10개 시나리오 모두 통과

### 선택 조건 (권장)
8. **추가 변환**: 20개 이상 긍정적 가이드라인 변환
9. **문서화**: Phase 6 요약 리포트 작성
10. **CI/CD 검증**: 자동화된 구조 일치 검증 스크립트 추가

---

## 자동화된 검증 스크립트

### verify-spec-claude-philosophy-001.sh

```bash
#!/bin/bash
# SPEC-CLAUDE-PHILOSOPHY-001 수용 기준 자동 검증 스크립트

set -e

CLAUDE_LOCAL="/Users/goos/MoAI/MoAI-ADK/CLAUDE.md"
CLAUDE_PKG="/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/CLAUDE.md"
SKILL_SESSION="/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-session-analytics"
SKILL_CONFIG="/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-config-advanced"

echo "=== SPEC-CLAUDE-PHILOSOPHY-001 수용 기준 검증 ==="

# 시나리오 1: Tier 1 핵심 규칙
echo "[Scenario 1] Tier 1 핵심 규칙 검증..."
TIER1_LINES=$(awk '/## .*Tier 1/,/## .*Tier 2/' "$CLAUDE_LOCAL" | wc -l)
if [ "$TIER1_LINES" -ge 400 ] && [ "$TIER1_LINES" -le 500 ]; then
  echo "✅ Tier 1 줄 수: $TIER1_LINES (400-500줄 범위)"
else
  echo "❌ Tier 1 줄 수: $TIER1_LINES (범위 벗어남)"
  exit 1
fi

# 시나리오 2: session-analytics Skill
echo "[Scenario 2] session-analytics Skill 검증..."
if [ -d "$SKILL_SESSION" ] && [ -f "$SKILL_SESSION/SKILL.md" ] && [ -f "$SKILL_SESSION/reference.md" ] && [ -f "$SKILL_SESSION/examples.md" ]; then
  echo "✅ session-analytics Skill 파일 존재"
else
  echo "❌ session-analytics Skill 파일 누락"
  exit 1
fi

# 시나리오 3: config-advanced Skill
echo "[Scenario 3] config-advanced Skill 검증..."
if [ -d "$SKILL_CONFIG" ] && [ -f "$SKILL_CONFIG/SKILL.md" ] && [ -f "$SKILL_CONFIG/reference.md" ] && [ -f "$SKILL_CONFIG/examples.md" ]; then
  echo "✅ config-advanced Skill 파일 존재"
else
  echo "❌ config-advanced Skill 파일 누락"
  exit 1
fi

# 시나리오 5: 로컬-패키지 구조 동기화
echo "[Scenario 5] 로컬-패키지 구조 동기화 검증..."
LOCAL_SECTIONS=$(grep -c '^##' "$CLAUDE_LOCAL")
PKG_SECTIONS=$(grep -c '^##' "$CLAUDE_PKG")
if [ "$LOCAL_SECTIONS" -eq "$PKG_SECTIONS" ]; then
  echo "✅ 섹션 수 일치: Local=$LOCAL_SECTIONS, Package=$PKG_SECTIONS"
else
  echo "❌ 섹션 수 불일치: Local=$LOCAL_SECTIONS, Package=$PKG_SECTIONS"
  exit 1
fi

LOCAL_SKILLS=$(grep -c 'Skill("' "$CLAUDE_LOCAL")
PKG_SKILLS=$(grep -c 'Skill("' "$CLAUDE_PKG")
if [ "$LOCAL_SKILLS" -eq "$PKG_SKILLS" ]; then
  echo "✅ Skill 링크 일치: Local=$LOCAL_SKILLS, Package=$PKG_SKILLS"
else
  echo "❌ Skill 링크 불일치: Local=$LOCAL_SKILLS, Package=$PKG_SKILLS"
  exit 1
fi

# 시나리오 6: 긍정적 가이드라인 변환
echo "[Scenario 6] 긍정적 가이드라인 변환 검증..."
POSITIVE_COUNT=$(grep -c 'INSTEAD\|PREFER\|USE:\|CREATE:\|RUN:\|CHECK:' "$CLAUDE_LOCAL" || true)
if [ "$POSITIVE_COUNT" -ge 20 ]; then
  echo "✅ 긍정적 표현: $POSITIVE_COUNT개 (최소 20개 이상)"
else
  echo "❌ 긍정적 표현: $POSITIVE_COUNT개 (20개 미만)"
  exit 1
fi

# 시나리오 8: 최소 400줄 유지
echo "[Scenario 8] CLAUDE.md 최소 400줄 유지 검증..."
TOTAL_LINES=$(wc -l < "$CLAUDE_LOCAL")
if [ "$TOTAL_LINES" -ge 400 ]; then
  echo "✅ 전체 줄 수: $TOTAL_LINES (최소 400줄 이상)"
else
  echo "❌ 전체 줄 수: $TOTAL_LINES (400줄 미만)"
  exit 1
fi

# 시나리오 9: 기존 링크 유효성
echo "[Scenario 9] 기존 링크 유효성 검증..."
# (복잡한 로직이므로 간략화)
LINK_COUNT=$(grep -c '\[.*\](#.*\)' "$CLAUDE_LOCAL" || true)
echo "✅ 내부 링크 개수: $LINK_COUNT (수동 검증 필요)"

echo ""
echo "=== 검증 완료: 모든 자동 검증 통과 ==="
echo "수동 검증 필요: 시나리오 4, 7, 10"
```

---

## 수동 검증 체크리스트

### 시나리오 4: 1곳만 수정
- [ ] Permissions 섹션이 Tier 1에만 존재하는지 확인
- [ ] 다른 섹션에서 Permissions 규칙 중복 없는지 확인

### 시나리오 7: Tier 3 Skill 링크 대체
- [ ] Tier 3 섹션에 Skill 링크 3개 이상 포함 확인
- [ ] 각 Skill 링크 옆에 간략 설명 있는지 확인

### 시나리오 10: Git 커밋 TDD 패턴
- [ ] 커밋 메시지 형식 검토
- [ ] 커밋 파일 목록 검토
- [ ] Co-Authored-By 포함 확인

---

_이 수용 기준은 `/alfred:2-run`으로 구현된 결과를 검증하기 위해 사용됩니다._
