---
id: SPEC-CLAUDEMD-DIET-V2-001
title: "CLAUDE.md 2nd-round diet (405 → ~300 lines, rule-SSOT pointer-ization + §16 path-scoped extraction)"
version: "0.1.1"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "CLAUDE.md + internal/template/templates/CLAUDE.md"
lifecycle: spec-anchored
tags: "claude-md, diet, always-loaded, context-budget, template-first, rule-ssot, pointer-ization, official-docs-audit-2026-07"
era: V3R6
tier: M
---

## HISTORY

- 2026-07-08 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier M, 4 artifacts: spec.md + plan.md + acceptance.md + progress.md). Track 2 of the official Claude Code docs audit 2026-07 (memory: `project_official_docs_audit_2026_07`, session 2a40de12). Independent SSOT-verification uncovered one major surprise: §16 Context Search Protocol has NO existing rule-SSOT (unlike §5/§8/§11/§14/§15) — extracting it requires creating a new path-scoped rule file, not a simple pointer-ization. Realistic line-count ceiling derived at ~300L (not the official 200L aspiration), because the diet respects the HARD constraint that §1/§3/§4 canonical content is not movable. `status: draft`.
- 2026-07-08 — v0.1.1 — manager-spec — iter-2 audit revision (plan-auditor iter-1 FAIL 0.81, 1 BLOCKING + 5 recommended). D1 (BLOCKING): AC-CMD2-003 regex bug — OLD `^@import` matched 0 lines (the actual syntax is Obsidian-style `@<path>` embed, NOT markdown `@import`); fixed regex to `^@\.moai/config/sections/(user|language)\.yaml$` across spec.md + acceptance.md + plan.md. D2 (SHOULD-FIX, verification-claim-integrity §1.1 surface 2): AC-CMD2-008 command form was never test-run; corrected `moai spec lint SPEC-CLAUDEMD-DIET-V2-001` (ParseFailure) → `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` (실측 `✓ No findings`); ran ALL 10 AC verification commands this iteration, verbatim observed outputs recorded in progress.md §E.1. D3 (SHOULD-FIX, scope-honesty): §A.4 reframed "200L impossible" → "200L scope choice" with theoretical-pointer option disclosed (§1's 11-bullet HARD list is partially pointer-ized to `moai-constitution.md` already; further pointer-ization is theoretically possible but out of scope). D4 (MINOR): AC-CMD2-010 baselines corrected to awk-measured §1=29 / §3=12 / §4=50 with ±1 tolerance. D5 (MINOR): §D Out-of-Scope heading "§1/§3/§4" → "§1/§2/§3/§4" reconciled with body. D6 (MINOR): SPEC-V3R5-CLAUDE-REFRESH-001 "superseded scope" → "completed" (status 확인 실측). Iter-1 plan-auditor confirmed §16 SSOT absence (MP-2 PASS), section arithmetic, 1st-round collision, 3rd-round out-of-scope declaration, 10 REQ GEARS compliance — all settled, not re-litigated. `status: draft`.

---

## A. Context / Background

### A.1 2nd-round diet trigger

공식 Claude Code 문서 감사(2026-07, memory `project_official_docs_audit_2026_07`)는 CLAUDE.md가 405라인으로 공식 권장(약 200라인)의 약 2배임을 식별했다. 이는 Track 2 P1 백로그로 지정되었다. 1st-round 다이어트(`SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001`, completed, 2026-06-22)가 650→409라인(−241라인)을 달성한 후, 본 2nd-round는 동일한 파일의 후속 다이어트를 다룬다 — 1st-round를 re-open하는 것이 아니다.

### A.2 1st-round와의 관계

`SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001`은 `status: completed`이며 본 SPEC은 그 연장선에 있다. 그러나 범위는 겹치지 않는다:

- 1st-round는 per-line-test deletion + changelog footer 제거 + 기존 SSOT로의 pointer-화를 다루었다.
- 2nd-round(본 SPEC)는 1st-round가 보존하기로 결정한 섹션들 중 6개 후보(§5, §8, §11, §14, §15, §16)를 재검토하여 추가로 축소한다.
- 1st-round iter-2 감사(D1)는 §16을 "0 SSOT hits → KEEP"으로 분류했었다 — 이 전례는 본 SPEC이 §16 처리를 위해 새 SSOT 룰 파일 생성을 요구하는 근거다. 본 SPEC은 1st-round를 정면으로 부정하지 않고, 1st-round가 남긴 전제("SSOT가 생기면 pointer-화 가능")를 실행한다.

### A.3 SSOT-verification surprise (plan-phase 독립 검증 결과)

사전 검증에서 6개 후보의 추출 대상 SSOT를 독립 확인한 결과, 하나의 주요 불일치를 발견했다:

| Section | Lines | User-identified SSOT | Actual SSOT exists? | Extraction type |
|---------|-------|----------------------|---------------------|-----------------|
| §16 Context Search Protocol | 48 | (user suggested `context-search.md`) | **NO** — file does not exist | New rule creation OR in-place compress |
| §15 Agent Teams | 43 | `spec-workflow.md` + `dynamic-workflows.md` | YES (18KB + 18KB) | Pointer-ization |
| §5 SPEC-Based Workflow | 35 | `spec-workflow.md` (33KB) | YES | Pointer-ization |
| §11 Error Handling | 19 | `agent-common-protocol.md` § Error Recovery | YES | Pointer-ization |
| §14 Parallel Execution | 16 | `agent-common-protocol.md` § Parallel Execution | YES | Pointer-ization |
| §8 User Interaction | 9 | `askuser-protocol.md` | YES | Pointer-ization |

**핵심 발견**: §16 L339의 기존 참조 줄은 `context-window-management.md`와 `session-handoff.md`를 가리키지만, 이들은 §16의 "When to Search / Search Process / Token Budget" 콘텐츠와 **다른 우려**(context window 임계값 + paste-ready resume 형식)를 다룬다. §16 자체의 콘텐츠에 대한 SSOT는 존재하지 않는다. 따라서 §16 추출은 다른 5개 후보와 근본적으로 다른 작업이다.

### A.4 현실적 라인 목표 도출 (scope choice, NOT physical impossibility — D3 fix iter-2)

200라인 목표(@official recommendation)는 본 SPEC의 **scope choice**이다. §1 Core Identity / §2 Request Processing / §3 Command Reference / §4 Agent Catalog는 CLAUDE.md 자체의 canonical home으로 보호한다(D.4 참조). 이것이 물리적 불가능인 것은 아니다 — §1의 11-bullet HARD list는 이미 `.claude/rules/moai/core/moai-constitution.md`(287L, 실측 2026-07-08)로 부분 pointer-화되어 있어, 추가 pointer-화는 이론적으로 가능하다. 그러나 본 SPEC은 canonical content를 보존하기로 **선택**했으므로, 후보별 감소량 합산은:

- §16: 48L → 5L pointer = −43L (Option B, 신규 룰 파일 생성) OR 48L → 15L = −33L (Option A, in-place 압축)
- §15: 43L → 20L = −23L (CG Mode ASCII + Dynamic Workflows prose → pointer)
- §5: 35L → 15L = −20L (Agent Chain → pointer)
- §11: 19L → 8L = −11L
- §14: 16L → 8L = −8L
- §8: 9L → 5L = −4L

합산: −109L (§16 Option B) 또는 −99L (§16 Option A). 405 − 109 = **296L** (Option B), 405 − 99 = **306L** (Option A). 현실적 MUST 목표를 ≤ 320L로, SHOULD 목표를 ≤ 210L(공식 권장)로 설정한다. 200L는 후속 3rd-round 다이어트(본 SPEC out of scope)에서 다룬다.

### A.5 Track 2 of official-docs-audit-2026-07

본 SPEC은 4-track 감사 중 Track 2 단일 항목이다. 다른 트랙:
- Track 1: agent-memory orphan 정리 (20건) — 별도 SPEC
- Track 3: agentic_loop Go loader — 별도 SPEC (본 SPEC out of scope)
- Track 4: optional 후속 — 별도 판단

---

## B. Requirements (GEARS notation)

### B.1 핵심 다이어트 의무 (Ubiquitous)

**REQ-CMD2-001** (Ubiquitous) — The CLAUDE.md file shall have its line count reduced from the 2026-07-08 measured baseline of 405 lines to a post-diet ceiling of ≤ 320 lines (MUST), targeting ≤ 210 lines (SHOULD, official Claude Code best-practice alignment).

### B.2 기존 SSOT 기반 pointer-화 (When + capability gate)

**REQ-CMD2-002** (Event-driven) — **When** a CLAUDE.md section's distinctive prose content is verifiably duplicated in an existing path-scoped rule file under `.claude/rules/moai/**`, the diet shall reduce that section to at most a one-line prose pointer to the SSOT rule file.

**REQ-CMD2-003** (Capability gate) — **Where** a section's distinctive content has no existing rule-SSOT (specifically §16 Context Search Protocol as of 2026-07-08), the diet shall NOT collapse the section to a pointer until a new path-scoped rule file is created as the canonical home, OR the section is compressed in-place without losing behavioral coverage.

### B.3 §16 특별 처리 (신규 SSOT 생성)

**REQ-CMD2-004** (Compound) — **Where** the §16 Context Search Protocol content is moved out of CLAUDE.md **While** template-internal-content-isolation doctrine (CLAUDE.local.md §25) applies, the diet shall create a new rule file at `.claude/rules/moai/workflow/context-search.md` with a `paths:` frontmatter restricting load to relevant contexts (e.g., `paths: "**/.claude/projects/**"` or on-demand) so that the extraction actually reduces the always-loaded token sum, NOT merely relocates content to another always-loaded file.

### B.4 Template-First 동기화 (State-driven)

**REQ-CMD2-005** (State-driven) — **While** both `CLAUDE.md` (local) and `internal/template/templates/CLAUDE.md` (template source) exist as byte-identical mirrors (verified 2026-07-08: `diff` exit 0), the diet shall apply every edit to BOTH files in lockstep and verify byte-parity at every milestone commit.

### B.5 행위 불변 (Unwanted behavior)

**REQ-CMD2-006** (Unwanted behavior) — The CLAUDE.md diet shall not alter any agent-observable behavior, including: agent routing, command set semantics, frontmatter-status transitions of other SPECs, the 8-retained-agent catalog, the archived-agent rejection list, or the command reference table.

**REQ-CMD2-007** (Unwanted behavior) — The diet shall not remove any `[HARD]`-tagged rule, `[ZONE:*]`-tagged policy marker, or the `@<path>` embed directives at §9 (`@.moai/config/sections/user.yaml` + `@.moai/config/sections/language.yaml`, Obsidian-style transclusion — NOT markdown `@import`; D1 fix iter-2) — these are load-bearing and were preserved through the 1st-round diet (precedent: 1st-round AC-CMD-003).

### B.6 Template 중립성 (Capability gate)

**REQ-CMD2-008** (Capability gate) — **Where** a new rule file is created under `internal/template/templates/.claude/rules/moai/workflow/context-search.md`, the rule file shall be content-neutral per §25 template-internal-isolation doctrine — no SPEC IDs, no internal dates, no commit SHAs, no moai-adk-internal development references — so that the CI guard `internal/template/internal_content_leak_test.go` continues to PASS.

### B.7 Always-loaded 예산 감소 (Ubiquitous)

**REQ-CMD2-009** (Ubiquitous) — The diet shall reduce the always-loaded token sum attributable to CLAUDE.md content, measured by line count reduction AND by confirming that any newly-created extraction-target rule file is path-scoped (NOT always-loaded itself).

### B.8 파생 목표 (range, not hard number)

**REQ-CMD2-010** (Ubiquitous) — The diet shall derive its post-edit target from the per-section reduction arithmetic documented in plan.md §C.4, not from an aspirational hard number.

---

## C. Constraints

1. **Template-First (HARD)**: `CLAUDE.md` + `internal/template/templates/CLAUDE.md` must stay byte-identical after every milestone. Plan orders edits template-first, then live mirror.
2. **render-SSOT preservation**: §1 Core Identity (HARD rules list), §3 Command Reference, §4 Agent Catalog table, §2 Request Processing pipeline — these are canonical to CLAUDE.md itself and MUST NOT be extracted. Only extract content whose canonical SSOT exists in a rule file.
3. **§25 Template Internal-Content Isolation**: any new rule file under `internal/template/templates/.claude/rules/` must pass CI guard `internal/template/internal_content_leak_test.go` (no SPEC IDs, internal dates, SHAs).
4. **Always-loaded budget discipline**: extracting to an always-loaded rule file would defeat the purpose. New rule files MUST declare `paths:` frontmatter or otherwise load on-demand.
5. **No behavior change**: prose/pointer refactoring only. No agent routing, no command semantics, no frontmatter-status transitions of other SPECs.
6. **Pre-existing content correctness**: any content corrections discovered during the diet (e.g., a reference that already points to a stale path) are recorded as run-phase findings, NOT silently fixed in the same edit — the diet is not a license to refactor behavior.
7. **@embed honesty**: the 2 `@<path>` embed lines (§9 `@.moai/config/sections/user.yaml` + `@.moai/config/sections/language.yaml`, Obsidian-style transclusion) are structure-only and contribute 0 to the token reduction; they MUST NOT be counted toward the line-reduction attribution (1st-round precedent AC-CMD-004).

---

## D. Out of Scope / Exclusions

### Out of Scope — §1/§2/§3/§4 canonical content extraction (D5 fix iter-2: heading reconciled with body)

- §1 Core Identity, §2 Request Processing pipeline, §3 Command Reference, §4 Agent Catalog table은 CLAUDE.md 자체의 canonical home이다. 본 SPEC은 이들을 이동하거나 pointer-화하지 않는다. 공식 200라인 권장을 위해 이들을 건드리는 것은 3rd-round 다이어트(별도 SPEC)의 판단이다. (D3 scope-honesty: §1 HARD list의 `moai-constitution.md` 추가 pointer-화는 이론적으로 가능하나 본 SPEC 범위 밖임.)

### Out of Scope — 8-retained-agent catalog 또는 archived-agent 리스트 수정

- CLAUDE.md §4의 8-retained-agent 표와 archived-agent 리스트(12개)는 본 SPEC의 범위 밖이다. 이들은 agent 카탈로그 SSOT이며 다이어트가 아닌 agent-team-rebuild 변경 사안이다.

### Out of Scope — 행위 변경 (behavior change)

- agent routing, command semantics, frontmatter-status 전이 규칙, HARD rule 텍스트의 의미 변경 — 어느 것도 본 SPEC 범위가 아니다. 본 SPEC은 prose/pointer refactoring만 수행한다.

### Out of Scope — Track 3 agentic_loop Go loader

- 공식 문서 감사의 Track 3 (agentic_loop Go loader)은 별도 SPEC이다. 본 SPEC은 CLAUDE.md 다이어트만 다룬다.

### Out of Scope — 40-item template-artifact audit (CRITICAL CMD-14)

- CLAUDE.md 외의 템플릿 자산 40건 감사 항목은 별도 백로그(CRITICAL CMD-14)이며 본 SPEC 범위가 아니다.

### Out of Scope — 1st-round SPEC re-open

- `SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001`은 completed 상태이며 본 SPEC은 그것을 re-open하지 않는다. 1st-round가 보존하기로 결정한 섹션들(§16 포함)을 후속 다이어트로 다루되, 1st-round의 AC를 건드리지 않는다.

### Out of Scope — 3rd-round 다이어트 (200L 달성 시도)

- 공식 권장 200라인 달성을 위한 추가 감소는 §1-§4 canonical 훼손 또는 agent-team-rebuild 수준의 구조 재편이 필요하며, 본 SPEC의 후속으로 판단 보류한다. 본 SPEC은 현실적 ceiling ~300L를 목표로 한다.

### Out of Scope — Go 코드 / lint 엔진 / spec-lint 규칙 수정

- 본 SPEC은 CLAUDE.md와 (필요시) 신규 룰 파일 `.claude/rules/moai/workflow/context-search.md`만 다룬다. `internal/spec/lint.go` 또는 다른 Go 코드 수정은 범위 밖이다.

---

## E. Acceptance Criteria Reference

상세 AC는 `acceptance.md`의 §D 매트릭스를 참조. 핵심 AC 토큰:

- **AC-CMD2-001**: `wc -l CLAUDE.md ≤ 320` (MUST), `≤ 210` (SHOULD)
- **AC-CMD2-002**: `diff CLAUDE.md internal/template/templates/CLAUDE.md` exit 0 (byte-parity)
- **AC-CMD2-003**: `[HARD]` count ≥ 14 preserved (1st-round ceiling), `[ZONE:*]` count preserved, `@<path>` embed count == 2 preserved (Obsidian-style transclusion `@.moai/config/sections/{user,language}.yaml`, NOT markdown `@import` — D1 fix iter-2)
- **AC-CMD2-004**: always-loaded sum reduction — 새 룰 파일이 path-scoped인지 확인
- **AC-CMD2-005**: no behavior change (agent catalog, command set, archived-agent list unchanged)
- **AC-CMD2-006**: §16 extraction completeness (Option A or B 실행 증거)
- **AC-CMD2-007**: template neutrality (CI guard `TestTemplateNeutralityAudit` PASS, `internal_content_leak_test.go` PASS)
- **AC-CMD2-008**: GEARS lint clean (`moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` — D2 fix iter-2: subcommand expects file path, not SPEC-ID)
- **AC-CMD2-009**: distinctive-content grep (POINTER 처리된 섹션의 distinctive 콘텐츠가 실제로 SSOT에 존재하는지 재확인)
- **AC-CMD2-010**: §1/§3/§4 negative scope guard (canonical content line count 보존)

---

## F. Evidence (verification-claim-integrity)

### F.1 사전 검증 where observed (2026-07-08, this session — iter-2 re-verified)

- `wc -l CLAUDE.md internal/template/templates/CLAUDE.md` → `405 / 405 / 810 total`
- `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo $?` → DIFF_EXIT=0 (byte-parity)
- `ls .claude/rules/moai/workflow/ | grep context-search` → no match (file does not exist)
- `find .claude/rules internal/template/templates/.claude/rules -iname '*context*search*'` → no match (confirmed §16 SSOT absence; iter-1 plan-auditor CONFIRMED MP-2)
- `cat .moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/spec.md | head -15` → `status: completed` (1st-round not re-opened)
- `ls .moai/specs/ | grep -iE 'claudemd-diet|claudemd-v2'` → only `SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001` (no collision)
- `grep -n '^## ' CLAUDE.md` → §1..§17 section line numbers verified (§16 L337-385 = 48L, §15 L293-336 = 43L, §5 L126-161 = 35L, §11 L235-254 = 19L, §14 L276-292 = 16L, §8 L196-205 = 9L)
- `ls .claude/rules/moai/{workflow,core}/*.md` → all 4 SSOT files exist (spec-workflow.md 33KB, dynamic-workflows.md 18KB, agent-common-protocol.md 29KB, askuser-protocol.md 25KB)
- `grep -c '\[HARD\]' CLAUDE.md` → 14 (실측)
- `grep -c '\[ZONE:' CLAUDE.md` → 14 (실측)
- `grep -c '^@import' CLAUDE.md` → **0** (D1 bug proven — OLD regex matched nothing)
- `grep -cE '^@\.moai/config/sections/(user|language)\.yaml$' CLAUDE.md` → **2** (D1 fix verified)
- `awk '/^## 1\. Core Identity/,/^## 2\./' CLAUDE.md | wc -l` → 29 (D4 actual; iter-1 plan-audit D4 corrected baseline)
- `awk '/^## 3\. Command Reference/,/^## 4\./' CLAUDE.md | wc -l` → 12
- `awk '/^## 4\. Agent Catalog/,/^## 5\./' CLAUDE.md | wc -l` → 50
- `awk '/^## 2\. Request Processing/,/^## 3\./' CLAUDE.md | wc -l` → 36 (D5 — §2 confirmed canonical, 34L body)
- `grep -c '^| \`manager-' CLAUDE.md` → 4 (D4-corollary: iter-1 spec had erroneous baseline 5)
- `grep -c 'Use the.*subagent' CLAUDE.md` → 11 (D4-corollary: iter-1 spec had erroneous baseline 9)
- `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md` → `✓ No findings — all SPEC documents are valid` (D2 fix verified — file-path form works)
- `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 0.538s` (AC-CMD2-007 baseline PASS)
- `go test ./internal/template/ -run TestInternalContentLeak -count=1` → `ok ... [no tests to run]` (test name not found — see AC-CMD2-007 note)
- `head -15 .moai/specs/SPEC-V3R5-CLAUDE-REFRESH-001/spec.md` → `status: completed` (D6 — NOT superseded)
- AC-CMD2-009 distinctive-content baseline grep (precondition data):
  - `grep -c 'Phase 1.*plan-phase.*manager-spec' spec-workflow.md` → **0** (§5 Agent Chain distinctive content NOT in SSOT — M2 risk)
  - `grep -c 'Error Recovery Pattern' agent-common-protocol.md` → 1 (§11 OK)
  - `grep -c 'Parallel Execution' agent-common-protocol.md` → 1 (§14 OK)
  - `grep -c 'CG Mode\|moai cg' spec-workflow.md dynamic-workflows.md` → **0 in both** (§15 CG Mode distinctive content NOT in SSOT — M2 risk)
  - `grep -c 'AskUserQuestion' askuser-protocol.md` → 50 (§8 OK)

### F.2 Gaps (plan-phase에서 명시적으로 관측하지 않은 것)

- 각 섹션의 **정확한** 감소 후 라인 수 — run-phase가 편집을 수행한 후에만 측정 가능. 본 SPEC의 ≤ 320L ceiling은 사전 산술에 근거하며, run-phase가 1-2라인 오차를 허용한다.
- 새 `context-search.md` 룰 파일의 정확한 `paths:` 값 — run-phase가 `~/.claude/projects/` 경로 패턴을 확정해야 한다. 본 SPEC은 path-scoped 요구사항(REQ-CMD2-004)만 명시한다.
- 200L SHOULD 목표 달성 가능성 — §1-§4 canonical을 건드리지 않는 한 불가능으로 예측되나, run-phase의 압축 품질에 따라 ≤ 290L까지는 도달 가능할 수 있다.

### F.3 Residual risk

- §16 새 룰 파일 생성이 template-neutrality CI guard를 통과하지 못할 가능성 — §16 콘텐츠 자체는 SPEC ID를 포함하지 않으므로 낮은 위험. 그러나 run-phase가 예시를 추가할 때 internal reference가 섞일 수 있다.
- 1st-round가 KEEP하기로 한 섹션(§11 recovery/resumable, §14 operational bullets, §15 CG-Mode ASCII)을 pointer-화하려는 시도가 1st-round D1 판결("distinctive content UNIQUE, 0 SSOT hits")과 충돌할 가능성 — 본 SPEC의 AC-CMD2-009는 이 충돌을 run-phase에서 재검증한다.
- 공식 200L 권장 미달성(예상 ~300L)이 사용자 기대와 불일치 — 본 SPEC은 명시적으로 200L를 SHOULD(비 MUST)로 설정하고, 3rd-round를 out of scope로 선언하여 위험을 문서화한다.

---

## G. SPEC ID Pre-Write Self-Check (recorded per protocol)

Decomposition trace (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`):

```
decomposition: SPEC ✓ | CLAUDEMD ✓ | DIET ✓ | V2 ✓ | 001 ✓ → PASS
```

Segment-by-segment verification:
- `SPEC` — literal first segment ✓
- `CLAUDEMD` — `[A-Z][A-Z0-9]*` (C-L-A-U-D-E-M-D, all uppercase letters, length ≥ 1) ✓
- `DIET` — `[A-Z][A-Z0-9]*` (D-I-E-T, all uppercase letters, length ≥ 1) ✓
- `V2` — `[A-Z][A-Z0-9]*` (V uppercase + 2 ∈ [A-Z0-9], length ≥ 1) ✓
- `001` — `\d{3}` digit-only end anchor (no alpha suffix) ✓

AC sub-ID convention note: acceptance.md uses `AC-CMD2-001a` / `AC-CMD2-001b` style sub-IDs where needed — these are acceptance-criteria sub-IDs (allowed trailing alpha), NOT SPEC IDs.

---

## H. Cross-References

- **1st-round predecessor**: `.moai/specs/SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001/` (status: completed, 2026-06-22) — 650→409 lines; its iter-2 D1 verdict on §16 ("0 SSOT hits → KEEP") is the direct precedent this 2nd-round addresses by creating the missing SSOT.
- **Audit origin**: memory `project_official_docs_audit_2026_07` (session 2a40de12) — 4-track official Claude Code docs audit 2026-07.
- **Extraction SSOTs (existing, verified 2026-07-08)**:
  - `.claude/rules/moai/workflow/spec-workflow.md` (33KB) — §5 Agent Chain + §15 Agent Teams target
  - `.claude/rules/moai/workflow/dynamic-workflows.md` (18KB) — §15 Dynamic Workflows prose target
  - `.claude/rules/moai/core/agent-common-protocol.md` (29KB) — §11 Error Recovery + §14 Parallel Execution target
  - `.claude/rules/moai/core/askuser-protocol.md` (25KB) — §8 User Interaction target
- **Extraction SSOT (to be created by this SPEC)**: `.claude/rules/moai/workflow/context-search.md` — §16 Context Search Protocol target
- **Template-First doctrine**: CLAUDE.local.md §2 [HARD] Template-First Rule
- **Template neutrality doctrine**: CLAUDE.local.md §25 + `.moai/docs/template-internal-isolation-doctrine.md`
- **Coding-standards size limit**: `.claude/rules/moai/development/coding-standards.md` § File Size Limits (CLAUDE.md 40K char heuristic + official 200-line target)
- **Verification-claim integrity**: `.claude/rules/moai/core/verification-claim-integrity.md` (§1.1 surface 3 — defect-claim hazard)
- **Related but distinct SPECs (non-overlap confirmed)**:
  - `SPEC-CCSYNC-CLAUDEMD-001` (completed) — Claude Code instruction-layer doc sync, different scope
  - `SPEC-STEERING-ALIGN-LOCAL-DIET-001` — CLAUDE.local.md diet, different file
  - `SPEC-RULE-DIET-002` — rule-file diet, different target
  - `SPEC-V3R5-CLAUDE-REFRESH-001` — earlier refresh, `status: completed` (D6 fix iter-2: 실측 confirmed `status: completed` NOT superseded — "Architecture Truth Reconciliation + Bundle A Settings Fix", 2026-05-18~19, different scope from this diet SPEC)
