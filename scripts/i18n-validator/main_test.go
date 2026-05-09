// Package main은 i18n-validator의 통합 테스트를 제공합니다.
// Package main provides integration tests for the i18n-validator tool.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- W6-T01: AST Parser tests ---

// TestParseTestFile_ExtractsTestifyEqualLiteral는 assert.Equal에서 LockedLiteral을 추출하는지 검증합니다.
func TestParseTestFile_ExtractsTestifyEqualLiteral(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/assert"

func TestFoo(t *testing.T) {
	actual := "hello"
	assert.Equal(t, "expected", actual)
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	// expect at least 1 — the expected literal at idx=0.
	// idx=1 ("actual" local var) may also produce an unresolved identifier entry;
	// Layer 2 will drop it since "actual" is not in the package symbol table.
	if len(lits) < 1 {
		t.Fatalf("expected at least 1 LockedLiteral, got %d", len(lits))
	}
	// The first entry must be the locked expected value.
	found := false
	for _, l := range lits {
		if l.Text == "expected" && l.AssertionRef.Method == "Equal" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected LockedLiteral with Text=%q Method=%q not found in %v", "expected", "Equal", lits)
	}
}

// TestParseTestFile_DetectsRequireContains는 require.Contains에서 LockedLiteral을 추출하는지 검증합니다.
func TestParseTestFile_DetectsRequireContains(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/require"

func TestContains(t *testing.T) {
	require.Contains(t, "haystack needle", "needle")
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	// require.Contains(t, haystack, needle) — both string args are extracted
	if len(lits) < 1 {
		t.Fatalf("expected at least 1 LockedLiteral, got %d", len(lits))
	}

	// Verify "needle" is captured
	found := false
	for _, l := range lits {
		if l.Text == "needle" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("needle literal not found in extracted literals: %v", lits)
	}
}

// TestParseTestFile_HandlesSuiteReceiver は suite receiver (s.Equal) を検証します。
func TestParseTestFile_HandlesSuiteReceiver(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/suite"

type MySuite struct {
	suite.Suite
}

func (s *MySuite) TestEqual() {
	actual := "got"
	s.Equal("expected", actual)
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	// expect at least 1 — the expected literal at idx=0.
	// idx=1 ("actual" local var) may produce an unresolved identifier entry; Layer 2 drops it.
	if len(lits) < 1 {
		t.Fatalf("expected at least 1 LockedLiteral, got %d: %v", len(lits), lits)
	}
	found := false
	for _, l := range lits {
		if l.Text == "expected" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected LockedLiteral with Text=%q not found in %v", "expected", lits)
	}
}

// TestParseTestFile_IgnoresNonAssertionLiterals は 非assertion呼び出しを無視します。
func TestParseTestFile_IgnoresNonAssertionLiterals(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "fmt"
import "testing"

func TestFoo(t *testing.T) {
	fmt.Println("hello")
	_ = "not an assertion"
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	if len(lits) != 0 {
		t.Errorf("expected 0 LockedLiterals for non-assertion code, got %d: %v", len(lits), lits)
	}
}

// TestParseTestFile_HandlesIdentifierReference は識別子参照を転送する。
func TestParseTestFile_HandlesIdentifierReference(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/assert"

const expectedConst = "hello"

func TestIdent(t *testing.T) {
	assert.Equal(t, expectedConst, "actual")
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	// expectedConst is an identifier ref — should produce at least 1 entry
	if len(lits) == 0 {
		t.Fatal("expected at least 1 LockedLiteral (identifier reference), got 0")
	}
}

// --- W6-T04: Magic Comment tests ---

// TestMagicComment_ExemptOnSameLine は同一行コメントで Translatable=true になることを検証。
func TestMagicComment_ExemptOnSameLine(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/assert"

const errMsg = "Failed to load config" // i18n:translatable

func TestMsg(t *testing.T) {
	assert.Equal(t, errMsg, "config not found")
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	if len(lits) == 0 {
		t.Fatal("expected 1 LockedLiteral")
	}

	// The "Failed to load config" const should be marked translatable
	for _, l := range lits {
		if l.Text == "Failed to load config" && !l.Translatable {
			t.Errorf("LockedLiteral %q should be Translatable=true (same-line comment)", l.Text)
		}
	}
}

// TestMagicComment_ExemptOnPrecedingLine は直前行コメントで Translatable=true になることを検証。
func TestMagicComment_ExemptOnPrecedingLine(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/assert"

// i18n:translatable
const errMsg2 = "Failed to load config2"

func TestMsg2(t *testing.T) {
	assert.Equal(t, errMsg2, "config not found")
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	// "Failed to load config2" should be translatable
	for _, l := range lits {
		if l.Text == "Failed to load config2" && !l.Translatable {
			t.Errorf("LockedLiteral %q should be Translatable=true (preceding line comment)", l.Text)
		}
	}
}

// TestMagicComment_RejectsTypo は typo の "translateable" (eが余分) を拒否します。
func TestMagicComment_RejectsTypo(t *testing.T) {
	t.Parallel()

	src := `package foo_test

import "testing"
import "github.com/stretchr/testify/assert"

const typoMsg = "typo message" // i18n:translateable

func TestTypo(t *testing.T) {
	assert.Equal(t, typoMsg, "x")
}
`
	f := writeTempGoFile(t, src)
	fset, astFile := parseGoFile(t, f)

	lits := extractLockedLiterals(fset, astFile, f)

	for _, l := range lits {
		if l.Text == "typo message" && l.Translatable {
			t.Errorf("LockedLiteral %q should NOT be Translatable (typo in comment token)", l.Text)
		}
	}
}

// --- W6-T05: --all-files oracle tests ---

// TestAllFiles_NoMismatch は不一致なしのケースで exit 0 を検証します。
func TestAllFiles_NoMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write a Go test file with a const that matches the assertion
	goFile := filepath.Join(dir, "normal_test.go")
	if err := os.WriteFile(goFile, []byte(`package normal_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const ExpectedMsg = "hello world"

func TestNormal(t *testing.T) {
	assert.Equal(t, ExpectedMsg, ExpectedMsg)
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	violations := runAllFilesOracle(dir)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}

// TestAllFiles_LockedLiteralMismatch は不一致時に exit 1 相当を検証します。
func TestAllFiles_LockedLiteralMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// const says "Y" but test asserts "X" (mismatch)
	if err := os.WriteFile(filepath.Join(dir, "mismatch_test.go"), []byte(`package mismatch_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const MsgY = "Y"

func TestMismatch(t *testing.T) {
	assert.Equal(t, "X", MsgY)
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	violations := runAllFilesOracle(dir)
	if len(violations) == 0 {
		t.Error("expected violations for const/test mismatch, got 0")
	}
}

// TestAllFiles_TranslatableLiteralMismatch は translatable コメントで exempt を検証します。
func TestAllFiles_TranslatableLiteralMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "translatable_test.go"), []byte(`package translatable_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const TransMsg = "Original" // i18n:translatable

func TestTranslatable(t *testing.T) {
	assert.Equal(t, "Translated", TransMsg)
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	violations := runAllFilesOracle(dir)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for translatable literal, got %d", len(violations))
	}
}

// --- W6-T07: Budget tests ---

// TestBudget_FullRepoScanWithin30Sec は実際の repo で30秒以内に完了することを検証します。
func TestBudget_FullRepoScanWithin30Sec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full repo scan in short mode")
	}
	t.Parallel()

	// Use the worktree root (determined via git rev-parse or working directory)
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Skipf("cannot determine repo root: %v", err)
	}

	start := time.Now()
	_ = runAllFilesOracle(repoRoot)
	elapsed := time.Since(start)

	if elapsed > 30*time.Second {
		t.Errorf("full repo scan took %v, want <= 30s", elapsed)
	}
}

// TestBudget_TimeoutExitOnExcess は deadline 超過時に exit code 4 と正しいメッセージを検証します。
func TestBudget_TimeoutExitOnExcess(t *testing.T) {
	t.Parallel()

	// Create a synthetic corpus with many files to trigger timeout.
	// Use unique filenames so all 100 files are written (previous "i%10"
	// pattern caused name collisions and only 10 unique files were emitted,
	// which left the 1ns budget able to finish in time on fast runners).
	dir := t.TempDir()
	for i := range 100 {
		name := filepath.Join(dir, fmt.Sprintf("file_%03d_test.go", i))
		_ = os.WriteFile(name, []byte(`package foobar_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestX(t *testing.T) {
	assert.Equal(t, "msg", "msg")
}
`), 0o644)
	}

	// Run with a near-zero budget to force timeout
	result := runWithBudget(dir, 1*time.Nanosecond)
	if result.ExitCode != 4 {
		t.Errorf("expected exit code 4 for budget exceeded, got %d", result.ExitCode)
	}
	if !strings.Contains(result.Stderr, "budget") {
		t.Errorf("expected 'budget' in stderr, got: %q", result.Stderr)
	}
}

// --- Helpers ---

func writeTempGoFile(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "test_file_test.go")
	if err := os.WriteFile(f, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}
