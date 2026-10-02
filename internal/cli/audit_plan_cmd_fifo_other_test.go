//go:build !unix

package cli

// audit_plan_cmd_fifo_other_test.go — the non-unix counterpart of
// audit_plan_cmd_fifo_unix_test.go: a platform without Mkfifo reports the fixture
// as unsupported and the (j5) cases skip.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001

import "errors"

// mkfifoForTest reports that named pipes cannot be created here.
func mkfifoForTest(string) error {
	return errors.ErrUnsupported
}
