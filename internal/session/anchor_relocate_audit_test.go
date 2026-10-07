package session

// SPEC-SESSION-ANCHOR-ATTR-001 W2 — relocation audit + ownership plausibility
// (REQ-SAA-003..006). The last-writer-wins cwd rewrite in RelocateSession was
// previously invisible: no audit row, no ownership plausibility judgment. A
// wrong-tree relocation (the t1339 shape) therefore left no trace and fed the
// disposal guard a misdirected anchor.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// auditTestRegistry builds a registry over a fresh project root with the
// given entries already on disk, and returns it with its project root.
func auditTestRegistry(t *testing.T, entries []Entry) (*Registry, string) {
	t.Helper()
	root := t.TempDir()
	writeAnchorRegistry(t, root, entries)
	return NewRegistry(filepath.Join(root, DefaultRegistryPath), nil), root
}

// readRelocationAuditLog parses the anchor-relocation audit log under the
// project root the registry path implies.
func readRelocationAuditLog(t *testing.T, projectRoot string) []RelocationAudit {
	t.Helper()
	data, err := os.ReadFile(RelocationAuditPath(projectRoot))
	if err != nil {
		t.Fatalf("read relocation audit log: %v", err)
	}
	var rows []RelocationAudit
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row RelocationAudit
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			rows = append(rows, row)
		}
	}
	return rows
}

// TestRelocateSessionAuditRowFields covers REQ-SAA-003: every cwd rewrite
// appends an audit row carrying session_id, the previous cwd, the new cwd,
// the trigger hook event, and a timestamp.
func TestRelocateSessionAuditRowFields(t *testing.T) {
	host, _ := os.Hostname()
	reg, root := auditTestRegistry(t, []Entry{
		{SessionID: "sess-audit", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: os.Getpid(), Host: host, CWD: "/tree/old"},
	})

	// The plain entry point also audits (plan M2: every relocation gets a row).
	if err := reg.RelocateSession("sess-audit", "/tree/new"); err != nil {
		t.Fatalf("RelocateSession: %v", err)
	}

	rows := readRelocationAuditLog(t, root)
	if len(rows) != 1 {
		t.Fatalf("want exactly 1 audit row, got %d", len(rows))
	}
	row := rows[0]
	if row.SessionID != "sess-audit" {
		t.Errorf("session_id = %q, want sess-audit", row.SessionID)
	}
	if row.FromCwd != "/tree/old" {
		t.Errorf("from_cwd = %q, want /tree/old", row.FromCwd)
	}
	if row.ToCwd != "/tree/new" {
		t.Errorf("to_cwd = %q, want /tree/new", row.ToCwd)
	}
	if row.Trigger == "" {
		t.Error("trigger must be recorded (non-empty), got empty")
	}
	if row.Timestamp == "" {
		t.Error("timestamp must be recorded (non-empty), got empty")
	}
	if _, err := time.Parse(time.RFC3339, row.Timestamp); err != nil {
		t.Errorf("timestamp %q is not RFC3339: %v", row.Timestamp, err)
	}
}

// TestRelocateSessionAudit_MissingEntryNoRow: no rewrite, no audit row —
// the audit records relocations, not attempts.
func TestRelocateSessionAudit_MissingEntryNoRow(t *testing.T) {
	reg, root := auditTestRegistry(t, []Entry{
		{SessionID: "sess-a", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: os.Getpid(), CWD: "/tree/old"},
	})

	if err := reg.RelocateSession("sess-unknown", "/tree/new"); err != nil {
		t.Fatalf("missing-entry relocation must stay a no-op, got: %v", err)
	}
	if _, err := os.Stat(RelocationAuditPath(root)); !os.IsNotExist(err) {
		t.Errorf("audit log must not exist without a rewrite, stat err: %v", err)
	}
}

// TestRelocateOwnershipFlag covers the REQ-SAA-004 case table, one arm per
// case. The target-tree lock reason format is the launcher's
// `claude session <card-id> (pid <n> ...)` shape.
func TestRelocateOwnershipFlag(t *testing.T) {
	host, _ := os.Hostname()

	t.Run("self-owned: lock pid == entry pid, no flag, proceeds", func(t *testing.T) {
		reg, root := auditTestRegistry(t, []Entry{
			{SessionID: "sess-1", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: 4242, Host: host, CWD: "/tree/old"},
		})
		lock := LockInfo{Locked: true, Reason: "claude session t931 (pid 4242 start Sun Sep 28)"}

		audit, err := reg.RelocateSessionWithOptions("sess-1", "/tree/other", RelocationOptions{
			Trigger: "CwdChanged", TargetLock: &lock,
		})
		if err != nil {
			t.Fatalf("relocate: %v", err)
		}
		if audit.Owner != OwnerSelf || audit.Flagged {
			t.Errorf("owner = %q flagged=%v, want self-owned without flag", audit.Owner, audit.Flagged)
		}
		if audit.HolderPID != 4242 {
			t.Errorf("holder_pid = %d, want 4242", audit.HolderPID)
		}
		entries, _ := reg.Query("")
		if entries[0].CWD != "/tree/other" {
			t.Errorf("self-owned relocation must proceed, cwd = %q", entries[0].CWD)
		}
		if rows := readRelocationAuditLog(t, root); len(rows) != 1 {
			t.Errorf("want 1 audit row, got %d", len(rows))
		}
	})

	t.Run("other live card: different pid, holder alive, flagged + proceeds", func(t *testing.T) {
		reg, _ := auditTestRegistry(t, []Entry{
			// The entry pid (4243) differs from the lock holder pid; the
			// holder is THIS test process — alive by construction, so the
			// case is another live card, not a dead-holder lock.
			{SessionID: "sess-1", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: 4243, Host: host, CWD: "/tree/old"},
		})
		otherLock := LockInfo{Locked: true, Reason: "claude session t-other (pid " + strconv.Itoa(os.Getpid()) + ")"}

		audit, err := reg.RelocateSessionWithOptions("sess-1", "/tree/other", RelocationOptions{
			Trigger: "CwdChanged", TargetLock: &otherLock,
		})
		if err != nil {
			t.Fatalf("relocate: %v", err)
		}
		if audit.Owner != OwnerOtherLive || !audit.Flagged {
			t.Errorf("owner = %q flagged=%v, want other-live flagged", audit.Owner, audit.Flagged)
		}
		if audit.HolderCardID != "t-other" {
			t.Errorf("holder_card_id = %q, want t-other", audit.HolderCardID)
		}
		entries, _ := reg.Query("")
		if entries[0].CWD != "/tree/other" {
			t.Errorf("advisory default must proceed with the relocation, cwd = %q", entries[0].CWD)
		}
	})

	t.Run("unreadable: lock reason unparsable, no flag, fail-closed reading kept", func(t *testing.T) {
		reg, _ := auditTestRegistry(t, []Entry{
			{SessionID: "sess-1", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: os.Getpid(), Host: host, CWD: "/tree/old"},
		})
		lock := LockInfo{Locked: true, Reason: "manually locked by operator"}

		audit, err := reg.RelocateSessionWithOptions("sess-1", "/tree/other", RelocationOptions{
			Trigger: "CwdChanged", TargetLock: &lock,
		})
		if err != nil {
			t.Fatalf("relocate: %v", err)
		}
		if audit.Owner != OwnerUnreadable || audit.Flagged {
			t.Errorf("owner = %q flagged=%v, want unreadable without flag", audit.Owner, audit.Flagged)
		}
		entries, _ := reg.Query("")
		if entries[0].CWD != "/tree/other" {
			t.Errorf("unreadable must not block the relocation, cwd = %q", entries[0].CWD)
		}
	})

	t.Run("registry-only: no lock, other anchored entries, flagged + proceeds", func(t *testing.T) {
		reg, _ := auditTestRegistry(t, []Entry{
			{SessionID: "sess-1", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: os.Getpid(), Host: host, CWD: "/tree/old"},
			{SessionID: "sess-other", SpecID: "SPEC-Y", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: os.Getpid(), Host: host, CWD: "/tree/other/deep"},
		})

		audit, err := reg.RelocateSessionWithOptions("sess-1", "/tree/other", RelocationOptions{
			Trigger: "CwdChanged", TargetLock: nil,
			// The caller resolves this: one OTHER live entry sits inside the
			// target tree (sess-other, cwd /tree/other/deep), so the anchor
			// evidence is registry-only.
			AnchoredOthers: 1,
		})
		if err != nil {
			t.Fatalf("relocate: %v", err)
		}
		if audit.Owner != OwnerRegistryOnly || !audit.Flagged {
			t.Errorf("owner = %q flagged=%v, want registry-only flagged", audit.Owner, audit.Flagged)
		}
		entries, _ := reg.Query("")
		if entries[0].CWD != "/tree/other" {
			t.Errorf("registry-only advisory must proceed, cwd = %q", entries[0].CWD)
		}
	})
}

// TestRelocateOwnershipGuardBlocks covers REQ-SAA-005's two arms: with the
// guard enabled a flagged relocation is refused and the refusal recorded;
// with it disabled the advisory default proceeds.
func TestRelocateOwnershipGuardBlocks(t *testing.T) {
	host, _ := os.Hostname()
	// The lock holder is THIS test process — alive by construction — while
	// every entry pid below differs from it, so both arms exercise the
	// other-live-card case rather than dead-holder.
	otherLock := LockInfo{Locked: true, Reason: "claude session t-other (pid " + strconv.Itoa(os.Getpid()) + ")"}

	t.Run("guard on: flagged relocation refused + refusal recorded", func(t *testing.T) {
		reg, root := auditTestRegistry(t, []Entry{
			{SessionID: "sess-1", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: 4243, Host: host, CWD: "/tree/old"},
		})

		_, err := reg.RelocateSessionWithOptions("sess-1", "/tree/other", RelocationOptions{
			Trigger: "CwdChanged", TargetLock: &otherLock, RefuseFlagged: true,
		})
		if !errors.Is(err, ErrRelocationRefused) {
			t.Fatalf("want ErrRelocationRefused, got %v", err)
		}
		entries, _ := reg.Query("")
		if entries[0].CWD != "/tree/old" {
			t.Errorf("refused relocation must not rewrite cwd, got %q", entries[0].CWD)
		}
		rows := readRelocationAuditLog(t, root)
		if len(rows) != 1 || !rows[0].Refused {
			t.Fatalf("refusal must be recorded in the audit log, got %+v", rows)
		}
	})

	t.Run("guard off (default): flagged relocation proceeds advisory", func(t *testing.T) {
		reg, _ := auditTestRegistry(t, []Entry{
			{SessionID: "sess-1", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: 4243, Host: host, CWD: "/tree/old"},
		})

		audit, err := reg.RelocateSessionWithOptions("sess-1", "/tree/other", RelocationOptions{
			Trigger: "CwdChanged", TargetLock: &otherLock, RefuseFlagged: false,
		})
		if err != nil {
			t.Fatalf("advisory default must proceed, got: %v", err)
		}
		if !audit.Flagged || audit.Refused {
			t.Errorf("audit = flagged=%v refused=%v, want flagged advisory without refusal", audit.Flagged, audit.Refused)
		}
	})
}

// TestParseLockCardID covers the card-id token extraction the ownership
// comparison records.
func TestParseLockCardID(t *testing.T) {
	cases := []struct {
		reason string
		want   string
		ok     bool
	}{
		{"claude session t931 (pid 4242 start Sun Aug 23)", "t931", true},
		{"moai codex session t931 (pid 4242)", "t931", true},
		{"claude session t931", "t931", true},
		{"manually locked by operator", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := ParseLockCardID(c.reason)
		if got != c.want || ok != c.ok {
			t.Errorf("ParseLockCardID(%q) = (%q, %v), want (%q, %v)", c.reason, got, ok, c.want, c.ok)
		}
	}
}

func TestProjectRootOfRegistryNativePath(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, DefaultRegistryPath)
	if got := projectRootOfRegistry(registry); got != root {
		t.Fatalf("projectRootOfRegistry(%q) = %q, want %q", registry, got, root)
	}
	if got := projectRootOfRegistry(registry + ".backup"); got != "" {
		t.Fatalf("non-registry suffix must not claim a project: %q", got)
	}
}
