# SPEC-V3R6-SKILL-COMPRESS-001 Plan — Top-5 Skill Body Compression

## 1. Implementation Strategy

5개 skill을 **차례대로 1개씩** 압축한다. 한 번에 모두 진행 금지 (회귀 위험 격리). 각 skill에 대해 동일한 Compression Loop을 적용한다.

### 1.0 공통 Compression Loop (skill-agnostic)

각 skill에 대해 다음 5-step을 적용:

1. **Inventory**: Level 2 body의 모든 H2/H3 섹션 enumeration + word count per section (`awk` 또는 manual count).
2. **Triage**: 각 섹션을 다음 3 카테고리로 분류:
   - **KEEP_INLINE** — 핵심 워크플로우, decision logic, trigger keyword 인용 → 유지 (압축 표현으로 rewrite 가능)
   - **MOVE_TO_LEVEL_3** — verbose example, multi-language code listing, exhaustive reference table → `<skill>/references/<topic>.md`로 이전 + Level 2에 link 1줄로 대체
   - **REMOVE** — 중복, deprecated 안내, 다른 skill로 옮겨진 absorbed 내용 → 삭제 + HISTORY에 justification 1줄
3. **Verify keyword coverage**: 각 `triggers.keywords` entry에 대해 압축 후 body에서 grep — REQ-SCM-009 (keyword 또는 synonym 보존) 보장.
4. **Word count check**: `wc -w SKILL.md` 결과가 target cap 이하인지 확인 (예: testing ≤ 2,000w).
5. **Template mirror**: `cp -p` 또는 `diff -q` 확인하여 `internal/template/templates/.claude/skills/<skill>/`와 byte-identical 보장.

### 1.1 `moai-workflow-testing` 압축 (3,153w → ≤ 2,000w; -1,150w 이상)

**현재 구조** (M1 inventory 기반 예상):
- H2 Quick Reference (~250w)
- H2 Implementation Guide (~1,800w, DDD testing / debugging / refactoring / performance / code review)
- H2 Advanced Patterns (~600w, multi-language adaptations)
- H2 Quality Metrics / Works Well With / Rationalizations (~500w)

**Compression strategy**:
- KEEP_INLINE: Quick Reference, TRUST 5 핵심 5-dimension, DDD 워크플로우 (PRESERVE 단계 핵심), PR review 패턴 헤드라인, Rationalizations / Red Flags / Verification (evolvable blocks 보존)
- MOVE_TO_LEVEL_3:
  - `references/multi-language-adaptations.md` ← Python/JS/TS/Go/Rust 도구 mapping 표
  - `references/automated-code-review-detail.md` ← 5-agent confidence scoring 상세
  - `references/performance-profiling-walkthrough.md` ← profiling 6-step 상세
  - (modules/ 디렉토리는 별도 SPEC scope, 본 SPEC는 SKILL.md body 텍스트만 외부화)
- REMOVE: redundant 5-step 반복 설명, Stage 1-6 outline 중 narrative duplicates

**예상 결과**: ~1,800w (target hit)

### 1.2 `moai-workflow-spec` 압축 (2,394w → ≤ 1,700w; -700w 이상)

**현재 구조** (예상):
- Quick Reference + EARS Five Patterns 정의
- Implementation Guide (Constitution Reference, SPEC Workflow Stages 6단계, EARS Deep Dive, Requirement Clarification Process)
- Plan-Run-Sync Integration + Parallel Worktree
- Advanced + Quality Metrics

**Compression strategy**:
- KEEP_INLINE: EARS Five Patterns 핵심 (5 patterns × 1 line each), SPEC 3-file structure (spec.md/plan.md/acceptance.md), Plan-Run-Sync 6-stage 헤드라인, Lifecycle 3-Level (spec-first/spec-anchored/spec-as-source)
- MOVE_TO_LEVEL_3:
  - `references/ears-examples.md` ← EARS 5 패턴 각각 예제 + Test strategy 상세
  - `references/clarification-workflow.md` ← Step 0 Assumption Analysis / Step 0.5 Root Cause / Step 1-4 상세
  - `references/sdd-2025-lifecycle.md` ← Lifecycle transition rules 상세
  - `references/migration-guide.md` (이미 reference/ 디렉토리에 있음) 재이용
- REMOVE: 중복된 EARS pattern 설명 (Quick + Deep Dive에서 같은 내용), Quality Metrics 일부 중복

**예상 결과**: ~1,500w (target hit)

### 1.3 `moai-workflow-project` 압축 (2,068w → ≤ 1,400w; -668w 이상)

**현재 구조** (예상):
- Quick Reference (Module Architecture)
- Implementation Guide (Documentation / Language / Template Optimization)
- Core Workflows (Initialization / Documentation Generation / Template Optimization)
- Advanced + Resources
- Template Optimization (absorbed) / Docs Generation (absorbed) / JIT Document Loading (absorbed)

**Compression strategy**:
- KEEP_INLINE: 4 core capabilities (documentation / language init / template optimization / jit-docs) headline 각 1-2줄, configuration structure (project type / language config) 핵심, performance metrics short table
- MOVE_TO_LEVEL_3:
  - `references/initialization-workflow.md` ← Complete project initialization 3-step + parameters 상세
  - `references/multilingual-localization.md` ← Language detection / agent prompt localization 상세
  - `references/template-library.md` ← FastAPI/React/Vue/Next.js boilerplate enumeration
  - `references/docs-generators.md` ← Sphinx/MkDocs/TypeDoc/OpenAPI/Nextra 상세
- REMOVE: 중복된 Capabilities 나열 (Quick + Implementation), absorbed sections의 verbose explanation (이미 다른 SPEC들에서 absorb 완료)

**예상 결과**: ~1,200w (target hit)

### 1.4 `moai-domain-design-handoff` 압축 (2,039w → ≤ 1,600w; -439w 이상)

**현재 구조** (예상, no modules/references 디렉토리):
- Quick Reference (5-file handoff bundle 정의)
- Implementation Guide (prompt template, context, references, acceptance, checklist)
- Brand-absent fallback / Section regeneration
- Phase 7 integration 상세

**Compression strategy**:
- KEEP_INLINE: 5-file bundle 정의 (prompt/context/references/acceptance/checklist 1줄씩), handoff 워크플로우 7-Phase 헤드라인, brand-absent fallback decision tree headline
- MOVE_TO_LEVEL_3 (신규 디렉토리 생성):
  - `references/handoff-bundle-templates.md` ← 5 file content templates 상세 + example
  - `references/brand-absent-fallback-workflow.md` ← fallback decision tree + section regeneration steps
  - `references/phase7-integration.md` ← /moai brain Phase 7 integration 상세
- REMOVE: redundant phase enumeration

**예상 결과**: ~1,400w (target hit)

### 1.5 `moai-meta-harness` 압축 (2,010w → ≤ 1,600w; -410w 이상)

**현재 구조** (예상):
- Quick Reference (meta-skill purpose, revfactory/harness attribution)
- 7-Phase workflow (revfactory/harness pattern adaptation)
- Role profile / Template Generation
- Dynamic harness artifact 생성 (`moai-harness-*`, `.claude/agents/harness/*`, `.moai/harness/*`)
- Apache-2.0 attribution

**Compression strategy**:
- KEEP_INLINE: 7-Phase workflow 순서 + 각 Phase 헤드라인 1줄, role profile 5종 enumeration (researcher / analyst / architect / implementer / reviewer), dynamic artifact 3 디렉토리 (skills, agents, harness configs), Apache-2.0 attribution notice
- MOVE_TO_LEVEL_3 (신규 디렉토리, 단 `seeds/`는 보존):
  - `references/phase-by-phase-walkthrough.md` ← 7-Phase 각 단계 input/output/agent 상세
  - `references/role-profile-definitions.md` ← 5 role profile 각 상세 + isolation policy
  - `references/dynamic-artifact-spec.md` ← `moai-harness-*` 생성 패턴 + 등록 절차
- REMOVE: redundant attribution repetition (NOTICE.md에 이미 있음 — link만 보존)

**예상 결과**: ~1,400w (target hit)

## 2. Files Affected

### 2.1 In-scope Modified (10 SKILL.md, byte-paired)

| Local | Template Mirror |
|---|---|
| `.claude/skills/moai-workflow-testing/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-testing/SKILL.md` |
| `.claude/skills/moai-workflow-spec/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md` |
| `.claude/skills/moai-workflow-project/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` |
| `.claude/skills/moai-domain-design-handoff/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-design-handoff/SKILL.md` |
| `.claude/skills/moai-meta-harness/SKILL.md` | `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` |

### 2.2 In-scope New (Level 3 references, ~15 files, byte-paired)

추정 신규 파일 (실제 개수는 M1 inventory + Triage 결과에 따라 ±3 변동 가능):

| Skill | New files (예상 1-3개 each) |
|---|---|
| moai-workflow-testing | `references/multi-language-adaptations.md`, `references/automated-code-review-detail.md`, `references/performance-profiling-walkthrough.md` (3개) |
| moai-workflow-spec | `references/ears-examples.md`, `references/clarification-workflow.md`, `references/sdd-2025-lifecycle.md` (3개; `reference/migration-guide.md`는 기존 존재) |
| moai-workflow-project | `references/initialization-workflow.md`, `references/multilingual-localization.md`, `references/template-library.md`, `references/docs-generators.md` (4개) |
| moai-domain-design-handoff | `references/handoff-bundle-templates.md`, `references/brand-absent-fallback-workflow.md`, `references/phase7-integration.md` (3개; 신규 `references/` 디렉토리) |
| moai-meta-harness | `references/phase-by-phase-walkthrough.md`, `references/role-profile-definitions.md`, `references/dynamic-artifact-spec.md` (3개; 신규 `references/` 디렉토리. `seeds/` 보존) |

**합계 신규 ~16 files × 2 (local + template mirror) = 32 file creations**

### 2.3 Catalog Hash Regeneration

`internal/template/catalog.yaml`에 5 skill 모두 등록된 경우 (SPEC-V3R6-CATALOG-SSOT-001 메모리 참조: moai-harness-learner / moai-meta-harness hash 갱신 선례), SKILL.md 콘텐츠 변경 시 hash 재생성 의무:

- `go run gen-catalog-hashes.go --all` 1회 실행 (M7 milestone)
- 변경된 hash entries 검증 (5 skill 모두 또는 일부)

### 2.4 Touchless (PRESERVE 의무)

- 28+ 다른 skill 디렉토리 일체 불변
- `.claude/rules/` 일체 (RULES-COMPRESS-001 / RULES-PATH-SCOPE-001 담당)
- `.claude/agents/` 일체
- Foundation skills (`moai-foundation-core`, `moai-foundation-cc`, `moai-foundation-quality`, `moai-foundation-thinking`) — 별도 SPEC 후보
- 다른 SPEC 디렉토리 (4 Wave 1 SPECs, 9 Wave 0/V3R6 SPECs)
- runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/`, `internal/hook/.moai/`)
- testing skill의 `modules/` 22 sub-디렉토리 콘텐츠 (link 정합성만 보장)
- `seeds/` (meta-harness)

## 3. Implementation Order

| Milestone | Scope | Deliverable | Verification |
|---|---|---|---|
| M1 | Baseline measurement + git status hygiene + 5 skill inventory | `wc -w` 5 skill = expected (3153/2394/2068/2039/2010) + git log 직접 수정 commit 부재 확인 | 정량 baseline 기록 (in run-phase progress.md) |
| M2 | `moai-workflow-testing` 압축 | SKILL.md ≤ 2,000w + Level 3 3 files + template mirror | AC-SCM-001 / AC-SCM-007 / AC-SCM-009 |
| M3 | `moai-workflow-spec` 압축 | SKILL.md ≤ 1,700w + Level 3 3 files + template mirror | AC-SCM-002 / AC-SCM-007 / AC-SCM-009 |
| M4 | `moai-workflow-project` 압축 | SKILL.md ≤ 1,400w + Level 3 4 files + template mirror | AC-SCM-003 / AC-SCM-007 / AC-SCM-009 |
| M5 | `moai-domain-design-handoff` 압축 | SKILL.md ≤ 1,600w + Level 3 3 files + template mirror | AC-SCM-004 / AC-SCM-007 / AC-SCM-009 |
| M6 | `moai-meta-harness` 압축 | SKILL.md ≤ 1,600w + Level 3 3 files + template mirror | AC-SCM-005 / AC-SCM-007 / AC-SCM-009 |
| M7 | Catalog hash regen + aggregate check | `go run gen-catalog-hashes.go --all` + `wc -w` sum ≤ 8,200w | AC-SCM-006 / AC-SCM-010 |
| M8 | Regression suite + cross-platform | `go test ./internal/template/...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 + frontmatter byte-identity 검증 | AC-SCM-008 / AC-SCM-011 / AC-SCM-012 |

Implementation order는 **risk-ordered**: M2 (testing) 가장 큰 절감 (-1,353w) 먼저 → Compression Loop 패턴 검증 → M3-M6 점진 적용. M7-M8 aggregate validation.

## 4. Pre-flight Checks

manager-develop이 run-phase 시작 시 다음을 확인 (B 카테고리 known issues 일부 적용):

```bash
# 1. 현재 branch + HEAD
git branch --show-current  # main 또는 feat/SPEC-V3R6-SKILL-COMPRESS-001
git rev-parse HEAD

# 2. Baseline word count 5 skill (현재 측정값과 일치 확인)
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  wc -w .claude/skills/$s/SKILL.md
done
# Expected: 3153 / 2394 / 2068 / 2039 / 2010

# 3. Template mirror byte-identical baseline
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  diff -q .claude/skills/$s/SKILL.md internal/template/templates/.claude/skills/$s/SKILL.md
done
# Expected: 모두 "identical"

# 4. Cross-platform build baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...

# 5. 영향 패키지 cross-SPEC 충돌 사전 스캔
grep -r "Retired\|superseded" .claude/skills/moai-workflow-testing/ .claude/skills/moai-workflow-spec/ .claude/skills/moai-workflow-project/ .claude/skills/moai-domain-design-handoff/ .claude/skills/moai-meta-harness/ 2>/dev/null || echo "no conflicts"

# 6. Frontmatter trigger 보존 baseline (압축 후 비교용)
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  awk '/^triggers:/,/^---$/' .claude/skills/$s/SKILL.md > /tmp/trigger-$s-before.txt
done

# 7. Catalog 등록 여부 사전 확인
grep -A 1 "moai-workflow-testing\|moai-workflow-spec\|moai-workflow-project\|moai-domain-design-handoff\|moai-meta-harness" internal/template/catalog.yaml | head -30

# 8. Wave 1 다른 3 SPEC 디렉토리 격리
ls .moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/ .moai/specs/SPEC-V3R6-RULES-COMPRESS-001/ .moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/ 2>/dev/null
# 본 SPEC는 이들과 file overlap 0
```

## 5. Brownfield Strategy

### 5.1 압축 대상 (Modify, in-scope)

5 SKILL.md (local + template mirror).

### 5.2 Level 3 분리 (Add, in-scope)

`<skill>/references/` 신규 .md 파일 (local + template mirror).

### 5.3 PRESERVE (Touchless, out-of-scope)

본 SPEC가 건드리지 않는 영역:

- 5 skill의 frontmatter (Level 1 metadata) — REQ-SCM-008 HARD
- 5 skill의 기존 modules/, reference/, schemas/, templates/, seeds/ 콘텐츠
- 28+ 다른 skill (foundation-* / domain-* / workflow-* / meta-* 나머지)
- `.claude/rules/` 전체
- `.claude/agents/` 전체
- `.claude/commands/` 전체
- 모든 SPEC 디렉토리 (본 SPEC 자체 외)
- internal/ Go code 일체 (단 `internal/template/catalog.yaml`의 hash entry는 M7에서 regenerate; 다른 yaml field 불변)
- runtime-managed files
- 무관 untracked / modified files

### 5.4 Greenfield 부분

Level 3 references/ 디렉토리는 design-handoff + meta-harness의 경우 신규 생성 (현재 부재). `mkdir -p` + 파일 생성으로 처리.

## 6. PRESERVE List (구체적 enumeration)

다음 파일/디렉토리는 본 SPEC 수정 절대 금지:

```
.claude/skills/moai-foundation-*/
.claude/skills/moai-domain-* (단 moai-domain-design-handoff 제외)
.claude/skills/moai-workflow-* (단 testing/spec/project 제외)
.claude/skills/moai-meta-* (단 harness 제외)
.claude/skills/moai-platform-*/
.claude/skills/moai-context7-integration/
.claude/skills/moai-tool-*/
.claude/skills/moai-cc-*/
.claude/skills/moai/  (nested namespace 별개)
.claude/rules/
.claude/agents/
.claude/commands/
.claude/hooks/
.claude/output-styles/
.claude/agent-memory/  (CLAUDE.local.md §2)
.moai/specs/ (단 본 SPEC 자체 디렉토리 제외)
.moai/research/
.moai/docs/
.moai/config/
.moai/state/  (runtime-managed)
.moai/harness/  (runtime-managed)
.moai/cache/
.moai/logs/
.moai/manifest.json
internal/hook/.moai/  (runtime-managed)
internal/template/templates/ (단 5 skill mirror 제외)
internal/ Go code (단 catalog.yaml hash entry 제외)
docs-site/
Makefile
go.mod / go.sum
README.md / CHANGELOG.md / CLAUDE.md
CLAUDE.local.md
.git/
```

특히 다음 untracked files은 본 SPEC 작업 중 절대 건드리지 마라:

- `.moai/research/moai-adk-current-state-2026-05-22.md`
- `.moai/research/v3.0-design-2026-05-22.md`  (참조만, 수정 금지)
- `docs-site/content/{en,ja,ko,zh}/book/`
- `docs-site/data/menu/extra.yaml`
- `docs-site/layouts/_default/redirect.html`
- `docs-site/scripts/gen_menu.py`
- `docs-site/static/book/`
- `internal/hook/.moai/`
- `internal/template/github_tmpl_parse_test.go`

## 7. Risk Mitigation Tactics

| Risk | Mitigation in Implementation |
|---|---|
| R-SCM-001 Trigger Logic 누락 | M2-M6 각 milestone 끝에 keyword grep 자동 검증 (AC-SCM-007); 압축 전후 keyword list diff |
| R-SCM-002 Level 3 미등록 | REQ-SCM-011 cross-reference 무결성 (AC-SCM-009); Level 3는 frontmatter 등록 불필요 |
| R-SCM-003 Template Drift | M2-M6 각 milestone 끝에 `diff -rq` 검증 (AC-SCM-008); MultiEdit pair pattern (local + template 동시) |
| R-SCM-004 Dangling Links | `grep` + `test -f` cross-ref 정합성 (AC-SCM-009) |
| R-SCM-005 Baseline Drift | M1 Pre-flight에서 baseline 재측정 의무; target은 absolute cap |

## 8. Coordination with Parallel Wave 1 SPECs

본 SPEC는 다음 3 Wave 1 SPECs와 **file overlap 0**이지만 같은 Wave 1 batch에 머지될 수 있음:

- `SPEC-V3R6-RULES-PATH-SCOPE-001` (Tier M) — `.claude/rules/moai/` paths frontmatter scope
- `SPEC-V3R6-RULES-COMPRESS-001` (Tier S) — `.claude/rules/moai/` rule body compression
- `SPEC-V3R6-SKILL-CONSOLIDATE-001` (Tier M) — skill 통폐합 (agency / copywriter / designer 등)

**File overlap 검증**:
- 본 SPEC: `.claude/skills/moai-workflow-testing|spec|project/`, `.claude/skills/moai-domain-design-handoff/`, `.claude/skills/moai-meta-harness/` + template mirrors
- RULES-PATH-SCOPE-001 / RULES-COMPRESS-001: `.claude/rules/moai/` 전용
- SKILL-CONSOLIDATE-001: 다른 skill 그룹 (agency / copywriter / designer / team-pattern-cookbook 등) — 본 SPEC 대상 5 skill과 disjoint

겹침 위험 영역: `internal/template/catalog.yaml` hash entries — 본 SPEC가 5 skill hash 갱신, SKILL-CONSOLIDATE-001도 catalog touch 가능성. M7 hash regen 시 `--all` flag로 전체 갱신하므로 충돌 없음.

## 9. Self-estimation (plan-auditor preview)

Tier M PASS threshold = 0.80.

| Dimension | Self-est | Rationale |
|---|---|---|
| D1 Completeness | 0.90 | 12 REQs (009-012 trigger preservation / cross-ref / catalog 포함), 5 Risks (mitigation explicit), Out of Scope 8 항목 |
| D2 Traceability | 0.92 | REQ↔AC 100% (REQ-SCM-001..012 ↔ AC-SCM-001..012, traceability matrix in acceptance.md) |
| D3 Testability | 0.88 | AC binary verifiable (`wc -w`, `grep`, `diff`, `go test`), edge cases 명시 |
| D4 Risk Coverage | 0.85 | R-SCM-001/003 mitigation에 자동 검증 명시; R-SCM-002/005 likelihood Low |
| **Weighted** | **~0.89** | **PASS (margin +0.09 above 0.80)** |

플랜 auditor 호출은 orchestrator 별도 진행 (본 위임 scope 외).

## 10. Token / Effort Estimate

본 SPEC implementation은 **5 skill 순차 압축** + Level 3 분리 작업. priority 라벨로만 표기 (시간 추정 금지 per CLAUDE.md §11):

- Priority: P1
- Sequence: M1 → M2 → M3 → M4 → M5 → M6 → M7 → M8 (8 milestones strict sequential)
- Parallel opportunity: 본 SPEC 내부에서 M2-M6는 독립적이므로 향후 Agent Teams 모드에서 implementer × 5 parallel 가능 (단 본 SPEC 자체는 sequential strategy로 작성)
