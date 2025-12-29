---
name: moai-lang-elixir
description: Elixir 1.17+ development specialist covering Phoenix 1.7, LiveView, Ecto, and OTP patterns. Use when developing real-time applications, distributed systems, or Phoenix projects.
version: 1.0.0
updated: 2025-12-07
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

Elixir 1.17+ Development Specialist - Phoenix 1.7, LiveView, Ecto, OTP patterns, and functional programming.

Auto-Triggers: `.ex`, `.exs` files, `mix.exs`, `config/`, Phoenix/LiveView discussions

Core Capabilities:
- Elixir 1.17: Pattern matching, pipes, protocols, behaviours, macros
- Phoenix 1.7: Controllers, LiveView, Channels, PubSub, Verified Routes
- Ecto: Schemas, Changesets, Queries, Migrations, Multi
- OTP: GenServer, Supervisor, Agent, Task, Registry
- ExUnit: Testing with setup, describe, async
- Mix: Build tool, tasks, releases
- Oban: Background job processing

### Quick Patterns

Phoenix Controller:
```elixir
defmodule MyAppWeb.UserController do
  use MyAppWeb, :controller

  alias MyApp.Accounts

  def show(conn, %{"id" => id}) do
    user = Accounts.get_user!(id)
    render(conn, :show, user: user)
  end

  def create(conn, %{"user" => user_params}) do
    case Accounts.create_user(user_params) do
      {:ok, user} ->
        conn
        |> put_flash(:info, "User created successfully.")
        |> redirect(to: ~p"/users/#{user}")

      {:error, %Ecto.Changeset{} = changeset} ->
        render(conn, :new, changeset: changeset)
    end
  end
end
```

Ecto Schema with Changeset:
```elixir
defmodule MyApp.Accounts.User do
  use Ecto.Schema
  import Ecto.Changeset

  schema "users" do
    field :name, :string
    field :email, :string
    field :password_hash, :string
    field :password, :string, virtual: true

    timestamps()
  end

  def changeset(user, attrs) do
    user
    |> cast(attrs, [:name, :email, :password])
    |> validate_required([:name, :email, :password])
    |> validate_format(:email, ~r/@/)
    |> validate_length(:password, min: 8)
    |> unique_constraint(:email)
  end
end
```

GenServer Pattern:
```elixir
defmodule MyApp.Counter do
  use GenServer

  def start_link(initial_value) do
    GenServer.start_link(__MODULE__, initial_value, name: __MODULE__)
  end

  def increment, do: GenServer.call(__MODULE__, :increment)
  def get_count, do: GenServer.call(__MODULE__, :get)

  @impl true
  def init(initial_value), do: {:ok, initial_value}

  @impl true
  def handle_call(:increment, _from, count), do: {:reply, count + 1, count + 1}
  def handle_call(:get, _from, count), do: {:reply, count, count}
end
```

---

## Implementation Guide (5 minutes)

### Elixir 1.17 Features

Pattern Matching Advanced:
```elixir
def process_message(%{type: :email, to: to} = message) when is_binary(to) do
  send_email(message)
end

def process_message(%{type: :sms, phone: phone}) when byte_size(phone) == 10 do
  send_sms(phone)
end

def process_message(_), do: {:error, :invalid_message}
```

Pipe Operator with Error Handling:
```elixir
def process_order_safe(params) do
  with {:ok, validated} <- validate_order(params),
       {:ok, total} <- calculate_total(validated),
       {:ok, discounted} <- apply_discounts(total),
       {:ok, order} <- create_order(discounted) do
    {:ok, order}
  else
    {:error, reason} -> {:error, reason}
  end
end
```

Protocols for Polymorphism:
```elixir
defprotocol Stringify do
  @doc "Converts a data structure to string"
  def to_string(data)
end

defimpl Stringify, for: Map do
  def to_string(map), do: Jason.encode!(map)
end

defimpl Stringify, for: List do
  def to_string(list), do: Enum.join(list, ", ")
end
```

### Phoenix 1.7 Patterns

LiveView Component:
```elixir
defmodule MyAppWeb.CounterLive do
  use MyAppWeb, :live_view

  def mount(_params, _session, socket) do
    {:ok, assign(socket, count: 0)}
  end

  def handle_event("increment", _, socket) do
    {:noreply, update(socket, :count, &(&1 + 1))}
  end

  def render(assigns) do
    ~H"""
    <div class="counter">
      <h1>Count: <%= @count %></h1>
      <button phx-click="increment">Increment</button>
    </div>
    """
  end
end
```

LiveView Form with Changesets:
```elixir
defmodule MyAppWeb.UserFormLive do
  use MyAppWeb, :live_view

  alias MyApp.Accounts
  alias MyApp.Accounts.User

  def mount(_params, _session, socket) do
    changeset = Accounts.change_user(%User{})
    {:ok, assign(socket, form: to_form(changeset))}
  end

  def handle_event("validate", %{"user" => user_params}, socket) do
    changeset =
      %User{}
      |> Accounts.change_user(user_params)
      |> Map.put(:action, :validate)

    {:noreply, assign(socket, form: to_form(changeset))}
  end

  def handle_event("save", %{"user" => user_params}, socket) do
    case Accounts.create_user(user_params) do
      {:ok, user} ->
        {:noreply,
         socket
         |> put_flash(:info, "User created!")
         |> push_navigate(to: ~p"/users/#{user}")}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, form: to_form(changeset))}
    end
  end

  def render(assigns) do
    ~H"""
    <.form for={@form} phx-change="validate" phx-submit="save">
      <.input field={@form[:name]} label="Name" />
      <.input field={@form[:email]} type="email" label="Email" />
      <.button>Save</.button>
    </.form>
    """
  end
end
```

Phoenix Channels:
```elixir
defmodule MyAppWeb.RoomChannel do
  use MyAppWeb, :channel

  @impl true
  def join("room:" <> room_id, _params, socket) do
    send(self(), :after_join)
    {:ok, assign(socket, :room_id, room_id)}
  end

  @impl true
  def handle_info(:after_join, socket) do
    push(socket, "presence_state", MyAppWeb.Presence.list(socket))
    {:noreply, socket}
  end

  @impl true
  def handle_in("new_message", %{"body" => body}, socket) do
    broadcast!(socket, "new_message", %{
      user_id: socket.assigns.user_id,
      body: body
    })
    {:noreply, socket}
  end
end
```

Verified Routes:
```elixir
# In router.ex
scope "/", MyAppWeb do
  pipe_through :browser

  live "/users", UserLive.Index, :index
  live "/users/:id", UserLive.Show, :show
end

# Usage with ~p sigil
~p"/users"           # "/users"
~p"/users/#{user}"   # "/users/123"
```

### Ecto Patterns

Multi for Transactions:
```elixir
def transfer_funds(from_account, to_account, amount) do
  Ecto.Multi.new()
  |> Ecto.Multi.update(:withdraw, withdraw_changeset(from_account, amount))
  |> Ecto.Multi.update(:deposit, deposit_changeset(to_account, amount))
  |> Ecto.Multi.insert(:transaction, fn %{withdraw: from, deposit: to} ->
    Transaction.changeset(%Transaction{}, %{
      from_account_id: from.id,
      to_account_id: to.id,
      amount: amount
    })
  end)
  |> Repo.transaction()
end
```

Query Composition:
```elixir
defmodule MyApp.Accounts.UserQuery do
  import Ecto.Query

  def base, do: from(u in User)

  def active(query \\ base()) do
    from u in query, where: u.active == true
  end

  def with_posts(query \\ base()) do
    from u in query, preload: [:posts]
  end
end

# Usage
User
|> UserQuery.active()
|> UserQuery.with_posts()
|> Repo.all()
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Production deployment with releases
- Distributed systems with libcluster
- Advanced LiveView patterns (streams, components)
- OTP supervision trees and dynamic supervisors
- Telemetry and observability
- Security best practices
- CI/CD integration patterns

See:
- [Advanced Patterns](modules/advanced-patterns.md) - Complete advanced patterns guide

---

## Context7 Library Mappings

```
/elixir-lang/elixir - Elixir language documentation
/phoenixframework/phoenix - Phoenix web framework
/phoenixframework/phoenix_live_view - LiveView real-time UI
/elixir-ecto/ecto - Database wrapper and query language
/sorentwo/oban - Background job processing
```

---

## Works Well With

- `moai-domain-backend` - REST API and microservices architecture
- `moai-domain-database` - SQL patterns and query optimization
- `moai-workflow-testing` - TDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-platform-deploy` - Deployment and infrastructure

---

## Troubleshooting

Common Issues:

Elixir Version Check:
```bash
elixir --version  # Should be 1.17+
mix --version     # Mix build tool version
```

Dependency Issues:
```bash
mix deps.get      # Fetch dependencies
mix deps.compile  # Compile dependencies
mix clean         # Clean build artifacts
```

Database Migrations:
```bash
mix ecto.create   # Create database
mix ecto.migrate  # Run migrations
mix ecto.rollback # Rollback last migration
```

Phoenix Server:
```bash
mix phx.server           # Start server
iex -S mix phx.server    # Start with IEx
MIX_ENV=prod mix release # Build release
```

LiveView Not Loading:
- Check websocket connection in browser console
- Verify endpoint configuration for websocket
- Ensure Phoenix.LiveView is in mix.exs dependencies

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
