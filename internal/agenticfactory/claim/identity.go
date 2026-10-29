package claim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

const maxAgents = 1000

func (s *Store) agentsDir() string { return filepath.Join(s.dir, "agents") }

func (s *Store) agentPath(n int) string {
	return filepath.Join(s.agentsDir(), fmt.Sprintf("agent-%d.json", n))
}

func readAgentFile(path string) (core.Agent, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return core.Agent{}, err
	}
	var a core.Agent
	if err := json.Unmarshal(b, &a); err != nil {
		return core.Agent{}, fmt.Errorf("%w: corrupt agent record: %v", core.ErrStoreUnavailable, err)
	}
	return a, nil
}

// agents returns every registered agent.
func (s *Store) agents() ([]core.Agent, error) {
	ents, err := os.ReadDir(s.agentsDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var out []core.Agent
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "agent-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		a, err := readAgentFile(filepath.Join(s.agentsDir(), e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// RegisterAgent gives a session its agent number. The number is taken by
// hard-linking a fully written file to agent-<n>.json, which fails if that
// name exists, so concurrent registrations (even from separate processes)
// can never receive the same number and a reader never sees a half-written
// record. Registering a session again returns its existing identity.
func (s *Store) RegisterAgent(sessionID string) (core.Agent, error) {
	if _, err := core.ParseSessionID(sessionID); err != nil {
		return core.Agent{}, err
	}
	if err := os.MkdirAll(s.agentsDir(), 0o700); err != nil {
		return core.Agent{}, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if existing, err := s.agentOf(sessionID); err != nil || existing.Number != 0 {
		return existing, err
	}
	for n := 1; n <= maxAgents; n++ {
		a := core.Agent{SessionID: sessionID, Number: n}
		linked, err := s.linkAgent(a)
		if err != nil {
			return core.Agent{}, err
		}
		if !linked {
			continue
		}
		// Two registrations of the same session may both have won a number;
		// the lowest stays, the other gives its number back.
		low, err := s.agentOf(sessionID)
		if err != nil {
			return core.Agent{}, err
		}
		if low.Number != 0 && low.Number < n {
			_ = os.Remove(s.agentPath(n))
			return low, nil
		}
		return a, nil
	}
	return core.Agent{}, fmt.Errorf("%w: no free agent number up to %d", core.ErrStoreUnavailable, maxAgents)
}

func (s *Store) agentOf(sessionID string) (core.Agent, error) {
	all, err := s.agents()
	if err != nil {
		return core.Agent{}, err
	}
	var best core.Agent
	for _, a := range all {
		if a.SessionID == sessionID && (best.Number == 0 || a.Number < best.Number) {
			best = a
		}
	}
	return best, nil
}

// linkAgent publishes agent-<n>.json atomically; false means the number is taken.
func (s *Store) linkAgent(a core.Agent) (bool, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	tmp, err := os.CreateTemp(s.agentsDir(), "agent.tmp-*")
	if err != nil {
		return false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := os.Link(tmp.Name(), s.agentPath(a.Number)); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return true, nil
}

// ResumeAgent confirms that session id and number are the registered pair and
// that every live claim the session holds carries the same display name. It
// returns the identity and the claims the agent still owns, by Issue number.
// Claims acquired without an agent identity are not counted.
func (s *Store) ResumeAgent(sessionID string, number int) (core.Agent, []Record, error) {
	a, err := readAgentFile(s.agentPath(number))
	if errors.Is(err, os.ErrNotExist) {
		return core.Agent{}, nil, core.ErrNotHeld
	}
	if err != nil {
		return core.Agent{}, nil, err
	}
	if a.SessionID != sessionID || a.Number != number {
		return core.Agent{}, nil, fmt.Errorf("%w: agent-%d is not session %q", core.ErrIdentityMismatch, number, sessionID)
	}
	recs, err := s.List()
	if err != nil {
		return core.Agent{}, nil, err
	}
	var owned []Record
	for _, r := range recs {
		if r.Holder != sessionID || r.AgentName == "" {
			continue
		}
		if r.AgentName != a.DisplayName() {
			return core.Agent{}, nil, fmt.Errorf("%w: issue-%d is bound to %q, not %q", core.ErrIdentityMismatch, r.Issue, r.AgentName, a.DisplayName())
		}
		owned = append(owned, r)
	}
	return a, owned, nil
}

// verifyAgent confirms a is still the registered holder of its number. A
// released identity, or one whose number now belongs to another session, is
// refused so it can never write under a name it no longer owns.
func (s *Store) verifyAgent(a core.Agent) error {
	cur, err := readAgentFile(s.agentPath(a.Number))
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: agent-%d is not registered", core.ErrIdentityMismatch, a.Number)
	}
	if err != nil {
		return err
	}
	if cur != a {
		return fmt.Errorf("%w: agent-%d belongs to another session", core.ErrIdentityMismatch, a.Number)
	}
	return nil
}

// ReleaseAgent returns an agent number so another session can take it. Only
// the registered session may release, and only once it holds no live claim;
// a PR it already handed off keeps its own session id in the receipt, so
// attribution survives the number being reused.
func (s *Store) ReleaseAgent(sessionID string, number int) error {
	a, err := readAgentFile(s.agentPath(number))
	if errors.Is(err, os.ErrNotExist) {
		return core.ErrNotHeld
	}
	if err != nil {
		return err
	}
	if a.SessionID != sessionID {
		return fmt.Errorf("%w: agent-%d is not session %q", core.ErrIdentityMismatch, number, sessionID)
	}
	recs, err := s.List()
	if err != nil {
		return err
	}
	for _, r := range recs {
		if r.Holder == sessionID {
			return fmt.Errorf("%w: agent-%d still holds issue-%d", core.ErrHeld, number, r.Issue)
		}
	}
	if err := os.Remove(s.agentPath(number)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return nil
}
