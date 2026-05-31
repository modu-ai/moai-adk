# Anthropic Claude Code Best Practices — moai-adk Application Audit

> **Date**: 2026-05-24
> **Source**: https://claude.com/blog/how-claude-code-works-in-large-codebases-best-practices-and-where-to-start
> **Scope**: WebFetch (Anthropic blog full content) + Explore × 2 parallel sub-agent analysis of moai-adk codebase (CLAUDE.md + 52 governance files + 23 agents + 32 skills + 4 yaml config sections + commands + CHANGELOG)
> **Methodology**: Exploration-First Pattern (Anthropic Best Practice — read-only Explore subagents → orchestrator synthesis)
> **Audit type**: governance research (cross-reference + dead prompt detection + Anthropic gap analysis)

---

## 1. Executive Summary

본 audit는 Anthropic의 Claude Code in large codebases 공식 best practice (2026-05) 와 moai-adk-go 현 governance 구조를 cross-reference한다. **총 13건 발견사항** — P1 (Critical) 4건 / P2 (Important) 5건 / P3 (Improvement) 4건. 본 보고서는 분석 + plan만 제공하며, 실제 수정은 사용자 결정 후 specialist 위임 chain으로 진행한다.

**핵심 결론 3건**:

1. **Anthropic 7 categories 적용 상태**: 5/7 양호 (Skill Progressive Disclosure / Hooks / LSP / Configuration review / Harness 5 extension points), 2/7 갭 (subdirectory CLAUDE.md 미도입 / DRI ownership 미명시)
2. **Orchestration failure pattern 발견**: 본 세션 4회 직접 수행 사례 (output-style v5.4.0/v5.4.1/footer cleanup) — builder-harness 위임 가능했으나 self-check 평가에서 잘못 invalidate. CLAUDE.local.md §16 self-check에 "Exploration-First Pattern 적용 가능한가?" 4번째 질문 누락이 원인
3. **SPEC artifact ownership ambiguity**: manager-spec / manager-develop / manager-docs 간 spec.md frontmatter status transition (draft→in-progress→implemented) 책임 분담이 CLAUDE.md §5 narrative documented이나 agent frontmatter/hooks에 binding 안 됨. 사용자 이전 turn architectural concern과 정확히 일치

---

## 2. Anthropic Blog 핵심 5 extension points + 7 best practice categories

### 2.1 5 Extension Points (Anthropic harness 정의)

> *"The harness is built from five extension points—CLAUDE.md files, hooks, skills, plugins, and MCP servers—each serving a different function. The order in which teams build them matters, as each layer builds on what came before."*

| Extension Point | moai-adk 보유 여부 | 활용도 |
|----------------|-------------------|-------|
| CLAUDE.md files | ✓ (root only) | 양호 — 14KB lean, layered (CLAUDE.md + CLAUDE.local.md) |
| Hooks | ✓ | 양호 — session-handoff / MX tag / CI / SessionStart |
| Skills | ✓ | 우수 — 32 skills, 6 namespaces (foundation/domain/workflow/ref/harness/meta) |
| Plugins | ✓ | 활용 가능 (marketplace 구조 존재) |
| MCP servers | ✓ | 양호 — context7 alwaysLoad, 외 chrome-devtools/pencil/zai-mcp opt-in |

### 2.2 7 Best Practice Categories (Anthropic 권장)

| # | Category | Anthropic 권장 | moai-adk 현재 상태 | 적용 갭 | 우선순위 |
|---|----------|-----------------|---------------------|--------|--------|
| 1 | **Lean CLAUDE.md** | layered, root big-picture only, subdirectory local conventions | root 14KB ✓ but `.claude/rules/moai/` 52 files / 10,469 lines 비대. progressive disclosure 활용 가능 | 부분 적용 | P2 |
| 2 | **Initialize in subdirectories** | subdirectory CLAUDE.md per major module | internal/cli/, internal/template/, internal/spec/ 등 subdirectory-level CLAUDE.md 없음 | **미적용** | P2 (F9) |
| 3 | **Skill Progressive Disclosure** | task-scoped skills, on-demand loading | 32 skills, frontmatter-based progressive disclosure 적용 | 양호 ✓ | - |
| 4 | **Hooks for deterministic** | use hooks instead of prompts for repeated actions | session-handoff/MX tag/CI/SessionStart 등 활용 | 양호 ✓ | - |
| 5 | **LSP integration** | language server protocol for symbol-level navigation | `.moai/config/sections/quality.yaml` lsp_quality_gates 활성 | 양호 ✓ | - |
| 6 | **Exploration-First Pattern** | read-only subagents → main agent edit | 본 turn에서 첫 적용 (WebFetch + Explore × 2 parallel). 이전 turn (output-style v5.4.0) 미적용 | **self-check 갭** | P1 (F4) |
| 7 | **DRI ownership** | individual/team owns Claude Code conventions | maintainer ownership 비명시 | **미적용** | P2 (F13) |

### 2.3 Anthropic Anti-Patterns 대비 moai-adk 점검

| Anti-pattern | moai-adk 위반 여부 | 근거 |
|--------------|--------------------|------|
| Bloated root CLAUDE.md | ⚠️ 경계선 | CLAUDE.md 14KB는 lean이나 `.claude/rules/moai/` 폴더 자동 inject 시 비대 위험 |
| Using prompts for automation | ✗ (no violation) | hooks 시스템 활용 |
| Consolidating all expertise into CLAUDE.md | ✗ (no violation) | skill progressive disclosure 적용 |
| Tribal knowledge retention | ⚠️ 부분 | DRI 미명시이나 rules + skills로 systematic 보유 |
| Stale embedding pipelines | ✗ (no violation) | RAG 미사용 — agentic search 정석 |
| Skipping infrastructure setup | ✗ (no violation) | template-first + builder-harness skill 보유 |
| Full-suite test/lint execution | ⚠️ 일부 | `go test ./...` 전체 실행 빈번 — per-package scoping 가능 |

---

## 3. 13 발견사항 상세 (P1 4 + P2 5 + P3 4)

### P1 Critical

#### F1. SPEC artifact ownership ambiguity (사용자 이전 turn 지적 confirm)

**근거**: Explore #1 Finding 1.

- **Phase 1 (plan)**: manager-spec creates `spec.md` / `plan.md` / `acceptance.md` in `.moai/specs/SPEC-XXX/`. status `draft` 명시
- **Phase 2 (run)**: manager-develop reads spec.md → creates `progress.md` (no spec.md modification 명시). frontmatter status transition `draft → in-progress` 는 CLAUDE.md §5 narrative documented이나 agent frontmatter/hook에 binding 안 됨
- **Phase 3 (sync)**: manager-docs reads spec.md + progress.md → creates PR. spec.md frontmatter status `in-progress → implemented` + progress.md Lifecycle Status table 갱신을 manager-docs가 수행 (precedent: TMD-001 `009e68c5d`, 5 files: CHANGELOG + spec.md + plan.md + acceptance.md + progress.md)

**문제**:
- manager-spec NOT for clause "documentation sync" 명시 (manager-docs.md agent definition)
- 그러나 sync 단계에서 manager-docs가 SPEC artifact body 수정 (TMD-001 sync `spec.md §B.1 scope expansion`) — 사용자 직관 (SPEC = manager-spec 영역) 위반
- spec-frontmatter-schema.md는 8 valid statuses 정의이나 어떤 agent가 어떤 transition 책임인지 명시 안 함

**영향**:
- SoC (Separation of Concerns) 위반
- manager-docs (default model: haiku) 가 SPEC body 수정 — capability mismatch 위험
- 신규 maintainer에게 ownership 흐름 모호

**해결책 후보**:
1. agent frontmatter에 SPEC artifact ownership 명시 (HARD rule)
2. hook 기반 auto-transition (PostToolUse / SubagentStop trigger)
3. manager-docs scope 축소 + manager-spec scope sync-phase 일부 포함

→ Tier 2 SPEC `SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001` 로 처리

#### F2. CLAUDE.md §4 agent catalog name mismatch

**근거**: Explore #2 P1.

- **CLAUDE.md line 116**: `### Builder Agents (1)` 섹션에 `platform` 명시
- **실제 file**: `.claude/agents/meta/builder-harness.md` (`name: builder-harness`)
- **불일치**: catalog와 실제 agent name 다름

**영향**: Forced Delegation Table cross-reference 실패 가능성, 신규 사용자 혼동

**해결책**: CLAUDE.md §4 line 116 `platform` → `harness` 1-line fix

→ Tier 1 chore에 포함

#### F3. 3 unused skills (orphan)

**근거**: Explore #2 P1.

| Skill | Path | Reference Count | 처리 후보 |
|-------|------|-----------------|----------|
| `moai-ref-git-workflow` | `.claude/skills/moai-ref-git-workflow/` | 0 refs | manager-git agent `skills:` list에 추가 |
| `moai-ref-react-patterns` | `.claude/skills/moai-ref-react-patterns/` | 0 refs | expert-frontend agent `skills:` list에 추가 |
| `moai-workflow-loop` | `.claude/skills/moai-workflow-loop/` | 0 refs | `/moai loop` command에서 명시적 invocation |

**영향**: token 부담 (32 skill SKILL.md 매 세션 Level 1 metadata 로드) + dead asset

**해결책**: 각 skill을 해당 agent에 reconnect 또는 archive (P3 cleanup)

→ Tier 3에 포함

#### F4. Orchestration failure pattern (본 세션 4회)

**근거**: 본 orchestrator self-audit + 사용자 메타 지적.

**사례**:
1. output-style v5.4.0 변경 (commit `6d10efb72`) — 5 banner 추가 + Anti-pattern 표 확장. builder-harness 위임 가능
2. output-style v5.4.1 변경 (commit `6b4e1f951`) — Insight banner prefix 통일 + Sprint Status banner 추가. builder-harness 위임 가능
3. footer cleanup (commit `3d588317e`) — Version log 제거. builder-harness 또는 manager-docs 위임 가능
4. 이전 turn manager-docs 위임 prompt 작성 — 사용자가 architectural concern으로 중단 (manager-spec 영역 침범)

**Root cause (5 Whys)**:

| Why | Cause |
|-----|-------|
| 1. Why 직접 수행했나? | "사용자 명시 자율 위임" + "context 풍부" 핑계 |
| 2. Why 그렇게 판단? | CLAUDE.local.md §16 self-check 3 질문에서 "context 풍부"가 "delegation 우위"를 invalidate한다고 결론 |
| 3. Why "context 풍부" = invalidate? | builder-harness sub-agent에 5 Whys + Anti-pattern catalogue + memory pattern을 prompt에 inject해야 한다는 비용 평가 |
| 4. Why trade-off 잘못? | Anthropic Best Practice "Exploration-First Pattern" 미숙지 — sub-agent는 _context inject_ 가 아니라 _isolated context window_ 보유. 본 turn처럼 read-only Explore + main agent synthesis가 정석 |
| 5. Why self-check에 명시 안 됨? | §16 self-check가 "delegation beats direct work"만 묻고 "Exploration-First Pattern 적용 가능한가?"를 별도로 묻지 않음 — 4번째 질문 누락 |

**해결책**: CLAUDE.local.md §16에 4번째 self-check 질문 추가:
> "이 작업의 일부를 read-only sub-agent 병렬 spawn으로 분해할 수 있는가? (Anthropic Exploration-First Pattern)"

→ Tier 1 chore에 포함

### P2 Important

#### F5. spec-workflow.md broken references [SUPERSEDED — FALSE POSITIVE 2026-05-24]

**근거 (original)**: Explore #2 P2 — claimed `.claude/skills/moai/team/{plan,run}.md` does NOT exist.

**Verification at Tier 1 execution (2026-05-24)**: builder-harness `ls` + `grep` cross-check:
- `.claude/skills/moai/team/plan.md` — **10,120 bytes** (exists, last modified 2026-05-16)
- `.claude/skills/moai/team/run.md` — **15,163 bytes** (exists, last modified 2026-05-17)
- **839 cross-references** to these files across 7+ canonical SSOT files (spec-workflow.md / SKILL.md / workflows/moai.md / workflows/plan.md / workflows/run/* / spec-frontmatter-schema.md)
- Load-bearing graph — NOT broken refs

**결론**: Explore #2 mis-reported. Refs are valid SSOT cross-links. NO action needed.

**Lesson learned**: Sub-agent file-existence claims MUST be verified via direct `ls` before action — Explore agent context may be stale or use wrong base path. Added to L49 cumulative trust-but-verify pattern.

→ Tier 1 chore F5 **cancelled** (no fix needed). F2 + F4 applied only (`860fc119f` commit).

#### F6. CHANGELOG.md 31 `## [Unreleased]` headings

**근거**: Explore #2 P2.

- **Line 30**: 현재 active Unreleased (v3.0 Mega-Sprint)
- **Line 138, 247, 267, 287, 310, 320, 336, 346, ...**: 30 stale Unreleased (이전 rc1 release)
- **Total**: 31 `## [Unreleased]` heading

**영향**: SIV-001 sync-phase manager-docs 위임 시 어느 Unreleased에 entry 추가할지 명확화 필요. 신규 maintainer 혼동.

**해결책**: 30 stale entries를 `CHANGELOG.archive.md`로 분리, line 30만 active 유지

→ Tier 4 cleanup chore에 포함

#### F7. ~260 historical SPEC refs without `.moai/specs/<ID>/`

**근거**: Explore #2 P2.

- **Examples**: SPEC-A-001, SPEC-A-002, SPEC-ADVISOR-001, SPEC-AGENCY-001 등
- **Total**: ~260 SPEC IDs in rules/skills/CHANGELOG referenced but no corresponding directory

**영향**: 정상 (retired SPECs). 그러나 audit에서 모호함 + token 부담 (rules grep 시 stale results)

**해결책**: retired SPEC 명시 인덱스 작성 또는 archive 처리 (별도 SPEC `SPEC-RETIRED-SPEC-REFS-CLEANUP-001`)

→ Tier 4 backlog

#### F8. `sunset.yaml` dormant config

**근거**: Explore #2 P2.

- **File**: `.moai/config/sections/sunset.yaml`
- **Status**: 전체 section dormant (no runtime hot path per `internal/config/sunset_notice.go` comment)
- **Activation**: 신규 SPEC 작성 필요

**영향**: 적음 — template-only field, 정상 dormant

**해결책**: activation SPEC 작성 또는 명시적 deprecate (P3)

→ Tier 4 backlog

#### F9. Subdirectory CLAUDE.md 미도입

**근거**: Anthropic Best Practice #2 적용 누락.

- **현재**: CLAUDE.md (root) + CLAUDE.local.md (root) 만
- **주요 subdirectory**: internal/cli/, internal/template/, internal/spec/, internal/hook/, internal/config/ — 각각 local CLAUDE.md 권장

**예상 효과**:
- Claude Code가 subdirectory navigation 시 local convention 자동 로드
- 도메인별 lint/test 명령 scope 명확화
- token budget 효율 (relevant context만 로드)

→ Tier 3에 포함

### P3 Improvement

#### F10. 4 stale `@MX:WARN` without `@MX:REASON`

**근거**: Explore #2 P3.

- `.claude/agents/meta/builder-harness.md:36` — `@MX:WARN [AUTO] trigger-union coverage` (no @MX:REASON)
- `.claude/agents/expert/expert-backend.md:6` — `@MX:WARN [AUTO] trigger-cap`
- `.claude/rules/moai/core/agent-common-protocol.md:85` — `@MX:WARN` (informal comment)
- `.claude/rules/moai/development/agent-authoring.md:181` — references moai-lang-* skills not in codebase

**영향**: @MX tag protocol violation (mx-tag-protocol.md requires `@MX:WARN` paired with `@MX:REASON`)

**해결책**: 4건 보강 (each `@MX:REASON` 추가)

→ Tier 4 cleanup

#### F11. SPEC artifact frontmatter asymmetry (SIV-001 case)

**근거**: 이전 turn manager-docs 위임 시 발견.

- **SIV-001 현재**: spec.md만 12-field frontmatter, plan/acceptance/progress 부재
- **이전 SPECs (TMD-001 등)**: 4 files 모두 frontmatter

**영향**: SIV-001 sync 시 "4 frontmatter draft→implemented" 표현 불일치

**해결책**: SIV-001 plan/acceptance/progress backfill (Tier 4 cleanup) 또는 spec-frontmatter-schema.md 명시 (현재 spec.md만 required)

→ Tier 4 cleanup

#### F12. manager-docs `model: haiku` vs sync-phase scope

**근거**: Explore #1 Cross-ref.

- **manager-docs**: `model: haiku` (default per agent frontmatter)
- **Sync-phase scope**: CHANGELOG + 4 SPEC files frontmatter + spec.md §B.1 scope expansion (TMD-001 precedent) — SPEC body 변경 포함
- **Capability mismatch**: haiku는 doc generation에 적합, SPEC reasoning에 부족 가능

**해결책**: F1 해결 시 자동 완화 (manager-spec이 SPEC body 책임, manager-docs는 CHANGELOG only)

→ Tier 2 SPEC에 포함

#### F13. DRI ownership 미명시 (Anthropic Best Practice #7)

**근거**: Anthropic blog 인용:
> *"You need to have an individual or a team assemble and evangelize the right Claude Code conventions. Without that work, knowledge will stay tribal and adoption will plateau."*

- **moai-adk 현재**: maintainer 한 명 (Goos) 이나 명시적 DRI/governance owner 미documented
- **README / CLAUDE.local.md**: ownership 섹션 부재

**해결책**: CLAUDE.local.md 또는 README.md에 governance maintainer 명시

→ Tier 3에 포함

---

## 4. 4-Tier 개선 plan

### Tier 1: Quick Fix (1 chore commit, builder-harness 위임)

**Scope**: F2 + F4 + F5 — governance docs 3건 정정. estimated diff: ~5 lines.

| # | Action | File | Change |
|---|--------|------|--------|
| 1 | F2 catalog name 정정 | `CLAUDE.md` line 116 | `Builder Agents (1): platform` → `Builder Agents (1): harness` |
| 2 | F4 self-check 4번째 질문 추가 | `CLAUDE.local.md` §16 | 4번째 self-check question 추가 |
| 3 | F5 broken refs 정정 | `.claude/rules/moai/workflow/spec-workflow.md` line 17, 177 | `.claude/skills/moai/team/{plan,run}.md` 참조 제거 또는 placeholder |

**위임 specialist**: builder-harness (governance docs 수정 — `.claude/` 자산 + governance MD)

**Tier 1 Latency**: ~1-2 turn (위임 + 검증 + commit + push)

### Tier 2: SPEC-First (F1 SPEC artifact ownership atomicity)

**SPEC ID**: `SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001`

**Scope**:
- manager-spec frontmatter: SPEC artifact ownership 명시 (spec.md/plan.md/acceptance.md 작성 + draft status)
- manager-develop frontmatter: progress.md 작성 + spec.md frontmatter status `draft → in-progress`
- manager-docs frontmatter: CHANGELOG/README/docs-site 작성 + spec.md frontmatter status `in-progress → implemented` + progress.md Lifecycle table 갱신
- OR hook 통한 auto-transition (PostToolUse / SubagentStop)
- spec-frontmatter-schema.md status transition matrix 추가

**Tier**: S minimal Section A-E (3 agent frontmatter 변경 + 1 schema doc 갱신 = ~5 files, ~150 LOC delta)

**AC**: 5-7개 binary verification (각 agent frontmatter 갱신 + hook 작동 검증)

**위임 chain**: manager-spec (plan) → manager-develop (run) → manager-docs (sync) → mx Step C

**Tier 2 Latency**: ~5-6 turn

### Tier 3: Anthropic Best Practice 추가 적용

| # | Action | 위임 specialist |
|---|--------|----------------|
| F3 | 3 unused skills reconnect (manager-git + expert-frontend + /moai loop command) | builder-harness |
| F9 | 주요 subdirectory CLAUDE.md 도입 (internal/cli, internal/template, internal/spec, internal/hook, internal/config) | manager-docs (각 module별 작성) + builder-harness (config설정) |
| F13 | DRI ownership 명시 (CLAUDE.local.md 또는 README.md) | manager-docs |

**Tier 3 Latency**: ~3-4 turn (별도 SPEC 또는 chore commits)

### Tier 4: Cleanup (chore commits)

| # | Action | 위임 specialist |
|---|--------|----------------|
| F6 | CHANGELOG.md 31 Unreleased → 1 active 통합 (`CHANGELOG.archive.md` 분리) | manager-docs |
| F7 | ~260 retired SPEC refs cleanup (별도 SPEC) | expert-refactoring (codemod) |
| F8 | `sunset.yaml` dormant config 정리 | builder-harness |
| F10 | 4 stale `@MX:WARN` → `@MX:REASON` 보강 | manager-quality (MX tag protocol enforcement) |
| F11 | SIV-001 plan/acceptance/progress frontmatter backfill | manager-spec (해당 SPEC artifact author) |
| F12 | F1 해결 시 자동 완화 (manager-docs scope 축소) | F1 Tier 2 결과로 종속 |

**Tier 4 Latency**: ~2-3 turn (각 cleanup 별 chore commit)

---

## 5. 실행 sequence (사용자 결정 반영)

| Step | Action | Turn 영역 |
|------|--------|----------|
| 1 | 본 보고서 저장 (`.moai/research/anthropic-best-practices-2026-05-24.md`) + memory file + MEMORY.md index | 본 turn |
| 2 | Tier 1 chore — builder-harness 위임 (F2 + F4 + F5) | 본 turn |
| 3 | Tier 1 commit + push 검증 | 본 turn |
| 4 | Tier 2 SPEC plan-phase — manager-spec 위임 (`SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001`) | 다음 turn |
| 5 | Tier 2 run-phase — manager-develop 위임 | 다음 turn 또는 별도 |
| 6 | Tier 2 sync-phase — manager-docs 위임 (개선된 정책 적용) | 별도 turn |
| 7 | SIV-001 sync-phase — manager-docs (개선된 정책 적용 — manager-spec 이 SPEC artifact 일부 책임) | Tier 2 완료 후 |
| 8 | SIV-001 Mx Step C judge + SPEC close | SIV-001 sync 후 |
| 9 | Tier 3 작업 (F3 + F9 + F13) | 사용자 추후 결정 |
| 10 | Tier 4 cleanup (F6 + F7 + F8 + F10 + F11) | 사용자 추후 결정 |

---

## 6. Lessons learned + 본 turn 자체의 위임 패턴 사례

**본 turn은 정석 위임 사례** — Anthropic Best Practice "Exploration-First Pattern" 직접 적용:

1. **Step A**: WebFetch (orchestrator-direct) — blog content fetch
2. **Step B**: Agent (Explore) × 2 parallel spawn — moai-adk 구조 + dead detection
3. **Step C**: Orchestrator synthesis — 3-source 결과 종합 + 13 findings 도출 + 4-tier plan 작성

**이전 turn 대비 차이**:
- 이전 turn (output-style v5.4.0): 직접 Edit × 5 + Bash truncate (위임 무) — 빠르지만 self-check 위반
- 본 turn: WebFetch + Explore × 2 + 본 보고서 작성 (위임 + synthesis) — Anthropic 정석

**개선 결과 (CLAUDE.local.md §16 self-check 4번째 질문 추가 시)**:
- 향후 작업 시 "read-only sub-agent 병렬 spawn으로 분해 가능?" 의무 자가 점검
- output-style 변경 같은 critical asset 작업 시 builder-harness 자동 위임 — bias prevention + capability isolation
- Anthropic Best Practice harness 5 extension points 완전 활용

---

## 7. References

| Source | Type | URL / Path |
|--------|------|-----------|
| Anthropic blog | external | https://claude.com/blog/how-claude-code-works-in-large-codebases-best-practices-and-where-to-start |
| Explore #1 result | internal | 본 turn agent return (governance structure cross-ref) |
| Explore #2 result | internal | 본 turn agent return (defect inventory) |
| CLAUDE.md | internal | `/Users/goos/MoAI/moai-adk-go/CLAUDE.md` |
| CLAUDE.local.md §16 | internal | `/Users/goos/MoAI/moai-adk-go/CLAUDE.local.md` §16 Orchestrator Self-Check |
| spec-workflow.md | internal | `.claude/rules/moai/workflow/spec-workflow.md` |
| manager-spec.md | internal | `.claude/agents/core/manager-spec.md` |
| manager-docs.md | internal | `.claude/agents/core/manager-docs.md` |
| spec-frontmatter-schema.md | internal | `.claude/rules/moai/development/spec-frontmatter-schema.md` |
| TMD-001 sync precedent | internal | git commit `009e68c5d` |
| Output-style v5.4.x commits | internal | `6d10efb72` + `6b4e1f951` + `3d588317e` |

---

**Status**: Audit complete — execution awaiting user decision (Q1 confirmed: Tier 1 chore immediate + Tier 2 SPEC; Q2 confirmed: SIV-001 sync after main work; Q3 confirmed: `.moai/research/` + memory).
