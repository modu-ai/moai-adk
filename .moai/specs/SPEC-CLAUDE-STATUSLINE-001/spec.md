---
id: CLAUDE-STATUSLINE-001
version: 1.2.0
status: draft
created: 2025-11-07
updated: 2025-11-07
author: @GOOS
priority: high
category: feature
labels:
  - claude-code
  - ux
  - status-display
  - workflow-optimization
  - version-tracking
---

## HISTORY

### v1.2.0 (2025-11-07)
- **FEATURE**: MoAI-ADK 버전 정보 표시 기능 추가
- **FEATURE**: 업데이트 안내 기능 추가 (아이콘 + 최신 버전 표시)
- **UPDATE**: 상태줄 레이아웃에 [VERSION] 필드 추가
- **UPDATE**: 요구사항 2개 추가 (@REQ:STATUSLINE-UBQ-006, @REQ:STATUSLINE-EVENT-006)
- **UPDATE**: 색상 팔레트에 업데이트 알림 색상 추가
- **SCOPE**: 7가지 핵심 정보 표시 (모델, 시간, 디렉토리, 버전, Git branch, Git 상태, 작업 상태)
- **RATIONALE**: MoAI-ADK 프로젝트 버전을 한 눈에 파악하고 업데이트 가용성을 실시간으로 인지

### v1.1.0 (2025-11-07)
- **UPDATE**: 누적 비용([COST]) 정보 제거
- **SCOPE**: 핵심 5가지 정보만 표시 (모델, 시간, 디렉토리, Git branch, 작업 상태)
- **RATIONALE**: 비용 추적은 세션 로그에서 확인 가능하므로 상태줄에서 제거

### v1.0.0 (2025-11-07)
- **INITIAL**: MoAI-ADK 개발자를 위한 Claude Code 상태줄 기능 명세
- **AUTHOR**: @GOOS
- **SCOPE**: 개발 진행 상황, 프로젝트 상태, Git 정보를 통합 표시
- **CONTEXT**: Alfred 워크플로우 진행 상황을 한 눈에 파악

---

# SPEC: Claude Code 상태줄 (Statusline) 기능

## @SPEC:CLAUDE-STATUSLINE-001

MoAI-ADK 개발자가 Claude Code 상태줄에서 실시간 모델 정보, 세션 시간, 프로젝트 상태, Git 정보, Alfred 워크플로우 진행 상황을 통합 확인할 수 있는 기능 명세

---

## 1. Environment (환경)

### Required Dependencies
- **Claude Code**: v2.0.30 이상
- **MoAI-ADK**: v0.20.1+ (Python 3.10+)
- **Configuration**: `.moai/config.json` 설정 완료
- **Workflow**: `/alfred:0-project` 완료 상태
- **Git**: Local repository 활성화 (branch management)
- **Session State**: `.moai/memory/last-session-state.json` 지원

### System Requirements
- Python 3.10 or higher
- Git with `git branch`, `git status`, `git log` 지원
- Read access to `.moai/config.json`, `.moai/specs/`, `CLAUDE.md`
- File system access to project root directory

### Development Environment
- Claude Code IDE with statusline customization API
- Terminal emulator with ANSI color support (256-color)
- MoAI-ADK 프로젝트의 활성 세션

### Performance Constraints
- **Update Frequency**: 300ms (0.3초) 이하의 주기적 갱신
- **Cache Duration**: 5초 이상의 캐싱으로 성능 최적화
- **Maximum Display**: 80자 이내 (표준 터미널 너비)
- **Memory Footprint**: <5MB 메모리 사용

---

## 2. Assumptions (가정)

### User Behavior
- MoAI-ADK 개발자는 하루에 여러 번 장시간 Claude Code를 사용
- 활성 SPEC 작업 중에 진행 상황을 자주 확인하려고 함
- Git branch 및 uncommitted changes 상태를 신속하게 파악해야 함
- 현재 사용 중인 모델과 세션 시간을 알고 싶음

### System Behavior
- Alfred 워크플로우는 `/alfred:0-project` → `/alfred:1-plan` → `/alfred:2-run` → `/alfred:3-sync` 순서로 진행
- SPEC은 `.moai/specs/SPEC-{ID}/` 디렉토리 구조로 저장
- Git 상태는 로컬 저장소에서 빠르게 조회 가능
- Session metrics는 `.moai/logs/sessions/` 또는 메모리에 저장

### Technical Constraints
- Claude Code statusline API는 300ms 주기로 호출 (update frequency limit)
- Git 명령어 실행은 캐싱이 필수 (디스크 I/O 최소화)
- 색상 표현은 ANSI 256-color palette 사용 (호환성)
- 이모지는 선택적이지만 시각적 구분에 효과적

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous Requirements (항상 표시되어야 함)

@REQ:STATUSLINE-UBQ-001
**모델 및 세션 정보 표시**
- GIVEN: Claude Code 세션이 활성화되어 있을 때
- WHEN: 상태줄을 렌더링할 때
- THEN: 현재 모델 이름 (예: `Haiku 4.5`) 과 세션 경과 시간을 표시해야 함

@REQ:STATUSLINE-UBQ-002
**현재 working directory 표시**
- GIVEN: 프로젝트 디렉토리가 설정된 상태
- WHEN: 상태줄을 렌더링할 때
- THEN: 현재 디렉토리의 마지막 경로 부분 (예: `MoAI-ADK`) 또는 상대 경로를 표시해야 함

@REQ:STATUSLINE-UBQ-003
**현재 Git branch 표시**
- GIVEN: Git repository 활성화 상태
- WHEN: 상태줄을 렌더링할 때
- THEN: 현재 branch 이름 (예: `feature/SPEC-AUTH-001`, `develop`, `main`) 을 표시해야 함

@REQ:STATUSLINE-UBQ-004
**Git 저장소 상태 표시**
- GIVEN: Git repository에서 변경이 발생했을 때
- WHEN: 상태줄을 렌더링할 때
- THEN: Staged changes (+N), Unstaged changes (M N), Untracked files (?) 의 개수를 표시해야 함

@REQ:STATUSLINE-UBQ-005
**활성 Alfred 작업 및 단계 표시**
- GIVEN: `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync` 등 Alfred 명령이 실행 중일 때
- WHEN: 상태줄을 렌더링할 때
- THEN: 현재 실행 중인 Alfred 명령의 이름과 진행 상태 (예: `[PLAN]`, `[RUN-GREEN]`, `[SYNC]`) 를 표시해야 함

@REQ:STATUSLINE-UBQ-006
**MoAI-ADK 버전 표시**
- GIVEN: MoAI-ADK 프로젝트가 활성화되어 있을 때
- WHEN: 상태줄을 렌더링할 때
- THEN: MoAI-ADK의 현재 버전 (예: `v0.20.1` 또는 `0.20.1`) 을 표시해야 함

### 3.2 Event-Driven Requirements (특정 이벤트 발생 시)

@REQ:STATUSLINE-EVENT-001
**활성 SPEC ID 표시**
- GIVEN: `/alfred:1-plan` 또는 `/alfred:2-run` 명령이 SPEC을 작업 중일 때
- WHEN: 상태줄을 업데이트할 때
- THEN: 활성 SPEC의 ID (예: `SPEC-AUTH-001`) 와 현재 단계 (예: `RED`, `GREEN`, `REFACTOR`) 를 표시해야 함

@REQ:STATUSLINE-EVENT-002
**TDD 사이클 단계 표시**
- GIVEN: `/alfred:2-run` 명령이 TDD 사이클 중일 때
- WHEN: 상태줄을 업데이트할 때
- THEN: 현재 단계 (RED/GREEN/REFACTOR) 와 현재 작업 중인 테스트/코드 파일 정보를 표시해야 함

@REQ:STATUSLINE-EVENT-003
**활성 TodoWrite 작업 표시**
- GIVEN: TodoWrite tool로 작업 목록이 추적 중일 때
- WHEN: 상태줄을 업데이트할 때
- THEN: 현재 in_progress 작업의 개수와 completed 작업의 진행률을 표시해야 함

@REQ:STATUSLINE-EVENT-004
**경고 및 오류 표시**
- GIVEN: 테스트 실패, 빌드 오류, Git 충돌 등이 발생했을 때
- WHEN: 상태줄을 업데이트할 때
- THEN: 경고 아이콘과 함께 오류 유형 (예: `⚠ TESTS FAILED`, `✗ CONFLICTS`) 를 표시해야 함

@REQ:STATUSLINE-EVENT-006
**업데이트 안내 표시**
- GIVEN: MoAI-ADK 업데이트가 가능할 때
- WHEN: 상태줄을 업데이트할 때 (300초 캐싱, 60초 주기로 확인)
- THEN: 버전 정보 옆에 업데이트 아이콘 (⬆️ 또는 [UPDATE]) 과 함께 최신 버전 번호를 표시해야 함
- 색상: 주황색(38;5;208) 또는 파란색(38;5;33)
- 예시: `0.20.1 ⬆️ 0.21.0` 또는 `0.20.1 [UPDATE]`

### 3.3 State-Driven Requirements (상태에 따라 변경)

@REQ:STATUSLINE-STATE-001
**branch 색상 동적 변경**
- GIVEN: 현재 branch가 feature, develop, 또는 main일 때
- WHEN: 상태줄을 렌더링할 때
- THEN: 다음 색상 규칙을 적용해야 함:
  - feature/* → Yellow/Orange (작업 진행 중)
  - develop → Cyan/Blue (통합 브랜치)
  - main → Green (릴리스 브랜치)

@REQ:STATUSLINE-STATE-002
**Git 변경 상태 색상 표시**
- GIVEN: Git에서 변경 사항이 감지되었을 때
- WHEN: 상태줄을 렌더링할 때
- THEN: 변경 유형별로 색상을 다르게 표시해야 함:
  - Staged (+) → Green
  - Unstaged (M) → Yellow/Orange
  - Untracked (?) → Red/Pink
  - Clean → No indicator

@REQ:STATUSLINE-STATE-003
**세션 시간 기반 상태 표시**
- GIVEN: 세션 경과 시간이 증가할 때
- WHEN: 상태줄을 렌더링할 때
- THEN: 시간 형식을 동적으로 변경해야 함:
  - 5분 이내 → `5m30s` (초 단위)
  - 5-60분 → `15m` (분 단위)
  - 1시간 이상 → `2h 30m` (시간:분)

### 3.4 Optional Requirements (선택적 표시)

@REQ:STATUSLINE-OPT-001
**최근 커밋 메시지 스니펫 표시**
- GIVEN: 개발자가 설정에서 상세 정보 표시를 활성화했을 때
- WHEN: 상태줄의 확장 영역을 렌더링할 때
- THEN: 최근 커밋의 첫 50자를 표시해야 함

@REQ:STATUSLINE-OPT-002
**활성 SPEC 목록 미니 표시**
- GIVEN: 개발자가 '다중 SPEC 모드'를 활성화했을 때
- WHEN: 상태줄의 확장 영역을 렌더링할 때
- THEN: 현재 프로젝트의 활성 SPEC 3개의 ID 목록을 표시해야 함

@REQ:STATUSLINE-OPT-003
**AI 토큰 사용량 표시 (고급)**
- GIVEN: Claude Code API에서 토큰 메트릭을 제공하고 개발자가 활성화했을 때
- WHEN: 상태줄을 렌더링할 때
- THEN: 누적 input/output 토큰 비율 (예: `I:5K O:2K`) 을 표시해야 함

---

## 4. Specifications (기술 명세)

### 4.1 상태줄 레이아웃 구조

```
[MODEL] [DURATION] | [DIR] | [VERSION] [UPDATE-INDICATOR] | [BRANCH] | [GIT-STATUS] | [ACTIVE-TASK]
```

**각 섹션 설명:**
- `[MODEL]`: 현재 모델명 (예: `H 4.5` = Haiku 4.5, `S 4.5` = Sonnet 4.5)
- `[DURATION]`: 세션 경과 시간 (예: `5m`, `1h 30m`)
- `[DIR]`: 프로젝트명 또는 current directory의 마지막 경로
- `[VERSION]`: MoAI-ADK 버전 (예: `v0.20.1` 또는 `0.20.1`)
- `[UPDATE-INDICATOR]`: 업데이트 가능 여부 (예: `⬆️`, `[UPDATE]`, 또는 공백)
- `[BRANCH]`: Git branch 이름 (색상 코드 포함)
- `[GIT-STATUS]`: 변경 사항 지표 (예: `+3 M2 ?1`)
- `[ACTIVE-TASK]`: Alfred 작업 또는 TodoWrite 진행 상황

### 4.2 디스플레이 모드

#### Mode 1: Compact (기본, 80자 제한)
```
H 4.5 | 5m | MoAI-ADK | 0.20.1 | feature/SPEC-AUTH-001 | +2 M1 | [PLAN]
```

**버전 업데이트 있을 때:**
```
H 4.5 | 5m | MoAI-ADK | 0.20.1 ⬆️ 0.21.0 | feature/SPEC-AUTH-001 | +2 M1 | [PLAN]
```

#### Mode 2: Extended (120자, 시간 추적)
```
Haiku 4.5 | 1h 30m | /Users/goos/MoAI/MoAI-ADK | v0.20.1 | feature/SPEC-AUTH-001 (develop) | +5 M3 ?2 | [RUN-GREEN]
```

**버전 업데이트 있을 때:**
```
Haiku 4.5 | 1h 30m | /Users/goos/MoAI/MoAI-ADK | v0.20.1 (latest: v0.21.0) | feature/SPEC-AUTH-001 | +5 M3 ?2 | [RUN-GREEN]
```

#### Mode 3: Minimal (40자, 극도로 제한된 환경)
```
H | 5m | 0.20.1 | feature/AUTH | +2M
```

**버전 업데이트 있을 때:**
```
H | 5m | 0.20.1↑ | feature/AUTH | +2M
```

### 4.3 색상 팔레트 (ANSI 256-color)

| 요소 | 색상 코드 | 용도 |
|------|---------|------|
| Model | `38;5;33` (Blue) | 모델명 강조 |
| Version | `38;5;33` (Blue) | 버전 정보 |
| Feature branch | `38;5;226` (Yellow) | 작업 진행 중 표시 |
| Develop branch | `38;5;51` (Cyan) | 통합 브랜치 |
| Main branch | `38;5;46` (Green) | 릴리스 브랜치 |
| Staged (+) | `38;5;46` (Green) | 커밋 준비 완료 |
| Modified (M) | `38;5;208` (Orange) | 수정됨 |
| Untracked (?) | `38;5;196` (Red) | 추적 안 됨 |
| Update available | `38;5;208` (Orange) | 업데이트 가능 알림 |
| Success | `38;5;46` (Green) | 성공 상태 |
| Error | `38;5;196` (Red) | 오류 상태 |

### 4.4 이모지 및 기호

| 기호 | 의미 | 사용 시점 |
|------|------|----------|
| 🔵 (또는 `●`) | Alfred 작업 활성 | `/alfred` 명령 실행 중 |
| 🟡 (또는 `◐`) | TDD 진행 중 | RED/GREEN/REFACTOR 단계 |
| 🟢 (또는 `✓`) | 테스트 통과 | 모든 테스트 성공 |
| 🔴 (또는 `✗`) | 오류/실패 | 테스트 실패, 빌드 오류 |
| ⚠️ (또는 `!`) | 경고 | 미저장 변경 |
| ⬆️ (또는 `↑`) | 업데이트 가능 | 새로운 버전 있을 때 |
| 📝 | 작업 추적 중 | TodoWrite 활성 |
| 💾 | 미저장 변경 | Uncommitted changes |

### 4.5 데이터 수집 및 캐싱 전략

```python
# 캐싱 계층 구조
class StatuslineCache:
    git_info:        # 5초 캐싱
        branch, changed_files, staged_count

    session_metrics: # 10초 캐싱
        current_duration

    version_info:    # 60초 캐싱
        current_version (읽기: .moai/config.json)

    update_check:    # 300초 캐싱
        latest_version (PyPI 또는 GitHub API)
        update_available (boolean)

    active_task:     # 1초 캐싱 (자주 업데이트)
        alfred_command, spec_id, tdd_stage

    project_info:    # 60초 캐싱
        project_name, active_specs, config
```

**캐시 갱신 트리거:**
- File system change event (`.moai/` 폴더, `.moai/config.json` 감지)
- Git event (branch change, file modification)
- Session state update (duration)
- Version file change (`.moai/config.json` 변경 감지 즉시 갱신)
- Update check timeout (300초마다 새로 확인)
- Explicit refresh (300ms 주기)

### 4.6 성능 최적화 기법

1. **Lazy Loading**: 필요한 정보만 조회
   - Git info: `git status --porcelain` (fast)
   - Session metrics: 메모리 캐시에서 읽기
   - Active task: `.moai/memory/last-session-state.json` (fast read)

2. **Batch Operations**: 여러 정보를 한 번에 수집
   - `git status -b --porcelain --short` (한 번에 branch + changes 조회)
   - `.moai/specs/` 디렉토리 캐싱 (매 시간 1회)

3. **Background Refresh**: 비용이 많이 드는 작업은 백그라운드에서 처리
   - 최근 커밋 조회 (optional 정보)
   - SPEC 목록 업데이트

### 4.7 에러 처리 및 Fallback

```
상황                          표시 내용
───────────────────────────────────────────
Git 명령 실패                 [GIT N/A] (회색)
Session metrics 읽기 실패     0s (회색)
버전 읽기 실패                [???] (회색)
업데이트 확인 실패            아무것도 표시하지 않음 (실패는 무시)
Alfred 작업 상태 조회 실패    [?] (물음표)
디렉토리 권한 오류            [RESTRICTED] (회색)
```

---

## 5. Constraints & Considerations (제약사항)

### 5.1 성능 제약
- **Update Frequency**: 최대 300ms (Claude Code statusline API limit)
- **Cache Duration**: 5초 이상 (디스크 I/O 최소화)
- **Maximum CPU**: <2% (statusline update 시)
- **Memory**: <5MB (캐시 포함)

### 5.2 호환성 제약
- **Terminal**: 256-color 지원 (또는 fallback to 16-color)
- **OS**: macOS (Darwin), Linux, Windows (WSL)
- **Python**: 3.10 이상 (f-string, dataclass 지원)
- **Git**: 2.20+ (git status -b 지원)

### 5.3 정보 보안 제약
- **API Keys**: 상태줄에 절대 표시하면 안 됨
- **File Paths**: 민감한 경로명은 마스킹해야 함 (예: `~` 사용)
- **Personal Info**: 사용자명, 이메일 절대 표시 금지

### 5.4 사용자 경험 제약
- **읽기 시간**: 한 번에 모든 정보를 3초 이내에 읽을 수 있어야 함
- **시각적 복잡도**: 색상 사용은 최대 4가지 (파란색, 초록색, 노랑색, 빨간색)
- **이모지 지원**: 터미널이 이모지를 지원하지 않으면 기호로 fallback
- **Accessibility**: 색상에만 의존하지 않기 (기호나 텍스트 병행)

---

## 6. Traceability

### Related SPECs
- @SPEC:CLAUDE-CODE-FEATURES-001 - Claude Code v2.0.30+ 신규 기능 통합
- @SPEC:ALF-WORKFLOW-001 - Alfred 워크플로우 4단계 프로세스

### Related Code Modules
- @CODE:STATUSLINE-ENGINE-001 - 상태줄 렌더링 엔진
- @CODE:CACHE-MANAGER-001 - 캐싱 및 성능 최적화
- @CODE:GIT-INFO-COLLECTOR-001 - Git 정보 수집
- @CODE:SESSION-METRICS-001 - 세션 메트릭 추적
- @CODE:VERSION-READER-001 - MoAI-ADK 버전 정보 읽기
- @CODE:UPDATE-CHECKER-001 - 업데이트 가용성 확인

### Related Documentation
- @DOC:STATUSLINE-CONFIG-001 - 설정 및 커스터마이제이션 가이드
- @DOC:STATUSLINE-EXAMPLES-001 - 실제 사용 예시
- @DOC:PERFORMANCE-GUIDE-001 - 성능 최적화 가이드

---

## 7. Version & Changelog

### v1.2.0 (2025-11-07) - Version Display & Update Notifications
- MoAI-ADK 버전 정보 표시 기능 추가
- 업데이트 안내 기능 추가 (아이콘 + 최신 버전)
- Ubiquitous Requirements 6개 (5 → 6)
- Event-Driven Requirements 5개 (4 → 5)
- 상태줄 레이아웃에 [VERSION] 필드 추가
- 캐싱 전략에 version_info (60초) + update_check (300초) 추가
- 색상 팔레트에 "Update available" (Orange) 추가
- 이모지 및 기호에 업데이트 아이콘 추가
- 7가지 정보 → 8가지 정보로 확장 (버전 + 업데이트 아이콘)
- 기술 모듈 2개 추가 (version_reader.py, update_checker.py)

### v1.1.0 (2025-11-07) - Cost Removal
- 누적 비용([COST]) 정보 제거
- Ubiquitous Requirements 5개로 축소
- 6가지 정보 → 5가지 정보로 단순화
- 비용 관련 State-Driven Requirements 제거

### v1.0.0 (2025-11-07) - Initial Release
- EARS 방식의 상태줄 요구사항 정의
- 7가지 핵심 기능 정의
- 3가지 디스플레이 모드 (Compact/Extended/Minimal)
- 색상 팔레트 및 캐싱 전략 정의
