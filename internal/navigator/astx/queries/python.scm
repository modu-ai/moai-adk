; astx — Python symbol captures.
(function_definition name: (identifier) @symbol.function)
(class_definition name: (identifier) @symbol.type)

; Call/import captures (name-based grade).
(function_definition name: (identifier) @code.caller) @func.node
(call function: (identifier) @code.call)
(call function: (attribute attribute: (identifier) @code.call))
(import_statement name: (dotted_name) @code.import)
(import_from_statement module_name: (dotted_name) @code.import)
