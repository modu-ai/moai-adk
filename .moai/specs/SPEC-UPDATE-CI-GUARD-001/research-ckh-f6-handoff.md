# Research Input — F6 Pattern-Coverage Gaps (from SPEC-CONFIG-KEY-HONESTY-001 M6)

> Handoff from SPEC-CONFIG-KEY-HONESTY-001 (REQ-CKH-013, AC-CKH-013).
> Provenance: M6 measured three internal-content leaks in shipped template
> config files (`internal/template/templates/.moai/config/**`) that the CI guard
> (`internal/template/internal_content_leak_test.go`) could not see. M6 removed
> the leaked *content* (comment rewrites in workflow.yaml + llm.yaml); the
> *pattern-coverage gap* that let them pass is handed to this SPEC (E5) for a
> design decision. This SPEC does NOT modify the guard (plan §F M6, §H).

## The three measured evidence points

Each gap is stated with the exact measurement M6 took at HEAD `3f8f8458a`
(origin/main carries the same leaks at the same sites, so the gaps are live
upstream, not worktree-local).

### Gap 1 — `SPEC-AGENT-ARCH-V2-001` matches no registered C1/C1c family

The shipped `workflow.yaml` cited `SPEC-AGENT-ARCH-V2-001` in two comment lines
(the `model_routing` deprecation note and the `model_routing_profiles` policy
note). The CI guard's enumerated-family design did not catch them:

```
$ grep -rnE 'SPEC-[A-Z0-9]' internal/template/templates/.moai/config/  # pre-M6
.../sections/workflow.yaml:82:  # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). ...
.../sections/workflow.yaml:102:  # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001 ...
```

The guard at `internal/template/internal_content_leak_test.go` matches SPEC IDs
only via registered families. The C1 tree-wide class covers
`SPEC-(V3R[2-6]|AGENCY|WORKTREE)-` and a narrower class covers
`SPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}`. `SPEC-AGENT-ARCH-V2-001`
belongs to **no registered family** — the `AGENT-ARCH-V2` domain token is not
in any C1/C1c pattern, so the two citations passed the guard undetected.

**Design question for E5**: should the guard add a `AGENT-ARCH` family, move to
a broader `SPEC-[A-Z-]+-[0-9]+` wildcard, or keep the enumerated design and
accept that new SPEC domains must register their family proactively? The
enumerated design is deliberate (a generic wildcard would flag pedagogical
placeholders like `{SPEC-ID}`, `<SPEC-ID>`, `SPEC-XXX` throughout skill bodies),
so the wildcard question is a real trade-off, not a trivial fix.

### Gap 2 — C6 matches `PR #N` but not `issue #N`

The shipped `llm.yaml` cited `issue #653` in a comment explaining the GLM
context-window upstream misreport:

```
$ grep -rn 'issue #' internal/template/templates/.moai/config/  # pre-M6
.../sections/llm.yaml:179:  # (issue #653). Claude Code reports context_window_size ...
```

The C6 class matches `PR #N` (the `PR #` token followed by digits) but **not**
`issue #N`. The internal issue number `#653` therefore passed the guard
undetected.

**Design question for E5**: should C6 be widened to cover `issue #N` (and
possibly other internal-tracker tokens like `ticket #N`), or is the `PR #N`
scope intentional (only PR references are considered internal-content leaks,
while issue references are considered acceptable documentation)? Note that the
local `CLAUDE.local.md` and several `.claude/rules/` files DO cite `issue #653`
/ `Issue #653` — those are local/rule surfaces and outside the
`internal/template/templates/**` neutrality scope, so they are not in bounds for
this guard. The question is only whether the *shipped template* scope should
reject `issue #N`.

### Gap 3 — no class covers `plan.md §D D6`-shaped artifact citations

The shipped `workflow.yaml` cited the internal artifact `plan.md §D D6` (an
internal plan-document section reference) alongside the SPEC ID:

```
.../sections/workflow.yaml:82:  # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). ...
```

No guard class covers internal artifact citations of the `plan.md §D D6` shape
(an internal plan-document section reference). The C-classes enumerated in the
guard cover SPEC IDs, PR/issue numbers, `/Users/` paths, `REQ-`/`AC-` tokens,
`CLAUDE.local` references, ISO dates, and commit SHAs — but not internal
document cross-references like `plan.md §D D6` or `design.md §D.5` (the latter
also appeared in the workflow.yaml:102 leak).

**Design question for E5**: should a new class cover internal document-section
citations (`<file>.md §<section>`) in shipped templates, or are these considered
acceptable documentation cross-references? The hazard is that a shipped comment
citing `plan.md §D D6` points a distributed user at an internal artifact they do
not possess — the same honesty failure SPEC-CONFIG-KEY-HONESTY-001 exists to
address for config keys.

## What M6 did vs what E5 owns

| Surface | Owner | Action |
|---|---|---|
| The three leaked comment lines (workflow.yaml:82/102, llm.yaml:179) | SPEC-CONFIG-KEY-HONESTY-001 M6 | **Removed** — comments rewritten to drop the internal citation while preserving the mechanism description (REQ-CKH-012). Verified by spec.md §A.9 greps returning only generic placeholders. |
| The pattern-coverage gaps that let the leaks pass | SPEC-UPDATE-CI-GUARD-001 (E5) | **Handed off** — this note. E5 decides whether to extend/widen the guard classes or keep the enumerated design and accept the residual. |
| `internal/template/internal_content_leak_test.go` | SPEC-UPDATE-CI-GUARD-001 (E5) | **Not modified by M6** (REQ-CKH-013, plan §F M6, §H). Verified unmodified: `git diff --stat ed70e4354 HEAD -- internal/template/internal_content_leak_test.go` returns empty. |

## Verification of the handoff (AC-CKH-013)

The three evidence points above are the load-bearing content of this note. The
AC-CKH-013 grep:

```
$ grep -cE 'SPEC-AGENT-ARCH-V2-001|issue #N|plan\.md §' \
    .moai/specs/SPEC-UPDATE-CI-GUARD-001/research-ckh-f6-handoff.md
```

returns a count `>= 3` (each evidence-point section heading and body references
its token), satisfying the criterion.
