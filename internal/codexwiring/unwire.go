package codexwiring

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// DisableCommand is the command that removes MoAI's Codex wiring. The update
// path names it when it finds wiring the configured harness no longer uses.
const DisableCommand = "moai tool disable codex"

// RemovedPart is one part (or, with Part empty, one whole file) unwire took out.
type RemovedPart struct {
	Path string
	Part string
}

// UnwireResult reports what `moai tool disable codex` removed and what it left.
type UnwireResult struct {
	Removed   []RemovedPart
	Kept      []Refusal
	Conflicts []Conflict
	Recovered []RecoveryOutcome
}

// partStatus is what unwire found for one created part.
type partStatus int

const (
	partAbsent partStatus = iota
	partRemoved
	partModified
)

// Unwire removes the parts of the Codex wiring MoAI created and can prove are
// unchanged (REQ-DHR-005), under the wiring lock and after recovery
// (REQ-DHR-004). A part with no record, recorded preexisting or unknown, a
// modified part, and a symbolic-link target are left untouched and reported
// (REQ-DHR-006). A whole file is deleted only when MoAI created it and it is
// still exactly what MoAI last wrote. The sidecar then records the disable so
// the update-path refresh does not wire the project again.
func Unwire(projectRoot string, warn io.Writer) (UnwireResult, error) {
	return unwireWith(projectRoot, warn, defaultPassOptions())
}

func unwireWith(root string, warn io.Writer, opts passOptions) (UnwireResult, error) {
	evidence := wiringEvidence(root)
	release, err := acquireWiringLock(root)
	if err != nil {
		return UnwireResult{}, err
	}
	defer release()
	recovered, err := recoverLocked(root, warn)
	res := UnwireResult{Recovered: recovered}
	if err != nil {
		return res, fmt.Errorf("recover interrupted wiring change: %w", err)
	}
	recorded, _ := readManifestFiles(root)
	p := newPass(root, nil, warn, opts, evidence)
	u := &unwirer{p: p, res: &res, evidence: evidence}
	for _, rel := range []string{HooksRelPath, ConfigRelPath} {
		if err := u.file(rel, recorded[rel]); err != nil {
			return res, err
		}
	}
	res.Kept = append(res.Kept, p.res.Refusals...)
	res.Conflicts = p.res.Conflicts
	// Only a project MoAI has wired carries a disable: one it never wired has
	// nothing for the marker to hold back.
	if evidence || len(res.Removed) > 0 {
		if err := markDisabled(root); err != nil {
			warnf(warn, "record the disable in %s: %v", SidecarPath, err)
		}
	}
	if len(res.Conflicts) > 0 {
		return res, ErrWiringConflict
	}
	return res, nil
}

type unwirer struct {
	p        *pass
	res      *UnwireResult
	evidence bool
}

func (u *unwirer) keep(rel, part string, reason RefusalReason, detail string) {
	u.res.Kept = append(u.res.Kept, Refusal{Path: rel, Part: part, Reason: reason, Detail: detail})
}

// file unwires one wiring file.
func (u *unwirer) file(rel string, rec manifest.FileEntry) error {
	root := u.p.root
	target := filepath.Join(root, filepath.FromSlash(rel))
	if u.p.opts.guards.lstat {
		if why := boundaryViolation(root, rel); why != "" {
			u.keep(rel, "", ReasonSymlinkBoundary, why)
			return nil
		}
	}
	body, err := os.ReadFile(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", rel, err)
	}
	if len(rec.Parts) == 0 {
		if u.evidence {
			u.keep(rel, "", ReasonUnknownOrigin, "MoAI wired this project before part records existed; remove its entries by hand if they are not yours")
		} else {
			u.keep(rel, "", ReasonNoProvenance, "no MoAI ownership record")
		}
		return nil
	}
	cur := sha256Hex(body)
	var created, keepParts []manifest.Part
	for _, part := range rec.Parts {
		switch {
		case part.Kind == manifest.PartWholeFile:
			if part.Origin == manifest.OriginCreated && part.Hash == cur {
				return u.removeWhole(rel, target)
			}
		case part.Origin == manifest.OriginPreexisting:
			u.keep(rel, part.Key, ReasonUserOwned, "present before MoAI wiring")
			keepParts = append(keepParts, part)
		case part.Origin == manifest.OriginCreated:
			created = append(created, part)
		default:
			u.keep(rel, part.Key, ReasonUnknownOrigin, "MoAI cannot prove it wrote this part; remove it by hand if it is not yours")
			keepParts = append(keepParts, part)
		}
	}

	var next []byte
	var status map[int]partStatus
	if rel == HooksRelPath {
		next, status, err = unwireHooks(body, created)
		if err != nil {
			u.keep(rel, "", ReasonUnparseable, err.Error())
			return nil
		}
	} else {
		next, status = unwireConfig(body, created)
	}
	var gone []RemovedPart
	for i, part := range created {
		switch status[i] {
		case partRemoved:
			gone = append(gone, RemovedPart{Path: rel, Part: part.Key})
		case partModified:
			u.keep(rel, part.Key, ReasonModified, "changed since MoAI wrote it")
			keepParts = append(keepParts, part)
		}
	}
	if bytes.Equal(next, body) {
		return nil
	}
	var entry *manifest.FileEntry
	if len(keepParts) > 0 {
		post := sha256Hex(next)
		entry = &manifest.FileEntry{Provenance: manifest.GeneratedManaged, DeployedHash: post, CurrentHash: post, Parts: keepParts}
	}
	written, err := u.p.write(rel, next, entry)
	if err != nil {
		return err
	}
	if written {
		u.res.Removed = append(u.res.Removed, gone...)
	}
	return nil
}

func (u *unwirer) removeWhole(rel, target string) error {
	if err := u.p.remove(rel); err != nil {
		return err
	}
	if _, err := os.Lstat(target); errors.Is(err, os.ErrNotExist) {
		u.res.Removed = append(u.res.Removed, RemovedPart{Path: rel})
	}
	return nil
}

// unwireHooks removes the created hook handlers and the created description
// whose current JSON value hashes to the record. Every other top-level value,
// entry, matcher, and handler is carried through as parsed JSON, in array
// order. The document is re-serialized the way the generator writes it.
func unwireHooks(body []byte, created []manifest.Part) ([]byte, map[int]partStatus, error) {
	status := map[int]partStatus{}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return nil, nil, fmt.Errorf("parse hooks.json: %w", err)
	}
	handlers := map[string]int{} // event|command -> index into created
	for i, part := range created {
		switch part.Kind {
		case manifest.PartHookHandler:
			handlers[part.Key] = i
		case manifest.PartJSONKey:
			raw, ok := top[part.Key]
			if !ok {
				continue
			}
			if compactHash(raw) == part.Hash {
				delete(top, part.Key)
				status[i] = partRemoved
			} else {
				status[i] = partModified
			}
		}
	}
	doc := map[string]any{}
	for k, v := range top {
		if k != "hooks" {
			doc[k] = v
		}
	}
	if rawHooks, ok := top["hooks"]; ok {
		var byEvent map[string][]json.RawMessage
		if err := json.Unmarshal(rawHooks, &byEvent); err != nil {
			return nil, nil, fmt.Errorf("parse hooks block: %w", err)
		}
		events := map[string][]json.RawMessage{}
		for event, entries := range byEvent {
			var kept []json.RawMessage
			for _, raw := range entries {
				entry, drop, err := unwireEntry(raw, event, handlers, created, status)
				if err != nil {
					return nil, nil, err
				}
				if !drop {
					kept = append(kept, entry)
				}
			}
			if len(kept) > 0 || len(entries) == 0 {
				events[event] = kept
			}
		}
		doc["hooks"] = events
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("render hooks.json: %w", err)
	}
	return append(out, '\n'), status, nil
}

// unwireEntry drops the recorded, unchanged MoAI handlers from one entry.
// drop reports that the entry held only such handlers.
func unwireEntry(raw json.RawMessage, event string, handlers map[string]int, created []manifest.Part, status map[int]partStatus) (json.RawMessage, bool, error) {
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, false, fmt.Errorf("parse %s entry: %w", event, err)
	}
	var list []json.RawMessage
	if err := json.Unmarshal(entry["hooks"], &list); err != nil || len(list) == 0 {
		return raw, false, nil
	}
	var kept []json.RawMessage
	for _, h := range list {
		var cmd struct {
			Command string `json:"command"`
		}
		_ = json.Unmarshal(h, &cmd)
		i, ok := handlers[event+"|"+cmd.Command]
		if !ok {
			kept = append(kept, h)
			continue
		}
		if handlerHash(h) == created[i].Hash {
			status[i] = partRemoved
			continue
		}
		if status[i] != partRemoved {
			status[i] = partModified
		}
		kept = append(kept, h)
	}
	if len(kept) == len(list) {
		return raw, false, nil
	}
	if len(kept) == 0 {
		return nil, true, nil
	}
	rebuilt, err := json.Marshal(kept)
	if err != nil {
		return nil, false, err
	}
	entry["hooks"] = rebuilt
	out, err := json.Marshal(entry)
	return out, false, err
}

// handlerHash hashes a handler the way the generator records it. A handler
// carrying exactly the generator's three keys is re-marshalled in the
// generator's field order, so a tool that re-serialized the file with its
// keys in another order does not make the handler look modified; object key
// order is not part of the comparison.
func handlerHash(raw json.RawMessage) string {
	var keys map[string]json.RawMessage
	var h handlerJSON
	if json.Unmarshal(raw, &keys) == nil && len(keys) == 3 &&
		keys["type"] != nil && keys["command"] != nil && keys["timeout"] != nil &&
		json.Unmarshal(raw, &h) == nil {
		if b, err := json.Marshal(h); err == nil {
			return sha256Hex(b)
		}
	}
	return compactHash(raw)
}

// compactHash hashes the compact form of a JSON value, the form the
// generator records.
func compactHash(raw json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return ""
	}
	return sha256Hex(buf.Bytes())
}

// unwireConfig cuts each created part's recorded region out of body. Every
// byte outside the removed regions is kept.
func unwireConfig(body []byte, created []manifest.Part) ([]byte, map[int]partStatus) {
	status := map[int]partStatus{}
	text := string(body)
	for i, part := range created {
		if part.Region == "" || sha256Hex([]byte(part.Region)) != part.Hash {
			if configPartPresent(text, part) {
				status[i] = partModified
			}
			continue
		}
		var cut bool
		if part.Kind == manifest.PartTOMLKey {
			text, cut = cutStatusLine(text, part.Region)
		} else if strings.Count(text, part.Region) == 1 {
			text, cut = strings.Replace(text, part.Region, "", 1), true
		}
		switch {
		case cut:
			status[i] = partRemoved
		case configPartPresent(text, part):
			status[i] = partModified
		}
	}
	return []byte(text), status
}

// configPartPresent reports whether the surface a part names exists at all,
// which separates a modified part from one the user already removed.
func configPartPresent(text string, part manifest.Part) bool {
	switch part.Key {
	case PartKeyMCPTable:
		return tablePresent(text, mcpMoaiTableRe)
	case PartKeyTUITable:
		return tablePresent(text, tuiTableRe)
	case PartKeyStatusLine:
		_, key := tuiState([]byte(text))
		return key
	}
	return false
}

// cutStatusLine removes region when it is the line directly after a [tui]
// header, where the generator inserts it.
func cutStatusLine(text, region string) (string, bool) {
	lines := strings.SplitAfter(text, "\n")
	for i := 0; i+1 < len(lines); i++ {
		if tuiTableRe.MatchString(strings.TrimSuffix(lines[i], "\n")) && lines[i+1] == region {
			return strings.Join(append(lines[:i+1:i+1], lines[i+2:]...), ""), true
		}
	}
	return text, false
}

// markDisabled records the disable in the sidecar, with the hashes of what
// the wiring files are now.
func markDisabled(root string) error {
	var hooks, config []byte
	if b, err := os.ReadFile(filepath.Join(root, HooksRelPath)); err == nil {
		hooks = b
	}
	if b, err := os.ReadFile(filepath.Join(root, ConfigRelPath)); err == nil {
		config = b
	}
	doc := sidecarDoc{HooksSHA256: sha256Hex(hooks), ConfigSHA256: sha256Hex(config), Disabled: true}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(root, SidecarPath), append(raw, '\n'))
}

// OwnedWiringFiles lists the wiring files `moai tool disable codex` has work
// for: files whose record holds a part MoAI created, and, in a project an
// older binary wired (a sidecar but no part records), every wiring file
// present. A disabled project lists nothing. It reads only.
func OwnedWiringFiles(projectRoot string) []string {
	if wiringDisabled(projectRoot) {
		return nil
	}
	recorded, _ := readManifestFiles(projectRoot)
	_, sidecar, _ := LoadSidecar(projectRoot)
	anyRecord := len(recorded[HooksRelPath].Parts) > 0 || len(recorded[ConfigRelPath].Parts) > 0
	var out []string
	for _, rel := range []string{HooksRelPath, ConfigRelPath} {
		if _, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(rel))); err != nil {
			continue
		}
		owned := sidecar && !anyRecord
		for _, part := range recorded[rel].Parts {
			if part.Origin == manifest.OriginCreated {
				owned = true
			}
		}
		if owned {
			out = append(out, rel)
		}
	}
	return out
}
