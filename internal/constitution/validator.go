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
	// SentinelDrift는 registry clause 가 source file 의 실제 텍스트와 불일치할 때 사용된다.
	SentinelDrift = "DRIFT"

	// SentinelSourceFileMissing은 registry entry 가 참조하는 source file 이 존재하지 않을 때 사용된다.
	// Exit code 2.
	SentinelSourceFileMissing = "SOURCE_FILE_MISSING"

	// SentinelZoneUnregistered는 source file 에서 [HARD] rule 이 발견되었으나 registry 에 없을 때 사용된다.
	SentinelZoneUnregistered = "ZONE_UNREGISTERED"

	// SentinelFrozenWithoutCanary는 Frozen zone entry 가 canary_gate:false 일 때 사용된다.
	SentinelFrozenWithoutCanary = "FROZEN_WITHOUT_CANARY"

	// SentinelAnchorNotFound는 registry entry 의 anchor 가 source file 내 존재하지 않을 때 사용된다.
	SentinelAnchorNotFound = "ANCHOR_NOT_FOUND"

	// SentinelDuplicateID는 registry 에 동일 ID 가 두 번 이상 사용될 때 사용된다. 항상 exit 1.
	SentinelDuplicateID = "DUPLICATE_ID"

	// SentinelStaleEntry는 entry 타임스탬프가 90일 이상 경과할 때 사용된다 (warning only).
	SentinelStaleEntry = "STALE_ENTRY"

	// SentinelDuplicateZoneMarker는 동일 [HARD] rule 라인에 ZONE 마커가 2개 이상일 때 사용된다 (warning only).
	SentinelDuplicateZoneMarker = "DUPLICATE_ZONE_MARKER"

	// SentinelInvalidZoneClass는 zone_class 가 4-enum 외 값일 때 사용된다.
	SentinelInvalidZoneClass = "INVALID_ZONE_CLASS"

	// skipValidateEnvKey는 로컬 bypass 환경변수 이름.
	skipValidateEnvKey = "MOAI_CONSTITUTION_SKIP_VALIDATE"
)

// validZoneClasses는 zone_class 4-enum 허용 값 집합.
var validZoneClasses = map[string]bool{
	"frozen-canonical":      true,
	"frozen-safety":         true,
	"evolvable-tuning":      true,
	"evolvable-experimental": true,
}

// ValidateStatus는 validate 명령 결과의 전체 상태를 나타낸다.
type ValidateStatus string

const (
	// ValidateStatusOK는 모든 검증이 통과된 상태.
	ValidateStatusOK ValidateStatus = "ok"

	// ValidateStatusDrift는 하나 이상의 drift 또는 오류가 발견된 상태.
	ValidateStatusDrift ValidateStatus = "drift"

	// ValidateStatusSkipped는 MOAI_CONSTITUTION_SKIP_VALIDATE=1 로 우회된 상태.
	ValidateStatusSkipped ValidateStatus = "skipped"
)

// ValidateOptions는 Validate 함수의 옵션을 담는 구조체.
type ValidateOptions struct {
	// RegistryPath는 zone-registry.md 파일 경로.
	RegistryPath string

	// ProjectDir는 source file 경로를 resolve 하는 기준 디렉토리.
	ProjectDir string

	// Strict는 --strict 플래그 (경고를 오류로 승격하지 않음; 기존 의미는 future --fail-on-warning).
	Strict bool

	// FailOnWarning은 경고를 오류로 승격한다 (--strict --fail-on-warning 조합).
	FailOnWarning bool
}

// ValidationEntry는 단일 검증 오류/경고 항목을 나타낸다.
type ValidationEntry struct {
	// ID는 관련 registry entry ID (없으면 빈 문자열).
	ID string `json:"id,omitempty"`

	// File은 관련 source file 경로.
	File string `json:"file,omitempty"`

	// Anchor는 관련 anchor.
	Anchor string `json:"anchor,omitempty"`

	// SentinelKey는 오류 종류 코드 (e.g. "DRIFT", "SOURCE_FILE_MISSING").
	SentinelKey string `json:"status"`

	// Detail은 사람이 읽을 수 있는 상세 설명.
	Detail string `json:"detail,omitempty"`
}

// ValidationResult는 Validate 함수의 전체 결과를 담는다.
type ValidationResult struct {
	// Status는 전체 상태 ("ok" | "drift" | "skipped").
	Status ValidateStatus `json:"status"`

	// DriftCount는 DRIFT 항목 수.
	DriftCount int `json:"drift_count"`

	// MissingCount는 SOURCE_FILE_MISSING 항목 수.
	MissingCount int `json:"missing_count"`

	// UnregisteredCount는 ZONE_UNREGISTERED 항목 수.
	UnregisteredCount int `json:"unregistered_count"`

	// Entries는 오류/경고 항목 목록.
	Entries []ValidationEntry `json:"entries"`

	// Warnings는 warning 전용 메시지 (STALE_ENTRY, DUPLICATE_ZONE_MARKER 등).
	Warnings []string `json:"warnings,omitempty"`

	// Skipped는 MOAI_CONSTITUTION_SKIP_VALIDATE 우회 여부.
	Skipped bool `json:"skipped,omitempty"`
}

// ValidationError는 검증 실패를 나타내는 오류 타입.
// SOURCE_FILE_MISSING (exit 2) 또는 DUPLICATE_ID (exit 1) 같은 치명적 오류에 사용.
type ValidationError struct {
	SentinelKey string
	Message     string
	Result      ValidationResult
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("constitution validate: %s: %s", e.SentinelKey, e.Message)
}

// AsValidationError는 err 가 *ValidationError 인지 확인하고 target 에 대입한다.
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

// hardRuleRegexp는 코드 펜스 외부의 [HARD] 규칙을 매칭한다.
var hardRuleRegexp = regexp.MustCompile(`\[HARD\]`)

// zoneMarkerRegexp는 [ZONE:Frozen] 또는 [ZONE:Evolvable] 마커를 매칭한다.
var zoneMarkerRegexp = regexp.MustCompile(`\[ZONE:(Frozen|Evolvable)\]`)

// multiSpaceRegexp는 복수의 공백을 단일 공백으로 정규화하기 위한 패턴.
var multiSpaceRegexp = regexp.MustCompile(`\s+`)

// normalizeWhitespace는 텍스트의 다중 공백을 단일 공백으로 정규화한다.
func normalizeWhitespace(s string) string {
	return strings.TrimSpace(multiSpaceRegexp.ReplaceAllString(s, " "))
}

// Validate는 zone-registry 와 constitution source file 의 정합성을 검증한다.
//
// 반환값:
//   - (result, nil): DRIFT, FROZEN_WITHOUT_CANARY, INVALID_ZONE_CLASS 등 비치명적 오류
//   - (result, *ValidationError): SOURCE_FILE_MISSING 또는 DUPLICATE_ID 같은 치명적 오류
//   - (result{Skipped:true}, nil): MOAI_CONSTITUTION_SKIP_VALIDATE=1 환경변수 우회
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

// checkDuplicateZoneMarkers는 source content 에서 동일 라인에 ZONE 마커가 2개 이상인 경우 warning 을 추가한다.
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

// stripCodeFences는 마크다운 코드 펜스(```...```) 내부 콘텐츠를 제거한다.
// [HARD] 가 코드 예시로 사용되는 경우 false-positive 를 방지한다 (EC-CDL-005).
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

// isAbsPath는 경로가 절대 경로인지 확인한다.
func isAbsPath(p string) bool {
	return len(p) > 0 && p[0] == '/'
}

// joinPath는 두 경로를 결합한다 (filepath.Join 과 동일하지만 import 없이).
func joinPath(base, rel string) string {
	if base == "" {
		return rel
	}
	if strings.HasSuffix(base, "/") {
		return base + rel
	}
	return base + "/" + rel
}

// truncate는 문자열을 maxLen 길이로 자른다.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// min은 두 int 중 작은 값을 반환한다 (Go 1.21 min builtin 과 충돌 방지를 위해 로컬 정의).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
