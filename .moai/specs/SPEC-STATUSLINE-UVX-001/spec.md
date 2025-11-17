---
spec_id: SPEC-STATUSLINE-UVX-001
title: "uvx 기반 Statusline 실행 - 모든 OS 지원"
version: 1.0
status: draft
created_at: 2025-11-17
author: Claude Code (spec-builder)
priority: high
impact_area:
  - Infrastructure
  - Cross-Platform Compatibility
  - Performance Optimization
---

## HISTORY

| 버전 | 날짜 | 변경사항 | 작성자 |
|------|------|---------|--------|
| 1.0 | 2025-11-17 | 초기 SPEC 작성 | Claude Code (spec-builder) |

---

## 개요

현재 Claude Code statusline은 `uv run python -m moai_adk.statusline.main` 방식으로 실행되어 모든 OS에서 추가 프로세스 오버헤드를 발생시킨다. 본 SPEC은 `uvx moai-adk statusline` 명령어를 통해 모든 OS(Windows, macOS, Linux)에서 일관된 방식으로 statusline을 실행하도록 표준화하는 것을 목표로 한다.

### 핵심 목표
- 모든 OS에서 `uvx` 방식으로 일관된 실행
- Windows에서 다단계 프로세스 생성 오버헤드 제거
- 현재 다단계 폴백 메커니즘 유지 (안정성)
- 템플릿 파일 동기화 (모든 신규 프로젝트에 자동 적용)

---

## 환경 (Environment)

### 시스템 요구사항
- Python 3.11+
- uv ≥ 0.4.0 (uvx 명령어 지원)
- 모든 OS: Windows (10+), macOS (10.14+), Linux (모든 주요 배포판)

### 현재 상태
- `.claude/settings.json`: `uv run python -m moai_adk.statusline.main` 정의
- `src/moai_adk/statusline/main.py`: 메인 엔트리포인트
- `pyproject.toml`: `moai-adk` CLI 스크립트 정의
- `moai-adk` 패키지: PyPI에 배포됨

### 배포 환경
- 로컬 개발 환경: MoAI-ADK 소스 직접 실행
- 패키지 설치 환경: PyPI에서 `moai-adk` 설치
- 템플릿 프로젝트: 신규 프로젝트에서 설정 자동 복사

---

## 가정 (Assumptions)

### 기술적 가정
1. **uvx 캐시**: uvx는 자동으로 도구를 캐시하여 반복 실행 시 빠른 성능 보장
2. **진입점 자동화**: `pyproject.toml`의 `[project.scripts]` 섹션이 CLI 스크립트 생성
3. **다단계 폴백**: uvx 실패 시 `uv run` 방식으로 자동 폴백 가능
4. **환경 변수 전달**: Claude Code의 JSON 컨텍스트가 uvx 프로세스로 전달 가능

### 운영 가정
1. 모든 프로젝트는 `uv`를 설치한 환경에서 실행
2. .claude/settings.json은 사용자 기본 설정 (권장)
3. 새로운 프로젝트는 템플릿에서 설정 자동 복사
4. 템플릿 변경은 즉시 신규 프로젝트에 적용 (기존 프로젝트는 수동 마이그레이션)

---

## 요구사항 (Requirements)

### 기능 요구사항

#### R1.1: uvx 기반 CLI 진입점 구현
- `moai-adk statusline` 명령어가 statusline 실행
- `pyproject.toml`의 `[project.scripts]` 섹션에 새로운 진입점 추가
- 진입점은 `moai_adk.statusline.main:main` 함수 호출

#### R1.2: 모든 OS에서 일관된 실행
- Windows, macOS, Linux에서 동일한 명령어 동작
- OS별 경로 문제 없음 (uvx가 자동 처리)
- stderr/stdout 인코딩 일관성 유지

#### R1.3: .claude/settings.json 업데이트
- `statusLine.command` 필드를 `uvx moai-adk statusline`로 변경
- 로컬 프로젝트에서 즉시 적용

#### R1.4: 템플릿 동기화
- `src/moai_adk/templates/.claude/settings.json`에 변경사항 반영
- 신규 프로젝트는 uvx 방식으로 자동 설정

#### R1.5: 다단계 폴백 메커니즘
- uvx 실패 시 `uv run python -m moai_adk.statusline.main` 폴백
- 폴백 로직은 Claude Code 설정 레벨이 아닌 워퍼 스크립트로 구현 (향후)

### 제약사항

#### C1: Windows 프로세스 오버헤드
- 현재: `uv run` 방식 (다단계 프로세스)
- 목표: `uvx` 방식으로 프로세스 최소화
- 예상 개선: 프로세스 체인 3단계 → 1-2단계 (약 30-40% 성능 개선)

#### C2: 역호환성
- MoAI-ADK 0.25.10 이상에서 지원
- 기존 `uv run` 방식도 사용 가능하지만 권장되지 않음

#### C3: 패키지 배포
- PyPI의 moai-adk 패키지에 새로운 CLI 진입점 포함
- 설치된 패키지는 `uvx moai-adk statusline` 즉시 실행 가능

---

## 명세 (Specifications)

### ubiquitous (항상 참)
- **SPEC**: 시스템은 항상 Claude Code의 JSON 컨텍스트를 stdin으로 수신하여 statusline 렌더링
- **구현**: `main.py`의 `read_session_context()` 함수가 stdin에서 JSON 읽음

### event-driven (이벤트 기반)
- **SPEC**: Claude Code 세션 시작 시 `.claude/settings.json`의 statusLine 명령어 실행
- **현재**: `uv run python -m moai_adk.statusline.main` 호출
- **변경 후**: `uvx moai-adk statusline` 호출
- **동작**:
  ```json
  {
    "statusLine": {
      "type": "command",
      "command": "uvx moai-adk statusline",
      "padding": 0
    }
  }
  ```

### unwanted (원치 않는 동작)
- **문제**: uvx 명령어가 설치되지 않은 환경에서 statusline 실행 실패
- **방지**: 사전 조건 확인 또는 폴백 메커니즘 구현
- **구현 방식**:
  - 워퍼 스크립트 (`.claude/bin/statusline.sh` 또는 `.py`) 작성
  - 워퍼는 uvx 사용 가능 여부 확인 후 폴백

### state-driven (상태 기반)
- **SPEC**: uvx 캐시가 활성 상태일 때, statusline 실행 성능 최적화
- **구현**: uvx는 자동으로 도구 캐시 유지
- **추적**: 캐시 효율성은 `~/.cache/uv` (Linux/macOS) 또는 `%LOCALAPPDATA%\uv` (Windows)에서 확인 가능

### optional (선택적)
- **기능**: 사용자가 로컬 개발 시 `uv run` 방식으로 직접 실행 가능
- **사용**: `uv run python -m moai_adk.statusline.main`
- **목적**: 디버깅, 개발, 또는 uvx 미설치 환경에서의 임시 해결

---

## 추적 (Traceability)

### 태그 매핑
- `@TAG(R1.1)`: `pyproject.toml` 진입점 추가 → `test_statusline_uvx_entry_point()`
- `@TAG(R1.2)`: 모든 OS 테스트 → `test_statusline_windows()`, `test_statusline_posix()`
- `@TAG(R1.3)`: `.claude/settings.json` 수정 → `test_settings_json_update()`
- `@TAG(R1.4)`: 템플릿 동기화 → `test_template_sync()`
- `@TAG(C1)`: Windows 성능 테스트 → `benchmark_process_overhead()`

### 의존성
- `SPEC-STATUSLINE-UVX-001` → `pyproject.toml` (CLI 진입점 추가)
- `SPEC-STATUSLINE-UVX-001` → `.claude/settings.json` (명령어 업데이트)
- `SPEC-STATUSLINE-UVX-001` → `src/moai_adk/templates/.claude/settings.json` (템플릿 동기화)

---

## 기술 고려사항

### uvx vs uv run 비교

| 항목 | uv run | uvx |
|------|--------|-----|
| 명령어 | `uv run python -m module` | `uvx tool-name` |
| 프로세스 수 | 3단계 (uv → python → module) | 1-2단계 (캐시 시) |
| 캐시 | 자동 | 자동 |
| OS 호환성 | 모든 OS | 모든 OS |
| 성능 | 느림 (특히 Windows) | 빠름 |
| 복잡도 | 간단 | 간단 |

### 다단계 폴백 아키텍처
```
Claude Code statusline command
  ↓
uvx moai-adk statusline (최우선)
  ↓ [실패 시]
uv run python -m moai_adk.statusline.main (폴백)
  ↓ [실패 시]
직접 python 호출 (최후 수단)
```

---

## 성능 영향 분석

### Windows 프로세스 오버헤드 (예상)
- **현재**: 프로세스 체인 3단계 (uv.exe → python.exe → statusline 로직)
- **개선 후**: uvx 캐시 히트 시 1단계, 미스 시 2단계
- **예상 개선**: 30-40% 성능 향상

### 캐시 전략
- **첫 실행**: uvx가 moai-adk 다운로드 및 캐시 (약 2-3초)
- **후속 실행**: 캐시에서 직접 로드 (약 100-200ms)
- **캐시 무효화**: moai-adk 버전 업데이트 시만 필요

---

## 마이그레이션 전략

### 단계별 적용

#### Phase 1: MoAI-ADK 코드 변경 (영향도: 낮음)
1. `pyproject.toml`에 새로운 CLI 진입점 추가
2. `.claude/settings.json` (로컬) 업데이트
3. 단위 테스트 작성

#### Phase 2: 템플릿 동기화 (영향도: 중간)
1. `src/moai_adk/templates/.claude/settings.json` 업데이트
2. 신규 프로젝트는 uvx 방식으로 자동 설정
3. 기존 프로젝트는 수동 마이그레이션 필요 (선택적)

#### Phase 3: 폴백 메커니즘 (향후, 영향도: 낮음)
1. 워퍼 스크립트 작성 (선택적)
2. UV_NO_CACHE 환경변수 등으로 캐시 제어

---

## 위험 및 완화 방안

### 위험 R1: uvx 미설치 환경
- **가능성**: 낮음 (uv를 설치했으면 uvx 포함)
- **영향도**: 높음 (statusline 작동 불가)
- **완화**: 폴백 메커니즘 구현 또는 명확한 오류 메시지 제공

### 위험 R2: 캐시 오염
- **가능성**: 낮음 (uvx 자체 안정성 높음)
- **영향도**: 중간 (캐시 재생성 필요)
- **완화**: `UV_CACHE_DIR` 환경변수로 캐시 위치 명시

### 위험 R3: 템플릿 동기화 미확인
- **가능성**: 중간 (신규 프로젝트에서만 영향)
- **영향도**: 낮음 (uvx 미지원 환경에서는 수동 설정)
- **완화**: 마이그레이션 가이드 문서 작성

---

