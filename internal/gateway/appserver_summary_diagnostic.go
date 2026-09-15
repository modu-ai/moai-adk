package gateway

// AppServerSummaryDiagnostic contains only request-shape observations. It is not
// part of a provider request or public response and carries no header values.
type AppServerSummaryDiagnostic struct {
	Classified, AgentHeaderPresent, PromptMatch bool
	LastRole, LastContent                       string
}

type appServerDiagnosticError struct {
	err        error
	diagnostic AppServerSummaryDiagnostic
}

func (e *appServerDiagnosticError) Error() string { return e.err.Error() }
func (e *appServerDiagnosticError) Unwrap() error { return e.err }
