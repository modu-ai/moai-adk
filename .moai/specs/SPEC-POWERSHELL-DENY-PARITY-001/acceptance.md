# Acceptance — SPEC-POWERSHELL-DENY-PARITY-001

Branch column: **both** = binds Branch P and Branch N; **P** = binds only when M1 confirms the gap; **N** = binds only when M1 shows Bash denies already reach PowerShell. ACs of the branch not taken are recorded as `N/A (branch <X>)`, not as PASS.

Paths are repository-relative from the repository root. `TMPL` = `internal/template/templates/.claude/settings.json.tmpl`. `BASE` = the local `develop` SHA absorbed in M0 (recorded in §E.2).

## §D AC Matrix

| AC | Branch | REQ | Verification |
|---|---|---|---|
| AC-PSD-001 | both | REQ-PSD-001, REQ-PSD-002 | progress §E.2 carries one row per arm (A, B, C, D) with: full command line, exit code, `system/init` tools list, PowerShell `tool_use` command, `permission_denied` / `permission_denials` content, target existence after the run, turn count, timeout status — one capped run each |
| AC-PSD-002 | both | REQ-PSD-003 | §E.2 records the search for a model-free rule-evaluation command and its result (found and used, or not found with the sources consulted) |
| AC-PSD-003 | both | REQ-PSD-004 | §E.2 records `pwsh` path and major version, `claude --version`, and managed-settings presence, captured before arm A |
| AC-PSD-004 | both | REQ-PSD-005, REQ-PSD-006, REQ-PSD-007 | §E.2 marks V1–V6 (plan §C.1) each with the evidence line it rests on, and names the branch; the named branch equals the outcome-table row selected by arm A. If any V fails, §E.2 names the failed condition and the run-phase stopped with a blocker report (no later milestone commit exists) |
| AC-PSD-005 | both | REQ-PSD-008 | §E.2 records arm D (denial present or absent, `victim-dir/keep.txt` existence) and labels it as an observation that selected no branch |
| AC-PSD-006 | P | REQ-PSD-009, REQ-PSD-010 | a script extracting the `TMPL` deny array shows: every Bash deny is either mapped with its `PowerShell(...)` counterpart present or listed on the exclusion list with a reason; the count of `PowerShell(` denies equals the D1 scope size minus exclusions; tool-policy.yaml carries the same entries with `tool: "PowerShell"` |
| AC-PSD-007 | P | REQ-PSD-011 | `grep -c 'PowerShell(.*\\\\:' internal/template/templates/.claude/settings.json.tmpl` → `0`. Positive control (measured at plan time): the same pattern with the `Bash(` prefix on `3c65a9f01` (pre-t1207) → `5`, so the pattern detects the escape |
| AC-PSD-008 | P | REQ-PSD-012 | the over-block Go test passes on the final tree: no benign sample of plan §C.3 matches a `PowerShell(...)` deny and every known-deny control matches. Positive control: replacing the matcher with an always-false function makes the known-deny assertions fail (observed once, output in §E.2). Alias- or compound-dependent samples are listed as a Gap |
| AC-PSD-009 | P | REQ-PSD-013 | `make tool-policy-drift-check` exit 0; `make build` exit 0 |
| AC-PSD-010 | P | REQ-PSD-014 | the closed-world guard passes on the final tree and fails, with the offending rule named, under each of four observed mutations of a scratch copy of the deny list: (a) one counterpart removed; (b) a new `Bash(...)` deny added with neither counterpart nor exclusion; (c) `\:` injected into a `PowerShell(...)` rule; (d) a `PowerShell(...)` rule combining a middle `*` with `:*` |
| AC-PSD-011 | both | REQ-PSD-015 | template-neutrality and internal-content-leak tests pass (`go test ./internal/template/...`) |
| AC-PSD-012 | both | REQ-PSD-017, REQ-PSD-018 | a set comparison of `git show BASE:TMPL` against the final `TMPL`: the `Bash(` deny set is identical, `PowerShell(` occurs 0 times in `allow` and `ask`, `defaultMode` is unchanged. Positive control: the same script run against a copy with one `Bash(` deny removed reports that difference |
| AC-PSD-013 | both | REQ-PSD-016 | each of `docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md` contains the branch's note, and all four files change in the same commit (`git show --stat <commit>` lists all four) |
| AC-PSD-014 | N | REQ-PSD-007 | `git diff --quiet BASE -- internal/template/templates/.claude/settings.json.tmpl .moai/config/sections/tool-policy.yaml` exits 0, and no test file asserting PowerShell rule parity is added |

## §D.1 Scenarios

**AC-PSD-S1 — gap confirmed (REQ-PSD-001, REQ-PSD-006)**
Given a scratch project whose only deny rule is `Bash(git clean -fdx:*)`, with the PowerShell tool as the only built-in tool and V1–V6 holding
When a capped session runs `git clean -fdx` through the PowerShell tool
Then no denial is recorded for that PowerShell call, `victim.txt` is deleted, and §E.2 names Branch P.

**AC-PSD-S2 — gap absent (REQ-PSD-007)**
Given V1–V6 hold and arm A records a denial on the PowerShell call with `victim.txt` surviving
When M1 closes
Then no `PowerShell(...)` rule and no parity test are added, and only the §E.2 evidence and the 4-locale docs note land.

**AC-PSD-S3 — inconclusive (REQ-PSD-005)**
Given arm B's deny rule `PowerShell(git clean -fdx:*)`
When arm B ends with `victim.txt` deleted, or the `system/init` tools list contains `Bash`
Then the run-phase names the failed validity condition in §E.2, returns a blocker report, and selects neither branch.

**AC-PSD-S4 — built-in observation is not attributed (REQ-PSD-008)**
Given arm D runs `Remove-Item -Recurse -Force ./victim-dir/*` with no custom deny rule
When the run ends
Then §E.2 records the observed outcome as evidence about Claude Code's built-in check, and no REQ of this SPEC is marked satisfied by it.

**AC-PSD-S5 — residual deny covered (REQ-PSD-009, REQ-PSD-014)** — Branch P
Given the final template deny list and the matcher model of plan §C.3
When `git push --force origin main` is evaluated against the `PowerShell(...)` denies
Then it matches a counterpart rule (no built-in check covers git, so the match is attributable to this SPEC's rules).

**AC-PSD-S6 — no over-block (REQ-PSD-012)** — Branch P
Given the final template deny list
When `Remove-Item ./build -Recurse -Force`, `git push origin HEAD`, or `truncate -s 0 app.log` is evaluated against the `PowerShell(...)` denies
Then none matches (the `truncate` case is resolved by exclusion or recorded operator acceptance, per plan §C.3).

**AC-PSD-S7 — guard is closed-world (REQ-PSD-014)** — Branch P
Given the closed-world guard on the final tree
When a new `Bash(...)` deny is added to a scratch copy with neither a counterpart nor an exclusion entry
Then the test fails naming that Bash rule.

## §D.2 Definition of Done

- M1 preconditions, arms, validity marks, and branch recorded before any rule or docs change.
- Every AC of the taken branch PASS with verbatim command output in progress §E.2; every AC of the other branch recorded `N/A (branch <X>)`.
- Scoped tests green locally (`internal/template`, `internal/config/toolpolicy`); full suite via CI on the develop push.
- Open decisions D1–D4 resolved and recorded.
- Residual risk recorded in §E.2: the measurement ran on macOS with the opt-in tool, and its transfer to Windows is inferred, not observed.
