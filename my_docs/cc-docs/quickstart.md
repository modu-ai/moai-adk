# Claude Code Quickstart Guide

## Before You Begin

Make sure you have:

- A terminal or command prompt open
- A code project to work with

## Step 1: Install Claude Code

### NPM Install

For Node.js 18 or newer:

```bash
npm install -g @anthropic-ai/claude-code
```

### Native Install

**macOS, Linux, WSL:**

```bash
curl -fsSL claude.ai/install.sh | bash
```

**Windows PowerShell:**

```powershell
irm https://claude.ai/install.ps1 | iex
```

## Step 2: Start Your First Session

Open your terminal in a project directory and start Claude Code:

```bash
cd /path/to/your/project
claude
```

## Step 3: Ask Your First Question

Try understanding your codebase:

```
> what does this project do?
> what technologies does this project use?
> where is the main entry point?
```

You can also ask about Claude Code's capabilities:

```
> what can Claude Code do?
> how do I use slash commands in Claude Code?
```

## Step 4: Make Your First Code Change

Try a simple task:

```
> add a hello world function to the main file
```

Claude Code will:

1. Find the appropriate file
2. Show proposed changes
3. Ask for your approval
4. Make the edit

## Step 5: Use Git with Claude Code

Make Git operations conversational:

```
> what files have I changed?
> commit my changes with a descriptive message
> create a new branch called feature/quickstart
```

## Essential Commands

| Command             | What it Does               | Example                             |
| ------------------- | -------------------------- | ----------------------------------- |
| `claude`            | Start interactive mode     | `claude`                            |
| `> "task"`          | Run a one-time task        | `> "fix the build error"`           |
| `claude -p "query"` | Run one-off query          | `claude -p "explain this function"` |
| `claude commit`     | Create a Git commit        | `claude commit`                     |
| `/clear`            | Clear conversation history | `/clear`                            |
| `/help`             | Get help on commands       | `/help`                             |
| `/config`           | Configure settings         | `/config`                           |

## Common Workflows

### Debugging

```
> the app crashes when I click submit
> can you investigate and fix it?
```

### Refactoring

```
> refactor this function to be more readable
> extract the validation logic into a separate function
```

### Testing

```
> write tests for the authentication module
> run the tests and fix any failures
```

### Documentation

```
> add JSDoc comments to all public functions
> create a README with setup instructions
```

## Tips for Success

1. **Be specific**: Clear instructions get better results
2. **Use context**: Reference files and functions directly
3. **Iterate**: Refine your requests based on results
4. **Review changes**: Always review before approving
5. **Learn shortcuts**: Use `/` commands for efficiency

## Next Steps

- Run `/init` to create a CLAUDE.md for your project
- Explore `/agents` to create specialized AI assistants
- Try `/mcp` to connect external tools
- Read the full documentation at docs.anthropic.com/claude-code
