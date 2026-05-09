package cli

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"time"

"github.com/spf13/cobra"

"github.com/modu-ai/moai-adk/internal/mx"
)

// newMxQueryCmd 'moai mx query' subcommand creates.
// Supports all filter flags defined in SPEC-V3R2-SPC-004 REQ-SPC-004-001.
func newMxQueryCmd() *cobra.Command {
var specID string
var kind string
var fanInMin int
var danger string
var filePrefix string
var since string
var limit int
var offset int
var format string
var includeTests bool

cmd := &cobra.Command{
Use: "query",
Short: "Query @MX tag sidecar index",
Long: `Query @MX tag sidecar index with structured tag format.

Apply multiple filters with AND combination and output in JSON (default), table, or markdown format.

Examples:
moai mx query --spec SPEC-AUTH-001 --kind anchor
moai mx query --fan-in-min 3 --kind anchor
moai mx query --danger concurrency
moai mx query --file-prefix internal/auth/ --format table`,
RunE: func(cmd *cobra.Command, args []string) error {
// verify project root
projectRoot, err := findProjectRootFn()
if err != nil {
return fmt.Errorf("failed to find project root: %w", err)
}

// validate KIND (REQ-SPC-004-041)
if kind != "" {
validKinds := map[string]bool{
"note": true, "warn": true, "anchor": true,
"todo": true, "legacy": true,
"NOTE": true, "WARN": true, "ANCHOR": true,
"TODO": true, "LEGACY": true,
}
if !validKinds[kind] {
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "InvalidQuery: --kind value ' '%s' is invalid. Allowed values: note, warn, anchor, todo, legacy\n", kind)
return &mx.InvalidQueryError{
Field: "kind",
Value: kind,
Message: "Allowed values: note, warn, anchor, todo, legacy",
}
}
}

// parse SINCE
var sinceTime time.Time
if since != "" {
parsed, err := time.Parse(time.RFC3339, since)
if err != nil {
return &mx.InvalidQueryError{
Field: "since",
Value: since,
Message: "RFC3339 format required (e.g.: 2006-01-02T15:04:05Z)",
}
}
sinceTime = parsed
}

// validate FORMAT
validFormats := map[string]bool{"json": true, "table": true, "markdown": true}
if format != "" && !validFormats[format] {
return &mx.InvalidQueryError{
Field: "format",
Value: format,
Message: "Allowed values: json, table, markdown",
}
}
if format == "" {
format = "json"
}

// verify sidecar path
stateDir := filepath.Join(projectRoot, ".moai", "state")
mgr := mx.NewManager(stateDir)

// verify sidecar file exists (REQ-SPC-004-013)
sidecarPath := filepath.Join(stateDir, mx.SidecarFileName)
if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
"SidecarUnavailable: sidecar index does not exist — run '/moai mx --full' to rebuild index\n")
return fmt.Errorf("SidecarUnavailable: no sidecar index")
}

// create Resolver
resolver := mx.NewResolver(mgr)

// convert KIND string to TagKind
var tagKind mx.TagKind
if kind != "" {
switch kind {
case "note", "NOTE":
tagKind = mx.MXNote
case "warn", "WARN":
tagKind = mx.MXWarn
case "anchor", "ANCHOR":
tagKind = mx.MXAnchor
case "todo", "TODO":
tagKind = mx.MXTodo
case "legacy", "LEGACY":
tagKind = mx.MXLegacy
}
}

// execute query
query := mx.Query{
SpecID: specID,
Kind: tagKind,
FanInMin: fanInMin,
Danger: danger,
FilePrefix: filePrefix,
Since: sinceTime,
Limit: limit,
Offset: offset,
IncludeTests: includeTests,
}

result, err := resolver.Resolve(query)
if err != nil {
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
return err
}

// render per output format
switch format {
case "table":
_, _ = fmt.Fprint(cmd.OutOrStdout(), mx.FormatTable(result))

case "markdown":
_, _ = fmt.Fprint(cmd.OutOrStdout(), mx.FormatMarkdown(result))

default: // json
data, err := json.MarshalIndent(result.Tags, "", " ")
if err != nil {
return fmt.Errorf("JSON serialization failed: %w", err)
}

if result.TruncationNotice {
_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
"TruncationNotice: showing only %d of %d total. Use --limit flag to see more.\n",
result.TotalCount, len(result.Tags))
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
}

return nil
},
}

// REQ-SPC-004-001 flag Register
cmd.Flags().StringVar(&specID, "spec", "", "SPEC ID filter (e.g.: SPEC-AUTH-001)")
cmd.Flags().StringVar(&kind, "kind", "", "TAG kind/type filter (note|warn|anchor|todo|legacy)")
cmd.Flags().IntVar(&fanInMin, "fan-in-min", 0, "minimum fan-in count filter (ANCHOR only)")
cmd.Flags().StringVar(&danger, "danger", "", "danger category filter (WARN only)")
cmd.Flags().StringVar(&filePrefix, "file-prefix", "", "file path prefix filter")
cmd.Flags().StringVar(&since, "since", "", "minimum LastSeenAt time filter (RFC3339 format)")
cmd.Flags().IntVar(&limit, "limit", 0, "maximum return count (default 100, max 10000)")
cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset (default 0)")
cmd.Flags().StringVar(&format, "format", "json", "output format (json|table|markdown)")
cmd.Flags().BoolVar(&includeTests, "include-tests", false, "include test file references in fan-in calculation")

return cmd
}
