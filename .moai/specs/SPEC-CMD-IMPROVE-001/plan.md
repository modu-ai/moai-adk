
**SPEC ID**: CMD-IMPROVE-001
**Title**: Commands 레이어 컨텍스트 전달 및 Resume 기능 통합 개선
**Author**: @goos
**Created**: 2025-11-12
**Priority**: HIGH

---

## 목표 (Objectives)

MoAI-ADK Commands 레이어의 안정성 및 사용자 경험을 개선하기 위해 다음 두 가지 핵심 기능을 구현:

1. **명시적 컨텍스트 전달 시스템**: Phase 간 데이터를 JSON 형식으로 저장하고 Agent 호출 시 명시적으로 전달
2. **Resume 기능**: 중단된 Command를 저장된 상태로부터 재개 가능

**예상 효과**:
- 25건의 아키텍처 이슈 해결 (CRITICAL 8건, HIGH 17건)
- 사용자 워크플로우 중단 시 재시작 비용 절감
- Agent 간 데이터 전달의 신뢰성 및 추적 가능성 확보

---

## Phase 1: 명시적 컨텍스트 전달 시스템 (Week 1-4)

### Week 1: 아키텍처 설계 및 핵심 모듈 구현

#### 1.1 컨텍스트 매니저 설계 (Priority: CRITICAL)

**목표**: Phase 간 데이터 흐름을 관리하는 핵심 모듈 설계

**작업 내용**:
- `context_manager.py` 모듈 생성
- 컨텍스트 저장/로드 인터페이스 정의
- JSON 스키마 설계 (Phase 메타데이터, 출력, 파일 경로)

**산출물**:
```python
# src/moai_adk/core/context_manager.py
class ContextManager:
    def save_phase_result(self, phase: str, data: dict) -> str:
        """Phase 결과를 JSON 파일로 저장"""
        pass

    def load_phase_result(self, phase: str) -> dict:
        """이전 Phase 결과 로드"""
        pass

    def validate_context(self, data: dict) -> bool:
        """컨텍스트 데이터 스키마 검증"""
        pass
```

**검증 방법**:
- JSON 스키마 검증 테스트
- 저장/로드 사이클 테스트

---

#### 1.2 절대 경로 변환 모듈 (Priority: HIGH)

**목표**: 상대 경로를 절대 경로로 변환하고 존재 여부를 검증

**작업 내용**:
- `path_validator.py` 모듈 생성
- Project root 기준 경로 변환 함수 구현
- 경로 존재 여부 검증 (디렉토리/파일 구분)

**산출물**:
```python
# src/moai_adk/core/path_validator.py
def validate_and_convert_path(
    relative_path: str,
    project_root: str,
    must_exist: bool = False
) -> str:
    """상대 경로를 절대 경로로 변환하고 검증"""
    pass
```

**검증 방법**:
- 다양한 경로 형식 테스트 (`.`, `..`, 절대 경로)
- Project root 외부 경로 접근 차단 테스트

---

#### 1.3 템플릿 변수 치환 엔진 (Priority: HIGH)

**목표**: `{{VARIABLE}}` 형식의 변수를 런타임 값으로 치환

**작업 내용**:
- `template_engine.py` 모듈 생성
- 변수 치환 함수 구현 (단순 문자열 치환 → Jinja2 활용 검토)
- 미치환 변수 검증 로직 추가

**산출물**:
```python
# src/moai_adk/core/template_engine.py
def replace_template_vars(text: str, context: dict) -> str:
    """템플릿 변수를 실제 값으로 치환"""
    pass

def validate_no_unsubstituted_vars(text: str) -> bool:
    """미치환 변수 검사"""
    pass
```

**검증 방법**:
- 다양한 변수 조합 테스트
- 중첩 변수 및 특수 문자 처리 테스트

---

### Week 2: 0-project.md 리팩토링

#### 2.1 Phase 결과 저장 로직 추가 (Priority: CRITICAL)

**목표**: `/alfred:0-project` 완료 시 결과를 JSON 파일로 저장

**작업 내용**:
- `0-project.md`에 컨텍스트 저장 로직 추가
- 생성된 파일 경로를 절대 경로로 변환
- `.moai/memory/command-state/0-project-{timestamp}.json` 생성

**저장 데이터 예시**:
```json
{
  "phase": "0-project",
  "timestamp": "2025-11-12T10:30:00Z",
  "status": "completed",
  "outputs": {
    "project_name": "MyProject",
    "mode": "personal",
    "conversation_language": "ko",
    "tech_stack": ["python", "fastapi"]
  },
  "files_created": [
    "/Users/goos/MoAI/MyProject/.moai/project/product.md",
    "/Users/goos/MoAI/MyProject/.moai/project/structure.md",
    "/Users/goos/MoAI/MyProject/.moai/project/tech.md"
  ],
  "next_phase": "1-plan"
}
```

**검증 방법**:
- `/alfred:0-project` 실행 후 JSON 파일 생성 확인
- `jq` 명령으로 JSON 구조 검증

---

#### 2.2 세션 의존 참조 제거 (Priority: CRITICAL)

**목표**: `위에서 언급한`, `앞서 분석한` 등의 표현을 명시적 참조로 변경

**작업 내용**:
- Placeholder 치환: `{{PROJECT_NAME}}`, `{{MODE}}`, `{{TECH_STACK}}`
- Agent 호출 시 명시적 컨텍스트 전달

**변경 전**:
```markdown
위에서 정의한 프로젝트 정보를 바탕으로...
```

**변경 후**:
```markdown
다음 프로젝트 정보를 바탕으로:
- 프로젝트명: {{PROJECT_NAME}}
- 모드: {{MODE}}
- 기술 스택: {{TECH_STACK}}
```

**검증 방법**:
- Grep으로 세션 의존 표현 검색 (0건 확인)
- Agent 호출 로그에서 치환된 변수 확인

---

### Week 3: 1-plan.md 및 2-run.md 리팩토링

#### 3.1 1-plan.md 컨텍스트 로드 로직 추가 (Priority: HIGH)

**목표**: `/alfred:1-plan` 실행 시 0-project 결과를 로드하여 활용

**작업 내용**:
- `0-project-*.json` 파일 로드
- 템플릿 변수 치환 후 plan-agent 호출
- Plan 생성 결과를 `1-plan-{spec_id}-{timestamp}.json`으로 저장

**저장 데이터 예시**:
```json
{
  "phase": "1-plan",
  "spec_id": "SPEC-AUTH-001",
  "timestamp": "2025-11-12T11:00:00Z",
  "status": "completed",
  "outputs": {
    "spec_files_created": [
      "/Users/goos/MoAI/MyProject/.moai/specs/SPEC-AUTH-001/spec.md",
      "/Users/goos/MoAI/MyProject/.moai/specs/SPEC-AUTH-001/plan.md",
      "/Users/goos/MoAI/MyProject/.moai/specs/SPEC-AUTH-001/acceptance.md"
    ],
    "implementation_priority": "high",
    "estimated_complexity": "medium"
  },
  "next_phase": "2-run"
}
```

**검증 방법**:
- 0-project 없이 1-plan 실행 시 에러 메시지 확인
- 로드된 컨텍스트가 Agent prompt에 포함되었는지 확인

---

#### 3.2 2-run.md 컨텍스트 로드 및 저장 (Priority: HIGH)

**목표**: `/alfred:2-run` 실행 시 1-plan 결과를 로드하고 구현 진행 상황 저장

**작업 내용**:
- `1-plan-{spec_id}-*.json` 파일 로드
- TDD 사이클 각 단계(RED/GREEN/REFACTOR) 완료 시 상태 업데이트
- `2-run-{spec_id}-{timestamp}.json` 파일에 진행 상황 저장

**저장 데이터 예시**:
```json
{
  "phase": "2-run",
  "spec_id": "SPEC-AUTH-001",
  "timestamp": "2025-11-12T12:00:00Z",
  "status": "in_progress",
  "completed_steps": [
    "spec_validation",
    "test_setup",
    "red_phase"
  ],
  "pending_steps": [
    "green_phase",
    "refactor_phase",
    "integration_test"
  ],
  "outputs": {
    "tests_written": ["/Users/goos/MoAI/MyProject/tests/test_auth.py"],
    "code_implemented": ["/Users/goos/MoAI/MyProject/src/auth.py"]
  },
  "next_phase": "3-sync"
}
```

**검증 방법**:
- TDD 사이클 중단 시 상태 파일 생성 확인
- `completed_steps` 배열이 단계별로 증가하는지 확인

---

### Week 4: 3-sync.md 리팩토링 및 통합 테스트

#### 4.1 3-sync.md 컨텍스트 로드 (Priority: MEDIUM)

**목표**: `/alfred:3-sync` 실행 시 2-run 결과를 로드하여 동기화 작업 수행

**작업 내용**:
- `2-run-{spec_id}-*.json` 파일 로드
- 구현된 코드 및 테스트 파일 경로를 doc-syncer에 전달
- 동기화 완료 후 `3-sync-{spec_id}-{timestamp}.json` 저장

**검증 방법**:
- 2-run 없이 3-sync 실행 시 에러 메시지 확인
- 동기화 대상 파일 목록이 올바른지 확인

---

#### 4.2 통합 테스트 시나리오 작성 (Priority: HIGH)

**목표**: 0-project → 1-plan → 2-run → 3-sync 전체 워크플로우 검증

**테스트 시나리오**:
1. **정상 흐름**: 각 Phase 순차 실행 후 컨텍스트 전달 확인
2. **Phase 누락**: 중간 Phase 없이 다음 Phase 실행 시 에러 발생 확인
3. **데이터 일관성**: 저장/로드 사이클 후 데이터 무결성 확인

**테스트 파일**:
- `tests/integration/test_full_workflow.py`

**검증 방법**:
- 모든 시나리오 통과 확인 (pytest 실행)
- 커버리지 90% 이상 달성

---

## Phase 2: Resume 기능 구현 (Week 5-8)

### Week 5: Resume 핸들러 설계 및 구현

#### 5.1 Resume 핸들러 모듈 설계 (Priority: HIGH)

**목표**: 저장된 상태로부터 Command를 재개하는 핵심 로직 구현

**작업 내용**:
- `resume_handler.py` 모듈 생성
- Resume 상태 로드 인터페이스 정의
- Timestamp 유효성 검증 로직 추가

**산출물**:
```python
# src/moai_adk/core/resume_handler.py
class ResumeHandler:
    def load_resume_state(self, command: str, spec_id: str) -> dict:
        """저장된 Resume 상태 로드"""
        pass

    def validate_timestamp(self, state: dict) -> bool:
        """Timestamp 유효성 검증 (30일 이내)"""
        pass

    def resume_from_step(self, state: dict, step: str) -> None:
        """특정 단계부터 Command 재개"""
        pass
```

**검증 방법**:
- 유효한 상태 로드 테스트
- 만료된 상태 검증 테스트

---

#### 5.2 Resume 명령 파싱 (Priority: MEDIUM)

**목표**: `/alfred:resume` 명령을 파싱하여 적절한 핸들러 호출

**작업 내용**:
- Resume 명령 형식 정의: `/alfred:resume {command} {spec_id} [--from=step]`
- 명령 파싱 로직 구현
- 존재하지 않는 상태 파일 처리

**예시**:
```bash
# 기본 재개
/alfred:resume 2-run SPEC-AUTH-001

# 특정 단계부터 재개
/alfred:resume 2-run SPEC-AUTH-001 --from=integration_test

# 처음부터 재실행
/alfred:resume 2-run SPEC-AUTH-001 --restart
```

**검증 방법**:
- 다양한 명령 형식 파싱 테스트
- 잘못된 명령 형식에 대한 에러 메시지 확인

---

### Week 6: Resume 로직 통합

#### 6.1 2-run.md에 Resume 로직 추가 (Priority: HIGH)

**목표**: `/alfred:2-run` 실행 시 Resume 상태가 있으면 자동 재개

**작업 내용**:
- Command 시작 시 Resume 상태 파일 검색
- 상태가 있으면 사용자에게 재개 여부 확인 (AskUserQuestion)
- 재개 시 `completed_steps` 단계는 스킵하고 `pending_steps`부터 실행

**사용자 확인 메시지**:
```
저장된 작업 상태가 발견되었습니다.
- SPEC ID: SPEC-AUTH-001
- 중단 시점: 2025-11-12 12:00
- 완료된 단계: spec_validation, test_setup, red_phase
- 대기 중인 단계: green_phase, refactor_phase, integration_test

이전 작업을 재개하시겠습니까?
1. 예 (대기 중인 단계부터 재개)
2. 아니오 (처음부터 재시작)
```

**검증 방법**:
- Resume 상태가 있는 경우와 없는 경우 동작 확인
- 단계 스킵 로직 검증

---

#### 6.2 상태 자동 만료 로직 (Priority: MEDIUM)

**목표**: 30일 경과한 Resume 상태 자동 무효화

**작업 내용**:
- Resume 상태 로드 시 `expiry` 필드 검증
- 만료된 상태 감지 시 경고 메시지 출력 및 삭제
- `.moai/memory/command-state/` 디렉토리 정리 스크립트 추가

**만료 처리 로직**:
```python
import time
from datetime import datetime, timedelta

def is_expired(state: dict) -> bool:
    """Resume 상태 만료 여부 확인"""
    expiry = datetime.fromisoformat(state['expiry'])
    return datetime.now() > expiry

def cleanup_expired_states(state_dir: str) -> int:
    """만료된 상태 파일 삭제"""
    # 30일 이상 경과한 파일 삭제
    pass
```

**검증 방법**:
- 만료된 상태 로드 시 에러 발생 확인
- 정리 스크립트 실행 후 파일 삭제 확인

---

### Week 7: 사용자 가이드 및 문서화

#### 7.1 Resume 기능 사용자 가이드 작성 (Priority: HIGH)

**목표**: 사용자가 Resume 기능을 쉽게 이해하고 활용할 수 있도록 문서화

**작성 내용**:
- Resume 기능 개요 및 사용 사례
- 명령 사용법 및 옵션 설명
- 문제 해결 가이드 (FAQ)

**문서 위치**: `docs/commands/resume-guide.md`

**주요 섹션**:
1. **개요**: Resume 기능이란?
2. **사용 방법**: 기본 명령 및 옵션
3. **제한 사항**: 30일 만료, 파일 시스템 의존성
4. **문제 해결**: 상태 파일 손상, 만료된 상태 처리

**검증 방법**:
- 베타 사용자 피드백 수집
- 문서 링크 유효성 검증

---

#### 7.2 개발자 문서 작성 (Priority: MEDIUM)

**목표**: 내부 개발자를 위한 아키텍처 및 API 문서 작성

**작성 내용**:
- 컨텍스트 매니저 아키텍처 설명
- Resume 핸들러 API 레퍼런스
- 확장 가이드 (새로운 Command 추가 시 고려사항)

**문서 위치**:
- `docs/architecture/command-state-management.md`
- `docs/api/context-manager.md`
- `docs/api/resume-handler.md`

**검증 방법**:
- 문서 리뷰 세션 진행
- API 예제 코드 실행 가능 여부 확인

---

### Week 8: E2E 테스트 및 베타 배포

#### 8.1 E2E 테스트 시나리오 작성 (Priority: HIGH)

**목표**: 실제 사용자 워크플로우를 시뮬레이션하는 E2E 테스트 작성

**테스트 시나리오**:
1. **정상 Resume**: 2-run 중단 후 재개하여 완료
2. **만료된 상태**: 30일 경과한 상태로 Resume 시도 시 에러
3. **충돌 해결**: Resume 중 코드베이스 변경 감지 시 경고

**테스트 파일**:
- `tests/e2e/test_resume_scenarios.py`

**검증 방법**:
- 모든 시나리오 통과 확인
- 실제 프로젝트에서 수동 테스트 수행

---

#### 8.2 베타 배포 및 피드백 수집 (Priority: HIGH)

**목표**: 초기 사용자에게 Resume 기능을 배포하고 피드백 수집

**작업 내용**:
- 베타 버전 릴리즈 (v0.21.0-beta)
- 베타 사용자 그룹에 안내 메일 발송
- 피드백 수집 채널 운영 (GitHub Discussions, Slack)

**주요 피드백 항목**:
- Resume 기능 사용 편의성
- 오류 메시지 명확성
- 문서 완성도

**검증 방법**:
- 피드백 수집 및 이슈 트래킹
- 주요 버그 수정 후 정식 릴리즈 (v0.21.0)

---

## 기술 스택 (Tech Stack)

### Core Technologies
- **Python 3.9+**: 핵심 로직 구현
- **pathlib**: 경로 관리 및 검증
- **json**: 컨텍스트 데이터 저장/로드
- **datetime**: Timestamp 및 만료 검증

### Testing Tools
- **pytest**: Unit 및 Integration 테스트
- **pytest-mock**: 파일 I/O 모킹
- **pytest-cov**: 커버리지 측정

### Documentation Tools
- **Markdown**: 사용자 가이드 작성
- **MkDocs**: API 문서 생성 (선택 사항)

---

## 의존성 및 제약사항 (Dependencies & Constraints)

### Dependencies
- **MoAI-ADK Core**: 기존 Command 인프라 활용
- **Git**: Resume 상태 검증 시 현재 브랜치 정보 필요
- **파일 시스템**: JSON 저장/로드를 위한 디스크 접근 필수

### Constraints
1. **하위 호환성**: 기존 `.moai/config.json` 구조 변경 불가
2. **메모리 제약**: 대규모 프로젝트에서 전체 컨텍스트를 메모리에 로드할 수 없음
3. **파일 시스템 의존성**: 컨테이너/VM 재시작 시 상태 유지 보장 필요

---

## 위험 요소 및 대응 방안 (Risks & Mitigations)

### Risk 1: JSON 파싱 오류
- **발생 가능성**: MEDIUM
- **영향도**: HIGH
- **대응 방안**:
  - JSON 스키마 검증 추가
  - 손상된 파일 감지 시 백업 파일 로드
  - 사용자에게 명확한 복구 가이드 제공

### Risk 2: 디스크 공간 부족
- **발생 가능성**: LOW
- **영향도**: MEDIUM
- **대응 방안**:
  - 30일 이상 경과한 상태 파일 자동 삭제
  - 최대 50개 상태 파일 제한
  - 디스크 공간 부족 시 경고 메시지 출력

### Risk 3: 병렬 실행 충돌
- **발생 가능성**: LOW (Personal 모드에서는 거의 없음)
- **영향도**: HIGH
- **대응 방안**:
  - 파일 잠금(file locking) 메커니즘 추가
  - 중복 실행 감지 및 경고
  - Team 모드에서만 활성화

### Risk 4: 코드베이스 불일치
- **발생 가능성**: MEDIUM
- **영향도**: HIGH
- **대응 방안**:
  - Resume 시 Git 상태 검증 (브랜치, 커밋 해시)
  - 불일치 감지 시 사용자에게 경고 및 재시작 권장
  - 로그에 불일치 세부 정보 기록

---

## 마일스톤 및 우선순위 (Milestones & Priorities)

### Primary Goals (MUST HAVE)
- [x] Week 1: 컨텍스트 매니저 핵심 모듈 완성
- [x] Week 2: command_helpers.py 구현 완료 (2025-11-12)
  - 27 tests, 90.41% coverage
- [ ] Week 3-4: 4개 Command 파일 리팩토링
- [ ] Week 5-6: Resume 핸들러 구현
- [ ] Week 7-8: 문서화 및 E2E 테스트

### Secondary Goals (SHOULD HAVE)
- [ ] 디버깅 로그 출력 기능
- [ ] Resume 상태 자동 정리 스크립트
- [ ] Team 모드 병렬 실행 충돌 방지

### Final Goals (NICE TO HAVE)
- [ ] Resume 재시작 옵션 (`--from`, `--restart`)
- [ ] 상태 파일 암호화 (민감 정보 보호)
- [ ] 웹 UI를 통한 Resume 상태 모니터링

---

## Definition of Done (완료 기준)

### Phase 1: 컨텍스트 전달 시스템
- [ ] 4개 Command 파일에서 명시적 컨텍스트 전달 구현 완료
- [ ] `.moai/memory/command-state/` 디렉토리 생성 및 JSON 저장 기능 구현
- [ ] 템플릿 변수 치환 엔진 구현 및 검증
- [ ] 절대 경로 변환 및 검증 함수 구현
- [ ] Unit 테스트 커버리지 90% 이상
- [ ] 통합 테스트 시나리오 5개 이상 통과

### Phase 2: Resume 기능
- [ ] Resume 명령 파싱 및 상태 로드 기능 구현
- [ ] Timestamp 검증 및 자동 만료 로직 구현
- [ ] 단계별 재개 메커니즘 구현
- [ ] 사용자 가이드 문서 작성 완료
- [ ] E2E 테스트 시나리오 3개 이상 통과
- [ ] 베타 사용자 피드백 수집 및 반영

---

## 참고 자료 (References)

- **SPEC 문서**: `.moai/specs/SPEC-CMD-IMPROVE-001/spec.md`
- **Acceptance 기준**: `.moai/specs/SPEC-CMD-IMPROVE-001/acceptance.md`
- **Commands 파일**:
  - `src/moai_adk/templates/.claude/commands/alfred/0-project.md`
  - `src/moai_adk/templates/.claude/commands/alfred/1-plan.md`
  - `src/moai_adk/templates/.claude/commands/alfred/2-run.md`
  - `src/moai_adk/templates/.claude/commands/alfred/3-sync.md`
- **관련 이슈**: GitHub Issue #XXX (생성 예정)
