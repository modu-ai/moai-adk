package config

// deploy_mode.go — SPEC-INIT-SHRINK-001 REQ-009 (OD-5 settled (a)
// 2026-10-03): the project deploy-mode record. The key is
// `deployment_mode: plugin|local` in .moai/config/sections/llm.yaml, read
// through the same loadYAMLFile seam as llm.harness so the two records
// cannot drift apart in format. The write side is
// template.ApplyDeployMode (the ApplyHarness line-patch shape).
//
// A record-less project reads "" — the value that routes update to the
// migration path (REQ-015). An out-of-set value also reads "": a project
// whose record cannot be trusted is treated as record-less, never as an
// invented mode.

import (
	"path/filepath"
)

// deployModeFileWrapper reads only the deployment_mode key of llm.yaml;
// the loader is non-strict, so every other key is ignored.
type deployModeFileWrapper struct {
	LLM struct {
		DeploymentMode string `yaml:"deployment_mode"`
	} `yaml:"llm"`
}

// validDeployModes is the closed set of deployment_mode values.
var validDeployModes = map[string]struct{}{
	"plugin": {},
	"local":  {},
}

// IsValidDeployMode reports whether value is a member of the closed set
// (plugin, local). The empty string is NOT a member: callers that mean "no
// record" handle that case themselves.
func IsValidDeployMode(value string) bool {
	_, ok := validDeployModes[value]
	return ok
}

// ReadDeployMode returns the recorded deployment_mode for the project rooted
// at projectRoot: the file value when it is a member of the closed set, ""
// otherwise. An absent file, an absent key, or an out-of-set value all read
// as "" — the record-less project shape.
func ReadDeployMode(projectRoot string) string {
	return ReadDeployModeFrom(filepath.Join(projectRoot, ".moai", "config", "sections"))
}

// ReadDeployModeFrom is ReadDeployMode against an explicit sections
// directory — the form update's restore step uses to read the PRE-UPDATE
// backup, whose directory layout is a sections/ tree.
func ReadDeployModeFrom(sectionsDir string) string {
	wrapper := &deployModeFileWrapper{}
	if _, err := loadYAMLFile(sectionsDir, "llm.yaml", wrapper); err != nil {
		return ""
	}
	if IsValidDeployMode(wrapper.LLM.DeploymentMode) {
		return wrapper.LLM.DeploymentMode
	}
	return ""
}
