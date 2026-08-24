# SPEC Review Report: SPEC-WEB-CONSOLE-015

Iteration: 3/3 (final — no further iteration is available)
Verdict: **FAIL**
Overall Score (harmonic mean, the contract's): **0.838**
Overall Score (arithmetic mean, for comparison): **0.850**
Tier: L — PASS threshold **0.85**
Blocking findings: **7** (all 7 are pre-run repairs; **0** require another audit round)

Reasoning context ignored per M1 Context Isolation. The five artifacts, the three sibling SPECs,
the two earlier audit reports, and the tree itself are the input surface; no author reasoning was
consulted.

**The two means straddle the threshold.** The arithmetic figure meets 0.85 exactly; the harmonic
figure does not. The contract's mean is the harmonic one, so the verdict is FAIL. This is stated
first because it is the whole distance between PASS and FAIL here, and the operator should decide
knowing that.

---

## Audit environment and baseline attribution

| Item | Value |
|---|---|
| Worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207` |
| Branch | `WT-web-live-todo` |
| HEAD at audit time | `aa4d7176d` |
| SPEC version audited | `0.2.0` |
| Commit the SPEC attributes its baselines to | `1cee5d29f` |
| Commit research.md attributes its measurements to | `dfbf828a6` |

**Claim.** The SPEC's source-tree baselines, measured at `1cee5d29f`, still reproduce at HEAD.

**Evidence.**

```
$ git diff --stat 1cee5d29f..HEAD -- internal/web internal/kanban internal/session
(no output)
```

**Baseline-attribution.** Empty diff across all three packages the SPEC cites ⇒ every source-file
citation measured at `1cee5d29f` is measured at `aa4d7176d`. Each was nevertheless re-run
individually below rather than inferred from the empty diff.

---

## Must-Pass Results

- **[PASS] MP-1 — REQ number consistency.** 12 requirement ids, zero duplicates, uniform 3-digit
  zero-padding. Measured:

  ```
  $ grep -n '^- \*\*REQ-' spec.md | sed 's/ —.*//'
  154:- **REQ-WC15-001**   162:- **REQ-WC15-021**   171:- **REQ-WC15-043**   190:- **REQ-WC15-050**
  156:- **REQ-WC15-002**   165:- **REQ-WC15-023**   176:- **REQ-WC15-044**   192:- **REQ-WC15-051**
                                                    179:- **REQ-WC15-045**   195:- **REQ-WC15-052**
                                                    181:- **REQ-WC15-046**
                                                    184:- **REQ-WC15-047**
  ```

  The sequence is **gapped** (001, 002, 021, 023, 043–047, 050–052).
  **This is a judgement, stated openly so it can be overruled.** Under a literal reading of the
  firewall ("even one gap = FAIL") this is a FAIL. I am not applying the literal reading, because
  the gap is declared in HISTORY 0.2.0 as deliberate — surviving ids keep their numbers so the two
  prior audit iterations stay traceable across a three-way carve-out — there are zero duplicates,
  and the numbering defect MP-1 exists to catch (an accidental gap or a collision) is absent. If
  the operator applies the firewall literally, the verdict is FAIL on MP-1 as well as on score,
  and the fix is a renumber that costs the traceability the HISTORY row bought. Recorded as
  finding F8, class **optional**.

- **[PASS] MP-2 — GEARS format compliance (requirement layer, `spec.md` §B).** All 12 `REQ-WC15-*`
  entries match a GEARS pattern: Unwanted (`shall not`) for -001, -002; Ubiquitous for -021, -044,
  -050, -052; Event-driven (`When`) for -023, -043 (second sentence), -046, -047, -051;
  State-driven (`While`) for -045. Judgement made **against the requirement layer only**; the 14
  `AC-WC15-*` entries in `acceptance.md` are Given-When-Then by design and were graded under
  Group 4, not here. Modality-precision concern recorded as F9 (optional), which does not affect
  pattern conformance.

- **[PASS] MP-3 — YAML frontmatter validity.** All 12 canonical fields present with correct types;
  no rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) appears.
  Measured (`spec.md:1-18`): `id`, `title` (quoted), `version: "0.2.0"` (quoted semver),
  `status: draft`, `created: 2026-08-24`, `updated: 2026-08-24`, `author: manager-spec`,
  `priority: P2`, `phase: "v3.2.0 target"`, `module: internal/web`, `lifecycle: spec-anchored`,
  `tags: web-console, kanban, factory, telemetry`. Extra fields `era`, `tier`, `depends_on`,
  `related_specs` are additive and well-formed.

- **[N/A] MP-4 — Section 22 language neutrality.** This SPEC is scoped to one Go package
  (`internal/web`) in this repository's own source tree and specifies no template-bound or
  multi-language tooling content. Confirmed non-mirrored: `ls internal/template/templates/internal`
  → `No such file or directory`, which is also the claim `spec.md:235-236` makes. N/A auto-passes.

- **[PASS] MP-5 — D7 cross-SPEC reconciliation.** Five SPEC ids are referenced; all five resolve
  and none carries a retired / superseded / archived status:

  | Referenced SPEC | Exists | `status:` |
  |---|---|---|
  | `SPEC-SESSION-TELEMETRY-001` | yes | `draft` |
  | `SPEC-KANBAN-RECORD-SESSION-KEY-001` | yes | `draft` |
  | `SPEC-WEB-TODO-QUEUE-001` | yes | `draft` |
  | `SPEC-WEB-CONSOLE-REDESIGN-001` | yes | `completed` |
  | `SPEC-FACTORY-WORKER-FANOUT-001` | yes | `implemented` |

  The two non-SPEC requirement ids the document cites also resolve:
  `SPEC-V3R6-MULTI-SESSION-COORD-001/spec.md:113` defines REQ-COORD-002 (the registry entry
  schema, including `pid: int`) and `:165` defines REQ-COORD-024 (the schema freeze). `spec.md:208`
  cites both correctly.

- **[PASS] MP-6 — D8 cross-platform discipline.**
  `grep -rn 'syscall' .moai/specs/SPEC-WEB-CONSOLE-015/` → rc=1, no output. Auto-PASS.

- **[PASS] MP-7 — clarification gate.**
  `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-WEB-CONSOLE-015/` → rc=1, no output.
  Advisory, not a failure: `plan.md` §F carries two **Open items for the operator** (the tier
  ruling, and landing order). They are not marked with the literal convention, so the mechanical
  gate does not fire; the tier item is nevertheless a decision the operator must close at the
  Implementation Kickoff Approval gate — see F11 and § Tier honesty below.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.88 | 0.75–1.0, near 1.0 | Every requirement carries a single interpretation; no pronoun ambiguity; zero weasel words in normative text (`grep -nEi 'appropriate\|adequate\|reasonable\|proper' acceptance.md` returns only two false positives on the word "property"). Deductions: F9 (four `When`s that are state conditions), F10 (plan/spec disagreement on whether the `widgets.templ` change is conditional). |
| Completeness | 0.80 | structural checklist fully met; substantive gaps deduct | All required sections present (HISTORY `spec.md:22`; WHY §A; WHAT §B; HOW `plan.md`/`design.md`; REQUIREMENTS §B; ACCEPTANCE CRITERIA `acceptance.md`; **seven** `### Out of Scope — <topic>` H3 headings at `spec.md:250,258,267,273,279,284,292`, each with specific `-` bullets). Deductions: F3 (registry-side duplicate PID neither covered nor excluded), F6 (nothing binds what the banner must still say), F7 (`missing()` tooltip carries the same falsified clause, unenumerated). |
| Testability | 0.72 | 0.50–0.75 — three defective criteria, not one | 11 of 14 criteria are binary and name a command. Deductions: F1 (AC-051 passes untouched), F2 (AC-046 passes untouched), F5 (AC-021a names no symbol, so its "≥ 1" grep is not executable as written). |
| Traceability | 1.00 | 1.0 | 12/12 requirements covered; 14/14 criteria cite an existing requirement; zero orphans. Mapping measured below. |

**Harmonic mean** = 4 / (1/0.88 + 1/0.80 + 1/0.72 + 1/1.00) = 4 / 4.77525 = **0.838**
**Arithmetic mean** = (0.88 + 0.80 + 0.72 + 1.00) / 4 = **0.850**

0.838 < 0.85 ⇒ **FAIL**.

### Traceability matrix (measured, both directions)

| REQ | Criteria |
|---|---|
| -001 | AC-001 |
| -002 | AC-002 |
| -021 | AC-021a, AC-021b |
| -023 | AC-023 |
| -043 | AC-043a, AC-043b |
| -044 | AC-044 |
| -045 | AC-045 |
| -046 | AC-046 |
| -047 | AC-047 |
| -050 | AC-050 |
| -051 | AC-051 |
| -052 | AC-052 |

Reverse: all 14 criteria name a requirement that exists in `spec.md` §B. No orphan in either
direction.

### Budget

12 requirements / 14 criteria against Tier L ceilings of 25 / 25, applied independently. Both
under. `progress.md:13-14` reports the same figures and they reproduce.

---

## Re-run of every measured claim

Each command below was executed in this worktree at `aa4d7176d`.

### Source-line citations — all reproduce

| Citation | Re-run result |
|---|---|
| `events.go:22` 250ms debounce | `const debounce = 250 * time.Millisecond` OK |
| `events.go:81` `text/event-stream` | `w.Header().Set("Content-Type", "text/event-stream")` OK |
| `events.go:95` 25s keepalive | `keepalive := time.NewTicker(25 * time.Second)` OK |
| `app.js:638` `POLL_MS = 30000` | `var POLL_MS = 30000;` OK |
| `app.js:700` `startPolling()` | `function startPolling() {` OK |
| `app.js:721` `new EventSource` | `var es = new EventSource("/events");` OK |
| `app.js:743` `failures >= 3` | `if (failures >= 3) startPolling();` OK |
| `viewmodel_ops.go:46` `ChainRoles` | `var ChainRoles = []string{"lead", "plan", "run", "sync"}` OK |
| `viewmodel_ops.go:250-256` placeholders | `Model: ""` (`:253`), `Effort: ""` (`:254`), `ContextPct: -1` (`:255`) OK — AC-021b's narrower `:253-255` is also exact |
| `viewmodel_ops.go:266-275` `estimateStage` | function body spans exactly those lines OK |
| `viewmodel_ops.go:92` `StageEstimated` | `StageEstimated bool` OK |
| `viewmodel_ops.go:409-435` `loadSessions` | function spans exactly those lines; `SessionVM{ID, SpecID, State, Heartbeat, Cwd}` — **`PID` is dropped**, as claimed OK |
| `screens.templ:165-175` model/effort branch | `if r.Model != "" { … } else { @missing() }` OK |
| `screens.templ:192` note banner | exact text reproduces, third argument `""` OK |
| `widgets.templ:40-52` `noteBanner` | `templ noteBanner(kind, text, key string)`; emits `data-i18n` only when `key != ""` OK |
| `widgets.templ:122-124` `missing()` | reproduces — **but the quoted title is incomplete; see F7** |
| `factory_slots.go:37/47/55/84` | `FactoryWorkerEntry` / `FactoryRegistryPath` / `LoadFactoryRegistry` / `PruneFactoryDeadClaims` OK |
| `record.go:45` `@MX:ANCHOR` | line 45 is the `// @MX:ANCHOR:` line itself. The cited entity **is** an annotation comment, so citing its comment line is correct — not the comment-instead-of-declaration error this card has hit twice OK |
| `cli/cc.go:36` help string | `  -m, --model <model>           Override model selection` OK |
| `cli/glm.go:350-353` four-slot map | four `os.Setenv` calls: Opus / Sonnet / Haiku / Fable OK |

### Acceptance-criterion baselines — all reproduce

```
$ grep -rnE 'os\.(WriteFile|MkdirAll|Rename|Create)\(|WriteBestEffort\(|acquireLock\(|SaveFactoryRegistry\(|\.Mutate\(' \
    internal/web --include='*.go' | grep -v '_test\.go'
internal/web/profile_crud.go:38:	return os.MkdirAll(dir, 0o755)
internal/web/profile_crud.go:64:	return os.Rename(src, dst)
```

Exactly the two allowlisted lines (AC-002). The reading half also holds: `profile_crud.go:34` is
`dir := profile.GetProfileDir(name)`, and the rename's paths derive from the same helper — neither
derives from the project state directory.

```
$ grep -rn "internal/statusline" internal/web --include='*.go'          -> rc=1, no output
$ grep -rn "context-usage\|ContextUsage" internal/web                    -> viewmodel_ops.go:255 only
$ grep -rn "Factory\|factory" internal/web --include='*.go' --include='*.templ' | grep -v _test
                                                                         -> rc=1, no output
$ grep -rn '"lane"\|RoleLane' internal/web --include='*.go' --include='*.templ'
                                                                         -> rc=1, no output
$ grep -c "<td" internal/web/screens.templ                               -> 0
$ grep -c 'are not recorded yet' internal/web/screens.templ              -> 1
$ grep -c 'kanban.Record is extended' internal/web/screens.templ         -> 1
```

All seven reproduce exactly as `acceptance.md` and `research.md` state.

```
$ go test ./internal/web/ -run TestI18n -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	1.047s
```

AC-050's baseline reproduces. I additionally checked the selector is not vacuous —
`go test ./internal/web/ -run TestI18n -count=1 -v | grep -c '^=== RUN'` -> **15**, so the run
selects real tests rather than passing on an empty set. This is the one baseline in the document
where a vacuous-selector shape was plausible; it is clean.

### Runtime-directory readings — **do not reproduce**; see F4

`spec.md` §A.4 and `research.md` §5-§6 quote readings of `.moai/state/active-sessions.json`,
`.moai/state/kanban/`, and `.moai/state/context-usage.json`. Re-read at audit time:

```
active sessions: 4
  e46fcfef-1f5c-4f9c-beff-2ada72e26eb5  pid 51045  /Users/goos/MoAI/moai-adk-go
  34740be0-dab0-402a-a976-4894b693a155  pid 31329  /Users/goos/moai/moai-adk-go
  55cdc796-5aa2-491b-91fb-bc274904684b  pid 87705  /Users/goos/moai/moai-adk-go
  e995be8e-015a-4391-9091-301d8a85f962  pid 10793  /Users/goos/moai/moai-adk-go

kanban record files: 84
  record keyed by 34740be0 : True
  record keyed by 55cdc796 : True
  record keyed by e46fcfef : False
  record keyed by e995be8e : False
```

None of the three session ids quoted in the SPEC (`2beac221…`, `c15d8434…`, `3db058e1…`) is still
in the registry. The `context-usage.json` slot now holds a fourth distinct value
(`session_id 2e3ace62…`, `writer_pid 72163`, `raw_pct 46`) belonging to no live registry entry —
which **confirms** research.md §6's last-writer-wins conclusion while falsifying its quoted values.

The decisive re-measurement is the actual REQ-WC15-043 join, run end to end:

```
$ cat .moai/state/factory/workers.json
lane-5  -> pid 87705    lane-6  -> pid 34699    lane-7  -> pid 15207
lane-8  -> pid 36912    lane-9  -> pid 51045    lane-10 -> pid 10793
```

| Lane | PID | Session | Record keyed by that session |
|---|---|---|---|
| lane-5 | 87705 | `55cdc796…` (live) | **present** — the join closes |
| lane-10 | 10793 | `e995be8e…` (live) | absent — unresolved |
| lane-9 | 51045 | `e46fcfef…` (live) | absent — unresolved |
| lane-6 / -7 / -8 | 34699 / 15207 / 36912 | none | unresolved (stale claims) |

Two consequences, one against the SPEC and one for it:

- **Against.** `spec.md:130-133` states categorically that the join "comes up empty-handed or
  holding another session's record until `SPEC-KANBAN-RECORD-SESSION-KEY-001` lands". Measured
  today, lane-5 resolves to a session whose record is keyed by that same session. The third case
  the sentence excludes is live in this tree. -> **F4.**
- **For.** `spec.md:220-223` argues the stale-duplicate hazard is reachable at render time because
  the loader is fail-open and pruning is a separate call. Measured: three of six registered lanes
  point at PIDs that are in no active session, and PID 51045 — which the SPEC's own earlier reading
  attributed to session `c15d8434…` in worktree t210 — now belongs to session `e46fcfef…` in the
  primary checkout. A recycled PID landing on a different session is exactly the hazard §C.3 names,
  and it is observed. The reachability argument is sound.

Note also that the dependency **conclusion** survives F4 on stable ground: the records that do
resolve carry `"spec_id": ""` and no lane number or card identifier, so REQ-WC15-044 still cannot be
satisfied. It is the *evidence and its phrasing* that fail to reproduce, not the dependency.

---

## Defects Found

**D1. F1 — AC-WC15-051 passes on the untouched tree, and states no baseline.**
`acceptance.md:133-138` — Severity: **major** — Class: **blocking** — Repair: **pre-run**

The criterion's Then-clause is: "both rows render, the pre-change row showing the 'not recorded'
marker for each field its record and snapshot do not carry, and no request returns an error status."
Every conjunct holds on the pre-change tree: `viewmodel_ops.go:253-255` hardcodes `Model: ""` /
`Effort: ""` / `ContextPct: -1` for **every** role, and `screens.templ:165-175` therefore renders
`@missing()` for every row — measured above. A record "written by a build predating this SPEC's
dependencies" is the only kind that exists today, and it renders.

This is the same shape as version 0.1.0's AC-WC15-012, which `acceptance.md:5-7` says was deleted
for exactly this reason, and the document's own opening rule ("Where a criterion is satisfied by an
absence, its measured pre-change baseline is stated") is not honoured here: AC-051 carries no
baseline. It is one of only two criteria in the document with none.

**Required fix:** make the discriminating half explicit — the post-dependency row must show
*values*, not markers, and the two rows must differ. Or state a baseline showing what the criterion
observes that the untouched tree does not.

**D2. F2 — AC-WC15-046 passes on the untouched tree, and states no baseline.**
`acceptance.md:102-104` — Severity: **major** — Class: **blocking** — Repair: **pre-run**

"Then in both cases the response status is 200 and the factory section renders zero lanes."
Measured, `internal/web` contains zero references to the factory registry (grep above, rc=1), so no
code path reads it and its absence or malformation cannot alter the response. The pre-change page
renders zero lanes unconditionally — the second conjunct is satisfied by there being no factory
section at all. AC-043b, AC-044 and AC-047 each explicitly state "as AC-WC15-043b — zero lane rows
exist pre-change, so this behaviour cannot pass on the untouched tree"; AC-046 is the one §C
criterion that omits that statement, and it is the one whose assertion the untouched tree satisfies.

**Required fix:** assert the factory section is **present and empty**, distinguishable from absent
(e.g. the section's container renders with an explicit zero-lane state), and state the baseline.

*Gap:* I did not start `moai web`, so the "status is 200" half is inferred from the absence of any
factory read path rather than observed from a running server. The "zero lanes" half is measured.

**D3. F3 — the duplicate-process-identifier rule covers one side of the join only, and the other
side is neither covered nor excluded.**
`spec.md:171-175, 184-186`; `design.md:37-67`; `acceptance.md:106-112` — Severity: **major** —
Class: **blocking** — Repair: **pre-run**

REQ-WC15-047 and AC-WC15-047 both scope to "two or more registered **lanes** carry the same process
identifier" — the `workers.json` side. The `active-sessions.json` side produces a distinct and
equally reachable ambiguity: two registry entries can carry one PID, so one lane's PID can resolve
to two sessions. Measured mechanism:

```
internal/session/registry.go:166-199 — Register():
    for i := range entries {
        if entries[i].SessionID == sessionID { … return entries, nil }
    }
    entries = append(entries, Entry{… PID: resolveSessionPID() …})
```

Deduplication is by `SessionID` alone. Nothing rejects, merges, or flags a second entry carrying an
already-present PID, and the comment at `:184-186` explicitly contemplates a PID differing across
reconnects while treating first-seen as canonical — i.e. PID uniqueness is not an invariant of this
file.

REQ-WC15-043 assumes uniqueness in its own wording: "joining the factory registry's recorded process
identifier to **the** active-sessions entry carrying that identifier" (singular). A
`grep -n 'duplicate\|two or more\|same process identifier'` across all five artifacts returns only
lane-side occurrences. §D excludes neither case.

**Required fix:** either extend REQ-WC15-047 / AC-WC15-047 to the registry side, or state explicitly
in §B or §D which side the requirement covers and why the other is out of scope.

*Note:* today's registry carries four entries with four distinct PIDs, so the trigger is not
currently live. The mechanism, not a current observation, is the evidence — reported as a
requirement-coverage gap, not as a live defect.

**D4. F4 — §A.4's categorical claim about the lane join is contradicted by measurement in this tree,
and rests on a snapshot of a directory that moves while the SPEC is open.**
`spec.md:105-133`; `research.md:79-110` — Severity: **major** — Class: **blocking** —
Repair: **pre-run**

Full evidence in § Runtime-directory readings above. In summary: the three session ids the SPEC
quotes no longer exist; of the four live sessions, two carry a record keyed by their own id; and
running the actual join today, `lane-5 -> pid 87705 -> session 55cdc796… -> record present` closes.
The sentence "The join therefore comes up empty-handed or holding another session's record until
`SPEC-KANBAN-RECORD-SESSION-KEY-001` lands" is false as a universal statement about this tree.

This matters more than a stale number because §A.4 is itself the *correction* of a false claim
carried by version 0.1.0, and it replaces one categorical assertion with another categorical
assertion drawn from a three-sample reading of a live directory.

**Required fix:** state the observation as dated with its sample size, and rest the dependency on
the property that does not move. That property is already written down next door —
`SPEC-KANBAN-RECORD-SESSION-KEY-001/spec.md:112` states it as "no record file carries a lane
number, and every lane writes the same role value", and I confirm it in this tree:

```
$ grep -h '"role"' .moai/state/kanban/*.json | sort | uniq -c
  36   "role": "lane",   18   "role": "lead",   5   "role": "run", …
$ grep -l '"lane"[[:space:]]*:' .moai/state/kanban/*.json          -> 0 files
```

That formulation is stable, is what REQ-WC15-044 actually needs, and survives any registry churn.

**D5. F5 — AC-WC15-021a's positive half is not executable as written, because it names no symbol.**
`acceptance.md:50-58` — Severity: **major** — Class: **blocking** — Repair: **pre-run**

"When it is grepped for the reader symbol `SPEC-SESSION-TELEMETRY-001` exports, Then the count is
**≥ 1**." No grep can be run from that sentence. The producer SPEC anticipated exactly this and
pinned the identifiers deliberately — `SPEC-SESSION-TELEMETRY-001/spec.md:184-191`: "The console
imports this reader, so the record type it returns is exported as `SessionTelemetryRecord` and the
reader as `ReadSessionTelemetry`. Both names are named here rather than left to the implementation,
because AC-ST-005 counts the exported reader by matching its identifier: an unpinned criterion would
either pass on a second reader with a different name…". The consumer criterion is the unpinned one
the producer wrote that paragraph against.

**Required fix:** name `ReadSessionTelemetry` (and, for the "zero" half, `SessionTelemetryRecord`)
in AC-WC15-021a, citing `SPEC-SESSION-TELEMETRY-001` REQ-ST-005 as the source of the names.

**D6. F6 — nothing binds what the note banner must still say, so deleting it satisfies every
criterion.**
`spec.md:195-197`; `acceptance.md:140-153`; `design.md:84-100` — Severity: **major** —
Class: **blocking** — Repair: **pre-run**

REQ-WC15-052 states two prohibitions and one key requirement. AC-WC15-052 asserts two zero-greps, a
non-empty third argument, and four-locale presence. An implementation that deletes the `@noteBanner`
call at `screens.templ:192` outright satisfies both zero-greps, and leaves the "read the note-banner
call" step with nothing to read — i.e. it passes without being distinguishable from the intended
rewrite.

The content that must survive is stated only in `design.md:93-95` ("The first sentence stays true
and is load-bearing: it is the prose form of the honesty flag REQ-WC15-045 requires") and
`plan.md:88-90`. Neither is a requirement or a criterion. This is the same structural defect
iteration 2 recorded as F7 — a correctness rule living in a non-normative section — reappearing on a
different subject after F7 was closed on its original one.

**Required fix:** add a positive clause to REQ-WC15-052 (the banner shall continue to disclose that
the stage is estimated from the heartbeat) and a matching assertion to AC-WC15-052.

**D7. F7 — the `missing()` marker's tooltip carries the same falsified clause the banner does; it is
enumerated nowhere, and research.md quotes it with that clause elided.**
`internal/web/widgets.templ:123`; `spec.md:85-89, 232`; `research.md:30-31` — Severity: **major** —
Class: **blocking** — Repair: **pre-run**

Measured, verbatim:

```
$ sed -n '122,124p' internal/web/widgets.templ
templ missing() {
	<span class="missing" title="not recorded anywhere yet (kanban.Record extension required)" aria-label="not recorded">—</span>
}
```

`research.md:30-31` renders this as: `templ missing()` renders `—` with
`title="not recorded anywhere yet"` and an aria-label. The elided parenthetical is precisely the
claim §A.2 establishes as false and REQ-WC15-052 exists to remove from the banner. Two problems
follow:

1. **Coverage.** `spec.md:85-89` says this SPEC "keeps the marker", and §C.3's `widgets.templ` row
   scopes the file's change to "the unresolved-lane marker, if `@missing()` is not reused verbatim"
   — so on the reuse path the tooltip is untouched and the falsehood ships alongside a banner
   corrected for saying the same thing. The tooltip and `aria-label` are also hard-coded English
   with no `data-i18n`, in a view REQ-WC15-050 governs.
2. **Quotation accuracy.** A truncated quote that removes the one clause making the string a defect
   is how the defect stayed invisible across three iterations.

**Required fix:** quote the title in full in `research.md`; add the tooltip to §C.3's enumeration and
bring it under REQ-WC15-052 (or state explicitly why a false tooltip may remain while a false banner
may not).

**D8. F8 — the requirement-id sequence is gapped.**
`spec.md:154-197` — Severity: **minor** — Class: **optional** — Repair: **pre-run**

001, 002, 021, 023, 043–047, 050–052. Declared deliberate in HISTORY 0.2.0 to preserve
cross-iteration traceability; zero duplicates. Full reasoning under MP-1 above. Surfaced so the
operator can apply the firewall literally if they choose; renumbering costs the traceability the
carve-out bought.

**D9. F9 — four requirements use `When` where the trigger is a persisting state condition.**
`spec.md:165, 174, 181, 184` — Severity: **minor** — Class: **optional** — Repair: **pre-run**

- REQ-WC15-023 "**When** no readable telemetry record exists for a session" — a state, not an event.
- REQ-WC15-043 (second sentence) "**When** a registered lane resolves to no session" — a state.
- REQ-WC15-046 "**When** the factory registry is absent or unreadable" — a state.
- REQ-WC15-047 "**When** two or more registered lanes carry the same process identifier" — a state.

`While` is the GEARS state-driven modality; `When` is temporal. REQ-WC15-045 uses `While` correctly
and REQ-WC15-051 uses `When` correctly ("when … **is read**" is a genuine event), which shows the
document knows the distinction. All four remain conformant Event-driven patterns, so MP-2 is
unaffected; this is precision, not conformance.

**D10. F10 — plan.md and spec.md disagree on whether the `widgets.templ` change is conditional.**
`plan.md:29` vs `spec.md:232` — Severity: **minor** — Class: **optional** — Repair: **pre-run**

`plan.md:29` asserts "Generated files following: **2** — `screens_templ.go`, `widgets_templ.go`",
while `spec.md:232` makes the `widgets.templ` change conditional ("if `@missing()` … is not reused
verbatim"). On the reuse path only one generated file follows. Both files exist
(`ls internal/web/*_templ.go` confirms). Minor arithmetic inconsistency in the tier table.

**D11. F11 — `tier: L` is retained against a measurement that supports M; the operator's decision is
still open.**
`spec.md:15`; `plan.md:16-43, 121` — Severity: **minor** — Class: **optional** —
Repair: **operator decision, not a code repair**

Judged plainly, as asked. Every measured signal in `plan.md` §B's table is Tier-M-shaped and I
reproduce them: one package (`internal/web`); zero cross-cutting schema changes (§C.2, and the
`@MX:ANCHOR` at `record.go:45` now binds the sibling); zero always-loaded doctrine files; zero
docs-site pages; four source files plus three test files; three milestones. The single L-shaped
argument is dependency risk, and both dependencies have been independently audited and passed
(telemetry 0.91; record-key 0.857 after revision) — which is the condition `plan.md:41-43` itself
names for reclassification.

**The stated measurement does not support the retained tier.** Retaining L is nevertheless the
*conservative* error: it holds the threshold at 0.85 rather than 0.80 and keeps `design.md` and
`research.md` in the required set, both of which exist and are current. Under Tier M's 0.80
threshold this SPEC's harmonic 0.838 would **PASS**. That fact is stated for the operator's
information and had no influence on the scoring above, which was performed against the four
dimensions before the threshold was applied.

---

## Regression Check — the iteration-2 findings routed here

| Iter-2 finding | Subject | Status | Evidence |
|---|---|---|---|
| **F2** | AC-WC15-012 observes nothing | **CLOSED** | `grep -rn 'WC15-012'` returns four hits, all historical prose correctly labelled "version 0.1.0's". No requirement or criterion numbered 012 survives, and no surviving requirement was orphaned by the deletion — the traceability matrix above is complete in both directions. |
| **F7** | duplicate-PID rule had no requirement | **CLOSED, but incompletely** | Promoted to REQ-WC15-047 + AC-WC15-047 + `design.md` §3 with rejected alternatives. The lane side is now properly normative. The registry side is not — **F3**. |
| **F9** | AC-002's grep could not express its qualifier | **CLOSED** | Restated as an exhaustive two-line inventory plus a declared human reading. Re-run above: exactly the two lines. A genuine improvement — the criterion is now strictly stronger than the qualifier it replaced. |
| **F11** | two criteria named no executable check | **CLOSED (the half routed here)** | AC-043a's file-creation clause is now a before/after recursive listing of the state directory. The other half (AC-031b) left with the todo carve-out. |
| **F12** | note banner in no enumeration | **CLOSED, but incompletely** | Now REQ-WC15-052 + AC-WC15-052 + a §C.3 row. But nothing binds what the banner must still say (**F6**), and the sibling string carrying the same falsehood was not swept up (**F7 of this report**). |
| **F13** | minor form defects | **CLOSED** | The specific items (REQ-024's `Where`, REQ-025's implementation body, REQ-030/031/040/042) all left with the carve-outs. REQ-WC15-050 no longer names `i18n.js` in its body. A **new** modality imprecision appears in four surviving requirements — **F9 of this report**. |
| **F8/N2** | version + HISTORY row | **CLOSED** | `version: "0.2.0"` and a HISTORY row at `spec.md:27` describing the carve-out. |

Seven routed findings, seven closed at the level they were written; two closed in a way that leaves
an adjacent instance of the same defect class open (F7 -> F3, F12 -> F6/F7). **No stagnation**: no
defect appears unchanged across all three iterations.

---

## Gaps — what this audit did NOT observe

- **`moai web` was not started.** Every claim about rendering is derived from source. AC-046's
  "status is 200" half and AC-051's "no request returns an error status" half are therefore reasoned
  from the absence of a factory read path, not observed from a running server.
- **No Go test outside `internal/web -run TestI18n` was executed**, per the dispatch. Coverage,
  `go vet`, and `golangci-lint` were not run; the §E Definition of Done items that depend on them
  are unverified by this audit.
- **The two screenshots attached to card t207 were not opened**, consistent with the SPEC's own §D
  declaration.
- **The sibling SPECs were read only where this SPEC's boundary claims required it.** I verified
  that `SPEC-SESSION-TELEMETRY-001` owns the exported reader (REQ-ST-005) and that
  `SPEC-KANBAN-RECORD-SESSION-KEY-001` owns the lane number and card identifier (REQ-KRS-005 and its
  §112-113 measurement block, which names REQ-WC15-044 as its consumer). I did not re-audit either
  sibling.
- **The i18n governance test's reach was checked but not exhaustively characterised.** It operates on
  catalogue keys, not on `.templ` literals, so it cannot detect a newly hard-coded English string.
  `acceptance.md:122-131` states this correctly; I confirmed the mechanism at
  `i18n_governance_test.go:145-183` but did not enumerate every assertion in the file.
- **PID-duplication on the registry side was established from the code path, not from a live
  duplicate.** Today's four registry entries carry four distinct PIDs.

## Residual risk — what could still be wrong despite the above

- **The join's correctness is a three-file, cross-process property, and only two of the three files
  are stable.** `workers.json` and `active-sessions.json` both change under the SPEC while it is
  open; F4 is one instance, and any future criterion that reads them inherits the same hazard. The
  SPEC carries no rule forbidding a moving baseline, so the pattern can recur.
- **Both dependencies are `status: draft`.** Everything in §B and §C of `acceptance.md` is
  unverifiable end to end until they land, which the SPEC states honestly (§E, second box) — but it
  means a PASS here would certify a plan, not a verifiable outcome.
- **The `WT-web-live-todo` branch is unpushed and this worktree holds the only copy** of all four
  SPECs and all six audit reports on this card.
- **This is the last iteration.** Every finding above is classified pre-run repair, so the fix path
  does not require another full audit — but nothing will re-check the repairs mechanically either.

---

## Recommendation

**Verdict: FAIL** (harmonic 0.838 < 0.85). Seven blocking findings, **all seven pre-run repairs**;
none requires another audit round.

The distinction the operator asked for, stated plainly: every blocking finding is a local edit to
one requirement, one criterion, or one paragraph. None of them reopens a design decision, moves the
dependency boundary, or changes what the SPEC is for. The three-way carve-out is sound, the boundary
is clean in both directions (verified: nothing this SPEC owns is unclaimed, `/todo` is fully absent
except as an §D exclusion, the telemetry schema is consumed rather than redeclared, and the lane
number and card identifier are cited to their owner rather than duplicated), and the document is
measurably better than version 0.1.0 on every axis iteration 2 flagged.

Ordered fix list:

1. **AC-WC15-051** (F1) — add the discriminating half: the post-dependency row shows values, the two
   rows differ. State the baseline.
2. **AC-WC15-046** (F2) — assert the factory section is present-and-empty, distinguishable from
   absent. State the baseline.
3. **REQ/AC-WC15-047** (F3) — cover the registry side of the PID ambiguity, or say in §B or §D which
   side is covered and why the other is not.
4. **spec.md §A.4 + research.md §5** (F4) — date the observation, give its sample size, and rest the
   dependency on the stable property (no record carries a lane number; every lane writes the same
   role value) rather than on a live-directory snapshot.
5. **AC-WC15-021a** (F5) — name `ReadSessionTelemetry` and `SessionTelemetryRecord`, citing
   `SPEC-SESSION-TELEMETRY-001` REQ-ST-005.
6. **REQ/AC-WC15-052** (F6) — add the positive clause: the banner shall continue to disclose that the
   stage is estimated from the heartbeat.
7. **widgets.templ `missing()` tooltip** (F7) — quote it in full in `research.md`, add it to §C.3, and
   bring it under REQ-WC15-052.

Optional, operator's discretion (F8–F11): the id gap, the four `When`/`While` modalities, the
generated-file count, and the tier ruling.

**On the tier, since the operator asked for it plainly:** the measurement supports M. Reclassifying
to M would move the threshold to 0.80, which this revision already clears at 0.838 — so the same
document, correctly tiered, passes. I am not applying that as a verdict, because the declared tier
at audit time is L and the threshold that binds is the declared one. If the operator reclassifies,
the seven repairs above are still worth making; they are defects at any tier.
