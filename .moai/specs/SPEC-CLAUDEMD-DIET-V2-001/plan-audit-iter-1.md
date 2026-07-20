# SPEC Review Report: SPEC-CLAUDEMD-DIET-V2-001
Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.81** (harmonic mean; Tier M threshold 0.80 — score clears threshold but BLOCKING defect D1 forces FAIL)
skip-eligible (≥ 0.90): **NO**

Audit performed: 2026-07-08 (this session, fresh context — M1 Context Isolation)
Auditor: plan-auditor (independent, adversarial)
Reasoning context ignored per M1 Context Isolation — verdict based solely on spec.md / plan.md / acceptance.md / progress.md + live mechanical verification.
User instruction: "Write your verdict to .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/plan-audit-iter-1.md (create it)."

---

## Must-Pass Results

### MP-1 GEARS format compliance — **PASS**
Evidence:
- 10 REQs (REQ-CMD2-001..010) all match GEARS patterns verbatim:
  - REQ-001 (Ubiquitous) `spec.md:L79` — "The CLAUDE.md file shall have its line count reduced..."
  - REQ-002 (Event-driven) `spec.md:L83` — "**When** ... the diet shall reduce ..."
  - REQ-003 (Capability gate) `spec.md:L85` — "**Where** ... the diet shall NOT collapse ..."
  - REQ-004 (Compound Where+While) `spec.md:L89` — "**Where** ... **While** ... the diet shall create ..."
  - REQ-005 (State-driven) `spec.md:L93` — "**While** both ... exist as byte-identical mirrors ..., the diet shall apply ..."
  - REQ-006 / REQ-007 (Unwanted) `spec.md:L97, L99` — "shall not alter ..." / "shall not remove ..."
  - REQ-008 (Capability gate) `spec.md:L103` — "**Where** a new rule file is created ..., the rule file shall ..."
  - REQ-009 / REQ-010 (Ubiquitous) `spec.md:L107, L111`
- 10 ACs (AC-CMD2-001..010) all carry mechanical verification commands (`wc -l`, `diff`, `grep -c`, `go test`, `awk`).
- AC-CMD2-009 explicitly annotated as run-phase precondition (`acceptance.md:L212`).

### MP-2 §16 SSOT absence — **PASS (manager-spec's central claim INDEPENDENTLY CONFIRMED)**
Live verification commands run (verbatim output):
```
$ find .claude/rules internal/template/templates/.claude/rules -iname '*context*'
.claude/rules/moai/workflow/context-window-management.md
.claude/rules/.moai/state/context-usage.json
internal/template/templates/.claude/rules/moai/workflow/context-window-management.md
$ find ... -iname '*search*'
(no output)
$ grep -rn 'When to Search\|Search Process\|Token Budget\|previous Claude Code sessions\|session index' \
    .claude/rules/moai/ internal/template/templates/.claude/rules/moai/
.claude/rules/moai/workflow/spec-workflow.md:11:| Phase | Command | Agent | Token Budget | Purpose |
(thematic match only — table header in spec-workflow.md, unrelated to §16's "Maximum 5,000 tokens per injection")
```
- The only `*context*` rule file is `context-window-management.md`, which covers context-window **thresholds** (1M = 50%, 200K = 90%) — NOT §16's "When to Search / Search Process / Token Budget for previous-session injection".
- §16's 1st-round pointer line指向 `context-window-management.md` + `session-handoff.md` — both cover **different concerns** (thresholds + paste-ready resume format). The manager-spec's identification of this misdirection (`spec.md:L51`) is correct.
- No rule SSOT exists for §16's actual content. CONFIRMED.

### MP-3 Ceiling arithmetic — **PARTIAL PASS (arithmetic correct; framing overstated — see D3)**
Live per-section line counts (awk output) vs spec.md §F.1 / plan.md §C.4 claims:

| Section | spec.md claim | Actual (awk) | Δ |
|---------|---------------|--------------|---|
| §16 | 48L | 48L | 0 ✓ |
| §15 | 43L | 43L | 0 ✓ |
| §5  | 35L | 35L | 0 ✓ |
| §11 | 19L | 19L | 0 ✓ |
| §14 | 16L | 16L | 0 ✓ |
| §8  | 9L  | 9L  | 0 ✓ |
| §1  | (canonical, ~28) | 27L | -1 |
| §3  | (canonical, ~11) | 10L | -1 |
| §4  | (canonical, ~49) | 48L | -1 |

Reduction arithmetic (plan.md §C.4):
- Opt B: 48+43+35+19+16+9 = 170 baseline; 5+20+15+8+8+5 = 61 target; 170 − 61 = **−109**; 405 − 109 = **296L** ✓
- Opt A: 71 target; 170 − 71 = **−99**; 405 − 99 = **306L** ✓
- Arithmetically correct.

Caveat (D3 below): "200L is impossible" framing is scope-bound, not fact-bound.

### MP-4 No behavior-change disguised as diet + Template-First byte-parity AC — **PASS with D1 caveat**
- REQ-CMD2-006 (`spec.md:L97`) explicitly forbids behavior change.
- AC-CMD2-002 byte-parity `diff` exit 0 — present and mechanical ✓.
- AC-CMD2-005 agent catalog counts + archived-agent grep + decision-tree count — present ✓.
- AC-CMD2-007 template neutrality CI guard — present ✓.
- AC-CMD2-003 `[HARD]` / `[ZONE:*]` portions correct (baselines 14 / 14 verified live).
- **D1 caveat**: AC-CMD2-003's `@import` regex is factually wrong — see defect list.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity in 1-2 places | D3 (200L framing overstates §1/§3/§4 immovability — `spec.md:L55`, `plan.md:L137`); D5 (§2 canonical declaration heading/body mismatch `spec.md:L129-131`) |
| Completeness | 0.85 | 0.75-0.85 — solid with one AC verification gap | All required sections present; 12 frontmatter fields + optional `era`/`tier`; 7 `### Out of Scope —` H3 subheadings each with concrete bullets; D2 gap: AC-CMD2-008 lint invocation fails |
| Testability | 0.70 | 0.50-0.75 — multiple ACs with mechanical defects | D1 (BLOCKING: AC-CMD2-003 `^@import` regex wrong); D2 (AC-CMD2-008 `moai spec lint <SPEC-ID>` ParseFailure); D4 (AC-CMD2-010 baselines off-by-one within ±1 tolerance) |
| Traceability | 1.00 | 1.0 — full bidirectional coverage | `acceptance.md:L262-277` table: every REQ → ≥1 AC; every AC → valid REQ; no orphans |

Aggregate (harmonic mean, per agent-common-protocol § Skeptical Evaluation Stance):
H = 4 / (1/0.75 + 1/0.85 + 1/0.70 + 1/1.00) = 4 / 4.938 = **0.810**

---

## Defects Found

### D1 — **BLOCKING** — AC-CMD2-003 `@import` regex factually wrong; will mis-verify
- **Location**: `acceptance.md:L53-54, L60`; `spec.md:L99, L123, L169`
- **Claim**: "`@import` count == 2 (both trees equal) — user.yaml + language.yaml"
- **Live verification**:
  ```
  $ grep -c '^@import' CLAUDE.md
  0
  $ grep -nE '^@' CLAUDE.md
  210:@.moai/config/sections/user.yaml
  211:@.moai/config/sections/language.yaml
  $ grep -cE '^@\.' CLAUDE.md
  2
  ```
- **Root cause**: CLAUDE.md §9 uses **Obsidian-style `@<path>` embeds** (single `@` prefix), NOT `@import <path>` directives. The 1st-round precedent AC-CMD-004 (inherited here at `spec.md:L123`) appears to have miscalled the syntax, and this SPEC inherited the misnomer verbatim.
- **Impact**: At run-phase, `grep -c '^@import' CLAUDE.md` returns 0. Compared against documented baseline "2", the AC will either (a) FAIL spuriously, OR (b) the run-phase will "discover" baseline 0 and the AC **vacuously passes** while never verifying the two `@.moai/config/sections/*.yaml` embeds are preserved. Either path fails to test what the AC claims.
- **Fix**: Replace `grep -c '^@import'` with `grep -cE '^@\.moai/config/sections/(user|language)\.yaml$'` (baseline: 2). Update `spec.md:L123` and AC-CMD2-003 baselines.

### D2 — **SHOULD-FIX** — AC-CMD2-008 `moai spec lint <SPEC-ID>` invocation ParseFailure
- **Location**: `acceptance.md:L170`
- **Live verification**:
  ```
  $ which moai
  /Users/goos/goos/bin/moai
  $ moai spec lint SPEC-CLAUDEMD-DIET-V2-001
  SEVERITY  CODE          FILE                       LINE  MESSAGE
  ERROR     ParseFailure  SPEC-CLAUDEMD-DIET-V2-001  1     SPEC parsing failed: failed to read file:
                                                         open SPEC-CLAUDEMD-DIET-V2-001: no such file or directory
  ```
- **Root cause**: The `moai spec lint` subcommand treats its argument as a **file path**, not a SPEC-ID. AC-CMD2-008's invocation form does not match the CLI contract.
- **Impact**: AC-CMD2-008 cannot PASS as written; fallback (`go test ./internal/spec/ -run TestSpecLint`) tests the lint engine generically but does not verify this SPEC's compliance. Violates verification-claim-integrity §1.1 surface 2.
- **Fix**: Investigate the correct invocation form (likely `moai spec lint .moai/specs/SPEC-CLAUDEMD-DIET-V2-001/spec.md`, or a different subcommand).

### D3 — **SHOULD-FIX (scope-honesty)** — "200L unreachable" framing overstates §1/§3/§4 immovability
- **Location**: `spec.md:L55, L137` (`§A.4`), `plan.md:L137` (`§D.5`)
- **Claim**: "200라인 목표는 HARD 제약(§1 Core Identity / §3 Command Reference / §4 Agent Catalog는 CLAUDE.md 자체의 canonical home이므로 이동 불가) 하에서는 달성 불가능하다."
- **Counter-evidence**:
  - `.claude/rules/moai/core/moai-constitution.md` exists (287 lines, verified live). §1 ALREADY pointer-izes "Core principles (1-4) and six Agent Core Behaviors" to it (`CLAUDE.md:L29`). The 11-bullet `[HARD]` rules list at `CLAUDE.md:L9-L22` is a SUMMARY of constitution content — theoretically pointer-izable (would collapse §1 from 27L → ~5L).
  - §4 (48L) — the Agent Catalog table is genuinely canonical, but surrounding prose (Selection Decision Tree 9L, Dynamic Team Generation 5L, Archived Agents paragraph 8L, claude-code-guide disambiguation 4L) is at least partly pointer-izable.
  - Conservative additional reduction available outside the SPEC's chosen scope: §1 −22L, §2 −29L, §4 −18L ≈ −69L. 296L (Opt B) − 69L ≈ 227L.
- **Impact**: The 200L SHOULD is scope-bound, not fact-bound. The Out-of-Scope declaration makes this **legitimate scope management**, but §A.4 prose could honestly state "200L is unreachable **within this SPEC's chosen scope**."
- **Fix**: Tighten §A.4 / §D.5 prose.

### D4 — **MINOR** — AC-CMD2-010 §1/§3/§4 baselines systematically off-by-one
- **Location**: `acceptance.md:L224, L228, L232`
- **Claims**: §1 ~28L; §3 ~11L; §4 ~49L. **Actual**: §1 = 27L; §3 = 10L; §4 = 48L.
- Within ±1 tolerance — non-blocking but sloppy.

### D5 — **MINOR** — §2 (Request Processing Pipeline, 34L) canonical-declaration inconsistency
- **Location**: `spec.md:L129-131`
- **Issue**: Out-of-Scope heading reads "Out of Scope — §1/§3/§4 canonical content extraction" (omits §2), but the body lists "§1 Core Identity, §2 Request Processing pipeline, §3 Command Reference, §4 Agent Catalog table" (includes §2). §2 is not a diet candidate so defense-in-depth gap not real risk. AC-CMD2-010 also omits §2.
- **Fix**: Add §2 to heading + AC-CMD2-010 guard, OR remove §2 from body list.

### D6 — **MINOR** — SPEC-V3R5-CLAUDE-REFRESH-001 prose-status mismatch (D7 non-blocking)
- **Location**: `spec.md:L244`
- **Claim**: "superseded scope". **Actual**: `status: completed`.
- D7 reconciliation: `completed` ∉ {retired, superseded, archived} → no BLOCKING.

---

## Independent Verification Commands Run (verbatim output)

### Group A — frontmatter + REQ/AC structural
```
$ grep -c '\[HARD\]' CLAUDE.md           → 14
$ grep -c '\[ZONE:' CLAUDE.md            → 14
$ grep -c '^@import' CLAUDE.md           → 0     ← D1 root cause
$ grep -cE '^@\.' CLAUDE.md              → 2     ← the actual @path embeds
```

### Group B — document structure + section enumeration
```
$ wc -l CLAUDE.md internal/template/templates/CLAUDE.md
   405 CLAUDE.md
   405 internal/template/templates/CLAUDE.md
   810 total

$ diff CLAUDE.md internal/template/templates/CLAUDE.md && echo IDENTICAL
IDENTICAL   (byte-parity confirmed at baseline)
```

### Group C — cross-SPEC reconciliation (D7)
```
SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001: status: completed
SPEC-CCSYNC-CLAUDEMD-001: status: completed
SPEC-STEERING-ALIGN-LOCAL-DIET-001: status: completed
SPEC-RULE-DIET-002: status: completed
SPEC-V3R5-CLAUDE-REFRESH-001: status: completed
```
D7 verdict: **no referenced SPEC is in {retired, superseded, archived}** → no BLOCKING. SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001 confirmed `completed` (not re-opened — collision-check PASS).

### Group D — §16 SSOT absence (MP-2 critical)
See MP-2 section above.

### Group E — moai spec lint invocation (D2)
See D2 above. `moai spec lint <SPEC-ID>` → ParseFailure.

### Group F — per-section line counts (MP-3 arithmetic)
See MP-3 table. All 6 diet-candidate section counts match spec.md §F.1 claims exactly.

---

## Chain-of-Verification Pass (M6 self-critique)

Second-look re-read:

1. **REQ sequencing end-to-end** — REQ-CMD2-001..010 sequential, no gaps, no duplicates (PASS).
2. **Traceability per-REQ** — every REQ mapped to ≥1 AC; every AC mapped to valid REQ; no orphans (PASS, score 1.0).
3. **Out-of-Scope specificity** — 7 `### Out of Scope —` H3 subheadings, each with concrete `-` bullets (PASS).
4. **Contradictions between requirements** — REQ-002 / REQ-003 / REQ-004 are COMPLEMENTARY, not contradictory (PASS).
5. **1st-round inheritance** — AC-CMD2-009 inheritance from 1st-round AC-CMD-009 plausible. HOWEVER the inherited "`@import` token-neutrality honesty" in 1st-round title suggests D1 root cause predates this SPEC.
6. **D8 (syscall) auto-check** — SPEC body does not mention `syscall`; D8 auto-PASS.
7. **Test run of AC-CMD2-008** — `moai spec lint` invocation fails (D2). NOT TESTED by author.

**New defects in second pass**: none beyond the 6 listed.

---

## Recommendation

**Verdict: FAIL** — iteration 1. BLOCKING defect D1 must be fixed before run-phase entry.

### Required fixes for iteration 2 PASS

1. **(BLOCKING, D1)** Fix AC-CMD2-003 `@import` regex → `grep -cE '^@\.moai/config/sections/(user|language)\.yaml$'` (baseline: 2). Update `spec.md:L99, L123, L169` prose from `@import` → `@path embed`.
2. **(SHOULD-FIX, D2)** Fix AC-CMD2-008 invocation — investigate correct `moai spec lint` form.
3. **(SHOULD-FIX, D3)** Tighten 200L framing: "unreachable **within this SPEC's chosen scope**".
4. **(MINOR, D4)** Update AC-CMD2-010 baselines: §1 → 27, §3 → 10, §4 → 48.
5. **(MINOR, D5)** Resolve §2 canonical-declaration inconsistency.
6. **(MINOR, D6)** Change "superseded scope" → "completed scope" for SPEC-V3R5-CLAUDE-REFRESH-001.

### What is already correct (do NOT re-open)

- §16 SSOT absence (MP-2) — independently confirmed.
- Per-section line-count arithmetic (MP-3) — adds up.
- Template-First byte-parity discipline — sound.
- 1st-round collision avoidance — confirmed completed.
- 3rd-round Out-of-Scope declaration — legitimate scope management, NOT hidden debt.
- All 10 REQs use canonical GEARS notation (MP-1).

### skip-eligibility

**NOT skip-eligible** (aggregate 0.81 < 0.90). Phase 0.5 plan-auditor MUST re-run on iteration 2 after D1 fix.

Note: skip-eligibility governs ONLY Phase 0.5 verdict re-execution. Implementation Kickoff Approval (plan→run HUMAN GATE) remains mandatory regardless — CLAUDE.local.md §19.1, orchestration-mode-selection.md §C.3.

---

## Iteration 2 entry conditions

Re-audit after manager-spec addresses D1 (mandatory) and ideally D2-D6. Predicted post-fix score: ~0.88-0.92 if D1 and D2 both fixed (Testability → 0.90+, Completeness → 0.95, Clarity → 0.85). Tier M threshold 0.80 — likely PASS on iteration 2.
