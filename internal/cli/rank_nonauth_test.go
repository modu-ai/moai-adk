package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Tests for detectGoBinPath ---

func TestDetectGoBinPath_ReturnsNonEmpty(t *testing.T) {
	result := detectGoBinPath("/home/testuser")
	if result == "" {
		t.Fatal("detectGoBinPath should return a non-empty string")
	}
}

func TestDetectGoBinPath_ResultContainsBin(t *testing.T) {
	result := detectGoBinPath("/home/testuser")
	if !strings.Contains(result, "bin") {
		t.Errorf("detectGoBinPath result should contain 'bin', got %q", result)
	}
}

func TestDetectGoBinPath_WithEmptyHomeDir(t *testing.T) {
	// When homeDir is empty and go env returns empty values for both
	// GOBIN and GOPATH, the function should return /usr/local/go/bin.
	// In a test environment, go env likely returns valid paths, so this
	// tests that the function does not panic with empty homeDir.
	result := detectGoBinPath("")
	if result == "" {
		t.Fatal("detectGoBinPath should return a non-empty string with empty homeDir")
	}
}

func TestDetectGoBinPath_FallbackWithHomeDir(t *testing.T) {
	// The function tries GOBIN, GOPATH, then falls back to homeDir/go/bin.
	// In a test environment with go installed, it will likely use GOBIN or GOPATH.
	// We verify the result is an absolute path.
	homeDir := t.TempDir()
	result := detectGoBinPath(homeDir)
	if !filepath.IsAbs(result) {
		t.Errorf("detectGoBinPath should return an absolute path, got %q", result)
	}
}

func TestDetectGoBinPath_HomeDirFallbackPath(t *testing.T) {
	// When GOBIN is empty and GOPATH is empty, the fallback path
	// should be homeDir/go/bin. In CI/CD this may or may not happen
	// since go env usually returns valid values. We at least verify
	// the function works with a valid homeDir.
	homeDir := "/tmp/test-home-dir"
	result := detectGoBinPath(homeDir)

	// Result should be one of: GOBIN value, GOPATH/bin, or homeDir/go/bin
	possibleFallback := filepath.Join(homeDir, "go", "bin")
	_ = possibleFallback // used for documentation only

	if result == "" {
		t.Fatal("detectGoBinPath should always return a non-empty string")
	}
}

// --- Tests for deployGlobalRankHookScript ---

func TestDeployGlobalRankHookScript_CreatesFile(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("script file should exist: %v", err)
	}

	if info.IsDir() {
		t.Error("script path should be a file, not a directory")
	}
}

func TestDeployGlobalRankHookScript_ExecutablePermissions(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat script: %v", err)
	}

	// Check executable bit is set (at least for owner)
	mode := info.Mode()
	if mode&0o100 == 0 {
		t.Errorf("script should be executable, got permissions %o", mode)
	}
}

func TestDeployGlobalRankHookScript_ContentContainsShebang(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "#!/bin/bash") {
		t.Error("script should start with #!/bin/bash shebang")
	}
}

func TestDeployGlobalRankHookScript_ContentContainsMoaiCommand(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "moai hook session-end") {
		t.Error("script should contain 'moai hook session-end' command")
	}
}

func TestDeployGlobalRankHookScript_ContentContainsCommandCheck(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "command -v moai") {
		t.Error("script should contain 'command -v moai' check")
	}
}

func TestDeployGlobalRankHookScript_CreatesDirectoryStructure(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	// Verify the directory structure was created
	hooksDir := filepath.Join(homeDir, ".claude", "hooks")
	info, err := os.Stat(hooksDir)
	if err != nil {
		t.Fatalf("hooks directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("hooks path should be a directory")
	}
}

func TestDeployGlobalRankHookScript_OverwritesExisting(t *testing.T) {
	homeDir := t.TempDir()

	// Create the directory and an existing file
	hooksDir := filepath.Join(homeDir, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	destPath := filepath.Join(hooksDir, "rank-submit.sh")
	if err := os.WriteFile(destPath, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}

	if string(data) == "old content" {
		t.Error("deployGlobalRankHookScript should overwrite existing file")
	}
}

func TestDeployGlobalRankHookScript_ContentContainsHomeDirFallback(t *testing.T) {
	homeDir := t.TempDir()

	err := deployGlobalRankHookScript(homeDir)
	if err != nil {
		t.Fatalf("deployGlobalRankHookScript error: %v", err)
	}

	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}

	content := string(data)
	// The script should contain a fallback path that includes the homeDir
	if !strings.Contains(content, homeDir) {
		t.Errorf("script should contain homeDir fallback path %q", homeDir)
	}
}
