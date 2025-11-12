# @SPEC:SKILLS-REDESIGN-001 검수 기준 (Acceptance Criteria)

> **목적**: Skills 4-Tier 아키텍처 재설계 완료 후 검증 기준 정의
> **형식**: Given-When-Then (Gherkin 스타일)
> **검증자**: Alfred SuperAgent

---

## ✅ AC-1: 스킬 개수 및 Tier 구조

### AC-1.1: 총 스킬 개수는 44개여야 한다

**Given**: MoAI-ADK 프로젝트 진입
**When**: `.claude/skills/` 디렉토리 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
find .claude/skills -name "SKILL.md" | wc -l
# 결과: 44 (46 - 2 = 44)
```

**검증 기준**:
- ✅ 정확히 44개 (41-47개는 실패)
- ✅ SKILL.md는 모두 유효한 파일

**증거**:
```bash
# 현재: 46개
# Phase 2 후: 44개 (template-generator, feature-selector 삭제됨)
```

---

### AC-1.2: Tier 1 (Foundation)은 정확히 6개여야 한다

**Given**: Phase 1 완료
**When**: Foundation 스킬 개수 조회
**Then**: 다음이 성립해야 한다

```bash
ls -d .claude/skills/moai-foundation-* | wc -l
# 결과: 6
```

**Foundation 6개 목록**:
1. ✅ `moai-foundation-trust` (기존: trust-validation)
2. ✅ `moai-foundation-tags` (기존: tag-scanning)
3. ✅ `moai-foundation-specs` (기존: spec-metadata-validation)
4. ✅ `moai-foundation-ears` (기존: ears-authoring)
5. ✅ `moai-foundation-git` (기존: git-workflow)
6. ✅ `moai-foundation-langs` (기존: language-detection)

**증거 파일**:
```bash
ls -la .claude/skills/moai-foundation-*/SKILL.md
```

---

### AC-1.3: Tier 2 (Essentials)은 정확히 4개여야 한다

**Given**: Phase 2 완료
**When**: Essentials 스킬 개수 조회
**Then**: 다음이 성립해야 한다

```bash
ls -d .claude/skills/moai-essentials-* | wc -l
# 결과: 4
```

**Essentials 4개 목록**:
1. ✅ `moai-essentials-review` (기존: code-reviewer)
2. ✅ `moai-essentials-debug` (기존: debugger-pro)
3. ✅ `moai-essentials-refactor` (기존: refactoring-coach)
4. ✅ `moai-essentials-perf` (기존: performance-optimizer)

---

### AC-1.4: Tier 3 (Language)는 정확히 24개여야 한다

**Given**: Phase 3 완료
**When**: Language 스킬 개수 조회
**Then**: 다음이 성립해야 한다

```bash
ls -d .claude/skills/moai-lang-* | wc -l
# 결과: 24
```

**검증**: 모든 24개 언어 스킬 존재 확인
```bash
moai-lang-python moai-lang-typescript moai-lang-javascript moai-lang-java
moai-lang-go moai-lang-rust moai-lang-ruby moai-lang-dart moai-lang-swift
moai-lang-kotlin moai-lang-scala moai-lang-clojure moai-lang-elixir
moai-lang-haskell moai-lang-c moai-lang-cpp moai-lang-csharp
moai-lang-php moai-lang-lua moai-lang-shell moai-lang-sql moai-lang-julia
moai-lang-r
```

---

### AC-1.5: Tier 4 (Domain)은 정확히 9개여야 한다

**Given**: Phase 3 완료
**When**: Domain 스킬 개수 조회
**Then**: 다음이 성립해야 한다

```bash
ls -d .claude/skills/moai-domain-* | wc -l
# 결과: 9
```

**Domain 9개 목록**:
1. ✅ `moai-domain-backend`
2. ✅ `moai-domain-frontend`
3. ✅ `moai-domain-database`
4. ✅ `moai-domain-devops`
5. ✅ `moai-domain-web-api`
6. ✅ `moai-domain-security`
7. ✅ `moai-domain-cli-tool`
8. ✅ `moai-domain-data-science`
9. ✅ `moai-domain-ml`

---

## ✅ AC-2: SKILL.md 표준화

### AC-2.1: 모든 스킬에 allowed-tools 필드가 있어야 한다

**Given**: Phase 1-3 완료
**When**: 모든 SKILL.md 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-*/SKILL.md; do
  grep -q "^allowed-tools:" "$skill" || echo "MISSING: $skill"
done
# 결과: (아무것도 출력되지 않음 = 모두 있음)
```

**검증 기준**:
- ✅ 44개 모두 `allowed-tools:` 필드 존재
- ✅ 필드 값은 배열 형식 (`- Read`, `- Bash` 등)

**예시**:
```yaml
allowed-tools:
  - Read
  - Bash
  - Write
```

---

### AC-2.2: 모든 스킬에 "Works well with" 섹션이 있어야 한다

**Given**: Phase 1-3 완료
**When**: 모든 SKILL.md 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-*/SKILL.md; do
  grep -q "## Works well with" "$skill" || echo "MISSING: $skill"
done
# 결과: (아무것도 출력되지 않음 = 모두 있음)
```

**검증 기준**:
- ✅ 44개 모두 `## Works well with` 섹션 존재
- ✅ 최소 2개 이상의 관련 스킬 나열

**예시**:
```markdown
## Works well with
- moai-foundation-trust (TRUST 검증)
- moai-foundation-tags (TAG 체인)
```

---

### AC-2.3: version, author, license, tags 필드는 제거되어야 한다

**Given**: Phase 1-3 완료
**When**: SKILL.md YAML frontmatter 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-*/SKILL.md; do
  grep -E "^(version|author|license|tags):" "$skill" && echo "FOUND: $skill"
done
# 결과: (아무것도 출력되지 않음 = 모두 제거됨)
```

**검증 기준**:
- ✅ 44개 모두 제거됨
- ✅ 오직 다음만 남음:
  - name
  - description
  - allowed-tools

---

### AC-2.4: description 필드는 200자 이하여야 한다

**Given**: Phase 1-3 완료
**When**: description 필드 길이 측정
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-*/SKILL.md; do
  desc=$(grep "^description:" "$skill" | cut -d' ' -f2-)
  len=${#desc}
  if [[ $len -gt 200 ]]; then
    echo "TOO LONG ($len chars): $skill"
  fi
done
# 결과: (아무것도 출력되지 않음 = 모두 200자 이하)
```

**검증 기준**:
- ✅ 44개 모두 ≤200 chars

---

## ✅ AC-3: 파일 크기 제약

### AC-3.1: 각 Language 스킬은 100줄 이하여야 한다

**Given**: Phase 3 완료
**When**: Language 스킬 SKILL.md 크기 측정
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-lang-*/SKILL.md; do
  lines=$(wc -l < "$skill")
  if [[ $lines -gt 100 ]]; then
    echo "TOO LARGE ($lines lines): $skill"
  fi
done
# 결과: (아무것도 출력되지 않음 = 모두 100줄 이하)
```

**검증 기준**:
- ✅ 24개 모두 ≤100줄

**이유**: Progressive Disclosure - 간결한 스킬이 로드 성능 향상

---

### AC-3.2: 각 Domain 스킬은 100줄 이하여야 한다

**Given**: Phase 3 완료
**When**: Domain 스킬 SKILL.md 크기 측정
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-domain-*/SKILL.md; do
  lines=$(wc -l < "$skill")
  if [[ $lines -gt 100 ]]; then
    echo "TOO LARGE ($lines lines): $skill"
  fi
done
# 결과: (아무것도 출력되지 않음 = 모두 100줄 이하)
```

**검증 기준**:
- ✅ 9개 모두 ≤100줄

---

### AC-3.3: Foundation 스킬 총합은 500줄 이하여야 한다

**Given**: Phase 1 완료
**When**: Foundation 6개 스킬 SKILL.md 총 줄 수 측정
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
total=0
for skill in .claude/skills/moai-foundation-*/SKILL.md; do
  lines=$(wc -l < "$skill")
  total=$((total + lines))
done
echo "Total Foundation lines: $total"
if [[ $total -le 500 ]]; then
  echo "✅ PASS"
else
  echo "❌ FAIL: $total > 500"
fi
```

**검증 기준**:
- ✅ Foundation 6개 합계 ≤500줄

---

## ✅ AC-4: Progressive Disclosure 메커니즘

### AC-4.1: 모든 Tier 3 스킬에 "auto-load: true"가 있어야 한다

**Given**: Phase 3 완료
**When**: Language 스킬 YAML 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-lang-*/SKILL.md; do
  grep -q "auto-load.*true" "$skill" || echo "MISSING: $skill"
done
# 결과: (아무것도 출력되지 않음 = 모두 있음)
```

**검증 기준**:
- ✅ 24개 모두 `auto-load: true` 설정

---

### AC-4.2: 모든 Tier 4 스킬에 "auto-load: false"가 있어야 한다

**Given**: Phase 3 완료
**When**: Domain 스킬 YAML 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
for skill in .claude/skills/moai-domain-*/SKILL.md; do
  grep -q "auto-load.*false" "$skill" || echo "MISSING: $skill"
done
# 결과: (아무것도 출력되지 않음 = 모두 있음)
```

**검증 기준**:
- ✅ 9개 모두 `auto-load: false` 설정

---

## ✅ AC-5: Templates 동기화

### AC-5.1: templates 디렉토리도 동기화되어야 한다

**Given**: Phase 1-4 완료
**When**: templates/.claude/skills/ 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
diff <(find .claude/skills -name "SKILL.md" | sort) \
     <(find templates/.claude/skills -name "SKILL.md" | sort)
# 결과: (차이 없음 = 동일)
```

**검증 기준**:
- ✅ .claude/skills/ = templates/.claude/skills/
- ✅ 파일 개수 동일
- ✅ SKILL.md 내용 동일

---

### AC-5.2: 삭제된 스킬도 templates에서 삭제되어야 한다

**Given**: Phase 2 완료
**When**: templates 디렉토리 확인
**Then**: 다음이 성립해야 한다

```bash
# 검증 명령어
find templates/.claude/skills -name "moai-alfred-template-generator" -o -name "moai-alfred-feature-selector"
# 결과: (아무것도 출력되지 않음 = 모두 삭제됨)
```

**검증 기준**:
- ✅ moai-alfred-template-generator 없음
- ✅ moai-alfred-feature-selector 없음

---

## ✅ AC-6: 워크플로우 통합

### AC-6.1: /alfred:1-plan 실행 시 Tier 1 스킬만 로드되어야 한다

**Given**: Phase 4 완료
**When**: `/alfred:1-plan "테스트 기능"` 실행
**Then**: 다음이 성립해야 한다

```
✅ 로드 스킬:
- moai-foundation-ears (SPEC 작성 가이드)
- moai-foundation-specs (메타데이터 검증)
- moai-foundation-git (브랜치 생성)

❌ 로드 안 됨:
- Tier 2 스킬 (essentials)
- Tier 3 스킬 (language)
- Tier 4 스킬 (domain)
```

**검증 방법**:
1. 콘솔 로그 확인
2. 생성된 feature 브랜치 확인
3. SPEC 파일 확인

---

### AC-6.2: /alfred:2-run 실행 시 Language 스킬 자동 로드

**Given**: Python 프로젝트, Phase 4 완료
**When**: `/alfred:2-run AUTH-001` 실행
**Then**: 다음이 성립해야 한다

```
✅ 로드 스킬:
- moai-foundation-langs (언어 감지)
- moai-lang-python (자동 로드)
- moai-essentials-review (선택적)

❌ 로드 안 됨:
- moai-lang-typescript, moai-lang-java 등 (다른 23개 언어)
```

**검증 방법**:
1. moai-foundation-langs 실행 확인
2. Python 감지 로그 확인
3. moai-lang-python 스킬만 로드됨 확인

---

### AC-6.3: /alfred:3-sync 실행 시 Tier 1 스킬 조합 작동

**Given**: Phase 4 완료
**When**: `/alfred:3-sync` 실행
**Then**: 다음이 성립해야 한다

```
순차 실행:
1. moai-foundation-tags (TAG 인벤토리)
2. moai-foundation-specs (SPEC 검증)
3. moai-foundation-trust (TRUST 검증)
4. moai-foundation-git (PR 상태 변경)

결과 파일:
- tag-inventory.md
- spec-validation-report.md
- trust-report.md
```

**검증 방법**:
1. 4개 파일 생성 확인
2. PR 상태 변경 확인
3. 콘솔 로그 순서 확인

---

## ✅ AC-7: 삭제 기능 이관

### AC-7.1: template-generator 기능이 moai-claude-code로 이관되어야 한다

**Given**: Phase 2 완료
**When**: moai-claude-code 스킬 확인
**Then**: 다음이 성립해야 한다

```
✅ templates/ 디렉토리 존재:
- templates/agent-full.md
- templates/command-full.md
- templates/skill-full.md
- templates/plugin-full.json
- templates/settings-full.json
```

**검증 방법**:
1. moai-claude-code/templates/ 존재 확인
2. 5개 템플릿 파일 확인

---

### AC-7.2: feature-selector 기능이 Commands에 이관되어야 한다

**Given**: Phase 2 완료
**When**: /alfred:1-plan 커맨드 확인
**Then**: 다음이 성립해야 한다

```
/alfred:1-plan 내부 로직:
- 기능 타입 선택 (new, update, refactor 등)
- 언어 선택 (Python, TypeScript 등)
- 도메인 선택 (Backend, Mobile 등)
```

**검증 방법**:
1. /alfred:1-plan 실행
2. 기능/언어/도메인 선택 가능 확인

---

## 📊 전체 검수 요약

| AC | 항목 | 검증 | 상태 |
|----|------|------|------|
| AC-1.1 | 총 스킬 44개 | `find .claude/skills -name SKILL.md \| wc -l` | 대기 |
| AC-1.2 | Foundation 6개 | `ls -d .claude/skills/moai-foundation-*` | 대기 |
| AC-1.3 | Essentials 4개 | `ls -d .claude/skills/moai-essentials-*` | 대기 |
| AC-1.4 | Language 24개 | `ls -d .claude/skills/moai-lang-*` | 대기 |
| AC-1.5 | Domain 9개 | `ls -d .claude/skills/moai-domain-*` | 대기 |
| AC-2.1 | allowed-tools 44개 | grep 확인 | 대기 |
| AC-2.2 | Works well with 44개 | grep 확인 | 대기 |
| AC-2.3 | 제거 필드 | grep 반대 | 대기 |
| AC-2.4 | description <200 | 길이 검증 | 대기 |
| AC-3.1 | Language <100줄 | wc -l | 대기 |
| AC-3.2 | Domain <100줄 | wc -l | 대기 |
| AC-3.3 | Foundation <500줄 | 합계 | 대기 |
| AC-4.1 | Language auto-load | grep true | 대기 |
| AC-4.2 | Domain auto-load | grep false | 대기 |
| AC-5.1 | Templates 동기화 | diff | 대기 |
| AC-5.2 | 2개 삭제 | find | 대기 |
| AC-6.1 | /alfred:1-plan | 수동 테스트 | 대기 |
| AC-6.2 | /alfred:2-run | 수동 테스트 | 대기 |
| AC-6.3 | /alfred:3-sync | 수동 테스트 | 대기 |
| AC-7.1 | 기능 이관 | 파일 확인 | 대기 |
| AC-7.2 | 기능 이관 | 기능 테스트 | 대기 |

**합계**: 21개 AC 모두 ✅ 통과 시 프로젝트 완료

---

**작성**: SPEC-SKILLS-REDESIGN-001 검수 기준
**최종 업데이트**: 2025-10-19
