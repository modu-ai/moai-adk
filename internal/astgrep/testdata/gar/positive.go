package gar

import "errors"

func doSomething() error { return errors.New("boom") }

// PositiveCase contains the one true violation: an unwrapped `return err`
// inside an `if err != nil` block.
func PositiveCase() error {
	err := doSomething()
	if err != nil {
		return err // WANT: matched as go-error-not-wrapped
	}
	return nil
}
