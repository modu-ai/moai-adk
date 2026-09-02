# SPEC Review Report: SPEC-SESSION-TELEMETRY-001

Iteration: 1/3 (full audit — this SPEC has not been audited before)
Tier: **L** — PASS threshold **0.85**
Verdict: **PASS**
Overall Score: **0.91** (harmonic mean of the four dimensions)
Blocking-class findings: **2** (F3, F7) — neither is a must-pass criterion, so neither forces FAIL; both should be closed by an edit before run-phase entry.

Auditor context isolation: this report was produced from the five SPEC artifacts plus the tree
itself. Reasoning context from the SPEC author was not supplied and would have been ignored per M1.
The three context documents named in the assignment (`spec-split-design.md`,
`plan-audit-iter2-independent.md`, `SPEC-WEB-CONSOLE-015/`) were read as context, not audited.

Audit tree: `<worktree>`, `HEAD = ee039da30`.
The SPEC attributes every baseline to `dfbf828a6`. That commit is one behind `HEAD`, and the
single intervening commit (`ee039da30 plan(t207): split SPEC-WEB-CONSOLE-015 into four SPECs`)
touches no file this SPEC measures — verified:

```
$ git diff --stat dfbf828a6..HEAD -- internal docs-site .claude
(no output)
```

So every re-measurement below is directly comparable to the SPEC's stated baseline.

---

## Must-Pass Results

- **[PASS] MP-1 — REQ number consistency.** `grep -o 'REQ-ST-[0-9]*' spec.md | sort -u` →
  `REQ-ST-001 … REQ-ST-009`, nine ids, no gap, no duplicate, uniform three-digit padding. Nit
  (not a failure): the §B presentation order is non-monotonic — REQ-ST-007 appears between -004
  and -005 because §B.1 groups by subject. Folded into F9.
- **[PASS] MP-2 — GEARS format compliance.** Judged against the **requirement layer** (`REQ-ST-*`
  in `spec.md` §B) only; the Given-When-Then entries in `acceptance.md` are the verification layer
  and are graded under Group 4. All nine match a GEARS pattern: -001/-002/-005/-006 ubiquitous
  (`The X shall …`), -003 ubiquitous + event-driven compound, -004 `Where` used correctly for a
  static-configuration gate (a session running against a non-Claude backend is fixed by the
  `ANTHROPIC_DEFAULT_*` env slots at launch), -007/-008/-009 event-driven `When`. -008/-009's
  trigger ("When the per-session path is adopted") is a change-event rather than a runtime one —
  form-valid, and it is the exact correction the parent audit's F13 asked for (`Where` → `When`).
- **[PASS] MP-3 — YAML frontmatter validity.** All 12 canonical fields present with correct types:
  `id`, `title`, `version: "0.1.0"` (quoted), `status: draft`, `created`/`updated` ISO dates,
  `author`, `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated
  string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`)
  appears. Three non-canonical extras (`era`, `tier`, `related_specs`) are additive.
- **[PASS] MP-4 — language neutrality.** N/A-equivalent: the SPEC is scoped to this repository's
  own Go source and its own doctrine/docs surfaces. It names no per-language toolchain and
  promotes no language. The two template-mirrored doctrine files it edits carry language-neutral
  prose about a state file.
- **[PASS] MP-5 — D7 cross-SPEC reconciliation.** SPEC ids referenced in the body:
  `SPEC-WEB-CONSOLE-015` (`status: draft`), `SPEC-HANDOFF-THRESHOLD-001` (`status: completed`).
  Neither is `retired` / `superseded` / `archived`; no reconciliation obligation fires. No
  referenced SPEC is missing from `.moai/specs/`. No BLOCKING finding.
- **[PASS] MP-6 — D8 cross-platform discipline.** `grep -rn 'syscall' .moai/specs/SPEC-SESSION-TELEMETRY-001/`
  → 0. Auto-PASS per D8-4.
- **[PASS] MP-7 — clarification gate.** `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-SESSION-TELEMETRY-001/`
  → 0 across all five artifacts plus `progress.md`.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0, near 1.0 | Nine requirements, each with a single reading. Deductions: REQ-ST-003's second cause is unreachable under D-3 (F2); REQ-ST-008/-009 phrase a change-process obligation in system-behaviour form (F9). |
| Completeness | 0.92 | 0.75–1.0, near 1.0 | HISTORY, background, requirements, constraints, six `### Out of Scope — …` H3 headings each with specific bullets (`spec.md:202,209,214,219,224,229`), a separate `acceptance.md` with §E edge cases and an 8-item §F Definition of Done, a `design.md` decision record, and a `research.md` carrying its own §8 Gaps and §9 Residual-risk. Deduction: `internal/spec/drift_cache.go:24` is enumerated as a consumer surface and carries a DoD line but no requirement and no criterion (F5). |
| Testability | 0.85 | 0.75 band | Every criterion names a command; every absence-satisfied criterion states a measured pre-change baseline, and all of them reproduce (below). Deductions: AC-ST-005 pins a symbol name no requirement fixes and that D-1 renames away from (F3); AC-ST-006 states a post-change hit count no requirement fixes (F4); AC-ST-009(b)'s baseline is unstated (F8). |
| Traceability | 1.00 | 1.0 | The §D table maps all nine requirements to eleven criteria; I checked both directions independently against the criterion headings. No orphan criterion, no uncovered requirement. |

**Aggregate: harmonic mean = 4 / (1/0.90 + 1/0.92 + 1/0.85 + 1/1.00) = 4 / 4.3746 = 0.9144 → 0.91.**

The harmonic mean is used deliberately, per the skeptical-evaluation contract
(`agent-common-protocol.md` § Skeptical Evaluation Stance) — an arithmetic mean of the same four
values is 0.9175, which is close here only because no dimension is weak. Had Testability landed at
0.50 the arithmetic mean would still clear 0.85 while the harmonic mean would not.

0.91 ≥ 0.85, and all seven must-pass criteria pass ⇒ **PASS**.

---

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

Every measured claim in `spec.md` §A, `acceptance.md`'s baselines, and `research.md` §1–§7 was
re-run against this tree. All of them reproduce, with **one line-number miss** (F1) and **one
narrower-than-stated premise** (F6).

### Evidence

**1. The consumer-surface counts (`spec.md` §A.4) — all four reproduce exactly.**

```
$ grep -rln "state/context-usage.json" .claude internal/template/templates
.claude/rules/moai/workflow/context-window-management.md
.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management.md

$ grep -rln "context-usage.json" docs-site/content | wc -l
      12

$ grep -rln "context-usage.json" docs-site/content
docs-site/content/ja/advanced/statusline.md
docs-site/content/ja/advanced/token-budget.md
docs-site/content/ja/cli-reference/tokens.md
docs-site/content/zh/advanced/statusline.md
docs-site/content/zh/advanced/token-budget.md
docs-site/content/zh/cli-reference/tokens.md
docs-site/content/ko/advanced/statusline.md
docs-site/content/ko/advanced/token-budget.md
docs-site/content/ko/cli-reference/tokens.md
docs-site/content/en/advanced/statusline.md
docs-site/content/en/advanced/token-budget.md
docs-site/content/en/cli-reference/tokens.md
```

Twelve files, exactly three per locale across `en`/`ko`/`ja`/`zh` — AC-ST-011(c)'s baseline is
correct as stated.

**2. Both mirror pairs are byte-identical today (`spec.md` §C.3, AC-ST-009(c)).**

```
$ diff -q .claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md
pair1 rc=0
$ diff -q .claude/rules/moai/workflow/context-window-management-detail.md internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md
pair2 rc=0
```

**3. AC-ST-006's baseline table reproduces line-for-line.**

```
$ grep -rn '"raw_pct"' internal/
internal/statusline/context_usage.go:63:	RawPct            float64 `json:"raw_pct"`
internal/statusline/context_usage_test.go:150:		"context_window_size", "tokens_used", "raw_pct", "stage", "band",
internal/cli/tokens.go:86:	RawPct            float64 `json:"raw_pct"`
internal/cli/tokens_test.go:283:	snap := `{"schema_version":1,…,"raw_pct":50.0,…}`

$ grep -rln '"raw_pct"' internal/ | grep -v '^internal/statusline' | wc -l
       2
```

**4. AC-ST-007's removal-half baseline reproduces exactly — four hits, correctly classified.**

```
$ grep -rn 'context-usage' internal/cli/
internal/cli/tokens.go:30:const tokensContextSnapshotFilename = "context-usage.json"
internal/cli/tokens.go:79:// tokensContextSnapshot is the subset of the statusline context-usage.json
internal/cli/tokens.go:426:			"…embed the context-usage snapshot when present, and append one JSON "
internal/cli/tokens_test.go:284:	if err := os.WriteFile(filepath.Join(stateDir, "context-usage.json"), …
```

This is the parent audit's F10 defect, closed: `:79` is named as the doc comment and `:81` as the
declaration, and `:426` is named as the out-of-scope help string. Verified — `grep -n 'type
tokensContextSnapshot' internal/cli/tokens.go` → `81:`.

**5. AC-ST-005's exported-reader baseline is 0, as stated.**

```
$ grep -rE '^func Read[A-Za-z]*ContextUsage' internal/statusline/*.go | wc -l
       0
```

**6. AC-ST-002's grep baseline is 0, as stated.**

```
$ grep -rn 'CurrentSideChannelFile\|current-session-id' internal/statusline/ | wc -l
       0
```

**7. AC-ST-001's baseline holds right now.**

```
$ ls -la .moai/state/context-usage.json
-rw-r--r--@ 1 goos  staff  271 Aug 24 13:43 .moai/state/context-usage.json
$ ls -d .moai/state/context-usage
ls: .moai/state/context-usage: No such file or directory
```

**8. Line-number citations — 13 of 14 land on the declaration they name.** Verified with
`grep -n` / `sed -n` per symbol:

| Cited | Verified |
|---|---|
| `context_usage.go:56` `contextUsageRecord` | ✅ `56:type contextUsageRecord struct {` |
| `context_usage.go:63` `"raw_pct"` | ✅ |
| `context_usage.go:134` write path | ✅ `134:	path := filepath.Join(stateDir, "context-usage.json")` |
| `context_usage.go:186` `readContextUsage` | ✅ |
| `context_usage.go:203/216/236` | ✅ all three are the `func` lines |
| `context_usage.go:27` `contextUsageSchemaVersion` | ❌ **`:27` is the doc comment; the const is at `:28`** — F1 |
| `builder.go:168` `writeContextUsage(...)` | ✅ |
| `builder.go:286-287` effort threading | ✅ |
| `types.go:69` `Effort *EffortInfo` | ✅ |
| `types.go:131` `EffortInfo{Level}` | ✅ (type at `:131`, field at `:132` — cited as the struct, accepted) |
| `types.go:148` `ModelInfo.DisplayName` | ✅ |
| `metrics.go:51` / `:197` `resolveGLMModelName` | ✅ both |
| `tokens.go:30` / `:79` / `:81` / `:86` / `:393-397` | ✅ all, with comment-vs-declaration correctly distinguished |
| `registry.go:39` / `:52` | ✅ both |
| `spec/drift_cache.go:24` | ✅ and correctly described as a comment |
| `profile.go:73` `ModelEffort`, `:92` `EffectiveProfile`, `:70-72` doc | ✅ all three |
| `context-window-management.md:100` | ✅ the line naming `<projectDir>/.moai/state/context-usage.json` |

**9. The doctrine quotations in `research.md` §7 are verbatim.**

```
$ grep -n 'writer_pid\|last-writer-wins\|discriminator' .claude/rules/moai/workflow/context-window-management-detail.md
50:- `session_id` / `writer_pid` / `captured_at` — validity-guard inputs (see §2)
59:- **Real session id on both sides**: valid only when the record's `session_id`
60:  equals the current session id (last-writer-wins). A differing id → treat as
65:  sessions both lack a real id and share one file, the `writer_pid` discriminator
```

**10. The four routed parent findings are closed.**

| Parent finding | Closure in this SPEC | Verified |
|---|---|---|
| **F1** — named producer cannot produce model/effort | Producer relocated to the statusline (D-1); `spec.md` §A.2 measures why the launcher cannot; REQ-ST-004 + D-5 handle the GLM backend through the existing `resolveGLMModelName` | ✅ `metrics.go:51,197` exist and are on the render path |
| **F5** — 12 docs-site files with no requirement, criterion, or DoD | REQ-ST-009 **and** AC-ST-011 **and** DoD line 6 — all three | ✅ AC-ST-011's grep scope is `docs-site/content`, which covers all twelve (evidence 1) |
| **F10** — incomplete removal baseline, `:79` cited as a declaration | AC-ST-007's baseline names all four hits and distinguishes `:79` comment from `:81` declaration | ✅ evidence 4 |
| **F13** — `Where` used for a temporal condition | REQ-ST-008/-009 now use `When`; the only `Where` (REQ-ST-004) is a genuine static-configuration gate | ✅ |

**11. The hard cut is asserted by decision, not by accident (audit item 5).** No criterion
presupposes a dual-write window. AC-ST-001's second half asserts the old path's *absence*, and
both `acceptance.md` (AC-ST-001) and `design.md` §1.2 state explicitly that this half is
legitimate only under D-3 and would fail by construction under a dual-write window. `design.md`
§3's population table states the pre-change population is not read at all — consistent with "no
fallback read". No fallback appears anywhere in the requirement set.

**12. The `internal/cli` reader migration asserts survival, not just removal (audit item 4).**
AC-ST-007 has two named halves: a **Removal** half (the constant and the duplicate declaration
gone) and a **Preservation** half (`moai tokens` still emits the context block for a readable
record; exits zero with the block absent otherwise). The silent-failure mechanism is correctly
characterised — `readTokensContextSnapshot` returns `nil` on any read error at
`tokens.go:393-397`, verified — and `plan.md` M1's sequencing (consolidate before moving) exists
precisely to defuse it.

**13. The record-schema widening addresses pre-change records and leaves the throttle question
open (audit item 6).** `contextUsageSchemaVersion` is bumped (§C.1, M3, `design.md` §3 with a
stated reason for bumping on an additive change); AC-ST-010 pins absent-field tolerance using
writer-produced bytes; `plan.md` §F item 1 records the throttle interaction as **Unmeasured**
with a named measurement and a named fallback. That is an honest open item, not an assumption —
but its stated hypothesis is contradicted by this repository's own doctrine (F7).

**14. Requirement / criterion budget (audit item 7).** 9 requirements, 11 criteria, against Tier L
ceilings of 25 and 25 applied independently. Both well inside.

### Baseline-attribution

Every figure above was measured in this run, in the worktree
`<worktree>`, at `HEAD = ee039da30`, with the
`dfbf828a6..HEAD` diff over `internal docs-site .claude` empty (shown at the top of this report).
Nothing is carried from the parent SPEC's audit reports, from `spec-split-design.md`, or from the
SPEC's own citations — where a figure agrees with one of those, it agrees because it was re-run.

### Gaps — what this audit did NOT observe

1. **No Go code was compiled or executed.** `go build`, `go vet`, `go test` were not run, per the
   assignment's instruction. Every code claim is a read of the tree.
2. **The runtime payload capture (`spec.md` §A.2, `research.md` §4) is not reproducible from the
   tree.** The claim that Claude Code 2.1.241 delivers `effort.level` and
   `model.display_name` on the render input rests on a one-off stdin capture the author performed
   and describes. I could not re-observe it. The static half — that `types.go` declares both
   fields and `builder.go` threads `Effort` — I did verify.
3. **The §A.1 three-live-sessions observation is a moment-in-time record and has since changed.**
   `.moai/state/context-usage.json` now holds `3db058e1…` / `raw_pct 39` (the t207 session). The
   *structural* claim — one slot, no per-session directory — reproduces.
4. **The twelve docs-site pages were counted, not read.** How much of each page M6 must rewrite is
   unassessed. `research.md` §8 item 5 records the same gap.
5. **The two sibling SPECs from the same split** (`SPEC-KANBAN-RECORD-SESSION-KEY-001`,
   `SPEC-WEB-TODO-QUEUE-001`) were not audited; only the `SPEC-WEB-CONSOLE-015` boundary was
   checked, as assigned.
6. **`progress.md`'s uniqueness check no longer reproduces** — it reports
   `ls .moai/specs | grep -i "SESSION-TELEMETRY"` returning no match, which was true before this
   SPEC's directory existed and is false now. Informational; not carried as a finding.

### Residual-risk

- **The boundary with `SPEC-WEB-CONSOLE-015` holds in both directions, as far as documents can
  show it.** The parent's `depends_on: [SPEC-SESSION-TELEMETRY-001, …]` and its
  `### Out of Scope — session telemetry production (owner: SPEC-SESSION-TELEMETRY-001)` match this
  SPEC's `### Out of Scope — console presentation`; the parent declares nothing under
  `internal/web` that duplicates the record schema, and this SPEC's DoD carries
  "No file under `internal/web` modified". Two asymmetries survive: the parent's canonical
  `related_specs` frontmatter does **not** list `SPEC-SESSION-TELEMETRY-001` (the dependency lives
  only in the non-canonical `depends_on` key), and the parent's requirement ids are deliberately
  gapped after the carve-out. Neither is this SPEC's defect; both are noted so a later reader does
  not read the gap as loss.
- **`moai tokens`'s post-change session key source is unstated but available.** REQ-ST-006 requires
  the migration without saying which session's record `moai tokens` reads. It is implementable:
  `tokens.go:275-287` resolves a `sessionID` (from `--session` or the transcript filename) and
  `:338` calls `readTokensContextSnapshot(stateDir)` with that id already in scope. The run will
  find the answer; the SPEC does not hand it over.
- **Six criteria are grep-shaped.** They are well-baselined, but a grep asserts the presence or
  absence of a string, never that the surrounding prose is correct. AC-ST-011 in particular can
  pass on twelve pages whose surrounding explanation is now wrong. `research.md` §8 item 5 names
  the same limit.
- **M5 edits an always-loaded doctrine file.** An error there reaches every session of every
  project that takes the template. The SPEC recognises this (`research.md` §9) and answers it with
  the mirror-pair `diff -q` criterion, which detects a one-sided edit but not a wrong edit made to
  both sides.

---

## Defects Found

**D1 (F1).** `spec.md` §C.1 (`spec.md:161` region), `design.md` §1.1, `plan.md` M3 — all three
cite `context_usage.go:27` for `contextUsageSchemaVersion`. Measured, `:27` is the doc comment and
`:28` is the `const`. — Severity: **minor** — Class: **optional** — Required fix: change `:27` to
`:28` in all three documents, or cite it as `:27-28` the way `tokens.go:79/:81` is already
disambiguated three lines earlier in the same SPEC. *(This is the shape of the parent's F10; the
SPEC closed it for `tokens.go` and reproduced it once for `context_usage.go`.)*

**D2 (F2).** `spec.md` REQ-ST-003 names two causes for an absent value: the render input omitted
it, **or** "the record was written by a build predating this schema". The second cause is
unreachable under D-3: `design.md` §3's population table states that pre-change records live at
the single-slot path, that the path is gone, and that there is no fallback read — so no
pre-change record is ever read. AC-ST-010 exercises the case with a synthetic fixture generated in
the test. — Severity: **minor** — Class: **optional** — Required fix: either drop the second cause
from REQ-ST-003 and let AC-ST-010 stand on its stated fixture rationale, or state in the
requirement that the clause is defensive rather than reachable, so a run does not go looking for a
migration path D-3 forecloses.

**D3 (F3).** `acceptance.md` AC-ST-005 asserts
`grep -rEc '^func Read[A-Za-z]*ContextUsage' internal/statusline/*.go` sums to exactly 1. That
pins the exported reader's **name** to `Read…ContextUsage`. REQ-ST-005 fixes no name, and D-1 /
`plan.md` M3 rename the record type to `sessionTelemetryRecord` while the SPEC's own vocabulary
throughout is "session telemetry". An implementation that exports `ReadSessionTelemetry` — a name
the SPEC's own decisions point at — satisfies REQ-ST-005 and **fails** AC-ST-005. — Severity:
**minor** — Class: **blocking** (internal inconsistency between a criterion and a ratified
decision in the same SPEC) — Required fix: either state the reader's name as a constraint in §C so
the coupling is deliberate, or rewrite the criterion to count exported readers without pinning the
identifier (e.g. assert exactly one exported symbol whose signature returns the record type, or
name the intended symbol explicitly in REQ-ST-005).

**D4 (F7).** `design.md` §1.1 ("The values are stable for a session's lifetime, which is why they
are expected not to disturb the write throttle") and `plan.md` §F item 1 ("both values are stable
for a session's lifetime") rest on a premise this repository's own always-loaded doctrine
contradicts:

```
$ grep -n 'mid-session model or effort switch' .claude/rules/moai/workflow/cache-aware-execution.md
27:10. **A mid-session model or effort switch busts the cache** … Changing model or effort
    mid-session (thinking budget included — `MAX_THINKING_TOKENS`) discards the prompt cache …
```

Model and effort are user-switchable mid-session. The consequence the SPEC does not record runs in
the opposite direction from the one it does: `sameSemanticPayload` (`context_usage.go:203`)
compares `session_id`, `stage`, `context_window_size`, and integer `raw_pct` only, so if the new
fields are **excluded** from the comparison — which is exactly the fallback §F item 1 prescribes —
a mid-session effort or model change is not persisted until an unrelated context value moves. The
record then holds a value that is present and wrong, a state REQ-ST-003's "not recorded" path does
not cover. — Severity: **minor** — Class: **blocking** (an unrecorded failure mode of the
prescribed fallback) — Required fix: correct the stability premise, and extend §F item 1 to record
the staleness direction: if the fields are excluded from the throttle comparison, state what makes
a changed model or effort reach disk.

**D5 (F4).** `acceptance.md` AC-ST-006 states "the post-change assertion is 4 hits in 2 files".
The binding half — every `"raw_pct"` hit lies inside `internal/statusline` — is sound and
well-baselined. The hit count is not fixed by any requirement: `internal/statusline` carries 2
hits today, and reaching 4 depends on where `tokens_test.go:283`'s fixture lands, which the
criterion describes only as "moves with it". A correct implementation landing on 2, 3, or 5 hits
fails the stated number. — Severity: **minor** — Class: **optional** — Required fix: drop the
count, or state it as an expectation rather than an assertion and name the destination of the
migrated fixture.

**D6 (F5).** `internal/spec/drift_cache.go:24` is enumerated as a consumer surface in `spec.md`
§A.4 (count 1) and carries a Definition-of-Done checkbox, but no requirement and no criterion
covers it. This is a small instance of the parent's F5 shape (enumerated ⇒ swept by prose ⇒ can
ship stale), differing in that a DoD line exists — and a DoD line is self-checked by the
implementer, not asserted by a command. — Severity: **minor** — Class: **optional** — Required
fix: fold the comment into AC-ST-009 as a fourth half (`grep -rn 'context-usage.json'
internal/spec/` → 0), or state explicitly that a NOTE-level comment is deliberately DoD-only.

**D7 (F6).** `spec.md` §A.2's first row claims `moai cc` "never parses or sets a model", cited to
`grep -rn '"-m"\|"--model"\|Model' internal/cli/cc.go` → 0 lines. The grep reproduces (0 lines),
but the claim is narrower than the conclusion it supports: `moai cc` **documents** `-m, --model` at
`cc.go:36` and passes unparsed arguments straight through —
`cc.go:225 return unifiedLaunch(profileName, "claude", filteredArgs)` — so a model can be present
in the launcher's argv even though the launcher never reads it. Two secondary points: the cited
command's own output (0) does not produce the `:36` exception the sentence then names (a different,
unquoted grep does), and `research.md` §3 presents both as one block. The conclusion is unaffected
— effort never appears in argv and the model usually does not — but the premise is stated more
strongly than measured. — Severity: **minor** — Class: **optional** — Required fix: restate as
"`moai cc` does not read or record a model; `-m/--model` is passed through to `claude` unparsed
(`cc.go:36`, `:225`)", and split the two greps in `research.md` §3.

**D8 (F8).** `acceptance.md` AC-ST-009 states baselines for halves (a) and (c) but not for half
(b) (`grep -rln "state/context-usage/" .claude internal/template/templates`). Measured, the
baseline is **0** — the pattern cannot match `state/context-usage.json` because the following
character is `.`, not `/` — so the half is genuinely new information; it simply does not say so
while its two siblings do. — Severity: **minor** — Class: **optional** — Required fix: state the
measured 0 baseline for half (b).

**D9 (F9) — form nits, grouped.** — Severity: **minor** — Class: **optional**
- `research.md`'s `$ command` blocks are annotated and reformatted rather than verbatim: `←`
  annotations appended to output lines (§3, §5), `_ =` and `//nolint:errcheck` stripped from the
  `glm.go` `sed` output, `types.go:131`'s three-line struct rendered as one line, and the
  docs-site listing brace-collapsed to `content/{en,ja,ko,zh}/…`. Every result reproduces when
  re-run, so this is a presentation defect and not a false claim — but
  `verification-claim-integrity.md` §3 requires the Evidence section to carry the command **plus
  its verbatim output**. Fix: paste raw output and put annotations outside the block.
- `spec.md` §B presents the requirements out of numeric order (001, 002, 003, 004, **007**, 005,
  006, 008, 009) because §B.1 groups by subject. No gap or duplicate; renumber or accept.
- REQ-ST-008 and REQ-ST-009 express a change-process obligation ("updated in the same change") in
  system-behaviour form. Form-valid GEARS and the correct fix for the parent's F13, but neither is
  observable at runtime — both are observable only as a repository state, which is what their
  criteria actually assert.
- `acceptance.md` AC-ST-011(a) greps `context-usage.json` with an unescaped `.`. Harmless against
  the intended post-change string (`context-usage/<session-id>.json` does not match), noted only
  because the criterion asserts an exact count.

---

## Regression Check

Not applicable — iteration 1.

---

## Recommendation

**PASS at 0.91**, above the Tier L threshold of 0.85, with all seven must-pass criteria passing.

This is a materially stronger SPEC than its parent on exactly the axes the parent failed. All four
routed findings (F1, F5, F10, F13) are closed with requirement + criterion + DoD coverage rather
than prose, every absence-satisfied criterion carries a baseline and every baseline I re-ran
reproduces, the hard cut is asserted by stated decision rather than by accident, the
`internal/cli` silent-break hazard is both requirement-covered and sequenced against in
`plan.md` M1/§E, and `research.md` §8 volunteers six gaps of its own — including the two the
plan then carries as §F run-phase measurements.

Before run-phase entry, close the two blocking-class findings by edit — neither needs a
re-audit iteration:

1. **D3** — decouple AC-ST-005 from the reader's identifier, or fix the name as a §C constraint.
   As written, the SPEC's own rename decision (D-1) points at a name its own criterion rejects.
2. **D4** — correct the "stable for a session's lifetime" premise and record the staleness
   direction in `plan.md` §F item 1, so the run does not adopt the prescribed throttle fallback
   without seeing what it costs.

The seven optional findings (D1, D2, D5–D9) are surfaced for the orchestrator's discretion. D1
(the `:27` → `:28` citation) and D8 (the unstated half-(b) baseline) are one-line edits worth
taking while the other two are in hand; the rest are judgement calls the author is entitled to
decline.
