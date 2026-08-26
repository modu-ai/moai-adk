; astx — Rust symbol captures.
(function_item name: (identifier) @symbol.function)
(struct_item name: (type_identifier) @symbol.type)
(enum_item name: (type_identifier) @symbol.type)
(trait_item name: (type_identifier) @symbol.type)

; Call/import captures (name-based grade).
(function_item name: (identifier) @code.caller) @func.node
(call_expression function: (identifier) @code.call)
(call_expression function: (scoped_identifier name: (identifier) @code.call))
(use_declaration argument: (scoped_identifier) @code.import)
