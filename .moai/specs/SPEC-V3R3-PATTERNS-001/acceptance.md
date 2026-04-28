---
id: SPEC-V3R3-PATTERNS-001
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: High
labels: [patterns, harness, acceptance, v3r3]
issue_number: null
---

# SPEC-V3R3-PATTERNS-001 — Acceptance Criteria

## 1. Definition of Done

본 SPEC은 다음을 모두 만족하면 완료(`status: completed`):

- AC-001 ~ AC-005 모두 PASS
- TRUST 5 quality gate 통과 (Tested / Readable / Unified / Secured / Trackable)
- `go test ./internal/template/...` green
- `make build` 성공 (embedded.go 재생성 정상)
- plan-auditor (있다면) PASS verdict 또는 BYPASSED with documented reason

---

## 2. Acceptance Criteria

### AC-001 — Agent Patterns 산출물 (REQ-PAT-001)

**Given** v3.0.0 R3 Phase A baseline 환경 (DEF-007, ARCH-003 완료),
**When** Wave 2 완료 후 `.claude/rules/moai/development/agent-patterns.md`가 작성된 경우,
**Then** 다음을 모두 만족해야 한다:

- 파일 존재: `test -f .claude/rules/moai/development/agent-patterns.md`
- 6 patterns 명시: grep으로 "Team", "Sub-agent", "Hybrid", "Orchestrator", "Specialist", "Pipeline" 6개 키워드 모두 발견
- 각 pattern 섹션에 다음 4 요소 포함: definition / when to use / example / anti-pattern
- 상단에 Apache 2.0 attribution 주석 포함: `grep -q "revfactory/harness" file && grep -q "Apache" file`
- frontmatter `paths` 글로브가 `.claude/agents/**` 또는 동등한 패턴 포함
- Template mirror 동기화: `internal/template/templates/.claude/rules/moai/development/agent-patterns.md`이 byte-identical

**검증 명령**:
```bash
diff .claude/rules/moai/development/agent-patterns.md \
     internal/template/templates/.claude/rules/moai/development/agent-patterns.md
# Exit code 0 (no diff) expected
```

---

### AC-002 — Boundary Verification + A/B Testing 산출물 (REQ-PAT-002, REQ-PAT-003)

**Given** Wave 2 진행,
**When** boundary-verification.md + skill-ab-testing.md 작성 완료된 경우,
**Then** 다음을 모두 만족해야 한다:

- `.claude/rules/moai/quality/boundary-verification.md` 존재 (디렉토리 quality/ 신규 생성됨)
- boundary-verification: 7 bug case 명시 — 각 case에 symptom / root cause / boundary condition / verification strategy 4 요소 포함
- skill-ab-testing: A/B methodology 명시 — with-skill vs baseline / evaluation criteria / sample size / statistical significance 4 요소 포함
- 두 파일 모두 frontmatter `paths` 글로브 적절히 설정 (boundary-verification은 test 파일, skill-ab-testing은 skill 파일)
- 두 파일 모두 Apache 2.0 attribution 상단 주석
- Template mirror 양쪽 byte-identical

**검증 명령**:
```bash
# 7 bug case 확인 (case 헤더 카운트)
grep -c "^### Case" .claude/rules/moai/quality/boundary-verification.md
# Output: 7 expected

# A/B methodology 핵심 키워드 확인
grep -E "(sample size|baseline|statistical)" .claude/rules/moai/development/skill-ab-testing.md
# At least 3 matches expected
```

---

### AC-003 — Team Cookbook + Orchestrator Templates 산출물 (REQ-PAT-004, REQ-PAT-005)

**Given** Wave 2 진행,
**When** team-pattern-cookbook.md + orchestrator-templates.md 작성 완료된 경우,
**Then** 다음을 모두 만족해야 한다:

- `.claude/rules/moai/workflow/team-pattern-cookbook.md` 존재
- 5 팀 예시 명시: research / implementation / review / design / debug — 5개 모두 포함
- 각 팀 예시에 role profile / file ownership / communication / shutdown sequence 4 요소
- `.claude/rules/moai/development/orchestrator-templates.md` 존재
- 3 templates 명시: Team-orchestrator / Sub-orchestrator / Hybrid-orchestrator
- 각 template에 when to use / how to spawn / error recovery / escalation 4 요소
- 두 파일 모두 Apache 2.0 attribution + Template mirror 동기화

**검증 명령**:
```bash
# 5 팀 헤더 확인
grep -E "^## (Research|Implementation|Review|Design|Debug)" \
  .claude/rules/moai/workflow/team-pattern-cookbook.md | wc -l
# Output: 5 expected

# 3 orchestrator templates 확인
grep -E "^## (Team|Sub|Hybrid)-orchestrator" \
  .claude/rules/moai/development/orchestrator-templates.md | wc -l
# Output: 3 expected
```

---

### AC-004 — Skill Writing Craft 산출물 (REQ-PAT-006)

**Given** Wave 2 진행,
**When** skill-writing-craft.md 작성 완료된 경우,
**Then** 다음을 모두 만족해야 한다:

- `.claude/rules/moai/development/skill-writing-craft.md` 존재
- 3 핵심 영역 모두 다룸:
  1. Description craft — when to trigger / when to skip
  2. Body structure — 3-level progressive disclosure
  3. Schema validation — frontmatter rules
- Apache 2.0 attribution + Template mirror 동기화
- frontmatter `paths` 글로브가 `.claude/skills/**/SKILL.md` 또는 동등 패턴 포함

**검증 명령**:
```bash
grep -E "(description|progressive disclosure|frontmatter)" \
  .claude/rules/moai/development/skill-writing-craft.md | wc -l
# At least 6 matches (각 키워드 2회 이상) expected
```

---

### AC-005 — Apache 2.0 Compliance + Template-First Sync (REQ-PAT-007, REQ-PAT-008)

**Given** Wave 2 + Wave 3 완료,
**When** 6 rule 파일 + NOTICE 모두 작성/동기화된 경우,
**Then** 다음을 모두 만족해야 한다:

- `.claude/rules/moai/NOTICE.md` 존재
- NOTICE.md에 다음 모두 명시:
  - revfactory/harness 저장소 URL
  - Apache 2.0 license URL
  - 흡수된 6 reference 파일 경로 일람
  - import 날짜 (2026-04-26)
  - attribution clause 텍스트
  - Apache 2.0 NOTICE 파일 사본 (harness가 가지고 있다면)
- 6 산출 파일 모두 상단에 동일한 attribution 주석 포함
- Template mirror 동기화: 7 파일 (6 rule + NOTICE) 모두 `internal/template/templates/.claude/rules/moai/` 하위에 byte-identical 사본 존재
- `make build` 성공
- `go test ./internal/template/...` green
- 16-language neutrality 검증: 산출 파일 6개에서 특정 언어 하드코딩 grep 검사 0건 (Python/Go/JavaScript 등 단독 표기 없이 multi-language 또는 의사코드)

**검증 명령**:
```bash
# 7 파일 모두 mirror 동기화 확인
for f in agent-patterns.md skill-ab-testing.md orchestrator-templates.md skill-writing-craft.md; do
  diff .claude/rules/moai/development/$f \
       internal/template/templates/.claude/rules/moai/development/$f
done

diff .claude/rules/moai/quality/boundary-verification.md \
     internal/template/templates/.claude/rules/moai/quality/boundary-verification.md

diff .claude/rules/moai/workflow/team-pattern-cookbook.md \
     internal/template/templates/.claude/rules/moai/workflow/team-pattern-cookbook.md

diff .claude/rules/moai/NOTICE.md \
     internal/template/templates/.claude/rules/moai/NOTICE.md

# All diffs should exit 0

# Apache attribution 7 파일 모두 확인
for f in $(find .claude/rules/moai/{development,quality,workflow} -name "*.md" -newer .moai/specs/SPEC-V3R3-PATTERNS-001/spec.md); do
  grep -q "revfactory/harness" "$f" || echo "MISSING attribution: $f"
done
# No output expected (all have attribution)

# Template build + test
make build && go test ./internal/template/...
```

---

## 3. Edge Cases

### EC-001 — harness 저장소 LICENSE가 Apache 2.0이 아닌 경우

**Detection**: M1에서 `cat /tmp/harness-analysis/harness/LICENSE` 결과가 Apache 2.0이 아님.

**Response**:
1. SPEC abort.
2. AskUserQuestion으로 사용자에게 보고.
3. 사용자 결정에 따라: (A) 다른 라이선스로 SPEC 재작성, (B) prior art 흡수 포기, (C) 사용자가 라이선스 호환성 직접 확인 후 진행.

### EC-002 — harness reference 파일 일부 부재

**Detection**: M1에서 6 파일 중 일부가 `skills/harness/references/`에 없음.

**Response**:
1. 누락된 파일을 progress.md에 기록.
2. AskUserQuestion으로 사용자에게 보고.
3. 사용자 결정: (A) 부재한 파일은 MoAI-ADK 자력 작성, (B) 해당 파일 산출 제외 후 부분 SPEC 진행, (C) SPEC abort.

### EC-003 — 흡수된 docs가 Markdown 외 형식 (RST, AsciiDoc 등)

**Detection**: M1에서 reference 파일 확장자/형식 확인 시.

**Response**: Markdown으로 자동 변환 (Pandoc 또는 수작업). 변환 사실을 attribution 섹션에 명시: "Converted from <format> to Markdown on 2026-04-26."

### EC-004 — 흡수 docs가 너무 길어 progressive_disclosure가 필요한 경우

**Detection**: M2에서 작성한 파일이 500 LOC 초과.

**Response**: SKILL.md 패턴 적용 — 메인 파일은 Quick Reference (200 LOC 이내) + `modules/` 디렉토리에 deep-dive 분리. frontmatter에 `progressive_disclosure: { enabled: true, level1_tokens: ~1500, level2_tokens: ~5000 }` 추가.

### EC-005 — Template mirror 동기화 실패 (`make build`)

**Detection**: M3에서 `make build` exit code 비-0.

**Response**:
1. 에러 로그 분석.
2. `internal/template/templates/.claude/rules/moai/` 경로/파일명 정확성 재확인.
3. `internal/template/embedded.go` 재생성 정상성 확인.
4. 5번 retry 한도 초과 시 사용자에게 보고.

### EC-006 — frontmatter `paths` 글로브가 의도와 다르게 매칭

**Detection**: M4 validation에서 의도된 파일 수정 시 rule이 로드되지 않거나, 의도되지 않은 파일에서 로드됨.

**Response**:
1. 글로브 패턴 재검토 (Claude Code paths 글로브 문법 참조).
2. `paths: "**/*.go,**/*_test.go"` 같은 다중 패턴 사용 검토.
3. 수정 후 재검증.

---

## 4. Test Strategy

### Phase 1 — Build Tests (Wave 3)

```bash
make build
# Verify: internal/template/embedded.go modified
# Verify: exit code 0

go test ./internal/template/...
# Verify: all tests green
# Verify: commands_audit_test.go pass
```

### Phase 2 — Content Verification (Wave 3)

```bash
# Apache attribution 7 파일 (6 rule + NOTICE)
files=(
  ".claude/rules/moai/development/agent-patterns.md"
  ".claude/rules/moai/development/skill-ab-testing.md"
  ".claude/rules/moai/development/orchestrator-templates.md"
  ".claude/rules/moai/development/skill-writing-craft.md"
  ".claude/rules/moai/quality/boundary-verification.md"
  ".claude/rules/moai/workflow/team-pattern-cookbook.md"
  ".claude/rules/moai/NOTICE.md"
)
for f in "${files[@]}"; do
  test -f "$f" && echo "OK: $f" || echo "MISSING: $f"
  grep -q "revfactory/harness" "$f" || echo "NO ATTRIBUTION: $f"
  grep -q "Apache" "$f" || echo "NO APACHE REF: $f"
done
```

### Phase 3 — Mirror Verification (Wave 3)

```bash
# byte-identical mirror
local_root=".claude/rules/moai"
template_root="internal/template/templates/.claude/rules/moai"
for sub in development/agent-patterns.md development/skill-ab-testing.md development/orchestrator-templates.md development/skill-writing-craft.md quality/boundary-verification.md workflow/team-pattern-cookbook.md NOTICE.md; do
  diff "$local_root/$sub" "$template_root/$sub" || echo "MIRROR DRIFT: $sub"
done
```

### Phase 4 — Frontmatter `paths` Glob Verification (Wave 4)

```bash
# 각 파일의 frontmatter paths가 적절한지 manual review
for f in "${files[@]}"; do
  if [[ "$f" != *NOTICE.md ]]; then
    head -20 "$f" | grep -q "paths:" || echo "NO paths frontmatter: $f"
  fi
done
```

### Phase 5 — Language Neutrality (Wave 4)

```bash
# 16-language hardcoding 검사
banned_patterns=("\\.py\\b" "\\.go\\b" "\\.ts\\b" "\\.rs\\b" "pip install" "go get" "npm install")
for f in "${files[@]}"; do
  for p in "${banned_patterns[@]}"; do
    if grep -E "$p" "$f" > /dev/null; then
      # 단독 표기인지 multi-language 컨텍스트인지 manual review 필요
      echo "POTENTIAL HARDCODING in $f: pattern $p"
    fi
  done
done
# Manual review of flagged matches
```

---

## 5. Quality Gate Criteria (TRUST 5)

- **Tested**: 본 SPEC은 markdown artifact 산출 — 회귀 테스트는 `internal/template/...` 기존 테스트로 충분. 추가 테스트 불필요.
- **Readable**: 산출 파일 6개가 명확한 헤더/섹션 구조. ko/en 정책 준수 (rule 파일은 영어).
- **Unified**: 6 파일 모두 동일한 frontmatter 구조 + 동일한 attribution 패턴. NOTICE.md 단일 진실 공급원.
- **Secured**: 외부 코드 흡수 시 라이선스 검증 의무 — Apache 2.0 명시 + NOTICE 보존.
- **Trackable**: SPEC-V3R3-PATTERNS-001 ID로 commit 추적, history 표 frontmatter, attribution date 명시.
