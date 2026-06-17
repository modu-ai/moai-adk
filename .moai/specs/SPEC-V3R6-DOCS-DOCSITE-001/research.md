# Research — SPEC-V3R6-DOCS-DOCSITE-001

> docs-site 4-locale drift inventory against docs-truth canonical facts.
> 이 파일은 run-phase에서 각 AC이 어떤 drift를 커버하는지 매핑하는 근거 자료이다.
> **모든 사실은 docs-truth.md가 인용한 PRIMARY source로 재검증되었다** (2026-06-17, tree `7666bd178`).

## §A. Methodology

1. docs-truth.md 5 axes 전부 1차 source에서 재검증
2. `docs-site/content/{en,ko,ja,zh}/` (420페이지)에 대해 grep 기반 facts-bearing surface 탐지
3. 발견된 facts-bearing page-family에 대해 4 locale별 drift 측정
4. 인접 drift (5 axes 밖)는 별도 기록, scope 밖 명시

### §A.1 docs-truth facts 재검증 결과 (2026-06-17, tree `7666bd178`)

| Fact | docs-truth claim | Primary source | Re-verification |
|------|------------------|----------------|-----------------|
| §1 Agent catalog | 8 retained (7 MoAI-custom + 1 Explore built-in) | `ls -1 .claude/agents/moai/*.md` | **PASS** → 7 files |
| §2 SPEC status enum | 8 lowercase values | `internal/spec/status.go` `ValidStatuses` | PASS (docs-truth 인용 그대로) |
| §3 Frontmatter | 12 required fields | `internal/spec/lint.go` `FrontmatterSchemaRule` | PASS |
| §4.1 Terminal verbs | `moai --help` + `internal/cli/` (40 files, 119 AddCommand) | `grep -rnE '\.AddCommand\(' internal/cli/` | PASS |
| §4.2 `/moai` commands | 17 commands | `ls -1 .claude/commands/moai/*.md` | **PASS** → 17 files |
| §5 GLM tier-models | `glm-5.2[1m]` High/Opus, `glm-4.7` Medium/Sonnet, `glm-4.5-air` Low/Haiku | `internal/config/defaults.go` lines 40-57 | **PASS** → `DefaultGLMHigh = "glm-5.2[1m]"` 확인 |

**모든 docs-truth fact가 1차 source에서 재확인되었다. 실패한 fact 없음 (blocker 없음).**

## §B. Facts-Bearing Page Inventory

docs-site에서 5 docs-truth axes 중 하나 이상의 사실을 실제로 담고 있는 page-family:

| # | Page-family (relative to locale/) | Locales | Axes carried | Stale? |
|---|-----------------------------------|---------|--------------|--------|
| 1 | `core-concepts/what-is-moai-adk.md` | en, ko, ja, zh | §1 (agents), §2 (status, partial) | **ja/zh stale** ("28 agents", "52 skills", mermaid lists archived as active) |
| 2 | `advanced/agent-guide.md` | en, ko, ja, zh | §1 (agents) | en/ko clean (8 retained + archived framing); **ja/zh need verification** |
| 3 | `multi-llm/_index.md` | en, ko, ja, zh | §5 (GLM) | **ko stale** (`GLM-5.1` as Opus); **en/ja/zh stubs** (parity gap) |
| 4 | `multi-llm/model-policy.md` | en, ko, ja, zh | §5 (GLM) | **ko full** (3001 bytes); **en/ja/zh stubs** (222 bytes — parity gap) |
| 5 | `getting-started/faq.md` | en, ko, ja, zh | §1 (agents — model assignment) | **en/ko carry 8-retained table; ja/zh missing it** |
| 6 | `getting-started/introduction.md` | en, ko, ja, zh | §1 (agents) | en/ko carry "8 agents"; **ja/zh need verification** |

**Facts-bearing surface**: 6 page-family × 4 locale. D2 live-grep 재조사 결과, **실제 편집 대상은 ≈10-11개 파일** (4 page-family × 4 locale의 사실-정정 + stub-확장 혼합). Tier L 근거는 단순 파일 수(Tier M 5-15 범위)가 아니라 (a) 4-locale 동시 조정 의무, (b) 6 facts-axis × page-family 교차 매항, (c) no-IA-redesign 제약 하 내용 사실만 정정, (d) 모든 사실의 primary-source 추적 가능성 요건이 결합된 것이다 (D9 — "24 pages / 4× exceed" 과장 framing 제거).

### §B.1 Page-family별 상세 drift

#### PF-1: `core-concepts/what-is-moai-adk.md` (가장 가시적인 parity failure)

**D2 re-survey (live grep, tree `8108d4311`, 2026-06-17)** — stale "28" agent-count per locale (digit-boundary protected regex `(^|[^0-9])28\s*(specialized\s*)?agents?|(^|[^0-9])28\s*개.*(전문\s*)?에이전트|(^|[^0-9])28\s*个.*Agent|(^|[^0-9])28\s*(の|個).*エージェント`):

| Locale | stale "28" count | stale lines | positive "8" count | Archived framing |
|--------|-------------------|-------------|---------------------|------------------|
| en | **0** | — | 2 (L7, L226 "8 retained agents") | ✓ (L226 "12 archived ... were consolidated") |
| ko | **2** | L226 active prose "28개의 전문 에이전트", L666 code-block "28개 AI 에이전트 정의" | 4 (L7, L24, L48, L498) | ✓ (L226) |
| ja | **5** | L7, L48, L226, L496, L670 (전체 pre-consolidation 상태) | **0** (positive "8" 부재 — 최악 drift) | ✗ (0 archived framing) |
| zh | **1** | L674 code-block "28个AI Agent定义" (L226 prose는 이미 "8个专业Agent"로 정정됨) | 5 (L7, L24, L48, L226, L500) | partial |

**Drift 분포 재확정 (D2)**: en=0, ko=2, ja=5, zh=1. **ja가 primary prose-drift locale** (5 stale, 0 positive). ko는 L226 active prose 1곳 + L666 code-block 1곳. zh는 L674 code-block 1곳만 (prose는 이미 최신화). en은 clean. 이전 research.md (iter-1)의 "en/ko clean, zh heavily drifted" 주장은 **오류**였다 (Verify-Don't-Assume 위반 — 정확한 분포는 ja primary, ko 1 active prose + 1 code-block, zh 1 code-block only, en 0).

**Mermaid structural drift (D3 re-survey)**: ko/ja/zh 3개 locale의 mermaid diagram `Managers` subgraph가 `M6["manager-spec ..."]`, `M7["manager-docs ..."]` phantom duplicate 슬롯으로 "(8)" count를 맞추고, `M5["sync-auditor ..."]`를 Manager로 잘못 분류한다 (sync-auditor는 evaluator). en은 이 구조적 drift가 없다 (phantom M6/M7 = 0). Live grep `grep -cE 'M6\["manager-spec|M6\["manager-docs|M7\["manager-spec|M7\["manager-docs'`: en=0, ko=2, ja=2, zh=2.

#### PF-2: `advanced/agent-guide.md`

| Locale | "8 retained" | Archived historical section |
|--------|--------------|-----------------------------|
| en | ✓ (lines 23, 57, 195) | ✓ (lines 213-216, "Twelve agents were archived on 2026-05-25") |
| ko | ✓ | ✓ (lines 214-217) |
| ja | (verify) | (verify) |
| zh | (verify) | (verify) |

**Drift**: en/ko clean. ja/zh는 run-phase M2에서 verification + 정정 필요.

#### PF-3: `multi-llm/_index.md` (GLM-5.1 stale + 4-locale content 격차)

| Locale | Bytes | GLM-5.1 stale? | glm-5.2[1m] present? |
|--------|-------|----------------|----------------------|
| en | 208 (stub) | ✗ | ✗ (stub) |
| ko | 2680 | **✓ (line 17, 23: `GLM-5.1` as Opus)** | ✗ (stale) |
| ja | 208 (stub) | ✗ | ✗ (stub) |
| zh | 208 (stub) | ✗ | ✗ (stub) |

**Drift**: ko만 full content를 가지나 **`GLM-5.1` (stale)을 Opus tier로 나열**. en/ja/zh는 near-empty stub (parity gap). M3에서 ko 정정 + en/ja/zh 확장 (parity) 필요.

ko/multi-llm/_index.md line 17, 23-25 현재 내용:
```
| **모델** | GLM-5.1, GLM-4.7, GLM-4.5-Air, 무료 모델 |     ← STALE (GLM-5.1)
| Opus | GLM-5.1 | $2.00 | $8.00 |                  ← STALE
| Sonnet | GLM-4.7 | $0.60 | $2.20 |                 ← OK
| Haiku | GLM-4.5-Air | $0.20 | $1.10 |              ← OK
```
정정 후 (docs-truth §5 기준):
```
| Opus | glm-5.2[1m] | ... |   ← 1M context
| Sonnet | glm-4.7 | ... |
| Haiku | glm-4.5-air | ... |
```

#### PF-4: `multi-llm/model-policy.md`

| Locale | Bytes |
|--------|-------|
| en | 222 (stub) |
| ko | 3001 (full) |
| ja | 222 (stub) |
| zh | 222 (stub) |

**Drift**: ko만 full content. en/ja/zh stub. M3에서 4-locale parity 필요.

#### PF-5: `getting-started/faq.md`

| Locale | "8 retained agents" model-assignment table |
|--------|---------------------------------------------|
| en | ✓ (line 87: "8 retained agents use model assignment based on tier") |
| ko | ✓ |
| ja | **✗ (0 matches)** |
| zh | **✗ (0 matches)** |

**Drift**: en/ko carry the model-assignment table; ja/zh missing. M5에서 4-locale parity.

#### PF-6: `getting-started/introduction.md`

en line 133, 163: "**8** specialized AI agents". 4 locale verification는 M5에서.

## §C. Drift Inventory (D-001 .. D-008 + ADJ-1)

| ID | Page-family | Locale(s) | Drift | Severity | AC | Milestone |
|----|-------------|-----------|-------|----------|-----|-----------|
| D-001 | core-concepts/what-is-moai-adk.md | ja (5 stale, primary), ko (1 active prose + 1 code-block), zh (1 code-block) | "28 agents" stale count — digit-boundary protected grep verified live: en=0/ko=2/ja=5/zh=1 | Critical | AC-001, AC-005 | M2 |
| D-001b | core-concepts/what-is-moai-adk.md (mermaid) | ko, ja, zh | `M6/M7` phantom duplicate manager-spec/docs slots + `M5 sync-auditor` misclassified as Manager — structural, not archived-name | High | AC-002 | M2 |
| D-002 | core-concepts/what-is-moai-adk.md | en, ko | (verify-only) archived names in correct historical framing | — (OK) | AC-002 | M2 verify |
| D-003 | advanced/agent-guide.md | ja, zh | verify archived historical framing consistency | High | AC-002 | M2 |
| D-004 | multi-llm/_index.md | ko | `GLM-5.1` as Opus tier (stale — `grep -rcE 'glm-5\.1\|GLM-5\.1' ko/multi-llm/` = 1 file) | Critical | AC-003 | M3 |
| D-004b | multi-llm/_index.md | en, ja, zh, ko | canonical `glm-5.2[1m]` absent in ALL 4 locales (`grep -rcE 'glm-5\.2' {en,ko,ja,zh}/multi-llm/` = 0 files each) | High | AC-003, AC-004 | M3 |
| D-005 | multi-llm/_index.md | en, ja, zh | near-empty stubs (parity gap) | Medium | AC-004 | M3 |
| D-006 | multi-llm/model-policy.md | en, ja, zh | stubs vs ko full (parity gap) | Medium | AC-004 | M3 |
| D-007 | getting-started/faq.md | en, ko, ja, zh | model-assignment TABLE already exists in all 4 locales (awk row-count = 4 each); D5 re-survey: no missing-table drift, only potential content-fact drift | Low (downgraded) | AC-008 | M5 verify |
| D-008 | getting-started/introduction.md | ja, zh | verify "8 agents" parity | Low | AC-005 | M5 |
| D-009 | multiple (what-is-moai-adk.md L216, ko L49+L216) | en (1), ko (3 occurrences), ja (1), zh (0) | "18 languages" — FROZEN rule CONST-V3R2-004 위반 (16 languages). **IN-SCOPE per REQ-DOCSITE-008** (D6). Live: `grep -rE '(^\|[^0-9])18\s*(languages?\|개.*언어\|言語\|个.*语言)'` = en=1,ko=3,ja=1,zh=0 occurrences | Medium | AC-011 | M4 |
| ADJ-1 | multiple (introduction, skill-guide, builder-agents) | en, ko, ja, zh | "31 skills" adjacent drift — OUT OF SCOPE (skill count is uncoupled, not a FROZEN rule) | Low (out of scope) | — | — (research only) |

### §C.1 Per-locale drift summary (D2 re-survey, live grep verified)

| Locale | Stale facts | Parity gaps | Overall staleness |
|--------|-------------|-------------|-------------------|
| en | minimal (clean post-README-001 wave) | multi-llm stubs (D-005/006), 18-languages L216 (D-009) | Lowest |
| ko | GLM-5.1 (D-004), "28" L226 active prose + L666 code-block (D-001), mermaid M6/M7 phantom (D-001b), 18-languages L49+L216 ×3 (D-009) | — | Medium |
| ja | **"28" 5곳 (L7/L48/L226/L496/L670) — primary prose-drift locale, positive "8" = 0** (D-001), mermaid M6/M7 phantom (D-001b), agent-guide verify (D-003), multi-llm stub (D-005/006), 18-languages L216 (D-009) | extensive | **Highest** |
| zh | "28" L674 code-block only (D-001), mermaid M6/M7 phantom (D-001b), agent-guide verify (D-003), multi-llm stub (D-005/006) | extensive | High |

**결론 (D2 정정)**: **ja가 가장 stale** (5 stale "28" + 0 positive "8" + 4 parity gaps). 이전 research.md (iter-1)의 "en/ko clean, zh heavily drifted" 주장은 Verify-Don't-Assume 위반으로 **오류**. 정확한 분포: **ja primary (5 stale), ko (2: 1 active prose + 1 code-block), zh (1 code-block only), en (0)**. zh는 prose가 이미 "8个专业Agent"로 정정되어 code-block 1곳만 남은 partial-done 상태이다.

## §D. Adjacent Drift (OUT OF SCOPE — 기록만)

본 SPEC은 docs-truth 6 axes (5 original + language count per REQ-DOCSITE-008/D6)에 한정. 아래 drift는 인접하나 scope 밖:

1. **"31 skills"** — en/getting-started/introduction.md (line 133, 156, 163), en/advanced/skill-guide.md (line 65, 127, 167), en/core-concepts/what-is-moai-adk.md (line 7, 48, 652), en/advanced/builder-agents.md (line 15). 실제 skill count는 별도 검증 필요. Skill count는 FROZEN rule이 아니며 uncoupled되어 있어 **별도 SPEC 권장** (SKILL-COUNT-001 가칭). **OUT OF SCOPE**.
2. **"34,220 lines of Go code / 32 packages"** — en/getting-started/introduction.md line 132. 시간에 따라 변동. 정적 fact 아님. **out of scope**.

**"18 languages"는 본 SPEC scope에 포함된다** (REQ-DOCSITE-008 / D-009 / D6). 16-language count는 FROZEN HARD rule CONST-V3R2-004이므로, "18"을 방치하면 조정 완료 페이지에 FROZEN rule 모순이 잔류한다. Live grep `(^[^0-9])18\s*(languages?|개.*언어|言語|个.*语言)`: en=1 occurrence (L216), ko=3 (L49+L216×2), ja=1 (L216), zh=0. M4에서 4-locale `18 → 16` 조정.

## §E. i18n Parity Tooling

`scripts/docs-i18n-check.sh` (8KB, executable) 가 존재하며, `.github/workflows/docs-i18n-check.yml` CI workflow가 이를 호출한다.

- **목적**: docs-site 4-locale structural parity 검증 (페이지 존재 여부, 메타데이터 일치)
- **AC 활용**: AC-004 (GLM parity), AC-009 (umbrella)에서 `bash scripts/docs-i18n-check.sh` 호출로 기계적 검증
- **한계**: 이 도구는 structural parity (페이지 존재/메타데이터)를 검증하며, **content fact correctness는 검증하지 않음**. 따라서 본 SPEC의 ACs는 grep 기반 fact 검증 + i18n-check 기반 structural parity를 병용한다.

추가 도구: `scripts/i18n-validator/` 디렉토리 (별도 validator). run-phase에서 활용 가능.

## §F. Risks & Open Questions

1. **ja/zh mermaid diagram 재작성 범위** — ja/zh `core-concepts/what-is-moai-adk.md`의 mermaid diagram (lines ~241-267)은 archived agent들을 active Manager/Expert 노드로 나열. 이를 8 retained agents로 재작성해야 하지만, diagram 구조가 복잡할 수 있음. M2에서 en/ko의 mermaid diagram을 참조 템플릿으로 사용.
2. **multi-llm 4-locale 확장 시 번역 품질** — ko의 GLM tier-models table을 en/ja/zh로 확장 시, 단순 기계번역이 아닌 각 locale의 자연스러운 표현으로 번역 필요. 단, 사실(fact)은 동일해야 함.
3. **"researcher" 문자열 disambiguation** — 16 파일이 "researcher"를 언급하나 대부분은 비-에이전트 일반 명사. archived 에이전트 `researcher`만 정정 대상. per-file human review 필요 (AC-002의 grep은 same-line archived framing을 요구하므로 일반 명사 사용은 자동 통과).
4. **en/ko already-clean 페이지 verify-only 처리** — en/ko core-concepts/what-is-moai-adk.md는 수정 불필요. 단, AC-002 mermaid check는 en/ko에도 적용되므로 en/ko mermaid diagram이 archived 노드를 active로 나열하지 않는지 확인 필요.

## §G. Cross-References

- `.moai/project/codemaps/docs-truth.md` — canonical facts checklist
- `.moai/specs/SPEC-V3R6-DOCS-CODEMAPS-V3-001/` — docs-truth baseline 저술 SPEC
- `.moai/specs/SPEC-V3R6-DOCS-V3-README-001/` — README 2-file 선행 SPEC (`7666bd178`)
- `scripts/docs-i18n-check.sh` — i18n parity tool
- `.github/workflows/docs-i18n-check.yml` — CI parity gate
- `internal/config/defaults.go` lines 40-57 — GLM tier-models primary source
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — archived-agent migration table
