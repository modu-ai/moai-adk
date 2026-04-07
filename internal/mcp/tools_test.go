package mcp

import "testing"

func TestDefineLSPTools_Count(t *testing.T) {
	t.Parallel()

	tools := defineLSPTools()
	const wantCount = 6
	if len(tools) != wantCount {
		t.Errorf("tool count: got %d want %d", len(tools), wantCount)
	}
}

func TestDefineLSPTools_Names(t *testing.T) {
	t.Parallel()

	tools := defineLSPTools()
	wantNames := []string{
		"goto_definition",
		"find_references",
		"hover",
		"document_symbols",
		"diagnostics",
		"rename",
	}

	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}

	for _, want := range wantNames {
		if !names[want] {
			t.Errorf("missing tool %q in defineLSPTools()", want)
		}
	}
}

func TestDefineLSPTools_Schemas(t *testing.T) {
	t.Parallel()

	tools := defineLSPTools()
	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			t.Parallel()

			if tool.Description == "" {
				t.Errorf("tool %q has empty description", tool.Name)
			}
			if tool.InputSchema.Type != "object" {
				t.Errorf("tool %q: InputSchema.Type = %q want %q", tool.Name, tool.InputSchema.Type, "object")
			}
			if len(tool.InputSchema.Properties) == 0 {
				t.Errorf("tool %q: InputSchema.Properties is empty", tool.Name)
			}
			if len(tool.InputSchema.Required) == 0 {
				t.Errorf("tool %q: InputSchema.Required is empty", tool.Name)
			}
		})
	}
}

func TestDefineLSPTools_FilePositionTools(t *testing.T) {
	t.Parallel()

	// These tools require file, line, and column.
	positionTools := map[string]bool{
		"goto_definition": true,
		"find_references": true,
		"hover":           true,
	}

	for _, tool := range defineLSPTools() {
		if !positionTools[tool.Name] {
			continue
		}
		t.Run(tool.Name, func(t *testing.T) {
			t.Parallel()

			required := make(map[string]bool)
			for _, r := range tool.InputSchema.Required {
				required[r] = true
			}

			for _, field := range []string{"file", "line", "column"} {
				if !required[field] {
					t.Errorf("tool %q: %q should be required", tool.Name, field)
				}
				if _, ok := tool.InputSchema.Properties[field]; !ok {
					t.Errorf("tool %q: missing property %q", tool.Name, field)
				}
			}
		})
	}
}

func TestDefineLSPTools_RenameHasNewName(t *testing.T) {
	t.Parallel()

	for _, tool := range defineLSPTools() {
		if tool.Name != "rename" {
			continue
		}
		if _, ok := tool.InputSchema.Properties["new_name"]; !ok {
			t.Error("rename tool: missing 'new_name' property")
		}
		hasNewName := false
		for _, r := range tool.InputSchema.Required {
			if r == "new_name" {
				hasNewName = true
				break
			}
		}
		if !hasNewName {
			t.Error("rename tool: 'new_name' should be required")
		}
	}
}

func TestFilePositionSchema(t *testing.T) {
	t.Parallel()

	schema := filePositionSchema()

	if schema.Type != "object" {
		t.Errorf("Type: got %q want %q", schema.Type, "object")
	}

	for _, field := range []string{"file", "line", "column"} {
		if _, ok := schema.Properties[field]; !ok {
			t.Errorf("missing property %q", field)
		}
	}

	required := make(map[string]bool)
	for _, r := range schema.Required {
		required[r] = true
	}
	for _, field := range []string{"file", "line", "column"} {
		if !required[field] {
			t.Errorf("property %q should be required", field)
		}
	}
}
