# SPEC Review Report: SPEC-CODEX-HOOK-ADAPTER-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.63** (harmonic mean; Tier M PASS threshold 0.80)

Audited tree: worktree `.claude/worktrees/t83`, branch `WT-hook-adapter`, fixed commit `268bc3cfd`
(verified at start and re-verified at end of audit — unchanged). Artifacts audited: spec.md (220L),
plan.md (117L), acceptance.md (141L). Reasoning context ignored per M1 Context Isolation; only the
SPEC artifacts and the measured-evidence files they cite were used.

Evidence consulted (all claims below were verified against these, not taken from the SPEC):

- `.moai/reports/t83/precondition-measurement.md` (round 2), `precondition-measurement-round3.md`
- `.moai/reports/t83/probe/` — 11 files (3 payload captures + 8 event streams), tracked on this branch
- `.moai/reports/t91/README.md` (M0) and `t91/hook-payloads/` (8 golden payloads, **untracked**)
- `internal/cli/hook.go:38-72` — the MoAI hook dispatcher's actual command set

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-1..REQ-7, sequential, no gaps, no duplicates
  (`grep -c '^### REQ-' spec.md` → 7; headings at spec.md:L107, L114, L124, L130, L137, L143, L149).
- **[PASS] MP-2 EARS/GEARS format compliance** (judged against the requirement layer, `spec.md §C`):
  REQ-1/2/3/5/6 use `WHERE [condition], the <subject> SHALL [response]` (spec.md:L109, L116-122,
  L126-128, L139-146, L145-147); REQ-4 uses WHERE + SHALL + SHALL NOT (spec.md:L132-135); REQ-7 is
  the negative form "The adapter SHALL NOT …" (spec.md:L151). All 7 match GEARS patterns. The
  Given-When-Then entries in acceptance.md §D.2 are ACs (verification layer) and are correctly not
  penalized here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (spec.md:L2-15); `created`/`updated`/`tags`/`id` canonical names (no snake_case aliases); id
  matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; version quoted semver; `status: draft` in enum; `phase:
  "v3.2"` is a release label (not a prohibited stage token); optional fields `era: V3R6`, `tier: M`
  valid. Verified against `FrontmatterSchemaRule` (`internal/spec/lint.go:741-794`) — priority is
  checked non-empty, `medium` passes.
- **[PASS/N/A] MP-4 language neutrality** — single-language scoped SPEC (Go codebase, codex-cli
  target); no 16-language enumeration obligation. N/A: single-language SPEC, auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-…' spec.md` finds only the SPEC's
  own ID; `related_specs: []`; no external SPEC references → no reconciliation duty, no BLOCKING
  finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → 0 → auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no
  matches (research.md does not exist at Tier M; plan.md checked, clean).

Must-pass firewall: clean. The FAIL verdict comes from the aggregate score (0.63 < 0.80) plus five
blocking defects below — not from a must-pass breach.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.50 | Multiple requirements require interpretation; a reasonable engineer might implement differently than intended | REQ-1's "eight events with a counterpart" (spec.md:L109-112) cannot be reconciled with §B's exclusions (spec.md:L90-104) nor with the dispatcher — see D1. REQ-3's "diagnostic" sink is an undefined term (spec.md:L126-128; acceptance.md:L80 "assertion on the diagnostic sink"). REQ-7's invariance has three phrasings with different scopes (spec.md:L26 vs L91-93 vs L151) — see D8. |
| Completeness | 0.75 | One non-critical structural deficiency; frontmatter complete | HISTORY (§H), WHY (§A), WHAT/scope (§B), REQUIREMENTS (§C), AC (acceptance.md §D), Out-of-Scope content all present and substantive. But the Out-of-Scope entries are bold paragraphs (spec.md:L90, L95, L98, L102), not `### Out of Scope — <topic>` H3 sub-headings with `-` bullets — fails the project's own `OutOfScopeRule` (`internal/spec/lint.go:861-907` requires a `###`-prefixed "out of scope" heading and a `-` bullet; neither exists) → `MissingExclusions` error at lint/CI. See D3. |
| Testability | 0.50 | Several ACs require judgment calls or cannot be executed as written | AC-REQ-6 is unsatisfiable in the execution tree for 3 of its 6 required events (D2). AC-REQ-5a/5b test the empty set — no config emitter exists in this SPEC's scope (D4). AC-REQ-7's "no change to decision logic" judged via `git diff --stat` needs a judgment call (D8). AC-REQ-3b's "a branch added later without one fails this AC" has no mechanical enforcement (D9). The remaining ACs are exemplary — marker-grep methodology with exit-code-is-never-evidence (acceptance.md:L3-5), byte-identical pass-through, count-asserted tables — but four of fifteen carry the defects above. |
| Traceability | 1.00 | Full bidirectional coverage | Every REQ has ≥1 AC (REQ-1:1a/1b; REQ-2:2a-2d; REQ-3:3a/3b; REQ-4:4a/4b/4c; REQ-5:5a/5b; REQ-6:6; REQ-7:7 — acceptance.md §D matrix L9-25); every AC maps to an existing REQ; §D.1 states the MUST-PASS/SHOULD split and its rationale (L27-31). No orphaned ACs, no uncovered REQs. Tier M budget respected (7 ≤ 16 REQ, 15 ≤ 16 AC). |

Harmonic mean = 4 / (1/0.50 + 1/0.75 + 1/0.50 + 1/1.00) = **0.63** < 0.80 → **FAIL**.

---

## Defects Found

**D1. REQ-1-count — spec.md:L109-112 (and acceptance.md:L35-41, plan.md:L68-71) — REQ-1's
"the eight events with a counterpart" is internally unreconcilable and the eight pairs are never
enumerated. — Severity: major — Class: blocking**
The Codex event universe is 11 (t91 README §1). §B excludes five (SubagentStop retired;
PermissionRequest/PreCompact/PostCompact/SubagentStart "have no MoAI counterpart", spec.md:L98-103),
implying **six** wired events. Yet the MoAI dispatcher registers argument counterparts for **all
eleven** Codex events, including the four §B declares counterpart-less: `compact` (hook.go:51,
EventPreCompact), `post-compact` (hook.go:62), `permission-request` (hook.go:56), `subagent-start`
(hook.go:54). So "no MoAI counterpart" is factually false — the exclusions are a policy choice, not
an absence — and the count is 6 (wired) or 11 (counterparts exist), never 8. AC-REQ-1a pads to
eight with "and the two the dispatcher distinguishes internally" (acceptance.md:L38-39), two pairs
named nowhere in any artifact. A reasonable engineer will build a 6-, 8-, or 11-pair table depending
on which sentence they read first.
**Required fix**: enumerate the full mapping table inside REQ-1 or AC-REQ-1a (Codex name →
dispatcher argument for each pair); restate §B's exclusions as policy scoping with the real
rationale; make the asserted count match the enumerated table.

**D2. GOLDEN-FIXTURE-AVAILABILITY — spec.md:L143-147, acceptance.md:L114-119, plan.md:L24-26 —
AC-REQ-6 is unsatisfiable in the execution tree for PostToolUse, SessionStart, and SessionEnd. —
Severity: major — Class: blocking**
REQ-6/AC-REQ-6 require golden tests over payload dumps "under `.moai/reports/t91/hook-payloads/` or
`.moai/reports/t83/probe/`" covering six events. Measured state of the fixed tree: the tracked
`t83/probe/` contains payload captures for only **three** events (`PreToolUse.observed.json`,
`Stop.observed.json`, `UserPromptSubmit.observed.json`; the `run-*.jsonl` files are Codex event
streams, not hook-stdin payloads). The dumps for the other three events exist **only** in
`.moai/reports/t91/hook-payloads/` — which `git ls-files` shows is **untracked** (0 tracked files)
and exists only in the primary checkout: it is absent from this worktree, from the branch, and from
any CI run. plan.md §C pre-flight 2 ("both resolve") therefore already fails in the worktree where
run-phase executes. This is exactly the worktree/primary path-dependence hazard: the MUST-PASS
golden AC cannot be executed as written.
**Required fix**: commit the needed t91 payloads on this branch, or vendor the six payload dumps
verbatim into a tracked `testdata/` directory (citing origin), and re-point REQ-6/AC-REQ-6 at the
tracked location.

**D3. OUT-OF-SCOPE-CONVENTION — spec.md:L90-104 — the Out-of-Scope section fails the project's own
lint rule. — Severity: major — Class: blocking**
`OutOfScopeRule` (`internal/spec/lint.go:861-907`) requires an H3/H4 heading containing "out of
scope" followed by a `-` bullet item. The SPEC's four exclusion entries are bold paragraph lead-ins
(`**Out of Scope — …**:`) with no `###` heading and no bullets, so the rule's scan never enters the
section and emits `MissingExclusions` at error severity — a CI-tier failure the moment the SPEC
lands. The *content* of the exclusions is excellent (specific, reasoned, each with a measurement
basis); only the form is wrong.
**Required fix**: convert each `**Out of Scope — <topic>**:` lead-in to a
`### Out of Scope — <topic>` heading with at least one `-` bullet beneath it.

**D4. AC-REQ-5-VACUITY — spec.md:L137-141, acceptance.md:L102-111, plan.md:L89-92 — AC-REQ-5a/5b
as written test the empty set. — Severity: major — Class: blocking**
REQ-5 constrains "anything in this repository [that] emits a Codex hooks config", but this SPEC
builds no emitter — §B excludes the wiring generator (t88/M4) and plan M5 states "No generator is
built here; this milestone is the constraint other work reads." At this SPEC's close the repository
emits zero Codex hooks configs, so "Given any Codex hooks config this repository emits, … it has no
top-level `version`" and the whitelist-subset check pass vacuously — two MUST-PASS ACs that cannot
fail. The measurement basis for the constraint itself (version key kills the file; field whitelist)
is solid — the defect is only that the AC's subject does not exist in scope.
**Required fix**: bind AC-REQ-5a/5b to the M5 deliverable (the whitelist/validator unit): assert
that every config object the constraint unit accepts/produces validates against the
measured-accepted set, with negative samples (`version` key, unknown field) rejected.

**D5. FINDING-D-ACCURACY — spec.md:L67-74 (§A Finding D), L192-201 (§F), acceptance.md:L102-107 —
"with no warning or error" is contradicted by the SPEC's own cited evidence. — Severity: major —
Class: blocking**
Finding D states the top-level `version` key disables the hooks file "with no warning or error", §F
says the generated file would "do nothing, with nothing reporting that", and AC-REQ-5a's
"silently" repeats it. But the SPEC's own §D-cited artifact, `probe/run-versionkey-kills-file.jsonl`
line 4, carries an explicit error item: `"failed to parse hooks config …: unknown field 'version',
expected 'description' or 'hooks' at line 2 column 11"` — file, field, line, and column all named in
the `--json` event stream. The operational conclusion (file disabled; don't emit the key) survives,
but the "fails silently" framing that anchors REQ-3's rationale and §F's blocker narrative is
overstated for this failure mode. The *project-level* non-firing case IS genuinely silent — verified:
`run-projectlevel-nofire.jsonl` contains zero error items — so §F's concern stands there and only
there.
**Required fix**: restate Finding D as "the file is disabled; the failure surfaces only as a JSONL
error item in the `--json` stream, with no interactive warning", scope §F's "nothing reporting
that" to the project-level case, and align AC-REQ-5a's "Why it matters" wording.

**D6. REQ-3-SINK — spec.md:L124-128, acceptance.md:L76-80 — the discard diagnostic's sink is
unspecified and its interaction with REQ-4's stderr pass-through is undefined. — Severity: minor —
Class: optional**
REQ-3 mandates a diagnostic for every undeliverable message but never says where it goes; AC-REQ-3a
defers to "the diagnostic sink", an undefined term. If the sink is stderr while the underlying hook
exited 2, appending a diagnostic line would corrupt exactly the blocking-reason / continuation-prompt
text REQ-4 mandates passing through unmodified. Current MoAI hooks make the collision unlikely
(sync gate's `systemMessage` rides exit 0), but nothing in the SPEC constrains it.
**Required fix**: name the sink in REQ-3 (e.g., a dedicated log path, or stderr only when the
underlying hook did not exit 2) and state the interaction rule with REQ-4.

**D7. REQ-4-EVIDENCE-TIER — spec.md:L130-135, acceptance.md:L60-61 — normative stderr classes
assigned on declaration-only evidence, including two events §B excludes from wiring. —
Severity: minor — Class: optional**
Measured classes are only PreToolUse (blocking reason) and Stop (continuation prompt). REQ-4
additionally assigns blocking-reason treatment to `UserPromptSubmit` and `PermissionRequest` and
continuation treatment to `SubagentStop` — all grounded in binary strings, none observed (round-2 §5
explicitly did not re-test exit 2 on other events), and `PermissionRequest`/`SubagentStop` are
events §B says are "not wired here", making their clauses dead branches. AC-REQ-4a/4b correctly
cover only the two measured events and AC-REQ-4c marks the rest unmeasured, so the AC layer is
honest — the REQ text is not. AC-REQ-2b's "the binary carries that error string per event"
similarly overstates (the empty-reason string was located for Stop/SubagentStop only).
**Required fix**: annotate the three unmeasured classes in REQ-4 as declared-not-measured, and drop
or explicitly mark the excluded events.

**D8. REQ-7-SCOPE — spec.md:L26 vs L90-93 vs L149-152, acceptance.md:L121-126 — the
`internal/hook` invariance is stated at three different scopes. — Severity: minor — Class: optional**
§A reports the card premise "105 files / 22,659 LOC stay untouched"; §B says "the parsing and
decision logic stays unchanged"; REQ-7 prohibits modifying "decision logic"; AC-REQ-7 filters
`git diff --stat -- internal/hook/` and routes a non-empty result to "justification". Whether a NEW
adapter file placed under `internal/hook/` (touching no existing logic) satisfies REQ-7 is
undecidable from the text, and a diff-stat cannot mechanically distinguish decision-logic changes
from any other change — the AC's fallback ("requires justification") is a judgment call, weakening
its binary-testability.
**Required fix**: pin the adapter's package location outside `internal/hook/` (or explicitly permit
new non-decision files), and make the AC criterion mechanical — e.g., zero modifications to
pre-existing files under `internal/hook/`.

**D9. AC-REQ-3B-ENFORCEMENT — acceptance.md:L83-86 — "a branch added later without one fails this
AC" has no mechanical backing. — Severity: minor — Class: optional**
The discard-branch enumeration lives in the test itself; a discard branch added to the adapter
without a matching test case passes every existing check silently. The claim only holds if the AC
is re-audited with coverage data.
**Required fix**: tie AC-REQ-3b to a coverage assertion on the adapter's discard paths, or to a
branch-count constant shared between implementation and test.

---

## What the SPEC gets right (for calibration)

The measurement discipline is the strongest part of this SPEC and should be preserved through
revision: scratch `CODEX_HOME` isolation with mtime/hash verification of the real user config
(report §0/§6), single-variable A/B for the `version`-key conclusion, an explicit retraction of the
round-2 matcher hypothesis, masked absolute paths in preserved evidence, and probe artifacts
actually committed on the branch. §E and acceptance §D.3 state the unmeasured surface honestly
instead of hiding it, the AC-REQ-2 series is fully measurement-grounded (2c's live-check even
specifies the marker-grep methodology and why exit code is never evidence), and the partial-mapping
scope decision (three inert keys only, everything measured-working passes through byte-identical)
is exactly the thin seam the evidence supports. Focus items (a) and (b) from the review brief came
back clean: no AC hangs on an unmeasured Codex behavior (the continue-series rewrite is asserted at
the transform level; the live ACs ride measured channels), and the length-not-content diagnostic is
checkable once the sink is named (D6).

## Recommendation (for manager-spec, iteration 2)

1. Enumerate the eight-pair table in the SPEC and reconcile the count with §B (D1) — this is the
   highest-change-likelihood decision and gates REQ-1/AC-REQ-1a/M1.
2. Make the golden fixtures reachable: commit or vendor the t91 payloads, re-point REQ-6 (D2).
3. Convert the four Out-of-Scope bold lead-ins to H3 headings with bullets (D3) — mechanical.
4. Reword AC-REQ-5a/5b against the M5 validator artifact (D4).
5. Correct Finding D / §F / AC-REQ-5a "silently" wording against the probe's own error item, and
   scope the silence claim to the project-level case (D5).
6. Optional but cheap while editing: D6 (name the diagnostic sink + REQ-4 interaction), D7 (mark
   declared-not-measured classes), D8 (pin the adapter package), D9 (coverage-tie AC-REQ-3b).
