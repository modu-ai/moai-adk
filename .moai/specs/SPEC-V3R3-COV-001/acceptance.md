---
spec_id: SPEC-V3R3-COV-001
title: Acceptance Criteria — Mobile Native Coverage
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-COV-001/spec.md
---

# Acceptance Criteria — SPEC-V3R3-COV-001

## AC-COV001-01: expert-mobile agent 파일 존재 (local + template)

**Given** the expert-mobile agent should be added
**When** the implementation is applied
**Then** both `.claude/agents/moai/expert-mobile.md` and `internal/template/templates/.claude/agents/moai/expert-mobile.md` exist with valid frontmatter (name, description, triggers, tools, model, permissionMode)

### Verification

```bash
test -f .claude/agents/moai/expert-mobile.md || echo "MISSING local"
test -f internal/template/templates/.claude/agents/moai/expert-mobile.md || echo "MISSING template"

# Frontmatter required fields
for path in .claude/agents/moai/expert-mobile.md internal/template/templates/.claude/agents/moai/expert-mobile.md; do
  for field in "name: expert-mobile" "description:" "tools:" "model:" "permissionMode:"; do
    grep -q "^$field" "$path" || echo "MISSING $field in $path"
  done
done

# Triggers (multilingual)
grep -E "EN:.*mobile.*ios.*android" .claude/agents/moai/expert-mobile.md || echo "MISSING EN triggers"
grep -E "KO:.*모바일" .claude/agents/moai/expert-mobile.md || echo "MISSING KO triggers"

# Expected: empty output
```

Maps to: REQ-COV001-001, REQ-COV001-002

---

## AC-COV001-02: 3 신규 skill 모두 SKILL.md + 최소 module 존재

**Given** the implementation creates 3 new mobile skills
**When** inspecting the skill directories
**Then** each skill has `SKILL.md` and at least one module file

### Verification

```bash
for skill in moai-domain-mobile moai-framework-react-native moai-framework-flutter-deep; do
  for base in .claude/skills internal/template/templates/.claude/skills; do
    test -f "$base/$skill/SKILL.md" || echo "MISSING SKILL.md: $base/$skill"
    module_count=$(ls -1 "$base/$skill/modules/" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$module_count" -lt 1 ]; then
      echo "MISSING modules: $base/$skill"
    fi
  done
done

# progressive_disclosure block (DEF-007 baseline)
for skill in moai-domain-mobile moai-framework-react-native moai-framework-flutter-deep; do
  grep -q "progressive_disclosure:" ".claude/skills/$skill/SKILL.md" || echo "MISSING PD: $skill"
done
# Expected: empty output
```

Maps to: REQ-COV001-003, REQ-COV001-004, REQ-COV001-005

---

## AC-COV001-03: 4 mobile paradigm modules 존재 (`moai-domain-mobile/modules/`)

**Given** moai-domain-mobile skill is created
**When** inspecting its modules directory
**Then** 4 paradigm modules exist: `ios-native.md`, `android-native.md`, `react-native.md`, `flutter.md`

### Verification

```bash
for base in .claude/skills internal/template/templates/.claude/skills; do
  for module in ios-native.md android-native.md react-native.md flutter.md; do
    test -f "$base/moai-domain-mobile/modules/$module" \
      && echo "OK: $base/moai-domain-mobile/modules/$module" \
      || echo "MISSING: $base/moai-domain-mobile/modules/$module"
  done
done
# Expected: 8 OK lines (4 modules × 2 paths)
```

Maps to: REQ-COV001-003

---

## AC-COV001-04: 4-mobile-strategy comparison guide 존재 + 5 selection criteria 포함

**Given** moai-domain-mobile/modules/ exists
**When** inspecting strategy-comparison.md
**Then** the file exists at both paths and contains a comparison table with at least 5 criteria covering all 4 paradigms

### Verification

```bash
for base in .claude/skills internal/template/templates/.claude/skills; do
  path="$base/moai-domain-mobile/modules/strategy-comparison.md"
  test -f "$path" || echo "MISSING: $path"
  # Must include all 4 paradigms
  grep -q "iOS native" "$path" || echo "MISSING iOS native: $path"
  grep -q "Android native" "$path" || echo "MISSING Android native: $path"
  grep -q "React Native" "$path" || echo "MISSING RN: $path"
  grep -q "Flutter" "$path" || echo "MISSING Flutter: $path"
  # At least 5 criteria (rough proxy: at least 5 table rows)
  table_rows=$(grep -c "^|" "$path")
  if [ "$table_rows" -lt 6 ]; then  # 1 header + 5 rows minimum
    echo "INSUFFICIENT criteria: $path ($table_rows rows)"
  fi
done
# Expected: empty output
```

Maps to: REQ-COV001-006

---

## AC-COV001-05: routing keyword가 추가됨

**Given** the routing keyword definition file exists (location TBD via grep at implementation time)
**When** searching for mobile keywords
**Then** the routing definition includes mobile, ios, android, swift, kotlin, react-native, flutter

### Verification

```bash
# 후보 routing 정의 파일들 확인
ROUTING_FILES=$(grep -rl "expert-backend\\|expert-frontend" .claude/skills/moai/ 2>/dev/null | head -5)
echo "Candidate routing files: $ROUTING_FILES"

# At least one of these files should now reference expert-mobile
grep -rE "expert-mobile" .claude/skills/moai/ \
  && echo "OK: routing references expert-mobile" \
  || echo "MISSING: expert-mobile not in routing"

# Mobile keyword presence in routing area
grep -rE "mobile|ios|android|swift|kotlin|react-native|flutter" .claude/skills/moai/ | head -5
```

Maps to: REQ-COV001-007

---

## AC-COV001-06: Template + local 동기화 완료

**Given** all new files are created at both paths
**When** comparing template and local
**Then** files are identical

### Verification

```bash
for path in .claude/agents/moai/expert-mobile.md \
            .claude/skills/moai-domain-mobile/SKILL.md \
            .claude/skills/moai-framework-react-native/SKILL.md \
            .claude/skills/moai-framework-flutter-deep/SKILL.md; do
  template="internal/template/templates/$path"
  diff -q "$path" "$template" \
    && echo "OK: $path synced" \
    || echo "DRIFT: $path"
done

# Modules
for skill in moai-domain-mobile moai-framework-react-native moai-framework-flutter-deep; do
  diff -rq ".claude/skills/$skill/modules/" "internal/template/templates/.claude/skills/$skill/modules/" 2>&1 \
    | grep -v "^Common subdirectories" \
    || echo "OK: $skill modules synced"
done
# Expected: only OK lines, no DRIFT
```

Maps to: REQ-COV001-008

---

## AC-COV001-07: make build + go test 통과

**Given** all new files added
**When** running build and test
**Then** both succeed

### Verification

```bash
make build
# Expected: exit 0
go test -count=1 ./internal/template/...
# Expected: PASS
```

Maps to: REQ-COV001-008

---

## AC-COV001-08: 기존 moai-framework-flutter 무수정 (`SPEC-FW-FLUTTER-001` 보존)

**Given** the existing moai-framework-flutter skill exists
**When** the new mobile coverage is added
**Then** the existing skill body is unchanged

### Verification

```bash
# Existing skill location
EXISTING="moai-framework-flutter"
test -d ".claude/skills/$EXISTING" || echo "EXISTING NOT FOUND (skipping AC-08)"

# git diff for the EXISTING skill
git diff .claude/skills/$EXISTING/ internal/template/templates/.claude/skills/$EXISTING/ 2>/dev/null | head -5
# Expected: empty (no diff)
git diff --name-only .claude/skills/$EXISTING/ | wc -l
# Expected: 0
```

Maps to: REQ-COV001-011

---

## AC-COV001-09: expert-mobile frontmatter Agent tool 포함 + Escalation Protocol body

**Given** the expert-mobile agent is created
**When** inspecting frontmatter and body
**Then** frontmatter `tools:` includes `Agent` (per ARCH-003 baseline) and body includes `## Escalation Protocol` section

### Verification

```bash
for path in .claude/agents/moai/expert-mobile.md internal/template/templates/.claude/agents/moai/expert-mobile.md; do
  grep -E "^tools:.*\\bAgent\\b" "$path" > /dev/null \
    && echo "OK Agent tool: $path" \
    || echo "MISSING Agent tool: $path"
  grep -q "^## Escalation Protocol" "$path" \
    && echo "OK Escalation Protocol: $path" \
    || echo "MISSING Escalation Protocol: $path"
done
# Expected: 4 OK lines
```

Maps to: REQ-COV001-009

---

## Edge Cases

### EC-1: 기존 routing 정의 파일 위치 불명확
If the routing definition file location is unclear at implementation time, T-A4-1 MUST first run `grep -rn "expert-backend\\|expert-frontend" .claude/` to identify candidates, document the chosen file in implementation notes, then add the mobile routing keyword.

### EC-2: moai-framework-flutter와 moai-framework-flutter-deep 사용자 혼동
Both skills' SKILL.md description MUST clearly state scope:
- `moai-framework-flutter`: "Basic Flutter UI, widget composition. For full-stack Flutter use moai-framework-flutter-deep."
- `moai-framework-flutter-deep`: "Full-stack Flutter (state mgmt, navigation, networking, platform channels). Excludes Firebase. For UI basics use moai-framework-flutter."

### EC-3: 4 paradigm 동등 취급 어려움 시
If team or user demands skew toward 1-2 paradigms (e.g., RN + Flutter only), the strategy-comparison.md MUST still cover all 4 with equal depth — selection bias is forbidden.

### EC-4: 신규 skill SKILL.md에 progressive_disclosure 누락
If created without DEF-007 baseline, the implementation MUST detect and add the block (level1_tokens: 100, level2_tokens: 5000) before completion.

---

## Definition of Done

- [ ] AC-COV001-01: expert-mobile 파일 존재 (local + template) + frontmatter 검증
- [ ] AC-COV001-02: 3 skills SKILL.md + module 최소 1개 존재 + PD 블록
- [ ] AC-COV001-03: moai-domain-mobile에 4 paradigm modules
- [ ] AC-COV001-04: strategy-comparison.md + 5 criteria
- [ ] AC-COV001-05: routing keyword 추가
- [ ] AC-COV001-06: template-local 동기화
- [ ] AC-COV001-07: make build + go test 통과
- [ ] AC-COV001-08: 기존 moai-framework-flutter 무수정
- [ ] AC-COV001-09: expert-mobile Agent tool + Escalation Protocol
- [ ] Edge cases EC-1/2/3/4 처리
