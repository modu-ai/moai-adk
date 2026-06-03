package config

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestConfigStructCreation(t *testing.T) {
	t.Parallel()

	cfg := Config{
		User:     models.UserConfig{Name: "TestUser"},
		Language: models.LanguageConfig{ConversationLanguage: "ko"},
		Quality:  models.QualityConfig{DevelopmentMode: models.ModeDDD},
		Project:  models.ProjectConfig{},
		GitStrategy: GitStrategyConfig{
			AutoBranch:   true,
			BranchPrefix: "feature/",
			CommitStyle:  "conventional",
		},
		System: SystemConfig{
			Version:  "1.0.0",
			LogLevel: "debug",
		},
		LLM: LLMConfig{
			DefaultModel: "sonnet",
		},
		Pricing: PricingConfig{
			TokenBudget: 100000,
		},
		Ralph: RalphConfig{
			MaxIterations: 3,
		},
		Workflow: WorkflowConfig{
			PlanTokens: 30000,
		},
	}

	if cfg.User.Name != "TestUser" {
		t.Errorf("User.Name: got %q, want %q", cfg.User.Name, "TestUser")
	}
	if cfg.Language.ConversationLanguage != "ko" {
		t.Errorf("Language.ConversationLanguage: got %q, want %q", cfg.Language.ConversationLanguage, "ko")
	}
	if cfg.Quality.DevelopmentMode != models.ModeDDD {
		t.Errorf("Quality.DevelopmentMode: got %q, want %q", cfg.Quality.DevelopmentMode, models.ModeDDD)
	}
	if cfg.GitStrategy.AutoBranch != true {
		t.Error("GitStrategy.AutoBranch: expected true")
	}
	if cfg.GitStrategy.BranchPrefix != "feature/" {
		t.Errorf("GitStrategy.BranchPrefix: got %q, want %q", cfg.GitStrategy.BranchPrefix, "feature/")
	}
	if cfg.System.LogLevel != "debug" {
		t.Errorf("System.LogLevel: got %q, want %q", cfg.System.LogLevel, "debug")
	}
	if cfg.LLM.DefaultModel != "sonnet" {
		t.Errorf("LLM.DefaultModel: got %q, want %q", cfg.LLM.DefaultModel, "sonnet")
	}
	if cfg.Pricing.TokenBudget != 100000 {
		t.Errorf("Pricing.TokenBudget: got %d, want %d", cfg.Pricing.TokenBudget, 100000)
	}
	if cfg.Ralph.MaxIterations != 3 {
		t.Errorf("Ralph.MaxIterations: got %d, want %d", cfg.Ralph.MaxIterations, 3)
	}
	if cfg.Workflow.PlanTokens != 30000 {
		t.Errorf("Workflow.PlanTokens: got %d, want %d", cfg.Workflow.PlanTokens, 30000)
	}
}

func TestConfigZeroValue(t *testing.T) {
	t.Parallel()

	var cfg Config
	if cfg.User.Name != "" {
		t.Errorf("zero-value User.Name: got %q, want empty", cfg.User.Name)
	}
	if cfg.Quality.DevelopmentMode != "" {
		t.Errorf("zero-value Quality.DevelopmentMode: got %q, want empty", cfg.Quality.DevelopmentMode)
	}
	if cfg.GitStrategy.AutoBranch != false {
		t.Error("zero-value GitStrategy.AutoBranch: expected false")
	}
	if cfg.Pricing.TokenBudget != 0 {
		t.Errorf("zero-value Pricing.TokenBudget: got %d, want 0", cfg.Pricing.TokenBudget)
	}
}

func TestIsValidSectionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"user is valid", "user", true},
		{"language is valid", "language", true},
		{"quality is valid", "quality", true},
		{"project is valid", "project", true},
		{"git_strategy is valid", "git_strategy", true},
		{"git_convention is valid", "git_convention", true},
		{"system is valid", "system", true},
		{"llm is valid", "llm", true},
		{"pricing is valid", "pricing", true},
		{"ralph is valid", "ralph", true},
		{"workflow is valid", "workflow", true},
		{"statusline is valid", "statusline", true},
		{"empty string is invalid", "", false},
		{"unknown section is invalid", "unknown", false},
		{"User uppercase is invalid", "User", false},
		{"QUALITY uppercase is invalid", "QUALITY", false},
		{"git-strategy with hyphen is invalid", "git-strategy", false},
		{"space-padded is invalid", " user ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidSectionName(tt.input); got != tt.want {
				t.Errorf("IsValidSectionName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidSectionNames(t *testing.T) {
	t.Parallel()

	names := ValidSectionNames()

	// Verify count
	if len(names) != 16 {
		t.Fatalf("expected 16 section names, got %d", len(names))
	}

	// Verify all expected names are present
	expected := map[string]bool{
		"user": true, "language": true, "quality": true, "project": true,
		"git_strategy": true, "git_convention": true, "system": true, "llm": true,
		"pricing": true, "ralph": true, "workflow": true, "state": true,
		"statusline": true, "gate": true, "sunset": true, "research": true,
	}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected section name: %q", name)
		}
	}
}

func TestValidSectionNamesReturnsCopy(t *testing.T) {
	t.Parallel()

	names1 := ValidSectionNames()
	names2 := ValidSectionNames()

	// Mutating one slice must not affect the other
	names1[0] = "MUTATED"
	if names2[0] == "MUTATED" {
		t.Error("ValidSectionNames() returned the same underlying slice, expected a copy")
	}
}

func TestGitStrategyConfigFields(t *testing.T) {
	t.Parallel()

	cfg := GitStrategyConfig{
		AutoBranch:        true,
		BranchPrefix:      "moai/",
		CommitStyle:       "conventional",
		WorktreeRoot:      "/tmp/worktree",
		Provider:          "gitlab",
		GitLabInstanceURL: "https://gitlab.company.com",
	}
	if !cfg.AutoBranch {
		t.Error("AutoBranch: expected true")
	}
	if cfg.BranchPrefix != "moai/" {
		t.Errorf("BranchPrefix: got %q, want %q", cfg.BranchPrefix, "moai/")
	}
	if cfg.CommitStyle != "conventional" {
		t.Errorf("CommitStyle: got %q, want %q", cfg.CommitStyle, "conventional")
	}
	if cfg.WorktreeRoot != "/tmp/worktree" {
		t.Errorf("WorktreeRoot: got %q, want %q", cfg.WorktreeRoot, "/tmp/worktree")
	}
	if cfg.Provider != "gitlab" {
		t.Errorf("Provider: got %q, want %q", cfg.Provider, "gitlab")
	}
	if cfg.GitLabInstanceURL != "https://gitlab.company.com" {
		t.Errorf("GitLabInstanceURL: got %q, want %q", cfg.GitLabInstanceURL, "https://gitlab.company.com")
	}
}

func TestSystemConfigFields(t *testing.T) {
	t.Parallel()

	cfg := SystemConfig{
		Version:        "2.0.0",
		LogLevel:       "warn",
		LogFormat:      "json",
		NoColor:        true,
		NonInteractive: true,
	}
	if cfg.Version != "2.0.0" {
		t.Errorf("Version: got %q, want %q", cfg.Version, "2.0.0")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, "warn")
	}
	if cfg.LogFormat != "json" {
		t.Errorf("LogFormat: got %q, want %q", cfg.LogFormat, "json")
	}
	if !cfg.NoColor {
		t.Error("NoColor: expected true")
	}
	if !cfg.NonInteractive {
		t.Error("NonInteractive: expected true")
	}
}

func TestLLMConfigFields(t *testing.T) {
	t.Parallel()

	cfg := LLMConfig{
		DefaultModel: "opus",
		QualityModel: "opus",
		SpeedModel:   "haiku",
	}
	if cfg.DefaultModel != "opus" {
		t.Errorf("DefaultModel: got %q, want %q", cfg.DefaultModel, "opus")
	}
	if cfg.QualityModel != "opus" {
		t.Errorf("QualityModel: got %q, want %q", cfg.QualityModel, "opus")
	}
	if cfg.SpeedModel != "haiku" {
		t.Errorf("SpeedModel: got %q, want %q", cfg.SpeedModel, "haiku")
	}
}

func TestPricingConfigFields(t *testing.T) {
	t.Parallel()

	cfg := PricingConfig{
		TokenBudget:  500000,
		CostTracking: true,
	}
	if cfg.TokenBudget != 500000 {
		t.Errorf("TokenBudget: got %d, want %d", cfg.TokenBudget, 500000)
	}
	if !cfg.CostTracking {
		t.Error("CostTracking: expected true")
	}
}

func TestRalphConfigFields(t *testing.T) {
	t.Parallel()

	cfg := RalphConfig{
		MaxIterations: 10,
		AutoConverge:  true,
		HumanReview:   false,
	}
	if cfg.MaxIterations != 10 {
		t.Errorf("MaxIterations: got %d, want %d", cfg.MaxIterations, 10)
	}
	if !cfg.AutoConverge {
		t.Error("AutoConverge: expected true")
	}
	if cfg.HumanReview {
		t.Error("HumanReview: expected false")
	}
}

// TestWorkflowConfigFields verifies the deprecated FLAT fields remain
// constructible for SPEC-CONFIG-001 backward-compat after the Option (c) rename
// (AutoClear bool → AutoClearLegacy; AutoSelection → AutoSelectionLegacy;
// PlanTokens/RunTokens/SyncTokens retained).
func TestWorkflowConfigFields(t *testing.T) {
	t.Parallel()

	cfg := WorkflowConfig{
		AutoClearLegacy: true,
		PlanTokens:      30000,
		RunTokens:       180000,
		SyncTokens:      40000,
	}
	if !cfg.AutoClearLegacy {
		t.Error("AutoClearLegacy: expected true")
	}
	if cfg.PlanTokens != 30000 {
		t.Errorf("PlanTokens: got %d, want %d", cfg.PlanTokens, 30000)
	}
	if cfg.RunTokens != 180000 {
		t.Errorf("RunTokens: got %d, want %d", cfg.RunTokens, 180000)
	}
	if cfg.SyncTokens != 40000 {
		t.Errorf("SyncTokens: got %d, want %d", cfg.SyncTokens, 40000)
	}
}

// TestWorkflowConfigNestedFieldReachability verifies every nested dotted path
// declared by AC-WSE-001 exists on the WorkflowConfig type (REQ-WSE-001).
func TestWorkflowConfigNestedFieldReachability(t *testing.T) {
	t.Parallel()

	type pathCheck struct {
		typ  reflect.Type
		path []string // nested field names
	}
	wfType := reflect.TypeOf(WorkflowConfig{})
	checks := []pathCheck{
		{wfType, []string{"AutoClear", "Enabled"}},
		{wfType, []string{"AutoClear", "AfterPlan"}},
		{wfType, []string{"AutoClear", "AfterRun"}},
		{wfType, []string{"AutoClear", "TokenThreshold"}},
		{wfType, []string{"Completion", "DetectInOutput"}},
		{wfType, []string{"Completion", "Markers", "Complete"}},
		{wfType, []string{"Completion", "Markers", "Done"}},
		{wfType, []string{"DefaultMode"}},
		{wfType, []string{"ExecutionMode"}},
		{wfType, []string{"LoopPrevention", "FailurePatternDetection"}},
		{wfType, []string{"LoopPrevention", "MaxIterations"}},
		{wfType, []string{"LoopPrevention", "MaxRetriesPerOperation"}},
		{wfType, []string{"TokenBudget", "Plan"}},
		{wfType, []string{"TokenBudget", "Run"}},
		{wfType, []string{"TokenBudget", "Sync"}},
		{wfType, []string{"Worktree", "AutoCleanup"}},
		{wfType, []string{"Worktree", "AutoCreate"}},
		{wfType, []string{"Worktree", "AutoMerge"}},
		{wfType, []string{"Worktree", "SessionNamePattern"}},
		{wfType, []string{"Worktree", "TmuxPreferred"}},
	}
	for _, c := range checks {
		cur := c.typ
		full := strings.Join(c.path, ".")
		for _, name := range c.path {
			f, ok := cur.FieldByName(name)
			if !ok {
				t.Errorf("path %q: field %q not found on %s", full, name, cur)
				break
			}
			cur = f.Type
		}
	}
}

// TestTeamConfigStructShape asserts TeamConfig and RoleProfileEntry have exactly
// the expected exported field sets and that Patterns is absent (REQ-WSE-002,
// EXCL-WSE-004).
func TestTeamConfigStructShape(t *testing.T) {
	t.Parallel()

	wantTeamFields := map[string]bool{
		"AutoSelection":       true,
		"Enabled":             true,
		"MaxTeammates":        true,
		"DefaultModel":        true,
		"DelegateMode":        true,
		"RequirePlanApproval": true,
		"RoleProfileKeys":     true,
		"RoleProfiles":        true,
	}
	gotTeamFields := map[string]bool{}
	for _, f := range reflect.VisibleFields(reflect.TypeOf(TeamConfig{})) {
		gotTeamFields[f.Name] = true
	}
	if gotTeamFields["Patterns"] {
		t.Error("TeamConfig must NOT contain a Patterns field (EXCL-WSE-004)")
	}
	for name := range wantTeamFields {
		if !gotTeamFields[name] {
			t.Errorf("TeamConfig missing expected field %q", name)
		}
	}
	for name := range gotTeamFields {
		if !wantTeamFields[name] {
			t.Errorf("TeamConfig has unexpected field %q", name)
		}
	}

	wantEntryFields := map[string]bool{
		"Description": true,
		"Isolation":   true,
		"Mode":        true,
		"Model":       true,
	}
	gotEntryFields := map[string]bool{}
	for _, f := range reflect.VisibleFields(reflect.TypeOf(RoleProfileEntry{})) {
		gotEntryFields[f.Name] = true
	}
	for name := range wantEntryFields {
		if !gotEntryFields[name] {
			t.Errorf("RoleProfileEntry missing expected field %q", name)
		}
	}
	for name := range gotEntryFields {
		if !wantEntryFields[name] {
			t.Errorf("RoleProfileEntry has unexpected field %q", name)
		}
	}
}

// TestWorkflowConfigLegacyFieldsRenamed verifies the FLAT-field rename +
// deprecation pattern (REQ-WSE-004): AutoClearLegacy / AutoSelectionLegacy exist
// with the correct types, and the legacy scalar AutoClear bool no longer exists.
func TestWorkflowConfigLegacyFieldsRenamed(t *testing.T) {
	t.Parallel()

	wfType := reflect.TypeOf(WorkflowConfig{})

	acl, ok := wfType.FieldByName("AutoClearLegacy")
	if !ok {
		t.Fatal("AutoClearLegacy field missing")
	}
	if acl.Type.Kind() != reflect.Bool {
		t.Errorf("AutoClearLegacy: got kind %s, want bool", acl.Type.Kind())
	}
	if tag := acl.Tag.Get("yaml"); tag != "-" {
		t.Errorf("AutoClearLegacy yaml tag: got %q, want %q", tag, "-")
	}

	asl, ok := wfType.FieldByName("AutoSelectionLegacy")
	if !ok {
		t.Fatal("AutoSelectionLegacy field missing")
	}
	if asl.Type != reflect.TypeOf(TeamAutoSelectionConfig{}) {
		t.Errorf("AutoSelectionLegacy: got type %s, want TeamAutoSelectionConfig", asl.Type)
	}

	// AutoClear is now a nested struct, NOT a bool.
	ac, ok := wfType.FieldByName("AutoClear")
	if !ok {
		t.Fatal("AutoClear (nested) field missing")
	}
	if ac.Type.Kind() == reflect.Bool {
		t.Error("AutoClear must no longer be a bool (renamed to AutoClearLegacy)")
	}

	// The deprecated token fields retain their identifiers with yaml:"-".
	for _, name := range []string{"PlanTokens", "RunTokens", "SyncTokens"} {
		f, ok := wfType.FieldByName(name)
		if !ok {
			t.Errorf("%s field missing", name)
			continue
		}
		if tag := f.Tag.Get("yaml"); tag != "-" {
			t.Errorf("%s yaml tag: got %q, want %q", name, tag, "-")
		}
	}
}

// TestWorkflowAccessorMethods verifies the Option (c) accessors return the
// nested values (REQ-WSE-005).
func TestWorkflowAccessorMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(c *Config)
		assert func(t *testing.T, c *Config)
	}{
		{
			name: "AutoClearEnabled",
			setup: func(c *Config) {
				c.Workflow.AutoClear.Enabled = true
			},
			assert: func(t *testing.T, c *Config) {
				if got := c.WorkflowAutoClearEnabled(); got != true {
					t.Errorf("WorkflowAutoClearEnabled(): got %v, want true", got)
				}
			},
		},
		{
			name: "PlanTokens",
			setup: func(c *Config) {
				c.Workflow.TokenBudget.Plan = 12345
			},
			assert: func(t *testing.T, c *Config) {
				if got := c.WorkflowPlanTokens(); got != 12345 {
					t.Errorf("WorkflowPlanTokens(): got %d, want 12345", got)
				}
			},
		},
		{
			name: "RunTokens",
			setup: func(c *Config) {
				c.Workflow.TokenBudget.Run = 67890
			},
			assert: func(t *testing.T, c *Config) {
				if got := c.WorkflowRunTokens(); got != 67890 {
					t.Errorf("WorkflowRunTokens(): got %d, want 67890", got)
				}
			},
		},
		{
			name: "SyncTokens",
			setup: func(c *Config) {
				c.Workflow.TokenBudget.Sync = 24680
			},
			assert: func(t *testing.T, c *Config) {
				if got := c.WorkflowSyncTokens(); got != 24680 {
					t.Errorf("WorkflowSyncTokens(): got %d, want 24680", got)
				}
			},
		},
		{
			name: "TeamAutoSelection",
			setup: func(c *Config) {
				c.Workflow.Team.AutoSelection = TeamAutoSelectionConfig{
					MinDomainsForTeam:  4,
					MinFilesForTeam:    11,
					MinComplexityScore: 8,
				}
			},
			assert: func(t *testing.T, c *Config) {
				got := c.WorkflowTeamAutoSelection()
				want := TeamAutoSelectionConfig{MinDomainsForTeam: 4, MinFilesForTeam: 11, MinComplexityScore: 8}
				if got != want {
					t.Errorf("WorkflowTeamAutoSelection(): got %+v, want %+v", got, want)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Config{}
			tt.setup(c)
			tt.assert(t, c)
		})
	}
}

func TestLSPQualityGatesFields(t *testing.T) {
	t.Parallel()

	gates := LSPQualityGates{
		Enabled: true,
		Plan:    PlanGate{RequireBaseline: true},
		Run: RunGate{
			MaxErrors:       0,
			MaxTypeErrors:   0,
			MaxLintErrors:   0,
			AllowRegression: false,
		},
		Sync: SyncGate{
			MaxErrors:       0,
			MaxWarnings:     10,
			RequireCleanLSP: true,
		},
		CacheTTLSeconds: 5,
		TimeoutSeconds:  3,
	}

	if !gates.Enabled {
		t.Error("Enabled: expected true")
	}
	if !gates.Plan.RequireBaseline {
		t.Error("Plan.RequireBaseline: expected true")
	}
	if gates.Run.MaxErrors != 0 {
		t.Errorf("Run.MaxErrors: got %d, want 0", gates.Run.MaxErrors)
	}
	if gates.Run.AllowRegression {
		t.Error("Run.AllowRegression: expected false")
	}
	if gates.Sync.MaxWarnings != 10 {
		t.Errorf("Sync.MaxWarnings: got %d, want 10", gates.Sync.MaxWarnings)
	}
	if !gates.Sync.RequireCleanLSP {
		t.Error("Sync.RequireCleanLSP: expected true")
	}
	if gates.CacheTTLSeconds != 5 {
		t.Errorf("CacheTTLSeconds: got %d, want 5", gates.CacheTTLSeconds)
	}
	if gates.TimeoutSeconds != 3 {
		t.Errorf("TimeoutSeconds: got %d, want 3", gates.TimeoutSeconds)
	}
}

// TestRoleProfile_Sandbox_DefaultByRole verifies RoleProfile.Sandbox field exists
// and can be set to the 4 valid sandbox values.
// T-RT003-09: Extend types_test.go for SPEC-V3R2-RT-003 config types.
func TestRoleProfile_Sandbox_DefaultByRole(t *testing.T) {
	t.Parallel()

	validValues := []string{"none", "bubblewrap", "seatbelt", "docker", ""}

	for _, v := range validValues {
		rp := RoleProfile{Sandbox: v}
		if rp.Sandbox != v {
			t.Errorf("RoleProfile.Sandbox: got %q, want %q", rp.Sandbox, v)
		}
	}
}

// TestSecuritySandbox_Fields verifies SecuritySandbox struct has all required fields.
// T-RT003-09: SPEC-V3R2-RT-003 REQ-008/020/030/031.
func TestSecuritySandbox_Fields(t *testing.T) {
	t.Parallel()

	ss := SecuritySandbox{
		Required:         true,
		NetworkAllowlist: []string{"internal.company.com"},
		EnvScrubExtra:    []string{"MY_SECRET_TOKEN"},
		DockerImage:      "moai/sandbox:v1",
	}

	if !ss.Required {
		t.Error("SecuritySandbox.Required: expected true")
	}
	if len(ss.NetworkAllowlist) != 1 || ss.NetworkAllowlist[0] != "internal.company.com" {
		t.Errorf("SecuritySandbox.NetworkAllowlist: got %v, want [internal.company.com]", ss.NetworkAllowlist)
	}
	if len(ss.EnvScrubExtra) != 1 || ss.EnvScrubExtra[0] != "MY_SECRET_TOKEN" {
		t.Errorf("SecuritySandbox.EnvScrubExtra: got %v, want [MY_SECRET_TOKEN]", ss.EnvScrubExtra)
	}
	if ss.DockerImage != "moai/sandbox:v1" {
		t.Errorf("SecuritySandbox.DockerImage: got %q, want %q", ss.DockerImage, "moai/sandbox:v1")
	}
}

// T-MIG003-05: TestSunsetConfig_DORMANT_GodocMarker verifies that the godoc block
// before type SunsetConfig struct contains the DORMANT literal and the
// "Activation deferred to a future SPEC" phrase.
// Maps to: REQ-MIG003-006/015, AC-MIG003-06
func TestSunsetConfig_DORMANT_GodocMarker(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("os.ReadFile(types.go): %v", err)
	}

	content := string(data)

	// Find the position of the struct declaration
	structDecl := "type SunsetConfig struct"
	idx := strings.Index(content, structDecl)
	if idx < 0 {
		t.Fatalf("types.go does not contain %q", structDecl)
	}

	// Extract the 500 chars before the struct declaration (should contain the godoc)
	start := idx - 500
	if start < 0 {
		start = 0
	}
	preceding := content[start:idx]

	if !strings.Contains(preceding, "DORMANT") {
		t.Errorf("godoc before SunsetConfig does not contain 'DORMANT'\npreceding text:\n%s", preceding)
	}
	if !strings.Contains(preceding, "Activation deferred to a future SPEC") {
		t.Errorf("godoc before SunsetConfig does not contain 'Activation deferred to a future SPEC'\npreceding text:\n%s", preceding)
	}
}

// T-MIG003-05: TestRootConfig_HasFourNewSectionFields verifies via reflection that
// the root Config struct exposes the 4 new section fields.
// Maps to: REQ-MIG003-001, AC-MIG003-01
func TestRootConfig_HasFourNewSectionFields(t *testing.T) {
	t.Parallel()

	required := []string{"Constitution", "ContextSearch", "Interview", "Design"}
	cfgType := reflect.TypeOf(Config{})

	for _, fieldName := range required {
		_, ok := cfgType.FieldByName(fieldName)
		if !ok {
			t.Errorf("Config struct missing field %q (REQ-MIG003-001)", fieldName)
		}
	}
}
