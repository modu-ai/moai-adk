---
id: SPEC-V3R6-AGENT-MODEL-ROUTING-001-PLAN
title: "Plan — Agent 23개 모델 명시 라우팅 (Tier L Section A-E)"
version: "0.2.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents"
lifecycle: spec-anchored
tags: "agent, model-routing, opus, sonnet, haiku, cost-optimization, sprint-2, v3.0, plan"
tier: L
---

# Plan — SPEC-V3R6-AGENT-MODEL-ROUTING-001 Tier L

본 plan은 `.claude/rules/moai/development/manager-develop-prompt-template.md` Section A-E **REQUIRED for Tier L** 5-section 구조 + 6 milestones (M1~M6) 분할로 작성된다.

---

## Section A — Context (위치 + 분기 + SPEC 산출물 경로)

### A.1 작업 위치 + 분기

- **Project root**: `/Users/goos/MoAI/moai-adk-go`
- **Main HEAD baseline**: `0abcda296` (2026-05-23, V3R6 UPDATE Audit 3 SPECs plan-complete)
- **Sprint 1 Lane A 진행 상황**: 4 SPECs 중 1개 implementation 머지 완료 (RULES-PATH-SCOPE-001 commit `7ed4c841c`); 잔여 3개 SPECs (RULES-COMPRESS / SKILL-CONSOLIDATE / SKILL-COMPRESS) plan-complete + run-pending
- **본 SPEC 위치**: Sprint 2 Tier L 첫 SPEC. Plan-phase는 Sprint 1 잔여 SPECs와 병렬 가능. Run-phase entry는 Sprint 1 머지 후 권장
- **Branch convention**: `feat/SPEC-V3R6-AGENT-MODEL-ROUTING-001` (Tier L Hybrid Trunk PR 의무)

### A.2 SPEC 산출물 5 artifacts (Tier L)

| 파일 | 라인 수 (목표) | 역할 |
|------|----------------|------|
| `.moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/spec.md` | ~310 lines | EARS 8 REQs + 5 NFR + 23-agent inventory + 6 Risks + 3 h3 Out of Scope |
| `.moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/plan.md` | ~480 lines (본 파일) | Section A-E + 6 milestones |
| `.moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/acceptance.md` | ~430 lines | 13 binary ACs + 100% traceability + DoD |
| `.moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/design.md` | ~250 lines | 4-subdir 아키텍처 + alternatives + decision log |
| `.moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/research.md` | ~310 lines | Baseline 측정 방법론 + A/B framework + cost-saving 검증 |

### A.3 plan-auditor verdict

- **iter 1**: REVISE 0.633 (Tier M threshold 0.80 미만 -0.167)
- **iter 2 본 작업**: 4 BLOCKING (B1 23-agent inventory / B2 AC paths / B3 module path / B4 cost math + opus count) + 5 SHOULD-FIX (S1 sprint-round / S2 constitution / S3 Tier M→L / S4 batch_api / S5 baseline) 적용
- **iter 2 PASS threshold**: 0.85 (Tier L 재분류 결과)

### A.4 기존 인프라 (PRESERVE 대상 + EXTEND 대상)

**PRESERVE (변경 금지)**:
- 23 agent files의 body 내용 (frontmatter `model:` 필드만 변경; description / tools / allowed-tools 보존)
- `internal/agent/` 디렉토리 (존재하지 않음, 신규 작성 금지)
- `internal/loader/` 디렉토리 (존재하지 않음, 신규 작성 금지)
- `.claude/settings.json` `effortLevel` 필드
- `CLAUDE_MODEL` env var override 메커니즘

**EXTEND (확장 대상)**:
- 23 agent files의 frontmatter `model:` 필드 (inherit → opus | sonnet | haiku)
- `researcher.md` frontmatter에 batch_api opt-in key 1개 추가
- `internal/template/templates/.claude/agents/{core,expert,harness,meta}/*.md` 23 mirror 파일
- `docs-site/content/{en,ko,ja,zh}/` agent catalog 페이지 4-locale

---

## Section B — Known Issues 자동 주입 (Tier L MANDATORY 8 카테고리)

### B1. Cross-platform Build Tags

- **본 SPEC 영향도**: **무관** (frontmatter markdown만 변경, syscall 코드 0건)
- **검증**: 본 SPEC scope 외. M5 template mirror 동기화 시 `go test ./internal/template/...` exit 0 확인만

### B2. Cross-SPEC 정책 충돌 사전 스캔

- **충돌 후보 1**: SPEC-V3R6-PROMPT-CACHE-001 (Sprint 2 sibling) — sync-phase 분리 의무 (R-AMR-004)
- **충돌 후보 2**: SPEC-V3R6-BACKEND-ROUTING-001 (Sprint 3) — orthogonal (model tier vs provider)
- **충돌 후보 3**: `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy 라인 55 — **expert-refactoring opus carve-out 정렬 완료** (iter 2 S2 적용)
- **검증 명령**: `grep -r "Retired\|TestHarnessRetirement\|superseded" .claude/agents/ || echo "no conflicts"` (예상 출력: no conflicts)

### B3. C-HRA-008 / Subagent Boundary Discipline

- **본 SPEC 영향도**: **무관** (agent frontmatter만 변경, AskUserQuestion 호출 추가 없음)
- **검증 명령**: `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/agents/ | grep -v "// "` → 기존 baseline 유지 확인

### B4. Frontmatter Canonical Schema

- **본 SPEC spec.md 자체**: 12-field canonical 준수 (created/updated/tags, snake_case alias 0건)
- **Agent file frontmatter**: SPEC frontmatter와 다른 schema (model + tools + description) — 무관
- **검증**: spec.md / plan.md / acceptance.md / design.md / research.md 5 파일 모두 12-field 검증 (M0 pre-flight)

### B5. CI 3-tier 인지

- **spec-lint**: 본 SPEC 5 artifacts에 spec-lint baseline 확인 (M0)
- **golangci-lint**: 무관 (Go 코드 변경 0건)
- **Test (per OS)**: 무관 (코드 변경 0건). M5 template mirror 동기화 시 `go test ./internal/template/...`만 확인

### B6. spec-lint Heading 규약

- **본 SPEC**: § 3.3 Out of Scope에 `### Out of Scope: <topic>` h3 4개 명시 (model retirement / runtime override / backend routing / prompt caching policy). MissingExclusions 회피.

### B7. observer.go / capture path resolution

- **본 SPEC 영향도**: **무관** (runtime hook 변경 없음)

### B8. Working Tree Hygiene

- **본 SPEC**: 23 agent files + 23 template mirror + 4-locale docs = 50+ files. 무관 untracked files (예: `.moai/research/`, `.moai/state/`) 변경 금지
- **검증**: 매 milestone 종료 시 `git status --porcelain` 무관 파일 확인

### B9 (iter 2 NEW) — git remote sync 자동 실행 금지

- **본 SPEC manager-develop 위임 시 명시**: prompt에서 `git pull --rebase origin main`, `git fetch + reset` 자동 실행 **prohibition** 명시 의무 (lessons B9 from CODE-COMMENTS-EN-001 Wave 2 incident).
- **Rationale**: manager-develop autonomous git remote sync는 Wave 1 SHA rewrite로 origin/main과 mid-task divergence 유발 가능.

---

## Section C — Pre-flight Check List (착수 전 의무 검증)

manager-develop이 M1 착수 전 실행:

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current                                    # → feat/SPEC-V3R6-AGENT-MODEL-ROUTING-001
git rev-parse HEAD                                            # → 0abcda296 (or later if Sprint 1 merged)

# 2. 23-agent inventory 정확성 검증 (B1 BLOCKING)
find .claude/agents -name "*.md" -type f | wc -l              # → 23
find .claude/agents -mindepth 1 -maxdepth 1 -type d           # → core, expert, harness, meta

# 3. 현재 model 분포 baseline
grep -l 'model: inherit' .claude/agents/{core,expert,harness,meta}/*.md | wc -l    # → 21
grep -l 'model: haiku'   .claude/agents/{core,expert,harness,meta}/*.md | wc -l    # → 2 (manager-git, manager-docs)
grep -l 'model: opus'    .claude/agents/{core,expert,harness,meta}/*.md | wc -l    # → 0
grep -l 'model: sonnet'  .claude/agents/{core,expert,harness,meta}/*.md | wc -l    # → 0

# 4. Template mirror 정합성
find internal/template/templates/.claude/agents -name "*.md" -type f | wc -l       # → 23 (local과 동일)

# 5. Cross-SPEC 충돌 사전 스캔
grep -r "Retired\|TestHarnessRetirement\|superseded" .claude/agents/ || echo "no conflicts"

# 6. SPEC 5 artifacts 12-field frontmatter
for f in spec.md plan.md acceptance.md design.md research.md; do
  head -20 .moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/$f | grep -E '^(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags):' | wc -l
done
# Each → 12

# 7. constitution opus tier list 정합
grep "expert-refactoring" .claude/rules/moai/core/moai-constitution.md | head -3
# → "Effort level selection: reasoning-intensive agents (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring) → effort: xhigh or high"
```

---

## Section D — Constraints (DO NOT VIOLATE)

### D.1 PRESERVE 대상 파일 enumeration

```
PRESERVE (1 byte도 변경 금지, frontmatter `model:` 필드 외):
- .claude/agents/core/manager-{brain,develop,docs,git,project,quality,spec,strategy}.md (8)
- .claude/agents/expert/expert-{backend,devops,frontend,performance,refactoring,security}.md (6)
- .claude/agents/harness/{cli-template,hook-ci,quality,workflow}-specialist.md (4)
- .claude/agents/meta/{builder-harness,claude-code-guide,evaluator-active,plan-auditor,researcher}.md (5)
- internal/template/templates/.claude/agents/{core,expert,harness,meta}/*.md (23 mirror)
- internal/agent/ (디렉토리 존재하지 않음, 신규 작성 금지)
- internal/loader/ (디렉토리 존재하지 않음, 신규 작성 금지)
- .claude/settings.json (effortLevel 보존)
```

### D.2 무관 untracked/modified 파일 list (변경 금지)

- `.moai/harness/usage-log.jsonl` (runtime-managed)
- `.moai/state/` (runtime-managed)
- 다른 SPEC dirs (`.moai/specs/SPEC-V3R6-*` 다른 SPECs)
- `docs-site/data/menu/extra.yaml` (다른 SPEC scope)

### D.3 금지 명령

- `git commit --no-verify` (pre-commit hook bypass 금지)
- `git push --force-with-lease` to main (Hybrid Trunk Tier L = feat branch + PR)
- `git rebase --no-edit` (invalid flag)
- `git pull --rebase origin main` (manager-develop 위임 시 prompt에 명시 금지 — B9)

### D.4 사용 의무 명령

- Conventional Commits: `feat(SPEC-V3R6-AGENT-MODEL-ROUTING-001): <subject>`
- 🗿 MoAI trailer (commit message 끝)
- 4-subdir glob 사용 (`.claude/agents/{core,expert,harness,meta}/*.md`)

### D.5 Binary Constraints

- `grep -l 'model: inherit' .claude/agents/{core,expert,harness,meta}/*.md | wc -l` → **0** (REQ-AMR-008)
- `grep -l 'model: opus' .claude/agents/{core,expert,harness,meta}/*.md | wc -l` → **7** (REQ-AMR-NF-010)
- `grep -l 'model: sonnet' .claude/agents/{core,expert,harness,meta}/*.md | wc -l` → **13** (REQ-AMR-003)
- `grep -l 'model: haiku' .claude/agents/{core,expert,harness,meta}/*.md | wc -l` → **3** (REQ-AMR-004)
- Template mirror diff: `diff -q .claude/agents/<sub>/<agent>.md internal/template/templates/.claude/agents/<sub>/<agent>.md` → 0 mismatches (REQ-AMR-NF-009)

---

## Section E — Self-Verification Deliverables

각 milestone 종료 시 manager-develop이 보고 의무 항목:

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Expected Output |
|----|--------|---------------------|-----------------|
| AC-AMR-001 | PASS/FAIL | `find .claude/agents -name "*.md" -type f \| wc -l` | `23` |
| AC-AMR-002 | PASS/FAIL | `grep -l 'model: opus' .claude/agents/{core,expert,harness,meta}/*.md \| wc -l` | `7` |
| AC-AMR-003 | PASS/FAIL | `grep -l 'model: sonnet' .claude/agents/{core,expert,harness,meta}/*.md \| wc -l` | `13` |
| AC-AMR-004 | PASS/FAIL | `grep -l 'model: haiku' .claude/agents/{core,expert,harness,meta}/*.md \| wc -l` | `3` |
| AC-AMR-005 | PASS/FAIL | researcher.md batch_api opt-in key 1개 명시 | (1 of: batch_api / use_batch_api / invocation_mode) |
| AC-AMR-006 | PASS/FAIL | `wc -l .moai/state/agent-model-baseline.jsonl` | `≥ 23` |
| AC-AMR-007 | PASS/FAIL | manager-quality regression report ±5% | (report path + within bound) |
| AC-AMR-008 | PASS/FAIL | `grep -l 'model: inherit' .claude/agents/{core,expert,harness,meta}/*.md \| wc -l` | `0` |
| AC-AMR-009 | PASS/FAIL | `diff -q .claude/agents/<sub>/<agent>.md internal/template/templates/.claude/agents/<sub>/<agent>.md` for 23 pairs | `0 mismatches` |
| AC-AMR-010 | PASS/FAIL | `make build` exit 0 + template embedded.go regenerated | (exit 0) |
| AC-AMR-011 | PASS/FAIL | `moai doctor` exit 0 | `exit 0` |
| AC-AMR-012 | PASS/FAIL | constitution opus tier verbatim 정렬 | (grep PASS) |
| AC-AMR-013 | PASS/FAIL | docs-site parity ratio ≤ 1.20 | (ratio ≤ 1.20) |

### E2. Cross-Platform Build 결과 (M5 only)

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ go test ./internal/template/...         → PASS
```

### E3. Coverage 측정

- **본 SPEC**: Go 코드 변경 0건 → coverage 변경 0% (baseline preservation 검증만)

### E4. Subagent Boundary Grep

```
$ grep -rn 'AskUserQuestion' .claude/agents/ | grep -v "// " | grep -v _test.go
(baseline preservation only — 본 SPEC scope 외 변경 없음)
```

### E5. Lint Status

```
$ moai spec lint .moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/
(5 artifacts 모두 12-field canonical PASS)
```

### E6. Branch HEAD + Push 상태

- M1~M6 단계별 commit SHA 리스트
- `git push origin feat/SPEC-V3R6-AGENT-MODEL-ROUTING-001` 결과
- PR 생성 결과 (`gh pr create`)

### E7. Blocker Report

- M1 (Claude Code SDK batch_api key 확인 시 SDK 문서 변경 또는 key 미확정 시)
- M2~M4 (agent 호출 실패 시) → orchestrator에 structured 보고

---

## Milestones (M1-M6)

### M1 — Baseline Measurement + Claude Code SDK Verification (Priority: Critical)

**Goal**: 23 agents 변경 직전 baseline JSONL 생성 + Claude Code SDK batch_api canonical key 확정 + opus tier 7개 dry-run 검증.

**Tasks**:
1. orchestrator-direct (M1 ONLY): WebFetch Claude Code SDK 공식 문서로 batch_api canonical key 확인 (`batch_api` / `use_batch_api` / `invocation_mode` 중 어느 것이 표준)
2. orchestrator-direct: claude-code-guide agent 보조 호출 (선택) — 회귀 조사
3. manager-develop: 23 agents baseline JSONL 생성 (`.moai/state/agent-model-baseline.jsonl`)
   - 23 entries: `{agent_name, baseline_input_tokens_avg, baseline_output_tokens_avg, baseline_quality_score, measurement_window_start, measurement_window_end}`
   - 측정 window: 최근 7일 또는 가용 데이터 (구체 method는 research.md §A/B framework 참조)
4. manager-develop: 23 agents 모두 정상 호출 가능 dry-run (1개 sample 호출 + 응답 확인)

**Deliverables**:
- `.moai/state/agent-model-baseline.jsonl` (≥ 23 lines)
- batch_api canonical key 확정 (M2 의존, AC-AMR-005)
- 23 agents 호출 가능성 검증 보고

**AC Verified**: AC-AMR-005 (key 확정), AC-AMR-006 (baseline ≥ 23 entries)

---

### M2 — Opus Tier Migration (7 agents) (Priority: High)

**Goal**: opus tier 7 agents의 frontmatter `model:` 필드를 `inherit` → `opus`로 변경. constitution 정렬 의무.

**Tasks**:
1. 4-subdir 7 agents enumeration:
   - `core/manager-develop.md`
   - `core/manager-spec.md`
   - `core/manager-strategy.md`
   - `expert/expert-security.md`
   - `expert/expert-refactoring.md` (**constitution-aligned, iter 2 S2 적용**)
   - `meta/plan-auditor.md`
   - `meta/evaluator-active.md`
2. 각 file frontmatter `model: inherit` → `model: opus` Edit (body 보존)
3. Template mirror 동시 갱신 (`internal/template/templates/.claude/agents/<sub>/<agent>.md` 7 files)
4. M2 종료 후 7 agents 1개 sample 호출 (manager-develop 또는 plan-auditor) — fallback chain 회귀 검증
5. constitution opus tier list 직접 인용 commit message 추가

**Deliverables**:
- 7 agent files (local) + 7 mirror files = 14 files modified
- M2 commit: `feat(SPEC-V3R6-AGENT-MODEL-ROUTING-001): opus tier 7 agents migration (constitution-aligned)`

**AC Verified**: AC-AMR-002 (opus = 7), AC-AMR-009 (mirror diff = 0), AC-AMR-011 (moai doctor exit 0), AC-AMR-012 (constitution 정렬)

---

### M3 — Sonnet Tier Migration (13 agents) (Priority: High)

**Goal**: sonnet tier 13 agents의 frontmatter `model:` 필드를 `inherit` → `sonnet`로 변경.

**Tasks**:
1. 4-subdir 13 agents enumeration:
   - `core/manager-brain.md`
   - `core/manager-project.md`
   - `core/manager-quality.md`
   - `expert/expert-backend.md`
   - `expert/expert-devops.md`
   - `expert/expert-frontend.md`
   - `expert/expert-performance.md`
   - `harness/cli-template-specialist.md` (**NEW iter 2 inventory**)
   - `harness/hook-ci-specialist.md` (**NEW iter 2 inventory**)
   - `harness/quality-specialist.md` (**NEW iter 2 inventory**)
   - `harness/workflow-specialist.md` (**NEW iter 2 inventory**)
   - `meta/builder-harness.md`
   - `meta/claude-code-guide.md`
2. 각 file frontmatter `model: inherit` → `model: sonnet` Edit (body 보존)
3. Template mirror 동시 갱신 (13 files)
4. M3 종료 후 13 agents 1개 sample 호출 — fallback chain 회귀 검증

**Deliverables**:
- 13 agent files (local) + 13 mirror files = 26 files modified
- M3 commit: `feat(SPEC-V3R6-AGENT-MODEL-ROUTING-001): sonnet tier 13 agents migration (4 harness specialists)`

**AC Verified**: AC-AMR-003 (sonnet = 13), AC-AMR-009 (mirror diff = 0), AC-AMR-011 (moai doctor exit 0)

---

### M4 — Haiku Tier Migration (3 agents) + researcher Batch API opt-in (Priority: High)

**Goal**: haiku tier 3 agents 변경 + researcher batch_api opt-in 적용.

**Tasks**:
1. 4-subdir 3 agents enumeration:
   - `core/manager-docs.md` (**이미 haiku, no change** — verification only)
   - `core/manager-git.md` (**이미 haiku, no change** — verification only)
   - `meta/researcher.md` (**NEW: inherit → haiku + batch_api opt-in**)
2. researcher.md frontmatter:
   - `model: inherit` → `model: haiku`
   - M1에서 확정한 batch_api canonical key 추가 (`batch_api: true` OR `use_batch_api: true` OR `invocation_mode: batch`)
3. Template mirror 동시 갱신 (1 file: researcher.md)
4. M4 종료 후 researcher 1개 sample 호출 + 응답 시간 측정 (Batch API async wait 검증)
5. R-AMR-002 mitigation: caller pattern 변경 없음 검증

**Deliverables**:
- 1 agent file (researcher.md local) + 1 mirror file = 2 files modified (researcher만)
- M4 commit: `feat(SPEC-V3R6-AGENT-MODEL-ROUTING-001): haiku tier + researcher batch_api opt-in`

**AC Verified**: AC-AMR-004 (haiku = 3), AC-AMR-005 (batch_api key), AC-AMR-008 (inherit = 0), AC-AMR-009 (mirror diff = 0)

---

### M5 — Template Embedded Build + Inherit Count Zero Verification (Priority: Medium)

**Goal**: `make build`로 template embedded.go 재생성 + `model: inherit` count 0 검증 + cross-platform build.

**Tasks**:
1. `make build` 실행 → `internal/template/embedded.go` 갱신
2. `go test ./internal/template/...` 실행 → 23 mirror 통합 검증
3. Cross-platform build 검증:
   - `go build ./...` → exit 0
   - `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
4. AC-AMR-008 binary 검증: `grep -l 'model: inherit' .claude/agents/{core,expert,harness,meta}/*.md | wc -l` → 0
5. AC-AMR-NF-010 verification: `grep -l 'model: opus' ... | wc -l` → 7

**Deliverables**:
- `internal/template/embedded.go` 재생성 (binary regeneration)
- M5 commit: `chore(SPEC-V3R6-AGENT-MODEL-ROUTING-001): template embedded.go regeneration + inherit count 0 verification`

**AC Verified**: AC-AMR-008 (inherit = 0), AC-AMR-010 (make build exit 0), AC-AMR-NF-010 (opus = 7)

---

### M6 — docs-site 4-locale Catalog + chore (Priority: Medium)

**Goal**: docs-site agent catalog 4-locale 동시 갱신 + post-run manager-quality regression validation + chore (status: draft → implemented, version 0.2.0 → 0.3.0).

**Tasks**:
1. docs-site agent catalog 페이지 4-locale 갱신:
   - `docs-site/content/en/agents.md` (또는 catalog 페이지)
   - `docs-site/content/ko/agents.md`
   - `docs-site/content/ja/agents.md`
   - `docs-site/content/zh/agents.md`
   - 23 agents의 opus/sonnet/haiku 분류 표 작성 + Hugo Goldmark attribute (`{#agent-catalog}` 등)
2. Parity ratio ≤ 1.20 검증 (현재 baseline 1.149 from GEARS-MIGRATION-001 patterns)
3. manager-quality 위임 (post-run): REQ-AMR-007 regression bound ±5% 검증
   - 23 agents × 5 task category × N samples 측정
   - Out-of-bound 발견 시 해당 agent revert (opus 복귀)
4. chore: 5 artifacts status `draft → implemented`, version `0.2.0 → 0.3.0`
5. progress.md 작성 (M1~M6 evidence + B1~B4 BLOCKING resolution + S1~S5 SHOULD-FIX 적용 기록)

**Deliverables**:
- 4-locale agent catalog 4 files modified
- progress.md (M1~M6 evidence)
- M6 commits:
  - `feat(SPEC-V3R6-AGENT-MODEL-ROUTING-001): docs-site 4-locale agent catalog`
  - `chore(SPEC-V3R6-AGENT-MODEL-ROUTING-001): status implemented v0.3.0 + progress`

**AC Verified**: AC-AMR-007 (regression ±5%), AC-AMR-013 (parity ≤ 1.20)

---

## Cross-Sprint Coordination

### R-AMR-004 Mitigation: Sync-Phase Ordering with SPEC-V3R6-PROMPT-CACHE-001

| Phase | Order | Action |
|-------|-------|--------|
| Plan-phase | Parallel | AGENT-MODEL-ROUTING-001 + PROMPT-CACHE-001 plan-phase 동시 진행 가능 (markdown only) |
| Run-phase | **AMR first** | AGENT-MODEL-ROUTING-001 run-phase 머지 → PROMPT-CACHE-001 run-phase 진입 |
| Sync-phase | **AMR first** | AGENT-MODEL-ROUTING-001 sync PR 머지 → PROMPT-CACHE-001 sync PR (sequence enforced) |

**Why**: model 명시 → cache_write 손익분기 모델별 다름 (Sonnet 1h cache_write +100% / 1h hit -90%, Opus와 다른 break-even). PROMPT-CACHE-001 SPEC body에 "AMR 머지 후 진입" pre-condition 명시.

### Cross-SPEC depends_on / related_specs

본 SPEC frontmatter:

```yaml
depends_on: [SPEC-V3R6-RULES-PATH-SCOPE-001]
related_specs: [SPEC-V3R6-PROMPT-CACHE-001, SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001, SPEC-V3R6-BACKEND-ROUTING-001]
```

PROMPT-CACHE-001 SPEC frontmatter (개정 권장):

```yaml
depends_on: [SPEC-V3R6-AGENT-MODEL-ROUTING-001]
related_specs: [SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001]
```

---

## Out of Scope (Plan-Level)

### Out of Scope: Migration tool automation

본 SPEC는 manual frontmatter edit (Edit tool)로 진행. 향후 별도 SPEC (가칭 `SPEC-V3R7-AGENT-MIGRATION-TOOL-001`)로 `moai agent model set <agent> <tier>` CLI 명령 자동화 검토.

### Out of Scope: A/B test framework Go implementation

REQ-AMR-006 baseline JSONL은 manual measurement (M1 task). A/B test orchestration Go framework은 별도 SPEC.

### Out of Scope: GLM / Z.AI backend integration

본 SPEC는 Anthropic Claude family (Opus 4.7 / Sonnet 4.6 / Haiku 4.5) 내 tier 분배만. GLM-4.6 / Z.AI 백엔드 라우팅은 Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001 sibling work (orthogonal). 두 SPECs 합성 (예: GLM + sonnet tier)는 design.md §Layer 6 future work.
