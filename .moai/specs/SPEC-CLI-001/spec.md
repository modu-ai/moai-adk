---
id: CLI-001
version: 0.1.0
status: completed
created: 2025-10-15
updated: 2025-10-15
priority: high
category: feature
labels:
  - cli
  - usability
  - diagnostics
depends_on: []
blocks: []
related_specs:
  - TEST-COVERAGE-001
scope:
  packages:
    - src/moai_adk/cli/commands
    - src/moai_adk/core/project
  files:
    - doctor.py
    - status.py
    - restore.py
    - checker.py
---


## HISTORY

### v0.1.0 (2025-10-15)
- **COMPLETED**: doctor 명령어 고도화 (20개 언어 도구 체인 검증) 구현 완료
- **REVIEW**: N/A (Personal implementation)
- **CHANGES**:
  - Phase 1: 언어별 도구 체인 매핑 (LANGUAGE_TOOLS 상수, checker.py 431 LOC)
  - Phase 2: doctor --verbose, --fix, --export, --check 옵션 추가
  - 50개 테스트 100% 통과, 커버리지 91.58% (doctor.py)
  - 지원 언어: Python, TypeScript, JavaScript, Java, Go, Rust, Dart, Swift, Kotlin, C#, PHP, Ruby, Elixir, Scala, Clojure, Haskell, C, C++, Lua, OCaml (20개)
- **RELATED**:
  - RED 커밋: 4568654 (doctor 언어별 도구 체인 검증 테스트 추가)
  - GREEN 커밋: bc0074a (20개 언어 도구 체인 검증 구현 완료)

### v0.0.1 (2025-10-15)
- **INITIAL**: CLI 명령어 고도화 명세 최초 작성
- **SCOPE**: doctor/status/restore 명령어 기능 확장
- **CONTEXT**: 85.61% 테스트 커버리지 달성 후 사용성 개선 단계 진입
- **MOTIVATION**:
  - 현재 doctor는 4개 기본 체크만 수행 (Python, Git, .moai, config.json)
  - 20개 언어별 도구 체인 검증 부재
  - 문제 발견 시 자동 수정 제안 및 가이드 부재
  - status는 정적 정보만 표시 (TAG 체인 무결성, 코드 품질 지표 부재)
  - restore는 기본 복원만 지원 (선택적 복원, 롤백 히스토리 부재)

## Environment (환경)

- **Python 버전**: 3.13.1
- **CLI 프레임워크**: Click 8.1.x
- **Rich 라이브러리**: 13.7.x (터미널 UI)
- **현재 CLI 명령어**: doctor, status, restore, init, update, backup
- **지원 언어**: 20개 (Python, TypeScript, JavaScript, Java, Go, Rust, Dart, Swift, Kotlin, C#, PHP, Ruby, Elixir, Scala, Clojure, Haskell, C, C++, Lua, OCaml)

## Assumptions (가정)

- Click 프레임워크의 옵션 확장이 가능하다 (`--verbose`, `--fix`, `--detail`, `--dry-run`)
- `src/moai_adk/core/template/languages.py`의 `LANGUAGE_TEMPLATES` 매핑을 활용한다
- 진단 실행 시간은 5초 이내로 제한한다 (사용자 경험)
- 모든 CLI 개선은 기존 테스트 커버리지 85% 이상을 유지한다
- 자동 수정 기능은 사용자 명시적 승인 후에만 실행한다

## Requirements (요구사항)

### Ubiquitous (필수 요구사항)

- 시스템은 20개 언어별 도구 체인 진단을 제공해야 한다
- 시스템은 doctor/status/restore 명령어에 상세 옵션을 제공해야 한다
- 시스템은 진단 결과에 대한 실행 가능한 가이드를 제공해야 한다
- 시스템은 모든 CLI 출력에 Rich 라이브러리를 사용하여 가독성을 보장해야 한다

### Event-driven (이벤트 기반 요구사항)

- WHEN `moai doctor` 실행 시, 시스템은 현재 프로젝트 언어를 감지하고 필수 도구를 검증해야 한다
- WHEN `moai doctor --verbose` 실행 시, 시스템은 모든 선택 도구 및 버전 정보를 표시해야 한다
- WHEN `moai doctor --fix` 실행 시, 시스템은 누락된 도구에 대한 설치 명령어를 제안하고 사용자 승인 후 실행해야 한다
- WHEN `moai status --detail` 실행 시, 시스템은 TAG 체인 무결성, 테스트 커버리지, 코드 품질 지표를 추가 표시해야 한다
- WHEN `moai restore --list` 실행 시, 시스템은 사용 가능한 모든 백업 목록과 메타데이터를 표시해야 한다
- WHEN `moai restore --dry-run <backup-id>` 실행 시, 시스템은 복원될 파일 목록과 변경 사항을 미리 보여야 한다
- WHEN 진단 중 치명적 문제 발견 시, 시스템은 빨간색 경고와 함께 해결 가이드를 표시해야 한다

### State-driven (상태 기반 요구사항)

- WHILE 프로젝트가 초기화되지 않았을 때 (.moai 미존재), status는 `moai init .` 실행 안내를 표시해야 한다
- WHILE Python 프로젝트일 때, doctor는 pytest, mypy, ruff 설치 여부를 확인해야 한다
- WHILE TypeScript 프로젝트일 때, doctor는 Vitest, Biome 설치 여부를 확인해야 한다
- WHILE 백업이 존재하지 않을 때, restore는 `moai backup` 실행 안내를 표시해야 한다
- WHILE 진단이 5초를 초과할 때, 시스템은 진행 상황 표시줄을 표시해야 한다

### Optional (선택적 기능)

- WHERE `--json` 옵션이 제공되면, 시스템은 모든 출력을 JSON 형식으로 반환할 수 있다
- WHERE CI/CD 환경에서 실행 시, 시스템은 Rich UI를 비활성화하고 플레인 텍스트로 출력할 수 있다
- WHERE `--export <file>` 옵션이 제공되면, doctor는 진단 결과를 파일로 저장할 수 있다
- WHERE 프로젝트가 Team 모드일 때, status는 GitHub PR 상태를 추가로 표시할 수 있다

### Constraints (제약사항)

- 진단 실행 시간은 5초를 초과하지 않아야 한다
- IF 자동 수정 기능 사용 시, 시스템은 반드시 사용자 승인을 받아야 한다
- IF 언어별 도구가 여러 대안이 있을 경우 (예: pytest/unittest), 시스템은 권장 도구를 우선 표시해야 한다
- 모든 새로운 CLI 기능은 85% 이상의 테스트 커버리지를 유지해야 한다
- doctor/status/restore 명령어는 오프라인 환경에서도 작동해야 한다 (네트워크 의존성 최소화)

## Implementation Details (구현 세부사항)

### Phase 1: doctor 고도화 (+50% 진단 범위)

**목표**: 기본 4개 체크 → 언어별 도구 체인 검증

1. **언어 감지 및 도구 매핑**:
   - `core/project/detector.py`의 언어 감지 로직 활용
   - `core/template/languages.py`의 `LANGUAGE_TEMPLATES` 기반 도구 매핑
   - 언어별 필수/선택 도구 정의:
     ```python
     LANGUAGE_TOOLS = {
         "python": {
             "required": ["python3", "pip"],
             "recommended": ["pytest", "mypy", "ruff"],
             "optional": ["black", "pylint"]
         },
         "typescript": {
             "required": ["node", "npm"],
             "recommended": ["vitest", "biome"],
             "optional": ["typescript", "eslint"]
         },
         # ... 20개 언어 전체
     }
     ```

2. **doctor 명령어 옵션 확장**:
   - `--verbose`: 모든 도구 버전 정보 표시
   - `--fix`: 누락 도구 자동 설치 제안
   - `--export <file>`: 진단 결과 JSON 저장
   - `--check <tool>`: 특정 도구만 검증

3. **진단 결과 UI 개선**:
   - Rich Table: 도구명, 버전, 상태, 권장사항
   - 색상 코딩: ✓ (초록), ⚠ (노랑), ✗ (빨강)
   - 실행 가능 명령어 표시: `pip install pytest`

### Phase 2: status 고도화 (+3개 품질 지표)

**목표**: 정적 정보 → 동적 품질 지표 추가

1. **TAG 체인 무결성 검증**:
   - `rg '@(SPEC|TEST|CODE|DOC):' -n` 실행
   - 고아 TAG 탐지: SPEC은 있는데 CODE 없음
   - 끊어진 체인 탐지: CODE는 있는데 SPEC 없음
   - 결과 요약: "TAG 체인: 100% (0 orphans, 0 broken)"

2. **테스트 커버리지 표시**:
   - `pytest --cov --cov-report=json` 결과 파싱
   - 커버리지 백분율 및 목표 대비 상태
   - 결과: "Test Coverage: 85.61% ✓ (goal: 85%)"

3. **코드 품질 지표**:
   - 린터 경고 개수 (ruff, biome 등)
   - 복잡도 임계값 초과 함수 개수
   - 결과: "Code Quality: 0 warnings, 0 complexity issues"

4. **status 명령어 옵션**:
   - `--detail`: TAG 체인 + 커버리지 + 품질 지표 표시
   - `--json`: JSON 형식 출력 (CI/CD 통합용)

### Phase 3: restore 고도화 (+선택적 복원)

**목표**: 전체 복원 → 선택적 복원 + 미리보기

1. **백업 메타데이터 확장**:
   - 현재: 타임스탬프만 기록
   - 개선: 백업 ID, 타임스탬프, 파일 목록, 변경 사항
   - 포맷: `.moai/backups/latest.json` (메타데이터), `.moai-backups/<backup-id>/` (백업 내용)

2. **restore 명령어 옵션**:
   - `--list`: 사용 가능한 백업 목록 표시
   - `--dry-run <backup-id>`: 복원 미리보기 (변경 사항 표시)
   - `--select <files>`: 특정 파일만 선택적 복원
   - `--compare <backup-id>`: 현재 상태와 백업 비교

3. **복원 안전장치**:
   - Git dirty state 감지: 미커밋 변경 있으면 경고
   - 백업 전 자동 스냅샷 생성
   - 복원 후 롤백 기능 제공

## Acceptance Criteria (인수 기준)

### AC-1: doctor --verbose 언어별 도구 진단

- **Given**: Python 프로젝트에서 pytest는 설치되어 있지만 mypy는 설치되지 않았을 때
- **When**: `moai doctor --verbose` 실행
- **Then**:
  - Python 프로젝트 감지 표시
  - pytest ✓ (버전 8.4.2) 표시
  - mypy ✗ (not installed) 표시
  - 설치 명령어 제안: `pip install mypy`
  - 전체 진단 시간 < 5초

### AC-2: doctor --fix 자동 수정 제안

- **Given**: TypeScript 프로젝트에서 Vitest가 설치되지 않았을 때
- **When**: `moai doctor --fix` 실행
- **Then**:
  - Vitest 미설치 감지
  - 설치 명령어 표시: `npm install -D vitest`
  - 사용자 승인 요청: "Install Vitest? (y/n)"
  - "y" 선택 시 설치 실행 및 결과 표시

### AC-3: status --detail 품질 지표 표시

- **Given**: 프로젝트에 1개 SPEC, 85.61% 커버리지, 0개 린터 경고가 있을 때
- **When**: `moai status --detail` 실행
- **Then**:
  - 기본 정보 표시 (Mode, Locale, SPECs, Branch, Git Status)
  - TAG 체인 무결성: "100% (0 orphans, 0 broken)"
  - 테스트 커버리지: "85.61% ✓ (goal: 85%)"
  - 코드 품질: "0 warnings, 0 complexity issues"

### AC-4: restore --list 백업 목록 표시

- **Given**: 3개 백업이 존재할 때
- **When**: `moai restore --list` 실행
- **Then**:
  - 백업 ID, 타임스탬프, 파일 개수가 테이블로 표시
  - 최신 백업이 상단에 표시 (역순)
  - 각 백업의 용량 표시

### AC-5: restore --dry-run 복원 미리보기

- **Given**: backup-001이 존재하고 현재 config.json이 수정된 상태일 때
- **When**: `moai restore --dry-run backup-001` 실행
- **Then**:
  - 복원될 파일 목록 표시
  - config.json 변경 사항 diff 표시
  - "실제 복원하려면 `moai restore backup-001` 실행" 안내

### AC-6: doctor 오프라인 동작

- **Given**: 인터넷 연결 없이 로컬 환경일 때
- **When**: `moai doctor` 실행
- **Then**:
  - 로컬 도구 체크 정상 수행
  - 네트워크 에러 없이 완료
  - 온라인 기능 (버전 체크) 스킵 표시

### AC-7: status CI/CD JSON 출력

- **Given**: CI/CD 환경에서 실행할 때
- **When**: `moai status --json` 실행
- **Then**:
  - JSON 형식으로 모든 정보 출력
  - Rich UI 비활성화
  - 파싱 가능한 구조화된 데이터

### AC-8: doctor 진행 상황 표시

- **Given**: 20개 언어 도구 검증이 5초 이상 소요될 때
- **When**: `moai doctor --verbose` 실행
- **Then**:
  - 진행 상황 표시줄 표시
  - 현재 검사 중인 도구 표시
  - 완료 백분율 업데이트

### AC-9: restore 선택적 복원

- **Given**: backup-001에 10개 파일이 있을 때
- **When**: `moai restore backup-001 --select config.json,product.md` 실행
- **Then**:
  - 지정된 2개 파일만 복원
  - 나머지 8개 파일은 건드리지 않음
  - 복원 완료 메시지 표시

### AC-10: doctor 언어 미감지 처리

- **Given**: 언어를 감지할 수 없는 프로젝트일 때
- **When**: `moai doctor` 실행
- **Then**:
  - "Language: Unknown" 표시
  - 기본 도구만 검증 (Python, Git)
  - 언어 설정 가이드 표시: `.moai/config.json` 수동 설정

### AC-11: status TAG 체인 끊어짐 감지

- **When**: `moai status --detail` 실행
- **Then**:
  - TAG 체인 무결성: "⚠ 1 broken chain"
  - 끊어진 TAG 목록 표시: "AUTH-001: CODE exists but SPEC missing"
  - 수정 가이드: `/alfred:1-plan AUTH-001` 실행 안내

### AC-12: restore Git dirty state 경고

- **Given**: Git에 미커밋 변경사항이 있을 때
- **When**: `moai restore backup-001` 실행
- **Then**:
  - 경고 표시: "⚠ Uncommitted changes detected"
  - Git status 요약 표시
  - 계속 진행 여부 확인: "Proceed? (y/n)"


- **TEST**:
  - `tests/unit/test_doctor.py`
  - `tests/unit/test_status.py`
  - `tests/unit/test_restore.py`
  - `tests/integration/test_cli_advanced.py`
- **CODE**:
  - `src/moai_adk/cli/commands/doctor.py`
  - `src/moai_adk/cli/commands/status.py`
  - `src/moai_adk/cli/commands/restore.py`
  - `src/moai_adk/core/project/checker.py`
- **DOC**: `README.md` (CLI Reference 섹션)

## References

- **TRUST 원칙**: `.moai/memory/development-guide.md#trust-5원칙`
- **언어 지원**: `src/moai_adk/core/template/languages.py`
- **Click 프레임워크**: https://click.palletsprojects.com/
- **Rich 라이브러리**: https://rich.readthedocs.io/
- **관련 SPEC**: `SPEC-TEST-COVERAGE-001` (테스트 커버리지 85% 기준)

## Non-Goals (범위 외)

- **GUI 인터페이스**: CLI만 지원, GUI는 향후 고려
- **원격 진단**: 로컬 환경만 지원, 원격 서버 진단은 별도 SPEC
- **플러그인 시스템**: 커스텀 진단 플러그인은 향후 고려
- **자동 업데이트**: doctor는 진단만 수행, 도구 자동 업데이트는 사용자 판단

## Migration Plan (기존 사용자 영향)

**기존 명령어 호환성**: 100% 유지
- `moai doctor` (옵션 없음): 기존과 동일하게 동작
- `moai status` (옵션 없음): 기존과 동일하게 동작
- `moai restore <backup-id>`: 기존과 동일하게 동작

**새로운 옵션**: 선택적 사용
- 모든 새 옵션 (`--verbose`, `--detail`, `--dry-run`)은 선택적
- 기존 사용자 워크플로우에 영향 없음

**문서 업데이트**: README.md CLI Reference 섹션에 새 옵션 추가
