//go:build !darwin

package mergeq

// clearExtendedACL does nothing off macOS. What a Linux default ACL or a native Windows ACL makes of a 0600 file in
// a 0700 directory has not been measured for this log, so nothing here claims the log is free of them; the log's mode
// and its directory are what this code sets, and the doc comment of FailureLog says so.
func clearExtendedACL(string) error { return nil }
