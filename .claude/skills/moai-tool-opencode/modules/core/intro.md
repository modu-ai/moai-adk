---
source: https://opencode.ai/docs/
fetched: 2026-01-08
title: Introduction to OpenCode
---

# Introduction to OpenCode

## Overview

OpenCode is an open source AI coding agent available in multiple formats: terminal interface, desktop application, and IDE extension.

## Installation Methods

### Quick Install

```bash
curl -fsSL https://opencode.ai/install | bash
```

### Package Managers

**Node.js (npm, Bun, pnpm, Yarn)**

```bash
npm install -g opencode
# or
bun install -g opencode
# or
pnpm install -g opencode
# or
yarn global add opencode
```

**Homebrew (macOS/Linux)**

```bash
brew install opencode
```

**Paru (Arch Linux)**

```bash
paru -S opencode
```

**Chocolatey (Windows)**

```bash
choco install opencode
```

**Scoop (Windows)**

```bash
scoop install opencode
```

**Mise**

```bash
mise use -g opencode
```

**Docker**

Docker support is available for containerized environments.

## Core Setup Steps

### 1. Prerequisites

- Modern terminal emulator (WezTerm, Alacritty, Ghostty, Kitty)
- LLM API keys

### 2. Configuration

Users connect via `/connect` command, authenticate at opencode.ai/auth, and input API credentials.

### 3. Project Initialization

Navigate to project directory and run `opencode`, then execute `/init` to analyze the codebase and generate `AGENTS.md`.

## Primary Use Cases

The documentation highlights four main workflows:

### Question Asking

Query codebase understanding using `@` syntax for file references.

```
@src/main.ts What does this file do?
```

### Feature Planning

Toggle Plan mode (Tab key) to review implementation approaches before building.

### Direct Changes

Request specific modifications with detailed context.

### Version Control

Undo/redo changes via `/undo` and `/redo` commands.

## Additional Features

### Sharing

`/share` command creates shareable conversation links.

### Customization

- Theme selection
- Keybind configuration
- Code formatter setup
- Custom commands

### Documentation

References to config, providers, and troubleshooting resources.

## Best Practices

The page emphasizes providing detailed context and treating the AI agent like a junior developer for optimal results.
