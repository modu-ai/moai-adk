// doctor_todo_ghost_test.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// REQ-TSP-042: the doctor ghost-inventory check's three states.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

func TestDoctorTodoGhostInventory(t *testing.T) {
	t.Run("no ghosts is OK", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvHome, t.TempDir())

		check := checkTodoGhostInventory(root, false)

		if check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK (message %q)", check.Status, check.Message)
		}
		if check.Name != todoGhostInventoryCheckName {
			t.Errorf("name = %q, want %q", check.Name, todoGhostInventoryCheckName)
		}
	})

	t.Run("ghosts present is WARN with path, class, and bytes", func(t *testing.T) {
		root, homeDB, _ := staleStoreFixture(t)
		// A home ghost json beside the engine database the base fixture
		// plants — the canonical dir carries the .db, so the json is a
		// leftover, exactly the measured 2026-09-29 artifact.
		ghostJSON := filepath.Join(filepath.Dir(homeDB), "backlog.json")
		if err := os.WriteFile(ghostJSON, []byte(strings.Repeat("g", 40)), 0o600); err != nil {
			t.Fatalf("plant ghost json: %v", err)
		}

		check := checkTodoGhostInventory(root, false)

		if check.Status != uikit.CheckWarn {
			t.Errorf("status = %v, want WARN (message %q)", check.Status, check.Message)
		}
		joined := check.Message + "\n" + check.Detail
		if !strings.Contains(joined, "backlog.json") {
			t.Errorf("inventory does not name the ghost path:\n%s\n%s", check.Message, check.Detail)
		}
		if !strings.Contains(joined, "legacy-json") {
			t.Errorf("inventory does not name the ghost class:\n%s\n%s", check.Message, check.Detail)
		}
		if !strings.Contains(joined, "40 bytes") {
			t.Errorf("inventory does not carry the byte size:\n%s\n%s", check.Message, check.Detail)
		}
	})

	t.Run("unreadable ghost artifact is FAIL", func(t *testing.T) {
		root, _, _ := staleStoreFixture(t)
		// A ghost path the detector lists but cannot stat as a regular
		// file: a directory named like a session record.
		weird := filepath.Join(root, ".moai", "state", "todo", "0a0b0c0d-0000-4000-8000-00000000dead.json")
		if err := os.MkdirAll(weird, 0o755); err != nil {
			t.Fatalf("plant non-regular session-record name: %v", err)
		}

		check := checkTodoGhostInventory(root, false)

		if check.Status != uikit.CheckFail {
			t.Errorf("status = %v, want FAIL (message %q)", check.Status, check.Message)
		}
		if !strings.Contains(check.Message+check.Detail, "unreadable") {
			t.Errorf("message does not carry the unreadable verdict: %q", check.Message)
		}
	})
}
