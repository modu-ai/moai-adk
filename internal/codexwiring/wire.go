package codexwiring

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// sidecarDoc is the JSON shape of the trust sidecar (.moai/state/codex-wiring.json).
type sidecarDoc struct {
	HooksSHA256  string `json:"hooks_sha256"`
	ConfigSHA256 string `json:"config_sha256"`
}

// ErrValidationRefused marks a REQ-CW-003 refusal: rendered bytes failed the
// whitelist gate and NOTHING was written. Callers surface it as a loud
// failure, not a silent skip.
var ErrValidationRefused = errors.New("codex wiring refused: rendered hooks.json fails the whitelist")

// Wire generates or refreshes the Codex wiring of the project at projectRoot:
//
//   - .codex/hooks.json — merged render, whitelist-gated before any write
//   - .codex/config.toml — [mcp_servers.moai] + [tui].status_line, both
//     create-if-absent
//   - .moai/state/codex-wiring.json — trust sidecar (sha256 of the generated
//     content), written only in a run that wrote hooks.json
//
// Guidance (REQ-CW-008): first creation prints FirstTrustGuidance; a
// content-changing regeneration prints ReTrustGuidance; an unchanged
// regeneration prints nothing.
//
// Failure posture (spec §F): wiring is best-effort — IO problems warn and the
// pass continues; the ONLY hard failure is a validation refusal (REQ-CW-003),
// returned as ErrValidationRefused with the violating keys named.
func Wire(projectRoot string, out, warn io.Writer) (Result, error) {
	return wireWith(projectRoot, out, warn, defaultPassOptions())
}

// RefreshWiring is the update-path entry (REQ-CW-009): it refreshes wiring
// ONLY in projects that already carry a wiring file. File existence is the
// user's standing opt-in — a `--agent claude` (or flag-absent) init left no
// wiring behind, and an update must not create any. An interrupted wiring
// change is recovered first even when no wiring file survived it
// (REQ-DHR-004).
func RefreshWiring(projectRoot string, out, warn io.Writer) (Result, error) {
	return runPass(projectRoot, out, warn, defaultPassOptions(), true)
}

// wireWith is Wire with explicit pass options (tests switch guards and set
// the interruption seam).
func wireWith(projectRoot string, out, warn io.Writer, opts passOptions) (Result, error) {
	return runPass(projectRoot, out, warn, opts, false)
}

// runPass takes the wiring lock, recovers interrupted changes, and then
// wires. gated limits the wiring (not the recovery) to projects that already
// carry a wiring file. A multi-file pass is not atomic as a whole: each file
// change is journaled and recoverable on its own (REQ-DHR-002).
func runPass(projectRoot string, out, warn io.Writer, opts passOptions, gated bool) (Result, error) {
	if gated && !wiringFilesExist(projectRoot) && !journalExists(projectRoot) {
		return Result{}, nil
	}
	evidence := wiringEvidence(projectRoot)
	release, err := acquireWiringLock(projectRoot)
	if err != nil {
		return Result{}, err
	}
	defer release()
	recovered, err := recoverLocked(projectRoot, warn)
	if err != nil {
		return Result{Recovered: recovered}, fmt.Errorf("recover interrupted wiring change: %w", err)
	}
	if gated && !wiringFilesExist(projectRoot) {
		return Result{Recovered: recovered}, nil
	}
	p := newPass(projectRoot, out, warn, opts, evidence)
	p.res.Recovered = recovered
	err = p.wireProject()
	return p.res, err
}

// journalExists reports whether a wiring journal is present.
func journalExists(projectRoot string) bool {
	_, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(JournalRelPath)))
	return err == nil
}

// wireProject is the shared body of Wire and RefreshWiring, run under the
// wiring lock after recovery.
func (p *pass) wireProject() error {
	projectRoot, out, warn := p.root, p.out, p.warn
	res := &p.res
	recorded, _ := readManifestFiles(projectRoot)

	hooksPath := filepath.Join(projectRoot, HooksRelPath)
	existing, readErr := os.ReadFile(hooksPath)
	res.HooksExisted = readErr == nil

	var rendered []byte
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		warnf(warn, "read %s: %v — leaving it untouched", HooksRelPath, readErr)
		res.HooksSkipped = true
	}
	if !res.HooksSkipped {
		merged, mergeErr := RenderHooks(existing)
		if mergeErr != nil {
			// Unparseable user document: never modify it, warn, continue with
			// the rest of the wiring (§B edge case).
			warnf(warn, "%s is unparseable (%v); leaving the file untouched and skipping hook wiring this run", HooksRelPath, mergeErr)
			res.HooksSkipped = true
		} else {
			rendered = merged
		}
	}

	if rendered != nil {
		// REQ-CW-003 gate: validate BEFORE the write. Violating bytes never
		// reach disk; the pass aborts with the violations named.
		violations, verr := codexadapter.ValidateConfig(rendered)
		if verr != nil {
			return fmt.Errorf("%w: %v", ErrValidationRefused, verr)
		}
		if len(violations) > 0 {
			names := make([]string, len(violations))
			for i, v := range violations {
				names[i] = v.Error()
			}
			return fmt.Errorf("%w: %s", ErrValidationRefused, strings.Join(names, "; "))
		}

		pre := ""
		if res.HooksExisted {
			pre = sha256Hex(existing)
		}
		post := sha256Hex(rendered)
		entry := intendedEntry(mergeParts(recorded[HooksRelPath].Parts, hooksParts(existing, res.HooksExisted), p.evidence, pre, post), post)
		if res.HooksExisted && bytesEqual(rendered, existing) {
			res.HooksChanged = false // unchanged regeneration — no write, no guidance
			p.recordUnchanged(HooksRelPath, recorded, entry)
		} else {
			written, err := p.write(HooksRelPath, rendered, entry)
			if err != nil {
				if p.crashed {
					return err
				}
				warnf(warn, "write %s: %v", HooksRelPath, err)
			}
			if written {
				res.HooksWritten = true
				res.HooksChanged = true
			}
		}
	}

	// config.toml: create-if-absent for both surfaces (idempotent by design).
	cfgPath := filepath.Join(projectRoot, ConfigRelPath)
	cfgExisting, cfgErr := os.ReadFile(cfgPath)
	if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
		warnf(warn, "read %s: %v", ConfigRelPath, cfgErr)
	} else {
		cfgExisted := cfgErr == nil
		cfgNext, cfgParts := renderConfig(cfgExisting, cfgExisted)
		pre := ""
		if cfgExisted {
			pre = sha256Hex(cfgExisting)
		}
		post := sha256Hex(cfgNext)
		entry := intendedEntry(mergeParts(recorded[ConfigRelPath].Parts, cfgParts, p.evidence, pre, post), post)
		if cfgExisted && bytesEqual(cfgNext, cfgExisting) {
			p.recordUnchanged(ConfigRelPath, recorded, entry)
		} else {
			written, err := p.write(ConfigRelPath, cfgNext, entry)
			if err != nil {
				if p.crashed {
					return err
				}
				warnf(warn, "write %s: %v", ConfigRelPath, err)
			}
			res.ConfigWritten = written
		}
	}

	// Trust sidecar: record the generated content hash so doctor can detect
	// divergence and regenerations can detect change (REQ-CW-008). Written in
	// a run that generated hooks.json, OR when the baseline was lost (e.g. a
	// force-reinit resets .moai/state while an unchanged hooks.json survives)
	// and the on-disk hooks still byte-match the render — the current file IS
	// the last generated content, verified, so re-recording restores the
	// divergence signal instead of silently degrading to "no baseline".
	sidecarNeeded := res.HooksWritten
	if !sidecarNeeded && !res.HooksSkipped && len(rendered) > 0 {
		if _, present, serr := LoadSidecar(projectRoot); serr == nil && !present {
			if onDisk, rerr := os.ReadFile(hooksPath); rerr == nil && bytesEqual(onDisk, rendered) {
				sidecarNeeded = true
			}
		}
	}
	if sidecarNeeded {
		finalHooks, err := os.ReadFile(hooksPath)
		if err != nil {
			warnf(warn, "re-read %s for sidecar: %v", HooksRelPath, err)
		} else {
			finalCfg := cfgExisting
			if res.ConfigWritten {
				if c, err := os.ReadFile(cfgPath); err == nil {
					finalCfg = c
				}
			}
			if err := writeSidecar(projectRoot, finalHooks, finalCfg); err != nil {
				warnf(warn, "write %s: %v", SidecarPath, err)
			}
		}
	}

	if out != nil {
		switch {
		case res.HooksWritten && !res.HooksExisted:
			_, _ = fmt.Fprintln(out, FirstTrustGuidance)
		case res.HooksWritten && res.HooksExisted:
			_, _ = fmt.Fprintln(out, ReTrustGuidance)
		}
	}

	if len(res.Conflicts) > 0 {
		paths := make([]string, len(res.Conflicts))
		for i, c := range res.Conflicts {
			paths[i] = c.Path
		}
		return fmt.Errorf("%w: %s", ErrWiringConflict, strings.Join(paths, ", "))
	}
	return nil
}

// recordUnchanged brings the part record of a file the pass did not need to
// rewrite up to date. No wiring file changes, so no journal entry is needed.
func (p *pass) recordUnchanged(rel string, recorded map[string]manifest.FileEntry, entry manifest.FileEntry) {
	if old, ok := recorded[rel]; ok && reflect.DeepEqual(old, entry) {
		return
	}
	if err := applyProvenance(p.root, rel, &entry, p.warn); err != nil {
		warnf(p.warn, "record %s ownership: %v", rel, err)
	}
}

// writeSidecar records the sha256 of the generated content.
func writeSidecar(projectRoot string, hooks, config []byte) error {
	doc := sidecarDoc{
		HooksSHA256:  sha256Hex(hooks),
		ConfigSHA256: sha256Hex(config),
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sidecar: %w", err)
	}
	path := filepath.Join(projectRoot, SidecarPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create sidecar directory: %w", err)
	}
	return writeAtomic(path, append(raw, '\n'))
}

// LoadSidecar reads the trust sidecar, if present. A missing sidecar is not an
// error (doctor treats it as "no baseline recorded").
func LoadSidecar(projectRoot string) (sidecarDoc, bool, error) {
	raw, err := os.ReadFile(filepath.Join(projectRoot, SidecarPath))
	if errors.Is(err, os.ErrNotExist) {
		return sidecarDoc{}, false, nil
	}
	if err != nil {
		return sidecarDoc{}, false, fmt.Errorf("read sidecar: %w", err)
	}
	var doc sidecarDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return sidecarDoc{}, false, fmt.Errorf("parse sidecar: %w", err)
	}
	return doc, true, nil
}

// sha256Hex returns the hex sha256 of b.
func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// bytesEqual is a tiny equality helper (avoids importing bytes for two uses).
func bytesEqual(a, b []byte) bool {
	return string(a) == string(b)
}

// writeAtomic writes content to path via temp-file + rename in the target
// directory, so a reader never observes a torn file (D2 atomic write).
func writeAtomic(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".codexwiring-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op after a successful rename
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename into place: %w", err)
	}
	return nil
}
