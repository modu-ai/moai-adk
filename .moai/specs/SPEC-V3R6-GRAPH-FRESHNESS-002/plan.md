# Implementation Plan — SPEC-V3R6-GRAPH-FRESHNESS-002

## §A. Context

- **Work location**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t279`, branch `WT-t250-followup`, baseline tree `c9eed8ac6` (includes #1648 squash `6786c3fa4`). Branch HEAD at authoring: `52f7ba135` (the M0 restamp — provenance.json only; the RED-now pins at `c9eed8ac6` are unaffected). Run-phase commits land on this branch; the card-level PR to `main` follows the repo-local all-tier PR policy (Route B only — the Route A direct-push allowance in the delegation template's B9 is **overridden** here).
- **SPEC artifacts**: `.moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-002/{spec,plan,acceptance,progress}.md` + `research.md`. Predecessor artifacts at `.moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-001/` (status `implemented`, version 1.1.0 — M4 amends to 1.2.0).
- **Scope definition (authoritative)**: `.moai/reports/t279/triage-table.md` sections A-1..A-4. The three verify reports carry the per-finding file:line RED-now baselines pinned at `c9eed8ac6`.
- **Predecessor close state today**: `§E.4 sync_commit_sha: pending-backfill`; no `§E.5` section (schema-correct — none is wanted); `§E.3 ac_pass_with_debt_count: 2` (AC-GF-012, AC-GF-022 — recorded debt, non-blocking per the code-verified close path). See research.md §4 for the close-path analysis.

## §B. Known Issues (filtered for this SPEC's domain)

- **B4 Frontmatter canonical schema** — predecessor edits touch `version`/`updated` only; canonical field names (`created`/`updated`/`tags`) preserved.
- **B5 CI 3-tier awareness** — spec-lint, golangci-lint, per-OS test can each fail; classify NEW vs pre-existing against the pre-flight baseline.
- **B6 spec-lint heading convention** — this SPEC's exclusions use `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule`).
- **B8 Working tree hygiene** — stage by explicit pathspec; `.moai/state/`, `.moai/cache/`, unrelated untracked files untouched.
- **B10 Scope discipline / PRESERVE** — the 33 already-fixed sites (triage §E) are citation targets, not edit targets; no drive-by refactors of adjacent code.
- **Test isolation** — `t.TempDir()` everywhere; no OTEL env in parallel tests; `filepath.Abs` for user-supplied paths.
- **Self-gate hazard** — editing `.moai/project/codemaps/dependencies.md` (M3) does NOT trip the graph-freshness gate: the codemaps metric counts described-source files (`internal/`, `cmd/`, `pkg/`), and `.moai/` is not described-source. Do not "fix" a stale-verdict that does not exist.
- **Stamp/build ordering (local check hygiene)** — any codemaps-set mutation stales the edges fingerprint: stamping itself (it mutates provenance.json — hence M0's settled stamp-then-rebuild order), M3's dependencies.md edit, and even this SPEC's own artifact additions under `.moai/specs/`. Re-run `moai graph build` before reading a local `moai graph check`; the CI job bootstraps the mechanical layers itself, so this is a local-hygiene rule, not a CI concern.

## §C. Pre-flight (before M1)

```bash
git -C <worktree> branch --show-current && git -C <worktree> rev-parse --short HEAD
go build ./... && CGO_ENABLED=0 go build ./...
go test ./internal/graph/ ./internal/cli/ ./internal/hook/quality/ ./internal/navigator/astx/ ./internal/config/ -count=1   # targeted baseline only
golangci-lint run --timeout=2m 2>&1 | tail -5
moai spec lint .moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-002/spec.md
```

Baseline outputs recorded before any change, so NEW-vs-pre-existing classification is possible.

## §D. Constraints

- **PRESERVE**: the 33 already-fixed file:line sites cited in triage §E; predecessor `§E.2`/`§E.3` audit evidence (M4 appends/authors, never rewrites recorded history); `.moai/reports/t279/` (evidence base, read-only); the D1 deviation record and its comments.
- **Forbidden**: `go test ./...` locally (CLAUDE.local.md §4/§6); `git add -A`/sweep staging; `--no-verify`; force-push; editing predecessor `spec.md`/`acceptance.md` bodies outside the M4 manager-spec re-delegation.
- **Ownership routing**: M1-M3 → manager-develop; M4 body corrections + §E.4 SHA backfill → manager-spec re-delegation (orchestrator-mediated, D-NEW-1 pattern, separate commit `feat(SPEC-V3R6-GRAPH-FRESHNESS-001): correction-and-close amendment (t279 M4)`); close CLI invoked afterward — the tool performs its own atomic commit.
- **Required**: Conventional Commits with card id `t279` in every commit message on the branch; per-milestone commits.
- **Stamp reachability (HARD, sync-phase)** — the PR's final codemaps stamp keeps naming a main-reachable commit (`c9eed8ac6`); restamping against the branch HEAD is FORBIDDEN (squash merge re-orphans it — REQ-GFR-014/AC-GFR-016, triage §F5). M1-M3 churn (~15-20 described-source files) stays under threshold 40 by design; if a refresh ever becomes necessary, it names another main-reachable commit and is recorded.

## §E. Self-Verification (per delegation, attribution triple a/b/c)

E1 AC binary matrix (command + verbatim output + tree SHA) · E2 cross-platform (`go build ./...`, `GOOS=windows go build ./...`, plus `CGO_ENABLED=0` test legs for graph/astx) · E3 coverage on touched packages · E5 lint (NEW vs baseline) · E6 commit/push state · E8 RED evidence for every new test (failing run observed before GREEN; the triage's grep-zero baselines are the standing RED-now record). Report per the 5-section evidence format.

## §F. Milestones

### M0 — codemaps provenance restamp (ALREADY EXECUTED — precondition, not a delegation unit)

- Delivered pre-issuance as branch commit `52f7ba135`: restamp against main-reachable `c9eed8ac6`, repairing the `0d15864ae90b` orphan that the #1648 squash merge left (main's graph-freshness check was red for every inheriting PR — lane-4 #1662 measured).
- Measured chain (triage-table.md §F5): `moai mx scan --quiet` → `moai graph build` → `moai graph stamp codemaps` → **`moai graph build` re-run** (the stamp mutates provenance.json and stales the edges fingerprint — build-after-stamp is the settled order) → `moai graph check`: all three layers fresh, exit 0.
- Run/sync phases inherit this state. The forward obligation is REQ-GFR-014 (AC-GFR-016): the final PR stamp stays main-reachable; no branch-HEAD restamp at any phase.

### M1 — test-policy remediation (triage A-1 #1–#10)

| # | comment-id | file (current line) | action | AC |
|---|---|---|---|---|
| 1 | 3855001995 | internal/graph/citation_test.go (new TC) | test the RegionHash≠excerpt internal-consistency branch (citation.go:148-152) | AC-GFR-002 |
| 2 | 3855002004 | internal/graph/codequery_test.go:42-45,146,175 | `!cgo` skip guard: `t.Fatal` → `t.Skip` when extraction unsupported | AC-GFR-001 |
| 3 | 3855002013 | internal/graph/codequery_test.go:255-259 | positive `inA` assertion alongside the ":B"-absence check | AC-GFR-005(a) |
| 4 | 3855149281 | internal/graph/check_test.go:199,225 | `strconv` replaces hand-rolled itoa/padding (quoteJSON half already fixed) | AC-GFR-005(b) |
| 5 | 3855001906 | internal/cli/graph_check_test.go:59-61,168,231 | `json.Marshal`-built provenance fixtures (Windows-path JSON breakage — real defect) | AC-GFR-005(c) |
| 6 | 3855001928 | internal/cli/mcp_code_tools_test.go:113,125,153,178,198 **+ :80** | `toolText`-style shape-check helper; 6 unguarded `res.Content[0]` sites (Minor-2a absorbed — :80 panic risk) | AC-GFR-006 |
| 7 | 3855001933 | internal/cli/mcp_code_tools_test.go:188 | assert Symbol/Via match content (tags already match the wire) | AC-GFR-003 |
| 8 | 3855001948 | internal/cli/mcp_code_tools_test.go (new TC) | table: required-parameter rejection per handler + literal `..` path case | AC-GFR-003 |
| 9 | 3855002099 | internal/hook/quality/gate_graph_freshness_test.go:113-132 | fixture stamps all three layers; assert the "all layers fresh" notice (gate.go:1216) | AC-GFR-004 |
| 10 | 3855149237 | internal/cli/graph_refresh_test.go:152-185 | budget-overrun determinism via injected duration (new seam in graph_refresh_cli.go) | AC-GFR-007 |

Commit: `test(SPEC-V3R6-GRAPH-FRESHNESS-002): M1 test-policy remediation (t279)`.

### M2 — code-policy remediation (triage A-2 #12–#21 + Minor-2b)

| # | comment-id | file (current line) | action | AC |
|---|---|---|---|---|
| 12 | 3855149289 | internal/graph/check.go:109-117 | error path returns layer reports OR doc-contract corrected — contradiction gone | AC-GFR-008(a) |
| 13 | 3855149309 | internal/graph/check.go:254-255 | `sidecarAbsentReason` → const + accurate comment | AC-GFR-008(b) |
| 14 | 3855149315 | internal/graph/check.go:309,327 | `MXIndexNeedsRefresh` threshold caller-injected (no `DefaultThresholds()` inside) | AC-GFR-008(c) |
| 15 | 3855149325 | internal/graph/meta.go:109-120 + check.go:348-369 | one shared fingerprint-comparison helper for probe + check | AC-GFR-008(d) |
| 16 | 3855149332 | internal/graph/symbol.go:31-34,94-97 | `%w` wrapping at both bare-return sites | AC-GFR-009(a) |
| 17 | 3855001978 | internal/cli/mcp_code_tools.go:26-36,81-87 | package `toolJSON`/`toolErr` replace `jsonToolResult`/`NewToolResultError` | AC-GFR-009(b) |
| 18 | 3855149248 | internal/cli/graph_stamp.go:46-68 | CLI-boundary fs error hygiene — no absolute local path leak to the user | AC-GFR-009(c) |
| 19 | 3855149254 | internal/cli/graph.go:150 | refresh decision evaluates the selected `--edges` artifact | AC-GFR-010 |
| 20 | 3855149357 | internal/mx/provenance.go:157-158 | `gitOut` comment stops overclaiming (clean `git status --porcelain` is empty+nil) | AC-GFR-009(d) |
| 21 | 3855001991 | internal/config/testdata/shipped_key_inventory.yaml:380-394 | 5 `graph_freshness` keys R→W with the three reader citations | AC-GFR-011 |
| 22b | (residual) | internal/graph/codequery.go:153,244,323 + internal/cli/mcp_server.go:506 | swapped CR-ID comments fixed; scan-window literal 8 → named const; "shared by MCP description" overclaim corrected (Minor-2b) | AC-GFR-009(e) |

Commit: `fix(SPEC-V3R6-GRAPH-FRESHNESS-002): M2 code-policy remediation (t279)`.

### M3 — astx + docs/wording (triage A-1 #11 + A-4 #26–#30)

| # | comment-id | file (current line) | action | AC |
|---|---|---|---|---|
| 11 | 3855002141 | internal/navigator/astx/*_test.go (5 untagged files) | `//go:build cgo` on CGO-positive tests + `!cgo` fallback test | AC-GFR-012(b,c) |
| 30 | 3855002146 | internal/navigator/astx/queries/go.scm:18 | `raw_string_literal` alternative for import_spec path capture + named test | AC-GFR-012(a) |
| 26 | 3855001858 | .moai/project/codemaps/dependencies.md:77,116 | hook summary bullet gains `graph` (matches the :77 Mermaid edge) — surgical edit, NOT a regeneration (see §B self-gate note) | AC-GFR-013 |
| 27 | 3855001863 | .moai/reports/t250/m5-baseline.md:18-19 | developer-local transcript path → repository-relative label | AC-GFR-013 |
| 28 | 3855001901 | CHANGELOG.md:46 | cumulative MCP count "21 to 24" → "25 to 28" (prior entry :97 = 25; mechanical count = 28) | AC-GFR-013 |
| 29 | 3855149226 | docs-site/content/ko/cli-reference/graph.md:53 | "오래했으면" → "오래되었으면" (ko-only content; hugo build gate) | AC-GFR-013 |

Commit: `fix(SPEC-V3R6-GRAPH-FRESHNESS-002): M3 astx + docs remediation (t279)`.

### M4 — predecessor SPEC correction + close (triage A-3 #23–#25 + card close task)

Sequencing (ownership-sensitive; close path code-verified at `internal/spec/closer.go` — research.md §4):

1. **manager-spec re-delegation** (orchestrator-mediated; separate commit):
   - `spec.md:87` REQ-GF-004 gains the third When-clause: exit 2 for not-comparable system errors — no verdict, failing operation named (matches graph-freshness.yml:48, docs-site 0/1/2, and the implemented F4 fix).
   - `acceptance.md` §D.1: AC-GF-008 moves SHOULD → MUST (mutant kill already observed, progress.md:75).
   - `acceptance.md` AC-GF-020 Then-clause gains the non-Go declaration-set qualifier.
   - `spec.md` frontmatter: `version: "1.2.0"`, `updated` advanced.
   - NO §E.5 section is authored — the modern 4-section schema retired it (`spec-frontmatter-schema.md`), and the close's precondition 2 accepts the 3-phase predicate (§E.4 marker + non-empty `sync_commit_sha`; closer.go:702-714). §E.2/§E.3 recorded history is NOT rewritten.
   - Commit: `feat(SPEC-V3R6-GRAPH-FRESHNESS-001): correction-and-close amendment (t279 M4)`.
2. **Manual SHA backfill (before the close — the ordering is load-bearing)**: set `§E.4 sync_commit_sha: 2fc4b40a6` in the predecessor's progress.md (D3 placeholder-backfill exemption surface). Why first: `needsSHABackfill` (closer.go:397-405) recognizes only four backfillable forms — empty, `(this commit)`, `(pending)`, `<pending>` — and the predecessor's prose placeholder (`pending-backfill — …`, progress.md:261) matches NONE of them, so the close's auto-backfill (closer.go:324-329) never fires for this SPEC: without the manual backfill, the close SUCCEEDS and freezes the placeholder as the permanent §E.4 value. The ordering prevents placeholder-freezing. (The `resolveRecentSpecCommitSHA` resolution — most recent SPEC-ID-mentioning commit, closer.go:430-434 — exists in code but is unreachable for this placeholder form.)
3. **Close attempt** (output recorded verbatim into THIS SPEC's progress.md §E.2):
   - `moai spec close SPEC-V3R6-GRAPH-FRESHNESS-001` → record exit code + output.
   - On observed precondition failure (exit 1): `moai spec close SPEC-V3R6-GRAPH-FRESHNESS-001 --backfill-only` → record exit code + output.
   - No pre-decision. Code-read expectation: the full close passes all four preconditions (§E.2 present; §E.4+SHA satisfies precondition 2; acceptance.md carries 0 genuine debt/FAIL markers so preconditions 3-4 hold — `hasGenuinePassWithDebtVerdict` scans acceptance.md only, closer.go:633/:667; status is `implemented`) — but expectation is not a verdict; the fallback stays encoded.
4. **End state**: predecessor `status: completed`; `§E.4 sync_commit_sha: 2fc4b40a6`.

Close-commit note: the tool stages exactly the predecessor's spec.md + progress.md by explicit path (closer.go:340-351) and generates its own subject — `chore(SPEC-V3R6-GRAPH-FRESHNESS-001): Mx-phase audit-ready signal + 3-phase close` (closer.go:354). That machine-generated commit is the one commit on the branch whose subject carries no t279 card id (it does carry the full SPEC-ID); its traceability rides the dispatch's `card:` field and the surrounding t279 commits. Record the actually-produced subject in §E.2.

### Decision-reversibility ordering (within-milestone)

Inside each milestone, land the highest-change-likelihood decisions first: M2's #14 (threshold injection — signature change) and #15 (shared helper — behavior-adjacent) before comment-only items; M4's body corrections before the close (irreversible status transition).

## §G. Anti-Patterns

- Regenerating codemaps to "fix" the dependencies.md bullet — regeneration is the LLM writer's debt (AC-GF-012), explicitly out of scope; the bullet is a surgical edit.
- Rewriting predecessor §E.2/§E.3 recorded evidence while amending — M4 edits the three named body locations, bumps the version, and backfills the §E.4 SHA only.
- Pre-deciding the close path in the delegation prompt — the run phase must observe the CLI output and decide.
- Clearing AC-GF-012/022 debt opportunistically — recorded, not cleared.
- Running the local full suite to "be thorough" — targeted packages only; CI judges the full matrix.
- Restamping codemaps against the branch HEAD (or "refreshing the stamp because the branch moved") — the squash merge re-orphans it and re-reddens main; the M0 stamp at `c9eed8ac6` carries to merge (REQ-GFR-014).
- Reading a local `moai graph check` edges-stale verdict after content changes without re-running `moai graph build` first — the staleness is the fingerprint contract working, not a defect (§B stamp/build ordering).

## §H. Cross-References

- spec.md §D (REQ-GFR-001..014) ↔ acceptance.md §D.2 (traceability)
- `.moai/reports/t279/triage-table.md` (scope SSOT) + 3 verify reports (RED-now baselines)
- Predecessor: `.moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-001/{spec,acceptance,progress}.md`
- Ownership matrix: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix + § Forbidden ownership crossings
- Close CLI: `moai spec close --help` + `internal/spec/closer.go` (code-verified precondition semantics, backfill resolution, staging, generated subject — research.md §4)
