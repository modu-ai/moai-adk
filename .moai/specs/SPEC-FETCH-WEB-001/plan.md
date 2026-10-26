# SPEC-FETCH-WEB-001 — Implementation Plan

> Milestones are ordered by decision-reversibility: the output contract and pipeline data model (M1-M3) carry the choices most likely to change on review; template packaging (M7) is mechanical and last.

---

## §A. Context

Card t1968. Authoritative scope: `.moai/reports/leader-plan-20261023/jevfetch-analysis-20261026.md` (leader analysis of `github.com/a2sys-ai/jevfetch`, MIT). Dependency: `SPEC-JEV-CORE-001` (the `internal/jev` package, the `workflow.jev.enabled` gate, the credential reader).

### §A.1 Verified tree facts (measured in this worktree at `ec9091a80c`)

| Fact | Source | Observed |
|---|---|---|
| Jev client surface | `internal/jev/jev.go` | `New(enabled bool) *Client`, `(c *Client) Ask(ctx, Request) Result`, `UsageLog() []Usage` / `record(u Usage)`, `Result.OK()`, `Result.NoticeLine()` |
| Approximate token estimator EXISTS | `internal/jev/jev.go:366-369` | `EstimateTokens(s string) int` — conservative 4-bytes-per-token heuristic; **no new estimator needed** |
| Jev config gate | `internal/config/types.go:684-694, 1052-1061` | `Workflow.Jev WorkflowJevConfig` — bare `Enabled bool \`yaml:"enabled"\``; alias + endpoint are compiled constants |
| MCP tool catalog | `internal/cli/mcp_server.go:393, 808` | tool-name map (`"jev_ask": true`) + dispatch block per tool |
| `moai web` verb collision | `internal/cli/web.go:35` | `web [--port N] [--no-open]` — the web UI; confirms why the new command is `fetch web`, not a `web` sub-verb |
| Template skills layout | `internal/template/templates/.claude/skills/` | one dir per skill (`moai-*`); settings template exists at `templates/.claude/settings.json.tmpl` |
| `internal/jevmeasure` | `internal/jevmeasure/measure.go` | measurement frame (two-arm runner, constant-answer baseline, fitted thresholds) exists — bench cases extend it |
| No existing fetch/pipeline artifacts | RED ledger (acceptance.md §G) | package, command, MCP tool, cache, bench, skill all absent; 0 collisions |

### §A.2 Scope boundary — t1762 (concurrent card, another lane)

Card **t1762** owns the `internal/jev` **ledger/consumers** surfaces. This SPEC's `internal/jev` edits are confined to the **client cache/retry surface**: disk cache, Retry-After backoff, semaphore, usage new-vs-cache separation, alias-staleness warning. If implementation discovers a needed seam on the t1762 side, it returns a blocker report — it does not edit across the boundary. Merge-order risk: both cards touch `internal/jev/jev.go`; whoever lands second rebases mechanically (the surfaces are disjoint by this constraint).

## §B. Reviewable decisions (review these first)

- **B1 — Off-gate behavior (pin).** While `workflow.jev.enabled=false`: zero network requests **including the page fetch** (the pipeline cannot run without Jev, so fetching the page would be spend-adjacent I/O with no use), a stderr notice naming WebFetch as the fallback, **exit code 2**. Exit 2 (not 1): the command is *unavailable*, the same class as `fetch_error`, and scripts distinguishing "no answer on the page" (1) from "could not / would not fetch" (2) keep working. Upstream jevfetch has no off-gate; this is a moai addition pinned here.
- **B2 — Exit-code mapping (port contract).** 0 = `filtered`/`passthrough`; 1 = `no_match`/`empty`; 2 = `fetch_error` (+ off-gate per B1). This matches upstream's observable contract and is asserted per status in unit tests.
- **B3 — Package layout.** New package `internal/fetchweb` owns blocks + scoring + pipeline + the page fetcher; the CLI layer (`internal/cli`) and MCP dispatch stay thin (arg parsing → pipeline call → output shaping). No logic in `internal/cli`.
- **B4 — Token approximation accepted.** All thresholds (3000-token windows, 2000-token passthrough, 300-token backfill, 400-char context is *characters*, not tokens) use `internal/jev.EstimateTokens`. Windows may split at slightly different points than upstream's tiktoken counts; golden cases are authored against the Go estimator's counts, not upstream's.
- **B5 — Cache key uses the interpreted model id** recorded from each response (never `jev-latest`). A stale entry (alias moved to a newer model) is *kept* (still a valid answer for its model) but warned about; no silent invalidation, no deletion.
- **B6 — Bench command surface.** A `jevmeasure` bench-case loader (question, expected quote, page snapshot) + a compare command writing a `baseline.json` and diffing a run against it. It records containment and token-savings; it asserts **no** upstream numeric claim (Gap carried).

## §C. Pre-flight (run before M1)

```bash
git rev-parse HEAD && git branch --show-current
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
grep -rn "cache/jev" internal/jev || echo "no cache yet (expected)"
grep -r "Retired\|superseded" internal/jev/jevmeasure 2>/dev/null || echo "no SPEC conflicts"
```

## §D. Constraints (mirror of spec.md §D — binding)

1. Default-off; zero network while off; no new keys; no spend.
2. No live Jev call / live page fetch in any test — recorded responses or fake servers only.
3. No Python deps; no tokenizer dep; approximate tokens documented at use sites.
4. No headless-render fallback (t1965 follow-up).
5. Naming fixed: `moai fetch web` / `fetch_web` / `moai-fetch-web` / `Bash(moai fetch web:*)`.
6. Template neutrality on all application-4 artifacts.
7. `internal/jev` edits confined to the client surface (t1762 boundary, REQ-FW-016).

## §E. Self-verification (per-milestone)

- `go test ./internal/fetchweb/... ./internal/jev/... ./internal/jevmeasure/...` (change-scoped, not `./...` — lane-local discipline).
- `GOOS=windows GOARCH=amd64 go build ./...` after each Go milestone.
- `golangci-lint run` on changed packages; `make build` before M7 template parity.
- RED-first: every M1-M3 unit test is authored before its implementation and its failing output captured (TDD mode; `quality.yaml` constitution.development_mode governs at run time).

## §F. Milestones (priority order = reversibility order)

- **M1 (High) — `internal/fetchweb` block splitting + token accounting.** Block kinds, heading paths, chrome removal, `<aside>` separation; golden-case fixtures. Data model most likely to change on review. (REQ-FW-001, REQ-FW-002; AC-FW-001)
- **M2 (High) — Two-stage scoring + selection rules.** Stage-1 windows/0.6; stage-2 standalone+neighbor/0.8, same-heading 0.6; first-block keep, <300-token backfill, header-row accompaniment, verbatim order-preserving emission. Against a fake answerer injected at the Jev seam. (REQ-FW-004..006; AC-FW-002, AC-FW-003)
- **M3 (High) — CLI command + output contract + off-gate.** `moai fetch web`, statuses, exit codes, `--json`, bounded page fetch, WebFetch fallback notice. (REQ-FW-003, REQ-FW-007, REQ-FW-008; AC-FW-004..007)
- **M4 (Medium) — MCP `fetch_web`.** Catalog entry + dispatch sharing the M3 pipeline. (REQ-FW-009; AC-FW-006)
- **M5 (Medium) — `internal/jev` client hardening.** Disk cache (SHA+model key, atomic replace), Retry-After backoff, semaphore, usage separation, alias-staleness warning. Disjoint from t1762 by §A.2. (REQ-FW-011..013, REQ-FW-016; AC-FW-009..011)
- **M6 (Medium) — `internal/jevmeasure` bench.** Case format, baseline.json compare command, containment + token-savings report. (REQ-FW-014; AC-FW-012)
- **M7 (Low, mechanical) — Template packaging.** Skill `moai-fetch-web`, permission rule, glm-web-tooling routing row, Codex `network_access` note; `make build` template-parity + neutrality guard. (REQ-FW-015; AC-FW-013)

## §G. Anti-patterns

- **AP-1** — Live network in tests (even a "harmless" real URL fetch): violates REQ-FW-010; fake servers only.
- **AP-2** — Paraphrasing block text anywhere in the output path: the port's whole point is verbatim.
- **AP-3** — Editing `internal/jev` ledger/consumers code "while in there": t1762 boundary breach (REQ-FW-016).
- **AP-4** — Shipping `enabled: true` or machine paths in templates: neutrality violation.
- **AP-5** — Asserting upstream's 47% figure or tiktoken-exact counts in tests: the approximation is documented, not compensated.
- **AP-6** — A `web read`-style verb under `moai web`: collision with the web UI (operator-fixed naming).

## §H. Cross-references

- spec.md §C (REQ-FW-001..016), acceptance.md (AC-FW-001..014 + RED ledger)
- `SPEC-JEV-CORE-001` (dependency) · `SPEC-JEV-OPTIN-MEASURE-001` (measurement discipline M6 extends)
- Card t1762 (boundary), card t1965 (headless-render follow-up linkage)
- `moai-mcp-tools-catalogue.md` (MCP catalog conventions), `hns-moaiadk-patterns` skill (add-a-template workflow)
