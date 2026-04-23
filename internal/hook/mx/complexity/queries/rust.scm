; McCabe decision nodes for Rust source files.
; Each captured node adds 1 to the cyclomatic complexity count.
; Note: `if let` and `while let` use `if_expression`/`while_expression` with
; a `let_condition` child — counted separately as decision nodes.

[
  (if_expression)
  (while_expression)
  (for_expression)
  (loop_expression)
  (match_arm)
  (let_condition)
  (binary_expression operator: "&&")
  (binary_expression operator: "||")
] @decision

; IfBranches: count if_expression nodes (includes if-let via let_condition parent).
(if_expression) @if_branch
