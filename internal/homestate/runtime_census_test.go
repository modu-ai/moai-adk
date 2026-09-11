package homestate

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRuntimeCensusFailsClosedWhenWorktreeInventoryUnavailable(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	linked := filepath.Join(base, "linked")
	for _, args := range [][]string{{"init", "-q", primary}, {"-C", primary, "config", "user.email", "test@example.com"}, {"-C", primary, "config", "user.name", "Test"}, {"-C", primary, "commit", "--allow-empty", "-qm", "init"}, {"-C", primary, "worktree", "add", "-qb", "linked-failure", linked}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	t.Setenv("MOAI_HOME", filepath.Join(base, "home"))
	registry := filepath.Join(linked, ".moai", "state", "active-sessions.json")
	if err := os.MkdirAll(filepath.Dir(registry), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registry, []byte(`[{"pid":`+strconv.Itoa(os.Getpid())+`}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	census, err := readRuntimeCensusWithWorktreeList(primary, func(string) ([]byte, error) { return nil, errors.New("inventory unavailable") })
	if err == nil {
		t.Fatalf("worktree inventory failure returned determinate false-zero census: %+v", census)
	}
}

func TestRuntimeCensusFailsClosedOnMalformedOrUnavailableWorktreeInventory(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	for name, list := range map[string]func(string) ([]byte, error){
		"empty":              func(string) ([]byte, error) { return nil, nil },
		"malformed":          func(string) ([]byte, error) { return []byte("garbage\n"), nil },
		"missing-root":       func(string) ([]byte, error) { return []byte("worktree " + filepath.Join(root, "missing") + "\n"), nil },
		"canonical-mismatch": func(string) ([]byte, error) { return []byte("worktree " + other + "\n"), nil },
	} {
		t.Run(name, func(t *testing.T) {
			if census, err := readRuntimeCensusWithWorktreeList(root, list); err == nil {
				t.Fatalf("inventory %s returned determinate census %+v", name, census)
			}
		})
	}
}

func TestRuntimeCensusCountsLiveAndIgnoresProvablyDead(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	state := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(filepath.Join(state, "mcp-server"), 0o700); err != nil {
		t.Fatal(err)
	}
	pid := strconv.Itoa(os.Getpid())
	if err := os.WriteFile(filepath.Join(state, "active-sessions.json"), []byte(`[{"pid":`+pid+`},{"pid":-1}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "mcp-server", "live.json"), []byte(`{"pid":`+pid+`}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "mcp-server", "ignored.txt"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(state, "mcp-server", "ignored.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	factory, err := OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := factory.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES('live',?,?,?),('dead',-1,?,?)`, os.Getpid(), now, now, now, now); err != nil {
		t.Fatal(err)
	}
	if err := factory.Close(); err != nil {
		t.Fatal(err)
	}
	census, err := ReadRuntimeCensus(root)
	if err != nil {
		t.Fatal(err)
	}
	if census.ActiveSessions != 1 || census.ActiveFactoryWorkers != 1 || census.ActiveMCPServers != 1 || census.Total() != 3 || census.Fingerprint != "1:1:1" {
		t.Fatalf("census=%+v", census)
	}
}

func TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	linked := filepath.Join(base, "linked")
	for _, args := range [][]string{{"init", "-q", primary}, {"-C", primary, "config", "user.email", "test@example.com"}, {"-C", primary, "config", "user.name", "Test"}, {"-C", primary, "commit", "--allow-empty", "-qm", "init"}, {"-C", primary, "worktree", "add", "-qb", "linked-test", linked}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	t.Setenv("MOAI_HOME", filepath.Join(base, "home"))
	registry := filepath.Join(linked, ".moai", "state", "active-sessions.json")
	if err := os.MkdirAll(filepath.Dir(registry), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registry, []byte(`[{"pid":`+strconv.Itoa(os.Getpid())+`}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	census, err := ReadRuntimeCensus(linked)
	if err != nil || census.ActiveSessions != 1 {
		t.Fatalf("census=%+v err=%v", census, err)
	}
}

func TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	linked := filepath.Join(base, "linked")
	for _, args := range [][]string{{"init", "-q", primary}, {"-C", primary, "config", "user.email", "test@example.com"}, {"-C", primary, "config", "user.name", "Test"}, {"-C", primary, "commit", "--allow-empty", "-qm", "init"}, {"-C", primary, "worktree", "add", "-qb", "linked-dedup", linked}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	t.Setenv("MOAI_HOME", filepath.Join(base, "home"))
	raw := []byte(`[{"pid":` + strconv.Itoa(os.Getpid()) + `}]`)
	for _, root := range []string{primary, linked} {
		registry := filepath.Join(root, ".moai", "state", "active-sessions.json")
		if err := os.MkdirAll(filepath.Dir(registry), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(registry, raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	first, err := ReadRuntimeCensus(linked)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ReadRuntimeCensus(primary)
	if err != nil {
		t.Fatal(err)
	}
	if first.ActiveSessions != 1 || second.ActiveSessions != 1 {
		t.Fatalf("dedup failed: linked=%+v primary=%+v", first, second)
	}
	if first.Fingerprint != "1:0:0" || second.Fingerprint != first.Fingerprint {
		t.Fatalf("unstable fingerprint: linked=%q primary=%q", first.Fingerprint, second.Fingerprint)
	}
}

func TestRuntimeCensusRejectsCorruptRegistries(t *testing.T) {
	for _, fixture := range []struct {
		name string
		make func(t *testing.T, root string)
		want string
	}{
		{"session", func(t *testing.T, root string) {
			p := filepath.Join(root, ".moai", "state", "active-sessions.json")
			if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("{"), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "session census"},
		{"factory", func(t *testing.T, root string) {
			p, err := FactoryDBPath(root)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("not sqlite"), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "factory census"},
		{"mcp", func(t *testing.T, root string) {
			p := filepath.Join(root, ".moai", "state", "mcp-server", "bad.json")
			if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(`{"pid":0}`), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "mcp census"},
		{"session-unreadable", func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, ".moai", "state", "active-sessions.json"), 0o700); err != nil {
				t.Fatal(err)
			}
		}, "session census"},
		{"mcp-unreadable", func(t *testing.T, root string) {
			p := filepath.Join(root, ".moai", "state", "mcp-server")
			if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("not a directory"), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "mcp census"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
			fixture.make(t, root)
			_, err := ReadRuntimeCensus(root)
			if err == nil || !strings.Contains(err.Error(), fixture.want) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
