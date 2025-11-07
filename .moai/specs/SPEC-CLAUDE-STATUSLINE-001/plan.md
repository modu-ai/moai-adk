---
spec_id: CLAUDE-STATUSLINE-001
version: 1.2.0
created: 2025-11-07
---

# 구현 계획 및 기술 전략

## 프로젝트 개요

MoAI-ADK 개발자가 Claude Code 상태줄에서 실시간으로 현재 모델, 세션 시간, 프로젝트 상태, Git 정보, Alfred 워크플로우 진행 상황을 통합 확인할 수 있는 기능 구현

---

## 1차 목표: 핵심 상태줄 엔진 개발

### 작업 1: 기본 상태줄 렌더러 구현
**목표**: Compact 모드에서 기본 정보 7가지를 올바르게 포맷팅하는 렌더러 개발 (버전 정보 포함)

**기술 접근:**
- Python dataclass로 StatuslineData 구조 정의 (version_info 필드 추가)
- 각 정보별 수집/포맷팅 함수 분리 (단일 책임 원칙)
- f-string으로 Compact 레이아웃 구현
- ANSI 색상 코드 적용 (256-color 팔레트)

**구현 범위:**
- `src/moai_adk/statusline/renderer.py` - 핵심 렌더러
- `src/moai_adk/statusline/formatter.py` - 정보 포맷팅
- `src/moai_adk/statusline/colors.py` - 색상 팔레트 정의 (업데이트 색상 추가)
- Unit tests: `tests/statusline/test_renderer.py`

**산출물:**
- StatuslineRenderer 클래스 (render() 메서드)
- 각 정보별 collect_* 메서드
- Compact 모드 완전 지원 (7가지 정보 포함)

---

### 작업 2: Git 정보 수집 모듈
**목표**: Git branch, changes 정보를 빠르게 수집/캐싱하는 모듈 개발

**기술 접근:**
- `subprocess.run('git status -b --porcelain')` 한 번 호출로 모든 정보 수집
- 정규표현식으로 branch 이름 추출
- 변경 사항 count 계산 (staged/unstaged/untracked)
- 5초 TTL 캐싱으로 성능 최적화

**구현 범위:**
- `src/moai_adk/statusline/git_collector.py` - Git 정보 수집
- `src/moai_adk/statusline/cache.py` - 캐싱 레이어
- Unit tests: `tests/statusline/test_git_collector.py`

**산출물:**
- GitInfo dataclass
- GitCollector 클래스 (collect_git_info() 메서드)
- 오류 처리 및 fallback 로직

---

### 작업 3: 세션 메트릭 추적 모듈
**목표**: 세션 경과 시간을 추적/전달하는 모듈 개발

**기술 접근:**
- `.moai/memory/last-session-state.json` 읽기 (현재 세션)
- 시간 계산: 세션 시작 시간 기반 경과 시간 계산
- JSON 파싱 및 캐싱 (10초 TTL)

**구현 범위:**
- `src/moai_adk/statusline/metrics_tracker.py` - 메트릭 추적
- `src/moai_adk/statusline/session_reader.py` - 세션 상태 읽기
- Unit tests: `tests/statusline/test_metrics_tracker.py`

**산출물:**
- SessionMetrics dataclass
- MetricsTracker 클래스 (track_duration())
- 시간 포맷팅 (분/시간 단위 동적 변환)

---

### 작업 4: Alfred 작업 상태 감지 모듈
**목표**: 현재 실행 중인 Alfred 명령과 SPEC 진행 상황을 감지하는 모듈 개발

**기술 접근:**
- `CLAUDE.md` 파싱으로 현재 작업 상태 확인
- `.moai/memory/last-session-state.json` 에서 active_task 정보 읽기
- `.moai/specs/` 디렉토리 스캔으로 활성 SPEC 감지
- TDD 단계 감지: Git 커밋 메시지 또는 TodoWrite 상태 기반

**구현 범위:**
- `src/moai_adk/statusline/alfred_detector.py` - Alfred 작업 감지
- `src/moai_adk/statusline/spec_tracker.py` - SPEC 추적
- Unit tests: `tests/statusline/test_alfred_detector.py`

**산출물:**
- AlfredTask dataclass (command, spec_id, stage)
- AlfredDetector 클래스 (detect_current_task())
- TDD 단계 판별 로직

---

### 작업 5: MoAI-ADK 버전 정보 읽기 모듈
**목표**: MoAI-ADK 버전을 `.moai/config.json`에서 읽어 상태줄에 표시하는 모듈 개발

**기술 접근:**
- `.moai/config.json` 읽기 및 JSON 파싱
- `version` 필드 추출 (예: "0.20.1")
- 캐싱 (60초 TTL): 파일 변경 감지로 즉시 갱신
- 오류 처리: 파일 없음/파싱 실패 시 `[???]` 표시

**구현 범위:**
- `src/moai_adk/statusline/version_reader.py` - 버전 정보 읽기
- Unit tests: `tests/statusline/test_version_reader.py`

**산출물:**
- VersionInfo dataclass (current_version, last_updated)
- VersionReader 클래스 (read_version())
- 60초 캐싱 로직 + 파일 변경 감지

---

### 작업 6: 업데이트 확인 모듈
**목표**: PyPI 또는 GitHub API에서 최신 MoAI-ADK 버전을 조회하여 업데이트 안내 표시

**기술 접근:**
- PyPI API 또는 GitHub API 호출 (https://pypi.org/pypi/moai-adk/json)
- 최신 버전 추출 (release.version)
- 버전 비교: `current_version < latest_version`
- 캐싱 (300초 TTL): 최소 API 호출 빈도 유지
- 오류 처리: API 실패 시 아무것도 표시하지 않음 (무시)

**구현 범위:**
- `src/moai_adk/statusline/update_checker.py` - 업데이트 확인
- `src/moai_adk/statusline/api_client.py` - PyPI/GitHub API 클라이언트
- Unit tests: `tests/statusline/test_update_checker.py`

**산출물:**
- UpdateInfo dataclass (latest_version, available, last_checked)
- UpdateChecker 클래스 (check_for_update())
- 300초 캐싱 로직 + 비교 알고리즘

---

## 2차 목표: 고급 기능 개발

### 작업 7: 다중 디스플레이 모드 지원
**목표**: Compact, Extended, Minimal 모드를 모두 지원 (버전 정보 포함)

**기술 접근:**
- 템플릿 기반 렌더링 (각 모드별 템플릿)
- 정보 가시성 레벨 (FULL, COMPACT, MINIMAL)
- 동적 너비 계산 및 텍스트 truncation
- 설정 파일에서 preferred mode 읽기 (`.moai/config.json`)
- 버전 정보 표시 형식 차별화 (compact: `0.20.1`, extended: `v0.20.1`)

**구현 범위:**
- `src/moai_adk/statusline/renderer.py` 확장 (render_extended, render_minimal)
- Template 관리 (각 모드별)
- Unit tests: `tests/statusline/test_modes.py`

**산출물:**
- render_compact(), render_extended(), render_minimal() 메서드
- 각 모드의 정확한 포맷팅 구현 (버전 + 업데이트 아이콘 포함)

---

### 작업 8: 색상 팔레트 및 테마 시스템
**목표**: ANSI 색상 자동 적용 및 테마 커스터마이제이션 지원 (업데이트 색상 포함)

**기술 접근:**
- 색상 팔레트 JSON 정의 (256-color 매핑, 업데이트 색상 추가)
- 테마 선택: light, dark, high-contrast
- 자동 색상 선택 (터미널 배경색 감지)
- Fallback: ANSI 16-color 또는 plain text

**구현 범위:**
- `src/moai_adk/statusline/themes.py` - 테마 시스템
- `src/moai_adk/statusline/color_manager.py` - 색상 관리 (update_available 색상 추가)
- Unit tests: `tests/statusline/test_colors.py`

**산출물:**
- Theme dataclass
- ColorManager 클래스 (get_color(), apply_theme())
- 기본 테마 3가지 (light, dark, high-contrast)
- 업데이트 알림 색상 지원

---

### 작업 9: 통합 및 Claude Code API 연동
**목표**: Claude Code statusline API와 통합하여 실제 상태줄에 표시

**기술 접근:**
- Claude Code extension API 문서 확인
- HTTP 또는 IPC 통신 채널 설정
- 300ms 주기 갱신 구현
- 상태줄 업데이트 요청 전달

**구현 범위:**
- `src/moai_adk/statusline/api_client.py` - Claude Code API 클라이언트
- `src/moai_adk/statusline/integration.py` - 통합 계층
- Unit tests: `tests/statusline/test_integration.py`

**산출물:**
- StatuslineClient 클래스 (send_update())
- 300ms 타이머 구현
- API 에러 처리 및 재시도 로직

---

## 3차 목표: 성능 최적화 및 배포

### 작업 10: 종합 성능 최적화
**목표**: CPU/메모리 <2%/5MB 제약 충족, 캐싱 전략 최적화

**기술 접근:**
- 프로파일링: `cProfile`, `memory_profiler` 사용
- 캐싱 계층 구조 검증 (5초, 10초, 60초 TTL)
- 병렬 처리: 백그라운드 스레드에서 slow operations 수행
- 점진적 업데이트: 중요한 정보부터 먼저 표시

**구현 범위:**
- `src/moai_adk/statusline/performance.py` - 성능 관리
- Benchmarks: `tests/statusline/benchmarks/`
- 부하 테스트 시나리오

**산출물:**
- 성능 메트릭 리포트 (CPU, 메모리, 응답시간)
- 캐싱 효율 분석
- 최적화 전/후 비교

---

### 작업 11: 오류 처리 및 Fallback 전략
**목표**: 예외 상황에서도 상태줄이 graceful하게 작동

**기술 접근:**
- 각 정보 수집 함수에 try-except 추가
- 실패 시 기본값 또는 회색 표시
- 로깅 (`.moai/logs/statusline.log`)
- 사용자 알림 (옵션)

**구현 범위:**
- Exception handling in all collectors
- Fallback templates
- Logging configuration
- Unit tests: `tests/statusline/test_error_handling.py`

**산출물:**
- 모든 예외 상황에서의 안전한 동작
- 상세한 로그 (디버깅용)

---

### 작업 12: 사용자 문서 및 가이드
**목표**: 사용자가 쉽게 설정하고 사용할 수 있도록 문서 작성

**작업 내용:**
- Quick Start 가이드 (5분 안에 상태줄 활성화)
- 설정 옵션 문서 (`.moai/config.json`)
- 커스터마이제이션 예시 (색상, 모드, 정보 선택)
- 트러블슈팅 가이드

**산출물:**
- `docs/guides/statusline/setup.md` - 설정 가이드
- `docs/guides/statusline/examples.md` - 사용 예시
- `docs/guides/statusline/troubleshooting.md` - 트러블슈팅

---

## 마일스톤 구조

### Primary Goal (1주)
1. 기본 상태줄 렌더러 구현 (7가지 정보 포함)
2. Git 정보 수집 모듈
3. 세션 메트릭 추적 모듈
4. Alfred 작업 상태 감지
5. MoAI-ADK 버전 정보 읽기 모듈
6. 업데이트 확인 모듈

**목표**: Compact 모드 완전 작동 (버전 + 업데이트 안내 포함)

### Secondary Goal (2주)
7. 다중 디스플레이 모드 지원 (버전 정보 포함)
8. 색상 팔레트 및 테마 시스템 (업데이트 색상 포함)
9. Claude Code API 통합

**목표**: Extended, Minimal 모드 + 테마 지원 (버전 표시 차별화)

### Final Goal (3주)
10. 성능 최적화 및 벤치마킹
11. 오류 처리 및 Fallback (버전/업데이트 오류 포함)
12. 사용자 문서 및 가이드

**목표**: 프로덕션 배포 준비 완료

---

## 기술 스택

### 언어 및 프레임워크
- **Python**: 3.10+ (f-string, dataclass, subprocess)
- **Libraries**:
  - `dataclasses` - 데이터 구조 정의
  - `subprocess` - Git 명령 실행
  - `json` - 세션 상태 읽기, 설정 파일 읽기
  - `threading` - 백그라운드 작업
  - `pathlib` - 파일 경로 관리
  - `requests` - PyPI/GitHub API 호출 (선택적)
  - `urllib` - API 호출 (requests 없을 경우)

### 테스트 프레임워크
- **pytest**: 유닛 테스트
- **pytest-cov**: 커버리지 리포팅
- **pytest-mock**: Mock 객체
- **pytest-benchmark**: 성능 테스트

### 개발 도구
- **Black**: 코드 포맷팅
- **Pylint**: 정적 분석
- **mypy**: 타입 체킹
- **cProfile**: 성능 프로파일링

---

## 아키텍처 설계

### 계층 구조
```
┌─────────────────────────────────────┐
│  Claude Code API Integration Layer   │ (statusline_integration.py)
├─────────────────────────────────────┤
│  Statusline Renderer                │ (renderer.py, formatter.py)
├─────────────────────────────────────┤
│  Collectors & Detectors             │
│  ├─ Git Collector                   │ (git_collector.py)
│  ├─ Metrics Tracker                 │ (metrics_tracker.py)
│  ├─ Alfred Detector                 │ (alfred_detector.py)
│  └─ Session Reader                  │ (session_reader.py)
├─────────────────────────────────────┤
│  Cache & Performance Layer          │ (cache.py, performance.py)
├─────────────────────────────────────┤
│  File System & Git Layer            │ (.moai/, .git/)
└─────────────────────────────────────┘
```

### 모듈 책임
- **Renderer**: 정보를 상태줄 텍스트로 변환
- **Collectors**: Git, Session, Alfred 정보 수집
- **Cache**: 성능 최적화 (캐싱)
- **Colors**: 색상 관리 및 테마
- **API Client**: Claude Code와 통신

---

## 위험 요소 및 대응

| 위험 | 영향도 | 대응 방안 |
|-----|--------|---------|
| Git 명령 시간 초과 | High | subprocess timeout 설정, 캐싱 강화 |
| Claude Code API 변경 | Medium | 버전 검사, fallback 메커니즘 |
| 색상 지원 부재 | Low | ANSI 16-color fallback, plain text |
| 파일 권한 오류 | Low | try-except, graceful degradation |
| 메모리 누수 | Medium | 정기적인 메모리 프로파일링 |

---

## 성공 기준

1. **기능성**: 6가지 핵심 정보가 모두 표시됨
2. **성능**: CPU <2%, 메모리 <5MB
3. **신뢰성**: 99% 이상의 uptime (오류 시에도 표시)
4. **사용성**: 5분 이내 설정, 3가지 모드 지원
5. **호환성**: macOS, Linux, Windows(WSL) 모두 지원

---

## 배포 계획

### 단계 1: 알파 테스트 (내부)
- MoAI-ADK 핵심 팀이 사용해보기
- 피드백 수집 및 개선

### 단계 2: 베타 테스트 (선택 사용자)
- GitHub에 공개 후 선택 사용자 참여
- 각 OS 환경에서 테스트

### 단계 3: GA 릴리스
- v0.20.2 이상에 포함
- 공식 문서 배포
- 릴리스 노트 작성

---

## 리소스 및 의존성

### 필수 리소스
- Git 저장소 access
- `.moai/` 디렉토리 구조
- Claude Code statusline API 문서

### 외부 의존성
- Claude Code v2.0.30+
- Python 3.10+
- Standard library (subprocess, pathlib, json, threading)

### 선택적 의존성
- `rich` - 고급 서식 지정 (옵션)
- `psutil` - 시스템 정보 (옵션)
