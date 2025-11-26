# SPEC-SKILL-PORTFOLIO-OPT-001: 인수 기준 (Acceptance Criteria)

---
id: SPEC-SKILL-PORTFOLIO-OPT-001-ACCEPTANCE
created_at: 2025-11-22
updated_at: 2025-11-22
version: 1.0.0
---

## 개요 (Overview)

본 인수 기준은 **MoAI-ADK Skills Portfolio 최적화 및 표준화** 작업의 **완료 조건**을 정의합니다.

모든 7개 요구사항(REQ-001 ~ REQ-007)에 대해 **Given-When-Then 형식**의 테스트 시나리오를 제공하며, 각 시나리오는 **자동 검증 가능**해야 합니다.

---

## REQ-001: 카테고리 통합 (32 → 10 티어)

### AC-001-1: 모든 스킬이 10개 티어 중 하나에 할당됨

**Given**: 총 127개 스킬이 존재하고, 10개 티어가 정의되어 있음
**When**: 티어 할당 스크립트를 실행함
**Then**:
- [ ] 127개 모든 스킬에 `category_tier` 필드가 존재함
- [ ] `category_tier` 값이 1-10 또는 "special" 중 하나임
- [ ] 미분류 스킬 수 = 0개

**검증 방법**:
```python
def test_all_skills_assigned_to_tier():
    skills = load_all_skills()
    for skill in skills:
        assert "category_tier" in skill.metadata
        assert skill.metadata["category_tier"] in range(1, 11) or \
               skill.metadata["category_tier"] == "special"

    unassigned_skills = [s for s in skills if "category_tier" not in s.metadata]
    assert len(unassigned_skills) == 0
```

---

### AC-001-2: 티어별 스킬 수가 균형적으로 분포됨

**Given**: 10개 티어에 127개 스킬이 할당됨
**When**: 티어별 스킬 수를 집계함
**Then**:
- [ ] 각 티어의 스킬 수가 1-15개 범위 내에 있음 (과도한 집중 방지)
- [ ] 가장 많은 티어와 가장 적은 티어의 차이가 20개 이하임

**검증 방법**:
```python
def test_tier_distribution_balanced():
    skills = load_all_skills()
    tier_counts = {}
    for skill in skills:
        tier = skill.metadata.get("category_tier")
        tier_counts[tier] = tier_counts.get(tier, 0) + 1

    for tier, count in tier_counts.items():
        if tier != "special":
            assert 1 <= count <= 15, f"Tier {tier} has {count} skills (expected 1-15)"

    min_count = min(tier_counts.values())
    max_count = max(tier_counts.values())
    assert (max_count - min_count) <= 20
```

---

### AC-001-3: 티어별 목록 문서 생성됨

**Given**: 10개 티어에 127개 스킬이 할당됨
**When**: 티어별 목록 생성 스크립트를 실행함
**Then**:
- [ ] `.moai/reports/skill-tiers-2025-11-22.md` 파일이 생성됨
- [ ] 파일 내용에 10개 티어별 표가 포함됨
- [ ] 각 티어의 총 스킬 수가 표시됨

**검증 방법**:
```python
def test_tier_list_document_generated():
    tier_report_path = ".moai/reports/skill-tiers-2025-11-22.md"
    assert os.path.exists(tier_report_path)

    with open(tier_report_path, "r") as f:
        content = f.read()

    # 10개 티어 섹션 존재 검증
    for tier_num in range(1, 11):
        assert f"Tier {tier_num}" in content

    # 총 스킬 수 표시 검증
    assert "총 127개" in content or "Total: 127" in content
```

---

## REQ-002: 중복 스킬 병합 (134 → 127개)

### AC-002-1: 중복 스킬이 0개로 축소됨

**Given**: 초기 134개 스킬 중 7개가 중복 대상임
**When**: 중복 병합 스크립트를 실행함
**Then**:
- [ ] 총 스킬 수가 127개로 감소함
- [ ] `moai-docs-generation`, `moai-docs-linting` 등 원본 스킬이 `archived/`로 이동됨
- [ ] 병합된 스킬(`moai-docs-toolkit`, `moai-docs-validation`)이 원본 기능을 모두 포함함

**검증 방법**:
```python
def test_duplicate_skills_merged():
    skills = load_all_skills()
    assert len(skills) == 127, f"Expected 127 skills, got {len(skills)}"

    # 원본 스킬이 archived로 이동되었는지 확인
    archived_skills = [
        "moai-docs-generation",
        "moai-docs-linting",
        # ... 기타 중복 스킬
    ]
    for archived_skill in archived_skills:
        archived_path = f".claude/skills/archived/{archived_skill}"
        assert os.path.exists(archived_path), f"{archived_skill} not archived"
```

---

### AC-002-2: 병합된 스킬이 원본 기능을 모두 포함함

**Given**: `moai-docs-generation`과 `moai-docs-toolkit`이 병합됨
**When**: 병합된 `moai-docs-toolkit` 스킬을 확인함
**Then**:
- [ ] `moai-docs-toolkit/SKILL.md`에 원본 두 스킬의 모든 기능이 설명됨
- [ ] 모듈화됨: `modules/generation.md`, `modules/validation.md` 존재
- [ ] 테스트 커버리지 100% 유지 (원본 기능 모두 테스트됨)

**검증 방법**:
```python
def test_merged_skill_complete():
    merged_skill = load_skill("moai-docs-toolkit")

    # 원본 기능 키워드 포함 확인
    content = merged_skill.read_skill_md()
    assert "generation" in content.lower()
    assert "validation" in content.lower()

    # 모듈 파일 존재 확인
    assert os.path.exists(".claude/skills/moai-docs-toolkit/modules/generation.md")
    assert os.path.exists(".claude/skills/moai-docs-toolkit/modules/validation.md")
```

---

## REQ-003: 비표준 명명 수정 (moai-google-nano-banana)

### AC-003-1: 스킬 이름이 표준 규칙을 준수함

**Given**: `moai-domain-nano-banana` 스킬이 비표준 명명임
**When**: 스킬 이름을 `moai-google-nano-banana`로 변경함
**Then**:
- [ ] 디렉토리 이름이 `moai-google-nano-banana`로 변경됨
- [ ] SKILL.md 내 `name` 필드가 `moai-google-nano-banana`로 업데이트됨
- [ ] 모든 스킬 이름이 Claude Code 표준을 준수함 (소문자, 숫자, 하이픈, 최대 64글자)

**검증 방법**:
```python
def test_skill_naming_compliance():
    skills = load_all_skills()
    for skill in skills:
        name = skill.metadata["name"]

        # 소문자, 숫자, 하이픈만 허용
        assert re.match(r'^[a-z0-9-]+$', name), f"Invalid name: {name}"

        # 최대 64글자
        assert len(name) <= 64, f"Name too long: {name} ({len(name)} chars)"

        # 예약어 금지
        assert "anthropic" not in name and "claude" not in name.split("-")
```

---

### AC-003-2: 하위 호환성이 유지됨

**Given**: 스킬 이름이 `moai-domain-nano-banana`에서 `moai-google-nano-banana`로 변경됨
**When**: 기존 에이전트가 원래 이름을 참조함
**Then**:
- [ ] Alias가 생성되어 원래 이름으로도 접근 가능함
- [ ] Migration guide가 작성되어 변경 사항이 문서화됨
- [ ] 35개 에이전트 중 해당 스킬을 참조하는 에이전트가 모두 정상 작동함

**검증 방법**:
```python
def test_renamed_skill_backward_compatibility():
    # Alias 존재 확인
    alias_path = ".claude/skills/moai-domain-nano-banana"
    assert os.path.islink(alias_path) or os.path.exists(alias_path)

    # Migration guide 존재 확인
    migration_guide = ".moai/docs/skill-migration-guide.md"
    assert os.path.exists(migration_guide)

    with open(migration_guide, "r") as f:
        content = f.read()
        assert "moai-domain-nano-banana" in content
        assert "moai-google-nano-banana" in content
```

---

## REQ-004: 메타데이터 표준화 (127개 스킬)

### AC-004-1: 모든 스킬에 필수 필드 7개가 존재함

**Given**: 총 127개 스킬이 존재함
**When**: 메타데이터 검증 스크립트를 실행함
**Then**:
- [ ] 127개 모든 스킬에 다음 필드가 존재함:
  - `name`
  - `description`
  - `version`
  - `modularized`
  - `allowed-tools` (선택사항이지만 권장)
  - `last_updated`
  - `compliance_score`

**검증 방법**:
```python
def test_all_skills_have_required_metadata():
    skills = load_all_skills()
    required_fields = ["name", "description", "version", "modularized", "last_updated", "compliance_score"]

    for skill in skills:
        for field in required_fields:
            assert field in skill.metadata, f"Skill {skill.name} missing field: {field}"
```

---

### AC-004-2: 버전 필드가 Semantic Versioning 형식임

**Given**: 모든 스킬에 `version` 필드가 존재함
**When**: 버전 형식을 검증함
**Then**:
- [ ] 모든 `version` 값이 `X.Y.Z` 형식임 (예: `1.0.0`, `3.2.1`)
- [ ] 잘못된 형식 (예: `v1.0`, `1.0`, `latest`) 0개

**검증 방법**:
```python
def test_version_semantic_format():
    skills = load_all_skills()
    semantic_version_pattern = r'^\d+\.\d+\.\d+$'

    for skill in skills:
        version = skill.metadata.get("version")
        assert re.match(semantic_version_pattern, version), \
               f"Invalid version format: {skill.name} = {version}"
```

---

### AC-004-3: 설명 길이가 100-200글자 범위임

**Given**: 모든 스킬에 `description` 필드가 존재함
**When**: 설명 길이를 측정함
**Then**:
- [ ] 최소 60% 스킬의 설명이 100-200글자 범위 내에 있음
- [ ] 100글자 미만 스킬이 20% 이하로 감소함 (현재 31%)
- [ ] 300글자 초과 스킬이 0개로 감소함 (현재 6%)

**검증 방법**:
```python
def test_description_length():
    skills = load_all_skills()
    optimal_count = 0
    too_short_count = 0
    too_long_count = 0

    for skill in skills:
        desc_length = len(skill.metadata.get("description", ""))

        if 100 <= desc_length <= 200:
            optimal_count += 1
        elif desc_length < 100:
            too_short_count += 1
        elif desc_length > 300:
            too_long_count += 1

    # 최적 비율 60% 이상
    optimal_ratio = optimal_count / len(skills)
    assert optimal_ratio >= 0.60, f"Optimal ratio: {optimal_ratio:.2%} (expected ≥60%)"

    # 짧은 설명 20% 이하
    too_short_ratio = too_short_count / len(skills)
    assert too_short_ratio <= 0.20, f"Too short: {too_short_ratio:.2%} (expected ≤20%)"

    # 긴 설명 0%
    assert too_long_count == 0, f"Too long: {too_long_count} skills"
```

---

## REQ-005: 신규 필수 스킬 5개 추가

### AC-005-1: 5개 신규 스킬이 생성됨

**Given**: 다음 5개 스킬이 필요함:
- moai-core-code-templates
- moai-security-api-versioning
- moai-essentials-testing-integration
- moai-essentials-performance-profiling
- moai-security-accessibility-wcag3

**When**: 신규 스킬 생성 스크립트를 실행함
**Then**:
- [ ] 5개 모든 스킬 디렉토리가 `.claude/skills/`에 생성됨
- [ ] 각 스킬에 `SKILL.md` 파일이 존재함
- [ ] 각 스킬에 필수 메타데이터 7개 필드가 포함됨

**검증 방법**:
```python
def test_new_skills_created():
    new_skills = [
        "moai-core-code-templates",
        "moai-security-api-versioning",
        "moai-essentials-testing-integration",
        "moai-essentials-performance-profiling",
        "moai-security-accessibility-wcag3"
    ]

    for skill_name in new_skills:
        skill_path = f".claude/skills/{skill_name}/SKILL.md"
        assert os.path.exists(skill_path), f"Skill not found: {skill_name}"

        skill = load_skill(skill_name)
        assert "name" in skill.metadata
        assert "description" in skill.metadata
        assert "version" in skill.metadata
```

---

### AC-005-2: 신규 스킬이 모듈화되어 있음

**Given**: 5개 신규 스킬이 생성됨
**When**: 각 스킬의 파일 구조를 확인함
**Then**:
- [ ] 각 스킬에 `modules/` 디렉토리가 존재함
- [ ] 각 스킬의 `modularized` 필드가 `true`임
- [ ] 각 스킬에 `examples.md` 파일이 존재함

**검증 방법**:
```python
def test_new_skills_modularized():
    new_skills = [
        "moai-core-code-templates",
        "moai-security-api-versioning",
        "moai-essentials-testing-integration",
        "moai-essentials-performance-profiling",
        "moai-security-accessibility-wcag3"
    ]

    for skill_name in new_skills:
        modules_path = f".claude/skills/{skill_name}/modules"
        assert os.path.exists(modules_path), f"Modules missing: {skill_name}"

        skill = load_skill(skill_name)
        assert skill.metadata.get("modularized") == True

        examples_path = f".claude/skills/{skill_name}/examples.md"
        assert os.path.exists(examples_path), f"Examples missing: {skill_name}"
```

---

### AC-005-3: 신규 스킬이 에이전트와 통합됨

**Given**: 5개 신규 스킬이 생성됨
**When**: 에이전트 파일을 확인함
**Then**:
- [ ] 각 신규 스킬이 최소 1개 이상의 에이전트에 매핑됨
- [ ] `agent_coverage` 필드에 참조 에이전트 목록이 포함됨
- [ ] `.moai/memory/agents.md` 파일에 신규 스킬 참조가 추가됨

**검증 방법**:
```python
def test_new_skills_integration_with_agents():
    new_skills = [
        "moai-core-code-templates",
        "moai-security-api-versioning",
        "moai-essentials-testing-integration",
        "moai-essentials-performance-profiling",
        "moai-security-accessibility-wcag3"
    ]

    for skill_name in new_skills:
        skill = load_skill(skill_name)
        agent_coverage = skill.metadata.get("agent_coverage", [])
        assert len(agent_coverage) >= 1, f"No agent mapped to {skill_name}"

    # agents.md에 참조 확인
    with open(".moai/memory/agents.md", "r") as f:
        agents_content = f.read()

    for skill_name in new_skills:
        assert skill_name in agents_content, f"{skill_name} not in agents.md"
```

---

## REQ-006: Auto-Trigger 로직 구현

### AC-006-1: CLAUDE.md에 Auto-Trigger 섹션이 추가됨

**Given**: CLAUDE.md 파일이 존재함
**When**: Auto-Trigger 로직 섹션을 추가함
**Then**:
- [ ] CLAUDE.md 파일 내 "Rule 8: Config 기반 자동 동작" 섹션에 Auto-Trigger 로직이 포함됨
- [ ] 키워드 매핑 테이블이 존재함 (최소 20개 키워드-스킬 쌍)
- [ ] 각 키워드에 대응하는 스킬과 에이전트가 명시됨

**검증 방법**:
```python
def test_claude_md_auto_trigger_section():
    with open("CLAUDE.md", "r") as f:
        content = f.read()

    assert "Auto-Trigger 로직" in content or "Auto-Trigger Logic" in content
    assert "키워드" in content or "keyword" in content.lower()
    assert "스킬" in content or "skill" in content.lower()

    # 최소 20개 키워드-스킬 쌍 검증
    keyword_skill_pairs = re.findall(r'\|.*\|.*\|.*\|', content)
    assert len(keyword_skill_pairs) >= 20
```

---

### AC-006-2: 모든 스킬에 auto_trigger_keywords 필드가 존재함

**Given**: 총 127개 스킬이 존재함
**When**: 메타데이터를 확인함
**Then**:
- [ ] 127개 모든 스킬에 `auto_trigger_keywords` 배열이 존재함
- [ ] 각 배열에 최소 1개 이상의 키워드가 포함됨
- [ ] 키워드가 소문자로 정규화되어 있음

**검증 방법**:
```python
def test_auto_trigger_keywords_in_all_skills():
    skills = load_all_skills()

    for skill in skills:
        keywords = skill.metadata.get("auto_trigger_keywords", [])
        assert len(keywords) >= 1, f"No keywords in {skill.name}"

        for keyword in keywords:
            assert keyword == keyword.lower(), f"Keyword not lowercase: {keyword}"
```

---

### AC-006-3: 키워드 매칭 정확도가 95% 이상임

**Given**: Auto-Trigger 로직이 구현됨
**When**: 100개 테스트 케이스를 실행함
**Then**:
- [ ] 키워드 매칭 정확도가 95% 이상임 (95개 이상 성공)
- [ ] 잘못된 스킬 선택이 5% 이하임
- [ ] Fallback 메커니즘이 작동함 (매칭 실패 시 사용자 선택 프롬프트)

**검증 방법**:
```python
def test_auto_trigger_keyword_matching():
    test_cases = [
        ("Implement JWT authentication", "moai-security-auth"),
        ("Optimize Python FastAPI performance", "moai-essentials-perf"),
        ("Create SPEC for user profile", "moai-foundation-specs"),
        # ... 97개 추가 테스트 케이스
    ]

    correct_matches = 0
    for user_request, expected_skill in test_cases:
        selected_skill = auto_trigger_logic(user_request)
        if selected_skill == expected_skill:
            correct_matches += 1

    accuracy = correct_matches / len(test_cases)
    assert accuracy >= 0.95, f"Accuracy: {accuracy:.2%} (expected ≥95%)"
```

---

## REQ-007: Agent-Skill 커버리지 85% 달성

### AC-007-1: 85% 이상의 에이전트가 스킬을 참조함

**Given**: 총 35개 에이전트가 존재함
**When**: 각 에이전트의 스킬 참조를 확인함
**Then**:
- [ ] 최소 30개 이상의 에이전트가 스킬을 참조함 (85% = 29.75개)
- [ ] 각 에이전트의 스킬 참조가 명시적으로 문서화됨 (agents.md 또는 에이전트 파일)

**검증 방법**:
```python
def test_agent_skill_coverage():
    agents = load_all_agents()  # 35개 에이전트
    agents_with_skills = 0

    for agent in agents:
        if agent.has_skill_references():
            agents_with_skills += 1

    coverage = agents_with_skills / len(agents)
    assert coverage >= 0.85, f"Coverage: {coverage:.2%} (expected ≥85%)"
    assert agents_with_skills >= 30
```

---

### AC-007-2: 참조된 스킬이 실제로 존재함

**Given**: 30개 이상의 에이전트가 스킬을 참조함
**When**: 참조된 스킬의 존재를 확인함
**Then**:
- [ ] 모든 참조된 스킬 이름이 실제 스킬 디렉토리와 일치함
- [ ] 깨진 참조 0개 (존재하지 않는 스킬 참조 없음)

**검증 방법**:
```python
def test_agent_skill_references_valid():
    agents = load_all_agents()
    all_skills = load_all_skills()
    skill_names = {skill.metadata["name"] for skill in all_skills}

    broken_references = []

    for agent in agents:
        referenced_skills = agent.get_skill_references()
        for skill_name in referenced_skills:
            if skill_name not in skill_names:
                broken_references.append((agent.name, skill_name))

    assert len(broken_references) == 0, f"Broken references: {broken_references}"
```

---

### AC-007-3: agent_coverage 필드가 모든 스킬에 존재함

**Given**: 총 127개 스킬이 존재함
**When**: 메타데이터를 확인함
**Then**:
- [ ] 127개 모든 스킬에 `agent_coverage` 배열이 존재함
- [ ] 각 배열에 최소 0개 이상의 에이전트가 포함됨 (일부 스킬은 0개 가능)
- [ ] 평균적으로 스킬당 2-3개 에이전트 참조가 있음

**검증 방법**:
```python
def test_agent_coverage_field_in_all_skills():
    skills = load_all_skills()

    total_agent_references = 0

    for skill in skills:
        agent_coverage = skill.metadata.get("agent_coverage", [])
        assert isinstance(agent_coverage, list), f"Invalid agent_coverage in {skill.name}"
        total_agent_references += len(agent_coverage)

    avg_references = total_agent_references / len(skills)
    assert 2 <= avg_references <= 3, f"Avg references: {avg_references:.2f} (expected 2-3)"
```

---

## 통합 품질 게이트 (Overall Quality Gates)

### QG-1: Overall Compliance Score ≥95%

**Given**: 모든 7개 REQ가 구현됨
**When**: 전체 준수율을 계산함
**Then**:
- [ ] Overall Compliance Score가 95% 이상임 (현재 72%)
- [ ] 모든 스킬이 Claude Code 공식 표준을 준수함

**검증 방법**:
```python
def test_overall_compliance_score():
    skills = load_all_skills()

    compliance_scores = [skill.metadata.get("compliance_score", 0) for skill in skills]
    overall_score = sum(compliance_scores) / len(skills)

    assert overall_score >= 95, f"Overall compliance: {overall_score:.2f}% (expected ≥95%)"
```

---

### QG-2: 모든 테스트 케이스 통과

**Given**: 총 25개 테스트 케이스가 작성됨
**When**: pytest를 실행함
**Then**:
- [ ] 25개 모든 테스트 케이스가 통과함 (100% pass rate)
- [ ] 테스트 커버리지 ≥90%

**검증 방법**:
```bash
pytest tests/test_spec_skill_portfolio_opt_001.py -v --cov=.claude/skills --cov-report=term-missing
```

**예상 출력**:
```
========== 25 passed in 10.25s ==========
Coverage: 92%
```

---

### QG-3: 문서화 완료

**Given**: 모든 구현이 완료됨
**When**: 문서를 확인함
**Then**:
- [ ] `.moai/docs/skill-migration-guide.md` 존재
- [ ] `.moai/reports/skill-tiers-2025-11-22.md` 존재
- [ ] CLAUDE.md에 Auto-Trigger 로직 섹션 추가됨
- [ ] 모든 신규 스킬에 examples.md 파일 존재

**검증 방법**:
```python
def test_documentation_complete():
    required_docs = [
        ".moai/docs/skill-migration-guide.md",
        ".moai/reports/skill-tiers-2025-11-22.md",
        "CLAUDE.md"
    ]

    for doc in required_docs:
        assert os.path.exists(doc), f"Missing doc: {doc}"

    new_skills = [
        "moai-core-code-templates",
        "moai-security-api-versioning",
        "moai-essentials-testing-integration",
        "moai-essentials-performance-profiling",
        "moai-security-accessibility-wcag3"
    ]

    for skill_name in new_skills:
        examples_path = f".claude/skills/{skill_name}/examples.md"
        assert os.path.exists(examples_path), f"Missing examples: {skill_name}"
```

---

## Definition of Done (DoD)

### Phase 1 DoD
- [ ] 모든 AC-001, AC-002, AC-003, AC-004 테스트 통과
- [ ] 메타데이터 표준화 완료 (127개 스킬)
- [ ] 중복 스킬 병합 완료 (134 → 127개)
- [ ] 비표준 명명 수정 완료

### Phase 2 DoD
- [ ] 모든 AC-001 테스트 통과
- [ ] 10개 티어 정의 및 할당 완료
- [ ] 티어별 목록 문서 생성 완료

### Phase 3 DoD
- [ ] 파일 크기 초과 스킬 0개
- [ ] 모듈화 완성도 ≥30%
- [ ] Progressive Disclosure 일관성 검증 완료

### Phase 4 DoD
- [ ] 모든 AC-005, AC-006, AC-007 테스트 통과
- [ ] 신규 스킬 5개 생성 및 에이전트 통합 완료
- [ ] Auto-Trigger 로직 CLAUDE.md 통합 완료
- [ ] Agent-Skill 커버리지 ≥85%

### Overall DoD
- [ ] 모든 25개 테스트 케이스 통과 (100% pass rate)
- [ ] Overall Compliance Score ≥95%
- [ ] 모든 문서화 완료
- [ ] Git 커밋 완료 (Phase별 atomic commit)

---

**Last Updated**: 2025-11-22
**Version**: 1.0.0
**Test Coverage Target**: ≥90%
