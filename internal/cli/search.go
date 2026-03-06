//go:build !race

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/search"
)

// searchCmd is the session history search subcommand.
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search indexed session history",
	Long: `Search indexed Claude Code session history using SQLite FTS5 full-text search.

Supports Korean/CJK text search via trigram tokenizer.
Sessions are automatically indexed via SessionEnd hook.

Examples:
  moai search "JWT 인증"           # Full-text search
  moai search "인덱스" --branch main  # Filter by branch
  moai search "플로우" --role user    # Filter by role
  moai search --index-session abc123  # Manually index a session`,
	GroupID: "tools",
	RunE:    runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Search filter flags.
	searchCmd.Flags().StringP("branch", "b", "", "git branch filter")
	searchCmd.Flags().String("since", "", "start date filter (RFC3339, e.g. 2026-01-01T00:00:00Z)")
	searchCmd.Flags().String("until", "", "end date filter (RFC3339, e.g. 2026-12-31T23:59:59Z)")
	searchCmd.Flags().String("role", "", "role filter: user|assistant")
	searchCmd.Flags().IntP("limit", "n", 20, "maximum number of results")

	// Indexing flags.
	searchCmd.Flags().String("index-session", "", "manually index a specific session ID")
	searchCmd.Flags().String("project-path", "", "project path to use during indexing")
	searchCmd.Flags().String("git-branch", "", "git branch to use during indexing")
}

// runSearch is the execution handler for the search command.
// If --index-session is provided, it indexes the session and exits.
// Otherwise it accepts a query argument and prints results.
func runSearch(cmd *cobra.Command, args []string) error {
	sessionID, _ := cmd.Flags().GetString("index-session")

	// Indexing mode.
	if sessionID != "" {
		return runIndexSession(cmd, sessionID)
	}

	// Search mode: query argument is required.
	if len(args) == 0 {
		return errors.New("search query is required, e.g.: moai search \"JWT auth\"")
	}

	return runSearchQuery(cmd, args[0])
}

// runIndexSession locates the JSONL file for the given session ID and indexes it.
func runIndexSession(cmd *cobra.Command, sessionID string) error {
	projectPath, _ := cmd.Flags().GetString("project-path")
	gitBranch, _ := cmd.Flags().GetString("git-branch")

	dbPath, err := searchDBPath()
	if err != nil {
		return err
	}

	db, err := search.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := search.CreateTables(db); err != nil {
		return fmt.Errorf("failed to initialize DB tables: %w", err)
	}

	// Locate the JSONL file under ~/.claude/projects/.
	filePath, err := findSessionJSONL(sessionID)
	if err != nil {
		return fmt.Errorf("session JSONL file not found (session=%s): %w", sessionID, err)
	}

	if err := search.IndexSession(db, sessionID, filePath, gitBranch, projectPath); err != nil {
		return fmt.Errorf("failed to index session: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "session indexed: %s\n", sessionID)
	return nil
}

// runSearchQuery executes the query and prints results.
func runSearchQuery(cmd *cobra.Command, query string) error {
	branch, _ := cmd.Flags().GetString("branch")
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	role, _ := cmd.Flags().GetString("role")
	limit, _ := cmd.Flags().GetInt("limit")

	dbPath, err := searchDBPath()
	if err != nil {
		return err
	}

	db, err := search.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := search.CreateTables(db); err != nil {
		return fmt.Errorf("failed to initialize DB tables: %w", err)
	}

	opts := search.SearchOptions{
		Query:  query,
		Branch: branch,
		Since:  since,
		Until:  until,
		Role:   role,
		Limit:  limit,
	}

	results, err := search.Search(db, opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	printSearchResults(cmd, query, results)
	return nil
}

// printSearchResults renders search results to the terminal.
func printSearchResults(cmd *cobra.Command, query string, results []search.SearchResult) {
	out := cmd.OutOrStdout()

	if len(results) == 0 {
		_, _ = fmt.Fprintf(out, "no results for: %q\n", query)
		return
	}

	// Print header.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#C45A3C", Dark: "#DA7756",
	})
	_, _ = fmt.Fprintln(out, headerStyle.Render(fmt.Sprintf("results for %q (%d found)", query, len(results))))
	_, _ = fmt.Fprintln(out)

	// Print each result as a card.
	roleStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#059669", Dark: "#10B981",
	})
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#9CA3AF", Dark: "#6B7280",
	})
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}).
		Padding(0, 2)

	for i, r := range results {
		// Build card content.
		meta := fmt.Sprintf("%s  %s  %s",
			roleStyle.Render(r.Role),
			mutedStyle.Render(r.GitBranch),
			mutedStyle.Render(r.Timestamp),
		)
		excerpt := r.Excerpt
		if excerpt == "" {
			excerpt = "(empty)"
		}
		// Truncate session ID for display.
		shortSession := r.SessionID
		if len(shortSession) > 12 {
			shortSession = shortSession[:12] + "..."
		}

		content := meta + "\n" + excerpt + "\n" + mutedStyle.Render(fmt.Sprintf("session: %s", shortSession))
		_, _ = fmt.Fprintln(out, borderStyle.Render(content))

		// No separator after the last result.
		if i < len(results)-1 {
			_, _ = fmt.Fprintln(out)
		}
	}
}

// searchDBPath returns the path to the search database file.
// Path: ~/.moai/search/sessions.db
func searchDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(homeDir, defs.MoAIDir, defs.SearchSubdir, defs.SearchDB), nil
}

// findSessionJSONL searches subdirectories under ~/.claude/projects/ for the
// {sessionId}.jsonl file and returns its path.
func findSessionJSONL(sessionID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	targetFile := sessionID + ".jsonl"

	// Walk subdirectories to find the JSONL file.
	var found string
	err = filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible entries.
			return nil
		}
		if !d.IsDir() && d.Name() == targetFile {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if found == "" {
		return "", fmt.Errorf("%s.jsonl not found under %s",
			sessionID, projectsDir)
	}

	return found, nil
}
