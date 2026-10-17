# SPEC-LANE-JOIN-RUN-SELECT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-17

Basis for the explicit signal: `moai spec lint` exited 0 on the SPEC directory in both the plain and the `--strict` form (record below). This is the author's own lint, not an audit verdict; no plan audit has run.

- Authoring: spec.md, plan.md, acceptance.md, decision-index.md written by manager-spec (Tier M, card t1853, Class C; the A-part of the split of card t1848). Tree `fb749d372e61a6b10c53091362c76b4412827427`, branch `WT-10-17-class-9`. Not yet audited.
- Counts: 7 requirements (Tier M ceiling 16), 8 acceptance criteria (ceiling 16), 3 milestones (M1-M3).
- SPEC ID self-check, executed as Bash: `ID="SPEC-LANE-JOIN-RUN-SELECT-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`; `ls .moai/specs | grep -i "LANE-JOIN-RUN\|RUN-SELECT"` printed nothing (grep exit 1), so the ID is unique at this tree.
- RED-now evidence: `acceptance.md` ledger E1-E3 (empty sweeps, exit 0, swept-count FAIL; 3 + 2 + 9 named tests), E4 (three existing guards GREEN). No behavioral RED was observed.
- Decision rows: Q1 `DEFAULT-APPLIED` (implementation-level); none open.
- Lint (command, judging build, tree): a binary built from this tree (`go build -o <scratchpad>/bin/moai ./cmd/moai` run in the card tree at HEAD `fb749d372e61a6b10c53091362c76b4412827427`, exit 0) and invoked by path: `moai spec lint <SPEC dir>` printed `✓ No findings — all SPEC documents are valid`, exit 0; `moai spec lint --strict <SPEC dir>` printed the same line, exit 0. That binary carries no commit stamp (`moai version` prints `v3.1.3 none built unknown` because a bare `go build` omits the Makefile's `LDFLAGS`), so its provenance is the build command at the HEAD above, not a self-reported SHA. The installed `moai` is `v3.2.0-rc.36-t1822c` (commit `ee981a932`); `git merge-base --is-ancestor ee981a932 HEAD` exited 0, so it lags this tree and was not used for the cited result (verification-claim-integrity §2.2).
- Hygiene sweep over the five files: no `origin/*` ref, no duration-style time estimate (`grep -nEi` for the patterns printed nothing). Line numbers cited in the SPEC are coordinates at the tree above and are stated as such in `spec.md` §C and `plan.md` §A.
- Gaps carried forward (full list in `spec.md` §H): no behavioral RED; the E4 guards ran in a shell carrying lane environment variables (TestMain clears ambient factory state); the read-path write-free claim is scoped to the `runs` and `events` tables; the Korean strings are authored but not read by a native reviewer beyond the author; docs-site pages not read in full; no plan-auditor or cross-model audit has run.
- Residual judgment calls: (1) the card says the launchers "pick a run implicitly" while this tree already refuses with two or more survivors (`spec.md` §B states the observed behavior); (2) Q1 default is read-only re-presentation (no reconciler); (3) the one-row table never offers cleanup (`CanCleanup` needs two or more rows); (4) the tree-built lint binary has no commit stamp; (5) AC-LJR-007 is a single test name carrying five scenarios (the N2 scenario set), kept as one criterion so its sweep count is one.
- Inherited history (carried, not reset): the predecessor SPEC-FACTORY-RUN-SELECT-CORE-001 (card t1848) plan audits FAIL 0.81 -> 0.89 -> 0.90; D1-D5 and N1 closed and carried in `spec.md` §F; N2 resolved by REQ-LJR-005 / REQ-LJR-006 and pinned by AC-LJR-007 (plan §C.4). The three reports were read in the sibling tree `.moai/worktrees/t1848/.moai/reports/t1848/`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
