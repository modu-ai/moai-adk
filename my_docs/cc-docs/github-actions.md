# Claude Code GitHub Actions

Learn about integrating Claude Code into your development workflow with Claude Code GitHub Actions.

## Why use Claude Code GitHub Actions?

- **Instant PR creation**: Describe what you need, and Claude creates a complete PR with all necessary changes
- **Automated code implementation**: Turn issues into working code with a single command
- **Follows your standards**: Claude respects your `CLAUDE.md` guidelines and existing code patterns
- **Simple setup**: Get started in minutes with our installer and API key
- **Secure by default**: Your code stays on Github's runners

## What can Claude do?

Claude Code provides a powerful GitHub Action that transforms how you work with code.

### Claude Code Action

This GitHub Action allows you to run Claude Code within your GitHub Actions workflows. You can use this to build any custom workflow on top of Claude Code.

## Setup

### Quick setup

The easiest way to set up this action is through Claude Code in the terminal. Just open claude and run `/install-github-app`. This command will guide you through setting up the GitHub app and required secrets.

*Note: You must be a repository admin to install the GitHub app and add secrets*

### Manual setup

If the `/install-github-app` command fails or you prefer manual setup:

1. **Install the Claude GitHub app** to your repository: [https://github.com/apps/claude](https://github.com/apps/claude)
2. **Add ANTHROPIC_API_KEY** to your repository secrets
3. **Copy the workflow file** from [examples/claude.yml](https://github.com/anthropics/claude-code-action/blob/main/examples/claude.yml) into your repository's `.github/workflows/`

## Upgrading from Beta

Claude Code GitHub Actions v1.0 introduces breaking changes that require updating your workflow files.

### Essential changes

1. **Update the action version**: Change `@beta` to `@v1`
2. **Remove mode configuration**: Delete `mode: "tag"` or `mode: "agent"`
3. **Update prompt inputs**: Replace `direct_prompt` with `prompt`
4. **Move CLI options**: Convert `max_turns`, `system_prompt`, etc. to `claude_args`

### Automatic mode detection

The action now automatically detects whether to run in:
- **Interactive mode**: Responds to `@claude` mentions in issues and PRs
- **Automation mode**: Runs immediately with a provided prompt

## Configuration Examples

### Basic workflow configuration

```yaml
name: Claude Code Action
on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]

jobs:
  claude:
    runs-on: ubuntu-latest
    steps:
      - uses: anthropics/claude-code-action@v1
        with:
          anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
```

### Advanced configuration with custom options

```yaml
- uses: anthropics/claude-code-action@v1
  with:
    prompt: "Review this PR for security issues"
    anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
    claude_args: |
      --system-prompt "Follow our coding standards"
      --max-turns 10
```

## Example Use Cases

- **Turn Issues into PRs**:
  ```
  @claude implement this feature based on the issue description
  ```

- **Get Implementation Help**:
  ```
  @claude how should I implement user authentication for this endpoint?
  ```

- **Fix Bugs Quickly**:
  ```
  @claude fix the TypeError in the user dashboard component
  ```

## Best Practices

- Create a `CLAUDE.md` to define coding standards
- Use GitHub Secrets for API keys
- Review Claude's suggestions before merging
- Optimize performance with concise configurations
- Be mindful of CI/API costs

## Cloud Provider Integration

Supports integration with:

- **Anthropic API** (Direct)
- **AWS Bedrock**
- **Google Vertex AI**

### Authentication Requirements

- Cloud project with AI services enabled
- Workload Identity Federation (for cloud providers)
- Service account with permissions
- GitHub App (recommended)

## Cost Considerations

- Consumes GitHub Actions minutes
- API token usage varies by task complexity
- Optimize by:
  - Using specific `@claude` commands
  - Setting `max_turns` limits
  - Configuring reasonable timeouts

## SDK Integration

Claude Code GitHub Actions is built on top of the Claude Code SDK, which enables programmatic integration of Claude Code into your applications. You can use the SDK to build custom automation workflows beyond GitHub Actions.

## Workflow Configuration

Detailed workflow examples available in the [Claude Code Action repository](https://github.com/anthropics/claude-code-action)

**Note**: Claude Code GitHub Actions v1.0 is now generally available (GA) as of 2025.
