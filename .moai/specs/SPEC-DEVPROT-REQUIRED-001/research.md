# Research: develop 브랜치 보호 상태검사 필수 승격 설계 (card t324)

> Investigator: orchestrator lane session (card t324, worktree `.claude/worktrees/t324`,
> branch `WT-devprot-required`, base `origin/develop` @ `fa8ff89ba`, measured 2026-09-02).
> All repo-side numbers below are from this run against this tree; external semantics are
> docs-verified with URLs. This file is research evidence, not the SPEC.

## 1. Current state (measured this run)

### 1.1 develop — unprotected

```
gh api repos/modu-ai/moai-adk/branches/develop/protection
→ {"message":"Branch not protected",...,"status":"404"}
```

### 1.2 main — protected (live read, 2026-09-02)

`gh api repos/modu-ai/moai-adk/branches/main/protection`:

| Setting | Live value |
|---|---|
| `required_status_checks.strict` | `false` (loose) |
| `required_status_checks.contexts` | `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`, `Release PR Multi-OS Gate` |
| `required_pull_request_reviews` | enabled, `required_approving_review_count: 0`, `dismiss_stale_reviews: true` |
| `enforce_admins.enabled` | **`true`** |
| `required_linear_history` / `allow_force_pushes` / `allow_deletions` | false / false / false |
| `required_conversation_resolution` | `false` |

**Doctrine drift note (observed, same API read):** `.moai/docs/git-local-workflow-doctrine.md`
§23.2 (2026-07-20 snapshot) records `strict: true`, `required_conversation_resolution: true`,
4 contexts. Live today: `strict: false`, conversation-resolution `false`, 5 contexts
(`Release PR Multi-OS Gate` added later, t406 era). The doctrine table is a stale snapshot —
the SPEC should not treat §23.2 numbers as current.

## 2. Workflow inventory — what fires on push to develop

Measured from `.github/workflows/*` (this tree, base `fa8ff89ba`):

| Workflow | push → develop | Contexts it can report on develop push | Required-check eligible? |
|---|---|---|---|
| `ci.yml` | **yes** (`branches: [main, develop]`, ci.yml:17-18) | `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Race Test`, `Constitution Check`, `Integration Tests (…)` | **yes — see 2.1** |
| `codeql.yml` | **yes** (codeql.yml:4-5) | `Analyze (Go) (go)` — always reported on push (analyze runs on push events regardless of go_code — see 2.2 refutation note) | **yes** |
| `lsel-leak-guard.yaml` | **yes** (lsel-leak-guard.yaml:10-11) | (own job) | eligible (not on main's list) |
| `test-install.yml` | paths-gated (`install.sh` only, test-install.yml:5-7) | — on most pushes workflow never fires → contexts stay **Pending** | **no** (docs §Pending hazard) |
| `release-pr-multi-os.yml` | no (`pull_request: [main]` only) | `Release PR Multi-OS Gate` structurally absent on develop SHAs | **no — permanent block if required** |
| `graph-freshness.yml` | no (`pull_request: [main, develop]`, graph-freshness.yml:8-9) | — | no (PR-only) |
| `docs-i18n-check` / `spec-lint` / `template-neutrality-check` / `review-quality-gate` / `auto-merge` / community / label / release-drafter / spec-status-auto-sync / claude | no (PR / check_run / issue triggers) | — | no |

### 2.1 ci.yml — context reliability on develop push (the good case)

ci.yml is a `detect` (dorny/paths-filter) + conditional-jobs workflow; it **always fires** on
push to develop, and the load-bearing trick is the paired skip-marker jobs:

- `test` → `Test (${{ matrix.os }})`, `if: needs.detect.outputs.go_code == 'true'` (ci.yml:114-118)
- `test-skip-marker` → **same name** `Test (${{ matrix.os }})`,
  `if: needs.detect.outputs.go_code != 'true'` (ci.yml:316-332) — reports `Test (ubuntu-latest)`
  = success on docs-only/chore-only pushes. GitHub docs: a job skipped by a conditional
  "will report its status as 'Success'".

→ **`Test (ubuntu-latest)` is reported on every ci.yml run** (real run OR skip-marker).
- `lint` (ci.yml:422) — no conditional → always reported.
- `build` (ci.yml:463) — no conditional; matrix includes `linux/amd64` (ci.yml:470-472) →
  `Build (linux/amd64)` always reported.
- `test-race` `Race Test`: conditional (`go_code == 'true' && !startsWith(github.head_ref, 'release/')`,
  ci.yml:257) with **no skip-marker** → not reported on docs-only push. Not required-check-eligible
  without a marker companion change.
- `constitution-check` (ci.yml:533) — no conditional shown → always reported.
- Concurrency: `concurrency: … cancel-in-progress: true` (ci.yml:35-37) — group key needs a look
  during design (ref-scoped groups would not cancel cross-ref duplicate runs).

### 2.2 codeql.yml — the PR-only skip-marker gap (the bad case)

> **[REFUTED 2026-09-02 — do not rely on this section's premise.]** The initial read
> below inferred that `Analyze (Go) (go)` goes unreported on docs-only direct pushes.
> That inference was wrong: the `analyze` job's actual condition (codeql.yml:59-63,
> re-read in full during SPEC authoring) is
> `go_code == 'true' || github.event_name == 'push' || github.event_name == 'schedule'` —
> it runs on EVERY push regardless of go_code. Empirical confirmation: develop HEAD
> `fa8ff89ba` is a docs-only merge (first-parent diff all-`.md`) and `Analyze (Go) (go)`
> reported success on that SHA (read-only `gh api .../check-runs`, this run). The
> PR-only skip-marker is not a gap — push events never needed it. `Analyze (Go) (go)`
> IS required-check eligible on develop without companion changes. SPEC
> SPEC-DEVPROT-REQUIRED-001 carries the corrected fact (REQ-DPR-004).

- `analyze` → `Analyze (Go)` + language matrix (renders `Analyze (Go) (go)`),
  `needs: detect` (codeql.yml:52-55), gated on go_code.
- `analyze-skip-marker`: `if: needs.detect.outputs.go_code != 'true' && github.event_name == 'pull_request'`
  (codeql.yml:107-124) — **the marker is PR-only.**

→ On a **direct push to develop** whose merge is docs-only/chore-only (`go_code != 'true'`),
neither job runs → `Analyze (Go) (go)` is **never reported** → stays Pending → any direct push
of such a merge to a branch requiring that context is rejected (GH006 "expected").
**Requiring this context on develop needs a companion change widening the marker to push events.**

## 3. GitHub semantics (docs-verified, 2026-09-02)

Sources (fetched this run):
- https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches
- https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/troubleshooting-required-status-checks

Verified statements load-bearing for the design:

1. **Direct push is gated**: "You won't be able to push local changes to a protected branch
   until all required status checks pass" — error form GH006: `Required status check "ci-build"
   is failing`. (Missing context variant surfaces as "is expected" in practice.)
2. **Checks are SHA-scoped**: "Required check needs to succeed against the latest commit SHA …
   Checks that were triggered using a previous commit SHA will not be used." Success statuses =
   `success`, `skipped`, `neutral`.
3. **The pre-verified-SHA pathway**: "After all required status checks pass, any commits must
   either be pushed to another branch and then merged **or pushed directly to the protected
   branch**." Statuses attach to the SHA repo-globally — a SHA verified green on any branch
   satisfies the gate when pushed elsewhere. **This is the mechanical basis of the
   merge-commit pre-verification option (B1).**
4. **PR local-merge allowance**: "Pull requests that are up-to-date and pass required status
   checks can be merged locally and pushed to the protected branch. This can be done without
   status checks running on the merge commit itself." — scoped to *pull requests that exist*;
   the no-PR lane model cannot rely on it without opening PRs (see B4, rejected).
5. **Workflow-level skip → Pending → blocked**: "If a workflow is skipped due to path
   filtering, branch filtering or a commit message, then checks associated with that workflow
   will remain in a 'Pending' state" — job-level conditional skip reports "Success".
6. **Admin bypass default**: "By default, the restrictions of a branch protection rule don't
   apply to people with admin permissions … You can optionally apply the restrictions to
   administrators" (= enforce_admins / "Do not allow bypassing").
7. **Configurability precondition**: to select a context as required it must have completed
   successfully in the repo within the past 7 days.

## 4. The core conflict: merge commit has no statuses

The integration window does `git merge --no-ff <card-branch>` → **new merge-commit SHA** →
`git push origin develop`. That SHA has never run CI; no required context exists on it at push
time. Card-branch pushes do not trigger ci.yml (branches-filtered), so card HEADs carry no
statuses either. Consequences under each enforcement level:

- enforce_admins=false + admin push: bypassed at push; CI runs **post-push** (push trigger is
  already active) — verification exists but is not a gate.
- enforce_admins=true (or non-admin): push **rejected** — the no-PR model breaks as-is.

## 5. Design axes — option space

### (a) Which contexts to require on develop

- **Safe set (no companion changes needed)**: `Test (ubuntu-latest)`, `Lint`,
  `Build (linux/amd64)` — unconditionally reported by ci.yml on every develop push (§2.1) —
  **plus `Analyze (Go) (go)`** (correction per the §2.2 refutation note: the codeql analyze
  job runs on every push event, so this context is also unconditionally reported on develop
  pushes; it is already required on main, so including it keeps develop/main parity).
- **Conditional set (needs companion change)**: `Race Test` (conditional, no skip-marker).
  `Constitution Check` appears unconditional but carries `continue-on-error` (ci.yml:537,
  audit-verified) — treat as ineligible until that is reconciled at design time.
- **Excluded**: `Release PR Multi-OS Gate` (structurally absent, §2), test-install contexts
  (paths-filtered → Pending hazard).

### (b) Lane push pass-through path

| Option | Mechanism | Companion changes | Cost / risk |
|---|---|---|---|
| **B1 — merge-commit 사전검증 (verify-branch seeding)** | Window: merge → push merge SHA to a verify ref → CI greens the SHA → push same SHA to develop (allowed per §3.3) | ci.yml (+codeql if used) push triggers widened to a verify pattern (e.g. `verify/**`) | Window hold extends by CI latency; duplicate run on develop push; broken merge caught BEFORE landing (today: only after) |
| **B2 — admin bypass (enforce_admins=false)** | Required checks bind non-admins only; operator credential (admin) pushes bypass; CI still runs post-push on develop | none | "필수" never gates the only actor that pushes (lanes push as admin); protection = force-push/deletion/non-admin block + post-push verification |
| **B3 — PR-mandatory on develop** | Mirror main's model | — | Contradicts card premise (무PR push 모델 공존); rejected as primary |
| **B4 — PR local-merge allowance** | Open card→develop PRs, merge locally, push (docs §3.4) | PR per card | Requires PR objects the model explicitly avoids; mechanics unverified without a PR; rejected |

### (c) enforce_admins

- `true`: real pre-push gate for everyone incl. operator — viable **only after** B1 is live.
- `false`: B2 semantics — model untouched, gate advisory for admins.
- Staged shape: B2 now → B1 companion changes → B1+true (operator decision at each step;
  the protection application itself stays an operator gh api gate per the card).

## 6. Risks / implicit contracts

1. **Window serialization × CI latency (B1)**: integration lock hold time grows by the CI run
   (~10-20 min full ci.yml on a green tree, this repo's runs). Batching multiple card merges
   per window amortizes it but blurs per-card CI attribution — operator trade-off.
2. **Rollout order matters**: applying enforce_admins=true (or required checks, for non-admins)
   BEFORE the B1 verify path works bricks the integration window. Order: workflow changes →
   window procedure update → protection application.
3. **7-day green precondition** (§3.7): the chosen contexts must already run green on develop
   (they do today — ci.yml/codeql fire on develop pushes).
4. **Duplicate CI on B1's second push** (verify → develop): same SHA, two push events. Cost,
   not correctness; concurrency-group keying is a design knob.
5. **`strict` (up-to-date) is PR-only semantics** — keep `strict: false` on develop; irrelevant
   to direct pushes.
6. **Doctrine drift** (§1.2): update §23.2-style tables if this SPEC lands, or cite live API as
   SSOT for protection state.
7. **Job-name uniqueness** (docs tip): required-check names must be unique across workflows —
   `Test (ubuntu-latest)` appears twice *within* ci.yml by design (paired marker), which is the
   documented-acceptable pattern; a same-named job in ANOTHER workflow would be ambiguous.

## 7. Recommendations (input to SPEC, not the SPEC)

1. Require the safe set (a) on develop: `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`,
   `Analyze (Go) (go)` (corrected — main parity, no companion change needed).
2. ~~Treat `Analyze (Go) (go)` as a phase-2 add-on gated on the codeql marker widening~~ —
   RETIRED by the §2.2 refutation (kept struck so the decision history stays legible).
3. Present B1/B2 as the operator decision axis with the staged path B2 → B1 → enforce_admins
   decision, per §5(c) — the SPEC delivers mechanics + runbook, the operator applies.
4. The SPEC is design-only (this card): no settings change, no workflow edit in scope; the
   companion workflow edits and window-procedure update are future run-phase work of this SPEC.
