# Acceptance — SPEC-EPIC-STATUS-001

> Verification layer for the Epic Status producer. Each AC is binary-testable Given-When-Then. GEARS obligations live in spec.md §B (REQ-ES-001..013); this file does NOT restate them.

---

## §A. Severity legend

- **MUST**: blocks `implemented → completed` transition. Failure returns to `in-progress`.
- **SHOULD**: blocks only under `--strict` audit; non-blocking otherwise.
- **NICE**: informational; never blocks.

---

## §B. AC ↔ REQ traceability matrix

| AC | REQ coverage | Severity | Milestone |
|----|----|----|----|
| AC-ES-001 | REQ-ES-001 (disk-grounded), REQ-ES-002 (reuse path) | MUST | M0 |
| AC-ES-002 | REQ-ES-002 (no scanner fork, observation-only) | MUST | M0 |
| AC-ES-003 | REQ-ES-003 (prefix glob discovery) | MUST | M0 |
| AC-ES-004 | REQ-ES-004 (title regex Mx extraction) | MUST | M0 |
| AC-ES-005 | REQ-ES-005 (orphan-Mx detection against canonical list) | MUST | M1, M2 |
| AC-ES-006 | REQ-ES-006 (status join classification) | MUST | M2 |
| AC-ES-007 | REQ-ES-006 (divide-by-zero / total: 0 handling) | MUST | M2 |
| AC-ES-008 | REQ-ES-007, REQ-ES-008 (JSON shape + human mode) | MUST | M3 |
| AC-ES-009 | REQ-ES-008 (additive-only forward-compat) | MUST | M3 |
| AC-ES-010 | REQ-ES-009 (CLI verb registration) | MUST | M4 |
| AC-ES-011 | REQ-ES-010 (non-interactive) | MUST | M4 |
| AC-ES-012 | REQ-ES-011 (template-neutrality) | MUST | M4 |
| AC-ES-013 | REQ-ES-013 (no persisted epic store) | MUST | M2 |
| AC-ES-014 | REQ-ES-012 (factory touchpoint — docs only) | SHOULD | M5 |
| AC-ES-015 | REQ-ES-001, REQ-ES-008 (baseline attribution HEAD SHA) | MUST | M5 |

---

## §C. Given-When-Then acceptance criteria

### AC-ES-001 — Producer reads disk only (no working-memory dependency)

**Given** a project with `.moai/specs/SPEC-NAVIGATOR-SYNC-001/` through `-003/` on disk (the BAS epic fixture) AND a fresh Go process with no prior in-memory state
**When** `moai epic status NAVIGATOR-SYNC --json` is invoked
**Then** the producer emits a JSON document whose `milestones` array is populated from the on-disk SPEC frontmatter, irrespective of any conversation transcript or session state, and `baseline_attribution` is the current HEAD SHA.

### AC-ES-002 — Reuses `spec.ListDocs` + `spec.Audit`, no scanner fork, no mutation

**Given** a `t.TempDir()` project root populated with a fixture SPEC tree
**When** the producer runs `DiscoverEpic` + `JoinStatus` against it
**Then** the call stack includes `spec.ListDocs` and `spec.Audit` (verifiable via a spy or by mocking at the package boundary), AND a file-write sentinel (`ioutil.Discard`-style canary or a post-call mtime check on every file under `.moai/specs/`) shows ZERO mutations.

### AC-ES-003 — Prefix glob over `.moai/specs/`

**Given** the on-disk catalog has `SPEC-NAVIGATOR-SYNC-001/`, `SPEC-NAVIGATOR-SYNC-002/`, `SPEC-NAVIGATOR-SYNC-003/`, AND `SPEC-AUTH-001/` (unrelated)
**When** `moai epic status NAVIGATOR-SYNC --json` is invoked
**Then** the candidate set is exactly the three Navigator-Sync SPECs, `SPEC-AUTH-001` does NOT appear in `milestones`, `untracked_specs`, or anywhere in the output.

### AC-ES-003b — Empty match set is a clean empty, NOT an error

**Given** a project with no `SPEC-NONEXIST-*` dirs
**When** `moai epic status NONEXIST --json` is invoked
**Then** the producer exits 0 with output `{"epic":"NONEXIST","milestones":[],"done":0,"total":0,"pct":0,...}` and the human mode prints `🎯 NONEXIST — no SPECs matched prefix 'NONEXIST'`.

### AC-ES-004 — Mx extraction via `(<TOKEN> M<N>)` title regex

**Given** three SPECs with titles containing `(BAS M0)`, `(BAS M1)`, `(BAS M4)` respectively
**When** the producer runs Mx extraction
**Then** the Mx→SPEC map is `{M0: SPEC-NAVIGATOR-SYNC-001, M1: SPEC-NAVIGATOR-SYNC-003, M4: SPEC-NAVIGATOR-SYNC-002}` (or the actual owning SPEC per dir order), AND a SPEC with no marker is recorded in `untracked_specs` (NOT silently dropped).

### AC-ES-004b — Generic token (non-BAS)

**Given** a SPEC titled `"Custom Epic (EPICX M3) — sub-feature"` under prefix `CUSTOM`
**When** `moai epic status CUSTOM --marker EPICX --json` is invoked
**Then** the producer recognizes the `EPICX` token and maps M3 to that SPEC.

### AC-ES-005 — Orphan-Mx detection against design-report canonical list

**Given** the BAS design report at `.moai/reports/navigator-redesign-bas-20260805.html` (with the §7 slice table listing M0..M5) AND only M0/M1/M4 covered by SPECs
**When** `moai epic status NAVIGATOR-SYNC --json` is invoked (auto-discovering the design report per design.md §4)
**Then** `orphan_mx: ["M2", "M3", "M5"]` appears in the output, AND those three Mx entries have `"covered": false, "status": "absent"`.

### AC-ES-005b — No design report → omit `orphan_mx`

**Given** an epic with no discoverable design report
**When** `moai epic status <prefix> --json` is invoked
**Then** the JSON output omits the `orphan_mx` field entirely (REQ-ES-005 + design.md §5 lock the omit-when-empty form); the producer does NOT report orphans it cannot ground in a canonical list.

### AC-ES-006 — Status join per SPEC (done / in-progress / planned / absent)

**Given** M0 owned by a SPEC with `status: completed` + non-empty `sync_commit_sha`, M1 owned by `status: in-progress`, M2 owned by `status: draft`, M3 NOT owned by any SPEC
**When** the producer runs JoinStatus
**Then** M0 is `done`, M1 is `in-progress`, M2 is `planned`, M3 is `absent` (REQ-ES-006 classification), AND `done = 1`, `total = 4`, `pct = 25`.

### AC-ES-007 — Divide-by-zero / empty epic handling

**Given** an epic whose every Mx is `absent` (total = 0 covered + 0 canonical) OR an empty match set (total = 0)
**When** the producer computes `pct`
**Then** `pct` is `0` (NOT a NaN, NOT an error exit), AND `done = 0`, `total = 0`.

### AC-ES-008 — JSON shape matches frozen contract

**Given** the fixture BAS epic (3 covered + 2 orphan + 1 untracked, design report present)
**When** `moai epic status NAVIGATOR-SYNC --json` is invoked
**Then** the output is a single JSON document on stdout matching field-for-field the shape in spec.md §B.1, with stable key ordering, AND stderr is empty (no diagnostics on success).

### AC-ES-008b — Human mode mirrors Progress Board grammar

**Given** the same fixture
**When** `moai epic status NAVIGATOR-SYNC` is invoked (no `--json`)
**Then** the output's first non-blank line matches `/^🎯 .* ▓+░+ +3\/6 \(50%\)$/` (Progress Board bar grammar), AND each Mx has a status icon from `{🟢, 🟡, ⬜}` per the legend at `moai.md:660-668`, AND the rendering respects the `conversation_language` translation (Korean → `에픽 진행:`).

### AC-ES-009 — Additive-only forward-compat

**Given** a future release that adds a new field (e.g. `in_flight_worktrees`) to the JSON shape
**When** a consumer written against the v0.1.0 shape parses the future JSON
**Then** parsing succeeds and the unknown field is tolerated (forward-compat parse rule); the consumer does NOT crash on unknown fields. (Verified by a parse-tolerance test in M3.)

### AC-ES-010 — CLI verb registration under `moai epic`

**Given** the built `moai` binary
**When** `moai epic --help` is invoked
**Then** the output lists `status` as a subcommand, AND `moai epic status --help` lists the `<prefix>` positional and the `--json`, `--design-report`, `--marker` flags, AND `moai --help` lists `epic` in the command groups (mirroring how `spec` appears).

### AC-ES-011 — Non-interactive CLI code path

**Given** the `internal/cli/epic.go` source file
**When** grep is run for `AskUserQuestion`, `bufio.NewReader`, `fmt.Scanln`, `ReadString` over the file
**Then** zero matches (the CLI is read + print + exit only — REQ-ES-010 / C-HRA-008).

### AC-ES-012 — Template-neutrality (no mirror created)

**Given** the plan-phase commits under `.moai/specs/SPEC-EPIC-STATUS-001/**`
**When** `internal/template/split_namespace_test.go` and `internal/template/internal_content_leak_test.go` run
**Then** the tests do NOT flag this SPEC's artifacts (they are NOT mirrored to `internal/template/templates/`), AND the run-phase commits that add `internal/epic/*.go` and `internal/cli/epic.go` also do NOT add anything under `internal/template/templates/`.

### AC-ES-013 — No persisted epic store

**Given** a project before and after `moai epic status <prefix>` invocation
**When** the filesystem diff is computed (find . -newer <pre-marker> -type f)
**Then** ZERO new files are created under `.moai/` (no `epic.json`, no `epic-state.json`, no caches, no logs beyond stderr), AND the producer's source code contains zero `os.WriteFile` / `os.Create` / `ioutil.WriteFile` calls (verified by grep).

### AC-ES-014 — Factory touchpoint (docs only, SHOULD)

**Given** the docs-site factory-mode page exists (or its v3.2.0 successor)
**When** the page is rendered
**Then** it mentions `moai epic status <prefix>` as the visibility primitive for in-factory epic context (a single sentence is sufficient). This AC is SHOULD severity because the docs-site page is owned by another SPEC's lifecycle; the producer need only stabilize the CLI surface so the docs team has a target.

### AC-ES-015 — Baseline attribution HEAD SHA

**Given** a git project at HEAD SHA `X`
**When** `moai epic status <prefix> --json` is invoked
**Then** `baseline_attribution: "X"` appears in the JSON output, AND in a non-git context (`t.TempDir()` without `git init`), `baseline_attribution: ""` (empty string, fail-open per KI-6).

---

## §D. Quality gate criteria (TRUST 5 — binding at sync-phase)

- **Tested**: `go test -cover ./internal/epic/... ./internal/cli/epic_test.go` ≥ 85% coverage. Characterization test against the BAS epic fixture (read-only against the actual `.moai/reports/navigator-redesign-bas-20260805.html`).
- **Readable**: godoc on every exported symbol (`DiscoverEpic`, `ExtractMx`, `JoinStatus`, `ParseDesignReport`, `RenderJSON`, `RenderHuman`, `EpicStatus`, `MilestoneEntry`); English comments per `code_comments: en`.
- **Unified**: `gofmt -l .` empty; `goimports -l .` empty; `golangci-lint run ./internal/epic/... ./internal/cli/...` zero findings.
- **Secured**: input validation on `--design-report <path>` (path-traversal allowlist — must resolve under `.moai/reports/` or be an explicit absolute path); zero shell exec; zero file-write primitives in the producer's call graph.
- **Trackable**: Conventional Commit subjects `feat(SPEC-EPIC-STATUS-001): M0 ...` etc; sync commit subject `docs(SPEC-EPIC-STATUS-001): sync-phase artifacts` per the Status Transition Ownership Matrix.

---

## §E. Definition of Done

- All MUST-severity ACs (AC-ES-001..AC-ES-013, AC-ES-015) PASS with attributable evidence (test output, command + verbatim result, baseline HEAD SHA).
- AC-ES-014 (SHOULD) either PASS or explicitly deferred with rationale.
- `progress.md §E.2` (run-phase evidence) + `§E.3` (run-phase audit-ready) populated by manager-develop.
- `progress.md §E.4` (sync-phase audit-ready) populated by manager-docs on the sync commit.
- ZERO push, ZERO PR created from this branch (per user instruction); the work stays local on `feat/factory-bootstrap-guidance`.

---

## §F. Edge cases (non-blocking, but tested)

- **E1**: a SPEC whose title has `(BAS M0) — foo (BAS M1)` (double marker) → first wins, M1 in `extra_mx` (KI-4).
- **E2**: a SPEC dir whose `spec.md` is malformed (frontmatter parse fails) → recorded in `untracked_specs` with a parse-error note, does NOT abort the scan (mirrors `DocRecord.ParseError` per-record surfacing).
- **E3**: design report exists but the slice-table regex doesn't match (reformatted HTML) → fail-open, treat as no-design-report (KI-1).
- **E4**: `--marker` flag conflicts with the inferred token → `--marker` wins, with a stderr info-level notice.
- **E5**: prefix matches the `<prefix>-` boundary ambiguously (e.g. `NAV` matches `SPEC-NAVIGATOR-SYNC-*` AND `SPEC-NAV-EXTRA-*`) → all matches included; the producer does NOT guess; the user narrows the prefix.

---

## §G. Forward-looking checks (advisory, not blocking)

- **FL-1**: when the banner-consumption follow-up SPEC is authored, it SHOULD cite this SPEC's §B.1 JSON shape as its input contract and SHOULD add a regression test that the banner's render matches the producer's `--json` for a fixed fixture.
- **FL-2**: when `moai epic list` (multi-epic rollup) is authored, it SHOULD reuse `DiscoverEpic` + `JoinStatus` without modification; if either's signature must change, that change is additive (new optional `Options` field, NOT a breaking signature change).
