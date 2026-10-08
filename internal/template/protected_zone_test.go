package template_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// zoneRuntimeAllowed is the set of runtime_paths entries the shipped manifest may
// carry: each is created at init, update or run time rather than shipped as a
// template file. A new entry must be argued into this list, which keeps
// runtime_paths from becoming a place to park dead entries.
var zoneRuntimeAllowed = map[string]bool{
	".claude/settings.json":                                true,
	".claude/settings.local.json":                          true,
	".moai/project/protected-zone.yaml":                    true,
	".moai/harness/learning-history/rate-limit-state.json": true,
	".moai/harness/learning-history/":                      true,
	".moai/harness/usage-log.jsonl":                        true,
	"**/CLAUDE.md":                                         true,
	"**/AGENTS.md":                                         true,
	"**/AGENTS.local.md":                                   true,
	// The USER install roots: created at init/update time under the user's
	// home, never shipped as template files. The guard scopes their
	// protection to manifest-tracked files at match time.
	"user-root:claude-skills/": true,
	"user-root:claude-agents/": true,
	"user-root:agents-skills/": true,
	"user-root:codex-agents/":  true,
}

// zoneShippedFloor is the number of paths and runtime_paths entries the shipped
// manifest must carry, so a sweep over a shrunken file cannot read as a pass.
const (
	zoneShippedPathsFloor   = 17
	zoneShippedRuntimeFloor = 8
)

func zoneRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile = <root>/internal/template/protected_zone_test.go
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func zoneShippedBytes(t *testing.T) []byte {
	t.Helper()
	tree, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	data, err := fs.ReadFile(tree, config.ProtectedZoneShippedRel)
	if err != nil {
		t.Fatalf("the shipped manifest is not in the embedded templates: %v", err)
	}
	return data
}

func TestProtectedZone(t *testing.T) {
	t.Run("Shipped", func(t *testing.T) {
		data := zoneShippedBytes(t)
		entries, err := config.ParseProtectedZone(data, true, config.ProtectedZoneShippedRel)
		if err != nil {
			t.Fatalf("shipped manifest does not parse: %v", err)
		}
		zone := config.ProtectedZone{Entries: entries}

		// regression_tests ships empty: this repository's test layout is not a user project's
		for _, e := range entries {
			if e.Category == "regression_tests" {
				t.Errorf("regression_tests must ship empty, found %q", e.Raw)
			}
		}

		// both manifest files list themselves
		for _, self := range []string{config.ProtectedZoneShippedRel, config.ProtectedZoneOverlayRel} {
			found := false
			for _, e := range entries {
				if e.Match(config.FoldZoneText(self)) {
					found = true
				}
			}
			if !found {
				t.Errorf("the shipped manifest does not cover %s", self)
			}
		}
		// the shipped file lists its own path explicitly, not only through a directory entry
		explicit := false
		for _, e := range entries {
			if e.Raw == config.ProtectedZoneShippedRel {
				explicit = true
			}
		}
		if !explicit {
			t.Errorf("%s is not an explicit entry of itself", config.ProtectedZoneShippedRel)
		}

		// every paths entry exists in the template tree; every runtime entry is a known runtime one
		tree, err := template.EmbeddedTemplates()
		if err != nil {
			t.Fatal(err)
		}
		sweep, err := config.SweepZoneEntries(zone, config.ProtectedZoneShippedRel, tree)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range sweep.Dead {
			t.Errorf("dead paths entry %q (%s): matches nothing in the template tree", d.Raw, d.Category)
		}
		for _, e := range entries {
			if e.Runtime && !zoneRuntimeAllowed[e.Raw] {
				t.Errorf("runtime_paths entry %q is not one the manifest may create at runtime", e.Raw)
			}
		}
		if sweep.Resolved < zoneShippedPathsFloor || sweep.Skipped < zoneShippedRuntimeFloor {
			t.Fatalf("swept resolved=%d skipped=%d, want at least %d/%d", sweep.Resolved, sweep.Skipped, zoneShippedPathsFloor, zoneShippedRuntimeFloor)
		}
		t.Logf("swept paths resolved=%d runtime skipped=%d", sweep.Resolved, sweep.Skipped)
	})

	t.Run("Neutrality", func(t *testing.T) {
		data := zoneShippedBytes(t)
		forbidden := map[string]*regexp.Regexp{
			"SPEC identifier":        regexp.MustCompile(`\bSPEC-[A-Z0-9-]+\b`),
			"requirement token":      regexp.MustCompile(`\b(REQ|AC)-[A-Z0-9-]+\b`),
			"audit citation":         regexp.MustCompile(`Audit [0-9]+|Finding [A-Z][0-9]`),
			"internal date":          regexp.MustCompile(`\b20[0-9]{2}-[01][0-9]-[0-3][0-9]\b`),
			"commit hash":            regexp.MustCompile(`\b[0-9a-f]{7,40}\b`),
			"card id":                regexp.MustCompile(`\bt[0-9]{3,4}\b`),
			"legacy local file name": regexp.MustCompile(`CLAUDE\.local\.md`),
			"absolute user path":     regexp.MustCompile(`/Users/|/home/`),
		}
		swept := 0
		for name, re := range forbidden {
			swept++
			if loc := re.FindIndex(data); loc != nil {
				t.Errorf("shipped manifest carries a %s: %q", name, data[loc[0]:loc[1]])
			}
		}
		// the dogfood overlay is where the legacy local-instruction entry lives
		overlay, err := os.ReadFile(filepath.Join(zoneRepoRoot(t), filepath.FromSlash(config.ProtectedZoneOverlayRel)))
		if err != nil {
			t.Fatalf("dogfood overlay: %v", err)
		}
		swept++
		if !bytes.Contains(overlay, []byte("**/CLAUDE.local.md")) {
			t.Errorf("the dogfood overlay must carry **/CLAUDE.local.md, which the shipped manifest cannot")
		}
		if swept != len(forbidden)+1 {
			t.Fatalf("swept %d checks", swept)
		}
		t.Logf("swept=%d", swept)
	})

	t.Run("SupersetOfShipped", func(t *testing.T) {
		root := zoneRepoRoot(t)
		// the local dogfood manifest is the shipped file byte for byte
		local, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(config.ProtectedZoneShippedRel)))
		if err != nil {
			t.Fatalf("local dogfood manifest: %v", err)
		}
		if shipped := zoneShippedBytes(t); !bytes.Equal(local, shipped) {
			t.Errorf("local %s differs from the shipped manifest", config.ProtectedZoneShippedRel)
		}

		// the dogfood effective zone (shipped + overlay) contains every shipped entry
		load := config.LoadProtectedZone(root)
		if load.State != config.ZoneStateOK {
			t.Fatalf("dogfood zone state=%q file=%q err=%v", load.State, load.InvalidFile, load.Err)
		}
		shippedEntries, err := config.ParseProtectedZone(zoneShippedBytes(t), true, config.ProtectedZoneShippedRel)
		if err != nil {
			t.Fatal(err)
		}
		have := map[string]bool{}
		for _, e := range load.Zone.Entries {
			have[e.Category+"\x00"+e.Raw] = true
		}
		swept := 0
		for _, e := range shippedEntries {
			swept++
			if !have[e.Category+"\x00"+e.Raw] {
				t.Errorf("the dogfood zone lacks shipped entry %s %q", e.Category, e.Raw)
			}
		}
		if swept == 0 {
			t.Fatal("no shipped entries swept")
		}
		t.Logf("swept=%d dogfood entries=%d", swept, len(load.Zone.Entries))
	})
}
