package factory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
)

func TestMigrationPublishesThroughPortablePrivateStaging(t *testing.T) {
	dir := privateBacklogFixtureDir(t)
	queue := filepath.Join(dir, "backlog.json")
	if err := os.WriteFile(queue, []byte(migrationFixture), 0600); err != nil {
		t.Fatal(err)
	}
	checksum := sha256File(t, queue)
	var expected BacklogRecord
	if err := json.Unmarshal([]byte(migrationFixture), &expected); err != nil {
		t.Fatal(err)
	}
	normalizeBacklogRecord(&expected)
	store := NewBacklogStore(queue)
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBacklogParity(&expected, got); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(store.EnginePath()); err != nil {
		t.Fatal(err)
	}
	if checksum != sha256File(t, queue+".migrated") {
		t.Fatal("legacy rollback bytes changed")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".backlog-migrate-") {
			t.Fatal("successful publication left private staging artifacts")
		}
	}
}
