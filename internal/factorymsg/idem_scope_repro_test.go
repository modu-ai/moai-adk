package factorymsg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// idemScopeEvidenceEnv names the directory the reproduction writes its
// evidence file into. The acceptance command passes it as a literal path
// relative to this package directory.
const idemScopeEvidenceEnv = "MOAI_T1100_EVIDENCE_DIR"

// idemScopeEvidence is the measured record. Field names are the keys the
// acceptance judge reads; do not rename them without updating that judge.
type idemScopeEvidence struct {
	Outcome           string   `json:"outcome"`
	UniqueConstraint  string   `json:"unique_constraint"`
	SenderSessions    []string `json:"sender_sessions"`
	SenderGenerations []int64  `json:"sender_generations"`
	FirstSendID       string   `json:"first_send_id"`
	SecondSendResult  string   `json:"second_send_result"`
	Rows              int      `json:"rows"`
	ClaimedIDs        int      `json:"claimed_ids"`
	ControlClaimed    int      `json:"control_claimed"`
	NotRunReasons     []string `json:"not_run_reasons,omitempty"`
}

var uniqueClause = regexp.MustCompile(`UNIQUE\s*\([^)]*\)`)

// storeUniqueConstraints reads the uniqueness rules the messages table was
// actually created with, from the database schema rather than from source.
func storeUniqueConstraints(t *testing.T, s *Store) string {
	t.Helper()
	rows, err := s.db.QueryContext(context.Background(), `SELECT sql FROM sqlite_master WHERE tbl_name='messages' AND sql IS NOT NULL ORDER BY type DESC, name`)
	if err != nil {
		t.Fatalf("read messages schema: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var found []string
	for rows.Next() {
		var sqlText string
		if err := rows.Scan(&sqlText); err != nil {
			t.Fatalf("scan messages schema: %v", err)
		}
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlText)), "CREATE UNIQUE INDEX") {
			found = append(found, strings.TrimSpace(sqlText))
			continue
		}
		found = append(found, uniqueClause.FindAllString(sqlText, -1)...)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate messages schema: %v", err)
	}
	return strings.Join(found, "; ")
}

// TestIdemScopeRestartReproduction measures, without asserting a preferred
// answer, whether a sender that restarts on the same lane slot (new session
// UUID, higher generation) and retries with the same idempotency key produces
// a second deliverable message row.
//
// With MOAI_T1100_EVIDENCE_DIR set, the evidence JSON is written there and a
// single tag line carrying its sha256 is printed to stdout. Without it, the
// same measurement runs and its path-traversal checks still fail the test, but
// no file or tag line is produced, so a plain package run stays green.
func TestIdemScopeRestartReproduction(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	s, err := Open(t.TempDir(), "run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	// Only the restarted sender process is live; the original sender process
	// is treated as gone, which is what a restart means.
	const restartedStart = "t1100-restarted-process-start"
	s.ownerCurrent = func(pid int, start string) bool { return pid == os.Getpid() && start == restartedStart }
	ctx := context.Background()

	sender := testPeer(s.runID, "t1100-sender-session-1", 1)
	sender.Slot = "lane-1"
	recipient := testPeer(s.runID, "t1100-recipient-session", 1)
	recipient.Slot = "lane-2"
	recipient.ProcessStart = restartedStart
	if sender, err = s.RegisterPeer(ctx, sender); err != nil {
		t.Fatalf("register sender: %v", err)
	}
	if recipient, err = s.RegisterPeer(ctx, recipient); err != nil {
		t.Fatalf("register recipient: %v", err)
	}

	const key = "t1100-idem-k"
	request := func(from Peer, idem string) SendRequest {
		return SendRequest{From: from, To: recipient, Kind: KindStatusReport, IdempotencyKey: idem, TaskRef: "t1100", CorrelationID: "c-t1100", TTL: time.Hour, Payload: []byte("t1100 idempotency reproduction")}
	}
	ev := idemScopeEvidence{UniqueConstraint: storeUniqueConstraints(t, s)}
	notRun := func(reason string) { ev.NotRunReasons = append(ev.NotRunReasons, reason) }
	countRows := func(where string, arg string) int {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM messages WHERE `+where, arg).Scan(&n); err != nil {
			t.Fatalf("count messages: %v", err)
		}
		return n
	}

	first, err := s.Send(ctx, request(sender, key))
	if err != nil {
		notRun("first send error: " + err.Error())
	} else {
		ev.FirstSendID = first.ID
		if countRows(`id=?`, first.ID) != 1 {
			notRun("first send left no row")
		}
	}

	restarted := sender
	restarted.SessionUUID = "t1100-sender-session-2"
	restarted.ProcessStart = restartedStart
	restarted.Generation = 1
	if restarted, err = s.RegisterPeer(ctx, restarted); err != nil {
		notRun("re-registration error: " + err.Error())
	}
	ev.SenderSessions = []string{sender.SessionUUID, restarted.SessionUUID}
	ev.SenderGenerations = []int64{sender.Generation, restarted.Generation}
	if restarted.SessionUUID == sender.SessionUUID || restarted.Generation <= sender.Generation {
		notRun("re-registration did not yield a different session UUID with a higher generation")
	}

	second, err := s.Send(ctx, request(restarted, key))
	switch {
	case err != nil:
		ev.SecondSendResult = "error"
		notRun("second send error: " + err.Error())
	case second.ID == ev.FirstSendID:
		ev.SecondSendResult = "same-id"
	default:
		ev.SecondSendResult = "new-id"
	}

	control, err := s.Send(ctx, request(restarted, "t1100-control"))
	if err != nil {
		notRun("control send error: " + err.Error())
	}

	ev.Rows = countRows(`idem_key=?`, key)
	keyIDs := map[string]bool{}
	idRows, err := s.db.QueryContext(ctx, `SELECT id FROM messages WHERE idem_key=?`, key)
	if err != nil {
		t.Fatalf("list keyed messages: %v", err)
	}
	for idRows.Next() {
		var id string
		if err := idRows.Scan(&id); err != nil {
			t.Fatalf("scan keyed message: %v", err)
		}
		keyIDs[id] = true
	}
	if err := idRows.Close(); err != nil {
		t.Fatalf("close keyed messages: %v", err)
	}

	claims, err := s.Claim(ctx, recipient, MaxBatch, time.Minute)
	if err != nil {
		t.Fatalf("recipient claim: %v", err)
	}
	claimedKey := map[string]bool{}
	for _, c := range claims {
		if keyIDs[c.ID] {
			claimedKey[c.ID] = true
		}
		if control.ID != "" && c.ID == control.ID {
			ev.ControlClaimed++
		}
	}
	ev.ClaimedIDs = len(claimedKey)
	if ev.ControlClaimed != 1 {
		notRun(fmt.Sprintf("control message claimed %d times", ev.ControlClaimed))
	}

	switch {
	case len(ev.NotRunReasons) > 0:
		ev.Outcome = "NOT_RUN"
	case ev.Rows >= 2 && ev.ClaimedIDs >= 2 && ev.SecondSendResult == "new-id":
		ev.Outcome = "reproduced"
	case ev.Rows == 1 && ev.ClaimedIDs == 1 && ev.SecondSendResult == "same-id":
		ev.Outcome = "not-reproduced"
	default:
		ev.Outcome = "unclassified"
	}

	body, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, '\n')
	if dir := os.Getenv(idemScopeEvidenceEnv); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create evidence dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "ac020-evidence.json"), body, 0o644); err != nil {
			t.Fatalf("write evidence: %v", err)
		}
		sum := sha256.Sum256(body)
		// Printed to stdout, not via t.Log, so the go test -json Output field
		// is exactly this line with no file:line prefix or indentation.
		fmt.Println("IDEM_SCOPE_REPRO_SHA256 " + hex.EncodeToString(sum[:]))
	} else {
		t.Logf("%s unset: measurement ran, evidence file not written", idemScopeEvidenceEnv)
	}

	switch ev.Outcome {
	case "NOT_RUN":
		fmt.Println("NOT_RUN: " + strings.Join(ev.NotRunReasons, "; "))
		t.Fatalf("measurement path not traversed: %v", ev.NotRunReasons)
	case "unclassified":
		t.Fatalf("measurement matched neither outcome: %s", body)
	}
	t.Logf("idempotency scope outcome: %s", ev.Outcome)
}
