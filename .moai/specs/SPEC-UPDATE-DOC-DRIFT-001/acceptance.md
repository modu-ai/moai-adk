# SPEC-UPDATE-DOC-DRIFT-001 — Acceptance Criteria

Rewritten at v0.3.0 against worktree HEAD **`7f61332ef`** (branch `docs/spec-doc-drift-rewrite`,
based on `origin/main`). The v0.2.0 baseline `d5336214e` is **not an ancestor of this tree** —
`git merge-base --is-ancestor d5336214e HEAD` exits `1` — so no v0.2.0 figure was carried over.
Every baseline below was observed by running the stated command on this tree.

## §A Discipline

1. **Every AC states a command runnable as written from the repository root, and its expected
   observed output** — not merely its exit code. A criterion phrased as a property with no command is
   not an AC.

2. **Every documentation AC is paired with a code-fact AC.** This SPEC's subject is documentation
   correctness, which makes the vacuity hazard acute: a criterion phrased as "the file no longer
   contains string X" is satisfied by *deleting a sentence* without correcting anything. Each
   documentation criterion is therefore accompanied by a sibling criterion asserting the code fact
   the corrected prose must agree with, and each pair states its **falsification** — what would have
   to be true for the criterion to fail.

3. **Absence is never asserted alone.** Every removal count (`expect 0`) is paired, in the same
   criterion, with a replacement count (`expect >= 1`) over the same scope. A tree where both are `0`
   fails: that is deletion, not correction.

4. **Baselines are observed and attributed to `7f61332ef`.** Each AC carries its measured
   pre-change baseline so a reviewer can distinguish a real change from a no-op. Criteria whose
   baseline already equals the expected value are labelled **preservation guards** and never counted
   as change detectors.

5. **`git stash` is prohibited.** This checkout is shared with concurrent sessions and `git stash` is
   repository-global. Falsification uses a scratch copy under `/tmp`, never a tree mutation.

6. **A command that cannot observe its own expectation is not an AC.** Clause 1 requires a command;
   this clause requires the command to be *capable of seeing* what the criterion claims. Five failure
   shapes are excluded by construction, four inherited from the v0.2.0 plan-audit and one added here:
   arithmetically unsatisfiable counts (`grep -c` over one line cannot exceed `1`); a printed line
   paired with a prose judgment and no predicate; an expectation outside the command's window; a
   guard that goes inert once its change is committed; and — new at v0.3.0 — a predicate defeated by
   quotation (clause 8).

7. **`go test -run <pattern>` exits 0 on zero matches.** Any `-run` criterion therefore additionally
   requires the verbatim `--- PASS: <exact test name>` line. This SPEC adds no test of its own
   (§B M1 is retired), so the rule binds only AC-UDD-001's regression guard.

8. **Corrections state the current truth; they do not quote the retracted claim** *(new at v0.3.0)*.
   This clause exists because the tree demonstrated the hazard. The 2026-08-01 in-place correction to
   `CLAUDE.local.md` §2.2 retracts its predecessor **by quoting it** — "종전 이 절은 `…로더 부재 →
   … 항상 컴파일 기본값 false`라고 적었으나 두 주장 모두 현재 main에서 거짓이다". The retraction is
   correct prose and a defeated predicate: `grep -c '로더 부재' CLAUDE.local.md` still prints `1`, so
   the v0.2.0 criterion expecting `0` fails against a tree where the requirement is satisfied. An
   absence-grep cannot distinguish an assertion from its quoted retraction.

   The resolution is a constraint on the edit, not a cleverer regex — §2.2 is a single physical line,
   so every line-scoped carve-out is degenerate on it. The M5 rewrite therefore **removes the
   quotation** and states the mechanism directly, keeping only the dated marker and the
   `v3.0.1`-vs-`main` release-version distinction (which is a substantive fact, not a quotation).
   With the quotation gone, the removal counts in AC-UDD-014 are decidable again.

## §B Acceptance criteria

### M5 — `CLAUDE.local.md` §2.2 ast-grep gate (REQ-UDD-002, REQ-UDD-003)

> **Section-range anchor.** All §2.2 criteria extract the section by content anchor, never by line
> number (`:141` at v0.2.0 is `:146` today — the consolidation moved it):
>
> ```bash
> R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
> ```
>
> Observed at `7f61332ef`: the range resolves to 3 lines (the section body, a blank line, and the
> terminating heading). Residual coupling, recorded not hidden: if §2.2 is ever moved out from under
> the `### [HARD] settings.local.json` heading, this anchor must move with it.

#### AC-UDD-014 (documentation) — §2.2 states the default without quoting the retracted claim

```bash
R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
grep -c '로더 부재' CLAUDE.local.md
grep -c '항상 컴파일 기본값 false' CLAUDE.local.md
sed -n "$R" CLAUDE.local.md | grep -cE '권고 모드|advisory|WarnOnlyMode'
```

Expected: `0`, `0`, `>= 1`. The first two enforce §A clause 8 — the retracted claim is not quoted —
and the third is the replacement assertion: the section still states the gate's actual default. A
tree where all three are `0` fails.

Baseline at `7f61332ef`: `1`, `1`, `1`. The first two are the quotation inside the 2026-08-01
retraction; the third is already satisfied and is a **preservation guard** (§A clause 4) — it fails
only if the M5 rewrite deletes the default statement while removing the quotation.

**Falsification**: fails if the quotation survives (counts stay `1`), or if the rewrite removes the
quotation *and* the default statement together, leaving the gate's existence unmentioned (plan.md
AP-1).

#### AC-UDD-015 (code fact) — loader present, `gate.yaml` shipped, default enabled

```bash
grep -rn 'loadGateSection' internal/config/ | grep -c 'loader\.go\|loader_gate\.go'
ls internal/template/templates/.moai/config/sections/gate.yaml .moai/config/sections/gate.yaml
grep -n -A5 'AstGrepGate: AstGrepGateConfig' internal/config/defaults.go
```

Expected: the first prints `>= 2` (declaration + call site), `ls` succeeds on both paths, and the
default block shows `Enabled: true`.

Baseline (verbatim, `7f61332ef`):

```
$ grep -rn 'loadGateSection' internal/config/ | grep -c 'loader\.go\|loader_gate\.go'
3
$ ls internal/template/templates/.moai/config/sections/gate.yaml .moai/config/sections/gate.yaml
.moai/config/sections/gate.yaml
internal/template/templates/.moai/config/sections/gate.yaml
$ grep -n -A5 'AstGrepGate: AstGrepGateConfig' internal/config/defaults.go
438:		AstGrepGate: AstGrepGateConfig{
439-			Enabled:      true,
440-			BlockOnError: false,
441-			WarnOnlyMode: true,
```

This is a **regression guard** for the discharged half of REQ-UDD-002, retained at v0.3.0 rather than
retired with it: it fails if `loadGateSection` is removed, `gate.yaml` is unshipped, or the default
flips to `false` — any of which would make the *original* §2.2 text correct again and this SPEC's
whole M5 direction wrong.

#### AC-UDD-016 (documentation) — §2.2 states the gate's invocation scope

```bash
R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
sed -n "$R" CLAUDE.local.md | grep -cE 'git commit|IsGitCommit'
sed -n "$R" CLAUDE.local.md | grep -cE 'IsAutonomyTierCommitGateOff|autonomy|자율성 티어'
```

Expected: both `>= 1`. The corrected text names both conjuncts of the trigger condition — `git commit`
Bash invocations, and the autonomy-tier opt-out — so a reader can bound a default-on gate's blast
radius.

Baseline at `7f61332ef`: `0` and `0`. §2.2 gives no invocation scope at all.

**Falsification**: fails if the correction states default-on without the scope qualifier (plan.md
AP-2), or if it names only `git commit` and omits the autonomy-tier conjunct that spec.md §A.8 drift
2 records as new since v0.2.0.

#### AC-UDD-017 (code fact) — the gate is invoked only under the two-conjunct condition

```bash
grep -n -A2 'IsGitCommit(command)' internal/hook/pre_tool.go
```

Expected: the quality gate is constructed inside a branch conditioned on **both**
`quality.IsGitCommit(command)` and `!config.IsAutonomyTierCommitGateOff(...)`, so the scope AC-UDD-016
asserts is the scope the code implements.

Baseline (verbatim, `7f61332ef`):

```
447:		if quality.IsGitCommit(command) && !config.IsAutonomyTierCommitGateOff(config.AutonomyTier()) {
448:			gate := quality.NewQualityGate(h.loadGateConfig())
449-			passed, output := gate.Run(ctx)
```

**Falsification**: fails if either conjunct is removed, which would make AC-UDD-016's required text
an overstatement or an understatement of the real trigger.

#### AC-UDD-018 (documentation) — §2.2 records that the suppression check blocks unconditionally

```bash
R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
sed -n "$R" CLAUDE.local.md | grep -cE 'suppression|sg-independent|억제 정책'
sed -n "$R" CLAUDE.local.md | grep -cE 'WarnOnlyMode와 무관|WarnOnlyMode 와 무관|무관하게 차단|unconditional'
sed -n "$R" CLAUDE.local.md | grep -c '차단(blocking)만이'
```

Expected: `>= 1`, `>= 1`, `0`. The first two are the replacement assertions — the section names the
sg-independent suppression-policy check and states that its block does not depend on `WarnOnlyMode`.
The third removes the false claim that blocking is reachable only via a `gate.yaml` opt-in. Under
§A clause 8 the third is decidable because the rewrite does not quote what it retracts.

Baseline at `7f61332ef`: `0`, `0`, `1`.

**Falsification**: fails if §2.2 keeps "차단(blocking)만이 `gate.yaml` opt-in이다" (third count stays
`1`); fails if that clause is deleted without stating what *can* block (first two stay `0`); and
fails if the correction claims the whole gate blocks unconditionally, since the required token pairs
the unconditional block with the named suppression check rather than with the gate as a whole.

#### AC-UDD-019 (code fact) — the suppression check is sg-independent, blocks, and is not gated by `WarnOnlyMode`

```bash
sed -n '41,90p' internal/hook/quality/astgrep_gate.go \
  | grep -nE 'Suppression policy check|ast-grep scan|WarnOnlyMode|ErrScannerUnavailable|return false'
```

Expected, in this relative order: the step-1 banner, a `return false` path, the step-2 banner, the
`WarnOnlyMode` field assignment, and the `ErrScannerUnavailable` branch. The **ordering is the
assertion**: `return false` precedes every appearance of `WarnOnlyMode`, which is what makes "the
suppression block is not gated by `WarnOnlyMode`" a measured fact rather than a reading.

Baseline (verbatim, `7f61332ef`):

```
$ sed -n '41,90p' internal/hook/quality/astgrep_gate.go \
    | grep -nE 'Suppression policy check|ast-grep scan|WarnOnlyMode|ErrScannerUnavailable|return false'
6:	// ── 1. Suppression policy check (sg-independent, pure-Go) ─────────────────
20:		return false, strings.TrimSpace(sb.String())
23:	// ── 2. ast-grep scan (depends on sg CLI) ─────────────────────────────────
28:		WarnOnlyMode: cfg.WarnOnlyMode,
39:		if errors.Is(err, astgrep.ErrScannerUnavailable) {
```

Relative lines 6/20/23/28/39 are absolute 46/60/63/68/79. The window `41,90` is chosen to contain the
`ErrScannerUnavailable` branch at `:79`; the v0.2.0 window `41,70` could not see it, which §A clause 6
rejects.

**Falsification**: fails if `WarnOnlyMode` appears before the `return false` hit — i.e. the
suppression block becomes conditional — in which case §2.2's "blocking is opt-in" text would be true
and AC-UDD-018's correction wrong. Also fails if step 1 is removed or made sg-dependent.

#### AC-UDD-020 — §2.2's unaffected content is preserved

```bash
R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
sed -n "$R" CLAUDE.local.md | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l
```

Expected: `3` — all three preserved topics survive the rewrite: the dogfood-experimental rationale
for not mirroring the language subdirectory tree, the `sgconfig.yml` `utils` ruleDir issue, and the
deferral of the 16-language ruleset. None of the findings touches them.

Baseline at `7f61332ef`: `3` — a **preservation guard**.

Counting *distinct matched tokens* rather than matching lines is required: §2.2 is one physical line,
so `grep -c` is bounded above by `1` and any threshold above that is unreachable in every possible
tree (§A clause 6).

**Falsification**: fails at `2` or below if the rewrite drops any of the three topics. §C.2 exercises
this against a deleted copy.

### M2 — `.moai/docs/local-dev-settings-intent.md` §22.8 worktree toggles (REQ-UDD-004)

> **Re-anchored at v0.3.0.** `CLAUDE.local.md` has no §22: `grep -n '§22' CLAUDE.local.md` returns
> one line, `:506`, the References entry. The content lives at
> `.moai/docs/local-dev-settings-intent.md:58`, still under its `### §22.8 …` heading.
>
> ```bash
> W='/§22.8 web worktree/,/^### §22\.9/p'
> ```
>
> Observed at `7f61332ef`: the range resolves to 13 lines.

#### AC-UDD-004 (documentation) — §22.8 states the measured reader status of each toggle

```bash
W='/§22.8 web worktree/,/^### §22\.9/p'
F=.moai/docs/local-dev-settings-intent.md
sed -n "$W" "$F" | grep 'auto_cleanup' | grep -c '없다'
sed -n "$W" "$F" | grep -c 'session_worktree'
sed -n "$W" "$F" | grep -cE 'auto_cleanup|auto_merge|auto_create'
```

Expected: `0`, `>= 1`, `>= 2`. The first removes the false claim — no line may name `auto_cleanup`
and assert it has no reader. The second is the replacement assertion: the section cites the actual
reader site (`internal/cli/session_worktree.go`, and/or `session_worktree_prmerge.go`). The third is
a preservation guard: all three toggles are still named, so the correction is not a deletion.

Baseline at `7f61332ef`: `1`, `0`, `3`.

**Falsification**: fails if `auto_cleanup` is still described as unread; fails if the correction
removes the `auto_cleanup` bullet instead of correcting it (second count stays `0`); and fails if the
section is trimmed to fewer than two named toggles, which would destroy the `[HARD]` protection the
section places on the `false` defaults.

#### AC-UDD-005 (code fact) — the reader inventory §22.8 must state

```bash
grep -rn 'Worktree\.AutoCleanup' --include='*.go' internal cmd pkg | grep -v '_test.go'
grep -rn 'Worktree\.AutoMerge'   --include='*.go' internal cmd pkg | grep -v '_test.go' | wc -l
grep -rn 'Worktree\.AutoCreate'  --include='*.go' internal cmd pkg | grep -v '_test.go'
```

Expected: the first prints **two** lines (`session_worktree.go`, `session_worktree_prmerge.go`), the
second prints `0`, and the third prints exactly one line (`worktree_advisory.go`).

Baseline (verbatim, `7f61332ef`):

```
$ grep -rn 'Worktree\.AutoCleanup' --include='*.go' internal cmd pkg | grep -v '_test.go'
internal/cli/session_worktree.go:584:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
internal/cli/session_worktree_prmerge.go:122:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
$ grep -rn 'Worktree\.AutoMerge' --include='*.go' internal cmd pkg | grep -v '_test.go' | wc -l
0
$ grep -rn 'Worktree\.AutoCreate' --include='*.go' internal cmd pkg | grep -v '_test.go'
internal/cli/worktree_advisory.go:60:	return cfg.Workflow.Worktree.AutoCreate
```

The `Worktree.` selector prefix is what makes this discriminating: bare `AutoMerge` is a homonym
(`internal/github/pr_merger.go` PR-merge options, `internal/cli/worktree/errors.go`
`NewAutoMergeBlockedError`), and none of those touch the config key.

**Falsification**: fails if the `AutoCleanup` count returns to `0` — the state the v0.2.0 baseline
recorded — in which case §22.8's *original* text was right and REQ-UDD-004's inverted correction is
wrong. This is the criterion that keeps the polarity inversion measured rather than assumed. It also
fails if `AutoMerge` gains a reader, which would make the surviving half of the sentence false too.

#### AC-UDD-006 [RETIRED] — **RETIRED at v0.3.0** (REQ-UDD-005 retired)

The E4 cross-reference this criterion required is present:
`grep -c 'CONFIG-KEY-HONESTY' .moai/docs/local-dev-settings-intent.md` → `2` (`:60`, `:65`), and
`SPEC-CONFIG-KEY-HONESTY-001` is `status: completed`. The obligation is discharged; the criterion is
retained here as a record, not as an open gate.

### M3 — `internal/config/CLAUDE.md` env overrides (REQ-UDD-008, REQ-UDD-009, REQ-UDD-010)

#### AC-UDD-007 (documentation) — the priority order names only implemented overrides

```bash
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/CLAUDE.md
grep -cE 'MOAI_DEVELOPMENT_MODE|MOAI_LOG_LEVEL|MOAI_LOG_FORMAT|MOAI_NO_COLOR' internal/config/CLAUDE.md
```

Expected: `0`, then `>= 1`. A tree where both are `0` fails — that is deletion, not correction.

Baseline at `7f61332ef`: `2`, `0`.

**Falsification**: fails if the two names are removed without the implemented set replacing them, or
if the file still asserts a priority order the code does not implement.

#### AC-UDD-008 (code fact) — `applyEnvOverrides` reads exactly the documented set

```bash
grep -n -A13 'func applyEnvOverrides' internal/config/manager.go
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go
```

Expected: `applyEnvOverrides` reads exactly `EnvDevelopmentMode`, `EnvLogLevel`, `EnvLogFormat`,
`EnvNoColor` — four, no more — and the `envkeys.go` count prints `0` (exit 1).

Baseline (verbatim, `7f61332ef`):

```
398:func applyEnvOverrides(cfg *Config) {
399-	if mode := os.Getenv(EnvDevelopmentMode); mode != "" { … }
402-	if level := os.Getenv(EnvLogLevel); level != "" { … }
405-	if format := os.Getenv(EnvLogFormat); format != "" { … }
408-	if noColor := os.Getenv(EnvNoColor); noColor == "true" || noColor == "1" { … }
411-}

$ grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go
0
exit=1
```

**Falsification**: fails if `envkeys.go` gains either constant — which would mean the correction
direction is inverted and `internal/config/CLAUDE.md` was right. This is what makes the contradiction
resolved by measurement rather than by recency (plan.md AP-3).

#### AC-UDD-009 (documentation) — the `envkeys.go` convention example names a real constant

```bash
grep -c 'EnvUserName' internal/config/CLAUDE.md
grep -c 'EnvDevelopmentMode' internal/config/CLAUDE.md
grep -c 'EnvDevelopmentMode *= *"MOAI_DEVELOPMENT_MODE"' internal/config/envkeys.go
```

Expected: `0`, `>= 1`, `1`. The bullet stops citing a constant `envkeys.go` does not declare, cites
one it does, and that citation is confirmed at the declaration site. A tree where the first two are
both `0` fails — that is deletion of the example, not correction of it.

Baseline at `7f61332ef`: `1`, `0`, `1`.

**Falsification**: fails if the example is deleted rather than replaced, or if `EnvUserName` survives
anywhere in the file.

#### AC-UDD-010 (documentation) — the test instruction is scoped to implemented overrides

```bash
L=$(grep -n 't\.Setenv' internal/config/CLAUDE.md | cut -d: -f1)
echo "line=$L"
sed -n "${L}p" internal/config/CLAUDE.md | grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG'
sed -n "${L}p" internal/config/CLAUDE.md | grep -cE 'MOAI_DEVELOPMENT_MODE|MOAI_LOG_LEVEL|MOAI_LOG_FORMAT|MOAI_NO_COLOR'
```

Expected: `$L` resolves to a single line number, the second count is `0`, the third is `>= 1`. The
bullet carrying the `t.Setenv` instruction no longer names an unimplemented variable and does name
the implemented set that "this priority" now refers to, so an agent following the instruction
literally writes a test that can pass.

Baseline at `7f61332ef`: `line=12`, `1`, `0`. The instruction shares line 12 with the priority
sentence, so the antecedent of "this priority" is an unimplemented behaviour.

**Falsification**: fails if the two unimplemented names leave the file but the `t.Setenv` bullet is
left pointing at "this priority" with no implemented antecedent (third count stays `0`). Also fails
if `$L` resolves to more than one line, meaning the instruction was duplicated and the scoping is
ambiguous.

### M4 — nonexistent paths: `config.yaml` and `.agency/` (REQ-UDD-006, REQ-UDD-007)

#### AC-UDD-011 (documentation) — no instruction file names the nonexistent `config.yaml`

```bash
grep -c 'config/config.yaml' CLAUDE.local.md
grep -cE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' internal/config/CLAUDE.md
sed -n '/^## 9\. Configuration System/,/^## 10\./p' CLAUDE.local.md | grep -c 'sections/\*.yaml'
```

Expected: `0`, `0`, `>= 1`. The two removals are paired with the replacement assertion that §9 still
describes the actual `sections/*.yaml` layout.

Baseline at `7f61332ef`: `1`, `2`, `2`. The `CLAUDE.local.md` count dropped from the v0.2.0 figure of
`2` because the §5 site left with the consolidation — it is now AC-UDD-013's target, not this one's.
The third is a **preservation guard**.

The `internal/config/CLAUDE.md` regex is deliberately wide (`.{0,2}` for code-span backticks): `:5`
writes the claim as `` `config.yaml` (main) ``, which a pattern assuming a bare adjacent token cannot
see, and REQ-UDD-006 names both `:5` and `:11`. Narrowing the regex would silently re-exempt `:5`.

**Falsification**: fails if any of the three sites survives, if §9's layout description is deleted
along with the wrong sentence, or if a future author narrows the regex — the wide pattern is the
criterion, not an implementation detail of it.

#### AC-UDD-012 (code fact) — the path is absent at both the local and the template location

```bash
ls .moai/config/config.yaml ; echo "local exit=$?"
ls internal/template/templates/.moai/config/config.yaml ; echo "template exit=$?"
ls .moai/config/
```

Expected: both `ls` print `No such file or directory` and exit non-zero, and `ls .moai/config/` lists
exactly `astgrep-rules`, `evaluator-profiles`, `sections`.

Baseline (verbatim, `7f61332ef`):

```
ls: .moai/config/config.yaml: No such file or directory
local exit=1
ls: internal/template/templates/.moai/config/config.yaml: No such file or directory
template exit=1
astgrep-rules  evaluator-profiles  sections
```

**Falsification**: fails if either file has since been created — in which case the documentation was
right and the fix direction is inverted.

#### AC-UDD-025 (documentation, new at v0.3.0) — the `.agency/` template-source instruction is removed

```bash
grep -c 'internal/template/templates/.agency' CLAUDE.local.md
grep -c '\.agency' CLAUDE.local.md
grep -cE 'migrate agency|레거시|legacy' CLAUDE.local.md
```

Expected: `0`, `>= 1`, `>= 1`. The nonexistent template-source path is gone; `.agency/` is still
mentioned (it is live in the code); and the surviving mention describes it in the direction the code
implements — a legacy layout read *out of* a user project by `moai migrate agency` — rather than as a
template-managed output directory.

Baseline at `7f61332ef`: `1`, `4`, `0`.

The second count is a **deletion guard**, not a change detector: silently dropping every `.agency/`
mention would satisfy the first count while removing a fact contributors still need.

**Falsification**: fails if `:88` keeps naming `internal/template/templates/.agency/`; fails if the
`[HARD]` Template-First rule keeps instructing that new `.agency/` files be mirrored (first count
non-zero); and fails if `.agency/` is scrubbed entirely (second count `0`) or retained without the
directional correction (third count `0`).

#### AC-UDD-026 (code fact, new at v0.3.0) — `.agency/` is inbound-only and has no template source

```bash
ls -d .agency internal/template/templates/.agency
grep -rn '"\.agency"' --include='*.go' internal | grep -v '_test.go' | wc -l
grep -rn '"\.agency"' --include='*.go' internal | grep -v '_test.go'
```

Expected: both `ls` fail; the count is `>= 3`; and every printed site is a *read* of a path inside a
user project — migration source, v2 fingerprint detection, residue cleanup — with no write into a
template tree.

Baseline (verbatim, `7f61332ef`):

```
ls: .agency: No such file or directory
ls: internal/template/templates/.agency: No such file or directory
       5
internal/cli/update_residue_cleanup.go:84:	if _, agencyStatErr := os.Stat(filepath.Join(projectRoot, ".agency")); agencyStatErr == nil {
internal/cli/v2_detection.go:282:	agencyDir := filepath.Join(projectRoot, ".agency")
internal/cli/migrate_agency.go:200:	agencyDir := filepath.Join(r.projectRoot, ".agency")
internal/cli/migrate_agency.go:472:	techPrefPath := filepath.Join(r.projectRoot, ".agency", "context", "tech-preferences.md")
internal/defs/dirs.go:134:		Path:            ".agency",
```

The threshold is `>= 3` rather than `== 5` deliberately: the criterion asserts that `.agency/` is
read from user projects in several places, not that exactly five call sites exist. Pinning the exact
count would make an unrelated refactor fail a documentation criterion.

**Falsification**: fails if `internal/template/templates/.agency/` is created — in which case the
Template-First rule was right and AC-UDD-025's correction is wrong. This is the pairing that keeps
"the instruction is unperformable" measured rather than asserted.

#### AC-UDD-013 (documentation) — the version-sync checklist entry is performable

```bash
F=.moai/docs/version-management.md
grep -c 'internal/template/templates/.moai/config/config.yaml' "$F"
grep -c '\.moai/config/sections/system.yaml' "$F"
grep -cE '\{\{\.Version\}\}|render-time|렌더' "$F"
```

Expected: `0`, `>= 1`, `>= 1`. The unperformable entry is removed; the entry that *is* performable
(`.moai/config/sections/system.yaml`, which exists and carries `version: v3.1.0-rc.2`) is preserved;
and a note records that the template side needs no manual bump because it is render-time-injected.

Baseline at `7f61332ef`: `1`, `2`, `0`.

**Falsification**: fails if the entry is deleted with no note, leaving a reader to assume the
template version is unmanaged (third count `0`); fails if the `system.yaml` line is removed with it
(second count `0`); and fails if the entry is replaced by another template path rather than removed —
which is caught by AC-UDD-027, since no such path exists.

#### AC-UDD-027 (code fact, new at v0.3.0) — the template version is render-time-injected

```bash
ls internal/template/templates/.moai/config/sections/system.yaml
grep -n -A2 '^moai:' internal/template/templates/.moai/config/sections/system.yaml.tmpl
```

Expected: the `ls` fails — there is no hand-bumped template `system.yaml` — and the `.tmpl` shows
`version: "{{.Version}}"`, confirming the value is substituted when the template is rendered.

Baseline (verbatim, `7f61332ef`):

```
ls: internal/template/templates/.moai/config/sections/system.yaml: No such file or directory
4:moai:
5-  # MoAI-ADK version
6-  version: "{{.Version}}"
```

**Falsification**: fails if a literal template `system.yaml` appears, or if the `.tmpl` stops using
the `{{.Version}}` substitution — either of which would mean a template-side manual bump *is*
required and AC-UDD-013's removal is wrong. This is the measurement plan.md AP-6 demands before any
replacement path is asserted.

### M1 — `--dry-run` contract (REQ-UDD-011, REQ-UDD-012, REQ-UDD-013 — all retired)

#### AC-UDD-001 (regression guard) — the flag's promise is kept and still met

```bash
grep -n 'Show planned archive and install operations' internal/cli/update.go
grep -c 'emitDryRunReinstallPlan' internal/cli/update.go
go test -run 'TestUpdateDryRun_EmitsCleanReinstallPlan' -count=1 -v ./internal/cli/ 2>&1 | grep -E '^--- (PASS|FAIL)'
```

Expected: the help text still promises both halves; `emitDryRunReinstallPlan` is referenced `>= 2`
times (definition + the call inside the dry-run branch); and the sibling SPEC's reachability test
prints a verbatim `--- PASS: TestUpdateDryRun_EmitsCleanReinstallPlan` line. Per §A clause 7 the
`--- PASS:` line is required, because `-run` exits `0` on zero matches.

Baseline (verbatim, `7f61332ef`):

```
$ grep -n 'Show planned archive and install operations' internal/cli/update.go
81:	updateCmd.Flags().Bool("dry-run", false, "Show planned archive and install operations without modifying the filesystem")
$ grep -c 'emitDryRunReinstallPlan' internal/cli/update.go
3
$ grep -n 'func TestUpdateDryRun_EmitsCleanReinstallPlan' internal/cli/update_dry_run_reach_test.go
186:func TestUpdateDryRun_EmitsCleanReinstallPlan(t *testing.T) {
```

All three already hold — this is a **preservation guard** for a retired requirement, not a change
detector. It fails if a future change drops `install` from the help text (reverting to option A), or
removes the reachability wiring the sibling landed.

#### AC-UDD-002 [RETIRED], AC-UDD-003 [RETIRED] — **RETIRED at v0.3.0**

AC-UDD-002 [RETIRED] specified two bespoke tests (`TestUpdateDryRunRendersCleanReinstallPlan`,
`TestUpdateDryRunNoMutation`) for an implementation this SPEC no longer performs. The reachability
half is covered by AC-UDD-001's use of the sibling's existing test; the no-mutation half is the
sibling's own guarantee, recorded at `internal/cli/update.go:551-555`. AC-UDD-003 [RETIRED] required
`progress.md` §E.2 to record the option-B execution — the execution happened in
`SPEC-UPDATE-REINSTALL-LOOP-002`, whose `progress.md` carries it, so requiring a second record here
would be a claim about work this SPEC did not do.

### Cross-cutting

#### AC-UDD-024 (new at v0.3.0) — the duplicate ownership is resolved on both sides

```bash
grep -c 'SPEC-UPDATE-DOC-DRIFT-001' .moai/specs/SPEC-INTERNAL-ARCH-001/spec.md
grep -c 'REQ-ARCH-006' .moai/specs/SPEC-UPDATE-DOC-DRIFT-001/spec.md
```

Expected: both `>= 1`. This SPEC records the resolution (spec.md §A.7) and `SPEC-INTERNAL-ARCH-001`
carries the reciprocal cross-reference marking REQ-ARCH-006 superseded.

Baseline at `7f61332ef`: `0` and `>= 1` after this rewrite — the first count is the open half.

**This criterion is expected to fail until the owner of `SPEC-INTERNAL-ARCH-001` applies the
cross-reference.** That is the intended state, not a defect in the criterion: this rewrite
deliberately does not edit that SPEC (spec.md §A.7, §C), and a criterion that passed anyway would
assume away the very thing it exists to detect. It is the only AC in this SPEC that is red by
construction at plan-phase.

**Falsification**: fails if either side lacks the cross-reference. Passes only when both SPECs agree
on who owns the `internal/config/CLAUDE.md` env-var fix.

#### AC-UDD-021 — no template-tree file is modified

```bash
git diff --stat 7f61332ef..HEAD -- internal/template/templates/
git log --oneline 7f61332ef..HEAD -- internal/template/templates/ | wc -l
git diff --stat -- internal/template/templates/
```

Expected: the first produces no output, the second prints `0`, the third produces no output
(NFR-UDD-002). The baseline-relative form is required because the run-phase workflow commits its
edits — an unstaged-only check falls silent at exactly the moment the constraint is violated.

Baseline at `7f61332ef` (the baseline commit itself): all three empty / `0`.

Supporting fact, that none of the four target files is mirrored:

```
$ ls internal/template/templates/.moai/docs/
agent-lint.md  generic-patterns-guide.md
$ find internal/template/templates -name 'CLAUDE.md'
internal/template/templates/CLAUDE.md
$ ls internal/template/templates/internal
ls: internal/template/templates/internal: No such file or directory
```

**Falsification**: fails if any template-tree file is modified, committed or not.

#### AC-UDD-022 — the module still builds and vets clean

```bash
go build ./... && echo "build=0"
go vet ./... && echo "vet=0"
```

Expected: `build=0` and `vet=0`.

Baseline (observed, `7f61332ef`): `build=0`, `vet=0`.

The v0.2.0 form also ran three test packages. That scope is dropped at v0.3.0 because this SPEC now
edits **no Go file at all** — every code-touching requirement (REQ-UDD-011/012/013) is retired — so
binding a documentation SPEC's Definition of Done to any test suite would gate closure on conditions
unrelated to it. `go build` / `go vet` are retained as the cheap whole-module guard that a markdown
edit did not somehow break the tree. AC-UDD-023 is the criterion that makes "edits no Go file"
mechanical rather than asserted.

Recorded as known and out of scope, carried forward from v0.2.0 so it is not rediscovered as new:
`TestBranchGuard_Latency` in `internal/hook` is load-sensitive and fails under a parallel full-suite
run while passing alone. It is not diagnosed here.

#### AC-UDD-023 (rewritten at v0.3.0) — this SPEC modifies no Go file

```bash
git diff --name-only 7f61332ef..HEAD | grep -c '\.go$'
git diff --name-only 7f61332ef..HEAD | grep -vc '^\.moai/specs/\|^CLAUDE\.local\.md$\|^internal/config/CLAUDE\.md$\|^\.moai/docs/'
```

Expected: `0` and `0`. No Go file differs from the baseline, and every changed path lies inside this
SPEC's declared write surface: its own SPEC directory, the two always-loaded instruction files, and
the two re-anchored `.moai/docs/` targets.

Baseline at `7f61332ef` (the baseline commit itself): `0` and `0`.

**Why this replaces the v0.2.0 form.** That criterion snapshotted `git status --porcelain` before and
after a test run to prove tests wrote nothing outside `t.TempDir()`. With no tests added, it measures
nothing. The scope constraint — a documentation SPEC that touches only documentation — is the
property actually worth asserting, and a baseline-relative name-diff asserts it without depending on
whether the edits are committed yet.

**Falsification**: fails if any Go file is edited, or if the SPEC's edits reach a path outside the
five declared prefixes.

## §C Falsification procedures

Each guard must be shown to FAIL against uncorrected content. `git stash` is prohibited (§A clause 5);
falsification uses a scratch copy under `/tmp`.

### C-1 — the documentation criteria actually fail against the current text

```bash
mkdir -p /tmp/udd-falsify
cp CLAUDE.local.md /tmp/udd-falsify/CLAUDE.local.md
cp internal/config/CLAUDE.md /tmp/udd-falsify/config-CLAUDE.md
cp .moai/docs/version-management.md .moai/docs/local-dev-settings-intent.md /tmp/udd-falsify/
grep -c '로더 부재' /tmp/udd-falsify/CLAUDE.local.md                                      # AC-UDD-014
grep -c 'config/config.yaml' /tmp/udd-falsify/CLAUDE.local.md                             # AC-UDD-011
grep -c 'internal/template/templates/.agency' /tmp/udd-falsify/CLAUDE.local.md            # AC-UDD-025
grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' /tmp/udd-falsify/config-CLAUDE.md        # AC-UDD-007
grep -c 'EnvUserName' /tmp/udd-falsify/config-CLAUDE.md                                   # AC-UDD-009
grep -cE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' /tmp/udd-falsify/config-CLAUDE.md  # AC-UDD-011
grep -c 'internal/template/templates/.moai/config/config.yaml' /tmp/udd-falsify/version-management.md  # AC-UDD-013
grep 'auto_cleanup' /tmp/udd-falsify/local-dev-settings-intent.md | grep -c '없다'        # AC-UDD-004
```

Expected: `1`, `1`, `1`, `2`, `1`, `2`, `1`, `1` — every criterion's expected post-fix value (`0`) is
contradicted, proving each is load-bearing rather than trivially satisfied. Any line printing `0`
would mean that criterion passes against the defective tree and detects nothing.

Observed at `7f61332ef`: `1`, `1`, `1`, `2`, `1`, `2`, `1`, `1`.

The `internal/config/CLAUDE.md` copy is renamed to `config-CLAUDE.md` in the scratch directory so it
cannot collide with `CLAUDE.local.md`'s basename or with a stray `CLAUDE.md`; the v0.2.0 procedure
copied both into one directory and then grepped `/tmp/udd-falsify/CLAUDE.md`, whose provenance was
ambiguous.

### C-2 — the deletion-vacuity hazard is actually excluded

```bash
mkdir -p /tmp/udd-falsify
L=$(grep -n '§2.2 astgrep-rules' CLAUDE.local.md | cut -d: -f1)
sed "${L}s|.*|> **§2.2 astgrep-rules 로컬 전용 예외**: (removed)|" CLAUDE.local.md > /tmp/udd-falsify/deleted.md
R='/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p'
sed -n "$R" /tmp/udd-falsify/deleted.md | grep -cE '권고 모드|advisory|WarnOnlyMode'
sed -n "$R" /tmp/udd-falsify/deleted.md | grep -cE 'suppression|sg-independent|억제 정책'
sed -n "$R" /tmp/udd-falsify/deleted.md | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l
```

Expected: `0`, `0`, `0` — a deletion-only "fix" simultaneously fails AC-UDD-014's positive half
(`>= 1`), AC-UDD-018's positive half (`>= 1`), and AC-UDD-020 (`3`). If AC-UDD-014 passed against
this deleted copy, it would be an absence-only grep and the vacuity hazard would be unguarded
(plan.md AP-8).

The line number is resolved by `grep` rather than hardcoded, so the procedure survives the next time
the file is consolidated — the hardcoded `141` in the v0.2.0 form now points at an unrelated line.

### C-3 — the code-fact pairing catches an inverted fix direction

```bash
cp internal/config/envkeys.go /tmp/udd-falsify/envkeys.go
printf '\nconst EnvUserName = "MOAI_USER_NAME"\n' >> /tmp/udd-falsify/envkeys.go
grep -c 'MOAI_USER_NAME' /tmp/udd-falsify/envkeys.go
```

Expected: `1` — AC-UDD-008's expected `0` is contradicted, so the criterion would fail and flag that
`internal/config/CLAUDE.md` was right after all. This proves the contradiction is resolved by
measurement, not by which file is more recent (plan.md AP-3).

### C-4 (new at v0.3.0) — the polarity inversion is measured, not assumed

```bash
grep -rn 'Worktree\.AutoCleanup' --include='*.go' internal cmd pkg | grep -v '_test.go' | wc -l
sed -n '578,590p' internal/cli/session_worktree.go
```

Expected: the count is `2`, and the printed context shows the read gating disposal — an early
`return` when the toggle is `false` — rather than selecting between two advisory strings.

Observed at `7f61332ef`:

```
2
	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
		// REQ-SW-008: default-manual — auto_cleanup is OFF (the distributed
		// default per CLAUDE.local.md §22.8). The worktree PERSISTS after exit;
		// the user disposes explicitly via `moai worktree done` / `remove`.
		return
	}
```

This is the procedure that justifies inverting REQ-UDD-004 rather than executing it as v0.2.0 wrote
it. Had the count been `0`, the v0.2.0 correction would have been right and the inversion wrong. It
is run *before* the M2 edit, not after.

## §D Definition of Done

- AC-UDD-004, 005, 007, 008, 009, 010, 011, 012, 013, 014, 015, 016, 017, 018, 019, 020, 021, 022,
  023, 025, 026, 027 produce their stated observable output.
- AC-UDD-001 holds as a preservation guard (the retired M1 has not regressed).
- AC-UDD-024 is **expected red at plan-phase** and closes only when the owner of
  `SPEC-INTERNAL-ARCH-001` applies the reciprocal cross-reference. Closing this SPEC with AC-UDD-024
  red is permitted, provided the open half is named in `progress.md` §E.1 rather than silently
  passed.
- AC-UDD-002 [RETIRED], AC-UDD-003 [RETIRED], AC-UDD-006 [RETIRED] are retired; their retirement evidence is recorded above and
  is not re-litigated.
- Falsification procedures C-1 through C-4 produce their stated contradictions.
- Every documentation correction cites the `file:line` or content-anchored symbol it was verified
  against (NFR-UDD-004).
- No Go file is modified (AC-UDD-023) and `internal/template/templates/**` is unmodified relative to
  `7f61332ef` (AC-UDD-021).
- No criterion is closed on a command that cannot observe its own expectation (§A clause 6), and no
  correction retracts a claim by quoting it (§A clause 8).
- `progress.md` §E.2 cites the observed command output for every claim, per
  `.claude/rules/moai/core/verification-claim-integrity.md`.
