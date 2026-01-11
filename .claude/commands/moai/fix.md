---
description: "Agentic auto-fix - Parallel scan with autonomous correction"
argument-hint: "[--dry] [--parallel] [--level N] [file_path]"
type: utility
allowed-tools: Task, AskUserQuestion, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git diff --name-only HEAD

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:fix - Agentic Auto-Fix

## Core Principle: 완전 자율 수정

AI가 스스로 이슈를 찾고 수정합니다.

```
START: 이슈 감지
  ↓
AI: 병렬 스캔 → 분류 → 수정 → 검증
  ↓
AI: 완료 마커 추가
```

## Command Purpose

LSP 오류, linting 이슈를 자율적으로 수정:

1. **병렬 스캔** (LSP + AST-grep + Linters 동시)
2. **자율 분류** (Level 1-4)
3. **자동 수정** (Level 1-2)
4. **검증**
5. **보고**

Target: $ARGUMENTS

## Quick Start

```bash
# 기본 수정
/moai:fix

# 병렬 스캔
/moai:fix --parallel

# 미리보기
/moai:fix --dry

# 특정 파일
/moai:fix src/auth.py

# 수정 레벨 제한
/moai:fix --level 2
```

## Command Options

| 옵션 | 축약 | 설명 | 기본값 |
|------|------|------|--------|
| `--dry` | --dry-run | 미리보기만 | 적용 |
| `--parallel` | - | 병렬 스캔 | 권장 |
| `--level N` | - | 최대 수정 레벨 | 3 |
| `--errors` | --errors-only | 에러만 수정 | 전체 |
| `--security` | --include-security | 보안 포함 | 제외 |
| `--no-fmt` | --no-format | 포맷팅 스킵 | 포함 |

## Parallel Scan (병렬 스캔)

```bash
# 순차 (30초)
LSP → AST → Linter

# 병렬 (8초)
LSP   ├─┐
      ├─→ 통합 (3.75배)
AST   ├─┤
     ├─┘
Linter
```

## Auto-Fix Levels

| Level | 설명 | 승인 | 예시 |
|-------|------|------|------|
| 1 | 즉시 | 불필요 | import, 공백 |
| 2 | 안전 | 로그 | 변수改名, 타입 |
| 3 | 승인 | 필요 | 로직, API |
| 4 | 수동 | 불가능 | 보안, 아키텍처 |

## TODO-Obsessive Rule

```
[HARD] TODO 관리 규칙

1. 즉시 생성: 이슈 → TODO
2. 즉시 진행: [ ] → [in_progress]
3. 즉시 완료: [in_progress] → [x]
4. 증거: 모든 수정에 출처
```

## Output Format

### 미리보기

```markdown
## Fix: Dry Run

### Scan (0.8s, parallel)
- LSP: 12 issues
- AST-grep: 0 security
- Linter: 5 issues

### Level 1 (12건)
- src/auth.py: import, formatting
- src/api/routes.py: import order
- tests/test_auth.py: whitespace

### Level 2 (3건)
- src/auth.py:45 - 'usr' → 'user'
- src/api/routes.py:78 - type 추가
- src/models.py:23 - dataclass?

### Level 4 (2건)
- src/auth.py:67 - logic error
- src/api/routes.py:112 - SQL injection

No changes (--dry).
```

### 완료

```markdown
## Fix: Complete

### Applied
- Level 1: 12 issues
- Level 2: 3 issues
- Level 3: 0 issues

### Evidence
**src/auth.py:5** - Removed unused `os`, `sys`
**src/auth.py:23** - Fixed whitespace
**src/api/routes.py:12** - Sorted imports

### Remaining (Level 4)
1. src/auth.py:67 - logic error
2. src/api/routes.py:112 - SQL injection

### Next
/moai:loop --parallel  # 루프로 계속
```

## Quick Reference

```bash
# 수정
/moai:fix

# 병렬
/moai:fix --parallel

# 미리보기
/moai:fix --dry

# 에러만
/moai:fix --errors

# 특정 파일
/moai:fix src/auth.py
```

---

## EXECUTION DIRECTIVE

1. $ARGUMENTS 파싱
2. 병렬 스캔 (--parallel)
3. 분류 (Level 1-4)
4. TODO 생성
5. Level 1-2 수정
6. Level 3 승인 요청
7. 검증
8. 보고 (증거 포함)

---

Version: 2.1.0
Last Updated: 2026-01-11
Core: Agentic AI Auto-Fix
