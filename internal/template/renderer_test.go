package template

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRendererRender(t *testing.T) {
	t.Run("successful_render", func(t *testing.T) {
		fs := fstest.MapFS{
			"CLAUDE.md.tmpl": &fstest.MapFile{
				Data: []byte("# {{.ProjectName}}\n\nVersion: {{.Version}}\n"),
			},
		}
		r := NewRenderer(fs)

		data := map[string]string{
			"ProjectName": "MoAI-ADK",
			"Version":     "1.0.0",
		}

		result, err := r.Render("CLAUDE.md.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}

		expected := "# MoAI-ADK\n\nVersion: 1.0.0\n"
		if string(result) != expected {
			t.Errorf("Render result = %q, want %q", string(result), expected)
		}
	})

	t.Run("missing_key_strict_mode", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.tmpl": &fstest.MapFile{
				Data: []byte("Hello {{.Name}}, your role is {{.Role}}"),
			},
		}
		r := NewRenderer(fs)

		// Only provide Name, not Role
		data := map[string]string{
			"Name": "GOOS",
		}

		_, err := r.Render("test.tmpl", data)
		if err == nil {
			t.Fatal("expected error for missing key")
		}
		if !errors.Is(err, ErrMissingTemplateKey) {
			t.Errorf("expected ErrMissingTemplateKey, got: %v", err)
		}
	})

	t.Run("nonexistent_template", func(t *testing.T) {
		fs := fstest.MapFS{}
		r := NewRenderer(fs)

		_, err := r.Render("nonexistent.tmpl", nil)
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
		if !errors.Is(err, ErrTemplateNotFound) {
			t.Errorf("expected ErrTemplateNotFound, got: %v", err)
		}
	})

	t.Run("no_unexpanded_tokens_in_result", func(t *testing.T) {
		fs := fstest.MapFS{
			"config.tmpl": &fstest.MapFile{
				Data: []byte("name: {{.Name}}\nversion: {{.Version}}"),
			},
		}
		r := NewRenderer(fs)

		data := map[string]string{
			"Name":    "test-project",
			"Version": "2.0.0",
		}

		result, err := r.Render("config.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}

		content := string(result)
		if strings.Contains(content, "{{.") {
			t.Errorf("result contains unexpanded Go template token: %s", content)
		}
	})

	t.Run("complex_template_with_conditionals", func(t *testing.T) {
		fs := fstest.MapFS{
			"complex.tmpl": &fstest.MapFile{
				Data: []byte(`{{if .Enabled}}Feature ON{{else}}Feature OFF{{end}}`),
			},
		}
		r := NewRenderer(fs)

		data := map[string]bool{"Enabled": true}
		result, err := r.Render("complex.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if string(result) != "Feature ON" {
			t.Errorf("result = %q, want %q", string(result), "Feature ON")
		}
	})

	t.Run("empty_template", func(t *testing.T) {
		fs := fstest.MapFS{
			"empty.tmpl": &fstest.MapFile{
				Data: []byte(""),
			},
		}
		r := NewRenderer(fs)

		result, err := r.Render("empty.tmpl", nil)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d bytes", len(result))
		}
	})

	t.Run("template_with_range", func(t *testing.T) {
		fs := fstest.MapFS{
			"list.tmpl": &fstest.MapFile{
				Data: []byte("{{range .Items}}- {{.}}\n{{end}}"),
			},
		}
		r := NewRenderer(fs)

		data := map[string][]string{
			"Items": {"alpha", "beta", "gamma"},
		}

		result, err := r.Render("list.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}

		expected := "- alpha\n- beta\n- gamma\n"
		if string(result) != expected {
			t.Errorf("result = %q, want %q", string(result), expected)
		}
	})
}

func TestRendererPassthroughTokens(t *testing.T) {
	fs := fstest.MapFS{
		"hooks.tmpl": &fstest.MapFile{
			Data: []byte(`{"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\"", "name": "{{.Name}}"}`),
		},
	}
	r := NewRenderer(fs)

	data := map[string]string{"Name": "test"}
	result, err := r.Render("hooks.tmpl", data)
	if err != nil {
		t.Fatalf("expected passthrough of $CLAUDE_PROJECT_DIR, got error: %v", err)
	}

	content := string(result)
	if !strings.Contains(content, "$CLAUDE_PROJECT_DIR") {
		t.Error("$CLAUDE_PROJECT_DIR should be preserved in output")
	}
}

func TestJsonEscapeTemplateFunc(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unix_path_unchanged",
			input: "/usr/local/go/bin:/usr/bin",
			want:  "/usr/local/go/bin:/usr/bin",
		},
		{
			name:  "windows_path_backslashes_escaped",
			input: `C:\Users\user\go\bin;C:\Windows\system32`,
			want:  `C:\\Users\\user\\go\\bin;C:\\Windows\\system32`,
		},
		{
			name:  "double_quotes_escaped",
			input: `path with "quotes"`,
			want:  `path with \"quotes\"`,
		},
		{
			name:  "tab_and_newline_escaped",
			input: "line1\tvalue\nline2",
			want:  `line1\tvalue\nline2`,
		},
		{
			name:  "empty_string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := templateFuncMap["jsonEscape"].(func(string) string)
			got := fn(tt.input)
			if got != tt.want {
				t.Errorf("jsonEscape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPosixPathTemplateFunc(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "windows_path_converted",
			input: `C:\Users\user\go\bin`,
			want:  "C:/Users/user/go/bin",
		},
		{
			name:  "unix_path_unchanged",
			input: "/home/user/go/bin",
			want:  "/home/user/go/bin",
		},
		{
			name:  "mixed_separators",
			input: `C:\Users/user\projects`,
			want:  "C:/Users/user/projects",
		},
		{
			name:  "empty_string",
			input: "",
			want:  "",
		},
		{
			name:  "no_separators",
			input: "moai",
			want:  "moai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := templateFuncMap["posixPath"].(func(string) string)
			got := fn(tt.input)
			if got != tt.want {
				t.Errorf("posixPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJsonEscapeInTemplate(t *testing.T) {
	t.Run("windows_path_produces_valid_json", func(t *testing.T) {
		fs := fstest.MapFS{
			"settings.json.tmpl": &fstest.MapFile{
				Data: []byte(`{"env":{"PATH":"{{jsonEscape .SmartPATH}}"}}`),
			},
		}
		r := NewRenderer(fs)

		data := map[string]string{
			"SmartPATH": `C:\Users\user\go\bin;C:\Windows\system32`,
		}

		result, err := r.Render("settings.json.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}

		// Verify the output is valid JSON
		var parsed map[string]any
		if err := json.Unmarshal(result, &parsed); err != nil {
			t.Fatalf("rendered output is not valid JSON: %v\noutput: %s", err, string(result))
		}

		// Verify the PATH value round-trips correctly
		env, ok := parsed["env"].(map[string]any)
		if !ok {
			t.Fatal("expected env key in JSON output")
		}
		pathVal, ok := env["PATH"].(string)
		if !ok {
			t.Fatal("expected PATH key in env")
		}
		want := `C:\Users\user\go\bin;C:\Windows\system32`
		if pathVal != want {
			t.Errorf("PATH = %q, want %q", pathVal, want)
		}
	})

	t.Run("unix_path_produces_valid_json", func(t *testing.T) {
		fs := fstest.MapFS{
			"settings.json.tmpl": &fstest.MapFile{
				Data: []byte(`{"env":{"PATH":"{{jsonEscape .SmartPATH}}"}}`),
			},
		}
		r := NewRenderer(fs)

		data := map[string]string{
			"SmartPATH": "/home/user/go/bin:/usr/local/bin:/usr/bin",
		}

		result, err := r.Render("settings.json.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(result, &parsed); err != nil {
			t.Fatalf("rendered output is not valid JSON: %v\noutput: %s", err, string(result))
		}
	})
}

func TestPosixPathInTemplate(t *testing.T) {
	t.Run("windows_gobinpath_converted_in_shell", func(t *testing.T) {
		fs := fstest.MapFS{
			"hook.sh.tmpl": &fstest.MapFile{
				Data: []byte(`if [ -f "{{posixPath .GoBinPath}}/moai" ]; then` + "\n" +
					`	exec "{{posixPath .GoBinPath}}/moai" hook start` + "\n" +
					`fi`),
			},
		}
		r := NewRenderer(fs)

		data := map[string]string{
			"GoBinPath": `C:\Users\user\go\bin`,
		}

		result, err := r.Render("hook.sh.tmpl", data)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}

		content := string(result)
		if strings.Contains(content, `\`) {
			t.Errorf("output still contains backslashes: %s", content)
		}
		if !strings.Contains(content, "C:/Users/user/go/bin/moai") {
			t.Errorf("expected POSIX path in output, got: %s", content)
		}
	})
}

func TestUnexpandedTokenDetection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		match   bool
	}{
		{"dollar_brace", "${SHELL}", true},
		{"double_brace", "{{VAR}}", true},
		{"go_template_dot", "{{.Name}}", true},
		{"dollar_var", "$HOME", true},
		{"normal_text", "hello world", false},
		{"dollar_lowercase", "$foo", false},        // pattern requires uppercase
		{"empty_braces", "${}", false},             // no var name
		{"partial_dollar", "cost is $5.00", false}, // $ followed by digit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unexpandedTokenPattern.MatchString(tt.content)
			if got != tt.match {
				t.Errorf("pattern match for %q = %v, want %v", tt.content, got, tt.match)
			}
		})
	}
}
