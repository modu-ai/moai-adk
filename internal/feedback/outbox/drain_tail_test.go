package outbox

// The oversized-log test (review gate finding, P2): a log grown past the
// cap made the sent-history read answer "no history" for EVERYTHING — a
// fingerprint already sent inside the dedupe window re-queued and would
// publish again. The read keeps the log's TAIL instead: recent rows are
// the part the window can still suppress, older rows fall outside the
// window check anyway, and a row cut in half at the read offset fails JSON
// parsing like any malformed line. The read is also capped at one cap-
// sized allocation, never the file's full size.

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSentHistoryKeepsTheTailPastTheCap(t *testing.T) {
	spoolFixture(t)
	consentOn(t)

	logPath, err := StorePath(OutboxFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	old := time.Now().AddDate(-1, 0, 0).UTC().Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)

	// Filler rows: valid JSON, but a year old — outside every dedupe
	// window, so their suppression verdict is false either way.
	filler := fmt.Sprintf(`{"outcome":"sent","fingerprint":"%s","at":"%s"}`+"\n", strings.Repeat("b", 32), old)
	var b strings.Builder
	for b.Len() <= 1024*1024+4096 {
		b.WriteString(filler)
	}
	b.WriteString(fmt.Sprintf(`{"outcome":"sent","fingerprint":"%s","at":"%s"}`+"\n", strings.Repeat("c", 32), recent))
	if werr := os.WriteFile(logPath, []byte(b.String()), 0o600); werr != nil {
		t.Fatalf("seed oversized log: %v", werr)
	}

	if got := SentHistoryHasFingerprintWithin(strings.Repeat("c", 32), time.Now(), 7); !got {
		t.Fatal("the most recent sent row was lost to the size cap — an in-window fingerprint would re-queue and publish again")
	}
	// The filler's year-old rows are outside the window: not suppressed,
	// before or after the tail cut.
	if got := SentHistoryHasFingerprintWithin(strings.Repeat("b", 32), time.Now(), 7); got {
		t.Fatal("a year-old filler row suppressed a fingerprint — the window check must bound the tail read's answers too")
	}
}
