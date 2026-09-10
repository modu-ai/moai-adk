# Research — SPEC-TODO-LANDING-EVIDENCE-001

Every measurement this SPEC rests on: the command, its verbatim output, and the tree it ran against.
Nothing here is carried over from another tree or another time. What was **not** measured is in
`spec.md` §G, not omitted.

**Tree**: worktree `.claude/worktrees/t359`, branch `WT-landing-evidence`, HEAD `e50964ad3`.
**Measured**: 2026-09-03.
That HEAD is the default and holds unless a block says otherwise; a measurement taken at a later
HEAD of this same worktree carries its own SHA inline at the point of use, and today §R.10.1's
ordering note (`c0cfb2520`) is the only one.

---

## §R.1 Half A is landed (the stated dependency precondition)

```
$ grep -n "^status:" .moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md
5:status: completed

$ git merge-base --is-ancestor c9f712232 develop; echo "rc=$?"
rc=0
```

Both re-run in this worktree. The ancestry check decays as `develop` moves; it is re-read at
run-phase entry per `plan.md` §C rather than trusted from here.

## §R.2 `REQ-TODO-013` permits additive change (D1, first half — CONFIRMED)

```
$ grep -n "REQ-TODO-013" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
59:- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record
shape — `{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}` with
`state ∈ {queued, picked, dropped}` — changing it only additively (the high-water mark, per
REQ-TODO-009).
```

(Line wrapped for display; the source is one line.) The requirement constrains the *manner* of
change and names its own precedent. It does not freeze the field set.

## §R.3 The two SPECs agree (D1, second half — the card's premise is FALSIFIED)

```
$ awk 'NR==51 {print}' .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md
- 항목당 5필드(`id`, `text`, `added_at`, `spec_id`, `state`)는 SPEC-KANBAN-TODO-CLI-001 REQ-TODO-013
이 고정한 계약이다. 다만 그 SPEC의 §E는 **자기 범위**의 out-of-scope 선언이지 영구 동결이 아니고,
같은 REQ가 "additively" 변경을 허용한다 — `last_seq` 최상위 필드가 그 선례다.
```

The clause after 다만 is the whole point: §E is a scope-local declaration, not a permanent freeze,
and the same REQ permits additive change. `SPEC-TODO-ANALYSIS-001` does **not** judge the opposite of
§R.2 — it states the same reading. Neither SPEC was edited, and neither needs to be. (Their
measured statuses are `in-progress` and `completed` respectively — §R.10.1; v0.1.0 called both
`completed`.)

The precedent it names is present in the tree:

```
$ grep -n "type BacklogRecord" -A6 internal/kanban/backlog_store.go
187:type BacklogRecord struct {
188:	Version  int                   `json:"version"`
189:	LastSeq  int                   `json:"last_seq"`
190:	Items    []BacklogItem         `json:"items"`
...
```

## §R.4 The constraints that shape the design (D2 and the third constraint)

```
$ grep -n "REQ-1.10" -A5 .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md
251:**REQ-1.10** — The resolver shall not name, return, or otherwise claim which
252:commit delivered a card. The `landed` outcome is a boolean fact about
253:`origin/main` and nothing more. (Grounds: §C.2 — a card's first matching commit
254:may be another card's report commit that merely mentions it, so any
255:"first match is the delivering commit" reading attributes wrongly.)

259:**REQ-2.1** — [HARD] The read surface shall leave
260:`.moai/state/kanban/backlog.json` byte-identical across an invocation, and shall
261:write no field, no `findings[]` entry, and no timestamp.
```

REQ-1.10's **grounds** are the load-bearing part, and are why §B.3 resolves on provenance rather
than asking for permission. REQ-2.1 is why the writer cannot be `todo pr` (§B.2).

## §R.5 The schema-freeze guard is column-blind (measured, not inferred)

A temporary probe (`internal/kanban/zz_t359_probe_test.go`) planted
`ALTER TABLE items ADD COLUMN bogus TEXT` and re-ran all four assertions
`TestTodoHistoryAddsNoSchemaChange` makes. Full output:
`.moai/reports/t359/measure-guard-column-blind.txt`.

```
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
=== RUN   TestT359ProbeGuardIsColumnBlind
    A1 table set     = "archived_findings archived_items findings items meta" -> match=true
    A2 CHECK present = true
    A2 stored items SQL now = CREATE TABLE items (
      seq      INTEGER PRIMARY KEY,
      ...
      state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
    , bogus TEXT)
    A3 index set     = "idx_items_state" -> match=true
    A4 schema_version = "1" -> match=true
    COLUMN SET (unasserted by the guard) = "seq id text added_at spec_id state bogus"
--- PASS: TestT359ProbeGuardIsColumnBlind (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.408s
```

Two facts follow. The guard makes **no** column-set assertion, so this SPEC's own column would land
silently (⇒ REQ-TLE-019). And the `strings.Contains` CHECK test survives an added column because
SQLite appends it *after* the closing constraint — visible in the A2 output above.

The probe was deleted immediately after the measurement:

```
$ rm internal/kanban/zz_t359_probe_test.go
$ git status --short
?? .moai/reports/t359/
```

No source file under `internal/` is modified.

## §R.6 The items SQL surface, and why the downgrade claim is buildable

```
$ grep -rn "SELECT .*FROM items\|INSERT INTO items" internal/kanban/*.go | grep -v _test
internal/kanban/backlog_migrate.go:60:		`SELECT id, text, added_at, spec_id, state FROM items ORDER BY seq`)
internal/kanban/backlog_migrate.go:276:			`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (?, ?, ?, ?, ?, ?)`,
internal/kanban/backlog_migrate.go:340:		`SELECT COUNT(*) FROM items WHERE state = ?`, string(BacklogStateQueued)).Scan(&n); err != nil {
internal/kanban/backlog_migrate.go:351:	rows, err := e.db.QueryContext(ctx, `SELECT state, COUNT(*) FROM items GROUP BY state`)

$ grep -rn "SELECT \*" internal/kanban/ internal/cli/todo_pr.go; echo "rc=$?"
rc=1
```

Every statement names its columns; there is no `SELECT *`. This is the basis for REQ-TLE-018 — and
it is an argument, not a demonstration: no pre-change binary was actually run against a post-change
database (`spec.md` §G).

## §R.7 `ALTER TABLE` is new to this codebase

```
$ grep -rn "ALTER TABLE" internal/kanban/ internal/cli/; echo "rc=$?"
rc=1
```

No output. There is no in-tree idempotence pattern to copy, which is why `design.md` §2 states one
rather than assuming the reader supplies it.

## §R.8 The version gate rejects any bump

The v0.1.0 revision labelled this block `NR>=292 && NR<=298` while showing only five of the seven
lines — the leading and trailing braces were absent, so the "verbatim output" was not verbatim.
Re-measured and re-pasted in full:

```
$ awk 'NR>=292 && NR<=298 {printf "%d: %s\n", NR, $0}' internal/kanban/backlog_sqlite.go
292: 		}
293: 	case backlogSchemaVersion:
294: 		// current layout
295: 	default:
296: 		return fmt.Errorf("schema %s: unsupported schema_version %q (want %q): %w",
297: 			e.dbPath, version, backlogSchemaVersion, ErrBacklogCorrupt)
298: 	}
```

A `schema_version` bump is a downgrade break for every operator queue in the field, not a feature —
hence the [HARD] no-bump constraint in `plan.md` §D and the §D exclusion in `spec.md`.

## §R.9 Baseline test state

```
$ go test ./internal/kanban/ -run 'TestTodoHistoryAddsNoSchemaChange' -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.408s
```

Scoped to the guard under discussion. A full pre-edit baseline over the two touched packages is
`plan.md` §C's pre-flight obligation and belongs to run-phase entry, not to this document — running
it here would produce a figure that has decayed by the time it is used.

---

## §R.10 Measurements added at v0.2.0 (plan-audit iteration 1 remediation)

### §R.10.1 The cross-referenced SPEC statuses (D1)

```
$ grep -m1 '^status:' .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md \
    .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md \
    .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md \
    .moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md
.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:status: in-progress
.moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:status: in-progress
.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md:status: completed
.moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md:status: completed
```

(Full output, verbatim including paths, and **one capture of a non-deterministic ordering**. The
divergence from the argument order is the shell's, not `grep -m1`'s: `grep` here resolves to a shell
function wrapping a parallel grep — `type grep` reports it as a function from the session's
`shell-snapshots/` file — and that wrapper's output order varies between runs of the identical
command. `/usr/bin/grep -m1`, given the same arguments, emits in argument order deterministically
(6/6 identical runs, measured in this tree at HEAD `c0cfb2520`), so the property does not belong to
`grep -m1`; v0.3.0 attributed it there and that attribution is **withdrawn**. The four values are
not in dispute and match re-measurement; the ordering shown is simply one of several this shell
produces, and a reader re-running the command will often see another. v0.2.0 additionally pasted
this block with the `.moai/specs/` prefix stripped from every line and did not disclose the trim,
the same class of lapse as D11; re-measured and re-pasted untrimmed here.)

v0.1.0 asserted `completed` for the first two. That was inherited from the card text and never
measured — the document opened by claiming every figure in it was measured in this tree, and this
one was not. `spec.md` §A.3b carries the correction and the re-grounded argument.

### §R.10.2 `PRLinkOutcome` carries no delivering-commit field (§A.3b)

```
$ grep -n "type PRLinkOutcome" -A14 internal/kanban/prlink.go
101:type PRLinkOutcome struct {
102:	CardID string `json:"card_id"`
104:	Kind PRLinkKind `json:"outcome"`
107:	PRs []int `json:"pr,omitempty"`
111:	PRState string `json:"pr_state,omitempty"`
113:	Confidence PRLinkConfidence `json:"confidence,omitempty"`
114:}
```

(Comment lines elided for width; the field lines are verbatim at their stated line numbers.) Five
fields, none of them a delivering commit. This is one of the two live-in-the-tree properties §A.3b
relies on in place of a lifecycle status.

### §R.10.3 The `todo pr` git subprocess seam (D6)

```
$ sed -n '150,154p' internal/kanban/prlink_landed.go
	args, err := LandedGrepArgs(ref, cardID)
	if err != nil {
		return LandingUnknown, err
	}
	out, err := q.Run("git", args...)

$ sed -n '57,60p' internal/cli/todo_pr.go
var todoRunCommand kanban.CommandRunner = func(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), todoPRSubprocessTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
```

`moai todo pr` executes `git` in the project working directory. Git may write inside `.git/` during
a read, so AC-TLE-014's project-root byte-identity assertion excludes `.git/` and adds a
non-empty + queue-directory-present positive control so the exclusion cannot hollow it out.

### §R.10.4 The inherited prompt guard already covers the new verb (D10)

```
$ sed -n '451,480p' internal/cli/todo_test.go
func TestTodoCmd_NoAskUserQuestion(t *testing.T) {
	sources, err := filepath.Glob("todo*.go")
	...
	// Positive control on the scan itself: a glob that matched nothing
	// would report every file clean without reading one.
	if scanned < 2 {
		t.Errorf("guard scanned %d todo sources, want the whole surface", scanned)
	}

	// Negative control: the guard must flag a synthetic violation.
	if _, bad := todoPromptGuard("x := AskUserQuestion()"); !bad {
		t.Error("guard must detect an AskUserQuestion reference (negative control)")
	}
}
```

The glob is `todo*.go`, so `todo_landed.go` is covered automatically once it exists — and the
existing guard carries a `scanned < 2` positive control and a synthetic-violation negative control
that a fresh grep conjunct in AC-TLE-007 would not. The conjunct was removed rather than kept as a
weaker duplicate.

### §R.10.5 What v0.2.0 did NOT measure

- **The validation commands were not executed.** `git rev-parse --verify <sha>^{commit}` and
  `git merge-base --is-ancestor` are named in REQ-TLE-020 and `design.md` §3 from their documented
  semantics; neither was run against a fixture here, because the fixture they need (a commit
  deliberately unreachable from the record's ref) is AC-TLE-020's to build in run-phase.
- **The plan-audit's `archived_items` blindness probe was not re-run.** It is cited from
  `.moai/reports/t359/plan-audit.md` Hunt 2 as an independent second observation beside this SPEC's
  own `items` probe (§R.5). Both are recorded as decaying: the guard belongs to
  `SPEC-TODO-ARCHIVE-QUERY-001`, and run-phase re-measures rather than citing either.
- **No pre-change binary was run.** Unchanged from v0.1.0 and now stated as a standing gap in
  `spec.md` §G rather than deferred to AC-TLE-018 (D7).
