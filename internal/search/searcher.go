//go:build !race

package search

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchOptions holds the query parameters for a search request.
type SearchOptions struct {
	Query  string // FTS5 full-text query (3+ runes) or LIKE query (1-2 runes)
	Branch string // git branch filter; empty means all branches
	Since  string // start date filter (RFC3339); empty means no lower bound
	Until  string // end date filter (RFC3339); empty means no upper bound
	Role   string // role filter: "user" or "assistant"; empty means all
	Limit  int    // maximum number of results; 0 uses defaultLimit
}

// SearchResult represents a single search hit.
type SearchResult struct {
	SessionID   string
	Role        string
	Excerpt     string  // excerpt text, up to 200 runes
	Timestamp   string
	GitBranch   string
	ProjectPath string
	Score       float64 // BM25 rank (more negative = more relevant). 0.0 for LIKE fallback.
}

// defaultLimit is the maximum number of results returned when Limit is not specified.
const defaultLimit = 20

// fts5MinQueryRunes is the minimum query length (Unicode rune count) required by the
// FTS5 trigram tokenizer. Queries shorter than this (including 2-rune CJK words) use
// the LIKE fallback.
const fts5MinQueryRunes = 3

// Search searches indexed sessions for messages matching opts.Query.
//
// Search strategy depends on query length:
//   - 3+ runes: FTS5 MATCH (BM25 rank-ordered, high performance)
//   - 1-2 runes: LIKE '%query%' fallback (supports short CJK terms like "인증", "認証")
//
// Returns an empty (non-nil) slice when no results are found.
func Search(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}

	queryRunes := []rune(opts.Query)
	if len(queryRunes) >= fts5MinQueryRunes {
		return searchFTS5(db, opts)
	}
	return searchLike(db, opts)
}

// searchFTS5 runs a full-text search using FTS5 MATCH with BM25 ranking.
// Requires a query of at least fts5MinQueryRunes Unicode characters (trigram tokenizer constraint).
func searchFTS5(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, "messages MATCH ?")
	args = append(args, opts.Query)

	appendCommonFilters(&conditions, &args, opts)
	args = append(args, opts.Limit)

	query := fmt.Sprintf(`
		SELECT m.session_id, m.role,
		       snippet(messages, 3, '', '', '...', 64) AS excerpt,
		       m.timestamp, s.git_branch, s.project_path,
		       rank AS score
		FROM messages m
		JOIN sessions s ON m.session_id = s.session_id
		WHERE %s
		ORDER BY rank
		LIMIT ?`,
		strings.Join(conditions, " AND "),
	)

	return runSearchQuery(db, query, args)
}

// searchLike performs a LIKE-based pattern search.
// Handles 1-2 rune CJK queries (e.g., Korean "인증", Japanese "認証", Chinese "认证")
// that the FTS5 trigram tokenizer cannot handle. Score is always 0.0.
func searchLike(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	var conditions []string
	var args []any

	// Bind as a LIKE parameter; SQLite evaluates '%' || ? || '%' server-side.
	conditions = append(conditions, "m.text LIKE '%' || ? || '%'")
	args = append(args, opts.Query)

	appendCommonFilters(&conditions, &args, opts)
	args = append(args, opts.Limit)

	query := fmt.Sprintf(`
		SELECT m.session_id, m.role,
		       m.text AS excerpt,
		       m.timestamp, s.git_branch, s.project_path,
		       0.0 AS score
		FROM messages m
		JOIN sessions s ON m.session_id = s.session_id
		WHERE %s
		ORDER BY m.timestamp DESC
		LIMIT ?`,
		strings.Join(conditions, " AND "),
	)

	return runSearchQuery(db, query, args)
}

// appendCommonFilters appends branch, role, and date range filters to the condition slice.
func appendCommonFilters(conditions *[]string, args *[]any, opts SearchOptions) {
	if opts.Branch != "" {
		*conditions = append(*conditions, "s.git_branch = ?")
		*args = append(*args, opts.Branch)
	}
	if opts.Role != "" {
		*conditions = append(*conditions, "m.role = ?")
		*args = append(*args, opts.Role)
	}
	if opts.Since != "" {
		*conditions = append(*conditions, "m.timestamp >= ?")
		*args = append(*args, opts.Since)
	}
	if opts.Until != "" {
		*conditions = append(*conditions, "m.timestamp <= ?")
		*args = append(*args, opts.Until)
	}
}

// runSearchQuery executes a SQL query and returns the matching SearchResult slice.
func runSearchQuery(db *sql.DB, query string, args []any) ([]SearchResult, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	results := []SearchResult{}
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(
			&r.SessionID, &r.Role, &r.Excerpt,
			&r.Timestamp, &r.GitBranch, &r.ProjectPath,
			&r.Score,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}
		// Truncate excerpt to 200 runes (rune-safe for multi-byte CJK characters).
		if runes := []rune(r.Excerpt); len(runes) > 200 {
			r.Excerpt = string(runes[:200])
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("result iteration failed: %w", err)
	}

	return results, nil
}
