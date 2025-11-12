# 인수 기준: @SPEC:CMD-IMPROVE-001

**SPEC ID**: CMD-IMPROVE-001
**Title**: Commands 레이어 컨텍스트 전달 및 Resume 기능 통합 개선
**Author**: @goos
**Created**: 2025-11-12
**Priority**: HIGH

---

## 개요 (Overview)

본 문서는 SPEC-CMD-IMPROVE-001의 인수 기준을 정의합니다. Given-When-Then 형식의 시나리오를 통해 구현 완료 여부를 검증합니다.

**핵심 검증 항목**:
1. 명시적 컨텍스트 전달 시스템 동작 확인
2. Resume 기능 정상 작동 확인
3. 오류 상황 처리 확인
4. 사용자 경험 개선 확인

---

## Phase 1: 명시적 컨텍스트 전달 시스템

### Scenario 1.1: Phase 결과 저장 및 로드

#### 테스트 케이스 1.1.1: 0-project 결과 저장

**Given**: 사용자가 새 프로젝트를 초기화함
- Command: `/alfred:0-project`
- 프로젝트명: TestProject
- 모드: personal
- 언어: ko

**When**: 0-project 명령이 성공적으로 완료됨

**Then**: 시스템은 다음을 수행해야 함
1. `.moai/memory/command-state/0-project-{timestamp}.json` 파일 생성
2. JSON 파일에 다음 정보 포함:
   - `phase`: "0-project"
   - `status`: "completed"
   - `outputs.project_name`: "TestProject"
   - `outputs.mode`: "personal"
   - `outputs.conversation_language`: "ko"
   - `files_created`: 절대 경로 배열
3. 파일 권한: 읽기/쓰기 가능 (644)

**검증 명령**:
```bash
# 파일 존재 확인
ls -la .moai/memory/command-state/0-project-*.json

# JSON 구조 검증
jq '.phase == "0-project" and .status == "completed"' .moai/memory/command-state/0-project-*.json

# 출력 데이터 검증
jq '.outputs | has("project_name") and has("mode")' .moai/memory/command-state/0-project-*.json
```

---

#### 테스트 케이스 1.1.2: 1-plan이 0-project 결과 로드

**Given**: 0-project가 완료되어 JSON 파일이 생성됨
- 파일 경로: `.moai/memory/command-state/0-project-20251112.json`
- 저장된 데이터: `project_name: TestProject`, `mode: personal`

**When**: 사용자가 `/alfred:1-plan "사용자 인증"` 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. `0-project-*.json` 파일 자동 로드
2. 템플릿 변수 치환:
   - `{{PROJECT_NAME}}` → "TestProject"
   - `{{MODE}}` → "personal"
3. plan-agent 호출 시 치환된 컨텍스트를 prompt에 포함
4. `1-plan-SPEC-XXX-{timestamp}.json` 파일 생성

**검증 명령**:
```bash
# 1-plan 결과 파일 존재 확인
ls -la .moai/memory/command-state/1-plan-*.json

# 컨텍스트 참조 검증 (Agent 호출 로그에서 확인)
# {{PROJECT_NAME}} 문자열이 "TestProject"로 치환되었는지 확인
```

---

### Scenario 1.2: 절대 경로 변환 및 검증

#### 테스트 케이스 1.2.1: 상대 경로를 절대 경로로 변환

**Given**: Phase 결과에 상대 경로가 포함됨
- 상대 경로: `.moai/project/product.md`
- Project root: `/Users/goos/MoAI/TestProject`

**When**: 다음 Phase에서 해당 파일을 참조함

**Then**: 시스템은 다음을 수행해야 함
1. 상대 경로를 절대 경로로 변환: `/Users/goos/MoAI/TestProject/.moai/project/product.md`
2. 파일 존재 여부 검증
3. Project root 외부 경로 접근 시도 차단

**검증 코드**:
```python
from moai_adk.core.path_validator import validate_and_convert_path

project_root = "/Users/goos/MoAI/TestProject"
relative_path = ".moai/project/product.md"

abs_path = validate_and_convert_path(relative_path, project_root)
assert abs_path == "/Users/goos/MoAI/TestProject/.moai/project/product.md"
assert abs_path.startswith(project_root)
```

---

#### 테스트 케이스 1.2.2: 존재하지 않는 경로 처리

**Given**: Agent가 존재하지 않는 파일 경로를 참조함
- 경로: `.moai/project/nonexistent.md`

**When**: 경로 검증 함수가 호출됨

**Then**: 시스템은 다음을 수행해야 함
1. 파일이 존재하지 않음을 감지
2. 부모 디렉토리 존재 여부 확인
3. 부모 디렉토리가 존재하면 경로 허용 (생성 예정 파일)
4. 부모 디렉토리가 없으면 명확한 에러 메시지 출력

**검증 코드**:
```python
import pytest
from moai_adk.core.path_validator import validate_and_convert_path

project_root = "/Users/goos/MoAI/TestProject"
invalid_path = "nonexistent_dir/file.md"

with pytest.raises(FileNotFoundError) as exc_info:
    validate_and_convert_path(invalid_path, project_root, must_exist=True)

assert "Parent directory not found" in str(exc_info.value)
```

---

### Scenario 1.3: 템플릿 변수 치환

#### 테스트 케이스 1.3.1: 단일 변수 치환

**Given**: Agent prompt에 템플릿 변수가 포함됨
- Prompt: "프로젝트 {{PROJECT_NAME}}의 SPEC을 생성합니다."
- 컨텍스트: `{"PROJECT_NAME": "TestProject"}`

**When**: 템플릿 엔진이 변수를 치환함

**Then**: 시스템은 다음을 수행해야 함
1. `{{PROJECT_NAME}}`을 "TestProject"로 치환
2. 결과 문자열에 미치환 변수가 없음을 확인

**검증 코드**:
```python
from moai_adk.core.template_engine import replace_template_vars

prompt = "프로젝트 {{PROJECT_NAME}}의 SPEC을 생성합니다."
context = {"PROJECT_NAME": "TestProject"}

result = replace_template_vars(prompt, context)
assert result == "프로젝트 TestProject의 SPEC을 생성합니다."
assert "{{" not in result
```

---

#### 테스트 케이스 1.3.2: 미치환 변수 검증

**Given**: Agent prompt에 치환되지 않은 변수가 남아있음
- Prompt: "프로젝트 {{PROJECT_NAME}}의 {{UNDEFINED_VAR}}을 처리합니다."
- 컨텍스트: `{"PROJECT_NAME": "TestProject"}`

**When**: 검증 함수가 호출됨

**Then**: 시스템은 다음을 수행해야 함
1. 미치환 변수 `{{UNDEFINED_VAR}}` 검출
2. 명확한 에러 메시지 출력: "Unsubstituted template variables: ['{{UNDEFINED_VAR}}']"
3. Agent 호출 중단

**검증 코드**:
```python
import pytest
from moai_adk.core.template_engine import validate_no_unsubstituted_vars

prompt_with_unsubstituted = "프로젝트 TestProject의 {{UNDEFINED_VAR}}을 처리합니다."

with pytest.raises(ValueError) as exc_info:
    validate_no_unsubstituted_vars(prompt_with_unsubstituted)

assert "Unsubstituted template variables" in str(exc_info.value)
assert "UNDEFINED_VAR" in str(exc_info.value)
```

---

## Phase 2: Resume 기능

### Scenario 2.1: Resume 상태 저장 및 로드

#### 테스트 케이스 2.1.1: 2-run 중단 시 상태 저장

**Given**: 사용자가 `/alfred:2-run SPEC-AUTH-001` 실행 중
- 완료된 단계: spec_validation, test_setup, red_phase
- 대기 중인 단계: green_phase, refactor_phase, integration_test

**When**: 사용자가 작업을 중단함 (Ctrl+C 또는 세션 종료)

**Then**: 시스템은 다음을 수행해야 함
1. `.moai/memory/command-state/2-run-SPEC-AUTH-001-{timestamp}.json` 파일 생성
2. JSON 파일에 다음 정보 포함:
   - `command`: "alfred:2-run"
   - `spec_id`: "SPEC-AUTH-001"
   - `current_phase`: "implementation"
   - `completed_steps`: ["spec_validation", "test_setup", "red_phase"]
   - `pending_steps`: ["green_phase", "refactor_phase", "integration_test"]
   - `timestamp`: 중단 시각
   - `expiry`: 중단 시각 + 30일
3. 파일이 원자적으로 저장됨 (atomic write)

**검증 명령**:
```bash
# Resume 상태 파일 존재 확인
ls -la .moai/memory/command-state/2-run-SPEC-AUTH-001-*.json

# 저장된 단계 검증
jq '.completed_steps | length == 3' .moai/memory/command-state/2-run-SPEC-AUTH-001-*.json
jq '.pending_steps | length == 3' .moai/memory/command-state/2-run-SPEC-AUTH-001-*.json
```

---

#### 테스트 케이스 2.1.2: Resume 명령으로 재개

**Given**: 저장된 Resume 상태가 존재함
- 파일: `.moai/memory/command-state/2-run-SPEC-AUTH-001-20251112.json`
- 완료된 단계: spec_validation, test_setup, red_phase
- 대기 중인 단계: green_phase, refactor_phase, integration_test

**When**: 사용자가 `/alfred:resume 2-run SPEC-AUTH-001` 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. Resume 상태 파일 자동 검색 및 로드
2. Timestamp 유효성 검증 (30일 이내)
3. 사용자에게 재개 확인 메시지 출력
4. 승인 시 `pending_steps`의 첫 번째 단계(green_phase)부터 실행
5. 완료된 단계는 스킵하고 로그에 기록

**검증 로그 예시**:
```
[INFO] Resume 상태 로드 완료: SPEC-AUTH-001 (저장 시각: 2025-11-12 12:00)
[INFO] 완료된 단계 스킵: spec_validation, test_setup, red_phase
[INFO] green_phase부터 재개합니다...
```

---

### Scenario 2.2: Timestamp 유효성 검증

#### 테스트 케이스 2.2.1: 유효한 상태 (30일 이내)

**Given**: Resume 상태가 20일 전에 저장됨
- 저장 시각: 2025-10-23 12:00
- 현재 시각: 2025-11-12 12:00
- 만료 시각: 2025-11-22 12:00

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. 경과 시간 계산: 20일
2. 만료 시각과 비교: 20일 < 30일 (유효)
3. Resume 진행 허용
4. 경고 메시지 없음

**검증 코드**:
```python
from datetime import datetime, timedelta
from moai_adk.core.resume_handler import validate_timestamp

state = {
    "timestamp": "2025-10-23T12:00:00Z",
    "expiry": "2025-11-22T12:00:00Z"
}

current_time = datetime(2025, 11, 12, 12, 0, 0)
assert validate_timestamp(state, current_time) is True
```

---

#### 테스트 케이스 2.2.2: 만료된 상태 (30일 초과)

**Given**: Resume 상태가 35일 전에 저장됨
- 저장 시각: 2025-10-08 12:00
- 현재 시각: 2025-11-12 12:00
- 만료 시각: 2025-11-07 12:00

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. 경과 시간 계산: 35일
2. 만료 시각과 비교: 35일 > 30일 (만료)
3. 명확한 에러 메시지 출력:
   ```
   ERROR: Resume 상태가 만료되었습니다.
   - 저장 시각: 2025-10-08 12:00
   - 만료 시각: 2025-11-07 12:00
   - 권장 조치: 처음부터 재시작하십시오.
   ```
4. Resume 진행 차단
5. 만료된 상태 파일 자동 삭제 (선택 사항)

**검증 코드**:
```python
import pytest
from datetime import datetime
from moai_adk.core.resume_handler import validate_timestamp

state = {
    "timestamp": "2025-10-08T12:00:00Z",
    "expiry": "2025-11-07T12:00:00Z"
}

current_time = datetime(2025, 11, 12, 12, 0, 0)

with pytest.raises(ValueError) as exc_info:
    validate_timestamp(state, current_time, raise_on_expired=True)

assert "만료되었습니다" in str(exc_info.value)
```

---

### Scenario 2.3: 단계별 재개 메커니즘

#### 테스트 케이스 2.3.1: 완료된 단계 스킵

**Given**: Resume 상태에 완료된 단계가 기록됨
- 완료된 단계: spec_validation, test_setup, red_phase

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. 완료된 단계를 순회하며 각 단계 스킵
2. 로그에 스킵된 단계 기록:
   ```
   [SKIP] spec_validation (already completed)
   [SKIP] test_setup (already completed)
   [SKIP] red_phase (already completed)
   ```
3. `pending_steps`의 첫 번째 단계로 이동
4. 첫 번째 대기 단계(green_phase) 실행

**검증 방법**:
- 로그 출력 확인
- 각 단계의 실행 시간 측정 (스킵된 단계는 0초)

---

#### 테스트 케이스 2.3.2: 대기 중인 단계 실행

**Given**: Resume 상태에 대기 중인 단계가 기록됨
- 대기 중인 단계: green_phase, refactor_phase, integration_test

**When**: 완료된 단계를 스킵한 후 재개함

**Then**: 시스템은 다음을 수행해야 함
1. 첫 번째 대기 단계(green_phase) 실행
2. green_phase 완료 시 `completed_steps`에 추가
3. `pending_steps`에서 green_phase 제거
4. 다음 단계(refactor_phase) 자동 진행
5. 각 단계 완료 시 상태 파일 업데이트

**검증 명령**:
```bash
# 단계 완료 후 상태 파일 확인
jq '.completed_steps | contains(["green_phase"])' .moai/memory/command-state/2-run-SPEC-AUTH-001-*.json
jq '.pending_steps | contains(["refactor_phase", "integration_test"])' .moai/memory/command-state/2-run-SPEC-AUTH-001-*.json
```

---

### Scenario 2.4: Resume 옵션 (선택적 기능)

#### 테스트 케이스 2.4.1: 특정 단계부터 재시작

**Given**: Resume 상태가 존재하지만 특정 단계부터 재시작하고 싶음
- 완료된 단계: spec_validation, test_setup, red_phase

**When**: 사용자가 `/alfred:resume 2-run SPEC-AUTH-001 --from=integration_test` 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. Resume 상태 로드
2. `--from` 옵션 파싱: integration_test
3. integration_test 이전 단계들을 모두 완료된 것으로 표시
4. integration_test부터 실행 시작

**검증 방법**:
- 로그에서 강제 재시작 메시지 확인
- integration_test 이전 단계 실행 시간 0초 확인

---

#### 테스트 케이스 2.4.2: 처음부터 재실행

**Given**: Resume 상태가 존재하지만 처음부터 재실행하고 싶음

**When**: 사용자가 `/alfred:resume 2-run SPEC-AUTH-001 --restart` 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. Resume 상태 로드 (참고용)
2. `--restart` 옵션 감지
3. 저장된 상태 무시
4. 첫 번째 단계(spec_validation)부터 실행
5. 기존 Resume 상태 파일 백업 또는 삭제

**검증 방법**:
- 모든 단계가 처음부터 실행되는지 확인
- 실행 로그에 "Ignoring saved state" 메시지 확인

---

## Phase 3: 오류 상황 처리

### Scenario 3.1: JSON 저장 실패

#### 테스트 케이스 3.1.1: 디스크 공간 부족

**Given**: Phase 완료 후 JSON 저장 시도
- 디스크 공간: 0 MB 남음

**When**: 상태 저장 함수를 호출함

**Then**: 시스템은 다음을 수행해야 함
1. 디스크 공간 부족 감지
2. 명확한 에러 메시지 출력:
   ```
   ERROR: 디스크 공간이 부족하여 상태를 저장할 수 없습니다.
   - 필요 공간: 10 KB
   - 사용 가능 공간: 0 KB
   - 권장 조치: 불필요한 파일을 삭제하고 다시 시도하십시오.
   ```
3. 임시 메모리에 상태 백업 유지
4. 사용자에게 재시도 옵션 제공

**검증 코드**:
```python
import pytest
from unittest.mock import patch
from moai_adk.core.context_manager import save_phase_result

with patch('os.replace', side_effect=OSError("No space left on device")):
    result = save_phase_result("0-project", {"test": "data"})
    assert result.status == "save_failed"
    assert result.backup_available is True
```

---

#### 테스트 케이스 3.1.2: 권한 오류

**Given**: `.moai/memory/command-state/` 디렉토리에 쓰기 권한 없음

**When**: 상태 저장 함수를 호출함

**Then**: 시스템은 다음을 수행해야 함
1. 권한 오류 감지
2. 명확한 에러 메시지 출력:
   ```
   ERROR: 상태 파일을 저장할 권한이 없습니다.
   - 디렉토리: .moai/memory/command-state/
   - 권장 조치: chmod 755 .moai/memory/command-state/
   ```
3. 작업 중단

**검증 명령**:
```bash
# 권한 제거 후 테스트
chmod 000 .moai/memory/command-state/

# 저장 시도 (실패 예상)
python -c "from moai_adk.core.context_manager import save_phase_result; save_phase_result('test', {})"

# 권한 복구
chmod 755 .moai/memory/command-state/
```

---

### Scenario 3.2: Resume 상태 파일 손상

#### 테스트 케이스 3.2.1: 잘못된 JSON 형식

**Given**: Resume 상태 파일이 손상됨
- 파일 내용: `{"command": "alfred:2-run", "spec_id": ...` (닫히지 않은 중괄호)

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. JSON 파싱 오류 감지
2. 명확한 에러 메시지 출력:
   ```
   ERROR: Resume 상태 파일이 손상되었습니다.
   - 파일: .moai/memory/command-state/2-run-SPEC-AUTH-001-20251112.json
   - 오류: Expecting ',' delimiter (line 3, column 5)
   - 권장 조치: 파일을 삭제하고 처음부터 재시작하십시오.
   ```
3. 백업 파일이 있으면 복구 시도
4. 복구 불가능하면 사용자에게 재시작 권장

**검증 코드**:
```python
import pytest
from moai_adk.core.resume_handler import load_resume_state

corrupted_file = ".moai/memory/command-state/corrupted.json"
# 파일 내용: 잘못된 JSON

with pytest.raises(ValueError) as exc_info:
    load_resume_state("2-run", "SPEC-AUTH-001")

assert "손상되었습니다" in str(exc_info.value)
```

---

#### 테스트 케이스 3.2.2: 필수 필드 누락

**Given**: Resume 상태 파일에 필수 필드 누락
- 파일 내용: `{"command": "alfred:2-run"}` (`spec_id` 누락)

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. 스키마 검증 실패 감지
2. 명확한 에러 메시지 출력:
   ```
   ERROR: Resume 상태 파일이 유효하지 않습니다.
   - 누락된 필드: spec_id, completed_steps, pending_steps
   - 권장 조치: 파일을 삭제하고 처음부터 재시작하십시오.
   ```
3. Resume 진행 차단

**검증 코드**:
```python
import pytest
from moai_adk.core.resume_handler import validate_resume_schema

invalid_state = {"command": "alfred:2-run"}

with pytest.raises(ValueError) as exc_info:
    validate_resume_schema(invalid_state)

assert "누락된 필드" in str(exc_info.value)
```

---

### Scenario 3.3: 코드베이스 불일치

#### 테스트 케이스 3.3.1: 브랜치 변경 감지

**Given**: Resume 상태에 브랜치 정보 기록됨
- 저장된 브랜치: feature/SPEC-AUTH-001
- 현재 브랜치: main

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. Git 상태 검증
2. 브랜치 불일치 감지
3. 경고 메시지 출력:
   ```
   WARNING: Resume 상태와 현재 브랜치가 다릅니다.
   - 저장된 브랜치: feature/SPEC-AUTH-001
   - 현재 브랜치: main
   - 권장 조치: git checkout feature/SPEC-AUTH-001
   ```
4. 사용자에게 재개 여부 확인 (AskUserQuestion)

**검증 명령**:
```bash
# 브랜치 변경
git checkout main

# Resume 시도 (경고 예상)
/alfred:resume 2-run SPEC-AUTH-001
```

---

#### 테스트 케이스 3.3.2: 커밋 해시 불일치 감지

**Given**: Resume 상태에 커밋 해시 기록됨
- 저장된 커밋: abc123def456
- 현재 커밋: xyz789uvw012

**When**: Resume 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. Git 상태 검증
2. 커밋 불일치 감지
3. 경고 메시지 출력:
   ```
   WARNING: Resume 상태 저장 이후 코드가 변경되었습니다.
   - 저장된 커밋: abc123def456
   - 현재 커밋: xyz789uvw012
   - 변경된 파일: 5개
   - 권장 조치: 변경 사항을 검토하고 충돌이 없는지 확인하십시오.
   ```
4. 사용자에게 재개 여부 확인

**검증 명령**:
```bash
# 현재 커밋 확인
git rev-parse HEAD

# Resume 상태 파일의 커밋과 비교
jq -r '.context.last_commit' .moai/memory/command-state/2-run-SPEC-AUTH-001-*.json
```

---

## Phase 4: 사용자 경험 개선

### Scenario 4.1: 다음 단계 안내

#### 테스트 케이스 4.1.1: Phase 완료 후 안내 메시지

**Given**: 0-project가 성공적으로 완료됨

**When**: 명령 실행이 종료됨

**Then**: 시스템은 다음을 수행해야 함
1. 완료 메시지 출력:
   ```
   ✓ 프로젝트 초기화가 완료되었습니다.

   다음 단계:
   1. SPEC 생성: /alfred:1-plan "기능 설명"
   2. 문서 확인: .moai/project/product.md
   ```
2. 생성된 파일 목록 표시
3. 다음 단계 추천 명령 제공

**검증 방법**:
- 출력 메시지에 "다음 단계" 섹션 존재 확인
- 추천 명령이 실행 가능한지 확인

---

#### 테스트 케이스 4.1.2: Resume 가능 상태 알림

**Given**: 2-run 중단 후 상태가 저장됨

**When**: 사용자가 다시 세션을 시작함

**Then**: 시스템은 다음을 수행해야 함
1. 세션 시작 시 Resume 가능 상태 감지
2. 알림 메시지 출력:
   ```
   📌 중단된 작업이 있습니다:
   - SPEC ID: SPEC-AUTH-001
   - 중단 시점: 2025-11-12 12:00 (3일 전)
   - 진행률: 60% (3/5 단계 완료)

   재개하려면: /alfred:resume 2-run SPEC-AUTH-001
   ```
3. Resume 명령 자동 완성 제안

**검증 방법**:
- 세션 시작 로그에서 알림 메시지 확인
- Resume 명령 실행 가능 확인

---

### Scenario 4.2: 디버깅 로그 출력 (선택적)

#### 테스트 케이스 4.2.1: --debug 플래그 사용

**Given**: 사용자가 디버깅 정보를 확인하고 싶음

**When**: `/alfred:1-plan "사용자 인증" --debug` 명령을 실행함

**Then**: 시스템은 다음을 수행해야 함
1. 디버깅 로그 출력 활성화
2. 컨텍스트 로드 과정 출력:
   ```
   [DEBUG] Loading context from .moai/memory/command-state/0-project-20251112.json
   [DEBUG] Loaded data: {"project_name": "TestProject", "mode": "personal"}
   [DEBUG] Replacing template variable: {{PROJECT_NAME}} → TestProject
   [DEBUG] Validating path: .moai/project/product.md
   [DEBUG] Converted to absolute path: /Users/goos/MoAI/TestProject/.moai/project/product.md
   [DEBUG] Calling Task(subagent_type="plan-agent", prompt="...")
   ```
3. Agent 호출 파라미터 출력
4. 저장/로드 타이밍 출력

**검증 방법**:
- 로그에 `[DEBUG]` 접두사가 있는 메시지 확인
- 디버깅 정보가 작업 흐름 이해에 도움이 되는지 확인

---

## 종합 검증 체크리스트

### Phase 1: 컨텍스트 전달 시스템
- [ ] Phase 결과 JSON 파일 생성 확인
- [ ] JSON 스키마 검증 통과
- [ ] 절대 경로 변환 정확성 확인
- [ ] 템플릿 변수 치환 완료 (미치환 변수 0건)
- [ ] Agent 호출 시 명시적 컨텍스트 전달 확인
- [ ] 원자적 파일 쓰기 동작 확인
- [ ] Unit 테스트 커버리지 90% 이상

### Phase 2: Resume 기능
- [ ] Resume 상태 저장 및 로드 정상 작동
- [ ] Timestamp 유효성 검증 정확성 (30일 기준)
- [ ] 만료된 상태 자동 무효화 확인
- [ ] 완료된 단계 스킵 동작 확인
- [ ] 대기 중인 단계 정상 실행 확인
- [ ] 사용자 재개 확인 메시지 출력
- [ ] E2E 테스트 시나리오 3개 이상 통과

### Phase 3: 오류 처리
- [ ] JSON 파싱 오류 처리 확인
- [ ] 디스크 공간 부족 에러 메시지 명확성
- [ ] 권한 오류 에러 메시지 명확성
- [ ] 손상된 파일 복구 시도 확인
- [ ] 코드베이스 불일치 경고 메시지 출력
- [ ] 모든 에러 상황에서 명확한 복구 가이드 제공

### Phase 4: 사용자 경험
- [ ] Phase 완료 후 다음 단계 안내 제공
- [ ] Resume 가능 상태 알림 메시지 출력
- [ ] 디버깅 로그 출력 기능 동작 (선택적)
- [ ] 사용자 가이드 문서 작성 완료
- [ ] 베타 사용자 피드백 수집 및 반영

---

## 성능 검증

### 컨텍스트 저장/로드 성능
- **목표**: JSON 파일 저장/로드 < 100ms
- **측정 방법**: Python `time.perf_counter()` 사용
- **검증 기준**: 10MB 이하 JSON 파일에서 목표 달성

### Resume 상태 검증 성능
- **목표**: Timestamp 검증 < 10ms
- **측정 방법**: 1000회 반복 테스트
- **검증 기준**: 평균 응답 시간 < 10ms

---

## 보안 검증

### 경로 탐색 공격 방지
- **시나리오**: 악의적 사용자가 `../../etc/passwd` 경로 접근 시도
- **검증**: Project root 외부 경로 접근 차단 확인

### JSON 인젝션 방지
- **시나리오**: 사용자 입력에 특수 문자(`"`, `{`, `}`) 포함
- **검증**: JSON 이스케이프 처리 확인

---

## 최종 승인 기준

**모든 테스트 케이스가 통과되고 다음 조건을 만족하면 인수 완료**:
1. Unit 테스트 커버리지 90% 이상
2. 통합 테스트 시나리오 5개 이상 통과
3. E2E 테스트 시나리오 3개 이상 통과
4. 베타 사용자 피드백 반영 완료
5. 사용자 가이드 및 개발자 문서 작성 완료
6. 성능 및 보안 검증 통과

---

**Author**: @goos
**Last Updated**: 2025-11-12
**Version**: 0.0.1
