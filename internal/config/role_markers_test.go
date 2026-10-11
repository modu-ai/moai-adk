package config

import (
	"os"
	"strings"
	"testing"
)

// TestRoleMarkerRegistrySeed pins the Q8 seed: the registry lists exactly
// the factory lane and leader markers over the same constants the launchers
// stamp, lane first (a lane environment carries BOTH keys - the worker key
// is the discriminator). A change here is a deliberate registry edit, never
// an accident.
func TestRoleMarkerRegistrySeed(t *testing.T) {
	got := RoleMarkerRegistry()
	if len(got) != 2 {
		t.Fatalf("registry size = %d, want 2 (Q8 seed: factory-lane + factory-leader)", len(got))
	}
	want := []RoleMarker{
		{Name: "factory-lane", EnvKey: EnvMoaiFactoryWorker},
		{Name: "factory-leader", EnvKey: EnvMoaiFactoryWorkers},
	}
	for i, m := range got {
		if m != want[i] {
			t.Errorf("registry[%d] = %+v, want %+v", i, m, want[i])
		}
		if m.Name == "" || m.EnvKey == "" {
			t.Errorf("registry[%d] carries an empty name or env key", i)
		}
	}
}

func TestDetectRoleMarker(t *testing.T) {
	tests := []struct {
		name  string
		env   map[string]string
		want  string
		found bool
	}{
		{name: "lane", env: map[string]string{EnvMoaiFactoryWorker: "lane-1"}, want: "factory-lane", found: true},
		{name: "leader", env: map[string]string{EnvMoaiFactoryWorkers: "3"}, want: "factory-leader", found: true},
		{name: "lane_env_carries_both_keys_lane_wins", env: map[string]string{
			EnvMoaiFactoryWorkers: "3",
			EnvMoaiFactoryWorker:  "lane-2",
		}, want: "factory-lane", found: true},
		{name: "unmarked", env: map[string]string{}, want: "", found: false},
		{name: "empty-value-is-unmarked", env: map[string]string{EnvMoaiFactoryWorker: ""}, want: "", found: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ok := DetectRoleMarker(func(key string) string { return tt.env[key] })
			if ok != tt.found {
				t.Fatalf("found = %v, want %v", ok, tt.found)
			}
			if ok && m.Name != tt.want {
				t.Errorf("role = %q, want %q", m.Name, tt.want)
			}
		})
	}
}

func TestDetectRoleMarkerRegistryDerived(t *testing.T) {
	// The guarded marker set derives FROM the registry: detection consults
	// every registered env key, not a hand-written list. Stamp each
	// registry entry's key one at a time and observe detection follows it.
	for _, m := range RoleMarkerRegistry() {
		t.Run(m.Name, func(t *testing.T) {
			stamp := map[string]string{m.EnvKey: "1"}
			got, ok := DetectRoleMarker(func(key string) string { return stamp[key] })
			if !ok || got.Name != m.Name {
				t.Errorf("registry entry %s not detected from its own key %s", m.Name, m.EnvKey)
			}
		})
	}
}

func TestDetectRoleMarkerEnvBacked(t *testing.T) {
	// t.Setenv-bounded real-environment run: the production lookup is
	// os.Getenv, so prove the seam composes with it. Ambient launch
	// variables are scrubbed first - a lane session's own environment
	// carries MOAI_FACTORY_WORKERS, and an unscrubbed lookup would read it
	// instead of the stamped key (parallel-test contamination guarded by
	// t.Setenv restore).
	t.Setenv(EnvMoaiFactoryWorkers, "")
	_ = os.Unsetenv(EnvMoaiFactoryWorkers)
	t.Setenv(EnvMoaiFactoryWorker, "lane-9")
	m, ok := DetectRoleMarker(os.Getenv)
	if !ok || m.Name != "factory-lane" {
		t.Errorf("os.Getenv lookup: got (%+v, %v), want factory-lane", m, ok)
	}
}

func TestExtractRoleCoreRegions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		marked  bool
	}{
		{
			name:    "unmarked file",
			content: "# Rule\n\nplain body with no markers\n",
			want:    nil,
			marked:  false,
		},
		{
			name: "two regions in order",
			content: "head\n" +
				RoleCoreMarkerStart + "\nA block\n" + RoleCoreMarkerEnd + "\n" +
				"middle\n" +
				RoleCoreMarkerStart + "\nB block\nline two\n" + RoleCoreMarkerEnd + "\n" +
				"tail\n",
			want:   []string{"\nA block\n", "\nB block\nline two\n"},
			marked: true,
		},
		{
			name:    "empty adjacent pair is marked with an empty core",
			content: "intro\n" + RoleCoreMarkerStart + RoleCoreMarkerEnd + "\nbody\n",
			want:    []string{""},
			marked:  true,
		},
		{
			name:    "start marker only",
			content: "intro\n" + RoleCoreMarkerStart + "\nunclosed\n",
			want:    nil,
			marked:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, marked := ExtractRoleCoreRegions(tt.content)
			if marked != tt.marked {
				t.Fatalf("marked = %v, want %v", marked, tt.marked)
			}
			if strings.Join(got, "\x00") != strings.Join(tt.want, "\x00") {
				t.Errorf("regions = %q, want %q", got, tt.want)
			}
		})
	}
}
