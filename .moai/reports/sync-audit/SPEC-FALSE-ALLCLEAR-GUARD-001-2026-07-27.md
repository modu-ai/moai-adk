# Sync Audit — SPEC-FALSE-ALLCLEAR-GUARD-001

**Auditor**: sync-auditor (independent; did not author the implementation)
**Date**: 2026-07-27
**Subject**: `feat/false-allclear-guard-001` @ `63c3df110`, merge-base `6763aff3b`
**Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/faclear`
**Profile**: `default` (SPEC declares no `evaluator_profile`; `harness.yaml default_profile: "default"`)

---

## Verdict

**CONDITIONAL PASS — 89.9 / 100** (weighted harmonic mean; arithmetic 90.1)

| Dimension | Weight | Score | Pass Threshold (default profile) | Verdict |
|-----------|-------:|------:|----------------------------------|---------|
| Functionality | 40% | 90 | All acceptance criteria PASS | **PASS** (must-pass) |
| Security | 25% | 95 | No Critical/High findings | **PASS** (must-pass) |
| Craft | 20% | 82 | Coverage >= 85% | PASS-with-debt |
| Consistency | 15% | 93 | No major pattern violations | PASS |

**Must-pass firewall**: both must-pass dimensions (Functionality, Security) clear independently. No Critical/High security finding. All 18 acceptance criteria PASS under independent execution.

**Why CONDITIONAL, not PASS**: the implementation is sound and I could not break it. The condition is F1 — `progress.md` §E.2 / §E.3 are still `_<pending run-phase>_` placeholders while both run-phase milestones are committed. `acceptance.md` §E Definition of Done requires the 18 criteria's verbatim output, the three falsification round trips, and the AC-FAG-017 baseline to be recorded there, and `spec-frontmatter-schema.md` assigns §E.2/§E.3 to manager-develop as **run-phase** duties. In a SPEC whose entire subject is "do not report success you did not observe", shipping the code with an empty evidence surface is the one finding I will not wave through.

---

## Findings (most severe first)

### F1 — [MEDIUM] `progress.md` run-phase evidence surface is empty — CONFIRMED

**Location**: `.moai/specs/SPEC-FALSE-ALLCLEAR-GUARD-001/progress.md` §E.2, §E.3

Both sections read `_<pending run-phase>_` while `a56bfb58c` (M1) and `63c3df110` (M2) are committed.

`acceptance.md` §E Definition of Done requires, unchecked:
- "All 18 criteria executed, verbatim command output recorded in `progress.md` §E.2."
- "AC-FAG-017 skip-count baseline measured on the unmodified tree **before** M1 begins and recorded in `progress.md` §E.2."
- "All three falsification round trips … executed and recorded" / "AC-FAG-004 both halves executed with both byte counts recorded" / "AC-FAG-011 both falsification forms … executed and recorded".

`spec-frontmatter-schema.md` § progress.md Section Map assigns §E.2 (Run-phase Evidence) and §E.3 (Run-phase Audit-Ready Signal) to **manager-develop, run-phase**. The run phase is complete, so these are overdue, not deferred.

**Failure scenario**: the SPEC closes with no attributable baseline. A future reader cannot tell which criteria were executed versus assumed — the exact `verification-claim-integrity.md` §1.1 surface-2 hazard. Concretely, the AC-FAG-017 skip baseline was never recorded, so the criterion had nothing to compare against even in principle.

**Mitigating**: I independently reproduced the full evidence set (this report). The evidence *exists*; it is simply unrecorded.

**Required fix**: populate `progress.md` §E.2 with the per-AC verbatim command output and §E.3 with the run-phase audit-ready signal. The measurements in this report's Commands section may be cited directly.

---

### F2 — [MEDIUM] AC-FAG-017's skip-count sub-assertion is vacuous — CONFIRMED

**Location**: `acceptance.md` AC-FAG-017 (lines 482-488)

```bash
go test ./... > /tmp/fag17.log 2>&1; echo "exit=$?"
grep -c '^--- SKIP' /tmp/fag17.log      # expect: <= the pre-M1 baseline
```

`go test ./...` without `-v` **never emits `--- SKIP` lines**. Measured on the real log: `grep -c '^--- SKIP'` = **0**, and `grep -c 'SKIP'` (any substring) = **0**, while the verbose run of the same tree contains **56** real `--- SKIP` lines. The count is structurally pinned at 0 regardless of how many tests skip, so the criterion cannot detect any newly-added skip.

This is one of the traps the SPEC's own §A.1 catalogues ("`--- SKIP` never appears in non-verbose output" was named in the audit brief), reproduced inside the criterion meant to enforce REQ-FAG-031. Compounded by F1: the baseline it compares against was never recorded.

**Failure scenario**: a future change converts a failing test to `t.Skip` to accommodate itself. AC-FAG-017 reports PASS. This is a false all-clear inside the false-all-clear SPEC.

**Mitigating — the underlying requirement is satisfied**: I verified REQ-FAG-031 directly rather than through the criterion. Verbose skip count is **56 at merge-base and 56 at HEAD**, and the named-skip set diff is **empty in both directions** (`comm -13` and `comm -23` both produced no output). No test was newly skipped. The risk is a missing future guard, not a present regression.

**Required fix**: change the criterion to `go test -v ./... > log; grep -c '^ *--- SKIP' log`, record the baseline (56 on this host), and note that the count is host-dependent — the two sg-gated skips this SPEC adds fire only where `sg` is absent.

---

### F3 — [LOW-MEDIUM] AC-FAG-008's `install` sub-assertion is vacuous at M2 evaluation time — CONFIRMED

**Location**: `acceptance.md` AC-FAG-008 line 280 — `grep -ci 'install' /tmp/fag8.err  # expect: >= 1`

Recorded baseline is 0, measured at `6763aff3b` (pre-M1, when slog was discarded). I built the **M1-only** commit `a56bfb58c` and ran the same command: `grep -ci install` on stderr = **1**. M1 alone makes the scanner's pre-existing `slog.Warn` hint (`"install from https://ast-grep.github.io/..."`, `scanner.go:263`) visible, so the assertion passes with **zero M2 change**.

Self-reported by the implementer; confirmed here by execution.

**Mitigating — the criterion survives on its other two sub-assertions**, both verified against the same M1-only binary:

| Sub-assertion | M1-only | HEAD | Discriminates? |
|---|---|---|---|
| exit code | **0** | **1** | YES |
| `grep -c 'no findings'` stdout | **1** | **0** | YES |
| `grep -ci install` stderr | **1** | 2 | **NO** |

**Required fix**: re-anchor to the M2-specific literal, e.g. `grep -c 'scan did not run' /tmp/fag8.err` (M1-only = 0, HEAD = 1), or re-baseline the `install` count against M1 and expect 2.

---

### F4 — [LOW] AC-FAG-015 mislabels which command is the registration guard — CONFIRMED

**Location**: `acceptance.md` AC-FAG-015 line 445 — "The second command is the registration guard: a `checkAstGrep` written but never added to `systemChecks` would leave the JSON count at 0."

The prose attributes the guard role to `grep -c 'exec.LookPath' internal/cli/doctor.go >= 3`, but describes the behavior of the *first* command. Falsified in a probe worktree by deleting the `{"ast-grep CLI", checkAstGrep}` registration while keeping the function:

- cmd1 (JSON grep): **1 → 0** — this is the registration guard.
- cmd2 (`exec.LookPath` count): **3 → 3** — unchanged; proves only that the function was written.

Harmless in effect: cmd1 does the job, and three golden tests (`TestDoctorGolden_{Light,Dark,NoColor}`) also fail on unregistration.

**Required fix**: swap the sentence to attribute the guard to the first command.

---

### F5 — [LOW] `isHookCommand` arg walk is fragile to a future value-taking global flag — code-reading inference, currently unreachable

**Location**: `internal/cli/logging.go:80-89`

```go
for _, arg := range args {
    if strings.HasPrefix(arg, "-") { continue }
    return arg == hookCmd.Name()
}
```

The walk skips every `-`-prefixed token and treats the first non-flag token as the subcommand. If a global flag that consumes a **following value** were added (`moai --config-dir /x hook pre-tool`), `/x` would be read as the subcommand, `isHookCommand` would return false, and the `moai hook` path would install the **stderr** handler — writing warn records onto the stream the Claude Code runtime consumes. That is the precise failure M1's carve-out exists to prevent.

**Refuted as currently reachable**: `rootCmd` registers no persistent flags (`grep -rn 'PersistentFlags' internal/cli/root.go` → no match), so no global flag consumes a following value today. Empirically, `moai hook pre-tool` yields **0 stderr bytes**, including under `MOAI_LOG_LEVEL=debug`.

Same shape as the pre-existing `isTrivialCommand` walk it deliberately mirrors, so this is inherited, not introduced.

**Required fix (defensive, optional)**: add a regression test asserting `isHookCommand` for a value-flag form, or a comment pinning the "no value-taking global flags" precondition.

---

### F6 — [LOW] `gateNotice` is dropped when the gate passes but a later Bash check returns `ask`/`deny` — code-reading inference

**Location**: `internal/hook/pre_tool.go:404` (assign) vs `:414-419` (early return) vs `:461-465` (emit)

`gateNotice` is set at 404, but `checkBashCommand` at 407 can early-return `NewDenyOutput` / `NewAskOutput` at 414/418, bypassing the emission at 461. The ast-grep skip notice is lost in exactly the case where a commit is being scrutinized.

Defensible — the deny/ask reason is the more urgent message and `HookOutput` carries one reason field. Recorded as an observation, not a required fix.

---

### F7 — [INFO] M1 makes 1050 bytes of warn noise visible on every non-hook invocation — CONFIRMED, root-caused

Three records now print on every non-hook `moai` command:

1. `gopls bridge disabled — quality gates fall back to CLI tools only` (`deps.go:179`)
2. `cwd fallback used (CLAUDE_PROJECT_DIR not set)` × 2 (`NewPreToolHandlerWithScanner`, `NewSubagentStartHandlerWithConfig`)

**Root-caused #1 as truthful, not a misconfiguration.** `gopls` IS installed (`/Users/goos/go/bin/gopls`), so I ran `MOAI_LOG_LEVEL=info` and got the discriminating record: `gopls bridge disabled (lsp.yaml: enabled=false)`. Confirmed at `.moai/config/sections/lsp.yaml:43` → `enabled: false`. The bridge is off by deliberate local config.

**Judgment**: acceptable UX, with one asymmetry worth a follow-up. `deps.go:131` already logs the *configured-off* case at **Info**, while `deps.go:179` re-reports the same state at **Warn**. A configured-off optional subsystem is an Info fact; Warn should be reserved for `NewBridge` failing (`deps.go:125`, already Warn). Down-levelling `deps.go:179` when the disable is configured would remove ~1/3 of the new noise without losing a real signal. The two `cwd fallback` records are genuine advisories (`CLAUDE_PROJECT_DIR` is legitimately unset outside a Claude session) and correctly at Warn.

Out of this SPEC's declared scope (§C: "This SPEC changes which records are *emitted*, not what any individual callsite says"). Follow-up candidate.

---

### F8 — [INFO] `MOAI_SG_VERSION_OVERRIDE` declared inline rather than in `envkeys.go` — judged ACCEPTABLE

`CLAUDE.local.md` §14 states `[HARD] 환경변수명 → internal/config/envkeys.go에 상수 정의 후 참조`. I tested both of the implementer's arguments:

- **Consistency (holds)**: `envkeys.go` contains only *runtime-behavior* env vars (`MOAI_CONFIG_DIR`, `MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR`, `MOAI_STATUSLINE_*`, `MOAI_SKIP_BINARY_UPDATE`, `MOAI_GLM_NO_AUTO_TOOLS`). All four *test-determinism* override vars (`MOAI_GIT_VERSION_OVERRIDE`, `MOAI_GH_VERSION_OVERRIDE`, `MOAI_GOOS_OVERRIDE`, `MOAI_GOARCH_OVERRIDE`) live inline in `banner.go`. Routing only `SG` through `envkeys.go` would make it the sole outlier in its own class.
- **Import cost (weaker than stated)**: no cycle would result — `uikit` imports `printer`/`settings`/`tui`; `config` imports only `defs`/`pkg/models`. The cost is one import, not a cycle.

**Judgment**: ACCEPT. §14's [HARD] rule is best read as binding runtime-config env names; this is a test fixture following an established four-member sub-convention. Not scored as a defect. If the rule is meant to bind test overrides too, that is a separate 4-file cleanup, not this SPEC's debt.

---

### F9 — [INFO] Four rewritten scanner tests no longer exercise their named subject on an sg-less host

`assertScanErrForRulesDirCase` (`scanner_test.go:20-40`) routes `TestScanner_SGNotAvailable`, `TestScanner_EmptyRulesDir`, `TestScanner_RulesDirNotExist`, and `TestParseSGFindings_EmptyOutput`. Because `Scan` probes availability *before* the rules-dir check, on a host without `sg` all four assert the sentinel and never reach the rules-dir logic they are named for. Honest and documented in the helper's comment; noted so CI-without-sg coverage is not overestimated.

---

## Priority-target findings — the implementer's claims, independently tested

### 1. Does the gate reason reach the user? — CONFIRMED, including on the wire

I re-ran **both** falsification forms myself in detached probe worktrees at HEAD, and added a third of my own.

| Form | Mutation | Result |
|------|----------|--------|
| **A** (implementer) | delete `out.SystemMessage = gateNotice` in `pre_tool.go` | **FAIL** at assertion 1 |
| **B** (implementer) | restore `gate.go`'s discarding pass branch + `return true, ""` | **FAIL** at assertion 1 |
| **C** (auditor-added) | revert `astgrep_gate.go` sentinel branch to `return true, ""` | **FAIL** — 6 tests |

Form B is the one that proves *propagation* rather than final assignment, and it fails — `gate.go`'s pass-path plumbing is load-bearing. Form C is caught by `TestAstGrepGateReasons`, `TestPreTool_AstGrepSkipReasonSurfaces`, `TestRunAstGrepGateV2_NoSgCLI`, `TestRunAstGrepGateV2_ProjectDirPathVariants`, `TestRunAstGrepGateV2_ScanDegradedReason`, `TestRunAstGrepGateV2_TableDriven`.

**Beyond the implementer's evidence — end-to-end on the wire.** The Go test asserts at the handler frame. I drove the real binary through the actual hook entry point with a stripped PATH and a synthetic project:

```json
{
  "systemMessage": "ast-grep scan skipped: the sg CLI was not found, so no rules ran (install from https://ast-grep.github.io/guide/quick-start.html)",
  "hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow"}
}
```

The reason reaches Claude Code's structured output, the decision is `allow` (not `deny`), and hook stderr is **0 bytes** — so the emission channel choice (`SystemMessage`, not `slog`) is correct given M1's unconditional hook carve-out. This is the single strongest piece of evidence in the audit and it was not in the implementer's claim set.

**Regression risk retired**: the gate never denies on sg-absence — verified at three frames (AC-FAG-013's two tests, AC-FAG-011 assertion 2, and Form C's `passed` assertions).

**Residual risk**: exactly one test (`TestPreTool_AstGrepSkipReasonSurfaces`) guards the three-frame chain. Forms A and B each kill only that one test. Deleting it would silently unprotect the whole propagation.

### 2. Is the sentinel distinguishable from a real scan error? — CONFIRMED, both directions

`TestScannerScan_UnavailableSentinel` passes with 3/3 subtests. Both directions verified:

- **Absent → sentinel**: `errors.Is(err, ErrScannerUnavailable)` true; message names `sg` and carries the install URL.
- **Untrusted path → NOT sentinel**: `ValidateBinary` runs *first* (`scanner.go:249`) and returns a non-sentinel error; subtest 2 asserts `errors.Is(...)` is **false**, so the sentinel cannot swallow unrelated failures.
- **Clean scan → nil**: executed here (not skipped) because `sg` is present.

Caller-side discrimination verified at runtime: with `sg` present, `moai ast-grep ./internal` exits **1** from 19,006 real findings with `grep -c 'scan did not run'` on stderr = **0**. The two exit-1 causes are distinguishable by stderr content. Ordering in `astgrep.go:94` (`errors.Is` sentinel branch first, generic wrap second) is correct.

### 3. Were the rewritten tests weakened? — NO. Inverted, and one strengthened.

All six checked by reading the diff:

| Test | Before | After | Verdict |
|---|---|---|---|
| `TestRunAstGrepGateV2_NoSgCLI` | `output == ""` | `output == astGrepReasonScannerUnavailable` | **Inverted** — fails if the gate reverts |
| `TestScanner_SGNotAvailable` | `err == nil` | sentinel when sg absent, `nil` when present | **Strictly stronger** in both branches |
| `TestScanner_EmptyRulesDir` | `err == nil` | same helper | **Strictly stronger** |
| `TestScanner_RulesDirNotExist` | `err == nil` | same helper | **Strictly stronger** |
| `TestParseSGFindings_EmptyOutput` | `err == nil` | same helper | **Strictly stronger** |
| `TestAstGrepCmd_SarifFormat` | `t.Skip` on empty output | `exec.LookPath` precondition skip + `t.Fatalf` when sg present | **Strengthened** — a silent skip became an honest precondition plus a hard failure |

No assertion was deleted or softened. `TestAstGrepCmd_SarifFormat` is a genuine improvement: it previously inferred "no sg" from empty output, which would also have swallowed a real SARIF regression; it now fails loudly when `sg` is present and stdout is empty. It also splits stdout/stderr capture, which the old merged buffer would have corrupted.

Form-A/B/C evidence above independently confirms the inverted assertions still bite.

### 4. AC vacuity sweep — 2 vacuous sub-assertions of 18 criteria

See F2, F3, F4. Every other criterion was tested for discrimination; two that looked suspicious turned out to be sound:

- **AC-FAG-005** (`grep -c 'level='` on stdout) looked vacuous because stdout is 0 bytes in its sg-absent setup. I ran the constructive falsification the criterion itself proposes — rebuilt with `dest: os.Stdout` — and the count went **0 → 4**. It discriminates. `TestLoggingHandlerSelection` also fails (2 subtests). Additionally verified under `sg` present with **38,004 lines** of real stdout: `level=` and `msg=` counts both **0**. No leakage even under heavy stdout traffic.
- **AC-FAG-001** (`grep -c 'slog.SetDefault'` = 1) does not match a comment; the single hit is real code at `logging.go:73`.

### 5. M1 stdout/stderr discipline — CONFIRMED across all three formats

`internal/cli/CLAUDE.md` requires stdout machine-readable, stderr human. Under stripped PATH, all of `--format=text|json|sarif` produce **stdout 0 bytes, 0 install-mentions, 0 `level=`, 0 `msg=`**, with the guidance on stderr only. Under `sg` present, stdout carries 38,004 lines of findings with zero slog contamination. No JSON/SARIF consumer can be corrupted.

Bonus verification not in the claim set — `MOAI_LOG_LEVEL` works end-to-end at the CLI, not merely in the unit test:

| Setting | Observed |
|---|---|
| unset | 4 WARN records on stderr |
| `error` | **0** WARN records; stderr = 148 bytes (only M2's own guidance) |
| `info` | 3 INFO records appear, including the discriminating `gopls bridge disabled (lsp.yaml: enabled=false)` |
| `bogus-not-a-level` | falls back to warn (4 WARN, 0 INFO); exit unaffected |
| hook path + `debug` | **0 stderr bytes** — carve-out is unconditional |

The `error` case also proves M2's guidance is a direct write, not a `slog` record (REQ-FAG-014) — it survives level suppression.

### 6. Newly visible warnings — see F7 (observation, root-caused, one follow-up recommended)

---

## Coverage

The implementer claimed deltas but I had no baseline, so I measured **both ends** in a detached worktree at `6763aff3b`.

| Package | Baseline | HEAD | Delta | vs 85% target |
|---|---:|---:|---:|---|
| `internal/astgrep` | 88.7% | **89.0%** | +0.3 | above |
| `internal/hook/quality` | 87.5% | **87.6%** | +0.1 | above |
| `internal/hook` | 83.3% | **83.4%** | +0.1 | below (pre-existing) |
| `internal/cli` | 75.3% | **75.4%** | +0.1 | below (pre-existing) |
| `internal/cli/uikit` | 98.8% | **98.8%** | 0.0 | above |

**Judgment on `internal/cli`**: M1/M2 did **not** worsen it — it rose 75.3 → 75.4. The implementer's "75.4 flat" slightly under-claims; the true delta is a marginal improvement. The sub-85 state is inherited debt across a large package, and this SPEC added `logging.go` (90 lines) with 176 lines of accompanying tests. Raising `internal/cli` to 85 is a separate, much larger effort and is correctly out of scope. The SPEC's own §F gate is relative ("not below the pre-change measurement") and is satisfied on every package.

The profile's absolute "Coverage below 85% = Craft FAIL" hard threshold is applied to the packages this SPEC owns and creates — `internal/astgrep` (89.0) and `internal/hook/quality` (87.6) — both of which clear it. Craft is scored PASS-with-debt at 82 rather than FAIL, and Craft is not a must-pass dimension under this profile.

---

## Security

No Critical or High findings.

| Surface | Finding |
|---|---|
| Sentinel path | `ValidateBinary` runs **before** `isSGAvailable` (`scanner.go:249` vs `:257`), so an untrusted path is rejected before any probe. Verified behaviorally by subtest 2. |
| `exec.LookPath` | Used in `isSGAvailable` (pre-existing) and the new `checkAstGrep`. Both take a fixed literal `"sg"`, not user input. |
| `trustedBinaryPrefixes()` | **Untouched** — diff of `scanner.go` contains no `[+-]` line matching `ValidateBinary`/`trustedBinaryPrefixes`/`ContainsAny`/`IsAbs`. Shell-metachar and `..` traversal defenses intact. |
| Error-message disclosure | Sentinel wraps `%q not found in PATH` with the binary *token*. The CLI's sentinel branch prints a **fixed** string and returns `exitCodeError` with a fixed msg — `err.Error()` never reaches user-facing stderr, so no path leaks there. The `binary=` slog attr under a stripped PATH prints only `sg`. |
| Availability (DoS-adjacent) | An absent optional tool must never block every commit. Verified at three frames plus Form C: `passed` is always `true` on the sentinel path. |
| `panic()` | **0** introduced (`git diff | grep '^\+' | grep -c 'panic('` → 0). |
| New disclosure surface | `MOAI_LOG_LEVEL=debug` now surfaces debug records on non-hook paths. Contained: default is `warn`, opt-in only, and this SPEC adds no callsites (§C explicitly excludes callsite content). Hook path remains unconditionally silent. |
| `doctor` check severity | `Warn`, never `Fail` — correctly avoids misreporting an optional tool as a broken install and avoids polluting the `--fix` list with an unfixable entry (REQ-FAG-027). Verified: sg absent → `status: "warn"`; sg present → `status: "ok"`. |

---

## Consistency

| Check | Result |
|---|---|
| Scope discipline | **Clean.** `git diff <merge-base>..HEAD --stat` = 32 files, all within the surfaces declared in `spec.md` §E. Zero drive-by changes. (A plain `git diff origin/main` is misleading here — origin advanced 8 commits, so it shows unrelated reversions.) |
| Golden files | Only the sg row + column widening + `4 ok`→`5 ok` + `Pass 11`→`Pass 12`. **No unrelated drift** across all three variants. Wired with `MOAI_SG_VERSION_OVERRIDE` in 8 test functions, matching the git/gh sibling pattern exactly. |
| Error wrapping | `fmt.Errorf(... %w ...)` throughout; no string concatenation. |
| Exit codes | 0/1 only; the two exit-1 causes distinguishable by stderr. `ast-edit` correctly retains exit 0 with unchanged source (anchored diff = 0 lines). |
| Comments | English; unusually good — they record *why* (e.g. `gate.go:283-285` names the exact discarded-reason defect). |
| `panic()` / drive-by refactor | None. |
| Env-var convention | See F8 — judged acceptable. |

---

## Acceptance Criteria — independent verdicts

| AC | Covers | My verdict | Note |
|----|--------|-----------|------|
| AC-FAG-001 | REQ-001 | **PASS** | `slog.SetDefault` non-test count 2→**1** (`logging.go:73`). Not a comment match. |
| AC-FAG-002 | REQ-004/005/006/007 | **PASS** | `TestResolveLogLevel` exists (1), **8** subtest PASS (≥5 required). `config.EnvLogLevel` referenced in 2 files; literal `"MOAI_LOG_LEVEL"` in non-test = **0**. Also verified end-to-end via the binary. |
| AC-FAG-003 | REQ-002/003 | **PASS** | `TestLoggingHandlerSelection` exists (1), **7** subtest PASS (≥4 required). |
| AC-FAG-004 | REQ-003/030 | **PASS** | (a) stderr 1050 bytes, `ast-grep (sg) CLI not found` ≥1. (b) M1-only build reproduces the pre-M2 shape; the pre-change binary at merge-base emits 0 stderr bytes. |
| AC-FAG-005 | REQ-008 | **PASS** | Guard holds (0/0) **and discriminates** — constructive falsification `dest: os.Stdout` flips it 0→4. Also 0/0 under 38,004 lines of real stdout. |
| AC-FAG-006 | REQ-009/030 | **PASS** | Half 1: exit 0, stdout 1 line, stderr **0 bytes**. Half 2: `TestTrivialPathSkipsFullInit` exists, 1 PASS. |
| AC-FAG-007 | REQ-010/011/012/013 | **PASS** | 3/3 subtests PASS, including subtest 3 **executed** (not skipped) — `sg` present. |
| AC-FAG-008 | REQ-014/015 | **PASS (1 of 3 sub-assertions vacuous)** | exit 0→**1** ✓, `no findings` 1→**0** ✓, `install` count **vacuous** (F3). Positive control: sg present → exit 1 from findings, `scan did not run` = 0. |
| AC-FAG-009 | REQ-016 | **PASS** | text/json/sarif → stdout install-mentions 0/0/0. |
| AC-FAG-010 | REQ-017 | **PASS** | exit **0**, `nothing to apply` ×1, anchored diff of `astedit.go` = 0 lines. Asymmetry preserved. |
| AC-FAG-011 | REQ-018/021/022/023/030 | **PASS** | Test exists, 1 PASS. **Form A and Form B both independently re-run by me → both FAIL.** Plus end-to-end `systemMessage` on the wire. |
| AC-FAG-012 | REQ-019/032 | **PASS** | `TestAstGrepGateReasons` exists, **2** subtest PASS. Both reasons are package-level constants; degraded branch additionally covered end-to-end by `TestRunAstGrepGateV2_ScanDegradedReason` via a fake `sg`. |
| AC-FAG-013 | REQ-020 | **PASS** | `-list` **2**, `--- PASS` **2**, `--- FAIL` **0**. |
| AC-FAG-014 | REQ-028/029 | **PASS** | (a) en/ja/ko/zh all 2 lines. (b) `guide/quick-start` in **4/4**. (c) stale claim **0**. **Human obligation discharged**: I read all four pages; each states the asymmetry (`ast-grep` non-zero / `ast-edit` zero) and carries install guidance. |
| AC-FAG-015 | REQ-024/026 | **PASS (cmd2 mislabeled — F4)** | JSON grep 0→**2**; exactly 1 matching check row. `exec.LookPath` count **3**. |
| AC-FAG-016 | REQ-025/027 | **PASS** | sg absent → `status: "warn"` + install URL. sg present → `status: "ok"`, `ast-grep 0.40.5`. |
| AC-FAG-017 | REQ-031 | **PASS on the suite half; skip half VACUOUS (F2)** | `go test ./...` exit **0**, **105 ok / 0 FAIL**. Skip half cannot discriminate — but I verified the requirement directly: skip count **56 = 56**, named-skip delta **empty both directions**. |
| AC-FAG-018 | standing constraints | **PASS** | vet exit 0; `GOOS=windows` exit 0; `GOOS=linux` exit 0; scoped `gofmt -l` **0**; and `gofmt -l` over **every** `.go` file this branch touches = **0**. |

**Summary**: 18/18 PASS. Two sub-assertions vacuous (F2, F3); one prose mislabel (F4). Both requirements behind the vacuous sub-assertions were independently verified to hold.

---

## Commands run and observed output

All from `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/faclear` @ `63c3df110`. Binary built to a scratchpad path. Every falsification ran in a `git worktree add --detach` probe, all removed afterward; final `git status --porcelain` = **0 lines**, HEAD and branch unchanged, no `git stash` used.

### Environment / controls
```
command -v sg                        → /opt/homebrew/bin/sg  (exit 0)
sg --version                         → ast-grep 0.40.5
PATH="$EMPTY" command -v sg          → (no output), exit 1
go build -o <scratch>/moai ./cmd/moai → exit 0
git merge-base origin/main HEAD      → 6763aff3b8d54d926a9fd0d12df7cc9395df6cfc
git diff <merge-base>..HEAD --stat   → 32 files, 1845 insertions, 109 deletions
```

### AC-FAG-008 / 009 / 005 — CLI behavior, sg absent
```
env PATH=$EMPTY moai ast-grep ./internal/astgrep   → exit=1
  stdout 0 bytes; stderr 1050 bytes
  stderr tail: "ast-grep: scan did not run — the ast-grep (sg) CLI was not found.
                Install it from https://ast-grep.github.io/guide/quick-start.html, then re-run."
  grep -ci install (stderr) = 2 ; grep -c 'no findings' (stdout) = 0
--format=text|json|sarif → stdout_bytes 0/0/0, install-mentions 0/0/0, level= 0/0/0, msg= 0/0/0
```

### Positive control — sg present
```
moai ast-grep ./internal   → exit=1  (19006 findings, 38004 stdout lines)
  grep -c 'scan did not run' stderr = 0    ← two exit-1 causes distinguishable
  grep -c 'level=' stdout = 0 ; grep -c 'msg=' stdout = 0
```

### AC-FAG-006 / 010
```
moai --version   → exit=0, stdout 1 line ("moai-adk v3.0.0"), stderr 0 bytes
env PATH=$EMPTY moai ast-edit --pattern x --rewrite y --lang go ./internal/astgrep
  → exit=0, stdout "ast-grep (sg) is not installed; nothing to apply."
  git diff --stat 6763aff3b -- internal/cli/astedit.go | wc -l → 0
```

### AC-FAG-015 / 016 — doctor
```
sg absent : {"name":"ast-grep CLI","status":"warn","message":"sg not found — 'moai ast-grep' cannot scan
             and the commit gate skips its rules; install from https://ast-grep.github.io/guide/quick-start.html"}
sg present: {"name":"ast-grep CLI","status":"ok","message":"ast-grep 0.40.5"}
grep -c 'exec.LookPath' internal/cli/doctor.go → 3
```

### Go-test ACs (Shape 1 — existence + actual PASS lines)
```
TestResolveLogLevel                    list=1, subtest PASS=8
TestLoggingHandlerSelection            list=1, subtest PASS=7
TestTrivialPathSkipsFullInit           list=1, PASS=1
TestScannerScan_UnavailableSentinel    list=1, subtest PASS=3  (unavailable / non-sentinel / clean_scan)
TestPreTool_AstGrepSkipReasonSurfaces  list=1, PASS=1
TestAstGrepGateReasons                 list=1, subtest PASS=2
TestQualityGate_Run_AstGrep*           list=2, PASS=2, FAIL=0
grep -rn 'slog.SetDefault' internal/ cmd/ --include='*.go' | grep -v _test → 1 (logging.go:73)
grep -rln 'config.EnvLogLevel' internal/cli/ → 2 files
grep -rn '"MOAI_LOG_LEVEL"' internal/cli/ | grep -v _test → 0
```

### Falsifications (detached probe worktrees at HEAD)
```
FORM B — restore gate.go discarding pass branch + return true, "":
  --- FAIL: TestPreTool_AstGrepSkipReasonSurfaces
      pre_tool_astgrep_reason_test.go:78: SystemMessage is empty: the ast-grep skip
      reason was dropped somewhere in the three-frame chain
  full ./internal/hook/... → 1 failing test; 9 sibling packages still ok

FORM A — delete `out.SystemMessage = gateNotice` in pre_tool.go:
  --- FAIL: TestPreTool_AstGrepSkipReasonSurfaces  (same assertion)

FORM C (auditor-added) — revert astgrep_gate.go sentinel branch to (true, ""):
  --- FAIL: TestAstGrepGateReasons
  --- FAIL: TestPreTool_AstGrepSkipReasonSurfaces
  --- FAIL: TestRunAstGrepGateV2_NoSgCLI
  --- FAIL: TestRunAstGrepGateV2_ProjectDirPathVariants
  --- FAIL: TestRunAstGrepGateV2_ScanDegradedReason
  --- FAIL: TestRunAstGrepGateV2_TableDriven

AC-FAG-005 constructive — logging.go dest: os.Stderr → os.Stdout:
  grep -c 'level=' stdout = 4  (was 0)   → criterion DOES discriminate
  --- FAIL: TestLoggingHandlerSelection/{update,bare_invocation_writes_stderr}

AC-FAG-015 registration — delete {"ast-grep CLI", checkAstGrep} from systemChecks:
  cmd1 JSON grep : 1 → 0   ← the real registration guard
  cmd2 LookPath  : 3 → 3   ← NOT a registration guard (F4)
  --- FAIL: TestDoctorGolden_{Light,Dark,NoColor}

M1-ONLY build (a56bfb58c) — AC-FAG-008 vacuity:
  exit=0 ; stdout "no findings" ; grep -ci install stderr = 1   ← passes with no M2 change (F3)
```

### End-to-end on the wire
```
echo '{"session_id":"t","hook_event_name":"PreToolUse","tool_name":"Bash",
       "tool_input":{"command":"git commit -m \"x\""}}' \
 | env PATH=$EMPTY CLAUDE_PROJECT_DIR=$PD moai hook pre-tool
→ exit=0, stderr 0 bytes, stdout:
{"systemMessage":"ast-grep scan skipped: the sg CLI was not found, so no rules ran
  (install from https://ast-grep.github.io/guide/quick-start.html)",
 "hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}
```

### MOAI_LOG_LEVEL end-to-end + hook carve-out
```
(unset)              → 4 WARN on stderr
MOAI_LOG_LEVEL=error → 0 WARN; stderr 148 bytes (M2 guidance only — survives suppression)
MOAI_LOG_LEVEL=info  → 3 INFO, incl. "gopls bridge disabled (lsp.yaml: enabled=false)"
MOAI_LOG_LEVEL=bogus-not-a-level → 4 WARN, 0 INFO, exit unaffected
moai hook pre-tool                          → stderr 0 bytes, valid JSON stdout
moai hook pre-tool (MOAI_LOG_LEVEL=debug)   → stderr 0 bytes   ← carve-out unconditional
grep -n 'enabled' .moai/config/sections/lsp.yaml → line 43: enabled: false
```

### AC-FAG-014 — docs
```
per-locale anchored diff: en→2  ja→2  ko→2  zh→2
grep -l 'guide/quick-start' docs-site/content/*/cli-reference/ast-grep.md | wc -l → 4
grep -c 'exits without an error' docs-site/content/en/cli-reference/ast-grep.md → 0
(all four pages read in full — each states ast-grep non-zero / ast-edit zero)
```

### AC-FAG-017 / 018 + coverage
```
go test -count=1 ./...                        → exit=0, 105 ok, 0 FAIL
go vet ./...                                  → exit=0
GOOS=windows GOARCH=amd64 go build ./...      → exit=0
GOOS=linux   GOARCH=amd64 go build ./...      → exit=0
gofmt -l <SPEC file list + new test files>    → 0
git diff --name-only <mb>..HEAD -- '*.go' | xargs gofmt -l | wc -l → 0

skip analysis:
  grep -c '^--- SKIP' (non-verbose log) → 0    ← structurally pinned; cannot discriminate (F2)
  grep -c 'SKIP'      (non-verbose log) → 0
  verbose '--- SKIP'  baseline=56  HEAD=56
  comm -13 base.skips head.skips → (empty)     ← no new skip
  comm -23 base.skips head.skips → (empty)     ← none removed

coverage (baseline @6763aff3b → HEAD):
  internal/astgrep      88.7 → 89.0
  internal/hook/quality 87.5 → 87.6
  internal/hook         83.3 → 83.4
  internal/cli          75.3 → 75.4
  internal/cli/uikit    98.8 → 98.8
```

### Golden files / scope
```
git diff <mb>..HEAD -- internal/cli/testdata/ → only the ast-grep row, column widening,
  "4 ok"→"5 ok", "Pass 11"→"Pass 12" across dark/light/nocolor. No unrelated drift.
git diff <mb>..HEAD -- '*.go' | grep '^+' | grep -c 'panic('  → 0
diff of internal/astgrep/scanner.go touching ValidateBinary/trustedBinaryPrefixes → none
grep -rn 'PersistentFlags' internal/cli/root.go → no match  (F5 refuted as reachable)
uikit imports: printer, settings, tui   |   config imports: defs, pkg/models  (no cycle — F8)
```

### internal/web flake (not attributable)
```
go test -v ./... → FAIL github.com/modu-ai/moai-adk/internal/web  (package-level; no --- FAIL test line)
git diff --name-only <mb>..HEAD -- internal/web/ | wc -l → 0
standalone re-runs: exit 0, 0, 0 ; verbose re-run: ok 0.720s
→ pre-existing known flake under whole-suite verbose load; 0 files touched by this branch
```

---

## Gaps — what I did NOT verify

1. **Windows / Linux runtime behavior.** Cross-*compilation* passes for both, but I executed nothing on either. All shell criteria are POSIX-only (the SPEC's own G-8). In particular `isHookCommand`, the `%q`-quoted binary in the sentinel message, and `exec.LookPath("sg")` (which resolves `sg.exe` on Windows) are unexercised at runtime.
2. **`sg`-absent host.** Every runtime check used a PATH strip on a host where `sg` IS installed. I did not test on a machine genuinely without `sg`, so I have not observed the two new conditional skips actually firing, nor CI coverage under that condition (F9).
3. **Claude Code's rendering of `systemMessage` on a PreToolUse allow.** I verified the field is emitted in the hook's stdout JSON. Whether the Claude Code runtime *surfaces it to the user* on the allow path is a runtime-contract question I did not test — no live Claude Code session was driven. If the runtime silently drops `systemMessage` for PreToolUse allow decisions, REQ-FAG-022's user-visible intent would be unmet despite every test passing. **This is the largest residual risk in the SPEC.**
4. **`golangci-lint`.** Not run (not part of any AC). Only `go vet` was executed.
5. **`gofumpt`.** The repo's actual formatter per `Makefile:61`; not installed, per the SPEC's own §C note. The 106 pre-existing `gofmt`-unclean files were not re-checked (correctly out of scope).
6. **Degraded-reason branch in production.** `astGrepReasonScanDegraded` is covered by a test using a fake `sg` script, but I did not observe it through the real CLI/gate with a genuinely malfunctioning `sg`.
7. **Concurrency / race.** `go test -race` not run. `configureLogging` calls `slog.SetDefault` once from `Execute()` before any goroutine, so I judge the risk low by code reading, but did not verify.
8. **Docs-site build.** The 4-locale markdown was read but Hugo was not built; rendering and link validity unverified.
9. **F5 / F6** are code-reading inferences. F5's unreachability IS confirmed by execution; its future-fragility claim is not (it describes a hypothetical flag that does not exist).
10. **Plan-audit history** at `.moai/reports/plan-audit/` was not read — I audited the implementation against the shipped contract, deliberately without the prior auditors' framing.

---

## Recommendations

1. **[Required before close]** Populate `progress.md` §E.2 / §E.3 (F1). The Commands section above can be cited directly.
2. **[Required before close]** Fix AC-FAG-017's skip command to the `-v` form and record the baseline `56` (F2). Note host-dependence.
3. **[Recommended]** Re-anchor AC-FAG-008's third sub-assertion to `grep -c 'scan did not run'` (F3); fix AC-FAG-015's guard-attribution sentence (F4).
4. **[Recommended]** Close Gap 3 — drive one real Claude Code `git commit` on a machine without `sg` and confirm the skip notice is user-visible. This is the only way to know REQ-FAG-022 achieved its intent rather than merely its mechanism.
5. **[Follow-up SPEC]** Down-level `deps.go:179` to Info when the gopls bridge is *configured* off, matching the sibling at `deps.go:131`; keep Warn for genuine init failure at `deps.go:125` (F7).
6. **[Follow-up SPEC]** `MOAI_LOG_FORMAT` is now a *visibly misleading* knob, as `spec.md` §C candidly records. This SPEC converted it from invisibly-inert to actively wrong. Worth scheduling rather than leaving in the exclusions list.
7. **[Optional]** Add a second guard for the three-frame propagation, or a comment on `TestPreTool_AstGrepSkipReasonSurfaces` warning that it is the sole protection (Forms A and B each kill only that one test).

---

## Assessment

The implementation is correct, well-tested, and I could not break it — Forms A, B, and my own Form C all fail as they should, `errors.Is` discriminates in both directions, the reason reaches Claude Code's actual wire format, and no rewritten test was weakened (one was materially strengthened). Scope is clean, coverage improved on all five touched packages, and every cross-platform and format gate is green.

The findings are concentrated in the acceptance criteria rather than the code. Two sub-assertions cannot discriminate, and in a SPEC whose subject is false all-clears that is worth naming plainly — a passing vacuous criterion is the same defect class the SPEC exists to eliminate. Both underlying requirements nonetheless hold; I verified them directly rather than trusting the criteria. Combined with the empty run-phase evidence surface, the pattern is a strong implementation carrying a weaker verification record. Closing F1 and F2 converts this to an unqualified PASS.

**Report path**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/faclear/.moai/reports/sync-audit/SPEC-FALSE-ALLCLEAR-GUARD-001-2026-07-27.md`
