# t190 — sweep of `243eb07ef` for further dropped cross-references

Card class B (investigation), Tier S. Split out of t189's Gaps. Worked directly on
`release/v3.1.3`, no worktree, per the lead's dispatch.

**Outcome: no restoration warranted. Zero files changed.** The evidence below is the
basis, including why a clean result under the card's stated criterion is weaker than it
first appears.

---

## Claim

1. `243eb07ef` reduced 11 always-loaded files. Across all of them it removed the direct
   naming of 14 destination `.md` files (15 rows, one duplicated across stubs).
2. Every one of those destinations is still reachable: 5 remain named in the stub itself,
   and the other 9 moved into a companion the stub names explicitly in surviving prose.
3. No other assertion in the tree is broken by the commit — the three packages that assert
   on rule-file content pass on the current release HEAD.
4. **The criterion the card names does not, by itself, separate t189 from these 9.** Stated
   here because a "0 findings" report would otherwise read as stronger than the measurement
   supports.

---

## Evidence

### Scope of the commit

```
$ git show --stat 243eb07ef
  main-checkout-branch-guard  11,865 ->  6,395   moai-constitution   18,958 -> 15,433
  moai-mcp-tools               7,357 ->  2,389   agent-common-protocol 27,043 -> 24,645
  verification-claim-integrity 13,140 ->  8,224  askuser-protocol    23,504 -> 21,822
  cross-session-messaging     16,672 -> 11,823   native-idiom-and-register 4,967 -> 3,952
  context-window-management   13,009 ->  8,828   CLAUDE.md           20,748 -> 19,766
                                                 skill-routing        5,825 ->  5,595
```

### Method

For each of the 11 stubs: take every `*.md` basename appearing on a removed diff line and
not on an added one, then check whether that basename still appears (a) in the stub at the
current release HEAD and (b) in the stub's companion. Script: `/tmp/t190_sweep.py`
(throwaway; the table below is its verbatim output).

### Result, measured at HEAD `3590d9f40`

```
stub                            dropped destination                   stub   comp   verdict
moai-constitution.md            agent-authoring.md                    True   True   still in stub
moai-constitution.md            run.md                                False  True   companion only
verification-claim-integrity.md agent-common-protocol.md              False  True   companion only
verification-claim-integrity.md brain.md                              False  True   companion only
verification-claim-integrity.md context-discovery.md                  False  True   companion only
verification-claim-integrity.md manager-develop-prompt-template.md    False  True   companion only
verification-claim-integrity.md moai-constitution.md                  False  True   companion only
verification-claim-integrity.md moai.md                               True   True   still in stub
verification-claim-integrity.md plan.md                               False  True   companion only
verification-claim-integrity.md proposal.md                           False  True   companion only
verification-claim-integrity.md verification-batch-pattern.md         False  True   companion only
agent-common-protocol.md        session-handoff.md                    True   True   still in stub
context-window-management.md    session-handoff.md                    True   True   still in stub
cross-session-messaging.md      kanban-dispatch.md                    True   True   still in stub
cross-session-messaging.md      session-handoff.md                    True   True   still in stub
rows=15 unreachable=0
```

`agent-authoring.md` reads `True` in the stub because t189 restored it; the other 14 rows
are untouched by that fix.

### The 9 relocations carry a surviving signpost

Both stubs that dropped a name kept prose naming the companion the content moved to:

- `verification-claim-integrity.md:64` — "…live in the detail companion
  `verification-claim-integrity-detail.md`. Load it when composing an evidence-bearing
  report for the first time, or when tracing a clause back to its originating failure."
- `moai-constitution.md:165-167` — "Categories, the 50-file cap and archive path, the
  repo-local inbox drain contract, auto-capture triggers, the domain-matching algorithm,
  and the workflow integration points: `moai-constitution-detail.md` § Lessons Protocol."

### The destinations resolve

```
$ find .claude -name run.md -o -name plan.md -o -name context-discovery.md \
    -o -name manager-develop-prompt-template.md -o -name verification-batch-pattern.md \
    -o -name agent-common-protocol.md -o -name moai-constitution.md
.claude/commands/moai/run.md                     .claude/skills/moai/workflows/run.md
.claude/commands/moai/plan.md                    .claude/skills/moai/workflows/plan.md
.claude/rules/moai/core/moai-constitution.md     .claude/skills/moai/workflows/plan/context-discovery.md
.claude/rules/moai/core/agent-common-protocol.md
.claude/rules/moai/development/manager-develop-prompt-template.md
.claude/rules/moai/workflow/verification-batch-pattern.md
```

`brain.md` and `proposal.md` do not resolve, and correctly so: they appear at
`verification-claim-integrity-detail.md:87,94` inside the incident record about
recommending retention of a **retired** feature. Naming things that no longer exist is that
record's subject, not a broken link.

### No other assertion is broken

```
$ go test ./internal/cli/agentlint/ ./internal/constitution/ ./internal/runtime/ -count=1
ok  github.com/modu-ai/moai-adk/internal/cli/agentlint    0.802s
ok  github.com/modu-ai/moai-adk/internal/constitution     1.121s
ok  github.com/modu-ai/moai-adk/internal/runtime          1.488s
```

---

## Baseline-attribution

Measured by me in `.claude/worktrees/release-v313` on `release/v3.1.3` at HEAD
`3590d9f40` (`git rev-list --count --left-right origin/release/v3.1.3...HEAD` → `0 0`;
`git rev-parse -q --verify MERGE_HEAD` → empty; no tracked modifications before or after).

HEAD moved during this card — an earlier pass of the same sweep ran against the pre-`t182`
release ref. The table above is the re-run at `3590d9f40` and is the attributed one; both
passes produced identical rows.

## The negative control, and what it shows

Before trusting a `unreachable=0` result, the detector was run against `f7e5a05bd` — the
tree where t189's regression was still live and `TestConstitutionCrossReference` was
failing:

```
$ python3 /tmp/t190_sweep_pre.py | grep -E 'UNREACHABLE|rows='
rows=15 unreachable=0
```

**The detector did not flag the known positive.** Under the card's criterion — can the
reader reach the destination — t189 classified as "companion only", exactly like the other
9, because `moai-constitution-detail.md` named `agent-authoring.md` throughout. t189 was a
real regression for a different reason: `agent_lint_test.go:1248` asserts the literal
string `agent-authoring.md` inside `moai-constitution.md`.

So this sweep's clean result should be read as: **no destination was orphaned, and no
assertion is red** — not as "the diet was verified harmless by the reachability criterion",
which the negative control shows that criterion cannot establish on its own. The
assertion-package run above is the check that actually discriminates, and it is green.

A secondary pass looked for the t189 *shape* — prose describing a pointer without naming its
file — by grepping the 11 stubs for the word "pointer". One hit, in
`cross-session-messaging.md:46` ("a pointer into shared source of truth"), unrelated to file
references. No paraphrased file pointer remains.

---

## Gaps — what was NOT observed

- **Basename granularity.** The detector compares bare filenames, so a stub that still
  mentions a file *somewhere* counts as retaining it even if the specific pointer became a
  paraphrase. Five rows read `still in stub` on that basis and were not read line by line.
- **Section anchors were not checked.** A surviving `file.md § Section` pointer whose named
  section no longer exists would pass every check here. Nothing measured section validity.
- **Only `243eb07ef` was swept.** Other commits in the diet series (t82 had several
  milestones) were not examined; the card scoped this to one commit.
- **Template mirrors were not separately swept.** The mirror-parity test (`internal/template`)
  is what holds them equal, and it was not re-run in this card — t189 ran it green at
  `eff16c371`, one merge earlier.
- **No full-suite run.** Zero files changed, so nothing could regress; the batch PR's CI is
  the full-suite measurement.

## Residual-risk

- **The reachability criterion is weak, as the negative control shows.** A future diet pass
  could drop a pointer that some test asserts, and this card's method would not predict it.
  What catches that class is the assertion itself — and only where one exists. Most rule
  cross-references have no test behind them.
- **Two-hop reachability is a real cost even when it is not a defect.** Nine destinations now
  require reading the companion first. That was the diet's deliberate trade, recorded here so
  it is a known property rather than a rediscovery.
