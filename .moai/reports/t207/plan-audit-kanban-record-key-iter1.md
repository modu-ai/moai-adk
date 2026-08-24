# SPEC Review Report: SPEC-KANBAN-RECORD-SESSION-KEY-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.75** (harmonic mean of the four dimensions; threshold 0.80)
Blocking findings: **5** — Optional findings: **6**
Premise reproduced: **YES, in shape — with a fresher and stronger instance than the SPEC's own; the SPEC's three cited session ids no longer exist.**

Reasoning context ignored per M1 Context Isolation. The audit reads only the three SPEC artifacts,
the code they cite, the two sibling SPECs named for boundary checking, and the live on-disk state.
Nothing was repaired or edited.

Audited tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`, HEAD `ee039da30`, branch
`WT-web-live-todo`. The SPEC's own measurements are attributed to `dfbf828a6`, which is an ancestor
commit in this tree (`git log --oneline -1 dfbf828a6` →
`dfbf828a6 plan(t207): measure the session-id slots and find the kanban record key defect (t207)`).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -o 'REQ-KRS-[0-9]*' spec.md | sort -u` → `REQ-KRS-001`
  … `REQ-KRS-008`, eight ids, sequential, zero-padded, no gap, no duplicate. Requirement-entry count
  `grep -c '^- \*\*REQ-KRS-' spec.md` → `8`, matching the id count exactly (no id appears twice as a
  definition).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md` §B) only, per M3 § Scope. All eight are GEARS-shaped: REQ-KRS-001/002/003 ubiquitous
  (`A session's kanban record shall be keyed by…`) with a trailing `shall not` unwanted clause;
  REQ-KRS-004 `**Where** a session is a factory lane, its record shall carry…`; REQ-KRS-005
  ubiquitous plus `**When** neither source yields a value, the field shall be left empty`;
  REQ-KRS-006 ubiquitous; REQ-KRS-007 ubiquitous plus `**When** a record written by a build
  predating this schema is read…`; REQ-KRS-008 `**When** a launch fact … is unavailable … the
  session shall record what it has`. No Given-When-Then entry appears in the requirement layer. The
  Given-When-Then form of every `AC-KRS-xxx` in `acceptance.md` is the correct verification-layer
  format and is graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present with correct types:
  `id`, `title`, `version: "0.1.0"` (quoted), `status: draft` (in the 8-value enum per
  `.claude/rules/moai/development/spec-frontmatter-schema.md:71`), `created: 2026-08-24`,
  `updated: 2026-08-24` (both ISO), `author`, `priority: P2`, `phase`, `module`, `lifecycle:
  spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at`/`updated_at`/`labels`/`spec_id`) appears. Extra fields `era`, `tier: M`,
  `related_specs` are additive and do not violate the schema.
- **[N/A → PASS] MP-4 language neutrality.** Single-language scope: the SPEC touches
  `internal/kanban`, `internal/cli`, `internal/config`, `internal/hook` only. Verified there is no
  template mirror for these paths — `ls internal/template/templates/internal` →
  `No such file or directory`, which confirms §C.3's claim. No 16-language enumeration obligation
  arises.
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** `grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` over the
  three artifacts yields exactly `SPEC-WEB-CONSOLE-015`, `SPEC-SESSION-TELEMETRY-001` (plus the
  SPEC's own id). Both directories exist and `grep -h '^status:'` on each returns `status: draft` —
  neither is `retired`/`superseded`/`archived`, so no reconciliation obligation fires. No BLOCKING
  D7 finding. (The stale *section* citation into `SPEC-WEB-CONSOLE-015` is D4 below; it is a
  citation-accuracy defect, not a D7 lifecycle finding.)
- **[N/A → PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` → `0`. D8-4
  auto-PASS; no cross-platform concern to gate.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION'` over the SPEC directory → no
  match (`rc=1`). `research.md` is correctly absent at Tier M; `plan.md` carries an
  `## §H Open verification items` section that explicitly declares its two items as
  "**not** blockers on this plan" — that is the correct disposition for a plan-phase open item and
  is not a clarification marker.

No must-pass criterion fails. The FAIL verdict is carried by the aggregate score (0.75 < 0.80) and
by the five blocking findings enumerated below.

---

## Premise Verification (the claim attacked hardest)

**Claim under test** — `kanban.Record` is keyed by the launching session's identifier because
`recordKanbanSession` resolves the key from a single project-wide slot, and the launcher runs before
the session it launches exists.

### The mechanism reproduces exactly, line by line

Every link in §A.1's chain was re-read at the cited line and matches:

```
$ grep -n 'func recordKanbanSession\|resolveLaunchSessionID("")\|kanban.WriteBestEffort' internal/cli/kanban.go
472:func recordKanbanSession(specID, backend, role string) {
474:	sessionID := resolveLaunchSessionID("")
478:	kanban.WriteBestEffort(projectRoot, kanban.NewRecord(sessionID, specID, backend).WithRole(role))
```

```
$ sed -n '126,134p' internal/cli/launcher_blockcap_infinite.go
func resolveLaunchSessionID(sessionOverride string) string {
	if sessionOverride != "" { return sessionOverride }
	if id, _, ok := resolveCurrentSessionID(); ok { return id }
	return ""
}
```

```
$ sed -n '52p' internal/session/registry.go
const CurrentSideChannelFile = ".moai/state/current-session-id.txt"
```

…whose own doc at `registry.go:45` reads "is overwritten on every SessionStart", and
`internal/hook/session_start.go:313-314` performs that overwrite with `input.SessionID`. All four
citations verify. §A.1 is sound.

### The consequence reproduces — with different data, and a stronger instance

The SPEC's three cited live sessions are **gone**. Re-measured:

```
$ python3 -c "…"  /Users/goos/MoAI/moai-adk-go/.moai/state/active-sessions.json
f85c1634-c55d-4ba9-9c72-94c1a67c85ee  pid 51045  cwd …/worktrees/t210  started 04:35:38Z
33468939-9717-4f5d-8d77-1582255bbc41  pid 34699  cwd …/worktrees/t209  started 04:36:01Z
2e3ace62-e8cf-4dfb-9742-fb4ba42750d6  pid 83078  cwd /Users/goos/moai/moai-adk-go  started 04:43:03.966Z
```

None of `2beac221…`, `c15d8434…`, `3db058e1…` appears; `ls .moai/state/kanban/3db058e1*.json` →
`no matches found`. **The SPEC's specific citation is stale.** Treated per the audit brief: this is a
*stale-but-still-true* observation, correctly attributed by the SPEC to a named commit and a named
date, so it is honest evidence about a past instant — not a claim that no longer holds.

The **shape** reproduces, and today's data is a cleaner reproduction than the SPEC's own:

```
$ ls .moai/state/kanban/f85c1634-….json   → No such file or directory
$ ls .moai/state/kanban/2e3ace62-….json   → No such file or directory
$ ls .moai/state/kanban/33468939-….json   → exists (201 bytes, mtime 13:43)
$ cat .moai/state/kanban/33468939-9717-4f5d-8d77-1582255bbc41.json
{ "session_id": "33468939-…", "spec_id": "", "role": "lead", "backend": "glm",
  "entered_at": "2026-08-24T04:43:00Z", "deepscan_dir": "", "verify_reentries": 0 }
```

The record filed under `33468939…` cannot describe `33468939…`: that session's own
`started_at` is `04:36:01Z`, seven minutes *before* the record's `entered_at` of `04:43:00Z`. Three
seconds *after* that instant, `2e3ace62…` registers — a session with no record of its own. The
record filed under the parent therefore describes the child, exactly as §A.1 predicts, and the
timestamps pin the launcher-runs-first ordering to within four seconds. This is a **live, fresher,
better-attributed instance of the defect than the one the SPEC cites.**

The SPEC's fourth citation still reproduces verbatim: `cat
.moai/state/kanban/d281730e-a47e-4f82-878e-5fd0ddc4dcb9.json` returns the quoted JSON with
`"role": "lane"` unchanged.

**Verdict on the premise: sound.** Both faults (single slot, launcher-before-child) are verified in
code and observed in live data. Nothing built on the premise fails for want of the premise.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — "minor ambiguity in one or two requirements" | REQ-KRS-005 admits two implementations for a primary-checkout session (D3). Every other requirement has a single reading; §A.1's chain table and §C.1's blast-radius table are unusually precise. |
| Completeness | 1.00 | 1.0 — all sections + frontmatter + Out of Scope | HISTORY, §A Background, §B Requirements, §C Constraints, §D Exclusions all present. Six `### Out of Scope — <topic>` H3 sub-headings, each with a specific `-` bullet. Frontmatter complete (MP-3). The disposition of the existing record files is stated (§C.4) as asked. |
| Testability | 0.50 | 0.50 — "several ACs … require judgment calls to evaluate" | Three verification statements do not do what they claim: AC-KRS-002's grep half is mis-baselined and unsatisfiable (D1); AC-KRS-001's baseline is not re-measurable as cited (D5); the DoD item pinned to "the 84 existing files" is not checkable (D5). The other six criteria are binary, command-named, and free of weasel words. |
| Traceability | 1.00 | 1.0 — every REQ covered, every AC parented, no orphan | §D table maps 8 REQ → 9 AC. Verified both directions: `grep -o 'AC-KRS-[0-9]*' acceptance.md \| sort -u` → 001…009, all nine appear in the table; every REQ-KRS-001…008 appears with at least one criterion. REQ-KRS-001 carries three (001, 002, 009). No AC references a non-existent REQ. |

**Aggregate = harmonic mean** (per the skeptical-evaluation contract, `agent-common-protocol.md`
§ Skeptical Evaluation Stance — "score quality as the harmonic mean of dimensions, not the
average"):

```
4 / (1/0.75 + 1/1.00 + 1/0.50 + 1/1.00)
  = 4 / (1.3333 + 1.0000 + 2.0000 + 1.0000)
  = 4 / 5.3333
  = 0.750
```

**0.750 < 0.80 (Tier M threshold).** The arithmetic mean would be 0.8125 — above threshold — which
is precisely why the contract specifies the harmonic mean: it refuses to let three strong dimensions
buy off one weak one.

---

## Defects Found (structured defect-list)

**D1.** `AC-KRS-002-GREP-CONTRADICTION` — `acceptance.md` §A (AC-KRS-002) — **The criterion's grep
half is both mis-baselined and unsatisfiable, and satisfying it would violate this SPEC's own t221
exclusion and its own Definition of Done.** The criterion states the grep
`grep -rn 'CurrentSideChannelFile|current-session-id|resolveLaunchSessionID' internal/kanban/ internal/hook/session_start*.go`
"returns zero hits", and asserts "Baseline: the same grep returns zero on the pre-change tree, so
the grep half asserts preservation". Measured on the untouched tree:

```
$ grep -rn 'CurrentSideChannelFile\|current-session-id\|resolveLaunchSessionID' internal/kanban/ internal/hook/session_start*.go
internal/hook/session_start_additional_context_test.go:58:  // side-channel file (.moai/state/current-session-id.txt) so `moai session
internal/hook/session_start_additional_context_test.go:77:  sidecar := filepath.Join(projectDir, session.CurrentSideChannelFile)
internal/hook/session_start_additional_context_test.go:188: sidecar := filepath.Join(projectDir, session.CurrentSideChannelFile)
internal/hook/session_start.go:288:  // side-channel file (.moai/state/current-session-id.txt) so `moai session
internal/hook/session_start.go:313:  sidecar := filepath.Join(input.ProjectDir, session.CurrentSideChannelFile)
internal/hook/session_start.go:315:  slog.Warn("session start: failed to write current-session-id side-channel file (non-blocking)",
```

**Six hits, not zero.** Three of them are in `internal/hook/session_start.go` — the sidecar *writer*
that §D's t221 exclusion and `acceptance.md` §F both require to remain untouched ("`.moai/state/current-session-id.txt`
and its writer unchanged (the t221 boundary)"). The only way to make this grep return zero is to
delete that writer, which the SPEC forbids. The defect is compounded by the milestone shape: M2
places the new record write into `internal/hook/session_start.go` "(or a new kanban-record sibling
in the same package)" — the `session_start*.go` glob is exactly where the sidecar write already
lives, so the file set can never be clean.
— Severity: **critical** — Class: **blocking** — Required fix: narrow the grep's file set to the
*new* record-write path only (e.g. the new sibling file by name, plus `internal/kanban/`),
excluding `session_start.go`'s sidecar block; and re-state the baseline as the measured six hits
with the expected post-change count, or drop the grep half and let the two-identifier fixture — the
half that genuinely observes something — carry the criterion alone.

**D2.** `M2-CONSUMER-PREMISE-FALSE` — `plan.md` §A and §D (M2, "Ships alone") — **The plan asserts
there is no consumer of the record; there is one, and it reads every file in the directory.**
`plan.md` §A: "No consumer changes — `internal/web` and `internal/statusline` are both out of
scope". `plan.md` §D M2: "No consumer reads either yet, so a duplicate is inert — and shipping M2
alone is what makes M3 a removal rather than a swap." Measured:

```
$ grep -rn "internal/kanban\"" internal | grep -v _test
internal/web/viewmodel_ops.go:23:	"github.com/modu-ai/moai-adk/internal/kanban"
…
$ grep -rn "KanbanRecord" internal/web | grep -v _test
internal/web/viewmodel_ops.go:49:type KanbanRecord = kanban.Record
internal/web/viewmodel_ops.go:439:func loadKanbanRecords(root string) []KanbanRecord
internal/web/viewmodel_ops.go:227:func buildChain(records []KanbanRecord, sessions map[string]SessionVM, cardID string) ChainVM
```

`loadKanbanRecords` (`viewmodel_ops.go:439-461`) globs `.moai/state/kanban/*.json` and unmarshals
each into `kanban.Record`; `buildChain` (`:227-232`) then collapses the set into
`byRole := map[string]KanbanRecord{}` — **last write wins per role**, over `os.ReadDir` order, which
is lexicographic by session id. During M2's declared dual-writer interim each launch produces two
records carrying the **same role** (the launcher's, under the parent key; the session's, under its
own), so which one the console renders is decided by an arbitrary id comparison. M2 therefore does
not "ship alone" inertly: it can make the rendered chain non-deterministic where today it is merely
wrong. The claim is an unverified premise of exactly the shape `verification-claim-integrity.md`
§1.1 surface 4 names.
— Severity: **major** — Class: **blocking** — Required fix: re-measure the consumer set and correct
§A and §D; then either collapse M2+M3 into one milestone (write moves and launcher write is removed
in the same change, so no interim duplicate exists), or state explicitly why an arbitrary-winner
interim in `buildChain` is acceptable for the life of one milestone.

**D3.** `REQ-KRS-005-PRESCRIBES-A-GUESS` — `spec.md` §B.2 REQ-KRS-005, with `acceptance.md` §E
bullet 2 — **The requirement's own anti-guess clause is defeated by its derivation rule for any
session outside a card worktree.** REQ-KRS-005: "A record shall carry **the card identifier the
session is working** … otherwise from the basename of the session's worktree root. **When** neither
source yields a value, the field shall be left empty rather than guessed." The basename source
always yields a value for any session inside a git checkout, so the empty branch is unreachable
except when git itself fails. Measured — one of the three live sessions right now
(`2e3ace62…`, `active-sessions.json`) has `cwd: /Users/goos/moai/moai-adk-go`, a primary checkout,
not `.claude/worktrees/<card-id>`; its basename is `moai-adk-go`, which is not a card identifier.
`acceptance.md` §E blesses this — "The basename is recorded as-is; the field says what the session
was standing in" — but that is a *different field* from the one REQ-KRS-005 names, and the consumer
that will read it (`SPEC-WEB-CONSOLE-015` REQ-WC15-044: "the console shall present … the card
identifier") will render `moai-adk-go` as a card. The field's contract says card identifier; its
derivation delivers "whatever directory the session stood in". This is a guess dressed as a value,
and the case is live rather than hypothetical.
— Severity: **major** — Class: **blocking** — Required fix: either (a) constrain the derivation —
take the basename only when the worktree root's parent is `.claude/worktrees`, leave the field empty
otherwise, which makes the "rather than guessed" clause reachable and true; or (b) rename the field
and the requirement to what the derivation actually delivers (the worktree basename) and state in
§D that resolving it to a real card is the consumer's problem — but then reconcile that with
`SPEC-WEB-CONSOLE-015` REQ-WC15-044, which asks for a card identifier.

**D4.** `STALE-PARENT-SECTION-CITATION` — `spec.md` §A.3, `acceptance.md` AC-KRS-009, `plan.md` §I —
**Three citations point at `SPEC-WEB-CONSOLE-015` §A.5 for a claim that has moved and been
withdrawn, and describe it in the present tense.** `spec.md` §A.3 opens "`SPEC-WEB-CONSOLE-015` §A.5
**asserts** that … 'is a join that closes on today's data with no new state file'". Measured in the
parent:

```
$ grep -n '^### A\.' .moai/specs/SPEC-WEB-CONSOLE-015/spec.md
105:### A.4 The lane join — and the correction of what version 0.1.0 claimed about it
135:### A.5 Dependencies
$ sed -n '112,113p' .moai/specs/SPEC-WEB-CONSOLE-015/spec.md
Version 0.1.0 §A.5 asserted this "closes on today's data with no new state file". **That claim is
withdrawn.**
```

The parent's §A.5 is now "Dependencies"; the join claim lives at §A.4 and is explicitly withdrawn,
with the parent's own dependency bullet pointing at "§A.4". A reader following any of the three
citations lands in the wrong section and is told the parent asserts something it has retracted.
— Severity: **minor** — Class: **blocking** — Required fix: retarget all three citations to
`SPEC-WEB-CONSOLE-015` §A.4 and change "asserts" to the past tense with the withdrawal noted — the
technical conclusion (the join does not close) is unaffected and is independently correct.

**D5.** `MUTABLE-BASELINE-COUNTS` — `spec.md` §A.6 / §C.4 / §D, `acceptance.md` AC-KRS-004 and §F —
**"84 existing record files" does not reproduce, and a Definition-of-Done item is pinned to a count
that changes while the SPEC is open.** The figure appears five times (§A.6 "the 84 existing record
files", §C.4 "The 84 files under `.moai/state/kanban/`", §D "the 84 files already under", AC-KRS-004
"**0** of 84 files", §F "The 84 existing files … neither migrated nor deleted"). Measured:

```
$ ls .moai/state/kanban/*.json | wc -l
      81
$ grep -l '"session_id"' .moai/state/kanban/*.json | wc -l
      79
$ grep -L '"session_id"' .moai/state/kanban/*.json
backlog.json
leads.json
```

**79 record files**, not 84; the `*.json` glob the SPEC's own greps use additionally sweeps in
`backlog.json` (51 KB of queue state) and `leads.json`, which are not records. The directory is live:
`33468939-….json` was created at 13:43 today, *after* the SPEC was authored at 13:25. A DoD checkbox
reading "The 84 existing files … neither migrated nor deleted" is therefore not binary-checkable at
merge — the true count will differ again, and a reviewer cannot tell a drift from a violation.
The role histogram the SPEC cites *does* still reproduce (`grep -h '"role"' … | sort | uniq -c` → 34
`"role": "lane"`), and the lane-key absence reproduces under both regex dialects
(`grep -l '"lane"\s*:'` and `grep -l '"lane"[[:space:]]*:'` both → 0), so those two measurements
stand.
— Severity: **minor** — Class: **blocking** — Required fix: replace the fixed count with the
property that is actually stable and checkable — "no file under `.moai/state/kanban/` is migrated,
repaired, or deleted by this change" — and, where a count is quoted as evidence, attribute it to its
measurement instant and exclude `backlog.json` / `leads.json` from the record glob.

**D6.** `BACKEND-ONLY-CLAIM-OVERSTATED` — `spec.md` §A.5 and `acceptance.md` AC-KRS-006 baseline —
The claim that the backend "exists **only** as a literal argument at the eight `recordKanbanSession`
call sites", attributed to `grep -rn "BackendGLM\|BackendClaude" internal/ | grep -v _test`, is not
what that command returns. Measured: it returns the eight call sites **plus** `record.go:23,24,75`
(which AC-KRS-006 does acknowledge) **plus nine unacknowledged hits** in
`internal/cli/mcp_convergence.go` (`:61`, `:63`, `:90`, `:274`, `:380`, `:500`, `:519`, `:552`) — a
same-named, unrelated constant pair in the same package. The load-bearing half of the claim survives
(`grep -rn 'BACKEND' internal/config/envkeys.go` → zero lines, so no environment key names the
backend; verified), and AC-KRS-006's own criterion greps only `cc.go` and `glm.go`, which is
correctly scoped. Only the narrative "only" is wrong.
— Severity: **minor** — Class: **optional** — Required fix: qualify the claim to the record's
backend argument, or scope the cited command to `internal/cli/cc.go internal/cli/glm.go`.

**D7.** `AC-KRS-003-EVIDENCE-MISMATCH` — `acceptance.md` §A (AC-KRS-003) — The criterion's half (a)
greps `kanban.WriteBestEffort\|kanban.Write(`, but the displayed evidence block is the output of a
different command (`grep -rn 'recordKanbanSession(' internal/cli/`, nine hits). The prose then
correctly reconciles them ("Half (a) therefore goes from one hit to zero"), and both measurements
verify (`grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/cli/` → exactly one hit,
`kanban.go:478`; the nine-hit block reproduces verbatim). The defect is presentational: the evidence
shown is not the evidence for the criterion stated.
— Severity: **minor** — Class: **optional** — Required fix: show the output of the criterion's own
grep, and keep the nine-hit call-site listing as separate supporting context.

**D8.** `COMMENT-CITED-AS-CODE` — `acceptance.md` §B (AC-KRS-004 baseline) — "`grep -n 'Lane\|Card'
internal/kanban/record.go` returns only two hits, **both inside the role setter's known-set
check** (`record.go:119`, `:126`)". Measured: the grep returns exactly those two lines, but `:119`
is a doc-comment line ("// the kanban roles (lead + the three companions) plus RoleLane, which a")
and only `:126` is the check. The count is right; the characterisation of one hit is not. This is
the same defect shape the parent SPEC's iteration-2 audit recorded (a cited line number resolving to
a comment), in a mild form.
— Severity: **minor** — Class: **optional** — Required fix: describe `:119` as the setter's doc
comment.

**D9.** `REQ-KRS-006-PRESCRIBES-TRANSPORT` — `spec.md` §B.2 REQ-KRS-006 — The requirement names the
mechanism ("shall be conveyed to that session **through its launch environment**") rather than the
property (that the session can observe the fact at all). This is HOW inside a requirement body —
the defect shape the parent audit flagged. Mitigating: the transport *is* substantively the
deliverable here (§A.5 measures that no such key exists), the plan records the rejected alternative
("Deriving the backend from `ANTHROPIC_BASE_URL` … is a guess dressed as a measurement", §G), and
`internal/kanban/factory_slots.go:34-36` confirms the launcher `exec`s without forking so the
environment genuinely propagates. Reported for completeness, not as a correctness problem.
— Severity: **minor** — Class: **optional** — Required fix: none required; optionally restate as
"shall be observable by the launched session" with the environment named in §C as the chosen means.

**D10.** `GEARS-WHERE-SEMANTICS` — `spec.md` §B.2 REQ-KRS-004 — "**Where** a session is a factory
lane" uses the GEARS `Where` pattern, whose canonical sense is a capability gate / feature flag /
static configuration. Being a factory lane is a launch-time property of a session, closer to
`While` (state-driven). Syntactically valid GEARS; semantically a stretch. Did not affect the MP-2
verdict.
— Severity: **minor** — Class: **optional** — Required fix: none required; `While` would be a
closer fit.

**D11.** `AC-KRS-008-PASSES-TODAY` — `acceptance.md` §C (AC-KRS-008) — The criterion is satisfied on
the untouched tree, which `acceptance.md`'s own preamble declares a defect ("A criterion that
already passes on the untouched tree is a defect, not a criterion"). The SPEC discloses this
explicitly and justifies it as a preservation guard across the move, which is the correct handling —
and post-change the criterion is non-vacuous, because the write it guards will exist. Verified that
the fail-open property it preserves is real: `record.go:174` `func WriteBestEffort(projectRoot
string, rec *Record) { _ = Write(...) }` under the `@MX:NOTE` "the absent error return is the
fail-open guarantee, not an oversight". Recorded as a disclosed, justified exception rather than a
finding to route.
— Severity: **minor** — Class: **optional** — Required fix: none.

---

## Items Verified Clean (no defect)

These were audited specifically and found sound; recording them so the Gaps section below is
honest about what *was* observed.

1. **Schema additivity (brief item 3).** The `@MX:ANCHOR` at `record.go:45` reproduces verbatim,
   including its reason ("both sides plus the sync-phase dedup gate bind to these JSON keys, so a
   renamed key breaks readers this package cannot see"). REQ-KRS-007 covers all three obligations —
   additive, `omitempty`, no rename — and adds the pre-change-bytes read and the
   rewrite-preserves-keys property. AC-KRS-007(a) constrains its fixture to writer-produced bytes
   ("not hand-authored"), which is the right call: `Write` uses `json.MarshalIndent(rec, "", "  ")`
   (`record.go:157`) so a hand-indented fixture would fail a correct implementation. The existing
   `Role string \`json:"role,omitempty"\`` field is precedent, and its doc comment already states
   the property ("omitempty keeps pre-existing records byte-identical on rewrite"). Note correctly
   scoped: the four non-`omitempty` existing keys (`spec_id`, `backend`, `deepscan_dir`,
   `verify_reentries`) are untouched, and REQ-KRS-007 binds only "every field added".
2. **The lane number (brief item 4).** REQ-KRS-004 makes it data distinct from the role, and
   `plan.md` §D M1 plus §G forbid routing it through `WithRole`. The hazard is real and verified:
   `record.go:126` `if role == RoleLead || role == RoleLane || isCompanionRole(role)` silently drops
   anything else, so a `lane-3` string would vanish. The zero-value justification is **measured, not
   asserted** — `SplitFactoryLaneLabel` (`bootstrap.go:271-281`) returns `0, false` when
   `err != nil || n < 1`, so lanes genuinely number from 1 and 0 is unreachable by legitimate data.
   `internal/cli/factory.go:75` (`WorkerNumber int // n of -f lane-<n>; 0 unless the lane form`) is
   existing precedent for the same convention. The "N lanes are indistinguishable today" baseline is
   measured (34 `"role": "lane"` rows, zero lane-number keys).
3. **The card identifier vs `MOAI_KANBAN_ID` (brief item 5).** No conflation. `envkeys.go:167-173`
   verifies as cited: `EnvMoaiKanbanID` "carries the **run** identifier that distinguishes one
   kanban run from another"; both `spec.md` §A.4 and `plan.md` §F name it as explicitly *not* the
   source. (The derivation defect is D3, which is a different problem.)
4. **The t221 boundary (brief item 6).** No requirement reaches into the sidecar. REQ-KRS-001 only
   forbids keying from it; §D's exclusion states the boundary and the reason; `plan.md` §F and §G
   both restate it. The one place the boundary leaks is AC-KRS-002's grep half — D1.
5. **Two writers at the terminal state (brief item 7).** Correctly excluded: `plan.md` §F rejects
   "keeping the launcher as the writer 'for the fields it already knows'" with the reason, §G lists
   it as an anti-pattern, and AC-KRS-003 pins it with both a grep and a file-count fixture. The
   defect is the *interim* only — D2.
6. **Reachability of the launch facts (brief item 1).** Every row of §A.5 re-verified at its cited
   line: `MOAI_KANBAN` set `kanban.go:171` / read `session_start_kanban.go:53`; `MOAI_KANBAN_LABEL`
   set `:315` / read `:50`; `MOAI_FACTORY_WORKERS` set `factory.go:253` / read
   `session_start_factory.go:51`; `MOAI_FACTORY_WORKER` set `factory.go:289` / read `:48`;
   `MOAI_KANBAN_SPEC` set only inside `enterKanbanMode` at `kanban.go:174-176` — confirmed absent
   from `enterKanbanCompanionMode` (`:305-322`) and `enterFactoryWorkerMode` (`factory.go:283-298`),
   both re-read in full. Both "No" rows verified: `grep -rn 'BACKEND' internal/config/envkeys.go`
   and `grep -rn 'CARD' internal/config/envkeys.go` each return zero lines. REQ-KRS-006 covers
   exactly the three missing facts; nothing missing is assumed present.
7. **Sibling boundaries.** `SPEC-WEB-CONSOLE-015` is consumer-only on this surface: its
   REQ-WC15-043/044/047 present and join, and its `design.md:13,18` delegate G-1 (card derivation)
   and G-6 (lane-number type) to this SPEC by name — which `plan.md` §F restates as inherited. Its
   §D carries `### Out of Scope — kanban record keying and lane identity (owner:
   SPEC-KANBAN-RECORD-SESSION-KEY-001)`. `SPEC-SESSION-TELEMETRY-001` §A.3 independently declines to
   read the sidecar for the same reason. **Nothing owned here is unclaimed; nothing handed over is
   duplicated.** The one genuine overlap risk — the card identifier's meaning — is D3.
8. **Budget (brief item 9).** 8 requirements, 9 criteria, against Tier M ceilings of 16 and 16
   applied independently. Both well inside. `plan.md` §B states the same numbers and adds the right
   discipline ("If scope grows past that, the answer is to cut scope, not to raise the ceiling").

---

## Gaps — what was NOT observed

- **No Go test, build, or vet was run.** Per the audit brief ("Do not run the full Go test suite.
  Targeted greps and file reads only"), all code claims rest on reading source at cited lines. In
  particular, D2's duplicate-record consequence is derived from reading `buildChain`'s `byRole`
  last-wins map and `loadKanbanRecords`'s unfiltered `*.json` glob — it was **not** exercised by
  constructing two same-role records and rendering the console. The map's last-wins behaviour is
  plain in the source (`viewmodel_ops.go:228-232`); the *rendered* consequence is inferred.
- ~~**`backlog.json` / `leads.json` behaviour inside `loadKanbanRecords` was not tested.**~~
  **Closed after re-measurement.** Both are JSON *objects* with keys `KanbanRecord` does not declare
  (`{"version":1,"last_seq":226,"items":[…]}` and `{"lead":{"pid":58131,…}}`), so
  `json.Unmarshal` **succeeds** — Go ignores unknown fields — and each yields a zero-valued
  `KanbanRecord`. They therefore survive into `loadKanbanRecords`' return slice but are dropped by
  `buildChain`'s `if role := roleOf(r); role != ""` guard (`viewmodel_ops.go:229`) and by
  `chainCardID`'s `if r.SpecID != ""` (`:657`). Net effect on the console: they inflate
  `len(records)`, which only feeds `ChainVM.Present` (`:233`). Not a defect of this SPEC, and it
  changes no finding — but it does confirm D5's point that the `*.json` glob is not a record glob.
- **The `moai web` console was not started.** Whether misattributed rows render today is unobserved
  — the same gap `plan.md` §H records honestly for itself.
- **No factory run was launched.** AC-KRS-009's end-to-end join (`workers.json[lane-N].PID →
  active-sessions → session_id → record`) was verified structurally (both PID fields exist at
  `factory_slots.go:38` and `registry.go:92`, both re-read) and by the absence of records for
  registered sessions, but not by executing a factory run.
- **`.moai/state/active-sessions.json` liveness was not probed.** The three entries were read as
  registry data; `kill -0` was not run against pids 51045 / 34699 / 83078, so "three live sessions"
  is a registry claim, not a process observation. The premise finding does not depend on it — the
  record/registry key mismatch and the seven-minute timestamp gap hold regardless of whether the
  processes are still running.
- **Cross-model audit not invoked.** No `audit_multi` / codex / GLM second opinion was requested for
  this audit; the verdict is single-auditor.
- **`git rev-parse --show-toplevel` from the primary checkout was not run** (the worktree guard
  refuses cross-tree git). D3's `moai-adk-go` basename is derived from the recorded
  `cwd: /Users/goos/moai/moai-adk-go` in the session registry plus path arithmetic, not from
  executing the command there. In this worktree the command was run and returns
  `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`, basename `t207`, confirming the
  card-worktree half of the derivation.

Every checklist group (Groups 1-8) was executed; no criterion was left unevaluated. The audit is
complete — the gaps above are bounded to the five open items named (the sixth was closed on
re-measurement), and none of them is load-bearing for the verdict.

---

## Residual Risk

- **The record directory is live state and moved during this audit.** A file appeared at 13:43
  (`33468939-….json`) while the SPEC was being audited. Any count, any session-id citation, and any
  "no record exists for S" observation in this report — including the ones supporting the premise —
  is true of an instant, not of a steady state. This is the same hazard D5 names in the SPEC.
- **D2's severity depends on whether M2 actually ships alone.** If the run phase collapses M2 and M3
  into one commit, the dual-writer interim never exists on any tree a consumer reads and the finding
  reduces to a documentation correction. If M2 ships as its own release, the console's chain view is
  non-deterministic until M3 lands.
- **D3 may be a wording defect rather than a design defect.** If the intent was always "the worktree
  basename, whatever it is", then the requirement and the parent's REQ-WC15-044 need reconciling but
  no code changes. If the intent was a real card identifier, the derivation needs the containment
  check. The two readings produce different implementations, which is exactly why it scores against
  Clarity.
- **The premise is verified for the launch shapes measured; one shape was not exercised.** A `moai
  glm` factory lane's end-to-end environment propagation is recorded as an open item in `plan.md`
  §H rather than measured. The launcher's `exec`-without-fork property (`factory_slots.go:34-36`)
  makes propagation very likely, but likely is not observed.
- **This audit read the code at `ee039da30`; the SPEC measured at `dfbf828a6`.** Every citation
  checked here still resolves, so no drift was detected between those commits on the cited lines —
  but only the cited lines were checked, not the whole diff.

---

## Recommendation

FAIL at 0.750 against a 0.80 threshold, with five blocking findings. This is a strong SPEC with a
verified premise, honest evidence discipline, clean traceability, and correctly held sibling
boundaries — it fails on three verification statements that do not verify what they claim and two
claims that measurement contradicts. All five are cheap to fix and none requires re-scoping.

In priority order:

1. **D1 — repair AC-KRS-002.** Narrow the grep's file set away from `session_start.go`'s sidecar
   block, or drop the grep half entirely. As written the criterion cannot pass without violating
   §D's t221 exclusion and `acceptance.md` §F. This is the one finding that would send an
   implementer to break the SPEC's own boundary.
2. **D2 — re-measure the consumer set and correct `plan.md` §A and §D M2.**
   `internal/web/viewmodel_ops.go:439` reads every record file and `:228` collapses them by role,
   last-wins. Then decide explicitly: collapse M2+M3, or accept and document the interim.
3. **D3 — decide what the card field means.** Either constrain the derivation to
   `.claude/worktrees/<basename>` so the "rather than guessed" clause becomes reachable, or rename
   the field to the worktree basename and reconcile with `SPEC-WEB-CONSOLE-015` REQ-WC15-044.
4. **D4 — retarget the three `SPEC-WEB-CONSOLE-015` §A.5 citations to §A.4** and change "asserts" to
   the past tense with the withdrawal noted.
5. **D5 — replace the fixed "84" with the checkable property** ("no file … is migrated, repaired, or
   deleted"), and exclude `backlog.json` / `leads.json` from any record glob used as evidence.

The six optional findings (D6-D11) are surfaced for the orchestrator's discretion and are not
routed: fixing them would improve the SPEC's evidence precision but none affects correctness, and
per M6 a list of optional findings does not by itself justify a FAIL and was not used to manufacture
one — the harmonic mean and the five blocking findings carry the verdict on their own.

One thing the re-audit should carry forward: **the premise now has better evidence than the SPEC
cites.** The `33468939…` / `2e3ace62…` pair — a record whose `entered_at` precedes its supposed
subject's `started_at` by seven minutes and follows the actual child's registration by three seconds
— demonstrates the launcher-before-child ordering with timestamps, which the original three-session
citation did not. AC-KRS-001's baseline would be stronger stated that way, and `acceptance.md` §F
already requires re-measurement at merge.
