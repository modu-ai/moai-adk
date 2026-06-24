# plan.md — SPEC-V3R6-DOCS-COVERAGE-001

> Implementation plan for docs-site skill-count reconciliation. Tier L (4-locale × multi-page-family + ja structural rewrite). 5-artifact plan-phase set (design.md omitted — no architectural decisions).

---

## §A. Context

본 SPEC은 Sprint 14 Docs-v3 코호트 4/5로, DOCSITE-001에서 연기된 skill-count 축을 소유한다. iter-2 재도출로 4-locale 허구적 스킬명 분포가 정정됨 (총 95개: en:10, ko:37, ja:37, zh:11). 핵심 작업:

1. **en/zh**: "31" → "32" count correction + Domain 카테고리에 `moai-domain-humanize` 추가 (8→9) + specialized 합계 30→31. category header 구조는 6 canonical로 정확. 추가로, 개념 설명 영역(Mermaid·코드 예제·ASCII tree·자동 로드 시나리오·callout)의 허구적 스킬명 본문 정제 (en: 10, zh: 11).
2. **ja 및 ko**: `advanced/skill-guide.md` 전체 구조 재작성 (각 9 fictional categories → 6 canonical categories, 각 37 nonexistent skill names 제거, 총 74). 다른 page-families는 en/zh와 동일하게 count correction. (ko 발견은 iter-2 재도출로 정정 — iter-1은 ko를 count-patch 대상으로 오분류.)

Run-phase는 manager-develop가 DDD cycle_type으로 실행 (문서 정합이므로 사실 기반 correction, characterization-test 불필요 — 대신 grep-based AC verification이 primary 검증).

---

## §B. Known Issues (사전 식별된 이슈)

| Issue | Impact | Mitigation |
|-------|--------|------------|
| ja 및 ko `skill-guide.md` 전체 재작성 필요 (각 37개 허구적 스킬명) | run-phase LOC 증가 | en/zh의 6-category 구조를 template으로 사용; M2(ko) + M3(ja)에서 독립 milestone으로 분리. iter-1은 ko를 count-patch로 오분류 → iter-2 재도출로 정정 |
| en/zh `skill-guide.md` 개념 설명 영역 허구적 스킬명 잔존 (en:10, zh:11) | category header는 정확하나 본문 정제 필요 | M4.5에서 허구적 스킬명을 real-skill equivalent로 교체 (예: `moai-lang-python` → `moai-domain-backend` + rules/moai/languages/ 안내). REQ-009 / AC-011이 0-match 검증 |
| `update.md`는 4-locale 모두 존재하나 statusline string은 en/zh에만 | 4-locale parity 예외 케이스 | research.md Coverage Map이 명시; ko/ja `update.md`는 count claim 무소유 — AC-001에서 per-locale grep이 자연스럽게 처리 (statusline string 없는 파일은 0-match로 정상) |
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
- [x] GEARS format REQs (REQ-001~009, Ubiquitous/Event-detected/State-driven/Capability gate/Unwanted 패턴 사용)
- [x] §E Exclusions 최소 1개 이상 (5개 exclusions 명시)
- [x] AC 11개 (10 MUST + 1 SHOULD), per-locale digit-boundary-anchored grep
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

**Verification**: AC-001 (en: 0 residual "31"+skill), AC-002 (en: ≥1 "32"+skill), AC-003 (en: humanize ≥1), AC-004 (en: Domain=9). (en in-body fictional-name cleanup은 M4.5에서 수행; AC-011은 M4.5 통합 검증.)

### M2 — ko locale structural rewrite (Priority High — iter-2 scope-up from count patch)

> **iter-2 revision (2026-06-18):** iter-1 scoped M2 to ko count correction only ("L62, L194 + humanize 추가"). Independent re-derivation confirmed ko `advanced/skill-guide.md` carries the IDENTICAL pre-v3 fictional taxonomy as ja (9 fictional categories, 37 nonexistent skill-name references, 3 missing canonical categories). M2 now mirrors M3's structural-rewrite scope.

ko locale은 단순 count patch가 아닌 구조 재작성 (M3 ja와 동일한 패턴):

- `advanced/skill-guide.md` **전체 카테고리 섹션 재작성**: 9 fictional categories (Foundation/Workflow/Domain/Language/Platform/Library/Tool/Framework/Design Tools) → 6 canonical categories (Foundation/Workflow/Domain/Reference/Meta-Harness/Design). 37 nonexistent skill names (`moai-lang-*` 16, `moai-platform-*` 4, `moai-library-*` 4, `moai-framework-*` 1, `moai-foundation-claude/philosopher/context` 3, 기타) 제거. 3 missing canonical categories (Reference, Meta/Harness, correctly-labeled Design) 도입. Domain에 `moai-domain-humanize` 추가 (4→9). Foundation sub-count 5→4, Workflow 11→10 정정.
- `advanced/builder-agents.md` L16 count correction
- `getting-started/introduction.md` L133, L156, L163 count correction
- `core-concepts/what-is-moai-adk.md` L7, L48, L267, L652 count correction (이미 claim 존재)

**Verification**: AC-006 (ko structural rewrite — `for loc in ja ko` loop의 ko 분기), AC-001~AC-004 ko locale, AC-007 (ko 6 categories).

### M3 — ja locale structural rewrite (Priority High — most complex)

ja locale은 단순 count patch가 아닌 구조 재작성 (M2 ko와 동일한 패턴):

- `advanced/skill-guide.md` **전체 카테고리 섹션 재작성**: 9 fictional categories (Foundation/Workflow/Domain/Language/Platform/Library/Tool/Framework/Design Tools) → 6 canonical categories (Foundation/Workflow/Domain/Reference/Meta-Harness/Design). 37 nonexistent skill names 제거 (M2와 동일 magnitude).
- `advanced/builder-agents.md` L17 count correction
- `getting-started/introduction.md` L133, L156, L163 count correction
- `core-concepts/what-is-moai-adk.md` L267 count correction + parity를 위해 L7/L48/L661에 count claim 추가 (현재 비대칭)

**Verification**: AC-006 (ja structural rewrite — `for loc in ja ko` loop의 ja 분기), AC-001~AC-004 ja locale, AC-007 (ja 6 categories).

### M4 — zh locale count correction + in-body cleanup (Priority High)

zh locale 5 pages. en과 동일 count pattern, locale-native idiom "32个技能"/"32项技能" 사용.

- `advanced/builder-agents.md` L15
- `advanced/skill-guide.md` L61 (sub-count breakdown의 Domain(8)→Domain(9) 수정 포함), L160
- `getting-started/introduction.md` L133, L156, L163
- `getting-started/update.md` L396
- `core-concepts/what-is-moai-adk.md` L7, L48, L267, L661

**Verification**: AC-001~AC-004 zh locale.

### M4.5 — en/zh in-body fictional-name cleanup (Priority High — iter-2 new milestone)

> **iter-2 addition (2026-06-18):** independent grep found en and zh `advanced/skill-guide.md` carry correct 6-canonical-category headers but still reference nonexistent skill names inside conceptual illustrations. en: 10 matches (Mermaid nodes L185/L333/L335, explicit-invocation L218/L220, ASCII tree L244, frontmatter example L266, auto-load comments L309/L314, callout L369). zh: 11 matches (analogous positions L45/L178/L325/L327/L211/L213/L236/L258/L301/L306/L360). ko+ja are handled by M2/M3 structural rewrite.

en 및 zh `advanced/skill-guide.md`의 개념 설명 영역 허구적 스킬명 정제:

- 각 허구적 참조를 real-skill equivalent로 교체 또는 재구성. 예: `moai-lang-python` → `moai-domain-backend` (Python 패턴은 `rules/moai/languages/python.md`를 통해 제공된다는 안내와 함께); `moai-library-mermaid` → `moai-ref-api-patterns` 또는 실제 라이브러리 참조로 교체; `moai-platform-supabase` → 해당 플랫폼 통합이 rules 기반으로 제공된다는 안내로 재구성.
- locale-native idiom 유지 (기계 번역 금지, REQ-007); Mermaid 노드 라벨, 코드 예제, ASCII tree, callout 모두 정정.
- M1(en)과 M4(zh)의 count correction과 동일 commit boundary에서 수행 (4-locale parity, REQ-006).

**Verification**: AC-011 (en/zh in-body fictional-name elimination — `for loc in en zh` loop, 0-match each).

### M5 — 4-locale parity verification + primary-source cross-check (Priority Medium)

모든 milestone 완료 후 통합 검증:

- AC-005 (sub-count sum = 31 specialized, mechanical find+grep) 전 locale 확인
- AC-007 (4-locale parity: 동일 count 32 + 동일 6-category 구조) parity diff
- AC-009 (primary-source `find` 출력 재현) 독립 실행
- AC-010 (`moai spec lint` 0 findings 재확인)
- AC-011 (en/zh in-body fictional-name 0-match) M4.5 완료 재확인

**Verification**: AC-005, AC-007, AC-009, AC-010, AC-011.

### M6 — Sync-phase handoff (Priority Low)

sync-phase 진입 준비:

- 모든 AC PASS 확인
- diff inspection으로 DOCSITE-001 6축 비중첩 재확인
- CHANGELOG entry 초안 (skill-count reconciliation)
- docs-truth.md §6 "Skill Count (32)" axis 추가 후속 (별도 commit, 본 SPEC scope 밖 — §D.4 forward-looking)

---

## §G. Anti-Patterns (회피해야 할 패턴)

1. **AP-COV-001 — 단일 로케일 correction 후 commit**: 4-locale parity 위반. 모든 correction은 단일 commit boundary 내에서 4 로케일 동시 적용. M1~M4.5는 논리적 분류이지 commit 분리가 아님 (단일 run-phase commit 권장).
2. **AP-COV-002 — glob grep로 전체 로케일 통합 검사**: per-locale drift mask. AC-001/AC-006/AC-011의 `for loc` loop를 우회해서는 안 됨.
3. **AP-COV-003 — ja OR ko locale에 en/zh patch 단순 적용**: ja 및 ko는 구조적 재작성 필요. count patch만 적용하면 fictional taxonomy가 잔존하여 AC-006 FAIL. (iter-1이 ko를 count-patch로 오분류한 것이 본 anti-pattern의 실제 사례 — iter-2 재도출로 정정.)
4. **AP-COV-004 — count를 "32"로 바꾸되 humanize 추가 누락**: AC-003 FAIL. count correction과 humanize 추가는 동일 milestone 내에서 수행.
5. **AP-COV-005 — user-owned harness skills를 count에 포함**: 34가 아닌 32. template-shipped만 count 대상. §E.4 Exclusions 방어.
6. **AP-COV-006 — en/zh category header만 정정하고 개념 설명 영역 허구적 스킬명 방치**: AC-011 FAIL. Mermaid·코드 예제·ASCII tree·callout 등 개념 설명 영역의 허구적 스킬명도 동일 milestone(M4.5)에서 정제. (iter-2 추가 — orchestrator 독립 grep이 발견.)

---

## §H. Cross-References

- spec.md §B REQs — 구현의 요구사항 근거
- acceptance.md §D ACs — milestone 검증 기준
- research.md §1~§3 — primary-source evidence + coverage map
- `.moai/specs/SPEC-V3R6-DOCS-DOCSITE-001/` — 선행 SPEC (6축 정합 완료)
- `.moai/project/codemaps/docs-truth.md` — docs-truth checklist (§6 axis 추가 후속)
