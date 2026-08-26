; SPEC-PROJECT-NAVIGATOR-003 astx — Go symbol captures.
; Captures use the @symbol.<kind> naming convention so the extractor groups
; symbols by kind without language-specific Go logic.

(function_declaration name: (identifier) @symbol.function)

(method_declaration name: (field_identifier) @symbol.method)

(type_declaration (type_spec name: (type_identifier) @symbol.type))

; Call/import captures (name-based grade). One match pairs the enclosing
; declaration's name (@code.caller) with its full node (@func.node) so
; consumers join call sites to callers by line containment.
(function_declaration name: (identifier) @code.caller) @func.node
(method_declaration name: (field_identifier) @code.caller) @func.node
(call_expression function: (identifier) @code.call)
(call_expression function: (selector_expression field: (field_identifier) @code.call))
(import_spec path: (interpreted_string_literal) @code.import)
