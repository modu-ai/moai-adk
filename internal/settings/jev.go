package settings

// jev.go — the ONE persistence path both Jev opt-in entrances drive
// (SPEC-JEV-OPTIN-MEASURE-001 REQ-JEVO-002).
//
// Two entrances exist: the `moai init` wizard question and the `moai web`
// settings screen. The console already reaches this package generically —
// parseSchemaForm collects the submitted bool into an edit map and hands it to
// ApplySchemaEdits. The wizard has no form, so it needs a named call; that call
// is SetJevEnabled, and it is deliberately a thin wrapper over the SAME
// ApplySchemaEdits rather than a second writer.
//
// The distinction matters because the write is a nested-key patch: the seam
// copies the whole section and mutates only the targeted nested field, so every
// sibling key and comment in workflow.yaml rides through byte-identical. A
// parallel writer would be a second implementation of that property, and the
// property is invisible until it breaks.

import "strconv"

// JevEnabledField is the schema field name of the Jev opt-in toggle. It is
// exported so the wizard names the same target the console renders; a test
// pins it against the FieldDef so the two cannot drift.
//
// The underlying key is workflow.jev.enabled in workflow.yaml — the SAME key
// the core capability's gate reads (internal/config WorkflowJevConfig). This
// SPEC introduces no new config key.
const JevEnabledField = "workflow.jev.enabled"

// SetJevEnabled persists the Jev opt-in for the project at projectRoot.
//
// It is the wizard's entrance. It performs no validation of its own: the value
// is a bool, the field is declared in the schema, and ApplySchemaEdits owns the
// value-invariant write gate (a submission that would not change the persisted
// value never reaches the file).
func SetJevEnabled(projectRoot string, enabled bool) error {
	return ApplySchemaEdits(projectRoot, map[string]string{
		JevEnabledField: strconv.FormatBool(enabled),
	})
}
