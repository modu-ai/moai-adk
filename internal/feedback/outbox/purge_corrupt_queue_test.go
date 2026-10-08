package outbox

// The corrupted-queue purge test (review gate finding, P2): the purge's
// queue step ran through the queue's mutation path, whose LOAD step fails
// on unparsable JSON — so a corrupted queue file ABORTED the purge, and the
// queue (with the removal of every store behind it) survived. The user
// asked for the data GONE: the removal proceeds regardless of the parse
// outcome. The lock the mutation path takes still serializes against an
// in-flight writer; only the parse verdict is ignored.

import (
	"os"
	"testing"
)

func TestPurgeRemovesACorruptedQueue(t *testing.T) {
	spoolFixture(t) // establishes the test's isolated MOAI_HOME
	consentOn(t)

	queuePath, err := StorePath(QueueFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := os.WriteFile(queuePath, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed corrupted queue: %v", err)
	}

	if err := PurgeStores(); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if _, serr := os.Lstat(queuePath); !os.IsNotExist(serr) {
		t.Fatalf("the corrupted queue file survived the purge (lstat: %v)", serr)
	}
}
