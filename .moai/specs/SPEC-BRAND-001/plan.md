# SPEC-BRAND-001 구현 계획

## 개요

**목표**: MoAI-ADK 프로젝트의 브랜딩을 "Claude Code"에서 "AI-Agent Alfred"로 통일하여 프로젝트 정체성을 강화하고 일관성을 확보합니다.

**핵심 전략**: Git 히스토리를 보존하면서, 새로운 커밋부터 점진적으로 새 브랜딩을 적용합니다.

---

## Phase 1: CLAUDE.md 문서 업데이트

### 우선순위: High

#### 작업 내용
1. **Line 14 변경**:
   ```markdown
   # 변경 전
   - **역할**: Claude Code 워크플로우의 중앙 오케스트레이터

   # 변경 후
   - **역할**: AI-Agent Alfred 워크플로우의 중앙 오케스트레이터
   ```

2. **Line 52 변경**:
   ```markdown
   # 변경 전
   | **cc-manager** 🛠️ | 데브옵스 엔지니어 | Claude Code 설정 | `@agent-cc-manager` | 설정 필요 시 |

   # 변경 후
   | **cc-manager** 🛠️ | 데브옵스 엔지니어 | MoAI-ADK 설정 | `@agent-cc-manager` | 설정 필요 시 |
   ```

#### 검증 방법
```bash
# "Claude Code" 참조가 외부 노출 부분에서 제거되었는지 확인
rg "Claude Code" -n CLAUDE.md

# 새 브랜딩 확인
rg "AI-Agent Alfred" -n CLAUDE.md
```

#### 예상 결과
- CLAUDE.md에서 "Claude Code" 참조 0건 (외부 노출 부분)
- "AI-Agent Alfred" 참조 1건 이상

---

## Phase 2: Git 커밋 템플릿 적용 (미래 커밋)

### 우선순위: High

#### 작업 내용

**커밋 메시지 템플릿**:
```
<type>(<scope>): <subject>

<body>

🤖 Generated with AI-Agent Alfred

Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>
```

**템플릿 구조**:
- **브랜딩 문구**: `🤖 Generated with AI-Agent Alfred`
- **Co-Authored-By**: `AI-Agent Alfred <noreply@anthropic.com>`
- **이메일 유지**: `noreply@anthropic.com` (Claude 연동)

#### 적용 방법

**수동 커밋 시**:
```bash
git commit -m "$(cat <<'EOF'
feat(brand): Update branding to AI-Agent Alfred

Update CLAUDE.md to reflect new branding.

🤖 Generated with AI-Agent Alfred

Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>
EOF
)"
```

**git-manager 에이전트 위임 시**:
- git-manager가 자동으로 새 템플릿 적용
- HEREDOC 형식으로 포맷팅 보장

#### 검증 방법
```bash
# 최근 커밋 메시지 확인
git log -1 --pretty=format:"%B"

# 새 브랜딩 문구 확인
git log --all --grep="Generated with AI-Agent Alfred"
```

#### 예상 결과
- 새로운 커밋 메시지에 "🤖 Generated with AI-Agent Alfred" 포함
- Co-Authored-By 서명에 "AI-Agent Alfred" 포함

---

## Phase 3: README.md 검토 및 업데이트 (선택적)

### 우선순위: Medium

#### 작업 내용

**Hero 섹션 강화 (선택적)**:
```markdown
# MoAI-ADK

**AI-Agent Alfred**와 함께하는 SPEC-First TDD 개발

Alfred는 모두의AI(MoAI)가 개발한 공식 SuperAgent로, 일관성 있고 추적 가능한 코드 품질을 보장합니다.
```

**내부 섹션 (컨텍스트 명확 시 "Alfred" 허용)**:
```markdown
## Alfred 페르소나

Alfred는 9개 전문 에이전트를 조율하는 중앙 오케스트레이터입니다.
```

#### 검증 방법
```bash
# README.md에서 브랜딩 확인
rg "AI-Agent Alfred" -n README.md

# 컨텍스트 명확한 "Alfred" 사용 확인
rg "Alfred" -n README.md
```

#### 예상 결과
- Hero 섹션: "AI-Agent Alfred" 명시 (선택적)
- 내부 설명: "Alfred" 단독 사용 허용

---

## Phase 4: GitHub 템플릿 업데이트 (선택적)

### 우선순위: Medium

#### 작업 내용

**PR 템플릿** (`.github/PULL_REQUEST_TEMPLATE.md`):
```markdown
## Summary
- Brief description of changes

## Test Plan
- [ ] Test 1
- [ ] Test 2

## Related Issues
- Closes #XXX

🤖 Generated with AI-Agent Alfred
```

**Issue 템플릿** (`.github/ISSUE_TEMPLATE/feature_request.md`):
```markdown
---
name: Feature Request
about: Suggest a new feature
---

## Description
[Describe the feature]

## Use Case
[Explain the use case]

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

🤖 Generated with AI-Agent Alfred
```

#### 검증 방법
```bash
# GitHub 템플릿 존재 여부 확인
ls -la .github/

# 템플릿 내용 확인
rg "Generated with AI-Agent Alfred" -n .github/
```

#### 예상 결과
- PR/Issue 템플릿에 브랜딩 문구 포함
- GitHub 웹 UI에서 확인 가능

---

## Phase 5: 에이전트 지침서 업데이트 (선택적)

### 우선순위: Low

#### 작업 내용

**에이전트 지침서 예시**:
```markdown
# git-manager Agent

**역할**: AI-Agent Alfred 생태계에서 Git 워크플로우를 담당하는 릴리스 엔지니어

## 커밋 메시지 템플릿

모든 커밋 메시지는 다음 형식을 따릅니다:

```
🤖 Generated with AI-Agent Alfred

Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>
```
```

#### 검증 방법
```bash
# 에이전트 지침서 존재 여부 확인
find .moai -name "*agent*.md"

# 브랜딩 참조 확인
rg "AI-Agent Alfred" -n .moai/
```

#### 예상 결과
- 에이전트 지침서에서 일관된 브랜딩 사용

---

## 마일스톤 및 우선순위

### 1차 목표 (필수)
- ✅ CLAUDE.md 업데이트 (Line 14, 52)
- ✅ Git 커밋 템플릿 적용 (미래 커밋부터)
- ✅ 브랜딩 일관성 검증 (`rg` 명령어)

### 2차 목표 (권장)
- 📋 README.md 검토 및 업데이트 (선택적)
- 📋 GitHub 템플릿 업데이트 (존재 시)

### 3차 목표 (선택적)
- 📋 에이전트 지침서 업데이트

---

## 기술적 접근 방법

### 1. Git 히스토리 보존 전략

**금지 사항**:
- ❌ `git rebase -i` (히스토리 재작성)
- ❌ `git commit --amend` (기존 커밋 수정)
- ❌ `git filter-branch` (히스토리 필터링)

**허용 사항**:
- ✅ 새로운 커밋에 새 브랜딩 적용
- ✅ 문서 파일 직접 수정 (CLAUDE.md, README.md)
- ✅ `.github/` 템플릿 파일 추가/수정

### 2. 브랜딩 일관성 검증 스크립트

```bash
#!/bin/bash
# scripts/verify-branding.sh

echo "Verifying branding consistency..."

# Check new branding
echo "1. Checking new branding..."
rg "Generated with AI-Agent Alfred" -n

# Check Co-Authored-By
echo "2. Checking Co-Authored-By..."
rg "Co-Authored-By: AI-Agent Alfred" -n

# Check CLAUDE.md
echo "3. Checking CLAUDE.md..."
rg "Claude Code" -n CLAUDE.md && echo "WARNING: Old branding found in CLAUDE.md" || echo "OK"

# Summary
echo "Branding verification complete."
```

### 3. 컨텍스트 명확성 가이드

**전체 명칭 사용 (필수)**:
- Git 커밋 메시지 푸터
- GitHub PR/Issue 제목 및 본문
- 외부 블로그, 발표 자료
- 프로젝트 소개 섹션

**"Alfred" 단독 사용 허용**:
- README.md 내부 설명 (예: "Alfred 페르소나")
- 에이전트 간 협업 설명 (예: "Alfred가 조율")
- 컨텍스트가 명확한 문서 내부

---

## 리스크 및 대응 방안

### Risk 1: Git 히스토리 재작성 실수
**완화 전략**:
- Git 작업은 git-manager 에이전트에게만 위임
- 수동 커밋 시 HEREDOC 템플릿 사용
- `git log` 검증 후 원격 푸시

### Risk 2: 기존 "Claude Code" 참조 누락
**완화 전략**:
- `rg "Claude Code" -n` 전체 스캔
- CLAUDE.md 외 다른 문서도 검증
- Legacy 참조는 주석으로 보존 가능

### Risk 3: 브랜딩 혼용 (일부 "Alfred", 일부 전체 명칭)
**완화 전략**:
- 컨텍스트 명확성 가이드 준수
- 외부 노출은 항상 전체 명칭
- 문서 작성 시 스타일 가이드 참조

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] CLAUDE.md에서 "Claude Code" → "AI-Agent Alfred" 변경 완료
- [ ] `rg "Claude Code" -n CLAUDE.md` 결과 0건 (외부 노출 부분)
- [ ] 새로운 커밋 메시지에 "🤖 Generated with AI-Agent Alfred" 포함
- [ ] Co-Authored-By 서명에 "AI-Agent Alfred" 포함
- [ ] Git 히스토리 보존 확인 (`git log --oneline` 검증)

### 선택적 조건
- [ ] README.md 브랜딩 강화 (Hero 섹션)
- [ ] GitHub 템플릿 업데이트
- [ ] 에이전트 지침서 브랜딩 통일

### 검증 조건
- [ ] `rg "Generated with AI-Agent Alfred" -n` 결과 > 0건
- [ ] `rg "Co-Authored-By: AI-Agent Alfred" -n` 결과 > 0건
- [ ] 브랜딩 일관성 검증 스크립트 통과

---

## 다음 단계

1. **구현**: CLAUDE.md 업데이트 (Phase 1)
2. **커밋**: 새 브랜딩으로 커밋 생성 (Phase 2)
3. **검증**: `rg` 명령어로 브랜딩 일관성 확인
4. **선택적**: README.md, GitHub 템플릿 업데이트 (Phase 3-4)
5. **동기화**: `/alfred:3-sync` 실행 (TAG 체인 검증)

---

_이 계획은 우선순위 기반 마일스톤으로 구성되어 있으며, 시간 예측은 포함하지 않습니다._
