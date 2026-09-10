# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.2.1 — applies the reduced plan-audit iteration 1 verdict (FAIL 0.71, `.moai/reports/t622/plan-audit-reduced-iter1.md`), the operator's OD-2 flag-semantics sub-decision, and the lead's per-mode approval ruling (both 2026-09-11, via lead). 0.2.0 was committed as `24e20ec50`; the verdict on top as `eea2be13b`.
- **Tier**: M kept; judged REQ 12 and AC 15 are within 16/16, affected files 17 exceed the Tier M file guidance (reported to the lead; spec.md §C.4). **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: HEAD `eea2be13b`; the eight scope files (local and template), `manager-git.toml`, the `/moai sync` command sources, the guarding test files, `docs-site/content`, and `Makefile` have no diff against `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` (`git diff --stat` empty, exit 0), so plan-time counts on these copies are base counts.
- **Audit count**: reduced iteration 1 FAIL 0.71; this revision is for reduced iteration 2. Full-scope iterations before the split: FAIL 0.67 and FAIL 0.75.
- **Pre-write self-checks**: SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → `PASS` (re-run for 0.2.1 in the return). Frontmatter 12 canonical fields present.
- **Numbering**: original IDs kept with one-line placeholders for moved (REQ 007-010, 016-023; AC 007-010, 017-024) and withdrawn (011, 012) numbers; new REQ-GDP-025 and 026, new AC-GDP-026 to 029.
- **Counts**: REQ 26 numbers — judged 12 (001-006, 013-015, 024-026), moved 12, withdrawn 2. AC 29 numbers — judged 15 (001-006, 013-016, 025-029), moved 12, withdrawn 2.
- **Scope added in 0.2.1**: `.claude/skills/moai/SKILL.md:140`, `.claude/skills/moai/references/reference.md:161`, `.claude/skills/moai/workflows/sync/quality-gates-context.md:30` and `:101`, `.claude/skills/moai/workflows/sync.md` Flags line (local 114 / template 104), plus the per-mode rule in `manager-git.md` PR Auto-Merge section and `delivery.md` Step 3.4 (local and template).
- **Consumer sweep** (`/usr/bin/grep -rn -e '--merge' -e '--no-merge' -e '--auto-merge' .claude internal/template/templates/.claude internal/template/templates/.agents docs-site` → 167 lines): in scope — the decided lines above plus `delivery.md` 337-338/348-349 and `manager-git.md` 148/166. Blockers (outside the decided list but required for consistency) — X1 `workflows/sync.md` usage line (local 95 / template 85), X2 `.claude/commands/moai/sync.md:3` and template `sync.md.tmpl:3` argument-hint, X3 `delivery.md:404`, X4 docs-site `workflow-commands/moai-sync.md` and `core-concepts/what-is-moai-adk.md` in en/ko/ja/zh. Out of scope (different meaning) — `gh pr merge --merge` (hns-release-specialist.md, moai-ref-git-workflow SKILL.md:134, docs-site worktree/examples.md), `--merged-only`, `git branch --merged`, `docs-site/README.md:348` `moai update --merge`. Published `.agents/skills/moai-sync/SKILL.md` (local and template) contains no `merge` (`grep -i merge` exit 1).
- **New-file parity, neutrality, Frozen (plan time)**:
  - Local/template diff: `quality-gates-context.md` exit 0; `moai/SKILL.md` exit 1 (20 hunks, 125c125 … 273,274c273,274, 392d391; line 140 outside); `references/reference.md` exit 1 (229d228); `workflows/sync.md` exit 1 (65,74d64, 81c71).
  - Tests naming the new files check other properties, not local/template parity: `backlog_json_disclosure_mirror_test.go:24` (embedded = template source, SKILL.md), `template_neutrality_audit_test.go:141` (SKILL.md C2 allowlist), `agent_frontmatter_audit_test.go:399`/`:407` (frontmatter), `agentless_audit_test.go:44` (sync.md listed as implementation skill). Mirror allowlists contain none of them. Guard per new file: diff only; template copies are walked by `TestTemplateNoInternalContentLeak`; neutrality of their added template lines is judged by AC-GDP-015 over `git diff $BASE -- $T/`.
  - `[ZONE:` lines in the four new files: 0 (local and template). zone-registry.md entries naming them: 0 (local and template). Positive control — entries naming agent-common-protocol.md: 13 (local and template).
- **Base counts and controls for new or changed ACs (template copies)**:
  - AC-GDP-001 toml: `.toml` `## Synchronization` section diff against template manager-git.md section exit 0 (11 lines), so the toml paragraph and autofail counts equal the md counts (1, 1).
  - AC-GDP-006 broadened detector: delivery Step 3.4 section hits at lines 8 and 20 (file 337, 349). Audit mutant fixture (4 lines) → lines 2 and 3 hit, correct line 4 not hit.
  - AC-GDP-014: `.toml` `## Synchronization` (11 lines) and `## PR Auto-Merge` (9 lines) sections diff against template manager-git.md exit 0 and 0. Mutant (toml section with "AND all approvals obtained" replaced by "once checks pass") → diff exit 1.
  - AC-GDP-026: fragments — skill 1 line, reference 3, qgc-args 4, qgc-flags 6, sync 1, delivery 52; `--auto-merge` count 0 in all six; violating `--merge` lines 7 (skill 1, ref 2, qgc-args 4, qgc-flags 4, sync 1, delivery 9 and 20). Mutant fixture (5 lines) → lines 1, 2, 3 hit; correct alias line 4 and `--merged-only` line 5 not hit.
  - AC-GDP-027: `--no-merge` lines only in delivery fragment (8, 19); (i) violations 8 and 19; (ii) violations 8 ("NOT set") and 19 ("Skip"). Mutant fixture (4 lines) → (i) lines 1, 3; (ii) lines 1, 2, 3; correct line 4 not hit by either.
  - AC-GDP-028 (detector narrowed to "team mode" after "team" matched "teammates"): manager-git PR Auto-Merge section 9 lines; (a) 0 and 0 (manager-git, delivery); (b) 0. Mutant fixture (6 lines) → (a) lines 2, 3; (b) line 2; the personal/manual line mentioning teammates is not hit.
  - AC-GDP-029: (a) 0 and 0; (b) 0. Same fixture → (a) lines 4, 6; (b) line 4; line 5 (manual missing) does not satisfy (a).
  - AC-GDP-025 registry control: 13 per copy (see above).
- **Carried from 0.2.0 (unchanged base)**: AC-GDP-001 paragraphs 1 / autofail 1 / listgroup 0; AC-GDP-002 block 9 lines, fetch-without-rev-list 1, standalone rev-list at block line 5, joined 0, rev-list 1, session command 1, table rows 8; AC-GDP-004 squash 2, merge_method 0; AC-GDP-005 six `--squash` lines (manager-git 32/114, delivery 343/355, toml 26/108); AC-GDP-006 doc-execution subsection hit at line 8, `manager-[g]it[.]md` mentions 0/0; AC-GDP-013 diffs manager-git 0, acp 0, delivery 275c275/278c278/479,480c479, doc-execution 138,143d137; AC-GDP-015 fixture 5 letter / 2 digit tokens; AC-GDP-025 four clause counts 1 per copy, Frozen mutant caught.
- **Post-write verification (0.2.1)**:
  - MP-1: `**REQ-GDP-NNN**` definitions equal REQ-GDP-001..026 exactly once each (diff against the generated sequence exit 0). Judged 12, moved placeholders 12, withdrawn 2. AC `### AC-GDP-NNN — ` headings 15. `### Out of Scope —` H3 5.
  - AC-GDP-015 controls on spec.md: SPEC-ID lines 8, REQ lines 50, date lines 15, SHA-shaped tokens 54, `:980ccdc56` 1.
  - AC-GDP-025 Frozen-line judge corrected during verification: the first grep no longer carries `-n` (the line-number prefix would have made `^[-+]` match nothing). Fixture diff with removed, added, and context `[ZONE:Frozen]` lines → 2 hits, context line not hit.
- **plan_status**: audit-ready (blockers X1-X4 reported to the lead)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
