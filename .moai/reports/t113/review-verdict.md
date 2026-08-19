# t113 review verdict — review column removed: five-column board, sync gate absorbs the review verdict (+ t107 riders 1·2)

- Reviewer: review lane session (dispatched by lead-tjv7iy, lens: default 4-perspective + closed-set·doc-restructure precision)
- Card: t113 · Worktree: `.claude/worktrees/t113` · Branch: `WT-t113` @ `327c6453d` (base `d169c4aec` = release tip incl. merge t60)
- Delta reviewed: `d169c4aec..327c6453d` — 11 files, +168/−85 (code: kanban×5 + web×1; docs: dispatch stub+companion local+template; evidence committed)
- Evidence read: `git show 327c6453d:.moai/reports/t113/evidence.md` (VCI 5-section) — claims reproduced below
- Method: card-worktree writes blocked by session isolation → verdict in release-v311 tree (lead pre-authorized). Dynamic checks via `git archive 327c6453d` → `/tmp/t113-review` (committed state exactly).

## Verdict: PASS

## Dispatch focus items (7/7 verified)

| # | Focus | Check performed (reviewer-run unless noted) | Result |
|---|-------|---------------------------------------------|--------|
| 1 | Column closed set | Diff read: `column.go` 5 constants (`backlog→plan→run→sync→done`), `allColumns` closed/unexported, `ParseColumn` rejects `review` (falls to "not one of the five columns"); `reconcile.go` in-progress pair → `run‖sync` + comments; `viewmodel_ops.go` `ChainRoles` = 4 roles; 3 tests updated exactly as claimed (5-value enumeration + `review`/`reviewed` added to reject list + HasOwningSession 3 working + round-trip fixture + collision pair). `grep -rn ColumnReview internal/` on extract → **0** | PASS |
| 2 | dispatch 6→3 restructure = RELOCATION, not deletion | Stub + companion diffs read in full. Board table: `review` row removed, `sync` row now "Review verdict (lenses per card), docs, CHANGELOG, PR"; lens field "`lens` appears only in a `sync` dispatch"; CodeRabbit section retained with "does not leave `sync`"; "Review lens selection" section retained, reframed "the sync gate runs that review" with the lens table intact; card classes A (`plan` skipped) / B (`run → sync`) / C ("all three working columns") re-passed consistently in both files. `[HARD]` count: stub 19→**20**, companion 2→**3** (no loss, +1 each = rider 1). Residual-phrase grep (`review column`·`six column`·`four columns`·`review-a1b2c3`) on both files → **0**. Stub+companion split discipline preserved ([HARD]/prohibitions in stub; tables/rationale in companion) | PASS |
| 3 | Mirror parity + build | `diff -q` in extract: kanban-dispatch.md + kanban-dispatch-detail.md local↔template → **identical**. `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Mirror|Parity|Leak'` → ok 1.374s. `make build` in extract → rc=0; catalog.yaml committed hash `7a87919718716dca` == regenerated `7a87919718716dca` (determinism reproduced) | PASS |
| 4 | Budget | `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget$` → **PASS at base AND tip** (reviewer-run both extracts; guard prints no number on success). Stub (always-loaded) bytes 23,544 → 24,235 = **+691B ≈ +173 tokens** — rider 1's [HARD] section, consumed from the t114 headroom (703); under-budget so no compensating-trim obligation triggers. Companion is `paths:`-gated (lazy) → no budget impact | PASS (net +173t from headroom) |
| 5 | i18n no-op | `grep -ri review internal/web --include='*.templ'` → only `workflow.codex.review_gate.enabled` (codex config, not the board); i18n.js hits are file-header comments + plan-preview keys; `"review"` literals in internal/ are codex-protocol/test surfaces. `ChainRoles` consumers: definition + single renderer (`viewmodel_ops.go:235`) — sole view-A source as claimed; view B (`pipelineColumns`) never had review (comment-only diff) | PASS |
| 6 | Rider 1 (stub [HARD] rule) | New "## Report milestones ↔ queue cards" [HARD] in stub — card-id NEUTRAL ("the delivering card id or an explicit new-card marker", no tNN) ✓ template-neutrality safe; companion same-named section carries rationale + `moai graph build && moai graph query --milestones-no-card` + queue semantics (queued/picked; dropped does not qualify) + `<card-id>`-neutral git resolution. Stub/companion pairing exactly per the rule/procedure split | PASS |
| 7 | Rider 2 (primary §7 insertion, outside branch diff) | Read primary `unified-board-design-20260817.md`: S6 card cell `t110` **alone** (no `tNN→tNN` compound — parser trap avoided); 실측 cell "재발행 t84→t110 — 머지 6b44bdd2e"; summary "8개 → 8개, 전 매핑 완결 (신규 0)"; lineage note covers all 3 (t84→t110 · t105→t106 · t38→t111); caveat states BOTH the parser limit (card칸 화살표 금지) and the done/never-issued ambiguity. Bonus: S7 실측 reflects t59's merge (d0f946d62) — current-state accurate | PASS |

### Gap ① — partially CLOSED by reviewer (machine verification of the live table)

Ran the t107-built binary (`/tmp/t107-rv-moai`, graph code byte-identical to WT-t107) against a copy of the CURRENT primary report:

```
17 edges — report-milestone: 8, milestone-card: 9
S6  claimed t110 — not in live queue: t110   ← corrected cell parses as the single claim
milestones without a live card: 6 of 8
```

All 6 flags (S0 t109 / S3 t56 / S4 t55 / S5 t85 / S6 t110 / S7 t58+t59) resolve via `git log --grep 'merge:'` to **verified merges** (2c70e7aed / 400dde787 / 1ea829c76 / 162f74d99 / 6b44bdd2e / b8a25b62f + d0f946d62 — each reviewer-checked); S1 (t108)·S2 (t113) live in queue → unflagged. **Zero real gaps.** Remaining: the DISTRIBUTED binary gains `graph` only after t107 merges — rerun on the installed binary then (one command).

## Reviewer-run verification (baseline attribution — committed tree at 327c6453d, /tmp/t113-review)

```
$ go vet ./internal/kanban/ ./internal/web/                → rc=0
$ go test ./internal/kanban/ -count=1                      → ok 14.794s
$ go test ./internal/web/ -count=1 -timeout 300s           → ok 2.933s
$ go test ./internal/config/ -run TestAlwaysLoadedTokenBudget$ -count=1 → PASS (also PASS at base d169c4aec)
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Mirror|Parity|Leak' → ok 1.374s
$ golangci-lint run internal/kanban/... internal/web/...   → 0 issues
$ make build (extract)                                     → rc=0; catalog.yaml hash identical pre/post
```

## 4-perspective

- **Functionality**: closed set enforced at the only constructor; reconcile/admission/web all consistent; backward-compat reasoning (LoadBoard does not validate column values; stale `review` values surface as reconcile-inconsistent, not corruption) sound and code-verified by reviewer.
- **Security**: no new input surface; set stays unexported and closed. OK.
- **Craft**: minimal diffs, comments updated everywhere the old count lived ("no sixth value"), test renames honest (Six→Five), reject list GREW (`review`, `reviewed`) — good defensive testing.
- **Consistency**: stub↔companion↔code↔template all state the same 5/3 structure; [HARD] count net-increased; no stale phrasing survives (grep 0).

## Gaps (per dispatch, recorded)

- **① t107-merge follow-up**: machine re-verification on the INSTALLED primary binary after t107 lands (`moai graph build && moai graph query --milestones-no-card`) — pre-merge evidence above already shows the table correct. **Context**: t107's M2 (CLI caveat reissue sentence) is NOT in t113; it is the subject of the t107 re-review dispatched in parallel — see the t107 verdict file's re-review section.
- **② docs 6-column notation debt** (this card deliberately out of scope): docs-site kanban pages (ko/en/ja/zh), README.ko "여섯 칸", t59 glossary 4-locale, kanban-mode.md "네 단계 (plan → run → review → sync)" — recommend a follow-up card (4-locale same-PR obligation surface).

## Residual risks

- Stale `review` values in existing board state files will keep rendering as inconsistent until the card moves — operational note recommended (lane already flags this).
- Lens-table authority now lives in the sync-gate frame; a future split of "sync" into more gates would need to re-home the lens table (noted, not actionable now).
