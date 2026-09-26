# t1259 — what a `develop` push actually verifies, measured

Measured by the lane (orchestrator) on 2026-09-26 in worktree
`.claude/worktrees/t1259`, at HEAD `28476f1a9`. Closes two gaps plan-audit iter2
recorded as unobserved, both of which bear on the D3-residual repair.

## Gap 1 — the `develop` required-check set

Claim: **`develop` has no required-check set, because the branch is not protected at all.**

```
$ gh api repos/modu-ai/moai-adk/branches/develop/protection --jq '.required_status_checks.contexts'
{"message":"Branch not protected", ... "status":"404"}
gh: Branch not protected (HTTP 404)
```

Consequence for the criterion: a clause phrased as "the required checks pass" has no
referent on `develop`. What exists is the set of workflow runs the push triggers, which
is a different thing — nothing blocks a merge on their result. A criterion should name
the **workflow run** it reads, not a protection gate that is absent.

Note the asymmetry with `main`, which `CLAUDE.local.md` §18 records as protected with
`enforce_admins: true`. Do not carry `main`'s posture onto `develop`.

## Gap 2 — whether the docs-site is built on a `develop` push

Claim: **the docs-site IS built, but by Vercel, conditionally, and not by Actions.**

Actions does not build it — confirmed again at this HEAD:

```
$ grep -rln 'hugo' .github/workflows/
(no output, exit 1)
```

Vercel does, and the configuration is in-repo:

```
$ sed -n '1,18p' docs-site/vercel.json
  "ignoreCommand": "git rev-parse -q --verify \"${VERCEL_GIT_PREVIOUS_SHA}^{commit}\" >/dev/null 2>&1 && git diff --quiet \"$VERCEL_GIT_PREVIOUS_SHA\" HEAD -- ':(top)docs-site' || exit 1",
  "framework": "hugo",
  "buildCommand": "hugo --minify --gc",
  "github": { "silent": true }
```

Three properties follow, and each one weakens a naive "the run includes the docs-site
build" clause:

1. **It is a different system with a different head.** The build is a Vercel deployment,
   not a job inside the Actions run the criterion names.
2. **It is conditional.** `ignoreCommand` skips the build unless the push changed
   something under `docs-site`. A `develop` push touching only Go files produces no
   docs-site build, so a criterion asserting one unconditionally is false on most pushes.
3. **`github.silent: true`** means the deployment does not report back onto the commit, so
   its result is not readable from the same surface as the Actions checks.

## What remains unmeasured

**Whether the Vercel project deploys `develop` at all** is NOT established here. That is
project-side configuration (production branch, preview-branch scope) and is not in the
repository, so it cannot be read from this tree. `git`-integration is configured and the
build command exists; which branches trigger it is not shown by any file measured above.

This is the load-bearing gap: it means a re-sited clause pointing at a Vercel deployment
would rest on an unverified premise. Dropping the build clause needs no such premise.

## Residual risk

Branch protection and Vercel project settings are both mutable outside this repository, so
both readings decay. The protection read in particular would flip silently if `develop` is
protected later — nothing in the tree would change.
