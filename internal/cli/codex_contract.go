package cli

// codex_contract.go — SPEC-CODEX-INIT-001 M4+M5+M6 (REQ-CI-005..008,
// REQ-CI-011): the instruction contract of the AGENTS.md-primary product.
// Connection only — never content: a missing AGENTS.md is created with a
// minimal body; an existing AGENTS.md is preserved byte-for-byte. A legacy
// CLAUDE.md is inert here — the contract neither reads nor writes it (the
// product ships no CLAUDE.md; the legacy CLAUDE.local.md move belongs to
// `moai migrate local-instructions`). Every write is a per-file temp+rename;
// no truncating write path exists here at all (AC-CI-010's open cell
// observes that).
//
// Containment comes FIRST (M4 before M5 — plan §D): before any touch of an
// instruction path, every existing path component must Lstat as a plain
// directory (the leaf: a regular file), and the symlink-and-..-resolved
// absolute path must stay inside the project root (REQ-CI-011 — judged by
// STRUCTURE, not by a snapshot; a symlink placed inside the project would
// pass any snapshot while writing outside). Refusal is issued WITHOUT
// reading the path — REQ-CI-011 forbids read too, so a guard that reads and
// discards is not a guard.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Instruction paths — the single source the contract guards. AGENTS.md is
// the sole instruction file; AGENTS.local.md is the user-owned local file
// neither harness reaches through an import — Codex through the launcher's
// developer_instructions injection (SPEC-INSTRUCTION-FILES-UNIFY-001
// REQ-IFU-002/006), Claude through the moai launcher. The legacy
// CLAUDE.local.md is read ahead of it by the launcher but never written
// here.
const (
	codexAgentsRelPath        = "AGENTS.md"
	codexLocalInstructionName = "AGENTS.local.md"
	codexClaudeLocalName      = "CLAUDE.local.md"
)

// Created body — minimal and non-empty (AC-CI-005 requires a created
// AGENTS.md to carry at least one non-space character; body QUALITY is out
// of this SPEC's scope). A created AGENTS.md imports nothing — it is the
// neutral contract.
const codexCreatedAgentsBody = "# AGENTS.md\n"

// codexInstructionRelPathsFn is the path-table seam: the instruction paths
// the contract guards, in guard order. The tests override it with variant
// spellings (docs/<name>, ..-escapes) to drive the parent-component escape
// axis through the real launch verbs.
var codexInstructionRelPathsFn = defaultCodexInstructionRelPaths

func defaultCodexInstructionRelPaths() []string {
	return []string{codexAgentsRelPath, codexLocalInstructionName}
}

// Filesystem seams — EVERY touch of an instruction path goes through one of
// these, which is what lets the acceptance count reads, refuse truncating
// writes, and inject Lstat modes mechanically.
var (
	codexLstatFn        = os.Lstat
	codexReadFileFn     = os.ReadFile
	codexCreateTempFn   = os.CreateTemp
	codexRenameFn       = os.Rename
	codexEvalSymlinksFn = filepath.EvalSymlinks
)

// codexPathGuardError names the refused instruction path — the diagnostic
// the cells require on every refusal (AC-CI-011).
type codexPathGuardError struct {
	Rel    string
	Reason string
}

func (e *codexPathGuardError) Error() string {
	return fmt.Sprintf("refusing instruction path %q: %s", e.Rel, e.Reason)
}

// codexGuardInstructionPath judges rel (slash-separated, relative to the
// project root) by structure:
//
//   - every existing component except the leaf must Lstat as a DIRECTORY —
//     a symlink, FIFO, socket, or device there is refused;
//   - the leaf, when it exists, must be a REGULAR file — the closed-set
//     judgement is IsRegular, never an enumeration of kinds;
//   - the first missing component ends the walk (the remainder is created);
//   - the ..-resolved absolute path must remain inside the project root.
//
// It returns whether the leaf currently exists, and never opens or reads the
// path itself (Lstat is a component probe, not a read).
func codexGuardInstructionPath(projectRoot, rel string) (bool, error) {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return false, &codexPathGuardError{Rel: rel, Reason: fmt.Sprintf("resolve project root: %v", err)}
	}
	if resolved, rerr := codexEvalSymlinksFn(absRoot); rerr == nil && resolved != "" {
		absRoot = resolved
	}

	cur := absRoot
	exists := true
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for i, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			cur = filepath.Dir(cur)
			continue
		}
		probe := filepath.Join(cur, filepath.FromSlash(part))
		info, lerr := codexLstatFn(probe)
		if lerr != nil {
			if os.IsNotExist(lerr) {
				exists = false
				break
			}
			return false, &codexPathGuardError{Rel: rel, Reason: fmt.Sprintf("probe %q: %v", part, lerr)}
		}
		isLeaf := i == len(parts)-1
		switch {
		case isLeaf && !info.Mode().IsRegular():
			return false, &codexPathGuardError{Rel: rel, Reason: "not a regular file (" + codexModeName(info.Mode()) + ")"}
		case !isLeaf && !info.Mode().IsDir():
			return false, &codexPathGuardError{Rel: rel, Reason: fmt.Sprintf("parent component %q is not a directory (%s)", part, codexModeName(info.Mode()))}
		}
		cur = probe
		exists = true
	}

	// Containment: the ..-resolved path must stay inside the project root.
	// No lexical Clean-before-resolve shortcut survives this — the walk
	// above already refused every symlink component.
	relToRoot, rerr := filepath.Rel(absRoot, cur)
	if rerr != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) || filepath.IsAbs(relToRoot) {
		return exists, &codexPathGuardError{Rel: rel, Reason: "resolves outside the project root"}
	}
	return exists, nil
}

// codexModeName renders a mode for DIAGNOSTICS only — the judgement itself
// is the closed-set IsRegular/IsDir structure above.
func codexModeName(m os.FileMode) string {
	switch {
	case m.IsDir():
		return "directory"
	case m&os.ModeSymlink != 0:
		return "symlink"
	case m&os.ModeNamedPipe != 0:
		return "named pipe"
	case m&os.ModeSocket != 0:
		return "socket"
	case m&os.ModeDevice != 0 && m&os.ModeCharDevice != 0:
		return "character device"
	case m&os.ModeDevice != 0:
		return "device"
	case m&os.ModeIrregular != 0:
		return "irregular"
	default:
		return m.String()
	}
}

// @MX:ANCHOR: [AUTO] the contract is the single writer of instruction links
// for the codex init gate — reached by both launch verbs and both launch
// paths, containment-guarded before any touch.
// @MX:REASON: ordering invariant (guard → plan → stage-all → rename-all)
// is what keeps a mid-write failure from ever truncating a user file
// (REQ-CI-006/011, AC-CI-010's open cell).
func secureCodexInstructionContract(req codexContractRequest) error {
	root := req.ProjectRoot
	rels := codexInstructionRelPathsFn()
	if len(rels) != 2 {
		return fmt.Errorf("instruction path table must name exactly two paths, got %d", len(rels))
	}

	// 1. Containment guards — BEFORE any read or write of the paths.
	exists := make([]bool, len(rels))
	for i, rel := range rels {
		ok, gerr := codexGuardInstructionPath(root, rel)
		if gerr != nil {
			return gerr
		}
		exists[i] = ok
	}

	// 2. Read + plan — compute EVERY intended change before applying any.
	// Only AGENTS.md is ever written (created when absent); AGENTS.local.md
	// is guard-only, and a legacy CLAUDE.md is neither read nor written.
	type codexContractPlan struct {
		rel     string
		content []byte
	}
	var plans []codexContractPlan

	_, aerr := codexReadFileFn(filepath.Join(root, filepath.FromSlash(rels[0])))
	agentsExists := aerr == nil
	if aerr != nil && !errors.Is(aerr, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", rels[0], aerr)
	}
	if !agentsExists {
		plans = append(plans, codexContractPlan{rel: rels[0], content: []byte(codexCreatedAgentsBody)})
	}

	// 3. Stage ALL temp files first. A staging failure leaves every target
	// byte-identical — the failure-path cell (E2) observes exactly that.
	type codexStagedFile struct {
		rel    string
		tmp    string
		target string
	}
	staged := make([]codexStagedFile, 0, len(plans))
	defer func() {
		// Best-effort residue cleanup: any temp that never reached its
		// rename must not linger in the project tree. Our temp paths only —
		// never a target path.
		for _, s := range staged {
			if s.tmp == "" {
				continue
			}
			if err := os.Remove(s.tmp); err != nil && !errors.Is(err, os.ErrNotExist) {
				codexGatePrintf(req.Err, "codex init: residue %s: %v\n", s.tmp, err)
			}
		}
	}()
	for _, p := range plans {
		dir := filepath.Dir(filepath.Join(root, filepath.FromSlash(p.rel)))
		tmp, terr := codexCreateTempFn(dir, ".codex-init-*")
		if terr != nil {
			return fmt.Errorf("stage %s: %w", p.rel, terr)
		}
		if _, werr := tmp.Write(p.content); werr != nil {
			_ = tmp.Close() // the write error is the one reported
			return fmt.Errorf("write %s: %w", p.rel, werr)
		}
		if cherr := tmp.Chmod(0o644); cherr != nil {
			_ = tmp.Close()
			return fmt.Errorf("chmod %s: %w", p.rel, cherr)
		}
		if cerr2 := tmp.Close(); cerr2 != nil {
			return fmt.Errorf("close %s: %w", p.rel, cerr2)
		}
		staged = append(staged, codexStagedFile{
			rel:    p.rel,
			tmp:    tmp.Name(),
			target: filepath.Join(root, filepath.FromSlash(p.rel)),
		})
	}

	// 4. Install — one rename per target file, in table order.
	for i := range staged {
		if rerr := codexRenameFn(staged[i].tmp, staged[i].target); rerr != nil {
			staged[i].tmp = "" // renamed: the cleanup must not remove it
			return fmt.Errorf("install %s: %w", staged[i].rel, rerr)
		}
		staged[i].tmp = ""
	}
	if len(staged) > 0 {
		codexGatePrintf(req.Out, "codex init: instruction file ready (%s)\n", rels[0])
	}
	return nil
}
