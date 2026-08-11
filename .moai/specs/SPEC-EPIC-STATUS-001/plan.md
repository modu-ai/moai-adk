# Plan — SPEC-EPIC-STATUS-001

> Implementation plan for the Epic Status producer. Derived from spec.md (SSOT for requirements). Order is decision-reversibility-first: the highest-change-likelihood decisions (epic-discovery precedence, JSON shape) lead; the mechanical CLI-registration steps close.

---

## §A. Context (one-paragraph summary)

The Epic Status producer is a read-only CLI (`moai epic status <prefix>`) that derives an epic's milestone progress map from `.moai/specs/SPEC-*/spec.md` frontmatter + optional design report, and emits a banner-SSOT JSON shape. It composes the existing `spec.ListDocs` + `spec.Audit` read path (the same pair `internal/web/board.go:88-131` already composes) and registers a new top-level `moai epic` command family. The producer is the first concrete step toward making Factory Mode a real epic orchestrator; backlog/card/dispatch/quorum surfaces are explicitly out of scope (owned by the KANBAN BOARD / WORKTREE / BOOTSTRAP SPECs).

### Baseline attribution

- **HEAD at plan-phase authoring**: `9fa242ddae3e5c7e9a80c2b47bd03d38b4c1b5ed` (branch `feat/factory-bootstrap-guidance`).
- **ListDocs measured at**: `internal/spec/listdocs.go:36` (function signature, observation-only contract).
- **Audit measured at**: `internal/spec/audit.go:156`.
- **buildBoardView measured at**: `internal/web/board.go:88-131`.
- **Banner templates measured at**: `.claude/output-styles/moai/moai.md:572-610` (Epic Status), `:636-680` (Progress Board).
- **BAS Mx canonical list measured at**: `.moai/reports/navigator-redesign-bas-20260805.html` §7 slice table.
- **`(BAS Mx)` marker convention measured against**: three Navigator-Sync titles in `.moai/specs/SPEC-NAVIGATOR-SYNC-{001,002,003}/spec.md`.

---

## §B. Known Issues + Risks

| ID | Issue / Risk | Mitigation |
|----|----|----|
| KI-1 | Design-report §6/§7 parsing is HTML-fragile — the report format is hand-authored HTML, not a structured schema; a regex parse can break on reformatting | Restrict the parser to a narrow pattern (the `<table>` slice table in §7 with `<tr><td>M<N> ...</td>` rows, plus an `<h2>N. 슬라이스</h2>` anchor). Fail-open: if the regex does not match, treat the design report as unavailable (fall back to marker union, omit `orphan_mx`). |
| KI-2 | The `(BAS Mx)` title-marker convention is currently BAS-specific; other epics may use different tokens (`(EPIC-X M3)`, etc.) | The marker regex is token-generic (`\(([A-Z][A-Z0-9-]*)\s+M(\d+)\)`); the `--marker <token>` flag overrides the default token inference. The default inference is "first marker token seen in the prefix-matched set" (design.md §3 documents the inference rule). |
| KI-3 | `spec.Audit` may be slow on large catalogs (it scans every SPEC dir) | The producer MUST NOT scan the whole catalog when the prefix already narrows the set; however `spec.ListDocs` walks the whole `.moai/specs/` dir. We accept the cost (per `board.go` precedent — the web board already calls both on every render). A future optimization is a prefix-filtered `ListDocs` variant; out of scope here. |
| KI-4 | A SPEC with two `(Mx)` markers in its title (rare, but possible) would ambiguous the Mx→SPEC map | Rule: first marker wins; the second is recorded in `extra_mx` (warning, not error). |
| KI-5 | Worktree in-flight state is invisible (the lateral-fragility hazard in spec.md §A.1) | Acknowledged residual risk; acceptance.md §C records it. Worktree-aware state is owned by `SPEC-KANBAN-WORKTREE-001`. |
| KI-6 | HEAD SHA reading via `git rev-parse HEAD` in a non-git context (sand-boxed test, fresh `t.TempDir()`) returns empty | Fail-open: `baseline_attribution` becomes `""` (empty string); the JSON shape's field is present but empty, NOT omitted (preserves shape stability). |

---

## §C. Pre-flight (zero-code, plan-phase only)

This is a plan-phase commit (markdown-only). Pre-flight verification is documentation-only:

- [x] SPEC ID regex PASS (`SPEC-EPIC-STATUS-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`).
- [x] No collision (no `SPEC-EPIC*` exists at HEAD `9fa242dda`).
- [x] Frontmatter 12-field canonical (spec.md head).
- [x] `phase: "v3.2.0 target"` — NOT a lifecycle-stage name (avoid the `FrontmatterPhaseInvalid` lint).
- [x] §C Out of Scope carries ≥1 `### Out of Scope — <topic>` H3 with bullets.
- [x] ZERO `.go` files in this commit (markdown-only plan-phase; `SKIP_MOAI_PRECOMMIT=1` bypass rationale: "0 .go files — plan-phase markdown only" if the pre-commit gate false-positives).
- [x] No push, no PR (user instruction).

---

## §D. Constraints (reiterated from spec.md §D)

- Worktree-anchored to `/Users/goos/.moai/worktrees/kanban`; no primary-checkout writes.
- Read-only at runtime (observation-only contract).
- Non-interactive (zero `AskUserQuestion`).
- No new external Go dependencies.
- Baseline-attributed (`baseline_attribution` HEAD SHA in every `--json` output).
- Template-neutrality: the SPEC's artifacts are NOT mirrored to `internal/template/templates/` (no CI guard triggered).

---

## §E. Self-Verification (Plan-phase)

Plan-phase self-check (per spec-frontmatter-schema.md progress.md §E.1):

- [x] Epic-discovery precedence documented in spec.md §A.5 + design.md §3 (3-strategy chain with strict order).
- [x] JSON shape frozen in spec.md §B.1 + design.md §5 (additive-only forward-compat).
- [x] AC ↔ REQ traceability matrix present in acceptance.md §B.
- [x] Out-of-scope boundary to KANBAN siblings explicit in spec.md §C.
- [x] Reuse-don't-fork: `spec.ListDocs` + `spec.Audit` cited with file:line in spec.md §A.2 + REQ-ES-002.
- [x] Template-neutrality constraint recorded in REQ-ES-011 + §C.
- [x] Baseline-attribution HEAD SHA `9fa242ddae3e5c7e9a80c2b47bd03d38b4c1b5ed` cited throughout.

---

## §F. Milestones (priority-ordered; no time estimates)

### M0 — Library skeleton + discovery primitives (Priority High)

`internal/epic/` new package. Pure functions, no CLI wiring yet. Outputs: `discover.go`, `status.go`, `status_test.go`.

- `DiscoverEpic(prefix string, opts Options) (*EpicCandidates, error)` — prefix glob over `spec.ListDocs` result; returns matched + unmatched split.
- `ExtractMx(records []spec.DocRecord, token string) (map[string]string, error)` — title-regex extraction; returns Mx→SPEC-ID map.
- `JoinStatus(records []spec.DocRecord, audit *spec.AuditResult, mxMap map[string]string) ([]MilestoneEntry, error)` — per-Mx status join (REQ-ES-006).
- Table-driven tests over a fixture mimicking the BAS epic (3 covered Mx + 2 orphan + 1 untracked).

AC binding: AC-ES-001..AC-ES-004.

### M1 — Design-report parser (Priority High)

`internal/epic/designreport.go` + `_test.go`. Parses `.moai/reports/<basename>-<epic-token-lower>-*.html` §7 slice table. Fail-open on miss (KI-1).

- `ParseDesignReport(path string) (*CanonicalMilestones, error)` — regex over the slice table; returns Mx list + labels.
- `DiscoverDesignReport(epicToken string, reportsDir string) (string, error)` — naming-rule discovery.
- Tests against the actual `.moai/reports/navigator-redesign-bas-20260805.html` fixture (read-only; no fixture duplication).

AC binding: AC-ES-005.

### M2 — Status join + orphan detection (Priority High)

Wires M0 + M1 into the final `EpicStatus` struct. Computes `done` / `total` / `pct` / `orphan_mx` / `extra_mx` / `untracked_specs`. Handles `total: 0` divide-by-zero (REQ-ES-013 pct rule).

AC binding: AC-ES-006, AC-ES-007.

### M3 — JSON + human renderers (Priority Medium)

`internal/epic/render.go` + `_test.go`. JSON matches the frozen shape in spec.md §B.1. Human mirrors Progress Board grammar (`🎯 <epic> ▓▓░░ N/M (pct%)`).

- `RenderJSON(s *EpicStatus) ([]byte, error)` — `json.MarshalIndent` stable ordering.
- `RenderHuman(s *EpicStatus, locale string) (string, error)` — Progress Board grammar translation.
- Tests assert byte-stable JSON for a fixed fixture (regression guard).

AC binding: AC-ES-008, AC-ES-009.

### M4 — CLI verb registration (Priority Medium)

`internal/cli/epic.go` + `epic_test.go`. Registers `moai epic status <prefix>` under a new `moai epic` parent (REQ-ES-009). Flags: `--json`, `--design-report`, `--marker`. Non-interactive (REQ-ES-010).

- `newEpicCmd() *cobra.Command` — parent.
- `newEpicStatusCmd() *cobra.Command` — status subcommand.
- `rootCmd.AddCommand(newEpicCmd())` — registration in `init()` or in `internal/cli/root.go` next to the `newSpecCmd()` group (follow the established pattern at `internal/cli/spec.go:10-39`).
- Integration tests via `cmd := newEpicStatusCmd(); cmd.SetArgs([]string{...})` (mirrors `spec_audit_test.go:124`).

AC binding: AC-ES-010, AC-ES-011.

### M5 — Baseline attribution + factory touchpoint documentation (Priority Low)

- `baseline_attribution` HEAD SHA reading via `git rev-parse HEAD` (fail-open to empty string in non-git contexts — KI-6).
- Docs-site pointer: a paragraph in the factory-mode docs page noting `moai epic status` as a visibility primitive (factory touchpoint per REQ-ES-012; this is a doc-only deliverable, NOT a code wire-up — the wire-up is owned by the Factory/Kanban Bootstrap SPEC family).

AC binding: AC-ES-012.

### M6 — Sync-phase close (owned by manager-docs, NOT this plan)

3-phase close: this plan covers M0-M5 (run-phase). manager-docs owns the sync commit that carries `implemented → completed` + this plan.md's `sync_commit_sha` backfill, per the Status Transition Ownership Matrix.

---

## §G. Anti-Patterns (do NOT)

- **AP-ES-001**: DO NOT fork `spec.ListDocs` into a new scanner. The producer composes the existing function; if a prefix-filtered variant becomes necessary, add it to `internal/spec/` (a separate enhancement SPEC), do NOT duplicate the walk logic.
- **AP-ES-002**: DO NOT auto-emit the Epic Status banner from inside the producer. The producer prints its human-mode output (Progress Board grammar) but does NOT call into `.claude/output-styles/`. Banner wiring is a separate concern (spec.md §C).
- **AP-ES-003**: DO NOT route through `nav-graph.json`. The navigator graph does not exist on disk at HEAD `9fa242dda` and its schema has no epic/milestone node type (spec.md §C).
- **AP-ES-004**: DO NOT introduce `epic.json` or any persisted store. The producer is derived on-demand; persistence is the KANBAN BOARD SPEC's territory (REQ-ES-013).
- **AP-ES-005**: DO NOT introduce interactive prompts in the CLI. `moai epic status` is read + print + exit (REQ-ES-010).
- **AP-ES-006**: DO NOT omit `baseline_attribution`. Every progress claim in `--json` carries the HEAD SHA per `verification-claim-integrity.md` §2.

---

## §H. Cross-References

- spec.md §A.5 — epic-discovery precedence (the locked design decision).
- spec.md §B.1 — JSON shape contract (frozen, additive-only).
- acceptance.md §B — AC ↔ REQ traceability matrix.
- design.md §3 — epic-discovery 3-strategy chain (detailed rationale).
- design.md §4 — design-report discovery rule.
- design.md §5 — JSON shape rationale + additive-only contract.
- research.md §2 — `spec.ListDocs` + `spec.Audit` + `buildBoardView` measured prior art.
- `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/plan.md` — sibling plan; follows the same Tier L 6-artifact structure.
