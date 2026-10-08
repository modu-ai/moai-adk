package hook

// handler_names.go — the capture package's registration feed
// (SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-011, design.md section 4).
//
// bugreport's hook-detail carrier admits only a registered event paired
// with a registered handler name, and registration shape-checks both names
// so no "/", ".", or path-shaped string can enter a payload. internal/hook
// owns the production identity tables: this file registers, at package
// initialisation, every EventType constant the package declares and the
// type name of every handler the production wiring (internal/cli/deps.go)
// registers.
//
// The cli-side guard TestEveryRegisteredHookHandlerHasBugreportName asserts
// set equality between THIS list and the wiring, in both directions: a
// handler wired without a registration fails CI, and so does a registration
// left behind by a removed handler.

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// bugreportHandlerTypeNames are the concrete handler types the production
// wiring registers. WithEscalationConfig mutates the inner handler in place
// and returns it unchanged, so a wrapped handler's %T identity is the inner
// type's — the wrapped names below appear once.
var bugreportHandlerTypeNames = []string{
	"sessionStartHandler",
	"sessionEndHandler",
	"autoUpdateHandler",
	"handoffInjectHandler",
	"sessionStartCompactHandler",
	"stopHandler",
	"preToolHandler",
	"postToolHandler",
	"postToolGuardianHandler",
	"compactHandler",
	"postToolUseFailureHandler",
	"notificationHandler",
	"subagentStartHandler",
	"userPromptSubmitHandler",
	"permissionRequestHandler",
	"teammateIdleHandler",
	"taskCompletedHandler",
	"worktreeCreateHandler",
	"worktreeRemoveHandler",
	"postCompactHandler",
	"instructionsLoadedHandler",
	"stopFailureHandler",
	"subagentStopHandler",
	"taskCreatedHandler",
	"permissionDeniedHandler",
	"configChangeHandler",
	"cwdChangedHandler",
	"fileChangedHandler",
	"elicitationHandler",
	"elicitationResultHandler",
}

// BugreportHandlerName returns the identity a handler carries in the
// capture package's table: its Go type name with the package qualifier and
// pointer marker stripped — the only identity the Handler interface offers
// (types.go:624 has no name method, so %T is what there is).
func BugreportHandlerName(h Handler) string {
	s := fmt.Sprintf("%T", h)
	s = strings.TrimPrefix(s, "*")
	if i := strings.LastIndexByte(s, '/'); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[i+1:]
	}
	return s
}

// bugreportDetail builds the closed detail for a live dispatch, or nil when
// the event/handler pair is not registered. Nil means the capture call below
// drops the signal — fail-closed in the safe direction: no capture rather
// than an unvalidated identity. The cli guard keeps the table honest so a
// wired-but-unregistered handler is a loud CI failure, not silence.
func bugreportDetail(event EventType, h Handler) bugreport.Detail {
	hd, err := bugreport.RegisterHookIdentity(string(event), BugreportHandlerName(h))
	if err != nil {
		return nil
	}
	return hd
}

func init() {
	// The declared set, in types.go declaration order. The cli parser guard
	// (TestEveryRegisteredHookHandlerHasBugreportName) asserts this equals
	// the EventType constants the file declares, so adding a constant there
	// without registering it here fails CI.
	for _, e := range []EventType{
		EventSessionStart,
		EventPreToolUse,
		EventPostToolUse,
		EventSessionEnd,
		EventStop,
		EventSubagentStop,
		EventPreCompact,
		EventPostToolUseFailure,
		EventNotification,
		EventSubagentStart,
		EventUserPromptSubmit,
		EventPermissionRequest,
		EventTeammateIdle,
		EventTaskCompleted,
		EventWorktreeCreate,
		EventWorktreeRemove,
		EventPostCompact,
		EventInstructionsLoaded,
		EventStopFailure,
		EventConfigChange,
		EventTaskCreated,
		EventCwdChanged,
		EventFileChanged,
		EventElicitation,
		EventElicitationResult,
		EventPermissionDenied,
		EventPostToolBatch,
		EventUserPromptExpansion,
		EventMessageDisplay,
		EventSetup,
	} {
		_ = bugreport.RegisterEvent(string(e))
	}
	for _, n := range bugreportHandlerTypeNames {
		_ = bugreport.RegisterHandlerName(n)
	}
}
