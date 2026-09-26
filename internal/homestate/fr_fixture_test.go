package homestate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// frRepo is a fixture git repository shaped for the F1 evidence readers.
//
//	main:        C (adds the artifact)
//	feature:     C ── F
//	integration: C ── M (merge --no-ff feature; tree(M) == tree(F))
//
// The working tree is checked out on integration, so HEAD = M and C is an
// ancestor of HEAD. With a remote, origin/integration = M.
type frRepo struct {
	Dir         string
	Commit      string
	Artifact    string
	Merge       string
	MergeTree   string
	Remeasure   string
	Integration string
}

const frIntegration = "integration"

// frGitEnv isolates git from the developer's global and system config.
func frGitEnv(t *testing.T) {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, []byte("[user]\n\tname = F1 Test\n\temail = f1@example.invalid\n[init]\n\tdefaultBranch = main\n"), 0o600); err != nil {
		t.Fatalf("write git config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func frGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func frWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func frNewRepo(t *testing.T, withRemote bool) frRepo {
	t.Helper()
	frGitEnv(t)
	base := t.TempDir()
	dir := filepath.Join(base, "repo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	frGit(t, dir, "init", "-q", "-b", "main")
	artifact := ".moai/specs/SPEC-FIXTURE-001/spec.md"
	frWrite(t, filepath.Join(dir, artifact), "fixture spec\n")
	frWrite(t, filepath.Join(dir, "a.txt"), "a\n")
	frGit(t, dir, "add", "-A")
	frGit(t, dir, "commit", "-q", "-m", "C")
	c := frGit(t, dir, "rev-parse", "HEAD")
	frGit(t, dir, "branch", frIntegration)
	frGit(t, dir, "checkout", "-q", "-b", "feature")
	frWrite(t, filepath.Join(dir, "b.txt"), "b\n")
	frGit(t, dir, "add", "-A")
	frGit(t, dir, "commit", "-q", "-m", "F")
	frGit(t, dir, "checkout", "-q", frIntegration)
	frGit(t, dir, "merge", "-q", "--no-ff", "-m", "M", "feature")
	m := frGit(t, dir, "rev-parse", "HEAD")
	tree := frGit(t, dir, "rev-parse", m+"^{tree}")
	if withRemote {
		bare := filepath.Join(base, "origin.git")
		frGit(t, base, "init", "-q", "--bare", bare)
		frGit(t, dir, "remote", "add", "origin", bare)
		frGit(t, dir, "push", "-q", "origin", frIntegration)
	}
	remeasure := filepath.Join(base, "remeasure.md")
	frWrite(t, remeasure, "re-measured on merge "+m+"\n")
	return frRepo{Dir: dir, Commit: c, Artifact: artifact, Merge: m, MergeTree: tree, Remeasure: remeasure, Integration: frIntegration}
}

// frWriteVerdict writes a card verdict file in the machine-readable form the
// E-VERDICT reader consumes.
func frWriteVerdict(t *testing.T, dir, cardID, name, verdict, sha string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Audit report\n\nProse summary.\n\n")
	if verdict != "" {
		b.WriteString("verdict: " + verdict + "\n")
	}
	if sha != "" {
		b.WriteString("audited_sha: " + sha + "\n")
	}
	frWrite(t, filepath.Join(dir, ".moai", "reports", cardID, name), b.String())
}

// frPlace inserts a card row directly, the way a fixture places a card in a
// state without walking the machine to it.
func frPlace(t *testing.T, db *FactoryDB, c Card) {
	t.Helper()
	if c.Version == 0 {
		c.Version = 1
	}
	if c.UpdatedAt == "" {
		c.UpdatedAt = "2026-09-26T00:00:00Z"
	}
	_, err := db.DB.Exec(`INSERT INTO cards(`+cardSelectColumns+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.RunID, c.CardID, c.OwnerLabel, c.State, c.Version, c.EvidencePath, c.UpdatedAt,
		c.Stage, c.LeaseHolder, c.LeaseExpiresAt, c.HeartbeatAt, c.DecisionGate, c.DecisionQuestion, c.DecisionResume,
		c.Decider, c.DecidedAt, c.FailureReason, c.HintPrefer, c.HintAfter, c.SpecID, c.WorktreePath, c.EvidenceSHA,
		c.MergeSHA, c.MergeTree, c.RemeasurePath, c.ContractSpecID, c.ContractSHA256, c.ContractSignedAt, c.ContractEvent)
	if err != nil {
		t.Fatalf("place card %s: %v", c.CardID, err)
	}
}

func frRegisterWorker(t *testing.T, db *FactoryDB, label string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,1,'2026-09-26T00:00:00Z','2026-09-26T00:00:00Z')`, label); err != nil {
		t.Fatalf("register worker %s: %v", label, err)
	}
}

// frRowDump renders a card row plus the events count, the "row and event log
// unchanged" witness used by every refusal assertion.
func frRowDump(t *testing.T, db *FactoryDB, runID, cardID string) string {
	t.Helper()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s: %v", cardID, err)
	}
	var events int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&events); err != nil {
		t.Fatalf("count events: %v", err)
	}
	return fmt.Sprintf("%+v events=%d", c, events)
}

func frEvents(t *testing.T, db *FactoryDB, kind string) []string {
	t.Helper()
	rows, err := db.DB.Query(`SELECT payload_json FROM events WHERE kind=? ORDER BY seq`, kind)
	if err != nil {
		t.Fatalf("query events: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

var frNow = time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)

func frLeaseUntil(d time.Duration) string { return frNow.Add(d).Format(time.RFC3339Nano) }

func frOpen(t *testing.T) *FactoryDB {
	t.Helper()
	return openSandboxFactory(t)
}
