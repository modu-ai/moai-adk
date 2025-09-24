# SPEC-008: MoAI-ADK v0.1.0 Production Release

## @REQ:RELEASE-001 배경

### 현재 상황
- 버전 불일치: pyproject.toml(0.2.1) ≠ _version.py(0.2.2) ≠ VERSION(0.2.1)
- 템플릿 동기화 실패: 프로젝트 .claude/.moai와 templates/ 불일치
- TAG 추적성 부족: 소스 코드 16-Core @TAG 커버리지 0%
- 레거시 코드 존재: 사용하지 않는 함수, 중복 코드, TODO 마크 다수

### 해결해야 할 문제
1. **버전 관리 혼란**: 3개 파일의 서로 다른 버전 정보
2. **설치형 패키지 불완전**: 템플릿이 실제 프로젝트와 동기화되지 않음
3. **추적성 공백**: 요구사항부터 구현까지 TAG 체인 단절
4. **코드 품질**: 프로덕션 수준에 미달하는 코드 정리 필요

## @DESIGN:PKG-001 설계 목표

### 1. 정식 버전 0.1.0 출시
```yaml
Version Strategy:
  Current: 0.2.x (Beta/Development)
  Target:  0.1.0 (Production/Stable)
  Rationale: 첫 정식 릴리즈로서 의미적 버전 부여
  Status: "Development Status :: 5 - Production/Stable"
```

### 2. 완전한 패키지 재구성
```
MoAI-ADK Architecture v0.1.0:
├── 통일된 버전 관리
├── 동기화된 템플릿 시스템
├── 100% TAG 추적성
├── 클린 코드 표준
└── 자동화된 CI/CD
```

### 3. 설치형 템플릿 완성
- `.claude/`: 현재 프로젝트 → templates/ 완전 복사
- `.moai/`: 현재 프로젝트 → templates/ 완전 복사
- `CLAUDE.md`: 최신 버전 → templates/ 동기화

### 4. 16-Core TAG 완전 적용
```markdown
Primary Chain:
@REQ:XXX-001 → @DESIGN:XXX-001 → @TASK:XXX-001 → @TEST:XXX-001

Implementation Chain:
@FEATURE:XXX-001 → @API:XXX-001 → @UI:XXX-001 → @DATA:XXX-001

Quality Chain:
@PERF:XXX-001 → @SEC:XXX-001 → @DOCS:XXX-001 → @TAG:XXX-001

Project Chain:
@VISION:XXX-001 → @STRUCT:XXX-001 → @TECH:XXX-001 → @ADR:XXX-001
```

## @TASK:IMPL-001 구현 명세

### Phase 1: 버전 통일 시스템
**목표**: 모든 버전 파일을 0.1.0으로 통일

#### 1.1 버전 파일 수정
```python
# pyproject.toml
version = "0.1.0"
classifiers = ["Development Status :: 5 - Production/Stable"]

# src/moai_adk/_version.py
__version__ = "0.1.0"
VERSIONS = {
    "moai_adk": "0.1.0",
    "core": "0.1.0",
    "templates": "0.1.0",
    "hooks": "0.1.0",
    "agents": "0.1.0",
    "commands": "0.1.0",
    # ... 모든 컴포넌트 0.1.0 통일
}

# src/moai_adk/resources/VERSION
0.1.0
```

#### 1.2 CHANGELOG.md 업데이트
```markdown
# Changelog

## [0.1.0] - 2024-09-24
### 🎉 First Production Release
- 정식 버전 0.1.0 출시
- 완전한 패키지 재구성 완료
- 100% TAG 추적성 구현
- 설치형 템플릿 시스템 완성
```

### Phase 2: 코드 클린업
**목표**: 레거시 코드 제거 및 품질 개선

#### 2.1 제거 대상 파일/함수
```python
# core/validator.py - 중복 검증 함수들
- duplicate_validation_function()
- legacy_check_method()

# cli/helpers.py - 미사용 유틸리티
- unused_helper_function()
- temporary_debug_util()

# install/post_install.py - 레거시 코드
- old_installation_method()
- deprecated_setup_function()
```

#### 2.2 통합 대상 모듈
```python
# config.py + core/config_manager.py → core/config_manager.py
# 중복된 설정 관리 로직 통합

# core/exceptions.py 정리
# 미사용 예외 클래스 제거: LegacyError, DeprecatedWarning
```

### Phase 3: 템플릿 동기화 시스템
**목표**: 현재 프로젝트와 templates/ 완전 동기화

#### 3.1 동기화 구현
```python
# core/template_sync.py (신규 생성)
class TemplateSync:
    def sync_claude_directory(self):
        """현재 .claude/ → templates/.claude/ 복사"""

    def sync_moai_directory(self):
        """현재 .moai/ → templates/.moai/ 복사"""

    def sync_claude_md(self):
        """현재 CLAUDE.md → templates/CLAUDE.md 복사"""

    def validate_sync(self):
        """동기화 상태 검증"""
```

#### 3.2 설치 시 템플릿 적용
```python
# install/installer.py 수정
def install_templates(self):
    """templates/ → 타겟 프로젝트로 설치"""
    self.copy_template_directory(".claude")
    self.copy_template_directory(".moai")
    self.copy_template_file("CLAUDE.md")
```

### Phase 4: 16-Core TAG 적용
**목표**: 모든 소스 파일에 TAG 추가

#### 4.1 TAG 적용 예시
```python
"""CLI command entry points

@REQ:CLI-001 Command line interface requirements
@DESIGN:CMD-001 Command pattern implementation design
@TASK:INIT-001 Initialize command implementation
@TEST:CLI-001 CLI command tests
"""

class Commands:
    """Main CLI commands handler

    @FEATURE:CLI-001 CLI command execution
    @API:CMD-001 Command API interface
    """

    def init_command(self):
        """Initialize MoAI project

        @TASK:INIT-001 Project initialization implementation
        @TEST:INIT-001 Initialization test coverage
        """
```

### Phase 5: TestPyPI 배포 및 검증
**목표**: TestPyPI를 통한 안전한 배포 검증

#### 5.1 TestPyPI 배포 설정
```yaml
# .github/workflows/test-release.yml
name: Test Release
on:
  push:
    branches: [ develop ]

jobs:
  test-release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install build dependencies
        run: pip install build twine

      - name: Build package
        run: python -m build

      - name: Test install from TestPyPI
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.TEST_PYPI_TOKEN }}
        run: |
          twine upload --repository testpypi dist/*
          pip install --index-url https://test.pypi.org/simple/ moai-adk
```

## @TEST:ACCEPT-001 수락 기준

### 1. 버전 통일 확인
```bash
Given: 버전 불일치 상태 (0.2.1 ≠ 0.2.2)
When: 버전 통일 스크립트 실행
Then: 모든 버전 파일이 0.1.0으로 통일됨
```

### 2. 템플릿 동기화 확인
```bash
Given: 프로젝트 .claude/.moai와 templates/ 불일치
When: 템플릿 동기화 실행
Then: diff 명령어로 완전 일치 확인됨
```

### 3. TAG 추적성 확인
```bash
Given: 소스 코드 TAG 커버리지 0%
When: TAG 적용 및 인덱스 생성
Then: tags.json에서 100% 커버리지 확인됨
```

### 4. TestPyPI 설치 확인
```bash
Given: TestPyPI에서 설치 시도
When: pip install --index-url https://test.pypi.org/simple/ moai-adk==0.1.0
Then: 성공적 설치 및 moai 명령어 실행 가능
```

### 5. 품질 게이트 통과
```bash
Given: 코드 품질 검사
When: make test && make validate
Then:
  - 테스트 커버리지 ≥ 85%
  - 개발 가이드 위반 0건
  - 빌드 에러 0건
```

## @PERF:OPT-001 성능 최적화

### 빌드 시간 단축
- 불필요한 파일 제거로 패키지 크기 감소
- 병렬 테스트 실행으로 CI 시간 단축

### 설치 속도 개선
- 의존성 최적화
- 템플릿 복사 효율화

## @SEC:SECURITY-001 보안 고려사항

### 민감정보 보호
- TestPyPI 토큰은 GitHub Secrets 사용
- 소스 코드에서 하드코딩된 정보 제거

### 코드 검증
- 모든 템플릿 파일 보안 스캔
- 설치 시 권한 검증

## @DOCS:DOC-001 문서화

### README.md 업데이트
- 0.1.0 정식 릴리즈 안내
- 새로운 설치 방법 가이드

### API 문서 생성
- 주요 모듈 docstring 완성
- 사용자 가이드 업데이트

## @TAG:TRACE-001 추적성 보장

### TAG 체인 완성
```markdown
@REQ:RELEASE-001 → @DESIGN:PKG-001 → @TASK:IMPL-001 → @TEST:ACCEPT-001
└── @FEATURE:TEMPLATE-001 → @API:SYNC-001 → @DATA:INDEX-001
└── @PERF:OPT-001 → @SEC:SECURITY-001 → @DOCS:DOC-001
```

### 이력 관리
- 모든 변경사항 CHANGELOG.md 기록
- Git 태그와 버전 연결
- 커밋 메시지에 SPEC-008 참조

---

**이 SPEC은 MoAI-ADK의 첫 정식 릴리즈를 위한 완전한 패키지 재구성을 정의합니다.**