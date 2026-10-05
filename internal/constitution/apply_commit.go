package constitution

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Three-file atomic apply (SPEC-CON-AMEND-APPLY-001 REQ-CAA-010, REQ-CAA-011,
// REQ-CAA-018): back up every existing file, write every new content to a
// temporary file beside its target, rename source, registry, then log into
// place; on any failure restore all three from the backups.

// fileChange is one of the three files the apply step writes.
type fileChange struct {
	role    string // "rule file", "registry", "evolution log"
	path    string
	oldData []byte
	existed bool
	mode    fs.FileMode
	newData []byte
	backup  string // backup file path; "" when the file did not exist
	temp    string // temporary file path; "" once consumed or never written
}

// readForChange reads a target's pre-apply bytes and mode. A file that does
// not exist yields existed=false only when allowMissing is set.
func readForChange(role, path string, allowMissing bool) (*fileChange, error) {
	c := &fileChange{role: role, path: path, mode: 0o644}
	info, err := os.Stat(path)
	switch {
	case err == nil:
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil, fmt.Errorf("%s %s: %w", role, path, rerr)
		}
		c.oldData, c.existed, c.mode = data, true, info.Mode().Perm()
	case allowMissing && errors.Is(err, fs.ErrNotExist):
	default:
		return nil, fmt.Errorf("%s %s: %w", role, path, err)
	}
	return c, nil
}

// writeSibling writes data to a new file beside path named
// .<base>.amend-<kind>-<random>, so a leftover inside .claude/rules is never
// loaded as a markdown rule and every rename stays on one filesystem.
func writeSibling(path, kind string, data []byte, mode fs.FileMode) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".amend-"+kind+"-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	_, werr := f.Write(data)
	cerr := f.Chmod(mode)
	if err := errors.Join(werr, cerr, f.Close()); err != nil {
		_ = os.Remove(name)
		return "", err
	}
	return name, nil
}

// restoreFile is the default restore operation: it writes a file's pre-apply
// bytes back, or removes the file when it did not exist before the apply.
func restoreFile(path string, data []byte, existed bool) error {
	if !existed {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}

// renameOp returns the forward-rename seam.
func (p *Pipeline) renameOp() func(string, string) error {
	if p.rename != nil {
		return p.rename
	}
	return os.Rename
}

// restoreOp returns the restore seam; it never routes through renameOp.
func (p *Pipeline) restoreOp() func(string, []byte, bool) error {
	if p.restore != nil {
		return p.restore
	}
	return restoreFile
}

// commitChanges writes the three files atomically as a group.
//
// @MX:WARN: [AUTO] multi-file write with rollback; a partial failure leaves recovery to the backups
// @MX:REASON: REQ-CAA-010/018 — every failure path must restore all three files, and a failed restore must keep every backup and name it
func (p *Pipeline) commitChanges(changes []*fileChange) error {
	fail := func(step string, cause error) error {
		if rerr := p.restoreAll(changes); rerr != nil {
			removeTemps(changes)
			return fmt.Errorf("%s failed (%v); %w", step, cause, rerr)
		}
		removeTemps(changes)
		removeBackups(changes)
		return fmt.Errorf("%s failed; the rule file, the registry, and the evolution log were restored: %w", step, cause)
	}

	for _, c := range changes {
		if !c.existed {
			continue
		}
		b, err := writeSibling(c.path, "bak", c.oldData, c.mode)
		if err != nil {
			return fail("backup of "+c.role+" "+c.path, err)
		}
		c.backup = b
	}
	for _, c := range changes {
		tmp, err := writeSibling(c.path, "tmp", c.newData, c.mode)
		if err != nil {
			return fail("temporary write of "+c.role+" "+c.path, err)
		}
		c.temp = tmp
	}
	rename := p.renameOp()
	for _, c := range changes {
		if err := rename(c.temp, c.path); err != nil {
			return fail("rename of "+c.role+" "+c.path, err)
		}
		c.temp = ""
	}
	removeBackups(changes)
	return nil
}

// restoreAll restores every file from its pre-apply bytes, whether or not its
// rename reported success. When any restore fails, it deletes no backup and
// returns an error naming every backup path and every failed restore.
func (p *Pipeline) restoreAll(changes []*fileChange) error {
	restore := p.restoreOp()
	var failed []string
	for _, c := range changes {
		if err := restore(c.path, c.oldData, c.existed); err != nil {
			failed = append(failed, fmt.Sprintf("restore of %s %s: %v", c.role, c.path, err))
		}
	}
	if len(failed) == 0 {
		return nil
	}
	var backups []string
	for _, c := range changes {
		if c.backup != "" {
			backups = append(backups, c.backup)
		}
	}
	return fmt.Errorf("%s; pre-apply backups kept for manual recovery: %s",
		strings.Join(failed, "; "), strings.Join(backups, ", "))
}

// removeTemps removes every temporary file not yet renamed into place.
func removeTemps(changes []*fileChange) {
	for _, c := range changes {
		if c.temp != "" {
			_ = os.Remove(c.temp)
			c.temp = ""
		}
	}
}

// removeBackups removes every backup file.
func removeBackups(changes []*fileChange) {
	for _, c := range changes {
		if c.backup != "" {
			_ = os.Remove(c.backup)
			c.backup = ""
		}
	}
}
