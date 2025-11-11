---
title: "Claude Code Statusline 통합"
description: "Claude Code 터미널 상태 표시줄에 실시간 개발 상황 표시"
---

# Claude Code Statusline 통합

MoAI-ADK statusline은 Claude Code 터미널 상태 표시줄에 **실시간 개발 상황**을 표시합니다. 모델, 버전, Git 브랜치, 파일 변경 상황을 한눈에 파악할 수 있습니다.

## 📊 상태줄 포맷

### 기본 형식

```
🤖 Haiku 4.5 | 🗿 Ver 0.20.1 | 📊 Git: develop | Changes: +0 M0 ?0
```

| 항목 | 아이콘 | 의미 | 예시 |
|------|------|------|------|
| **모델** | 🤖 | 사용 중인 Claude 모델 | Haiku 4.5, Sonnet 4.5 |
| **버전** | 🗿 | MoAI-ADK 버전 | 0.20.1 |
| **Git 브랜치** | 📊 | 현재 작업 중인 브랜치 | develop, feature/SPEC-001 |
| **Changes** | - | Git 파일 변경 상태 | +0 M0 ?0 |

## 📝 Changes 표기 설명

Changes 필드는 Git 작업 상태를 세 부분으로 표시합니다:

```
Changes: +staged Mmodified ?untracked

+0  = 스테이징된 파일 개수 (git add된 파일)
M0  = 수정된 파일 개수 (git add 안 된 파일)
?0  = 추적되지 않는 새 파일 개수
```

### 상태별 의미

| 상태 | 표시 | 의미 |
|------|------|------|
| **정상** | `Changes: +0 M0 ?0` | 모든 변경사항 commit됨 |
| **수정됨** | `Changes: +0 M2 ?0` | 2개 파일 수정됨 (git add 필요) |
| **새 파일** | `Changes: +0 M0 ?1` | 새 파일 1개 (git add 필요) |
| **준비됨** | `Changes: +3 M0 ?0` | 3개 파일 준비됨 (commit 가능) |
| **진행 중** | `Changes: +2 M1 ?1` | 복합 상태: staged + modified + untracked |

## 🎯 렌더링 모드

### 1. Compact 모드 (기본)

길이: 80자 이내

```
🤖 Haiku 4.5 | 🗿 Ver 0.20.1 | 📊 Git: develop | Changes: +0 M0 ?0
```

**특징**:
- 가장 일반적인 정보만 표시
- 빠른 상태 확인에 최적화
- 터미널 공간 효율적

### 2. Extended 모드

길이: 120자 이내

```
🤖 Haiku 4.5 | 🗿 Ver 0.20.1 | 📊 Git: feature/SPEC-001 | Changes: +2 M1 ?0 | [PLAN]
```

**특징**:
- 현재 작업 단계 표시 ([PLAN], [RUN], [SYNC])
- 긴 브랜치 이름 완전 표시
- 상세한 Git 상태 정보

### 3. Minimal 모드

길이: 40자 이내

```
🤖 H 4.5 | 🗿 Ver 0.20.1
```

**특징**:
- 최소한의 정보만 표시
- 화면 공간이 제한된 환경에 적합
- 핵심 정보만 빠르게 확인

## ⚙️ 설정 방법

### 1. 설정 파일 편집

`.claude/settings.json` 파일에서 statusline 설정:

```json
{
  "statusLine": {
    "type": "command",
    "command": "python3 -m moai_adk.statusline.main",
    "padding": 1
  }
}
```

### 2. 환경 변수 설정

```bash
# Compact 모드 (기본)
export MOAI_STATUSLINE_MODE=compact

# Extended 모드
export MOAI_STATUSLINE_MODE=extended

# Minimal 모드
export MOAI_STATUSLINE_MODE=minimal
```

### 3. 동적 모드 전환

```bash
# 세션 중 모드 변경
moai-adk statusline --mode extended

# 현재 설정 확인
moai-adk statusline --show-config
```

## 🔧 상태 표시줄 구현

### Python 모듈 구조

```python
# src/moai_adk/statusline/main.py
class StatusLineRenderer:
    def __init__(self, mode="compact"):
        self.mode = mode
        self.git_info = GitInfo()
        self.moai_info = MoAIInfo()

    def render(self) -> str:
        """상태줄 렌더링"""
        if self.mode == "compact":
            return self.render_compact()
        elif self.mode == "extended":
            return self.render_extended()
        elif self.mode == "minimal":
            return self.render_minimal()

    def render_compact(self) -> str:
        """Compact 모드 렌더링"""
        model = self.get_claude_model()
        version = self.get_moai_version()
        branch = self.git_info.get_current_branch()
        changes = self.git_info.get_file_changes()

        return f"🤖 {model} | 🗿 Ver {version} | 📊 Git: {branch} | Changes: {changes}"
```

### Git 정보 추출

```python
class GitInfo:
    def get_current_branch(self) -> str:
        """현재 Git 브랜치 가져오기"""
        try:
            result = subprocess.run(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"],
                capture_output=True, text=True
            )
            return result.stdout.strip()
        except subprocess.CalledProcessError:
            return "unknown"

    def get_file_changes(self) -> str:
        """파일 변경 상태 계산"""
        try:
            # Staged files
            staged = subprocess.run(
                ["git", "diff", "--cached", "--name-only"],
                capture_output=True, text=True
            )
            staged_count = len(staged.stdout.strip().split('\n')) if staged.stdout.strip() else 0

            # Modified files
            modified = subprocess.run(
                ["git", "diff", "--name-only"],
                capture_output=True, text=True
            )
            modified_count = len(modified.stdout.strip().split('\n')) if modified.stdout.strip() else 0

            # Untracked files
            untracked = subprocess.run(
                ["git", "ls-files", "--others", "--exclude-standard"],
                capture_output=True, text=True
            )
            untracked_count = len(untracked.stdout.strip().split('\n')) if untracked.stdout.strip() else 0

            return f"+{staged_count} M{modified_count} ?{untracked_count}"
        except subprocess.CalledProcessError:
            return "+0 M0 ?0"
```

## 📊 실시간 정보 표시

### 모델 정보

```python
def get_claude_model(self) -> str:
    """현재 Claude 모델 정보 가져오기"""
    # Claude API 또는 환경 변수에서 모델 정보 추출
    model = os.environ.get("CLAUDE_MODEL", "Haiku")
    version = os.environ.get("CLAUDE_VERSION", "4.5")
    return f"{model} {version}"
```

### MoAI-ADK 버전 정보

```python
def get_moai_version(self) -> str:
    """MoAI-ADK 버전 정보 가져오기"""
    try:
        from moai_adk import __version__
        return __version__
    except ImportError:
        return "unknown"
```

### 파이프라인 상태

```python
def get_pipeline_status(self) -> str:
    """현재 MoAI-ADK 파이프라인 상태"""
    config = self.load_moai_config()
    stage = config.get("pipeline", {}).get("current_stage", "unknown")

    stage_mapping = {
        "initialized": "[INIT]",
        "planning": "[PLAN]",
        "implementation": "[RUN]",
        "sync": "[SYNC]"
    }

    return stage_mapping.get(stage, "[UNKNOWN]")
```

## 🎨 커스터마이징

### 1. 색상 테마 설정

```json
{
  "statusLine": {
    "theme": {
      "model": "blue",
      "version": "green",
      "git": "yellow",
      "changes": "red"
    }
  }
}
```

### 2. 사용자 정의 포맷

```json
{
  "statusLine": {
    "custom_format": "🚀 {model} | {version} | {branch} | {changes}",
    "show_timestamp": true,
    "show_workspace": true
  }
}
```

### 3. 필터 설정

```json
{
  "statusLine": {
    "filters": {
      "hide_empty_changes": true,
      "shorten_branch_names": true,
      "show_only_dirty": false
    }
  }
}
```

## 🔍 디버깅 및 문제 해결

### 상태 표시줄이 표시되지 않을 때

```bash
# 상태 표시줄 모듈 직접 실행
python3 -m moai_adk.statusline.main

# Git 저장소 확인
git status

# Claude Code 설정 확인
cat ~/.claude/settings.json
```

### 일반적인 문제

| 문제 | 원인 | 해결책 |
|------|------|--------|
| **Git 정보 없음** | Git 저장소 아님 | `git init` 실행 |
| **모델 정보 없음** | 환경 변수 누락 | `CLAUDE_MODEL` 설정 |
| **버전 정보 없음** | MoAI-ADK 설치 안됨 | `pip install moai-adk` |
| **권한 오류** | 실행 권한 없음 | `chmod +x statusline.py` |

## 📈 성능 최적화

### 캐싱 전략

```python
class CachedStatusInfo:
    def __init__(self, cache_ttl=30):
        self.cache = {}
        self.cache_ttl = cache_ttl

    def get_git_branch(self) -> str:
        """캐시된 Git 브랜치 정보"""
        if "branch" not in self.cache or self.is_cache_expired("branch"):
            self.cache["branch"] = self.fetch_git_branch()
            self.cache["branch_time"] = time.time()

        return self.cache["branch"]
```

### 비동기 업데이트

```python
import asyncio

async def update_status_async():
    """비동기 상태 정보 업데이트"""
    tasks = [
        fetch_git_info(),
        fetch_model_info(),
        fetch_pipeline_status()
    ]

    results = await asyncio.gather(*tasks)
    return format_statusline(*results)
```

## 🎯 사용 팁

### 1. 워크플로우 최적화

```bash
# 개발 시작 전 상태 확인
echo "Current status: $(moai-adk statusline)"

# 커밋 전 변경사항 확인
git status
echo "Statusline: $(moai-adk statusline --mode extended)"
```

### 2. 자동화 통합

```bash
# Pre-commit hook
#!/bin/sh
# .git/hooks/pre-commit

echo "Pre-commit status: $(moai-adk statusline)"
if moai-adk status-check; then
    exit 0
else
    echo "❌ Status check failed"
    exit 1
fi
```

### 3. 팀 협업

```bash
# 팀 표준 설정 공유
echo "export MOAI_STATUSLINE_MODE=extended" >> team-setup.sh

# 브랜치 명명 규칙 강제
git config --global moai.statusline.branch-prefix "feature/SPEC-"
```

## 📋 향후 개선 계획

- [ ] **다중 모델 지원**: 동시에 여러 Claude 모델 표시
- [ ] **성능 메트릭**: 실시간 성능 지표 추가
- [ ] **알림 통합**: 중요 이벤트 알림 표시
- [ ] **팀 상태**: 팀원들의 작업 상태 공유
- [ ] **프로젝트 통계**: 프로젝트 진행률 시각화

Statusline 통합을 통해 MoAI-ADK는 개발자에게 실시간 프로젝트 상태를 직관적으로 제공하여 더 효율적인 개발 경험을 제공합니다.