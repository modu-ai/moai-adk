package outbox

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
)

// A legacy broad log must become private before a new record can be appended;
// preserving prior records is as important as the native privacy boundary.
func TestAppendOutboxNarrowsExistingLogWithoutTruncation(t *testing.T) {
	spoolFixture(t)
	if err := AppendOutbox(OutboxRow{Outcome: "queued", Title: "first", Body: "first body"}); err != nil {
		t.Fatal(err)
	}
	path, err := StorePath(OutboxFileName)
	if err != nil {
		t.Fatal(err)
	}
	testguard.MakeFilePublic(t, path)
	if err := privatefile.Check(path); err == nil {
		t.Fatal("fixture must begin publicly readable")
	}
	if err := AppendOutbox(OutboxRow{Outcome: "sent", Title: "second", Body: "second body"}); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(path); err != nil {
		t.Fatalf("append left a broad log: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(strings.TrimSpace(string(raw)), "\n") != 1 || !strings.Contains(string(raw), `"title":"first"`) || !strings.Contains(string(raw), `"title":"second"`) {
		t.Fatalf("append did not preserve both records: %s", raw)
	}
}
