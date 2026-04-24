# Elixir violation fixture — demonstrates patterns matched by elixir rules.
# This file should produce >= 1 finding when scanned with the elixir rule set.

defmodule Example do
  def bad_code(config) do
    # Matches elixir-unused-var: assignment to _unused
    _unused = :some_value

    # Matches elixir-unsafe-map-access: Map.fetch! raises KeyError on missing key
    value = Map.fetch!(config, :required_key)

    # Matches elixir-todo-marker: stub raise
    raise "TODO"

    value
  end
end
