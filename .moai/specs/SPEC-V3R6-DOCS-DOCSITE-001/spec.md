---
id: SPEC-V3R6-DOCS-DOCSITE-001
title: "docs-site 4-locale content reconciliation against docs-truth canonical facts"
version: "0.2.0"
status: draft
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
era: V3R6
tags: "docs-site, i18n, docs-truth, reconciliation, 4-locale"
depends_on: [SPEC-V3R6-DOCS-CODEMAPS-V3-001, SPEC-V3R6-DOCS-V3-README-001]
---

# SPEC-V3R6-DOCS-DOCSITE-001 — docs-site 4-locale content reconciliation

## §A. History / Context

이 SPEC은 Sprint 14 Docs-v3 코호트의 3번째 SPEC이다. 선행 관계:

- **SPEC-V3R6-DOCS-CODEMAPS-V3-001** (completed, `4a6f4b4d3`) — `.moai/project/codemaps/docs-truth.md` canonical facts baseline 저술. 5개 facts axis (agent catalog / SPEC status enum / frontmatter 12 fields / CLI surface / GLM tier-models) 정의.
- **SPEC-V3R6-DOCS-V3-README-001** (closed, `7666bd178`) — repo-root README.md/README.ko.md 2-file reconciliation. en/ko 사이의 사실(facts) drift 해소.

DOCSITE-001는 동일한 docs-truth baseline을 따라 **`docs-site/content/{en,ko,ja,zh}/**/*.md` 4-locale 전체**의 사실 drift를 해소한다. README-001이 2-file scope였던 반면, DOCSITE-001는 4-locale 트리(각 105페이지) 중 사실을 실제로 담고 있는 facts-bearing page-families를 대상으로 한다. research.md §B의 live-grip 재조사(D2) 결과, 실제 edit surface는 4 page-family × 4 locale ≈ **10-11개 실제 편집 파일** (일부 locale은 stub 확장, 일부는 사실 정정). Tier L 근거는 단순 edit-count가 아니라 (a) 4-locale 동시 조정 의무, (b) 6 facts-axis × page-family 교차 매항, (c) IA 재설계 금지 under no-IA-redesign 제약, (d) 모든 사실의 primary-source 추적 가능성 요건이 결합된 것이다 (research.md §B.1 Tier L 근거 재서술 참조).

**이것은 사실 조정(factual reconciliation) SPEC이다.** docs-site IA 재설계, 스타일링 변경, 빌드 설정 수정이 아니다. 모든 사실 주장(factual claim)은 docs-truth.md가 인용한 PRIMARY source로 거슬러 올라가 추적되어야 한다.

## §B. Requirements (GEARS)

### REQ-DOCSITE-001 — Agent catalog: 8 retained (4-locale)

The docs-site pages **shall** state the MoAI agent catalog as exactly **8 retained agents** (7 MoAI-custom + 1 Anthropic built-in `Explore`), matching `.claude/agents/moai/*.md` (= 7 files) + the Explore built-in, across all 4 locales (en, ko, ja, zh).

**When** a docs-site page references the agent-catalog count, the page **shall** state "8 retained agents" (locale-translated) — NOT "28 agents", "27 agents", or any pre-consolidation count.

### REQ-DOCSITE-002 — Archived agents: historical context only (4-locale)

**When** a docs-site page names archived agents (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the 6 `expert-*` agents), the page **shall** frame them exclusively as archived/consolidated historical context — explicitly marking them as NOT active and citing `.claude/rules/moai/workflow/archived-agent-rejection.md` for migration guidance — across all 4 locales.

**Where** a docs-site page renders an agent-catalog diagram (mermaid or tabular), the diagram **shall** list ONLY the 8 retained agents as active — archived agent names MUST NOT appear as active nodes.

### REQ-DOCSITE-003 — GLM tier-models: glm-5.2[1m] as High/Opus (4-locale)

**Where** a docs-site page in `multi-llm/` enumerates the GLM→Claude tier mapping, the page **shall** state the default High/Opus tier model as `glm-5.2[1m]` (1M context), Medium/Sonnet as `glm-4.7`, Low/Haiku as `glm-4.5-air` — matching `internal/config/defaults.go` lines 40-57 — across all 4 locales.

**When** a docs-site `multi-llm/` page names the Opus-tier GLM model, the page **shall not** state `glm-5.1` (the pre-activation default). The `glm-5.2[1m]` activation is the canonical post-2026-06-15 fact.

### REQ-DOCSITE-004 — GLM facts 4-locale parity

The docs-site **shall** carry the GLM tier-models facts (REQ-DOCSITE-003) with structurally equivalent content across all 4 locales — each locale's `multi-llm/` section MUST carry the tier-models table OR all 4 locales MUST uniformly omit it. The current state (ko carries a full table; en/ja/zh are near-empty stubs) is a parity gap requiring reconciliation.

### REQ-DOCSITE-005 — Agent catalog 4-locale parity

The docs-site **shall** carry the 8-retained-agents fact (REQ-DOCSITE-001) and the archived-agents historical framing (REQ-DOCSITE-002) with consistent facts across all 4 locales. The current state — en/ko `core-concepts/what-is-moai-adk.md` updated to "8 retained" but ja/zh still carrying "28 agents" / "52 skills" and a mermaid diagram listing archived agents as active — is a 4-locale parity failure requiring reconciliation.

### REQ-DOCSITE-006 — CLI surface: 17 `/moai` commands (4-locale)

**Where** a docs-site page enumerates the `/moai` slash-command set, the page **shall** state the count as **17 commands** (matching `ls -1 .claude/commands/moai/*.md | wc -l` → 17), across all 4 locales. Stale counts (e.g. historical "10", "12") MUST be removed.

### REQ-DOCSITE-007 — SPEC status enum + frontmatter (4-locale)

**Where** a docs-site page enumerates the SPEC status lifecycle or the frontmatter required-fields list, the page **shall** state the 8-value status enum (`draft` · `planned` · `in-progress` · `implemented` · `completed` · `superseded` · `archived` · `rejected`) and the 12 required frontmatter fields — matching `internal/spec/status.go` and `internal/spec/lint.go` — across all 4 locales.

### REQ-DOCSITE-008 — Language count: 16 (not 18) 4-locale reconciliation

**Where** a docs-site page states the supported-language count, the page **shall** state **16 languages** — matching CLAUDE.local.md §15 (16-language neutrality) and the FROZEN HARD rule CONST-V3R2-004 (`.claude/rules/moai/development/coding-standards.md` § Language Policy) — across all 4 locales. Stale "18 languages" / "18개 언어" / "18言語" / "18种语言" claims MUST be corrected to 16. This REQ is in-scope (unlike "31 skills") because the 16-language count is a FROZEN rule (CONST-V3R2-004), so leaving "18" in reconciled pages would preserve a contradiction with a canonical HARD rule.

## §C. Acceptance Criteria Summary

See `acceptance.md` for the full AC matrix (§D). Each AC is mechanically verifiable via `grep -rl` / `wc -l` / `diff` across the 4 locale trees. ACs enforce both the per-fact correctness AND the 4-locale parity invariant.

## §D. Constraints

1. **Factual reconciliation only** — docs-site IA 재설계 금지. 페이지 병합/분리/이동 금지.
2. **6 docs-truth axes** — agent catalog, SPEC status enum, frontmatter, CLI surface, GLM tier-models, **language count (16, per FROZEN CONST-V3R2-004)**. 인접 drift 중 "31 skills"는 research.md에 기록하되 DOCSITE-001 scope에서 제외 (별도 SPEC 권장). "18 languages"는 FROZEN rule CONST-V3R2-004 위반이므로 본 SPEC scope에 포함 (REQ-DOCSITE-008).
3. **4-locale parity is load-bearing** — 단일 locale만 수정하고 다른 3개를 방치하면 parity 위반이다. 각 facts axis의 수정은 4-locale 동시 적용.
4. **No invented facts** — 모든 사실은 docs-truth.md가 인용한 primary source로 거슬러 올라가 추적 가능해야 한다.
5. **Styling/build config excluded** — `docs-site/nextra.config.*`, `vercel.json` 수정 금지. content facts only.

## §E. Exclusions (What NOT to Build)

### Out of Scope — Adjacent and excluded surfaces

- **Repo-root README.md / README.ko.md** — 이미 README-001 (`7666bd178`)이 처리. 중복 금지.
- **Go 코드 / CLAUDE.md / `.claude/` 설정** — 이 SPEC의 scope 밖. docs-site `content/` 트리만.
- **docs-site IA 재설계** — 섹션 재구조, 페이지 병합/분리, 네비게이션 변경 금지.
- **docs-site 스타일링 / 빌드 설정** — `nextra.config.*`, `vercel.json`, `_meta.yaml` 구조 변경 금지 (내용 사실 정정에 수반되는 최소한의 frontmatter 조정은 허용).
- **"31 skills" 인접 drift** — 별도 SPEC 대상. DOCSITE-001는 docs-truth 5 axes + language-count(REQ-008)에 한정. research.md에 기록만.
- **"18 languages"** — FROZEN rule CONST-V3R2-004 위반이므로 **본 SPEC scope 포함** (REQ-DOCSITE-008). "31 skills"와 달리 16-language count는 codified HARD rule이므로 방치 시 모순 잔류.
- **Lang-translated prose 품질 개선** — 번역 품질 polish가 아닌, 사실(fact) 정정이 목적.

## §F. Cross-References

- `.moai/project/codemaps/docs-truth.md` — canonical facts checklist (navigation aid, NOT SSOT)
- `.moai/specs/SPEC-V3R6-DOCS-CODEMAPS-V3-001/` — docs-truth baseline 저술 SPEC (completed)
- `.moai/specs/SPEC-V3R6-DOCS-V3-README-001/` — repo-root README 2-file reconciliation (closed)
- `internal/spec/status.go` — §2 SPEC status enum primary source
- `internal/spec/lint.go` — §3 frontmatter 12 required fields primary source
- `internal/config/defaults.go` lines 40-57 — §5 GLM tier-models primary source
- `.claude/agents/moai/*.md` (7 files) + `Explore` built-in — §1 agent catalog primary source
- `.claude/commands/moai/*.md` (17 files) — §4.2 `/moai` command set primary source
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — archived-agent migration table
- `scripts/docs-i18n-check.sh` — i18n parity-check tool (ACs에서 활용 가능)
- `.github/workflows/docs-i18n-check.yml` — CI parity gate

## §G. Verification Approach

모든 AC는 `grep -rl`, `wc -l`, `diff` 명령으로 4 locale 트리에 걸쳐 기계적으로 검증 가능하다:

- **부정 grep (stale facts 제거 확인)**: `grep -rl 'glm-5\.1' docs-site/content/{en,ko,ja,zh}/multi-llm/` → empty
- **긍정 grep (canonical facts 존재 확인)**: `grep -rl 'glm-5\.2\[1m\]' docs-site/content/{en,ko,ja,zh}/multi-llm/` → ≥1 file per locale (parity)
- **Parity diff**: `diff <(grep -c '8 retained\|8 Retained' en/.../agent-guide.md) <(grep -c '...' ja/.../agent-guide.md)` → 0

상세 AC 표는 `acceptance.md` §D 참조.

## §H. Risks

1. **Parity regression** — 한 locale만 수정하고 push하면 CI i18n check가 fail하거나, 통과하더라도 parity 위반 상태로 잔류. 4-locale 동시 수정 강제.
2. **Archived-agent 이름 leak 재발** — archived 이름을 "active" context에 실수로 재도입할 위험. 모든 수정은 historical-framing-only 원칙 준수.
3. **mermaid diagram 동기화** — mermaid 안의 agent 노드 목록은 prose와 별도로 관리됨. prose만 고치고 diagram을 잊으면 parity 위반 잔류.両方 수정.
4. **"31 skills" scope creep** — 인접 drift를 만지기 시작하면 SPEC이 무한 확장. 5 axes 엄격 준수.
