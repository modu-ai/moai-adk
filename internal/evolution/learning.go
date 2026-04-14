package evolution

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// reLearningID matches the canonical learning ID format: LEARN-YYYYMMDD-NNN.
// 형식 외의 ID는 경로 순회 등 보안 위험이 있으므로 거부한다.
var reLearningID = regexp.MustCompile(`^LEARN-\d{8}-\d{3}$`)

// validateLearningID checks that id matches the required LEARN-YYYYMMDD-NNN format.
// Returns ErrInvalidLearningID if the format is not satisfied.
func validateLearningID(id string) error {
	if !reLearningID.MatchString(id) {
		return ErrInvalidLearningID
	}
	return nil
}

// learningsDir returns the directory where learning files are stored.
func learningsDir(projectRoot string) string {
	return filepath.Join(projectRoot, defs.MoAIDir, defs.EvolutionSubdir, "learnings")
}

// learningFilePath returns the path to the learning file for the given ID.
func learningFilePath(projectRoot, id string) string {
	return filepath.Join(learningsDir(projectRoot), id+".md")
}

// CreateLearning writes a new LearningEntry to disk.
//
// Preconditions:
//   - The entry ID must match LEARN-YYYYMMDD-NNN format (보안: 경로 순회 방지).
//   - The active learning count must be below MaxActiveLearnings.
//   - The entry ID must not already exist.
//
// Returns ErrInvalidLearningID when the ID format is invalid.
// Returns ErrMaxLearningsReached when the cap is hit.
func CreateLearning(projectRoot string, entry *LearningEntry) error {
	// ID 형식 검증: 경로 순회 및 임의 파일 덮어쓰기 방지
	if err := validateLearningID(entry.ID); err != nil {
		return err
	}

	dir := learningsDir(projectRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("evolution: mkdir learnings: %w", err)
	}

	// Count active learnings.
	active, err := countActiveLearnings(projectRoot)
	if err != nil {
		return err
	}
	if active >= MaxActiveLearnings {
		return ErrMaxLearningsReached
	}

	content := marshalLearning(entry)
	path := learningFilePath(projectRoot, entry.ID)
	// Write atomically.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("evolution: write learning temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evolution: rename learning: %w", err)
	}
	return nil
}

// LoadLearningByID reads and parses the learning file identified by id.
// Returns nil (no error) when the file does not exist.
// Returns ErrInvalidLearningID when the ID format is invalid.
func LoadLearningByID(projectRoot, id string) (*LearningEntry, error) {
	if err := validateLearningID(id); err != nil {
		return nil, err
	}
	path := learningFilePath(projectRoot, id)
	return LoadLearning(path)
}

// LoadLearning reads and parses the learning file at path.
// Returns nil (no error) when the file does not exist.
func LoadLearning(path string) (*LearningEntry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("evolution: read learning %s: %w", path, err)
	}
	return unmarshalLearning(string(data))
}

// UpdateLearning applies fn to the learning entry identified by id,
// then persists the modified entry.
// Returns ErrInvalidLearningID when the ID format is invalid.
func UpdateLearning(projectRoot, id string, fn func(*LearningEntry)) error {
	if err := validateLearningID(id); err != nil {
		return err
	}
	path := learningFilePath(projectRoot, id)
	entry, err := LoadLearning(path)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("evolution: learning %q not found", id)
	}
	fn(entry)
	content := marshalLearning(entry)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("evolution: write updated learning: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evolution: rename updated learning: %w", err)
	}
	return nil
}

// ListLearnings returns all learning entries matching the provided filter.
func ListLearnings(projectRoot string, filter LearningFilter) ([]*LearningEntry, error) {
	dir := learningsDir(projectRoot)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("evolution: readdir learnings: %w", err)
	}

	var result []*LearningEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		entry, err := LoadLearning(path)
		if err != nil || entry == nil {
			continue
		}

		if filter.ExcludeArchived && entry.Status == StatusArchived {
			continue
		}
		if filter.Status != "" && entry.Status != filter.Status {
			continue
		}
		if filter.SkillID != "" && entry.SkillID != filter.SkillID {
			continue
		}

		result = append(result, entry)
	}
	return result, nil
}

// ArchiveOldLearnings moves the oldest entries to StatusArchived when the
// active count exceeds maxActive.  Entries are ordered by creation date
// (oldest first for archiving).
func ArchiveOldLearnings(projectRoot string, maxActive int) error {
	entries, err := ListLearnings(projectRoot, LearningFilter{ExcludeArchived: true})
	if err != nil {
		return err
	}
	excess := len(entries) - maxActive
	if excess <= 0 {
		return nil
	}

	// Sort by Created ascending (simple insertion sort for small slices).
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].Created.Before(entries[j-1].Created); j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}

	for i := 0; i < excess; i++ {
		err := UpdateLearning(projectRoot, entries[i].ID, func(e *LearningEntry) {
			e.Status = StatusArchived
			e.Updated = time.Now()
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// countActiveLearnings returns the number of non-archived learning files.
func countActiveLearnings(projectRoot string) (int, error) {
	dir := learningsDir(projectRoot)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("evolution: readdir for count: %w", err)
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		entry, err := LoadLearning(path)
		if err != nil || entry == nil {
			continue
		}
		if entry.Status != StatusArchived {
			count++
		}
	}
	return count, nil
}

// ---- Serialisation --------------------------------------------------------
//
// Learning files are stored as Markdown with a YAML-like header section.
// The format is intentionally simple so it can be parsed with regex/string
// matching without a YAML dependency.
//
// Format:
//
//	# Learning: <ID>
//
//	Status: <status>
//	SkillID: <skill-id>
//	ZoneID: <zone-id>
//	Observations: <n>
//	Confidence: <f>
//	Created: <RFC3339>
//	Updated: <RFC3339>
//
//	## Observation
//
//	<observation text>
//
//	## Evidence
//
//	- SessionID: <id>
//	  Date: <date>
//	  Context: <context>
//
//	## ProposedChange
//
//	TargetFile: <path>
//	ZoneID: <zone>
//	Addition: |
//	  <addition lines>

func marshalLearning(e *LearningEntry) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Learning: %s\n\n", e.ID)
	fmt.Fprintf(&sb, "Status: %s\n", e.Status)
	fmt.Fprintf(&sb, "SkillID: %s\n", e.SkillID)
	fmt.Fprintf(&sb, "ZoneID: %s\n", e.ZoneID)
	fmt.Fprintf(&sb, "Observations: %d\n", e.Observations)
	fmt.Fprintf(&sb, "Confidence: %.4f\n", e.Confidence)
	fmt.Fprintf(&sb, "Created: %s\n", e.Created.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "Updated: %s\n", e.Updated.UTC().Format(time.RFC3339))
	sb.WriteString("\n## Observation\n\n")
	sb.WriteString(e.Observation)
	sb.WriteString("\n")

	if len(e.Evidence) > 0 {
		sb.WriteString("\n## Evidence\n\n")
		for _, ev := range e.Evidence {
			fmt.Fprintf(&sb, "- SessionID: %s\n", ev.SessionID)
			fmt.Fprintf(&sb, "  Date: %s\n", ev.Date)
			fmt.Fprintf(&sb, "  Context: %s\n", strings.ReplaceAll(ev.Context, "\n", " "))
		}
	}

	if e.ProposedChange != nil {
		sb.WriteString("\n## ProposedChange\n\n")
		fmt.Fprintf(&sb, "TargetFile: %s\n", e.ProposedChange.TargetFile)
		fmt.Fprintf(&sb, "ZoneID: %s\n", e.ProposedChange.ZoneID)
		sb.WriteString("Addition: |\n")
		for _, line := range strings.Split(e.ProposedChange.Addition, "\n") {
			fmt.Fprintf(&sb, "  %s\n", line)
		}
	}

	return sb.String()
}

// Header field regexes.
var (
	reHeaderID       = regexp.MustCompile(`^#\s+Learning:\s+(\S+)`)
	reHeaderField    = regexp.MustCompile(`^([A-Za-z]+):\s*(.*)`)
	reEvidenceStart  = regexp.MustCompile(`^-\s+SessionID:\s+(.+)`)
	reEvidenceField  = regexp.MustCompile(`^\s+(Date|Context):\s+(.+)`)
)

func unmarshalLearning(content string) (*LearningEntry, error) {
	var e LearningEntry
	lines := strings.Split(content, "\n")

	section := "header" // "header" | "observation" | "evidence" | "proposed"
	var obsLines []string
	var additionLines []string
	inAddition := false
	var currentEvidence *EvidenceEntry

	for _, line := range lines {
		switch {
		case reHeaderID.MatchString(line):
			m := reHeaderID.FindStringSubmatch(line)
			e.ID = m[1]
			section = "header"

		case strings.TrimSpace(line) == "## Observation":
			if currentEvidence != nil {
				e.Evidence = append(e.Evidence, *currentEvidence)
				currentEvidence = nil
			}
			section = "observation"

		case strings.TrimSpace(line) == "## Evidence":
			section = "evidence"

		case strings.TrimSpace(line) == "## ProposedChange":
			if currentEvidence != nil {
				e.Evidence = append(e.Evidence, *currentEvidence)
				currentEvidence = nil
			}
			section = "proposed"
			if e.ProposedChange == nil {
				e.ProposedChange = &ProposedChange{}
			}
			inAddition = false

		default:
			switch section {
			case "header":
				if m := reHeaderField.FindStringSubmatch(line); m != nil {
					parseHeaderField(&e, m[1], m[2])
				}

			case "observation":
				obsLines = append(obsLines, line)

			case "evidence":
				if m := reEvidenceStart.FindStringSubmatch(line); m != nil {
					if currentEvidence != nil {
						e.Evidence = append(e.Evidence, *currentEvidence)
					}
					currentEvidence = &EvidenceEntry{SessionID: strings.TrimSpace(m[1])}
				} else if currentEvidence != nil {
					if m := reEvidenceField.FindStringSubmatch(line); m != nil {
						switch m[1] {
						case "Date":
							currentEvidence.Date = strings.TrimSpace(m[2])
						case "Context":
							currentEvidence.Context = strings.TrimSpace(m[2])
						}
					}
				}

			case "proposed":
				if e.ProposedChange == nil {
					e.ProposedChange = &ProposedChange{}
				}
				if !inAddition {
					if m := reHeaderField.FindStringSubmatch(line); m != nil {
						switch m[1] {
						case "TargetFile":
							e.ProposedChange.TargetFile = strings.TrimSpace(m[2])
						case "ZoneID":
							e.ProposedChange.ZoneID = strings.TrimSpace(m[2])
						case "Addition":
							inAddition = true
						}
					}
				} else {
					// Strip two-space indent from addition lines.
					if strings.HasPrefix(line, "  ") {
						additionLines = append(additionLines, line[2:])
					} else if line == "" {
						additionLines = append(additionLines, "")
					}
				}
			}
		}
	}

	// Flush pending evidence.
	if currentEvidence != nil {
		e.Evidence = append(e.Evidence, *currentEvidence)
	}

	// Trim leading/trailing blank lines from observation.
	e.Observation = strings.TrimSpace(strings.Join(obsLines, "\n"))

	// Trim trailing blank line from addition.
	if e.ProposedChange != nil && len(additionLines) > 0 {
		e.ProposedChange.Addition = strings.TrimRight(strings.Join(additionLines, "\n"), "\n")
	}

	if e.ID == "" {
		return nil, fmt.Errorf("evolution: malformed learning file (missing ID)")
	}
	return &e, nil
}

func parseHeaderField(e *LearningEntry, key, value string) {
	value = strings.TrimSpace(value)
	switch key {
	case "Status":
		e.Status = Status(value)
	case "SkillID":
		e.SkillID = value
	case "ZoneID":
		e.ZoneID = value
	case "Observations":
		if n, err := strconv.Atoi(value); err == nil {
			e.Observations = n
		}
	case "Confidence":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			e.Confidence = f
		}
	case "Created":
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			e.Created = t
		}
	case "Updated":
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			e.Updated = t
		}
	}
}
