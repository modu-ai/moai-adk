package foundation

import "fmt"

// RequirementType represents an EARS (Easy Approach to Requirements Syntax) pattern type.
type RequirementType string

const (
	// Ubiquitous represents a requirement that always applies to the system.
	Ubiquitous RequirementType = "ubiquitous"

	// EventDriven represents a trigger-response requirement.
	EventDriven RequirementType = "event_driven"

	// UnwantedBehavior represents a prohibition or unwanted scenario requirement.
	UnwantedBehavior RequirementType = "unwanted_behavior"

	// StateDriven represents a state-conditional requirement.
	StateDriven RequirementType = "state_driven"

	// Optional represents an optional feature requirement.
	Optional RequirementType = "optional"
)

// String returns the string representation of the RequirementType.
func (r RequirementType) String() string {
	return string(r)
}

// IsValid checks whether the RequirementType is one of the defined constants.
func (r RequirementType) IsValid() bool {
	switch r {
	case Ubiquitous, EventDriven, UnwantedBehavior, StateDriven, Optional:
		return true
	}
	return false
}

// AllRequirementTypes returns all valid EARS requirement types.
func AllRequirementTypes() []RequirementType {
	return []RequirementType{
		Ubiquitous,
		EventDriven,
		UnwantedBehavior,
		StateDriven,
		Optional,
	}
}

// EARSTemplate holds the EARS pattern template for a requirement type.
type EARSTemplate struct {
	Type        RequirementType `json:"type"`
	Name        string          `json:"name"`
	Template    string          `json:"template"`
	Description string          `json:"description"`
}

// earsTemplates stores the predefined EARS pattern templates.
var earsTemplates = map[RequirementType]EARSTemplate{
	Ubiquitous: {
		Type:        Ubiquitous,
		Name:        "Ubiquitous",
		Template:    "The <system> shall <response>.",
		Description: "A requirement that always applies to the system without any preconditions.",
	},
	EventDriven: {
		Type:        EventDriven,
		Name:        "Event-Driven",
		Template:    "When <event>, the <system> shall <response>.",
		Description: "A requirement triggered by a specific event or action.",
	},
	UnwantedBehavior: {
		Type:        UnwantedBehavior,
		Name:        "Unwanted Behavior",
		Template:    "If <unwanted condition>, then the <system> shall <response>.",
		Description: "A requirement that handles undesirable or exceptional scenarios.",
	},
	StateDriven: {
		Type:        StateDriven,
		Name:        "State-Driven",
		Template:    "While <state>, the <system> shall <response>.",
		Description: "A requirement conditional on a specific system state being active.",
	},
	Optional: {
		Type:        Optional,
		Name:        "Optional",
		Template:    "Where <feature>, the <system> shall <response>.",
		Description: "A requirement for an optional feature that may or may not be present.",
	},
}

// GetEARSTemplate returns the EARS template for the given requirement type.
func GetEARSTemplate(rt RequirementType) (EARSTemplate, error) {
	tmpl, ok := earsTemplates[rt]
	if !ok {
		return EARSTemplate{}, fmt.Errorf("%w: %s", ErrInvalidRequirementType, rt)
	}
	return tmpl, nil
}

// GetAllEARSTemplates returns all EARS templates in canonical order.
func GetAllEARSTemplates() []EARSTemplate {
	types := AllRequirementTypes()
	templates := make([]EARSTemplate, 0, len(types))
	for _, rt := range types {
		templates = append(templates, earsTemplates[rt])
	}
	return templates
}

// Requirement represents a single EARS requirement with metadata.
type Requirement struct {
	Type               RequirementType `json:"type"`
	ID                 string          `json:"id"`
	Description        string          `json:"description"`
	AcceptanceCriteria []string        `json:"acceptance_criteria,omitempty"`
}

// Format produces EARS-formatted text for the requirement.
// Returns an empty string if the receiver is nil.
func (r *Requirement) Format() string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("[%s] %s (Type: %s)", r.ID, r.Description, r.Type)
}

// Validate checks that the requirement has all required fields populated and valid.
func (r *Requirement) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: nil requirement", ErrInvalidRequirementType)
	}
	if r.ID == "" {
		return fmt.Errorf("requirement ID cannot be empty")
	}
	if r.Description == "" {
		return fmt.Errorf("requirement description cannot be empty")
	}
	if !r.Type.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidRequirementType, r.Type)
	}
	return nil
}

// RequirementSet is an ordered collection of EARS requirements with index-based lookup.
type RequirementSet struct {
	requirements []*Requirement
	index        map[string]int
}

// NewRequirementSet creates a new empty RequirementSet.
func NewRequirementSet() *RequirementSet {
	return &RequirementSet{
		requirements: make([]*Requirement, 0),
		index:        make(map[string]int),
	}
}

// Add adds a requirement to the set.
// Returns an error if the requirement is nil, invalid, or has a duplicate ID.
func (rs *RequirementSet) Add(r *Requirement) error {
	if r == nil {
		return fmt.Errorf("cannot add nil requirement")
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if _, exists := rs.index[r.ID]; exists {
		return fmt.Errorf("requirement with ID %q already exists", r.ID)
	}
	rs.index[r.ID] = len(rs.requirements)
	rs.requirements = append(rs.requirements, r)
	return nil
}

// Get retrieves a requirement by ID.
// Returns a RequirementNotFoundError if the ID is not found.
func (rs *RequirementSet) Get(id string) (*Requirement, error) {
	if id == "" {
		return nil, &RequirementNotFoundError{ID: id}
	}
	idx, ok := rs.index[id]
	if !ok {
		return nil, &RequirementNotFoundError{ID: id}
	}
	return rs.requirements[idx], nil
}

// Filter returns all requirements matching the given type.
// Returns an empty (non-nil) slice if no requirements match.
func (rs *RequirementSet) Filter(rt RequirementType) []*Requirement {
	var result []*Requirement
	for _, r := range rs.requirements {
		if r.Type == rt {
			result = append(result, r)
		}
	}
	if result == nil {
		return []*Requirement{}
	}
	return result
}

// All returns a copy of all requirements in the set.
func (rs *RequirementSet) All() []*Requirement {
	result := make([]*Requirement, len(rs.requirements))
	copy(result, rs.requirements)
	return result
}

// Len returns the number of requirements in the set.
func (rs *RequirementSet) Len() int {
	return len(rs.requirements)
}

// Validate validates all requirements in the set and returns any errors found.
func (rs *RequirementSet) Validate() []error {
	var errs []error
	for _, r := range rs.requirements {
		if err := r.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("requirement %s: %w", r.ID, err))
		}
	}
	return errs
}
