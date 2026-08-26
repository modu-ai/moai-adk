# SPEC Review Report: SPEC-CODEX-VERDICT-SYNTH-001 — iter2 (narrow re-audit)

Iteration: 2 (lead-directed delta re-audit)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.8025** — moved **monotonically up** from iter1's 0.7625 (+0.0400). No STOP escalation.

Reasoning context ignored per M1 Context Isolation.

| Item | Value |
|---|---|
| Measurement tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t229` @ `9cb18cf4b`, branch `WT-audit-verdict-converge` |
| Delta audited | `910f3ffed..9cb18cf4b` — 4 files, +125/−26, SPEC artifacts only (no source changed) |
| Scope | D1 / D2 / D3 / D9 disposition + new-defect sweep on the delta. Items iter1 cleared were NOT re-derived. |
| Live codex probe | NOT run |
| `internal/cli` tests | NOT run — no finding depends on a test run |

**Process note (not a defect):** the Tier S plan-audit ceiling is 1 iteration
(`harness.plan_audit_tier_ceilings`). This iteration is a lead-directed narrow re-audit of an
enumerated defect delta, not an auditor-initiated retry. It does not change the verdict authority.

---

## Must-Pass Results (re-verified at v0.4.0)

- **[PASS] MP-1** — `REQ-CVS-001` … `-004` at `spec.md:119/123/127/131`. Count 4, sequential, no gap or duplicate. REQ bodies unchanged by the delta.
- **[PASS] MP-2** — requirement-layer GEARS unchanged by the delta; iter1's judgment stands. P-CONS was deliberately placed as a §A.5 **property** and an AC criterion rather than a new REQ, so no new requirement text entered the layer.
- **[PASS] MP-3** — 12 canonical fields present; `version: "0.4.0"`, `updated: 2026-08-25` (valid ISO). No rejected snake_case alias.
- **[N/A] MP-4** — single-language SPEC.
- **[PASS] MP-5 D7** — external references still limited to `SPEC-AUDIT-MULTI-MODEL-001` (`status: completed`). No BLOCKING.
- **[PASS] MP-6 D8** — `syscall` count 0 across all three artifacts.
- **[PASS] MP-7** — `grep -rn '\[NEEDS CLARIFICATION'` exits 1.

**Budget** — REQ **4/8**, AC **6/8**. Both inside the Tier S ceilings, applied independently.

---

## Category Scores

| Dimension | iter1 | iter2 | Movement |
|-----------|-------|-------|----------|
| Clarity | 0.75 | **0.78** | D2's false [HARD] premise removed and P-CONS stated crisply with an explicit anti-formulation warning; offset by two NEW false claims (N1, N2) and a duplicated paragraph (N3). Net small gain. |
| Completeness | 0.75 | **0.85** | D1's uncovered case now covered by AC-CVS-006; D4's open item closed by measurement; M2 gained an explicit entry gate. Landing route still unstated (iter1 D8, optional, unrepaired). |
| Testability | 0.80 | **0.83** | AC-CVS-006 measurably kills 6 of the 9 non-correct implementation variants I built; AC-CVS-001's RED expectation corrected. Offset by N1 (a false kill claim), N2 (a false RED claim), and one redundant row (K5). |
| Traceability | 0.75 | **0.75** | Unchanged. REQ-CVS-003 still maps to AC-CVS-004 alone; AC-CVS-006 traces to a §A.5 property rather than a REQ (N7), and the AC-004 ↔ AC-006 relationship is unstated (N4). |

Aggregate = mean(0.78, 0.85, 0.83, 0.75) = **0.8025** ≥ 0.75 (Tier S threshold).

---

## Method — how the discrimination claims were measured

Reasoning about which mutant a witness row kills is exactly the kind of claim that fails silently,
so I measured it. I replicated `codexFindingBullet` (`internal/cli/mcp_codex.go:1115`) and
`codexStatedVerdict` (`:1130`) **verbatim**, added a `codexScoredVerdict` built to `plan.md` §C.3's
narrow contract (line head + uppercase verdict word + space + a 0..1 decimal), then wrote **eleven
implementation variants** — the current two-signal shape, each named mutant (e / f / g and their
plausible sub-shapes), two mis-ranking mutants, and a correct set-max P-CONS implementation — and
ran all eight §B-2 rows through every one.

```
impl \ row                        K1   K2   K3   K4   K5   K6   K7   K8   | killed by
V0  current tree (2 signals)      ok   ok   X    ok   ok   ok   X    ok   | K3,K7
Ve  mutant(e) naive assign        X    X    ok   X    ok   ok   ok   ok   | K1,K2,K4
Ve2 naive, reversed order         ok   ok   X    ok   ok   ok   X    ok   | K3,K7
Vf1 mutant(f) first-in-text       ok   X    X    ok   ok   ok   ok   ok   | K2,K3
Vf2 mutant(f) last-in-text        X    ok   ok   X    ok   ok   X    ok   | K1,K4,K7
Vg1 mutant(g) pair, bullet-last   ok   ok   ok   ok   ok   ok   ok   ok   | SURVIVES ALL
Vg2 mutant(g) pair, early-return  ok   ok   ok   ok   ok   X    ok   ok   | K6
Vh  bullet-needs-stated           X    X    ok   X    X    ok   ok   ok   | K1,K2,K4,K5
Vrk wrong rank (pass>inconcl)     ok   ok   ok   X    ok   ok   X    ok   | K4,K7
Val always-fail                   ok   ok   ok   X    ok   ok   X    X    | K4,K7,K8
Vok correct P-CONS                ok   ok   ok   ok   ok   ok   ok   ok   | SURVIVES ALL
```

```
-- per-row values --
K1  want=fail          V0=fail          Ve=pass          Vg1=fail          Vg2=fail
K2  want=fail          V0=fail          Ve=pass          Vg1=fail          Vg2=fail
K3  want=fail          V0=pass          Ve=fail          Vg1=fail          Vg2=fail
K4  want=inconclusive  V0=inconclusive  Ve=pass          Vg1=inconclusive  Vg2=inconclusive
K5  want=fail          V0=fail          Ve=fail          Vg1=fail          Vg2=fail
K6  want=fail          V0=fail          Ve=fail          Vg1=fail          Vg2=inconclusive
K7  want=inconclusive  V0=pass          Ve=inconclusive  Vg1=inconclusive  Vg2=inconclusive
K8  want=pass          V0=pass          Ve=pass          Vg1=pass          Vg2=pass
```

`Vok` surviving every row and `Val` / `Vrk` dying confirms the corpus is neither vacuous nor
over-constrained. The scratch harness was removed after measurement; the tree is unmodified.

---

## 1. D1 repair — **REPAIRED** (core), with two new claim defects

### The discriminator, applied

| Test | Result |
|---|---|
| Does a **fourth signal** require editing AC-CVS-006's assertion? | **No.** The rule is `max(S)` over the signal set; the assertion is `got == row.want` applied table-driven. A fourth signal needs a new row, not a new assertion. |
| Does adding a **K9 row** require editing the assertion? | **No.** §B-2 states this explicitly ("행만 추가한다. 단언문도, 이 원칙 목록도 바뀌지 않는다"). |
| Is the rule a **set** operation, not an assignment-sequence statement? | **Yes**, and emphatically. `spec.md` §A.5 P-CONS reads "신호 **집합**의 최댓값" and then names the failure mode the lead warned about and forbids it: "'나중 신호가 앞선 신호를 덮지 않는다' 같은 **순서에 관한 규칙으로 쓰면 안 된다** — 그 형태는 신호가 셋인 동안만 유효하고 넷째가 들어오는 순간 다시 뚫린다." `plan.md` §C.5 repeats the prohibition. |
| Is the AC free of implementation-form assertions? | **Yes.** `acceptance.md:164` — "채택값만 단언하며 내부 구현 형태에 관해 아무것도 요구하지 않는다"; `spec.md` §C adds the same constraint for AC-CVS-006. |

### Mutant (e) independence — **VERIFIED**

`Ve` (the naive one-more-assignment M2, in `plan.md` §C.1's stated order) is killed by **K1, K2, K4**
and by nothing in AC-CVS-001…005: it leaves the M1 fall-through correct, so C1–C7 still synthesize
`inconclusive` in adversarial (AC-001 ✓), C8 → `fail` (AC-002 ✓), native no-signal → `pass`
(AC-003 ✓), `Verdict: pass` + bullet → `fail` (AC-004 ✓), and the live-probe body → `inconclusive`
(AC-005 ✓). **AC-CVS-006 is the sole AC that kills it.** The independence claim is not inflated.

The `plan.md` M2 entry gate ("착수 전제: acceptance.md 에 AC-CVS-006 이 테스트로 존재해야 한다") is
the correct structural placement of that guard.

### Per-row witness audit — measured, not accepted

| Row | Claim in §B-2 | Measured | Disposition |
|---|---|---|---|
| K1 | 기본 세탁 반례 (mutant e) | kills `Ve`, `Vf2`, `Vh` | **claim holds** |
| K2 | 순서 무관 (mutant f) | kills `Vf1` where K1 does not — `Vf1` gives K1=`fail`, K2=`pass` on an identical signal set | **claim holds** (K3 is a second witness for `Vf1`) |
| K3 | 반대 방향 세탁 | kills `V0`, `Ve2`, `Vf1` — pins the opposite assignment order | **claim holds** |
| K4 | 순위 중간값 | kills `Ve`, `Vf2`, `Vrk`, `Val` | **claim holds** |
| K5 | **쌍 특수화를 가른다 (mutant g)** | kills `Vh` only — and `Vh` is already killed by K1, K2 and K4. **K5 has no unique kill, and kills neither mutant-(g) variant.** | **claim FALSE** → N1 |
| K6 | 3-신호, 쌍 특수화 (mutant g) | **sole** killer of `Vg2` (the early-returning pair specialization) | **claim holds for the early-return shape only**; false for the bullet-last shape → N1 |
| K7 | 순서 무관 (두 번째 증인) | kills `V0`, `Ve2`, `Vf2`, `Vrk`, `Val`; no unique kill | holds as a *second* witness, but it is a signal-**role swap**, not a text-order swap → N5 |
| K8 | 과잉 보수 방지 | kills `Val` | **claim holds** |

**Net:** the D1 repair achieves what it was commissioned to achieve — the laundering path that
iter1 found is now closed by a genuinely order-independent, signal-count-independent AC. Two of the
table's own labels do not survive measurement (N1), and the RED-baseline claim does not either (N2).

## 2. D2 repair — **REPAIRED**, in both places, with no overcorrection

Every citation the rewrite introduced is exact in this tree:

| Cited | Measured |
|---|---|
| `isBlockVerdict`, `codex_review_gate.go:116-117` | `116: func isBlockVerdict(verdict string) bool` / `117: return strings.HasPrefix(strings.ToLower(strings.TrimSpace(verdict)), "fail")` ✅ |
| `codex_review_gate.go:109` | `109: return allow, nil // pass / inconclusive ⇒ ALLOW` ✅ |
| `mcp_convergence.go:126-129` | the fail-open comment: "when all required backends returned VerdictInconclusive (or a pass/inconclusive mix with no required FAIL), the overall verdict falls back to the claude verdict" ✅ — and it **does** support the claim; the actual condition also covers a pass/inconclusive *mix*, so the SPEC's paraphrase is if anything narrower than the code, never broader |
| `codex_review_rpc_test.go:119` | `119: "": "pass",` ✅ — the empty-string contract the pin preserves |

- No blocking claim is reintroduced. `spec.md` §A.6 now states plainly that flipping the native
  default "정상 변경이 차단되지는 않는다".
- **No overcorrection.** §A.6 closes with "반대 방향으로도 과장하지 않는다 — `fail` 은 여전히 차단하므로
  게이트가 판정값과 무관한 것은 아니다", which is the exact guard the lead asked for.
- Both surfaces were corrected — `spec.md` §A.6 **and** `acceptance.md` AC-CVS-003's mutant (b) note
  — so the two do not disagree. The AC's tag also moved from `[게이트 보호]` to `[보고 정확성]`,
  matching the corrected rationale. `plan.md` §F's risk row was corrected in the same pass.

For the record: iter1 cited `isBlockVerdict` as `117-119` and the allow-return as `:108`. The
v0.4.0 citations (`116-117`, `:109`) are the precise ones.

## 3. D3 — **REPAIRED**

`acceptance.md:97` now reads "corpus 8건 전부(C1~C8)에서 실패해야 한다", `plan.md` M1 carries the
same correction, and a new DoD line requires recording the RED for all eight. Matches iter1's
measurement (C1–C8 all synthesize `pass` on the current tree).

## 4. D9 — **REPAIRED**, both citations verified at HEAD

| Citation | Measured at `9cb18cf4b` |
|---|---|
| `spec.md` §D → `mcp_codex.go:1155` | `1155: Findings:  []Finding{},` ✅ |
| `spec.md` §F → `mcp_convergence.go:367` | `367: type backendCallFn func(ctx context.Context, backend, target, focus, projectRoot string) ReviewOutput` ✅ |

iter1's optional **D4** was also closed by measurement (`plan.md` §C.4 now records that
`review-output.schema.json` exists nowhere in the repository). iter1's D5–D8 remain open and
optional; none was in this iteration's scope.

---

## New Defects (v0.4.0 delta)

**N1.** `acceptance.md:63` (K5/K6 table rows) + `acceptance.md:170` (mutant (g) note) + `plan.md:147`
(안티패턴 3-b) + `acceptance.md:189` (DoD line) — **The claim that K5 and K6 catch pair
specialization is false as stated, and the mutant (g) prose contradicts itself.** The prose says the
pair-specialized implementation "오늘의 세 신호에서는 옳게 동작한다" and then says "K5와 K6이 잡는다";
both cannot hold. Measured: `Vg1` (stated×scored special branch, bullet applied last — the natural
shape) **survives all eight rows**, and correctly so, because the bullet signal only ever contributes
`fail`, the top of the rank, so applying it last is always P-CONS-correct with three signals. K6 does
kill `Vg2`, the early-*returning* variant — a different mutant, one that is NOT correct today. K5
kills neither, and has **no unique kill at all**: the only variant it catches (`Vh`, bullets
conditional on a stated verdict) is already caught by K1, K2 and K4. The DoD line "K5·K6 이 stated ×
scored 특수화 구현을 실제로 가르는지 확인" is therefore **unsatisfiable as written** — half of it
cannot be verified because it is not true. — Severity: **major** — Class: **blocking** — Required
fix: (a) split mutant (g) into its two shapes and state that the bullet-last shape is
**undetectable with three signals and is not a behavioral defect today** — it is a latent structural
risk that only a fourth signal can expose; (b) attribute K6 to the early-return shape only;
(c) relabel K5 for what it actually witnesses (a `scored × bullet` pair where no stated verdict
exists, which catches bullet-application made conditional on `stated`) or drop it; (d) amend the DoD
line to match.

**N2.** `acceptance.md:171-172` ("[중요] 이 AC 는 RED 로 시작하지 않는다") + `plan.md:110`
("가드 AC 이므로 RED 기대 없음") + `acceptance.md:187` (DoD) — **AC-CVS-006 does partially start RED,
so the blanket claim is false.** Measured against the current tree (`V0`, two signals): **K3 →
`pass` (want `fail`)** and **K7 → `pass` (want `inconclusive`)** — both fail today, because with no
scored signal only `stated` is read, and in both rows `stated` is the *lenient* member of the set.
The result is independent of whether M1's fall-through fix is applied, since `stated` matches in both
rows. The claim is true only of K1/K2/K4 (the mutant-(e) witnesses), which is the substantive point
the note wanted to make. The note's own closing sentence warns against asserting an unmeasured RED
expectation; the blanket form makes the mirror-image unobserved claim. Practical consequence: the DoD
instructs the run phase **not** to record a RED baseline for this AC, so two genuinely-red rows would
arrive unexplained — and the cheap wrong response is to "correct" the expectations to match the
current tree, which would silently delete K3 and K7's discriminating power. — Severity: **major** —
Class: **blocking** — Required fix: restate as "K1/K2/K4/K5/K6/K8 are already green today; **K3 and
K7 are RED today** (measured: K3 → `pass`, want `fail`; K7 → `pass`, want `inconclusive`) because the
third signal does not yet exist", and amend the DoD line to require recording exactly that two-row
RED set.

**N3.** `acceptance.md:171-172` — the `[중요] 이 AC 는 RED 로 시작하지 않는다` blockquote appears
**twice**, near-verbatim; the two differ only in "K1 은" vs "첫 번째 Given 은". Both were added by
this delta. — Severity: minor — Class: optional — Required fix: delete one (the surviving copy needs
N2's correction anyway).

**N4.** `acceptance.md` AC-CVS-004 vs AC-CVS-006 — **the relationship is unstated.** AC-004's verdict
assertion (`Verdict: pass` body + a `[P1]` bullet ⇒ `fail`) is precisely the `stated × bullet`
instance of P-CONS, so it is a special case of AC-006's general rule. They are not redundant —
AC-004 additionally pins the `SynthesisNote` record and the `converge` propagation, which AC-006 does
not touch, and `stated × bullet` is a pair that §B-2 **does not contain** (K5 is `scored × bullet`,
K6 is three-signal). But a reader cannot tell which AC owns the conservative-adoption rule. —
Severity: minor — Class: optional — Required fix: one sentence in AC-CVS-004 stating that its verdict
half is the `stated × bullet` instance of P-CONS (owned generally by AC-CVS-006) and that AC-004 owns
the recording and convergence assertions; or move the pair into §B-2 as a K row and leave AC-004 the
recording only.

**N5.** `acceptance.md:61` (K7 row label) — labelled "순서 무관 (두 번째 증인)", but K7 is not a
text-order variant of K4: it swaps **which signal carries which value** (K4 = stated `inconclusive` ×
scored `pass`; K7 = scored `inconclusive` × stated `pass`). The row's own parenthetical
("담는 신호만 뒤바꿈") is accurate; the summary column is not. The only true text-order pair is K1/K2,
which §B-2's own construction principle requires. — Severity: minor — Class: optional — Required fix:
relabel K7 as a signal-role-swap witness.

**N6.** `acceptance.md` §B-2 table + AC-CVS-006 — the expected value is **transcribed by hand** per
row rather than derived from the signal-set column. §B-2 defines it as derived
("기대값은 언제나 집합의 최댓값이며 손으로 예외를 두지 않는다"), but nothing requires the test to
compute it, so a future K9 row carrying a mistaken `want` would pass silently — the same
self-fulfilling shape §B warns about for format lists. — Severity: minor — Class: optional —
Required fix: require the test to compute the expectation from the signal-set column via a
test-local rank helper. This constrains the *test*, not the implementation, so it does not violate
`spec.md` §C's implementation-independence rule.

**N7.** `acceptance.md:156` — AC-CVS-006 traces to `spec.md §A.5 P-CONS`, a named SPEC property,
rather than to a `REQ-CVS-*`. Under a strict reading (every AC references a valid REQ) it is an
orphan AC. The choice is deliberate and its rationale is stated — P-CONS was kept out of the
requirement layer on purpose — so this is recorded as a traceability observation, not a
misplacement. — Severity: minor — Class: optional — Required fix: none required; optionally note in
§B or §C that AC-CVS-006 is bound to a §A.5 property by design.

---

## Delta items verified clean (no defect)

- **C7 → C5 mode-wiring witness move.** The consequence is stated correctly: `runTurn` short-circuits
  empty review text at `mcp_codex.go:702-703` (verified — `if reviewText == "" { return
  inconclusiveReview(...), errors.New(...) }`) before the synthesizer, so C7 is unreachable in
  production and C5 (reachable prose) is the right witness. C7's retained native-`pass` pin **does
  not** contradict AC-CVS-001: AC-001 binds the adversarial path, AC-003 the native path, and the
  mode split is the point. The pin is additionally required by `codex_review_rpc_test.go:119`.
- **K8 vs AC-CVS-001.** No conflict. §B members are unrecognized-format by construction; §B-2 rows
  are recognized-format by construction, so K8's `pass` expectation cannot collide with "no §B corpus
  member is `pass`".
- **`plan.md` M2 entry gate.** Correctly placed and correctly motivated.
- **`plan.md` §C.4** — the `mcp_convergence.go:134` C3-invariant citation is unchanged and still exact.
- **Budget** — 4 REQ / 6 AC, both inside the Tier S ceilings.

---

## Recommendation

**PASS-WITH-DEBT**, score **0.8025**, moved monotonically up from iter1's **0.7625**.

The character of the debt has changed, and that matters more than the score. iter1's D1 was a hole
that let a **wrong implementation land green** — a §0 violation the SPEC would have introduced itself.
That hole is closed: AC-CVS-006 is a genuine set-based property, it is order- and signal-count-
independent by the stated discriminator, and it is measurably the only AC that kills mutant (e).
iter2's remaining debt is entirely about the **accuracy of claims made about the tests**, not about
the tests' power. Nothing in N1–N7 permits a wrong implementation to pass.

Two items are owed before run-phase entry, both cheap and both text-only:

1. **N2 before M1's RED is recorded.** K3 and K7 are red today; the DoD currently says not to look.
   Fix the claim and the DoD line together, or the run phase records a baseline it was told not to
   expect and may "repair" the two rows that carry the discriminating power.
2. **N1 before M2's entry gate is evaluated.** The gate asks a run-phase actor to verify that K5 and
   K6 discriminate pair specialization. Half of that cannot be verified because it is not true, and
   an unsatisfiable gate item either blocks or gets waved through — both bad outcomes.

N3 should ride the same edit (it is a duplicated paragraph). N4–N7 are optional and left to the
orchestrator's discretion; routing all of them into a third revision would cost more than it buys.
