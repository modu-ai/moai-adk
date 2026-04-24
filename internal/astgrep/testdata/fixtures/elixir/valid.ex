# Elixir valid fixture — no rule violations.
# This file should produce 0 findings when scanned with the elixir rule set.

defmodule Example do
  def good_code do
    result = :some_value
    {:ok, value} = Map.fetch(%{key: "val"}, :key)
    value
  end

  def implemented do
    {:ok, "result"}
  end
end
