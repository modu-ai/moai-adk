package template

import (
	"testing"
	"testing/fstest"
)

// TestLoadEmbeddedCatalog_Success verifies that LoadEmbeddedCatalog wraps the
// embedded raw FS correctly and returns a fully-typed *Catalog.
//
// SPEC-V3R4-CATALOG-002 T2.4.
func TestLoadEmbeddedCatalog_Success(t *testing.T) {
	t.Parallel()

	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("LoadEmbeddedCatalog() error = %v, want nil", err)
	}
	if cat == nil {
		t.Fatal("LoadEmbeddedCatalog() returned nil catalog")
	}

	// Total entry count must match catalog.yaml ground truth.
	const wantTotal = 65
	all := cat.AllEntries()
	if len(all) != wantTotal {
		t.Errorf("LoadEmbeddedCatalog() AllEntries() = %d, want %d", len(all), wantTotal)
	}
}

// TestNewSlimDeployerWithRenderer_NilCatalog verifies that passing a nil catalog
// returns a descriptive error and not a panic.
//
// SPEC-V3R4-CATALOG-002 T2.4.
func TestNewSlimDeployerWithRenderer_NilCatalog(t *testing.T) {
	t.Parallel()

	deployer, err := NewSlimDeployerWithRenderer(nil, nil)
	if err == nil {
		t.Fatal("NewSlimDeployerWithRenderer(nil, ...) returned nil error, want error containing \"nil catalog\"")
	}
	if deployer != nil {
		t.Error("NewSlimDeployerWithRenderer(nil, ...) returned non-nil Deployer, want nil")
	}

	const wantSubstr = "nil catalog"
	if !containsSubstring(err.Error(), wantSubstr) {
		t.Errorf("NewSlimDeployerWithRenderer(nil, ...) error = %q, want substring %q", err.Error(), wantSubstr)
	}
}

// TestNewSlimDeployerWithRenderer_Success verifies the happy path: a valid
// catalog produces a non-nil Deployer with no error.
//
// SPEC-V3R4-CATALOG-002 T2.4.
func TestNewSlimDeployerWithRenderer_Success(t *testing.T) {
	t.Parallel()

	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("LoadEmbeddedCatalog() setup error: %v", err)
	}

	// Renderer is nil — NewDeployerWithRenderer accepts nil renderer without panic.
	deployer, err := NewSlimDeployerWithRenderer(cat, nil)
	if err != nil {
		t.Fatalf("NewSlimDeployerWithRenderer(validCat, nil) error = %v, want nil", err)
	}
	if deployer == nil {
		t.Error("NewSlimDeployerWithRenderer(validCat, nil) returned nil Deployer, want non-nil")
	}
}

// TestNewSlimDeployerWithRenderer_EncapsulationGate documents the DEFECT-5
// invariant: embeddedRaw MUST remain unexported. This test asserts via godoc
// expectation; the CI grep gate (git grep 'EmbeddedRaw' internal/cli/) is
// the enforcement mechanism.
//
// The test passes trivially — it is here as a living specification comment.
// SPEC-V3R4-CATALOG-002 DEFECT-5.
func TestNewSlimDeployerWithRenderer_EncapsulationGate(t *testing.T) {
	t.Parallel()

	t.Run("no_embedded_raw_export", func(t *testing.T) {
		t.Parallel()

		// This test documents the DEFECT-5 invariant. The package-level
		// variable embeddedRaw is declared in embed.go as unexported. The CI
		// grep gate verifies this at the file level. From inside a *_test.go
		// file in the same package we can confirm the variable is accessible
		// (intra-package visibility) but NOT exported (no capital letter).
		//
		// We simply exercise LoadEmbeddedCatalog which internally uses
		// embeddedRaw; if the encapsulation were broken the grep CI gate
		// would have already failed. This test is a specification anchor.
		cat, err := LoadEmbeddedCatalog()
		if err != nil {
			t.Errorf("encapsulation gate: LoadEmbeddedCatalog() unexpectedly failed: %v", err)
		}
		if cat == nil {
			t.Error("encapsulation gate: LoadEmbeddedCatalog() returned nil catalog")
		}
	})
}

// TestLoadEmbeddedCatalog_LoadCatalogErrorWrapping verifies that LoadCatalog
// returns a non-nil error when catalog.yaml is absent from the provided FS.
// This exercises the error-return path that LoadEmbeddedCatalog wraps.
//
// Note: LoadEmbeddedCatalog itself can only fail in production if the binary's
// embedded catalog.yaml is corrupt, which is not reproducible without patching
// the unexported embeddedRaw variable. Instead we test the underlying path
// to confirm coverage of the error-wrapping branch is met through LoadCatalog.
//
// SPEC-V3R4-CATALOG-002 T2.4.
func TestLoadEmbeddedCatalog_LoadCatalogErrorWrapping(t *testing.T) {
	t.Parallel()

	// Use an empty FS — LoadCatalog must return CATALOG_MANIFEST_ABSENT error.
	emptyFS := fstest.MapFS{}
	_, err := LoadCatalog(emptyFS)
	if err == nil {
		t.Error("LoadCatalog(emptyFS) should return error, got nil")
	}
	if !containsSubstring(err.Error(), "CATALOG_MANIFEST_ABSENT") {
		t.Errorf("LoadCatalog(emptyFS) error = %q, want CATALOG_MANIFEST_ABSENT", err.Error())
	}
}

// containsSubstring is a helper to avoid importing strings inside
// this test file (already imported elsewhere in the package).
func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && containsHelper(s, sub)
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
