# SPEC Review Report: SPEC-CLI-STATE-DIR-BOUND-001 — Iteration 3

Iteration: 3 (operator-authorized beyond the Tier M ceiling of 2; the ceiling was reached at iter2 and escalated rather than silently looped)
Verdict: **PASS**
Overall Score: **0.80** (harmonic mean; arithmetic 0.8125)
Grading against: **Tier M PASS threshold = 0.80** (`spec-workflow.md` § SPEC Complexity Tier)
Score trajectory: iter1 0.60 → iter2 0.667 → iter3 **0.80** — monotonic improvement, no STOP signal

**The score sits exactly on the threshold, not above it.** I did not round up: harmonic 0.80 meets `≥ 0.80` and nothing more. The residual debt in § Residual debt the run phase inherits is what separates this from a comfortable pass, and the orchestrator should carry it into the run-phase delegation prompt rather than treat PASS as "nothing left to do".

Reasoning context ignored per M1 Context Isolation. The author's closure summary was read only to locate claims; every closure was re-derived against the artifacts and the tree.

---

## §1 Claim — closure verification

| iter2 MUST-FIX | Author's claim | My verdict |
|---|---|---|
| E6 — 4th undeclared consequence (env priority + `clean` delete radius) | REQ-1 split on read/destructive axis; REQ-9 excludes `clean`; REQ-6 premise rewritten; §E → 4 rows; AC-015 | **CLOSED** |
| E4 — REQ-1 vs REQ-8 contradiction | "그대로" narrowed to *no walk*; env value explicitly `EvalSymlinks`-normalized; AC-001/AC-010 rewritten to exercise the env branch | **CLOSED — and better than the fix I asked for** |
| E7 — REQ-7 head vs failure clause | head gains explicit case-4 carve-out; 4-case table; AC-009 case 4; AC-008 negative scoped to the success path | **CLOSED** |
| E5 — hook writer unnamed | `chain_event.go:67` in §A, §F, REQ-7 rationale #1; R4 rewritten as "case 2 is the expected state" | **CLOSED** |
| E1 — AC-003 wrong count | call-only grep pattern; per-file table re-measured; `chain.go:61` added to AC-014; AP-17 codifies the lesson | **CLOSED — numbers verified correct** |

All five closed. Two of them (E4, E1) were closed with a generalization rather than a point fix, which is the stronger form.

---

## §2 Evidence — verbatim

### E1 closure — AC-003's new command, run verbatim

```
$ grep -c "err := findStateDir()" internal/cli/clean.go internal/cli/tokens.go internal/cli/chain.go internal/cli/state.go
internal/cli/state.go:2
internal/cli/clean.go:1
internal/cli/chain.go:2
internal/cli/tokens.go:1
```

Total 6, matching the AC's "변경 전 (현재 실측)" column cell-for-cell (`clean 1 / tokens 1 / chain 2 / state 2 / 합 6`). The call-only pattern correctly excludes the definition at `state.go:212` and the comment at `chain.go:61` that produced iter2's 8. `plan.md` §G AP-17 records the failure mode by name so it is not re-derived.

### Every remaining AC judging command, run verbatim

```
$ grep -c "falls back to findStateDir() directory walk" internal/cli/chain.go   → 1   (AC-014 cmd3 expects 0 post-change) ✓ stated
$ grep -c "EnvClaudeProjectDir\|CLAUDE_PROJECT_DIR" internal/cli/clean.go        → 0   (AC-015 cmd2 expects 0, preservation)  ✓ stated
$ grep -c "err := findStateDir()" internal/cli/clean.go                          → 1   (AC-015 cmd3 expects 0 post-change)     ✓ stated
$ grep -c "EvalSymlinks" internal/cli/state.go                                   → 0   (AC-010 cmd2 expects ≥1 post-change)    ✓ stated
```

Every command executes cleanly and returns exactly the current value the AC states. No wrong-expected-number survives.

### E4 closure — the fix generalizes past what iter2 asked for

I asked for a carve-out sentence plus one AC. The SPEC delivered that (`spec.md:116`, REQ-1's "그대로의 의미" paragraph), and `acceptance.md` §D.2 additionally names *why* iter2's AC concealed the contradiction:

> [HARD] **양쪽을 정규화하는 비교는 REQ-8 을 검증하지 못한다.** `normPath(got) == normPath(want)` 는 구현이 정규화를 하든 안 하든 통과한다 — iter2 의 AC-001 이 그 형태였고, 그래서 REQ-1 과 REQ-8 의 모순을 가렸다.

and pins the corrected shape (`got != normPath(t, want)` — raw actual, normalized expected), then codifies it as `plan.md` AP-10. AC-010 now carries two explicit sub-branches (walk / env), closing the "the scenario never said whether the env was set" hole I flagged.

### `-run` pattern non-vacuity (re-checked)

All 11 patterns resolve: `TestFindStateDirFromWalksUp` exists (`tokens_state_dir_test.go:32`); `-run 'M2'` matches 6 existing `TestStateM2_*`; the remaining nine are explicitly marked **작성 대상: 신규** with `--- PASS: <name>` plus "`--- SKIP` / `no tests to run` 없음" as the expected observation. AP-16 records the iter1 vacuous-AC failure mode by name.

### Re-probe — is the read/destructive split coherent, or does it move the hazard?

**REQ-9 (`clean` anchored on cwd) is safe in a worktree — measured.**

```
$ pwd                      → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t164
$ ls -d .moai .moai/state  → drwxr-xr-x  .moai
                             drwxr-xr-x  .moai/state
```

This worktree carries its own `.moai` **and** `.moai/state`, so `FindProjectRoot` from inside stops at the worktree root and `clean` targets the worktree — correct. A worktree carrying `.moai` but no `.moai/state` stops at the worktree root and **fails** per REQ-3 class A — also safe. The escape to the primary checkout requires a worktree with **no `.moai` at all**, which does not arise on any branch carrying tracked `.moai/` content (`git ls-files .moai` is non-empty: `config`, `docs`, `specs`, …). See F5 for the unstated precondition.

Notably it is REQ-3's marker change — the thing REQ-3 itself calls an intentional regression (class B) — that supplies REQ-9's protection here. The SPEC never connects the two.

**`tokens.go` append under the env: defensible, not a hazard.** `internal/session/anchor.go:26-33` documents, as a measured fact (2026-08-17), that the `moai cc -w` launcher *pins the session's project to the checkout it launched from*. Under that model, a worktree session writing its token ledger into the launching checkout's `.moai/state` is consistent with the launcher's own project attribution rather than a mis-attribution. The SPEC's "worst case is read/append in the named project" holds for tokens.

**`chain.go:67` MkdirAll: no mode change.** It is the one consumer that already reads the env (`chain.go:64`), so REQ-1 moves nothing there; REQ-7 changes only *which* directory it creates, and that is declared as §E row 3.

### Re-probe — §E table completeness

I looked for a fifth **behavior change** and did not find one. The four rows (marker / normalization / chain path / env priority) cover every observable change this SPEC causes, and the 6-way recount is arithmetically sound — I verified the "1 existing" claim independently:

```
$ grep -rn "EnvClaudeProjectDir" internal/ | grep -v _test.go   → chain.go:64 is the only reader among the six consumers
```

4 new + 1 existing + 1 excluded = 6. Correct.

What I did find is **under-analysis of row 4** (F2, F3) and two factual errors outside the table (F1) — recorded as debt below, not as a missing row.

### MP re-checks

```
$ grep -c "^### REQ-" spec.md          → 9    (Tier M ceiling 16 — within budget)
   AC count 15                           (Tier M ceiling 16 — within budget)
$ grep -c "^### Out of Scope" spec.md  → 6    (each with `-` bullets)
$ grep -c syscall spec.md              → 0    (D8 auto-PASS)
$ grep -c 'NEEDS CLARIFICATION' plan.md → 0
$ grep -n "^status:" .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001/spec.md → status: completed
```

---

## §3 Baseline-attribution

All evidence gathered in this run, against this tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t164`, branch `main`. No file in the tree was modified by this audit. Artifacts read at `spec.md` 294 lines / `plan.md` 254 / `acceptance.md` 427. The worktree `.moai` probe and every AC grep were executed in this session's shell.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-1…REQ-9 present, no gaps, no duplicates, uniformly unpadded. Count 9 ≤ Tier M ceiling 16. (REQ-9 is authored out of numeric order in the body — placed beside REQ-1 for the read/destructive contrast — which is presentation, not a numbering defect.)
- **[PASS] MP-2 GEARS compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`); Given-When-Then in `acceptance.md` is the correct verification-layer format, graded under Group 4. New this iteration: REQ-9 (`… state 디렉터리를 해석할 때, `CLAUDE_PROJECT_DIR` 을 참조해서는 안 되며(SHALL NOT) … 기준으로 삼아야 한다(SHALL)`) is event-driven + unwanted-behavior; conformant.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields, canonical names, no rejected alias; `version: "0.3.0"` quoted, `priority: High`, `module: internal/cli`.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/cli`).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one referenced SPEC, exists, `status: completed`. No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` count 0. Auto-PASS.
- **[PASS] MP-7 clarification gate** — no markers; `research.md` absent (Tier M).

No must-pass failure.

---

## Category Scores

| Dimension | Score | Band | Δ vs iter2 | Evidence |
|---|---|---|---|---|
| Clarity | 0.75 | 0.75 | = (composition improved) | Both iter2 contradictions are genuinely resolved — REQ-1/REQ-8 via the "그대로 = 걷기 없음" determination (`spec.md:116`) and REQ-7 head/failure via the explicit case-4 exception (`spec.md:171,199`). One scope defect remains: REQ-6's "걷기 구현은 **정확히 하나**" is unscoped and literally false on landing (F1). A reasonable engineer resolves it as `internal/cli`-scoped because §B says so, so it resolves consistently — 0.75, not lower. |
| Completeness | 0.75 | 0.75 | ↑ from 0.50 | The §E 4-row behavior-change table is **complete** — I probed for a fifth change and found none, and the 6-way recount verifies. Structure complete; 6 × `### Out of Scope` H3 with bullets; `clean.go:136`, `m2SetupState`, and `chain_event.go` all now named because the author went looking. Deductions: R2 still counts "규약이 셋" when there are at least four (F1), and row 4's blanket safety claim is under-verified for the registry consumer (F2). |
| Testability | 0.75 | 0.75 | = | Every judging command verified verbatim against the tree, every stated current value correct, every `-run` pattern resolving. The §D.2 [HARD] anti-concealment rule and AC-009's anti-skip instruction are both sharper than the fixes iter2 requested. Softest spot: AC-009 cases 2 and 4 assert "관측 가능한 경고 … (stderr 또는 printer)" without pinning a stream or a string, so a tester exercises judgment on what counts — and both of REQ-7's riskiest branches hinge on exactly that assertion. |
| Traceability | 1.00 | 1.0 | ↑ from 0.75 | All 9 REQs carry ≥1 AC; every AC maps to a valid REQ or is explicitly marked 횡단 (AC-012/013/014); no orphans; iter1's two mis-mappings corrected by number at `acceptance.md:44`; the §D.2 entry-point naming table makes AC-003/007/015 binary and carries a rename-correspondence obligation into `progress.md`. Meets the 1.0 anchor as written. |

Harmonic mean = 4 / (1/0.75 + 1/0.75 + 1/0.75 + 1/1.00) = 4 / 5.00 = **0.80**. Arithmetic = 0.8125.

---

## Residual debt the run phase inherits

None of the following blocks run-phase entry. Each is verified, and each should be carried into the run-phase delegation prompt (Section A: Context) so the implementer sees it.

**F1 — REQ-6's "exactly one walk" is unscoped and false on landing; R2 miscounts the fragmentation — Severity: major**

`spec.md:165` states, unscoped: *"state 해석의 상향 걷기 구현은 **정확히 하나**여야 하며(SHALL)"*. A second unbounded upward walk over the identical `.moai/state` marker survives untouched in another package:

```go
// internal/hook/cwd_changed_relocate.go:78
func findRegistryUpward(dir string) (string, bool) {
	for {
		candidate := filepath.Join(dir, session.DefaultRegistryPath)   // ".moai/state/active-sessions.json"
		if _, err := os.Stat(candidate); err == nil { return candidate, true }
		parent := filepath.Dir(dir)
		if parent == dir { return "", false }
		dir = parent
	}
}
```

Same marker, same shape, **no home guard** — the exact defect class this SPEC exists to remove. It is named nowhere: not in §B, not in any Out-of-Scope entry, not in R2. And `spec.md:256` R2 asserts *"규약이 셋(`findStateDir`, `FindProjectRoot`, `findRepoRoot`)"* — the count is wrong; there are at least four, and the fourth is the one that walks the very marker REQ-6 claims to unify.

The intent is unambiguous (§B scopes the whole SPEC to `internal/cli`), so this is a wording and accounting defect, not a design defect. Fix: scope REQ-6 to `internal/cli`, correct R2's count, and add `findRegistryUpward` as an Out-of-Scope entry or a follow-up card. Note also that no AC could falsify REQ-6's "exactly one" clause — AC-003 cmd3 (`grep -c "for {" internal/cli/state.go`) sees only `state.go`.

**F2 — the session registry has two locations by design, and row 4's safety claim does not account for it — Severity: minor-major**

`internal/session/anchor.go:26-33` records a measured fact:

> TWO registries are consulted, because the two lane launch shapes register in different places (measured 2026-08-17): 1. the tree-LOCAL registry … 2. the CALLER's project registry (`CLAUDE_PROJECT_DIR`, else CWD) — **where `moai cc -w` lanes actually register** … while the worktree itself carries no local registry at all.

`LiveAnchoredSessions` reads **both** precisely because neither alone is complete. `loadRegistryForOverlay` (`chain.go:353`) reads exactly one, and REQ-1 changes *which* one. This does not make REQ-1 wrong — for the `cc -w` shape, env-first lands on the registry where those lanes actually register, which is an improvement. But the SPEC's blanket justification ("최악의 결과가 지목된 프로젝트에서 읽거나 추가하는 것") was asserted without noticing there are two registries, so for this one consumer the claim is unverified rather than established. One sentence in row 4 citing `anchor.go` would close it, and it strengthens the SPEC rather than weakening it.

**F3 — `clean` and the other five can resolve to different projects in the same session, with no signal — Severity: major**

REQ-9 is the right call and the probe confirms it prevents the delete escape. Its coherence cost is undeclared: with `CLAUDE_PROJECT_DIR` set and diverging from the cwd root, `moai state dump` reads project B while `moai clean` deletes in project A, in the same session, silently. An operator who inspects and then cleans is looking at one project and acting on another.

`spec.md` R8 acknowledges the *classification* can age; it does not acknowledge the *simultaneous divergence*. The cheap mitigation is one line of output from `clean` naming the resolved root (which also makes AC-015's post-merge dry-run check in `acceptance.md` §D.6 self-evidencing). No AC and no §E note currently requires it.

**F4 — REQ-1 vs REQ-4: the env-set-but-rootless case is unspecified — Severity: minor**

REQ-1: with the env set, use it as the project root and do not walk. REQ-4's `<cwd>/.moai/state` fallback is conditioned on *"프로젝트 루트를 찾지 못할 때"*. With the env set the root is found by fiat, so when `<env>` contains no `.moai` at all, `resolveTokensStateDir` writes to `<env>/.moai/state` (created lazily) and REQ-4's fallback never fires. AC-006's Given explicitly excludes the env case, so nothing pins the behavior. Probably intended; state it.

**F5 — REQ-9's safety rests on an unstated precondition — Severity: minor**

REQ-9 is safe in a worktree *because the worktree carries its own `.moai` marker* — measured above. Were a worktree to carry no `.moai`, the cwd-anchored walk would climb `.claude/worktrees/` → `.claude/` → the primary checkout and delete there, reproducing the exact hazard REQ-9 closes. The shape does not arise on branches with tracked `.moai/` content, so this is a documentation gap, not a live hole. Worth one sentence in REQ-9, along with the observation that REQ-3's marker change — the change REQ-3 calls an intentional regression — is what supplies the protection.

---

## Regression Check (iter1 D1-D11, iter2 E1-E7)

| Prior finding | Status |
|---|---|
| iter1 D1 — REQ-5 false premise | **RESOLVED, held across two revisions** — subtest 1 preserved in `spec.md:155,159`, `plan.md` §B / M4 / AP-1, AC-005; no invert instruction survives |
| iter1 D2/D3 — chain path, EvalSymlinks | **RESOLVED** (REQ-7, REQ-8) |
| iter1 D4 — vacuous AC-008 | **RESOLVED** (AC-012 + AP-16) |
| iter1 D5 → iter2 E1 — AC-003 count | **RESOLVED** — call-only pattern, numbers verified by execution, AP-17 codifies |
| iter1 D6/D7/D8/D9/D10/D11 | **RESOLVED** (verified in iter2, unchanged) |
| iter2 E4 — REQ-1 vs REQ-8 | **RESOLVED** (REQ-1 determination + §D.2 [HARD] + AP-10 + AC-010 branch B) |
| iter2 E5 — hook writer | **RESOLVED** (§A, §F, REQ-7 rationale 1, R4 rewrite) |
| iter2 E6 — 4th consequence | **RESOLVED** (REQ-1 split + REQ-9 + REQ-6 rewrite + §E row 4 + AC-015 + plan D6/M3/AP-4/AP-5) |
| iter2 E7 — REQ-7 self-contradiction | **RESOLVED** (case-4 carve-out + 4-case table + AC-009 case 4 + AC-008 scoping) |

No defect appears unchanged across iterations. No stagnation.

---

## §4 Gaps (what this audit did NOT verify)

- I did not run `go test ./internal/cli/...` — no baseline established; I make no claim about the suite's current colour.
- I did not verify Windows behavior. AC-009 case 4's permission-based failure reproduction is flagged by the SPEC itself as possibly non-reproducible there; I did not test whether the proposed injectable-seam alternative is achievable in the current `chain.go` shape.
- I did not enumerate every upward walk in the repo — F1 names one I found while probing the registry; there may be others. R2's count should be re-derived by the author rather than taken from my "at least four".
- `CLAUDE_PROJECT_DIR` was unset in this session (`printenv` → exit 1), so the *frequency* of the row-4 mode change remains unmeasured across all three iterations. Its treatment rests on the delete-radius argument, which REQ-9 now removes.
- I did not verify that `findStateDirNoEnv` (the §D.2 naming convention) is free of collisions with existing identifiers.

## §5 Residual risk

- F3 is the item most likely to surface as a user-visible surprise after landing, and it is the one with no AC. If the run phase implements REQ-9 without an output line naming the resolved root, the divergence becomes discoverable only by an operator noticing that nothing was deleted where they expected.
- `plan.md` §E D3 (home-boundary seam) is still undecided by design, and its R6 hazard — production green, tests red — remains live until M1. The decision criterion ("does production take the same code path as the test?") is now explicit, which is the right mitigation, but the risk is deferred rather than closed.
- REQ-7's migration is still the only irreversible operation in scope, and R4 now correctly states that case 2 is the expected state for users with data. The "warning text accurately describes the situation" mitigation is only as good as the warning string, which no AC pins (see the Testability note).

---

## Recommendation

**PASS at 0.80**, exactly meeting the Tier M threshold. Proceed to Implementation Kickoff Approval — which remains mandatory and score-independent; this PASS governs only Phase 1 verdict re-execution and never auto-bypasses the plan→run human gate.

The three-iteration arc is genuine: every MUST-FIX I raised was closed, two of them by generalizing the fix into a named anti-pattern rather than patching the instance, and the one design question that mattered (E6's read/destructive axis) was answered with a structural split that survives adversarial probing. `clean` no longer follows the environment; the split is coherent rather than hazard-shifting, and I verified the worktree case that would have falsified it.

Carry F1-F5 into the run-phase delegation prompt as declared debt. Suggested handling:

1. **F1** — fix in the SPEC before M1 (scope REQ-6 to `internal/cli`; correct R2's count; add `findRegistryUpward` as Out-of-Scope or a follow-up card). It is a five-word scope qualifier plus two sentences, and leaving a literally-false requirement in an approved SPEC invites a future reader to act on it.
2. **F3** — decide during M3 whether `clean` names its resolved root in output. If yes, add the assertion to AC-015; if no, record the decision in `progress.md`.
3. **F2, F4, F5** — one sentence each, folded into the M7 documentation milestone.

Follow-up cards worth opening now, since three audits have each surfaced one: the `findRegistryUpward` home-guard gap (F1), the hook writer's hardcoded path literal (already noted in §B as a follow-up candidate), and the Windows CI ancestor-`.moai/state` diagnosis (R1).

**VERDICT: PASS 0.80**
