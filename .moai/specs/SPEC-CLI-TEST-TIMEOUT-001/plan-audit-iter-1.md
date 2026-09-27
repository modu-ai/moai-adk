# Plan-Audit Report — SPEC-CLI-TEST-TIMEOUT-001 (Iteration 1)

Auditor: plan-auditor (independent, fresh-judgment). Tree: worktree `t1253`, branch `WT-cli-test-duration`,
HEAD pinned at `b4f798dccc8b2cf3951edea5f62e2b34740f633c` (verified equal to the SPEC's attributed baseline).
Tier M — PASS threshold 0.80 (harmonic mean), must-pass firewall active.

## Verdict

**FAIL** — 0.83 (harmonic mean above the 0.80 Tier M threshold, but must-pass MP-2 Completeness
fails on D1; the firewall overrides the score).

## Claim

The SPEC's arithmetic, measurement attributions, GEARS conformance, traceability, and AC
testability are sound and independently verified. The SPEC's central scope claim — "on every
sanctioned local test entry point" — is **not** delivered as written: `make ci-local` reaches
`go test -race -count=1 -short ./...` (scripts/ci-mirror/lib/go.sh:25) with the 10m default and
is neither covered nor documented as excluded.

## Evidence

All commands run in this audit, this tree, HEAD `b4f798dcc`. Verbatim outputs:

1. **Local measurement figures reproduce from the raw stream** (`.moai/reports/t1253/go-test-cli.json`, 8.15MB):
   - `jq -s '[.[] | select(.Action=="pass" and .Test != null)] | length'` → `7594`
   - skip count → `50`
   - package pass event → `{"Elapsed": 1118.093}` (exit 0)
   - thresholds → ge1s `399`, ge5s `42`, ge10s `7` (exact match to spec.md §A.1 / measure-meta.txt)
   - top-25 Elapsed sum → `278.58` (claimed 278.6 = 25%: 278.58/1118.093 = 24.92%)
   - pass-only leaf sum → `1444.66` (claimed 1444.7; the 1450.94 pass+skip figure differs only by skip events carrying 6.28s — measure-meta's label is pass-only and accurate)
   - parallel factor → 1444.66/1118.093 = 1.292 ≈ 1.29x ✓
2. **CI comparator figures verified against the upstream artifacts** (stronger than measure-meta attribution alone):
   - `gh api repos/modu-ai/moai-adk/actions/runs/36228023389` → `{"conclusion":"success","head_sha":"b4f798dccc8b2cf3951edea5f62e2b34740f633c","name":"CI","status":"completed"}` — run exists, succeeded, head equals the baseline pin.
   - `gh run download 36228023389 -n test-stream-ci-race-ubuntu-latest` + jq → `RACE internal/cli Elapsed: 885.287` ✓
   - `-n test-stream-ci-test-ubuntu-latest` + jq → `NORACE internal/cli Elapsed: 308.601` ✓
3. **D1/D2/D3 derivations recompute**:
   - D1: 885.287 × 3.62 = 3204.7 ≈ 3205s = 53.4 min ✓; 30m (1800s) insufficient ✓; 3600/3204.7 = 1.1235 ≈ 1.12x ✓
   - Cross-check: 885.287/308.601 = 2.8683; 1118.093 × 2.868 = 3206.7 ≈ 3207s ✓ (convergent)
   - D2: 1800/1118.093 = 1.6098 ≈ 1.61x ✓
   - Amplification: 1118.093/308.601 = 3.6233 ≈ 3.62x ✓; default-600s exceeded by 518.1s ✓
4. **RED-now verified on the pinned tree**: `Makefile:104` (`test`), `:107` (`test-verbose`), `:110` (`test-codex-live`), `:194` (`test-race-short`) and `CLAUDE.local.md:265` (§4), `:394` (§6 HARD line) — none carries `-timeout`. AC-001/AC-002 are genuinely RED at baseline.
5. **CI quote accurate**: `.github/workflows/ci.yml:308-311` carries verbatim "Go's 10m default kills internal/cli, whose -race runtime measures 546-1183s" and `-timeout 20m`; `release-pr-multi-os.yml:210` carries `-race -timeout 25m ./...`. REQ-SCOPE-008's premise holds and no requirement touches CI.
6. **The missed entry point (D1 defect)**:
   - `Makefile:160-161`: `ci-local: ## Run CI mirror locally (lint + vet + test + cross-compile)` → `@./scripts/ci-mirror/run.sh`
   - `scripts/ci-mirror/run.sh`: dispatches per-language modules from `lib/`
   - `scripts/ci-mirror/lib/go.sh:25`: `go test -race -count=1 -short ./... || exit 2` — **no `-timeout`**
   - `scripts/ci-mirror` is tracked and NOT template-distributed (`ls internal/template/templates/scripts/ci-mirror` → absent), i.e. a repo-local developer surface — squarely inside the SPEC's declared territory (it edits two other repo-local surfaces).
   - Hazard: CI race `internal/cli` alone is 885.287s > the 600s default **before** any local amplification; whether `-short` shaves it below 600s is unmeasured (the SPEC itself discloses "no short-mode duration measurement exists" in REQ-TIMEOUT-003). `make ci-local` is help-listed and one command away from the t1171 panic class.
7. **Un-inventoried go-test surfaces (D2 defect)**: `Makefile:180`/`:184` (`tui-snapshot`, `tui-snapshot-verify`), `Makefile:39,48,52,59,67` (agents-emit / commands-emit / tool-policy golden checks), `scripts/ac-baseline/check-staged.sh:26`. All practically fast (single `-run`-filtered or small-package), but the SPEC never inventories or dismisses them, so "every" is unverified by its own text.
8. **Structure checks**: frontmatter carries all 12 canonical fields + `tier: M` (no rejected aliases); 9 REQs, sequential REQ-TIMEOUT-001..006 / REQ-DOC-007 / REQ-SCOPE-008 / REQ-COORD-009, all GEARS-conformant ("shall" ubiquitous ×6, When-triggered ×1, "shall not" unwanted ×1 — spec.md:95-103); all 4 `### Out of Scope —` H3 headings carry bullets; SPEC-HEAVY-TEST-SLOT-001 exists with `status: completed` (no D7 blocking); no `[NEEDS CLARIFICATION]` markers in any artifact; progress.md §E.1 present and correctly shaped (`plan_status: audit-ready`, baseline_head, evidence paths).

## Baseline-attribution

- Every figure above was measured **in this audit run, against this tree** at HEAD `b4f798dccc8b2cf3951edea5f62e2b34740f633c` (verified via `git rev-parse HEAD`), or against the GitHub artifact set of run 36228023389 whose `head_sha` equals that same pin. The local raw stream, measure-meta.txt, and the upstream CI artifacts are mutually consistent with zero discrepancies.

## Dimension Scores (harmonic mean, rubric-anchored)

| Dimension | Score | Basis |
|-----------|-------|-------|
| MP-1 Clarity | 0.95 | GEARS clean on all 9 REQs; every -timeout value carries a derivation that recomputes (Evidence 3). "sanctioned" undefined (D4). |
| MP-2 Completeness | 0.60 | 4 named Makefile targets + 2 CLAUDE.local.md recipes covered; rejections recorded with evidence; coordination premises present; RED-now verified. BUT ci-local/ci-mirror entry point missed entirely (D1, blocking), tui-snapshot*/golden/ac-baseline surfaces un-inventoried (D2), title's "every" claim unmet. |
| MP-3 Testability | 1.00 | All 8 ACs binary-testable; AC-004 bounded (one package, 35m probe ceiling, slot-serialized, single compound invocation, exit-0 + no-panic + green count); mutant probes specified for AC-001/AC-008. |
| MP-4 Traceability | 0.90 | All 9 REQs have ≥1 AC; all AC mappings resolve to existing anchors; frontmatter valid. AC-004→"§A" and AC-007→"§C" anchor to sections, not REQ tokens (D3, minor). |

Harmonic mean = 4 / (1/0.95 + 1/0.60 + 1/1.00 + 1/0.90) = **0.83**.

## Gaps (what this audit did NOT observe)

- A direct local **race** run of `./internal/cli/` or `./...` (the 60m value's projected basis) — the SPEC discloses this; I did not run one either (load discipline; the M3 probe is run-phase work).
- A `-short`-mode duration measurement for `internal/cli` — so I cannot establish whether `make ci-local`'s race-short full-suite run actually exceeds 600s today, only that it is unmeasured and unprotected.
- The CI figures were verified against run 36228023389's artifacts (downloaded and recomputed) — I did not re-run CI myself.
- t1252's TestMain fix state and t1219's eventual scope: both cards are external to this tree; the coordination premises were checked for presence and internal coherence only (AC-006's own bar).

## Residual-risk

- The 3.62x amplification is a single-sample figure taken under declining, non-stationary load (27.07 → 13.02, with 4 foreign `go test` processes alive at launch per measure-meta). The SPEC handles this correctly — projection disclosed in plan.md §B, thin 1.12x headroom acknowledged in D1, REQ-TIMEOUT-006 re-derivation trigger — so the residual is operational (a >60m local race panic triggers re-derivation) rather than a defect.
- If D1 is resolved by documented exclusion rather than coverage, the panic hazard in `make ci-local` remains live in the tree; the exclusion text should say so plainly.

## Defects Found

**D1. (BLOCKING — must-pass MP-2)** — `spec.md` §A/§B.2/§E; `Makefile:160-161`; `scripts/ci-mirror/lib/go.sh:25`.
`make ci-local` reaches `go test -race -count=1 -short ./...` with no `-timeout` — a help-listed, first-class local entry point in the exact defect class the SPEC claims to close (§A: "The LOCAL surfaces still relying on the 10m default are the defect class this SPEC closes"). CI race `internal/cli` is 885.287s > the 600s default even before local load amplification, and no `-short` measurement exists to establish safety (the SPEC's own REQ-TIMEOUT-003 disclosure). The surface appears nowhere in the SPEC — not in requirements, not in Out of Scope.
**Required fix (either)**: (a) add coverage — one more requirement + M1 line putting `-timeout 60m` (D1 class; note the `-short` nuance REQ-TIMEOUT-003 already carries) + derivation comment on the `go.sh:25` line per REQ-DOC-007, and extend AC-001's target list; or (b) add `### Out of Scope — ci-local (scripts/ci-mirror)` to §E stating the exclusion rationale (§6-discouraged surface, unmeasured -short class) and reword §A's closure claim so it no longer asserts the whole defect class. File: spec.md.

**D2. (SHOULD-FIX — MP-2)** — `spec.md` §E; `Makefile:39,48,52,59,67,180,184`; `scripts/ac-baseline/check-staged.sh:26`.
Seven additional go-test-carrying local surfaces (tui-snapshot, tui-snapshot-verify, the agents-emit/commands-emit/tool-policy golden checks, ac-baseline staged check) are neither covered nor inventoried-and-dismissed. All are practically fast (single `-run`-filtered or small-package), so exclusion is defensible — but the title's "every" claim requires the inventory step, which is absent.
**Required fix**: add an entry-point inventory (table or §E bullet) naming these surfaces with the dismissal basis (e.g., "-run-filtered single-test invocations; 600s default unreachable by measurement class"). File: spec.md.

**D3. (MINOR — MP-4)** — `acceptance.md` §D.2.
AC-004 → "§A problem closure" and AC-007 → "§C" anchor to document sections rather than REQ tokens. No orphan (every REQ is covered), but the two ACs have no requirement-level owner.
**Required fix**: either annotate the two cells as intentional section-anchors, or add one requirement (e.g., REQ-VERIFY-010: "The SPEC shall close only after a bounded post-change re-verification completes exit 0") to own AC-004. File: acceptance.md.

**D4. (MINOR — Clarity)** — `spec.md` §A/title.
"sanctioned" is never defined; the entire scope boundary rides on the word, which is how D1 fell out.
**Required fix**: one sentence in §A defining the sanctioned set (e.g., "Makefile targets whose go test invocation runs a measured class that can exceed the 600s default, plus the CLAUDE.local.md §4/§6 recipes"). File: spec.md.

**D5. (MINOR)** — `spec.md` §D; `CLAUDE.local.md:532`.
§13 carries another `go test ./...` mention outside §6; the SPEC defers "§6 full-suite command lines" to t1219 but is silent on §13's line's ownership.
**Required fix**: one clause in §D extending (or explicitly excluding) §13's line from t1219's stated scope. File: spec.md.

## Recommendation

FAIL on must-pass MP-2. The fixes are small and additive: D1 is either one covered recipe line or one documented exclusion; D2 is one inventory paragraph; D3-D5 are single sentences. None disturbs the verified measurement base, derivations, or AC machinery — iteration 2 should be a cheap confirming re-audit scoped to the defect delta (D1 first). The arithmetic and evidence chain in this SPEC are the strongest I have verified in this audit stream; the defect is purely a scope-boundary omission around the word "sanctioned."
