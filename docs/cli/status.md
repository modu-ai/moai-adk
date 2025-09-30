# moai status

`moai status` 명령어는 현재 프로젝트의 MoAI-ADK 설정 상태를 빠르게 확인하는 진단 도구입니다. 프로젝트가 MoAI-ADK로 초기화되었는지, Claude Code 통합이 활성화되었는지, 버전 정보는 어떠한지, 업데이트가 필요한지 등을 한눈에 파악할 수 있습니다. 프로젝트 타입(Full MoAI Project, Partial MoAI Project, Claude Project, Regular Directory)을 자동으로 감지하여 현재 상태를 명확히 표시합니다.

이 명령어는 프로젝트 온보딩, 트러블슈팅, 팀 협업 시 개발 환경 확인에 매우 유용합니다. 새로운 팀원이 프로젝트를 클론한 후 `moai status`를 실행하면 MoAI-ADK가 제대로 설정되었는지 즉시 확인할 수 있으며, 문제가 있다면 구체적인 권장 사항을 제공합니다. 또한 `--verbose` 플래그를 사용하면 `.moai`, `.claude` 디렉토리의 파일 개수, `CLAUDE.md` 존재 여부 등 세부 정보까지 확인할 수 있습니다.

StatusCommand 클래스는 프로젝트 경로를 스캔하여 `.moai` 디렉토리, `.claude` 디렉토리, `CLAUDE.md` 메모리 파일, `.git` 저장소의 존재 여부를 확인합니다. 각 컴포넌트가 있으면 ✅, 없으면 ❌로 표시하여 시각적으로 명확한 피드백을 제공합니다. 버전 정보는 `package.json`에서 패키지 버전을, `.moai/version.json`에서 템플릿 버전을 읽어와 비교하여 업데이트 필요 여부를 판단합니다.

이 명령어는 읽기 전용 작업만 수행하므로 프로젝트에 어떠한 변경도 가하지 않습니다. 언제든지 안전하게 실행할 수 있으며, CI/CD 파이프라인에서 프로젝트 상태 검증 단계로도 활용할 수 있습니다. 특히 멀티 프로젝트 모노레포 환경에서 각 프로젝트의 MoAI-ADK 설정 상태를 빠르게 확인하는 데 유용합니다.

## 개요

`moai status`는 다음 정보를 제공합니다:

1. **프로젝트 경로 및 타입**
   - 절대 경로 표시
   - 프로젝트 타입 자동 분류 (Full/Partial/Claude/Regular)

2. **MoAI-ADK 컴포넌트 상태**
   - MoAI 시스템 (`.moai` 디렉토리)
   - Claude 통합 (`.claude` 디렉토리)
   - 메모리 파일 (`CLAUDE.md`)
   - Git 저장소 (`.git` 디렉토리)

3. **버전 정보**
   - 설치된 패키지 버전
   - 프로젝트 템플릿 버전
   - 사용 가능한 업데이트 (있는 경우)

4. **상세 정보 (--verbose)**
   - 각 디렉토리의 파일 개수
   - 전체 MoAI 파일 통계

5. **권장 사항**
   - MoAI 초기화 필요 여부
   - Git 저장소 초기화 권장
   - 업데이트 권장

프로젝트 타입 분류는 다음과 같이 이루어집니다:
- **MoAI Project (Full)**: `.moai`와 `.claude` 모두 존재
- **MoAI Project (Partial)**: `.moai`만 존재
- **Claude Project**: `.claude`만 존재
- **Regular Directory**: 둘 다 없음

이 분류를 통해 프로젝트의 설정 완성도를 즉시 파악할 수 있습니다. Full MoAI Project는 3단계 워크플로우(`/moai:1-spec`, `/moai:2-build`, `/moai:3-sync`)를 완전히 활용할 수 있는 상태를 의미합니다.

## 기본 사용법

```bash
moai status [options]
```

### 옵션

| 옵션 | 설명 |
|------|------|
| `--verbose`, `-v` | 파일 개수 등 상세 정보 표시 |
| `--project-path <path>` | 상태를 확인할 프로젝트 경로 지정 (기본: 현재 디렉토리) |
| `-h, --help` | 도움말 메시지 표시 |

### 주요 기능

- **빠른 상태 확인**: 프로젝트 설정 완성도를 5초 내 파악
- **프로젝트 타입 분류**: 4가지 프로젝트 타입 자동 감지
- **버전 추적**: 패키지 및 템플릿 버전 비교
- **업데이트 알림**: 템플릿이 오래되었을 때 경고
- **컴포넌트 체크**: 필수 디렉토리 및 파일 존재 여부 검증
- **Git 통합 확인**: 버전 관리 설정 상태 확인
- **파일 통계** (verbose): 프로젝트 규모 파악
- **실행 가능한 권장 사항**: 문제 해결을 위한 구체적인 명령어 제시

## 사용 예시

### 1. 기본 상태 확인

가장 일반적인 사용법은 프로젝트 디렉토리에서 옵션 없이 `moai status`를 실행하는 것입니다.

```bash
cd my-project
moai status
```

**출력 예시 (Full MoAI Project)**:

```
📊 MoAI-ADK Project Status

📂 Project: /Users/user/projects/my-project
   Type: MoAI Project (Full)

🗿 MoAI-ADK Components:
   MoAI System: ✅
   Claude Integration: ✅
   Memory File: ✅
   Git Repository: ✅

🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.1
```

이 출력은 프로젝트가 완전히 설정되어 있음을 나타냅니다. 모든 컴포넌트가 ✅로 표시되고, 버전이 일치하며, 권장 사항이 없습니다. 이 상태에서는 즉시 MoAI-ADK의 모든 기능을 사용할 수 있습니다.

각 컴포넌트의 의미:
- **MoAI System**: `.moai/` 디렉토리에 프로젝트 설정, 메모리, 스크립트가 저장됩니다.
- **Claude Integration**: `.claude/` 디렉토리에 에이전트, 명령어, 훅, 출력 스타일이 있습니다.
- **Memory File**: `CLAUDE.md`는 프로젝트 컨텍스트를 Claude에게 전달하는 핵심 파일입니다.
- **Git Repository**: `.git/` 디렉토리가 있어 버전 관리가 활성화되어 있습니다.

### 2. 부분적으로 설정된 프로젝트

일부 컴포넌트만 있는 경우, 명확한 권장 사항이 제공됩니다.

```bash
moai status
```

**출력 예시 (Partial MoAI Project)**:

```
📊 MoAI-ADK Project Status

📂 Project: /Users/user/projects/partial-project
   Type: MoAI Project (Partial)

🗿 MoAI-ADK Components:
   MoAI System: ✅
   Claude Integration: ❌
   Memory File: ✅
   Git Repository: ❌

🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.1

💡 Recommendations:
   - Run 'moai init' to initialize MoAI-ADK
   - Initialize Git repository: git init
```

이 출력은 프로젝트가 부분적으로만 설정되었음을 나타냅니다. `.moai` 디렉토리와 `CLAUDE.md`는 있지만 `.claude` 디렉토리와 `.git` 저장소가 누락되었습니다. 권장 사항 섹션에서 `moai init`을 다시 실행하여 누락된 Claude 통합을 추가하고, `git init`으로 버전 관리를 활성화하라고 제안합니다.

Partial 상태에서는 MoAI-ADK의 일부 기능만 사용할 수 있습니다. Claude Integration이 없으면 7개의 전문 에이전트(`spec-builder`, `code-builder`, `doc-syncer`, `cc-manager`, `debug-helper`, `git-manager`, `trust-checker`)를 사용할 수 없으므로 `moai init`을 실행하여 Full 상태로 전환하는 것이 권장됩니다.

### 3. MoAI 미초기화 프로젝트

MoAI-ADK가 전혀 설정되지 않은 일반 디렉토리의 경우입니다.

```bash
moai status
```

**출력 예시 (Regular Directory)**:

```
📊 MoAI-ADK Project Status

📂 Project: /Users/user/projects/regular-project
   Type: Regular Directory

🗿 MoAI-ADK Components:
   MoAI System: ❌
   Claude Integration: ❌
   Memory File: ❌
   Git Repository: ✅

🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.1

💡 Recommendations:
   - Run 'moai init' to initialize MoAI-ADK
```

이 출력은 프로젝트가 MoAI-ADK로 초기화되지 않았음을 나타냅니다. Git 저장소만 있고 MoAI 관련 컴포넌트는 모두 누락되었습니다. `moai init`을 실행하여 프로젝트를 MoAI-ADK 프로젝트로 전환할 수 있습니다.

Regular Directory 상태에서는 MoAI-ADK의 어떤 기능도 사용할 수 없습니다. SPEC-First TDD 워크플로우, Claude Code 통합, @TAG 추적성 시스템 모두 `.moai`와 `.claude` 디렉토리가 필요하기 때문입니다. `moai init`은 필요한 모든 파일과 디렉토리를 자동으로 생성하여 프로젝트를 Full MoAI Project로 만들어줍니다.

### 4. 상세 정보 포함 (--verbose)

`--verbose` 플래그를 사용하면 파일 개수 통계가 추가로 표시됩니다.

```bash
moai status --verbose
```

**출력 예시**:

```
📊 MoAI-ADK Project Status

📂 Project: /Users/user/projects/my-project
   Type: MoAI Project (Full)

🗿 MoAI-ADK Components:
   MoAI System: ✅
   Claude Integration: ✅
   Memory File: ✅
   Git Repository: ✅

🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.1

📁 File Counts:
   .moai: 47 files
   .claude: 23 files
   CLAUDE.md: 1 files
```

파일 개수 정보는 프로젝트 규모를 파악하는 데 유용합니다. `.moai` 디렉토리에는 메모리 파일, 프로젝트 설정, 스크립트 등이 저장되며, `.claude` 디렉토리에는 에이전트 정의, 명령어, 훅, 출력 스타일이 포함됩니다. 파일 개수가 예상보다 많거나 적다면 프로젝트 설정에 문제가 있을 수 있습니다.

`countProjectFiles()` 메서드는 각 디렉토리를 재귀적으로 순회하여 전체 파일 개수를 계산합니다. `.moai/memory/`, `.moai/project/`, `.moai/scripts/` 같은 하위 디렉토리의 파일도 모두 포함됩니다. 이를 통해 프로젝트가 표준 템플릿 구조를 따르는지 확인할 수 있습니다.

### 5. 템플릿 업데이트 필요 시

템플릿 버전이 오래된 경우 경고 메시지가 표시됩니다.

```bash
moai status
```

**출력 예시**:

```
📊 MoAI-ADK Project Status

📂 Project: /Users/user/projects/my-project
   Type: MoAI Project (Full)

🗿 MoAI-ADK Components:
   MoAI System: ✅
   Claude Integration: ✅
   Memory File: ✅
   Git Repository: ✅

🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.0
   Available template update: v0.0.1
   ⚠️  Templates are outdated. Run 'moai update' to refresh.

💡 Recommendations:
   - Run 'moai update' to update project templates
```

템플릿 버전 불일치는 프로젝트가 오래된 템플릿으로 생성되었거나, MoAI-ADK 패키지는 업그레이드했지만 프로젝트 템플릿은 업데이트하지 않았을 때 발생합니다. 이 경우 `moai update` 명령어를 실행하여 `.moai`와 `.claude` 디렉토리의 파일을 최신 버전으로 업데이트할 수 있습니다.

템플릿 업데이트는 새로운 에이전트, 개선된 훅, 버그 수정 등을 포함할 수 있으므로 정기적으로 업데이트하는 것이 권장됩니다. `moai update --check` 명령어로 먼저 사용 가능한 업데이트를 확인한 후, `moai update`를 실행하여 실제로 업데이트할 수 있습니다. 업데이트 전 자동으로 백업이 생성되므로 문제 발생 시 복원할 수 있습니다.

### 6. 다른 프로젝트 경로 확인

현재 디렉토리가 아닌 다른 프로젝트의 상태를 확인할 때 `--project-path` 옵션을 사용합니다.

```bash
moai status --project-path /path/to/other/project
```

**출력 예시**:

```
📊 MoAI-ADK Project Status

📂 Project: /path/to/other/project
   Type: Claude Project

🗿 MoAI-ADK Components:
   MoAI System: ❌
   Claude Integration: ✅
   Memory File: ❌
   Git Repository: ✅

🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.1

💡 Recommendations:
   - Run 'moai init' to initialize MoAI-ADK
```

이 예시는 Claude Code만 설정되어 있고 MoAI-ADK는 초기화되지 않은 프로젝트를 보여줍니다. 이러한 "Claude Project" 타입은 기존에 Claude Code를 사용하던 프로젝트에 MoAI-ADK를 추가로 도입할 때 자주 나타납니다. `moai init`을 실행하면 기존 `.claude` 설정을 유지하면서 `.moai` 디렉토리와 `CLAUDE.md` 파일을 추가합니다.

`--project-path` 옵션은 멀티 프로젝트 모노레포 환경에서 특히 유용합니다. 각 프로젝트의 MoAI-ADK 설정 상태를 한 번에 확인하는 스크립트를 작성할 수 있습니다:

```bash
#!/bin/bash
for project in projects/*; do
  echo "Checking $project..."
  moai status --project-path "$project"
  echo ""
done
```

### 7. CI/CD 파이프라인에서 사용

CI/CD 환경에서 프로젝트 설정 검증 단계로 활용할 수 있습니다.

**GitHub Actions 예시**:

```yaml
name: Verify MoAI-ADK Setup
on: [push, pull_request]

jobs:
  verify-setup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install MoAI-ADK
        run: npm install -g moai-adk

      - name: Check Project Status
        run: moai status --verbose

      - name: Verify Full Setup
        run: |
          # Check if all components are initialized
          if [ ! -d ".moai" ] || [ ! -d ".claude" ]; then
            echo "MoAI-ADK not fully initialized!"
            exit 1
          fi
```

**GitLab CI 예시**:

```yaml
stages:
  - verify

verify-moai-setup:
  stage: verify
  image: node:20
  script:
    - npm install -g moai-adk
    - moai status --verbose
    - |
      if [ ! -f "CLAUDE.md" ]; then
        echo "CLAUDE.md missing!"
        exit 1
      fi
```

CI/CD에서 `moai status`를 사용하면 다음과 같은 이점이 있습니다:

1. **일관성 보장**: 모든 PR이 Full MoAI Project 상태를 유지하도록 강제
2. **빠른 실패**: 설정 문제를 빌드 초기에 감지하여 시간 절약
3. **자동 검증**: 수동 확인 없이 프로젝트 설정 완성도 검증
4. **문서화**: 빌드 로그에 프로젝트 상태가 기록되어 추적 가능

`moai status`는 읽기 전용 작업만 수행하므로 CI/CD에서 안전하게 실행할 수 있으며, 종료 코드를 통해 스크립트에서 성공/실패를 판단할 수 있습니다.

## 출력 섹션 상세 설명

`moai status`의 출력은 여러 섹션으로 구성되며, 각 섹션은 프로젝트의 특정 측면을 나타냅니다.

### 1. 프로젝트 정보 (📂 Project)

```
📂 Project: /Users/user/projects/my-project
   Type: MoAI Project (Full)
```

**프로젝트 경로**: `checkProjectStatus()` 메서드가 `path.resolve()`를 사용하여 절대 경로로 변환합니다. 이를 통해 현재 작업 디렉토리와 관계없이 정확한 위치를 표시합니다.

**프로젝트 타입**: 4가지 유형으로 분류됩니다.

| 타입 | 조건 | 설명 |
|------|------|------|
| **MoAI Project (Full)** | `.moai` ✅ `.claude` ✅ | 완전히 설정된 MoAI-ADK 프로젝트 |
| **MoAI Project (Partial)** | `.moai` ✅ `.claude` ❌ | MoAI 시스템만 있고 Claude 통합 누락 |
| **Claude Project** | `.moai` ❌ `.claude` ✅ | Claude Code만 설정된 프로젝트 |
| **Regular Directory** | `.moai` ❌ `.claude` ❌ | MoAI-ADK 미초기화 |

프로젝트 타입은 사용 가능한 기능을 즉시 파악하는 데 도움이 됩니다. Full 타입만 모든 MoAI-ADK 기능을 완전히 활용할 수 있습니다.

### 2. 컴포넌트 상태 (🗿 MoAI-ADK Components)

```
🗿 MoAI-ADK Components:
   MoAI System: ✅
   Claude Integration: ✅
   Memory File: ✅
   Git Repository: ✅
```

각 컴포넌트는 `fs.pathExists()`로 검증됩니다:

**MoAI System (`.moai/`)**: MoAI-ADK의 핵심 시스템 파일이 저장되는 디렉토리입니다. 다음 하위 디렉토리를 포함합니다:
- `.moai/memory/`: 개발 가이드, 프로젝트 정의
- `.moai/project/`: product.md, structure.md, tech.md
- `.moai/scripts/`: 자동화 스크립트
- TAG 시스템: 소스코드에만 존재 (CODE-FIRST, `rg '@TAG'` 명령으로 검색)
- `.moai/version.json`: 템플릿 버전 정보

**Claude Integration (`.claude/`)**: Claude Code와의 통합을 위한 디렉토리입니다. 다음 요소를 포함합니다:
- `.claude/agents/moai/`: 7개 전문 에이전트 (`spec-builder`, `code-builder`, `doc-syncer`, `cc-manager`, `debug-helper`, `git-manager`, `trust-checker`)
- `.claude/commands/moai/`: 5개 슬래시 명령어 (`/moai:0-project`, `/moai:1-spec`, `/moai:2-build`, `/moai:3-sync`)
- `.claude/hooks/moai/`: 8개 이벤트 훅 (보안, 모니터링)
- `.claude/output-styles/`: 5개 출력 스타일 (학습, 페어, 초보)

**Memory File (`CLAUDE.md`)**: Claude에게 프로젝트 컨텍스트를 전달하는 핵심 메모리 파일입니다. 프로젝트 정보, 개발 철학, @TAG 시스템, TRUST 5원칙 등이 포함됩니다. 이 파일이 없으면 Claude가 프로젝트를 이해하지 못하므로 필수입니다.

**Git Repository (`.git/`)**: 버전 관리 시스템입니다. MoAI-ADK의 3단계 워크플로우는 Git 브랜치 전략과 밀접하게 통합되어 있으므로 Git 저장소가 있는 것이 강력히 권장됩니다.

### 3. 버전 정보 (🧭 Versions)

```
🧭 Versions:
   Package: v0.0.1
   Templates: v0.0.1
```

**Package**: 현재 설치된 MoAI-ADK npm 패키지 버전입니다. `package.json`에서 읽어옵니다. 이 버전은 `npm install -g moai-adk@<version>` 또는 `bun add -g moai-adk@<version>` 명령으로 설치한 CLI 도구의 버전을 나타냅니다.

**Templates**: 프로젝트에 설치된 템플릿 버전입니다. `.moai/version.json` 파일에서 `template_version` 필드를 읽어옵니다. 이 버전은 `moai init` 또는 `moai update` 명령으로 프로젝트에 복사된 템플릿 파일들의 버전을 나타냅니다.

**버전 불일치**: 두 버전이 다르면 `outdated: true`가 설정되고 경고 메시지가 표시됩니다. 예를 들어, MoAI-ADK 패키지를 v0.0.2로 업그레이드했지만 프로젝트 템플릿은 v0.0.1인 경우, `moai update`를 실행하여 템플릿을 최신 버전으로 업데이트할 수 있습니다.

**Available template update**: 사용 가능한 업데이트 버전이 표시됩니다. 이는 패키지 버전과 템플릿 버전을 비교하여 결정됩니다. 새로운 기능, 버그 수정, 보안 패치가 포함될 수 있으므로 정기적으로 업데이트하는 것이 권장됩니다.

### 4. 파일 개수 (📁 File Counts) - verbose 모드

```
📁 File Counts:
   .moai: 47 files
   .claude: 23 files
   CLAUDE.md: 1 files
```

`countProjectFiles()` 메서드가 각 디렉토리를 재귀적으로 스캔하여 파일 개수를 계산합니다. 이 정보는 프로젝트 구조가 표준 템플릿과 일치하는지 확인하는 데 유용합니다.

**예상 파일 개수** (표준 템플릿 기준):
- `.moai`: 약 40-50개 (메모리 파일, 프로젝트 문서, 스크립트)
- `.claude`: 약 20-30개 (에이전트, 명령어, 훅, 출력 스타일)
- `CLAUDE.md`: 1개 (메인 메모리 파일)

파일 개수가 예상 범위를 크게 벗어나면:
- **너무 적음**: 템플릿이 불완전하게 설치되었을 수 있습니다. `moai update --force`로 재설치를 시도하세요.
- **너무 많음**: 사용자 정의 파일이나 백업 파일이 포함되었을 수 있습니다. 불필요한 파일을 정리하세요.

이 통계는 `--verbose` 플래그를 사용할 때만 표시되며, 일상적인 확인에는 필요하지 않으므로 기본적으로 숨겨져 있습니다.

### 5. 권장 사항 (Recommendations)

```
💡 Recommendations:
   - Run 'moai init' to initialize MoAI-ADK
   - Initialize Git repository: git init
```

권장 사항은 프로젝트 상태 분석 결과를 바탕으로 자동 생성됩니다:

**MoAI 미초기화** (`!status.moaiInitialized`): `.moai` 디렉토리가 없으면 `moai init` 실행을 권장합니다. 이 명령어는 필요한 모든 디렉토리 구조와 설정 파일을 생성합니다.

**Git 미초기화** (`!status.gitRepository`): `.git` 디렉토리가 없으면 `git init` 실행을 권장합니다. Git은 MoAI-ADK의 브랜치 기반 워크플로우에 필수적이며, `/moai:1-spec`과 `/moai:2-build` 단계에서 자동으로 브랜치를 생성합니다.

**템플릿 업데이트** (`versions.outdated`): 템플릿 버전이 패키지 버전보다 낮으면 `moai update` 실행을 권장합니다. 최신 에이전트 개선 사항과 버그 수정을 받을 수 있습니다.

이러한 권장 사항은 실행 가능한 명령어 형태로 제공되어, 사용자가 복사하여 즉시 실행할 수 있습니다. 각 권장 사항은 프로젝트를 Full MoAI Project 상태로 만들기 위한 단계입니다.

## 프로젝트 타입별 시나리오

각 프로젝트 타입에서 할 수 있는 작업과 제한 사항을 상세히 설명합니다.

### Full MoAI Project

**상태**: `.moai` ✅ `.claude` ✅ `CLAUDE.md` ✅ `.git` ✅

**사용 가능한 기능**:
- ✅ 3단계 워크플로우 (`/moai:1-spec`, `/moai:2-build`, `/moai:3-sync`)
- ✅ 7개 전문 에이전트 (`spec-builder`, `code-builder`, `doc-syncer`, 등)
- ✅ SPEC-First TDD 방법론
- ✅ @TAG 추적성 시스템
- ✅ Git 브랜치 자동 관리
- ✅ Living Document 자동 동기화
- ✅ TRUST 5원칙 자동 검증

**권장 작업**: 정기적으로 `moai status`와 `moai doctor`를 실행하여 환경 최신 상태 유지

### Partial MoAI Project

**상태**: `.moai` ✅ `.claude` ❌

**제한 사항**:
- ❌ Claude Code 통합 불가
- ❌ 전문 에이전트 사용 불가
- ❌ 슬래시 명령어 사용 불가
- ❌ 이벤트 훅 작동 안 함

**사용 가능한 기능**:
- ✅ MoAI CLI 명령어 (`moai init`, `doctor`, `status`, `update`, `restore`)
- ✅ 프로젝트 문서 구조 (`.moai/project/`)
- ✅ 스크립트 수동 실행 (`.moai/scripts/`)

**Full로 전환**: `moai init` 실행 (기존 `.moai` 유지하면서 `.claude` 추가)

### Claude Project

**상태**: `.moai` ❌ `.claude` ✅

**제한 사항**:
- ❌ SPEC-First TDD 워크플로우 불가
- ❌ @TAG 시스템 없음
- ❌ 프로젝트 문서 구조 없음
- ❌ MoAI 자동화 스크립트 없음

**사용 가능한 기능**:
- ✅ 기본 Claude Code 기능 (에이전트가 MoAI 전용이 아닌 경우)
- ✅ 커스텀 에이전트 (`.claude/agents/`에 있는 경우)

**Full로 전환**: `moai init` 실행 (기존 `.claude` 유지하면서 `.moai` 추가)

### Regular Directory

**상태**: `.moai` ❌ `.claude` ❌

**제한 사항**:
- ❌ MoAI-ADK 기능 전혀 사용 불가
- ❌ Claude Code 통합 없음

**사용 가능한 기능**:
- ✅ MoAI CLI 명령어 일부 (`moai init`, `moai doctor`)

**Full로 전환**: `moai init` 실행 (처음부터 모든 구조 생성)

## 문제 해결

`moai status` 실행 중 발생할 수 있는 문제와 해결 방법을 안내합니다.

### 1. "Not a MoAI-ADK project" 메시지

**증상**: 프로젝트가 초기화되지 않았다는 메시지

**원인**: `.moai` 디렉토리가 존재하지 않음

**해결책**:

```bash
# 1. MoAI-ADK 초기화
moai init

# 2. 초기화 성공 확인
moai status

# 3. 모든 컴포넌트가 ✅인지 확인
```

초기화 후에도 문제가 지속되면 `.moai` 디렉토리 생성 권한을 확인하세요.

### 2. Claude Integration ❌ 표시

**증상**: `.claude` 디렉토리가 감지되지 않음

**원인**: Claude Code 통합이 설치되지 않았거나 삭제됨

**해결책**:

```bash
# 1. 다시 초기화 (기존 .moai 유지)
moai init

# 2. Claude 디렉토리 존재 확인
ls -la .claude/

# 3. 에이전트 파일 확인
ls -la .claude/agents/moai/
```

`.claude` 디렉토리를 수동으로 삭제했다면 `moai init`이 자동으로 복원합니다.

### 3. Memory File ❌ 표시

**증상**: `CLAUDE.md` 파일이 감지되지 않음

**원인**: 메모리 파일이 삭제되었거나 이름이 변경됨

**해결책**:

```bash
# 1. 템플릿에서 CLAUDE.md 복원
moai update --resources-only

# 2. 또는 수동으로 생성
touch CLAUDE.md

# 3. 파일 존재 확인
moai status
```

`CLAUDE.md`는 프로젝트 컨텍스트를 Claude에게 전달하는 핵심 파일이므로 반드시 필요합니다.

### 4. Git Repository ❌ 표시

**증상**: Git 저장소가 감지되지 않음

**원인**: Git이 초기화되지 않았거나 `.git` 디렉토리가 삭제됨

**해결책**:

```bash
# 1. Git 초기화
git init

# 2. 초기 커밋
git add .
git commit -m "Initial commit"

# 3. 상태 확인
moai status
```

Git은 MoAI-ADK의 브랜치 기반 워크플로우에 필수적이므로 초기화하는 것이 강력히 권장됩니다.

### 5. Templates Outdated 경고

**증상**: "Templates are outdated" 경고 메시지

**원인**: 프로젝트 템플릿 버전이 패키지 버전보다 낮음

**해결책**:

```bash
# 1. 사용 가능한 업데이트 확인
moai update --check

# 2. 템플릿 업데이트
moai update

# 3. 업데이트 확인
moai status
```

업데이트 전 자동으로 백업이 생성되므로 안전하게 업데이트할 수 있습니다.

### 6. 파일 개수가 예상보다 적음 (verbose)

**증상**: `.moai`나 `.claude` 파일 개수가 10개 미만

**원인**: 템플릿 설치가 불완전하거나 파일이 삭제됨

**해결책**:

```bash
# 1. 백업 생성 (안전을 위해)
cp -r .moai .moai.backup
cp -r .claude .claude.backup

# 2. 강제 업데이트
moai update --force

# 3. 파일 개수 확인
moai status --verbose

# 4. 문제 지속 시 재초기화
rm -rf .moai .claude
moai init
```

강제 업데이트는 모든 템플릿 파일을 덮어쓰므로 사용자 정의 수정 사항이 손실될 수 있습니다.

### 7. 권한 오류

**증상**: "EACCES: permission denied" 오류

**원인**: 디렉토리 읽기 권한이 없음

**해결책**:

```bash
# 1. 현재 디렉토리 권한 확인
ls -la

# 2. 필요시 권한 부여
chmod -R u+r .moai .claude

# 3. 상태 재확인
moai status

# 4. 여전히 실패하면 sudo 사용 (비권장)
# sudo moai status
```

권한 문제는 주로 CI/CD 환경이나 다른 사용자가 생성한 프로젝트에서 발생합니다.

## 고급 사용법

### 스크립트에서 상태 확인

`moai status`의 결과를 스크립트에서 활용할 수 있습니다.

```bash
#!/bin/bash
# check-moai-setup.sh

# 상태 확인 및 출력 파싱
output=$(moai status)

# MoAI System 확인
if echo "$output" | grep -q "MoAI System: ✅"; then
  echo "✅ MoAI System is initialized"
else
  echo "❌ MoAI System is NOT initialized"
  echo "Running moai init..."
  moai init
fi

# Claude Integration 확인
if echo "$output" | grep -q "Claude Integration: ✅"; then
  echo "✅ Claude Integration is active"
else
  echo "❌ Claude Integration is missing"
  exit 1
fi
```

### JSON 출력 (향후 버전)

현재는 사람이 읽기 쉬운 텍스트 형식만 지원하지만, v0.0.4에서는 `--json` 플래그를 지원할 예정입니다.

```bash
# 향후 버전 (v0.0.4)
moai status --json > status.json
```

**예상 JSON 구조**:

```json
{
  "project": {
    "path": "/Users/user/projects/my-project",
    "type": "full"
  },
  "components": {
    "moaiSystem": true,
    "claudeIntegration": true,
    "memoryFile": true,
    "gitRepository": true
  },
  "versions": {
    "package": "0.0.1",
    "templates": "0.0.1",
    "outdated": false
  },
  "fileCounts": {
    ".moai": 47,
    ".claude": 23,
    "CLAUDE.md": 1,
    "total": 71
  },
  "recommendations": []
}
```

JSON 출력은 대시보드, 알림 시스템, 분석 도구와의 통합을 쉽게 만듭니다.

### 멀티 프로젝트 배치 확인

여러 프로젝트의 상태를 한 번에 확인하는 스크립트:

```bash
#!/bin/bash
# check-all-projects.sh

echo "Checking all MoAI-ADK projects..."
echo ""

for dir in */; do
  if [ -d "$dir/.moai" ] || [ -d "$dir/.claude" ]; then
    echo "==== $dir ===="
    moai status --project-path "$dir"
    echo ""
  fi
done
```

이 스크립트는 현재 디렉토리의 모든 하위 디렉토리를 순회하며 MoAI-ADK 프로젝트만 상태를 확인합니다.

### 모니터링 및 알림

프로젝트 상태를 모니터링하고 문제 발생 시 알림을 보내는 시스템:

```bash
#!/bin/bash
# monitor-moai-status.sh

status_output=$(moai status 2>&1)

# 업데이트 필요 검사
if echo "$status_output" | grep -q "outdated"; then
  curl -X POST "https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"MoAI-ADK templates are outdated in $(pwd)\"}"
fi

# 컴포넌트 누락 검사
if echo "$status_output" | grep -q "❌"; then
  curl -X POST "https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"MoAI-ADK setup incomplete in $(pwd)\"}"
fi
```

이 스크립트를 cron 작업으로 설정하여 정기적으로 프로젝트 상태를 모니터링할 수 있습니다.

## 관련 명령어

- [`moai init`](./init.md) - 프로젝트 초기화
- [`moai doctor`](./doctor.md) - 시스템 요구사항 진단
- [`moai update`](./update.md) - 템플릿 및 리소스 업데이트
- [`moai restore`](./restore.md) - 백업에서 복원

## 참고 자료

- [MoAI-ADK 공식 문서](https://adk.mo.ai.kr)
- [커뮤니티 포럼](https://mo.ai.kr) (오픈 예정)
- [3단계 워크플로우 가이드](/guide/workflow.md)
- [프로젝트 구조 이해하기](/getting-started/project-structure.md)

---

`moai status`는 프로젝트의 MoAI-ADK 설정 상태를 빠르게 파악하는 필수 도구입니다. 프로젝트를 처음 클론한 후, 오랜만에 작업을 재개할 때, 또는 팀원과 환경을 동기화할 때 실행하세요. 문제가 발생하면 [GitHub Issues](https://github.com/your-org/moai-adk/issues)에 보고해 주시기 바랍니다.