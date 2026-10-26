# SPEC-FETCH-WEB-001 — Acceptance Criteria

Given-When-Then is the verification layer's format; the GEARS requirement wording lives in `spec.md` §C and is not restated here.

Every release-blocking criterion below is **two-celled** per `.claude/rules/moai/development/verification-completeness.md` §2: a RED-now cell in the §G evidence ledger (command + verbatim stdout + exit code, pinned to the pre-implementation tree), and a green-path cell naming the milestone that flips it. Criteria without a RED cell (AC-FW-002, AC-FW-003, AC-FW-004, AC-FW-008, AC-FW-010, AC-FW-011, AC-FW-014) are behavior-of-new-code assertions whose RED state is the absence of the entire package (ledger row R1 covers them collectively); they are release-blocking with R1 as their starting observation.

## §A. Pipeline core

**AC-FW-001** — Given golden-case HTML fixtures covering paragraphs, list items, table rows, code blocks, nested heading paths, page chrome, and an `<aside>` note, When the block splitter runs, Then each expected block is produced with its correct heading path, chrome contributes zero blocks, and the aside content is a separately-tagged block. Method: table-driven golden tests in `internal/fetchweb`. Green path: M1. RED: ledger R1.

**AC-FW-002** — Given a fake Jev answerer returning controlled probabilities, When the two-stage scorer runs, Then: a block with stage-1 p ≥ 0.6 becomes a candidate; a candidate with stage-2 standalone-plus-context p ≥ 0.8 is kept; a candidate below 0.8 is kept only when a same-heading neighbor scores ≥ 0.6; the first block is kept regardless; a kept set under 300 tokens is backfilled in descending stage-1 score order; and every kept table row is accompanied by its header row. Method: one sub-test per selection rule with the fake answerer pinned per rule. Green path: M2. RED: ledger R1.

**AC-FW-003** — Given a document whose split token count is ≤ 2,000 and a Jev seam instrumented to fail the test on any call, When the pipeline runs, Then the output status is `passthrough`, the whole document is present verbatim, and zero scoring calls were made. Green path: M2 (unit) / M3 (command-level). RED: ledger R1.

**AC-FW-004** — Given command-level invocations against fixture pages and a fake Jev server, When each terminal condition is exercised (relevant blocks, ≤2,000-token page, no block survives selection, empty page, unreachable page), Then the statuses `filtered` / `passthrough` / `no_match` / `empty` / `fetch_error` map to exit codes 0 / 0 / 1 / 1 / 2 respectively, the first output line begins `#`, each block is preceded by a `##` heading-path line, block text is verbatim, and `--json` emits the same information as a parseable JSON object. Green path: M3. RED: ledger R1 (no command exists).

## §B. Command and MCP surface

**AC-FW-005** — Given the CLI command tree, When `moai fetch web --help` is invoked on a built binary, Then the command exists with the exact name `fetch web`, takes `<url>` and `"<question>"`, and documents `--json`. Green path: M3. RED: ledger R2 (`grep -rn "moai fetch web" internal/cli` — zero hits at `ec9091a80c`).

**AC-FW-006** — Given the MCP tool catalog, When it is enumerated, Then a `fetch_web` tool is present and its dispatch reaches the same pipeline as the CLI command (one shared entry point, no duplicated scoring logic). Method: catalog test + a dispatch-path test asserting both surfaces call the same pipeline function. Green path: M4. RED: ledger R3 (`grep -n '"fetch_web"' internal/cli/mcp_server.go` — zero hits at `ec9091a80c`).

**AC-FW-007** — Given `workflow.jev.enabled=false` and an HTTP transport instrumented to fail the test on any outbound request, When `moai fetch web <url> "<question>"` runs, Then zero network requests occur (no page fetch, no Jev call), a WebFetch fallback notice is emitted, and the exit code is 2. This is the card's named zero-network proof. Green path: M3. RED: ledger R1 + R2.

**AC-FW-008** — Given a fake page server serving a fixture page and a fake Jev server returning a scripted answer sequence, When one E2E command invocation runs against both, Then the pipeline completes end to end and the output satisfies the AC-FW-004 contract for the `filtered` case. No real endpoint is contacted (asserted by binding both fakes and failing on any other dial). Green path: M3 (E2E test in the same milestone). RED: ledger R1.

## §C. Client hardening (internal/jev, client surface only)

**AC-FW-009** — Given two identical Jev requests differing only in interpretation model id and a cache directory under a temp HOME, When both are issued, Then two distinct cache files exist (the model id is in the key); and given a cache hit for the same body+model, When the same request is re-issued with the network transport closed, Then the cached answer is served. Cache writes are atomic (no partial file observable). Green path: M5. RED: ledger R4 (`grep -rn "cache/jev" internal/jev` — zero hits at `ec9091a80c`).

**AC-FW-010** — Given a fake Jev server responding 429 with `Retry-After: 1`, When a request is sent, Then the client waits the parsed delay and retries, and the retry count is bounded; and given N concurrent requests and a semaphore of width W, When they are issued, Then at most W are in flight simultaneously (observed server-side by the fake). Green path: M5. RED: ledger R4.

**AC-FW-011** — Given one cache-served and one network-served answer, When usage aggregation is read, Then the two are reported in separated counters (new vs cache); and given a cache entry whose model id is older than the currently interpreted model, When it is served, Then a staleness warning naming the cached model id is emitted. Green path: M5. RED: ledger R4.

## §D. Bench (internal/jevmeasure)

**AC-FW-012** — Given a bench case file (question, expected quote, page snapshot) and a recorded run result, When the baseline comparison command runs, Then it reports per case whether the expected quote is contained in the output and the token-savings rate versus the snapshot's token count, plus a set-level summary; and a second run against the stored `baseline.json` reports pass/fail per case deterministically. Method: golden test with a fixed fixture; no network. Green path: M6. RED: ledger R5 (`grep -rn "baseline.json" internal/jevmeasure` — zero hits at `ec9091a80c`).

## §E. Template packaging

**AC-FW-013** — Given the template tree after M7, When enumerated, Then: the skill directory `moai-fetch-web` exists with a SKILL.md stating when to prefer it over WebFetch and how `--json` and `no_match` are handled; the settings template carries `Bash(moai fetch web:*)`; the GLM web-tooling routing rule carries a "jev active → fetch_web first" row; and the Codex packaging carries the `network_access` note. The template-neutrality CI guard stays green (no `enabled: true`, no machine-specific values). Green path: M7. RED: ledger R6 (skill dir absent) + R7 (permission rule zero hits in `settings.json.tmpl`, which exists).

## §F. Test isolation

**AC-FW-014** — Given the full test suite of every package this SPEC touches, When the production Jev endpoint constant is grepped across `_test.go` files and when every dial in these tests is audited, Then no test reaches a real endpoint — all Jev and page interactions run against fake servers or recorded fixtures. Method: a guard test asserting test HTTP targets are loopback/httptest servers, plus a repo-level grep check. Green path: M3 onward, verified at close. RED: ledger R1.

## §F.1 Traceability (AC → REQ)

| AC | Requirement coverage |
|---|---|
| AC-FW-001 | maps REQ-FW-001, REQ-FW-002 |
| AC-FW-002 | maps REQ-FW-004, REQ-FW-005, REQ-FW-006 |
| AC-FW-003 | maps REQ-FW-003 |
| AC-FW-004 | maps REQ-FW-007 |
| AC-FW-005 | maps REQ-FW-007 |
| AC-FW-006 | maps REQ-FW-009 |
| AC-FW-007 | maps REQ-FW-008 |
| AC-FW-008 | maps REQ-FW-007, REQ-FW-010 |
| AC-FW-009 | maps REQ-FW-011 |
| AC-FW-010 | maps REQ-FW-012 |
| AC-FW-011 | maps REQ-FW-013 |
| AC-FW-012 | maps REQ-FW-014 |
| AC-FW-013 | maps REQ-FW-015 |
| AC-FW-014 | maps REQ-FW-010, REQ-FW-016 |

## §G. RED-now evidence ledger (document-level pin)

**Tree SHA (pins every criterion above unless a criterion carries its own): `ec9091a80c090132ae4dda6c0588b9810c20de53`.** All rows observed 2026-10-26 in the card worktree `.moai/worktrees/t1968`. RED reason in every row: the artifact the criterion measures does not exist on the pre-implementation tree — absence of this SPEC's own deliverables, no unrelated-file contamination.

| Row | Command (single invocation, read-only) | Verbatim stdout | Exit code | Cited by |
|---|---|---|---|---|
| R1 | `ls internal/fetchweb` | `ls: internal/fetchweb: No such file or directory` | 1 | AC-FW-001..004, AC-FW-007, AC-FW-008, AC-FW-014 |
| R2 | `grep -rn "moai fetch web" internal/cli` | (empty) | 1 | AC-FW-005, AC-FW-007 |
| R3 | `grep -n '"fetch_web"' internal/cli/mcp_server.go` | (empty) | 1 | AC-FW-006 |
| R4 | `grep -rn "cache/jev" internal/jev` | (empty) | 1 | AC-FW-009..011 |
| R5 | `grep -rn "baseline.json" internal/jevmeasure` | (empty) | 1 | AC-FW-012 |
| R6 | `ls internal/template/templates/.claude/skills/moai-fetch-web` | `ls: internal/template/templates/.claude/skills/moai-fetch-web: No such file or directory` | 1 | AC-FW-013 |
| R7 | `grep -rn "moai fetch web" internal/template/templates/.claude/settings.json.tmpl` | (empty) — target file confirmed to exist (`ls` on it, exit 0) in the same observation session | 1 | AC-FW-013 |

## §H. Quality gates

- `go test ./internal/fetchweb/... ./internal/jev/... ./internal/jevmeasure/...` passes; full-suite verdict from CI.
- `golangci-lint run` clean on changed packages; `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` pass.
- `make build` succeeds; template-parity and template-neutrality guards green.

## §I. Definition of Done

1. Every AC above passes with observed evidence per the two-cell discipline.
2. The card's three named verdicts hold: unit tests over block-splitting golden cases, scoring selection rules, output format, and exit codes (AC-FW-001/002/004); one E2E test against a fake Jev server (AC-FW-008); a zero-network-when-off proof (AC-FW-007).
3. No test in the suite performs a live Jev call or live page fetch (AC-FW-014).
4. The t1762 boundary is intact — the final diff touches no `internal/jev` ledger/consumers file (REQ-FW-016).
