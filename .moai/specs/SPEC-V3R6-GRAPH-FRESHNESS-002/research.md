# Research — SPEC-V3R6-GRAPH-FRESHNESS-002

## §1. Provenance and triage method

Card t279 is the follow-up to t250 (#1648, squash `6786c3fa4`, SPEC-V3R6-GRAPH-FRESHNESS-001). The CodeRabbit round-2 review dump (`.moai/reports/t250/cr-round2-comments.md`) produced **69 findings**. A just-completed triage verified all 69 against the current tree `c9eed8ac6` (branch `WT-t250-followup`; `6786c3fa4` confirmed ancestor):

- Method: full sweep of the 69 findings against the current tree, split across three read-only investigations with file:line evidence per finding — `verify-graph.md` (20 findings, internal/graph slice), `verify-cli.md` (32 findings, internal/cli + mcp + mx + hook + config-testdata slice), `verify-docs.md` (17 findings, docs/SPEC/workflow slice).
- The verify reports also resolved two sub-questions the raw dump could not: the Minor-2 candidates (the `mcp_code_tools_test.go:80` unguarded index — confirmed panic-risk; and the `codequery.go` residual quickwins), and the dump's stale "✅ Addressed" markers (re-verified against the tree — #6 holds; #2 and #4 do NOT; the regen/narrowing commits never touched the cited lines).
- The card's premise ("연기 31 + Minor 2 = 33 unaddressed") matched the measured still-valid count exactly (33) — card-premise verification per the standing investigation discipline.

## §2. The 69 → 29 reconciliation

| Verdict | Count | Section in triage-table.md |
|---|---|---|
| Adopted (this SPEC) | 29 | A-1 (11) · A-2 (10 + absorbed Minor-2b) · A-3 (3) · A-4 (5) |
| Follow-up candidates | 5 | B (F1 workflow SHA pinning · F2 ctx propagation · F3 fingerprint anchoring · F4 scanned-package evidence · F5 squash-merge × stamp reachability countermeasures — §7 below) |
| Rejected (invalid premise) | 2 | C (R1 doc totals match registration 28/24 exactly — CR miscount; R2 absolute tree_root is REQ-GF-003-mandated tracked-artifact design, documented at check.go:153-161 citing the comment-id) |
| Deferred-by-design | 1 | D (D1 unconfigured-skip silence — recorded deviation, AC-GF-006) |
| Already fixed | 33 | E (fix locations cited per finding for thread-close use) |

Counted allocations inside the adopted 29: A-1 #1–#10 execute in M1; A-1 #11 (astx CGO tags) executes in M3 with the astx group (triage's own grouping note — one package, one milestone); Minor-2a is absorbed into A-1 #6 (the `:80` site is the same defect family as 3855001928); Minor-2b is absorbed into A-2 #22b.

## §3. SPEC ID selection and uniqueness

- Candidate: `SPEC-V3R6-GRAPH-FRESHNESS-002` (second SPEC in the GRAPH-FRESHNESS domain).
- Regex pre-write self-check (executed, verbatim): `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`.
- Uniqueness: `ls .moai/specs/ | grep -i GRAPH` in THIS worktree → only `SPEC-V3R6-GRAPH-FRESHNESS-001` (plus the coincidental substring match `SPEC-SKILLPORT-SVG-INFOGRAPHIC-001`, a different domain); the primary checkout `.moai/specs/` shows the same single FRESHNESS entry. No collision; the fallback ID was not needed.
- Frontmatter: 12 canonical fields (schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`) + `era: V3R6` + `tier: M` + `related_specs`. The explicit `era:` is deliberate: at plan-phase the progress.md skeleton carries §E.2–§E.4 headings but no `sync_commit_sha` yet, which would heuristically classify as V3R5 (H-3) — H-override pins V3R6.

## §4. Close-path analysis (predecessor SPEC-V3R6-GRAPH-FRESHNESS-001) — code-verified

Two sources, both read this run: `moai spec close --help` and the implementation at `internal/spec/closer.go`. The implementation supersedes two assumptions the help text's shorthand planted.

**Precondition semantics (`validatePreconditions`, closer.go:694-739).** Four preconditions: (1) §E.2 sync section present; (2) §E.5 Mx section present (legacy) **OR** the 3-phase predicate — §E.4 marker + non-empty `sync_commit_sha` (closer.go:702-714; the code comment cites the schema's §E.5 retirement explicitly — a 3-phase SPEC honors the schema by omitting §E.5 and must still close); (3) all AC PASS — determined ONLY from acceptance.md FAIL markers (`**FAIL**` / `| FAIL |` / `| FAILED |`, closer.go:635-637); (4) no genuine PASS-WITH-DEBT — `hasGenuinePassWithDebtVerdict` (closer.go:667-687) scans **acceptance.md only**, anchored to table-cell-start or bold verdict forms, deliberately ignoring prose mentions such as "no PASS-WITH-DEBT". Full-close mode additionally requires spec.md status `implemented` (or `completed`).

**Predecessor state vs the code (verified this run):**

| Precondition | Evidence | Verdict |
|---|---|---|
| §E.2 present | progress.md carries §E.2 | pass |
| §E.5 or 3-phase predicate | §E.4 present; `sync_commit_sha: pending-backfill` → satisfied once the SHA is backfilled (M4 step 2) | pass after backfill |
| AC all pass | grep of predecessor acceptance.md → 0 FAIL markers | pass |
| No genuine PASS-WITH-DEBT | grep → 0 table-cell/bold debt markers; the debt lives in progress.md §E.3 `ac_pass_with_debt_count: 2`, which the check never reads | pass |
| status | `implemented` | pass (full-close eligible) |

Code-read expectation: the full close succeeds without `--force` and without `--backfill-only`. Expectation is still not a verdict — REQ-GFR-013/AC-GFR-015 keep the attempt-then-observed-fallback shape and forbid pre-deciding.

**Backfill ordering (load-bearing — audit D1 correction).** `needsSHABackfill` (closer.go:397-405) recognizes exactly four backfillable forms: empty, `(this commit)`, `(pending)`, `<pending>`. The predecessor's actual field value (progress.md:261) is the prose placeholder `pending-backfill — the single sync commit on WT-graph-freshness cannot contain its own SHA …` — lowercased/trimmed it matches NONE of the four, so `needsSHABackfill` returns false and the close's auto-backfill (closer.go:324-329) never fires for this SPEC. The placeholder also satisfies the 3-phase predicate's non-empty check, so the close SUCCEEDS with the placeholder intact: without the manual backfill, `pending-backfill — <prose>` is frozen as the permanent §E.4 `sync_commit_sha`. Hence M4's order: manual backfill of `2fc4b40a6` into §E.4 FIRST (D3 placeholder-backfill exemption — a sanctioned surface). The `resolveRecentSpecCommitSHA` resolution path (most recent SPEC-ID-mentioning commit, closer.go:430-434) exists in code but is UNREACHABLE for this placeholder form — it would have been a wrong-SHA hazard only had the placeholder used a recognized form (post-M4-amendment it would resolve a t279 commit).

**Necessary-vs-sufficient (CR #1665 3865025108, stated plainly).** The enforced precondition — `validatePreconditions` precondition 2, closer.go:694-739 — tests only that the §E.4 `sync_commit_sha` string is **non-empty**. Passing that check is *necessary, not sufficient*, for a genuine sync commit: the literal sentinel `pending-backfill — …` passes the emptiness check while being prose, not a SHA, and no SHA-format validation exists anywhere on that path (`needsSHABackfill`'s four-form match is the only stricter gate, and this form does not match it). Where the enforced check lives: entirely in `internal/spec/closer.go` (`validatePreconditions` precondition 2 + `needsSHABackfill`); the pre-close manual-backfill ordering rule is the discipline that actually guarantees a real SHA lands — the code does not enforce it. (Corollary, recorded not fixed: the close would equally accept any non-empty non-placeholder string that is not a real SHA — a SHA-format guard in `validatePreconditions` would be a code change, out of this SPEC's scope.)

**Close-commit mechanics.** The close stages EXACTLY the predecessor's spec.md + progress.md by explicit path (closer.go:340-351) and generates its own subject: `chore(SPEC-V3R6-GRAPH-FRESHNESS-001): Mx-phase audit-ready signal + 3-phase close` (closer.go:354). That machine-generated commit is the one commit on the branch whose subject carries no t279 card id (it does carry the full SPEC-ID); traceability rides the dispatch's `card:` field and the surrounding t279 commits.

**Divergence from the triage table.** The triage's A-3 row 4 ("§E.5(Mx) 저작 → `moai spec close` → …") predates this code verification; this SPEC supersedes it — no §E.5 is authored, and the backfill precedes the close. The 2 pass-with-debt ACs (AC-GF-012/022) remain legitimate recorded debt this SPEC does not clear (spec.md §F).

`sync_commit_sha` backfill value: `2fc4b40a6` — designated by the integrating lead for the predecessor's sync commit (the §E.4 placeholder records "the integrating lead backfills"). Encoded as the end-state field value in AC-GFR-015.

## §5. RED-now baselines (evidence pinning)

All AC baselines are pinned to tree `c9eed8ac6` — a fixed SHA the delivering branch stands on, not a moving ref. The three verify reports are the per-finding evidence record (file:line + grep-zero results where applicable). Representative grep-zero baselines relied on by ACs:

- `grep -n 'does not cover' internal/graph/*_test.go` → 0 (AC-GFR-002)
- No `toolText` helper; `res.Content[0].(mcp.TextContent).Text` one-value assertions at 5 sites + :80 (AC-GFR-006)
- 5 untagged astx test files; `grep -l "go:build" internal/navigator/astx/*_test.go` → none (AC-GFR-012)
- `grep -rn "오래되었으면" docs-site/content/` → 0 (AC-GFR-013)
- Five `graph_freshness` inventory entries `class: R` / `evidence: none` (AC-GFR-011)

The `AC-GFR-*` token set has zero pre-implementation tree hits by construction (new domain prefix, never used before this SPEC).

## §6. Deviations and judgment calls

1. **research.md at Tier M** — the Tier M artifact set is spec/plan/acceptance (+progress); research.md is a Tier L artifact. Included here at the brief's explicit request (provenance: triage method, reconciliation, close-path analysis). Recorded as an intentional deviation, not a tier misclassification — the frontmatter stays `tier: M` per the card classification.
2. **M1 carries 10 of A-1's 11 findings** — finding #11 (astx CGO tags) executes in M3 per the triage's grouping note; the brief's "M1 = 11 findings" counts the A-1 section. Adopted total unchanged (29).
3. **No `depends_on`** — predecessor status is `implemented`, and strict fulfillment requires `completed`; a depends_on would block the run pre-flight circularly (M4 completes the predecessor). `related_specs` carries the relation instead.
4. **Predecessor amendment is not an `amendment_of` lifecycle event** — the predecessor is `implemented` (not `completed`), so the `completed → in-progress (amendment)` machinery does not apply; the M4 body corrections + version bump + the §E.4 SHA backfill (D3 exemption surface) are non-transition content edits, routed through manager-spec re-delegation per the ownership matrix.
5. **CR thread resolution excluded from run scope** — the triage's §"PR #1648 스레드 정리 매핑" (42 unresolved threads → 0) is card-level sync-phase work; recorded in spec.md §F so the boundary is explicit rather than implicit.
6. **M0 recorded as a precondition, not a REQ/AC-bearing milestone** — the restamp was delivered before SPEC issuance (commit `52f7ba135`); an AC for already-delivered work would be PASS-at-authoring, so M0 lives in plan.md §F as an already-executed precondition + spec.md §B.1/§C, and only its forward obligation became requirements (REQ-GFR-014 / AC-GFR-016). Follow-up count also updated 4 → 5 (F5, §7).

## §7. M0 restamp and the squash-merge × stamp structural defect (triage-table.md §F5)

- **Incident**: the #1648 squash merge (`6786c3fa4`) orphaned t250's stamp commit `0d15864ae90b` — the commit object existed only on the branch, not in main's history. Main's graph-freshness check turned not-comparable → red, inherited by every subsequent PR (measured on lane-4 #1662; t274 completion report).
- **Worktree measurement** (at `c9eed8ac6`): `git merge-base --is-ancestor 0d15864ae HEAD` → rc=1 — orphan confirmed. The locally-observed fresh verdict (the object was fetched) vs the CI-observed exit 2 (the shallow checkout lacks the object) are two shapes of one root cause: the stamp named a branch-local commit.
- **Executed countermeasure** (t279's first commit, `52f7ba135` — the branch HEAD at SPEC authoring): restamp against main-reachable `c9eed8ac6`. Measured chain: `mx scan --quiet` → `graph build` → `graph stamp codemaps` → **`graph build` re-run** (settled order — the stamp mutates provenance.json and stales the edges fingerprint) → `graph check`: 3/3 layers fresh, exit 0.
- **Recurrence constraint in this SPEC**: REQ-GFR-014 / AC-GFR-016 — final PR stamp main-reachable; branch-HEAD restamp forbidden; M1-M3 churn (~15-20 described-source files) absorbed under threshold 40 by design.
- **Countermeasure proposals — follow-up card F5, out of scope** (spec.md §F): (1) CI pre-merge stamp-reachability guard in the graph-freshness job — verify provenance.json's commit_sha is an ancestor of the PR base (origin/main); the only point that catches an orphan stamp before merge (recommended); (2) `moai graph stamp codemaps --commit <main-ancestor>` explicit-commit mode (ergonomic support); (3) recorded-but-not-proposed heavier alternatives — post-merge main restamp routine (heavy, not recommended) and a merge-commit strategy switch (record only).
