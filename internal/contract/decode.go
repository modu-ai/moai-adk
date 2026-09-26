package contract

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// ErrSchemaInvalid is wrapped by every Decode error. Verify maps any decode
// failure to exactly `reasons == ["schema_invalid"]` and stops evaluating.
var ErrSchemaInvalid = errors.New("contract: schema invalid")

func schemaErr(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrSchemaInvalid, fmt.Sprintf(format, a...))
}

// Decode strictly decodes a contract.yaml document (REQ-CONTRACT-003).
//
// Rejected, each as an error wrapping ErrSchemaInvalid: an empty document, a
// document whose root is not a mapping, more than one YAML document, a
// duplicate key, a value of the wrong type, and a field the schema does not
// define at any depth. The decoded contract records which sections carried a
// non-null value, so HasSection can tell an absent section from one whose
// value happens to be the zero value.
func Decode(raw []byte) (*Contract, error) {
	root, err := parseSingleDocument(raw)
	if err != nil {
		return nil, err
	}
	if root.Kind != yaml.MappingNode {
		return nil, schemaErr("document root is not a mapping")
	}

	// yaml.Node.Decode does not honor KnownFields, so the strict decode runs
	// over the raw bytes with a fresh decoder.
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	var c Contract
	if err := dec.Decode(&c); err != nil {
		return nil, schemaErr("%v", err)
	}
	c.present = presentKeys(root)
	return &c, nil
}

// parseSingleDocument parses raw leniently and returns the root content node
// of its only document.
func parseSingleDocument(raw []byte) (*yaml.Node, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, schemaErr("empty document")
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	var doc yaml.Node
	if err := dec.Decode(&doc); err != nil {
		return nil, schemaErr("%v", err)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err == nil {
		return nil, schemaErr("more than one YAML document")
	} else if !errors.Is(err, io.EOF) {
		return nil, schemaErr("%v", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 {
		return nil, schemaErr("empty document")
	}
	return doc.Content[0], nil
}

// presentKeys returns the top-level keys (and the nested ownership keys, as
// "ownership.<key>") whose value is not YAML null.
func presentKeys(root *yaml.Node) map[string]bool {
	present := map[string]bool{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		key, val := root.Content[i].Value, root.Content[i+1]
		if isNull(val) {
			continue
		}
		present[key] = true
		if key == SectionOwnership && val.Kind == yaml.MappingNode {
			for j := 0; j+1 < len(val.Content); j += 2 {
				if !isNull(val.Content[j+1]) {
					present[SectionOwnership+"."+val.Content[j].Value] = true
				}
			}
		}
	}
	return present
}

func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.ShortTag() == "!!null"
}

// hasTopLevelSignature reports whether raw, parsed leniently, has a non-null
// top-level `signature` key. Verify uses it to choose the State of a
// document that failed strict decoding.
func hasTopLevelSignature(raw []byte) bool {
	root, err := parseSingleDocument(raw)
	if err != nil || root.Kind != yaml.MappingNode {
		return false
	}
	return presentKeys(root)[SectionSignature]
}

// HasSection reports whether the contract carries the named section
// (a Section* constant). For a decoded contract this is the decode-time
// presence of a non-null value; for a Contract built in Go it falls back to
// a non-zero check.
func (c *Contract) HasSection(name string) bool {
	if c == nil {
		return false
	}
	if c.present != nil {
		return c.present[name]
	}
	switch name {
	case SectionAcceptance:
		return c.Acceptance != nil
	case SectionInvariants:
		return c.Invariants != nil
	case SectionOwnership:
		return c.Ownership != nil
	case SectionOwnershipWrite:
		return c.Ownership != nil && c.Ownership.Write != nil
	case SectionOwnershipNever:
		return c.Ownership != nil && c.Ownership.Never != nil
	case SectionOwnershipScratch:
		return c.Ownership != nil && c.Ownership.Scratch != nil
	case SectionApproach:
		return c.Approach != ""
	case SectionActions:
		return c.Actions != nil
	case SectionReobserve:
		return c.Reobserve != nil
	case SectionReview:
		return c.Review != nil
	case SectionBudget:
		return c.Budget != nil
	case SectionEscalateOn:
		return c.EscalateOn != nil
	case SectionPlanAudit:
		return c.PlanAudit != nil
	case SectionSignature:
		return c.Signature != nil
	}
	return false
}

// MissingSections returns, in RequiredSections order, the required sections
// the contract lacks. It includes SectionBudget when absent; whether a
// missing budget is an error depends on the signature state (Verify).
func (c *Contract) MissingSections() []string {
	var missing []string
	for _, s := range RequiredSections {
		if !c.HasSection(s) {
			missing = append(missing, s)
		}
	}
	return missing
}
