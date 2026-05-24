---
id: SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001
title: "Anthropic Best-Practice Audit Tier 3 — Subdirectory CLAUDE.md (F3+F9) + Programmatic DRI Ownership Verification (F13)"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules + internal/spec"
lifecycle: spec-anchored
tags: "anthropic-best-practice, audit-tier-3, subdirectory-claude-md, dri-verification, ci-automation, governance"
tier: M
depends_on: [SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001]
related_specs: [SPEC-V3R6-MULTI-SESSION-COORD-001]
---

# SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 — Anthropic Best-Practice Audit Tier 3 (F3 + F9 + F13)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial creation (plan-phase). Tier M SPEC derived from `.moai/research/anthropic-best-practices-2026-05-24.md` §3 (F3 unused-skills reconnect 처리는 Tier 4 backlog로 분리) + §3 F9 (subdirectory CLAUDE.md) + §3 F13 (DRI ownership). 본 Tier 3 SPEC은 Tier 1 (F2+F4+F5, builder-harness chore, 완료) 및 Tier 2 (F1+F12, SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001, 완료)의 후속이며 Anthropic Best Practice 7 categories 중 미적용 2 categories (Category #2 subdirectory initialization + Category #7 DRI ownership at programmatic granularity)를 본 SPEC scope로 한정한다. |

---

## §A. Why this SPEC

### §A.1 Problem statement — 2 Anthropic best-practice categories가 아직 binding이 약하다

Tier 1 chore (`860fc119f`, 2026-05-24)는 governance docs 3건 정정 (F2 + F4 + F5)으로 self-check 갭을 해소했다. Tier 2 SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (`a25476e7e`, 2026-05-24)는 manager-spec / manager-develop / manager-docs agent body에 SPEC artifact ownership 매트릭스를 선언적으로 binding했고 `spec-frontmatter-schema.md`에 § Status Transition Ownership Matrix 7-row SSOT를 추가했다. 그러나 두 가지 Anthropic best-practice category가 아직 보강되지 않았다:

1. **F9 — Subdirectory CLAUDE.md 미도입** (Anthropic best-practice #2, P2). 현재 CLAUDE.md (root) + CLAUDE.local.md (root) 만 존재한다. Anthropic 권장: 주요 module마다 local CLAUDE.md를 두어 (a) Claude Code가 해당 디렉토리에서 작업할 때 local convention을 자동 로드하고 (b) 각 module의 lint/test/build 명령 scope를 명확화하며 (c) token budget을 효율화한다 (relevant context만 로드). moai-adk-go에는 5개 주요 module이 존재한다: `internal/cli/`, `internal/template/`, `internal/spec/`, `internal/hook/`, `internal/config/`.

2. **F13 — DRI ownership 명시 부분 적용** (Anthropic best-practice #7, P2). Tier 2 SPEC ARR-001은 **agent-artifact 레벨**의 DRI를 schema-level SSOT에 명시했으나, **프로그래밍 레벨에서 실제 commit이 매트릭스를 준수했는지 검증하는 자동화는 없다**. Anthropic 인용: *"You need to have an individual or a team assemble and evangelize the right Claude Code conventions. Without that work, knowledge will stay tribal and adoption will plateau."* 본 SPEC은 ARR-001의 선언적 ownership matrix를 보완하는 **검증 측면 (verification side)** — git commit 사실과 schema-level ownership matrix의 일치도를 lint-time 또는 CI-time에 확인하는 mechanism — 을 도입한다. ARR-001의 hook-based enforcement는 REQ-ARR-009로 명시 deferred 되었으므로 본 SPEC이 그 후속 작업이다.

### §A.2 Evidence — 두 가지 갭에 대한 측정 가능한 신호

#### §A.2.1 F9 신호 — 5 module 모두 local CLAUDE.md 부재

루트 cwd 기준 `find . -name CLAUDE.md -not -path "./node_modules/*"` 실행 결과 (2026-05-25):

- `./CLAUDE.md` (root, 11K)
- `./CLAUDE.local.md` (root user-only, 41K)
- 총 2개 — 모두 root 위치

주요 module (`internal/cli/`, `internal/template/`, `internal/spec/`, `internal/hook/`, `internal/config/`) 각각의 디렉토리에는 CLAUDE.md가 없다. 결과적으로 Claude Code agent가 예컨대 `internal/spec/lint.go`를 수정할 때 root CLAUDE.md만 로드되며, `internal/spec/` 도메인 특화 convention (예: `FrontmatterSchemaRule` 패턴, `Promotion` schema, `Tier S/M/L` 분류 규칙)이 즉시 surfaced 되지 않는다. 이는 (1) token budget 비효율 (관련 없는 root convention 매번 로드) (2) 도메인 무지 (agent가 module 특화 패턴을 모르고 정정 작업 진행) (3) cross-module 회귀 위험 (한 module의 변경이 다른 module의 invariant를 위반) 의 잠재 비용을 누적한다.

#### §A.2.2 F13 신호 — Ownership matrix 위반 commit 검출 메커니즘 부재

Tier 2 SPEC ARR-001은 7-row `Status Transition Ownership Matrix`를 정의했다:

| Transition | Owning agent |
|------------|--------------|
| `(none) → draft` | manager-spec |
| `draft → in-progress` | manager-develop |
| `in-progress → implemented` | manager-docs |
| `implemented → completed` | manager-docs OR orchestrator |
| `* → superseded` | manager-spec |
| `* → archived` | manager-docs |
| `* → rejected` | manager-docs (orchestrator decision) |

그러나 현재 lint 또는 CI는 다음을 검증하지 않는다: "이 commit이 spec.md frontmatter `status: in-progress → implemented`를 수행했는가? 수행했다면 commit message가 `docs(SPEC-...): sync-phase artifacts` 또는 `chore(SPEC-...): sync-phase artifacts` 패턴을 만족하는가? 패턴이 일치한다면 (manager-docs 정황) `spec.md` / `plan.md` / `acceptance.md` body 영역은 변경되지 않았는가? (manager-docs의 forbidden ownership crossing 위반 여부 — schema.md "Forbidden ownership crossings" 절)". 이 검증 부재는 다음 risk를 누적한다: (1) 신규 maintainer가 ARR-001 matrix를 모르고 작업 → 실수로 manager-docs가 spec.md body 수정 → CI는 통과 → SSOT 회귀. (2) 사용자 다중 세션 시 (예: SPEC-V3R6-MULTI-SESSION-COORD-001 staging-area race 사례) 어느 세션이 어느 transition을 수행했는지 audit trail이 commit message format alone에 의존 → 비명시적.

#### §A.2.3 Anthropic best-practice alignment 신호

`.moai/research/anthropic-best-practices-2026-05-24.md` §2.2 표 7 categories 중 5/7은 "양호 ✓"로 분류되어 있고 (Skill Progressive Disclosure / Hooks / LSP / Exploration-First Pattern 등), 2/7만 갭 — 정확히 본 SPEC이 다루는 F9 (Category #2) + F13 (Category #7) 이다. 따라서 본 SPEC 완료 후 moai-adk-go는 Anthropic best-practice 7/7 적용 완성 상태에 도달한다.

### §A.3 Cost of the gap (5 concrete failure modes)

1. **Token budget non-deterministic surge** — agent가 specific module 작업 시 무관 root convention이 통째 로드되어 context window를 점유. `internal/cli/` 작업에 `.claude/rules/moai/design/constitution.md` (frozen zones) 까지 매번 load 되는 patternal cost.
2. **Domain knowledge invisibility** — `Promotion` struct (V3R4 self-evolving harness) 와 같은 module-internal contract가 root CLAUDE.md에 surfaced 안 되어 새 agent가 `internal/spec/` 또는 `internal/cli/harness/` 수정 시 contract drift 위험 (예: L56 plan.md prose vs actual Go API drift 사례).
3. **Forbidden ownership crossing 위반 latent** — manager-docs가 sync-phase에서 spec.md body 수정 시 (ARR-001 forbids) — CI 무신호 → human-review 의존 → 검토자가 ownership matrix 모를 시 회귀.
4. **Multi-session audit trail 모호** — 동일 SPEC을 여러 세션이 동시 작업할 때 (예: COORD-001 4 race cases) 각 transition을 누가 수행했는지 commit author alone로는 불충분 (manager-spec / manager-docs 모두 `Goos Kim`이 author).
5. **Anthropic best-practice incomplete adoption** — public-facing 메시지 ("MoAI-ADK applies Anthropic best practices") 가 5/7 부분 적용 상태 — 본 SPEC 후 7/7 완성.

### §A.4 Why subdirectory CLAUDE.md + programmatic DRI verification rather than alternatives

대안 평가 (3 alternatives × 2 갭):

**F9 (subdirectory CLAUDE.md) 대안**:
- (A1) 5 module CLAUDE.md 도입 (본 SPEC 채택) — Anthropic 권장 정석, automatic loading on cwd change
- (A2) root CLAUDE.md에 module-specific 절 추가 — token bloat (현 11K → 30K+ 예상), Anthropic 권장 anti-pattern (bloated root)
- (A3) skill body에 module convention 명시 — skill loading은 conditional, automatic cwd loading 아님

선택: A1. moai-adk-go는 다중 module Go monorepo이고 각 module이 독립 contract를 가지므로 Anthropic 권장이 정확히 fit.

**F13 (DRI ownership verification) 대안**:
- (B1) spec-lint rule 확장 (본 SPEC 채택) — lint-time detection, `internal/spec/lint.go` extension, 기존 인프라 활용
- (B2) PostToolUse hook 도입 — execution-time, 그러나 hook agent attribution 어려움 (subagent_type field 직접 검사 필요), ARR-001 REQ-009로 deferred
- (B3) CI workflow (`.github/workflows/`) 추가 — GitHub Actions step, post-push, 그러나 push 후 detection은 too-late

선택: B1. lint-time이 shift-left 정석, 기존 `FrontmatterSchemaRule` 옆에 `OwnershipTransitionRule` 추가하는 cohesion 좋은 확장. B2는 ARR-001 forward-looking 으로 명시 deferred되어 본 SPEC scope 외.

### §A.5 Anthropic Best Practice alignment 완성

본 SPEC 완료 시 moai-adk-go는 Anthropic 7 categories 모두 적용 완성:

| # | Category | Pre-Tier-3 상태 | Post-Tier-3 상태 |
|---|----------|----------------|-----------------|
| 1 | Lean CLAUDE.md (layered) | 부분 적용 (root only) | **완성** (root + 5 module local) |
| 2 | Initialize in subdirectories | 미적용 | **완성** (5 module CLAUDE.md) |
| 3 | Skill Progressive Disclosure | 양호 ✓ | 양호 ✓ |
| 4 | Hooks for deterministic | 양호 ✓ | 양호 ✓ |
| 5 | LSP integration | 양호 ✓ | 양호 ✓ |
| 6 | Exploration-First Pattern | 양호 ✓ (self-check #4) | 양호 ✓ |
| 7 | DRI ownership | 선언적만 (ARR-001) | **완성** (선언적 + 프로그래밍 검증) |

---

## §B. Scope

### §B.1 In-scope (run-phase artifacts, 10 files)

**Plan-phase artifacts** (created by this SPEC's plan-phase, 4 files):

- `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/spec.md` (this file)
- `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/plan.md`
- `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/acceptance.md`
- `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/progress.md`

**Run-phase artifacts** (modified or created by future `/moai run`, 10 files):

| # | Path | Operation | Expected delta |
|---|------|-----------|---------------|
| 1 | `internal/cli/CLAUDE.md` | CREATE — CLI module local conventions (cobra root cmd, subcommand layout, `internal/cli/<group>/` packaging pattern, AskUserQuestion subagent boundary HARD) | +60-90 LOC |
| 2 | `internal/template/CLAUDE.md` | CREATE — template embedding system, `internal/template/templates/` Template-First Rule, `go:embed` boundary, 16-language neutrality | +60-90 LOC |
| 3 | `internal/spec/CLAUDE.md` | CREATE — spec linter architecture (`FrontmatterSchemaRule` / future `OwnershipTransitionRule`), AC matrix conventions, Tier S/M/L classification entry point | +60-90 LOC |
| 4 | `internal/hook/CLAUDE.md` | CREATE — hook handler conventions (`session_start.go` / `subagent_stop.go` / `post_tool_use.go`), `CLAUDE_PROJECT_DIR` resolution, observer.go path resolution (B7 issue prevention) | +60-90 LOC |
| 5 | `internal/config/CLAUDE.md` | CREATE — config section schema (`harness.yaml` / `quality.yaml` / `language.yaml` / `system.yaml`), config-merge precedence, envkey convention | +60-90 LOC |
| 6 | `internal/spec/lint.go` | EXTEND — new `OwnershipTransitionRule` struct + `Check()` method validating commit-message-vs-frontmatter-status-transition consistency | +120-180 LOC |
| 7 | `internal/spec/lint_test.go` | EXTEND — table-driven tests for `OwnershipTransitionRule` (PASS cases: each of 7 canonical transitions; FAIL cases: forbidden crossings, format mismatches) | +180-240 LOC |
| 8 | `.claude/rules/moai/development/spec-frontmatter-schema.md` | EXTEND — short cross-reference subsection pointing to `OwnershipTransitionRule` lint code + new finding code documentation | +20-30 LOC |
| 9 | `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` | MIRROR — byte-identical mirror of #8 per CLAUDE.local.md §2 [HARD] Template-First Rule | same as #8 |
| 10 | `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/progress.md` | EXTEND — run-phase evidence sections (§E.2 / §E.3 / §E.4 / §E.5) populated during run/sync/Mx phases | +60-100 LOC |

Total run-phase delta: **10 files modified or created** (5 CREATE module CLAUDE.md + 2 EXTEND `internal/spec/` lint+test + 1 EXTEND schema doc + 1 mirror + 1 progress.md). Total LOC delta: **~700-1000 LOC** — Tier M envelope (300-1000 LOC, 5-15 files acceptable).

### §B.2 Non-Goals (out of scope)

[HARD] 본 SPEC은 다음을 변경하지 않는다 (deferred to follow-up SPECs):

1. **`.claude/rules/moai/core/agent-common-protocol.md` 수정 금지** — SPEC-V3R6-MULTI-SESSION-COORD-001이 동시 run-phase 진행 중 (§Pre-Spawn Sync Check L4 scope reinforcement). 본 SPEC은 해당 파일을 read-only로 참조만 한다.
2. **`internal/governance/*` 디렉토리 손대지 않음** — SPEC-V3R6-MULTI-SESSION-COORD-001 scope. 본 SPEC은 `internal/spec/` 만 다룬다.
3. **PostToolUse hook 기반 enforcement** — ARR-001 REQ-009로 명시 deferred. 본 SPEC은 lint-time 검증만 (B1 alternative). hook 기반 enforcement는 향후 별도 SPEC.
4. **agent body sections (`.claude/agents/core/manager-{spec,develop,docs}.md`) 추가 수정** — ARR-001 M1 commit `e6ad82031`로 이미 완료. 본 SPEC은 manager body는 read-only.
5. **F3 unused skills reconnect** — Anthropic audit Tier 3 candidates 중 F3 (moai-ref-git-workflow / moai-ref-react-patterns / moai-workflow-loop 0-ref orphan) 은 별도 chore commit 또는 Tier 4 backlog로 분리. 본 SPEC scope 축소 위함 (Tier M envelope 유지).
6. **Subdirectory CLAUDE.md 추가 module** — `internal/governance/` (COORD-001 신규 생성 예정), `pkg/` (외부 export API), `cmd/moai/` (entry point) 등도 후속 CLAUDE.md 후보이나 본 SPEC scope에서 제외. F9 5-module 우선.
7. **CHANGELOG.md `[Unreleased]` 정리** — F6 Tier 4 backlog.
8. **DRI ownership documentation in README.md or CLAUDE.local.md prose** — F13 prose 차원은 별도 docs PR. 본 SPEC은 프로그래밍 검증만.

### §B.3 Out of Scope (explicit boundary clauses for plan-auditor §7 lint compliance)

#### §B.3.1 Cross-SPEC scope discipline

- **NOT MODIFYING** `.claude/rules/moai/core/agent-common-protocol.md` (COORD-001 active)
- **NOT MODIFYING** `internal/governance/*` (COORD-001 active, new package)
- **NOT MODIFYING** `.claude/agents/core/manager-*.md` body (ARR-001 already operationalized)
- **NOT MODIFYING** `.claude/rules/moai/development/agent-authoring.md` (separate SPEC if agent-frontmatter ownership field formalized)

#### §B.3.2 Hook-layer enforcement deferral

- **NOT IMPLEMENTING** PostToolUse hook validating agent-vs-transition (ARR-001 REQ-009 explicit deferral)
- **NOT IMPLEMENTING** SubagentStop hook validating commit attribution (would conflict with COORD-001 race detection scope)
- **DEFERRED TO** follow-up SPEC: `SPEC-V3R6-OWNERSHIP-HOOK-ENFORCEMENT-001` (post-Tier-3 lint-rule observation period required)

#### §B.3.3 F3 unused skills reconnect deferral

- **NOT RECONNECTING** moai-ref-git-workflow → manager-git skills:list
- **NOT RECONNECTING** moai-ref-react-patterns → expert-frontend skills:list
- **NOT INVOKING** moai-workflow-loop in /moai loop command body
- **DEFERRED TO** Tier 4 chore commit (3 simple frontmatter edits, separate from Tier 3 SPEC scope)

#### §B.3.4 Subdirectory CLAUDE.md candidate exclusions

- **NOT CREATING** `internal/governance/CLAUDE.md` (COORD-001 will own this when COORD-001 establishes the package)
- **NOT CREATING** `pkg/CLAUDE.md` (separate exported-API SPEC if needed)
- **NOT CREATING** `cmd/moai/CLAUDE.md` (low fan-out, root CLAUDE.md suffices)
- **NOT CREATING** `scripts/CLAUDE.md`, `.github/CLAUDE.md`, or any infra directory CLAUDE.md (Anthropic best-practice is module-level, not infra-level)

#### §B.3.5 DRI prose documentation deferral

- **NOT MODIFYING** `README.md` to add DRI ownership / governance maintainer section (separate docs PR)
- **NOT MODIFYING** `CLAUDE.local.md` to add DRI maintainer (user-local file, separate edit if user chooses)
- **DEFERRED TO** docs-site update or README.md follow-up — F13 prose dimension separate from programmatic verification

#### §B.3.6 Run-phase commit scope

- Implementation MUST land in **at most 3 commits** (M1 = subdirectory CLAUDE.md × 5, M2 = OwnershipTransitionRule lint, M3 = schema doc cross-ref + template mirror)
- NO cascade-style mid-run scope expansion permitted; if mid-run scope adjustment needed → blocker report + orchestrator re-delegate to manager-spec for body edit (D-NEW-1 inline-fix pattern from SIV-001 precedent)

---

## §C. Requirements (EARS format)

### §C.1 F9 — Subdirectory CLAUDE.md (REQ-AAT-001..006)

**REQ-AAT-001 (Ubiquitous)** — The 5 subdirectory CLAUDE.md files (`internal/cli/`, `internal/template/`, `internal/spec/`, `internal/hook/`, `internal/config/`) **shall** be created in run-phase and exist at the canonical paths listed in §B.1 rows #1..#5.

**REQ-AAT-002 (Event-Driven)** — **When** Claude Code agent operates in any of the 5 module directories listed in REQ-AAT-001, the system **shall** automatically load the corresponding subdirectory CLAUDE.md (per Anthropic best-practice #2 native loading behavior — no MoAI-side wiring required; the file presence is sufficient).

**REQ-AAT-003 (Ubiquitous)** — Each subdirectory CLAUDE.md **shall** contain the following sections at minimum: (a) module purpose statement (1-3 sentences), (b) key files / packages list with one-line annotations, (c) module-specific conventions (e.g., naming, error wrapping, test isolation pattern), (d) cross-references to root CLAUDE.md and related rules.

**REQ-AAT-004 (Ubiquitous)** — Each subdirectory CLAUDE.md **shall** be lean — size between 60 and 200 LOC. Bloated subdirectory CLAUDE.md (>200 LOC) violates Anthropic best-practice principle (lean layering).

**REQ-AAT-005 (Unwanted)** — Subdirectory CLAUDE.md **shall not** duplicate content already present in root CLAUDE.md. If a convention applies project-wide, it belongs in root; if module-specific, it belongs in subdirectory. Detection mechanism: `diff <root-section> <subdir-section>` should not yield >50% line overlap.

**REQ-AAT-006 (Optional)** — Where the module has language-specific conventions (e.g., Go build tags, Python virtualenv layout, TypeScript tsconfig), the subdirectory CLAUDE.md **shall** surface those conventions in a dedicated "Language & Tooling" subsection.

### §C.2 F13 — Programmatic DRI Ownership Verification (REQ-AAT-007..012)

**REQ-AAT-007 (Ubiquitous)** — A new lint rule `OwnershipTransitionRule` **shall** be implemented in `internal/spec/lint.go` with finding code `OwnershipTransitionInvalid` and severity Warning (same as `FrontmatterInvalid` for consistency).

**REQ-AAT-008 (Event-Driven)** — **When** the lint scans a SPEC directory containing both `spec.md` and a recent git commit (HEAD~3 default lookback window), the rule **shall** detect frontmatter `status:` transitions versus the commit subject pattern (per the `Status Transition Ownership Matrix` in spec-frontmatter-schema.md) and emit a finding when the commit subject pattern does not match the expected pattern for the observed transition.

**REQ-AAT-009 (Ubiquitous)** — The rule **shall** validate the 4 most-common transitions: `(none) → draft`, `draft → in-progress`, `in-progress → implemented`, `implemented → completed`. The 3 less-common transitions (`* → superseded` / `* → archived` / `* → rejected`) **may** be validated when explicitly enabled via lint config (default: enabled for all 7 transitions).

**REQ-AAT-010 (State-Driven)** — **While** the rule scans, it **shall** maintain a per-SPEC transition state map (previous-status → current-status) by parsing git log for that SPEC's frontmatter changes — using `git log --follow .moai/specs/SPEC-{ID}/spec.md -p` to extract `status:` line deltas. When the rule cannot reach git (e.g., outside git repo), the rule **shall** emit `OwnershipTransitionUnreachable` (Severity Info) and continue without blocking.

**REQ-AAT-011 (Unwanted)** — The rule **shall not** auto-modify any file. Findings are surfaced for human/CI review; no mutation is permitted. The rule is observation-only (consistent with all existing rules in `internal/spec/lint.go`).

**REQ-AAT-012 (Ubiquitous)** — The rule **shall** be covered by table-driven tests in `internal/spec/lint_test.go` with at least 7 PASS scenarios (one per canonical transition) and at least 5 FAIL scenarios (forbidden crossings, format mismatches, regex non-match, multi-transition single-commit ambiguity, unreachable-git fallback).

### §C.3 Cross-cutting / integration (REQ-AAT-013..015)

**REQ-AAT-013 (Ubiquitous)** — The schema doc `.claude/rules/moai/development/spec-frontmatter-schema.md` **shall** be extended with a short cross-reference subsection (~20-30 LOC) pointing to `OwnershipTransitionRule` (file + line range + finding-code documentation). The template mirror at `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` **shall** be byte-identical per CLAUDE.local.md §2 Template-First Rule.

**REQ-AAT-014 (Unwanted)** — The implementation **shall not** modify `.claude/rules/moai/core/agent-common-protocol.md` (active in COORD-001 scope) or `internal/governance/*` (COORD-001 scope) or `.claude/agents/core/manager-*.md` body (ARR-001 closed scope).

**REQ-AAT-015 (Ubiquitous)** — Cross-platform build **shall** pass (linux/amd64 + darwin/arm64 + darwin/amd64 + windows/amd64). Subdirectory CLAUDE.md files are markdown — no build impact. `internal/spec/lint.go` extension MUST not introduce platform-specific syscalls.

---

## §D. Acceptance Criteria summary (full matrix in acceptance.md)

본 SPEC은 **10개 mandatory AC**로 구성된다. 자세한 Given-When-Then 검증 명세는 `acceptance.md`를 참조한다.

| AC | Type | One-line summary |
|----|------|------------------|
| AC-AAT-001 | mandatory | 5 subdirectory CLAUDE.md 파일 존재 + 경로 일치 |
| AC-AAT-002 | mandatory | 각 subdirectory CLAUDE.md 60-200 LOC 범위 + 4 필수 섹션 |
| AC-AAT-003 | mandatory | root CLAUDE.md vs subdirectory diff <50% overlap |
| AC-AAT-004 | mandatory | `OwnershipTransitionRule` struct + Check() method 존재 |
| AC-AAT-005 | mandatory | 7 canonical transition lint PASS scenarios |
| AC-AAT-006 | mandatory | 5 lint FAIL scenarios (forbidden crossing + format mismatch) |
| AC-AAT-007 | mandatory | `git log --follow` parser graceful degradation (unreachable-git) |
| AC-AAT-008 | mandatory | schema doc cross-ref subsection + template mirror byte-identical |
| AC-AAT-009 | mandatory | 4-platform cross-build 통과 + go vet + golangci-lint 0 issues |
| AC-AAT-010 | mandatory | Disjoint scope verification — COORD-001 / ARR-001 / 다른 SPEC 파일 변경 0 |

---

## §E. Constitution Reference

본 SPEC은 다음 constitution / SSOT를 준수한다:

- **Anthropic Best Practice 7 categories** (외부 SSOT: `.moai/research/anthropic-best-practices-2026-05-24.md` §2.2)
- **Status Transition Ownership Matrix** (내부 SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix, ARR-001 도입)
- **Template-First Rule** (`CLAUDE.local.md` §2 [HARD])
- **16-language neutrality** (`internal/template/templates/` 하위 contents는 언어 중립)
- **manager-develop-prompt-template.md** Section A-E (Tier M REQUIRED per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier)

---

## §F. Risk and Mitigation

| Risk | Likelihood | Severity | Mitigation |
|------|-----------|----------|------------|
| 5 subdirectory CLAUDE.md prose가 root와 중복 → REQ-AAT-005 위반 | Medium | Medium | M1 작성 중 `diff` 자가 점검 + plan-auditor iter-1 review |
| `OwnershipTransitionRule` git log parser가 worktree 환경에서 fail | Low | Medium | REQ-AAT-010 graceful degradation 명시 — `OwnershipTransitionUnreachable` Info severity fallback |
| Sibling SPEC (COORD-001) 동시 변경으로 인한 conflict | Medium | High | §B.3 strict disjoint scope + run-phase pre-spawn fetch (CLAUDE.local.md §23.8 race mitigation) |
| Template mirror drift (REQ-AAT-013) | Low | Medium | M3 commit에서 `diff .claude/rules/.../schema.md internal/template/templates/.../schema.md` 자가 검증 |
| 7-transition lint rule이 false-positive 빈발 | Medium | Low | REQ-AAT-009 default-enabled subset, lint config로 opt-out 가능; observation period 후 hook 도입 (ARR-001 REQ-009 deferred path) |
| Mid-run scope expansion 유혹 (Tier M envelope 위반) | Medium | Medium | §B.3.6 명시 max 3 commit cap + blocker report + manager-spec re-delegate pattern |

---

## §G. References

| Source | Type | URL / Path |
|--------|------|-----------|
| Anthropic blog | external SSOT | https://claude.com/blog/how-claude-code-works-in-large-codebases-best-practices-and-where-to-start |
| Audit research doc | internal | `.moai/research/anthropic-best-practices-2026-05-24.md` |
| ARR-001 SPEC (Tier 2 antecedent) | internal | `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` |
| Status Transition Ownership Matrix (SSOT) | internal | `.claude/rules/moai/development/spec-frontmatter-schema.md` |
| Lint architecture | internal | `internal/spec/lint.go` (`FrontmatterSchemaRule` 패턴 참조) |
| Template-First Rule | internal | `CLAUDE.local.md` §2 [HARD] |
| Tier S/M/L classification | internal | `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier |
| manager-develop prompt template | internal | `.claude/rules/moai/development/manager-develop-prompt-template.md` |
