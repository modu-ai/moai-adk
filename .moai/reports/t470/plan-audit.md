# Plan Audit — SPEC-QUEUE-UPGRADE-PROOF-001 (card t470)

Auditor: `plan-auditor` · Iteration 1/2 (Tier M ceiling) · Date 2026-09-03

Tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t470`, branch
`WT-queue-upgrade-proof`, HEAD `38d1526b1`, SPEC base `4e4607abe`
(`git merge-base --is-ancestor 4e4607abe HEAD` → yes).

Reasoning context ignored per M1 Context Isolation. The dispatch's scope
discipline (proof-not-repair, `REQ-QUP-009`) was consumed as an audit
CONSTRAINT — it is what this audit grades against — not as evidence about the
artifacts.

---

## Verdict

| | |
|---|---|
| **Verdict** | **FAIL** |
| Aggregate quality score | **0.875** (Tier M threshold 0.80 — score is ABOVE threshold) |
| Cause of FAIL | Two must-pass firewall failures (MP-3, MP-7). A must-pass failure is not compensable by score. |
| Blocking defects | D1, D2, D3, D4, D5, D6 |
| Optional defects | D7, D8, D9, D10 |

The score and the verdict disagree deliberately, and both numbers are reported
because they carry different information. On substance this is a strong SPEC:
every measured claim in it re-verified exactly, including line-level ones. What
fails it is (a) a frontmatter defect that renders the SPEC mechanically
unlintable, and (b) the open clarification gate, which is score-independent by
construction. D1/D2 are one-line edits; D3/D4 are two acceptance-criterion
rewrites. None requires re-authoring the SPEC.

---

## Must-Pass Results

| ID | Criterion | State | Evidence |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | `grep -n '^### REQ-' spec.md` → `REQ-QUP-001`(L81) `002`(L88) `003`(L95) `004`(L101) `005`(L106) `006`(L112) `007`(L119) `008`(L126) `009`(L131). Sequential, no gap, no duplicate, uniform 3-digit padding. |
| MP-2 | GEARS format compliance | **PASS** | Judged against the **requirement layer** (`REQ-XXX` in `spec.md`) only; ACs graded under Group 4. 001-004 event-driven (`**When** … shall`); 005/006/008 ubiquitous (`The … shall`); 007 `**Where** … shall`; 009 unwanted (`shall not change`). 9/9 match a GEARS pattern. |
| MP-3 | YAML frontmatter validity | **FAIL** | Two independent violations — see D1 (`tags:` sequence, decode error) and D2 (`lifecycle: spec-first` outside the enum). Measured below. |
| MP-4 | §22 language neutrality | **N/A (auto-pass)** | Single-language scope: `module: internal/kanban`, a Go package in this repository. No multi-language tooling surface named. |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md plan.md acceptance.md \| sort -u` → `SPEC-QUEUE-UPGRADE-PROOF-001` only (self). No external SPEC referenced, so no retired/superseded/archived status to reconcile. No BLOCKING finding. |
| MP-6 | D8 cross-platform discipline | **PASS (auto)** | `grep -c 'syscall' spec.md plan.md acceptance.md` → `0`, `0`, `0`. D8-4 auto-PASS. |
| MP-7 | Clarification gate | **FAIL** | `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/` → 4 matches: `plan.md:10`, `plan.md:22`, `progress.md:13`, `progress.md:14`. See D5. |

### MP-3 measurement

`internal/spec/lint.go:497-503` declares `SPECFrontmatter.Tags` as `string`
(`yaml:"tags"`). The SPEC writes it as a YAML sequence (`spec.md:13`).

Probe (standalone module, `gopkg.in/yaml.v3@v3.0.1`, decoding the SPEC's own
frontmatter block extracted with `sed -n '2,14p' spec.md`):

```
err = yaml: unmarshal errors:
  line 12: cannot unmarshal !!seq into string
fm  = {Version:0.1.0 Priority:HIGH Lifecycle:spec-first Tags:}
```

Blast radius, read from `internal/spec/lint.go:619-623`:

```go
fm, body, err := extractFrontmatter(content)
if err != nil {
    doc.ParseError = fmt.Errorf("frontmatter parsing error: %w", err)
    return doc
}
```

`parseSPECDoc` returns EARLY. The SPEC therefore reaches **no lint rule at
all** — not `FrontmatterSchemaRule`, not the GEARS modality check, not the
coverage or out-of-scope rules. It is not "a SPEC with a warning"; it is a SPEC
invisible to the audit tooling. That is why this is critical rather than
cosmetic.

Corpus position (`grep -rl '^tags: \[' .moai/specs/*/spec.md | wc -l` → `4` of
`746`): the sequence form is a 4-SPEC outlier, not a house style.

`lifecycle: spec-first` is outside the SSOT enum
(`spec-anchored | spec-lite | exploratory`,
`.claude/rules/moai/development/spec-frontmatter-schema.md` § Field Reference).
Corpus: `grep -rh '^lifecycle:' .moai/specs/*/spec.md | sort | uniq -c` →
`651 spec-anchored`, `97 completed`, `3 spec-first`, `1 exploratory`,
`1 design-only`. The lint rule does not enum-check this field, so the defect is
schema-conformance only — but MP-3 binds the SSOT schema, not the lint subset.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.95 | 1.0 | Every requirement has one reading. No pronoun ambiguity. Normative text avoids "should"/"reasonable" throughout. Deducted for D9 (one imprecise line citation) and a mild pattern-selection quibble on `REQ-QUP-007` (`Where` is GEARS' capability-gate/feature-flag form; the condition here is a data shape, closer to `While`). Neither changes what an implementer would build. |
| Completeness | 0.85 | 0.75-1.0 | All required sections present (`spec.md` HISTORY L17, §A Context L26, §B L56, §C Requirements L79, §D Constraints L137, §E Exclusions L149; `acceptance.md` §A matrix, §B scenarios, §C quality gate, §D DoD). Five `### Out of Scope — <topic>` H3s (L155/162/174/184/192), each with specific `-` bullets AND a stated reason. All 12 frontmatter fields are PRESENT; deducted because two carry invalid values (D1, D2). |
| Testability | 0.80 | 0.75 | Most ACs are binary and command-checkable, and two are notably well built (see below). Deducted for D3 (the RED-establishing mutation does not establish RED) and D4 (one AC limb is unconditionally true). |
| Traceability | 0.90 | 0.75-1.0 | Every `REQ-QUP-001..009` carries ≥1 AC (001→001a+001b; 002-009 one each) — no uncovered REQ. Every AC except one names a valid REQ. Deducted for D6: `AC-QUP-010` traces to `—` (acceptance.md:23), an orphan by the document's own matrix. |

Aggregate (arithmetic mean): **0.875**. Tier M PASS threshold 0.80.

---

## Defects Found

**D1** — `spec.md:L13` — `tags: [queue, upgrade, migration, testing, kanban]` is
a YAML sequence decoded into a `string` field, which errors the whole
frontmatter decode and makes `parseSPECDoc` return before any lint rule runs.
The SPEC is mechanically unlintable and unauditable by `moai spec audit`.
— Severity: **critical** — Class: **blocking**
— Required fix: replace with the comma-separated string form,
`tags: "queue, upgrade, migration, testing, kanban"`.

**D2** — `spec.md:L12` — `lifecycle: spec-first` is not a member of the SSOT
enum `spec-anchored | spec-lite | exploratory`.
— Severity: **critical** (MP-3 binding) — Class: **blocking**
— Required fix: `lifecycle: spec-anchored`.

**D3** — `acceptance.md:L110-119` (`AC-QUP-010`) — **the RED-establishing
mutation does not establish RED.** This is the card's single anti-vacuity
guard, and as written it is itself vacuous.

The prescribed mutation is "seeding the fixture under the CURRENT directory
name instead of the legacy one, so no relocation is required", with the
expectation that `AC-QUP-002` then fails. Trace it against the production code:

- `internal/kanban/state_dir.go:81-84` — `resolveStateDir` returns the current
  directory unconditionally when it exists; no relocation is attempted and none
  is needed.
- `internal/kanban/backlog_migrate.go:434` — `migrateLegacyBacklog(queuePath)`
  is keyed on the RESOLVED queue path, not on the directory's name. A
  `backlog.json` sitting in `.moai/state/todo/` migrates exactly as one in
  `.moai/state/kanban/` does.
- `internal/kanban/backlog_migrate.go:558-567` — `quarantineLegacyBacklog`
  renames it to `.migrated` on the same terms.

`AC-QUP-002`'s assertion pair is "`<root>/.moai/state/todo/` exists and holds
the queue, **and** `<root>/.moai/state/kanban/` no longer exists". Under the
mutation the legacy directory is never created, so the second limb — a negative
existence check — passes **vacuously**, and the first limb passes because the
fixture was seeded there. `AC-QUP-003` and `AC-QUP-004` likewise still pass,
because migration and quarantine both still run. The mutation produces GREEN on
every criterion, so it demonstrates nothing about what the test asserts.
— Severity: **major** — Class: **blocking**
— Required fix: EITHER (a) make `AC-QUP-002` assert relocation POSITIVELY — seed
a sentinel file inside the legacy directory (a session registry file is the
natural one, since `state_dir.go:12-16` documents that the relocation moves the
directory rather than a file list) and assert that sentinel appears under
`.moai/state/todo/` afterwards; OR (b) choose a mutation that genuinely breaks
the composed path, e.g. swapping the non-adopting `kanban.BacklogPathForRoot`
for `BacklogPathForRootAdopting` at the test's entry, which suppresses the
relocation while leaving the legacy fixture in place. Option (b) must not touch
production code — it is a test-local entry substitution.

**D4** — `acceptance.md:L99-100` (`AC-QUP-008`, third limb) — the live-queue
protection is verified by "`git status --short .moai/state/` reporting nothing
for that path after the run". Measured: `.moai/state/` is gitignored
(`git check-ignore -v .moai/state/todo/backlog.db` → `.gitignore:311
.moai/state/`), so that command reports nothing whether or not the live queue
was mutated. The check cannot fail, so it asserts nothing.
— Severity: **major** — Class: **blocking**
— Required fix: replace with a before/after digest comparison of the real file,
e.g. record `shasum -a 256 .moai/state/todo/backlog.db` (and its mtime) before
the test run and assert equality after. The other two limbs of `AC-QUP-008`
(all paths under `t.TempDir()`; resolved queue root asserted inside the temp
dir before any command runs) are sound and close the root-resolution trap — keep
them unchanged.

**D5** — `plan.md:L10`, `plan.md:L22` (echoed at `progress.md:L13-14`) — two
unresolved `[NEEDS CLARIFICATION]` markers: `G2 definition` and `downgrade
intent vs quarantine rename`. MP-7 makes any unresolved marker a must-pass
failure at audit time.
— Severity: **critical** (gate) — Class: **blocking**
— Required fix: the ORCHESTRATOR resolves each topic with the dispatcher via
`AskUserQuestion` before Implementation Kickoff Approval. **No SPEC edit is
required and none should be made by the author** — this audit explicitly did
not resolve them. On the quality of the markers themselves, both are
well-formed: each states the two mechanical facts that conflict, names the
judgment being deferred and to whom, records the consequence of arriving
unresolved (`plan.md:L18-20` — G1 ships, G2 closes as unstarted rather than
silently dropped), and correctly proposes no change on its own account. A third
party could act on either without re-deriving it. That is the correct handling
of an open question; it is the open state, not the writing, that fails MP-7.

**D6** — `acceptance.md:L23` — `AC-QUP-010`'s Requirement column is `—`. It is
the only AC with no backing requirement, so the card's anti-vacuity guard is
unanchored in the requirement layer.
— Severity: **minor** — Class: **blocking** (traceability is a criterion this
document states)
— Required fix: EITHER map it to `REQ-QUP-005` (the regression-guard
requirement, which a vacuous test would not satisfy), OR add a short
`REQ-QUP-010` in the unwanted form — "The composed-path proof shall not be
accepted on a GREEN it has never been shown capable of failing."

**D7** — `spec.md:L9` — `priority: HIGH`. The SSOT enum is
`P0|P1|P2|P3` or `High|Medium|Low|Critical`; the uppercase spelling matches
neither as written. Not lint-enforced.
— Severity: minor — Class: **optional**
— Suggested fix: `priority: High` (or `P1`).

**D8** — `spec.md:L3`, `spec.md:L4` — `version: 0.1.0` and the `title:` value
are unquoted; the schema specifies both as quoted strings. YAML still decodes
both as strings, so nothing breaks today.
— Severity: minor — Class: **optional**
— Suggested fix: quote both.

**D9** — `plan.md:L106` — cites `todo_queue_root_test.go:100` as the
`initGitRepo(t, primary)` + `t.Setenv("CLAUDE_PROJECT_DIR", primary)` pattern.
Read at that line: the test is
`TestResolveTodoQueueRoot_SubdirectoryResolvesToRepoRoot`, which does call
`initGitRepo(t, primary)` but sets `CLAUDE_PROJECT_DIR` to `sub`, not
`primary`. The pattern the plan describes does exist in the file; the citation
picks a line where one half of it differs.
— Severity: minor — Class: **optional**
— Suggested fix: cite the line whose `Setenv` target is the repo root, or note
the subdirectory variance.

**D10** — `acceptance.md:L105` — `AC-QUP-009` says
`git diff --stat <base>..HEAD` with `<base>` left as a placeholder, while the
pinned SHA `4e4607abe` is already stated in `plan.md:L3` and `progress.md:L3`.
— Severity: minor — Class: **optional**
— Suggested fix: inline the pinned SHA. (Note the SPEC is already correct on
the harder version of this: it pins a SHA rather than citing a moving ref, so
`verification-claim-integrity.md` §2.1 is satisfied.)

---

## What re-verified exactly — evidence for the PASS dimensions

Every quantitative claim in `spec.md §A` and `plan.md §B/§C` was re-measured in
this tree. All matched, several to the line number:

| SPEC claim | Command | Observed |
|---|---|---|
| No release tag carries the SQLite merge | `git tag --contains 3cb258d62 \| wc -l` | `0` |
| v3.1.2 `BacklogRecord` has exactly 3 fields, at `backlog_store.go:75-78` | `git show v3.1.2:internal/kanban/backlog_store.go \| grep -A6 'type BacklogRecord struct'` | L75-79: `Version`, `LastSeq`, `Items`. No `Findings`, no `Archived`. |
| develop adds the two fields at `backlog_store.go:187-192` | `grep -A10 'type BacklogRecord struct' internal/kanban/backlog_store.go` | L187-193: adds `Findings`, `Archived` |
| `backlogVersion` is `1` on both sides | `grep 'backlogVersion = '` at `v3.1.2` and HEAD | v3.1.2 L47 `= 1`; HEAD L48 `= 1` |
| v3.1.2 carries `queued`/`picked`/`dropped` | `git show v3.1.2:… \| grep 'BacklogState = "'` | L54/56/58 — all three present |
| `moai update` cannot touch the queue | `grep -rn "internal/kanban" internal/cli/update/ \| wc -l` | `0` |
| `moai doctor` says nothing about the queue | `grep -rln "kanban\|backlog" internal/cli/doctor*.go \| wc -l` | `0` |
| No CLI test seeds the legacy directory | `grep -rn "LegacyStateDirForRoot" internal/cli/ \| wc -l` | `0` |
| `todoBacklogPath` is the adopting entrance | `sed -n '50,62p' internal/cli/todo.go` | L56-58 `todoBacklogPath` → `kanban.BacklogPathForRootAdopting(root)` |
| `resolveTodoQueueRoot` uses the adopting root resolver | `grep -n -A8 'func resolveTodoQueueRoot' internal/cli/todo.go` | L71-73 → `ResolveTodoQueueRootAdopting(resolveProjectDir())` |
| Per-layer test counts 5 / 10 | `grep -c '^func Test' internal/kanban/state_dir_test.go internal/kanban/backlog_migrate_test.go` | `5` / `10` |
| `TestConcurrencyStress` is in-process, single store | `sed -n '130,140p' internal/kanban/backlog_concurrency_test.go` | L135 `func TestConcurrencyStress`, L138 `store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))` — one store object, goroutines. The SPEC's correction of the common framing is right. |
| `backlogMigratedSuffix` at `backlog_migrate.go:41` | `sed -n '38,42p' internal/kanban/backlog_migrate.go` | L41 `const backlogMigratedSuffix = ".migrated"` |
| Quarantine is a pure rename (so `AC-QUP-003`'s byte-identity assertion is sound) | `sed -n '558,567p' internal/kanban/backlog_migrate.go` | `os.Rename(queuePath, target)`; existing `.migrated` never overwritten |
| Relocation is a pure rename, never copy-then-delete | `sed -n '108,120p' internal/kanban/state_dir.go` | `relocateStateDir` → `os.Rename(from, to)` |
| Only SPEC artifacts changed so far | `git diff --stat 4e4607abe..HEAD` | 4 files, all under `.moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/`, +559 |

Two acceptance criteria deserve positive credit rather than silence:

- **`AC-QUP-001b` is non-vacuous by construction.** It seeds `last_seq`
  strictly above the highest item id and asserts the next issued id is
  `last_seq + 1`. `internal/kanban/backlog_store.go:772-773`
  (`normalizeBacklogRecord`) raises `LastSeq` to the max present seq only when
  it is LOWER, so the seeded high-water mark survives untouched, and
  `backlog_store.go:695-697` issues `t{LastSeq+1}`. The criterion therefore
  distinguishes "the high-water mark crossed the composition" from "it was
  re-derived from the items" — which is exactly what it claims to do.
- **`AC-QUP-005` pre-empts the zero-match selector**: "a zero-match selector is
  a failure of this criterion, not a pass". `go test -run <name>` exits 0 on
  zero matches, so this clause is load-bearing and correctly stated.

---

## Production-code observations — deliberately NOT counted against the score

Recorded separately per the dispatch's inversion of the usual reflex, and
stated as candidates for a DIFFERENT card. **None of these depresses any
dimension score, and this audit recommends no production change under t470.**
`REQ-QUP-009` is correct and this audit endorses it.

1. **The refused-relocation silence.** When `relocateStateDir` fails
   (`state_dir.go:99-101`), `resolveStateDir` returns the legacy directory and
   the user keeps running READ-ONLY on the old layout with no diagnostic on any
   surface — `moai doctor` included (measured: 0 references). The SPEC already
   records this at `spec.md:L179-182` as the one situation where the G5 silence
   bites. I found nothing to add and nothing to correct. Not a t470 defect.
2. **The downgrade-comment tension.** `state_dir.go:141-145` asserts the queue
   document keeps the name `backlog.json` so "an older binary reads only this
   file"; `backlog_migrate.go:558-567` renames it to `.migrated` and the
   directory has moved. Read literally the two cannot both hold. The SPEC
   escalates this rather than resolving it (`plan.md:L22-45`), which is the
   correct handling — I did not resolve it either, and it is the dispatcher's
   judgment. Not a t470 defect.
3. **No new production defect was found.** I looked, at the surfaces the SPEC
   names: the fail-closed migration ordering (parity asserted BEFORE authority
   flips, `backlog_migrate.go:459-468`), the state-C double-check
   (`backlog_migrate.go:435-438`), the non-overwriting quarantine, and the
   `internal/cli/update` isolation. Each behaved as the SPEC describes. The
   dispatcher's premise — that the mechanism is sound — held under every check I
   ran. I am withholding no production finding.

---

## Evidence-bearing summary

**Claim.** SPEC-QUEUE-UPGRADE-PROOF-001 scores 0.875 against a Tier M threshold
of 0.80 and fails two must-pass criteria (MP-3, MP-7), yielding FAIL. Six
blocking defects, four optional. No production change is recommended.

**Evidence.** The command/output pairs are inline above — § MP-3 measurement
(yaml.v3 probe + `lint.go:619-623`), § What re-verified exactly (16 rows), and
the per-defect measurements in D3 (three source-line traces) and D4
(`git check-ignore`).

**Baseline-attribution.** All measurements taken in this run, in the worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t470` at HEAD `38d1526b1`,
against SPEC base `4e4607abe`. Tag-relative facts read through
`git show v3.1.2:…` in the same tree. No figure is carried over from another
tree, package, or point in time.

**Gaps — explicitly NOT observed.**
- **No test was executed.** `go test ./internal/kanban/... ./internal/cli/` was
  not run: `./internal/cli/` exceeds 600s on this machine (C-3) and the audit is
  read-only. All D3/D4 conclusions rest on source reading, not on execution.
  D3 in particular is a static trace of three call sites; it predicts what the
  mutation would do and has not been observed doing it.
- The yaml.v3 probe reproduced the decode in a standalone module at
  `gopkg.in/yaml.v3@v3.0.1`, NOT through `internal/spec`'s own build. The repo's
  yaml version was not pinned-checked. `parseSPECDoc`'s early return was read,
  not executed.
- `moai spec audit` / `mcp__moai__spec_audit` was not invoked against this
  SPEC — which is itself moot, since D1 means the tool could not have parsed it.
- Iteration-2 regression check: not applicable (first iteration).
- No cross-model backend was consulted (`audit_multi` not invoked).

**Residual-risk.**
- D3's remedy option (b) — substituting the non-adopting path — is my
  suggestion, unverified by execution; the run phase may find option (a)
  cleaner. The DEFECT (the named mutation does not produce RED) is
  independently established by the three cited call sites and does not depend on
  which remedy is chosen.
- Fixing D1 makes this SPEC lintable for the first time. Lint rules that have
  never run against it may then surface findings this audit did not anticipate,
  because no rule has ever seen the document. Re-run `moai spec audit` after
  D1/D2 land.
- The score 0.875 was computed as an unweighted mean of four dimensions; a
  harmonic mean would sit slightly lower (0.872) and would not change the
  verdict, which is firewall-driven rather than score-driven.

---

## Recommendation

Route in this order. Every item is a bounded edit; none re-opens the SPEC's
design.

1. **`spec.md:L13`** — `tags: "queue, upgrade, migration, testing, kanban"` (D1).
2. **`spec.md:L12`** — `lifecycle: spec-anchored` (D2).
3. **`acceptance.md:L110-119`** — rewrite `AC-QUP-010`'s mutation so it actually
   produces RED, per D3. This is the highest-value fix in the list: it is the
   criterion protecting the card's entire value.
4. **`acceptance.md:L99-100`** — replace the gitignored `git status` limb of
   `AC-QUP-008` with a digest comparison (D4). Leave the other two limbs.
5. **`acceptance.md:L23`** — give `AC-QUP-010` a requirement (D6).
6. **ORCHESTRATOR, not the author** — resolve both `[NEEDS CLARIFICATION]`
   topics with the dispatcher via `AskUserQuestion` before Implementation
   Kickoff Approval (D5). The markers stay in `plan.md` until answered; do not
   delete them to clear the gate.
7. Optional, at the orchestrator's discretion: D7, D8, D9, D10.

After 1-6, re-audit is scoped to that enumerated defect delta plus a regression
check, not a from-scratch audit (Tier M ceiling: 2 iterations).

---

## Operational Notes (unverified)

- `measured` — the sequence-form `tags:` is a 4-of-746 outlier in this corpus
  (`grep -rl '^tags: \[' .moai/specs/*/spec.md | wc -l` → `4`;
  `grep -rl '^tags:' .moai/specs/*/spec.md | wc -l` → `746`). The other three
  are presumably unlintable for the same reason. Measure whether that matters:
  `for f in $(grep -rl '^tags: \[' .moai/specs/*/spec.md); do echo "$f"; done`
  and run `moai spec audit` against each.
- `inferred` — `lifecycle:` accepts values outside the SSOT enum silently
  because `FrontmatterSchemaRule` (`internal/spec/lint.go:978-1046`) checks
  presence, the id pattern, semver shape, and the phase token set, but performs
  no enum membership test for `lifecycle`, `priority`, or `status`. The rule
  body was read; the consequence (that `design-only` and `completed` also pass
  unflagged) is inference from that reading, not an observed run. Measure with a
  fixture SPEC carrying `lifecycle: nonsense` before treating it as fact.
- `inferred` — a lint rule that enum-checks `lifecycle`/`priority` would flag a
  non-trivial slice of the existing corpus (97 `completed`, 25 lowercase
  `high`, 34 `"P2 Medium"`), so adding one is a corpus-migration question rather
  than a one-line guard. Named so a later reader does not scope it as small.
