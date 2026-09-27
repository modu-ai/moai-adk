package codexwiring

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func partByKey(parts []manifest.Part, key string) (manifest.Part, bool) {
	for _, p := range parts {
		if p.Key == key {
			return p, true
		}
	}
	return manifest.Part{}, false
}

// TestCodexWiringPartOrigins covers the REQ-DHR-001 origin rules: a part MoAI
// writes is created (with its inserted region), a part already present in a
// project without wiring evidence is preexisting, one present alongside
// earlier evidence but without a record is unknown, and no later pass
// promotes preexisting or unknown to created.
func TestCodexWiringPartOrigins(t *testing.T) {
	wire := func(t *testing.T, root string) {
		t.Helper()
		var out, warn bytes.Buffer
		if _, err := Wire(root, &out, &warn); err != nil {
			t.Fatalf("Wire: %v (%s)", err, warn.String())
		}
	}
	cfgParts := func(t *testing.T, root string) []manifest.Part {
		t.Helper()
		e, ok := manifestEntry(t, root, ConfigRelPath)
		if !ok {
			t.Fatal("no config.toml record")
		}
		return e.Parts
	}

	t.Run("fresh_project_created", func(t *testing.T) {
		root := t.TempDir()
		wire(t, root)
		parts := cfgParts(t, root)
		for _, key := range []string{PartKeyMCPTable, PartKeyTUITable} {
			p, ok := partByKey(parts, key)
			if !ok || p.Origin != manifest.OriginCreated || p.Region == "" {
				t.Fatalf("%s part=%+v ok=%v", key, p, ok)
			}
		}
		if w, ok := partByKey(parts, ""); !ok || w.Kind != manifest.PartWholeFile || w.Origin != manifest.OriginCreated {
			t.Fatalf("whole-file part=%+v ok=%v", w, ok)
		}
		cfg := readFile(t, filepath.Join(root, ConfigRelPath))
		mcp, _ := partByKey(parts, PartKeyMCPTable)
		tui, _ := partByKey(parts, PartKeyTUITable)
		if string(cfg) != mcp.Region+tui.Region {
			t.Fatalf("regions do not reassemble the created file:\nfile=%q\nregions=%q+%q", cfg, mcp.Region, tui.Region)
		}
	})

	t.Run("status_line_key_into_user_tui", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ConfigRelPath), "[tui]\ntheme = \"dark\"\n")
		wire(t, root)
		p, ok := partByKey(cfgParts(t, root), PartKeyStatusLine)
		if !ok || p.Kind != manifest.PartTOMLKey || p.Origin != manifest.OriginCreated || p.Region != statusLineDefaultTOML+"\n" {
			t.Fatalf("status_line part=%+v ok=%v", p, ok)
		}
	})

	t.Run("preexisting_without_evidence", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ConfigRelPath), "[mcp_servers.moai]\ncommand = \"mine\"\n")
		wire(t, root)
		p, _ := partByKey(cfgParts(t, root), PartKeyMCPTable)
		if p.Origin != manifest.OriginPreexisting || p.Hash != "" {
			t.Fatalf("pre-existing table recorded %+v", p)
		}
		wire(t, root)
		if p, _ := partByKey(cfgParts(t, root), PartKeyMCPTable); p.Origin != manifest.OriginPreexisting {
			t.Fatalf("second pass promoted preexisting to %q", p.Origin)
		}
	})

	t.Run("unknown_with_sidecar_evidence", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ConfigRelPath), "[mcp_servers.moai]\ncommand = \"moai\"\n")
		writeFile(t, filepath.Join(root, SidecarPath), "{}\n")
		wire(t, root)
		p, _ := partByKey(cfgParts(t, root), PartKeyMCPTable)
		if p.Origin != manifest.OriginUnknown {
			t.Fatalf("table under earlier evidence recorded %q, want unknown", p.Origin)
		}
		wire(t, root)
		if p, _ := partByKey(cfgParts(t, root), PartKeyMCPTable); p.Origin != manifest.OriginUnknown {
			t.Fatalf("re-wiring promoted unknown to %q", p.Origin)
		}
	})
}
