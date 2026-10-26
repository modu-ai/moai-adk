---
id: SPEC-JEV-LEDGER-001
title: "Acceptance — Jev decision ledger recorder and per-consumer config surface"
version: "0.2.0"
created: 2026-10-26
updated: 2026-10-26
---

# Acceptance — SPEC-JEV-LEDGER-001

**Document-level tree pin:** every RED-now cell below was measured on `0c2d2aa4c230d2a6671037cefea260ede18e295d` (§L). No criterion carries its own pin. Each RED-now cell cites an evidence-ledger entry holding the command, the verbatim stdout, and the exit code.

Each criterion has the Given-When-Then, a RED-now cell (§L) saying why it is red, a green path naming the milestone that flips it and the passing output, and the three-part check spec: (a) WHEN, (b) the failing INPUT, (c) who sees the red. Release-blocking criteria also carry a mutant probe and a continued-firing answer; §C gives the per-family continued-firing table (`.claude/rules/moai/development/verification-completeness.md` §1.2, §1.3, §2, §2.1).

Migrated criteria (AC-002, -003, -004, -005, -008, -015, -016, and parts of -001, -006, -012, -013, -014) are listed in spec.md §D.1.

---

## §A. Criteria

### AC-JEVL-001 — nothing on disk until an append (REQ-JEVL-001) [release-blocking]

- **Given** a temp project root whose `workflow.yaml` sets `workflow.jev.enabled` to false in one run and true in another, with two consumers configured.
- **When** (i) the configuration is loaded through the real loader and each consumer is resolved, and (ii) a recorder is constructed for that root and no append is made.
- **Then** no `.moai/state/jev` path exists under the root after (i) or (ii), in either gate setting. Positive control in the same test: one append on the same recorder creates `.moai/state/jev/decisions.jsonl` holding exactly one line, so the zero is attributable to the absence of an append and not to a recorder that cannot write.
- RED-now: §L E5 (`grep -rn decisions.jsonl internal` → empty, exit 1). Red because no recorder and no ledger path exist; it flips at M2, when `internal/jevledger/ledger.go` names the path.
- Green path: config half `--- PASS: TestLoadWorkflow_JevCreatesNothing ` passes at M1 and is recorded as partial. **Final at M2:** that line plus `--- PASS: TestLedger_NothingUntilAppend `.
- Check spec: (a) every CI run, `go test ./internal/config/ ./internal/jevledger/`. (b) failing input: a recorder whose constructor runs `MkdirAll` eagerly → the directory exists after (ii) → red; a loader that touches `.moai/state/jev` while resolving → red after (i). (c) `go test` exit 1 in the required Test job.
- Mutant probe: a recorder that never writes anything satisfies "nothing exists" — the positive control turns it red. A recorder that creates the directory but defers the file until the first row passes (i)/(ii) only if it also creates nothing at construction, which is the requirement; no surviving mutant found.
- Continued firing: both test names are listed in `progress.md` §E.2 at run time; a dropped or renamed test removes a named `--- PASS` line the self-verification matrix requires.

### AC-JEVL-006 — mode enum and clamp at the resolver (REQ-JEVL-006)

- **Given** consumer configs with `mode` = `shadow`, `advise`, `steer`, `confirm`, absent, and `bogus`.
- **When** the resolver runs on each.
- **Then** the effective mode is `shadow` in all six, the configured value is returned verbatim (empty when absent), and only `bogus` yields a warning.
- RED-now: §L E1 (`grep -n Consumers internal/config/types.go` → empty, exit 1). Red because no consumer config type exists.
- Green path: final at M1. `--- PASS: TestResolveJevConsumer_Mode/* ` (6 subtests; fewer than 6 PASS lines is a gap).
- Mutant probe: a resolver that passes `advise` through as effective → red.
- Check spec: (a) CI. (b) the pass-through mutant. (c) Test job.

### AC-JEVL-007 — threshold bounds and per-consumer quarantine through the real loader (REQ-JEVL-007, REQ-JEVL-008) [release-blocking]

- **Given** a temp project whose `.moai/config/sections/workflow.yaml` sets `workflow.jev.enabled: true`, one other non-default `workflow` setting, and seven consumers whose `threshold` is `0`, `1`, `0.5`, `-0.1`, `1.5`, `"high"` (a YAML string), and absent; one further consumer has a mapping in place of `threshold`. Separately, a second temp project whose `workflow.jev.consumers` is the scalar `"x"`, with the same `enabled: true` and the same other setting.
- **When** each file is loaded through the real loader (`internal/config/loader.go`, `loadWorkflowSection`) and each consumer is resolved, with stdout, stderr, and the default `slog` handler captured.
- **Then** `0`, `1`, and `0.5` are kept; `-0.1`, `1.5`, the string, and the mapping resolve to absent, each with a warning returned as data; absent resolves to absent without a warning; `0` is distinguishable from absent. `enabled` and the other setting keep their file values, not the defaults. All eight consumers are present in the loaded map. For the scalar project, the map is empty, exactly one section-level warning is returned as data, and `enabled` and the other setting keep their file values. In both projects, captured stdout and stderr are empty and the `slog` capture holds no record.
- RED-now: §L E2 (`ls internal/config/jev_consumers.go` → no such file, exit 1). Red because the resolver file does not exist.
- Green path: final at M1. `--- PASS: TestResolveJevConsumer_Threshold/* `, `--- PASS: TestLoadWorkflow_MalformedConsumerQuarantined `, `--- PASS: TestLoadWorkflow_NonMappingConsumersQuarantined `.
- Check spec: (a) every CI run, `go test ./internal/config/`. (b) failing inputs: the mutants below. (c) Test job exit 1.
- Mutant probes: (1) an open-interval check rejecting `0` → red. (2) A plain `Threshold *float64` field: `yaml.Unmarshal` fails on `"high"`, `loadWorkflowSection` logs and keeps the defaults (`loader.go:239-241`), so `enabled` and the other setting revert and the `slog` capture is non-empty → red. A resolver-only test passes this mutant, which is why the test goes through the loader. (3) A typed `map[string]JevConsumerConfig` for `consumers` fails on `"x"` the same way → the second project goes red. (4) A resolver that writes its warning with `slog.Warn` instead of returning it → `slog` capture non-empty → red.
- Continued firing: §C row 1; a loader path change that bypassed `loadWorkflowSection` would make the "other setting keeps its file value" assertion fail, not pass vacuously.

### AC-JEVL-009 — ledger row fields (REQ-JEVL-009, REQ-JEVL-010, REQ-JEVL-006, REQ-JEVL-008) [release-blocking]

- **Given** two answered rows of one call (`model_id: jev-1.13.0`, `request_model: jev-latest`, `rubric_version: r1`, choice question p = 0.72, noul question p = 0.31, `mode: shadow`, `configured_mode: advise`, one row with `config_warning: "mode bogus"`), and one unanswered row (`model_id: ""`, `request_model: jev-latest`, probability absent). The caller supplies `action: advise` and `outcome_label: correct` on every row.
- **When** the recorder appends the three rows.
- **Then** each line decodes to an object holding `model_id`, `request_model`, `rubric_version`, the matching `question_id`, `probability` equal to the supplied raw value, `mode: "shadow"`, `configured_mode: "advise"`, `action: "none"`, `outcome_label: "unlabeled"`, plus `ts`, `consumer`, `event`, `harness`, `availability`, `input_tokens`, and `call_id`. The warned row carries `config_warning: "mode bogus"`; the others carry no `config_warning`. On the unanswered row, `model_id` is `""` and `probability` is `null`. No row's `model_id` equals `jev.ModelID`.
- RED-now: §L E4 (`grep -rn outcome_label internal` → empty, exit 1). Red because no row type exists anywhere in `internal`.
- Green path: final at M2. `--- PASS: TestLedgerRow_Fields `, `--- PASS: TestLedgerRow_UnansweredNullProbability `, `--- PASS: TestLedgerRow_RecorderOwnsActionAndOutcome `.
- Check spec: (a) every CI run, `go test ./internal/jevledger/`. (b) failing inputs: the mutants below. (c) Test job exit 1.
- Mutant probes: (1) a recorder that fills an empty `model_id` from `request_model` writes `jev-latest` on the unanswered row → red. (2) A recorder that passes the caller's `action` / `outcome_label` through → `advise` / `correct` → red. (3) A recorder that writes `0` for an absent probability → not `null` → red.
- Continued firing: the three names are listed in `progress.md` §E.2; the field assertions decode every line, so an empty file fails the line-count precondition instead of passing.

### AC-JEVL-010 — append, permissions, no interleaving (REQ-JEVL-011)

- **Given** a temp project root.
- **When** 50 goroutines append one row each.
- **Then** the file has 50 lines and each parses as JSON. On POSIX the directory mode is 0700 and the file mode is 0600.
- RED-now: §L E5. Red because no ledger path exists.
- Green path: M2. `--- PASS: TestLedger_ConcurrentAppend `, `--- PASS: TestLedger_Modes ` (skipped by name on windows; the file compiles there because it uses `os.FileMode` only).
- Check spec: (a) CI on linux and darwin. (b) mutant: a buffered writer flushing partial lines → a parse failure. (c) Test job.

### AC-JEVL-011 — ledger path stays ignored (REQ-JEVL-011) [regression-guard]

- **Given** the repository `.gitignore`.
- **When** `git check-ignore -v .moai/state/jev/decisions.jsonl` runs.
- **Then** exit 0 with a matching rule.
- RED-now: none. §L E10 shows the path is already ignored (`.gitignore:401:.moai/state/`). By the two-cell discipline this criterion would be vacuous as a gate, so it is a **regression-guard** and is not counted as a pass for this work.
- Green path: unchanged while `.moai/state/` stays ignored in both `.gitignore` copies.
- Check spec: (a) plan time and the M2 self-verification. (b) removing `.moai/state/` from `.gitignore` → exit 1. (c) the self-verification table.

### AC-JEVL-012 — write failure and non-regular ledger path are dropped silently (REQ-JEVL-011, REQ-JEVL-012) [release-blocking]

- **Given** four project roots, each tested by calling the recorder directly (the hook-level twin migrated to -002):
  - (i) `.moai/state/jev` is a regular file whose bytes are recorded first, so `mkdir` fails.
  - (ii) `.moai/state/jev/decisions.jsonl` is a FIFO with no reader.
  - (iii) `.moai/state/jev/decisions.jsonl` is a symbolic link to a file outside the root whose bytes are recorded first.
  - (iv) `.moai/state/jev/decisions.jsonl` is a FIFO **with an open reader**: before the append, the test opens the read side with `O_RDONLY|O_NONBLOCK` and keeps it open through the call.
- **When** one append of a valid row runs in each root, with stdout and stderr captured.
- **Then** in every case the append returns a dropped result and does not panic, and captured stdout and stderr are empty. (i): the file at `.moai/state/jev` keeps its recorded bytes and no `decisions.jsonl` exists under the root. (ii): the path is still a FIFO. (iii): the link is still a link and its target's bytes equal the recorded bytes. (iv): the path is still a FIFO, and a non-blocking read on the test's reader after the append returns zero bytes — no byte of the row reached the pipe.
- **Platform split (audit R1).** Cases (ii) and (iv), the FIFO fixture, and every Unix-only flag live in `internal/jevledger/fifo_unix_test.go` under `//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly`, the constraint on `internal/runtime/regular_file_race_unix_test.go:1`. Its counterpart `internal/jevledger/fifo_other_test.go` carries the complement constraint and one test that skips with the named reason "named pipes unsupported on this platform", following `internal/cli/audit_plan_cmd_fifo_other_test.go:1`. Cases (i) and (iii) are portable and live in `ledger_test.go`; (iii) skips by name where `os.Symlink` fails. **A run-time `t.Skip` inside an untagged file is not sufficient**: `syscall.Mkfifo` and `syscall.O_NOFOLLOW` do not exist on windows (§L E11, E12), so such a file fails to compile there and no test in the package runs at all. The DoD (§D) therefore checks the windows test compile directly.
- **Hang guard and cleanup (FIFO cases).** Each FIFO lives under `t.TempDir()`. The append runs in a goroutine, and the test waits with `select` against a 2 s timer; on expiry the test fails instead of hanging. In case (ii), a `t.Cleanup` registered before the call opens the read side with `O_RDONLY|O_NONBLOCK` and closes it, so a mutant blocked in a write open is released and the goroutine ends. The 2 s timer is a hang guard for the test, not a timing claim about the recorder. The package `go test -timeout` is the outer bound.
- RED-now: §L E6 (`ls internal/jevledger/open_unix.go` → no such file, exit 1) and §L E7 (`ls internal/jevledger/fifo_unix_test.go` → no such file, exit 1). Red because neither the write-side open nor the platform-split fixture exists; both flip at M2.
- Green path: final at M2. On linux/darwin: `--- PASS: TestLedger_WriteFailureDropped `, `--- PASS: TestLedger_NonRegularPathDropped/symlink `, `--- PASS: TestLedger_FIFODropped/no-reader `, `--- PASS: TestLedger_FIFODropped/with-reader `. On windows: `--- SKIP: TestLedger_FIFOCasesUnsupportedHere` with the named reason.
- Check spec: (a) every CI run, `go test ./internal/jevledger/`, all four cases on linux and darwin. (b) failing inputs: the mutants below. (c) Test job exit 1; for the compile split, `GOOS=windows go vet ./internal/jevledger/` non-zero.
- Mutant probes:
  1. The drop printed to stderr → captured stderr non-empty → red.
  2. The plain open without `O_NONBLOCK` → case (ii) blocks; the timer fires → red; the cleanup releases the goroutine.
  3. An open without `O_NOFOLLOW` and without a handle check → case (iii) appends to the link target → bytes differ → red.
  4. **Keeps `O_NOFOLLOW|O_NONBLOCK` and drops the opened-handle regular-file check** (audit R2) → cases (i)-(iii) still pass, because each is refused before the check; case (iv) opens successfully, the row is written into the pipe, and the test's reader reads the row's bytes → red. This is the case that exists to catch this mutant.
  5. An open without `O_NOFOLLOW` that keeps the handle check → case (iii) follows the link to a regular file, the check passes, and the target is appended to → red.
  6. FIFO fixture placed in the untagged `ledger_test.go` with only a run-time skip → `GOOS=windows go vet ./internal/jevledger/` fails with `undefined: syscall.Mkfifo` → the DoD command is non-zero → red.
- Continued firing: the four linux/darwin names and the windows SKIP name are listed in `progress.md` §E.2. On windows the SKIP line is what shows the FIFO cases are absent there; a missing SKIP line is a gap.

### AC-JEVL-013 — budget constants and their derivation (REQ-JEVL-013)

- **Given** the compiled defaults.
- **When** the constants test runs.
- **Then** the budget is 5 s, the margin is 1 s, the derived offset at the defaults is 4 s, and the same derivation applied to a 250 ms budget and a 50 ms margin yields 200 ms.
- RED-now: §L E8 (`grep -n JevHook internal/config/defaults.go` → empty, exit 1). Red because no budget constant exists. The stem is fixed by M1; if M1 picks a name without `JevHook`, M1 re-points this cell's command in the same commit.
- Green path: final at M1. `--- PASS: TestJevHookBudgetConstants `.
- Mutant probes: (1) margin declared as 0.5 s → red. (2) a derivation that returns a stored 4 s instead of computing budget − margin → the 250/50 ms assertion gets 4 s ≠ 200 ms → red.
- Check spec: (a) CI. (b) either mutant. (c) Test job.

### AC-JEVL-014 — boundary: no request path, no credential, no read-back (REQ-JEVL-014, REQ-JEVL-001, REQ-JEVL-011)

- **Given** the tree after M2.
- **When** an import-set test parses every non-test Go file in `internal/jevledger` and `internal/config/jev_consumers.go`, and the existing `TestJevCallPath_HasExactlyTheDeclaredConsumers` runs unchanged.
- **Then** every import is standard library, and none is `net`, `net/http`, `github.com/modu-ai/moai-adk/internal/jev`, or `github.com/modu-ai/moai-adk/internal/jevcred`. No non-test file in `internal/jevledger` contains `os.Open(`, `os.ReadFile(`, or `O_RDONLY`. The existing allow-list test passes with its allow-list unchanged. Positive controls: the classifier flags `net/http`, and the scan counts at least one file.
- RED-now: §L E3 (`ls internal/jevledger` → no such file, exit 1). Red because the package does not exist.
- Green path: final at M2. `--- PASS: TestLedgerPackage_ImportBoundary `, `--- PASS: TestLedgerPackage_NeverReadsLedger `, `--- PASS: TestJevCallPath_HasExactlyTheDeclaredConsumers `.
- Mutant probes: (1) the recorder imports `internal/jev` → red. (2) the recorder opens the ledger with `os.Open` to check its last byte → red. The read-back check is lexical; a read through another spelling is a residual risk owned by review.
- Check spec: (a) CI. (b) either mutant. (c) Test job.

### AC-JEVL-017 — template default (REQ-JEVL-017)

- **Given** the distributed template `internal/template/templates/.moai/config/sections/workflow.yaml`.
- **When** it is parsed.
- **Then** `workflow.jev.enabled` is false and `workflow.jev.consumers` is an empty mapping.
- RED-now: §L E9 (`grep -n consumers internal/template/templates/.moai/config/sections/workflow.yaml` → empty, exit 1). Red because the template block does not exist.
- Green path: final at M1. `--- PASS: TestTemplateJevConsumersEmpty `, plus the DoD `make build` line.
- Mutant probe: a template shipping a sample consumer entry → map not empty → red.
- Check spec: (a) CI. (b) the sample-entry mutant. (c) Test job.

### AC-JEVL-018 — append after malformed lines (REQ-JEVL-018)

- **Given** an existing ledger holding one valid row line, then the line `not json`, then the line `{"truncated":`, each ending in `\n`, with its bytes recorded.
- **When** the recorder appends one row.
- **Then** the file's first N bytes equal the recorded bytes, exactly one line follows them, and that line decodes as a full row.
- RED-now: §L E5. Red because no recorder exists.
- Green path: final at M2. `--- PASS: TestLedger_AppendAfterMalformedLines `.
- Mutant probes: (1) an open with `O_TRUNC` → prefix differs → red. (2) a recorder that drops the row when the file is non-empty and not valid JSON throughout → no new line → red.
- Check spec: (a) CI. (b) either mutant. (c) Test job.

---

## §B. Edge cases covered above

- Unanswered rows: AC-009.
- Windows: AC-010 modes skip by name; AC-012 FIFO cases skip by name through the tagged counterpart; the windows compile is checked by the DoD.
- A FIFO with a reader: AC-012 (iv).
- Malformed lines already in the ledger: AC-018.

## §C. Continued-firing answers (§1.3)

| Check family | If it stopped firing tomorrow, what would be different? |
|---|---|
| Go tests in `internal/config` and `internal/jevledger` (all ACs except 011) | Each test name is unique and listed in `progress.md` §E.2. A selector change that drops one shows as a missing named `--- PASS` line, which the self-verification matrix treats as a gap, not a pass. The required CI Test job runs whole packages, so a local selector has no effect there. |
| Path and zero-result tests (AC-001, AC-012, AC-014) | Each zero-result assertion carries its own positive control in the same test: AC-001's single append creates the file; AC-012 (iv)'s reader is opened before the call, so a zero read is attributable to the recorder; AC-014's scan must count at least one file and its classifier must flag `net/http`. A test that silently fell onto a dead fixture fails on the control instead of passing. |
| Windows compile of the FIFO split (AC-012) | The windows SKIP line `TestLedger_FIFOCasesUnsupportedHere` appears in the windows job's output. If the tagged pair were removed, the windows run would either fail to compile (DoD red) or lose that SKIP line. |
| Regression guard (AC-011) | Not a CI test. If it stopped being checked nothing would differ; the template `.gitignore` copy (`internal/template/templates/.gitignore:243`) is the only other carrier. Stated as a residual risk. |

## §D. Quality gates and Definition of Done

- All release-blocking ACs (001, 007, 009, 012) PASS at their final milestone (plan.md §E closure map) with verbatim `--- PASS` lines in `progress.md` §E.2. A partial-milestone PASS line is progress evidence, not closure.
- `go build ./...` exits 0 and `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- **Windows test compile:** `GOOS=windows go vet ./internal/jevledger/ ./internal/config/` exits 0. This is the command that compiles the test files for windows; `go build` does not compile test files and cannot stand in for it. Baseline before M1: `GOOS=windows go vet ./internal/config/` → `exit=0` (§L B1).
- `go test -cover` ≥ 85% for `internal/jevledger` and for `internal/config/jev_consumers.go`.
- `golangci-lint run` shows no new findings against the baseline.
- `make build` exits 0 after the template edit, and `git status --short` afterwards lists only files named in plan.md §E.
- No file in `internal/jev` and no file in `internal/cli` is changed by this SPEC: `git diff --name-only 0c2d2aa4c2 HEAD -- internal/jev internal/cli` prints nothing at the M2 self-verification (exit 0, empty stdout), measured before any absorption of other cards' commits.
- No user-facing question channel in the new code (C-HRA-008 grep returns 0).

---

## §L. Evidence ledger (RED-now cells)

**Tree:** `0c2d2aa4c230d2a6671037cefea260ede18e295d` (`git rev-parse HEAD` in card worktree `t1762`, 2026-10-26, 0.2.0 re-collection). This ledger supersedes every 0.1.x entry (E1-E10, E1', E3', E6' on `d1aa4ed656` / `82aff00a6d`), which belonged to the larger scope.

**Collection method.** Each command was run once, as written, in its own Bash invocation of the form `<command>; echo "exit=$?"`, with no `sh -c` wrapper. The printed `exit=N` is the cited command's own status. The **Command** column holds the bare command a reader re-runs. `ls` reports a missing path on stderr, so its stdout is empty.

| Id | Command | Verbatim stdout | Exit | Why red (or not) | Flips at |
|---|---|---|---|---|---|
| E1 | `grep -n Consumers internal/config/types.go` | *(empty)* | 1 | no consumer config type | M1 (`types.go`) |
| E2 | `ls internal/config/jev_consumers.go` | *(empty; stderr `ls: internal/config/jev_consumers.go: No such file or directory`)* | 1 | no resolver file | M1 |
| E3 | `ls internal/jevledger` | *(empty; stderr `ls: internal/jevledger: No such file or directory`)* | 1 | no recorder package | M2 |
| E4 | `grep -rn outcome_label internal` | *(empty)* | 1 | no row type anywhere in `internal` | M2 (`ledger.go`) |
| E5 | `grep -rn decisions.jsonl internal` | *(empty)* | 1 | no ledger path anywhere in `internal` | M2 (`ledger.go`) |
| E6 | `ls internal/jevledger/open_unix.go` | *(empty; stderr `ls: internal/jevledger/open_unix.go: No such file or directory`)* | 1 | no write-side regular-file open | M2 |
| E7 | `ls internal/jevledger/fifo_unix_test.go` | *(empty; stderr `ls: internal/jevledger/fifo_unix_test.go: No such file or directory`)* | 1 | no platform-split FIFO fixture | M2 |
| E8 | `grep -n JevHook internal/config/defaults.go` | *(empty)* | 1 | no budget constant | M1 (`defaults.go`) |
| E9 | `grep -n consumers internal/template/templates/.moai/config/sections/workflow.yaml` | *(empty)* | 1 | no template block | M1 (template) |
| E10 | `git check-ignore -v .moai/state/jev/decisions.jsonl` | `.gitignore:401:.moai/state/	.moai/state/jev/decisions.jsonl` | 0 | already green: regression-guard only (AC-011) | — |
| E11 | `GOOS=windows go doc syscall.Mkfifo` | `doc: no symbol Mkfifo in package syscall` (stream not separated) | 1 | premise of the R1 split, not an AC cell | — |
| E12 | `GOOS=windows go doc syscall.O_NOFOLLOW` | `doc: no symbol O_NOFOLLOW in package syscall` (stream not separated) | 1 | premise of the R1 split, not an AC cell | — |
| B1 | `GOOS=windows go vet ./internal/config/` | *(empty)* | 0 | DoD baseline: the windows test compile is green before M1 | stays green |

Each red flips only through a file this SPEC's milestones create or edit, and none depends on a file outside plan.md §E. The tool that ran these commands shows stdout and stderr together; the `ls` messages are attributed to stderr by `ls`'s convention, and E11/E12 are recorded without a stream attribution. The observation that decides each cell is the text plus the exit code.

Release-blocking cells: AC-001 → E5; AC-007 → E2; AC-009 → E4; AC-012 → E6 and E7. Each has command, stdout, exit code, and the document-level pin.
