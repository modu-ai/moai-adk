package convention

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ConventionConfig
		wantErr bool
		check   func(t *testing.T, conv *Convention)
	}{
		{
			name: "valid config",
			cfg: ConventionConfig{
				Name:      "test",
				Pattern:   `^(feat|fix): .+`,
				Types:     []string{"feat", "fix"},
				MaxLength: 72,
				Examples:  []string{"feat: example"},
			},
			wantErr: false,
			check: func(t *testing.T, conv *Convention) {
				if conv.Name != "test" {
					t.Errorf("Name = %q, want %q", conv.Name, "test")
				}
				if conv.MaxLength != 72 {
					t.Errorf("MaxLength = %d, want %d", conv.MaxLength, 72)
				}
				if len(conv.Types) != 2 {
					t.Errorf("Types length = %d, want %d", len(conv.Types), 2)
				}
				if conv.Pattern == nil {
					t.Error("Pattern is nil")
				}
			},
		},
		{
			name: "empty pattern returns error",
			cfg: ConventionConfig{
				Name:    "empty",
				Pattern: "",
			},
			wantErr: true,
		},
		{
			name: "invalid regex returns error",
			cfg: ConventionConfig{
				Name:    "bad-regex",
				Pattern: `^(unclosed`,
			},
			wantErr: true,
		},
		{
			name: "default max length when zero",
			cfg: ConventionConfig{
				Name:      "no-max",
				Pattern:   `^.+`,
				MaxLength: 0,
			},
			wantErr: false,
			check: func(t *testing.T, conv *Convention) {
				if conv.MaxLength != 100 {
					t.Errorf("MaxLength = %d, want default 100", conv.MaxLength)
				}
			},
		},
		{
			name: "default max length when negative",
			cfg: ConventionConfig{
				Name:      "neg-max",
				Pattern:   `^.+`,
				MaxLength: -5,
			},
			wantErr: false,
			check: func(t *testing.T, conv *Convention) {
				if conv.MaxLength != 100 {
					t.Errorf("MaxLength = %d, want default 100", conv.MaxLength)
				}
			},
		},
		{
			name: "preserves scopes and required",
			cfg: ConventionConfig{
				Name:     "full",
				Pattern:  `^.+`,
				Scopes:   []string{"api", "cli"},
				Required: []string{"type", "description"},
			},
			wantErr: false,
			check: func(t *testing.T, conv *Convention) {
				if len(conv.Scopes) != 2 {
					t.Errorf("Scopes length = %d, want %d", len(conv.Scopes), 2)
				}
				if len(conv.Required) != 2 {
					t.Errorf("Required length = %d, want %d", len(conv.Required), 2)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv, err := Parse(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && conv != nil {
				tt.check(t, conv)
			}
		})
	}
}

func TestParseBuiltin(t *testing.T) {
	builtins := BuiltinNames()
	for _, name := range builtins {
		t.Run(name, func(t *testing.T) {
			conv, err := ParseBuiltin(name)
			if err != nil {
				t.Fatalf("ParseBuiltin(%q) error = %v", name, err)
			}
			if conv == nil {
				t.Fatal("ParseBuiltin returned nil convention")
			}
			if conv.Name != name {
				t.Errorf("Name = %q, want %q", conv.Name, name)
			}
			if conv.Pattern == nil {
				t.Error("Pattern is nil")
			}
			if conv.MaxLength <= 0 {
				t.Errorf("MaxLength = %d, want > 0", conv.MaxLength)
			}
		})
	}
}

func TestParseBuiltin_UnknownName(t *testing.T) {
	_, err := ParseBuiltin("nonexistent")
	if err == nil {
		t.Fatal("ParseBuiltin(nonexistent) expected error, got nil")
	}
}
