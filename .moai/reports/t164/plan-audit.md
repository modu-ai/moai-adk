# SPEC Review Report: SPEC-CLI-STATE-DIR-BOUND-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.60** (harmonic mean; arithmetic mean 0.625). Tier M PASS threshold = 0.80.

Reasoning context ignored per M1 Context Isolation. Only the three artifacts and the actual tree were read.

---

## §1 Claim — the six load-bearing claims the lead asked me to verify

| # | SPEC claim | Verdict |
|---|---|---|
| 1 | `internal/cli/state.go:231` `findStateDirFrom` is an unbounded walk to filesystem root | **TRUE** |
| 2 | `internal/core/project/root.go:20` `FindProjectRoot` carries a `paths.Home()` guard and errors rather than climbing through home | **TRUE** |
| 3 | `internal/cli` already imports `internal/core/project` (no cycle introduced) | **TRUE** |
| 4 | All six consumers exist at the cited line numbers | **TRUE** (all six exact) |
| 5 | `TestFindStateDirFromWalksUp` pins the OPPOSITE contract and would fail under the recommended change | **FALSE** — see D1 |
| 6 | `internal/cli/harness.go:101` `resolveProjectRoot` is the flag-then-cwd convention | **TRUE** |

---

## §2 Evidence — verbatim

**Claim 1 — `sed -n '212,250p' internal/cli/state.go`:**
```
212:func findStateDir() (string, error) {
231:func findStateDirFrom(start string) (string, error) {
	dir := start
	for {
		stateDir := filepath.Join(dir, ".moai", "state")
		if info, err := os.Stat(stateDir); err == nil && info.IsDir() { return stateDir, nil }
		parent := filepath.Dir(dir)
		if parent == dir { break }     // reached filesystem root
		dir = parent
	}
	return "", fmt.Errorf(".moai/state/ directory not found from %s", start)
```
Comment verbatim: `The walk is bottom-up and unbounded ... inherits any ~/.moai/state on the machine`. Confirmed.

**Claim 2 — `sed -n '20,66p' internal/core/project/root.go`:**
```go
homeDir, _ := paths.Home()
...
for {
    if homeDir != "" && absDir == homeDir {
        return "", fmt.Errorf("not in a MoAI project (no .moai directory found in project directories)")
    }
    moaiPath := filepath.Join(absDir, ".moai")
    ...
}
```
Confirmed: `paths.Home()` + `EvalSymlinks` normalization, errors at the home boundary rather than climbing through it. `FindProjectRoot` is at line 20.

**Claim 3 — `grep -rl "internal/core/project" internal/cli/`:**
```
internal/cli/init_autonomy_wizard.go   internal/cli/reporter.go
internal/cli/init_workflow_flags.go    internal/cli/update_template_sync.go
internal/cli/init.go                   (+ 3 _test.go)
```
5 non-test files. No cycle introduced. Confirmed.

**Claim 4 — `grep -rn "findStateDir" internal/`:**
```
internal/cli/clean.go:65    stateDir, err := findStateDir()
internal/cli/tokens.go:377      if dir, err := findStateDir(); err == nil {
internal/cli/chain.go:67            stateDir, err := findStateDir()
internal/cli/chain.go:353   stateDir, err := findStateDir()
internal/cli/state.go:78    stateDir, err := findStateDir()
internal/cli/state.go:154   stateDir, err := findStateDir()
```
All six exact. **One further dependent the SPEC never names:** `internal/cli/state_m2_test.go:37` — `m2SetupState` chdirs a test into a temp tree "so findStateDir resolves it". Not in the SPEC's consumer table, not in plan.md §F M4's file list.

**Claim 5 — `internal/cli/tokens_state_dir_test.go` (read in full).** See D1 below; the premise is false.

**Claim 6 — `sed -n '101,117p' internal/cli/harness.go`:** `resolveProjectRoot` = `--project-root` flag → inherited flag → `os.Getwd()`. No walk. Confirmed.

**D7 (MP-5):** `grep -n "^status:" .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001/spec.md` → `status: completed`. Not retired/superseded/archived. No BLOCKING.
**D8 (MP-6):** `grep -c syscall spec.md` → `0`. Auto-PASS.
**MP-7:** `grep -rn 'NEEDS CLARIFICATION' <spec dir>` → exit 1, no matches.

---

## §3 Baseline-attribution

All evidence gathered in this run, against this tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t164`, branch `main`. No file in the tree was modified by this audit. The symlink probe (D3) was run as `python3 -c "import os,tempfile;..."` in this session on this machine (darwin 25.5.0).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-1…REQ-6 at `spec.md:89,95,101,107,111,117`. Sequential, no gaps, no duplicates, uniformly unpadded.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`), not the AC layer. All six carry an explicit trigger + `SHALL` / `SHALL NOT`: REQ-1/2/3/4 are When-shaped event-driven; REQ-5 is When-shaped; REQ-6 is ubiquitous + unwanted. Given-When-Then in `acceptance.md` is the correct verification-layer format and is graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present at `spec.md:2-15`, canonical names (`created`/`updated`/`tags`/`id`), no rejected snake_case alias. Three cosmetic deviations recorded as minor (D8-D10 below), none of which the decoder rejects.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/cli`). Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one referenced SPEC (`SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001`), exists, `status: completed`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears 0 times in the SPEC body. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — no `[NEEDS CLARIFICATION]` markers in `plan.md`; `research.md` absent (Tier M).

**No must-pass failure. The FAIL is driven by the rubric score (0.60 < 0.80) and by three blocking correctness defects below.**

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Five of six REQs have a single interpretation. REQ-1's hoist (`spec.md:91-93`) leaves the `chain.go` target path genuinely under-determined — two engineers produce two different paths (D2). |
| Completeness | 0.50 | 0.50 | `spec.md:134` claims exhaustively "**동작 변경 1건 명시**: REQ-3". The claim is false: two further behavior changes are undeclared (D2 chain path, D3 symlink normalization). An exhaustive inventory that is not exhaustive is a completeness failure, not a nitpick. |
| Testability | 0.50 | 0.50 | AC-008 is vacuous (D4). AC-003 first command and AC-007 have no binary criterion (D5, D6). AC-001/002/004's path-equality assertions are unimplementable as written on darwin (D3). |
| Traceability | 0.75 | 0.75 | Every REQ has ≥1 AC and the §D.1 map is explicit. Two mis-mappings: AC-003's body is REQ-6's wording but is mapped to REQ-2; AC-008 is mapped to REQ-3 but its subject is the §E fail-open constraint, not REQ-3 (D7). |

Harmonic mean = 4 / (1/0.75 + 1/0.50 + 1/0.50 + 1/0.75) = **0.60**.

---

## Defects Found

**D1. REQ-5/AC-005 rest on a factually false premise — `spec.md:111-115`, `plan.md` §B.1, `acceptance.md` AC-005 — Severity: critical — Class: blocking**

`spec.md:113` and `plan.md` §B.1 both assert that `TestFindStateDirFromWalksUp` pins "정확히 반대 계약" and that it will fail when REQ-2/REQ-3 land. I read the test. It does not.

Subtest 1 (`"an ancestor state dir wins over the starting directory"`, `tokens_state_dir_test.go:32-47`) builds `root := t.TempDir()`, puts `.moai/state` at `root`, and starts at `root/AppData/Local/Temp/TestSomething/001`. Under the recommended change, `FindProjectRoot`-style resolution walks up from the start, finds `.moai` at `root` — **still the same ancestor** — and `root/.moai/state` exists, so it returns exactly what the subtest asserts. No home boundary sits between the start and `root` on any platform (on Windows `$HOME` is *above* `root`, so the guard fires later, not earlier). **The subtest passes unchanged.**

What the change actually alters is only resolution that would *cross the home boundary* — which is subtest 3's territory, not subtest 1's. The recommended change **preserves** ancestor-wins inside a project tree; `FindProjectRoot` is itself an upward walk.

Consequence: `plan.md` §F M2 instructs the implementer to **invert** subtest 1 ("조상 미끼를 주장하지 않음을 단언"), and AC-005 makes that inversion a MUST-PASS gate. Following that instruction pins a contract the implementation does not and should not produce — the run-phase would either write a test that fails against a correct implementation, or bend the implementation to satisfy a wrong test.

**Required fix:** re-derive REQ-5 from the actual test. State that subtest 1 (ancestor-wins within a project tree) is **preserved**, and that the contract change binds only subtest 3 (which becomes a deterministic error assertion) plus any new home-boundary case. Rewrite AC-005 accordingly and remove the "반전" instruction from `plan.md` §F M2 and AP-1.

---

**D2. The `CLAUDE_PROJECT_DIR` hoist silently unifies two paths that differ today — `internal/cli/chain.go:29,64-72`; `spec.md:91-93` (REQ-1), `acceptance.md` AC-007 — Severity: critical — Class: blocking**

The lead asked whether the two branches produce the same path today. **They do not.**

```go
const ChainStateDir = ".moai/state/chain"          // chain.go:29

if projDir := os.Getenv(config.EnvClaudeProjectDir); projDir != "" {
    chainDir = filepath.Join(projDir, ChainStateDir)          // <proj>/.moai/state/chain
} else {
    stateDir, err := findStateDir()
    chainDir = filepath.Join(filepath.Dir(stateDir), "chain") // <root>/.moai/chain
}
```

`filepath.Dir("<root>/.moai/state")` = `<root>/.moai`, so the walk branch yields `<root>/.moai/**chain**` while the env branch yields `<root>/.moai/**state/chain**`. Even with a perfectly-correct root, the same project writes its chain events to two different locations depending on whether the env var happens to be set. Then `os.MkdirAll` creates whichever one it picked.

REQ-1 mandates hoisting the env check into the shared helper. That hoist **forces a choice** between these two paths, and either choice relocates existing `chain-events` data for one class of users — an undeclared, data-affecting behavior change. `plan.md` §E D2 notices the arithmetic (`<found>/../chain`) and even observes that `chain.go` "실제로는 프로젝트 루트를 원한다", but stops at "구현자 재량" without ever stating that the two branches disagree, without naming the canonical target, and without a migration note. AC-007 only checks that the inline branch disappeared — it would pass under either choice, including the one that abandons existing data.

Neither directory currently exists in this checkout (`ls -d .moai/state/chain .moai/chain` → both `No such file or directory`), so this repo has no data at risk — but the SPEC ships to every user, and `moai chain` has been writing to one of these for its whole life.

**Required fix:** add an explicit requirement naming the canonical chain directory, declare it as a behavior change in `spec.md` §E alongside REQ-3, and add an AC that asserts the resolved chain path for BOTH the env-set and env-unset cases (they must now agree). State whether existing files at the abandoned path are migrated, read as a fallback, or knowingly orphaned.

---

**D3. `FindProjectRoot`'s `EvalSymlinks` changes returned path strings on darwin — `internal/core/project/root.go:27-30`; unmentioned anywhere in the SPEC — Severity: critical — Class: blocking**

`FindProjectRoot` normalizes with `filepath.EvalSymlinks`. `findStateDirFrom` does not. On macOS `/var` is a symlink (`ls -ld /var` → `lrwxr-xr-x /var -> private/var`), and `t.TempDir()` returns the **unresolved** form while `getcwd(3)` returns the **resolved** form. Measured in this session:

```
$ python3 -c "import os,tempfile;d=tempfile.mkdtemp();os.chdir(d);print(d);print(os.getcwd())"
/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/tmp8b0_7lt7
/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/tmp8b0_7lt7
equal = False
```

Today `findStateDirFrom(start)` returns a path with **whatever prefix `start` carried**, so a test comparing against a `t.TempDir()`-derived expectation matches. After delegation through an `EvalSymlinks`-normalizing resolver, the returned path carries the `/private/var` prefix while the test's expected value carries `/var` — string comparison fails on the author's own platform.

This lands directly on: existing subtest 2 (`tokens_state_dir_test.go:49-62`, compares `got != own`), and the three new ACs whose Then-clause is a path equality — AC-001 ("결과는 `root/explicit/.moai/state`"), AC-002, AC-004. `plan.md` §E D3 discusses the home-boundary injection seam at length and never mentions symlink normalization; `spec.md` §E residual risks R1-R3 do not include it.

It is not a blocker for the design — it is a blocker for the ACs as written, and it is the single most likely thing to burn M2/M3.

**Required fix:** declare the normalization consequence in `spec.md` §E, and amend AC-001/AC-002/AC-004 so their assertions compare `EvalSymlinks`-normalized paths (or assert on the resolved form explicitly). State whether `findStateDirFrom`'s returned path is contractually normalized.

---

**D4. AC-008 is vacuous — it passes with zero verification — `acceptance.md` AC-008 — Severity: major — Class: blocking**

Judging command: `go test ./internal/cli/ -run 'Overlay|RegistryForOverlay' -v`. No such test exists — `grep -rn "Overlay" internal/cli/chain_test.go` returns nothing, and no `_test.go` in `internal/cli` matches either name. The command therefore prints `ok ... [no tests to run]` and exits 0. A reader marks AC-008 PASS having verified nothing about the fail-open invariant it exists to protect.

This matters more than usual here, because AC-004 deliberately *increases* how often resolution fails, which is exactly what drives traffic through `loadRegistryForOverlay`'s fail-open path (`chain.go:352-361`).

**Required fix:** name the test that must be **written** (as AC-002 and AC-004 correctly do), not a `-run` pattern against tests that do not exist. State the expected observation as a non-empty `--- PASS: <name>` line.

---

**D5. AC-003's first judging command has no binary criterion — `acceptance.md` AC-003 — Severity: major — Class: blocking**

`grep -n "findStateDir\|resolveTokensStateDir\|resolveChainStore" internal/cli/*.go | grep -v _test.go` with expected observation "첫 명령이 6개 호출 지점을 모두 공유 헬퍼 호출로 보여준다". No expected count, no expected line set; whether the output "shows the six going through the shared helper" is a human judgment. Today that grep returns 8 lines (6 call sites + 2 comment/definition lines at `tokens.go:373` and `chain.go:61`), so even the count is not self-evident.

The second command (`grep -c "for {" internal/cli/state.go`, expect 0) **is** properly binary — current value is 1, and that one loop is the walk, so 0 is the correct post-condition. Keep it.

**Required fix:** state the exact expected line count and the six expected `file:line` tokens, or replace with a mechanical assertion.

---

**D6. AC-007's expected observation is an "either/or" judgment — `acceptance.md` AC-007 — Severity: minor — Class: blocking**

"인라인 분기가 사라졌거나 헬퍼 위임으로 바뀌었다" — two acceptable outcomes, neither with a count. `grep -n "EnvClaudeProjectDir" internal/cli/chain.go` currently returns **two** hits: line 64 (`resolveChainStore`, in scope) and line 83 (`resolveCWD`, out of scope). The AC correctly warns not to confuse them but never says the expected post-state is exactly one remaining hit at line ~83. As written a reader can mark PASS on either output.

**Required fix:** expected observation = exactly 1 match, in `resolveCWD`, and 0 matches inside `resolveChainStore`.

---

**D7. Two AC→REQ mis-mappings — `acceptance.md` §D matrix + §D.1 — Severity: minor — Class: blocking**

AC-003's body is verbatim REQ-6's subject ("소비자 6곳 전부가 동일한 공유 헬퍼를 경유", `spec.md:117-119`) but both the matrix and §D.1 map it to REQ-2. AC-008's subject is the §E fail-open constraint (`spec.md:135`), not REQ-3's stop-and-fail behavior, yet it is mapped to REQ-3. The net effect: REQ-6 appears to carry two ACs when it carries one, and REQ-3's second AC does not test REQ-3.

**Required fix:** map AC-003 → REQ-6; either promote the fail-open constraint to a requirement and map AC-008 to it, or mark AC-008 cross-cutting like AC-009/AC-010.

---

**D8. `internal/cli/state_m2_test.go:37` `m2SetupState` is an unnamed dependent — Severity: minor — Class: blocking**

The helper chdirs a test into a temp tree "so findStateDir resolves it" and is used by the M2 golden-master state tests. It is not in `spec.md` §A's consumer table nor in `plan.md` §F M4's file list. It is unlikely to break (it works through the directory, not a path comparison, and the symlinked and resolved paths are the same physical directory), but the SPEC's blast-radius claim is "measured 6 consumers" and that measurement missed a test-side dependent on the same resolution.

**Required fix:** name it in `spec.md` §A or `plan.md` §F M4 with a one-line statement of why it is unaffected.

---

**D9. `module: cli` is not path-like — `spec.md:11` — Severity: minor — Class: optional**

Schema requires a path-like value; siblings use `internal/example`. Lint only checks non-empty, so this passes mechanically. `internal/cli` would be correct.

**D10. `version: 0.1.0` unquoted and `priority: high` lowercase — `spec.md:4,9` — Severity: minor — Class: optional**

Schema documents `version` as a quoted semver string and the priority enum as `P0..P3` / `High|Medium|Low|Critical` (capitalized). `internal/spec/lint.go:756` only tests non-empty, so neither is rejected. Cosmetic drift from the SSOT.

**D11. §D.2's Given-When-Then blocks name no line-level anchor for the "미끼" home boundary — Severity: minor — Class: optional**

AC-002 requires a directory that "plays the role of `$HOME`" but the resolution reads `paths.Home()` internally with no injection point — `plan.md` §E D3 correctly identifies this as an open M1 decision. Leaving the decision open is defensible; what is not stated is that AC-002 is **unrunnable** until that decision is made, so it cannot be scheduled before M1's D3 choice.

---

## Design probes the lead asked for — answers

**Does REQ-3 break an existing caller or test?** No existing test asserts ancestor-wins across a home boundary, and `m2SetupState` (D8) is unaffected. But REQ-3 introduces a regression class the SPEC does not analyze: `FindProjectRoot`'s marker is `.moai` while `findStateDir`'s is `.moai/state`. Today a walk from inside a subdirectory that owns a bare `.moai` (no `state`) **skips it** and keeps climbing to the real project root — success. After delegation, resolution **stops** at that subdirectory and fails, even though a valid project-root state dir sits above. This repo has a documented history of exactly that shape (the `internal/hook/.moai/` anomaly, `manager-develop-prompt-template.md` §B7). AC-004 does pin these semantics deliberately, so it is a declared consequence rather than a hidden one — but the SPEC nowhere enumerates the trees this newly breaks, and "옳은 자리에서 실패" is the justification offered for a case where the *right* answer was reachable and is now refused. Folded into D2/completeness rather than raised as a separate must-fix.

**Is the marker mismatch fully handled?** No — see the paragraph above. There is no AC for "nested bare `.moai` below a valid project root".

**Is the `CLAUDE_PROJECT_DIR` hoist safe for `chain.go:67`?** No. See **D2** — the two branches produce different paths today and the hoist forces an undeclared choice between them.

**Are the ACs falsifiable?** Seven of ten are. AC-008 is vacuous (D4); AC-003's first command and AC-007 require judgment (D5, D6). AC-001/002/004 are falsifiable in shape but unimplementable as written on darwin (D3).

---

## §4 Gaps (what this audit did NOT verify)

- I did not run `go test ./internal/cli/...` — no baseline was established, so I make no claim about the current suite's colour.
- I did not verify Windows behavior; the `%USERPROFILE%`-shaped reasoning in D1 is derived from the test file's own comment plus the code, not measured.
- I did not check whether any downstream user has data in `.moai/chain` or `.moai/state/chain` beyond this checkout (both absent here).
- I did not audit `internal/config/token_budget_guard.go:51` — the SPEC scopes it out and I accept that scoping.
- I did not evaluate whether Tier M is the correct tier; the SPEC's file/LOC estimate was not independently measured.

## §5 Residual risk

- D3's severity depends on the implementer's D3-seam choice: routing through `os.Getwd()` inside `FindProjectRoot` (rather than a start-injecting variant) may hide the divergence from production while still breaking the tests, which is the worse failure — a green production path with red tests invites the wrong fix.
- D1's correction may reduce REQ-5 to near-nothing. If subtest 1 is preserved and subtest 3 is merely strengthened, REQ-5 no longer carries a distinct requirement and may fold into REQ-2's AC set. That is a scope reduction, not a defect.
- The chain-path decision (D2) may turn out to need its own SPEC if migration of existing event files is wanted.

---

## Recommendation

FAIL. Route the eight blocking defects (D1-D8) back to `manager-spec`. In priority order:

1. **D1** — re-derive REQ-5/AC-005 from the actual test; remove the "invert subtest 1" instruction from `plan.md` §F M2 and AP-1. This one, unfixed, actively misdirects the run-phase.
2. **D2** — name the canonical chain directory, declare it as a behavior change in §E, add a both-branches-agree AC, state the migration disposition.
3. **D3** — declare `EvalSymlinks` normalization in §E; amend AC-001/002/004 to compare normalized paths.
4. **D4/D5/D6** — make AC-008, AC-003 (first command), and AC-007 binary.
5. **D7** — fix the two AC→REQ mappings.
6. **D8** — name `state_m2_test.go`'s `m2SetupState` and state why it is unaffected.
7. Also correct `spec.md:134`'s exhaustive "동작 변경 1건" claim once D2 and D3 are declared.

D9-D11 are optional; surface them to the orchestrator but do not gate on them.

The core design judgment — converge on `internal/core/project`'s protected convention rather than adding a depth cap or a fourth flag — is sound and well-argued, and claims 1, 2, 3, 4 and 6 all held under verification. This FAIL is about three undeclared consequences and a false premise, not about the direction.

**VERDICT: FAIL 0.60**

### MUST-FIX
- D1 — REQ-5/AC-005 false premise (`TestFindStateDirFromWalksUp` subtest 1 does NOT invert)
- D2 — undeclared `chain.go` path unification (`.moai/state/chain` vs `.moai/chain`)
- D3 — undeclared `EvalSymlinks` normalization; AC-001/002/004 unimplementable as written on darwin
- D4 — AC-008 vacuous (no such test exists; passes with zero verification)
- D5 — AC-003 first command non-binary
- D6 — AC-007 expectation is an either/or judgment
- D7 — AC-003→REQ-2 and AC-008→REQ-3 mis-mappings
- D8 — `state_m2_test.go:37` unnamed dependent
