package constitution

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Sentinel error keys for constitution validate command.
// Each constant matches the string key used in ValidationEntry.SentinelKey
// and in CLI output / JSON results.
const (
	// SentinelDrift is used when a registry clause does not match the actual text of the source file.
	SentinelDrift = "DRIFT"

	// SentinelSourceFileMissing is used when the source file referenced by a registry entry does not exist.
	// Exit code 2.
	SentinelSourceFileMissing = "SOURCE_FILE_MISSING"

	// SentinelZoneUnregistered is used when a [HARD] rule is found in a source file but is not present in the registry.
	SentinelZoneUnregistered = "ZONE_UNREGISTERED"

	// SentinelFrozenWithoutCanary is used when a Frozen zone entry has canary_gate:false.
	SentinelFrozenWithoutCanary = "FROZEN_WITHOUT_CANARY"

	// SentinelAnchorNotFound is used when the anchor in a registry entry does not exist in the source file.
	SentinelAnchorNotFound = "ANCHOR_NOT_FOUND"

	// SentinelDuplicateID is used when the same ID is used more than once in the registry. Always exit 1.
	SentinelDuplicateID = "DUPLICATE_ID"

	// SentinelStaleEntry is used when an entry timestamp is older than 90 days (warning only).
	SentinelStaleEntry = "STALE_ENTRY"

	// SentinelDuplicateZoneMarker is used when a single [HARD] rule line contains 2 or more ZONE markers (warning only).
	SentinelDuplicateZoneMarker = "DUPLICATE_ZONE_MARKER"

	// SentinelInvalidZoneClass is used when zone_class is outside the 4-enum allowed values.
	SentinelInvalidZoneClass = "INVALID_ZONE_CLASS"

	// skipValidateEnvKey is the name of the local bypass environment variable.
	skipValidateEnvKey = "MOAI_CONSTITUTION_SKIP_VALIDATE"
)

// validZoneClasses is the set of allowed values for the zone_class 4-enum.
var validZoneClasses = map[string]bool{
	"frozen-canonical":      true,
	"frozen-safety":         true,
	"evolvable-tuning":      true,
	"evolvable-experimental": true,
}

// ValidateStatus represents the overall status of the validate command result.
type ValidateStatus string

const (
	// ValidateStatusOK indicates all validations passed.
	ValidateStatusOK ValidateStatus = "ok"

	// ValidateStatusDrift indicates one or more drifts or errors were found.
	ValidateStatusDrift ValidateStatus = "drift"

	// ValidateStatusSkipped indicates the validation was bypassed via MOAI_CONSTITUTION_SKIP_VALIDATE=1.
	ValidateStatusSkipped ValidateStatus = "skipped"
)

// ValidateOptions holds the options for the Validate function.
type ValidateOptions struct {
	// RegistryPath is the path to the zone-registry.md file.
	RegistryPath string

	// ProjectDir is the base directory used to resolve source file paths.
	ProjectDir string

	// Strict is the --strict flag (does not promote warnings to errors; the legacy
	// meaning is reserved for the future --fail-on-warning behavior).
	Strict bool

	// FailOnWarning promotes warnings to errors (combined with --strict --fail-on-warning).
	FailOnWarning bool
}

// ValidationEntry represents a single validation error or warning item.
type ValidationEntry struct {
	// ID is the related registry entry ID (empty string when absent).
	ID string `json:"id,omitempty"`

	// File is the related source file path.
	File string `json:"file,omitempty"`

	// Anchor is the related anchor.
	Anchor string `json:"anchor,omitempty"`

	// SentinelKey is the error category code (e.g. "DRIFT", "SOURCE_FILE_MISSING").
	SentinelKey string `json:"status"`

	// Detail is the human-readable detailed description.
	Detail string `json:"detail,omitempty"`
}

// ValidationResult holds the overall result of the Validate function.
type ValidationResult struct {
	// Status is the overall status ("ok" | "drift" | "skipped").
	Status ValidateStatus `json:"status"`

	// DriftCount is the number of DRIFT entries.
	DriftCount int `json:"drift_count"`

	// MissingCount is the number of SOURCE_FILE_MISSING entries.
	MissingCount int `json:"missing_count"`

	// UnregisteredCount is the number of ZONE_UNREGISTERED entries.
	UnregisteredCount int `json:"unregistered_count"`

	// Entries is the list of error/warning items.
	Entries []ValidationEntry `json:"entries"`

	// Warnings holds warning-only messages (STALE_ENTRY, DUPLICATE_ZONE_MARKER, etc.).
	Warnings []string `json:"warnings,omitempty"`

	// Skipped indicates whether MOAI_CONSTITUTION_SKIP_VALIDATE bypassed the validation.
	Skipped bool `json:"skipped,omitempty"`
}

// ValidationError is the error type representing a validation failure.
// Used for fatal errors such as SOURCE_FILE_MISSING (exit 2) or DUPLICATE_ID (exit 1).
type ValidationError struct {
	SentinelKey string
	Message     string
	Result      ValidationResult
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("constitution validate: %s: %s", e.SentinelKey, e.Message)
}

// AsValidationError checks whether err is a *ValidationError and, if so, assigns it to target.
func AsValidationError(err error, target **ValidationError) bool {
	if err == nil {
		return false
	}
	if ve, ok := err.(*ValidationError); ok {
		*target = ve
		return true
	}
	return false
}

// zoneMarkerRegexp matches [ZONE:Frozen] or [ZONE:Evolvable] markers.
var zoneMarkerRegexp = regexp.MustCompile(`\[ZONE:(Frozen|Evolvable)\]`)

// multiSpaceRegexp is the pattern used to normalize multiple whitespace
// characters into a single space.
var multiSpaceRegexp = regexp.MustCompile(`\s+`)

// normalizeWhitespace normalizes multiple whitespace in the text into single spaces.
func normalizeWhitespace(s string) string {
	return strings.TrimSpace(multiSpaceRegexp.ReplaceAllString(s, " "))
}

// Validate verifies the consistency between zone-registry and constitution source files.
//
// Return values:
//   - (result, nil): non-fatal errors such as DRIFT, FROZEN_WITHOUT_CANARY, INVALID_ZONE_CLASS
//   - (result, *ValidationError): fatal errors such as SOURCE_FILE_MISSING or DUPLICATE_ID
//   - (result{Skipped:true}, nil): bypassed via MOAI_CONSTITUTION_SKIP_VALIDATE=1 environment variable
func Validate(opts ValidateOptions) (ValidationResult, error) {
	// MOAI_CONSTITUTION_SKIP_VALIDATE=1 bypass check
	if os.Getenv(skipValidateEnvKey) == "1" {
		_, _ = fmt.Fprintf(os.Stderr, "WARN: validation skipped (%s=1)\n", skipValidateEnvKey)
		return ValidationResult{
			Status:  ValidateStatusSkipped,
			Skipped: true,
		}, nil
	}

	reg, err := LoadRegistry(opts.RegistryPath, opts.ProjectDir)
	if err != nil {
		// Check if it's a duplicate ID error from loader
		if strings.Contains(err.Error(), "duplicate ID") {
			result := ValidationResult{
				Status:  ValidateStatusDrift,
				Entries: []ValidationEntry{{SentinelKey: SentinelDuplicateID, Detail: err.Error()}},
			}
			return result, &ValidationError{
				SentinelKey: SentinelDuplicateID,
				Message:     err.Error(),
				Result:      result,
			}
		}
		return ValidationResult{}, fmt.Errorf("load registry: %w", err)
	}

	var result ValidationResult
	result.Status = ValidateStatusOK
	result.Entries = []ValidationEntry{}
	result.Warnings = []string{}

	// Track seen source files to avoid duplicate reads
	sourceCache := make(map[string]string)

	for _, entry := range reg.Entries {
		// 1. INVALID_ZONE_CLASS check
		if entry.ZoneClass != "" && !validZoneClasses[entry.ZoneClass] {
			result.Entries = append(result.Entries, ValidationEntry{
				ID:          entry.ID,
				File:        entry.File,
				SentinelKey: SentinelInvalidZoneClass,
				Detail:      fmt.Sprintf("zone_class %q is not one of the 4 allowed values", entry.ZoneClass),
			})
		}

		// 2. FROZEN_WITHOUT_CANARY check
		if entry.Zone == ZoneFrozen && !entry.CanaryGate {
			result.Entries = append(result.Entries, ValidationEntry{
				ID:          entry.ID,
				File:        entry.File,
				Anchor:      entry.Anchor,
				SentinelKey: SentinelFrozenWithoutCanary,
				Detail:      "Frozen zone entry must have canary_gate: true",
			})
		}

		// 3. SOURCE_FILE_MISSING check
		sourcePath := entry.File
		if !isAbsPath(sourcePath) {
			sourcePath = joinPath(opts.ProjectDir, sourcePath)
		}
		sourceContent, cached := sourceCache[sourcePath]
		if !cached {
			data, readErr := os.ReadFile(sourcePath) // #nosec G304 -- registry-controlled, project-scoped
			if readErr != nil {
				if os.IsNotExist(readErr) {
					result.MissingCount++
					result.Entries = append(result.Entries, ValidationEntry{
						ID:          entry.ID,
						File:        entry.File,
						Anchor:      entry.Anchor,
						SentinelKey: SentinelSourceFileMissing,
						Detail:      fmt.Sprintf("file %q not found", entry.File),
					})
					// Build a ValidationError since SOURCE_FILE_MISSING is fatal (exit 2)
					finalResult := result
					finalResult.Status = ValidateStatusDrift
					return finalResult, &ValidationError{
						SentinelKey: SentinelSourceFileMissing,
						Message:     fmt.Sprintf("file %q not found", entry.File),
						Result:      finalResult,
					}
				}
				return ValidationResult{}, fmt.Errorf("read source %q: %w", entry.File, readErr)
			}
			sourceContent = string(data)
			sourceCache[sourcePath] = sourceContent
		}

		// 4. DUPLICATE_ZONE_MARKER warning check (scan source for this entry's file)
		checkDuplicateZoneMarkers(sourceContent, entry.File, &result)

		// 5. DRIFT check — clause text must appear (as substring after whitespace normalization) in source
		normalizedClause := normalizeWhitespace(entry.Clause)
		normalizedSource := normalizeWhitespace(stripCodeFences(sourceContent))
		if normalizedClause != "" && !strings.Contains(normalizedSource, normalizedClause) {
			result.DriftCount++
			result.Entries = append(result.Entries, ValidationEntry{
				ID:          entry.ID,
				File:        entry.File,
				Anchor:      entry.Anchor,
				SentinelKey: SentinelDrift,
				Detail:      fmt.Sprintf("clause %q not found in source %q", truncate(normalizedClause, 80), entry.File),
			})
		}
	}

	// Finalize status
	if len(result.Entries) > 0 {
		result.Status = ValidateStatusDrift
	} else {
		result.Status = ValidateStatusOK
	}

	// Recount sentinels
	result.DriftCount = 0
	result.MissingCount = 0
	result.UnregisteredCount = 0
	for _, e := range result.Entries {
		switch e.SentinelKey {
		case SentinelDrift:
			result.DriftCount++
		case SentinelSourceFileMissing:
			result.MissingCount++
		case SentinelZoneUnregistered:
			result.UnregisteredCount++
		}
	}

	return result, nil
}

// checkDuplicateZoneMarkers adds a warning when a single line in the source content contains two or more ZONE markers.
func checkDuplicateZoneMarkers(content, filePath string, result *ValidationResult) {
	seen := make(map[string]bool) // avoid duplicate warnings for same file
	for _, line := range strings.Split(content, "\n") {
		matches := zoneMarkerRegexp.FindAllString(line, -1)
		if len(matches) >= 2 {
			key := fmt.Sprintf("%s:%s", filePath, line[:min(len(line), 60)])
			if !seen[key] {
				seen[key] = true
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("%s: line has multiple ZONE markers: %s", SentinelDuplicateZoneMarker, strings.TrimSpace(line[:min(len(line), 80)])))
			}
		}
	}
}

// stripCodeFences removes the content inside markdown code fences (```...```).
// This prevents false positives when [HARD] is used as a code example (EC-CDL-005).
func stripCodeFences(content string) string {
	var result strings.Builder
	inFence := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			result.WriteString("\n") // preserve line count for readability
			continue
		}
		if !inFence {
			result.WriteString(line)
			result.WriteString("\n")
		} else {
			result.WriteString("\n") // blank placeholder to preserve structure
		}
	}
	return result.String()
}

// isAbsPath reports whether the path is an absolute path.
func isAbsPath(p string) bool {
	return len(p) > 0 && p[0] == '/'
}

// joinPath joins two paths (equivalent to filepath.Join but without the import).
func joinPath(base, rel string) string {
	if base == "" {
		return rel
	}
	if strings.HasSuffix(base, "/") {
		return base + rel
	}
	return base + "/" + rel
}

// truncate trims a string to at most maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// min returns the smaller of two ints (defined locally to avoid conflict with the Go 1.21 min builtin).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
