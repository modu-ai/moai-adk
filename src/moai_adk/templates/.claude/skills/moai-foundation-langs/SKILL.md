---
name: moai-foundation-langs
version: 4.0.0
updated: 2025-11-20
status: stable
tier: foundation
description: Programming language detection and setup patterns
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Language & Stack Detection

**Automatic Language Recognition & Setup**

> **Detection**: File extensions, config files, package managers  
> **Support**: 25+ languages

---

## Overview

Automatically detect project language and recommend appropriate tooling.

### Detection Methods

1.  **File Extensions**: `.py`, `.js`, `.go`, `.rs`
2.  **Config Files**: `package.json`, `Cargo.toml`, `go.mod`
3.  **Lock Files**: `package-lock.json`, `Cargo.lock`, `go.sum`

---

## Language Patterns

### Python

**Detection**:

- Files: `*.py`, `requirements.txt`, `pyproject.toml`
- Tools: `pip`, `uv`, `poetry`

**Setup**:

```bash
# Modern (uv)
uv init
uv add fastapi pytest

# Traditional
python -m venv .venv
pip install -r requirements.txt
```

### JavaScript/TypeScript

**Detection**:

- Files: `*.js`, `*.ts`, `package.json`, `tsconfig.json`
- Tools: `npm`, `yarn`, `pnpm`

**Setup**:

```bash
npm install
npm run build
```

### Go

**Detection**:

- Files: `*.go`, `go.mod`, `go.sum`

**Setup**:

```bash
go mod download
go build ./...
```

### Rust

**Detection**:

- Files: `*.rs`, `Cargo.toml`, `Cargo.lock`

**Setup**:

```bash
cargo build
cargo test
```

---

## Multi-Language Projects

**Monorepo Detection**:

```
project/
├── backend/     # Python
│   └── pyproject.toml
├── frontend/    # TypeScript
│   └── package.json
└── services/    # Go
    └── go.mod
```

**Recommendation**: Use workspace tools (Turborepo, Nx, Lerna)

---

## Validation Checklist

- [ ] **Detection**: Language correctly identified?
- [ ] **Version**: Specific version recommended?
- [ ] **Tools**: Package manager specified?
- [ ] **Setup**: Installation steps provided?

---

## Related Skills

- `moai-domain-backend`: Backend development
- `moai-domain-frontend`: Frontend development

---

**Last Updated**: 2025-11-20
