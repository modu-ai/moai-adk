# Plan — SPEC-V3R6-DOCS-DOCSITE-001

> Implementation plan for docs-site 4-locale content reconciliation.
> Tier: **L** (4-locale × 6-facts-axis factual reconciliation under no-IA-redesign + primary-source-traceability constraints).

## §A. Context

DOCSITE-001는 docs-truth baseline(`.moai/project/codemaps/docs-truth.md`)을 따라 `docs-site/content/{en,ko,ja,zh}/**/*.md` 4-locale 중 사실을 담고 있는 페이지들을 조정한다. research.md §B의 live-grep drift inventory에 따르면 6개 page-family가 6 docs-truth axes(agent catalog / SPEC status / frontmatter / CLI / GLM tier / language count) 중 하나 이상의 사실을 담고 있다. D2 재조사 결과 실제 edit surface는 **≈10-11개 편집 파일** (4 page-family × 4 locale, stub-확장 + 사실-정정 혼합)이다.

**Tier L 근거 (edit-count 이상)**: 단순 파일 수(10-11)는 Tier M(5-15 files) 범위에 들지만, (a) 4-locale 동시 조정 의무 (단일 locale 수정 = parity 위반), (b) 6 facts-axis × page-family 교차 매항, (c) IA 재설계 금지 under no-IA-redesign 제약 (구조 변경 없이 내용 사실만 정정), (d) 모든 사실의 primary-source 추적 가능성 요건이 결합되어 5-artifact Tier L ceremony가 정당화된다. "24 pages / 4× exceed" 과장 framing은 제거 (D9).

**가장 큰 parity gap**: `core-concepts/what-is-moai-adk.md`의 en은 "8 retained agents"로 최신화되었고 ko/zh는 부분 최신화(코드 블록 "28" 잔류), **ja는 전체 pre-consolidation 상태** ("28の専門エージェント" 5곳, positive "8" 0곳). 또한 ko/ja/zh 3개 locale의 mermaid diagram이 `M6/M7` phantom duplicate manager-spec/docs 슬롯으로 count를 맞추는 구조적 drift를 갖는다 (D3). 이것이 본 SPEC의 가장 가시적인 drift이다.

## §B. Known Issues (from research.md §C drift inventory)

| Issue ID | Page-family | Locales affected | Drift type |
|----------|-------------|------------------|------------|
| D-001 | `core-concepts/what-is-moai-adk.md` | ja (5 stale, primary), ko (2: L226 prose + L666 code-block), zh (1: L674 code-block) | "28 agents" stale count — live-grep verified: en=0/ko=2/ja=5/zh=1 |
| D-001b | `core-concepts/what-is-moai-adk.md` (mermaid) | ko, ja, zh | `M6/M7` phantom duplicate manager-spec/docs slots + `M5 sync-auditor` misclassified as Manager (structural) — en=0/ko=2/ja=2/zh=2 |
| D-002 | `core-concepts/what-is-moai-adk.md` | en, ko | archived-agent names appear ONLY in correct historical framing (OK — verify only) |
| D-003 | `advanced/agent-guide.md` | en, ko, ja, zh | verify all 4 locales carry "8 retained" + archived historical framing consistently |
| D-004 | `multi-llm/_index.md` | ko | `GLM-5.1` as Opus tier (stale — 1 file) |
| D-004b | `multi-llm/_index.md` | en, ko, ja, zh | canonical `glm-5.2[1m]` absent in ALL 4 locales (0 files each) |
| D-005 | `multi-llm/_index.md` | en, ja, zh | near-empty stubs — no GLM tier-models table (parity gap with ko) |
| D-006 | `multi-llm/model-policy.md` | en, ja, zh (vs ko) | ko has full content (2680 bytes); en/ja/zh are stubs (222 bytes) |
| D-007 | `getting-started/faq.md` | en, ko, ja, zh | model-assignment table ALREADY present in all 4 locales (4 rows each) — verify-only, NOT missing (D5 re-survey) |
| D-008 | `getting-started/introduction.md` | ja, zh | verify "8 agents" claim parity (en/ko have it) |
| D-009 | multiple (`what-is-moai-adk.md` L216, ko L49+L216) | en (1 occ), ko (3 occ), ja (1 occ), zh (0) | "18 languages" — FROZEN rule CONST-V3R2-004 위반 (16 languages). **IN-SCOPE** (REQ-DOCSITE-008) |
| ADJ-1 | multiple (introduction, skill-guide, builder-agents) | en, ko, ja, zh | "31 skills" adjacent drift — OUT OF SCOPE (skill count uncoupled, not a FROZEN rule) |

## §C. Pre-flight Checks

run-phase 진입 전 확인 항목:

1. `git status` clean — docs-site content 수정 중 충돌 방지
2. `moai spec audit --json` 본 SPEC이 draft 상태로 인식되는지 확인
3. `scripts/docs-i18n-check.sh` baseline 실행 — 수정 전 현재 parity 상태 캡처
4. `ls -1 .claude/agents/moai/*.md | wc -l` → 7, `ls -1 .claude/commands/moai/*.md | wc -l` → 17 재확인 (docs-truth facts re-verification)

## §D. Constraints (from spec.md §D)

1. Factual reconciliation only — IA 재설계 금지
2. 5 docs-truth axes only — "31 skills" / "18 languages" 제외
3. 4-locale parity load-bearing — 단일 locale 수정 금지
4. Styling/build config excluded
5. No invented facts — 모든 사실은 primary source 추적

## §E. Self-Verification (Plan-phase)

- [x] docs-truth.md 6 axes 전부 1차 source에서 재검증 완료 (7 agents, 17 commands, `glm-5.2[1m]`, 8 status values, 12 frontmatter fields, 16 languages per CONST-V3R2-004)
- [x] 4 locale × 6 page-family drift inventory 구축 + D2 live-grep 재조사 (research.md §B/§C)
- [x] Tier L 근거 재확정 (≈10-11 실제 편집 파일 + 4-locale 동시 조정 + 6 facts-axis 교차 + no-IA-redesign + primary-source 추적 — D9 과장 제거)
- [x] scripts/docs-i18n-check.sh 존재 확인 (AC tooling 활용 가능)
- [x] README-001 (`7666bd178`)이 repo-root README 처리 — 본 SPEC은 docs-site content만
- [x] **iter-2**: 모든 AC regex를 live tree(`8108d4311`)에서 사전 검증 — digit-boundary 보정, locale-native idiom, no-op check 제거 (D1/D3/D4/D5/D10)

## §F. Milestones (Tier L — 6 milestones, facts-axis × locale-pair)

### M1 — Drift inventory 확정 + AC 바인딩

**Scope**: research.md §B drift inventory를 확정하고, 각 AC가 어떤 drift를 커버하는지 매핑. 수정 전 baseline grep 결과를 `progress.md` §E.2에 캡처.

**Files**: `research.md`, `acceptance.md`, `progress.md`

**Exit criteria**:
- drift inventory 8개 issue (D-001..D-008) 전부 AC에 매핑
- baseline grep output (`grep -rl 'glm-5\.1'`, `grep -rl '28.*agent'` etc.) 캡처

### M2 — Agent catalog 4-locale 조정 + mermaid structural rewrite (REQ-DOCSITE-001, -002, -005)

**Scope**: `core-concepts/what-is-moai-adk.md` + `advanced/agent-guide.md` 전 4 locale에 걸쳐 "8 retained agents" 사실과 archived historical framing 통일. **D3 추가**: ko/ja/zh mermaid `Managers` subgraph의 phantom M6/M7 duplicate slots + M5 sync-auditor misclassification 구조적 drift 재작성 (en mermaid를 템플릿으로 사용).

**Files** (live-grep D1/D3 확정):
- `docs-site/content/ja/core-concepts/what-is-moai-adk.md` — **5곳 "28" stale → "8 retained"** (L7, L48, L226, L496, L670), positive "8" 형식 도입 (현재 0), mermaid M6/M7 phantom 제거 + 4-manager + 2-evaluator + 1-builder + Explore 구조로 재작성
- `docs-site/content/ko/core-concepts/what-is-moai-adk.md` — **L226 active prose "28개의 전문 에이전트" → "8개"**, L666 code-block "28개 AI 에이전트 정의" → "8개", mermaid M6/M7 phantom 제거
- `docs-site/content/zh/core-concepts/what-is-moai-adk.md` — **L674 code-block "28个AI Agent定义" → "8个"** (prose L226 이미 done), mermaid M6/M7 phantom 제거
- `docs-site/content/{en,ko,ja,zh}/advanced/agent-guide.md` — 4-locale archived framing 일관성 검증 + 정정 (en/ko는 verify-only)

**Exit criteria** (AC-001/AC-002/AC-005 digit-boundary protected greps):
- `for loc in en ko ja zh; do grep -cE '(^|[^0-9])28\s*(specialized\s*)?agents?|...' docs-site/content/$loc/core-concepts/what-is-moai-adk.md; done` → all 0 (PRE-FIX: en=0/ko=2/ja=5/zh=1)
- `for loc in en ko ja zh; do grep -cE 'M6\["manager-spec|M6\["manager-docs|M7\["manager-spec|M7\["manager-docs' ...; done` → all 0 (PRE-FIX: en=0/ko=2/ja=2/zh=2)
- 4 locale 모두 positive "8" ≥ 1 (PRE-FIX: en=2/ko=4/ja=0/zh=5 → POST-FIX ja ≥ 1)

### M3 — GLM tier-models 4-locale 조정 (REQ-DOCSITE-003, -004)

**Scope**: `multi-llm/_index.md` + `multi-llm/model-policy.md` 전 4 locale에 걸쳐 `glm-5.2[1m]` canonical 사실 통일. **D10 정정**: ko의 stale `GLM-5.1` 제거 + en/ja/zh stubs를 ko와 동등한 내용으로 확장 (parity). 현재 4 locale 모두 `glm-5.2` 부재 (0 files each).

**Files**:
- `docs-site/content/ko/multi-llm/_index.md` — `GLM-5.1` → `glm-5.2[1m]` 정정 (L17, L23)
- `docs-site/content/{en,ja,zh}/multi-llm/_index.md` — ko 기준으로 tier-models table 확장 (4-locale parity), `glm-5.2[1m]` 포함
- `docs-site/content/{en,ja,zh}/multi-llm/model-policy.md` — stub → full content (ko 기준)

**Exit criteria** (AC-003/AC-004 per-locale greps):
- `for loc in en ko ja zh; do grep -rcE 'glm-5\.1|GLM-5\.1' docs-site/content/$loc/multi-llm/ | grep -v ':0$' | wc -l; done` → all 0 (PRE-FIX: en=0/ko=1/ja=0/zh=0)
- `for loc in en ko ja zh; do grep -rcE 'glm-5\.2' docs-site/content/$loc/multi-llm/ | grep -v ':0$' | wc -l; done` → all ≥ 1 (PRE-FIX: all 0)
- `scripts/docs-i18n-check.sh` multi-llm section parity PASS

### M4 — CLI surface + SPEC status/frontmatter + language count 4-locale 조정 (REQ-DOCSITE-006, -007, -008)

**Scope**: CLI `/moai` command count (17), SPEC status enum / frontmatter 12 fields, **language count 16 (D6/REQ-008)**. research.md §B 확정 시 해당 page-family 식별.

**Files**:
- `utility-commands/moai.md`, `advanced/` 내 SPEC lifecycle 페이지 (CLI/status/frontmatter — research.md 확정 후 특정)
- `docs-site/content/{en,ko,ja,zh}/core-concepts/what-is-moai-adk.md` — **"18 languages" → "16 languages"** 조정 (en L216, ko L49+L216, ja L216; zh는 이미 clean)
- 기타 "18 languages" 출현 페이지 (research.md §C D-009 참조)

**Exit criteria** (AC-006/AC-007/AC-011):
- `/moai` command count "17"이 4 locale에서 일관 (stale count 제거)
- SPEC status enum 8-value / frontmatter 12-field가 담긴 페이지가 4 locale에서 사실 일치
- `for loc in en ko ja zh; do grep -rE '(^|[^0-9])18\s*(languages?|개.*언어|言語|个.*语言)' docs-site/content/$loc/ | wc -l; done` → all 0 (PRE-FIX: en=1/ko=3/ja=1/zh=0)

### M5 — FAQ + introduction 4-locale 조정 (REQ-DOCSITE-001, -005 보강)

**Scope**: `getting-started/faq.md` (**D5: verify-only — table 이미 4-locale 보유**) + `getting-started/introduction.md` (4-locale "8 agents" parity 검증).

**Files**:
- `docs-site/content/{en,ko,ja,zh}/getting-started/faq.md` — **verify-only** (D5 live-grep: 4 locale 모두 model-assignment table 4 rows 보유). content-fact drift만 있는지 검증, table 추가 불필요.
- `docs-site/content/{en,ko,ja,zh}/getting-started/introduction.md` — "8 agents" claim parity 검증

**Exit criteria** (AC-008/AC-005):
- `for loc in en ko ja zh; do awk '/model.*assign|모델.*할당|モデル.*割り当て|模型.*分配|8\s*(个|保留|retained|개)/{f=1} f&&/^\|.*manager-(spec|develop|docs|git)/{n++} END{print n+0}' docs-site/content/$loc/getting-started/faq.md; done` → all ≥ 4 (PRE-FIX: all 4)
- `getting-started/introduction.md` "8 agents" claim 4-locale 일치

### M6 — Final AC sweep + parity 검증

**Scope**: acceptance.md §D ACs 전부 4 locale에 걸쳐 기계적 실행. `scripts/docs-i18n-check.sh` full run. parity regression 없음 확인.

**Files**: 모든 docs-site content (검증 only, 추가 수정 없음)

**Exit criteria**:
- acceptance.md §D 모든 AC PASS
- `scripts/docs-i18n-check.sh` exit 0
- 4 locale × 5 docs-truth axes 모두 일치
- `moai spec audit` 본 SPEC lint 0 findings

## §G. Anti-Patterns

- **AP-1**: 단일 locale만 수정 후 push — parity 위반. 각 milestone은 4-locale 동시 수정.
- **AP-2**: "31 skills" / "18 languages" 인접 drift 만지기 — scope creep. research.md에 기록만.
- **AP-3**: mermaid diagram을 prose 수정에서 누락 — diagram이 archived agent를 active로 잔류. prose + diagram 동시 수정.
- **AP-4**: docs-truth.md를 SSOT로 인용 — docs-truth.md는 navigation aid. primary source 인용.
- **AP-5**: archived-agent 이름을 active context에 재도입 — historical-framing-only 원칙 위반.

## §H. Cross-References

- `spec.md` §B REQ-DOCSITE-001..008 — 본 plan이 커버하는 requirements (008 = language count, D6)
- `acceptance.md` §D — milestone exit criteria의 기계적 검증 ACs (AC-001..011, iter-2 digit-boundary 보정 + live-grep evidence)
- `research.md` §B/§C — drift inventory (D-001..D-009 + D-001b/D-004b, D2 re-survey) 상세
- `.moai/specs/SPEC-V3R6-DOCS-CODEMAPS-V3-001/` — docs-truth baseline
- `.moai/specs/SPEC-V3R6-DOCS-V3-README-001/` — repo-root README 선행 SPEC
