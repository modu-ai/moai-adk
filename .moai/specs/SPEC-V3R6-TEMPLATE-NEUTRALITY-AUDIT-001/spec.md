---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — dev incident/path/refs sanitization for 16-language template distribution"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-30
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, neutrality, distribution, ci-guard"
tier: L
related_specs: [SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Template Neutrality Audit

## §1 Goal

`internal/template/templates/` 하위 138개 unique file에 잔존하는 **개발자-local 또는 dev-incident 흔적**(macOS-biased absolute paths / V3R[0-9] dev version refs / 2026-05-XX dates / feedback_·memory refs / CLAUDE.local.md refs / PR #N / commit hash)을 **범용적(template-grade)** 컨텐츠로 정정한다. 16-language 사용자 프로젝트에 배포되는 모든 template 파일이 다음 4개 neutrality 기준을 만족해야 한다:

1. **No absolute path bias** — `/Users/...`, `/home/...` 등 OS-specific 경로 placeholder 금지 → `$HOME/...` 또는 `~/...` 통일
2. **No dev-incident traceability leakage** — V3R[0-9] dev version sigils, specific commit hashes, specific PR #N references → allow-list 정의 후 잔존은 doctrine 인용 경우만 허용
3. **No maintainer-personal information** — author 필드 `GOOS Kim` → `Author Name` (이미 fc47f31a7 Critical fix 완료, 본 SPEC은 재발 방지 CI guard 추가)
4. **No local-only file refs** — `CLAUDE.local.md` 참조는 template 배포 부적합 (사용자 프로젝트에는 존재하지 않음)

**False positive** identification: Go `GOOS=windows`, `GOOS=linux`, `GOOS=darwin` cross-compile environment variable은 보존(audit script exclude).

## §2 Non-Goals

본 SPEC은 다음을 **수행하지 않는다**:

### §2.1 Out of Scope — Template neutrality audit boundary

- **C3 (generic ISO date `2026-0[5-9]`) enforcement** — DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (`internal_content_leak_test.go` strict-tier `S1-internal-date`). 본 SPEC은 date class를 scan하지 않는다 (dual-allow-list drift 회피).
- **C7 (commit hash) enforcement** — DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (`internal_content_leak_test.go` strict-tier `S2-short-sha-sentence-final`). 본 SPEC은 commit-hash class를 scan하지 않는다.
- Sprint 1 in-progress SPEC directories (`SPEC-V3R6-AGENT-MODEL-ROUTING-001/`, `SPEC-V3R6-HOOK-ASYNC-EXPAND-001/`, `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/`, `SPEC-V3R6-PROMPT-CACHE-001/`) 직접 수정 — 각 SPEC 개별 plan/run 흐름에서 처리
- `docs-site/content/{en,ko,ja,zh}/book/` untracked 잔존 디렉토리 정리 — 별도 docs-site cleanup SPEC
- `internal/hook/.moai/` working dir leak 정리 — git-strategy cleanup SPEC
- `.moai/specs/.moai/` symlink/dup 정리 — workspace hygiene SPEC
- `settings.local.json`의 maintainer-specific 값 (e.g., `teammateMode`, `defaultMode`) — runtime-managed, §2 [HARD] settings.local.json Separation
- `EXCL-CCE-001` 패턴의 user-facing string literal (Korean nodejs install URLs, GitLab descriptions, BODP literal `(권장)`) — translation 불필요
- Code 파일 (`internal/*.go`, `pkg/*.go`)의 dev-incident 주석 — 본 SPEC scope는 `internal/template/templates/` 한정 (Go code dev-incident 정리는 SPEC-V3R6-CODE-COMMENTS-EN-001 또는 별도 SPEC)
- Template Go template variable rendering 정정 (`{{.GoBinPath}}` 등 `internal/template/render_*.go`의 Go template logic은 무관)

## §3 Background

### 3.1 발견 경위

2026-05-23 commit `fc47f31a7`에서 Critical 4 violations을 즉시 fix하였음:
- `internal/template/templates/.claude/skills/moai/SKILL.md` 2 lines: `/Users/goos/MoAI/moai-adk-go/...` → `${CLAUDE_SKILL_DIR}/...`
- `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` 2 lines: `author: GOOS Kim` → `author: Author Name`
- 신규 `internal/template/templates/.claude/rules/moai/development/sprint-round-naming.md` SSOT (V3R6 / PR #819 refs 제거된 generic version)

그러나 본 세션 추가 audit 결과 (baseline re-measured 2026-05-30, absolute-path grep at HEAD `ecda4ef04`):
- **C1 macOS-biased path placeholder 4 files** (`worktree-integration.md`, `run/context-loading.md`, `moai-foundation-cc/examples.md`, `moai-workflow-loop/examples.md`) — 모두 `/Users/user/...` 또는 `/Users/john/...` generic placeholder이나 macOS-only path syntax — **KEEP (NEUTRALITY-unique)**
- **C2 V3R[0-9] bare-narrative sigils 7 files** — broad `V3R[0-9]` regex가 341 hits를 매칭하나 그중 299 (88%)는 `SPEC-V3R…` / `CONST-V3R…` / `REQ-V3R…` ID-embedded substring이며 이는 ISOLATION-001의 `C1-spec-id` leak-test class 소유. C2는 **bare-narrative sigil (ID-embedded 아닌 7 files)** 만 scope — **KEEP (NEUTRALITY-unique, narrowed)**. ID-embedded 299 sanitization은 ISOLATION-001 job (forbidden cross-SPEC scope bleed).
- **C3 2026-05-XX dates 39 files** — incident date allow-list 외 generalize — **DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001** (그 SPEC의 `internal_content_leak_test.go` strict-tier `S1-internal-date` `\b202[6-9]-MM-DD\b` 가 본 class를 enforce; 본 SPEC은 중복 enforcement 제거. 자세한 partition은 §3.3 참조)
- **C4 feedback_/memory refs 9 files** — rule/doctrine 인용 정상, 외 generalize — **KEEP (NEUTRALITY-unique)**. `internal_content_leak_test.go`는 memory *path* class (`~/.claude/projects/-Users-` / `.moai/backups/agent-archive-`)만 enforce하며, 본 C4의 `feedback_` / `memory.md` *substring reference* class는 default/strict 어느 모드에서도 enforce하지 않음 (§3.3 partition 표 참조)
- **C5 CLAUDE.local.md refs 3 files** — local-only file ref는 template에 부적합 — **KEEP (NEUTRALITY-unique)** (baseline 10→3, partial prior cleanup 반영)
- **C6 PR #N refs 3 files** — 특정 PR 번호는 template 부적합 — **KEEP (NEUTRALITY-unique)**
- **C7 commit hash refs 2 files** — 특정 hash는 template 부적합 — **DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001** (그 SPEC의 strict-tier `S2-short-sha-sentence-final` `\b[0-9a-f]{7,8}([\s\.,;:!?]|$)` 가 본 class를 enforce)
- **C8 False positive** `GOOS=(linux|windows|darwin)` Go env var 4 hits / 3 files — **KEEP (NEUTRALITY-unique, PRESERVE)**

**Rescope (v0.1.1)**: 본 SPEC은 plan-audit iter-1 (FAIL 0.71) 후 **NEUTRALITY-unique 카테고리 (C1/C2/C4/C5/C6/C8)** 로 rescope되었으며, **C3/C7** 는 이미 shipped된 sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (status `completed`, `internal/template/internal_content_leak_test.go` 16810B) 의 strict-tier pattern class로 defer한다. 이로써 같은 Go package (`internal/template/`) 안에 date/SHA class에 대한 dual-allow-list drift를 제거한다. 자세한 partition은 §3.3 참조.

KEPT class baselines (re-measured 2026-05-30 at HEAD `1162b0de8`, point-in-time; run-phase M1/M3 re-measures before fixing): C1=4, **C2=7 (bare-narrative only; broad `V3R[0-9]`=341 hits of which 299 ID-embedded are ISOLATION-owned)**, C4=9, C5=3, C6=3, C8=3 files. KEPT class scope (overlaps account for dedupe). Tier L migration SPEC scope.

### 3.2 Template-First Rule 부작용

`.claude/rules/moai/development/template-first-rule.md`는 모든 신규 rule이 `internal/template/templates/` 동시 미러를 의무화한다. 그 결과 **dev-specific incident rule** (예: V3R6 round-splitting lessons, sprint-round naming PR #819 history)이 user-facing template에 자동 배포될 위험. 본 SPEC은:
- Audit script (`internal/template/template_neutrality_audit_test.go`)로 재발 차단
- CI workflow (`template-neutrality-check.yaml`)으로 PR-level enforcement
- Template-First Rule guideline 보강 — acceptable content range 명문화

### 3.3 분류 매트릭스 필요성 + ISOLATION partition (rescope v0.1.1)

KEPT case-by-case 카테고리 (C2 V3R refs 73 / C4 memory refs 9)는 일괄 removal하면 안 됨. 각 카테고리에 **case-by-case classify**가 필요:
- **PRESERVE (allow-list)**: rule SSOT 인용, decision rationale, doctrine citation (예: "Per session-handoff.md, the 5 triggers are ..." — V3R5 ID로 등록된 HARD clause는 보존)
- **GENERALIZE**: incident-specific 표현을 패턴 표현으로 변경 (예: "a 2026 stream-stall incident" 또는 "a prior stream-stall incident")
- **REMOVE**: 단순 dev-history 흔적 (제거)

이 매트릭스는 `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md`에 single source로 기록되며, M2~M5에서 manager-develop이 참조한다.

#### ISOLATION-001 deconfliction (C3/C7 defer, C4 keep)

shipped sibling **SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001** (status `completed`, `internal/template/internal_content_leak_test.go` 16810B, same Go package `internal/template/`)가 date/SHA/memory-path class를 enforce하므로, 본 SPEC은 중복(dual-allow-list) 회피를 위해 다음 partition을 적용한다. (leak-test pattern 검증 2026-05-30):

| NEUTRALITY 카테고리 | leak-test enforcement | partition 결정 |
|---|---|---|
| C1 `/Users/` paths | 미enforce | **KEEP (NEUTRALITY)** |
| C2 bare-narrative `V3R[0-9]` (NOT ID-embedded) | 미enforce | **KEEP (NEUTRALITY — narrowed to 7 bare-narrative files)** |
| C2 ID-embedded `SPEC-V3R…` / `REQ-V3R…` (299 of 341 hits) | ISOLATION `C1-spec-id` leak-test class | **DEFER to ISOLATION (not a C2 target)** |
| C3 generic ISO date `2026-0[5-9]` | enforce via strict-tier `S1-internal-date` `\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b` (opt-in `MOAI_TEMPLATE_LEAK_STRICT=1`; §25.1 evolution policy의 future-tightening tier) | **DEFER to ISOLATION** |
| C4 `feedback_` / `memory.md` *substring reference* | **미enforce** (default 및 strict 모두). leak-test C5는 memory *path* class (`~/.claude/projects/-Users-` / `.moai/backups/agent-archive-`)만 enforce하며 이는 본 C4와 disjoint pattern | **KEEP (NEUTRALITY)** |
| C5 `CLAUDE.local.md` refs | 미enforce | **KEEP (NEUTRALITY)** |
| C6 `PR #N` refs | 미enforce | **KEEP (NEUTRALITY)** |
| C7 commit hash | enforce via strict-tier `S2-short-sha-sentence-final` `\b[0-9a-f]{7,8}([\s\.,;:!?]\|$)` (opt-in strict mode) | **DEFER to ISOLATION** |
| C8 `GOOS=` Go env var | 미enforce (false-positive PRESERVE) | **KEEP (NEUTRALITY)** |

**C4 discrepancy note**: plan-audit iter-1 SCOPE 권고는 C3/C4/C7 일괄 defer였으나, leak-test 실측 결과 **C4 (`feedback_`/`memory.md` substring reference) class는 leak test가 enforce하지 않음** (default/strict 모두). leak-test C5 path class와 disjoint pattern이므로 C4를 silently drop하면 enforcement gap이 발생한다. 따라서 C4는 NEUTRALITY에 **유지**한다 (per orchestrator guardrail: "If the leak test does NOT enforce one of C3/C4/C7, keep that category in NEUTRALITY"). C3/C7만 defer.

**dual-allow-list drift 제거**: defer된 C3/C7은 본 SPEC의 audit script (REQ-TNA-009)가 scan하지 않으므로, `internal/template/` package 안에 date/SHA class에 대한 allow-list가 ISOLATION의 leak-test 하나만 존재한다 (drift 원천 제거).

**C2 narrow note (v0.1.2 M3 blocker resolution)**: M3 진입 시 manager-develop이 broad `V3R[0-9]` regex의 ≤18 target과 73-file baseline이 **달성 불가능**임을 구조적 C2 conflict로 발견하고 blocker를 반환했다. 근본 원인: broad `V3R[0-9]` (341 hits) 중 299 (88%)가 `SPEC-V3R…` / `CONST-V3R…` / `REQ-V3R…` ID-embedded substring이며, 이를 sanitize하려면 ISOLATION-001의 `C1-spec-id` leak-test class 영역을 침범해야 한다 (forbidden cross-SPEC scope bleed). 해소 (Option A, user-approved): C2 detection을 **bare-narrative `V3R[0-9]` (ID-prefix token에 직접 선행되지 않는 sigil)** 으로 narrow한다 (REQ-TNA-002 § C2 detection scope 참조). 이로써 C2는 7 bare-narrative files만 owns하고, 299 ID-embedded matches는 ISOLATION-001로 defer되어 disjointness가 보존된다 (C3/C7과 동일 disjoint discipline).

### 3.4 Known Pre-existing Issues (out of scope)

#### Pre-existing package RED at run-phase baseline

[HARD] At this SPEC's run-phase baseline (HEAD `1162b0de8`), the `internal/template` Go package is **already RED** — `go test ./internal/template/...` reports **13 pre-existing failing test functions** that exist independently of this SPEC's work. **This SPEC does NOT fix these failures; they are explicitly out of scope.**

The 13 pre-existing failing test functions (orchestrator-measured 2026-05-30):

1. `TestTemplateNoInternalContentLeak` — the ISOLATION-001 leak test, **RED despite ISOLATION-001 `status: completed`**
2. `TestRuleTemplateMirrorDrift` — template ↔ `.claude/` byte-parity drift
3. `TestLateBranchTemplateMirror` — Late-branch doctrine mirror parity
4. `TestManagerDevelopActiveAgentPresent`
5. `TestManagerDevelopIsActiveAgent`
6. `TestEmbeddedTemplates_AgentDefinitions`
7. `TestAgentFrontmatterAudit`
8. `TestBuilderSkillPathStructure`
9. `TestTemplateAgentsStructure`
10. `TestSettingsTemplateHookEventCount`
11. `TestContractSchemaVerification`
12. `TestBackwardCompatibility`
13. `TestContractAssertionsNaturalLanguage`

**Implication for this SPEC's verification**: the new audit script (REQ-TNA-009 `TestTemplateNeutralityAudit`) is designed to PASS as an **independent test function** — it is verified in isolation via `go test ./internal/template/... -run TestTemplateNeutralityAudit` (AC-TNA-008), NOT via the package-wide `go test ./internal/template/...`. The package-level run remains **RED** until a separate cleanup SPEC addresses the 13 pre-existing failures. Acceptance of this SPEC MUST use the `-run TestTemplateNeutralityAudit` targeted form; a package-wide green is NOT a precondition and MUST NOT block this SPEC's closure.

**Recommended follow-up (successor SPEC)**: a dedicated `internal/template` package-RED cleanup SPEC (provisional `SPEC-V3R6-TEMPLATE-PACKAGE-RED-CLEANUP-001`, Tier M) should triage and fix the 13 failures — notably `TestTemplateNoInternalContentLeak` (which is RED despite ISOLATION-001 being marked `completed`, indicating an unresolved leak or an over-broad strict-tier assertion) and the 3 mirror-drift tests (`TestRuleTemplateMirrorDrift` / `TestLateBranchTemplateMirror` / `TestEmbeddedTemplates_AgentDefinitions`). That successor SPEC owns the package-wide green; this NEUTRALITY SPEC is scoped to its own 6 kept-class fixes + the disjoint audit script only.

#### F3 — template ↔ `.claude/` mirror-drift caveat (4 byte-parity files)

[HARD] Four files touched by C2/C4/C5/C6 are on the `rule_template_mirror_test.go` byte-parity allow-list, meaning their `internal/template/templates/…` mirror MUST stay byte-identical to their `.claude/…` counterpart:

- `manager-develop-prompt-template.md` (C2 bare-narrative target + C8 GOOS preserve)
- `manager-spec.md` (C2 bare-narrative — SPEC-ID decomposition self-check example)
- `spec-workflow.md` (C2/C4 context)
- `manager-git.md` (C6 PR-ref context)

Some of these `.claude/` mirrors are already genericized while the template mirrors lag (pre-existing drift surfaced by `TestRuleTemplateMirrorDrift` RED). When run-phase manager-develop edits any of these 4 files, it MUST keep template ↔ `.claude/` mirror parity: edit **both sides** in the same milestone, OR verify the `.claude/` side already matches the intended generic form before editing only the template side. Failing to maintain parity widens the existing `TestRuleTemplateMirrorDrift` failure rather than resolving it.

## §4 EARS GEARS Requirements

다음 13개 REQ는 GEARS notation (Where / When / While / If-Then) self-dogfooding 원칙을 따른다. EARS-only "The system shall …" 패턴 사용 시에도 GEARS-compatible context 선언과 결합한다.

**Rescope status (v0.1.1)**: REQ-TNA-001/002/004/005/006/008 (kept classes C1/C2/C4/C5/C6/C8) + 인프라 REQ-TNA-009..013 은 본 SPEC에서 active하다. **REQ-TNA-003 (C3 dates)** 및 **REQ-TNA-007 (C7 commit hash)** 는 sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (`internal/template/internal_content_leak_test.go` strict-tier classes)로 **DEFERRED** 되었다. REQ 번호는 contiguity 보존을 위해 재번호하지 않고 원번호 유지 + 명시적 DEFERRED status line으로 표기한다 (§3.3 partition 참조).

### REQ-TNA-001 — macOS-biased path placeholder removal (C1)

**Where** the template tree (`internal/template/templates/**`) contains an absolute path placeholder beginning with `/Users/`, `/home/`, or any OS-specific user home prefix, **the system shall** replace the prefix with `$HOME` or `~` while preserving the remaining path segments, file role context, and surrounding example narrative.

Affected baseline (verified 2026-05-23): 4 files / 8 lines
- `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` — lines 260, 261, 270, 273
- `internal/template/templates/.claude/skills/moai/workflows/run/context-loading.md` — lines 177, 185, 188
- `internal/template/templates/.claude/skills/moai-foundation-cc/references/examples.md` — line 646
- `internal/template/templates/.claude/skills/moai-workflow-loop/references/examples.md` — line 396

### REQ-TNA-002 — V3R[0-9] bare-narrative dev version sigils classification (C2)

**Where** the template tree contains a **bare-narrative** occurrence of `V3R[0-9]` — i.e. a `V3R[0-9]` substring NOT immediately preceded by an ID-prefix token (`SPEC-`, `CONST-`, `REQ-`) nor part of any larger `[A-Za-z0-9-]` identifier — **the system shall** classify each occurrence per the migration matrix into PRESERVE (rule SSOT citation / decision record / named-doctrine citation), GENERALIZE (incident pattern), or REMOVE (dev-history trace). **If** the occurrence is classified as PRESERVE, **then** the file shall be added to the allow-list defined in `migration-matrix.md` §C2 with explicit rationale; otherwise the occurrence shall be transformed per the classified action.

#### C2 detection scope — bare-narrative only (ID-embedded matches are out of scope)

[HARD] The C2 detection target is the **bare-narrative version sigil** ONLY. The broad `V3R[0-9]` regex collides with the SPEC-ID / CONST-registry-ID / REQ-ID domain that is owned by the sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001's `C1-spec-id` leak-test class. C2 owns ONLY the bare-narrative sigil so the two SPECs stay disjoint (the same disjointness discipline already applied to C3/C7 in §3.3).

Ground-truth measurement (orchestrator-measured, absolute-path grep, 2026-05-30 HEAD `1162b0de8`):

| Subset | Count | Owner |
|---|---|---|
| Total `V3R[0-9]` occurrences | 341 | — |
| `SPEC-V3R[0-9]` (ID-embedded) | 165 | ISOLATION-001 `C1-spec-id` (DEFER) |
| `CONST-V3R[0-9]` (registry-ID-embedded) | 130 | zone-registry SSOT (PRESERVE — see §C2 allow-list) |
| `REQ-V3R[0-9]` (ID-embedded) | 4 | ISOLATION-001 leak-test (DEFER) |
| **ID-embedded subtotal** | **299 (88%)** | **NOT a C2 target** |
| **Bare-narrative `V3R[0-9]`** | **7 files** | **C2 target (this REQ)** |

The old `V3R[0-9]` regex + the ≤18 allow-list target were unachievable without sanitizing the 299 ID-embedded matches — which is ISOLATION-001's job (forbidden cross-SPEC scope bleed). The narrowed C2 measures only the 7 bare-narrative files.

#### Two-pass detection approach for the Go audit script (REQ-TNA-009)

Go's `regexp` package (RE2) does NOT support lookbehind, so the audit script MUST implement bare-narrative detection as an explicit two-pass exclusion (a single Go regex cannot express "not preceded by an ID prefix"):

1. **Pass 1 — find all candidates**: match `\bV3R[0-9]` on each line (RE2-compatible word-boundary anchor).
2. **Pass 2 — exclude ID-embedded hits**: for each Pass-1 hit, inspect the immediately preceding non-space token. Drop the hit when the `V3R` is part of a larger identifier — concretely, drop it when the match is reached via any of the prefix forms `SPEC-V3R`, `CONST-V3R`, `REQ-V3R`, OR when the preceding character is a member of `[A-Za-z0-9-]` (i.e. the `V3R` continues an identifier token). Equivalently, the script MAY pre-compile the negative forms `SPEC-V3R[0-9]`, `CONST-V3R[0-9]`, `REQ-V3R[0-9]` and subtract their hit set from the Pass-1 hit set, then additionally drop any residual hit whose preceding rune is `[A-Za-z0-9-]`.

A shell-verifiable equivalent (perl PCRE supports negative lookbehind) is `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'`, used by AC-TNA-002 below. The Go two-pass exclusion MUST produce the same file set as this perl form.

Baseline (orchestrator-measured 2026-05-30 at HEAD `1162b0de8`, bare-narrative grep `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'`): **7 bare-narrative files** (replaces the stale 70/73 broad-regex counts). The 7 files are: `manager-spec.md`, `zone-registry.md`, `manager-develop-prompt-template.md`, `moai-harness-learner/SKILL.md`, `moai/SKILL.md`, `moai/workflows/harness.md`, `moai-meta-harness/SKILL.md`. Point-in-time; run-phase M3 re-measures the bare-narrative set before fixing.

**Post-fix target**: ≤ allow-list count. Of the 7 bare-narrative files, the legitimate PRESERVE entries (zone-registry V3R2/V3R5 namespace decision record + CONST-registry section headers; manager-spec.md SPEC-ID decomposition self-check example; the harness `V3R4 Self-Evolving` authoritative-SPEC decision-record citations in `moai-harness-learner/SKILL.md`, `moai/SKILL.md`, `harness.md`, `moai-meta-harness/SKILL.md`) stay; the remainder (`manager-develop-prompt-template.md` `V3R4 HARNESS retirement` example + `다른 V3R6 SPEC` generic) is GENERALIZE/REMOVE. Expected post-fix bare-narrative count: **6 PRESERVE** (matching the allow-list) after `manager-develop-prompt-template.md` is generalized.

### REQ-TNA-003 — Date refs classification (C3) — **[DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]**

**[DEFERRED]** Date-class enforcement (generic ISO `2026-0[5-9]-[0-9]{2}`) is delegated to the shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001, whose `internal/template/internal_content_leak_test.go` strict-tier `S1-internal-date` class (`\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b`, opt-in via `MOAI_TEMPLATE_LEAK_STRICT=1`, tracked under its §25.1 evolution policy) already owns this class. This SPEC does NOT re-scan generic dates (avoids dual-allow-list drift within the `internal/template/` Go package). The REQ number is preserved for traceability contiguity; no acceptance criterion is emitted for C3 (see acceptance.md AC-TNA-003 deferred marker).

Baseline (re-measured 2026-05-30): 39 files (was 32 at 2026-05-23; +7 drift) — informational only; enforcement owned by ISOLATION strict tier.

### REQ-TNA-004 — feedback_/memory refs classification (C4) — **[KEPT — NEUTRALITY-unique]**

**Where** the template tree contains a substring matching `feedback_` or `memory\.md`, **the system shall** classify each occurrence. **If** the reference cites a rule SSOT or doctrine pattern (e.g., "Lessons Protocol writes to memory `lessons.md`"), **then** PRESERVE; **otherwise** the reference shall be removed or replaced with a generic "auto-memory" phrasing.

**Why kept (not deferred)**: the shipped `internal_content_leak_test.go` does NOT enforce the `feedback_` / `memory.md` *substring reference* class in either default or strict mode. Its C5 class enforces only memory *paths* (`~/.claude/projects/-Users-` / `.moai/backups/agent-archive-`), which is a disjoint pattern from this C4 substring class. Deferring C4 would silently drop enforcement of a class no other test covers — therefore C4 remains a NEUTRALITY-unique requirement (verified 2026-05-30 against `internal_content_leak_test.go` pattern catalog; see §3.3 partition table).

Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`): 9 files (unchanged from 2026-05-23). Point-in-time; run-phase M3 re-measures before fixing.

### REQ-TNA-005 — CLAUDE.local.md refs handling (C5)

**Where** the template tree references `CLAUDE.local.md` directly, **the system shall** remove the reference because `CLAUDE.local.md` is documented as a maintainer-only local file (per `CLAUDE.local.md` self-description). **If** the surrounding context discusses local-only configuration in the abstract, **then** the reference shall be replaced with a generic statement (e.g., "local override file" or "machine-specific configuration").

Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`): **3 files** (was 10 at 2026-05-23; −7 from partial prior cleanup). Point-in-time; run-phase M3 re-measures before fixing.

### REQ-TNA-006 — PR #N refs removal (C6)

**Where** the template tree contains a substring matching `PR #[0-9]+`, **the system shall** remove the specific PR number. **If** the surrounding context discusses a rule precedent that requires historical evidence, **then** the PR ref shall be replaced with a generic phrase (e.g., "a prior round of the same rule" or "the originating PR" without a specific number).

Baseline (verified 2026-05-23): 3 files. Examples include `manager-develop-prompt-template.md`, `sync/delivery.md`, `scripts/ci-mirror/cross-compile.sh`.

### REQ-TNA-007 — Commit hash refs removal (C7) — **[DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]**

**[DEFERRED]** Commit-hash-class enforcement is delegated to the shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001, whose `internal/template/internal_content_leak_test.go` strict-tier `S2-short-sha-sentence-final` class (`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`, opt-in via `MOAI_TEMPLATE_LEAK_STRICT=1`) already owns this class with a more conservative, false-positive-aware detection than the broad `[a-f0-9]{7,40}` regex this REQ originally proposed. This SPEC does NOT re-scan commit hashes (avoids dual-allow-list drift within the `internal/template/` Go package). The REQ number is preserved for traceability contiguity; no acceptance criterion is emitted for C7 (see acceptance.md AC-TNA-007 deferred marker).

Baseline (re-measured 2026-05-30): ~2 files — informational only; enforcement owned by ISOLATION strict tier. The original broad C7 regex was FP-saturated (45 raw hits) and depended on a not-yet-existing test, which is precisely the hard problem ISOLATION already solved conservatively.

### REQ-TNA-008 — False positive exclusion (`GOOS=...` Go env var) (C8)

**Where** the template tree contains a substring matching `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` in the context of Go cross-compilation (e.g., shell scripts, Go build commands), **the system shall** preserve the occurrence unchanged. The audit script (REQ-TNA-009) shall exclude this pattern from violation reports.

Baseline (verified 2026-05-23): 4 hits / 3 files
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md`
- `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md`
- `internal/template/templates/scripts/ci-mirror/cross-compile.sh`

### REQ-TNA-009 — Audit script (`template_neutrality_audit_test.go`) — **[SCOPED to kept classes C1/C2/C4/C5/C6/C8]**

**Where** `internal/template/` is the Go test discovery root, **the system shall** provide a new test file `internal/template/template_neutrality_audit_test.go` containing the function `TestTemplateNeutralityAudit`. **When** invoked via `go test ./internal/template/... -run TestTemplateNeutralityAudit`, the test shall scan `internal/template/templates/**` and emit a single FAIL with a structured violation report when category C1, C5, or C6 (binary patterns) finds a hit outside its allow-list, and a WARN log line (not a FAIL) when C2 (bare-narrative `V3R[0-9]`) or C4 (case-by-case patterns) find a hit outside the migration-matrix allow-list. **If** the C8 false-positive pattern (`GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)`) matches, **then** the hit shall be excluded from all reports.

**[C2 two-pass detection]** Because Go's `regexp` (RE2) lacks lookbehind, the C2 subtest MUST implement the **bare-narrative two-pass exclusion** specified in REQ-TNA-002 § "Two-pass detection approach": Pass 1 matches `\bV3R[0-9]`; Pass 2 drops hits reached via `SPEC-V3R` / `CONST-V3R` / `REQ-V3R` ID prefixes or whose preceding rune is `[A-Za-z0-9-]`. The C2 subtest's resulting file set MUST equal the perl PCRE form `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'` (AC-TNA-002). The C2 subtest MUST NOT match the 299 ID-embedded `SPEC-V3R…` / `CONST-V3R…` / `REQ-V3R…` substrings (those belong to ISOLATION-001's `C1-spec-id` class).

**[run-phase isolation note]** This audit script is run and verified in **isolation** (`-run TestTemplateNeutralityAudit`), NOT via the package-wide `go test ./internal/template/...` — the package is RED at run-phase baseline with 13 pre-existing failures unrelated to this SPEC (see §3.4). The new test function is designed to PASS standalone.

**[SCOPE] C3 / C7 are NOT scanned by this test.** The date class (C3) and commit-hash class (C7) are owned by the shipped sibling `internal/template/internal_content_leak_test.go` (strict-tier `S1-internal-date` / `S2-short-sha-sentence-final`). Re-scanning them here would create two test files in the same Go package (`internal/template/`) with two divergent allow-lists for one pattern class — a guaranteed drift source. This audit script's pattern set is therefore **disjoint** from `internal_content_leak_test.go`.

**[DECISION] New disjoint test file, NOT an extension of the leak test.** REQ-TNA-009 creates a NEW `template_neutrality_audit_test.go` rather than extending `internal_content_leak_test.go`, because the two tests have different severity semantics (NEUTRALITY C2/C4 are WARN-level advisory; the leak test is all-FAIL) and different ownership lifecycles (NEUTRALITY is the active SPEC; ISOLATION is `completed`/frozen). The pattern sets MUST remain disjoint: NEUTRALITY enforces {C1, C2, C4, C5, C6, C8-exclude}; the leak test enforces {C1-spec-id, C2-req-ac, C3-audit, C4-finding/archive-date, C5-memory-path} + strict {S1-date, S2-sha, S3-req-ac}. No pattern is enforced by both files.

### REQ-TNA-010 — CI guard workflow (PR touching template/)

**Where** a GitHub PR modifies any file under `internal/template/templates/**`, **the system shall** trigger the `template-neutrality-check.yaml` workflow which runs `TestTemplateNeutralityAudit` and surfaces failures as a required status check. **If** the workflow detects a C1/C5/C6 violation (kept binary classes), **then** the PR check shall fail; **if** only C2/C4 WARN-level findings are emitted, **then** the check shall pass with annotations. (C3/C7 are out of this workflow's scope — owned by the leak-test gate per the ISOLATION SPEC; see REQ-TNA-009 SCOPE note.)

### REQ-TNA-011 — Migration matrix (`migration-matrix.md`)

**Where** the SPEC directory is `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/`, **the system shall** include a file `migration-matrix.md` containing one section per category (C1–C8) with: (a) the regex / detection rule, (b) the action policy (PRESERVE / GENERALIZE / REMOVE / DEFERRED), (c) the allow-list (file paths + rationale), (d) the baseline count, and (e) the post-fix expected count. **If** a category has zero allow-list entries, **then** the allow-list section shall explicitly state "Empty (no exceptions)". **If** a category is deferred to the ISOLATION SPEC (C3, C7), **then** its section shall carry a `DEFERRED` action policy and reference the owning leak-test class rather than an allow-list.

### REQ-TNA-012 — Template-First Rule guideline reinforcement

**Where** the canonical Template-First Rule is documented (currently in `CLAUDE.local.md` §2 and any equivalent rule file), **the system shall** add an `Acceptable Content Range for Templates` subsection enumerating: (a) what is acceptable in template content (rule SSOT citations, doctrine name references, GENERIC examples), (b) what is rejected (specific PR / commit / date / V3R refs, absolute paths, maintainer-personal names, local-only file refs). **If** a future rule introduces template-bound content, **then** the rule author shall verify against this guideline before committing.

### REQ-TNA-013 — Documentation for future template additions

**Where** a contributor authors a new file destined for `internal/template/templates/`, **the system shall** provide a discoverable checklist (referenced from `CLAUDE.local.md` and from REQ-TNA-012's subsection) that the contributor follows. **If** the checklist verification fails (e.g., the new file references a specific PR #), **then** the contributor shall remediate before opening the PR; the CI guard (REQ-TNA-010) shall act as the safety net catching missed manual verifications.

## §5 Acceptance Criteria reference

See `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/acceptance.md` for the 13 binary AC scenarios (AC-TNA-001 through AC-TNA-013) and the Given-When-Then test scenarios.
