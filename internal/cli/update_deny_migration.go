package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// retiredV2DenyEntries lists the permission deny rules the v2-era settings.json
// template shipped but the v3 template retired (issue #1101). The v2->v3
// clean-reinstall path preserves the user's settings.json wholesale (no embedded
// base exists for a 3-way merge — only settings.json.tmpl ships in the embedded
// FS), so without this one-shot strip the retired entries survive every upgrade
// and produce Claude Code startup warnings. Matching is exact-string only so
// genuine user customizations are never touched. Source of truth: the entries
// removed from settings.json.tmpl by commit "chore(settings): trim redundant
// Write/Grep/Glob secret-path deny rules".
var retiredV2DenyEntries = []string{
	"Write(./secrets/**)",
	"Write(~/.ssh/**)",
	"Write(~/.aws/**)",
	"Write(~/.config/gcloud/**)",
	"Grep(./secrets/**)",
	"Grep(~/.ssh/**)",
	"Grep(~/.aws/**)",
	"Grep(~/.config/gcloud/**)",
	"Glob(./secrets/**)",
	"Glob(~/.ssh/**)",
	"Glob(~/.aws/**)",
	"Glob(~/.config/gcloud/**)",
}

// stripRetiredV2DenyEntries removes the retired v2 deny rules from
// .claude/settings.json under projectRoot. It round-trips the file as
// map[string]any so unknown keys are never wiped (SPEC-CLIFIX-CRITICAL-001
// precedent), preserves list order, and rewrites the file ONLY when at least
// one retired entry was actually removed — a v3-clean file and a second run
// are byte-for-byte no-ops. A missing file, missing permissions key, or
// missing deny list is a silent no-op.
func stripRetiredV2DenyEntries(projectRoot string, out io.Writer) error {
	path := filepath.Join(projectRoot, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read settings.json: %w", err)
	}

	m := make(map[string]any)
	if err := json.Unmarshal(data, &m); err != nil {
		// A malformed settings.json is not this migration's problem; leave it
		// untouched rather than failing the whole clean-reinstall.
		return nil
	}

	perms, ok := m["permissions"].(map[string]any)
	if !ok {
		return nil
	}
	deny, ok := perms["deny"].([]any)
	if !ok {
		return nil
	}

	retired := make(map[string]struct{}, len(retiredV2DenyEntries))
	for _, e := range retiredV2DenyEntries {
		retired[e] = struct{}{}
	}

	kept := make([]any, 0, len(deny))
	removed := 0
	for _, e := range deny {
		if s, isStr := e.(string); isStr {
			if _, isRetired := retired[s]; isRetired {
				removed++
				continue
			}
		}
		kept = append(kept, e)
	}
	if removed == 0 {
		return nil
	}

	perms["deny"] = kept
	outData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.json: %w", err)
	}
	outData = append(outData, '\n')

	info, statErr := os.Stat(path)
	mode := os.FileMode(0o644)
	if statErr == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(path, outData, mode); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}
	_, _ = fmt.Fprintf(out, "[clean-reinstall] Removed %d retired v2 permission deny entries from settings.json\n", removed)
	return nil
}
