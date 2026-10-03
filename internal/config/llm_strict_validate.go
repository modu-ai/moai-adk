package config

// llm_strict_validate.go — the strict type-fidelity check over the STORED llm
// section (REQ-AFR-015, SPEC-WEB-AGENTFM-RESTORE-001 v0.3.0 M7, card t1421).
//
// The section loader is deliberately lenient: a genuine yaml type mismatch
// (a sequence where a string is expected) makes loadLLMSection warn and fall
// back to defaults, so a read path never fails on a broken file. That
// leniency is correct for readers and wrong for the write boundary — a save
// that proceeded would normalize the defect away silently. The console save
// path therefore re-checks the stored section through this function before
// anything is written, and joins a positive finding into its atomic-reject
// set (REQ-AFR-006/007 flow): the re-render carries a per-field error naming
// the offending key and every persisted file stays byte-identical.
//
// Two defect classes are checked here, both measured against the yaml.v3
// decoder this tree uses:
//
//  1. genuine type mismatches — reused from strictUnmarshalSection, the SAME
//     typed-unmarshal pass the settings resolver runs over section files
//     (one definition of "what llm.yaml must type-check as" in the tree);
//  2. string-coercible opt-in values — yaml.v3 coerces "yes"/"on"/"1" into
//     TRUE when the target field is bool, so a quoted or bare "yes" survives
//     the typed pass as an opt-in the operator never wrote as one. The
//     opt-in contract admits an explicit boolean only (REQ-AFR-015), so
//     agent_overrides_consume is additionally required to carry the !!bool
//     tag; any other tag joins the reject with a ConfigTypeError naming the
//     key. The check is scoped to this one key — other llm bools keep the
//     historical lenient decode.

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// consumeKey is the one key the tag-strictness check scopes to.
const consumeKey = "agent_overrides_consume"

// ValidateLLMYAMLSection reads <projectRoot>/.moai/config/sections/llm.yaml
// and strictly type-checks it against the llm section's Go types. It returns
// the *ConfigTypeError naming the offending key on a genuine type mismatch
// or on a non-boolean agent_overrides_consume scalar, and nil when the file
// is absent (greenfield projects have nothing to reject), unreadable, or
// clean.
func ValidateLLMYAMLSection(projectRoot string) error {
	path := filepath.Join(filepath.Clean(projectRoot), ".moai", "config", "sections", "llm.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if typeErr := strictUnmarshalSection(data, "llm", path); typeErr != nil {
		return typeErr
	}
	if tagErr := validateConsumeKeyBool(data, path); tagErr != nil {
		return tagErr
	}
	return nil
}

// validateConsumeKeyBool requires a stored agent_overrides_consume scalar to
// carry the !!bool tag (see the file header for the coercion it closes). It
// runs AFTER the typed pass, so everything structurally ill-formed (a
// non-mapping document, a scalar llm value, an unparseable file) has already
// been rejected upstream — the shape checks here exist only to walk safely,
// and an absent key (at either level) is nothing to reject.
func validateConsumeKeyBool(data []byte, path string) *ConfigTypeError {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil || doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != "llm" {
			continue
		}
		llm := root.Content[i+1]
		for j := 0; j+1 < len(llm.Content); j += 2 {
			if llm.Content[j].Value != consumeKey {
				continue
			}
			scalar := llm.Content[j+1]
			if scalar.Tag == "!!bool" {
				return nil
			}
			return &ConfigTypeError{
				File:         path,
				Key:          consumeKey,
				ExpectedType: "bool",
				ActualValue:  scalar.Value,
			}
		}
		return nil // key absent — nothing to reject
	}
	return nil // no llm mapping — the typed pass owns the shape
}
