package config

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestJevConsumers_MergeKeyWarningSurvivesCacheUpgrade (leader audit F-A,
// repair cycle R2 of card t1979): the D2 merge-key warning is emitted only in
// JevConsumers.UnmarshalYAML — the parse path. configCacheSchemaVersion was
// still 14 when D2 added that warning, so a cache written by pre-D2 code
// (e.g. origin/main 4c4217a19d) stays a schema-version HIT after the upgrade,
// and the cached empty warnings silently shadow the one warning a direct load
// of the same workflow.yaml yields. The fixture below writes a real
// v14-shape cache file (the cacheFile JSON of cache.go: schema_version 14, a
// fingerprint of the section dir, a Config whose JevConsumers holds the
// merge-key consumer with EMPTY warnings — exactly what a pre-D2 binary
// wrote: the Warnings field existed, the merge-key branch did not) and
// asserts the warning IS present after a cache-path load. With the schema
// version left at 14 the stale cache is served and the warning is lost.
func TestJevConsumers_MergeKeyWarningSurvivesCacheUpgrade(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	// A workflow.yaml whose consumer c1 uses a YAML merge key: a direct load
	// yields exactly one section warning (TestJevConsumers_EntryMergeKeyWarns
	// pins the decode behavior; this premise check pins the fixture).
	mustWriteFile(t, filepath.Join(sectionsDir, "workflow.yaml"),
		"workflow:\n"+
			"  jev:\n"+
			"    consumers:\n"+
			"      base: &base {mode: advise, threshold: 0.25}\n"+
			"      c1:\n"+
			"        <<: *base\n"+
			"      ok: {mode: steer}\n")

	// The stale cache payload as a pre-D2 (4c4217a19d) binary would have
	// written it: the same Config JSON, schema_version 14, and a fingerprint
	// that still matches the section files on disk.
	direct, err := NewLoader().Load(dir)
	if err != nil {
		t.Fatalf("direct load: %v", err)
	}
	w := direct.Workflow.Jev.Consumers.SectionWarnings()
	if len(w) != 1 || !strings.Contains(w[0], "merge key") {
		t.Fatalf("direct load: section warnings = %v, want exactly one naming the merge key (fixture premise)", w)
	}
	stale := *direct
	stale.Workflow.Jev.Consumers.Warnings = nil // the pre-D2 decode: no merge-key branch

	fp, err := computeFingerprint(sectionsDir)
	if err != nil {
		t.Fatalf("fingerprint: %v", err)
	}
	cf := cacheFile{
		SchemaVersion: 14, // the pre-D2 schema, NOT configCacheSchemaVersion, which must move past it
		WrittenAt:     time.Now().UTC(),
		Fingerprint:   fp,
		Config:        &stale,
	}
	data, err := json.Marshal(&cf)
	if err != nil {
		t.Fatalf("marshal stale cache: %v", err)
	}
	// writeCache's own MkdirAll is not on this path — the stale cache is
	// hand-written, so the state directory must be created here.
	mustMkdir(t, filepath.Dir(cacheFilePath(dir)))
	mustWriteFile(t, cacheFilePath(dir), string(data))

	// Load through the cache path. On a schema HIT the cached empty warnings
	// are served and the merge-key warning is lost (F-A); on a MISS the
	// re-merge re-parses workflow.yaml and the warning is present.
	cfg, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("LoadWithCache over stale v14 cache: %v", err)
	}
	got := cfg.Workflow.Jev.Consumers.SectionWarnings()
	if len(got) != 1 || !strings.Contains(got[0], "merge key") {
		t.Errorf("merge-key warning lost over a v14 cache: section warnings = %v, want the one warning a direct load yields", got)
	}
}
