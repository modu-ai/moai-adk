# 🤝 Contributing to MoAI-ADK

**English version below | 아래에 한국어 버전이 있습니다**

---

## 📑 Table of Contents

- [English](#english-version)
- [한국어](#한국어-버전)

---

# English Version

## 🤝 Contributing to MoAI-ADK

Thank you for contributing to the MoAI-ADK project! This document guides you on how to effectively contribute to the project.

---

## 📋 Table of Contents (English)

- [Issue Creation Guide](#issue-creation-guide)
  - [Bug Report](#bug-report)
  - [Feature Request](#feature-request)
- [Pull Request Guide](#pull-request-guide)
- [Development Environment Setup](#development-environment-setup)
- [Code Contribution Guide](#code-contribution-guide)

---

## Issue Creation Guide

### Bug Report

Found a bug? Please create an issue with the following information:

**Title Format**: `[Bug] Brief description of the bug`

**Required Information**:

```markdown
## 🐛 Bug Description

Provide a clear and concise description of what the bug is.

## 🔄 Steps to Reproduce

1. What command did you execute?
2. What input did you provide?
3. What action did you perform?
4. At what point did the error occur?

## 💥 Expected vs Actual Behavior

- **Expected Behavior**: How should it work?
- **Actual Behavior**: How did it actually work?

## 🖥️ Environment Information

- **OS**: (e.g., macOS 14.0, Ubuntu 22.04, Windows 11)
- **Python Version**: (e.g., 3.11.0)
- **MoAI-ADK Version**: (e.g., v0.14.0)
- **Claude Code Version**: (Optional)

## 📸 Screenshots or Logs

Please attach error messages, screenshots, or logs if possible.

```bash
# Example error log
Error: Cannot find module '...'
    at Function.Module._resolveFilename ...
```

## 🔍 Additional Information

Provide any additional context or information related to the bug.
```

**Example**:

```markdown
## 🐛 Bug Description

Executing `/moai:2-run` command fails during the implementation validation step.

## 🔄 Steps to Reproduce

1. Initialize project with `python -m moai_adk init .`
2. Run `/moai:1-plan "User Authentication"` to create Plan & SPEC
3. Execute `/moai:2-run SPEC-AUTH-001`
4. Error occurs during implementation validation step

## 💥 Expected vs Actual Behavior

- **Expected Behavior**: TDD cycle should complete successfully
- **Actual Behavior**: Implementation validation failed

## 🖥️ Environment Information

- **OS**: macOS 14.2
- **Python Version**: 3.11.0
- **MoAI-ADK Version**: v0.14.0

## 📸 Screenshots or Logs

```bash
Error: Implementation validation failed
Please ensure all tests are passing before proceeding
```
```

---

### Feature Request

Want to propose a new feature?

**Title Format**: `[Feature Request] Feature Name`

**Required Information**:

```markdown
## 💡 Feature Proposal

Provide a clear and concise description of the proposed feature.

## 🎯 Problem This Solves

What problem does this feature solve? What inconvenience exists in the current workflow?

## ✨ Proposed Solution

Describe in detail how the feature should work.

**Expected Usage**:
```bash
# Command example
python -m moai_adk new-feature --option
```

## 🔄 Considered Alternatives

Have you considered other alternatives or solutions?

## 📚 Additional Information

Are there relevant documentation, references, or similar tools that provide this feature?
```

**Example**:

```markdown
## 💡 Feature Proposal

Auto-export SPEC documents to PDF

## 🎯 Problem This Solves

Currently, to share SPEC documents with external stakeholders, manual conversion from Markdown is required.
Non-developer stakeholders find it difficult to read Markdown format.

## ✨ Proposed Solution

Propose adding `python -m moai_adk export` command to export SPEC documents to PDF.

**Expected Usage**:
```bash
# Export specific SPEC to PDF
python -m moai_adk export SPEC-AUTH-001 --format pdf

# Export all SPECs to PDF
python -m moai_adk export --all --format pdf --output ./exports
```

## 🔄 Considered Alternatives

- Manual conversion using Pandoc
- Host documentation on GitHub Pages

## 📚 Additional Information

Reference: [Pandoc Markdown to PDF](https://pandoc.org/MANUAL.html#creating-a-pdf)
```

---

## Pull Request Guide

Before submitting a Pull Request, please verify the following:

### PR Submission Checklist

- [ ] **SPEC Written**: Is there a SPEC document for the changes? (`/moai:1-plan`)
- [ ] **TDD Completed**: Have you completed the RED-GREEN-REFACTOR cycle? (`/moai:2-run`)
- [ ] **Documentation Synchronized**: Has the Living Document been updated? (`/moai:3-sync`)
- [ ] **TRUST 5 Principles Followed**:
  - [ ] **T**est: Are tests written? (Coverage ≥85%)
  - [ ] **R**eadable: Is code readable? (Function ≤50 LOC, File ≤300 LOC)
  - [ ] **U**nified: Are consistent patterns used?
  - [ ] **S**ecured: Are there no security vulnerabilities?

### PR Template

MoAI-ADK uses an [automatic PR template](.github/PULL_REQUEST_TEMPLATE.md).
The `/moai:3-sync` command automatically fills in most of the information.

**Parts you need to manually complete**:
- Verify SPEC ID
- Summarize changes
- Document test scenarios

---

## Development Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 2. Install uv Package Manager (if needed)

**Windows Users (RECOMMENDED)**:
```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

**macOS/Linux Users**:
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**WSL Users (Windows Subsystem for Linux)**:
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Platform Notes**:
- 🟢 **Windows (PowerShell)**: Recommended for Windows users - most stable
- 🟡 **WSL**: Works but has environment setup overhead
- ✅ **macOS/Linux**: Use native bash installation

### 3. Install Dependencies

```bash
# Recommended: uv (fast installation)
uv pip install -e ".[dev]"

# Or use standard pip
pip install -e ".[dev]"
```

### 4. Use MoAI-ADK Locally

```bash
# Check CLI version
python -m moai_adk --version

# Check help
python -m moai_adk --help
```

### 5. Run in Development Mode

```bash
# Run tests
uv run pytest -n auto

# Run code quality checks
uv run ruff check
uv run mypy src
```

### 6. Understanding Alfred Configuration (Important!)

The core of MoAI-ADK is **Alfred** (MoAI SuperAgent). Alfred's behavior is defined in 4 documents in the `.claude/` directory:

#### 📄 Essential Reading: 4-Document Architecture

| Document | Size | When to Read | Key Content |
|----------|------|-------------|-------------|
| **CLAUDE.md** | ~7kb | Before starting development | Alfred's identity, core directives, 3-step workflow |
| **CLAUDE-AGENTS-GUIDE.md** | ~14kb | When you need a specific Agent | 19 Sub-agent team structure, 55 Skills classification |
| **CLAUDE-RULES.md** | ~17kb | When understanding decision rules | Skill invocation rules, user question rules, TRUST 5 gates |
| **CLAUDE-PRACTICES.md** | ~8kb | When you want real workflow examples | JIT context patterns, practical workflows |

#### 🎯 Key Developer Knowledge (Summary)

**Alfred's 3 Core Responsibilities**:
1. **SPEC-First**: Define requirements before code
2. **Automated TDD**: Execute RED → GREEN → REFACTOR cycle
3. **Automatic Document Sync**: Keep code and docs synchronized

**Understand the 4-Layer Architecture**:
- 📌 **Commands** (`/moai:0-3`): Workflow entry points
- 🤖 **Sub-agents** (19): Specialists for each phase
- 📚 **Skills** (55): Reusable knowledge base
- 🛡️ **Hooks**: Safety checks and validation

#### 💡 Tips

- Need to modify `.claude/` files? **Usually not**. Defaults are optimized.
- When proposing new features, refer to "Skill Invocation Rules" in **CLAUDE-RULES.md**.
- If Alfred's behavior seems off, check "Alfred's Core Directives" in **CLAUDE.md** first.

---

## Code Contribution Guide

### Follow MoAI-ADK 3-Step Workflow

MoAI-ADK follows the **SPEC-First TDD** methodology. All code changes must follow these steps:

#### Step 1: Plan & Write SPEC (`/moai:1-plan`)

```bash
/moai:1-plan "Feature description"
```

- Write requirements in EARS format
- Creates `.moai/specs/SPEC-{ID}/spec.md`
- Automatically creates feature branch

#### Step 2: Execute TDD (`/moai:2-run`)

```bash
/moai:2-run SPEC-{ID}
```

- **REFACTOR**: Improve code quality

#### Step 3: Synchronize Documentation (`/moai:3-sync`)

```bash
/moai:3-sync
```

- Update Living Document
- Convert to PR Ready state

### Code Style Guide

**TypeScript Code**:
- Function: ≤50 LOC
- File: ≤300 LOC
- Parameters: ≤5
- Complexity: ≤10

```typescript
export class AuthService {
  async login(username: string, password: string): Promise<Token> {
    // Implementation
  }
}
```

**Writing Tests**:
```typescript
describe('AuthService', () => {
  it('should authenticate valid credentials', async () => {
    // Given
    const authService = new AuthService();

    // When
    const token = await authService.login('user', 'pass');

    // Then
    expect(token).toBeDefined();
  });
});
```

---

## 💬 Have Questions?

- **General Questions**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **Bug Reports**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **Real-time Chat**: (Discord link coming soon)

---

## 🙏 Code of Conduct

The MoAI-ADK project is committed to creating an environment that is open and welcoming to all.

**What We Encourage**:
- ✅ Respectful and considerate behavior
- ✅ Constructive feedback
- ✅ Collaborative problem-solving
- ✅ Diversity and inclusion

**What Is Not Acceptable**:
- ❌ Offensive or aggressive language
- ❌ Harassment or discrimination
- ❌ Disrespectful or unprofessional behavior

---

**Thank you for all contributions!** 🪿

Let's make MoAI-ADK a better tool together!

---

---

# 한국어 버전

## 🤝 MoAI-ADK에 기여하기

MoAI-ADK 프로젝트에 기여해주셔서 감사합니다! 이 문서는 프로젝트에 효과적으로 기여하는 방법을 안내합니다.

---

## 📋 목차

- [이슈 작성 가이드](#이슈-작성-가이드)
  - [버그 리포트](#버그-리포트)
  - [기능 제안](#기능-제안)
- [Pull Request 가이드](#pull-request-가이드)
- [개발 환경 설정](#개발-환경-설정)
- [코드 기여 가이드](#코드-기여-가이드)

---

## 이슈 작성 가이드

### 버그 리포트

버그를 발견하셨나요? 다음 정보를 포함하여 이슈를 작성해주세요:

**제목 형식**: `[Bug] 간단한 버그 설명`

**필수 포함 사항**:

```markdown
## 🐛 버그 설명

버그에 대한 명확하고 간결한 설명을 작성해주세요.

## 🔄 재현 단계

1. 어떤 명령어를 실행했는지
2. 어떤 입력을 제공했는지
3. 어떤 동작을 수행했는지
4. 오류가 발생한 시점

## 💥 예상 동작 vs 실제 동작

- **예상 동작**: 어떻게 작동해야 하는지
- **실제 동작**: 실제로 어떻게 작동했는지

## 🖥️ 환경 정보

- **OS**: (예: macOS 14.0, Ubuntu 22.04, Windows 11)
- **Python 버전**: (예: 3.11.0)
- **MoAI-ADK 버전**: (예: v0.14.0)
- **Claude Code 버전**: (선택사항)

## 📸 스크린샷 또는 로그

가능하면 에러 메시지, 스크린샷, 또는 로그를 첨부해주세요.

```bash
# 에러 로그 예시
Error: Cannot find module '...'
    at Function.Module._resolveFilename ...
```

## 🔍 추가 정보

버그와 관련된 추가 정보나 컨텍스트를 제공해주세요.
```

**예시**:

```markdown
## 🐛 버그 설명

`/moai:2-run` 명령 실행 시 구현 검증 단계에서 오류가 발생합니다.

## 🔄 재현 단계

1. `python -m moai_adk init .` 명령으로 프로젝트 초기화
2. `/moai:1-plan "사용자 인증"` 실행하여 Plan & SPEC 생성
3. `/moai:2-run SPEC-AUTH-001` 실행
4. 구현 검증 단계에서 오류 발생

## 💥 예상 동작 vs 실제 동작

- **예상 동작**: TDD 사이클이 정상적으로 완료되어야 함

## 🖥️ 환경 정보

- **OS**: macOS 14.2
- **Python 버전**: 3.11.0
- **MoAI-ADK 버전**: v0.14.0

## 📸 스크린샷 또는 로그

```bash
Error: Implementation validation failed
Please ensure all tests are passing before proceeding
```
```

---

### 기능 제안

새로운 기능을 제안하고 싶으신가요?

**제목 형식**: `[Feature Request] 기능 이름`

**필수 포함 사항**:

```markdown
## 💡 기능 제안

제안하는 기능에 대한 명확하고 간결한 설명을 작성해주세요.

## 🎯 해결하려는 문제

이 기능이 어떤 문제를 해결하나요? 현재 워크플로우에서 어떤 불편함이 있나요?

## ✨ 제안하는 해결 방법

기능이 어떻게 작동해야 하는지 구체적으로 설명해주세요.

**예상 사용 방법**:
```bash
# 명령어 예시
python -m moai_adk new-feature --option
```

## 🔄 대안 고려

다른 대안이나 해결 방법을 고려해보셨나요?

## 📚 추가 정보

관련 문서, 레퍼런스, 또는 유사한 기능을 제공하는 도구가 있나요?
```

**예시**:

```markdown
## 💡 기능 제안

SPEC 문서를 자동으로 PDF로 내보내는 기능

## 🎯 해결하려는 문제

현재 SPEC 문서를 외부 이해관계자와 공유하려면 Markdown을 수동으로 변환해야 합니다.
비개발자 이해관계자는 Markdown 형식을 읽기 어려워합니다.

## ✨ 제안하는 해결 방법

`python -m moai_adk export` 명령어로 SPEC 문서를 PDF로 내보낼 수 있도록 제안합니다.

**예상 사용 방법**:
```bash
# 특정 SPEC을 PDF로 내보내기
python -m moai_adk export SPEC-AUTH-001 --format pdf

# 모든 SPEC을 PDF로 내보내기
python -m moai_adk export --all --format pdf --output ./exports
```

## 🔄 대안 고려

- Pandoc을 사용한 수동 변환
- GitHub Pages로 웹 문서 호스팅

## 📚 추가 정보

참고: [Pandoc Markdown to PDF](https://pandoc.org/MANUAL.html#creating-a-pdf)
```

---

## Pull Request 가이드

Pull Request를 제출하기 전에 다음 사항을 확인해주세요:

### PR 제출 체크리스트

- [ ] **SPEC 작성**: 변경 사항에 대한 SPEC 문서가 있습니까? (`/moai:1-plan`)
- [ ] **TDD 완료**: RED-GREEN-REFACTOR 사이클을 완료했습니까? (`/moai:2-run`)
- [ ] **문서 동기화**: Living Document가 업데이트되었습니까? (`/moai:3-sync`)
- [ ] **TRUST 5원칙 준수**:
  - [ ] **T**est: 테스트가 작성되었습니까? (커버리지 ≥85%)
  - [ ] **R**eadable: 코드가 읽기 쉽습니까? (함수 ≤50 LOC, 파일 ≤300 LOC)
  - [ ] **U**nified: 일관된 패턴을 사용했습니까?
  - [ ] **S**ecured: 보안 취약점이 없습니까?

### PR 템플릿

MoAI-ADK는 [자동 PR 템플릿](.github/PULL_REQUEST_TEMPLATE.md)을 사용합니다.
`/moai:3-sync` 명령이 대부분의 정보를 자동으로 채워줍니다.

**수동으로 작성해야 할 부분**:
- SPEC ID 확인
- 변경 사항 요약
- 테스트 시나리오

---

## 개발 환경 설정

### 1. 저장소 클론

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 2. uv 패키지 관리자 설치 (필요할 경우)

**Windows 사용자 (권장)**:
```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

**macOS/Linux 사용자**:
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**WSL 사용자 (Windows Subsystem for Linux)**:
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**플랫폼별 주의사항**:
- 🟢 **Windows (PowerShell)**: Windows 사용자를 위한 권장 방법 - 가장 안정적
- 🟡 **WSL**: 작동하지만 환경 설정 오버헤드 발생
- ✅ **macOS/Linux**: 기본 bash 설치 사용

### 3. 의존성 설치

```bash
# uv 권장 (빠른 설치)
uv pip install -e ".[dev]"

# 또는 표준 pip 사용
pip install -e ".[dev]"
```

### 4. 로컬에서 MoAI-ADK 사용

```bash
# CLI 버전 확인
python -m moai_adk --version

# 도움말 확인
python -m moai_adk --help
```

### 5. 개발 모드 실행

```bash
# 테스트 실행
uv run pytest -n auto

# 코드 품질 검사
uv run ruff check
uv run mypy src
```

### 6. Alfred의 설정 문서 이해하기 (중요!)

MoAI-ADK의 핵심은 **Alfred** (MoAI SuperAgent)입니다. Alfred의 동작 방식은 `.claude/` 디렉토리의 4개 문서로 정의됩니다:

#### 📄 필수 읽기: 4-Document Architecture

| 문서 | 크기 | 언제 읽을까? | 주요 내용 |
|------|------|-----------|----------|
| **CLAUDE.md** | ~7kb | 개발 시작 시 | Alfred의 정체성, 핵심 지령, 3단계 워크플로우 |
| **CLAUDE-AGENTS-GUIDE.md** | ~14kb | 어떤 Agent가 필요할 때 | 19개 Sub-agent 팀 구조, 55개 Skills 분류 |
| **CLAUDE-RULES.md** | ~17kb | 의사결정 규칙을 이해하고 싶을 때 | Skill 호출 규칙, 사용자 질문 규칙, TRUST 5 게이트 |
| **CLAUDE-PRACTICES.md** | ~8kb | 실제 워크플로우 예제를 원할 때 | JIT 컨텍스트 패턴, 실전 워크플로우 |

#### 🎯 개발자가 알아야 할 것 (요약)

**Alfred의 3가지 핵심 의무**:
1. **SPEC-First**: 코드 전에 요구사항 정의
2. **TDD 자동 실행**: RED → GREEN → REFACTOR 순환
3. **문서 자동 동기화**: 코드와 문서 항상 일치

**4개 계층 구조를 이해하세요**:
- 📌 **Commands** (`/moai:0-3`): 워크플로우 진입점
- 🤖 **Sub-agents** (19명): 각 단계별 전문가
- 📚 **Skills** (55개): 재사용 가능한 지식 기지
- 🛡️ **Hooks**: 안전장치 및 검증

#### 💡 팁

- `.claude/` 파일을 수정해야 하나? **대부분 안 합니다**. 기본값이 최적화되어 있습니다.
- 새 기능을 제안할 때는 **CLAUDE-RULES.md**의 "Skill Invocation Rules" 섹션을 참고하세요.
- Alfred의 동작이 이상하면 **CLAUDE.md**의 "Alfred's Core Directives"를 먼저 확인하세요.

---

## 코드 기여 가이드

### MoAI-ADK 3단계 워크플로우 따르기

MoAI-ADK는 **SPEC-First TDD** 방법론을 따릅니다. 모든 코드 변경은 다음 단계를 거쳐야 합니다:

#### 1단계: Plan & SPEC 작성 (`/moai:1-plan`)

```bash
/moai:1-plan "기여하려는 기능 설명"
```

- EARS 방식으로 요구사항 작성
- `.moai/specs/SPEC-{ID}/spec.md` 생성
- feature 브랜치 자동 생성

#### 2단계: TDD 실행 (`/moai:2-run`)

```bash
/moai:2-run SPEC-{ID}
```

- **REFACTOR**: 코드 품질 개선

#### 3단계: 문서 동기화 (`/moai:3-sync`)

```bash
/moai:3-sync
```

- Living Document 업데이트
- PR Ready 전환

### 코드 스타일 가이드

**TypeScript 코드**:
- 함수: ≤50 LOC
- 파일: ≤300 LOC
- 매개변수: ≤5개
- 복잡도: ≤10

```typescript
export class AuthService {
  async login(username: string, password: string): Promise<Token> {
    // 구현
  }
}
```

**테스트 작성**:
```typescript
describe('AuthService', () => {
  it('should authenticate valid credentials', async () => {
    // Given
    const authService = new AuthService();

    // When
    const token = await authService.login('user', 'pass');

    // Then
    expect(token).toBeDefined();
  });
});
```

---

## 💬 질문이 있으신가요?

- **일반 질문**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **버그 리포트**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **실시간 대화**: (Discord 링크 추가 예정)

---

## 🙏 기여자 행동 강령

MoAI-ADK 프로젝트는 모두에게 열려 있고 환영받는 환경을 만들기 위해 노력합니다.

**우리가 지향하는 것**:
- ✅ 존중하고 배려하는 태도
- ✅ 건설적인 피드백
- ✅ 협력적인 문제 해결
- ✅ 다양성과 포용성

**허용되지 않는 것**:
- ❌ 모욕적이거나 공격적인 언어
- ❌ 괴롭힘이나 차별
- ❌ 무례하거나 비전문적인 행동

---

**모든 기여에 감사드립니다!** 🪿

MoAI-ADK를 함께 더 나은 도구로 만들어가요!
