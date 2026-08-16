package fixtures

// Valid Go fixture — a clean sample produces zero findings against the
// shipped ruleset (.moai/astgrep-rules via sgconfig.yml).
//
// It deliberately avoids every shipped matcher: no interface{} (use any),
// no time.Now().Sub, no string([]byte{...}), no blank-error assignment,
// no bare `return err`, no sentinel errors.New var, no bare go func literal,
// no defer-in-loop, no unguarded channel send, no unchecked map lookup
// assignment, no md5.New, no shell exec, no template.HTML, no log.Printf,
// no credential-looking literals, and no http handler shape.

import "fmt"

// Clean wraps an error with context using the %w verb and returns it.
func Clean(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
