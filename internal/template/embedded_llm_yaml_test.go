package template

import (
	"io/fs"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// embeddedLLMYAMLPath is the shipped llm.yaml as it appears INSIDE the embedded
// filesystem. EmbeddedTemplates() returns fs.Sub(embeddedRaw, "templates"), so
// the "templates/" prefix is already stripped and the key carries none.
const embeddedLLMYAMLPath = ".moai/config/sections/llm.yaml"

// TestEmbeddedLLMYAMLMatchesMatrix asserts that the llm.yaml compiled into the
// binary agrees with DefaultProfileMatrix() for the three agents whose profile
// cells the plan/sync effort rebalance moves.
//
// This is the only way to observe the embed: `moai model profile` loads the
// on-disk project config through config.NewLoader().Load(...) and never reads
// the embedded FS, so its output is identical whether or not the binary was
// rebuilt after a template edit. Reading EmbeddedTemplates() directly is what
// makes a stale embed visible.
//
// Why it matters: llm.profiles is consulted BEFORE the Go default, so a shipped
// config carrying stale cells would shadow the Go matrix back to its pre-change
// values on every new install. Nothing else in the repo compares the shipped
// llm.yaml against DefaultProfileMatrix().
func TestEmbeddedLLMYAMLMatchesMatrix(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}

	raw, err := fs.ReadFile(fsys, embeddedLLMYAMLPath)
	if err != nil {
		t.Fatalf("read %s from embedded FS: %v", embeddedLLMYAMLPath, err)
	}

	var wrapper struct {
		LLM struct {
			Profiles map[string]map[string]config.ModelEffort `yaml:"profiles"`
		} `yaml:"llm"`
	}
	if err := yaml.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("unmarshal embedded %s: %v", embeddedLLMYAMLPath, err)
	}
	if len(wrapper.LLM.Profiles) == 0 {
		t.Fatalf("embedded %s carries no llm.profiles block — the comparison below would be vacuous", embeddedLLMYAMLPath)
	}

	matrix := DefaultProfileMatrix()
	agents := []string{"manager-spec", "plan-auditor", "manager-docs"}

	for _, profile := range []string{PerformanceTierHigh, PerformanceTierMedium} {
		column, ok := wrapper.LLM.Profiles[profile]
		if !ok {
			t.Errorf("embedded llm.profiles has no %q column", profile)
			continue
		}
		for _, agent := range agents {
			want, inMatrix := matrix[profile][agent]
			if !inMatrix {
				t.Errorf("%s/%s: absent from DefaultProfileMatrix()", profile, agent)
				continue
			}
			got, present := column[agent]
			if !present {
				t.Errorf("%s/%s: absent from the embedded llm.yaml", profile, agent)
				continue
			}
			if got != want {
				t.Errorf("%s/%s: embedded llm.yaml has %+v, DefaultProfileMatrix() has %+v — the shipped config would shadow the Go matrix",
					profile, agent, got, want)
			}
		}
	}
}
