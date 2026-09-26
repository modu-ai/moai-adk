# Acceptance — SPEC-ROLE-NAMING-DOCS-001

Version 0.1.0 · 2026-09-26 · manager-spec · card t1257

All commands run in the card worktree against the edited tree unless stated. `$R` is `.moai/reports/t1257`; `$SP` is a scratch directory holding `files.txt` from `git ls-files`.

## §A Acceptance Criteria

### AC-RND-001 — Gate verified before substitution

- **Given** the run phase is about to start M4, M5, or M6
- **When** progress.md is read
- **Then** it carries a gate row naming the t1256 SPEC status read from develop (verbatim `status:` line) or the operator-recorded code-layer decision, and recorded answers to research.md §F Q1–Q5.

### AC-RND-002 — Halt when the gate fails

- **Given** the gate of AC-RND-001 does not hold
- **When** the run phase reaches a substitution milestone
- **Then** it returns a blocker report and `git diff --stat <plan-commit>..HEAD -- internal/template/templates .claude docs-site README.md README.ko.md README.ja.md README.zh.md CLAUDE.md AGENTS.md CLAUDE.local.md` prints nothing.

### AC-RND-003 — Target vocabulary present

- **Given** M4–M6 are complete
- **When** the glossary page and `kanban-dispatch.md` (template copy) are read
- **Then** both name the coordinating session `leader` and the card-carrying session `lane`, and the glossary carries the operator's model statement.

### AC-RND-004 — Identifier forms match the code

- **Given** the develop HEAD the edit targets
- **When** every CLI token, session notation, env var, and sentinel written in the edited docs is checked against the code-layer vocabulary table in progress.md
- **Then** zero documented forms are ones the code does not accept (for each `-f <token>` form in docs, `moai cc --help` or the factory parser test named in the table accepts it).

### AC-RND-005 — Other meanings untouched

- **Given** the plan-time raw table `$R/raw/per-file-class.tsv`
- **When** the inventory is re-run and the `other-meaning:*` and `plain-english` rows are compared file by file
- **Then** their counts are unchanged, and every changed line that the ledger classes as cg / Agent Teams / Lane A-B / Epic Lane / doc-companion / leaf-worker is marked "kept".

### AC-RND-006 — Disposition ledger complete

- **Given** the final diff
- **When** `git diff <plan-commit>..HEAD --unified=0` changed lines carrying a role token are joined against the ledger by path and line
- **Then** every such line has a ledger row, and the ledger is committed under `$R/`.

### AC-RND-007 — HARD clauses preserved

- **Given** every file touched by M4–M6
- **When** `grep -c '\[HARD\]' <file>` runs on the plan commit and on HEAD
- **Then** the counts are equal for every file, and each touched `[HARD]` line's ledger row states "subject noun only".

### AC-RND-008 — No dangling anchors

- **Given** the edited tree
- **When** the section-reference scan of inventory §6.2 and `raw/scripts/anchors.py` are re-run
- **Then** every `§ <text>` reference containing a role word resolves to an existing heading or bold paragraph, and no reference names an anchor text that exists only at the plan commit.

### AC-RND-009 — Template-First and build

- **Given** each mirrored file changed
- **When** the commit order and `make build` output are inspected
- **Then** the template copy changed in the same or an earlier commit than the local copy, and `make build` exits 0.

### AC-RND-010 — Codex copies regenerated, not hand-edited

- **Given** an agent definition under `internal/template/templates/.claude/agents/moai/` changed
- **When** `make agents-emit-check` runs
- **Then** it exits 0, and the commit that changed any `.codex/agents/moai/*.toml` also changed its `.md` source.

### AC-RND-011 — `manager-lead` follows the recorded option

- **Given** the operator's Q4 answer in progress.md
- **When** `grep -rn 'manager-lead' internal/template/templates/.claude/agents/moai/ CLAUDE.md` runs
- **Then** under option A the file `manager-lead.md` still exists and its definition sentence names it the leader's coordination agent; under B/C/D this SPEC's diff contains no agent-file rename.

### AC-RND-012 — Template neutrality

- **Given** the template edits
- **When** the template-neutrality check script from `.github/workflows/template-neutrality-check.yaml` runs locally
- **Then** it reports no violation, and `grep -rnE 'SPEC-[A-Z]|\bt1257\b|2026-09-26' <changed template files>` prints nothing new.

### AC-RND-013 — Four locales and four READMEs together

- **Given** the M6 change set
- **When** the changed docs-site pages are grouped by page path
- **Then** each changed page path changed in all four locales, all four README files changed, and the lexicon in progress.md was committed before the first docs commit.

### AC-RND-014 — Glossary disambiguation

- **Given** `docs-site/content/{en,ko,ja,zh}/core-concepts/kanban-board-terms.md`
- **When** each is read
- **Then** each defines leader and lane and carries one line separating the Kanban/Factory leader from the `moai cg` leader pane and the Agent Teams leader.

### AC-RND-015 — `CLAUDE.local.md` edited only in the develop copy

- **Given** M5
- **When** `git log --format=%h -- CLAUDE.local.md` on the card branch is read
- **Then** the §4.1 change exists as a commit on the card branch, and no instruction in the ledger or progress.md edits or restores the primary checkout's copy.

### AC-RND-016 — No Go edits; pinned strings routed

- **Given** the final diff
- **When** `git diff --name-only <plan-commit>..HEAD -- '*.go'` runs
- **Then** it prints nothing, and every doc edit that would change a string asserted by a test in inventory §7 appears in progress.md as a blocker routed to the code-layer card.

### AC-RND-017 — Frozen records untouched

- **Given** the final diff
- **When** `git diff --name-only <plan-commit>..HEAD -- CHANGELOG.md '.moai/specs/**'` runs
- **Then** it lists only this SPEC's own directory, and no HISTORY-section line or deprecated-alias disclosure is rewritten.

### AC-RND-018 — No model change in a naming commit

- **Given** the ledger rows for the clauses "sole producer", "promotion is the operator's act", and the per-column companion description
- **When** the before and after texts are compared
- **Then** only the role nouns differ, unless progress.md records a Q2/Q3 model-change answer and names the separate SPEC that carries it.

### AC-RND-019 — Local-only files stay local

- **Given** the local-only role-bearing files in inventory §5
- **When** `git diff --name-status <plan-commit>..HEAD -- internal/template/templates` runs
- **Then** it adds no file mirroring any of them.

### AC-RND-020 — Zero residue

- **Given** the edited tree
- **When** `git ls-files > $SP/files.txt && python3 $R/raw/scripts/inv.py $SP $R/raw-final` runs
- **Then** the `role`-class counts of `lead`, `companion`, and Factory-sense `worker` on the in-scope surfaces equal the number of ledger rows marked "kept" for those classes.

## §B Traceability

| REQ | AC |
|---|---|
| REQ-RND-001 | AC-RND-003 |
| REQ-RND-002 | AC-RND-001 |
| REQ-RND-003 | AC-RND-002 |
| REQ-RND-004 | AC-RND-004 |
| REQ-RND-005 | AC-RND-005 |
| REQ-RND-006 | AC-RND-006 |
| REQ-RND-007 | AC-RND-007 |
| REQ-RND-008 | AC-RND-008 |
| REQ-RND-009 | AC-RND-009 |
| REQ-RND-010 | AC-RND-010 |
| REQ-RND-011 | AC-RND-011 |
| REQ-RND-012 | AC-RND-012 |
| REQ-RND-013 | AC-RND-013 |
| REQ-RND-014 | AC-RND-014 |
| REQ-RND-015 | AC-RND-015 |
| REQ-RND-016 | AC-RND-016 |
| REQ-RND-017 | AC-RND-017 |
| REQ-RND-018 | AC-RND-018 |
| REQ-RND-019 | AC-RND-019 |
| REQ-RND-020 | AC-RND-020 |

## §C Edge Cases

- A line mixes the Kanban lead and the cg leader (e.g., a comparison table): split the disposition by match, not by line; the ledger records both.
- A heading renamed in the template copy but the local copy differs (`manager-lead.md`): anchors are checked per copy.
- A zh page uses 主控 for the lead and 主导 elsewhere for a non-role meaning: the lexicon names which word is the role word; the other stays.
- The t1256 decision keeps `worker-N` as the notation: docs keep `worker-N` for the identifier and use "lane" only in prose (REQ-RND-004 wins).

## §D Quality Gates and Definition of Done

- All 20 criteria PASS with the command and verbatim output in progress.md §E.2.
- `make build`, `make agents-emit-check`, `make embed-check` exit 0; `go test ./internal/template/... ./internal/kanban/...` passes; hugo build warning-free.
- plan-auditor PASS at the Tier L threshold (0.85) before run.
- Merge to develop through the integration window; the leader pushes.
