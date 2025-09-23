# MoAI-ADK 설치 및 초기화

> **완전 자동화된 설치 시스템** - pip 기반 PyPI 패키지 설치
> **Last Updated**: 2025-09-23
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
moai --version
# 출력 예: MoAI-ADK 0.x.y
```

### 시스템 요구사항

- **Python**: 3.11+
- **운영체제**: Windows, macOS, Linux
- **디스크 공간**: 50MB (전역 리소스)
- **권한**: 일반 사용자 권한 (심볼릭 링크 불필요)

선택 의존성(모드별 권장):
- 개인 모드: `watchdog` (자동 체크포인트 파일 감시)
- 팀 모드: GitHub CLI(`gh`), Anthropic GitHub App (PR 자동화)

## Stage 2: 프로젝트 초기화

### 새 프로젝트 생성 (모드 선택)

```bash
# 새 프로젝트 생성 (개인 모드, 기본값)
moai init --personal myapp

# 프로젝트 디렉토리로 이동
cd myapp
```

### 기존 프로젝트에 설치 (모드 지정 가능)

```bash
# 기존 프로젝트에 설치
cd existing-project
moai init --team .

```

### 모드 전환/확인

모드 전환은 `/moai:0-project update` 마법사를 다시 실행해 팀 구성/협업 형태를 재조정하는 것이 기본입니다. CLI로 직접 전환해야 할 때만 아래 명령을 사용하세요.

```bash
moai config --mode team     # 개인 → 팀 전환 (수동)
moai config --mode personal # 팀 → 개인 전환 (수동)
moai config --show          # 현재 모드 출력
```

## 📋 초기화 프로세스(요약)

SimplifiedInstaller 기준 실제 초기화 단계는 아래와 같습니다.

1) 프로젝트 디렉토리 준비
- 현재 디렉토리 또는 지정 경로에 프로젝트 폴더 생성/재사용
- `--force`가 아닌 한 기존 파일은 보존 (`.git/`는 항상 보존)

2) Claude 리소스 설치 (.claude/)
- 패키지 내장 리소스를 복사하여 에이전트/명령/훅/메모리/스타일 배치

3) MoAI 리소스 설치 (.moai/)
- 문서/스크립트/인덱스/템플릿 복사
- `templates.mode=package`인 경우 `_templates/` 복사 생략(생성 시 패키지 템플릿 폴백)

4) 템플릿 버전 기록
- `.moai/version.json`에 템플릿/패키지 버전과 `last_updated` 기록

5) 보조 디렉토리 구성
- `.claude/logs`, `.moai/project`, `.moai/specs`, `.moai/reports` 등 빈 디렉토리 생성

6) GitHub 워크플로우 (옵션)
- `include_github=true`일 때 `.github/workflows/` 복사

7) 프로젝트 메모리 생성
- `CLAUDE.md` 설치 및 기술 스택 기반 메모리 템플릿 렌더링(`.moai/memory/*.md`)

8) 설정 파일 생성
- `.claude/settings.json`, `.moai/config.json` 생성 및 검증

9) Git 초기화 (옵션)
- 저장소 초기화, `.gitignore` 생성

10) 설치 검증 및 다음 단계 안내
- 필수 리소스 존재 검증, 가이드 출력

## 🌟 패키지 내장 리소스 시스템 (v0.1.13)

### 패키지 내장 리소스 분석

MoAI-ADK v0.1.17+부터 패키지 내장 리소스 시스템을 사용합니다:

```python
# 패키지 내장 리소스 접근
from importlib import resources
self.resources_root = resources.files('moai_adk.resources')
self.templates_root = self.resources_root / 'templates'

# 각 프로젝트로 복사되는 리소스 (기본)
.claude/agents/moai/      # MoAI 핵심/통합 에이전트(project-manager, spec-builder, code-builder, doc-syncer, git-manager, codex-bridge, gemini-bridge 등)
.claude/commands/moai/    # 6개 슬래시 명령어
.moai/_templates/         # 문서 템플릿들 (templates.mode=package일 때는 복사 생략)
```


### 파일 복사 아키텍처

- **완전 독립성**: 각 프로젝트가 완전히 독립된 파일 복사본 사용
- **크로스 플랫폼**: Windows/macOS/Linux 모든 환경에서 동일한 동작
- **안정성**: shutil.copytree 기반의 안전한 파일 복사

### 설치 및 활용

```bash
# 표준 설치 (모든 플랫폼)
moai init project

# 상태 확인
moai status -v  # 상세 상태 확인
```

### (선택) 외부 브레인스토밍 CLI 확인

```bash
which codex  || echo "🔧 Codex CLI 미설치 – npm install -g @openai/codex 또는 brew install codex"
which gemini || echo "🔧 Gemini CLI 미설치 – npm install -g @google/gemini-cli 또는 brew install gemini-cli"
```

`/moai:0-project` 인터뷰에서 project-manager 에이전트가 위 명령과 동일한 검사를 수행하고, 설치/로그인 방법만 안내합니다. 자동 설치는 수행하지 않습니다.

```json
// 브레인스토밍 설정 예시
{
  "brainstorming": {
    "enabled": true,
    "providers": ["claude", "codex", "gemini"]
  }
}
```

`providers` 배열에는 항상 `"claude"` 를 유지하고, 추가로 사용할 엔진을 선택합니다.

## 🛠️ 설치 후 확인

### 기본 상태 확인

```bash
# 간단한 상태 확인
moai status

# 상세 상태 확인
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

## 🔄 업데이트 시스템 (예: vX.Y.Z)

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

내부 도구로 제공되는 VersionSyncManager를 통해 문서/설정의 버전 문자열을 점검/동기화할 수 있습니다.

```bash
# 드라이런(실제 변경 없음)
python -m moai_adk.core.version_sync --dry-run

# 동기화 검증만
python -m moai_adk.core.version_sync --verify

# 버전 업데이트 스크립트 생성(선택)
python -m moai_adk.core.version_sync --create-script
```

일반 사용자는 `moai update`만으로 충분합니다. VersionSyncManager는 패키지 개발/문서 유지보수 목적의 고급 도구입니다.

## 🚨 문제 해결

### 일반적인 설치 문제

1. **권한 오류 (Windows)**
   ```bash
   # 해결방법: 파일 복사 모드로 설치
   moai init . --force-copy
   ```

2. **Python 버전 호환성**
   ```bash
   # Python 3.11+ 확인
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

### 리소스/템플릿 업데이트

```bash
# 템플릿 버전 확인
moai update --check

# 리소스만 갱신(자동 백업 생성)
moai update --resources-only

# 패키지 업그레이드 포함 전체 업데이트 안내
moai update
```

업데이트 실행 시 `.moai/version.json`이 최신 템플릿 버전으로 갱신되며, 기존 리소스는 `.moai_backup_*` 디렉터리에 자동 백업됩니다.
```

## ✅ 설치 완료 체크리스트

- [ ] `moai --version` 명령어 정상 동작
- [ ] `moai status` 모든 컴포넌트 정상
- [ ] `.claude/` 디렉토리 구조 완성
- [ ] `.moai/` 디렉토리 구조 완성
- [ ] Hook 시스템 정상 동작
- [ ] TAG 시스템 초기화 완료
- [ ] CI/CD 파이프라인 설정 완료

설치 완료 후 `/moai:1-project [프로젝트이름]` 명령어로 프로젝트별 설정을 진행하세요.

## 📚 관련 문서

- **[시스템 개요](01-overview.md)**: MoAI-ADK 전체 소개 및 주요 기능
- **[패키지 구조](package-structure.md)**: 설치된 패키지의 내부 구조 이해
- **[아키텍처](04-architecture.md)**: 생성되는 프로젝트 구조 상세 설명
- **[명령어 시스템](08-commands.md)**: 설치 후 사용 가능한 CLI 명령어
- **[대화형 마법사](06-wizard.md)**: 프로젝트 초기화 및 설정
- **[빌드 시스템](build-system.md)**: 개발 환경 설정 및 빌드
