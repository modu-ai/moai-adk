# Progress — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audited-pass
audit_ready: true
plan_audit_iteration_3: PASS 0.91, verdict file `.moai/reports/t1725/plan-audit-verdict-iter3.md`, audited_sha ed1512b869a820c8ff33eb248eab447a0119a402, plan_artifact_hash aa66a04ecd01b76ee264ead4c52fc42d152f8a33fe24725f12f39396041aca81, receipt rcpt-e182b42500cf048189d6bdbd (re-runs rcpt-ef41fab8f889725087eb1f75, rcpt-34eef81c2d33487bc250a9c0)
plan_revision: 0.3.0
plan_commit: 6d335cd483 (feat(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan-phase artifacts (tier M, 3 artifacts), card t1725)
plan_followup_commits: b28da6f891 (progress.md and decision-index.md), bb7d928f68 (cite REQ in ACs, QE-1 relabel)
revision_1_commit: e85847fd33 (fix(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan revision 1 after audit FAIL 0.69 (D1-D12), card t1725)
revision_commit: c9ec672020 (revision 2, fix(SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001): plan revision 2 after audit FAIL 0.84 (D-N1..D-N7), card t1725; full c9ec672020afdb5300408fa58fe5ace443fe1f4e; backfilled by the follow-up commit because a commit cannot cite its own hash)
card: t1725 (Class C, docs-only), lane-26, run tn1p4e
parent: SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (child row SWEEPE; decisions Q13, Q17, Q24/N2)
measurement_pin: 74ad197380a0 (the tree the plan-phase counts were taken on; full 74ad197380a0711cdb9816d4e37c94d29af730ea; docs-site, .claude, internal, cmd, and scripts are identical to HEAD e85847fd33, acceptance.md S-00)
plan_audit: iteration 1 FAIL 0.69 (tier M threshold 0.80, MP-8 failed), verdict file `.moai/reports/t1725/plan-audit-verdict.md`, audited_sha bb7d928f682d34279d5d3584537fec0c3e68dd25, receipt rcpt-93c8df0b6680dfc7dac83fb9; iteration 2 FAIL 0.84 (above the 0.80 bar, but the required codex gate returned fail; blocking D-N1 to D-N3, optional D-N4 to D-N7), verdict file `.moai/reports/t1725/plan-audit-verdict-iter2.md`, audited_sha e85847fd331b14320d4825338d2923830f4d1bb0, receipt rcpt-90f562481259842169400426; revision 0.3.0 answers D-N1 to D-N5 and D-N7 and leaves D-N6 as recorded planner defaults; iteration 3 (delta on D-N1 to D-N3) not yet run
open_decisions: see decision-index.md. Q13, Q17, and Q24 carry operator verdicts (operator records d-20261019T015908Z-220a for Q13 and Q17, d-20261014T060405Z-0fcd for Q24). QE-1 to QE-6 carry DEFAULT-APPLIED planner verdicts (QE-1 to QE-5 at 2026-10-19T02:27:27Z, QE-6 at 2026-10-19T02:51:00Z, manager-spec); the leader records the default application of all six on the decision board before Kickoff (iteration-2 defect D-N6).

### Plan-phase measurements (source: .moai/reports/t1725/progress.md, manager-spec, 2026-10-19, pin 74ad197380a0; revision 0.3.0 re-measured every cell it touched, acceptance.md § D)

- Closing pattern (umbrella E-04) on the twelve page files: 10 rows, 18 lines (faq x4 `/moai init`; moai-feedback x4 `fb|bug|issue`; en foundations line 39 `relationship`; en verification-claim-integrity line 84 `brain`). The ja and zh foundations files print no row. Line 31 of each moai-feedback page (`# Short aliases` and translations) is not a match; it is the block label line (QE-6, REQ-SWEEPE-004).
- Complement census (`-P`, CJK-safe boundary): the same 18 lines; no unknown word beyond the six.
- Router Priority 1 (SKILL.md lines 69-70, repo and template copies): `init` is an alias of `project`, `fb` of `feedback`; `bug`, `issue`, `relationship`, `brain` are not. So the FAQ claim that `/moai init` "does nothing" is false, and the umbrella's word "unregistered" is looser than router resolution (QE-3).
- `moai brain` (CLI form, en verification-claim-integrity line 84) exits 1 with `Unknown command` on a tree build (rebuilt in revision 0.3.0: stdout empty, stderr 654 bytes, exit 1); umbrella families E-01..E-03 miss it. E-01..E-03 are 0 in the page files. Every other inline `moai <word>` token resolves (S-06b, S-06c parent-keyed, S-06d-1 to S-06d-5).
- Neither en faq line 24 nor en verification-claim-integrity line 84 carries an example marker; no HTML comment exists in either file.
- ja foundations line 40 and zh foundations line 40 translate the edited English clause; the ko row has no `/moai` clause (S-07, QE-4). The linked `commands.md` headings share no wording across the four locales (en 238 "Where the /moai Command Sits", ko 238 "/moai 명령은 어디에 얹혀 있나", ja 178 "...との関係", zh 178 "...的 /moai 命令的关系"); accepted residual risk (acceptance.md § G).
- Tree build VCS stamp (478285ab4c1f, modified) differs from HEAD; not used as provenance.
- Operator record d-20261019T015908Z-220a read through `moai decision read --scope card:t1725` (revision 0.1.0): Q13 replace-else-delete, Q17 four locales.
- Content fence (revision 0.3.0): fixture repositories `repo` and `repo2` (under `.moai/state/t1725-ctl/`, removed before the commit) hold the pin blobs, a `good` branch with exactly the pinned edits, and mutant branches. Observed: the iteration-2 same-count overwrite mutant keeps numstat `2 4` and fails S-18-en, S-16-en, and S-17 (control C-2); the English-in-every-locale mutants pass the earlier single-alternation S-12a and S-14 and fail every per-locale cell (control C-5); the parent-keyed S-06c kills `moai spec queue`, `moai feedback setup`, and `moai profile audit` (control C-6); the marker-anywhere mutant of the QE-2 alternate fails S-H1a (control C-10).
- Traceability (revision 0.3.0): the Group 4 awk verb was refused by the worktree-session guard and not retried; the two readings were taken with plain commands (plan.md § F): 7 requirement definitions, 11 mapping rows, no uncovered requirement, no orphan. Derived, not the verb's `COLLECTED:` line.
- `moai spec lint` on spec.md (tree build, revision 0.3.0): `0 error(s), 0 warning(s)`, exit 0, one INFO `OwnershipTransitionUnmeasured` for the first plan commit 6d335cd483 (no `Authored-By-Agent` trailer; a historical commit that is not rewritten).
- `spec_audit` MCP tool (`filter_spec` this SPEC, `project_root` this worktree), revision 0.3.0: era V3R6 (H-5), one INFO finding (`EraAutoDetected`), `modern_era_clean: 1`, `total_specs: 1`, no MUST-FIX finding.

### Plan-phase gaps carried forward

- The card's named source evidence (`.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md`) is absent; the file and line list is verified against the current tree only.
- The Q13 and Q17 authority anchors are operator records, not committed files; under the authority register they are not citable as DECIDED or POLICY-COVERED, so those rows stay FOUNDER with the verdict recorded.
- The green paths of the content cells (S-11, S-12a-xx, S-14-xx, S-16-xx, S-17, S-18-xx) were observed on fixtures, not on the real pages; the docs-site pages are run-phase work (acceptance.md G-11, G-13).
- The Group 4 traceability verb and the CN-4 ordering verb were not run (guard refusal of inline `awk`); plain-command readings stand in for them (acceptance.md G-15).
- The native register of the ja, ko, and zh pinned prose (including the new labels) was not reviewed by a native reader (acceptance.md G-8).

## §E.2 Run-phase Evidence

Run agent: manager-develop (cycle_type tdd adapted to docs: RED cells, edits, GREEN cells). Base tree for the RED run: HEAD bdb77bdb0b (clean). Edits committed in 8f86c6bece (full 8f86c6bece4195f52897d4049297ddbd00c7fefc), which also carries the `draft -> in-progress` transition on spec.md (`status:` only; `updated:` already 2026-10-19). Every GREEN cell ran with `HEAD` = 8f86c6bece after that commit, in this worktree, against this tree. Pin for the fence cells: 74ad197380a0. Every cell below is one `git` invocation followed by `; echo "exit=$?"` outside the cell; stdout is verbatim.

Pre-flight S-00 (`git diff --stat 74ad197380a0 HEAD -- docs-site .claude internal cmd scripts`): stdout empty, exit 0 (no docs-site, .claude, internal, cmd, or scripts change since the pin).

| Cell | RED (before edits, HEAD bdb77bdb0b) | GREEN (HEAD 8f86c6bece) | Status |
|------|-------------------------------------|--------------------------|--------|
| S-01 closing pattern, 12 files | 10 rows (en foundations 1, en vci 1, faq x4 1 each, feedback x4 3 each), exit 0 | stdout empty, exit 1 | PASS |
| S-02 complement `-P` | the same 10 rows, exit 0 | stdout empty, exit 1 | PASS |
| S-03 no `slash-example` marker | (not run RED; precondition green at the pin) | stdout empty, exit 1 | PASS |
| S-05a / S-05b / S-05c CLI families | (precondition green at the pin) | each stdout empty, exit 1 | PASS |
| S-06a `moai brain` | `HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`, exit 0 | stdout empty, exit 1 | PASS |
| S-06b first-level verb outside the set | `HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`, exit 0 | stdout empty, exit 1 | PASS |
| S-06c parent-keyed sub-verb | (precondition green at the pin) | stdout empty, exit 1 | PASS |
| S-11 `> /moai:feedback` at its position | stdout empty, exit 1 | four rows, `HEAD:docs-site/content/{en,ja,ko,zh}/utility-commands/moai-feedback.md:32:> /moai:feedback`, exit 0 | PASS |
| S-12a-en / -ja / -zh pinned FAQ line | each stdout empty, exit 1 | one row each at line 24 with the pinned line, exit 0 | PASS |
| S-12a-ko pinned FAQ line | stdout empty, exit 1 | one row at line 28 with the pinned line, exit 0 | PASS |
| S-12b false claim banned | four rows (`faq.md:1` in en, ja, ko, zh), exit 0 | stdout empty, exit 1 | PASS |
| S-14-en pinned foundations row | stdout empty, exit 1 | `HEAD:docs-site/content/en/claude-code/foundations/_index.md:39:` + pinned row, exit 0 | PASS |
| S-14-ja / S-14-zh pinned row | each stdout empty, exit 1 | one row each at line 40 with the pinned row, exit 0 | PASS |
| S-15 retained fragment | stdout empty, exit 1 | `HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`, exit 0 | PASS |
| S-18-en / -ja / -ko / -zh pinned label | each stdout empty, exit 1 | one row each at line 31 (`# Direct slash command`, `# 直接スラッシュコマンド`, `# 직접 슬래시 명령`, `# 直接斜杠命令`), exit 0 | PASS |
| S-13a names (`git diff --name-only 74ad197380a0 HEAD -- . ':!<SPEC dir>' ':!.moai/reports/t1725'`) | stdout empty, exit 0 | exactly the twelve page paths in the AC-PRC-001f order, exit 0 | PASS |
| S-13b numstat (same pathspec) | stdout empty, exit 0 | exactly the twelve rows of AC-PRC-001f (`1 1` x8, `2 4` x4 for the feedback pages), exit 0 | PASS |
| S-16-en/-ja/-ko/-zh whole-line diff | stdout empty, exit 0 | the four pinned blocks verbatim (one hunk `@@ -31,4 +31,2 @@` each; new blob ids 2bd283a9d1a3, 15386d1976fe, 27c00aab4302, dbc9b9f89eb6), exit 0 (run as one invocation over the four paths) | PASS |
| S-17 blob ids of the twelve page files | the twelve unedited ids (d7d6bdc523c9..., 1223289b8580..., 62fd34466239..., a74cfdaac50b..., 94c5b7edf351..., 1e0a41308691..., 87ac8379f286..., e8df4954906e..., bc6086e4c65d..., efd8920f97a9..., 7293a1c812d1..., e542d848c744...), exit 0 | the twelve pinned ids, in order: 99b8de197d71ff79c0a9f827e9bbdc1404f919cf, 62c87efeaf51e300860e9c1ed7c1c3961dd5ccf3, e7de4ec667bf81235612c50acaa5f31d1eb54fe7, 2bd283a9d1a3ab5dd64b170a7115f20bc03fede3, f6fe617f76157513b1df8ff9cf6e3de59d575d19, a137a94f2cbe66f7172efdcc66cc83a67b082af5, 15386d1976fecb9509a3ae481ebd7620cd27d53f, 42ef2ea422f35066220b7700cab58b6a347260d1, 27c00aab43024fe9126e8c219600c327967a8de9, 8d10d0cfef812cc65289d8b8d5d75c56d62b25c4, 26a422462ea47b4fbbed38867fe03966433ac500, dbc9b9f89eb65728dadccee5417c0e200c156c5d, exit 0 | PASS |
| S-10 `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` (plain bash script, no Go invoked) | (precondition green at the pin) | `Errors:   0`, `Warnings: 0`, `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.`, exit 0 | PASS |

Acceptance-criterion roll-up: AC-PRC-001a (S-01) PASS; b (S-02) PASS; c (S-05a to S-05c) PASS; d (S-11, S-18-xx) PASS; e (S-10) PASS; f (S-13a, S-13b, S-16-xx, S-17) PASS; g (S-03) PASS; h (S-06a to S-06c) PASS; i (S-12a-xx, S-12b) PASS; j (S-14-xx) PASS; k (S-15) PASS. No deviation from the pinned text was needed.

Gaps (not observed in this run):
- S-06d-1 to S-06d-5 (tree-build verb surface) were not re-run: heavy Go execution was frozen by the leader (no `go build`), and no moai binary built from this tree was used; the verified verb sets of S-06b and S-06c rest on the plan-phase observation of S-06d (acceptance.md § D).
- Hugo was not built; the rendered output of the replacement lines was not observed.
- The ja, ko, and zh pinned prose (including the new labels) was not reviewed by a native reviewer (acceptance.md G-8); S-10 checks parity, frontmatter, H1, and glossary only, not register.
- Tool provenance (verification-claim-integrity 2.2): the git cells use the installed `git`; S-10 is a shell script run from this tree. No moai build was invoked, so no judging-build commit applies.
- The sibling FAQ bullet about `moai plan` ("does nothing") stays untouched by decision QE-5; recorded as a follow-up observation only.
- Whole verbatim stdout and exit codes of every cell are in `.moai/reports/t1725/run-evidence.md` (gitignored, machine-local); the table above carries the deciding lines.

Residual risk: the ja and zh foundations rows now say "位置づけ" and "定位" while their linked pages' headings still say "関係" and "关系" (accepted residual, decision QE-4). The CI run on the project's integration branch owns the repository-wide test verdict; it is PENDING at report time.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-19
run_commit_sha: 8f86c6bece4195f52897d4049297ddbd00c7fefc
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (docs-only card; no fetch performed, HEAD read before commit: bdb77bdb0b on WT-getting-started-faq)
l44_post_push_fetch: not-applicable (no push; lanes do not push)
new_warnings_or_lints_introduced: 0 (S-10 strict: Errors 0, Warnings 0)
cross_platform_build:
  linux: not-applicable (docs-only)
  windows: not-applicable (docs-only)
total_run_phase_files: 12 page files + spec.md status line + this progress.md
m1_to_mN_commit_strategy: single combined commit for M1 and M2 (8f86c6bece), then this evidence commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
audit_ready: true
sync_status: complete
sync_complete_at: 2026-10-19
sync_commit_sha: pending-backfill
ac_source: .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001/acceptance.md
tier: M
ac_count: 11
changelog_entry_position: "[Unreleased] ### Changed, first entry"
frontmatter_status_transitions: in-progress -> implemented -> completed (spec.md, single sync commit; updated: stays 2026-10-19)
b12_self_test_a: pass (grep -c SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 CHANGELOG.md returned 0 before emission)
b12_self_test_b: pass by grep reading (11 distinct identifiers AC-PRC-001a..k, 0 reserved tokens, 0 ambiguous, live=11 matches the run-phase ac_pass_count 11); the canonical awk counter was not run because the worktree guard refused `awk -f`
b12_self_test_c: pass (the twelve page paths named in the entry are the twelve names in git diff --name-only 74ad197380a0 8f86c6bece, all present at HEAD)
```

What was synced: one CHANGELOG bullet (`CHANGELOG.md` `[Unreleased]` `### Changed`, first entry) and the 3-phase close (`spec.md` `status:` in-progress to completed). No README or docs-site change was made in the sync commit: the twelve page edits are the run-phase deliverable (commit 8f86c6bece), and this SPEC's docs obligation ends there. No `spec.md` HISTORY row was added, matching the closed siblings SPEC-FACTORY-RETIRE-GUARD-001 and SPEC-LANE-JOIN-RUN-SELECT-001, and no SPEC body text was touched.

Evidence paths: run-phase cells in §E.2 above; run-phase audit signal in §E.3; whole verbatim cell output in `.moai/reports/t1725/run-evidence.md` (gitignored, machine-local, not citable); plan audit verdict `.moai/reports/t1725/plan-audit-verdict-iter3.md`.

Gaps carried over from the run phase (not observed in the sync phase either):
- S-06d-1 to S-06d-5 (tree-build verb surface) were not run: heavy Go execution is frozen by the leader.
- Hugo was not built; the rendered output of the replacement lines was not observed.
- The ja, ko, and zh pinned prose (including the new labels) had no native-register review.
- The sync phase ran no new page cell; it relies on the run-phase GREEN cells at HEAD 8f86c6bece, and the sync commit changes no docs-site file.
- `sync_commit_sha` holds the `pending-backfill` placeholder because a commit cannot cite its own hash; the real SHA is owed to a following commit.

Residual risk: as recorded in §E.2 (ja and zh foundations row wording versus the linked `commands.md` headings, accepted under decision QE-4). The CI verdict on the integration branch is not part of this evidence. The plan-artifact hash input changes with this commit because `spec.md` carries the new status line; the iteration-3 audit hash `aa66a04ecd01b76ee264ead4c52fc42d152f8a33fe24725f12f39396041aca81` therefore no longer matches, which is expected for the close.
