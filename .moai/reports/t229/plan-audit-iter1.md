# SPEC Review Report: SPEC-CODEX-VERDICT-SYNTH-001

Iteration: 1/1 (Tier S ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.7625** (Tier S threshold 0.75)

Reasoning context ignored per M1 Context Isolation. Judgment rests on the four SPEC
artifacts, the evidence base named in the dispatch, and my own measurements against
this worktree.

| Item | Value |
|---|---|
| Measurement tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t229` @ `910f3ffed` (base `294b4b6ab` = `origin/main`) |
| Every `file:line` below | re-verified in THIS tree, not the primary checkout |
| Live codex probe | NOT run (dispatch constraint) |
| Package tests | NOT run — no finding below depends on a test run |

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CVS-001` … `REQ-CVS-004` at `spec.md:104/108/112/116`. Sequential, no gap, no duplicate, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-CVS-*` in `spec.md`), not the ACs. REQ-001 state-driven (`While … the system shall …`), REQ-002 ubiquitous, REQ-003 `Where …` , REQ-004 ubiquitous + `shall not` (GEARS canonical negative). No informal language, no Given/When/Then presented as a REQ. See D5/D6 for two pattern-precision findings that are **not** MP-2 failures.
- **[PASS] MP-3 YAML frontmatter validity** — `spec.md:1-15` carries all 12 canonical fields (`id`,`title`,`version` quoted,`status`,`created`,`updated`,`author`,`priority`,`phase`,`module`,`lifecycle`,`tags`) plus `tier: S`. No rejected snake_case alias (`created_at`/`updated_at`/`labels`/`spec_id`) present. `plan.md` and `acceptance.md` mirror the same 13 keys.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/cli` of this tool itself). Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the only external reference is `SPEC-AUDIT-MULTI-MODEL-001`; measured `status: completed`, which is not in {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` returns 0 across `spec.md`, `plan.md`, `acceptance.md`. Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-VERDICT-SYNTH-001/` exits 1 (no match). `research.md` absent, as expected at Tier S.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Strong: `spec.md:27-33` governing principle, `spec.md:122-128` AC-shape constraint, `plan.md:61` records a rejected alternative. Deducted for D2 (a [HARD] premise at `spec.md:94` that measures false) and D6/D7 ambiguities. |
| Completeness | 0.75 | 0.75 | All sections present incl. 5 `### Out of Scope — <topic>` H3s with bullets (`spec.md:144-164`). Deducted for D1's uncovered case, D4's unresolved open item, and no landing step in `plan.md`. |
| Testability | 0.80 | 0.75-1.0 | All 5 ACs are binary-testable Go assertions; each names the mutant it kills. AC-CVS-001 verified genuinely property-shaped. Deducted for D3 (understated RED set). |
| Traceability | 0.75 | 0.75 | No orphan AC, no uncovered REQ. Exactly one indirect/partial mapping: REQ-CVS-003 is worded over *any* two diverging signals, AC-CVS-004 witnesses only the stated×bullet pair (D1). |

Aggregate = mean(0.75, 0.75, 0.80, 0.75) = **0.7625** ≥ 0.75.

---

## Adversarial hunt — the six items the dispatch named

### 1. Is AC-CVS-001 actually property-shaped? — **YES, verified**

The assertion is `Verdict != "pass"` applied by iterating the corpus (`acceptance.md:59-63`).
Adding a member of the unrecognized class requires no assertion edit, which is the SPEC's own
discriminator (`acceptance.md:34`). It is deliberately a **negative** property, not `== inconclusive`
— correct, since a positive equality would break on a member that legitimately synthesizes `fail`.

**Independence from AC-CVS-002 — verified, they kill different mutants.** Against the named
mutant (a) (one score regex added, `verdict := "pass"` fall-through kept):

| Corpus member | Mutant (a) result | AC-CVS-001 | AC-CVS-002 |
|---|---|---|---|
| C1 blocking-table | `pass` (no signal matches) | **FAILS** | not exercised |
| C8 scored `FAIL 0.75` | `fail` (score regex reads it) | passes | **passes** |

So AC-CVS-002 alone lets mutant (a) through and AC-CVS-001 alone kills it. The coverage claim
at `acceptance.md:65` is true. The reciprocal constraint also holds: a "return `inconclusive` for
everything" mutant satisfies AC-CVS-001 and dies on AC-CVS-002's `fail` assertion; a "return `fail`
for everything in adversarial" mutant satisfies both and dies on AC-CVS-005's live-probe clause
(`inconclusive` required). The AC set is mutually constraining as claimed.

### 2. Does any AC observe nothing? — measured, see the RED/pin split below

I replicated `codexFindingBullet` (`internal/cli/mcp_codex.go:1115`) and `codexStatedVerdict`
(`:1130`) verbatim plus the synthesis body (`:1144-1151`) in a throwaway program and measured
every corpus member and AC input. Result:

```
C1 blocking-table              -> pass
C2 numbered-list               -> pass
C3 json-blob                   -> pass
C4 title-only                  -> pass
C5 prose-1line                 -> pass
C6 korean                      -> pass
C7 empty                       -> pass
C8 scored-FAIL                 -> pass
AC002 prose-falsepos           -> pass
AC003 native-clean             -> pass
AC004 diverge                  -> fail
AC005 probe-head               -> inconclusive
3sig label-fail+score-PASS     -> fail
```

No AC observes nothing. AC-CVS-003 and AC-CVS-005 are pure pins **and say so**; that is honest,
not a defect. One precision error falls out of this table — see D3.

### 3. Is the native/adversarial seam reachable? — **YES, the claim is exact**

Traced in this tree:

| Entry point | Method passed | Mode |
|---|---|---|
| `HandleCodexReviewGate` (`internal/cli/codex_review_gate.go:66`) → `runCodexReviewRPC` (`:89`) | `codexMethodReviewStart` | native |
| `handleCodexAudit` (`mcp_codex.go:1186`) | `codexMethodReviewStart` (`:1212`) or `codexMethodTurnStart` (`:1219`) | both |
| `performCodexAudit` (`mcp_convergence.go:409`) → `codexReviewRPC` (`:425`) | `codexMethodTurnStart` | adversarial |
| `runCodexTaskTurn` (`codex_task.go:87`) → `session.runTurn` (`:97`) | `codexMethodTurnStart` | adversarial |

All four converge on `runTurn` (`mcp_codex.go:680`), whose signature is
`(ctx context.Context, method string, params map[string]any)` — it **does** already carry `method`,
and `synthesizeReviewOutput` is called from exactly one production site inside it (`:705`).
`spec.md:96` and `plan.md:49` are correct. **Not a blocking defect.**

Consequence the SPEC handles correctly: `codex_task` also rides the adversarial branch, so REQ-CVS-001
changes its `Verdict`. I verified `plan.md:120`'s mitigation premise — `codex_task` consumes only
`Summary` (`codex_task.go:251` and `:318` both assign `result.Output = out.Summary`); the sole
`Verdict` reference in that file is the timeout construction at `:107`. The premise is **true**.

### 4. Does the regression AC protect the gate? — the AC is right, its stated reason is **wrong**

AC-CVS-003 does pin the native bullet-less clean pass, and the empty-string half is forced by the
existing test's `"": "pass"` case (`codex_review_rpc_test.go:120`), so the pin is coherent.

But the *rationale* is false. `isBlockVerdict` (`codex_review_gate.go:117-119`) blocks only on a
`fail` prefix, and the gate's terminal line reads `return allow, nil // pass / inconclusive ⇒ ALLOW`
(`:108`). Flipping native to `inconclusive` therefore **cannot** block a normal change through this
gate. See D2 — the requirement survives on other grounds, the premise does not.

Mode coherence of the empty-string case: internally consistent (adversarial → not-pass by
AC-CVS-001, native → pass by AC-CVS-003), but the input is unreachable in production — `runTurn`
short-circuits empty review text into `inconclusive` at `mcp_codex.go:702-703` before the
synthesizer is ever called. See D7.

### 5. Scope honesty — the demotion of conservative ordering is **wrong as applied**. This is D1.

`spec.md:90` demotes ordering on the ground that "실측표 7행 어디에도 현재 조합이 잘못된 값을 낸
반례가 없다". That is true — **of a two-signal implementation**. This SPEC's own M2 adds a third
signal, and `plan.md:43` prescribes the list order `codexStatedVerdict` → `codexScoredVerdict` →
`codexFindingBullet`. Under that order, with the existing assign-and-overwrite body:

```
body:  "Verdict: fail\n\nPASS 0.95 / 1.00"
       stated → fail  →  scored overwrites → pass  →  no bullet  →  final: pass
```

A stated merge-blocking verdict laundered into `pass` — the exact `spec.md:29` violation this
SPEC exists to prevent. Measured today the same body yields `fail` (table above, last row),
because the scored signal does not yet exist; the counterexample is created **by this SPEC**, and
it is not in the 7-row table. The demotion's premise does not transfer. Finding D1.

### 6. Out-of-scope integrity — **clean, nothing load-bearing**

| Declared out of scope | Load-bearing for any in-scope AC? |
|---|---|
| t234 / #1632 `Findings: []Finding{}` | No. AC-CVS-004's record assertion targets the new `SynthesisNote` (`plan.md:73`), never `Findings`. |
| t246 worktree tree misread | No AC references tree selection. |
| t248 binary-commit stamping | No. `spec.md:174` bans live-probe evidence outright, which removes the dependency rather than deferring it. |
| session_id / convergence persistence | No. AC-CVS-004's second half asserts on the value returned by `converge` (`mcp_convergence.go:135`), a pure function over `[]PerBackendVerdict`. Persistence is untouched. |
| GLM backend | No AC touches `performGLMAudit`. |

### 7. Repo policy conflict — **NEGATIVE, no conflict found**

`.claude/rules/local/repo-local-pr-policy.md` disables Route A for all tiers. `plan.md` prescribes
**no** main-direct push — it prescribes no landing route at all (§D ends at M4, §E is self-verification).
Silence is not a violation. Recorded as D8 (minor) because the run-phase orchestrator must consult
the policy independently.

---

## Defects Found

**D1.** `spec.md:88-90` (§A.5) + `plan.md:79` (§C.5) + `acceptance.md:93-106` (AC-CVS-004) — The
conservative-ordering demotion cites a measured table taken against a **two-signal** implementation,
but M2 introduces a third signal and `plan.md:43` fixes its position between stated and bullet. Under
that order a body carrying both `Verdict: fail` and `PASS 0.95 / 1.00` synthesizes `pass`, violating
`spec.md:29`. REQ-CVS-003 is worded generically ("two verdict signals diverge") but AC-CVS-004
witnesses only the stated×bullet pair, so an implementation that passes all five ACs can ship this
laundering. — Severity: **critical** — Class: **blocking** — Required fix: add one AC (or a second
Given/When/Then block on AC-CVS-004) pinning stated×scored precedence, e.g. `Verdict: fail` +
`PASS 0.95 / 1.00` ⇒ `fail`, and `PASS 0.95 / 1.00` + `- [P1] …` ⇒ `fail`. Gate M2: do not start it
until that AC exists.

**D2.** `spec.md:94` (§A.6, [HARD]) + `acceptance.md:90` (mutant (b)) — Both assert that flipping the
default to `inconclusive` in both modes breaks `HandleCodexReviewGate`'s clean-pass and thereby
blocks a normal change. Measured false: `isBlockVerdict` (`codex_review_gate.go:117-119`) matches a
`fail` prefix only, and `codex_review_gate.go:108` returns ALLOW for `pass / inconclusive` alike. The
real cost of that mutant is loss of signal, not a false block — `converge` distinguishes the two
(all-required-`inconclusive` falls back to the claude verdict, `mcp_convergence.go:125-128`), and
the existing test pins `pass` at `codex_review_rpc_test.go:120`. REQ-CVS-004 stands; its stated
reason does not. — Severity: **major** — Class: **blocking** — Required fix: rewrite `spec.md:94`
and the mutant (b) note to name the actual harm (a native `inconclusive` is indistinguishable from
fail-open at the convergence layer and discards the pinned `pass` contract), and drop the
"정상 변경을 차단" claim.

**D3.** `acceptance.md:66` — "사전구현 트리에서 이 AC 는 **C1·C5·C7 에서 실패해야 한다**" understates
the RED set. Measured: **all eight** members C1-C8 synthesize `pass` today. An implementer taking the
named three as the complete expected-RED list, per the DoD RED-evidence item at `acceptance.md:137`,
records a partial baseline. — Severity: minor — Class: blocking — Required fix: state C1-C8 (all
eight), or say "at minimum C1·C5·C7" and require the run-phase to record the actual RED set.

**D4.** `plan.md:75` — "미확인 사항: `review-output.schema.json` 이 외부 계약으로 고정돼 있다면…"
left open into the audit. Measured: **no such file exists** in the repository —
`grep -rl 'review-output.schema.json' . --exclude-dir=.git` returns only prose references (SPEC docs,
`internal/cli/*.go` comments, two SKILL.md copies). Adding `SynthesisNote` breaks no file-backed
schema. — Severity: minor — Class: optional — Required fix: close the item in-place with that
measurement rather than deferring it to kickoff.

**D5.** `spec.md:112` (REQ-CVS-003) — uses the `Where` pattern for a runtime data condition ("two
verdict signals diverge within one backend's review body"). GEARS reframes `Where` as a capability
gate / feature flag / static config; a per-body runtime condition is `When`. Not scored as an MP-2
failure — the sentence is one of the five patterns and is well-formed — but the pattern selection is
imprecise. — Severity: minor — Class: optional — Required fix: `When two verdict signals diverge …`.

**D6.** `spec.md:112` and `spec.md:116` — both REQs are compound. REQ-CVS-003 binds two subjects
(the system, the convergence engine) across four behaviors; REQ-CVS-004 binds two unrelated
preservation duties (native `pass`, `codex_task` output text). Compound requirements weaken the
one-REQ-one-outcome mapping that Traceability relies on. — Severity: minor — Class: optional —
Required fix: split REQ-CVS-003 into synthesis-side and convergence-side requirements; consider
splitting REQ-CVS-004.

**D7.** `acceptance.md:88` and `:91` — the empty string is used as the discriminating witness that
mode wiring exists ("같은 입력이 모드에 따라 갈린다"). Production never delivers it: `runTurn` returns
`inconclusiveReview("codex review produced no verdict text")` at `mcp_codex.go:702-703` before the
synthesizer. The pin itself is correct (it preserves `codex_review_rpc_test.go:120`), but the
mode-split demonstration rests on an unreachable input. — Severity: minor — Class: optional —
Required fix: add a reachable discriminating witness — the same prose body (e.g. C5) asserted
non-`pass` in adversarial and `pass` in native.

**D8.** `plan.md:81-104` (§D) — no landing/integration milestone. This repo forces Route B for all
tiers (`.claude/rules/local/repo-local-pr-policy.md`, `enforce_admins: true`). No conflict exists —
the plan prescribes no main-direct push — but the route is unstated. — Severity: minor — Class:
optional — Required fix: name the PR route in §D or §E.

**D9.** Stale line citations. This card was derailed twice by stale-tree citations, so precision here
is load-bearing: (a) `spec.md:134` cites the `Findings: []Finding{}` hardcode at `mcp_codex.go:1152`;
measured **1155** (1152 is `return ReviewOutput{`). (b) `spec.md:173` cites `backendCallFn` at
`mcp_convergence.go:368`; measured **367**. Verified-correct citations, for contrast:
`mcp_codex.go:1145`, `:1130`, `:680`, `mcp_convergence.go:135`, `plan.md:73`'s `:134`,
`codex_review_gate.go:66`, `codex_review_rpc_test.go:114`, `mcp_codex.go:1144-1156`. — Severity:
minor — Class: blocking — Required fix: correct the two line numbers against `910f3ffed`.

**Not defects — recorded so a later reader does not re-litigate them:** the `runTurn` `method` seam
(claim exact); `codex_task` consuming only `Summary` (premise true); AC-CVS-001's property shape
(genuine); AC-001/AC-002 mutant independence (verified distinct); out-of-scope integrity (clean);
repo PR policy (no conflict); MP-2 on REQ-CVS-003's `Where` (pattern-precision only, see D5).

---

## RED vs regression pin — expected behavior on the current tree

Every row measured against `910f3ffed`. All five ACs additionally fail to compile until
`synthesizeReviewOutput` takes `method`; the semantics below are what matters after that.

| AC | Current-tree expectation | Class |
|---|---|---|
| AC-CVS-001 | **FAILS** on **all eight** C1-C8 (every one synthesizes `pass`) — not the three named at `acceptance.md:66` | **RED** |
| AC-CVS-002 | **FAILS** — C8 `FAIL 0.75 / 1.00` → `pass`, wants `fail`. The prose false-positive clause also **FAILS** today (`the suite reported PASS 12 times…` → `pass`, wants non-`pass`) | **RED** (both clauses) |
| AC-CVS-003 | **PASSES** today — native clean prose → `pass`, empty → `pass` | **pure pin** (gate protection; self-labelled) |
| AC-CVS-004 | **split**: the value assertion (`fail`) **PASSES** today; the `SynthesisNote` record assertion and the `converge` `DisagreementFlag` assertion **FAIL** (field does not exist) | **mixed** — record half RED, value half pin |
| AC-CVS-005 | **PASSES** today on all three clauses — existing 4-input test green; live-probe body → `inconclusive`; `codex_task` output unchanged | **pure pin** (self-labelled at `acceptance.md:116`) |

Net: two ACs carry genuine RED detection, one is mixed, two are pure regression pins. That is a
healthy split for a Tier S card — the pins are declared as pins rather than dressed up as new checks.
The one precision error is D3.

---

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria clear, and the aggregate 0.7625 clears the Tier S
threshold of 0.75. The SPEC's central claim survives adversarial testing: AC-CVS-001 is genuinely
property-shaped, it kills mutant (a) independently of AC-CVS-002, the mode seam exists exactly as
described, and the out-of-scope boundary is clean.

Two blocking items are owed before the milestones they touch:

1. **D1 gates M2.** Do not start the scored-verdict recognizer until an AC pins stated×scored
   precedence. Without it, M2 can reintroduce the SPEC's own target defect and every AC stays green.
2. **D2 gates M1's documentation.** Correct the false gate premise at `spec.md:94` and mutant (b)
   before it is used to justify anything further; the requirement it supports is sound, the reason is not.

D3 and D9 are cheap corrections that should ride the same edit (D9 especially — this card has a
two-round history of stale citations). D4-D8 are optional and left to the orchestrator's discretion;
routing all of them into a revision would cost more than it buys.
