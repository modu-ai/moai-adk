# Templates API Reference

> MoAI-ADK Python v0.3.0 Templates Module API Documentation

MoAI-ADK의 템플릿 시스템으로, 프로젝트 초기화 시 사용되는 문서 템플릿을 관리합니다.

---

## 모듈 구조

```
templates/
├── CLAUDE.md                    # Claude 설정 템플릿 (프로젝트 루트)
└── .moai/
    ├── config.json             # 프로젝트 설정 템플릿
    ├── project/
    │   ├── product.md          # 제품 개요 템플릿
    │   ├── structure.md        # 디렉토리 구조 템플릿
    │   └── tech.md             # 기술 스택 템플릿
    └── memory/
        ├── development-guide.md # 개발 가이드 템플릿
        └── spec-metadata.md     # SPEC 메타데이터 가이드 템플릿
```

---

## 템플릿 시스템 개요

MoAI-ADK는 Jinja2 스타일 변수 치환을 사용하는 템플릿 시스템을 제공합니다.

### 템플릿 변수 규칙

| 변수명 | 설명 | 기본값 | 예시 |
|--------|------|--------|------|
| `{{PROJECT_NAME}}` | 프로젝트 이름 | 디렉토리명 | `my-awesome-app` |
| `{{PROJECT_DESCRIPTION}}` | 프로젝트 설명 | (빈 문자열) | `A MoAI-ADK project` |
| `{{PROJECT_VERSION}}` | 프로젝트 버전 | `0.1.0` | `1.0.0` |
| `{{PROJECT_MODE}}` | 프로젝트 모드 | `personal` | `team` |
| `{{AUTHOR}}` | 작성자 | (Git user.name) | `Goos` |
| `{{LANGUAGE}}` | 주 프로그래밍 언어 | (자동 감지) | `python` |

### 템플릿 렌더링 흐름

```
1. moai init . 실행
   ↓
2. ProjectInitializer.initialize() 호출
   ↓
3. 템플릿 변수 수집
   - PROJECT_NAME: 디렉토리명 추출
   - LANGUAGE: LanguageDetector로 감지
   - AUTHOR: Git config user.name 조회
   ↓
4. templates/ 디렉토리 파일 복사
   ↓
5. 변수 치환 ({{VAR}} → 실제 값)
   ↓
6. .moai/ 디렉토리에 최종 파일 생성
```

---

## 템플릿 파일 상세

### CLAUDE.md

> **경로**: `templates/CLAUDE.md`
>
> **목적지**: `{PROJECT_ROOT}/CLAUDE.md`

프로젝트의 Claude 설정 파일 템플릿입니다. Alfred SuperAgent와 9개 전문 에이전트의 워크플로우를 정의합니다.

#### 포함 내용

- **Alfred SuperAgent 소개**: 페르소나, 오케스트레이션 전략
- **9개 전문 에이전트 생태계**: 각 에이전트의 역할과 전문 영역
- **3단계 개발 워크플로우**: `/alfred:1-spec`, `/alfred:2-build`, `/alfred:3-sync`
- **@TAG Lifecycle**: SPEC → TEST → CODE → DOC 체계
- **TRUST 5원칙**: 테스트, 가독성, 통합, 보안, 추적성
- **언어별 코드 규칙**: 공통 제약 및 품질 기준
- **TDD 워크플로우 체크리스트**: 단계별 체크리스트

#### 템플릿 변수

```markdown
# {{PROJECT_NAME}} - MoAI-Agentic Development Kit

## 프로젝트 정보

- **이름**: {{PROJECT_NAME}}
- **설명**: {{PROJECT_DESCRIPTION}}
- **버전**: {{PROJECT_VERSION}}
- **모드**: {{PROJECT_MODE}}
```

#### 사용 예시

```python
from pathlib import Path
from moai_adk.core.project import ProjectInitializer

# 프로젝트 초기화
initializer = ProjectInitializer("/path/to/project")
result = initializer.initialize()

# CLAUDE.md 생성됨
# /path/to/project/CLAUDE.md
```

---

### config.json

> **경로**: `templates/.moai/config.json`
>
> **목적지**: `{PROJECT_ROOT}/.moai/config.json`

프로젝트 설정 파일 템플릿입니다.

#### 템플릿 구조

```json
{
  "projectName": "{{PROJECT_NAME}}",
  "mode": "{{PROJECT_MODE}}",
  "locale": "{{LOCALE}}",
  "language": "{{LANGUAGE}}"
}
```

#### 필드 설명

| 필드 | 타입 | 설명 | 예시 |
|------|------|------|------|
| `projectName` | string | 프로젝트 이름 | `"my-awesome-app"` |
| `mode` | string | 프로젝트 모드 (`personal`, `team`) | `"team"` |
| `locale` | string | 로케일 (`ko`, `en`, `ja`, `zh`) | `"ko"` |
| `language` | string | 주 프로그래밍 언어 | `"python"` |

#### 생성 예시

```bash
$ moai init . --mode team --locale ko
```

생성되는 `.moai/config.json`:

```json
{
  "projectName": "my-project",
  "mode": "team",
  "locale": "ko",
  "language": "python"
}
```

---

### product.md

> **경로**: `templates/.moai/project/product.md`
>
> **목적지**: `{PROJECT_ROOT}/.moai/project/product.md`

제품 개요 문서 템플릿입니다. 프로젝트의 목표, 핵심 기능, 사용자 페르소나를 정의합니다.

#### 템플릿 구조

```markdown
# {{PROJECT_NAME}} - 제품 개요

## 프로젝트 정보

- **이름**: {{PROJECT_NAME}}
- **설명**: {{PROJECT_DESCRIPTION}}
- **버전**: {{PROJECT_VERSION}}
- **작성자**: @{{AUTHOR}}
- **작성일**: {{CREATED_DATE}}

---

## 제품 비전

> 이 프로젝트가 해결하려는 핵심 문제와 비전을 작성하세요.

(작성 필요)

---

## 핵심 기능

### 1. (기능 1)

- **목적**: (작성 필요)
- **대상 사용자**: (작성 필요)
- **핵심 가치**: (작성 필요)

### 2. (기능 2)

- **목적**: (작성 필요)
- **대상 사용자**: (작성 필요)
- **핵심 가치**: (작성 필요)

---

## 사용자 페르소나

### 페르소나 1: (이름)

- **배경**: (작성 필요)
- **목표**: (작성 필요)
- **Pain Points**: (작성 필요)

---

## 성공 지표

- (작성 필요)

---

**최종 업데이트**: {{UPDATED_DATE}}
**작성자**: @{{AUTHOR}}
```

#### 작성 가이드

`product.md`는 프로젝트 초기화 후 `/alfred:0-project` 명령어로 자동 작성됩니다:

```bash
# 1. 프로젝트 초기화
$ moai init .

# 2. product.md 자동 작성
Claude Code에서:
/alfred:0-project
```

---

### structure.md

> **경로**: `templates/.moai/project/structure.md`
>
> **목적지**: `{PROJECT_ROOT}/.moai/project/structure.md`

디렉토리 구조 및 모듈 설계 문서 템플릿입니다.

#### 템플릿 구조

```markdown
# {{PROJECT_NAME}} - 디렉토리 구조

## 프로젝트 구조

```

{{PROJECT_NAME}}/
├── src/                  # 소스 코드
│   └── (작성 필요)
├── tests/                # 테스트 코드
│   └── (작성 필요)
├── docs/                 # 문서
│   └── (작성 필요)
├── .moai/                # MoAI-ADK 설정
│   ├── config.json
│   ├── project/
│   ├── specs/
│   ├── memory/
│   └── backup/
└── README.md

```

---

## 모듈 설계

### 1. (모듈 1)

- **경로**: `src/(작성 필요)`
- **역할**: (작성 필요)
- **의존성**: (작성 필요)

### 2. (모듈 2)

- **경로**: `src/(작성 필요)`
- **역할**: (작성 필요)
- **의존성**: (작성 필요)

---

## 디렉토리 명명 규칙

- (작성 필요)

---

**최종 업데이트**: {{UPDATED_DATE}}
**작성자**: @{{AUTHOR}}
```

#### 작성 가이드

`structure.md`도 `/alfred:0-project` 명령어로 자동 작성됩니다. Alfred는 현재 프로젝트의 파일 구조를 스캔하여 자동으로 문서를 생성합니다.

---

### tech.md

> **경로**: `templates/.moai/project/tech.md`
>
> **목적지**: `{PROJECT_ROOT}/.moai/project/tech.md`

기술 스택 및 도구 체인 문서 템플릿입니다.

#### 템플릿 구조

```markdown
# {{PROJECT_NAME}} - 기술 스택

## 주 프로그래밍 언어

- **언어**: {{LANGUAGE}}
- **버전**: (작성 필요)

---

## 프레임워크 및 라이브러리

### 핵심 프레임워크

- (작성 필요)

### 주요 라이브러리

- (작성 필요)

---

## 개발 도구

### 테스트

- **프레임워크**: (작성 필요)
- **커버리지 목표**: 85% 이상
- **테스트 전략**: (작성 필요)

### 린터/포맷터

- **린터**: (작성 필요)
- **포맷터**: (작성 필요)

### 타입 검사

- **도구**: (작성 필요)

---

## 빌드 및 배포

- **빌드 도구**: (작성 필요)
- **패키지 매니저**: (작성 필요)
- **CI/CD**: (작성 필요)

---

## TRUST 5원칙 구현

### T - Test First

- **도구**: (작성 필요)
- **전략**: (작성 필요)

### R - Readable

- **린터**: (작성 필요)
- **포맷터**: (작성 필요)

### U - Unified

- **타입 시스템**: (작성 필요)

### S - Secured

- **보안 도구**: (작성 필요)

### T - Trackable

- **@TAG 시스템**: CODE-FIRST 스캔 방식

---

**최종 업데이트**: {{UPDATED_DATE}}
**작성자**: @{{AUTHOR}}
```

#### 언어별 기술 스택 예시

**Python 프로젝트**:

```markdown
## 주 프로그래밍 언어

- **언어**: Python
- **버전**: 3.11+

## 프레임워크 및 라이브러리

### 핵심 프레임워크

- FastAPI (웹 프레임워크)
- SQLAlchemy (ORM)

### 주요 라이브러리

- pydantic (데이터 검증)
- httpx (HTTP 클라이언트)

## 개발 도구

### 테스트

- **프레임워크**: pytest
- **커버리지**: pytest-cov
- **테스트 전략**: TDD (Red-Green-Refactor)

### 린터/포맷터

- **린터**: ruff
- **포맷터**: black

### 타입 검사

- **도구**: mypy
```

**TypeScript 프로젝트**:

```markdown
## 주 프로그래밍 언어

- **언어**: TypeScript
- **버전**: 5.3+

## 프레임워크 및 라이브러리

### 핵심 프레임워크

- Node.js (런타임)
- Express (웹 프레임워크)

### 주요 라이브러리

- zod (스키마 검증)
- prisma (ORM)

## 개발 도구

### 테스트

- **프레임워크**: Vitest
- **커버리지**: @vitest/coverage-v8
- **테스트 전략**: TDD (Red-Green-Refactor)

### 린터/포맷터

- **린터**: Biome
- **포맷터**: Biome

### 타입 검사

- **도구**: tsc (TypeScript Compiler)
```

---

### development-guide.md

> **경로**: `templates/.moai/memory/development-guide.md`
>
> **목적지**: `{PROJECT_ROOT}/.moai/memory/development-guide.md`

MoAI-ADK 개발 가이드 템플릿입니다. SPEC-First TDD 워크플로우, TRUST 5원칙, @TAG 시스템을 설명합니다.

#### 포함 내용

- **SPEC 우선 TDD 워크플로우**: 3단계 개발 루프
- **EARS 요구사항 작성법**: 5가지 구문 패턴
- **Context Engineering**: JIT Retrieval, Compaction
- **TRUST 5원칙**: 테스트, 가독성, 통합, 보안, 추적성
- **@TAG 시스템**: TAG BLOCK 템플릿, 사용 규칙
- **개발 원칙**: 코드 제약, 품질 기준, 리팩토링 규칙
- **언어별 도구 매핑**: Python, TypeScript, Java, Go, Rust 등

#### 템플릿 변수

```markdown
# {{PROJECT_NAME}} 개발 가이드

> "명세 없으면 코드 없다. 테스트 없으면 구현 없다."

MoAI-ADK 범용 개발 툴킷을 사용하는 모든 에이전트와 개발자를 위한 통합 가드레일이다.
```

---

### spec-metadata.md

> **경로**: `templates/.moai/memory/spec-metadata.md`
>
> **목적지**: `{PROJECT_ROOT}/.moai/memory/spec-metadata.md`

SPEC 메타데이터 구조 가이드 템플릿입니다. 모든 SPEC 문서가 따라야 하는 메타데이터 표준을 정의합니다.

#### 포함 내용

- **메타데이터 구조 개요**: 필수 7개 + 선택 9개 필드
- **필수 필드**: id, version, status, created, updated, author, priority
- **선택 필드**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **메타데이터 검증**: 필수 필드/형식 검증 명령어
- **마이그레이션 가이드**: 기존 SPEC 업데이트 방법
- **설계 원칙**: DRY, Context-Aware, Traceable, Maintainable, Simple First

#### SPEC 메타데이터 예시

```yaml
---
# 필수 필드 (7개)
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @{{AUTHOR}}
priority: high

# 선택 필드
category: security
labels:
  - authentication
  - jwt

depends_on:
  - USER-001

scope:
  packages:
    - src/core/auth
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-09-15)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
```

---

## 템플릿 커스터마이징

### 프로젝트별 템플릿 수정

프로젝트 초기화 후 생성된 템플릿 파일을 자유롭게 수정할 수 있습니다:

```bash
# 1. 프로젝트 초기화
$ moai init .

# 2. 생성된 템플릿 파일 수정
$ vim .moai/project/product.md
$ vim .moai/project/structure.md
$ vim .moai/project/tech.md

# 3. Alfred로 자동 작성
Claude Code에서:
/alfred:0-project
```

### 커스텀 템플릿 추가

프로젝트별 커스텀 템플릿을 추가할 수 있습니다:

```bash
.moai/
└── project/
    ├── product.md           # 기본 템플릿
    ├── structure.md         # 기본 템플릿
    ├── tech.md              # 기본 템플릿
    ├── deployment.md        # 커스텀 템플릿 (배포 가이드)
    └── troubleshooting.md   # 커스텀 템플릿 (트러블슈팅)
```

---

## 템플릿 변수 수집 로직

### ProjectInitializer.initialize() 내부 동작

```python
from moai_adk.core.project import ProjectInitializer, LanguageDetector
from pathlib import Path
import subprocess

def collect_template_variables(project_path: Path) -> dict[str, str]:
    """템플릿 변수 수집"""
    variables = {}

    # 1. PROJECT_NAME: 디렉토리명
    variables["PROJECT_NAME"] = project_path.name

    # 2. LANGUAGE: 자동 감지
    detector = LanguageDetector(project_path)
    variables["LANGUAGE"] = detector.detect() or "generic"

    # 3. AUTHOR: Git config에서 조회
    try:
        result = subprocess.run(
            ["git", "config", "user.name"],
            capture_output=True,
            text=True,
            cwd=project_path
        )
        variables["AUTHOR"] = result.stdout.strip() or "Unknown"
    except Exception:
        variables["AUTHOR"] = "Unknown"

    # 4. 기본값
    variables["PROJECT_DESCRIPTION"] = f"A MoAI-ADK project built with {variables['LANGUAGE']}"
    variables["PROJECT_VERSION"] = "0.1.0"
    variables["PROJECT_MODE"] = "personal"  # 사용자 옵션으로 변경 가능

    return variables

# 사용 예시
project_path = Path("/path/to/project")
variables = collect_template_variables(project_path)

print(variables)
# 출력:
# {
#     'PROJECT_NAME': 'my-project',
#     'LANGUAGE': 'python',
#     'AUTHOR': 'Goos',
#     'PROJECT_DESCRIPTION': 'A MoAI-ADK project built with python',
#     'PROJECT_VERSION': '0.1.0',
#     'PROJECT_MODE': 'personal'
# }
```

---

## 템플릿 렌더링 예시

### 간단한 변수 치환

```python
import re
from pathlib import Path

def render_template(template_content: str, variables: dict[str, str]) -> str:
    """Jinja2 스타일 변수 치환"""
    result = template_content

    for key, value in variables.items():
        pattern = r'\{\{' + key + r'\}\}'
        result = re.sub(pattern, value, result)

    return result

# 템플릿 읽기
template_path = Path("templates/CLAUDE.md")
template_content = template_path.read_text()

# 변수 치환
variables = {
    "PROJECT_NAME": "my-awesome-app",
    "PROJECT_DESCRIPTION": "An awesome web application",
    "PROJECT_VERSION": "1.0.0",
    "PROJECT_MODE": "team",
    "AUTHOR": "Goos",
    "LANGUAGE": "python"
}

rendered = render_template(template_content, variables)

# 결과 저장
output_path = Path("/path/to/project/CLAUDE.md")
output_path.write_text(rendered)

print(f"템플릿 렌더링 완료: {output_path}")
```

---

## 템플릿 검증

### 템플릿 무결성 확인

```python
import re
from pathlib import Path

def validate_template(template_path: Path) -> list[str]:
    """템플릿 파일의 미치환 변수 확인"""
    content = template_path.read_text()

    # {{VAR}} 패턴 추출
    unresolved = re.findall(r'\{\{([A-Z_]+)\}\}', content)

    if unresolved:
        print(f"⚠️ {template_path.name}에 미치환 변수가 있습니다:")
        for var in set(unresolved):
            print(f"  - {{{{{{var}}}}}}")
        return list(set(unresolved))
    else:
        print(f"✅ {template_path.name} 검증 완료 (모든 변수 치환됨)")
        return []

# 사용 예시
template_path = Path(".moai/project/product.md")
unresolved_vars = validate_template(template_path)

if unresolved_vars:
    print(f"\n미치환 변수: {unresolved_vars}")
    print("템플릿 렌더링을 다시 실행하세요.")
```

---

## 통합 사용 예시

### 완전한 프로젝트 초기화 워크플로우

```bash
# 1. 새 프로젝트 디렉토리 생성
$ mkdir my-awesome-app
$ cd my-awesome-app

# 2. Git 저장소 초기화
$ git init
$ git config user.name "Goos"
$ git config user.email "goos@example.com"

# 3. Python 프로젝트 파일 생성
$ touch main.py
$ echo "print('Hello, World!')" > main.py

# 4. MoAI-ADK 프로젝트 초기화
$ moai init . --mode team --locale ko
Initializing MoAI-ADK project...
✓ Language detected: python
✓ Mode: team
✓ Locale: ko
✓ Created .moai/ directory
✓ Generated config.json
✓ Copied template files
✓ Rendered variables
Project initialized successfully!

# 5. 생성된 파일 확인
$ tree -a -L 3
.
├── .git/
├── .moai/
│   ├── config.json
│   ├── project/
│   │   ├── product.md
│   │   ├── structure.md
│   │   └── tech.md
│   ├── memory/
│   │   ├── development-guide.md
│   │   └── spec-metadata.md
│   ├── specs/
│   └── backup/
├── CLAUDE.md
└── main.py

# 6. CLAUDE.md 확인
$ head -20 CLAUDE.md
# my-awesome-app - MoAI-Agentic Development Kit

**SPEC-First TDD Development with Alfred SuperAgent**

---

## ▶◀ Meet Alfred: Your MoAI SuperAgent

**Alfred**는 모두의AI(MoAI)가 설계한 MoAI-ADK의 공식 SuperAgent입니다.

### Alfred 페르소나

- **정체성**: 모두의 AI 집사 ▶◀ - 정확하고 예의 바르며, 모든 요청을 체계적으로 처리
- **역할**: MoAI-ADK 워크플로우의 중앙 오케스트레이터
- **책임**: 사용자 요청 분석 → 적절한 전문 에이전트 위임 → 결과 통합 보고
- **목표**: SPEC-First TDD 방법론을 통한 완벽한 코드 품질 보장

# 7. config.json 확인
$ cat .moai/config.json
{
  "projectName": "my-awesome-app",
  "mode": "team",
  "locale": "ko",
  "language": "python"
}

# 8. Claude Code에서 문서 자동 작성
# /alfred:0-project
```

---

## Python API 사용 예시

### 템플릿 시스템 프로그래밍 방식 사용

```python
from pathlib import Path
from moai_adk.core.project import ProjectInitializer

def setup_new_project(
    project_name: str,
    language: str = None,
    mode: str = "team",
    locale: str = "ko"
) -> dict[str, str]:
    """새 프로젝트 설정 헬퍼 함수"""

    # 1. 프로젝트 디렉토리 생성
    project_path = Path.cwd() / project_name
    project_path.mkdir(exist_ok=True)

    # 2. 초기화
    initializer = ProjectInitializer(project_path)
    result = initializer.initialize(
        mode=mode,
        locale=locale,
        language=language
    )

    print(f"✅ 프로젝트 '{result['path']}' 초기화 완료")
    print(f"   - 언어: {result['language']}")
    print(f"   - 모드: {result['mode']}")
    print(f"   - 로케일: {result['locale']}")

    return result

# 사용 예시
if __name__ == "__main__":
    # Python 프로젝트 생성
    setup_new_project(
        project_name="my-python-app",
        language="python",
        mode="team",
        locale="ko"
    )

    # TypeScript 프로젝트 생성
    setup_new_project(
        project_name="my-ts-app",
        language="typescript",
        mode="personal",
        locale="en"
    )

    # 언어 자동 감지 프로젝트
    setup_new_project(
        project_name="my-auto-detected-app",
        language=None,  # 자동 감지
        mode="team",
        locale="ja"
    )
```

---

## 참고 문서

- **Core API**: `docs/api/core.md` - ProjectInitializer, LanguageDetector 상세 문서
- **CLI API**: `docs/api/cli.md` - moai init 명령어 사용법
- **개발 가이드**: `.moai/memory/development-guide.md` - SPEC-First TDD 워크플로우
- **SPEC 메타데이터**: `.moai/memory/spec-metadata.md` - SPEC 문서 표준

---

**최종 업데이트**: 2025-10-14
**버전**: v0.3.0
**작성자**: @doc-syncer
