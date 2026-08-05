; SPEC-PROJECT-NAVIGATOR-003 astx — Go symbol captures.
; Captures use the @symbol.<kind> naming convention so the extractor groups
; symbols by kind without language-specific Go logic.

(function_declaration name: (identifier) @symbol.function)

(method_declaration name: (field_identifier) @symbol.method)

(type_declaration (type_spec name: (type_identifier) @symbol.type))
