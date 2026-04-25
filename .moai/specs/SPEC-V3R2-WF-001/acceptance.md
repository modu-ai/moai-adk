---
spec_id: SPEC-V3R2-WF-001
title: Acceptance Criteria — Skill Consolidation (48 → 24)
version: "0.1.0"
status: draft
created: 2026-04-24
updated: 2026-04-24
author: manager-spec (acceptance.md generation)
related_plan: .moai/specs/SPEC-V3R2-WF-001/plan.md
related_spec: .moai/specs/SPEC-V3R2-WF-001/spec.md
---

# 수용 기준 — SPEC-V3R2-WF-001 Skill Consolidation (48 → 24)

> **형식**: 각 AC 는 Given-When-Then (GWT) 로 서술되며, 자동 검증을 위한 구체 `grep`/`ls`/`diff` 명령과 수동 관찰 방법을 포함한다.
> **참조**: spec.md §6 (AC-WF001-01 ~ 15), plan.md §9 Definition of Done, tasks.md 의 각 Checkpoint.

---

## AC-WF001-01: v3 skill directory count = 24

**Maps**: REQ-WF001-001

**Given** v2.13.2 skill tree 가 `.claude/skills/` 에 48 디렉터리를 보유한 상태
**When** SPEC-V3R2-WF-001 의 모든 7 Wave 가 머지되고 `make build` 가 실행됨
**Then** `.claude/skills/` 의 디렉터리 수는 정확히 **24** 이어야 한다

### Verification

```bash
# Automated check
actual=$(ls -d /Users/goos/MoAI/moai-adk-go/.claude/skills/*/ | wc -l | tr -d ' ')
[ "$actual" = "24" ] || { echo "FAIL: expected 24, got $actual"; exit 1; }

# Template parity
template=$(ls -d /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/*/ | wc -l | tr -d ' ')
[ "$template" = "24" ] || { echo "FAIL: template mismatch: $template"; exit 1; }
```

### Evidence
- Wave 1.7 `wave-1.7-report.md` 의 "Skill count" 섹션
- `ls -d .claude/skills/*/` 출력 = 24 라인

---

## AC-WF001-02: 48 entries × single verdict

**Maps**: REQ-WF001-002

**Given** R4 audit 의 `Per-skill audit table` 이 48 entry 를 정의한 상태
**When** §6.2 판정표 (spec.md lines 225-274) 와 `.moai/decisions/skill-rename-map.yaml` 이 대조됨
**Then** 48 entry 전부가 **정확히 하나의** {KEEP, REFACTOR, MERGE, RETIRE} verdict 를 가져야 한다

### Verification

```bash
# Count §6.2 table rows (exclude header + separator)
rows=$(awk '/^\| [0-9]+ \|/' /Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-V3R2-WF-001/spec.md | wc -l | tr -d ' ')
[ "$rows" = "48" ] || { echo "FAIL: §6.2 has $rows rows"; exit 1; }

# Each row has "R4 verdict" and "v3 action" columns (verdict ∈ {KEEP, REFACTOR, MERGE, RETIRE})
awk '/^\| [0-9]+ \|/' /Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-V3R2-WF-001/spec.md | \
  awk -F'|' '{gsub(/^ +| +$/, "", $3); print $3}' | \
  grep -vE '^(KEEP|REFACTOR|MERGE|RETIRE)' | wc -l  # must be 0
```

### Evidence
- §6.2 판정표 라인별 column 3 (R4 verdict) 이 정규식 `^(KEEP|REFACTOR|MERGE|RETIRE)` 일치
- `skill-rename-map.yaml` 의 merges/retires/refactors/unchanged_keep 합계 = 48 (또는 entry 일부가 여러 카테고리 겹쳐도 48 unique 이름)

---

## AC-WF001-03: Thinking triplet trigger union

**Maps**: REQ-WF001-007

**Given** `moai-foundation-thinking`, `moai-foundation-philosopher`, `moai-workflow-thinking` 3개 skill 이 Wave 시작 시 독립 존재
**When** Wave 1.2 (content merge) + Wave 1.3 (trigger union) 가 완료됨
**Then** 병합된 `moai-foundation-thinking/SKILL.md` 의 `triggers:` 에 3개 source 의 모든 trigger 가 union 으로 포함되어야 한다

### Verification

```bash
# Extract source triggers (before archive)
philosopher_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .moai/archive/skills/v3.0/moai-foundation-philosopher/SKILL.md | grep "^  - " | sort -u)
workflow_thinking_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .moai/archive/skills/v3.0/moai-workflow-thinking/SKILL.md | grep "^  - " | sort -u)

# Extract target triggers
target_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .claude/skills/moai-foundation-thinking/SKILL.md | grep "^  - " | sort -u)

# Verify source ⊆ target (case-insensitive)
comm -23 <(echo "$philosopher_triggers" | tr '[:upper:]' '[:lower:]') <(echo "$target_triggers" | tr '[:upper:]' '[:lower:]')  # must be empty
comm -23 <(echo "$workflow_thinking_triggers" | tr '[:upper:]' '[:lower:]') <(echo "$target_triggers" | tr '[:upper:]' '[:lower:]')  # must be empty
```

### Evidence
- `moai-foundation-thinking/SKILL.md` frontmatter 의 triggers 수가 3 sources 의 union (dedup 후) 과 일치
- Wave 1.3 task T1.3-1 의 output log

---

## AC-WF001-04: RETIRE archive with RETIRED.md

**Maps**: REQ-WF001-008

**Given** `moai-tool-svg` 가 v2.13.2 에 존재한 상태 (§6.2 line 268: RETIRE)
**When** Wave 1.4 T1.4-11 이 실행됨
**Then** `.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md` 가 존재하고 substitute guidance 를 포함해야 한다

### Verification

```bash
[ -f /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md ] || { echo "FAIL: archive missing"; exit 1; }

# RETIRED.md content check
grep -q "^\*\*Date\*\*:" /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md
grep -q "^\*\*Verdict\*\*:" /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md
grep -q "^\*\*Substitute\*\*:" /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md
grep -q "^\*\*SPEC\*\*: SPEC-V3R2-WF-001" /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-tool-svg/RETIRED.md

# Original SKILL.md preserved
[ -f /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-tool-svg/SKILL.md ]
```

### Evidence
- `RETIRED.md` 파일 존재 및 6개 frontmatter 필드 (Date, Verdict, Substitute, SPEC, Migration, Rationale) 완비
- 원본 `SKILL.md` 도 archive 하위에 보존됨

---

## AC-WF001-05: Template + local byte-identical

**Maps**: REQ-WF001-006

**Given** 모든 Wave 가 완료된 상태
**When** `diff -rq .claude/skills internal/template/templates/.claude/skills` 를 실행함
**Then** 출력이 **empty** 이어야 한다 (즉, 두 트리가 디렉터리 이름 및 파일 내용 모두 동일)

### Verification

```bash
result=$(diff -rq /Users/goos/MoAI/moai-adk-go/.claude/skills \
                  /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills 2>&1)
[ -z "$result" ] || { echo "FAIL: $result"; exit 1; }
```

### Evidence
- Wave 1.2, 1.3, 1.4, 1.5 각 checkpoint 에서 `diff -rq` empty 확인
- Wave 1.7 T1.7-2 의 최종 확인

### Note
Archive (`.moai/archive/skills/v3.0/`) 는 **template 에 없으므로** `.claude/skills/` 및 `internal/template/templates/.claude/skills/` 비교 범위 외. OQ-2 해소 방침에 따라 archive 는 dev-only artifact.

---

## AC-WF001-06: Agency FROZEN skill 불변

**Maps**: REQ-WF001-005

**Given** `moai-domain-copywriting`, `moai-domain-brand-design` 이 Wave 시작 시 특정 해시로 baseline 기록된 상태 (T1.1-1 baseline)
**When** 모든 Wave 가 완료된 후 동일 파일의 해시를 다시 계산
**Then** 해시가 baseline 과 **정확히 일치** 해야 한다 (byte-identical, agency 계약 위반 없음)

### Verification

```bash
# Wave 1.1 baseline recorded
baseline=$(cat .moai/specs/SPEC-V3R2-WF-001/baseline-hashes.txt | grep "moai-domain-copywriting\|moai-domain-brand-design")

# Current hashes
current=$(shasum -a 256 \
  /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-domain-copywriting/SKILL.md \
  /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-domain-brand-design/SKILL.md)

diff <(echo "$baseline" | awk '{print $1}') <(echo "$current" | awk '{print $1}')  # must be empty
```

### Evidence
- Wave 1.7 T1.7-3 의 hash comparison log
- `.claude/rules/moai/design/constitution.md` §3 FROZEN zone 미변경 증명

---

## AC-WF001-07: Agent prompt retired-name 치환

**Maps**: REQ-WF001-014

**Given** v2 agent prompt 가 `moai-foundation-philosopher`, `moai-workflow-thinking`, 등의 retired skill 이름을 문자열 리터럴로 하드코딩한 상태
**When** Wave 1.6 가 완료됨
**Then** 동일 agent prompt 파일에 retired skill 이름이 **0 occurrence** 이고, 매핑된 new target 이름으로 대체되어 있어야 한다

### Verification

```bash
# Retired names set
retired_names="moai-foundation-philosopher|moai-workflow-thinking|moai-design-craft|moai-design-tools|moai-domain-uiux|moai-platform-database-cloud|moai-workflow-templates|moai-docs-generation|moai-workflow-jit-docs|moai-tool-svg|moai-foundation-context"

# Grep in agents (local)
count=$(grep -rl "$retired_names" /Users/goos/MoAI/moai-adk-go/.claude/agents/ 2>/dev/null | wc -l | tr -d ' ')
[ "$count" = "0" ] || { echo "FAIL: $count agent files still reference retired names"; exit 1; }

# Grep in agents (template)
count=$(grep -rl "$retired_names" /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/ 2>/dev/null | wc -l | tr -d ' ')
[ "$count" = "0" ] || { echo "FAIL: $count template agent files still reference retired names"; exit 1; }

# Target names are present (sanity check)
grep -l "moai-foundation-thinking" /Users/goos/MoAI/moai-adk-go/.claude/agents/moai/*.md | wc -l  # ≥ 1
```

### Evidence
- Wave 1.6 checkpoint T1.6-END 의 grep log
- 4개 수정된 agent 파일: `expert-frontend.md`, `manager-project.md`, `builder-skill.md`, `manager-docs.md`

---

## AC-WF001-08: Archive 없이 삭제 시 CI reject

**Maps**: REQ-WF001-015 (Unwanted Behavior)

**Given** 가상 시나리오: skill 디렉터리 삭제가 commit 에 포함되었으나 `.moai/archive/skills/v3.0/<name>/` 가 없는 경우
**When** CI guard rail `SKILL_RETIRE_NO_ARCHIVE` 가 실행됨
**Then** commit 이 reject 되어야 한다

### Verification

**본 SPEC 은 dry-run 검증만 포함 (OQ-6)**. 실제 CI 구현은 follow-up SPEC 대상.

```bash
# Dry-run logic (Wave 1.7 manual check)
for archived in $(ls /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/ 2>/dev/null); do
  [ -f "/Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/$archived/RETIRED.md" ] || {
    echo "FAIL: $archived has no RETIRED.md"; exit 1;
  }
done
```

### Evidence
- Wave 1.7 T1.7-5 의 archive completeness log
- Follow-up SPEC reference (`SKILL_RETIRE_NO_ARCHIVE` hook 구현)

---

## AC-WF001-09: templates/docs-generation bundled resource 이관

**Maps**: REQ-WF001-010

**Given** `moai-workflow-templates/schemas/` 및 `moai-workflow-templates/templates/` (또는 `modules/`, `references/`) 가 bundled 상태로 존재
**When** Wave 1.2 T1.2-2 가 완료됨
**Then** 해당 bundled 리소스가 `moai-workflow-project/modules/` 또는 `moai-workflow-project/references/` 하위로 재배치되고 원본 위치는 Wave 1.4 에서 archive 됨

### Verification

```bash
# Post-Wave 1.4: originals archived
[ -d /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-workflow-templates ]
[ -d /Users/goos/MoAI/moai-adk-go/.moai/archive/skills/v3.0/moai-docs-generation ]

# Target has absorbed content
[ -d /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-project/modules ]
[ -d /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-project/references ]

# Specific resource check (예시, actual files depend on source content)
# ls /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-project/modules/ | wc -l  ≥ source count
```

### Evidence
- `moai-workflow-project/modules/` 에 source 의 reusable 자원이 포함됨
- archive 에 원본 보존 (Wave 1.4 archived SKILL.md)

---

## AC-WF001-10: moai-foundation-thinking 내부 섹션 포함

**Maps**: REQ-WF001-003

**Given** Wave 1.2 T1.2-1 이 완료된 상태
**When** `moai-foundation-thinking/SKILL.md` 가 inspect 됨
**Then** 파일은 다음을 포함해야 한다:
  - `## First Principles` 또는 동등 섹션 (philosopher 흡수)
  - `## Sequential Thinking MCP` 또는 동등 섹션 (workflow-thinking 흡수)

### Verification

```bash
grep -q "First Principles" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-foundation-thinking/SKILL.md
grep -q "Sequential Thinking" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-foundation-thinking/SKILL.md

# Level 2 token budget (5000 ceiling)
word_count=$(wc -w < /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-foundation-thinking/SKILL.md)
estimated_tokens=$(( word_count * 3 / 4 ))  # rough estimate: 0.75 token/word
[ "$estimated_tokens" -le 5000 ] || echo "WARN: Level 2 token estimate $estimated_tokens exceeds 5000"
```

### Evidence
- SKILL.md grep 일치
- Wave 1.2 T1.2-1 의 token budget log

---

## AC-WF001-11: moai-design-system trigger union

**Maps**: REQ-WF001-007

**Given** `moai-design-craft`, `moai-domain-uiux`, `moai-design-tools` 3개 source (UI side)
**When** Wave 1.3 T1.3-3 이 완료됨
**Then** `moai-design-system/SKILL.md` 의 frontmatter `triggers:` 는 3개 source 의 trigger union 이어야 한다

### Verification

```bash
# Extract source triggers (from archive after Wave 1.4)
craft_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .moai/archive/skills/v3.0/moai-design-craft/SKILL.md | grep "^  - " | sort -u)
uiux_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .moai/archive/skills/v3.0/moai-domain-uiux/SKILL.md | grep "^  - " | sort -u)
tools_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .moai/archive/skills/v3.0/moai-design-tools/SKILL.md | grep "^  - " | sort -u)

# Target
target_triggers=$(awk '/^triggers:/,/^[a-z_-]+:/' .claude/skills/moai-design-system/SKILL.md | grep "^  - " | sort -u)

# Union subset check (craft ∪ uiux ⊆ target); tools Pencil portion only
comm -23 <(echo "$craft_triggers" | tr '[:upper:]' '[:lower:]') <(echo "$target_triggers" | tr '[:upper:]' '[:lower:]') | wc -l  # 0
comm -23 <(echo "$uiux_triggers" | tr '[:upper:]' '[:lower:]') <(echo "$target_triggers" | tr '[:upper:]' '[:lower:]') | wc -l  # 0
```

### Evidence
- `moai-design-system/SKILL.md` frontmatter inspect log
- Wave 1.3 T1.3-3 결과

---

## AC-WF001-12: moai-domain-database cloud vendor 섹션

**Maps**: REQ-WF001-003

**Given** `moai-platform-database-cloud/SKILL.md` 의 cloud vendor content
**When** Wave 1.2 T1.2-4 완료 후 `moai-domain-database/SKILL.md` 를 inspect
**Then** `## Cloud Vendor Guide (absorbed from moai-platform-database-cloud)` 또는 동등 섹션이 존재해야 한다

### Verification

```bash
grep -q "Cloud Vendor" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-domain-database/SKILL.md
grep -q "absorbed from moai-platform-database-cloud" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-domain-database/SKILL.md

# References 이관
[ -d /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-domain-database/references ]
```

### Evidence
- SKILL.md grep 일치
- `references/` 디렉터리 존재 (cloud reference 파일 포함)

---

## AC-WF001-13: UNCLEAR skill telemetry window

**Maps**: REQ-WF001-013

**Given** `moai-framework-electron`, `moai-platform-chrome-extension` 이 R4 audit 에서 UNCLEAR 분류
**When** Wave 1.5 T1.5-7, T1.5-8 완료
**Then** 두 skill 은 v3 tree 에 **존속** 하고, SKILL.md 에 `## Telemetry Window` 섹션이 추가되어 있어야 한다

### Verification

```bash
# 존속 확인
[ -d /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-framework-electron ]
[ -d /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-platform-chrome-extension ]

# Telemetry Window 섹션 존재
grep -q "## Telemetry Window" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-framework-electron/SKILL.md
grep -q "## Telemetry Window" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-platform-chrome-extension/SKILL.md

# 60-day window 언급
grep -q "60-day\|60 day" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-framework-electron/SKILL.md
grep -q "60-day\|60 day" /Users/goos/MoAI/moai-adk-go/.claude/skills/moai-platform-chrome-extension/SKILL.md
```

### Evidence
- 2개 skill 존속 + Telemetry Window 섹션 grep 일치

---

## AC-WF001-14: REFACTOR 섹션 R4 audit line 링크

**Maps**: REQ-WF001-011

**Given** REFACTOR verdict 를 받은 skill (예: `moai-workflow-testing`)
**When** Wave 1.5 의 해당 task 가 완료됨
**Then** skill SKILL.md 에 `## Refactor Notes` 섹션이 존재하고, R4 audit 또는 §6.2 line reference 를 포함해야 한다

### Verification

```bash
for skill in moai-workflow-testing moai-domain-backend moai-domain-frontend moai-domain-database moai-platform-deployment moai-platform-auth; do
  grep -q "## Refactor Notes" /Users/goos/MoAI/moai-adk-go/.claude/skills/$skill/SKILL.md || {
    echo "FAIL: $skill has no Refactor Notes"; exit 1;
  }
  grep -q "R4 audit\|SPEC-V3R2-WF-001" /Users/goos/MoAI/moai-adk-go/.claude/skills/$skill/SKILL.md || {
    echo "FAIL: $skill Refactor Notes has no audit reference"; exit 1;
  }
done
```

### Evidence
- 6개 REFACTOR skill (archive 되지 않은 것) 각 SKILL.md grep 일치

---

## AC-WF001-15: related-skills 참조 자동 재작성

**Maps**: REQ-WF001-017 (Complex: State + Event)

**Given** 임의 skill 의 `related-skills:` 가 retired skill 이름 (예: `moai-foundation-philosopher`) 을 참조
**When** Wave 1.3 (frontmatter union) 또는 Wave 1.6 (agent prompt rewrite) 단계에서 §6.2 mapping 이 적용됨
**Then** retired 이름은 target 이름 (예: `moai-foundation-thinking`) 으로 rewrite 되거나 alias 로 명시되어 유지됨

### Verification

```bash
# Surviving skills 의 SKILL.md 에서 related-skills 가 retired 이름만 단독으로 참조하지 않음
# (alias 형태로 남는 것은 허용, but 단독 reference 는 없어야 함)
for skill_dir in /Users/goos/MoAI/moai-adk-go/.claude/skills/*/; do
  related=$(awk '/^related-skills:/,/^[a-z_-]+:/' $skill_dir/SKILL.md 2>/dev/null | grep "^  - " || true)
  # No retired name appears as sole reference (each retired name must co-exist with target)
  echo "$related" | grep -E "moai-(foundation-philosopher|workflow-thinking|design-craft|design-tools|domain-uiux|platform-database-cloud|workflow-templates|docs-generation|workflow-jit-docs|tool-svg|foundation-context)" || true
done

# Commands / rules 에서 직접 참조 제거
count=$(grep -rl "moai-foundation-philosopher\|moai-workflow-thinking\|moai-design-craft\|moai-domain-uiux\|moai-platform-database-cloud" \
  /Users/goos/MoAI/moai-adk-go/.claude/commands/ \
  /Users/goos/MoAI/moai-adk-go/.claude/rules/ 2>/dev/null | wc -l | tr -d ' ')
[ "$count" = "0" ] || { echo "FAIL: $count files still reference retired names"; exit 1; }
```

### Evidence
- `.claude/commands/`, `.claude/rules/` 내 retired 이름 참조 0
- `.claude/skills/` 내 retired 이름은 alias 형태로만 존재 (Wave 1.3 이 보존한 related-skills)

---

## Edge Cases (보조 검증)

### EC-1: `make build` idempotency

**Given** 모든 Wave 완료 후 `make build` 를 **2회** 연속 실행
**Then** 2차 실행의 `internal/template/embedded.go` 가 1차와 byte-identical

```bash
make build && shasum -a 256 internal/template/embedded.go > /tmp/first.sha
make build && shasum -a 256 internal/template/embedded.go > /tmp/second.sha
diff /tmp/first.sha /tmp/second.sha  # empty
```

### EC-2: Wave 역순 revert 복구

**Given** Wave 1.7 까지 완료된 상태
**When** Wave 1.7 → 1.6 → 1.5 → 1.4 → 1.3 → 1.2 순서로 `git revert` 를 연속 수행
**Then** 결과 상태가 v2.13.2 baseline (48 skills) 와 동일

```bash
# (hypothetical dry-run)
# git revert <sha-1.7>..<sha-1.2>
actual=$(ls -d .claude/skills/*/ | wc -l | tr -d ' ')
[ "$actual" = "48" ] || echo "FAIL: revert did not restore 48 skills"
```

### EC-3: Level 2 token budget 미초과

**Given** 4개 merge target 의 SKILL.md
**Then** 각 파일의 Level 2 body 가 5000 token (추정) 이하

```bash
for target in moai-foundation-thinking moai-workflow-project moai-design-system moai-domain-database; do
  wc=$(wc -w < /Users/goos/MoAI/moai-adk-go/.claude/skills/$target/SKILL.md)
  est_tokens=$(( wc * 3 / 4 ))
  echo "$target: $est_tokens tokens (limit 5000)"
  [ "$est_tokens" -le 5000 ] || echo "WARN: $target exceeds Level 2 budget"
done
```

### EC-4: skill-rename-map.yaml 스키마 검증

**Given** `.moai/decisions/skill-rename-map.yaml` 존재
**When** Python YAML parser 로 load
**Then** `version`, `generated_by`, `merges`, `retires`, `refactors`, `unchanged_keep` 필드 존재

```bash
python3 -c "
import yaml, sys
data = yaml.safe_load(open('/Users/goos/MoAI/moai-adk-go/.moai/decisions/skill-rename-map.yaml'))
assert 'version' in data
assert 'generated_by' in data and data['generated_by'] == 'SPEC-V3R2-WF-001'
assert 'merges' in data and isinstance(data['merges'], list)
assert 'retires' in data and isinstance(data['retires'], list)
assert 'unchanged_keep' in data and isinstance(data['unchanged_keep'], list)
print('OK')
" || { echo "FAIL: skill-rename-map.yaml schema invalid"; exit 1; }
```

---

## Definition of Done (DoD) — 최종 체크리스트

본 SPEC 은 다음 **모든** 조건이 만족될 때 완료로 간주된다:

- [ ] AC-WF001-01 PASS: `.claude/skills/` = 24 directories (template 동등)
- [ ] AC-WF001-02 PASS: §6.2 판정표 48 entry 전부에 verdict 존재
- [ ] AC-WF001-03 PASS: thinking triplet trigger union
- [ ] AC-WF001-04 PASS: `moai-tool-svg/RETIRED.md` 생성
- [ ] AC-WF001-05 PASS: `diff -rq` empty
- [ ] AC-WF001-06 PASS: agency FROZEN hash match
- [ ] AC-WF001-07 PASS: agent prompt 내 retired 이름 0 occurrence
- [ ] AC-WF001-08 PASS: archive-less deletion CI dry-run 통과
- [ ] AC-WF001-09 PASS: templates/docs-generation bundled resource 이관됨
- [ ] AC-WF001-10 PASS: `moai-foundation-thinking` 섹션 포함
- [ ] AC-WF001-11 PASS: `moai-design-system` trigger union
- [ ] AC-WF001-12 PASS: `moai-domain-database` cloud section
- [ ] AC-WF001-13 PASS: 2개 UNCLEAR skill Telemetry Window
- [ ] AC-WF001-14 PASS: 6개 REFACTOR skill 에 Refactor Notes 섹션
- [ ] AC-WF001-15 PASS: related-skills 자동 재작성
- [ ] EC-1 ~ 4 edge case 4건 모두 PASS
- [ ] `make build` 성공, `go test ./...`, `go vet ./...` 통과
- [ ] OPEN QUESTIONS (OQ-1 ~ OQ-7) 모두 user 확인으로 해소
- [ ] `wave-1.7-report.md` 작성, `baseline-hashes.txt` 정리

**AC 총 15개 + Edge Case 4개 = 19개 관찰가능 검증 포인트**. 모두 PASS 시 `<moai>COMPLETE</moai>` 마커 부여.
