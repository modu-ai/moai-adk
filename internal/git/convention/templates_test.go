package convention

import (
	"regexp"
	"sort"
	"testing"
)

func TestBuiltinNames(t *testing.T) {
	names := BuiltinNames()

	if len(names) == 0 {
		t.Fatal("BuiltinNames() returned empty list")
	}

	expected := []string{"conventional-commits", "angular", "karma"}
	sort.Strings(expected)
	sort.Strings(names)

	if len(names) != len(expected) {
		t.Fatalf("BuiltinNames() length = %d, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("BuiltinNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestGetBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantNil  bool
	}{
		{name: "conventional-commits", wantName: "conventional-commits"},
		{name: "angular", wantName: "angular"},
		{name: "karma", wantName: "karma"},
		{name: "nonexistent", wantNil: true},
		{name: "", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetBuiltin(tt.name)
			if tt.wantNil {
				if cfg != nil {
					t.Errorf("GetBuiltin(%q) = %v, want nil", tt.name, cfg)
				}
				return
			}
			if cfg == nil {
				t.Fatalf("GetBuiltin(%q) = nil, want non-nil", tt.name)
			}
			if cfg.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", cfg.Name, tt.wantName)
			}
		})
	}
}

func TestBuiltinPatternsCompile(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			if cfg.Pattern == "" {
				t.Error("Pattern is empty")
			}
			_, err := regexp.Compile(cfg.Pattern)
			if err != nil {
				t.Errorf("Pattern %q failed to compile: %v", cfg.Pattern, err)
			}
		})
	}
}

func TestBuiltinConventionsHaveExamples(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			if len(cfg.Examples) == 0 {
				t.Error("convention has no examples")
			}
		})
	}
}

func TestBuiltinConventionsHaveTypes(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			if len(cfg.Types) == 0 {
				t.Error("convention has no types")
			}
		})
	}
}

func TestBuiltinExamplesMatchPatterns(t *testing.T) {
	for name, cfg := range builtinConventions {
		t.Run(name, func(t *testing.T) {
			re, err := regexp.Compile(cfg.Pattern)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			for _, example := range cfg.Examples {
				if !re.MatchString(example) {
					t.Errorf("example %q does not match pattern %q", example, cfg.Pattern)
				}
			}
		})
	}
}
