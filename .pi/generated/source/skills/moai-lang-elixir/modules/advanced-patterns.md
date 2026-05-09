# Elixir Advanced Patterns

## Ecto Advanced Patterns

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
    embeds_one :shipping_adddess, Adddess, on_replace: :update
    embeds_many :items, Item, on_replace: :delete

    timestamps()
  end

  def changeset(order, attrs) do
    order
    |> cast(attrs, [:status])
    |> cast_embed(:shipping_adddess, required: true)
    |> cast_embed(:items, required: true)
  end
end

defmodule MyApp.Order.Adddess do
  use Ecto.Schema
  import Ecto.Changeset

  embedded_schema do
    field :street, :string
    field :city, :string
    field :zip, :string
  end

  def changeset(adddess, attrs) do
    adddess
    |> cast(attrs, [:street, :city, :zip])
    |> validate_required([:street, :city, :zip])
  end
end
```

## OTP Advanced Patterns

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

## ExUnit Advanced Testing

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

## Oban Background Jobs

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

Unique Jobs:
```elixir
defmodule MyApp.Workers.UniqueWorker do
  use Oban.Worker,
    queue: :default,
    unique: [period: 60, states: [:available, :scheduled, :executing]]

  @impl Oban.Worker
  def perform(%Oban.Job{args: args}) do
    # Only one job with these args will run within 60 seconds
    :ok
  end
end
```

## Production Deployment

Releases Configuration:
```elixir
# config/runtime.exs
import Config

if config_env() == :prod do
  database_url =
    System.get_env("DATABASE_URL") ||
      raise "DATABASE_URL environment variable is not set"

  config :my_app, MyApp.Repo,
    url: database_url,
    pool_size: String.to_integer(System.get_env("POOL_SIZE") || "10")

  secret_key_base =
    System.get_env("SECRET_KEY_BASE") ||
      raise "SECRET_KEY_BASE environment variable is not set"

  config :my_app, MyAppWeb.Endpoint,
    http: [port: String.to_integer(System.get_env("PORT") || "4000")],
    secret_key_base: secret_key_base
end
```

Dockerfile:
```dockerfile
FROM elixir:1.17-alpine AS build

RUN apk add --no-cache build-base npm git

WORKDIR /app

RUN mix local.hex --force && mix local.rebar --force

ENV MIX_ENV=prod

COPY mix.exs mix.lock ./
COPY config config
RUN mix deps.get --only $MIX_ENV
RUN mix deps.compile

COPY lib lib
COPY priv priv
COPY assets assets

RUN mix assets.deploy
RUN mix compile
RUN mix release

FROM alpine:3.18 AS app

RUN apk add --no-cache libstdc++ openssl ncurses-libs

WORKDIR /app

COPY --from=build /app/_build/prod/rel/my_app ./

ENV HOME=/app

CMD ["bin/my_app", "start"]
```

## Distributed Systems with libcluster

Cluster Configuration:
```elixir
# config/prod.exs
config :libcluster,
  topologies: [
    k8s: [
      strategy: Elixir.Cluster.Strategy.Kubernetes,
      config: [
        kubernetes_selector: "app=my-app",
        kubernetes_node_basename: "my_app"
      ]
    ]
  ]
```

Distributed GenServer:
```elixir
defmodule MyApp.DistributedCache do
  use GenServer

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts, name: {:global, __MODULE__})
  end

  def get(key) do
    GenServer.call({:global, __MODULE__}, {:get, key})
  end

  def put(key, value) do
    GenServer.call({:global, __MODULE__}, {:put, key, value})
  end

  @impl true
  def handle_call({:get, key}, _from, state) do
    {:reply, Map.get(state, key), state}
  end

  @impl true
  def handle_call({:put, key, value}, _from, state) do
    {:reply, :ok, Map.put(state, key, value)}
  end
end
```

## Telemetry and Observability

Telemetry Events:
```elixir
defmodule MyApp.Telemetry do
  require Logger

  def attach_handlers do
    :telemetry.attach_many(
      "my-app-handlers",
      [
        [:my_app, :repo, :query],
        [:phoenix, :endpoint, :stop],
        [:my_app, :api, :request]
      ],
      &handle_event/4,
      nil
    )
  end

  defp handle_event([:my_app, :repo, :query], measurements, metadata, _config) do
    Logger.debug(
      "Query: #{metadata.query} - Duration: #{measurements.total_time / 1_000_000}ms"
    )
  end

  defp handle_event([:phoenix, :endpoint, :stop], measurements, metadata, _config) do
    Logger.info(
      "Request: #{metadata.conn.request_path} - Duration: #{measurements.duration / 1_000_000}ms"
    )
  end
end
```

## Advanced LiveView Patterns

LiveView Streams:
```elixir
defmodule MyAppWeb.PostsLive do
  use MyAppWeb, :live_view

  def mount(_params, _session, socket) do
    posts = Posts.list_posts()
    {:ok, stream(socket, :posts, posts)}
  end

  def handle_event("delete", %{"id" => id}, socket) do
    post = Posts.get_post!(id)
    {:ok, _} = Posts.delete_post(post)
    {:noreply, stream_delete(socket, :posts, post)}
  end

  def render(assigns) do
    ~H"""
    <div id="posts" phx-update="stream">
      <div :for={{dom_id, post} <- @streams.posts} id={dom_id}>
        <%= post.title %>
        <button phx-click="delete" phx-value-id={post.id}>Delete</button>
      </div>
    </div>
    """
  end
end
```

LiveView Components:
```elixir
defmodule MyAppWeb.Components.Modal do
  use Phoenix.Component

  attr :id, :string, required: true
  attr :show, :boolean, default: false
  slot :inner_block, required: true

  def modal(assigns) do
    ~H"""
    <div
      id={@id}
      phx-mounted={@show && show_modal(@id)}
      phx-remove={hide_modal(@id)}
      class="relative z-50 hidden"
    >
      <div class="fixed inset-0 bg-black/50" />
      <div class="fixed inset-0 flex items-center justify-center">
        <div class="bg-white rounded-lg p-6">
          <%= render_slot(@inner_block) %>
        </div>
      </div>
    </div>
    """
  end
end
```

## Security Best Practices

Authentication with Guardian:
```elixir
defmodule MyApp.Guardian do
  use Guardian, otp_app: :my_app

  def subject_for_token(user, _claims) do
    {:ok, to_string(user.id)}
  end

  def resource_from_claims(%{"sub" => id}) do
    user = MyApp.Accounts.get_user!(id)
    {:ok, user}
  rescue
    Ecto.NoResultsError -> {:error, :resource_not_found}
  end
end

# Plug for authentication
defmodule MyAppWeb.AuthPipeline do
  use Guardian.Plug.Pipeline,
    otp_app: :my_app,
    module: MyApp.Guardian,
    error_handler: MyAppWeb.AuthErrorHandler

  plug Guardian.Plug.VerifyHeader, scheme: "Bearer"
  plug Guardian.Plug.EnsureAuthenticated
  plug Guardian.Plug.LoadResource
end
```

Input Validation:
```elixir
defmodule MyApp.Accounts.UserValidator do
  import Ecto.Changeset

  def validate_user_input(changeset) do
    changeset
    |> validate_required([:email, :password])
    |> validate_format(:email, ~r/^[^\s@]+@[^\s@]+\.[^\s@]+$/)
    |> validate_length(:password, min: 8, max: 72)
    |> validate_password_strength()
    |> unique_constraint(:email)
  end

  defp validate_password_strength(changeset) do
    validate_change(changeset, :password, fn :password, password ->
      cond do
        not String.match?(password, ~r/[A-Z]/) ->
          [password: "must contain at least one uppercase letter"]
        not String.match?(password, ~r/[a-z]/) ->
          [password: "must contain at least one lowercase letter"]
        not String.match?(password, ~r/[0-9]/) ->
          [password: "must contain at least one number"]
        true ->
          []
      end
    end)
  end
end
```
