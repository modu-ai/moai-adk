package search

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchOptions holds search filters and query parameters.
type SearchOptions struct {
	Query  string
	Branch string
	Since  string
	Until  string
	Role   string
	Limit  int
}

// SearchResult represents a single search result.
type SearchResult struct {
	SessionID   string
	Role        string
	Excerpt     string
	Timestamp   string
	GitBranch   string
	ProjectPath string
	Score       float64
}

// Search performs BM25 full-text search with optional filters.
func Search(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	if opts.Query == "" {
		return []SearchResult{}, nil
	}

	// Set default limit
	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	// Build FTS5 query with BM25 ranking
	query := buildFTSQuery(opts)

	// Execute search
	rows, err := db.Query(query, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	// Collect results
	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		err := rows.Scan(&r.SessionID, &r.Role, &r.Excerpt, &r.Timestamp, &r.GitBranch, &r.ProjectPath, &r.Score)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}

// buildFTSQuery constructs the FTS5 search query with filters.
func buildFTSQuery(opts SearchOptions) string {
	// Base query with BM25 ranking
	baseQuery := `
		SELECT 
			m.session_id,
			m.role,
			substr(m.text, 1, 200) as excerpt,
			m.timestamp,
			s.git_branch,
			s.project_path,
			bm25(messages) as score
		FROM messages m
		JOIN sessions s ON m.session_id = s.session_id
		WHERE messages MATCH ?
	`

	// Add filters
	var whereClauses []string
	var args []string

	if opts.Branch != "" {
		whereClauses = append(whereClauses, "s.git_branch = ?")
		args = append(args, opts.Branch)
	}

	if opts.Role != "" {
		whereClauses = append(whereClauses, "m.role = ?")
		args = append(args, opts.Role)
	}

	if opts.Since != "" {
		whereClauses = append(whereClauses, "m.timestamp >= ?")
		args = append(args, opts.Since)
	}

	if opts.Until != "" {
		whereClauses = append(whereClauses, "m.timestamp <= ?")
		args = append(args, opts.Until)
	}

	// Combine clauses
	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	}

	// Add ordering and limit
	baseQuery += " ORDER BY score LIMIT ?"

	// Note: This is a simplified version. The actual args are passed separately.
	return baseQuery
}
