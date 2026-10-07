package bugreport

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Detail is the closed union the optional payload field admits (REQ-ANON-011).
// The interface is unexported-sealed: only the two concrete carriers below —
// and nothing a caller can define — satisfy it, so `detail` is an enum-like
// dedicated type on every path, never a free string.
type Detail interface {
	isDetail()

	// Token returns the canonical string form used on the wire and read back
	// through ParseDetail. Only closed carriers implement it, so the wire
	// form is closed-set membership by construction.
	Token() string
}

// TemplateToken is the closed six-member token set for the template and
// harness kinds. internal/cli maps template sentinels to these tokens at the
// emit sites (design.md section 2); the token determines both the kind it may
// ride on and the verdict the attribution applies.
type TemplateToken string

const (
	TokenPathTraversal     TemplateToken = "path_traversal"
	TokenNotFound          TemplateToken = "not_found"
	TokenPreserveIntegrity TemplateToken = "preserve_integrity"
	TokenMissingKey        TemplateToken = "missing_key"
	TokenUnexpandedToken   TemplateToken = "unexpanded_token"
	TokenInvalidJSON       TemplateToken = "invalid_json"
)

// allTokens is the closed set in pinned order (TestTemplateTokenSetIsExactlySix).
var allTokens = []TemplateToken{
	TokenPathTraversal,
	TokenNotFound,
	TokenPreserveIntegrity,
	TokenMissingKey,
	TokenUnexpandedToken,
	TokenInvalidJSON,
}

// AllTemplateTokens returns the six tokens in pinned order.
func AllTemplateTokens() []TemplateToken {
	out := make([]TemplateToken, len(allTokens))
	copy(out, allTokens)
	return out
}

// Valid reports membership in the closed set.
func (t TemplateToken) Valid() bool {
	for _, member := range allTokens {
		if t == member {
			return true
		}
	}
	return false
}

// ValidForKind reports whether the token may ride on kind k, per the register
// rows: template_deploy_failure takes the deploy sentinels, harness_defect
// the render/validate ones, and no other kind takes a token at all.
func (t TemplateToken) ValidForKind(k Kind) bool {
	if !t.Valid() || !k.Valid() {
		return false
	}
	switch k {
	case KindTemplateDeployFailure:
		return t == TokenPathTraversal || t == TokenNotFound || t == TokenPreserveIntegrity
	case KindHarnessDefect:
		return t == TokenMissingKey || t == TokenUnexpandedToken || t == TokenInvalidJSON
	default:
		return false
	}
}

func (t TemplateToken) Token() string { return string(t) }
func (t TemplateToken) isDetail()     {}

// HookDetail is the closed carrier for the hook kinds: a registered event
// paired with a registered handler name. The identifiers are obtainable only
// through registration (RegisterHookIdentity), which shape-checks both names,
// so no "/", ".", or path-shaped string can enter a payload through a hook
// detail.
type HookDetail struct {
	EventID   string
	HandlerID string
}

func (h HookDetail) Token() string { return "event=" + h.EventID + " handler=" + h.HandlerID }
func (h HookDetail) isDetail()     {}

// identityNamePattern is the shape every registered name must match:
// an ASCII letter, then up to 63 ASCII letters or digits. No "/", ".",
// "_", ":", space, or leading digit can pass, which is what keeps a path
// shape out of the payload through a hook identity (design.md section 4).
var identityNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{0,63}$`)

// hookIdentityRegistry holds the registered events and handler names. It is
// written once at package initialisation by internal/hook (M3 wiring) and read
// by ParseDetail and the internal/cli guard; the mutex serialises the writes
// tests make. This is a write-once registry, not mutable pipeline state.
var hookIdentityRegistry = struct {
	sync.Mutex
	events   map[string]bool
	handlers map[string]bool
}{
	events:   map[string]bool{},
	handlers: map[string]bool{},
}

// registerName validates the shape and inserts the name into the given set.
func registerName(set map[string]bool, name, what string) error {
	if !identityNamePattern.MatchString(name) {
		return fmt.Errorf("bugreport: refusing %s name %q: must match %s", what, name, identityNamePattern)
	}
	set[name] = true
	return nil
}

// RegisterEvent registers one hook event name. internal/hook calls it at
// package initialisation for each of its EventType constants; the parser guard
// in internal/cli asserts the registered set equals the declared constants.
func RegisterEvent(event string) error {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	return registerName(hookIdentityRegistry.events, event, "event")
}

// RegisterHandlerName registers one production handler name. internal/hook
// calls it at package initialisation for each handler its production wiring
// registers; the guard TestEveryRegisteredHookHandlerHasBugreportName asserts
// the wiring and the table agree.
func RegisterHandlerName(name string) error {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	return registerName(hookIdentityRegistry.handlers, name, "handler")
}

// RegisterHookIdentity returns the HookDetail for a registered event and a
// registered handler, refusing either unregistered or malformed name. It is
// the only way a HookDetail comes to exist.
func RegisterHookIdentity(event, handler string) (HookDetail, error) {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	if !identityNamePattern.MatchString(event) {
		return HookDetail{}, fmt.Errorf("bugreport: refusing event name %q: must match %s", event, identityNamePattern)
	}
	if !identityNamePattern.MatchString(handler) {
		return HookDetail{}, fmt.Errorf("bugreport: refusing handler name %q: must match %s", handler, identityNamePattern)
	}
	if !hookIdentityRegistry.events[event] {
		return HookDetail{}, fmt.Errorf("bugreport: event %q is not registered", event)
	}
	if !hookIdentityRegistry.handlers[handler] {
		return HookDetail{}, fmt.Errorf("bugreport: handler %q is not registered", handler)
	}
	return HookDetail{EventID: event, HandlerID: handler}, nil
}

// EventRegistered reports whether the event name is in the table.
func EventRegistered(event string) bool {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	return hookIdentityRegistry.events[event]
}

// HandlerRegistered reports whether the handler name is in the table.
func HandlerRegistered(name string) bool {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	return hookIdentityRegistry.handlers[name]
}

// RegisteredEventCount / RegisteredHandlerCount report the table sizes for the
// internal/cli guard.
func RegisteredEventCount() int {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	return len(hookIdentityRegistry.events)
}

func RegisteredHandlerCount() int {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	return len(hookIdentityRegistry.handlers)
}

// RegisteredEventNames returns the registered event names, sorted, for the
// internal/cli guard that asserts the table equals internal/hook's declared
// EventType constants.
func RegisteredEventNames() []string {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	out := make([]string, 0, len(hookIdentityRegistry.events))
	for name := range hookIdentityRegistry.events {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// RegisteredHandlerNames returns the registered handler names, sorted, for
// the same guard's wiring-vs-table set equality.
func RegisteredHandlerNames() []string {
	hookIdentityRegistry.Lock()
	defer hookIdentityRegistry.Unlock()
	out := make([]string, 0, len(hookIdentityRegistry.handlers))
	for name := range hookIdentityRegistry.handlers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// ErrDetailRejected is the sentinel every detail rejection wraps, so a caller
// can distinguish "this detail is not in the closed set" from an unexpected
// storage error.
var ErrDetailRejected = errors.New("bugreport: detail is outside the closed set for this kind")

// ParseDetail is the read-back validator (REQ-ANON-011): a queued payload's
// detail field is untrusted input, and this re-validates it by closed-set
// membership before anything downstream sees it.
//
// For the template and harness kinds s must be one of the six tokens, valid
// for the kind. For the hook kinds s must be the canonical "event=<E>
// handler=<H>" form with both names registered. The panic and internal_error
// kinds accept no detail at all.
func ParseDetail(kind Kind, s string) (Detail, error) {
	if !kind.Valid() {
		return nil, fmt.Errorf("%w: unknown kind %q", ErrDetailRejected, kind)
	}
	switch kind {
	case KindPanic, KindInternalError:
		return nil, fmt.Errorf("%w: kind %s takes no detail", ErrDetailRejected, kind)
	case KindTemplateDeployFailure, KindHarnessDefect:
		tok := TemplateToken(s)
		if !tok.ValidForKind(kind) {
			return nil, fmt.Errorf("%w: token %q is not valid for kind %s", ErrDetailRejected, s, kind)
		}
		return tok, nil
	case KindHookHandlerFailure, KindHookTimeout:
		event, handler, ok := parseHookToken(s)
		if !ok {
			return nil, fmt.Errorf("%w: %q is not the canonical hook detail form", ErrDetailRejected, s)
		}
		hookIdentityRegistry.Lock()
		defer hookIdentityRegistry.Unlock()
		if !hookIdentityRegistry.events[event] {
			return nil, fmt.Errorf("%w: event %q is not registered", ErrDetailRejected, event)
		}
		if !hookIdentityRegistry.handlers[handler] {
			return nil, fmt.Errorf("%w: handler %q is not registered", ErrDetailRejected, handler)
		}
		return HookDetail{EventID: event, HandlerID: handler}, nil
	default:
		return nil, fmt.Errorf("%w: unknown kind %q", ErrDetailRejected, kind)
	}
}

// parseHookToken parses the canonical "event=<E> handler=<H>" form strictly:
// exactly two space-separated key=value fields, event first.
func parseHookToken(s string) (event, handler string, ok bool) {
	var eventPart, handlerPart string
	n := 0
	for _, field := range strings.SplitN(s, " ", 3) {
		switch {
		case strings.HasPrefix(field, "event=") && n == 0:
			eventPart = strings.TrimPrefix(field, "event=")
			n = 1
		case strings.HasPrefix(field, "handler=") && n == 1:
			handlerPart = strings.TrimPrefix(field, "handler=")
			n = 2
		default:
			return "", "", false
		}
	}
	if n != 2 || eventPart == "" || handlerPart == "" || strings.Contains(handlerPart, " ") {
		return "", "", false
	}
	return eventPart, handlerPart, true
}
