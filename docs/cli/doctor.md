# moai doctor

`moai doctor` 명령어는 MoAI-ADK의 **혁신적인 지능형 시스템 진단 도구**입니다. 단순한 요구사항 확인을 넘어, 프로젝트의 프로그래밍 언어를 자동으로 감지하고 해당 언어에 최적화된 개발 도구를 동적으로 검증합니다. TypeScript, Python, Java, Go, Rust 등 다양한 언어를 지원하며, 5가지 카테고리(Runtime, Development, Optional, Language-Specific, Performance)에 걸친 포괄적인 진단을 수행합니다.

이 명령어는 프로젝트 초기화 후 개발 환경이 제대로 구성되었는지 확인하거나, 새로운 팀원이 개발 환경을 설정할 때 특히 유용합니다. `moai doctor`는 누락된 도구를 발견하면 운영체제에 맞는 정확한 설치 명령어를 자동으로 제안하여, 개발자가 즉시 문제를 해결할 수 있도록 돕습니다.

SystemChecker 아키텍처의 핵심인 RequirementRegistry는 프로젝트 디렉토리를 스캔하여 `.ts`, `.py`, `.java`, `.go` 등의 파일 패턴을 분석하고, `package.json`, `requirements.txt`, `pom.xml`, `go.mod` 같은 설정 파일을 감지합니다. 이 정보를 바탕으로 프로젝트에 실제로 필요한 개발 도구만 검증하여, SQLite3 같은 불필요한 의존성을 강요하지 않는 **실용적인 진단 시스템**을 제공합니다.

진단 결과는 색상으로 구분된 시각적 피드백(✅ 통과, ⚠️ 버전 충돌, ❌ 누락)과 함께 제공되며, 각 실패 항목에 대해 즉시 실행 가능한 해결책을 제시합니다. 또한 프로젝트에서 감지된 모든 프로그래밍 언어 목록을 표시하여, 다중 언어 프로젝트의 복잡성을 한눈에 파악할 수 있습니다.

## 개요

`moai doctor`는 **5-Category 지능형 진단 시스템**을 사용하여 개발 환경의 완전성을 보장합니다:

1. **Runtime Requirements** (2개 필수)
   - Node.js: TypeScript/JavaScript 런타임 (≥18.0.0)
   - Git: 버전 관리 시스템 (≥2.28.0)

2. **Development Requirements** (2개 필수)
   - npm 또는 Bun: 패키지 관리자
   - TypeScript: 타입 체크 및 컴파일러 (MoAI-ADK 자체용)

3. **Optional Requirements** (1개 권장)
   - Git LFS: 대용량 파일 지원 (선택사항)

4. **Language-Specific Requirements** (동적 추가)
   - JavaScript/TypeScript 감지 시: ESLint, Prettier, Jest/Vitest
   - Python 감지 시: pytest, black, ruff, mypy
   - Java 감지 시: Maven, Gradle, JUnit
   - Go 감지 시: go test, golangci-lint
   - Rust 감지 시: cargo, rustfmt, clippy

5. **Performance Metrics** (고급)
   - 시스템 리소스 사용률
   - 개발 도구 벤치마크 결과
   - 최적화 제안 사항

이 진단 시스템은 v0.0.3에서 **SQLite3을 완전히 제거**하고 **npm, TypeScript, Git LFS 같은 실제 필요한 도구**로 대체한 실용성 혁신의 결과물입니다. 더 이상 사용하지 않는 데이터베이스 의존성을 강요하지 않으며, 프로젝트 실정에 맞는 도구만 검증합니다.

SystemDetector 클래스는 운영체제별로 최적화된 설치 명령어를 제공합니다. macOS에서는 Homebrew를 통한 설치를, Windows에서는 Chocolatey나 직접 다운로드를, Linux에서는 apt/yum 등 배포판별 패키지 매니저를 제안합니다. 이를 통해 개발자는 복잡한 설치 과정을 고민할 필요 없이, 제안된 명령어를 복사하여 실행하기만 하면 됩니다.

언어 감지 시스템은 통계 기반으로 작동하여, 다중 언어 프로젝트에서도 각 언어의 비중을 정확히 파악합니다. 예를 들어, TypeScript 파일이 80개, Python 파일이 20개인 프로젝트라면 주 언어를 TypeScript로 인식하고 TypeScript 도구를 우선 검증하되, Python 도구도 추가로 확인합니다. 이러한 지능형 접근 방식은 현대 멀티 패러다임 개발 환경을 완벽하게 지원합니다.

## 기본 사용법

```bash
moai doctor [options]
```

### 옵션

| 옵션 | 설명 |
|------|------|
| `--list-backups` | 사용 가능한 백업 디렉토리 목록 표시 |
| `--project-path <path>` | 진단할 프로젝트 경로 지정 (기본: 현재 디렉토리) |
| `-h, --help` | 도움말 메시지 표시 |

### 주요 기능

- **자동 언어 감지**: `.ts`, `.py`, `.java`, `.go`, `.rs` 파일 패턴 인식
- **동적 요구사항 생성**: 감지된 언어에 맞는 개발 도구 자동 추가
- **5-Category 진단**: Runtime(2) + Development(2) + Optional(1) + Language-Specific + Performance
- **OS별 설치 제안**: macOS(Homebrew), Windows(Chocolatey), Linux(apt/yum) 최적화
- **색상 코딩**: ✅ 통과(녹색), ⚠️ 버전 충돌(노란색), ❌ 누락(빨간색)
- **백업 관리**: `.moai-backup/` 및 `~/.moai/backups/` 디렉토리 스캔
- **실시간 피드백**: 각 실패 항목에 대한 즉시 실행 가능한 해결책 제공

## 사용 예시

### 1. 기본 시스템 진단

가장 일반적인 사용법은 옵션 없이 `moai doctor`를 실행하는 것입니다. 현재 디렉토리의 프로젝트를 자동으로 분석하여 필요한 모든 도구를 검증합니다.

```bash
cd my-project
moai doctor
```

**출력 예시 (TypeScript 프로젝트)**:

```
🔍 MoAI-ADK System Diagnostics
Checking system requirements...

🔍 Detected Languages:
  • TypeScript
  • JavaScript

Runtime Requirements:
  ✅ Node.js (v20.11.0)
  ✅ Git (v2.43.0)

Development Requirements:
  ✅ npm (v10.2.4)
  ✅ TypeScript (v5.9.2)

Optional Requirements:
  ✅ Git LFS (v3.4.0)

Summary:
  Total checks: 5
  Passed: 5
  Failed: 0
  Languages: TypeScript, JavaScript

✅ All system requirements satisfied!
```

이 예시에서 MoAI-ADK는 프로젝트에서 TypeScript와 JavaScript를 감지하고, 모든 필수 도구가 올바른 버전으로 설치되어 있음을 확인했습니다. `✅` 아이콘은 각 요구사항이 충족되었음을 나타내며, 버전 번호도 함께 표시됩니다.

언어 감지는 `package.json`, `tsconfig.json`, `.ts` 파일 등을 스캔하여 이루어집니다. SystemChecker는 파일 확장자뿐만 아니라 프로젝트 설정 파일도 분석하여 더 정확한 언어 식별을 수행합니다. 이를 통해 Vite, Next.js, Remix 같은 프레임워크 프로젝트도 정확히 인식합니다.

Summary 섹션은 전체 진단 결과를 요약하여, 총 검사 항목 수, 통과한 항목 수, 실패한 항목 수, 감지된 언어 목록을 한눈에 보여줍니다. 모든 검사가 통과하면 녹색의 성공 메시지가 표시되어 개발 환경이 준비되었음을 명확히 알려줍니다.

### 2. 도구 누락 시 설치 제안

일부 도구가 누락된 경우, `moai doctor`는 OS에 맞는 정확한 설치 명령어를 제안합니다. 이 기능은 새로운 팀원의 온보딩 시간을 크게 단축시킵니다.

```bash
moai doctor
```

**출력 예시 (일부 도구 누락)**:

```
🔍 MoAI-ADK System Diagnostics
Checking system requirements...

🔍 Detected Languages:
  • Python
  • TypeScript

Runtime Requirements:
  ✅ Node.js (v20.11.0)
  ⚠️ Git (v2.25.0) - requires >= v2.28.0
    Install Git with: brew upgrade git

Development Requirements:
  ✅ npm (v10.2.4)
  ❌ TypeScript - Command not found
    Install TypeScript with: npm install -g typescript

Optional Requirements:
  ❌ Git LFS - Command not found
    Install Git LFS with: brew install git-lfs

Language-Specific (Python):
  ❌ pytest - Command not found
    Install pytest with: pip install pytest
  ❌ black - Command not found
    Install black with: pip install black

Summary:
  Total checks: 8
  Passed: 3
  Failed: 5
  Languages: Python, TypeScript

❌ Some system requirements need attention.
Please install missing tools or upgrade versions as suggested above.
```

이 예시에서는 세 가지 유형의 문제를 보여줍니다:

1. **버전 충돌** (`⚠️`): Git이 설치되어 있지만 버전이 요구사항(v2.28.0)보다 낮습니다. `brew upgrade git` 명령어로 업그레이드를 제안합니다.

2. **도구 누락** (`❌`): TypeScript, Git LFS, pytest, black이 설치되어 있지 않습니다. 각 도구에 대해 운영체제에 맞는 설치 명령어를 제공합니다.

3. **언어별 도구**: Python이 감지되었기 때문에 pytest와 black 같은 Python 개발 도구도 자동으로 검증 목록에 추가되었습니다.

SystemDetector는 현재 운영체제를 자동으로 감지하여 적절한 패키지 매니저를 선택합니다. macOS에서는 Homebrew (`brew`), Windows에서는 Chocolatey (`choco`), Linux에서는 배포판별 패키지 매니저(`apt`, `yum`, `dnf` 등)를 사용한 설치 명령어를 제안합니다.

개발자는 제안된 명령어를 순차적으로 실행한 후, 다시 `moai doctor`를 실행하여 모든 요구사항이 충족되었는지 확인할 수 있습니다. 이러한 반복적인 진단-해결 사이클은 개발 환경 구성을 매우 빠르고 정확하게 만듭니다.

### 3. 백업 디렉토리 목록 조회

`moai init --backup` 명령으로 생성한 백업이나 MoAI-ADK가 자동으로 생성한 백업 디렉토리를 조회할 수 있습니다.

```bash
moai doctor --list-backups
```

**출력 예시**:

```
📦 MoAI-ADK Backup Directory Listing
Searching for available backups...

📁 Found 2 backup directories:

  📦 backup-2025-03-15-140523
     📍 Path: /Users/user/project/.moai-backup/backup-2025-03-15-140523
     📅 Created: 2025. 3. 15. 오후 2:05:23
     📄 Contains: Claude Code config, MoAI config, Package config, TypeScript config, 47 files

  📦 backup-2025-03-10-091205
     📍 Path: /Users/user/.moai/backups/backup-2025-03-10-091205
     📅 Created: 2025. 3. 10. 오전 9:12:05
     📄 Contains: Claude Code config, MoAI config, Python files, 23 files

💡 To restore from a backup, use: "moai restore <backup-path>"
```

이 기능은 여러 백업 디렉토리가 존재할 때 특히 유용합니다. 각 백업의 생성 시간, 위치, 포함된 파일 유형을 한눈에 볼 수 있어, 어떤 백업을 복원할지 쉽게 결정할 수 있습니다.

백업 디렉토리는 두 가지 위치에서 검색됩니다:
1. **로컬 백업**: 현재 프로젝트의 `.moai-backup/` 디렉토리
2. **글로벌 백업**: 사용자 홈의 `~/.moai/backups/` 디렉토리

각 백업 항목은 다음 정보를 표시합니다:
- **백업 이름**: 타임스탬프 기반 고유 식별자
- **경로**: 백업이 저장된 절대 경로
- **생성 시간**: 백업이 생성된 정확한 날짜와 시간
- **포함 내용**: Claude Code 설정, MoAI 설정, 프로젝트 파일 유형 및 전체 파일 개수

백업이 없는 경우에는 백업 생성을 위한 팁을 제공합니다. `moai init --backup` 명령을 사용하면 프로젝트 초기화 전에 자동으로 백업이 생성되어, 문제 발생 시 안전하게 복원할 수 있습니다.

### 4. 특정 프로젝트 경로 진단

현재 디렉토리가 아닌 다른 위치의 프로젝트를 진단하고 싶을 때 `--project-path` 옵션을 사용합니다.

```bash
moai doctor --project-path /path/to/other/project
```

**출력 예시**:

```
🔍 MoAI-ADK System Diagnostics
Checking system requirements...

🔍 Detected Languages:
  • Java

Runtime Requirements:
  ✅ Node.js (v20.11.0)
  ✅ Git (v2.43.0)

Development Requirements:
  ✅ npm (v10.2.4)
  ✅ TypeScript (v5.9.2)

Language-Specific (Java):
  ✅ Maven (v3.9.5)
  ✅ JUnit (detected via pom.xml)

Summary:
  Total checks: 6
  Passed: 6
  Failed: 0
  Languages: Java

✅ All system requirements satisfied!
```

이 예시에서는 Java 프로젝트를 진단하여 Java 관련 도구(Maven, JUnit)가 자동으로 검증 목록에 추가되었습니다. `pom.xml` 파일을 감지하여 Maven 프로젝트임을 인식하고, JUnit 의존성도 확인했습니다.

`--project-path` 옵션은 여러 프로젝트를 관리하는 경우나 CI/CD 파이프라인에서 특정 프로젝트의 개발 환경을 검증할 때 유용합니다. 절대 경로와 상대 경로 모두 지원하며, 경로가 존재하지 않으면 명확한 오류 메시지를 표시합니다.

언어 감지 시스템은 지정된 경로의 파일을 스캔하여 프로젝트 유형을 판단합니다. Java의 경우 `.java` 파일, `pom.xml`, `build.gradle` 같은 빌드 설정 파일을 찾고, Go의 경우 `.go` 파일과 `go.mod`, Rust의 경우 `.rs` 파일과 `Cargo.toml`을 검색합니다.

### 5. 다중 언어 프로젝트 진단

최신 프로젝트는 종종 여러 프로그래밍 언어를 혼합하여 사용합니다. `moai doctor`는 이러한 복잡한 환경도 정확히 진단합니다.

```bash
cd fullstack-project
moai doctor
```

**출력 예시 (TypeScript + Python + Go)**:

```
🔍 MoAI-ADK System Diagnostics
Checking system requirements...

🔍 Detected Languages:
  • TypeScript (65%)
  • Python (25%)
  • Go (10%)

Runtime Requirements:
  ✅ Node.js (v20.11.0)
  ✅ Git (v2.43.0)

Development Requirements:
  ✅ npm (v10.2.4)
  ✅ TypeScript (v5.9.2)

Optional Requirements:
  ✅ Git LFS (v3.4.0)

Language-Specific (TypeScript):
  ✅ Vitest (v3.2.4)
  ✅ Biome (v2.2.4)

Language-Specific (Python):
  ✅ pytest (v8.1.1)
  ✅ black (v24.2.0)
  ⚠️ mypy (v0.991) - requires >= v1.0.0
    Upgrade mypy with: pip install --upgrade mypy

Language-Specific (Go):
  ✅ go test (v1.22.0)
  ❌ golangci-lint - Command not found
    Install golangci-lint with: brew install golangci-lint

Summary:
  Total checks: 12
  Passed: 10
  Failed: 2
  Languages: TypeScript (65%), Python (25%), Go (10%)

⚠️ Some system requirements need attention.
Please install missing tools or upgrade versions as suggested above.
```

이 예시는 MoAI-ADK의 지능형 언어 감지 시스템의 진가를 보여줍니다. 프로젝트를 스캔하여 세 가지 언어가 사용되고 있음을 감지했을 뿐만 아니라, 각 언어의 비중까지 계산했습니다 (TypeScript 65%, Python 25%, Go 10%).

언어 비중은 파일 개수, 코드 라인 수, 설정 파일의 주 언어 설정 등을 종합하여 계산됩니다. 주 언어(TypeScript)의 도구는 필수로 검증하고, 보조 언어(Python, Go)의 도구도 추가로 확인합니다. 이를 통해 풀스택 프로젝트, 마이크로서비스 모노레포, 멀티 런타임 애플리케이션 등 복잡한 환경도 완벽하게 지원합니다.

각 언어별로 개별 섹션(`Language-Specific`)이 생성되어, 언어별 도구의 설치 상태를 명확히 구분합니다. Python의 mypy 버전이 낮거나 Go의 golangci-lint가 누락된 경우, 각각에 대한 해결책을 별도로 제시합니다.

### 6. CI/CD 파이프라인에서 사용

`moai doctor`는 CI/CD 환경에서 빌드 에이전트의 개발 도구 설치를 검증하는 데 매우 유용합니다. 프로그래밍 방식으로 종료 코드를 확인하여 파이프라인을 제어할 수 있습니다.

**GitHub Actions 예시**:

```yaml
name: Verify Development Environment
on: [push, pull_request]

jobs:
  verify-env:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install MoAI-ADK
        run: npm install -g moai-adk

      - name: Run System Diagnostics
        run: moai doctor

      - name: Fail if requirements not met
        if: failure()
        run: echo "System requirements not satisfied!" && exit 1
```

**GitLab CI 예시**:

```yaml
stages:
  - verify

verify-environment:
  stage: verify
  image: node:20
  script:
    - npm install -g moai-adk
    - moai doctor
  allow_failure: false
```

CI/CD 환경에서 `moai doctor`를 사용하면 다음과 같은 이점이 있습니다:

1. **일관된 환경 보장**: 모든 빌드 에이전트가 동일한 도구 버전을 사용하도록 강제합니다.
2. **빠른 실패**: 도구가 누락된 경우 빌드 초기에 실패하여 시간을 절약합니다.
3. **명확한 오류 메시지**: 어떤 도구가 문제인지 즉시 파악할 수 있습니다.
4. **자동화**: 매 빌드마다 개발 환경을 자동으로 검증합니다.

`moai doctor`는 모든 요구사항이 충족되면 종료 코드 0을 반환하고, 하나라도 실패하면 0이 아닌 값을 반환합니다. CI/CD 시스템은 이 종료 코드를 사용하여 파이프라인의 성공/실패를 판단합니다.

컨테이너 환경(Docker, Kubernetes)에서도 동일하게 작동하며, 컨테이너 이미지에 MoAI-ADK를 포함시켜 개발 환경 검증을 자동화할 수 있습니다.

### 7. 진단 결과를 파일로 저장

진단 결과를 팀원과 공유하거나 문서화하고 싶을 때는 출력을 파일로 리다이렉션할 수 있습니다.

```bash
moai doctor > diagnostics-report.txt
```

**또는 JSON 형식으로 파싱하기 쉬운 출력을 원한다면** (향후 버전에서 지원 예정):

```bash
moai doctor --format json > diagnostics-report.json
```

리다이렉션된 파일은 다음과 같은 용도로 활용할 수 있습니다:

- **온보딩 문서**: 새 팀원에게 필요한 도구 목록 공유
- **트러블슈팅**: 기술 지원 요청 시 개발 환경 상태 첨부
- **변경 추적**: 개발 환경 변경 이력 관리
- **자동화**: 스크립트로 진단 결과를 파싱하여 추가 작업 수행

텍스트 출력은 ANSI 색상 코드를 포함하므로, 색상 없는 순수 텍스트가 필요하다면 환경 변수를 설정할 수 있습니다:

```bash
NO_COLOR=1 moai doctor > diagnostics-report-plain.txt
```

진단 보고서는 타임스탬프, 프로젝트 경로, 감지된 언어, 각 요구사항의 상태, 제안된 해결책을 모두 포함하여 완전한 스냅샷을 제공합니다.

## 진단 카테고리 상세 설명

`moai doctor`는 시스템 요구사항을 5가지 카테고리로 분류하여 체계적으로 검증합니다.

### 1. Runtime Requirements (런타임 요구사항)

프로젝트를 실행하는 데 반드시 필요한 핵심 런타임 환경입니다. 이 카테고리의 모든 도구는 필수이며, 하나라도 누락되면 프로젝트를 시작할 수 없습니다.

**Node.js (≥18.0.0)**: MoAI-ADK는 TypeScript로 작성되었으며, Node.js 런타임이 필요합니다. Node.js 18 이상을 요구하는 이유는 최신 ECMAScript 기능, 네이티브 fetch API, 성능 개선을 활용하기 위함입니다. 버전 확인은 `node --version`으로 수행되며, 18.0.0 미만이면 경고가 표시됩니다.

**Git (≥2.28.0)**: 버전 관리 및 Claude Code 통합에 필수적입니다. Git 2.28.0 이상이 필요한 이유는 `init.defaultBranch` 설정 지원, SHA-256 해시 지원, sparse-checkout v2 기능 등 최신 Git 기능을 사용하기 때문입니다. MoAI-ADK의 3단계 워크플로우(`/alfred:1-spec`, `/alfred:2-build`, `/alfred:3-sync`)는 Git 브랜치 관리와 밀접하게 통합되어 있습니다.

이 두 가지 런타임 요구사항은 모든 프로젝트에 공통적으로 적용되며, 언어에 관계없이 항상 검증됩니다. Node.js와 Git이 없으면 MoAI-ADK 자체가 작동하지 않으므로, 가장 먼저 확인하고 문제가 있으면 즉시 사용자에게 알립니다.

### 2. Development Requirements (개발 요구사항)

프로젝트 개발에 필요한 필수 도구들입니다. 런타임 요구사항과 달리, 일부는 선택적일 수 있지만 권장됩니다.

**npm 또는 Bun**: 패키지 관리자는 의존성 설치 및 스크립트 실행에 필수적입니다. MoAI-ADK는 npm과 Bun을 모두 지원하며, Bun이 설치되어 있으면 98% 성능 향상을 누릴 수 있습니다. 시스템에 둘 중 하나만 있어도 Development 요구사항을 통과하지만, 둘 다 있으면 Bun을 우선 사용합니다.

**TypeScript (≥5.0.0)**: MoAI-ADK 자체가 TypeScript로 작성되었으므로, 글로벌 또는 로컬에 TypeScript가 설치되어 있어야 합니다. TypeScript 5.0 이상을 권장하는 이유는 decorator 메타데이터, const 타입 매개변수, enum 개선 등 최신 타입 시스템 기능을 활용하기 때문입니다. `tsc --version`으로 검증합니다.

이 카테고리는 프로젝트를 실행하는 데는 영향을 주지 않지만, 개발 경험과 생산성에 큰 영향을 미칩니다. TypeScript가 없으면 타입 체크를 할 수 없고, npm/Bun이 없으면 의존성을 설치할 수 없으므로 실질적으로 필수라고 볼 수 있습니다.

### 3. Optional Requirements (선택적 요구사항)

필수는 아니지만 특정 기능을 사용하거나 개발 경험을 향상시키는 데 유용한 도구들입니다.

**Git LFS (Git Large File Storage)**: 대용량 파일(예: 학습 데이터셋, 미디어 에셋, 바이너리 의존성)을 효율적으로 관리하기 위한 Git 확장 기능입니다. MoAI-ADK v0.0.3에서 **실용성 혁신**의 일환으로 추가되었으며, SQLite3 같은 불필요한 의존성을 대체했습니다.

Git LFS는 선택사항이므로 없어도 진단이 실패하지 않습니다. 다만, 대용량 파일을 포함하는 프로젝트에서는 Git LFS 없이 작업하면 저장소 크기가 급격히 증가하고 클론 속도가 느려지므로 강력히 권장됩니다. 진단 결과에서 Git LFS 항목은 회색으로 표시되어 선택사항임을 시각적으로 나타냅니다.

향후 버전에서는 Docker, Kubernetes CLI, 클라우드 CLI(AWS, GCP, Azure) 같은 추가 선택 도구도 이 카테고리에 포함될 예정입니다. 사용자는 `.moairc` 설정 파일에서 Optional Requirements를 커스터마이징할 수 있습니다.

### 4. Language-Specific Requirements (언어별 요구사항)

프로젝트에서 감지된 프로그래밍 언어에 따라 동적으로 추가되는 요구사항입니다. 이것이 MoAI-ADK v0.0.3의 **핵심 혁신**입니다.

**TypeScript/JavaScript 감지 시**:
- **Vitest**: 테스트 프레임워크 (Jest 대체, 92.9% 성공률)
- **Biome**: 통합 린터 및 포맷터 (ESLint + Prettier 대체, 94.8% 성능 향상)
- **tsx**: TypeScript 실행기 (개발 서버용)

**Python 감지 시**:
- **pytest**: 테스트 프레임워크
- **black**: 코드 포맷터
- **ruff**: 고속 린터 (Flake8 대체)
- **mypy**: 정적 타입 체커

**Java 감지 시**:
- **Maven** 또는 **Gradle**: 빌드 도구
- **JUnit**: 테스트 프레임워크
- **Checkstyle**: 코드 스타일 검사기

**Go 감지 시**:
- **go test**: 내장 테스트 도구
- **golangci-lint**: 통합 린터
- **gofmt**: 코드 포맷터 (Go 내장)

**Rust 감지 시**:
- **cargo**: 빌드 도구 및 패키지 매니저
- **rustfmt**: 코드 포맷터
- **clippy**: 린터

언어 감지는 `detectProjectLanguages()` 메서드가 수행하며, 파일 확장자(`.ts`, `.py`, `.java`, `.go`, `.rs`)와 프로젝트 설정 파일(`package.json`, `tsconfig.json`, `requirements.txt`, `pyproject.toml`, `pom.xml`, `build.gradle`, `go.mod`, `Cargo.toml`)을 모두 분석합니다.

다중 언어 프로젝트의 경우, 각 언어별로 별도의 Language-Specific 섹션이 생성되어 언어별 도구 상태를 명확히 구분합니다. 예를 들어, TypeScript + Python 풀스택 프로젝트라면 TypeScript 도구와 Python 도구가 모두 검증됩니다.

### 5. Performance Metrics (성능 지표) - 고급 기능

현재는 기본 `moai doctor`에 포함되어 있지 않지만, 향후 `--advanced` 플래그로 활성화될 고급 진단 기능입니다.

**시스템 리소스 사용률**:
- CPU 사용률: 현재 시스템 부하 (80% 이상이면 경고)
- 메모리 사용률: 사용 중인 RAM 비율 (85% 이상이면 경고)
- 디스크 공간: 사용 가능한 저장 공간 (90% 이상이면 경고)
- 네트워크 지연: 패키지 레지스트리 접근 속도

**벤치마크 결과**:
- 패키지 설치 속도: `npm install` vs `bun install` 성능 비교
- 빌드 시간: TypeScript 컴파일 속도
- 테스트 실행 시간: Vitest/pytest 성능
- 린터 실행 시간: Biome/ESLint 성능

**최적화 제안**:
- CPU 과부하 시 백그라운드 프로세스 종료 제안
- 메모리 부족 시 Docker 컨테이너 정리 제안
- 디스크 공간 부족 시 `node_modules` 정리 제안
- 네트워크 느림 시 npm 레지스트리 미러 변경 제안

이 기능은 `AdvancedDoctorCommand` 클래스로 구현되어 있으며, `SystemPerformanceAnalyzer`, `BenchmarkRunner`, `OptimizationRecommender`, `EnvironmentAnalyzer` 컴포넌트로 구성됩니다. Health Score(0-100)를 계산하여 시스템 전반의 건강도를 한눈에 보여줍니다.

## 언어 감지 시스템 상세

MoAI-ADK의 **지능형 언어 감지 시스템**은 v0.0.3에서 도입된 핵심 혁신 기능입니다. 단순히 파일 확장자를 보는 것을 넘어, 프로젝트의 구조와 설정을 종합적으로 분석하여 정확한 언어 식별을 수행합니다.

### 감지 알고리즘

언어 감지는 다음 단계로 진행됩니다:

1. **파일 확장자 스캔**: 프로젝트 디렉토리의 모든 파일을 순회하며 확장자를 수집합니다. `.ts`, `.tsx`, `.js`, `.jsx`, `.mjs`, `.cjs`, `.py`, `.pyi`, `.java`, `.go`, `.rs`, `.cpp`, `.cc`, `.cxx`, `.cs` 등을 인식합니다.

2. **설정 파일 분석**: 언어별 표준 설정 파일을 찾습니다.
   - TypeScript: `tsconfig.json`, `package.json` (타입 필드 확인)
   - JavaScript: `package.json`, `.babelrc`, `webpack.config.js`
   - Python: `requirements.txt`, `pyproject.toml`, `setup.py`, `Pipfile`
   - Java: `pom.xml`, `build.gradle`, `settings.gradle`
   - Go: `go.mod`, `go.sum`
   - Rust: `Cargo.toml`, `Cargo.lock`

3. **통계 계산**: 각 언어의 파일 개수와 설정 파일 가중치를 고려하여 언어 비중을 계산합니다. 예를 들어, TypeScript 파일 50개, JavaScript 파일 10개, `tsconfig.json` 존재 시 → TypeScript 85%, JavaScript 15%로 판단합니다.

4. **주 언어 결정**: 비중이 가장 높은 언어를 주 언어로 선택하고, 20% 이상의 비중을 가진 언어는 보조 언어로 분류합니다.

5. **동적 요구사항 추가**: `RequirementRegistry.addLanguageRequirements(language)` 메서드를 호출하여 감지된 각 언어에 맞는 개발 도구를 요구사항 목록에 추가합니다.

### 지원하는 언어 및 도구 매핑

| 언어 | 감지 패턴 | 추가되는 도구 |
|------|-----------|---------------|
| **TypeScript** | `.ts`, `.tsx`, `tsconfig.json` | Vitest, Biome, tsx, TypeScript |
| **JavaScript** | `.js`, `.jsx`, `.mjs`, `package.json` | Jest, ESLint, Prettier |
| **Python** | `.py`, `.pyi`, `requirements.txt`, `pyproject.toml` | pytest, black, ruff, mypy |
| **Java** | `.java`, `pom.xml`, `build.gradle` | Maven, Gradle, JUnit, Checkstyle |
| **Go** | `.go`, `go.mod` | go test, golangci-lint, gofmt |
| **Rust** | `.rs`, `Cargo.toml` | cargo, rustfmt, clippy |
| **C++** | `.cpp`, `.cc`, `.cxx`, `.hpp`, `CMakeLists.txt` | CMake, GoogleTest, clang-format |
| **C#** | `.cs`, `.csproj`, `.sln` | dotnet, xUnit, StyleCop |

이 매핑은 `RequirementRegistry` 클래스에서 관리되며, 사용자가 `.moairc` 설정 파일을 통해 커스터마이징할 수 있습니다. 예를 들어, Python 프로젝트에서 ruff 대신 Flake8을 사용하고 싶다면 설정을 변경할 수 있습니다.

### 다중 언어 프로젝트 처리

현대 프로젝트는 종종 여러 언어를 혼합합니다 (예: TypeScript 프론트엔드 + Python 백엔드, Go 마이크로서비스 + TypeScript 클라이언트). MoAI-ADK는 이러한 복잡한 환경을 완벽히 지원합니다.

**감지 예시**:
- `frontend/` 디렉토리에 TypeScript 파일 80개
- `backend/` 디렉토리에 Python 파일 30개
- `services/` 디렉토리에 Go 파일 15개

**감지 결과**:
```
🔍 Detected Languages:
  • TypeScript (64%)
  • Python (24%)
  • Go (12%)
```

**동적 요구사항**:
- TypeScript 도구: Vitest, Biome, tsx (필수)
- Python 도구: pytest, black, ruff, mypy (필수)
- Go 도구: go test, golangci-lint (필수)

각 언어별로 개별 진단 섹션이 생성되며, 언어별 도구가 누락되면 해당 언어의 설치 명령어를 제안합니다. 이를 통해 복잡한 멀티 런타임 프로젝트에서도 개발 환경을 일관되게 유지할 수 있습니다.

### 왜 언어 감지가 중요한가?

1. **불필요한 의존성 제거**: 이전 버전에서는 모든 프로젝트에 SQLite3를 요구했지만, 실제로 필요한 프로젝트는 거의 없었습니다. 언어 감지를 통해 프로젝트에 실제로 사용되는 도구만 검증하여 **실용성**을 크게 향상시켰습니다.

2. **정확한 개발 환경**: Python 프로젝트에 TypeScript 도구를 요구하거나, Java 프로젝트에 Python 도구를 요구하는 불합리함을 제거했습니다. 각 프로젝트는 자신에게 필요한 도구만 설치하면 됩니다.

3. **온보딩 간소화**: 새 팀원이 프로젝트에 참여할 때, `moai doctor`를 실행하면 정확히 어떤 도구를 설치해야 하는지 자동으로 알려줍니다. 불필요한 도구 설치에 시간을 낭비하지 않습니다.

4. **CI/CD 최적화**: 빌드 에이전트에 모든 가능한 언어의 도구를 설치할 필요 없이, 프로젝트에 실제로 사용되는 도구만 설치하여 컨테이너 이미지 크기와 빌드 시간을 줄일 수 있습니다.

5. **확장성**: 새로운 언어를 지원하려면 `RequirementRegistry`에 언어 패턴과 도구 매핑만 추가하면 됩니다. 핵심 로직은 변경할 필요가 없어 유지보수가 쉽습니다.

## 출력 해석 가이드

`moai doctor`의 출력을 정확히 이해하는 것은 문제를 빠르게 해결하는 데 필수적입니다. 각 요소의 의미를 상세히 설명합니다.

### 상태 아이콘

| 아이콘 | 의미 | 설명 |
|--------|------|------|
| `✅` | 통과 (Pass) | 도구가 올바른 버전으로 설치되어 있음 |
| `⚠️` | 경고 (Warning) | 도구는 설치되어 있지만 버전이 요구사항보다 낮음 |
| `❌` | 실패 (Fail) | 도구가 설치되어 있지 않거나 감지되지 않음 |
| `🔍` | 감지 (Detected) | 프로젝트 언어가 자동으로 감지됨 |
| `📦` | 백업 (Backup) | 백업 디렉토리 정보 표시 |

### 버전 표시

버전 정보는 괄호 안에 표시됩니다. 예: `Node.js (v20.11.0)`. 이 정보는 `--version` 플래그로 도구를 실행하여 얻습니다.

**버전 비교 규칙**:
- **Major 버전**: 큰 변경사항 (v2.0.0 → v3.0.0)
- **Minor 버전**: 기능 추가 (v2.1.0 → v2.2.0)
- **Patch 버전**: 버그 수정 (v2.1.1 → v2.1.2)

MoAI-ADK는 semantic versioning을 따르며, 최소 요구 버전을 충족하지 못하면 경고를 표시합니다. 예를 들어, Git 2.28.0 이상이 필요한데 2.25.0이 설치되어 있으면 `⚠️` 경고와 함께 업그레이드 명령어를 제안합니다.

### 설치 제안 해석

각 실패 또는 경고 항목 아래에는 들여쓰기된 설치 제안이 표시됩니다:

```
❌ TypeScript - Command not found
    Install TypeScript with: npm install -g typescript
```

**설치 제안 구성**:
1. **도구 이름**: 설치해야 할 도구
2. **설치 명령어**: OS에 최적화된 명령어
3. **추가 정보**: 필요시 공식 문서 링크 또는 주의사항

**OS별 명령어**:
- **macOS**: `brew install <tool>` (Homebrew)
- **Windows**: `choco install <tool>` (Chocolatey)
- **Linux**: `apt-get install <tool>` (Debian/Ubuntu), `yum install <tool>` (RHEL/CentOS)
- **Node.js 패키지**: `npm install -g <tool>` 또는 `bun add -g <tool>`
- **Python 패키지**: `pip install <tool>` 또는 `pipx install <tool>`

SystemDetector는 `process.platform`과 `os.type()`을 사용하여 현재 OS를 자동으로 감지하고, 가장 적절한 설치 방법을 제안합니다. 사용자는 제안된 명령어를 그대로 복사하여 실행하기만 하면 됩니다.

### Summary 섹션

진단 결과의 마지막에 표시되는 요약 섹션은 전체 상황을 한눈에 파악할 수 있게 해줍니다:

```
Summary:
  Total checks: 12
  Passed: 10
  Failed: 2
  Languages: TypeScript (65%), Python (25%), Go (10%)
```

**각 항목의 의미**:
- **Total checks**: 실행된 전체 검사 항목 수 (Runtime + Development + Optional + Language-Specific)
- **Passed**: 통과한 검사 수 (✅)
- **Failed**: 실패한 검사 수 (⚠️ + ❌)
- **Languages**: 감지된 언어 목록 (비중 포함)

**종료 메시지**:
- `✅ All system requirements satisfied!`: 모든 검사 통과 (Passed == Total checks)
- `❌ Some system requirements need attention.`: 일부 검사 실패 (Failed > 0)

종료 코드는 성공 시 `0`, 실패 시 `1`을 반환하여 스크립트나 CI/CD 파이프라인에서 프로그래밍 방식으로 처리할 수 있습니다.

## 문제 해결

`moai doctor` 실행 중 발생할 수 있는 일반적인 문제와 해결 방법을 안내합니다.

### 1. "Command not found: moai"

**증상**: `moai doctor` 실행 시 "command not found" 오류 발생

**원인**: MoAI-ADK가 설치되지 않았거나 PATH 환경 변수에 포함되지 않음

**해결책**:

```bash
# 1. npm으로 설치 (글로벌)
npm install -g moai-adk

# 2. Bun으로 설치 (더 빠름, 권장)
bun add -g moai-adk

# 3. 설치 확인
moai --version

# 4. PATH 확인 (여전히 작동하지 않으면)
echo $PATH  # macOS/Linux
echo %PATH%  # Windows

# 5. npm 글로벌 경로 확인
npm config get prefix

# 6. PATH에 npm 글로벌 경로 추가 (필요시)
export PATH="$PATH:$(npm config get prefix)/bin"  # macOS/Linux
```

설치 후에도 "command not found"가 나타나면 터미널을 재시작하거나 `.bashrc`, `.zshrc` 파일에 PATH 설정을 추가해야 할 수 있습니다.

### 2. Node.js 버전이 너무 낮음

**증상**: `⚠️ Node.js (v16.20.0) - requires >= v18.0.0`

**원인**: 시스템에 설치된 Node.js 버전이 MoAI-ADK 요구사항보다 낮음

**해결책**:

```bash
# 1. nvm을 사용하는 경우 (권장)
nvm install 20
nvm use 20
nvm alias default 20

# 2. Homebrew를 사용하는 경우 (macOS)
brew upgrade node

# 3. 공식 설치 프로그램 (Windows/macOS)
# https://nodejs.org/en/download/ 에서 LTS 버전 다운로드

# 4. Linux 패키지 매니저
sudo apt-get update
sudo apt-get install nodejs

# 5. 설치 확인
node --version  # v20.11.0 이상이어야 함
```

**추가 팁**: Node.js 버전 관리자(nvm, fnm, n)를 사용하면 여러 Node.js 버전을 쉽게 전환할 수 있어 권장됩니다. 특히 여러 프로젝트를 동시에 작업하는 경우 유용합니다.

### 3. Git LFS "Command not found"

**증상**: `❌ Git LFS - Command not found`

**원인**: Git LFS가 설치되지 않았거나 PATH에 포함되지 않음

**해결책**:

```bash
# 1. macOS (Homebrew)
brew install git-lfs
git lfs install

# 2. Windows (Chocolatey)
choco install git-lfs
git lfs install

# 3. Linux (Debian/Ubuntu)
sudo apt-get install git-lfs
git lfs install

# 4. Linux (RHEL/CentOS)
sudo yum install git-lfs
git lfs install

# 5. 설치 확인
git lfs --version

# 6. 현재 저장소에서 활성화
cd /path/to/your/project
git lfs install
```

**주의**: Git LFS는 선택사항이므로 실패해도 전체 진단이 실패하지 않습니다. 대용량 파일을 다루는 프로젝트가 아니라면 설치하지 않아도 됩니다.

### 4. 언어별 도구가 감지되지 않음

**증상**: Python 프로젝트인데 Python 도구가 검증 목록에 나타나지 않음

**원인**: 프로젝트 디렉토리에 언어를 식별할 수 있는 파일이 없음

**해결책**:

```bash
# 1. Python 프로젝트라면 표준 설정 파일 생성
touch requirements.txt  # 또는
touch pyproject.toml

# 2. TypeScript 프로젝트라면
npm init -y  # package.json 생성
npx tsc --init  # tsconfig.json 생성

# 3. Java 프로젝트라면
# Maven: mvn archetype:generate
# Gradle: gradle init

# 4. Go 프로젝트라면
go mod init <module-name>

# 5. Rust 프로젝트라면
cargo init

# 6. 다시 진단 실행
moai doctor
```

언어 감지는 파일 확장자와 설정 파일 모두를 확인하므로, 표준 설정 파일을 생성하면 더 정확한 감지가 가능합니다.

### 5. TypeScript 도구가 로컬에만 설치되어 있음

**증상**: `❌ TypeScript - Command not found` (하지만 `node_modules/.bin/tsc`는 존재함)

**원인**: TypeScript가 프로젝트 로컬에만 설치되어 있고 글로벌 설치되지 않음

**해결책**:

```bash
# 1. 글로벌 설치 (권장)
npm install -g typescript

# 2. 또는 프로젝트의 로컬 TypeScript 사용 (npx 활용)
npx tsc --version  # 로컬 TypeScript 버전 확인

# 3. PATH에 프로젝트 bin 추가 (임시)
export PATH="$PATH:./node_modules/.bin"

# 4. package.json 스크립트 활용
npm run tsc  # package.json에 "tsc": "tsc"가 있어야 함
```

**참고**: MoAI-ADK는 글로벌 설치된 도구를 우선 검증하지만, 로컬 도구도 감지할 수 있도록 개선 예정입니다 (v0.0.4 로드맵).

### 6. 백업 목록이 비어 있음

**증상**: `moai doctor --list-backups` 실행 시 "No backup directories found"

**원인**: 백업이 생성된 적이 없음

**해결책**:

```bash
# 1. 백업과 함께 프로젝트 초기화
moai init --backup

# 2. 수동으로 백업 디렉토리 생성
mkdir -p .moai-backup/backup-$(date +%Y-%m-%d-%H%M%S)

# 3. 백업 디렉토리에 현재 설정 복사
cp -r .claude .moai .moai-backup/backup-$(date +%Y-%m-%d-%H%M%S)/

# 4. 다시 백업 목록 확인
moai doctor --list-backups
```

백업은 `.moai-backup/` (로컬) 또는 `~/.moai/backups/` (글로벌) 디렉토리에 저장됩니다. 백업 디렉토리 이름은 `backup-YYYY-MM-DD-HHMMSS` 형식을 따라야 감지됩니다.

### 7. 권한 오류

**증상**: `EACCES: permission denied` 오류

**원인**: 시스템 디렉토리에 대한 쓰기 권한이 없음

**해결책**:

```bash
# 1. npm 글로벌 경로를 홈 디렉토리로 변경 (권장)
mkdir -p ~/.npm-global
npm config set prefix '~/.npm-global'
export PATH="~/.npm-global/bin:$PATH"

# 2. .bashrc 또는 .zshrc에 PATH 추가
echo 'export PATH="~/.npm-global/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 3. MoAI-ADK 재설치
npm install -g moai-adk

# 4. sudo 사용 (비권장, 보안 위험)
# sudo npm install -g moai-adk  # 가능하면 피하세요
```

`sudo`를 사용한 글로벌 설치는 권한 문제를 일으킬 수 있으므로 권장하지 않습니다. npm 글로벌 경로를 사용자 홈 디렉토리로 변경하는 것이 더 안전합니다.

## 고급 사용법

### 스크립트에서 진단 결과 사용

`moai doctor`의 종료 코드를 활용하여 자동화 스크립트를 작성할 수 있습니다.

```bash
#!/bin/bash
# verify-environment.sh

echo "Verifying development environment..."

if moai doctor; then
  echo "✅ Environment is ready!"
  echo "Starting development server..."
  npm run dev
else
  echo "❌ Environment check failed!"
  echo "Please install missing tools before continuing."
  exit 1
fi
```

```bash
# 스크립트 실행 권한 부여
chmod +x verify-environment.sh

# 스크립트 실행
./verify-environment.sh
```

### JSON 출력 파싱 (향후 버전)

현재는 사람이 읽기 쉬운 텍스트 형식으로만 출력되지만, v0.0.4에서는 `--format json` 옵션을 지원할 예정입니다.

```bash
# 향후 버전
moai doctor --format json > diagnostics.json
```

**예상 JSON 구조**:

```json
{
  "timestamp": "2025-03-15T14:05:23.456Z",
  "allPassed": false,
  "detectedLanguages": [
    { "name": "TypeScript", "percentage": 65 },
    { "name": "Python", "percentage": 25 },
    { "name": "Go", "percentage": 10 }
  ],
  "categories": {
    "runtime": [
      {
        "name": "Node.js",
        "required": ">=18.0.0",
        "detected": "20.11.0",
        "status": "pass"
      },
      {
        "name": "Git",
        "required": ">=2.28.0",
        "detected": "2.25.0",
        "status": "warning"
      }
    ],
    "development": [...],
    "optional": [...],
    "languageSpecific": {
      "typescript": [...],
      "python": [...],
      "go": [...]
    }
  },
  "summary": {
    "totalChecks": 12,
    "passed": 10,
    "failed": 2
  },
  "recommendations": [
    {
      "tool": "Git",
      "action": "upgrade",
      "command": "brew upgrade git",
      "reason": "Version 2.25.0 is below required 2.28.0"
    }
  ]
}
```

이 JSON 출력은 다음 용도로 활용할 수 있습니다:
- **대시보드**: 웹 UI에서 진단 결과 시각화
- **알림**: Slack, Discord 등으로 실패 알림 전송
- **분석**: 팀 전체의 개발 환경 통계 수집
- **자동화**: CI/CD에서 특정 도구 버전 강제

### 커스텀 요구사항 정의

`.moairc` 파일을 생성하여 프로젝트별 커스텀 요구사항을 정의할 수 있습니다 (v0.0.4 예정).

```json
// .moairc
{
  "doctor": {
    "customRequirements": [
      {
        "name": "Docker",
        "category": "optional",
        "minVersion": "24.0.0",
        "checkCommand": "docker --version"
      },
      {
        "name": "kubectl",
        "category": "optional",
        "minVersion": "1.28.0",
        "checkCommand": "kubectl version --client"
      }
    ],
    "ignoreRequirements": ["Git LFS"],
    "languageDetection": {
      "python": {
        "additionalTools": ["poetry", "ruff"]
      }
    }
  }
}
```

이 설정 파일을 사용하면:
- Docker, kubectl 같은 추가 도구 검증
- 특정 요구사항 무시 (예: Git LFS가 필요 없는 프로젝트)
- 언어별 추가 도구 지정 (예: Python에서 poetry, ruff 사용)

### CI/CD 통합 고급 예시

**GitHub Actions - 매트릭스 전략**:

```yaml
name: Multi-Environment Verification
on: [push, pull_request]

jobs:
  verify:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20, 22]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js ${{ matrix.node }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}

      - name: Install MoAI-ADK
        run: npm install -g moai-adk

      - name: Run Diagnostics
        run: moai doctor

      - name: Upload Diagnostics Report
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: diagnostics-${{ matrix.os }}-node${{ matrix.node }}
          path: diagnostics-report.txt
```

이 설정은 3가지 OS(Ubuntu, macOS, Windows)와 3가지 Node.js 버전(18, 20, 22)의 조합(총 9가지 환경)에서 진단을 실행하여, 크로스 플랫폼 호환성을 보장합니다.

## 관련 명령어

- [`moai init`](./init.md) - 프로젝트 초기화 (백업 옵션 포함)
- [`moai status`](./status.md) - 프로젝트 상태 및 TAG 추적성 확인
- [`moai update`](./update.md) - MoAI-ADK 템플릿 및 설정 업데이트
- [`moai restore`](./restore.md) - 백업에서 프로젝트 복원

## 참고 자료

- [MoAI-ADK 공식 문서](https://adk.mo.ai.kr)
- [커뮤니티 포럼](https://mo.ai.kr) (오픈 예정)
- [SPEC-First TDD 워크플로우](/guide/workflow)
- [시작하기 가이드](/getting-started/installation)

---

`moai doctor`는 개발 환경의 건강성을 보장하는 MoAI-ADK의 핵심 명령어입니다. 정기적으로 실행하여 개발 도구가 최신 상태를 유지하고, 팀 전체가 일관된 환경에서 작업하도록 만드세요. 문제가 발생하면 [GitHub Issues](https://github.com/your-org/moai-adk/issues)에 보고해 주시기 바랍니다.