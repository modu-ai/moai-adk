# SPEC-SYNC-SHA-SLOT-FORMAT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored in worktree `.claude/worktrees/t299`, branch `WT-sha-slot-format`,
frozen baseline `BASELINE_SHA = a6bbbf82b`. Tier S, era V3R6.

Artifacts: `spec.md`, `plan.md`, `acceptance.md`, this file. Note the Tier-S deviation: `acceptance.md`
is present although Tier S normally inlines its AC in `spec.md` §3 — written on the lead's explicit
instruction, because the regression triad's falsifiability contract needs the room.

Measurements recorded in `spec.md` §B, each with the command that produced it. Commands the plan phase
actually ran, for reproduction:

```
grep -h '^sync_commit_sha:' .moai/specs/*/progress.md | wc -l                                  → 346
… | sed 's/^sync_commit_sha:[[:space:]]*//' | grep -cE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'  → 334
… | grep -vE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'                                            → 12 (11 SPECs)
… | grep -c '#'                                                                                 → 105
grep -rh 'pending-backfill' .moai/specs/*/progress.md | grep -oE 'pending-backfill[a-zA-Z0-9-]*' | sort | uniq -c
                                                                                                → 28 spellings
diff .claude/rules/moai/development/spec-frontmatter-schema.md \
     internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md      → exit 0
```

### Plan audit iter-1 → v0.2.0 remediation

Verdict FAIL 0.80 (`.moai/reports/t299/plan-audit-iter1.md`), three blocking defects, all remediated
at spec.md v0.2.0. The central correction: the flagged set is **5**, not 12 — REQ-SSF-005's exemption
removes seven placeholders — and all five sit in `completed` SPECs, so the predicted non-advisory
count is **0**, not 2. Re-derived here:

```
python3 .moai/reports/t299/grammar_check.py          → total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
… | grep -E '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)' | grep -c '#'   → 99  (of 334 conforming)
grep -h '^mx_commit_sha:' .moai/specs/*/progress.md | wc -l         → 85
… | grep -vcE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'               → 9   (3 deliberate declarations)
grep -A6 'var eraDemotableCodes' internal/spec/lint.go              → MissingExclusions, FrontmatterInvalid
```

The `pending-backfill` family command in `spec.md` §B.3 now excludes this SPEC's own directory: the
verbatim command quoted in this file self-matches twice, which moved the cited 29 to 31 for anyone
re-running it.

Open questions carried to the operator: the nine `completed`-SPEC slots (repair or leave — `spec.md`
§E) and the t354 coordination recommendation (`spec.md` §D.4).

## §E.2 Run-phase Evidence

Run phase executed in worktree `.claude/worktrees/t299`, branch `WT-sha-slot-format`, entered at HEAD
`09bb632e6`. Frozen baseline for every diff-based criterion: `BASELINE_SHA = a6bbbf82b` (the literal
recorded at plan phase, not re-resolved from a moving ref). Milestone commits:

| Milestone | Commit | Subject |
|---|---|---|
| M1 | `ccf307dac` | value grammar + shared predicate |
| M2 | `d65eef053` | invert the closer's backfill test |
| M3 | `902290416` | SyncSHASlotFormat lint rule |

### E.2.1 — Inventory 1: the five findings (observed, NOT repaired)

Measured, not predicted. `go run ./cmd/moai spec lint --strict --json` in this tree, filtered to
`SyncSHASlotFormat` — the full JSON is at `.moai/reports/t299/corpus-lint.json` and the extract at
`.moai/reports/t299/ac-ssf-006-findings.txt`:

```
$ jq -r '.[] | select(.code=="SyncSHASlotFormat") | "\(.file):\(.line)  advisory=\(.advisory // false)  severity=\(.severity)"'
…/.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md:47   advisory=true  severity=warning
…/.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md:259  advisory=true  severity=warning
…/.moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/progress.md:103       advisory=true  severity=warning
…/.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/progress.md:45         advisory=true  severity=warning
…/.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/progress.md:149            advisory=true  severity=warning
```

Values held, per the plan-phase classifier: `null`, `<pending>`, `TBD`, `null`, `pending`. All five
owning `spec.md` files carry `status: completed`.

**This is the observed-not-repaired list.** Whether closed history is rewritten to satisfy a rule
written afterwards is an operator decision, not a lane's — it is the same judgment `applyEraDemotion`
already encodes by demoting terminal-status findings. This card changed none of them; §E.2.4 measures
that.

### E.2.2 — Inventory 2: the seven exempt placeholders (a scheduling list, NOT a work item)

These are the slots REQ-SSF-005 requires the rule to stay SILENT about. They are not findings and not
a defect list: each is a slot still owed a real SHA, so this is the inventory of SPECs whose OWN close
will populate them. Useful to the lead for scheduling; never a work item for this lane.

| SPEC | Token held |
|---|---|
| `SPEC-AUDIT-SNAPSHOT-001` | `pending-backfill-SYNC` |
| `SPEC-BACKLOG-LOCK-BUDGET-001` | `pending-backfill-sync` |
| `SPEC-INFINITE-GOAL-001` | `pending-backfill` |
| `SPEC-UPDATE-YAML-PRESERVE-001` | `pending-backfill-after-merge` |
| `SPEC-V3R6-AUDIT-MODEL-PIN-001` | `pending-backfill-sync` |
| `SPEC-V3R6-LIFECYCLE-CLOSE-THREEPHASE-001` | `pending-backfill-sync` |
| `SPEC-VERIFICATION-COMPLETENESS-001` | `pending-backfill` |

`SPEC-BACKLOG-LOCK-BUDGET-001` is card t354's SPEC and `SPEC-V3R6-AUDIT-MODEL-PIN-001` is
`implemented`; both are named in `spec.md` §E as explicitly out of scope, and neither was touched.

Note the shape of the whole picture, because reading it wrong is the anticipated mistake: **twelve
non-SHA values, five findings.** Seven are exempt, not missing. A run reporting 12 findings has broken
AC-SSF-003.

### E.2.3 — The `mx_commit_sha` observation the operator asked for

`spec.md` §E ACCEPTS, as out-of-scope, that three deliberate `mx_commit_sha` declarations enter the
backfill predicate's scope after the inversion. That the INVERTED IMPLEMENTATION actually behaves that
way was not observed at plan time. It is observed now, by running the implemented predicate against
the three literal values — `TestNeedsSHABackfill_MxDeliberateDeclarationsInScope` in
`internal/spec/closer_syncsha_test.go`:

| Value | Owning SPEC | `needsSHABackfill` BEFORE (allowlist) | AFTER (inverted) |
|---|---|---|---|
| `<NA>` | `SPEC-CLIFIX-CONCURRENCY-001` | false | **true** |
| `_<pending Mx-phase>_` | `SPEC-V3R6-BASH-RISK-GOVERNANCE-001` | false | **true** |
| `_(not applicable — this SPEC removes the Mx-phase concept; REQ-LR-004/007)_` | `SPEC-V3R6-LIFECYCLE-REDESIGN-001` | false | **true** |

The BEFORE column is measured, not assumed: the same test was observed RED against the pre-inversion
predicate (`.moai/reports/t299/mutations/ac-ssf-004-red.txt` carries the same run's failure block for
the sibling criterion; the mx rows failed identically in the M2 RED capture). **The result confirms
the accepted boundary rather than contradicting it**: all three now return `true`, so a later close
would overwrite each with `(this commit)`.

**No close path was triggered against them in this run.** Measured:

```
$ git diff --stat a6bbbf82b -- .moai/specs/SPEC-CLIFIX-CONCURRENCY-001 \
    .moai/specs/SPEC-V3R6-BASH-RISK-GOVERNANCE-001 .moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001
(no output)
```

The exposure stays PROSPECTIVE, not live: all three owning SPECs are `completed`, so no close is
expected against them, and the Mx phase is retired by the third SPEC in the list — the class shrinks
rather than grows. No `mx` guard was written (`spec.md` §E); a durable "not applicable" declaration in
that field is an mx-path requirement for its own card to raise.

### E.2.4 — No corpus repair

```
$ git diff a6bbbf82b -- .moai/specs | grep -E '^[+-]sync_commit_sha:'
(no output)
$ git diff --stat a6bbbf82b -- .moai/specs
 .moai/specs/SPEC-SYNC-SHA-SLOT-FORMAT-001/{acceptance,plan,progress,spec}.md | 986 ++++++++
```

The only `.moai/specs` paths in the diff are this SPEC's own four artifacts (which did not exist at
`a6bbbf82b`). Two `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/` fixtures were rewritten in the working
tree by `go test` — a known behavior of that package's perf fixture, unrelated to this card — and were
restored with an explicit-pathspec `git restore` rather than committed.

## §E.3 Run-phase Audit-Ready Signal

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Baseline-attribution.** Every row below was measured in this run, in this tree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t299`), against HEAD `902290416`, with diff-based
criteria decided against the frozen literal `a6bbbf82b`. No figure is carried from the plan phase or
from another tree.

### AC PASS/FAIL matrix

Every criterion names the command that decided it AND the mutation that was planted, observed red, and
reverted. A criterion reported without its mutation having been run is not reported.

| AC | Verdict | Command | Observed output | Mutation run | RED evidence |
|---|---|---|---|---|---|
| AC-SSF-001 (triad a) | **PASS** | `go test ./internal/spec/ -run TestSyncSHASlot_FlagsProse` | `ok … 2.226s` — 1 finding, `progress.md`, `warning`, `Advisory=false` | fixture value → `a6bbbf82b` | `mutations/ac-ssf-001-red.txt`: `expected exactly 1 SyncSHASlotFormat finding, got 0` |
| AC-SSF-002 (triad b) | **PASS** | `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnSHA` | `ok` — 0 findings across `sha-short` / `sha-full` / `sha-quoted` / `sha-annotated` | SHA pattern → `{40}` | `mutations/ac-ssf-002-red.txt`: `sha-short`, `sha-quoted` **and** `sha-annotated` each produce 1 finding (broader than the stated prediction of two; `sha-full` correctly stays silent) |
| AC-SSF-003 (triad c) | **PASS** | `go test ./internal/spec/ -run TestSyncSHASlot_SilentOnPlaceholder` | `ok` — 0 findings on `placeholder` / `placeholder-suffixed` | placeholder branch disabled | `mutations/ac-ssf-003-red.txt`: both fixtures produce 1 finding each |
| AC-SSF-004 | **PASS** | `go test ./internal/spec/ -run TestNeedsSHABackfill_OutOfAllowlistPlaceholder` | `ok` | four-value `switch` restored | `mutations/ac-ssf-004-red.txt`: all 8 rows fail, `pending-backfill-sync` first |
| AC-SSF-005 | **PASS** | `go test ./internal/spec/ -run TestNeedsSHABackfill_LegacySetPreserved` | `ok` — all four retired values return `true` | `isCommitSHAToken("")` → `true` | `mutations/ac-ssf-005-red.txt`: `needsSHABackfill("") = false, want true` |
| AC-SSF-006 | **PASS** | `go run ./cmd/moai spec lint --strict --json`, filtered | `{"total": 5, "non_advisory": 0}` — exactly the five files+lines of §B.5.1 | severity → `error` | `mutations/ac-ssf-006-red.txt`: `{"total": 5, "non_advisory": 5, "severities": ["error"]}` |
| AC-SSF-007 | **PASS** | `grep -rn 'isCommitSHAToken' internal/spec --include='*.go' \| grep -v _test` | 1 definition (`syncsha.go:107`) + 2 non-test call sites (`closer.go:422`, `lint_syncsha.go:103`); 3 further hits are doc comments | (structural — a second inlined copy would show the lint file no longer referencing the predicate) | — |
| AC-SSF-008 (SHOULD) | **PASS** | `git diff a6bbbf82b -- internal/spec/lint.go` | one added rule entry + its comment block, 12 insertions, 0 deletions; no existing entry reordered or removed | — | — |
| AC-SSF-009 | **PASS** | `git diff --stat a6bbbf82b -- internal/spec/era.go` | (empty) | — | — |
| AC-SSF-010 | **PASS** | `grep -A6 'var eraDemotableCodes' internal/spec/lint.go` + `go test -run TestSyncSHASlotFormat_AbsentFromEraDemotableCodes` | map holds exactly `MissingExclusions` and `FrontmatterInvalid` | `"SyncSHASlotFormat": true` added | `mutations/ac-ssf-010-red.txt`: `has 3 entries, want exactly 2` |

**The regression triad is reported whole.** AC-SSF-001, -002 and -003 are one instrument in three
parts; each leg alone is satisfied by a degenerate rule, and only all three together separate a
working guard from a switched-off one.

### AC-SSF-006: the measurement matched the derivation, and the divergence check that says so

Predicted (spec.md §B.6, derived from `lint.go:61/220/284/1134` plus a classifier): **5 total /
0 non-advisory**. Measured: **5 total / 0 non-advisory**, on the same five files and lines. No
divergence to report, and the rule was not tuned toward the number — the corpus run was executed once,
after M3 landed, and its first result is the one recorded here.

The "contributes nothing to the `--strict` exit status" half is measured separately, because the
overall run exits 1:

```
$ jq -r '[.[] | select(.severity=="error" or (.severity=="warning" and (.advisory != true)))]
         | group_by(.code) | map({code: .[0].code, n: length}) | .[] | "\(.code)\t\(.n)"'
MovingRefUnpinned	12
```

The strict exit status is driven entirely by a pre-existing rule. `SyncSHASlotFormat` contributes
zero, and the AC-SSF-006 mutation shows why that is a real shelter rather than an accident: promoting
the severity to `error` moves the non-advisory count from 0 to 5 immediately.

### Verification scope

```
$ go test ./internal/spec/...                     → ok  github.com/modu-ai/moai-adk/internal/spec  36.688s
$ go test -cover ./internal/spec/...              → ok  …  coverage: 89.3% of statements
$ go vet ./internal/spec/...                      → exit 0, no output
$ golangci-lint run --timeout=5m ./internal/spec/... → exit 0, "0 issues."
$ go build ./...                                  → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...        → exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/spec/... → exit 0 (test files compile on windows)
$ go test ./internal/cli/... ./internal/epic/... ./internal/harness/router/... \
          ./internal/hook/... ./internal/navigator/tiers/... ./internal/web/...
                                                  → all ok (the six packages importing internal/spec)
```

The dependent-package set was derived rather than guessed:
`grep -rln 'internal/spec"' --include='*.go' internal/ cmd/ pkg/ | xargs -n1 dirname | sort -u`. The
full local suite was NOT run; CI decides the whole tree.

### Gaps — what was explicitly NOT observed

- **CI has not run.** The branch is unpushed by instruction (the lead owns integration), so no
  clean-environment or darwin/windows-matrix verdict exists. Local green is an early signal, not the
  verdict.
- **No SPEC was actually closed.** The closer's inverted predicate was exercised as a unit, not
  through an end-to-end `moai spec close` against a real SPEC. That the predicate returns `true` is
  measured; that a close then writes the resolved SHA correctly is inherited from the pre-existing
  close path and was not re-verified.
- **The `mx_commit_sha` overwrite was not observed happening.** §E.2.3 measures the predicate's answer
  and the absence of any close against the three SPECs; it does not demonstrate the overwrite itself,
  which is the prospective behavior `spec.md` §E accepts.
- **Reachability of any SHA value was not checked** — limitation L2, out of scope by design.
- **The fourth reader** (`internal/epic/status.go` `syncShaYAMLPattern`) was not unified into the
  shared predicate; it remains the recorded follow-up candidate of `spec.md` §E.
- **Card t357 was not investigated**, per the operator's instruction that REQ-SSF-007's premise holds
  as written (t357 M2 sets `SeverityError` at its own emission site, with no global warning→error
  promotion path). AC-SSF-010 therefore decides the requirement as written; the contingency is not
  live and was not re-derived from this tree.

### Residual-risk — what could still be wrong despite what was observed

- **The line-shape anchor is narrow by choice.** The rule matches `sync_commit_sha:` at column 0,
  matching the classifier the prediction was derived with. A slot written with leading whitespace, or
  as a markdown list item (a form `extractProgressField` DOES accept), is invisible to the rule. The
  choice keeps the measured count comparable to the predicted one; the cost is a false-negative class.
- **L1 is live in the corpus, unmeasured.** A 7-40 character all-hex English word in the slot passes
  as a SHA. No corpus scan for that class was run, so its size is unknown.
- **The `--strict` shelter depends on `terminalStatusEnum` and on the five SPECs staying `completed`.**
  If one of them re-opened (the `completed → in-progress` amendment transition), its finding would
  stop being advisory and the strict exit status would change — with no signal that this rule caused
  it.
- **Fixture era classification is inferred, not pinned.** The fixtures carry no `era:` frontmatter and
  rely on H-4 firing. `TestSyncSHASlot_FlagsProse` asserting `Advisory == false` is what makes that a
  measurement, but only the prose fixture is covered by that assertion — a silently-demoted SHA or
  placeholder fixture would still report 0 findings and read as a pass.
- **Placeholder-prefix case-sensitivity is a judgment.** `Pending-Backfill` would be flagged. No corpus
  slot uses that spelling today, so the decision is currently free of cost; a future author who
  capitalizes gets a finding rather than an exemption.

### Audit-ready block

```yaml
run_complete_at: 2026-08-29
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (branch unpushed by instruction; lead owns integration)
l44_post_push_fetch: not-run (no push performed)
new_warnings_or_lints_introduced: none (go vet exit 0; golangci-lint ./internal/spec/... "0 issues.")
cross_platform_build:
  darwin_arm64: pass (go build ./... exit 0)
  windows_amd64: pass (GOOS=windows go build ./... exit 0; GOOS=windows go vet ./internal/spec/... exit 0)
total_run_phase_files: 30 changed vs a6bbbf82b (4 SPEC artifacts + 6 Go source/test + 14 fixture + 6 evidence)
m1_to_mN_commit_strategy: one commit per milestone, three total (ccf307dac, d65eef053, 902290416), unpushed
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
