---
id: SPEC-UPDATE-DOC-DRIFT-001
title: "always-loaded instruction drift: maintainer documentation that asserts a mechanism the code contradicts is not a stale comment — it is an input that misdirects every agent session"
version: "0.2.0"
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: P2
phase: "v3.0.2"
module: "docs"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "documentation, drift, claude-local, instruction-correctness, always-loaded, dry-run, flag-contract, config, update"
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-KEY-HONESTY-001, SPEC-UPDATE-CI-GUARD-001]
depends_on: []
---

# SPEC-UPDATE-DOC-DRIFT-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Epic SPEC 6 of 6 — the closing SPEC of the four-lens audit of `moai update` / `.moai/config`. Findings F1-F5 each re-verified while authoring; F1 found false in three independent ways rather than the two supplied; F3 found false at the template path as well as the local path; F1's user-impact claim independently disproved. Three drifts recorded (§A.6). |
| 0.2.0 | 2026-07-31 | Plan-audit revision (iteration 1 verdict **FAIL, 0.65** against the Tier M threshold 0.80; Must-Pass 7/7 PASS — the failure was entirely in verification design, Testability 0.50). D1-D11 resolved, D12 folded into D1's edit, D16 resolved; D13-D15 and D17 deferred. **§A.5's option-B cost framing was found inverted and is rewritten against measurement**: the non-mutating clean-reinstall renderer already exists (`update_clean_install.go:186-198`) and is already wired (`update.go:360`), so option B is a local re-ordering, not new construction — the `--dry-run` decision is settled as **option B**, preserving non-mutation. §A.5's `:305-325` citation corrected to `:306-326` (deny-rule migration) / `:328` (clean-reinstall detection). §A.2's output block declared abbreviated and its unrelated-symbol enumeration completed. `acceptance.md` gained §A clause 7 (anti-vacuity); ten criteria rewritten under it. |

## §A Problem / Motivation

The five preceding SPECs of this Epic are about code and config that lie to the **user**. This one is
about documentation that lies to the **maintainer** — and therefore to every agent session, because
`CLAUDE.local.md` and `internal/config/CLAUDE.md` are loaded as project instructions on every turn.

That loading property is the whole reason this is a SPEC rather than a cleanup chore. A false
statement inside an always-loaded instruction file is not a stale comment sitting unread next to the
code it describes. It is an **input**: it enters the model's context before any work begins, it is
read as authoritative project doctrine, and it actively steers work in the wrong direction. A
maintainer who reads "the gate is off by default" does not think to check whether the gate fires. An
agent that reads the same sentence has no independent reason to check either — the instruction file
is precisely the surface it is told to trust.

Two of the five findings below are not merely wrong but wrong in a way that would suppress
investigation: §A.1 states a mechanism that would make a whole subsystem inert (so nobody looks at
it), and §A.4 is a direct contradiction between two files that are *both* always-loaded, so whichever
one a reader happens to weight wins silently.

Sibling SPECs own adjacent halves and are not restated here: `SPEC-CONFIG-KEY-HONESTY-001` (E4) owns
the code-and-config layer of unread config keys; `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) owns whether
the clean-reinstall loop should trigger; `SPEC-UPDATE-CI-GUARD-001` (E5) owns CI guard pattern design.
**This SPEC owns exactly one thing: making the documentation say what the code does.**

Every finding below was re-verified while authoring. Where the verification contradicted the
finding as supplied, the contradiction is recorded rather than silently corrected (§A.6).

### A.1 `CLAUDE.local.md` §2.2 is wrong three times over about the ast-grep gate (F1)

`CLAUDE.local.md:141` (§2.2) states that the ast-grep pre-tool gate is off by default, giving as the
mechanism: `ast-grep pre-tool gate는 기본 OFF(`gate.yaml`/`gate` 로더 부재 → `AstGrepGate.Enabled`
항상 컴파일 기본값 false)`, and concludes `실제 영향은 명시적 `moai ast-grep` CLI 경로 + `sg` 설치
시에만 발생`.

Four claims are packed into that sentence. **All four are false at HEAD.**

**(a) The loader is not absent.** `internal/config/loader_gate.go:20` declares
`func (l *Loader) loadGateSection(dir string, cfg *Config)`, and `internal/config/loader.go:89`
calls it:

```
$ grep -rn 'loadGateSection' internal/config/
internal/config/loader_gate.go:14:// loadGateSection loads the gate configuration section from gate.yaml.
internal/config/loader_gate.go:20:func (l *Loader) loadGateSection(dir string, cfg *Config) {
internal/config/loader.go:89:	l.loadGateSection(sectionsDir, cfg)
```

The file's own header comment (`loader_gate.go:5-12`) states verbatim that it was added *because*
"no loader path read gate.yaml, so the ast-grep pre-tool gate (AstGrepGate.Enabled) had no path to
true" — §2.2 is a description of the state that existed **before this file landed**, preserved
unamended after the condition it describes was removed.

**(b) `gate.yaml` is not absent — it is shipped.** Both the template source and the local project
carry it:

```
$ ls internal/template/templates/.moai/config/sections/gate.yaml .moai/config/sections/gate.yaml
internal/template/templates/.moai/config/sections/gate.yaml
.moai/config/sections/gate.yaml
```

and its shipped content sets the gate on explicitly (`gate.yaml:13-25`): `gate.enabled: true`,
`ast_grep_gate.enabled: true`, `block_on_error: false`, `warn_only_mode: true`.

**(c) The compiled default is `true`, not `false`.** `internal/config/defaults.go:316-322`:

```go
// The ast-grep sub-gate is ON by default in advisory mode (findings
// reported, commits never blocked); blocking is opt-in via gate.yaml.
AstGrepGate: AstGrepGateConfig{
	Enabled:      true,
	BlockOnError: false,
	WarnOnlyMode: true,
},
```

So the gate is on in **advisory mode** by default — the opposite of §2.2's "always compiled default
false".

**(d) The impact is not confined to an explicit CLI path with `sg` installed.** This claim required
its own check rather than inheritance from (a)-(c), and it does not survive. `RunAstGrepGateV2`
(`internal/hook/quality/astgrep_gate.go:41`) runs **two** steps, and only the second depends on `sg`:

```go
// ── 1. Suppression policy check (sg-independent, pure-Go) ─────────────────
sourceFiles := walkSourceFiles(projectDir)
for _, fp := range sourceFiles {
	allViolations = append(allViolations, checkSuppressionPairing(fp)...)
}
if len(allViolations) > 0 {
	...
	return false, strings.TrimSpace(sb.String())      // ← blocks, with sg absent
}
// ── 2. ast-grep scan (depends on sg CLI) ─────────────────────────────────
```

Step 1 walks every source file in pure Go and can return `false` — a **block** — with `sg` nowhere on
the system. Step 2 degrades gracefully when `sg` is missing (`ErrScannerUnavailable` → pass), and
`WarnOnlyMode: true` prevents step 2 from blocking even when it does find something. So the one path
that can actually block a commit is the one §2.2 asserts cannot run at all.

**The precise scope, stated so this SPEC does not overclaim in the other direction.** The gate is not
invoked on every tool call: `internal/hook/pre_tool.go:430-431` gates it on
`quality.IsGitCommit(command)`, so it fires on `git commit` Bash invocations only. That is a
meaningful narrowing of blast radius — and it is also not what §2.2 says. The correct statement is
"on by default in advisory mode, evaluated at `git commit`, whose sg-independent suppression-policy
check can block"; §2.2 says "inert".

This is the highest-severity finding in the SPEC. A maintainer reading §2.2 concludes the subsystem
cannot fire, and therefore has no reason to investigate a commit that is unexpectedly blocked by it.

### A.2 `CLAUDE.local.md` §22.8 presents three toggles as policy; two have no reader (F2)

§22.8 presents `AutoCleanup`, `AutoCreate`, and `AutoMerge` as an intentional default-OFF **policy** —
"웹 콘솔의 worktree 자동화는 사용자의 **명시적 opt-in** ... 있을 때만 동작한다" — and marks it
`[HARD] ... 의도된 정책. 감사/동기화 시 "결함"으로 되돌리지 말 것`. The prose implies all three gate
behaviour, and that setting one to `true` turns something on.

Measured at HEAD. The survey grep returns 33 lines; the block below is an **abbreviated extract**,
not verbatim output — it collapses contiguous line ranges and omits the unrelated-symbol hits
enumerated in the paragraph that follows:

```
$ grep -rn 'AutoCleanup\|AutoCreate\|AutoMerge' --include='*.go' internal cmd pkg \
    | grep -v '_test.go' | grep -v main-fork          # abbreviated: 33 lines total
internal/config/types.go:485-487        (declarations)
internal/config/defaults.go:545-547     (defaults, all three false)
internal/cli/worktree_advisory.go:29    autoCreate := readWorktreeAutoCreate(projectRoot)
internal/cli/worktree_advisory.go:60    return cfg.Workflow.Worktree.AutoCreate
```

The decisive assertion is not this survey but the **selector-scoped** count, which is exhaustive and
is what AC-UDD-005 asserts:

```
$ grep -rn 'Worktree\.AutoCleanup\|Worktree\.AutoMerge' --include='*.go' internal cmd pkg | grep -v '_test.go' | wc -l
0
$ grep -rn 'Worktree\.AutoCreate' --include='*.go' internal cmd pkg | grep -v '_test.go'
internal/cli/worktree_advisory.go:60:	return cfg.Workflow.Worktree.AutoCreate
```

Only `AutoCreate` is read in production, and per `SPEC-CONFIG-KEY-HONESTY-001` §A.6 that read
selects between two advisory sentences — it does not gate worktree creation. `AutoCleanup` and
`AutoMerge` have zero production readers. The remaining survey hits are unrelated symbols and
comments, all of which the `Worktree.`-prefixed selector correctly excludes as homonyms:
`worktree/done.go:29,31` `AutoCleanupFlag` (a CLI flag name), `worktree/new.go:439,457,458,459`
`ShouldAutoMerge` (a `--no-merge` flag helper), `worktree/errors.go:56,60`
`NewAutoMergeBlockedError` (a worktree error constructor), `worktree_advisory.go:48,51` (the
`readWorktreeAutoCreate` doc comment and signature), `defaults.go:543-544` (the mutation-rationale
comment), the `internal/github/` `AutoMerge` fields at `pr_merger.go:12,13,54,55,113,116,117,159,164,230`
and `errors.go:34,35` (PR-merge options), and `pkg/models/config.go:172` (a separate model struct).

Setting `auto_cleanup: true` or `auto_merge: true` in `workflow.yaml` therefore changes nothing —
while §22.8 tells the reader it is the opt-in switch.

**Boundary — the sharpest in this SPEC.** `SPEC-CONFIG-KEY-HONESTY-001` §A.6 already owns the
**code-and-config** layer of this finding: the unread keys themselves, and the triage of whether to
implement, deprecate, or mark them. This SPEC owns **only the documentation defect** — that §22.8
asserts an enforcement that does not exist — and the obligation to reconcile the prose with whatever
E4's triage decides. It does not re-specify the key triage, propose wiring, or take a position on
which triage class the keys land in. See §C.

### A.3 `CLAUDE.local.md` §5 and §9 reference a file that exists at neither path (F3)

Both sections point at `config.yaml`:

```
$ grep -n 'config/config.yaml' CLAUDE.local.md
250:- internal/template/templates/.moai/config/config.yaml (moai.version)
415:**Project config** (`.moai/config/config.yaml`):
```

Line 415 (§9) calls it the "Main configuration file". Line 250 (§5, *Files Requiring Version Sync*)
lists the template-tree path among the files a release must bump.

Neither path resolves:

```
$ ls .moai/config/config.yaml
ls: .moai/config/config.yaml: No such file or directory
$ ls internal/template/templates/.moai/config/config.yaml
ls: internal/template/templates/.moai/config/config.yaml: No such file or directory
$ ls .moai/config/
astgrep-rules  evaluator-profiles  sections
```

**This is worse than a stale local-path reference, and the template path was checked rather than
assumed symmetric.** §5 is a release checklist. It instructs the releaser to bump `moai.version` in a
file that does not exist, so the step is unperformable: a releaser either silently skips it (and the
checklist has a dead line that erodes trust in the rest) or creates the file to satisfy it (and
introduces a config file nothing loads). The §9 claim compounds it by naming a nonexistent file as
the *main* config, when the actual layout is `sections/*.yaml` only.

`internal/config/CLAUDE.md:5` repeats the same claim ("`config.yaml` (main) plus `sections/*.yaml`")
and `:11` says "Main `config.yaml` aggregates references" — so the nonexistent file is asserted on
two always-loaded surfaces.

### A.4 `internal/config/CLAUDE.md` and `CLAUDE.local.md` §9 directly contradict each other (F4)

`internal/config/CLAUDE.md:12` states the configuration priority order:

> **Configuration priority order**: (1) Environment variables (`MOAI_USER_NAME`,
> `MOAI_CONVERSATION_LANG`, ...) override file values. ... Tests MUST verify this priority via
> `t.Setenv` + fixture file combinations.

and `:13` instructs that these names live as constants in `envkeys.go`, giving
`EnvUserName = "MOAI_USER_NAME"` as the worked example.

`CLAUDE.local.md` §9 states the opposite in an explicit NOTE:

> `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` are **NOT currently implemented** — no code in
> `internal/`/`pkg/`/`cmd/` reads them.

**Both files are loaded as project instructions.** The contradiction was resolved against the code
rather than by assuming the more recent file wins:

```
$ grep -n 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go
(no output)
```

`envkeys.go` declares `EnvConfigDir`, `EnvDevelopmentMode`, `EnvLogLevel`, `EnvLogFormat`,
`EnvNoColor`, and many others — but no `EnvUserName` and no `EnvConversationLang`. And
`applyEnvOverrides` (`internal/config/manager.go:393-406`) reads exactly four:

```go
func applyEnvOverrides(cfg *Config) {
	if mode := os.Getenv(EnvDevelopmentMode); mode != "" { ... }
	if level := os.Getenv(EnvLogLevel); level != "" { ... }
	if format := os.Getenv(EnvLogFormat); format != "" { ... }
	if noColor := os.Getenv(EnvNoColor); noColor == "true" || noColor == "1" { ... }
}
```

**`CLAUDE.local.md` §9 is correct; `internal/config/CLAUDE.md:12-13` is wrong** — and the example it
chose to illustrate the `envkeys.go` convention is one of the two names the file does not contain.

**The second-order defect.** `:12` does not merely misstate a fact; it issues an instruction:
"Tests MUST verify this priority via `t.Setenv` + fixture file combinations." For an unimplemented
variable that is an instruction to write a test that cannot pass. An agent following the file
literally would either write a failing test or, worse, implement the env override to make the test
pass — an unrequested behaviour change originating in a documentation error.

**Who a fix reaches.** `internal/config/CLAUDE.md` is git-tracked but **not** shipped: the template
tree contains exactly one `CLAUDE.md` (`internal/template/templates/CLAUDE.md`), and there is no
`internal/template/templates/internal/` directory. It is therefore repo-local maintainer
documentation, like `CLAUDE.local.md` — its audience is moai-adk-go contributors and every agent
session in this repository, not downstream users. Both fixes are repo-local; neither ships.

### A.5 `--dry-run` cannot preview the operation most worth previewing (F5)

The flag's registered help text (`internal/cli/update.go:69`):

```go
updateCmd.Flags().Bool("dry-run", false, "Show planned archive and install operations without modifying the filesystem")
```

The handler (`internal/cli/update.go:293-304` — the comment at `:293`, the branch at `:294-304`) is
an early return placed *before* every mutating step that follows it:

```go
// --dry-run: print planned operations without mutating the filesystem
if getBoolFlag(cmd, "dry-run") {
	cwd, err := os.Getwd()
	...
	emitWorktreeAdvisory(out, cwd)
	return dryRunArchiveLegacySkills(cwd, out)
}
```

`dryRunArchiveLegacySkills` (`internal/cli/update_archive.go:339-353`) iterates `legacySkillIDs`,
prints one `[dry-run] archive: <id>` line per legacy skill present, and returns. It previews the
**archive** half only.

**What actually sits after the early return, measured.** The early return closes at `:304`. What
follows is *not* the clean-reinstall path:

```
$ sed -n '306,313p;328,329p;359,360p' internal/cli/update.go
306:	// Retired-deny-rule migration on the v3 path (issue #1101 follow-up). The
...
312:	// Placement is deliberate: after the --binary / --dry-run early-returns (so
313:	// a dry run never mutates) but BEFORE the version-match short-circuit
328:	// SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-002: detect v2 fingerprint
329:	// and short-circuit to the clean-reinstall code path when the project is
359:			result, runErr := runCleanReinstall(ctx, cwd, CleanReinstallOptions{
360:				DryRun: getBoolFlag(cmd, "dry-run"),
```

`:306-326` is the **retired-deny-rule migration**, which mutates — it calls
`stripRetiredV2DenyEntries(cwd, out)`, rewriting `settings.json`. The clean-reinstall *detection*
begins at `:328` and calls `runCleanReinstall` at `:359`. The comment at `:312` records that the
early-return placement is deliberate: a dry run must not reach that mutating migration. That
reasoning is sound and this SPEC preserves it.

The defect is the **contract**: the help text promises "archive **and install** operations", and the
install half is never previewed. `moai update --dry-run` on a v2 project prints legacy-skill
archiving and says nothing about the clean reinstall that is about to replace the user's tree.

**This is the one finding with both a documentation face and a behavioural face, and the scope is
narrow deliberately.** This SPEC owns the **contract mismatch** — the help text and the behaviour
must agree. It does not own whether clean-reinstall should trigger at all; that is
`SPEC-UPDATE-REINSTALL-LOOP-002` (§C).

#### The install half is unreachable by *placement*, not by construction — and the renderer already exists

An earlier draft of this section stated that "the install half is unreachable by construction" and
that extending the flag "requires constructing the plan without executing it — the plan-construction
path would have to be separable from the mutation path, which is real work and may not belong in
this Epic at all". **Both statements are false, and the second is inverted.** The probe this SPEC's
own plan recommended was run, and it found the separation already built:

```
$ sed -n '50,53p;184,198p' internal/cli/update_clean_install.go
 50:type CleanReinstallOptions struct {
 51:	// DryRun: when true, emit planned actions but make no filesystem mutations.
 52:	// REQ-VVCR-028.
 53:	DryRun bool
...
184:	// Dry-run early return — emit planned actions and stop before any
185:	// filesystem mutation.
186:	if opts.DryRun {
187:		… "[clean-reinstall] DRY-RUN — no filesystem mutations performed"
188:		… "Would back up %d files into .moai/backups/v2-to-v3-<stamp>/"
190:		… "Would remove %d deprecated paths"
194:		… "Would auto-invoke `moai migrate agency` for .agency/ contents"
197:		return result, nil
198:	}
```

A non-mutating clean-reinstall plan renderer exists at `update_clean_install.go:186-198`
(REQ-VVCR-028), and `update.go:360` already wires `DryRun: getBoolFlag(cmd, "dry-run")` into
`runCleanReinstall`. The renderer is unreachable for one reason only: the CLI's own `--dry-run`
early return at `:294` fires before the v2-detection block at `:328` is reached. Nothing needs to be
built; the plumbing has to reach code that is already written and already wired.

The cost comparison the earlier draft rested on was therefore backwards — it charged the option that
requires a local re-ordering with the cost of an unbuilt subsystem.

#### Decision — option B, settled

| Option | Change | Actual cost | What is lost |
|---|---|---|---|
| A — narrow the text | Rewrite the description to "Show planned archive operations…". | One line. | The highest-consequence operation stays unpreviewable — a user running `--dry-run` before a v2 update still learns nothing about the reinstall about to replace their tree. |
| **B — make the renderer reachable** *(chosen)* | Hoist v2 detection above the `--dry-run` branch and have that branch render the resulting plan and return. | A local re-ordering. The renderer (`update_clean_install.go:186-198`) and its wiring (`update.go:360`) already exist. | Nothing. |

**Option B is the resolution, and it preserves non-mutation rather than trading it away.** The
implementation constraint is explicit: the `--dry-run` early return is **not** relocated past
`:306-326`. Moving it there would put `stripRetiredV2DenyEntries` — a `settings.json` rewrite — on
the dry-run path, contradicting the `:312` rationale that a dry run never mutates. Detection is
hoisted *above* the branch instead; the branch itself keeps returning before any mutating call.

This matches the sibling `SPEC-UPDATE-REINSTALL-LOOP-002` plan.md §E M4, which reached the same
resolution independently and records the relocate-the-return variant as a **confirmed defect**. The
two SPECs must not push this early return in opposite directions.

Residual verification owed to run-phase, not assumed here: that the renderer's prologue —
`buildPreserveInventory` / `computeInventoryHashes` (`:179`) and `scanDeprecatedPaths` (`:189`) —
performs no filesystem write. The code path was read, not executed. AC-UDD-002 assertion 2 is the
criterion that settles it, and REQ-UDD-013 routes the finding to E1 if it turns out the plan cannot
be rendered without mutation.

### A.6 Drift recorded while authoring

Three discrepancies between the findings as supplied and the tree measured while authoring. None
reverses a finding; each strengthens or narrows one, and each is recorded rather than silently folded
in.

**Drift 1 — RESOLVED: all six SPECs now share baseline `d5336214e`.** As recorded at authoring time,
this SPEC was written in the worktree `.claude/worktrees/epic-update-config` on branch
`plan/epic-update-config-audit` at HEAD `7225a8b7a`, while the five sibling SPECs recorded
verification against the divergent local branch `main` at `1d4e4f7da`. The two trees differed by
three files:

```
 internal/cli/update_clean_install.go               | 11 ++-
 .../cli/update_clean_install_merge_notice_test.go  | 85 ++++++++++++++++++++++
 internal/hook/pre_tool.go                          | 53 ++++++++------
```

`pre_tool.go` was the only one touching a file this SPEC cites, and its delta was a refactor — lazy
project-root resolution via a `projectRootResolver` embed — leaving the `AstGrepGate` config mapping
(`:661-672`) and the `IsGitCommit` gating (`:430-431`) unchanged, so §A.1's conclusions held on both
trees.

The divergence no longer exists. The Epic branch has since been merged with `origin/main`, producing
**baseline HEAD `d5336214e`** (`git rev-list --count HEAD..origin/main` → 0). `1d4e4f7da` is not an
ancestor of that merge — it was a stale divergent local branch, never the trunk — while both
`7225a8b7a` and `9426bf49b` are ancestors of `d5336214e`. All six Epic SPECs are therefore
re-attributed to the single baseline `d5336214e`, and every acceptance criterion below records that
tree. Every baseline figure in this SPEC was re-observed at `d5336214e`; the only change is
AC-UDD-011's `internal/config/CLAUDE.md` count, corrected from a mis-recorded `2` to the measured
`1` (see that criterion).

**Drift 2 — F1 is false in three ways, not two.** The finding as supplied named two false halves
(loader absent; compiled default false). Measured, `gate.yaml` **is shipped** in both the template
tree and the local project, and sets `ast_grep_gate.enabled: true` explicitly — so the "gate.yaml
부재" half of §2.2's parenthetical is independently false as well. §A.1 records all three, plus the
fourth (impact claim) which the finding flagged for separate verification and which also fails.

**Drift 3 — F3's template-tree path is also absent.** The finding as supplied asserted only that
`.moai/config/config.yaml` does not exist and directed that the template path be checked rather than
assumed. Checked: `internal/template/templates/.moai/config/config.yaml` does not exist either. §5's
release checklist therefore names an unbumpable file, which is a release-process consequence and not
merely a local-path inaccuracy. §A.3 is written against the stronger measured fact.

## §B Requirements (GEARS)

### B.1 Always-loaded instruction correctness

**REQ-UDD-001** — **Where** an always-loaded instruction file (`CLAUDE.local.md`,
`internal/config/CLAUDE.md`) states a mechanism, the stated mechanism shall match the code at the
cited site. **When** a reader follows the stated mechanism to the code, the code shall exhibit the
stated behaviour.

**REQ-UDD-002** — `CLAUDE.local.md` §2.2 shall state the ast-grep pre-tool gate's actual default —
enabled in advisory mode — and shall not assert the absence of `loadGateSection`, the absence of a
shipped `gate.yaml`, or a compiled default of `false`. The corrected text shall name the gate's
actual invocation scope (evaluated on `git commit` per `internal/hook/pre_tool.go:430-431`) so the
correction does not overstate the blast radius in the opposite direction.

**REQ-UDD-003** — **Where** §2.2 describes the gate's user-visible impact, it shall record that the
suppression-policy check (`internal/hook/quality/astgrep_gate.go`, step 1) is sg-independent and can
return a blocking result, and shall not state that impact requires an explicit `moai ast-grep`
invocation with `sg` installed.

### B.2 Documented policy versus code reachability

**REQ-UDD-004** — **Where** `CLAUDE.local.md` §22.8 describes `workflow.worktree.*` toggles as an
opt-in policy, the description shall state each toggle's actual production reader status:
`auto_cleanup` and `auto_merge` unread, `auto_create` read only to select advisory wording.

**REQ-UDD-005** — The §22.8 correction shall be reconciled with whatever triage class
`SPEC-CONFIG-KEY-HONESTY-001` §A.6 assigns to those keys, and shall not itself propose implementing,
deprecating, or removing them. **Where** E4's triage lands after this SPEC's edit, the §22.8 text
shall be re-checked against that outcome rather than left asserting a status E4 has changed.

### B.3 References to nonexistent paths

**REQ-UDD-006** — An always-loaded instruction file shall not name a configuration file path that
does not exist. `CLAUDE.local.md` §9, `CLAUDE.local.md` §5, and `internal/config/CLAUDE.md` shall
stop describing `.moai/config/config.yaml` (or its template-tree counterpart) as the main
configuration file.

**REQ-UDD-007** — **Where** `CLAUDE.local.md` §5 lists files requiring a version bump at release, the
list shall contain only files that exist, so every checklist line is performable. The
`internal/template/templates/.moai/config/config.yaml` entry shall be removed or replaced with the
file that actually carries the shipped version.

### B.4 Contradiction between two always-loaded files

**REQ-UDD-008** — Two always-loaded instruction files shall not assert contradictory facts about the
same mechanism. `internal/config/CLAUDE.md`'s configuration-priority statement shall be corrected to
match the implemented override set — `MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`,
`MOAI_NO_COLOR` (per `applyEnvOverrides`), plus `MOAI_CONFIG_DIR` for the config-directory location —
and shall stop citing `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` as implemented overrides.

**REQ-UDD-009** — **Where** `internal/config/CLAUDE.md` illustrates the `envkeys.go` constant
convention, the example shall name a constant that `envkeys.go` actually declares.

**REQ-UDD-010** — An instruction file shall not require a test whose subject is unimplemented.
`internal/config/CLAUDE.md`'s "Tests MUST verify this priority via `t.Setenv` + fixture file
combinations" shall be scoped to the implemented override set, so following the instruction literally
does not produce an unpassable test or an unrequested behaviour change.

### B.5 Flag contract agreement

**REQ-UDD-011** — The `--dry-run` flag's registered help text and its actual behaviour shall agree.
**Where** the flag previews only archive operations, the help text shall not promise install
operations; **where** the help text promises install operations, the flag shall render the
clean-reinstall plan without mutating the filesystem.

**REQ-UDD-012** — The choice between narrowing the help text and extending the flag shall be recorded
as an explicit decision with its trade-off, not resolved implicitly during implementation. **Where**
the extend option is selected, the plan-construction path shall be demonstrated to perform no
filesystem mutation, preserving the `internal/cli/update.go:312` placement rationale.

*Decision recorded (v0.2.0): **option B** — make the existing non-mutating renderer reachable
(§A.5). The demonstration obligation this requirement attaches to option B is therefore live, and is
discharged by AC-UDD-002 assertion 2. The reachability fix hoists v2 detection above the `--dry-run`
branch; it does **not** relocate the early return past `internal/cli/update.go:306-326`, which would
place a `settings.json` rewrite on the dry-run path and reverse the `:312` rationale.*

**REQ-UDD-013** — This SPEC shall not change whether clean-reinstall triggers, nor its sequencing.
**Where** work on REQ-UDD-011 reveals that the plan cannot be constructed without entering the
mutating path, that finding shall be reported to `SPEC-UPDATE-REINSTALL-LOOP-002` rather than
resolved here.

### B.6 Non-functional

**NFR-UDD-001** — Every test added by this SPEC shall confine its filesystem writes to `t.TempDir()`.

**NFR-UDD-002** — No file under `internal/template/templates/**` shall be modified by this SPEC.
`CLAUDE.local.md` and `internal/config/CLAUDE.md` are repo-local maintainer documentation and are
never mirrored into the template tree (per the template isolation doctrine, `CLAUDE.local.md` §25).

**NFR-UDD-003** — Go sources added by this SPEC shall use `snake_case.go` filenames, wrap errors with
`fmt.Errorf("...: %w", err)`, and carry English comments and godoc.

**NFR-UDD-004** — Each documentation correction shall cite the code site it was verified against
(`file:line` or a content-anchored symbol name), so a future reader can re-verify without
re-deriving the measurement.

## §C Exclusions

### Out of Scope — the worktree-key triage

- Deciding whether `workflow.worktree.auto_cleanup` / `auto_merge` / `auto_create` should be
  implemented, deprecated, marked reserved, or removed; adding readers for them; and any change to
  `internal/config/defaults.go` or `internal/config/types.go` on their account. Owned by
  `SPEC-CONFIG-KEY-HONESTY-001` §A.6 / REQ-CKH-009. This SPEC corrects only the §22.8 prose and
  reconciles it with E4's outcome (REQ-UDD-005).

### Out of Scope — clean-reinstall behaviour

- Whether the clean-reinstall loop should trigger, its v2-fingerprint detection, its sequencing, and
  its interaction with `isMoAIProject`. Owned by `SPEC-UPDATE-REINSTALL-LOOP-002`. REQ-UDD-011 binds
  the `--dry-run` **contract** only; REQ-UDD-013 routes any deeper finding to E1.

### Out of Scope — CI guard pattern design

- Extending, generalising, or restructuring the leak-detection or neutrality patterns, coverage
  gating, and the `paths-filter` gating set. Owned by `SPEC-UPDATE-CI-GUARD-001` (E5). This SPEC adds
  no CI workflow change and no guard-pattern change.

### Out of Scope — merge behaviour and merge-base provenance

- The three-way merge's zero-value skip, `ValuesEqual` semantics, `systemFields` depth, and merge
  atomicity. Owned by `SPEC-UPDATE-YAML-PRESERVE-001`.
- The provenance of the merge base (base drawn from the new template rather than a snapshot of the
  old one). Owned by the unwritten `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`.

### Out of Scope — template mirroring of maintainer documentation

- Mirroring `CLAUDE.local.md` or `internal/config/CLAUDE.md` into `internal/template/templates/**`.
  Both are repo-local maintainer documentation; a template mirror would violate the template
  internal-content isolation doctrine (`CLAUDE.local.md` §25). Verified: the template tree contains
  exactly one `CLAUDE.md` (`internal/template/templates/CLAUDE.md`) and no
  `internal/template/templates/internal/` directory, so neither file ships today and neither shall be
  made to (NFR-UDD-002).

### Out of Scope — code changes in this Epic

- No code changes. This Epic's six SPECs are plan-phase artefacts only; run-phase implementation is a
  separate, separately-approved step. REQ-UDD-011's behavioural half, if the extend option is chosen
  at the plan.md §F.5 decision point, is a run-phase decision **recorded** here, not performed here.

### Out of Scope — a general documentation audit

- Auditing `CLAUDE.local.md` sections other than §2.2, §5, §9, and §22.8, or `internal/config/CLAUDE.md`
  statements other than the configuration-priority and `envkeys.go`-convention bullets. This SPEC
  fixes the five measured drifts; a general sweep is a separate scope.
