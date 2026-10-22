//go:build !windows

package privatefile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAppendRefusesHardlinkWithoutChangingExternalAsset(t *testing.T) {
	for _, mode := range []os.FileMode{0o644, 0o600} {
		t.Run(fmt.Sprintf("mode-%04o", mode), func(t *testing.T) {
			dir := t.TempDir()
			outside := filepath.Join(dir, "external-asset")
			alias := filepath.Join(dir, "outbox.log")
			const payload = "user-owned content must remain unchanged"
			if err := os.WriteFile(outside, []byte(payload), mode); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(outside, mode); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(outside, alias); err != nil {
				t.Fatal(err)
			}
			opened, err := OpenAppend(alias)
			if opened != nil {
				if closeErr := opened.Close(); closeErr != nil {
					t.Fatal(closeErr)
				}
			}
			if err == nil {
				t.Error("multiply linked external asset was admitted for append")
			}
			info, statErr := os.Stat(outside)
			if statErr != nil {
				t.Fatal(statErr)
			}
			if got := info.Mode().Perm(); got != mode {
				t.Errorf("external mode changed: got %04o want %04o", got, mode)
			}
			data, readErr := os.ReadFile(outside)
			if readErr != nil || string(data) != payload {
				t.Fatalf("external payload changed: data=%q err=%v", data, readErr)
			}
		})
	}
}
