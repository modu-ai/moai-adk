# @SPEC:TEST-COVERAGE-001 인수 기준

## 개요

- **SPEC ID**: TEST-COVERAGE-001
- **목표**: CLI 및 Git 모듈 테스트 커버리지 85% 달성
- **인수 조건**: 3개 핵심 시나리오 검증

---

## AC-1: 실패 테스트 수정

### Given (전제 조건)
- 100개의 실패 테스트가 존재한다
  - CLI 통합 테스트: 30개 실패
  - moai_hooks.py 테스트: 49개 실패
  - Git Utils 테스트: 6개 실패
  - 기타 테스트: 15개 실패
- pytest 8.4.2 및 pytest-cov 6.0.0이 설치되어 있다
- 현재 커버리지: 72.06%

### When (실행 조건)
1. 각 실패 테스트를 수정한다
2. 다음 명령어를 실행한다:
   ```bash
   pytest -v --cov --cov-report=term-missing
   ```

### Then (예상 결과)
- ✅ 모든 테스트가 통과해야 한다 (0 failed, 235+ passed)
- ✅ 커버리지가 80% 이상이어야 한다
- ✅ 테스트 실행 시간이 30초 이내여야 한다
- ✅ 커버리지 리포트가 터미널에 출력되어야 한다

### 검증 방법
```bash
# 1. 테스트 실행
pytest -v --cov --cov-report=term-missing

# 2. 결과 확인
# Expected output:
# ======================== 235 passed in 28.50s ========================
# Coverage: 80.5%

# 3. 실패 테스트 확인
pytest --tb=short | grep "FAILED"
# Expected: (no output)
```

### 성공 기준
- [ ] 실패 테스트 0개
- [ ] 커버리지 80% 이상
- [ ] 테스트 실행 시간 30초 이내

---

## AC-2: Git 모듈 단위 테스트

### Given (전제 조건)
- Git 모듈 커버리지가 0%이다
  - `core/git/__init__.py`: 0%
  - `core/git/manager.py`: 0%
  - `core/git/branch.py`: 0%
  - `core/git/commit.py`: 0%
- Git 저장소가 초기화되어 있다
- 임시 디렉토리를 사용한 테스트 환경이 준비되어 있다

### When (실행 조건)
1. Git 모듈 단위 테스트를 작성한다:
   - `tests/unit/test_git_manager.py` (GitManager 클래스)
   - `tests/unit/test_git_branch.py` (브랜치 생성/전환)
   - `tests/unit/test_git_commit.py` (커밋 메시지 생성)

2. 다음 명령어를 실행한다:
   ```bash
   pytest tests/unit/test_git_*.py -v --cov=src/moai_adk/core/git
   ```

### Then (예상 결과)
- ✅ `core/git/manager.py` 커버리지가 60% 이상이어야 한다
- ✅ `core/git/branch.py` 커버리지가 70% 이상이어야 한다
- ✅ `core/git/commit.py` 커버리지가 70% 이상이어야 한다
- ✅ 모든 Git 작업이 임시 디렉토리에서 실행되어야 한다
- ✅ 테스트 종료 시 임시 디렉토리가 삭제되어야 한다

### 검증 방법
```bash
# 1. Git 모듈 테스트 실행
pytest tests/unit/test_git_*.py -v --cov=src/moai_adk/core/git --cov-report=term-missing

# 2. 커버리지 확인
# Expected output:
# src/moai_adk/core/git/manager.py    65%
# src/moai_adk/core/git/branch.py     72%
# src/moai_adk/core/git/commit.py     75%

# 3. 임시 디렉토리 정리 확인
ls /tmp | grep moai_test_
# Expected: (no output)
```

### 성공 기준
- [ ] Git 모듈 평균 커버리지 60% 이상
- [ ] manager.py: 60% 이상
- [ ] branch.py: 70% 이상
- [ ] commit.py: 70% 이상
- [ ] 임시 디렉토리 자동 정리

---

## AC-3: 전체 커버리지 목표 달성

### Given (전제 조건)
- Phase 1 (실패 테스트 수정) 완료: 커버리지 80%
- Phase 2 (Git 모듈 테스트) 완료: Git 모듈 60% 이상
- 모든 테스트가 작성되어 있다
- CI/CD 파이프라인이 설정되어 있다

### When (실행 조건)
1. 전체 테스트를 실행한다:
   ```bash
   pytest --cov --cov-report=html --cov-report=term-missing
   ```

2. CI/CD 빌드를 트리거한다:
   ```bash
   git push origin feature/SPEC-TEST-COVERAGE-001
   ```

### Then (예상 결과)
- ✅ 전체 커버리지가 85% 이상이어야 한다
- ✅ 실패 테스트가 0개여야 한다
- ✅ 커버리지 리포트가 HTML로 생성되어야 한다 (`htmlcov/index.html`)
- ✅ CI/CD 빌드가 통과해야 한다
- ✅ 커버리지 배지가 업데이트되어야 한다

### 검증 방법
```bash
# 1. 전체 테스트 실행
pytest --cov --cov-report=html --cov-report=term-missing

# 2. 커버리지 확인
# Expected output:
# TOTAL    859    730    85%

# 3. HTML 리포트 확인
open htmlcov/index.html

# 4. CI/CD 빌드 확인
gh run list --workflow=tests.yml
# Expected: ✓ Tests / 85% coverage
```

### 성공 기준
- [ ] 전체 커버리지 85% 이상
- [ ] 실패 테스트 0개
- [ ] HTML 리포트 생성
- [ ] CI/CD 빌드 통과
- [ ] 커버리지 배지 업데이트

---

## 종합 인수 조건

### 필수 조건 (Must Have)
1. ✅ 실패 테스트 0개
2. ✅ 전체 커버리지 85% 이상
3. ✅ Git 모듈 커버리지 60% 이상
4. ✅ 테스트 실행 시간 30초 이내

### 권장 조건 (Should Have)
1. ✅ 커버리지 HTML 리포트 생성
2. ✅ CI/CD 빌드 통과
3. ✅ 커버리지 배지 표시

### 선택 조건 (Nice to Have)
1. ⭕ 커버리지 90% 이상
2. ⭕ pytest-xdist 병렬 테스트 실행
3. ⭕ 테스트 실행 시간 20초 이내

---

## 최종 검증 체크리스트

### Phase 1 완료 확인
- [ ] CLI 통합 테스트 30개 수정 완료
- [ ] moai_hooks.py 테스트 49개 수정 완료
- [ ] Git Utils 테스트 6개 수정 완료
- [ ] 기타 테스트 15개 수정 완료
- [ ] 커버리지 80% 이상 달성

### Phase 2 완료 확인
- [ ] Git Manager 테스트 작성 완료
- [ ] Git Branch 테스트 작성 완료
- [ ] Git Commit 테스트 작성 완료
- [ ] Git 통합 테스트 작성 완료
- [ ] Git 모듈 커버리지 60% 이상 달성

### Phase 3 완료 확인
- [ ] 엣지 케이스 테스트 추가 완료
- [ ] CI/CD 설정 완료
- [ ] 커버리지 배지 추가 완료
- [ ] 전체 커버리지 85% 이상 유지

---

## 테스트 시나리오 예시

### 시나리오 1: CLI 통합 테스트
```python
def test_init_command_success(tmp_path):
    """Given: 빈 디렉토리
    When: moai init . 실행
    Then: .moai/ 디렉토리 생성"""
    # 테스트 구현
```

### 시나리오 2: Git 브랜치 생성
```python
def test_create_feature_branch(temp_git_repo):
    """Given: develop 브랜치
    When: feature/SPEC-AUTH-001 브랜치 생성
    Then: 브랜치 생성 성공"""
    # 테스트 구현
```

### 시나리오 3: 커밋 메시지 생성
```python
def test_commit_message_locale_ko():
    """Given: locale = 'ko'
    When: RED 단계 커밋 메시지 생성
    Then: '🔴 RED: ...' 형식"""
    # 테스트 구현
```

---

## 참조 문서

- **SPEC 문서**: `.moai/specs/SPEC-TEST-COVERAGE-001/spec.md`
- **구현 계획**: `.moai/specs/SPEC-TEST-COVERAGE-001/plan.md`
- **TRUST 원칙**: `.moai/memory/development-guide.md#trust-5원칙`
- **pytest 가이드**: https://docs.pytest.org/
- **커버리지 가이드**: https://coverage.readthedocs.io/
