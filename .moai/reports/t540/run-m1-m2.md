# t540 run-phase evidence — SPEC-CODEX-SKILL-PATH-SLASH-001, M1 + M2

Card t540 · worktree `.claude/worktrees/t540` · branch `WT-codex-path-escape` ·
cycle_type **tdd** · Tier M.

Scope of this run: **M1 (separator seam) and M2 (publisher) ONLY.** M3 (reader-side conversion) and
M0/AC-CSPS-001 (the Windows gate) are NOT in this run and are recorded as open in §5.

Card base, re-derived at read time (`git merge-base origin/develop HEAD`):
`9ce7926377236aa837294cce9a214cb545751e7c`. Commits: `2d66ebad4` (M1), `dc71e8ef9` (M2).

---

## 1. Claim

1. `toConfigPath(p, sep)` / `fromConfigPath(p, sep)` exist in `internal/cli/codex_config_path.go`
   with `var configPathSeparator = filepath.Separator` as the injection point, following the
   package's existing seam convention (`osStatFn`, `codexUserHomeDir`). The conversion is
   separator-aware and is the identity when `sep == '/'`.
2. `upsertCodexSkillDisable` converts the incoming path with `toConfigPath` **before** the
   "cannot carry verbatim" guard, and compares both sides of its single path comparison on the
   `toConfigPath` form, so an existing backslash-shaped entry is Updated rather than duplicated.
3. AC-CSPS-002, AC-CSPS-003 (both arms), AC-CSPS-007 (both arms) are GREEN, and AC-CSPS-004 arm B
   (M1's stated exit) is GREEN. t502 F4 is closed.
4. AC-CSPS-006's predicate is satisfied and AC-CSPS-008's probe is empty under a non-zero control.

## 2. Evidence

### 2.1 M1 RED 1 — the functions do not exist

```
$ go test ./internal/cli/... -run 'TestConfigPath' -timeout 1800s
internal/cli/codex_config_path_test.go:89:5: undefined: configPathSeparator
internal/cli/codex_config_path_test.go:90:68: undefined: configPathSeparator
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

### 2.2 M1 RED 2 — the identity-stub mutant (mandatory mutation arm)

Bodies replaced with `return p` for every separator:

```
$ go test ./internal/cli/ -run 'TestConfigPath' -timeout 1800s -v
=== RUN   TestConfigPathToConfigPathConvertsHostSeparator
    codex_config_path_test.go:30: toConfigPath(windows form, '\\') = "C:\\Users\\u\\.codex\\skills\\probe\\SKILL.md", want "C:/Users/u/.codex/skills/probe/SKILL.md"
--- FAIL: TestConfigPathToConfigPathConvertsHostSeparator (0.00s)
=== RUN   TestConfigPathFromConfigPathRestoresHostSeparator
    codex_config_path_test.go:41: fromConfigPath(slash form, '\\') = "C:/Users/u/SKILL.md", want "C:\\Users\\u\\SKILL.md"
--- FAIL: TestConfigPathFromConfigPathRestoresHostSeparator (0.00s)
=== RUN   TestConfigPathIsIdentityOnSlashSeparator
--- PASS: TestConfigPathIsIdentityOnSlashSeparator (0.00s)
=== RUN   TestConfigPathRoundTripsUnderWindowsSeparator
--- PASS: TestConfigPathRoundTripsUnderWindowsSeparator (0.00s)
=== RUN   TestConfigPathSeparatorSeamTracksTheHost
--- PASS: TestConfigPathSeparatorSeamTracksTheHost (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.303s
```

### 2.3 M1 GREEN

```
$ go test ./internal/cli/ -run 'TestConfigPath' -timeout 1800s -v
--- PASS: TestConfigPathToConfigPathConvertsHostSeparator (0.00s)
--- PASS: TestConfigPathFromConfigPathRestoresHostSeparator (0.00s)
--- PASS: TestConfigPathIsIdentityOnSlashSeparator (0.00s)
--- PASS: TestConfigPathRoundTripsUnderWindowsSeparator (0.00s)
--- PASS: TestConfigPathSeparatorSeamTracksTheHost (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.198s
```

### 2.4 M2 RED — publisher not yet converting

```
$ go test ./internal/cli/ -run 'TestUpsertCodexSkillDisable(PublishesWindowsPathInSlashForm|StillRefusesUnrepresentableChars|RefusesBackslashOnSlashHost|UpdatesExistingBackslashEntry|SkipsMixedShapeDuplicates)' -timeout 1800s -v

=== RUN   TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm
    codex_skills_disable_path_test.go:81: action = 3 (reason "the path contains a character this config format cannot carry verbatim (\"C:\\\\Users\\\\u\\\\.codex\\\\skills\\\\probe\\\\SKILL.md\")"), want appended
--- FAIL: TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm (0.00s)
=== RUN   TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars
--- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars (0.00s)
    --- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars///quote (0.00s)
    --- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars///lf (0.00s)
    --- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars///cr (0.00s)
    --- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars/\/quote (0.00s)
    --- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars/\/lf (0.00s)
    --- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars/\/cr (0.00s)
=== RUN   TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost
--- PASS: TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost (0.00s)
=== RUN   TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry
    codex_skills_disable_path_test.go:211: action = 3 (reason "the path contains a character this config format cannot carry verbatim (\"C:\\\\Users\\\\u\\\\.codex\\\\skills\\\\probe\\\\SKILL.md\")"), want updated
--- FAIL: TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry (0.00s)
=== RUN   TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates
    codex_skills_disable_path_test.go:254: reason = "the path contains a character this config format cannot carry verbatim (\"C:\\\\Users\\\\u\\\\.codex\\\\skills\\\\probe\\\\SKILL.md\")", want the duplicate-count skip naming 2 entries
--- FAIL: TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.964s
```

**The two AC-CSPS-003 arms are GREEN here and are reported as regression guards, not as RED
evidence.** The guard they assert already existed; what they discriminate is the
unconditional-`ReplaceAll` mutant (§2.6).

### 2.5 M2 GREEN — new and pre-existing upsert tests together

```
$ go test ./internal/cli/ -run 'TestUpsertCodexSkillDisable|TestConfigPath' -timeout 1800s -v
--- PASS: TestConfigPathToConfigPathConvertsHostSeparator (0.00s)
--- PASS: TestConfigPathFromConfigPathRestoresHostSeparator (0.00s)
--- PASS: TestConfigPathIsIdentityOnSlashSeparator (0.00s)
--- PASS: TestConfigPathRoundTripsUnderWindowsSeparator (0.00s)
--- PASS: TestConfigPathSeparatorSeamTracksTheHost (0.00s)
--- PASS: TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm (0.00s)
--- PASS: TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars (0.00s)
--- PASS: TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost (0.00s)
--- PASS: TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry (0.00s)
--- PASS: TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates (0.00s)
--- PASS: TestUpsertCodexSkillDisableAppendsOneEntryWithFalse (0.00s)
--- PASS: TestUpsertCodexSkillDisableInsertsMissingEnabledKey (0.00s)
--- PASS: TestUpsertCodexSkillDisableUpdatesExistingEntry (0.00s)
--- PASS: TestUpsertCodexSkillDisableIdempotentBytes (0.00s)
--- PASS: TestUpsertCodexSkillDisablePreservesSurroundings (0.00s)
--- PASS: TestUpsertCodexSkillDisableSkipsUnrecognisedEntry (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.995s
```

### 2.6 Mutants

Both mandatory mutants were applied to a copy-restored `codex_config_path.go`; the file was
restored from a pristine copy afterwards and `git diff --stat -- internal/cli/codex_config_path.go`
returned empty.

**Mutant 1 — identity stub** (`toConfigPath`/`fromConfigPath` return their input for every
separator):

```
--- FAIL: TestConfigPathToConfigPathConvertsHostSeparator
--- FAIL: TestConfigPathFromConfigPathRestoresHostSeparator          <- AC-CSPS-004 arm B
--- FAIL: TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm <- AC-CSPS-002
--- FAIL: TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry   <- AC-CSPS-007 arm 1
--- FAIL: TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates       <- AC-CSPS-007 arm 2
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.013s
```

CAUGHT, and the DoD requirement (must fail AC-CSPS-004 arm B **and** AC-CSPS-002) is met.

*Mutants NOT caught by, recorded because they draw the guard's boundary:*
`TestConfigPathIsIdentityOnSlashSeparator`, `TestConfigPathRoundTripsUnderWindowsSeparator` (an
identity function round-trips trivially — this test asserts a property the stub also has, so it is
not a discriminator), `TestConfigPathSeparatorSeamTracksTheHost`, and all AC-CSPS-003 arms.

**Mutant 2 — unconditional `strings.ReplaceAll(p, "\\", "/")` in `toConfigPath`:**

```
--- FAIL: TestConfigPathIsIdentityOnSlashSeparator
--- FAIL: TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost     <- AC-CSPS-003 arm 2
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.004s
```

CAUGHT by AC-CSPS-003 arm 2, exactly as REQ-CSPS-010 requires.

*NOT caught by:* `TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars` (expected — quote /
LF / CR are untouched by a backslash replacement), and every AC-CSPS-002 / 007 assertion (they set
the seam to `'\\'`, where both formulations agree).

### 2.7 AC-CSPS-006 — executed-test control

```
$ go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t540/ac-006-base.log 2>&1; echo "rc=$?"
rc=0
$ /usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006-base.log
6886
```
(base taken at pre-flight on `b4ce67468`, **before** M1 landed)

```
$ go test ./internal/cli/... -timeout 1800s -v > .moai/reports/t540/ac-006.log 2>&1; echo "rc=$?"
rc=0
$ /usr/bin/grep -c -- '--- PASS: ' .moai/reports/t540/ac-006.log
6902
$ /usr/bin/grep -c -- '--- FAIL: ' .moai/reports/t540/ac-006.log
0
```

Predicate `AFTER >= BEFORE AND AFTER > 0`: **6902 >= 6886 and 6902 > 0 → PASS.** The delta of +16 is
exactly the tests added (5 `TestConfigPath*` + 1 + 7 + 1 + 1 + 1 in the new disable-path file), so
no pre-existing test was lost. No pre-existing test file was edited: the two changed test files are
both new (§2.9).

### 2.8 Build, vet, lint

```
$ go build ./...                           → rc=0, no output
$ GOOS=windows GOARCH=amd64 go build ./... → rc=0, no output
$ go vet ./internal/cli/...                → rc=0, no output
$ golangci-lint run --timeout=5m ./internal/cli/...
0 issues.
```

### 2.9 AC-CSPS-008 — scope pin (control first)

Captured at `dc71e8ef9` (the M2 commit), **before** this report and the `progress.md` update landed
— so a re-run at a later HEAD legitimately shows more control rows. The probe is what the criterion
asserts, and it is empty at every HEAD on this branch.

```
$ git merge-base origin/develop HEAD
9ce7926377236aa837294cce9a214cb545751e7c

$ git diff --name-only 9ce7926377236aa837294cce9a214cb545751e7c..HEAD
.moai/reports/t540/ac-006-base.log
.moai/reports/t540/ac-006.log
.moai/reports/t540/lab/ts.go
.moai/reports/t540/plan-audit.md
.moai/reports/t540/reader-census.md
.moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/acceptance.md
.moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/plan.md
.moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/progress.md
.moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/spec.md
internal/cli/codex_config_path.go
internal/cli/codex_config_path_test.go
internal/cli/codex_skills_disable.go
internal/cli/codex_skills_disable_path_test.go

$ git diff --name-only 9ce7926377236aa837294cce9a214cb545751e7c..HEAD -- internal/codexwiring/skills.go
(no output)
```

Control non-empty → measurable. Probe empty → the parser is unmodified.

**Probe discriminator.** The same pathspec form against a path known to be changed returns a row,
so the empty probe above is a real absence rather than an always-empty filter:

```
$ git diff --name-only 9ce7926377236aa837294cce9a214cb545751e7c..HEAD -- internal/cli/codex_config_path.go
internal/cli/codex_config_path.go
```

This is a filter-liveness check, not the parser-touch mutant the DoD names — that mutant needs a
commit touching the parser, and AC-CSPS-008 is M4's exit, outside this run.

### 2.10 Coverage of the changed code

Focused profile, from the M1/M2 tests only (stated as such — this is NOT a package-coverage figure):

```
$ go test ./internal/cli/ -run 'TestConfigPath|TestUpsertCodexSkillDisable' -timeout 1800s -coverprofile=<scratch>/cover.out
ok  	github.com/modu-ai/moai-adk/internal/cli	0.951s	coverage: 6.0% of statements

$ go tool cover -func=<scratch>/cover.out | grep -E 'codex_config_path|upsertCodexSkillDisable'
internal/cli/codex_config_path.go:51:	toConfigPath		100.0%
internal/cli/codex_config_path.go:61:	fromConfigPath		100.0%
internal/cli/codex_skills_disable.go:231:	upsertCodexSkillDisable	92.0%
```

## 3. Baseline-attribution

Every figure above was produced in this run, on this tree, on `darwin`:

- Card base re-derived at read time: `git merge-base origin/develop HEAD` → `9ce792637…`. The
  literal `9ce792637` in the SPEC is a dated anchor and was not used as a range edge.
- AC-CSPS-006 base measured at `b4ce67468` (pre-M1), the AFTER measurement at `dc71e8ef9` (post-M2),
  with the identical command.
- `filepath.ToSlash`/`FromSlash` identity on this host is the SPEC's M6, taken from
  `.moai/reports/t540/lab/ts.go`; it was NOT re-run in this session and is cited as a prior
  measurement, not a fresh one (§4).
- Commits: `2d66ebad4` (M1), `dc71e8ef9` (M2). Both messages contain `t540`.

## 4. Gaps — what was NOT observed

1. **AC-CSPS-001 (the Windows gate) was not attempted.** No Windows host exists here; nothing about
   Codex's resolution of a forward-slash `path` on Windows was measured. It remains open and
   landing is gated on it.
2. **M3 was not implemented.** `codex_skills_prune.go` and `doctor_codex.go` were not read for
   change and not modified. AC-CSPS-004 arms A / A' / C and AC-CSPS-005 were not attempted; only
   arm B (the pure function) is green.
3. **The M6 separator probe was not re-run in this session.** `go run .moai/reports/t540/lab/ts.go`
   is cited from the SPEC's record, not re-stamped here.
4. **No parser-touch mutant was executed** for AC-CSPS-008 — only the filter-liveness discriminator
   in §2.9. AC-CSPS-008 is M4's exit.
5. **No full-repository test run.** Scope was `./internal/cli/...` per the card's constraint; the
   full-suite verdict belongs to CI on a pushed head, and nothing was pushed.
6. **No Codex config was written.** No experiment in this run touched `~/.codex/config.toml` or set
   `CODEX_HOME`; the run is pure-function and in-memory `[]byte` merging only.
7. **`upsertCodexSkillDisable`'s remaining 8%** was not attributed to specific uncovered branches.

## 5. Residual-risk

- The whole direction still rests on AC-CSPS-001. If Codex on Windows does NOT resolve a
  forward-slash `path`, this run's publisher change is worse than the status quo there: it converts
  a visible refusal into a published entry naming a path Codex cannot resolve. **Nothing was
  pushed, and landing is correctly gated.**
- The publisher now emits slash form while M3 has not landed, so the tree is in an intermediate
  state: on a Windows host a slash-declared entry would be read back by `prune`/`doctor` without the
  inverse conversion. That is inference I2 territory, unobserved here, but it is the reason M3
  should not be deferred past a landing decision.
- `classifyCodexSkillPath` remains non-injectable (`spec.md §G` gap 4), so nothing in this run
  measured how a `C:\…` declaration classifies on Windows.
- The mixed-shape duplicate arm (AC-CSPS-007 arm 2) pins that the branch is still reached, but the
  resulting user experience — a config holding both shapes now always skips — is a behaviour change
  a Windows user could meet. It matches the criterion as written and is flagged here rather than
  assumed benign.
