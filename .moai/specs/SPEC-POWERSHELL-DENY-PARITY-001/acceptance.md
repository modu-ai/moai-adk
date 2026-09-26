# Acceptance — SPEC-POWERSHELL-DENY-PARITY-001

## §D AC Matrix

| AC | REQ | Verification |
|---|---|---|
| AC-PSD-001 | REQ-PSD-001, REQ-PSD-002 | progress §E.2 carries per-arm rows (A, B, control C): command, exit code, `victim/` existence, invoked tool name, deny message — all from one capped run each |
| AC-PSD-002 | REQ-PSD-003 | progress §E.2 records the search for a model-free rule-evaluation command and its result (found + used, or not found with the source consulted) |
| AC-PSD-003 | REQ-PSD-004, REQ-PSD-005 | the branch taken after M1 is named in §E.2 and matches the observed arm outcomes |
| AC-PSD-004 | REQ-PSD-006, REQ-PSD-007 | every in-scope row of plan §C.2 has a `tool: "PowerShell"` deny entry in tool-policy.yaml and a `PowerShell(...)` string in the template deny list |
| AC-PSD-005 | REQ-PSD-008 | `grep -c 'PowerShell(.*\\\\:' internal/template/templates/.claude/settings.json.tmpl` → 0 |
| AC-PSD-006 | REQ-PSD-009 | Go test asserts no PowerShell deny rule matches any benign-sample string of plan §C.3 |
| AC-PSD-007 | REQ-PSD-010 | `make tool-policy-drift-check` exit 0; `make build` exit 0 |
| AC-PSD-008 | REQ-PSD-011 | new parity test fails when one counterpart is removed from the template (mutation observed), passes on the final tree |
| AC-PSD-009 | REQ-PSD-012 | template neutrality + internal-content-leak tests pass |
| AC-PSD-010 | REQ-PSD-013 | the four `advanced/settings-json.md` locale files each mention PowerShell deny coverage in the same commit |
| AC-PSD-011 | REQ-PSD-014, REQ-PSD-015 | `git diff develop -- settings.json.tmpl` shows no removed/changed `Bash(` deny line and no `PowerShell(` string in allow/ask |

## §D.1 Scenarios

**AC-PSD-S1 — gap confirmed**
Given a scratch project whose only deny rule is `Bash(Remove-Item *)`
When a capped PowerShell-tool session runs `Remove-Item -Recurse -Force victim`
Then `victim/` is deleted, and with `PowerShell(Remove-Item *)` instead the directory survives and a deny event is recorded.

**AC-PSD-S2 — gap absent**
Given arm A leaves `victim/` in place with a deny event
When M1 closes
Then no `PowerShell(...)` rule is added, and only the regression guard and the docs note land.

**AC-PSD-S3 — alias coverage**
Given the final template deny list
When the command `rd /s /q C:/` or `rm -r -fo C:/` is evaluated against the PowerShell rules (mechanically, or recorded as a Gap if only model arms exist)
Then it is denied through alias canonicalization.

**AC-PSD-S4 — no over-block**
Given the final template deny list
When `Remove-Item ./build -Recurse -Force` or `git push origin HEAD` is evaluated
Then it is not denied.

**AC-PSD-S5 — parity guard bites**
Given the parity test on the final tree
When one PowerShell counterpart is deleted from the template
Then the test fails naming the missing Bash rule.

## §D.2 Definition of Done

- M1 evidence recorded before any rule change; branch decision traceable.
- All ACs PASS with verbatim command output in progress §E.2.
- Scoped tests green locally (`internal/template`, `internal/config/toolpolicy`); full suite via CI on develop push.
- Open decisions D1–D4 resolved and recorded.
