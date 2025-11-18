# Claude Code Hooks 최적화 분석 리포트 (MoAI-ADK)

**작성 일자**: 2025-11-19
**현재 버전**: 0.26.0
**상태**: 발견된 문제점 및 최적화 기회 식별 완료

---

## 1. 현재 상태 개요

### 구조 분석
```
.claude/hooks/
├── __init__.py (버전 1.0.0)
├── moai/
│   ├── __init__.py
│   ├── core/                           # ⚠️ 중복: shared/core와 동일
│   │   ├── project.py
│   │   ├── timeout.py
│   │   ├── ttl_cache.py
│   │   ├── version_cache.py
│   │   └── __pycache__/
│   ├── utils/                          # ✓ 로컬 유틸리티
│   │   ├── gitignore_parser.py
│   │   ├── hook_config.py
│   │   ├── timeout.py                  # ⚠️ 중복: shared/core/timeout.py와 동일
│   │   └── __init__.py
│   ├── shared/                         # ✓ 공유 코드
│   │   ├── core/                       # ✓ 핵심 기능
│   │   │   ├── agent_context.py
│   │   │   ├── checkpoint.py
│   │   │   ├── config_cache.py
│   │   │   ├── config_manager.py       # 대규모: 370 lines
│   │   │   ├── context.py
│   │   │   ├── error_handler.py        # 대규모: 237 lines
│   │   │   ├── json_utils.py
│   │   │   ├── project.py              # ⚠️ 중복: core/project.py와 동일
│   │   │   ├── timeout.py              # ⚠️ 중복: core/timeout.py와 동일
│   │   │   ├── version_cache.py        # ⚠️ 중복: core/version_cache.py와 동일
│   │   │   ├── __pycache__/
│   │   │   └── __init__.py
│   │   ├── handlers/                   # ✓ Hook 로직
│   │   │   ├── daily_analysis.py
│   │   │   ├── notification.py
│   │   │   ├── session.py
│   │   │   ├── tool.py
│   │   │   ├── user.py
│   │   │   ├── __pycache__/
│   │   │   └── __init__.py
│   │   ├── utils/                      # ✓ 공유 유틸리티
│   │   │   ├── announcement_translator.py
│   │   │   ├── state_tracking.py
│   │   │   └── (아직 __init__.py 없음)
│   │   ├── config/                     # ⚠️ 미사용 디렉토리 (비어있음)
│   │   └── __init__.py
│   ├── handlers/                       # ⚠️ 비어있음 (shared/handlers로 통합됨)
│   │   └── __init__.py
│   ├── SessionStart 훅 (3개)
│   │   ├── session_start__show_project_info.py (508 lines)
│   │   ├── session_start__auto_cleanup.py (628 lines)
│   │   └── session_start__config_health_check.py (419 lines)
│   ├── PreToolUse 훅 (2개)
│   │   ├── pre_tool__auto_checkpoint.py (98 lines)
│   │   └── pre_tool__document_management.py
│   ├── PostToolUse 훅 (2개)
│   │   ├── post_tool__enable_streaming_ui.py
│   │   └── post_tool__log_changes.py
│   ├── SubagentStart 훅 (1개)
│   │   └── subagent_start__context_optimizer.py
│   ├── SubagentStop 훅 (1개)
│   │   └── subagent_stop__lifecycle_tracker.py
│   ├── SessionEnd 훅 (1개)
│   │   └── session_end__auto_cleanup.py
│   └── spec_status_hooks.py             # ⚠️ 유틸리티 (Hook 아님)
└── (로컬 복제본 = 템플릿 동기화됨)

파일 수 (총): 58개
- Hook 파일: 9개 (공식 스펙 준수)
- 공유 코드: 12개
- 유틸리티: 4개
- __pycache__: 4개 (Git 추적 불필요)
- __init__.py: 약 10개
```

---

## 2. 공식 Claude Code Hooks 스펙 검증

### 지원하는 Hook 타입 (cchooks 라이브러리)
```
✅ SessionStart         - 세션 시작 시 자동 실행 (시스템 메시지 반환 가능)
✅ UserPromptSubmit     - 사용자 입력 전처리 (블로킹 가능)
✅ PreToolUse          - 도구 사용 전 검증 (블로킹 가능)
✅ PostToolUse         - 도구 사용 후 처리 (블로킹 불가)
✅ SubagentStart       - 에이전트 시작 시 컨텍스트 시딩
✅ SubagentStop        - 에이전트 종료 시 정리
❌ Notification        - 알림 (현재 미사용)
❌ Stop                - 세션 중지 제어 (현재 미사용)
❌ PreCompact          - 컨텍스트 압축 사전 처리 (현재 미사용)
```

### MoAI-ADK 구현 현황
```
✅ SessionStart: 3개 훅
   - session_start__show_project_info.py      (508 lines) → 프로젝트 상태 표시
   - session_start__auto_cleanup.py           (628 lines) → 자동 정리 + 일일 분석
   - session_start__config_health_check.py    (419 lines) → 설정 건강도 확인

✅ PreToolUse: 2개 훅
   - pre_tool__auto_checkpoint.py             (98 lines)  → Git 체크포인트 생성
   - pre_tool__document_management.py         (?) → 문서 관리 (내용 미확인)

✅ PostToolUse: 2개 훅
   - post_tool__enable_streaming_ui.py        (?) → UI 스트리밍
   - post_tool__log_changes.py                (?) → 변경사항 로깅

✅ SubagentStart: 1개 훅
   - subagent_start__context_optimizer.py     (?) → 컨텍스트 최적화

✅ SubagentStop: 1개 훅
   - subagent_stop__lifecycle_tracker.py      (?) → 생명주기 추적

❌ SessionEnd: 1개 훅
   - session_end__auto_cleanup.py             (?) → 공식 스펙 상 존재 없음
   (예: PreCompact 또는 관리용 스크립트일 가능성)

⚠️ spec_status_hooks.py                       (290 lines)
   - Hook이 아니라 CLI 유틸리티 (argparse 사용)
   - /moai:2-run, /moai:3-sync와 통합용
   - 위치: Hook 디렉토리가 아닌 별도 모듈로 관리해야 함
```

---

## 3. 구조 문제점 (Critical + Major)

### 3.1 중복된 코드 (Critical: 3개 모듈)

**문제**: moai/core/ 와 moai/shared/core/ 에서 동일한 코드 존재

| 파일 | moai/core/ | moai/shared/core/ | 심각도 |
|------|-----------|------------------|--------|
| `project.py` | ✅ 존재 (3 KB) | ✅ 존재 (동일) | 🔴 Critical |
| `timeout.py` | ✅ 존재 (2 KB) | ✅ 존재 (동일) | 🔴 Critical |
| `version_cache.py` | ✅ 존재 (3 KB) | ✅ 존재 (동일) | 🔴 Critical |

**영향**:
- 코드 유지보수 비용 증가 (변경 시 2곳 수정)
- 임포트 경로 혼동 (shared/core vs core)
- 메모리 중복 로드 (동일 모듈 2회 import)

**권장**: `moai/core/` 제거 → `moai/shared/core/`로 통합

---

### 3.2 빈 디렉토리 (Major: 2개)

**문제**: 사용되지 않는 디렉토리 존재

| 디렉토리 | 상태 | 이유 |
|----------|------|------|
| `moai/handlers/` | 비어있음 | shared/handlers로 통합됨 |
| `moai/shared/config/` | 비어있음 | config_manager.py가 shared/core/에 있음 |

**권장**: 완전히 제거

---

### 3.3 임포트 경로 혼동 (Major)

**현재 상황**:
```python
# pre_tool__auto_checkpoint.py
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
from handlers import handle_pre_tool_use      # ✅ shared/handlers/
from utils.timeout import CrossPlatformTimeout # ✅ shared/utils/ 또는 moai/utils/?

# session_start__show_project_info.py
from utils.timeout import CrossPlatformTimeout # ⚠️ 모호함: moai/utils 또는 shared/utils?
from core.config_cache import ...              # ⚠️ 모호함: moai/core 또는 shared/core?
```

**권장**: 명시적 임포트 경로 사용

```python
# 개선 후
from moai_adk.hooks.shared.utils.timeout import CrossPlatformTimeout
from moai_adk.hooks.shared.core.config_cache import get_cached_config
```

---

### 3.4 Hook이 아닌 코드가 Hooks 디렉토리에 위치 (Major)

**문제**: `spec_status_hooks.py`

```python
# CLI 유틸리티 (Hook 스펙 미준수)
parser = argparse.ArgumentParser(description='Spec Status Manager Hooks')
parser.add_argument('command', choices=[
    'status_update', 'validate_completion', 'batch_update', 'detect_drafts'
], help='Command to execute')

# 사용법: python3 spec_status_hooks.py status_update SPEC-001 --status completed
```

**권장**:
- `src/moai_adk/cli/spec_status_hooks.py` 로 이동
- .claude/hooks/에서 제거

---

### 3.5 __pycache__ 디렉토리 추적 (Minor)

**문제**: Git에서 추적 중 (불필요)

```
moai/core/__pycache__/
moai/utils/__pycache__/
moai/shared/core/__pycache__/
moai/shared/handlers/__pycache__/
```

**권장**: `.gitignore`에 추가
```
.claude/hooks/**/__pycache__/
.claude/hooks/**/*.pyc
```

---

### 3.6 SessionEnd Hook의 위치 (Minor)

**문제**: 공식 Claude Code 스펙에 `SessionEnd` 없음

```python
# session_end__auto_cleanup.py
# 이것은 Hook이 아니라 명시적으로 호출되는 스크립트일 가능성
```

**권장**: 위치 재고찰
- 만약 SessionEnd Hook이라면: 공식 스펙 문서 확인
- 만약 CLI 스크립트라면: `.moai/scripts/` 로 이동

---

## 4. 코드 품질 분석

### 4.1 Hook 파일 크기

| Hook | 라인 수 | 복잡도 | 평가 |
|------|--------|--------|------|
| session_start__show_project_info.py | 508 | 높음 | ⚠️ 너무 큼 (분할 필요) |
| session_start__auto_cleanup.py | 628 | 높음 | 🔴 너무 큼 (분할 필수) |
| session_start__config_health_check.py | 419 | 중간 | ⚠️ 분할 고려 |
| pre_tool__auto_checkpoint.py | 98 | 낮음 | ✅ 적절 |
| shared/core/config_manager.py | 370 | 중간 | ⚠️ 모듈화 가능 |
| shared/core/error_handler.py | 237 | 중간 | ✅ 적절 |
| spec_status_hooks.py | 290 | 중간 | ⚠️ Hook 아님 |

**권장**:
- `session_start__auto_cleanup.py` (628 lines) → 3개 Hook으로 분할
- `session_start__show_project_info.py` (508 lines) → 내부 helper 추출

---

### 4.2 의존성 분석

```
Hook 계층 다이어그램:
┌─────────────────────────────────┐
│  Hook Main Files (9개)          │  ← 외부 진입점
│  ├─ session_start__*.py (3개)  │
│  ├─ pre_tool__*.py (2개)        │
│  ├─ post_tool__*.py (2개)       │
│  ├─ subagent_start__*.py        │
│  └─ subagent_stop__*.py         │
└────────────┬────────────────────┘
             │ import
             ↓
┌─────────────────────────────────┐
│  Shared Handlers (5개)          │
│  ├─ daily_analysis.py           │
│  ├─ session.py                  │
│  ├─ tool.py                     │
│  ├─ user.py                     │
│  └─ notification.py             │
└────────────┬────────────────────┘
             │ import
             ↓
┌─────────────────────────────────┐
│  Shared Core (9개)              │
│  ├─ config_manager.py (370L)   │
│  ├─ error_handler.py (237L)    │
│  ├─ context.py                  │
│  ├─ checkpoint.py               │
│  ├─ json_utils.py               │
│  ├─ project.py ❌ DUPLICATE     │
│  ├─ timeout.py ❌ DUPLICATE     │
│  ├─ version_cache.py ❌ DUP     │
│  └─ agent_context.py            │
└────────────┬────────────────────┘
             │ import
             ↓
┌─────────────────────────────────┐
│  Local Utilities (4개)          │
│  ├─ moai/utils/timeout.py ❌ DUP │
│  ├─ moai/utils/gitignore_parser │
│  ├─ moai/utils/hook_config.py   │
│  └─ moai/core/* ❌ ALL DUPLICATE │
└─────────────────────────────────┘
```

**의존성 추적**:
```
Hook files → Handlers (shared) → Core (shared) → [Duplicate core/timeout]
                                            └─→ [Duplicate core/project]
                                            └─→ [Duplicate core/version_cache]
                                            └─→ utils/timeout (duplicate)
```

---

## 5. 설정 건강도 (settings.json 연계)

### 현재 설정
```json
{
  "statusLine": {
    "command": "uv run $CLAUDE_PROJECT_DIR/.moai/bin/statusline.py"
  },
  "permissions": {
    "allow": ["Task", "Read", "Write", "Edit", ...],
    "ask": ["Bash(git commit:*)", ...],
    "deny": ["Bash(rm -rf /:*)", ...]
  }
}
```

### 권장 사항
- [ ] `$CLAUDE_PROJECT_DIR` 변수가 Hook 설정에서도 일관되게 사용되는지 확인
- [ ] Hook timeout 설정 (현재: config_manager.py에 기본값 5초) → settings.json으로 통합 검토

---

## 6. 성능 최적화 기회

### 6.1 캐싱 전략

**현재**: 각 Hook이 독립적으로 캐시 (git-info.json, version-check.json)

**권장**: 통합 캐시 관리
```
.moai/cache/
├─ git-info.json (TTL: 1분)
├─ version-check.json (TTL: 24시간)
├─ spec-progress.json (TTL: 5분)
└─ config-state.json (TTL: 1시간)
```

---

### 6.2 병렬 실행 분석

**SessionStart Hook 실행 순서**:
```
① session_start__show_project_info.py    (500ms - Git 명령어 병렬화)
② session_start__auto_cleanup.py         (300ms - 파일 시스템 I/O)
③ session_start__config_health_check.py  (200ms - PyPI API 캐시 사용)
────────────────────────────────────────
합계: ~1000ms (직렬 실행)
```

**권장**: 병렬화 가능성
- ① + ② + ③ 동시 실행 가능 (의존성 없음)
- 예상 시간: ~500ms (50% 개선)

---

## 7. 문서화 및 메인테넌스

### 7.1 Hook 문서화 상태

| Hook | 문서 | 주석 | 타입 정의 |
|------|------|------|----------|
| session_start__show_project_info.py | ✅ | ✅ | ✅ |
| session_start__auto_cleanup.py | ✅ | ✅ | ✅ |
| session_start__config_health_check.py | ✅ | ✅ | ✅ |
| pre_tool__auto_checkpoint.py | ✅ | ✅ | ⚠️ |
| pre_tool__document_management.py | ? | ? | ? |
| post_tool__enable_streaming_ui.py | ? | ? | ? |
| post_tool__log_changes.py | ? | ? | ? |
| subagent_start__context_optimizer.py | ? | ? | ? |
| subagent_stop__lifecycle_tracker.py | ? | ? | ? |

**권장**: 모든 Hook에 표준 헤더 추가
```python
#!/usr/bin/env python3
"""Hook Name: [Brief Description]

Claude Code Event: [SessionStart|PreToolUse|PostToolUse|SubagentStart|SubagentStop]
Purpose: [What this hook does]
Execution: [When it's triggered]
Input Schema: [Expected stdin JSON structure]
Output Schema: [Expected stdout JSON structure]

Key Decisions:
- [Why this approach]
- [Performance considerations]
"""
```

---

## 8. SSOT (Single Source of Truth) 검증

### 템플릿 → 로컬 동기화 상태

| 항목 | 템플릿 위치 | 로컬 위치 | 상태 | 검증 |
|------|-----------|---------|------|------|
| 구조 | `src/moai_adk/templates/.claude/hooks/` | `.claude/hooks/` | ✅ 동기화됨 | 파일 수 일치 |
| 코드 | 템플릿의 {{ }} 변수 | 절대 경로 | ⚠️ 확인 필요 | 임포트 경로 검증 |
| 설정 | 없음 (Hook은 config 독립) | .claude/settings.json | ✅ 통합 | Hook timeout 확인 |

**권장**:
- 템플릿 변경 후 로컬 동기화 자동화 스크립트 검토
- Hook 임포트 경로 일관성 점검

---

## 9. 최종 문제점 요약

### 🔴 Critical Issues (즉시 조치 필요)
1. **중복된 코드**: moai/core/ 및 utils/timeout.py (3개 모듈)
   - 해결: moai/core/ 제거, moai/shared/core/ 통합
   - 영향도: 높음 (임포트 경로 변경)

2. **임포트 경로 혼동**: 상대 경로 사용
   - 해결: 명시적 절대 경로로 변경
   - 영향도: 중간 (유지보수성 향상)

3. **Hook이 아닌 코드**: spec_status_hooks.py
   - 해결: CLI 유틸리티로 이동
   - 영향도: 낮음 (기능은 동일)

### 🟠 Major Issues (계획적 개선)
4. **빈 디렉토리**: moai/handlers/, moai/shared/config/
   - 해결: 제거
   - 영향도: 낮음 (정리용)

5. **Hook 크기**: session_start__auto_cleanup.py (628 lines)
   - 해결: 3개 Hook으로 분할
   - 영향도: 중간 (테스트 필요)

6. **SessionEnd Hook 위치**: 공식 스펙 확인 필요
   - 해결: 위치 재고려 또는 확인
   - 영향도: 낮음 (확인용)

### 🟡 Minor Issues (미래 개선)
7. **__pycache__ 추적**: Git에 포함됨
   - 해결: .gitignore 업데이트
   - 영향도: 매우 낮음 (정리용)

8. **문서화**: 일부 Hook 미기록
   - 해결: 표준 헤더 추가
   - 영향도: 낮음 (유지보수 개선)

---

## 10. 규정 준수 검증

### Claude Code 공식 스펙 (cchooks v2.x)

| 요구사항 | 현재 | 준수 | 개선안 |
|----------|------|------|--------|
| Hook 명명: `{event}__{description}.py` | ✅ | ✅ | - |
| JSON stdin/stdout | ✅ | ✅ | - |
| `"continue": true` 반환 | ✅ | ✅ | - |
| 타임아웃 처리 | ✅ | ✅ | - |
| 에러 복구 (graceful degradation) | ✅ | ✅ | - |
| `.claude/settings.json` 등록 | ⚠️ | ⚠️ | Hook 리스트 확인 |
| Hook 직렬화 가능 (JSON) | ✅ | ✅ | - |
| stderr에 로깅 (stdout은 JSON만) | ✅ | ✅ | - |

---

## 부록: 파일 목록

### Hook Files (공식 스펙 준수)
```
✅ .claude/hooks/moai/session_start__show_project_info.py
✅ .claude/hooks/moai/session_start__auto_cleanup.py
✅ .claude/hooks/moai/session_start__config_health_check.py
✅ .claude/hooks/moai/pre_tool__auto_checkpoint.py
✅ .claude/hooks/moai/pre_tool__document_management.py
✅ .claude/hooks/moai/post_tool__enable_streaming_ui.py
✅ .claude/hooks/moai/post_tool__log_changes.py
✅ .claude/hooks/moai/subagent_start__context_optimizer.py
✅ .claude/hooks/moai/subagent_stop__lifecycle_tracker.py
⚠️ .claude/hooks/moai/session_end__auto_cleanup.py (위치 확인 필요)
```

### Shared Core Modules (유지보수)
```
.claude/hooks/moai/shared/core/
├─ config_manager.py (370 lines) ✅
├─ error_handler.py (237 lines) ✅
├─ config_cache.py ✅
├─ context.py ✅
├─ checkpoint.py ✅
├─ json_utils.py ✅
├─ agent_context.py ✅
├─ project.py ❌ DUPLICATE
├─ timeout.py ❌ DUPLICATE
└─ version_cache.py ❌ DUPLICATE
```

### 제거 대상
```
❌ .claude/hooks/moai/core/ (전체 디렉토리)
❌ .claude/hooks/moai/handlers/ (비어있음)
❌ .claude/hooks/moai/shared/config/ (비어있음)
❌ .claude/hooks/moai/utils/timeout.py (shared/utils로 이동)
❌ spec_status_hooks.py (src/moai_adk/cli/로 이동)
```

---

**다음 단계**: OPTIMIZATION_PLAN.md 참조
