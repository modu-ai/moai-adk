---

name: moai-lang-elixir
description: Elixir functional programming, OTP patterns, and Phoenix framework best practices
allowed-tools: [Read, Bash, WebFetch]

---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, elixir, moai  


# Elixir Functional Programming — Enterprise

## Quick Reference (30 seconds)

Elixir는 함수형 프로그래밍, 불변성(Immutability), OTP 감독 트리(Supervision Trees)를 통해
고동시성(High-Concurrency) 분산 시스템을 구축합니다. 경량 프로세스, 패턴 매칭,
"let it crash" 철학으로 자가 복구 시스템을 만들 수 있습니다.

**핵심 개념**:
- **불변성**: 모든 데이터 구조 불변
- **패턴 매칭**: 함수 절(Function Clauses) 및 Case 식
- **동시성**: 액터 모델 기반 경량 프로세스 (수백만 개 동시 실행)
- **자동 복구**: OTP 감독 트리 (프로세스 재시작, 캐스케이딩 복구)

**에코시스템** (November 2025):
- **Phoenix 1.7**: 웹 프레임워크 (라우팅, 미들웨어, WebSocket)
- **LiveView**: 서버 렌더링 실시간 UI (JavaScript 불필요)
- **Ecto 3.11**: 데이터베이스 라이브러리 (쿼리 빌더, 마이그레이션)
- **GenServer**: 일반 서버 동작 (상태 관리, 동시성 처리)


## Implementation Guide

### 1. 패턴 매칭 (Pattern Matching)

**기본 패턴 매칭**:
```elixir
# 함수 절 기반 패턴 매칭
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

**구조적 분해 (Destructuring)**:
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

### 2. GenServer 패턴 (상태 관리)

**기본 GenServer**:
```elixir
defmodule Counter do
  use GenServer

  def start_link(initial_count \\ 0) do
    GenServer.start_link(__MODULE__, initial_count, name: __MODULE__)
  end

  # 초기화
  def init(initial_count) do
    {:ok, initial_count}
  end

  # 동기 호출 (응답 대기)
  def increment(amount \\ 1) do
    GenServer.call(__MODULE__, {:increment, amount})
  end

  # 비동기 호출 (응답 미대기)
  def log(message) do
    GenServer.cast(__MODULE__, {:log, message})
  end

  # 핸들러: 동기 호출
  def handle_call({:increment, amount}, _from, count) do
    new_count = count + amount
    {:reply, new_count, new_count}
  end

  # 핸들러: 비동기 호출
  def handle_cast({:log, message}, state) do
    IO.puts("Log: #{message}")
    {:noreply, state}
  end
end
```

### 3. OTP 감독 트리 (Supervision)

**감독자 구조**:
```elixir
defmodule MyApp.Supervisor do
  use Supervisor

  def start_link(opts) do
    Supervisor.start_link(__MODULE__, opts, name: __MODULE__)
  end

  def init(_opts) do
    children = [
      # 데이터베이스 연결 풀
      {Ecto.Repo, []},

      # 메시지 브로커
      {Phoenix.PubSub, name: MyApp.PubSub},

      # 커스텀 워커
      {Counter, 0}
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end
end

# 감독 전략:
# :one_for_one - 자식 프로세스만 재시작
# :one_for_all - 모든 자식 프로세스 재시작
# :rest_for_one - 해당 프로세스 이후의 자식들 재시작
```

### 4. 함수형 프로그래밍 패턴

**Map/Filter/Reduce**:
```elixir
defmodule List do
  def process_data(list) do
    list
    |> Enum.map(&(&1 * 2))              # 각 원소를 2배
    |> Enum.filter(&(&1 > 10))          # 10 이상만 필터링
    |> Enum.reduce(0, &+/2)             # 합계 계산
  end
end

iex> List.process_data([1, 5, 7, 10, 15])
64
```

**Pipe 연산자 활용**:
```elixir
defmodule DataPipeline do
  def transform(data) do
    data
    |> parse_json()          # 1단계
    |> validate()            # 2단계
    |> enrich_with_metadata()  # 3단계
    |> persist_to_db()       # 4단계
  end
end
```

### 5. Ecto 데이터베이스 패턴

**Changeset 검증**:
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

### 6. Phoenix LiveView 실시간 UI

**기본 LiveView**:
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
- **패턴 매칭 활용**: 복잡한 if/else 대신 패턴 매칭 사용
- **감독 계층 설계**: 프로세스 체계적 관리 (let it crash)
- **불변성 활용**: 스레드 안전성 보장
- **Ecto Changeset 검증**: 데이터 일관성 보장
- **파이프 연산자**: 순차적 변환 체인
- **타입 스펙 지정**: 함수 계약 정의 (`@spec`)

### ❌ DON'T
- **상태 변경**: Tuple/Map 반환으로 새로운 상태 생성
- **에러 처리 생략**: `{:ok, result} | {:error, reason}` 패턴 활용
- **OTP 무시**: 재구현 대신 표준 라이브러리 사용
- **깊은 중첩**: 패턴 매칭으로 가독성 향상
- **프로세스 과다 생성**: 풀 구조 활용


## Works Well With

- `moai-domain-backend` (백엔드 아키텍처)
- `moai-domain-devops` (배포 및 모니터링)
- `moai-domain-database` (데이터 모델링)


**Version**: 2.0.0 | **Last Updated**: 2025-11-21 | **Lines**: 255
