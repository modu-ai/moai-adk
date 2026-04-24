---
id: SPEC-V3R2-ORC-002
title: "Agent Common Protocol CI Lint (moai agent lint)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 3 — Agent Cleanup"
module: "internal/cli/agent_lint.go, internal/cli/cmd/agent.go, .claude/agents/moai/, internal/template/templates/.claude/agents/moai/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-ORC-001
related_problem:
  - P-A01
  - P-A04
  - P-A13
  - P-A18
  - P-A19
  - P-A23
related_theme: "Layer 4 — Orchestration, Master §4.4, §8 BC-V3R2-004, BC-V3R2-005"
breaking: true
bc_id: [BC-V3R2-004]
lifecycle: spec-anchored
tags: "agent, lint, ci, common-protocol, askuserquestion, moai-constitution, v3r2"
---

# SPEC-V3R2-ORC-002: Agent Common Protocol CI Lint (moai agent lint)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

R5 audit identifies 9 agents embedding the literal string `AskUserQuestion` in their bodies — a direct violation of `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary [HARD] rule "Subagents MUST NOT prompt the user" (problem P-A01, CRITICAL). Four agents declare a dead `Agent` tool in their tools list (subagents cannot spawn sub-subagents, P-A04). Several agents duplicate identical Skeptical-Evaluator Mandate blocks (P-A13), repeat boilerplate `--deepthink` description lines 22 times (P-A19), and carry inconsistent skill-preload sets (P-A23).

This SPEC introduces `moai agent lint` as a first-class CI lint command that scans every `.claude/agents/moai/*.md` file for:
- (i) literal `AskUserQuestion` occurrences in body text,
- (ii) `Agent` declared in the tools CSV,
- (iii) missing `effort:` frontmatter field,
- (iv) dead hook configuration (hook matcher references a tool not in the tools list),
- (v) `isolation: worktree` sanity for write-heavy role profiles (flagged as CI warning, promoted to error by SPEC-V3R2-ORC-004),
- (vi) presence of `--deepthink` description boilerplate,
- (vii) drift vs the common Skeptical-Evaluator Mandate block.

The lint ships as a Go binary subcommand (`moai agent lint`) and a CI GitHub Action step. Violations fail the CI build; exit code 0 on clean lint, exit code 1 on violation, exit code 2 on malformed frontmatter.

### 1.1 Background

`agent-common-protocol.md` lists 7 HARD rules. R5 §Common Protocol compliance table counts 9 violations of Rule 1 alone. The remediation pattern is well-defined: every "Use AskUserQuestion to clarify X" line becomes "If X is ambiguous, return a structured blocker report to the orchestrator." But enforcement requires CI guardrails; otherwise the pattern re-emerges with every new agent.

This SPEC also extracts the duplicate Skeptical-Evaluator Mandate block (appears verbatim in both `manager-quality.md` and `evaluator-active.md`, 6 bullets) into a dedicated section of `agent-common-protocol.md` labeled `§Skeptical Evaluation Stance`. Both agents then reference the common section; the duplicate is deleted. The lint checks that the 6 bullets appear at most once across the agent roster (sole occurrence in the rule file).

*Source: r5-agent-audit.md §Common Protocol compliance table; problem-catalog.md P-A01, P-A04, P-A13, P-A18, P-A19, P-A23; agent-common-protocol.md §User Interaction Boundary.*

### 1.2 Non-Goals

- Rewriting the 9 violating agent bodies beyond what ORC-001 scope already covers (ORC-001 handles roster changes; this SPEC handles lint rules only).
- Effort-level matrix design and publication (SPEC-V3R2-ORC-003 owns the matrix table; this SPEC merely checks presence of the field).
- `isolation: worktree` mandatory enforcement (SPEC-V3R2-ORC-004 upgrades the lint from warning to error for write-heavy role profiles).
- Hook handler implementation (SPEC-V3R2-RT-001 and SPEC-V3R2-RT-006).
- Modifying `.claude/rules/moai/core/agent-common-protocol.md` beyond extracting the Skeptical-Evaluator Stance section.
- Adding new subagents or skills.
- Implementing a runtime check during agent spawn; this SPEC is build-time lint only.

---

## 2. Scope (범위)

### 2.1 In Scope

- Create `internal/cli/agent_lint.go` implementing the lint engine: frontmatter parser, body scanner, rule-check pipeline.
- Register `moai agent lint` subcommand in `internal/cli/cmd/agent.go` (or current location where subcommand registration lives).
- Lint rules:
  - LR-01 reject literal `AskUserQuestion` anywhere in body or frontmatter description.
  - LR-02 reject `Agent` token in tools list.
  - LR-03 warn on missing `effort:` frontmatter field (promoted to error in ORC-003).
  - LR-04 reject dead hook entries whose matcher references a tool absent from the `tools:` list (e.g., expert-debug hook matching `Write|Edit` while tools list has no Write or Edit).
  - LR-05 warn on missing `isolation: worktree` for role profiles {implementer, tester, designer}; promoted to error by ORC-004.
  - LR-06 warn on `--deepthink` boilerplate in description field (reduction target).
  - LR-07 reject duplicate Skeptical-Evaluator Mandate block appearing more than once across the roster (canonical location: agent-common-protocol.md).
  - LR-08 warn on skill-preload drift (agents in same category with divergent required preloads).
- Extract Skeptical-Evaluator Mandate block into `.claude/rules/moai/core/agent-common-protocol.md` §"Skeptical Evaluation Stance" with the 6 canonical bullets.
- Remove the duplicate block from both `manager-quality.md` and `evaluator-active.md`; replace with a reference comment.
- Remove the dead `Agent` tool declaration from expert-security, builder-agent, builder-skill, builder-plugin (per ORC-001 those last three retire; expert-security remains and needs cleanup).
- Integrate lint into `.github/workflows/` CI as a required check.
- Emit `--format=json` output mode for IDE integrations and PR review bots.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Runtime enforcement of the lint rules during agent spawn (build-time only).
- Automatic body rewriter that fixes `AskUserQuestion` violations (manual fix + blocker-report pattern).
- Effort-level matrix content (SPEC-V3R2-ORC-003 ships the matrix; this SPEC only checks presence).
- Changing the frontmatter schema beyond presence checks (SPEC-V3-AGT-001 governs schema).
- Lint rules targeting skill files (different lint owner, SPEC-V3R2-WF-001 territory).
- Lint rules targeting command files (SPEC-V3R2-WF-002 governs thin-wrapper enforcement).
- Lint rules targeting hook wrappers (`.claude/hooks/moai/*.sh`).
- AskUserQuestion occurrences inside code examples fenced in ``` blocks that document the orchestrator's correct usage (documentation allowance, see Assumptions).
- Performance optimization of the lint engine beyond O(N) file pass.

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-alpha.3+ (Phase 3 target)
- Go version: 1.23+ (matches project go.mod)
- CI target: GitHub Actions (existing `.github/workflows/`); runs on every push + PR
- Invocation modes:
  - CLI: `moai agent lint` (default scans `.claude/agents/moai/`)
  - CLI with path: `moai agent lint --path internal/template/templates/.claude/agents/moai/`
  - CLI with JSON output: `moai agent lint --format=json`
  - CLI strict: `moai agent lint --strict` (promotes LR-03, LR-05, LR-06, LR-08 warnings to errors)
- Exit codes: 0 = clean; 1 = violation; 2 = malformed frontmatter; 3 = IO error
- Affected files: 22 current agents + 17 v3r2 final agents (both trees scanned during migration)

---

## 4. Assumptions (가정)

- SPEC-V3R2-CON-001 is landed; FROZEN/EVOLVABLE zone model is live; `.claude/rules/moai/core/agent-common-protocol.md` is classified as FROZEN for core rules and EVOLVABLE for the extracted Skeptical-Evaluator Stance section.
- SPEC-V3R2-ORC-001 merges first or in the same PR wave; its new/retired agent file deltas are visible to the lint scanner.
- `AskUserQuestion` occurrences inside Markdown code fences (```) are documentation of correct orchestrator usage and are exempt from LR-01 (scanner skips fenced-code regions).
- Every agent frontmatter uses valid YAML parsable by the existing Go YAML library (no exotic tags).
- CI runtime budget allows sub-second lint (<500ms per 22 files).
- The new `manager-cycle.md` and `builder-platform.md` (from ORC-001) pass all LR rules by construction because their templates are co-designed with this SPEC.
- Contributors running `moai agent lint` locally before push is the norm; pre-commit hook integration is optional.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-ORC-002-001 (Ubiquitous) — 바이너리 서브커맨드**
The `moai` binary **shall** expose `moai agent lint` as a subcommand that accepts flags `--path`, `--format=text|json`, `--strict`, and `--help`.

**REQ-ORC-002-002 (Ubiquitous) — 기본 스캔 경로**
Without the `--path` flag, `moai agent lint` **shall** scan both `.claude/agents/moai/*.md` and `internal/template/templates/.claude/agents/moai/*.md` relative to the repository root.

**REQ-ORC-002-003 (Ubiquitous) — 여덟 린트 규칙**
The lint engine **shall** implement exactly 8 lint rules (LR-01 through LR-08) as specified in §2.1 In Scope.

**REQ-ORC-002-004 (Ubiquitous) — 종료 코드 의미**
The binary **shall** return exit code 0 for clean runs, 1 for any LR-01, LR-02, LR-04, or LR-07 violation, 2 for malformed frontmatter, 3 for IO error, and 1 for LR-03, LR-05, LR-06, LR-08 violations ONLY when `--strict` is set (warning otherwise).

**REQ-ORC-002-005 (Ubiquitous) — 공통 프로토콜 변경**
The file `.claude/rules/moai/core/agent-common-protocol.md` **shall** gain a new section `§Skeptical Evaluation Stance` containing the 6-bullet Skeptical-Evaluator Mandate block previously duplicated in `manager-quality.md` and `evaluator-active.md`; the section **shall** be classified EVOLVABLE under the CON-001 zone model.

### 5.2 Event-Driven (이벤트 기반)

**REQ-ORC-002-006 (Event-Driven) — LR-01 위반 감지**
**When** the body scanner encounters the literal token `AskUserQuestion` outside a fenced code block, the lint engine **shall** emit a violation record with `rule: LR-01`, `severity: error`, file path, line number, and surrounding context (2 lines).

**REQ-ORC-002-007 (Event-Driven) — LR-02 위반 감지**
**When** the frontmatter parser finds the token `Agent` in the `tools:` CSV list, the lint engine **shall** emit a violation record with `rule: LR-02`, `severity: error`, file path, and exact token match.

**REQ-ORC-002-008 (Event-Driven) — LR-04 dead hook 감지**
**When** an agent declares a hook with `matcher:` field whose regex-matched tool name does not appear in the agent's `tools:` CSV, the lint engine **shall** emit a violation with `rule: LR-04`, `severity: error`, file path, hook event name, and mismatched matcher pattern.

**REQ-ORC-002-009 (Event-Driven) — LR-07 중복 블록 감지**
**When** the lint engine scans the full agent tree and finds the 6-bullet Skeptical-Evaluator Mandate block in more than one location (including `agent-common-protocol.md` as the allowed canonical location), the engine **shall** emit LR-07 violations for every occurrence beyond the first (first = the rule file).

**REQ-ORC-002-010 (Event-Driven) — JSON 출력 형식**
**When** invoked with `--format=json`, the lint engine **shall** emit a single JSON document with fields `version`, `summary: {total, errors, warnings}`, and `violations: [{rule, severity, file, line, message}]`.

### 5.3 State-Driven (상태 기반)

**REQ-ORC-002-011 (State-Driven) — CI 통합**
**While** the CI pipeline is active on push and PR events, the workflow **shall** invoke `moai agent lint` as a required status check; a non-zero exit code **shall** block PR merge.

**REQ-ORC-002-012 (State-Driven) — strict 모드 활성**
**While** `--strict` is set, the lint engine **shall** treat LR-03 (missing effort), LR-05 (missing isolation for write-heavy roles), LR-06 (--deepthink boilerplate), and LR-08 (skill-preload drift) as exit-code-1 errors instead of warnings.

### 5.4 Optional (선택)

**REQ-ORC-002-013 (Optional) — pre-commit 훅 통합**
**Where** the developer has installed a pre-commit framework, the repository's `.pre-commit-config.yaml` **shall** list `moai agent lint --path .claude/agents/moai/` as a pre-commit hook.

**REQ-ORC-002-014 (Optional) — IDE 통합 JSON**
**Where** an editor integration consumes `--format=json` output, the schema version field **shall** be stable across v3.0.0 minor versions; breaking schema changes **shall** bump the `version` field.

### 5.5 Unwanted Behavior

**REQ-ORC-002-015 (Unwanted Behavior) — fenced code 오탐지 금지**
**If** the body scanner encounters `AskUserQuestion` inside a triple-backtick fenced code block, **then** the scanner **shall** NOT emit an LR-01 violation (documentation of correct orchestrator usage is allowed).

**REQ-ORC-002-016 (Unwanted Behavior) — malformed frontmatter 처리**
**If** an agent file has invalid YAML frontmatter that cannot be parsed, **then** the lint engine **shall** emit exit code 2 with a single violation record `rule: PARSE_ERROR` naming the file and the YAML parser error message; other files in the scan set continue to be linted.

**REQ-ORC-002-017 (Unwanted Behavior) — 템플릿/로컬 드리프트 금지**
**If** the scan of `.claude/agents/moai/*.md` produces a different violation set than the scan of `internal/template/templates/.claude/agents/moai/*.md`, **then** the lint engine **shall** emit warning `LINT_TREE_DRIFT` listing the diverging files (the CI gate from CLAUDE.local.md §2 Template-First catches byte drift; this lint flag specifically catches semantic drift introduced by rule skew).

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-ORC-002-01**: `moai agent lint --help` prints the subcommand usage with all documented flags.
- **AC-ORC-002-02**: Running `moai agent lint` on a tree containing the v2.13.2 roster (22 agents, 9 violations per R5) exits with code 1 and reports exactly the 9 LR-01 violations listed in R5 §Common Protocol compliance table.
- **AC-ORC-002-03**: Running `moai agent lint` on the v3r2 post-consolidation roster (17 agents produced by ORC-001) exits with code 0.
- **AC-ORC-002-04**: `moai agent lint --format=json` output is valid JSON with the documented schema; `jq '.summary.errors'` matches the count from text output.
- **AC-ORC-002-05**: Introducing a fresh `AskUserQuestion` line in any agent body causes CI to fail with exit code 1; CI-status check becomes red.
- **AC-ORC-002-06**: Introducing a dead-hook configuration (matcher `Write` with no Write in tools) triggers LR-04 with exit code 1.
- **AC-ORC-002-07**: Adding the 6-bullet Skeptical block to a second agent body triggers LR-07.
- **AC-ORC-002-08**: In non-strict mode, a missing `effort:` field emits a warning only; exit code 0; violation count via JSON shows warnings non-zero.
- **AC-ORC-002-09**: In `--strict` mode, the same missing `effort:` causes exit code 1.
- **AC-ORC-002-10**: Fenced code blocks containing `AskUserQuestion` are not flagged; negative test fixture confirms.
- **AC-ORC-002-11**: Malformed YAML frontmatter triggers exit code 2; other files continue to be linted.
- **AC-ORC-002-12**: `.claude/rules/moai/core/agent-common-protocol.md` contains exactly one occurrence of the Skeptical-Evaluator Mandate 6-bullet block.

---

## 7. Constraints (제약)

- [HARD] FROZEN constitution preservation (Master §1.3, CON-001): `agent-common-protocol.md` §User Interaction Boundary remains FROZEN; this SPEC adds §Skeptical Evaluation Stance as EVOLVABLE but does not modify the FROZEN section.
- [HARD] Template-First (CLAUDE.local.md §2): agent file edits land under `internal/template/templates/` first; lint scans both trees.
- [HARD] No false positives on fenced code (REQ-015).
- [HARD] Lint must complete under 500ms on full roster scan for CI budget.
- [HARD] No network calls during lint (reproducible CI).
- [HARD] Go-only binary; no external tools (no ripgrep dependency; use Go regex + bufio.Scanner).
- [HARD] Must support both `/*.md` extensions and frontmatter parsing compatible with existing template/deployer.go code.
- [HARD] JSON schema version field is stable through v3.0.0 minor releases.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                           | Impact | Mitigation                                                                                         |
|----------------------------------------------------------------|--------|----------------------------------------------------------------------------------------------------|
| LR-01 false positive inside fenced code confuses contributors | HIGH   | REQ-015 explicit exemption; regression test fixtures in _test.go                                   |
| LR-07 canonical-block detection false-positive on paraphrases | HIGH   | Fingerprint the 6-bullet block by normalizing whitespace + bullet markers; match semantic identity |
| Contributors bypass lint with `# nolint` markers              | MEDIUM | No nolint escape hatch for LR-01 (too safety-critical); warnings have no escape hatch either       |
| Template vs local tree drift reintroduces violations in one tree | MEDIUM | REQ-017 LINT_TREE_DRIFT warning + CI template-first diff check                                    |
| JSON schema changes break IDE integrations                    | MEDIUM | Version field per REQ-014; semver for schema                                                       |
| Pre-commit integration too slow for tight loops               | LOW    | Scoped scan via `--path` flag; parallelize file scanning                                           |
| agent-common-protocol.md amendment triggers CON-002 gate      | MEDIUM | §Skeptical Evaluation Stance is EVOLVABLE; amendment goes through graduation protocol              |
| Lint reports noisy diffs on v2 → v3 transition                | LOW    | Two-tree drift warning is non-blocking; clear after ORC-001 lands                                  |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification) — zone model prerequisite for classifying the new §Skeptical Evaluation Stance section.
- **SPEC-V3R2-ORC-001** (Agent roster consolidation) — the v3r2 roster is the baseline the lint validates against.

### 9.2 Blocks

- **SPEC-V3R2-ORC-003** (Effort matrix) — ORC-003 promotes LR-03 from warning to error after effort is populated.
- **SPEC-V3R2-ORC-004** (Worktree MUST) — ORC-004 promotes LR-05 from warning to error.
- **SPEC-V3R2-MIG-001** (v2→v3 migrator) — MIG-001 runs `moai agent lint` after rewriting legacy SPEC references.

### 9.3 Related

- **SPEC-V3R2-CON-002** (Constitutional amendment protocol) — §Skeptical Evaluation Stance section addition runs through the 5-layer safety gate.
- **SPEC-V3R2-CON-003** (Constitution consolidation) — touches the same rules file but does not overlap this SPEC's extraction scope.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01; REQ-002 → AC-02, AC-03; REQ-003 → AC-02..AC-10; REQ-004 → AC-08, AC-09, AC-11; REQ-005 → AC-12; REQ-006 → AC-02, AC-05, AC-10; REQ-007 → LR-02 regression test in acceptance.md; REQ-008 → AC-06; REQ-009 → AC-07, AC-12; REQ-010 → AC-04; REQ-011 → AC-05; REQ-012 → AC-08, AC-09; REQ-013 → pre-commit fixture in acceptance.md; REQ-014 → version-field regression; REQ-015 → AC-10; REQ-016 → AC-11; REQ-017 → two-tree drift fixture.
- Total REQ count: 17 (Ubiquitous 5, Event-Driven 5, State-Driven 2, Optional 2, Unwanted 3)
- Expected AC count: 12
- Wave 1/2 sources:
  - `r5-agent-audit.md` §Common Protocol compliance table (9 violations enumerated)
  - `r5-agent-audit.md` §Tool scope audit (4 dead `Agent` tool declarations)
  - `r5-agent-audit.md` §Additional systemic findings (dead hook config, --deepthink boilerplate, skill preload drift)
  - `problem-catalog.md` P-A01 (CRITICAL), P-A04, P-A13, P-A18, P-A19, P-A23
  - `pattern-library.md` X-1 (Markdown + YAML Frontmatter as contract → enables lint)
  - `agent-common-protocol.md` §User Interaction Boundary (FROZEN rule protected by this SPEC)
  - `major-v3-master.md` §4.4 Layer 4 Orchestration, Appendix A (P-A01 maps to ORC-001 + ORC-002)
- Code-side paths:
  - `internal/cli/agent_lint.go` (new, REQ-001, REQ-003, REQ-006..009)
  - `internal/cli/agent_lint_test.go` (new, REQ-015, REQ-016, all AC regression fixtures)
  - `internal/cli/cmd/agent.go` (modified, REQ-001 subcommand registration)
  - `.claude/rules/moai/core/agent-common-protocol.md` (modified, REQ-005)
  - `.claude/agents/moai/manager-quality.md` (modified — remove duplicate Skeptical block, reference common)
  - `.claude/agents/moai/evaluator-active.md` (modified — remove duplicate Skeptical block, reference common)
  - `.claude/agents/moai/expert-security.md` (modified — drop dead `Agent` tool per P-A04)
  - `.github/workflows/ci.yaml` (modified, REQ-011)
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1053` (§11.4 ORC-002 definition)
  - `docs/design/major-v3-master.md:L963` (§8 BC-V3R2-004)
  - `docs/design/major-v3-master.md:L964` (§8 BC-V3R2-005)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-A01, P-A04, P-A13, P-A19)

---

End of SPEC.
