# Acceptance — SPEC-PLUGIN-LOAD-SCOPE-001

All commands are plain, separately invocable, and run from the card worktree root. The exit code is its
own field. A path written `evidence/...`, `verdict.md`, `check-evidence.sh` or `check-verdict.sh` without a
directory prefix is relative to `.moai/reports/t1434/`. Every checker prints its counters, exits non-zero on
any failure, and exits non-zero when its swept set is empty (`verification-completeness.md` §1.1, empty
sweep). Run-phase artifacts under `.moai/reports/t1434/` are **local-only** (`.gitignore:235`); the durable
carrier the follow-up cards read is the `progress.md §E.2` block checked by AC-016.

## §D AC Matrix

| AC | Requirement | Description | Severity | Verification |
|----|-------------|-------------|----------|--------------|
| AC-001 | REQ-001 | Repository writes confined; no product path touched, committed or uncommitted | Should (regression guard) | Direct |
| AC-002 | REQ-002 | Environment recorded first; scrub list enumerated from the live environment; every command logged scrubbed | Must | Direct |
| AC-003 | REQ-003 | Static commands isolated: scratch homes go from empty to non-empty, no credentials in them | Must | Direct |
| AC-004 | REQ-004 | Real-home commands are observe-only; no registry-writing verb touches a real home | Must | Direct |
| AC-005 | REQ-005 | Protected-set manifests per real-home command, attributed AMBIENT or LEAK, builder shown able to fail | Must | Direct |
| AC-006 | REQ-006 | Fixture set complete: count, sentinels, manifests, Codex marketplace manifest, recomputed hash | Must | Direct |
| AC-007 | REQ-007 | Claude static evidence per fixture and per command kind, with the executed-test count | Must | Direct |
| AC-008 | REQ-008 | Runtime cells carry sentinel, project control, and a captured control value; zero sweep fails | Must | Direct |
| AC-009 | REQ-009, REQ-014 | Every cell has a five-field card whose quoted deciding line occurs in the cited raw file | Must | Direct |
| AC-010 | REQ-009, REQ-014 | Both checkers are shown able to fail, one mutant per counter and per mode, re-executed | Must | Direct |
| AC-011 | REQ-010 | Settings-hook equivalence fully observed: both executions, both expansions, cutoff seconds | Must | Direct |
| AC-012 | REQ-011 | Rules and instruction rows carry plugin and project sentinels, three runs, the R11 configuration scope | Must | Direct |
| AC-013 | REQ-012 | Mods row follows the fixed result table, with a nonzero executed-test count | Must | Direct |
| AC-014 | REQ-013 | Codex measured under a scratch home; only successful adds counted; no real-home Codex write | Must | Direct |
| AC-015 | REQ-015, REQ-016 | Verdict table: 14 rows, vocabulary, home consistency, consequence, derived yield floor | Must | Direct |
| AC-016 | REQ-014, REQ-016 | Versions equal at start and end, hash recomputed, non-goal line, durable carrier in progress.md §E.2 | Must | Direct |

Budget check (Tier M ceilings 16/16, counted independently): 16 requirements, 16 acceptance criteria.

## RED-now Ledger (keyed per criterion)

The deliverables do not exist at the pinned tree, so every criterion that reads a deliverable is red for a
stated reason: the work has not run. Each cell was executed in this revision on the worktree whose HEAD is
`676293144`. **Pin caveat:** the files these cells read live under `.moai/reports/t1434/` (local-only, never
in a commit), so the SHA names the tracked tree and the cell records the worktree's state at measurement
time, not a recoverable commit. The iteration-1 ledger row L-1 (`test -d .moai/reports/t1434`) is **retired**:
the plan-auditor created that directory (`plan-audit.md` now sits in it), so the command no longer exits 1;
a cell that is not red is not a RED-now cell. The `test` commands print nothing; the exit code is the field.

| Key | Command | Verbatim stdout | Exit code | Tree SHA | Why red |
|-----|---------|-----------------|-----------|----------|---------|
| RN-001 | `git diff --name-only 802a72235..676293144 -- .moai/specs` | the four paths `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/acceptance.md`, `plan.md`, `progress.md`, `spec.md` | 0 | 676293144 | AC-001 is a regression guard, green at arrival by construction (the real form `git status --porcelain -- internal cmd pkg .claude .mcp.json CLAUDE.md AGENTS.md` prints nothing, exit 0). This is its mutant-input cell: the same instrument over a pathspec that includes a touched path prints a non-empty list, so the instrument is not blind |
| RN-002 | `test -s .moai/reports/t1434/evidence/commands.log` | (empty) | 1 | 676293144 | the command log does not exist; written from M1 |
| RN-003 | `test -s .moai/reports/t1434/evidence/scratch-home-after.txt` | (empty) | 1 | 676293144 | no scratch-home listing exists; written at M1 |
| RN-004 | `grep -c "^CMD: home=real" .moai/reports/t1434/evidence/commands.log` | (empty; stderr `ugrep: warning: .moai/reports/t1434/evidence/commands.log: No such file or directory`) | 2 | 676293144 | the command log does not exist, so no real-home line can be audited |
| RN-005 | `test -s .moai/reports/t1434/evidence/negative-control-manifest.txt` | (empty) | 1 | 676293144 | the manifest-builder control has not run; written at M1 |
| RN-006 | `test -s .moai/reports/t1434/evidence/probe-counts.txt` | (empty) | 1 | 676293144 | the probe has not printed its counts; written at M1 |
| RN-007 | `test -d .moai/reports/t1434/evidence/raw` | (empty) | 1 | 676293144 | no raw evidence directory; created at M2 |
| RN-008 | `test -f .moai/reports/t1434/check-verdict.sh` | (empty) | 1 | 676293144 | the verdict checker is not written; written at M1, finished at M5 |
| RN-009 | `test -f .moai/reports/t1434/verdict.md` | (empty) | 1 | 676293144 | the verdict is not written; written at M5 |
| RN-010 | `test -s .moai/reports/t1434/evidence/negative-control-checker.txt` | (empty) | 1 | 676293144 | the checker mutants have not run; written at M5 |
| RN-011 | `test -s .moai/reports/t1434/evidence/raw/claude-runtime-R08.txt` | (empty) | 1 | 676293144 | the R08 runtime session has not run; written at M3 |
| RN-012 | `test -s .moai/reports/t1434/evidence/raw/claude-runtime-R09.txt` | (empty) | 1 | 676293144 | the R09 runtime session has not run; written at M3 |
| RN-013 | `test -s .moai/reports/t1434/evidence/raw/claude-runtime-R04.txt` | (empty) | 1 | 676293144 | the R04 runtime session has not run; written at M3 |
| RN-014 | `test -s .moai/reports/t1434/evidence/raw/codex-plugin-list.txt` | (empty) | 1 | 676293144 | no Codex list output exists; written at M4 |
| RN-015 | `grep -c "^| ID | Component" .moai/reports/t1434/verdict.md` | (empty; stderr `ugrep: warning: .moai/reports/t1434/verdict.md: No such file or directory`) | 2 | 676293144 | the verdict table does not exist; written at M5 |
| RN-016 | `grep -c "evidence index" .moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/progress.md` | `0` | 1 | 676293144 | `progress.md §E.2` is a placeholder; the durable carrier is written at M5 |

## Green-path Ledger (keyed per criterion)

Each green path names the milestone that flips the criterion and what the passing output becomes. A green
path that would need an unrelated change to flip is disqualified (`verification-completeness.md` §2).

| Key | Flips at | Passing output |
|-----|----------|----------------|
| GP-001 | M5 (the progress.md write) | both product-path forms print nothing, `git diff --name-only BASE..HEAD` lists only `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/` paths, `git status --porcelain --ignored -- .moai/reports/t1434` prints `!! .moai/reports/t1434/` |
| GP-002 | M1 | `check-evidence.sh commands` prints `CMDS=<n> SCRUB-MATCH=<n> ENV-NAMES-MATCH=1 CMD-COUNT-MATCH=1 SECRET-VALUES=0 RESULT=PASS`, exit 0, `<n>` equal to the probe's `CMD-COUNT` |
| GP-003 | M1 (static phase at M2) | `check-evidence.sh isolation` prints `TOOLS=2 ISOLATED=<a> FAILED=<b> ... CRED-FILES=0 RESULT=PASS` with `a+b=2`, exit 0 |
| GP-004 | M3 | `check-evidence.sh observe-only` prints `REAL-CMDS=<n> FORBIDDEN-REAL=0 REAL-OFF-ALLOWLIST=0 SCRATCH-NO-OVERRIDE=0 REAL-WITH-OVERRIDE=0 RESULT=PASS`, exit 0 |
| GP-005 | M1 (builder control), M3 (pairs) | `check-evidence.sh manifests` prints `PAIRS=<n+1> REAL-CMDS=<n> LEAK=0 BUILDER-CONTROL=PASS RESULT=PASS`, exit 0 |
| GP-006 | M1 | `check-evidence.sh fixtures` prints `FIXTURES=15 ROWS=14 ... SHA-RECOMPUTED=1 RESULT=PASS`, exit 0 |
| GP-007 | M2 | `check-evidence.sh static` prints `FIXTURES=15 VALIDATE-PLAIN=15 VALIDATE-STRICT=15 INSTALL-FILES=15 ... TESTS-MATCH=1 RESULT=PASS`, exit 0 |
| GP-008 | M3 (finished M5) | `check-verdict.sh runtime verdict.md` prints `SWEPT=<n>` with `n >= 1` and `BAD-CONTROL=0 BAD-RUNS=0 RESULT=PASS`, exit 0 |
| GP-009 | M5 | `check-verdict.sh cells verdict.md` prints `CELLS=28 CARDLESS=0 MISSING-FIELDS=0 ORPHAN-RAW=0 QUOTE-MISMATCH=0 EMPTY-REASON=0 RESULT=PASS`, exit 0 |
| GP-010 | M5 | `check-evidence.sh negative-controls` prints `MUTANTS=<m> MUTANTS-RED=<m> RESULT=PASS` with `m` equal to the counters plus modes the scripts list, exit 0 |
| GP-011 | M3 | `check-verdict.sh r08 evidence/raw/claude-runtime-R08.txt` prints each token separately with a recomputed `EXPANSION` and `SECONDS`, `RESULT=PASS`, exit 0; or the R08 Claude cell is `UNOBSERVED` with a quoted proof |
| GP-012 | M3 | `check-verdict.sh rules verdict.md` prints `RUNS-OK=3 CONTROL-OK=3 CONFIG-SCOPE=1 RESULT=PASS`, exit 0 |
| GP-013 | M2 (test count), M3 (marker) | `check-verdict.sh r04 verdict.md` prints `TESTS-EXECUTED=<n> TESTS-FOUND=<n> OUTCOME=<allowed> CELL-MATCH=1 RESULT=PASS`, exit 0 |
| GP-014 | M4 | `check-verdict.sh codex verdict.md` prints `ADDS-OK=<a> LISTED=<a> VALIDATE-LINES=0 REAL-CODEX-WRITES=0 RESULT=PASS`, exit 0 |
| GP-015 | M5 | `check-verdict.sh table verdict.md` prints `ROWS=14 HEADER=1 ... FLOOR=7 FLOOR-MET=7 RESULT=PASS`, exit 0 |
| GP-016 | M5 | `check-evidence.sh stamps` prints `STAMPED=<n> RAW=<n> VERSION-EQUAL=1 SHA-RECOMPUTED=1 NONGOAL=1 CARRIER-ROWS=14 CARRIER-SHA=1 RESULT=PASS`, exit 0 |

## §D.1 AC-001 — Repository writes confined (regression guard, RN-001 / GP-001)

**Given** the card branch after the run phase, and the base printed by `git merge-base develop HEAD` (run as its own invocation; call its output BASE)
**When** these four forms run, each as its own invocation: `git diff --name-only BASE..HEAD -- internal cmd pkg .claude .mcp.json CLAUDE.md AGENTS.md`, `git status --porcelain -- internal cmd pkg .claude .mcp.json CLAUDE.md AGENTS.md`, `git diff --name-only BASE..HEAD`, and `git status --porcelain --ignored -- .moai/reports/t1434`
**Then** the first two print nothing (the second catches an uncommitted edit the first cannot), the third is non-empty and every path begins `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/` (the only tracked write; `.moai/reports/t1434/` is ignored and never appears), and the fourth prints `!! .moai/reports/t1434/`, which shows the report directory is local-only. An empty third form means the measurement is reported as unmeasured, not as a pass.
**Turns red when** any product path is modified, committed or not (RN-001 shows the instrument fires on a touched path), or when a path outside `.moai/specs/SPEC-PLUGIN-LOAD-SCOPE-001/` is tracked-changed.

## §D.2 AC-002 — Environment recorded, scrub enumerated live, every command logged (RN-002 / GP-002)

**Given** `evidence/env.txt`, `evidence/env-scrub.txt`, `evidence/commands.log`, `evidence/probe-counts.txt`
**When** `sh .moai/reports/t1434/check-evidence.sh commands` runs
**Then** it prints `CMDS`, `SCRUB-MATCH`, `ENV-NAMES-MATCH`, `CMD-COUNT-MATCH`, `SECRET-VALUES`, `RESULT=PASS` and exits 0. `env.txt` carries `CLAUDE_CONFIG_DIR=`, `CODEX_HOME=`, an `AUTH-STATUS` section, and the live `CLAUDE_CODE_*`/`MOAI_*` names, and is its first written section; the name set in `env-scrub.txt` equals the live-name set in `env.txt` in both directions (a name the probe kept is listed with its reason); every `CMD:` line carries `id=`, `home=scratch` or `home=real`, and an unset list byte-equal to `env-scrub.txt`; `CMDS` equals the `CMD-COUNT` the probe printed from its own counter and is at least 1; no value for a name matching TOKEN, KEY or SECRET appears in `env.txt` (`SECRET-VALUES=0`).
**Turns red when** a `CMD:` line omits the unset prefix, `env-scrub.txt` omits a live name (for example `CLAUDE_CODE_MESSAGING_TOKEN`), a command ran without a log line (counter mismatch), the log is empty (zero sweep), or a secret value was written.

## §D.3 AC-003 — Static commands isolated by observation (RN-003 / GP-003)

**Given** `evidence/scratch-home-before.txt` and `evidence/scratch-home-after.txt`, one section per tool (claude, codex)
**When** `sh .moai/reports/t1434/check-evidence.sh isolation` runs
**Then** it prints `TOOLS=2 ISOLATED=<a> FAILED=<b> BEFORE-EMPTY BEFORE-NONEMPTY AFTER-NONEMPTY CRED-FILES RESULT=PASS` and exits 0. For every isolated tool the before-listing has zero entries and the after-listing has at least one, including the tool's registry file (`plugins/installed_plugins.json` for Claude); `CRED-FILES=0` means no listed name matches `credential|auth|token|keychain|login`; a tool that is not isolated carries `ISOLATION-FAILED <tool> raw=<file>` with a quoted line, and the checker then requires zero registry-writing `CMD:` lines for that tool in `commands.log` (the route actually taken, not an assumed one); `a+b` equals `TOOLS`.
**Turns red when** a before-listing is non-empty, an after-listing is empty with no `ISOLATION-FAILED` proof, a credential-named file sits in a scratch home, or a failed-isolation tool has a registry-writing command logged.

## §D.4 AC-004 — Real-home commands are observe-only (RN-004 / GP-004)

**Given** `evidence/commands.log`, `evidence/forbidden-verbs.txt`, `evidence/readonly-verbs.txt`
**When** `sh .moai/reports/t1434/check-evidence.sh observe-only` runs
**Then** it prints `REAL-CMDS=<n> FORBIDDEN-REAL=0 REAL-OFF-ALLOWLIST=0 SCRATCH-NO-OVERRIDE=0 REAL-WITH-OVERRIDE=0 RESULT=PASS` and exits 0: no `home=real` line contains a verb from `forbidden-verbs.txt` (install, uninstall, marketplace add/remove/update, enable, disable, configure, plugin update, Codex plugin add/remove); every `home=real` line is a `--plugin-dir` session or a verb from `readonly-verbs.txt` (`--version`, `auth status`, `login status`, `plugin list`); every `home=scratch` line carries the `CLAUDE_CONFIG_DIR=<scratch>/config-claude` or `CODEX_HOME=<scratch>/config-codex` override for its tool, and every `home=real` line carries neither.
**Turns red when** a real-home line carries a forbidden verb (mutant: `home=real ... plugin install`), a `home=scratch` line lacks its override (mutant: the same command with the override stripped, which would write the real home), or the log is empty.

## §D.5 AC-005 — Protected-set manifests, attribution, builder control (RN-005 / GP-005)

**Given** `evidence/protected-t0.sha256`, `evidence/manifest/<id>-before.sha256` and `-after.sha256`, `evidence/manifest-diff-<id>.txt`, `evidence/protected-excluded.txt`, `evidence/negative-control-manifest.txt`
**When** `sh .moai/reports/t1434/check-evidence.sh manifests` runs
**Then** it prints `ROWS REQUIRED-ROWS PAIRS REAL-CMDS LEAK EXCLUDED-DECLARED BUILDER-CONTROL RESULT=PASS` and exits 0. `ROWS` equals both the line count of `protected-t0.sha256` and the `ROWS-MANIFEST` the builder printed in `probe-counts.txt` (an equality, not a threshold); the rows `claude:installed_plugins.json`, `claude:known_marketplaces.json`, `claude:settings.json` (of the resolved profile) and `codex:config.toml#plugins-tables` are all present; `PAIRS` equals the number of `home=real` `CMD:` lines plus one (the static-phase window); every row that differs in a `manifest-diff` is classified `AMBIENT` or `LEAK` by whether its changed content names a fixture plugin, the scratch marketplace, the scratch path, or a `SENTINEL_` string, and `LEAK=0`; `protected-excluded.txt` names `.claude.json`, `sessions/`, `session-env/`, `debug/`, `file-history/`, `.codex-global-state.json` and Codex sqlite state, each with a reason. The builder control runs the **builder** on `$SCRATCH/ctl-tree`: `negative-control-manifest.txt` carries `CONTROL unchanged ROWS-DIFF=0`, `CONTROL changed-bytes ROWS-DIFF=1` (one file's bytes changed, same path and length, which a path-hashing builder cannot see), and `CONTROL planted-identifier LEAK=1`, and `BUILDER-CONTROL=PASS` requires all three.
**Turns red when** a required row is missing, `PAIRS` differs from the real-home command count, a changed row names a fixture identifier (`LEAK`), an excluded entry has no reason, or the builder hashes paths instead of content (the changed-bytes control then reads `ROWS-DIFF=0`).

## §D.6 AC-006 — Fixture set complete (RN-006 / GP-006)

**Given** `evidence/probe-counts.txt`, `evidence/fixture-manifest.txt`, `evidence/fixture-sentinels.txt`, and the scratch tree
**When** `sh .moai/reports/t1434/check-evidence.sh fixtures` runs
**Then** it prints `FIXTURES=<n> ROWS=<r> SENTINELS SENTINELS-UNIQUE CLAUDE-MANIFESTS CODEX-MANIFESTS CODEX-MARKETPLACE SHA-RECOMPUTED RESULT=PASS` and exits 0, with `n` equal to `r+1` (`r` is the count of `^| R[0-9][0-9] ` rows of §5 in `spec.md`, which the probe read, so `n` is 15), `SENTINELS` and `SENTINELS-UNIQUE` both equal `n` and each sentinel occurring exactly once in the tree, `CLAUDE-MANIFESTS` and `CODEX-MANIFESTS` each equal to `n`, `CODEX-MARKETPLACE=1` (exactly one non-empty `.agents/plugins/marketplace.json`), and `SHA-RECOMPUTED=1` meaning the sha256 of `fixture-manifest.txt` equals the `fixture_sha256` the probe stamped.
**Turns red when** a fixture lacks its sentinel, two fixtures share one, the count differs from the row count plus one, the Codex marketplace manifest is absent, or the recomputed hash differs from the stamp.

## §D.7 AC-007 — Claude static evidence per fixture and command kind (RN-007 / GP-007)

**Given** the raw files `claude-validate-<ID>`, `claude-validate-strict-<ID>`, `claude-install-<ID>`, `claude-details-<ID>`, `claude-test-<ID>` after M2
**When** `sh .moai/reports/t1434/check-evidence.sh static` runs
**Then** it prints `FIXTURES VALIDATE-PLAIN VALIDATE-STRICT INSTALL-FILES INSTALL-OK DETAILS TEST-FIXTURES TESTS-MATCH RESULT=PASS` and exits 0: `VALIDATE-PLAIN`, `VALIDATE-STRICT` and `INSTALL-FILES` each equal `FIXTURES` (every file ends with an `EXIT=` trailer; a failing install is an observation, not a fault); `INSTALL-OK` is the count of installs whose trailer is `EXIT=0` and whose output contains the verbatim `Successfully installed plugin` line; `DETAILS` equals `INSTALL-OK` and each details file contains an `Always-on` line; `TEST-FIXTURES` equals the number of fixtures holding `*.test.ts` files in `fixture-manifest.txt`, and for each, `TESTS-MATCH=1` means the executed-test count quoted from the runner output equals the number of `*.test.ts` files (a runner count of zero, or no count printed, is an empty sweep: the R04 card then reads `UNOBSERVED(test-count-not-printed; raw; quote)`).
**Turns red when** a command kind has fewer files than fixtures, an install claims success without the verbatim line, `plugin test` ran on a fixture with zero tests, or the executed count differs from the test-file count.

## §D.8 AC-008 — Runtime cells carry sentinel, control, and captured value (RN-008 / GP-008)

**Given** `verdict.md` and `evidence/raw/claude-runtime-<ID>.txt` for the runtime rows R01-R13 (R14 has no runtime channel; the row set is the rows of §5 whose `channels` column is `static+runtime`)
**When** `sh .moai/reports/t1434/check-verdict.sh runtime .moai/reports/t1434/verdict.md` runs
**Then** it iterates the **cells of `verdict.md`** (not a glob of files), prints `RUNTIME-ROWS=<k> SWEPT=<n> NOT-SWEPT-UNOBSERVED=<u> BAD-CONTROL=0 BAD-RUNS=0 RESULT=PASS`, and exits 0 only when `n >= 1` and `n+u` equals `k`. For every non-`UNOBSERVED` runtime cell the raw file exists, is non-empty, and contains a `PLUGIN-SENTINEL:` line, a `PROJECT-CONTROL-SENTINEL:` line, and a `CONTROL-FIRED:` line carrying a captured value that includes the run's token (a bare token with no value is `BAD-CONTROL`); a model-quoted channel carries `RUNS=3 AGREE=3` for an absence claim (`BAD-RUNS` counts any other); the number of `claude -p` runs in `commands.log` does not exceed the printed `RUNTIME-RUN-CAP`.
**Turns red when** the sweep is empty (every runtime cell `UNOBSERVED`, or no verdict table), a cell is `PLUGIN-OK` or `PROJECT-ONLY` with no runtime file, the control line is the literal token with no value, an absence claim rests on fewer than three agreeing runs, or the run cap is exceeded. All-runtime-blocked is a real outcome that fails this criterion and returns a blocker report; it is not relaxed (edge case EC-5).

## §D.9 AC-009 — Every cell has a card whose quoted line is in the cited raw file (RN-009 / GP-009)

**Given** the 28 cards `evidence/cells/R01-claude.md` through `R14-codex.md`, each cell in `verdict.md` written `<VERDICT> [cell:<ID>-<tool>]`, each card's Evidence field carrying `DECIDING-LINE: "<quote>"` and naming one raw file
**When** `sh .moai/reports/t1434/check-verdict.sh cells .moai/reports/t1434/verdict.md` runs (the script contains no git invocation)
**Then** it prints `CELLS=28 CARDLESS=0 MISSING-FIELDS=0 ORPHAN-RAW=0 QUOTE-MISMATCH=0 EMPTY-REASON=0 RESULT=PASS` and exits 0, where `CELLS` equals twice the row count; `CARDLESS` counts cells without a card; `MISSING-FIELDS` counts cards lacking any of Claim, Evidence, Baseline-attribution, Gaps, Residual-risk; `ORPHAN-RAW` counts non-`UNOBSERVED` cells whose cited raw file does not exist or is empty; `QUOTE-MISMATCH` counts cards (of any verdict) whose `DECIDING-LINE` quote is not a fixed-string match inside the cited raw file; `EMPTY-REASON` counts `UNOBSERVED` cells whose reason is empty or lacks `raw=<file>; quote="<line>"` (the block is proven by a line of a raw file, and that line is checked as in `QUOTE-MISMATCH`).
**Turns red when** one card is removed, a field is deleted, a cited raw file is renamed or emptied, a quote is altered, or an `UNOBSERVED(anything)` has no raw proof (mutant: 26 cells `UNOBSERVED(x)` has `EMPTY-REASON=26`); a checker computing only the raw-file test fails the per-counter mutants of AC-010.

## §D.10 AC-010 — Both checkers can fail, re-executed (RN-010 / GP-010)

**Given** the mutant inputs under `evidence/mutants/` (one verdict mutant per counter of `check-verdict.sh cells`: `CARDLESS`, `MISSING-FIELDS`, `ORPHAN-RAW`, `QUOTE-MISMATCH`, `EMPTY-REASON`, plus one `ZERO-SWEEP` mutant with an empty table; one evidence-root mutant per mode of `check-evidence.sh`), and the recorded results `evidence/negative-control-checker.txt` and `evidence/negative-control-evidence.txt`
**When** `sh .moai/reports/t1434/check-evidence.sh negative-controls` runs
**Then** it **re-executes** every mutant (it does not merely read the recorded files), prints `MODES=<a> MUTANTS=<m> MUTANTS-RED=<m> RESULT=PASS` and exits 0, where `a` is the number of modes `check-evidence.sh` lists in its own usage, `m` equals the number of counters `check-verdict.sh` lists plus one (`ZERO-SWEEP`) plus `a`, every mutant exits non-zero and its named counter is at least 1, and each recorded line `MUTANT=<name> EXIT=<n> COUNTER=<name>=<v>` matches the re-executed result. A mutant that exits 0 is a defect against this criterion.
**Turns red when** any counter or mode has no mutant (a checker that computes only `ORPHAN-RAW` passes the other four mutants, so `MUTANTS-RED < MUTANTS`), a mutant exits 0, or a recorded line disagrees with the re-run.

## §D.11 AC-011 — Settings-hook equivalence fully observed (RN-011 / GP-011)

**Given** `evidence/raw/claude-runtime-R08.txt` and `evidence/raw/r08-entry-diff.txt`
**When** `sh .moai/reports/t1434/check-verdict.sh r08 .moai/reports/t1434/evidence/raw/claude-runtime-R08.txt` runs
**Then** it prints one counter per token and `RESULT=PASS`, exit 0: `PLUGIN-HOOK-EXEC` and `PROJECT-HOOK-EXEC` each present as a line carrying the run token; `PROJECT_DIR_EXPANDED=<value>` and `PLUGIN_ROOT_EXPANDED=<value>` each present with a captured value, and the checker recomputes `EXPANSION` for each as `expanded` (value equals the scratch project path from `env.txt` for the project directory, the fixture path for the plugin root), `literal` (value is the unexpanded `${...}` text) or `empty`, so an observed non-expansion is a valid observation and is never typed in by hand; `START=<epoch>`, `END=<epoch>` or `END=ABSENT`, and `SESSION-END=<epoch>` present, with `TIMEOUT-CUTOFF` and `SECONDS` recomputed from them (cutoff is `yes` when `END` is absent and `START` is present); `r08-entry-diff.txt` shows the fixture entry equal to the template entry except for the final script-path element. Alternatively the R08 Claude cell is `UNOBSERVED(<reason>; raw; quote)` and AC-009 checks its proof.
**Turns red when** five identical `PLUGIN-HOOK-EXEC` lines stand in for the five tokens, `PROJECT_DIR_EXPANDED=` has an empty value recorded as expanded, the epochs are absent, or the fixture entry differs from the template entry in more than the script path.

## §D.12 AC-012 — Rules and instruction rows, three runs, configuration scope (RN-012 / GP-012)

**Given** `evidence/raw/claude-runtime-R09.txt`, `-R10.txt`, `-R11.txt`, the R11 card, `evidence/env.txt`, `evidence/raw/claude-validate-R11.txt`
**When** `sh .moai/reports/t1434/check-verdict.sh rules .moai/reports/t1434/verdict.md` runs
**Then** it prints `RUNS-OK=3 CONTROL-OK=3 CONFIG-SCOPE=1 RESULT=PASS` and exits 0: each of R09, R10, R11 raw files carries `PLUGIN-SENTINEL: PRESENT|ABSENT` and `PROJECT-CONTROL-SENTINEL: PRESENT`; R10 carries `PHASE=start` and `PHASE=after-touch` readings; a `PROJECT-ONLY` cell requires `PLUGIN-SENTINEL: ABSENT` in all three runs (`RUNS=3 AGREE=3`) with the control present; the R11 card carries `CONFIG-SCOPE: instructionFiles=<value>` equal to the value in `env.txt`; the R11 card carries `VALIDATE-MESSAGE: "<verbatim line>"` matched as a fixed string in `claude-validate-R11.txt`, and the cell is not decided by that message.
**Turns red when** a `PROJECT-ONLY` verdict has the control absent, only two runs agree, the R11 card omits the configuration value, or the validate message is missing or altered.

## §D.13 AC-013 — Mods row follows the fixed result table (RN-013 / GP-013)

**Given** the R04 fixture bytes in `fixture-manifest.txt`, `evidence/raw/claude-test-R04.txt`, `evidence/raw/claude-runtime-R04.txt`, and the R04 Claude cell
**When** `sh .moai/reports/t1434/check-verdict.sh r04 .moai/reports/t1434/verdict.md` runs
**Then** it prints `TESTS-EXECUTED=<n> TESTS-FOUND=<n> OUTCOME=<allowed> CELL-MATCH=1 RESULT=PASS` and exits 0: the fixture `hooks/hooks.json` is the payload `{ "modules": ["./register.tsx"] }`; `TESTS-EXECUTED` equals `TESTS-FOUND` and is at least 1 (zero executed is an empty sweep, so a mod fixture with no test file fails); the runtime file carries the control marker (`HOOK-CONTROL: <token>`, the same marker logic as a plugin hook in the same session) and the mod marker (`MOD-MARKER: <token>`), and the checker derives `OUTCOME` from the plan's fixed table (control and mod marker present: `PLUGIN-OK` allowed; control present, mod marker absent: at most `PARTIAL(static-only ...)` and the runtime result `UNOBSERVED(modules-not-loaded-under-claude-p; raw; quote)`; control absent: `UNOBSERVED(control-not-fired)`); `CELL-MATCH=1` means the recorded R04 Claude cell is within the allowed outcome.
**Turns red when** the cell reads `PLUGIN-OK` while the mod marker is absent, the control is absent but the cell claims an observation, the executed test count is zero, or the module payload differs.

## §D.14 AC-014 — Codex under a scratch home, only successful adds counted (RN-014 / GP-014)

**Given** `evidence/commands.log`, `evidence/raw/codex-marketplace-add.txt`, `codex-plugin-add-<ID>.txt`, `codex-plugin-list.txt`, `codex-marketplace-schema-<n>.txt`
**When** `sh .moai/reports/t1434/check-verdict.sh codex .moai/reports/t1434/verdict.md` runs
**Then** it prints `ADDS-OK=<a> LISTED=<a> VALIDATE-LINES=0 REAL-CODEX-WRITES=0 RESULT=PASS` and exits 0, on the route actually taken: `ADDS-OK` counts `codex plugin add` raw files whose trailer is `EXIT=0`, and `LISTED` counts those whose plugin name appears in `codex-plugin-list.txt`, the two equal (a failed add never counts; a successful add absent from the list is a mismatch); `VALIDATE-LINES` is the number of `codex plugin validate` lines in `commands.log` and is 0 (the subcommand is absent per the observed `codex plugin --help`); `REAL-CODEX-WRITES` counts Codex `plugin add`, `plugin remove` or `marketplace` lines with `home=real` and is 0; when AC-003 reports Codex `ISOLATION-FAILED`, `ADDS-OK` is 0 and every Codex cell is `UNOBSERVED` with a quoted proof; each rejection message that shaped the Codex marketplace manifest is saved as a `codex-marketplace-schema-<n>` file.
**Turns red when** a failed add is counted, `codex plugin list` evidence is missing, a Codex write targets a real home, or a `codex plugin validate` line appears.

## §D.15 AC-015 — Verdict table: shape, consistency, derived yield floor (RN-015 / GP-015)

**Given** `.moai/reports/t1434/verdict.md` and `spec.md` §5
**When** `sh .moai/reports/t1434/check-verdict.sh table .moai/reports/t1434/verdict.md` runs
**Then** it prints `ROWS=14 HEADER=1 VOCAB-BAD=0 HOME-BAD=0 INCONSISTENT=0 CONSEQ-BAD=0 FLOOR=<f> FLOOR-MET=<f> RESULT=PASS` and exits 0. `ROWS` equals the row count of §5 and `HEADER=1` is the single line `| ID | Component | Claude plugin | Codex plugin | Recommended home | Consequence |`; `VOCAB-BAD` counts cells outside `PLUGIN-OK`, `PROJECT-ONLY`, `PARTIAL(...)`, `UNOBSERVED(...)`; `HOME-BAD` counts homes outside `plugin`, `project-scaffold`, `split(...)`, `undetermined(...)`; `INCONSISTENT` counts rows where the home contradicts the Claude cell (`plugin` needs `PLUGIN-OK`, `project-scaffold` needs `PROJECT-ONLY`, `split` needs `PARTIAL`, `UNOBSERVED` needs `undetermined`); `CONSEQ-BAD` counts consequence cells lacking a `marketplace:` clause or an `init-shrink:` clause. `FLOOR` is derived, not typed: the checker counts the rows marked `floor=yes` in §5 of `spec.md` (R01, R02, R03, R05, R06, R07, R14, so 7), and `FLOOR-MET` counts those whose Claude cell is not `UNOBSERVED` and whose card cites a static raw file (`claude-validate-*`, `claude-details-*` or `claude-install-*`) with a quote that is a fixed-string match in it; the two are equal.
**Turns red when** every home reads `plugin` while every cell reads `PROJECT-ONLY` (`INCONSISTENT=14`), the table is 14 empty rows, or 26 cells are `UNOBSERVED(x)` (`FLOOR-MET` at most 2 against `FLOOR=7`). A floor row that cannot honestly reach a non-`UNOBSERVED` Claude cell fails this criterion and returns a blocker report (for example R05, whose static channel is not documented); the criterion is not relaxed to pass.

## §D.16 AC-016 — Versions, hash, non-goal, durable carrier (RN-016 / GP-016)

**Given** every raw file beginning `# stamp: claude=<v> codex=<v> fixture_sha256=<h> date=<d>`, `evidence/version-start.txt`, `evidence/version-end.txt`, `verdict.md`, and `progress.md §E.2`
**When** `sh .moai/reports/t1434/check-evidence.sh stamps` runs
**Then** it prints `RAW=<n> STAMPED=<n> VERSION-EQUAL=1 SHA-RECOMPUTED=1 NONGOAL=1 CARRIER-ROWS=14 CARRIER-SHA=1 RESULT=PASS` and exits 0: `STAMPED` equals `RAW` (the count of files under `evidence/raw/`) and is at least 1; `VERSION-EQUAL=1` means the start and end readings of `claude --version` and `codex --version` are equal by recomputation (a `diff` of the two files, not a token); every stamp's `fixture_sha256` equals the value the checker recomputes from `fixture-manifest.txt`; `verdict.md` carries `Non-goal: user-global install mode removal` exactly once; `progress.md §E.2` carries the 14 table rows (`^| R[0-9][0-9] `) identical to the rows of `verdict.md`, a line `verdict_sha256: <h>` equal to the sha256 the checker computes over `verdict.md`, and an `evidence index:` block whose `RAW-FILES=<n>` equals `RAW`.
**Turns red when** a raw file has no stamp, the version readings differ (auto-update mid-run: the batch is re-run), a stamp hash differs from the recomputed one, the non-goal line is missing, `progress.md §E.2` is still a placeholder (RN-016), or `verdict.md` was edited after its hash was recorded.

## Edge Cases

- **EC-1 — Static isolation fails for one tool.** Record `ISOLATION-FAILED <tool>` with the quoted proof, run no registry-writing command for that tool, and carry its static cells as `UNOBSERVED`; AC-003 judges the route actually taken (D3).
- **EC-2 — A fixture fails validate.** A failing validate is an observation (record verbatim), not a probe fault; the cell may still be measured at runtime if the plugin loads.
- **EC-3 — A command is refused (permission, guard).** The refusal is named in the card's Gaps with what was done instead; the cell is `UNOBSERVED` unless an observed channel remains (`verification-claim-integrity.md` §3.1).
- **EC-4 — Auto-update changes a CLI version mid-run.** AC-016 fails; the batch is re-run.
- **EC-5 — Every runtime row is blocked** (authentication, approval, run cap). AC-008 fails with `SWEPT=0` and the run returns a blocker report to the orchestrator; no criterion is weakened.
- **EC-6 — A model-dependent channel disagrees between its three runs.** The cell is `PARTIAL(nondeterministic)` or `UNOBSERVED`; never `PLUGIN-OK`, and never an absence claim.
- **EC-7 — Scrubbing a `CLAUDE_CODE_*` name breaks authentication on the real route.** Keep that name, list it in `env-scrub.txt` with the observed reason, re-run; AC-002 then requires the listed reason.
- **EC-8 — Another session rewrites a protected real-home row during a command.** The row is reported `AMBIENT` (it names no fixture identifier) and left untouched; only a `LEAK` fails AC-005.

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Isolation | scratch homes empty to non-empty, no credentials, real commands observe-only | AC-003, AC-004 |
| Attribution | every changed real-home row `AMBIENT` or `LEAK`, `LEAK=0`, builder control observed red on changed bytes | AC-005 |
| Verdict integrity | cells checker exit 0 and every counter/mode mutant exits non-zero when re-executed | AC-009, AC-010 |
| Observation honesty | every non-`UNOBSERVED` runtime cell has sentinel, control with captured value, three runs for absence | AC-008, AC-011, AC-012, AC-013 |
| Yield | `FLOOR-MET` equals the derived `FLOOR` (rows marked `floor=yes`) | AC-015 |
| Durable carrier | `progress.md §E.2` rows and hash match `verdict.md` | AC-016 |
| Product code untouched | empty committed and uncommitted diff over product paths with non-empty control | AC-001 |
| plan-auditor | PASS at the Tier M threshold | auditor report |

## Definition of Done

- [ ] All 16 acceptance criteria green with verbatim evidence under `.moai/reports/t1434/evidence/` (local-only) and carried into `progress.md §E.2`
- [ ] `.moai/reports/t1434/verdict.md` carries the 14-row table, both columns, recommended-home, consequence, version stamps, and the non-goal line
- [ ] Every checker mutant and the manifest-builder control observed failing and re-executable
- [ ] No registry- or settings-writing command ran against a real home; every changed protected real-home row is attributed `AMBIENT` or `LEAK` with `LEAK=0`
- [ ] No moai product code, template, `.mcp.json`, or settings change in the card diff, committed or uncommitted
- [ ] Unobservable cells are `UNOBSERVED` with a raw file and a quoted line, none inferred from documentation or from validate output alone
- [ ] plan-auditor PASS recorded before run-phase entry
