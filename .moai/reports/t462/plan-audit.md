# SPEC Review Report: SPEC-CODEX-E2E-MEASURE-001
Iteration: 1/2 (Tier M ceiling = 2)
Verdict: FAIL
Overall Score: 0.75 (Tier M PASS threshold: 0.80)

Auditor: plan-auditor (Claude anchor) + codex backend (required) + glm backend (inconclusive, fail-open)
Tree audited: `WT-codex-e2e` @ `c7a4d0360` (base `e9c6a8564`, verified live)
Report stream: this file (dispatch-designated location, `.moai/reports/t462/plan-audit.md`)

## Verdict summary

The SPEC is structurally sound — frontmatter, GEARS requirement layer, Out of Scope
discipline, D7/D8 hygiene, and the two-prong measurement scope all hold, and every
inventory count re-executes exactly. But three blocking-class defects break the card at
run phase if the plan is followed to the letter: the live-test gate inventory is factually
wrong (D2), the plan's test commands cannot produce the swept-count evidence its own
MUST-PASS AC demands (D7), and the "codex-axis test surface" definition undercounts while
M4's non-recursive package patterns silently skip genuine codex-axis subpackage tests
(D8). All fixes are a single focused manager-spec amendment pass.

Convergence record: the Claude anchor sent to `mcp__moai__audit_multi` was
`pass-with-debt`. Codex returned FAIL with 8 findings; GLM was inconclusive (fail-open).
Each load-bearing codex finding was then re-verified first-hand in this tree (findings 2,
3, 5 — see Defects D7, D8, D9 and the Verification Evidence section). The verdict moved
to FAIL on that independently-verified evidence, not on codex's authority. Tool result:
`overall_verdict: fail`, `residual_risk_note: "required-backend FAIL: codex"`,
`fail_open_backends: ["glm"]`.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — REQ-CEM-001..011 sequential, no gaps, no
  duplicates, consistent `CEM` domain prefix and 3-digit padding (spec.md:76-112).
- [PASS] MP-2 EARS/GEARS compliance (requirement layer only) — all 11 REQs anchor on one
  of the five patterns: Ubiquitous (001, 006), When (002, 005, 007, 011), Where (003),
  While (004, 009), Unwanted/negative (008, 010). ACs are correctly Given-When-Then
  (verification layer) and are NOT penalized here. Caveat logged as D6: the repo's own
  spec-lint flags REQ-CEM-003/005/009 as malformed (compound/parenthetical/colon forms)
  — warning severity, 0 errors under `--strict` (codex run; binary-lag caveat: installed
  `moai` v3.2.0-rc.0 built from ancestor `e79c010b8`).
- [PASS] MP-3 YAML frontmatter validity — all 12 canonical fields present with correct
  types and canonical names (spec.md:1-16): `id`, `title` (quoted), `version: "0.1.0"`
  (quoted semver), `status: draft` (valid enum), `created`/`updated: 2026-09-03`,
  `author`, `priority: P1`, `phase: "v3.1.4 target"` (release label, not a prohibited
  lifecycle-stage value), `module`, `lifecycle: spec-anchored`, `tags` (comma string).
  No snake_case aliases. Optional `tier: M` + `related_specs` well-formed. plan.md and
  acceptance.md are stateless on the status axis (no `status:` field) per the schema's
  artifact-statelessness rule. Multi-segment ID matches repo-wide norm.
- [N/A] MP-4 language neutrality — single-language (Go) project-scoped SPEC; auto-passes.
- [PASS] MP-5 D7 cross-SPEC reconciliation — all three referenced SPECs
  (SPEC-CODEX-WIRING-001, SPEC-CODEX-LAUNCHER-001, SPEC-CODEX-REVIEW-TARGET-001) exist
  in `.moai/specs/` with `status: completed` (none retired/superseded/archived → no
  BLOCKING). SPEC-ID extraction over all three artifacts found no other references.
- [PASS] MP-6 D8 cross-platform discipline — `grep -c syscall` = 0 across spec.md,
  plan.md, acceptance.md → auto-pass.
- [PASS] MP-7 clarification gate — `grep -rn '\[NEEDS CLARIFICATION'` over plan.md and
  spec.md: no matches. research.md absent (Tier M artifact set: spec/plan/acceptance) →
  N/A for that file.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | minor ambiguity in 1-2+ requirements | D1 (17-vs-16 miscount, spec.md:48), D2 (§B "live test" definition wrong for 3 of 4 files), D6 (3 parser-flagged REQ forms) — each resolvable, but a run-phase executor hits them |
| Completeness | 0.75 | sections complete; core content gap | All sections + 4 `### Out of Scope — <topic>` H3 subsections with bullets present; but §B's "codex-axis test surface = the 73-file union" definition is materially incomplete (D8: ≥50 lexical-codex test files unclassified; subpackage tests never executed) |
| Testability | 0.75 | 1+ AC not satisfiable as planned | AC-CEM-002's `=== RUN` count ≥ 1 is unreachable from plan M4 steps 1/2 (no `-v`, D7); AC-CEM-005's Given is factually wrong (D2); AC-CEM-011's fixed diff base is brittle (D3). The AC form itself is exemplary — every AC names its failure mode |
| Traceability | 0.75 | indirect mapping | Every REQ has ≥1 AC and every AC maps to ≥1 REQ via exact numbering parallel + plan.md §H; but zero AC text cites a REQ id (repo lint: 11 uncovered-REQ warnings) |

Aggregate: 0.75 < 0.80 threshold, independent of the blocking defects below.

## Defects Found

Blocking-class (fix before run phase; each alone bars PASS):

- **D1..D11** are listed with artifact:line, severity, class, and required fix:

D1. minor / optional — spec.md:48 + inventory-baseline.md §1(b) — dependency-axis prose
labels `internal/cli (17)` but enumerates 16 files; the re-executed total (22) and union
(73) only reconcile with 16. Fix: correct the label to (16) in both files.

D2. major / **blocking** — inventory-baseline.md:103-106, spec.md §B:68-69, plan.md §B.5
+ M4-step-5 — Live-gate inventory is factually wrong. Claim: 4 files gated on
`MOAI_CODEX_LIVE_PROBE`, "All SKIP by default". Measured (this audit):
`codex_live_protocol_probe_test.go` alone uses that gate (opt-in);
`codex_review_gate_live_test.go:35` and `codex_review_target_live_test.go:37` are opt-OUT
via `MOAI_SKIP_LIVE_CODEX` (default-RUN when the codex binary is on PATH — and it is, at
`/Users/goos/.local/bin/codex`); `audit_pin_live_test.go:38` uses `MOAI_AUDIT_PIN_LIVE`
(opt-in). Consequences: AC-CEM-005's Given clause is false; M4 step 5's SKIP inventory is
mis-scoped; prong A's whole-package `internal/cli` run will execute live-codex tests with
nothing enabled, contradicting REQ-CEM-010's opt-in premise. Fix: attribute the three
gate classes correctly in all three artifacts; decide at Implementation Kickoff Approval
whether prong A sets `MOAI_SKIP_LIVE_CODEX=1` (and `MOAI_AUDIT_PIN_LIVE` stays unset).

D3. minor / optional — acceptance.md:84-86 — AC-CEM-011 pins `git diff --name-only
e9c6a8564..HEAD` literally; plan M2 / spec §D.2 anticipate base movement (t451/t452
absorption), after which the diff includes foreign files and the AC fails spuriously.
Fix: diff against the pinned base recorded in progress.md §E (re-read at close), or
against the card's own commit set.

D4. minor / optional — acceptance.md:97 — Git hygiene is one of the five HARD dispatch
constraints but its AC (AC-CEM-012) is SHOULD-PASS; the other four HARD constraints all
land on MUST-PASS ACs. Fix: promote AC-CEM-012 to MUST-PASS or record the downgrade as a
deliberate lead-level choice.

D5. minor / optional (upgraded by lint corroboration) — acceptance.md (all ACs) — no AC
text references a REQ id; the repo's spec-lint emits 11 uncovered-REQ warnings (codex
run, corroborated by inspection: acceptance.md contains zero `REQ-CEM` strings). Mapping
is derivable but unstated. Fix: add `(→ REQ-CEM-00N)` to each AC.

D6. minor / optional — spec.md:82-86, 91-94, 103-105 — REQ-CEM-003 (parenthetical +
compound response), REQ-CEM-005 (two shall-clauses), REQ-CEM-009 (tautological
`unmerged-or-merged` While-condition, colon extension) are flagged by the repo GEARS
parser and drift from single-response canonical form. Intent is clear in all three. Fix:
tighten to one-line canonical clauses (009 → Ubiquitous).

D7. major / **blocking** — plan.md M4 steps 1, 2, 4 vs acceptance.md AC-CEM-002/004 —
the commanded runs (`go test ./internal/codexwiring/... -count=1` etc.) carry no `-v` or
`-json`; plain `go test` output contains zero `=== RUN` lines, so AC-CEM-002's
"swept (`=== RUN`) count ≥ 1" (MUST-PASS) and AC-CEM-004's swept counts are
unsatisfiable as commanded — the AC fails a valid run by construction (step 3 alone has
`-v`, showing the omission in 1/2/4 is an oversight, not a choice). Fix: add `-v`
(or `-json` with `Action:"run"` counting and a named parser) to M4 steps 1, 2, 4.

D8. major / **blocking** — spec.md §B:64 + plan.md M4 steps 3-4 vs REQ-CEM-003's own
words — two coupled holes in the surface definition/execution:
  (i) REQ-CEM-003 says "execute the whole package", but M4 uses non-recursive patterns
  (`./internal/cli/`, `./internal/template/`) — Go package semantics exclude subpackages,
  so codex-axis tests in `internal/cli/wizard` (≥4 files, incl.
  `agent_wiring_question_test.go` — claude/codex/both wiring selection) and
  `internal/template/agentemit` (≥5 files, incl. `golden_test.go:172`
  `TestRealSetCodexShape` pinning the codex realset shape) never execute in prong A.
  (ii) §B defines "codex-axis test surface" flatly as the 73-file union, but a broad
  lexical sweep (`grep -rli codex internal/ --include='*_test.go'`, minus codex-named
  files and the two whole packages) finds 72 files of which the symbol-axis regex catches
  only 22 — a 50-file delta with no inclusion/exclusion classification. Several are
  genuine codex-axis tests outside the union entirely (`internal/config/audit_models_test.go`
  — 22 codex refs, codex audit-model serialization; `internal/settings/audit_pin_fields_test.go`;
  `internal/web/*audit*` family). The card intent ("execute the FULL existing codex-axis
  test surface") is not met by the current definition. Fix: make M4 patterns recursive
  (`./internal/cli/...`, `./internal/template/...`), extend or re-derive the inventory
  with a checked-in manifest classifying every lexical-codex candidate
  (included/excluded + reason), and re-state the union arithmetic; alternatively bound
  the claim explicitly as a defined two-axis heuristic with the misses named as known
  exclusions — but that weaker form conflicts with the card intent and needs the lead's
  sign-off.

D9. major / blocking-adjacent — plan.md M3 — the positive-control recipe has two
defects: (a) its primary example ("a `.codex/config.toml` with a key the codexadapter
`ValidateConfig` whitelist must reject") is wrong about the code — `ValidateConfig`
validates hooks.json (`codex_readiness.go:116` binding, `doctor_codex.go:56` call,
`codexwiring/wire.go:86`); config.toml is existence-checked only
(`classifyCodexWiring`), so the suggested break would NOT be detected and the card's
hardest gate (AC-CEM-006) would record a failed control. The in-sentence alternative
(broken hooks.json stray key) IS correct. (b) `moai codex status` resolves to the
INSTALLED binary (`/Users/goos/go/bin/moai`, v3.2.0-rc.0, built from ancestor
`e79c010b8` — behind this tree's `c7a4d0360`), violating REQ-CEM-001's tree-pinning
discipline. Fix: lead with the hooks.json stray-key break with an exact expected
diagnostic, and invoke the tree (`go run ./cmd/moai ...` or a fresh `make build`) for
every CLI-surface check.

D10. should-fix / optional — plan.md §C + spec.md §D.1 — `CODEX_HOME` isolation is
assignment-shaped ("Create throwaway CODEX_HOME=…") rather than exported/command-scoped,
and M4's test commands carry no `CODEX_HOME` prefix; the non-mutation tripwire hashes
only `config.toml` and lists top-level skills (misses `auth.json`, hooks.json, nested
skill content). Existing repo tests carry their own t.TempDir isolation, so the residual
risk is narrow — but the tripwire should be a full-tree snapshot (`find ~/.codex -type f`
+ shasum) and the preflight should export the variable in the same invocation as each
command that needs it.

D11. advisory / optional — acceptance.md (MUST-PASS set) — no per-AC RED-now/green-path
evidence cells (verification-completeness §2). For a measurement SPEC the pinned
baseline partially substitutes (G1-G8 establishing commands, machine-state table, all
pinned to `e9c6a8564`), and the report-existence ACs are trivially red at plan phase;
the formal ledger is still absent. Surfaced for the amendment pass, not blocking.

### Adjudications requested by the dispatch

- **M1 (internal/cli WHOLE vs `-run` selector): SOUND.** The rationale (heterogeneous
  test names across 50+ files; whole-package is the only selector whose completeness is
  trivially established, per the swept-count/empty-sweep discipline) holds, the decision
  is explicitly marked for lead confirmation at Implementation Kickoff Approval with a
  swept-count-proof requirement on any override, and the whole-package 1800s standalone
  form is what dispatch constraint 4 itself authorizes. Codex's counter (553 cli test
  files → ~499 unrelated; unrelated reds pollute a codex measurement) is a real
  attribution caveat but is already adjudicated by the constraint and handled by M1's
  kickoff gate + report discipline. Note: D8(i) strengthens the case for `./internal/cli/...`
  (recursive) rather than changing the whole-vs-selector decision.
- **AC ordering (positive-control gate first): SOUND as structure, defective in recipe.**
  M3 precedes M4 in milestone order; AC-CEM-006 is MUST-PASS with an explicit
  timestamp/ordering-before-verdict requirement; AP-1 names the vacuous-zero trap;
  AC-CEM-008's grep-method control is viable (verified: `grep -c doctor
  e2e/cli/tux3_journeys.sh` → 7 > 0 on the same file where codex → 0). The ordering gate
  is the SPEC's strongest element — but see D9: its M3 example validates the wrong
  artifact with an unpinned binary.
- **Peripheral 5 packages:** correctly in scope via the verified dependency axis, bounded
  runtime (whole-but-cheap, template separate for embed cost). Not dominating. Note D7
  (no `-v`) and D8(i) (`./internal/template/` non-recursive) both touch step 4.
- **No vacuous live-probe AC:** confirmed — no AC requires executing a probe; AC-CEM-005
  requires enumerating skips (but its Given is wrong — D2).
- **Divergence recording:** exemplary — the 49-skills drift is recorded with measured
  values (all re-verified by this audit: 0 `[skills.*]` sections, 18 `[plugins."moai-*]`
  blocks all enabled=true, skills dir = `.system` + `hatch-pet`), and no check builds on
  the wrong number; AC-CEM-006's Given uses only the confirmed fact.
- **Inventory reproducibility:** confirmed — all four §A commands recorded verbatim and
  re-executed by this audit with identical counts (38 / 6 / 7 / 22 → 44 / 29 / 73).

## Verification Evidence (commands run by this audit, this tree, `c7a4d0360`)

- `git rev-parse --short HEAD` → `c7a4d0360`; `git branch --show-current` → `WT-codex-e2e`;
  parent `e9c6a8564` — matches dispatch.
- The four §A inventory commands re-run verbatim → 38, 6, 7, 22 (identical to SPEC);
  dependency-axis cli-only count → **16** (not 17 — D1).
- `grep -ric codex e2e/cli/tux3_journeys.sh` → 0 (exit 1); `grep -c doctor …` → 7;
  `find e2e -type f | wc -l` → 1. Single-e2e-file claim and grep-control token verified.
- Code citations verified: `resolveCodexHomeDir` mcp_codex.go:1758 (env-first at :1759),
  `codexHomeEnvVar` :1730, `Use: "codex [cli | status | app]"` codex_launcher.go:293,
  `codex-review-gate` hook.go:221, MCP tools mcp_server.go:270/294, live-probe const
  codex_live_protocol_probe_test.go:31, `ValidateConfig` codexadapter/config.go:53,
  codexwiring package doc (t83 Finding D / hooks.json whitelist) — all exact.
- Machine state re-verified read-only: skills dir, config.toml shapes (see Divergence
  recording above).
- Live-gate enumeration: `grep -rl 'MOAI_SKIP_LIVE_CODEX|MOAI_AUDIT_PIN_LIVE|MOAI_CODEX_LIVE_PROBE'`
  → exactly the 4 named files, with per-file gate constants read at the cited lines (D2).
- Surface delta: broad-lexical 72 vs symbol-axis 22 → 50 unclassified files (D8);
  codex-named misses enumerated (wizard ×4, agentemit ×5, config, settings, …).
- `command -v codex` → `/Users/goos/.local/bin/codex` (present); `command -v moai` →
  `/Users/goos/go/bin/moai` v3.2.0-rc.0 (ancestor build — D9b).
- Cross-model: `mcp__moai__audit_multi(project_root=<this worktree>, target=baseBranch)`
  → codex FAIL (8 findings, findings 2/3/5 re-verified here), glm inconclusive,
  overall fail.

## Gaps (what this audit did NOT observe)

- No test suite was executed (plan-phase audit; `internal/cli` whole-run is run-phase
  work). D7's no-`=== RUN`-without-`-v` claim rests on Go output semantics plus codex's
  re-execution of codexwiring/codexadapter (`ok … 0.603s`, no RUN lines), not on this
  audit's own run.
- `moai spec lint --strict` was not re-run by this audit (binary lag vs tree); the
  14-warning figure is codex's run, corroborated by direct inspection of the three REQ
  forms and the AC text.
- The 50-file lexical delta was classified by sampling (wizard, agentemit, config,
  settings, web families inspected); not every one of the 50 was individually read.
  D8's blocking classification rests on the verified subpackage-execution hole (i), which
  does not depend on the classification.

## Residual Risk

- Codex's challenge that measurement-only conversion of "all codex-related e2e" might not
  be authorized is resolved by the dispatch itself (measurement-only, authoring out of
  scope) — not a defect.
- If the lead chooses to accept D8's weaker alternative (bounded two-axis heuristic with
  named exclusions), prong A's headline ("execute the full existing codex-axis test
  surface") must be re-worded to match, or the report will overclaim.

## Recommendation (for manager-spec, iteration 2)

1. Fix D7: add `-v` (or `-json` + parser) to plan M4 steps 1, 2, 4.
2. Fix D8: recursive patterns `./internal/cli/...`, `./internal/template/...` in M4;
   extend the inventory manifest to classify the 50-file lexical delta
   (included/excluded + reason) and re-state the union arithmetic in spec.md §A/§B.
3. Fix D2: correct the live-gate attribution (three gate classes) in baseline §4,
   spec.md §B, plan §B.5/M4-5; add the kickoff decision on `MOAI_SKIP_LIVE_CODEX`.
4. Fix D9: hooks.json-stray-key as the primary M3 positive control with expected
   diagnostic; tree-pinned binary (`go run ./cmd/moai`) for CLI-surface checks.
5. Fix D1 (17→16), D3 (diff base from progress-pinned base), D4 (AC-CEM-012 severity),
   D5 (REQ refs in ACs), D6 (three REQ forms), D10 (export + full-tree snapshot).
6. Re-audit scope (iteration 2, per the retry loop): the enumerated defect delta D1-D10
   plus a regression check; D11 remains optional.

Defects D2, D7, D8 are the verdict-bearing set; everything else can ride the same
amendment pass.

---

# Iteration 2/2 — Re-audit (defect-delta scope)

Verdict: **PASS-WITH-DEBT**
Overall Score: **0.875** (Tier M threshold 0.80 — cleared)
Repair audited: commit `519f766d4` on `WT-codex-e2e` (+284/−146, 5 files — verified:
all 5 under `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/` + `.moai/reports/t462/`; `git diff
e9c6a8564..HEAD -- e2e/` empty).
Scope: iter-1 defect delta + regression check, per the Retry Loop Contract. No
from-scratch re-audit; no second cross-model pass (delta-scoped iteration; iter-1 codex
findings were already first-hand-verified in iter 1).

## Regression Check — iter-1 defects

| Iter-1 defect | Status | Evidence (verified at `519f766d4`) |
|---|---|---|
| D1 (17-vs-16 label) | **RESOLVED** | spec.md:49 + baseline §1(b) now "(16)"; my re-execution of the symbol axis → 22 with cli-only = 16 — arithmetic reconciles |
| D2 (live-gate misattribution) | **RESOLVED** | spec.md §B:80-88 three gates with correct semantics (PROBE opt-in :31 / SKIP_LIVE_CODEX opt-out :35,:37 + default `=1` policy / AUDIT_PIN_LIVE opt-in :38) — matches my iter-1 gate measurements; REQ-CEM-014; AC-CEM-005 corrected Given + failure modes (incl. running WITHOUT `=1`); baseline §4 corrected table; plan §B.5 + M1-D2 kickoff item |
| D3 (fixed diff base) | **RESOLVED** | `BASE_SHA=e9c6a8564` named constant (acceptance header), AC-CEM-011 uses `${BASE_SHA}`, explicit re-pin clause + §D.2 edge case |
| D4 (AC-CEM-012 severity) | **RESOLVED** | AC-CEM-012 MUST-PASS (§D.1, bolded) |
| D5 (no REQ citations in ACs) | **RESOLVED** | all 12 ACs carry `(MUST\|SHOULD, maps REQ-CEM-XXX)`; all 14 REQs covered; `moai spec lint --strict spec.md` → "No findings" (run by this audit) |
| D6 (REQ form) | **RESOLVED** | 003/005/009 single-shall; 010 live-probe clause corrected (both opt-in gates named); lint clean |
| D7 (no `-v` ⇒ no `=== RUN`) | **RESOLVED** | M4 steps 1/2/3a/3b all carry `-v` (+ header note); AC-CEM-002 RED: "a run without `-v` (no `=== RUN` lines ⇒ count unsatisfiable)"; AC-CEM-003 RED: "missing `-v`" |
| D8 (surface undercount + non-recursive holes) | **RESOLVED in substance** | lexicon axis added (spec §A row; baseline §1c manifest classifying all 50); union 123 arithmetic self-consistent (44+7+22+50); recursive patterns `./internal/cli/...`, `./internal/template/...` in REQ-CEM-012/013 + M4; 5 new packages added. Residuals R-1/R-5 below |
| D9 (control recipe) | **RESOLVED** | M3 binary pin (`go run ./cmd/moai` / fresh build; AP-6 ancestor-binary); hooks.json-first primary — verified against `ValidateConfig` callers (`codex_readiness.go:116`, `doctor_codex.go:56`); secondary config.toml variant verified DETECTABLE (`doctor_codex.go` InspectMCPTable reports "[mcp_servers.moai] table missing") |
| D10 (isolation enforcement) | **PARTIALLY RESOLVED** | command-scoped `CODEX_HOME` per invocation ✓ (plan §C bullet 3, AC-CEM-009); BUT the `~/.codex` tripwire remains narrow (config.toml sha + top-level skills ls) — the added "whole-tree snapshot" covers the REPO tree (`git status` + HEAD), not the codex home (R-3) |
| D11 (RED cells) | **RESOLVED** | explicit RED cells on all 12 ACs (nine MUST-PASS each named); BASE_SHA pin applies to scope ACs |

Tier M ceilings: 14 REQs ≤ 16, 12 ACs ≤ 16 ✓. New REQ-CEM-012/013/014 all Where-pattern,
single-shall ✓. Frontmatter unchanged (MP-3 holds). Scope discipline held: journey
authoring still Out of Scope (§G, REQ-CEM-008), G1–G8 verbatim-unchanged in baseline §3,
no files outside the two allowed roots.

## Iter-2 residual defects (none execution-breaking)

R1. major / blocking-routing (fix BEFORE M2 executes; one-line edits) — spec.md:98-100
(REQ-CEM-002) still reads "re-measure **both** inventory axes … (44 filename / 29
dependency / **73** union)" — the superseded two-axis counts — directly contradicting
AC-CEM-001 (its own mapped AC: 7+22 / 50 (27+23) / union **123**), spec §A/§B, and plan
M2. Same family: §A heading "both axes" (line 36) and §F table "two-axis inventory"
(line 159). Cross-layer revision sweep missed the requirement layer. Cannot misdirect
execution (AC-CEM-001 + M2 carry the correct counts) but the drift baseline is
ambiguous as normatively written. Fix: refresh REQ-CEM-002 to three axes / 123 and the
two stale labels.

R2. minor / optional — spec.md §G ("Full-suite local verification") still scopes
execution to "what REQ-CEM-003 names"; the surface is now named by REQ-CEM-003/012/013.

R3. minor / optional — D10 remnant: `~/.codex` non-mutation tripwire still hashes only
`config.toml` + lists top-level skills (misses `auth.json`, `.codex/hooks.json`, nested
skill content). Command-scoping (resolved half) reduces exposure; a `find ~/.codex
-type f` + shasum snapshot would close it.

R4. nit — plan M3 says the installed moai "reports v3.1.3-era provenance"; measured
(this audit, iter 1): `v3.2.0-rc.0`, built from ancestor `e79c010b8`. Conclusion
(ancestor build, must not be used) holds; the version label is wrong.

R5. minor / optional — 3 codex-named test files exist OUTSIDE the three axis packages
(`internal/web/mcp_codex_surface_test.go`, `internal/web/codex_card_sentinel_test.go`,
`internal/template/codex_agents_deploy_test.go`) and are counted by NO axis (filename
axis is cli-scoped; dependency and lexicon axes exclude codex-named files): true
codex-named surface repo-wide is 41, union counts 123 where the exhaustive figure is
126. EXECUTION covers them (web and template are both recursive-pattern packages), so
no behavioral blind spot — this is a counting-completeness blemish in the inventory the
lead will cut follow-up cards from. Fix: repo-wide filename find, or name the 3 as
counted-elsewhere exclusions.

(nit) spec.md `version:` stays "0.1.0" while acceptance.md bumped to "0.2.0" — sibling
version divergence, schema-permissible.

## Iter-2 verification evidence (this audit, tree `519f766d4`)

- `git rev-parse --short HEAD` → `519f766d4`; branch `WT-codex-e2e`; base-anchored diff
  → exactly 5 files, all in allowed roots; `-- e2e/` diff empty.
- `moai spec lint --strict .moai/specs/SPEC-CODEX-E2E-MEASURE-001/spec.md` →
  "✓ No findings — all SPEC documents are valid" (binary-lag caveat: installed moai
  built from `e79c010b8`, an ancestor whose delta to this tree is this card's markdown
  only — lint engine identical in practice).
- Lexicon sweep re-built with the iter-1 commands: broad 72, symbol axis 22, comm-delta
  **50**; per-file `grep -ci codex` over all 50 → **27 ≥5 refs / 23 in 1–4** — the
  baseline's classification reproduces exactly.
- Per-file spot counts: mcp_convergence_participant 29, agentemit_test 25,
  audit_models 22, audit_pin_fields 16, sessionmsg/agent 5 (all ≥5, behavioral);
  version_stamp_registry 4, wizard/agent_wiring_question 3 (incidental) — all match
  baseline claims.
- `find internal/ -name '*codex*_test.go'` outside cli/codexwiring/codexadapter → 3
  files (R5).
- doctor_codex.go `InspectMCPTable` read: "[mcp_servers.moai] table missing from
  config.toml" is a reported problem — M3's secondary control variant is detectable.

## Gaps (not observed at iter 2)

- No test execution (still plan-phase; run-phase work).
- Codex-named-file contents in R5 were located, not individually read — the
  execution-coverage claim rests on package membership (web, template), not per-file
  inspection.
- Lint run via the ancestor-built binary (see caveat above).

## Iter-2 verdict rationale

All three iter-1 blocking defects (D2, D7, D8) are resolved and verified by
re-measurement; the plan's commands as now written produce the evidence their ACs
demand; the surface definition is measured-complete at the manifest level; lint is
clean. Score 0.875 ≥ 0.80 with zero execution-breaking residuals → not a clean PASS
only because R-1 is an internal-consistency defect inside a normative REQ (one-line
fix). Route: land R-1 (+R-2/R-5 label fixes if convenient) as a trivial amendment
before run-phase M2, or accept as documented debt — AC-CEM-001 and plan M2 carry the
correct counts and bind the executor either way. Implementation Kickoff Approval
remains the lead's gate regardless of this verdict.
