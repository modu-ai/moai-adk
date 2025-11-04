# Git Workflow Best Practices - MoAI-ADK

**문서 언어**: 한국어
**대상**: MoAI-ADK 개발자 및 사용자
**작성일**: 2025-11-04
**버전**: 1.0

---

## 📋 목차

1. [SPEC-First 개발 워크플로우](#spec-first-개발-워크플로우)
2. [브랜치 분기 문제 예방](#브랜치-분기-문제-예방)
3. [Git 명령어 가이드](#git-명령어-가이드)
4. [문제 해결](#문제-해결)
5. [FAQ](#faq)

---

## 🎯 SPEC-First 개발 워크플로우

### 정상적인 워크플로우

```
┌─────────────────────────────────────────────────────────┐
│ Phase 0: 초기화 (/alfred:0-project)                    │
└──────────────────┬──────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────┐
│ Phase 1: 계획 (/alfred:1-plan)                         │
│ ✅ develop 최신 상태                                    │
│ ✅ 병합되지 않은 브랜치 확인                             │
│ ✅ feature/SPEC-XXX 브랜치 생성                         │
└──────────────────┬──────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────┐
│ Phase 2: 구현 (/alfred:2-run SPEC-XXX)                │
│ ✅ TDD 사이클 (RED → GREEN → REFACTOR)                 │
│ ✅ 자주 커밋                                             │
│ ✅ feature/SPEC-XXX 에서만 작업                         │
└──────────────────┬──────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────┐
│ Phase 3: 동기화 (/alfred:3-sync)                      │
│ ✅ PR 생성 (feature/SPEC-XXX → develop)               │
│ ✅ 코드 리뷰 받기                                       │
│ ✅ develop에 병합                                       │
└──────────────────┬──────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────┐
│ SPEC-XXX 완료! 다음 SPEC 시작                           │
│ 이 과정을 반복합니다.                                    │
└─────────────────────────────────────────────────────────┘
```

---

## ⚠️ 브랜치 분기 문제 예방

### 문제 상황

Issue #179에서 발생한 브랜치 분기 문제를 설명합니다.

#### 🔴 잘못된 워크플로우 (Issue #179 원인)

```
develop (최신)
    │
    ├─ SPEC-DOCS-005 브랜치 생성 & 작업 ✅
    │  └─ 889개 파일 추가 및 커밋
    │
    ✅ SPEC-DOCS-005 원격 push
    │
    ❌ develop 업데이트 없이 바로 SPEC-NAV-001 생성
    │
    └─ SPEC-NAV-001 브랜치 생성 (오래된 develop에서)
       └─ SPEC 문서만 3개 추가

결과: SPEC-NAV-001이 SPEC-DOCS-005의 889개 파일 접근 불가!
```

#### ✅ 올바른 워크플로우

```
develop (최신)
    │
    ├─ SPEC-DOCS-005 브랜치 생성 & 작업
    │  └─ 889개 파일 추가
    │
    ✅ develop에 병합 (PR 통과 후)
    │
    ✅ develop 다시 pull (최신 버전 획득)
    │
    └─ SPEC-NAV-001 브랜치 생성 (새로운 develop에서)
       └─ 889개 파일 + SPEC 문서로 작업 가능!

결과: SPEC-NAV-001이 SPEC-DOCS-005의 모든 파일 접근 가능! ✅
```

---

## 📋 Pre-Branch Creation 체크리스트

새로운 SPEC 브랜치를 생성하기 전에 **반드시** 확인하세요:

### Step 1: develop 최신화

```bash
# develop으로 전환
git checkout develop

# 최신 버전 가져오기
git pull origin develop

# 확인
git log -1 --oneline
```

✅ **확인**: 로컬 develop과 origin/develop이 동일한지 확인

### Step 2: 병합되지 않은 브랜치 확인

```bash
# 병합되지 않은 브랜치 목록
git branch --no-merged develop

# 예상 출력:
# (병합된 브랜치만 있으면 아무것도 안 보임)
# feature/SPEC-001
# feature/SPEC-002
```

⚠️ **주의**: 병합되지 않은 브랜치가 있으면, **먼저 병합하거나 삭제**해야 함

```bash
# 병합되지 않은 브랜치 병합하기
git checkout feature/SPEC-001
git merge develop  # 또는
git rebase develop

# 또는 삭제하기
git branch -D feature/SPEC-001
```

### Step 3: 현재 상태 확인

```bash
# 현재 브랜치 상태
git status

# 예상 출력:
# On branch develop
# Your branch is up to date with 'origin/develop'.
# nothing to commit, working tree clean
```

✅ **조건**:
- [ ] "On branch develop" 확인
- [ ] "up to date with 'origin/develop'" 확인
- [ ] "working tree clean" 확인

### Step 4: 파일 수 확인

```bash
# 현재 develop의 파일 수
git ls-files | wc -l

# 예상 출력: 1000+ (정확한 수는 버전에 따라 다름)
```

✅ **조건**: 파일 수가 충분히 많아야 함 (100개 이상)

---

## 🛠️ Git 명령어 가이드

### 1. SPEC 브랜치 안전하게 생성하기

```bash
# 안전한 브랜치 생성 (권장)
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-XXX  # XXX는 SPEC 번호

# 또는 alias 사용 (설정했다면)
git safebranch feature/SPEC-XXX
```

### 2. 현재 상황 파악하기

```bash
# 간결한 상태 보기
git status

# 브랜치 목록 (업스트림 정보 포함)
git branch -vv

# 브랜치 파일 수 비교
echo "Current: $(git ls-files | wc -l) files"
git ls-files | wc -l

# develop과의 파일 수 비교
echo "Develop: $(git ls-tree -r --name-only develop | wc -l) files"
```

### 3. 브랜치 동기화하기

```bash
# develop의 최신 변경사항 가져오기
git fetch origin develop

# develop과 비교하여 차이 확인
git log develop..HEAD --oneline

# develop 최신 버전 가져오기
git merge develop  # 또는
git rebase develop

# 동기화 확인
git log -1 --oneline  # develop 이후 커밋 보임
```

---

## 🔍 문제 해결

### 상황 1: "너무 많은 파일이 삭제되었습니다"

**증상**:
```bash
$ git status
On branch feature/SPEC-NAV-001

Changes not staged for commit:
  deleted:  docs-site/src/app/page.tsx
  deleted:  scripts/shopby_analyzer.py
  ... (889 files)
```

**원인**: 브랜치 분기 문제 또는 잘못된 체크아웃

**해결책**:

```bash
# 현재 브랜치의 base 확인
git merge-base --is-ancestor develop HEAD

# 현재 파일 수 확인
git ls-files | wc -l

# develop과의 파일 수 비교
git ls-tree -r --name-only develop | wc -l

# 차이가 크면, 브랜치 동기화 필요
git merge develop  # 안전한 방법
```

### 상황 2: "develop과 동기화하고 싶습니다"

```bash
# Option 1: Merge (안전함, 권장)
git checkout feature/SPEC-XXX
git merge develop
git push origin feature/SPEC-XXX

# Option 2: Rebase (깔끔함, force push 필요)
git checkout feature/SPEC-XXX
git rebase develop
git push origin feature/SPEC-XXX --force-with-lease
```

### 상황 3: "develop이 뒤처져 있습니다"

```bash
# develop 업데이트
git checkout develop
git pull origin develop

# 현재 브랜치도 업데이트
git checkout feature/SPEC-XXX
git merge develop
```

### 상황 4: "실수로 파일을 삭제했습니다"

```bash
# 변경사항 취소 (커밋 전)
git checkout -- <file_path>  # 또는
git restore <file_path>

# 커밋된 경우
git revert <commit_hash>
```

---

## ❓ FAQ

### Q1: `/alfred:1-plan` 명령어는 항상 develop에서 분기하나요?

**A**: 네, `/alfred:1-plan`은 **항상** develop에서 분기하도록 설계되었습니다.

수동으로 브랜치를 생성하면 이 보장이 없으므로, **항상 `/alfred:1-plan` 사용**하세요.

---

### Q2: 여러 SPEC을 동시에 작업할 수 있나요?

**A**: 가능하지만 신중해야 합니다.

**안전한 방법**:
```
develop
  ├─ feature/SPEC-001 (병합 후)
  │  ├─ merge to develop ✅
  │  └─ git pull origin develop
  │
  └─ feature/SPEC-002 (새로운 develop에서)
      └─ 안전 ✅
```

**위험한 방법**:
```
develop
  ├─ feature/SPEC-001 (병합 안 함!)
  │
  └─ feature/SPEC-002 (분기 생성)
      └─ SPEC-001의 파일에 접근 불가 ❌
```

---

### Q3: 작은 파일만 변경했는데 왜 889개 파일이 "삭제"되었나요?

**A**: Git의 diff 알고리즘 때문입니다.

브랜치를 전환할 때, 두 브랜치의 **파일 목록 차이**를 계산합니다:

```
feature/SPEC-NAV-001에만 있는 파일: 삭제된 것처럼 표시
feature/SPEC-DOCS-005에만 있는 파일: 추가된 것처럼 표시

실제로는: 브랜치에 파일이 없을 뿐 삭제된 것 아님!
```

---

### Q4: 현재 브랜치의 파일 수를 어떻게 확인하나요?

**A**: 다음 명령어로 확인할 수 있습니다.

```bash
# 현재 브랜치의 파일 수
git ls-files | wc -l

# develop의 파일 수
git ls-tree -r --name-only develop | wc -l

# 차이 계산
CURRENT=$(git ls-files | wc -l)
DEVELOP=$(git ls-tree -r --name-only develop | wc -l)
DIFF=$((CURRENT - DEVELOP))
echo "파일 수 차이: $DIFF"
```

---

### Q5: 실수로 오래된 커밋에서 브랜치를 생성했습니다. 어떻게 하나요?

**A**: 다음 중 하나 선택:

**Option 1: Rebase (권장)**
```bash
git checkout <your_branch>
git rebase develop
git push origin <your_branch> --force-with-lease
```

**Option 2: Reset & Re-branch**
```bash
git branch -D <your_branch>  # 기존 브랜치 삭제
git checkout develop
git pull origin develop
git checkout -b <your_branch>
```

---

### Q6: git 명령어가 무섭습니다. 어떻게 하나요?

**A**: 안전한 도구들을 사용하세요:

1. **GUI 도구**: VSCode, GitKraken, SourceTree
2. **Alfred 명령어**: `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
3. **Alias**: 자주 사용하는 명령어 단축

**추천**: 가능하면 **Alfred 명령어만 사용**하세요. 안전합니다!

---

## 🛡️ Git Hooks 설정 (선택)

브랜치 분기 문제를 미리 감지하려면 Git Hook을 설정할 수 있습니다.

### Pre-checkout Hook 설정

`.git/hooks/pre-checkout` 파일 생성:

```bash
#!/bin/bash

# MoAI-ADK Branch Divergence Warning

TARGET_BRANCH="$1"
CURRENT_FILE_COUNT=$(git ls-files 2>/dev/null | wc -l)
TARGET_FILE_COUNT=$(git ls-tree -r --name-only "$TARGET_BRANCH" 2>/dev/null | wc -l)

if [ $? -eq 0 ]; then
    DIFF=$((CURRENT_FILE_COUNT - TARGET_FILE_COUNT))

    if [ ${DIFF#-} -gt 100 ]; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "⚠️  WARNING: Significant file count difference detected"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "  Current branch: $CURRENT_FILE_COUNT files"
        echo "  Target branch:  $TARGET_FILE_COUNT files"
        echo "  Difference:     ${DIFF#-} files"
        echo ""
        echo "This may indicate a branch divergence issue."
        echo "Make sure this is expected before proceeding."
        echo ""
        read -p "Continue with checkout? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Checkout cancelled."
            exit 1
        fi
    fi
fi

exit 0
```

실행 권한 부여:

```bash
chmod +x .git/hooks/pre-checkout
```

---

## 📞 추가 지원

문제가 발생하면:

1. **GitHub Issue 검색**: 유사한 문제가 이미 보고되었을 수 있음
2. **GitHub Issue 생성**: 새로운 문제는 상세히 설명하고 보고
3. **Alfred 도움**: `/alfred:2-run` 중 오류 발생 시 자동으로 제안

---

## 📚 참고 자료

- **Issue #179**: [브랜치 분기 문제 분석](https://github.com/modu-ai/moai-adk/issues/179)
- **CLAUDE.md**: [MoAI-ADK 프로젝트 지침](./CLAUDE.md)
- **Alfred 가이드**: [Alfred 워크플로우](./alfred-workflow.md)

---

**문서 버전**: 1.0
**마지막 업데이트**: 2025-11-04
**작성자**: Alfred (MoAI-ADK SuperAgent)

🤖 Generated with Claude Code
Co-Authored-By: 🎩 Alfred@MoAI
