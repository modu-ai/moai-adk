# t47 review verdict — README ko new-skeleton promoted to canonical; en/ja/zh re-derived

- Reviewer: review lane session (dispatched by lead-tjv7iy, lens: default 4-perspective + 4-locale documentation precision)
- Card: t47 · Worktree: `.claude/worktrees/t47` · Branch: `WT-t47` (4 commits on base `ca8c0b593`: `3fa31ef4d` rules → `3c887d508` ko canonical → `c07804841` re-derive → `81b3f0d5a` evidence)
- Delta reviewed: `ca8c0b593..WT-t47` — 11 files (5 rule/skill/Runner + 4 READMEs + 2 evidence docs); **no docs-site file, no template file** (hns-* surfaces are user-owned harness files, correctly unmirrored)
- Evidence read: `.moai/reports/t47/verification.md` (VCI 5-section, honest gaps) + `absorption-map.md` (26-row disposition table)
- Method: card-worktree writes blocked by session isolation → verdict in release-v311 tree. Measurements via `git archive WT-t47` → `/tmp/t47-review`.

## Verdict: PASS — with 1 one-word nit and 2 recorded 부기 (lead-record items)

## Dispatch focus items (8/8 checked)

| # | Focus | Check performed (reviewer-run unless noted) | Result |
|---|-------|---------------------------------------------|--------|
| 1 | Rule change en→ko, docs-site chain untouched | Diffs of both SKILL.md files read: README chain row flipped (`README.ko.md` canonical → derive `README.md`/ja/zh), version bumps, chain-history notes correctly state the docs-site chain was ALREADY ko-canonical (no docs-site chain text altered). Runner `derivedLocalesFor('readme-only')` re-run → returns exactly the 3 ko-canonical derivation targets (verbatim JSON match with evidence). Residual en-canonical phrasing grep → only NEW ko-canonical statements (en named as derived) — clean. hns-* user-owned: no template mirrors exist (namespace is dev-only per the harness doctrine; diff touches none) | PASS |
| 2 | Absorption lossless (esp. TRUST5) | `absorption-map.md` read in full: 26 rows, every en-skeleton element has an explicit disposition (흡수/대응/중복 폐기/부분/요약 + two fact corrections). Epigraph absorption spot-verified in ja (`:36` 「」quote) and zh (`:36` ""quote). **TRUST5 judgment: intentional, documented decision — not hidden loss**: the map itself declares "요약 수준 흡수" as a Gap with a follow-up-card recommendation; the five-letter expansion + key numbers (85%·lint 0·type 0·Conventional) survive; the per-item verification-method detail architecturally belongs to docs-site, which carries TRUST-depth pages (verified existing: `docs-site/content/ko/_index.md`, `moai-loop.md`, `moai-gate.md`). Sound README-entry/docs-depth split | PASS |
| 3 | Re-derivation parity | Reviewer-measured on extract: H2 = 12 ×4, H3 = 59 ×4, fences = 42 ×4, lines = 756 ×4 — **exact**. Self-locale docs links = 20 in each; `adk.mo.ai.kr/ko/` body residuals in en/ja/zh = **0** (only the locale-table Korean row, as declared) | PASS |
| 4 | Fact correction 36→33 cells | `~/go/bin/moai model profile --json | grep -c '"agent":'` → **11** (reviewer-run) → 33 cells; "12 = catalog count incl. Explore" matches CLAUDE.md §4's retained-catalog statement. Correction propagated to all 4 READMEs (lane-verified, structure parity makes it 4-file by construction) | PASS |
| 5 | t59 "five sessions" subsection | Heading present at `:75` in ko (`### 다섯 세션이 쓰는 말`) and en (`### Words the five sessions share`); ja/zh carry the translated body (kanban paragraph verified translated) and the 59-H3 structural parity guarantees the heading slot. NOTE: the subsection's CONTENT still describes the OLD 6-column board — see 부기 B | PASS (content currency → 부기 B) |
| 6 | Docs discipline | URL blacklist (`docs.moai-ai.dev\|adk.moai.com\|adk.moai.kr`) → **0** across 4 files; Mermaid: 12 blocks (3×4) **all `flowchart TD`**; internal targets verified existing: CHANGELOG/LICENSE/CONTRIBUTING + 5 named pngs; locale infographics = **18** (6×3); language-switcher header contract: 3 outbound locale links in each file, self unlinked (`README.md:12-15` verified, count=3 ×4). Emoji scan: lane-attributed (positions enumerated: agent-table cost circles, code-block interiors, 💡 tip — consistent with reviewer spot-reads; prose emoji not observed in sampled sections) | PASS |
| 7 | Style unification | `grep 합니다\|습니다 README.ko.md` → **exactly 2 hits, both declared intentional**: `:115` quoted agent speech ("테스트가 통과했습니다" inside quotation marks) and `:755` credits line (만들었습니다). Prose is uniformly 해라체. Boundary handling appropriate | PASS |
| 8 | docs-site unchanged | 0 docs-site files in the 11-file diff (verified by name list); hugo-skip rationale valid — the card's surface is the GitHub README set; docs-site was already ko-canonical so chain coherence holds without site edits | PASS |

## 4-perspective

- **Functionality**: canonical flip is mechanically wired (Runner + skills agree; no residual en-canonical text); parity numbers are exact.
- **Security**: no executable surface; all external URLs on the adk.mo.ai.kr whitelist pattern; no secrets. OK.
- **Craft**: the absorption map is exemplary loss-accounting (dispositions + fact corrections + declared gaps); evidence honestly scopes what machines could not check (rendering, native-speaker quality).
- **Consistency**: 4-locale structural identity + switcher contract + blacklist + TD-only mermaid all reproduce.

## Reviewer-run verification (baseline attribution — committed tree at WT-t47, /tmp/t47-review)

```
$ grep -c '^## ' README{,.ko,.ja,.zh}.md      → 12/12/12/12
$ grep -c '^### ' …                            → 59 ×4
$ grep -c '^```' …                             → 42 ×4
$ wc -l …                                      → 756 ×4 (3,024 total)
$ self-locale link counts                      → 20 ×4
$ adk.mo.ai.kr/ko/ body residual (en/ja/zh)    → 0
$ URL blacklist grep                           → 0
$ mermaid direction                            → 12 × flowchart TD
$ ~/go/bin/moai model profile --json rows      → 11
$ node … derivedLocalesFor('readme-only')      → 3 ko-canonical targets (verbatim)
$ internal targets ls + infographic count      → 3 docs + 5 png + 18 infographics
$ switcher outbound links                      → 3 ×4
$ 합니다/습니다 in ko prose                       → 2 (both declared intentional)
```

## 부기 (lead-record items, per dispatch)

- **A — one-word nit**: `hns-oss-docs-i18n-rules/SKILL.md:34` says the ko skeleton is "11-section" — the shipped skeleton has **12 H2** (FAQ added, per the absorption map's own "골격 11→12" note). Fix "11-section" → "12-section" (or drop the number) with the next touch of that file; not worth a card alone.
- **B — deferred board-terminology debt is now 4-locale uniform**: the ko canonical itself still describes the OLD board (`:46` 다섯 개 terminals / `plan·run·review·sync`, `:60` `--name review-<run-id>`, `:66` "여섯 칸" chain, `:80` mermaid with `review`), faithfully re-derived into en/ja/zh (3 `plan…review…sync` matches each). This is the pre-existing debt t113's review recorded as Gap ② (README.ko named in t113's residual-risk) — **not a t47 defect** (t47's scope was the canonical flip; the debt predates it), but the re-derivation means the superseded board model now ships at release quality in 4 locales. Recommend scheduling the t113-Gap-② follow-up card promptly, with scope = README ×4 + docs-site kanban pages ×4 + t59 glossary (+ eventually the `kanban-five-sessions.png` image itself).
- **C — lead-record (pre-flagged by dispatch)**: ja/zh sentence-level native quality (GLM-authored) and external anchor/section-name breakage risk for sites citing the old en skeleton (14→12 H2) — release-notes mention advisable.

## Gaps

- Emoji-position scan and GitHub visual rendering (table breakage) lane-attributed/unchecked — machine checks cover structure and links; human eye remains for render.
- ja/zh heading text for the five-sessions subsection verified by structural parity + body spot-check, not by exact heading grep.
