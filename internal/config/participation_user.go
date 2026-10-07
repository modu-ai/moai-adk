package config

// participation_user.go — the user-scoped consent reader
// (SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-001, DEC-1).
//
// Consent is a property of the person and their GitHub account, not of a
// repository, so it lives in a file under the user's own moai home directory
// — <moai home>/config/participation.yaml — beside, and never inside,
// config/sections/. The section resolver merges project and user tier files
// by priority, so a consent key that took part in that merge could be decided
// by a cloned repository's project file; this reader never opens one, and a
// tracked project file cannot enable anything by construction (the mutant
// fixture in the reader's tests pins that).
//
// The reader is fail-CLOSED: a missing file, key, or directory, an unreadable
// or unparseable file, or a wrongly typed value yields the zero consent —
// enabled false, asked false, repository empty. There is no error return to
// mishandle: the caller cannot accidentally treat a read failure as consent.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/modu-ai/moai-adk/internal/paths"
	"gopkg.in/yaml.v3"
)

// participationFileName is the consent file's name under <moai home>/config.
const participationFileName = "participation.yaml"

// UserParticipation is the consent state the pipeline reads. Repository is
// the user's optional override of the publication target; it is used only
// when well-formed (see UserParticipationRepository).
type UserParticipation struct {
	Enabled    bool
	Asked      bool
	Repository string
}

// userParticipationYAML is the file's schema: every value nested under the
// single `participation:` key.
type userParticipationYAML struct {
	Participation struct {
		Enabled    *bool  `yaml:"enabled"`
		Asked      *bool  `yaml:"asked"`
		Repository string `yaml:"repository"`
	} `yaml:"participation"`
}

// UserParticipationFilePath returns <moai home>/config/participation.yaml.
func UserParticipationFilePath() (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", fmt.Errorf("resolve moai home for participation: %w", err)
	}
	return home + string(os.PathSeparator) + "config" + string(os.PathSeparator) + participationFileName, nil
}

// participationRepositoryPattern is the shape a user repository override must
// match to be used: owner/name with repository-name characters.
var participationRepositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// ReadUserParticipation reads the consent state, fail-closed. Any absence,
// unreadability, parse failure, or wrongly typed value reads as no consent.
//
// @MX:ANCHOR: [AUTO] user-scoped consent reader — capture, drain, sender, preview, console, wizard, and update all read consent through it
// @MX:REASON: a second consent path (or a project-tier fallback) would let a tracked repository file enable publication from the user's account (REQ-ANON-001, DEC-1)
// @MX:WARN: [AUTO] fail-closed reader — every failure path must read as no consent
// @MX:REASON: an error surfaced as success here would capture, queue, and publish without consent; the tests pin the absent, malformed, directory, and wrongly-typed arms
func ReadUserParticipation() UserParticipation {
	path, err := UserParticipationFilePath()
	if err != nil {
		return UserParticipation{}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return UserParticipation{}
	}
	// Exactly ONE YAML document is admitted. yaml.Unmarshal decodes only the
	// first document and silently ignores a tail — so `enabled: true` followed
	// by `---` and a damaged document would read as consent with its broken
	// remainder unseen. The fail-closed contract refuses that shape outright:
	// the second decode must reach EOF or the file reads as no consent.
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	var doc userParticipationYAML
	if err := dec.Decode(&doc); err != nil {
		return UserParticipation{}
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return UserParticipation{}
	}
	up := UserParticipation{}
	if doc.Participation.Enabled != nil {
		up.Enabled = *doc.Participation.Enabled
	}
	if doc.Participation.Asked != nil {
		up.Asked = *doc.Participation.Asked
	}
	if participationRepositoryPattern.MatchString(doc.Participation.Repository) {
		up.Repository = doc.Participation.Repository
	}
	return up
}

// UserParticipationRepository returns the publication target: the user-scoped
// override when well-formed, else the compiled default. The project-tier
// feedback.repository is never read here (REQ-ANON-024) — a cloned repository
// must not redirect where the user's account posts.
func UserParticipationRepository() string {
	if repo := ReadUserParticipation().Repository; repo != "" {
		return repo
	}
	return DefaultFeedbackRepository
}
