# Acceptance — SPEC-ROLE-NAMING-DOCS-001

Version 0.3.0 · 2026-09-26 · manager-spec · card t1257

All commands run in the card worktree against the edited tree unless stated. `$R` is `.moai/reports/t1257`; `$SP` is a scratch directory holding `files.txt` from `git ls-files`; `<plan-commit>` is the commit that carries this SPEC's final plan version.

## §A Acceptance Criteria

### AC-RND-001 — Gate verified before substitution

- **Given** the run phase is about to start M4, M5, or M6
- **When** progress.md is read
- **Then** it carries a gate row with the verbatim output of `git show develop:.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md | grep '^status:'` reading `implemented` or `completed`, and a row confirming the code-layer term table lists no accepted legacy spelling.

### AC-RND-002 — Halt when the gate fails

- **Given** the gate of AC-RND-001 does not hold
- **When** the run phase reaches a substitution milestone
- **Then** it returns a blocker report and `git diff --stat <plan-commit>..HEAD -- internal/template/templates .claude docs-site README.md README.ko.md README.ja.md README.zh.md CLAUDE.md AGENTS.md CLAUDE.local.md` prints nothing.

### AC-RND-003 — Target vocabulary present

- **Given** M4–M6 are complete
- **When** the glossary page and `kanban-dispatch.md` (template copy) are read
- **Then** both name the coordinating session `leader` and the Factory card-carrying session `lane`, and the glossary carries the operator's model statement.

### AC-RND-004 — Canonical identifiers only

- **Given** the edited tree
- **When** `grep -rnE -- '-f (worker|agent)\b|\b(worker|agent)-<n>|\b(worker|agent)-[0-9]+\b' <in-scope doc surfaces>` runs
- **Then** it prints only lines the ledger marks "kept" as another meaning (none describing a Factory session); and every form printed by `grep -rhoE -- '-f [a-z]+(-<n>|-[0-9]+)?' <in-scope doc surfaces> | sort -u` appears in the join-token or numbered-label row of the `SPEC-ROLE-NAMING-CODE-001` design.md §3 term table read from develop at the gate HEAD (`git show develop:.moai/specs/SPEC-ROLE-NAMING-CODE-001/design.md`), with the extracted list and the matched rows recorded in progress.md.

### AC-RND-005 — Other meanings untouched

- **Given** the plan-time raw table `$R/raw/per-file-class.tsv`
- **When** the inventory is re-run and the `other-meaning:*` and `plain-english` rows are compared file by file
- **Then** their counts are unchanged except for the qualifier words added by AC-RND-021, and every changed line that the ledger classes as cg / Agent Teams / Lane A-B / Epic Lane / doc-companion / leaf-worker is marked "kept" or "qualified".

### AC-RND-006 — Disposition ledger complete

- **Given** the final diff
- **When** the changed lines of `git diff <plan-commit>..HEAD --unified=0` that carry a role token are joined against the ledger by path and line
- **Then** every such line has a ledger row, and `git ls-files $R/` lists the ledger and every re-run output cited as evidence (they were added with `git add -f`, since `.moai/reports/*` is gitignored).

### AC-RND-007 — Non-target HARD clauses preserved

- **Given** every file touched by M4–M6
- **When** `grep -c '\[HARD\]' <file>` runs on the plan commit and on HEAD
- **Then** the counts are equal for every file that carries no amendment target of REQ-RND-018 to REQ-RND-020, not less than the plan-commit count for every file that carries one (e.g., both copies of `kanban-dispatch.md`, 38 markers each at `e62c3e183`), and each touched `[HARD]` line outside the amendment targets has a ledger row stating "subject noun only".

### AC-RND-008 — No dangling anchors

- **Given** the edited tree
- **When** (1) `git ls-files > $SP/files.txt` is taken on the edited tree, the plan-time `$R/raw/headings.tsv` is copied into `$R/raw-anchor-check/`, and `python3 $R/raw/scripts/anchors.py $SP $R/raw-anchor-check` runs; and (2) the section-reference scan of inventory §6.2 is re-run on the edited tree
- **Then** (1) in `$R/raw-anchor-check/anchors.tsv`, every heading the ledger marks as renamed shows `ref_files` = 0 for its old text; and (2) every `§ <text>` reference containing a role word resolves to an existing heading or bold paragraph (including `§ Lane spawn authority`).

### AC-RND-009 — Template-First and build

- **Given** each mirrored file changed
- **When** the commit order and `make build` output are inspected
- **Then** the template copy changed in the same or an earlier commit than the local copy, `make build` exits 0, and `git diff --name-status <plan-commit>..HEAD -- internal/template/templates` adds no file mirroring a local-only file of inventory §5.

### AC-RND-010 — Codex copies regenerated, not hand-edited

- **Given** an agent definition under `internal/template/templates/.claude/agents/moai/` changed
- **When** `make agents-emit-check` runs
- **Then** it exits 0, and every commit that changed a `.codex/agents/moai/*.toml` also changed its `.md` source.

### AC-RND-011 — `manager-lead` name kept

- **Given** the edited tree
- **When** `ls internal/template/templates/.claude/agents/moai/manager-lead.md .claude/agents/moai/manager-lead.md` and `git diff --name-status <plan-commit>..HEAD | grep -i 'manager-lead'` run
- **Then** both files exist, no rename (`R`) or delete (`D`) status appears, and each definition site named in REQ-RND-011 carries the kept-name sentence.

### AC-RND-012 — Template neutrality

- **Given** the template edits
- **When** the template-neutrality check from `.github/workflows/template-neutrality-check.yaml` runs locally
- **Then** it reports no violation, and `grep -nE 'SPEC-[A-Z]|\bt1257\b|2026-09-26' <changed template files>` prints nothing new.

### AC-RND-013 — Four locales, one lexicon

- **Given** the M6 change set
- **When** changed docs-site pages are grouped by page path and the zh role words are counted with Python `str.count`
- **Then** each changed page path changed in all four locales, all four README files changed, and in zh the role-sense occurrences of 主控 / 领导 / 负责人 are 0 with every role occurrence reading 主导 (or 主导会话) and 泳道.

### AC-RND-014 — Glossary disambiguation

- **Given** `docs-site/content/{en,ko,ja,zh}/core-concepts/kanban-board-terms.md`
- **When** each is read
- **Then** each defines leader and lane and carries one line separating the factory leader from the `moai cg` leader pane and the Agent Teams team lead.

### AC-RND-015 — `CLAUDE.local.md` edited only in the develop copy

- **Given** M5
- **When** `git log --format=%h -- CLAUDE.local.md` on the card branch is read
- **Then** the §4.1 change exists as a commit on the card branch, and no instruction in the ledger or progress.md edits or restores the primary checkout's copy.

### AC-RND-016 — No Go edits; pinned strings routed

- **Given** the final diff
- **When** `git diff --name-only <plan-commit>..HEAD -- '*.go'` runs
- **Then** it prints nothing, and every doc edit that would change a string asserted by a test in inventory §7 appears in progress.md as a blocker routed to the code-layer card.

### AC-RND-017 — No legacy alias described; frozen records untouched

- **Given** the edited tree and the plan-time list `$R/raw/q1-legacy-alias-lines.txt`
- **When** `grep -rnE -- '-f (agent|worker)\b|(agent|worker)-<n>|legacy label|legacy (agent|lane|worker)|deprecated alias' <in-scope doc surfaces>` runs (`lane-<n>` and `-f lane` are canonical and are not matched), and `git diff --name-only <plan-commit>..HEAD -- CHANGELOG.md '.moai/specs/**'` runs
- **Then** the first prints nothing describing a Factory session spelling, and the second lists only this SPEC's own directory.

### AC-RND-018 — Pre-dispatch clauses name the actor per obligation

- **Given** `kanban-dispatch.md` in the template and local copies, and the ledger rows for the clauses at `e62c3e183` lines L31, L33, L37, L39, L41, L49, L266
- **When** each amended clause is read, and `grep -c 'the dispatching party (the leader, or a lane that promoted the card itself)' <copy>` runs on each copy
- **Then**, per copy:
  - the promotion clause (ledger row L31) names exactly two promoters — the operator, and a lane promoting an already-queued card to itself — still states that the leader never promotes on its own initiative, and still carries its `[HARD]` marker;
  - the grep prints at least 4, and the phrase is the grammatical subject of each of the four obligations: the PR/landed cross-check (row L37), the completed-SPEC cross-check (row L39), surfacing an open PR or landed card for the operator to confirm or withdraw (row L41), and the A/B/C class assignment (row L49);
  - one sentence states that a lane that promoted a card itself performs those four obligations itself and reports each result to the leader before starting work;
  - the "What stays forbidden" paragraph (row L33) keeps every prohibition on the leader and contains no sentence forbidding a lane's self-promotion of an already-queued card;
  - the Factory Mode paragraph (row L266) names both the leader's dispatch and the lane's self-promotion as ways a lane receives a card.

### AC-RND-019 — Production clause meaning unchanged

- **Given** the ledger row for "The lead is the queue's sole producer"
- **When** the before and after texts are compared word by word
- **Then** only the role noun differs, the standing-source exception is intact, and no sentence grants a lane a production right.

### AC-RND-020 — Echoes amended together

- **Given** the plan-time echo list `$R/raw/q3-clause-echoes.tsv` (124 lines, 51 files, produced by `python3 .moai/reports/t1257/raw/scripts/echoes.py`)
- **When** the same command is re-run on the edited tree into `$R/raw-final/q3-clause-echoes.tsv`
- **Then** every plan-time line and every re-run line has a ledger row marked "amended", "consistent with the amended clause", or "other meaning" (the `internal/cli/todo.go` rows marked "routed to code layer"); every "amended" row changed in one commit; no remaining re-run line states that only the operator promotes or that a lane never picks a card; and no touched file lost a `[HARD]` marker.

### AC-RND-021 — Leader homonym qualified at first occurrence

- **Given** every in-scope file whose ledger rows classify an occurrence of leader / lead as one of the three role senses of REQ-RND-021 (plain-English uses and identifiers such as `manager-lead` excluded)
- **When** the first line of each sense in each file is located from the ledger
- **Then** that line contains the sense's qualifier in the file's locale exactly as tabled in REQ-RND-021 — factory leader / team lead(er) / cg leader pane; 팩토리 리더 / 팀 리더 / cg 리더 패인; ファクトリーリーダー / チームリーダー / cg リーダーペイン; 工厂主导 / 团队队长 / cg 领队窗格.

### AC-RND-022 — Kanban companions stay companions

- **Given** the edited `kanban-dispatch.md`, `kanban-board-terms.md` (four locales), and `advanced/kanban-mode.md` (four locales)
- **When** the plan / run / sync session descriptions are read
- **Then** they are called companions, and no line calls a Kanban column session a lane.

### AC-RND-023 — Auxiliary-role definitions

- **Given** the definition sites of foreman (`moai-kanban-foreman/SKILL.md`, `loop.md`), deputy (`manager-lead.md` § Deputy dispatch surface, `kanban-dispatch.md` § Deputy dispatch surface), and coordinator (docs-site `advanced/manager-lead.md` title)
- **When** each site is read
- **Then** the term keeps its name and carries one line defining it as an auxiliary role of the leader.

### AC-RND-024 — Zero residue

- **Given** the edited tree
- **When** `git ls-files > $SP/files.txt && python3 $R/raw/scripts/inv.py $SP $R/raw-final` runs
- **Then** the `role`-class counts of `lead` and of Factory-sense `worker` on the in-scope surfaces equal the ledger rows marked "kept" for those classes, and the AC-RND-017 grep prints nothing.

### AC-RND-025 — Alias-disclosure lines removed or rewritten

- **Given** the 42 plan-time lines of `$R/raw/q1-legacy-alias-lines.txt`
- **When** each line's file is read at HEAD and the ledger row for each line is located
- **Then** each row is marked "removed" or "rewritten to canonical", no row is marked "kept", and every rewritten line naming `lane-<n>` presents it as the current form with no legacy or deprecated wording.

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
| REQ-RND-021 | AC-RND-021 |
| REQ-RND-022 | AC-RND-022 |
| REQ-RND-023 | AC-RND-023 |
| REQ-RND-024 | AC-RND-024 |
| REQ-RND-025 | AC-RND-025 |

## §C Edge Cases

- A line mixes the factory leader and the cg leader (e.g., a comparison table): disposition by match, not by line; the ledger records both.
- A heading renamed in the template copy but the local copy differs (`manager-lead.md`): anchors are checked per copy.
- A zh page uses 负责人 for a non-role meaning (e.g., a document owner): it stays; only role-sense occurrences move to 主导.
- A file uses only one sense of leader: only that sense is qualified at its first occurrence.
- The code-layer table at the gate still lists an accepted legacy spelling: the gate fails (AC-RND-001) — documents never describe it.

## §D Quality Gates and Definition of Done

- All 25 criteria PASS with the command and verbatim output in progress.md §E.2.
- `make build`, `make agents-emit-check`, `make embed-check` exit 0; `go test ./internal/template/... ./internal/kanban/...` passes; hugo build warning-free.
- plan-auditor PASS at the Tier L threshold (0.85) before run.
- Merge to develop through the integration window; the leader pushes.
