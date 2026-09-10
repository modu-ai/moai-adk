# SPEC-SYNC-STRATEGY-KEY-001 — Independent Sync Audit (card t303)

Auditor: sync-auditor (cold start, no prior involvement).
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t303`, branch `WT-strategy-key-sync`, tip `cd21e9ccb`, base develop `812ee01fc`.
Tier M, PASS threshold 80.

**Verdict: PASS-WITH-DEBT — weighted harmonic mean 90.21. Must-pass firewall CLEARED (Functionality 96 ≥ 80, Security 93 ≥ 80). Zero blocking findings; three non-blocking evidence-honesty defects enumerated below.**

---

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 96/100 | PASS (must-pass) | 12/12 AC re-measured GREEN independently; 9/9 MUST-PASS green — see §1 |
| Security (25%) | 93/100 | PASS (must-pass) | Change strictly tightens delivery: explicit unmatched-value stop (no PR, no push), no silent default (`grep 'Default strategy\|Otherwise,'` → 0 hits), "never force-push" in the WT-* route, precondition stop. No secrets, no injection surface. F4 is the only deduction. |
| Craft (20%) | 78/100 | PASS | F1 + F2 + F3 land here: two evidence rows in `progress.md` do not reproduce as written |
| Consistency (15%) | 90/100 | PASS | Mirror parity holds (only the two documented local-only diffs); neutrality guards green; CHANGELOG placement correct. F4/F5 minor. |

Weighted harmonic mean = `1 / (0.40/96 + 0.25/93 + 0.20/78 + 0.15/90)` = **90.21**.

---

## §1 Object 1 — the implementation (12 ACs, re-measured on `cd21e9ccb`)

Nothing below is taken from `progress.md` §E.2. Every cell is a command this auditor ran on this tree.

| AC | Sev | Command | Verbatim output | Verdict |
|---|---|---|---|---|
| 001 | MUST | `grep -c 'git_strategy.{mode}.workflow' T/…/delivery.md` | `6` | PASS |
| 001 | MUST | `grep -n 'git_strategy' T/…/delivery.md` | canonical reads at L19/20/29/31/229/232/234/387/441; the only legacy mention is L33, inside the fallback block | PASS |
| 002 | MUST | `grep -rn 'spec_git_workflow' internal/template/templates/` | 1 hit — `…/sync/delivery.md:33`, line begins `**Legacy key fallback (deprecated, removed in v3.3.0).**` | PASS (refinement legitimate — §3) |
| 003 | SHOULD | delivery.md L33-44 read directly | fallback block present; names canonical key, removal version v3.3.0, deprecation warning, 5-row mapping table, canonical-wins rule | PASS |
| 004 | MUST | delivery.md L29 read directly | `…stop the delivery step and report the offending value together with the canonical domain {github-flow, git-flow}. Do not create a pull request and do not push. A missing workflow subkey under the active mode is an unmatched value, not a default — there is no fallback strategy.` | PASS |
| 005 | MUST | `grep -c 'WT-' D` / `awk` WT-block `\| grep -c 'gh pr'` | `6` / `0` | PASS |
| 006 | MUST | `awk '/^##### Strategy:/{s=$0} /matches no defined route/{c[s]++} END{...}' D` | `##### Strategy: github-flow 1` + `##### Strategy: git-flow 1` (exactly 2) | PASS |
| 006 | MUST | `grep -n 'Default strategy\|Otherwise,' D` | (no output, rc=1) | PASS |
| 007 | SHOULD | `grep -n 'delivery.md\|merge --no-ff' .claude/rules/local/gitflow-lane-protocol.md` | 2 hits, both citations (L29, L119); zero `merge --no-ff` — the dev rule cites, does not restate | PASS |
| 008 | MUST | `grep -rn 'gitflow-lane-protocol' T/` | (no output, rc=1) | PASS |
| 008 | MUST | `grep -n 'workflow:' T/.moai/config/sections/git-strategy.yaml.tmpl` | `13: workflow: github-flow` / `45:` / `81:` — no private value leaked | PASS |
| 008 | MUST | `git diff d29b8942e -- internal/template/templates/ \| grep '^+' \| grep -oE 'SPEC-[A-Z0-9-]+-[0-9]{3}' \| wc -l` | `0` | PASS |
| 009 | MUST | `grep -c 'github.spec_git_workflow' internal/config/testdata/shipped_key_inventory.yaml` | `0` | PASS |
| 009 | MUST | `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1 -v` | `--- PASS` + 4 named subtests ran; `non-vacuity: 911 shipped keys, 974 inventory entries, 344 struct fields`; `ok … 3.626s` | PASS (non-vacuous: 4 subtests matched) |
| 010 | MUST | `grep -n 'auto_branch\|spec_git_workflow' T/…/tab_schema.json` | 3 hits (L964/966/982), all `git_strategy.{mode}.automation.auto_branch`; zero legacy-key hits | PASS |
| 011 | SHOULD | `sed -n '25p'` on template and local `doc-execution.md` | both: `- Read project configuration: git_strategy.mode, git_strategy.{mode}.workflow, conversation_language` | PASS |
| 012.1 | MUST | `grep -rn 'spec_git_workflow' .claude/skills/ .moai/config/sections/system.yaml` | 1 hit — local `delivery.md:33`, the same D1 sentinel | PASS (see F3) |
| 012.2 | MUST | `diff T/…/tab_schema.json L/…/tab_schema.json` | (empty, rc=0) | PASS |
| 012.3 | MUST | `grep -c 'SPEC-SYNC-PARALLEL-DOCS-001' L/…/doc-execution.md` | `3` — A5 block intact | PASS |
| 012.4 | MUST | `grep -n 'workflow:\|mode:' .moai/config/sections/git-strategy.yaml` (primary checkout) | `2: mode: manual` / `8: workflow: git-flow` / 35 / 62 `github-flow` — in-domain | PASS |

Mirror-parity spot checks (independent of §E.2):

- `diff T/…/delivery.md L/…/delivery.md` → 2 lines only, both in the footer changelog block. Documented local-only drift.
- `diff T/…/doc-execution.md L/…/doc-execution.md` → the SPEC-SYNC-PARALLEL-DOCS-001 A5 attribution block only. Documented local-only content, preserved.
- `go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestRuleTemplateMirror' -count=1 -v` → `TestRuleTemplateMirrorDrift`, `TestTemplateNeutralityAudit` (6 subtests), `TestTemplateNeutralityAuditC8Preserve` all `--- PASS`; `ok … 0.651s`. Non-vacuous.

**Object 1 conclusion: the implementation genuinely satisfies all 12 acceptance criteria. No functional defect found.**

---

## §2 Object 2 — the sync closure's honesty

### 2.1 Cross-card claims — both VERIFIED

§E.4 asserts OBS-1 was resolved by card t316 (`6310dbf28`) and B5 by card t308 (`da301bbe1`). This is the shape that normally goes unchallenged. Both were read.

- `git log -1 6310dbf28` → merge, `merge(WT-tabschema-autobranch): integrate card t316 — drop dead-path auto_branch questions from the interview schema`.
  `git diff 6310dbf28^1 6310dbf28 -- T/…/tab_schema.json | grep -E '^[-+].*auto_branch'` → six `-` lines removing exactly `git_strategy.personal.auto_branch` ×3 and `git_strategy.team.auto_branch` ×3; **zero `+` lines**, so t303's own rebind was not touched. `merge-base --is-ancestor 6310dbf28 HEAD` → `ANCESTOR-YES`. Claim accurate.
- `git show da301bbe1 -- CLAUDE.local.md` → the diff replaces `workflow: gitflow` with `workflow: git-flow` in the §2.3 check command and in the manual-restore instruction. `merge-base --is-ancestor da301bbe1 HEAD` → `ANCESTOR-YES`. Claim accurate.

Both are correctly credited to the other card rather than claimed as this SPEC's own work. That is the right disposition.

### 2.2 Frontmatter scope — FACT, not convenience

`grep -rn '^status:' .moai/specs/SPEC-SYNC-STRATEGY-KEY-001/` → single hit, `spec.md:5:status: implemented`.
`grep -c '^---'` → `plan.md: 0`, `progress.md: 0`, `acceptance.md: 2` (both body horizontal rules — the file opens with `# SPEC-…`), `spec.md: 7` (frontmatter + body rules).
Only `spec.md` carries a frontmatter block, so only `spec.md`'s `status` could transition. The §E.4 note is accurate and the disclosure is appropriately explicit.

### 2.3 Status transition — WARRANTED

`in-progress → implemented`, deliberately not `completed`, with the stated reason that the independent audit had not yet run and the `completed` call belongs to the dispatching lead. This is the conservative and correct choice. All 12 ACs are green, so `implemented` is earned rather than asserted.

### 2.4 CHANGELOG entry — ACCURATE

Read verbatim from `git show e9f288473 -- CHANGELOG.md`. Every checkable statement reproduces: the 1-hit refined grep with the baseline of 10 disclosed inline; `{github-flow, git-flow}`; the stop-on-unmatched behaviour; the 3-hit `automation.auto_branch` rebind with byte-identical mirror; "12 acceptance criteria (AC-SYK-001..012): 12 PASS, 0 FAIL"; run commits `c374d9605` → `63b4628a6`; merge `7ed6edb3e` (`git log -1 --format='%H %p %s' 7ed6edb3e` → parents `daf206903 8f519407b`, so `8f519407b` is the branch tip merged in — correct). Both open debts are named. Placement matches the stated position (first bullet under `[Unreleased] → Changed`, before SPEC-TABSCHEMA-AUTOBRANCH-001).

### 2.5 The `docs_impact: none` claim — VERIFIED

`grep -rln 'spec_git_workflow' docs-site` → rc=1, no output.

### 2.6 Open debt disposition — HONEST

OBS-2 and OBS-3 are left OPEN and reported, and OBS-3 carries the explicit admission "Run locally instead, exit 0 — not independently re-verified on CI by this sync commit." That is the correct treatment of an unobserved check.

---

## §3 Rulings on the two flagged judgment calls

### AC-SYK-002's "documented refinement" — LEGITIMATE SCOPING, not a weakened assertion

The criterion text pre-authorizes it verbatim: *"if that blocks this grep, the sentinel `spec_git_workflow` must appear only in the fallback and the AC grep is refined to exclude the deprecation block — state the refinement in the evidence, do not weaken the removal claim."*

Three conditions were checked, all hold:
1. The surviving hit is exactly one, and it is inside the fallback block (`delivery.md:33`, the line that opens the D1 block).
2. The refinement is stated in the evidence — in `progress.md` §E.2, in the commit message, and in the CHANGELOG entry, each time with the baseline of 10 disclosed.
3. The removal claim is not weakened: 10 → 1, and the surviving 1 is required to exist by AC-SYK-003 (a SHOULD that mandates the deprecation fallback). A literal 0 would make AC-002 and AC-003 mutually unsatisfiable.

Ruling: **legitimate**. Not a finding.

### AC-SYK-004 / 005 / 006 as prose — a STATED STOP, not a missing default

The distinction the brief asks for is the right one, and the text passes it on the affirmative side.

- **AC-004 (L29)** does not merely omit a default. It (a) commands the stop, (b) specifies what to report, (c) forbids both PR and push, (d) closes the missing-subkey hole by name — *"A missing `workflow` subkey under the active mode is an unmatched value, not a default"* — and (e) asserts the non-existence of a fallback. The negative probe confirms the old `Default strategy (if not configured): github_flow` line is gone.
- **AC-006** has an explicit terminal clause in *both* strategy blocks, mechanically attributed one-per-block by the awk probe. The git-flow one goes further: *"Do not improvise a route and do not fall back to another strategy's behavior."* That forecloses the two ways prose usually leaks.
- **AC-005** is 8 numbered steps with first-match order declared normative, zero `gh pr` tokens in the block, an explicit "never force-push", and a stop-and-report precondition when the integration worktree is absent.

Ruling: the text **forbids** a silent default; it does not merely fail to mention one. The anti-mutant pair holds — removing the WT route reddens 005, keeping fall-through reddens 006.

### AC-count disambiguation — CONCLUSION RIGHT, EXPLANATION WRONG

See F1.

---

## §4 Findings

| ID | Severity | Blocking | Location | Finding | Required fix |
|---|---|---|---|---|---|
| F1 | Medium | No | `progress.md` §E.4, `ac_count_used_for_b12_self_test` | The stated command does not reproduce the stated output. §E.4 says *"a raw `AC-([A-Z0-9]+-)*[0-9]+` regex over the whole file returns 24 because L111 also carries a shorthand … alias."* Measured: `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| wc -l` → **37** (`grep -c` → 20 lines). The **24** is the count of full `AC-SYK-NNN` tokens (`grep -oE 'AC-SYK-[0-9]{3}' \| wc -l` → 24), which the L111 shorthand table contributes **zero** to — L111 carries `AC-001`-style aliases, which that narrower regex cannot match. The real cause of 24-vs-12 is ACs cross-referencing each other inside their own bodies (AC-005 ×4, AC-002 ×3, AC-003 ×3, …). L111's 13 shorthand tokens are exactly the 37−24 remainder. The **conclusion is correct and independently verified**: `grep -c '^### AC-SYK-[0-9]{3}' acceptance.md` → **12**. VCI §2 (baseline-integrity attribution). | Rewrite the `ac_count_used_for_b12_self_test` note to cite the command that actually returns 24 (`grep -oE 'AC-SYK-[0-9]{3}' … \| wc -l`) and attribute the surplus to intra-AC cross-references, keeping the L111 shorthand table as a separate, secondary observation about the wider regex. |
| F2 | Low | No | `progress.md` §E.2, invariant row "Adjacent template guards" | Vacuous-green. The cited selector `-run 'TestInternalContentLeak\|TestRuleTemplateMirror\|TestCommandsAudit'` names three tests; `grep -rn 'func TestInternalContentLeak\|func TestCommandsAudit' internal/` returns **no output** — neither exists anywhere under `internal/`. Only `TestRuleTemplateMirrorDrift` matched. The `ok … 0.249s` is real but establishes one guard, not three. Carried into the sync closure unreviewed. | Correct the row to name the test that exists (`TestRuleTemplateMirrorDrift`), or add the guards that were intended. Never cite a `-run` selector without the matched-case count. |
| F3 | Low | No | `acceptance.md` AC-SYK-012 sub-criterion 1 vs `progress.md` §E.2 | AC-012.1 literally requires the local dead-key grep `= 0`; actual is **1** (the same D1 sentinel). §E.2 records `(1) 0` with the prefix "local dead-key grep refined". Unlike AC-002, AC-012.1's own text carries **no** refinement clause, so a refined figure is reported against an unrefined criterion. Substantively identical justification to AC-002 and materially harmless — but the disclosure is thinner than the one AC-002 received. | Either state the refinement inline on the AC-012 row the way the AC-002 row does, or amend AC-012.1 to inherit AC-002's fallback-exclusion clause. |
| F4 | Low | No | `T/.claude/skills/moai/workflows/sync/delivery.md`, git-flow `WT-*` route step 6 | The step is labelled *"Push the integration branch"* but hardcodes `git push origin develop`. Under a release-branch integration batch (`release/vX.Y.Z`, the model this repo's own lane doctrine uses — `AGENTS.md` §3, `kanban-dispatch.md`) the literal command targets the wrong branch while the label reads correct. Not an AC violation: AC-SYK-005 specified only "push the integration branch". | Optional. Parameterize as `git push origin <integration-branch>` with `develop` given as the git-flow default, so label and command agree. |
| F5 | Info | No | `CHANGELOG.md`, the new `[Unreleased] → Changed` first bullet | The entry omits the trailing `🗿 MoAI` marker that both adjacent entries (SPEC-TABSCHEMA-AUTOBRANCH-001, SPEC-GLM-FLASH-DEFAULT-001) carry. Cosmetic. | Optional. Append the marker for section consistency. |

**Blocking findings: 0.** F1-F3 are documentation-honesty defects in the closure artifact; none of them changes an AC verdict, and none is grounds to withhold the merge. They are debt to be corrected, which is why the verdict carries the `-WITH-DEBT` suffix rather than a bare PASS.

Stated plainly, because the brief asked for it: **the implementation is sound.** Twelve criteria, twelve independent green re-measurements, both card regressions genuinely closed, both cross-card resolution claims true and correctly attributed, mirror parity intact with the two documented local-only differences preserved. No findings were manufactured to pad this list, and F1 was not softened.

---

## §5 Evidence-bearing summary

**Claim.** SPEC-SYNC-STRATEGY-KEY-001's implementation satisfies all 12 acceptance criteria on tree `cd21e9ccb`, and its sync closure is substantively honest with three correctable evidence defects.

**Evidence.** The 21 command/output rows in §1, the four verification blocks in §2, and the five findings in §4. Every output is verbatim from a command this auditor ran in this worktree in this run. Test selectors were checked for non-vacuity: `TestShippedConfigKeysHaveReaders` matched 4 subtests over 911 shipped keys; the template-guard selector matched 3 test functions with 6 named subtests.

**Baseline-attribution.** All measurements taken on `git rev-parse --show-toplevel` = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t303`, `HEAD` = `cd21e9ccb`, `git status --short` empty at audit start. The AC-008 diff probe is attributed to the acceptance file's own stated baseline `d29b8942e` (not §E.2's `0931789b6`); it returns 0 on both framings. Cross-card claims attributed to `6310dbf28` and `da301bbe1`, both confirmed ancestors of `HEAD` by `merge-base --is-ancestor`.

**Gaps — explicitly NOT observed.**
1. CI was not consulted. No `gh pr checks` or CodeRabbit read was performed; the branch carries no PR under this repo's git-flow card model. OBS-3 (whether the template-neutrality guard fires on CI for a card branch) therefore remains unverified by this audit as well as by the sync commit.
2. `make build` was not re-run and `bin/moai` was not scanned. The embedded-asset parity claim in §E.2 is carried forward unverified by this auditor; the two source copies were verified byte-identical, which is a weaker statement.
3. No behavioural execution of the delivery workflow. AC-003/004/005/006 are prose instructions in a skill file; they were judged as text, which is the only judgement available. Whether an agent reading them actually stops was not and cannot be measured here.
4. The full Go suite was not run — only the two targeted packages the change touches, per the repo's local-suite prohibition.
5. `spec.md` and `plan.md` bodies were read only in the regions bearing on the ACs; no full-text review of either.

**Residual risk.**
- Prose-encoded control flow (the unmatched-value stop, the WT-* route, the two terminal clauses) has no mechanical enforcement. A future edit can delete any of them and every grep-shaped AC here would need re-running to notice. The `awk` per-block probe in AC-006 is the closest thing to a regression guard and it lives only in the acceptance file, not in CI.
- F4's hardcoded `develop` is the most likely operational surprise: a lane following the route literally during a release-branch batch pushes to the wrong branch.
- The D1 fallback keeps the retired key alive until v3.3.0. AC-SYK-002's refined form will silently keep passing after the fallback should have been removed unless the §D.5 forward-looking check is actually executed at that release.
- OBS-2 (interview question overlap) remains open and will surface to a user running `/moai project` in personal or team mode as a duplicated-feeling question, not as an error.

---

## §6 Verdict

**PASS-WITH-DEBT — 90.21 / threshold 80.**

Must-pass firewall, stated explicitly: Functionality 96 ≥ 80 **PASS**; Security 93 ≥ 80 **PASS**. Both must-pass dimensions cleared independently, so the firewall does not force a FAIL.

Debt carried into close, none blocking: **F1** (Medium — §E.4's AC-count command does not reproduce; conclusion correct, explanation wrong), **F2** (Low — a `-run` selector naming two tests that do not exist), **F3** (Low — a refined figure reported against an unrefined criterion). **F4** and **F5** are optional.

The recommendation on the `completed` transition is the dispatching lead's, per §E.4. From this audit's side there is no functional obstacle to it; correcting F1 and F2 in the closure artifact first would be the cleaner order.

---

Auditor: sync-auditor · Audit tree `cd21e9ccb` · This report is uncommitted by design.
