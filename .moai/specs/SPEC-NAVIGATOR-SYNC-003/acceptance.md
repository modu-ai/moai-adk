# SPEC-NAVIGATOR-SYNC-003 — Acceptance Criteria

> Tier L acceptance artifact. The verification layer: each AC is `AC-NS3-NNN`, written Given-When-Then, binary-testable.
> GEARS lives in the REQUIREMENT layer (spec.md §C REQ-NS3-NNN); this file does NOT restate requirements.

## §A. Test Methodology

### §A.1 Non-overlap grep-pair protocol

The non-overlap ACs (AC-NS3-016 / 017 / 018) use the two-lens grep-pair protocol (per `feedback_deadfield_grep_pair_protocol.md`, carried forward from M0 design.md §5):

- **Lens 1 — source grep**: the production source under `internal/navigator/tiers/` + `internal/cli/navigator_tiers.go` MUST NOT literally name a forbidden path. A source file that mentions a forbidden path in a comment trips Lens 1 even though the binary never touches it. Source-comment hygiene: forbidden LSEL path fragments (`lessons-inbox`, `state/lsel`, `memory/feedback_`, `hns-lsel`) and forbidden predecessor write paths (`capability-map.md`, `audit-report`, `capability-symbols`) MUST NOT appear as literal strings in production source comments.
- **Lens 2 — runtime fixture**: a temp-dir fixture runs the tier enrichment, snapshots the tree before and after, diffs, and asserts the only NEW paths are within the 6 allowed write surfaces (REQ-NS3-018). A bare `lsel` substring is FORBIDDEN as a grep pattern (it would match the SPEC's own non-overlap test file).

### §A.2 Attribution obligation (per `verification-claim-integrity.md` §2)

Every AC's Evidence names the command + the observable output form. A PASS is valid only while command + observed output remain attributable. The success metric (AC-NS3-022) additionally binds §1.1 surface 3 (no narrative-claim substitution).

## §B. AC Matrix (trace to REQ)

| AC | REQ | One-line |
|----|-----|----------|
| AC-NS3-001 | REQ-NS3-001 | contract node additive |
| AC-NS3-002 | REQ-NS3-002 | contract drift build-enforced, graph fail-open |
| AC-NS3-003 | REQ-NS3-003 | 3 contract surfaces + empty-registry degrade |
| AC-NS3-004 | REQ-NS3-004 | module_tree.json authored scaffold |
| AC-NS3-005 | REQ-NS3-005 | overview.md Kiro 7 sections |
| AC-NS3-006 | REQ-NS3-006 | blueprint drift is debt, NOT build failure |
| AC-NS3-007 | REQ-NS3-007 | blueprint node + module/owns edges additive |
| AC-NS3-008 | REQ-NS3-008 | ADR 4 fields + missing-ADR degrade |
| AC-NS3-009 | REQ-NS3-009 | supersede: new ADR + Status flip, body immutable |
| AC-NS3-010 | REQ-NS3-010 | decision node enriched in overlay |
| AC-NS3-011 | REQ-NS3-011 | per-symbol signature + declaration + references |
| AC-NS3-012 | REQ-NS3-012 | narrative metadata.json last_updated_commit gate |
| AC-NS3-013 | REQ-NS3-013 | Go astx path; SCIP absent → graceful |
| AC-NS3-014 | REQ-NS3-014 | symbol node enriched in overlay |
| AC-NS3-015 | REQ-NS3-015 | 2-tier separable; deterministic without LLM |
| AC-NS3-016 | REQ-NS3-016 | consumer-only on M0/M1 producers |
| AC-NS3-017 | REQ-NS3-017 | no predecessor + no LSEL surface |
| AC-NS3-018 | REQ-NS3-018 | write-surface isolation; nav-graph never overwritten |
| AC-NS3-019 | REQ-NS3-019 | provenance no wall-clock; byte-identical re-run |
| AC-NS3-020 | REQ-NS3-020 | fail-open each error mode |
| AC-NS3-021 | REQ-NS3-021 | template-first + neutrality green |
| AC-NS3-022 | REQ-NS3-022 | ≥40% reads-reduction via named fixture |

## §C. Edge cases + quality gates

- **Edge — `nav-graph.json` absent (M0 not yet run)**: tier enrichment fail-opens (AC-NS3-020); emits no `tiers.json`; exit 0; one log line.
- **Edge — ADR file missing for a `@NAV:DEC-<id>`**: decision node remains with `adr_path` empty (AC-NS3-008); logged at debug, not an error.
- **Edge — symbol with no narrative**: structure-only enrichment (AC-NS3-014); `narrative_path` empty.
- **Edge — non-Go language (SCIP absent)**: deterministic path degrades to "no per-symbol structure for this language" (AC-NS3-013); does NOT block.
- **Edge — human-edited `module_tree.json`**: a plain `/moai project` run does NOT overwrite it (AC-NS3-004); `--rescaffold` opt-in required.
- **Edge — prior ADR superseded**: prior body byte-identical except the `Status:` line (AC-NS3-009).
- **Quality gate**: `go test ./internal/navigator/tiers/...` ≥85% coverage; `golangci-lint` 0 issues; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `make build` regenerates catalog; `internal/template/internal_content_leak_test.go` green.

## §D. Given-When-Then Scenarios

### AC-NS3-001 (contract node additive)
**Given** a fixture `nav-graph.json` with decision/spec/symbol nodes, **When** the tier enrichment runs with a `contracts.yaml` registering one schema contract, **Then** `tiers.json` contains a `contract` entity-type node AND the original three entity types remain present and unchanged in shape.

### AC-NS3-002 (contract drift build-enforced, graph fail-open)
**Given** a registered contract whose declaration and implementation diverge, **When** the drift check runs in CI mode (`TIER_CONTRACT_CI=1`), **Then** the drift check exits non-zero AND `tiers.json` is still emitted (graph fail-open) AND a diagnostic line is appended to `.moai/logs/navigator-sync.log`.

### AC-NS3-003 (3 contract surfaces + empty-registry degrade)
**Given** the default moai-adk-go configuration, **When** contract enumeration runs, **Then** the three surfaces (nav-graph schema, hook JSON schemas, CLI flag schemas) are recognized; **and given** an empty `contracts.yaml`, **then** zero contract nodes are emitted and the run exits 0.

### AC-NS3-004 (module_tree.json authored scaffold)
**Given** a `.moai/project/codemaps/dependencies.md` and NO existing `module_tree.json`, **When** the tier enrichment runs, **Then** a draft `module_tree.json` is scaffolded; **and given** a human-edited `module_tree.json`, **when** a plain run executes (no `--rescaffold`), **then** the file is NOT overwritten (byte-identical).

### AC-NS3-005 (overview.md Kiro 7 sections)
**Given** a `module_tree.json` entry, **When** the overview template is instantiated, **Then** the generated `overview.md` carries all seven Kiro sections as headings (Component Architecture, Data Flow, Data Model, Error Handling, Test Strategy, Implementation Approach, Migration) AND a provenance block with `last_updated_commit`.

### AC-NS3-006 (blueprint drift is debt, NOT build failure)
**Given** a `module_tree.json` whose `depends_on` diverges from the actual code's import graph, **When** `/moai project` runs, **Then** the run exits 0 (no build failure) AND no test fails AND a documentation-debt signal is emitted to the M1 detect channel (`.moai/state/navigator-detect/`). Lens: `grep -rn "blueprint.*drift.*fail\|drift.*Fail\|t.Fatal.*blueprint" internal/` returns 0 matches in test files (negative test — the drift-fail path does not exist).

### AC-NS3-007 (blueprint node + module/owns edges additive)
**Given** a `module_tree.json` with two modules where A depends_on B, **When** the tier enrichment runs, **Then** `tiers.json` contains two `blueprint` nodes, one `module-edge` (A→B), and `owns-edge` entries joining each blueprint to the symbols extracted within it; AND M0's existing edge types are present and unchanged.

### AC-NS3-008 (ADR 4 fields + missing-ADR degrade)
**Given** a `@NAV:DEC-AUTH-STRATEGY` token and a `.moai/decisions/AUTH-STRATEGY.md` with the four fields, **When** ADR resolution runs, **Then** the decision node's `adr_path` points to the file AND the four fields are parseable; **and given** a `@NAV:DEC-ORPHAN` token with no ADR file, **then** `adr_path` is empty and the node remains (degrade, no error).

### AC-NS3-009 (supersede: new ADR + Status flip, body immutable)
**Given** an existing ADR `AUTH-STRATEGY.md` with Status `Accepted`, **When** a supersede operation creates `AUTH-STRATEGY-V2.md` with `supersedes: AUTH-STRATEGY`, **Then** `AUTH-STRATEGY.md`'s body is byte-identical to its pre-supersede content EXCEPT the `Status:` line which reads `Superseded`, AND `tiers.json` carries a `superseded_by` edge from `AUTH-STRATEGY` to `AUTH-STRATEGY-V2`.

### AC-NS3-010 (decision node enriched in overlay)
**Given** M0 decision nodes in `nav-graph.json`, **When** the tier enrichment runs, **Then** `tiers.json` carries `adr_path` and `superseded_by` fields per decision node; AND `nav-graph.json` itself is byte-identical before and after (M0 owns it).

### AC-NS3-011 (per-symbol signature + declaration + references)
**Given** a Go fixture package with a function `ParseHeader`, **When** the astx-extended structure extraction runs, **Then** the per-symbol record for `ParseHeader` carries a `signature`, a `declaration_path` + `declaration_line`, and a `references` list with ≥1 caller location.

### AC-NS3-012 (narrative metadata.json last_updated_commit gate)
**Given** a symbol narrative file with `metadata.json` `last_updated_commit: <sha-old>` and a deterministic record unchanged since `<sha-old>`, **When** a narrative refresh runs, **Then** the narrative is NOT re-drafted (gate holds); **and given** the deterministic record changed at `<sha-new>`, **then** the narrative IS re-drafted and `metadata.json` updates to `<sha-new>`.

### AC-NS3-013 (Go astx path; SCIP absent → graceful)
**Given** the M4 configuration with no SCIP indexer vendored, **When** symbol extraction runs for a Go file, **Then** the astx/Go path produces structured records; **and when** extraction is requested for a non-Go language with no SCIP configured, **then** the layer emits zero per-symbol records for that language, logs a debug line, and exits 0 (graceful degrade, no SCIP dependency).

### AC-NS3-014 (symbol node enriched in overlay)
**Given** M0 symbol nodes, **When** the tier enrichment runs, **Then** `tiers.json` carries `signature`, `declaration_path`, `declaration_line`, `references` (capped at N), and `narrative_path` (empty when no narrative) per symbol node.

### AC-NS3-015 (2-tier separable; deterministic without LLM)
**Given** the deterministic-layer-only configuration (LLM narrative path stubbed/disabled), **When** the tier enrichment runs, **Then** `tiers.json` is still emitted with the deterministic structure fields populated AND narrative_path fields empty AND the run exits 0 (deterministic layer is NOT blocked by LLM availability).

### AC-NS3-016 (consumer-only on M0/M1 producers)
**Given** the M4 run-phase commit range, **When** Lens 1 `git diff <range> -- internal/navigator/sync/ internal/navigator/detect/ internal/hook/navigator_detect.go internal/hook/navigator_detect_nonoverlap_test.go internal/hook/post_tool.go` runs, **Then** the diff is empty (M0/M1 producers untouched); Lens 2 grep: `internal/navigator/tiers/` source imports `internal/navigator/sync` and `internal/navigator/astx` as read-only consumers (no Write to those packages' output paths).

### AC-NS3-017 (no predecessor + no LSEL surface)
**Given** the production source under `internal/navigator/tiers/`, **When** Lens 1 source-grep runs for forbidden path fragments (`capability-map.md`, `audit-report`, `capability-symbols`, `lessons-inbox`, `state/lsel`, `memory/feedback_`, `hns-lsel`), **Then** 0 matches in non-test source; Lens 2 runtime fixture asserts the temp-dir diff writes none of those paths.

### AC-NS3-018 (write-surface isolation; nav-graph never overwritten)
**Given** a temp-dir fixture with a pre-existing `nav-graph.json`, **When** the tier enrichment runs, **Then** the only NEW paths are within the 6 allowed surfaces (`blueprint/module_tree.json`, `blueprint/<module>/overview.md`, `blueprint/contracts.yaml`, `decisions/<dec-id>.md` new only, `navigator/tiers.json`, `navigator/symbols/<symbol>.md`) AND `nav-graph.json`'s content hash is identical before and after (overlay, not overwrite).

### AC-NS3-019 (provenance no wall-clock; byte-identical re-run)
**Given** the tier enrichment run twice on the same HEAD with no input change, **When** the two `tiers.json` outputs are compared, **Then** they are byte-identical; AND the `captured_at` field equals `git log -1 --format=%cI` of HEAD (no `time.Now()`); AND `grep -rn "time.Now()" internal/navigator/tiers/` excluding narrative-drafting code returns 0 matches in the deterministic path.

### AC-NS3-020 (fail-open each error mode)
**Given** each error mode in turn — (a) `nav-graph.json` absent, (b) `tiers.json` input unparseable, (c) astx extraction error, (d) narrative file absent, (e) per-tier timeout — **When** the tier enrichment runs, **Then** each yields exit 0, no tier enrichment for the affected node, no user-facing error, and one diagnostic line in `.moai/logs/navigator-sync.log`.

### AC-NS3-021 (template-first + neutrality green)
**Given** the 5 author-facing template surfaces authored under `internal/template/templates/`, **When** `make build` runs and `go test ./internal/template/ -run TestInternalContentLeak` executes, **Then** both exit 0; AND `grep -nE 'SPEC-NAVIGATOR-SYNC|REQ-NS3|AC-NS3|2026-08-06'` on each template file returns 0 matches (neutrality); AND the local mirrors under `.moai/project/blueprint/`, `.moai/decisions/adr-template.md`, `.moai/project/navigator/symbols/` are byte-identical to their template sources.

### AC-NS3-022 (≥40% reads-reduction via named fixture)

**Comparator methodology — strategy-proof by construction (REQ-NS3-015 separability).** The ≥40% delta is produced by a DETERMINISTIC file-read simulator with no LLM in the loop (a fixed procedure, honoring the 2-tier separability so the metric never depends on an LLM's reading choices). (1) **Read definition** — one "read" = one file opened and fully loaded into the simulator's context (file-granular, NOT per-chunk; the atomic unit is the file because the blueprint's value proposition is reducing the COUNT of files an orienting agent must open, and per-chunk would conflate the metric with token-budgeting). (2) **Reading agent** — a fixed scripted simulator parameterized by an orientation question, NOT an LLM. (3) **Reading strategy** — for each question the simulator consults the blueprint layer first (`module_tree.json` → follow `depends_on` edges → each referenced `overview.md`) in dependency-order, then — only if the answer is absent from the blueprint layer — descends into source files in filepath-sorted order; the "without-blueprint" run skips the blueprint layer and starts directly at source files in the same filepath-sorted order. (4) **Termination condition** — the simulator stops when it loads a file containing the question's target anchor (a fixed keyword/symbol/identifier match, NOT an LLM relevance judgment) OR a fixed file cap (30 files) is reached, whichever comes first. The "with-blueprint" and "without-blueprint" runs share the IDENTICAL agent, strategy, and termination, differing ONLY in blueprint-layer availability — so the reads-count delta isolates the blueprint's orientation value and cannot be inflated by a baseline-choice artifact; the anti-hardcoding guard below additionally forbids asserting a constant `P` without computing it.

**Given** the static orientation-task corpus at `internal/navigator/tiers/testdata/reads-corpus/` (N tasks, each a (repo-state, orientation-question) pair) and the deterministic read-count comparator, **When** the comparator runs the corpus twice — once with blueprint pre-read enabled, once without — **Then** the emitted percentage `P = (without_reads - with_reads) / without_reads * 100` is the observed output (NOT a hardcoded constant) AND `P ≥ 40`. The acceptance reads the observed `P` from the test output; a test that asserts a hardcoded `P` value without computing it is a narrative-claim violation (AP-NS3-008).
