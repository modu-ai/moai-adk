; McCabe decision nodes for Python source files.
; Each captured node adds 1 to the cyclomatic complexity count.

[
  (if_statement)
  (elif_clause)
  (for_statement)
  (while_statement)
  (except_clause)
  (boolean_operator operator: "and")
  (boolean_operator operator: "or")
  (conditional_expression)
] @decision

; IfBranches: count if_statement and elif_clause nodes separately.
[
  (if_statement)
  (elif_clause)
] @if_branch
