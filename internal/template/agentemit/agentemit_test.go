// agentemit_test.go — SPEC-CODEX-DUAL-AGENTS-001 MS1 specification tests.
//
// Test-first suite for the agent dual-publication emitter (Option A: the .md
// files ARE the neutral layer; the Codex TOML is a transform of
// (.md x agents-codex.yaml manifest)).
//
// Fixtures live in fstest.MapFS (no disk writes). The TOML decoder used for
// round-trip assertions is an INDEPENDENT test-side implementation of the
// emitted grammar subset (basic strings, multi-line literal strings, arrays),
// written from the TOML spec — deliberately not shared with writer.go so a
// writer bug cannot be mirrored by its own validator. The real consumer
// (codex-cli) parses the same artifacts in the MS2 probe smoke.
package agentemit_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// fixtureMD builds a minimal agent .md source in the template frontmatter
// contract (field order name, description block scalar |, tools CSV, model,
// effort, optional skills; body follows the closing delimiter).
func fixtureMD(name, description, tools, model, effort string, skills []string, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: |\n")
	for _, line := range strings.Split(strings.TrimSuffix(description, "\n"), "\n") {
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("tools: " + tools + "\n")
	b.WriteString("model: " + model + "\n")
	b.WriteString("effort: " + effort + "\n")
	if len(skills) > 0 {
		b.WriteString("skills:\n")
		for _, s := range skills {
			b.WriteString("  - " + s + "\n")
		}
	}
	b.WriteString("---\n")
	b.WriteString(body)
	return b.String()
}

// fullFixtureSet returns a representative fixture set exercising: an MCP
// carrier (manager-develop-like), a non-carrier with a skills preload
// (builder-harness-like), and a no-skills agent (plan-auditor-like) whose body
// contains a bare "---" horizontal rule (the frontmatter splitter must anchor
// on the FIRST closing delimiter only).
func fullFixtureSet() fstest.MapFS {
	return fstest.MapFS{
		"agents/mdcarrier.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"mdcarrier",
			"Carrier agent.\nSecond description line.\n",
			"Read, Grep, Glob, Bash, TaskCreate, mcp__moai__spec_audit, mcp__moai__verify_snapshot",
			"inherit", "high",
			[]string{"moai-foundation-core"},
			"\n# mdcarrier\n\nBody line.\n\n---\n\nAfter the rule.\n",
		))},
		"agents/plainagent.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"plainagent",
			"Plain agent with no MCP tools.\n",
			"Read, Write, Edit, Grep, Glob, Bash",
			"inherit", "medium",
			nil, // no skills field at all (plan-auditor-like)
			"\n# plainagent\n\nPlain body.\n",
		))},
		"agents/twoskills.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"twoskills",
			"Two-skill agent.\n",
			"Read, Grep, Glob, Bash, Skill",
			"sonnet", "low",
			[]string{"moai-foundation-core", "moai-workflow-project"},
			"\n# twoskills\n\nBody.\n",
		))},
	}
}

// ---------------------------------------------------------------------------
// Group A — loader (neutral-layer parse contract)
// ---------------------------------------------------------------------------

// TestParseAgentDocParsesFrontmatterContract verifies the .md structure
// contract: block-scalar description, CSV tools, optional skills (0 and 2
// entries both valid), and a body preserved byte-exact even when it contains
// a bare "---" horizontal rule after the closing delimiter (plan-auditor-like).
func TestParseAgentDocParsesFrontmatterContract(t *testing.T) {
	set := fullFixtureSet()
	for _, file := range []string{"agents/mdcarrier.md", "agents/plainagent.md", "agents/twoskills.md"} {
		data, err := set.ReadFile(file)
		if err != nil {
			t.Fatalf("read fixture %s: %v", file, err)
		}
		doc, err := agentemit.ParseAgentDoc(file, data)
		if err != nil {
			t.Fatalf("ParseAgentDoc(%s): %v", file, err)
		}
		if doc.Name == "" {
			t.Errorf("%s: name empty", file)
		}
		if doc.Description == "" || !strings.Contains(doc.Description, "\n") {
			t.Errorf("%s: description must be the decoded multi-line block scalar, got %q", file, doc.Description)
		}
		if len(doc.Tools) == 0 {
			t.Errorf("%s: tools empty", file)
		}
	}
	// Body byte-exactness including the interior "---" rule.
	data, _ := set.ReadFile("agents/mdcarrier.md")
	doc, err := agentemit.ParseAgentDoc("agents/mdcarrier.md", data)
	if err != nil {
		t.Fatalf("ParseAgentDoc: %v", err)
	}
	wantBody := "\n# mdcarrier\n\nBody line.\n\n---\n\nAfter the rule.\n"
	if string(doc.Body) != wantBody {
		t.Errorf("body not byte-exact:\n got %q\nwant %q", doc.Body, wantBody)
	}
	// Skills cardinality: plainagent 0, twoskills 2, mdcarrier 1.
	plain, _ := agentemit.ParseAgentDoc("agents/plainagent.md", mustBytes(t, set, "agents/plainagent.md"))
	if len(plain.Skills) != 0 {
		t.Errorf("plainagent: want 0 skills, got %v", plain.Skills)
	}
	two, _ := agentemit.ParseAgentDoc("agents/twoskills.md", mustBytes(t, set, "agents/twoskills.md"))
	if len(two.Skills) != 2 {
		t.Errorf("twoskills: want 2 skills, got %v", two.Skills)
	}
	carrier, _ := agentemit.ParseAgentDoc("agents/mdcarrier.md", mustBytes(t, set, "agents/mdcarrier.md"))
	if len(carrier.Skills) != 1 {
		t.Errorf("mdcarrier: want 1 skill, got %v", carrier.Skills)
	}
}

// TestParseAgentDocRejectsBrokenSources covers loader fail-closed cases:
// missing name and name/file-stem mismatch (the emitted TOML name must be
// unambiguous — a mismatch would publish two identities for one agent).
func TestParseAgentDocRejectsBrokenSources(t *testing.T) {
	cases := map[string]struct {
		file string
		data string
		want string
	}{
		"missing name": {
			file: "agents/noname.md",
			data: "---\ndescription: |\n  x\ntools: Read\nmodel: inherit\neffort: low\n---\nbody\n",
			want: "name",
		},
		"name stem mismatch": {
			file: "agents/othername.md",
			data: fixtureMD("differentname", "d.\n", "Read", "inherit", "low", nil, "body\n"),
			want: "differentname",
		},
	}
	for label, tc := range cases {
		_, err := agentemit.ParseAgentDoc(tc.file, []byte(tc.data))
		if err == nil {
			t.Errorf("%s: want error, got nil", label)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q must name the offending value %q", label, err, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Group B — emitter + TOML writer
// ---------------------------------------------------------------------------

// TestEmitAllRoundTripsBodyVerbatim is AC-003's core: developer_instructions
// decodes byte-equal to the .md body; name equals the frontmatter name;
// description is non-empty. Parsed with the independent test-side decoder.
func TestEmitAllRoundTripsBodyVerbatim(t *testing.T) {
	man := mustManifest(t)
	set := fullFixtureSet()
	pub, err := agentemit.EmitAll(set, "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	wantBody := "\n# mdcarrier\n\nBody line.\n\n---\n\nAfter the rule.\n"
	tomlData, ok := pub.CodexTOML[".codex/agents/moai/mdcarrier.toml"]
	if !ok {
		t.Fatalf("emitted TOML path .codex/agents/moai/mdcarrier.toml missing; have %v", keysOf(pub.CodexTOML))
	}
	doc, err := decodeTOML(string(tomlData))
	if err != nil {
		t.Fatalf("independent decode failed: %v\n%s", err, tomlData)
	}
	if got := doc["developer_instructions"]; got != wantBody {
		t.Errorf("developer_instructions not byte-equal to .md body:\n got %q\nwant %q", got, wantBody)
	}
	if got := doc["name"]; got != "mdcarrier" {
		t.Errorf("name = %q, want mdcarrier", got)
	}
	if got := doc["description"]; got == "" {
		t.Error("description empty")
	}
}

// TestEmitAllMCPServerMapping is AC-007: exactly the agents carrying any
// mcp__moai__* token declare mcp_servers containing "moai"; the others carry
// no mcp_servers key at all.
func TestEmitAllMCPServerMapping(t *testing.T) {
	man := mustManifest(t)
	pub, err := agentemit.EmitAll(fullFixtureSet(), "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	carrier := mustDecoded(t, pub, ".codex/agents/moai/mdcarrier.toml")
	servers, ok := carrier["mcp_servers"].([]string)
	if !ok || len(servers) != 1 || servers[0] != "moai" {
		t.Errorf("carrier mcp_servers = %#v, want [\"moai\"]", carrier["mcp_servers"])
	}
	for _, plain := range []string{".codex/agents/moai/plainagent.toml", ".codex/agents/moai/twoskills.toml"} {
		doc := mustDecoded(t, pub, plain)
		if _, has := doc["mcp_servers"]; has {
			t.Errorf("%s: non-carrier must not declare mcp_servers", plain)
		}
	}
}

// TestEmitAllEffortMappingPerManifest is AC-008: each model_reasoning_effort
// equals the manifest-mapped value for its source effort.
func TestEmitAllEffortMappingPerManifest(t *testing.T) {
	man := mustManifest(t)
	pub, err := agentemit.EmitAll(fullFixtureSet(), "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	want := map[string]string{
		".codex/agents/moai/mdcarrier.toml":  "high",
		".codex/agents/moai/plainagent.toml": "medium",
		".codex/agents/moai/twoskills.toml":  "low",
	}
	for path, wantEffort := range want {
		doc := mustDecoded(t, pub, path)
		got, ok := doc["model_reasoning_effort"].(string)
		if !ok {
			t.Errorf("%s: model_reasoning_effort missing/not a string", path)
			continue
		}
		if got != wantEffort {
			t.Errorf("%s: model_reasoning_effort = %q, want %q", path, got, wantEffort)
		}
	}
}

// TestEmitAllOmitsModel is AC-009: zero emitted files carry a model key, even
// when the source .md pins a Claude alias (sonnet) — the pin is a documented
// drop in the manifest, never an emitted value.
func TestEmitAllOmitsModel(t *testing.T) {
	man := mustManifest(t)
	pub, err := agentemit.EmitAll(fullFixtureSet(), "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	for path := range pub.CodexTOML {
		doc := mustDecoded(t, pub, path)
		if _, has := doc["model"]; has {
			t.Errorf("%s: model key must be omitted (R-011)", path)
		}
	}
}

// TestEmitAllSandboxPerMeasuredSet locks the P-01 probe outcome: the
// manifest emits sandbox_mode = "workspace-write" (a member of the
// runtime-measured value set {read-only, workspace-write, danger-full-access},
// codex-cli 0.147.0) on every agent; and a manifest variant whose value set
// is unconfirmed (emit: false) omits the key entirely — the ship-omitted
// fallback, never a guess.
func TestEmitAllSandboxPerMeasuredSet(t *testing.T) {
	man := mustManifest(t)
	pub, err := agentemit.EmitAll(fullFixtureSet(), "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	for path := range pub.CodexTOML {
		doc := mustDecoded(t, pub, path)
		if got := doc["sandbox_mode"]; got != "workspace-write" {
			t.Errorf("%s: sandbox_mode = %#v, want workspace-write (P-01 measured set member)", path, doc["sandbox_mode"])
		}
	}

	// Ship-omitted fallback face: an unconfirmed manifest variant omits the key.
	man.Fields["sandbox_mode"].Emit = false
	pub, err = agentemit.EmitAll(fullFixtureSet(), "agents", man)
	if err != nil {
		t.Fatalf("EmitAll (omit variant): %v", err)
	}
	for path := range pub.CodexTOML {
		doc := mustDecoded(t, pub, path)
		if _, has := doc["sandbox_mode"]; has {
			t.Errorf("%s: sandbox_mode emitted from an unconfirmed manifest variant (ship-omitted rule)", path)
		}
	}
}

// TestEmitAllDeterministic is AC-004's in-process face: two emissions over
// identical inputs produce byte-identical artifacts, and no output embeds an
// environment-derived marker (absolute path, temp dir).
func TestEmitAllDeterministic(t *testing.T) {
	man := mustManifest(t)
	set := fullFixtureSet()
	pub1, err := agentemit.EmitAll(set, "agents", man)
	if err != nil {
		t.Fatalf("EmitAll #1: %v", err)
	}
	pub2, err := agentemit.EmitAll(set, "agents", man)
	if err != nil {
		t.Fatalf("EmitAll #2: %v", err)
	}
	for path, want := range pub1.CodexTOML {
		got, ok := pub2.CodexTOML[path]
		if !ok {
			t.Fatalf("%s missing in second emission", path)
		}
		if string(got) != string(want) {
			t.Errorf("%s not deterministic", path)
		}
	}
	for path, data := range pub1.CodexTOML {
		s := string(data)
		for _, marker := range []string{"/Users/", "/tmp/", "var/folders", "os.TempDir"} {
			if strings.Contains(s, marker) {
				t.Errorf("%s: output contains environment-derived marker %q (R-006)", path, marker)
			}
		}
	}
}

// TestEmitAllMarkdownIdentityIsPassThrough pins Option A: the .md publication
// is identity — emitted markdown bytes equal source bytes (the emitter never
// re-renders the neutral layer; R-002/R-003 by construction).
func TestEmitAllMarkdownIdentityIsPassThrough(t *testing.T) {
	man := mustManifest(t)
	set := fullFixtureSet()
	pub, err := agentemit.EmitAll(set, "agents", man)
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	for file, want := range set {
		if !strings.HasSuffix(file, ".md") {
			continue
		}
		got, ok := pub.Markdown[file]
		if !ok {
			t.Errorf("%s missing from markdown publication", file)
			continue
		}
		if string(got) != string(want.Data) {
			t.Errorf("%s: markdown publication must be byte-identical pass-through", file)
		}
	}
}

// TestEmitAllFailClosedNegatives drives every fail-closed validator (R-008)
// and asserts each error names the offending file AND token/value while the
// returned publication is nil (no partial artifact set).
func TestEmitAllFailClosedNegatives(t *testing.T) {
	base := func(name, tools, effort, body string) fstest.MapFS {
		return fstest.MapFS{
			"agents/" + name + ".md": &fstest.MapFile{Data: []byte(fixtureMD(
				name, "d.\n", tools, "inherit", effort, nil, body,
			))},
		}
	}
	cases := []struct {
		label string
		fset  fstest.MapFS
		want1 string // file name fragment
		want2 string // offending token/value
	}{
		{
			label: "unknown tool token (AC-005)",
			fset:  base("badtoken", "Read, Teleport, Bash", "low", "b\n"),
			want1: "badtoken.md", want2: "Teleport",
		},
		{
			label: "empty tool token from trailing comma",
			fset:  base("emptytok", "Read, Bash,", "low", "b\n"),
			want1: "emptytok.md", want2: "empty",
		},
		{
			label: "unmapped effort value (AC-006)",
			fset:  base("badeffort", "Read, Bash", "ultra", "b\n"),
			want1: "badeffort.md", want2: "ultra",
		},
		{
			label: "unrepresentable body (TOML literal delimiter)",
			fset:  base("delimbody", "Read, Bash", "low", "contains ''' delimiter\n"),
			want1: "delimbody.md", want2: "'''",
		},
	}
	for _, tc := range cases {
		pub, err := agentemit.EmitAll(tc.fset, "agents", mustManifest(t))
		if err == nil {
			t.Errorf("%s: want error, got nil", tc.label)
			continue
		}
		if pub != nil {
			t.Errorf("%s: must return no partial artifact set, got %#v", tc.label, pub)
		}
		msg := err.Error()
		if !strings.Contains(msg, tc.want1) {
			t.Errorf("%s: error %q must name the file %q", tc.label, msg, tc.want1)
		}
		if !strings.Contains(msg, tc.want2) {
			t.Errorf("%s: error %q must name the offending token/value %q", tc.label, msg, tc.want2)
		}
	}
}

// TestEmitAllFailClosedDuplicateName covers the D.2 duplicate-name edge: two
// .md files declaring the same frontmatter name are a Codex namespace
// collision — emission fails closed naming the name.
func TestEmitAllFailClosedDuplicateName(t *testing.T) {
	set := fstest.MapFS{
		"agents/dupA.md": &fstest.MapFile{Data: []byte(fixtureMD("duplicated", "a.\n", "Read", "inherit", "low", nil, "a\n"))},
		"agents/dupB.md": &fstest.MapFile{Data: []byte(fixtureMD("duplicated", "b.\n", "Read", "inherit", "low", nil, "b\n"))},
	}
	pub, err := agentemit.EmitAll(set, "agents", mustManifest(t))
	if err == nil {
		t.Fatal("want duplicate-name error, got nil")
	}
	if pub != nil {
		t.Error("must return no partial artifact set")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Errorf("error %q must name the duplicated agent name", err)
	}
}

// ---------------------------------------------------------------------------
// Group C — manifest (mapping-table authority)
// ---------------------------------------------------------------------------

// TestLoadManifestSelfValidates is AC-013's mechanical face: the embedded
// manifest parses, records the measured codex-cli version, carries a valid
// layout mode, gives every semantic class exactly one disposition with a
// non-empty rationale, and documents the manager-git model pin drop (AC-009).
func TestLoadManifestSelfValidates(t *testing.T) {
	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if man.CodexMeasuredVersion == "" {
		t.Error("codex_measured_version must be recorded (t91 pin)")
	}
	switch man.Layout.Mode {
	case "subdirectory", "flat_prefix":
	default:
		t.Errorf("layout mode %q invalid", man.Layout.Mode)
	}
	if len(man.Classes) == 0 {
		t.Fatal("classes empty")
	}
	valid := map[string]bool{
		"no-field": true, "consequence": true, "emit-field": true,
		"documented-drop": true, "deferred-m1": true, "omit": true,
		"correspondence-note": true,
	}
	for _, c := range man.Classes {
		if !valid[c.Disposition] {
			t.Errorf("class %s: invalid disposition %q", c.Class, c.Disposition)
		}
		if strings.TrimSpace(c.Rationale) == "" {
			t.Errorf("class %s: disposition %q carries no rationale (AC-013: no silent discard)", c.Class, c.Disposition)
		}
	}
	// manager-git sonnet pin must be a documented drop (AC-009).
	found := false
	for _, d := range man.DocumentedDrops {
		if strings.Contains(d.ID, "manager-git") && d.Rationale != "" {
			found = true
		}
	}
	if !found {
		t.Error("manager-git model pin documented drop missing or rationale-less")
	}
	// Every token class referenced by ToolClasses must have a disposition row.
	rows := map[string]bool{}
	for _, c := range man.Classes {
		rows[c.Class] = true
	}
	for token, class := range man.ToolClasses {
		if !rows[class] {
			t.Errorf("tool %s maps to class %s which has no disposition row", token, class)
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustManifest(t *testing.T) agentemit.Manifest {
	t.Helper()
	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	return man
}

func mustBytes(t *testing.T, fsys fstest.MapFS, path string) []byte {
	t.Helper()
	data, err := fsys.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func mustDecoded(t *testing.T, pub *agentemit.Publication, path string) map[string]any {
	t.Helper()
	data, ok := pub.CodexTOML[path]
	if !ok {
		t.Fatalf("%s missing from publication; have %v", path, keysOf(pub.CodexTOML))
	}
	doc, err := decodeTOML(string(data))
	if err != nil {
		t.Fatalf("independent decode of %s failed: %v\n%s", path, err, data)
	}
	return doc
}

func keysOf(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
