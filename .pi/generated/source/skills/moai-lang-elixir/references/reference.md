# Elixir 1.17+ Complete Reference

## Language Features Reference

### Elixir 1.17 Feature Matrix

| Feature                | Status | Release | Production Ready |
| ---------------------- | ------ | ------- | ---------------- |
| set_theoretic_types    | Stable | 1.17    | Yes              |
| Duration module        | Stable | 1.17    | Yes              |
| Process.info/1 default | Stable | 1.17    | Yes              |
| Improved Documentation | Stable | 1.17    | Yes              |
| Pattern Matching       | Stable | 1.0+    | Yes              |
| Protocols              | Stable | 1.0+    | Yes              |
| Metaprogramming        | Stable | 1.0+    | Yes              |
| OTP Integration        | Stable | 1.0+    | Yes              |

### Set-Theoretic Types (Elixir 1.17+)

Type System Improvements:

```elixir
# Better type inference for guards
defmodule TypeExample do
  @spec check_value(integer() | binary()) :: :number | :string
  def check_value(value) when is_integer(value), do: :number
  def check_value(value) when is_binary(value), do: :string

  # Compiler now understands type narrowing
  @spec process(integer() | binary()) :: integer()
  def process(value) when is_integer(value) do
    value * 2
  end

  def process(value) when is_binary(value) do
    String.length(value)
  end
end
```

### Duration Module

New Duration type for time calculations:

```elixir
# Create durations
duration = Duration.new!(hour: 2, minute: 30, second: 15)

# Duration arithmetic
start_time = ~U[2024-01-01 10:00:00Z]
end_time = DateTime.add(start_time, duration)

# Duration components
duration.hour    # 2
duration.minute  # 30
duration.second  # 15

# Negative durations
negative_duration = Duration.new!(hour: -1, minute: -30)

# Duration comparison
Duration.compare(duration, negative_duration)  # :gt
```

---

## Phoenix Framework Reference

### Phoenix 1.7 Architecture

Application Structure:

```
lib/
├── my_app/              # Business logic context
│   ├── application.ex
│   ├── repo.ex
│   └── accounts/
│       └── user.ex
└── my_app_web/          # Web interface
    ├── endpoint.ex
    ├── router.ex
    ├── components/      # New in 1.7
    │   ├── core_components.ex
    │   └── layouts.ex
    ├── controllers/
    └── live/
```

### Phoenix.Component (Function Components)

Core Components Pattern:

```elixir
defmodule MyAppWeb.CoreComponents do
  use Phoenix.Component
  alias Phoenix.LiveView.JS

  @doc """
  Renders a button.
  """
  attr :type, :string, default: "button"
  attr :class, :string, default: nil
  attr :rest, :global, include: ~w(disabled form name value)

  slot :inner_block, required: true

  def button(assigns) do
    ~H"""
    <button
      type={@type}
      class={[
        "phx-submit-loading:opacity-75 rounded-lg bg-zinc-900 hover:bg-zinc-700",
        "py-2 px-3 text-sm font-semibold leading-6 text-white",
        @class
      ]}
      {@rest}
    >
      <%= render_slot(@inner_block) %>
    </button>
    """
  end

  @doc """
  Renders a table with sortable columns.
  """
  attr :id, :string, required: true
  attr :rows, :list, required: true
  attr :row_id, :any, default: nil, doc: "function or key to get row ID"
  attr :row_click, :any, default: nil
  attr :row_item, :any, default: &Function.identity/1

  slot :col, required: true do
    attr :label, :string
    attr :sortable, :boolean
  end

  slot :action, doc: "actions for each row"

  def table(assigns) do
    assigns =
      with %{rows: %Phoenix.LiveView.LiveStream{}} <- assigns do
        assign(assigns, row_id: assigns.row_id || fn {id, _item} -> id end)
      end

    ~H"""
    <div class="overflow-y-auto px-4 sm:overflow-visible sm:px-0">
      <table class="w-[40rem] mt-11 sm:w-full">
        <thead class="text-sm text-left leading-6 text-zinc-500">
          <tr>
            <th :for={col <- @col} class="p-0 pb-4 pr-6 font-normal"><%= col[:label] %></th>
            <th :if={@action != []} class="relative p-0 pb-4">
              <span class="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody class="relative divide-y divide-zinc-100 border-t border-zinc-200">
          <tr :for={row <- @rows} id={@row_id && @row_id.(row)} class="group hover:bg-zinc-50">
            <td
              :for={{col, i} <- Enum.with_index(@col)}
              phx-click={@row_click && @row_click.(row)}
              class={["relative p-0", @row_click && "hover:cursor-pointer"]}
            >
              <div class="block py-4 pr-6">
                <span class="absolute -inset-y-px right-0 -left-4 group-hover:bg-zinc-50 sm:rounded-l-xl" />
                <span class={["relative", i == 0 && "font-semibold text-zinc-900"]}>
                  <%= render_slot(col, @row_item.(row)) %>
                </span>
              </div>
            </td>
            <td :if={@action != []} class="relative w-14 p-0">
              <div class="relative whitespace-nowrap py-4 text-right text-sm font-medium">
                <span class="absolute -inset-y-px -right-4 left-0 group-hover:bg-zinc-50 sm:rounded-r-xl" />
                <span :for={action <- @action} class="relative ml-4 font-semibold leading-6 text-zinc-900 hover:text-zinc-700">
                  <%= render_slot(action, @row_item.(row)) %>
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    """
  end
end
```

### Verified Routes

Route Helpers Replacement:

```elixir
# router.ex
scope "/", MyAppWeb do
  pipe_through :browser

  get "/", PageController, :home
  live "/users", UserLive.Index, :index
  live "/users/:id", UserLive.Show, :show

  resources "/posts", PostController
end

# Usage with ~p sigil (compile-time verification)
~p"/"                          # "/"
~p"/users"                     # "/users"
~p"/users/#{user}"             # "/users/123"
~p"/users/#{user}/edit"        # "/users/123/edit"
~p"/posts/#{post}"             # "/posts/42"

# Path with query params
~p"/search?#{[q: "elixir", page: 2]}"  # "/search?q=elixir&page=2"

# Fragment
~p"/docs#installation"         # "/docs#installation"

# External URLs (requires uri_info configuration)
url(~p"/users")                # "https://example.com/users"
```

### Phoenix.LiveView Streams

Efficient List Updates:

```elixir
defmodule MyAppWeb.UserLive.Index do
  use MyAppWeb, :live_view

  @impl true
  def mount(_params, _session, socket) do
    {:ok, stream(socket, :users, Accounts.list_users())}
  end

  @impl true
  def handle_event("delete", %{"id" => id}, socket) do
    user = Accounts.get_user!(id)
    {:ok, _} = Accounts.delete_user(user)

    # Remove from stream (efficient DOM update)
    {:noreply, stream_delete(socket, :users, user)}
  end

  @impl true
  def handle_info({:user_created, user}, socket) do
    # Insert at beginning of stream
    {:noreply, stream_insert(socket, :users, user, at: 0)}
  end

  @impl true
  def handle_info({:user_updated, user}, socket) do
    # Update existing item in stream
    {:noreply, stream_insert(socket, :users, user)}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <.table id="users" rows={@streams.users}>
      <:col :let={{_id, user}} label="Name"><%= user.name %></:col>
      <:col :let={{_id, user}} label="Email"><%= user.email %></:col>
      <:action :let={{id, user}}>
        <.link phx-click={JS.push("delete", value: %{id: user.id}) |> hide("##{id}")}>
          Delete
        </.link>
      </:action>
    </.table>
    """
  end
end
```

---

## Ecto Reference

### Ecto 3.12 Query API

Query Composition Patterns:

```elixir
import Ecto.Query

# Basic queries
from u in User, select: u
from u in User, where: u.age > 18

# Join queries
from u in User,
  join: p in assoc(u, :posts),
  where: p.published == true,
  select: {u, count(p.id)},
  group_by: u.id

# Subqueries
popular_posts_query = from p in Post, where: p.views > 1000

from u in User,
  join: p in subquery(popular_posts_query),
  on: p.user_id == u.id

# Dynamic queries
def filter_users(query, filters) do
  Enum.reduce(filters, query, fn
    {:name, name}, query ->
      from q in query, where: ilike(q.name, ^"%#{name}%")

    {:min_age, age}, query ->
      from q in query, where: q.age >= ^age

    {:verified, true}, query ->
      from q in query, where: not is_nil(q.verified_at)

    _, query ->
      query
  end)
end

# Window functions
from p in Post,
  select: %{
    id: p.id,
    title: p.title,
    row_number: over(row_number(), :posts_partition)
  },
  windows: [posts_partition: [partition_by: p.user_id, order_by: p.inserted_at]]

# Common Table Expressions (CTE)
initial_query = from p in Post, where: p.published == true
recursive_query = from p in Post, where: p.parent_id in subquery(initial_query)

from p in Post,
  recursive_ctes: true,
  with_cte: "published_posts",
  as: ^union_all(initial_query, ^recursive_query)
```

### Ecto.Multi for Transactions

Complex Transaction Patterns:

```elixir
defmodule MyApp.Transfers do
  import Ecto.Query
  alias Ecto.Multi
  alias MyApp.{Repo, Account, Transaction}

  def transfer_money(from_account_id, to_account_id, amount) do
    Multi.new()
    |> Multi.run(:validate_amount, fn _repo, _changes ->
      if amount > 0, do: {:ok, amount}, else: {:error, :invalid_amount}
    end)
    |> Multi.run(:from_account, fn repo, _changes ->
      case repo.get(Account, from_account_id) do
        nil -> {:error, :from_account_not_found}
        account -> {:ok, account}
      end
    end)
    |> Multi.run(:to_account, fn repo, _changes ->
      case repo.get(Account, to_account_id) do
        nil -> {:error, :to_account_not_found}
        account -> {:ok, account}
      end
    end)
    |> Multi.run(:check_balance, fn _repo, %{from_account: account} ->
      if account.balance >= amount do
        {:ok, :sufficient_balance}
      else
        {:error, :insufficient_funds}
      end
    end)
    |> Multi.update(:withdraw, fn %{from_account: account} ->
      Account.changeset(account, %{balance: account.balance - amount})
    end)
    |> Multi.update(:deposit, fn %{to_account: account} ->
      Account.changeset(account, %{balance: account.balance + amount})
    end)
    |> Multi.insert(:transaction, fn %{from_account: from, to_account: to} ->
      Transaction.changeset(%Transaction{}, %{
        from_account_id: from.id,
        to_account_id: to.id,
        amount: amount,
        status: :completed
      })
    end)
    |> Multi.run(:notify, fn _repo, %{transaction: txn} ->
      # Send notifications
      MyApp.Notifications.send_transfer_notification(txn)
      {:ok, :notified}
    end)
    |> Repo.transaction()
  end
end
```

### Ecto Schema Types

Custom Type Definition:

```elixir
defmodule MyApp.Encrypted.Binary do
  use Ecto.Type

  def type, do: :binary

  def cast(value) when is_binary(value), do: {:ok, value}
  def cast(_), do: :error

  def load(value) when is_binary(value) do
    {:ok, decrypt(value)}
  end

  def dump(value) when is_binary(value) do
    {:ok, encrypt(value)}
  end

  defp encrypt(value) do
    # Implement encryption
    :crypto.strong_rand_bytes(16)
    |> then(&:crypto.crypto_one_time(:aes_256_cbc, key(), &1, value, true))
  end

  defp decrypt(value) do
    # Implement decryption
    <<iv::binary-16, ciphertext::binary>> = value
    :crypto.crypto_one_time(:aes_256_cbc, key(), iv, ciphertext, false)
  end

  defp key do
    Application.get_env(:my_app, :encryption_key)
  end
end

# Usage in schema
defmodule MyApp.Secret do
  use Ecto.Schema

  schema "secrets" do
    field :name, :string
    field :value, MyApp.Encrypted.Binary

    timestamps()
  end
end
```

---

## OTP Behaviors Reference

### GenServer Complete API

```elixir
defmodule MyApp.Cache do
  use GenServer

  # Client API

  def start_link(opts) do
    name = Keyword.get(opts, :name, __MODULE__)
    GenServer.start_link(__MODULE__, opts, name: name)
  end

  def get(server \\ __MODULE__, key) do
    GenServer.call(server, {:get, key})
  end

  def put(server \\ __MODULE__, key, value) do
    GenServer.cast(server, {:put, key, value})
  end

  def delete(server \\ __MODULE__, key) do
    GenServer.call(server, {:delete, key})
  end

  def clear(server \\ __MODULE__) do
    GenServer.cast(server, :clear)
  end

  # Server Callbacks

  @impl true
  def init(opts) do
    max_size = Keyword.get(opts, :max_size, 1000)
    ttl = Keyword.get(opts, :ttl, :timer.hours(1))

    state = %{
      cache: %{},
      max_size: max_size,
      ttl: ttl,
      access_order: []
    }

    {:ok, state}
  end

  @impl true
  def handle_call({:get, key}, _from, state) do
    case Map.get(state.cache, key) do
      nil ->
        {:reply, nil, state}

      {value, timestamp} ->
        if expired?(timestamp, state.ttl) do
          new_cache = Map.delete(state.cache, key)
          {:reply, nil, %{state | cache: new_cache}}
        else
          # Update access order (LRU)
          new_access_order = [key | List.delete(state.access_order, key)]
          {:reply, value, %{state | access_order: new_access_order}}
        end
    end
  end

  def handle_call({:delete, key}, _from, state) do
    new_cache = Map.delete(state.cache, key)
    new_access_order = List.delete(state.access_order, key)
    {:reply, :ok, %{state | cache: new_cache, access_order: new_access_order}}
  end

  @impl true
  def handle_cast({:put, key, value}, state) do
    timestamp = System.monotonic_time(:millisecond)
    new_cache = Map.put(state.cache, key, {value, timestamp})
    new_access_order = [key | List.delete(state.access_order, key)]

    state = %{state | cache: new_cache, access_order: new_access_order}

    # Evict oldest if over max_size
    state =
      if map_size(state.cache) > state.max_size do
        evict_oldest(state)
      else
        state
      end

    {:noreply, state}
  end

  def handle_cast(:clear, state) do
    {:noreply, %{state | cache: %{}, access_order: []}}
  end

  @impl true
  def handle_info(:cleanup, state) do
    now = System.monotonic_time(:millisecond)

    new_cache =
      state.cache
      |> Enum.reject(fn {_key, {_value, timestamp}} ->
        expired?(timestamp, state.ttl)
      end)
      |> Map.new()

    schedule_cleanup()
    {:noreply, %{state | cache: new_cache}}
  end

  # Private Functions

  defp expired?(timestamp, ttl) do
    System.monotonic_time(:millisecond) - timestamp > ttl
  end

  defp evict_oldest(state) do
    oldest_key = List.last(state.access_order)
    new_cache = Map.delete(state.cache, oldest_key)
    new_access_order = List.delete_at(state.access_order, -1)
    %{state | cache: new_cache, access_order: new_access_order}
  end

  defp schedule_cleanup do
    Process.send_after(self(), :cleanup, :timer.minutes(5))
  end
end
```

### Supervisor Strategies

```elixir
defmodule MyApp.Supervisor do
  use Supervisor

  def start_link(init_arg) do
    Supervisor.start_link(__MODULE__, init_arg, name: __MODULE__)
  end

  @impl true
  def init(_init_arg) do
    children = [
      # One-for-one: Only restart failed child
      {MyApp.Cache, name: MyApp.Cache, strategy: :one_for_one},

      # Rest-for-one: Restart failed child and all started after it
      Supervisor.child_spec({MyApp.DatabaseWorker, []}, restart: :permanent),
      Supervisor.child_spec({MyApp.ApiClient, []}, restart: :permanent),

      # One-for-all: Restart all children if one fails (use separate supervisor)
      {MyApp.ClusterSupervisor, strategy: :one_for_all}
    ]

    # Strategy options:
    # - :one_for_one - only restart failed process
    # - :one_for_all - restart all children if one fails
    # - :rest_for_one - restart failed and all started after it
    Supervisor.init(children, strategy: :one_for_one, max_restarts: 3, max_seconds: 5)
  end
end
```

---

## Mix Tasks Reference

### Common Mix Commands

```bash
# Project management
mix new my_app                    # Create new Mix project
mix new my_app --sup             # Create with supervision tree
mix phoenix.new my_app_web       # Create Phoenix project
mix phoenix.new my_app_web --no-ecto  # Without database

# Dependencies
mix deps.get                     # Fetch dependencies
mix deps.update <package>        # Update specific package
mix deps.clean <package>         # Clean compiled package
mix deps.tree                    # Show dependency tree

# Database (Ecto)
mix ecto.create                  # Create database
mix ecto.drop                    # Drop database
mix ecto.migrate                 # Run migrations
mix ecto.rollback                # Rollback last migration
mix ecto.rollback --step 3       # Rollback 3 migrations
mix ecto.reset                   # Drop, create, migrate
mix ecto.gen.migration add_users_table  # Generate migration

# Testing
mix test                         # Run all tests
mix test test/my_app_test.exs    # Run specific file
mix test --only slow             # Run tagged tests
mix test --cover                 # With coverage report
mix test --trace                 # Detailed test output

# Code quality
mix format                       # Format code
mix format --check-formatted     # Check formatting
mix credo                        # Static code analysis
mix dialyzer                     # Type checking

# Phoenix-specific
mix phx.server                   # Start Phoenix server
mix phx.gen.html Accounts User users name:string email:string  # Generate HTML resources
mix phx.gen.live Blog Post posts title:string body:text  # Generate LiveView resources
mix phx.gen.json Api User users name:string  # Generate JSON API
mix phx.gen.context Accounts User users name:string  # Generate context only
mix phx.routes                   # List all routes

# Release
mix release                      # Build release
mix release.init                 # Initialize release config
MIX_ENV=prod mix release         # Production release
```

### Custom Mix Task

```elixir
defmodule Mix.Tasks.MyApp.DatabaseReport do
  use Mix.Task

  @shortdoc "Generate database report"
  @moduledoc """
  Generate a report of database statistics.

      mix my_app.database_report

  Options:
      --format - Output format (json, csv, table). Default: table
      --output - Output file path
  """

  @impl Mix.Task
  def run(args) do
    Mix.Task.run("app.start")

    {opts, _args, _invalid} =
      OptionParser.parse(args,
        strict: [format: :string, output: :string],
        aliases: [f: :format, o: :output]
      )

    format = Keyword.get(opts, :format, "table")
    output_file = Keyword.get(opts, :output)

    report = generate_report()

    formatted_output =
      case format do
        "json" -> Jason.encode!(report, pretty: true)
        "csv" -> to_csv(report)
        _ -> to_table(report)
      end

    if output_file do
      File.write!(output_file, formatted_output)
      Mix.shell().info("Report written to #{output_file}")
    else
      Mix.shell().info(formatted_output)
    end
  end

  defp generate_report do
    alias MyApp.Repo

    %{
      total_users: Repo.aggregate(MyApp.Accounts.User, :count, :id),
      total_posts: Repo.aggregate(MyApp.Blog.Post, :count, :id),
      database_size: get_database_size()
    }
  end

  defp get_database_size do
    result = Repo.query!("SELECT pg_size_pretty(pg_database_size(current_database()))")
    result.rows |> List.first() |> List.first()
  end

  defp to_csv(report) do
    report
    |> Enum.map(fn {key, value} -> "#{key},#{value}" end)
    |> Enum.join("\n")
  end

  defp to_table(report) do
    report
    |> Enum.map(fn {key, value} ->
      "  #{String.pad_trailing(to_string(key), 20)} | #{value}"
    end)
    |> Enum.join("\n")
  end
end
```

---

## Configuration Patterns

### Runtime Configuration

```elixir
# config/runtime.exs
import Config

if config_env() == :prod do
  # Database configuration
  database_url =
    System.get_env("DATABASE_URL") ||
      raise """
      environment variable DATABASE_URL is missing.
      """

  maybe_ipv6 = if System.get_env("ECTO_IPV6") in ~w(true 1), do: [:inet6], else: []

  config :my_app, MyApp.Repo,
    url: database_url,
    pool_size: String.to_integer(System.get_env("POOL_SIZE") || "10"),
    socket_options: maybe_ipv6

  # Phoenix Endpoint configuration
  secret_key_base =
    System.get_env("SECRET_KEY_BASE") ||
      raise """
      environment variable SECRET_KEY_BASE is missing.
      """

  host = System.get_env("PHX_HOST") || "example.com"
  port = String.to_integer(System.get_env("PORT") || "4000")

  config :my_app, MyAppWeb.Endpoint,
    url: [host: host, port: 443, scheme: "https"],
    http: [
      ip: {0, 0, 0, 0, 0, 0, 0, 0},
      port: port
    ],
    secret_key_base: secret_key_base

  # Email configuration
  config :my_app, MyApp.Mailer,
    adapter: Swoosh.Adapters.Mailgun,
    api_key: System.get_env("MAILGUN_API_KEY"),
    domain: System.get_env("MAILGUN_DOMAIN")

  # Oban configuration
  config :my_app, Oban,
    repo: MyApp.Repo,
    queues: [
      default: String.to_integer(System.get_env("OBAN_DEFAULT_QUEUE") || "10"),
      mailers: String.to_integer(System.get_env("OBAN_MAILER_QUEUE") || "20"),
      events: String.to_integer(System.get_env("OBAN_EVENTS_QUEUE") || "50")
    ]
end
```

---

## Performance Optimization

### Query Optimization

```elixir
# Bad: N+1 queries
users = Repo.all(User)
Enum.map(users, fn user ->
  posts = Repo.all(from p in Post, where: p.user_id == ^user.id)
  {user, posts}
end)

# Good: Preloading
users = User |> Repo.all() |> Repo.preload(:posts)

# Good: Join with custom select
from u in User,
  left_join: p in assoc(u, :posts),
  group_by: u.id,
  select: %{user: u, post_count: count(p.id)}

# Use selectinload for has_many associations
Repo.all(from u in User, preload: [posts: :comments])

# Use joinedload for belongs_to associations
Repo.all(from p in Post, preload: [:user, :category])
```

### Process Optimization

```elixir
# Use Task.async_stream for parallel processing
1..100
|> Task.async_stream(&expensive_operation/1, max_concurrency: 10)
|> Enum.to_list()

# Use ETS for fast in-memory storage
:ets.new(:my_cache, [:named_table, :public, read_concurrency: true])
:ets.insert(:my_cache, {:key, "value"})
:ets.lookup(:my_cache, :key)

# Use Registry for process discovery
{:ok, _} = Registry.start_link(keys: :unique, name: MyApp.Registry)
Registry.register(MyApp.Registry, "user:123", nil)
Registry.lookup(MyApp.Registry, "user:123")
```

---

## Context7 Integration

Library ID Resolution:

```elixir
# Available Elixir libraries in Context7

# Elixir Language
library_id = "elixir-lang/elixir"
topics = "pattern matching, protocols, macros, processes"

# Phoenix Framework
library_id = "phoenixframework/phoenix"
topics = "controllers, LiveView, channels, PubSub, routing"

# Phoenix LiveView
library_id = "phoenixframework/phoenix_live_view"
topics = "components, hooks, streams, live navigation"

# Ecto
library_id = "elixir-ecto/ecto"
topics = "schemas, queries, changesets, multi, migrations"

# Oban
library_id = "sorentwo/oban"
topics = "workers, queues, cron, plugins"

# Absinthe (GraphQL)
library_id = "absinthe-graphql/absinthe"
topics = "schema, resolvers, subscriptions"
```

---

Last Updated: 2026-01-10
Version: 1.0.0
