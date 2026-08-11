# Design — SPEC-EPIC-STATUS-001

> Design rationale for the epic-discovery mechanism, the JSON shape contract, and the design-report parser. Complements spec.md §A.5 (epic-discovery precedence) and §B.1 (JSON shape). This is a Tier L design artifact — it locks the decisions most likely to change before run-phase implementation.

---

## §1. Design questions + locked answers

| Q | Options considered | Locked answer | Rationale |
|---|---|---|---|
| Where does the producer live? | (a) `internal/cli/epic.go` flat, (b) `internal/epic/` package + CLI wiring | **(b)** | The discovery + join + render logic is reusable (a future `moai epic list` and the banner consumption layer both want it); a package boundary keeps the CLI thin. Mirrors `internal/spec/` (library) + `internal/cli/spec_*.go` (wiring) split. |
| Top-level `moai epic` or under `moai spec epic`? | (a) `moai spec epic status <prefix>`, (b) `moai epic status <prefix>` | **(b)** | Epics are a distinct abstraction from SPECs (an epic groups SPECs; a SPEC does not own its epic). The Navigator-Sync family already uses the prefix-as-epic convention without a `spec epic` parent. A top-level `epic` family leaves room for `moai epic list`, `moic epic init` (future), without crowding `moai spec`. |
| Derived epic or persisted epic store? | (a) `epic.json` persisted, (b) derive on-demand | **(b)** | The KANBAN BOARD SPEC owns the persisted card store. Deriving the epic map on-demand avoids a second source of truth and matches `board.go`'s derive-on-render pattern. |
| Design-report as a parser input or a manual flag only? | (a) auto-discover always, (b) `--design-report` flag only, (c) auto-discover with flag override | **(c)** | The user's primary epic (BAS) has its design report at a discoverable path; auto-discovery removes friction. The `--design-report` flag handles non-standard locations and is also the test entry point. |
| Marker token inferred or fixed? | (a) hardcode `BAS`, (b) infer from prefix, (c) require `--marker` | **(b) with override** | The marker convention may generalize beyond BAS; inference (first marker token seen in the prefix-matched set) handles the 90% case; `--marker <token>` is the override. |

---

## §2. Architecture (data flow)

```
   moai epic status <prefix> [--json] [--design-report path] [--marker token]
                │
                ▼
        ┌──────────────────────────────┐
        │   internal/cli/epic.go        │  thin cobra wiring
        │   newEpicStatusCmd()         │  flags → Options
        └──────────────┬───────────────┘
                       │
                       ▼
        ┌──────────────────────────────┐
        │   internal/epic/discover.go  │
        │   DiscoverEpic(prefix, opts) │
        │     ↓ calls                  │
        │   spec.ListDocs(projectRoot) │ ← REUSED (listdocs.go:36)
        │     ↓ prefix-filter          │
        │   EpicCandidates{matched,...}│
        └──────────────┬───────────────┘
                       │
                       ▼
        ┌──────────────────────────────┐
        │   internal/epic/designreport │
        │   DiscoverDesignReport(...)  │
        │   ParseDesignReport(path)    │ ← read-only on .moai/reports/
        └──────────────┬───────────────┘
                       │
                       ▼
        ┌──────────────────────────────┐
        │   internal/epic/status.go    │
        │   ExtractMx(records, token)  │ ← title regex
        │   JoinStatus(records, audit, │ ← spec.Audit (audit.go:156)
        │              mxMap, canon)   │
        │   → EpicStatus               │
        └──────────────┬───────────────┘
                       │
                       ▼
        ┌──────────────────────────────┐
        │   internal/epic/render.go    │
        │   RenderJSON(s)              │ ← frozen shape (spec.md §B.1)
        │   RenderHuman(s, locale)     │ ← Progress Board grammar
        └──────────────────────────────┘
```

The producer never reaches into git internals (the `git rev-parse HEAD` call for `baseline_attribution` is a single stdlib `os/exec` call, fail-open to empty string — KI-6).

---

## §3. Epic-discovery precedence (3-strategy chain, strict order)

The producer computes an `EpicCandidates` set + an `EpicStatus` view-model in three strictly-ordered stages. Each stage's output feeds the next; an earlier stage's signal is NEVER overridden by a later one.

### Stage 1 — Prefix glob (the epic selector)

`DiscoverEpic(prefix, opts)` calls `spec.ListDocs(projectRoot)` once and filters the returned `[]DocRecord` to those whose frontmatter `id` (or directory name fallback) starts with `SPEC-<prefix>-`. The prefix is case-sensitive uppercased (SPEC IDs are uppercase per `internal/spec` `specIDPattern`).

- Empty prefix → CLI flag-parse error ("prefix is required").
- Empty match set → NOT an error (AC-ES-003b); the producer continues with an empty `EpicCandidates.matched` and returns the empty-epic shape.

### Stage 2 — Title-regex Mx extraction (the Mx→SPEC map)

`ExtractMx(matched, token)` scans each `DocRecord.Frontmatter.Title` for the regex `\(([A-Z][A-Z0-9-]*)\s+M(\d+)\)` and builds `map[Mx]SpecID`.

- The `<TOKEN>` is the inferred default = the most-frequent token seen across the matched set (mode), with ties broken by first-seen. The `--marker <token>` flag overrides this (AC-ES-004b).
- A SPEC whose title has no marker → recorded in `EpicStatus.untracked_specs`, NOT silently dropped.
- A SPEC with multiple markers → first wins, remainder in `extra_mx` (KI-4 / edge case E1).

### Stage 3 — Design-report canonical list (optional orphan grounding)

`DiscoverDesignReport(epicToken, reportsDir)` looks for `.moai/reports/<*>-<epic-token-lower>-*.html` (case-insensitive). The naming rule is documented in §4 below. If found, `ParseDesignReport(path)` extracts the M0..Mx canonical list + labels.

- The canonical list is the ONLY basis for `orphan_mx` (REQ-ES-005). Without a design report, `orphan_mx` is omitted (omit-when-empty, per REQ-ES-005 + §5 JSON shape).
- Orphan = canonical Mx NOT covered by any SPEC's marker.
- Extra = SPEC's marker Mx NOT in the canonical list.

---

## §4. Design-report discovery rule

The auto-discovery rule for the design report:

```
scan <reportsDir> (default: .moai/reports/)
  for files matching the pattern:
    regexp: ^.*-[<token-lowercased>]-[a-z0-9]+\.html$
  return the lexicographically-first match (deterministic)
```

For the BAS epic (`token = BAS`), the rule matches `navigator-redesign-bas-20260805.html` (verified at HEAD `9fa242dda` — file exists). The rule is fail-open: zero matches → no canonical list → `orphan_mx` omitted.

The `--design-report <path>` flag bypasses discovery (the user points at the file explicitly). The path is validated against a path-traversal allowlist: it MUST resolve under `.moai/reports/` (resolved absolute path starts with the project's `.moai/reports/` absolute path), OR be an absolute path the caller passed explicitly (which the caller is presumed to have validated). Relative paths escaping `.moai/reports/` are rejected with a flag-parse error.

The design-report parser is HTML-narrow (KI-1): it scans for the slice-table `<table>` block following an `<h2>N. 슬라이스</h2>` heading (the heading text is locale-agnostic in practice — the BAS report uses Korean `슬라이스`, but the parser keys on the `<table>` structure with `<tr><td>M<N> ...</td>` rows, not the heading text). A regex extracts `M<N>` + the cell text up to the first `<` as the label. Fail-open: if the regex doesn't match, treat as no-design-report.

---

## §5. JSON shape rationale (additive-only contract)

The shape (spec.md §B.1) is locked at v0.1.0. The rationale per field:

| Field | Why locked | Forward-compat rule |
|---|---|---|
| `epic`, `epic_token` | Consumers key display + i18n on these | Never rename |
| `milestones[].id` (`M0`..`Mx`) | Display + ordered render | Format never changes |
| `milestones[].status` (4-value enum) | Banner icon mapping depends on it | Add new values ONLY at the end; never remove old |
| `milestones[].covered` (bool) | Banner's `done` counter keys on `status == "done" && covered` | Never change semantics |
| `milestones[].spec_id`, `spec_status`, `sync_commit_sha` | Banner click-through target + attribution | Add optional fields (e.g. `spec_tier`); never remove |
| `done`, `total`, `pct` | Banner's aggregate bar keys on these | Computation rule (REQ-ES-006, REQ-ES-013) never changes |
| `orphan_mx`, `extra_mx`, `untracked_specs` | Orphan detection output | Omitted when no canonical list (per REQ-ES-005); field name stable |
| `design_report` (path or empty) | Provenance for the canonical list | Type stable (string) |
| `baseline_attribution` (HEAD SHA) | `verification-claim-integrity.md` §2 contract | Type stable (string, possibly empty in non-git) |

Additive-only rule: a future release MAY add fields; consumers MUST tolerate unknown fields (forward-compat parse rule — verified by AC-ES-009). The shape itself NEVER removes or renames a field once shipped.

---

## §6. Alternatives considered + rejected

### A1 — Persisted `epic.json` written by an `epic init` verb

Rejected: creates a second source of truth alongside `.moai/specs/`. The kanban card store is the KANBAN BOARD SPEC's territory. The producer's value proposition is that it derives from signals already on disk.

### A2 — Route through the navigator graph

Rejected: `nav-graph.json` does NOT exist on disk at HEAD `9fa242dda` (verified by the parallel Explore audits 2026-08-11), and its schema (per `.claude/rules/moai/workflow/nav-tokens.md`) has no `epic`/`milestone` node type — only `decision`, `spec`, `symbol`. Adding those node types is a Navigator-Sync concern.

### A3 — Use only frontmatter `phase:` field for epic grouping

Rejected: `phase:` names a release target (per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Prohibited phase values), not an epic grouping. The BAS epic's SPECs share `phase: "v3.0.0 target"` but so do many unrelated SPECs — `phase:` is too coarse.

### A4 — Add a `epic:` frontmatter field to every SPEC

Rejected: introduces a new authoring convention that requires every existing SPEC to be touched for backfill. The `(BAS Mx)` title-marker convention is already in measured use (3 Navigator-Sync SPECs at HEAD) and derives the Mx map with zero new convention cost.

### A5 — Wire the banner consumption in this SPEC

Rejected: the producer is the higher-change-likelihood decision; banner consumption is a separate concern with its own rendering rules. Locking the JSON shape here lets the banner-wiring follow-up SPEC consume it without re-litigating the producer contract.

---

## §7. Risks revisited (cross-ref plan.md §B)

The plan.md §B Known Issues table carries the operational mitigations (KI-1..KI-6). The design-level risks:

- **DR-1**: the title-marker convention is social, not enforced — a SPEC author may forget the `(BAS Mx)` marker. Mitigation: the producer surfaces `untracked_specs` so a missing marker is visible, not silent.
- **DR-2**: the design-report HTML format is unstable — the maintainer may reformat the slice table. Mitigation: fail-open + the `--design-report` flag is the test entry point (tests pin a fixture copy's parseability).
- **DR-3**: the inferred-token heuristic may pick the wrong token when an epic has multiple tokens (e.g. a SPEC migrated from one epic to another). Mitigation: `--marker` flag override + stderr notice when the override differs from the inferred default.
