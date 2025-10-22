# moai-lang-elixir - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Mix & ExUnit

```bash
# Create new Elixir project
mix new my_app
cd my_app

# Create Phoenix web project
mix phx.new my_app_web
cd my_app_web

# Project structure created:
# my_app/
# ├── lib/
# │   └── my_app.ex
# ├── test/
# │   ├── my_app_test.exs
# │   └── test_helper.exs
# ├── mix.exs
# └── .formatter.exs
```

**mix.exs configuration**:
```elixir
defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      version: "0.1.0",
      elixir: "~> 1.19",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      test_coverage: [tool: ExCoveralls],
      preferred_cli_env: [
        coveralls: :test,
        "coveralls.detail": :test,
        "coveralls.html": :test
      ]
    ]
  end

  def application do
    [
      extra_applications: [:logger]
    ]
  end

  defp deps do
    [
      {:credo, "~> 1.7", only: [:dev, :test], runtime: false},
      {:dialyxir, "~> 1.4", only: [:dev, :test], runtime: false},
      {:excoveralls, "~> 0.18", only: :test},
      {:ex_doc, "~> 0.34", only: :dev, runtime: false}
    ]
  end
end
```

## Example 2: TDD Workflow with ExUnit

**RED: Write failing test**
```elixir
# test/calculator_test.exs
defmodule CalculatorTest do
  use ExUnit.Case
  # @TEST:CALC-001

  describe "add/2" do
    test "adds two positive numbers" do
      assert Calculator.add(2, 3) == 5
    end

    test "adds negative numbers" do
      assert Calculator.add(-2, -3) == -5
    end

    test "adds zero" do
      assert Calculator.add(5, 0) == 5
    end
  end

  describe "divide/2" do
    test "divides two numbers" do
      assert Calculator.divide(10, 2) == 5.0
    end

    test "returns error on division by zero" do
      assert Calculator.divide(10, 0) == {:error, :division_by_zero}
    end
  end
end
```

**GREEN: Implement feature**
```elixir
# lib/calculator.ex
defmodule Calculator do
  @moduledoc """
  @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: calculator_test.exs
  Basic calculator with arithmetic operations
  """

  @doc """
  Adds two numbers.

  ## Examples

      iex> Calculator.add(2, 3)
      5

      iex> Calculator.add(-1, 1)
      0

  """
  def add(a, b), do: a + b

  @doc """
  Subtracts two numbers.
  """
  def subtract(a, b), do: a - b

  @doc """
  Divides two numbers.

  Returns `{:error, :division_by_zero}` if denominator is zero.

  ## Examples

      iex> Calculator.divide(10, 2)
      5.0

      iex> Calculator.divide(10, 0)
      {:error, :division_by_zero}

  """
  def divide(_a, 0), do: {:error, :division_by_zero}
  def divide(a, b), do: a / b
end
```

**REFACTOR: Improve with pattern matching and error handling**
```elixir
# lib/calculator.ex (improved)
defmodule Calculator do
  @moduledoc """
  @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: calculator_test.exs
  Calculator with enhanced error handling
  """

  @type result :: {:ok, number()} | {:error, atom()}

  @spec add(number(), number()) :: number()
  def add(a, b) when is_number(a) and is_number(b), do: a + b

  @spec subtract(number(), number()) :: number()
  def subtract(a, b) when is_number(a) and is_number(b), do: a - b

  @spec divide(number(), number()) :: result()
  def divide(_a, 0), do: {:error, :division_by_zero}
  def divide(a, b) when is_number(a) and is_number(b), do: {:ok, a / b}

  @spec multiply(number(), number()) :: number()
  def multiply(a, b) when is_number(a) and is_number(b), do: a * b
end
```

## Example 3: GenServer for State Management

```elixir
# test/counter_server_test.exs
defmodule CounterServerTest do
  use ExUnit.Case
  # @TEST:STATE-001

  describe "CounterServer" do
    setup do
      {:ok, pid} = CounterServer.start_link(initial_value: 0)
      %{pid: pid}
    end

    test "starts with initial value", %{pid: pid} do
      assert CounterServer.get(pid) == 0
    end

    test "increments counter", %{pid: pid} do
      CounterServer.increment(pid)
      assert CounterServer.get(pid) == 1
    end

    test "decrements counter", %{pid: pid} do
      CounterServer.increment(pid)
      CounterServer.increment(pid)
      CounterServer.decrement(pid)
      assert CounterServer.get(pid) == 1
    end

    test "resets counter", %{pid: pid} do
      CounterServer.increment(pid)
      CounterServer.reset(pid)
      assert CounterServer.get(pid) == 0
    end
  end
end
```

**Implementation**:
```elixir
# lib/counter_server.ex
defmodule CounterServer do
  @moduledoc """
  @CODE:STATE-001 | SPEC: SPEC-STATE-001.md | TEST: counter_server_test.exs
  Counter GenServer with state management
  """
  use GenServer

  # Client API

  def start_link(opts \\ []) do
    initial_value = Keyword.get(opts, :initial_value, 0)
    GenServer.start_link(__MODULE__, initial_value, name: __MODULE__)
  end

  def get(pid), do: GenServer.call(pid, :get)
  def increment(pid), do: GenServer.cast(pid, :increment)
  def decrement(pid), do: GenServer.cast(pid, :decrement)
  def reset(pid), do: GenServer.cast(pid, :reset)

  # Server Callbacks

  @impl true
  def init(initial_value) do
    {:ok, initial_value}
  end

  @impl true
  def handle_call(:get, _from, state) do
    {:reply, state, state}
  end

  @impl true
  def handle_cast(:increment, state) do
    {:noreply, state + 1}
  end

  @impl true
  def handle_cast(:decrement, state) do
    {:noreply, state - 1}
  end

  @impl true
  def handle_cast(:reset, _state) do
    {:noreply, 0}
  end
end
```

## Example 4: Async Testing with Task

```elixir
# test/async_fetcher_test.exs
defmodule AsyncFetcherTest do
  use ExUnit.Case, async: true
  # @TEST:ASYNC-001

  describe "fetch_all/1" do
    test "fetches multiple items concurrently" do
      ids = [1, 2, 3, 4, 5]
      start_time = System.monotonic_time(:millisecond)

      results = AsyncFetcher.fetch_all(ids)

      duration = System.monotonic_time(:millisecond) - start_time

      assert length(results) == 5
      # Should complete in parallel, not sequentially
      assert duration < 1500
    end

    test "handles partial failures gracefully" do
      ids = [1, -1, 2, -2, 3]

      results = AsyncFetcher.fetch_all(ids)

      assert length(results) == 5
      assert Enum.count(results, &match?({:ok, _}, &1)) == 3
      assert Enum.count(results, &match?({:error, _}, &1)) == 2
    end
  end
end
```

**Implementation**:
```elixir
# lib/async_fetcher.ex
defmodule AsyncFetcher do
  @moduledoc """
  @CODE:ASYNC-001 | SPEC: SPEC-ASYNC-001.md | TEST: async_fetcher_test.exs
  Async data fetcher with concurrent operations
  """

  def fetch_all(ids) do
    ids
    |> Enum.map(&Task.async(fn -> fetch_item(&1) end))
    |> Enum.map(&Task.await/1)
  end

  defp fetch_item(id) when id < 0 do
    {:error, :invalid_id}
  end

  defp fetch_item(id) do
    # Simulate API call
    Process.sleep(500)
    {:ok, %{id: id, data: "Data for #{id}"}}
  end
end
```

## Example 5: Mocking with Mox

```elixir
# test/user_service_test.exs
defmodule UserServiceTest do
  use ExUnit.Case, async: true
  # @TEST:USER-001

  import Mox

  # Make sure mocks are verified when test exits
  setup :verify_on_exit!

  describe "create_user/1" do
    test "creates user with valid email" do
      email = "test@example.com"

      UserRepoMock
      |> expect(:find_by_email, fn ^email -> nil end)
      |> expect(:save, fn user -> {:ok, user} end)

      assert {:ok, user} = UserService.create_user(email)
      assert user.email == email
    end

    test "rejects invalid email" do
      assert {:error, :invalid_email} = UserService.create_user("invalid")
    end

    test "handles duplicate email" do
      email = "existing@example.com"
      existing_user = %{id: "123", email: email}

      UserRepoMock
      |> expect(:find_by_email, fn ^email -> existing_user end)

      assert {:error, :duplicate_email} = UserService.create_user(email)
    end
  end
end
```

**Service implementation**:
```elixir
# lib/user_service.ex
defmodule UserService do
  @moduledoc """
  @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: user_service_test.exs
  User management with validation
  """

  @repo Application.compile_env(:my_app, :user_repo, UserRepo)

  def create_user(email) do
    with :ok <- validate_email(email),
         nil <- @repo.find_by_email(email),
         user <- build_user(email),
         {:ok, saved_user} <- @repo.save(user) do
      {:ok, saved_user}
    else
      {:error, reason} -> {:error, reason}
      %{} -> {:error, :duplicate_email}
    end
  end

  defp validate_email(email) do
    if email =~ ~r/^[\w._%+-]+@[\w.-]+\.[A-Za-z]{2,}$/ do
      :ok
    else
      {:error, :invalid_email}
    end
  end

  defp build_user(email) do
    %{
      id: generate_id(),
      email: email,
      created_at: DateTime.utc_now()
    }
  end

  defp generate_id, do: :crypto.strong_rand_bytes(16) |> Base.encode64()
end
```

**Mock configuration** (`test/test_helper.exs`):
```elixir
Mox.defmock(UserRepoMock, for: UserRepoBehaviour)
Application.put_env(:my_app, :user_repo, UserRepoMock)

ExUnit.start()
```

## Example 6: Doctests

```elixir
# lib/string_utils.ex
defmodule StringUtils do
  @moduledoc """
  @CODE:STR-001 | SPEC: SPEC-STR-001.md
  String utility functions with doctests
  """

  @doc """
  Trims whitespace from both ends of a string.

  ## Examples

      iex> StringUtils.trim("  hello  ")
      "hello"

      iex> StringUtils.trim("world")
      "world"

      iex> StringUtils.trim("")
      ""

  """
  def trim(str) when is_binary(str) do
    String.trim(str)
  end

  @doc """
  Capitalizes the first letter of each word.

  ## Examples

      iex> StringUtils.title_case("hello world")
      "Hello World"

      iex> StringUtils.title_case("ELIXIR")
      "Elixir"

  """
  def title_case(str) when is_binary(str) do
    str
    |> String.downcase()
    |> String.split()
    |> Enum.map(&String.capitalize/1)
    |> Enum.join(" ")
  end
end
```

**Test with doctests**:
```elixir
# test/string_utils_test.exs
defmodule StringUtilsTest do
  use ExUnit.Case
  doctest StringUtils

  # Additional tests beyond doctests
  describe "trim/1" do
    test "handles tabs and newlines" do
      assert StringUtils.trim("\t\nhello\n\t") == "hello"
    end
  end
end
```

## Example 7: Quality Gate Check

```bash
# Run all tests
mix test

# Run with coverage
mix coveralls

# Generate HTML coverage report
mix coveralls.html
open cover/excoveralls.html

# Run static analysis (Credo)
mix credo --strict

# Run type checking (Dialyzer)
mix dialyzer

# Format code
mix format

# Check formatting
mix format --check-formatted

# TRUST 5 validation
echo "T - Test coverage"
mix coveralls.html
# Verify coverage ≥85% in cover/excoveralls.html

echo "R - Readable code"
mix credo --strict
mix format --check-formatted

echo "U - Unified types"
mix dialyzer

echo "S - Security"
mix hex.audit

echo "T - Trackable with @TAG"
rg '@(CODE|TEST|SPEC):' -n lib/ test/ --type elixir
```

## Example 8: Property-Based Testing with StreamData

```elixir
# test/math_properties_test.exs
defmodule MathPropertiesTest do
  use ExUnit.Case
  use ExUnitProperties

  property "addition is commutative" do
    check all a <- integer(),
              b <- integer() do
      assert Calculator.add(a, b) == Calculator.add(b, a)
    end
  end

  property "multiplication by zero returns zero" do
    check all n <- integer() do
      assert Calculator.multiply(n, 0) == 0
    end
  end

  property "string reversal is involutive" do
    check all str <- string(:printable) do
      assert str |> String.reverse() |> String.reverse() == str
    end
  end
end
```

---

## TRUST 5 Integration

### Test Coverage (≥85%)
```bash
mix coveralls.html
# Check: cover/excoveralls.html
```

### Readable Code
```bash
mix credo --strict
mix format --check-formatted
```

### Unified Types
```bash
mix dialyzer
```

### Security
```bash
mix hex.audit
mix deps.audit
```

### Trackable with @TAG
```bash
rg '@(CODE|TEST|SPEC):' -n lib/ test/ --type elixir
```

---

_For detailed CLI reference, see [reference.md](reference.md)_
