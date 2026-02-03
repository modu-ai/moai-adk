# Elixir Production-Ready Code Examples

## Complete Phoenix Application

### Project Structure

```
phoenix_app/
├── lib/
│   ├── my_app/
│   │   ├── application.ex
│   │   ├── repo.ex
│   │   ├── accounts/
│   │   │   ├── user.ex
│   │   │   └── user_token.ex
│   │   ├── mailer.ex
│   │   └── workers/
│   │       └── email_worker.ex
│   └── my_app_web/
│       ├── endpoint.ex
│       ├── router.ex
│       ├── telemetry.ex
│       ├── components/
│       │   ├── core_components.ex
│       │   └── layouts.ex
│       ├── controllers/
│       │   ├── user_controller.ex
│       │   └── user_json.ex
│       └── live/
│           ├── user_live/
│           │   ├── index.ex
│           │   ├── show.ex
│           │   └── form_component.ex
│           └── user_auth.ex
├── test/
│   ├── my_app/
│   │   └── accounts_test.exs
│   ├── my_app_web/
│   │   ├── controllers/
│   │   │   └── user_controller_test.exs
│   │   └── live/
│   │       └── user_live_test.exs
│   ├── support/
│   │   ├── conn_case.ex
│   │   ├── data_case.ex
│   │   └── fixtures/
│   │       └── accounts_fixtures.ex
│   └── test_helper.exs
├── priv/
│   └── repo/
│       ├── migrations/
│       └── seeds.exs
├── config/
│   ├── config.exs
│   ├── dev.exs
│   ├── test.exs
│   ├── prod.exs
│   └── runtime.exs
├── mix.exs
└── Dockerfile
```

### Application Entry Point

```elixir
# lib/my_app/application.ex
defmodule MyApp.Application do
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    children = [
      # Telemetry supervisor
      MyAppWeb.Telemetry,

      # Database connection
      MyApp.Repo,

      # PubSub system
      {Phoenix.PubSub, name: MyApp.PubSub},

      # Finch HTTP client
      {Finch, name: MyApp.Finch},

      # Background job processing
      {Oban, Application.fetch_env!(:my_app, Oban)},

      # Presence tracking
      MyAppWeb.Presence,

      # Endpoint - must be last
      MyAppWeb.Endpoint
    ]

    opts = [strategy: :one_for_one, name: MyApp.Supervisor]
    Supervisor.start_link(children, opts)
  end

  @impl true
  def config_change(changed, _new, removed) do
    MyAppWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end
```

### Configuration

```elixir
# config/config.exs
import Config

config :my_app,
  ecto_repos: [MyApp.Repo],
  generators: [timestamp_type: :utc_datetime]

config :my_app, MyApp.Repo,
  migration_primary_key: [type: :binary_id],
  migration_timestamps: [type: :utc_datetime]

config :my_app, MyAppWeb.Endpoint,
  url: [host: "localhost"],
  adapter: Bandit.PhoenixAdapter,
  render_errors: [
    formats: [html: MyAppWeb.ErrorHTML, json: MyAppWeb.ErrorJSON],
    layout: false
  ],
  pubsub_server: MyApp.PubSub,
  live_view: [signing_salt: "YOUR_SECRET"]

config :my_app, Oban,
  engine: Oban.Engines.Basic,
  queues: [default: 10, mailers: 20, events: 50],
  repo: MyApp.Repo

config :esbuild,
  version: "0.17.11",
  my_app: [
    args: ~w(js/app.js --bundle --target=es2017 --outdir=../priv/static/assets),
    cd: Path.expand("../assets", __DIR__),
    env: %{"NODE_PATH" => Path.expand("../deps", __DIR__)}
  ]

config :tailwind,
  version: "3.4.0",
  my_app: [
    args: ~w(
      --config=tailwind.config.js
      --input=css/app.css
      --output=../priv/static/assets/app.css
    ),
    cd: Path.expand("../assets", __DIR__)
  ]

import_config "#{config_env()}.exs"
```

```elixir
# config/runtime.exs
import Config

if config_env() == :prod do
  database_url =
    System.get_env("DATABASE_URL") ||
      raise """
      environment variable DATABASE_URL is missing.
      For example: ecto://USER:PASS@HOST/DATABASE
      """

  config :my_app, MyApp.Repo,
    url: database_url,
    pool_size: String.to_integer(System.get_env("POOL_SIZE") || "10"),
    ssl: true,
    ssl_opts: [
      verify: :verify_peer,
      cacertfile: "/etc/ssl/cert.pem",
      server_name_indication: ~c"myapp.fly.dev",
      customize_hostname_check: [
        match_fun: :public_key.pkix_verify_hostname_match_fun(:https)
      ]
    ]

  secret_key_base =
    System.get_env("SECRET_KEY_BASE") ||
      raise """
      environment variable SECRET_KEY_BASE is missing.
      You can generate one by calling: mix phx.gen.secret
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
end
```

### Ecto Schema and Changeset

```elixir
# lib/my_app/accounts/user.ex
defmodule MyApp.Accounts.User do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id

  schema "users" do
    field :email, :string
    field :name, :string
    field :password, :string, virtual: true, redact: true
    field :hashed_password, :string, redact: true
    field :confirmed_at, :utc_datetime
    field :is_admin, :boolean, default: false

    has_many :tokens, MyApp.Accounts.UserToken
    has_many :posts, MyApp.Blog.Post

    timestamps(type: :utc_datetime)
  end

  @doc """
  A user changeset for registration.
  """
  def registration_changeset(user, attrs, opts \\ []) do
    user
    |> cast(attrs, [:email, :name, :password])
    |> validate_email(opts)
    |> validate_name()
    |> validate_password(opts)
  end

  defp validate_email(changeset, opts) do
    changeset
    |> validate_required([:email])
    |> validate_format(:email, ~r/^[^\s]+@[^\s]+$/, message: "must have the @ sign and no spaces")
    |> validate_length(:email, max: 160)
    |> maybe_validate_unique_email(opts)
  end

  defp validate_name(changeset) do
    changeset
    |> validate_required([:name])
    |> validate_length(:name, min: 2, max: 100)
  end

  defp validate_password(changeset, opts) do
    changeset
    |> validate_required([:password])
    |> validate_length(:password, min: 12, max: 72)
    |> validate_format(:password, ~r/[a-z]/, message: "at least one lower case character")
    |> validate_format(:password, ~r/[A-Z]/, message: "at least one upper case character")
    |> validate_format(:password, ~r/[!?@#$%^&*_0-9]/, message: "at least one digit or punctuation character")
    |> maybe_hash_password(opts)
  end

  defp maybe_hash_password(changeset, opts) do
    hash_password? = Keyword.get(opts, :hash_password, true)
    password = get_change(changeset, :password)

    if hash_password? && password && changeset.valid? do
      changeset
      |> validate_length(:password, max: 72, count: :bytes)
      |> put_change(:hashed_password, Bcrypt.hash_pwd_salt(password))
      |> delete_change(:password)
    else
      changeset
    end
  end

  defp maybe_validate_unique_email(changeset, opts) do
    if Keyword.get(opts, :validate_email, true) do
      changeset
      |> unsafe_validate_unique(:email, MyApp.Repo)
      |> unique_constraint(:email)
    else
      changeset
    end
  end

  @doc """
  A user changeset for changing the email.
  """
  def email_changeset(user, attrs, opts \\ []) do
    user
    |> cast(attrs, [:email])
    |> validate_email(opts)
    |> case do
      %{changes: %{email: _}} = changeset -> changeset
      %{} = changeset -> add_error(changeset, :email, "did not change")
    end
  end

  @doc """
  Verifies the password.
  """
  def valid_password?(%MyApp.Accounts.User{hashed_password: hashed_password}, password)
      when is_binary(hashed_password) and byte_size(password) > 0 do
    Bcrypt.verify_pass(password, hashed_password)
  end

  def valid_password?(_, _) do
    Bcrypt.no_user_verify()
    false
  end
end
```

### Context Module

```elixir
# lib/my_app/accounts.ex
defmodule MyApp.Accounts do
  @moduledoc """
  The Accounts context.
  """

  import Ecto.Query, warn: false
  alias MyApp.Repo
  alias MyApp.Accounts.{User, UserToken}

  ## User Registration

  @doc """
  Registers a user.
  """
  def register_user(attrs) do
    %User{}
    |> User.registration_changeset(attrs)
    |> Repo.insert()
  end

  ## User Queries

  @doc """
  Gets a user by email.
  """
  def get_user_by_email(email) when is_binary(email) do
    Repo.get_by(User, email: email)
  end

  @doc """
  Gets a user by email and password.
  """
  def get_user_by_email_and_password(email, password)
      when is_binary(email) and is_binary(password) do
    user = Repo.get_by(User, email: email)
    if User.valid_password?(user, password), do: user
  end

  @doc """
  Gets a single user.
  """
  def get_user!(id), do: Repo.get!(User, id)

  @doc """
  Lists all users with optional filters.
  """
  def list_users(filters \\ %{}) do
    User
    |> apply_filters(filters)
    |> Repo.all()
  end

  defp apply_filters(query, filters) do
    Enum.reduce(filters, query, fn
      {:search, term}, query ->
        pattern = "%#{term}%"
        where(query, [u], ilike(u.name, ^pattern) or ilike(u.email, ^pattern))

      {:is_admin, value}, query ->
        where(query, [u], u.is_admin == ^value)

      {:confirmed, true}, query ->
        where(query, [u], not is_nil(u.confirmed_at))

      {:confirmed, false}, query ->
        where(query, [u], is_nil(u.confirmed_at))

      _, query ->
        query
    end)
  end

  @doc """
  Paginated user listing.
  """
  def list_users_paginated(page \\ 1, per_page \\ 20) do
    offset = (page - 1) * per_page

    query =
      from u in User,
        order_by: [desc: u.inserted_at],
        limit: ^per_page,
        offset: ^offset

    users = Repo.all(query)
    total_count = Repo.aggregate(User, :count, :id)

    %{
      users: users,
      page: page,
      per_page: per_page,
      total_count: total_count,
      total_pages: ceil(total_count / per_page)
    }
  end

  ## User Updates

  @doc """
  Updates a user.
  """
  def update_user(%User{} = user, attrs) do
    user
    |> User.email_changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Confirms a user by setting `confirmed_at`.
  """
  def confirm_user(user) do
    user
    |> Ecto.Changeset.change(confirmed_at: DateTime.utc_now() |> DateTime.truncate(:second))
    |> Repo.update()
  end

  ## Session Tokens

  @doc """
  Generates a session token.
  """
  def generate_user_session_token(user) do
    {token, user_token} = UserToken.build_session_token(user)
    Repo.insert!(user_token)
    token
  end

  @doc """
  Gets the user with the given signed token.
  """
  def get_user_by_session_token(token) do
    {:ok, query} = UserToken.verify_session_token_query(token)
    Repo.one(query)
  end

  @doc """
  Deletes the signed token with the given context.
  """
  def delete_user_session_token(token) do
    Repo.delete_all(UserToken.by_token_and_context_query(token, "session"))
    :ok
  end
end
```

### Phoenix Controller

```elixir
# lib/my_app_web/controllers/user_controller.ex
defmodule MyAppWeb.UserController do
  use MyAppWeb, :controller

  alias MyApp.Accounts
  alias MyApp.Accounts.User

  action_fallback MyAppWeb.FallbackController

  def index(conn, params) do
    page = Map.get(params, "page", "1") |> String.to_integer()
    per_page = Map.get(params, "per_page", "20") |> String.to_integer()

    result = Accounts.list_users_paginated(page, per_page)
    render(conn, :index, result)
  end

  def create(conn, %{"user" => user_params}) do
    with {:ok, %User{} = user} <- Accounts.register_user(user_params) do
      conn
      |> put_status(:created)
      |> put_resp_header("location", ~p"/api/users/#{user}")
      |> render(:show, user: user)
    end
  end

  def show(conn, %{"id" => id}) do
    user = Accounts.get_user!(id)
    render(conn, :show, user: user)
  end

  def update(conn, %{"id" => id, "user" => user_params}) do
    user = Accounts.get_user!(id)

    with {:ok, %User{} = user} <- Accounts.update_user(user, user_params) do
      render(conn, :show, user: user)
    end
  end

  def delete(conn, %{"id" => id}) do
    user = Accounts.get_user!(id)

    with {:ok, %User{}} <- Accounts.delete_user(user) do
      send_resp(conn, :no_content, "")
    end
  end
end
```

```elixir
# lib/my_app_web/controllers/user_json.ex
defmodule MyAppWeb.UserJSON do
  alias MyApp.Accounts.User

  @doc """
  Renders a list of users.
  """
  def index(%{users: users, page: page, per_page: per_page, total_count: total_count, total_pages: total_pages}) do
    %{
      data: for(user <- users, do: data(user)),
      pagination: %{
        page: page,
        per_page: per_page,
        total_count: total_count,
        total_pages: total_pages
      }
    }
  end

  @doc """
  Renders a single user.
  """
  def show(%{user: user}) do
    %{data: data(user)}
  end

  defp data(%User{} = user) do
    %{
      id: user.id,
      email: user.email,
      name: user.name,
      is_admin: user.is_admin,
      confirmed_at: user.confirmed_at,
      inserted_at: user.inserted_at
    }
  end
end
```

### Phoenix LiveView

```elixir
# lib/my_app_web/live/user_live/index.ex
defmodule MyAppWeb.UserLive.Index do
  use MyAppWeb, :live_view

  alias MyApp.Accounts
  alias MyApp.Accounts.User

  @impl true
  def mount(_params, _session, socket) do
    if connected?(socket) do
      Phoenix.PubSub.subscribe(MyApp.PubSub, "users")
    end

    {:ok, stream(socket, :users, Accounts.list_users())}
  end

  @impl true
  def handle_params(params, _url, socket) do
    {:noreply, apply_action(socket, socket.assigns.live_action, params)}
  end

  defp apply_action(socket, :edit, %{"id" => id}) do
    socket
    |> assign(:page_title, "Edit User")
    |> assign(:user, Accounts.get_user!(id))
  end

  defp apply_action(socket, :new, _params) do
    socket
    |> assign(:page_title, "New User")
    |> assign(:user, %User{})
  end

  defp apply_action(socket, :index, _params) do
    socket
    |> assign(:page_title, "Listing Users")
    |> assign(:user, nil)
  end

  @impl true
  def handle_info({MyAppWeb.UserLive.FormComponent, {:saved, user}}, socket) do
    {:noreply, stream_insert(socket, :users, user)}
  end

  @impl true
  def handle_info({:user_created, user}, socket) do
    {:noreply, stream_insert(socket, :users, user, at: 0)}
  end

  @impl true
  def handle_event("delete", %{"id" => id}, socket) do
    user = Accounts.get_user!(id)
    {:ok, _} = Accounts.delete_user(user)

    {:noreply, stream_delete(socket, :users, user)}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <.header>
      Listing Users
      <:actions>
        <.link patch={~p"/users/new"}>
          <.button>New User</.button>
        </.link>
      </:actions>
    </.header>

    <.table id="users" rows={@streams.users} row_click={fn {_id, user} -> JS.navigate(~p"/users/#{user}") end}>
      <:col :let={{_id, user}} label="Name"><%= user.name %></:col>
      <:col :let={{_id, user}} label="Email"><%= user.email %></:col>
      <:col :let={{_id, user}} label="Confirmed">
        <%= if user.confirmed_at, do: "Yes", else: "No" %>
      </:col>
      <:action :let={{_id, user}}>
        <div class="sr-only">
          <.link navigate={~p"/users/#{user}"}>Show</.link>
        </div>
        <.link patch={~p"/users/#{user}/edit"}>Edit</.link>
      </:action>
      <:action :let={{id, user}}>
        <.link phx-click={JS.push("delete", value: %{id: user.id}) |> hide("##{id}")} data-confirm="Are you sure?">
          Delete
        </.link>
      </:action>
    </.table>

    <.modal :if={@live_action in [:new, :edit]} id="user-modal" show on_cancel={JS.patch(~p"/users")}>
      <.live_component
        module={MyAppWeb.UserLive.FormComponent}
        id={@user.id || :new}
        title={@page_title}
        action={@live_action}
        user={@user}
        patch={~p"/users"}
      />
    </.modal>
    """
  end
end
```

```elixir
# lib/my_app_web/live/user_live/form_component.ex
defmodule MyAppWeb.UserLive.FormComponent do
  use MyAppWeb, :live_component

  alias MyApp.Accounts

  @impl true
  def render(assigns) do
    ~H"""
    <div>
      <.header>
        <%= @title %>
        <:subtitle>Use this form to manage user records in your database.</:subtitle>
      </.header>

      <.simple_form
        for={@form}
        id="user-form"
        phx-target={@myself}
        phx-change="validate"
        phx-submit="save"
      >
        <.input field={@form[:name]} type="text" label="Name" />
        <.input field={@form[:email]} type="email" label="Email" />
        <.input :if={@action == :new} field={@form[:password]} type="password" label="Password" />
        <:actions>
          <.button phx-disable-with="Saving...">Save User</.button>
        </:actions>
      </.simple_form>
    </div>
    """
  end

  @impl true
  def update(%{user: user} = assigns, socket) do
    {:ok,
     socket
     |> assign(assigns)
     |> assign_new(:form, fn ->
       to_form(Accounts.change_user(user))
     end)}
  end

  @impl true
  def handle_event("validate", %{"user" => user_params}, socket) do
    changeset = Accounts.change_user(socket.assigns.user, user_params)
    {:noreply, assign(socket, form: to_form(changeset, action: :validate))}
  end

  def handle_event("save", %{"user" => user_params}, socket) do
    save_user(socket, socket.assigns.action, user_params)
  end

  defp save_user(socket, :edit, user_params) do
    case Accounts.update_user(socket.assigns.user, user_params) do
      {:ok, user} ->
        notify_parent({:saved, user})

        {:noreply,
         socket
         |> put_flash(:info, "User updated successfully")
         |> push_patch(to: socket.assigns.patch)}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, form: to_form(changeset))}
    end
  end

  defp save_user(socket, :new, user_params) do
    case Accounts.register_user(user_params) do
      {:ok, user} ->
        notify_parent({:saved, user})
        Phoenix.PubSub.broadcast(MyApp.PubSub, "users", {:user_created, user})

        {:noreply,
         socket
         |> put_flash(:info, "User created successfully")
         |> push_patch(to: socket.assigns.patch)}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, form: to_form(changeset))}
    end
  end

  defp notify_parent(msg), do: send(self(), {__MODULE__, msg})
end
```

---

## GenServer and OTP Patterns

### GenServer with State Management

```elixir
defmodule MyApp.Services.RateLimiter do
  use GenServer
  require Logger

  @cleanup_interval :timer.seconds(60)

  defstruct [:max_requests, :window_ms, :requests]

  ## Client API

  def start_link(opts) do
    name = Keyword.get(opts, :name, __MODULE__)
    GenServer.start_link(__MODULE__, opts, name: name)
  end

  def check_rate(server \\ __MODULE__, key, limit \\ nil) do
    GenServer.call(server, {:check_rate, key, limit})
  end

  def reset(server \\ __MODULE__, key) do
    GenServer.cast(server, {:reset, key})
  end

  ## Server Callbacks

  @impl true
  def init(opts) do
    max_requests = Keyword.get(opts, :max_requests, 100)
    window_ms = Keyword.get(opts, :window_ms, 60_000)

    schedule_cleanup()

    state = %__MODULE__{
      max_requests: max_requests,
      window_ms: window_ms,
      requests: %{}
    }

    {:ok, state}
  end

  @impl true
  def handle_call({:check_rate, key, limit}, _from, state) do
    limit = limit || state.max_requests
    now = System.monotonic_time(:millisecond)
    cutoff = now - state.window_ms

    # Get timestamps for this key, filtered by window
    timestamps =
      state.requests
      |> Map.get(key, [])
      |> Enum.filter(&(&1 > cutoff))

    count = length(timestamps)

    if count >= limit do
      {:reply, {:error, :rate_limit_exceeded, count}, state}
    else
      new_timestamps = [now | timestamps]
      new_requests = Map.put(state.requests, key, new_timestamps)
      new_state = %{state | requests: new_requests}
      {:reply, {:ok, count + 1}, new_state}
    end
  end

  @impl true
  def handle_cast({:reset, key}, state) do
    new_requests = Map.delete(state.requests, key)
    {:noreply, %{state | requests: new_requests}}
  end

  @impl true
  def handle_info(:cleanup, state) do
    now = System.monotonic_time(:millisecond)
    cutoff = now - state.window_ms

    new_requests =
      state.requests
      |> Enum.map(fn {key, timestamps} ->
        {key, Enum.filter(timestamps, &(&1 > cutoff))}
      end)
      |> Enum.reject(fn {_key, timestamps} -> timestamps == [] end)
      |> Map.new()

    Logger.debug("Cleaned up rate limiter. Keys before: #{map_size(state.requests)}, after: #{map_size(new_requests)}")

    schedule_cleanup()
    {:noreply, %{state | requests: new_requests}}
  end

  defp schedule_cleanup do
    Process.send_after(self(), :cleanup, @cleanup_interval)
  end
end
```

### Dynamic Supervisor

```elixir
defmodule MyApp.Workers.DynamicSupervisor do
  use DynamicSupervisor

  def start_link(init_arg) do
    DynamicSupervisor.start_link(__MODULE__, init_arg, name: __MODULE__)
  end

  @impl true
  def init(_init_arg) do
    DynamicSupervisor.init(strategy: :one_for_one)
  end

  def start_worker(user_id) do
    spec = {MyApp.Workers.UserWorker, user_id: user_id}
    DynamicSupervisor.start_child(__MODULE__, spec)
  end

  def stop_worker(pid) do
    DynamicSupervisor.terminate_child(__MODULE__, pid)
  end

  def list_workers do
    DynamicSupervisor.which_children(__MODULE__)
  end
end

defmodule MyApp.Workers.UserWorker do
  use GenServer, restart: :temporary

  def start_link(opts) do
    user_id = Keyword.fetch!(opts, :user_id)
    GenServer.start_link(__MODULE__, user_id, name: via_tuple(user_id))
  end

  defp via_tuple(user_id) do
    {:via, Registry, {MyApp.Registry, {__MODULE__, user_id}}}
  end

  @impl true
  def init(user_id) do
    {:ok, %{user_id: user_id, status: :idle}}
  end

  @impl true
  def handle_call(:get_status, _from, state) do
    {:reply, state.status, state}
  end

  @impl true
  def handle_cast({:update_status, new_status}, state) do
    {:noreply, %{state | status: new_status}}
  end
end
```

---

## Oban Background Jobs

```elixir
# lib/my_app/workers/email_worker.ex
defmodule MyApp.Workers.EmailWorker do
  use Oban.Worker, queue: :mailers, max_attempts: 3

  alias MyApp.Mailer
  alias MyApp.Emails

  @impl Oban.Worker
  def perform(%Oban.Job{args: %{"type" => "welcome", "user_id" => user_id}}) do
    user = MyApp.Accounts.get_user!(user_id)

    user
    |> Emails.welcome_email()
    |> Mailer.deliver()

    :ok
  end

  def perform(%Oban.Job{args: %{"type" => "password_reset", "user_id" => user_id, "token" => token}}) do
    user = MyApp.Accounts.get_user!(user_id)

    user
    |> Emails.password_reset_email(token)
    |> Mailer.deliver()

    :ok
  end

  # Schedule welcome email
  def schedule_welcome_email(user_id) do
    %{type: "welcome", user_id: user_id}
    |> new()
    |> Oban.insert()
  end

  # Schedule with delay
  def schedule_password_reset(user_id, token, delay_seconds \\ 0) do
    %{type: "password_reset", user_id: user_id, token: token}
    |> new(schedule_in: delay_seconds)
    |> Oban.insert()
  end
end
```

---

## ExUnit Testing

```elixir
# test/my_app/accounts_test.exs
defmodule MyApp.AccountsTest do
  use MyApp.DataCase

  alias MyApp.Accounts

  describe "register_user/1" do
    test "registers user with valid data" do
      valid_attrs = %{
        email: "test@example.com",
        name: "Test User",
        password: "SecurePass123!"
      }

      assert {:ok, user} = Accounts.register_user(valid_attrs)
      assert user.email == "test@example.com"
      assert user.name == "Test User"
      assert user.hashed_password != nil
      assert user.password == nil
    end

    test "returns error with invalid email" do
      invalid_attrs = %{
        email: "invalid-email",
        name: "Test User",
        password: "SecurePass123!"
      }

      assert {:error, changeset} = Accounts.register_user(invalid_attrs)
      assert "must have the @ sign and no spaces" in errors_on(changeset).email
    end

    test "returns error with short password" do
      invalid_attrs = %{
        email: "test@example.com",
        name: "Test User",
        password: "short"
      }

      assert {:error, changeset} = Accounts.register_user(invalid_attrs)
      assert "should be at least 12 character(s)" in errors_on(changeset).password
    end
  end

  describe "get_user_by_email_and_password/2" do
    setup do
      user = user_fixture()
      %{user: user}
    end

    test "returns user with correct password", %{user: user} do
      assert returned_user = Accounts.get_user_by_email_and_password(user.email, "Password123!")
      assert returned_user.id == user.id
    end

    test "returns nil with wrong password", %{user: user} do
      assert Accounts.get_user_by_email_and_password(user.email, "wrongpass") == nil
    end
  end
end
```

```elixir
# test/my_app_web/live/user_live_test.exs
defmodule MyAppWeb.UserLiveTest do
  use MyAppWeb.ConnCase

  import Phoenix.LiveViewTest
  import MyApp.AccountsFixtures

  describe "Index" do
    setup [:create_user]

    test "lists all users", %{conn: conn, user: user} do
      {:ok, _index_live, html} = live(conn, ~p"/users")

      assert html =~ "Listing Users"
      assert html =~ user.email
    end

    test "saves new user", %{conn: conn} do
      {:ok, index_live, _html} = live(conn, ~p"/users")

      assert index_live |> element("a", "New User") |> render_click() =~
               "New User"

      assert_patch(index_live, ~p"/users/new")

      assert index_live
             |> form("#user-form", user: %{email: "invalid"})
             |> render_change() =~ "must have the @ sign"

      assert index_live
             |> form("#user-form", user: valid_user_attributes())
             |> render_submit()

      assert_patch(index_live, ~p"/users")

      html = render(index_live)
      assert html =~ "User created successfully"
      assert html =~ "test@example.com"
    end

    test "deletes user in listing", %{conn: conn, user: user} do
      {:ok, index_live, _html} = live(conn, ~p"/users")

      assert index_live |> element("#users-#{user.id} a", "Delete") |> render_click()
      refute has_element?(index_live, "#users-#{user.id}")
    end
  end

  defp create_user(_) do
    user = user_fixture()
    %{user: user}
  end
end
```

---

## Production Dockerfile

```dockerfile
# Dockerfile
ARG ELIXIR_VERSION=1.17.3
ARG OTP_VERSION=27.1.2
ARG DEBIAN_VERSION=bookworm-20241016-slim

FROM hexpm/elixir:${ELIXIR_VERSION}-erlang-${OTP_VERSION}-debian-${DEBIAN_VERSION} AS builder

# Install build dependencies
RUN apt-get update -y && apt-get install -y build-essential git curl \
    && apt-get clean && rm -f /var/lib/apt/lists/*_*

# Prepare build directory
WORKDIR /app

# Install hex + rebar
RUN mix local.hex --force && \
    mix local.rebar --force

# Set build ENV
ENV MIX_ENV="prod"

# Install mix dependencies
COPY mix.exs mix.lock ./
RUN mix deps.get --only $MIX_ENV
RUN mkdir config

# Copy compile-time config files before compiling dependencies
COPY config/config.exs config/${MIX_ENV}.exs config/
RUN mix deps.compile

# Copy application code
COPY priv priv
COPY lib lib
COPY assets assets

# Compile assets
RUN mix assets.deploy

# Compile the release
RUN mix compile

# Create release
COPY config/runtime.exs config/
COPY rel rel
RUN mix release

# Start a new build stage for a smaller runtime image
FROM debian:${DEBIAN_VERSION} AS runtime

RUN apt-get update -y && \
  apt-get install -y libstdc++6 openssl libncurses5 locales ca-certificates \
  && apt-get clean && rm -f /var/lib/apt/lists/*_*

# Set the locale
RUN sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen && locale-gen

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

WORKDIR "/app"
RUN chown nobody /app

# Set runner ENV
ENV MIX_ENV="prod"

# Copy built release
COPY --from=builder --chown=nobody:root /app/_build/${MIX_ENV}/rel/my_app ./

USER nobody

# Expose port
EXPOSE 4000

CMD ["/app/bin/server"]
```

---

Last Updated: 2026-01-10
Version: 1.0.0
