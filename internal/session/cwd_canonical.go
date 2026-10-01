package session

import (
	"os"
	"path/filepath"
)

// canonicalCWD reports dir's on-disk spelling where the platform can answer
// it (t1293): a case-insensitive filesystem resolves any letter-case variant
// to the same directory, so a logical PWD that differs only in case from the
// stored name would fork the registry's cwd axis and defeat every literal
// path comparison downstream (anchor detection, factory lane admission —
// t1290 F1 measured lanes registered as /Users/goos/moai/... while git and
// lsof name the same tree /Users/goos/MoAI/...). Fail-open: an unresolvable
// input returns dir unchanged — normalization must never block registration.
func canonicalCWD(dir string) string {
	if dir == "" {
		return dir
	}
	if resolved := realpathPlatform(dir); resolved != "" {
		return resolved
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return dir
}

// canonicalCWDFromProcess reports the current process working directory in
// its on-disk spelling. It is os.Getwd plus the canonicalCWD normalization.
func canonicalCWDFromProcess() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return canonicalCWD(wd), nil
}
