---
name: moai-lang-elixir
version: 3.0.0
updated: "2025-11-19"
status: stable
description: Elixir and Phoenix framework best practices for concurrent systems, OTP patterns, LiveView applications, and production deployment. Use when building Elixir/Phoenix applications or concurrent systems.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Elixir & Phoenix Framework

Production-grade Elixir/Phoenix development with OTP, LiveView, and concurrent patterns.

## Quick Start

**Create Phoenix App**:

```bash
# Install Phoenix
mix archive.install hex phx_new

# Create new app
mix phx.new my_app --database postgres

# Setup database
cd my_app
mix ecto.create

# Start server
mix phx.server
# Visit: http://localhost:4000
```

**Simple LiveView**:

```elixir
defmodule MyAppWeb.CounterLive do
  use Phoenix.LiveView

  def mount(_params, _session, socket) do
    {:ok, assign(socket, count: 0)}
  end

  def render(assigns) do
    ~H"""
    <div>
      <h1>Count: <%= @count %></h1>
      <button phx-click="increment">+</button>
      <button phx-click="decrement">-</button>
    </div>
    """
  end

  def handle_event("increment", _,  socket) do
    {:noreply, update(socket, :count, &(&1 + 1))}
  end

  def handle_event("decrement", _, socket) do
    {:noreply, update(socket, :count, &(&1 - 1))}
  end
end
```

---

## Core Concepts

### Why Elixir?

| Feature             | Benefit                            | Use Case                           |
| ------------------- | ---------------------------------- | ---------------------------------- |
| **Concurrency**     | Millions of processes              | Real-time systems, chat, gaming    |
| **Fault Tolerance** | Supervision trees, self-healing    | High uptime requirements           |
| **Scalability**     | Distributed, horizontal scaling    | Microservices, distributed systems |
| **Productivity**    | Pattern matching, pipe operator    | Rapid development                  |
| **OTP**             | Battle-tested concurrent framework | Production systems                 |

---

## Pattern Matching

```elixir
# Basic patterns
{:ok, result} = {:ok, 42}  # result = 42
{:error, _} = {:error, "failed"}

# Function clauses
defmodule Math do
  def divide(_n, 0), do: {:error, "division by zero"}
  def divide(n, m), do: {:ok, n / m}
end

# With guards
def categorize(age) when age < 18, do: "minor"
def categorize(age) when age >= 18 and age < 65, do: "adult"
def categorize(_age), do: "senior"

# Case expressions
case HTTP.get(url) do
  {:ok, %{status: 200, body: body}} -> process(body)
  {:ok, %{status: 404}} -> {:error, :not_found}
  {:error, reason} -> {:error, reason}
end
```

---

## OTP Patterns

### GenServer (State Management)

```elixir
defmodule Counter do
  use GenServer

  # Client API
  def start_link(initial_value) do
    GenServer.start_link(__MODULE__, initial_value, name: __MODULE__)
  end

  def increment do
    GenServer.call(__MODULE__, :increment)
  end

  def get_value do
    GenServer.call(__MODULE__, :get)
  end

  # Server Callbacks
  @impl true
  def init(initial_value) do
    {:ok, initial_value}
  end

  @impl true
  def handle_call(:increment, _from, state) do
    {:reply, state + 1, state + 1}
  end

  @impl true
  def handle_call(:get, _from, state) do
    {:reply, state, state}
  end
end

# Usage
{:ok, _pid} = Counter.start_link(0)
Counter.increment()  # 1
Counter.get_value()  # 1
```

### Supervisor (Fault Tolerance)

```elixir
defmodule MyApp.Application do
  use Application

  def start(_type, _args) do
    children = [
      # Database connection pool
      {Ecto.Repo, repo: MyApp.Repo},

      # PubSub
      {Phoenix.PubSub, name: MyApp.PubSub},

      # Endpoint (HTTP server)
      MyAppWeb.Endpoint,

      # Custom workers
      {MyApp.Worker, name: MyApp.Worker},
    ]

    opts = [strategy: :one_for_one, name: MyApp.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
```

---

## Phoenix Framework

### Router

```elixir
defmodule MyAppWeb.Router do
  use MyAppWeb, :router

  pipeline :browser do
    plug :accepts, ["html"]
    plug :fetch_session
    plug :fetch_live_flash
    plug :put_root_layout, {MyAppWeb.LayoutView, :root}
    plug :protect_from_forgery
    plug :put_secure_browser_headers
  end

  pipeline :api do
    plug :accepts, ["json"]
    plug MyAppWeb.Auth.Pipeline
  end

  scope "/", MyAppWeb do
    pipe_through :browser

    live "/", PageLive, :index
    live "/users", UserLive.Index, :index
    live "/users/:id", UserLive.Show, :show
  end

  scope "/api", MyAppWeb do
    pipe_through :api

    resources "/users", UserController, except: [:new, :edit]
    post "/auth/login", AuthController, :login
  end
end
```

### Context (Business Logic)

```elixir
defmodule MyApp.Accounts do
  @moduledoc """
  The Accounts context - handles user management
 """

  alias MyApp.Repo
  alias MyApp.Accounts.User

  def list_users do
    Repo.all(User)
  end

  def get_user!(id), do: Repo.get!(User, id)

  def create_user(attrs \\ %{}) do
    %User{}
    |> User.changeset(attrs)
    |> Repo.insert()
  end

  def update_user(%User{} = user, attrs) do
    user
    |> User.changeset(attrs)
    |> Repo.update()
  end

  def delete_user(%User{} = user) do
    Repo.delete(user)
  end
end
```

---

## LiveView Real-Time

### Full LiveView Example

```elixir
defmodule MyAppWeb.DashboardLive do
  use MyAppWeb, :live_view

  alias MyApp.Metrics

  @impl true
  def mount(_params, _session, socket) do
    if connected?(socket) do
      # Subscribe to updates
      Phoenix.PubSub.subscribe(MyApp.PubSub, "metrics:updates")

      # Schedule periodic updates
      :timer.send_interval(5000, self(), :tick)
    end

    {:ok, assign(socket, metrics: load_metrics())}
  end

  @impl true
  def handle_info(:tick, socket) do
    {:noreply, assign(socket, metrics: load_metrics())}
  end

  @impl true
  def handle_info({:metric_updated, metric}, socket) do
    metrics = update_metric(socket.assigns.metrics, metric)
    {:noreply, assign(socket, metrics: metrics)}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="grid grid-cols-3 gap-4">
      <%= for metric <- @metrics do %>
        <div class="metric-card">
          <h3><%= metric.name %></h3>
          <p class="text-2xl"><%= metric.value %></p>
          <span class={"text-sm #{trend_color(metric.trend)}"}>
            <%= format_trend(metric.trend) %>
          </span>
        </div>
      <% end %>
    </div>
    """
  end

  defp load_metrics, do: Metrics.list_metrics()
  defp trend_color(:up), do: "text-green-500"
  defp trend_color(:down), do: "text-red-500"
  defp format_trend(:up), do: "↑"
  defp format_trend(:down), do: "↓"
end
```

---

## Ecto (Database)

### Schema & Changeset

```elixir
defmodule MyApp.Accounts.User do
  use Ecto.Schema
  import Ecto.Changeset

  schema "users" do
    field :email, :string
    field :name, :string
    field :age, :integer
    field :password, :string, virtual: true
    field :hashed_password, :string

    has_many :posts, MyApp.Content.Post

    timestamps()
  end

  def changeset(user, attrs) do
    user
    |> cast(attrs, [:email, :name, :age, :password])
    |> validate_required([:email, :name])
    |> validate_format(:email, ~r/@/)
    |> validate_number(:age, greater_than: 0)
    |> validate_length(:password, min: 8)
    |> unique_constraint(:email)
    |> hash_password()
  end

  defp hash_password(changeset) do
    case changeset do
      %Ecto.Changeset{valid?: true, changes: %{password: password}} ->
        put_change(changeset, :hashed_password, Bcrypt.hash_pwd_salt(password))
      _ ->
        changeset
    end
  end
end
```

### Queries

```elixir
import Ecto.Query

# Basic queries
Repo.all(User)
Repo.get(User, 1)
Repo.get_by(User, email: "user@example.com")

# Complex query
from(u in User,
  where: u.age > 18,
  join: p in assoc(u, :posts),
  where: p.published == true,
  preload: [posts: p],
  select: {u.name, count(p.id)},
  group_by: u.id,
  order_by: [desc: count(p.id)],
  limit: 10
)
|> Repo.all()

# Composition
query = from(u in User)
query = where(query, [u], u.age > 18)
query = order_by(query, [u], desc: u.inserted_at)
Repo.all(query)
```

---

## Testing

```elixir
defmodule MyApp.AccountsTest do
  use MyApp.DataCase, async: true

  alias MyApp.Accounts

  describe "users" do
    test "create_user/1 with valid data creates user" do
      attrs = %{email: "test@example.com", name: "Test User"}

      assert {:ok, user} = Accounts.create_user(attrs)
      assert user.email == "test@example.com"
      assert user.name == "Test User"
    end

    test "create_user/1 with invalid email returns error" do
      attrs = %{email: "invalid", name: "Test"}

      assert {:error, changeset} = Accounts.create_user(attrs)
      assert "has invalid format" in errors_on(changeset).email
    end
  end
end

# LiveView testing
defmodule MyAppWeb.CounterLiveTest do
  use MyAppWeb.ConnCase

  import Phoenix.LiveViewTest

  test "increments counter", %{conn: conn} do
    {:ok, view, _html} = live(conn, "/counter")

    assert view |> element("button", "+") |> render_click()
    assert render(view) =~ "Count: 1"
  end
end
```

---

## Production Deployment

### Release with Mix

```bash
# Build release
MIX_ENV=prod mix release

# Run release
_build/prod/rel/my_app/bin/my_app start

# Or as daemon
_build/prod/rel/my_app/bin/my_app daemon
```

### Docker

```dockerfile
FROM elixir:1.15-alpine AS builder

WORKDIR /app

# Install dependencies
RUN mix local.hex --force && \
    mix local.rebar --force

# Copy mix files
COPY mix.exs mix.lock ./
RUN mix deps.get --only prod

# Copy app
COPY . .

# Compile and build release
RUN MIX_ENV=prod mix compile
RUN MIX_ENV=prod mix release

# Runtime image
FROM alpine:3.18

RUN apk add --no-cache openssl ncurses-libs

WORKDIR /app

COPY --from=builder /app/_build/prod/rel/my_app ./

CMD ["./bin/my_app", "start"]
```

---

## Best Practices

✅ **DO**:

- Use pattern matching extensively
- Leverage OTP for concurrency
- Write tests (ExUnit is excellent)
- Use contexts for business logic
- Handle errors explicitly (`{:ok, val}` / `{:error, reason}`)
- Use LiveView for real-time UIs

❌ **DON'T**:

- Mutate state (use immutable data)
- Use global state (use GenServer/Agent)
- Skip supervision trees
- Ignore errors (always handle)
- Over-use macros

---

## Advanced Topics

For detailed patterns:

- **[examples.md](examples.md)**: Complete apps, real-time systems, distributed setups
- **[reference.md](reference.md)**: OTP behaviors, macros, metaprogramming

**Related Skills**:

- `moai-domain-backend`: Backend patterns
- `moai-essentials-perf`: Performance optimization
- `moai-testing-integration`: Testing strategies

---

**Ecosystem**: Elixir 1.15+, Phoenix 1.7+, LiveView 0.20+, Ecto 3.10+

**Version**: 3.0.0  
**Last Updated**: 2025-11-19  
**Status**: Production Ready
