# SPEC Review Report: SPEC-AGENCY-ABSORB-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: 0.62

> Reasoning context ignored per M1 Context Isolation. Audit performed exclusively against `spec.md`, `acceptance.md`, and `plan.md` in `.moai/specs/SPEC-AGENCY-ABSORB-001/`. User's prompt-level interpretation (including the claim of "69 REQ-NNN") was disregarded — actual REQ count verified from spec.md is **58**.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**
  REQ numbers use a category-prefixed scheme (`REQ-{CATEGORY}-{NNN}`) and are sequential within each category with consistent 3-digit zero-padding, no gaps, no duplicates.
  Verified: REQ-ROUTE 001–008 (spec.md:L92–118), REQ-SKILL 001–014 (spec.md:L124–183), REQ-MIGRATE 001–012 (spec.md:L189–235), REQ-DIR 001–003 (spec.md:L241–260), REQ-DETECT 001–003 (spec.md:L266–273), REQ-DEPRECATE 001–004 (spec.md:L279–295), REQ-CONST 001–004 (spec.md:L301–311), REQ-BRIEF 001–003 (spec.md:L317–324), REQ-FALLBACK 001–003 (spec.md:L330–337), REQ-REMOVE 001–004 (spec.md:L343–353). Total = 58 REQs.
  Note: 58 ≠ the "69 REQ-NNN" figure quoted in the delegation prompt. The prompt figure is incorrect per spec.md.

- **[FAIL] MP-2 EARS format compliance**
  Two requirements violate EARS pattern grammar:
  - **REQ-FALLBACK-003** (spec.md:L337): Mixes mandatory modal with permissive modal. Text: "the `moai-domain-brand-design` 스킬 **shall** 공개 Figma 파일 URL에서 디자인 토큰을 추출하는 보조 모드를 제공할 **수 있다**(Phase 2 로드맵)." Korean `수 있다` = "may / be able to", which contradicts the mandatory `shall`. This is the exact "mixed informal/formal within a single criterion" failure mode flagged in M3.
  - **REQ-REMOVE-002** (spec.md:L347): Missing mandatory modal entirely. Text: "The `copywriter.md`, `learner.md` 에이전트는 제거 대신 신규 스킬로 내용 **흡수된다**... 파일은 흡수 완료 후 **삭제된다**." Declarative passive voice with no `shall`; does not match any of the five EARS patterns.
  - Secondary issue (not individually disqualifying but compounding): REQ-SKILL-008 (spec.md:L161) and REQ-SKILL-012 (spec.md:L177) each concatenate a mandatory EARS clause with a non-normative "roadmap" or "optional" tail statement inside the same numbered requirement, blurring testability.

- **[FAIL] MP-3 YAML frontmatter validity**
  Two required fields are missing or mis-named in the spec.md frontmatter (spec.md:L1–10). Same defects appear in acceptance.md:L1–9 and plan.md:L1–9.
  - `created_at` — **missing**. The spec uses `created` (spec.md:L5) instead of the required field name `created_at`.
  - `labels` — **missing entirely**. No `labels` field of any type appears in frontmatter.
  - Present: `id`, `version`, `status`, `priority` (four of six required fields).
  Per M5/MP-3 rubric, any single missing required field = FAIL. Two missing fields confirm the FAIL.

- **[PASS] MP-4 Section 22 language neutrality**
  The SPEC explicitly constrains template-bound content to 16-language neutrality (spec.md:L374: "템플릿(`internal/template/templates/`)은 16개 언어 중립성 유지"). No language-specific LSP, formatter, or runtime is elevated to "primary" status within requirements. Mentions of "Go" refer to the implementation runtime of moai-adk itself (spec.md:L68), which is explicitly separated from user-project language scope. No bias found.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension      | Score | Rubric Band | Evidence |
|----------------|-------|-------------|----------|
| Clarity        | 0.65  | ~0.75 band, reduced by one concrete contradiction | REQ-ROUTE-003 (spec.md:L101) is Ubiquitous ("**shall** 첫 번째 옵션을 'Claude Design 위임(권장)'으로 표시") yet REQ-ROUTE-006 (spec.md:L112) demands "경로 B를 기본 선택으로 **재지정한다**" for Pro-or-below users. The spec provides no conditional override clause reconciling the two. Implementers may interpret "first option" differently. AC-ROUTE-004 (acceptance.md:L152–158) attempts a resolution but is not reflected back into the REQ text. |
| Completeness   | 0.55  | between 0.50 and 0.75, pulled down by frontmatter gaps and missing risk REQs | Document sections all present (HISTORY, Goal, Scope, Requirements, Acceptance Criteria, Exclusions with 6 specific entries at spec.md:L57–62, Constraints, Risks, Dependencies, Traceability). However: (a) YAML frontmatter missing two required fields (see MP-3); (b) Claude Design handoff **bundle version mismatch** is listed only as a Risk (spec.md:L387) with no normative REQ, despite being flagged in the delegation audit points; (c) **interrupt/SIGTERM** handling for migration is not addressed anywhere; (d) Windows-specific permission-bit semantics for REQ-MIGRATE-012 (spec.md:L235) are acknowledged in plan.md:L233 as a risk but lack a corresponding REQ. |
| Testability    | 0.60  | between 0.50 and 0.75 | Most ACs are binary-testable with measurable thresholds (e.g., AC-MIGRATE-001 "원본과 바이트 단위로 동일", acceptance.md:L36; AC-SKILL-003 WCAG 4.5:1, acceptance.md:L201). But: REQ-FALLBACK-003 is **explicitly excluded from DoD** (acceptance.md:L468 "Phase 2 로드맵 (DoD 제외)") — an untested requirement that survives in the SPEC. Five REQs are tagged "(암시)" implicit-only coverage in the traceability matrix (REQ-BRIEF-001, REQ-BRIEF-002, REQ-DIR-003, REQ-DETECT-003, REQ-SKILL-004 at acceptance.md:L426, L451, L454, L463, L464), meaning no AC directly evaluates them. Weasel-word scan is otherwise clean within normative sections. |
| Traceability   | 0.60  | between 0.50 and 0.75 | Matrix present (acceptance.md:L411–472) and structurally covers all 58 REQs. But: (a) REQ-FALLBACK-003 mapping is "Phase 2 로드맵 (DoD 제외)" — effectively uncovered; (b) REQ-DEPRECATE-003 mapping is "Release 계획 (DoD §5)" — deferred, not an AC; (c) five "(암시)" entries reduce the effective AC→REQ direct-link count. Also: AC-CONST-002 (acceptance.md:L313) tests "Section 5 Safety Architecture" while the owning REQ-CONST-003 (spec.md:L308) references "Section 5 Layer 5 Human Oversight" — section-number mismatch between test and requirement. |

---

## Defects Found

**D1. spec.md:L83** — Time prediction "(약 6주)" inside Assumption bullet: "최소 1 마이너 버전 주기(약 6주)". This violates the MoAI HARD rule "Never use time predictions in plans or reports" (see `.claude/rules/moai/core/agent-common-protocol.md` Time Estimation section). — **Severity: critical**

**D2. spec.md:L1–10, acceptance.md:L1–9, plan.md:L1–9** — YAML frontmatter missing required field `labels` and uses the name `created` where `created_at` is required. Two required fields failing = automatic MP-3 FAIL. — **Severity: critical**

**D3. spec.md:L337 (REQ-FALLBACK-003)** — EARS grammar violation: "the ... 스킬 **shall** ... 제공할 **수 있다**" combines mandatory `shall` with permissive Korean modal `수 있다` (may / be able to). Fails MP-2. — **Severity: critical**

**D4. spec.md:L347 (REQ-REMOVE-002)** — EARS grammar violation: declarative form without `shall`. "...에이전트는 ... **흡수된다**... 파일은 ... **삭제된다**." Does not match any of the five EARS patterns. Fails MP-2. — **Severity: critical**

**D5. spec.md:L101 vs L112 (REQ-ROUTE-003 vs REQ-ROUTE-006)** — Internal contradiction. REQ-ROUTE-003 unconditionally requires "첫 번째 옵션을 'Claude Design 위임(권장)'으로 표시"; REQ-ROUTE-006 requires "경로 B를 기본 선택으로 재지정" for Pro-or-below users. No override clause or conditional scoping ties the two together in the normative text. Implementer must reverse-engineer the intended precedence from AC-ROUTE-004. — **Severity: major**

**D6. spec.md §8 (Risks table, L387)** — Bundle **version mismatch** handling for Claude Design handoff bundles is flagged as a risk but has no corresponding REQ. The delegation audit explicitly asks whether bundle format parsing failures are defined; they are not in normative form. — **Severity: major**

**D7. spec.md §5.3 (REQ-MIGRATE-001 through REQ-MIGRATE-012)** — **Interrupt / SIGTERM / process-kill during migration** is not specified anywhere in requirements or acceptance. REQ-MIGRATE-006 covers *internal step failure* rollback but not external termination. Without explicit requirement, a Ctrl+C at Phase 3 could leave both `.agency/` and `.moai/` in an inconsistent state with no documented recovery. — **Severity: major**

**D8. acceptance.md:L468 (REQ-FALLBACK-003 coverage)** — Requirement survives in SPEC but is explicitly excluded from DoD: "Phase 2 로드맵 (DoD 제외)". Every REQ that is not testable *now* should either be removed from the SPEC or marked with a future-phase marker in its own entry. As written, REQ-FALLBACK-003 is a normative requirement that the DoD declines to enforce. — **Severity: major**

**D9. spec.md:L235 (REQ-MIGRATE-012)** — "모든 파일 복사 시 권한(permission mode)을 보존한다" is not platform-qualified. plan.md:L233 lists Windows permission semantics as a risk, but no REQ captures the platform-specific acceptance criterion. AC-MIGRATE-001 (acceptance.md:L40) asserts "권한 비트가 원본과 일치" without qualifying Windows NTFS/ACL behavior. — **Severity: major**

**D10. spec.md:L161 (REQ-SKILL-008)** — Requirement body mixes a normative "shall ... 1차 지원한다" clause with a non-normative roadmap claim "DOCX/PPTX/PDF/Canva 링크는 로드맵(Phase 2)으로 표시한다." The second sentence is neither a Ubiquitous requirement nor testable in the current release. Roadmap items belong in §2.1/§2.2 Scope or in a separate "Future Work" section, not inside a numbered REQ. — **Severity: minor**

**D11. spec.md:L177 (REQ-SKILL-012)** — Same pattern as D10: State-Driven EARS clause ("**While** harness level이 `thorough`... **shall** Sprint Contract 협상을 필수로 요구한다") concatenated with an additional Ubiquitous optional clause ("`standard` harness level에서는 선택적이다") in the same REQ. Should be split into two numbered requirements. — **Severity: minor**

**D12. acceptance.md:L313 vs spec.md:L308** — Section-number drift. REQ-CONST-003 names "Section 5 Layer 5 Human Oversight" as the escalation target; AC-CONST-002 tests modifications to "Section 5 Safety Architecture". Both happen to live in Section 5 of the constitution, but the test and the REQ point at different sub-items. — **Severity: minor**

**D13. spec.md:L199–205 (REQ-MIGRATE-004 Step 6)** — Scope confusion. Step 6 of the migration command is a user-`.claude/rules/` relocation. Per CLAUDE.local.md §2 Local-Only Files and Template-First Rule, user `.claude/` mutation from a migration command conflicts with the principle that `.claude/` is runtime-managed. The requirement is ambiguous about whether the source file at `.claude/rules/agency/constitution.md` (user copy) will exist in all target projects, and whether M1 (template deployment) already provides the destination file. — **Severity: minor**

**D14. acceptance.md:L426, L451, L454, L463, L464** — Five requirements (REQ-SKILL-004, REQ-DIR-003, REQ-DETECT-003, REQ-BRIEF-001, REQ-BRIEF-002) are marked "(암시)" in the traceability matrix — indirect/implicit coverage only. A rigorous audit requires direct AC→REQ links. Implicit coverage is insufficient for downstream TDD test generation. — **Severity: minor**

---

## Chain-of-Verification Pass

Second-look findings: Re-read every REQ text end-to-end to verify no EARS violation was missed beyond REQ-FALLBACK-003 and REQ-REMOVE-002. Confirmed no further missing-`shall` cases. Re-verified REQ numbering sequence by counting within each category block (ROUTE 8, SKILL 14, MIGRATE 12, DIR 3, DETECT 3, DEPRECATE 4, CONST 4, BRIEF 3, FALLBACK 3, REMOVE 4) = 58; no gaps, no duplicates. Re-checked Exclusions (spec.md:L57–62): six specific entries, not vague.

New defects surfaced in second pass:
- **D13** (scope confusion in migration Step 6) was missed on first pass — added above.
- **D12** (constitution section-number drift between REQ and AC) was missed on first pass — added above.

Re-confirmed traceability matrix by counting "(암시)" tags: 5 implicit mappings + 1 explicit "(DoD 제외)" exclusion + 1 "Release 계획 (DoD §5)" deferral = 7 weak mappings out of 58 REQs (12% weak coverage).

---

## Regression Check (Iteration 2+ only)

N/A — this is iteration 1.

---

## Recommendation

This SPEC **FAILS** the audit and must be revised before proceeding to the Run phase. Three must-pass criteria are affected (MP-2 EARS, MP-3 YAML) and the overall quality profile is pulled below the 0.75 passing threshold by ambiguity and coverage gaps.

Actionable fix list for manager-spec, ordered by severity:

1. **spec.md:L83** — Remove the "(약 6주)" time prediction. Replace with a version-ordering statement such as "최소 1 마이너 버전 주기 동안 유지된다" (already present — just delete the parenthetical).

2. **spec.md:L1–10, acceptance.md:L1–9, plan.md:L1–9** — Add `labels: [...]` (YAML array or CSV string) to all three frontmatter blocks. Rename `created` → `created_at` (or add `created_at` alongside and deprecate `created`). Recommend: `labels: [agency, migration, design, absorption]`.

3. **spec.md:L337 (REQ-FALLBACK-003)** — Either remove the "수 있다" tail and keep `shall` as the operative modal, or demote the entire REQ to a roadmap entry in §2.2 Exclusions ("Phase 2 로드맵"). If kept as a REQ, rewrite in pure Optional EARS: "**Where** 사용자가 Figma 계정을 연결한 경우, the `moai-domain-brand-design` 스킬 **shall** 공개 Figma 파일 URL에서 디자인 토큰을 추출하는 보조 모드를 제공한다."

4. **spec.md:L347 (REQ-REMOVE-002)** — Rewrite in Ubiquitous EARS: "The 흡수 릴리스 **shall** `copywriter.md` 및 `learner.md` 에이전트의 내용을 신규 스킬(REQ-SKILL-001 ~ REQ-SKILL-003 참조)로 흡수한 뒤 두 파일을 삭제한다."

5. **spec.md:L101 and L112** — Resolve the REQ-ROUTE-003 vs REQ-ROUTE-006 contradiction. Recommended: add a qualifier to REQ-ROUTE-003 such as "기본 구독 플랜 가정 하에" and make REQ-ROUTE-006 the explicit override: "**When** ... Pro 이하 감지, **the** 시스템 **shall** REQ-ROUTE-003의 순서를 뒤집어 경로 B를 첫 번째로 표시한다."

6. **spec.md §5.7 or §5.2** — Add a new REQ for Claude Design handoff **bundle version mismatch**: e.g., "REQ-SKILL-015 (Unwanted): **If** 감지된 bundle 포맷 버전이 지원 화이트리스트에 없으면, **then** the 스킬 **shall** `DESIGN_IMPORT_UNSUPPORTED_VERSION` 오류를 반환하고 최신 지원 버전 목록을 stderr에 출력한다." Add corresponding AC.

7. **spec.md §5.3** — Add a new REQ for migration interrupt handling: e.g., "REQ-MIGRATE-013 (Unwanted): **If** 마이그레이션 도중 SIGINT/SIGTERM이 수신되면, **then** the 커맨드 **shall** 현재 Phase 완료 후 트랜잭션 로그를 flush하고 롤백을 시도하며, 재실행 시 `~/.moai/.migrate-tx-<timestamp>.json`에서 복구 가능하다." Add corresponding AC.

8. **spec.md:L235 (REQ-MIGRATE-012)** — Split into platform-qualified sub-requirements: one for POSIX permission bits (macOS/Linux), one for Windows ACL-equivalent or documented no-op. Reference CLAUDE.local.md lessons.md #7.

9. **acceptance.md:L468 (REQ-FALLBACK-003)** — Either add a concrete AC now or move the REQ itself to §2.2 Exclusions / a "Phase 2 Roadmap" section. Do not leave a normative requirement outside DoD scope.

10. **spec.md:L161 and L177** — Split REQ-SKILL-008 and REQ-SKILL-012 into two numbered requirements each. The roadmap / secondary-state tails become their own numbered REQs (or move to Exclusions for REQ-SKILL-008 roadmap tail).

11. **acceptance.md:L313** — Align section reference to match REQ-CONST-003 exactly (either "Section 5 Layer 5 Human Oversight" in both, or update REQ text).

12. **acceptance.md:L426, L451, L454, L463, L464** — Convert the five "(암시)" traceability entries into direct AC references by writing explicit AC scenarios. No "implicit" coverage should remain in a passing SPEC.

13. **spec.md:L199–205 (REQ-MIGRATE-004 Step 6)** — Clarify scope. Either (a) state that Step 6 only runs if the user-copy of `.claude/rules/agency/constitution.md` exists (otherwise relies on M1 template delivery), or (b) remove Step 6 from the migration command entirely and document the relocation as a template-install-time concern instead.

Estimated iteration cap: **3 iterations**. If iteration 3 still fails on any MP criterion, escalate to user intervention per plan-auditor retry contract.

---

**Files reviewed (absolute paths):**
- `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-AGENCY-ABSORB-001/spec.md`
- `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-AGENCY-ABSORB-001/acceptance.md`
- `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-AGENCY-ABSORB-001/plan.md`

**Report written to:** `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-AGENCY-ABSORB-001/audit-report.md`
