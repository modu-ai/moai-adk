# t107 review verdict — report milestone→card cross-check gate (report.go + `--milestones-no-card`)

- Reviewer: review lane session (dispatched by lead-tjv7iy, lens: default 4-perspective + parser·CLI precision)
- Card: t107 · Worktree: `.claude/worktrees/t107` · Branch: `WT-t107` @ `c5eed2456` (base `6b44bdd2e` = origin/release/v3.1.1 at branch point)
- Integration note: origin/release/v3.1.1 has since advanced to `d169c4aec` (merge t60) — no file overlap with this delta (internal/graph, internal/cli/graph.go), conflict surface minimal at self-integration.
- Delta reviewed: `6b44bdd2e..c5eed2456` — 7 files, +659/−13, ALL under `internal/` (no template surface; mirror-parity/neutrality N/A — lane claim confirmed by reviewer `--stat`).
- Evidence read: card worktree `.moai/reports/t107/evidence.md` (105 lines, 5-section) + `unified-board-card-crosscheck.md` (paste-ready).
- Method note: card-worktree lock held by the idle run session (pid 53876, no activity since 15:29) AND worktree isolation blocks card-worktree writes — verdict therefore lives in the lead's release-v311 tree (same as t103/t110). Review performed from the release-v311 tree reading `WT-t107` refs/committed blobs; dynamic verification via `git archive c5eed2456` → `/tmp/t107-review` (tests the committed state exactly, immune to uncommitted drift; binary built to `/tmp/t107-rv-moai`; demo root `/tmp/t107demo2` = primary report copy + spliced section, real primary queue).

## Verdict: FAIL — 2 must-fix (both small; all other dimensions PASS-grade)

The card is functionally sound and the demo reproduces exactly; the failure is concentrated in the interpretation layer that the card itself exists to mechanize — the reissue lineage trap the lead pre-flagged (focus 6).

### Must-fix 1 (artifact — deliverable (1), paste-blocker)

The paste-ready table's S6 row is factually wrong and would request a duplicate card for already-delivered work:

> `| S6 | 공장장 세션 --chief | [신규 발급 필요] | 원 주장 t84 — 미발급 (git grep 0건) |` … summary "S6 신규 필요"

Reviewer-verified lineage: `git log --grep 'merge: t110'` → `6b44bdd2e merge: t110 — … review-PASS(허브)` (landed); `merge: t84` → **0**. Lead pre-confirmed the t84→t110 reissue covers S6. The lane followed exactly the caveat's prescribed rule (`git grep merge: tNN` → 0건 → "미발급") and reached the wrong conclusion — live proof that the rule is insufficient (see must-fix 2).

Fix: S6 card cell → `t110`; 실측 cell → reissue lineage (e.g. "재발행 t84→t110 — 머지 6b44bdd2e"); summary → 8 매핑·신규 0. **The lead must NOT paste the current block** — it is waiting for lead application and would propagate the error into the primary report. (Do not write "t84 → t110" in the card cell itself — the parser would claim both ids; lineage belongs in the 실측 column, which is not card-parsed.) Recommended: append a one-line correction note to evidence.md's t84 finding rather than rewriting it.

### Must-fix 2 (code — one caveat sentence + test assertion)

The reissue limitation is stated nowhere: `git grep -i 'reissue|재발행'` over the committed tree returns only an unrelated pre-existing regex (`reIssuesSummary` in internal/hook/quality/linter.go). `milestoneNoCardCaveat`'s decision rule — "(완결이면 통과, 미발급이면 새 카드)" — yields a false "new card" verdict for any id that was reissued and landed under a new id. Three such lineages are live on this board (t84→t110, t105→t106, t38→t111), so the false-positive class is not hypothetical.

Fix: add one sentence to `milestoneNoCardCaveat` (internal/cli/graph.go) — zero `merge: tNN` matches does NOT prove the work is undone; the card may have been reissued under a new id (board lineage / merge-title sweep before issuing a new card). Optionally mirror in the `query` Long help; extend the existing caveat assertion in `TestGraphQueryCmd_MiltestonesNoCard`. ~1 const string + 1 test line.

## Dispatch focus items (6/6 checked)

| # | Focus | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | report.go parser precision | Full read + 6 tests + live demo. Section = `## Card Cross-Check` prefix (locale gloss ok); FIRST table in section only (the report's own §7 table correctly skipped — the test discriminates because S0 exists only in the cross-check table); card column located by exact header cell (`card`/`카드`), never by position; tNN word-bounded, read ONLY from that cell — demo proves the 실측 column's "원 주조 t84" is NOT claimed (S6 = no card). Ragged rows / interior empty cells handled; per-row and per-edge dedup; deterministic emission (sorted names, Build-applied EdgeLess) | PASS |
| 2 | `--milestones-no-card` queue parity + fail-open | Source-verified: `liveQueueCards` → `todoBacklogPath(resolveTodoQueueRoot())` — the identical chain `moai todo` uses (git common-dir parent = primary root, t106 interpretation). Queue-unreadable → explicit "card-vs-queue comparison skipped" NOTE; hermetic CLI test corrupts the queue and asserts claimed milestones are NOT flagged while no-card rows still are | PASS |
| 3 | t100 interface untouched | `query.go` delta is pure append (`MilestoneClaim`/`MilestoneClaims`); FindCallers/BlastRadius bodies unmodified; `--blast` verified live | PASS |
| 4 | Live demo reproduction | Reviewer-rebuilt binary (from c5eed2456), demo root = primary report copy + spliced section, real primary queue: build `16 edges — report-milestone: 8, milestone-card: 8` exact; query `milestones without a live card: 6 of 8` exact; `--blast t59` → 2 nodes (report + S7) exact. Drift: S7 line now reads `t58,t59` because t59 merged (`d0f946d62`) after the lane's 15–16h snapshot — the variable-state drift the evidence itself warns about; queue at review time has live = t108/t113 only | PASS (drift explained) |
| 5 | Template surface untouched | All 7 delta files under `internal/`; no `.claude/`·`.moai/` template path in the diff → make build / mirror-parity / neutrality not required | PASS |
| 6 | "6 of 8" reissue-trap interpretation | Tool cannot distinguish stale reference vs real gap AND does not say so (grep = 0); the artifact itself mis-called S6 (t84→t110 landed, lead-confirmed, reviewer-verified) | **FAIL** → must-fixes 1+2 |

## 4-perspective (default lens)

- **Functionality**: core mechanics correct (demo byte-level reproduction modulo documented queue drift); defect confined to the interpretation layer (caveat rule + one table row).
- **Security**: read-only layer over `.moai/reports` + queue file; no writes, no injection surface beyond bounded regex parsing. OK.
- **Craft**: strong — fail-open consistent with sibling layers, stem-qualified milestone nodes (collision test), seam injection mirrors the house pattern (`userHomeDirFn`), `TestGraphCmd_NoAskUserQuestion` guard, output-determinism contract preserved, tests hermetic (real primary queue never read).
- **Consistency**: kind naming / help text / build counts follow the existing graph CLI conventions; evidence file is honest (declared the queue-mutability gap — vindicated within hours by the t59 drift).

## Reviewer-run verification (baseline attribution — this run, committed tree at c5eed2456)

```
$ git archive c5eed2456 → /tmp/t107-review (committed state exactly)
$ go build -o /tmp/t107-rv-moai ./cmd/moai        → BUILD-OK
$ go vet ./internal/graph/ ./internal/cli/        → rc=0 (VET-OK)
$ go test ./internal/graph/ -count=1              → ok 3.022s
$ go test ./internal/cli/ -run 'TestGraph' -count=1 → ok 1.276s
$ golangci-lint run internal/graph/... internal/cli/ → 0 issues (rc=0)
$ /tmp/t107-rv-moai graph build --root /tmp/t107demo2     → 16 edges (8 + 8) [verbatim in §focus 4]
$ /tmp/t107-rv-moai graph query … --milestones-no-card    → "milestones without a live card: 6 of 8"
$ /tmp/t107-rv-moai graph query … --blast t59             → "blast radius of t59: 2" (report + S7)
```

## Gaps

- Uncommitted drift in the card worktree unobservable (lock held by idle run session) — verdict binds the committed state `c5eed2456`, which is the surface integration merges.
- Full `internal/cli` suite (lane: 291s) and darwin/windows cross-compile gates deferred to CI (lane-local discipline).
- Queue snapshot is time-of-review (20:08 KST); flags are query-time-valid only.

## Residual risks

- Reports deviating from the exact header/column names vanish silently (the aggregate "no Card Cross-Check sections found" note fires only when zero sections parse overall) — inherent to the fail-open layer; detection coverage rides discipline-(2) adoption, as the evidence itself states.
- After must-fix 1, S6/t110 will still print as "not in live queue" (t110 done, row removed) — correctly resolved by `git grep merge: t110` → found → 통과; the post-fix caveat will explain the reissue class.

## Re-review protocol (lightweight)

On fix: (a) read the caveat-const diff + S6 row + test assertion, (b) re-run the two targeted test commands, (c) one demo query for output shape. No full re-review needed.

---

# Re-review @ `8460ca593` (2026-08-17, dispatched after t113 PASS)

- Range: `c5eed2456` (reviewed) → `411158492` (release-tip sync, 0 conflicts) → `8460ca593` (discipline (2) reflection: stub [HARD] section + 5 offsetting trims + evidence/paste-file committed).
- Original-M1 disposition: **resolved OUTSIDE this branch** — the lead routed it to t113 rider 2, which applied the corrected table to primary (verified PASS in the t113 verdict, incl. machine verification: S6 parses as single `t110` claim, 0 real gaps).
- Original-M2 disposition: **still open on NO branch** (`git show 8460ca593:internal/cli/graph.go | grep 재발행|reissued` → 0).

## Re-review verdict: FAIL maintained — must-fix updated (M1′ replaces M1; M2 unchanged)

The new work on this branch (release sync + discipline + budget offsets) is itself clean — verified below. The FAIL stands on the two items the recipe named, one changed shape:

### M1′ (new — the committed paste-ready file is the WRONG version)

`8460ca593` commits `.moai/reports/t107/unified-board-card-crosscheck.md` **unchanged from the state my FAIL rejected**: S6 row still `[신규 발급 필요] · 원 주장 t84 — 미발급 (git grep 0건)`, summary still "S6 신규 필요", and the header notes still claim `live: t108·t113·t59` (stale — t59 merged, d0f946d62). Committing it makes the known-wrong artifact a paste source that REGRESSES the corrected primary table if anyone uses it. Fix (minutes): update the block to match the applied primary section (S6 card cell `t110` alone, lineage in the 실측 column, summary 8/8·신규 0, queue note refreshed) — or replace the body with a superseded tombstone pointing at the primary section. Recommended also: one correction line next to evidence.md's "t84 = 0건 — 갭 잔존분" finding (now known wrong: t84→t110 reissue landed).

### M2 (unchanged — CLI caveat sentence + test assertion)

Still absent. Note: the doctrine layer (t113 companion rule) and the report-level caveat now exist, but the string the USER sees in `--milestones-no-card` output still prescribes only `git grep merge: tNN` and mis-guides on reissued ids. 1 const sentence + 1 test line in `internal/cli/graph.go` / `graph_cmd_test.go`.

## Recipe items ③④ (done, PASS-grade)

- **③ Tests + demo (reviewer-run on `git archive 8460ca593`)**: `go test ./internal/graph/` ok 1.037s · `go test ./internal/cli/ -run TestGraph` ok 2.350s · `TestAlwaysLoadedTokenBudget` PASS · mirrors byte-identical (both files). Graph code is **byte-identical to the reviewed `c5eed2456`** (`git diff c5eed2456..8460ca593 -- internal/graph/ internal/cli/graph.go internal/cli/graph_cmd_test.go` → empty), so the full original verification binds unchanged. Budget: stub 23,544 → **23,542 bytes (−2B net)** — the lane's probe measurement (75,794 → 75,793 tokens, headroom 207) is consistent and honest (one-off probe deleted after measuring; methodology stated). Demo query 1회: run — and it incidentally exercised the **home-fallback queue path** (cwd outside a git repo → `resolveTodoQueueRoot` falls back → different queue → "8 of 8"): designed behavior, honestly reflected in flags; canonical primary-queue demo (6 of 8, all merges) stands from the t113 review. Advisory (not must-fix): a `--root`-pointed query from a non-repo cwd mixes graph-from-root with queue-from-home — worth one help-text line someday.
- **④ Duplicate [HARD] section (lead's dispatch error, acknowledged by lead)**: t107's version confirmed on this branch AND t113's on WT-t113. Neutrality of t107's version: no SPEC-ids/dates/SHAs; contains the `(`tNN`)` format literal (internal id-format hint) vs t113's fully neutral "the delivering card id"; mechanics inlined in the stub vs t113's companion placement; **no companion pairing** on t107's side. **Recommendation: t113's version survives; t107's section is dropped at integration — but t107's 5 TRIMS ARE PRESERVED** (verdict's-home relocation to companion + CodeRabbit/recipe sentence trims — they produced the −2B net and are good edits independent of the duplicate). Integration order: t113 merges first (PASS final) → t107's integration merge resolves the stub conflict as "t113's section + t107's trims".

## Close-out condition

FAIL → PASS when M1′ (file corrected or tombstoned) and M2 (caveat + test) land on this branch; verification is then two greps + one targeted test run (recipe unchanged).

---

# Close-out @ `1010546e6` (2026-08-17, bundle re-review) — **verdict: PASS**

Fix commit `1010546e6` ("re-issuance caveat + canonical crosscheck copy") verified against the recipe:

- **M1′ ✓** — `.moai/reports/t107/unified-board-card-crosscheck.md` is now a declared **archive copy**: header states 정본 = primary § Card Cross-Check and that this file is NOT a paste source; an explicit 재심 M1′ 정정 note records the S6 mis-call (t84→t110 reissue, merge 6b44bdd2e, 8/8 mapping, "grep 0건 ≠ 재발행 아님"). The copied block is **byte-identical to the primary section** (reviewer diff: 14/14 lines, 0 differences). evidence.md's "t84 = 0건" line now carries the inline correction annotation (append-not-rewrite, as recommended).
- **M2 ✓** — `milestoneNoCardCaveat` gains the third line: "grep 0건 ≠ 작업 미완 — 카드가 재발행되었을 수 있으니 새 카드 발급 전에 계보를 확인." (verbatim, committed). `graph_cmd_test.go` +1 assertion: `"새 카드 발급 전에 계보를 확인"` in the `TestGraphQueryCmd_MilestonesNoCard` want-list, commented "zero-hit grep != work never done; check re-issuance lineage (re-review M2)".
- **Parser no-pollution claim ✓** — `report.go:39-42` comment documents the top-level-`*.md`-only scan (subdirectories hold working notes); code behavior was verified in the original review.
- **Targeted tests ✓ (reviewer-run on `git archive 1010546e6`)**: `TestGraphQueryCmd_MilestonesNoCard` **PASS** · `TestGraphQueryCmd_MilestonesNoCardQueueUnreadableSkipsComparison` **PASS** (`ok internal/cli 0.843s`).

**Verdict: FAIL → PASS.** Integration note (per lead's plan): the branch still carries its own duplicate discipline section (from `8460ca593`) alongside t113's landed version — resolve the stub at the release merge keeping **t113's section + t107's 5 trims** (see the re-review ④ recommendation above).
