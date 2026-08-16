// migrate_profiles.go — `moai migrate profiles`.
//
// Consolidates the per-profile memory stores into the default one.
//
// `moai cc -p <name>` sets CLAUDE_CONFIG_DIR, and Claude Code puts
// projects/<slug>/memory under whatever that points at. The effect is that
// the same project accumulates a separate memory per profile, and a session
// launched one way cannot recall what a session launched the other way
// learned. Memory is project-scoped by nature; the profile is an outer key
// that does not belong on it.
//
// The migration is explicit, never automatic: `moai update` only reports that
// there is something to merge. Moving hundreds of files inside a user's home
// as a side effect of an update is the shape of an accident this project has
// already had once.
package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// migrateProfileOpts controls one migration run.
type migrateProfileOpts struct {
	// DryRun reports what would happen and writes nothing.
	DryRun bool
}

// migrateProfileResult is what a run did (or would do, under DryRun).
type migrateProfileResult struct {
	// Profiles are the profile names that held memory for this project.
	Profiles []string
	// Moved counts files that landed in the default store under their own name.
	Moved int
	// Renamed counts files kept under a profile-suffixed name because the
	// default store already had that name with different content.
	Renamed int
	// Skipped counts files already present with identical content.
	Skipped int
	// IndexLines counts index entries carried into the default MEMORY.md.
	IndexLines int
}

// profileMemoryStores returns the per-profile memory directories that hold
// something for this project, keyed by profile name.
func profileMemoryStores(projectRoot string) (map[string]string, error) {
	home, err := userHomeDir()
	if err != nil {
		return nil, fmt.Errorf("migrate profiles: resolve home: %w", err)
	}
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("migrate profiles: resolve project root: %w", err)
	}
	slug := memoryProjectSlug(abs)

	profilesRoot := filepath.Join(home, ".moai", "claude-profiles")
	entries, err := os.ReadDir(profilesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("migrate profiles: read %s: %w", profilesRoot, err)
	}

	stores := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(profilesRoot, e.Name(), "projects", slug, "memory")
		names, err := topicFileNames(dir)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			continue
		}
		stores[e.Name()] = dir
	}
	return stores, nil
}

// topicFileNames lists .md topic files directly under dir, excluding the index.
func topicFileNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("migrate profiles: read %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && e.Name() != memoryIndexName {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// defaultMemoryStore is where the migration consolidates to: the store a
// session uses when no profile is active.
func defaultMemoryStore(projectRoot string) (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("migrate profiles: resolve home: %w", err)
	}
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("migrate profiles: resolve project root: %w", err)
	}
	return filepath.Join(home, ".claude", "projects", memoryProjectSlug(abs), "memory"), nil
}

// migrateProfileMemory merges every profile store for this project into the
// default one.
//
// Collisions never overwrite. A file whose name already exists in the target
// is compared byte for byte: identical content is already migrated and is
// skipped, differing content is kept alongside under a profile-suffixed name.
// Losing one of two same-named memories would be silent and unrecoverable,
// and a duplicate is cheap by comparison.
func migrateProfileMemory(projectRoot string, opts migrateProfileOpts) (*migrateProfileResult, error) {
	stores, err := profileMemoryStores(projectRoot)
	if err != nil {
		return nil, err
	}
	res := &migrateProfileResult{}
	if len(stores) == 0 {
		return res, nil
	}

	target, err := defaultMemoryStore(projectRoot)
	if err != nil {
		return nil, err
	}
	if !opts.DryRun {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return nil, fmt.Errorf("migrate profiles: create %s: %w", target, err)
		}
	}

	profiles := make([]string, 0, len(stores))
	for name := range stores {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	res.Profiles = profiles

	// renames maps the original filename to the name it landed under, so the
	// index merge can rewrite the link of anything that was suffixed.
	renames := map[string]map[string]string{}

	for _, profile := range profiles {
		src := stores[profile]
		names, err := topicFileNames(src)
		if err != nil {
			return nil, err
		}
		renames[profile] = map[string]string{}

		for _, name := range names {
			srcPath := filepath.Join(src, name)
			dstName := name
			dstPath := filepath.Join(target, dstName)

			existing, statErr := os.ReadFile(dstPath)
			switch {
			case statErr == nil:
				incoming, readErr := os.ReadFile(srcPath)
				if readErr != nil {
					return nil, fmt.Errorf("migrate profiles: read %s: %w", srcPath, readErr)
				}
				if bytes.Equal(existing, incoming) {
					res.Skipped++
					if !opts.DryRun {
						if err := os.Remove(srcPath); err != nil {
							return nil, fmt.Errorf("migrate profiles: drop duplicate %s: %w", srcPath, err)
						}
					}
					continue
				}
				dstName = suffixedMemoryName(name, profile)
				dstPath = filepath.Join(target, dstName)
				renames[profile][name] = dstName
				res.Renamed++
			case os.IsNotExist(statErr):
				res.Moved++
			default:
				return nil, fmt.Errorf("migrate profiles: stat %s: %w", dstPath, statErr)
			}

			if opts.DryRun {
				continue
			}
			if err := moveFile(srcPath, dstPath); err != nil {
				return nil, err
			}
		}
	}

	lines, err := mergeMemoryIndexes(target, stores, profiles, renames, opts.DryRun)
	if err != nil {
		return nil, err
	}
	res.IndexLines = lines
	return res, nil
}

// suffixedMemoryName appends the profile name before the extension.
func suffixedMemoryName(name, profile string) string {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s__%s%s", stem, profile, ext)
}

// moveFile relocates src to dst, falling back to copy+remove when the two
// sit on different filesystems (a profile root and the home root need not be
// on the same volume).
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("migrate profiles: read %s: %w", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("migrate profiles: write %s: %w", dst, err)
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("migrate profiles: remove %s: %w", src, err)
	}
	return nil
}

// mergeMemoryIndexes appends each profile index's entries to the default
// index, skipping links the default already carries and rewriting the link of
// any file that had to be renamed. Existing entries are never rewritten or
// reordered — the default index is the one a session has been reading.
func mergeMemoryIndexes(target string, stores map[string]string, profiles []string, renames map[string]map[string]string, dryRun bool) (int, error) {
	targetIndex := filepath.Join(target, memoryIndexName)
	existing, err := os.ReadFile(targetIndex)
	if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("migrate profiles: read %s: %w", targetIndex, err)
	}

	present := map[string]bool{}
	for _, m := range markdownLinkTargetCLI.FindAllStringSubmatch(string(existing), -1) {
		present[filepath.Base(m[1])] = true
	}

	var added []string
	for _, profile := range profiles {
		data, err := os.ReadFile(filepath.Join(stores[profile], memoryIndexName))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return 0, fmt.Errorf("migrate profiles: read %s index: %w", profile, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			matches := markdownLinkTargetCLI.FindAllStringSubmatch(line, -1)
			if len(matches) == 0 {
				continue
			}
			name := filepath.Base(matches[0][1])
			if renamed, ok := renames[profile][name]; ok {
				line = strings.Replace(line, matches[0][1], renamed, 1)
				name = renamed
			}
			if present[name] {
				continue
			}
			present[name] = true
			added = append(added, line)
		}
	}
	if len(added) == 0 || dryRun {
		return len(added), nil
	}

	var out bytes.Buffer
	out.Write(existing)
	if len(existing) > 0 && !bytes.HasSuffix(existing, []byte("\n")) {
		out.WriteString("\n")
	}
	out.WriteString("\n## Merged from profiles\n")
	for _, line := range added {
		out.WriteString(line + "\n")
	}
	if err := os.WriteFile(targetIndex, out.Bytes(), 0o644); err != nil {
		return 0, fmt.Errorf("migrate profiles: write %s: %w", targetIndex, err)
	}
	return len(added), nil
}

// renderMigrateProfileResult prints the outcome.
func renderMigrateProfileResult(out io.Writer, res *migrateProfileResult, dryRun bool) {
	if len(res.Profiles) == 0 {
		_, _ = fmt.Fprintln(out, "no profile memory found for this project — nothing to migrate")
		return
	}
	verb := "migrated"
	if dryRun {
		verb = "would migrate"
	}
	_, _ = fmt.Fprintf(out, "profiles: %s\n", strings.Join(res.Profiles, ", "))
	_, _ = fmt.Fprintf(out, "%s %d file(s); %d kept alongside on name collision; %d already present\n",
		verb, res.Moved, res.Renamed, res.Skipped)
	_, _ = fmt.Fprintf(out, "index entries carried: %d\n", res.IndexLines)
	if res.Renamed > 0 {
		_, _ = fmt.Fprintln(out,
			"collisions were kept under <name>__<profile>.md — neither version was overwritten")
	}
	if dryRun {
		_, _ = fmt.Fprintln(out, "\ndry run: nothing was written. Re-run without --dry-run to apply.")
	}
}

// migrateProfileAdvisory returns the one-line notice `moai update` prints
// when profile memory is still lying around, or "" when there is none.
// It never migrates: an update that moved hundreds of files inside a home
// directory as a side effect is precisely the accident this project has
// already had once.
func migrateProfileAdvisory(projectRoot string) string {
	stores, err := profileMemoryStores(projectRoot)
	if err != nil || len(stores) == 0 {
		return ""
	}
	total := 0
	for _, dir := range stores {
		names, err := topicFileNames(dir)
		if err != nil {
			return ""
		}
		total += len(names)
	}
	if total == 0 {
		return ""
	}
	return fmt.Sprintf(
		"%d memory file(s) sit in %d profile store(s) that a default session cannot read. "+
			"Run `moai migrate profiles --dry-run` to preview merging them.",
		total, len(stores))
}

func newMigrateProfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles",
		Short: "Merge per-profile memory into the default store",
		Long: `Merge this project's per-profile memory stores into the default store.

` + "`moai cc -p <name>`" + ` sets CLAUDE_CONFIG_DIR, and Claude Code keeps
projects/<slug>/memory under whatever that points at — so the same project
accumulates a separate memory per profile, and a session launched one way
cannot recall what a session launched the other way learned.

Collisions never overwrite. A name that already exists in the default store is
compared byte for byte: identical content is skipped, differing content is kept
alongside as <name>__<profile>.md.

Sessions (transcripts, file-history, session-env) are NOT migrated here; they
span six directories and are a separate step.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			res, err := migrateProfileMemory(resolveProjectDir(), migrateProfileOpts{DryRun: dryRun})
			if err != nil {
				return err
			}
			renderMigrateProfileResult(cmd.OutOrStdout(), res, dryRun)
			return nil
		},
	}
	cmd.Flags().Bool("dry-run", false, "Show planned actions without modifying files")
	return cmd
}

func init() {
	migrateCmd.AddCommand(newMigrateProfilesCmd())
}
