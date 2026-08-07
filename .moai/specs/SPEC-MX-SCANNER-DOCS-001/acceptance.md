# acceptance.md — SPEC-MX-SCANNER-DOCS-001

## §D. Acceptance Criteria Matrix

Each criterion is binary-testable. Severity: MUST-PASS for P1/P2 (gating), SHOULD for P3 (advisory).

### AC-MSD-001 — rotRisk documented accurately (MUST-PASS, traces REQ-MSD-001)

**Given** the new docs-site page in any of the 4 locales,
**When** a reader reads the rotRisk subsection,
**Then** the prose states ALL THREE facts: (a) `rotRisk` is a field on `@MX:DEBT` tags only; (b) the value is `"no-trigger"` when `@MX:UPGRADE` is absent and empty (omitted from JSON) when present; (c) an absent `@MX:CEILING` is a quality note, NOT the rot gate.

### AC-MSD-002 — LSP fan-in condition documented accurately (MUST-PASS, traces REQ-MSD-002)

**Given** the new docs-site page in any locale,
**When** a reader reads the LSP fan-in subsection,
**Then** the prose states ALL FOUR facts: (a) fan-in is LSP-primary via `textDocument/references`; (b) `fan_in_method` discloses the engine (`"lsp"` or `"textual"`); (c) the default (non-strict) mode silently falls back to textual grep on LSP unavailability or error; (d) `MOAI_MX_QUERY_STRICT=1` raises `LSPRequiredError` instead of falling back.

### AC-MSD-003 — CGO complexity path documented accurately (MUST-PASS, traces REQ-MSD-003)

**Given** the new docs-site page in any locale,
**When** a reader reads the CGO complexity subsection,
**Then** the prose states ALL FOUR facts: (a) measurement is a CGO-gated tree-sitter implementation; (b) a non-CGO build emits `Result{Supported: false}` for every language as a hard stub; (c) on CGO builds, scaffolded languages / files >1 MiB / parse or query errors also yield `Supported: false`; (d) `Supported: false` is a silent skip, never an error.

### AC-MSD-004 — Scan automation timing documented accurately (MUST-PASS, traces REQ-MSD-004)

**Given** the new docs-site page AND the README FAQ in any locale,
**When** a reader reads the scan automation subsection,
**Then** the prose names ALL FIVE lifecycle points: (1) explicit `moai mx scan` CLI; (2) SessionStart deferred cold-start scan (time-boxed ~2s, fail-open); (3) PostToolUse validation hook (reads sidecar, does not rebuild); (4) SessionEnd batch validation hook; (5) sync-workflow enforcement gate (P1/P2 blocking, `--skip-mx` escape).

### AC-MSD-005 — 4-locale docs-site page exists with parity (MUST-PASS, traces REQ-MSD-007)

**Given** the run phase has completed,
**When** the verifier lists `docs-site/content/{ko,en,ja,zh}/advanced/mx-scanner-internals.md`,
**Then** all 4 files exist AND have the SAME SET of H2 headings (same count, same order, same anchors after locale translation) AND the Mermaid diagram source is byte-identical across the 4 files.

### AC-MSD-006 — Cross-reference links added in 4 locales (MUST-PASS, traces REQ-MSD-007)

**Given** the run phase has completed,
**When** the verifier greps `docs-site/content/{ko,en,ja,zh}/advanced/mx-tags.md` for a link to `mx-scanner-internals`,
**Then** all 4 locale files contain the link AND grepping `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-mx.md` likewise yields 4 hits.

### AC-MSD-007 — Mermaid diagrams TD-only (MUST-PASS, traces REQ-MSD-005)

**Given** the new docs-site page in any locale,
**When** the verifier greps the page for `flowchart LR|graph LR|flowchart RL|graph RL`,
**Then** zero matches AND at least one `flowchart TD` or `graph TB` block exists.

### AC-MSD-008 — No body-text emoji (MUST-PASS, traces REQ-MSD-006)

**Given** the new docs-site page in any locale,
**When** the verifier runs the body-emoji scan from `hns-oss-docs-verify`,
**Then** zero decorative body-text emoji AND the `{{</* icon */>}}` shortcode is used for any callout marker (typography arrows `→ ← ↓ ✓ ✗` and branding emoji inside code-block examples are permitted and excluded from the scan).

### AC-MSD-009 — README FAQ 4-file parity (MUST-PASS, traces REQ-MSD-008)

**Given** the run phase has completed,
**When** the verifier compares `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` FAQ sections,
**Then** all 4 files contain FAQ entries for rotRisk, LSP fan-in fallback, CGO complexity stub, and scan automation timing, with heading text and section order matching across the 4 files (canonical = en).

### AC-MSD-010 — No scanner source modified (MUST-PASS, traces REQ-MSD-009)

**Given** the run phase has completed,
**When** the verifier runs `git diff --name-only origin/main...HEAD` and filters for paths under `internal/mx/`, `internal/hook/mx/`, or `internal/cli/mx_scan.go`,
**Then** the filtered list is empty (zero scanner source files touched by this SPEC).

### AC-MSD-011 — URL whitelist respected (MUST-PASS, traces REQ-MSD-010)

**Given** the new docs-site page and README FAQ entries in any locale,
**When** the verifier greps for `docs.moai-ai.dev|adk.moai.com|adk.moai.kr`,
**Then** zero matches (only `adk.mo.ai.kr` is a valid docs-site domain).

### AC-MSD-012 — Menu wiring present (SHOULD, traces REQ-MSD-007)

**Given** the new docs-site page exists in all 4 locales,
**When** the verifier inspects `docs-site/data/menu/main.yaml` and the 4 `content/<locale>/_meta.yaml` files,
**Then** the new page is listed in the sidebar navigation in all 4 locales.

## §D.1 Severity Classification

| AC | Severity | Traces REQ |
|---|---|---|
| AC-MSD-001 | MUST-PASS | REQ-MSD-001 |
| AC-MSD-002 | MUST-PASS | REQ-MSD-002 |
| AC-MSD-003 | MUST-PASS | REQ-MSD-003 |
| AC-MSD-004 | MUST-PASS | REQ-MSD-004 |
| AC-MSD-005 | MUST-PASS | REQ-MSD-007 |
| AC-MSD-006 | MUST-PASS | REQ-MSD-007 |
| AC-MSD-007 | MUST-PASS | REQ-MSD-005 |
| AC-MSD-008 | MUST-PASS | REQ-MSD-006 |
| AC-MSD-009 | MUST-PASS | REQ-MSD-008 |
| AC-MSD-010 | MUST-PASS | REQ-MSD-009 |
| AC-MSD-011 | MUST-PASS | REQ-MSD-010 |
| AC-MSD-012 | SHOULD | REQ-MSD-007 |

## §D.2 Edge Cases

- A locale translation that silently drops the `@MX:CEILING` "quality note, not rot gate" clause would create a factually narrower claim than the canonical. The 4-locale parity check (AC-MSD-005) catches section-heading parity but NOT prose-fact parity; the docs author MUST verify each locale carries all three rotRisk facts (AC-MSD-001 applied per-locale).
- A Mermaid diagram that renders TD in the canonical locale but is "reformatted" to LR in a derived locale (because a translator thought it looked better) would pass visual review but fail AC-MSD-007. The diagram source is treated as code — byte-identical across locales.
- A README FAQ entry that links to a docs-site URL with a forbidden domain (e.g. a stale `adk.moai.kr` link copied from an older README section) would fail AC-MSD-011.

## §D.3 Quality Gate Criteria (Definition of Done)

- All MUST-PASS AC (AC-MSD-001..011) demonstrate PASS with cited verifier output.
- AC-MSD-012 (SHOULD) is PASS or carries an explicit deferral note.
- The `hns-oss-docs-verify` recipe exits 0 (warning-free Hugo build, sitemap exists, URL blacklist clean, Mermaid TD-only, 4-locale file+section parity, README heading parity, body-emoji clean).
- `git diff --name-only origin/main...HEAD` shows ONLY documentation paths (docs-site content, README files, menu/_meta config); zero paths under `internal/`.

## §D.4 Indirect Verification

- Scanner behavior claims (rotRisk values, fan_in_method values, Supported:false triggers) are indirectly verified by the §C research citations in `plan.md` — each claim traces to a source file and line. The docs author does NOT re-run the scanner; the citations are the evidence.
- The "no scanner source modified" gate (AC-MSD-010) is the negative-space verification: the SPEC's documentation claims remain accurate because the underlying code is unchanged.

## §D.5 Forward-Looking Checks

- If a future SPEC changes the rotRisk semantics (e.g. introduces a third value beyond `"no-trigger"`/empty), the docs page MUST be updated in the same PR. The "no scanner source modified" gate applies ONLY to THIS SPEC — it is not a permanent freeze on those paths.
- If the cold-start scan timeout default changes, the docs page's "~2s ceiling" claim must be re-verified. The citation in plan.md §C.4 (`internal/hook/session_start.go:1223` `DefaultSessionStartDriftTimeout`) is the anchor.
