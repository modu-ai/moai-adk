# Template Optimizer - 템플릿 최적화 에이전트 설계

> **feature-selector가 선택한 Skills를 기반으로 CLAUDE.md 맞춤형 생성 및 불필요한 파일 정리**

---

## 🎯 목적

feature-selector가 선택한 3~9개 Skills를 기반으로:
1. **CLAUDE.md 맞춤형 생성**: 선택된 Skills만 문서화
2. **불필요한 Skills 삭제**: 40개 미선택 Skills를 백업 후 삭제
3. **config.json 업데이트**: `optimized: true` 플래그 설정
4. **디스크 절약**: 12.8 MB 절약 (84%)

---

## 📥 입력

**feature-selector의 출력**:
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
  "total_selected": 9
}
```

---

## 🔄 작업 흐름 (3단계)

### 1단계: CLAUDE.md 맞춤형 생성

#### 템플릿 구조

```markdown
# {{PROJECT_NAME}} - MoAI-Agentic Development Kit

## ▶◀ Meet Alfred: Your MoAI SuperAgent

...

## 🎯 활성화된 Skills ({{TOTAL_SELECTED}}개)

{{SKILLS_SECTION}}

---

## 핵심 철학
...
```

#### Skills 섹션 생성 로직

```python
def generate_skills_section(selected_skills):
    """선택된 Skills를 Tier별로 그룹화하여 문서화"""

    # Tier별 그룹화
    skills_by_tier = {
        1: [],  # Foundation
        2: [],  # Language
        3: [],  # Domain
        4: []   # Essentials
    }

    for skill in selected_skills:
        tier = skill["tier"]
        name = skill["name"]

        # Skill 정보 읽기
        skill_md = read(f".claude/skills/{name}/SKILL.md")

        skills_by_tier[tier].append({
            "name": name,
            "description": skill_md["description"],
            "tier_name": skill_md.get("tier_name", "")
        })

    # Markdown 생성
    output = "## 🎯 활성화된 Skills\n\n"

    # Tier 1: Foundation
    if skills_by_tier[1]:
        output += "### Tier 1: Foundation (핵심 기반)\n\n"
        for skill in skills_by_tier[1]:
            output += f"- **{skill['name']}**: {skill['description']}\n"
        output += "\n"

    # Tier 2: Language
    if skills_by_tier[2]:
        output += "### Tier 2: Language (언어별 도구)\n\n"
        for skill in skills_by_tier[2]:
            output += f"- **{skill['name']}**: {skill['description']}\n"
        output += "\n"

    # Tier 3: Domain
    if skills_by_tier[3]:
        output += "### Tier 3: Domain (도메인 전문)\n\n"
        for skill in skills_by_tier[3]:
            output += f"- **{skill['name']}**: {skill['description']}\n"
        output += "\n"

    # Tier 4: Essentials (선택적)
    if skills_by_tier[4]:
        output += "### Tier 4: Essentials (추가 도구)\n\n"
        for skill in skills_by_tier[4]:
            output += f"- **{skill['name']}**: {skill['description']}\n"
        output += "\n"

    output += "---\n\n"
    output += f"**총 {len(selected_skills)}개 Skills 활성화** (49개 → {len(selected_skills)}개, "
    output += f"{100 - int(len(selected_skills) / 49 * 100)}% 감소)\n\n"

    return output
```

#### 최종 CLAUDE.md 생성

```python
def generate_claude_md(selected_skills, project_info):
    """맞춤형 CLAUDE.md 생성"""

    # 템플릿 읽기
    template = read("src/moai_adk/templates/.moai/CLAUDE.md")

    # Skills 섹션 생성
    skills_section = generate_skills_section(selected_skills)

    # 변수 치환
    claude_md = template.replace("{{SKILLS_SECTION}}", skills_section)
    claude_md = claude_md.replace("{{TOTAL_SELECTED}}", str(len(selected_skills)))
    claude_md = claude_md.replace("{{PROJECT_NAME}}", project_info["name"])

    # 저장
    write(".moai/CLAUDE.md", claude_md)

    return claude_md
```

---

### 2단계: 불필요한 Skills 삭제

#### 백업 후 삭제

```python
def cleanup_skills(selected_skills):
    """미선택 Skills를 백업 후 삭제"""

    # 선택된 Skills 이름 목록
    selected_names = [skill["name"] for skill in selected_skills]

    # 모든 Skills 목록
    all_skills = glob(".claude/skills/*/SKILL.md")

    # 삭제할 Skills
    to_delete = []

    for skill_path in all_skills:
        skill_name = extract_skill_name(skill_path)

        if skill_name not in selected_names:
            to_delete.append(skill_name)

    # 백업 디렉토리 생성
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    backup_dir = f".moai-backups/{timestamp}/skills/"
    mkdir(backup_dir)

    # 백업 및 삭제
    deleted_count = 0
    for skill_name in to_delete:
        skill_dir = f".claude/skills/{skill_name}/"

        # 백업
        backup_path = f"{backup_dir}{skill_name}/"
        copy(skill_dir, backup_path)

        # 삭제
        remove(skill_dir)
        deleted_count += 1

    return {
        "deleted_count": deleted_count,
        "backup_dir": backup_dir,
        "deleted_skills": to_delete
    }
```

#### 디스크 절약 계산

```python
def calculate_disk_saving(deleted_skills):
    """삭제된 Skills의 디스크 절약량 계산"""

    total_saved = 0

    for skill_name in deleted_skills:
        skill_dir = f".moai-backups/{timestamp}/skills/{skill_name}/"
        skill_size = get_dir_size(skill_dir)  # bytes
        total_saved += skill_size

    return {
        "saved_bytes": total_saved,
        "saved_mb": round(total_saved / 1024 / 1024, 1),
        "percentage": int((len(deleted_skills) / 49) * 100)
    }
```

---

### 3단계: config.json 업데이트

```python
def update_config(selected_skills):
    """config.json에 최적화 정보 저장"""

    # 기존 config 읽기
    config = read(".moai/config.json")

    # 업데이트
    config["optimized"] = True
    config["selected_skills"] = [skill["name"] for skill in selected_skills]
    config["optimization_date"] = datetime.now().isoformat()

    # 저장
    write(".moai/config.json", json.dumps(config, indent=2))

    return config
```

---

## 📤 산출물

### 1. 맞춤형 CLAUDE.md

```markdown
# MoAI-ADK - MoAI-Agentic Development Kit

## ▶◀ Meet Alfred: Your MoAI SuperAgent

...

## 🎯 활성화된 Skills (9개)

### Tier 1: Foundation (핵심 기반)

- **moai-foundation-specs**: SPEC 메타데이터 표준 제공
- **moai-foundation-ears**: EARS 요구사항 작성 방법론
- **moai-foundation-tags**: @TAG 시스템 관리
- **moai-foundation-trust**: TRUST 5원칙 검증
- **moai-foundation-langs**: LanguageInterface 표준 제공
- **moai-foundation-git**: Git 워크플로우 자동화

### Tier 2: Language (언어별 도구)

- **moai-lang-python**: Python 최적 도구 (pytest, ruff, black, mypy, uv)

### Tier 3: Domain (도메인 전문)

- **moai-domain-backend**: 백엔드 아키텍처, API 설계, 캐싱 전략
- **moai-domain-web-api**: REST API, GraphQL 설계 패턴

---

**총 9개 Skills 활성화** (49개 → 9개, 82% 감소)
```

### 2. 삭제 보고서

```json
{
  "deleted_skills": [
    "moai-lang-typescript",
    "moai-lang-java",
    "moai-lang-go",
    "moai-lang-rust",
    "moai-domain-frontend",
    "moai-domain-mobile-app",
    "moai-domain-ml",
    "moai-essentials-debug",
    "moai-essentials-review",
    "moai-essentials-refactor",
    "moai-essentials-perf",
    // ... 총 40개
  ],
  "deleted_count": 40,
  "backup_dir": ".moai-backups/20251020-153000/skills/",
  "disk_saving": {
    "saved_mb": 12.8,
    "percentage": 82
  }
}
```

### 3. 업데이트된 config.json

```json
{
  "project": {
    "name": "MoAI-ADK",
    "version": "0.4.0",
    "mode": "personal",
    "locale": "ko"
  },
  "optimized": true,
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
  ],
  "optimization_date": "2025-10-20T15:30:00"
}
```

---

## 🔄 Alfred에게 반환하는 메시지

```markdown
✅ 템플릿 최적화 완료

### 생성된 파일
- ✅ `.moai/CLAUDE.md` (맞춤형, 9개 Skills만 문서화)

### 삭제된 파일
- 🗑️ 40개 Skills 삭제 (백업: `.moai-backups/20251020-153000/skills/`)

### 디스크 절약
- 💾 12.8 MB 절약 (15.2 MB → 2.4 MB, 84% 감소)

### 업데이트
- ⚙️ `.moai/config.json` 업데이트 (`optimized: true`)

### 다음 단계
1. 맞춤형 CLAUDE.md 확인
2. 프로젝트 개발 시작 (/alfred:1-plan)
```

---

## ⚙️ 에이전트 YAML Frontmatter

```yaml
---
name: template-optimizer
description: "Use when: 템플릿 최적화 및 파일 정리가 필요할 때. /alfred:0-project 커맨드에서 호출"
tools: Read, Write, Edit, MultiEdit, Bash, Glob
model: haiku
skills:
  - moai-claude-code  # Claude Code 템플릿 관리
---
```

---

## 📊 성능 비교

### Before (최적화 전)

```
.claude/skills/
├─ moai-foundation-specs/
├─ moai-foundation-ears/
├─ moai-lang-python/
├─ moai-lang-typescript/
├─ moai-lang-java/
├─ moai-lang-go/
├─ moai-domain-backend/
├─ moai-domain-frontend/
├─ moai-domain-mobile-app/
├─ ...
└─ (총 49개, 15.2 MB)
```

### After (최적화 후)

```
.claude/skills/
├─ moai-foundation-specs/
├─ moai-foundation-ears/
├─ moai-foundation-tags/
├─ moai-foundation-trust/
├─ moai-foundation-langs/
├─ moai-foundation-git/
├─ moai-lang-python/
├─ moai-domain-backend/
└─ moai-domain-web-api/
(총 9개, 2.4 MB)

.moai-backups/20251020-153000/skills/
├─ moai-lang-typescript/
├─ moai-lang-java/
├─ moai-lang-go/
├─ ...
└─ (40개 백업)
```

---

## 🎯 복원 기능 (선택적)

**사용자가 삭제된 Skill을 다시 활성화하고 싶을 때**:

```python
def restore_skill(skill_name):
    """백업된 Skill을 복원"""

    # 최신 백업 찾기
    backups = glob(".moai-backups/*/skills/{skill_name}/")
    latest_backup = sorted(backups)[-1]

    # 복원
    copy(latest_backup, f".claude/skills/{skill_name}/")

    # config.json 업데이트
    config = read(".moai/config.json")
    config["selected_skills"].append(skill_name)
    write(".moai/config.json", json.dumps(config, indent=2))

    return f"✅ {skill_name} 복원 완료"
```

**사용 예시**:
```bash
# moai-adk CLI
moai-adk restore-skill moai-lang-typescript
```

---

## 🚨 주의사항

### 1. 백업 필수

- 모든 삭제 작업 전 `.moai-backups/` 디렉토리에 백업
- 백업 경로: `.moai-backups/{timestamp}/skills/{skill-name}/`

### 2. config.json 검증

- `optimized: true` 플래그 설정 후 재초기화 방지
- 재초기화 시 백업 병합 워크플로우 실행

### 3. CLAUDE.md 충돌 방지

- 기존 CLAUDE.md가 사용자 커스터마이징을 포함하는 경우 백업 후 병합
- 템플릿 기본값 vs 사용자 커스터마이징 자동 탐지

---

## 🔗 feature-selector와의 연계

```python
# /alfred:0-project 실행 흐름

# Phase 3: 최적화
feature_result = Task(subagent="feature-selector", ...)
→ selected_skills: [9개]

Task(
    subagent="template-optimizer",
    prompt=f"""
    선택된 Skills: {feature_result['selected_skills']}

    작업:
    1. CLAUDE.md 맞춤형 생성
    2. 불필요한 40개 Skills 삭제
    3. config.json 업데이트
    """
)
→ 최적화 완료 보고
```

---

**작성자**: @Alfred
**최종 업데이트**: 2025-10-20
