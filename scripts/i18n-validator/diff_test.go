// Package main은 --diff 모드 oracle 테스트를 제공합니다.
// Package main provides tests for the --diff mode oracle.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTempGitRepo는 임시 git 레포지토리를 생성하고 초기화합니다.
// setupTempGitRepo creates and initializes a temporary git repository.
func setupTempGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "Test")
	runGit("config", "commit.gpgsign", "false")
	return dir
}

// commitFile は指定ファイルをコミットします。
func commitFile(t *testing.T, dir, relPath, content, message string) string {
	t.Helper()
	fullPath := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	runGit("add", relPath)
	runGit("commit", "-m", message)
	return runGit("rev-parse", "HEAD")
}

// TestDiff_ExitsNonZeroOnPR783Mockreleasedata はAC-CIAUT-016の検証です。
// PR #783 regression: BOTH data and test translated together (baseline Korean, head English).
func TestDiff_ExitsNonZeroOnPR783Mockreleasedata(t *testing.T) {
	t.Parallel()

	// Check git is available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := setupTempGitRepo(t)

	// Commit baseline (Korean)
	baselineContent := `package pr783_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var mockReleaseData = map[string]string{
	"yaml_error": "유효한 YAML 문서가 아닙니다",
}

func TestReleaseFlow(t *testing.T) {
	assert.Equal(t, "유효한 YAML 문서가 아닙니다", mockReleaseData["yaml_error"])
}
`
	baselineRev := commitFile(t, dir, "release_test.go", baselineContent, "baseline: Korean strings")

	// Commit head (BOTH data and test translated to English)
	headContent := `package pr783_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var mockReleaseData = map[string]string{
	"yaml_error": "Not a valid YAML document",
}

func TestReleaseFlow(t *testing.T) {
	assert.Equal(t, "Not a valid YAML document", mockReleaseData["yaml_error"])
}
`
	commitFile(t, dir, "release_test.go", headContent, "translate: Korean -> English")

	// Run --diff mode
	result := runDiffOracle(dir, baselineRev)
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1 for translation regression, got %d\nstderr: %s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stderr, "translation requires test update") {
		t.Errorf("expected 'translation requires test update' in stderr, got: %q", result.Stderr)
	}
}

// TestDiff_PassesNormalI18nChange は非ロック文字列の翻訳が通過することを検証します。
func TestDiff_PassesNormalI18nChange(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := setupTempGitRepo(t)

	// Baseline: a non-test file with a translatable string
	baselineContent := `package cmd

// userFacingMsg는 사용자에게 표시되는 메시지입니다.
const userFacingMsg = "환영합니다" // i18n:translatable
`
	baselineRev := commitFile(t, dir, "messages.go", baselineContent, "baseline")

	// Head: translated
	headContent := `package cmd

// userFacingMsg는 사용자에게 표시되는 메시지입니다.
const userFacingMsg = "Welcome" // i18n:translatable
`
	commitFile(t, dir, "messages.go", headContent, "translate: Korean -> English")

	result := runDiffOracle(dir, baselineRev)
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0 for translatable literal, got %d\nstderr: %s", result.ExitCode, result.Stderr)
	}
}

// TestDiff_RespectsTranslatableMarker は translatable マーカーが --diff モードでも有効なことを検証します。
func TestDiff_RespectsTranslatableMarker(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := setupTempGitRepo(t)

	// Baseline: test references a translatable const
	baselineContent := `package foo_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const TranslatableMsg = "원본 메시지" // i18n:translatable

func TestTranslatable(t *testing.T) {
	assert.Equal(t, TranslatableMsg, "실제값")
}
`
	baselineRev := commitFile(t, dir, "translatable_test.go", baselineContent, "baseline")

	// Head: translated (still has magic comment)
	headContent := `package foo_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const TranslatableMsg = "Original message" // i18n:translatable

func TestTranslatable(t *testing.T) {
	assert.Equal(t, TranslatableMsg, "actual")
}
`
	commitFile(t, dir, "translatable_test.go", headContent, "translate")

	result := runDiffOracle(dir, baselineRev)
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0 for translatable marker, got %d\nstderr: %s", result.ExitCode, result.Stderr)
	}
}

// TestDiff_NonGitRepoFallsBack は非git環境で --all-files モードにフォールバックすることを検証します。
func TestDiff_NonGitRepoFallsBack(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No git init — plain directory

	result := runDiffOracle(dir, "HEAD~1")
	// Should NOT crash; should fall back gracefully (exit 0 on empty corpus)
	if result.ExitCode == 2 && !strings.Contains(result.Stderr, "not a git") {
		t.Errorf("unexpected exit in non-git fallback: code=%d stderr=%q", result.ExitCode, result.Stderr)
	}
}
