package hook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// LedgerHeartbeatEvery is the least time between two ledger writes of one
// session (ledger contract §6, heartbeat throttle).
const LedgerHeartbeatEvery = 30 * time.Second

// ledgerHeartbeatTimeout bounds one beat: the seam runs on a latency-sensitive
// hook path and its result is advisory.
const ledgerHeartbeatTimeout = 2 * time.Second

// LedgerBinding is what a claimed card tells its session's hooks: which ledger
// file, which card, which fencing epoch. The claiming side writes it.
type LedgerBinding struct {
	DB     string `json:"db"`
	CardID string `json:"card_id"`
	Epoch  int64  `json:"epoch"`
}

var (
	ledgerHeartbeatNow = time.Now
	// ledgerBeat is the one DB write; tests replace it.
	ledgerBeat = func(db, cardID string, epoch int64) error {
		ctx, cancel := context.WithTimeout(context.Background(), ledgerHeartbeatTimeout)
		defer cancel()
		st, err := ledger.Open(ctx, db)
		if err != nil {
			return err
		}
		defer func() { _ = st.Close() }()
		return st.Heartbeat(ctx, cardID, epoch, nil)
	}
)

func ledgerHeartbeatDir(root string) string {
	return filepath.Join(root, ".moai", "state", "ledger-heartbeat")
}

// ledgerSessionFile names a per-session file, refusing an id that could leave
// the state directory.
func ledgerSessionFile(root, sessionID, ext string) (string, error) {
	if sessionID == "" || sessionID == "." || sessionID == ".." || strings.ContainsAny(sessionID, `/\`) {
		return "", fmt.Errorf("ledger heartbeat: unusable session id %q", sessionID)
	}
	return filepath.Join(ledgerHeartbeatDir(root), sessionID+ext), nil
}

// BindLedgerHeartbeat records the card this session holds so its hooks can
// renew the lease. Called after a claim.
func BindLedgerHeartbeat(root, sessionID string, b LedgerBinding) error {
	path, err := ledgerSessionFile(root, sessionID, ".json")
	if err != nil {
		return err
	}
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LedgerHeartbeatSeam renews the ledger lease of the card this session holds, at
// most once per LedgerHeartbeatEvery. A session with no binding does nothing and
// creates nothing. The throttle lives in a per-session stamp file, so a call
// inside the window costs two small file reads and no database open. Fail-open:
// every error returns silently; the stamp is written before the beat, so a
// failing ledger is not retried inside the window.
func LedgerHeartbeatSeam(input *HookInput) {
	if input == nil || input.SessionID == "" {
		return
	}
	root := resolveProjectRoot(input)
	if root == "" {
		return
	}
	bindPath, err := ledgerSessionFile(root, input.SessionID, ".json")
	if err != nil {
		return
	}
	raw, err := os.ReadFile(bindPath)
	if err != nil {
		return
	}
	var b LedgerBinding
	if json.Unmarshal(raw, &b) != nil || b.DB == "" || b.CardID == "" || b.Epoch < 1 {
		return
	}
	stampPath, _ := ledgerSessionFile(root, input.SessionID, ".stamp")
	now := ledgerHeartbeatNow()
	if last, ok := readLedgerStamp(stampPath); ok && now.Sub(last) < LedgerHeartbeatEvery {
		return
	}
	if err := os.WriteFile(stampPath, []byte(strconv.FormatInt(now.UnixNano(), 10)), 0o600); err != nil {
		return
	}
	switch err := ledgerBeat(b.DB, b.CardID, b.Epoch); {
	case err == nil:
	case errors.Is(err, ledger.ErrStaleEpoch):
		// The card was claimed again: this epoch is dead, stop renewing it.
		_ = os.Remove(bindPath)
	default:
		slog.Debug("ledger heartbeat seam: beat failed (non-blocking)", "card", b.CardID, "error", err.Error())
	}
}

func readLedgerStamp(path string) (time.Time, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(0, n), true
}
