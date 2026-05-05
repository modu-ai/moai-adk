package cli

import (
"fmt"
"os"
"path/filepath"
"strings"

"github.com/spf13/cobra"
"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecViewCmd 'moai spec view' subcommand creates.
func newSpecViewCmd() *cobra.Command {
var shapeTrace bool

cmd := &cobra.Command{
Use: "view <SPEC-ID>",
Short: "View acceptance criteria in tree structure",
Long: `Display acceptance criteria from a SPEC document in a hierarchical tree format.

Examples:
moai spec view SPEC-SPC-001
moai spec view SPEC-SPC-001 --shape-trace`,
RunE: func(cmd *cobra.Command, args []string) error {
if len(args) < 1 {
return cmd.Help()
}

specID := args[0]
return viewAcceptanceCriteria(cmd, specID, shapeTrace)
},
}

cmd.Flags().BoolVar(&shapeTrace, "shape-trace", false, "Include node depth and parent ID in output")

return cmd
}

// viewAcceptanceCriteria reads Acceptance Criteria from SPEC document and outputs in tree structure.
func viewAcceptanceCriteria(cmd *cobra.Command, specID string, shapeTrace bool) error {
// find project root
projectRoot, err := findProjectRootFn()
if err != nil {
return fmt.Errorf("failed to find project root: %w", err)
}

// locate SPEC directory
specDir := filepath.Join(projectRoot, ".moai", "specs", specID)
specPath := filepath.Join(specDir, "spec.md")

// verify spec.md file exists
if _, err := os.Stat(specPath); os.IsNotExist(err) {
return fmt.Errorf("spec.md not found for %s at %s", specID, specPath)
}

// read file
content, err := os.ReadFile(specPath)
if err != nil {
return fmt.Errorf("failed to read spec.md: %w", err)
}

// Parse Acceptance Criteria
criteria, parseErrors := spec.ParseAcceptanceCriteria(string(content), false)

if len(parseErrors) > 0 {
// output warnings but continue
for _, err := range parseErrors {
switch e := err.(type) {
case *spec.DanglingRequirementReference:
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
case *spec.MissingRequirementMapping:
// handle only missing REQ mapping warnings for leaf nodes
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
default:
// other errors are fatal
return fmt.Errorf("parse error: %w", err)
}
}
}

// output tree
if len(criteria) == 0 {
_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No acceptance criteria found in %s\n", specID)
return nil
}

_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Acceptance Criteria for %s:\n\n", specID)
printTree(cmd, criteria, "", shapeTrace, 0, "")

return nil
}

// printTree recursively outputs Acceptance Criteria in tree format.
func printTree(cmd *cobra.Command, criteria []spec.Acceptance, prefix string, shapeTrace bool, depth int, parentID string) {
for i, ac := range criteria {
// determine tree glyph
var glyph string
var childPrefix string

if i == len(criteria)-1 {
glyph = "└── "
childPrefix = prefix + " "
} else {
glyph = "├── "
childPrefix = prefix + "│ "
}

// output basic information
line := prefix + glyph + formatAcceptanceNode(ac, shapeTrace, depth, parentID)
_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)

// recursively output child nodes
if len(ac.Children) > 0 {
printTree(cmd, ac.Children, childPrefix, shapeTrace, depth+1, ac.ID)
}
}
}

// formatAcceptanceNode formats a single Acceptance Criteria node.
func formatAcceptanceNode(ac spec.Acceptance, shapeTrace bool, depth int, parentID string) string {
var parts []string

// ID
parts = append(parts, ac.ID)

// Given/When/Then
if ac.Given != "" {
parts = append(parts, ac.Given)
}
if ac.When != "" {
parts = append(parts, ac.When)
}
if ac.Then != "" {
parts = append(parts, ac.Then)
}

// REQ mapping
if len(ac.RequirementIDs) > 0 {
reqList := make([]string, len(ac.RequirementIDs))
for i, reqID := range ac.RequirementIDs {
reqList[i] = "REQ-" + reqID
}
parts = append(parts, fmt.Sprintf("(maps %s)", strings.Join(reqList, ", ")))
}

// Shape trace information
if shapeTrace {
traceInfo := fmt.Sprintf("[depth:%d", depth)
if parentID != "" {
traceInfo += fmt.Sprintf(", parent:%s", parentID)
}
traceInfo += "]"
parts = append(parts, traceInfo)
}

return strings.Join(parts, ": ")
}
