package contract

import "errors"

// ErrSchemaInvalid is wrapped by every Decode error. Verify maps any decode
// failure to exactly `reasons == ["schema_invalid"]` and stops evaluating.
var ErrSchemaInvalid = errors.New("contract: schema invalid")

// Decode strictly decodes a contract.yaml document.
func Decode(raw []byte) (*Contract, error) {
	return nil, errors.New("not implemented")
}

// HasSection reports whether the decoded document carried the named section.
func (c *Contract) HasSection(name string) bool {
	return false
}

// MissingSections returns the required sections the contract lacks.
func (c *Contract) MissingSections() []string {
	return nil
}
