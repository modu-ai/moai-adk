# SPEC-UPDATE-DOC-DRIFT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: []
code_baseline: d5336214e
worktree_head_at_revision: 145e601c9
branch: plan/epic-update-config-audit
worktree: .claude/worktrees/epic-update-config
findings: 5 (§A.1 - §A.5)
requirements: 17 (REQ-UDD-001..013, NFR-UDD-001..004)
acceptance_criteria: 23 (AC-UDD-001..023), 14 of which form 7 documentation/code-fact pairs
drift_recorded: 3 (§A.6)
open_decisions: 1 (plan.md §F M2 — §22.8 [HARD] marker scope)
settled_decisions: 1 (plan.md §F M1 — `--dry-run` resolved as option B, v0.2.0)
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.65
    threshold: 0.80
    dimensions: {clarity: 0.70, completeness: 0.85, testability: 0.50, traceability: 0.65}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7, D8, D9, D10, D11, D12, D16]
    deferred: [D13, D14, D15, D17]
    report: .moai/reports/plan-audit/SPEC-UPDATE-DOC-DRIFT-001.md
```

### Baseline

- **Code baseline** `d5336214e`; **worktree HEAD at the plan-audit revision** `145e601c9`, a
  descendant that changes SPEC documents only. Observed:
  `git merge-base --is-ancestor d5336214e HEAD` → exit `0`;
  `git diff --name-only d5336214e HEAD | grep -cv '\.md$'` → `0`;
  `git diff --stat d5336214e HEAD -- '*.go'` → empty.
- This resolves plan-audit **D16**: the v0.1.0 artifacts named only `d5336214e` while plan.md §C
  promised "or a recorded successor" without recording one. All three artifacts now name both.

### v0.2.0 — plan-audit revision (D1-D12 + D16 resolved)

Iteration-1 verdict on v0.1.0: **FAIL, 0.65** against the Tier M threshold 0.80. Must-Pass 7/7
passed; the failure was concentrated in Testability (0.50). Two audit-confirmed strengths were
preserved unchanged: §A.6 drift 1 (this SPEC's baseline correction was right — `1d4e4f7da` was a
stale divergent branch, and the `update.go:312` citation was correctly left unshifted while siblings
moved +22), and falsification procedures C-1/C-2/C-3, which the audit re-ran and found to work.

| Defect | Severity | Resolution |
|---|---|---|
| D1 | critical | Option B's cost framing was **inverted**. spec.md §A.5 and plan.md §F M1 rewritten against measurement: the non-mutating renderer exists at `update_clean_install.go:186-198` and is wired at `update.go:360`; B is a local re-ordering, not new construction. **Decision settled: option B**, with the non-relocation constraint stated. AC-UDD-002 gained the no-mutation assertion the audit's fix item 4 required. |
| D2 | major | AC-UDD-020 was **arithmetically unsatisfiable** (`grep -c` over one line vs `>= 2`) and its baseline false. Replaced with a distinct-token count (`grep -o \| sort -u \| wc -l`), expected `3`, observed `3`. |
| D3 | major | AC-UDD-009 rebuilt on the AC-UDD-007 removal/replacement pattern (`EnvUserName` → `0`, `EnvDevelopmentMode` → `>= 1`). |
| D4 | major | AC-UDD-010 given a mechanical predicate scoped to the resolved `t.Setenv` line number. |
| D5 | major | AC-UDD-011's regex widened to tolerate backticks so `internal/config/CLAUDE.md:5` is in scope; the scope-narrowing sentence ("which remains the assertion this criterion is about") removed. Corrected baseline for the widened form: `2`. |
| D6 | major | AC-UDD-019's window `41,70` → `41,85` (the expected fact is at `:79`), `ErrScannerUnavailable` given its own command, and the non-verbatim baseline block replaced with observed output. |
| D7 | major | AC-UDD-021 made baseline-relative (`d5336214e..HEAD` diff + commit count); the unstaged-only form went inert once a change was committed. |
| D8 | major | AC-UDD-023 converted to a before/after `git status --porcelain` delta; the absolute check could not separate the test run's writes from this SPEC's own edits. |
| D9 | major | AC-UDD-022's "green" baseline was **unobserved** — the full suite exits 1 on `TestBranchGuard_Latency` (load-sensitive; passes alone). Scoped to the three packages this SPEC touches, with `go build ./...` / `go vet ./...` retained module-wide. |
| D10 | major | AC-UDD-005's "exactly …" enumeration was incomplete and already failed at baseline. Replaced with the selector-scoped assertion `Worktree.AutoCleanup\|Worktree.AutoMerge` → `0`; spec.md §A.2's output block declared abbreviated (33 lines) with the homonym set enumerated. |
| D11 | major | `sed -n '141p'` replaced by a content-anchored section range across AC-UDD-014c / 016 / 018 / 020, per plan.md D6. Resolves D2 structurally as well. |
| D12 | minor | spec.md §A.5's `:305-325` corrected: `:306-326` is the retired-deny-rule migration (mutating), clean-reinstall detection begins at `:328`, `runCleanReinstall` is called at `:359`. Folded into the D1 edit per the audit's own recommendation. |
| D16 | minor | Worktree HEAD `145e601c9` recorded in all three artifacts (see Baseline above). |

Deferred to a later revision, unchanged here: **D13** (REQ→AC back-references; `REQ-UDD-001`,
`REQ-UDD-013`, `NFR-UDD-003` have no directly-bound AC), **D14 residue** (AC-UDD-006's loose `E4`
token and its unverified "no proposal" half — AC-UDD-001/003 and C-4 were fixed under D1),
**D15** (AC-UDD-012's third command is `ls` alias-sensitive), **D17** (info —
`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` does not exist; spec.md self-declares it as unwritten, so
this is a forward reference, not a broken one).

### Observed at revision time (2026-07-31, HEAD `145e601c9`)

Every baseline changed in this revision was re-run, not re-attributed:

- `sed -n '/§2.2 astgrep-rules/,/^### \[HARD\] settings.local.json/p' CLAUDE.local.md | grep -oE 'dogfood|sgconfig\.yml|utils' | sort -u | wc -l` → `3` (D2)
- `sed -n '141p' CLAUDE.local.md | grep -cE 'dogfood|sgconfig\.yml|utils'` → `1`, disproving the recorded `>= 2` (D2)
- `grep -c 'EnvUserName' internal/config/CLAUDE.md` → `1`; `grep -c 'EnvDevelopmentMode' internal/config/CLAUDE.md` → `0` (D3)
- `grep -n 't\.Setenv' internal/config/CLAUDE.md` → line `12`; that line matches the unimplemented pair `1`, the implemented set `0` (D4)
- `grep -cE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' internal/config/CLAUDE.md` → `2` (`:5` and `:11`) (D5)
- `sed -n '41,85p' internal/hook/quality/astgrep_gate.go | grep -nE '…'` → relative `6, 20, 23, 39`; `grep -n -A2 'ErrScannerUnavailable' …` → `79: … 80: return true, astGrepReasonScannerUnavailable` (D6)
- `git diff --stat d5336214e..HEAD -- internal/template/templates/` → empty; `git log --oneline d5336214e..HEAD -- internal/template/templates/ | wc -l` → `0` (D7)
- before/after `git status --porcelain` around `go test ./internal/config/` → `diff` exit `0`, no output (D8)
- `go build ./...` → `0`; `go vet ./...` → `0`; `go test -count=1 ./internal/cli/ ./internal/config/ ./internal/hook/quality/` → `ok` ×3 (156.762s / 1.884s / 4.174s) (D9)
- `grep -rn 'Worktree\.AutoCleanup\|Worktree\.AutoMerge' … | wc -l` → `0`; `grep -rn 'Worktree\.AutoCreate' …` → one line, `worktree_advisory.go:60`; survey grep → `33` lines (D10)
- `grep -n '§2.2 astgrep-rules' CLAUDE.local.md` → `141:`; terminating anchor at `:143` (D11)
- `sed -n '306,313p;328,329p;359,360p' internal/cli/update.go` → deny-rule migration at `:306`, placement rationale at `:312-313`, v2 detection at `:328`, `runCleanReinstall` + `DryRun` wiring at `:359-360` (D1, D12)
- `sed -n '50,53p;184,198p' internal/cli/update_clean_install.go` → `CleanReinstallOptions.DryRun` and the non-mutating renderer with `return result, nil` at `:197` (D1)

### Residual risk

- `TestBranchGuard_Latency` (`internal/hook`) is load-sensitive: it fails under the parallel full
  suite and passes alone. It is **not diagnosed here** — this SPEC changes no Go file in that package
  — and AC-UDD-022 is scoped around it rather than through it. If a future revision widens the test
  scope, this flake returns.
- The `§2.2` section-range anchor terminates on `### [HARD] settings.local.json Separation`. Moving
  §2.2 out from under that heading breaks the four criteria that use it.
- REQ-UDD-005's E4 reconciliation is a **time coupling**, not a dependency edge: whichever of this
  SPEC and `SPEC-CONFIG-KEY-HONESTY-001` lands second must re-check the §22.8 text against the
  other's outcome.

### Epic run order (dependency sequencing)

`SPEC-UPDATE-DOC-DRIFT-001` declares `depends_on: []` — it records `related_specs` only — so its
run-phase `Depends_on Pre-flight Check` is trivially satisfied. Every SPEC in this Epic is currently
`status: draft`, which is the expected state for an Epic whose members have not yet run; it is a
sequencing fact, not a per-SPEC defect.

The order below is consistent with the orders recorded in `SPEC-UPDATE-REINSTALL-LOOP-002` §E.1,
`SPEC-UPDATE-DATA-SURVIVAL-001` §E.1, `SPEC-CONFIG-TIER-PERSIST-001` §E.1,
`SPEC-CONFIG-KEY-HONESTY-001` §E.1, and `SPEC-UPDATE-CI-GUARD-001` §E.1:

| Order | SPEC | Constraint |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | Epic entry point |
| 2 | `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) | `depends_on: [E1]` |
| 3 | `SPEC-CONFIG-TIER-PERSIST-001` (E3) | `depends_on: [E2]` |
| 4 | `SPEC-CONFIG-KEY-HONESTY-001` (E4) | `depends_on: [E3]` |
| 5+ | `SPEC-UPDATE-CI-GUARD-001` (E5), `SPEC-UPDATE-YAML-PRESERVE-001`, **this SPEC** (E6) | no `depends_on` edge |

**Do not invoke `/moai run` on this SPEC with `--ignore-deps`.** There is nothing to bypass — the
frontmatter declares no dependency — so an override would substitute a flag for an ordering that
costs nothing to honour, and would set a precedent the siblings with real edges must not follow.

Two **soft** ordering preferences bind this SPEC without being dependency edges, and both are
handled inside the SPEC rather than by frontmatter:

- **M2 versus E4.** `SPEC-CONFIG-KEY-HONESTY-001` §A.6 owns the worktree-key triage that M2's §22.8
  text must agree with. REQ-UDD-005 makes this a reconciliation obligation on whichever SPEC lands
  second, not a blocker. Landing this SPEC first leaves §22.8 accurate as of today and flagged for
  re-check; landing E4 first lets M2 write the final text directly. **Preference: after E4**, which
  the order above already gives.
- **M1 versus E1.** Both SPECs touch the `--dry-run` branch of `internal/cli/update.go`, and both
  settle it the same way (make the existing non-mutating renderer reachable; never relocate the
  early return). E1 owns the reachability change as REQ-RIL2-024/025; this SPEC owns the help-text
  contract. **Preference: after E1**, so M1 verifies a reachability fix that has already landed
  rather than authoring a competing one. If this SPEC runs first, M1 must implement the hoist itself
  and E1's M4 becomes a no-op verification — either order is correct, but the two must not both
  edit that branch concurrently.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
