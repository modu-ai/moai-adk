# SPEC Review Report: SPEC-HOOK-WIRING-DRIFT-001

Iteration: 2/2 (Tier M ceiling — terminal)
Verdict: **PASS**
Overall Score: **0.862** (harmonic mean; arithmetic 0.868)
Threshold applied: **0.80** (Tier M, SSOT). Also clears 0.85 — see § Threshold note.
Audited at: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t216`, HEAD `6331d505c`, branch `WT-hook-wiring-drift`

Reasoning context ignored per M1 Context Isolation. Position confirmed before any read:
`git rev-parse --show-toplevel` → `.../worktrees/t216`; `git rev-parse --short HEAD` → `6331d505c`;
`git branch --show-current` → `WT-hook-wiring-drift`; working tree carries one untracked file
(`?? .moai/reports/t216/plan-audit.md`, iteration 1's own report). The delta since iteration 1 is a single
commit `6331d505c` touching **only** the four SPEC artifacts (`--stat`: +618/-308, no source changes), so
every `Pre-impl observed:` value remains independently re-runnable at this HEAD.

## Threshold note (process finding, not a SPEC defect)

The dispatch instructed me to score against **0.85**. That is the **Tier L** value. This SPEC declares
`tier: M` (spec.md:14), and the SSOT — `.claude/rules/moai/workflow/spec-workflow.md:138-142`,
§ SPEC Complexity Tier, column *plan-auditor PASS threshold* — reads:

| Tier | plan-auditor PASS threshold |
|---|---|
| S | 0.75 |
| M | **0.80** |
| L | 0.85 |

My agent contract directs me to read this table rather than a copy. Two consequences, stated plainly:

1. **This iteration's verdict is insensitive to the discrepancy.** 0.862 clears both 0.80 and 0.85, so
   PASS holds on either reading. The finding is informative, not load-bearing.
2. **Iteration 1's verdict was not.** It scored 0.807 and returned FAIL against 0.85. Against the correct
   Tier M threshold of 0.80, **0.807 was a PASS**. Iteration 1's findings were real and the v0.2.0 revision
   materially improved the document — but the FAIL verdict itself rested on a misapplied threshold. I record
   this against my own prior instance rather than leaving it in the file.

## Tooling note (mandatory disclosure)

`mcp__moai__spec_audit` / `spec_progress` as exposed to this session carry **no `project_root` field** and
return the primary checkout's catalog (627 SPECs, 0 matches for this SPEC). They are **not** used as evidence
anywhere below. Substitute: `./bin/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md` on the
**tree-built** binary → `✓ No findings — all SPEC documents are valid`. This is an improvement on iteration 1,
which could only run the stale globally-installed v3.1.2.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-HWD-001` … `REQ-HWD-014`, enumerated by
  `grep -n '^- \*\*REQ-HWD-' spec.md` at lines 122/126/133/139/159/166/172/178/184/188/192/198/211/215.
  Sequential, no gap, no duplicate, uniform 3-digit padding. The three additions this iteration
  (012/013/014) extend the sequence cleanly.
- **[PASS] MP-2 GEARS format compliance** — judgment made against the **requirement layer** (`REQ-XXX` in
  spec.md §B), never against an AC. All 14 carry a canonical pattern label and canonical form:
  Ubiquitous (001, 002, 006, 007, 010, 012, 013), Event-driven (003, 005, 009), Unwanted (004, 008, 011, 014).
  Iteration 1's D9 (`(Event-detected)`, not a GEARS pattern name) is **closed** — grep confirms no
  `Event-detected` label survives.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (spec.md:1-15): `version: "0.2.0"` quoted semver; `status: draft` ∈ the 8-value enum; `priority: P2`;
  `created`/`updated` ISO dates; `lifecycle: spec-anchored`; `tags` comma-separated. No rejected snake_case
  alias. `tier: M` present as an optional field. `moai spec lint` clean on the tree-built binary.
- **[N/A] MP-4 language neutrality** — the SPEC targets this repository's own hook/settings tooling; no
  per-language tooling is named. N/A auto-passes. (Template *neutrality* — a different obligation — is now
  bound by REQ-HWD-014/AC-HWD-016 and verified below.)
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — extracted references: `SPEC-CHAIN-CORE-001`,
  `SPEC-PROJECT-NAVIGATOR-001`. Both directories exist; `grep -m1 '^status:'` → `completed` on both, outside
  `{retired, superseded, archived}`. No BLOCKING finding. Improved this iteration: §G-1/§G-2 now name the
  amendment path for a completed SPEC (iteration 1's D13).
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/`
  → rc=1, no match. `plan.md` clean; `research.md` absent (Tier M does not require it).

No must-pass failure.

## Category Scores (rubric-anchored, re-justified — not inherited)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 band, one instance | REQ-HWD-004's scope mismatch (iteration 1's D7) is **resolved**: the requirement is now scoped to "under the project root" with an explicit rationale block (spec.md:139-157), and requirement and observable agree in wording. REQ-HWD-002 remains compound but is one coherent obligation. **One fresh instance of the same shape**: REQ-HWD-012 (spec.md:198-204) names **six** surfaces (three files + their template twins) while AC-HWD-018 observes **four** — finding N2. One requirement carrying a scope disagreement, down from two. |
| Completeness | 0.92 | 1.0 band minus evidence-block defects | HISTORY (both versions), §A Context (+§A.2 disproved premises), §B Requirements, §C Constraints, §D file classification (now 11 rows incl. the two new template pairs), §E NFR, §F Exclusions (four `### Out of Scope — <topic>` H3s, each with specific `-` bullets), §G follow-ups, §H cross-refs; acceptance.md §D matrix + §D.5 DoD + **new §D.6 Residual risk**. Iteration 1's D4 §C-5 shortfall is **fully closed**: all three formerly-deferred baselines are measured and I reproduced all three independently (below). Deduction: two `Pre-impl observed:` values do not reproduce at this HEAD (N4, N5). |
| Testability | 0.80 | above the 0.75 band, not clean | 16 criteria, no weasel words, every criterion mechanically decidable. Genuine, verified depth gains: AC-HWD-013 now carries a **behavioural** falsifier I executed (the inverted-test claim reproduces exactly); AC-HWD-017 is a real mechanism gate backed by a [HARD] injectable-parameter constraint at plan.md:133; AC-HWD-011 moved to git-state assertions with three named fresh mutants; AC-HWD-015 widened to three files. Defective criteria fell 6 → 2. Against that: **AC-HWD-007 is unsatisfiable as written** (N1, critical, empirically demonstrated) and **AC-HWD-009 is defeated by a mutant I constructed and tested** (N3). |
| Traceability | 0.95 | 1.0 band, small deduction | Iteration 1's D12 is **closed**. Matrix verified row-by-row: 16 live ACs, every one mapping to a `REQ-XXX`; every one of the 14 REQs covered (001→003; 002→004; 003→005/006/017; 004→007; 005→008; 006→009; 007→010; 008→011; 009→012; 010→013; 011→014; 012→018; 013→015; 014→016). No orphan, no uncovered REQ. The two retired rows are retained with `—` so iteration-1 references still resolve — deliberate and documented. Deduction: REQ-HWD-012's criterion does not span the requirement's own stated surface set (N2) — coverage depth, not broken linkage. |

Harmonic mean = 4 / (1/0.80 + 1/0.92 + 1/0.80 + 1/0.95) = 4 / 4.6396 = **0.862**.

Sensitivity: the verdict does not turn on the Testability figure. Testability would have to fall to ≈0.62 —
the "several ACs require judgment calls" band, plainly false for this document — before the harmonic mean
dropped below 0.80.

## Falsifiability re-runs — mine, at HEAD `6331d505c`

Every command below was executed by this auditor from the worktree root, on the tree-built `./bin/moai`
where a binary was required.

| AC | Author's `Pre-impl observed:` | My independent re-run | Verdict |
|---|---|---|---|
| 003 (a) | `chain-event.sh` → `0`; parity test absent | `grep -rl 'HookEntryParity' internal/` → rc=1; `grep -c 'chain-event.sh' .claude/settings.json` → `0` | **MATCH** |
| 003 (b) | `entries 1 async 0 if-scoped 0` | unchanged since iteration 1; template target confirmed at `settings.json.tmpl:78/86/94` | **MATCH** |
| 004 | `ls .moai/state/chain/` → No such file or directory; `CreateNodeAtSpawn` 1 hit 0 callers | `grep -rn CreateNodeAtSpawn internal \| grep -v _test` → 2 lines, both `internal/chain/populate.go` (:42 doc comment, :53 definition), **0 callers** | **MATCH** — D10 correction verified |
| 005 / 006 | `bin/moai doctor 2>&1 \| grep -ci 'hook wiring'` → `0`; `grep -c 'Hook Wiring' doctor.go` → `0` | both `0` on the tree-built binary | **MATCH** |
| 007 | sha256 `57fc6d11…8a7324e`; porcelain one line, unchanged after doctor | `shasum -a 256` → `57fc6d11506a4cfd198dc4de1ecea27baa23bea9087a862adaa90e5008a7324e`; porcelain → `?? .moai/reports/t216/plan-audit.md`, unchanged after `./bin/moai doctor` | **MATCH on the worktree** — but see **N1**: the criterion's own Given is a *fixture*, and it fails there |
| 008 | `make build` then `./bin/moai doctor; echo $?` → **EXIT=0** | `./bin/moai doctor > /dev/null 2>&1; echo $?` → **`0`**; `grep -ci 'hook wiring'` → `0` | **MATCH** — D4 deferral closed |
| 009 | 7 present / 4 absent; per-name integers incl. `handle-elicitation` **2** | present/absent split **7 / 4 — MATCH**. `team-ac-verify` **7** — MATCH. `chain-event` `handle-agent-hook` `handle-session-start-compact` `handle-session-start-navigator` all **0** — MATCH. **`grep -c handle-elicitation` → `1`, not 2** | **PARTIAL** — split correct, one integer wrong (**N4**) |
| 010 | `statusLine` → **0**; `handle-agent-hook` → **0**; `"type": "command"` → 34 | `0`, `0`, `34` | **MATCH** — D4 deferral closed |
| 011 | `git ls-files .claude/hooks/moai/` → **43**; template → **47**; porcelain empty both; `-size 0` → 0 | `43`, `47`, both porcelain empty, `0` | **EXACT MATCH** |
| 011 miscount | v0.1.0's `50` came from an aliased long-format `ls` | **Reproduced and root-caused.** `ls … \| wc -l` → **50**; `ls -1 … \| wc -l` → **49**; `diff` of the listing against `git ls-files` shows exactly the extra entries `.` and `..`. So 47 tracked + `.` + `..` + the `total` line = 50. The author's explanation is **exactly right** | **MATCH — verified, not accepted** |
| 012 | `SidecarUnavailable`, `EXIT=1`, no index written — measured on the **v3.1.2** binary; re-run named a gap in §D.6 | **Gap closed by me.** Constructed a fixture project (`.moai/state/` empty, one `@MX:DEBT` tag) and ran the **tree-built** `bin/moai mx query --kind DEBT`: stderr `SidecarUnavailable: sidecar index does not exist — run 'moai mx scan' to build the index`; TUI `ERROR Sidecarunavailable: no sidecar index.`; **EXIT=1**; `.moai/state/mx-index.json` absent afterwards | **EXACT MATCH on the tree-built binary** — iteration 1 Gap #1 and §D.6 item 2 both closed |
| 013 (a) | package-scoped: `runMXColdStartScan` **8**, `mxIndexNeedsRebuild` **12**, `mxScanNeeded` **6** | `8`, `12`, `6` | **EXACT MATCH** |
| 013 (b) | `session_start.go:554-560` declares **5** params, fifth `mxScanNeeded bool` | confirmed verbatim: `func (h *sessionStartHandler) spawnDeferredAdvisoryScans(` at 554, params 555-559, `mxScanNeeded bool` at 559, `) <-chan …` at 560 | **EXACT MATCH** |
| 013 (c) | the existing integration test **passes today** asserting the index IS created | **Executed.** `go test ./internal/hook/ -run TestSessionStartMXColdStartIntegration -v -count=1` → `INFO session start (deferred): MX index built via cold-start scan … tags=1` / `--- PASS (0.00s)` / `ok … 0.667s` | **EXACT MATCH** — the inverted assertion genuinely fails pre-implementation |
| 013 harness | tests at `session_start_mx_test.go:83,121,135,159` | confirmed: 83/121/135 are `TestRunMXColdStartScan_{BuildsIndex,WriteErrorFailOpen,TinyTimeoutFailOpen}` declarations, 159 is `TestSessionStartMXColdStartIntegration`. Call sites are 90/129/148 | **MATCH** (labelling nit → N7) |
| 014 | `1`, at `session_start.go:253` | `grep -c 'durable side effects still land'` → `1` | **EXACT MATCH** |
| 015 | mirrors identical | `diff -q` for **all three** M3 pairs (`hook-independence.md`, `agent-common-protocol.md`, `agent-common-protocol-reference.md`) → rc=0 each | **MATCH** — the widened three-file form has a verified baseline |
| 016 | `SPEC-[A-Z]\|REQ-[A-Z]` → **0**; `\bt[0-9]{2,3}\b` → **0** | `0`, `0` | **MATCH** — D4 deferral closed |
| 017 | `Hook Wiring` → 0; `HookEntryParity` → rc=1; `grep -rl 'zz-fixture-only' .` → **rc=1, no match** | first two **MATCH**. Third: `grep -rl 'zz-fixture-only' .` → **rc=0**, returns `.moai/specs/SPEC-HOOK-WIRING-DRIFT-001/acceptance.md` | **DOES NOT REPRODUCE** (**N5**) |
| 018 | `grep -c 'dormant'` → **1** in each of the four surfaces; lines `:38` and `:291` | `1` / `1` / `1` / `1`. Both quoted lines verbatim: `agent-common-protocol.md:38` and `agent-common-protocol-reference.md:291` | **EXACT MATCH** |

**Baseline-attribution.** All values above were measured in this run, against this tree, at HEAD `6331d505c`,
from `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t216`. No figure is carried over from iteration 1; where
a value coincides with iteration 1's, it was re-executed rather than recalled.

## The v0.2.0 claims the dispatch asked me to test

### The AC-001/AC-002 retirement — subsumption **HOLDS**

I retrieved the v0.1.0 text (`git show 4842760a7:…/acceptance.md`) and compared clause by clause.

- **AC-HWD-001** asserted `grep -c 'chain-event.sh' .claude/settings.json` → `1`. **AC-HWD-003 (a)** asserts
  the **parsed** `SubagentStop` group contains an entry whose script is `chain-event.sh`, with `timeout: 5`,
  `type: command`, and no `async` key. That is strictly stronger on every axis: it asserts the containing
  event group, the transport type, the timeout, and the absence of `async` — none of which a name grep
  observes. AC-001's own `Mutant:` note conceded it was defeatable ("a comment … containing `chain-event.sh`
  anywhere in the file passes this grep while wiring nothing"), so the retirement removes a criterion that
  was **self-declared insufficient**.
- **AC-HWD-002** asserted `entries 3 async 3 if-scoped 3` plus the three exact `if` values.
  **AC-HWD-003 (b)** asserts the count is exactly 3, **each** with `async: true` **and `timeout: 5`**, and the
  `if` values are exactly those three. "Each with `async: true`" subsumes `async 3`; "their `if` values are
  exactly {…}" subsumes `if-scoped 3`; `timeout: 5` is **added**.

**Ruling: this is not depth traded for headroom.** The merged criterion observes strictly more than the two it
replaced. The ceiling arithmetic is also correct — `spec-workflow.md:146-152` sets the Tier M ceilings at 16
requirements **and** 16 criteria independently, so 14 REQs / 16 ACs is within both, and adding AC-017 + AC-018
without the retirement would have hit 18 > 16.

One observation, not a finding: the retirement leaves M1 with no criterion runnable as a plain command at
M1 time — AC-003's artifact is a Go test built in M2. plan.md §A already reconciles this explicitly
("land the test **red** with M2 … then M1 turns it green. Do not weaken the test to make it pass early"),
and AC-003's `Pre-impl observed:` preserves the two retired commands' values as sub-evidence. Coherent.

### D1's three clauses — **CLOSED, and (c) is the strongest thing in the document**

I ran the test. `TestSessionStartMXColdStartIntegration` passes today, logging `MX index built via
cold-start scan … tags=1` — so the inverted form ("after `Handle` returns, `.moai/state/mx-index.json` does
**not** exist") fails on the pre-implementation tree by construction. That is a genuine behavioural falsifier,
not a lexical one, and it defeats both fresh mutants the author names: relocating to another package leaves
the index written, and inlining the scan body leaves the index written. Combined with the package-scoped
greps (verified 8/12/6) and the arity clause promoted into the Then clause (verified 5 params), **mutant M-1
is dead**.

### D2's fix — **CLOSED**

`plan.md:133` carries the only `[HARD]` in the file: *"The template source is an injectable parameter … the
check MUST take its template source — an `fs.FS`, or the rendered bytes — as a parameter rather than reaching
for the embedded FS internally, and production MUST pass `template.EmbeddedTemplates()` through that same
seam."* The load-bearing detail is right: the seam is on the **check**, not only the helper, and AC-017's Then
observes the **check's own returned message**. A hardcoded expected-entry list cannot name `zz-fixture-only.sh`,
a name it has never seen. M-2 is closed.

### D3's fix — **CLOSED**

The `[HARD] Fixture construction` note (acceptance.md, head of §M2) states all three fixtures are built
**from the render**, never by copying the project, and spells out why: at M2 time
`grep -c 'chain-event.sh' .claude/settings.json` → `0`, so there is nothing in the project to remove. The
ordering incoherence is resolved.

### D4 — **CLOSED**, all three reproduced

No deferral class was carved; §C-5 stays absolute. I independently re-measured all three: AC-010 `statusLine`
→ `0`, AC-016 neutrality → `0`/`0`, AC-008 `./bin/moai doctor; echo $?` → `0`.

### D5 — **CLOSED**, and the latent miscount is real

43 / 47 verified. The `50` → `47` correction is not just asserted — I reproduced the wrong number
(`ls … | wc -l` → 50) and confirmed the stated cause by diffing the listing against `git ls-files`: the extra
entries are exactly `.`, `..`, plus the `total` line from the aliased long format. The false "none
constructible" claim is retracted and replaced with three named fresh mutants, all of which genuinely fail
against the new clauses.

### D6 — **PARTIALLY CLOSED** → finding N2

Scoping in rather than excluding was the right call, and the always-loaded surface
(`agent-common-protocol.md`) is now covered. But the fix is incomplete in the inverse direction — see N2.

### D7 — **CLOSED on the mismatch; the fix introduced N1**

Requirement and observable now agree in wording, and the residual out-of-tree gap is recorded openly in
§D.6 rather than papered over. But the widened observable is unsatisfiable against its own Given — see N1.

### The acknowledged uncovered boundary (`~/.moai/`) — **legitimate scoping, not a shrunk requirement**

I judge the narrowing sound, on three grounds. (1) The alternative — keeping "any other file on disk" — is an
obligation **no criterion can observe**, and an unobservable clause in a requirement set is precisely the
"reads as a gate but is not one" defect this SPEC's own method targets. (2) The narrowing is *disclosed at
both ends*: spec.md:145-153 explains it inside the requirement, and §D.6 item 1 records the residual gap as
residual risk. (3) The requirement did not lose the behaviour that motivated it — the realistic mutant
(`.moai/logs/doctor-drift.log`) is now **caught**, where under v0.1.0 it passed. A requirement shrunk to fit
its criterion would have dropped the motivating case; this one added it. Verdict: legitimate.

### AC-009's substring counts — the 7/4 split **holds**, one integer **does not**

The dispatch asked specifically whether the 7/4 split survives the substring imprecision. It does — I measured
name-by-name: present = `handle-elicitation`, `handle-elicitation-result`, `handle-notification`,
`handle-task-created`, `handle-worktree-create`, `handle-worktree-remove`, `team-ac-verify` (**7**);
absent = `chain-event`, `handle-agent-hook`, `handle-session-start-compact`,
`handle-session-start-navigator` (**4**). But the per-name integer is wrong and its stated explanation is a
misdiagnosis — see N4.

## Mutants I constructed this iteration

**N1-M — defeats AC-HWD-007 by making it unsatisfiable. Severity: critical. CONSTRUCTED AND EXECUTED.**

AC-HWD-007's snapshot (iii) requires "a listing of name+size+mtime for every file under `.moai/logs/` and
`.moai/state/`" to be **unchanged** after `bin/moai doctor` runs against the fixture. `moai doctor` writes
`.moai/state/config-cache.json` on its first run in a project. Demonstrated:

```
$ mkdir -p /tmp/t216fx/.moai/state /tmp/t216fx/.claude
$ cp .claude/settings.json /tmp/t216fx/.claude/settings.json
$ cd /tmp/t216fx && find .moai/state -type f | sort > before.txt   # 0 lines
$ .../t216/bin/moai doctor > /dev/null 2>&1; echo $?
0
$ find .moai/state -type f | sort > after.txt; diff before.txt after.txt
0a1
> .moai/state/config-cache.json
```

I controlled for the obvious confounder. In the **worktree** (where a config-cache already exists), a
single-invocation snapshot → `doctor` → snapshot shows **no** delta (`SAME_INVOCATION_RC=0`); an earlier
apparent delta was my own session's trace log crossing a tool-call boundary, and a no-doctor control over the
same interval showed no change. So the write is real, is `doctor`'s, and is **first-run only** — a second run
leaves `config-cache.json` byte- and mtime-identical (`8334 1787582060` both times).

That last detail is what makes this a blocker rather than a curiosity: AC-007's `Harness correction:` runs
doctor **twice** and would pass, while the Then clause — compared against a snapshot taken before the
**first** run on a **freshly constructed** fixture, which the `[HARD]` fixture note mandates — **fails**,
regardless of how the hook-wiring check is implemented. The criterion cannot be satisfied by correct code.

Note the irony that pins the diagnosis: the author's own `Fresh mutant attempt (iter 1)` (ii) reads *"Write a
cache under `.moai/state/` — **fails** on snapshot (iii)."* `moai doctor` already writes a cache under
`.moai/state/`. The requirement is about the **diagnostic**; the criterion observes the **whole doctor run**.

**N2-M — defeats REQ-HWD-012 through the AC-HWD-018 / AC-HWD-009 / AC-HWD-010 gap. Severity: major.**

REQ-HWD-012 (spec.md:198-204) names **six** surfaces: `hook-independence.md`, `agent-common-protocol.md`,
`agent-common-protocol-reference.md`, "and their template twins". AC-HWD-018's Given names **four** —
`agent-common-protocol.md`, `agent-common-protocol-reference.md`, and *their two* twins.
`hook-independence.md` and its twin are outside every clause of AC-018.

Do AC-009 / AC-010 cover them? **No.** AC-009 requires each of the 11 names to appear on a line carrying a
disposition class; AC-010 requires the 33-entry and agent-frontmatter statements. Neither touches the
"dormant" wording. And `hook-independence.md` carries **5** occurrences of `dormant`, including exactly the
registered-but-gated reading REQ-HWD-012 forbids:

```
:87  | (g) configured timeout | … | 10s (TaskCompleted) — **dormant / not wired** |
:89  > **Row (g) caveat — `team-ac-verify.sh` is dormant.** …
:91  > It is forward-looking (activates only under harness `thorough` + team mode prerequisites).
```

Mutant: correct the four AC-018 surfaces; add a disposition row `team-ac-verify.sh | open-question` to
satisfy AC-009; record the two corrections for AC-010; **leave lines 87-93 untouched**. All 16 criteria pass;
REQ-HWD-012 is violated on 2 of its 6 named surfaces, one of which is distributed to 16 languages.

AC-HWD-018's `Fresh mutant attempt (iter 1)` asserts precisely this mutant fails: *"(i) Correct all four files
but leave `hook-independence.md` on the old wording — … fails AC-HWD-009 + AC-HWD-010, which bind the same
correction in `hook-independence.md`."* **That claim is false.** It is the same overclaim class as v0.1.0's
"none constructible" (D5) — an unverified closure note recorded as a closure.

A caution for the fix: lines 96/106/107 of the same file use "dormant" correctly, about *other* surfaces
(worktree lifecycle wrappers, `moai hook spec-status`). A blanket removal would be wrong; the correction is
line-specific.

**N3-M — defeats AC-HWD-009. Severity: major. CONSTRUCTED AND EXECUTED.**

AC-HWD-009's Then clause: *"each appears at least once, on a line carrying one of the five disposition
classes."* One line can carry all 11 names. I built the mutant and ran the criterion against it:

```
$ cat /tmp/t216mut/mutant-rule.md
| `chain-event.sh`, `handle-agent-hook.sh`, … , `team-ac-verify.sh` | dead-by-decision |

$ grep -o -e chain-event.sh -e … -e team-ac-verify.sh mutant-rule.md | sort -u | wc -l
11                     # all 11 names present
$ grep -cE 'reachable-via-template-settings|reachable-via-agent-frontmatter|\
reachable-via-in-binary-registry|dead-by-decision|open-question' mutant-rule.md
1                      # and that one line carries a disposition class
```

**The criterion passes.** Yet the mutant assigns `dead-by-decision` to `chain-event.sh` (which M1 makes
`reachable-via-template-settings`), to `handle-agent-hook.sh` (`reachable-via-agent-frontmatter` — §A.2's own
correction), and to the two open questions — erasing exactly the distinction the investigation established.

This is the blanket-heading mutant AC-009's `Mutant:` note claims to close: *"Closed by requiring a
disposition class **per name** … on the same line."* The Then clause does not say "per name"; it says each
name appears on *a* line carrying *one of* the five classes. The mitigation lives in the `Mutant:` prose,
not in the gate — the identical shape as iteration 1's D1 and D2. AC-009 was **not** rewritten this iteration
and so received no `Fresh mutant attempt`, which is how it survived.

A realistic (non-contrived) instance: grouping the four retired obs-only events **plus** `chain-event.sh` on
one `dead-by-decision` line is a plausible authoring shortcut, and it records a wrong disposition while
passing.

**Mutants I attempted and could NOT construct** — AC-HWD-003 (full tuple, both directions, plus the [HARD]
observed-failure obligation), AC-HWD-004, AC-HWD-011 (git-state clauses close both M-4 variants; the count and
porcelain clauses are genuinely non-redundant as the author argues), AC-HWD-012 (tag-returned AND index-exists
closes the empty-result mutant; the stale and corrupt harness inputs close the absent-only mutant),
AC-HWD-013 (see above — behaviourally closed), AC-HWD-014, AC-HWD-015, AC-HWD-016, AC-HWD-017. Nine criteria
resisted attack.

**An attack that failed, reported because it failed.** I hypothesised that `moai doctor` mutates
`.moai/logs/` in the worktree and would break AC-007 there too. A controlled single-invocation test
(`SAME_INVOCATION_RC=0`) plus a no-doctor control disproved it. The apparent delta was my own session's trace
file. AC-007's snapshot (iii) is sound **in the worktree**; it is only the fixture path that breaks.

## Regression Check — iteration 1 defects

| # | Status | Evidence |
|---|---|---|
| **D1** AC-013 single-file greps | **CLOSED** | (a) package-scoped 8/12/6 re-measured; (b) arity 5 params at `session_start.go:554-560` verified, clause promoted into the Then; (c) inverted-test falsifier **executed** — currently PASSes asserting the index IS created. M-1 dead. |
| **D2** mechanism unbound | **CLOSED** | New AC-HWD-017 + `[HARD]` injectable-parameter constraint at `plan.md:133`; seam on the check, Then observes the check's returned message. `zz-fixture-only` confirmed absent from all source. |
| **D3** fixtures unconstructible pre-M1 | **CLOSED** | `[HARD]` fixture-construction note; all three fixtures built from the render. |
| **D4** three deferred baselines | **CLOSED** | All three measured; all three reproduced by me (`0`, `0`, exit `0`). No deferral class carved; §C-5 remains absolute. |
| **D5** AC-011 false "none constructible" | **CLOSED** | git-state clauses verified 43/47/empty/0; claim retracted; three fresh mutants named and each genuinely fails. The 50→47 miscount reproduced and root-caused. |
| **D6** remaining "dormant" surfaces | **PARTIALLY CLOSED** | The two named surfaces + twins are now scoped in via REQ-HWD-012/AC-HWD-018 (`dormant` = 1 in each, verified). **Residue**: the inverse gap — `hook-independence.md` (5 occurrences) and its twin have no criterion binding the corrected reading → **N2**. |
| **D7** REQ-004 scope mismatch | **CLOSED** (fix introduced N1) | Requirement narrowed to project root with rationale; AC-007 widened to porcelain + gitignored snapshots. Wording and observable now agree. The widened observable is unsatisfiable → **N1**. |
| **D8** AC-009 miscount | **PARTIALLY CLOSED** | 7/4 split corrected and independently verified; `team-ac-verify` added at 7; "three"→"four" corrected. **Residue**: new wrong integer `handle-elicitation = 2` → **N4**. |
| **D9** `(Event-detected)` label | **CLOSED** | grep confirms no `Event-detected` survives; REQ-005/009 both `(Event-driven)`. |
| **D10** AC-004 `ls .moai/state/chain/` | **CLOSED** | Restated as measured in this worktree with the correct explanation (the v0.1.0 listing was taken in the primary checkout); the load-bearing premise re-verified tree-independently. |
| **D11** §C-4 "§G only" | **CLOSED in spec.md** | §C-4 now reads "…appear in this SPEC's own artifacts (§G, frontmatter `tags`, plan.md, progress.md) and **never in `internal/template/templates/**`**". **The same false claim reappears at `plan.md:267`** → **N6**. |
| **D12** constraint-orphaned ACs | **CLOSED** | REQ-HWD-013/014 promoted; matrix verified — all 16 ACs trace to a REQ, all 14 REQs covered. |
| **D13** closed-SPEC owners | **CLOSED** | §G-1 and §G-2 now name the amendment path (`completed → in-progress`) carried by t242/t243. Both SPECs re-confirmed `status: completed`. |

**No iteration-1 defect is unresolved as stated.** The automatic-FAIL-on-unresolved-defect trigger does not
fire. Three findings below are **new**, two of them arising inside iteration-1 fixes.

## Regression check — items that PASSED in iteration 1

- **Scope discipline: PASS, improved.** t242/t243/t244 remain routed as undecided owner decisions
  (`spec.md` §G-1/2/3, §F "Out of Scope — the three owner decisions"); `progress.md:30` still lists them as
  `deferred`. D13's amendment-path clause is a *routing* correction, not a decision. AC-HWD-018 states
  explicitly that it is "a **factual correction, not a decision**: whether team mode should fire the hook
  remains §G-3 / card t244 and is untouched." No decision leaked.
- **Template-First and neutrality: PASS, strengthened.** Now requirement-backed (REQ-HWD-013/014, previously
  constraints only). `diff -q` verified identical for **all three** M3 mirror pairs; neutrality baselines
  `0`/`0`; §D file classification expanded to cover the two new template pairs including the always-loaded
  one.
- **`moai update` prohibition: PASS.** `grep -n 'moai update' acceptance.md` → rc=1, **no match**. §C-3 intact
  with the CLAUDE.local.md §2.3 post-update deletion check as the conditional guard.
- **M1 inertness framing: PASS, unchanged.** REQ-HWD-002 + AC-HWD-004 intact; §G-1's "Until then the entry M1
  delivers is parity, not function" intact. Premise re-verified: `CreateNodeAtSpawn` has 2 non-test
  occurrences, both in `internal/chain/populate.go` (doc comment :42, definition :53) — **0 callers**.

## Defects Found (structured defect-list)

**N1.** AC-HWD-007 is unsatisfiable as written — `acceptance.md` § "AC-HWD-007 — the diagnostic changes
nothing under the project root" (Then clause, snapshot (iii)) — `bin/moai doctor` writes
`.moai/state/config-cache.json` on its first run in a project, so snapshot (iii) changes against the freshly
constructed fixture the `[HARD]` fixture note mandates, regardless of how the hook-wiring check is
implemented. Demonstrated empirically above, with the confounder controlled. REQ-HWD-004 constrains the
**diagnostic**; the criterion observes the **entire doctor run**. — Severity: **critical** — Class:
**blocking** — Required fix: bind the assertion to what the requirement actually constrains. Either (a) scope
snapshot (iii) to exclude paths written by pre-existing `doctor` machinery, naming `config-cache.json`
explicitly and saying why; or (b) warm the fixture — run `bin/moai doctor` once, *then* take the pre-run
snapshot — and state that as part of the fixture construction; or (c) assert against the check function's own
effects rather than the whole run. Option (b) is closest to the requirement's intent and costs one line in
the fixture note.

**N2.** REQ-HWD-012 names six surfaces; AC-HWD-018 observes four — `spec.md` §B REQ-HWD-012 (spec.md:198-204)
vs `acceptance.md` § "AC-HWD-018" (Given) — `hook-independence.md` and its template twin carry the
registered-but-gated "dormant" reading at `.claude/rules/moai/development/hook-independence.md:87` and the
`:89-93` caveat block ("forward-looking (activates only under harness `thorough` + team mode
prerequisites)"), and **no criterion binds their correction**: AC-009 binds disposition classes, AC-010 binds
the two audit corrections, neither touches the wording. Mutant constructed above. AC-018's
`Fresh mutant attempt (iter 1)` claims this mutant fails — that claim is false, the same overclaim class as
v0.1.0's "none constructible". — Severity: **major** — Class: **blocking** — Required fix: extend AC-HWD-018's
Given to all six surfaces REQ-HWD-012 names, and correct the fresh-mutant note to record that AC-009/AC-010
do **not** close the `hook-independence.md` case. Add a caution that `hook-independence.md:96,106,107` use
"dormant" correctly about other surfaces, so the edit is line-specific.

**N3.** AC-HWD-009 is defeated by a blanket-line mutant — `acceptance.md` § "AC-HWD-009 — every one of the 11
carries a disposition" (Then clause) — "each appears at least once, on a line carrying one of the five
disposition classes" is satisfied by a single line carrying all 11 names and one class. Constructed and
executed above: 11 distinct names present, 1 disposition-class line, criterion passes while assigning a wrong
class to at least three names. The "per name" mitigation lives in the `Mutant:` prose, not the Then clause —
the same shape as iteration 1's D1 and D2. AC-009 was not rewritten this iteration and so carries no
`Fresh mutant attempt`. — Severity: **major** — Class: **blocking** — Required fix: restate the Then clause as
one row per name — e.g. "each of the 11 appears on a line that names **that script and no other**, carrying
exactly one of the five disposition classes" — and add a `Fresh mutant attempt` recording the blanket-line
mutant and how the rewrite closes it.

**N4.** AC-HWD-009's `handle-elicitation` baseline does not reproduce, and its stated explanation is a
misdiagnosis — `acceptance.md` § AC-HWD-009 `Pre-impl observed:` and `§D.6` item 4 — The record says
`handle-elicitation` **2** "(the count includes `handle-elicitation-result` as a substring)". Measured:
`grep -c handle-elicitation .claude/rules/moai/development/hook-independence.md` → **1**. `grep -c` counts
matching **lines**, not occurrences, and both names sit on the same line (`:102`), so a same-line substring
inflates nothing. The 7/4 split is unaffected and I verified it independently. — Severity: **minor** —
Class: **blocking** (measured-evidence block; §C-5 is [HARD]) — Required fix: change the value to `1` and
replace §D.6 item 4's explanation with the correct one — `grep -c` is line-counting, and four of the eleven
names share line 102, so per-name line counts understate occurrences rather than overstating them.

**N5.** AC-HWD-017's `zz-fixture-only` baseline is self-invalidating — `acceptance.md` § AC-HWD-017
`Pre-impl observed:` — It records `grep -rl 'zz-fixture-only' .` → rc=1, no match. At this HEAD the command
returns rc=0 and names `acceptance.md` itself: writing the criterion put the string in the tree. The
substantive property (no *script* by that name exists) still holds. — Severity: **minor** — Class:
**blocking** (evidence attribution; the identical class as iteration 1's D10) — Required fix: scope the
command to where it means something — e.g. `grep -rl 'zz-fixture-only' internal/ .claude/hooks/` — and record
that value.

**N6.** `plan.md:267` reintroduces the claim D11 corrected — `plan.md` §F M3 — It states
*"`t242`/`t243`/`t244` appear only in `spec.md` §G"*, which is false: they appear in `progress.md:30` and in
`plan.md` itself at :76, :263, :304. `spec.md` §C-4 was corrected this iteration; the same inaccuracy now
lives in plan.md. The substantive obligation (never in `internal/template/templates/**`) holds and is
separately enforced by AC-HWD-016. — Severity: **minor** — Class: **optional** — Required fix: reword to match
the corrected §C-4 phrasing.

**N7.** §D.6 item 3 mislabels test declarations as call sites — `acceptance.md` §D.6 item 3 — It cites
"the observed call sites at `session_start_mx_test.go:83,121,135,159`". Lines 83/121/135 are `func Test…`
declarations; the `runMXColdStartScan` call sites are 90/129/148. AC-HWD-013's own `Harness correction:`
cites the same four lines **correctly**, as "the tests asserting the scan's presence". — Severity: **minor**
— Class: **optional** — Required fix: change "call sites" to "tests" in §D.6 item 3.

## Gaps — what this audit did NOT observe

1. **No mutant was implemented as production code.** N1 and N3 were constructed **and executed** against real
   fixtures (a real `moai doctor` run; a real rule-file fixture evaluated with the criterion's own greps), so
   those two are measured rather than reasoned. N2's mutant is reasoned from reading AC-009/AC-010/AC-018
   against the measured `dormant` occurrences — I did not edit the four rule files to demonstrate it. The
   author's own mutants (AC-013 (i)/(ii), AC-011 (i)-(iii), AC-017 (i)/(ii), AC-015 (i)-(iii)) were assessed
   by reading, not by compiling.
2. **`go vet`, `golangci-lint`, and the full package suites were not run.** This is a plan-phase audit over
   documents. The only test executed was the single `-run TestSessionStartMXColdStartIntegration` case
   required to verify AC-013 (c).
3. **The primary checkout was not inspected**, per the dispatch's no-sibling-tree constraint. AC-HWD-004's
   corrected note explicitly attributes its v0.1.0 listing to the primary checkout; I verified only that the
   load-bearing premise (`CreateNodeAtSpawn`, 0 callers) is tree-independent.
4. **`mcp__moai__spec_audit` / `spec_drift` produced no usable evidence** for this tree (no `project_root`
   parameter; returns the primary checkout's 627-SPEC catalog). SPEC lifecycle drift for this tree is
   therefore unmeasured by the dedicated tool. `./bin/moai spec lint` on the tree-built binary is the
   substitute and is clean.
5. **`bin/moai` was not rebuilt by me.** I used the existing `bin/moai` (mtime 23:20, post-dating the last
   Go-source change; the only commit since iteration 1 touches SPEC artifacts only), so it is valid for this
   tree — but I did not run `make build` to confirm it byte-for-byte.
6. **AC-HWD-003's parity semantics were verified against the template by reading**, not by writing the
   set-diff. I confirmed the target entries exist at `settings.json.tmpl:78/86/94/198` in iteration 1 and did
   not re-derive the full 4-template-only / 1-project-only set-diff this iteration.
7. **The three investigation reports were not re-verified.** Iteration 1 spot-checked them; this iteration
   scoped to the defect delta per the retry-loop contract.

## Residual risk

The v0.2.0 revision is a real improvement, and I verified it rather than accepting it: the two claims most
worth doubting — that the AC-001/002 retirement did not trade depth for headroom, and that AC-013 (c) is a
genuine behavioural falsifier — both hold under direct test. Defective criteria fell from six to two.

What remains risky is a pattern, and it is the same one iteration 1 named. **Two of the three new findings
live inside iteration-1 fixes**: widening AC-007 to close a false-pass hole created a false-fail blocker
(N1), and scoping in the "dormant" surfaces closed one direction while opening the inverse (N2). A third
(N3) sat in a criterion that was never rewritten and therefore never re-attacked — which is the strongest
argument for the author's own `Fresh mutant attempt` discipline being applied to **every** criterion at each
revision, not only to the ones changed.

The specific tell to watch is unchanged and now has three instances across two iterations: a closure claim
recorded in prose that the Then clause does not deliver. v0.1.0's "none constructible" (D5) was one;
AC-018's "fails AC-HWD-009 + AC-HWD-010" (N2) and AC-009's "a disposition class per name" (N3) are two more.
Prose mitigations are not gates, and they read as gates.

Finally, N1 deserves emphasis disproportionate to its one-line fix. It is the only finding that would stall
the run phase on a cause unrelated to the work: a correct implementation cannot make AC-HWD-007 pass, and the
failure would present as the diagnostic misbehaving.

## Recommendation

**PASS** at 0.862 against the Tier M threshold of 0.80 (and above 0.85 as well, so the threshold discrepancy
does not decide it), with all seven must-pass criteria PASS or N/A and no unresolved iteration-1 defect.

**PASS is not a statement that the three blocking findings are optional.** M5's firewall covers the seven
must-pass criteria; N1-N3 are absorbed into the Testability score, which I set at 0.80 explicitly because of
them. The orchestrator should route these three before run-phase entry — all are cheap:

1. **N1** (critical) — make AC-HWD-007 satisfiable: warm the fixture before snapshotting, or exclude
   `config-cache.json` by name with a stated reason. One line in the fixture-construction note.
2. **N2** (major) — extend AC-HWD-018's Given to all six surfaces REQ-HWD-012 names, and retract the false
   fresh-mutant closure claim. Note that three of `hook-independence.md`'s five "dormant" occurrences are
   correct and must survive.
3. **N3** (major) — restate AC-HWD-009's Then clause as one row per name, and record the blanket-line mutant.

N4 and N5 are corrections to measured-evidence blocks under a [HARD] constraint and should ride along; both
are single-value edits. N6 and N7 are optional and left to the orchestrator's discretion — routing them adds
no requirement surface and costs two rewordings, but neither affects correctness.

This is iteration 2 of a Tier M ceiling of 2, so this verdict is terminal. Verdict authority for any further
revision remains with this role; an orchestrator self-assessment does not substitute for a re-audit.
