# Acceptance — SPEC-PROJECT-NAVIGATOR-002

> Each AC is a binary-testable Given-When-Then scenario. REQs live in `spec.md` §C; this file is the verification layer. Severity: M (Milestone-blocking) unless noted. Traceability: AC-NA-XXX ↔ REQ-NA-XXX.

## §D. AC Matrix

### AC-NA-001 — Audit produces report

**Given** a fixture project with design docs (`.moai/project/{product,structure,tech}.md`) containing a `## Core Features` section AND a populated `.moai/project/navigator/capability-map.md` (the 001 output)
**When** `/moai project --audit` is invoked (equivalently: the audit script `scripts/navigator-audit.sh` is run with `CLAUDE_PROJECT_DIR` pointing at the fixture)
**Then** exactly two files exist under `.moai/project/navigator/`: `audit-report.md` AND `audit-report.json` — and no other top-level audit file is present.

### AC-NA-002 — Audit is read-only over its inputs (no regeneration side-effect)

**Given** the fixture project with a populated `capability-map.md` and `progress-map.md`
**When** the audit runs
**Then** the byte content of `.moai/project/navigator/capability-map.md`, `.moai/project/navigator/progress-map.md`, `.moai/project/navigator/navigator.md`, `.moai/project/product.md`, `.moai/project/structure.md`, `.moai/project/tech.md` is byte-identical before and after the audit (verified by `sha256sum` comparison); AND the audit script's process tree shows NO invocation of `navigator-regen.sh`. The capability-map the user sees in the report is exactly what was on disk before the audit ran.

### AC-NA-003 — Missing-SPEC detection

**Given** the fixture's `product.md` names a feature "Real-time Collaboration" under `## Core Features`, AND no capability-map row's title / implementation-path matches "Real-time Collaboration" under the heuristic, AND no override entry covers it
**When** the audit runs
**Then** `audit-report.md`'s `## Missing SPECs` section contains an entry naming "Real-time Collaboration" with its provenance (`.moai/project/product.md` + the heading path), AND `audit-report.json`'s `missing[]` array contains a corresponding object with `design_name`, `source.file`, `source.heading_path`, and `closest_match: null`.

### AC-NA-004 — Orphan-SPEC detection

**Given** the fixture's `capability-map.md` carries a row whose implementation-path column reads `internal/crm` (last path segment `crm`, 3 characters — under the ≥4-character floor of REQ-NA-007(c), so `crm` will NOT produce a `module-token` match), and the row's spec-id is `SPEC-X-999` with title "Legacy CRM Integration" and status `in-progress`, AND no design-doc feature matches "Legacy CRM Integration" or the `crm` path token under the heuristic, AND no override entry covers it, AND its status (`in-progress`) is NOT in the `{superseded, archived, rejected}` excluded set
**When** the audit runs
**Then** `audit-report.md`'s `## Orphan SPECs` section contains an entry naming `SPEC-X-999` / "Legacy CRM Integration" / `internal/crm` with provenance (spec-id + title + implementation-path), AND `audit-report.json`'s `orphan[]` array contains a corresponding object.

### AC-NA-005 — Dual output, stable schema, no extras

**Given** the audit has run
**When** the output files are inspected
**Then** (a) `audit-report.json` parses as valid JSON AND contains exactly the top-level keys `audit_at`, `audit_commit`, `inputs`, `missing`, `orphan`, `matched` — no more, no less; (b) `audit-report.md` contains exactly the three sections `## Missing SPECs`, `## Orphan SPECs`, `## Matched` in that order; (c) no third top-level audit file exists under `.moai/project/navigator/` (no `audit-summary.txt`, no `audit-debug.log`, etc.).

### AC-NA-006 — Provenance per row

**Given** a regenerated `audit-report.md` and `audit-report.json`
**When** each candidate row is inspected
**Then** every Missing SPEC entry carries `source.file` + `source.heading_path` (the design doc that named it); every Orphan SPEC entry carries `spec_id` + `title` + `implementation_path` (the capability-map row it came from); every Matched entry carries `match_basis` ∈ `{exact, substring, module-token, override}`. A grep for rows missing the required provenance field returns zero matches.

### AC-NA-007 — Header-driven token-normalized heuristic tolerates phrasing divergence

**Given** (1) TWO fixture variants of the capability-map, differing ONLY in header spelling: variant-A's header follows 001's spec.md spelling (`capability | owning-spec | status | implementation-path | commit-sha | captured-at`, capability-first) and variant-B's header follows 001's acceptance.md AC-PN-013 spelling (`spec-id | title | status | implementation-path | commit-sha | captured-at`, spec-id-first); (2) `product.md` names "Project Initialization (`moai init`)" and a capability-map row whose title/name column reads "CLI Tool — init template selection" with implementation-path `internal/cli/init`; (3) a second fixture row whose implementation-path is `internal/crm` (last segment `crm`, 3 chars)
**When** the audit runs WITHOUT an override file on each header-spelling variant
**Then** (a) BOTH header-spelling variants parse correctly — header-driven column resolution (REQ-NA-007) recognizes the feature/name and spec-id columns under either spelling, and data rows are extracted successfully (NOT skipped as missing-column warnings); (b) the "Project Initialization" / "CLI Tool — init template selection" pair is considered matched via either `module-token` basis (last path segment `init`, ≥4 chars, appears as a token in the normalized feature name) OR `substring` basis (`init` ≥4 chars appears in both normalized strings), and recorded in `audit-report.md`'s `## Matched` section with the corresponding `match_basis`; (c) the `internal/crm` row's last segment `crm` (3 chars, under the ≥4-char floor) does NOT produce a `module-token` match on the token `crm` alone — verifying the floor rejects trivially short segments. The fixture exercises at least one of each basis: `exact`, `substring`, `module-token`.

### AC-NA-008 — Override file honored

**Given** the fixture's `audit-known-matches.yaml` contains (a) `match: [{design_name: "Autonomy Loop", spec_id: "SPEC-AUTONOMY-WORKFLOW-001"}]` and (b) `ignore: ["SPEC-DEPRECATED-001", "Old Feature Name"]`, AND the heuristic would otherwise emit "Autonomy Loop" as Missing and "SPEC-DEPRECATED-001" as Orphan
**When** the audit runs
**Then** (1) "Autonomy Loop" appears in the `## Matched` section with `match_basis: override`, NOT in `## Missing SPECs`; (2) "SPEC-DEPRECATED-001" appears in NEITHER `## Orphan SPECs` NOR `## Matched`; (3) "Old Feature Name" appears in NEITHER `## Missing SPECs` NOR `## Matched`. The override file is loaded BEFORE the heuristic.

### AC-NA-009 — Idempotence (byte-identical on re-run)

**Given** the fixture at commit SHA `C` with design docs + capability-map + override file unchanged
**When** the audit procedure runs twice in succession with no intervening commit
**Then** `diff` between the two `audit-report.md` outputs is empty AND `diff` between the two `audit-report.json` outputs is empty (byte-identical). The `audit_at` field in JSON is sourced from `git log --format=%cI` for HEAD (NOT wall-clock), so it is stable across re-runs at the same commit.

### AC-NA-010 — Fail-open on missing inputs

**Given** four separate fixture variants: (a) no `.moai/project/*.md` design docs; (b) no `.moai/project/navigator/capability-map.md`; (c) no `.moai/specs/SPEC-*` directories; (d) all three missing (freshly-initialized project)
**When** the audit runs on each variant
**Then** in every case the audit exits code 0, writes a minimal `audit-report.md` containing a literal "no inputs available — <naming the missing input>" placeholder, AND appends a warning line to `.moai/logs/navigator-warnings.log` with a UTC timestamp + the missing-input name. The audit never aborts with a non-zero exit on missing inputs.

### AC-NA-011 — Boundary non-overlap with LSEL + SPEC-003

**Given** the audit script has run on a fully-populated fixture project
**When** the set of files read and written by the audit procedure is enumerated (via a fixture that records every `open()` / redirect target the script touches)
**Then** the read set contains NO LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`) AND no SPEC-003 surface (no tree-sitter grammar directory, no AST extraction helper); AND the write set contains NO LSEL surface AND NO SPEC-003 surface. The write set is exactly `{audit-report.md, audit-report.json}` under `.moai/project/navigator/` plus an append to `.moai/logs/navigator-warnings.log`.

### AC-NA-012 — Template neutrality + 16-language neutrality

**Given** the template-distributed audit surfaces (`moai-workflow-project/SKILL.md` audit-mode section, `references/navigator-audit.md`, `scripts/navigator-audit.sh`, `settings.json.tmpl` if touched)
**When** the CI template-neutrality guard runs (`internal/template/internal_content_leak_test.go` extended with sentinels `SPEC-PROJECT-NAVIGATOR-002` and `REQ-NA-`)
**Then** the guard finds zero matches for: internal SPEC IDs, REQ tokens (`REQ-NA-`), internal dates, commit SHAs (C2/C3/C7 forbidden classes per CLAUDE.local.md §25.1). AND separately: a fixture project whose primary language is NOT Go (e.g. Python-only, with `pyproject.toml` and `.py` sources) — the audit runs successfully without any Go-specific assumption, and the output format is identical to the Go-fixture case modulo the fixture's own design-doc + capability-map content.

## §D.1 Severity classification

- **Milestone-blocking (M)**: AC-NA-001, 002, 003, 004, 005, 009, 010, 011 — core invariants (output shape, read-only, missing + orphan detection, schema, idempotence, fail-open, boundary).
- **Important (I)**: AC-NA-006, 007, 008, 012 — feature completeness (provenance, heuristic tolerance, override, neutrality).

## §D.2 Indirect verification

- REQ-NA-006 provenance: verified by AC-NA-006 (grep for missing-field) AND indirectly by AC-NA-009 (idempotence implies determinism implies the provenance field is sourced consistently).
- REQ-NA-002 read-only: verified by AC-NA-002 across three surfaces (capability-map byte-equal, progress-map byte-equal, no regen subprocess).
- REQ-NA-010 fail-open: verified by AC-NA-010 across four missing-input variants.

## §D.3 Closure gates (Definition of Done)

- All 12 ACs PASS with observed evidence (command + output cited per `.claude/rules/moai/core/verification-claim-integrity.md` §2).
- `make build` clean after template mirror.
- CI template-neutrality guard green (AC-NA-012).
- plan-auditor PASS on this artifact set (M5).
- Implementation Kickoff Approval human gate passed (constraint: PR-mandatory + enforce_admins:true).
- Sync-phase merge of the implementation PR lands the audit feature on `main`.

## §D.4 Forward-looking checks (deferred to follow-up SPECs)

- LSEL consumption of `audit-report.json`'s `missing[]` array → a future SPEC or LSEL amendment.
- SPEC-003 enriched capability-map rows feeding 002's heuristic → SPEC-003's ACs; 002's algorithm is unchanged (it operates on text fields 003 enriches).
- These are listed here so reviewers know the boundary; they are NOT in scope for 002's Definition of Done.
