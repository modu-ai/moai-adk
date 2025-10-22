# Elixir 1.18+ CLI Reference — Tool Command Matrix

**Framework**: Elixir 1.18+ + Mix + ExUnit + Credo 1.7+

---

## Elixir Runtime Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `elixir --version` | Check Elixir version | `elixir --version` → `Elixir 1.18.4` |
| `elixir script.exs` | Run Elixir script | `elixir script.exs` |
| `elixir -e "code"` | Execute Elixir code inline | `elixir -e "IO.puts System.version"` |
| `elixir --help` | Show Elixir help | `elixir --help` |
| `iex` | Start interactive Elixir shell | `iex` |
| `iex -S mix` | Start IEx with Mix project | `iex -S mix` |

---

## Build Tool (Mix)

| Command | Purpose | Example |
|---------|---------|---------|
| `mix new project` | Create new project | `mix new my_app` |
| `mix new --sup project` | Create with supervisor | `mix new --sup my_app` |
| `mix compile` | Compile project | `mix compile` |
| `mix run` | Run project | `mix run` |
| `mix test` | Run tests | `mix test` |
| `mix test test_file.exs` | Run specific test file | `mix test test/calculator_test.exs` |
| `mix test test_file.exs:10` | Run test at line 10 | `mix test test/calculator_test.exs:10` |
| `mix test --trace` | Run tests with detailed trace | `mix test --trace` |
| `mix test --failed` | Run only failed tests | `mix test --failed` |
| `mix test --stale` | Run stale tests | `mix test --stale` |
| `mix test --cover` | Run tests with coverage | `mix test --cover` |
| `mix deps.get` | Install dependencies | `mix deps.get` |
| `mix deps.update --all` | Update all dependencies | `mix deps.update --all` |
| `mix deps.tree` | Show dependency tree | `mix deps.tree` |
| `mix hex.info package` | Show package info | `mix hex.info phoenix` |
| `mix ecto.create` | Create database | `mix ecto.create` |
| `mix ecto.migrate` | Run migrations | `mix ecto.migrate` |
| `mix ecto.rollback` | Rollback migration | `mix ecto.rollback` |
| `mix phx.new web_app` | Create Phoenix app | `mix phx.new my_app_web` |
| `mix phx.server` | Start Phoenix server | `mix phx.server` |
| `mix release` | Build production release | `mix release` |
| `mix clean` | Clean build artifacts | `mix clean` |
| `mix format` | Format code | `mix format` |
| `mix format --check-formatted` | Check if formatting needed | `mix format --check-formatted` |

---

## Testing Framework (ExUnit)

| Command | Purpose | Example |
|---------|---------|---------|
| `mix test` | Run all tests | `mix test` |
| `mix test --trace` | Verbose test output | `mix test --trace` |
| `mix test --only focus` | Run tagged tests | `mix test --only focus` |
| `mix test --exclude slow` | Skip tagged tests | `mix test --exclude slow` |
| `mix test --seed 12345` | Run with specific seed | `mix test --seed 12345` |
| `mix test --max-failures 3` | Stop after 3 failures | `mix test --max-failures 3` |

**Test Structure**:
```elixir
defmodule CalculatorTest do
  use ExUnit.Case
  # @TEST:CALC-001

  describe "add/2" do
    test "adds two positive numbers" do
      assert Calculator.add(2, 3) == 5
    end

    @tag :slow
    test "adds large numbers" do
      assert Calculator.add(1_000_000, 2_000_000) == 3_000_000
    end
  end
end
```

**New in Elixir 1.18**: Parameterized tests and test grouping
```elixir
defmodule CalculatorTest do
  use ExUnit.Case, async: true

  # Tests in same group don't run concurrently
  @tag group: :math
  test "addition" do
    assert Calculator.add(1, 2) == 3
  end

  @tag group: :math
  test "subtraction" do
    assert Calculator.subtract(5, 3) == 2
  end
end
```

---

## Code Quality (Credo 1.7+)

| Command | Purpose | Example |
|---------|---------|---------|
| `mix credo` | Run code analysis | `mix credo` |
| `mix credo --strict` | Strict analysis | `mix credo --strict` |
| `mix credo list` | List all issues | `mix credo list` |
| `mix credo suggest` | Get suggestions | `mix credo suggest` |
| `mix credo --all` | Check all files | `mix credo --all` |
| `mix credo --format oneline` | One-line format | `mix credo --format oneline` |

**.credo.exs Configuration**:
```elixir
%{
  configs: [
    %{
      name: "default",
      strict: true,
      color: true,
      files: %{
        included: ["lib/", "test/"],
        excluded: [~r"/_build/", ~r"/deps/"]
      },
      checks: %{
        enabled: [
          {Credo.Check.Readability.ModuleDoc, []},
          {Credo.Check.Refactor.CyclomaticComplexity, [max_complexity: 10]}
        ],
        disabled: [
          {Credo.Check.Design.AliasUsage, []}
        ]
      }
    }
  ]
}
```

---

## Type Checking (Dialyzer)

| Command | Purpose | Example |
|---------|---------|---------|
| `mix dialyzer` | Run type analysis | `mix dialyzer` |
| `mix dialyzer --halt-exit-status` | Exit with error on issues | `mix dialyzer --halt-exit-status` |

**mix.exs Configuration**:
```elixir
defp deps do
  [
    {:dialyxir, "~> 1.4", only: [:dev, :test], runtime: false}
  ]
end
```

**Type Specifications**:
```elixir
defmodule Calculator do
  @spec add(number(), number()) :: number()
  def add(a, b), do: a + b

  @spec divide(number(), number()) :: {:ok, number()} | {:error, atom()}
  def divide(_a, 0), do: {:error, :division_by_zero}
  def divide(a, b), do: {:ok, a / b}
end
```

---

## Code Coverage (ExCoveralls)

| Command | Purpose | Example |
|---------|---------|---------|
| `mix coveralls` | Generate coverage | `mix coveralls` |
| `mix coveralls.html` | Generate HTML report | `mix coveralls.html` |
| `mix coveralls.detail` | Detailed coverage | `mix coveralls.detail` |
| `mix coveralls.github` | Upload to GitHub | `mix coveralls.github` |

**mix.exs Configuration**:
```elixir
def project do
  [
    test_coverage: [tool: ExCoveralls],
    preferred_cli_env: [
      coveralls: :test,
      "coveralls.detail": :test,
      "coveralls.html": :test
    ]
  ]
end

defp deps do
  [
    {:excoveralls, "~> 0.18", only: :test}
  ]
end
```

---

## JSON Support (New in Elixir 1.18)

**Built-in JSON Module**:
```elixir
# Encode
JSON.encode!(%{name: "Alice", age: 30})
#=> "{"name":"Alice","age":30}"

# Decode
JSON.decode!("{"name":"Alice"}")
#=> %{"name" => "Alice"}
```

---

## Interactive Development (IEx)

### IEx Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `h Module.function` | Show help | `h Enum.map` |
| `i value` | Inspect value | `i [1, 2, 3]` |
| `v()` | Show last value | `v()` |
| `v(5)` | Show 5th value | `v(5)` |
| `c "file.ex"` | Compile file | `c "lib/calculator.ex"` |
| `r Module` | Reload module | `r Calculator` |
| `respawn` | Restart IEx session | `respawn` |
| `clear` | Clear screen | `clear` |

---

## Project Structure

**Standard Layout**:
```
my_app/
├── lib/
│   ├── my_app.ex
│   └── my_app/
│       └── calculator.ex
├── test/
│   ├── test_helper.exs
│   └── my_app/
│       └── calculator_test.exs
├── mix.exs
├── .formatter.exs
└── .credo.exs
```

**mix.exs Template**:
```elixir
defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      version: "0.1.0",
      elixir: "~> 1.18",
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

**.formatter.exs**:
```elixir
[
  inputs: ["{mix,.formatter}.exs", "{config,lib,test}/**/*.{ex,exs}"],
  line_length: 100
]
```

---

## Combined Workflow (Quality Gate)

**Before Commit** (all must pass):

```bash
#!/bin/bash
set -e

echo "Running Elixir quality gate checks..."

# 1. Format code
echo "1. Checking formatting..."
mix format --check-formatted

# 2. Run tests with coverage
echo "2. Running tests..."
mix coveralls.html

# 3. Check coverage threshold
echo "3. Checking coverage..."
# View cover/excoveralls.html

# 4. Run static analysis
echo "4. Running Credo..."
mix credo --strict

# 5. Run type checking
echo "5. Running Dialyzer..."
mix dialyzer

echo "✅ All quality gates passed!"
```

**Save as** `bin/quality-gate.sh`:
```bash
chmod +x bin/quality-gate.sh
./bin/quality-gate.sh
```

---

## CI/CD Integration (GitHub Actions)

**.github/workflows/test.yml**:
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        elixir: ['1.17', '1.18']
        otp: ['26', '27']

    steps:
      - uses: actions/checkout@v4

      - name: Set up Elixir
        uses: erlef/setup-beam@v1
        with:
          elixir-version: ${{ matrix.elixir }}
          otp-version: ${{ matrix.otp }}

      - name: Cache dependencies
        uses: actions/cache@v4
        with:
          path: |
            deps
            _build
          key: ${{ runner.os }}-mix-${{ hashFiles('**/mix.lock') }}

      - name: Install dependencies
        run: mix deps.get

      - name: Check formatting
        run: mix format --check-formatted

      - name: Run tests
        run: mix test

      - name: Run coverage
        run: mix coveralls.github

      - name: Credo
        run: mix credo --strict

      - name: Dialyzer
        run: mix dialyzer
```

---

## Phoenix Framework Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `mix phx.new app` | Create Phoenix app | `mix phx.new my_app_web` |
| `mix phx.server` | Start server | `mix phx.server` |
| `mix phx.gen.html` | Generate HTML scaffold | `mix phx.gen.html Accounts User users` |
| `mix phx.gen.json` | Generate JSON API | `mix phx.gen.json Accounts User users` |
| `mix phx.gen.context` | Generate context | `mix phx.gen.context Accounts User users` |
| `mix phx.routes` | Show all routes | `mix phx.routes` |

---

## TRUST 5 Principles Integration

### T - Test First (ExUnit)
```bash
mix test --cover
```

### R - Readable (Credo + Format)
```bash
mix credo --strict
mix format --check-formatted
```

### U - Unified Types (Dialyzer)
```bash
mix dialyzer
```

### S - Security
```bash
mix hex.audit
mix deps.audit
```

### T - Trackable (@TAG)
```bash
rg '@(CODE|TEST|SPEC):' -n lib/ test/ --type elixir
```

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: Elixir 1.18+ CLI Tools Reference
