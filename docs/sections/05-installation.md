# MoAI-ADK 설치 및 초기화

> **완전 자동화된 설치 시스템** - pip 기반 PyPI 패키지 설치
> **Last Updated**: 2025-09-16 | **Package Version**: v0.1.15
> **Difficulty**: 🟢 Basic

## 🚀 설치 과정 개요

MoAI-ADK 설치는 2단계로 구성됩니다:

1. **Stage 1**: Python 패키지 설치
2. **Stage 2**: 프로젝트 초기화

## Stage 1: Python 패키지 설치

### 기본 설치

```bash
# pip을 통한 설치
pip install moai-adk

# 또는 개발 모드 설치
cd moai_adk
pip install -e .

# 설치 확인
python -m moai_adk --version
# 또는
moai --version
```

### 시스템 요구사항

- **Python**: 3.8 이상 (3.8-3.13 완전 호환성 검증 완료)
- **운영체제**: Windows, macOS, Linux
- **디스크 공간**: 50MB (전역 리소스)
- **권한**: 일반 사용자 권한 (Windows에서 심볼릭 링크 사용 시 관리자 권한 권장)

### ✅ v0.1.15 안정성 보장

**완전히 테스트된 안정 버전**:
- 🧪 **7가지 핵심 기능** 완전 동작 검증
- 🐍 **Python 3.8-3.13** 교차 호환성 확인
- 🚫 **치명적 버그 5개** 완전 수정
- ⚡ **설치 성공률 100%** 달성

## Stage 2: 프로젝트 초기화

### 새 프로젝트 생성

```bash
# 새 프로젝트 생성
moai init myapp

# 프로젝트 디렉토리로 이동
cd myapp
```

### 기존 프로젝트에 설치

```bash
# 기존 프로젝트에 설치
cd existing-project
moai init .
```

## 📋 17단계 초기화 프로세스

초기화 과정에서 다음 17단계가 자동으로 실행됩니다:

### 1. 디렉토리 구조 생성
- `.claude/` 및 `.moai/` 디렉토리 구조 생성
- 필요한 하위 디렉토리 자동 구성

### 2. 기존 코드 자동 스캔
- `pyproject.toml`, `setup.py`, `requirements.txt` 분석
- 프로젝트 언어 및 프레임워크 자동 감지

### 3. 스캔 결과 프리필
- `/moai:1-project init` 마법사에 자동 반영
- 감지된 정보로 질문 사전 답변

### 4. AI 에이전트 시스템 설치
- 11개 전문 에이전트 설치 (`.claude/agents/moai/`)
- 에이전트별 역할과 도구 설정

### 5. 슬래시 명령어 설치
- 6개 MoAI 명령어 설치 (`.claude/commands/moai/`)
- 연번순 명령어 체계 구성

### 6. MoAI Hook 스크립트 구성
- 5개 핵심 Hook 설치
- `settings.json`으로 Hook 설정 통합

### 7. 문서 템플릿 시스템 설치
- SPEC, PLAN, TASKS 템플릿 설치
- 동적 템플릿 엔진 구성

### 8. 메모리 시스템 설치
- 프로젝트 가이드라인, Constitution, ADR 템플릿 설치
- Claude Code 메모리 파일 구성

### 9. GitHub CI/CD 워크플로우 설치
- Constitution 자동 검증 파이프라인 설치
- 다중 언어 지원 구성

### 10. 검증 스크립트 설치
- 9개 품질 검증 도구 설치 (`.moai/scripts/`)
- 통합 테스트 실행 스크립트 구성

### 11. 14-Core TAG 시스템 초기화
- 추적성 매트릭스 및 인덱스 생성
- TAG 무결성 검사 시스템 구성

### 12. Claude Code 설정 최적화
- MoAI 통합 설정 적용
- Hook 시스템 활성화

### 13. Constitution 5원칙 구성
- 품질 보장 시스템 활성화
- 자동 검증 게이트 설정

### 14. Steering 문서 생성
- 프로젝트 방향성 문서 템플릿 준비
- 동적 생성 시스템 구성

### 15. 프로젝트 메모리 생성
- `CLAUDE.md` 시스템 구성
- 프로젝트별 메모리 설정

### 16. 자동 버전 관리 시스템 설치 (v0.1.14)
- `scripts/update_version.py`: 독립실행형 버전 관리 스크립트
- `TemplateEngine`: 자동 버전 변수 주입 시스템
- `CLI update-version`: 개발자용 버전 동기화 명령어
- 24개 파일 패턴 매칭 규칙 적용

### 17. MoAI Output Styles 설치
- 5개 맞춤형 스타일 설치:
  - `expert.md`: 간결하고 효율적인 전문가 모드
  - `beginner.md`: 상세한 설명과 단계별 안내
  - `study.md`: 깊이 있는 원리와 심화 학습
  - `mentor.md`: 1:1 멘토링과 페어 프로그래밍
  - `audit.md`: 코드 품질 지속적 검증 개선

## 🌟 패키지 내장 리소스 시스템 (v0.1.13)

### 패키지 내장 리소스 분석

MoAI-ADK v0.1.15부터 패키지 내장 리소스 시스템을 사용합니다:

```python
# 패키지 내장 리소스 접근
from importlib import resources
self.resources_root = resources.files('moai_adk.resources')
self.templates_root = self.resources_root / 'templates'

# 각 프로젝트로 복사되는 리소스
.claude/agents/moai/      # 11개 에이전트 파일
.claude/commands/moai/    # 6개 슬래시 명령어
.moai/templates/          # 문서 템플릿들
```

### 파일 복사 아키텍처

- **완전 독립성**: 각 프로젝트가 완전히 독립된 파일 복사본 사용
- **크로스 플랫폼**: Windows/macOS/Linux 모든 환경에서 동일한 동작
- **안정성**: shutil.copytree 기반의 안전한 파일 복사

### 설치 및 활용

```bash
# 표준 설치 (모든 플랫폼)
moai init project

# 권한 확인
moai status -v  # 상세 상태 확인 (심볼릭 링크 포함)
```

## 🛠️ 설치 후 확인

### 기본 상태 확인

```bash
# 간단한 상태 확인
moai status

# 상세 상태 확인 (심볼릭 링크 포함)
moai status -v
```

### 시스템 무결성 검사

```bash
# 전체 시스템 검증
python .moai/scripts/run-tests.sh

# 개별 검증
python .moai/scripts/check_constitution.py
python .moai/scripts/validate_tags.py
python .moai/scripts/check-traceability.py
```

## 🔄 업데이트 시스템 (v0.1.14)

### 사용자용 업데이트

```bash
# 완전 자동 업데이트 (패키지 + 리소스)
moai update

# 업데이트 가능 여부 확인
moai update --check

# 선택적 업데이트
moai update --package-only     # 패키지만 업그레이드
moai update --resources-only   # 글로벌 리소스만 업데이트
```

### 개발자용 버전 관리

```bash
# 전체 버전 동기화
moai update-version 0.1.14

# 안전한 사전 테스트
moai update-version 0.1.14 --dry-run

# 검증 포함
moai update-version 0.1.14 --verify

# Git 커밋 제외
moai update-version 0.1.14 --no-git
```

## 🚨 문제 해결

### 일반적인 설치 문제

1. **권한 오류 (Windows)**
   ```bash
   # 해결방법: 파일 복사 모드로 설치
   moai init . --force-copy
   ```

2. **Python 버전 호환성**
   ```bash
   # Python 3.8+ 확인
   python --version

   # 가상환경 사용 권장
   python -m venv moai-env
   source moai-env/bin/activate  # Linux/macOS
   # 또는
   moai-env\Scripts\activate     # Windows
   ```

3. **기존 설치와 충돌**
   ```bash
   # 기존 설치 제거 후 재설치
   pip uninstall moai-adk
   pip install moai-adk
   ```

### Hook 실행 문제

```bash
# Hook 권한 설정 (Linux/macOS)
chmod +x .claude/hooks/moai/*.py

# Hook 테스트
python .claude/hooks/moai/test_hook.py
```

### TAG 시스템 문제

```bash
# TAG 무결성 자동 수정
python .moai/scripts/repair_tags.py --execute

# 추적성 검증
python .moai/scripts/check-traceability.py --verbose
```

## ✅ 설치 완료 체크리스트

- [ ] `moai --version` 명령어 정상 동작
- [ ] `moai status` 모든 컴포넌트 정상
- [ ] `.claude/` 디렉토리 구조 완성
- [ ] `.moai/` 디렉토리 구조 완성
- [ ] Hook 시스템 정상 동작
- [ ] TAG 시스템 초기화 완료
- [ ] CI/CD 파이프라인 설정 완료

설치 완료 후 `/moai:1-project init` 명령어로 프로젝트별 설정을 진행하세요.

## 📚 관련 문서

- **[시스템 개요](01-overview.md)**: MoAI-ADK 전체 소개 및 주요 기능
- **[패키지 구조](package-structure.md)**: 설치된 패키지의 내부 구조 이해
- **[아키텍처](04-architecture.md)**: 생성되는 프로젝트 구조 상세 설명
- **[명령어 시스템](08-commands.md)**: 설치 후 사용 가능한 CLI 명령어
- **[대화형 마법사](06-wizard.md)**: 프로젝트 초기화 및 설정
- **[빌드 시스템](build-system.md)**: 개발 환경 설정 및 빌드