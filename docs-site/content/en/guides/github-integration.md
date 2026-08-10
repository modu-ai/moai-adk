---
title: "GitHub Integration Guide"
description: "Parse issues and link them to SPECs with the moai github subcommand"
draft: false
weight: 10
---

MoAI-ADK's GitHub integration provides a lightweight CLI tool that parses
GitHub issues and links them to SPEC documents. Every command fetches the
current repository's issue data through the locally installed `gh` CLI.

> **Scope note**: This page covers only the `moai github` subcommand that
> actually ships and the GitHub Actions assets that come with it. The
> "multi-LLM review panel" that attaches multiple LLMs to a PR as a panel is
> not included in the current distributed release.

## Prerequisites

- MoAI-ADK installed (macOS · Linux · Windows)
- GitHub CLI (`gh`) installed and authenticated (`gh auth login`)
- A GitHub repository

## The moai github subcommand

`moai github` provides two active subcommands. Both support the `--dry-run`
flag, which lets you preview the work to be done without making any actual
changes.

### Issue parsing: `moai github parse-issue`

```bash
moai github parse-issue 123
```

Using the `gh` CLI, it fetches the issue with the given number and prints its
number · title · author · labels · body summary · comment count as a card.

### SPEC linking: `moai github link-spec`

```bash
moai github link-spec 123 SPEC-ISSUE-123
```

It creates a bidirectional link between a GitHub issue and a SPEC document, and
stores that mapping in `.moai/github-spec-registry.json`. The SPEC ID is
format-validated before it is stored.

```bash
# Check the plan only, without making actual changes
moai github link-spec 123 SPEC-ISSUE-123 --dry-run
```

## GitHub Actions assets shipped alongside

`moai init` deploys the following two assets under `.github/`.

### Label Sync workflow (`.github/workflows/label-sync.yml`)

It synchronizes repository labels, treating `.github/labels.yml` as the single
source of truth.

- **Trigger**: `workflow_dispatch` (manual, supports a `dry_run` input), or
  automatically when `.github/labels.yml` / the workflow file is pushed to
  `main`
- **Permissions**: `issues: write`, `pull-requests: write`, `contents: read`
- **Behavior**: the EndBug/label-sync action reflects `labels.yml` → repo labels

### detect-language composite action (`.github/actions/detect-language/action.yml`)

It detects the primary language based on the repository's first source-file
extension and emits it as the `language` output.

- **Supported languages (16)**: Go, Python, TypeScript, JavaScript, Rust, Java,
  Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift
- **Implementation note**: it uses `find ... -print -quit` to exit immediately
  after the first match, avoiding a broken-pipe failure under `set -o pipefail`

## Troubleshooting

### When the `gh` command is not found

The `moai github` subcommand depends on the local `gh` CLI. Confirm the
installation with `gh --version`, and finish authentication with
`gh auth login`.

### When an issue cannot be fetched

Check that the current directory is inside the target repository's working
tree, and that `gh` has access to that repo.

### SPEC ID validation failure

`link-spec` accepts only a valid SPEC ID that follows the `SPEC-` prefix.
Check the ID format and re-run.

## Next steps

- [CLI Reference](/en/workflow-commands/)
- [Workflow Settings Reference](/en/advanced/settings-json/)
- [Security Policy](/en/advanced/security-notes/)
