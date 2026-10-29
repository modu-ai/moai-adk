package cli

// checklist_plan.go — the group/item plan table of the init checklist
// (SPEC-TUI-CHECKLIST-001 design.md §4, REQ-TCL-010). The table is the one
// place that names the groups and items; call sites refer to the constants
// below, never to string literals, and each key resolves to a "checklist.*"
// catalogue entry (checklist.group.<group>, checklist.item.<group>.<item>).

const (
	clGroupTemplates = "templates"
	clGroupConfig    = "config"
	clGroupHooks     = "hooks"
	clGroupMCP       = "mcp"
	clGroupVerify    = "verify"

	clItemTemplatesDetect   = "detect"
	clItemTemplatesValidate = "validate"
	clItemTemplatesDeploy   = "deploy"

	clItemConfigAutonomy = "autonomy"
	clItemConfigProfile  = "profile"
	clItemConfigOptins   = "optins"
	clItemConfigHarness  = "harness"
	clItemConfigGlobals  = "globals"

	clItemHooksPrePush   = "prepush"
	clItemHooksPreCommit = "precommit"

	clItemMCPEntry = "entry"

	clItemVerifyManifest = "manifest"
)

// checklistPlanGroup is one titled group and its items in code-execution
// order.
type checklistPlanGroup struct {
	Key   string
	Items []string
}

// initChecklistPlan is the init plan table: five groups, thirteen items, the
// clean-run screen of design.md §1.
var initChecklistPlan = []checklistPlanGroup{
	{clGroupTemplates, []string{clItemTemplatesDetect, clItemTemplatesValidate, clItemTemplatesDeploy}},
	{clGroupConfig, []string{clItemConfigAutonomy, clItemConfigProfile, clItemConfigOptins, clItemConfigHarness, clItemConfigGlobals}},
	{clGroupHooks, []string{clItemHooksPrePush, clItemHooksPreCommit}},
	{clGroupMCP, []string{clItemMCPEntry}},
	{clGroupVerify, []string{clItemVerifyManifest}},
}
