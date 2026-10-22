package privatefile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/testguard"
)

func TestWriteFilePostPublicationRefusalRevokesSecret(t *testing.T) {
	for _, scenario := range []string{"policy-refused", "policy-unreadable", "concurrent-replacement"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "credential")
			owned := path
			const secret = "SYNTHETIC_NEW_API_KEY=must-not-survive-refusal"
			const successor = "unrelated successor must remain byte-identical"
			policyErr := errors.New("injected native policy read failure")
			checkedPublished := false
			err := writeFileWithCheck(path, []byte(secret), func(candidate string) error {
				if candidate != path {
					return Check(candidate)
				}
				checkedPublished = true
				got, err := publicationReadbackForTest(path)
				if err != nil || string(got) != secret {
					t.Fatalf("fixture did not reach real publication: payload=%q err=%v", got, err)
				}
				switch scenario {
				case "policy-refused":
					testguard.MakeFilePublic(t, path)
					policyErr = Check(path)
					if policyErr == nil {
						t.Fatal("fixture must actually fail native policy readback")
					}
				case "concurrent-replacement":
					owned = filepath.Join(dir, "moved-publication")
					if err := os.Rename(path, owned); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte(successor), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				return policyErr
			})
			if !checkedPublished || !errors.Is(err, policyErr) {
				t.Fatalf("published policy error was lost: checked=%v error=%v", checkedPublished, err)
			}
			if data, readErr := os.ReadFile(owned); readErr == nil {
				if len(data) != 0 || strings.Contains(string(data), secret) {
					t.Errorf("refused publication retained secret bytes: %q", data)
				}
			} else if !os.IsNotExist(readErr) {
				t.Fatalf("read refused publication: %v", readErr)
			}
			if scenario == "concurrent-replacement" {
				data, readErr := os.ReadFile(path)
				if readErr != nil || string(data) != successor {
					t.Fatalf("another writer's replacement was changed: payload=%q err=%v", data, readErr)
				}
			}
		})
	}
}
