---
id: SPEC-V3R6-AGENT-MODEL-ROUTING-001-ACCEPTANCE
title: "Acceptance — Agent 23개 모델 명시 라우팅 (13 binary ACs)"
version: "0.2.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents"
lifecycle: spec-anchored
tags: "agent, model-routing, opus, sonnet, haiku, cost-optimization, sprint-2, v3.0, acceptance"
tier: L
---

# Acceptance — SPEC-V3R6-AGENT-MODEL-ROUTING-001 Tier L

본 acceptance는 **13 binary ACs** (각각 PASS/FAIL 단일 답변 가능) + Given-When-Then 시나리오 + 100% REQ ↔ AC traceability + Definition of Done으로 구성된다.

---

## 1. Binary Acceptance Criteria (13 ACs)

### AC-AMR-001: 23-agent inventory 정확성

**Statement**: 4-subdirectory glob에서 정확히 23 agent files 발견.

**Verification Command**:
```bash
find .claude/agents -name "*.md" -type f | wc -l
```

**Expected Output**: `23`

**Status**: PASS if output equals `23`, FAIL otherwise.

**Given-When-Then**:
- **Given** the 4-subdirectory layout `.claude/agents/{core, expert, harness, meta}/`
- **When** the verification command is executed
- **Then** the output equals exactly `23`

**Traces to**: REQ-AMR-001

---

### AC-AMR-002: Opus tier 7 agents 명시

**Statement**: 7 agents의 frontmatter에 `model: opus` 명시.

**Verification Command**:
```bash
grep -l 'model: opus' .claude/agents/core/*.md .claude/agents/expert/*.md .claude/agents/harness/*.md .claude/agents/meta/*.md | wc -l
```

**Expected Output**: `7`

**Expected Members**:
- `core/manager-develop.md`
- `core/manager-spec.md`
- `core/manager-strategy.md`
- `expert/expert-security.md`
- `expert/expert-refactoring.md` (constitution-aligned)
- `meta/plan-auditor.md`
- `meta/evaluator-active.md`

**Status**: PASS if output equals `7` AND member list matches, FAIL otherwise.

**Given-When-Then**:
- **Given** opus tier 7 agents의 frontmatter migration 완료
- **When** grep `model: opus` 명령 실행
- **Then** 정확히 7개 files match (예상 멤버 일치)

**Traces to**: REQ-AMR-002, REQ-AMR-NF-010

---

### AC-AMR-003: Sonnet tier 13 agents 명시

**Statement**: 13 agents의 frontmatter에 `model: sonnet` 명시.

**Verification Command**:
```bash
grep -l 'model: sonnet' .claude/agents/core/*.md .claude/agents/expert/*.md .claude/agents/harness/*.md .claude/agents/meta/*.md | wc -l
```

**Expected Output**: `13`

**Expected Members**:
- `core/manager-brain.md`
- `core/manager-project.md`
- `core/manager-quality.md`
- `expert/expert-backend.md`
- `expert/expert-devops.md`
- `expert/expert-frontend.md`
- `expert/expert-performance.md`
- `harness/cli-template-specialist.md` (iter 2 new inventory)
- `harness/hook-ci-specialist.md` (iter 2 new inventory)
- `harness/quality-specialist.md` (iter 2 new inventory)
- `harness/workflow-specialist.md` (iter 2 new inventory)
- `meta/builder-harness.md`
- `meta/claude-code-guide.md`

**Status**: PASS if output equals `13` AND member list matches, FAIL otherwise.

**Given-When-Then**:
- **Given** sonnet tier 13 agents migration 완료 (4 harness specialists 포함)
- **When** grep `model: sonnet` 명령 실행
- **Then** 정확히 13개 files match

**Traces to**: REQ-AMR-003

---

### AC-AMR-004: Haiku tier 3 agents 명시

**Statement**: 3 agents의 frontmatter에 `model: haiku` 명시.

**Verification Command**:
```bash
grep -l 'model: haiku' .claude/agents/core/*.md .claude/agents/expert/*.md .claude/agents/harness/*.md .claude/agents/meta/*.md | wc -l
```

**Expected Output**: `3`

**Expected Members**:
- `core/manager-docs.md` (pre-existing)
- `core/manager-git.md` (pre-existing)
- `meta/researcher.md` (NEW: inherit → haiku)

**Status**: PASS if output equals `3` AND member list matches, FAIL otherwise.

**Given-When-Then**:
- **Given** haiku tier 3 agents migration 완료 (researcher 신규)
- **When** grep `model: haiku` 명령 실행
- **Then** 정확히 3개 files match

**Traces to**: REQ-AMR-004

---

### AC-AMR-005: researcher batch_api opt-in key

**Statement**: `researcher.md` frontmatter에 batch_api opt-in key 중 정확히 1개 명시.

**Verification Command**:
```bash
grep -E '^(batch_api: true|use_batch_api: true|invocation_mode: batch)$' .claude/agents/meta/researcher.md | wc -l
```

**Expected Output**: `1`

**Accepted Keys** (M1에서 Claude Code SDK 공식 문서로 canonical key 확정):
- `batch_api: true`
- `use_batch_api: true`
- `invocation_mode: batch`

**Status**: PASS if exactly one of three keys present, FAIL otherwise.

**Given-When-Then**:
- **Given** M1에서 Claude Code SDK batch_api canonical key 확정
- **When** researcher.md frontmatter에 1개 key 명시
- **Then** grep 출력 = 1

**Traces to**: REQ-AMR-005

---

### AC-AMR-006: Baseline JSONL ≥ 23 entries

**Statement**: `.moai/state/agent-model-baseline.jsonl` 파일 존재 + 23 entries 이상.

**Verification Command**:
```bash
test -f .moai/state/agent-model-baseline.jsonl && wc -l < .moai/state/agent-model-baseline.jsonl
```

**Expected Output**: integer `≥ 23`

**Required Fields per Entry**:
- `agent_name` (string)
- `baseline_input_tokens_avg` (number)
- `baseline_output_tokens_avg` (number)
- `baseline_quality_score` (number, 0-1)
- `measurement_window_start` (ISO 8601 timestamp)
- `measurement_window_end` (ISO 8601 timestamp)

**Status**: PASS if file exists AND line count `≥ 23`, FAIL otherwise.

**Given-When-Then**:
- **Given** M1 baseline measurement 완료
- **When** baseline JSONL 파일 line count 측정
- **Then** ≥ 23 entries 확인

**Traces to**: REQ-AMR-006

---

### AC-AMR-007: Quality regression bound ±5%

**Statement**: post-run manager-quality validation에서 sonnet/haiku tier agents의 quality score가 baseline ±5% 이내.

**Verification Command**:
```bash
test -f .moai/reports/quality-regression/SPEC-V3R6-AGENT-MODEL-ROUTING-001.json && \
  jq '.agents[] | select(.regression_pct > 0.05 or .regression_pct < -0.05)' \
    .moai/reports/quality-regression/SPEC-V3R6-AGENT-MODEL-ROUTING-001.json | wc -l
```

**Expected Output**: `0` (no out-of-bound entries)

**Report Structure** (`.moai/reports/quality-regression/SPEC-V3R6-AGENT-MODEL-ROUTING-001.json`):
```json
{
  "spec_id": "SPEC-V3R6-AGENT-MODEL-ROUTING-001",
  "measurement_date": "2026-05-XX",
  "agents": [
    {
      "agent_name": "expert-backend",
      "baseline_quality_score": 0.87,
      "post_change_quality_score": 0.85,
      "regression_pct": -0.023,
      "within_bound": true
    }
    // ... 23 entries
  ]
}
```

**Status**: PASS if all 23 agents within ±5% bound, FAIL if any agent exceeds.

**Given-When-Then**:
- **Given** M6 post-run manager-quality validation 완료
- **When** regression report 생성
- **Then** 23 agents 모두 within_bound = true

**Traces to**: REQ-AMR-007

---

### AC-AMR-008: model: inherit count = 0

**Statement**: post-migration `model: inherit` 명시 agents = 0건.

**Verification Command**:
```bash
grep -l 'model: inherit' .claude/agents/core/*.md .claude/agents/expert/*.md .claude/agents/harness/*.md .claude/agents/meta/*.md | wc -l
```

**Expected Output**: `0`

**Status**: PASS if output equals `0`, FAIL otherwise.

**Given-When-Then**:
- **Given** M2+M3+M4 migration 완료
- **When** grep `model: inherit` 명령 실행
- **Then** 0개 files match (모두 명시적 model 분류)

**Traces to**: REQ-AMR-008, REQ-AMR-001

---

### AC-AMR-009: Template mirror byte-identical

**Statement**: 23 agent files의 local과 `internal/template/templates/.claude/agents/<sub>/<agent>.md` mirror가 byte-identical.

**Verification Command**:
```bash
mismatches=0
for sub in core expert harness meta; do
  for f in .claude/agents/$sub/*.md; do
    name=$(basename "$f")
    if ! diff -q "$f" "internal/template/templates/.claude/agents/$sub/$name" >/dev/null 2>&1; then
      mismatches=$((mismatches + 1))
    fi
  done
done
echo "$mismatches"
```

**Expected Output**: `0`

**Status**: PASS if 0 mismatches across all 23 pairs, FAIL otherwise.

**Given-When-Then**:
- **Given** M2+M3+M4 milestone마다 template mirror 동시 갱신
- **When** 23 pair diff 검증
- **Then** 0 mismatches

**Traces to**: REQ-AMR-NF-009

---

### AC-AMR-010: make build + embedded.go regeneration

**Statement**: `make build` exit 0 + `internal/template/embedded.go` 재생성.

**Verification Command**:
```bash
make build && test -s internal/template/embedded.go && echo "PASS"
```

**Expected Output**: `PASS`

**Status**: PASS if make build exits 0 AND embedded.go is non-empty after build, FAIL otherwise.

**Given-When-Then**:
- **Given** M5 단계에서 23 agent 변경 완료
- **When** `make build` 실행
- **Then** exit 0 + embedded.go 재생성

**Traces to**: REQ-AMR-NF-009 (template mirror sync mechanism)

---

### AC-AMR-011: moai doctor exit 0

**Statement**: `moai doctor` 명령 실행 시 exit 0 (no regression).

**Verification Command**:
```bash
moai doctor; echo "EXIT_CODE=$?"
```

**Expected Output**: `EXIT_CODE=0`

**Status**: PASS if exit code = 0, FAIL otherwise.

**Given-When-Then**:
- **Given** M2~M4 milestone마다 agent migration 적용
- **When** `moai doctor` 명령 실행
- **Then** exit 0 (warnings 가능, errors 0)

**Traces to**: REQ-AMR-NF-011 (no Go code change, no doctor regression)

---

### AC-AMR-012: Constitution opus tier verbatim 정렬

**Statement**: 본 SPEC의 opus tier 7 agents가 `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy의 reasoning-intensive list와 일치.

**Verification Command**:
```bash
grep -E "(manager-spec|manager-strategy|plan-auditor|evaluator-active|expert-security|expert-refactoring|manager-develop)" \
  .claude/rules/moai/core/moai-constitution.md | head -1 && \
  grep -l 'model: opus' .claude/agents/{core,expert,harness,meta}/*.md
```

**Expected Output**:
- constitution line containing all 6 reasoning-intensive names (manager-develop은 implementation-intensive로 분류되지만 본 SPEC에서 opus assignment)
- 7 file paths matching opus tier members from AC-AMR-002

**Status**: PASS if constitution line contains the 6 names (expert-refactoring 포함) AND 7 opus agents enumerated correctly, FAIL otherwise.

**Given-When-Then**:
- **Given** `.claude/rules/moai/core/moai-constitution.md` line 55 verbatim ("reasoning-intensive agents (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, **expert-refactoring**)")
- **When** SPEC opus tier 7 agents migration 적용
- **Then** 정렬 verified (constitution 본문 우선)

**Traces to**: REQ-AMR-002 + Background §1.5

---

### AC-AMR-013: docs-site 4-locale parity ≤ 1.20

**Statement**: docs-site `content/{en,ko,ja,zh}/` agent catalog 4-locale 동시 갱신 + parity ratio ≤ 1.20.

**Verification Command**:
```bash
# Parity ratio = max(locale_word_count) / min(locale_word_count)
en_wc=$(wc -w < docs-site/content/en/agents.md 2>/dev/null || echo 0)
ko_wc=$(wc -w < docs-site/content/ko/agents.md 2>/dev/null || echo 0)
ja_wc=$(wc -w < docs-site/content/ja/agents.md 2>/dev/null || echo 0)
zh_wc=$(wc -w < docs-site/content/zh/agents.md 2>/dev/null || echo 0)
max_wc=$(echo -e "$en_wc\n$ko_wc\n$ja_wc\n$zh_wc" | sort -n | tail -1)
min_wc=$(echo -e "$en_wc\n$ko_wc\n$ja_wc\n$zh_wc" | sort -n | head -1)
test "$min_wc" -gt 0 && python3 -c "print(round($max_wc / $min_wc, 3))"
```

**Expected Output**: numeric ratio `≤ 1.20`

**Status**: PASS if ratio ≤ 1.20, FAIL otherwise.

**Given-When-Then**:
- **Given** M6 docs-site 4-locale catalog 갱신 완료
- **When** parity ratio 계산
- **Then** ratio ≤ 1.20 (GEARS-MIGRATION-001 baseline 1.149 참조)

**Traces to**: REQ-AMR-NF-013

---

## 2. Traceability Matrix (REQ ↔ AC, 100% coverage)

| REQ ID | REQ Type | AC ID(s) | Coverage |
|--------|----------|----------|----------|
| REQ-AMR-001 (Ubiquitous — 23/23 명시) | Functional | AC-AMR-001, AC-AMR-008 | ✓ |
| REQ-AMR-002 (Where — opus 7) | Functional | AC-AMR-002, AC-AMR-012 | ✓ |
| REQ-AMR-003 (Where — sonnet 13) | Functional | AC-AMR-003 | ✓ |
| REQ-AMR-004 (Where — haiku 3) | Functional | AC-AMR-004 | ✓ |
| REQ-AMR-005 (When + Where — researcher batch_api) | Functional | AC-AMR-005 | ✓ |
| REQ-AMR-006 (Ubiquitous — baseline measurement) | Functional | AC-AMR-006 | ✓ |
| REQ-AMR-007 (When + Where — regression ±5%) | Functional | AC-AMR-007 | ✓ |
| REQ-AMR-008 (Ubiquitous — inherit = 0) | Functional | AC-AMR-008 | ✓ |
| REQ-AMR-NF-009 (Template mirror sync) | Non-Functional | AC-AMR-009, AC-AMR-010 | ✓ |
| REQ-AMR-NF-010 (Opus 7/23 = 70% off) | Non-Functional | AC-AMR-002 (count = 7) | ✓ |
| REQ-AMR-NF-011 (No Go code change) | Non-Functional | AC-AMR-011 | ✓ |
| REQ-AMR-NF-012 (Env-var override preserved) | Non-Functional | AC-AMR-011 (moai doctor exit 0 = override mechanism intact) | ✓ |
| REQ-AMR-NF-013 (docs-site parity ≤ 1.20) | Non-Functional | AC-AMR-013 | ✓ |

**Coverage**: 13/13 REQs traced to ≥ 1 AC, 13/13 ACs trace to ≥ 1 REQ. **100% bidirectional traceability**.

---

## 3. Definition of Done (DoD)

본 SPEC가 `status: implemented` 변경 가능 조건:

### 3.1 Functional Completion

- [ ] AC-AMR-001 PASS (23-agent inventory)
- [ ] AC-AMR-002 PASS (opus = 7)
- [ ] AC-AMR-003 PASS (sonnet = 13)
- [ ] AC-AMR-004 PASS (haiku = 3)
- [ ] AC-AMR-005 PASS (researcher batch_api opt-in)
- [ ] AC-AMR-008 PASS (inherit = 0)

### 3.2 Quality Validation

- [ ] AC-AMR-006 PASS (baseline ≥ 23 entries)
- [ ] AC-AMR-007 PASS (regression ±5% all 23 agents)
- [ ] AC-AMR-011 PASS (moai doctor exit 0)
- [ ] AC-AMR-012 PASS (constitution opus tier 정렬)

### 3.3 Infrastructure Integrity

- [ ] AC-AMR-009 PASS (23 mirror byte-identical)
- [ ] AC-AMR-010 PASS (make build + embedded.go)
- [ ] AC-AMR-013 PASS (docs-site 4-locale parity ≤ 1.20)

### 3.4 Process Completion

- [ ] All 6 milestones (M1~M6) committed
- [ ] Feat branch merged to main (Hybrid Trunk Tier L PR squash)
- [ ] progress.md written (M1~M6 evidence + iter 2 BLOCKING/SHOULD-FIX resolution)
- [ ] 5 artifacts status `draft → implemented`, version `0.2.0 → 0.3.0`
- [ ] PROMPT-CACHE-001 cross-Sprint coordination documented (sync-phase ordering)

### 3.5 Plan-Auditor Approval

- [ ] iter 2 plan-auditor verdict: PASS ≥ 0.85 (Tier L threshold)
- [ ] 0 BLOCKING items remaining

---

## 4. Edge Cases

### 4.1 R-AMR-003 Triggered: Agent fallback chain failure

**Scenario**: M2 또는 M3에서 1개 agent가 명시적 model 지정 후 fallback chain 의존으로 호출 실패.

**Detection**: M2/M3 종료 sample 호출에서 응답 0 또는 timeout.

**Resolution**:
1. 해당 agent만 `model: inherit` 또는 `model: opus`로 즉시 revert (1 commit)
2. Cross-Sprint coordination ledger 업데이트
3. AC-AMR-008 PARTIAL FAIL 보고 (해당 agent 제외 22/23 inherit 제거)
4. Out-of-bound report에 fallback 의존 agent 명시 + design.md decision log 추가

### 4.2 AC-AMR-007 FAIL: 1 agent quality regression > ±5%

**Scenario**: post-run manager-quality validation에서 sonnet/haiku tier agent 1개가 baseline 대비 -5% 초과 저하.

**Detection**: `.moai/reports/quality-regression/SPEC-V3R6-AGENT-MODEL-ROUTING-001.json`의 `agents[].regression_pct < -0.05`.

**Resolution**:
1. 해당 agent revert (sonnet → opus 또는 haiku → sonnet, 1 commit)
2. AC-AMR-007 RE-VERIFY
3. design.md decision log에 revert 사유 + baseline JSONL 인용
4. PROMPT-CACHE-001 sync-phase ordering 영향 평가 (cache_write 손익분기 모델 변경)

### 4.3 AC-AMR-005 FAIL: Claude Code SDK batch_api canonical key 미확정

**Scenario**: M1 WebFetch에서 Claude Code SDK 공식 문서가 명확한 canonical key 미정의.

**Detection**: M1 orchestrator-direct WebFetch 응답에서 3개 후보 (`batch_api` / `use_batch_api` / `invocation_mode`) 중 명시적 선택 불가.

**Resolution**:
1. claude-code-guide agent 회귀 조사 (선택 task)
2. AskUserQuestion으로 user에게 canonical key 선택 의뢰 (M1 orchestrator-direct only, manager-develop 위임 시 blocker report)
3. 사용자 결정 후 M4 진행

### 4.4 R-AMR-005 Triggered: docs-site locale drift

**Scenario**: M6 docs-site 4-locale 갱신 시 1개 locale (예: zh) 누락.

**Detection**: AC-AMR-013 parity ratio > 1.20.

**Resolution**:
1. 누락 locale 즉시 갱신
2. docs-i18n-check baseline 확인
3. AC-AMR-013 RE-VERIFY

### 4.5 Cross-Sprint R-AMR-004: PROMPT-CACHE-001 sync 충돌

**Scenario**: AGENT-MODEL-ROUTING-001 sync PR과 PROMPT-CACHE-001 sync PR이 same-day 머지 → A/B regression detection이 cache 효과와 혼합.

**Detection**: GitHub PR 머지 타임스탬프 비교.

**Resolution**:
1. PROMPT-CACHE-001 sync PR을 AGENT-MODEL-ROUTING-001 sync PR 머지 후 24시간 이상 지연
2. AGENT-MODEL-ROUTING-001 sync PR 머지 후 manager-quality regression validation 즉시 실행
3. baseline JSONL pre-cache state 확보

---

## 5. Verification Schedule

| Milestone | ACs Verified | Stage |
|-----------|--------------|-------|
| M1 | AC-AMR-005, AC-AMR-006 | Baseline + SDK verification |
| M2 | AC-AMR-002, AC-AMR-009 (partial), AC-AMR-011, AC-AMR-012 | Opus tier |
| M3 | AC-AMR-003, AC-AMR-009 (partial), AC-AMR-011 | Sonnet tier |
| M4 | AC-AMR-004, AC-AMR-005, AC-AMR-008, AC-AMR-009 (partial) | Haiku + batch_api |
| M5 | AC-AMR-009 (full 23 pairs), AC-AMR-010 | Template build + inherit 0 |
| M6 | AC-AMR-001 (re-verify), AC-AMR-007, AC-AMR-013 | Regression + docs-site |
| Post-merge | All 13 ACs re-verified | DoD validation |

---

## 6. Out of Scope (Acceptance-Level)

### Out of Scope: Quality eval set Go implementation

REQ-AMR-007 quality score measurement은 M6 manager-quality 위임에서 manual eval set 적용. Go-based eval framework (`internal/eval/`)는 별도 SPEC (가칭 `SPEC-V3R7-AGENT-EVAL-FRAMEWORK-001`).

### Out of Scope: Continuous regression monitoring

본 SPEC는 1-time post-migration validation만. CI-integrated continuous regression monitoring (예: nightly cron with eval set replay)는 향후 SPEC.

### Out of Scope: Per-agent model override CLI

본 SPEC는 manual frontmatter edit. `moai agent model set <agent> <tier>` CLI 명령은 plan.md Out of Scope §Migration tool automation 참조.
