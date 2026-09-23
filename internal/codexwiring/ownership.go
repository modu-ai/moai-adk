package codexwiring

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// Part keys (design §A.1).
const (
	// PartKeyDescription is the top-level hooks.json description.
	PartKeyDescription = "description"
	// PartKeyMCPTable is the [mcp_servers.moai] table.
	PartKeyMCPTable = "mcp_servers.moai"
	// PartKeyTUITable is a [tui] table MoAI appended as a whole.
	PartKeyTUITable = "tui"
	// PartKeyStatusLine is a status_line assignment inside a user [tui].
	PartKeyStatusLine = "tui.status_line"
)

// derivedPart is a part observed in one pass: either written by MoAI in this
// pass (createdNow) or found already present.
type derivedPart struct {
	part       manifest.Part
	createdNow bool
}

// partFamily groups keys that describe the same surface, so a record kept
// for one form (a whole appended [tui]) is not joined by a second record for
// the other form (its status_line key) on a later pass.
func partFamily(key string) string {
	if key == PartKeyTUITable || key == PartKeyStatusLine {
		return PartKeyTUITable
	}
	return key
}

// wiringEvidence reports whether the project carries earlier MoAI wiring
// evidence: a manifest part record for a wiring file, the trust sidecar, or
// the wiring journal. It reads only; an unreadable manifest counts as
// evidence, so an unprovable part resolves to unknown rather than to
// preexisting.
func wiringEvidence(projectRoot string) bool {
	for _, rel := range []string{SidecarPath, JournalRelPath} {
		if _, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(rel))); err == nil {
			return true
		}
	}
	files, ok := readManifestFiles(projectRoot)
	if !ok {
		return true
	}
	for _, rel := range []string{HooksRelPath, ConfigRelPath} {
		if len(files[rel].Parts) > 0 {
			return true
		}
	}
	return false
}

// readManifestFiles parses the manifest without Load's corrupt-file rename,
// so read-only callers never move it. ok is false when it exists but cannot
// be parsed.
func readManifestFiles(projectRoot string) (map[string]manifest.FileEntry, bool) {
	raw, err := os.ReadFile(filepath.Join(projectRoot, defs.MoAIDir, defs.ManifestJSON))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]manifest.FileEntry{}, true
	}
	if err != nil {
		return nil, false
	}
	var mf manifest.Manifest
	if err := json.Unmarshal(raw, &mf); err != nil {
		return nil, false
	}
	if mf.Files == nil {
		mf.Files = map[string]manifest.FileEntry{}
	}
	return mf.Files, true
}

// hooksParts derives the parts of a hooks.json render from the file as it was
// before the pass.
func hooksParts(existing []byte, existed bool) []derivedPart {
	var parts []derivedPart
	if !existed {
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartWholeFile, Origin: manifest.OriginCreated}, createdNow: true})
	}
	var top map[string]json.RawMessage
	_ = json.Unmarshal(existing, &top)
	if _, has := top["description"]; has {
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartJSONKey, Key: PartKeyDescription}})
	} else {
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartJSONKey, Key: PartKeyDescription, Origin: manifest.OriginCreated, Hash: sha256Hex([]byte(mustQuote(moaiHooksDescription)))}, createdNow: true})
	}
	for _, row := range codexadapter.EventTable {
		if !row.Adapted {
			continue
		}
		h := handlerJSON{Type: "command", Command: "moai hook " + row.DispatcherArg + harnessCodexSuffix, Timeout: moaiHandlerTimeout(row.CodexEvent)}
		raw, err := json.Marshal(h)
		if err != nil {
			continue
		}
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartHookHandler, Key: string(row.CodexEvent) + "|" + h.Command, Origin: manifest.OriginCreated, Hash: sha256Hex(raw)}, createdNow: true})
	}
	return parts
}

// renderConfig returns the ensured config.toml and the parts of that render.
// Each region is the exact text MoAI inserts, including the blank separator
// line appendSection puts before an appended table; the newline appendSection
// adds to terminate an unterminated last line is outside every region.
func renderConfig(existing []byte, existed bool) ([]byte, []derivedPart) {
	var parts []derivedPart
	if !existed {
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartWholeFile, Origin: manifest.OriginCreated}, createdNow: true})
	}
	withMCP := EnsureMCPTable(existing)
	if tablePresent(string(existing), mcpMoaiTableRe) {
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartTOMLTable, Key: PartKeyMCPTable}})
	} else {
		parts = append(parts, regionPart(manifest.PartTOMLTable, PartKeyMCPTable, appendedRegion(existing, withMCP)))
	}
	next := EnsureStatusLine(withMCP)
	tuiPresent, keyPresent := tuiState(withMCP)
	switch {
	case !tuiPresent:
		parts = append(parts, regionPart(manifest.PartTOMLTable, PartKeyTUITable, appendedRegion(withMCP, next)))
	case !keyPresent:
		parts = append(parts, regionPart(manifest.PartTOMLKey, PartKeyStatusLine, statusLineDefaultTOML+"\n"))
	default:
		parts = append(parts, derivedPart{part: manifest.Part{Kind: manifest.PartTOMLKey, Key: PartKeyStatusLine}})
	}
	return next, parts
}

func regionPart(kind manifest.PartKind, key, region string) derivedPart {
	return derivedPart{part: manifest.Part{Kind: kind, Key: key, Origin: manifest.OriginCreated, Hash: sha256Hex([]byte(region)), Region: region}, createdNow: true}
}

// appendedRegion is the suffix appendSection added to before: next minus the
// before bytes and the one terminating newline appendSection may have added.
func appendedRegion(before, next []byte) string {
	base := string(before)
	if base != "" && !strings.HasSuffix(base, "\n") {
		base += "\n"
	}
	return strings.TrimPrefix(string(next), base)
}

// tuiState reports whether body has a [tui] table and whether that table
// assigns status_line.
func tuiState(body []byte) (tablePresent, keyPresent bool) {
	lines := splitLines(string(body))
	for i, line := range lines {
		if !tuiTableRe.MatchString(line) {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if anyTableRe.MatchString(lines[j]) {
				break
			}
			if statusLineKeyRe.MatchString(lines[j]) {
				return true, true
			}
		}
		return true, false
	}
	return false, false
}

// mergeParts combines the recorded parts of a file with the parts observed in
// this pass (design §A.1):
//
//   - hook handlers are replaced by the current table (the generator replaces
//     them every run, so they are always MoAI's);
//   - a recorded part is kept as recorded, so preexisting or unknown is never
//     promoted to created;
//   - a part MoAI wrote in this pass is created;
//   - a part found present without a record is preexisting when the project
//     carries no earlier wiring evidence, unknown otherwise;
//   - a created whole-file record follows MoAI's write only while the file
//     is still exactly what MoAI last wrote; once anything else changed it,
//     the claim is dropped and ownership falls back to parts.
func mergeParts(recorded []manifest.Part, derived []derivedPart, evidence bool, preHash, postHash string) []manifest.Part {
	var out []manifest.Part
	byFamily := map[string]manifest.Part{}
	var whole *manifest.Part
	for i := range recorded {
		p := recorded[i]
		switch p.Kind {
		case manifest.PartHookHandler:
			continue
		case manifest.PartWholeFile:
			whole = &p
		default:
			byFamily[partFamily(p.Key)] = p
		}
	}
	if whole != nil {
		if whole.Origin != manifest.OriginCreated {
			out = append(out, *whole)
		} else if whole.Hash == preHash {
			w := *whole
			w.Hash = postHash
			out = append(out, w)
		}
	}
	for _, d := range derived {
		p := d.part
		switch p.Kind {
		case manifest.PartWholeFile:
			if whole == nil && d.createdNow {
				p.Hash = postHash
				out = append(out, p)
			}
		case manifest.PartHookHandler:
			out = append(out, p)
		default:
			fam := partFamily(p.Key)
			if rec, ok := byFamily[fam]; ok {
				out = append(out, rec)
				delete(byFamily, fam)
				continue
			}
			if !d.createdNow {
				p.Origin = manifest.OriginPreexisting
				if evidence {
					p.Origin = manifest.OriginUnknown
				}
				p.Hash, p.Region = "", ""
			}
			out = append(out, p)
		}
	}
	return out
}

// intendedEntry is the manifest record a pass will apply for rel.
func intendedEntry(parts []manifest.Part, postHash string) manifest.FileEntry {
	return manifest.FileEntry{Provenance: manifest.GeneratedManaged, DeployedHash: postHash, CurrentHash: postHash, Parts: parts}
}
