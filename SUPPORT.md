# Getting help

MoAI ADK is maintained by a small team. Please use the channel that fits your question so that answers stay findable for the next person.

| You want to… | Go to |
| --- | --- |
| Ask how to do something, or share an idea | [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) |
| Report a bug or propose a concrete change | [GitHub Issues](https://github.com/modu-ai/moai-adk/issues), using one of the Issue forms |
| Report a security vulnerability | the private channel in [SECURITY.md](SECURITY.md), never a public Issue |
| Contribute code or documentation | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Chat in real time | [Discord](https://discord.gg/Z7E7Mdc5aN) |

Discord is for conversation. Anything that should be remembered (a decision, a bug, a workaround) belongs in an Issue or a Discussion.

## Before you ask

- Say which generation you use: the Rust release candidate (branch `v2026.11.0`) or the Go 3.x release (branch `main`). Both commands are called `moai`, so include the output of `moai --version`: only the release candidate's line ends with `(rust)`.
- Include your OS and CPU architecture, and the host and its version (Claude Code or Codex).
- For an error, paste the exact command and its output, with secrets and personal paths removed.

## What to expect

The release candidate is not released yet and its host end-to-end tests have not run (see [docs/support-matrix.md](docs/support-matrix.md)), so some answers will be "not supported yet". Response times are best effort; a target has not been set yet (pending maintainer decision).

Documentation: [README](README.md) · [한국어 README](README.ko.md) · Go 3.x documentation at https://adk.mo.ai.kr
