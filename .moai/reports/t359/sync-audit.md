# Sync Audit — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

- Auditor: sync-auditor (independent, single writer)
- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
- Branch: `WT-landing-evidence`
- HEAD measured by this auditor at audit start: `f5d57fca4`, working tree clean
- Tier L · PASS threshold 0.85 · profile `default` (`harness.yaml` `default_profile: "default"`)
- Mode: flat weighted-percentage (`harness.yaml` sets no `evaluator_mode: hierarchical`)

---

## Overall Verdict

**FAIL — 0.75 vs the Tier L threshold 0.85. Score-driven; the must-pass firewall did NOT fire.**

Both must-pass dimensions PASS independently: Functionality (all 21 acceptance criteria met) and
Security (no Critical or High findings). The FAIL is arithmetic — every dimension sits one rubric
anchor below its ceiling — and it is routable: **one blocking finding (F1)**, plus one confirmed
availability defect the implementer had already escalated as an open operator decision (F2).

This is not a verdict against the work's honesty. The card's own disclosure record is the most
complete this auditor has read on any card: it names its vacuous clause, its stale criterion, its
undemonstrated RED, its absent linter and coverage, and its inherited build blocker, none of which
had to be volunteered. **F1 was found by doing precisely the thing the implementer disclosed as
not done** — `progress.md` Residual-risk: *"`landed` is unexercised as a shipped command … a defect
that lives only in cobra wiring or in the real binary's git-subprocess path would not have been
caught by anything in this card."* That gap was real, and it contained a defect. Disclosure of a gap
correctly predicts risk; it does not discharge it.

### Dimension Scores

| Dimension | Score | Verdict | Rubric anchor cited | Evidence |
|---|---|---|---|---|
| Functionality (40%) | 0.75 | PASS (must-pass) | *"All primary acceptance criteria pass; minor edge cases missing"* | 21/21 AC met; `go test` green on both packages; 8 criteria independently re-verified against a binary built from this tree |
| Security (25%) | 0.75 | PASS (must-pass) | *"No Critical/High findings; Medium findings documented with mitigations"* | One Medium (F1), mechanically demonstrated. Nearest anchor; falls short of its second clause — F1 was **not** documented or mitigated |
| Craft (20%) | 0.75 | PASS | *"Coverage >= 80%, minor style issues, acceptable naming"* | `internal/kanban` 86.2%, `internal/cli` 80.4%, `golangci-lint` **0 issues** |
| Consistency (15%) | 0.75 | PASS | *"Minor deviations from conventions; no structural inconsistencies"* | F1 is a localized deviation: the new function omits the sanitizer its same-package sibling applies with an explicit written rationale |

Weighted total: `0.40(0.75) + 0.25(0.75) + 0.20(0.75) + 0.15(0.75)` = **0.75**.

Must-pass firewall: Functionality PASS, Security PASS. Neither forces the FAIL; the score does.

---

## Findings

Reported at coverage, not filtered for importance. Confidence and severity are stated per finding;
`blocking` marks the ones that should gate the verdict being revisited.

### F1 — `ReadPrimarySpecStatus` joins an unvalidated `spec_id`, reading a file outside the project root
- **Severity: Medium · Confidence: HIGH (mechanically demonstrated) · BLOCKING**
- **Location:** `internal/kanban/status_read.go:281-289` (new in this card, +23 lines)
- **Reached from:** `internal/cli/todo_landed.go:300` ← `readCardSpecStatus` ← `moai todo landed`
- **Source of the tainted value:** `internal/cli/todo.go:584`, `moai todo next <n> --spec <ID>`, whose own comment records the decision: *"Recorded as-is: the store is not a SPEC registry; normalization is out of scope"*

The function builds `filepath.Join(primaryRoot, ".moai", "specs", specID, "spec.md")` with no
validation of `specID`. A `..`-bearing value escapes the project root. `filepath.Join` cleans an
absolute value back under the root, so absolute paths do **not** escape — only relative traversal
does.

This is a **deviation from a sibling in the same file**, not an oversight in isolation.
`ReadCardStatus` (`status_read.go:105-113`) guards the identical value class and states why:

> *"The spec identifier is interpolated into a worktree path, a git-show ref, and the primary-checkout spec.md path. A traversal-shaped value (`..`, separator, absolute) must be refused at this boundary rather than reach any join — the same sanitizer the CLI applies at its SPEC-ID boundaries (`internal/cli/specid.ValidateSpecID` …)."*

The new function reaches the same join shape and calls no sanitizer.

**Consequence.** The first `^status:\s*(\S+)$` line of an arbitrary readable file is stored in the
queue's `landing.spec_status` and rendered by `moai todo pr` to every reader of the shared queue.
It is a read-only, single-line disclosure — no write, no execution — which is why this is Medium and
not High: the operator or lane supplying `--spec` already holds read access on a shared checkout.
It is not Low, because the queue crosses actors (a lane sets `spec_id`; the lead's session renders
it) and because the value is persisted, not transient.

It is also a **correctness** deviation, which is what makes it blocking. REQ-TLE-005 and REQ-TLE-010
require the record to carry *"that SPEC's frontmatter `status`"*. Under a traversal `spec_id` the
verb stores a status read from a document that is not that card's SPEC, while
`LandingEvidence.Validate` (`landing_evidence.go:146`) — which polices every other field's
invariants — does not constrain `SpecStatus` at all.

**Required fix.** Call `specid.ValidateSpecID(specID)` in `ReadPrimarySpecStatus` before the join
and return `("", false)` on rejection, so a traversal-shaped `spec_id` maps onto the existing
`LandingSpecStatusUnknown` marker — the explicit "asked, unanswered" outcome REQ-TLE-010 already
defines — rather than onto a foreign file's contents. Validating at the `--spec` admission boundary
instead would be a wider blast radius and is the operator's call, not this fix's.

### F2 — one undecodable `landing` value makes the whole queue unreadable, and `--clear` cannot reach it
- **Severity: Medium · Confidence: HIGH (mechanically demonstrated) · not blocking — already escalated**
- **Location:** the decode-on-load path, `internal/kanban` (`decode landing evidence`), reached by every `todo` verb
- Disclosed by the implementer in `progress.md` as *"The open operator decision — recorded OPEN, not resolved"*

Independently confirmed: with a single corrupt cell, `moai todo` (list), `moai todo pr`, and
`moai todo landed <id> --clear` **all** exit 1 on the same load error. The documented escape hatch
is unreachable because it must load before it can clear. M5 extended the same choice to
`archived_items`, so the cost is paid on both tables. The consequence matches the SPEC's own
`§D.1` blocking-data class verbatim — *"an operator queue in the field becomes unopenable"* — yet no
acceptance criterion covers a malformed stored value.

Not classified blocking here for one reason only: the implementer surfaced it as an OPEN operator
decision with its data-loss rationale stated, which is exactly the shape the lead routes. The
rationale is sound (the encoder is the column's only writer, so an undecodable value is external
corruption and dropping it would destroy an operator record). Recommended bounded fix, preserving
that rationale: let `--clear` operate without a full decode of the column it is about to erase.

### F3 — the `#nosec` annotation asserts the very property that is unverified
- **Severity: Low · Confidence: HIGH · not blocking (subsumed by F1's fix)**
- **Location:** `internal/kanban/status_read.go:285` — `// #nosec G304 -- project-local SPEC path`

G304 is *"file path provided as taint input"* — the exact class of F1. The justification asserts the
path is project-local, which is the unchecked premise. The suppression silences the scanner that
would otherwise have flagged this line. Worth separate mention because removing the traversal
without revisiting the annotation leaves a false warrant in place.

### F4 — user-supplied `--ref` reaches git subcommands with no `--` end-of-options separator
- **Severity: Low · Confidence: HIGH (probed; currently not exploitable) · not blocking**
- **Location:** `internal/cli/todo_landed.go:161, 216`

`ref` is interpolated into `git rev-parse --verify --quiet <ref>^{commit}` and
`git merge-base --is-ancestor <resolved> <ref>` without `--`. Probed with
`--ref '--output=/tmp/t359_pwn'`: refused with rc=1 and **no file created** — neither reachable
subcommand exposes a write-capable option. Latent hardening only; a future subcommand change would
make it live.

#### F4 — REMEDIATION CORRECTED (appended by the lane orchestrator, post-verdict)

The finding above stands unchanged and is not disputed: user-supplied `--ref` did reach git
subcommands with no end-of-options guard, and the probe recorded with it is a real measurement.
**Only the prescribed token is corrected.** This note is appended rather than edited into the body
so the auditor's text stays as written.

**The original prescription**, quoted verbatim from the heading and body above: *"with no `--`
end-of-options separator"* … *"without `--`"*. That names two different git tokens as one. Applied
literally it breaks the verb. Measured on `git version 2.50.1 (Apple Git-155)`:

```
$ git rev-parse --verify --quiet 'HEAD^{commit}'                    → rc=0  <sha>
$ git rev-parse --verify --quiet -- 'HEAD^{commit}'                 → rc=1  NO OUTPUT   ← breaks
$ git rev-parse --verify --quiet --end-of-options 'HEAD^{commit}'   → rc=0  <sha>
$ git merge-base --is-ancestor --end-of-options HEAD~1 HEAD         → rc=0
$ git rev-parse --verify --quiet --end-of-options '--output=/tmp/t359_probe'
                                                                    → rc=1, no file created
```

In `rev-parse`, `--` separates revisions from **paths**, so a revision placed after it is read as a
path and resolves to nothing. `--end-of-options` is the token that stops option parsing while
leaving the operand a revision. Reproduced independently by the lane lead and by the fixing agent.

**Adopted: `--end-of-options`**, at **three** call sites rather than the two this finding lists —
`todo_landed.go:161`, `:205` (`validateSuppliedSHA`'s own `rev-parse` on `--sha`, same class, absent
from the finding's text), and `:216`. Guarding two of three would have left an incoherent partial
guard. It introduces a git 2.24+ (Nov 2019) floor where this project has declared none.

**One correction in the other direction, owed to the finding.** The lane orchestrator wrote that
guarded and unguarded were "indistinguishable by any probe available today". That is true of the
WRITE probe only. At the raw-ref site the injection is mechanically observable in isolation:

```
$ git merge-base --is-ancestor <sha> --independent                  → rc=129
    error: options '--independent' and '--is-ancestor' cannot be used together   ← consumed as an OPTION
$ git merge-base --is-ancestor --end-of-options <sha> --independent → rc=128
    fatal: Not a valid object name --independent                                 ← treated as an OPERAND
```

It remains unreachable through the CLI, and the structural reason is sharper than this finding's
"neither subcommand exposes a write-capable option": the `^{commit}` gate in `buildLandingEvidence`
runs first and refuses every option-shaped ref, so the raw-ref site never receives one. Not
exploitable today **by construction**; never "safe", since the immunity rests on incidental string
concatenation. `TestLandedOptionShapedRefIsRefused` is the alarm if that changes.

Fix landed in `b6e09cd31`; evidence in `.moai/reports/t359/f4`-bearing sections of
`.moai/reports/t359/f1-fix-evidence.md` and its F4 companion.

### F5 — AC-TLE-010 cannot fail on a malformed `spec_id`, so REQ-TLE-010 is violable while its criterion passes
- **Severity: Low-Medium · Confidence: HIGH · not blocking (pairs with F1)**
- **Location:** `acceptance.md:126-135`

The criterion tests two cases — a readable fixture SPEC and a nonexistent one — and passes honestly
on both (verified below). Neither varies the *shape* of `spec_id`, so the criterion is structurally
incapable of detecting F1. This is the "a criterion can read as strict and measure nothing" pattern
the card's own Residual-risk names, occurring once more than the three instances it lists.

### F6 — docs-site omits the new verb in all four locales
- **Severity: Low · Confidence: HIGH (measured) · not blocking — out of module, correctly handed off**

`docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md`: `grep -c 'todo landed'` = **0 in all
four**. A real user-facing gap. Outside the `module:` list, and the DoD's final line forbids
touching it, so the implementer's disposition (a follow-up card, the operator's to issue) is correct
— this auditor meets it as an inheritance, not a defect of this card.

### F7 — the `[HARD]` operator-act paragraph is unguarded against mirror drift
- **Severity: Low · Confidence: HIGH · not blocking** — disclosed by the implementer

Measured: the two doctrine surfaces are currently **byte-identical in full** (`diff` rc=0), which is
stronger than AC-TLE-021 requires. The gap is that AC-TLE-021 compares two extracted rows only, so
nothing would *catch* future drift on that paragraph.

### F8 — `make build` blocked at `agents-emit-check` — INHERITED, confirmed not this card's doing
- **Severity: Low · Confidence: HIGH (measured) · not blocking for this card**

Reproduced: rc=2, drift on `.codex/agents/moai/sync-auditor.toml`. Confirmed inherited on two
independent measurements: this card's diff against the merge-base touches **zero** files under
`.claude/agents/`, `templates/.claude/agents/`, or `templates/.codex/agents/`; and
`git merge-base --is-ancestor b65e7e5f6 HEAD` returns 1, so the regenerating commit is not yet on
this branch. The implementer's account is accurate. The integration lane must re-measure `make build`
on the **merged** tree rather than inherit the prediction that it clears.

### F9 — `internal/cli` package coverage is 80.4%, below the profile's 85% threshold
- **Severity: Low · Confidence: HIGH for the figure, NOT ESTABLISHED for its cause · not blocking**

Measured 80.4%. **I did not measure the pre-change baseline**, so I cannot claim the shortfall is
inherited — see Gaps. What is measured: this card's own files sit at or above the bar
(`todo_landed.go` 85.9%, `todo_pr.go` 91.6%, `landing_evidence.go` 94.4%, `backlog_sqlite.go` 85.9%).
Scored at the 0.75 anchor (*"Coverage >= 80%"*), not as a Craft FAIL.

### Non-finding, recorded because it looked like one

`go tool cover` reports `ReadPrimarySpecStatus` at **0.0%** in the `internal/kanban` profile. That is
a cross-package attribution artifact, **not** an untested function: `TestTodoLanded_SpecStatusRead
NeverInvented` (`internal/cli/todo_landed_test.go:417`) exercises it through the real call path with
real fixture files, and its second phase rewrites the fixture's frontmatter and re-asserts, which
defeats a hard-coded return. AC-TLE-010 is honestly verified. Recorded so a later reader does not
re-raise it.

---

## Independent Verification

### Claim
All four gates the dispatch names are green at `f5d57fca4`; the DoD's two unmeasured items (project
linter, coverage) are clean/measured; and eight acceptance criteria hold when re-verified against a
binary built from this tree rather than against in-process fixtures.

### Evidence

Gates — every verdict taken from `$?`, unfiltered:

```
$ go test ./internal/kanban/... -count=1 -timeout 600s   ; echo rc=$?
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.741s
kanban rc=0

$ go test ./internal/cli/... -count=1 -timeout 900s      ; echo rc=$?
ok  	github.com/modu-ai/moai-adk/internal/cli	448.610s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	2.578s
   [15 further ./internal/cli/... packages, all ok]
cli rc=0

$ go vet ./internal/kanban/... ./internal/cli/...        ; echo rc=$?
vet rc=0        (zero bytes of output)
```

Unfiltered FAIL scan of the captured output, guarding against a `--- FAIL:` below a display cut:

```
$ /usr/bin/grep -c -- '--- FAIL\|^FAIL' kanban.txt   →  0
$ /usr/bin/grep -c -- '--- FAIL\|^FAIL' cli.txt      →  0
```

The two DoD items the run phase recorded as never performed:

```
$ golangci-lint run ./internal/kanban/... ./internal/cli/...   ; echo rc=$?
0 issues.
lint rc=0

$ go test ./internal/kanban/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.373s	coverage: 86.2% of statements
$ go test ./internal/cli/ -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/cli	431.928s	coverage: 80.4% of statements
```

Per-file statement coverage, computed from `-coverprofile` (not a mean of function percentages,
which is a different statistic):

```
todo_landed.go       stmts=85   covered=73  = 85.9%
todo_pr.go           stmts=83   covered=76  = 91.6%
landing_evidence.go  stmts=36   covered=34  = 94.4%
backlog_sqlite.go    stmts=92   covered=79  = 85.9%
status_read.go       stmts=84   covered=67  = 79.8%   (majority pre-existing code)
backlog_migrate.go   stmts=311  covered=244 = 78.5%   (majority pre-existing code)
```

**F1, demonstrated end-to-end** — a binary built from this tree (`go build ./cmd/moai`), an isolated
git project, and a `spec.md` planted **outside** the project root:

```
$ moai todo add "audit probe card"          → t1
$ moai todo next 1 --spec '../../../secret' → picked t1 audit probe card
$ moai todo landed t1 --ref HEAD            → landed t1 ref=HEAD marker=ref-head   (rc=0)

$ strings .moai/state/todo/backlog.db | grep spec_status
{"ref":"HEAD","ref_head":"8b05f465104d742b058b54926d771c7fc9441209",
 "observed_at":"2026-09-08T03:04:24Z","spec_status":"PWNED-TRAVERSAL"}
```

`PWNED-TRAVERSAL` is the frontmatter of a file outside the project root. Inferred from nothing; the
stored bytes were read back out of the queue database.

**F2, demonstrated** — one cell corrupted via `sqlite3`, then every read verb, exit codes unfiltered:

```
$ moai todo               ; echo rc=$?   → rc=1   load backlog …: decode landing evidence: invalid character 'n'
$ moai todo landed t1 --clear ; echo rc=$? → rc=1   (same error — the escape hatch cannot run)
$ moai todo pr            ; echo rc=$?   → rc=1
```

**Criteria re-verified against the real binary** (the disclosed "unexercised as a shipped command"
gap), each naming the assertion that fires:

| Criterion | Observation | Firing assertion |
|---|---|---|
| AC-TLE-015 / 021 | `fields=7`, `[7]="column count probe"` | tab-field count = 7 **and** card text is last |
| AC-TLE-006 | card with no record renders `[6]=` empty | evidence cell empty, not `not-landed` |
| AC-TLE-016 | `(ref-head)` → `(operator)` after `--sha` | marker string differs by provenance, not by SHA value |
| AC-TLE-009 | re-record replaced; `--clear` → `[6]=` empty | one record, then none |
| AC-TLE-008 | `state=queued` across record / replace / clear | state unchanged by every evidence write |
| AC-TLE-020 (existence) | `the existence check failed: "deadbeef…" names no commit` | rc=1, named check, nothing written |
| AC-TLE-020 (reachability) | `the reachability check failed: commit 637eb035… is not reachable from HEAD` | rc=1, **distinct** branch from existence |
| AC-TLE-014 | `shasum` of every non-`.git` file, before vs after `todo pr` | identical → byte-identical, rc=0 |
| AC-TLE-018 | live DB `meta` table: `('schema_version','1')` | version not bumped |
| AC-TLE-001/002/019 | `PRAGMA table_info` on both tables | `landing TEXT` nullable, no default, last column, on both |

Assertions checked for vacuity rather than assumed:
- **AC-TLE-011** (`prlink_landed_attribution_test.go`) probes both the full SHA **and** the
  7-character abbreviation — the width the leak actually stores, which is the repair the card
  discloses — carries a fixture-premise check at `:133` and a poisoned-string positive control at
  `:188-191`. Not vacuous.
- **AC-TLE-019** pins exact ordered `(name,type,notnull,dflt_value)` tuples per table via string
  equality on `PRAGMA` output, with `quote()` separating `''` from `NULL`. String equality has no
  vacuity mode; a plant on either table cannot pass. No mutant planted — the tuples were instead
  confirmed against a live database (above), and the shared tree has one writer, so I declined a
  source mutation.
- **No SQL injection**: the only interpolated statement, `backlog_sqlite.go:363`, ranges over the
  compile-time constant `landingCarryingTables` with a constant column name.

### Baseline-attribution
Every figure above was produced in this run, in this tree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`), at HEAD `f5d57fca4`, working tree clean at
audit start and carrying only this auditor's own commits thereafter. Nothing is carried from
`.moai/reports/t359/sync-verify/` (the sync agent's captures at `b306a6148`), from `progress.md`
§E.2/§E.3, or from the lead's dispatch — the dispatch's HEAD was re-read and matched, and every
other figure it offered was treated as a claim to check. The probe binary was built from this HEAD;
the two probe projects are throwaway git repositories under the session scratchpad, never this tree.

### Gaps — what was NOT observed
- **Pre-change coverage baseline.** I did not measure `internal/cli` coverage at the merge-base, so
  the 80.4% figure is **not attributed** to this card or to its predecessor. F9 states this rather
  than resolving it. The reasoning that a 306-line addition cannot have moved a multi-thousand-line
  package from ≥85% to 80.4% is an argument, not a measurement.
- **No mutants planted by this auditor.** The five run-phase mutants and AC-TLE-019's three plants
  are re-read as records, not re-observed. I declined to mutate a shared tree; where a criterion's
  discriminating power was in question I established it by construction or by an independent probe
  instead. AC-TLE-018's disclosed undemonstrated RED is therefore confirmed only as a *reading* of
  the criterion text.
- **Cross-platform.** darwin/arm64 only. No linux, no windows, no `GOOS` cross-build. CI's verdict.
- **No `go test -race`**, and no concurrency probe of `ensureLandingColumn` under concurrent open.
- **The merged tree is unmeasured.** Every green here is this branch's tip in isolation.
  `make build` must be re-measured after the absorb; `catalog.yaml` is modified by both this card
  and `b65e7e5f6` and should be expected to conflict and be resolved by regenerating.
- **The full local suite was not run**, by `CLAUDE.local.md` §4 policy. Packages outside
  `internal/kanban` and `internal/cli` are unmeasured at this HEAD.
- **F1's blast radius beyond `spec_status`** was not swept: I confirmed one sink, and did not
  enumerate every other consumer of a stored `spec_id` across the codebase.
- The `.moai/reports/t359/sync-verify/` captures were **not** re-derived line by line; I re-ran the
  gates myself instead, so those files are neither corroborated nor contradicted here.

### Residual-risk
- **F1's severity rests on a threat model, and the model is arguable.** I scored Medium because
  `--spec` is an operator act on a local CLI where the actor already holds file-read access. A lead
  who regards the shared queue as a genuine trust boundary between lanes would reasonably score it
  High — and under this profile High is a must-pass failure, which would move the verdict from a
  score-driven FAIL to a firewall FAIL. The measurement is not in doubt; its severity classification
  is a judgement the lead may overrule.
- **The score is flat 0.75 because the rubric's anchors are 0.25 apart.** With a 0.85 threshold,
  a Tier L card effectively needs 1.00s. One genuinely-clean dimension would not have changed the
  outcome (Craft at 1.00 yields 0.80, still below 0.85). The FAIL is real but it is close to the
  rubric's granularity, and the distance from PASS is smaller than the token "FAIL" suggests.
- **Absence of further findings is not evidence of their absence.** I read the implementation and
  probed the paths the dispatch's map named; a defect in the JSON⇄SQLite downgrade path under a
  hand-edited `backlog.json`, or in the archive round trip, would plausibly have escaped this audit.
- **The audit's own probes wrote only to the scratchpad**, but the coverage and gate runs put this
  machine under sustained load for roughly 25 minutes; a flaky timing-sensitive test elsewhere could
  have been perturbed by that load without appearing in these packages' results.

---

## Recommendations

1. **Fix F1** — the one blocking item. `specid.ValidateSpecID` in `ReadPrimarySpecStatus`, returning
   `("", false)` on rejection so it degrades to the existing `LandingSpecStatusUnknown` marker.
   Revisit the `#nosec` annotation (F3) in the same edit. A criterion closing F5 would be the honest
   companion, since AC-TLE-010 as written cannot detect the regression.
2. **Decide F2** — it is the lead's, as the implementer said. The bounded option that keeps the
   data-loss rationale intact is letting `--clear` run without decoding the column it erases.
3. **Re-measure on the merged tree** — `make build`, and the `catalog.yaml` conflict, resolved by
   regenerating rather than by picking a side.
4. **Issue the docs-site follow-up card (F6)** — four locales, verb absent, correctly out of scope
   here.
5. **Carry F4 and F7 as hardening**, not as gates.

