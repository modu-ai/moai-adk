# Model Configuration

Learn about the Claude Code model configuration, including model aliases like `opusplan`

## Available models

For the `model` setting in Claude Code, you can either configure:

- A **model alias**
- A full **model name**
- For Bedrock, an ARN

### Model aliases

Model aliases provide a convenient way to select model settings without remembering exact version numbers:

| Model alias | Behavior |
|------------|----------|
| **`default`** | Recommended model setting, depending on your account type |
| **`sonnet`** | Uses the latest Sonnet model (currently Sonnet 4) for daily coding tasks |
| **`opus`** | Uses the most capable Opus model (currently Opus 4.1) for complex reasoning |
| **`haiku`** | Uses the fast and efficient Haiku model for simple tasks |
| **`sonnet[1m]`** | Uses Sonnet with a 1 million token context window for long sessions |
| **`opusplan`** | Special mode that uses `opus` during plan mode, then switches to `sonnet` for execution |

### Setting your model

You can configure your model in several ways, listed in order of priority:

1. **During session** - Use `/model <alias|name>` to switch models mid-session
2. **At startup** - Launch with `claude --model <alias|name>`
3. **Environment variable** - Set `ANTHROPIC_MODEL=<alias|name>`
4. **Settings** - Configure permanently in your settings file using the `model` field.

Example usage:

```bash
# Start with Opus
claude --model opus

# Switch to Sonnet during session
/model sonnet
```

Example settings file:

```json
{
    "permissions": {
        "allow": ["Read(**)", "Write(**)"],
        "deny": []
    },
    "model": "opus"
}
```

## Special model behavior

### `default` model setting

The behavior of `default` depends on your account type:
- **Free accounts**: Uses the most appropriate free model
- **Pro accounts**: Uses the recommended Pro model setting
- **Team/Enterprise**: Uses organization-defined default

### `opusplan` hybrid mode

The `opusplan` alias enables intelligent model switching:

- **Plan phase**: Uses `opus` for complex reasoning, architecture decisions, and strategic planning
- **Execution phase**: Automatically switches to `sonnet` for code generation and implementation
- **Benefits**: Combines Opus's reasoning capabilities with Sonnet's efficiency

Example workflow with `opusplan`:

```bash
claude --model opusplan
> Plan a REST API for user authentication

# Uses Opus for planning phase
# Automatically switches to Sonnet for implementation
```

### Extended context models

You can enable extended context windows by adding the `[1m]` suffix to full model names:

```bash
# Enable 1 million token context window
claude --model claude-3-5-sonnet-20241022[1m]
```

**Note**: Extended context models have different pricing than standard models.

## Environment variables

### Model alias overrides

You can customize the behavior of model aliases using environment variables:

```bash
# Override default Opus model
export ANTHROPIC_DEFAULT_OPUS_MODEL="claude-3-opus-20240229"

# Override default Sonnet model
export ANTHROPIC_DEFAULT_SONNET_MODEL="claude-3-5-sonnet-20241022"

# Override default Haiku model
export ANTHROPIC_DEFAULT_HAIKU_MODEL="claude-3-haiku-20240307"
```

### Sub-agent model configuration

Control which model sub-agents use:

```bash
# Set model for sub-agents
export CLAUDE_CODE_SUBAGENT_MODEL="sonnet"
```

### Default model selection

Set the default model for all Claude Code sessions:

```bash
# Set default model
export ANTHROPIC_MODEL="opus"

# Or use model alias
export ANTHROPIC_MODEL="opusplan"
```

## Model switching during sessions

### In-session model changes

You can switch models at any time during a conversation:

```bash
# Switch to a different model
/model opus

# Switch to extended context
/model sonnet[1m]

# Switch to hybrid mode
/model opusplan
```

### Model-specific commands

Some commands may benefit from specific models:

```bash
# Use Opus for complex planning
/model opus
> Design the architecture for a microservices system

# Switch to Sonnet for implementation
/model sonnet
> Implement the user service from the plan above
```

## Best practices

### Choosing the right model

- **`opus`**: Complex reasoning, architecture decisions, code reviews
- **`sonnet`**: Daily coding tasks, debugging, refactoring
- **`haiku`**: Simple queries, quick fixes, documentation
- **`opusplan`**: Best of both worlds with automatic switching
- **`[1m]` models**: Large codebases, long conversation contexts

### Cost optimization

- Use `haiku` for simple tasks to minimize costs
- Use `opusplan` for balanced performance and cost
- Reserve `opus` for complex reasoning tasks
- Consider extended context models only when necessary

### Performance considerations

- `haiku`: Fastest response times
- `sonnet`: Good balance of speed and capability
- `opus`: Slower but most capable
- Extended context models: Slower due to increased context processing

## Configuration examples

### Personal development setup

```json
{
    "model": "sonnet",
    "permissions": {
        "allow": ["Read(**)", "Write(**)", "Bash"]
    }
}
```

### Team collaboration setup

```json
{
    "model": "opusplan",
    "permissions": {
        "allow": ["Read(**)", "Write(**/src/**)", "Bash(npm *)", "Bash(git *)"],
        "deny": ["Write(**/.env)", "Bash(rm *)"]
    }
}
```

### Enterprise setup with custom model

```json
{
    "model": "claude-3-opus-20240229",
    "permissions": {
        "allow": ["Read(**/src/**)", "Write(**/src/**)"],
        "deny": ["Bash(rm -rf *)"]
    },
    "env": {
        "ANTHROPIC_DEFAULT_OPUS_MODEL": "claude-3-opus-20240229",
        "CLAUDE_CODE_SUBAGENT_MODEL": "sonnet"
    }
}
```

## Troubleshooting

### Model not available

If a model is not available on your account:

1. Check your account type and available models
2. Verify the model name or alias is correct
3. Try using `default` as a fallback
4. Contact support if issues persist

### Performance issues

If experiencing slow responses:

1. Consider switching to `haiku` or `sonnet`
2. Avoid extended context models unless necessary
3. Clear conversation history if context becomes too large
4. Use `/model` to switch to a faster model mid-session

### Environment variable conflicts

If environment variables aren't working:

1. Restart your terminal session
2. Verify environment variable syntax
3. Check settings file doesn't override environment variables
4. Use `claude --model` flag as highest priority override
