package sessionmsg

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
)

// AgentCapabilities is the A2A AgentCard capabilities subset the broker
// records (design.md §3.1 — reference shape, security_schemes excluded).
type AgentCapabilities struct {
	Messaging bool `json:"messaging"`
}

// AgentRecord is one registered agent: an A2A AgentCard reference shape plus
// broker-owned identity fields (REQ-CSM-003). Persisted at
// agents/<agentId>.json.
type AgentRecord struct {
	AgentID       string            `json:"agentId"`
	Kind          string            `json:"kind"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Version       string            `json:"version"`
	Capabilities  AgentCapabilities `json:"capabilities"`
	CWD           string            `json:"cwd,omitempty"`
	PID           int               `json:"pid,omitempty"`
	Host          string            `json:"host,omitempty"`
	RegisteredAt  time.Time         `json:"registeredAt"`
	LastHeartbeat time.Time         `json:"lastHeartbeat"`
}

// AgentInfo is the ListAgents reporting shape (design.md §6: id, name, kind,
// online, pending count).
type AgentInfo struct {
	AgentID       string    `json:"agentId"`
	Kind          string    `json:"kind"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Online        bool      `json:"online"`
	PendingCount  int       `json:"pendingCount"`
	LastHeartbeat time.Time `json:"lastHeartbeat"`
}

// validKind reports whether kind is a registerable agent kind.
func validKind(kind string) bool {
	return kind == KindClaude || kind == KindCodex
}

// newAgentID mints an agentId of the form "<kind>-<hex8>" (design.md §3.2).
func newAgentID(kind string) (string, error) {
	b := make([]byte, 4) // 4 bytes → 8 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("sessionmsg: agent id entropy: %w", err)
	}
	return fmt.Sprintf("%s-%x", kind, b), nil
}

// Register idempotently registers kind+name and returns the stable agentId
// (REQ-CSM-003): an existing kind+name record gets its heartbeat refreshed
// (and its description when a non-empty one is supplied); a new record is
// written atomically to agents/<agentId>.json in the A2A AgentCard
// reference shape. The whole scan-then-create critical section runs under
// the registration advisory lock, so concurrent registrations of the same
// kind+name converge on one agentId.
//
// @MX:ANCHOR: [AUTO] broker registration primitive — idempotent kind+name to stable agentId
// @MX:REASON: fan_in >= 3 (M2 session_msg_register handler, e2e registration of both session kinds, re-register recovery after agentId loss). Changing id minting or the agents/ layout re-addresses every mailbox.
func (s *Store) Register(kind, name, description string) (AgentRecord, error) {
	if !validKind(kind) {
		return AgentRecord{}, fmt.Errorf("sessionmsg: register: unknown kind %q (want %q or %q)", kind, KindClaude, KindCodex)
	}
	if name == "" {
		return AgentRecord{}, errors.New("sessionmsg: register: name is empty")
	}

	var out AgentRecord
	err := s.withAgentLock(lockNameRegister, func() error {
		records, err := s.readAllAgents()
		if err != nil {
			return err
		}
		now := s.clock.Now().UTC()

		existing := map[string]bool{}
		for _, rec := range records {
			existing[rec.AgentID] = true
			if rec.Kind == kind && rec.Name == name {
				rec.LastHeartbeat = now
				if description != "" {
					rec.Description = description
				}
				if err := writeJSONAtomic(s.agentPath(rec.AgentID), rec); err != nil {
					return err
				}
				out = rec
				return nil
			}
		}

		var id string
		for attempt := 0; attempt < 8; attempt++ {
			candidate, err := newAgentID(kind)
			if err != nil {
				return err
			}
			if !existing[candidate] {
				id = candidate
				break
			}
		}
		if id == "" {
			return errors.New("sessionmsg: register: agent id collision exhausted retries")
		}

		host, _ := os.Hostname()
		cwd, _ := os.Getwd()
		rec := AgentRecord{
			AgentID:       id,
			Kind:          kind,
			Name:          name,
			Description:   description,
			Version:       "1",
			Capabilities:  AgentCapabilities{Messaging: true},
			CWD:           cwd,
			PID:           os.Getpid(),
			Host:          host,
			RegisteredAt:  now,
			LastHeartbeat: now,
		}
		if err := writeJSONAtomic(s.agentPath(id), rec); err != nil {
			return err
		}
		out = rec
		return nil
	})
	if err != nil {
		return AgentRecord{}, err
	}
	return out, nil
}

// ListAgents reports every registered agent with its online state
// (heartbeat age within config.DefaultSessionMsgAgentOfflineMinutes,
// REQ-CSM-004) and pending mailbox count. Reads are lock-free and eventually
// consistent (registry precedent).
func (s *Store) ListAgents() ([]AgentInfo, error) {
	records, err := s.readAllAgents()
	if err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	offline := time.Duration(config.DefaultSessionMsgAgentOfflineMinutes) * time.Minute
	infos := make([]AgentInfo, 0, len(records))
	for _, rec := range records {
		infos = append(infos, AgentInfo{
			AgentID:       rec.AgentID,
			Kind:          rec.Kind,
			Name:          rec.Name,
			Description:   rec.Description,
			Online:        now.Sub(rec.LastHeartbeat) <= offline,
			PendingCount:  countFiles(s.pendingDir(rec.AgentID)),
			LastHeartbeat: rec.LastHeartbeat,
		})
	}
	return infos, nil
}

// heartbeat refreshes one agent's LastHeartbeat under the agent's own
// advisory lock. Idempotent on a missing record (registry precedent) — the
// mailbox/poll paths call it after releasing the mailbox lock, so this never
// nests inside another withAgentLock scope.
func (s *Store) heartbeat(agentID string) error {
	return s.withAgentLock(agentID, func() error {
		rec, err := s.readAgent(agentID)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		rec.LastHeartbeat = s.clock.Now().UTC()
		return writeJSONAtomic(s.agentPath(agentID), rec)
	})
}

// readAllAgents loads every agents/*.json record. Directory order (sorted by
// filename) gives deterministic listing; a missing directory is an empty
// registry.
func (s *Store) readAllAgents() ([]AgentRecord, error) {
	entries, err := os.ReadDir(s.agentsDir())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("sessionmsg: read agents dir: %w", err)
	}
	var records []AgentRecord
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := atomicfile.ReadFile(filepath.Join(s.agentsDir(), e.Name()))
		if err != nil {
			return nil, fmt.Errorf("sessionmsg: read agent %s: %w", e.Name(), err)
		}
		var rec AgentRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, fmt.Errorf("sessionmsg: parse agent %s: %w", e.Name(), err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// readAgent loads one agent record. A missing file returns the raw error with
// its fs.ErrNotExist chain intact, for the caller to classify with errors.Is.
func (s *Store) readAgent(agentID string) (AgentRecord, error) {
	data, err := atomicfile.ReadFile(s.agentPath(agentID))
	if err != nil {
		return AgentRecord{}, err
	}
	var rec AgentRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return AgentRecord{}, fmt.Errorf("sessionmsg: parse agent %s: %w", agentID, err)
	}
	return rec, nil
}

// countFiles counts regular files in dir (0 when the directory is absent).
func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			n++
		}
	}
	return n
}
