package uikit

import "github.com/modu-ai/moai-adk/internal/settings"

// TuiLabel holds a field's TUI widget title + description.
//
// Moved from internal/cli/schema_bridge.go as part of the uikit kernel extraction
// (SPEC-CLI-UIKIT-KERNEL-001). The profileSetupText-coupled bridge maps stay in
// package cli; this leaf package holds only the type + the resolver dispatch.
type TuiLabel struct {
	Title string
	Desc  string
}

// SchemaBridgeResolver resolves a schema key + locale to a TuiLabel.
//
// Registered by package cli at init() (the actual map lookup + profileSetupText
// coupling lives in cli/schema_bridge.go). This callback pattern keeps uikit
// free of profileSetupText references (REQ-CUK-007 b-ii resolution: uikit has
// zero profileSetupText refs, verified by AC-CUK-007).
type SchemaBridgeResolver func(schemaKey, locale string) (TuiLabel, bool)

// schemaBridgeResolver holds the cli-registered resolver. nil until cli's
// init() runs; package cli always runs init before any test or caller.
var schemaBridgeResolver SchemaBridgeResolver

// RegisterSchemaBridge registers the profileSetupText-aware resolver.
// Called by package cli init() exactly once at program start.
func RegisterSchemaBridge(fn SchemaBridgeResolver) {
	schemaBridgeResolver = fn
}

// SchemaKeyToTUIField resolves a schema field's i18n key to a TuiLabel for the
// given locale. The second return value reports whether the key has a bridge
// entry. Dispatches to the cli-registered resolver (b-ii split).
//
// Moved from internal/cli/schema_bridge.go (SPEC-CLI-UIKIT-KERNEL-001).
func SchemaKeyToTUIField(schemaKey, locale string) (TuiLabel, bool) {
	if schemaBridgeResolver != nil {
		return schemaBridgeResolver(schemaKey, locale)
	}
	return TuiLabel{}, false
}

// FieldDefTUILabel resolves a schema FieldDef's I18nKey to a TuiLabel.
//
// Moved from internal/cli/schema_bridge.go (SPEC-CLI-UIKIT-KERNEL-001).
func FieldDefTUILabel(f settings.FieldDef, locale string) (TuiLabel, bool) {
	return SchemaKeyToTUIField(f.I18nKey, locale)
}
