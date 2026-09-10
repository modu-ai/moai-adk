package homestate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

type RuntimeCensus struct {
	ActiveSessions       int
	ActiveFactoryWorkers int
	ActiveMCPServers     int
	Fingerprint          string
}

func (c RuntimeCensus) Total() int {
	return c.ActiveSessions + c.ActiveFactoryWorkers + c.ActiveMCPServers
}

func ReadRuntimeCensus(projectRoot string) (RuntimeCensus, error) {
	return readRuntimeCensusWithWorktreeList(projectRoot, func(root string) ([]byte, error) {
		if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
			return []byte("worktree " + root + "\n"), nil
		}
		return exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
	})
}

func readRuntimeCensusWithWorktreeList(projectRoot string, listWorktrees func(string) ([]byte, error)) (RuntimeCensus, error) {
	caller, err := filepath.Abs(projectRoot)
	if err != nil {
		return RuntimeCensus{}, err
	}
	root := CanonicalProjectRoot(projectRoot)
	var c RuntimeCensus
	roots := map[string]bool{root: true, filepath.Clean(caller): true}
	out, listErr := listWorktrees(root)
	if listErr != nil {
		return RuntimeCensus{}, fmt.Errorf("worktree census inventory: %w", listErr)
	}
	foundWorktree := false
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			foundWorktree = true
			candidate, absErr := filepath.Abs(strings.TrimPrefix(line, "worktree "))
			if absErr != nil {
				return RuntimeCensus{}, fmt.Errorf("worktree census path: %w", absErr)
			}
			info, statErr := os.Stat(candidate)
			if statErr != nil || !info.IsDir() {
				return RuntimeCensus{}, fmt.Errorf("worktree census path %q is unavailable", candidate)
			}
			candidate = filepath.Clean(candidate)
			if CanonicalProjectRoot(candidate) != root {
				return RuntimeCensus{}, fmt.Errorf("worktree census canonical root mismatch: %s", candidate)
			}
			roots[candidate] = true
		}
	}
	if !foundWorktree {
		return RuntimeCensus{}, fmt.Errorf("worktree census inventory is empty or malformed")
	}
	seenSessionPID := map[int]bool{}
	for registryRoot := range roots {
		registry := filepath.Join(registryRoot, ".moai", "state", "active-sessions.json")
		raw, readErr := os.ReadFile(registry)
		if readErr == nil {
			var rows []struct {
				PID int `json:"pid"`
			}
			if err := json.Unmarshal(raw, &rows); err != nil {
				return c, fmt.Errorf("session census: %w", err)
			}
			for _, row := range rows {
				if seenSessionPID[row.PID] {
					continue
				}
				seenSessionPID[row.PID] = true
				state := platformPIDState(row.PID)
				if state == ProcessIdentityIndeterminate {
					return c, fmt.Errorf("session census indeterminate pid %d", row.PID)
				}
				if state == ProcessIdentityLive {
					c.ActiveSessions++
				}
			}
		} else if !os.IsNotExist(readErr) {
			return c, fmt.Errorf("session census: %w", readErr)
		}
	}
	factoryPath, err := FactoryDBPath(root)
	if err != nil {
		return c, err
	}
	if _, statErr := os.Stat(factoryPath); statErr == nil {
		values := url.Values{"mode": {"ro"}}
		dsn := (&url.URL{Scheme: "file", Path: filepath.ToSlash(factoryPath), RawQuery: values.Encode()}).String()
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			return c, err
		}
		rows, err := db.Query(`SELECT pid FROM workers`)
		if err != nil {
			_ = db.Close()
			return c, fmt.Errorf("factory census: %w", err)
		}
		for rows.Next() {
			var pid int
			if err := rows.Scan(&pid); err != nil {
				_ = rows.Close()
				_ = db.Close()
				return c, err
			}
			state := platformPIDState(pid)
			if state == ProcessIdentityIndeterminate {
				_ = rows.Close()
				_ = db.Close()
				return c, fmt.Errorf("factory census indeterminate pid %d", pid)
			}
			if state == ProcessIdentityLive {
				c.ActiveFactoryWorkers++
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			_ = db.Close()
			return c, fmt.Errorf("factory census: %w", err)
		}
		_ = rows.Close()
		_ = db.Close()
	} else if !os.IsNotExist(statErr) {
		return c, fmt.Errorf("factory census: %w", statErr)
	}
	mcpDir := filepath.Join(root, ".moai", "state", "mcp-server")
	if entries, err := os.ReadDir(mcpDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(mcpDir, entry.Name()))
			if err != nil {
				return c, fmt.Errorf("mcp census: %w", err)
			}
			var row struct {
				PID int `json:"pid"`
			}
			if json.Unmarshal(raw, &row) != nil || row.PID <= 0 {
				return c, fmt.Errorf("mcp census invalid %s", entry.Name())
			}
			state := platformPIDState(row.PID)
			if state == ProcessIdentityIndeterminate {
				return c, fmt.Errorf("mcp census indeterminate pid %d", row.PID)
			}
			if state == ProcessIdentityLive {
				c.ActiveMCPServers++
			}
		}
	} else if !os.IsNotExist(err) {
		return c, fmt.Errorf("mcp census: %w", err)
	}
	c.Fingerprint = strconv.Itoa(c.ActiveSessions) + ":" + strconv.Itoa(c.ActiveFactoryWorkers) + ":" + strconv.Itoa(c.ActiveMCPServers)
	return c, nil
}
