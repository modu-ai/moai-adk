# SPEC-UPDATE-DOC-DRIFT-001 — Implementation Plan

Milestones are ordered by **decision-reversibility**: the two milestones carrying genuine decisions
(M1 the `--dry-run` contract choice, M2 the E4 reconciliation dependency) come first, and the
mechanical text corrections that carry no open decision come last.

## §A Context

Baseline tree: HEAD `7225a8b7a`, branch `plan/epic-update-config-audit`, in the worktree
`.claude/worktrees/epic-update-config`. This differs from the five sibling SPECs' recorded baseline
(`main` HEAD `1d4e4f7da`) — see spec.md §A.6 drift 1 for the three-file delta and why §A.1's
conclusions hold on both trees.

Every `file:line` reference in spec.md §A was verified against this tree while authoring. Three
drifts were found and recorded rather than silently folded in (spec.md §A.6): the baseline-HEAD
difference, F1 being false in three ways rather than two, and F3's template-tree path being absent
as well as the local one.

Two properties this plan depends on:

- **The two target files are repo-local, not shipped.** `internal/template/templates/` contains
  exactly one `CLAUDE.md` and no `internal/` subtree, so neither `CLAUDE.local.md` nor
  `internal/config/CLAUDE.md` reaches a downstream user. Every edit in this SPEC is contributor-facing
  and agent-facing only.
- **Both files are always-loaded.** They enter every session's context before work begins, which is
  why a wrong sentence is an input rather than a stale comment (spec.md §A opening).

## §B Known issues carried into this SPEC

1. **E4 dependency is one-directional and unsequenced.** `SPEC-CONFIG-KEY-HONESTY-001` §A.6 owns the
   worktree-key triage; M2's §22.8 text depends on its outcome. The two SPECs may land in either
   order — M2 records the reconciliation obligation (REQ-UDD-005) rather than blocking on E4.
2. **The `--dry-run` decision may escalate out of this Epic.** If the clean-reinstall plan cannot be
   constructed without entering the mutating path, REQ-UDD-013 routes the finding to E1 and M1
   settles on the narrow-the-text option. This is a plausible outcome, not a failure mode.
3. **This checkout is shared with concurrent sessions.** `git stash` is prohibited. Falsification for
   documentation criteria uses a scratch copy under `t.TempDir()` or `/tmp`, never a tree mutation.
4. **Documentation ACs are vulnerable to vacuous satisfaction.** A criterion phrased as "the file no
   longer contains string X" is satisfiable by deletion without correction. Every documentation AC in
   acceptance.md is therefore paired with a code-fact assertion (acceptance.md §A clause 2).

## §C Pre-flight

```bash
git rev-parse --short HEAD                       # expect 7225a8b7a or a recorded successor
git branch --show-current                        # expect plan/epic-update-config-audit
go build ./... && go vet ./internal/config/... ./internal/hook/... ./internal/cli/...
go test -count=1 ./internal/config/... ./internal/hook/... ./internal/cli/...
```

## §D Constraints

- **D1 — no template-tree edits.** `internal/template/templates/**` is untouched (NFR-UDD-002).
  Neither target file is mirrored there, and neither shall be made to be.
- **D2 — no code changes in this Epic.** Plan-phase artefacts only. M1's behavioural half, if the
  extend option is selected, is a run-phase decision recorded in this plan, not performed.
- **D3 — cite the verification site.** Every correction names the `file:line` or content-anchored
  symbol it was verified against (NFR-UDD-004), so a future reader re-verifies rather than re-derives.
- **D4 — correct, do not merely delete.** Removing an offending sentence without stating the true
  mechanism satisfies a grep and leaves the reader with no guidance. Each correction states what the
  code actually does.
- **D5 — do not overclaim in the opposite direction.** §A.1's correction must state the gate's
  narrow invocation scope (`git commit` only) alongside its default-on status; a correction that
  implies the gate fires on every tool call is a new error, not a fix.
- **D6 — line numbers drift.** Anchor every criterion on a content token where one exists; use
  `file:line` only as a locating aid, never as the sole matcher.

## §E Self-verification

Each milestone closes only when its acceptance.md criteria print the stated observable output. §E of
`progress.md` carries the run-phase evidence.

## §F Milestones

### M1 — `--dry-run` contract decision and resolution (REQ-UDD-011, REQ-UDD-012, REQ-UDD-013)

The highest-change-likelihood decision in this SPEC, and the only one with a behavioural face.

**The decision, stated for review rather than resolved silently.** The help text
(`internal/cli/update.go:69`) promises "planned archive **and install** operations"; the handler
(`:293-304`) early-returns into `dryRunArchiveLegacySkills`, which previews archiving only. Two ways
to make them agree:

| Option | Change | Cost | What is lost |
|---|---|---|---|
| **A — narrow the text** | Rewrite the flag description to "Show planned archive operations without modifying the filesystem". | One line. No behaviour change. Lands entirely within this Epic's no-code-change constraint. | The highest-consequence operation — clean reinstall replacing the user's tree — remains unpreviewable. A user who runs `--dry-run` before a v2 update still learns nothing about it. |
| **B — extend the flag** | Render the clean-reinstall plan from the dry-run branch. | Requires the plan-construction path to be separable from the mutation path. Real implementation work, and its feasibility is unverified. | Nothing, if feasible. If the plan cannot be constructed without entering the mutating path, the `:312` placement rationale ("so a dry run never mutates") forbids B outright. |

**Recommended sequencing**: establish B's feasibility first with a read-only probe — determine
whether the clean-reinstall plan can be computed without any filesystem write — then choose. A B
that turns out infeasible collapses to A with the reason recorded; an A chosen without probing
forecloses B silently.

**Escalation**: if the probe shows the plan cannot be constructed without mutation, that is a finding
about clean-reinstall's structure, not about the flag. Report it to
`SPEC-UPDATE-REINSTALL-LOOP-002` per REQ-UDD-013 and settle on A.

**Deliverable either way**: the help text and the behaviour agree, and the decision plus its
rationale is recorded in `progress.md` §E.2.

### M2 — §22.8 worktree-toggle reconciliation (REQ-UDD-004, REQ-UDD-005)

The second decision, and the one with a cross-SPEC dependency.

`CLAUDE.local.md` §22.8 currently presents three toggles as an opt-in policy under a `[HARD]` marker.
Measured, only `auto_create` has a production reader, and that read selects advisory wording rather
than gating creation (spec.md §A.2).

**The correction states reader status, not triage.** The §22.8 text must say which toggles are read
and what the read does — `auto_cleanup` and `auto_merge` unread; `auto_create` read at
`internal/cli/worktree_advisory.go:29`/`:60` to select between two advisory sentences. It must not
propose implementing, deprecating, or removing them; that is `SPEC-CONFIG-KEY-HONESTY-001` §A.6 /
REQ-CKH-009 territory (spec.md §C).

**The dependency, and why it is not a blocker.** E4's triage may reclassify these keys (wire / mark
reserved / delete). REQ-UDD-005 makes the reconciliation an obligation rather than a sequencing
constraint: whichever SPEC lands second re-checks the §22.8 text against the other's outcome. Landing
this SPEC first leaves §22.8 accurate as of today and flagged for re-check; landing E4 first lets M2
write the final text directly.

**Decision to surface at review**: whether the `[HARD] ... 의도된 정책. 감사/동기화 시 "결함"으로
되돌리지 말 것` marker survives. The marker protects the `false` defaults from being "corrected" back
to `true` by a future audit — a real protection worth keeping — but it currently sits attached to a
policy claim that overstates enforcement. Options are to keep the marker scoped narrowly to the
defaults themselves, or to move it to whatever E4 records. Not resolved here.

### M3 — `internal/config/CLAUDE.md` env-override correction (REQ-UDD-008, REQ-UDD-009, REQ-UDD-010)

Mechanical once the contradiction is resolved, and the resolution direction is already measured.

`internal/config/CLAUDE.md:12` claims `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` override file
values; `:13` cites `EnvUserName` as the `envkeys.go` convention example. Neither constant exists
(`grep -n 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go` → no output), and
`applyEnvOverrides` (`internal/config/manager.go:393-406`) reads exactly four names.
`CLAUDE.local.md` §9 already states this correctly.

Three edits:

1. **`:12` priority order** — replace the two nonexistent names with the implemented set
   (`MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR`, plus
   `MOAI_CONFIG_DIR` for the config-directory location), citing `applyEnvOverrides` as the site.
2. **`:13` convention example** — replace `EnvUserName = "MOAI_USER_NAME"` with a constant
   `envkeys.go` actually declares (e.g. `EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"`). The
   convention itself (constants live in `envkeys.go`; no inline `os.Getenv("MOAI_*")`) is correct and
   is preserved.
3. **`:12` test instruction** — scope "Tests MUST verify this priority via `t.Setenv` + fixture file
   combinations" to the implemented set, so following it literally cannot produce an unpassable test
   (REQ-UDD-010).

Edit 3 is the one that matters beyond accuracy: as written, the instruction directs an agent toward
either a failing test or an unrequested env-override implementation.

### M4 — nonexistent `config.yaml` references (REQ-UDD-006, REQ-UDD-007)

Mechanical. Three sites, on two files, asserting a file that exists at neither the local nor the
template path (spec.md §A.3, §A.6 drift 3).

- `CLAUDE.local.md:415` (§9) — stop calling `.moai/config/config.yaml` the "Main configuration
  file"; describe the actual `sections/*.yaml` layout.
- `CLAUDE.local.md:250` (§5, *Files Requiring Version Sync*) — remove the
  `internal/template/templates/.moai/config/config.yaml` entry or replace it with the file that
  actually carries the shipped version. This is the release-process consequence: the line as written
  is unperformable.
- `internal/config/CLAUDE.md:5` and `:11` — the same claim ("`config.yaml` (main) plus
  `sections/*.yaml`"; "Main `config.yaml` aggregates references"), corrected identically.

**Note for M4's replacement text**: determine what does carry the shipped version before rewriting
the §5 entry — `.moai/config/sections/system.yaml` (`moai.version`) is the adjacent line and the
likely answer, but the run-phase edit should verify rather than assume, since asserting the wrong
replacement re-creates the defect in a new location.

### M5 — §2.2 ast-grep gate correction (REQ-UDD-002, REQ-UDD-003)

Largest text rewrite, but no open decision — the correct statement is fully determined by the
measurements in spec.md §A.1.

`CLAUDE.local.md:141` currently asserts four things, all false: loader absent, `gate.yaml` absent,
compiled default `false`, and impact confined to explicit `moai ast-grep` with `sg` installed.

The corrected statement must carry all four true facts **and** the scope narrowing (D5):

- The loader exists: `internal/config/loader_gate.go:20` `loadGateSection`, called from
  `internal/config/loader.go:89`.
- `gate.yaml` is shipped (template + local) and sets `ast_grep_gate.enabled: true`.
- The compiled default is `Enabled: true, BlockOnError: false, WarnOnlyMode: true`
  (`internal/config/defaults.go:318-322`) — on, in advisory mode.
- The gate is evaluated on `git commit` Bash invocations only
  (`internal/hook/pre_tool.go:430-431`, `quality.IsGitCommit`), **not** on every tool call.
- The suppression-policy check (`internal/hook/quality/astgrep_gate.go` step 1) is sg-independent
  pure Go and can return a blocking result; the sg-dependent scan (step 2) degrades gracefully when
  `sg` is absent and cannot block under `WarnOnlyMode: true`.

Everything else in §2.2 — the dogfood-experimental rationale for not mirroring the language
subdirectory tree, the `sgconfig.yml` `utils` ruleDir issue, the deferral of the 16-language ruleset
to a follow-up SPEC — is unaffected by these findings and is preserved verbatim.

## §G Anti-patterns

- **AP-1 — deleting the offending sentence instead of correcting it.** A grep-satisfying deletion
  leaves the reader with no statement of the actual mechanism, which is worse than a wrong one in the
  §2.2 case because the gate's existence then goes entirely unmentioned (D4).
- **AP-2 — overcorrecting §2.2 into "the gate fires on every tool call".** It fires on `git commit`
  only. A correction that drops the scope qualifier trades one wrong instruction for another (D5).
- **AP-3 — resolving the F4 contradiction by recency.** `CLAUDE.local.md` §9 wins because
  `envkeys.go` and `applyEnvOverrides` say so, not because it was written later. The measurement is
  the arbiter.
- **AP-4 — re-specifying the worktree-key triage in M2.** §22.8's correction states reader status.
  Implement / deprecate / mark is E4's decision (spec.md §C).
- **AP-5 — picking the `--dry-run` option silently.** M1's A-vs-B choice has a real trade-off and a
  feasibility question; resolving it inside an implementation commit hides the decision.
- **AP-6 — asserting a replacement path in M4 without checking it.** Replacing one nonexistent path
  with another unverified one reproduces the defect.
- **AP-7 — mirroring either file into the template tree.** Both are repo-local; a mirror violates the
  isolation doctrine (NFR-UDD-002, spec.md §C).
- **AP-8 — an AC that only greps for absence.** Every documentation criterion pairs with a code-fact
  criterion (acceptance.md §A clause 2); an absence-only criterion is satisfied by deletion.

## §H Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` — every claim in `progress.md` §E must
  cite an observed command.
- `CLAUDE.local.md` §2.2 (M5), §5 (M4), §9 (M4), §22.8 (M2), §25 (template isolation — NFR-UDD-002).
- `internal/config/CLAUDE.md` — M3 and M4 target; repo-local, not shipped.
- `internal/config/loader_gate.go`, `internal/config/defaults.go:318-322`,
  `internal/hook/pre_tool.go:430-431`, `internal/hook/quality/astgrep_gate.go` — M5 evidence sites.
- `internal/config/envkeys.go`, `internal/config/manager.go:393-406` — M3 evidence sites.
- `internal/cli/update.go:69`, `:293-304`, `:312`; `internal/cli/update_archive.go:339-353` — M1
  evidence sites.
- `SPEC-CONFIG-KEY-HONESTY-001` §A.6 — M2's reconciliation dependency.
- `SPEC-UPDATE-REINSTALL-LOOP-002` — M1's escalation target (REQ-UDD-013).
