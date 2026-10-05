# SPEC Review Report: SPEC-HOOK-MATCHER-POWERSHELL-001
Iteration: 2/2 (Tier M ceiling per harness.plan_audit_tier_ceilings — FINAL)
Verdict: PASS
Overall Score: 0.87 (Tier M PASS threshold 0.80; iter-1 0.77 → iter-2 0.87, no regression, no STOP)

Card: t1224. Tree audited: worktree `.claude/worktrees/t1224`, branch `WT-hook-matcher-powershell`, HEAD `0f4a8e9c0` (SPEC v0.2.0), working tree clean. Revision delta: `git diff --stat e609f57b0 0f4a8e9c0` touches only the four SPEC files plus the iter-1 report. Develop-side facts were measured against local `develop` = `35ab8cff3`.

Reasoning context ignored per M1 Context Isolation. The invocation's focus list was used only to prioritize checks.

Cross-model second opinion: not run. `grep -rn audit_model .moai/config/sections/` exited 1, so no key is set and this is a Claude-only audit. `mcp__moai__spec_audit` (project_root = this worktree, filter = this SPEC) returned `drift_findings: []` and `modern_era_clean: 1`. `moai spec lint SPEC-HOOK-MATCHER-POWERSHELL-001 --json` returned `[]`.

The verdict is PASS because every must-pass criterion holds, the score is above threshold, and all 14 iter-1 defects are resolved. Three new minor blocking text defects (N1–N3) must still be fixed before run. They are one-line corrections that the orchestrator can check by diff, and none of them touches a must-pass criterion.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-HMP-001..016 each appear once, sequentially, with zero padding. `grep -c '^- \*\*REQ-HMP'` returns 16, which is at the Tier M ceiling of 16 but not over it. There are 13 ACs (`grep -c '^### AC-HMP'` returns 13), within the 16 ceiling.
- [PASS] MP-2 GEARS compliance, judged on the requirement layer (spec.md §B) only. Every entry uses a GEARS form:
  - Ubiquitous: L80, L81, L86, L103, L104, L105, and L109 (an embedded When clause inside a ubiquitous frame).
  - Unwanted, "shall not": L82 and L87.
  - Event-driven: L91, L94, L95, L99, and L110.
  - Where+When compound: L92 and L93.

  The acceptance.md Given-When-Then scenarios are verification-layer entries and were graded under Group 4, not here.
- [PASS] MP-3 frontmatter: spec.md:L2–L13 carry all 12 canonical fields with correct types. `version: "0.2.0"` is quoted, `created` and `updated` are ISO dates, `priority` is P1, `lifecycle` is spec-anchored, and `tags` is a comma string. No rejected aliases appear.
- [N/A] MP-4: the SPEC concerns the Bash and PowerShell host tools, not the 16 programming-language toolchains.
- [PASS] MP-5 D7: SPEC-POWERSHELL-DENY-PARITY-001 has `status: completed` in this tree. SPEC-HOOK-STDIN-FAILCLOSED-001 has `status: completed` on develop, and its REQ-HSF-001 (develop spec.md:179) matches the SPEC's citation. SPEC-DUAL-HARNESS-HOOK-PARITY-001 is in-progress on its branch. None of the three is retired, superseded, or archived. The two SPECs not in this tree are now located in §C (L115–L116), which resolves the iter-1 D14 SHOULD.
- [PASS] MP-6 D8: `grep -c syscall` returns 0 in all four SPEC files.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION'` over the SPEC dir exited 1. research.md is absent (Tier M).

## Category Scores (0.0–1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.82 | 0.75 band, upper | D2/D4/D6 ambiguities are gone: hook identity is defined at spec.md:L103, the envelope vs `tool_input` split at L94, and no-decode at L95. Residual: REQ-HMP-005's scope (L87) reaches test files (N1). The encoded-command spelling set (L95) is asserted without a source (N3). The integration-lock log root is unstated (plan.md:L10, N5). |
| Completeness | 0.92 | 1.0 band minus minor | The site inventory is now complete. On base and on develop, the AST-relevant non-test `"bash"` literals are exactly pre_tool.go ×4, post_tool.go ×2, evidence_writer.go ×3, and cli/hook.go ×1 (L788; develop L822). normalize.go:65/74 and cli/hook.go:662 (develop L696) are comments. The only wrapper naming `"Bash"` is the pre-tool wrapper, now site 14 (L74). All sections are present, including six `### Out of Scope —` H3s. |
| Testability | 0.82 | 0.75 band, upper | AC-HMP-012 now has a FAIL branch (acceptance.md:L122). AC-HMP-011 is static, with a positive control and swept count ≥1 (L111–L114). AC-HMP-002 carries a non-empty control (L36). Residual: AC-HMP-012's first Then is unconditional and cannot hold on the INCONCLUSIVE path (N2). AC-HMP-003's wrong-script mutant cannot be built on the real template (N4). The arm C provenance check admits a `-dirty` build (N6). |
| Traceability | 0.95 | 1.0 band | All 16 REQs map to ≥1 AC in the matrix (acceptance.md:L9–L21). All 13 ACs have a Given-When-Then body. No AC cites a nonexistent REQ. Minor: AC-HMP-011 verifies only the MOAI_HOME and no-Parallel halves of REQ-HMP-014, not the `.moai/state` clause (N7). |

Aggregate (harmonic mean) = 4 / (1/0.82 + 1/0.92 + 1/0.82 + 1/0.95) ≈ 0.87.

## Regression Check — iteration-1 defects

| ID | Iter-1 defect | Status | Evidence |
|---|---|---|---|
| D1 | Shell wrapper `tool_name` branch missed | RESOLVED | spec.md:L74 adds site 14, class (b). The template tmpl:44 and local .sh:39 locations match my grep. REQ-HMP-002 (L81) owns wrapper behavior via D6, and REQ-HMP-013(ii) (L104) scans `*.sh.tmpl`/`*.sh`. AC-HMP-013 has been added. Only one wrapper names `"Bash"` (measured), so the scan's base count is 1. |
| D2 | REQ-HMP-009 contradicts the landed fail-closed dispatcher | RESOLVED | L94 now covers only the parseable-envelope case and defers to REQ-HSF-001. Develop `internal/cli/hook.go:288-296` confirms `answerStdinParseFailure` runs before any tool name is read. `HookInput.ToolInput` is `json.RawMessage` (types.go:217). AC-HMP-008 is split into two concrete cases (L79–L86). |
| D3 | LIVE had no FAIL branch; arm C binary unattributable | RESOLVED (residual N6) | FAIL is defined in REQ-HMP-015 (L109), AC-HMP-012 (L122), and plan.md:L81. Arm C pins PATH and records `binary-version.txt` (plan.md:L77, L79). `moai version` prints the BuildID from `git describe --tags --dirty` (Makefile:19; observed `moai_cp/20260925_122548-14-ga8a9b9376`), so a `-g<sha>` token exists. |
| D4 | Hook identity undefined for the `command: "bash"` shape | RESOLVED (residual N4) | L103 defines the identity as the final `args` element with `${CLAUDE_PROJECT_DIR}/` stripped. Template L53 has final element `${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-pre-tool.sh`, so the definition yields `.claude/hooks/moai/handle-pre-tool.sh`, consistent with AC-HMP-003 L43. |
| D5 | D2 option (A) cited a nonexistent integration-lock audit log | RESOLVED (residual N5) | plan.md:L10 names each guard's target. `grep -n -i 'audit\|\.jsonl\|\.log' internal/hook/integration_lock_guard.go` returns 0 lines, confirming the "creates a new log" framing. AC-HMP-009 L94–L96 adds the integration-lock case. |
| D6 | `-EncodedCommand` trigger undecidable | RESOLVED (new N3 on the spelling set) | L95: "unclassifiable regardless of its payload and shall not be decoded". AC-HMP-009 L97 adds a no-decode probe. |
| D7 | Payload capture ordered after M3; t1211 evidence uncited | RESOLVED (new N2 on the INCONCLUSIVE path) | Arms A and B moved into M0 (plan.md:L28). §A.1 L33 cites the t1211 evidence and labels it transcript-level. `.moai/reports/t1211/m1/A.jsonl` and `B.jsonl` exist and are tracked (`git ls-files`). |
| D8 | AC-HMP-011 vacuous under a process-level MOAI_HOME | RESOLVED | §C L119 no longer exports MOAI_HOME. AC-HMP-011 is a static `go/ast` check with a positive control and swept ≥1 (L107–L114). The no-`t.Parallel` rule is stated at L105 and plan L49. |
| D9 | `BASE` undefined | RESOLVED | acceptance.md:L3 defines `BASE` as `$(git merge-base develop HEAD)`, evaluated before merge, and acknowledges post-merge vacuity. The non-empty control is at L36. |
| D10 | AC-002/007/011 lacked scenarios | RESOLVED | Scenarios are at L32–L37, L69–L77, and L107–L114. |
| D11 | Source guard evadable by constant or EqualFold | RESOLVED | L104 flags any `bash` literal in any case. plan.md:L42–L43 adds an AST BasicLit scan plus both positive controls. On over-breadth: every lowercase `"bash"` literal on develop is in `_test.go` files, so the non-test scan has no spurious hits today. |
| D12 | Slot-lease indirection behavior unstated | RESOLVED | spec.md:L64, plan.md:L13 (D2-slot row), and REQ-HMP-010 (L95) all state it. |
| D13 | LIVE isolation details | RESOLVED (residual folded into N6) | plan.md:L77 sets `MOAI_HOOK_STDERR_LOG`, and the wrapper allowlist (tmpl:6-8) admits `$CLAUDE_PROJECT_DIR/.moai/logs/`. Arm B now expects ≥1 payload. `MOAI_BRANCH_GUARD_EXEMPT` must be unset (plan.md:L79). |
| D14 | D7 SHOULD: referenced SPECs absent from this tree | RESOLVED | §C L115–L116 name each SPEC's location. |

No stagnation, and no defect is unresolved.

## Defects Found (structured defect-list)

N1. REQ-SCOPE-FALSE-AT-ARRIVAL — spec.md:L87 (REQ-HMP-005) vs L104 (REQ-HMP-013), plan.md:L47 — REQ-HMP-005 says "the hook package and the hook CLI shall not carry a string literal equal to `bash` in any letter case outside that shared predicate". A Go package includes its `_test.go` files. On develop, `internal/hook` test files carry well over 100 `"Bash"`/`"bash"` literals (git grep, e.g. pre_tool_test.go, normalize_test.go, wrapper_test.go). M3 also requires new Bash-control tests that must name `"Bash"`. REQ-HMP-013 and AC-HMP-004 correctly scope the check to non-test files, so REQ-005 is wider than its own verifier, false on arrival, and contradicted by M3 (verification-completeness.md §3). The v0.2.0 rewording from "compare against the literal" to "carry a string literal" introduced this. — Severity: minor — Class: blocking — Required fix: insert "in non-test source files" into REQ-HMP-005, matching REQ-HMP-013(i).

N2. INCONCLUSIVE-PATH CONTRADICTION — acceptance.md:L120 (AC-HMP-012 first Then), plan.md:L47 (M3 precondition), plan.md:L28, acceptance.md:L146 (DoD) — The DoD and REQ-HMP-015 accept INCONCLUSIVE as a deliverable LIVE outcome. However:
  - AC-HMP-012's first Then ("the arm B payload's `tool_name`, `tool_input` keys, and `tool_response` keys are recorded … before M1 begins") is unconditional, so it cannot hold when arm B is INCONCLUSIVE.
  - M3's precondition, "M0 payload capture recorded (REQ-HMP-016)", reads as blocking M3 on the same path.
  - M0's stop rule covers only a captured-but-wrong payload, not a missing capture.

  The run phase therefore cannot tell whether an INCONCLUSIVE arm B blocks M1 through M3 or lets them proceed on fail-safe handling (§A.1 L33 implies proceed). — Severity: minor — Class: blocking — Required fix: make AC-HMP-012's first Then conditional ("when arm B captured a payload"), and state in plan M0/M3 that an INCONCLUSIVE arm B lets M1–M3 proceed on REQ-HMP-009 fail-safe handling, with no live-path claim for REQ-HMP-006/011.

N3. UNCITED ALIAS SET — spec.md:L95 (REQ-HMP-010), plan.md:L12, acceptance.md:L138 — The encoded-command trigger set is fixed as "`-EncodedCommand` or its short forms `-enc`, `-ec`, `-e`", with no source cited. The M0 measurement (plan.md:L26) tests only whether the guard parser classifies these constructs. It does not test which spellings pwsh itself accepts. PowerShell's command-line parameter binding is commonly documented to accept unambiguous prefixes, which would admit spellings such as `-en`, `-enco`, or `-encoded`, and possibly `--`/`/` switch prefixes. That is a hypothesis: this audit could not measure it, because `pwsh` is present (`command -v pwsh` → `/usr/local/bin/pwsh`) but running it was refused by the worktree-isolation guard (see Gaps). If the set is incomplete, an unlisted spelling reaches neither the parser nor D2. Under D2=(B) that is a silent deny bypass; under (A) it is an unlogged allow. — Severity: minor — Class: blocking — Required fix: either define the trigger structurally ("any argument that the invoked pwsh/powershell resolves to the EncodedCommand parameter"), or add an M0 step that measures the accepted spellings against the installed pwsh and records them, then pin the M4 test to the measured set.

N4. MUTANT NOT CONSTRUCTIBLE ON THE REAL TEMPLATE — acceptance.md:L46, plan.md:L35 — In the template, every PreToolUse registration points to `handle-pre-tool.sh` (L53, L64, L75, L86). No "different PreToolUse script" exists to host the wrong-script mutant. The plan's example moves the registration onto `handle-harness-observe.sh`, which is a PostToolUse, opt-in-only entry (tmpl:130-141), so the event key catches it, not the script key. — Severity: minor — Class: optional — Required fix: state that mutant (m2) runs on a synthetic settings fixture carrying a second PreToolUse script.

N5. LOG ROOT UNSTATED — plan.md:L10, acceptance.md:L95 — `.moai/logs/integration-lock-audit.log` does not say which root it sits under. The branch-guard log is projectDir-relative (branch_guard.go:47). The integration-lock guard fires inside the release worktree, which may differ from the primary checkout. On scope creep: the new file appears only under D2=(A), and its cost is disclosed to the operator, so it is acceptable. — Severity: minor — Class: optional — Required fix: name the root (for example, the handler's projectDir, like the branch guard).

N6. ARM C PROVENANCE AND ISOLATION RESIDUALS — plan.md:L77, L79 — Three residuals remain:
  - A binary built from a dirty tree prints `…-g<sha>-dirty`, so its `-g<sha>` still equals HEAD even though it contains uncommitted code.
  - A HEAD sitting exactly on a tag prints no `-g` token, which forces INCONCLUSIVE.
  - `MOAI_HOOK_STDERR_LOG` survives the wrapper allowlist only if `<arm>` string-matches `$CLAUDE_PROJECT_DIR`. A `/tmp` vs `/private/tmp` spelling mismatch on macOS would silently fall back to the real `$HOME/.moai/logs`.

  — Severity: minor — Class: optional — Required fix: add validity conditions: "BuildID carries no `-dirty`", and "arm C's stderr log exists under `<arm>/.moai/logs/`". Build the scratch path from `pwd -P`.

N7. PARTIAL VERIFIER — acceptance.md:L109–L114 vs spec.md:L105 — AC-HMP-011's list of modified tests is self-reported (progress.md §E.2), so a modified test left off the list escapes. REQ-HMP-014's "shall not read or write the repository's `.moai/state`" clause has no check; it rests on fixture design alone (temp git repos). — Severity: minor — Class: optional — Required fix: derive the modified-test set from `git diff $BASE..HEAD -- '*_test.go'`. Optionally, state that the `.moai/state` clause is covered by the temp-repo fixture requirement of AC-HMP-006/007.

Must fix before run: N1, N2, N3. All three are text edits to spec.md, acceptance.md, or plan.md. The orchestrator can verify them by diff without a third audit iteration, since Tier M has no iteration 3. N4–N7 are left to the orchestrator's discretion.

## Gaps (not observed in this audit)

- pwsh encoded-command spelling acceptance: the measurement (`pwsh -NoProfile -NonInteractive <flag> <b64>` over 10 spellings) was refused by the worktree-isolation guard both times it was attempted. N3 therefore rests on the SPEC lacking a citation, not on a measured defect.
- The hook payload shape for PowerShell remains unmeasured by design. It is deferred to M0 arms A and B.
- The AC-HMP-011 checker, the AST guard, and the parity test are specifications, not code. Their failure modes are judged from the text only.

## Recommendation

PASS (0.87 ≥ 0.80; MP-1 through MP-7 pass or are N/A with evidence above; all 14 iter-1 defects resolved). Before Implementation Kickoff, manager-spec applies:

1. N1: spec.md:L87, add "in non-test source files".
2. N2: acceptance.md:L120, make the Then conditional. In plan.md M0/M3, state that an INCONCLUSIVE arm B lets M1–M3 proceed on fail-safe handling with no live-path claim.
3. N3: spec.md:L95, plan.md:L12, and acceptance.md:L138, replace the enumerated spellings with a structural definition, or add an M0 pwsh spelling measurement.

Decision ownership (plan.md Open Decisions):

- **Operator:** D1 (matcher form), D2 (unclassifiable policy), D4 (sibling cards: `IsWriteOperation` parity, and hooks on Windows without Git Bash), and D6 (wrapper Risk-Amplifier on PowerShell; this changes distributed template behavior).
  - D2's options should be presented after N3 is fixed, because the deny-bypass exposure under (B) depends on the spelling set.
  - Option (A)'s cost includes one new log file for the integration lock, and its root should be named (N5).
- **Lead:** D3 (PostToolUse sites 5–6 converted for parity only) and D5 (LIVE arm C). D5 is now decidable: the provenance controls exist, and the N6 refinements are optional.
