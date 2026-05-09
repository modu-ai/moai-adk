// Package main은 i18n-validator의 lockset 및 cross-file resolver 테스트를 제공합니다.
// Package main provides tests for the lockset and cross-file resolver.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

// --- W6-T02: Cross-File Resolver tests ---

// TestResolveIdentifier_ConstFromSamePkg はパッケージレベル定数を解決します。
func TestResolveIdentifier_ConstFromSamePkg(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// decl file
	if err := os.WriteFile(filepath.Join(dir, "messages.go"), []byte(`package foo

const HelloMsg = "hello world"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	// test file referencing the const
	if err := os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte(`package foo

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHello(t *testing.T) {
	assert.Equal(t, HelloMsg, "actual")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	builder := newLocksetBuilder()
	lockset := builder.build(dir)

	// Should have at least one locked literal for HelloMsg
	if len(lockset.Literals) == 0 {
		t.Fatal("expected at least 1 LockedLiteral for const reference, got 0")
	}
	found := false
	for _, l := range lockset.Literals {
		if l.Text == "hello world" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("'hello world' literal not found in lockset: %v", lockset.Literals)
	}
}

// TestResolveIdentifier_MapLiteralValue はマップリテラルの値を解決します。
func TestResolveIdentifier_MapLiteralValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data_test.go"), []byte(`package foo_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var mockData = map[string]string{
	"title": "유효한 YAML 문서가 아닙니다",
}

func TestData(t *testing.T) {
	assert.Equal(t, mockData["title"], "actual")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	builder := newLocksetBuilder()
	lockset := builder.build(dir)

	// mockData["title"] = "유효한 YAML 문서가 아닙니다" should be locked
	found := false
	for _, l := range lockset.Literals {
		if l.Text == "유효한 YAML 문서가 아닙니다" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("map literal value not resolved in lockset: %v", lockset.Literals)
	}
}

// TestResolveIdentifier_ExternalPackage は外部パッケージ識別子をスキップします。
func TestResolveIdentifier_ExternalPackage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ext_test.go"), []byte(`package foo_test

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
)

var ErrExternal = errors.New("external error")

func TestExternal(t *testing.T) {
	assert.Equal(t, ErrExternal.Error(), "external error")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not panic; graceful skip for external identifiers
	builder := newLocksetBuilder()
	lockset := builder.build(dir)
	_ = lockset // no crash is the assertion
}

// TestResolveIdentifier_NotFoundGracefulSkip は未宣言識別子をスキップします。
func TestResolveIdentifier_NotFoundGracefulSkip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notfound_test.go"), []byte(`package foo_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	assert.Equal(t, undeclaredConst, "actual")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not panic; graceful skip
	builder := newLocksetBuilder()
	lockset := builder.build(dir)
	_ = lockset
}

// --- W6-T03: Lockset Builder tests ---

// TestBuildLockset_FullCorpusScan は複数ファイルから lockset を構築します。
func TestBuildLockset_FullCorpusScan(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "corpus_test.go"), []byte(`package corpus_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const Msg1 = "message one"
const Msg2 = "message two"

func TestOne(t *testing.T) {
	assert.Equal(t, Msg1, "x")
}

func TestTwo(t *testing.T) {
	assert.Equal(t, Msg2, "y")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	builder := newLocksetBuilder()
	lockset := builder.build(dir)

	if len(lockset.Literals) < 2 {
		t.Errorf("expected at least 2 locked literals, got %d", len(lockset.Literals))
	}
}

// TestBuildLockset_SkipsExclusions は vendor/ を無視することを検証します。
func TestBuildLockset_SkipsExclusions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// vendor file — should be skipped
	vendorDir := filepath.Join(dir, "vendor", "pkg")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vendorDir, "vendor_test.go"), []byte(`package pkg_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestVendor(t *testing.T) {
	assert.Equal(t, "vendor string", "x")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	builder := newLocksetBuilder()
	lockset := builder.build(dir)

	for _, l := range lockset.Literals {
		if l.Text == "vendor string" {
			t.Errorf("vendor string should be excluded from lockset")
		}
	}
}

// TestBuildLockset_DuplicateLiteralsAtDifferentSites は同一定数への複数参照を検証します。
func TestBuildLockset_DuplicateLiteralsAtDifferentSites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dup_test.go"), []byte(`package dup_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const SharedMsg = "shared message"

func TestA(t *testing.T) {
	assert.Equal(t, SharedMsg, "x")
}

func TestB(t *testing.T) {
	assert.Equal(t, SharedMsg, "y")
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	builder := newLocksetBuilder()
	lockset := builder.build(dir)

	// Only 1 entry per declaration site (first seen wins)
	count := 0
	for _, l := range lockset.Literals {
		if l.Text == "shared message" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 lockset entry for shared const, got %d", count)
	}
}
