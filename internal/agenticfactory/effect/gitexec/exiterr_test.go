package gitexec

import (
	"strings"
	"testing"
)

func TestAnExitErrorNeverEchoesACredentialOrPersonalData(t *testing.T) {
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	err := &ExitError{Verb: "ls-remote", Code: 128, Stderr: "fatal: https://x:" + secret + "@github.com for someone@example.com in /Users/someone/work"}
	for _, bad := range []string{secret, "someone@example.com", "/Users/someone"} {
		if strings.Contains(err.Error(), bad) {
			t.Errorf("Error() holds %q: %s", bad, err.Error())
		}
	}
	if !strings.Contains(err.Error(), "ls-remote") || !strings.Contains(err.Error(), "exited 128") {
		t.Errorf("the verb and code are lost: %s", err.Error())
	}
}
