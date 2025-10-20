# 🏗️ MoAI-ADK Architecture v2.0 (Claude Code 공식 표준)

> **Commands → Sub-agents → Skills 3-Tier 아키텍처**
>
> Claude Code 공식 문서 기반 설계

---

## 📐 아키텍처 개요

### Core Principles (Claude Code Official)

1. **Commands (Slash Commands)**: User-invoked, 워크플로우 오케스트레이션
2. **Agents (Sub-agents)**: Task tool로 위임, 독립 컨텍스트, 전문 작업 수행
3. **Skills**: Model-invoked, SKILL.md (YAML frontmatter), 자동 호출 (description 기반)

### 3-Tier Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ User Request                                                │
└─────────────────────────────────────────────────────────────┘
                          ↓ (slash command)
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Commands (4개)                                     │
│ ================================================            │
│ - User-invoked (slash commands)                             │
│ - 워크플로우 오케스트레이션                                 │
│ - Phase 1 (계획) → Phase 2 (실행)                           │
│ ================================================            │
│ Files: .claude/commands/alfred/*.md                         │
│                                                             │
│ ┌─────────────────┐ ┌─────────────────┐                    │
│ │ 0-project       │ │ 1-plan          │                    │
│ │ 프로젝트 초기화 │ │ SPEC 작성       │                    │
│ └─────────────────┘ └─────────────────┘                    │
│ ┌─────────────────┐ ┌─────────────────┐                    │
│ │ 2-run           │ │ 3-sync          │                    │
│ │ TDD 구현        │ │ 문서 동기화     │                    │
│ └─────────────────┘ └─────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
                          ↓ (Task tool)
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Sub-agents (17개)                                  │
│ ================================================            │
│ - Task tool로 위임                                          │
│ - 독립 컨텍스트 (각자의 메모리)                            │
│ - Skills 조합으로 전문성 확보                               │
│ ================================================            │
│ Files: .claude/agents/alfred/*.md                           │
│                                                             │
│ ┌─────────────────────────────────────────┐                │
│ │ Core Agents (9개)                       │                │
│ │ - spec-builder (Sonnet)                 │                │
│ │ - tdd-implementer (Sonnet)              │                │
│ │ - doc-syncer (Haiku)                    │                │
│ │ - tag-agent, git-manager (Haiku)        │                │
│ │ - debug-helper (Sonnet)                 │                │
│ │ - trust-checker (Haiku)                 │                │
│ │ - cc-manager, project-manager (Sonnet)  │                │
│ └─────────────────────────────────────────┘                │
│                                                             │
│ ┌─────────────────────────────────────────┐                │
│ │ 0-project Sub-agents (6개)              │                │
│ │ - language-detector (Haiku)             │                │
│ │ - backup-merger (Sonnet)                │                │
│ │ - project-interviewer (Sonnet)          │                │
│ │ - document-generator (Haiku)            │                │
│ │ - feature-selector (Haiku)              │                │
│ │ - template-optimizer (Haiku)            │                │
│ └─────────────────────────────────────────┘                │
│                                                             │
│ ┌─────────────────────────────────────────┐                │
│ │ Built-in (Claude Code 제공, 2개)        │                │
│ │ - Explore (Haiku): 코드베이스 탐색      │                │
│ │ - general-purpose (Sonnet): 범용 작업   │                │
│ └─────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────┘
                          ↓ (자동 호출)
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Skills (49개)                                      │
│ ================================================            │
│ - Model-invoked (Claude가 자동 판단)                        │
│ - SKILL.md (YAML frontmatter + Markdown)                    │
│ - description 기반 자동 호출                                │
│ ================================================            │
│ Files: .claude/skills/*/SKILL.md                            │
│                                                             │
│ ┌──────────────────────┐ ┌──────────────────────┐          │
│ │ Tier 1: Foundation   │ │ Tier 2: Language     │          │
│ │ (6개)                │ │ (28개)               │          │
│ │ - specs, ears, tags  │ │ - python, typescript │          │
│ │ - trust, langs, git  │ │ - java, go, rust...  │          │
│ └──────────────────────┘ └──────────────────────┘          │
│ ┌──────────────────────┐ ┌──────────────────────┐          │
│ │ Tier 3: Domain       │ │ Tier 4: Essentials   │          │
│ │ (10개)               │ │ (5개)                │          │
│ │ - backend, frontend  │ │ - debug, review      │          │
│ │ - mobile, ml, db...  │ │ - perf, refactor...  │          │
│ └──────────────────────┘ └──────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 워크플로우 예시: `/alfred:0-project`

### Phase 1: 분석 및 계획 수립

```markdown
User: "/alfred:0-project"

┌─────────────────────────────────────────┐
│ Command: 0-project.md                   │
│ (Alfred가 Command 프롬프트 실행)        │
└─────────────────────────────────────────┘
                ↓
        ┌───────────────┐
        │ 병렬 실행     │
        └───────────────┘
        ↓               ↓
┌──────────────┐  ┌──────────────┐
│ language-    │  │ backup-      │
│ detector     │  │ merger       │
│              │  │              │
│ Skills:      │  │ Skills:      │
│ - langs      │  │ (없음)       │
└──────────────┘  └──────────────┘
        ↓               ↓
    (결과 반환)     (결과 반환)
        ↓
        └───────┬─────────┘
                ↓
┌────────────────────────────┐
│ project-interviewer        │
│ (순차 실행: 병렬 완료 후) │
│                            │
│ Skills:                    │
│ - specs (메타데이터)       │
│ - ears (요구사항 작성)     │
└────────────────────────────┘
                ↓
┌────────────────────────────┐
│ Alfred: 계획 보고서 생성   │
│ + 사용자 승인 대기         │
└────────────────────────────┘
```

### Phase 2: 실행 (사용자 승인 후)

```markdown
User: "진행"

┌────────────────────────────┐
│ document-generator         │
│ (product/structure/tech.md)│
│                            │
│ Skills:                    │
│ - specs (YAML Front Matter)│
│ - ears (EARS 구문 적용)    │
│ - langs (언어별 템플릿)    │
└────────────────────────────┘
                ↓
┌────────────────────────────┐
│ Alfred: config.json 생성   │
└────────────────────────────┘
```

### Phase 3: 최적화 (선택적)

```markdown
User: "스킬 최적화 진행"

┌────────────────────────────┐
│ feature-selector           │
│ (49개 → 3~9개 선택)        │
│                            │
│ 입력:                      │
│ - product.md (도메인)      │
│ - tech.md (언어 스택)      │
│                            │
│ 출력:                      │
│ - selected_skills: [...]   │
└────────────────────────────┘
                ↓
┌────────────────────────────┐
│ template-optimizer         │
│ (불필요한 41개 삭제)       │
│                            │
│ Skills:                    │
│ - claude-code (템플릿 관리)│
│                            │
│ 작업:                      │
│ 1. CLAUDE.md 맞춤형 생성   │
│ 2. 불필요 스킬 삭제        │
│ 3. config.json 업데이트    │
└────────────────────────────┘
                ↓
┌────────────────────────────┐
│ Alfred: 완료 보고          │
│ - 84% 디스크 절약          │
│ - 컨텍스트 비용 절감       │
└────────────────────────────┘
```

---

## 📝 파일 구조 및 형식

### 1. Commands (Slash Commands)

**위치**: `.claude/commands/alfred/*.md`

**형식** (YAML frontmatter + Markdown):
```yaml
---
name: alfred:0-project
description: 프로젝트 문서 초기화 - product/structure/tech.md 생성 및 언어별 최적화 설정
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash(ls:*)
  - Bash(grep:*)
  - Task
---

# 📋 MoAI-ADK 0단계: 프로젝트 문서 초기화

## 실행 흐름 (2단계)

### Phase 1: 분석 및 계획 수립

#### 1.1 언어 감지 (병렬)
```python
Task(
    subagent_type="language-detector",
    description="프로젝트 언어 및 프레임워크 자동 감지",
    prompt="..."
)
```

#### 1.2 백업 병합 (병렬, 조건부)
...

### Phase 2: 실행
...
```

**핵심**:
- `Task` tool만 사용하여 Sub-agents에게 위임
- 순차/병렬 실행 명시
- Phase 1 (계획) → Phase 2 (실행) 2단계 워크플로우

---

### 2. Sub-agents

**위치**: `.claude/agents/alfred/*.md`

**형식** (YAML frontmatter + Markdown):
```yaml
---
name: language-detector
description: "Use when: 프로젝트 언어 및 프레임워크 자동 감지가 필요할 때"
tools: Read, Bash, Grep, Glob
model: haiku
skills:
  - moai-foundation-langs
---

# Language Detector - 언어 감지 에이전트

## 🎭 에이전트 페르소나
**아이콘**: 🔍
**직무**: 기술 분석가
**전문 영역**: 언어/프레임워크 감지, LanguageInterface 표준 제공

## 🎯 핵심 역할
- 설정 파일 스캔 (package.json, pyproject.toml, go.mod...)
- 언어 감지 (Python, TypeScript, Java, Go...)
- LanguageInterface JSON 응답 생성

## 📦 산출물
```json
{
  "language": "Python",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "type_checker": "mypy",
  "package_manager": "uv"
}
```

## 🔗 사용 Skills
- **moai-foundation-langs**: LanguageInterface 표준 제공
```

**핵심**:
- `skills` 필드에 사용하는 Skills 목록 명시
- Skills는 Claude가 자동 호출 (Model-invoked)
- Sub-agent는 Skills의 도메인 지식을 활용

---

### 3. Skills

**위치**: `.claude/skills/*/SKILL.md`

**형식** (YAML frontmatter + Markdown):
```yaml
---
name: moai-foundation-langs
tier: 1
description: Auto-detects project language and framework (package.json, pyproject.toml, etc) and provides LanguageInterface standard
allowed-tools:
- Read
- Bash
- Write
- Edit
---

# Alfred Language Detection & LanguageInterface

## What it does
Automatically detects project's primary language and framework by scanning configuration files, then recommends appropriate testing tools and linters.

## When to use
- "언어 감지", "프로젝트 언어 확인", "테스트 도구 추천"
- Automatically invoked by language-detector agent

## LanguageInterface Definition
```yaml
interface:
  language: "Python"
  test_framework: "pytest"
  linter: "ruff"
  formatter: "black"
  type_checker: "mypy"
  package_manager: "uv"
  version_requirement: ">=3.11"
```

## Works well with
- moai-lang-python
- moai-lang-typescript
- All other moai-lang-* skills
```

**핵심**:
- `tier` 필드로 계층 구조 명시
- `description`이 자동 호출 조건 (Claude가 판단)
- `allowed-tools`로 사용 가능한 도구 제한

---

## 🎯 Sub-agent ↔ Skills 관계

### Mapping Table

| Sub-agent            | 사용 Skills                                      | 호출 시점                      |
| -------------------- | ------------------------------------------------ | ------------------------------ |
| **language-detector** | moai-foundation-langs                            | 언어 감지 필요 시              |
| **spec-builder**      | moai-foundation-specs, moai-foundation-ears      | SPEC 작성 시                   |
| **tdd-implementer**   | moai-lang-python, moai-lang-typescript, ...      | TDD 구현 시 (언어별)           |
| **doc-syncer**        | moai-foundation-tags, moai-foundation-trust      | 문서 동기화, TAG 검증 시       |
| **debug-helper**      | moai-essentials-debug, moai-lang-* (언어별)      | 오류 진단 시                   |
| **trust-checker**     | moai-foundation-trust                            | TRUST 5원칙 검증 시            |
| **git-manager**       | moai-foundation-git                              | Git 작업 시                    |
| **feature-selector**  | (없음)                                           | Skills 선택 로직 (순수 알고리즘) |
| **template-optimizer**| moai-claude-code                                 | 템플릿 최적화 시               |

**원칙**:
- Sub-agent는 `skills` 필드에 사용할 Skills 명시
- Claude가 Sub-agent 실행 시 해당 Skills 자동 참조
- Skills는 도메인 지식 제공, Sub-agent는 로직 실행

---

## 🚀 feature-selector 로직 설계

### 입력

```json
{
  "project": {
    "language": "Python",
    "framework": "FastAPI",
    "domain": "backend"
  },
  "team": {
    "priority_areas": ["security", "performance"]
  }
}
```

### 선택 알고리즘

```python
# Tier 1 (Foundation): 항상 포함 (6개)
TIER_1_CORE = [
    "moai-foundation-specs",
    "moai-foundation-ears",
    "moai-foundation-tags",
    "moai-foundation-trust",
    "moai-foundation-langs",
    "moai-foundation-git"
]

# Tier 2 (Language): 언어별 1개 선택
TIER_2_MAP = {
    "Python": "moai-lang-python",
    "TypeScript": "moai-lang-typescript",
    "Java": "moai-lang-java",
    # ... 28개 언어
}

# Tier 3 (Domain): 도메인별 0~3개 선택
TIER_3_MAP = {
    "backend": ["moai-domain-backend", "moai-domain-web-api"],
    "frontend": ["moai-domain-frontend"],
    "mobile": ["moai-domain-mobile-app"],
    # ... 10개 도메인
}

# Tier 4 (Essentials): 우선순위별 0~2개 선택
TIER_4_MAP = {
    "security": "moai-domain-security",
    "performance": "moai-essentials-perf",
    "refactor": "moai-essentials-refactor",
    # ... 5개 Essentials
}

def select_skills(project, team):
    selected = []

    # Tier 1: 항상 포함
    selected.extend(TIER_1_CORE)  # 6개

    # Tier 2: 언어별 1개
    lang_skill = TIER_2_MAP[project["language"]]
    selected.append(lang_skill)  # +1개 = 7개

    # Tier 3: 도메인별 0~3개
    domain_skills = TIER_3_MAP[project["domain"]]
    selected.extend(domain_skills[:2])  # +2개 = 9개 (예시)

    # Tier 4: 우선순위별 0~2개 (선택적)
    # for priority in team["priority_areas"]:
    #     selected.append(TIER_4_MAP[priority])

    return selected  # 총 9개 (권장: 3~9개)
```

### 출력

```json
{
  "selected_skills": [
    {"tier": 1, "name": "moai-foundation-specs"},
    {"tier": 1, "name": "moai-foundation-ears"},
    {"tier": 1, "name": "moai-foundation-tags"},
    {"tier": 1, "name": "moai-foundation-trust"},
    {"tier": 1, "name": "moai-foundation-langs"},
    {"tier": 1, "name": "moai-foundation-git"},
    {"tier": 2, "name": "moai-lang-python"},
    {"tier": 3, "name": "moai-domain-backend"},
    {"tier": 3, "name": "moai-domain-web-api"}
  ],
  "total_selected": 9,
  "reduction": "49개 → 9개 (82% 감소)",
  "disk_saving": "12.8 MB"
}
```

---

## 🛠️ template-optimizer 로직 설계

### 입력

```json
{
  "selected_skills": [
    "moai-foundation-specs",
    "moai-foundation-ears",
    "moai-foundation-tags",
    "moai-foundation-trust",
    "moai-foundation-langs",
    "moai-foundation-git",
    "moai-lang-python",
    "moai-domain-backend",
    "moai-domain-web-api"
  ]
}
```

### 작업 흐름

```python
# 1. CLAUDE.md 맞춤형 생성
def generate_claude_md(selected_skills):
    template = read(".moai/templates/CLAUDE.md")

    # Skills 섹션 생성
    skills_section = "## 🎯 활성화된 Skills\n\n"
    for skill in selected_skills:
        skill_info = read(f".claude/skills/{skill}/SKILL.md")
        skills_section += f"- **{skill}**: {skill_info['description']}\n"

    # 템플릿 변수 치환
    claude_md = template.replace("{{SKILLS}}", skills_section)

    write(".moai/CLAUDE.md", claude_md)

# 2. 불필요한 스킬 파일 삭제
def cleanup_skills(selected_skills):
    all_skills = glob(".claude/skills/*/SKILL.md")

    for skill_path in all_skills:
        skill_name = extract_name(skill_path)

        if skill_name not in selected_skills:
            # 백업 후 삭제
            backup(skill_path, f".moai-backups/{timestamp}/{skill_name}")
            delete(skill_path)

# 3. config.json 업데이트
def update_config():
    config = read(".moai/config.json")
    config["optimized"] = True
    config["selected_skills"] = selected_skills
    write(".moai/config.json", config)
```

### 산출물

1. **맞춤형 CLAUDE.md**: 9개 Skills만 문서화
2. **삭제된 Skills**: 40개 (백업 보관)
3. **디스크 절약**: 12.8 MB
4. **config.json**: `optimized: true`, `selected_skills: [...]`

---

## 📊 성능 및 효과

### Before (현재)

- **Skills 개수**: 49개
- **컨텍스트 로딩**: 모든 Skills description 로드
- **디스크 사용**: 15.2 MB
- **처리 속도**: 느림 (Claude가 49개 Skills 중 선택)

### After (최적화 후)

- **Skills 개수**: 9개 (82% 감소)
- **컨텍스트 로딩**: 프로젝트 관련 Skills만
- **디스크 사용**: 2.4 MB (84% 절감)
- **처리 속도**: 빠름 (선택지 9개로 축소)

### 병렬 처리 효과

**순차 실행 (기존)**:
```
language-detector (5s) → backup-merger (3s) → project-interviewer (10s)
= 총 18초
```

**병렬 실행 (개선)**:
```
language-detector (5s)  ┐
                        ├─ 병렬 (5초)
backup-merger (3s)      ┘
                        ↓
project-interviewer (10s)
= 총 15초 (17% 단축)
```

---

## 🎯 적용 가이드

### 1단계: Commands 리팩토링

```bash
# Before (991 lines)
/alfred:0-project
→ 직접 처리 로직 (Bash, Read, Write...)

# After (300 lines)
/alfred:0-project
→ Task tool만 사용
→ Sub-agents에게 위임
```

### 2단계: Sub-agents 표준화

```bash
# 모든 .claude/agents/alfred/*.md 파일에 추가
---
skills:
  - moai-foundation-langs  # 사용하는 Skills 목록
---
```

### 3단계: Skills Tier 구조 명시

```bash
# 모든 .claude/skills/*/SKILL.md 파일에 추가
---
tier: 1  # 1 (Foundation), 2 (Language), 3 (Domain), 4 (Essentials)
---
```

### 4단계: 0-project 워크플로우 구현

```bash
# feature-selector, template-optimizer 추가
.claude/agents/alfred/
  ├─ feature-selector.md  (NEW)
  └─ template-optimizer.md  (NEW)
```

---

## 📚 참고 문서

- **Claude Code 공식 문서**: https://docs.claude.com/en/docs/claude-code/skills
- **Agent Skills 설계 패턴**: Anthropic Engineering Blog
- **Task Tool 사용법**: .moai/memory/development-guide.md
- **LanguageInterface 표준**: .claude/skills/moai-foundation-langs/SKILL.md

---

**작성자**: @Alfred
**버전**: v2.0
**최종 업데이트**: 2025-10-20
