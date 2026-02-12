package convention

import (
	"testing"
)

func TestNewManager_CreatesInstance(t *testing.T) {
	m := NewManager("/some/path")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Convention() != nil {
		t.Error("Convention() should be nil before loading")
	}
}

func TestManager_LoadConvention_Builtin(t *testing.T) {
	tests := []struct {
		name    string
		conv    string
		wantErr bool
	}{
		{"conventional-commits", "conventional-commits", false},
		{"angular", "angular", false},
		{"karma", "karma", false},
		{"unknown", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager("/unused")
			err := m.LoadConvention(tt.conv)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConvention(%q) error = %v, wantErr %v", tt.conv, err, tt.wantErr)
				return
			}
			if !tt.wantErr && m.Convention() == nil {
				t.Error("Convention() should not be nil after loading builtin")
			}
			if !tt.wantErr && m.Convention().Name != tt.conv {
				t.Errorf("Convention().Name = %q, want %q", m.Convention().Name, tt.conv)
			}
		})
	}
}

func TestManager_LoadConvention_Auto(t *testing.T) {
	repoRoot := findGitRoot(t)

	m := NewManager(repoRoot)
	err := m.LoadConvention("auto")
	if err != nil {
		t.Fatalf("LoadConvention(auto) error = %v", err)
	}
	if m.Convention() == nil {
		t.Error("Convention() should not be nil after auto-detection")
	}
}

func TestManager_LoadConvention_AutoFallback(t *testing.T) {
	// Use a temp dir with no git history -- auto should fallback to
	// conventional-commits.
	tmpDir := t.TempDir()

	m := NewManager(tmpDir)
	err := m.LoadConvention("auto")
	if err != nil {
		t.Fatalf("LoadConvention(auto) with fallback error = %v", err)
	}
	if m.Convention() == nil {
		t.Fatal("Convention() should not be nil after auto-fallback")
	}
	if m.Convention().Name != "conventional-commits" {
		t.Errorf("Convention().Name = %q, want %q", m.Convention().Name, "conventional-commits")
	}
}

func TestManager_LoadFromConfig_Valid(t *testing.T) {
	m := NewManager("/unused")
	err := m.LoadFromConfig(ConventionConfig{
		Name:    "custom",
		Pattern: `^(feat|fix): .+`,
		Types:   []string{"feat", "fix"},
	})
	if err != nil {
		t.Fatalf("LoadFromConfig error = %v", err)
	}
	if m.Convention() == nil {
		t.Error("Convention() should not be nil")
	}
	if m.Convention().Name != "custom" {
		t.Errorf("Convention().Name = %q, want %q", m.Convention().Name, "custom")
	}
}

func TestManager_LoadFromConfig_Invalid(t *testing.T) {
	m := NewManager("/unused")
	err := m.LoadFromConfig(ConventionConfig{
		Name:    "bad",
		Pattern: `^(unclosed`,
	})
	if err == nil {
		t.Error("LoadFromConfig should fail with invalid regex")
	}
	if m.Convention() != nil {
		t.Error("Convention() should remain nil after failed load")
	}
}

func TestManager_LoadFromConfig_EmptyPattern(t *testing.T) {
	m := NewManager("/unused")
	err := m.LoadFromConfig(ConventionConfig{
		Name:    "empty",
		Pattern: "",
	})
	if err == nil {
		t.Error("LoadFromConfig should fail with empty pattern")
	}
}

func TestManager_ValidateMessage(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("conventional-commits"); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}

	tests := []struct {
		name    string
		message string
		valid   bool
	}{
		{"valid feat", "feat(auth): add JWT", true},
		{"valid fix", "fix: resolve bug", true},
		{"invalid", "random message", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.ValidateMessage(tt.message)
			if result.Valid != tt.valid {
				t.Errorf("ValidateMessage(%q).Valid = %v, want %v", tt.message, result.Valid, tt.valid)
			}
		})
	}
}

func TestManager_ValidateMessage_NoConvention(t *testing.T) {
	m := NewManager("/unused")
	result := m.ValidateMessage("anything goes")
	if !result.Valid {
		t.Error("ValidateMessage with no convention should always be valid")
	}
}

func TestManager_ValidateMessages(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("conventional-commits"); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}

	messages := []string{
		"feat: add feature",
		"bad message",
		"fix: resolve bug",
	}

	results := m.ValidateMessages(messages)
	if len(results) != 3 {
		t.Fatalf("ValidateMessages returned %d results, want 3", len(results))
	}

	if !results[0].Valid {
		t.Error("results[0] should be valid")
	}
	if results[1].Valid {
		t.Error("results[1] should be invalid")
	}
	if !results[2].Valid {
		t.Error("results[2] should be valid")
	}
}

func TestManager_ValidateMessages_Empty(t *testing.T) {
	m := NewManager("/unused")
	results := m.ValidateMessages(nil)
	if len(results) != 0 {
		t.Errorf("ValidateMessages(nil) returned %d results, want 0", len(results))
	}
}

func TestManager_Convention_ReturnsNilBeforeLoad(t *testing.T) {
	m := NewManager("/unused")
	if m.Convention() != nil {
		t.Error("Convention() should be nil before any load")
	}
}

func TestManager_Convention_ReturnsLoadedConvention(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("angular"); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}
	conv := m.Convention()
	if conv == nil {
		t.Fatal("Convention() should not be nil after loading")
	}
	if conv.Name != "angular" {
		t.Errorf("Convention().Name = %q, want %q", conv.Name, "angular")
	}
}
