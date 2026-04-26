//go:build integration

package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestConstitutionList_RealRegistry is an integration test using the actual zone-registry.md.
// Runs only in environments where the actual registry file exists.
// Related to AC-CON-001-001, AC-CON-001-002.
func TestConstitutionList_RealRegistry(t *testing.T) {
	// Find the actual zone-registry.md path
	registryPath := os.Getenv("MOAI_CONSTITUTION_REGISTRY")
	if registryPath == "" {
		// Search relative to the project root
		cwd, err := os.Getwd()
		if err != nil {
			t.Skipf("cannot determine cwd: %v", err)
		}
		// Walk up from internal/cli to the project root
		projectRoot := filepath.Join(cwd, "..", "..")
		registryPath = filepath.Join(projectRoot, ".claude", "rules", "moai", "core", "zone-registry.md")
	}

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		t.Skipf("zone-registry.md not found at %q - skipping integration test", registryPath)
	}

	projectDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(registryPath)))))

	t.Run("list all entries", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, nil, "", "table")
		if err != nil {
			t.Fatalf("runConstitutionList error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "CONST-V3R2-001") {
			t.Errorf("actual registry must contain CONST-V3R2-001\noutput: %s", output)
		}
	})

	t.Run("filter frozen zone", func(t *testing.T) {
		frozen := constitution.ZoneFrozen
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, &frozen, "", "table")
		if err != nil {
			t.Fatalf("--zone frozen error: %v", err)
		}
		output := buf.String()
		if strings.Contains(output, "Evolvable") {
			t.Errorf("Frozen filter result must not contain Evolvable")
		}
	})

	t.Run("json format valid", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, nil, "", "json")
		if err != nil {
			t.Fatalf("--format json error: %v", err)
		}
		var result struct {
			Entries []map[string]any `json:"entries"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("JSON parse error: %v\noutput: %s", err, buf.String())
		}
		if len(result.Entries) == 0 {
			t.Error("actual registry JSON must have entries")
		}
		t.Logf("actual registry: %d entries", len(result.Entries))
	})

	t.Run("minimum frozen entries", func(t *testing.T) {
		frozen := constitution.ZoneFrozen
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, &frozen, "", "json")
		if err != nil {
			t.Fatalf("--zone frozen --format json error: %v", err)
		}
		var result struct {
			Entries []map[string]any `json:"entries"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("JSON parse error: %v", err)
		}
		// AC-CON-001-006: must have at least 7 Frozen invariant clauses
		const minFrozen = 7
		if len(result.Entries) < minFrozen {
			t.Errorf("Frozen entries = %d, want >= %d (7 canonical invariants)", len(result.Entries), minFrozen)
		}
	})
}

// TestConstitutionGuard_RealRegistry is a guard integration test using the actual registry.
func TestConstitutionGuard_RealRegistry(t *testing.T) {
	registryPath := os.Getenv("MOAI_CONSTITUTION_REGISTRY")
	if registryPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			t.Skipf("cannot determine cwd: %v", err)
		}
		projectRoot := filepath.Join(cwd, "..", "..")
		registryPath = filepath.Join(projectRoot, ".claude", "rules", "moai", "core", "zone-registry.md")
	}

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		t.Skipf("zone-registry.md not found at %q - skipping integration test", registryPath)
	}

	projectDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(registryPath)))))

	t.Run("no violations is OK", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionGuard(&buf, io.Discard, projectDir, registryPath, []string{})
		if err != nil {
			t.Errorf("expected nil when no violations: %v", err)
		}
	})

	t.Run("CONST-V3R2-001 is Frozen violation", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionGuard(&buf, io.Discard, projectDir, registryPath, []string{"CONST-V3R2-001"})
		if err == nil {
			t.Error("changing CONST-V3R2-001 is a Frozen zone violation and must return an error")
		}
	})
}
