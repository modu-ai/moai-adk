# @SPEC:TEST-COVERAGE-001 구현 계획서

## 개요

- **SPEC ID**: TEST-COVERAGE-001
- **목표**: CLI 및 Git 모듈 테스트 커버리지 85% 달성
- **현재 상태**: 72.06% (619/859 statements)
- **목표 상태**: 85% (730/859 statements)
- **Gap**: 12.94% (111 statements)

## 구현 전략

### Phase 1: 실패 테스트 수정
**목표**: 72% → 80% (+8%)

#### 1.1 CLI 통합 테스트 수정 (우선순위 High)
- [ ] `test_cli.py`: `test_no_arguments` 수정
- [ ] `test_cli_commands.py`: 초기화 명령어 테스트 수정
- [ ] `test_commands.py`: InitCommand, StatusCommand, RestoreCommand 수정
- **예상 결과**: 30개 실패 테스트 수정 완료

#### 1.2 moai_hooks.py 테스트 수정 (우선순위 High)
- [ ] Language Detection 테스트 18개 수정
- [ ] JIT Context 테스트 6개 수정
- [ ] Hook Handlers 테스트 11개 수정
- [ ] Integration 테스트 4개 수정
- **예상 결과**: 49개 실패 테스트 수정 완료

#### 1.3 Git Utils 테스트 수정 (우선순위 Medium)
- [ ] `test_git_utils.py`: 3개 실패 테스트 수정
- [ ] `test_git_manager.py`: 3개 실패 테스트 수정
- **예상 결과**: 6개 실패 테스트 수정 완료

#### 1.4 기타 테스트 수정 (우선순위 Medium)
- [ ] Template Processor: 9개 테스트 수정
- [ ] Project Initializer: 10개 테스트 수정
- [ ] Backup Utils: 3개 테스트 수정
- **예상 결과**: 22개 실패 테스트 수정 완료

**Phase 1 완료 조건**:
- ✅ 실패 테스트 0개
- ✅ 커버리지 80% 이상
- ✅ 모든 테스트 통과

---

### Phase 2: Git 모듈 테스트 추가
**목표**: 80% → 85% (+5%)

#### 2.1 단위 테스트 작성 (우선순위 High)
- [ ] `tests/unit/test_git_manager.py` 작성
  - GitManager 초기화 테스트
  - 브랜치 생성 로직 테스트
  - 커밋 메시지 생성 테스트
  - 모드별 브랜치 전략 테스트

- [ ] `tests/unit/test_git_branch.py` 작성
  - 브랜치 생성 테스트
  - 브랜치 전환 테스트
  - 브랜치 삭제 테스트
  - 원격 브랜치 동기화 테스트

- [ ] `tests/unit/test_git_commit.py` 작성
  - 커밋 메시지 생성 테스트 (locale별)
  - 커밋 생성 테스트
  - TDD 단계별 커밋 테스트 (RED/GREEN/REFACTOR/DOCS)

**예상 테스트 수**: 약 30개

#### 2.2 통합 테스트 작성 (우선순위 Medium)
- [ ] `tests/integration/test_git_workflow.py` 작성
  - Personal 모드 전체 워크플로우 테스트
  - Team 모드 전체 워크플로우 테스트
  - 브랜치 → 커밋 → PR 생성 플로우 테스트

**예상 테스트 수**: 약 10개

#### 2.3 엣지 케이스 테스트 (우선순위 Low)
- [ ] Git 충돌 처리 테스트
- [ ] 원격 저장소 없는 경우 테스트
- [ ] gh CLI 미설치 경우 테스트
- [ ] 권한 오류 처리 테스트

**Phase 2 완료 조건**:
- ✅ Git 모듈 커버리지 60% 이상
- ✅ 전체 커버리지 85% 이상
- ✅ 모든 테스트 통과

---

### Phase 3: 엣지 케이스 및 유지보수
**목표**: 85% 유지

#### 3.1 추가 엣지 케이스 (필요 시)
- [ ] Template Processor 엣지 케이스 추가
- [ ] Backup Utils 엣지 케이스 추가
- [ ] Project Initializer 엣지 케이스 추가

#### 3.2 CI/CD 설정
- [ ] GitHub Actions 워크플로우 업데이트
- [ ] 커버리지 임계값 85%로 설정
- [ ] 커버리지 배지 추가 (README.md)

**Phase 3 완료 조건**:
- ✅ 커버리지 85% 이상 유지
- ✅ CI/CD 빌드 통과
- ✅ 커버리지 배지 표시

---

## 우선순위별 작업 순서

### 1차 목표 (Critical)
1. CLI 통합 테스트 수정 (30개)
2. moai_hooks.py 테스트 수정 (49개)
→ 커버리지 78% 달성

### 2차 목표 (High)
3. Git Utils 테스트 수정 (6개)
4. Git Manager 단위 테스트 작성 (30개)
→ 커버리지 85% 달성

### 3차 목표 (Medium)
5. Git 통합 테스트 작성 (10개)
6. 기타 엣지 케이스 수정 (22개)
→ 커버리지 85% 유지

---

## 위험 요소 및 완화 방안

### 위험 요소 1: 실패 테스트 수정 시간 초과
- **완화 방안**: 우선순위별 수정 (CLI → moai_hooks → Git Utils)
- **대안**: Phase 1과 Phase 2 병렬 진행

### 위험 요소 2: Git 모듈 테스트 복잡도
- **완화 방안**: 임시 디렉토리 사용, 테스트 격리
- **대안**: Mock 객체 활용 (unittest.mock)

### 위험 요소 3: 커버리지 85% 미달
- **완화 방안**: 엣지 케이스 추가 테스트 작성
- **대안**: 목표 임계값 조정 (80% → 85%)

---

## 기술적 접근 방법

### 테스트 격리 전략
```python
import tempfile
import shutil
from pathlib import Path

@pytest.fixture
def temp_git_repo():
    """임시 Git 저장소 생성"""
    temp_dir = tempfile.mkdtemp(prefix="moai_test_")
    yield Path(temp_dir)
    shutil.rmtree(temp_dir)
```

### 커버리지 측정 명령어
```bash
# 전체 커버리지 측정
pytest --cov --cov-report=term-missing --cov-report=html

# 특정 모듈만 측정
pytest tests/unit/test_git_*.py --cov=src/moai_adk/core/git

# 커버리지 85% 미만 시 실패
pytest --cov --cov-fail-under=85
```

### TDD 사이클
```bash
# RED: 실패하는 테스트 작성
pytest tests/unit/test_git_manager.py -k test_create_branch

# GREEN: 최소한의 구현
# (코드 작성)

# REFACTOR: 코드 품질 개선
pytest --cov=src/moai_adk/core/git/manager.py
```

---

## 성공 기준

1. ✅ 실패 테스트 0개
2. ✅ 전체 커버리지 85% 이상
3. ✅ Git 모듈 커버리지 60% 이상
4. ✅ 테스트 실행 시간 30초 이내
5. ✅ CI/CD 빌드 통과

---

## 참조 문서

- **TRUST 원칙**: `.moai/memory/development-guide.md#trust-5원칙`
- **SPEC 문서**: `.moai/specs/SPEC-TEST-COVERAGE-001/spec.md`
- **Acceptance Criteria**: `.moai/specs/SPEC-TEST-COVERAGE-001/acceptance.md`
- **pytest 설정**: `pyproject.toml` [tool.pytest.ini_options]
- **커버리지 설정**: `pyproject.toml` [tool.coverage.run]
