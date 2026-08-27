# acceptance.md — SPEC-GITFLOW-DOCTRINE-ALIGN-001

Verification layer for the three-document git-flow doctrine alignment. Paths abbreviations used below:

- `$D` = `.moai/docs/git-workflow-doctrine.md`
- `$L` = `.moai/docs/git-local-workflow-doctrine.md`
- `$R` = `.claude/rules/local/repo-local-pr-policy.md`

All RED-now observations were executed in THIS session on THIS tree, base d29b8942e (baseline-integrity attribution; outputs quoted verbatim in research.md §B–§D). Two-cell discipline (verification-completeness.md §2) applies: every AC below states its observed-red input and the milestone that flips it.

## §D AC Matrix

| AC | Requirement | Verifies | Severity | RED-now | Flipped by |
|----|-------------|----------|----------|---------|------------|
| AC-GDA-001 | REQ-GDA-002 | D1 (line 15 rationale prose) | Must | RED (measured) | M1 |
| AC-GDA-002 | REQ-GDA-003 | D2+D4 (❌ develop bullets gone) | Must | RED (2 hits measured) | M1 |
| AC-GDA-003 | REQ-GDA-003 | Replacement prohibitions present in both lists | Must | RED (0 hits measured) | M1 |
| AC-GDA-004 | REQ-GDA-001, -002 | D3 (§18.3.1 annotated + restated) | Must | RED (0 × `[RETIRED 2026-08-27]`) | M1 |
| AC-GDA-005 | REQ-GDA-004 | D5+D9 (both §23.7 routing bullets scoped) | Must | RED (L150 un-annotated; L151 `항상 …` always-PR clause live) | M2 |
| AC-GDA-006 | REQ-GDA-004 | D6 (§23.9 re-scoped + preserved) | Must | RED (0 markers) | M2 |
| AC-GDA-007 | REQ-GDA-005 | D7+D8 (origin-premise rewrite) | Must | RED (Route-B line live; 0 branch-model `develop` statements — only incidental `manager-develop` substring @ L11) | M3 |
| AC-GDA-008 | REQ-GDA-006 | Scope fence holds | Must | GREEN at plan time (vacuously) — flips to meaningfully PASS after M1-M3 touch the targets | M4 |

Requirement coverage: REQ-GDA-001 → AC-004 (and partially AC-002/003 negatives); REQ-GDA-002 → AC-001, AC-004; REQ-GDA-003 → AC-002, AC-003; REQ-GDA-004 → AC-005, AC-006; REQ-GDA-005 → AC-007; REQ-GDA-006 → AC-008. Every requirement traces to ≥1 AC.

## AC-GDA-001 — Rationale prose carries dated retirement

- **Given** the working tree at any commit containing `.moai/docs/git-workflow-doctrine.md`.
- **When** running: `grep -n "장점이 부재" .moai/docs/git-workflow-doctrine.md`
- **Then** PASS iff EVERY returned line either contains the dated marker form `[RETIRED 2026-08-27]` on itself OR is bracketed by `~~…~~` OR sits inside a `>` blockquote opened by a `[RETIRED 2026-08-27]` header no more than 3 lines above; additionally PASS requires ≥1 operative statement in the same region asserting the reversal (grep `"뒤집혔다"` finds a hit at or below the header notice).
- **Command form**: `grep -n "장점이 부재" $D | grep -vE '\[RETIRED 2026-08-27\]|~~|>'` → must output nothing (exit-safety note: inspect printed lines, do not rely on bare exit codes through pipes).
- **RED-now evidence**: single hit at line 15, no marker anywhere nearby (verbatim in research.md §B/D1).
- **Green path**: M1 step 1 wraps the paragraph as retained-as-history and appends the operative correction.
- Mutant probe: deleting the sentence entirely ALSO passes the negative half — accepted, because plan.md M1 directs annotation-preferred replacement and the deletion is visible diff-wise; the positive half (≥1 reversal statement in region) still binds whether annotated or rewritten.

## AC-GDA-002 — No forbidden-list bullet declares develop creation prohibited

- **Given** the aligned tree.
- **When** running: `grep -nE '❌.*develop.*브랜치 생성' .moai/docs/git-workflow-doctrine.md`
- **Then** zero matches.
- **Baseline attribution**: same command today → exactly 2 hits, lines 52 and 351 (d29b8942e). Count-equals-zero is adopted WITH its red-on-record (2), satisfying the empty-set proof obligation.
- **Green path**: M1 steps 2–3 replace both bullets in place.

## AC-GDA-003 — Corrected prohibitions stand in the lists

- **Given** the aligned tree.
- **When** running the pair:
  - `sed -n '/금지 사항/,/^### §18.1/p' .moai/docs/git-workflow-doctrine.md | grep -cE '^[- ]*❌.*(main.*분기|main.*생성)'` → ≥ 1
  - `sed -n '/^### §18.10/,/^### §18.11/p' .moai/docs/git-workflow-doctrine.md | grep -cE '^[- ]*❌.*(main.*(분기|시작)|카드 단위 PR)'` → ≥ 1
- **Then** both ranges contain ≥1 ❌ bullet naming a CURRENT true prohibition (card branch from `main`; card-level PR into `main`), each referencing `develop` as the correct origin.
- **Red-on-record**: both counts are 0 today (the only ❌ …develop constructs are the retired ones counted in AC-GDA-002).
- **Green path**: M1 steps 2–3 exact drafts in plan.md.

## AC-GDA-004 — §18.3.1 annotated and restated for the new model

- **Given** the aligned tree.
- **When** running:
  - `sed -n '/### §18.3.1/,/^### §18.4/p' .moai/docs/git-workflow-doctrine.md | grep -c '\[RETIRED 2026-08-27\]'` → ≥ 1
  - same range: `grep -cE 'develop.*(merge|병합|통합)|(merge --no-ff)'` → ≥ 1
- **Then** the section carries the dated retirement marker AND at least one current-rule statement expressing develop-based card integration.
- **Red-on-record**: `[RETIRED 2026-08-27]` count over ALL THREE targets is 0 today (measured via `grep -c` batch); the section sweep yields 7 pre-transition PR-routing premises (research.md §B/D3).
- **Green path**: M1 step 4.

## AC-GDA-005 — §23.7 routing bullets scoped to the surviving truths

- **Given** the aligned tree.
- **When** running BOTH:
  - check (a): `sed -n '/^### §23.7/,/^### §23.9/p' .moai/docs/git-local-workflow-doctrine.md | grep -n '모든 tier' | grep -vE '\[RETIRED 2026-08-27\]|~~'`
  - check (b): `sed -n '/^### §23.7/,/^### §23.9/p' .moai/docs/git-local-workflow-doctrine.md | grep -n '항상' | grep -vE '\[RETIRED 2026-08-27\]|~~|^>'`
- **Then** NEITHER returns output — every surviving "모든 tier … PR 경유" claim AND every surviving `항상 … 흐름` route-prescription is annotated, struck, or inside a blockquote — AND the same range contains ≥1 `[HARD]` bullet stating the main direct-push prohibition remains (grep `enforce_admins` → ≥1 hit in range).
- **Red-on-record** (both measured this session on d29b8942e): (a) single surviving hit at file line 150, un-annotated; (b) single hit at file line 151 (`- [HARD] \`\`git push origin main\`\` 금지 — 시도 시 server-side rejected. 항상 feat/fix/chore/docs/release 브랜치 → gh pr create → CI green → gh pr merge 흐름.`), un-struck. Check (b) was added at plan-audit iter-1 (auditor D1); its red observation postdates the original battery — recorded here as its first RED-now cell rather than claimed otherwise.
- **Green path**: M2 step 1's two sub-bullets — L150 scoped rewrite; L151 keeps the true server-side-rejection half intact and strikes ONLY the route-prescription tail with a dated annotation, replacing it with release-path-only phrasing.

## AC-GDA-006 — §23.9 retire-and-preserve

- **Given** the aligned tree.
- **When** extracting `sed -n '/^### §23.9/,/^### §23.8/p' .moai/docs/git-local-workflow-doctrine.md` (note: §23.8 physically follows §23.9 in the file) and then:
  - counting the marker via FIXED-STRING search over the extracted range: `sed -n '/^### §23.9/,/^### §23.8/p' .moai/docs/git-local-workflow-doctrine.md | grep -F -c '[RETIRED 2026-08-27]'` → ≥ 1 (the `-F` flag is mandatory: bare `[RETIRED …]` parsed as a regex bracket-expression is an invalid character range under BSD grep/ugrep — plan-audit iter-1 D2). An equivalent hit from an escaped regex form also passes.
  - checking the lead sentence `모든 경우 PR 생성·머지는` does not appear UNLESS its line matches `[RETIRED 2026-08-27]|~~|^>`
  - finding a current-routing statement mentioning both `develop` (card integration, no PR) and `release/vX.Y.Z` → `main` PR (merge commit)
- **Then** all three sub-checks hold.
- **Red-on-record**: marker count 0 across the target files (measured); routing-flow items 1–3 all premised on per-change PRs (read directly at lines 171–175).
- **Green path**: M2 step 2.

## AC-GDA-007 — Origin-premise rule rewritten

- **Given** the aligned `.claude/rules/local/repo-local-pr-policy.md`.
- **When** running the triple:
  - `grep -n 'ALL tiers (S / M / L) use' $R` → every hit line (if any) contains `~~` (struck historical record); normally ZERO hits after rewrite
  - statement-pattern greps are the PASS criteria: `grep -nE 'branch(es)? FROM .?develop' $R` → ≥ 1 AND `grep -cE 'merge --no-ff' $R` → ≥ 1 (a branch-from statement and a merge-into statement each present). The raw occurrence count `grep -cE 'develop' $R` is INFORMATIONAL ONLY — it reads **1** today via the incidental identifier `manager-develop` at line 11 (re-measured post-audit), so it discriminates nothing and is not a pass criterion.
  - `grep -nE 'enforce_admins: true' $R` → ≥ 1 AND `grep -nE 'MUST NOT.*PR|NO card-level PR' $R` → ≥ 1
- **Then** the file prohibits main direct push in substance, states the card model, keeps the neutrality paragraph (grep `NOT mirrored` → ≥ 1).
- **Red-on-record**: `ALL tiers` line live at line 10 un-struck; branch-model `develop` statements = 0 — no line today asserts branching from or merging into `develop`; the file's only `develop` substring is the incidental identifier hit noted above (`grep -cE 'develop' $R` = 1 @ L11; wording corrected from "occurrences = 0" per plan-audit iter-1 D3). The 14-line file was inspected in full.
- **Green path**: M3 full draft carried verbatim in plan.md.

## AC-GDA-008 — Hard scope fence

- **Given** the integrated card branch BEFORE `origin/develop` push.
- **When** running:
  - `git diff --name-only d29b8942e -- CLAUDE.local.md internal/template/templates .github/workflows .claude/rules/local/gitflow-lane-protocol.md` → empty output
  - `git diff --name-only d29b8942e -- .moai .claude/rules/local` → every listed path is either one of the three targets or under `.moai/specs/SPEC-GITFLOW-DOCTRINE-ALIGN-001/` or `.moai/state/verify/`
- **Then** nothing outside the dispatched scope moved.
- **Pin note**: d29b8942e is a FIXED tree SHA, not a moving ref — a rebase onto newer `develop` invalidates this pinning and REQUIRES re-measurement of every baseline (verification-completeness.md §4).
- **Green path**: M4 battery run at card close.

## §D.1 Severity & grading

All eight criteria are **Must** — a violation of any one fails the card. There are no Should/May-tier criteria: a docs-alignment card is binary at the grep level by construction.

## §D.2 Indirect verification

Beyond the grep battery, the human-audit lens: read §18.0.1, §18.10, §23.7, §23.9 top-to-bottom after M1–M3 and confirm no sentence in standing (un-annotated) position contradicts gitflow-lane-protocol.md §§1–4. The battery catches every MEASURED defect; this read catches NEW contradictions the battery didn't anticipate. Findings from that read loop back as amendment candidates, not silent fixes.

## §D.3 Closure gates

- G1: All eight ACs PASS on the final tree (battery output captured at `.moai/state/verify/<session>/t310-ac-battery.txt`).
- G2: Zero unresolved clarification-gate markers across plan.md / research.md — the MP-7 letter-scan must find no non-self-referential occurrence of that marker idiom anywhere in the artifact set (paraphrased token-free per plan-audit iter-1 D5, so this gate no longer trips on its own home file).
- G3: Scope fence clean (AC-GDA-008 both sub-checks).

## §D.4 Forward-looking checks

- If a later card repairs research.md §F-1..F-5, it MUST re-measure at its own base SHA and may reuse this SPEC's grep idiom.
- If the operator ever restores an explicit `--pr` card hatch (F-6), amend REQ-GDA-004/GDA-005 rather than editing bodies outside a SPEC.

## Definition of Done

Card t310 closes when: AC battery fully green on this branch's tip; findings report cites battery evidence path; card merged into local `develop` per lane protocol §2–§4 and pushed to `origin/develop`; `origin/develop` CI green serves as the integration verdict (documentation-only diff expected to move docs-neutral workflows only).
