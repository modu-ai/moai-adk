package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureManifestIntegrity(t *testing.T) {
	dir := "testdata/validation-capture/claude-code-2.1.268"
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		File           string
		OriginalSHA256 string   `json:"original_sha256"`
		FixtureSHA256  string   `json:"fixture_sha256"`
		ChangedFields  []string `json:"changed_fields"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("manifest count: %d", len(rows))
	}
	for _, row := range rows {
		body, err := os.ReadFile(filepath.Join(dir, row.File))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(body)
		if hex.EncodeToString(sum[:]) != row.FixtureSHA256 {
			t.Fatalf("fixture hash mismatch: %s", row.File)
		}
		original, err := hex.DecodeString(row.OriginalSHA256)
		if err != nil || len(original) != 32 || len(row.ChangedFields) == 0 {
			t.Fatalf("invalid provenance: %s", row.File)
		}
		for _, marker := range []string{"/Users/", "/private/var/folders/", "/tmp/", "fe251a486f019c671"} {
			if strings.Contains(string(body), marker) {
				t.Fatalf("local identifier remains in %s", row.File)
			}
		}
		var parsed struct {
			Metadata struct {
				UserID string `json:"user_id"`
			}
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatal(err)
		}
		var user map[string]string
		if err := json.Unmarshal([]byte(parsed.Metadata.UserID), &user); err != nil {
			t.Fatal(err)
		}
		if user["device_id"] != "fixture-device_id" || user["session_id"] != "fixture-session_id" {
			t.Fatalf("metadata not anonymized: %s", row.File)
		}
	}
}
