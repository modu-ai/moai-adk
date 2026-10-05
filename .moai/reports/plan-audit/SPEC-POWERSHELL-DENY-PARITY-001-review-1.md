# SPEC Review Report: SPEC-POWERSHELL-DENY-PARITY-001
Iteration: 1/2 (Tier M ceiling = 2)
Verdict: FAIL
Overall Score: 0.66 (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. Audited tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1211`, HEAD `3c65a9f01`, branch `WT-powershell-deny`. Cross-model backends not invoked: no `audit_model` key in `.moai/config/sections/*.yaml` (grep empty).

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: REQ-PSD-001..015 contiguous, 3-digit padding, no duplicates (spec.md:L39-L62).
- [PASS] MP-2 GEARS compliance (requirement layer only): all 15 REQs are Ubiquitous / When / Where / shall-not forms, e.g. L41 "Where Claude Code offers a mechanical ... the measurement shall use it", L42 "When the measurement shows ... shall not add". ACs graded under Group 4, not here.
- [PASS] MP-3 frontmatter: 12 canonical fields present, `version: "0.1.0"` quoted, `status: draft`, `priority: P1`, `lifecycle: spec-anchored`, `tags` string (spec.md:L2-L13). `moai spec lint` → "No findings"; `spec_audit` (project_root = worktree) → 0 drift.
- [PASS] MP-4 language neutrality: the axis here is OS shell tools (Bash/PowerShell), not the 16 programming languages; the template already carries Windows `del`/`rmdir` forms. REQ-PSD-012 keeps the neutrality tests binding.
- [PASS] MP-5 D7: only reference SPEC-V3R6-TOOL-POLICY-SSOT-001, `status: completed` (not retired/superseded/archived).
- [PASS] MP-6 D8: `grep -c syscall` = 0 in all four files.
- [PASS] MP-7 clarification gate: `grep -rn 'NEEDS CLARIFICATION'` over the SPEC dir → no match (exit 1).

## Premise verification (focus item 2, 3)

| Claim | Command | Observed | Result |
|---|---|---|---|
| Template has 0 `PowerShell(` rules | python extract of `permissions.{allow,ask,deny}` from HEAD and `develop` template | HEAD deny 55 total / Bash 47 / PS 0; develop same; allow Bash 89 PS 0; ask 0 | TRUE (spec.md:L29) |
| 47 Bash denies after t1207 | same | 47 on develop | TRUE |
| t1207 in develop, not in HEAD | `git merge-base --is-ancestor baa054586 develop/HEAD` | develop=0, HEAD=1 | TRUE (spec.md:L33, plan L14) |
| t1207 landed form is literal `C:/` | develop tmpl + `git show develop:.moai/config/sections/tool-policy.yaml` L190 | `Bash(rm -rf C:/:*)`, `args_pattern: "Remove-Item -Recurse -Force C:/:*"`; HEAD still `C\\:/` (tool-policy.yaml L160-190, tmpl L553-557) | TRUE |
| `moai tool-policy build` + `--local-only` exist | `internal/cli/tool_policy.go` L35 `Use: "tool-policy"`, L72 `Use: "build"`, L172 `--local-only` | present | TRUE |
| Loader/codegen accept `tool: "PowerShell"` (SPEC left unchecked, §C L67) | `types.go` L86 `Tool string`, L119-124 `SettingsSpecifier` generic; `loader.go` L46 only checks non-empty; codegen emits `e.SettingsSpecifier()` for any tool | no tool-name allowlist anywhere | TRUE — the conditional generator change in §C is not needed |
| audit/source fields do not reach the template | no template mirror of `tool-policy.yaml` (`ls internal/template/templates/.moai/config/sections/ \| grep tool` empty); codegen writes only specifiers | — | plan §D concern resolved (no leak path) |
| `make tool-policy-drift-check` covers template | Makefile L66-68 | compares YAML vs **local** `.claude/settings.json` only | partial: template propagation is checked only by AC-PSD-004 string presence |
| Existing guard to extend | `git show develop:internal/template/settings_test.go` L55 `TestSettingsTemplateDenyWildcardSyntax` | present; checks `\:` and wildcard+`:*` mixing for `Bash(` prefix only | TRUE |
| Deny applies under bypassPermissions | `permission-modes.md` L30 (fetched 2026-09-26): "Deny rules block in every mode, including `bypassPermissions`." | — | TRUE — observable is sound on this axis |
| pwsh available for opt-in tool on macOS | `command -v pwsh` | `/usr/local/bin/pwsh`; `claude --version` = 2.1.283 | TRUE, but not declared as a precondition |

## Category Scores (rubric-anchored)
| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ-PSD-006/007 scope delegated to a "draft, finalised after M1" table (plan L37) and undecided D1; REQ-PSD-004 "regression guard asserting the measured behaviour" (spec L42) has no static realization under §C L68 |
| Completeness | 0.75 | 0.75 | All sections present; §A omits Claude Code's built-in PowerShell `Remove-Item` denies (defect D1), a material omission for the premise |
| Testability | 0.50 | 0.50 | AC-PSD-011 measured vacuous (D5); AC-PSD-006 has no positive control (D6); AC-PSD-S3 unattributable (D7); M1 decision rule incomplete (D2, D3) |
| Traceability | 0.75 | 0.75 | Matrix covers REQ-PSD-001..015 (acceptance L7-L17); scenarios S1-S5 cite no REQ; branch-conditional ACs contradict REQ-PSD-004 (D9) |

## Defects Found (structured defect-list)

D1. PREMISE-BUILTIN — spec.md:L25-L31, L48 (REQ-PSD-007); plan.md:L43-L46 — The background omits that Claude Code already denies `Remove-Item` on system paths in every mode. `permission-modes.md` § "Remove-Item in PowerShell" (L664-L676): "System paths: the filesystem root and its top-level directories, drive roots and their top-level directories, and your home directory. Claude Code denies the command in every mode, without asking you"; "Wildcards: a bare `*`, or any target ending in `/*` or `\*` ... denies the command in every mode"; the same system-path check covers `rd`/`rmdir`/`del`/`erase` via `cmd` (v2.1.283+). L543: these denies also apply in `bypassPermissions`. REQ-PSD-007's entire intent (drive root, Unix root, home, and their globs) is therefore already enforced natively, and D1's "filesystem-destructive subset only (card wording)" option is close to a no-op, while the real residual gap is the git/db/other rows. — Severity: major — Class: blocking — Required fix: add the built-in coverage to §A with the citation; restate REQ-PSD-007 as explicit defense-in-depth (or drop it); have M1 record the built-in behaviour; reframe D1 so the operator sees that the filesystem subset adds no enforcement beyond the built-in.

D2. M1-INCONCLUSIVE-INCOMPLETE — spec.md:L43 (REQ-PSD-005); plan.md:L35 — Inconclusive is defined only as "no deletion and no deny event" / "arm C does not delete". Not covered: (a) arm B deletes `victim/` — the positive control of the deny path failed (note `claude --help`: in `-p` mode "Settings files that fail validation are silently ignored", so a bad scratch settings file reads exactly like "rule did not block"); (b) the invoked tool in any arm is not `PowerShell`; (c) deletion happened via a tool other than PowerShell. Under the current text, outcome A-deletes / B-deletes / C-deletes has no defined branch and could be read as confirming the hypothesis. — Severity: major — Class: blocking — Required fix: add an explicit outcome table: proceed only when A deletes via PowerShell, B survives with a deny event on the PowerShell tool_use, C deletes via PowerShell; A-survives-with-deny and B-survives-with-deny and C-deletes → REQ-PSD-004; every other combination → REQ-PSD-005 blocker.

D3. M1-TOOL-ROUTE-CONFOUND — plan.md:L27-L33 — With `--max-turns 3` and the Bash tool still registered (on macOS the PowerShell tool is opt-in alongside Bash), a model denied in arm B can retry via `Bash rm -rf victim`, deleting the observable; arm A's outcome is likewise not attributable to the PowerShell path by directory state alone. — Severity: major — Class: blocking — Required fix: restrict the tool set per arm (e.g. `--tools PowerShell` or `--disallowedTools Bash,Write,Edit`) and decide each arm from the `stream-json` tool_use/tool_result events for the PowerShell call, with directory state as corroboration only.

D4. M1-ISOLATION-USER-SCOPE — spec.md:L40 (REQ-PSD-002); plan.md:L22-L24 — Isolation excludes only repository settings. User and managed scopes still load: `CLAUDE_CONFIG_DIR=/Users/goos/.moai/claude-profiles/moai-adk` in this environment, and `~/.claude/settings.json` carries `PreToolUse` and `PermissionRequest` hooks (measured). A user-scope PreToolUse hook returning deny/allow would contaminate every arm. pwsh 7+ presence is also undeclared. — Severity: major — Class: blocking — Required fix: pin sources with `--setting-sources project` (scratch settings only) and disable hooks (`--safe-mode`: "hooks ... disabled ... permissions work normally"), record the loaded sources per arm, and declare `pwsh` 7+ on PATH as a precondition.

D5. AC-VACUOUS-PATHSPEC — acceptance.md:L17 (AC-PSD-011) — `git diff develop -- settings.json.tmpl` matches only a repo-root file. Measured: `git diff --stat 4dcd4d8d4 develop -- settings.json.tmpl` → empty; `... -- internal/template/templates/.claude/settings.json.tmpl` → "1 file changed, 5 insertions(+), 5 deletions(-)". AC-PSD-011 passes on any change. — Severity: major — Class: blocking — Required fix: use the full path, and verify by set comparison of the rendered deny/allow/ask arrays (Bash deny set unchanged; no `PowerShell(` in allow/ask; `defaultMode` unchanged) rather than diff-line reading.

D6. AC-NO-POSITIVE-CONTROL — acceptance.md:L12 (AC-PSD-006); plan.md:L55 — The benign-sample Go test must reimplement Claude Code's matcher (wildcards, `:*`, case-insensitivity, alias canonicalization). A matcher that never matches passes every benign case; alias samples such as `rm ./tmp.txt` can only be evaluated if canonicalization is modeled. — Severity: major — Class: blocking — Required fix: require the same test to assert that known-deny strings (e.g. `Remove-Item -Recurse -Force C:/x`, `git push --force origin`) DO match; state which semantics the test models; mark alias-dependent benign samples as a Gap unless canonicalization is modeled.

D7. AC-UNATTRIBUTABLE — acceptance.md:L31-L34 (AC-PSD-S3) — `rd /s /q C:/` and `rm -r -fo C:/` target a drive root, which the built-in system-path deny blocks regardless of any rule (D1). A PASS cannot be attributed to the SPEC's rules. — Severity: major — Class: blocking — Required fix: drop S3 or restate it as "recorded as not attributable (built-in deny)"; alias coverage of the SPEC's rules can only be shown on a non-system target, which the SPEC's rules do not cover.

D8. D2-DEFECTIVE-SHAPE — plan.md:L51, L91 — D2's wider option `Remove-Item * C:/:*` mixes a middle `*` with the `:*` suffix. The develop guard `TestSettingsTemplateDenyWildcardSyntax` (L88-L90) rejects that shape for Bash as "mixes wildcard with legacy prefix syntax"; the same matcher treats the middle `*` as literal, so the rule would block nothing. — Severity: major — Class: blocking — Required fix: remove that option or express it in space-suffix syntax only; extend the mixing check and the `\:` check to the `PowerShell(` prefix in REQ-PSD-011.

D9. BRANCH-CONTRADICTION — spec.md:L42 vs L55, L57; acceptance.md:L10, L14, L16 — Under the REQ-PSD-004 branch (no PowerShell rules), REQ-PSD-011 ("fails when any in-scope Bash rule lacks its PowerShell counterpart") fails by construction, REQ-PSD-013 would state that denies "are declared for both the Bash and the PowerShell tools" (false), and AC-PSD-004/008/010 cannot pass. The "regression guard asserting the measured behaviour" in REQ-PSD-004 has no static form, since §C L68 forbids spawning Claude Code in `go test`. — Severity: major — Class: blocking — Required fix: condition REQ-PSD-011/013 and AC-PSD-004/005/006/008/010 on the REQ-PSD-006 branch; define the REQ-PSD-004 deliverable concretely (e.g. recorded evidence plus a docs note stating Bash rules already reach PowerShell) and drop the undefined runtime guard.

D10. GUARD-OPEN-WORLD — spec.md:L55 (REQ-PSD-011); acceptance.md:L14 — "In-scope" is defined by a hand-kept table, so the guard catches a deleted counterpart but not a Bash deny added later without a counterpart. That is the drift the guard exists to stop. — Severity: major — Class: blocking — Required fix: make the guard closed-world: every `Bash(` deny in the rendered template must either map to a declared `PowerShell(` counterpart or appear on an explicit exclusion list in the test; any unlisted Bash deny fails.

D11. SCENARIO-TRACE — acceptance.md:L21-L44 — Scenarios S1-S5 name no REQ-PSD-XXX. — Severity: minor — Class: optional — Required fix: add REQ references (S1→001/006, S2→004, S3→007, S4→009, S5→011).

D12. ALIAS-PREMISE — spec.md:L48 — REQ-PSD-007 asserts that `rm`, `del`, `rd`, `ri`, `rmdir`, `erase` canonicalize to `Remove-Item`. The docs say only that "Common aliases are canonicalized" (permissions.md L318) and do not list these aliases. — Severity: minor — Class: optional (may be moot after D1) — Required fix: mark as unverified or cite a source that lists them.

D13. GREP-CONTROL — acceptance.md:L11 (AC-PSD-005) — Returns 0 on today's tree trivially. The pattern itself works: its Bash-form positive control `grep -c 'Bash(.*\\\\:'` on HEAD returns 5. The mutation in AC-PSD-008 covers only a removed counterpart, not the `\:` branch. — Severity: minor — Class: optional — Required fix: record the positive control next to the AC and add a `\:` mutation to AC-PSD-008.

D14. PLATFORM-GENERALIZATION — plan.md:L27 — M1 runs on macOS (opt-in tool) while the at-risk users are on Windows. Carrying the result over to Windows is an inference. — Severity: minor — Class: optional — Required fix: record it under Residual-risk in §E.2.

D15. D3-CITATION — spec.md:L75 — Treating the matcher as out of scope is defensible, but the vendor doc states the risk directly (hooks.md, PowerShell input): "On Windows without Git Bash ... Claude Code doesn't register the Bash tool at all. A hook that matches only `Bash` never fires there." — Severity: minor — Class: optional — Required fix: cite it so the follow-up card's priority is evidence-based.

## Answers to the focus questions

1. **Can the measurement arms separate the hypotheses?** In principle, yes. A is the discriminating arm, B shows that a PowerShell deny actually blocks, and C shows that the PowerShell path runs at all. The observable is sound under `bypassPermissions`, since deny rules block in every mode (permission-modes.md L30). Caps are declared in the text before execution (REQ-PSD-002), and the outer `timeout 180` bounds each run from outside. As written, however, the result is not separable. Failure of arm B has no defined branch (D2). The Bash tool stays available as a fallback route (D3). User-scope settings and hooks are not isolated (D4). The inconclusive branch is a blocker report rather than a silent pass, but only for one failure shape (D2).
2. **Are the premise claims correct?** All verified as true (table above). The SPEC also omits a premise it needs: the built-in PowerShell `Remove-Item` denies (D1).
3. **Does the loader/codegen accept `PowerShell`?** Yes, with no code change needed; the tool name is a free string (types.go L86, loader.go L46, codegen via `SettingsSpecifier`).
4. **Are the ACs mechanically checkable, and are the guards non-vacuous?** AC-PSD-011 is vacuous (D5). The over-block test has no positive control (D6). S3 cannot be attributed to the SPEC's rules (D7). The guard catches a removed counterpart but not a new Bash rule that lacks one (D10). The `\:` check works: its positive control returned 5 (D13).
5. **Which decisions need the operator?** See below.
6. **Is the scope boundary sound?** Keeping the hook matcher (D3) out of scope is sound as a boundary. Cite the vendor text (D15).

## Decisions: operator vs lead

| Decision | Owner | Why |
|---|---|---|
| D1 parity scope | **Operator (at Kickoff)** | Changes the deny set shipped to every user project. After D1 above, "filesystem subset" adds no enforcement beyond the built-in, so the operator needs the corrected framing before choosing. |
| D4 accept a no-op close | **Operator (at Kickoff)** | Pre-accepts an outcome in which the card lands no rule change; this changes what the card delivers. |
| D3 hook matcher as a separate card | **Operator** (card issuance); lead proposes | Admitting a card to the queue is the operator's act. The lead may recommend it, citing hooks.md. |
| D2 pattern shape | **Lead / run-phase**, from M1 evidence within REQ-PSD-009 | A technical choice that the evidence settles. The defective `* ...:*` option must be removed first (D8). |

## Recommendation

FAIL. All must-pass rows pass, but the score of 0.66 is below the Tier M threshold of 0.80, and ten blocking majors remain. Fix in this order:

1. Rewrite §A and REQ-PSD-007 with the built-in `Remove-Item` coverage, and reframe D1 (D1).
2. Replace REQ-PSD-005 and plan §C.1 with a full outcome table; restrict the tool set; isolate the setting sources and hooks; declare pwsh (D2, D3, D4).
3. Condition the post-M1 REQs and ACs on the branch taken, and define the REQ-PSD-004 deliverable (D9).
4. Fix AC-PSD-011's pathspec; give AC-PSD-006 a positive control; drop or restate S3; make the guard closed-world; remove the mixed-wildcard D2 option (D5, D6, D7, D10, D8).
5. Handle the optional items D11 to D15 at the author's discretion.

The re-audit (iteration 2/2) is scoped to D1-D10.
