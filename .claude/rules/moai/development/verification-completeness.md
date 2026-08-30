---
description: "Verification completeness — a check, gate, acceptance criterion, rule, or assertion is incomplete until its failure has been observed on a known input"
paths: "**/.moai/specs/**,**/.claude/rules/**,internal/template/templates/.claude/rules/**,internal/hook/**,scripts/**,**/.moai/astgrep-rules/**,**/.moai/hooks/**"
---

# Verification Completeness

> Loading scope: this rule keys the surfaces where verification artifacts are authored — SPEC
> artifacts, rule files on both the local and template trees, hook and gate code, check scripts,
> and rule rulesets. It is paths-scoped and never joins the always-loaded surface.

**The single completion axis:** a verification artifact (check, gate, acceptance criterion,
rule, or assertion) is incomplete until its failure has been observed on a known input.

Claim integrity (`.claude/rules/moai/core/verification-claim-integrity.md`) stops an actor from
asserting a verification it never ran. This rule works one layer earlier: it stops an actor from
shipping a verification instrument that cannot deliver a verdict at all. The claim rule polices
false PASS reports; this rule polices checks with no observed failure on record — an instrument
whose red has never been seen has proven nothing, whatever its green output says.

## 1. The completion axis

### 1.1 Observed-failure completion

[ZONE:Evolvable] [HARD] Creating a check is not completing it. A verification artifact is
complete only when its failure has been executed and observed on a known failing input — an
input constructed to make the artifact red. Writing the check, wiring it into a gate, and
watching it pass green are pre-completion acts; until the red has been seen, the green is
uninterpreted output that could come from the check working, the check matching nothing, or the
check never running at all.

The recurring defect form is **report-not-verdict**: the tool prints its findings and exits with
the status of whatever ran last — a probe that prints a return code but returns the exit status
of its own trailing success message; a citation sweep that prints unverified citations and still
exits zero; an acceptance criterion that invokes a check but never uses its result, stalling
review at the same score round after round. Each produces output that looks like verification
while carrying no decision.

> Evidence: observed as a probe that printed a return code yet returned the exit status of its
> own trailing success message — always zero — so an injected failure never surfaced; the same
> defect reappeared one tool later in a citation sweep that printed unverified citations while
> exiting zero; on the criterion side, review rounds stalled at an identical score with
> comments of the form "invokes the check but never uses its result".

[ZONE:Evolvable] [HARD] **A pass whose swept set is empty asserts nothing.** A verification that
selected nothing still reports success — a test-name selector matching zero tests, a zero-hit grep
read as a pass, a suite filtered down to no cases — and the report is indistinguishable from one
where everything passed. Counting what was actually swept is therefore part of the same completion
act as observing the failure, not a refinement of it: a green with an empty swept set is
uninterpreted output, and reading it as a pass is an unobserved-verification claim.

The obligation this places on an actor is to establish the swept count before reading the verdict,
and on a verification instrument to make an empty sweep visible rather than silent. The runner
usually says so in its own words — go prints `[no tests to run]` or `[no test files]`, pytest
`no tests ran`, jest `No tests found`, vitest `No test files found`, cargo a zero pass count — and
those tokens are the cheapest available evidence. Where an instrument classifies runner output, the
empty-sweep judgment is made ahead of every other signal, because the exit code of a run that swept
nothing is the same zero a fully-passing run returns.

### 1.2 The three-part check spec

[ZONE:Evolvable] [HARD] Every check specification states three parts together, and a check with
any part missing is unfinished:

- **(a) WHEN** it must run to be meaningful. A check scheduled at a structurally always-green
  moment — before the condition it guards can possibly differ — proves nothing by passing.
- **(b) the INPUT** that turns it red. The specification names a known failing input, and the
  check is unfinished until that failure has actually been observed (§1.1).
- **(c) the failure's reachability** — who sees the red: which exit code, which log level, which
  trace. A red visible to nobody is a red that never happened.

Defect forms: (a)-missing produces an always-green check — a constraint demanded as a check was
scheduled at a milestone where its condition could not yet differ, so it passed vacuously;
(c)-missing produces an invisible red — a runner logged its own failure at debug level, silent
in production and absent from traces: the failure was reached but not observed.

> Evidence: observed as a release-ordering constraint demanded as a check yet scheduled at a
> milestone where the ordering could not yet differ — a check incapable of failing that passed
> anyway; and as a runner that logged its own failure at debug level, silent in production and
> absent from traces — reachability designed, observability absent.

### 1.3 Continued firing

[ZONE:Evolvable] [HARD] A check's completion does not survive a change to its trigger, its
deployment, or its branch model. §1.2(a) asks WHEN a check must run to be meaningful and answers
it once, at authoring time; nothing in that answer keeps holding. A check correctly scheduled on
Monday can stop firing on Friday because the event it subscribed to was abolished, because the
binary carrying it is older than the fix it was supposed to run, or because a selector that once
matched now matches nothing. The check is still in the tree, still correct, and no longer runs.

**This is absent execution, not suppressed failure, and the distinction decides what to build.**
§1.2(c) routes a red to someone who will see it; that clause presupposes a red exists. Here
nothing ran, so nothing could go red, and there is no signal to route — no exit code, no log
line, no trace. The accurate sentence is not "the check should have failed and did not"; it is
**"nothing failed, and there was nothing there to fail."** A green that follows is true about the
set that was selected and silent about the set that was intended, and nothing in the mechanism
compares the two.

Three ways a live check goes quiet, each observed:

- **Trigger.** A branch-model change abolished the event two guards subscribed to; both stopped
  running on the integration branch and neither announced it. Their absence from a default run
  listing was indistinguishable from a path filter declining to match.
- **Deployment.** A fix landed and the installed binary was built from an older commit, so the
  fixed code never ran. Three observers each independently reproduced the pre-fix behaviour and a
  card was issued for a defect that was already repaired.
- **Selection.** A test-name selector named three tests of which one existed. The run printed a
  passing line and a duration; the two absent tests never reached the exit code.

**The completion act.** A check specification is unfinished until it also states its
continued-firing answer: how a reader learns the check has stopped firing, as a stale-guard signal
that arrives without being asked for.
A verdict available only to whoever thinks to query it has relocated the problem into the person
expected to already know the question, which is precisely the person who does not have it: the
targeted query that recovers a missing guard can only be issued by someone who already suspects
the answer. Liveness that answers on demand is not liveness; it is a second thing to remember to
check.

The general form, and the test to apply to any check: **any check whose non-execution is
indistinguishable from its success has this defect.** Ask of each one — if this stopped running
tomorrow, what would be different in what I see? Where the honest answer is "nothing", the check
is complete against §1.1 and §1.2 and still unfinished here.

> Evidence: observed as two guards silently unsubscribed by a branch-model change, their
> non-firing indistinguishable from a non-matching path filter in the default listing; as a fix
> that landed while the installed binary predated it, reproduced as a live defect by three
> independent observers; and as a selector naming three tests of which one existed, the run
> printing a pass whose swept set was two-thirds empty.

## 2. Two-cell adoption discipline

[ZONE:Evolvable] [HARD] Adopting an acceptance criterion takes two cells authored as a pair: a
**RED-now cell** — the criterion observed red on the pre-implementation tree, pinned to the tree
it was measured on — and a **green path cell** naming which milestone flips it and what the
passing output becomes. A criterion with one cell is unadopted: RED-now alone does not certify
that this work can flip it, and a green path alone is a promise with no starting observation.

[ZONE:Evolvable] [HARD] RED must be red **for the right stated reason**, and the RED cell must
say why it is red. Three failure directions, only one of which RED-now catches by itself:

- **Vacuous** — green today; the criterion asserts nothing about the work. Caught by requiring
  RED-now.
- **Impossible** — red today and red forever; no correct work can satisfy it. Sails through
  RED-now, because it is red at arrival.
- **Wrong-reason red** — red at arrival and still red after implementation, because of
  pre-existing files the work never touches. Indistinguishable from the impossible direction
  unless the RED cell states why it is red.

A green path is disqualified when it runs through "someone fixes the unrelated files", or when
no change this work can make would flip it — either way the criterion does not measure this
work.

> Evidence: observed as a runner reporting "ok, 0 passed; 0 failed" with exit zero on a tree
> containing zero rule tests — a vacuous criterion passing without exercising anything — next to
> a count-equals-zero criterion that no correct work could ever satisfy — red at arrival and red
> forever; requiring RED-now caught the first and waved the second through, which is why the two
> cells are a pair.

[ZONE:Evolvable] [HARD] **Mutant probe.** Before adopting a criterion, try to write a mutant
that satisfies the criterion while violating its requirement. If such a mutant is writable, the
criterion is too shallow to adopt. Rule-pairing corollary: an invalid-cases-only rule passes an
all-matching mutant; a valid-cases-only rule passes a nothing-matching one — a ruleset asserted
by a single-direction criterion admits a mutant on the other direction.

> Evidence: observed as stalled review commentary of the form "invokes the check but never uses
> its result" — exactly the shape of criterion a mutant satisfies while the requirement is
> violated — and as a paired-rule review where an invalid-cases-only rule admitted an
> all-matching mutant while a valid-cases-only rule admitted a nothing-matching one, each
> direction passing its own one-sided criterion.

## 3. Cross-layer revision sweep

[ZONE:Evolvable] [HARD] The layer a rule constrains is its blind spot. When a criterion is
rescoped or revised, sweep the requirement and plan items it cites in the same pass — a revision
does not end in the file it started in. A revised criterion whose requirement still carries the
old scope is false at arrival: the requirement reaches files the work never touches, so the
pre-work tree already fails it for reasons this work cannot fix. The direct contradiction pair —
a requirement demanding a regeneration step while the revised criterion asserts the same output
does not change, with the plan instructing the forbidden side — is the same defect one layer up.

> Evidence: observed as a rescoped criterion whose requirement still scoped the entire template
> tree, false at arrival because the untouched tree already carried dozens of internal tokens
> the work never touched; and as a requirement demanding a regeneration step while its paired
> criterion asserted that output never changes, the plan instructing the forbidden side.

## 4. Evidence pinning

[ZONE:Evolvable] [HARD] Invariant assertions — byte-unchanged, preserved-surface, absence
claims — pin the tree SHA where the evidence was collected, never a moving branch name. A moving
ref advances under the assertion: work that changed nothing reads as dozens of changed files
because upstream moved, and the assertion silently falsifies itself between measurement and
reading.

Corollaries: never re-cite a measured divergence without re-measuring it — a divergence read
once and quoted again later is a number about a tree that no longer exists; on rebase,
re-measure and re-pin.

**Discriminator — pinning is not unconditional.** Before pinning, ask whether a moving ref
flipping the claim would be a true signal about the subject or spurious red from unrelated
upstream work. Provenance-style statements about the mainline itself — "this branch descends
from the remote mainline" — keep the moving ref, because pinning them to a fixed SHA would
weaken exactly what they assert. Without this discriminator, the next author pins everything
mechanically and provenance claims rot.

> Evidence: observed as a preserve-proof measured against a moving branch name that reported
> dozens of changed files under zero local work — upstream had simply advanced — while the same
> measurement at a fixed tree SHA was empty; the same document re-quoted a divergence hundreds
> of commits stale without re-measuring; one question separates the two cases: is the flip a
> signal about the subject, or noise from upstream?

## 5. Corollaries

- **Audit-verification.** To verify that a fix landed, do not grep for the fixed form — run the
  two command forms directly and observe them diverge. The cheapest modification that satisfies
  a grep for the fixed form is not the modification that fixes the defect; only observing the
  pre-fix form fail and the post-fix form pass on the same tree is a verdict.
- **Self-application.** A rule's own text must comply with the rule. A comment warning that a
  number is measured against a moving reference must not state its own number without pinning
  it. Applied to this rule: each clause above carries its observed failure, acceptance criteria
  written under this rule follow the two-cell pair of §2, and the evidence rows state the defect
  forms their clauses prohibit.

---

Version: 1.0.0
Classification: Path-scoped doctrine rule — binds verification-artifact authoring (checks, gates, acceptance criteria, rules, assertions); changes no gate semantics.
