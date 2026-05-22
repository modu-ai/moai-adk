// Package harness — HRN-003 Rubric struct, Markdown parser, and citation validation.
// REQ-HRN-003-003: 4 anchor levels (0.25, 0.50, 0.75, 1.00) FROZEN.
// REQ-HRN-003-005: .md profile-file parser.
// REQ-HRN-003-009: enforces rubric-anchor citation.
package harness

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// @MX:NOTE: [AUTO] FROZEN at {0.25, 0.50, 0.75, 1.00} per design-constitution §12 Mechanism 1; adding anchors requires CON-002 amendment
// @MX:WARN: [AUTO] FROZEN-zone constraint per CONST-V3R2-155; anchor set must remain {0.25, 0.50, 0.75, 1.00}
// @MX:REASON: design-constitution §12 Mechanism 1 — "Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0" (FROZEN)

// canonicalAnchors is the canonical set of rubric anchors (FROZEN).
// REQ-HRN-003-003, REQ-HRN-003-013.
var canonicalAnchors = map[float64]bool{
	0.25: true,
	0.50: true,
	0.75: true,
	1.00: true,
}

// canonicalAnchorStrings is the canonical string set of rubric anchors (FROZEN).
// REQ-HRN-003-009: used by ValidateCitation.
var canonicalAnchorStrings = map[string]bool{
	"0.25": true,
	"0.50": true,
	"0.75": true,
	"1.00": true,
}

// DimensionRubric is the rubric configuration for a single Dimension.
type DimensionRubric struct {
	// Weight is the weight of this dimension (0.0~1.0).
	Weight float64
	// PassThreshold is the minimum pass threshold for this dimension (>= 0.60 FROZEN floor).
	PassThreshold float64
	// Anchors is the map of anchor score -> description.
	// Canonical set: {0.25, 0.50, 0.75, 1.00}.
	Anchors map[float64]string
	// Aggregation is the aggregation mode for this dimension ("min" or "mean").
	// An empty value inherits the profile-level Rubric.Aggregation.
	Aggregation string
}

// Rubric is the scoring-criteria struct of an evaluator profile.
// REQ-HRN-003-003: includes 4 anchor levels.
// REQ-HRN-003-005: parsed from a .md file.
type Rubric struct {
	// ProfileName is the profile name (for example "default", "strict").
	ProfileName string
	// Dimensions is the Dimension -> DimensionRubric map.
	// Contains exactly four dimensions (Functionality, Security, Craft, Consistency).
	Dimensions map[Dimension]DimensionRubric
	// PassThreshold is the overall pass threshold (>= 0.60 FROZEN floor).
	PassThreshold float64
	// MustPass is the must-pass dimension list.
	// [Security] floor — narrower sets are not permitted (REQ-HRN-003-018).
	MustPass []Dimension
	// Aggregation is the default aggregation mode ("min" or "mean").
	Aggregation string
}

// Validate verifies the Rubric.
// All of the following conditions must hold:
//   - exactly 4 canonical dimensions (REQ-HRN-003-012)
//   - exactly 4 canonical anchors per dimension (REQ-HRN-003-013)
//   - pass_threshold >= 0.60 (REQ-HRN-003-014)
//   - aggregation is one of {"min", "mean"} (REQ-HRN-003-007)
//   - MustPass includes the [Security] floor (REQ-HRN-003-018)
func (r *Rubric) Validate() error {
	var errs []config.ValidationError

	// Must have exactly 4 canonical dimensions.
	if len(r.Dimensions) != 4 {
		errs = append(errs, config.ValidationError{
			Field:   "Dimensions",
			Message: fmt.Sprintf("must have exactly 4 dimensions, got %d (FROZEN per SPEC-V3R2-HRN-003 REQ-001)", len(r.Dimensions)),
			Value:   len(r.Dimensions),
			Wrapped: config.ErrInvalidConfig,
		})
	}

	// Verify that each dimension is canonical.
	for dim, dr := range r.Dimensions {
		if !dim.IsValid() {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v]", dim),
				Message: fmt.Sprintf("unknown dimension %v; only {Functionality, Security, Craft, Consistency} are valid (FROZEN per REQ-HRN-003-001)", dim),
				Value:   dim,
				Wrapped: config.ErrUnknownDimension,
			})
			continue
		}

		// Each dimension must have exactly 4 canonical anchors.
		if err := validateAnchors(dim, dr.Anchors); err != nil {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v].Anchors", dim),
				Message: err.Error(),
				Wrapped: config.ErrInvalidConfig,
			})
		}

		// Validate per-dimension pass_threshold >= 0.60.
		if dr.PassThreshold < 0.60 {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v].PassThreshold", dim),
				Message: fmt.Sprintf("pass_threshold %.2f is below FROZEN floor 0.60 (SPEC-V3R2-HRN-003 REQ-014)", dr.PassThreshold),
				Value:   dr.PassThreshold,
				Wrapped: config.ErrInvalidConfig,
			})
		}

		// Validate per-dimension aggregation.
		if dr.Aggregation != "" && dr.Aggregation != "min" && dr.Aggregation != "mean" {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v].Aggregation", dim),
				Message: fmt.Sprintf("aggregation %q is not valid; must be one of {min, mean}", dr.Aggregation),
				Value:   dr.Aggregation,
				Wrapped: config.ErrInvalidConfig,
			})
		}
	}

	// Validate overall pass_threshold >= 0.60.
	if r.PassThreshold < 0.60 {
		errs = append(errs, config.ValidationError{
			Field:   "PassThreshold",
			Message: fmt.Sprintf("pass_threshold %.2f is below FROZEN floor 0.60 (design-constitution §5, HRN-002 REQ-012)", r.PassThreshold),
			Value:   r.PassThreshold,
			Wrapped: config.ErrInvalidConfig,
		})
	}

	// Validate aggregation.
	if r.Aggregation != "min" && r.Aggregation != "mean" {
		errs = append(errs, config.ValidationError{
			Field:   "Aggregation",
			Message: fmt.Sprintf("aggregation %q is not valid; must be one of {min, mean}", r.Aggregation),
			Value:   r.Aggregation,
			Wrapped: config.ErrInvalidConfig,
		})
	}

	// Verify that MustPass includes the [Security] floor.
	// REQ-HRN-003-018: [Security] is the floor; narrower sets are not permitted.
	if err := validateMustPassFloor(r.MustPass); err != nil {
		errs = append(errs, config.ValidationError{
			Field:   "MustPass",
			Message: err.Error(),
			Wrapped: config.ErrMustPassBypassProhibited,
		})
	}

	if len(errs) > 0 {
		return &config.ValidationErrors{Errors: errs}
	}
	return nil
}

// validateAnchors verifies that the anchor set is exactly the canonical {0.25, 0.50, 0.75, 1.00}.
func validateAnchors(dim Dimension, anchors map[float64]string) error {
	if len(anchors) != 4 {
		return fmt.Errorf("dimension %v must have exactly 4 anchor levels, got %d (FROZEN per REQ-HRN-003-013)", dim, len(anchors))
	}
	for score := range anchors {
		if !canonicalAnchors[score] {
			return fmt.Errorf("dimension %v has non-canonical anchor %.2f; must be one of {0.25, 0.50, 0.75, 1.00}", dim, score)
		}
	}
	return nil
}

// validateMustPassFloor verifies that the MustPass set includes the [Security] floor.
// REQ-HRN-003-018: profiles MAY widen but MAY NOT narrow below [Security].
func validateMustPassFloor(mustPass []Dimension) error {
	hasSecurityFloor := false
	for _, d := range mustPass {
		if d == Security {
			hasSecurityFloor = true
			break
		}
	}
	if !hasSecurityFloor {
		return fmt.Errorf("MustPass set must include Security (floor per OQ3 default + design-constitution §12 Mechanism 3 FROZEN); attempt to narrow below [Security] is prohibited")
	}
	return nil
}

// ValidateCitation verifies that the SubCriterionScore.RubricAnchor field is one of the canonical anchor values.
// REQ-HRN-003-009: empty or non-canonical anchor returns ErrRubricCitationMissing.
// AC-HRN-003-05.a: empty field -> ErrRubricCitationMissing.
// AC-HRN-003-05.b: non-canonical value (e.g. "0.65") -> ErrRubricCitationMissing.
// AC-HRN-003-05.c: canonical value -> nil.
func (r *Rubric) ValidateCitation(score SubCriterionScore) error {
	if score.RubricAnchor == "" {
		return fmt.Errorf("%w: sub-criterion score missing rubric_anchor field (per design-constitution §12 Mechanism 1)", config.ErrRubricCitationMissing)
	}
	if !canonicalAnchorStrings[score.RubricAnchor] {
		return fmt.Errorf("%w: rubric_anchor %q is not a canonical anchor value; must be one of {\"0.25\", \"0.50\", \"0.75\", \"1.00\"}", config.ErrRubricCitationMissing, score.RubricAnchor)
	}
	return nil
}

// ─────────────────────────────────────────────
// Markdown profile parser
// ─────────────────────────────────────────────

// dimensionNameMap maps the dimension names from the profile file to the Dimension enum.
var dimensionNameMap = map[string]Dimension{
	"Functionality": Functionality,
	"Security":      Security,
	"Craft":         Craft,
	"Consistency":   Consistency,
}

// ParseRubricMarkdown parses an evaluator profile file in .md format and returns a Rubric.
// REQ-HRN-003-005: consumes the .moai/config/evaluator-profiles/{name}.md file.
// Parser structure:
//   - H2 "## Evaluation Dimensions" table -> Weight + PassThreshold per dim
//   - H2 "## Must-Pass Criteria" -> MustPass slice
//   - H2 "## Scoring Rubric" -> H3 per dimension -> 2-column score table -> anchor map
//
// Tolerant: allows extra whitespace; normalizes anchor scores ("1.0" -> "1.00") before float64 parsing.
func ParseRubricMarkdown(path string) (*Rubric, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ParseRubricMarkdown: open %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	// Profile name is extracted from the filename (without extension).
	profileName := extractProfileName(path)

	rubric := &Rubric{
		ProfileName:   profileName,
		PassThreshold: 0.60, // Default (FROZEN floor).
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security}, // Default.
		Dimensions:    make(map[Dimension]DimensionRubric),
	}

	// Parser state machine.
	type parseSection int
	const (
		sectionNone parseSection = iota
		sectionEvalDimensions
		sectionMustPass
		sectionScoringRubric
	)

	section := sectionNone
	currentRubricDim := Dimension(0)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")

		// Detect H2 section.
		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimPrefix(line, "## ")
			switch {
			case strings.Contains(heading, "Evaluation Dimensions"):
				section = sectionEvalDimensions
				currentRubricDim = 0
			case strings.Contains(heading, "Must-Pass Criteria"):
				section = sectionMustPass
				currentRubricDim = 0
			case strings.Contains(heading, "Scoring Rubric"):
				section = sectionScoringRubric
				currentRubricDim = 0
			default:
				section = sectionNone
				currentRubricDim = 0
			}
			continue
		}

		// Detect H3 section (per-dimension sections inside Scoring Rubric).
		if strings.HasPrefix(line, "### ") && section == sectionScoringRubric {
			heading := strings.TrimPrefix(line, "### ")
			// Extract the dimension name (for example "Functionality (40%)" -> "Functionality").
			dimName := strings.Fields(heading)[0]
			if dim, ok := dimensionNameMap[dimName]; ok {
				currentRubricDim = dim
				// Initialize the dimension when it does not yet exist.
				if _, exists := rubric.Dimensions[dim]; !exists {
					rubric.Dimensions[dim] = DimensionRubric{
						Weight:        0.25, // Default.
						PassThreshold: 0.60, // FROZEN floor.
						Anchors:       make(map[float64]string),
					}
				}
			} else {
				// Skip unknown dimension names inside Scoring Rubric.
				// Unknown dimensions in the Evaluation Dimensions table are validated by Validate().
				// frontend.md has sections such as "Craft & Functionality", "Design Quality", and "Originality";
				// the parser ignores them and maps UI-related content to Craft.
				currentRubricDim = 0
			}
			continue
		}

		// Parse table rows.
		if !strings.HasPrefix(line, "|") {
			continue
		}

		cells := parseTableRow(line)
		if len(cells) < 2 {
			continue
		}

		switch section {
		case sectionEvalDimensions:
			// "| Dimension | Weight | Pass Threshold |" format.
			if len(cells) < 3 {
				continue
			}
			dimName := cells[0]
			dim, ok := dimensionNameMap[dimName]
			if !ok {
				// Skip header rows and separator lines.
				continue
			}
			// Parse Weight (for example "40%" -> 0.40).
			weight := parsePercentOrFloat(cells[1])
			// PassThreshold is stored as text — parsed only when a number is present.
			passThreshold := 0.60 // Default.
			if pct := parsePercentOrFloat(cells[2]); pct > 0 {
				passThreshold = pct
			}

			dr := rubric.Dimensions[dim]
			if dr.Anchors == nil {
				dr.Anchors = make(map[float64]string)
			}
			dr.Weight = weight
			dr.PassThreshold = passThreshold
			rubric.Dimensions[dim] = dr

		case sectionMustPass:
			// "- Functionality: ..." format (list item).
			// Not a table row, so handled by the listItem parsing below.

		case sectionScoringRubric:
			// "| Score | Description |" format.
			if currentRubricDim == 0 {
				continue
			}
			scoreStr := cells[0]
			description := cells[1]
			if scoreStr == "Score" || strings.Contains(scoreStr, "---") {
				continue // Header or separator.
			}
			// Parse anchor score ("1.0" -> 1.00, "0.25" -> 0.25).
			anchor, err := parseAnchorScore(scoreStr)
			if err != nil {
				// Skip rows that cannot be parsed.
				continue
			}
			dr := rubric.Dimensions[currentRubricDim]
			if dr.Anchors == nil {
				dr.Anchors = make(map[float64]string)
			}
			dr.Anchors[anchor] = description
			rubric.Dimensions[currentRubricDim] = dr
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ParseRubricMarkdown: scan %q: %w", path, err)
	}

	// Re-parse Must-Pass Criteria (list format).
	// Reopen the file to parse the Must-Pass section.
	mustPass, err := parseMustPassSection(path)
	if err != nil {
		return nil, err
	}
	if len(mustPass) > 0 {
		rubric.MustPass = mustPass
	}

	// Initialize default Anchors for dimensions with empty Anchors.
	// (Dimensions created during EvalDimensions parsing but lacking a Rubric table.)
	for dim, dr := range rubric.Dimensions {
		if dr.Anchors == nil {
			dr.Anchors = make(map[float64]string)
			rubric.Dimensions[dim] = dr
		}
	}

	// For the strict.md profile, PassThreshold is set higher.
	// Rubric.Validate() guarantees the 0.60 floor, so we only guard against 0 here.
	if rubric.PassThreshold <= 0 {
		rubric.PassThreshold = 0.60
	}

	return rubric, nil
}

// parseMustPassSection parses the dimension list from the "## Must-Pass Criteria" section of the file.
func parseMustPassSection(path string) ([]Dimension, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parseMustPassSection: open %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var result []Dimension
	inSection := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")

		if strings.HasPrefix(line, "## ") {
			if strings.Contains(line, "Must-Pass Criteria") {
				inSection = true
			} else {
				inSection = false
			}
			continue
		}

		if !inSection {
			continue
		}

		// List items in the "- Functionality: ..." or "- Security: ..." format.
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			content := strings.TrimLeft(line, "-* \t")
			// Check whether the first word is a dimension name.
			parts := strings.SplitN(content, ":", 2)
			dimName := strings.TrimSpace(parts[0])
			if dim, ok := dimensionNameMap[dimName]; ok {
				// Prevent duplicate additions.
				found := false
				for _, d := range result {
					if d == dim {
						found = true
						break
					}
				}
				if !found {
					result = append(result, dim)
				}
			}
		}
	}

	return result, scanner.Err()
}

// parseTableRow parses a Markdown table row and returns a slice of cell contents.
// Strips the leading/trailing pipes (|) and trims each cell.
func parseTableRow(line string) []string {
	// Remove leading/trailing pipes.
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

// parsePercentOrFloat converts a string in "40%" or "0.40" format to float64.
func parsePercentOrFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return 0
		}
		return v / 100.0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseAnchorScore converts an anchor-score string to float64.
// Includes normalization such as "1.0" -> 1.00 and "0.25" -> 0.25.
func parseAnchorScore(s string) (float64, error) {
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	// Normalization: convert to float64; canonical-value validation is performed by Validate().
	return v, nil
}

// extractProfileName extracts the profile name from the file path.
// Example: ".moai/config/evaluator-profiles/default.md" -> "default".
func extractProfileName(path string) string {
	// Extract the filename after the final slash.
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	// Strip the extension.
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		return filename[:idx]
	}
	return filename
}

// Unwrap is needed for errors.Is to work with ErrUnknownDimension wrapped inside ValidationError.
var _ error = (*config.ValidationErrors)(nil)
