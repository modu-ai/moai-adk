---
name: moai-lang-elixir
description: Elixir functional programming, OTP patterns, and Phoenix framework best practices
version: 1.0.0
modularized: false
tags:
  - programming-language
  - enterprise
  - development
  - elixir
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, elixir, moai  


# Elixir Functional Programming — Enterprise

## Quick Reference (30 seconds)

Elixir builds high-concurrency distributed systems through functional programming, immutability, and OTP supervision trees. You can create self-healing systems with lightweight processes, pattern matching, and the "let it crash" philosophy.

**Core concepts**:
- **Immutability**: All data structures are immutable
- **Pattern matching**: Function clauses and case expressions
- **Concurrency**: Actor model-based lightweight processes (millions concurrent)
- **Self-healing**: OTP supervision trees (process restart, cascading recovery)

**Ecosystem** (November 2025):
- **Phoenix 1.7**: Web framework (routing, middleware, WebSocket)
- **LiveView**: Server-rendered real-time UI (no JavaScript required)
- **Ecto 3.11**: Database library (query builder, migrations)
- **GenServer**: Generic server behavior (state management, concurrency)


## Implementation Guide

### 1. Pattern Matching

**Basic pattern matching**:
```elixir
# Function clause-based pattern matching
defmodule Math do
  def factorial(0), do: 1
  def factorial(n) when n > 0 do
    n * factorial(n - 1)
  end

  def classify({:ok, value}), do: "Success: #{value}"
  def classify({:error, reason}), do: "Error: #{reason}"
end

iex> Math.factorial(5)
120
iex> Math.classify({:ok, "data"})
"Success: data"
```

**Destructuring**:
```elixir
defmodule User do
  def greet({:user, name, age}) do
    "Hello #{name}, age #{age}"
  end

  def extract_pair([h1, h2 | _tail]) do
    {h1, h2}
  end
end

iex> User.greet({:user, "Alice", 30})
"Hello Alice, age 30"
```

### 2. GenServer Pattern (State Management)

**Basic GenServer**:
```elixir
defmodule Counter do
  use GenServer

  def start_link(initial_count \\ 0) do
    GenServer.start_link(__MODULE__, initial_count, name: __MODULE__)
  end

  # Initialization
  def init(initial_count) do
    {:ok, initial_count}
  end

  # Synchronous call (wait for response)
  def increment(amount \\ 1) do
    GenServer.call(__MODULE__, {:increment, amount})
  end

  # Asynchronous call (no response wait)
  def log(message) do
    GenServer.cast(__MODULE__, {:log, message})
  end

  # Handler: synchronous call
  def handle_call({:increment, amount}, _from, count) do
    new_count = count + amount
    {:reply, new_count, new_count}
  end

  # Handler: asynchronous call
  def handle_cast({:log, message}, state) do
    IO.puts("Log: #{message}")
    {:noreply, state}
  end
end
```

### 3. OTP Supervision Trees

**Supervisor structure**:
```elixir
defmodule MyApp.Supervisor do
  use Supervisor

  def start_link(opts) do
    Supervisor.start_link(__MODULE__, opts, name: __MODULE__)
  end

  def init(_opts) do
    children = [
      # Database connection pool
      {Ecto.Repo, []},

      # Message broker
      {Phoenix.PubSub, name: MyApp.PubSub},

      # Custom worker
      {Counter, 0}
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end
end

# Supervision strategies:
# :one_for_one - Restart only child process
# :one_for_all - Restart all child processes
# :rest_for_one - Restart children after the failed one
```

### 4. Functional Programming Patterns

**Map/Filter/Reduce**:
```elixir
defmodule List do
  def process_data(list) do
    list
    |> Enum.map(&(&1 * 2))              # Double each element
    |> Enum.filter(&(&1 > 10))          # Filter 10 and above
    |> Enum.reduce(0, &+/2)             # Calculate sum
  end
end

iex> List.process_data([1, 5, 7, 10, 15])
64
```

**Pipe operator usage**:
```elixir
defmodule DataPipeline do
  def transform(data) do
    data
    |> parse_json()          # Step 1
    |> validate()            # Step 2
    |> enrich_with_metadata()  # Step 3
    |> persist_to_db()       # Step 4
  end
end
```

### 5. Ecto Database Patterns

**Changeset validation**:
```elixir
defmodule MyApp.User do
  use Ecto.Schema
  import Ecto.Changeset

  schema "users" do
    field :email, :string
    field :password, :string
    field :name, :string

    timestamps()
  end

  def create_changeset(user, attrs) do
    user
    |> cast(attrs, [:email, :password, :name])
    |> validate_required([:email, :password])
    |> validate_length(:password, min: 8)
    |> unique_constraint(:email)
  end
end
```

### 6. Phoenix LiveView Real-time UI

**Basic LiveView**:
```elixir
defmodule MyApp.CounterLive do
  use Phoenix.LiveView

  def mount(_params, _session, socket) do
    {:ok, assign(socket, count: 0)}
  end

  def render(assigns) do
    ~H"""
    <div>
      <h1>Count: <%= @count %></h1>
      <button phx-click="increment">+ 1</button>
      <button phx-click="decrement">- 1</button>
    </div>
    """
  end

  def handle_event("increment", _value, socket) do
    {:noreply, update(socket, :count, &(&1 + 1))}
  end

  def handle_event("decrement", _value, socket) do
    {:noreply, update(socket, :count, &(&1 - 1))}
  end
end
```


## Best Practices

### ✅ DO
- **Use pattern matching**: Instead of complex if/else, use pattern matching
- **Design supervision layers**: Systematic process management (let it crash)
- **Leverage immutability**: Ensure thread safety
- **Use Ecto Changesets**: Ensure data consistency
- **Use pipe operators**: Sequential transformation chains
- **Specify types**: Define function contracts (`@spec`)

### ❌ DON'T
- **Change state**: Create new state with Tuple/Map returns
- **Skip error handling**: Use `{:ok, result} | {:error, reason}` pattern
- **Ignore OTP**: Use standard library instead of reinventing
- **Deep nesting**: Improve readability with pattern matching
- **Overcreate processes**: Use pool structures


## Works Well With

- `moai-domain-backend` (backend architecture)
- `moai-domain-devops` (deployment and monitoring)
- `moai-domain-database` (data modeling)


**Version**: 2.0.0 | **Last Updated**: 2025-11-21 | **Lines**: 255
