# release-new.md 지침 상세 분석 보고서

**분석 대상**: `.claude/commands/alfred/release-new.md`
**파일 크기**: 3,384줄 (97KB)
**작성 일시**: 2025-11-04
**상태**: 로컬 개발용 (템플릿 동기화 금지)

---

## 🎯 핵심 목적

Python 패키지(moai-adk)의 **완전 자동화된 릴리즈 파이프라인**을 제공합니다.

**주요 기능**:
- ✅ 품질 검증 자동화 (pytest, ruff, mypy, bandit, pip-audit)
- ✅ 버전 관리 (SSOT: Single Source of Truth)
- ✅ GitFlow 워크플로우 자동화
- ✅ GitHub Actions 기반 CI/CD
- ✅ PyPI 배포 자동화
- ✅ TestPyPI 테스트 배포
- ✅ Dry-Run 시뮬레이션

---

## 📋 문서 구조 분석

### 1. YAML 프론트매터 (1-6줄)

```yaml
name: awesome:release-new
description: 패키지 배포 및 GitHub 릴리즈 자동화
argument-hint: "[patch|minor|major] [--dry-run] [--testpypi]..."
tools: Read, Write, Edit, Bash(git:*), Bash(gh:*), Bash(python:*), Bash(uv:*), Task
```

**구성 요소**:
- **name**: `awesome:release-new` (명령 ID)
- **description**: 패키지 배포 및 GitHub 릴리즈 자동화
- **argument-hint**: 사용자를 위한 인수 가이드
- **tools**: 필요한 도구 권한
  - Git 작업: `Bash(git:*)`
  - GitHub CLI: `Bash(gh:*)`
  - Python 실행: `Bash(python:*)`
  - uv 패키지 관리: `Bash(uv:*)`
  - Task 에이전트 호출 권한

---

### 2. 제목 및 개요 (7-30줄)

**섹션 내용**:
- 커맨드 목적 5가지 단계 정의
- SSOT 버전 관리 방식 설명
- 인수: `$ARGUMENTS` 처리 방식

**핵심 개념 - SSOT (Single Source of Truth)**:
```
✅ 버전은 pyproject.toml 한 곳에만 존재
✅ __init__.py는 importlib.metadata로 자동 로드
✅ 버전 업데이트는 pyproject.toml만 수정
```

이 방식으로:
- 버전 불일치 문제 제거
- 릴리즈 시 파일 수정 최소화
- 유지보수성 향상

---

### 3. Dry-Run 모드 가이드 (32-105줄)

**목적**: 릴리즈 프로세스를 **시뮬레이션**하고 실제 변경은 하지 않음

#### 3.1 사용 방법
```bash
/awesome:release-new [patch|minor|major] --dry-run
```

#### 3.2 실제 실행 작업 (변경 없음)
- Phase 0: 품질 검증 (pytest, ruff, mypy, bandit, pip-audit)
- 버전 계산 및 분석
- Git 로그 분석
- 릴리즈 계획 보고서 생성
- 변경사항 요약

#### 3.3 시뮬레이션만 하는 작업
- ~~파일 수정 (pyproject.toml)~~
- ~~Git 커밋 생성~~
- ~~Git 태그 생성~~
- ~~GitHub PR 생성~~
- ~~GitHub Release 생성~~
- ~~PyPI 배포~~

#### 3.4 결과 리포트 예시
```markdown
🔬 Dry-Run 시뮬레이션 완료 (실제 변경 없음)

📊 시뮬레이션 계획:
- 현재 버전: v0.13.0
- 목표 버전: v0.13.1 (patch)
- 변경사항: 5개 커밋
```

**워크플로우**:
1. Dry-Run 실행 → 계획 보고서 확인
2. 결과 만족 → `--dry-run` 제외하고 실제 실행

---

### 4. TestPyPI 배포 가이드 (106-350줄)

**목적**: 실제 PyPI 배포 전에 테스트 환경에서 검증

#### 4.1 TestPyPI란?
- URL: https://test.pypi.org/
- 목적: PyPI와 동일한 환경에서 테스트
- 특징: 테스트 패키지는 30일 후 자동 삭제

#### 4.2 사용 방법
```bash
/awesome:release-new minor --testpypi
/awesome:release-new minor --dry-run --testpypi  # 조합 사용
```

#### 4.3 TestPyPI 배포 워크플로우
```
/awesome:release-new [version] --testpypi
    ↓
Phase 0: 품질 검증 (동일)
    ↓
Phase 1: 버전 분석 (동일)
    ↓
Phase 1.5: 릴리즈 계획 보고서 (TestPyPI 표시)
    ↓
Phase 2: PR 관리 (생략됨 - GitHub에 영향 없음)
    ↓
Phase 3: TestPyPI 배포 (PyPI 대신 TestPyPI에만 배포)
    ↓
✅ TestPyPI 배포 완료
```

#### 4.4 초기 설정 (한 번만)
1. TestPyPI 토큰 생성 (https://test.pypi.org/manage/account/token/)
2. ~/.pypirc 또는 환경 변수 설정
   ```bash
   export UV_PUBLISH_TOKEN_TESTPYPI="pypi-AgEIcHlwaS5vcmcCJ..."
   ```

---

### 5. 핵심 실행 단계 분석

#### 5.1 Phase 0: 품질 검증 (필수, 549줄부터)

**검증 항목**:
1. **테스트**: `pytest --cov` (커버리지 확인)
2. **린트**: `ruff check` (코드 스타일)
3. **타입**: `mypy` (타입 안정성)
4. **보안**: `bandit` (보안 취약점)
5. **의존성**: `pip-audit` (의존성 보안)

**특징**:
- 검증 실패 시 → 릴리즈 중단
- Dry-Run 모드에서도 → **실제 실행** (검증은 항상 필요)
- CodeRabbit AI 자동 통합 가능

**검증 실패 처리**:
```bash
# 검증 실패 시
❌ 문제 해결
✅ 재시도: /awesome:release-new patch
```

---

#### 5.2 Phase 1: 버전 분석 및 검증

**수행 작업**:
1. 현재 프로젝트 버전 확인 (pyproject.toml, __init__.py)
2. 목표 버전 결정
   - 인수로 명시: `patch`, `minor`, `major`
   - 또는 자동 증가
3. Git 상태 확인 (커밋 가능 여부)
4. 변경사항 요약

**버전 규칙**:
```
patch: 0.13.0 → 0.13.1 (버그 수정)
minor: 0.13.0 → 0.14.0 (기능 추가)
major: 0.13.0 → 1.0.0 (주요 변경)
```

---

#### 5.3 Phase 1.5: 사용자 확인 (릴리즈 계획 보고서)

**생성되는 보고서 항목**:

```markdown
## 🚀 릴리즈 계획 보고서: v{new_version}

### 🧪 품질 검증 결과 (Phase 0)
- ✅ 테스트: 통과 (커버리지 87%)
- ✅ 린트: 통과
- ✅ 타입: 통과
- ✅ 보안: 통과

### 📊 버전 정보
- 현재 버전: v{current_version}
- 목표 버전: v{new_version}
- 버전 타입: {patch|minor|major}

### 📁 프로젝트 정보
- 프로젝트: moai-adk
- 현재 브랜치: {current_branch}
- 마지막 커밋: {commit_hash}

### 📝 변경사항 요약
- Added: N개 기능
- Fixed: N개 버그
- Documentation: N개 문서 업데이트

### 🔄 업데이트할 파일 (SSOT)
- [ ] pyproject.toml: {current} → {new}
- [ ] __init__.py: 수정 불필요 (자동 로드)

### 🚀 릴리즈 작업
- [ ] Git 커밋: 🔖 RELEASE: v{new_version}
- [ ] Git 태그: v{new_version} (Annotated)
- [ ] PyPI 배포: uv publish
- [ ] GitHub Release: gh release create
```

**사용자 응답 옵션**:
- "진행": 릴리즈 계속
- "수정 [내용]": 수정 후 재시도
- "중단": 릴리즈 중단

**Dry-Run 모드에서**: 사용자 승인을 기다리지 않고 "실제 릴리즈 명령" 제시

---

#### 5.4 Phase 2: GitFlow PR 관리 (1056줄부터)

**전제 조건**: Phase 1에서 사용자가 "진행" 선택

**모드 감지** (자동):
```bash
# 1. 프로젝트 모드 감지 (.moai/config.json)
project_mode=$(rg '"mode":\s*"([^"]+)"' .moai/config.json)

# 2. 워크플로우 감지
- GitFlow: develop 브랜치 존재 시
- Simplified: develop 브랜치 없을 시
```

**두 가지 모드**:

**Personal 모드**:
- PR 단계 자동 건너뜀
- 로컬 개발용
- Phase 3으로 직접 진행

**Team 모드**:
- Full GitFlow PR 프로세스 실행
- develop 브랜치 확인
- GitHub PR 생성 (Draft)
- CodeRabbit AI 자동 리뷰
- PR 병합 (GitHub 웹에서만)

#### 5.5 Step 2.0-2.3: 세부 단계

**Step 2.0**: 모드 및 워크플로우 자동 감지

**Step 2.1**: Develop 브랜치 확인
```bash
git show-ref --verify --quiet refs/heads/develop
# GitFlow 워크플로우 감지
```

**Step 2.2**: PR 생성
```bash
gh pr create \
  --base develop \
  --head feature/SPEC-XXX \
  --title "[RELEASE] v{new_version}" \
  --draft
```

**Step 2.3**: PR을 Ready for Review로 전환
- CodeRabbit AI 자동 리뷰 대기
- 품질 80% 이상 자동 승인
- GitHub 웹에서 병합

---

#### 5.6 Phase 3: GitHub Actions 자동 릴리즈 (1502줄부터)

**⚠️ 중요**: Phase 3은 **자동화됨** (로컬에서 할 작업 없음)

**자동 실행 워크플로우**:

1. **moai-gitflow.yml** (PR merge 시 자동 트리거)
   - Release commit 감지 (🔖 RELEASE: 패턴)
   - Git Tag 생성
   - GitHub Release 생성 (Draft)

2. **release.yml** (GitHub Release published 시 자동 트리거)
   - 패키지 빌드: `uv build`
   - PyPI 배포: `uv publish with PYPI_API_TOKEN`

**모니터링 방법**:

```bash
# GitHub Actions 실행 상태 확인
gh run list --branch main --limit 5 --json name,status,conclusion

# GitHub Release Draft 확인
gh release list --limit 3

# PyPI 배포 확인
curl -s https://pypi.org/pypi/moai-adk/json | python3 -c "import sys, json; data=json.load(sys.stdin); print('Latest:', data['info']['version'])"
```

**완료 확인**:
```
✅ GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/v{new_version}
✅ PyPI: https://pypi.org/project/moai-adk/{new_version}
✅ Git 태그: git tag -l "v{new_version}"
✅ GitHub Actions: 모두 success
```

---

## 🔄 전체 릴리즈 플로우

### 일반 릴리즈 (Team 모드)

```
1. /awesome:release-new patch
         ↓
2. Phase 0: 품질 검증
   - pytest, ruff, mypy, bandit, pip-audit
   - 실패 시 → 중단
         ↓
3. Phase 1: 버전 분석
   - 현재 버전 확인
   - 목표 버전 결정
         ↓
4. Phase 1.5: 릴리즈 계획 보고서
   - 사용자 승인 대기
   - "진행" / "수정" / "중단" 선택
         ↓
5. Phase 2: PR 관리
   - GitHub PR 생성 (Draft)
   - CodeRabbit AI 리뷰
   - GitHub 웹에서 병합
         ↓
6. Phase 3: GitHub Actions 자동 실행
   - moai-gitflow.yml 트리거
   - release.yml 트리거
   - PyPI 배포 완료
         ↓
✅ 릴리즈 완료
```

### Dry-Run 모드

```
/awesome:release-new patch --dry-run
         ↓
Phase 0: 품질 검증 (실제 실행)
Phase 1: 버전 분석 (시뮬레이션)
Phase 1.5: 릴리즈 계획 보고서 (출력)
Phase 2: PR 관리 (건너뜀)
         ↓
🔬 시뮬레이션 완료 보고서
실제 릴리즈 명령 제시
```

### TestPyPI 배포

```
/awesome:release-new minor --testpypi
         ↓
Phase 0: 품질 검증 (동일)
Phase 1: 버전 분석 (동일)
Phase 1.5: 릴리즈 계획 (TestPyPI 표시)
Phase 2: PR 관리 (생략)
         ↓
Phase 3: TestPyPI에만 배포
         ↓
✅ TestPyPI 배포 완료
```

---

## 🛠️ 기술 스택

### 사용 도구

| 도구 | 용도 | 커맨드 |
|------|------|--------|
| **pytest** | 테스트 실행 | `pytest --cov` |
| **ruff** | 린트 검사 | `ruff check` |
| **mypy** | 타입 체크 | `mypy src/` |
| **bandit** | 보안 검사 | `bandit -r src/` |
| **pip-audit** | 의존성 보안 | `pip-audit` |
| **uv** | 패키지 관리 | `uv publish` |
| **gh** | GitHub CLI | `gh pr create` |
| **git** | 버전 제어 | `git tag` |

### 환경 변수

| 변수 | 용도 |
|------|------|
| `PYPI_API_TOKEN` | PyPI 인증 |
| `UV_PUBLISH_TOKEN` | uv 발행 토큰 |
| `UV_PUBLISH_TOKEN_TESTPYPI` | TestPyPI 토큰 |

---

## ⚙️ 버전 관리 구조

### SSOT 구조

```python
# pyproject.toml (유일한 진실의 출처)
[project]
name = "moai-adk"
version = "0.14.0"

# src/moai_adk/__init__.py (동적 로드)
from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version("moai-adk")  # 자동으로 0.14.0 읽기
except PackageNotFoundError:
    __version__ = "0.14.0-dev"
```

### 업데이트 방법

1. **pyproject.toml만 수정**:
   ```
   version = "0.13.0" → "0.13.1"
   ```

2. **__init__.py는 수정하지 않음**:
   - `importlib.metadata`가 자동으로 새 버전 읽음

---

## 📊 주요 특징 및 장점

### 1. 완전 자동화
- 품질 검증부터 PyPI 배포까지 자동 처리
- 수동 개입 최소화

### 2. 안전성
- 모든 단계에서 검증 (품질, 버전, 보안)
- Dry-Run으로 사전 검증 가능
- 실패 시 즉시 중단

### 3. 유연성
- Dry-Run 모드: 시뮬레이션
- TestPyPI: 테스트 환경 검증
- Personal/Team 모드: 프로젝트 구조에 맞춤

### 4. 투명성
- 상세한 릴리즈 계획 보고서 제공
- 변경사항 명확하게 요약
- 사용자 승인 필수

### 5. 확장성
- GitHub Actions로 추가 자동화 가능
- CodeRabbit AI 통합
- 다양한 배포 환경 지원

---

## 🚀 사용 시나리오

### 시나리오 1: 개발 완료, 릴리즈 준비

```bash
# 1단계: Dry-Run으로 확인
/awesome:release-new patch --dry-run

# 2단계: 계획 검토 후 실제 릴리즈
/awesome:release-new patch
```

### 시나리오 2: 테스트 환경에서 먼저 검증

```bash
# 1단계: TestPyPI로 배포
/awesome:release-new minor --testpypi

# 2단계: 설치 테스트
pip install --index-url https://test.pypi.org/simple/ moai-adk

# 3단계: 문제 없으면 실제 PyPI 배포
/awesome:release-new minor
```

### 시나리오 3: 개발 중 버전 확인

```bash
# Dry-Run으로 버전 계산만 확인
/awesome:release-new minor --dry-run

# 결과 보고서만 확인하고 중단
```

---

## ⚠️ 주의사항 및 제약 사항

### 1. 브랜치 전제 조건
- **Team 모드**: develop 브랜치 필수 존재
- **Personal 모드**: 브랜치 요구사항 없음

### 2. Git 상태
- 모든 변경사항 커밋된 상태
- 미커밋 파일이 있으면 경고

### 3. GitHub 권한
- GitHub CLI 인증 필수
- PR 생성/병합 권한 필수

### 4. PyPI 토큰
- `PYPI_API_TOKEN` 환경 변수 설정 필수
- TestPyPI의 경우 `UV_PUBLISH_TOKEN_TESTPYPI` 설정

### 5. GitHub Actions 워크플로우
- moai-gitflow.yml 필수
- release.yml 필수
- 두 워크플로우가 자동 트리거되어야 함

---

## 🔍 문제 해결

### 문제 1: Phase 0 검증 실패

**증상**: pytest, ruff, mypy, bandit, pip-audit 중 하나 실패

**해결**:
1. 로컬에서 해당 도구 실행하여 문제 확인
2. 코드 수정
3. 릴리즈 재시도

### 문제 2: GitHub Actions 미작동

**증상**: PR 병합 후 GitHub Actions 실행 안 됨

**해결**:
1. moai-gitflow.yml, release.yml 존재 확인
2. 워크플로우 활성화 확인
3. GitHub Actions 실행 로그 확인

### 문제 3: PyPI 배포 실패

**증상**: uv publish 실패

**해결**:
1. `PYPI_API_TOKEN` 환경 변수 확인
2. 토큰 만료 여부 확인
3. 패키지명 중복 여부 확인

---

## 📌 핵심 명령어 정리

| 명령어 | 용도 |
|--------|------|
| `/awesome:release-new patch` | patch 버전 릴리즈 |
| `/awesome:release-new minor` | minor 버전 릴리즈 |
| `/awesome:release-new major` | major 버전 릴리즈 |
| `/awesome:release-new patch --dry-run` | Dry-Run 시뮬레이션 |
| `/awesome:release-new minor --testpypi` | TestPyPI 배포 |
| `/awesome:release-new patch --dry-run --testpypi` | Dry-Run + TestPyPI |

---

## 결론

**release-new.md**는 moai-adk Python 패키지의 **엔드-투-엔드 릴리즈 자동화**를 구현한 포괄적인 지침입니다.

### 핵심 설계 원칙

1. **자동화 우선**: 모든 반복 가능한 작업 자동화
2. **안전성 우선**: 품질 검증과 사용자 승인 의무화
3. **투명성 우선**: 상세한 계획과 진행 상황 제시
4. **유연성 우선**: 다양한 배포 시나리오 지원

### 이 지침의 가치

- **시간 절감**: 수동 릴리즈 작업 완전 자동화
- **오류 감소**: 자동 검증으로 인적 오류 제거
- **재현성**: 동일한 프로세스로 일관성 있는 릴리즈
- **추적성**: 모든 릴리즈 단계 기록 및 모니터링

---

**마지막 업데이트**: 2025-11-04
**파일 경로**: `.claude/commands/alfred/release-new.md`
**동기화 상태**: ⛔️ 로컬 전용 (템플릿 동기화 금지)
