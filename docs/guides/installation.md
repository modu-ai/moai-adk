# Installation

MoAI-ADK 설치 가이드입니다. 다양한 환경과 사용 사례에 맞는 설치 방법을 제공합니다.

## System Requirements

### Minimum Requirements

| Component | Requirement |
|-----------|------------|
| **OS** | macOS, Linux, Windows (WSL2) |
| **Node.js** | 18.0.0 or higher |
| **Package Manager** | npm, pnpm, yarn, or bun |
| **Disk Space** | ~100 MB |
| **Memory** | 2 GB RAM (recommended 4 GB) |

### Recommended Setup

- **Node.js**: v20.x LTS
- **Package Manager**: Bun 1.2.0+ (fastest)
- **Editor**: VSCode with Claude Code extension
- **Terminal**: iTerm2 (macOS) or Windows Terminal

---

## Installation Methods

### Method 1: Global Installation (권장)

전역 설치는 어디서든 `moai` 명령어를 사용할 수 있습니다.

::: code-group

```bash [bun (fastest)]
# Bun으로 전역 설치 (권장)
bun add -g moai-adk

# 확인
moai --version
```

```bash [npm]
# npm으로 전역 설치
npm install -g moai-adk

# 확인
moai --version
```

```bash [pnpm]
# pnpm으로 전역 설치
pnpm add -g moai-adk

# 확인
moai --version
```

```bash [yarn]
# Yarn으로 전역 설치
yarn global add moai-adk

# 확인
moai --version
```

:::

### Method 2: Local Installation (프로젝트별)

특정 프로젝트에만 설치하려는 경우:

::: code-group

```bash [bun]
cd your-project
bun add -D moai-adk

# npx 대신 bunx 사용
bunx moai --version
```

```bash [npm]
cd your-project
npm install --save-dev moai-adk

# npx로 실행
npx moai --version
```

```bash [pnpm]
cd your-project
pnpm add -D moai-adk

# pnpm exec로 실행
pnpm moai --version
```

```bash [yarn]
cd your-project
yarn add -D moai-adk

# yarn으로 실행
yarn moai --version
```

:::

### Method 3: From Source (개발자용)

최신 개발 버전을 사용하거나 기여하려는 경우:

```bash
# 레포지토리 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk/moai-adk-ts

# 의존성 설치
bun install

# 빌드
bun run build

# 로컬에서 링크
bun link

# 다른 프로젝트에서 사용
cd your-project
bun link moai-adk
```

---

## Post-Installation Setup

### 1. Verify Installation

설치가 완료되면 다음 명령어로 확인:

```bash
moai --version
# Expected output: 0.2.17

moai help
# Shows available commands
```

### 2. System Diagnostics

시스템 환경을 확인합니다:

```bash
moai doctor
```

**출력 예시**:

```
🔍 MoAI-ADK System Diagnostics

Environment
✅ Operating System: darwin (macOS)
✅ Node.js: v20.10.0
✅ Package Manager: bun v1.2.19
✅ Git: v2.39.0

Claude Code Integration
✅ Claude Code: Available
✅ .claude directory: Ready

Project Status
⚠️  No MoAI project detected
→  Run 'moai init .' to initialize

All checks passed! 🚀
```

### 3. Optional Dependencies

추가 기능을 위한 선택적 의존성:

#### GitHub CLI (PR 자동화)

```bash
# macOS
brew install gh

# Ubuntu/Debian
sudo apt install gh

# Windows
winget install GitHub.cli
```

인증:

```bash
gh auth login
```

#### Claude Code Extension

VSCode에서 Claude Code 확장 설치:

1. VSCode 열기
2. Extensions (⌘+Shift+X)
3. "Claude Code" 검색
4. Install

---

## Platform-Specific Instructions

### macOS

#### Homebrew로 Node.js 설치

```bash
# Homebrew 설치 (없는 경우)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Node.js 설치
brew install node

# Bun 설치 (권장)
curl -fsSL https://bun.sh/install | bash
```

#### PATH 설정

```bash
# ~/.zshrc 또는 ~/.bash_profile에 추가
export PATH="$HOME/.bun/bin:$PATH"

# 적용
source ~/.zshrc
```

### Linux (Ubuntu/Debian)

```bash
# Node.js 설치 (NodeSource)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Bun 설치
curl -fsSL https://bun.sh/install | bash

# PATH 설정
echo 'export PATH="$HOME/.bun/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Windows (WSL2)

```bash
# WSL2 설치 (PowerShell 관리자 권한)
wsl --install

# Ubuntu 재시작 후
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Bun 설치
curl -fsSL https://bun.sh/install | bash
```

---

## Updating MoAI-ADK

### CLI를 통한 업데이트 (v0.2.18+)

```bash
# Claude Code에서 실행
/alfred:9-update

# 또는 터미널에서
moai update
```

### 수동 업데이트

::: code-group

```bash [bun]
bun update -g moai-adk
```

```bash [npm]
npm update -g moai-adk
```

```bash [pnpm]
pnpm update -g moai-adk
```

```bash [yarn]
yarn global upgrade moai-adk
```

:::

### 특정 버전 설치

```bash
# Specific version
bun add -g moai-adk@0.2.17

# Latest beta
bun add -g moai-adk@beta

# Latest canary
bun add -g moai-adk@canary
```

---

## Uninstallation

### Global Uninstall

::: code-group

```bash [bun]
bun remove -g moai-adk
```

```bash [npm]
npm uninstall -g moai-adk
```

```bash [pnpm]
pnpm remove -g moai-adk
```

```bash [yarn]
yarn global remove moai-adk
```

:::

### Remove Project Files

```bash
# .moai 디렉토리 제거
rm -rf .moai

# .claude 디렉토리 제거 (선택)
rm -rf .claude

# CLAUDE.md 제거 (선택)
rm CLAUDE.md
```

---

## Troubleshooting

### Common Issues

#### Issue 1: Command not found

```bash
# PATH 확인
echo $PATH

# Global bin 디렉토리 확인
bun pm bin -g  # Bun
npm config get prefix  # npm
```

**해결책**:

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
export PATH="$HOME/.bun/bin:$PATH"
export PATH="$HOME/.npm-global/bin:$PATH"
```

#### Issue 2: Permission denied

```bash
# EACCES 오류 시
sudo chown -R $(whoami) ~/.npm
sudo chown -R $(whoami) ~/.bun
```

#### Issue 3: Version mismatch

```bash
# 캐시 정리
bun pm cache rm
npm cache clean --force

# 재설치
bun add -g moai-adk
```

### Get Help

```bash
# 시스템 진단
moai doctor --verbose

# 백업 목록 확인
moai doctor --list-backups

# 복원
moai restore .moai-backup/2025-10-11T13-00-00
```

---

## Next Steps

설치가 완료되었습니다! 이제 다음 단계로 진행하세요:

1. **[Getting Started](/guides/getting-started)** - 첫 프로젝트 시작
2. **[Quick Start](/guides/quick-start)** - 5분 튜토리얼
3. **[SPEC-First TDD](/guides/concepts/spec-first-tdd)** - 핵심 개념

---

<div style="text-align: center; margin-top: 40px;">
  <p>Installation complete! Ready to build with MoAI-ADK 🚀</p>
</div>
