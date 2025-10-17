# @DOC:START-INSTALL-001 | Chain: @SPEC:DOCS-003 -> @DOC:START-001

# Installation

MoAI-ADK를 설치하고 첫 프로젝트를 시작하는 방법을 안내합니다.

## Prerequisites

MoAI-ADK를 사용하기 전에 다음이 필요합니다:

- **Python 3.13+**: 최신 Python 버전
- **Claude Code**: Anthropic의 Claude Code CLI
- **Git**: 버전 관리 시스템

### Python 3.13 설치 확인

```bash
python --version
# 출력: Python 3.13.x
```

### Claude Code 설치

```bash
# Claude Code CLI 설치 (공식 문서 참조)
# https://claude.ai/claude-code
```

---

## PyPI에서 설치

MoAI-ADK는 PyPI에 패키지로 배포됩니다:

```bash
pip install moai-adk
```

### 설치 확인

```bash
moai-adk --version
# 출력: moai-adk v0.3.3
```

---

## 개발 모드 설치 (기여자용)

MoAI-ADK 개발에 참여하려면:

```bash
# 1. 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 2. 가상환경 생성
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate

# 3. editable 모드 설치
pip install -e ".[dev]"

# 4. 설치 확인
moai-adk --version
```

---

## 템플릿 프로젝트 생성

MoAI-ADK 프로젝트를 초기화합니다:

```bash
# 새 프로젝트 생성
moai-adk init my-project

# 프로젝트 디렉토리로 이동
cd my-project
```

### 생성된 프로젝트 구조

```
my-project/
├── .moai/                    # MoAI-ADK 설정 및 메모리
│   ├── config.json          # 프로젝트 설정
│   ├── memory/              # 개발 가이드
│   ├── specs/               # SPEC 문서
│   └── indexes/             # TAG 인덱스
├── .claude/                 # Claude Code 에이전트 정의
├── src/                     # 소스 코드
├── tests/                   # 테스트 코드
├── docs/                    # MkDocs 문서
└── pyproject.toml           # Python 프로젝트 설정
```

---

## 다음 단계

설치가 완료되었습니다. 이제 첫 프로젝트를 시작해보세요:

- [빠른 시작](quick-start.md) - 첫 SPEC 작성 및 구현
- [프로젝트 초기화](first-project.md) - TODO 앱 예제

---

**다음**: [Quick Start →](quick-start.md)
