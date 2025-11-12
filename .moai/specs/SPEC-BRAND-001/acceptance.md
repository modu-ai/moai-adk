
## 개요

이 문서는 "AI-Agent Alfred 브랜딩 일관성 통일" SPEC의 상세한 수락 기준을 정의합니다. 모든 수락 기준은 Given-When-Then 형식으로 작성되며, 검증 가능해야 합니다.

---

## Scenario 1: Git 커밋 메시지 브랜딩

### Given-When-Then

```gherkin
Given AI-Agent Alfred가 코드 변경을 커밋할 때
When 커밋 메시지를 작성하면
Then 커밋 메시지 푸터에 "🤖 Generated with AI-Agent Alfred" 문구가 포함되어야 한다
And 커밋 서명에 "Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>" 포함되어야 한다
And Git log에서 `git log -1 --pretty=format:"%B"` 실행 시 확인 가능해야 한다
```

### 검증 방법

```bash
# 최근 커밋 메시지 확인
git log -1 --pretty=format:"%B"

# 브랜딩 문구 확인
git log -1 --pretty=format:"%B" | grep "Generated with AI-Agent Alfred"

# Co-Authored-By 확인
git log -1 --pretty=format:"%B" | grep "Co-Authored-By: AI-Agent Alfred"
```

### 예상 출력

```
feat(brand): Update branding to AI-Agent Alfred

Update CLAUDE.md and commit template.

🤖 Generated with AI-Agent Alfred

Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>
```

### 실패 조건
- 브랜딩 문구가 누락된 경우
- Co-Authored-By 서명이 누락된 경우
- 이메일 주소가 `noreply@anthropic.com`이 아닌 경우

---

## Scenario 2: CLAUDE.md 문서 브랜딩 업데이트

### Given-When-Then

```gherkin
Given CLAUDE.md 문서를 업데이트할 때
When Line 14와 Line 52를 수정하면
Then Line 14: "AI-Agent Alfred 워크플로우의 중앙 오케스트레이터"로 변경되어야 한다
And Line 52: "MoAI-ADK 설정"으로 변경되어야 한다
And `rg "Claude Code" -n CLAUDE.md` 실행 시 결과가 0건이어야 한다 (외부 노출 부분)
```

### 검증 방법

```bash
# Line 14 확인
sed -n '14p' CLAUDE.md

# Line 52 확인
sed -n '52p' CLAUDE.md

# "Claude Code" 참조 확인 (외부 노출 부분에서 제거되었는지)
rg "Claude Code" -n CLAUDE.md

# 새 브랜딩 확인
rg "AI-Agent Alfred" -n CLAUDE.md
```

### 예상 출력

**Line 14**:
```
- **역할**: AI-Agent Alfred 워크플로우의 중앙 오케스트레이터
```

**Line 52**:
```
| **cc-manager** 🛠️ | 데브옵스 엔지니어 | MoAI-ADK 설정 | `@agent-cc-manager` | 설정 필요 시 |
```

**rg "Claude Code"**:
```
(결과 없음 또는 내부 컨텍스트에서만 존재)
```

### 실패 조건
- Line 14 또는 Line 52가 변경되지 않은 경우
- "Claude Code" 참조가 외부 노출 부분에 남아있는 경우

---

## Scenario 3: GitHub PR/Issue 템플릿 브랜딩

### Given-When-Then

```gherkin
Given GitHub PR 또는 Issue를 생성할 때
When 템플릿 본문을 작성하면
Then 본문 마지막에 "🤖 Generated with AI-Agent Alfred" 문구가 포함되어야 한다
And GitHub 웹 UI에서 확인 가능해야 한다
And 기존 PR/Issue는 변경하지 않아야 한다 (새 PR/Issue부터 적용)
```

### 검증 방법

```bash
# PR 템플릿 확인
cat .github/PULL_REQUEST_TEMPLATE.md

# Issue 템플릿 확인
cat .github/ISSUE_TEMPLATE/*.md

# 브랜딩 문구 확인
rg "Generated with AI-Agent Alfred" -n .github/
```

### 예상 출력

**PR 템플릿 예시**:
```markdown
## Summary
- Brief description

## Test Plan
- [ ] Test 1

🤖 Generated with AI-Agent Alfred
```

### 실패 조건
- 템플릿에 브랜딩 문구가 누락된 경우
- 기존 PR/Issue가 변경된 경우 (히스토리 재작성)

---

## Scenario 4: 브랜딩 일관성 전체 검증

### Given-When-Then

```gherkin
Given 프로젝트 전체 파일을 검증할 때
When `rg "Generated with AI-Agent Alfred" -n` 실행하면
Then CLAUDE.md, 커밋 메시지, GitHub 템플릿에서 확인되어야 한다
And 결과가 > 0건이어야 한다
When `rg "Co-Authored-By: AI-Agent Alfred" -n` 실행하면
Then Git 로그 또는 커밋 메시지에서 확인되어야 한다
And 결과가 > 0건이어야 한다
```

### 검증 방법

```bash
# 새 브랜딩 문구 전체 검색
rg "Generated with AI-Agent Alfred" -n

# Co-Authored-By 서명 전체 검색
rg "Co-Authored-By: AI-Agent Alfred" -n

# 결과 개수 확인
rg "Generated with AI-Agent Alfred" -n | wc -l
rg "Co-Authored-By: AI-Agent Alfred" -n | wc -l
```

### 예상 출력

```
# "Generated with" 검색 결과
.git/COMMIT_EDITMSG:5:🤖 Generated with AI-Agent Alfred
.github/PULL_REQUEST_TEMPLATE.md:10:🤖 Generated with AI-Agent Alfred
moai-adk-ts/src/core/git/constants/config-constants.ts:265:🤖 Generated with AI-Agent Alfred

# "Co-Authored-By" 검색 결과
.git/COMMIT_EDITMSG:7:Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>
```

### 실패 조건
- 검색 결과가 0건인 경우
- 새 브랜딩이 적용되지 않은 경우

---

## Scenario 5: 컨텍스트 명확성 검증

### Given-When-Then

```gherkin
Given README.md 내부 섹션을 작성할 때
When "Alfred 페르소나", "Alfred SuperAgent" 등 컨텍스트가 명확하면
Then "Alfred" 단독 사용을 허용해야 한다
And 독자가 "Alfred"가 MoAI-ADK AI Agent를 지칭함을 명확히 이해할 수 있어야 한다
When 외부 노출 섹션(커밋 메시지, PR 제목)을 작성하면
Then "AI-Agent Alfred" 전체 명칭을 사용해야 한다
And 컨텍스트 없이도 프로젝트 정체성이 명확해야 한다
```

### 검증 방법

```bash
# README.md에서 "Alfred" 단독 사용 확인 (컨텍스트 명확 시 허용)
rg "Alfred" -n README.md

# README.md에서 전체 명칭 사용 확인 (외부 노출 부분)
rg "AI-Agent Alfred" -n README.md

# 커밋 메시지에서 전체 명칭 사용 확인
git log --all --grep="Alfred" --pretty=format:"%s"
```

### 예상 출력

**README.md 내부 (컨텍스트 명확)**:
```markdown
## Alfred 페르소나

Alfred는 9개 전문 에이전트를 조율합니다.
```

**README.md Hero (외부 노출)**:
```markdown
# MoAI-ADK

**AI-Agent Alfred**와 함께하는 SPEC-First TDD 개발
```

### 실패 조건
- 외부 노출 부분에서 "Alfred" 단독 사용
- 컨텍스트 불명확한 내부 섹션에서 전체 명칭 누락

---

## Scenario 6: Git 히스토리 보존 검증

### Given-When-Then

```gherkin
Given 기존 Git 커밋 히스토리가 존재할 때
When 새로운 브랜딩을 적용하면
Then 기존 커밋 메시지는 변경하지 않아야 한다
And 새로운 커밋부터만 새 브랜딩을 사용해야 한다
And `git log --oneline` 실행 시 기존 히스토리가 보존되어야 한다
And Git 히스토리 재작성 명령어(`git rebase -i`, `git commit --amend`)를 사용하지 않아야 한다
```

### 검증 방법

```bash
# 전체 Git 히스토리 확인
git log --oneline

# 기존 "Claude Code" 커밋 확인 (변경되지 않았는지)
git log --all --grep="Generated with \[Claude Code\]"

# 새 브랜딩 커밋 확인
git log --all --grep="Generated with AI-Agent Alfred"

# 히스토리 무결성 확인 (커밋 해시 변경 없음)
git log --pretty=format:"%H %s" | head -20
```

### 예상 출력

```
# 기존 커밋 (보존)
3c41c3a feat(init): Add non-interactive mode support with TTY detection
🤖 Generated with [Claude Code](https://claude.com/claude-code)

# 새로운 커밋 (새 브랜딩)
abcdef1 feat(brand): Update branding to AI-Agent Alfred
🤖 Generated with AI-Agent Alfred
```

### 실패 조건
- 기존 커밋 해시가 변경된 경우 (히스토리 재작성)
- 기존 커밋 메시지가 수정된 경우
- `git rebase -i` 또는 `git commit --amend` 사용 흔적

---

## Scenario 7: README.md 브랜딩 강화 (선택적)

### Given-When-Then

```gherkin
Given README.md Hero 섹션을 업데이트할 때
When 프로젝트 소개 문구를 작성하면
Then "AI-Agent Alfred"를 명시해야 한다
And 독자가 프로젝트 정체성을 명확히 이해할 수 있어야 한다
And 컨텍스트 명확한 내부 섹션에서는 "Alfred" 단독 사용을 허용해야 한다
```

### 검증 방법

```bash
# README.md Hero 섹션 확인
head -20 README.md

# 브랜딩 문구 확인
rg "AI-Agent Alfred" -n README.md

# 컨텍스트 명확한 "Alfred" 사용 확인
rg "Alfred 페르소나|Alfred SuperAgent" -n README.md
```

### 예상 출력

**Hero 섹션**:
```markdown
# MoAI-ADK

**AI-Agent Alfred**와 함께하는 SPEC-First TDD 개발
```

**내부 섹션**:
```markdown
## Alfred 페르소나

Alfred는 모두의AI(MoAI)가 개발한 공식 SuperAgent입니다.
```

### 실패 조건
- Hero 섹션에서 전체 명칭이 누락된 경우
- 컨텍스트 불명확한 섹션에서 "Alfred" 단독 사용

---

## Scenario 8: 에이전트 지침서 브랜딩 통일 (선택적)

### Given-When-Then

```gherkin
Given 에이전트 지침서 파일을 업데이트할 때
When git-manager, code-builder 등의 에이전트 문서를 작성하면
Then "AI-Agent Alfred 생태계"를 참조해야 한다
And 에이전트 간 협업 설명에서 일관된 브랜딩을 유지해야 한다
And 커밋 템플릿 예시에 새 브랜딩을 포함해야 한다
```

### 검증 방법

```bash
# 에이전트 지침서 파일 검색
find .moai -name "*agent*.md"

# 브랜딩 참조 확인
rg "AI-Agent Alfred" -n .moai/

# 커밋 템플릿 예시 확인
rg "Generated with AI-Agent Alfred" -n .moai/
```

### 예상 출력

```
.moai/memory/agents/git-manager.md:10:**역할**: AI-Agent Alfred 생태계에서 Git 워크플로우 담당
.moai/memory/agents/git-manager.md:50:🤖 Generated with AI-Agent Alfred
```

### 실패 조건
- 에이전트 지침서에서 브랜딩 불일치
- 커밋 템플릿 예시에 구 브랜딩 사용

---

## 품질 게이트 (Quality Gates)

### 필수 통과 조건

1. **브랜딩 일관성**:
   - [ ] `rg "Generated with AI-Agent Alfred" -n` 결과 > 0건
   - [ ] `rg "Co-Authored-By: AI-Agent Alfred" -n` 결과 > 0건

2. **CLAUDE.md 업데이트**:
   - [ ] Line 14: "AI-Agent Alfred 워크플로우" 확인
   - [ ] Line 52: "MoAI-ADK 설정" 확인
   - [ ] `rg "Claude Code" -n CLAUDE.md` 결과 0건 (외부 노출 부분)

3. **Git 히스토리 보존**:
   - [ ] 기존 커밋 해시 변경 없음
   - [ ] 새로운 커밋부터 새 브랜딩 적용
   - [ ] `git log --oneline` 검증 성공

4. **컨텍스트 명확성**:
   - [ ] 외부 노출: 전체 명칭 사용
   - [ ] 내부 섹션: 컨텍스트 명확 시 "Alfred" 허용

### 선택적 통과 조건

5. **README.md 브랜딩**:
   - [ ] Hero 섹션에 "AI-Agent Alfred" 명시 (선택적)
   - [ ] 내부 섹션에서 컨텍스트 명확 시 "Alfred" 허용

6. **GitHub 템플릿**:
   - [ ] PR 템플릿에 브랜딩 문구 포함
   - [ ] Issue 템플릿에 브랜딩 문구 포함

7. **에이전트 지침서**:
   - [ ] 에이전트 문서에서 일관된 브랜딩
   - [ ] 커밋 템플릿 예시 업데이트

---

## Definition of Done (완료 조건)

### 최소 완료 조건 (Minimum Viable)

- ✅ CLAUDE.md에서 "Claude Code" → "AI-Agent Alfred" 변경 완료
- ✅ 새로운 Git 커밋에 새 브랜딩 적용
- ✅ Co-Authored-By 서명 업데이트
- ✅ Git 히스토리 보존 확인
- ✅ 브랜딩 일관성 검증 스크립트 통과

### 완전 완료 조건 (Full)

- ✅ 최소 완료 조건 모두 충족
- ✅ README.md 브랜딩 강화 (선택적)
- ✅ GitHub 템플릿 업데이트
- ✅ 에이전트 지침서 브랜딩 통일
- ✅ 모든 품질 게이트 통과

---

## 검증 체크리스트

### Phase 1: CLAUDE.md 검증
```bash
# 1. Line 14 확인
sed -n '14p' CLAUDE.md | grep "AI-Agent Alfred 워크플로우"

# 2. Line 52 확인
sed -n '52p' CLAUDE.md | grep "MoAI-ADK 설정"

# 3. "Claude Code" 제거 확인
rg "Claude Code" -n CLAUDE.md
```

### Phase 2: Git 커밋 검증
```bash
# 1. 최근 커밋 메시지 확인
git log -1 --pretty=format:"%B"

# 2. 브랜딩 문구 확인
git log -1 --pretty=format:"%B" | grep "Generated with AI-Agent Alfred"

# 3. Co-Authored-By 확인
git log -1 --pretty=format:"%B" | grep "Co-Authored-By: AI-Agent Alfred"
```

### Phase 3: 전체 브랜딩 검증
```bash
# 1. 새 브랜딩 전체 검색
rg "Generated with AI-Agent Alfred" -n

# 2. Co-Authored-By 전체 검색
rg "Co-Authored-By: AI-Agent Alfred" -n

# 3. 결과 개수 확인
echo "New branding count: $(rg 'Generated with AI-Agent Alfred' -n | wc -l)"
echo "Co-Authored-By count: $(rg 'Co-Authored-By: AI-Agent Alfred' -n | wc -l)"
```

### Phase 4: Git 히스토리 보존 검증
```bash
# 1. 전체 히스토리 확인
git log --oneline | head -20

# 2. 기존 "Claude Code" 커밋 확인 (보존 확인)
git log --all --grep="Generated with \[Claude Code\]"

# 3. 새 브랜딩 커밋 확인
git log --all --grep="Generated with AI-Agent Alfred"
```

---

_이 수락 기준은 Given-When-Then 형식으로 작성되어 있으며, 모든 조건은 검증 가능합니다._
