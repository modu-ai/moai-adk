# progress.md — SPEC-STAMP-REACHABILITY-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M canonical set (spec.md, plan.md, acceptance.md) + research.md + progress.md, written in worktree `.claude/worktrees/t291` (branch `WT-stamp-orphan-guard`, base `da791eb0a`), 2026-08-27.
- Requirements: 10 REQ (GEARS, pattern-annotated: REQ-SR-001..010); 10 AC (Given-When-Then) each carrying a RED-now baseline measured at base; REQ↔AC traceability both directions complete (acceptance.md §D.1). Counts below Tier M ceilings.
- Premise re-measurement (this run, verbatim outcomes): `git merge-base --is-ancestor 0d15864ae90b origin/main` rc≠0 (`NOT_ANCESTOR` — orphan fixture live); provenance `.commit_sha` read as `410da655f39d…` via jq; `git merge-base --is-ancestor 410da655f39d1d3731abf2f247e8e5353a9d0de5 origin/main` rc=0 (`CURRENT_STAMP_ANCESTOR` — guard-green baseline live). SPEC ID regex self-check executed as Bash: `PASS`; catalog uniqueness probe count 0.
- Countermeasure adjudication per lead dispatch: option 1 (CI pre-merge reachability guard) and option 2 (explicit stamp-commit mode) BOTH selected as in-scope milestones M2/M1; option 3 alternatives rejected-and-recorded only. Validation split recorded as a boundary decision: resolvability at stamp time, survivability enforced by the CI gate where it is decidable.
- Scope fences: #1666 recurrence classification stays with t278; provenance schema evolution; historical-orphan retrofit; mx/edges layers — all fenced in spec.md §E with H3 Out-of-Scope headings.
- Constraint inheritance: predecessor REQ-GFR-014 (never restamp vs branch-local HEAD) binds this SPEC's own delivery; described-source churn below threshold means the committed stamp carries unchanged **provided its verification-time `.commit_sha` remains main-reachable** — at-or-above threshold accepts the pre-merge red until a POST-MERGE main restamp outside delivery scope (spec.md §D).
- Delta round (2026-08-27, post plan-audit iter-2 PASS-WITH-DEBT 0.94): v0.2.1 wording remediation N2/N3/N4/N5 all applied; N3 authoritative numeral re-measured first-hand (`grep -n` single hit → check.go**:213** at HEAD 016dc0b8c; iter-2's :214 was an authoring mis-count); HISTORY row appended; version field bumped 0.2.0 → 0.2.1 for traceability across frontmatter/HISTORY/progress.
- Stamp-health moving-target observation, recorded as dated fact (audit-cycle context): on 2026-08-27 after upstream #1668 landed through HEAD 0364217e9+ lineage, the tracked `.moai/project/codemaps/provenance.json` names `commit_sha` = `a995e58fa69b40f3d76a76c0a5ca56d0594fb528` — locally present (`git cat-file -t` → commit) and NOT an ancestor of origin/main (`git merge-base --is-ancestor … origin/main` rc=1): a live SECOND occurrence of the F5 orphan class on main. Instance repair routes to lead triage OUTSIDE this card's scope; no requirement text changed on its account. Live-probe corroboration with the §A.1 canonical script returned rc=1 with the ancestry annotation naming that sha — the planned M2 guard would surface exactly this state red pre-merge.
- Constraint inheritance note superseded in part by the observation above: iter-2's "committed stamp expected to carry unchanged" premise was TRUE of its measurement moment (sha `410da655f39d…`, ancestor-verified rc=0, 2026-08-27 pre-#1668) and is retained as a date-pinned historical observation only (acceptance.md baseline-attribution, AC-SP-002(ii)).
- Evidence paths cited across artifacts resolve within the worktree tree (triage-table.md tracked on base; code anchors at graph_stamp.go/provenance.go/check.go/graph_check.go line-cited in plan.md §H).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier=M · scope ≈6-8 files (internal/cli/graph_stamp.go+tests, internal/mx/provenance.go+tests, .github/workflows/graph-freshness.yml, command Long text, docs-site cli-reference ×4 locales) · domains=3 (Go CLI / CI YAML / docs markdown) · language mix Go+YAML+MD · concurrency benefit LOW (coding-heavy, inter-milestone dependency M1→M2→M3 sequential in plan.md) · Agent Teams prereqs not requested.
- Mode evaluation: direct — not selected (multi-file semantic change) · fanout — not selected (Anthropic coding-task parallelism caveat: implementation is coding-heavy with milestone ordering; parallel spawns would race shared files) · sweep — not selected (<30 files, non-mechanical multi-rule edits, Workflow overhead unwarranted) · **serial — selected**.
- Decision: `serial` (one `manager-develop` sub-agent, milestones in plan.md order, TDD cycles per milestone).
- Justification: coding-heavy work plus strict M1→M2→M3 dependency ordering makes sequential single-agent execution both the safe default and the cache-economical one; a single §E evidence stream per milestone lands contiguously in §E.2. Recorded after Implementation Kickoff Approval (operator-approved 2026-08-27, relayed by lead session) and pre-spawn sync checks (divergence absorbed through merge `e514e695f`; same-SPEC active sessions `[]`).
