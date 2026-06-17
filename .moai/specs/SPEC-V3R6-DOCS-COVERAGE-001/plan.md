# plan.md — SPEC-V3R6-DOCS-COVERAGE-001

> Implementation plan for docs-site skill-count reconciliation. Tier L (4-locale × multi-page-family + ja structural rewrite). 5-artifact plan-phase set (design.md omitted — no architectural decisions).

---

## §A. Context

본 SPEC은 Sprint 14 Docs-v3 코호트 4/5로, DOCSITE-001에서 연기된 skill-count 축을 소유한다. 핵심 작업:

1. **en/ko/zh**: "31" → "32" count correction + Domain 카테고리에 `moai-domain-humanize` 추가 (8→9) + specialized 합계 30→31. 5개 page-family.
2. **ja**: `advanced/skill-guide.md` 전체 구조 재작성 (9 fictional categories → 6 canonical categories, 46+ nonexistent skill names 제거). 다른 page-families는 en/ko/zh와 동일하게 count correction.

Run-phase는 manager-develop가 DDD cycle_type으로 실행 (문서 정합이므로 사실 기반 correction, characterization-test 불필요 — 대신 grep-based AC verification이 primary 검증).

---

## §B. Known Issues (사전 식별된 이슈)

| Issue | Impact | Mitigation |
|-------|--------|------------|
| ja `skill-guide.md` 전체 재작성 필요 | run-phase LOC 증가 | en/ko/zh의 6-category 구조를 template으로 사용; M3에서 독립 milestone으로 분리 |
| `update.md`가 en/zh에만 존재 | 4-locale parity 예외 케이스 | research.md Coverage Map이 명시; AC-001에서 per-locale grep이 자연스럽게 처리 |
| ja `what-is-moai-adk.md` 비대칭 (L267만 "31") | parity 달성 시 ja에 count claim 추가 필요 | AC-007 parity 검증이 구조적 일치를 요구; minimal-edit이 아닌 parity 우선 |
| ASCII tree code block 내 count (L652/661) | code comment 내 factual claim | AC-001 digit-boundary grep이 code block도 포함; correction 대상 명시 |

---

## §C. Pre-flight (run-phase 진입 전 검증)

run-phase 진입 전 orchestrator가 확인할 항목:

1. `find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l` → 32 (canonical count 재확인)
2. `find .claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-domain-*' -exec basename {} \; | sort` → 9 entries 포함 humanize
3. 기준 baseline: HEAD a7c1b4d48 (DOCSITE-001 close) — 본 SPEC 진입 시점
4. `moai spec lint .moai/specs/SPEC-V3R6-DOCS-COVERAGE-001/spec.md` → 0 findings (이미 plan-phase에서 확인, run-phase 재확인)

---

## §D. Constraints (재확인)

spec.md §D의 5 제약사항이 run-phase에 그대로 적용됨. 특히:

- **제약 2 (4-locale parity)**: 모든 correction은 단일 commit boundary 내에서 4 로케일 동시 적용. ja 구조 재작성도 동일 commit에 포함.
- **제약 4 (primary-source traceability)**: count "32"는 반드시 `find` 출력에 근거. run-phase에서 count 변경 시 `find` 재실행으로 재확인.

---

## §E. Self-Verification (plan-phase 산출물 self-check)

- [x] SPEC ID regex decomposition: `SPEC ✓ | V3R6 ✓ | DOCS ✓ | COVERAGE ✓ | 001 ✓ → PASS`
- [x] Frontmatter 12 canonical fields 모두 존재 (id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags)
- [x] `era: V3R6` 명시 (H-2 transient misclassification 방지)
- [x] `depends_on: [SPEC-V3R6-DOCS-DOCSITE-001, SPEC-V3R6-DOCS-CODEMAPS-V3-001]` 설정
- [x] `created`/`updated` 사용 (snake_case alias 아님)
- [x] `tags` comma-separated string (YAML array 아님)
- [x] GEARS format REQs (REQ-001~008, Ubiquitous/Event-detected/State-driven/Capability gate/Unwanted 패턴 사용)
- [x] §E Exclusions 최소 1개 이상 (5개 exclusions 명시)
- [x] AC 10개 (9 MUST + 1 SHOULD), per-locale digit-boundary-anchored grep
- [x] research.md primary-source verbatim 출력 포함
- [x] DOCSITE-001 scope 비중첩 검증 (§E.1 명시)

---

## §F. Milestones (priority-based, no time estimates)

### M1 — en locale count correction (Priority High)

en locale 5 pages의 "31" → "32" correction. Page-families:

- `advanced/builder-agents.md` L15
- `advanced/skill-guide.md` L65, L127, L167 + Domain 섹션에 `moai-domain-humanize` 추가 + sub-count 8→9
- `getting-started/introduction.md` L133, L156, L163
- `getting-started/update.md` L396 (statusline string)
- `core-concepts/what-is-moai-adk.md` L7, L48, L267, L652

**Verification**: AC-001 (en: 0 residual "31"+skill), AC-002 (en: ≥1 "32"+skill), AC-003 (en: humanize ≥1), AC-004 (en: Domain=9).

### M2 — ko locale count correction (Priority High)

ko locale 4 pages. en과 동일한 correction pattern, locale-native idiom "32개 스킬" 사용.

- `advanced/builder-agents.md` L16
- `advanced/skill-guide.md` L62, L194 + humanize 추가
- `getting-started/introduction.md` L133, L156, L163
- `core-concepts/what-is-moai-adk.md` L7, L48, L267, L652

**Verification**: AC-001~AC-004 ko locale.

### M3 — ja locale structural rewrite (Priority High — most complex)

ja locale은 단순 count patch가 아닌 구조 재작성:

- `advanced/skill-guide.md` **전체 카테고리 섹션 재작성**: 9 fictional categories (Foundation/Workflow/Domain/Language/Platform/Library/Tool/Framework/Design Tools) → 6 canonical categories (Foundation/Workflow/Domain/Reference/Meta-Harness/Design). 46+ nonexistent skill names 제거.
- `advanced/builder-agents.md` L17 count correction
- `getting-started/introduction.md` L133, L156, L163 count correction
- `core-concepts/what-is-moai-adk.md` L267 count correction + parity를 위해 L7/L48/L661에 count claim 추가 (현재 비대칭)

**Verification**: AC-006 (ja structural rewrite), AC-001~AC-004 ja locale, AC-007 (ja 6 categories).

### M4 — zh locale count correction (Priority High)

zh locale 5 pages. en과 동일 pattern, locale-native idiom "32个技能"/"32项技能" 사용.

- `advanced/builder-agents.md` L15
- `advanced/skill-guide.md` L61 (sub-count breakdown의 Domain(8)→Domain(9) 수정 포함), L160
- `getting-started/introduction.md` L133, L156, L163
- `getting-started/update.md` L396
- `core-concepts/what-is-moai-adk.md` L7, L48, L267, L661

**Verification**: AC-001~AC-004 zh locale.

### M5 — 4-locale parity verification + primary-source cross-check (Priority Medium)

모든 milestone 완료 후 통합 검증:

- AC-005 (sub-count sum = 31 specialized) 전 locale 확인
- AC-007 (4-locale parity: 동일 count 32 + 동일 6-category 구조) parity diff
- AC-009 (primary-source `find` 출력 재현) 독립 실행
- AC-010 (`moai spec lint` 0 findings 재확인)

**Verification**: AC-005, AC-007, AC-009, AC-010.

### M6 — Sync-phase handoff (Priority Low)

sync-phase 진입 준비:

- 모든 AC PASS 확인
- diff inspection으로 DOCSITE-001 6축 비중첩 재확인
- CHANGELOG entry 초안 (skill-count reconciliation)
- docs-truth.md §6 "Skill Count (32)" axis 추가 후속 (별도 commit, 본 SPEC scope 밖 — §D.4 forward-looking)

---

## §G. Anti-Patterns (회근해야 할 패턴)

1. **AP-COV-001 — 단일 로케일 correction 후 commit**: 4-locale parity 위반. 모든 correction은 단일 commit boundary 내에서 4 로케일 동시 적용. M1~M4는 논리적 분류이지 commit 분리가 아님 (단일 run-phase commit 권장).
2. **AP-COV-002 — glob grep로 전체 로케일 통합 검사**: per-locale drift mask. AC-001의 `for loc` loop를 우회해서는 안 됨.
3. **AP-COV-003 — ja locale에 en/ko/zh patch 단순 적용**: ja는 구조적 재작성 필요. count patch만 적용하면 fictional taxonomy가 잔존하여 AC-006 FAIL.
4. **AP-COV-004 — count를 "32"로 바꾸되 humanize 추가 누락**: AC-003 FAIL. count correction과 humanize 추가는 동일 milestone 내에서 수행.
5. **AP-COV-005 — user-owned harness skills를 count에 포함**: 34가 아닌 32. template-shipped만 count 대상. §E.4 Exclusions 방어.

---

## §H. Cross-References

- spec.md §B REQs — 구현의 요구사항 근거
- acceptance.md §D ACs — milestone 검증 기준
- research.md §1~§3 — primary-source evidence + coverage map
- `.moai/specs/SPEC-V3R6-DOCS-DOCSITE-001/` — 선행 SPEC (6축 정합 완료)
- `.moai/project/codemaps/docs-truth.md` — docs-truth checklist (§6 axis 추가 후속)
