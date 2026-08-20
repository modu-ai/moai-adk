# t108 review verdict — orchestration mode catalog renamed 6→4 onto the spawn-count axis

- Reviewer: review lane session (dispatched by lead-tjv7iy, lens: default 4-perspective + exhaustive-rename·backward-compat precision)
- Card: t108 · Worktree: `.claude/worktrees/t108` · Branch: `WT-t108` @ `cdc3f0a6a` (body `ebf95526b` 81 files +641/−502 on base `d169c4aec`; tip sync `7a969dac8` brings the t113 merge; evidence `cdc3f0a6a`)
- Delta reviewed: `d169c4aec..cdc3f0a6a`. Note: internal/kanban+web changes inside this range are **t113's merge**, not t108's — t108's own Go footprint is `catalog.yaml` (data) only, matching the "Go 코드 0행" claim.
- Evidence read: `git show cdc3f0a6a:.moai/reports/t108/verify-summary.md` (VCI 5-section).
- Method: card-worktree writes blocked by session isolation → verdict in release-v311 tree. Dynamic checks via `git archive cdc3f0a6a` → `/tmp/t108-review`.

## Verdict: FAIL — 1 must-fix (2-line template-mirror sync; the card's central [HARD] claim is false as committed) + 2 advisory riders

The rename itself is excellent — 81 files consistently re-axised, honest compat design, docs fact-preserving, budget green. The failure is one mirror file the rename never touched, on the surface that ships to every user, plus the measurement artifact that hid it.

### Must-fix — `internal/template/templates/CLAUDE.md` still teaches deleted mode numbers

Reviewer grep over the claimed scope (`.claude/` + template tree + root `CLAUDE.md` + docs-site) finds **2 lines / 3 occurrences**, all in the template mirror `internal/template/templates/CLAUDE.md`:

- `:145` — "MoAI's own **Mode 4** band is 3-5 as an advisory … the team-size 3-5 advisory binds **Mode 3** teammates only"
- `:151` — "Genealogy: **Mode 3** (`--team`) was retired …"

The root `CLAUDE.md` WAS renamed ("fanout band … binds agent-team teammates"; "Genealogy: agent-team (`--team`)") — the mirror was missed (`git diff --name-only` confirms `internal/template/templates/CLAUDE.md` is NOT in the change set). Consequences:

1. The card's Claim 1 ("살아있는 전 표면 grep `Mode [1-6]|Mode 7` = **0건** [HARD]") is **false as committed** — the distributed template (`moai init`/`moai update` output) ships a CLAUDE.md referencing a catalog the same distribution no longer defines.
2. **CI will not catch it**: `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/` passes (31.709s, reviewer-run) — the parity suite does not compare root CLAUDE.md against its mirror. The fix must be manual and the grep evidence corrected.
3. Root cause of the false 0: the lane's evidence command was `grep -rn "Mode [1-6]\|Mode 7" .claude/ templates/ docs-site/ CLAUDE.md` — `templates/` does not exist at the repo root (real path `internal/template/templates/`); the missing directory was silently skipped, so the measurement never scanned the template tree. Same failure class as the CLI-existence lesson (measurement scope silently narrower than claimed).

Fix: sync the 2 mirror lines to the root wording; re-run the grep with the correct path; correct the evidence's grep command line (and its 0-row artifacts).

### Advisory riders (lead's call — both trivial)

- **R1** — stale axis name in shipped Go comments: `internal/config/model_routing.go:7` and `internal/config/types.go:451` both say "Phase 0.95's **Mode 1-6** shape axis" (orthogonality notes). Outside the card's declared surface, but live maintainer-facing comments naming a deleted axis — 2 small edits.
- **R2** — SSOT completeness: the three handoff-rendering surfaces all cross-reference "the `orchestration-mode-selection.md` rename note" for the legacy-token mapping, but the note maps old CATALOG names (`trivial`→`direct`, `sub-agent`→`serial`, `parallel`→`fanout`, `workflow`→`sweep`) and never lists the handoff ENUM tokens (`solo-sequential`/`parallel-subagents`/`dynamic-workflow`). The semantics agree (sub-agent↔solo-sequential etc.) and each rendering surface carries the full mapping inline, so no paste-time ambiguity exists today — but one added row would make the cross-reference resolve literally.

## Dispatch focus items (7/7 checked)

| # | Focus | Check performed (reviewer-run unless noted) | Result |
|---|-------|---------------------------------------------|--------|
| 1 | `grep 'Mode [1-6]\|Mode 7'` = 0 | Ran over extract `.claude/` + `internal/template/templates/` + root `CLAUDE.md` + `docs-site/` → **2 lines, all in templates/CLAUDE.md** (must-fix). Exceptions measured: `.moai/specs` = 766 lines (claim "~800" ✓), `CHANGELOG.md` = 2 entries (claim ✓) — immutable-record carve-out accurate | **FAIL** → must-fix |
| 2 | Backward-compat + mapping consistency | Old enum tokens appear at exactly 8 lines = 4 files × 2 (moai.md §8, session-handoff.md Block 1, session-handoff-examples.md ×2, + template mirrors) — all in compat-mapping context, zero live usage. The three RENDERING surfaces carry byte-identical semantics (`solo-sequential`/`parallel-subagents`/`dynamic-workflow` → `serial`/`fanout`/`sweep`, same order, "parse-accepted on read and map — never emitted"). SSOT note is the catalog-name axis, semantically consistent (see R2) | PASS |
| 3 | execution-rules.md header rename | Diff read: `#### Mode 1/2/3: Manual/Personal/Team` → `#### Git strategy: Manual/Personal/Team` ×3 — the judgment basis is confirmed by the surrounding content (git config JSON, use cases = git strategies, not orchestration modes); no referencing text survives (global grep 0) | PASS |
| 4 | Budget | `TestAlwaysLoadedTokenBudget` → **PASS** (reviewer-run, 0.451s). Headroom: guard is silent on success so no exact number; PASS ⇒ surface ≤ 76,000, consistent with the ≥200 claim given t113-adjusted prior state (≈530 pre-card). Lane's "+2줄" increment not independently re-derived (my superset line-count is net +18 over a broader file set — includes lazy companions); the PASS gate is the authoritative signal | PASS |
| 5 | Template | STRICT suite ok 31.709s + `make build` rc=0 (reviewer-run). Mirrors of all 36 template files in the diff updated with their local twins — EXCEPT the CLAUDE.md root↔mirror pair (must-fix). catalog.yaml updated in-diff (4 lines, file-list data) | PASS **except CLAUDE.md** |
| 6 | docs-site | 19 files = 4 pages × 4 locales + `moai-run.md` en/ko only — scope correct: ja/zh moai-run pages contain no `--mode`/catalog paragraph at all (nothing stale to update; verified by direct grep). Fact preservation spot-checked (ko `moai.md`: 6-모드 catalog sentence → 4-모드 with agent-team footnote framing; gates list unchanged). `hugo --quiet` → **rc=0, zero warnings** (reviewer-run) | PASS |
| 7 | agent-common-protocol untouched | `git diff d169c4aec..cdc3f0a6a -- …agent-common-protocol.md` → **0 lines**; numeric mode references 0 (covered by the focus-1 grep over `.claude/` = 0 outside the mirror) | PASS |

## 4-perspective

- **Functionality**: doctrine-only change, no runtime surface; handoff enum read-compat is honest (no Go parser exists — verified claim chain: CLI stores body verbatim).
- **Security**: no new surface. OK.
- **Craft**: the rename script (`rename-modes.pl`) as method evidence, immutable-record carve-out declared with measured counts, evidence honestly lists runtime-unverified gaps.
- **Consistency**: one mirror file breaks the otherwise-complete surface sweep; the false grep-0 is a measurement-scope artifact, not fabrication (the artifacts genuinely contain 0 rows — for a scope that silently omitted the template tree).

## Reviewer-run verification (baseline attribution — committed tree at cdc3f0a6a, /tmp/t108-review)

```
$ grep -rn 'Mode [1-6]\|Mode 7' .claude/ internal/template/templates/ CLAUDE.md docs-site/ → 2 (templates/CLAUDE.md:145,151)
$ grep -rn 'solo-sequential' .claude/ internal/template/templates/ CLAUDE.md            → 8 (4 compat surfaces × 2, incl. mirrors)
$ go test ./internal/config/ -run TestAlwaysLoadedTokenBudget$ -count=1                   → PASS
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1                       → ok 31.709s
$ make build (extract)                                                                    → rc=0
$ hugo --quiet -d /tmp/t108-hugo-out (docs-site)                                          → rc=0, no warnings
$ git diff …agent-common-protocol.md                                                      → 0 lines
$ grep exceptions: .moai/specs 766 lines · CHANGELOG 2 (matches declared carve-out)
```

## Gaps

- Exact budget number unobservable (guard silent on success) — PASS gate + arithmetic context only.
- Full cli/kanban suites not re-run (t113's merge content already verified in the t113 review; CI re-runs everything).
- Runtime progress.md token emission unverified (doctrine-only change; lane-declared).

## Residual risks

- Pasted legacy `mode:` bodies from old memory resolve via read-side mapping — indefinite by design (no sunset); a future enum-tightening would need a deprecation note.
- `fanout`/`sweep` token similarity to unrelated identifiers (FO-* fanout ids) — noted by the lane; catalog SSOT keeps single source.

## Close-out condition

FAIL → PASS when the template CLAUDE.md mirror is synced (2 lines) and the evidence's grep command/path corrected; re-verification is one grep + one diff. Riders R1/R2 may ride the same commit at the lead's discretion.

---

# Close-out @ `ab9cf2957` (2026-08-17, bundle re-review) — **verdict: PASS**

Fix commit `d692adabe` (+ evidence refresh `ab9cf2957`) verified against the recipe:

- **MUST-FIX ✓** — template mirror synced and committed: `git diff ab9cf2957:CLAUDE.md ab9cf2957:internal/template/templates/CLAUDE.md` → **0 lines** (root↔mirror byte-identical; the `:145`/`:151` lines now carry the renamed wording).
- **Committed-tree grep ✓ (methodology corrected as dispatched)** — `git grep 'Mode [1-6]\|Mode 7' ab9cf2957 -- .claude internal/template/templates CLAUDE.md docs-site` → **0 hits** across all live surfaces, measured against the commit object (the lane's evidence now records the commit-tree method; working-tree artifact files updated accordingly).
- **R1 ✓** — `internal/config/model_routing.go:7` and `types.go:451` now read "ORTHOGONAL to the Phase 4 mode-shape axis (**direct / serial / fanout / sweep**)" — the stale "Mode 1-6" axis name is gone from shipped comments.
- **R2 ✓** — the SSOT rename note now carries the handoff enum-token mapping (verified: `solo-sequential` now matches `orchestration-mode-selection.md:14` **and its template mirror** — previously absent from both), so the three handoff surfaces' cross-reference resolves literally.
- **Extended immutable-exception group ✓ (appropriate)** — the 8 hits in `.moai/docs/autonomous-workflow-strategy.md` are dated proposal/decision records ("Mode 6: workflow 추가 제안", "Phase 0.95 5-mode" time-point, AP-9 finding reference, SPEC proposal notes) — chronicle-class history in a local reference doc, same category as `.moai/specs`/CHANGELOG. Rewriting them would falsify the record; the classification stands.
- Budget note: the fix's always-loaded delta is the +1-line enum row in `orchestration-mode-selection.md` — negligible against the ≈500-token headroom; lane re-ran the budget gate (green-budget.txt updated in-commit).

**Verdict: FAIL → PASS.** Integration proceeds per the lead's order (t107 first, then t108 — no file overlap between the two fix sets and the release branch at ca8c0b593+).
