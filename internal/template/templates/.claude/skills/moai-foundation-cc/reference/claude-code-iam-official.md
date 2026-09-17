# Claude Code IAM & Permissions - Reference

> **Illustrative reference, NOT the official Claude Code specification.**
> This file is an MoAI-authored summary intended for orientation only. The
> authoritative, up-to-date IAM and permissions surface lives in the official
> Claude Code documentation at
> https://code.claude.com/docs/en/iam — consult that source before relying
> on any field for a production deployment. Earlier revisions of this file
> presented fabricated RBAC roles (`developer` / `securityReviewer` /
> `devopsEngineer` with `toolRestrictions`), an `enterprise.policies`
> framework, and SOC2 / ISO27001 compliance JSON as if they were real
> Claude Code IAM surfaces; they do not exist in the product and have been
> removed. What remains is the real, currently-shipped permission surface.

Source: https://code.claude.com/docs/en/iam

## Key Concepts

### What is Claude Code IAM?

Claude Code's Identity and Access Management surface is a **permission
mode + allow/ask/deny list** model, enforced at four settings scopes
(enterprise/managed → user → project → local). There is NO role-based
access control (RBAC), NO predefined `developer`/`reviewer`/`devops` roles,
and NO `toolRestrictions` sub-object — the earlier sections that described
those were illustrative-only and have been removed.

## Permission Modes

The `permissions.defaultMode` field accepts six values (official docs,
re-verified 2026-09-13):

| Mode | Behavior |
|------|----------|
| `default` | Prompts per tool on first use; CLI label "Manual"; alias `manual` (v2.1.200+) |
| `acceptEdits` | Auto-accepts file edits + common filesystem commands in the working directory |
| `plan` | Read-only; Claude cannot modify files or run non-read tools |
| `auto` | Auto-approves tool calls; a second model (the classifier) reviews actions instead of you |
| `dontAsk` | Auto-denies what would otherwise prompt; only allow-listed tools run |
| `bypassPermissions` | Skips prompts (except actions no mode auto-approves); gated by `disableBypassPermissionsMode` |

Kill switches (settable to `"disable"` in any settings file):
`permissions.disableBypassPermissionsMode` and
`permissions.disableAutoMode`.

Scope caveat (load-bearing for automation that writes settings): `"auto"`
and `"bypassPermissions"` set in `.claude/settings.json` or
`.claude/settings.local.json` (PROJECT scope) are silently ignored —
`auto` falls back to the built-in default and `bypassPermissions` starts
the session in Manual mode. All other values (`default`, `acceptEdits`,
`plan`, `dontAsk`) apply from any settings file; `auto` and
`bypassPermissions` take effect from `~/.claude/settings.json` (USER
scope), `--settings`, or managed settings. Version notes: the `manual`
alias requires CC v2.1.200+; the built-in auto-mode default requires
v2.1.228+ (macOS/Linux/WSL) or v2.1.233+ (Windows).

```json
{
  "permissions": {
    "defaultMode": "default"
  }
}
```

## Allow / Ask / Deny Lists

The `permissions` object carries three ordered lists of tool-path patterns.
Each entry is a tool name with an optional parenthesized argument pattern:

```json
{
  "permissions": {
    "allow": [
      "Read",
      "Glob",
      "Grep",
      "Bash(git status:*)",
      "Bash(git log:*)"
    ],
    "ask": [
      "Bash(rm:*)",
      "Bash(sudo:*)"
    ],
    "deny": [
      "Read(~/.ssh/**)",
      "Bash(rm -rf /:*)"
    ],
    "additionalDirectories": [
      "/path/to/extra/checkout"
    ]
  }
}
```

- `allow` — auto-approve matching tool calls (no prompt).
- `ask` — always prompt, even in `acceptEdits` / `bypassPermissions`-leaning flows.
- `deny` — block unconditionally; never callable.
- `additionalDirectories` — grant Read access to directories outside the project root.

### Pattern Grammar (illustrative)

- `Bash(git status:*)` — the `:` prefix-anchored command form for Bash.
- `Read(./docs/**)` — glob-restricted Read paths.
- `WebFetch(domain:example.com)` — domain-scoped WebFetch.

The authoritative pattern grammar (including the `~` home-prefix, the `//`
glob separator, and MCP-tool naming `mcp__server__tool`) lives at
https://code.claude.com/docs/en/iam.

## Settings Scopes (Enterprise / Managed → Local)

Claude Code resolves settings in a strict precedence order. Higher scopes
override lower ones, and `deny` entries at any scope are absolute:

1. **Enterprise / Managed settings** — `/etc/claude/settings.json` (Linux/macOS) or `/Library/Application Support/ClaudeCode/managed-settings.json` (macOS Managed). Organization-wide; users cannot override.
2. **User settings** — `~/.claude/settings.json`. Personal defaults across all projects.
3. **Project settings** — `.claude/settings.json`. Team-shared, checked into VCS.
4. **Local settings** — `.claude/settings.local.json`. Per-developer overrides; gitignored.

There is NO inline `enterprise: { policies: { ... } }` settings.json object
of the kind earlier revisions showed. Enterprise policy is expressed by
editing the managed-settings file at one of the paths above, not by an
in-document `enterprise` block.

## Enterprise Levers

The real enterprise-grade settings keys (set in the managed-settings file,
not in an inline `enterprise` block):

- `disableBypassPermissionsMode: true` — prevents agents from entering `bypassPermissions` (Claude Code v2.1.111+).
- `permissions.deny` at managed scope — absolute deny that no lower scope can override.
- `permissions.allow` / `permissions.ask` at managed scope — org-wide defaults.

There is no `enterprise.compliance` block with SOC2 / ISO27001 fields.
Compliance posture for Claude Code itself is documented by Anthropic's
trust center, not configured in `settings.json`.

## What Was Removed (for readers familiar with older revisions)

For anyone who learned this surface from an earlier draft of this file,
the following were illustrative-only fabrications and do NOT exist in
Claude Code:

- The four "Levels" (Read-only / Bash / File-Modification / Administrative) as a tiered approval system — Claude Code uses the four permission modes above, not a 4-tier approval ladder.
- Predefined RBAC roles: `developer`, `securityReviewer`, `devopsEngineer`.
- `toolRestrictions` sub-objects (`allowedCommands`, `blockedCommands`, `allowedDomains`, `allowedPaths`, `maxFileSize`, etc.). The real model is the allow/ask/deny list, not a per-tool restrictions object.
- Custom role definition (`customRole` with `inherits`).
- `enterprise.policies.tools` / `enterprise.policies.mcpServers` / `enterprise.policies.roles` / `enterprise.policies.compliance`.
- `policyEnforcement` (validation / overrides / monitoring) blocks.
- `mcpSecurity` validation / sandbox / monitoring blocks (MCP permissions are governed by the same allow/ask/deny list, scoped with `mcp__server__tool`).
- `webPermissions` / `fileSystemPermissions` top-level objects (use `WebFetch(domain:...)` and `Read(path:...)` patterns in the standard permissions lists instead).
- Python `validate_tool_usage()` pseudocode and real-time monitoring JSON.
- SOC 2 / ISO 27001 compliance JSON blocks.

If you need a capability that the real surface above does not list, it is
almost certainly NOT available — do not assume a hidden field. Check
https://code.claude.com/docs/en/iam first.

## Best Practices

- **Principle of least privilege**: start with `default` mode and a minimal `allow` list; widen only when a workflow genuinely needs it.
- **Deny at managed scope for hard boundaries**: put `Read(~/.ssh/**)`, `Bash(rm -rf /:*)`, and similar in the managed-settings `permissions.deny` so no lower scope or errant agent can override them.
- **Local overrides stay local**: keep machine-specific entries in `.claude/settings.local.json`; never commit them.
- **`additionalDirectories` is a read grant**, not a write grant — it lets Claude Read outside the project root, it does not auto-approve Write/Edit there.
- **Audit the allow list**: periodically review `.claude/settings.json` `permissions.allow` for over-broad patterns (`Bash(*)` is a common footgun).

For the authoritative and current IAM surface — including fields added
after this file was last reviewed — consult the official documentation at
https://code.claude.com/docs/en/iam.
