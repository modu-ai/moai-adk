package spec

import (
	"testing"
)

// TestAcceptance_ValidateID_Valid tests valid AC ID formats
func TestAcceptance_ValidateID_Valid(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"standard format", "AC-SPC-001-01"},
		{"with domain prefix", "AC-AUTH-002-15"},
		{"numeric domain", "AC-123-003-99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &Acceptance{ID: tt.id}
			if err := ac.ValidateID(); err != nil {
				t.Errorf("ValidateID() error = %v, want nil", err)
			}
		})
	}
}

// TestAcceptance_ValidateID_Invalid tests invalid AC ID formats
func TestAcceptance_ValidateID_Invalid(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"missing prefix", "SPC-001-01"},
		{"missing domain", "AC--001-01"},
		{"missing sequence", "AC-SPC-01"},
		{"lowercase", "ac-spc-001-01"},
		{"with child suffix", "AC-SPC-001-01.a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &Acceptance{ID: tt.id}
			if err := ac.ValidateID(); err == nil {
				t.Errorf("ValidateID() expected error for %s, got nil", tt.id)
			}
		})
	}
}

// TestAcceptance_GenerateChildID_Level1 tests level 1 child ID generation
func TestAcceptance_GenerateChildID_Level1(t *testing.T) {
	parent := &Acceptance{ID: "AC-SPC-001-05"}

	tests := []struct {
		index      int
		wantID     string
		expectErr  bool
	}{
		{0, "AC-SPC-001-05.a", false},
		{1, "AC-SPC-001-05.b", false},
		{25, "AC-SPC-001-05.z", false},
		{26, "", true},  // out of range
		{-1, "", true}, // negative
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gotID, err := parent.GenerateChildID(1, tt.index)
			if tt.expectErr {
				if err == nil {
					t.Errorf("GenerateChildID() expected error for index %d, got nil", tt.index)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateChildID() error = %v", err)
				}
				if gotID != tt.wantID {
					t.Errorf("GenerateChildID() = %v, want %v", gotID, tt.wantID)
				}
			}
		})
	}
}

// TestAcceptance_GenerateChildID_Level2 tests level 2 child ID generation
func TestAcceptance_GenerateChildID_Level2(t *testing.T) {
	parent := &Acceptance{ID: "AC-SPC-001-05.a"}

	tests := []struct {
		index      int
		wantID     string
		expectErr  bool
	}{
		{0, "AC-SPC-001-05.a.i", false},
		{1, "AC-SPC-001-05.a.ii", false},
		{10, "AC-SPC-001-05.a.xi", false},
		{25, "AC-SPC-001-05.a.xxvi", false},
		{26, "", true},  // out of range
		{-1, "", true}, // negative
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gotID, err := parent.GenerateChildID(2, tt.index)
			if tt.expectErr {
				if err == nil {
					t.Errorf("GenerateChildID() expected error for index %d, got nil", tt.index)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateChildID() error = %v", err)
				}
				if gotID != tt.wantID {
					t.Errorf("GenerateChildID() = %v, want %v", gotID, tt.wantID)
				}
			}
		})
	}
}

// TestAcceptance_Depth tests depth calculation
func TestAcceptance_Depth(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{"top level", "AC-SPC-001-05", 0},
		{"level 1 child", "AC-SPC-001-05.a", 1},
		{"level 1 child b", "AC-SPC-001-05.b", 1},
		{"level 2 child", "AC-SPC-001-05.a.i", 2},
		{"level 2 child xx", "AC-SPC-001-05.a.xx", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &Acceptance{ID: tt.id}
			if got := ac.Depth(); got != tt.want {
				t.Errorf("Depth() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAcceptance_InheritGiven tests Given inheritance
func TestAcceptance_InheritGiven(t *testing.T) {
	parent := &Acceptance{
		ID:    "AC-SPC-001-05",
		Given: "user is authenticated",
	}

	tests := []struct {
		name           string
		childGiven     string
		expectedGiven  string
	}{
		{"empty child inherits", "", "user is authenticated"},
		{"non-empty child keeps own", "user is admin", "user is admin"},
		{"nil parent", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := &Acceptance{
				ID:    "AC-SPC-001-05.a",
				Given: tt.childGiven,
			}

			if tt.name == "nil parent" {
				child.InheritGiven(nil)
			} else {
				child.InheritGiven(parent)
			}

			if child.Given != tt.expectedGiven {
				t.Errorf("InheritGiven() = %v, want %v", child.Given, tt.expectedGiven)
			}
		})
	}
}

// TestAcceptance_IsLeaf tests leaf node detection
func TestAcceptance_IsLeaf(t *testing.T) {
	tests := []struct {
		name     string
		children []Acceptance
		want     bool
	}{
		{"no children", nil, true},
		{"empty children", []Acceptance{}, true},
		{"with children", []Acceptance{{ID: "child"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &Acceptance{ID: "AC-SPC-001-05", Children: tt.children}
			if got := ac.IsLeaf(); got != tt.want {
				t.Errorf("IsLeaf() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAcceptance_CountLeaves tests leaf counting
func TestAcceptance_CountLeaves(t *testing.T) {
	tests := []struct {
		name     string
		tree     Acceptance
		expected int
	}{
		{
			name:     "single leaf",
			tree:     Acceptance{ID: "AC-SPC-001-05"},
			expected: 1,
		},
		{
			name: "parent with two children",
			tree: Acceptance{
				ID: "AC-SPC-001-05",
				Children: []Acceptance{
					{ID: "AC-SPC-001-05.a"},
					{ID: "AC-SPC-001-05.b"},
				},
			},
			expected: 2,
		},
		{
			name: "nested structure",
			tree: Acceptance{
				ID: "AC-SPC-001-05",
				Children: []Acceptance{
					{
						ID: "AC-SPC-001-05.a",
						Children: []Acceptance{
							{ID: "AC-SPC-001-05.a.i"},
							{ID: "AC-SPC-001-05.a.ii"},
						},
					},
					{ID: "AC-SPC-001-05.b"},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tree.CountLeaves(); got != tt.expected {
				t.Errorf("CountLeaves() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAcceptance_ValidateDepth tests depth validation
func TestAcceptance_ValidateDepth(t *testing.T) {
	tests := []struct {
		name        string
		tree        Acceptance
		expectError bool
	}{
		{
			name:        "valid depth 0",
			tree:        Acceptance{ID: "AC-SPC-001-05"},
			expectError: false,
		},
		{
			name: "valid depth 1",
			tree: Acceptance{
				ID: "AC-SPC-001-05.a",
			},
			expectError: false,
		},
		{
			name: "valid depth 2",
			tree: Acceptance{
				ID: "AC-SPC-001-05.a.i",
			},
			expectError: false,
		},
		{
			name: "invalid depth 3",
			tree: Acceptance{
				ID: "AC-SPC-001-05.a.i.1", // depth 3 (invalid)
			},
			expectError: true,
		},
		{
			name: "child exceeds max depth",
			tree: Acceptance{
				ID: "AC-SPC-001-05",
				Children: []Acceptance{
					{
						ID: "AC-SPC-001-05.a",
						Children: []Acceptance{
							{
								ID: "AC-SPC-001-05.a.i",
								Children: []Acceptance{
									{ID: "AC-SPC-001-05.a.i.1"}, // depth 3
								},
							},
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tree.ValidateDepth()
			if tt.expectError && err == nil {
				t.Errorf("ValidateDepth() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateDepth() unexpected error = %v", err)
			}
		})
	}
}

// TestExtractRequirementMappings tests REQ extraction from text
func TestExtractRequirementMappings(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single mapping",
			text:     "Verify authentication (maps REQ-AUTH-001)",
			expected: []string{"AUTH-001"},
		},
		{
			name:     "multiple mappings",
			text:     "Verify auth (maps REQ-AUTH-001) and session (maps REQ-SESS-002)",
			expected: []string{"AUTH-001", "SESS-002"},
		},
		{
			name:     "duplicates removed",
			text:     "Check (maps REQ-AUTH-001) and (maps REQ-AUTH-001) again",
			expected: []string{"AUTH-001"},
		},
		{
			name:     "no mappings",
			text:     "Just regular text",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractRequirementMappings(tt.text)
			if len(got) != len(tt.expected) {
				t.Errorf("ExtractRequirementMappings() length = %v, want %v", len(got), len(tt.expected))
				return
			}
			for i, reqID := range got {
				if reqID != tt.expected[i] {
					t.Errorf("ExtractRequirementMappings()[%d] = %v, want %v", i, reqID, tt.expected[i])
				}
			}
		})
	}
}

// TestAcceptance_ValidateRequirementMappings tests REQ mapping validation
func TestAcceptance_ValidateRequirementMappings(t *testing.T) {
	tests := []struct {
		name        string
		tree        Acceptance
		expectError bool
	}{
		{
			name:        "leaf with mapping",
			tree:        Acceptance{ID: "AC-SPC-001-05", RequirementIDs: []string{"SPC-001"}},
			expectError: false,
		},
		{
			name:        "leaf without mapping",
			tree:        Acceptance{ID: "AC-SPC-001-05"},
			expectError: true,
		},
		{
			name: "parent without mapping but children have it",
			tree: Acceptance{
				ID: "AC-SPC-001-05",
				Children: []Acceptance{
					{ID: "AC-SPC-001-05.a", RequirementIDs: []string{"SPC-001"}},
					{ID: "AC-SPC-001-05.b", RequirementIDs: []string{"SPC-002"}},
				},
			},
			expectError: false,
		},
		{
			name: "nested structure with all mappings",
			tree: Acceptance{
				ID: "AC-SPC-001-05",
				Children: []Acceptance{
					{
						ID: "AC-SPC-001-05.a",
						Children: []Acceptance{
							{ID: "AC-SPC-001-05.a.i", RequirementIDs: []string{"SPC-001"}},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "nested structure with missing mapping",
			tree: Acceptance{
				ID: "AC-SPC-001-05",
				Children: []Acceptance{
					{
						ID: "AC-SPC-001-05.a",
						Children: []Acceptance{
							{ID: "AC-SPC-001-05.a.i"}, // missing mapping
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.tree.ValidateRequirementMappings()
			hasError := len(errors) > 0
			if hasError != tt.expectError {
				t.Errorf("ValidateRequirementMappings() hasError = %v, want %v, errors: %v", hasError, tt.expectError, errors)
			}
		})
	}
}
