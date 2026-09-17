package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

type namingManifest struct {
	SchemaVersion        int                 `json:"schema_version"`
	SpecID               string              `json:"spec_id"`
	BaselineTree         string              `json:"baseline_tree"`
	GeneratedByMilestone string              `json:"generated_by_milestone"`
	SourceInventory      namingInventory     `json:"source_inventory"`
	PathMappings         []namingPathMapping `json:"path_mappings"`
	ProductionPackage    []string            `json:"production_package"`
	PackageTests         []string            `json:"package_tests"`
	ActiveImporters      []string            `json:"active_importers"`
	HistoricalUtilities  []namingHistorical  `json:"historical_utilities"`
	LegacyAllowlist      []namingLegacyAllow `json:"legacy_allowlist"`
	NewNames             map[string]string   `json:"new_names"`
}

type namingInventory struct {
	ProductionPackage   []string           `json:"production_package"`
	PackageTests        []string           `json:"package_tests"`
	ActiveImporters     []string           `json:"active_importers"`
	HistoricalUtilities []namingHistorical `json:"historical_utilities"`
}

type namingPathMapping struct {
	Category    string `json:"category"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type namingHistorical struct {
	Path           string   `json:"path"`
	AllowedSymbols []string `json:"allowed_symbols"`
	Access         string   `json:"access"`
}

type namingLegacyAllow struct {
	Classification   string   `json:"classification"`
	Path             string   `json:"path"`
	Symbols          []string `json:"symbols"`
	Access           string   `json:"access"`
	RemovalCondition string   `json:"removal_condition"`
}

type namingDiff struct{ missingFromTree, extraInTree []string }

func TestNamingMigrationManifestContract(t *testing.T) {
	root := namingRepoRoot(t)
	baseline := namingBaselineInventory(t, root)
	current := namingEffectiveInventory(t, root)
	path := filepath.Join(root, ".moai", "specs", "SPEC-MOAI-GATEWAY-001", "naming-migration-manifest.json")

	t.Run("schema", func(t *testing.T) {
		manifest, err := readNamingManifest(path)
		if err != nil {
			t.Errorf("schema_error=manifest: %v", err)
			return
		}
		if err := validateNamingSchema(manifest, baseline); err != nil {
			t.Errorf("schema_error=%v", err)
		}
	})
	t.Run("current-equality", func(t *testing.T) {
		manifest, err := readNamingManifest(path)
		if err != nil {
			manifest = namingManifest{}
		}
		diff := compareNamingInventory(manifest, current)
		if len(diff.missingFromTree) != 0 || len(diff.extraInTree) != 0 {
			t.Errorf("missing_from_tree=%v extra_in_tree=%v manifest_error=%v", diff.missingFromTree, diff.extraInTree, err)
		}
	})
	t.Run("missing-from-tree", func(t *testing.T) {
		fixture := current
		fixture.ProductionPackage = append([]string(nil), current.ProductionPackage...)
		fixture.ProductionPackage = append(fixture.ProductionPackage, "internal/orchestration/not-in-tree.go")
		slices.Sort(fixture.ProductionPackage)
		diff := compareNamingInventory(fixture, current)
		if !slices.Contains(diff.missingFromTree, "internal/orchestration/not-in-tree.go") {
			t.Errorf("missing_from_tree=%v, want injected path", diff.missingFromTree)
		}
		mappingFixture := namingMappingFixture(baseline)
		if err := validateNamingSchema(mappingFixture, baseline); err != nil {
			t.Fatalf("positive schema fixture: %v", err)
		}
		mappingFixture.PathMappings = mappingFixture.PathMappings[1:]
		if err := validateNamingSchema(mappingFixture, baseline); err == nil {
			t.Error("schema_error=path_mappings accepted missing source mapping")
		}
	})
	t.Run("extra-in-tree", func(t *testing.T) {
		fixture := current
		fixture.HistoricalUtilities = nil
		diff := compareNamingInventory(fixture, current)
		if len(diff.extraInTree) != 1 || diff.extraInTree[0] != "cmd/t657-merge/main.go" {
			t.Errorf("extra_in_tree=%v, want [cmd/t657-merge/main.go]", diff.extraInTree)
		}
		mappingFixture := namingMappingFixture(baseline)
		mappingFixture.PathMappings[1].Destination = mappingFixture.PathMappings[0].Destination
		if err := validateNamingSchema(mappingFixture, baseline); err == nil {
			t.Error("schema_error=path_mappings accepted duplicate destination")
		}
		mutations := []struct {
			name   string
			mutate func(*namingManifest)
		}{
			{"new_names", func(m *namingManifest) { delete(m.NewNames, "dispatch") }},
			{"legacy_allowlist", func(m *namingManifest) { m.LegacyAllowlist = m.LegacyAllowlist[:2] }},
			{"historical_utility", func(m *namingManifest) { m.HistoricalUtilities[0].AllowedSymbols = []string{"wrong"} }},
			{"cardinality_membership", func(m *namingManifest) { m.PackageTests[0] = strings.TrimSuffix(m.PackageTests[0], "_test.go") + ".go" }},
			{"mapping_category", func(m *namingManifest) { m.PathMappings[0].Category = "unknown" }},
		}
		for _, mutation := range mutations {
			candidate := namingMappingFixture(baseline)
			mutation.mutate(&candidate)
			if err := validateNamingSchema(candidate, baseline); err == nil {
				t.Errorf("schema_error=%s mutation accepted", mutation.name)
			}
		}
	})
	t.Run("legacy-read-new-write", func(t *testing.T) {
		root := t.TempDir()
		legacyDir := kanban.LegacyStateDirForRoot(root)
		if err := os.MkdirAll(legacyDir, 0700); err != nil {
			t.Fatal(err)
		}
		legacy := kanban.NewBacklogStore(filepath.Join(legacyDir, "backlog.json"))
		if err := legacy.Mutate(func(record *kanban.BacklogRecord) error {
			record.Items = append(record.Items, kanban.BacklogItem{ID: "t1", Text: "legacy", AddedAt: time.Now().UTC().Format(time.RFC3339), State: kanban.BacklogStateQueued})
			return nil
		}); err != nil {
			t.Fatalf("legacy write fixture: %v", err)
		}
		adopted := kanban.NewBacklogStore(kanban.BacklogPathForRootAdopting(root))
		record, err := adopted.LoadPure()
		if err != nil || len(record.Items) != 1 || record.Items[0].ID != "t1" {
			t.Errorf("legacy read got items=%v err=%v, want t1", record.Items, err)
		}
		if got := filepath.ToSlash(kanban.BacklogPathForRoot(root)); strings.Contains(got, "/kanban/") || !strings.Contains(got, "/todo/") {
			t.Errorf("new write path=%q, want Todo path without active kanban", got)
		}
	})
}

func namingRepoRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve naming test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(here), "..", ".."))
}

func namingBaselineInventory(t *testing.T, root string) namingInventory {
	t.Helper()
	const baseline = "4056f69e1c20d942d4f9fc7363d3d79bffde899a"
	out, err := exec.Command("git", "-C", root, "ls-tree", "-r", "--name-only", baseline).Output()
	if err != nil {
		t.Fatalf("baseline Go inventory: %v", err)
	}
	grep, grepErr := exec.Command("git", "-C", root, "grep", "-l", "github.com/modu-ai/moai-adk/internal/kanban", baseline, "--", "*.go").Output()
	if grepErr != nil {
		t.Fatalf("baseline importer inventory: %v", grepErr)
	}
	baselineImporters := map[string]bool{}
	for _, row := range strings.Fields(string(grep)) {
		baselineImporters[strings.TrimPrefix(row, baseline+":")] = true
	}
	var inventory namingInventory
	for _, path := range strings.Fields(string(out)) {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		if strings.HasPrefix(path, "internal/kanban/") {
			if strings.HasSuffix(path, "_test.go") {
				inventory.PackageTests = append(inventory.PackageTests, path)
			} else {
				inventory.ProductionPackage = append(inventory.ProductionPackage, path)
			}
			continue
		}
		if path == "cmd/t657-merge/main.go" {
			inventory.HistoricalUtilities = append(inventory.HistoricalUtilities, namingHistorical{Path: path, AllowedSymbols: []string{"t657"}, Access: "comment-only"})
			continue
		}
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		if baselineImporters[path] {
			inventory.ActiveImporters = append(inventory.ActiveImporters, path)
		}
	}
	slices.Sort(inventory.ProductionPackage)
	slices.Sort(inventory.PackageTests)
	slices.Sort(inventory.ActiveImporters)
	if len(inventory.ProductionPackage) != 47 || len(inventory.PackageTests) != 85 || len(inventory.ActiveImporters) != 36 || len(inventory.HistoricalUtilities) != 1 {
		t.Fatalf("baseline source_inventory counts production=%d tests=%d importers=%d historical=%d, want 47/85/36/1", len(inventory.ProductionPackage), len(inventory.PackageTests), len(inventory.ActiveImporters), len(inventory.HistoricalUtilities))
	}
	return inventory
}

func namingEffectiveInventory(t *testing.T, root string) namingManifest {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "--", "*.go").Output()
	if err != nil {
		t.Fatalf("effective Go inventory: %v", err)
	}
	m := namingManifest{}
	for _, path := range strings.Fields(string(out)) {
		if path == "internal/orchestration/naming_manifest_contract_test.go" {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			continue
		}
		if strings.HasPrefix(path, "internal/orchestration/") {
			if strings.HasSuffix(path, "_test.go") {
				m.PackageTests = append(m.PackageTests, path)
			} else {
				m.ProductionPackage = append(m.ProductionPackage, path)
			}
			continue
		}
		if path == "cmd/t657-merge/main.go" {
			m.HistoricalUtilities = append(m.HistoricalUtilities, namingHistorical{Path: path, AllowedSymbols: []string{"t657"}, Access: "comment-only"})
			continue
		}
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if strings.Contains(string(raw), "github.com/modu-ai/moai-adk/internal/orchestration") {
			m.ActiveImporters = append(m.ActiveImporters, path)
		}
	}
	slices.Sort(m.ProductionPackage)
	slices.Sort(m.PackageTests)
	slices.Sort(m.ActiveImporters)
	return m
}

func readNamingManifest(path string) (namingManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return namingManifest{}, fmt.Errorf("open naming-migration-manifest.json: %w", err)
	}
	var manifest namingManifest
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return namingManifest{}, err
	}
	return manifest, nil
}

func validateNamingSchema(m namingManifest, baseline namingInventory) error {
	if m.SchemaVersion != 1 || m.SpecID != "SPEC-MOAI-GATEWAY-001" || m.GeneratedByMilestone != "M14-R3" {
		return fmt.Errorf("root scalars version=%d spec=%q milestone=%q", m.SchemaVersion, m.SpecID, m.GeneratedByMilestone)
	}
	if m.BaselineTree != "4056f69e1c20d942d4f9fc7363d3d79bffde899a" {
		return fmt.Errorf("baseline_tree=%q want exact provenance anchor", m.BaselineTree)
	}
	if !equalNamingInventory(m.SourceInventory, baseline) {
		return fmt.Errorf("source_inventory differs from baseline tree")
	}
	wantNames := map[string]string{"execution": "Factory", "queue": "Todo", "ui_status": "Tasks", "dispatch": "Dispatch", "coordination": "Orchestration"}
	if !mapsEqualNaming(m.NewNames, wantNames) {
		return fmt.Errorf("new_names=%v want exact %v", m.NewNames, wantNames)
	}
	if !equalLegacyAllowlist(m.LegacyAllowlist, expectedLegacyAllowlist()) {
		return fmt.Errorf("legacy_allowlist=%+v want exact three migration paths/required symbols", m.LegacyAllowlist)
	}
	if len(m.HistoricalUtilities) != 1 || m.HistoricalUtilities[0].Path != "cmd/t657-merge/main.go" || m.HistoricalUtilities[0].Access != "comment-only" || !slices.Equal(m.HistoricalUtilities[0].AllowedSymbols, []string{"t657"}) {
		return fmt.Errorf("historical_utilities=%+v want exact t657 comment-only row", m.HistoricalUtilities)
	}
	if len(m.ProductionPackage) != 47 || len(m.PackageTests) != 85 || len(m.ActiveImporters) != 36 {
		return fmt.Errorf("current cardinality production=%d tests=%d importers=%d historical=%d want 47/85/36/1", len(m.ProductionPackage), len(m.PackageTests), len(m.ActiveImporters), len(m.HistoricalUtilities))
	}
	for _, p := range m.ProductionPackage {
		if !strings.HasPrefix(p, "internal/orchestration/") || strings.HasSuffix(p, "_test.go") {
			return fmt.Errorf("production_package category mismatch %s", p)
		}
	}
	for _, p := range m.PackageTests {
		if !strings.HasPrefix(p, "internal/orchestration/") || !strings.HasSuffix(p, "_test.go") {
			return fmt.Errorf("package_tests category mismatch %s", p)
		}
	}
	sources, destinations := map[string]bool{}, map[string]bool{}
	allowedCategories := map[string]bool{"production_package": true, "package_tests": true, "active_importers": true, "historical_utilities": true}
	sourceCategory := namingCategoryIndex(m.SourceInventory)
	destinationCategory := namingManifestCategoryIndex(m)
	previous := ""
	for _, row := range m.PathMappings {
		key := row.Category + "\x00" + row.Source + "\x00" + row.Destination
		if previous != "" && key <= previous {
			return fmt.Errorf("path_mappings must be sorted and unique")
		}
		previous = key
		if !allowedCategories[row.Category] || row.Source == "" || row.Destination == "" || sources[row.Source] || destinations[row.Destination] {
			return fmt.Errorf("path_mappings invalid or duplicate row %+v", row)
		}
		if sourceCategory[row.Source] != row.Category || destinationCategory[row.Destination] != row.Category {
			return fmt.Errorf("path_mappings category mismatch %+v", row)
		}
		sources[row.Source], destinations[row.Destination] = true, true
		if strings.HasPrefix(row.Source, "internal/kanban/") && row.Destination != strings.Replace(row.Source, "internal/kanban/", "internal/orchestration/", 1) {
			return fmt.Errorf("path_mappings target mismatch %+v", row)
		}
	}
	for _, source := range flattenNamingInventory(baseline) {
		if !sources[source] {
			return fmt.Errorf("path_mappings missing source %s", source)
		}
	}
	for _, destination := range flattenNamingManifest(m) {
		if !destinations[destination] {
			return fmt.Errorf("path_mappings missing destination %s", destination)
		}
	}
	for name, rows := range map[string][]string{"production_package": m.ProductionPackage, "package_tests": m.PackageTests, "active_importers": m.ActiveImporters} {
		if !slices.IsSorted(rows) || len(slices.Compact(append([]string(nil), rows...))) != len(rows) {
			return fmt.Errorf("%s must be sorted and unique", name)
		}
	}
	for _, row := range m.HistoricalUtilities {
		if row.Path == "" || row.Access != "comment-only" || !slices.IsSorted(row.AllowedSymbols) {
			return fmt.Errorf("historical_utilities invalid row %+v", row)
		}
	}
	for _, row := range m.LegacyAllowlist {
		if row.Classification == "" || row.Path == "" || row.Access != "read-only" || row.RemovalCondition == "" || !slices.IsSorted(row.Symbols) {
			return fmt.Errorf("legacy_allowlist invalid row %+v", row)
		}
	}
	return nil
}

func mapsEqualNaming(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range b {
		if a[key] != value {
			return false
		}
	}
	return true
}

func expectedLegacyAllowlist() []namingLegacyAllow {
	env := []string{"MOAI_KANBAN", "MOAI_KANBAN_BACKEND", "MOAI_KANBAN_CARD", "MOAI_KANBAN_ID", "MOAI_KANBAN_LABEL", "MOAI_KANBAN_LEAD_ADDR", "MOAI_KANBAN_LEAD_NAME", "MOAI_KANBAN_SETTINGS_INJECTED", "MOAI_KANBAN_SPEC"}
	prod := append([]string{"ReadLegacyDispatchEnv"}, env...)
	slices.Sort(prod)
	return []namingLegacyAllow{
		{Classification: "legacy-migration", Path: "internal/orchestration/legacy_env.go", Symbols: prod, Access: "read-only", RemovalCondition: "supported upgrade window ended and migration telemetry is zero"},
		{Classification: "legacy-migration-test", Path: "internal/orchestration/legacy_env_test.go", Symbols: env, Access: "read-only", RemovalCondition: "remove with production reader"},
		{Classification: "legacy-migration-test", Path: "internal/orchestration/legacy_store_test.go", Symbols: []string{"legacy_store_fixture"}, Access: "read-only", RemovalCondition: "remove with production reader"},
	}
}

func equalLegacyAllowlist(a, b []namingLegacyAllow) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Classification != b[i].Classification || a[i].Path != b[i].Path || a[i].Access != b[i].Access || a[i].RemovalCondition != b[i].RemovalCondition || !slices.Equal(a[i].Symbols, b[i].Symbols) {
			return false
		}
	}
	return true
}

func namingCategoryIndex(i namingInventory) map[string]string {
	out := map[string]string{}
	for _, p := range i.ProductionPackage {
		out[p] = "production_package"
	}
	for _, p := range i.PackageTests {
		out[p] = "package_tests"
	}
	for _, p := range i.ActiveImporters {
		out[p] = "active_importers"
	}
	for _, row := range i.HistoricalUtilities {
		out[row.Path] = "historical_utilities"
	}
	return out
}

func namingManifestCategoryIndex(m namingManifest) map[string]string {
	return namingCategoryIndex(namingInventory{ProductionPackage: m.ProductionPackage, PackageTests: m.PackageTests, ActiveImporters: m.ActiveImporters, HistoricalUtilities: m.HistoricalUtilities})
}

func equalNamingInventory(a, b namingInventory) bool {
	return slices.Equal(a.ProductionPackage, b.ProductionPackage) && slices.Equal(a.PackageTests, b.PackageTests) && slices.Equal(a.ActiveImporters, b.ActiveImporters) && slices.Equal(flattenNamingInventory(namingInventory{HistoricalUtilities: a.HistoricalUtilities}), flattenNamingInventory(namingInventory{HistoricalUtilities: b.HistoricalUtilities}))
}
func flattenNamingInventory(i namingInventory) []string {
	m := namingManifest{ProductionPackage: i.ProductionPackage, PackageTests: i.PackageTests, ActiveImporters: i.ActiveImporters, HistoricalUtilities: i.HistoricalUtilities}
	return flattenNamingManifest(m)
}
func flattenNamingManifest(m namingManifest) []string {
	out := append(append(append([]string(nil), m.ProductionPackage...), m.PackageTests...), m.ActiveImporters...)
	for _, row := range m.HistoricalUtilities {
		out = append(out, row.Path)
	}
	slices.Sort(out)
	return slices.Compact(out)
}
func namingMappingFixture(source namingInventory) namingManifest {
	m := namingManifest{SchemaVersion: 1, SpecID: "SPEC-MOAI-GATEWAY-001", BaselineTree: "4056f69e1c20d942d4f9fc7363d3d79bffde899a", GeneratedByMilestone: "M14-R3", SourceInventory: source, NewNames: map[string]string{"execution": "Factory", "queue": "Todo", "ui_status": "Tasks", "dispatch": "Dispatch", "coordination": "Orchestration"}}
	m.LegacyAllowlist = expectedLegacyAllowlist()
	for _, p := range source.ProductionPackage {
		d := strings.Replace(p, "internal/kanban/", "internal/orchestration/", 1)
		m.ProductionPackage = append(m.ProductionPackage, d)
		m.PathMappings = append(m.PathMappings, namingPathMapping{Category: "production_package", Source: p, Destination: d})
	}
	for _, p := range source.PackageTests {
		d := strings.Replace(p, "internal/kanban/", "internal/orchestration/", 1)
		m.PackageTests = append(m.PackageTests, d)
		m.PathMappings = append(m.PathMappings, namingPathMapping{Category: "package_tests", Source: p, Destination: d})
	}
	for _, p := range source.ActiveImporters {
		m.ActiveImporters = append(m.ActiveImporters, p)
		m.PathMappings = append(m.PathMappings, namingPathMapping{Category: "active_importers", Source: p, Destination: p})
	}
	for _, row := range source.HistoricalUtilities {
		m.HistoricalUtilities = append(m.HistoricalUtilities, row)
		m.PathMappings = append(m.PathMappings, namingPathMapping{Category: "historical_utilities", Source: row.Path, Destination: row.Path})
	}
	slices.SortFunc(m.PathMappings, func(a, b namingPathMapping) int {
		return strings.Compare(a.Category+"\x00"+a.Source+"\x00"+a.Destination, b.Category+"\x00"+b.Source+"\x00"+b.Destination)
	})
	return m
}

func compareNamingInventory(manifest, current namingManifest) namingDiff {
	flatten := func(m namingManifest) []string {
		return flattenNamingManifest(m)
	}
	want, got := flatten(manifest), flatten(current)
	diff := namingDiff{}
	for _, path := range want {
		if !slices.Contains(got, path) {
			diff.missingFromTree = append(diff.missingFromTree, path)
		}
	}
	for _, path := range got {
		if !slices.Contains(want, path) {
			diff.extraInTree = append(diff.extraInTree, path)
		}
	}
	return diff
}
