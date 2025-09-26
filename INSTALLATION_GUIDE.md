# MoAI-ADK 설치 가이드

## 📋 시스템 요구사항

- **Python**: 3.10 이상 (3.11, 3.12, 3.13 권장)
- **운영체제**: Windows, macOS, Linux
- **패키지 관리자**: pip (20.0+) 또는 uv (권장)

## 🚀 표준 설치 (PyPI)

```bash
# 최신 stable 버전 설치
pip install moai-adk

# 또는 uv 사용 (더 빠름)
uv pip install moai-adk

# 설치 확인
moai --version
```

## 🧪 개발 버전 설치 (TestPyPI)

TestPyPI에서 최신 개발 버전을 설치하려면 다음 단계를 따르세요:

### 1단계: 기존 패키지 제거

```bash
# 기존 설치된 moai-adk 제거
pip uninstall -y moai-adk

# pip 캐시 정리 (선택사항, 문제 발생 시 권장)
pip cache purge
```

### 2단계: TestPyPI에서 설치

**권장 방법** (모든 의존성 정상 설치):
```bash
pip install --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk
```

**TestPyPI 전용 설치** (일부 의존성 누락 가능):
```bash
pip install --no-cache-dir -i https://test.pypi.org/simple/ moai-adk
```

### 3단계: 설치 확인

```bash
# 버전 확인
moai --version

# 기본 기능 테스트
moai init test-project
cd test-project
ls -la .moai .claude
```

## 🔍 문제 해결

### Windows 사용자 공통 문제

#### 문제 1: Python 3.10에서 v0.1.7 설치됨

**증상**: `pip install moai-adk`가 v0.1.7을 설치함

**원인**: Python 3.10 사용자가 이전 버전 요구사항(>=3.11)으로 인해 구 버전 설치

**해결책**: v0.1.25부터 Python 3.10 지원 복원
```bash
# 기존 버전 제거
pip uninstall -y moai-adk

# 최신 버전 재설치
pip install --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk

# 버전 확인 (0.1.25 이상이어야 함)
moai --version
```

#### 문제 2: 의존성 백트래킹 경고

**증상**: pip이 여러 버전을 시도하며 긴 백트래킹 과정 실행

**원인**: TestPyPI에 일부 의존성(jsonschema 등) 누락

**해결책**: `--extra-index-url` 옵션 사용
```bash
pip install --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk
```

#### 문제 3: 권한 오류

**증상**: `Permission denied` 또는 `Access is denied`

**해결책**: 사용자 디렉토리에 설치
```bash
pip install --user --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk
```

### 일반적인 문제

#### 캐시 관련 문제

```bash
# pip 캐시 완전 정리
pip cache purge

# uv 캐시 정리
uv cache clean

# 재설치
pip install --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk
```

#### 네트워크 문제

```bash
# 타임아웃 늘리기
pip install --timeout=300 --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk

# 프록시 사용 (필요한 경우)
pip install --proxy http://proxy.company.com:8080 --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk
```

## 🔄 버전별 Python 호환성

| MoAI-ADK 버전 | Python 요구사항 | 비고 |
|--------------|---------------|------|
| 0.1.7 이하    | >=3.8         | 레거시 버전 |
| 0.1.8 ~ 0.1.24 | >=3.11      | 호환성 문제 존재 |
| **0.1.25+**  | **>=3.10**    | **권장 버전** |

## 📊 설치 성공 확인 체크리스트

- [ ] `moai --version`이 0.1.25 이상 출력
- [ ] `moai init test-project` 명령 성공
- [ ] `.moai/` 및 `.claude/` 디렉토리 생성 확인
- [ ] `moai config status` 명령 정상 동작

## 🆘 추가 지원

설치 문제가 계속되면:

1. **GitHub Issues**: https://github.com/modu-ai/moai-adk/issues
2. **Python 환경 정보** 포함:
   ```bash
   python --version
   pip --version
   pip show moai-adk
   ```
3. **전체 설치 로그** 첨부

## 🎯 빠른 설치 요약

**Python 3.10+ 사용자 (권장)**:
```bash
pip uninstall -y moai-adk
pip install --no-cache-dir -i https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple moai-adk
moai --version
```

이 방법으로 최신 개발 버전을 안정적으로 설치할 수 있습니다.