# MCP Recipe Catalogue

> Local-only (`.moai/docs/` convention). Covers opt-in recipes + skip rationale for the §3.7 trend-MCP tool set. Every opt-in recipe carries BOTH a copy-pasteable `.mcp.json` snippet AND a one-line `moai mcp add ...` equivalent — the two are byte-equivalent after normalization. Every opt-in recipe carries the "supply, do not redefine" note: these tools SUPPLY evidence to `/moai gate` LSP, `verification-claim-integrity` attributed-baseline, and `sync-auditor` 4-dim scoring; they do NOT replace those gates.

## Active default-on entries (shipped in the template `.mcp.json`)

The distributed `.mcp.json` carries exactly three active entries by default. They are secret-free npx-launched stdio servers and pass the §25 template-neutrality audit. No recipe action is required — they activate automatically when the user's Claude Code session loads the project.

## ast-grep (default-disabled)

`ast-grep` is a structural-search MCP server. It is default-disabled because the `uvx` (Python ≥ 3.11) runtime is not universally available; the user opts in by adding the entry. The template ships it in this catalogue ONLY — it is NOT in the active `mcpServers` map.

`.mcp.json` snippet:

```json
{
  "mcpServers": {
    "ast-grep": {
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec uvx ast-grep-mcp@latest"]
    }
  }
}
```

One-line CLI equivalent:

```
moai mcp add ast-grep --command /bin/bash --args -l --args -c --args 'exec uvx ast-grep-mcp@latest' --scope project
```

Supply, do not redefine: ast-grep's structural-search output SUPPLIES evidence to `/moai gate` LSP findings; the LSP gate (not ast-grep) remains the SSOT gate.

## Semgrep (opt-in)

`.mcp.json` snippet:

```json
{
  "mcpServers": {
    "semgrep": {
      "type": "http",
      "url": "https://semgrep.example.com/mcp",
      "headers": {
        "Authorization": "Bearer ${SEMGREP_API_TOKEN}"
      }
    }
  }
}
```

One-line CLI equivalent:

```
moai mcp add semgrep --type http --url https://semgrep.example.com/mcp --headers 'Authorization=Bearer ${SEMGREP_API_TOKEN}' --scope project
```

Supply, do not redefine: Semgrep findings SUPPLY evidence to `/moai gate` LSP and the Secured dimension of `sync-auditor` 4-dim scoring. The gate semantics, PASS/FAIL contract, and thresholds remain owned by `/moai gate` + `sync-auditor` (the SSOTs); Semgrep is a SUPPLY surface, not a gate replacement.

## GitHub MCP (opt-in)

`.mcp.json` snippet:

```json
{
  "mcpServers": {
    "github": {
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec npx -y @modelcontextprotocol/server-github@latest"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

One-line CLI equivalent:

```
moai mcp add github --command /bin/bash --args -l --args -c --args 'exec npx -y @modelcontextprotocol/server-github@latest' --env 'GITHUB_TOKEN=${GITHUB_TOKEN}' --scope project
```

Supply, do not redefine: GitHub MCP SUPPLIES repository / PR / issue evidence. The Trackable dimension of `sync-auditor` (Conventional Commits, PR trailers) remains owned by `sync-auditor`; GitHub MCP does not redefine it.

## Postgres / Neon (opt-in)

`.mcp.json` snippet:

```json
{
  "mcpServers": {
    "postgres": {
      "command": "/bin/bash",
      "args": ["-l", "-c", "exec npx -y @modelcontextprotocol/server-postgres@latest '${DATABASE_URL}'"]
    }
  }
}
```

One-line CLI equivalent:

```
moai mcp add postgres --command /bin/bash --args -l --args -c --args 'exec npx -y @modelcontextprotocol/server-postgres@latest ${DATABASE_URL}' --scope project
```

Note: `${DATABASE_URL}` expansion here happens inside the bash `-c` shell, not the Claude Code runtime env-var path. The user MUST set `DATABASE_URL` in their shell environment before launching Claude Code; the literal `${DATABASE_URL}` is preserved in `.mcp.json` (no resolved secrets).

Supply, do not redefine: Postgres query results SUPPLY evidence to verification claims; the attributed-baseline invariant (`verification-claim-integrity`) remains the SSOT for what counts as observed.

## Sentry (opt-in)

`.mcp.json` snippet:

```json
{
  "mcpServers": {
    "sentry": {
      "type": "http",
      "url": "https://sentry.example.com/api/0/mcp/",
      "headers": {
        "Authorization": "Bearer ${SENTRY_AUTH_TOKEN}"
      }
    }
  }
}
```

One-line CLI equivalent:

```
moai mcp add sentry --type http --url https://sentry.example.com/api/0/mcp/ --headers 'Authorization=Bearer ${SENTRY_AUTH_TOKEN}' --scope project
```

Supply, do not redefine: Sentry error evidence SUPPLIES the Secured dimension of `sync-auditor`; `sync-auditor` (not Sentry) remains the SSOT gate.

## Codecov (opt-in)

`.mcp.json` snippet:

```json
{
  "mcpServers": {
    "codecov": {
      "type": "http",
      "url": "https://codecov.example.com/mcp/",
      "headers": {
        "Authorization": "Bearer ${CODECOV_API_TOKEN}"
      }
    }
  }
}
```

One-line CLI equivalent:

```
moai mcp add codecov --type http --url https://codecov.example.com/mcp/ --headers 'Authorization=Bearer ${CODECOV_API_TOKEN}' --scope project
```

Supply, do not redefine: Codecov coverage data SUPPLIES the Tested dimension of `sync-auditor`; the coverage threshold (`/moai gate` LSP coverage check + `sync-auditor` Tested) remains the SSOT gate.

## Skip — Sequential Thinking

Sequential Thinking MCP is skipped. Rationale: the `ultrathink` keyword (Adaptive Thinking on Opus 4.7+, including Opus 5 and 4.8) supersedes Sequential Thinking as the deep-reasoning path. Adding a parallel MCP-backed reasoning surface would duplicate the `ultrathink` keyword's role and create two sources of truth for "deep reasoning".

## Skip — Filesystem

Filesystem MCP is skipped. Rationale: Claude Code's native Read / Write / Edit / Glob / Grep tools already cover filesystem access with a permission layer. Adding the Filesystem MCP would duplicate the native tools without adding capability, and would bypass the native permission layer (a security regression).

## Skip — Git

Git MCP is skipped. Rationale: the main-checkout branch guard (`.claude/rules/moai/workflow/main-checkout-branch-guard.md`) constrains `git switch` / `git reset --hard` / `git stash` / `git rebase` in the primary checkout for concurrency safety. A Git MCP server would operate outside that guard discipline and bypass the safety net; the Bash tool + the orchestrator-direct git operations are the SSOT surface.

## Skip — Memory / Knowledge Graph

Memory / KG MCP is skipped. Rationale: MoAI's auto-memory subsystem (`~/.claude/projects/<hash>/memory/` indexed by `MEMORY.md`) plus the per-agent `.claude/agent-memory/<agent-name>/` taxonomy already cover the persistent-memory surface. Adding a Memory MCP would duplicate the auto-memory store with a parallel namespace.

## Skip — Brave / Exa Search

Brave Search / Exa MCP is skipped. Rationale: Claude Code's native `WebSearch` + `WebFetch` (and under GLM backend, the z.ai `mcp__web_search_prime__webSearchPrime` / `mcp__web_reader__webReader`) already cover web search + URL verification. The MoAI anti-hallucination policy (`moai-constitution.md` § URL Verification) binds on the native surfaces; adding a Brave/Exa MCP would create a parallel search path outside the URL-verification discipline.

---

Cross-references:

- SPEC-TREND-MCP-001 (the run-phase SPEC that authored this catalogue).
- `.claude/rules/moai/core/settings-management.md` § MCP Configuration (the doctrine surface describing the multi-entry provisioning contract).
- `internal/template/templates/.mcp.json` (the template source — exactly 3 active entries).
- `internal/cli/mcp.go` (the generic `moai mcp add|remove|list` CLI).
