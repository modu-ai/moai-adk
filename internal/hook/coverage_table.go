// Resolution: KEEP — 27-event coverage table for audit_test and doctor hook CLI.
// SPEC-V3R2-RT-006 REQ-050, AC-12.
package hook

// EventResolution categorizes how each hook event is handled.
type EventResolution string

const (
	// ResolutionKeep means the handler has full business logic.
	ResolutionKeep EventResolution = "KEEP"
	// ResolutionUpgrade means the stub handler was upgraded with full logic.
	ResolutionUpgrade EventResolution = "UPGRADE"
	// ResolutionFix means the handler fixes a known bug.
	ResolutionFix EventResolution = "FIX"
	// ResolutionRetireObsOnly means removed from settings.json; kept as observability tap.
	ResolutionRetireObsOnly EventResolution = "RETIRE-OBS-ONLY"
	// ResolutionRemove means the orphan handler was deleted entirely.
	ResolutionRemove EventResolution = "REMOVE"
	// ResolutionComposite means the handler is bundled under another event (e.g., autoUpdate).
	ResolutionComposite EventResolution = "COMPOSITE"
)

// EventCoverageEntry describes a single hook event's coverage status.
type EventCoverageEntry struct {
	// EventName is the Claude Code hook event name (e.g., "SessionStart").
	EventName string
	// Resolution is the v3 resolution category.
	Resolution EventResolution
	// IsActive is true when the event is registered in settings.json.
	IsActive bool
	// ObservabilityOptIn is true when the event is listed in system.yaml hook.observability_events.
	ObservabilityOptIn bool
	// HandlerFile is the Go source file implementing the handler.
	HandlerFile string
}

// CoverageTable is the authoritative 26-event table after EventSetup retirement
// (SPEC-V3R2-MIG-002 M2.1). Originally 27 events per SPEC-V3R2-RT-006 §5.7.
// @MX:ANCHOR: [AUTO] CoverageTable is the authoritative 26-event inventory post-MIG-002
// @MX:REASON: fan_in=3, consumed by audit_test/doctor_hook CLI/doctor hook subcommand
var CoverageTable = []EventCoverageEntry{
	{EventName: "SessionStart", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "session_start.go"},
	{EventName: "SessionEnd", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "session_end.go"},
	{EventName: "PreToolUse", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "pre_tool.go"},
	{EventName: "PostToolUse", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "post_tool.go"},
	{EventName: "PostToolUseFailure", Resolution: ResolutionUpgrade, IsActive: true, HandlerFile: "post_tool_failure.go"},
	{EventName: "PreCompact", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "compact.go"},
	{EventName: "PostCompact", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "post_compact.go"},
	{EventName: "Stop", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "stop.go"},
	{EventName: "StopFailure", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "stop_failure.go"},
	{EventName: "SubagentStart", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "subagent_start.go"},
	{EventName: "SubagentStop", Resolution: ResolutionFix, IsActive: true, HandlerFile: "subagent_stop.go"},
	{EventName: "Notification", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: "notification.go"},
	{EventName: "UserPromptSubmit", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "user_prompt_submit.go"},
	{EventName: "PermissionRequest", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "permission_request.go"},
	{EventName: "PermissionDenied", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "permission_denied.go"},
	{EventName: "TeammateIdle", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "teammate_idle.go"},
	{EventName: "TaskCompleted", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "task_completed.go"},
	{EventName: "TaskCreated", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: "task_created.go"},
	{EventName: "WorktreeCreate", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "worktree_create.go"},
	{EventName: "WorktreeRemove", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "worktree_remove.go"},
	{EventName: "ConfigChange", Resolution: ResolutionUpgrade, IsActive: true, HandlerFile: "config_change.go"},
	{EventName: "CwdChanged", Resolution: ResolutionKeep, IsActive: true, HandlerFile: "cwd_changed.go"},
	{EventName: "FileChanged", Resolution: ResolutionUpgrade, IsActive: true, HandlerFile: "file_changed.go"},
	{EventName: "InstructionsLoaded", Resolution: ResolutionUpgrade, IsActive: true, HandlerFile: "instructions_loaded.go"},
	{EventName: "Elicitation", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: "elicitation.go"},
	{EventName: "ElicitationResult", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: "elicitation.go"},
	// Setup event retired: EventSetup constant + CLI binding removed by SPEC-V3R2-MIG-002 M2.1.
	// The event had no handler implementation; only an orphan constant + cobra subcommand.
	// Composite: autoUpdate is bundled under SessionStart in settings.json.
	{EventName: "AutoUpdate (SessionStart composite)", Resolution: ResolutionComposite, IsActive: true, HandlerFile: "auto_update.go"},
}

// CoverageSummary holds aggregate counts from the CoverageTable.
type CoverageSummary struct {
	Total           int `json:"total"`
	Keep            int `json:"keep"`
	Upgrade         int `json:"upgrade"`
	Fix             int `json:"fix"`
	RetireObsOnly   int `json:"retire"`
	Remove          int `json:"remove"`
	Composite       int `json:"composite"`
}

// Summarize computes aggregate counts from the CoverageTable.
func Summarize() CoverageSummary {
	var s CoverageSummary
	for _, e := range CoverageTable {
		s.Total++
		switch e.Resolution {
		case ResolutionKeep:
			s.Keep++
		case ResolutionUpgrade:
			s.Upgrade++
		case ResolutionFix:
			s.Fix++
		case ResolutionRetireObsOnly:
			s.RetireObsOnly++
		case ResolutionRemove:
			s.Remove++
		case ResolutionComposite:
			s.Composite++
		}
	}
	return s
}
