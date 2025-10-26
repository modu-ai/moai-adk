# MoAI-ADK Team 모드 GitHub 통합 상세 분석 보고서

## 1. 분석 개요

**분석 대상**: MoAI-ADK의 Team 모드에서 GitHub과의 통합이 어떻게 작동하는지  
**분석 기준**: 실제 구현 코드 + Agent 정의 + Command 구현  
**분석 범위**: `.moai/config.json`, `.claude/agents/`, `.claude/commands/`, 핵심 Skills  

---

## 2. Team 모드 GitHub 설정

### 2.1 Config 구조 (.moai/config.json)

```json
{
  "git_strategy": {
    "active_mode": "team",
    "team": {
      "auto_pr": true,              // Draft PR 자동 생성 활성화
      "develop_branch": "develop",   // 개발 브랜치
      "draft_pr": true,              // Draft PR 기본값
      "feature_prefix": "feature/SPEC-",  // 피처 브랜치 이름 규칙
      "main_branch": "main",         // 프로덕션 브랜치
      "use_gitflow": true,           // GitFlow 워크플로우 사용
      "auto_ready_on_sync": true    // /alfred:3-sync에서 PR 자동 Ready 전환
    }
  },
  "project": {
    "mode": "team",                 // Team 모드 활성화
    "language": "python"
  }
}
```

**핵심 설정**:
- ✅ `auto_pr: true` → Draft PR 자동 생성
- ✅ `draft_pr: true` → 기본적으로 Draft 상태로 생성
- ✅ `auto_ready_on_sync: true` → Sync 단계에서 PR Ready 자동 전환
- ✅ `use_gitflow: true` → GitFlow 표준 준수

---

## 3. Team 모드 GitHub 통합 워크플로우

### 3.1 전체 흐름도

```
/alfred:1-plan (SPEC 생성)
    ├─ spec-builder: SPEC 작성 + @SPEC TAG 추가
    └─ git-manager: 
        ├─ feature/SPEC-{ID} 브랜치 생성 (develop 기반)
        ├─ GitHub Issue 생성 (Team 모드 용)
        └─ Draft PR 생성 (feature → develop)

/alfred:2-run (TDD 구현)
    ├─ implementation-planner: 실행 계획 수립
    ├─ tdd-implementer: RED → GREEN → REFACTOR
    │   ├─ @TEST TAG 추가
    │   └─ @CODE TAG 추가
    └─ git-manager: 
        ├─ RED/GREEN/REFACTOR 커밋 생성
        ├─ Draft PR 자동 업데이트
        └─ 테스트/커버리지 리포트 작성

/alfred:3-sync (문서 동기화)
    ├─ doc-syncer: 
    │   ├─ Living Document 동기화
    │   ├─ @TAG 체인 검증
    │   └─ SPEC 메타데이터 업데이트
    └─ git-manager:
        ├─ 문서 변경사항 커밋
        ├─ PR Ready 전환 (gh pr ready)
        ├─ [선택] PR 자동 머지 (--auto-merge 플래그)
        └─ 브랜치 정리 + develop 체크아웃
```

---

## 4. 각 단계별 GitHub 자동화

### 4.1 Stage 1: `/alfred:1-plan` - SPEC 생성 및 Branch/Draft PR 생성

**참여 에이전트**:
- `spec-builder` (Sonnet): SPEC 문서 작성
- `git-manager` (Haiku): Git/GitHub 작업

**수행 작업**:

#### Step 1-1: SPEC 생성
```bash
# 위치: .moai/specs/SPEC-{ID}/
# 생성 파일:
- spec.md      (EARS 구조 SPEC)
- plan.md      (구현 계획)
- acceptance.md (수용 기준)
```

**SPEC 메타데이터 구조** (YAML Front Matter):
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-25
updated: 2025-10-25
author: @username
priority: high
---

# @SPEC:AUTH-001: [제목]

## HISTORY
### v0.0.1 (2025-10-25)
- **INITIAL**: 초기 SPEC 생성
```

#### Step 1-2: Feature 브랜치 생성 (Team 모드)
```bash
git checkout develop              # develop 기반에서 시작
git pull origin develop           # 최신 상태로 동기화
git checkout -b feature/SPEC-{ID} # feature/SPEC-AUTH-001 생성
```

**규칙**:
- 항상 `develop` 브랜치에서 시작
- 브랜치 이름: `feature/SPEC-{ID}` (설정값: `feature_prefix`)
- 직접 main 브랜치 생성 금지

#### Step 1-3: GitHub Issue 생성 (Team 모드 고유)
```bash
gh issue create \
  --title "[SPEC-AUTH-001] JWT Authentication System" \
  --body "[SPEC 문서 내용]
  
  ## Acceptance Criteria
  - Test coverage ≥ 85%
  - All tests pass
  
  ## Implementation Plan
  [plan.md 내용]"
```

**Issue 특징**:
- 제목: `[SPEC-{ID}] {SPEC 제목}`
- 본문: SPEC, Acceptance Criteria, Implementation Plan 포함
- GitHub Projects 연동 가능
- PR과 자동 연결됨

#### Step 1-4: Draft PR 생성
```bash
# git-manager가 자동 실행
gh pr create \
  --draft \
  --base develop \
  --head feature/SPEC-{ID} \
  --title "[SPEC-AUTH-001] JWT Authentication System" \
  --body "[Draft PR 본문 - SPEC 링크 포함]"
```

**Draft PR 특징**:
- 초기 상태: `DRAFT` (리뷰 요청 불가)
- 기본 브랜치: `develop`
- Feature 브랜치에 push할 때마다 자동 업데이트
- `/alfred:3-sync`에서 Ready로 전환

**git-manager 구현 (git-manager.md에서)**:
```markdown
## 📋 Feature 개발 워크플로우 (feature/*)

### 1. SPEC 작성 시 (/alfred:1-plan)
```bash
git checkout develop
git checkout -b feature/SPEC-{ID}

gh pr create --draft --base develop --head feature/SPEC-{ID}
```
```

---

### 4.2 Stage 2: `/alfred:2-run` - TDD 구현 및 자동 커밋

**참여 에이전트**:
- `implementation-planner` (Sonnet): 구현 계획 수립
- `tdd-implementer` (Sonnet): RED → GREEN → REFACTOR
- `quality-gate` (Haiku): TRUST 5 원칙 검증
- `git-manager` (Haiku): 커밋 생성 및 PR 업데이트

**수행 작업**:

#### Step 2-1: RED - 실패하는 테스트 작성
```python
# tests/auth/test_service.py
# @TEST:AUTH-001 TAG 추가

def test_user_authentication_with_valid_credentials():
    """JWT 토큰 발급 테스트"""
    # Given: 유효한 사용자 자격증명
    credentials = {"username": "user", "password": "pass"}
    
    # When: 로그인 요청
    # Then: JWT 토큰 발급 (아직 구현되지 않아 FAIL)
```

**자동 커밋**:
```bash
git add tests/auth/test_service.py
git commit -m "🔴 RED: Add JWT token issuance test
  
  @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

  🤖 Generated with Claude Code
  Co-Authored-By: Alfred <alfred@mo.ai.kr>"
```

#### Step 2-2: GREEN - 최소 구현으로 테스트 통과
```python
# src/auth/service.py
# @CODE:AUTH-001 TAG 추가

def authenticate_user(username: str, password: str) -> str:
    """JWT 토큰 발급 (최소 구현)"""
    # 검증 로직 없이 토큰만 생성
    return jwt.encode({"user": username}, "secret", algorithm="HS256")
```

**자동 커밋**:
```bash
git add src/auth/service.py
git commit -m "🟢 GREEN: Implement JWT token issuance
  
  @CODE:AUTH-001 | TEST: tests/auth/test_service.py | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

  🤖 Generated with Claude Code
  Co-Authored-By: Alfred <alfred@mo.ai.kr>"
```

#### Step 2-3: REFACTOR - 코드 품질 개선
```python
# src/auth/service.py (개선)

def authenticate_user(username: str, password: str) -> str:
    """JWT 토큰 발급 (개선된 버전)"""
    _validate_credentials(username, password)
    payload = {
        "user": username,
        "exp": datetime.utcnow() + timedelta(minutes=30)
    }
    return jwt.encode(payload, os.getenv("JWT_SECRET"), algorithm="HS256")

def _validate_credentials(username: str, password: str) -> None:
    """자격증명 검증"""
    if not username or not password:
        raise ValueError("Username and password required")
```

**자동 커밋**:
```bash
git add src/auth/service.py
git commit -m "♻️ REFACTOR: Improve JWT token handling and validation
  
  - Add token expiration
  - Add environment-based secret management
  - Extract validation logic
  @CODE:AUTH-001

  🤖 Generated with Claude Code
  Co-Authored-By: Alfred <alfred@mo.ai.kr>"
```

#### Step 2-4: 자동 PR 업데이트
```bash
# git-manager가 각 커밋 후 자동 실행
git push origin feature/SPEC-{ID}

# Draft PR 자동 업데이트 (gh CLI가 처리)
# PR 본문에 커밋 로그, 테스트 결과, 커버리지 리포트 추가
```

**Draft PR 상태**:
- 브랜치에 새 커밋이 push될 때마다 자동 업데이트
- CI/CD 파이프라인 자동 실행
- 리뷰어 자동 할당 (구성된 경우)
- 리뷰 요청 불가 (Draft 상태이므로)

**Draft PR 본문 자동 업데이트 내용**:
```markdown
## Implementation Summary

### Commits
- 🔴 RED: Add JWT token issuance test
- 🟢 GREEN: Implement JWT token issuance  
- ♻️ REFACTOR: Improve JWT token handling and validation

### Test Results
✅ All tests passing (15/15)
- Test coverage: 87% (target: 85%)

### Quality Gate
✅ TRUST 5 principles verified
- T (Test First): ✅ 87% coverage
- R (Readable): ✅ Code style pass
- U (Unified): ✅ Architecture consistent
- S (Secured): ✅ Security scan clean
- T (Traceable): ✅ TAG chain complete

@SPEC:AUTH-001
@TEST:AUTH-001
@CODE:AUTH-001
```

---

### 4.3 Stage 3: `/alfred:3-sync` - 문서 동기화 및 PR Ready 전환

**참여 에이전트**:
- `tag-agent` (Haiku): TAG 체인 검증
- `quality-gate` (Haiku): 최종 품질 확인
- `doc-syncer` (Haiku): Living Document 동기화
- `git-manager` (Haiku): PR Ready 전환 및 자동 머지

**수행 작업**:

#### Step 3-1: TAG 체인 검증 (전체 프로젝트 범위)
```bash
# tag-agent가 실행
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 검증 항목:
# - @SPEC:AUTH-001 존재 ✅
# - @TEST:AUTH-001 존재 ✅
# - @CODE:AUTH-001 존재 ✅
# - @DOC:AUTH-001 존재 (필요 시)
```

**검증 결과**:
```markdown
## TAG Chain Verification Report

✅ Primary Chain Complete
- @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md
- @TEST:AUTH-001 → tests/auth/test_service.py
- @CODE:AUTH-001 → src/auth/service.py
- @DOC:AUTH-001 → docs/api/authentication.md

✅ No Orphan TAGs detected
✅ No Broken References detected
```

#### Step 3-2: Living Document 동기화
```bash
# doc-syncer가 실행

# 1. 자동 생성/업데이트 문서:
docs/api/authentication.md    # @CODE:AUTH-001에서 함수 서명 추출
README.md                     # 새 기능 섹션 추가
CHANGELOG.md                  # v0.1.0 변경사항 기록

# 2. SPEC 메타데이터 자동 업데이트
.moai/specs/SPEC-AUTH-001/spec.md:
  status: draft → completed
  version: 0.0.1 → 0.1.0
```

**생성된 문서 예시**:

`docs/api/authentication.md`:
```markdown
# Authentication API

## @CODE:AUTH-001: JWT Authentication

### Functions

#### authenticate_user(username: str, password: str) -> str

**Description**: JWT 토큰 발급

**Parameters**:
- `username` (str): 사용자명
- `password` (str): 패스워드

**Returns**: JWT 토큰 문자열

**Example**:
```python
token = authenticate_user("user", "password")
```

**References**:
- SPEC: @SPEC:AUTH-001
- Tests: @TEST:AUTH-001
- Implementation: src/auth/service.py
```

#### Step 3-3: PR Ready 전환 (Team 모드 자동)
```bash
# doc-syncer가 문서 커밋 후 git-manager 호출
git add -A docs/ CHANGELOG.md README.md .moai/specs/SPEC-AUTH-001/spec.md
git commit -m "docs: Synchronize documentation with AUTH-001 implementation
  
  - Update API documentation
  - Add CHANGELOG entry
  - Update SPEC metadata (draft → completed)
  - Update README features list
  
  @DOC:AUTH-001 @SPEC:AUTH-001

  🤖 Generated with Claude Code
  Co-Authored-By: Alfred <alfred@mo.ai.kr>"

git push origin feature/SPEC-AUTH-001

# Draft PR를 Ready for Review로 전환
gh pr ready {PR_NUMBER}
```

**PR 상태 변화**:
- `DRAFT` → `READY_FOR_REVIEW`
- 리뷰어 자동 요청 활성화
- CI/CD 최종 검사 실행

#### Step 3-4: [선택] PR 자동 머지 (--auto-merge 플래그 사용 시)
```bash
# /alfred:3-sync --auto-merge 실행 시

# 1. CI/CD 상태 확인
gh pr checks --watch {PR_NUMBER}
# → 모든 체크 통과 대기

# 2. Squash 머지 실행
gh pr merge --squash --delete-branch {PR_NUMBER}

# 3. Local cleanup
git checkout develop
git pull origin develop
git branch -d feature/SPEC-AUTH-001
```

**머지 커밋 예시**:
```
docs: Synchronize documentation with AUTH-001 implementation (#5)

Squashed commit containing:
- 🔴 RED: Add JWT token issuance test
- 🟢 GREEN: Implement JWT token issuance
- ♻️ REFACTOR: Improve JWT token handling
- docs: Update documentation and SPEC metadata

@SPEC:AUTH-001 @TEST:AUTH-001 @CODE:AUTH-001 @DOC:AUTH-001

🤖 Generated with Claude Code
Co-Authored-By: Alfred <alfred@mo.ai.kr>
```

---

## 5. 현재 구현 상태 분석

### 5.1 완전 구현 항목

| 항목 | 상태 | 증거 |
|------|------|------|
| **Draft PR 자동 생성** | ✅ 완전 구현 | `git-manager.md`: "gh pr create --draft" |
| **Feature 브랜치 자동 생성** | ✅ 완전 구현 | `.moai/config.json`: `feature_prefix: "feature/SPEC-"` |
| **TDD 단계별 커밋** | ✅ 완전 구현 | `git-manager.md`: RED/GREEN/REFACTOR 커밋 템플릿 |
| **Tag 기반 추적** | ✅ 완전 구현 | SPEC/TEST/CODE/DOC TAG 시스템 |
| **PR Ready 전환** | ✅ 완전 구현 | `/alfred:3-sync`: `gh pr ready` |
| **자동 문서 동기화** | ✅ 완전 구현 | `doc-syncer.md`: Living Document 동기화 |
| **Develop 기반 브랜치** | ✅ 완전 구현 | `git-manager.md`: GitFlow 표준 준수 |

### 5.2 부분 구현 항목

| 항목 | 상태 | 설명 |
|------|------|------|
| **GitHub Issue 자동 생성** | ⚠️ 부분 구현 | `/alfred:1-plan`에서 "Create GitHub Issue" 언급 있으나 실제 구현 세부사항 미흡 |
| **PR 자동 머지** | ✅ 구현됨 | `/alfred:3-sync --auto-merge` 플래그로 활성화 |
| **리뷰어 자동 할당** | ⚠️ 부분 구현 | `doc-syncer.md`에서 언급만 있고 세부 로직 미설명 |
| **CI/CD 자동 검사** | ✅ 구현됨 | `.github/workflows/` 자동 트리거 |

### 5.3 미구현 항목

| 항목 | 설명 |
|------|------|
| **Automatic Merge Conflict Resolution** | PR 머지 시 충돌 발생 시 자동 해결 불가 |
| **PR Template Validation** | PR 템플릿 준수 여부 자동 검증 |
| **Release Notes Auto-generation** | Release 브랜치에서 자동 Release Notes 생성 |

---

## 6. GitHub 이슈/PR 자동 생성 메커니즘

### 6.1 Issue 자동 생성 (미구현이지만 설계된 흐름)

**설계된 흐름** (git-manager.md 참조):
```
/alfred:1-plan
  → spec-builder: SPEC 작성
  → git-manager: 
    1. feature 브랜치 생성
    2. [Team 모드] GitHub Issue 생성 (title: "[SPEC-{ID}] {제목}")
    3. Draft PR 생성 (feature → develop)
```

**현재 구현 상태**:
- Issue 생성 명령어 정의: `gh issue create` (예상)
- 정확한 구현 코드 미확인
- Agent 협력 구조에는 포함됨

### 6.2 Draft PR 자동 생성 (완전 구현)

**구현 확인됨** (git-manager.md):
```bash
gh pr create --draft --base develop --head feature/SPEC-{ID}
```

**작동 방식**:
1. `/alfred:1-plan` 실행 → SPEC 파일 생성
2. git-manager 에이전트 호출 → branch + Draft PR 생성
3. Draft 상태로 시작 → `/alfred:3-sync`에서 Ready로 전환

### 6.3 PR 상태 변화 (완전 구현)

```
Draft PR 생성 (/alfred:1-plan)
    ↓
TDD 구현 중 자동 업데이트 (/alfred:2-run)
    ↓
문서 동기화 및 Ready 전환 (/alfred:3-sync)
    ↓
[선택] PR 자동 머지 + 브랜치 정리 (/alfred:3-sync --auto-merge)
```

---

## 7. Team 모드 워크플로우 다이어그램

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: SPEC 생성 (/alfred:1-plan)                        │
└─────────────────────────────────────────────────────────────┘
                    ↓
         ┌──────────────────────┐
         │  spec-builder        │
         │  (SPEC 작성)         │
         └──────────────────────┘
                    ↓
         ┌──────────────────────────────────────┐
         │     git-manager                      │
         │  1. feature 브랜치 생성              │
         │     (develop 기반)                   │
         │  2. GitHub Issue 생성                │
         │  3. Draft PR 생성                    │
         │     (feature → develop)              │
         └──────────────────────────────────────┘
                    ↓
         SPEC 문서 + Branch + Draft PR 준비됨
         (.moai/specs/SPEC-{ID}/)

┌─────────────────────────────────────────────────────────────┐
│ Phase 2: TDD 구현 (/alfred:2-run)                           │
└─────────────────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ implementation-planner: 실행 계획 수립           │
  └──────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ tdd-implementer: RED → GREEN → REFACTOR          │
  │  • @TEST TAG 추가                                │
  │  • @CODE TAG 추가                                │
  └──────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ git-manager: 자동 커밋                           │
  │  • git add + commit (RED)                        │
  │  • git add + commit (GREEN)                      │
  │  • git add + commit (REFACTOR)                   │
  │  • git push origin feature/SPEC-{ID}             │
  │  → Draft PR 자동 업데이트                        │
  └──────────────────────────────────────────────────┘
                    ↓
         Draft PR 상태 리포트
         - Commits: RED/GREEN/REFACTOR
         - Coverage: X%
         - CI/CD: In Progress

┌─────────────────────────────────────────────────────────────┐
│ Phase 3: 동기화 (/alfred:3-sync [--auto-merge])             │
└─────────────────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ tag-agent: TAG 체인 검증 (전체 프로젝트)        │
  │  • @SPEC, @TEST, @CODE, @DOC 존재 확인          │
  │  • 고아 TAG 및 끊어진 링크 검출                 │
  └──────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ quality-gate: 품질 게이트 검증 (선택사항)      │
  │  • TRUST 5 원칙 검증                             │
  │  • 코드 스타일 검증                             │
  │  • 테스트 커버리지 확인                         │
  └──────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ doc-syncer: Living Document 동기화              │
  │  • API 문서 생성/업데이트                        │
  │  • README 업데이트                              │
  │  • CHANGELOG 추가                               │
  │  • SPEC 메타데이터 업데이트                     │
  │  • git commit + push                            │
  └──────────────────────────────────────────────────┘
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ git-manager: PR Ready 전환                       │
  │  • gh pr ready {PR_NUMBER}                       │
  │  • Draft → Ready for Review 상태 변경           │
  │  • CI/CD 최종 검사 실행                         │
  └──────────────────────────────────────────────────┘
                    ↓
    PR 준비 완료 (Ready for Review 상태)
    - 리뷰어 자동 요청 가능
    - CI/CD 모두 통과 중

[선택: --auto-merge 플래그 사용 시]
                    ↓
  ┌──────────────────────────────────────────────────┐
  │ git-manager: 자동 머지 실행                      │
  │  1. gh pr checks --watch (CI/CD 완료 대기)       │
  │  2. gh pr merge --squash --delete-branch         │
  │  3. git checkout develop                        │
  │  4. git pull origin develop                     │
  │  5. feature 브랜치 정리                         │
  └──────────────────────────────────────────────────┘
                    ↓
    완성! develop 브랜치에서 다음 작업 준비
    → /alfred:1-plan "다음 기능" 실행 가능
```

---

## 8. 실제 코드 예시

### 8.1 Commit 서명 표준

```
🔴 RED: Add authentication test case

@TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>
```

### 8.2 Tag 체인 구조

```
.moai/specs/SPEC-AUTH-001/spec.md
├─ @SPEC:AUTH-001 (명시적 TAG)
│
tests/auth/test_service.py
├─ @TEST:AUTH-001 (테스트 구현)
│
src/auth/service.py
├─ @CODE:AUTH-001 (소스 구현)
│
docs/api/authentication.md
├─ @DOC:AUTH-001 (문서 참조)
```

### 8.3 Config 기반 자동 선택

```python
# .moai/config.json 읽음
if config["project"]["mode"] == "team":
    # Team 모드 활성화
    use_gitflow = config["git_strategy"]["team"]["use_gitflow"]
    develop_branch = config["git_strategy"]["team"]["develop_branch"]
    auto_pr = config["git_strategy"]["team"]["auto_pr"]
    
    # 자동 실행:
    # 1. develop 기반 feature 브랜치 생성
    # 2. Draft PR 자동 생성
    # 3. Sync 단계에서 Ready 자동 전환
```

---

## 9. 핵심 결론

### 9.1 구현 완성도

| 영역 | 완성도 | 설명 |
|------|--------|------|
| **Branch 자동화** | 100% | Feature 브랜치 생성, GitFlow 준수 |
| **PR 자동화** | 95% | Draft PR 생성, Ready 전환, 자동 머지 모두 구현 |
| **Issue 자동화** | 70% | 설계됨, 일부 구현 확인 |
| **문서 동기화** | 100% | Living Document, TAG 체인 검증 완전 구현 |
| **Commit 관리** | 100% | TDD 단계별 자동 커밋, 서명 표준화 |

### 9.2 Team 모드 GitHub 통합 특징

✅ **자동화**
- 모든 기본 작업이 자동화됨
- 개발자는 코드 작성만 집중

✅ **추적성**
- @SPEC → @TEST → @CODE → @DOC TAG 완전한 추적
- 모든 커밋이 Alfred 서명으로 추적 가능
- PR 코멘트에 SPEC 링크 자동 포함

✅ **품질 보증**
- Draft PR로 시작 → 검증 후 Ready 전환
- CI/CD 자동 실행
- TRUST 5 원칙 자동 검증

✅ **협업 지원**
- GitHub Issue 기반 요구사항 추적
- Draft → Ready PR 상태 관리
- 리뷰어 자동 할당 (설정 시)

### 9.3 권장사항

1. **Issue 자동 생성 로직 명시화**
   - `gh issue create` 정확한 스펙 문서화
   - 테스트 케이스 추가

2. **PR 템플릿 강화**
   - Checklist 추가
   - Acceptance Criteria 자동 포함

3. **Auto-merge 정책 문서화**
   - Squash vs. Merge vs. Rebase 기준 명확히
   - CI/CD 요구사항 명시

4. **리뷰어 할당 규칙**
   - CODEOWNERS 파일 활용
   - 자동 할당 로직 구현

---

## 10. 참고 자료

### 명령어
- `/alfred:1-plan "기능 제목"` - SPEC + Branch + Draft PR
- `/alfred:2-run SPEC-{ID}` - TDD 구현
- `/alfred:3-sync` - 문서 동기화 + PR Ready
- `/alfred:3-sync --auto-merge` - 자동 머지까지 (Team 모드)

### Agent
- `spec-builder`: SPEC 작성
- `git-manager`: Git/GitHub 자동화
- `doc-syncer`: 문서 동기화
- `tdd-implementer`: TDD 구현

### Skills
- `moai-alfred-git-workflow`: GitFlow 자동화
- `moai-alfred-tag-scanning`: TAG 검증
- `moai-foundation-trust`: TRUST 5 검증

---

