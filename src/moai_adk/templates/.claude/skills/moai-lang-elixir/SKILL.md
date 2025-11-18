---
name: moai-lang-elixir
version: 4.0.0
updated: 2025-11-19
status: stable
category: Languages
description: Elixir 1.18+, Phoenix LiveView, OTP patterns, concurrency and distributed systems for building scalable, fault-tolerant applications
allowed-tools:
  - WebSearch
  - WebFetch
  - Read
  - Bash
related-skills:
  - moai-lang-erlang
  - moai-domain-backend
  - moai-essentials-perf
tags:
  - functional-programming
  - concurrency
  - distributed-systems
  - phoenix
  - otp
---

# Elixir: Functional Programming for Scalable Systems

## Overview

Elixir is a dynamic, functional programming language built on the Erlang Virtual Machine (BEAM), combining the power of functional programming with practical, pragmatic design. Born from José Valim's work at Plataformatec, Elixir brings together:

**Core Philosophy**: Elixir emphasizes **immutability**, **concurrency**, **distributed systems**, and **fault tolerance**. Unlike imperative languages, Elixir encourages you to think in terms of data transformation pipelines and message-passing actors rather than shared mutable state.

**Unique Advantages**:
- **Lightweight Concurrency**: Run hundreds of thousands of processes on a single machine without threads
- **Fault Tolerance**: "Let it crash" philosophy with automatic process supervision and recovery
- **Distributed Systems**: Built-in clustering and node communication for horizontal scaling
- **Hot Code Upgrades**: Update code without stopping the system (unique to Erlang/Elixir)
- **Real-time Capabilities**: LiveView framework for reactive web interfaces without JavaScript

**When to Use Elixir**:
- Real-time web applications (chat, notifications, dashboards)
- Microservices and distributed systems
- Data streaming and processing pipelines
- IoT and embedded systems with high concurrency
- Messaging systems and event-driven architectures
- Applications requiring 99.9999999% uptime ("nine nines")

**Not Ideal For**:
- CPU-intensive numerical computing (use Rust, C++)
- Traditional CRUD applications with simple requirements
- Desktop GUI applications

---

## 1. Elixir Fundamentals

### Pattern Matching: The Foundation of Elixir

Pattern matching is **not just destructuring**—it's a control flow mechanism that makes Elixir uniquely expressive. Every assignment, function call, and case statement uses pattern matching.

```elixir
# Basic pattern matching
{name, age} = {"Alice", 30}
# name = "Alice", age = 30

# List patterns
[head | tail] = [1, 2, 3, 4]
# head = 1, tail = [2, 3, 4]

# Ignoring values with underscore
{x, _, z} = {"a", "b", "c"}
# x = "a", z = "c"

# Nested patterns
user = %{"name" => "Bob", "profile" => %{"age" => 25, "city" => "NYC"}}
%{"name" => name, "profile" => %{"city" => city}} = user
# name = "Bob", city = "NYC"

# Guard clauses add conditions
def process({type, value}) when type in [:success, :ok] do
  {:ok, value}
end

def process({:error, reason}) do
  {:error, reason}
end

# Multiple clauses create elegant branching
def greet(%{"name" => name, "vip" => true}), do: "Welcome back, VIP #{name}!"
def greet(%{"name" => name}), do: "Hello, #{name}!"
```

**Why it matters**: Pattern matching eliminates null-checking boilerplate. Instead of `if something != nil and something.valid?`, you write elegant pattern matches that fail fast and communicate intent clearly.

### The Pipe Operator: Data Transformation Chains

The pipe operator `|>` threads the result of one function as the first argument to the next, creating readable data transformation chains.

```elixir
# Traditional nested approach (hard to read)
String.upcase(String.trim(input))

# Pipe operator (clear data flow)
input
|> String.trim()
|> String.upcase()

# Complex pipeline
users
|> Enum.filter(fn user -> user.age > 18 end)
|> Enum.map(fn user -> user.name end)
|> Enum.sort()
|> Enum.join(", ")

# Real-world example: Data processing pipeline
defmodule DataProcessor do
  def process_csv(file_path) do
    file_path
    |> File.read!()
    |> String.split("\n")
    |> Enum.map(&String.trim/1)
    |> Enum.filter(&(&1 != ""))
    |> Enum.map(&parse_line/1)
    |> Enum.filter(&valid?/1)
    |> Enum.reduce(%{}, &aggregate/2)
  end

  defp parse_line(line), do: String.split(line, ",")
  defp valid?(fields), do: length(fields) == 3
  defp aggregate(fields, acc), do: Map.update(acc, Enum.at(fields, 0), 1, &(&1 + 1))
end
```

**Advantage**: Pipes make data transformations explicit and readable, replacing nested function calls with a clear left-to-right flow.

### Modules and Functions: Organizing Code

Elixir organizes code into modules containing named functions. Unlike some functional languages, Elixir allows multiple function clauses with different arities.

```elixir
defmodule Math do
  # Multiple clauses: same name, different arity
  def add(a, b), do: a + b
  def add(a, b, c), do: a + b + c

  # Guards for type/condition checking
  def factorial(0), do: 1
  def factorial(n) when n > 0 do
    n * factorial(n - 1)
  end

  # Private function (only accessible within module)
  defp helper_function(x) do
    x * 2
  end

  # Default arguments
  def greet(name, greeting \\ "Hello") do
    "#{greeting}, #{name}!"
  end
end

# Calling functions
Math.add(2, 3)           # 5
Math.add(1, 2, 3)        # 6
Math.greet("Alice")      # "Hello, Alice!"
Math.greet("Bob", "Hi")  # "Hi, Bob!"
```

**Key Pattern**: Elixir encourages many small, focused functions over few large ones. Each clause handles a specific case, making code intention-revealing.

### Enum and Stream: Collection Processing

`Enum` provides eager evaluation (immediate execution), while `Stream` provides lazy evaluation (deferred execution until needed).

```elixir
# Enum: eager evaluation (processes immediately)
numbers = 1..1000000 |> Enum.to_list()
evens = Enum.filter(numbers, &(rem(&1, 2) == 0))
# Creates intermediate list in memory

# Stream: lazy evaluation (processes on-demand)
stream = 1..1000000
  |> Stream.filter(&(rem(&1, 2) == 0))
  |> Stream.map(&(&1 * 2))
  |> Stream.take(10)
  |> Enum.to_list()
# Processes only 10 elements

# Real-world comparison
defmodule FileProcessor do
  # Eager: loads entire file into memory
  def count_lines_eager(file_path) do
    file_path
    |> File.read!()
    |> String.split("\n")
    |> length()
  end

  # Lazy: streams line-by-line
  def count_lines_lazy(file_path) do
    file_path
    |> File.stream!()
    |> Stream.map(&String.trim/1)
    |> Enum.count()
  end
end

# Reduce for custom aggregation
colors = ["red", "green", "blue"]
result = Enum.reduce(colors, %{}, fn color, acc ->
  Map.put(acc, color, String.length(color))
end)
# %{"red" => 3, "green" => 5, "blue" => 4}
```

**Decision Rule**: Use `Enum` for small collections or final operations, use `Stream` for large data processing or infinite sequences.

---

## 2. Phoenix LiveView: Real-Time Web

### Understanding LiveView: WebSocket Magic

Phoenix LiveView eliminates the traditional divide between frontend and backend. Instead of writing JavaScript for real-time updates, you write Elixir code that runs on the server, with BEAM-managed WebSocket connections pushing updates to clients.

```elixir
# lib/my_app_web/live/counter_live.ex
defmodule MyAppWeb.CounterLive do
  use MyAppWeb, :live_view

  # Mount: called when client connects
  def mount(_params, _session, socket) do
    socket = assign(socket, count: 0, timestamp: DateTime.utc_now())
    {:ok, socket}
  end

  # Render: called whenever state changes
  def render(assigns) do
    ~H"""
    <div class="counter">
      <p>Count: <%= @count %></p>
      <p>Last updated: <%= @timestamp %></p>
      <button phx-click="increment">Increment</button>
      <button phx-click="decrement">Decrement</button>
      <button phx-click="reset">Reset</button>
    </div>
    """
  end

  # Event handlers: respond to client actions
  def handle_event("increment", _params, socket) do
    {:noreply, assign(socket, count: socket.assigns.count + 1, timestamp: DateTime.utc_now())}
  end

  def handle_event("decrement", _params, socket) do
    {:noreply, assign(socket, count: socket.assigns.count - 1, timestamp: DateTime.utc_now())}
  end

  def handle_event("reset", _params, socket) do
    {:noreply, assign(socket, count: 0, timestamp: DateTime.utc_now())}
  end
end
```

**How it works**: LiveView renders HTML on the server. When a user clicks a button, the browser sends an event via WebSocket. The server calls `handle_event`, updates state with `assign`, and `render` is called automatically. Only the changed HTML is sent back, updating the DOM.

### LiveView Components: Reusable UI Building Blocks

Components are stateful or stateless sub-views that encapsulate UI logic.

```elixir
# Stateless component (pure function)
defmodule MyAppWeb.Components.Card do
  use MyAppWeb, :component

  def card(assigns) do
    ~H"""
    <div class="card">
      <div class="card-header"><%= @title %></div>
      <div class="card-body"><%= render_slot(@inner_block) %></div>
    </div>
    """
  end
end

# Using stateless component
def render(assigns) do
  ~H"""
  <.card title="User Profile">
    <p>Name: <%= @user.name %></p>
    <p>Email: <%= @user.email %></p>
  </.card>
  """
end

# Stateful component (with handle_event)
defmodule MyAppWeb.Components.Modal do
  use MyAppWeb, :live_component

  def mount(socket) do
    {:ok, assign(socket, open: false)}
  end

  def render(assigns) do
    ~H"""
    <div id={"modal-#{@id}"} class={"modal #{if @open, do: 'open'}"}>
      <button phx-click="close" phx-target={@myself}>Close</button>
      <%= render_slot(@inner_block) %>
    </div>
    """
  end

  def handle_event("close", _params, socket) do
    {:noreply, assign(socket, open: false)}
  end
end

# Using stateful component
def render(assigns) do
  ~H"""
  <.live_component module={Modal} id="help-modal">
    <p>This is help content</p>
  </.live_component>
  """
end
```

**Architecture**: Stateless components for presentation, stateful components for interactive features. Components receive `assigns` (data) and return rendered HTML.

### State Management: Socket Assigns

All state in a LiveView is stored in `socket.assigns`, a map that triggers re-renders when changed.

```elixir
defmodule MyAppWeb.TodoLive do
  use MyAppWeb, :live_view

  def mount(_params, _session, socket) do
    {:ok, assign(socket, todos: [], input: "", filter: "all")}
  end

  def handle_event("add_todo", %{"value" => text}, socket) do
    new_todo = %{id: Enum.random(1..999999), text: text, done: false}
    todos = socket.assigns.todos ++ [new_todo]
    {:noreply, assign(socket, todos: todos, input: "")}
  end

  def handle_event("toggle_todo", %{"id" => id}, socket) do
    todos = Enum.map(socket.assigns.todos, fn todo ->
      if to_string(todo.id) == id, do: %{todo | done: !todo.done}, else: todo
    end)
    {:noreply, assign(socket, todos: todos)}
  end

  def handle_event("set_filter", %{"filter" => filter}, socket) do
    {:noreply, assign(socket, filter: filter)}
  end

  defp filtered_todos(todos, filter) do
    case filter do
      "active" -> Enum.filter(todos, &(!&1.done))
      "completed" -> Enum.filter(todos, &(&1.done))
      _ -> todos
    end
  end

  def render(assigns) do
    ~H"""
    <div class="todo-app">
      <input phx-change="update_input" value={@input} placeholder="Add todo...">
      <button phx-click="add_todo" phx-value-value={@input}>Add</button>

      <div class="filters">
        <button phx-click="set_filter" phx-value-filter="all">All</button>
        <button phx-click="set_filter" phx-value-filter="active">Active</button>
        <button phx-click="set_filter" phx-value-filter="completed">Completed</button>
      </div>

      <ul>
        <%= for todo <- filtered_todos(@todos, @filter) do %>
          <li>
            <input type="checkbox" checked={todo.done} phx-click="toggle_todo" phx-value-id={todo.id}>
            <%= todo.text %>
          </li>
        <% end %>
      </ul>
    </div>
    """
  end
end
```

**Key Insight**: LiveView's state model is simple: server maintains state in `assigns`, client sends events, state updates trigger automatic re-renders.

### Performance: Handling Scale

LiveView efficiently handles thousands of concurrent connections through BEAM's lightweight processes.

```elixir
defmodule MyAppWeb.DashboardLive do
  use MyAppWeb, :live_view

  # Subscribe to updates (using Phoenix.PubSub)
  def mount(_params, _session, socket) do
    Phoenix.PubSub.subscribe(MyApp.PubSub, "dashboard:updates")
    {:ok, assign(socket, metrics: %{}, page: 1, limit: 10)}
  end

  # Handle real-time updates from background processes
  def handle_info({:metrics_updated, new_metrics}, socket) do
    {:noreply, assign(socket, metrics: new_metrics)}
  end

  def handle_info({:user_joined, user}, socket) do
    {:noreply, assign(socket, user_count: socket.assigns.user_count + 1)}
  end

  # Pagination to avoid rendering large lists
  def handle_event("next_page", _params, socket) do
    page = socket.assigns.page + 1
    {:noreply, assign(socket, page: page)}
  end

  # Debounce expensive operations
  def handle_event("search", %{"query" => query}, socket) do
    # Schedule a delayed search
    Process.send_after(self(), {:perform_search, query}, 500)
    {:noreply, assign(socket, search_query: query)}
  end

  def handle_info({:perform_search, query}, socket) do
    results = MyApp.search(query)
    {:noreply, assign(socket, search_results: results)}
  end
end
```

**Optimization Tactics**: PubSub for one-to-many broadcasts, pagination for large lists, debouncing for expensive operations, and lazy loading for images.

### Testing LiveView

ExUnit supports testing LiveView with `.render_hook` and `.trigger_event`.

```elixir
defmodule MyAppWeb.CounterLiveTest do
  use MyAppWeb.ConnCase
  import Phoenix.LiveViewTest

  test "increments counter on button click" do
    {:ok, view, html} = live(conn, "/counter")
    
    assert html =~ "Count: 0"
    
    html = view |> element("button", "Increment") |> render_click()
    assert html =~ "Count: 1"
    
    html = view |> element("button", "Increment") |> render_click()
    assert html =~ "Count: 2"
  end

  test "resets counter" do
    {:ok, view, _html} = live(conn, "/counter")
    
    # Click increment multiple times
    for _ <- 1..5 do
      render_click(view, "increment")
    end
    
    # Reset
    html = render_click(view, "reset")
    assert html =~ "Count: 0"
  end
end
```

---

## 3. OTP: Open Telecom Platform

### GenServer: Managing State and Concurrency

GenServer is a behavior that encapsulates server logic, handling synchronous and asynchronous messages from multiple clients.

```elixir
# lib/my_app/counter.ex
defmodule MyApp.Counter do
  use GenServer
  require Logger

  def start_link(initial_value \\ 0) do
    GenServer.start_link(__MODULE__, initial_value, name: __MODULE__)
  end

  # Client API
  def get() do
    GenServer.call(__MODULE__, :get)
  end

  def increment(amount \\ 1) do
    GenServer.call(__MODULE__, {:increment, amount})
  end

  def decrement(amount \\ 1) do
    GenServer.call(__MODULE__, {:decrement, amount})
  end

  def reset() do
    GenServer.cast(__MODULE__, :reset)  # Async (cast)
  end

  # Server callbacks
  @impl true
  def init(initial_value) do
    Logger.info("Counter initialized with value: #{initial_value}")
    {:ok, initial_value}
  end

  @impl true
  def handle_call(:get, _from, state) do
    {:reply, state, state}
  end

  @impl true
  def handle_call({:increment, amount}, _from, state) do
    new_state = state + amount
    {:reply, new_state, new_state}
  end

  @impl true
  def handle_call({:decrement, amount}, _from, state) do
    new_state = state - amount
    {:reply, new_state, new_state}
  end

  @impl true
  def handle_cast(:reset, _state) do
    {:noreply, 0}
  end

  # Handle system messages
  @impl true
  def terminate(reason, state) do
    Logger.info("Counter terminating with state #{state}, reason: #{inspect(reason)}")
    :ok
  end
end

# Usage
{:ok, _pid} = MyApp.Counter.start_link(0)
MyApp.Counter.increment(5)    # Returns 5
MyApp.Counter.increment(3)    # Returns 8
MyApp.Counter.get()           # Returns 8
MyApp.Counter.reset()         # Async, returns :ok
```

**Call vs Cast**: `call` is synchronous (waits for reply), `cast` is asynchronous (fire-and-forget). Use `call` when you need a response, `cast` for notifications.

### Supervisor: Fault Tolerance and Process Management

Supervisors monitor child processes, automatically restarting them if they crash.

```elixir
# lib/my_app/supervisor.ex
defmodule MyApp.Supervisor do
  use Supervisor

  def start_link(arg) do
    Supervisor.start_link(__MODULE__, arg, name: __MODULE__)
  end

  @impl true
  def init(_arg) do
    children = [
      # Restart strategy, max_restarts, max_seconds
      # :permanent - always restarted
      # :temporary - never restarted
      # :transient - restarted only if abnormal exit
      {MyApp.Counter, 0},
      {MyApp.Database, []},
      {MyApp.Cache, [ttl: 3600]},
      # Worker pool
      {DynamicSupervisor, strategy: :one_for_one, name: MyApp.DynamicWorkers}
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end
end

# Supervision strategies
# :one_for_one - restart only the failed child
# :rest_for_one - restart failed child and all started after it
# :one_for_all - restart all children if one fails
# :simple_one_for_one - deprecated, use DynamicSupervisor instead

defmodule MyApp.DynamicWorkerSupervisor do
  use DynamicSupervisor

  def start_link(_arg) do
    DynamicSupervisor.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  @impl true
  def init(:ok) do
    DynamicSupervisor.init(strategy: :one_for_one, max_restarts: 3, max_seconds: 5)
  end

  def start_worker(args) do
    DynamicSupervisor.start_child(__MODULE__, {MyApp.Worker, args})
  end
end

# Supervisor tree for a real application
defmodule MyApp.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      MyApp.Repo,                          # Database
      {Phoenix.PubSub, name: MyApp.PubSub},
      MyAppWeb.Endpoint,                   # Phoenix web server
      MyApp.Supervisor,                    # Application supervisor
      {Registry, keys: :unique, name: MyApp.ProcessRegistry}
    ]

    Supervisor.start_link(children, strategy: :one_for_one)
  end

  def config_change(changed, _new, removed) do
    MyAppWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end
```

**Recovery Pattern**: The "let it crash" philosophy means you don't write defensive code; instead, you let processes crash and let supervisors restart them. This is often cleaner than error handling.

### Error Handling: Let It Crash Philosophy

Elixir's error handling philosophy differs from traditional try-catch approaches.

```elixir
# Traditional approach (not idiomatic in Elixir)
try do
  risky_operation()
rescue
  e in RuntimeError -> handle_error(e)
  e in SomeError -> handle_other_error(e)
after
  cleanup()
end

# Idiomatic Elixir: pattern matching
case risky_operation() do
  {:ok, result} -> process_result(result)
  {:error, reason} -> handle_error(reason)
end

# With! variants for crash-on-error
result = risky_operation!()  # Crashes if error tuple returned

# Real-world: Database operation
def get_user(user_id) do
  case Repo.get(User, user_id) do
    %User{} = user -> {:ok, user}
    nil -> {:error, :not_found}
  end
end

def fetch_user!(user_id) do
  case get_user(user_id) do
    {:ok, user} -> user
    {:error, reason} -> raise "User not found: #{inspect(reason)}"
  end
end

# Distributed system resilience
defmodule MyApp.ResilientServer do
  def call_remote_service(url) do
    case HTTPClient.get(url) do
      {:ok, response} -> {:ok, response}
      {:error, :timeout} -> {:error, :timeout}  # Retry externally
      {:error, :connection_refused} -> {:error, :unavailable}
    end
  end
end
```

**Key Pattern**: Return tuples `{:ok, value}` or `{:error, reason}` instead of raising exceptions. Let supervisors handle crashes.

### Distributed Systems: Clustering and Node Communication

Elixir runs on the Erlang VM, which has first-class support for distributed computing.

```elixir
# lib/my_app/cluster.ex
defmodule MyApp.Cluster do
  # Start node with clustering enabled
  # iex --sname node1@localhost -S mix phx.server
  # iex --sname node2@localhost -S mix phx.server

  def connect_node(node_name) do
    Node.connect(String.to_atom(node_name))
  end

  def get_all_nodes() do
    [Node.self() | Node.list()]
  end

  def call_remote(node, module, function, args) do
    :rpc.call(node, module, function, args)
  end

  def cast_remote(node, module, function, args) do
    :rpc.cast(node, module, function, args)
  end
end

# Global registry across cluster
defmodule MyApp.ProcessRegistry do
  def start_link(_) do
    Registry.start_link(keys: :unique, name: __MODULE__)
  end

  def register(name, value) do
    Registry.register(__MODULE__, name, value)
  end

  def lookup(name) do
    case Registry.lookup(__MODULE__, name) do
      [{pid, _value}] -> {:ok, pid}
      [] -> {:error, :not_found}
    end
  end

  def broadcast(topic, message) do
    Registry.dispatch(__MODULE__, topic, fn entries ->
      for {pid, _} <- entries, do: send(pid, message)
    end)
  end
end

# Distributed Supervisor
defmodule MyApp.DistributedSupervisor do
  use Supervisor

  def start_link(_) do
    Supervisor.start_link(__MODULE__, :ok, name: {:global, __MODULE__})
  end

  @impl true
  def init(:ok) do
    children = [
      {MyApp.Counter, 0}
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end

  # Access from any node
  def get_counter() do
    GenServer.call({:global, MyApp.Counter}, :get)
  end
end

# Cluster deployment example
# config/runtime.exs
config :libcluster,
  topologies: [
    kubernetes: [
      strategy: Cluster.Strategy.Kubernetes.DNS,
      config: [
        service: "my-app",
        namespace: "default",
        polling_interval: 10_000
      ]
    ]
  ]
```

**Use Cases**: Distributed caching, session replication, load balancing across nodes, and fault tolerance across multiple machines.

---

## 4. Concurrency and Parallelism

### Process Model: Lightweight Concurrency

Elixir processes are extremely lightweight (thousands per MB), enabling massive concurrency without threads.

```elixir
# Spawning processes
defmodule Concurrency do
  def simple_spawn() do
    pid = spawn(fn -> IO.puts("Hello from process") end)
    Process.alive?(pid)  # true
  end

  def spawn_with_link() do
    spawn_link(fn ->
      raise "Crash this process"
    end)
    # Parent also crashes (linked)
  end

  def spawn_monitor() do
    {pid, ref} = spawn_monitor(fn ->
      :timer.sleep(1000)
      IO.puts("Done")
    end)
    
    receive do
      {:DOWN, ^ref, :process, ^pid, reason} ->
        IO.puts("Process exited: #{inspect(reason)}")
    end
  end

  # Broadcasting to multiple processes
  def broadcast_to_processes(pids, message) do
    Enum.each(pids, fn pid -> send(pid, message) end)
  end

  # Worker pool pattern
  def worker_pool_example() do
    workers = Enum.map(1..4, fn _i ->
      spawn_link(&worker/0)
    end)

    # Distribute work
    Enum.each(1..10, fn job ->
      worker = Enum.random(workers)
      send(worker, {:work, job})
    end)
  end

  defp worker() do
    receive do
      {:work, job} ->
        IO.puts("Processing job: #{job}")
        worker()
    end
  end
end

# Process messaging
defmodule Messenger do
  def start() do
    pid = self()
    spawn(fn ->
      send(pid, {:greeting, "Hello from child"})
    end)
    
    receive do
      {:greeting, msg} -> IO.puts(msg)
    after
      1000 -> IO.puts("Timeout")
    end
  end

  def request_reply(server_pid) do
    send(server_pid, {:request, self(), "data"})
    
    receive do
      {:response, data} -> data
    end
  end
end
```

**Key Insight**: Processes are isolated, communicate only through messages, and are supervised. This isolation makes fault tolerance and parallelism natural.

### Actor Model: Message Passing Patterns

The actor model treats processes as actors that receive messages and respond independently.

```elixir
defmodule Counter.Actor do
  def start() do
    spawn(fn -> actor_loop(0) end)
  end

  defp actor_loop(count) do
    receive do
      {:inc, sender} ->
        new_count = count + 1
        send(sender, {:count, new_count})
        actor_loop(new_count)

      {:get, sender} ->
        send(sender, {:count, count})
        actor_loop(count)

      :stop ->
        :ok
    end
  end
end

# Using the actor
defmodule Counter.ActorTest do
  def test() do
    actor = Counter.Actor.start()
    
    send(actor, {:inc, self()})
    receive do
      {:count, n} -> IO.puts("Count: #{n}")
    end
    
    send(actor, {:get, self()})
    receive do
      {:count, n} -> IO.puts("Final count: #{n}")
    end
    
    send(actor, :stop)
  end
end

# Pub/Sub actor pattern
defmodule PubSub.Actor do
  def start() do
    spawn(fn -> pub_sub_loop(%{}) end)
  end

  defp pub_sub_loop(subscribers) do
    receive do
      {:subscribe, topic, pid} ->
        subscribers = Map.update(subscribers, topic, [pid], &[pid | &1])
        pub_sub_loop(subscribers)

      {:publish, topic, message} ->
        subs = Map.get(subscribers, topic, [])
        Enum.each(subs, fn pid -> send(pid, message) end)
        pub_sub_loop(subscribers)

      {:unsubscribe, topic, pid} ->
        subscribers = Map.update(subscribers, topic, [], &List.delete(&1, pid))
        pub_sub_loop(subscribers)
    end
  end
end
```

**Pattern**: Each actor owns its state, responds to messages asynchronously, and maintains its own loop.

### Task: Asynchronous Operations

Tasks are spawned processes designed for one-off async work or parallel operations.

```elixir
defmodule TaskExample do
  def async_operation() do
    task = Task.async(fn ->
      :timer.sleep(1000)
      "Result from async operation"
    end)
    
    # Do other work
    IO.puts("Doing other work...")
    
    # Wait for result (with timeout)
    result = Task.await(task, 5000)
    IO.puts(result)
  end

  def parallel_requests(urls) do
    urls
    |> Enum.map(&Task.async(fn -> fetch_url(&1) end))
    |> Enum.map(&Task.await(&1, 10000))
  end

  def map_reduce_parallel(list, map_fn, reduce_fn, initial_acc) do
    list
    |> Enum.map(&Task.async(fn -> map_fn.(&1) end))
    |> Enum.map(&Task.await/1)
    |> Enum.reduce(initial_acc, reduce_fn)
  end

  defp fetch_url(url) do
    # Simulate HTTP request
    {:ok, "Content from #{url}"}
  end

  # Supervised tasks
  def supervised_async(work_fn) do
    {:ok, task_pid} = Task.Supervisor.start_child(
      MyApp.TaskSupervisor,
      work_fn
    )
    task_pid
  end
end
```

**Use Cases**: Parallel API calls, background jobs, image processing, and anything that can run independently.

### Flow: Data Pipeline Parallelization

Flow is a higher-level abstraction for processing data in parallel.

```elixir
defmodule DataProcessing do
  # Sequential pipeline
  def sequential_pipeline(data) do
    data
    |> Enum.filter(&valid?/1)
    |> Enum.map(&transform/1)
    |> Enum.reduce(%{}, &aggregate/2)
  end

  # Parallel pipeline
  def parallel_pipeline(data) do
    data
    |> Flow.from_enumerable()
    |> Flow.filter(&valid?/1)
    |> Flow.map(&transform/1)
    |> Flow.reduce(fn -> %{} end, &aggregate/2)
    |> Enum.to_list()
  end

  # Parallel with partitioning
  def parallel_partitioned_pipeline(data) do
    data
    |> Flow.from_enumerable(max_demand: 100)
    |> Flow.partition(key: fn item -> item.category end)
    |> Flow.reduce(fn -> %{} end, fn item, acc ->
      Map.update(acc, item.category, [item], &[item | &1])
    end)
    |> Enum.to_list()
  end

  # Real-world: processing CSV files in parallel
  def process_large_csv(file_path, batch_size \\ 1000) do
    file_path
    |> File.stream!()
    |> Flow.from_enumerable(max_demand: batch_size)
    |> Flow.map(&parse_csv_line/1)
    |> Flow.filter(&valid_row?/1)
    |> Flow.map(&enrich_row/1)
    |> Flow.partition(key: fn row -> row.region end)
    |> Flow.reduce(fn -> %{} end, &aggregate_by_region/2)
    |> Enum.to_list()
  end

  defp valid?(item), do: item != nil
  defp transform(item), do: String.upcase(item)
  defp aggregate(item, acc), do: Map.update(acc, item, 1, &(&1 + 1))
  
  defp parse_csv_line(line), do: String.split(line, ",")
  defp valid_row?(row), do: length(row) == 5
  defp enrich_row(row), do: %{data: row, processed_at: DateTime.utc_now()}
  defp aggregate_by_region(row, acc), do: Map.update(acc, row.region, [row], &[row | &1])
end

# Performance comparison
defmodule FlowBenchmark do
  def benchmark() do
    data = 1..100000 |> Enum.to_list()

    {time_seq, _} = :timer.tc(fn ->
      DataProcessing.sequential_pipeline(data)
    end)

    {time_par, _} = :timer.tc(fn ->
      DataProcessing.parallel_pipeline(data)
    end)

    IO.puts("Sequential: #{time_seq} microseconds")
    IO.puts("Parallel: #{time_par} microseconds")
    IO.puts("Speedup: #{Float.round(time_seq / time_par, 2)}x")
  end
end
```

**When to use Flow**: Processing large datasets, CPU-bound operations parallelizable by partition key, and ETL pipelines.

---

## 5. Ecto Database Layer

### Schema Definition and Migrations

Ecto provides a declarative approach to database schemas with type safety.

```elixir
# lib/my_app/schemas/user.ex
defmodule MyApp.User do
  use Ecto.Schema
  import Ecto.Changeset

  schema "users" do
    field :name, :string
    field :email, :string
    field :age, :integer
    field :bio, :string
    field :is_active, :boolean, default: true
    field :inserted_at, :naive_datetime_usec, autogenerate: {DateTime, :utc_now, []}

    has_many :posts, MyApp.Post
    has_many :comments, through: [:posts, :comments]
    belongs_to :organization, MyApp.Organization

    timestamps(type: :utc_datetime_usec)
  end

  # Validation changeset
  def changeset(user, attrs) do
    user
    |> cast(attrs, [:name, :email, :age, :bio, :is_active])
    |> validate_required([:name, :email])
    |> validate_format(:email, ~r/@/)
    |> validate_number(:age, greater_than: 0, less_than: 150)
    |> validate_length(:bio, max: 500)
    |> unique_constraint(:email, name: "unique_email_index")
  end

  def registration_changeset(user, attrs) do
    changeset(user, attrs)
    |> validate_required([:password])
    |> put_password_hash()
  end

  defp put_password_hash(changeset) do
    case changeset do
      %Ecto.Changeset{valid?: true, changes: %{password: password}} ->
        put_change(changeset, :password_hash, hash_password(password))

      changeset ->
        changeset
    end
  end

  defp hash_password(password) do
    Bcrypt.hash_pwd_salt(password)
  end
end

# lib/my_app/schemas/post.ex
defmodule MyApp.Post do
  use Ecto.Schema
  import Ecto.Changeset

  schema "posts" do
    field :title, :string
    field :content, :string
    field :view_count, :integer, default: 0
    field :status, :string, default: "draft"

    belongs_to :user, MyApp.User
    has_many :comments, MyApp.Comment, on_delete: :delete_all

    timestamps()
  end

  def changeset(post, attrs) do
    post
    |> cast(attrs, [:title, :content, :status, :user_id])
    |> validate_required([:title, :content])
    |> validate_length(:title, min: 3, max: 200)
    |> foreign_key_constraint(:user_id)
  end
end

# Migration file: priv/repo/migrations/20240101120000_create_users.exs
defmodule MyApp.Repo.Migrations.CreateUsers do
  use Ecto.Migration

  def change do
    create table(:users) do
      add :name, :string, null: false
      add :email, :string, null: false
      add :age, :integer
      add :bio, :text
      add :is_active, :boolean, default: true
      add :organization_id, references(:organizations, on_delete: :nilify_all)

      timestamps(type: :utc_datetime_usec)
    end

    create unique_index(:users, [:email], name: "unique_email_index")
    create index(:users, [:organization_id])
  end
end

# Migration: priv/repo/migrations/20240102120000_create_posts.exs
defmodule MyApp.Repo.Migrations.CreatePosts do
  use Ecto.Migration

  def change do
    create table(:posts) do
      add :title, :string, null: false
      add :content, :text, null: false
      add :status, :string, default: "draft"
      add :view_count, :integer, default: 0
      add :user_id, references(:users, on_delete: :delete_all), null: false

      timestamps()
    end

    create index(:posts, [:user_id])
    create index(:posts, [:status])
  end
end
```

**Key Concepts**: Schema defines structure, changeset validates data, migrations define database evolution.

### Querying with Ecto

Ecto provides a composable query DSL that's translated to SQL.

```elixir
defmodule MyApp.UserQueries do
  import Ecto.Query
  alias MyApp.{User, Post, Repo}

  # Basic queries
  def get_user(id) do
    Repo.get(User, id)
  end

  def get_user_by_email(email) do
    Repo.get_by(User, email: email)
  end

  # Query composition
  def active_users() do
    from u in User,
      where: u.is_active == true,
      select: u
  end

  def users_with_posts() do
    from u in User,
      join: p in assoc(u, :posts),
      distinct: u.id,
      select: u
  end

  def recent_posts_by_user(user_id, limit \\ 10) do
    from p in Post,
      where: p.user_id == ^user_id,
      order_by: [desc: p.inserted_at],
      limit: ^limit,
      select: p
  end

  # Aggregation
  def user_post_count(user_id) do
    from p in Post,
      where: p.user_id == ^user_id,
      select: count(p.id)
  end

  def posts_by_status(status) do
    from p in Post,
      where: p.status == ^status,
      group_by: p.user_id,
      select: %{user_id: p.user_id, count: count(p.id)}
  end

  # Filtering and searching
  def search_users(query) do
    from u in User,
      where: like(u.name, ^"%#{query}%") or like(u.email, ^"%#{query}%"),
      select: u
  end

  # Pagination
  def paginate(queryable, page, page_size) do
    offset = (page - 1) * page_size
    from q in queryable,
      offset: ^offset,
      limit: ^page_size
  end

  # Preloading associations
  def users_with_posts_preloaded() do
    from u in User,
      preload: :posts
  end

  def users_with_nested_preload() do
    from u in User,
      preload: [posts: :comments]
  end

  # Execute queries
  def all_active_users() do
    active_users()
    |> Repo.all()
  end

  def count_all_users() do
    from u in User, select: count(u.id)
    |> Repo.one()
  end

  def users_and_count() do
    query = active_users()
    total = query |> Repo.aggregate(:count, :id)
    users = Repo.all(query)
    {users, total}
  end
end

# In LiveView or controller
def render_users(socket) do
  users = MyApp.UserQueries.active_users()
    |> MyApp.UserQueries.paginate(socket.assigns.page, 10)
    |> MyApp.Repo.all()
    |> MyApp.Repo.preload(:posts)

  assign(socket, users: users)
end
```

**Performance**: Preload associations to avoid N+1 queries, use `select` to fetch only needed columns, and paginate large result sets.

### Transactions and Consistency

Ecto transactions ensure ACID properties for multi-step operations.

```elixir
defmodule MyApp.TransactionExample do
  alias MyApp.Repo
  import Ecto.Changeset

  # Simple transaction
  def create_user_with_organization(user_attrs, org_attrs) do
    Repo.transaction(fn ->
      with {:ok, org} <- Repo.insert(MyApp.Organization.changeset(%MyApp.Organization{}, org_attrs)),
           user_attrs <- Map.put(user_attrs, :organization_id, org.id),
           {:ok, user} <- Repo.insert(MyApp.User.changeset(%MyApp.User{}, user_attrs)) do
        {org, user}
      else
        {:error, reason} -> Repo.rollback(reason)
      end
    end)
  end

  # Atomic operations
  def transfer_credits(from_user_id, to_user_id, amount) do
    Repo.transaction(fn ->
      from_user = Repo.get_for_update!(MyApp.User, from_user_id)
      to_user = Repo.get_for_update!(MyApp.User, to_user_id)

      from_balance = from_user.balance - amount
      to_balance = to_user.balance + amount

      if from_balance < 0 do
        Repo.rollback(:insufficient_funds)
      else
        {:ok, _} = Repo.update(change(from_user, balance: from_balance))
        {:ok, _} = Repo.update(change(to_user, balance: to_balance))
        :ok
      end
    end)
  end

  # Handling transaction results
  def create_user_safe(attrs) do
    case Repo.transaction(fn ->
      Repo.insert(MyApp.User.changeset(%MyApp.User{}, attrs))
    end) do
      {:ok, {:ok, user}} -> {:ok, user}
      {:ok, {:error, changeset}} -> {:error, changeset}
      {:error, reason} -> {:error, reason}
    end
  end
end
```

**Key Pattern**: Use `Repo.transaction/1` for multi-step operations, `Repo.get_for_update!/2` for locking.

---

## 6. API Development with Phoenix

### Building REST APIs

Phoenix is excellent for building JSON APIs with clear, composable patterns.

```elixir
# lib/my_app_web/controllers/user_controller.ex
defmodule MyAppWeb.UserController do
  use MyAppWeb, :controller
  alias MyApp.{User, Repo}
  import Ecto.Query

  action_fallback MyAppWeb.FallbackController

  # List users with filtering and pagination
  def index(conn, params) do
    page = String.to_integer(params["page"] || "1")
    page_size = String.to_integer(params["page_size"] || "10")
    status = params["status"] || "all"

    query = from u in User,
      where: u.is_active == true

    query = case status do
      "premium" -> where(query, [u], u.tier == "premium")
      "regular" -> where(query, [u], u.tier == "regular")
      _ -> query
    end

    total = Repo.aggregate(query, :count, :id)
    offset = (page - 1) * page_size

    users = query
      |> limit(^page_size)
      |> offset(^offset)
      |> Repo.all()

    conn
    |> put_resp_header("x-total-count", to_string(total))
    |> put_resp_header("x-page", to_string(page))
    |> json(users)
  end

  # Get single user
  def show(conn, %{"id" => id}) do
    user = Repo.get!(User, id)
    json(conn, user)
  end

  # Create user
  def create(conn, %{"user" => user_params}) do
    with {:ok, user} <- User.changeset(%User{}, user_params)
                        |> Repo.insert() do
      conn
      |> put_status(:created)
      |> put_resp_header("location", Routes.user_path(conn, :show, user))
      |> json(%{id: user.id, name: user.name, email: user.email})
    end
  end

  # Update user
  def update(conn, %{"id" => id, "user" => user_params}) do
    user = Repo.get!(User, id)
    
    with {:ok, updated_user} <- User.changeset(user, user_params)
                                |> Repo.update() do
      json(conn, updated_user)
    end
  end

  # Delete user
  def delete(conn, %{"id" => id}) do
    Repo.get!(User, id) |> Repo.delete!()
    send_resp(conn, :no_content, "")
  end
end

# lib/my_app_web/controllers/fallback_controller.ex
defmodule MyAppWeb.FallbackController do
  use MyAppWeb, :controller

  def call(conn, {:error, %Ecto.Changeset{} = changeset}) do
    conn
    |> put_status(:unprocessable_entity)
    |> json(%{errors: changeset_errors(changeset)})
  end

  def call(conn, {:error, :not_found}) do
    conn
    |> put_status(:not_found)
    |> json(%{error: "Resource not found"})
  end

  def call(conn, {:error, reason}) when is_atom(reason) do
    conn
    |> put_status(:bad_request)
    |> json(%{error: Atom.to_string(reason)})
  end

  defp changeset_errors(changeset) do
    Ecto.Changeset.traverse_errors(changeset, fn {msg, opts} ->
      Enum.reduce(opts, msg, fn {key, value}, acc ->
        String.replace(acc, "%{#{key}}", to_string(value))
      end)
    end)
  end
end

# lib/my_app_web/router.ex
defmodule MyAppWeb.Router do
  use MyAppWeb, :router

  scope "/api", MyAppWeb do
    pipe_through :api

    resources "/users", UserController
    resources "/posts", PostController
    resources "/comments", CommentController
  end

  # Protected routes
  scope "/api/protected", MyAppWeb do
    pipe_through [:api, :require_auth]
    
    post "/users/:id/admin", AdminController, :promote_to_admin
    delete "/users/:id", UserController, :delete
  end
end
```

**RESTful Pattern**: Use standard HTTP methods (GET, POST, PATCH, DELETE) with predictable URL structures.

### Authentication with Guardian

Guardian provides JWT-based authentication for APIs.

```elixir
# config/config.exs
config :guardian, Guardian,
  issuer: "my_app",
  secret_key: "your-secret-key-change-in-production",
  ttl: {30, :days}

# lib/my_app/authentication.ex
defmodule MyApp.Authentication do
  use Guardian, otp_app: :my_app
  alias MyApp.{User, Repo}

  def subject_for_token(user = %User{}, _claims) do
    {:ok, to_string(user.id)}
  end

  def resource_from_claims(claims) do
    case Repo.get(User, String.to_integer(claims["sub"])) do
      user = %User{} -> {:ok, user}
      _ -> {:error, :not_found}
    end
  end
end

# lib/my_app_web/plugs/auth_plug.ex
defmodule MyAppWeb.AuthPlug do
  use Guardian.Plug.Pipeline, otp_app: :my_app

  plug Guardian.Plug.VerifyHeader
  plug Guardian.Plug.EnsureAuthenticated
  plug Guardian.Plug.LoadResource
end

# lib/my_app_web/controllers/auth_controller.ex
defmodule MyAppWeb.AuthController do
  use MyAppWeb, :controller
  alias MyApp.{User, Repo}
  alias MyApp.Authentication

  def login(conn, %{"email" => email, "password" => password}) do
    case Repo.get_by(User, email: email) do
      user = %User{} ->
        if verify_password(password, user.password_hash) do
          {:ok, token, _claims} = Authentication.encode_and_sign(user)
          json(conn, %{token: token, user: user})
        else
          conn
          |> put_status(:unauthorized)
          |> json(%{error: "Invalid credentials"})
        end

      nil ->
        conn
        |> put_status(:unauthorized)
        |> json(%{error: "Invalid credentials"})
    end
  end

  def me(conn, _) do
    user = Guardian.Plug.current_resource(conn)
    json(conn, user)
  end

  defp verify_password(password, hash) do
    Bcrypt.verify_pass(password, hash)
  end
end

# lib/my_app_web/router.ex
defmodule MyAppWeb.Router do
  scope "/api", MyAppWeb do
    pipe_through :api

    post "/login", AuthController, :login
  end

  scope "/api", MyAppWeb do
    pipe_through [:api, :auth]

    get "/me", AuthController, :me
    resources "/users", UserController
  end
end
```

**Token Management**: Access tokens (short-lived), refresh tokens (long-lived), and secure storage on the client.

---

## 7. Testing and Quality Assurance

### Unit and Integration Testing with ExUnit

ExUnit is Elixir's built-in testing framework with excellent LiveView support.

```elixir
# test/my_app/user_test.exs
defmodule MyApp.UserTest do
  use ExUnit.Case
  alias MyApp.User

  describe "changeset/2" do
    test "valid attributes" do
      attrs = %{
        "name" => "Alice",
        "email" => "alice@example.com",
        "age" => 30
      }

      changeset = User.changeset(%User{}, attrs)
      assert changeset.valid?
    end

    test "requires email and name" do
      changeset = User.changeset(%User{}, %{})
      refute changeset.valid?
      assert "can't be blank" in errors_on(changeset).name
      assert "can't be blank" in errors_on(changeset).email
    end

    test "validates email format" do
      changeset = User.changeset(%User{}, %{"email" => "invalid"})
      refute changeset.valid?
    end

    test "unique email constraint" do
      # Requires database
      assert {:error, changeset} = insert_user(%{email: "test@example.com"})
      assert {:error, _} = insert_user(%{email: "test@example.com"})
    end
  end

  defp insert_user(attrs) do
    Repo.insert(User.changeset(%User{}, attrs))
  end

  defp errors_on(changeset) do
    Ecto.Changeset.traverse_errors(changeset, fn {msg, _opts} -> msg end)
  end
end

# test/my_app_web/controllers/user_controller_test.exs
defmodule MyAppWeb.UserControllerTest do
  use MyAppWeb.ConnCase
  import MyApp.Factory

  describe "GET /api/users" do
    test "lists users with pagination" do
      for i <- 1..15, do: insert(:user, name: "User #{i}")

      conn = get(build_conn(), "/api/users?page=1&page_size=10")
      assert json_response(conn, 200)
      assert length(json_response(conn, 200)) == 10
      assert get_resp_header(conn, "x-total-count") == ["15"]
    end
  end

  describe "POST /api/users" do
    test "creates user with valid attributes" do
      attrs = %{"name" => "John", "email" => "john@example.com"}
      conn = post(build_conn(), "/api/users", %{"user" => attrs})
      
      assert response(conn, 201)
      assert json_response(conn, 201)["id"]
    end

    test "rejects invalid email" do
      attrs = %{"name" => "John", "email" => "invalid"}
      conn = post(build_conn(), "/api/users", %{"user" => attrs})
      
      assert response(conn, 422)
    end
  end

  describe "DELETE /api/users/:id" do
    test "deletes user" do
      user = insert(:user)
      conn = delete(build_conn(), "/api/users/#{user.id}")
      
      assert response(conn, 204)
      assert Repo.get(User, user.id) == nil
    end
  end
end

# test/support/factory.ex
defmodule MyApp.Factory do
  use ExMachina.Ecto, repo: MyApp.Repo

  def user_factory do
    %MyApp.User{
      name: sequence(:name, &"User #{&1}"),
      email: sequence(:email, &"user#{&1}@example.com"),
      age: 30
    }
  end

  def post_factory do
    %MyApp.Post{
      title: "Test Post",
      content: "Test content",
      user: build(:user)
    }
  end
end

# test/my_app_web/live/counter_live_test.exs
defmodule MyAppWeb.CounterLiveTest do
  use MyAppWeb.ConnCase
  import Phoenix.LiveViewTest

  test "increments counter" do
    {:ok, view, html} = live(conn, "/counter")
    assert html =~ "Count: 0"

    render_click(view, "increment")
    assert render(view) =~ "Count: 1"
  end

  test "handles multiple state changes" do
    {:ok, view, _html} = live(conn, "/counter")

    render_click(view, "increment")
    render_click(view, "increment")
    render_click(view, "decrement")

    assert render(view) =~ "Count: 1"
  end

  test "broadcasts updates to other clients" do
    {:ok, view1, _} = live(conn, "/counter")
    {:ok, view2, _} = live(conn, "/counter")

    render_click(view1, "increment")

    # Both views should see the update
    assert render(view1) =~ "Count: 1"
    assert render(view2) =~ "Count: 1"
  end
end
```

**Coverage Target**: Aim for 85%+ coverage with a mix of unit tests (functions), integration tests (API endpoints), and LiveView tests (UI interactions).

### Code Quality with Credo

Credo analyzes code for style and maintainability issues.

```bash
# Install
mix erts.add credo

# Run checks
mix credo

# Generate report
mix credo --format=json > credo-report.json

# Strict mode
mix credo --strict
```

```elixir
# .credo.exs configuration
%{
  configs: [
    %{
      name: "default",
      files: %{
        included: ["lib/", "test/"],
        excluded: [~r"/_build/", ~r"/deps/"]
      },
      checks: [
        {Credo.Check.Design.DuplicatedCode, []}
      ]
    }
  ]
}
```

---

## 8. Deployment and DevOps

### Building Releases

Phoenix apps are deployed as Erlang releases with hot code upgrades.

```yaml
# mix.exs
def project do
  [
    app: :my_app,
    version: "0.1.0",
    elixir: "~> 1.14",
    compilers: [:gettext] ++ Mix.compilers(),
    start_permanent: Mix.env() == :prod,
    deps: deps(),
    releases: [
      my_app: [
        include_executables_for: [:unix],
        applications: [runtime_tools: :permanent]
      ]
    ]
  ]
end
```

```bash
# Create release
MIX_ENV=prod mix release

# Run release
./_build/prod/rel/my_app/bin/my_app start

# Interactive console
./_build/prod/rel/my_app/bin/my_app remote
```

**Hot Upgrade**: Erlang allows updating code without stopping the system, a unique capability.

### Docker Deployment

Multi-stage Docker build for minimal image size.

```dockerfile
# Dockerfile
FROM elixir:1.16 as builder

WORKDIR /app
RUN mix local.hex --force && mix local.rebar --force

COPY mix.exs mix.lock ./
COPY config config
RUN MIX_ENV=prod mix deps.get --only prod
RUN MIX_ENV=prod mix compile

COPY lib lib
COPY priv priv
RUN MIX_ENV=prod mix release

# Final stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    openssl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/_build/prod/rel/my_app .

ENV MIX_ENV=prod

EXPOSE 4000
CMD ["bin/my_app", "start"]
```

```bash
# Build and deploy
docker build -t my-app:latest .
docker push my-app:latest

# Run
docker run -p 4000:4000 -e PHX_HOST=example.com my-app:latest
```

### Monitoring with Telemetry

Telemetry collects metrics for observability.

```elixir
# lib/my_app/telemetry.ex
defmodule MyApp.Telemetry do
  use Supervisor
  import Telemetry.Metrics

  def start_link(arg) do
    Supervisor.start_link(__MODULE__, arg, name: __MODULE__)
  end

  @impl true
  def init(_arg) do
    children = [
      {TelemetryMetrics.Prometheus, [metrics: metrics()]}
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end

  def metrics do
    [
      # Phoenix LiveView metrics
      summary("phoenix.live_view.mount.stop", unit: {:native, :microsecond}),
      counter("phoenix.live_view.render.stop"),

      # HTTP metrics
      summary("phoenix.router_dispatch.stop", unit: {:native, :microsecond}),
      counter("phoenix.router_dispatch.exception"),

      # Database metrics
      summary("my_app.repo.query.total_time", unit: {:native, :microsecond}),
      counter("my_app.repo.query.queue_time", unit: {:native, :microsecond}),

      # Custom application metrics
      gauge("my_app.process.count"),
      counter("my_app.user.created")
    ]
  end
end

# lib/my_app/application.ex
def start(_type, _args) do
  children = [
    MyApp.Telemetry,  # Add telemetry
    MyApp.Repo,
    {Phoenix.PubSub, name: MyApp.PubSub},
    MyAppWeb.Endpoint
  ]

  Supervisor.start_link(children, strategy: :one_for_one, name: MyApp.Supervisor)
end

# Emit custom metrics
:telemetry.execute([:my_app, :user, :created], %{}, %{user_id: user.id})
```

**Observability Stack**: Combine with Prometheus/Grafana for dashboards and alerting.

---

## Best Practices and Patterns

### Immutability by Default

Elixir enforces immutability, making code predictable and concurrent-safe.

```elixir
# Anti-pattern: Attempting mutation (fails)
list = [1, 2, 3]
list[0] = 10  # Error!

# Correct: Create new list
list = [1, 2, 3]
new_list = [10 | tl(list)]  # [10, 2, 3]

# With pipes
new_list = list
  |> List.delete_at(0)
  |> List.insert_at(0, 10)
```

### Function-Oriented Design

Favor small, composable functions over large methods.

```elixir
# Anti-pattern: Large monolithic function
def process_user(user_data) do
  user = parse_json(user_data)
  validate_email(user)
  validate_age(user)
  save_to_db(user)
  send_welcome_email(user)
  log_user_created(user)
end

# Better: Composable pipeline
def process_user(user_data) do
  user_data
  |> parse_json()
  |> validate_email()
  |> validate_age()
  |> save_to_db()
  |> case do
    {:ok, user} ->
      send_welcome_email(user)
      log_user_created(user)
      {:ok, user}
    error -> error
  end
end
```

### Error Handling Strategy

Return tuples instead of raising exceptions in most cases.

```elixir
# Anti-pattern: Excessive exceptions
def get_user(id) do
  Repo.get!(User, id)  # Raises if not found
end

# Better: Return tuples
def get_user(id) do
  case Repo.get(User, id) do
    user = %User{} -> {:ok, user}
    nil -> {:error, :not_found}
  end
end

# Let supervisors handle critical failures
def start_critical_service() do
  {:ok, _pid} = GenServer.start_link(CriticalService, [])
  # Supervisor will restart if it crashes
end
```

### Performance Optimization

Profile before optimizing; Elixir is often fast enough.

```elixir
defmodule PerformanceExample do
  def benchmark_list_operations() do
    list = 1..10000 |> Enum.to_list()

    {time1, _} = :timer.tc(fn ->
      Enum.filter(list, &(rem(&1, 2) == 0))
      |> Enum.map(&(&1 * 2))
      |> Enum.sum()
    end)

    {time2, _} = :timer.tc(fn ->
      list
      |> Flow.from_enumerable()
      |> Flow.filter(&(rem(&1, 2) == 0))
      |> Flow.map(&(&1 * 2))
      |> Flow.reduce(fn -> 0 end, &+/2)
      |> Enum.to_list()
    end)

    IO.puts("Enum: #{time1}µs, Flow: #{time2}µs")
  end

  # Use profiling tools
  def profile_function() do
    :fprof.apply(&expensive_function/0, [])
    :fprof.profile()
  end

  defp expensive_function() do
    1..1000
    |> Enum.map(&complex_calculation/1)
    |> Enum.reduce(0, &+/2)
  end

  defp complex_calculation(n) do
    n * n + n * 2 + 1
  end
end
```

**Rules**: Use Enum for small data, Stream/Flow for large data, profile before optimizing.

---

## TRUST 5 Compliance

### Test-First: ExUnit Test Patterns

```elixir
# Every module has tests
defmodule MyApp.UserTest do
  use ExUnit.Case
  doctest MyApp.User

  setup do
    # Setup test data
    user = %MyApp.User{name: "Test", email: "test@example.com"}
    {:ok, user: user}
  end

  test "validates required fields", %{user: user} do
    changeset = MyApp.User.changeset(user, %{})
    refute changeset.valid?
  end

  test "accepts valid attributes" do
    attrs = %{name: "John", email: "john@example.com"}
    changeset = MyApp.User.changeset(%MyApp.User{}, attrs)
    assert changeset.valid?
  end
end
```

Target: 85%+ coverage, test behaviors not implementations.

### Readable: Clear Code with Patterns

- Use descriptive function names
- Follow Elixir conventions (pattern matching, guards)
- Leverage pipes for data flow
- Add doc strings to public functions

```elixir
defmodule MyApp.UserService do
  @doc """
  Retrieves a user by email, returning {:ok, user} or {:error, :not_found}.
  """
  def get_by_email(email) when is_binary(email) do
    case Repo.get_by(User, email: email) do
      user = %User{} -> {:ok, user}
      nil -> {:error, :not_found}
    end
  end
end
```

### Unified: Community Standards

- Follow Elixir style guide
- Use consistent naming (snake_case)
- Leverage existing libraries (Phoenix, Ecto)
- Document unusual patterns

### Secured: Authentication and Validation

- Use Guardian for authentication
- Validate all inputs via changesets
- Hash passwords with Bcrypt
- Use HTTPS in production

```elixir
def create_user(attrs) do
  %User{}
  |> User.changeset(attrs)
  |> Repo.insert()
end

# Changesets validate all fields
defmodule User do
  def changeset(user, attrs) do
    user
    |> cast(attrs, [:name, :email, :password])
    |> validate_required([:name, :email, :password])
    |> validate_format(:email, ~r/@/)
    |> put_password_hash()
    |> unique_constraint(:email)
  end
end
```

### Trackable: Version Management and Logs

- Use git tags for versions
- Document changes in CHANGELOG
- Use structured logging

```elixir
defmodule MyApp.UserLog do
  require Logger

  def log_user_created(user) do
    Logger.info("User created", user_id: user.id, email: user.email)
  end
end
```

---

## Further Reading and Resources

- **Official Elixir Guide**: https://elixir-lang.org/getting-started/introduction.html
- **Phoenix Framework**: https://www.phoenixframework.org/
- **Ecto Documentation**: https://hexdocs.pm/ecto/
- **Guardian Authentication**: https://github.com/ueberauth/guardian
- **Elixir School**: https://elixirschool.com/

---

**Version**: 4.0.0  
**Last Updated**: 2025-11-19  
**Framework Versions**: Elixir 1.18+, Phoenix 1.7.14+, Ecto 3.12+  
**Status**: Stable and Production-Ready
