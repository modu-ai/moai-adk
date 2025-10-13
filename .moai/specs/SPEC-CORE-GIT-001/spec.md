---
id: CORE-GIT-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high
category: feature
labels:
  - git
  - gitpython
  - version-control
depends_on:
  - PY314-001
scope:
  packages:
    - moai-adk-py/src/moai_adk/core/git/
  files:
    - git/manager.py
    - git/branch.py
    - git/commit.py
---

# @SPEC:CORE-GIT-001: GitPython 기반 Git 관리

## HISTORY

### v0.0.1 (2025-10-13)
- **INITIAL**: GitPython 기반 Git 워크플로우 관리 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: TypeScript simple-git를 GitPython으로 전환

---

## 개요

GitPython 라이브러리를 사용하여 Git 브랜치, 커밋, PR 관리를 자동화한다. SPEC-First TDD 워크플로우에 필요한 Git 작업을 추상화한다.

---

## Environment (환경 및 전제조건)

### 기술 스택
- **Git 라이브러리**: GitPython 3.1+
- **Git 클라이언트**: git 2.30+
- **GitHub CLI**: gh (PR 생성용)

### 기존 시스템
- TypeScript simple-git 기반
- feature/SPEC-XXX 브랜치 자동 생성
- TDD 단계별 커밋 (RED, GREEN, REFACTOR)
- Draft PR 자동 생성

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 Git 저장소 조작 API를 제공해야 한다
- 시스템은 브랜치 생성/전환을 지원해야 한다
- 시스템은 커밋 및 푸시를 지원해야 한다
- 시스템은 현재 브랜치 상태를 조회할 수 있어야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 브랜치 생성 요청이 오면, 시스템은 feature/SPEC-XXX 형식으로 생성해야 한다
- WHEN TDD 커밋이 요청되면, 시스템은 단계별 이모지를 포함해야 한다
- WHEN 푸시가 요청되면, 시스템은 원격 저장소에 동기화해야 한다
- WHEN Draft PR 생성이 요청되면, 시스템은 gh CLI를 호출해야 한다

### State-driven Requirements (상태 기반)
- WHILE Git 저장소가 더티(dirty) 상태일 때, 시스템은 경고를 표시해야 한다
- WHILE 브랜치가 원격보다 뒤처져 있을 때, 시스템은 풀(pull) 권장해야 한다

### Constraints (제약사항)
- 브랜치명은 `feature/SPEC-XXX` 형식이어야 한다
- 커밋 메시지는 locale 설정에 따라 한국어/영어로 작성되어야 한다
- 모든 Git 작업은 사용자 확인 후 실행되어야 한다 (autoCommit 제외)

---

## Specifications (상세 명세)

### 1. GitManager 클래스

```python
# moai_adk/core/git/manager.py
from git import Repo
from pathlib import Path

class GitManager:
    def __init__(self, repo_path: str = "."):
        self.repo = Repo(repo_path)
        self.git = self.repo.git

    def is_repo(self) -> bool:
        """Git 저장소 여부 확인"""
        try:
            _ = self.repo.git_dir
            return True
        except:
            return False

    def current_branch(self) -> str:
        """현재 브랜치명 반환"""
        return self.repo.active_branch.name

    def is_dirty(self) -> bool:
        """작업 디렉토리 변경사항 확인"""
        return self.repo.is_dirty()

    def create_branch(self, branch_name: str, from_branch: str = "develop") -> None:
        """새 브랜치 생성 및 전환"""
        self.git.checkout("-b", branch_name, from_branch)

    def commit(self, message: str, files: list[str] = None) -> None:
        """파일 스테이징 및 커밋"""
        if files:
            self.repo.index.add(files)
        else:
            self.repo.git.add(A=True)

        self.repo.index.commit(message)

    def push(self, branch: str = None, set_upstream: bool = False) -> None:
        """원격 저장소에 푸시"""
        if set_upstream:
            self.git.push("--set-upstream", "origin", branch or self.current_branch())
        else:
            self.git.push()
```

### 2. 브랜치 네이밍

```python
def generate_branch_name(spec_id: str) -> str:
    """SPEC ID로부터 브랜치명 생성"""
    return f"feature/SPEC-{spec_id}"

# 예시
generate_branch_name("AUTH-001")  # => "feature/SPEC-AUTH-001"
```

### 3. TDD 커밋 메시지 (Locale 기반)

```python
def format_commit_message(stage: str, description: str, locale: str = "ko") -> str:
    """TDD 단계별 커밋 메시지 생성"""
    templates = {
        "ko": {
            "red": "🔴 RED: {desc}",
            "green": "🟢 GREEN: {desc}",
            "refactor": "♻️ REFACTOR: {desc}",
            "docs": "📝 DOCS: {desc}",
        },
        "en": {
            "red": "🔴 RED: {desc}",
            "green": "🟢 GREEN: {desc}",
            "refactor": "♻️ REFACTOR: {desc}",
            "docs": "📝 DOCS: {desc}",
        }
    }

    template = templates.get(locale, templates["en"]).get(stage.lower())
    return template.format(desc=description)
```

### 4. Draft PR 생성 (gh CLI)

```python
import subprocess

def create_draft_pr(
    title: str,
    body: str,
    base: str = "develop",
    head: str = None
) -> str:
    """GitHub Draft PR 생성"""
    cmd = [
        "gh", "pr", "create",
        "--title", title,
        "--body", body,
        "--base", base,
        "--draft"
    ]

    if head:
        cmd.extend(["--head", head])

    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    return result.stdout.strip()  # PR URL 반환
```

### 5. Git 상태 조회

```python
def get_repo_status(manager: GitManager) -> dict:
    """저장소 상태 정보 반환"""
    return {
        "is_repo": manager.is_repo(),
        "current_branch": manager.current_branch(),
        "is_dirty": manager.is_dirty(),
        "untracked_files": manager.repo.untracked_files,
        "modified_files": [item.a_path for item in manager.repo.index.diff(None)],
    }
```

---

## Traceability (추적성)

- **SPEC ID**: @SPEC:CORE-GIT-001
- **Depends on**: PY314-001
- **TAG 체인**: @SPEC:CORE-GIT-001 → @TEST:CORE-GIT-001 → @CODE:CORE-GIT-001
