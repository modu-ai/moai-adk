; astx — Java symbol captures.
(method_declaration name: (identifier) @symbol.method)
(class_declaration name: (identifier) @symbol.type)
(interface_declaration name: (identifier) @symbol.type)
(enum_declaration name: (identifier) @symbol.type)

; Call/import captures (name-based grade).
(method_declaration name: (identifier) @code.caller) @func.node
(constructor_declaration name: (identifier) @code.caller) @func.node
(method_invocation name: (identifier) @code.call)
(import_declaration (scoped_identifier) @code.import)
