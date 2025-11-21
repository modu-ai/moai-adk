---
name: moai-lang-elixir
description: Elixir functional programming, OTP patterns, and Phoenix framework best practices
allowed-tools: [Read, Bash, WebFetch]
---

# Elixir Language Expert

## Quick Reference

Elixir functional programming patterns, OTP (Open Telecom Platform) supervision trees, and Phoenix web framework for high-concurrency distributed systems.

**Core Concepts**:
- **Immutability**: All data structures immutable
- **Pattern Matching**: Function clauses and case expressions
- **Concurrency**: Actor model with lightweight processes
- **Fault Tolerance**: Supervision trees (let it crash philosophy)

**Ecosystem** (November 2025):
- **Phoenix 1.7**: Web framework
- **Ecto 3.11**: Database library
- **GenServer**: Generic server behavior
- **LiveView**: Real-time UI without JavaScript

---

## Implementation Guide

**GenServer Pattern**:
```elixir
defmodule Counter do
  use GenServer

  def start_link(initial) do
    GenServer.start_link(__MODULE__, initial)
  end

  def init(count) do
    {:ok, count}
  end

  def handle_call(:increment, _from, count) do
    {:reply, count + 1, count + 1}
  end
end
```

**Supervision Tree**:
```elixir
children = [
  {Counter, 0},
  {Phoenix.PubSub, name: MyApp.PubSub}
]

Supervisor.start_link(children, strategy: :one_for_one)
```

---

## Best Practices

### ✅ DO
- Use pattern matching extensively
- Design supervision hierarchies (let it crash)
- Leverage immutability for thread safety
- Use Ecto changesets for validation

### ❌ DON'T
- Mutate state (use processes instead)
- Skip error handling (use {:ok, result} | {:error, reason} tuples)
- Ignore OTP principles (reinventing the wheel)

---

**Version**: 1.0.0 | **Last Updated**: 2025-11-21
