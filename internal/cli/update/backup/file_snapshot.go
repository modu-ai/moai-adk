package backup

// file_snapshot.go holds the machinery behind every per-file base snapshot the
// update merge reads. It was extracted from settings_snapshot.go when .mcp.json
// gained the same treatment (card t1029); the lifecycle, the manifest gate, and
// the staging/promotion split are unchanged from
// SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001, only parameterised.
//
// A file gets a snapshot namespace by declaring a fileSnapshot value; the
// exported Settings*/MCP* functions are thin wrappers over one each, so a
// caller never picks a namespace by string.

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// pendingSuffix marks the staging copy of a snapshot namespace.
const pendingSuffix = ".pending"

// fileSnapshot names one mergeable file's snapshot namespace under the shared
// cache root.
//
// subdir may be empty, which puts the copies directly under the snapshot root
// as their own sibling of sections/ and claude/. Where a subdir is used it
// drops the live path's leading dot, so the cache never holds a directory or
// file another tool might discover as project configuration.
type fileSnapshot struct {
	subdir string
	name   string
	// liveRel is the project-relative path of the file this records, in the
	// slash form the manifest keys use.
	liveRel string
	// The two prefixes are distinct per namespace so one file's failure can
	// never be read as another's.
	writeFailedPrefix   string
	promoteFailedPrefix string
}

// path returns the canonical base the next update merges against.
func (s fileSnapshot) path(projectRoot string) string {
	if s.subdir == "" {
		return filepath.Join(SnapshotDir(projectRoot), s.name)
	}
	return filepath.Join(SnapshotDir(projectRoot), s.subdir, s.name)
}

// pendingPath returns the staging copy of the render a flow deployed.
func (s fileSnapshot) pendingPath(projectRoot string) string {
	return s.path(projectRoot) + pendingSuffix
}

// load returns the canonical base when it exists, reads, and decodes as a JSON
// object. Anything else reports false and the merge keeps the derived base.
func (s fileSnapshot) load(projectRoot string) ([]byte, bool) {
	data, err := os.ReadFile(s.path(projectRoot))
	if err != nil {
		return nil, false
	}
	var doc map[string]any
	if json.Unmarshal(data, &doc) != nil || doc == nil {
		return nil, false
	}
	return data, true
}

// stageDeployed records the render the deploy just wrote as the staging copy.
//
// Whether the deploy wrote the file is decided from the deployer's own manifest
// record, never from the file alone: the entry must be template-managed and its
// template hash — the hash of the bytes the deployer wrote — must match the
// file. A deploy that skipped an existing user file records nothing, and the
// hash match also means the staged bytes are the deployer's render rather than
// a later rewrite.
func (s fileSnapshot) stageDeployed(projectRoot string, m manifest.Manager, warn io.Writer) {
	if m == nil {
		return
	}
	entry, ok := m.GetEntry(s.liveRel)
	if !ok || entry.Provenance != manifest.TemplateManaged {
		return
	}
	render, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(s.liveRel)))
	if err != nil || manifest.HashBytes(render) != entry.TemplateHash {
		return
	}
	pending := s.pendingPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(pending), defs.DirPerm); err != nil {
		warnLine(warn, s.writeFailedPrefix, err)
		return
	}
	if err := os.WriteFile(pending, render, defs.FilePerm); err != nil {
		warnLine(warn, s.writeFailedPrefix, err)
	}
}

// judgeLeftover settles a staging copy an earlier flow left behind because it
// stopped before its promotion decision. The live file still equal to the
// leftover means nothing reverted the render after the stop, so it is promoted;
// any difference discards it and keeps the prior base.
func (s fileSnapshot) judgeLeftover(projectRoot string, warn io.Writer) {
	leftover, err := os.ReadFile(s.pendingPath(projectRoot))
	if err != nil {
		return
	}
	live, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(s.liveRel)))
	if err == nil && bytes.Equal(live, leftover) {
		s.promote(projectRoot, warn)
		return
	}
	s.discard(projectRoot)
}

// settle ends a flow's snapshot lifecycle, after that flow's merge of this
// file. preserved reports whether the merge wrote the pre-flow user file back
// wholesale, in which case the live file does not reflect the render and the
// staging copy is discarded; otherwise it is promoted.
func (s fileSnapshot) settle(projectRoot string, preserved bool, warn io.Writer) {
	if _, err := os.Stat(s.pendingPath(projectRoot)); err != nil {
		return
	}
	if preserved {
		s.discard(projectRoot)
		return
	}
	s.promote(projectRoot, warn)
}

// promote moves the staging copy onto the canonical path with a same-directory
// rename, which replaces an existing file on every platform.
func (s fileSnapshot) promote(projectRoot string, warn io.Writer) {
	if err := os.Rename(s.pendingPath(projectRoot), s.path(projectRoot)); err != nil {
		warnLine(warn, s.promoteFailedPrefix, err)
	}
}

func (s fileSnapshot) discard(projectRoot string) {
	_ = os.Remove(s.pendingPath(projectRoot))
}
