# SPEC-UPDATE-DOC-DRIFT-001 — Implementation Plan

Rewritten at v0.3.0. Milestones are ordered by **decision-reversibility**: the milestones carrying
genuine judgment calls come first, the mechanical text corrections last. The v0.2.0 M1 (`--dry-run`)
is retired and recorded at §F.0 rather than deleted, so a reader of this plan can see why it is gone.

## §A Context

**Baseline** worktree HEAD `7f61332ef`, branch `docs/spec-doc-drift-rewrite`, based on `origin/main`.

The v0.2.0 baseline `d5336214e` is **not an ancestor of this tree** —
`git merge-base --is-ancestor d5336214e HEAD` exits `1` — so no v0.2.0 `file:line`, count, or
verbatim output was carried into v0.3.0. Everything in spec.md §A and acceptance.md §B was
re-observed here. Five cited files moved line numbers; four drifts are recorded at spec.md §A.8, one
of which (drift 3) reverses the direction of a correction this plan previously specified.

Three properties this plan depends on:

- **The always-loaded surface shrank.** `CLAUDE.local.md` was consolidated: former §18-§27 were
  externalized into `.moai/docs/*.md` and the file now runs §1-§17 plus a `## References` table
  (511 lines). Two of this SPEC's targets moved out with that change (spec.md §A.0), and the
  requirements that follow them carry a reduced-severity note: a `.moai/docs/` file is read on
  demand, not injected into every session.
- **None of the four target files is shipped.** `internal/template/templates/` contains exactly one
  `CLAUDE.md`, no `internal/` subtree, and a `.moai/docs/` holding only `agent-lint.md` and
  `generic-patterns-guide.md`. Every edit here is contributor-facing and agent-facing only.
- **This SPEC now touches no Go file.** Every code-touching requirement is retired (§F.0), so the
  change surface is markdown in five path prefixes. AC-UDD-023 makes that mechanical.

## §B Known issues carried into this SPEC

1. **The `auto_cleanup` correction is the opposite of what v0.2.0 specified.** Two production readers
   appeared since the v0.2.0 baseline. Executing the v0.2.0 M2 text as written would write a false
   statement into the doc. §F.2 is rewritten against the measurement, and acceptance.md §C-4 is run
   *before* the edit rather than after.
2. **A correction that quotes what it retracts defeats its own acceptance criterion.** The
   2026-08-01 fix to §2.2 is correct prose whose quoted retraction leaves `grep -c '로더 부재'` at
   `1`. acceptance.md §A clause 8 states the resulting constraint on the M5 edit: state the truth,
   do not quote the retracted claim.
3. **The duplicate with `SPEC-INTERNAL-ARCH-001` REQ-ARCH-006 is resolved on one side only.** That
   SPEC is under concurrent review in this session; editing it would race another agent. §F.6 reports
   rather than edits, and AC-UDD-024 is red by construction until its owner acts.
4. **This checkout is shared with concurrent sessions.** `git stash` is prohibited. Falsification
   uses a scratch copy under `/tmp`, never a tree mutation.
5. **Documentation ACs are vulnerable to vacuous satisfaction.** Every removal count is paired with a
   replacement count over the same scope (acceptance.md §A clause 3), and every documentation
   criterion is paired with a code-fact criterion (clause 2).

## §C Pre-flight

```bash
git rev-parse --short HEAD                    # expect 7f61332ef or a recorded successor
git branch --show-current                     # expect docs/spec-doc-drift-rewrite
git diff --name-only 7f61332ef..HEAD | grep -c '\.go$'   # expect 0 (no Go change in this SPEC)
go build ./... && go vet ./...                # expect exit 0, exit 0
```

No test package is in pre-flight scope. This SPEC edits no Go file, so binding it to any suite would
gate a documentation change on unrelated conditions — in particular on `TestBranchGuard_Latency` in
`internal/hook`, a load-sensitive performance test that fails under a parallel full-suite run and
passes alone (acceptance.md AC-UDD-022). `go build` / `go vet` still cover the whole module.

## §D Constraints

- **D1 — no template-tree edits.** `internal/template/templates/**` is untouched (NFR-UDD-002). None
  of the four target files is mirrored there, and none shall be made to be.
- **D2 — no code changes.** Documentation only. The behavioural half of the former M1 was landed by
  a sibling SPEC (§F.0) and is not re-performed here.
- **D3 — cite the verification site.** Every correction names the `file:line` or content-anchored
  symbol it was verified against (NFR-UDD-004).
- **D4 — correct, do not merely delete.** Removing an offending sentence without stating the true
  mechanism satisfies a grep and leaves the reader with nothing.
- **D5 — do not overclaim in the opposite direction.** §2.2's correction must carry the invocation
  scope (both conjuncts) alongside the blocking statement; a correction implying the gate fires on
  every tool call is a new error, not a fix.
- **D6 — do not retract by quotation.** Stating the current mechanism is the correction; quoting the
  superseded claim re-introduces its tokens and defeats the criterion (acceptance.md §A clause 8).
- **D7 — line numbers drift.** Anchor every criterion on a content token; use `file:line` only as a
  locating aid. The v0.3.0 rewrite exists largely because v0.2.0's anchors did not survive one
  consolidation.

## §E Self-verification

Each milestone closes only when its acceptance.md criteria print the stated observable output. §E of
`progress.md` carries the run-phase evidence. AC-UDD-024 is the single criterion permitted to close
red, and only with its open half named in `progress.md` §E.1.

## §F Milestones

### §F.0 — Retired milestone: the `--dry-run` contract (former M1)

Kept as a record so the retirement is auditable rather than silent.

v0.2.0 settled the A-vs-B choice as **option B** — make the existing non-mutating clean-reinstall
plan renderer reachable — and specified the constraint that the `--dry-run` early return must not
move past `stripRetiredV2DenyEntries`, which rewrites `settings.json`.

That work landed, in `SPEC-UPDATE-REINSTALL-LOOP-002` (status `completed`) as REQ-RIL2-024/025/026,
under exactly that constraint:

```
$ sed -n '344,355p' internal/cli/update.go
344:		// SPEC-UPDATE-REINSTALL-LOOP-002 REQ-RIL2-024/025 (M4): the v2
345:		// fingerprint is computed HERE — inside the dry-run branch, above the
346:		// deny-rule migration below — so the clean-reinstall and
347:		// residue-cleanup plans become reachable from `moai update --dry-run`.
353:		// The early return itself does NOT move (REQ-RIL2-026): it stays
354:		// ABOVE stripRetiredV2DenyEntries, which rewrites settings.json.
355:		return emitDryRunReinstallPlan(cmd.Context(), cwd, getBoolFlag(cmd, "force"), out, th)
```

REQ-UDD-011 and REQ-UDD-012 are satisfied; REQ-UDD-013, whose content was to escalate to that same
sibling if the plan could not be rendered without mutation, is moot. AC-UDD-001 remains as a
regression guard; AC-UDD-002 and AC-UDD-003 are retired.

The v0.2.0 worry that the two SPECs might push the early return in opposite directions did not
materialise — worth recording, because the coordination cost was paid in specification and the
outcome confirms it was not wasted.

### M2 — §22.8 worktree-toggle status, polarity inverted (REQ-UDD-004)

**First, because it is the milestone whose direction changed and the one that is easiest to get
wrong.** Executing the v0.2.0 text here would introduce a falsehood.

Target: `.moai/docs/local-dev-settings-intent.md:58-65` (§22.8), which received this content when
`CLAUDE.local.md` §18-§27 were externalized. `CLAUDE.local.md:506` is the References entry pointing
at it.

**Run acceptance.md §C-4 before editing.** It re-measures the reader count and prints the reader's
context. If `Worktree.AutoCleanup` returns `0` production readers, the v0.2.0 correction was right,
this inversion is wrong, and the milestone stops for re-specification.

The measured state (spec.md §A.2):

| Toggle | Production readers | What the read does |
|---|---|---|
| `auto_cleanup` | 2 — `session_worktree.go:584`, `session_worktree_prmerge.go:122` | **gates disposal**: `false` returns early and the worktree persists |
| `auto_merge` | 0 | — |
| `auto_create` | 1 — `worktree_advisory.go:60` | selects between two advisory sentences |

The current text says `auto_merge` / `auto_cleanup` have no reader and that no code path consumes
them. Half of that is now false, which is the worst available state: a reader who spot-checks
`auto_merge`, finds the doc right, and extends the confidence to `auto_cleanup`.

**The judgment call, surfaced rather than resolved here.** The shipped default for `auto_cleanup`
remains `false` while it now gates real behaviour. Whether that combination is itself a defect is
`SPEC-CONFIG-KEY-HONESTY-001` territory (spec.md §C) — this milestone records the reader status and
takes no position. The `[HARD] … 의도된 정책` marker protecting the `false` defaults from being
"corrected" back to `true` by a future audit is a real protection and is preserved; what changes is
only the reader-status sentence beneath it. AC-UDD-004's third count is the guard that the marker's
subject matter is not deleted along with the wrong sentence.

The two reader sites carry comments citing `CLAUDE.local.md §22.8`. Those still resolve through the
References table (spec.md §A.0), so they are left alone — see §G AP-11.

### M5 — §2.2 ast-grep gate: the second correction (REQ-UDD-002, REQ-UDD-003)

**Second, because it is the largest rewrite and carries the D6 constraint.**

Target: `CLAUDE.local.md:146` (§2.2 — `:141` at v0.2.0). The section is one physical line.

Its four original false claims were retracted in place on 2026-08-01, so the falsehood halves of both
requirements are discharged (spec.md §A.1). Three things remain to do, and one thing not to do.

**(a) Remove the quotation, keep the substance.** The retraction quotes its predecessor —
"종전 이 절은 `…로더 부재 → … 항상 컴파일 기본값 false`라고 적었으나 두 주장 모두 … 거짓이다" — which
leaves both retracted tokens in the file (D6). The corrected section states the mechanism directly.
The dated marker and the `v3.0.1`-vs-`main` release-version distinction are **substantive and are
kept**: they record that released `v3.0.1` genuinely has no loader (issue #1265) and that the
resolution is a release, not a code change. That is a fact about two trees, not a quotation.

**(b) State the invocation scope, both conjuncts** (REQ-UDD-002). The gate is evaluated at
`internal/hook/pre_tool.go:447` under
`quality.IsGitCommit(command) && !config.IsAutonomyTierCommitGateOff(config.AutonomyTier())`. The
second conjunct is new since v0.2.0 (spec.md §A.8 drift 2); naming only `git commit` would be
incomplete on this tree.

**(c) Correct the new misstatement** (REQ-UDD-003). The current text concludes that blocking is
reachable only via a `gate.yaml` opt-in. It is not: `RunAstGrepGateV2` step 1
(`internal/hook/quality/astgrep_gate.go:46-61`) walks source files in pure Go and returns `false` at
`:60` whenever a suppression-policy violation is found — `WarnOnlyMode` is passed only into the
step-2 scanner config at `:68` and does not reach that return. The corrected text states: the
suppression-policy check is sg-independent and blocks irrespective of `WarnOnlyMode`; the
sg-dependent scan degrades gracefully when `sg` is absent (`ErrScannerUnavailable` → pass, `:79`).

**What not to do.** The dogfood-experimental rationale, the `sgconfig.yml` `utils` ruleDir issue, and
the 16-language-ruleset deferral are untouched by any finding and are preserved (AC-UDD-020).

### M4 — nonexistent paths: `config.yaml` and `.agency/` (REQ-UDD-006)

**Third: mechanical for `config.yaml`, one judgment call for `.agency/`.**

`config.yaml` — three sites, unchanged in substance since v0.1.0, at two files:

- `CLAUDE.local.md:328` (§9) — stop calling `.moai/config/config.yaml` the "Main configuration file";
  describe the `sections/*.yaml`-only layout that the same section already describes four lines
  below, so the file stops contradicting itself.
- `internal/config/CLAUDE.md:5` and `:11` — the same claim (`` `config.yaml` (main) plus
  `sections/*.yaml` ``; "Main `config.yaml` aggregates references"), corrected identically.

`.agency/` — four sites in `CLAUDE.local.md`, newly folded into this requirement at v0.3.0:

- `:88` names `internal/template/templates/.agency/` as a template-source directory. It does not
  exist.
- `:94`, `:106`, `:108` instruct, inside a `[HARD]` Template-First rule, that new `.agency/` files be
  mirrored into that directory, and that a pre-commit check verify the mirror. Both are
  unperformable.

**The judgment call: how much to keep.** `.agency/` is live in the code but **inbound only** — a
legacy layout read *out of* a user project by `moai migrate agency` (`migrate_agency.go:200`), the v2
fingerprint detector (`v2_detection.go:282`), and residue cleanup (`update_residue_cleanup.go:84`),
with a `internal/defs/dirs.go:134` entry. Deleting every mention would remove a fact contributors
still need; keeping the Template-First arm asserts a direction that never existed. The correction
therefore **keeps** `.agency/` and **re-describes** it as legacy-inbound, removing it from the
template-source list and from the mirror instruction. AC-UDD-025's second and third counts are the
guards on each half of that.

### M4b — the version-sync checklist, re-anchored (REQ-UDD-007)

Target moved: `CLAUDE.local.md` §5 is now a three-line pointer, and the *Files Requiring Version
Sync* list lives at `.moai/docs/version-management.md:66-78`.

`:78` names `internal/template/templates/.moai/config/config.yaml (moai.version)` — the file
AC-UDD-012 shows does not exist. The line is unperformable: a releaser skips it silently or creates a
config file nothing loads.

**The replacement was measured before being specified** (§G AP-6). There is no
`internal/template/templates/.moai/config/sections/system.yaml` either; the template ships
`system.yaml.tmpl` with `version: "{{.Version}}"` at `:6`, injected at render time. So the template
side carries **no** hand-bumped version, and the correct edit is deletion of `:78` plus a one-line
note recording the render-time injection — not substitution of another path, which would re-create
the defect at a new location. `:77` (`.moai/config/sections/system.yaml`) is correct and stays.

### M3 — `internal/config/CLAUDE.md` env-override correction (REQ-UDD-008, REQ-UDD-009, REQ-UDD-010)

**Last among the edits: fully mechanical, direction already measured, no open decision.**

`:12` claims `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` override file values; `:13` cites
`EnvUserName` as the `envkeys.go` worked example. Neither constant exists, and `applyEnvOverrides`
(`internal/config/manager.go:398-411`) reads exactly four names. `CLAUDE.local.md` §9 already states
this correctly, so the contradiction is resolved against the code, not by recency (§G AP-3).

Three edits:

1. **`:12` priority order** — replace the two nonexistent names with the implemented set
   (`MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR`, plus
   `MOAI_CONFIG_DIR` for the config-directory location), citing `applyEnvOverrides` as the site.
2. **`:13` convention example** — replace `EnvUserName = "MOAI_USER_NAME"` with a constant
   `envkeys.go` declares; `EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"` is confirmed present. The
   convention itself — constants live in `envkeys.go`, no inline `os.Getenv("MOAI_*")` — is correct
   and preserved.
3. **`:12` test instruction** — scope "Tests MUST verify this priority via `t.Setenv` + fixture file
   combinations" to the implemented set.

Edit 3 is the one that matters beyond accuracy: as written, the instruction directs an agent toward
either a failing test or an unrequested env-override implementation.

### M6 — report the unresolved half of the duplicate (REQ-UDD-014)

Not an edit. `SPEC-INTERNAL-ARCH-001` REQ-ARCH-006 claims the same fix M3 performs; spec.md §A.7
resolves ownership in favour of this SPEC and states the three reasons.

The reciprocal edit — marking REQ-ARCH-006 superseded and pointing AC-ARCH-007 here — belongs to that
SPEC's owner and is **not** made from this worktree: that SPEC is under concurrent review in this
session (§B.3). This milestone's deliverable is the report, and AC-UDD-024 is the detector that keeps
the open half visible instead of assumed away. Closing this SPEC with AC-UDD-024 red is permitted,
provided `progress.md` §E.1 names it.

## §G Anti-patterns

- **AP-1 — deleting the offending sentence instead of correcting it.** A grep-satisfying deletion
  leaves the reader with no statement of the actual mechanism, which is worse than a wrong one in the
  §2.2 case because the gate's existence then goes unmentioned entirely (D4).
- **AP-2 — overcorrecting §2.2 into "the gate fires on every tool call".** It fires under two
  conjuncts. Dropping the scope qualifier trades one wrong instruction for another (D5).
- **AP-3 — resolving the env-var contradiction by recency.** `CLAUDE.local.md` §9 wins because
  `envkeys.go` and `applyEnvOverrides` say so, not because it was written later.
- **AP-4 — re-specifying the worktree-key triage in M2.** The correction states reader status.
  Implement / deprecate / mark is `SPEC-CONFIG-KEY-HONESTY-001` territory (spec.md §C).
- **AP-6 — asserting a replacement path without checking it.** Replacing one nonexistent path with
  another unverified one reproduces the defect. M4b's deletion-plus-note is the measured outcome of
  actually looking.
- **AP-7 — mirroring any target file into the template tree.** All four are repo-local (NFR-UDD-002).
- **AP-8 — an AC that only greps for absence.** Every removal count is paired with a replacement
  count (acceptance.md §A clause 3).
- **AP-11 — sweeping the retired `§18`-`§27` citations** *(new at v0.3.0)*. 44 Go comments and 8
  rule/doc sites cite the retired section numbers. They still resolve: the `## References` table
  preserves the §-number → file mapping and the receiving files keep their §-numbered headings. This
  was measured before it could be proposed as a defect, which is the point — a repo-wide sweep on a
  false premise is a large, confident, wrong change.
- **AP-12 — carrying a v0.2.0 figure into a v0.3.0 claim** *(new at v0.3.0)*. The old baseline is not
  an ancestor of this tree. Every number in these artifacts was re-observed; drift 3 shows what
  happens when one is not — the correction it specified would have written a falsehood.
- **AP-13 — editing `SPEC-INTERNAL-ARCH-001` from this worktree** *(new at v0.3.0)*. It is under
  concurrent review. §F.6 reports; it does not reach across.

## §H Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` — every claim in `progress.md` §E must
  cite an observed command.
- `CLAUDE.local.md` §2 / §2.2 (M5), §9 (M4), §2 Template-First `.agency/` arm (M4), `## References`
  table (context for the re-anchored targets).
- `internal/config/CLAUDE.md` — M3 and M4 target; repo-local, not shipped.
- `.moai/docs/local-dev-settings-intent.md` §22.8 — M2 target (received `CLAUDE.local.md` §22).
- `.moai/docs/version-management.md` §*Files Requiring Version Sync* — M4b target (received §5).
- `internal/config/loader_gate.go:20`, `internal/config/loader.go:89`,
  `internal/config/defaults.go:438-442`, `internal/hook/pre_tool.go:447-448`,
  `internal/hook/quality/astgrep_gate.go:46-79` — M5 evidence sites.
- `internal/config/envkeys.go`, `internal/config/manager.go:398-411` — M3 evidence sites.
- `internal/cli/session_worktree.go:584`, `internal/cli/session_worktree_prmerge.go:122`,
  `internal/cli/worktree_advisory.go:60` — M2 evidence sites.
- `internal/cli/migrate_agency.go:200`, `internal/cli/v2_detection.go:282`,
  `internal/cli/update_residue_cleanup.go:84`, `internal/defs/dirs.go:134` — M4's `.agency/`
  direction evidence.
- `internal/template/templates/.moai/config/sections/system.yaml.tmpl:6` — M4b's render-time
  injection evidence.
- `internal/cli/update.go:81`, `:344-355`, `:546-600`; `internal/cli/update_dry_run_reach_test.go:186`
  — §F.0 retirement evidence.
- `SPEC-UPDATE-REINSTALL-LOOP-002` (completed) — landed the retired M1.
- `SPEC-CONFIG-KEY-HONESTY-001` (completed) — discharged REQ-UDD-005.
- `SPEC-INTERNAL-ARCH-001` REQ-ARCH-006 / AC-ARCH-007 — the duplicate resolved at spec.md §A.7.
