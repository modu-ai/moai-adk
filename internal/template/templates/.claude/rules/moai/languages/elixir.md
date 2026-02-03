---
paths:
  - "**/*.ex"
  - "**/*.exs"
  - "**/mix.exs"
---

# Elixir Rules

Version: Elixir 1.17+

## Tooling

- Build: Mix
- Linting: Credo
- Formatting: mix format
- Testing: ExUnit
- Type checking: Dialyzer

## MUST

- Use pattern matching for control flow
- Use with for chained operations
- Use GenServer for stateful processes
- Use Supervisor trees for fault tolerance
- Write @spec for public functions
- Use @moduledoc and @doc for documentation

## MUST NOT

- Use try/catch for control flow
- Spawn processes without supervision
- Use mutable state (Agent misuse)
- Ignore dialyzer warnings
- Use string concatenation in loops
- Leave debug IO.inspect in production

## File Conventions

- *_test.exs for test files
- Use snake_case for modules and functions
- Use PascalCase for module names
- lib/ for application code
- test/ for test files

## Testing

- Use ExUnit with setup blocks
- Use Mox for mocking behaviors
- Use async: true for isolated tests
- Use ExMachina for test factories

## MoAI Integration

- Use Skill("moai-lang-elixir") for detailed patterns
- Follow TRUST 5 quality gates
