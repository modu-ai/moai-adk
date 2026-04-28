; McCabe decision nodes for Go source files.
; Each captured node adds 1 to the cyclomatic complexity count.

[
  (if_statement)
  (for_statement)
  (expression_case)
  (default_case)
  (communication_case)
  (type_case)
  (binary_expression operator: "&&")
  (binary_expression operator: "||")
] @decision

; IfBranches: count if_statement nodes separately.
(if_statement) @if_branch
