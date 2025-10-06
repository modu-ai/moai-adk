---
id: BRAND-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @goos
---

# @SPEC:BRAND-001: AI-Agent Alfred 브랜딩 일관성 통일

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: Claude Code → AI-Agent Alfred 브랜딩 변경
- **AUTHOR**: @goos
- **REVIEW**: @AI-Alfred
- **SCOPE**: Git 메시지, 문서, 커밋 서명, PR/Issue 전체 업데이트
- **CONTEXT**: 프로젝트 정체성 강화 및 브랜딩 일관성 확보

---

## Environment (환경 및 전제조건)

### 실행 환경
- **프로젝트**: MoAI-ADK (모두의AI Agentic Development Kit)
- **Git**: 브랜치 전략, 커밋 메시지, PR/Issue 템플릿
- **문서**: CLAUDE.md, README.md, 에이전트 지침서

### 기술 스택
- **VCS**: Git, GitHub (Issues, Pull Requests)
- **문서 형식**: Markdown
- **브랜딩 대상**: 커밋 메시지, PR 본문, Issue 본문, 프로젝트 문서

### 제약사항
- 기존 Git 히스토리는 변경하지 않음 (새로운 커밋부터 적용)
- README.md의 핵심 설명 섹션은 컨텍스트 명확 시 "Alfred" 단독 사용 허용
- Co-Authored-By 이메일 주소는 유지 (`noreply@anthropic.com`)

---

## Assumptions (가정사항)

1. **브랜딩 범위 가정**:
   - Git 커밋 메시지 푸터 (footer signature)
   - GitHub PR/Issue 템플릿
   - CLAUDE.md 프로젝트 문서
   - README.md 주요 설명 섹션

2. **컨텍스트 명확성 가정**:
   - "Alfred SuperAgent", "Alfred 페르소나" 등 컨텍스트가 명확한 경우 "Alfred" 단독 사용 허용
   - 외부 노출(커밋 메시지, PR 제목)은 항상 전체 명칭 사용

3. **기존 히스토리 가정**:
   - Git 히스토리 재작성 불가 (rebase --interactive 금지)
   - 기존 "Claude Code" 참조는 Legacy로 보존

4. **일관성 우선순위 가정**:
   - 우선순위 High: Git 커밋 서명, PR/Issue 템플릿
   - 우선순위 Medium: CLAUDE.md 문서
   - 우선순위 Low: README.md 내부 설명 (컨텍스트 명확 시 "Alfred" 허용)

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (기본 기능)

**UR-001**: 시스템은 모든 Git 커밋 메시지 푸터에 "🤖 Generated with AI-Agent Alfred" 문구를 포함해야 한다

**UR-002**: 시스템은 모든 Git 커밋에 "Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>" 서명을 포함해야 한다

**UR-003**: 시스템은 프로젝트 문서에서 브랜딩을 "AI-Agent Alfred"로 명시해야 한다

**UR-004**: 시스템은 외부 노출(커밋, PR 제목, Issue 제목)에서 전체 명칭을 사용해야 한다

**UR-005**: 시스템은 기존 Git 히스토리를 변경하지 않아야 한다 (새로운 커밋부터 적용)

---

### Event-driven Requirements (이벤트 기반)

**ER-001**: WHEN Git 커밋 생성 시, 시스템은 커밋 메시지 푸터에 "🤖 Generated with AI-Agent Alfred" 문구를 삽입해야 한다
- **조건**: git-manager 또는 수동 커밋 생성
- **동작**: 커밋 메시지 마지막 줄에 브랜딩 문구 추가
- **출력**: Git log에서 확인 가능

**ER-002**: WHEN GitHub PR/Issue 생성 시, 시스템은 본문에 "🤖 Generated with AI-Agent Alfred" 문구를 포함해야 한다
- **조건**: `gh pr create` 또는 `gh issue create` 실행
- **동작**: PR/Issue 본문 마지막에 브랜딩 문구 추가
- **출력**: GitHub 웹 UI에서 확인 가능

**ER-003**: WHEN CLAUDE.md 문서 업데이트 시, 시스템은 "AI-Agent Alfred" 전체 명칭을 사용해야 한다
- **조건**: 프로젝트 정체성 강조가 필요한 섹션
- **동작**: "Claude Code" → "AI-Agent Alfred" 변경
- **예외**: 에이전트 이름 단독 사용 시 "Alfred" 허용 (예: "Alfred 페르소나")

**ER-004**: WHEN 커밋 서명(Co-Authored-By) 생성 시, 시스템은 "AI-Agent Alfred" 명칭을 사용해야 한다
- **조건**: 모든 Git 커밋
- **동작**: `Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>` 형식
- **이메일**: `noreply@anthropic.com` 유지 (Claude 연동)

---

### State-driven Requirements (상태 기반)

**SR-001**: WHILE 프로젝트 전체에서 "AI-Agent Alfred" 브랜딩이 일관되어야 한다
- **상태**: 모든 Git 커밋, PR, Issue, 문서
- **동작**: 일관된 브랜딩 문구 사용
- **검증**: `rg "Generated with AI-Agent Alfred" -n` 결과 확인

**SR-002**: WHILE 문서 작성 중일 때, 컨텍스트가 명확하면 "Alfred" 단독 사용을 허용할 수 있다
- **상태**: README.md 내부 섹션, 에이전트 설명
- **조건**: "Alfred SuperAgent", "Alfred 페르소나" 등 명확한 컨텍스트
- **제약**: 외부 노출(커밋, PR 제목)은 항상 전체 명칭 사용

**SR-003**: WHILE 기존 Git 히스토리는 변경하지 않아야 한다
- **상태**: 과거 커밋 메시지 보존
- **동작**: 새로운 커밋부터만 새 브랜딩 적용
- **금지**: `git rebase -i`, `git commit --amend` (히스토리 재작성)

---

### Optional Features (선택적 기능)

**OF-001**: WHERE README.md 업데이트 시, 브랜딩 강조가 필요하면 "AI-Agent Alfred"를 사용할 수 있다
- **조건**: 프로젝트 소개, 핵심 기능 설명
- **구현**: README.md 상단 Hero 섹션
- **우선순위**: Medium

**OF-002**: WHERE 에이전트 지침서 업데이트 시, 브랜딩을 명시할 수 있다
- **조건**: 에이전트별 MD 파일 업데이트
- **구현**: 각 에이전트가 "AI-Agent Alfred"를 참조
- **우선순위**: Low

---

### Constraints (제약사항)

**C-001**: IF Git 히스토리 재작성이 필요하면, 시스템은 작업을 거부해야 한다
- **조건**: `git rebase -i`, `git commit --amend` 시도
- **동작**: 경고 메시지 출력 및 작업 중단
- **이유**: 기존 커밋 히스토리 보존 원칙

**C-002**: IF "Claude Code" 참조가 외부 노출되면, "AI-Agent Alfred"로 변경해야 한다
- **조건**: 커밋 메시지, PR 제목, Issue 제목
- **동작**: 자동 변경 또는 수동 검토 요청
- **검증**: `rg "Claude Code" -n` 결과 확인

**C-003**: Co-Authored-By 이메일 주소는 `noreply@anthropic.com`을 유지해야 한다
- **조건**: Claude API 연동 유지
- **동작**: 이메일 변경 금지
- **이유**: Claude 서비스와의 통합성 유지

**C-004**: IF 컨텍스트가 불명확하면, "Alfred" 대신 "AI-Agent Alfred"를 사용해야 한다
- **조건**: 외부 문서, 블로그, 발표 자료
- **동작**: 전체 명칭 사용
- **예시**: "Alfred는..." (X) → "AI-Agent Alfred는..." (O)

---

## Specifications (상세 명세)

### 1. Git 커밋 메시지 템플릿

**변경 전**:
```
feat(auth): Add JWT authentication

Implement JWT-based authentication system

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**변경 후**:
```
feat(auth): Add JWT authentication

Implement JWT-based authentication system

🤖 Generated with AI-Agent Alfred

Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>
```

---

### 2. CLAUDE.md 문서 업데이트

**변경 대상 1**: Line 14
```markdown
# 변경 전
- **역할**: Claude Code 워크플로우의 중앙 오케스트레이터

# 변경 후
- **역할**: AI-Agent Alfred 워크플로우의 중앙 오케스트레이터
```

**변경 대상 2**: Line 52 (cc-manager 설명)
```markdown
# 변경 전
| **cc-manager** 🛠️ | 데브옵스 엔지니어 | Claude Code 설정 | `@agent-cc-manager` | 설정 필요 시 |

# 변경 후 (옵션 1 - 전체 명칭)
| **cc-manager** 🛠️ | 데브옵스 엔지니어 | MoAI-ADK 설정 | `@agent-cc-manager` | 설정 필요 시 |

# 변경 후 (옵션 2 - 컨텍스트 명확 시)
| **cc-manager** 🛠️ | 데브옵스 엔지니어 | Alfred 설정 | `@agent-cc-manager` | 설정 필요 시 |
```

**권장**: 옵션 1 (MoAI-ADK 설정) - cc-manager가 Claude Code 특정 설정이 아닌 MoAI-ADK 프로젝트 전체 설정을 관리하므로

---

### 3. GitHub PR/Issue 템플릿 업데이트

**PR 템플릿 예시**:
```markdown
## Summary
- Implement JWT authentication system

## Test Plan
- [ ] Unit tests pass
- [ ] Integration tests pass

🤖 Generated with AI-Agent Alfred
```

**Issue 템플릿 예시**:
```markdown
## Description
Add support for JWT authentication

## Acceptance Criteria
- [ ] JWT token generation
- [ ] Token validation

🤖 Generated with AI-Agent Alfred
```

---

### 4. README.md 업데이트 가이드

**외부 노출 섹션 (전체 명칭 사용)**:
```markdown
# MoAI-ADK

**AI-Agent Alfred**와 함께하는 SPEC-First TDD 개발
```

**내부 설명 섹션 (컨텍스트 명확 시 "Alfred" 허용)**:
```markdown
## Alfred 페르소나

Alfred는 모두의AI(MoAI)가 개발한 SuperAgent입니다.
```

---

### 5. 변경 대상 파일 목록

#### 우선순위 High
1. **CLAUDE.md** (2곳):
   - Line 14: "Claude Code 워크플로우" → "AI-Agent Alfred 워크플로우"
   - Line 52: "Claude Code 설정" → "MoAI-ADK 설정"

2. **Git 커밋 템플릿** (미래 커밋):
   - 모든 새로운 커밋 메시지 푸터
   - Co-Authored-By 서명

#### 우선순위 Medium
3. **README.md**:
   - Hero 섹션 브랜딩 강화 (선택적)
   - 컨텍스트 명확한 내부 섹션은 "Alfred" 허용

4. **GitHub 템플릿**:
   - `.github/PULL_REQUEST_TEMPLATE.md` (존재 시)
   - `.github/ISSUE_TEMPLATE/*.md` (존재 시)

#### 우선순위 Low
5. **에이전트 지침서** (존재 시):
   - `.moai/memory/agents/*.md`

---

## Acceptance Criteria (수락 기준)

### AC1: Git 커밋 메시지 브랜딩
```gherkin
GIVEN AI-Agent Alfred가 커밋을 생성할 때
WHEN 커밋 메시지 푸터를 작성하면
THEN "🤖 Generated with AI-Agent Alfred" 문구가 포함되어야 한다
AND "Co-Authored-By: AI-Agent Alfred <noreply@anthropic.com>" 서명이 포함되어야 한다
AND Git log에서 확인 가능해야 한다
```

### AC2: CLAUDE.md 문서 브랜딩
```gherkin
GIVEN CLAUDE.md 문서를 업데이트할 때
WHEN 브랜딩 문구를 작성하면
THEN Line 14: "AI-Agent Alfred 워크플로우의 중앙 오케스트레이터"로 변경되어야 한다
AND Line 52: "MoAI-ADK 설정"으로 변경되어야 한다
AND `rg "Claude Code" -n CLAUDE.md` 결과가 0건이어야 한다 (외부 노출 부분)
```

### AC3: GitHub PR/Issue 브랜딩
```gherkin
GIVEN GitHub PR/Issue를 생성할 때
WHEN 본문 템플릿을 작성하면
THEN "🤖 Generated with AI-Agent Alfred" 문구가 포함되어야 한다
AND GitHub 웹 UI에서 확인 가능해야 한다
AND 기존 PR/Issue는 변경하지 않아야 한다
```

### AC4: 브랜딩 일관성 검증
```gherkin
GIVEN 프로젝트 전체 파일을 검증할 때
WHEN `rg "Generated with AI-Agent Alfred" -n` 실행하면
THEN 결과가 > 0건이어야 한다
AND CLAUDE.md, 커밋 메시지에서 확인되어야 한다
WHEN `rg "Co-Authored-By: AI-Agent Alfred" -n` 실행하면
THEN 결과가 > 0건이어야 한다
AND Git 로그에서 확인되어야 한다
```

### AC5: 컨텍스트 명확성 검증
```gherkin
GIVEN README.md 내부 섹션을 작성할 때
WHEN "Alfred 페르소나", "Alfred SuperAgent" 등 컨텍스트가 명확하면
THEN "Alfred" 단독 사용을 허용해야 한다
WHEN 외부 노출(커밋 메시지, PR 제목)을 작성하면
THEN "AI-Agent Alfred" 전체 명칭을 사용해야 한다
```

### AC6: Git 히스토리 보존
```gherkin
GIVEN 기존 Git 커밋 히스토리가 존재할 때
WHEN 새로운 브랜딩을 적용하면
THEN 기존 커밋 메시지는 변경하지 않아야 한다
AND 새로운 커밋부터만 새 브랜딩을 사용해야 한다
AND `git log --oneline` 결과에서 기존 히스토리가 보존되어야 한다
```

### AC7: 에이전트 지침서 업데이트 (선택적)
```gherkin
GIVEN 에이전트 지침서 파일이 존재할 때
WHEN 브랜딩 참조를 업데이트하면
THEN "AI-Agent Alfred"를 명시해야 한다
AND 에이전트 간 협업 설명에서 일관된 브랜딩을 유지해야 한다
```

---

## Traceability (@TAG 체인)

### TAG 체인 구조
```
@SPEC:BRAND-001 (본 문서)
  ↓
@TEST:BRAND-001 (없음 - 문서 업데이트는 수동 검증)
  ↓
@CODE:BRAND-001 (없음 - Git 템플릿 및 문서 변경)
  ├─ CLAUDE.md (Line 14, 52)
  ├─ Git 커밋 템플릿 (미래 커밋)
  └─ GitHub 템플릿 (선택적)
  ↓
@DOC:BRAND-001 (본 SPEC 문서)
```

### 검증 명령어
```bash
# SPEC 문서 확인
rg '@SPEC:BRAND-001' -n .moai/specs/

# 브랜딩 문구 확인
rg "Generated with AI-Agent Alfred" -n
rg "Co-Authored-By: AI-Agent Alfred" -n

# CLAUDE.md 업데이트 확인
rg "Claude Code" -n CLAUDE.md

# 전체 브랜딩 일관성 검증
rg "AI-Agent Alfred" -n
```

---

## 다음 단계

### 구현 단계
1. CLAUDE.md 업데이트 (Line 14, 52)
2. Git 커밋 템플릿 업데이트 (미래 커밋부터 적용)
3. README.md 검토 및 업데이트 (선택적)
4. GitHub 템플릿 업데이트 (존재 시)

### 검증 단계
1. `rg "Generated with AI-Agent Alfred" -n` 실행
2. `rg "Co-Authored-By: AI-Agent Alfred" -n` 실행
3. `rg "Claude Code" -n CLAUDE.md` 실행 (결과 0건 확인)
4. Git log에서 새로운 커밋 메시지 확인

### 동기화 단계
1. `/alfred:3-sync` 실행 (문서 동기화)
2. TAG 체인 검증 (수동)
3. Living Document 생성

---

_이 문서는 SPEC-First TDD 방법론에 따라 작성되었습니다._
_브랜딩 일관성 확보를 통한 프로젝트 정체성 강화가 목표입니다._
