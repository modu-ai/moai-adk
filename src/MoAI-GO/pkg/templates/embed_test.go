package template

import (
	"testing"
)

func TestListTemplates(t *testing.T) {
	files, err := ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}

	if len(files) == 0 {
		t.Error("ListTemplates() returned no files")
	}

	// Check for known files
	foundCLAUDE := false
	foundMoai := false
	for _, f := range files {
		if len(f) > 8 && f[:8] == ".claude/" {
			foundCLAUDE = true
		}
		if len(f) > 6 && f[:6] == ".moai/" {
			foundMoai = true
		}
	}

	if !foundCLAUDE {
		t.Error("No .claude templates found")
	}
	if !foundMoai {
		t.Error("No .moai templates found")
	}
}

func TestExtractTemplate(t *testing.T) {
	// Test extracting a known file
	content, err := ExtractTemplate(".moai/config/config.yaml")
	if err != nil {
		t.Fatalf("ExtractTemplate() failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("ExtractTemplate() returned empty content")
	}
}
