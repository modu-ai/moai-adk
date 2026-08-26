---
id: SPEC-STAMP-REACHABILITY-001
title: "Codemaps stamp reachability: CI pre-merge orphan guard and explicit-commit stamping mode"
version: "0.2.0"
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".github/workflows/graph-freshness.yml, internal/cli, internal/mx"
lifecycle: spec-anchored
tags: "graph, codemaps, provenance, stamp, ci-guard, squash-merge, reachability"
era: V3R6
tier: M
related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002]
---

# SPEC-STAMP-REACHABILITY-001 — Codemaps stamp reachability: CI pre-merge orphan guard and explicit-commit stamping mode

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.2.0 | 2026-08-27 | Iteration-2 remediation clearing plan-audit iteration 1 (PASS-WITH-DEBT 0.81). D1: push-event anchor validity covered by new AC-SP-011 (GREEN cell + mutant-catch cell + static anatomy assertions — no `continue-on-error`, single-step shell-internal base-ref emptiness test pinned as THE conditioning mechanism); guard contract text made executable via acceptance.md §A.1 canonical reference script. D2: §D threshold-crossing recovery restated honestly — pre-merge threshold crossing accepts the red until a POST-MERGE main restamp (lead/operator act, outside delivery scope); merge-base restamp cannot restore freshness under its own trigger condition. D3: threshold direction corrected to ≥ (`count >= th.CodemapsChangedFiles`, measured check.go:214 this tree). D4: REQ-SR-004 gains verifying instrument AC-SP-012 (static inspection, forbidden-token enumeration). D5: AC-SP-003 fixture made paste-executable with local source-path fetch; absence premise MEASURED (rc=128) before baking the assertion. D6/D7 adopted (greppable prohibition tokens; canonical-script citation). All six guard-branch cells executed and quoted in acceptance.md baseline-attribution. | manager-spec |
| 0.1.0 | 2026-08-27 | Initial plan-phase authoring. Root cause adopted from triage-table.md §F5 (card t291 = follow-up F5 of card t279); countermeasure options adjudicated: option 1 (CI reachability guard) + option 2 (--commit flag) selected, option 3 alternatives rejected-and-recorded. Orphan premise re-measured on worktree base da791eb0a (`merge-base --is-ancestor 0d15864ae90b origin/main` rc=1); false-positive premise verified (`410da655f…` IS an ancestor of origin/main). | manager-spec |

## §A. Problem Statement

Under squash merges (this repository's default merge practice), a codemaps provenance stamp that names a feature-branch HEAD commit gets **orphaned at merge time**: the squashed landing creates a new main-side commit, the named stamp commit exists only on the feature branch, and after merge nothing in main's history reaches it. Consequences, both observed (triage-table.md §F5):

- main's `graph check` judged the codemaps layer **not-comparable → exit 2 → red**, and every subsequently opened PR inherited the red through the workflow's pull_request trigger (lane-4 #1662 measured until t279's emergency restamp).
- **Environment-dependent symptom shapes**: local checkouts holding the fetched-orphaned object judged codemaps fresh; CI checkouts lacking the object reported exit 2. One defect, different visible failures per environment — proof the stamp's branch-locality is the cause, not either consumer.

The incident was closed operationally by SPEC-V3R6-GRAPH-FRESHNESS-002's M0 emergency restamp (`52f7ba135`, against main-reachable `c9eed8ac6`), which also froze the recurrence obligation REQ-GFR-014 ("never restamp against branch-local HEAD"). But that obligation is **human discipline with zero mechanical enforcement**: nothing catches an orphan-bound stamp between the moment an operator writes one and the moment its next squash merge ships it. This SPEC builds the two countermeasures selected by lead adjudication (triage §F5 proposals 1-2):

1. **CI pre-merge reachability guard** — a graph-freshness workflow step verifying that the tracked provenance's `commit_sha` is an ancestor of the PR base before any freshness verdict is reported. The only enforcement point positioned *before* a squash merge can orphan the stamp.
2. **Explicit stamp-commit mode** — `moai graph stamp codemaps --commit <rev>` so the operator names a merge-surviving revision instead of the unconditional-HEAD recording (the ergonomics gap behind REQ-GFR-014 being easy to violate accidentally).

**Why a new SPEC**: the predecessor series is closed (`SPEC-V3R6-GRAPH-FRESHNESS-002 status: completed` with a frozen audit trail); this is independent follow-up candidate F5 promoted to a card (t291), not an increment of a frozen SPEC.

## §B. Scope

### §B.1 In Scope

Three milestones (execution order M1 → M2 → M3):

- **M1 — explicit stamp-commit mode**: extend `moai graph stamp codemaps` with `--commit <rev>`; the named revision resolves to a full sha and is recorded verbatim as `commit_sha`; unresolved revisions and the dirty-tree combination are rejected; the flagless default path is byte-for-byte behavior-preserved. Files: `internal/cli/graph_stamp.go` (+wiring), `internal/mx/provenance.go` (+tests).
- **M2 — CI pre-merge reachability guard**: add one step to `.github/workflows/graph-freshness.yml` that reads the tracked `.moai/project/codemaps/provenance.json`, verifies object presence in checkout history and ancestry versus the PR base ref, and fails the job with attributable annotations otherwise. Placement after checkout, before build/bootstrap, so an orphan-bound stamp fails fast without paying Go setup minutes.
- **M3 — regression characterization + operator documentation**: preserve-and-pin the existing freshness contract (digest compare, dirty anchoring, described_roots fidelity, 0/1/2 exit codes); document the explicit-commit mode with the merge-surviving recipe in the 4-locale `docs-site` cli-reference and the command Long text; gather plan→run handoff evidence.

### §B.2 Boundary decisions

- The guard is a **new workflow step, not new Go code** consumed by CI: judging reachability requires only `git merge-base --is-ancestor`, `git cat-file -e`, and `jq` reading a tracked JSON file — capabilities ubuntu-latest runners ship; routing it through the freshly built binary would couple the guard's correctness to the very stamp machinery it audits (a binary-lag bug could silence its own detector).
- On `push` events (main runs) the same step still verifies **object presence** of the named commit — an orphaned stamp that lands anyway fails main's own run with a message naming the defect class instead of the generic not-comparable exit 2. Ancestry-vs-base is structurally vacuous there and is checked only where it discriminates (pull_request events).
- Stamp-time `--commit` validation deliberately does **NOT** check main-reachability: a local machine cannot decide merge-survivability (origin/main moves; the squash hasn't happened yet). Resolvability is validated at stamp time; survivability is enforced where it is decidable — the CI guard at the merge gate. This division is deliberate, not an omission.

## §C. Requirements (GEARS)

#### REQ-SR-001 — Pre-merge ancestry verdict (When)

**When** a `pull_request` event reaches the graph-freshness workflow, the job shall read the tracked `.moai/project/codemaps/provenance.json` and verify — before reporting any freshness verdict — that the provenance's `commit_sha` is an ancestor of the PR base branch (`origin/<base_ref>` resolved from `github.base_ref`); **when** ancestry does not hold, the job shall fail the reachability step with an error annotation naming both the stamped sha and the base ref.

#### REQ-SR-002 — Missing-object failure (When)

**When** the provenance `commit_sha` names a commit whose object does not exist anywhere in the checkout's history, the reachability guard shall fail the job with an error annotation naming the missing sha — object-absence must not be skipped on the assumption that a later check will catch it, because the later check's exit 2 names the condition generically and arrives after paid setup minutes.

#### REQ-SR-003 — Anchorless skip-with-reason (While)

**While** the provenance block carries no commit anchor (`dirty: true`, a `content_fingerprint`, no `commit_sha`), the guard shall record its outcome as a printed skip-with-reason line on the job log and leave the verdict to the existing freshness layers — never a silent green and never a hard failure.

#### REQ-SR-004 — Judgeability inputs (Ubiquitous)

The reachability guard shall judge stamp reachability from git graph topology and tracked repository files only; it shall read no filesystem modification time and no wall-clock timestamp (consistent with the REQ-GF-002 mtime ban).

#### REQ-SR-005 — Explicit-commit stamping (Where)

**Where** the operator passes `--commit <rev>` to `moai graph stamp codemaps`, the command shall resolve `<rev>` to a full commit sha via git (`git rev-parse --verify <rev>^{commit}`) and record that resolved sha verbatim as the provenance's `commit_sha`.

#### REQ-SR-006 — Mixed-anchor rejection (When)

**When** `--commit` is passed while the described roots carry uncommitted changes versus HEAD, the command shall exit non-zero with a user-facing error free of absolute local paths and shall write no provenance file — a named-commit anchor and a dirty content-fingerprint anchor are mutually exclusive honesty claims (REQ-GFR-014 lineage; provenance.go's baseProvenance records exactly one).

#### REQ-SR-007 — Default-path preservation (Ubiquitous)

Absent `--commit`, `moai graph stamp codemaps` shall behave exactly as before this SPEC: a clean described-source tree records HEAD; described-source changes record `dirty: true` plus the aggregate content fingerprint; the atomic tmp-write-then-rename install and the output format are unchanged.

#### REQ-SR-008 — Pre-merge green impossibility (Unwanted)

A pull request carrying an orphan-bound codemaps stamp shall not be able to present a green graph-freshness check at merge time: the reachability failure (REQ-SR-001/002) surfaces as the job's verdict while squash merges remain repository practice.

#### REQ-SR-009 — Freshness-contract regression lock (While)

**While** the freshness checker evaluates the codemaps layer, its existing contract shall remain unchanged: the endpoint diff of described-source working-tree content against the stamped generation commit (reverted churn counts zero), dirty-fingerprint anchoring, `described_roots` fidelity, threshold resolution from gate.yaml, and the 0/1/2 exit-code discipline — including the not-comparable → system-error path (REQ-GF-004).

#### REQ-SR-010 — Operator guidance (Where)

**Where** operator-facing `graph stamp` documentation lives (docs-site cli-reference graph page, all 4 locales; the command's Long text), it shall document the explicit-commit mode together with the merge-surviving recipe `--commit "$(git merge-base HEAD origin/main)"` and shall state plainly that restamping against branch-local HEAD re-orphans under squash merges (the `0d15864ae90b` incident as the worked example).

## §D. Constraints

- **No new dependencies**; no dependency-version changes; no changes to gate.yaml defaults.
- **Provenance schema stability**: schema_version stays 1; no field added or repurposed. `tree_root` remains an absolute path by design (CR round-2 R2 refutation stands; informational-only for the tracked codemaps layer).
- **Stamp reachability (HARD, inherited)**: at no phase of THIS SPEC does the delivering PR's codemaps stamp get refreshed against a branch-local HEAD — the REQ-GFR-014 ban is retained verbatim and unconditionally. Expected case: described-source churn stays below the threshold, so the committed `410da655f…` stamp carries to merge unchanged. **Threshold-crossing recovery restated honestly**: if described-source churn reaches the threshold (≥40 changed files — `count >= th.CodemapsChangedFiles`, measured `check.go:214` on the iteration-2 tree) before merge, the pre-merge codemaps-freshness red is **accepted and carried**, not papered over: a merge-base restamp cannot restore freshness in its own trigger case, because the checker diffs working-tree content against the content AT THE NAMED COMMIT — recording the pre-churn ancestor re-counts the same churn as stale (that was v0.1.0's ineffective advice). Recovery is a **POST-MERGE restamp on main** naming a main-reachable commit — a lead/operator act OUTSIDE this SPEC's delivery scope; no milestone here repairs a crossed threshold.
- **Verification discipline**: affected-package tests only (`go test ./internal/cli/... ./internal/mx/... ./internal/graph/...`); NEVER a local full-suite run (CLAUDE.local.md §4/§6) — the delivering PR's CI is the full-suite judge. Workflow-step logic is exercised as local shell invocations against fixtures (acceptance.md) before landing; the workflow YAML itself is syntax-checked, not live-executed by plan/run phases.
- **Docs-site 4-locale parity**: any cli-reference change touches en/ja/ko/zh in the same change-set (oss-docs i18n HARD rule).
- **Template neutrality**: `.github/workflows/graph-freshness.yml` is dev-only repository infrastructure and is NOT template-mirrored (templates carry only `label-sync.yml` — measured). No `make build` is triggered by these edits; no forbidden-class content enters `internal/template/templates/**`.
- **Scope discipline**: only M-milestone files above; `docs-site` hugo build verified when touched; PR #1666 (t278) staleness questions are adjudicated elsewhere.

## §E. Out-of-Scope Declarations

Per-hypothesis boundaries below; each topic fenced with its reason.

### Out of Scope — Rejected countermeasure alternatives (record-only)

- Per-graph-affecting-merge post-merge main-restamp chore PRs: heavyweight, human-memory-dependent, repairs after exposure instead of blocking before merge — rejected (triage §F5 proposal 3).
- Switching repository convention to merge-commit merges: an org-wide process change far exceeding this defect's blast radius — rejected (triage §F5 proposal 3).

### Out of Scope — Provenance schema evolution

- No schema_version bump, field addition, or semantic change to `tree_root` (absolute-by-design per CR R2), `dirty`, `content_fingerprint`, or `described_roots`. The guard is a pure reader of the existing block.

### Out of Scope — Retrofit of historical orphaned states

- Backfilling, rewriting history around, or re-auditing previously orphaned stamps beyond what t279 already repaired. The guard prevents recurrence; archaeology of past incidents ends where triage-table.md §F5 ends.

### Out of Scope — Other graph layers and consumers

- mx-index and edges layers carry no commit anchors (file-inventory/fingerprint provenance) — out of the guard's domain by construction. Answer-attribution consumers (REQ-GF-008), MCP code-query provenance display, and gate integration are untouched.

### Out of Scope — Recurrence instance triage for open PR #1666

- Whether #1666's observed graph-freshness failure is orphan-class or staleness-class belongs to the t278 card; this SPEC fixes the mechanism for all future cases without inspecting or repairing that PR.
