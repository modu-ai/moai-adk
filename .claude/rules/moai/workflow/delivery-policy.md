# Delivery Policy

This rule is the single owner of branch, push, and pull-request decisions for
plan, run, and sync.

1. Phase agents create commits but do not push, create PRs, merge PRs, or
   switch the shared primary checkout. `manager-git` owns those operations.
2. Tier S, M, and L all use the same protected delivery decision. A normal
   change is delivered through a feature branch and PR. The only no-PR path is
   an explicitly configured `git-flow` `WT-*` integration branch merged into
   the designated local integration worktree; it is still pushed by
   `manager-git` only after the local merge and required checks.
3. `--pr` selects the PR route but never grants an early push. The route table
   records the base branch, owner, approval requirement, and final push target.
4. A phase cannot infer a direct push from its Tier. If route metadata is
   missing or contradictory, stop delivery and return the resolved policy as a
   blocker.

The minimum delivery record is:

```text
route=<pr|wt-integration>
tier=<S|M|L>
owner=manager-git
base=<branch>
local_merge=<sha or none>
approval=<required|not-applicable>
push=<deferred-until-route-complete>
```
