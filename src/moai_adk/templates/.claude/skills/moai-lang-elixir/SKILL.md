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
    |> put_password_hash()
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

Set-Theoretic Types (Gradual Typing):
- Type annotations for better code documentation
- Compile-time warnings for type mismatches
- Enhanced dialyzer integration

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

Pipe Operator Best Practices:
```elixir
def process_order(order_params) do
  order_params
  |> validate_order()
  |> calculate_total()
  |> apply_discounts()
  |> create_order()
end

# With error handling using with
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
    {:ok, _} = MyAppWeb.Presence.track(socket, socket.assigns.user_id, %{
      online_at: System.system_time(:second)
    })
    {:noreply, socket}
  end

  @impl true
  def handle_in("new_message", %{"body" => body}, socket) do
    broadcast!(socket, "new_message", %{
      user_id: socket.assigns.user_id,
      body: body,
      inserted_at: DateTime.utc_now()
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
~p"/users?page=1"    # "/users?page=1"
```

### Ecto Advanced Patterns

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

  def by_email(query \\ base(), email) do
    from u in query, where: u.email == ^email
  end

  def with_posts(query \\ base()) do
    from u in query, preload: [:posts]
  end

  def order_by_recent(query \\ base()) do
    from u in query, order_by: [desc: u.inserted_at]
  end
end

# Usage
User
|> UserQuery.active()
|> UserQuery.with_posts()
|> UserQuery.order_by_recent()
|> Repo.all()
```

Embedded Schemas:
```elixir
defmodule MyApp.Order do
  use Ecto.Schema
  import Ecto.Changeset

  schema "orders" do
    field :status, :string
    embeds_one :shipping_address, Address, on_replace: :update
    embeds_many :items, Item, on_replace: :delete

    timestamps()
  end

  def changeset(order, attrs) do
    order
    |> cast(attrs, [:status])
    |> cast_embed(:shipping_address, required: true)
    |> cast_embed(:items, required: true)
  end
end

defmodule MyApp.Order.Address do
  use Ecto.Schema
  import Ecto.Changeset

  embedded_schema do
    field :street, :string
    field :city, :string
    field :zip, :string
  end

  def changeset(address, attrs) do
    address
    |> cast(attrs, [:street, :city, :zip])
    |> validate_required([:street, :city, :zip])
  end
end
```

### OTP Patterns

Supervisor Tree:
```elixir
defmodule MyApp.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      MyApp.Repo,
      MyAppWeb.Telemetry,
      {Phoenix.PubSub, name: MyApp.PubSub},
      MyAppWeb.Endpoint,
      {MyApp.Cache, []},
      {Task.Supervisor, name: MyApp.TaskSupervisor},
      MyApp.SchedulerSupervisor
    ]

    opts = [strategy: :one_for_one, name: MyApp.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
```

Dynamic Supervisor:
```elixir
defmodule MyApp.WorkerSupervisor do
  use DynamicSupervisor

  def start_link(init_arg) do
    DynamicSupervisor.start_link(__MODULE__, init_arg, name: __MODULE__)
  end

  @impl true
  def init(_init_arg) do
    DynamicSupervisor.init(strategy: :one_for_one)
  end

  def start_worker(args) do
    spec = {MyApp.Worker, args}
    DynamicSupervisor.start_child(__MODULE__, spec)
  end

  def stop_worker(pid) do
    DynamicSupervisor.terminate_child(__MODULE__, pid)
  end
end
```

Registry for Named Processes:
```elixir
# Start registry in application supervision tree
{Registry, keys: :unique, name: MyApp.Registry}

# GenServer with dynamic name
defmodule MyApp.Session do
  use GenServer

  def start_link(user_id) do
    GenServer.start_link(__MODULE__, user_id, name: via_tuple(user_id))
  end

  defp via_tuple(user_id) do
    {:via, Registry, {MyApp.Registry, {:session, user_id}}}
  end

  def get_session(user_id) do
    GenServer.call(via_tuple(user_id), :get)
  end
end
```

### ExUnit Testing

Async Tests with Setup:
```elixir
defmodule MyApp.AccountsTest do
  use MyApp.DataCase, async: true

  alias MyApp.Accounts

  describe "users" do
    setup do
      user = insert(:user)
      {:ok, user: user}
    end

    test "get_user!/1 returns the user with given id", %{user: user} do
      assert Accounts.get_user!(user.id) == user
    end

    test "create_user/1 with valid data creates a user" do
      valid_attrs = %{name: "Test", email: "test@example.com", password: "password123"}

      assert {:ok, %User{} = user} = Accounts.create_user(valid_attrs)
      assert user.name == "Test"
      assert user.email == "test@example.com"
    end

    test "create_user/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = Accounts.create_user(%{})
    end
  end
end
```

LiveView Testing:
```elixir
defmodule MyAppWeb.CounterLiveTest do
  use MyAppWeb.ConnCase

  import Phoenix.LiveViewTest

  test "renders counter", %{conn: conn} do
    {:ok, view, html} = live(conn, ~p"/counter")

    assert html =~ "Count: 0"
  end

  test "increments counter on click", %{conn: conn} do
    {:ok, view, _html} = live(conn, ~p"/counter")

    assert view
           |> element("button", "Increment")
           |> render_click() =~ "Count: 1"
  end
end
```

### Oban Background Jobs

Job Worker:
```elixir
defmodule MyApp.Workers.EmailWorker do
  use Oban.Worker, queue: :mailers, max_attempts: 3

  @impl Oban.Worker
  def perform(%Oban.Job{args: %{"email" => email, "template" => template}}) do
    case MyApp.Mailer.send_email(email, template) do
      {:ok, _} -> :ok
      {:error, reason} -> {:error, reason}
    end
  end
end

# Enqueue job
%{email: "user@example.com", template: "welcome"}
|> MyApp.Workers.EmailWorker.new()
|> Oban.insert()

# Scheduled job
%{email: "user@example.com", template: "reminder"}
|> MyApp.Workers.EmailWorker.new(scheduled_at: DateTime.add(DateTime.utc_now(), 3600))
|> Oban.insert()
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Production deployment with releases
- Distributed systems with libcluster
- Advanced LiveView patterns (streams, components)
- Telemetry and observability
- Security best practices
- CI/CD integration patterns

See:
- [reference.md](reference.md) - Complete reference documentation
- [examples.md](examples.md) - Production-ready code examples

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
- `moai-quality-testing` - TDD and testing strategies
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
