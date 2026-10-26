# Implementation Plan — SPEC-V3R6-DOCS-PREREWRITE-README-001 (card t1720, Tier M, docs-only)

## §A Context

- **Baseline tree.** Commit `cb30e05b3e5f64774b8dabd87e80c24332f4551d`, tree object `9c08830bdec23b81334840307ee745e468500174` (acceptance.md §A). All RED-now evidence is in acceptance.md §C, measured on that tree on 2026-10-26; nothing in this plan cites an umbrella figure without a fresh `R-` cell behind it.
- **Artifacts.** `spec.md` (requirements, decisions Q1/Q8/Q13/Q17 + QR-1..QR-3, edit surface), `plan.md` (this file), `acceptance.md` (AC matrix + evidence ledger).
- **Parent.** SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (umbrella, split by `d-20261014T055241Z-1bc6` / `d-20261014T060405Z-0fcd`); pre-split REQ/AC/decision texts imported by path and SHA from `bdb7b764`.
- **Siblings.** GATE (t1717) owns the cited-verbs producer; FLAG (t1726) owns flag citations; SWEEPA–SWEEPF (t1721–t1725, t1727) own docs-site page files. No `depends_on`: this child lands independently; its scoped checks do not need the GATE producer.

### §A.5 PRESERVE list (scope fence — REQ-README-004)

Everything outside: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `docs-site/hugo.toml` (lines 55–56 only), `docs-site/content/ko/core-concepts/what-is-moai-adk.md`, `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-README-001/`, `.moai/reports/t1720/`. In particular: the seven already-linking ko pages, every other docs-site page, `.github/workflows/**`, `internal/**`, `cmd/**`, `docs-site/vercel.json`, `assets/**`, and every tag and release on the remote (no `git tag`, no release creation, no push).

## §B Milestones

Decisions most likely to change come first. M2+M3 are ONE commit (REQ-PRC-003's same-commit clause).

- **M1 — SPEC baseline + RED-now evidence (this commit).** The three plan-phase artifacts land as one commit on the card branch. Evidence cells R-00 to R-27 (acceptance.md §C) are the baseline; the closing checks re-run them. Priority: High.
- **M2 + M3 — Q8 badge rollback and Q13 citations, one commit with the hugo.toml change.** Files: the four READMEs + `docs-site/hugo.toml`.
  - Badge (M2a, REQ-PRC-003): line 24 of each README, `Release-v3.2.0-blue.svg` → `Release-v3.1.2-blue.svg`.
  - hugo.toml (M3, same commit): line 55 `version = "v3.1.2"`, line 56 `releaseDate = "2026-08-21"` (v3.1.2's measured publish date, R-20c).
  - Q13 citations (M2b, REQ-README-001, decision QR-3): per locale, line 58 (drop the `moai cg` clause; `moai gpt` clause already names the registered `moai codex` equivalent — keep the sentence accurate without the flagged tokens) and the GPT-gateway callout (ko 291 / en 296 / ja 290 / zh 289 — delete the flagged `moai gpt` citation; the callout's dated withdrawal notice may remain with the citation removed, per Q13's default).
  - Verification: R-20a/R-20b re-run green at `v3.1.2`/`2026-08-21`; same-commit property shown by the parent-diff assertion `git diff <commit>^ <commit> -- README.md README.ko.md README.ja.md README.zh.md docs-site/hugo.toml` exhibiting all six line changes (4 badge + 2 hugo) in ONE diff (D3 repair; a file listing alone is not the check); the scoped E-01 count drops by the two non-paragraph lines per file (58, callout) with the paragraph lines remaining for M4.
  - Priority: High.
- **M4 — Q1 CG paragraphs and the what-is link.** Files: the four READMEs + `docs-site/content/ko/core-concepts/what-is-moai-adk.md`.
  - README.ko.md 386 and 591 (R-08r): each paragraph becomes a one-line link sentence to `/ko/multi-llm/cg-mode/` (REQ-README-002, Q1).
  - en/ja/zh counterpart paragraphs (en 387+592, ja 385+590, zh 584+589 — R-01n): each becomes a link sentence to that locale's cg-mode page (decision QR-2 default; these lines are flagged and must change for AC-PRC-001-r to close).
  - `what-is-moai-adk.md`: add the link; under the QR-1 default, replace its own sentence copy with the link sentence (link instead of repeat).
  - Verification: R-08r exits 1 with no output; R-01 scoped exits 1 with no output (all 16 flagged lines resolved across M2+M4); R-08b on what-is ≥ 1.
  - Priority: High.
- **M5 — verification sweep (sync).** Re-run every ledger cell on the closing tree (single-invocation form, fresh verbatim outputs, tree SHA pinned); run `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` from the worktree root (gap GR-5); confirm the diff fence with `git diff --name-only` against the M1 baseline listing only the §A.5 surfaces; record the fresh external `gh release list` observation with its UTC time (regression-guard, not a pass); land the sync/landing hand-off per the card flow (the lane does not merge its own card). Priority: Medium.

## §C Known issues / constraints

- **No release operations.** Q8's resolution labels existing state only. Creating a v3.2.0 release, publishing drafts, or tagging anything is out of scope (REQ-README-004; the "3.2 rollback" is recorded, not repaired, by this card).
- **Same-commit clause.** A commit that changes the badge lines but not hugo.toml lines 55–56 (or vice versa) violates REQ-PRC-003 — plan M2+M3 as one commit, never two.
- **Shared files.** The four READMEs are shared with the FLAG child (t1726) on different lines; the what-is page family is shared with the SWEEP children on different lines. Keep the diff line-scoped so the siblings' merges do not conflict.
- **Non-flagged lines are frozen.** REQ-README-004 leaves every non-flagged README line byte-identical, with the named exception of the line-24 Release badge of the four READMEs (authority: REQ-PRC-003 / Q8; D1 repair); drive-by prose fixes (including adjacent factual claims) are out of scope.
- **The lane does not merge its own card.** The landing hand-off follows the card flow (`moai factory complete` PR edge); no hand-run PR, merge, or push.

## §D Risks

- **Line-number drift between M1 and run phase.** The per-locale line numbers in R-01n are pins, not anchors; if a sibling lands first and shifts lines, re-run R-01n on the then-current tree before editing. The pattern, not the line number, defines the flagged set.
- **Locale wording quality.** The link sentences are new per-locale prose; they follow the README 4-locale sync contract (Korean primary, en/ja/zh derived) so the four files stay in section and phrasing parity.
- **Planner defaults unresolved.** Q1, Q13, Q17, QR-1..QR-3 carry planner defaults, not operator verdicts (acceptance.md GR-6). A run-phase deviation from a default is a blocker report, not a local rewording.

## §E Self-verification (run-phase §E, per manager-develop template)

At completion, report per the attribution triple (command / verbatim output / tree SHA):
1. AC matrix (AC-PRC-001-r, AC-PRC-003 guard, AC-PRC-004-r) with each closing cell's verbatim output.
2. Diff fence: `git diff --name-only <M1-baseline>..HEAD` listing only the §A.5 surfaces.
3. Same-commit property: parent-diff assertion `git diff <M2+M3-commit>^ <M2+M3-commit> -- README.md README.ko.md README.ja.md README.zh.md docs-site/hugo.toml` showing all six line changes (4 badge + 2 hugo) in ONE diff (D3 repair).
4. i18n strict check output (M5).
5. Fresh external release observation with UTC time (regression-guard record).

## §F Cross-references

- Umbrella: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (index; pre-split text at `bdb7b764`).
- i18n rules: `.moai/docs/docs-site-i18n-rules.md` lines 61, 65 (the Q17 obligation, outside the authority register).
- Card evidence path: `.moai/reports/t1720/`.

## §G Revision history

| Date | Revision | Author | Change |
|------|----------|--------|--------|
| 2026-10-26 | 0.1.0 | manager-spec | Initial plan-phase authoring (see spec.md HISTORY). |
| 2026-10-26 | 0.1.1 | manager-spec | Repair per plan-audit verdict FAIL 0.88 at commit `5fe7d20f46` (card t1720; verdict at `.moai/reports/t1720/plan-audit-verdict.md`, gitignored). M2+M3 verification and §E.3 same-commit check replaced with the parent-diff assertion (D3); §C line-freeze constraint now names the REQ-PRC-003/Q8 badge exception (D1). No milestone added, removed, or reordered. |
