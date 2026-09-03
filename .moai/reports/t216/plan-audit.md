# SPEC Review Report: SPEC-HOOK-WIRING-DRIFT-001
Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.81** (harmonic mean 0.807; arithmetic 0.8125) — threshold 0.85
Audited at: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t216`, HEAD `4842760a7`, branch `WT-hook-wiring-drift`

Reasoning context ignored per M1 Context Isolation. Position confirmed before any read:
`git rev-parse --show-toplevel` → `.../worktrees/t216`; `git rev-parse --short HEAD` → `4842760a7`; tree clean.
`git log 950cb4399..HEAD` → 2 commits, both SPEC artifacts only (`--stat`: 4 files, +1115, no source changes), so the
pre-implementation tree state the author measured against is intact and every `Pre-impl observed:` value is
independently re-runnable at this HEAD.

## Tooling note (mandatory disclosure)

`mcp__moai__spec_audit` as exposed to this session carries **no `project_root` field** — its schema is
`{filter_era, filter_spec, include_grandfathered}` only. Invoked with `filter_spec=SPEC-HOOK-WIRING-DRIFT-001`
it returned `{"total_specs":0,...}`, i.e. it acted on a catalog that is not this tree. It is **not** used as
evidence anywhere below. Fallback used: `moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md`
→ `No findings`, plus direct reads and re-runs. Caveat: the installed `moai` is `/Users/goos/go/bin/moai`
(v3.1.2, stale relative to this tree, as plan.md §C itself records).

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-HWD-001` … `REQ-HWD-011` in spec.md §B, sequential, no gap, no
  duplicate, uniform 3-digit padding. Verified by direct read of spec.md §B (M1: 001-002, M2: 003-005,
  M3: 006-008, M4: 009-011).
- **[PASS] MP-2 GEARS format compliance** — judgment made against the **requirement layer** (`REQ-XXX` in
  spec.md §B), never against an AC. All 11 match a canonical pattern: Ubiquitous (001, 002, 006, 007, 010),
  Event-driven (003 "When an operator runs `moai doctor`, the hook-wiring diagnostic shall …", 005, 009),
  Unwanted (004 "shall not modify …", 008 "shall not delete …", 011 "No comment … shall assert …").
  Defect noted below (D9): REQ-HWD-005 and REQ-HWD-009 are **labelled** `(Event-detected)`, which is not a
  GEARS pattern name — but both texts are canonical Event-driven form, so the criterion passes on form.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`version: "0.1.0"` quoted semver; `status: draft` ∈ the 8-value enum; `priority: P2`;
  `phase: "v3.1.4 target"` is a release target, not a prohibited lifecycle token; `lifecycle: spec-anchored`;
  `tags` comma-separated). No rejected snake_case alias. The multi-segment id `SPEC-HOOK-WIRING-DRIFT-001`
  is accepted by the **enforced** regex `internal/spec/lint.go:715` `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`
  (the single-segment regex printed in the schema doc's field table is a stale simplification, not this
  SPEC's defect). `moai spec lint` clean.
- **[N/A] MP-4 language neutrality** — the SPEC targets this repository's own hook/settings tooling. The M3
  template edit names shell wrapper scripts and disposition classes, not per-language tooling; no
  language-specific tool name appears. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — extracted references: `SPEC-CHAIN-CORE-001`,
  `SPEC-PROJECT-NAVIGATOR-001`. Both directories exist under `.moai/specs/`; both carry `status: completed`,
  which is outside `{retired, superseded, archived}`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/`
  → rc=1, no match. `research.md` absent (Tier M does not require it); `plan.md` clean.

No must-pass failure. The FAIL is driven by the aggregate score against the 0.85 threshold.

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 — minor ambiguity in one or two requirements | REQ-HWD-004 (spec.md §B, M2 block) says "shall not modify … **any other file on disk**" but its observable narrows to "every file under `.claude/`" — the two do not agree, and the gap is exploitable (mutant M-3). REQ-HWD-002 is compound (a Ubiquitous obligation + an Unwanted one in one entry). Everything else is single-interpretation. |
| Completeness | 0.90 | 1.0 band (all sections + full frontmatter) minus the §C-5 shortfall | HISTORY, §A Context, §B Requirements, §C Constraints, §D file classification, §E NFR, §F Exclusions (four `### Out of Scope — <topic>` H3s each with specific `-` bullets), §G follow-ups, §H cross-refs; acceptance.md §D matrix + §D.5 DoD. Deduction: spec.md §C-5 asserts "Each entry in `acceptance.md` carries a `Pre-impl observed:` value measured on this tree at HEAD `950cb4399`" — **three do not** (D4). |
| Testability | 0.75 | 0.75 — most criteria binary-testable, material gaps | 16 Given-When-Then criteria, no weasel words, per-criterion mutant analysis and `Harness correction:` obligations — genuinely above the norm. Against that: I defeated 4 criteria with constructible mutants (D1, D2, D5, D7), and 2 criteria's failing inputs cannot be built at their scheduled milestone (D3). |
| Traceability | 0.85 | between 0.75 and 1.0 | All 11 REQs have ≥1 AC (001→AC-001/002/003; 002→004; 003→005/006; 004→007; 005→008; 006→009; 007→010; 008→011; 009→012; 010→013; 011→014). No orphan pointing at a non-existent REQ. Deduction: AC-HWD-015 and AC-HWD-016 trace to **constraints** (§C-1, §C-4) rather than to any `REQ-XXX`, so two [HARD] obligations have acceptance criteria with no requirement-layer representation. |

Harmonic mean = 4 / (1/0.75 + 1/0.90 + 1/0.75 + 1/0.85) = **0.807**. Below 0.85.

## Per-criterion falsifiability — my own re-runs, not the author's

Every command below was re-executed by this auditor at HEAD `4842760a7` from the worktree root.

| AC | Author's `Pre-impl observed:` | My independent re-run | Verdict |
|---|---|---|---|
| 001 | `0` | `grep -c 'chain-event.sh' .claude/settings.json` → `0` (rc=1) | **MATCH** — falsifies today |
| 002 | `entries 1 async 0 if-scoped 0` / `[]` | identical, byte-for-byte | **MATCH** — falsifies today |
| 003 | no such test; parity false | `grep -rl 'HookEntryParity' internal/` → rc=1, no match. Target state independently verified: `settings.json.tmpl:198` carries the `chain-event.sh` SubagentStop entry; `:78/:86/:94` carry three `status-transition-ownership.sh` entries with `"if": "Write(**/.moai/specs/**)"` / `"Edit(…)"` / `"MultiEdit(…)"`, each `"timeout": 5, "async": true` | **MATCH** — falsifies today. Wording nit: `grep -rc` does not print "no matching files"; it prints per-file `0` and exits 0 |
| 004 | `ls .moai/state/chain/` → `.gitkeep` only | `.moai/state/chain/` **does not exist** in this worktree. `grep -rn CreateNodeAtSpawn internal \| grep -v _test` → 1 hit (`internal/chain/populate.go:53`, the definition), **0 callers** — matches | **PARTIAL** — the conclusion holds and is well-evidenced; the quoted `ls` output is not reproducible here (D10). Separately, the criterion largely already holds |
| 005 | `grep -c 'Hook Wiring' internal/cli/doctor.go` → `0` | `0` | **MATCH** |
| 006 | same `0`, project-only entry real | confirmed via the AC-002 re-run | **MATCH** |
| 007 | `sha256 = 57fc6d11…8a7324e` | `shasum -a 256 .claude/settings.json` → `57fc6d11506a4cfd198dc4de1ecea27baa23bea9087a862adaa90e5008a7324e` | **EXACT MATCH** |
| 008 | **deferred to run time** | not measured by the author; measurable at HEAD without any implementation | **§C-5 VIOLATION** (D4) |
| 009 | `chain-event` → 0, `handle-agent-hook` → 0 | both `0`. Also measured: `handle-elicitation` 1, `handle-notification` 1, `handle-task-created` 1, `handle-worktree-create` 1, `handle-worktree-remove` 1, `team-ac-verify` **7**, `handle-session-start-compact` 0, `handle-session-start-navigator` 0 | **MATCH** on the two cited values; the surrounding prose miscounts (D8) |
| 010 | `statusLine` in rule file → **deferred**; `"type": "command"` → 34 | `grep -c 'statusLine' .claude/rules/moai/development/hook-independence.md` → **0**; `grep -c '"type": "command"' .claude/settings.json` → **34** | **§C-5 VIOLATION** — I measured the deferred value in one command; it is `0` (D4) |
| 011 | `43` / `50` | `ls .claude/hooks/moai/*.sh \| wc -l` → `43`; `ls internal/template/templates/.claude/hooks/moai/ \| wc -l` → `50` | **MATCH** on values; criterion passes pre-implementation and admits mutants (D5) |
| 012 | `SidecarUnavailable`, `EXIT=1`, no index written | **NOT RE-RUN — GAP.** Requires a constructed temp project plus a binary built from this tree (`make build`); I did not build | **UNVERIFIED** |
| 013 | `runMXColdStartScan` 5, `mxScanNeeded` 6, `mxIndexNeedsRebuild` 4 | `5`, `6`, `4` | **EXACT MATCH** on values; criterion defeated by mutant M-1 (D1) |
| 014 | `1`, at `session_start.go:253` | `grep -n` → line `253`, text verbatim *"in the background (durable side effects still land) and sends"* inside the `case <-joinTimer.C:` branch. The contradicting accurate comment confirmed at `1529-1531`: *"the SessionStart process may exit and kill the helper goroutine … the scan only lands if it finishes before the process exits"* | **EXACT MATCH** |
| 015 | the two files are IDENTICAL | `diff -q` local mirror vs template source → rc=0, no output | **MATCH** (invariant criterion; the SPEC states this openly) |
| 016 | **deferred to run time** | `grep -cE 'SPEC-[A-Z]\|REQ-[A-Z]' internal/template/templates/.claude/rules/moai/development/hook-independence.md` → **0** (rc=1). I also portability-tested the card regex: `grep -cE '\bt[0-9]{2,3}\b'` on a fixture containing `t244` and `t2440` matched **only** the `t244` line — `\b` works on this platform, the guard is sound | **§C-5 VIOLATION** — measurable in one command; it is `0` (D4) |

**Ruling on the two declared deferrals the dispatch flagged.** Neither is legitimate, and there is a **third**
the dispatch did not name.

- **AC-HWD-016** — not an evasion (the true value `0` is a clean baseline that embarrasses nothing), but
  unjustifiable: one `grep` at HEAD produces it. §C-5 is absolute and this violates it.
- **AC-HWD-008** — the most defensible of the three, since the baseline exit status is only meaningful against
  a binary built from this tree. Still avoidable: `make build` is in the plan's own §C pre-flight, so the
  baseline was reachable at authoring time.
- **AC-HWD-010** — undeclared by the dispatch and equally deferred; the value is `0`.

## Mutants I constructed

**M-1 — defeats AC-HWD-013 (violates REQ-HWD-010). Severity: critical.**
The Then clause is three greps scoped to a **single file**. Move `runMXColdStartScan` and `mxIndexNeedsRebuild`
verbatim into a new sibling file `internal/hook/mx_scan.go` (same package), rename them
`startAdvisoryIndexRefresh` / `indexStale`, and rename the `mxScanNeeded` parameter to `refreshWanted`.
Result: `grep -c 'runMXColdStartScan' internal/hook/session_start.go` → `0`,
`grep -c 'mxScanNeeded'` → `0`, `grep -c 'mxIndexNeedsRebuild'` → `0` — **all three pass** — while the scan
still fires on every SessionStart, which is exactly what REQ-HWD-010 forbids ("The SessionStart handler shall
carry no MX cold-start scan"). The author's stated mitigation ("paired with a review requirement that the
`spawnDeferredAdvisoryScans` signature no longer carries a scan parameter") does not hold: it lives in the
`Mutant:` prose, not in the Then clause, and the mutant renames that parameter. Nor does the harness
correction catch it — I confirmed `internal/hook/session_start_mx_test.go:90,129,148` calls
`runMXColdStartScan`, and under the relocation those calls still resolve, so `go test ./internal/hook/...`
stays green.

**M-2 — defeats AC-HWD-005 AND AC-HWD-006 (violates REQ-HWD-003). Severity: critical.**
Both criteria assert only what appears in `moai doctor`'s **output**. Implement `checkHookWiringDrift` with a
hardcoded in-Go list of expected `(event, script)` pairs and diff the project's `settings.json` against that
list. It names `chain-event.sh` as template-only (AC-005 passes), names the unscoped
`status-transition-ownership.sh` entry as project-only (AC-006 passes), and never renders the template.
REQ-HWD-003 explicitly requires "render the settings template in memory" — the mechanism is the whole point,
because a hardcoded list rots against the template and re-creates the very drift class this SPEC exists to
close. AC-HWD-003's parity test does perform a real render, but it is a **separate artifact**: a hardcoded
doctor check plus a correct parity test passes all 16 criteria.

**M-3 — defeats AC-HWD-007 (violates REQ-HWD-004). Severity: major.**
REQ-HWD-004 forbids modifying "`.claude/settings.json` **or any other file on disk**", but the criterion only
observes `sha256`+`mtime` of `settings.json` and a recursive **content** checksum of `.claude/`. A diagnostic
that writes a cache, a log line, or a drift report under `.moai/` (e.g. `.moai/logs/doctor-drift.log`) passes
every assertion while violating the requirement as written. Doctor writing a log is a realistic implementation
choice, not a contrived one.

**M-4 — defeats AC-HWD-011, which the SPEC declares mutant-proof. Severity: major.**
acceptance.md states *"`Mutant:` none constructible — the criterion is a direct inventory comparison with no
interpretive gap."* The gap is that "not deleted" ≠ "name still returned by `ls`". Two passing mutants:
(a) `git rm --cached <wrapper>.sh` — removed from the repository (the distributed act REQ-HWD-008 forbids)
while the working-tree file remains, so `ls` and the name set are unchanged; (b) truncate a wrapper to 0
bytes — name and count unchanged, script functionally removed. The §D.5 DoD phrases the obligation as
"No file … was **added or removed**", which `ls` on a working tree cannot establish; `git ls-files` /
`git status --porcelain` can. Both mutants require deliberate perversity, so I rank the *practical* risk
low — but the **"none constructible" claim in the record is false**, and an overclaimed mutant-proof note is
the thing that stops the next reader from strengthening the criterion.

**Criteria I attacked and could NOT defeat** — AC-HWD-003 (full tuple `(event, matcher, script, if, timeout,
async)` in both directions closes the name-only and one-directional mutants, and the [HARD] observed-failure
obligation closes the never-run-red mutant), AC-HWD-004, AC-HWD-012 (asserting tag-returned AND index-exists
closes the empty-result mutant; the stale+corrupt harness inputs close the absent-only mutant), AC-HWD-014,
AC-HWD-015, AC-HWD-016. These six are adequate as written.

## Dispatch-directed checks

**(3) The report-only requirement — REQ-HWD-004 / AC-HWD-007: SATISFIED on the specific concern.**
AC-HWD-007 binds `sha256` AND `mtime` of `.claude/settings.json` plus a recursive checksum of `.claude/`
(acceptance.md, "Then" clause), and its `Harness correction:` is [HARD] *"run the diagnostic twice in a row
against the same drifting copy and record that the second run reports the same drift as the first. A
self-repairing check is detectable only by the second run."* A repair-on-run-1 / clean-on-run-2 diagnostic is
therefore caught. The residual hole is scope, not mechanism — see M-3.

**(4) Scope discipline: PASS.** All four out-of-scope items are stated as open questions and none is decided.
§F carries them as exclusions; §G-1 routes chain node population to *"SPEC-CHAIN-CORE-001's owner decides"*
(card t242); §G-2 routes the navigator regression to *"SPEC-PROJECT-NAVIGATOR-001's owner decides
restore-or-retire"* (t243); §G-3 says *"Decide whether team mode is meant to fire it"* with a conditional
("if yes it needs a `TaskCompleted` entry"), which recommends without deciding (t244); §G-4 is explicitly
carded-nowhere and untouched. plan.md §G lists *"Deciding §G-1/2/3 in passing"* as an anti-pattern.
One caveat: M3's rewording of `team-ac-verify.sh` from "dormant" to "not registered — activation decision
pending" is a **factual correction**, not a decision, and is legitimate — but it is partial (D6).

**(5) Template-First and neutrality: PASS on substance.** spec.md §D classifies the rule file as **template
source** edited "here **first**"; plan.md §F M3 states target = `internal/template/templates/…` first,
then `make build`, then mirror; AC-HWD-015 asserts mirror identity (I verified `diff -q` → identical today);
AC-HWD-016 asserts the forbidden classes and I verified its `\b` regex works on this platform. I re-measured
the template file: `grep -cE 'SPEC-[A-Z]|REQ-[A-Z]'` → `0`, and no card number appears. plan.md §F M3
explicitly states *"The `open-question` rows name the pending decision and **not** the card"*. The claim holds.
Minor precision defect: D11.

**(6) The `moai update` prohibition: PASS.** `grep -n 'moai update' acceptance.md` → rc=1, **no match** — no
criterion invokes it. §C-3 forbids the phrasing and carries the conditional post-update deletion check
(`git status --porcelain | grep '^ D'` → must be empty) should any step ever need it. plan.md §B.3 and §G
repeat the prohibition.

**(7) M1's honesty about inertness: PASS, and this is the SPEC's strongest section.** REQ-HWD-002 requires the
record to state inertness **and** forbids claiming the entry makes the ledger work. AC-HWD-004 makes it a
criterion with the false-comment mutant named explicitly. plan.md §B.1 and §F M1's closing note repeat it for
the commit message. §G-1 states *"Until then the entry M1 delivers is parity, not function."* I independently
confirmed the premise: `CreateNodeAtSpawn` has exactly one occurrence outside tests (its definition at
`internal/chain/populate.go:53`) and no caller. No requirement anywhere promises a working ledger.

**(8) Cross-milestone ordering: COHERENT for AC-003, INCOHERENT for AC-005/007.**
The AC-003 arrangement is sound and explicitly reconciled: plan.md §A states *"AC-HWD-003's parity test cannot
be green until M1 lands. Land the test **red** with M2 … then M1 turns it green. Do not weaken the test to make
it pass early."* The AC matrix attaches AC-003 to M1 while the artifact is built in M2 — a deliberate,
documented split, not drift. I additionally verified the render is deterministic enough to support it: the
`hooks` key spans `settings.json.tmpl:3-373`, and every `{{ }}` inside that span is `.HookOptIn.Enabled`
(lines 109/120/169/176/188/195/237/244) — the `.Platform` and `.GitMode` conditionals live at 376+, outside
`hooks`. So the plan's "read `HookOptIn` from `system.yaml`" approach fully determines the hook subtree.
The incoherence is elsewhere — see D3.

## Defects Found

**D1.** AC-HWD-013 — `acceptance.md` § "AC-HWD-013 — the SessionStart cold-start scan is gone" (Then clause) —
The criterion is three greps scoped to `internal/hook/session_start.go` alone, so relocating the scan to a
sibling file in the same package and renaming its symbols passes all three while the scan still fires
(mutant M-1, constructed above; survives `go test ./internal/hook/...`). The criterion does not observe
REQ-HWD-010. — Severity: **critical** — Class: **blocking** — Required fix: scope the greps to the package
(`grep -rc … internal/hook/`), promote the `spawnDeferredAdvisoryScans` arity check from the `Mutant:` prose
into the Then clause as a mechanical assertion, and add one behavioural clause: a SessionStart handler test
that, run in a temp project with no `mx-index.json`, observes that **no** `mx-index.json` is created.

**D2.** AC-HWD-005 / AC-HWD-006 — `acceptance.md` §M2 (both Then clauses) — Both assert only doctor's printed
output, so a hardcoded expected-entry list passes both without ever rendering the template (mutant M-2). The
"render the settings template in memory" clause of REQ-HWD-003 — the clause that stops the check rotting into
the same drift class it detects — is unobserved by any criterion. — Severity: **critical** — Class:
**blocking** — Required fix: add a criterion binding the mechanism, e.g. a unit test over the shared
keyed-entry helper (the one plan.md §F M2 "Extraction note" already mandates) that feeds it a **modified
in-memory template** carrying a script name absent from any hardcoded list, and asserts the drift line names
it. A hardcoded implementation cannot pass that.

**D3.** AC-HWD-005 and AC-HWD-007 — `acceptance.md` §M2 (Given clauses) vs `plan.md` §A milestone table —
Both presuppose a tree state that only exists **after M1**, which plan.md schedules **last**. AC-HWD-005's
failing input is *"a temporary project copy whose `.claude/settings.json` has one hook entry removed (the
`chain-event.sh` SubagentStop entry)"* — but at M2 time that entry is not there to remove (I measured:
`grep -c 'chain-event.sh' .claude/settings.json` → `0`). AC-HWD-007 requires running *"both on a drift-free
copy and on a drifting copy"* — a drift-free copy does not exist until M1 lands. A workaround exists (build
the drift-free copy from the rendered template) but is nowhere stated, so the run phase will stall on it. —
Severity: **major** — Class: **blocking** — Required fix: state the construction explicitly — the drift-free
copy is produced by rendering the template, not by copying the project — or add a milestone-order note
mirroring the AC-003 red-then-green reconciliation already in plan.md §A.

**D4.** §C-5 vs acceptance.md — `spec.md` §C item 5 vs `acceptance.md` AC-HWD-008 / AC-HWD-010 / AC-HWD-016 —
§C-5 states as [HARD] that *"Each entry in `acceptance.md` carries a `Pre-impl observed:` value measured on
this tree at HEAD `950cb4399`"*. **Three entries carry "to be recorded at run time" instead.** Two of the three
are single-command measurements I performed during this audit: AC-HWD-010's `statusLine` grep is **0** and
AC-HWD-016's neutrality grep is **0**. This is the exact failure class the dispatch flags as having sunk the
prior SPEC on this branch twice. — Severity: **major** — Class: **blocking** — Required fix: record all three
values (010 → `0`, 016 → `0`, 008 → run `make build` then `bin/moai doctor; echo $?` and record the integer),
or amend §C-5 to permit a named, justified deferral class and place AC-008 in it explicitly.

**D5.** AC-HWD-011 — `acceptance.md` § "AC-HWD-011 — nothing was deleted", `Mutant:` line — The record asserts
*"none constructible — the criterion is a direct inventory comparison with no interpretive gap."* The claim is
false: the interpretive gap is that "not deleted" ≠ "name returned by `ls`". Two passing mutants exist (M-4:
`git rm --cached` leaving the working-tree file; truncation to 0 bytes). The §D.5 DoD's own phrasing
("No file … was added or removed") is a claim `ls` cannot support. — Severity: **major** — Class: **blocking**
— Required fix: replace the `ls` comparison with `git ls-files .claude/hooks/moai/ | wc -l` plus a
`git status --porcelain` emptiness check for those paths, and correct the `Mutant:` note from "none
constructible" to the two named mutants and how the revised criterion closes them.

**D6.** M3's `team-ac-verify.sh` rewording is partial — `plan.md` §F M3 ("Also correct the existing
`team-ac-verify.sh` wording") vs `.claude/rules/moai/core/agent-common-protocol-reference.md:291` and
`.claude/rules/moai/core/agent-common-protocol.md` — M3 corrects "dormant" → "not registered — activation
decision pending" in `hook-independence.md` only. I measured two further surfaces still saying "dormant":
`agent-common-protocol-reference.md:291` (*"TaskCompleted in team mode (dormant — harness `thorough` + team
prerequisites)"*, which spec.md §G-3 itself cites as evidence) and `agent-common-protocol.md` (1 occurrence,
**always-loaded**). The investigation knew: `d2-unwired-scripts.md:112` lists both files plus their template
mirrors. After M3 lands, three surfaces disagree, and the always-loaded one still carries the wrong reading.
The SPEC neither schedules them nor excludes them. — Severity: **major** — Class: **blocking** — Required fix:
either extend M3 to the two files and their template mirrors (Template-First applies to both), or add an
explicit `### Out of Scope` bullet naming them and the reason, so the divergence is a recorded decision rather
than an oversight.

**D7.** REQ-HWD-004 scope mismatch — `spec.md` §B REQ-HWD-004 vs `acceptance.md` AC-HWD-007 — The requirement
forbids modifying "any other file **on disk**"; the criterion observes only `.claude/`. A diagnostic writing a
log or cache under `.moai/` passes while violating the requirement (mutant M-3). — Severity: **major** —
Class: **blocking** — Required fix: pick one and make both agree — either narrow REQ-HWD-004 to `.claude/`
(and say why the rest of the tree is out of the observable), or widen AC-HWD-007 to a whole-worktree
`git status --porcelain` emptiness assertion before and after the run.

**D8.** AC-HWD-009 `Pre-impl observed:` miscounts — `acceptance.md` § "AC-HWD-009 — every one of the 11 carries
a disposition" — The prose says *"the three newly-classified names"* and then lists **four**
(`chain-event.sh`, `handle-agent-hook.sh`, `handle-session-start-compact.sh`,
`handle-session-start-navigator.sh`). It also accounts for only 10 of the 11: `team-ac-verify.sh` is omitted
from the already-present set, though I measured it at **7** occurrences in the rule file. Correct split, which
I measured name-by-name: **7 present** (4 obs-only + 2 worktree + `team-ac-verify`), **4 absent**. — Severity:
**minor** — Class: **blocking** (arithmetic in a measured-evidence block) — Required fix: change "three" to
"four" and add `team-ac-verify.sh` to the present set with its count.

**D9.** Two REQs carry a non-GEARS pattern label — `spec.md` §B REQ-HWD-005 and REQ-HWD-009, both tagged
`(Event-detected)` — Not one of the five GEARS pattern names. Both texts are canonical Event-driven form, so
MP-2 passes on form, but the label will be copied forward. — Severity: **minor** — Class: **optional** —
Required fix: relabel both `(Event-driven)`.

**D10.** AC-HWD-004's cited evidence line is not reproducible in this tree — `acceptance.md` § AC-HWD-004,
`Pre-impl observed:` — It reads *"`ls .moai/state/chain/` → `.gitkeep` only, 0 bytes"*; in this worktree the
directory **does not exist at all** (`ls: .moai/state/chain/: No such file or directory`). The conclusion
(`events.jsonl` has never existed) is unaffected and is independently supported, but a cited command whose
output does not reproduce is an unattributed baseline. — Severity: **minor** — Class: **blocking** (evidence
attribution) — Required fix: restate as measured here, or scope the citation to the tree where it was taken.

**D11.** §C-4's "§G only" claim is inaccurate — `spec.md` §C item 4 — It states *"the card numbers live in
**§G of this SPEC only**"*. Card tokens in fact also appear in the frontmatter (`tags: … card-t216`), in
`plan.md` §D and §F(M1), and in `progress.md` (`card: t216`, `deferred: [t242, t243, t244]`). The substantive
obligation (never in `internal/template/templates/**`) holds and is separately enforced by AC-HWD-016. —
Severity: **minor** — Class: **optional** — Required fix: reword to "…and never in
`internal/template/templates/**`".

**D12.** AC-HWD-015 / AC-HWD-016 trace to constraints, not requirements — `acceptance.md` §D AC Matrix rows
mapping to `§C-1 Template-First` and `§C-4 neutrality` — Two [HARD] obligations carry acceptance criteria with
no `REQ-XXX` in the requirement layer. — Severity: **minor** — Class: **optional** — Required fix: promote
both to `REQ-HWD-012` / `REQ-HWD-013`, or note in the matrix that constraint-derived criteria are intentional.

**D13.** §G-1/§G-2 route decisions to owners of SPECs that are already closed — `spec.md` §G-1, §G-2 — I
measured `status: completed` on both `SPEC-CHAIN-CORE-001` and `SPEC-PROJECT-NAVIGATOR-001`. "The owner
decides" has no live process attached to a completed SPEC; the actual mechanism is the amendment path
(`completed → in-progress`, per the frontmatter schema) or the cards t242/t243, which the SPEC does name. —
Severity: **minor** — Class: **optional** — Required fix: one clause in each noting that the decision lands as
an amendment to the completed SPEC, carried by its card.

## Gaps — what this audit did NOT observe

1. **AC-HWD-012 was not re-run.** Its pre-impl value (`SidecarUnavailable`, `EXIT=1`, no `mx-index.json`)
   requires a constructed temp project **and** a binary built from this tree. I did not run `make build`, so
   this is the one `Pre-impl observed:` value I take on the author's word. It is also the criterion whose
   mutant analysis I judged adequate — that judgement rests on reading the criterion, not on running it.
2. **The primary checkout was not inspected.** AC-HWD-004's claim spans *"this worktree or the primary
   checkout"*; I verified only this worktree, per the dispatch's no-sibling-tree constraint.
3. **No mutant was actually implemented.** M-1 … M-4 are constructed by reading the criteria against the code
   I read; I wrote no mutant code and ran no mutant test. M-1's survival of `go test ./internal/hook/...` is
   reasoned from the observed call sites in `session_start_mx_test.go:90,129,148` plus Go package scoping, not
   from a compiled run.
4. **`mcp__moai__spec_audit` / `spec_drift` produced no usable evidence** (no `project_root` in the schema I
   received; returned `total_specs: 0`). SPEC lifecycle drift for this tree is therefore unmeasured by the
   dedicated tool; `moai spec lint` on `spec.md` is the substitute and it is clean, but the installed binary
   is v3.1.2 and predates this tree.
5. **`go vet` / `golangci-lint` / any test run were not executed.** This is a plan-phase audit over documents;
   no build or test evidence is claimed.
6. **The three investigation reports were spot-checked, not fully re-verified.** I confirmed
   `d2-unwired-scripts.md:18-21,63-69` backs the 33-not-34 correction and `:103` backs the
   `handle-agent-hook.sh` frontmatter reclassification; I did not re-derive every figure in d1/d2/d3.

## Residual risk

The reframing the dispatch asked me to verify **held**: I found no disproved card premise carried forward into
the requirement set. §A.2's five corrections are each traceable to a report and the two I spot-checked
reproduce. §G-4's die-at-exit twin claim also survives a check the SPEC did not make — I confirmed
`internal/hook/file_changed.go:57,70,114,117` uses a `sync.WaitGroup` whose only accessor is documented
*"for use with `testutil.WaitForAsync`. Package-internal; not exposed via the Handler interface"*, with no
production waiter, so the goroutine genuinely is unwaited in production and the SPEC's "asserted from code,
not measured" label is the right honesty rather than an overclaim.

What remains risky is narrower and specific: this SPEC's quality gate is its mutant analysis, and that
analysis is **self-reported**. Four of sixteen criteria are defeatable, and one of those four is the one the
author declared mutant-proof. The pattern to watch in a revision is a criterion whose stated mitigation lives
in the `Mutant:` prose rather than in the Then clause (D1 and D2 are both that shape) — prose mitigations are
not gates, and they read as gates.

## Recommendation

FAIL at 0.81 against a 0.85 threshold, with all seven must-pass criteria PASS or N/A. This is a strong SPEC
missing the bar, not a weak one: the disproved-premises section, the report-never-repair constraint, the
`moai update` prohibition, and the M1 inertness discipline are all verified sound and are above the norm for
this repository. Seven blocking findings, in fix order:

1. **D1** — rescope AC-HWD-013's greps to the package, promote the arity check into the Then clause, add the
   behavioural no-index-written assertion.
2. **D2** — add a mechanism-binding criterion for the doctor check (modified-in-memory-template unit test over
   the shared helper) so a hardcoded allowlist cannot pass.
3. **D3** — state how the drift-free copy and AC-005's failing input are constructed before M1 lands.
4. **D4** — record the three deferred `Pre-impl observed:` values (010 → `0`, 016 → `0`, 008 via
   `make build` + `bin/moai doctor; echo $?`), or carve a named deferral class into §C-5.
5. **D7** — reconcile REQ-HWD-004's "any other file on disk" with AC-HWD-007's `.claude/`-scoped observable.
6. **D5** — replace AC-HWD-011's `ls` comparison with a `git ls-files` + `git status --porcelain` form and
   correct the false "none constructible" note.
7. **D6** — scope in or explicitly exclude the two remaining "dormant" surfaces (`agent-common-protocol.md`,
   `agent-common-protocol-reference.md:291`, plus template mirrors).

D8 and D10 are cheap corrections to measured-evidence blocks and should ride along. D9, D11, D12, D13 are
optional and left to the orchestrator's discretion — routing all four into a revision would add requirement
surface this SPEC has not claimed and is not the shape of the problem.

Iteration 2 should be scoped to this enumerated defect delta plus a regression check, not a from-scratch
re-audit. Tier M ceiling is 2 iterations.
