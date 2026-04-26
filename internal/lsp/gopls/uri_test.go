package gopls

import (
	"runtime"
	"strings"
	"testing"
)

// TestPathToURI_EncodesSpaces verifies that spaces in paths are percent-encoded.
func TestPathToURI_EncodesSpaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix path test: skipping on Windows")
	}
	uri, err := pathToURI("/a b/main.go")
	if err != nil {
		t.Fatalf("pathToURI error: %v", err)
	}
	// spaces must be encoded as %20.
	if !strings.Contains(uri, "%20") {
		t.Errorf("pathToURI(%q) = %q, want %%20 encoding", "/a b/main.go", uri)
	}
	if !strings.HasPrefix(uri, "file:///") {
		t.Errorf("pathToURI(%q) = %q, want file:/// prefix", "/a b/main.go", uri)
	}
}

// TestPathToURI_EncodesUnicode verifies that Unicode paths are percent-encoded.
func TestPathToURI_EncodesUnicode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix path test: skipping on Windows")
	}
	uri, err := pathToURI("/내/main.go")
	if err != nil {
		t.Fatalf("pathToURI error: %v", err)
	}
	// Unicode characters must be percent-encoded.
	if !strings.Contains(uri, "%") {
		t.Errorf("pathToURI(%q) = %q, want percent encoding", "/내/main.go", uri)
	}
	if !strings.HasPrefix(uri, "file:///") {
		t.Errorf("pathToURI(%q) = %q, want file:/// prefix", "/내/main.go", uri)
	}
}

// TestPathToURI_WindowsDrive verifies that Windows drive paths are handled correctly.
// The logic must work correctly on Windows only; skipped on other platforms.
func TestPathToURI_WindowsDrive(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows drive path test: only runs on Windows")
	}
	uri, err := pathToURI(`C:\x\main.go`)
	if err != nil {
		t.Fatalf("pathToURI error: %v", err)
	}
	// Windows: must be in file:///C:/x/main.go format.
	want := "file:///C:/x/main.go"
	if uri != want {
		t.Errorf("pathToURI = %q, want %q", uri, want)
	}
}

// TestPathToURI_Idempotent verifies that converting the same absolute path twice yields the same result.
func TestPathToURI_Idempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix path test: skipping on Windows")
	}
	path := "/usr/local/src/main.go"
	uri1, err := pathToURI(path)
	if err != nil {
		t.Fatalf("pathToURI error: %v", err)
	}
	uri2, err := pathToURI(path)
	if err != nil {
		t.Fatalf("pathToURI(2) error: %v", err)
	}
	if uri1 != uri2 {
		t.Errorf("pathToURI idempotency violation: %q != %q", uri1, uri2)
	}
}

// TestPathToURI_RejectsEmpty verifies that an empty path is rejected.
func TestPathToURI_RejectsEmpty(t *testing.T) {
	_, err := pathToURI("")
	if err == nil {
		t.Error("empty path did not return an error")
	}
}
