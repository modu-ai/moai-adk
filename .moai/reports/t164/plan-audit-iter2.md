# SPEC Review Report: SPEC-CLI-STATE-DIR-BOUND-001 — Iteration 2

Iteration: 2/2 (Tier M ceiling — `harness.plan_audit_tier_ceilings`: S=1, M=2, L=3)
Verdict: **FAIL**
Overall Score: **0.667** (harmonic mean; arithmetic 0.688)
Grading against: **Tier M PASS threshold = 0.80** (`spec-workflow.md` § SPEC Complexity Tier)
Score trajectory: iter1 0.60 → iter2 0.667 — **improving, no STOP signal** (no score regression, so the LEAN scope-reduction escalation does not fire on that trigger)

Reasoning context ignored per M1 Context Isolation. Author's closure summary was read only to locate the claims; every closure below was re-verified against the artifacts and the tree independently.

Iteration 2 is scoped to the iter1 defect delta plus a regression check over D1-D11, per the Retry Loop Contract, and extended with the three re-probes the delegation named.

---

## §1 Claim — closure verification (author's 5 claimed closures + iter1's 8 MUST-FIX)

| iter1 MUST-FIX | Author's claim | My verdict |
|---|---|---|
| D1 — REQ-5 false premise | re-derived; subtest 1 PRESERVED; "invert" declared an error | **CLOSED** |
| D2 — chain path undeclared | REQ-7 fixes canonical + migration | **CLOSED with gaps** (E5, E7) |
| D3 — EvalSymlinks undeclared | REQ-8 declares it; ACs normalize | **CLOSED, but introduces E4** |
| D4 — AC-008 vacuous | renumbered AC-012, real named test | **CLOSED** |
| D5 — AC-003 non-binary | made binary with an exact count | **NOT CLOSED — the count is wrong (E1)** |
| D6 — AC-007 either/or | three exact-count commands | **CLOSED** (verified verbatim) |
| D7 — AC→REQ mis-mappings | AC-003→REQ-6, AC-008→AC-012 cross-cutting | **CLOSED** |
| D8 — `m2SetupState` unnamed | named in spec §A, plan §F M4, §H | **CLOSED** |
| §E "one behavior change" | replaced by a 3-row table | **INCOMPLETE — a fourth exists (E6)** |
| D9/D10 frontmatter | version quoted, priority `High`, module `internal/cli` | **CLOSED** |

Six of eight iter1 MUST-FIX are fully closed. Two are not (D5), and the D2/D3 closures each carry a residual defect. The §E exhaustiveness claim fails a second time.

---

## §2 Evidence — verbatim

### D1 closure — verified across four surfaces

`spec.md:141` (REQ-5): *"**서브테스트 1 은 무변경으로 통과한다.** … 서브테스트 1 을 반전시키라는 iter1 의 지시는 올바른 구현이 만들지 않는 계약을 고정하게 만드는 오도였다."*
`spec.md:137` table row: subtest 1 → `**보존.** 단언은 그대로 참이다`.
`plan.md:27` § "iter1 의 거짓 전제 (제거됨)" — same derivation, independently stated.
`plan.md:133` (§F M3): `서브테스트 1 **보존**(단언 의미 무변경)`.
`plan.md:176` (§G AP-1): inverting subtest 1 is now the **anti-pattern**.
`acceptance.md:159` (AC-005): `서브테스트 1 을 **반전시키지 않는다**`.

Grep for any surviving invert instruction: none. The reasoning reproduced in all four places matches my iter1 derivation exactly (no home boundary between the start and `root`; `FindProjectRoot` is itself an upward walk). **D1 is genuinely closed, not merely re-worded.**

### D6 closure — AC-007's three commands, run verbatim

```
$ grep -c "EnvClaudeProjectDir" internal/cli/chain.go                                  → 2   (AC expects 1 post-change)
$ awk '/^func resolveChainStore/,/^}/' internal/cli/chain.go | grep -c "Env…"          → 1   (AC expects 0 post-change)
$ awk '/^func resolveCWD/,/^}/' internal/cli/chain.go | grep -c "Env…"                 → 1   (AC expects 1, unchanged)
```
All three execute cleanly and report exactly the current values the AC states. Genuinely binary.

### D5 NOT closed — AC-003's command, run verbatim

```
$ grep -n "findStateDir()" internal/cli/clean.go internal/cli/tokens.go internal/cli/chain.go internal/cli/state.go
internal/cli/state.go:78:	stateDir, err := findStateDir()
internal/cli/state.go:154:	stateDir, err := findStateDir()
internal/cli/state.go:212:func findStateDir() (string, error) {          ← the DEFINITION
internal/cli/clean.go:65:	stateDir, err := findStateDir()
internal/cli/chain.go:61:// Prefers CLAUDE_PROJECT_DIR; falls back to findStateDir() directory walk.   ← a COMMENT
internal/cli/chain.go:67:		stateDir, err := findStateDir()
internal/cli/chain.go:353:	stateDir, err := findStateDir()
internal/cli/tokens.go:377:	if dir, err := findStateDir(); err == nil {
$ … | wc -l                                                                             → 8
```
See E1.

### Re-probe 1 — is REQ-7's canonical path consistent with every consumer?

```
$ grep -rn "filepath.Dir(stateDir)" internal/cli/
internal/cli/clean.go:136:	moaiDir := filepath.Dir(stateDir) // .moai/
internal/cli/chain.go:71:		chainDir = filepath.Join(filepath.Dir(stateDir), "chain")
```
Two sites, not one. `clean.go:136` uses the identical arithmetic to reach `<root>/.moai/config/sections/state.yaml` (`loadRetentionDays`). Under REQ-2 the state dir is `<root>/.moai/state`, so `Dir()` still yields `<root>/.moai` and the config path stays correct — **clean.go is not broken by REQ-7**. But it is a second, unnamed instance of the arithmetic REQ-7 calls "산술의 부산물", and the SPEC nowhere states it was examined. Minor.

The consequential finding is outside `internal/cli` entirely — see E5.

### Re-probe 2 — does REQ-8 introduce a new inconsistency?

```
$ grep -rn "EnvClaudeProjectDir" internal/ | grep -v _test.go
internal/config/envkeys.go:290:	EnvClaudeProjectDir = "CLAUDE_PROJECT_DIR"
internal/cli/hook.go:651  internal/cli/chain.go:64  internal/cli/chain.go:83
internal/cli/hook_pre_push.go:134  internal/hook/chain_event.go:57  internal/hook/chain_event.go:80
internal/hook/path_resolve.go:79   internal/hook/post_tool_metrics.go:99
internal/hook/protocol.go:116,128  internal/session/anchor.go:95
```
Twelve production readers, **every one a raw `os.Getenv` with no normalization anywhere**. See E4.

### Re-probe 3 — is every `-run` pattern non-vacuous?

| AC | pattern | status |
|---|---|---|
| AC-001 | `TestStateDirHonoursProjectDirEnv` | 신규, declared |
| AC-002 | `TestStateDirDoesNotCrossHomeBoundary` | 신규, declared |
| AC-004 | `TestStateDirStopsAtProjectRoot` | 신규, declared |
| AC-005 | `TestFindStateDirFromWalksUp` | **exists** (`tokens_state_dir_test.go:32`) |
| AC-006 | `TestResolveTokensStateDirFallsBackToCwd` | 신규, declared |
| AC-008 | `TestChainDirIsCanonicalUnderBothBranches` | 신규, declared |
| AC-009 | `TestChainLegacyEventsRelocation` | 신규, declared |
| AC-010 | `TestStateDirReturnsNormalizedPath` | 신규, declared |
| AC-011 | `TestStateDirStopsAtNestedBareMoai` | 신규, declared |
| AC-012 | `TestLoadRegistryForOverlayFailsOpen` | 신규, declared **with the iter1 failure mode named** |
| §D.3 | `-run 'M2'` | matches 6 existing `TestStateM2_*` tests — verified |

Every pattern resolves to an existing test or an explicitly-declared new one, and each new AC carries an expected observation of the form `--- PASS: <name>` with `--- SKIP` / `no tests to run` excluded. `plan.md:187` AP-12 codifies the iter1 lesson as a named anti-pattern. **The vacuous-AC class is closed.** Two contract gaps remain uncovered by any AC (E4, E7) — a different failure mode (missing AC, not vacuous AC).

### MP re-checks

```
$ grep -c "^### REQ-" spec.md            → 8    (Tier M ceiling 16 — within budget)
   AC count 14                             (Tier M ceiling 16 — within budget)
$ grep -c "^### Out of Scope" spec.md    → 5    (each with `-` bullets)
$ grep -c syscall spec.md                → 0    (D8 auto-PASS)
$ grep -rn 'NEEDS CLARIFICATION' <dir>   → exit 1, no matches
$ grep -n "^status:" .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001/spec.md → status: completed
```

---

## §3 Baseline-attribution

All evidence gathered in this run, against this tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t164`, branch `main`. No file in the tree was modified by this audit. Artifacts read at `spec.md` 243 lines / `plan.md` 203 / `acceptance.md` 341 / `progress.md` 51. The `CLAUDE_PROJECT_DIR` probe (`printenv CLAUDE_PROJECT_DIR` → exit 1, unset) was run in this session's shell.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-1…REQ-8 at `spec.md:104,110,116,127,131,147,151,175`. Sequential, no gaps, no duplicates, uniformly unpadded. Count 8 ≤ Tier M ceiling 16.
- **[PASS] MP-2 GEARS compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`); the Given-When-Then bodies in `acceptance.md` are the correct verification-layer format and are graded under Group 4. All eight REQs carry an explicit trigger plus `SHALL` / `SHALL NOT`. New this iteration: REQ-7 (`… 해석할 때, 결과는 … 이어야 한다(SHALL)`) is event-driven; REQ-8 (`… 반환하는 경로는 … 여야 한다(SHALL)`) is ubiquitous. Both conform.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present, canonical names, no rejected alias. iter1's D9/D10 cosmetics repaired: `version: "0.2.0"` now quoted, `priority: High` now matches the documented enum casing, `module: internal/cli` now path-like.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/cli`). Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one referenced SPEC, exists, `status: completed`. No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears 0 times. Auto-PASS.
- **[PASS] MP-7 clarification gate** — no `[NEEDS CLARIFICATION]` markers; `research.md` absent (Tier M).

**No must-pass failure.** The FAIL is driven by the rubric score (0.667 < 0.80) and by the five blocking defects below.

---

## Category Scores

| Dimension | Score | Band | Δ vs iter1 | Evidence |
|---|---|---|---|---|
| Clarity | 0.75 | 0.75 | ↑ 0.75 → 0.75 (composition changed) | The chain under-determination that drove iter1's finding is decisively resolved (REQ-7 + plan §E D2 explicitly replace "구현자 재량"). Two **new** edge-of-contract contradictions replace it: REQ-1 vs REQ-8 on the env branch (E4) and REQ-7's main SHALL vs its failure-path SHALL (E7). The mainline resolution contract is materially clearer than iter1; the ambiguity is now confined to two branches. |
| Completeness | 0.50 | 0.50 | undeclared consequences 3 → 1 | `spec.md:203` claims exhaustiveness a second time — "선언된 동작 변경 — 전 3건 (iter1 은 1건이라 적었고, 그것은 틀렸다)". A fourth exists and carries a delete radius (E6). A second chain-path writer outside the package is unnamed (E5). Structure itself is complete: HISTORY, WHY, WHAT, REQUIREMENTS, 5 × `### Out of Scope` H3 with bullets, frontmatter complete. |
| Testability | 0.75 | 0.75 | ↑ from 0.50 | Vacuous-AC class closed and codified (AP-12). Every `-run` pattern resolves. AC-007 verified binary and correct verbatim. AC-008 now fails under the wrong canonical choice — the exact iter1 hole it was written to close. Three reproduction-first obligations with progress.md citation. Remaining: AC-003's expected count is wrong (E1); REQ-7's failure path and the env-branch normalization have no AC (E7, E4). |
| Traceability | 0.75 | 0.75 | = | `§D.1` maps all 8 REQs; AC-012/013/014 explicitly marked cross-cutting; `acceptance.md:42` names and corrects both iter1 mis-mappings by number. Gap: REQ-7's `SHALL` failure disposition maps to no AC. |

Harmonic mean = 4 / (1/0.75 + 1/0.50 + 1/0.75 + 1/0.75) = **0.667**. Arithmetic = 0.688. Both below the Tier M threshold of 0.80.

---

## Defects Found

**E1. AC-003's expected count is wrong — the AC fails against a correct implementation — `acceptance.md:107-118`, `plan.md:35,41` — Severity: major — Class: blocking**

iter1's D5 said AC-003's first command had no binary criterion. iter2 gave it one — and the number is wrong. Run verbatim today it returns **8**, not the 6 the AC expects, because `grep "findStateDir()"` also matches:

- `internal/cli/state.go:212` — `func findStateDir() (string, error) {`, the **definition**;
- `internal/cli/chain.go:61` — a **comment**: `// Prefers CLAUDE_PROJECT_DIR; falls back to findStateDir() directory walk.`

The per-file table (`clean 1 / tokens 1 / chain 2 / state 2`) is wrong in the same two cells: actual is `chain 3 / state 3`.

After the change, the definition **survives** — it *is* the shared helper REQ-6 requires — so state.go cannot reach 2 without renaming or deleting the helper. An implementer chasing "exactly 6" is being pushed toward contorting the code to satisfy a mis-specified grep. A non-binary criterion has been replaced with a binary and false one, which is the worse of the two.

`plan.md:35` carries the same error (`grep -n "findStateDir()" internal/cli/*.go  # 호출 6곳 확인`) and `plan.md:41` states the same expectation, so the Pre-Flight step would report a mismatch before M1 even starts.

Separately, `chain.go:61`'s comment ("falls back to findStateDir() directory walk") becomes stale under REQ-1/REQ-2 but is covered by no AC — AC-014 scopes only `state.go` and `tokens.go`.

**Required fix:** count call sites with a pattern that excludes the definition and comments (e.g. `grep -n "err := findStateDir()\|, err := findStateDir()"`, or `grep -c "findStateDir()" | minus the known definition line`), and restate the per-file table against the measured post-change values. Add `chain.go:61` to AC-014's comment sweep.

---

**E2 (numbering reserved — folded into E1).**

---

**E4. REQ-1 and REQ-8 contradict on the `CLAUDE_PROJECT_DIR` branch, and no AC pins it — `spec.md:106` (REQ-1) vs `spec.md:177` (REQ-8) — Severity: critical — Class: blocking**

This is the new inconsistency the delegation asked me to probe. It is real.

- REQ-1: *"`CLAUDE_PROJECT_DIR` 이 설정되어 있을 때, 공유 state 해석 헬퍼는 그 값을 **그대로** 사용하고 …"* — use the env value **verbatim**.
- REQ-8: *"수렴 후 state 해석이 반환하는 경로는 `filepath.EvalSymlinks` 로 정규화된 형태여야 한다(SHALL)."* — **unconditional**, no branch carve-out.

Nothing normalizes `CLAUDE_PROJECT_DIR`. All twelve production readers are raw `os.Getenv` (evidence §2, re-probe 2). So when the env value is an unnormalized path — on darwin, any `/var/…` path, which is precisely what `t.TempDir()` produces — the two SHALLs cannot both hold: normalizing violates REQ-1's "그대로", returning raw violates REQ-8.

The AC set does not resolve it, and one AC actively conceals it:

- **AC-001** compares `normPath(결과) == normPath(root/explicit/.moai/state)` — normalizing **both sides** makes the assertion pass under either implementation. It cannot distinguish them.
- **AC-010** asserts `반환값 == filepath.EvalSymlinks(반환값)` — this **fails** on the env branch when the env value is unnormalized. But AC-010's scenario (`GIVEN darwin 에서 /var 가 …`) does not say whether the env var is set, so an implementer will naturally exercise the walk branch and the AC passes while the contradiction stays untested.

A reasonable engineer resolves this either way. That is the definition of the Clarity 0.50 anchor, confined here to one branch.

**Required fix:** state the resolution explicitly in REQ-1 or REQ-8 — either "the env value is normalized before use (REQ-1's 그대로 means *no walk*, not *no normalization*)" or "REQ-8 binds the walk branch only". Then add an AC that sets `CLAUDE_PROJECT_DIR` to a deliberately unnormalized path and asserts the chosen contract.

---

**E5. A second chain-directory writer outside `internal/cli` is unnamed, and it makes REQ-7's "both exist" case the LIKELY case rather than an edge case — `internal/hook/chain_event.go:67` — Severity: critical — Class: blocking**

The delegation asked whether REQ-7's canonical path is consistent with every consumer. Inside `internal/cli` it is. Outside it, there is a third constructor of the chain path that the SPEC never mentions:

```go
// internal/hook/chain_event.go:55-68
projectDir := payload.ProjectDir
if projectDir == "" { projectDir = os.Getenv(config.EnvClaudeProjectDir) }
if projectDir == "" { projectDir = payload.CWD }
...
chainDir := filepath.Join(projectDir, ".moai", "state", "chain")
storePath := filepath.Join(chainDir, "events.jsonl")
```

Two consequences, pulling in opposite directions:

1. **It independently corroborates REQ-7's choice.** The hook has always written `.moai/state/chain`, hardcoded — not via `ChainStateDir`, not via `resolveChainStore`. REQ-7's justification currently rests on the constant plus a semantic argument; this is stronger evidence and should be cited.

2. **It falsifies REQ-7's implicit framing of case 2 as an edge case.** On any machine where the chain hook has run *and* `moai chain` was invoked without `CLAUDE_PROJECT_DIR`, the hook wrote `.moai/state/chain/events.jsonl` while the CLI wrote `.moai/chain/events.jsonl`. **Both files already exist** — that is not a corner, it is the expected state of a machine that used both surfaces. REQ-7 case 2's disposition is "canonical wins, legacy untouched, warn, no merge" (`spec.md:171`), so those users permanently keep a second history no reader will ever open. `spec.md:222` R4 concedes "이력이 두 갈래로 남는다" but presents it as a residual risk of a rare path; the hook makes it the default outcome for the population most likely to have data.

The hook also re-implements the canonical path literal rather than sharing it, which is the same duplication REQ-6 forbids inside `internal/cli` — the SPEC neither scopes it out nor names it.

**Required fix:** name `internal/hook/chain_event.go:67` in `spec.md` §A and §F; cite it as corroborating evidence in REQ-7's rationale; restate R4 to say case 2 is the *expected* state on any machine that ran both surfaces, not an edge case; and either scope the hook's duplicate literal out explicitly (with a reason) or bring it under the canonical constant.

---

**E6. A FOURTH undeclared behavior change — the REQ-1 hoist gives five consumers `CLAUDE_PROJECT_DIR` priority they have never had, and one of them deletes — `spec.md:203-209` (§E table), `spec.md:106` (REQ-1) — Severity: critical — Class: blocking**

The delegation pre-declared that a fourth undeclared consequence is a MUST-FIX. One exists.

Today exactly one consumer reads `CLAUDE_PROJECT_DIR`: `resolveChainStore` (`chain.go:64`). The other five — `clean.go:65`, `tokens.go:377`, `chain.go:353`, `state.go:78`, `state.go:154` — call `findStateDir()`, which never reads it and always walks from cwd. REQ-1 hoists the env check into the shared helper, so **all six** begin honoring it.

For those five that is not a neutral gain; it is a change of resolution *mode*. Whenever `CLAUDE_PROJECT_DIR` is set and does not name the root the walk would have found, those five now resolve somewhere else than they did yesterday. And this repo carries a documented precedent for exactly that divergence: `main-checkout-branch-guard.md` § Mechanical Enforcement records that `$CLAUDE_PROJECT_DIR` did **not** track a worktree-resident agent's actual directory — "querying the primary checkout about itself always answered 'primary', which misclassified a worktree-resident agent's branch-state command" — which is why the discriminant was moved to `input.CWD`.

The blast radius is the largest in the SPEC. `internal/cli/clean.go:116` is `os.RemoveAll(path)`. Under the hoist, `moai clean` run inside a worktree while `CLAUDE_PROJECT_DIR` names the primary checkout resolves to the **primary checkout's** state dir and deletes there. That is the SPEC's own founding hazard — "a state directory the caller never named" — reintroduced through a different door, in the one consumer that deletes.

The §E table has three rows (marker, chain path, normalization). This is not among them, and `spec.md:203` asserts the table is complete ("전 3건"). The same exhaustiveness claim that iter1's D2/D3 falsified is falsified again, by one row instead of three.

**Required fix:** add a fourth `§E` row — "five consumers begin honoring `CLAUDE_PROJECT_DIR`; where it diverges from the walk result, resolution moves, and `clean` deletes at the new location". Add an AC that fixes the behavior for at least `runClean` when the env var names a directory other than the walk's answer. State whether the hoist is intended to bind `clean`'s delete path at all, or whether `clean` should be carved out.

---

**E7. REQ-7's migration-failure disposition contradicts REQ-7's main SHALL and is covered by no AC — `spec.md:153` vs `spec.md:170`; `acceptance.md:215-242` — Severity: major — Class: blocking**

REQ-7's head clause: the result *"`CLAUDE_PROJECT_DIR` 설정 여부와 무관하게 **`<project-root>/.moai/state/chain`** 이어야 한다(SHALL)"*.
REQ-7's failure clause: *"이전이 실패하면 fail-open — 그 호출은 **레거시 경로를 계속 쓰고** 사용자에게 관측 가능한 경고를 남겨야 한다(SHALL)"*.

On the migration-failure path the second SHALL returns `.moai/chain`, which the first SHALL forbids. The carve-out is clearly intended but is not written as one, so the two read as a contradiction.

AC-009 enumerates exactly three cases (legacy-only / both / canonical-only). **The failure case is absent** — the one path where user data is at risk and where the two SHALLs disagree has no test. AC-008 makes it worse by asserting unconditionally that neither branch yields `root/.moai/chain` (`acceptance.md:207`), which the fail-open path would violate.

**Required fix:** rewrite REQ-7's head clause with an explicit exception ("except on the migration-failure path defined below"); add AC-009 case 4 — migration fails (e.g. destination unwritable) → resolution returns the legacy path, a warning is observable, no data is lost; and scope AC-008's "neither is `.moai/chain`" assertion to the success path.

---

**E8. `clean.go:136` is a second, unnamed site of the `filepath.Dir(stateDir)` arithmetic REQ-7 calls a by-product — Severity: minor — Class: optional**

```
internal/cli/clean.go:136:	moaiDir := filepath.Dir(stateDir) // .moai/
```
`loadRetentionDays` derives `<root>/.moai/config/sections/state.yaml` this way. Under REQ-2 the state dir stays `<root>/.moai/state`, so `Dir()` still yields the right directory — **not broken**. But REQ-7 builds an argument on this arithmetic being unintended, and the SPEC never records that the only other site was examined and found safe. One sentence would close it.

**E9. REQ-7 names a specific function in a requirement — `spec.md:153` — Severity: minor — Class: optional**

*"`resolveChainStore` 가 chain 디렉터리를 해석할 때 …"* puts an implementation identifier in the requirement layer (RQ-4: requirements state WHAT, not the function that does it). GEARS-conformant and unambiguous, so it does not affect MP-2; noted only for form. "chain 디렉터리 해석" would carry the same meaning without binding the function name.

**E10. `plan.md:35,41` Pre-Flight repeats E1's wrong count — Severity: minor — Class: optional (folded into E1's fix)**

---

## Regression Check (iter1 defects D1-D11)

| iter1 | Status | Evidence |
|---|---|---|
| D1 — REQ-5 false premise | **RESOLVED** | Four surfaces corrected + AP-1 inverted; no invert instruction survives (§2) |
| D2 — chain path undeclared | **RESOLVED (with E5/E7 residue)** | REQ-7 + plan §E D2 + AC-008 binary under either choice |
| D3 — EvalSymlinks undeclared | **RESOLVED (introduces E4)** | REQ-8 + `normPath` convention + AP-7/AP-8 |
| D4 — AC-008 vacuous | **RESOLVED** | AC-012 with a real named test + AP-12 codifying the failure mode |
| D5 — AC-003 non-binary | **UNRESOLVED** | Now binary but the expected count is wrong — 8, not 6 (E1) |
| D6 — AC-007 either/or | **RESOLVED** | Three commands verified verbatim: 2 / 1 / 1 |
| D7 — AC→REQ mis-mappings | **RESOLVED** | `acceptance.md:42` corrects both by number |
| D8 — `m2SetupState` unnamed | **RESOLVED** | `spec.md:54`, `plan.md:148`, `plan.md:202`; `-run 'M2'` verified to match 6 tests |
| D9 — `module` not path-like | **RESOLVED** | `module: internal/cli` |
| D10 — version unquoted / priority casing | **RESOLVED** | `version: "0.2.0"`, `priority: High` |
| D11 — AC-002 unrunnable before D3 | **RESOLVED** | `acceptance.md:95` now states the schedule dependency explicitly |

No stagnation: no defect appears unchanged across both iterations. D5 is the sole carry-over and it changed shape (non-binary → binary-but-wrong), so it is a new mistake in the same location rather than a missed fix.

---

## §4 Gaps (what this audit did NOT verify)

- I did not run `go test ./internal/cli/...` — no baseline established; I make no claim about the suite's current colour.
- I did not measure how often `CLAUDE_PROJECT_DIR` is set in practice. `printenv CLAUDE_PROJECT_DIR` in this session returned exit 1 (unset), so E6's frequency is **unmeasured** — its severity rests on the delete radius and the documented worktree divergence, not on an observed set-rate.
- I did not verify Windows behavior; the marker and temp-under-home reasoning is derived from code and the test file's own comment, not measured.
- I did not enumerate whether any distributed user actually holds `.moai/chain/events.jsonl`. Neither directory exists in this checkout (`ls -d` → both absent), consistent with the SPEC's own statement.
- I did not audit `internal/hook/chain_event.go` beyond its path construction and env-resolution order.

## §5 Residual risk

- E4's resolution direction matters more than it looks: if the author resolves it by normalizing `CLAUDE_PROJECT_DIR`, that silently changes what eleven other readers of the same variable would see if they ever adopt the helper. If resolved by exempting the env branch from REQ-8, then AC-010 must state which branch it exercises or it will drift back to testing nothing.
- E6's fix may reveal that `clean` should not honor `CLAUDE_PROJECT_DIR` at all, which would make REQ-1 per-consumer rather than global and reopen REQ-6's "all six share one resolution" premise. That is a design question, not a wording fix, and it is the most likely reason this SPEC needs a third pass.
- plan.md §E D3's R6 hazard (production green, tests red) is now explicitly carried and given a decision criterion — good — but it remains undecided until M1, so AC-002/AC-011 stay unschedulable until then.

---

## Recommendation

FAIL at 0.667 against the Tier M threshold of 0.80. **This is iteration 2 of 2 for Tier M**, so the retry loop's ceiling is reached: the orchestrator must escalate to the user rather than iterate to a third pass on its own authority. Verdict authority for the SPEC stays here; the choice among the three escalation options is the user's.

The five blocking defects are narrow and localized — none touches the direction, and the delta from iter1 is substantial and real (six of eight MUST-FIX fully closed, the vacuous-AC class eliminated and codified, chain path decided, normalization declared). My assessment for the escalation:

- **PASS-with-debt is defensible** if E1, E4, E6, E7 are folded in as a pre-run correction before `/moai run` — E1 is a one-line grep fix, E4 and E7 are one carve-out sentence plus one AC each, E5 is a naming plus an R4 restatement. Only E6 carries a genuine design question (should `clean` honor the env var at all).
- **Scope reduction is worth considering** on one axis only: REQ-7's legacy migration is the sole logic this SPEC *adds*, it is the only irreversible operation in scope (R4), and E5 shows its blast radius is wider than modeled. Splitting REQ-7 into a follow-up SPEC would leave a clean convergence-only change here and give the data migration its own审 pass.

Fix order if iteration continues:

1. **E6** — the fourth `§E` row plus the `clean` decision. This is the one with a delete radius and the one that may change REQ-1's shape.
2. **E4** — resolve REQ-1 vs REQ-8 on the env branch; add the AC that distinguishes them.
3. **E7** — write REQ-7's failure carve-out as an exception; add AC-009 case 4; scope AC-008's negative assertion to the success path.
4. **E5** — name `internal/hook/chain_event.go:67`; cite it in REQ-7's rationale; restate R4.
5. **E1** — fix the grep pattern and the per-file table in both `acceptance.md` and `plan.md`; add `chain.go:61` to AC-014.

E8/E9/E10 are optional; surface them, do not gate on them.

**VERDICT: FAIL 0.667**

### MUST-FIX
- E1 — AC-003's expected count is wrong (8, not 6); binary-but-false, and `plan.md:35,41` repeats it
- E4 — REQ-1 ("그대로") contradicts REQ-8 (unconditional normalization) on the `CLAUDE_PROJECT_DIR` branch; AC-001 conceals it, AC-010 does not pin it
- E5 — `internal/hook/chain_event.go:67` is an unnamed second writer of `.moai/state/chain`; it corroborates REQ-7's choice AND makes case 2 the expected state, not an edge case
- E6 — FOURTH undeclared behavior change: five consumers begin honoring `CLAUDE_PROJECT_DIR`; `clean.go:116` `os.RemoveAll` follows it
- E7 — REQ-7's failure-path SHALL contradicts its head SHALL, has no AC, and AC-008's unconditional negative assertion conflicts with it
