# SPEC-UPDATE-DOC-DRIFT-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command runnable as written from the repository root, and its expected
   observed output** — not merely its exit code. A criterion phrased as a property with no command is
   not an AC.

2. **Every documentation AC is paired with a code-fact AC.** This SPEC's subject is documentation
   correctness, which makes the vacuity hazard acute: a criterion phrased as "the file no longer
   contains string X" is satisfied by *deleting a sentence* without correcting anything. Each
   documentation criterion below is therefore accompanied by a sibling criterion asserting the code
   fact the corrected prose must now agree with, and each pair states its **falsification** — what
   would have to be true for the criterion to fail.

3. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore additionally
   requires the verbatim `--- PASS: <exact test name>` line. Vacuity baseline recorded at HEAD
   `d5336214e`:

   ```
   $ go test -run 'TestUpdateDryRunHelpTextMatchesBehavior' ./internal/cli/ ; echo "exit=$?"
   ok  	github.com/modu-ai/moai-adk/internal/cli	4.690s [no tests to run]
   exit=0
   ```

   An AC whose only assertion is `exit 0` would pass against a tree with no test at all, and is
   rejected.

4. **Baselines are observed, and attributed to the code baseline `d5336214e`.** Every baseline in §B
   was produced by running the stated command in the worktree `.claude/worktrees/epic-update-config`,
   branch `plan/epic-update-config-audit`. The worktree HEAD at the plan-audit revision is
   **`145e601c9`** — a descendant of `d5336214e` that changes SPEC documents only:
   `git merge-base --is-ancestor d5336214e HEAD` exits `0`,
   `git diff --name-only d5336214e HEAD | grep -cv '\.md$'` prints `0`, and
   `git diff --stat d5336214e HEAD -- '*.go'` is empty. No Go source differs between the two, so
   every `file:line` and count below is attributable to `d5336214e`. Authoring happened at
   `7225a8b7a` on the same branch; the sibling SPECs' recorded `main` HEAD `1d4e4f7da` was a stale
   divergent branch (spec.md §A.6 drift 1). Each AC carries its observed pre-change baseline so a
   reviewer can distinguish a real change from a no-op.

5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions and `git stash` is
   repository-global. Falsification uses a scratch copy under `/tmp` or `t.TempDir()`, never a tree
   mutation.

6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-UDD-001).

7. **A command that cannot observe its own expectation is not an AC.** Clause 1 requires a command;
   this clause requires the command to be *capable of seeing* what the criterion claims. It is the
   clause the sibling SPECs carry (`SPEC-CONFIG-KEY-HONESTY-001` §A clause 7;
   `SPEC-UPDATE-REINSTALL-LOOP-002` §A clause 5) and it is the one this SPEC most needed: the
   plan-audit found 13 of 23 criteria non-discriminating, of which 5 are the intentional code-fact
   pairings of clause 2 — leaving 8 genuine defects plus one unobserved baseline. Three failure
   shapes were found and each is now excluded by construction:

   - **Arithmetically unsatisfiable** — `grep -c` over a one-line input cannot print `>= 2`
     (AC-UDD-020). Counting *distinct tokens* (`grep -o | sort -u | wc -l`) is the form that can.
   - **No predicate** — a command that prints a line and an expectation phrased as a prose judgment
     about that line asserts nothing mechanically (AC-UDD-009, AC-UDD-010). Every such criterion now
     uses the AC-UDD-007 pattern: removal-count `0` **and** replacement-count `>= 1`, so neither
     deletion nor omission passes.
   - **Expectation outside the command's window** — an expected fact at `:79` cannot be observed by a
     command that prints `:41-70` (AC-UDD-019). The window is widened, or the fact gets its own
     command.

   Two further shapes are excluded: a criterion whose predicate cannot distinguish the SPEC's own
   edits from the effect it measures (AC-UDD-023 — a delta, not an absolute), and a criterion that
   becomes inert once the change is committed (AC-UDD-021 — a baseline-relative diff, not an
   unstaged-only one).

   **Known residue, recorded rather than silently carried.** AC-UDD-006's `E4` token remains loose
   (a two-character match), and its "no proposal to implement/deprecate/delete" half is unverified
   prose. That is the residue of plan-audit defect D14, deferred to the next revision along with D13,
   D15 and D17; it is named here so the clause is not read as claiming coverage it does not have.

## §B Acceptance criteria

### M1 — `--dry-run` contract

> **The A-vs-B decision is settled: option B — preserve non-mutation and make the renderer
> reachable.** The three criteria below are written against that decision, not against an open
> choice. See spec.md §A.5 and plan.md §F.1 for the measured basis: the non-mutating plan renderer
> already exists (`internal/cli/update_clean_install.go:186-198`) and is already wired
> (`internal/cli/update.go:360`); the remaining work narrows an early return, it does not build a
> renderer.

#### AC-UDD-001 (documentation) — the help text keeps its install promise, and the promise is now met

```bash
sed -n '69p' internal/cli/update.go | grep -c 'install'
sed -n '69p' internal/cli/update.go | grep -c 'archive'
```

Expected: both counts `1`. Under option B the registered description retains **both** words because
the flag now previews both halves; AC-UDD-002 is what makes the retained promise true.

Baseline at HEAD `d5336214e`: both counts are already `1` —

```
$ sed -n '69p' internal/cli/update.go
	updateCmd.Flags().Bool("dry-run", false, "Show planned archive and install operations without modifying the filesystem")
```

This criterion is therefore a **preservation guard**, not a change detector, and it is deliberately
paired: it fails if a future implementer resolves the mismatch by silently reverting to option A
(dropping `install` from the text) after option B was chosen. The change detection lives in
AC-UDD-002; the two must both hold.

**Falsification**: fails if either word is removed from the description, or if the flag is deleted —
removal is not in scope.

#### AC-UDD-002 (code fact) — the dry-run branch reaches the plan renderer, and still mutates nothing

```bash
go test -run 'TestUpdateDryRunRendersCleanReinstallPlan' -count=1 -v ./internal/cli/
go test -run 'TestUpdateDryRunNoMutation' -count=1 -v ./internal/cli/
grep -n -A2 'Placement is deliberate' internal/cli/update.go
```

Expected, all three:

1. A verbatim `--- PASS: TestUpdateDryRunRendersCleanReinstallPlan` line, from a test that runs
   `moai update --dry-run` against a `t.TempDir()` fixture carrying a v2 fingerprint and asserts the
   captured output contains the literal `[clean-reinstall] DRY-RUN — no filesystem mutations performed`
   emitted at `update_clean_install.go:187`. This is the reachability assertion: at baseline the line
   is never printed because `update.go:294` returns first.
2. A verbatim `--- PASS: TestUpdateDryRunNoMutation` line, from a test that snapshots the fixture
   tree (path set **and** per-file content hash), runs the same dry-run, and asserts the snapshot is
   unchanged. This is the criterion that proves the renderer's own prologue —
   `buildPreserveInventory` / `computeInventoryHashes` at `update_clean_install.go:179` and
   `scanDeprecatedPaths` at `:189` — writes nothing, which the plan-audit explicitly left unverified
   (its §4 gap 1).
3. The placement-rationale comment survives, anchored on its content token rather than a line number:

   ```
   312:	// Placement is deliberate: after the --binary / --dry-run early-returns (so
   313-	// a dry run never mutates) but BEFORE the version-match short-circuit
   ```

Baseline at HEAD `d5336214e`: neither test exists — both `-run` selectors match nothing, which §A
clause 3 rejects as a pass, and the `--- PASS:` requirement is what converts that into a failure. The
comment grep already prints the two lines above.

**Falsification**: assertion 2 fails if the plan-construction path writes anything inside the fixture
— a temp file, a lock file, a cache entry, a backup directory. Assertion 1 fails if the
reachability fix is omitted or reverted. Assertion 3 fails if an implementer relocates the early
return past `update.go:306-326` (`stripRetiredV2DenyEntries`, which rewrites `settings.json`) —
the option the sibling `SPEC-UPDATE-REINSTALL-LOOP-002` plan.md §E M4 records as a **confirmed
defect**. The two SPECs must not push this early return in opposite directions.

#### AC-UDD-003 — the settled decision and its basis are recorded

```bash
P=.moai/specs/SPEC-UPDATE-DOC-DRIFT-001/progress.md
sed -n '/^## §E.2 Run-phase Evidence/,/^## §E.3/p' "$P" | grep -cE 'option B|reachab'
sed -n '/^## §E.2 Run-phase Evidence/,/^## §E.3/p' "$P" | grep -c 'update_clean_install.go'
```

Expected: both counts `>= 1`, with the surrounding `progress.md` §E.2 text naming the chosen option
(B), the rejected option (A — narrow the text), the reason (the renderer already exists and is
already wired, so B's cost is a local re-ordering rather than new construction), and the run-phase
observation that the reachability change landed.

**The `sed` window is load-bearing.** The decision is *also* recorded at plan-phase in §E.1 and in
plan.md §F.1, so an unscoped whole-file `grep` over `progress.md` would print `>= 1` before
run-phase does anything — the criterion would pass against an empty §E.2. Extracting §E.2 is what
makes the count discriminating (§A clause 7).

Baseline at HEAD `145e601c9`: §E.2 is the placeholder `_<pending run-phase>_`, so both scoped counts
are `0`, while the unscoped counts are already non-zero — which is the failure this window avoids.

### M2 — §22.8 worktree toggles

#### AC-UDD-004 (documentation) — §22.8 states per-toggle reader status

```bash
sed -n '/§22.8 web worktree/,/§22.9/p' CLAUDE.local.md \
  | grep -cE 'auto_cleanup|auto_merge|auto_create|AutoCleanup|AutoMerge|AutoCreate'
sed -n '/§22.8 web worktree/,/§22.9/p' CLAUDE.local.md | grep -c 'worktree_advisory'
```

Expected: the first count `>= 2` (unchanged — the toggles are still named), and the second count
`>= 1`, with the surrounding prose stating that `auto_cleanup` and `auto_merge` have no production
reader and that `auto_create` is read only at `internal/cli/worktree_advisory.go` to select advisory
wording.

Baseline at HEAD `d5336214e`: the first count is `2`; the second is `0`. The section describes all
three as governing web-console worktree automation and gives no reader status.

**Falsification**: fails if §22.8 still asserts that setting any of the three to `true` enables
behaviour, or if the toggles are simply removed from the section (which would satisfy an
absence-only grep while destroying the `[HARD]` protection on the `false` defaults — plan.md §F.2).

#### AC-UDD-005 (code fact) — the reader inventory §22.8 now states is the measured one

The load-bearing assertion is **mechanical**, because the previous "exactly the following sites"
enumeration was incomplete and therefore already failed at baseline (it omitted
`internal/cli/worktree/errors.go`, the `defaults.go:543-544` comment lines, and several
`worktree/new.go`, `worktree_advisory.go` and `internal/github/*` lines). Asserting field *reads*
directly is both shorter and exhaustive:

```bash
grep -rn 'Worktree\.AutoCleanup\|Worktree\.AutoMerge' --include='*.go' internal cmd pkg \
  | grep -v '_test.go' | wc -l
grep -rn 'Worktree\.AutoCreate' --include='*.go' internal cmd pkg | grep -v '_test.go'
```

Expected: the first prints `0` — no production code reads `Workflow.Worktree.AutoCleanup` or
`Workflow.Worktree.AutoMerge` through the config struct, which is precisely the claim §22.8's
corrected text makes. The second prints exactly one line,
`internal/cli/worktree_advisory.go:60:	return cfg.Workflow.Worktree.AutoCreate`, the single
production read, reached from `:29` to select advisory wording.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
$ grep -rn 'Worktree\.AutoCleanup\|Worktree\.AutoMerge' --include='*.go' internal cmd pkg | grep -v '_test.go' | wc -l
0
$ grep -rn 'Worktree\.AutoCreate' --include='*.go' internal cmd pkg | grep -v '_test.go'
internal/cli/worktree_advisory.go:60:	return cfg.Workflow.Worktree.AutoCreate
```

The selector prefix `Worktree.` is what makes this discriminating: a bare `AutoMerge` match is a
homonym — `internal/github/pr_merger.go` carries an unrelated `AutoMerge` PR-merge option, and
`internal/cli/worktree/errors.go:56,60` carries `NewAutoMergeBlockedError`. Neither touches the
config key, and neither should make this criterion fail.

The broad survey grep is retained as a **context command**, not as a criterion — its output is
recorded in spec.md §A.2 as an abbreviated extract, and the abbreviation is now declared there:

```bash
grep -rn 'AutoCleanup\|AutoCreate\|AutoMerge' --include='*.go' internal cmd pkg \
  | grep -v '_test.go' | grep -v main-fork
```

**Falsification**: fails if a production reader for `AutoCleanup` or `AutoMerge` has appeared — the
first count moves off `0` — which would mean `SPEC-CONFIG-KEY-HONESTY-001` wired them and §22.8's
corrected text is already stale, exactly the reconciliation REQ-UDD-005 exists to catch. It also
fails if the `AutoCreate` read is removed, leaving §22.8 asserting a reader that no longer exists.

#### AC-UDD-006 — the correction proposes no triage and records the E4 dependency

```bash
sed -n '/§22.8 web worktree/,/§22.9/p' CLAUDE.local.md | grep -cE 'CONFIG-KEY-HONESTY|E4'
```

Expected: a count `>= 1` — §22.8 cross-references the sibling SPEC that owns the triage — and the
section contains no proposal to implement, deprecate, or delete the keys (spec.md §C).

Baseline: count `0`; the section carries no cross-reference to the owning SPEC.

### M3 — `internal/config/CLAUDE.md` env overrides

#### AC-UDD-007 (documentation) — the priority order names only implemented overrides

```bash
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/CLAUDE.md
grep -cE 'MOAI_DEVELOPMENT_MODE|MOAI_LOG_LEVEL|MOAI_LOG_FORMAT|MOAI_NO_COLOR' internal/config/CLAUDE.md
```

Expected: the first count is `0` (the two unimplemented names are gone), and the second is `>= 1`
(the implemented set is named in their place). A tree where both counts are `0` fails — that is
deletion, not correction.

Baseline at HEAD `d5336214e`: first count `2`, second count `0`.

**Falsification**: fails if the two names are removed without the implemented set replacing them, or
if the file still asserts a priority order the code does not implement.

#### AC-UDD-008 (code fact) — `applyEnvOverrides` reads exactly the documented set

```bash
grep -n -A14 'func applyEnvOverrides' internal/config/manager.go
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go
```

Expected: `applyEnvOverrides` reads exactly `EnvDevelopmentMode`, `EnvLogLevel`, `EnvLogFormat`,
`EnvNoColor` — four overrides, no more — and the `grep -c` over `envkeys.go` prints `0` (exit 1),
confirming neither name is declared.

Baseline (verbatim, HEAD `d5336214e`):

```
393:func applyEnvOverrides(cfg *Config) {
394-	if mode := os.Getenv(EnvDevelopmentMode); mode != "" { ... }
397-	if level := os.Getenv(EnvLogLevel); level != "" { ... }
400-	if format := os.Getenv(EnvLogFormat); format != "" { ... }
403-	if noColor := os.Getenv(EnvNoColor); noColor == "true" || noColor == "1" { ... }
406-}

$ grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go
0
exit=1
```

**Falsification**: fails if `envkeys.go` gains either constant — which would mean the correction
direction was wrong and `internal/config/CLAUDE.md` was right after all. This is the criterion that
makes the F4 resolution measured rather than assumed-by-recency (plan.md §G AP-3).

#### AC-UDD-009 (code fact) — the `envkeys.go` convention example names a real constant

```bash
grep -c 'EnvUserName' internal/config/CLAUDE.md
grep -c 'EnvDevelopmentMode' internal/config/CLAUDE.md
grep -c 'EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"' internal/config/envkeys.go
```

Expected: `0`, then `>= 1`, then `1`. The convention bullet no longer cites the constant `envkeys.go`
does not declare, **and** it cites one it does, **and** that citation is confirmed against the
declaration site. A tree where the first two are both `0` fails — that is deletion of the example,
not correction of it.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
$ grep -c 'EnvUserName' internal/config/CLAUDE.md
1
$ grep -c 'EnvDevelopmentMode' internal/config/CLAUDE.md
0
$ grep -c 'EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"' internal/config/envkeys.go
1
```

**Why this replaces the previous form.** The earlier criterion ran `grep -n 'envkeys.go'` (which
merely prints the bullet, asserting nothing) paired with the `envkeys.go` declaration count (already
`1` at baseline). Both halves were satisfied by the *uncorrected* file, so `internal/config/CLAUDE.md:13`
could keep citing `EnvUserName = "MOAI_USER_NAME"` forever and the criterion would still pass. The
removal/replacement pair above is the AC-UDD-007 pattern, which cannot pass without the edit
(§A clause 7).

The `envkeys.go` convention itself — constants live in `envkeys.go`; no inline `os.Getenv("MOAI_*")`
— is correct and is preserved; only the worked example changes.

**Falsification**: fails if the example is deleted rather than replaced (first count `0`, second
count `0`), or if `EnvUserName` survives anywhere in the file.

#### AC-UDD-010 (documentation) — the test instruction is scoped to implemented overrides

```bash
L=$(grep -n 't\.Setenv' internal/config/CLAUDE.md | cut -d: -f1)
echo "line=$L"
sed -n "${L}p" internal/config/CLAUDE.md | grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG'
sed -n "${L}p" internal/config/CLAUDE.md | grep -cE 'MOAI_DEVELOPMENT_MODE|MOAI_LOG_LEVEL|MOAI_LOG_FORMAT|MOAI_NO_COLOR'
```

Expected: `$L` resolves to a single line number, the second count is `0`, and the third is `>= 1` —
the bullet carrying the `t.Setenv` instruction no longer names an unimplemented variable, and does
name the implemented set the instruction's "this priority" now refers to. An agent following the
instruction literally therefore writes a test that can pass.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
line=12
1
0
```

The instruction shares line 12 with the priority-order sentence naming `MOAI_USER_NAME` /
`MOAI_CONVERSATION_LANG`, so the antecedent of "this priority" is an unimplemented behaviour.

**Why this replaces the previous form.** The earlier criterion ran `grep -n -B2 -A2 't.Setenv'` and
expected the prose judgment "the instruction is scoped to the implemented override set". A command
that prints the unmodified line can be reported as having produced its expected output, because the
expectation was never mechanical (§A clause 7). Resolving the line number and asserting the
removal/replacement counts *on that line* is what makes REQ-UDD-010 enforceable.

**Falsification**: fails if the two unimplemented names are removed from the file but the
`t.Setenv` bullet is left pointing at "this priority" with no implemented antecedent — the third
count stays `0`. It also fails if `$L` resolves to more than one line, which would mean the
instruction was duplicated and the scoping is ambiguous.

### M4 — nonexistent `config.yaml`

#### AC-UDD-011 (documentation) — no instruction file names the nonexistent path

```bash
grep -n 'config/config.yaml' CLAUDE.local.md
grep -nE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' internal/config/CLAUDE.md
```

Expected: both produce no output (each exits 1). `CLAUDE.local.md` §9 describes the actual
`sections/*.yaml` layout, and `internal/config/CLAUDE.md` no longer asserts an aggregating main file
**at either of its two sites**.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
$ grep -c 'config/config.yaml' CLAUDE.local.md
2                    # :250 (§5 release checklist), :415 (§9 "Main configuration file")
$ grep -cE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' internal/config/CLAUDE.md
2                    # :5 and :11
```

**Two corrections are folded into this criterion, and the second reverses the first.**

The originally-recorded second figure of `2` was mis-attributed: it was never observed, and the
narrow regex `config\.yaml \(main\)` genuinely prints `1`, because
`internal/config/CLAUDE.md:5` writes the claim as `` `config.yaml` (main) `` — with backticks — so a
pattern assuming a bare token adjacent to a space does not match it. Correcting the *number* to `1`
was accurate for that regex.

But correcting the number silently narrowed the *criterion*: REQ-UDD-006 and plan.md §F.4 both name
`internal/config/CLAUDE.md:5` **and** `:11`, so a criterion that can only see `:11` verifies half the
requirement, and `:5` could stay uncorrected while this AC passed. The regex is therefore widened to
tolerate the backticks (`.{0,2}` for the code-span delimiters, `.?` for the leading one), which
restores the criterion to the scope its requirement states — and the correct observed baseline for
the widened form is `2`:

```
$ grep -nE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' internal/config/CLAUDE.md
5:...the layered configuration tree under `.moai/config/` — `config.yaml` (main) plus `sections/*.yaml`...
11:- **Section-file layout (CLAUDE.local.md §9)**: ... Main `config.yaml` aggregates references. ...
```

**Falsification**: fails if either site survives. It also fails if a future author reverts to the
narrow regex, which would silently re-exempt `:5` — the widened pattern is the criterion, not an
implementation detail of it.

#### AC-UDD-012 (code fact) — the path is absent at both the local and the template location

```bash
ls .moai/config/config.yaml ; echo "local exit=$?"
ls internal/template/templates/.moai/config/config.yaml ; echo "template exit=$?"
ls .moai/config/
```

Expected: both `ls` commands print `No such file or directory` and exit non-zero, and
`ls .moai/config/` lists exactly `astgrep-rules`, `evaluator-profiles`, `sections`.

Baseline (verbatim, HEAD `d5336214e`) — this is the measured fact the corrected prose must agree with,
and the template-path half is spec.md §A.6 drift 3:

```
ls: .moai/config/config.yaml: No such file or directory
ls: internal/template/templates/.moai/config/config.yaml: No such file or directory
astgrep-rules  evaluator-profiles  sections
```

**Falsification**: fails if either file has since been created — in which case the documentation was
right and the fix direction is inverted. The criterion is what prevents "the docs were wrong" from
being assumed rather than measured.

#### AC-UDD-013 (documentation) — the §5 release checklist entry is performable

```bash
sed -n '/Files Requiring Version Sync/,/Release Process/p' CLAUDE.local.md \
  | grep -E '^\-' | while read -r _ p _; do ls "$p" >/dev/null 2>&1 || echo "MISSING: $p"; done
```

Expected: no `MISSING:` line — every path the §5 checklist names resolves to an existing file, so
every checklist line is performable by a releaser.

Baseline: the loop prints `MISSING: internal/template/templates/.moai/config/config.yaml`. This is the
release-process consequence of §A.3, distinct from the §9 documentation inaccuracy.

**Falsification**: fails if the entry is deleted without determining what does carry the shipped
version (plan.md §F.4 note) — a checklist that silently stops covering the version-bearing file is a
regression, not a fix.

### M5 — §2.2 ast-grep gate

> **Section-range anchor (replaces `sed -n '141p'` in AC-UDD-014c / 016 / 018 / 020).** The four
> criteria below previously used `sed -n '141p' CLAUDE.local.md` as their **sole** locator, which
> plan.md D6 forbids ("use `file:line` only as a locating aid, never as the sole matcher") and which
> M5 itself breaks: M5 is the largest text rewrite in the SPEC and is required to state five distinct
> facts, so the corrected §2.2 will almost certainly span more than one line — at which point a
> single-line extractor silently stops seeing most of the section. All four now extract the section
> by content anchor:
>
> ```bash
> sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md
> ```
>
> Observed at HEAD `145e601c9`: `grep -n '§2.2 astgrep-rules' CLAUDE.local.md` → `141:`, and the
> terminating anchor `### [HARD] settings.local.json Separation` is at `:143`, so the range currently
> resolves to lines 141-143 (the section body, a blank line, and the next heading) and yields the
> same counts as the retired single-line form — while remaining correct after a multi-line rewrite.
> Line numbers below are retained as locating aids only.
>
> Residual coupling, recorded not hidden: the range's terminating anchor is still bound to the
> document's structure. If §2.2 is ever moved out from under the `### [HARD] settings.local.json`
> heading, the anchor must move with it.

#### AC-UDD-014 (documentation) — §2.2 states the gate's actual default

```bash
grep -c '로더 부재' CLAUDE.local.md
grep -c '항상 컴파일 기본값 false' CLAUDE.local.md
sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md \
  | grep -cE 'advisory|기본 ON|warn_only'
```

Expected: the first two counts are `0` (the two false mechanism claims are gone) and the third is
`>= 1` (the corrected text states the gate is on by default in advisory mode). A tree where all three
are `0` fails — that is deletion, not correction.

Baseline (observed at HEAD `145e601c9`, code baseline `d5336214e`): `1`, `1`, `0` respectively.

**Falsification**: fails if the false mechanism parenthetical is removed while the section still
concludes the gate is off by default, or if the whole gate discussion is deleted from §2.2 (the
dogfood-mirroring rationale that §2.2 legitimately owns would be lost — AC-UDD-020).

#### AC-UDD-015 (code fact) — loader present, `gate.yaml` shipped, default enabled

```bash
grep -rn 'loadGateSection' internal/config/
ls internal/template/templates/.moai/config/sections/gate.yaml .moai/config/sections/gate.yaml
grep -n -A5 'AstGrepGate: AstGrepGateConfig' internal/config/defaults.go
```

Expected: all three of §2.2's false claims are contradicted by the tree — `loadGateSection` is
declared and called; both `gate.yaml` files exist; the compiled default is `Enabled: true`.

Baseline (verbatim, HEAD `d5336214e`):

```
internal/config/loader_gate.go:20:func (l *Loader) loadGateSection(dir string, cfg *Config) {
internal/config/loader.go:89:	l.loadGateSection(sectionsDir, cfg)

internal/template/templates/.moai/config/sections/gate.yaml
.moai/config/sections/gate.yaml

318:		AstGrepGate: AstGrepGateConfig{
319-			Enabled:      true,
320-			BlockOnError: false,
321-			WarnOnlyMode: true,
```

The shipped `gate.yaml` sets `ast_grep_gate.enabled: true` explicitly, so the default holds via two
independent paths.

**Falsification**: fails if `loadGateSection` is removed, `gate.yaml` is unshipped, or the default
flips to `false` — any of which would make §2.2's original text correct and this SPEC's correction
wrong. This is the pairing that prevents "the doc is stale" from being asserted rather than measured.

#### AC-UDD-016 (documentation) — §2.2 states the gate's invocation scope

```bash
sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md \
  | grep -cE 'git commit|IsGitCommit'
```

Expected: count `>= 1` — the corrected text names the gate's narrow trigger (`git commit` Bash
invocations), so the correction does not imply the gate fires on every tool call.

Baseline (observed at HEAD `145e601c9`): count `0`; §2.2 gives no invocation scope at all.

**Falsification**: fails if the correction states default-on without the scope qualifier, which
trades one wrong instruction for another (plan.md §G AP-2).

#### AC-UDD-017 (code fact) — the gate is invoked only on `git commit`

```bash
grep -n -B2 -A2 'IsGitCommit' internal/hook/pre_tool.go
```

Expected: the quality gate is constructed inside an `if quality.IsGitCommit(command)` branch, so the
scope the corrected §2.2 asserts is the scope the code implements.

Baseline (HEAD `d5336214e`):

```
430:		if quality.IsGitCommit(command) {
431:			gate := quality.NewQualityGate(h.loadGateConfig())
```

#### AC-UDD-018 (documentation) — §2.2 records the sg-independent blocking path

```bash
sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md \
  | grep -cE 'suppression|sg-independent|sg 없이'
```

Expected: count `>= 1`, with the surrounding text stating that the suppression-policy check runs in
pure Go and can return a blocking result even when `sg` is absent, and that the sg-dependent scan
degrades gracefully.

Baseline (observed at HEAD `145e601c9`): count `0`; §2.2 asserts the opposite — that impact requires
an explicit `moai ast-grep` invocation with `sg` installed.

**Falsification**: fails if §2.2 retains the "sg 설치 시에만" claim, or if it drops the claim without
stating what can actually block.

#### AC-UDD-019 (code fact) — the suppression check is sg-independent and can block

```bash
sed -n '41,85p' internal/hook/quality/astgrep_gate.go \
  | grep -nE 'Suppression policy check|ast-grep scan|ErrScannerUnavailable|return false'
grep -n -A2 'ErrScannerUnavailable' internal/hook/quality/astgrep_gate.go
```

Expected, from the first command: four relative-line hits in this order — the step-1 banner, a
`return false, ...` path, the step-2 banner, and the `ErrScannerUnavailable` branch. The ordering is
the assertion: step 1 (`checkSuppressionPairing` over `walkSourceFiles`) executes and can block
*before* any `sg` invocation. From the second: the `ErrScannerUnavailable` branch returns `true`
(graceful degradation, not a block).

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
$ sed -n '41,85p' internal/hook/quality/astgrep_gate.go \
    | grep -nE 'Suppression policy check|ast-grep scan|ErrScannerUnavailable|return false'
6:	// ── 1. Suppression policy check (sg-independent, pure-Go) ─────────────────
20:		return false, strings.TrimSpace(sb.String())
23:	// ── 2. ast-grep scan (depends on sg CLI) ─────────────────────────────────
39:		if errors.Is(err, astgrep.ErrScannerUnavailable) {
```

Relative line 39 is absolute line **79**.

**Why the window moved from `41,70` to `41,85`.** The retired command printed `:41-70`, but half of
its stated expectation — "step 2 returns `true` on `ErrScannerUnavailable`" — lives at `:79`, outside
that window. The command could not observe what the criterion claimed (§A clause 7), and the
"baseline" block recorded beneath it was not the command's output: it was an edited composite that
included the `ErrScannerUnavailable` line the command never printed, which §A clause 4 forbids. The
window is widened to cover `:79`, the baseline is replaced with the actual observed output, and the
`ErrScannerUnavailable` branch additionally gets its own command so the `return true` half is
asserted rather than inferred.

**Falsification**: fails if step 1 is removed or made sg-dependent — the step-1 banner or the
`return false` hit disappears, or their order inverts relative to the step-2 banner — in which case
§2.2's original impact claim becomes true and AC-UDD-018's correction is wrong. It also fails if the
`ErrScannerUnavailable` branch starts returning `false`, which would make the sg-dependent scan
blocking and contradict AC-UDD-018's "degrades gracefully" half.

#### AC-UDD-020 — §2.2's unaffected content is preserved

```bash
sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md \
  | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l
```

Expected: `3` — all three preserved topics are still named: the dogfood-experimental rationale for
not mirroring the language subdirectory tree, the `sgconfig.yml` `utils` ruleDir issue, and (via the
same sentence) the deferral of the 16-language ruleset. None of the five findings touches them, so
all three must survive the M5 rewrite.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
$ sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md \
    | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l
3
```

**This criterion was previously impossible to satisfy, and its recorded baseline was false.** The
retired form was `sed -n '141p' … | grep -cE 'dogfood|sgconfig\.yml|utils'` expecting `>= 2`.
`grep -c` counts *matching lines*, not matches; its input was exactly one line, so its output is
bounded above by `1` and the expectation `>= 2` is unreachable in every possible tree — including a
perfectly corrected one. The recorded baseline "count `>= 2` at HEAD `d5336214e` (present today)"
was therefore never observed; the command actually prints `1`:

```
$ sed -n '141p' CLAUDE.local.md | grep -cE 'dogfood|sgconfig\.yml|utils'
1
```

Counting **distinct matched tokens** (`grep -o | sort -u | wc -l`) is the form that expresses the
intent — three topics preserved — and it is satisfiable, discriminating, and independent of how many
lines the rewritten section occupies. This is the §A clause 7 arithmetically-unsatisfiable shape.

**Falsification**: fails at `2` or below if the M5 rewrite drops any of the three topics — which is
the deletion hazard plan.md AP-1 names. §C.2 exercises exactly this against a deleted copy.

### Cross-cutting

#### AC-UDD-021 — no template-tree file is modified

```bash
git diff --stat d5336214e..HEAD -- internal/template/templates/
git log --oneline d5336214e..HEAD -- internal/template/templates/ | wc -l
git diff --stat -- internal/template/templates/
```

Expected: the first produces **no output**, the second prints `0`, and the third produces no output.
`internal/template/templates/**` is untouched relative to the code baseline (NFR-UDD-002). Both
target files are repo-local maintainer documentation and are never mirrored.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
$ git diff --stat d5336214e..HEAD -- internal/template/templates/
$ git log --oneline d5336214e..HEAD -- internal/template/templates/ | wc -l
0
$ git diff --stat -- internal/template/templates/
```

**Why the baseline-relative form was added.** The retired criterion ran only
`git diff --stat -- internal/template/templates/`, which sees **unstaged** changes. The run-phase
workflow commits its edits, so an implementer who modified a template file and committed it would
leave that command printing nothing — the guard would fall silent at exactly the moment
NFR-UDD-002 was violated. Comparing against `d5336214e` (and counting commits that touched the
path) closes the window; the unstaged check is retained as the third command so an in-flight
modification is caught before it is committed.

The supporting fact that neither target file ships is unchanged:

```
$ find internal/template/templates -name 'CLAUDE.md'
internal/template/templates/CLAUDE.md
$ ls internal/template/templates/internal
ls: internal/template/templates/internal: No such file or directory
```

**Falsification**: fails if any template-tree file is modified, whether the change is committed or
not.

#### AC-UDD-022 — the build is clean and this SPEC's change surface is green

```bash
go build ./... && echo "build=0"
go vet ./... && echo "vet=0"
go test -count=1 ./internal/cli/ ./internal/config/ ./internal/hook/quality/
```

Expected: `build=0`, `vet=0`, and three `ok` lines.

Baseline (verbatim, observed at HEAD `145e601c9`, code baseline `d5336214e`):

```
build=0
vet=0
ok  	github.com/modu-ai/moai-adk/internal/cli	156.762s
ok  	github.com/modu-ai/moai-adk/internal/config	1.884s
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	4.174s
```

**The recorded "green" of the retired form was never observed, and the full suite is not green.**
The retired criterion was `go build ./... && go vet ./... && go test -count=1 ./...` with the
baseline "green at HEAD `d5336214e` (plan.md §C pre-flight)". Run, the full suite exits **1**:

```
FAIL	github.com/modu-ai/moai-adk/internal/hook	32.712s
--- FAIL: TestBranchGuard_Latency (1.84s)
    pre_tool_branch_guard_integration_test.go:166: iteration 4: checkBranchState took 515.368334ms, ceiling 500ms
```

Re-run alone, the same test passes (`ok … 1.554s`) — it is a **load-sensitive performance test**,
failing only under the parallel full-suite run. Since `d5336214e`→HEAD changes SPEC documents only,
it was equally flaky at `d5336214e`, so the recorded baseline could not have been observed.

The scope is narrowed to the three packages this SPEC's change surface touches, which keeps the
criterion deterministic. `internal/hook` is deliberately **excluded** from the test scope while
`go build ./...` and `go vet ./...` still cover it — this SPEC edits no Go file in that package, and
binding a documentation SPEC's Definition of Done to an unrelated timing flake would make closure
non-deterministic for reasons that have nothing to do with the SPEC.

Recorded as known, not silently dropped: `TestBranchGuard_Latency`'s load sensitivity is a
pre-existing condition outside this SPEC's scope (spec.md §C names no `internal/hook` Go change). It
is not diagnosed here — see the residual-risk note in `progress.md` §E.1.

**Falsification**: fails if any of the three packages regresses, or if `go build` / `go vet` breaks
anywhere in the module — the two whole-module commands are retained precisely so narrowing the test
scope does not narrow the build/vet scope.

#### AC-UDD-023 — the test run itself creates and modifies nothing

```bash
mkdir -p /tmp/udd-verify
git status --porcelain > /tmp/udd-verify/before.txt
go test -count=1 ./internal/cli/ ./internal/config/
git status --porcelain > /tmp/udd-verify/after.txt
diff /tmp/udd-verify/before.txt /tmp/udd-verify/after.txt
```

Expected: the tests pass and `diff` produces **no output** and exits `0` — the test run added,
removed, or modified no tracked or untracked path (NFR-UDD-001).

Baseline (observed at HEAD `145e601c9`): `diff` exits `0` with no output.

**Why a delta replaces the absolute check.** The retired criterion asserted that
`git status --porcelain` "reports no files created or modified by the test run", but the command it
ran was a bare `git status --porcelain` — an **absolute** emptiness check. This SPEC's run-phase
edits `CLAUDE.local.md`, `internal/config/CLAUDE.md`, and `progress.md` by definition, so at the
moment the criterion is evaluated the working tree is necessarily non-empty and the criterion would
fail for a reason entirely unrelated to NFR-UDD-001. The predicate could not distinguish "a test
wrote a file" from "the SPEC edited a file" — the very quantities it exists to separate. The
before/after difference measures only the test run's own effect and is unaffected by the SPEC's
edits (§A clause 7).

**Falsification**: fails if any test writes outside `t.TempDir()` — a stray fixture, a cache entry,
or a modified tracked file appears in `after.txt` and not in `before.txt`.

## §C Falsification procedures

Each new guard must be shown to FAIL against uncorrected content. `git stash` is prohibited
(§A clause 5); falsification uses a scratch copy under `/tmp`.

### C-1 — the documentation criteria actually fail against the current text

```bash
mkdir -p /tmp/udd-falsify
cp CLAUDE.local.md internal/config/CLAUDE.md /tmp/udd-falsify/
# AC-UDD-014a / AC-UDD-007 / AC-UDD-011 (local half) against the UNCORRECTED copies
grep -c '로더 부재' /tmp/udd-falsify/CLAUDE.local.md
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' /tmp/udd-falsify/CLAUDE.md
grep -c 'config/config.yaml' /tmp/udd-falsify/CLAUDE.local.md
# AC-UDD-011 (widened regex, D5) and AC-UDD-009 (removal half, D3)
grep -cE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' /tmp/udd-falsify/CLAUDE.md
grep -c 'EnvUserName' /tmp/udd-falsify/CLAUDE.md
```

Expected: `1`, `2`, `2`, `2`, `1` — each criterion's expected post-fix value (`0`) is contradicted,
proving the criteria are load-bearing rather than trivially satisfied. A run producing `0` on any
line would mean that criterion passes against the defective tree and detects nothing.

Observed at HEAD `145e601c9`: `1`, `2`, `2`, `2`, `1`.

The last two lines were added with the criteria they falsify. The fourth exercises AC-UDD-011's
**widened** regex, whose whole purpose is to see `internal/config/CLAUDE.md:5` — the narrow form
returns `1` here and would leave `:5`'s falsification unproven. The fifth exercises AC-UDD-009's
removal half, which did not previously have a falsifiable command at all.

### C-2 — the deletion-vacuity hazard is actually excluded

```bash
mkdir -p /tmp/udd-falsify
sed '141s/.*/> **§2.2 astgrep-rules 로컬 전용 예외**: (removed)/' CLAUDE.local.md \
  > /tmp/udd-falsify/deleted.md
R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
sed -n "$R" /tmp/udd-falsify/deleted.md | grep -cE 'advisory|기본 ON|warn_only'
sed -n "$R" /tmp/udd-falsify/deleted.md | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l
```

Expected: `0` and `0` — a deletion-only "fix" fails AC-UDD-014's positive half (which requires
`>= 1`) and AC-UDD-020 (which requires `3`) simultaneously. A run where AC-UDD-014 passes against
this deleted copy means the criterion is an absence-only grep and the vacuity hazard is unguarded
(plan.md §G AP-8).

Observed at HEAD `145e601c9`:

```
$ sed -n "$R" /tmp/udd-falsify/deleted.md | grep -cE 'advisory|기본 ON|warn_only'
0
$ sed -n "$R" /tmp/udd-falsify/deleted.md | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l
0
```

**The extractor was updated with the criteria it falsifies.** This procedure previously used
`sed -n '141p'`, matching the criteria's retired single-line locator. Now that AC-UDD-014c and
AC-UDD-020 anchor on the section range, the falsification must use the same range — otherwise it
would demonstrate a contradiction against commands no criterion runs. The `sed` deletion preserves
the `§2.2 astgrep-rules` start anchor and leaves the `### [HARD] settings.local.json` terminator in
place, so the range still resolves (3 lines) on the mutated copy and the counts are genuine
observations rather than extraction failures.

### C-3 — the code-fact pairing catches an inverted fix direction

```bash
# in a scratch copy only: simulate envkeys.go gaining the constant
cp internal/config/envkeys.go /tmp/udd-falsify/envkeys.go
printf '\nconst EnvUserName = "MOAI_USER_NAME"\n' >> /tmp/udd-falsify/envkeys.go
grep -c 'MOAI_USER_NAME' /tmp/udd-falsify/envkeys.go
```

Expected: `1` — AC-UDD-008's expected `0` is contradicted, so the criterion would fail and flag that
`internal/config/CLAUDE.md` was right after all. This proves the F4 resolution rests on the
measurement rather than on which file is more recent.

### C-4 — the dry-run no-mutation guard is load-bearing

Two runs against a scratch copy carrying a deliberately-mutating dry-run stub (one that writes a
single file inside the fixture before returning). `git stash` is prohibited (§A clause 5); use a
scratch `git worktree` or `go test -overlay`.

```bash
# Run 1 — the real guard (tree-snapshot comparison) against the mutating stub
go test -run 'TestUpdateDryRunNoMutation' -count=1 -v ./internal/cli/ 2>&1 | grep -E '^--- (FAIL|PASS)'

# Run 2 — the guard weakened to a bare `err == nil` check, same mutating stub
go test -run 'TestUpdateDryRunNoMutation' -count=1 -v ./internal/cli/ 2>&1 | grep -E '^--- (FAIL|PASS)'
```

Expected, stated as two separate observed lines rather than as a negated sentence:

| Run | Guard | Expected line |
|---|---|---|
| 1 | real (tree-snapshot comparison) | `--- FAIL: TestUpdateDryRunNoMutation` |
| 2 | weakened (`err == nil` only) | `--- PASS: TestUpdateDryRunNoMutation` |

The pair is what proves the load-bearing assertion is the tree comparison: the real guard catches
the mutating stub, and removing *that specific assertion* is sufficient to let it through. If run 2
also prints `--- FAIL`, some other assertion is doing the work and the tree comparison is not the
guard it is claimed to be. If run 1 prints `--- PASS`, the stub is not actually mutating and the
procedure proves nothing — check the stub first.

**Why this replaces the previous wording.** The retired form read "**FAIL** is no longer produced …
i.e. the weakened guard passes where the real guard fails" — a double negative over an unstated
second run, from which the pass/fail direction of either run could not be read off directly. It also
named no observable line, so at plan-phase the command prints `no tests to run` and exits `0`
(§A clause 3), which is neither of the two outcomes the procedure distinguishes.

**Precondition**: both runs require `TestUpdateDryRunNoMutation` to exist (AC-UDD-002 assertion 2).
Until run-phase creates it, C-4 is not executable — that is expected, not a pass.

## §D Definition of Done

- All of AC-UDD-001 through AC-UDD-023 produce their stated observable output.
- All four falsification procedures C-1 through C-4 produce their stated contradiction against
  uncorrected content — C-4 as the two-run pair, both lines observed.
- Every documentation correction cites the `file:line` or content-anchored symbol it was verified
  against (NFR-UDD-004).
- `internal/template/templates/**` is unmodified relative to `d5336214e` (AC-UDD-021).
- The M1 `--dry-run` resolution is **option B** (settled at plan-phase; see spec.md §A.5, plan.md
  §F.1). §E.2 records its execution and the observed reachability evidence — not a re-litigation of
  the choice. The M2 `[HARD]`-marker decision remains open and MUST be recorded in `progress.md`
  §E.2 with its rationale, not resolved implicitly.
- No criterion is closed on a command that cannot observe its own expectation (§A clause 7).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`.
