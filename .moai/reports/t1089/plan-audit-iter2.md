# SPEC Review Report: SPEC-MODEL-OPUS55-001
Iteration: 2/2 (Tier M ceiling reached)
Verdict: PASS-WITH-DEBT
Overall Score: 0.87 (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. The coordinator's list of claimed resolutions was treated as a set of claims to verify with tools, and none was accepted without measurement.

Audited tree: `.claude/worktrees/t1089`, HEAD `98a9a55a3`, whose parent is `c4de30473` (the iter-1 report). `git diff --stat c4de30473 98a9a55a3` shows 5 files: the 4 SPEC artifacts and `.moai/reports/t1089/p4v2-rednow-6e75b74db.txt`. The scope is the D1–D15 delta plus any regression the repair introduced.

## Must-Pass Results

- [PASS] MP-1: 16 REQs (`grep -oE '^- \*\*REQ-OP55-[0-9]{3}' spec.md | wc -l` → 16) and 16 ACs (AC-009 split into a/b/c under a single number), sequential with no gaps.
- [PASS] MP-2: judged on the requirement layer. REQ-007 (spec.md:L58), REQ-009 (L63) and REQ-010 (L64) as rewritten are ubiquitous "shall" statements. `moai spec lint` → `✓ No findings`.
- [PASS] MP-3: `version: "0.1.1"` (quoted) and `updated: 2026-09-23`; the 12 fields are intact. `mcp__moai__spec_audit(project_root=<worktree>)` → `modern_era_clean: 1`.
- [N/A] MP-4: single-language scope, as in iter-1.
- [PASS] MP-5: the SPEC cross-references are unchanged since iter-1, so there is no BLOCKING D7 finding.
- [PASS] MP-6: no `syscall`.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-MODEL-OPUS55-001/` → rc=1. The literal is gone and plan.md:L214 now reads "0 open clarification markers".

## Category Scores

| Dimension | Score | Evidence |
|---|---|---|
| Clarity | 0.85 | REQ-007 (spec.md:L58) now matches `resolveLaunchEffort`, and REQ-009 (L63) is pinned to one line. The remaining weak point is N1: the C.6 row 9 KEEP rationale is internally contradictory. |
| Completeness | 0.90 | `.moai/project` is in scope (spec.md:L41, plan.md:L202), K6 is present (plan.md:L195), and the 14-row conflict table is present (plan.md:L151-L170). |
| Testability | 0.85 | Every re-run probe reproduces its RED-now value. The per-locale regexes are anchored. P6 can false-fire on a correct rewrite (N2). |
| Traceability | 0.90 | The acceptance.md:L7-L24 table covers all 16 REQs, and REQ-008's two halves are both exercised (AC-009a/b/c). |

The aggregate is 0.87, above 0.80. All 15 iter-1 defects are closed. One new SHOULD-FIX (N1) comes from the repair itself, which is why the verdict is PASS-WITH-DEBT and not PASS.

## Regression Check (iter-1 defects)

| ID | Status | Evidence (re-measured on `98a9a55a3`) |
|---|---|---|
| D1 | CLOSED | See D1 note below the table. |
| D2 | CLOSED (with N1) | See D2 note below the table. |
| D3 | CLOSED | plan.md §C.8 picks Option A: one English label `Opus 5.5 (Recommended)` in all 4 locales. `TestModelOptLabelsEnglishUnified` exists at console_ux_fix_test.go:197, and updating it is named in M2 (plan.md:L201). AC-007 pins count `4` plus the companion `"Opus 5"`: RED-now `4`, Green `0`. |
| D4 | CLOSED | AC-006b spans 9 files, and each returned `:0` when re-run (RED-now as claimed). P5 → rc=1 now. The P5 regex matches the mutant `Opus 5.5 defaults to \`effort: high\``. |
| D5 | CLOSED | AC-009a runs `TestGetProfileText_OpusAliasValues` (internal/cli). AC-009b runs `TestModelPolicyDescsAgreeWithProfileMatrix`, which exists at wizard/model_policy_matrix_agreement_test.go:23 and reads the translation store (L77-80). Each has its own mutant. |
| D6 | CLOSED | The P4 v2 roots include `.moai/project`. Its 6 new lines over v1 are `.moai/project/product.md:151`, `tech.md:13/14/15`, and `model_policy.go:54/78`, verified with `comm -13`. M3 lists the project docs. |
| D7 | CLOSED | REQ-009 is scoped to the `- opus = …` line in both model-policy.md copies. P7 awk → `0` now; its green value is `2`. |
| D8 | CLOSED | K6 is at plan.md:L195 and cites launcher.go:1139-1152. The §C heading at spec.md:L83 was reworded to "runtime changes other than the alias target", and L85 discloses the `--model claude-opus-5-5[1m]` change. |
| D9 | CLOSED | `grep -in 'four golden\|4 golden'` → no hit. Three files are named, along with `-update-golden` (plan.md:L184). |
| D10 | CLOSED | AC-008 anchors on `"[^"]*…"` with one command per locale marker. AC-007 uses exact per-locale strings, which match the current labels `Medium`/`중간`/`中`/`中` at i18n.js:443/1337/2116/2895. |
| D11 | CLOSED | See D11 note below the table. |
| D12 | CLOSED | REQ-015 (spec.md:L72) excepts `.moai/reports/t1089/`. The AC-014 command `git -C <wt> diff --name-only develop...HEAD -- … .moai/reports ':!.moai/reports/t1089'` → empty. |
| D13 | CLOSED | MP-7 grep → rc=1. |
| D14 | CLOSED | AC-009c adds the alias→`claude-opus-6` mutant, and both guards are expected to fail. |
| D15 | CLOSED | plan.md §C.1 (L132) states the `ModelIDOpus48` precedent, the card-wording reading, and the doc-comment update. AC-003 notes that it catches the doc comment. |

**D1.** REQ-007 now says "model-policy effort when a model policy is stored, otherwise Claude Code's own model default (`medium` on Opus 5.5)" and forbids an unconditional `medium`.
- `TestResolveLaunchEffort` is at launcher_test.go:900. Its five subtests are "explicit effort wins over model_policy", "model_policy fallback high/medium/low", and "both empty no override" (L907-L911).
- AC-007 pins the wording per locale (`model policy`/`모델 정책`/`モデルポリシー`/`模型策略` followed by `Opus 5.5`). The `opt.runtime_default` key has one consumer, schema.go:374.

**D2.** plan.md §C.6 is a 14-row table.
- I re-read the cited lines and each matches the quoted text: model-policy.md:157/234/240, agent-authoring.md:61/83/219, prompting-best-practices.md:35, session-handoff.md:88, verify-judge-effort-contract.md:11, CLAUDE.md:143, profile_matrix.go:280, and dynamic-workflows.md:135-138.
- P6, re-run: 11 lines, rc=0. That is identical to plan.md:L76-L86 and includes dynamic-workflows.md:126 in both copies and tech.md:15.
- dynamic-workflows is a SAME pair (`cmp` rc=0), and AC-010 now covers six pairs.
- One of the table's decisions is wrong; see N1.

**D11.** P4 v2 re-run → 153 lines, rc=0. The sorted diff against `p4v2-rednow-6e75b74db.txt` is **IDENTICAL**. I tested it against a planted file:
- `Opus 5 (superseded prompts) is the default` → now caught.
- `Opus5 is default` → now caught.
- `// superseded` deprecated row → exempt.
- `ModelIDOpus55` / `claude-opus-5-5` / `Opus 5.5` → silent.
- The new roots `.claude/{agents,commands,output-styles}` and `AGENTS.md` are included.
- Remaining limit: a line carrying the literal `(superseded)` anywhere is still exempt. See N3.

## Defects Found (new, introduced by or surfaced through the repair)

N1. C6-ROW9-CONTRADICTION: plan.md:L165 (§C.6 row 9) against .claude/rules/moai/development/prompting-best-practices.md:35 (template copy byte-identical, SAME pair). Severity: SHOULD-FIX (major). Class: blocking.
- Row 9 marks prompting-best-practices.md:35 as KEEP, with the reason "consistent with a `medium` session default". The line actually reads "`medium`/`low` only for speed-critical or simple tasks".
- That directly contradicts row 1's new constitution text, "moai's recommended session effort is `medium`", and the REQ-OP55-005 wizard/web medium recommendation. This is the exact high/xhigh-vs-medium conflict class that card requirement (3) asks the SPEC to resolve.
- P6 does not match this line, so no AC would catch it.
- agent-authoring.md:219 ("xhigh for coding/agentic, minimum high for intelligence-sensitive") in the same row is per-task escalation and is compatible. The KEEP decision is wrong only for prompting-best-practices.md:35.
- Required fix: split row 9. Mark prompting-best-practices.md:35 as REWRITE, in both copies, which are a SAME pair: drop "only for speed-critical or simple tasks" and say `medium` is the recommended session default, with `high`/`xhigh` raised per task. Add the file to the AC-006b positive list, or add a negative grep `'medium`/`low` only for'` → empty. This is a one-row plan edit, and it can be folded in before Implementation Kickoff without another audit iteration.

N2. P6-FALSE-FIRE-ON-CORRECT-TEXT: acceptance.md:L75, plan.md:L75. Severity: MINOR. Class: optional.
- P6's green state requires that none of the literal shapes `high: the default` or `defaults to \`effort: high\`` remain anywhere.
- Correct rewrites that REQ-010 itself asks for would still trip it, for example "`high`: the default on every model except Opus 5.5" or "Sonnet 5 defaults to `effort: high`". The run phase must word around the probe.
- Required fix: state in AC-006c that rewritten lines must avoid these literal shapes. Alternatively, allow shapes followed by an explicit Opus 5.5 exception on the same line.

N3. P4V2-PAREN-SUPERSEDED-ESCAPE: plan.md:L163 (§C.6 row 7), acceptance.md:L56. Severity: MINOR. Class: optional.
- `(superseded)` anywhere on a line still exempts it: a planted `Opus 5 is the default (superseded)` passes.
- Row 7 uses `(superseded)` to tag model-policy.md:234 ("On Opus 5, `low` and `medium` are stronger…"). That is a still-true vendor statement about Opus 5, not a superseded one.
- Required fix: add the escape-hatch limit to AC-004's "Known limit" line. Optionally, use `measured on Opus 5` or an explicit "(Opus 5 statement)" style for row 7 rather than `(superseded)`.

N4. EMPTY-OPTION-REFERS-TO-HIDDEN-FIELD: plan.md:L174 (C.7). Severity: MINOR. Class: optional.
- The new web label tells the user that "model policy effort if set" wins. However, `model_policy` has no web field: it was removed from the console per schema.go:350-352 and is set only through the TUI wizard.
- The wording is truthful, which is what REQ-007 requires, but a web user cannot see or change the input it names.
- Required fix: none required for correctness. Optionally, add "(set in `moai profile setup`)" or equivalent to the label text in C.7.

## Recommendation

PASS-WITH-DEBT at the Tier M ceiling. Every iter-1 defect D1–D15 is closed, with tool-reproduced evidence.

- The carried debt is N1 (major, blocking-class). It is a one-row correction to plan.md §C.6. It should be applied before the Implementation Kickoff Approval gate, or be written into the run-phase dispatch as a mandatory M3 item with its own AC.
- N2–N4 are optional wording and limit notes.
- Per the Retry Loop Contract, no iteration 3 is scheduled. If the orchestrator judges that N1 must be re-verified by an auditor, that is the operator's ceiling-extension decision.

## Gaps

- No Go tests were executed; this is a plan-phase audit. The existence of test names and subtests was checked by grep/read only.
- The AC-009c claim that the failure output prints the derived token `6` depends on the guards' message format after the K2 rewrite, which is not yet written.
- No cross-model backend was run. No `audit_model` key is configured.
