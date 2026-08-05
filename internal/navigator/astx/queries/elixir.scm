; astx — Elixir symbol captures. def/defp/defmodule are call expressions.
(call
  target: (identifier) @_kw
  (arguments (alias) @symbol.type)
  (#match? @_kw "^defmodule$"))
(call
  target: (identifier) @_kw
  (arguments (call target: (identifier) @symbol.function))
  (#match? @_kw "^(def|defp|defmacro)$"))
