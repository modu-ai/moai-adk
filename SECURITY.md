# Security Policy

## Reporting a vulnerability

**Please do not report security vulnerabilities in public Issues, pull requests or Discussions.**

Report them privately through GitHub's private vulnerability reporting:

1. Go to the repository's [Security advisories](https://github.com/modu-ai/moai-adk/security/advisories) page.
2. Click **Report a vulnerability** (direct link: [new report](https://github.com/modu-ai/moai-adk/security/advisories/new)).
3. Describe the problem, the affected version or commit, the steps to reproduce it, and the impact you expect.

Only the maintainers and you can see the report. We will work with you on a fix and on the timing of any public disclosure.

**Response time:** [PENDING MAINTAINER DECISION: the target for the first response and for an assessment has not been set yet. This section will state it once decided.]

## Versions

| Generation | Branch | State |
| --- | --- | --- |
| Rust release candidate `moai-cli 2026.11.0-rc.1` | `v2026.11.0` | prerelease, not yet released |
| Go 3.x (`moai`) | `main` | latest release `v3.1.2` |

Which versions receive security fixes is [PENDING MAINTAINER DECISION].

## Scope

In scope:

- the `moai` binary, its alias `moai-cli`, and their commands;
- files MoAI generates or applies into a project or a host configuration;
- hooks that MoAI generates;
- the installer and release artifacts, once they are published.

Out of scope (please report these to their owners):

- Claude Code itself (report to Anthropic) and Codex itself (report to OpenAI);
- third-party MCP servers.

## Handling secrets

MoAI ADK never asks you to paste keys or tokens into Issues, pull requests or chat. If you accidentally published a secret in this repository, revoke it with its provider first, then tell us through the private report above.
