# MoAI-ADK Claude Code Hook 시스템 구조 분석 및 개선 계획

**분석 일시**: 2025-11-04
**분석 대상**: MoAI-ADK Hook 시스템 아키텍처
**목적**: 패키지 템플릿과 로컬 Hook 구조 정규화 및 동기화 전략 수립

---

## 📋 Executive Summary

### 핵심 발견사항

1. **구조 불일치**: 로컬에만 존재하는 `session_start__daily_analysis.py` Hook 및 `shared/handlers/daily_analysis.py` 핸들러가 패키지 템플릿에 누락됨
2. **의존성 분리**: `session_analyzer.py`가 `.moai/scripts/`에 위치하여 Hook 시스템과 물리적으로 분리됨
3. **경로 의존성**: Hook이 `.moai/scripts/session_analyzer.py`를 subprocess로 호출하는 구조 (결합도 높음)
4. **스크립트 정체성 모호**: `.moai/scripts/`가 Hook 전용인지 프로젝트 공용 유틸리티인지 명확하지 않음

### 권장사항 요약

- **Phase 1**: 패키지 템플릿에 누락된 daily_analysis 관련 파일 추가
- **Phase 2**: `.moai/scripts/` 구조 재정의 (Hook 전용 vs 공용 분리)
- **Phase 3**: session_analyzer를 Hook 시스템 내부로 이동 (선택적)
- **Phase 4**: 패키지-로컬 동기화 검증 자동화

---

## 🔍 1. 현재 상태 분석 (As-Is)

### 1.1 패키지 템플릿 구조

**위치**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/hooks/alfred/`

```
src/moai_adk/templates/.claude/hooks/alfred/
├── core/                                   # Top-level 공용 유틸리티
│   ├── project.py                         # 프로젝트 메타데이터 읽기
│   ├── timeout.py                         # 타임아웃 관리 (구 버전)
│   ├── ttl_cache.py                       # TTL 캐시 유틸리티
│   └── version_cache.py                   # 버전 캐시 관리
│
├── handlers/                              # (현재 비어있음)
│   └── __init__.py
│
├── shared/                                # Hook 간 공유 모듈
│   ├── core/                              # 핵심 비즈니스 로직
│   │   ├── __init__.py
│   │   ├── checkpoint.py                  # 체크포인트 관리
│   │   ├── context.py                     # 컨텍스트 관리
│   │   ├── project.py                     # 프로젝트 정보 (중복?)
│   │   ├── tags.py                        # @TAG 시스템
│   │   └── version_cache.py               # 버전 캐시 (중복?)
│   │
│   └── handlers/                          # Hook 이벤트 핸들러
│       ├── __init__.py
│       ├── notification.py                # 알림 핸들러
│       ├── session.py                     # 세션 핸들러
│       ├── tool.py                        # Tool 핸들러
│       └── user.py                        # User 핸들러
│
├── utils/                                 # 범용 유틸리티
│   ├── __init__.py
│   └── timeout.py                         # 크로스플랫폼 타임아웃
│
├── post_tool__log_changes.py             # PostTool Hook
├── pre_tool__auto_checkpoint.py           # PreTool Hook
├── session_end__cleanup.py                # SessionEnd Hook
├── session_start__show_project_info.py    # SessionStart Hook
└── user_prompt__jit_load_docs.py          # UserPrompt Hook

총 파일 수: 22개 (Python 파일만, __pycache__ 제외)
```

### 1.2 로컬 Hook 구조

**위치**: `/Users/goos/MoAI/MoAI-ADK/.claude/hooks/alfred/`

```
.claude/hooks/alfred/
├── core/                                   # (패키지와 동일)
│   ├── project.py
│   ├── timeout.py
│   ├── ttl_cache.py
│   └── version_cache.py
│
├── handlers/                              # (패키지와 동일)
│   └── __init__.py
│
├── shared/                                # Hook 간 공유 모듈
│   ├── core/                              # (패키지와 동일)
│   │   ├── __init__.py
│   │   ├── checkpoint.py
│   │   ├── context.py
│   │   ├── project.py
│   │   ├── tags.py
│   │   └── version_cache.py
│   │
│   └── handlers/                          # Hook 이벤트 핸들러
│       ├── __init__.py
│       ├── daily_analysis.py              # ⚠️ 로컬 전용 (패키지에 없음)
│       ├── notification.py
│       ├── session.py
│       ├── tool.py
│       └── user.py
│
├── utils/                                 # (패키지와 동일)
│   ├── __init__.py
│   └── timeout.py
│
├── post_tool__log_changes.py             # (패키지와 동일)
├── pre_tool__auto_checkpoint.py           # (패키지와 동일)
├── session_end__cleanup.py                # (패키지와 동일)
├── session_start__daily_analysis.py       # ⚠️ 로컬 전용 (패키지에 없음)
├── session_start__show_project_info.py    # (패키지와 동일)
└── user_prompt__jit_load_docs.py          # (패키지와 동일)

총 파일 수: 24개 (Python 파일만, __pycache__ 제외)
```

### 1.3 .moai/scripts/ 구조

**위치**: `/Users/goos/MoAI/MoAI-ADK/.moai/scripts/`

```
.moai/scripts/
├── init-dev-config.sh          # 개발 환경 초기화 (71 lines)
├── session_analyzer.py         # Claude Code 세션 로그 분석기 (337 lines)
└── weekly_analysis.sh          # 주간 분석 래퍼 스크립트 (68 lines)

총 파일 수: 3개
```

**session_analyzer.py 역할**:
- Claude Code 세션 로그 파싱 (`~/.claude/projects/*/session-*.json`)
- Tool 사용, 오류, 권한 요청, Hook 실패 패턴 분석
- 마크다운 리포트 생성 (`.moai/reports/daily-analysis-YYYY-MM-DD.md`)

**의존성 관계**:
```
session_start__daily_analysis.py (Hook)
    ↓ imports
shared/handlers/daily_analysis.py (Handler)
    ↓ subprocess.run(["python3", ".moai/scripts/session_analyzer.py", ...])
.moai/scripts/session_analyzer.py (Standalone Script)
```

---

## 🚨 2. 문제점 목록 (Issues)

### 2.1 구조 불일치

| 항목 | 로컬 | 패키지 템플릿 | 상태 |
|------|------|--------------|------|
| `session_start__daily_analysis.py` | ✅ 존재 | ❌ 누락 | **불일치** |
| `shared/handlers/daily_analysis.py` | ✅ 존재 | ❌ 누락 | **불일치** |
| 기타 Hook 파일 | ✅ | ✅ | 일치 |

**영향**:
- 패키지 템플릿으로 새 프로젝트 초기화 시 일일 분석 기능 누락
- 로컬 개발 전용 기능으로 간주되어 글로벌 배포 안 됨

### 2.2 의존성 경로 오류

**현재 구조**:
```python
# .claude/hooks/alfred/shared/handlers/daily_analysis.py (line 75)
result = subprocess.run(
    ["python3", ".moai/scripts/session_analyzer.py", "--days", "1"],
    cwd=cwd,
    ...
)
```

**문제점**:
1. **Hard-coded 경로**: `.moai/scripts/session_analyzer.py` 고정
2. **디렉토리 결합**: Hook 시스템이 `.moai/` 디렉토리 존재를 가정
3. **실행 환경 가정**: `python3` 명령어 사용 가능 가정 (시스템 Python)

**잠재적 오류 시나리오**:
- `.moai/scripts/` 디렉토리 없음 → FileNotFoundError (현재 try-except로 처리)
- `python3` 없는 환경 (Windows py.exe만 있는 경우) → FileNotFoundError
- 가상환경 활성화 안 됨 → 의존성 누락 (json, pathlib은 표준 라이브러리라 괜찮음)

### 2.3 패키지-로컬 동기화 누락

**동기화 규칙 (CLAUDE.md 명시)**:
> 항상 @src/moai_adk/templates/.claude/ @src/moai_adk/templates/.moai/ @src/moai_adk/templates/CLAUDE.md 에 변경이 생기면 로컬 프로젝트 폴더에도 동기화를 항상 하도록 하자. 패키지 템플릿이 가장 우선이다.

**현재 상황**:
- `session_start__daily_analysis.py` 로컬 추가 → 패키지 템플릿 미반영
- `.moai/scripts/session_analyzer.py` 로컬 추가 → 패키지 템플릿 `.moai/` 구조 없음

**원인 분석**:
- daily_analysis 기능을 로컬 실험으로 시작했으나 패키지화 단계 누락
- `.moai/scripts/`가 패키지 템플릿에 포함되어야 하는지 결정 보류

### 2.4 스크립트 정체성 모호

**.moai/scripts/ 디렉토리 용도 불명확**:

**옵션 A**: Hook 전용 스크립트 저장소
- session_analyzer.py는 Hook만 사용
- 패키지 템플릿에 포함해야 함
- `.claude/hooks/alfred/shared/scripts/`로 이동 고려

**옵션 B**: 프로젝트 공용 유틸리티
- Hook 외에 사용자가 수동으로도 실행 가능
- `.moai/scripts/`에 유지하되 패키지 템플릿에 포함
- Hook은 이 디렉토리를 외부 의존성으로 참조

**옵션 C**: 하이브리드
- `init-dev-config.sh` → 로컬 개발 전용 (패키지 제외)
- `session_analyzer.py` → Hook + 수동 실행 (패키지 포함)
- `weekly_analysis.sh` → 수동 실행 전용 (패키지 포함)

**현재 사용 패턴**:
```bash
# Hook에서 자동 실행
session_start__daily_analysis.py → session_analyzer.py --days 1

# 사용자 수동 실행
python3 .moai/scripts/session_analyzer.py --days 7 --output .moai/reports/weekly.md
```

---

## 🎯 3. 개선 계획 (To-Be)

### 3.1 정규화된 디렉토리 구조

**목표**: Hook 시스템 내부 응집도 향상 + 외부 의존성 명확화

#### 옵션 1: Hook 내부 통합 (권장)

```
.claude/hooks/alfred/
├── core/                                   # Top-level 공용 유틸리티
│   ├── project.py
│   ├── timeout.py
│   ├── ttl_cache.py
│   └── version_cache.py
│
├── shared/                                # Hook 간 공유 모듈
│   ├── core/                              # 핵심 비즈니스 로직
│   │   ├── __init__.py
│   │   ├── checkpoint.py
│   │   ├── context.py
│   │   ├── project.py
│   │   ├── tags.py
│   │   └── version_cache.py
│   │
│   ├── handlers/                          # Hook 이벤트 핸들러
│   │   ├── __init__.py
│   │   ├── daily_analysis.py              # ✅ 패키지 템플릿 추가
│   │   ├── notification.py
│   │   ├── session.py
│   │   ├── tool.py
│   │   └── user.py
│   │
│   └── scripts/                           # ⭐ NEW: Hook 전용 스크립트
│       ├── __init__.py
│       └── session_analyzer.py            # ⭐ .moai/scripts/에서 이동
│
├── utils/                                 # 범용 유틸리티
│   ├── __init__.py
│   └── timeout.py
│
├── [Hook 파일들...]
├── session_start__daily_analysis.py       # ✅ 패키지 템플릿 추가
└── session_start__show_project_info.py

# 변경된 import 경로
# Before: subprocess.run(["python3", ".moai/scripts/session_analyzer.py", ...])
# After:  from shared.scripts import session_analyzer
#         session_analyzer.analyze_sessions(days=1, output_dir=reports_dir)
```

**장점**:
- Hook 시스템이 자기완결적 (self-contained)
- 패키지 배포 시 모든 의존성 포함
- import 경로 단순화 (subprocess → Python import)

**단점**:
- 사용자가 수동으로 session_analyzer 실행 시 경로 변경 필요
- `.claude/hooks/alfred/shared/scripts/session_analyzer.py --days 7`

#### 옵션 2: .moai/ 디렉토리 패키지 포함

```
# 패키지 템플릿에 추가
src/moai_adk/templates/.moai/
├── scripts/
│   ├── session_analyzer.py                # ✅ 패키지 템플릿 추가
│   └── weekly_analysis.sh                 # ✅ 패키지 템플릿 추가
└── cache/                                 # 사용자 생성 디렉토리 (템플릿 제외)

# Hook 구조는 현재 유지
.claude/hooks/alfred/
├── shared/handlers/daily_analysis.py      # ✅ 패키지 템플릿 추가
└── session_start__daily_analysis.py       # ✅ 패키지 템플릿 추가

# subprocess 경로는 현재 유지
subprocess.run(["python3", ".moai/scripts/session_analyzer.py", ...])
```

**장점**:
- 사용자 수동 실행 경로 변경 없음
- `.moai/scripts/` 용도 명확화 (프로젝트 공용 유틸리티)
- Hook과 스크립트의 느슨한 결합 유지

**단점**:
- 패키지 구조 복잡도 증가 (`.claude/` + `.moai/` 두 곳 관리)
- subprocess 호출 오버헤드 유지

### 3.2 파일 분류 기준

| 파일 | 위치 | 패키지 포함 | 사유 |
|------|------|------------|------|
| `session_start__daily_analysis.py` | `.claude/hooks/alfred/` | ✅ 예 | 핵심 Hook 기능 |
| `shared/handlers/daily_analysis.py` | `.claude/hooks/alfred/shared/handlers/` | ✅ 예 | Hook 핸들러 |
| `session_analyzer.py` | **옵션 1**: `.claude/hooks/alfred/shared/scripts/`<br>**옵션 2**: `.moai/scripts/` | ✅ 예 | 분석 로직 재사용 |
| `weekly_analysis.sh` | `.moai/scripts/` | ✅ 예 | 사용자 수동 실행 |
| `init-dev-config.sh` | `.moai/scripts/` | ❌ 아니오 | MoAI-ADK 로컬 개발 전용 |

### 3.3 최종 권장 구조 (옵션 2 선택)

**근거**:
- session_analyzer.py는 Hook 외에도 사용자 수동 실행 용도
- `.moai/scripts/` 디렉토리를 프로젝트 공용 유틸리티 저장소로 명확히 정의
- 느슨한 결합 유지 (Hook ↔ Scripts 분리)

```
# 패키지 템플릿 구조
src/moai_adk/templates/
├── .claude/
│   └── hooks/alfred/
│       ├── shared/handlers/
│       │   └── daily_analysis.py          # ⭐ 추가
│       └── session_start__daily_analysis.py # ⭐ 추가
│
└── .moai/
    └── scripts/
        ├── session_analyzer.py            # ⭐ 추가
        └── weekly_analysis.sh             # ⭐ 추가
        # init-dev-config.sh는 제외 (로컬 전용)
```

---

## 🛠️ 4. 구현 로드맵

### Phase 1: 패키지 템플릿 정규화 ✅

**목표**: 누락된 파일을 패키지 템플릿에 추가

**작업 항목**:
1. ✅ `src/moai_adk/templates/.claude/hooks/alfred/session_start__daily_analysis.py` 추가
2. ✅ `src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/daily_analysis.py` 추가
3. ✅ `src/moai_adk/templates/.moai/scripts/` 디렉토리 생성
4. ✅ `src/moai_adk/templates/.moai/scripts/session_analyzer.py` 복사
5. ✅ `src/moai_adk/templates/.moai/scripts/weekly_analysis.sh` 복사

**검증**:
```bash
# 파일 존재 확인
ls src/moai_adk/templates/.claude/hooks/alfred/session_start__daily_analysis.py
ls src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/daily_analysis.py
ls src/moai_adk/templates/.moai/scripts/session_analyzer.py
ls src/moai_adk/templates/.moai/scripts/weekly_analysis.sh

# 구조 비교
diff -r .claude/hooks/alfred/ src/moai_adk/templates/.claude/hooks/alfred/ \
  --exclude="__pycache__" --exclude="*.pyc"
```

**예상 시간**: 30분

---

### Phase 2: 로컬 구조 적용 ✅

**목표**: 로컬 MoAI-ADK 프로젝트를 패키지 템플릿과 동기화

**작업 항목**:
1. ✅ `.moai/scripts/` 디렉토리 정리
   - `init-dev-config.sh` 유지 (로컬 전용 표시)
   - `session_analyzer.py` 검토 (패키지 버전과 동일한지 확인)
   - `weekly_analysis.sh` 검토

2. ✅ `.claude/hooks/alfred/` 검토
   - 패키지 템플릿과 차이점 확인
   - 로컬 전용 실험 파일 식별

**검증**:
```bash
# 패키지 템플릿 → 로컬 동기화 테스트
rsync -av --dry-run \
  src/moai_adk/templates/.claude/hooks/alfred/ \
  .claude/hooks/alfred/ \
  --exclude="__pycache__"

# 차이점 리포트
diff -qr src/moai_adk/templates/.claude/hooks/alfred/ .claude/hooks/alfred/ \
  --exclude="__pycache__"
```

**예상 시간**: 20분

---

### Phase 3: 경로 및 import 수정 (선택적) 🔄

**목표**: subprocess 경로 안정화 및 오류 처리 강화

**옵션 A: 현재 구조 유지 (subprocess 방식)**

```python
# shared/handlers/daily_analysis.py
def run_session_analyzer(cwd: str) -> bool:
    """Run session analyzer for previous day (--days 1)"""
    # 절대 경로 사용으로 안정성 향상
    script_path = Path(cwd) / ".moai" / "scripts" / "session_analyzer.py"

    if not script_path.exists():
        # 스크립트 없음 - 조용히 실패
        return False

    try:
        result = subprocess.run(
            [sys.executable, str(script_path), "--days", "1"],  # sys.executable 사용
            cwd=cwd,
            timeout=4,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        return False
    except Exception:
        return False
```

**개선 사항**:
- `python3` → `sys.executable` (현재 Python 인터프리터 사용)
- 상대 경로 → 절대 경로 (`Path(cwd) / ".moai/scripts/..."`)
- 스크립트 존재 여부 사전 확인

**옵션 B: Python import 방식 (옵션 1 선택 시)**

```python
# shared/handlers/daily_analysis.py
import sys
from pathlib import Path

# .moai/scripts를 sys.path에 추가
scripts_dir = Path(cwd) / ".moai" / "scripts"
if scripts_dir.exists() and str(scripts_dir) not in sys.path:
    sys.path.insert(0, str(scripts_dir))

try:
    import session_analyzer

    analyzer = session_analyzer.SessionAnalyzer(days_back=1)
    result = analyzer.run_analysis()
    return result.success
except ImportError:
    return False
except Exception:
    return False
```

**장점**:
- subprocess 오버헤드 제거
- 직접적인 함수 호출 (성능 향상)

**단점**:
- session_analyzer.py를 모듈로 리팩토링 필요 (현재는 스크립트)

**권장**: **옵션 A** (subprocess 방식 유지 + 경로 개선)
- session_analyzer.py의 독립 실행 가능성 유지
- 최소 변경으로 안정성 향상

**예상 시간**: 40분

---

### Phase 4: 동기화 검증 자동화 ✅

**목표**: 패키지-로컬 구조 일관성 자동 검증

**구현**:

**4.1 Skill 생성**: `moai-alfred-template-sync-checker`

```yaml
---
name: moai-alfred-template-sync-checker
description: Verify MoAI-ADK package template and local project synchronization
model: haiku
---

# Template Synchronization Checker

Verifies that local MoAI-ADK project structure matches package templates.

## Usage

```python
from pathlib import Path
import subprocess

def check_template_sync():
    """Check .claude/ and .moai/ synchronization"""

    template_base = Path("src/moai_adk/templates")
    local_base = Path(".")

    dirs_to_check = [
        ".claude/hooks/alfred",
        ".claude/skills",
        ".moai/scripts",  # 스크립트만 체크 (cache, reports 제외)
    ]

    for dir_path in dirs_to_check:
        template_dir = template_base / dir_path
        local_dir = local_base / dir_path

        # diff -qr 실행
        result = subprocess.run(
            ["diff", "-qr", str(template_dir), str(local_dir), "--exclude=__pycache__"],
            capture_output=True,
            text=True,
        )

        if result.returncode != 0:
            print(f"❌ Mismatch in {dir_path}:")
            print(result.stdout)
        else:
            print(f"✅ {dir_path} synchronized")
```

## Check Rules

1. **Must Match**:
   - `.claude/hooks/alfred/*.py` (Hook files)
   - `.claude/hooks/alfred/shared/**/*.py` (Shared modules)
   - `.moai/scripts/*.py` (except init-dev-config.sh)

2. **Local Only**:
   - `.moai/scripts/init-dev-config.sh` (development-only)
   - `.moai/cache/` (runtime cache)
   - `.moai/reports/` (runtime reports)

3. **Exceptions**:
   - `__pycache__/` directories
   - `*.pyc` files
```

**4.2 Pre-commit Hook 통합** (선택적)

```bash
# .git/hooks/pre-commit
#!/bin/bash
# Check template sync before commit

python3 -c "
from pathlib import Path
import subprocess
import sys

template_hooks = Path('src/moai_adk/templates/.claude/hooks/alfred')
local_hooks = Path('.claude/hooks/alfred')

result = subprocess.run(
    ['diff', '-qr', str(template_hooks), str(local_hooks), '--exclude=__pycache__'],
    capture_output=True
)

if result.returncode != 0:
    print('⚠️ Warning: Hook templates out of sync')
    print(result.stdout.decode())
    sys.exit(1)
"
```

**예상 시간**: 1시간

---

## 📊 5. 마이그레이션 체크리스트

### Phase 1: 패키지 템플릿 정규화

- [ ] `src/moai_adk/templates/.claude/hooks/alfred/session_start__daily_analysis.py` 생성
- [ ] `src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/daily_analysis.py` 생성
- [ ] `src/moai_adk/templates/.moai/scripts/` 디렉토리 생성
- [ ] `session_analyzer.py` 복사 (로컬 → 패키지)
- [ ] `weekly_analysis.sh` 복사 (로컬 → 패키지)
- [ ] 패키지 템플릿 구조 검증 (트리 비교)

### Phase 2: 로컬 구조 적용

- [ ] `.moai/scripts/init-dev-config.sh` 로컬 전용 표시 (주석 추가)
- [ ] 로컬 Hook 파일과 패키지 템플릿 diff 확인
- [ ] 차이점 문서화 (의도적 차이 vs 동기화 누락)

### Phase 3: 경로 및 import 수정

- [ ] `shared/handlers/daily_analysis.py` 경로 개선
  - [ ] `python3` → `sys.executable`
  - [ ] 상대 경로 → 절대 경로
  - [ ] 스크립트 존재 확인 로직 추가
- [ ] 로컬 테스트 (daily_analysis Hook 실행)
- [ ] 오류 처리 강화 (timeout, FileNotFoundError)

### Phase 4: 동기화 검증 자동화

- [ ] `moai-alfred-template-sync-checker` Skill 생성
- [ ] 동기화 검증 스크립트 작성
- [ ] CI/CD 파이프라인 통합 (선택적)
- [ ] Pre-commit Hook 추가 (선택적)

---

## 🎯 6. 성공 기준

### 6.1 구조 일관성

**검증 명령어**:
```bash
# Hook 디렉토리 동기화 확인
diff -qr \
  src/moai_adk/templates/.claude/hooks/alfred/ \
  .claude/hooks/alfred/ \
  --exclude="__pycache__" \
  --exclude="*.pyc"

# 예상 출력: (로컬 전용 파일 제외 시)
# Only in .claude/hooks/alfred: init-dev-config.sh  (정상 - 로컬 전용)
```

**기대 결과**:
- 패키지 템플릿과 로컬이 동일 (로컬 전용 파일 제외)
- 모든 Hook 파일 쌍방향 존재

### 6.2 기능 동작

**테스트 시나리오**:
1. **새 프로젝트 초기화**:
   ```bash
   uv run moai init test-project
   cd test-project

   # daily_analysis Hook 존재 확인
   ls .claude/hooks/alfred/session_start__daily_analysis.py
   ls .moai/scripts/session_analyzer.py
   ```

2. **Hook 실행 테스트**:
   ```bash
   # Claude Code 세션 시작 (SessionStart Hook 자동 실행)
   claude-code

   # 로그 확인
   cat .moai/cache/last_analysis_date.json
   ls .moai/reports/daily-analysis-*.md
   ```

3. **수동 분석 실행**:
   ```bash
   python3 .moai/scripts/session_analyzer.py --days 7 \
     --output .moai/reports/manual-test.md

   # 리포트 생성 확인
   cat .moai/reports/manual-test.md
   ```

**기대 결과**:
- ✅ Hook 자동 실행 성공 (5초 내 완료)
- ✅ 리포트 정상 생성
- ✅ 오류 없음 (stderr 비어있음)

### 6.3 문서화

**필수 문서**:
- [ ] `.moai/scripts/README.md` 생성 (스크립트 사용법)
- [ ] `session_start__daily_analysis.py` docstring 업데이트 (경로 정보)
- [ ] `Skill("moai-alfred-session-analytics")` 업데이트 (구조 변경 반영)

---

## 📝 7. 위험 요소 및 완화 전략

| 위험 | 영향 | 완화 전략 |
|------|------|----------|
| 패키지 배포 시 `.moai/scripts/` 누락 | 높음 | `pyproject.toml`에 `include` 명시 |
| 로컬 전용 파일 실수로 패키지 포함 | 중간 | `.gitignore` 및 `MANIFEST.in` 검증 |
| subprocess 타임아웃 (느린 시스템) | 낮음 | 타임아웃 4→5초 증가 고려 |
| Python 인터프리터 경로 문제 | 낮음 | `sys.executable` 사용 |

---

## 🔗 8. 관련 문서

- **Skill**: `moai-alfred-session-analytics` - 세션 분석 가이드
- **Skill**: `moai-cc-hooks` - Hook 시스템 설계
- **SPEC**: `SPEC-CLAUDE-PHILOSOPHY-001` - CLAUDE.md 철학
- **문서**: `.moai/scripts/session_analyzer.py` - 분석기 스크립트

---

## 📌 9. 다음 단계

### 즉시 실행 (이번 세션)

1. ✅ **Phase 1 완료**: 패키지 템플릿에 누락 파일 추가
2. ✅ **Phase 2 시작**: 로컬 구조 검증

### 후속 작업 (다음 세션)

3. 🔄 **Phase 3**: 경로 개선 (선택적)
4. 🔄 **Phase 4**: 자동화 스크립트 구현

### 장기 계획

- [ ] `.moai/scripts/` 용도 명확화 문서 작성
- [ ] 다른 프로젝트 템플릿에도 동일 구조 적용 (Data Science, CLI, FastAPI)
- [ ] Hook 성능 프로파일링 (SessionStart 5초 제한 최적화)

---

**작성자**: 🎩 Alfred@MoAI
**검토자**: GOOS
**최종 업데이트**: 2025-11-04
**버전**: 1.0.0
