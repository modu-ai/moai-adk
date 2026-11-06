# Progress — SPEC-V3R6-DOCS-PREREWRITE-README-001 (card t1720)

Card t1720 (Class C, docs-only, tier M, lane-14). Branch `WT-readme-ko-cg`, worktree `.moai/worktrees/t1720`.
Base tree `cb30e05b3e`. Full lane evidence chain (plan audit iters, Kickoff, waits, watchdog): `.moai/reports/t1720/progress.md` (runtime record, gitignored).

## §E.1 Plan-phase Audit-Ready Signal

- plan artifacts: spec.md 0.1.2 / plan.md / acceptance.md revision 0.1.2, commits `5fe7d20f46` → `0412f03771` → `7f42ba8c36`
plan_status: audit-ready
audit_ready: true
- plan_complete_at: 2026-10-26T13:25Z
- audit chain: iter1 FAIL 0.88 (D1–D3, `rcpt-20423e764fa5dbfdf0b82938`) → repair `0412f03771` → iter2 FAIL (convergence, D4 lost receipt) → iter3 FAIL 0.88 (D4 resolved via `rcpt-bbd227b78fd24413a7a2389c`, D5 found) → repair `7f42ba8c36` → iter4-delta **PASS 0.88** (blocking 0, ≥ Tier M 0.80; verdict `.moai/reports/t1720/plan-audit-verdict-iter4-delta.md`, gitignored)
- receipts backing the PASS: `rcpt-a62c3d55bd93105e07f1689d` (iter4 auditor's own codex leg, D5 repair commit scope) + `rcpt-5d46314d25f3c294ab256a60` (lane codex_audit pass, delta_base `5fe7d20f46`)
- Kickoff: autonomous form (auto-semantics §9.1), decision record in `.moai/reports/t1720/progress.md` (2026-10-26T13:25Z)

## §E.2 Run-phase Evidence

Run closed on HEAD `2956448055` (2026-10-26T13:40Z). Two forward commits, no push, no tag:

- `9655ffb355` — README 4-locale (ko/en/ja/zh): stale CLI-citation cleanup (Q13), CG-retirement paragraphs replaced with canonical `https://adk.mo.ai.kr/<locale>/multi-llm/cg-mode/` link sentences (Q1/Q17), `what-is-moai-adk` cg-mode link added (Q1, QR-1..QR-3)
- `2956448055` — Release badges `v3.1.2` ×4 (README line 24 each) + `docs-site/hugo.toml` `version = "v3.1.2"` / `releaseDate = "2026-08-21"` in the SAME commit (REQ-PRC-003 same-commit clause)

Exit-verification, all cells GREEN at `2956448055` (measured in this run, this tree; lane spot-checks re-observed at the same head in `.moai/reports/t1720/card-review.md`):

| Cell | Observed |
|---|---|
| R-01 | absent-CLI-forms grep across 4 READMEs → empty, exit 1 (GREEN) |
| R-08r | scoped CG-sentence grep → empty, exit 1 (GREEN) |
| R-08b-r1 | `git grep -c -F "](https://adk.mo.ai.kr/ko/multi-llm/cg-mode/)" 2956448055 -- README.ko.md` → `2956448055:README.ko.md:2`, exit 0 (GREEN) |
| R-08b-r2 | what-is cg-mode link = 1 |
| R-20a | `Release-v3.1.2` 4× (1 per README), exit 0 |
| R-20b | hugo.toml v3.1.2 / 2026-08-21 |
| AC-PRC-003-r | `git diff 2956448055^ 2956448055` carries all six line changes in ONE diff |

hns-oss-docs-verify recipe (inlined): hugo build warning-free; sitemap OK; URL blacklist 0; Mermaid 0; 4-locale file-existence + section parity; README 4-file H2 parity 12/12/12/12; added body emoji 0.

Gaps (recorded, bounded):
- GR-A — 7 out-of-scope ko docs-site pages keep their CG-paragraph copies per spec; umbrella AC-PRC-004 full closure is outside this child (umbrella closes it).
- GR-B — pre-existing stale examples / emoji elsewhere: outside the edit surface (plan §A.5 PRESERVE).
- GR-C — repaired (sync-audit F1; the earlier "absent from the tree" claim was false — stale skill digest): `scripts/docs-i18n-check.sh` IS git-tracked (base, run tree, and HEAD). Strict run re-executed in this repair run: `DOCS_I18N_STRICT=1 bash scripts/docs-i18n-check.sh` → exit 0, "Errors:   0 / Warnings: 0 / OK: all 4 locales pass parity, frontmatter, H1, and glossary checks." The inlined verify recipe ran in addition, not as a substitute.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
audit_ready: true
run_complete_at: 2026-10-26T13:40Z
- closing tree: `2956448055` (`295644805581277c9b1a6e855849fcc7f7a8dfa2`)
- card-review: codex `mcp__moai__codex_review` scope=card (base `cb30e05b3e`, head `2956448055`) → **pass, 0 findings** — receipt `.moai/reports/t1720/card-review.md` (gitignored)

## §E.4 Sync-phase Audit-Ready Signal

Single sync commit (this commit) carries the 3-phase close: spec.md frontmatter `status: draft → completed` (merged `draft → implemented → completed` per the adjacent run/sync close) + `updated: 2026-10-26`; stages only `progress.md` + `spec.md` of this SPEC directory. No CHANGELOG (release-time artifact; release lens forbids it), no codemaps, no README/docs-site edits, no push, no tag.

sync_complete_at: 2026-10-26
sync_commit_sha: "96f443c9e9c8518a76dd3058280d11e914c412cf"
sync_status: complete (3-phase close — run 9655ffb355 + 2956448055, sync close 96f443c9e9, backfill afd26932a6, post-absorb HEAD cb36e91e76)
b12_self_test_a: not-applicable — CHANGELOG emission forbidden by the release lens for this card (no `[Unreleased]` entry written; pre-emission grep therefore not run)
frontmatter_status_transitions:
  spec_md_status: draft → completed (merged 3-phase close, single sync commit, owner manager-docs)
  updated_refreshed_to: 2026-10-26

Audit-ready statement: plan-audit final PASS stands on receipts `rcpt-a62c3d55bd93105e07f1689d` + `rcpt-5d46314d25f3c294ab256a60` (verdict `.moai/reports/t1720/plan-audit-verdict-iter4-delta.md`); card-review stands on `.moai/reports/t1720/card-review.md` (pass, 0 findings, head `2956448055`); run exit-verification all-green at closing tree `2956448055`. sync-auditor to run next (not run within this commit). `sync_commit_sha` is the canonical `pending-backfill-sync` placeholder — a commit cannot cite its own hash; backfill with the real SHA in the next forward commit on this branch (spec-frontmatter-schema.md § SHA placeholder backfill exemption).
