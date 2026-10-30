package core

// Roles whose evidence the gate counts (design 11). The names are provisional:
// SPEC-AGENTIC-FACTORY-IDENTITY-001 defines what a role is and how a session
// obtains one. Until that SPEC is complete, no attribution source exists.
const (
	RoleAgent       = "agent" // the session that executed the attempt (the writer)
	RoleReviewer    = "reviewer"
	RoleAuditor     = "auditor"
	RoleQA          = "qa"
	RoleCoordinator = "coordinator"
)

// AttributionLocked is the reason an attribution lookup gives while the
// identity SPEC is incomplete (DEC-AF-02): every review, audit, QA and
// coordinator result is unattributable, so none of it is counted.
const AttributionLocked = "identity-spec-incomplete"
