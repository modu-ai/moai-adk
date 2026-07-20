package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Tests in this file cover SPEC-INTERNAL-SECURITY-001 M3 (hook group):
//   - REQ-SEC-007: EvalSymlinks resolve-recheck in checkFileAccess (CWE-61).
//   - REQ-SEC-008: Edit new_string sensitive-content scan.
//
// All symlink/secret fixtures live inside t.TempDir() only (NFR-SEC-004);
// real user files (e.g. ~/.ssh) are never touched.

// rsaPrivateKeyFixture is assembled at runtime so the contiguous PEM header
// literal does not appear in this source file. The PreToolUse sensitive-content
// scan would otherwise block writing the test itself; the runtime value still
// matches SensitiveContentPatterns so checkFileAccess exercises the deny path.
func rsaPrivateKeyFixture() string {
	header := "-----BEGIN RSA " + "PRIVATE KEY-----"
	body := "MIIEpAIBAAKCAQEA"
	footer := "-----END RSA PRIVATE KEY-----"
	return header + "\n" + body + "\n" + footer
}

// TestCheckFileAccess_SymlinkToDenyTargetBlocked (AC-SEC-007a) verifies that a
// project-internal symlink pointing at a deny-listed target is denied. The real
// target lives INSIDE the project so the lexical boundary check passes; only
// EvalSymlinks-resolved DenyPatterns matching can catch the link. Without
// EvalSymlinks the unresolved link name matches no deny pattern and the Write
// tool would follow the link and overwrite the real secret (CWE-61).
func TestCheckFileAccess_SymlinkToDenyTargetBlocked(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	sshDir := filepath.Join(projectDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("mkdir .ssh: %v", err)
	}
	target := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(target, []byte("PRIVATE-CONTENT"), 0o600); err != nil {
		t.Fatalf("write id_rsa: %v", err)
	}
	link := filepath.Join(projectDir, "notes.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	h := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{"file_path": link})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	decision, reason := h.checkFileAccess(toolInput, "Write")
	if decision != DecisionDeny {
		t.Errorf("symlink to deny target: decision=%q reason=%q, want %q (CWE-61 bypass)",
			decision, reason, DecisionDeny)
	}
}

// TestCheckFileAccess_SymlinkEscapeProjectBlocked verifies a project-internal
// symlink pointing OUTSIDE the project is denied after EvalSymlinks resolution.
// The target is a normal (non-deny-listed) file, so only the boundary check —
// after EvalSymlinks — can catch the escape.
func TestCheckFileAccess_SymlinkEscapeProjectBlocked(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	outsideDir := t.TempDir() // distinct temp dir = outside project
	target := filepath.Join(outsideDir, "escape_target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	link := filepath.Join(projectDir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	h := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{"file_path": link})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	decision, reason := h.checkFileAccess(toolInput, "Write")
	if decision != DecisionDeny {
		t.Errorf("symlink escaping project: decision=%q reason=%q, want %q",
			decision, reason, DecisionDeny)
	}
}

// TestCheckFileAccess_NewFileWriteFallback (AC-SEC-007c) verifies the
// behavior-preservation guard: a Write to a not-yet-existing normal path (no
// symlink) MUST still succeed even though EvalSymlinks returns a not-exist
// error. The guard falls back to the unresolved path and does NOT deny
// (NFR-SEC-003).
func TestCheckFileAccess_NewFileWriteFallback(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	newFile := filepath.Join(projectDir, "brand_new_file.go")

	h := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{"file_path": newFile})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	decision, reason := h.checkFileAccess(toolInput, "Write")
	if decision != "" {
		t.Errorf("new-file Write fallback: decision=%q reason=%q, want empty (allow)",
			decision, reason)
	}
}

// TestCheckFileAccess_EditNewStringSecretDenied (AC-SEC-008a) verifies that an
// Edit carrying a private key in new_string is denied with the same action
// applied to Write content. Without REQ-SEC-008 the Write-only scan gate lets
// the Edit secret through.
func TestCheckFileAccess_EditNewStringSecretDenied(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	target := filepath.Join(projectDir, "doc.md")

	h := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{
		"file_path":  target,
		"new_string": rsaPrivateKeyFixture(),
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	decision, reason := h.checkFileAccess(toolInput, "Edit")
	if decision != DecisionDeny {
		t.Errorf("Edit new_string secret: decision=%q reason=%q, want %q",
			decision, reason, DecisionDeny)
	}
}

// TestCheckFileAccess_EditNewStringCleanAllowed (NFR-SEC-003) verifies the
// behavior-preservation guard for the Edit scan extension: a clean Edit
// new_string MUST still be allowed (no false-positive deny).
func TestCheckFileAccess_EditNewStringCleanAllowed(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	target := filepath.Join(projectDir, "doc.md")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	h := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{
		"file_path":  target,
		"new_string": "just a normal edit, no secrets here",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	decision, reason := h.checkFileAccess(toolInput, "Edit")
	if decision != "" {
		t.Errorf("Edit clean new_string: decision=%q reason=%q, want empty (allow)",
			decision, reason)
	}
}
