---
id: SPEC-FETCH-WEB-001
title: "moai fetch web — Go port of the jevfetch pipeline (verbatim web excerpt via two-stage Jev scoring)"
version: "0.2.2"
status: draft
created: 2026-10-26
updated: 2026-10-26
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/fetchweb (new), internal/jev, internal/jevmeasure, internal/cli, internal/cli/mcp_server.go, internal/template/templates"
lifecycle: spec-anchored
tags: "jevfetch, fetch-web, jev, noul-scoring, disk-cache, bench, mcp, template-packaging"
tier: M
depends_on: [SPEC-JEV-CORE-001]
---

# SPEC-FETCH-WEB-001 — moai fetch web (jevfetch pipeline port)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-26 | manager-spec | Initial plan-phase authoring. Card t1968. Authoritative design source: leader analysis `.moai/reports/leader-plan-20261023/jevfetch-analysis-20261026.md` (upstream `github.com/a2sys-ai/jevfetch`, MIT, Python 3.12). Scope = that analysis's applications 1-4 verbatim, including its not-porting list and recorded Gaps. Naming fixed by operator directive 2026-10-26: `moai fetch web` / `fetch_web` / `moai-fetch-web` / `Bash(moai fetch web:*)` (avoids the existing `moai web` verb). |
| 0.2.0 | 2026-10-26 | manager-spec | Plan revision per plan-audit iter1 FAIL 0.80 (report `.moai/reports/t1968/plan-audit.md`; all premises re-verified against source before rewrite). D1: no_match decision order pinned in REQ-FW-006 + AC-FW-004. D2: REQ-FW-015 / AC-FW-013 scoped to this SPEC's own shipped artifacts. D3: REQ-FW-012 composed with REQ-JEVC-008; AC-FW-010 negative cases added. D4: RED basis for AC-FW-010/011 re-based on behavior probes (ledger R8-R10); R1/R6 stderr relabel. D5: B1 exit-2 carve-out vs REQ-JEVC-007 recorded. D6: cache-key mechanism pinned in B5/M5. D7: t1762 boundary rewritten to real paths (`internal/jevledger`; `internal/jev` conflict premise retired as stale). D8: tokenizer-dependency AC guard folded into AC-FW-014. D9: plan milestones converted to headings. B1/B5/B6 lane-adjudicated and resolved. REQ/AC numbering unchanged (16 REQ / 14 AC, Tier M ceilings respected). |
| 0.2.1 | 2026-10-26 | manager-spec | Delta repair per plan-audit iter2 FAIL 0.82 (report `## iter2`). N1: AC-FW-002's first-block limb scoped to "whenever at least one stage-1 candidate exists" (REQ-FW-006 decision order). N2: B5 stale-selection policy pinned (miss under the resolved id probes the prior-model entry, serve with warning — keep-and-warn); AC-FW-011 method re-staged against that policy. N3: REQ-FW-003 empty-first clause (`empty` precedes passthrough). N4: stale "without a RED cell" list trimmed of AC-FW-010/011 (they cite R8-R10). REQ/AC numbering unchanged; no measured code path touched (RED ledger rows R1-R10 stay valid at pin `4f0e58197d`). |
| 0.2.2 | 2026-10-26 | manager-spec | Delta repair per plan-audit iter3 FAIL 0.87 (report `## iter3`; premise: the N5 staging trace — explicit override + alias-moved state under the unscoped probe falsifies AC-FW-009's first limb). N5: B5 stale-selection probe scoped with an applicability condition (probe applies only to alias-resolved lookups; an explicit model override that misses goes to the network and writes its own entry, never serving another model's stale entry); AC-FW-009 method mirrors the same assertion (staged explicit-override scenario yields two distinct cache files, no cross-model stale serve). REQ/AC numbering unchanged; markdown-only edit, RED ledger rows R1-R10 stay valid at pin `4f0e58197d`. |

> Upstream provenance: the pipeline semantics (block kinds, thresholds, statuses, exit codes) are a behavioral port of jevfetch (`jev.py`, `scoring.py`, `pipeline.py`, `blocks.py`). No Python source or dependency enters this repository; the port is Go-only against the existing `internal/jev` client.

## §A — Context and Motivation (WHY)

The repository has two web-reading surfaces today — the built-in WebFetch (a summary model rewrites the page) and, under GLM, the z.ai web reader — and neither returns **verbatim** excerpts. When the task needs "the exact passage that answers this question, unmodified", the harness has no tool: every path paraphrases.

jevfetch solves this with a different shape: split the page into blocks, ask Jev (TypeSafe System One noul yes/no probabilities) which blocks the question is about, and return the kept blocks **word-for-word**, with heading paths. Its bench claims (47% time reduction) are NOT reproduced here (Gap, recorded below); what is ported is the pipeline.

`internal/jev` already ships the client (`jev-latest` alias, interpreted-model validation, 429/529 kind separation), gated by `workflow.jev.enabled` (default **false** — no keys, no spend, zero network while off). This SPEC builds the fetch pipeline on that client, hardens the client's cache/retry surface, extends `internal/jevmeasure` to host jevfetch-style bench cases, and packages the result as templates.

## §B — Scope (WHAT this SPEC delivers)

Four applications, exactly as scoped by the leader analysis:

| # | Application | Surface |
|---|-------------|---------|
| 1 | `moai fetch web <url> "<question>"` CLI + `fetch_web` MCP tool — the Go port of the jevfetch blocks/scoring/pipeline | `internal/fetchweb` (new), `internal/cli`, `internal/cli/mcp_server.go` |
| 2 | `internal/jev` client hardening — **client cache/retry surface ONLY** | `internal/jev` |
| 3 | `internal/jevmeasure` bench case format + baseline.json comparison command | `internal/jevmeasure` |
| 4 | Template packaging — skill, permission rule, routing row, Codex note | `internal/template/templates/**` |

## §C — GEARS Requirements

### §C.1 Pipeline core — block splitting and token accounting

- **REQ-FW-001** (Ubiquitous): The block splitter shall split a fetched HTML page into blocks of exactly these kinds — paragraphs, list items, table rows, and code blocks — with each block carrying its heading path (the sequence of ancestor headings), page chrome (navigation, footers, scripts, styles, templates, and similar non-content furniture) removed before splitting, and `<aside>` note content emitted as blocks that are scored separately from the main flow.
- **REQ-FW-002** (Ubiquitous): The pipeline shall count tokens with the approximate estimator already shipped in `internal/jev` (`EstimateTokens`), shall add no tokenizer dependency of any kind, and every threshold comparison that depends on the count shall be documented at its use site as approximate (the upstream reference counts with tiktoken cl100k; the Go estimator is deliberately conservative and the two may diverge).
- **REQ-FW-003** (State-driven): **While** a document's total token count after splitting is 2,000 or fewer, the pipeline shall pass the whole document through unfiltered, report status `passthrough`, and make no scoring request at all — with the `empty` status decision (a page that splits to zero blocks) preceding passthrough, so this guarantee applies only to non-empty documents.

### §C.2 Two-stage Jev scoring and selection

- **REQ-FW-004** (Ubiquitous): Stage 1 of the scoring pipeline shall present blocks to Jev in windows of approximately 3,000 tokens and shall mark as a candidate every block whose relevance probability against the question is 0.6 or greater.
- **REQ-FW-005** (Ubiquitous): Stage 2 shall re-examine each candidate standalone **and** with a neighbor context of approximately 400 characters, and shall keep a candidate whose standalone-plus-context probability is 0.8 or greater; a candidate scoring below 0.8 shall still be kept **While** a same-heading neighbor block scores 0.6 or greater.
- **REQ-FW-006** (Ubiquitous): The selection stage shall decide `no_match` first: **When** zero blocks become stage-1 candidates, the pipeline reports `no_match` and no selection guarantee below applies. Once at least one stage-1 candidate exists, the selection stage shall preserve these guarantees: the first block of the document is always kept; **When** the kept set totals fewer than 300 tokens, blocks are added back in descending stage-1 score order until the threshold is met or blocks are exhausted; every kept table row is accompanied by its table's header row; and kept blocks are emitted in original document order with their text **verbatim** — no paraphrase, no truncation, no rewording. (Decision order pinned per plan-audit D1: `no_match` = zero stage-1 candidates; the first-block keep and the 300-token backfill apply only when at least one candidate exists, so `no_match` is reachable on any non-empty page where nothing scores ≥ 0.6 at stage 1.)

### §C.3 Command, output contract, MCP surface

- **REQ-FW-007** (Ubiquitous): The command `moai fetch web <url> "<question>"` shall fetch the target page with a bounded HTTP client (a timeout and a maximum response size), run the pipeline of §C.1-§C.2, and emit: a first line beginning `#` carrying one of the statuses `filtered` / `passthrough` / `no_match` / `empty` / `fetch_error`, then per kept block a line beginning `##` carrying the block's heading path followed by the block's verbatim text; the exit code shall be 0 for `filtered` and `passthrough`, 1 for `no_match` and `empty`, and 2 for `fetch_error`; and a `--json` flag shall emit the same information as a machine-readable object instead of the markdown-shaped text.
- **REQ-FW-008** (State-driven): **While** `workflow.jev.enabled` is `false` (the distributed default), the command shall make **zero network requests** — neither the page fetch nor any Jev call — and shall instead emit a WebFetch fallback notice telling the operator to use WebFetch (or enable Jev), exiting with code 2.
- **REQ-FW-009** (Ubiquitous): The MCP catalog shall expose a `fetch_web` tool that runs the same pipeline against the same output contract (status, heading-path-prefixed verbatim blocks, and the same exit-status mapping surfaced as the tool's result shape), so CLI and MCP callers observe identical semantics.
- **REQ-FW-010** (Unwanted): The test suite shall make **no live Jev call and no live target-page fetch**; every Jev interaction in tests shall run against recorded responses or a fake server, and every page-fetch interaction in tests shall run against a fake page server or fixture.

### §C.4 internal/jev client hardening — CLIENT surface only

- **REQ-FW-011** (Ubiquitous): The client shall cache Jev responses on disk under `~/.moai/cache/jev/`, keyed by a SHA of the request body **combined with the interpretation model id** (never the `jev-latest` alias), written by atomic replace so a concurrent reader never observes a partial cache file, and a cache hit shall be served without a network request.
- **REQ-FW-012** (Event-driven): **When** the endpoint responds 429, 529, or another 5xx with a `Retry-After` header, and the failing call satisfies the retry conditions REQ-JEVC-008 already permits (only calls whose repetition is provably free of side effects), the client shall parse the header and delay the retry by the parsed amount (within a bounded maximum and a bounded retry count); **When** a request is non-repeatable or its failure is ambiguous after delivery, the client shall not retry it (composing with `SPEC-JEV-CORE-001` REQ-JEVC-008 and its existing `TestRetry_*` safeguards, not overriding them); and the client shall cap concurrently in-flight requests with a semaphore.
- **REQ-FW-013** (Ubiquitous): The client's usage aggregation shall separate new network requests from cache-served answers, and **When** a cache entry was written against an interpretation model id that the alias has since moved past, the client shall emit a staleness warning naming the cached model id.

### §C.5 Bench — internal/jevmeasure

- **REQ-FW-014** (Ubiquitous): `internal/jevmeasure` shall accept bench cases in the jevfetch shape — a question, an expected quote, and a page snapshot — and shall provide a comparison command that, given a run result and a `baseline.json`, reports per case whether the expected quote is contained in the output and the token-savings rate versus the page snapshot's token count, plus a summary over the case set.

### §C.6 Template packaging

- **REQ-FW-015** (Capability gate): **Where** the templates are deployed, this SPEC's shipped artifacts shall be: the skill `moai-fetch-web` (whose SKILL.md states when to prefer it over WebFetch, and how `--json` output and the `no_match` status are to be handled), the settings permission rule `Bash(moai fetch web:*)`, a row in the GLM web-tooling routing rule stating "jev active → `fetch_web` first", and a Codex `network_access` note for the command — all template-neutral: no machine-specific values, and no `enabled: true` introduced by THIS SPEC's own shipped artifacts (in particular not `workflow.jev.enabled`). The template tree already carries 127 `enabled: true` occurrences owned by other templates (measured 2026-10-26, this tree); this SPEC removes none of them, and the existing template-neutrality CI guard stays green.

### §C.7 Boundary against the concurrent ledger/consumers card

- **REQ-FW-016** (Unwanted): This SPEC's changes inside `internal/jev` shall be confined to the **client** surface — the disk cache, the Retry-After backoff, the concurrency semaphore, the usage new-vs-cache separation, and the alias-staleness warning — and `internal/jev` has no other concurrent claimant: the ledger/consumers work of in-flight card **t1762** (`SPEC-JEV-LEDGER-001`) lives in its own new package `internal/jevledger` plus its own `internal/config` and `internal/template` surfaces and does not edit `internal/jev` (verified against t1762's SPEC text at commit `954adab25e`). This SPEC's implementation shall not modify, rename, or restructure `internal/jevledger` or t1762's config/template artifacts; if a needed seam is discovered there, the implementation returns a blocker rather than editing across the boundary.

## §D — Constraints

1. **No new API keys, no spend.** `workflow.jev.enabled` stays default-`false`; the capability ships INERT (same family as the existing gate — while off, the package constructs no request and makes no network call). Operator keys/credits are the only activation path.
2. **No live network in tests.** Recorded responses or fake servers only (REQ-FW-010); a test that would reach the real Jev endpoint or a real target page is a defect.
3. **No Python anything.** No httpx/bs4/tiktoken, no vendored Python, no Python toolchain step. HTML parsing and token counting are Go-only; token counting reuses `internal/jev.EstimateTokens` and is **approximate** (Constraint of record, upstream Gap).
4. **No headless-Chrome auto-detection or render fallback.** Excluded by the leader analysis (execution-environment dependence; install-prohibition principle). Deferred to a follow-up card with t1965 linkage.
5. **Naming is fixed by operator directive (2026-10-26).** Command `moai fetch web`, MCP tool `fetch_web`, skill `moai-fetch-web`, permission `Bash(moai fetch web:*)`. `moai web` is the existing web-UI command; the sub-verb collision is why the rename happened — do not reintroduce a `web` read verb.
6. **Exit-code and status compatibility with upstream jevfetch** is part of the port contract (REQ-FW-007); the off-gate behavior (REQ-FW-008) is a moai addition and is pinned in plan.md §B-D1.
7. **Template neutrality** (§25 family): everything in application 4 ships as templates with no machine-specific values; the template-neutrality CI guard must stay green.
8. **Upstream Gaps carried, not resolved:** the 47%-time-reduction bench figure is NOT reproduced or asserted anywhere; the ~$0.0015/call cost estimate is informational only; the token-count divergence vs tiktoken cl100k is accepted as approximate.

## §E — Exclusions (What this SPEC does NOT build)

### Out of Scope — Headless-render fallback
- No headless-Chrome detection, no render-then-retry for short static HTML (upstream's <200-token static-HTML retry path). Deferred: follows the t1965 aside/platform-browser linkage, then a follow-up card.
- No new browser dependency of any kind.

### Out of Scope — Python runtime and dependencies
- No httpx, bs4, tiktoken, or any Python packaging; the port is a behavioral reimplementation in Go.

### Out of Scope — jev ledger / consumers surfaces
- The ledger/consumers work of in-flight card t1762 (`SPEC-JEV-LEDGER-001`, another lane) lives in its own new package `internal/jevledger` plus its `internal/config` and `internal/template` surfaces — none of it inside `internal/jev`. This SPEC does not touch those paths (REQ-FW-016); if a needed seam is discovered there, the implementation returns a blocker rather than editing across the boundary.

### Out of Scope — Bench reproduction of upstream numbers
- The 47% time-reduction claim is not verified, not asserted in tests, and not cited as an expected value. The bench harness ships; upstream's numeric results do not.

### Out of Scope — WebFetch deny enforcement
- The optional "deny WebFetch" guidance in the skill is documentation only; this SPEC adds no hook, no settings enforcement, and no behavior change to WebFetch itself.

## §F — Cross-references

- `SPEC-JEV-CORE-001` — the `internal/jev` package, the `workflow.jev.enabled` key, the credential reader (dependency).
- `SPEC-JEV-OPTIN-MEASURE-001` / `SPEC-JEV-CONSUMERS-001` — measurement gate and consumer discipline this SPEC's bench work extends.
- `SPEC-JEV-LEDGER-001` — the ledger/consumers card (t1762) whose `internal/jevledger` / config / template surfaces this SPEC's boundary guard protects (REQ-FW-016).
- Leader analysis (authoritative scope source): `.moai/reports/leader-plan-20261023/jevfetch-analysis-20261026.md`.
- Upstream reference (design only): `https://github.com/a2sys-ai/jevfetch`.
