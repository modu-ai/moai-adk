# SPEC Review Report: SPEC-POWERSHELL-DENY-PARITY-001
Iteration: 2/2 (Tier M ceiling = 2, final)
Verdict: PASS
Overall Score: 0.89 (Tier M PASS threshold 0.80; harmonic mean of the four dimensions)

Reasoning context ignored per M1 Context Isolation. Audited tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1211`, HEAD `cb5c733fe`, branch `WT-powershell-deny`, SPEC v0.2.0. Scope: resolution of iter-1 D1–D15 plus defects introduced by the revision. Cross-model backends not invoked (no `audit_model` key, same as iter-1).

This is a conditional PASS. All must-pass rows pass, all ten iter-1 blocking defects are resolved, and the score is above the threshold. The revision introduced two new blocking-class minor defects (N1, N5). Both must be fixed before M1 runs. Each is a one-sentence edit that a grep can check (see Recommendation).

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: REQ-PSD-001..018 contiguous and unique (extracted `REQ-PSD-[0-9]{3}**` from spec.md L68–L94: 001…018 in order).
- [PASS] MP-2 GEARS (requirement layer only): all 18 REQs are Ubiquitous ("The run-phase shall", L68/L69/L75), When (L71–L74), Where (L70, L79–L82, L86–L87), or shall-not (L93–L94) forms. ACs are graded under Group 4, not here.
- [PASS] MP-3 frontmatter: 12 canonical fields present (spec.md L2–L13), `version: "0.2.0"` quoted, `updated: 2026-09-26`. `moai spec lint` → "✓ No findings"; `spec_audit` (project_root = worktree) → `drift_findings: []`.
- [PASS] MP-4 language neutrality: the axis is OS shells, not the 16 programming languages. REQ-PSD-015 (L88) keeps the neutrality and leak tests binding.
- [PASS] MP-5 D7: the only foreign reference is SPEC-V3R6-TOOL-POLICY-SSOT-001, whose status is `completed`.
- [PASS] MP-6 D8: `grep -c syscall` = 0 in all four files.
- [PASS] MP-7 clarification gate: no `[NEEDS CLARIFICATION` in plan.md; research.md is absent (Tier M).

## Regression Check — iter-1 defects

| ID | iter-1 defect | Status | Evidence |
|---|---|---|---|
| D1 | Built-in Remove-Item protection omitted | RESOLVED | spec §A.2 L34–L43 cites the built-in protection. Quotes re-verified against the fetched `permission-modes.md`: L668 system paths, L669 wildcards, L672 `cmd` check (v2.1.283+), L676 `CLAUDE_CODE_DISABLE_POWERSHELL_CMD_RM_DENY=1`, L543 bypassPermissions. Scope is reframed to the 38-row residual set (§A.3); the filesystem 9 are excluded as `builtin` (plan §C.2); D1(d) states that it adds nothing for the documented cases (plan L145). |
| D2 | Inconclusive outcome incompletely defined | RESOLVED | plan L51–L68: V1–V6 plus an outcome table whose "any other combination → INCONCLUSIVE" row closes the table; REQ-PSD-005 L72 names the failed controls. Arm B deleting the target now fails V4, and arm C being blocked fails V5. |
| D3 | Tool-route confound (Bash fallback) | RESOLVED | `--tools PowerShell` (plan L35); V2 requires `PowerShell` present and `Bash` absent in `system/init`; the stream-json denial record decides and directory state only corroborates (plan L70). |
| D4 | User-scope settings and hooks not isolated; pwsh undeclared | RESOLVED | `--setting-sources project --safe-mode --strict-mcp-config` (plan L34); REQ-PSD-004 (L71) makes pwsh ≥ 7 a precondition; managed-settings presence is recorded (plan L26). Flags verified in `claude --help` on 2.1.283 (see new-claims table). |
| D5 | Vacuous `settings.json.tmpl` pathspec | RESOLVED | `TMPL` is defined with the full path (acceptance L5). AC-PSD-012 is a set comparison with a positive control (L22), and AC-PSD-014 uses full paths (L24). |
| D6 | Over-block test has no positive control | RESOLVED | plan §C.3 L89–L98 names the modeled semantics and lists known-deny controls, with an always-false-matcher mutation. Alias- and compound-dependent samples are recorded as a Gap (AC-PSD-008 L18). |
| D7 | S3 not attributable (system-path target) | RESOLVED | S3 is now the inconclusive scenario (L38–L41). The built-in observation moves to arm D / S4 and is explicitly not attributed (L43–L46). The measurement targets a residual command, `git clean -fdx` (plan L41). |
| D8 | Defective `Remove-Item * C:/:*` shape | RESOLVED | Withdrawn (plan L87). REQ-PSD-011 L81 forbids a middle `*` combined with `:*`, and mutation (d) of AC-PSD-010 tests it. |
| D9 | Branch contradiction | RESOLVED | Branch P/N split (spec L64): REQ-009..014 are "Where Branch P is taken", REQ-007 defines the Branch N deliverable, REQ-016 is branch-specific, the runtime guard is dropped (§C L100), and the acceptance Branch column carries an N/A rule (L3). |
| D10 | Guard is open-world | RESOLVED | REQ-PSD-014 L87 and plan §C.4 make the guard closed-world (every Bash rule is either mapped or on the exclusion list, and every PowerShell rule maps to a Bash rule). AC-PSD-010 mutation (b) adds an unmapped Bash deny, and S7 covers it. |
| D11 | Scenarios untraced | RESOLVED | S1–S7 each name their REQ (acceptance L28–L58). |
| D12 | Alias premise unverified | RESOLVED (residue → N4) | The `rm/del/rd…` canonicalization claim is removed, and REQ-PSD-010 excludes alias heads. |
| D13 | Grep AC lacks a positive control | RESOLVED | AC-PSD-007 L17 records the control. Re-measured: `git show 3c65a9f01:TMPL \| grep -c 'Bash(.*\\\\:'` → `5`; the PowerShell pattern on HEAD → `0`. The `\:` mutation is AC-PSD-010 (c). |
| D14 | macOS→Windows generalization | RESOLVED | spec §C L101; acceptance DoD L69. |
| D15 | Hook-matcher hazard uncited | RESOLVED | spec §D L107 cites tools-reference. Verified: tools-reference.md L397 "Match `Bash\|PowerShell` in hooks … matching `Bash` alone is not enough"; L413 "The Bash tool remains available for POSIX scripts when Git Bash is installed". |

## Verification of new claims introduced by the revision

| Claim | Command | Observed | Result |
|---|---|---|---|
| 47 Bash denies = 9 fs + 14 git + 3 disk + 11 system + 10 db (38 residual) | extract of the `develop` (`e4e4d2624`) TMPL deny array | 55 deny / 47 `Bash(` / 0 `PowerShell(`. Rows 1–9 fs; git: 7 plain + 7 `git *` = 14; disk: `Clear-Disk`, `Format-Volume`, `format` = 3; system: `chmod -R 777`, `chmod 777`, `dd`, `mkfs`, `fdisk`, `reboot`, `shutdown`, `init`, `systemctl`, `kill -9`, `killall` = 11; db: 10 | TRUE (spec L47–L54) |
| t1207 in develop | `git merge-base --is-ancestor baa054586 develop` | exit 0 | TRUE |
| `--tools`, `--setting-sources`, `--safe-mode`, `--strict-mcp-config` exist in 2.1.283 | `claude --version`; `claude --help` | 2.1.283. `--tools <tools...>` "list of available tools from the built-in set". `--setting-sources` "(user, project, local)". `--safe-mode` "… hooks, MCP servers … disabled … Admin-managed (policy) settings still apply. Auth, model selection, built-in tools and plugins, and permissions work normally". `--strict-mcp-config` present | TRUE; the plan's quotes are faithful |
| Denials surface in stream-json | headless.md L300 | "`permission_denied` system messages … `permission_denials`" | TRUE |
| PowerShell rule shape and case-insensitivity | permissions.md L302, L318 | "`:*` suffix is equivalent to a trailing ` *`"; "Matching is case-insensitive" | TRUE |
| TRUNCATE case-fold over-block | same + deny list | `Bash(TRUNCATE:*)` is the only uppercase SQL head whose lowercase form is a common utility (`truncate`). The other uppercase heads (`DROP …`, `DELETE FROM`, `redis-cli FLUSH*`) have no benign lowercase utility | TRUE; the note is accurate and the plan does not let it pass silently |
| `kill` is the only alias head in the residual 38 | inspection of heads against PowerShell aliases | `kill` → `Stop-Process` (Windows). No other residual head (`format`, `init`, `dd`, `chmod`, `mongo`, …) is a PowerShell alias | TRUE (inspection; not machine-checked against `Get-Alias`) |

## Hypothesis separation (V1–V6)

The design separates the hypotheses. A and C differ only in the `Bash(...)` rule, so a denial seen in A but not in C is attributable to that rule. B shows that the scratch settings loaded and that the PowerShell deny path fires. C shows that the command runs through PowerShell with nothing else blocking it. V2 excludes the Bash route. `--setting-sources project` together with `--safe-mode` excludes user-scope rules and hooks, and managed settings are recorded. If `--safe-mode` also suppressed project permission rules, both A and B would show no denial, V4 would fail, and the result would be INCONCLUSIVE. That is a safe failure, never a false Branch P. Two execution details can still spoil the single-shot M1 without producing a wrong branch; they are N1 and N5.

## Category Scores (rubric-anchored)
| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 | Minor ambiguity: REQ-PSD-005 L72 "a cap is reached" vs plan V6 L62 "exceeded 3 turns" (N1); per-arm fixture reset unstated (N5); glued trailing `*` not in the matcher model (N3) |
| Completeness | 1.00 | 1.0 | HISTORY L19–L24, background §A, requirements §B, constraints §C, 5 Out-of-Scope H3s with bullets L105–L123, 12 frontmatter fields |
| Testability | 0.85 | 0.75–1.0 | Every AC is binary, with positive controls on AC-007/008/010/012. REQ-PSD-004's "stop without running any arm" is checked only through AC-004's V1 row |
| Traceability | 0.95 | 1.0 (−0.05 indirect) | All 18 REQs map (acceptance L11–L24); all S1–S7 carry REQ refs; REQ-PSD-004 is reached indirectly (AC-003 records the precondition, AC-004 the stop) |

## Defects Found (structured defect-list)

N1. CAP-REACHED-VS-EXCEEDED — spec.md:L72 (REQ-PSD-005) vs plan.md:L62 (V6), L24 — REQ-PSD-005 makes "a cap is reached" INCONCLUSIVE, while V6 says "exceeded 3 turns". `--max-turns 3` makes exceeding impossible, so V6 as written is vacuous. The realistic case is arm B or arm A: a denied model retries and ends on the turn cap (max-turns error). REQ-PSD-005 calls that INCONCLUSIVE; V6 calls it valid. Because M1 has no retries, this contradiction decides whether a correctly denied arm counts. — Severity: minor — Class: blocking — Required fix: make one rule govern. Recommended: V6 = "`timeout` did not fire (exit ≠ 124)". Treat hitting the turn cap as allowed when the recorded PowerShell `tool_use` and its denial or execution are present, and say so in REQ-PSD-005 ("the wall-clock cap is reached"). Otherwise, state explicitly that hitting max-turns is INCONCLUSIVE and prompt for a single tool call.

N5. PER-ARM-FIXTURE — plan.md:L28, L41–L48 — The plan describes one scratch project with one committed `.claude/settings.json`, yet each arm needs a different `permissions.deny`. Arm A (if the gap exists) or arm C deletes `victim.txt`, and `git clean -fdx` also deletes the untracked `victim-dir/` used by arm D. Run in sequence in one project, a later arm starts with its observable already gone, so V4/V5 fail and M1 ends INCONCLUSIVE with no retry allowed. — Severity: minor — Class: blocking — Required fix: state "a fresh scratch project per arm, created from the same recipe with that arm's settings". Otherwise, specify the exact reset between arms and record each arm's pre-run existence of the observable in §E.2.

N2. TRUNCATE-DECISION-OWNER — plan.md:L96; Open Decisions L139–L148 — "recording operator acceptance" of the `TRUNCATE` over-block is an operator decision that is missing from D1–D4. The D1 counts (37 / 27) also assume `TRUNCATE` ships. — Severity: minor — Class: optional — Required fix: fold it into D1 as a sub-choice ("ship `TRUNCATE` counterpart and accept the `truncate` utility over-block / exclude it with reason `case-fold over-block`") so the answer is collected at Kickoff.

N3. GLUED-TRAILING-STAR — plan.md:L91 — The matcher model defines `:*`, a trailing ` *`, and a non-trailing `*`, but not a trailing `*` glued to text (`git * reset --hard*`, `--force*`). The known-deny control `git -C repo reset --hard HEAD` (L94) depends on it. The test self-corrects, because the control goes red if the case is unmodeled. — Severity: minor — Class: optional — Required fix: add "a trailing `*` not preceded by a space matches any character sequence (including empty)".

N4. BUILTIN-ALIAS-INFERENCE — spec.md:L43 — "enforced … for `Remove-Item` (and its canonicalized aliases)": the built-in section (permission-modes.md L664–L676) names `Remove-Item` and the `cmd`-run `rd`/`rmdir`/`del`/`erase` only. Alias canonicalization is documented for permission rules (permissions.md L318), not for the built-in check. — Severity: minor — Class: optional — Required fix: mark "(and its canonicalized aliases)" as inferred, or add it to the §B Gap list next to native `rm`.

## Recommendation

PASS. Rationale per must-pass: MP-1 contiguous 001–018; MP-2 all REQs are GEARS forms (requirement layer); MP-3 lint clean and spec_audit shows no drift; MP-4 N/A-equivalent (OS-shell axis); MP-5 the reference is `completed`; MP-6 no `syscall`; MP-7 no markers. D1–D10 are all RESOLVED with line evidence, and D11–D15 are resolved as well.

Conditions before M1 executes. These are blocking-class and do not need a third audit; the lead confirms them mechanically:
1. N1: `grep -n 'V6' plan.md` and `grep -n 'cap is reached' spec.md` must state one consistent rule for the turn cap.
2. N5: plan §C.1 must state a fresh scratch project per arm (or an explicit reset with recorded pre-run existence).
N2–N4 are optional. N2 is best folded into the D1 Kickoff question.

## Operator decisions (Kickoff)

| Decision | Owner | Content |
|---|---|---|
| D1 parity scope | Operator (Kickoff) | Takes effect only under Branch P. (a) residual 38 minus `kill -9` = 37 counterparts, recommended as the default and covering every Bash deny the built-in does not; (b) git 14 + disk 3 + db 10 = 27, dropping the `dd`/`mkfs`/`chmod`/`shutdown` counterparts for opt-in `pwsh` on macOS/Linux; (c) git 14 only; (d) (a) plus the 9 filesystem rows, which adds nothing for the documented cases and affects only the unmeasured native-`rm`-from-`pwsh` case. Sub-choice N2: ship the `TRUNCATE` counterpart and accept the `truncate` utility over-block, or exclude it. |
| D4 accept Branch N as a valid close | Operator (Kickoff) | If M1 shows that Bash denies already reach PowerShell, the card lands no rule change and no Go guard. The deliverable is the §E.2 evidence plus the 4-locale docs note. |
| D3 hook matcher card | Operator issues; lead proposes | `Write\|Edit\|Bash` → `Write\|Edit\|Bash\|PowerShell` as a separate card, citing tools-reference L397. |
| D2 `Remove-Item` pattern shape | Lead / run-phase | Relevant only under D1(d); space-suffix syntax only. Not an operator decision. |
