package review

import "context"

// FakeExecutor is a recording Executor for the tests in this package; it lives in a _test.go file so it is not shipped.
type FakeExecutor struct {
	Result Result
	Err    error
	Calls  []Request
}

// Execute records the request and returns the canned result.
func (f *FakeExecutor) Execute(_ context.Context, req Request) (Result, error) {
	f.Calls = append(f.Calls, req)
	return f.Result, f.Err
}
