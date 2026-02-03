package foundation

import (
	"encoding/json"
	"testing"
)

func TestRequirementTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rt   RequirementType
		want string
	}{
		{name: "Ubiquitous", rt: Ubiquitous, want: "ubiquitous"},
		{name: "EventDriven", rt: EventDriven, want: "event_driven"},
		{name: "UnwantedBehavior", rt: UnwantedBehavior, want: "unwanted_behavior"},
		{name: "StateDriven", rt: StateDriven, want: "state_driven"},
		{name: "Optional", rt: Optional, want: "optional"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.rt.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequirementTypeIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rt   RequirementType
		want bool
	}{
		{name: "Ubiquitous", rt: Ubiquitous, want: true},
		{name: "EventDriven", rt: EventDriven, want: true},
		{name: "UnwantedBehavior", rt: UnwantedBehavior, want: true},
		{name: "StateDriven", rt: StateDriven, want: true},
		{name: "Optional", rt: Optional, want: true},
		{name: "empty_string", rt: RequirementType(""), want: false},
		{name: "invalid_type", rt: RequirementType("invalid"), want: false},
		{name: "partial_match", rt: RequirementType("ubiquitou"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.rt.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllRequirementTypes(t *testing.T) {
	t.Parallel()

	types := AllRequirementTypes()

	if len(types) != 5 {
		t.Errorf("AllRequirementTypes() returned %d types, want 5", len(types))
	}

	expected := map[RequirementType]bool{
		Ubiquitous:       true,
		EventDriven:      true,
		UnwantedBehavior: true,
		StateDriven:      true,
		Optional:         true,
	}

	for _, rt := range types {
		if !expected[rt] {
			t.Errorf("unexpected requirement type: %s", rt)
		}
		delete(expected, rt)
	}

	for rt := range expected {
		t.Errorf("missing requirement type: %s", rt)
	}
}

func TestGetEARSTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rt          RequirementType
		wantErr     bool
		wantName    string
		wantHasType bool
	}{
		{
			name:        "Ubiquitous",
			rt:          Ubiquitous,
			wantErr:     false,
			wantName:    "Ubiquitous",
			wantHasType: true,
		},
		{
			name:        "EventDriven",
			rt:          EventDriven,
			wantErr:     false,
			wantName:    "Event-Driven",
			wantHasType: true,
		},
		{
			name:        "UnwantedBehavior",
			rt:          UnwantedBehavior,
			wantErr:     false,
			wantName:    "Unwanted Behavior",
			wantHasType: true,
		},
		{
			name:        "StateDriven",
			rt:          StateDriven,
			wantErr:     false,
			wantName:    "State-Driven",
			wantHasType: true,
		},
		{
			name:        "Optional",
			rt:          Optional,
			wantErr:     false,
			wantName:    "Optional",
			wantHasType: true,
		},
		{
			name:    "invalid_type",
			rt:      RequirementType("invalid"),
			wantErr: true,
		},
		{
			name:    "empty_type",
			rt:      RequirementType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpl, err := GetEARSTemplate(tt.rt)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tmpl.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tmpl.Name, tt.wantName)
			}
			if tmpl.Template == "" {
				t.Error("Template should not be empty")
			}
			if tmpl.Description == "" {
				t.Error("Description should not be empty")
			}
			if tt.wantHasType && tmpl.Type != tt.rt {
				t.Errorf("Type = %q, want %q", tmpl.Type, tt.rt)
			}
		})
	}
}

func TestGetAllEARSTemplates(t *testing.T) {
	t.Parallel()

	templates := GetAllEARSTemplates()

	if len(templates) != 5 {
		t.Errorf("GetAllEARSTemplates() returned %d templates, want 5", len(templates))
	}

	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("template Name should not be empty")
		}
		if tmpl.Template == "" {
			t.Error("template Template should not be empty")
		}
		if tmpl.Description == "" {
			t.Error("template Description should not be empty")
		}
		if !tmpl.Type.IsValid() {
			t.Errorf("template Type %q is not valid", tmpl.Type)
		}
	}
}

func TestRequirementFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *Requirement
		want string
	}{
		{
			name: "ubiquitous_requirement",
			req: &Requirement{
				Type:        Ubiquitous,
				ID:          "REQ-001",
				Description: "The system shall log all API requests",
			},
			want: "[REQ-001] The system shall log all API requests (Type: ubiquitous)",
		},
		{
			name: "event_driven_requirement",
			req: &Requirement{
				Type:        EventDriven,
				ID:          "REQ-002",
				Description: "When user submits, the system shall validate",
			},
			want: "[REQ-002] When user submits, the system shall validate (Type: event_driven)",
		},
		{
			name: "nil_requirement",
			req:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.req.Format(); got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequirementValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *Requirement
		wantErr bool
	}{
		{
			name: "valid_requirement",
			req: &Requirement{
				Type:               Ubiquitous,
				ID:                 "REQ-001",
				Description:        "The system shall log all API requests",
				AcceptanceCriteria: []string{"Logs are stored"},
			},
			wantErr: false,
		},
		{
			name: "valid_without_acceptance_criteria",
			req: &Requirement{
				Type:        EventDriven,
				ID:          "REQ-002",
				Description: "When event occurs, system shall respond",
			},
			wantErr: false,
		},
		{
			name:    "nil_requirement",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty_ID",
			req: &Requirement{
				Type:        Ubiquitous,
				ID:          "",
				Description: "Some description",
			},
			wantErr: true,
		},
		{
			name: "empty_description",
			req: &Requirement{
				Type:        Ubiquitous,
				ID:          "REQ-001",
				Description: "",
			},
			wantErr: true,
		},
		{
			name: "invalid_type",
			req: &Requirement{
				Type:        RequirementType("invalid"),
				ID:          "REQ-001",
				Description: "Some description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.req.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRequirementSetAdd(t *testing.T) {
	t.Parallel()

	t.Run("add_valid_requirement", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		err := rs.Add(&Requirement{
			Type:        Ubiquitous,
			ID:          "REQ-001",
			Description: "The system shall log all API requests",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rs.Len() != 1 {
			t.Errorf("Len() = %d, want 1", rs.Len())
		}
	})

	t.Run("add_duplicate_ID", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		req := &Requirement{
			Type:        Ubiquitous,
			ID:          "REQ-001",
			Description: "First requirement",
		}
		if err := rs.Add(req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dup := &Requirement{
			Type:        EventDriven,
			ID:          "REQ-001",
			Description: "Duplicate ID",
		}
		if err := rs.Add(dup); err == nil {
			t.Error("expected error for duplicate ID, got nil")
		}
	})

	t.Run("add_nil_requirement", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		if err := rs.Add(nil); err == nil {
			t.Error("expected error for nil requirement, got nil")
		}
	})

	t.Run("add_invalid_requirement", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		err := rs.Add(&Requirement{
			Type:        RequirementType("invalid"),
			ID:          "REQ-001",
			Description: "Invalid type",
		})
		if err == nil {
			t.Error("expected error for invalid requirement, got nil")
		}
	})
}

func TestRequirementSetGet(t *testing.T) {
	t.Parallel()

	rs := NewRequirementSet()
	req := &Requirement{
		Type:        Ubiquitous,
		ID:          "REQ-001",
		Description: "The system shall log all API requests",
	}
	if err := rs.Add(req); err != nil {
		t.Fatalf("Add error: %v", err)
	}

	t.Run("get_existing", func(t *testing.T) {
		t.Parallel()
		got, err := rs.Get("REQ-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "REQ-001" {
			t.Errorf("ID = %q, want %q", got.ID, "REQ-001")
		}
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		t.Parallel()
		_, err := rs.Get("REQ-999")
		if err == nil {
			t.Error("expected error for nonexistent ID, got nil")
		}
	})

	t.Run("get_empty_ID", func(t *testing.T) {
		t.Parallel()
		_, err := rs.Get("")
		if err == nil {
			t.Error("expected error for empty ID, got nil")
		}
	})
}

func TestRequirementSetFilter(t *testing.T) {
	t.Parallel()

	rs := NewRequirementSet()
	reqs := []*Requirement{
		{Type: Ubiquitous, ID: "REQ-001", Description: "Req 1"},
		{Type: Ubiquitous, ID: "REQ-002", Description: "Req 2"},
		{Type: EventDriven, ID: "REQ-003", Description: "Req 3"},
		{Type: StateDriven, ID: "REQ-004", Description: "Req 4"},
	}
	for _, r := range reqs {
		if err := rs.Add(r); err != nil {
			t.Fatalf("Add error: %v", err)
		}
	}

	t.Run("filter_ubiquitous", func(t *testing.T) {
		t.Parallel()
		result := rs.Filter(Ubiquitous)
		if len(result) != 2 {
			t.Errorf("Filter(Ubiquitous) returned %d, want 2", len(result))
		}
	})

	t.Run("filter_event_driven", func(t *testing.T) {
		t.Parallel()
		result := rs.Filter(EventDriven)
		if len(result) != 1 {
			t.Errorf("Filter(EventDriven) returned %d, want 1", len(result))
		}
	})

	t.Run("filter_no_match", func(t *testing.T) {
		t.Parallel()
		result := rs.Filter(Optional)
		if result == nil {
			t.Error("Filter should return empty slice, not nil")
		}
		if len(result) != 0 {
			t.Errorf("Filter(Optional) returned %d, want 0", len(result))
		}
	})
}

func TestRequirementSetAll(t *testing.T) {
	t.Parallel()

	t.Run("empty_set", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		all := rs.All()
		if all == nil {
			t.Error("All() should return empty slice, not nil")
		}
		if len(all) != 0 {
			t.Errorf("All() returned %d, want 0", len(all))
		}
	})

	t.Run("populated_set", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		_ = rs.Add(&Requirement{Type: Ubiquitous, ID: "REQ-001", Description: "Req 1"})
		_ = rs.Add(&Requirement{Type: EventDriven, ID: "REQ-002", Description: "Req 2"})
		all := rs.All()
		if len(all) != 2 {
			t.Errorf("All() returned %d, want 2", len(all))
		}
	})
}

func TestRequirementSetValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid_set", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		_ = rs.Add(&Requirement{Type: Ubiquitous, ID: "REQ-001", Description: "Req 1"})
		_ = rs.Add(&Requirement{Type: EventDriven, ID: "REQ-002", Description: "Req 2"})
		errs := rs.Validate()
		if len(errs) != 0 {
			t.Errorf("Validate() returned %d errors, want 0", len(errs))
		}
	})

	t.Run("empty_set", func(t *testing.T) {
		t.Parallel()
		rs := NewRequirementSet()
		errs := rs.Validate()
		if len(errs) != 0 {
			t.Errorf("Validate() on empty set returned %d errors, want 0", len(errs))
		}
	})
}

func TestRequirementSetLen(t *testing.T) {
	t.Parallel()

	rs := NewRequirementSet()
	if rs.Len() != 0 {
		t.Errorf("empty set Len() = %d, want 0", rs.Len())
	}
	_ = rs.Add(&Requirement{Type: Ubiquitous, ID: "REQ-001", Description: "Req 1"})
	if rs.Len() != 1 {
		t.Errorf("after add Len() = %d, want 1", rs.Len())
	}
}

func TestRequirementJSONRoundTrip(t *testing.T) {
	t.Parallel()

	req := &Requirement{
		Type:               Ubiquitous,
		ID:                 "REQ-001",
		Description:        "The system shall log all API requests",
		AcceptanceCriteria: []string{"Logs are stored", "Logs have timestamps"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Requirement
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Type != req.Type {
		t.Errorf("Type = %q, want %q", got.Type, req.Type)
	}
	if got.ID != req.ID {
		t.Errorf("ID = %q, want %q", got.ID, req.ID)
	}
	if got.Description != req.Description {
		t.Errorf("Description = %q, want %q", got.Description, req.Description)
	}
	if len(got.AcceptanceCriteria) != len(req.AcceptanceCriteria) {
		t.Errorf("AcceptanceCriteria length = %d, want %d", len(got.AcceptanceCriteria), len(req.AcceptanceCriteria))
	}
}

func TestEARSTemplateJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tmpl, err := GetEARSTemplate(Ubiquitous)
	if err != nil {
		t.Fatalf("GetEARSTemplate error: %v", err)
	}

	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got EARSTemplate
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Type != tmpl.Type {
		t.Errorf("Type = %q, want %q", got.Type, tmpl.Type)
	}
	if got.Name != tmpl.Name {
		t.Errorf("Name = %q, want %q", got.Name, tmpl.Name)
	}
}
