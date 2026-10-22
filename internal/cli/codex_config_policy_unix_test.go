//go:build !windows

package cli

import (
	"fmt"
	"os"
	"testing"
)

func codexConfigPolicyFingerprint(t *testing.T, path string) string {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("mode=%04o", info.Mode().Perm())
}
