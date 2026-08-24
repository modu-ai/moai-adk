; astx — TypeScript symbol captures.
(function_declaration name: (identifier) @symbol.function)
(method_definition name: (property_identifier) @symbol.method)
(class_declaration name: (type_identifier) @symbol.type)
(interface_declaration name: (type_identifier) @symbol.type)

; Call/import captures (name-based grade).
(function_declaration name: (identifier) @code.caller) @func.node
(method_definition name: (property_identifier) @code.caller) @func.node
(call_expression function: (identifier) @code.call)
(call_expression function: (member_expression property: (property_identifier) @code.call))
(import_statement source: (string) @code.import)
