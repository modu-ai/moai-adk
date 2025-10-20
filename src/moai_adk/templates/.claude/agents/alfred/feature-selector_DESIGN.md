# Feature Selector - 기능 선택 에이전트 설계

> **49개 Skills 중 프로젝트에 최적화된 3~9개를 자동 선택하는 알고리즘 설계**

---

## 🎯 목적

`/alfred:0-project` 실행 시 생성된 project 문서 3개(product.md, structure.md, tech.md)를 기반으로:
- 49개 Skills 중 프로젝트에 필요한 **최소 3개 ~ 최대 9개**를 자동 선택
- 불필요한 Skills는 template-optimizer가 삭제하여 **컨텍스트 비용 절감** (84%)

---

## 📊 Skills Tier 구조

### Tier 1: Foundation (6개) - 항상 포함 ✅

**필수 Core Skills**:
```yaml
- moai-foundation-specs   # SPEC 메타데이터 표준
- moai-foundation-ears    # EARS 요구사항 작성
- moai-foundation-tags    # @TAG 시스템
- moai-foundation-trust   # TRUST 5원칙
- moai-foundation-langs   # LanguageInterface 표준
- moai-foundation-git     # Git 워크플로우
```

**이유**: MoAI-ADK의 핵심 워크플로우에 필수

---

### Tier 2: Language (28개) - 언어별 1개 선택 ✅

**입력**: `tech.md`의 `language` 필드

**Mapping Table**:
```python
LANGUAGE_SKILL_MAP = {
    # 주요 8개 언어
    "Python": "moai-lang-python",
    "TypeScript": "moai-lang-typescript",
    "Java": "moai-lang-java",
    "Go": "moai-lang-go",
    "Rust": "moai-lang-rust",
    "Dart": "moai-lang-dart",
    "Swift": "moai-lang-swift",
    "Kotlin": "moai-lang-kotlin",

    # 추가 20개 언어
    "JavaScript": "moai-lang-javascript",
    "C++": "moai-lang-cpp",
    "C": "moai-lang-c",
    "C#": "moai-lang-csharp",
    "Ruby": "moai-lang-ruby",
    "PHP": "moai-lang-php",
    "Elixir": "moai-lang-elixir",
    "Scala": "moai-lang-scala",
    "Clojure": "moai-lang-clojure",
    "Haskell": "moai-lang-haskell",
    "Lua": "moai-lang-lua",
    "R": "moai-lang-r",
    "Julia": "moai-lang-julia",
    "SQL": "moai-lang-sql",
    "Shell": "moai-lang-shell",
    # ... 총 28개
}
```

**선택 로직**:
```python
def select_language_skill(tech_md):
    language = tech_md["language"]  # "Python"
    return LANGUAGE_SKILL_MAP.get(language, "moai-lang-python")  # default
```

---

### Tier 3: Domain (10개) - 도메인별 0~3개 선택 ⚙️

**입력**: `product.md`의 도메인, `structure.md`의 아키텍처

**Mapping Table**:
```python
DOMAIN_SKILL_MAP = {
    # 핵심 도메인
    "backend": [
        "moai-domain-backend",      # 우선순위 1
        "moai-domain-web-api"       # 우선순위 2 (API 제공 시)
    ],
    "frontend": [
        "moai-domain-frontend"      # 우선순위 1
    ],
    "mobile": [
        "moai-domain-mobile-app"    # 우선순위 1
    ],
    "ml": [
        "moai-domain-ml"            # 우선순위 1
    ],
    "data-science": [
        "moai-domain-data-science"  # 우선순위 1
    ],
    "cli": [
        "moai-domain-cli-tool"      # 우선순위 1
    ],
    "database": [
        "moai-domain-database"      # 우선순위 1
    ],
    "devops": [
        "moai-domain-devops"        # 우선순위 1
    ],
    "security": [
        "moai-domain-security"      # 우선순위 1
    ]
}
```

**선택 로직**:
```python
def select_domain_skills(product_md, structure_md):
    domain = product_md["domain"]  # "backend"
    selected = []

    # 주 도메인 스킬 (최대 2개)
    domain_skills = DOMAIN_SKILL_MAP.get(domain, [])
    selected.extend(domain_skills[:2])

    # 아키텍처 기반 추가 도메인 (최대 1개)
    if "API" in structure_md["architecture"]:
        if "moai-domain-web-api" not in selected:
            selected.append("moai-domain-web-api")

    return selected  # 최대 3개
```

---

### Tier 4: Essentials (5개) - 우선순위별 0~2개 선택 (선택적) 📦

**입력**: `product.md`의 `team.priority_areas`

**Mapping Table**:
```python
ESSENTIALS_SKILL_MAP = {
    "debug": "moai-essentials-debug",
    "review": "moai-essentials-review",
    "refactor": "moai-essentials-refactor",
    "performance": "moai-essentials-perf"
}
```

**선택 로직** (선택적, 기본적으로 제외):
```python
def select_essentials_skills(product_md, max_count=0):
    """기본적으로 Essentials는 제외 (max_count=0)"""
    priority_areas = product_md.get("team", {}).get("priority_areas", [])
    selected = []

    for area in priority_areas[:max_count]:
        skill = ESSENTIALS_SKILL_MAP.get(area)
        if skill:
            selected.append(skill)

    return selected  # 최대 max_count개 (기본 0개)
```

---

## 🔄 전체 선택 알고리즘

```python
def select_optimal_skills(product_md, structure_md, tech_md):
    """
    프로젝트에 최적화된 Skills 선택

    입력:
    - product.md: 프로젝트 도메인, 팀 우선순위
    - structure.md: 아키텍처
    - tech.md: 언어, 프레임워크

    출력:
    - selected_skills: 최종 선택된 Skills (3~9개)
    """
    selected = []

    # Tier 1: Foundation (6개, 항상 포함)
    TIER_1_CORE = [
        "moai-foundation-specs",
        "moai-foundation-ears",
        "moai-foundation-tags",
        "moai-foundation-trust",
        "moai-foundation-langs",
        "moai-foundation-git"
    ]
    selected.extend(TIER_1_CORE)  # 6개

    # Tier 2: Language (1개, 언어별)
    lang_skill = select_language_skill(tech_md)
    selected.append(lang_skill)  # +1개 = 7개

    # Tier 3: Domain (0~3개, 도메인별)
    domain_skills = select_domain_skills(product_md, structure_md)
    selected.extend(domain_skills)  # +0~3개 = 7~10개

    # Tier 4: Essentials (0개, 기본적으로 제외)
    # essentials = select_essentials_skills(product_md, max_count=0)
    # selected.extend(essentials)

    # 최종 검증 (3~9개)
    if len(selected) < 3:
        raise ValueError("최소 3개 Skills 필요")
    if len(selected) > 9:
        # 도메인 스킬 축소
        selected = selected[:9]

    return selected
```

---

## 📋 입력 예시

### product.md (YAML Front Matter)

```yaml
---
domain: backend
team:
  mode: personal
  priority_areas:
    - security
    - performance
---
```

### structure.md

```markdown
## 아키텍처

- **유형**: 모놀리식 백엔드
- **API**: REST API 제공
- **데이터베이스**: PostgreSQL
```

### tech.md (YAML Front Matter)

```yaml
---
language: Python
framework: FastAPI
---
```

---

## 📤 출력 예시

### JSON 응답

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
  "reduction": {
    "before": 49,
    "after": 9,
    "percentage": 82
  },
  "disk_saving": {
    "before_mb": 15.2,
    "after_mb": 2.4,
    "saved_mb": 12.8
  }
}
```

---

## 🔗 template-optimizer로 전달

**다음 단계**: feature-selector가 선택한 Skills를 template-optimizer에게 전달

```python
# feature-selector 완료 후
Task(
    subagent_type="template-optimizer",
    description="불필요한 Skills 삭제 및 CLAUDE.md 맞춤형 생성",
    prompt=f"""
    선택된 Skills: {selected_skills}

    작업:
    1. CLAUDE.md 맞춤형 생성
    2. 불필요한 40개 Skills 삭제
    3. config.json 업데이트 (optimized: true)
    """
)
```

---

## ⚙️ 에이전트 YAML Frontmatter

```yaml
---
name: feature-selector
description: "Use when: 49개 스킬 중 3~9개 최적 선택이 필요할 때. /alfred:0-project 커맨드에서 호출"
tools: Read, Bash, TodoWrite
model: haiku
skills: []  # 순수 알고리즘, Skills 불필요
---
```

---

## 📊 기대 효과

### Before (최적화 전)

- **Skills 개수**: 49개
- **디스크 사용**: 15.2 MB
- **컨텍스트 로딩**: 모든 Skills description 로드 (느림)

### After (최적화 후)

- **Skills 개수**: 9개 (82% 감소)
- **디스크 사용**: 2.4 MB (84% 절약)
- **컨텍스트 로딩**: 프로젝트 관련 Skills만 (빠름)

---

**작성자**: @Alfred
**최종 업데이트**: 2025-10-20
