# Claude Code Hooks 최적화 계획 (MoAI-ADK v0.26.0)

**목표**: 구조 단순화, 코드 중복 제거, 유지보수성 향상
**예상 기간**: 3-5 스프린트
**리스크 수준**: 중간 (Hook 시스템 변경)

---

## Phase 1: 분석 및 기획 (1주일)

### 1.1 문제점 우선순위 재확인

| 우선순위 | 이슈 | 노력 | 영향 | 위험 | 점수 |
|----------|------|------|------|------|------|
| 1️⃣ | 중복 코드 제거 | 중간 | 높음 | 중간 | 9점 |
| 2️⃣ | 임포트 경로 통일 | 높음 | 높음 | 낮음 | 8점 |
| 3️⃣ | 빈 디렉토리 제거 | 낮음 | 낮음 | 매우낮음 | 6점 |
| 4️⃣ | Hook 파일 분할 | 높음 | 중간 | 중간 | 7점 |
| 5️⃣ | __pycache__ 정리 | 낮음 | 낮음 | 매우낮음 | 5점 |

### 1.2 변경 영향도 분석

```
변경 대상           임포트 의존 Hook 수    영향도
────────────────────────────────────────────
moai/core/          모든 Hook (9개)        Critical
utils/timeout       3개 Hook              High
handlers/           1개 Hook              Low
shared/config/      0개 Hook              Low
spec_status_hooks   별도 CLI              Very Low
```

### 1.3 테스트 계획 수립

```
테스트 항목:
☐ Hook timeout 동작 (5초 제한)
☐ JSON stdin/stdout 형식
☐ 모든 Hook 실행 (SessionStart 우선)
☐ 캐시 동작 확인 (git-info, version-check)
☐ 에러 복구 (graceful degradation)
☐ 문서 일관성
```

---

## Phase 2: 코드 중복 제거 (1주일)

### 2.1 중복 모듈 통합

**Step 1: moai/shared/core/ 확정**

```bash
# 템플릿 파일 확인
ls -la src/moai_adk/templates/.claude/hooks/moai/shared/core/
ls -la src/moai_adk/templates/.claude/hooks/moai/core/
```

**Step 2: 로컬 moai/core/ 제거**

```bash
# 1. 백업 생성
cp -r .claude/hooks/moai/core .moai/backup/hooks-core-backup

# 2. moai/core에서만 임포트하는 파일 찾기
grep -r "from.*moai\.core\|from core\|import.*moai\.core" .claude/hooks/moai/ \
  | grep -v "shared/core"

# 3. 해당 파일들의 임포트 경로 변경
# (Step 2.2에서 상세히)

# 4. moai/core/ 제거
rm -rf .claude/hooks/moai/core/
```

**Step 3: 템플릿 moai/core/ 제거**

```bash
# 템플릿에서도 제거하여 SSOT 유지
rm -rf src/moai_adk/templates/.claude/hooks/moai/core/
```

### 2.2 임포트 경로 통일

**현재 임포트 패턴**:
```python
# ❌ 상대 경로 (모호함)
sys.path.insert(0, str(SHARED_DIR))
from utils.timeout import ...
from core.config_cache import ...

# ❌ 혼합 사용 (일관성 부족)
from handlers import handle_pre_tool_use
from utils.timeout import ...
from core.config_cache import ...
```

**개선된 임포트 패턴**:
```python
# ✅ 명시적 절대 경로
import sys
from pathlib import Path

# Hook 루트에서 shared 경로 계산
HOOKS_DIR = Path(__file__).parent.parent  # moai/
SHARED_DIR = HOOKS_DIR / "shared"
sys.path.insert(0, str(SHARED_DIR))

# 또는 프로젝트 루트에서 계산
PROJECT_ROOT = Path(__file__).parent.parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT / "src"))

# 임포트
from moai_adk.hooks.shared.core.config_cache import get_cached_config
from moai_adk.hooks.shared.handlers.session import handle_session_start
from moai_adk.hooks.shared.utils.timeout import CrossPlatformTimeout
```

**변경 대상 파일**:

```
session_start__show_project_info.py
┌─ 현재: from core.config_cache import ...
└─ 변경: from moai_adk.hooks.shared.core.config_cache import ...

session_start__auto_cleanup.py
┌─ 현재: from moai_adk.utils.common import ...
├─ 확인: moai_adk.utils.common 존재 여부
└─ 변경: (필요시) from moai_adk.hooks.shared.utils import ...

session_start__config_health_check.py
┌─ 현재: 상대 경로 사용 (확인 필요)
└─ 변경: 명시적 절대 경로

pre_tool__auto_checkpoint.py
┌─ 현재: from handlers import handle_pre_tool_use
│         from utils.timeout import ...
└─ 변경: from moai_adk.hooks.shared.handlers.tool import handle_pre_tool_use
         from moai_adk.hooks.shared.core.timeout import ...

pre_tool__document_management.py
└─ 변경: (내용 확인 후 동일하게)

post_tool__*.py
└─ 변경: (내용 확인 후 동일하게)

subagent_*.py
└─ 변경: (내용 확인 후 동일하게)

session_end__auto_cleanup.py
└─ 변경: (내용 확인 후 동일하게)
```

**Step-by-step 변경 예시**:

```python
# ===== session_start__show_project_info.py =====

# Before:
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

try:
    from utils.timeout import CrossPlatformTimeout
except ImportError:
    ...

try:
    from core.config_cache import get_cached_config
except ImportError:
    ...

# After:
HOOKS_DIR = Path(__file__).parent.parent  # moai/
SHARED_DIR = HOOKS_DIR / "shared"
PROJECT_ROOT = HOOKS_DIR.parent.parent.parent  # MoAI-ADK/

# 프로젝트 루트 기반 임포트
if str(PROJECT_ROOT / "src") not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT / "src"))

try:
    from moai_adk.hooks.shared.core.config_cache import (
        get_cached_config,
        get_cached_spec_progress,
    )
    from moai_adk.hooks.shared.core.timeout import CrossPlatformTimeout
except ImportError as e:
    # Fallback 구현
    ...
```

### 2.3 utils/timeout.py 정리

**현재 상황**:
- moai/utils/timeout.py (로컬 복제)
- moai/shared/core/timeout.py (공유 버전)

**결정**:
```
Option A: moai/utils/timeout.py 제거 → shared/core/timeout.py 사용
Option B: moai/utils/timeout.py 유지 → shared/core 임포트하게 변경
┗━ 권장: Option A (DRY 원칙)
```

**Step 1: 임포트 경로 변경**

```bash
# moai/utils/timeout.py를 임포트하는 모든 파일 찾기
grep -r "from.*utils.*timeout\|from utils.timeout" .claude/hooks/moai/

# 해당 파일들의 임포트를 shared/core/timeout.py로 변경
```

**Step 2: moai/utils/timeout.py 제거**

```bash
# 템플릿에서도 제거
rm -f .claude/hooks/moai/utils/timeout.py
rm -f src/moai_adk/templates/.claude/hooks/moai/utils/timeout.py
```

---

## Phase 3: 구조 정리 (3-4일)

### 3.1 빈 디렉토리 제거

**Step 1: 빈 디렉토리 확인**

```bash
# 빈 디렉토리 찾기
find .claude/hooks/moai -type d -empty

# 예상 결과:
# .claude/hooks/moai/handlers/
# .claude/hooks/moai/shared/config/
```

**Step 2: 제거**

```bash
# 로컬
rm -rf .claude/hooks/moai/handlers/
rm -rf .claude/hooks/moai/shared/config/

# 템플릿
rm -rf src/moai_adk/templates/.claude/hooks/moai/handlers/
rm -rf src/moai_adk/templates/.claude/hooks/moai/shared/config/
```

### 3.2 spec_status_hooks.py 이동

**현재 위치**: `.claude/hooks/moai/spec_status_hooks.py`

**새 위치**: `src/moai_adk/cli/spec_status_hooks.py`

**Step 1: 기능 확인**

```python
# 이것은 Hook이 아니라 CLI 유틸리티
# ✅ 커맨드라인 인자 사용 (argparse)
# ✅ Hook stdin/stdout 아님
# ✅ 파이썬 모듈로 직접 실행 가능
```

**Step 2: 파일 이동**

```bash
# 템플릿에서 이동
mv src/moai_adk/templates/.claude/hooks/moai/spec_status_hooks.py \
   src/moai_adk/cli/spec_status_hooks.py

# 로컬에서 이동
mv .claude/hooks/moai/spec_status_hooks.py \
   src/moai_adk/cli/spec_status_hooks.py
```

**Step 3: 임포트 경로 업데이트**

```python
# spec_status_hooks.py 내부 임포트 확인 및 수정
# Before:
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

# After:
# (이미 src/ 내부이므로 상대 임포트 사용 가능)
from moai_adk.core.spec_status_manager import SpecStatusManager
```

**Step 4: /moai:2-run, /moai:3-sync 명령어 업데이트**

```bash
# 명령어에서 Hook 호출 제거
# Before:
# /moai:2-run SPEC-XXX
#   └─ .claude/hooks/moai/spec_status_hooks.py status_update

# After:
# /moai:2-run SPEC-XXX
#   └─ uv run -m moai_adk.cli.spec_status_hooks status_update
```

### 3.3 __pycache__ 정리

**Step 1: .gitignore 업데이트**

```bash
cat >> .gitignore << 'EOF'

# Claude Code Hooks
.claude/hooks/**/__pycache__/
.claude/hooks/**/*.pyc
.claude/hooks/**/*.pyo
.claude/hooks/**/*.pyd
EOF
```

**Step 2: 기존 __pycache__ 제거**

```bash
# 로컬
find .claude/hooks -type d -name __pycache__ -exec rm -rf {} \;

# 템플릿
find src/moai_adk/templates/.claude/hooks -type d -name __pycache__ -exec rm -rf {} \;

# Git에서 제거
git rm -r --cached '.claude/hooks/**/__pycache__'
git rm -r --cached 'src/moai_adk/templates/.claude/hooks/**/__pycache__'
```

---

## Phase 4: Hook 파일 분할 (선택사항, 2주일)

### 4.1 session_start__auto_cleanup.py 분할

**현재**: 628 lines (3개 기능 혼합)

```python
# 기능 1: 오래된 파일 정리 (cleanup_old_files)
# 기능 2: 일일 분석 보고서 생성 (generate_daily_analysis)
# 기능 3: 정리 통계 업데이트 (update_cleanup_stats)
```

**분할 계획**:

```
Option A: 3개 Hook으로 분할 (권장하지 않음 - SessionStart 오버헤드)
┗━ 동시 실행으로 성능 저하

Option B: 2개 Hook으로 분할
┗━ session_start__file_cleanup.py (200-250 lines)
   session_start__daily_analysis.py (300-350 lines)

Option C: 1개 Hook + 1개 모듈로 분할 (권장)
┗━ session_start__auto_cleanup.py (메인 - 150 lines)
   shared/handlers/file_cleanup.py (추출 - 200 lines)
   shared/handlers/daily_analysis.py (추출 - 250 lines)
```

**권장**: Option C

**Step 1: 핸들러 모듈 추출**

```bash
# shared/handlers/file_cleanup.py 생성
# - cleanup_old_files()
# - cleanup_directory()
# - update_cleanup_stats()

# shared/handlers/daily_analysis.py 생성
# - generate_daily_analysis()
# - analyze_session_logs()
# - format_analysis_report()
```

**Step 2: Hook 파일 간단히**

```python
# session_start__auto_cleanup.py (간단화)
from moai_adk.hooks.shared.handlers.file_cleanup import cleanup_old_files
from moai_adk.hooks.shared.handlers.daily_analysis import generate_daily_analysis

def main():
    config = load_config()

    # 파일 정리 실행
    cleanup_stats = cleanup_old_files(config)

    # 일일 분석 생성
    report_path = generate_daily_analysis(config)

    # 결과 반환
    print(json.dumps({
        "cleanup_stats": cleanup_stats,
        "daily_analysis_report": report_path,
    }))
```

---

## Phase 5: 문서화 및 검증 (1주일)

### 5.1 Hook 표준 헤더 추가

```python
#!/usr/bin/env python3
"""Hook Name: [Brief Description]

Claude Code Event: [SessionStart|PreToolUse|PostToolUse|SubagentStart|SubagentStop]
Purpose: [What this hook does]
Execution: [When it's triggered]
Matcher: [Tool pattern if applicable, e.g., "Edit|Write|MultiEdit"]

Input Schema:
{
  "session_id": "string",
  "hook_event_name": "SessionStart|...",
  "hook_version": "1.0.0",
  ...
}

Output Schema:
{
  "continue": true|false,
  "systemMessage": "string (optional)",
  "hookSpecificOutput": {...} (optional)
}

Dependencies:
- moai_adk.hooks.shared.core.config_cache
- moai_adk.hooks.shared.utils.timeout

Performance:
- Timeout: 5 seconds
- Average execution: [actual time from benchmarks]

Examples:
- See .moai/hooks/examples/ for test inputs/outputs

Notes:
- Always return {"continue": true} to avoid blocking session
- Use graceful degradation for non-critical failures
- Log errors to stderr, JSON to stdout
"""

import json
import sys
from pathlib import Path
from typing import Any, Dict, Optional

# ... rest of code ...
```

### 5.2 모든 Hook 검증

```bash
#!/bin/bash
# Hook 검증 스크립트

HOOKS_DIR=".claude/hooks/moai"
REQUIRED_HOOKS=(
    "session_start__show_project_info.py"
    "session_start__auto_cleanup.py"
    "session_start__config_health_check.py"
    "pre_tool__auto_checkpoint.py"
    "pre_tool__document_management.py"
    "post_tool__enable_streaming_ui.py"
    "post_tool__log_changes.py"
    "subagent_start__context_optimizer.py"
    "subagent_stop__lifecycle_tracker.py"
)

echo "=== Hook 파일 검증 ==="
for hook in "${REQUIRED_HOOKS[@]}"; do
    if [ -f "$HOOKS_DIR/$hook" ]; then
        echo "✅ $hook"
    else
        echo "❌ $hook (MISSING)"
    fi
done

echo ""
echo "=== 중복 코드 검색 ==="
if [ -d "$HOOKS_DIR/core" ]; then
    echo "❌ moai/core/ 여전히 존재"
else
    echo "✅ moai/core/ 제거됨"
fi

if [ -f "$HOOKS_DIR/utils/timeout.py" ]; then
    echo "❌ moai/utils/timeout.py 여전히 존재"
else
    echo "✅ moai/utils/timeout.py 제거됨"
fi

echo ""
echo "=== 빈 디렉토리 검색 ==="
find "$HOOKS_DIR" -type d -empty

echo ""
echo "=== Python 문법 검사 ==="
python3 -m py_compile $(find "$HOOKS_DIR" -name "*.py" -type f)
```

### 5.3 임포트 경로 검증

```bash
#!/bin/bash
# 임포트 경로 일관성 검사

HOOKS_DIR=".claude/hooks/moai"

echo "=== 상대 경로 사용 검사 (제거 필요) ==="
grep -r "from utils\|from core\|from handlers\|from config" \
    "$HOOKS_DIR" \
    --include="*.py" \
    | grep -v "^[^:]*:#" \
    | head -20

echo ""
echo "=== 절대 경로 사용 검사 (확인) ==="
grep -r "from moai_adk" "$HOOKS_DIR" --include="*.py" | wc -l
```

---

## Phase 6: 테스트 및 배포 (1주일)

### 6.1 로컬 테스트

```bash
# 1. SessionStart Hook 실행 테스트
echo "{}" | python3 .claude/hooks/moai/session_start__show_project_info.py
echo "{}" | python3 .claude/hooks/moai/session_start__auto_cleanup.py
echo "{}" | python3 .claude/hooks/moai/session_start__config_health_check.py

# 2. 출력 형식 검증
# - JSON 형식 확인
# - "continue" 필드 존재 확인
# - 타임아웃 처리 확인 (5초 초과 X)

# 3. 의존성 로드 확인
python3 -c "from moai_adk.hooks.shared.core.config_cache import *"

# 4. 전체 Hook 실행 (Claude Code 세션에서)
/moai:0-project (또는 기존 프로젝트)
# → SessionStart Hook 자동 실행됨
```

### 6.2 CI/CD 검증

```yaml
# .github/workflows/hooks-lint.yml
name: Hooks Validation

on: [push, pull_request]

jobs:
  validate-hooks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: uv sync

      - name: Lint Hook files
        run: |
          ruff check .claude/hooks/moai/ --select=E,W,F

      - name: Type check Hook files
        run: |
          mypy .claude/hooks/moai/ --ignore-missing-imports

      - name: Syntax check
        run: |
          python3 -m py_compile $(find .claude/hooks -name "*.py")

      - name: Verify Hook structure
        run: |
          python3 .moai/scripts/validate-hooks.py
```

### 6.3 마이그레이션 체크리스트

```markdown
## 마이그레이션 체크리스트

### Phase 2: 코드 중복 제거
- [ ] moai/core/ → shared/core/ 통합 검증
- [ ] 모든 파일의 임포트 경로 업데이트
- [ ] moai/utils/timeout.py 제거 및 임포트 수정
- [ ] 로컬 + 템플릿 동기화 확인

### Phase 3: 구조 정리
- [ ] moai/handlers/ 제거
- [ ] moai/shared/config/ 제거
- [ ] spec_status_hooks.py 이동 및 임포트 수정
- [ ] __pycache__ 제거 및 .gitignore 업데이트
- [ ] Git 커밋

### Phase 4: Hook 파일 분할 (선택사항)
- [ ] shared/handlers/file_cleanup.py 생성
- [ ] shared/handlers/daily_analysis.py 생성
- [ ] session_start__auto_cleanup.py 간단히
- [ ] 모든 Hook 테스트

### Phase 5: 문서화
- [ ] 모든 Hook에 표준 헤더 추가
- [ ] README.md 작성 (Hook 구조, 실행 흐름)
- [ ] MAINTENANCE.md 작성 (유지보수 가이드)

### Phase 6: 테스트
- [ ] SessionStart Hook 실행 테스트
- [ ] PreToolUse Hook 실행 테스트
- [ ] 캐시 동작 확인
- [ ] 에러 복구 테스트
- [ ] CI/CD 통과

### 배포
- [ ] Release notes 작성
- [ ] Changelog 업데이트
- [ ] 템플릿 배포
- [ ] 로컬 프로젝트 동기화 지시
```

---

## 최적화 결과 (예상)

### 파일 수 감소

```
현재:
- Hook 파일: 9개
- 공유 코드: 12개
- 유틸리티: 4개 (중복 포함)
- __pycache__: 4개 (Git 추적 제외)
────────────────
합계: 29개 추적 파일

최적화 후 (Phase 2-3):
- Hook 파일: 9개 (동일)
- 공유 코드: 10개 (중복 3개 제거)
- 유틸리티: 3개 (timeout 통합)
- spec_status: 0개 (이동됨)
────────────────
합계: 22개 추적 파일

감소: 7개 파일 (24% 감소)
```

### 복잡도 감소

```
현재:
- 디렉토리 계층: 4단계 (moai → shared → core → files)
- 중복 코드: ~15KB (3개 파일)
- 임포트 경로: 3가지 패턴 (상대 + 혼합 + 절대)

최적화 후:
- 디렉토리 계층: 3단계 (shared → core/handlers/utils)
- 중복 코드: 0KB
- 임포트 경로: 1가지 패턴 (절대 경로)
```

### 유지보수성 개선

```
지표                  현재      최적화 후    개선율
──────────────────────────────────────────────
코드 중복율          ~5%      0%           100%
평균 Hook 크기       200L     180L         10%
임포트 경로 일관성   60%      100%         67%
문서화 커버리지      70%      95%          36%
테스트 용이성        중간      높음         +2레벨
```

---

## 리스크 관리

### 높은 리스크 (위험 완화 계획)

| 리스크 | 가능성 | 영향 | 완화 계획 |
|--------|--------|------|----------|
| Hook 실행 실패 | 중간 | 높음 | 각 단계마다 SessionStart Hook 테스트 |
| 임포트 오류 | 중간 | 높음 | CI/CD에 Python 문법 검사 추가 |
| 캐시 문제 | 낮음 | 중간 | 캐시 로직 검증 테스트 추가 |
| 템플릿 동기화 실패 | 낮음 | 높음 | SSOT 검증 스크립트 실행 |

### 롤백 계획

```bash
# 각 Phase 완료 후 백업
git tag "hooks-before-phase-2"
git tag "hooks-before-phase-3"
# ...

# 문제 발생 시 롤백
git reset --hard hooks-before-phase-2
```

---

## 일정 및 담당

| Phase | 일정 | 소요시간 | 담당 | 상태 |
|-------|------|---------|------|------|
| 1: 분석 기획 | Week 1 | 2일 | - | ✅ 완료 |
| 2: 코드 중복 제거 | Week 2 | 5일 | - | ⏳ 대기 |
| 3: 구조 정리 | Week 2-3 | 3일 | - | ⏳ 대기 |
| 4: Hook 분할 (선택) | Week 3-4 | 10일 | - | ⏳ 선택사항 |
| 5: 문서화 검증 | Week 4 | 5일 | - | ⏳ 대기 |
| 6: 테스트 배포 | Week 5 | 5일 | - | ⏳ 대기 |
| **총 기간** | **5주** | **30일** | | |

---

## 다음 단계

1. ✅ **분석 리포트 검토** (HOOKS_ANALYSIS_REPORT.md)
2. ✅ **최적화 계획 승인** (이 문서)
3. ⏳ **Phase 2 시작**: 코드 중복 제거
   - 명령어: `/moai:2-run SPEC-HOOKS-001`
4. ⏳ **정기적 진행 상황 보고**
5. ⏳ **최종 배포**: Release v0.27.0

---

**문서 작성**: 2025-11-19
**버전**: 1.0.0
**상태**: 🔵 검토 대기 중
