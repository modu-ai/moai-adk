# P07 local TODO promotion — superseded (2026-11-10)

**Status: superseded.** User decision 2026-11-10T11:52Z (Issue #2384): "Remove the local TODO; use only GitHub/GitLab Issues. Without GitHub/GitLab the existing todo is not used. Remove SQLite too." The earlier P07 rule ("when a forge is connected later, preview existing local TODOs and promote only user-selected items", 07-B #2009) no longer applies. Its text and product path stay in git history (this file before #2384; `src/forge/promote.rs` until the removal commit).

## Rule

- **The work ledger is forge Issues only** (GitHub; GitLab when its adapter lands). There is no local TODO list, no local task store and no promotion from one.
- **No forge connected means no ledger.** Ledger operations never create, read or migrate a local TODO, and they never fall back to a local store. They answer with the shared forge error classification as it is (user decision 2026-11-10, #2389 comment 6774899518 row 37: keep the current behaviour and fix this document; the shared classification is not changed):
  - `gh` missing from `PATH`, or `gh` not signed in: error kind `transport`, `retryable: true` (the transport cannot reach the forge; signing in or installing `gh` and retrying succeeds);
  - no repository named in the request: error kind `invalid_input`, `retryable: false`.
  The test `forge_read_without_a_forge_creates_nothing_and_answers_transport` (`tests/ledger_forge_unconnected_cli.rs`) pins both answers and that no file is written. (This rule first said `blocked`/`retryable: false`; measured behaviour differed, Issue #2384 comment 6771930434.)
- **No SQLite.** No state, index, queue or TODO store uses SQLite. The stores that exist are file based and already implemented: the claim journal (`claim::journal_root`), the generation receipts (`apply::Receipt`), the forge outbox (`<project>/.moai-cli/forge/outbox/`) and the P09 home layout pointer. `Cargo.toml` has no SQLite dependency (checked at tip `2930b8631f`).
- **Out of scope** (preserved): the Go tool's own todo and SQLite features and any data under `~/.moai/`. Existing `<project>/.moai-cli/forge/promotions.json` files written by the removed preview/apply are left in place; nothing deletes user data.

## Inventory at tip `2930b8631f` (two independent counts)

Method A is `git grep -i -E` over tracked files. Method B (`target/agent5-2384/scan.py`, kept out of the tree) reads every tracked blob through `git cat-file --batch` and matches with Python `re`. Both report 585 tracked files and the same numbers:

| pattern | files | lines |
| --- | ---: | ---: |
| `todo-promote\|TodoPromote\|todo_promot` | 17 | 85 |
| `sqlite` | 16 | 39 |
| `로컬 TODO\|local TODO\|local todo` | 14 | 24 |

### Removed by the code card (waits for its merge-train turn; every open PR touches `src/main.rs`)

| path | what |
| --- | --- |
| `src/main.rs` | `TodoPromotePreview`/`TodoPromoteApply` variants, their operation names `todo.promote.plan`/`todo.promote.apply`, dispatch arms |
| `src/forge/promote.rs` (857 lines), `src/forge/mod.rs` `pub mod promote;` | the promotion module; not in the leader's first inventory |
| `tests/todo_promotion_cli.rs`, `tests/todo_promotion_test.py`, `experiments/todo_promotion.py` | tests and the in-memory experiment |
| `docs/design-review/appendix/interfaces.md` | the two lines `todo-promote-apply`, `todo-promote-preview` (the test `operations_doc_matches_code` compares the list with the code) |
| `docs/rc-file-selection.json` | the two entries for `experiments/todo_promotion.py` and `tests/todo_promotion_test.py` |
| `src/lib.rs:70`, `src/redact.rs:8` | doc comments that name `forge::promote` as a consumer of the shared secret-shape filter (the filter itself stays; promote only consumed it) |
| `.gitignore:41-42` | `*.sqlite`, `*.sqlite3` (no such file exists in the primary checkout to depth 4; leader-owned file) |
| `src/forge/outbox.rs:276,295-303` | test fixtures use the operation word `promote`; they test the marker grammar, not promotion, and may stay or be renamed |

After removal the code card checks for dead code left behind in `forge::{outbox, transport, write}` (clippy `-D warnings` fails on it) and adds the AC-1/AC-2 tests: the two subcommands exit 2 as usage errors and are absent from `--help`; a temporary project with no forge gives ledger operations the answers listed under Rule and creates no file.

### Changed in this document card

- this file (supersession, rule, inventory);
- `skills/moai/workflows/sync.md` step 4: the sentence about moving local TODOs to a newly connected forge is replaced by the forge-only rule, and `skills/moai/references/source-manifest.json` carries the new digest of that file.

### Kept: history, records and other systems (not current design)

| path | why it stays |
| --- | --- |
| `docs/decisions-2026-11-08.md:69` | decision log entry (history); the leader records the 2026-11-10 decision there |
| `docs/release-notes/2026.11.0-rc.1.md:84` | release history of #2009 |
| `docs/design-review/review-group2-core-brain-sync-owner.md:138,146`, `review-group3-install-host-plan-run-release.md:52,368`, `review-group4-past-designs.md`, `synthesis-improvement-report-2026-11-09.md`, `fixes/group2-0d4538a27fb88191a19ba261bd5a3442.md`, `fixes/top-section20-lessons.md` | dated review records. The SQLite lines there already say SQLite is not adopted or describe the Go factory; review-group3:52 records that promotion was implemented at that time |
| `docs/design-review/codex-statusline-official-scope-2026-11-10.md:52`, `statusline-rust-stage1-2026-11-10.md:177`, `docs/session-hub-research-2026-11-08.md:176` | Codex's or the Go tool's own SQLite, read as research; not a MoAI store |
| `tests/fixtures/github/issues-open.json:480`, `tests/fixtures/atlas-tag/go-corpus.txt:1386` | recorded fixtures (Issue title, Go source corpus) |
| `.moai-cli/manifests/issue-design-research*.json`, `github-issues.json`, `github-duplicate-map.json`, `baseline/rc1-inventory.*`, `.moai-cli/reports/rc1-issue-design-2026-11-08.html` | leader ledger mirror and design snapshots (history) |
| `.moai-cli/reports/issue-bodies/{1977,1983,1989,2005,2006,2009}*.md` | Issue body copies; the leader edits the live Issue bodies (proposals below) and refreshes the copies |

## Proposed Issue body edits (the leader applies them)

Each affected requirement gets one line: "Removed by the user decision 2026-11-10 (#2384): the work ledger is forge Issues only; no local TODO and no SQLite."

- #1983 (EP-08) and #2009 (07-B): the "local TODO selected promotion" section and its acceptance rows.
- #1977 (EP-02), #1989 (01-B), #2005 (06-A), #2006 (06-B): the sentences that name SQLite as a state store or index; the file stores listed under Rule replace them.
