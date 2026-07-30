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
   `7225a8b7a`:

   ```
   $ go test -run 'TestUpdateDryRunHelpTextMatchesBehavior' ./internal/cli/ ; echo "exit=$?"
   ok  	github.com/modu-ai/moai-adk/internal/cli	4.690s [no tests to run]
   exit=0
   ```

   An AC whose only assertion is `exit 0` would pass against a tree with no test at all, and is
   rejected.

4. **Baselines were recorded from this tree while authoring** — HEAD `7225a8b7a`, branch
   `plan/epic-update-config-audit`, in the worktree `.claude/worktrees/epic-update-config`. This
   differs from the sibling SPECs' `main` HEAD `1d4e4f7da`; see spec.md §A.6 drift 1. Each AC carries
   its observed pre-change baseline so a reviewer can distinguish a real change from a no-op.

5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions and `git stash` is
   repository-global. Falsification uses a scratch copy under `/tmp` or `t.TempDir()`, never a tree
   mutation.

6. **All fixtures use `t.TempDir()`** and touch no path outside it (NFR-UDD-001).

## §B Acceptance criteria

### M1 — `--dry-run` contract

#### AC-UDD-001 (documentation) — the help text no longer promises what the flag cannot do

```bash
grep -n 'dry-run' internal/cli/update.go | head -2
```

Expected: **either** the registered description at `internal/cli/update.go:69` no longer contains the
word `install` (option A — narrowed to archive operations), **or** it retains `install` and
AC-UDD-002 shows the clean-reinstall plan is rendered from the dry-run branch (option B). A tree
where the text says "archive and install" while `dryRunArchiveLegacySkills` remains the sole dry-run
action fails.

Baseline at HEAD `7225a8b7a` (the failing state):

```
69:	updateCmd.Flags().Bool("dry-run", false, "Show planned archive and install operations without modifying the filesystem")
293:	// --dry-run: print planned operations without mutating the filesystem
```

**Falsification**: this criterion fails if the help text still promises install operations AND the
dry-run branch still returns directly from `dryRunArchiveLegacySkills` without emitting any
clean-reinstall plan line. Deleting the flag entirely would also fail — the flag is not in scope for
removal.

#### AC-UDD-002 (code fact) — the dry-run path still performs no filesystem mutation

```bash
sed -n '290,320p' internal/cli/update.go
go test -run 'TestUpdateDryRunNoMutation' -count=1 -v ./internal/cli/
```

Expected: the `--dry-run` branch returns before the clean-reinstall path (`internal/cli/update.go:312`
comment intact — the placement rationale "so a dry run never mutates" is preserved), **and** a
`--- PASS: TestUpdateDryRunNoMutation` line from a test that runs the dry-run path against a
`t.TempDir()` fixture project and asserts the directory tree is byte-identical before and after.

Baseline: no such test exists — the `-run` selector matches nothing (§A clause 3). The current early
return is at `:293-304` with the rationale comment at `:312`.

**Falsification**: fails if option B is chosen and the plan-construction path writes anything —
including a temp file, a lock file, or a cache entry — inside the fixture project. This is the
criterion that makes option B's feasibility question answerable rather than assumed (plan.md §F.1).

#### AC-UDD-003 — the A-vs-B decision is recorded, not implicit

```bash
grep -cE 'narrow the (help )?text|extend the flag' .moai/specs/SPEC-UPDATE-DOC-DRIFT-001/progress.md
```

Expected: a count `>= 1`, with the surrounding `progress.md` §E.2 text naming the chosen option, the
rejected option, and the reason — including, where option A was chosen, whether B was probed for
feasibility first (plan.md §F.1 recommended sequencing).

Baseline: `progress.md` §E.2 is an empty placeholder at plan-phase.

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

Baseline at HEAD `7225a8b7a`: the first count is `2`; the second is `0`. The section describes all
three as governing web-console worktree automation and gives no reader status.

**Falsification**: fails if §22.8 still asserts that setting any of the three to `true` enables
behaviour, or if the toggles are simply removed from the section (which would satisfy an
absence-only grep while destroying the `[HARD]` protection on the `false` defaults — plan.md §F.2).

#### AC-UDD-005 (code fact) — the reader inventory §22.8 now states is the measured one

```bash
grep -rn 'AutoCleanup\|AutoCreate\|AutoMerge' --include='*.go' internal cmd pkg \
  | grep -v '_test.go' | grep -v main-fork
```

Expected: exactly the declaration sites (`internal/config/types.go:485-487`), the default sites
(`internal/config/defaults.go:545-547`), the single `AutoCreate` read
(`internal/cli/worktree_advisory.go:29`, `:60`), and the unrelated symbols
(`worktree/done.go` `AutoCleanupFlag`, `worktree/new.go` `ShouldAutoMerge`, `internal/github/*`
`AutoMerge`, `pkg/models/config.go:172`). No production read of `AutoCleanup` or `AutoMerge` appears.

Baseline: this is the verbatim measured state at HEAD `7225a8b7a` — the criterion asserts the
documentation now matches it.

**Falsification**: fails if a production reader for `AutoCleanup` or `AutoMerge` has appeared (which
would mean `SPEC-CONFIG-KEY-HONESTY-001` wired them and §22.8's corrected text is already stale —
exactly the reconciliation REQ-UDD-005 exists to catch).

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

Baseline at HEAD `7225a8b7a`: first count `2`, second count `0`.

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

Baseline (verbatim, HEAD `7225a8b7a`):

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
grep -n 'envkeys.go' internal/config/CLAUDE.md
grep -c 'EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"' internal/config/envkeys.go
```

Expected: the convention bullet in `internal/config/CLAUDE.md` cites a constant, and that exact
constant declaration is found in `envkeys.go` with count `>= 1`.

Baseline: the bullet cites `EnvUserName = "MOAI_USER_NAME"`, which `envkeys.go` does not declare
(AC-UDD-008 baseline). The `envkeys.go` convention itself is correct and is preserved.

#### AC-UDD-010 (documentation) — the test instruction is scoped to implemented overrides

```bash
grep -n -B2 -A2 't.Setenv' internal/config/CLAUDE.md
```

Expected: the "Tests MUST verify this priority via `t.Setenv` + fixture file combinations"
instruction is scoped to the implemented override set, so an agent following it literally writes a
test that can pass.

Baseline: the instruction sits immediately after the sentence naming `MOAI_USER_NAME` /
`MOAI_CONVERSATION_LANG`, directing verification of an unimplemented behaviour.

**Falsification**: fails if the instruction remains unscoped while the surrounding names change —
the instruction would then point at "this priority" with no antecedent, which is a different defect
rather than a fix.

### M4 — nonexistent `config.yaml`

#### AC-UDD-011 (documentation) — no instruction file names the nonexistent path

```bash
grep -n 'config/config.yaml' CLAUDE.local.md
grep -nE 'config\.yaml \(main\)|Main .config\.yaml' internal/config/CLAUDE.md
```

Expected: both produce no output (each exits 1). `CLAUDE.local.md` §9 describes the actual
`sections/*.yaml` layout, and `internal/config/CLAUDE.md` no longer asserts an aggregating main file.

Baseline at HEAD `7225a8b7a`:

```
$ grep -c 'config/config.yaml' CLAUDE.local.md
2                    # :250 (§5 release checklist), :415 (§9 "Main configuration file")
$ grep -cE 'config\.yaml \(main\)|Main .config\.yaml' internal/config/CLAUDE.md
2                    # :5 and :11
```

#### AC-UDD-012 (code fact) — the path is absent at both the local and the template location

```bash
ls .moai/config/config.yaml ; echo "local exit=$?"
ls internal/template/templates/.moai/config/config.yaml ; echo "template exit=$?"
ls .moai/config/
```

Expected: both `ls` commands print `No such file or directory` and exit non-zero, and
`ls .moai/config/` lists exactly `astgrep-rules`, `evaluator-profiles`, `sections`.

Baseline (verbatim, HEAD `7225a8b7a`) — this is the measured fact the corrected prose must agree with,
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

#### AC-UDD-014 (documentation) — §2.2 states the gate's actual default

```bash
grep -c '로더 부재' CLAUDE.local.md
grep -c '항상 컴파일 기본값 false' CLAUDE.local.md
sed -n '141p' CLAUDE.local.md | grep -cE 'advisory|기본 ON|warn_only'
```

Expected: the first two counts are `0` (the two false mechanism claims are gone) and the third is
`>= 1` (the corrected text states the gate is on by default in advisory mode). A tree where all three
are `0` fails — that is deletion, not correction.

Baseline at HEAD `7225a8b7a`: `1`, `1`, `0` respectively.

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

Baseline (verbatim, HEAD `7225a8b7a`):

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
sed -n '141p' CLAUDE.local.md | grep -cE 'git commit|IsGitCommit'
```

Expected: count `>= 1` — the corrected text names the gate's narrow trigger (`git commit` Bash
invocations), so the correction does not imply the gate fires on every tool call.

Baseline: count `0`; §2.2 gives no invocation scope at all.

**Falsification**: fails if the correction states default-on without the scope qualifier, which
trades one wrong instruction for another (plan.md §G AP-2).

#### AC-UDD-017 (code fact) — the gate is invoked only on `git commit`

```bash
grep -n -B2 -A2 'IsGitCommit' internal/hook/pre_tool.go
```

Expected: the quality gate is constructed inside an `if quality.IsGitCommit(command)` branch, so the
scope the corrected §2.2 asserts is the scope the code implements.

Baseline (HEAD `7225a8b7a`):

```
430:		if quality.IsGitCommit(command) {
431:			gate := quality.NewQualityGate(h.loadGateConfig())
```

#### AC-UDD-018 (documentation) — §2.2 records the sg-independent blocking path

```bash
sed -n '141p' CLAUDE.local.md | grep -cE 'suppression|sg-independent|sg 없이'
```

Expected: count `>= 1`, with the surrounding text stating that the suppression-policy check runs in
pure Go and can return a blocking result even when `sg` is absent, and that the sg-dependent scan
degrades gracefully.

Baseline: count `0`; §2.2 asserts the opposite — that impact requires an explicit `moai ast-grep`
invocation with `sg` installed.

**Falsification**: fails if §2.2 retains the "sg 설치 시에만" claim, or if it drops the claim without
stating what can actually block.

#### AC-UDD-019 (code fact) — the suppression check is sg-independent and can block

```bash
sed -n '41,70p' internal/hook/quality/astgrep_gate.go
```

Expected: step 1 (`checkSuppressionPairing` over `walkSourceFiles`) executes before any `sg`
invocation and contains a `return false, ...` path; step 2 (`scanner.Scan`) returns `true` on
`ErrScannerUnavailable`.

Baseline (HEAD `7225a8b7a`) — the two comment banners are the content anchors:

```
	// ── 1. Suppression policy check (sg-independent, pure-Go) ─────────────────
	...
	if len(allViolations) > 0 { ... return false, strings.TrimSpace(sb.String()) }
	// ── 2. ast-grep scan (depends on sg CLI) ─────────────────────────────────
	...
	if errors.Is(err, astgrep.ErrScannerUnavailable) { return true, astGrepReasonScannerUnavailable }
```

**Falsification**: fails if step 1 is removed or made sg-dependent, in which case §2.2's original
impact claim becomes true and AC-UDD-018's correction is wrong.

#### AC-UDD-020 — §2.2's unaffected content is preserved

```bash
sed -n '141p' CLAUDE.local.md | grep -cE 'dogfood|sgconfig\.yml|utils'
```

Expected: count `>= 2` — the dogfood-experimental rationale for not mirroring the language
subdirectory tree, the `sgconfig.yml` `utils` ruleDir issue, and the deferral of the 16-language
ruleset survive the rewrite. None of the five findings touches them.

Baseline: count `>= 2` at HEAD `7225a8b7a` (present today). This criterion guards against the M5
rewrite discarding correct content along with the incorrect.

### Cross-cutting

#### AC-UDD-021 — no template-tree file is modified

```bash
git diff --stat -- internal/template/templates/
```

Expected: **no output** — `internal/template/templates/**` is untouched (NFR-UDD-002). Both target
files are repo-local maintainer documentation and are never mirrored.

Baseline: verified at HEAD `7225a8b7a` that the template tree contains exactly one `CLAUDE.md`
(`internal/template/templates/CLAUDE.md`) and no `internal/template/templates/internal/` directory,
so neither target file ships:

```
$ find internal/template/templates -name 'CLAUDE.md'
internal/template/templates/CLAUDE.md
$ ls internal/template/templates/internal
ls: internal/template/templates/internal: No such file or directory
```

#### AC-UDD-022 — the full suite is green and the build is clean

```bash
go build ./... && go vet ./... && go test -count=1 ./...
```

Expected: exit 0 from all three.

Baseline: green at HEAD `7225a8b7a` (plan.md §C pre-flight).

#### AC-UDD-023 — no test writes outside `t.TempDir()`

```bash
go test -count=1 ./internal/cli/ ./internal/config/ && git status --porcelain
```

Expected: the tests pass and `git status --porcelain` reports no files created or modified by the
test run (NFR-UDD-001).

Baseline: clean tree before the run.

## §C Falsification procedures

Each new guard must be shown to FAIL against uncorrected content. `git stash` is prohibited
(§A clause 5); falsification uses a scratch copy under `/tmp`.

### C-1 — the documentation criteria actually fail against the current text

```bash
mkdir -p /tmp/udd-falsify
cp CLAUDE.local.md internal/config/CLAUDE.md /tmp/udd-falsify/
# run the AC-UDD-014 / AC-UDD-007 / AC-UDD-011 greps against the UNCORRECTED copies
grep -c '로더 부재' /tmp/udd-falsify/CLAUDE.local.md
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' /tmp/udd-falsify/CLAUDE.md
grep -c 'config/config.yaml' /tmp/udd-falsify/CLAUDE.local.md
```

Expected: `1`, `2`, `2` — each criterion's expected post-fix value (`0`) is contradicted, proving the
criteria are load-bearing rather than trivially satisfied. A run producing `0` here would mean the
criteria pass against the defective tree and detect nothing.

### C-2 — the deletion-vacuity hazard is actually excluded

```bash
sed '141s/.*/> **§2.2 astgrep-rules 로컬 전용 예외**: (removed)/' CLAUDE.local.md \
  > /tmp/udd-falsify/deleted.md
sed -n '141p' /tmp/udd-falsify/deleted.md | grep -cE 'advisory|기본 ON|warn_only'
sed -n '141p' /tmp/udd-falsify/deleted.md | grep -cE 'dogfood|sgconfig\.yml|utils'
```

Expected: both print `0` — so a deletion-only "fix" fails AC-UDD-014's positive half and AC-UDD-020
simultaneously. A run where AC-UDD-014 passes against this deleted copy means the criterion is an
absence-only grep and the vacuity hazard is unguarded (plan.md §G AP-8).

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

```bash
go test -run 'TestUpdateDryRunNoMutation' -count=1 -v ./internal/cli/
```

with the guard's tree-comparison assertion replaced by a bare `err == nil` check in a scratch
worktree.

Expected: **FAIL** is no longer produced against a deliberately-mutating dry-run stub — i.e. the
weakened guard passes where the real guard fails. A weakened guard that still fails would mean the
assertion under test is not the one doing the work.

## §D Definition of Done

- All of AC-UDD-001 through AC-UDD-023 produce their stated observable output.
- All four falsification procedures C-1 through C-4 produce their stated contradiction against
  uncorrected content.
- Every documentation correction cites the `file:line` or content-anchored symbol it was verified
  against (NFR-UDD-004).
- `internal/template/templates/**` is unmodified (AC-UDD-021).
- The M1 A-vs-B decision and the M2 `[HARD]`-marker decision are both recorded in `progress.md` §E.2
  with their rationale, not resolved implicitly.
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`.
