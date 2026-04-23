; McCabe decision nodes for TypeScript source files.
; Each captured node adds 1 to the cyclomatic complexity count.

[
  (if_statement)
  (for_statement)
  (for_in_statement)
  (while_statement)
  (do_statement)
  (switch_case)
  (catch_clause)
  (binary_expression operator: "&&")
  (binary_expression operator: "||")
  (binary_expression operator: "??")
  (ternary_expression)
] @decision

; IfBranches: count if_statement nodes separately.
(if_statement) @if_branch
