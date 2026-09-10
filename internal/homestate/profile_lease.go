package homestate

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/paths"
	_ "modernc.org/sqlite"
)

type ProfileLease struct {
	Token              string
	ProfileName        string
	ProfilePath        string
	ProjectKey         string
	PID                int
	ProcessFingerprint string
	SessionID          string
}

type ProfileLeaseStore struct {
	DB   *sql.DB
	Path string
}

type ProcessIdentityState string

const (
	ProcessIdentityLive          ProcessIdentityState = "live"
	ProcessIdentityDead          ProcessIdentityState = "dead"
	ProcessIdentityIndeterminate ProcessIdentityState = "indeterminate"
)

type LeaseProtection string

const (
	LeaseLive          LeaseProtection = "live"
	LeaseStale         LeaseProtection = "stale"
	LeaseIndeterminate LeaseProtection = "indeterminate"
	LeaseNone          LeaseProtection = "none"
)

func OpenProfileLeases() (*ProfileLeaseStore, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, "run")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "profile-leases.db")
	v := url.Values{}
	v.Add("_pragma", "journal_mode(WAL)")
	v.Add("_pragma", "busy_timeout(5000)")
	v.Add("_txlock", "immediate")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS profile_leases(token TEXT PRIMARY KEY,profile_name TEXT NOT NULL,profile_path TEXT NOT NULL,project_key TEXT NOT NULL DEFAULT '',pid INTEGER NOT NULL,process_fingerprint TEXT NOT NULL DEFAULT '',session_id TEXT NOT NULL DEFAULT '',state TEXT NOT NULL CHECK(state IN ('provisional','transferring','enriched','released')),transfer_deadline TEXT NOT NULL DEFAULT '',created_at TEXT NOT NULL,updated_at TEXT NOT NULL); CREATE INDEX IF NOT EXISTS profile_lease_path_state ON profile_leases(profile_path,state)`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`ALTER TABLE profile_leases ADD COLUMN transfer_deadline TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(err.Error(), "duplicate column") {
		_ = db.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := secureFactoryArtifacts(path); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &ProfileLeaseStore{DB: db, Path: path}, nil
}
func (s *ProfileLeaseStore) Close() error { return s.DB.Close() }
func leaseToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
func (s *ProfileLeaseStore) CreateProvisional(ctx context.Context, l ProfileLease) (string, error) {
	if l.Token == "" {
		var err error
		l.Token, err = leaseToken()
		if err != nil {
			return "", err
		}
	}
	path, err := filepath.Abs(l.ProfilePath)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	deadline := time.Now().UTC().Add(30 * time.Second).Format(time.RFC3339Nano)
	_, err = s.DB.ExecContext(ctx, `INSERT INTO profile_leases(token,profile_name,profile_path,project_key,pid,process_fingerprint,session_id,state,transfer_deadline,created_at,updated_at) VALUES(?,?,?,?,?,?,?,'provisional',?,?,?)`, l.Token, l.ProfileName, filepath.Clean(path), l.ProjectKey, l.PID, l.ProcessFingerprint, l.SessionID, deadline, now, now)
	return l.Token, err
}
func (s *ProfileLeaseStore) TransferToChild(ctx context.Context, token string, parentPID int, parentFingerprint string, childPID int, childFingerprint string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	deadline := time.Now().UTC().Add(30 * time.Second).Format(time.RFC3339Nano)
	res, err := s.DB.ExecContext(ctx, `UPDATE profile_leases SET state='transferring',pid=?,process_fingerprint=?,transfer_deadline=?,updated_at=? WHERE token=? AND state='provisional' AND pid=? AND process_fingerprint=?`, childPID, childFingerprint, deadline, now, token, parentPID, parentFingerprint)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return fmt.Errorf("profile lease transfer CAS conflict")
	}
	return nil
}
func (s *ProfileLeaseStore) Enrich(ctx context.Context, token, sessionID string, pid int, fingerprint string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.DB.ExecContext(ctx, `UPDATE profile_leases SET state='enriched',session_id=?,pid=?,process_fingerprint=?,updated_at=? WHERE token=? AND state IN ('provisional','transferring','enriched')`, sessionID, pid, fingerprint, now, token)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return fmt.Errorf("profile lease enrich CAS conflict")
	}
	return nil
}
func (s *ProfileLeaseStore) ReleaseSession(ctx context.Context, sessionID string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE profile_leases SET state='released',updated_at=? WHERE session_id=? AND state<>'released'`, time.Now().UTC().Format(time.RFC3339Nano), sessionID)
	return err
}
func (s *ProfileLeaseStore) ProfileProtection(ctx context.Context, profilePath string, probe func(int) (string, ProcessIdentityState)) (LeaseProtection, error) {
	abs, err := filepath.Abs(profilePath)
	if err != nil {
		return LeaseIndeterminate, err
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT pid,process_fingerprint,state,transfer_deadline FROM profile_leases WHERE profile_path=? AND state<>'released'`, filepath.Clean(abs))
	if err != nil {
		return LeaseIndeterminate, err
	}
	defer func() { _ = rows.Close() }()
	found := false
	staleOnly := true
	for rows.Next() {
		found = true
		var pid int
		var fingerprint, state, deadline string
		if err := rows.Scan(&pid, &fingerprint, &state, &deadline); err != nil {
			return LeaseIndeterminate, err
		}
		if fingerprint == "" {
			return LeaseIndeterminate, nil
		}
		got, status := probe(pid)
		if status == ProcessIdentityLive && got == fingerprint {
			return LeaseLive, nil
		}
		if (state == "provisional" || state == "transferring") && (deadline == "" || deadline > time.Now().UTC().Format(time.RFC3339Nano)) {
			return LeaseIndeterminate, nil
		}
		if status == ProcessIdentityIndeterminate {
			return LeaseIndeterminate, nil
		}
		if status == ProcessIdentityLive && got != fingerprint {
			continue
		}
		if status != ProcessIdentityDead {
			staleOnly = false
		}
	}
	if err := rows.Err(); err != nil {
		return LeaseIndeterminate, err
	}
	if !found {
		return LeaseNone, nil
	}
	if staleOnly {
		return LeaseStale, nil
	}
	return LeaseIndeterminate, nil
}

func CurrentProcessFingerprint() string {
	fp, state := ProbeProcessIdentity(os.Getpid())
	if state != ProcessIdentityLive {
		return ""
	}
	return fp
}
func ProbeProcessIdentity(pid int) (string, ProcessIdentityState) {
	state := platformPIDState(pid)
	if state != ProcessIdentityLive {
		return "", state
	}
	fingerprint, ok := platformProcessFingerprint(pid)
	if !ok {
		return "", ProcessIdentityIndeterminate
	}
	return fingerprint, ProcessIdentityLive
}

func ProfileProtectionAtHome(profilePath string) (LeaseProtection, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return LeaseIndeterminate, err
	}
	if _, err := os.Stat(filepath.Join(home, "run", "profile-leases.db")); os.IsNotExist(err) {
		return LeaseNone, nil
	} else if err != nil {
		return LeaseIndeterminate, err
	}
	s, err := OpenProfileLeases()
	if err != nil {
		return LeaseIndeterminate, err
	}
	defer func() { _ = s.Close() }()
	return s.ProfileProtection(context.Background(), profilePath, ProbeProcessIdentity)
}
