package cli

// SPEC-V3R6-MOAI-CLEAN-HOME-001 REQ-MCH-001/002/006: the advisory Home Disk
// Usage doctor check. Reports the ~/.moai breakdown, per-profile sizes,
// cross-profile duplicate clusters (report-only heuristic), releases count vs
// the current version, and a ~/.claude summary line marked report-only —
// clean --home never touches ~/.claude.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// profileCategoryStat is one profile top-level category's size and regular-file
// count — report metadata carried alongside a verified tree digest.
type profileCategoryStat struct {
	Size  int64
	Files int
}

// homeDuplicateCluster is a report-only duplicate-plugin finding. Membership
// is established by a SHA-256 digest over sorted paths, modes, and contents;
// size and file count are display metadata, not equality evidence.
type homeDuplicateCluster struct {
	Category string
	Profiles []string
	Size     int64
	Files    int
}

// homeDirStat is one named top-level or per-profile entry's disk stat.
type homeDirStat struct {
	Name  string
	Size  int64
	Files int
}

// homeDiskReport aggregates everything the check renders.
type homeDiskReport struct {
	Root             string
	TotalBytes       int64
	TopLevel         []homeDirStat
	Profiles         []homeDirStat
	ProfileCats      map[string]map[string]profileCategoryStat
	Clusters         []homeDuplicateCluster
	ReleaseCount     int
	CurrentVersion   string
	CleanableBytes   int64
	RetentionDays    int
	ClaudeBytes      int64
	ClaudeExists     bool
	InactiveProfiles []string
	OversizeProfiles []string
	InsecureHomeDirs int
}

// @MX:NOTE: [AUTO] Home Disk Usage check — advisory only; the WARN threshold is the compiled DefaultHomeDiskWarnBytes (config surface per CLAUDE.local.md §14)
func checkHomeDisk(_ bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Home Disk Usage"}

	root, err := paths.MoaiHome()
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = "cannot resolve ~/.moai home"
		check.Detail = err.Error()
		return check
	}
	if _, statErr := os.Stat(root); statErr != nil {
		check.Status = uikit.CheckOK
		check.Message = "no ~/.moai home found — nothing to report"
		return check
	}

	report := gatherHomeDiskReport(root)

	// Compact top-level summary for the one-line message (top 3 by size).
	top := make([]string, 0, 3)
	for i, s := range report.TopLevel {
		if i >= 3 {
			break
		}
		top = append(top, fmt.Sprintf("%s %s", s.Name, formatDiskBytes(s.Size)))
	}
	topStr := strings.Join(top, ", ")
	claudeStr := "absent"
	if report.ClaudeExists {
		claudeStr = formatDiskBytes(report.ClaudeBytes)
	}

	if report.CleanableBytes > config.DefaultHomeDiskWarnBytes {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf(
			"~/.moai %s (%s) — cleanable ~%s exceeds %s threshold; run 'moai clean --home' (dry-run by default); ~/.claude %s (report-only)",
			formatDiskBytes(report.TotalBytes), topStr,
			formatDiskBytes(report.CleanableBytes), formatDiskBytes(config.DefaultHomeDiskWarnBytes),
			claudeStr)
	} else {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf(
			"~/.moai %s (%s) — cleanable ~%s; ~/.claude %s (report-only)",
			formatDiskBytes(report.TotalBytes), topStr,
			formatDiskBytes(report.CleanableBytes), claudeStr)
	}
	check.Detail = renderHomeDiskDetail(report)
	return check
}

// gatherHomeDiskReport walks the home tree once per concern and fills the
// report the check renders.
func gatherHomeDiskReport(root string) homeDiskReport {
	report := homeDiskReport{
		Root:           root,
		ProfileCats:    map[string]map[string]profileCategoryStat{},
		CurrentVersion: version.GetVersion(),
	}

	// Top-level breakdown.
	if entries, err := os.ReadDir(root); err == nil {
		for _, e := range entries {
			size, files := homeEntrySize(filepath.Join(root, e.Name()), e)
			report.TotalBytes += size
			report.TopLevel = append(report.TopLevel, homeDirStat{Name: e.Name(), Size: size, Files: files})
		}
		sort.Slice(report.TopLevel, func(i, j int) bool {
			if report.TopLevel[i].Size != report.TopLevel[j].Size {
				return report.TopLevel[i].Size > report.TopLevel[j].Size
			}
			return report.TopLevel[i].Name < report.TopLevel[j].Name
		})
	}

	// Per-profile sizes + per-category stats (duplicate-cluster inputs).
	profilesDir := filepath.Join(root, defs.ClaudeProfilesSubdir)
	if profiles, err := os.ReadDir(profilesDir); err == nil {
		for _, p := range profiles {
			if !p.IsDir() {
				continue
			}
			profileRoot := filepath.Join(profilesDir, p.Name())
			size, files := homeEntrySize(profileRoot, p)
			report.Profiles = append(report.Profiles, homeDirStat{Name: p.Name(), Size: size, Files: files})
			if size > config.DefaultProfileMaxBytes {
				report.OversizeProfiles = append(report.OversizeProfiles, p.Name())
			}
			lastUsed := newestModTime(filepath.Join(profileRoot, "projects"))
			if lastUsed.IsZero() {
				if info, err := p.Info(); err == nil {
					lastUsed = info.ModTime()
				}
			}
			if !lastUsed.IsZero() && lastUsed.Before(time.Now().AddDate(0, 0, -config.DefaultProfileUnusedDays)) {
				report.InactiveProfiles = append(report.InactiveProfiles, p.Name())
			}
			cats := map[string]profileCategoryStat{}
			if catEntries, err := os.ReadDir(profileRoot); err == nil {
				for _, ce := range catEntries {
					catSize, catFiles := homeEntrySize(filepath.Join(profileRoot, ce.Name()), ce)
					cats[ce.Name()] = profileCategoryStat{Size: catSize, Files: catFiles}
				}
			}
			report.ProfileCats[p.Name()] = cats
		}
		sort.Slice(report.Profiles, func(i, j int) bool {
			if report.Profiles[i].Size != report.Profiles[j].Size {
				return report.Profiles[i].Size > report.Profiles[j].Size
			}
			return report.Profiles[i].Name < report.Profiles[j].Name
		})
	}

	report.Clusters = findPluginHashClusters(profilesDir, report.ProfileCats)
	_ = filepath.WalkDir(root, func(_ string, entry os.DirEntry, walkErr error) error {
		if walkErr == nil && entry.IsDir() {
			if info, err := entry.Info(); err == nil && info.Mode().Perm() != 0o700 {
				report.InsecureHomeDirs++
			}
		}
		return nil
	})

	// Releases count vs current version.
	releasesDir := filepath.Join(root, defs.ReleasesSubdir)
	if entries, err := os.ReadDir(releasesDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, "moai-") && !strings.HasSuffix(name, ".sha256") {
				report.ReleaseCount++
			}
		}
	}

	// Cleanable estimate — the SAME scanner clean --home deletes by, under the
	// home-tier retention, so doctor and clean cannot disagree.
	retention := config.DefaultHomeCleanRetentionDays
	if days, err := loadHomeRetentionDays(); err == nil {
		retention = days
	}
	report.RetentionDays = retention
	for _, c := range scanHomeCleanable(root, retention, config.DefaultReleaseKeep, report.CurrentVersion, time.Now()) {
		report.CleanableBytes += c.Size
	}

	// ~/.claude summary — report-only, never a mutation target.
	if home, err := paths.Home(); err == nil {
		claudeDir := filepath.Join(home, ".claude")
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			size, _, _ := walkHomeSize(claudeDir)
			report.ClaudeExists = true
			report.ClaudeBytes = size
		}
	}
	return report
}

// renderHomeDiskDetail renders the verbose surface: per-profile breakdown,
// duplicate clusters, releases vs current, cleanable estimate, ~/.claude line.
func renderHomeDiskDetail(report homeDiskReport) string {
	var lines []string
	if len(report.Profiles) > 0 {
		parts := make([]string, 0, len(report.Profiles))
		for _, p := range report.Profiles {
			cats := report.ProfileCats[p.Name]
			catNames := make([]string, 0, len(cats))
			for cat := range cats {
				catNames = append(catNames, cat)
			}
			sort.Strings(catNames)
			catParts := make([]string, 0, len(catNames))
			for _, cat := range catNames {
				catParts = append(catParts, fmt.Sprintf("%s %s", cat, formatDiskBytes(cats[cat].Size)))
			}
			parts = append(parts, fmt.Sprintf("%s %s (%s)", p.Name, formatDiskBytes(p.Size), strings.Join(catParts, ", ")))
		}
		lines = append(lines, "profiles: "+strings.Join(parts, "; "))
	}
	if len(report.Clusters) > 0 {
		clusterParts := make([]string, 0, len(report.Clusters))
		for _, c := range report.Clusters {
			clusterParts = append(clusterParts, fmt.Sprintf("%s duplicated across %s (%s each, %d file(s))",
				c.Category, strings.Join(c.Profiles, ", "), formatDiskBytes(c.Size), c.Files))
		}
		lines = append(lines, "duplicate clusters: "+strings.Join(clusterParts, "; ")+" — report-only")
	}
	if len(report.OversizeProfiles) > 0 {
		lines = append(lines, fmt.Sprintf("profiles above %s: %s", formatDiskBytes(config.DefaultProfileMaxBytes), strings.Join(report.OversizeProfiles, ", ")))
	}
	if len(report.InactiveProfiles) > 0 {
		lines = append(lines, fmt.Sprintf("profiles unused for %d+ days: %s (report-only)", config.DefaultProfileUnusedDays, strings.Join(report.InactiveProfiles, ", ")))
	}
	if report.InsecureHomeDirs > 0 {
		lines = append(lines, fmt.Sprintf("~/.moai directories not mode 0700: %d (repaired by 'moai clean --home --force')", report.InsecureHomeDirs))
	}
	lines = append(lines, fmt.Sprintf("releases: %d binary/binary-set entr(ies) vs current %s (keep %d beyond current)",
		report.ReleaseCount, report.CurrentVersion, config.DefaultReleaseKeep))
	lines = append(lines, fmt.Sprintf("cleanable estimate: ~%s under %dd retention — 'moai clean --home' (dry-run by default)",
		formatDiskBytes(report.CleanableBytes), report.RetentionDays))
	if report.ClaudeExists {
		lines = append(lines, fmt.Sprintf("~/.claude %s (report-only — moai clean never touches it)", formatDiskBytes(report.ClaudeBytes)))
	} else {
		lines = append(lines, "~/.claude absent (report-only — moai clean never touches it)")
	}
	return strings.Join(lines, "\n")
}

func newestModTime(root string) time.Time {
	var newest time.Time
	_ = filepath.WalkDir(root, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if info, infoErr := entry.Info(); infoErr == nil && info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	return newest
}

// findPluginHashClusters reports byte-identical plugin trees. The digest is
// over sorted relative paths, file modes and contents, so equal aggregate
// sizes alone never produce a duplicate claim.
func findPluginHashClusters(profilesDir string, perProfile map[string]map[string]profileCategoryStat) []homeDuplicateCluster {
	byDigest := map[string][]string{}
	stats := map[string]profileCategoryStat{}
	for profile, cats := range perProfile {
		stat, ok := cats["plugins"]
		if !ok || stat.Files == 0 {
			continue
		}
		digest, err := treeDigest(filepath.Join(profilesDir, profile, "plugins"))
		if err != nil {
			continue
		}
		byDigest[digest] = append(byDigest[digest], profile)
		stats[digest] = stat
	}
	var out []homeDuplicateCluster
	for digest, profiles := range byDigest {
		if len(profiles) < 2 {
			continue
		}
		sort.Strings(profiles)
		st := stats[digest]
		out = append(out, homeDuplicateCluster{Category: "plugins sha256=" + digest[:12], Profiles: profiles, Size: st.Size, Files: st.Files})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Category < out[j].Category })
	return out
}

func treeDigest(root string) (string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, path := range files {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		_, _ = io.WriteString(h, filepath.ToSlash(rel)+"\x00"+info.Mode().String()+"\x00")
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(h, file)
		closeErr := file.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// homeEntrySize returns the size and regular-file count of a ReadDir entry
// (walking directories).
func homeEntrySize(abs string, e os.DirEntry) (int64, int) {
	if !e.IsDir() {
		if info, err := e.Info(); err == nil {
			return info.Size(), 1
		}
		return 0, 0
	}
	size, files, err := walkHomeSize(abs)
	if err != nil {
		return 0, 0
	}
	return size, files
}

// formatDiskBytes renders a byte count in 1024-based du-style units.
func formatDiskBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(b)/float64(div), "KMGTPE"[exp])
}
