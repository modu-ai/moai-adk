# MoAI-ADK Technology Stack

## @TECH:STACK-001 언어 & 런타임

### Python ≥3.11 (주 언어)

- **선택 이유**: Claude Code 호환성, 풍부한 생태계, 크로스 플랫폼 지원
- **지원 버전**: 3.11, 3.12, 3.13 (pyproject.toml 기준)
- **배포 타겟**: PyPI, conda-forge (계획)
- **설치 요구사항**: pip ≥21.0, setuptools ≥61.0

### 멀티 플랫폼 지원 현황

| 플랫폼      | 지원 상태 | 검증 도구          | 주요 제약              |
| ----------- | --------- | ------------------ | ---------------------- |
| **Windows** | ✅ 완료   | colorama 기반 콘솔 | 경로 구분자 처리       |
| **macOS**   | ✅ 완료   | 기본 터미널        | 권한 설정 주의         |
| **Linux**   | ✅ 완료   | 표준 bash          | 배포판별 패키지 관리자 |

## @TECH:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. CLI 프레임워크

```toml
dependencies = [
    "click>=8.0.0",      # 명령어 인터페이스
    "colorama>=0.4.6",   # 크로스 플랫폼 콘솔 색상
    "toml>=0.10.0",      # 설정 파일 파싱
    "watchdog>=3.0.0",   # 파일 시스템 감시
]
```

### 2. 개발 도구체인

```toml
dev = [
    "pytest>=7.0",       # 테스트 프레임워크
    "pytest-cov>=4.0",   # 커버리지 측정
    "black>=22.0",       # 코드 포매터
    "isort>=5.10",       # import 정렬
    "mypy>=1.0",         # 정적 타입 검사
    "flake8>=5.0",       # 린터
]
```

### 3. 빌드 시스템

- **패키지 빌드**: setuptools + wheel
- **버전 관리**: 자체 개발 `version_sync.py` 시스템
- **배포 자동화**: GitHub Actions + PyPI API

## @TECH:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지 목표

- **현재 상태**: `.coverage` 파일 기준 측정
- **목표**: 85% 이상 (TRUST 원칙 기준)
- **측정 도구**: pytest-cov
- **실패 시 대응**: 회귀 테스트 추가, Waiver 검토

### 정적 분석 도구

| 도구       | 역할        | 설정 파일      | 실패 시 조치         |
| ---------- | ----------- | -------------- | -------------------- |
| **black**  | 코드 포매터 | pyproject.toml | 자동 수정            |
| **isort**  | import 정렬 | pyproject.toml | 자동 수정            |
| **mypy**   | 타입 검사   | pyproject.toml | 타입 어노테이션 추가 |
| **flake8** | 린터        | .flake8        | 코드 스타일 수정     |

### Makefile 기반 자동화

```bash
# 품질 검사 파이프라인
make build        # 버전 동기화 + 패키지 빌드
make test         # 종합 테스트 슈트
make validate     # 설정 파일 검증
make release      # 프로덕션 배포 준비
```

## @TECH:SECURITY-001 보안 정책 & 운영

### 비밀 관리

- **정책**: 환경 변수 기반, 코드 내 하드코딩 금지
- **검증 도구**: `check_secrets.py` 스크립트
- **Claude Code 통합**: `.claude/hooks/moai/pre_write_guard.py`

### 접근 제어

```json
// .claude/settings.json 예시
{
  "defaultMode": "acceptEdits",
  "overrides": {
    "claude/hooks/moai/policy_block.py": "ask",
    "critical_files": "deny"
  }
}
```

### 로깅 정책

- **구조화 로깅**: JSON Lines 포맷
- **로그 수준**: INFO (기본), DEBUG (개발)
- **민감정보 마스킹**: `***redacted***` 패턴 적용
- **감사 로그**: 모든 파일 변경, Git 작업 기록

### 보안 스크립트

| 스크립트                | 기능                   | 실행 주기      | 책임자      |
| ----------------------- | ---------------------- | -------------- | ----------- |
| `check_secrets.py`      | 비밀정보 검사          | pre-commit     | 개발자      |
| `check_constitution.py` | 개발 가이드 위반 검사 | 세션 시작      | MoAI 시스템 |
| `check_licenses.py`     | 라이선스 호환성        | 종속성 변경 시 | 관리자      |

## @TECH:DEPLOY-001 배포 채널 & 전략

### 1. PyPI (주 배포 채널)

- **패키지명**: `moai-adk`
- **릴리스 절차**: GitHub Actions 자동화
- **버전 정책**: Semantic Versioning (0.2.1 → 0.2.2)
- **rollback 전략**: PyPI에서 버전 삭제 후 재배포

### 2. 개발 설치

```bash
# 개발자 모드
pip install -e .

# 테스트 설치
pip install moai-adk[test]

# 전체 개발 환경
pip install moai-adk[dev]
```

### 3. 크로스 플랫폼 배포 완성 계획

| 채널            | 현재 상태 | 목표               | 배포 주기          |
| --------------- | --------- | ------------------ | ------------------ |
| **PyPI**        | ✅ 완료   | 안정 배포          | stable 태그 시     |
| **conda-forge** | 🔄 계획   | 과학 컴퓨팅 생태계 | PyPI 릴리스 후     |
| **Homebrew**    | 🔄 계획   | macOS 편의성       | 주요 버전 업데이트 |
| **winget**      | 🔄 계획   | Windows 패키지     | 주요 버전 업데이트 |

## Legacy Context

### 기술 스택 현황 요약

- **완성된 Python 패키지**: setuptools + wheel 기반
- **풍부한 개발 도구**: black, mypy, pytest 등 표준 도구체인
- **자동화된 빌드**: Makefile + Python 스크립트 조합
- **Claude Code 완전 통합**: 훅/에이전트 시스템

### @DEBT:TEST-COVERAGE-001 품질 개선 부채

1. **현재 커버리지 상태 불명**: 정확한 수치 측정 필요
2. **E2E 테스트 부족**: CLI 명령어 전체 시나리오 테스트 추가
3. **크로스 플랫폼 검증 부족**: Windows/Linux 환경 자동 테스트

### @TASK:CROSS-PLATFORM-001 크로스 플랫폼 완성 계획

1. **Windows 호환성 검증**: 경로 구분자, 권한 설정 테스트
2. **conda-forge 패키지 등록**: 과학 컴퓨팅 생태계 진입
3. **패키지 관리자 통합**: Homebrew, winget 지원 추가
4. **자동화된 배포**: 멀티 채널 동시 배포 파이프라인

### @TODO:TECH-UPGRADE-001 기술 스택 업그레이드 후보

1. **Python 3.14 지원**: 차기 Python 버전 호환성 준비
2. **Rich/Textual 도입**: CLI UX 개선 (진행률 바, 테이블 등)
3. **Pydantic v2**: 설정 검증 및 타입 안전성 강화
4. **GitHub Actions 최적화**: 빌드 시간 단축, 캐시 전략 개선

## 환경별 설정

### 개발 환경 (`dev`)

```bash
make setup          # 전체 개발 환경 구성
make dev            # 파일 감시 모드
make test-verbose   # 상세 테스트 실행
```

### CI/CD 환경 (`ci`)

```bash
make test-ci        # JUnit 리포트 + 커버리지
make build-clean    # 클린 빌드
make validate       # 설정 파일 검증
```

### 프로덕션 환경 (`production`)

```bash
make release        # 배포 준비 (테스트 + 검증 + 빌드)
```

---

_이 기술 스택은 `/moai:2-build` 실행 시 TDD 도구 선택과 품질 게이트 적용의 기준이 됩니다._
