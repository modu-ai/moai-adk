# SPEC-WORKTREE-BASEREF-001 — Sync-Phase Audit (card t313)

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t313` · branch `WT-worktree-baseref` · HEAD `023f7fd57`
Base: `d0bc4bba5` (plan artifacts) · merge-base with `origin/develop`: `3cb258d625eb2e095eedd255d5ba2827aa18547e`
Auditor: `sync-auditor`, independent. The implementer's `run-verdict.md` was treated as the claim under test; every AC below was re-run in this tree, not read off that file.
Evaluator profile: `default` (SPEC frontmatter carries no `evaluator_profile`; `harness.default_profile: "default"`). Weights 40/25/20/15; must-pass = Functionality + Security; hard threshold Craft coverage >= 85%.

---

## Verdict

**PASS — mergeable into `develop` as it stands.**

Overall score **0.90** (harmonic mean of the four dimension scores). Must-pass firewall: Functionality 0.95 PASS, Security 1.00 PASS — both clear independently.

| Dimension | Weight | Score | Verdict | Evidence (re-run in this tree) |
|---|---|---|---|---|
| Functionality | 40% | 0.95 | **PASS** | 16/16 MUST criteria reproduced independently; all five touched package trees green in full (`go test ./internal/{config,settings,web,hook,cli}/... -count=1` all rc 0) |
| Security | 25% | 1.00 | **PASS** | No Critical/High. Argument-vector review of the three `exec.Command` call sites: no shell, no interpolation, leading-`-` values gated by the resolvability predicate. One informational note (F5) |
| Craft | 20% | 0.80 | PASS-with-note | New code is well covered (`RunWorktreeBaseAlignment` 100.0%, `gitWorktreeAddArgs` 100.0%, `checkWorktreeBaseBranch` 90.0%); package aggregates sit below 85% on pre-existing baselines, not on this delta (F2) |
| Consistency | 15% | 0.90 | **PASS** | Seam idiom, fail-open shape, template-first parity, 4-locale i18n all match the surrounding conventions; `golangci-lint run` on the five packages → `0 issues.` |

Harmonic mean = 4 / (1/0.95 + 1/1.00 + 1/0.80 + 1/0.90) = **0.9048 → 0.90**, above the 0.80 line.

Nothing in the finding list below blocks a merge. Findings F1-F6 are recorded debt.

---

## A1 — Per-AC independent reproduction (16/16)

Every criterion's own command was re-executed. `=== RUN` counts are from my runs, not the implementer's captures.

| AC | Command re-run | `=== RUN` | rc | Verdict |
|---|---|---|---|---|
| 001 | `go test ./internal/config/... -run 'GitStrategy' -count=1 -v` + `grep -n worktree_base_branch internal/config/types.go` | 35 | 0 | PASS (with F1) |
| 002 | value-side grep on `git-strategy.yaml.tmpl` | n/a | grep rc **1** (no branch name on the value side) | PASS |
| 003 | `-run 'WorktreeBaseBranch.*(Unset\|Empty)'` → `TestWorktreeBaseBranchAlignmentUnsetIsSilentNoOp` | 1 | 0 | PASS |
| 004 | `-run 'WorktreeBaseBranch.*Match'` → `…AlignmentMatchIsSilentNoOp` | 1 | 0 | PASS |
| 005 | `-run 'WorktreeBaseBranch.*Mismatch'` → `…AlignmentMismatchWritesAndAnnounces` | 1 | 0 | PASS |
| 006 | `-run 'WorktreeBaseBranch.*(FailOpen\|GitError)'` → `…FailOpenOnGitError`, `…FailOpenOnOriginHeadError` | 2 | 0 | PASS |
| 007 | `go test ./internal/cli/... -run 'SessionWorktree.*Base' -timeout 600s` → incl. `TestSessionWorktreeConfiguredBaseCutsFromThatBranch` | 5 | 0 | PASS |
| 008 | `-run 'SessionWorktree.*(NoBase\|Unresolvable)'` | 3 | 0 | PASS |
| 008 (3rd assertion) | `-run 'SessionWorktree.*(SharedPredicate\|Resolver)'` → `…SharedPredicateResolverInvokedOnce`, `…SharedPredicateDefaultsToConsumerOnesHelper` | 2 | 0 | PASS |
| 009 | `-run 'Doctor.*WorktreeBaseBranch'` (4 named sub-states) + `./bin/moai doctor --check 'Worktree Base Branch'` | 7 | 0 | PASS |
| 010 | `./internal/web/... -run 'WorktreeBaseBranch'` (2) + `./internal/settings/... -run 'AllFields'` (1) | 2 / 1 | 0 | PASS |
| 011 | `./internal/settings/... -run 'WorktreeBaseBranch.*Type'` + `./internal/web/... -run 'WorktreeBaseBranch.*(Text\|FreeText)'` | 1 / 1 | 0 | PASS |
| 012 | `./internal/web/... ./internal/hook/... ./internal/cli/... -run 'WorktreeBaseBranch' -timeout 600s` | 21 | 0, 0 FAIL | PASS |
| 013 | template-counterpart + `.sh`/`.sh.tmpl` twin checks, this diff only | n/a | 0 NO-TEMPLATE-COUNTERPART, 0 DRIFT | PASS (with F3) |
| 014 | `./internal/settings/... -run 'GitStrategy.*RoundTrip'` → `TestGitStrategyWorktreeBaseBranchRoundTripPreservesManualKeys` | 1 | 0 | PASS (with F4) |
| 015 | `-run 'WorktreeBaseBranch.*Unresolvable'` → 2 tests | 2 | 0 | PASS |
| 016 | `-run 'WorktreeBaseBranch.*(Fires\|Registered\|Once\|LinkedWorktree\|NotPrimary)'` | 4 | 0 | PASS |

**No AC failed to reproduce.** No contradiction with the orchestrator's `go build` / `go vet` measurements: I re-ran `go vet ./internal/{config,hook,cli,settings,web}/...` → rc 0.

Full-suite regression check beyond the `-run` filters (the risk a filtered run cannot see):

```
go test ./internal/config/... ./internal/settings/... -count=1              rc=0
go test ./internal/web/...    ./internal/hook/...     -count=1              rc=0
go test ./internal/cli/...    -count=1 -timeout 900s                        rc=0
golangci-lint run ./internal/{config,hook,cli,settings,web}/...   → 0 issues.
```

The three regenerated doctor golden files (`doctor-{dark,light,nocolor}.golden`) are covered by that green `internal/cli` run.

---

## A2 — The vacuous-test trap

Nine criteria are discharged by a `-run` regex. For each, I re-ran with `-v` and counted `^=== RUN` myself. Counts are in the A1 table; **every one is >= 1, and the executed test names match the criterion's subject** (they are enumerated above rather than summarised, because a non-zero count against an unrelated test would be the same trap one level down).

There is **no vacuous pass** in this set. AC-WBR-011's schema half — the specific invocation `acceptance.md` §D.3 measured as `[no tests to run]` at plan time — now executes `TestWorktreeBaseBranchFieldTypeIsText`.

**F6 (criterion-authoring artifact, non-blocking).** §D.3's second mechanical clause, `grep -q '\[no tests to run\]' out.txt && echo "VACUOUS — criterion FAILS"`, fails **every** multi-package invocation in this file, including correct ones: `./internal/hook/...` expands to sibling packages (`hook/testutil`, `hook/trace`, …) that legitimately match no test and each print `[no tests to run]`. My runs record 10 such lines per hook invocation while the criterion's own test executes. The operative clause is the `=== RUN` count, and I applied it that way. If the string clause is meant to bind literally, it must be scoped to the owning package (`… | grep '^ok .*internal/hook\b' | grep -q 'no tests to run'`) — otherwise it will mislabel a correct run in every future SPEC that copies this stanza.

---

## A3 — The shared resolvability predicate (REQ-WBR-011)

**Verified at the code level, not behaviourally.** There is exactly one implementation.

- The sole authority: `internal/hook/worktree_base_branch.go:83` — `var WorktreeBaseBranchResolvable = worktreeBaseBranchResolvableReal`.
- Consumer 2 delegates **at call time**, holding no rule of its own: `internal/cli/session_worktree.go:67-69` —
  `sessionWorktreeBaseResolvable = func(branch string) bool { return hook.WorktreeBaseBranchResolvable(branch) }`.
  Call-time delegation (not a snapshot at package-init) is what makes the test seam and the production path the same rule.
- The doctor diagnostic delegates the same way: `internal/cli/doctor_worktree_base.go:25-27`.
- **No second rule exists.** `grep -rn` over `internal/` finds no `rev-parse --verify`, no `git branch --list`, and no local-branch check on this path. Consumer 2's only resolvability call site is `session_worktree.go:216`.

Exit-code discipline (`internal/cli/session_worktree.go` → `internal/hook/worktree_base_branch.go:185-190`):

```go
return exec.Command("git", "show-ref", "--verify", "refs/remotes/origin/"+branch).Run() == nil
```

Only `rc == 0` counts as resolvable — `.Run() == nil` is exactly that predicate, so a missing ref is correctly classified. Measured in this tree:

```
git show-ref --verify refs/remotes/origin/develop        → rc=0
git show-ref --verify refs/remotes/origin/no-such-branch → rc=128
```

An `rc == 1` implementation would have misclassified the second. The plumbing form is used, so BranchGuard's `\bgit\s+branch\b` refusal is avoided. **PASS.**

---

## A4 — The read seam and the primary-checkout gate

Both halves of AC-WBR-016 landed, and both go through `Handle` rather than calling the helper directly — which is what makes the seam count assert *registration*, the thing the criterion exists for.

- **Seam is a function variable**: `worktreeBaseBranchReadConfig = worktreeBaseBranchReadConfigReal` (`internal/hook/worktree_base_branch.go:52`), documented in place as the alignment-entry seam AC-WBR-016 counts.
- **Registered in the errgroup**: `internal/hook/session_start.go:175-182`, Task 5, `g.Go(func() error { worktreeBaseData = RunWorktreeBaseAlignment(input.ProjectDir); … })`.
- **Gate precedes the read** — `internal/hook/worktree_base_branch.go:91-100`:

```go
if !WorktreeBaseBranchInPrimaryCheckout() { return data }   // gate
configured := worktreeBaseBranchReadConfig(projectRoot)      // read
```

  So the entry-seam count from a linked worktree is structurally 0, not incidentally 0.
- **Half 1** (`worktree_base_branch_test.go:238-262`): sub-tests `set value` and `empty value`, both asserting `f.entryCalls == 1` after one `Handle`.
- **Half 2** (`:266-291`): `primary: false` with the exact AC-WBR-005 write-requiring state, asserting all four — `entryCalls == 0`, `setHeadCalls == 0`, empty stderr, nil error.

The discriminant's both directions were measured by hand: in this linked worktree `--git-dir` = `…/.git/worktrees/t313` vs `--git-common-dir` = `…/.git` (differ); in a throwaway clone both print `.git` (equal). **PASS.**

---

## A5 — The ModeProfile pass-through ruling (implementation, not the ruling)

**Implemented as ruled: genuinely pass-through, no scope creep.**

`internal/config/types.go:126-128` adds exactly three struct tags and nothing else:

```go
DevelopBranch       string `yaml:"develop_branch"`        // manual mode, git-flow only
ReleaseBranchPrefix string `yaml:"release_branch_prefix"` // manual mode, git-flow only
RCVersionFormat     string `yaml:"rc_version_format"`     // manual mode, git-flow only
```

Repository-wide reference count for each of the three identifiers, over `internal/`, `pkg/`, `cmd/`:

```
DevelopBranch       → internal/config/types.go:126   (1 hit — the declaration)
ReleaseBranchPrefix → internal/config/types.go:127   (1 hit)
RCVersionFormat     → internal/config/types.go:128   (1 hit)
```

**No accessor, no consumer, no reader, no writer, no validation, no `ActiveModeProfile()` change.** The only behaviour they can affect is yaml round-trip survival, which is precisely the ruled minimum. `spec.md` §C's out-of-scope exclusion (repairing the wider ModeProfile schema gap) is intact. `shipped_key_inventory.yaml` gained the corresponding 3 entries — the inventory guard's own bookkeeping, not new behaviour.

**AC-WBR-014 proves the round trip — but one notch weaker than the property (F4).** The write path is `LoadRaw → applyGitStrategyKey → SetSection → Save` (`internal/settings/sectionapply.go:88-149`, `:170-181`), i.e. a full struct re-marshal. The test (`internal/settings/worktree_base_branch_test.go:94-98`) asserts the three **key names** survive:

```go
for _, key := range []string{"develop_branch", "release_branch_prefix", "rc_version_format"} {
    if !strings.Contains(written, key) { t.Errorf("typed save dropped the unmodelled key %q…") }
}
```

A regression that preserved the keys while emptying their values (`develop_branch: ""`) would pass. The values do in fact survive — the loader populates the fields and the marshal writes them back — so this is a **guard-strength gap, not a live defect**. The real-world corroboration is that `.moai/config/sections/git-strategy.yaml` in this tree still carries `develop_branch: develop`, `release_branch_prefix: release/`, `rc_version_format: vX.Y.Z-rc.N` after the SPEC's own edit, whose diff is a single added line.

---

## A6 — The fake-only coverage of consumer 1: **CLOSED empirically, not by argument**

The implementer's disclosure is accurate as far as the unit tests go: in this session the primary-checkout gate returns false by construction, so every in-package consumer-1 assertion runs against fakes, and the three real seams show 0.0% statement coverage:

```
worktree_base_branch.go:88  RunWorktreeBaseAlignment              100.0%
worktree_base_branch.go:144 worktreeBaseBranchInPrimaryCheckoutReal 80.0%
worktree_base_branch.go:155 worktreeBaseBranchReadConfigReal         0.0%
worktree_base_branch.go:161 worktreeBaseBranchReadOriginHeadReal     0.0%
worktree_base_branch.go:170 worktreeBaseBranchSetHeadReal            0.0%
worktree_base_branch.go:185 worktreeBaseBranchResolvableReal        66.7%
```

**My independent read: the fake-only coverage was not sufficient to trust consumer 1 in production, so I ran the missing verification rather than recommending it.** It passes.

I built a throwaway repository (bare origin + clone, under the session scratchpad — this repository was never mutated), seeded a `git-strategy.yaml`, and drove the shipped binary's real SessionStart. A plain clone **is** a primary checkout, so the gate returns true and the whole path executes against real git.

**(a) happy path — mismatch → write + exactly one notice:**

```
$ echo '{"session_id":"audit-t313","cwd":"…/work","project_dir":"…/work"}' | ./moai hook session-start
rc=0
stderr: notice: refs/remotes/origin/HEAD realigned from main to develop per git_strategy.worktree_base_branch
$ git symbolic-ref refs/remotes/origin/HEAD
refs/remotes/origin/develop          # was refs/remotes/origin/main
```

**(b) idempotence — second run is a silent no-op:** re-run with the same config; **no alignment line on stderr** (the only stderr content was an unrelated multi-session `<system-reminder>` from a different SessionStart task), symref unchanged.

**(c) unresolvable — one diagnostic, no write:** with `worktree_base_branch: no-such-branch`,

```
rc=0
warning: git_strategy.worktree_base_branch names "no-such-branch", which has no remote-tracking
branch (refs/remotes/origin/no-such-branch); refs/remotes/origin/HEAD left unchanged — correct the setting
$ git symbolic-ref refs/remotes/origin/HEAD
refs/remotes/origin/develop          # unchanged
```

**(d) the read path in THIS real repository:** `./bin/moai doctor --check 'Worktree Base Branch'` →
`ok  Worktree Base Branch  refs/remotes/origin/HEAD names develop, matching git_strategy.worktree_base_branch`, rc 0. That exercises the real config read and the real `git symbolic-ref` against the actual repository.

**(e) the `set-head` primitive in isolation:** `git remote set-head origin develop` in the throwaway clone → rc 0, symref `main` → `develop`, exactly the contract `worktreeBaseBranchSetHeadReal` assumes.

All four real seams are therefore exercised end to end against a real repository, in all three interesting states. **No pre-merge verification remains outstanding on this axis.** The residual is now only that these runs are manual and captured in this report rather than in CI — recorded as F2.

---

## A7 — Template neutrality and re-embedding

**Neutral.** `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:12` ships `worktree_base_branch: ""`. The value-side probe:

```
grep 'worktree_base_branch' …/git-strategy.yaml.tmpl | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main
→ rc=1   (no match; a value-bearing hit would FAIL)
```

The comment names `main, develop, trunk` as examples — permitted by AC-WBR-002, which binds the value only. Mechanically pinned by `internal/template/worktree_base_branch_neutrality_test.go` → `TestWorktreeBaseBranchTemplateValueIsEmpty`, which I ran (rc 0, 1 `=== RUN`); this is N5's discharge.

The local `.moai/config/sections/git-strategy.yaml` carries `worktree_base_branch: develop`, which is the permitted repository-specific side.

**Re-embedded.** `//go:embed all:templates` compiles the template tree in at build time, so the source `.tmpl` is the single source of truth; `make build` produced `bin/moai` (68,955,874 bytes, 23:09) and the key is present in it: `strings bin/moai | grep -c worktree_base_branch` → **28**. The binary I used for the A6 runs is that binary.

---

## A8 — Carried-debt disposition (7 items + inherited)

| Item | Implementer's claim | My verdict | Evidence |
|---|---|---|---|
| **N1** — AC-013 check (1) false NO-TEMPLATE-COUNTERPART | repaired | **Confirmed** | The `.tmpl`-accepting probe was used; the two shipped surfaces resolve (`worktree-integration.md` plain form, `git-strategy.yaml.tmpl` form). 0 lines emitted. See F3 for a second, disclosed deviation |
| **N2** — AC-016 seam denotation | pinned pre-run | **Confirmed** | `acceptance.md:289-292` pins the alignment-entry (configured-value) read; the code names it as such at `worktree_base_branch.go:47-56` and the test counts that seam |
| **Seam obligation** (item 3) | discharged | **Confirmed** | `internal/hook` now carries its own seam block (`worktree_base_branch.go:38-70`), three members exported precisely so the cross-package guard can swap them |
| **N4** — check (2) dormant | left latent | **Confirmed correct** | `git diff --name-only BASE..HEAD -- '*/hooks/moai/*.sh' '*/hooks/moai/*.sh.tmpl'` → empty. Vacuous by construction, as the criterion itself intends. Repairing an untriggered check would have been scope creep |
| **N5** — AC-002's first command human-read | discharged | **Confirmed** | `internal/template/worktree_base_branch_neutrality_test.go` makes it mechanical |
| **N3** — REQ-012 two modalities | cosmetic, no action | **Confirmed** | No run-phase action was required |
| **Item 7** — both AC-016 halves separately mandatory | honoured | **Confirmed** | Two distinct top-level tests, neither subsumed by the other |
| **G2** — `EnterWorktree`'s origin/HEAD read inferred | fallback landed | **Confirmed** | The doctor item is the stated surfacing mechanism and it exists and runs |
| **G4** — `moai doctor` never executed | discharged | **Confirmed twice** | Implementer's `ac-wbr-009-doctor-run.txt`, and my own independent `./bin/moai doctor --check 'Worktree Base Branch'` above |
| **G6** — rendered attribute order unmeasured | both either-or branches WRONG; collapsed to measured form | **Confirmed — the claim is correct** | The renderer emits `<input class="in" type="text" id="…" name="…"`, with `id` **between** `type` and `name`. Neither `type="text" name="X"` nor `name="X" type="text"` occurs, so both branches of the plan-time condition would have failed a correct implementation. `TestWorktreeBaseBranchRendersFreeTextInput` asserts the single measured form and passes against the real renderer in my run — the measurement is empirically confirmed, not taken on trust |

---

## Findings

Format: `id [severity] [blocking|optional] file:line — description — what goes wrong if it ships`.

- **F1 [MINOR] [optional]** `internal/config/types.go:121,174` — AC-WBR-001's second command says "expect exactly 1 hit" for `grep -n worktree_base_branch internal/config/types.go`; the tree returns **2** (line 174 is the field, line 121 is a doc-comment mention added by the M5 pass-through comment block). Literal reading of the criterion FAILS; substantive reading PASSes, since the criterion's intent is one schema declaration. *If it ships:* nothing at runtime. A future re-run of the AC verbatim reads red on a correct tree. **Required fix (deferrable):** narrow the criterion to `grep -c '^\s*WorktreeBaseBranch string' internal/config/types.go` → 1, or accept the 2 and say why.

- **F2 [MINOR] [optional]** coverage — package aggregates are below the profile's 85% Craft threshold: `config` 80.6%, `cli` 79.6%, `web` 66.8% (`hook` 85.1%, `settings` 90.3% clear it). These are **pre-existing package baselines, not this SPEC's delta**: the code this SPEC added is well covered (`RunWorktreeBaseAlignment` 100.0%, `gitWorktreeAddArgs` 100.0%, `checkWorktreeBaseBranch` 90.0%). What is genuinely uncovered by automated tests are the four real git-wrapper seams (0.0%-66.7%), which I exercised **by hand** in A6 rather than in CI. *If it ships:* a future regression in `worktreeBaseBranchSetHeadReal` or `…ReadOriginHeadReal` is caught by no test. **Required fix (deferrable):** one `testing.Short()`-skipped integration test that builds a throwaway bare-origin clone and asserts the three A6 outcomes; that would convert my manual runs into a standing guard. Scored 0.80 rather than FAIL because the hard threshold is aimed at the change's own craft, and no aggregate moved backwards.

- **F3 [MINOR] [optional]** `acceptance.md:243` — AC-WBR-013 check (1) excludes only `^.moai/specs/`, so this diff's four `.moai/reports/t313/*.md` evidence files emit false `NO-TEMPLATE-COUNTERPART` lines. The implementer added `grep -v '^.moai/reports/'` and **disclosed the deviation in full** in `ac-wbr-013-parity.txt` rather than hiding it. I reproduced both forms and agree with the repair. *If it ships:* the same false failure recurs for every SPEC that writes report evidence. **Required fix (deferrable):** fold `^.moai/reports/` into the criterion's exclusion, alongside N1's `.tmpl` repair.

- **F4 [MINOR] [optional]** `internal/settings/worktree_base_branch_test.go:94-98` — AC-WBR-014's guard asserts the three unmodelled key **names** are present in the written file, not that their **values** survived. A regression writing `develop_branch: ""` passes the guard. The values do survive today (verified by reading the `LoadRaw → SetSection → Save` path and by the local `git-strategy.yaml` being intact after this SPEC's one-line edit), so this is a guard-strength gap. *If it ships:* a future ModeProfile change could silently blank three git-flow keys this repository depends on and the guard would stay green. **Required fix (deferrable):** assert the full `key: value` pairs (`develop_branch: develop`, `release_branch_prefix: release/`, `rc_version_format: vX.Y.Z-rc.N`) rather than the bare key names.

- **F5 [INFORMATIONAL] [optional]** `internal/hook/worktree_base_branch.go:171,190`, `internal/cli/session_worktree.go:250-254` — no `--` end-of-options separator on the three git invocations that carry the free-text value. `git show-ref --verify refs/remotes/origin/<v>` is inherently safe (the value is concatenated into a ref path, so a leading `-` cannot become an option). `git remote set-head origin <v>` and `git worktree add -b <b> <dest> <v>` would accept an option-shaped value in principle, **but both are gated behind the resolvability predicate**, so a value like `-d` or `--detach` is only reachable if a remote-tracking ref of that literal name exists. Every call uses `exec.Command` with a separate argument vector — no shell, no interpolation, so there is no command injection surface. *If it ships:* nothing exploitable. **Required fix (optional hardening):** add `--` before the branch operand in the two option-position calls.

- **F6 [MINOR] [optional]** `acceptance.md:335` — §D.3's `grep -q '\[no tests to run\]'` VACUOUS clause is over-broad; it fails every multi-package invocation in the file (10 such lines per `./internal/hook/...` run in my measurements) including correct ones. Detailed in A2. *If it ships:* future SPECs copying this stanza will mislabel correct runs as vacuous. **Required fix (deferrable):** scope the string check to the owning package, or drop it and keep the `=== RUN` count as the sole mechanical clause.

**Blocking findings: none.**

---

## Coverage statement

**Checked.** All 16 AC commands re-executed in this tree with `-v` and `=== RUN` counted independently; full test suites for all five touched package trees (`config`, `settings`, `web`, `hook`, `cli`) run to completion; `go vet` over the same five; `golangci-lint run` over the same five; the shared-predicate contract read at the code level across all three call sites with a repository-wide grep for competing implementations; the errgroup registration and gate ordering read at source; the ModeProfile pass-through claim verified by a repository-wide reference count; template neutrality verified by the value-side probe, by the neutrality test, and by `strings` on the built binary; template-first parity and the `.sh`/`.sh.tmpl` twin check re-derived from the diff; the card t316 boundary confirmed empty; and — the item the implementer flagged as its own residual risk — consumer 1's three real states driven end to end against a real repository through the shipped binary.

**Not checked (gaps, stated so they are not mistaken for passes).**

1. **The AC-WBR-012 mutation was not re-performed by me.** Re-running it requires deleting the `FieldDef` from `internal/settings/schema_sections.go`, which is a source edit this audit is forbidden from making. I read the implementer's capture (`ac-wbr-012-mutation.txt`): with the field removed, guard properties (1) and (2) fail and `MUTATED-RC=1`. The capture is internally consistent with the guard's source, and property (3) correctly still passes there (removing the schema entry does not break the config→consumer path, and the conjunction fails as a whole). I judge the claim credible but it is **attributed to the implementer, not independently reproduced**.
2. **Cross-platform.** All measurement is darwin/arm64. No Windows or Linux run; `git rev-parse --git-common-dir` path comparison and `git worktree add` operand handling were not exercised there. CI covers this on the PR head.
3. **Concurrency.** Two simultaneous SessionStarts against one repository were not run; the design argument (the second finds a match and no-ops, and the primary-checkout gate leaves exactly one writer) was read but not stress-tested.
4. **The web console was not driven through a browser.** The render half is asserted against `renderConsolePage(t)` output; no end-to-end form submit of the free-text field was performed, so the POST → `applyGitStrategyKey` → file path is covered by the round-trip unit test only.
5. **`EnterWorktree`'s actual read of `refs/remotes/origin/HEAD`** remains inferred from behaviour rather than read from Claude Code source — inherited gap G2, unchanged by this SPEC and explicitly mitigated by the doctor item.

## Residual risk

- The four real git-wrapper seams have no standing automated guard (F2). My A6 runs prove they work today; nothing prevents a silent regression tomorrow.
- The value-preservation property of AC-WBR-014 rests on a key-name assertion (F4). The property holds; the guard is weaker than the property.
- `worktree_base_branch` is repository-global metadata written by one process. The primary-checkout gate confines writers to one, but an external actor (another tool, a manual `git remote set-head`, a fresh clone) can still move `refs/remotes/origin/HEAD` between sessions. The design accepts this — realignment is idempotent and costs one notice line — and the doctor item surfaces it. This is inherent to the approach, not a defect in the implementation.
- Two SPEC-artifact housekeeping items sit outside this audit's write permission and belong to the sync phase: `spec.md` frontmatter still reads `status: in-progress`, and this audit's own report file is untracked.

## Working-tree note (this audit's own side effects)

Running `go test ./internal/hook/...` rewrote `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/{baseline,postchange}.md` — a known behaviour of that package's perf fixtures, unrelated to this SPEC. I restored both with `git restore`; `git status --short` is clean again apart from this report file. My throwaway git repository lives entirely under the session scratchpad; **this repository's `refs/remotes/origin/HEAD` was never written.**

---

## Mergeable?

**Yes — merge into `develop` as it stands.** The 16 MUST criteria reproduce independently, the five touched package trees are green in full, the two must-pass dimensions clear their thresholds on their own, and the one residual risk the implementer flagged for a merge decision (A6) is now closed by direct real-repository measurement rather than left open.

The six findings are all recorded debt. If any is worth a follow-up card, the ones with the most future leverage are **F4** (the guard that would not notice three config keys being blanked) and **F2** (no standing test over the real git seams); **F1**, **F3** and **F6** are criterion-text repairs worth folding into whichever SPEC next copies these stanzas.
