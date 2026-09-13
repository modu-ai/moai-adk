package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrPlatformUnsupported = errors.New("gateway private authentication storage is not implemented on this platform")

var ErrAuthState = errors.New("invalid or inaccessible gateway authentication state")
var ErrBroker = errors.New("gateway authentication broker failed")
var ErrRefresh = errors.New("gateway authentication refresh not verified")

// Broker runs only in a new MoAI-owned home. It must return only after its
// subprocess has terminated; a successful RPC alone is not a successful Run.
type Broker interface {
	Run(context.Context, string, bool) error
}

type Store struct {
	dir      string
	root     *os.Root
	platform *platformStore
}
type state struct {
	Generation uint64          `json:"generation"`
	Tombstone  bool            `json:"tombstone"`
	Auth       json.RawMessage `json:"auth,omitempty"`
}
type tokenData struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	ID      string `json:"id_token"`
	Account string `json:"account_id"`
}
type authFile struct {
	APIKey string    `json:"OPENAI_API_KEY,omitempty"`
	Mode   string    `json:"auth_mode"`
	Tokens tokenData `json:"tokens"`
}
type Status struct {
	Generation uint64
	LoggedIn   bool
	ExpiresAt  time.Time
	Method     string
}

// OpenStore requires a private absolute directory and never reads CODEX_HOME.
func OpenStore(dir string) (*Store, error) { return openPlatformStore(dir) }
func (s *Store) Close() error              { return errors.Join(s.root.Close(), s.closePlatform()) }
func (s *Store) lock(ctx context.Context, name string) (func(), error) {
	if e := s.validatePlatformRoot(); e != nil {
		return nil, e
	}
	info, e := s.root.Lstat(name)
	if e == nil && !privateRegular(filepath.Join(s.dir, name), info) {
		return nil, ErrAuthState
	}
	if e != nil && !errors.Is(e, os.ErrNotExist) {
		return nil, ErrAuthState
	}
	f, e := s.openLock(name)
	if e != nil {
		return nil, ErrAuthState
	}
	if e = lockFile(ctx, f); e != nil {
		_ = f.Close() // lock not held; discarding the descriptor
		return nil, e
	}
	return func() { unlockFile(f); _ = f.Close() }, nil // release path; unlock governs, flock ends at exit
}
func (s *Store) read() (state, error) {
	if e := s.validatePlatformRoot(); e != nil {
		return state{}, e
	}
	info, e := s.stateInfo()
	if errors.Is(e, os.ErrNotExist) {
		return state{Tombstone: true}, nil
	}
	if e != nil || !privateRegular(filepath.Join(s.dir, "state.json"), info) {
		return state{}, ErrAuthState
	}
	f, e := s.root.Open("state.json")
	if e != nil {
		return state{}, ErrAuthState
	}
	defer func() { _ = f.Close() }() // read-only state source; no write-back to lose
	if e := validateOpenedPrivateFile(f); e != nil {
		return state{}, e
	}
	current, e := f.Stat()
	if e != nil || !os.SameFile(info, current) {
		return state{}, ErrAuthState
	}
	raw, e := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if e != nil || len(raw) > 1<<20 {
		return state{}, ErrAuthState
	}
	var v state
	if json.Unmarshal(raw, &v) != nil || v.Generation == 0 || v.Tombstone && len(v.Auth) != 0 {
		return state{}, ErrAuthState
	}
	if !v.Tombstone {
		if _, _, e = parseAuth(v.Auth); e != nil {
			return state{}, e
		}
	}
	return v, nil
}
func (s *Store) write(v state) error {
	raw, e := json.Marshal(v)
	if e != nil {
		return ErrAuthState
	}
	return s.writePlatform(raw)
}
func parseAuth(raw []byte) (authFile, time.Time, error) {
	var a authFile
	if json.Unmarshal(raw, &a) != nil || a.Mode != "chatgpt" || a.APIKey != "" || a.Tokens.Access == "" || a.Tokens.Refresh == "" || a.Tokens.ID == "" || a.Tokens.Account == "" {
		return a, time.Time{}, ErrAuthState
	}
	for _, v := range []string{a.Tokens.Access, a.Tokens.Refresh, a.Tokens.Account} {
		if strings.ContainsAny(v, "\r\n\x00") {
			return a, time.Time{}, ErrAuthState
		}
	}
	parts := strings.Split(a.Tokens.Access, ".")
	if len(parts) != 3 {
		return a, time.Time{}, ErrAuthState
	}
	claims, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return a, time.Time{}, ErrAuthState
	}
	var expiry struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(claims, &expiry) != nil || expiry.Exp <= 0 {
		return a, time.Time{}, ErrAuthState
	}
	return a, time.Unix(expiry.Exp, 0), nil
}
func (s *Store) Status(ctx context.Context) (Status, error) {
	unlock, e := s.lock(ctx, "state.lock")
	if e != nil {
		return Status{}, e
	}
	defer unlock()
	v, e := s.read()
	if e != nil {
		return Status{}, e
	}
	out := Status{Generation: v.Generation, LoggedIn: !v.Tombstone, Method: "subscription"}
	if !v.Tombstone {
		_, out.ExpiresAt, e = parseAuth(v.Auth)
	}
	return out, e
}

// Owns checks reference provenance only; SendAuthorized separately checks its
// generation and the current logout state before transmitting credentials.
func (s *Store) Owns(ref CredentialRef) bool {
	r, ok := ref.(*storeRef)
	return s != nil && ok && r != nil && r.s == s
}

func (s *Store) Resolve() (CredentialRef, error) {
	status, e := s.Status(context.Background())
	if e != nil {
		return nil, e
	}
	if !status.LoggedIn {
		return nil, ErrCredentialAbsent
	}
	return &storeRef{s: s, expected: status.Generation}, nil
}
func (s *Store) Login(ctx context.Context, b Broker) (uint64, error) {
	return s.transact(ctx, 0, b, nil, false)
}
func (s *Store) Refresh(ctx context.Context, expected uint64, b Broker, verify func(context.Context, CredentialRef) error) (uint64, error) {
	if expected == 0 || verify == nil {
		return 0, ErrRefresh
	}
	return s.transact(ctx, expected, b, verify, true)
}

// @MX:ANCHOR: [AUTO] Serialize broker operations independently of logout state.
// @MX:REASON: Login and refresh must not block tombstone publication while waiting on a user or provider.
func (s *Store) transact(ctx context.Context, expected uint64, b Broker, verify func(context.Context, CredentialRef) error, refresh bool) (uint64, error) {
	if b == nil {
		return 0, ErrBroker
	}
	operation, e := s.lock(ctx, "operation.lock")
	if e != nil {
		return 0, e
	}
	defer operation()
	unlock, e := s.lock(ctx, "state.lock")
	if e != nil {
		return 0, e
	}
	before, e := s.read()
	unlock()
	if e != nil {
		return 0, e
	}
	if refresh {
		if before.Tombstone {
			return 0, ErrCredentialAbsent
		}
		if before.Generation != expected {
			return before.Generation, nil
		}
	}
	home, e := privateScratch(s.dir, "broker-")
	if e != nil {
		return 0, ErrAuthState
	}
	defer func() { _ = os.RemoveAll(home) }() // scratch cleanup; a leftover dotfile is harmless
	homeInfo, e := privatePathInfo(home, true)
	if e != nil {
		return 0, ErrAuthState
	}
	releaseHome, e := holdPrivateDirectory(home)
	if e != nil {
		return 0, ErrAuthState
	}
	defer releaseHome()
	if refresh {
		if e = seedPrivateFile(filepath.Join(home, "auth.json"), before.Auth); e != nil {
			return 0, ErrAuthState
		}
	}
	if b.Run(ctx, home, refresh) != nil {
		return 0, ErrBroker
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	currentHome, e := privatePathInfo(home, true)
	if e != nil || !currentHome.IsDir() || !os.SameFile(homeInfo, currentHome) {
		return 0, ErrAuthState
	}
	info, e := privatePathInfo(filepath.Join(home, "auth.json"), false)
	if e != nil || !privateRegular(filepath.Join(home, "auth.json"), info) || info.Size() > 1<<20 {
		return 0, ErrAuthState
	}
	raw, e := readPrivateBrokerFile(home, homeInfo, info)
	if e != nil {
		return 0, ErrAuthState
	}
	a, expiry, e := parseAuth(raw)
	if e != nil || !expiry.After(time.Now()) {
		return 0, ErrAuthState
	}
	if refresh {
		old, oldExpiry, e := parseAuth(before.Auth)
		if e != nil || old.Tokens.Account != a.Tokens.Account || old.Tokens.Access == a.Tokens.Access || !expiry.After(oldExpiry) {
			return 0, ErrRefresh
		}
		if verify(ctx, &snapshotRef{data: a, expiry: expiry, generation: before.Generation + 1}) != nil {
			return 0, ErrRefresh
		}
	}
	releaseHome()
	// Remove private broker artifacts before making the candidate canonical.
	// A cleanup failure must leave the previous generation unchanged.
	if e = os.RemoveAll(home); e != nil {
		return 0, ErrAuthState
	}
	unlock, e = s.lock(ctx, "state.lock")
	if e != nil {
		return 0, e
	}
	defer unlock()
	current, e := s.read()
	if e != nil {
		return 0, e
	}
	if current.Generation != before.Generation {
		return 0, ErrCredentialChanged
	}
	if current.Generation == ^uint64(0) {
		return 0, ErrAuthState
	}
	next := state{Generation: current.Generation + 1, Auth: raw}
	if e = s.write(next); e != nil {
		return 0, e
	}
	return next.Generation, nil
}

// Logout records local revocation only; it makes no remote revocation claim.
func (s *Store) Logout(ctx context.Context) (uint64, error) {
	generation, _, err := s.logoutSnapshot(ctx)
	return generation, err
}

func (s *Store) logoutSnapshot(ctx context.Context) (uint64, []byte, error) {
	unlock, e := s.lock(ctx, "state.lock")
	if e != nil {
		return 0, nil, e
	}
	defer unlock()
	v, e := s.read()
	if e != nil {
		return 0, nil, e
	}
	if v.Generation == ^uint64(0) {
		return 0, nil, ErrAuthState
	}
	next := state{Generation: v.Generation + 1, Tombstone: true}
	if e = s.write(next); e != nil {
		return 0, nil, e
	}
	return next.Generation, v.Auth, nil
}

type storeRef struct {
	s        *Store
	expected uint64
}

func (r *storeRef) Provider() ProviderID { return ProviderOpenAI }
func (r *storeRef) Redacted() string     { return "openai:subscription" }
func (r *storeRef) Generation() (uint64, error) {
	v, e := r.s.Status(context.Background())
	if e != nil {
		return 0, e
	}
	if !v.LoggedIn {
		return 0, ErrCredentialAbsent
	}
	return v.Generation, nil
}
func (r *storeRef) Apply(req *http.Request) error {
	unlock, e := r.s.lock(req.Context(), "state.lock")
	if e != nil {
		return e
	}
	defer unlock()
	v, e := r.s.read()
	if e != nil {
		return e
	}
	if v.Tombstone {
		return ErrCredentialAbsent
	}
	if v.Generation != r.expected {
		return ErrCredentialChanged
	}
	a, expiry, e := parseAuth(v.Auth)
	if e != nil {
		return e
	}
	return applySubscription(req, a, expiry)
}

type snapshotRef struct {
	data       authFile
	expiry     time.Time
	generation uint64
}

func (r *snapshotRef) Provider() ProviderID          { return ProviderOpenAI }
func (r *snapshotRef) Redacted() string              { return "openai:subscription" }
func (r *snapshotRef) Generation() (uint64, error)   { return r.generation, nil }
func (r *snapshotRef) Apply(req *http.Request) error { return applySubscription(req, r.data, r.expiry) }

const SubscriptionEndpoint = "https://chatgpt.com/backend-api/codex/responses"
const APIEndpoint = "https://api.openai.com/v1/responses"

func applySubscription(req *http.Request, a authFile, expiry time.Time) error {
	if req == nil || req.URL == nil || req.URL.String() != SubscriptionEndpoint || req.URL.User != nil || req.Host != "" && req.Host != "chatgpt.com" {
		return ErrWrongProvider
	}
	if !expiry.After(time.Now()) {
		return ErrCredentialAbsent
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Authorization", "Bearer "+a.Tokens.Access)
	req.Header.Set("ChatGPT-Account-Id", a.Tokens.Account)
	return nil
}

func readPrivateBrokerFile(home string, homeInfo, info os.FileInfo) ([]byte, error) {
	if privateDirectory(home) != nil {
		return nil, ErrAuthState
	}
	f, e := os.Open(filepath.Join(home, "auth.json"))
	if e != nil {
		return nil, ErrAuthState
	}
	defer func() { _ = f.Close() }() // read-only broker source; no write-back to lose
	if validateOpenedPrivateFile(f) != nil {
		return nil, ErrAuthState
	}
	fi, e := f.Stat()
	if e != nil || !os.SameFile(info, fi) {
		return nil, ErrAuthState
	}
	raw, e := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if e != nil || len(raw) > 1<<20 {
		return nil, ErrAuthState
	}
	hi, e := privatePathInfo(home, true)
	if e != nil || !os.SameFile(homeInfo, hi) || privateDirectory(home) != nil {
		return nil, ErrAuthState
	}
	return raw, nil
}
