package mx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type Query struct {
	// SpecID is the filter for connecting tag to specific SPEC (--spec flag).
	SpecID string

	// Kind is the tag type filter (--kind flag).
	Kind TagKind

	// FanInMin is the minimum fan-in value filter (--fan-in-min flag).
	FanInMin int

	// Danger is the danger category filter (--danger flag).
	// WARN tag's REASON text matches mx.yaml danger_categories: patterns.
	Danger string

	// FilePrefix is the file path prefix filter (--file-prefix flag).
	FilePrefix string

	// Since is the time-based filter on LastSeenAt (--since flag).
	Since time.Time

	// Limit is the maximum number of tags to return (--limit flag, default 100, maximum 10000).
	Limit int

	// Offset is the pagination offset (--offset flag, default 0).
	Offset int

	// IncludeTests determines whether to include test file callers during fan-in calculation (--include-tests flag).
	// Default value is false (exclude test file callers).
	IncludeTests bool

	// fanInCounter is the implementation to use for fan-in calculation. Uses TextualFanInCounter if nil.
	fanInCounter FanInCounter

	dangerMatcher *DangerCategoryMatcher

	// specAssociator is the implementation to use for SPEC connection.
	specAssociator *SpecAssociator

	// projectRoot is the project root path for fan-in calculation.
	projectRoot string
}

// DefaultLimit is the default value for --limit flag (REQ-SPC-004-007).
const DefaultLimit = 100

// MaxLimit is the maximum value for --limit flag (REQ-SPC-004-007).
const MaxLimit = 10000

// TagResult represents the result of a single @MX TAG query.
// Implements the JSON schema defined in REQ-SPC-004-005.
type TagResult struct {
	Kind     TagKind `json:"kind"`
	File     string  `json:"file"`
	Line     int     `json:"line"`
	Body     string  `json:"body"`
	Reason   string  `json:"reason,omitempty"`
	AnchorID string  `json:"anchor_id,omitempty"`

	// CreatedBy represents the entity that created the tag (agent name or "human").
	CreatedBy string `json:"created_by"`

	LastSeenAt time.Time `json:"last_seen_at"`

	// FanIn is the caller count of ANCHOR tags (ANCHOR only).
	FanIn int `json:"fan_in,omitempty"`

	// FanInMethod is the fan-in calculation method: "lsp" or "textual" (ANCHOR only).
	FanInMethod string `json:"fan_in_method,omitempty"`

	// DangerCategory is the danger category of WARN tags (WARN only).
	DangerCategory string `json:"danger_category,omitempty"`

	// SpecAssociations is the list of SPEC IDs connected to the tag.
	SpecAssociations []string `json:"spec_associations"`
}

// QueryResult represents the result of Resolve query.
type QueryResult struct {
	Tags []TagResult

	// TruncationNotice is true when results are truncated by Limit (REQ-SPC-004-021).
	TruncationNotice bool

	TotalCount int
}

// SidecarUnavailableError is returned when the sidecar index cannot be read (REQ-SPC-004-013).
type SidecarUnavailableError struct {
	Cause error
}

// Error implements the error interface.
func (e *SidecarUnavailableError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("SidecarUnavailable: %v -- run '/moai mx --full' to rebuild the index", e.Cause)
	}
	return "SidecarUnavailable: cannot read sidecar index -- run '/moai mx --full' to rebuild the index"
}

// Unwrap returns the underlying error.
func (e *SidecarUnavailableError) Unwrap() error {
	return e.Cause
}

// LSPRequiredError is returned when MOAI_MX_QUERY_STRICT=1 and LSP cannot be used (REQ-SPC-004-030).
type LSPRequiredError struct {
	Language string
}

// Error implements the error interface.
func (e *LSPRequiredError) Error() string {
	return fmt.Sprintf("LSPRequired: LSP server is not running for language '%s' (MOAI_MX_QUERY_STRICT=1)", e.Language)
}

// InvalidQueryError is returned when query parameters are syntactically incorrect (REQ-SPC-004-041).
type InvalidQueryError struct {
	Field   string
	Value   string
	Message string
}

// Error implements the error interface.
func (e *InvalidQueryError) Error() string {
	return fmt.Sprintf("InvalidQuery: field '%s' has invalid value '%s': %s", e.Field, e.Value, e.Message)
}

// validKinds is the list of allowed TagKind values.
var validKinds = map[TagKind]bool{
	MXNote:   true,
	MXWarn:   true,
	MXAnchor: true,
	MXTodo:   true,
	MXLegacy: true,
	"":       true, // string means all kinds
}

// resolveLimit normalizes the Limit value of the query.
// Returns DefaultLimit if 0 or negative, returns MaxLimit if exceeds MaxLimit.
func resolveLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

	// Resolve queries @MX TAG based on Query.
//
// @MX:ANCHOR: [AUTO] Resolve — invariant contract of Query API entry point
// @MX:REASON: fan_in >= 3 -- called by CLI mx_query.go, code map generation tools, evaluator (evaluator-active)
func (r *Resolver) Resolve(query Query) (QueryResult, error) {
	// 1. Validate query (REQ-SPC-004-041)
	if err := validateQuery(query); err != nil {
		return QueryResult{}, err
	}

	sidecarPath := r.manager.sidecarPath
	if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
		return QueryResult{}, &SidecarUnavailableError{Cause: err}
	}

	sidecar, err := r.manager.Load()
	if err != nil {
		return QueryResult{}, &SidecarUnavailableError{Cause: err}
	}

	dangerMatcher := query.dangerMatcher
	if dangerMatcher == nil {
		dangerMatcher = NewDangerCategoryMatcher(DangerCategoryConfig{})
	}

	specAssociator := query.specAssociator
	if specAssociator == nil {
		specAssociator = NewSpecAssociator(map[string][]string{})
	}

	fanInCounter := query.fanInCounter
	if fanInCounter == nil {
		fanInCounter = &TextualFanInCounter{}
	}

	// 4. Verify MOAI_MX_QUERY_STRICT (REQ-SPC-004-030)
	strictMode := os.Getenv("MOAI_MX_QUERY_STRICT") == "1"
	needsFanIn := query.FanInMin > 0
	if strictMode && needsFanIn {
		// LSP is needed in strict mode but textual fallback is always provided
		// LSP availability verification is delegated to fanInCounter implementation
		// Return LSPRequired error (no LSP client)
		return QueryResult{}, &LSPRequiredError{Language: "unknown"}
	}

	// 5. Apply all filters to all tags (AND combination, REQ-SPC-004-042)
	limit := resolveLimit(query.Limit)

	var matched []TagResult
	for _, tag := range sidecar.Tags {
		tagResult, ok := applyFilters(tag, query, dangerMatcher, specAssociator)
		if !ok {
			continue
		}

		// Fan-in calculation (when ANCHOR and FanInMin is configured)s
		if tag.Kind == MXAnchor && needsFanIn {
			count, method, err := fanInCounter.Count(context.Background(), tag, query.projectRoot, !query.IncludeTests)
			if err != nil {
				count = 0
				method = "textual"
			}
			tagResult.FanIn = count
			tagResult.FanInMethod = method

			// Fan-in minimum value filter
			if tagResult.FanIn < query.FanInMin {
				continue
			}
		}

		matched = append(matched, tagResult)
	}

	totalCount := len(matched)

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var paginated []TagResult
	if offset >= len(matched) {
		paginated = []TagResult{}
	} else {
		end := offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		paginated = matched[offset:end]
	}

	// 7. Process empty results (REQ-SPC-004-041: non-nil empty slice)s
	if paginated == nil {
		paginated = []TagResult{}
	}

	truncationNotice := totalCount > limit+offset

	return QueryResult{
		Tags:             paginated,
		TruncationNotice: truncationNotice,
		TotalCount:       totalCount,
	}, nil
}

// validateQuery validates the query parameters (REQ-SPC-004-041).
func validateQuery(query Query) error {
	if query.Kind != "" && !validKinds[query.Kind] {
		return &InvalidQueryError{
			Field:   "kind",
			Value:   string(query.Kind),
			Message: fmt.Sprintf("allowed values: note, warn, anchor, todo, legacy (actual: %s)", query.Kind),
		}
	}
	return nil
}

// Returns TagResult and true if passes filter.
func applyFilters(tag Tag, query Query, dangerMatcher *DangerCategoryMatcher, specAssociator *SpecAssociator) (TagResult, bool) {
	// KIND filter (REQ-SPC-004-001)
	if query.Kind != "" && tag.Kind != query.Kind {
		return TagResult{}, false
	}

	// File prefix filter (REQ-SPC-004-001)
	if query.FilePrefix != "" && !strings.HasPrefix(tag.File, query.FilePrefix) {
		return TagResult{}, false
	}

	// Since time filter (REQ-SPC-004-001)
	if !query.Since.IsZero() && tag.LastSeenAt.Before(query.Since) {
		return TagResult{}, false
	}

	// SPEC connection and filter (REQ-SPC-004-006, REQ-SPC-004-010)
	specAssociations := specAssociator.Associate(tag)
	if query.SpecID != "" {
		found := false
		for _, spec := range specAssociations {
			if spec == query.SpecID {
				found = true
				break
			}
		}
		if !found {
			return TagResult{}, false
		}
	}

	// Danger category filter (REQ-SPC-004-012, WARN only)
	var dangerCategory string
	if query.Danger != "" {
		if tag.Kind != MXWarn {
			return TagResult{}, false
		}
		if !dangerMatcher.Match(tag.Reason, query.Danger) {
			return TagResult{}, false
		}
		dangerCategory = query.Danger
	} else if tag.Kind == MXWarn && tag.Reason != "" {
		dangerCategory = dangerMatcher.CategoryOf(tag.Reason)
	}

	// Construct TagResult (REQ-SPC-004-005)
	if specAssociations == nil {
		specAssociations = []string{}
	}

	return TagResult{
		Kind:             tag.Kind,
		File:             tag.File,
		Line:             tag.Line,
		Body:             tag.Body,
		Reason:           tag.Reason,
		AnchorID:         tag.AnchorID,
		CreatedBy:        tag.CreatedBy,
		LastSeenAt:       tag.LastSeenAt,
		DangerCategory:   dangerCategory,
		SpecAssociations: specAssociations,
	}, true
}

// FormatMarkdown converts QueryResult to markdown table (REQ-SPC-004-031).
// Includes Kind, File, Line, Body, FanIn, Danger, SPECs columns.
// If TruncationNotice exists, warning blockquote is added before the table.
func FormatMarkdown(result QueryResult) string {
	var sb strings.Builder

	if result.TruncationNotice {
		_, _ = fmt.Fprintf(&sb, "> **TruncationNotice**: showing %d of %d total results.\n\n", len(result.Tags), result.TotalCount)
	}

	sb.WriteString("| Kind | File | Line | Body | FanIn | Danger | SPECs |\n")
	sb.WriteString("|------|------|------|------|-------|--------|-------|\n")

	for _, tag := range result.Tags {
		specs := strings.Join(tag.SpecAssociations, ", ")
		fanIn := ""
		if tag.FanIn > 0 || tag.FanInMethod != "" {
			fanIn = fmt.Sprintf("%d (%s)", tag.FanIn, tag.FanInMethod)
		}
		_, _ = fmt.Fprintf(&sb, "| %s | %s | %d | %s | %s | %s | %s |\n",
			tag.Kind, tag.File, tag.Line,
			truncateStr(tag.Body, 40),
			fanIn, tag.DangerCategory, specs)
	}

	return sb.String()
}

// FormatTable converts QueryResult to human-readable text table (REQ-SPC-004-004).
// Includes KIND, FILE, LINE, BODY columns and outputs "(결과 없음)" when empty.
func FormatTable(result QueryResult) string {
	if len(result.Tags) == 0 {
		return "(결과 없음)\n"
	}

	var sb strings.Builder

	if result.TruncationNotice {
		_, _ = fmt.Fprintf(&sb, "TruncationNotice: showing %d of %d total results.\n\n", len(result.Tags), result.TotalCount)
	}

	header := fmt.Sprintf("%-8s %-50s %5s %-40s\n", "KIND", "FILE", "LINE", "BODY")
	separator := strings.Repeat("-", len(header)) + "\n"

	sb.WriteString(header)
	sb.WriteString(separator)

	for _, tag := range result.Tags {
		_, _ = fmt.Fprintf(&sb, "%-8s %-50s %5d %-40s\n",
			tag.Kind,
			truncateStr(tag.File, 50),
			tag.Line,
			truncateStr(tag.Body, 40))
	}

	return sb.String()
}

// truncateStr truncates string with "..." if maxLen is exceeded.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
