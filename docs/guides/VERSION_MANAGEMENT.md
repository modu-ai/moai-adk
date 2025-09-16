# 🗿 MoAI-ADK 버전 관리 시스템

MoAI-ADK는 **완전 자동화된 버전 관리 시스템**을 제공합니다. 단일 소스에서 전체 프로젝트의 버전 정보를 일관되게 관리할 수 있습니다.

## 🎯 핵심 개념

### 중앙집중식 버전 관리
- **단일 소스**: `src/_version.py` 파일이 모든 버전 정보의 원천
- **자동 동기화**: 버전 변경 시 관련된 80여개 파일이 자동으로 업데이트
- **일관성 보장**: 버전 불일치 방지 및 실시간 검증

### 지원하는 버전 패턴
- Python 패키지: `pyproject.toml`, `__version__` 변수
- 소스 코드: 모든 `.py` 파일의 버전 참조
- 문서: 마크다운 파일의 버전 표기
- 설정: JSON, YAML 파일의 버전 필드  
- CI/CD: GitHub Actions 워크플로우
- 빌드: Makefile, 스크립트 파일

## 🚀 사용 방법

### 1. 현재 버전 확인
```bash
make version
# 출력: 🗿 MoAI-ADK v0.1.16
```

### 2. 버전 업데이트 (권장)
```bash
# 새 버전으로 업데이트
make version-set VERSION=0.2.0

# 결과:
# ✅ Updated _version.py to 0.2.0
# 🔄 Synchronizing version across all files...
# ✅ Version update completed to 0.2.0
```

### 3. 버전 동기화 확인
```bash
# 드라이 런 (실제 변경 없이 미리보기)
make version-sync-dry

# 실제 동기화 실행
make version-sync

# 일관성 검증
make version-verify
```

## 🔄 자동화 워크플로우

### GitHub Actions 자동 동기화
`src/_version.py` 파일이 변경되면 자동으로:

1. **변경 감지**: _version.py 파일 수정 감지
2. **전체 동기화**: 모든 관련 파일 업데이트
3. **일관성 검증**: 버전 불일치 확인
4. **자동 커밋**: 변경사항 자동 커밋
5. **태그 생성**: 릴리즈 태그 자동 생성 (main 브랜치)

### 수동 실행도 가능
```bash
# GitHub Actions 워크플로우 수동 실행
# Repository → Actions → Version Sync → Run workflow
# 버전 입력: 0.2.0
```

## 📁 관리 대상 파일들

### Python 패키지
- `pyproject.toml` - 패키지 메타데이터 버전
- `src/_version.py` - 중앙 버전 저장소
- `src/*.py` - 모든 Python 소스 파일

### 설정 파일
- `**/*.json` - JSON 설정의 version, moai_version 필드
- `.moai/config.json` - MoAI 설정 버전
- `.claude/settings.json` - Claude Code 설정

### 문서
- `README.md` - 프로젝트 설명 및 배지
- `CHANGELOG.md` - 변경 이력
- `**/*.md` - 모든 마크다운 문서

### CI/CD
- `.github/workflows/*.yml` - GitHub Actions 워크플로우
- `Makefile` - 빌드 시스템
- 각종 스크립트 파일

## 🛠️ 고급 사용법

### 버전 정보 프로그래밍 방식 접근
```python
from src._version import __version__, get_version, get_version_format

# 기본 버전
print(__version__)  # "0.1.16"

# 컴포넌트별 버전
print(get_version("core"))  # "0.1.16"
print(get_version("hooks")) # "0.1.16"

# 포맷된 버전 문자열
print(get_version_format("banner"))  # "🗿 MoAI-ADK v0.1.16"
print(get_version_format("short"))   # "v0.1.16"
```

### 커스텀 버전 패턴 추가
`src/version_sync.py`의 `_load_version_patterns()` 메서드에서 새 패턴 추가:

```python
"**/*.py": [
    {
        "pattern": r'MY_VERSION\s*=\s*"[^"]*"',
        "replacement": f'MY_VERSION = "{self.current_version}"',
        "description": "Custom version pattern"
    }
]
```

## ⚡ 빠른 시작

### 개발 중 버전 업데이트
```bash
# 1. 패치 버전 업데이트 (0.1.4 → 0.1.5)
make version-set VERSION=0.1.5

# 2. 마이너 버전 업데이트 (0.1.5 → 0.2.0) 
make version-set VERSION=0.2.0

# 3. 메이저 버전 업데이트 (0.2.0 → 1.0.0)
make version-set VERSION=1.0.0

# 4. Git에 반영
git add -A
git commit -m "bump version to v1.0.0"  
git tag v1.0.0
git push origin main --tags
```

### 릴리즈 전 최종 검증
```bash
# 1. 버전 일관성 확인
make version-verify

# 2. 전체 테스트
make test

# 3. 빌드 검증
make build

# 4. 릴리즈 준비 완료
make release
```

## 🔧 문제 해결

### 버전 불일치 발생 시
```bash
# 1. 현재 상황 파악
make version-verify

# 2. 강제 동기화
make version-sync

# 3. 재검증
make version-verify
```

### GitHub Actions 실패 시
```bash
# 1. 로컬에서 동기화 테스트
make version-sync-dry

# 2. _version.py 직접 확인
cat src/_version.py | grep __version__

# 3. 수동 동기화 후 푸시
make version-sync
git add -A && git commit -m "fix: version sync" && git push
```

### 새 파일에 버전 추가 시
새로운 템플릿이나 설정 파일을 추가할 때는 `src/version_sync.py`의 패턴 정의에 추가하세요:

```python
# 새 파일 타입 추가 예시
"**/*.yaml": [
    {
        "pattern": r'version:\s*[0-9]+\.[0-9]+\.[0-9]+',
        "replacement": f'version: {self.current_version}',
        "description": "YAML version field"
    }
]
```

## 📊 버전 관리 모범 사례

### 시맨틱 버전 관리 따르기
- **MAJOR.MINOR.PATCH** 형식 준수
- **MAJOR**: 호환성이 깨지는 변경사항
- **MINOR**: 새 기능 추가 (하위 호환성 유지)
- **PATCH**: 버그 수정

### 커밋 메시지 규칙
```bash
# 버전 업데이트 커밋
git commit -m "bump version to v0.2.0"

# 기능 추가 후
git commit -m "feat: add new agent system

bump version to v0.2.0"
```

### 브랜치 전략
- `main`: 안정 버전만 (자동 태깅)
- `develop`: 개발 버전 (동기화만)
- `feature/*`: 기능 개발 (동기화 스킵)

---

**🗿 MoAI-ADK v0.1.16** - 완전 자동화된 버전 관리로 개발에만 집중하세요!
