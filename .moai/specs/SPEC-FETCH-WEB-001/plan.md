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

Card **t1762** (`SPEC-JEV-LEDGER-001`) owns the ledger/consumers work, and it lives in a **new package `internal/jevledger`** plus its own `internal/config` and `internal/template` surfaces — its SPEC states it "does not edit `internal/jev/jev.go` or any other file in `internal/jev`" (read at commit `954adab25e`; re-read the then-current t1762 artifact at implementation time, per the audit's staleness caveat). The earlier premise that both cards touch `internal/jev/jev.go` is therefore RETIRED as stale: `internal/jev` is exclusively this SPEC's to harden on the **client surface** (disk cache, Retry-After backoff, semaphore, usage new-vs-cache separation, alias-staleness warning). The boundary this SPEC must respect is `internal/jevledger` and t1762's config/template artifacts: this SPEC's implementation does not modify, rename, or restructure them; if a needed seam is discovered there, it returns a blocker report. Residual merge-order risk is confined to `internal/config` and `internal/template` (both cards may touch nearby files); whoever lands second rebases mechanically.

## §B. Reviewable decisions (review these first)

- **B1 — Off-gate behavior (pin).** While `workflow.jev.enabled=false`: zero network requests **including the page fetch** (the pipeline cannot run without Jev, so fetching the page would be spend-adjacent I/O with no use), a stderr notice naming WebFetch as the fallback, **exit code 2**. Exit 2 (not 1): the command is *unavailable*, the same class as `fetch_error`, and scripts distinguishing "no answer on the page" (1) from "could not / would not fetch" (2) keep working. Upstream jevfetch has no off-gate; this is a moai addition pinned here. Reconciliation with `SPEC-JEV-CORE-001` REQ-JEVC-007 ("every consumer of an unavailable result shall … exit 0, and continue the workflow"): REQ-JEVC-007 binds in-process consumers that continue a workflow; `moai fetch web` is a dedicated one-shot read-only command, not a workflow continuation, and reports its own unavailability through its exit-code contract (exit 2, the `fetch_error` class). This carve-out is the lane's recorded reading (B1 lane-adjudicated); no operator escalation required.
- **B2 — Exit-code mapping (port contract).** 0 = `filtered`/`passthrough`; 1 = `no_match`/`empty`; 2 = `fetch_error` (+ off-gate per B1). This matches upstream's observable contract and is asserted per status in unit tests.
- **B3 — Package layout.** New package `internal/fetchweb` owns blocks + scoring + pipeline + the page fetcher; the CLI layer (`internal/cli`) and MCP dispatch stay thin (arg parsing → pipeline call → output shaping). No logic in `internal/cli`.
- **B4 — Token approximation accepted.** All thresholds (3000-token windows, 2000-token passthrough, 300-token backfill, 400-char context is *characters*, not tokens) use `internal/jev.EstimateTokens`. Windows may split at slightly different points than upstream's tiktoken counts; golden cases are authored against the Go estimator's counts, not upstream's.
- **B5 — Cache key uses the interpreted model id, never `jev-latest`; mechanism pinned (D6).** The `Request` type carries no model field, so the id is resolved before send by this lookup sequence: (1) an explicit per-request override when the caller provides one; (2) the configured default interpretation model (no such knob exists on this tree today — the lookup falls through to step (3)); (3) alias resolution from the persisted last-observed alias→model map stored under the cache dir (`~/.moai/cache/jev/models.json`, updated whenever a network answer reports its interpreted model id). A cache lookup is keyed by body-SHA + that resolved id. **Stale selection (keep-and-warn probe, plan-audit iter2 N2)**: on a miss under the resolved model id, the client probes the most recent prior-model entry for the same body-SHA and serves it with the staleness warning — without this probe the composite key would leave preserved older-model entries unreachable after an alias move, making the serve-stale branch unimplementable. **Restart-then-offline hit**: the map is disk-persisted, so a fresh process resolves the same id and serves the cached answer with the network closed. **Alias movement**: on a network answer whose reported model differs from the map's entry, the map is updated and the older entry's cached answers are *kept* (still valid answers for their model) but served with a staleness warning naming the cached model id — no silent invalidation, no deletion (keep-and-warn, lane-adjudicated B5, resolvable now that this mechanism is pinned).
- **B6 — Bench command surface.** A `moai jevmeasure bench` subcommand: bench-case loader (question, expected quote, page snapshot) + `--baseline <path>` compare mode writing a `baseline.json` and diffing a run against it (the concrete subcommand name is pinned here for the M6 delegation). It records containment and token-savings; it asserts **no** upstream numeric claim (Gap carried).

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

### M1 (High) — `internal/fetchweb` block splitting + token accounting

Block kinds, heading paths, chrome removal, `<aside>` separation; golden-case fixtures. Data model most likely to change on review. (REQ-FW-001, REQ-FW-002; AC-FW-001)

### M2 (High) — Two-stage scoring + selection rules

Stage-1 windows/0.6; stage-2 standalone+neighbor/0.8, same-heading 0.6; no_match decided first (zero stage-1 candidates, REQ-FW-006 decision order); first-block keep, <300-token backfill, header-row accompaniment, verbatim order-preserving emission. Against a fake answerer injected at the Jev seam. (REQ-FW-004..006; AC-FW-002, AC-FW-003)

### M3 (High) — CLI command + output contract + off-gate

`moai fetch web`, statuses, exit codes, `--json`, bounded page fetch, WebFetch fallback notice. (REQ-FW-003, REQ-FW-007, REQ-FW-008; AC-FW-004..007)

### M4 (Medium) — MCP `fetch_web`

Catalog entry + dispatch sharing the M3 pipeline. (REQ-FW-009; AC-FW-006)

### M5 (Medium) — `internal/jev` client hardening

Disk cache (body-SHA + resolved-model-id key per B5's pinned mechanism, atomic replace), Retry-After backoff composing with REQ-JEVC-008, semaphore, usage new-vs-cache separation, alias-staleness warning. `internal/jev` is exclusively this SPEC's under the revised §A.2 boundary. (REQ-FW-011..013, REQ-FW-016; AC-FW-009..011)

### M6 (Medium) — `internal/jevmeasure` bench

Case format, `moai jevmeasure bench --baseline` compare command (B6), containment + token-savings report. (REQ-FW-014; AC-FW-012)

### M7 (Low, mechanical) — Template packaging

Skill `moai-fetch-web`, permission rule, glm-web-tooling routing row, Codex `network_access` note; `make build` template-parity + neutrality guard. (REQ-FW-015; AC-FW-013)

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
