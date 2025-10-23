# MoAI-ADK (Agentic Development Kit)

[English](README.md) | [한국어](README.ko.md) | [ไทย](README.th.md) | [日本語](README.ja.md) | [中文](README.zh.md) | [हिन्दी](README.hi.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-87.84%25-brightgreen)](https://github.com/modu-ai/moai-adk)

> **MoAI-ADK delivers a seamless development workflow that naturally connects SPEC → TEST (TDD) → CODE → DOCUMENTATION with AI.**

---

## 1. MoAI-ADK at a Glance

MoAI-ADK transforms AI-powered development with three core principles. Use the navigation below to jump to the section that matches your needs.

If you're **new to MoAI-ADK**, start with "What is MoAI-ADK?".
If you want to **get started quickly**, jump straight to "5-Minute Quick Start".
If you've **already installed it and want to understand the concepts**, we recommend "5 Key Concepts".

| Question                           | Jump To                                                      |
| ---------------------------------- | ------------------------------------------------------------ |
| First time here—what is it?        | [What is MoAI-ADK?](#what-is-moai-adk)                       |
| How do I get started?              | [5-Minute Quick Start](#5-minute-quick-start)                |
| What's the basic flow?             | [Core Workflow (0 → 3)](#core-workflow-0--3)                 |
| What do Plan/Run/Sync commands do? | [Command Cheat Sheet](#command-cheat-sheet)                  |
| What are SPEC, TDD, TAG?           | [5 Key Concepts](#5-key-concepts)                            |
| Tell me about agents/Skills        | [Sub-agents & Skills Overview](#sub-agents--skills-overview) |
| I want a 4-week hands-on project   | [Second Practice: Mini Kanban Board](#second-practice-mini-kanban-board) |
| Want to dive deeper?               | [Additional Resources](#additional-resources)                |

---

## What is MoAI-ADK?

### The Problem: Trust Crisis in AI Development

Today, countless developers want help from Claude or ChatGPT, but can't shake one fundamental doubt: **"Can I really trust the code this AI generates?"**

The reality looks like this. Ask an AI to "build a login feature" and you'll get syntactically perfect code. But these problems keep repeating:

- **Unclear Requirements**: The basic question "What exactly should we build?" remains unanswered. Email/password login? OAuth? 2FA? Everything relies on guessing.
- **Missing Tests**: Most AIs only test the "happy path". Wrong password? Network error? Three months later, bugs explode in production.
- **Documentation Drift**: Code gets modified but docs stay the same. The question "Why is this code here?" keeps repeating.
- **Context Loss**: Even within the same project, you have to explain everything from scratch each time. Project structure, decision rationale, previous attempts—nothing gets recorded.
- **Impact Tracking Impossible**: When requirements change, you can't track which code is affected.

### The Solution: SPEC-First TDD with Alfred SuperAgent

**MoAI-ADK** (MoAI Agentic Development Kit) is an open-source framework designed to **systematically solve** these problems.

The core principle is simple yet powerful:

> **"No tests without code, no SPEC without tests"**

More precisely, it's the reverse order:

> **"SPEC comes first. No tests without SPEC. No complete documentation without tests and code."**

When you follow this order, magical things happen:

**1️⃣ Clear Requirements**
Write SPECs first with the `/alfred:1-plan` command. A vague request like "login feature" transforms into **clear requirements** like "WHEN valid credentials are provided, the system SHALL issue a JWT token". Alfred's spec-builder uses EARS syntax to create professional SPECs in just 3 minutes.

**2️⃣ Test Guarantee**
`/alfred:2-run` automatically performs Test-Driven Development (TDD). It proceeds in RED (failing test) → GREEN (minimal implementation) → REFACTOR (cleanup) order, **guaranteeing 85%+ test coverage**. No more "testing later". Tests drive code creation.

**3️⃣ Automatic Documentation Sync**
A single `/alfred:3-sync` command **synchronizes** all code, tests, and documentation. README, CHANGELOG, API docs, and Living Documents all update automatically. Six months later, code and docs still match.

**4️⃣ Tracking with @TAG System**
Every piece of code, test, and documentation gets a `@TAG:ID`. When requirements change later, one command—`rg "@SPEC:AUTH-001"`—**finds all related tests, implementations, and docs**. You gain confidence during refactoring.

**5️⃣ Alfred Remembers Context**
A team of AI agents collaborate to **remember** your project's structure, decision rationale, and work history. No need to repeat the same questions.

### MoAI-ADK's 3 Core Promises

For beginners to remember easily, MoAI-ADK's value simplifies to three things:

**First, SPEC comes before code**
Start by clearly defining what to build. Writing SPEC helps discover problems before implementation. Communication costs with teammates drop dramatically.

**Second, tests drive code (TDD)**
Write tests before implementation (RED). Implement minimally to pass tests (GREEN). Then clean up the code (REFACTOR). Result: fewer bugs, confidence in refactoring, code anyone can understand.

**Third, documentation and code always match**
One `/alfred:3-sync` command auto-updates all documentation. README, CHANGELOG, API docs, and Living Documents always sync with code. No more despair when modifying six-month-old code.

---

## Why Do You Need It?

### Real Challenges in AI Development

Modern AI-powered development faces various challenges. MoAI-ADK **systematically solves** all these problems:

| Concern                         | Traditional Approach Problem                       | MoAI-ADK Solution                                             |
| ------------------------------- | -------------------------------------------------- | ------------------------------------------------------------- |
| "Can't trust AI code"           | Implementation without tests, unclear verification | Enforces SPEC → TEST → CODE order, guarantees 85%+ coverage   |
| "Repeating same explanations"   | Context loss, unrecorded project history           | Alfred remembers everything, 19 AI team members collaborate   |
| "Hard to write prompts"         | Don't know how to write good prompts               | `/alfred` commands provide standardized prompts automatically |
| "Documentation always outdated" | Forget to update docs after code changes           | `/alfred:3-sync` auto-syncs with one command                  |
| "Don't know what changed where" | Hard to search code, unclear intent                | @TAG chain connects SPEC → TEST → CODE → DOC                  |
| "Team onboarding takes forever" | New members can't grasp code context               | Reading SPEC makes intent immediately clear                   |

### Benefits You Can Experience Right Now

From the moment you adopt MoAI-ADK, you'll feel:

- **Faster Development**: Clear SPEC reduces round-trip explanation time
- **Fewer Bugs**: SPEC-based tests catch issues early
- **Better Code Understanding**: @TAG and SPEC make intent immediately clear
- **Lower Maintenance Costs**: Code and docs always match
- **Efficient Team Collaboration**: Clear communication through SPEC and TAG

---

## 5-Minute Quick Start

Now let's start your first project with MoAI-ADK. Follow these 5 steps and in just **5 minutes** you'll have a project with SPEC, TDD, and documentation all connected.

### Step 1: Install uv (about 30 seconds)

First, install `uv`. `uv` is an ultra-fast Python package manager written in Rust. It's **10+ times faster** than traditional `pip` and works perfectly with MoAI-ADK.

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows (PowerShell)
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Verify installation
uv --version
# Output: uv 0.x.x
```

**Why uv?** MoAI-ADK is optimized to leverage uv's fast installation speed and stability. Perfect project isolation means no impact on other Python environments.

### Step 2: Install MoAI-ADK (about 1 minute)

Install MoAI-ADK as a global tool. This won't affect your project dependencies.

```bash
# Install in tool mode (recommended: runs in isolated environment)
uv tool install moai-adk

# Verify installation
moai-adk --version
# Output: MoAI-ADK v0.4.11
```

Once installed, you can use the `moai-adk` command anywhere.

### Step 3: Create Project (about 1 minute)

**To start a new project:**
```bash
moai-adk init my-project
cd my-project
```

**To add to an existing project:**
```bash
cd your-existing-project
moai-adk init .
```

This one command automatically generates:

```
my-project/
├── .moai/                   # MoAI-ADK project configuration
│   ├── config.json
│   ├── project/             # Project information
│   ├── specs/               # SPEC files
│   └── reports/             # Analysis reports
├── .claude/                 # Claude Code automation
│   ├── agents/              # AI team
│   ├── commands/            # /alfred commands
│   ├── skills/              # Claude Skills
│   └── settings.json
├── src/                     # Implementation code
├── tests/                   # Test code
├── docs/                    # Auto-generated documentation
└── README.md
```

### Step 4: Start Alfred in Claude Code (about 2 minutes)

Run Claude Code and invoke the Alfred SuperAgent:

```bash
# Run Claude Code
claude
```

Then enter this in Claude Code's command input:

```
/alfred:0-project
```

This command performs:

1. **Collect Project Info**: "Project name?", "Goals?", "Main language?"
2. **Auto-detect Tech Stack**: Automatically recognizes Python/JavaScript/Go, etc.
3. **Deploy Skill Packs**: Prepares necessary Skills for your project
4. **Generate Initial Report**: Project structure, suggested next steps

### Step 5: Write First SPEC (about 1 minute)

After project initialization completes, write your first feature as a SPEC:

```
/alfred:1-plan "User registration feature"
```

Automatically generated:
- `@SPEC:USER-001` - Unique ID assigned
- `.moai/specs/SPEC-USER-001/spec.md` - Professional SPEC in EARS format
- `feature/spec-user-001` - Git branch auto-created

### Step 6: TDD Implementation (about 3 minutes)

Once SPEC is written, implement using TDD:

```
/alfred:2-run USER-001
```

This command handles:
- 🔴 **RED**: Automatically write failing test (`@TEST:USER-001`)
- 🟢 **GREEN**: Minimal implementation to pass test (`@CODE:USER-001`)
- ♻️ **REFACTOR**: Improve code quality

### Step 7: Documentation Sync (about 1 minute)

Finally, auto-sync all documentation:

```
/alfred:3-sync
```

Automatically generated/updated:
- Living Document (API documentation)
- README updates
- CHANGELOG generation
- @TAG chain validation

### Complete!

After these 7 steps, everything is ready:

✅ Requirements specification (SPEC)
✅ Test code (85%+ coverage)
✅ Implementation code (tracked with @TAG)
✅ API documentation (auto-generated)
✅ Change history (CHANGELOG)
✅ Git commit history (RED/GREEN/REFACTOR)

**Everything completes in 15 minutes!**

### Verify Generated Results

Check if the generated results were properly created:

```bash
# 1. Check TAG chain (SPEC → TEST → CODE → DOC)
rg '@(SPEC|TEST|CODE):USER-001' -n

# 2. Run tests
pytest tests/ -v

# 3. Check generated documentation
cat docs/api/user.md
cat README.md
```

> 🔍 **Verification Command**: `moai-adk doctor` — Checks if Python/uv versions, `.moai/` structure, and agent/Skills configuration are all ready.
> ```bash
> moai-adk doctor
> ```
> All green checkmarks mean perfect readiness!

---

## Keeping MoAI-ADK Up-to-Date

### Check Version
```bash
# Check currently installed version
moai-adk --version

# Check latest version on PyPI
uv tool list  # Check current version of moai-adk
```

### Upgrading

#### Method 1: MoAI-ADK Built-in Update Command (Simplest)
```bash
# MoAI-ADK's own update command - also updates agent/Skills templates
moai-adk update

# Apply new templates to project after update (optional)
moai-adk init .
```

#### Method 2: Upgrade with uv tool command

**Upgrade specific tool (recommended)**
```bash
# Upgrade only moai-adk to latest version
uv tool upgrade moai-adk
```

**Upgrade all installed tools**
```bash
# Upgrade all uv tool installations to latest versions
uv tool update
```

**Install specific version**
```bash
# Reinstall specific version (e.g., 0.4.2)
uv tool install moai-adk==0.4.2
```

### Verify After Update
```bash
# 1. Check installed version
moai-adk --version

# 2. Verify project works correctly
moai-adk doctor

# 3. Apply new templates to existing project (if needed)
cd your-project
moai-adk init .  # Keeps existing code, updates only .moai/ structure and templates

# 4. Check updated features in Alfred
cd your-project
claude
/alfred:0-project  # Verify new features like language selection
```

> 💡 **Tip**:
> - `moai-adk update`: Updates MoAI-ADK package version + syncs agent/Skills templates
> - `moai-adk init .`: Applies new templates to existing project (keeps code safe)
> - Running both commands completes a full update
> - When major updates (minor/major) release, run these procedures to utilize new agents/Skills

---

## Core Workflow (0 → 3)

Alfred iteratively develops projects with four commands.

```mermaid
%%{init: {'theme':'neutral'}}%%
graph TD
    Start([User Request]) --> Init[0. Init<br/>/alfred:0-project]
    Init --> Plan[1. Plan & SPEC<br/>/alfred:1-plan]
    Plan --> Run[2. Run & TDD<br/>/alfred:2-run]
    Run --> Sync[3. Sync & Docs<br/>/alfred:3-sync]
    Sync --> Plan
    Sync -.-> End([Release])
```

### 0. INIT — Project Preparation
- Questions about project introduction, target, language, mode (locale)
- Auto-generates `.moai/config.json`, `.moai/project/*` 5 documents
- Language detection and recommended Skill Pack deployment (Foundation + Essentials + Domain/Language)
- Template cleanup, initial Git/backup checks

### 1. PLAN — Agree on What to Build
- Write SPEC with EARS template (includes `@SPEC:ID`)
- Organize Plan Board, implementation ideas, risk factors
- Auto-create branch/initial Draft PR in Team mode

### 2. RUN — Test-Driven Development (TDD)
- Phase 1 `implementation-planner`: Design libraries, folders, TAG layout
- Phase 2 `tdd-implementer`: RED (failing test) → GREEN (minimal implementation) → REFACTOR (cleanup)
- quality-gate verifies TRUST 5 principles, coverage changes

### 3. SYNC — Documentation & PR Organization
- Sync Living Document, README, CHANGELOG, etc.
- Validate TAG chain and recover orphan TAGs
- Generate Sync Report, transition Draft → Ready for Review, support `--auto-merge` option

---

## Command Cheat Sheet

| Command                        | What it does                                                      | Key Outputs                                                        |
| ------------------------------ | ----------------------------------------------------------------- | ------------------------------------------------------------------ |
| `/alfred:0-project`            | Collect project description, create config/docs, recommend Skills | `.moai/config.json`, `.moai/project/*`, initial report             |
| `/alfred:1-plan <description>` | Analyze requirements, draft SPEC, write Plan Board                | `.moai/specs/SPEC-*/spec.md`, plan/acceptance docs, feature branch |
| `/alfred:2-run <SPEC-ID>`      | Execute TDD, test/implement/refactor, verify quality              | `tests/`, `src/` implementation, quality report, TAG connection    |
| `/alfred:3-sync`               | Sync docs/README/CHANGELOG, organize TAG/PR status                | `docs/`, `.moai/reports/sync-report.md`, Ready PR                  |

> ❗ All commands maintain **Phase 0 (optional) → Phase 1 → Phase 2 → Phase 3** cycle structure. Alfred automatically reports execution status and next-step suggestions.

---

## 5 Key Concepts

MoAI-ADK consists of 5 key concepts. Each concept connects to the others, and together they create a powerful development system.

### Key Concept 1: SPEC-First (Requirements First)

**Metaphor**: Like building a house without an architect, you shouldn't code without a blueprint.

**Core Idea**: Before implementation, **clearly define "what to build"**. This isn't just documentation—it's an **executable spec** that both teams and AI can understand.

**EARS Syntax 5 Patterns**:

1. **Ubiquitous** (basic function): "The system SHALL provide JWT-based authentication"
2. **Event-driven** (conditional): "WHEN valid credentials are provided, the system SHALL issue a token"
3. **State-driven** (during state): "WHILE the user is authenticated, the system SHALL allow access to protected resources"
4. **Optional** (optional): "WHERE a refresh token exists, the system MAY issue a new token"
5. **Constraints** (constraints): "Token expiration time SHALL NOT exceed 15 minutes"

**How?** The `/alfred:1-plan` command automatically creates professional SPECs in EARS format.

**What You Get**:
- ✅ Clear requirements everyone on the team understands
- ✅ SPEC-based test cases (what to test is already defined)
- ✅ When requirements change, track all affected code with `@SPEC:ID` TAG

---

### Key Concept 2: TDD (Test-Driven Development)

**Metaphor**: Like finding the route after setting a destination, you set goals with tests, then write code.

**Core Idea**: Write tests **before** implementation. Like checking ingredients before cooking, this clarifies requirements before implementation.

**3-Step Cycle**:

1. **🔴 RED**: Write a failing test first
   - Each SPEC requirement becomes a test case
   - Must fail because implementation doesn't exist yet
   - Git commit: `test(AUTH-001): add failing test`

2. **🟢 GREEN**: Minimal implementation to pass the test
   - Make it pass using the simplest approach
   - Passing comes before perfection
   - Git commit: `feat(AUTH-001): implement minimal solution`

3. **♻️ REFACTOR**: Clean up and improve code
   - Apply TRUST 5 principles
   - Remove duplication, improve readability
   - Tests must still pass
   - Git commit: `refactor(AUTH-001): improve code quality`

**How?** The `/alfred:2-run` command automatically executes these 3 steps.

**What You Get**:
- ✅ Guaranteed 85%+ coverage (no code without tests)
- ✅ Refactoring confidence (always verifiable with tests)
- ✅ Clear Git history (trace RED → GREEN → REFACTOR process)

---

### Key Concept 3: @TAG System

**Metaphor**: Like package tracking numbers, you should be able to trace code's journey.

**Core Idea**: Add `@TAG:ID` to all SPECs, tests, code, and documentation to create **one-to-one correspondence**.

**TAG Chain**:
```
@SPEC:AUTH-001 (requirements)
    ↓
@TEST:AUTH-001 (test)
    ↓
@CODE:AUTH-001 (implementation)
    ↓
@DOC:AUTH-001 (documentation)
```

**TAG ID Rules**: `<Domain>-<3 digits>`
- AUTH-001, AUTH-002, AUTH-003...
- USER-001, USER-002...
- Once assigned, **never change**

**How to Use?** When requirements change:
```bash
# Find everything related to AUTH-001
rg '@TAG:AUTH-001' -n

# Result: Shows all SPEC, TEST, CODE, DOC at once
# → Clear what needs modification
```

**How?** The `/alfred:3-sync` command validates TAG chains and detects orphan TAGs (TAGs without correspondence).

**What You Get**:
- ✅ Clear intent for all code (reading SPEC explains why this code exists)
- ✅ Instantly identify all affected code during refactoring
- ✅ Code remains understandable 3 months later (trace TAG → SPEC)

---

### Key Concept 4: TRUST 5 Principles

**Metaphor**: Like a healthy body, good code must satisfy all 5 elements.

**Core Idea**: All code must follow these 5 principles. `/alfred:3-sync` automatically verifies them.

1. **🧪 Test First** (tests come first)
   - Test coverage ≥ 85%
   - All code protected by tests
   - Adding feature = adding test

2. **📖 Readable** (easy-to-read code)
   - Functions ≤ 50 lines, files ≤ 300 lines
   - Variable names reveal intent
   - Pass linters (ESLint/ruff/clippy)

3. **🎯 Unified** (consistent structure)
   - Maintain SPEC-based architecture
   - Same patterns repeat (reduces learning curve)
   - Type safety or runtime validation

4. **🔒 Secured** (security)
   - Input validation (defend against XSS, SQL Injection)
   - Password hashing (bcrypt, Argon2)
   - Protect sensitive information (environment variables)

5. **🔗 Trackable** (traceability)
   - Use @TAG system
   - Include TAG in Git commits
   - Document all decisions

**How?** The `/alfred:3-sync` command automatically performs TRUST verification.

**What You Get**:
- ✅ Production-quality code guaranteed
- ✅ Entire team develops with same standards
- ✅ Fewer bugs, prevent security vulnerabilities in advance

---

### Key Concept 5: Alfred SuperAgent

**Metaphor**: Like a personal assistant, Alfred handles all the complex work.

**Core Idea**: **19 AI agents** collaborate to automate the entire development process:

**Agent Composition**:
- **Alfred SuperAgent**: Overall orchestration (1)
- **Core Sub-agents**: Specialized tasks like SPEC writing, TDD implementation, documentation sync (10)
- **Zero-project Specialists**: Project initialization, language detection, etc. (6)
- **Built-in Agents**: General questions, codebase exploration (2)

**55 Claude Skills**:
- **Foundation** (6): TRUST/TAG/SPEC/Git/EARS principles
- **Essentials** (4): Debugging, performance, refactoring, code review
- **Alfred** (7): Workflow automation
- **Domain** (10): Backend, frontend, security, etc.
- **Language** (18): Python, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, C/C#, Ruby, PHP, SQL, Shell, and more
- **Ops** (1): Claude Code session management
- **Other** (2): Skill factory, Spec authoring

**How?** `/alfred:*` commands automatically activate the right expert team.

**What You Get**:
- ✅ No prompt writing needed (use standardized commands)
- ✅ Automatically remember project context (no repeating same questions)
- ✅ Auto-assemble optimal expert team (activate appropriate Sub-agents)

> **Want to learn more?** Check detailed rules in `.moai/memory/development-guide.md`.

---

## First Hands-on: Todo API Example

Let's now **experience MoAI-ADK's complete workflow** firsthand. We'll build a simple "Todo Management API" and see how SPEC, TDD, and documentation connect.

### Step 1: PLAN - Write SPEC (about 3 minutes)

```bash
/alfred:1-plan "Todo add, view, update, delete API"
```

**Execution Result**:

Alfred's **spec-builder** automatically generates:

```yaml
# .moai/specs/SPEC-TODO-001/spec.md

---
id: TODO-001
version: 0.0.1
status: draft
created: 2025-10-22
updated: 2025-10-22
author: @user
priority: high
---

# @SPEC:TODO-001: Todo Management API

## Ubiquitous Requirements
- The system SHALL be able to add todos
- The system SHALL be able to view all todos
- The system SHALL be able to update specific todos
- The system SHALL be able to delete todos

## Event-driven Requirements
- WHEN a new todo is requested via POST /todos, the system SHALL save the todo and return a 201 response
- WHEN GET /todos/{id} is requested with an existing todo ID, the system SHALL return that todo
- WHEN GET is requested with a non-existent todo ID, the system SHALL return a 404 error

## Constraints
- Todo title SHALL be minimum 1 character, maximum 200 characters
- Each todo SHALL automatically record creation time
```

**Also auto-generated**:
- 📋 `Plan Board`: Implementation ideas, risk factors, solution strategies
- ✅ `Acceptance Criteria`: Verification standards
- 🌿 `feature/spec-todo-001` Git branch

### Step 2: RUN - TDD Implementation (about 5 minutes)

```bash
/alfred:2-run TODO-001
```

**Phase 1: Establish Implementation Strategy**

The **implementation-planner** Sub-agent decides:
- 📚 Libraries: FastAPI + SQLAlchemy
- 📁 Folder structure: `src/todo/`, `tests/todo/`
- 🏷️ TAG design: `@CODE:TODO-001:API`, `@CODE:TODO-001:MODEL`, `@CODE:TODO-001:REPO`

**Phase 2: RED → GREEN → REFACTOR**

**🔴 RED: Write Tests First**

```python
# tests/test_todo_api.py
# @TEST:TODO-001 | SPEC: SPEC-TODO-001.md

import pytest
from src.todo.api import create_todo, get_todos

def test_create_todo_should_return_201_with_todo_id():
    """WHEN a new todo is requested via POST /todos,
    the system SHALL save the todo and return a 201 response"""
    response = create_todo({"title": "Buy groceries"})
    assert response.status_code == 201
    assert "id" in response.json()
    assert response.json()["title"] == "Buy groceries"

def test_get_todos_should_return_all_todos():
    """The system SHALL be able to view all todos"""
    create_todo({"title": "Task 1"})
    create_todo({"title": "Task 2"})

    response = get_todos()
    assert response.status_code == 200
    assert len(response.json()) >= 2

def test_get_todo_with_invalid_id_should_return_404():
    """WHEN GET is requested with a non-existent todo ID,
    the system SHALL return a 404 error"""
    response = get_todo(999)
    assert response.status_code == 404
```

**Execute**: `pytest tests/test_todo_api.py -v`
**Result**: ❌ All fail (create_todo not defined)

**Git Commit**:
```bash
git commit -m "🔴 test(TODO-001): add failing API tests"
```

**🟢 GREEN: Minimal Implementation**

```python
# src/todo/api.py
# @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md | TEST: tests/test_todo_api.py

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uuid

app = FastAPI()
todos_db = {}

class TodoRequest(BaseModel):
    title: str

@app.post("/todos", status_code=201)
def create_todo(todo: TodoRequest):
    """@CODE:TODO-001:API - POST endpoint"""
    todo_id = str(uuid.uuid4())
    todos_db[todo_id] = {"id": todo_id, "title": todo.title}
    return todos_db[todo_id]

@app.get("/todos")
def get_todos():
    """@CODE:TODO-001:API - GET all endpoint"""
    return list(todos_db.values())

@app.get("/todos/{todo_id}")
def get_todo(todo_id: str):
    """@CODE:TODO-001:API - GET by ID endpoint"""
    if todo_id not in todos_db:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todos_db[todo_id]
```

**Execute**: `pytest tests/test_todo_api.py -v`
**Result**: ✅ All pass!

**Git Commit**:
```bash
git commit -m "🟢 feat(TODO-001): implement minimal Todo API"
```

**♻️ REFACTOR: Code Improvement**

```python
# src/todo/models.py
# @CODE:TODO-001:MODEL | SPEC: SPEC-TODO-001.md

from datetime import datetime
from sqlalchemy import Column, String, DateTime
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class Todo(Base):
    """@CODE:TODO-001:MODEL - Todo data model"""
    __tablename__ = "todos"

    id = Column(String, primary_key=True)
    title = Column(String(200), nullable=False, index=True)  # Constraint: max 200 chars
    created_at = Column(DateTime, default=datetime.utcnow)   # Auto creation time

    def validate(self):
        """@CODE:TODO-001:MODEL - Validation"""
        if not self.title or len(self.title) > 200:
            raise ValueError("Title must be 1-200 characters")
```

**Execute**: `pytest tests/test_todo_api.py -v`
**Result**: ✅ Still all pass!

**Git Commit**:
```bash
git commit -m "♻️ refactor(TODO-001): add database models and validation"
```

**Phase 3: Quality Verification**

```bash
# TRUST verification
✅ Test First: 87% coverage
✅ Readable: All functions < 50 lines
✅ Unified: Consistent API patterns
✅ Secured: Input validation complete
✅ Trackable: All code has @TAG:TODO-001
```

### Step 3: SYNC - Documentation Sync (about 1 minute)

```bash
/alfred:3-sync
```

**Automatically Performed**:

1. **TAG Chain Validation**
   ```bash
   ✅ @SPEC:TODO-001 → .moai/specs/SPEC-TODO-001/spec.md
   ✅ @TEST:TODO-001 → tests/test_todo_api.py
   ✅ @CODE:TODO-001 → src/todo/ (3 files)
   ✅ @DOC:TODO-001 → docs/api/todo.md (auto-generated)

   TAG Chain Integrity: 100%
   Orphan TAGs: None
   ```

2. **Living Document Generation**
   ```markdown
   # @DOC:TODO-001: Todo Management API

   ## Overview
   REST API for managing tasks with CRUD operations.

   ## Endpoints

   ### Create Todo
   - Method: POST
   - URL: /todos
   - Request: {"title": "string (1-200 chars)"}
   - Response: 201 Created with todo object
   - Implemented in: @CODE:TODO-001:API
   - Tested in: @TEST:TODO-001

   ### Get All Todos
   - Method: GET
   - URL: /todos
   - Response: 200 OK with array of todos

   [... etc ...]
   ```

3. **README Update**
   ```markdown
   ## Features

   - ✅ Todo Management API (TODO-001)
   ```

4. **CHANGELOG Generation**
   ```markdown
   # Changelog

   ## [0.1.0] - 2025-10-22

   ### Added
   - Todo Management API with CRUD operations (@SPEC:TODO-001)
     - Create new todos
     - List all todos
     - Update existing todos
     - Delete todos

   ### Implementation Details
   - SPEC: .moai/specs/SPEC-TODO-001/spec.md
   - Tests: tests/test_todo_api.py (87% coverage)
   - Code: src/todo/ with models, API, repository layers
   ```

### Step 4: Verification (about 1 minute)

Let's verify everything generated is properly connected:

```bash
# 1️⃣ Check TAG chain
rg '@(SPEC|TEST|CODE|DOC):TODO-001' -n

# Output:
# .moai/specs/SPEC-TODO-001/spec.md:1: # @SPEC:TODO-001: Todo Management API
# tests/test_todo_api.py:2: # @TEST:TODO-001 | SPEC: SPEC-TODO-001.md
# src/todo/api.py:5: # @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md
# src/todo/models.py:5: # @CODE:TODO-001:MODEL | SPEC: SPEC-TODO-001.md
# docs/api/todo.md:1: # @DOC:TODO-001: Todo Management API


# 2️⃣ Run tests
pytest tests/test_todo_api.py -v
# ✅ test_create_todo_should_return_201_with_todo_id PASSED
# ✅ test_get_todos_should_return_all_todos PASSED
# ✅ test_get_todo_with_invalid_id_should_return_404 PASSED
# ✅ 3 passed in 0.05s


# 3️⃣ Check generated documentation
cat docs/api/todo.md              # API documentation auto-generated
cat README.md                      # Todo API added
cat CHANGELOG.md                   # Change history recorded


# 4️⃣ Check Git history
git log --oneline | head -5
# a1b2c3d ✅ sync(TODO-001): update docs and changelog
# f4e5d6c ♻️ refactor(TODO-001): add database models
# 7g8h9i0 🟢 feat(TODO-001): implement minimal API
# 1j2k3l4 🔴 test(TODO-001): add failing tests
# 5m6n7o8 🌿 Create feature/spec-todo-001 branch
```

### After 15 Minutes: Complete System

```
✅ SPEC written (3 minutes)
   └─ @SPEC:TODO-001 TAG assigned
   └─ Clear requirements in EARS format

✅ TDD implementation (5 minutes)
   └─ 🔴 RED: Tests written first
   └─ 🟢 GREEN: Minimal implementation
   └─ ♻️ REFACTOR: Quality improvement
   └─ @TEST:TODO-001, @CODE:TODO-001 TAGs assigned
   └─ 87% coverage, TRUST 5 principles verified

✅ Documentation sync (1 minute)
   └─ Living Document auto-generated
   └─ README, CHANGELOG updated
   └─ TAG chain validation complete
   └─ @DOC:TODO-001 TAG assigned
   └─ PR status: Draft → Ready for Review

Result:
- 📋 Clear SPEC (SPEC-TODO-001.md)
- 🧪 85%+ test coverage (test_todo_api.py)
- 💎 Production-quality code (src/todo/)
- 📖 Auto-generated API documentation (docs/api/todo.md)
- 📝 Change history tracking (CHANGELOG.md)
- 🔗 Everything connected with TAGs
```

> **This is MoAI-ADK's true power.** Not just a simple API implementation,
> but a **complete development artifact** with everything from SPEC through tests, code, and documentation consistently connected!

---

## Sub-agents & Skills Overview

Alfred works by combining multiple specialized agents with Claude Skills.

### Core Sub-agents (Plan → Run → Sync)

| Sub-agent         | Model  | Role                                                                    |
| ----------------- | ------ | ----------------------------------------------------------------------- |
| project-manager 📋 | Sonnet | Project initialization, metadata interviews                             |
| spec-builder 🏗️    | Sonnet | Plan board, EARS SPEC authoring                                         |
| code-builder 💎    | Sonnet | Performs complete TDD with `implementation-planner` + `tdd-implementer` |
| doc-syncer 📖      | Haiku  | Living Doc, README, CHANGELOG sync                                      |
| tag-agent 🏷️       | Haiku  | TAG inventory, orphan detection                                         |
| git-manager 🚀     | Haiku  | GitFlow, Draft/Ready, Auto Merge                                        |
| debug-helper 🔍    | Sonnet | Failure analysis, fix-forward strategy                                  |
| trust-checker ✅   | Haiku  | TRUST 5 quality gate                                                    |
| quality-gate 🛡️    | Haiku  | Coverage change and release blocker review                              |
| cc-manager 🛠️      | Sonnet | Claude Code session optimization, Skill deployment                      |

### Skills (Progressive Disclosure - v0.4 New!)

Alfred organizes Claude Skills in a 4-tier architecture using **Progressive Disclosure** to load Just-In-Time only when needed. Each Skill is a production-grade guide stored in `.claude/skills/` directory.

#### Foundation Tier
Core skills containing fundamental TRUST/TAG/SPEC/Git/EARS/Language principles

| Skill                   | Description                                                                        |
| ----------------------- | ---------------------------------------------------------------------------------- |
| `moai-foundation-trust` | TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable) verification |
| `moai-foundation-tags`  | @TAG markers scan and inventory generation (CODE-FIRST principle)                  |
| `moai-foundation-specs` | SPEC YAML frontmatter validation and HISTORY section management                   |
| `moai-foundation-ears`  | EARS (Easy Approach to Requirements Syntax) requirements writing guide             |
| `moai-foundation-git`   | Git workflow automation (branching, TDD commits, PR management)                    |
| `moai-foundation-langs` | Project language/framework auto-detection (package.json, pyproject.toml, etc.)     |

#### Essentials Tier
Core tools needed for daily development work

| Skill                      | Description                                                            |
| -------------------------- | ---------------------------------------------------------------------- |
| `moai-essentials-debug`    | Stack trace analysis, error pattern detection, quick diagnosis support |
| `moai-essentials-perf`     | Performance profiling, bottleneck detection, tuning strategies         |
| `moai-essentials-refactor` | Refactoring guide, design patterns, code improvement strategies        |
| `moai-essentials-review`   | Automated code review, SOLID principles, code smell detection          |

#### Alfred Tier
MoAI-ADK internal workflow orchestration skills

| Skill                                  | Description                                                                                            |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `moai-alfred-ears-authoring`           | EARS syntax validation and requirement pattern guidance                                                |
| `moai-alfred-git-workflow`             | MoAI-ADK conventions (feature branch, TDD commits, Draft PR) automation                                |
| `moai-alfred-language-detection`       | Project language/runtime detection and test tool recommendations                                       |
| `moai-alfred-spec-metadata-validation` | SPEC YAML frontmatter and HISTORY section consistency validation                                       |
| `moai-alfred-tag-scanning`             | Complete @TAG marker scan and inventory generation (CODE-FIRST principle)                              |
| `moai-alfred-trust-validation`         | TRUST 5-principles compliance verification                                                             |
| `moai-alfred-interactive-questions`    | Claude Code Tools AskUserQuestion TUI menu standardization                                             |

#### Domain Tier
Specialized domain expertise

| Skill                      | Description                                                                              |
| -------------------------- | ---------------------------------------------------------------------------------------- |
| `moai-domain-backend`      | Backend architecture, API design, scaling guide                                          |
| `moai-domain-cli-tool`     | CLI tool development, argument parsing, POSIX compliance, user-friendly help messages    |
| `moai-domain-data-science` | Data analysis, visualization, statistical modeling, reproducible research workflows      |
| `moai-domain-database`     | Database design, schema optimization, indexing strategies, migration management          |
| `moai-domain-devops`       | CI/CD pipelines, Docker containerization, Kubernetes orchestration, IaC                  |
| `moai-domain-frontend`     | React/Vue/Angular development, state management, performance optimization, accessibility |
| `moai-domain-ml`           | Machine learning model training, evaluation, deployment, MLOps workflows                 |
| `moai-domain-mobile-app`   | Flutter/React Native development, state management, native integration                   |
| `moai-domain-security`     | OWASP Top 10, static analysis (SAST), dependency security, secrets management            |
| `moai-domain-web-api`      | REST API, GraphQL design patterns, authentication, versioning, OpenAPI documentation     |

#### Language Tier
Programming language-specific best practices

| Skill                  | Description                                               |
| ---------------------- | --------------------------------------------------------- |
| `moai-lang-python`     | pytest, mypy, ruff, black, uv package management          |
| `moai-lang-typescript` | Vitest, Biome, strict typing, npm/pnpm                    |
| `moai-lang-javascript` | Jest, ESLint, Prettier, npm package management            |
| `moai-lang-go`         | go test, golint, gofmt, standard library                  |
| `moai-lang-rust`       | cargo test, clippy, rustfmt, ownership/borrow checker     |
| `moai-lang-java`       | JUnit, Maven/Gradle, Checkstyle, Spring Boot patterns     |
| `moai-lang-kotlin`     | JUnit, Gradle, ktlint, coroutines, extension functions    |
| `moai-lang-swift`      | XCTest, SwiftLint, iOS/macOS development patterns         |
| `moai-lang-dart`       | flutter test, dart analyze, Flutter widget patterns       |
| `moai-lang-csharp`     | xUnit, .NET tooling, LINQ, async/await patterns           |
| `moai-lang-cpp`        | Google Test, clang-format, modern C++ (C++17/20)          |
| `moai-lang-c`          | Unity test framework, cppcheck, Make build system         |
| `moai-lang-scala`      | ScalaTest, sbt, functional programming patterns           |
| `moai-lang-ruby`       | RSpec, RuboCop, Bundler, Rails patterns                   |
| `moai-lang-php`        | PHPUnit, Composer, PSR standards                          |
| `moai-lang-sql`        | Test frameworks, query optimization, migration management |
| `moai-lang-shell`      | bats, shellcheck, POSIX compliance                        |
| `moai-lang-r`          | testthat, lintr, data analysis patterns                   |

#### Claude Code Ops
Claude Code session management

| Skill              | Description                                                                        |
| ------------------ | ---------------------------------------------------------------------------------- |
| `moai-claude-code` | Claude Code agents, commands, skills, plugins, settings scaffolding and monitoring |

> **v0.4.6 New Feature**: Claude Skills organized in 4-tier architecture (100% complete in v0.4.6). Each Skill loads via Progressive Disclosure only when needed to minimize context cost. Organized in Foundation → Essentials → Alfred → Domain/Language/Ops tiers, with all skills including production-grade documentation and executable TDD examples.

---

## AI Model Selection Guide

| Scenario                                             | Default Model         | Why                                             |
| ---------------------------------------------------- | --------------------- | ----------------------------------------------- |
| Specifications, design, refactoring, problem solving | **Claude 4.5 Sonnet** | Strong in deep reasoning and structured writing |
| Document sync, TAG checks, Git automation            | **Claude 4.5 Haiku**  | Strong in rapid iteration, string processing    |

- Start with Haiku for patterned tasks; switch to Sonnet when complex judgment is needed.
- If you manually change models, noting "why switched" in logs helps collaboration.

---

## Claude Code Hooks Guide

MoAI-ADK provides 4 main **Claude Code Hooks** that seamlessly integrate with your development workflow. These hooks enable automatic checkpoints, JIT context loading, and session monitoring—all happening transparently in the background.

### What Are Hooks?

Hooks are **event-driven** scripts that trigger automatically at specific points in your Claude Code session. Think of them as safety guardrails and productivity boosters that work behind the scenes without interrupting your flow.

### Installed Hooks

#### 1. SessionStart (Session Initialization)

**Triggers**: When you start a Claude Code session in your project
**Purpose**: Display project status at a glance

**What You See**:
```
🚀 MoAI-ADK Session Started
   Language: Python
   Branch: develop
   Changes: 2 files
   SPEC Progress: 12/25 (48%)
```

**Why It Matters**: Instantly understand your project's current state without running multiple commands.

#### 2. PreToolUse (Before Tool Execution)

**Triggers**: Before executing file edits, Bash commands, or MultiEdit operations
**Purpose**: Detect risky operations and automatically create safety checkpoints + TAG Guard

**Protection Against**:
- `rm -rf` (file deletion)
- `git merge`, `git reset --hard` (Git dangerous operations)
- Editing critical files (`CLAUDE.md`, `config.json`)
- Mass edits (10+ files at once via MultiEdit)

**TAG Guard (New in v0.4.11)**:
Automatically detects missing @TAG annotations in changed files:
- Scans staged, modified, and untracked files
- Warns when SPEC/TEST/CODE/DOC files lack required @TAG markers
- Configurable rules via `.moai/tag-rules.json`
- Non-blocking (gentle reminder, doesn't stop execution)

**What You See**:
```
🛡️ Checkpoint created: before-delete-20251023-143000
   Operation: delete
```

Or when TAGs are missing:
```
⚠️ TAG 누락 감지: 생성/수정한 파일 중 @TAG가 없는 항목이 있습니다.
 - src/auth/service.py → 기대 태그: @CODE:
 - tests/test_auth.py → 기대 태그: @TEST:
권장 조치:
  1) SPEC/TEST/CODE/DOC 유형에 맞는 @TAG를 파일 상단 주석이나 헤더에 추가
  2) rg로 확인: rg '@(SPEC|TEST|CODE|DOC):' -n <경로>
```

**Why It Matters**: Prevents data loss from mistakes and ensures @TAG traceability. You can always restore from the checkpoint if something goes wrong.

#### 3. UserPromptSubmit (Prompt Input)

**Triggers**: When you submit a prompt to Claude
**Purpose**: JIT (Just-In-Time) context loading—automatically add relevant files

**How It Works**:
- You type: "Fix AUTH bug"
- Hook scans for AUTH-related files
- Auto-loads: SPEC, tests, implementation, docs related to AUTH
- Claude receives full context without you manually specifying files

**Why It Matters**: Saves time and ensures Claude has all the relevant context for your request.

#### 4. SessionEnd (Session Cleanup)

**Triggers**: When you close your Claude Code session
**Purpose**: Cleanup tasks and state preservation

**Why It Matters**: Ensures clean session transitions and proper state management.

### Technical Details

- **Location**: `.claude/hooks/alfred/`
- **Environment Variable**: `$CLAUDE_PROJECT_DIR` (dynamically references project root)
- **Performance**: Each hook executes in <100ms
- **Logging**: Errors output to stderr (stdout reserved for JSON payloads)

### How to Disable Hooks

If you need to temporarily disable hooks, edit `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [],     // Disabled
    "PreToolUse": [...]     // Still active
  }
}
```

### Troubleshooting

**Problem: Hook doesn't execute**
- ✅ Verify `.claude/settings.json` is properly configured
- ✅ Check `uv` is installed: `which uv`
- ✅ Ensure hook script has execute permissions: `chmod +x .claude/hooks/alfred/alfred_hooks.py`

**Problem: Performance degradation**
- ✅ Check if any hook exceeds 100ms execution time
- ✅ Disable unnecessary hooks
- ✅ Review error messages in stderr output

**Problem: Too many checkpoints created**
- ✅ Review PreToolUse trigger conditions
- ✅ Adjust detection thresholds in `core/checkpoint.py` if needed

### Installed Hooks (5 total)

| Hook | Status | Feature |
|------|--------|---------|
| **SessionStart** | ✅ Active | Project status summary (language, Git, SPEC progress, checkpoints) |
| **PreToolUse** | ✅ Active | Risk detection + auto checkpoint (critical-delete, delete, merge, script) + **TAG Guard** (missing @TAG detection) |
| **UserPromptSubmit** | ✅ Active | JIT context loading (auto-load related SPEC, tests, code, docs) |
| **PostToolUse** | ✅ Active | Auto-run tests after code changes (9 languages: Python, TS, JS, Go, Rust, Java, Kotlin, Swift, Dart) |
| **SessionEnd** | ✅ Active | Session cleanup and state saving |

### Future Enhancements

- **Notification**: Important event alerts (logging, notifications)
- **Stop/SubagentStop**: Cleanup when agents terminate
- Advanced security: `dd` commands, supply chain checks

### Learn More

- Comprehensive analysis: `.moai/reports/hooks-analysis-and-implementation.md`
- PostToolUse implementation: `.moai/reports/phase3-posttool-implementation-complete.md`
- Security enhancements: `.moai/reports/security-enhancement-critical-delete.md`
- Hook implementation: `.claude/hooks/alfred/`
- Hook tests: `tests/hooks/`

---

## Frequently Asked Questions (FAQ)

- **Q. Can I install on an existing project?**
  - A. Yes. Run `moai-adk init .` to add only the `.moai/` structure without touching existing code.
- **Q. How do I run tests?**
  - A. `/alfred:2-run` runs them first; rerun `pytest`, `pnpm test`, etc. per language as needed.
- **Q. How do I ensure documentation stays current?**
  - A. `/alfred:3-sync` generates a Sync Report. Check the report in Pull Requests.
- **Q. Can I work manually?**
  - A. Yes, but keep the SPEC → TEST → CODE → DOC order and always leave TAGs.

---

## Latest Updates (New!)

| Version    | Key Features                                                                         | Date       |
| ---------- | ------------------------------------------------------------------------------------ | ---------- |
| **v0.4.11** | ✨ TAG Guard system + CLAUDE.md formatting improvements + Code cleanup                | 2025-10-23 |
| **v0.4.10** | 🔧 Hook robustness improvements + Bilingual documentation + Template language config | 2025-10-23 |
| **v0.4.9** | 🎯 Hook JSON schema validation fixes + Comprehensive tests (468/468 passing)        | 2025-10-23 |
| **v0.4.8** | 🚀 Release automation + PyPI deployment + Skills refinement                          | 2025-10-23 |
| **v0.4.7** | 📖 Korean language optimization + SPEC-First principle documentation                 | 2025-10-22 |
| **v0.4.6** | 🎉 Complete Skills v2.0 (100% Production-Ready) + 85,000 lines official docs + 300+ TDD examples | 2025-10-22 |

> 📦 **Install Now**: `pip install moai-adk==0.4.11` or `uv tool install moai-adk==0.4.11`

---

##

## Second Practice: Mini Kanban Board

This section goes beyond the first Todo API example and outlines a full 4-week full‑stack project.

Let’s build a Mini Kanban Board web application designed to help you master MoAI‑ADK end‑to‑end. This project lets you experience every step of SPEC‑First TDD.

### Project Overview

- Backend: FastAPI + Pydantic v2 + uv + WebSocket (Python)
- Frontend: React 19 + TypeScript 5.9 + Vite + Zustand + TanStack Query
- Real-time: Multi‑client sync over WebSocket
- Storage: Local filesystem (.moai/specs/)
- DevOps: Docker Compose + GitHub Actions CI/CD + Playwright E2E

### 4‑Week Timeline

```mermaid
gantt
    title Mini Kanban Board - 4-week plan
    dateFormat YYYY-MM-DD

    section Phase 1: Backend Basics
    Define SPEC-001-004           :active, ch07-spec, 2025-11-03, 1d
    Implement SpecScanner (TDD)   :active, ch07-impl, 2025-11-04, 1d

    section Phase 2: Backend Advanced
    Implement REST API            :active, ch08-api, 2025-11-05, 1d
    WebSocket + File Watch        :active, ch08-ws, 2025-11-06, 1d

    section Phase 3: Frontend Basics
    React init + SPEC-009-012     :active, ch09-spec, 2025-11-10, 1d
    Kanban Board (TDD)            :active, ch09-impl, 2025-11-11, 1d

    section Phase 4: Advanced + Deploy
    E2E + CI/CD                   :active, ch10-e2e, 2025-11-12, 1d
    Docker Compose + Optimize     :active, ch10-deploy, 2025-11-13, 1d
```

### 16‑SPEC Roadmap

| Phase | SPEC ID | Title | Stack | Est. | Status |
|------|---------|-------|-------|------|--------|
| Backend Basics | SPEC-001 | SPEC file scanner | FastAPI + pathlib + YAML | 1h | 📋 |
|  | SPEC-002 | YAML metadata parser | Pydantic v2 validation | 1h | 📋 |
|  | SPEC-003 | GET /api/specs (list) | FastAPI router | 0.5h | 📋 |
|  | SPEC-004 | GET /api/specs/{id} (detail) | FastAPI router | 0.5h | 📋 |
| Backend Advanced | SPEC-005 | PATCH /api/specs/{id}/status | FastAPI + update | 1h | 📋 |
|  | SPEC-006 | GET /api/specs/summary | Aggregation | 0.5h | 📋 |
|  | SPEC-007 | File watcher | watchdog + async | 1h | 📋 |
|  | SPEC-008 | WebSocket events | FastAPI WebSocket | 1.5h | 📋 |
| Frontend Basics | SPEC-009 | Kanban layout | React + CSS Grid | 1.5h | 📋 |
|  | SPEC-010 | SPEC card component | React + TypeScript | 1h | 📋 |
|  | SPEC-011 | TanStack Query integration | useQuery + useMutation | 1.5h | 📋 |
|  | SPEC-012 | Drag & Drop | React Beautiful DnD | 1.5h | 📋 |
| Advanced + Deploy | SPEC-013 | E2E automated tests | Playwright | 1.5h | 📋 |
|  | SPEC-014 | GitHub Actions CI/CD | Test + Release | 1h | 📋 |
|  | SPEC-015 | Docker Compose deploy | Multi‑container | 1h | 📋 |
|  | SPEC-016 | Performance + extensions | Caching + WS tuning | 1.5h | 📋 |
|  |  | Overall |  | 20h |  |

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Mini Kanban Board — Architecture              │
└─────────────────────────────────────────────────────────────┘

┌──────────────────────┐         ┌────────────────────────┐
│   📱 Frontend        │         │   🖥️ Backend Server    │
│  (React 19 + Vite)   │◄───────►│ (FastAPI + Pydantic)   │
│                      │  REST   │                        │
│ ┌──────────────────┐ │ API +   │ ┌──────────────────┐   │
│ │ DashboardHeader  │ │WebSocket│ │ GET /api/specs   │   │
│ ├──────────────────┤ │         │ ├──────────────────┤   │
│ │ KanbanBoard      │ │         │ │ PATCH /api/specs/{id}││
│ │ ┌──────────────┐ │ │         │ │ /status          │   │
│ │ │ Column: Draft│ │ │         │ ├──────────────────┤   │
│ │ │ Column: Active││ │         │ │ WebSocket        │   │
│ │ │ Column: Done │ │ │         │ │ /ws              │   │
│ │ └──────────────┘ │ │         │ │                  │   │
│ ├──────────────────┤ │         │ ├──────────────────┤   │
│ │ SpecCard (DnD)   │ │         │ │ SpecScanner      │   │
│ ├──────────────────┤ │         │ │ (.moai/specs/)   │   │
│ │ SearchBar        │ │         │ ├──────────────────┤   │
│ └──────────────────┘ │         │ │ YAML Parser      │   │
│                      │         │ │ (Pydantic v2)    │   │
│ Zustand Store:       │         │ └──────────────────┘   │
│ • filterStore        │         │                        │
│ • uiStore            │         │ File System:           │
│                      │         │ .moai/specs/           │
│ TanStack Query:      │         │ SPEC-001/              │
│ • useQuery           │         │ SPEC-002/              │
│ • useMutation        │         │ ...                    │
└──────────────────────┘         └────────────────────────┘
         │                                    │
         │            WebSocket               │
         └────────────────────────────────────┘
              (Real-time Sync)
```

### Phase Details

#### Phase 1: Backend Basics (SPEC-001~004)

Goal: Build the core data scanning service with FastAPI + Pydantic v2 + uv

```bash
# 1) Initialize project
/alfred:0-project
# → creates .moai/, backend/, frontend/
# → configures .moai/config.json

# 2) Write SPECs (SPEC-001~004)
/alfred:1-plan
# → SPEC-001: SPEC file scanner
# → SPEC-002: YAML metadata parser
# → SPEC-003: GET /api/specs endpoint
# → SPEC-004: GET /api/specs/{id} endpoint

# 3) TDD (RED → GREEN → REFACTOR)
/alfred:2-run SPEC-001
/alfred:2-run SPEC-002
/alfred:2-run SPEC-003
/alfred:2-run SPEC-004
```

Key Concepts:
- FastAPI project structure
- Pydantic v2 validation
- YAML front matter parsing
- Dependency Injection
- First TDD cycle completed

#### Phase 2: Backend Advanced (SPEC-005~008)

Goal: Implement file watching and WebSocket real-time events

```bash
# REST endpoints
/alfred:2-run SPEC-005  # PATCH /api/specs/{id}/status
/alfred:2-run SPEC-006  # GET /api/specs/summary

# WebSocket + File Watcher
/alfred:2-run SPEC-007  # File watching (watchdog)
/alfred:2-run SPEC-008  # WebSocket broadcast

# TRUST 5 verification
/alfred:3-sync          # verify all principles
```

Key Concepts:
- File system monitoring (watchdog)
- FastAPI WebSocket endpoint
- Async event broadcast
- Automated TRUST 5 verification

#### Phase 3: Frontend Basics (SPEC-009~012)

Goal: Build Kanban UI with React 19 + TypeScript + Vite

```bash
# Initialize React + Vite
cd frontend
npm create vite@latest . -- --template react-ts

# TanStack Query + Zustand
npm install @tanstack/react-query zustand

# SPECs
/alfred:1-plan SPEC-009  # layout
/alfred:1-plan SPEC-010  # card component
/alfred:1-plan SPEC-011  # TanStack Query integration
/alfred:1-plan SPEC-012  # drag & drop

# TDD
/alfred:2-run SPEC-009
/alfred:2-run SPEC-010
/alfred:2-run SPEC-011
/alfred:2-run SPEC-012
```

Key Concepts:
- React 19 Hooks (useState, useEffect, useContext)
- TypeScript 5.9 strict typing
- TanStack Query (useQuery, useMutation)
- Zustand state management
- React Beautiful DnD drag & drop

#### Phase 4: Advanced + Deploy (SPEC-013~016)

Goal: E2E tests, CI/CD, Docker deployment, performance optimization

```bash
# E2E tests (Playwright)
/alfred:2-run SPEC-013

# GitHub Actions CI/CD
/alfred:2-run SPEC-014

# Docker Compose deploy
/alfred:2-run SPEC-015

# Performance optimization
/alfred:2-run SPEC-016
```

Key Concepts:
- Playwright E2E automation
- GitHub Actions workflows
- Docker multi-stage builds
- Production performance tuning

### Quick Start Guide

#### Step 1: Initialize project

```bash
# Install MoAI-ADK
pip install moai-adk==0.4.6

# Create project
mkdir mini-kanban-board && cd mini-kanban-board
git init

# Initialize with Alfred
/alfred:0-project
```

#### Step 2: Write SPECs

```bash
# Start planning
/alfred:1-plan

# Answer prompts:
# - Project name: Mini Kanban Board
# - Tech stack: FastAPI + React 19
# - Duration: 4-week practice project
```

#### Step 3: Start TDD

```bash
# Phase 1 (Backend basics)
/alfred:2-run SPEC-001  # first TDD cycle

# Phase 2 (Backend advanced)
/alfred:2-run SPEC-005
/alfred:2-run SPEC-006
/alfred:2-run SPEC-007
/alfred:2-run SPEC-008

# Phase 3 (Frontend basics)
cd frontend
/alfred:2-run SPEC-009
/alfred:2-run SPEC-010
/alfred:2-run SPEC-011
/alfred:2-run SPEC-012

# Phase 4 (Advanced + deploy)
/alfred:2-run SPEC-013
/alfred:2-run SPEC-014
/alfred:2-run SPEC-015
/alfred:2-run SPEC-016
```

---

## Additional Resources

| Purpose                   | Resource                                                             |
| ------------------------- | -------------------------------------------------------------------- |
| Skills detailed structure | `.claude/skills/` directory (56 Skills)                              |
| Sub-agent details         | `.claude/agents/alfred/` directory                                   |
| Workflow guide            | `.claude/commands/alfred/` (0-3 commands)                            |
| Development guidelines    | `.moai/memory/development-guide.md`, `.moai/memory/spec-metadata.md` |
| Release notes             | GitHub Releases: https://github.com/modu-ai/moai-adk/releases        |

---

## Community & Support

| Channel                  | Link                                                    |
| ------------------------ | ------------------------------------------------------- |
| **GitHub Repository**    | https://github.com/modu-ai/moai-adk                     |
| **Issues & Discussions** | https://github.com/modu-ai/moai-adk/issues              |
| **PyPI Package**         | https://pypi.org/project/moai-adk/ (Latest: v0.4.11)     |
| **Latest Release**       | https://github.com/modu-ai/moai-adk/releases/tag/v0.4.11 |
| **Documentation**        | See `.moai/`, `.claude/`, `docs/` within project        |

---

## 🚀 MoAI-ADK Philosophy

> **"No CODE without SPEC"**

MoAI-ADK is not simply a code generation tool. Alfred SuperAgent with its 19-member team and 56 Claude Skills together guarantee:

- ✅ **SPEC → TEST (TDD) → CODE → DOCS consistency**
- ✅ **Complete history tracking with @TAG system**
- ✅ **Guaranteed 87.84%+ coverage**
- ✅ **Iterative development with 4-stage workflow (0-project → 1-plan → 2-run → 3-sync)**
- ✅ **Collaborate with AI transparently and traceably**

Start a new experience of **trustworthy AI development** with Alfred! 🤖

---

**MoAI-ADK v0.4.11** — SPEC-First TDD with AI SuperAgent & Complete Skills v2.0 + TAG Guard
- 📦 PyPI: https://pypi.org/project/moai-adk/
- 🏠 GitHub: https://github.com/modu-ai/moai-adk
- 📝 License: MIT
- ⭐ Skills: 55+ Production-Ready Guides
- ✅ Tests: 467/476 Passing (85.60% coverage)
- 🏷️ TAG Guard: Automatic @TAG validation in PreToolUse Hook

---
