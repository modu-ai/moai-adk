# t344 verdict — prediction-ledger metric inversion promoted to general form

Card: t344 (G2a, lane-9) · Branch: WT-audit-verdict-owner · Base: local develop e85c55fa9

## Claim

The card's pre-work clause (t241 verdict: "read §A.5's current text and judge — promote the format to general form, or treat it as already closed") is adjudicated: **promote to general form**. The corrected metric convention — predictions written as defects SURVIVING TO ADOPTION, audit findings only as an inverse-sign auxiliary indicator, every expectation stating which of the two it measures — now lives in the doctrine that governs verification-artifact authoring (`.claude/rules/moai/development/verification-completeness.md` §6, byte-identical in both trees), instead of being trapped inside one SPEC's plan artifact.

## Evidence

1. **The inversion, measured**: `SPEC-VERIFICATION-COMPLETENESS-001/plan.md §A.5` — VC-2/4/5/6 predictions written as "감사 지적 0건"; the ledger's own defect note states the metric RISES as the audit layer works and instructs "다음 장부는 예측을 채택까지 살아남은 건수로 쓰고" — but that instruction exists only inside this plan.md.
2. **The "already closed" reading tested and REFUTED**: t241's C4 row claims "t333 이 채택함 — 발생/채택생존 표를 SPEC에 싣고". Not corroborated by any measurable artifact: (a) `grep "발생|생존"` over the SPEC dir → only the defect note itself and an unrelated D6 usage; (b) `git log -- plan.md` → two commits only (1d4f881ee t241, 7f5b6a947 t261), no t333 commit; (c) card t333's own reports (`.moai/reports/t333/{verdict,trigger-axis-observation}.md`) are SPEC-GUARD-LIVENESS-001 work with zero 장부/C4/생존 mentions (grep → 0 hits); (d) the rule file carried no survival convention pre-edit (grep → 0 hits). An "adopted" claim with no artifact is an unobserved claim (VCI §1.1 surface 3/4) — the already-closed reading cannot stand on it.
3. **The fix**: new `## 6. Prediction-ledger metrics — count survival, not detection` in `.claude/rules/moai/development/verification-completeness.md` [HARD] — predictions as survival-to-adoption counts, never "zero audit findings"; occurrence as inverse-sign auxiliary only; per-expectation labelling. Applied to BOTH trees (template source first per Template-First, then local mirror — the pair is byte-parity, `diff -q` rc=0 post-edit). Full diff: `.moai/reports/t344/section6-diffstat.txt` (68 lines, 2 files, identical hunk in each).
4. **Firing surface measured**: the rule's `paths:` frontmatter includes `**/.moai/specs/**` — a future SPEC author writing a prediction ledger loads this rule (frontmatter read directly, head -8). This is the exact audience the trapped plan.md note could never reach.
5. `make build` → EXIT=0.

## Baseline-attribution

All measurements this run, this tree (WT-audit-verdict-owner @ e99d7f7fe + working tree):
- inversion text: `sed -n '100,140p' .moai/specs/SPEC-VERIFICATION-COMPLETENESS-001/plan.md` (this run).
- refutation sweeps: 4 greps + 1 git log named in Evidence item 2, all rc observed this run.
- parity: `diff -q` both copies → IDENTICAL, rc=0, this run.
- build: `make build` → exit 0.

## Gaps (explicitly NOT observed)

- The SPEC plan.md §A.5 itself was NOT edited (SPEC body = manager-spec's domain per Status Transition Ownership Matrix; the ledger's defect note already instructs the correction at next-ledger authoring, and the general form now reinforces it from the rules layer). Editing those prediction cells would need a manager-spec delegation — judged out of scope for this card.
- No future ledger exists yet to observe the convention being applied — firing is by paths-scope loading, proven structurally (item 4), not behaviorally.
- The discrepancy between t241's C4 note and the artifacts is reported as measured; adjudicating WHY the note says t333 (typo vs lost adoption) is not decided here — the lead/operator owns that follow-up.

## Residual-risk

- The convention is doctrine-layer: an author who skips the paths-scoped rule (e.g., authors a ledger in a doc the scope does not match) writes outside its reach. Accepted — the ledger convention is a verification-artifact authoring concern, which is exactly the scope the rule keys.
- If the "t333 adopted" note later turns out to reference a real artifact neither of us found, this card's addition is still not redundant: that artifact (whatever it is) was invisible to a direct searcher, while the rules-layer convention is load-bearing at authoring time.
