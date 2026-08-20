# t59 review verdict — 칸반/팀 모드 용어 정의 + 문서화

- Reviewer: review lane session (`release-v311` worktree; independent of the run lane)
- Card: t59 · Worktree: `.claude/worktrees/t59` · Branch: `WT-t59` @ `bd06ea127` (base `931e4138a`)
- Lens: default 4-perspective + docs-site i18n verification (per dispatch)
- Evidence read: `.moai/reports/t59/evidence.md` (44 lines, 5-section)
- Delta: 13 files, +430/−2 — rule detail companion (local + template mirror), README.ko.md, docs-site 4-locale page + `_meta.yaml` ×4 + `data/menu/main.yaml`, evidence
- Verdict location note: card-worktree placement (`.claude/worktrees/t59/.moai/reports/t59/review-verdict.md`) was refused by the worktree-isolation guard (this session may only write inside `release-v311`) — filed here per the dispatch's fallback clause, with this reason recorded.

## Verdict: PASS — with one advisory for the lead (zh translation-term drift, §F)

## 1. Dispatch focus items (all six verified)

| # | Focus | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | docs-site 4-locale same-commit obligation | All 4 `kanban-board-terms.md` created in the same commit `bd06ea127`; ko canonical → en/ja/zh derived with facts, figures, code blocks, identifiers (`run-a1b2c3`, `WT-t0`, `a1b2c3`, `a1b2c4`, `WT-t1`), and Mermaid direction preserved verbatim (4-way read of the full diff). `_meta.yaml` ×4 registered with title + weight 40 between trust-5 and verification-claim-integrity; `main.yaml` entry in the same neighbor position with 4-locale name map + `ref` — icon field absent, matching both neighbors (trust-5, VCI) which also carry no icon. Internal link targets (kanban-mode, moai-todo, harness-engineering ×4 locales = 12 files) all exist. Section parity 5/5/5/5 reproduced | PASS |
| 2 | Glossary placement in the lazy companion | `kanban-dispatch-detail.md` (paths-scoped) gains "Terminology — the board vocabulary" ahead of "The board"; always-loaded `kanban-dispatch.md` absent from the diff (unmodified, as claimed). `TestAlwaysLoadedTokenBudget` re-run in the t59 worktree: `ok internal/config 0.526s` | PASS |
| 3 | README ko-only | diff touches `README.ko.md` only — README.md/en/ja/zh untouched. Addition is H3 (`### 다섯 세션이 쓰는 말`); H2 count reproduced at 11 (pre-existing ko deficit vs en/ja/zh unchanged — t47's territory). Sole external link `https://adk.mo.ai.kr/ko/core-concepts/kanban-board-terms` is whitelisted-domain | PASS |
| 4 | Template neutrality | Mirror byte-identical proven two ways: `git ls-tree bd06ea127` shows the same blob `9b0f270ea` for local and template paths; `git show \| shasum -a 256` gives identical digests (`da28a0fe…`) for both. Example identifiers are neutral (`a1b2c3`/`t0`); grep for `t59\|tjv7iy\|SPEC-[A-Z]\|2026-` across the template file and 4 new docs pages matches only the generic `<SPEC-ID>` placeholder and pre-existing lines — no internal identifiers. `go test ./internal/template/ -run 'Mirror\|Parity\|Leak'`: `ok 1.406s` | PASS |
| 5 | Docs discipline | Body emoji scan (perl, U+1F000–1FAFF / 2600–26FF / FE0F) over the 4 new pages: 0 matches. Mermaid forbidden directions (`flowchart\|graph` + `LR\|RL`): 0 matches; `flowchart TD` ×2 in each of the 4 files. URL blacklist (`docs.moai-ai.dev\|adk.moai.com\|adk.moai.kr`): 0 matches over docs-site/content + README.ko.md. hugo rebuilt in this review: exit 0, WARN/ERROR count 0, `Pages 179/177/177/177` (matches evidence), sitemap.xml present, `index.html` for the new page rendered in all 4 locales | PASS |
| 6 | Integration cautions recorded | evidence Residual-risk carries both: t113 (three-column consolidation would require updating the "six columns" wording across glossary/README/docs-site — t113 already declares the kanban-dispatch.md template-mirror companion) and t47 (README ko section title "다섯 세션이 쓰는 말" needs localization when en/ja/zh are re-derived) | PASS |

## 2. Evidence claim reproduction (11/11 claims dispositioned)

Reproduced this review with matching observations: hugo warning-free (exit 0, 0 WARN), sitemap, blacklist 0, TD-only 0, section parity 5/5/5/5, emoji 0, mirror byte-identical (strengthened to SHA-256 + blob-hash equality), Mirror/Parity/Leak ok, TestAlwaysLoadedTokenBudget ok, README H2 11 unchanged (addition is H3).

Substituted, not re-run: `make build` — the commit carries no catalog.yaml change (diff-level confirmation of the "catalog unchanged" claim) and the blob-hash equality proves the deployed template content; a rebuild would reproduce the binary, not new information.

Not re-run (gap carried, same as evidence): full 143×4 locale-parity ratchet — CI's M6 4-locale gate owns the final call; new-page-only parity plus zero modified existing pages (diff shows none) close the divergence path.

## 3. Baseline attribution

All observations are from THIS review session: git object reads against `bd06ea127` from `release-v311` (shared object store), file greps / hugo build / go test against the t59 worktree tree at `bd06ea127`. Outputs quoted verbatim in §1–2.

## 4. Gaps

- No browser-visual check of sidebar placement/new-badge (hugo build proves page generation only) — same gap as evidence.
- Locale-parity ratchet over the full site deferred to CI M6 (same as evidence).
- `make build` binary reproduction skipped (see §2 substitution rationale).

## 5. Residual risks

- t113 interaction: "six columns" wording in this card's outputs (rule glossary, README.ko, docs-site ×4) must be co-updated when t113 lands — recorded in evidence; recommend the t113 dispatch explicitly lists t59 artifacts.
- t47 interaction: the README.ko H3 section becomes a derivation source; title localization needed — recorded in evidence.
- zh term drift → §F (the one advisory).

## F. Advisory — zh translation-term drift (lead's discretion; not a FAIL)

The evidence claims "기존 번역어 일관성 유지: 리드 세션/동반 세션 (kanban-mode.md 관례)". Verified per locale:

- ko: existing kanban-mode.md uses "리드 세션"/"동반 세션" — new page matches. **consistent**
- en: "lead session"/"companion session" — matches. **consistent**
- ja: "リード" — matches. **consistent**
- zh: existing kanban-mode.md uses **主导会话** for lead (L106, L128) and **伴随会话** for companion (L128); the new zh glossary uses **主控** and **同伴会话** instead. **INCONSISTENT** — the consistency claim does not hold for zh.

Mitigating: the existing zh page itself is not fully uniform (L244 also uses 领导者), so "the" convention is arguable — but the most frequent rendering (主导会话 ×2 / 伴随会话) differs from the glossary's choice. A glossary page whose purpose is term unification introducing a second zh word pair for the same concepts splits the very vocabulary it defines.

Suggested fix (one-file string substitution, no build impact): in `docs-site/content/zh/core-concepts/kanban-board-terms.md`, 主控 → 主导会话 and 同伴会话 → 伴随会话 (occurrences in the table, lane-vs-column section, journey section, and the mermaid subgraph labels as applicable).

No HARD i18n rule is violated (blacklist, TD-only, emoji, 4-locale obligation, structure preservation all pass); this is a documentation-quality advisory. Lead may issue a rider before merge or defer to a follow-up card.
