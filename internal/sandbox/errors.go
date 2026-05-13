package sandbox

import "fmt"

var ErrSandboxBackendUnavailable = fmt.Errorf("sandbox backend unavailable")
var ErrSandboxProfileInvalid = fmt.Errorf("sandbox profile invalid")
var ErrSandboxRequired = fmt.Errorf("sandbox required")
var ErrSandboxOutputTruncated = fmt.Errorf("sandbox output truncated")
var ErrSandboxSetuidDenied = fmt.Errorf("setuid denied")

func IsSandboxBackendUnavailable(err error) bool {
	return err != nil && err == ErrSandboxBackendUnavailable
}

func IsSandboxOutputTruncated(err error) bool {
	return err != nil && err == ErrSandboxOutputTruncated
}

func IsSandboxSetuidDenied(err error) bool {
	return err != nil && err == ErrSandboxSetuidDenied
}
