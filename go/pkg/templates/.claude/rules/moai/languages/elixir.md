---
paths:
  - "**/*.ex"
  - "**/*.exs"
  - "**/mix.exs"
---

# Elixir Rules

Version: Elixir 1.17+

## Tooling

- Build: mix
- Testing: ExUnit
- Linting: Credo
- Formatting: mix format

## Best Practices (2026)

- Use pattern matching for control flow
- Use Phoenix 1.7+ contexts for domain separation
- Use OTP GenServer/Supervisor for stateful processes
- Use Task.Supervisor for concurrent operations
- Use Ecto.Multi for transactional operations

## LiveView 1.1 Patterns

```elixir
# Colocated Hooks - keep JavaScript close to Elixir
defmodule MyAppWeb.ChartLive do
  use MyAppWeb, :live_view

  # Hook is colocated with LiveView
  @impl true
  def mount(_params, _session, socket) do
    {:ok, assign(socket, data: [])}
  end
end
```

```javascript
// assets/js/hooks/chart.js
export const Chart = {
  mounted() {
    this.handleEvent("update-chart", ({data}) => {
      this.renderChart(data)
    })
  }
}
```

## OTP Patterns

```elixir
# Task.Supervisor for fire-and-forget tasks
Task.Supervisor.start_child(MyApp.TaskSupervisor, fn ->
  send_email(user)
end)

# Dynamic Supervisor for worker pools
DynamicSupervisor.start_child(MyApp.WorkerSupervisor, {Worker, args})
```

## Performance

- Use ETS for fast in-memory caching
- Use `Stream` for lazy enumeration
- Profile with `:observer.start()`
- Use connection pooling with DBConnection

## MoAI Integration

- Use Skill("moai-lang-elixir") for detailed patterns
- Follow TRUST 5 quality gates
