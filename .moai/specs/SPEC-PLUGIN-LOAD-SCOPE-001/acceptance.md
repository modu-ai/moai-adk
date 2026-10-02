# Acceptance — SPEC-PLUGIN-LOAD-SCOPE-001

All commands are plain, separately invocable, and run from the card worktree root. The exit code is its
own field. A path written `evidence/...`, `verdict.md`, `check-evidence.sh` or `check-verdict.sh` without a
directory prefix is relative to `.moai/reports/t1434/`. Every checker prints its counters, exits non-zero on
any failure, and exits non-zero when its swept set is empty (`verification-completeness.md` §1.1, empty
sweep). Run-phase artifacts under `.moai/reports/t1434/` are **local-only** (`.gitignore:235`); the durable
carrier the follow-up cards read is the `progress.md §E.2` block checked by AC-015.

**Which checker modes carry mutants.** `check-verdict.sh` decides the deliverable: every counter of its seven
modes carries a mutant (AC-009; the names are listed in `plan.md` §6, the single place that defines them).
`check-evidence.sh` is a **measurement helper**: its modes carry no mutants of their own, and a wrong reading
there is covered by the cell-level cross-checks of `check-verdict.sh` (AC-008) and by the red input each
criterion states under "Turns red when". The consequence is stated, not hidden: the `check-evidence.sh` modes
are not observed failing, and the verdict's Residual-risk says so.

## §D AC Matrix

| AC | Requirement | Description | Severity | Verification |
|----|-------------|-------------|----------|--------------|
| AC-001 | REQ-001 | Repository writes confined to the named run-commit paths; no product path touched, committed or uncommitted | Should (regression guard) | Direct |
| AC-002 | REQ-002 | Environment recorded first; scrub list enumerated from the live environment and re-enumerated by the checker; every command logged scrubbed | Must | Direct |
| AC-003 | REQ-003 | Static commands isolated: scratch homes go from empty to non-empty, no credentials in them | Must | Direct |
| AC-004 | REQ-004 | Real-home commands are observe-only; verb lists live in the checker; no registry-writing verb touches a real home | Must | Direct |
| AC-005 | REQ-005 | Protected-set manifests per real-home command, attributed AMBIENT or LEAK, LEAK handled by the stop rule, builder shown able to fail | Must | Direct |
| AC-006 | REQ-006, REQ-007 | Claude static evidence per fixture and per command kind, fixture count equal to the row count, executed-test count | Must | Direct |
| AC-007 | REQ-008 | Runtime cells carry sentinel, project control, and a captured control value; zero sweep fails; run cap pinned | Must | Direct |
| AC-008 | REQ-009, REQ-014 | Every cell has a five-field card whose quoted deciding line occurs in the cited raw file and names a row token | Must | Direct |
| AC-009 | REQ-009, REQ-014 | `check-verdict.sh` shown able to fail: a base that passes and one mutant per counter and per mode, re-executed | Must | Direct |
| AC-010 | REQ-010 | Settings-hook equivalence fully observed: both executions, both expansions from `ARG0`, cutoff seconds | Must | Direct |
| AC-011 | REQ-011 | Rules and instruction rows carry plugin and project sentinels, three runs, the R11 configuration scope | Must | Direct |
| AC-012 | REQ-012 | Mods row follows the fixed result table, with a nonzero executed-test count | Must | Direct |
| AC-013 | REQ-013 | Codex measured under a scratch home; only successful adds counted; no real-home Codex write | Must | Direct |
| AC-014 | REQ-015, REQ-016 | Verdict table: 14 rows, vocabulary, home consistency, consequence, derived yield floor tied to row tokens | Must | Direct |
| AC-015 | REQ-014, REQ-016 | Versions equal at start and end, hash recomputed, non-goal line, durable carrier in progress.md §E.2 | Must | Direct |

Budget check (Tier M ceilings 16/16, counted independently): 16 requirements, 15 acceptance criteria.

## RED-now Ledger (keyed per criterion)

The deliverables do not exist at the pinned tree, so every criterion that reads a deliverable is red for a
stated reason: the work has not run. Each cell was executed in iteration 3 on this worktree, whose committed
revision is `6d0d75af3`, with the exit code read through an `; echo EXIT=$?` wrapper (the tool shows nothing
for an exit-1 command with empty output). **Pin caveat:** the files these cells read live under
`.moai/reports/t1434/` (local-only, never in a commit), so the SHA names the committed tracked tree and the
cell records the worktree's state at measurement time, not a recoverable commit. The `test` commands print
nothing; the exit code is the field. RN-001 is the one cell whose exit code the wrapper could not read: the
worktree guard refuses a git command followed by `; echo`, so the plain command was run and the tool reported
no error on a command that printed four paths (recorded as 0, not separately displayed). The iteration-1
ledger row L-1 (`test -d .moai/reports/t1434`) stays retired: the plan-auditor created that directory.

| Key | Command | Verbatim stdout | Exit code | Tree SHA | Why red |
|-----|---------|-----------------|-----------|----------|---------|
| RN-001 | `git diff --name-only 802a72235..6d0d75af3 -- .moai/specs` | the four paths `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/acceptance.md`, `plan.md`, `progress.md`, `spec.md` | 0 | 6d0d75af3 | AC-001 is a regression guard whose product-path forms print nothing at arrival; as a whole it reads unmeasured (not green) until the run commit exists, because its third form must be non-empty. This is its mutant-input cell: the same git instrument over a pathspec that includes touched paths prints a non-empty list, so the instrument is not blind |
| RN-002 | `test -s .moai/reports/t1434/evidence/commands.log` | (empty) | 1 | 6d0d75af3 | the command log does not exist; written from M1 |
| RN-003 | `test -s .moai/reports/t1434/evidence/scratch-home-after.txt` | (empty) | 1 | 6d0d75af3 | no scratch-home listing exists; written at M1 |
| RN-004 | `grep -c "home=real" .moai/reports/t1434/evidence/commands.log` | (empty; stderr `ugrep: warning: .moai/reports/t1434/evidence/commands.log: No such file or directory`) | 2 | 6d0d75af3 | the command log does not exist, so no real-home line can be audited; the pattern `home=real` matches the specified `CMD: id=<n> home=real ...` line format |
| RN-005 | `test -s .moai/reports/t1434/evidence/negative-control-manifest.txt` | (empty) | 1 | 6d0d75af3 | the manifest-builder control has not run; written at M1 |
| RN-006 | `test -d .moai/reports/t1434/evidence/raw` | (empty) | 1 | 6d0d75af3 | no raw evidence directory; created at M1 |
| RN-007 | `test -f .moai/reports/t1434/check-verdict.sh` | (empty) | 1 | 6d0d75af3 | the verdict checker is not written; written at M1, finished at M5 |
| RN-008 | `test -f .moai/reports/t1434/verdict.md` | (empty) | 1 | 6d0d75af3 | the verdict is not written; written at M5 |
| RN-009 | `test -s .moai/reports/t1434/evidence/negative-control-checker.txt` | (empty) | 1 | 6d0d75af3 | the checker mutants have not run; written at M5 |
| RN-010 | `test -s .moai/reports/t1434/evidence/raw/claude-runtime-R08.txt` | (empty) | 1 | 6d0d75af3 | the R08 runtime session has not run; written at M3 |
| RN-011 | `test -s .moai/reports/t1434/evidence/raw/claude-runtime-R09.txt` | (empty) | 1 | 6d0d75af3 | the R09 runtime session has not run; written at M3 |
| RN-012 | `test -s .moai/reports/t1434/evidence/raw/claude-runtime-R04.txt` | (empty) | 1 | 6d0d75af3 | the R04 runtime session has not run; written at M3 |
| RN-013 | `test -s .moai/reports/t1434/evidence/raw/codex-plugin-list.txt` | (empty) | 1 | 6d0d75af3 | no Codex list output exists; written at M4 |
| RN-014 | `grep -c "^| ID | Component" .moai/reports/t1434/verdict.md` | (empty; stderr `ugrep: warning: .moai/reports/t1434/verdict.md: No such file or directory`) | 2 | 6d0d75af3 | the verdict table does not exist; written at M5 |
| RN-015 | `grep -c "evidence index" .moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md` | `0` | 1 | 6d0d75af3 | `progress.md §E.2` is a placeholder; the durable carrier is written at M5 |

## Green-path Ledger (keyed per criterion)

Each green path names the milestone that flips the criterion and what the passing output becomes. A green
path that would need an unrelated change to flip is disqualified (`verification-completeness.md` §2).

| Key | Flips at | Passing output |
|-----|----------|----------------|
| GP-001 | M5 (the run commit), after the orchestrator replaces `run_start_sha` with the SHA of the plan-final commit | the product-path forms print nothing, `git status --porcelain` lists only the two allowed paths (or nothing), `git diff --name-only BASE` lists only those paths and is non-empty, the `spec.md` diff holds only `status:`/`updated:` lines, and `git status --porcelain --ignored -- .moai/reports/t1434` prints `!! .moai/reports/t1434/` |
| GP-002 | M1 | `check-evidence.sh commands` prints `CMDS=<n> SCRUB-MATCH=<n> ENV-NAMES-MATCH=1 CMD-COUNT-MATCH=1 SECRET-VALUES=0 RESULT=PASS`, exit 0, `<n>` equal to the probe's `CMD-COUNT` |
| GP-003 | M1 (static phase at M2) | `check-evidence.sh isolation` prints `TOOLS=2 ISOLATED=<a> FAILED=<b> ... CRED-FILES=0 RESULT=PASS` with `a+b=2`, exit 0 |
| GP-004 | M3 | `check-evidence.sh observe-only` prints `REAL-CMDS=<n> FORBIDDEN-REAL=0 REAL-OFF-ALLOWLIST=0 SCRATCH-NO-OVERRIDE=0 REAL-WITH-OVERRIDE=0 ... RESULT=PASS`, exit 0 |
| GP-005 | M1 (builder control), M3 (pairs) | `check-evidence.sh manifests` prints `PAIRS=<n+1> REAL-CMDS=<n> LEAK=0 LEAK-FORBIDDEN=0 LEAK-UNREPORTED=0 LEAK-CONTINUED=0 BUILDER-CONTROL=PASS RESULT=PASS`, exit 0 |
| GP-006 | M2 | `check-evidence.sh static` prints `ROWS=14 FIXTURES=14 VALIDATE-PLAIN=14 VALIDATE-STRICT=14 INSTALL-FILES=14 ... TESTS-MATCH=1 RESULT=PASS`, exit 0 |
| GP-007 | M3 (finished M5) | `check-verdict.sh runtime verdict.md` prints `RUNTIME-ROWS=10 SWEPT=<n>` with `n >= 1`, `BAD-CONTROL=0 BAD-RUNS=0 RUNTIME-RUN-CAP=60 OVER-CAP=0 RESULT=PASS`, exit 0 |
| GP-008 | M5 | `check-verdict.sh cells verdict.md` prints `CELLS=28 CARDLESS=0 MISSING-FIELDS=0 ORPHAN-RAW=0 QUOTE-MISMATCH=0 EMPTY-REASON=0 RESULT=PASS`, exit 0 |
| GP-009 | M5 | `check-evidence.sh negative-controls` prints `BASES-GREEN=7 MUTANTS=40 MUTANTS-RED=40 RESULT=PASS`, exit 0 |
| GP-010 | M3 | `check-verdict.sh r08 evidence/raw/claude-runtime-R08.txt` prints `EXEC-BAD=0 EXPANSION-BAD=0 EPOCH-BAD=0 ENTRY-DIFF-BAD=0` with the recomputed `EXPANSION`, `TIMEOUT-CUTOFF` and `SECONDS` and `RESULT=PASS`, exit 0; or the R08 Claude cell is `UNOBSERVED` with a quoted proof |
| GP-011 | M3 | `check-verdict.sh rules verdict.md` prints `RUNS-BAD=0 CONTROL-BAD=0 CONFIG-SCOPE-BAD=0 VALIDATE-MSG-BAD=0 RESULT=PASS`, exit 0 |
| GP-012 | M2 (test count), M3 (marker) | `check-verdict.sh r04 verdict.md` prints `TESTS-EXECUTED=<n> TESTS-FOUND=<n> OUTCOME=<allowed> TESTS-GAP=0 PAYLOAD-BAD=0 CELL-MISMATCH=0 RESULT=PASS`, exit 0 |
| GP-013 | M4 | `check-verdict.sh codex verdict.md` prints `ADDS-OK=<a> LISTED=<a> ADD-MISMATCH=0 VALIDATE-LINES=0 REAL-CODEX-WRITES=0 RESULT=PASS`, exit 0 |
| GP-014 | M5 | `check-verdict.sh table verdict.md` prints `ROWS=14 HEADER=1 ... FLOOR=7 FLOOR-MET=7 RESULT=PASS`, exit 0 |
| GP-015 | M5 | `check-evidence.sh stamps` prints `STAMPED=<n> RAW=<n> VERSION-EQUAL=1 SHA-RECOMPUTED=1 NONGOAL=1 CARRIER-ROWS=14 CARRIER-SHA=1 RESULT=PASS`, exit 0 |

## §D.1 AC-001 — Repository writes confined to the named run-commit paths (regression guard, RN-001 / GP-001)

**Given** the card branch after the run phase, and BASE = the `run_start_sha:` value recorded in `progress.md §E.1`, read with `grep run_start_sha .moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md`. The value written at this revision is `6d0d75af3`, the audited commit; the orchestrator replaces it with the SHA of the commit that carries the final plan-phase revision, so a BASE older than the last change to `spec.md`, `plan.md` or `acceptance.md` makes the criterion red (the third form below lists those files).
**When** these five forms run, each as its own invocation: `git diff --name-only BASE -- internal cmd pkg .claude .mcp.json CLAUDE.md AGENTS.md` (BASE against the working tree, so committed and uncommitted tracked changes are both seen), `git status --porcelain` (unrestricted, so staged and untracked paths are seen), `git diff --name-only BASE`, `git diff -U0 BASE -- .moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/spec.md`, and `git status --porcelain --ignored -- .moai/reports/t1434`
**Then** the first prints nothing; the second prints nothing or only lines whose path is `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md` or `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/spec.md`; the third is non-empty and every line is one of those two paths (the run commit writes `progress.md §E.2` and `§E.3` and the `spec.md` status flip; `.moai/reports/t1434/` is ignored and never appears); in the fourth every changed content line (a line starting `+` or `-` other than the `+++` and `---` headers) begins `status:` or `updated:` after the sign; the fifth prints `!! .moai/reports/t1434/`, which shows the report directory is local-only. An empty third form means the measurement is reported as unmeasured, not as a pass.
**Turns red when** any product path is modified, committed or not (RN-001 shows the instrument fires on a touched path); a path other than the two allowed ones is changed in any way, including `plan.md` or `acceptance.md` (a criterion weakened and committed during the run) or a settings file such as `.moai/config/sections/workflow.yaml` or `Makefile` edited and left uncommitted (the second form lists it); a body line of `spec.md` changes (the fourth form).

## §D.2 AC-002 — Environment recorded, scrub enumerated live and re-enumerated, every command logged (RN-002 / GP-002)

**Given** `evidence/env.txt`, `evidence/env-scrub.txt`, `evidence/commands.log`, `evidence/probe-counts.txt`, and the environment of the checker's own process
**When** `sh .moai/reports/t1434/check-evidence.sh commands` runs
**Then** it prints `CMDS`, `SCRUB-MATCH`, `ENV-NAMES-MATCH`, `CMD-COUNT-MATCH`, `SECRET-VALUES`, `RESULT=PASS` and exits 0. `env.txt` carries `CLAUDE_CONFIG_DIR=`, `CODEX_HOME=`, an `AUTH-STATUS` section, and the live `CLAUDE_CODE_*`/`MOAI_*` names, and is its first written section; `ENV-NAMES-MATCH=1` means two things hold: the name set in `env-scrub.txt` equals the live-name set in `env.txt` in both directions (a name the probe kept is listed with its reason), and every `CLAUDE_CODE_*` and `MOAI_*` name the checker itself enumerates from its own environment at check time (names only) appears in `env-scrub.txt`, so the probe's own enumeration is not the only source; `SCRUB-MATCH` counts the `CMD:` lines that carry `id=`, `home=scratch` or `home=real`, and an unset list byte-equal to `env-scrub.txt`; `CMDS` equals the `CMD-COUNT` the probe printed from its own counter and is at least 1; no value for a name matching TOKEN, KEY or SECRET appears in `env.txt` (`SECRET-VALUES=0`).
**Turns red when** a `CMD:` line omits the unset prefix, a live name is missing from `env-scrub.txt` (mutant: the probe enumerates `CLAUDE_CODE_*` with a glob that skips `CLAUDE_CODE_MESSAGING_TOKEN`, so `env.txt` and `env-scrub.txt` agree and the checker's own enumeration still finds the name), a command ran without a log line (counter mismatch), the log is empty (zero sweep), or a secret value was written.

## §D.3 AC-003 — Static commands isolated by observation (RN-003 / GP-003)

**Given** `evidence/scratch-home-before.txt` and `evidence/scratch-home-after.txt`, one section per tool (claude, codex)
**When** `sh .moai/reports/t1434/check-evidence.sh isolation` runs
**Then** it prints `TOOLS=2 ISOLATED=<a> FAILED=<b> BEFORE-EMPTY BEFORE-NONEMPTY AFTER-NONEMPTY CRED-FILES RESULT=PASS` and exits 0. For every isolated tool the before-listing has zero entries and the after-listing, taken after the tool's first registry-writing command (the M1 trivial marketplace add), has at least one, including the tool's registry file (`plugins/installed_plugins.json` for Claude); `CRED-FILES=0` means no listed name matches `credential|auth|token|keychain|login`; a tool that is not isolated carries `ISOLATION-FAILED <tool> raw=<file>` with a quoted line, and the checker then requires zero registry-writing `CMD:` lines for that tool in `commands.log` after that verdict (the route actually taken, not an assumed one); `a+b` equals `TOOLS`.
**Turns red when** a before-listing is non-empty, an after-listing is empty with no `ISOLATION-FAILED` proof, a credential-named file sits in a scratch home, or a failed-isolation tool has a registry-writing command logged after the verdict.

## §D.4 AC-004 — Real-home commands are observe-only (RN-004 / GP-004)

**Given** `evidence/commands.log`, and the forbidden-verb and read-only-verb lists, which are literal text inside `check-evidence.sh` (mode `observe-only`), copied from REQ-004 in `spec.md`; the probe writes no verb list
**When** `sh .moai/reports/t1434/check-evidence.sh observe-only` runs
**Then** it prints `REAL-CMDS=<n> FORBIDDEN-REAL=0 REAL-OFF-ALLOWLIST=0 SCRATCH-NO-OVERRIDE=0 REAL-WITH-OVERRIDE=0 RESULT=PASS`, followed by the lines `FORBIDDEN-VERBS=<list>` and `READONLY-VERBS=<list>` (the checker's own lists, printed so a reviewer compares them with REQ-004), and exits 0: no `home=real` line contains a verb from the forbidden list (install, uninstall, marketplace add/remove/update, enable, disable, configure, plugin update, Codex plugin add/remove, `codex exec`); every `home=real` line is either a `claude -p` session carrying `--no-session-persistence` (the plugin run with `--plugin-dir`, or the control run in the scratch project without it, both allowed) or a verb from the read-only list (`--version`, `auth status`, `login status`, `plugin list`); every `home=scratch` line carries the `CLAUDE_CONFIG_DIR=<scratch>/config-claude` or `CODEX_HOME=<scratch>/config-codex` override for its tool, and every `home=real` line carries neither.
**Turns red when** a real-home line carries a forbidden verb (mutant: `home=real ... plugin install`, or `home=real ... codex exec`), a `home=real` `claude -p` line lacks `--no-session-persistence`, a `home=scratch` line lacks its override (mutant: the same command with the override stripped, which would write the real home), or the log is empty. A probe that writes an empty verb file changes nothing, because the checker does not read one.

## §D.5 AC-005 — Protected-set manifests, attribution, LEAK stop rule, builder control (RN-005 / GP-005)

**Given** `evidence/protected-t0.sha256`, `evidence/manifest/<id>-before.sha256` and `-after.sha256`, `evidence/manifest-diff-<id>.txt`, `evidence/protected-excluded.txt`, `evidence/negative-control-manifest.txt`, `evidence/commands.log`, and `verdict.md`
**When** `sh .moai/reports/t1434/check-evidence.sh manifests` runs
**Then** it prints `ROWS REQUIRED-ROWS PAIRS REAL-CMDS LEAK LEAK-FORBIDDEN LEAK-UNREPORTED LEAK-CONTINUED EXCLUDED-DECLARED BUILDER-CONTROL RESULT=PASS` and exits 0. `ROWS` equals both the line count of `protected-t0.sha256` and the `ROWS-MANIFEST` the builder printed in `probe-counts.txt` (an equality, not a threshold); the rows `claude:installed_plugins.json`, `claude:known_marketplaces.json`, `claude:settings.json` (of the resolved profile) and `codex:config.toml#plugins-tables` are all present; `PAIRS` equals the number of `home=real` `CMD:` lines plus one (the static-phase window); every row that differs in a `manifest-diff` is classified `AMBIENT` or `LEAK` by whether its changed content names a fixture plugin, the scratch marketplace, the scratch path, or a `SENTINEL_` string; `LEAK` counts the `LEAK` rows, and the stop rule holds: `LEAK-FORBIDDEN=0` (no `LEAK` row is attributed, through its before/after pair, to a command whose verb is forbidden; a forbidden-verb `LEAK` fails the criterion), `LEAK-UNREPORTED=0` (every `LEAK` row, which can only come from a permitted verb, is named by a `LEAK-FINDING: <cmd-id>` line in the Gaps of `verdict.md`), and `LEAK-CONTINUED=0` (no `home=real` command is logged after the command that produced the first `LEAK`); `protected-excluded.txt` names `.claude.json`, `sessions/`, `session-env/`, `debug/`, `file-history/`, `.codex-global-state.json` and Codex sqlite state, each with a reason. The builder control runs the **builder** on `$SCRATCH/ctl-tree`: `negative-control-manifest.txt` carries `CONTROL unchanged ROWS-DIFF=0`, `CONTROL changed-bytes ROWS-DIFF=1` (one file's bytes changed, same path and length, which a path-hashing builder cannot see), and `CONTROL planted-identifier LEAK=1`, and `BUILDER-CONTROL=PASS` requires all three.
**Turns red when** a required row is missing, `PAIRS` differs from the real-home command count, a `LEAK` is attributed to a forbidden verb, a `LEAK` has no `LEAK-FINDING` line, the probe kept running real-home commands after a `LEAK`, an excluded entry has no reason, or the builder hashes paths instead of content (the changed-bytes control then reads `ROWS-DIFF=0`). A changed row that names no fixture identifier is `AMBIENT` and turns nothing red.

## §D.6 AC-006 — Claude static evidence per fixture and command kind (RN-006 / GP-006)

**Given** the raw files `claude-validate-<ID>`, `claude-validate-strict-<ID>`, `claude-install-<ID>`, `claude-details-<ID>`, `claude-test-<ID>` after M2, `evidence/fixture-manifest.txt`, and §5 of `spec.md`
**When** `sh .moai/reports/t1434/check-evidence.sh static` runs
**Then** it prints `ROWS=<r> FIXTURES=<f> VALIDATE-PLAIN VALIDATE-STRICT INSTALL-FILES INSTALL-OK DETAILS TEST-FIXTURES TESTS-MATCH RESULT=PASS` and exits 0: `ROWS` is the count of `^| R[0-9][0-9] ` rows of §5 in `spec.md` (14) and `FIXTURES` is the count of fixture plugin directories in `fixture-manifest.txt` that carry both `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` (14), the two equal (one single-component fixture per row, no composite); `VALIDATE-PLAIN`, `VALIDATE-STRICT` and `INSTALL-FILES` each equal `FIXTURES` (every file ends with an `EXIT=` trailer; a failing install is an observation, not a fault); `INSTALL-OK` is the count of installs whose trailer is `EXIT=0` and whose output contains the verbatim `Successfully installed plugin` line; `DETAILS` equals `INSTALL-OK` and each details file contains an `Always-on` line; `TEST-FIXTURES` equals the number of fixtures holding `*.test.ts` files in `fixture-manifest.txt` (only the mod fixture does, and it holds exactly one test per `*.test.ts` file), and for each, `TESTS-MATCH=1` means the executed-test count quoted from the runner output equals the number of `*.test.ts` files (a runner count of zero, or no count printed, is an empty sweep: the R04 card then reads `UNOBSERVED(test-count-not-printed; raw; quote)`).
**Turns red when** the fixture count differs from the row count (a fixture is missing, or a composite is added), a command kind has fewer files than fixtures, an install claims success without the verbatim line, `plugin test` ran on a fixture with zero tests, or the executed count differs from the test-file count.

## §D.7 AC-007 — Runtime cells carry sentinel, control, and captured value (RN-007 / GP-007)

**Given** `verdict.md`, `evidence/commands.log` and `evidence/raw/claude-runtime-<ID>.txt` for the ten runtime rows R01-R04 and R06-R11 (the rows of §5 whose `channels` column is `static+runtime`; R05, R12, R13 and R14 are `static` rows with no runtime channel and are outside this sweep)
**When** `sh .moai/reports/t1434/check-verdict.sh runtime .moai/reports/t1434/verdict.md` runs
**Then** it iterates the **cells of `verdict.md`** (not a glob of files), prints `RUNTIME-ROWS=<k> SWEPT=<n> NOT-SWEPT-UNOBSERVED=<u> BAD-CONTROL=0 BAD-RUNS=0 RUNS=<r> RUNTIME-RUN-CAP=<c> OVER-CAP=0 RESULT=PASS`, and exits 0 only when `k` (counted from §5, so 10) satisfies `n >= 1` and `n+u` equals `k`. For every non-`UNOBSERVED` runtime cell the raw file exists, is non-empty, and contains a `PLUGIN-SENTINEL: <PRESENT|ABSENT> <sentinel>` line, a `PROJECT-CONTROL-SENTINEL: <PRESENT|ABSENT> <control sentinel>` line, and a `CONTROL-FIRED:` line carrying a captured value that includes the run's token (a bare token with no value is `BAD-CONTROL`); the R04 and R08 raw files carry these three generic lines in addition to their own token families (AC-010, AC-012); a model-quoted channel carries `RUNS=3 AGREE=3` for an absence claim (`BAD-RUNS` counts any other); a runtime cell counted in `u` is `UNOBSERVED(<reason>; raw=<file>; quote="<line>")`, because a block of a runtime channel (no authentication, a refused command, a version gap, a control that did not fire, the run cap) must still cite a raw file and quote (AC-008 `EMPTY-REASON`); `RUNS` is the count of `claude -p` lines in `commands.log`, the M1 observation runs included, and `RUNTIME-RUN-CAP` is the value the probe printed, which must equal `k` times 6 (60); `OVER-CAP` is non-zero when `c` differs from that value or `r` exceeds `c`.
**Turns red when** the sweep is empty (every runtime cell `UNOBSERVED`, or no verdict table), a cell is `PLUGIN-OK` or `PROJECT-ONLY` with no runtime file, the control line is the literal token with no value, an absence claim rests on fewer than three agreeing runs, the printed cap is not 60, or the run count exceeds it. All-runtime-blocked is a real outcome that fails this criterion and is reported as `blocked-on-measurement`; it is not relaxed (edge case EC-5).

## §D.8 AC-008 — Every cell has a card whose quoted line is in the cited raw file and relevant to the row (RN-008 / GP-008)

**Given** the 28 cards `evidence/cells/R01-claude.md` through `R14-codex.md`, each cell in `verdict.md` written `<VERDICT> [cell:<ID>-<tool>]`, each card's Evidence field carrying `DECIDING-LINE: "<quote>"` and naming one raw file, each Claude card of a floor row also carrying `STATIC-LINE: "<quote>"`, and the row tokens of `plan.md` §4
**When** `sh .moai/reports/t1434/check-verdict.sh cells .moai/reports/t1434/verdict.md` runs (the script contains no git invocation)
**Then** it prints `CELLS=28 CARDLESS=0 MISSING-FIELDS=0 ORPHAN-RAW=0 QUOTE-MISMATCH=0 EMPTY-REASON=0 RESULT=PASS` and exits 0, where `CELLS` equals twice the row count; `CARDLESS` counts cells without a card; `MISSING-FIELDS` counts cards lacking any of the five fields (Claim, Evidence, Baseline-attribution, Gaps, Residual-risk); `ORPHAN-RAW` counts non-`UNOBSERVED` cells whose cited raw file does not exist or is empty; `QUOTE-MISMATCH` counts cards (of any verdict) whose `DECIDING-LINE` quote, or `STATIC-LINE` quote, is not a fixed-string match inside the cited file (for the `STATIC-LINE`, inside one of that row's own `claude-validate-<ID>`, `claude-validate-strict-<ID>`, `claude-details-<ID>` or `claude-install-<ID>` files), or is a line that cannot decide (a line beginning `#`, or an `EXIT=` trailer), or, on a Claude cell that is not `UNOBSERVED`, contains none of that row's tokens (the sentinel prefix `SENTINEL_<ID>_`, the fixture plugin name, the details token); `EMPTY-REASON` counts cells whose leading vocabulary token is `UNOBSERVED` and whose reason is empty or lacks `raw=<file>; quote="<line>"` (the block is proven by a line of a raw file, checked for presence and eligibility as in `QUOTE-MISMATCH`; a block proof need not name a row token). A cell is classified by the vocabulary token before its first `(`: a static row's `PARTIAL(static-only: ...; runtime=UNOBSERVED(no-runtime-channel))` is a `PARTIAL` cell, and its `runtime=UNOBSERVED(no-runtime-channel)` text needs no raw file or quote because no runtime channel was designed for that row.
**Turns red when** one card is removed, a field is deleted, a cited raw file is renamed or emptied, a quote is altered, an `UNOBSERVED(anything)` has no raw proof (mutant: 26 cells `UNOBSERVED(x)` has `EMPTY-REASON=26`), or a generic line stands for a component (mutant: seven `PARTIAL(static-only)` floor cells whose `DECIDING-LINE` and `STATIC-LINE` are `EXIT=0` or the stamp line reads `QUOTE-MISMATCH=7` here and `FLOOR-MET=0` in AC-014). The per-counter mutants are AC-009's.

## §D.9 AC-009 — check-verdict.sh can fail, re-executed (RN-009 / GP-009)

**Given** the inputs under `evidence/mutants/` for the seven modes of `check-verdict.sh` (cells, runtime, table, r08, r04, rules, codex): one base input per mode that the checker must accept; one mutant per name listed in `plan.md` §6 (40 names: one per counter, two for `QUOTE-MISMATCH`, one `ZERO-SWEEP` per mode), each differing from its mode's base input by that single defect; and the recorded results `evidence/negative-control-checker.txt`
**When** `sh .moai/reports/t1434/check-evidence.sh negative-controls` runs
**Then** it reads the mutant names from `plan.md` §6, **re-executes** every base and every mutant (it does not merely read the recorded file), prints `BASES-GREEN=7 MUTANTS=40 MUTANTS-RED=40 RESULT=PASS` and exits 0: every base exits 0; every mutant exits non-zero and its named counter is at least 1; and each recorded line `MUTANT=<mode:COUNTER> EXIT=<n> COUNTER=<name>=<v>` matches the re-executed result. A mutant that exits 0, or a name without a mutant, is a defect against this criterion. `check-evidence.sh` is a measurement helper (see the header) and its modes carry no mutants; this criterion's own loop is therefore one of the readings not observed failing, named in the verdict's Residual-risk.
**Turns red when** any name of `plan.md` §6 has no mutant (a checker that computes only `ORPHAN-RAW` passes the other counters' mutants, so `MUTANTS-RED < MUTANTS`), a mutant exits 0, a base exits non-zero (a checker that fails every input would otherwise pass every mutant), or a recorded line disagrees with the re-run.

## §D.10 AC-010 — Settings-hook equivalence fully observed (RN-010 / GP-010)

**Given** `evidence/raw/claude-runtime-R08.txt`, `evidence/raw/r08-entry-diff.txt`, and `evidence/env.txt`
**When** `sh .moai/reports/t1434/check-verdict.sh r08 .moai/reports/t1434/evidence/raw/claude-runtime-R08.txt` runs
**Then** it prints `EXEC-BAD=0 EXPANSION-BAD=0 EPOCH-BAD=0 ENTRY-DIFF-BAD=0`, the recomputed `EXPANSION` of each copy, `TIMEOUT-CUTOFF` and `SECONDS`, and `RESULT=PASS`, exit 0: the file carries, besides its own tokens, the generic `PLUGIN-SENTINEL:`, `PROJECT-CONTROL-SENTINEL:` and `CONTROL-FIRED:` lines of AC-007; `PLUGIN-HOOK-EXEC: <PRESENT|ABSENT> <token>` and `PROJECT-HOOK-EXEC: <PRESENT|ABSENT> <token>` each present as a line (`EXEC-BAD` counts a missing line, and an `ABSENT` project line, the control, in a cell that is not `UNOBSERVED`; an `ABSENT` plugin line beside a `PRESENT` project line is a valid observation); for each copy whose line is `PRESENT`, `ARG0=<value>`, the argument the marker script received as `$0`, and the checker recomputes `EXPANSION` as `expanded` (the plugin copy's value starts with the fixture plugin path, the project copy's with the scratch project path from `env.txt`), `literal` (the value begins with the unexpanded `${`) or `empty`, so an observed non-expansion is a valid observation and is never typed in by hand (`EXPANSION-BAD` counts a present copy with no `ARG0` line or whose recorded expansion disagrees with the recomputation); `START=<epoch>`, `END=<epoch>` or `END=ABSENT`, and `SESSION-END=<epoch>` present as numbers, with `TIMEOUT-CUTOFF` recomputed as `yes` exactly when `END` is absent and `SESSION-END - START > 5` (otherwise `no`) and `SECONDS` recomputed as `END - START` only when `END` exists (`n/a` when it does not; `EPOCH-BAD` counts a missing or non-numeric epoch or a disagreeing cutoff); `r08-entry-diff.txt` shows each copy equal to the template entry except for the final script-path argument, `${CLAUDE_PLUGIN_ROOT}/hooks/marker.sh` in the plugin copy and `${CLAUDE_PROJECT_DIR}/marker.sh` in the project copy (`ENTRY-DIFF-BAD` counts a copy that differs in more). Alternatively the R08 Claude cell is `UNOBSERVED(<reason>; raw; quote)` and AC-008 checks its proof.
**Turns red when** five identical `PLUGIN-HOOK-EXEC` lines stand in for the five tokens, a present copy has no `ARG0` line or an empty value recorded as expanded, a session that ended before 5 seconds is read as a cutoff (`END` absent, `SESSION-END - START` at most 5), the epochs are absent, or a copy differs from the template entry in more than the final argument.

## §D.11 AC-011 — Rules and instruction rows, three runs, configuration scope (RN-011 / GP-011)

**Given** `evidence/raw/claude-runtime-R09.txt`, `-R10.txt`, `-R11.txt`, the R11 card, `evidence/env.txt`, `evidence/raw/claude-validate-R11.txt`
**When** `sh .moai/reports/t1434/check-verdict.sh rules .moai/reports/t1434/verdict.md` runs
**Then** it prints `RUNS-BAD=0 CONTROL-BAD=0 CONFIG-SCOPE-BAD=0 VALIDATE-MSG-BAD=0 RESULT=PASS` and exits 0: each of the R09, R10, R11 raw files carries `PLUGIN-SENTINEL: <PRESENT|ABSENT> <sentinel>` and `PROJECT-CONTROL-SENTINEL: PRESENT <control sentinel>` (`CONTROL-BAD` counts a file of a non-`UNOBSERVED` cell whose control line is missing or `ABSENT`); R10 carries `PHASE=start` and `PHASE=after-touch` readings; a `PROJECT-ONLY` cell requires `PLUGIN-SENTINEL: ABSENT` in all three runs (`RUNS=3 AGREE=3`) (`RUNS-BAD` counts a `PROJECT-ONLY` cell resting on fewer, and an R10 file missing either phase reading); the R11 card carries `CONFIG-SCOPE: instructionFiles=<value>` equal to the value in `env.txt` (`CONFIG-SCOPE-BAD` counts a missing or different value); the R11 card carries `VALIDATE-MESSAGE: "<verbatim line>"` matched as a fixed string in `claude-validate-R11.txt`, and the cell is not decided by that message (`VALIDATE-MSG-BAD` counts a missing or altered message, or a `DECIDING-LINE` equal to it).
**Turns red when** a `PROJECT-ONLY` verdict has the control absent, only two runs agree, the R11 card omits the configuration value, or the validate message is missing or altered.

## §D.12 AC-012 — Mods row follows the fixed result table (RN-012 / GP-012)

**Given** the R04 fixture bytes in `fixture-manifest.txt`, `evidence/raw/claude-test-R04.txt`, `evidence/raw/claude-runtime-R04.txt`, and the R04 Claude cell
**When** `sh .moai/reports/t1434/check-verdict.sh r04 .moai/reports/t1434/verdict.md` runs
**Then** it prints `TESTS-EXECUTED=<n> TESTS-FOUND=<n> OUTCOME=<allowed> TESTS-GAP=0 PAYLOAD-BAD=0 CELL-MISMATCH=0 RESULT=PASS` and exits 0: the fixture `hooks/hooks.json` is the payload `{ "modules": ["./register.tsx"] }` (`PAYLOAD-BAD` counts a different payload); `TESTS-EXECUTED` equals `TESTS-FOUND` and is at least 1 (`TESTS-GAP` counts a zero or a difference; zero executed is an empty sweep, so a mod fixture with no test file fails); the runtime file carries, besides the generic lines of AC-007, the control marker (`HOOK-CONTROL: <token>`, the same marker logic as a plugin hook in the same session) and the mod marker (`MOD-MARKER: <token>`), and the checker derives `OUTCOME` from the plan's fixed table (control and mod marker present: `PLUGIN-OK` allowed; control present, mod marker absent: at most `PARTIAL(static-only ...)` and the runtime result `UNOBSERVED(modules-not-loaded-under-claude-p; raw; quote)`; control absent: `UNOBSERVED(control-not-fired)`); `CELL-MISMATCH` counts a recorded R04 Claude cell outside the allowed outcome.
**Turns red when** the cell reads `PLUGIN-OK` while the mod marker is absent, the control is absent but the cell claims an observation, the executed test count is zero, or the module payload differs.

## §D.13 AC-013 — Codex under a scratch home, only successful adds counted (RN-013 / GP-013)

**Given** `evidence/commands.log`, `evidence/raw/codex-marketplace-add.txt`, `codex-plugin-add-<ID>.txt`, `codex-plugin-list.txt`, `codex-marketplace-schema-<n>.txt`
**When** `sh .moai/reports/t1434/check-verdict.sh codex .moai/reports/t1434/verdict.md` runs
**Then** it prints `ADDS-OK=<a> LISTED=<a> ADD-MISMATCH=0 VALIDATE-LINES=0 REAL-CODEX-WRITES=0 RESULT=PASS` and exits 0, on the route actually taken: `ADDS-OK` counts `codex plugin add` raw files whose trailer is `EXIT=0`, and `LISTED` counts those whose plugin name appears in `codex-plugin-list.txt`; `ADD-MISMATCH` counts the two differing (a failed add never counts; a successful add absent from the list is a mismatch) and counts `ADDS-OK >= 1` while AC-003 recorded Codex `ISOLATION-FAILED` (then `ADDS-OK` must be 0 and every Codex cell `UNOBSERVED` with a quoted proof); `VALIDATE-LINES` is the number of `codex plugin validate` lines in `commands.log` and is 0 (the subcommand is absent per the observed `codex plugin --help`); `REAL-CODEX-WRITES` counts Codex `plugin add`, `plugin remove`, `marketplace` or `exec` lines with `home=real` and is 0; each rejection message that shaped the Codex marketplace manifest is saved as a `codex-marketplace-schema-<n>` file.
**Turns red when** a failed add is counted, `codex plugin list` evidence is missing, a Codex write or a `codex exec` targets a real home, or a `codex plugin validate` line appears.

## §D.14 AC-014 — Verdict table: shape, consistency, derived yield floor (RN-014 / GP-014)

**Given** `.moai/reports/t1434/verdict.md`, `spec.md` §5, the cards and raw files of the floor rows, and the row tokens of `plan.md` §4
**When** `sh .moai/reports/t1434/check-verdict.sh table .moai/reports/t1434/verdict.md` runs
**Then** it prints `ROWS=14 HEADER=1 VOCAB-BAD=0 HOME-BAD=0 INCONSISTENT=0 CONSEQ-BAD=0 FLOOR=<f> FLOOR-MET=<f> RESULT=PASS` and exits 0. `ROWS` equals the row count of §5 and `HEADER=1` is the single line `| ID | Component | Claude plugin | Codex plugin | Recommended home | Consequence |`; `VOCAB-BAD` counts cells outside `PLUGIN-OK`, `PROJECT-ONLY`, `PARTIAL(...)`, `UNOBSERVED(...)` (classified by the token before the first `(`); `HOME-BAD` counts homes outside `plugin`, `project-scaffold`, `split(...)`, `undetermined(...)`; `INCONSISTENT` counts rows where the home contradicts the Claude cell (`plugin` needs `PLUGIN-OK`, `project-scaffold` needs `PROJECT-ONLY`, `split` needs `PARTIAL`, `UNOBSERVED` needs `undetermined`); `CONSEQ-BAD` counts consequence cells lacking a `marketplace:` clause or an `init-shrink:` clause. `FLOOR` is derived, not typed: the checker counts the rows marked `floor=yes` in §5 of `spec.md` (R01, R02, R03, R05, R06, R07, R14, so 7), and `FLOOR-MET` counts those whose Claude cell is not `UNOBSERVED` and whose card carries a `STATIC-LINE` that is a fixed-string match in one of that row's own static raw files (`claude-validate-<ID>`, `claude-validate-strict-<ID>`, `claude-details-<ID>`, `claude-install-<ID>`), is not a `#` or `EXIT=` line, and contains one of that row's tokens (the sentinel prefix `SENTINEL_<ID>_`, the fixture plugin name, the details token: `Skills (1)` for R01, `Agents (1)` for R02, `MCP servers (1)` for R06, `Hooks (1)` for R07, and for R03, R05 and R14, whose only static evidence is validate or install output, the plugin name `p-command`, `p-outstyle`, `p-manifest`); the two are equal. A floor row decided by runtime evidence (for example R03 `PLUGIN-OK` at runtime) still carries its `STATIC-LINE`, so an honest observed result does not fail the floor.
**Turns red when** every home reads `plugin` while every cell reads `PROJECT-ONLY` (`INCONSISTENT=14`), the table is 14 empty rows, 26 cells are `UNOBSERVED(x)` (`FLOOR-MET` at most 2 against `FLOOR=7`), or seven floor cells quote `EXIT=0` or the stamp line (`FLOOR-MET=0`). A floor row that cannot honestly reach a non-`UNOBSERVED` Claude cell fails this criterion and is reported as `blocked-on-measurement` (for example R05 if the install is rejected and nothing else names the fixture); the criterion is not relaxed to pass.

## §D.15 AC-015 — Versions, hash, non-goal, durable carrier (RN-015 / GP-015)

**Given** every raw file beginning `# stamp: claude=<v> codex=<v> fixture_sha256=<h> date=<d>`, `evidence/version-start.txt` (M1), `evidence/version-end.txt` (M5), `evidence/fixture-manifest.txt`, `verdict.md`, and `progress.md §E.2`
**When** `sh .moai/reports/t1434/check-evidence.sh stamps` runs
**Then** it prints `RAW=<n> STAMPED=<n> VERSION-EQUAL=1 SHA-RECOMPUTED=1 NONGOAL=1 CARRIER-ROWS=14 CARRIER-SHA=1 RESULT=PASS` and exits 0: `STAMPED` equals `RAW` (the count of files under `evidence/raw/`) and is at least 1; `VERSION-EQUAL=1` means the start and end readings of `claude --version` and `codex --version` are equal by recomputation (a `diff` of the two files, not a token); every stamp's `fixture_sha256` equals the value the checker recomputes from `fixture-manifest.txt`; `verdict.md` carries `Non-goal: user-global install mode removal` exactly once; `progress.md §E.2` carries the 14 table rows (`^| R[0-9][0-9] `) identical to the rows of `verdict.md`, a line `verdict_sha256: <h>` equal to the sha256 the checker computes over `verdict.md`, and an `evidence index:` block whose `RAW-FILES=<n>` equals `RAW`.
**Turns red when** a raw file has no stamp, the version readings differ (auto-update mid-run: the batch is re-run), a stamp hash differs from the recomputed one, the non-goal line is missing, `progress.md §E.2` is still a placeholder (RN-015), or `verdict.md` was edited after its hash was recorded.

## Edge Cases

- **EC-1 — Static isolation fails for one tool.** Record `ISOLATION-FAILED <tool>` with the quoted proof, run no further registry-writing command for that tool, and carry its static cells as `UNOBSERVED`; AC-003 judges the route actually taken (D3).
- **EC-2 — A fixture fails validate.** A failing validate is an observation (record verbatim), not a probe fault; the cell may still be measured at runtime if the plugin loads.
- **EC-3 — A command is refused (permission, guard).** The refusal is named in the card's Gaps with what was done instead; the cell is `UNOBSERVED` unless an observed channel remains (`verification-claim-integrity.md` §3.1).
- **EC-4 — Auto-update changes a CLI version mid-run.** AC-015 fails; the batch is re-run.
- **EC-5 — Every runtime row is blocked** (authentication, approval, run cap). AC-007 fails with `SWEPT=0`; the lane's blocker report goes to the leader and the card reports `blocked-on-measurement`; no criterion is weakened.
- **EC-6 — A model-dependent channel disagrees between its three runs.** The cell is `PARTIAL(nondeterministic)` or `UNOBSERVED`; never `PLUGIN-OK`, and never an absence claim.
- **EC-7 — Scrubbing a `CLAUDE_CODE_*` name breaks authentication on the real route.** Keep that name, list it in `env-scrub.txt` with the observed reason, re-run; AC-002 then requires the listed reason.
- **EC-8 — A protected real-home row changes during a command.** A row that names no fixture identifier is `AMBIENT` (possibly another session, possibly the nested session's own tool-internal refresh; not read as either) and is left untouched. A `LEAK` from a permitted verb stops the probe's real-home commands, is recorded as `LEAK-FINDING: <cmd-id>` in the verdict Gaps, and does not fail AC-005; a `LEAK` from a forbidden verb fails it.

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Isolation | scratch homes empty to non-empty, no credentials, real commands observe-only | AC-003, AC-004 |
| Attribution | every changed real-home row `AMBIENT` or `LEAK`, no forbidden-verb `LEAK`, builder control observed red on changed bytes | AC-005 |
| Verdict integrity | cells checker exit 0 and every `check-verdict.sh` counter mutant exits non-zero when re-executed against a passing base | AC-008, AC-009 |
| Observation honesty | every non-`UNOBSERVED` runtime cell has sentinel, control with captured value, three runs for absence | AC-007, AC-010, AC-011, AC-012 |
| Yield | `FLOOR-MET` equals the derived `FLOOR` (rows marked `floor=yes`), each floor cell tied to a row token | AC-014 |
| Durable carrier | `progress.md §E.2` rows and hash match `verdict.md` | AC-015 |
| Product code untouched | empty product-path diff against the run-start SHA, committed and uncommitted, with a non-empty control | AC-001 |
| plan-auditor | PASS at the Tier M threshold | auditor report |

## Definition of Done

- [ ] All 15 acceptance criteria green with verbatim evidence under `.moai/reports/t1434/evidence/` (local-only) and carried into `progress.md §E.2`
- [ ] `.moai/reports/t1434/verdict.md` carries the 14-row table, both columns, recommended-home, consequence, version stamps, and the non-goal line
- [ ] The manifest-builder control and all 40 `check-verdict.sh` mutants observed failing against 7 passing bases, and re-executable; the `check-evidence.sh` modes (measurement helpers) are stated in the verdict's Residual-risk as not observed failing
- [ ] No registry- or settings-writing command ran against a real home; every changed protected real-home row is attributed `AMBIENT` or `LEAK`, with no forbidden-verb `LEAK` and every permitted-verb `LEAK` written into the verdict Gaps
- [ ] No moai product code, template, `.mcp.json`, or settings change in the card diff, committed or uncommitted; the only tracked writes are the run commit's `progress.md §E.2` and `§E.3` blocks and the `spec.md` `status:` and `updated:` lines, written by manager-develop
- [ ] Unobservable cells are `UNOBSERVED` with a raw file and a quoted line, none inferred from documentation or from validate output alone
- [ ] plan-auditor PASS recorded before run-phase entry
- [ ] Blocked outcome (`blocked-on-measurement`: a floor row without a non-`UNOBSERVED` Claude cell, or every runtime row blocked): complete when the lane's blocker report to the leader names the partial table, the evidence path and the failing criteria (AC-007 or AC-014), the card reports `blocked-on-measurement`, and the leader chooses PASS-with-debt, re-plan or abandon; no criterion is relaxed
