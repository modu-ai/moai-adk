package bugreport

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// goFileShape matches a Go file path embedded in a frame name: an identifier
// ending in ".go" at a segment boundary, optionally followed by a line-number
// colon or another path segment. Anchored rather than a substring check, so
// function names like cli.goalProjectRoot are never misread as files.
var goFileShape = regexp.MustCompile(`(?:^|/)[A-Za-z0-9_\-]+\.go(?:$|[:/])`)

// ModulePrefix is the module path a stack frame's function name must begin
// with to be moai-internal (REQ-ANON-010). Frames carry the Function name
// only — file, line, and arguments are never read, because the build sets no
// -trimpath and a frame's file path would embed the builder's absolute paths
// (research.md section 4).
const ModulePrefix = "github.com/modu-ai/moai-adk/"

// capturePkgPrefix is this package's own prefix: the capture helpers sit
// between the recover site and the failing function in the stack, and a frame
// naming them carries no signal.
const capturePkgPrefix = ModulePrefix + "internal/bugreport."

// FrameLimit is the maximum number of moai-internal frames kept per signal,
// innermost first (design.md section 10). Resolved from internal/config so
// the ceiling has one owner.
func FrameLimit() int {
	return config.DefaultBugreportFrameLimit
}

// ValidateFrameName is the anchored allowlist a kept frame must pass: a bare
// Go symbol path — package path, dots, and the [...] generic placeholder the
// runtime prints for type arguments — and never a file path, line number, or
// argument text.
func ValidateFrameName(name string) error {
	if name == "" {
		return errors.New("bugreport: frame name is empty")
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "~") || strings.HasPrefix(name, "..") ||
		strings.Contains(name, "/../") {
		return fmt.Errorf("bugreport: frame %q is path-shaped", name)
	}
	// A Go FILE path in a frame is a path segment ending in ".go", followed
	// by end-of-string, a line-number colon, or another slash. The check is
	// anchored to the segment boundary on purpose: a substring containment
	// would misclassify ordinary function names that merely carry ".go"
	// inside them — cli.goalProjectRoot is a function, leak.go:42 is a file.
	if goFileShape.MatchString(name) {
		return fmt.Errorf("bugreport: frame %q carries a file-name shape", name)
	}
	for _, r := range name {
		allowed := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '.' || r == '/' || r == '_' || r == '*' || r == '(' || r == ')' ||
			r == '[' || r == ']' || r == ',' || r == '-'
		if !allowed {
			return fmt.Errorf("bugreport: frame %q carries character %q outside the symbol allowlist", name, r)
		}
	}
	return nil
}

// FilterFrames reduces a captured stack to the moai-internal function names
// (REQ-ANON-010): keep a frame only when its Function begins with the module
// prefix, strip the prefix, drop this package's own helper frames and the
// runtime's panic frame, keep at most FrameLimit innermost first. File, line,
// and argument fields are never read from the frame.
//
// @MX:NOTE: [AUTO] why file paths are never read — the build carries no
// -trimpath, so a frame's file path embeds the builder's absolute paths
// (design.md section 4). Function names only, ever.
func FilterFrames(frames []runtime.Frame) []string {
	limit := FrameLimit()
	out := make([]string, 0, limit)
	for _, f := range frames {
		name := f.Function
		if !strings.HasPrefix(name, ModulePrefix) {
			continue
		}
		if strings.HasPrefix(name, capturePkgPrefix) || name == "runtime.gopanic" {
			continue
		}
		stripped := strings.TrimPrefix(name, ModulePrefix)
		// Defense in depth: a module-prefixed name that fails the symbol
		// allowlist (a path or line number glued onto the prefix) is dropped
		// rather than passed to the validator downstream.
		if err := ValidateFrameName(stripped); err != nil {
			continue
		}
		out = append(out, stripped)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// FramesFromPC filters the stack a runtime.Callers slice describes. Capture
// call sites use this form; the runtime.Frame form exists for tests and for
// callers that already hold frames.
func FramesFromPC(pc []uintptr) []string {
	frames := runtime.CallersFrames(pc)
	limit := FrameLimit()
	out := make([]string, 0, limit)
	for {
		f, more := frames.Next()
		if name := f.Function; strings.HasPrefix(name, ModulePrefix) &&
			!strings.HasPrefix(name, capturePkgPrefix) && name != "runtime.gopanic" {
			if stripped := strings.TrimPrefix(name, ModulePrefix); ValidateFrameName(stripped) == nil {
				out = append(out, stripped)
				if len(out) >= limit {
					return out
				}
			}
		}
		if !more {
			break
		}
	}
	return out
}

// ErrZeroFramesStayLocal is refused by Build for a panic payload whose frame
// list is empty: a panic that yields no moai-internal frame stays local
// (REQ-ANON-010). Other kinds capture from a moai call site, so their empty
// frame lists are rejected as malformed input by Build's validators instead.
var ErrZeroFramesStayLocal = errors.New("bugreport: panic with no moai-internal frame stays local")

// IsLocalOnly reports whether a signal of kind k with the given frames must
// stay local by the zero-frame rule. Only the panic kind is subject to it.
func IsLocalOnly(k Kind, frames []string) bool {
	return k == KindPanic && len(frames) == 0
}
