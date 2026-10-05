# SPEC-HARNESS-DETACHED-PRUNE-001 — Acceptance Criteria

## §A Discipline and Tree Pin

Every criterion below adopts the two-cell discipline (`.claude/rules/moai/development/verification-completeness.md` §2): a RED-now cell observed on the pre-implementation tree and a GREEN path naming the milestone that flips it. All RED cells were measured in this plan session on **`d05d1d5f0`** — branch `WT-harness-prune-detached`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1497`, status clean, from the worktree root. Each RED cell states why it is red: the matched identifiers name code this SPEC will create, or — for the defect-presence cells (A, G) — the defect exists; no wrong-reason red. New-test cells use the package-wide `go test -list` corroboration so a bare selector matching nothing cannot read as a pass (LEDGER-DP-B and LEDGER-DP-D are that guard). Every check names WHEN it runs (its milestone exit), the input that turns it red, and who sees the red (the exit code of the milestone's exit gate, then the plan-auditor and the run-phase orchestrator via progress.md §E.1). AC-DP-007 is classified regression-guard (its green is the pre-existing state; no input on this tree turns it red today), never release-blocking, per §2.1's undecidable disposition.

GREEN-side executable commands live in the fenced ledger too wherever they carry alternation: a markdown table cell cannot hold a bare pipe, and the escaped form is a literal pipe in Go regexp — the iter1 audit's D1 defect — so AC-DP-001's GREEN commands were moved to LEDGER-DP-GREEN-A. LEDGER-DP-FORM records the grammar form-controls (Go-regexp selector pair; grep BRE alternation) so a zero-hit grep or a zero-match selector can never read as a dead-pattern pass. The `\|` sequences that remain inside fenced ledger commands are the quoted input bytes of those controls, not cell escaping.

## §B RED-Now Evidence Ledger

```
LEDGER-DP-A
  cmd:  grep -c "PruneStaleEntries" internal/harness/observer.go
  out:  2
  exit: 0
  why:  defect-presence cell for REQ-DP-001: the two synchronous prune calls
        this SPEC removes are present in the observer record path. Red because
        the defect exists; flips only when both call sites are gone (M1 GREEN
        expects this count to reach 0).
  tree: d05d1d5f0

LEDGER-DP-B
  cmd:  go test -list 'TestRecordExtendedEventDoesNotPrune|TestRecordEventDoesNotPrune|TestObserverRecordsWithoutPrune' ./internal/harness
  out:  ok  	github.com/modu-ai/moai-adk/internal/harness	0.792s
  exit: 0
  why:  package-wide corroboration for AC-DP-001's test names — the selector
        matches no test in the package, so the AC's test is genuinely new,
        not a pre-existing green.
  tree: d05d1d5f0

LEDGER-DP-C
  cmd:  grep -rn "MaybeSpawnRetentionPruner\|retention_spawn" internal/harness internal/cli
  out:  (no output)
  exit: 1
  why:  the spawn gate and its file do not exist; the only code that can
        satisfy this grep is what M2 creates.
  tree: d05d1d5f0

LEDGER-DP-D
  cmd:  go test -list 'TestMaybeSpawnRetentionPruner|TestSpawnGateSuppressesOnFreshStamp|TestSpawnFailureFailOpen|TestDetachedChildPrunes|TestDetachedChildDoubleSpawnCollapses' ./internal/harness
  out:  ok  	github.com/modu-ai/moai-adk/internal/harness	0.603s
  exit: 0
  why:  package-wide corroboration — none of the five gate/child AC test
        names exists yet (AC-DP-002/004/005/006).
  tree: d05d1d5f0

LEDGER-DP-E
  cmd:  ls internal/harness/retention_spawn_unix.go internal/harness/retention_spawn_windows.go
  out:  ls: internal/harness/retention_spawn_unix.go: No such file or directory
        ls: internal/harness/retention_spawn_windows.go: No such file or directory
  exit: 1
  why:  the platform-split detached-exec files of REQ-DP-006 do not exist;
        M2 creates both.
  tree: d05d1d5f0

LEDGER-DP-F
  cmd:  grep -rn "SysProcAttr" internal/harness
  out:  (no output)
  exit: 1
  why:  no detached-process attribute precedent exists in the package — the
        detached spawn is genuinely new code (the delegation's premise,
        re-measured).
  tree: d05d1d5f0

LEDGER-DP-G
  cmd:  go run ./cmd/moai hook retention-prune --log /tmp/x.jsonl --archive /tmp/arch
  out:     ERROR

          Unknown flag: --log.

          Try --help for usage.

        exit status 1
  exit: 1
  why:  the child verb of REQ-DP-003 is not registered — the invocation dies
        at the parent command's flag parsing instead of performing a prune.
        Red because the verb does not exist; flips only when M2 registers it
        (then --help exits 0 and the verb runs the prune path).
  tree: d05d1d5f0

LEDGER-DP-H
  cmd:  go test -list 'TestHookRetentionPruneVerb|TestHarnessObserveGateWiring' ./internal/cli
  out:  ok  	github.com/modu-ai/moai-adk/internal/cli	1.531s
  exit: 0
  why:  package-wide corroboration for AC-DP-002/003's cli-side test names —
        the selector matches no test in the package.
  tree: d05d1d5f0

LEDGER-DP-GREEN-A
  (AC-DP-001's GREEN commands, carried here rather than in the §C table:
  a table cell cannot hold a bare pipe and the escaped form is a literal
  pipe in Go regexp — the iter1 D1 defect. The `\|` in the first control
  command below is the quoted input bytes of the control itself.)

  GREEN (M1, count-first — the plan M1 exit's form):
  cmd:  go test -list '^(TestRecordExtendedEventDoesNotPrune|TestRecordEventDoesNotPrune)$' ./internal/harness
  expect:  lists exactly the two new tests — non-empty, exact count; an
           empty listing is a FAIL of this cell, not a pass
  cmd:  go test -run '^(TestRecordExtendedEventDoesNotPrune|TestRecordEventDoesNotPrune)$' ./internal/harness
  expect:  exit 0 with the tests executed; a run printing the runner's
           no-tests token is a FAIL of this cell, not a pass
  plus:  grep -c "PruneStaleEntries" internal/harness/observer.go → 0, exit 1
         (kept in the §C cell — no metacharacter); go test ./internal/harness
         exit 0 (the F5 drift net)

  syntax control (measured this tree d05d1d5f0 — the auditor's iter1
  positive control, re-executed by the authoring lane):
  cmd:  go test -run '^(TestPruneStaleEntriesRemovesOldEvents\|TestRecordEvent)$' ./internal/harness
  out:  ok  	github.com/modu-ai/moai-adk/internal/harness	(cached) [no tests to run]
  exit: 0
        ← the escaped form is the vacuous shape: zero tests selected, exit 0
  cmd:  go test -count=1 -v -run '^(TestPruneStaleEntriesRemovesOldEvents|TestRecordEvent)$' ./internal/harness
  out:  === RUN   TestPruneStaleEntriesRemovesOldEvents
        --- PASS: TestPruneStaleEntriesRemovesOldEvents (0.01s)
        PASS
        ok  	github.com/modu-ai/moai-adk/internal/harness	0.769s
  exit: 0
        ← the raw form SELECTS AND RUNS the existing test
  tree: d05d1d5f0

LEDGER-DP-FORM
  (grammar form-control for the grep BRE alternation used by LEDGER-DP-C
  and plan.md M3's boundary grep — the same two-character sequence as the
  D1 defect with a different, correct grammar; recorded so a zero hit can
  never read as a broken pattern. Measured this tree, d05d1d5f0.)
  cmd:  grep -c "PruneStaleEntries\|pruneSkipDuration" internal/harness/retention.go
  out:  8
  exit: 0
  why:  the escaped alternation matches known identifiers, so the form is
        live: LEDGER-DP-C's zero and plan M3's expected zero are absence
        findings, not dead-pattern artifacts.
  tree: d05d1d5f0

LEDGER-DP-GREEN-B
  (AC-DP-003's GREEN commands, carried in the ledger per the §A rule. The
  codex-review P2 against the original exit-0-only cell: an exit code alone
  cannot detect the verb failing to register — the hook dispatcher exits 0
  on unknown event names — so the registration half keys on the verb's own
  help LISTING CONTENT and the effect half on the test's prune assertions,
  never on an exit code alone. Registration keys on the verb-OWN help, not
  the parent listing, because the verb registers cobra-Hidden (spec D1) and
  a hidden verb never appears in the parent help.)

  GREEN registration half (M2):
  cmd:  go run ./cmd/moai hook retention-prune --help
  expect:  the output is the verb's OWN help — its usage line and Long text
           carry `retention-prune`; the parent hook help (today's output on
           an unregistered tree) does not
  decisive stage, measured red value this tree d05d1d5f0 (unregistered):
  cmd:  go run ./cmd/moai hook retention-prune --help | grep -c "retention-prune"
  out:  0
  exit: 1
        flips to >= 1 / exit 0 only when M2 registers the verb — the parent
        help cannot carry the string, so no unrelated change can flip it
  form note: the decisive stage is a pipe, outside §2.1's single-invocation
  letter — the LEDGER-ACR-J judgment-held disposition applies (four elements
  present, re-executing every round); the single-invocation form is the raw
  command above, readable directly for the missing usage line.
  tree: d05d1d5f0

  GREEN effect half (M2/M3, count-first):
  cmd:  go test -list '^(TestDetachedChildPrunes)$' ./internal/harness
  expect:  lists exactly 1 (LEDGER-DP-D corroborates the name is new)
  cmd:  go test -run '^(TestDetachedChildPrunes)$' ./internal/harness
  expect:  exit 0 — the test drives the verb's run function against a temp
           log seeded with an over-retention entry and asserts the prune's
           observable effects: the kept-line count shrinks and an archive
           member (<YYYY-MM>.jsonl.gz) is written; the run function enters
           through Retention.PruneStaleEntries (spec D5)
```

## §C Acceptance Criteria

| AC | Covers REQ | RED-now (ledger) | GREEN path (milestone + command) | Classification |
|----|-----------|------------------|----------------------------------|----------------|
| AC-DP-001 | REQ-DP-001 | LEDGER-DP-A (count `2`, exit 0 — the calls exist), LEDGER-DP-B (new tests corroborated absent) | M1: LEDGER-DP-GREEN-A (count-first — `-list` lists exactly the two tests, THEN `-run` exit 0 with tests executed; raw-pipe alternation, syntax control recorded in-ledger); `grep -c "PruneStaleEntries" internal/harness/observer.go` → `0` exit 1; `go test ./internal/harness` exit 0 | release-blocking |
| AC-DP-002 | REQ-DP-002, REQ-DP-004 | LEDGER-DP-C (gate absent), LEDGER-DP-D, LEDGER-DP-H (test names absent) | M2+M3: `TestMaybeSpawnRetentionPruner`, `TestSpawnGateSuppressesOnFreshStamp`, `TestSpawnFailureFailOpen` exit 0 — gate reads the stamp lock-free once, spawns only on stale-or-absent, suppresses on fresh, and a seam failure returns the error the wrapper logs at exit 0 | release-blocking |
| AC-DP-003 | REQ-DP-003 | LEDGER-DP-G (verb dies at flag parsing, exit 1), LEDGER-DP-H | M2+M3: LEDGER-DP-GREEN-B — registration half keys on the verb's OWN help LISTING CONTENT (usage line carrying `retention-prune`; an exit code alone is not the gate, and the parent help cannot flip it because the verb registers Hidden — codex P2), effect half is `TestDetachedChildPrunes` exit 0 count-first asserting the prune's observable effects (kept-line shrink + archive member) | release-blocking |
| AC-DP-004 | REQ-DP-003 (double-check), REQ-DP-005 (orphan harmlessness) | LEDGER-DP-D (`TestDetachedChildDoubleSpawnCollapses` absent) | M3: `TestDetachedChildDoubleSpawnCollapses` exit 0 — two children against one stamp file: the second reads the fresh stamp under the lock (`pruneLocked`, retention.go:187) and exits without a second rewrite; the attempt stamp written before the work (retention.go:195) keeps a killed/orphaned pruner from repeating within the interval | release-blocking |
| AC-DP-005 | REQ-DP-006 | LEDGER-DP-E (files absent), LEDGER-DP-F (no SysProcAttr anywhere in the package) | M2 files exist; M4: `GOOS=windows GOARCH=amd64 go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go vet ./internal/harness ./internal/cli` exit 0. Windows runtime behavior of the detached child stays documented-unobserved (spec §F F3) | release-blocking (build+vet half); runtime half unobserved-by-declaration |
| AC-DP-006 | REQ-DP-007 (seam), REQ-DP-002 (wrapper) | LEDGER-DP-C, LEDGER-DP-D, LEDGER-DP-H | M3: the seam is a function field replaced by a recording fake (asserted inside `TestMaybeSpawnRetentionPruner`/`TestSpawnGateSuppressesOnFreshStamp`); no test spawns a real detached child (plan M3 boundary grep: no `exec.Command` invocation from test files on the spawn path); the four handlers reach the gate through ONE wrapper (`TestHarnessObserveGateWiring`) | release-blocking |
| AC-DP-007 | REQ-DP-008 (semantics preserved) | none — the 1-hour interval, stamp-before-work, and atomic archive are correct on this tree today (`pruneSkipDuration = time.Hour`, retention.go:22; `TestPruneSkipsIfRecentlyPruned` green); no input turns it red before the work | M1/M2 re-verification: `go test ./internal/harness` exit 0 including the existing `TestPruneStaleEntries*` family and the t1467 M1 `retention_archive_atomic_test.go` suite — the semantics must still pass on the changed tree | regression-guard (never release-blocking — §A disposition) |

## §D Quality Gates and Definition of Done

- Every release-blocking AC's GREEN cell observed with verbatim command + output + exit code, recorded in progress.md §E.1, before run-phase exit.
- `go run ./cmd/moai spec lint SPEC-HARNESS-DETACHED-PRUNE-001 --strict` exit 0 (M4).
- `golangci-lint run` (CI-pinned v2.1.6): no NEW issues attributable to this SPEC (pre-existing baseline reported separately, per the E5 discipline).
- `go test -cover ./internal/harness` ≥ 85%.
- REQ-HL-001's contract (under 100 ms, never blocks) holds on the gate path: the wrapper's added synchronous work is one stamp read plus one spawn call — asserted by `TestSpawnFailureFailOpen` and the wrapper's construction; the parent verdict's under-5s-budget observation is the motivation, not a re-measured gate here.
- The detachment property (REQ-DP-005's parent-exits-without-wait / child-reparents) is verified by construction only, and deliberately carries no AC: no input on the pre-implementation tree turns it red (nothing spawns before M2 — LEDGER-DP-C), and REQ-DP-007 forecloses the runtime experiment that would observe reparenting. The gate path's no-`Wait` shape and the seam's recorded argv are the construction evidence (iter1 D4 disposition; plan.md §D carries the same line).
