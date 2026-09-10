# SPEC-WORKTREE-BASEREF-001 — Run-phase Verdict

Card **t313** · Tier M · cycle_type=tdd
Worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`
Measured at HEAD **`8c46460ff`** unless a row says otherwise. Start of run: `d0bc4bba5`.
Verbatim command output: `.moai/state/verify/t313/` — 50 files (persistent path, survives `/tmp` clearance).

> **Evidence-location caveat.** `.moai/state/` is gitignored (`.gitignore:284`), so those 50 files exist ONLY in this worktree and are destroyed when it is disposed. This verdict is tracked and survives; the verbatim output it cites does not. Anyone who needs the raw output after disposal must copy `.moai/state/verify/t313/` to the primary checkout first.

**Verdict: COMPLETE — 16/16 AC PASS, 0 FAIL, 0 PASS-WITH-DEBT, no blockers.**

---

## 1. Per-AC matrix

Every `-run` invocation carries `-v`; both the exit code AND the `=== RUN` count are recorded, per acceptance.md §D.3. A green exit code with a zero RUN count would be a FAIL; none occurred.

| AC | Verdict | Command | Exit | `=== RUN` | Evidence file |
|---|---|---|---|---|---|
| AC-WBR-001 Schema + neutral default | PASS | `go test ./internal/config -run 'GitStrategy' -count=1 -v` | 0 | 35 | `ac-wbr-001.txt` |
| AC-WBR-002 Template neutrality | PASS | value-side grep (below) + `go test ./internal/template -run 'WorktreeBaseBranchTemplate' -count=1 -v` | 1 (grep: no match = PASS) / 0 (guard) | n/a / 1 | `ac-wbr-002.txt`, `ac-wbr-002-guard.txt` |
| AC-WBR-003 Unset → byte-identical | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(Unset\|Empty)' -count=1 -v` | 0 | 1 | `ac-wbr-003.txt` |
| AC-WBR-004 Match → no write, no output | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Match' -count=1 -v` | 0 | 1 | `ac-wbr-004.txt` |
| AC-WBR-005 Mismatch → write + exactly one line | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Mismatch' -count=1 -v` | 0 | 1 | `ac-wbr-005.txt` |
| AC-WBR-006 Fail-open | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(FailOpen\|GitError)' -count=1 -v` | 0 | 2 | `ac-wbr-006.txt` |
| AC-WBR-007 `moai cc -w` cuts from the base | PASS | `go test ./internal/cli -run 'SessionWorktree.*Base' -count=1 -v -timeout 600s` | 0 | 5 | `ac-wbr-007.txt` |
| AC-WBR-008 Empty/unresolvable → today's argv | PASS | `... -run 'SessionWorktree.*(NoBase\|Unresolvable)' ...` | 0 | 3 | `ac-wbr-008a.txt` |
| AC-WBR-008 third assertion (shared predicate) | PASS | `... -run 'SessionWorktree.*(SharedPredicate\|Resolver)' ...` | 0 | 2 | `ac-wbr-008b.txt` |
| AC-WBR-009 Doctor, four distinct states | PASS | `go test ./internal/cli -run 'Doctor.*WorktreeBaseBranch' -count=1 -v -timeout 600s` | 0 | 7 | `ac-wbr-009.txt` |
| AC-WBR-009 manual confirmation | PASS | `./bin/moai doctor --check 'Worktree Base Branch'` | 0 | n/a | `ac-wbr-009-doctor-run.txt` |
| AC-WBR-010 Registered + rendered | PASS | `go test ./internal/web -run 'WorktreeBaseBranch' -count=1 -v` | 0 | 2 | `ac-wbr-010-web.txt` |
| AC-WBR-010 (schema half) | PASS | `go test ./internal/settings -run 'AllFields' -count=1 -v` | 0 | 1 | `ac-wbr-010-allfields.txt` |
| AC-WBR-011 Free-text control (schema half) | PASS | `go test ./internal/settings -run 'WorktreeBaseBranch.*Type' -count=1 -v` | 0 | 1 | `ac-wbr-011-schema.txt` |
| AC-WBR-011 (render half) | PASS | `go test ./internal/web -run 'WorktreeBaseBranch.*(Text\|FreeText)' -count=1 -v` | 0 | 1 | `ac-wbr-011-render.txt` |
| AC-WBR-012 Anti-dead-key guard, 3-part conjunction | PASS | `go test ./internal/web ./internal/hook ./internal/cli -run 'WorktreeBaseBranch' -count=1 -v -timeout 600s` | 0 | 21 | `ac-wbr-012.txt` |
| AC-WBR-012 mutation check | PASS | same command with the `FieldDef` removed from `gitStrategyFields()` | **1** (guard fails, as required) | 2 | `ac-wbr-012-mutation.txt` |
| AC-WBR-013 Template-first parity | PASS | `make build` + the diff-scoped probes (§3) | 0 | n/a | `ac-wbr-013-parity.txt`, `make-build-final.txt` |
| AC-WBR-014 Typed round trip preserves `manual.*` | PASS | `go test ./internal/settings -run 'GitStrategy.*RoundTrip' -count=1 -v` | 0 | 2 | `ac-wbr-014.txt` |
| AC-WBR-015 Unresolvable writes nothing | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Unresolvable' -count=1 -v` | 0 | 2 | `ac-wbr-015.txt` |
| AC-WBR-016 Firing point (**both halves**) | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(Fires\|Registered\|Once\|LinkedWorktree\|NotPrimary)' -count=1 -v` | 0 | 4 | `ac-wbr-016.txt` |

`[no tests to run]` count across every criterion-discharging invocation above: **0**.

### AC-WBR-002 verbatim

```
$ grep -n 'worktree_base_branch' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
12:  worktree_base_branch: ""
$ grep 'worktree_base_branch' <tmpl> | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main
value-side-branch-grep-rc=1 (1 = no match = PASS)
$ grep -c 'worktree_base_branch' internal/config/types.go
1
```

### AC-WBR-016 — both halves, separately

Debt item 7 asks that both halves be treated as separately mandatory. They are asserted by two distinct tests, not one:

- **Half 1** `TestWorktreeBaseBranchFiresExactlyOnceFromPrimaryCheckout` — sub-tests for a set value AND an empty value; the alignment-ENTRY read seam is invoked exactly 1 time in both.
- **Half 2** `TestWorktreeBaseBranchNotFiredFromLinkedWorktree` — the exact configuration AC-WBR-005 requires a write for; entry seam 0, write seam 0, stderr empty, `Handle` nil error.

The seam counted is the alignment-entry (configured-value) read pinned by plan.md §A D3.2 — not the `origin/HEAD` read, which stays a separate seam so half 1 holds on the empty (shipped-default) path.

---

## 2. Milestone commits

| Milestone | SHA | Subject |
|---|---|---|
| M1 | `81808d85b` | `feat(config): add git_strategy.worktree_base_branch schema key and neutral default` |
| M2 | `cf2955ed5` | `feat(hook): align refs/remotes/origin/HEAD from the configured worktree base branch` |
| M3 | `9e1ea4226` | `feat(cli): pass the configured base branch to git worktree add` |
| M4 | `5658988be` | `feat(cli): add the Worktree Base Branch doctor diagnostic` |
| M5 | `04c645a68` | `feat(web): expose worktree_base_branch in the console as a free-text field` |
| M6 | `8c46460ff` | `docs(worktree): document the stored card-worktree base branch and its two consumers` |

Every commit names card **t313** in its body. Staged by explicit pathspec only — no `git add -A`, no `git add .`, no `git commit -a`. `git rev-parse --short HEAD` and `git branch --show-current` were re-read immediately before each commit and matched the assumption every time. **Nothing was pushed and no PR was opened**, per the dispatch.

M1 also carried the `draft → in-progress` transition on `spec.md` (the only status transition this run performs; `plan.md` / `acceptance.md` / `progress.md` carry no YAML frontmatter, so `spec.md` is the whole surface).

---

## 3. Build, vet, lint, parity

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Cross-platform | `GOOS=linux GOARCH=amd64 go build ./...` | exit 0 |
| Vet | `go vet ./internal/{config,hook,cli,settings,web}/...` | exit 0, no output |
| Lint | `golangci-lint run --timeout=5m ./internal/{config,hook,cli,settings,web}/...` | exit 0, `0 issues.` |
| Template embed | `make build` | exit 0; `catalog.yaml` byte-unchanged (12899 bytes), so no hash-regen cascade |
| Mirror parity | `diff -q <template> <local>` on `worktree-integration.md` | exit 0 (byte-identical) |

Two `errcheck` findings (`fmt.Fprintf` return unchecked) were introduced by this run and fixed before the final measurement; the recorded lint run is clean. The `internal/cli` doctor golden snapshots were regenerated (`UPDATE_GOLDEN=1`) because the new diagnostic adds one row — the diff is that row plus column re-padding, nothing else.

**AC-WBR-013 parity, N1-repaired probe** (`BASE = 3cb258d625eb2e095eedd255d5ba2827aa18547e`):

```
.claude/rules/moai/workflow/worktree-integration.md
  -> internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md   (plain form) OK
.moai/config/sections/git-strategy.yaml
  -> internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl          (.tmpl form) OK
NO-TEMPLATE-COUNTERPART lines: 0
DRIFT lines: 0   (check (2) enumerates nothing — this SPEC touches no hook wrapper)
```

---

## 4. Carried debt — disposition

| # | Item | Disposition |
|---|---|---|
| 1 | N1 (AC-WBR-013 probe) | Already repaired in spec 0.3.1. Confirmed in use: the `.tmpl`-form probe is exactly what made `git-strategy.yaml` pass. |
| 2 | N2 (AC-WBR-016 seam denotation) | Already repaired in spec 0.3.1. Implemented as pinned — the entry seam is a separate function-variable from the `origin/HEAD` read, so half 1 holds on the empty path. |
| 3 | Seam obligation (M2/M3) | **DISCHARGED.** `internal/hook/worktree_base_branch.go` is new and introduces five function-variable seams on the `var x = xReal` idiom copied from `internal/cli/session_worktree.go:48-53`, plus a swap-and-restore helper. `internal/cli` gained `sessionWorktreeBaseResolvable`, which delegates at call time to the shared helper. |
| 4 | N4 (check (2) path mixing) | **UNTRIGGERED, UNREPAIRED.** Scope never grew to a hook wrapper, so the check enumerates nothing and the latent path-mixing defect was never exercised. Left for whoever first touches a wrapper. |
| 5 | N5 (AC-WBR-002 human-read) | **DISCHARGED.** `internal/template/worktree_base_branch_neutrality_test.go` asserts value-emptiness mechanically: it strips trailing comments, keeps the value side, and fails on a branch-named value. It also pins `found == 1`, so a deleted or duplicated key fails too. |
| 6 | N3 (two modalities in REQ-WBR-012) | Cosmetic; no action, as the verdict stated. |
| 7 | Folding traceability thinness | **HONOURED.** Both halves of AC-WBR-016 are separate named tests (§1). REQ-WBR-013's preservation clause is asserted by `TestGitStrategyWorktreeBaseBranchRoundTripPreservesManualKeys`, which failed before the fix and passes after — it is not an id-level check. |
| G2 | `EnterWorktree`'s `origin/HEAD` read is inferred | **STILL INFERRED.** Not repairable from this repository — Claude Code is not in it. The stated fallback is now real: the doctor item exists, is registered, and was executed. |
| G4 | `moai doctor` never executed | **DISCHARGED.** Built and run: `./bin/moai doctor --check 'Worktree Base Branch'` → exit 0, reachable by exact name, correctly reporting the match state in this tree (`origin/HEAD` = develop, config = develop). |
| G6 | Rendered attribute order unmeasured | **DISCHARGED — and the criterion's two branches were BOTH wrong.** Measured: the renderer emits `<input class="in" type="text" id="<name>" name="<name>"` (`internal/web/fieldsets_templ.go:1152-1175`), with `id` between `type` and `name`. Neither `type="text" name=` nor `name= type="text"` appears. The assertion is collapsed to the single measured form. |

---

## 5. Scope note — one file outside the plan §D write list

`internal/config/types.go` `ModeProfile` gained three pass-through fields: `develop_branch`, `release_branch_prefix`, `rc_version_format`.

This is stated rather than absorbed, because spec.md §C lists the `ModeProfile` schema gap as out of scope while REQ-WBR-013 requires that this SPEC's write path not drop those keys. The two are only jointly satisfiable by modelling them:

- **Measured, not assumed.** With the FieldDef and applier in place but the fields absent, `TestGitStrategyWorktreeBaseBranchRoundTripPreservesManualKeys` reported `typed save dropped the unmodelled key "develop_branch"` (and the other two) — `m5-settings-1.txt`. The web control would have silently deleted three keys this repository depends on the first time anyone edited the base branch.
- **No merge path exists.** `saveSection` (`internal/config/manager.go:418`) marshals the struct and writes it atomically; there is no merge-with-existing step to preserve unmodelled keys.
- **Routing around it was not available.** REQ-WBR-013 mandates the typed route (`FieldDef` in `gitStrategyFields()`, applied by `applyGitStrategyKey`), so the yamlpatch seam is not an option.

The fields carry no accessor and no consumer; they exist only to survive the round trip, which is the narrowest reading of "shall not newly expose that gap". The wider divergence §C excludes — no accessor, no `ActiveModeProfile()` surface, no behaviour — is untouched. **If the reviewer reads §C as forbidding even the pass-through fields, the correct alternative is to escalate AC-WBR-014 as a blocker and hold M5**, since the criterion's own instruction is to escalate rather than absorb; that is the objection, recorded here for the decision to be made rather than made silently.

### Three further cascade edits, all inside the SPEC's envelope

| File | Why |
|---|---|
| `internal/config/testdata/shipped_key_inventory.yaml` | A pre-existing anti-rot guard FAILS on any new shipped template key that is not triaged. Added as class `W` (wired, has a reader). |
| `internal/web/widget_policy_test.go` | `TestFreeTextWhitelist` FAILS on any new `TypeText` field. The entry records the operator ruling and its reason (a branch name has no closed domain), not a bare exemption. |
| `internal/cli/testdata/doctor-*.golden` (3) | The new diagnostic adds one row; regenerated with `UPDATE_GOLDEN=1`. |

---

## 6. Gaps — what was NOT verified

Stated explicitly. An empty Gaps section would itself be a claim.

1. **The end-to-end premise is still inferred (G2).** No test in this change proves that Claude Code's `EnterWorktree` reads `refs/remotes/origin/HEAD`. Consumer 1 is verified to WRITE the symref correctly; that the runtime then reads it rests on the observed base of this worktree plus the documented `fresh` semantics. If a future runtime changes that, consumer 1 silently stops mattering and only the doctor item would surface it.
2. **Consumer 1's happy path was never exercised against a real repository.** This session runs inside a linked worktree, so the primary-checkout gate returns false here by construction. Every consumer-1 assertion is against fakes. Consumer 2 IS exercised against real git (`TestSessionWorktreeConfiguredBaseCutsFromThatBranch` runs a real `git worktree add` and asserts the created tree's HEAD equals the base branch's commit).
3. **No push, no CI verdict.** Nothing was pushed and no PR was opened, per the dispatch. The full-suite verdict on a clean machine is unmeasured; `go test ./...` was deliberately not run locally. Package-scoped runs of all five affected packages pass.
4. **Coverage figures are package baselines, not deltas.** No pre-change baseline was captured, so no coverage delta is claimed. `internal/web` (66.8%) and `internal/cli` (79.6%) sit below the 85% target — both are pre-existing package-level figures this change did not measurably move, not a regression this run introduced.
5. **N4 remains latent** (§4 row 4).
6. **AC-WBR-012's literal command form.** The criterion writes `./internal/web/... ./internal/hook/... ./internal/cli/...`; the recorded discharging run uses the narrowed `./internal/web ./internal/hook ./internal/cli` so that §D.3's `[no tests to run]` pin is meaningful. The literal `/...` form was ALSO run (`ac-wbr-012-literal-form.txt`): exit 0, 21 `=== RUN`, and 52 `[no tests to run]` lines — every one from a sub-package that contains no matching test, none from a package where the guard should have run. Same 21 tests either way.
7. **`moai doctor` non-OK paths were exercised only through unit tests.** The CLI run observed the OK/match state; the two non-OK states and the no-origin state are covered by `TestDoctorWorktreeBaseBranchFourStates` against fakes, not by a live `moai doctor` invocation in a mismatched repository.

---

## 7. Residual risk

- **`refs/remotes/origin/HEAD` is repository-global and this SPEC now writes it automatically.** The primary-checkout gate removes multi-lane contention, but a second primary checkout of the same repository, or an external actor moving the symref by hand, still produces one attributable notice line on the next primary-checkout session start. That is the designed outcome, not a fight — only one automatic writer exists.
- **The setting lives inside the `moai update` wipe root.** `.moai/config` is restored to template defaults by `moai update`, which by REQ-WBR-003 means the empty neutral value. The failure direction is "does nothing", not "wrong branch". Inherited, not introduced.
- **The pass-through `ModeProfile` fields are load-bearing but invisible.** Nothing reads them; a future author tidying up "unused" fields would silently restore the key-dropping defect. The comment on them says so, but a comment is the only guard — no test asserts the fields exist by name, only that the round trip preserves the keys.
- **A first session in a fresh clone still gets the old base.** The symref is local metadata, so the alignment cannot run before the first session start in that clone. Inherent to the handle, stated in plan §A D1.

---

## 8. Blockers

**None.** No blocker report is outstanding. The one scope tension (§5) is recorded as an objection with the alternative named, not raised as a blocker, because the requirement layer (REQ-WBR-013, MUST) and the criterion (AC-WBR-014, MUST) both demand preservation and the implemented change is the narrowest way to provide it.
