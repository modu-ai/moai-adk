// Package cli — `moai mcp` generic atomic-RMW entry-management CLI
// (SPEC-TREND-MCP-001 M2).
//
// mcp.go implements three subcommands — `moai mcp add`, `moai mcp remove`,
// `moai mcp list` — that manage third-party MCP entries in the user's
// `.mcp.json` (project scope) or `~/.claude.json` (user scope). The CLI is a
// THIN wrapper that reuses the SAME atomic-RMW seam the GLM tools CLI
// (`glm_tools.go:541`) uses: `mutateClaudeJSONAtomic` (flock + compare-retry +
// backup-before-publish + idempotent-skip). No fork, no second lock convention
// (REQ-TMC-008 / AP-TMC-002). Secret-bearing HTTP/GLM entries stay in
// `glm_tools.go`; this generic CLI handles secret-free entries only
// (REQ-TMC-009: every `--env KEY=VAL` VALUE MUST be a ${VAR} literal).
//
// Subagent boundary (C-HRA-008 / REQ-TMC-010): the CLI MUST NOT call the
// orchestrator-only user-interaction channel. Every interaction is positional
// args + `--flag` defaults + structured stderr errors.
//
// @MX:NOTE: [AUTO] SPEC-TREND-MCP-001 M2 — generic entry-management CLI reusing mutateClaudeJSONAtomic unchanged.
// @MX:SPEC: SPEC-TREND-MCP-001
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// mcpSecretLiteralRe is the secret-hygiene validator (REQ-TMC-009). Every
// `--env KEY=VAL` VALUE MUST match this regex — a single ${VAR} literal that
// the Claude Code runtime expands at load. Resolved secrets are NEVER
// serialized into the git-tracked .mcp.json (C3 / AP-TMC-003).
//
// Lowercase / kebab-case / nested-expansion forms are rejected on purpose: the
// runtime expands exactly one ${VAR} per value, and the env-var name MUST
// match the shell convention ([A-Z_][A-Z0-9_]*).
var mcpSecretLiteralRe = regexp.MustCompile(`^\$\{[A-Z_][A-Z0-9_]*\}$`)

// mcpAddArgs is the parsed flag set for `moai mcp add`. It is the test-facing
// argument shape consumed by moaiMcpAdd; the cobra RunE adapter only translates
// cobra flags → this struct.
type mcpAddArgs struct {
	name    string   // mcpServers key
	command string   // stdio command (PATH-resolved)
	args    []string // stdio command args
	env     []string // KEY=VAL env pairs (VAL must be a ${VAR} literal)
	typeArg string   // "stdio" (default) or "http"
	url     string   // http only
	headers []string // http only (JSON-encoded KEY=VAL pairs)
	scope   string   // "project" (default) or "user"
}

// newMCPCmd builds the `moai mcp` parent command with three subcommands. It is
// registered under root.go AddCommand alongside the other subcommands.
func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP server entries in .mcp.json (project) or ~/.claude.json (user)",
		Long: `Manage third-party MCP server entries via the atomic-RMW guard.

Subcommands:
  add     Register a new MCP entry (idempotent; rejects resolved secrets).
  remove  Remove a named MCP entry (partial-delete safe; preserves unrelated entries).
  list    List active MCP entries (JSON when --json, plain text otherwise).

Every write goes through the SAME atomic-RMW seam the GLM tools CLI uses
(flock + compare-retry + backup-before-publish). No hand-editing of .mcp.json.
The --env KEY=VAL VALUE MUST be a ${VAR} literal; resolved secrets are rejected.`,
		SilenceUsage: true,
	}
	cmd.AddCommand(newMCPCmdAdd())
	cmd.AddCommand(newMCPCmdRemove())
	cmd.AddCommand(newMCPCmdList())
	return cmd
}

// newMCPCmdAdd builds `moai mcp add`.
func newMCPCmdAdd() *cobra.Command {
	var (
		command string
		args    []string
		env     []string
		typeArg string
		url     string
		headers []string
		scope   string
	)
	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Register a new MCP entry via the atomic-RMW guard",
		Long: `Register a new MCP entry in .mcp.json (project scope, default) or
~/.claude.json (user scope).

stdio entry (default):
	moai mcp add my-tool --command npx --args -y --args my-tool-mcp

http entry:
	moai mcp add semgrep --type http --url https://semgrep.example.com/mcp

Every --env KEY=VAL VALUE MUST be a ${VAR} literal; resolved secrets are
rejected (REQ-TMC-009). Reuses mutateClaudeJSONAtomic unchanged (REQ-TMC-008).`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, positional []string) error {
			if len(positional) == 0 {
				return fmt.Errorf("add requires a <name> argument")
			}
			dir, err := resolveMcpScopeDir(scope)
			if err != nil {
				return err
			}
			return moaiMcpAdd(dir, mcpAddArgs{
				name:    positional[0],
				command: command,
				args:    args,
				env:     env,
				typeArg: typeArg,
				url:     url,
				headers: headers,
				scope:   scope,
			})
		},
	}
	c.Flags().StringVar(&command, "command", "", "stdio command (PATH-resolved, no absolute path required)")
	c.Flags().StringArrayVar(&args, "args", nil, "stdio command args (repeatable, one token per flag)")
	c.Flags().StringArrayVar(&env, "env", nil, "env pair KEY=${VAR} (repeatable; VALUE must be a ${VAR} literal)")
	c.Flags().StringVar(&typeArg, "type", "stdio", "entry type: stdio (default) or http")
	c.Flags().StringVar(&url, "url", "", "http entry URL (required when --type http)")
	c.Flags().StringArrayVar(&headers, "headers", nil, "http entry header pair KEY=VAL (repeatable; VAL should be a ${VAR} literal)")
	c.Flags().StringVar(&scope, "scope", "project", "scope: project (default — .mcp.json) or user (~/.claude.json)")
	return c
}

// newMCPCmdRemove builds `moai mcp remove`.
func newMCPCmdRemove() *cobra.Command {
	var scope string
	c := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a named MCP entry (partial-delete safe)",
		Long: `Remove only the named entry; every unrelated mcpServers entry is
preserved via the SAME partial-delete safety contract the GLM tools CLI uses.`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, positional []string) error {
			if len(positional) == 0 {
				return fmt.Errorf("remove requires a <name> argument")
			}
			dir, err := resolveMcpScopeDir(scope)
			if err != nil {
				return err
			}
			return moaiMcpRemove(dir, positional[0], scope)
		},
	}
	c.Flags().StringVar(&scope, "scope", "project", "scope: project (default) or user")
	return c
}

// newMCPCmdList builds `moai mcp list`.
func newMCPCmdList() *cobra.Command {
	var (
		scope  string
		asJSON bool
	)
	c := &cobra.Command{
		Use:   "list",
		Short: "List active MCP entries (JSON when --json)",
		Long: `List the active mcpServers entries, distinguishing stdio from http
and flagging entries that carry ${VAR} literal env references.`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			dir, err := resolveMcpScopeDir(scope)
			if err != nil {
				return err
			}
			return moaiMcpList(c.OutOrStdout(), dir, scope, asJSON)
		},
	}
	c.Flags().StringVar(&scope, "scope", "project", "scope: project (default) or user")
	c.Flags().BoolVar(&asJSON, "json", false, "emit JSON instead of plain text")
	return c
}

// resolveMcpScopeDir returns the directory the scope-relative .mcp.json (or
// ~/.claude.json) lives under. moaiMcpAdd/mcpRemove/mcpList join the basename
// themselves, so this returns the dir only.
func resolveMcpScopeDir(scope string) (string, error) {
	// resolveConfigPath returns the full path; we split it back to a directory
	// so the moaiMcp* helpers can join a test-friendly relative path. This
	// keeps the path-resolution policy in exactly one place (resolveConfigPath).
	path, err := resolveConfigPath(scope)
	if err != nil {
		return "", err
	}
	// The project scope returns <cwd>/.mcp.json; user scope returns
	// <home>/.claude.json. The file basename differs, but the helpers consume
	// the directory + scope to re-resolve, so just return the dir.
	return filepathDir(path), nil
}

// filepathDir is a small wrapper kept local so the helpers do not pull in
// filepath explicitly.
func filepathDir(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		if i == 0 {
			return "/"
		}
		return p[:i]
	}
	return "."
}

// configPathForScope joins the scope-relative config filename under dir.
func configPathForScope(dir, scope string) string {
	if scope == "user" {
		return dir + "/.claude.json"
	}
	return dir + "/.mcp.json"
}

// moaiMcpAdd registers a new entry via mutateClaudeJSONAtomic. It is the
// test-facing seam: tests call it directly with a fixture dir + scope so
// resolveConfigPath does NOT depend on os.Getwd at test time.
func moaiMcpAdd(dir string, a mcpAddArgs) error {
	if a.name == "" {
		return fmt.Errorf("add requires a non-empty <name>")
	}
	// Validate --env values BEFORE entering the RMW (fail fast, no partial write).
	envMap, err := parseEnvFlags(a.env)
	if err != nil {
		return err
	}
	entry, err := buildMcpEntry(a)
	if err != nil {
		return err
	}
	configPath := configPathForScope(dir, a.scope)
	return mutateClaudeJSONAtomic(configPath, func(root map[string]any) (bool, error) {
		servers, _ := root["mcpServers"].(map[string]any)
		if servers == nil {
			servers = map[string]any{}
		}
		if existing, ok := servers[a.name].(map[string]any); ok && mcpEntryEqual(existing, entry) {
			root["mcpServers"] = servers
			return false, nil // idempotent skip — no write, no backup
		}
		servers[a.name] = entry
		root["mcpServers"] = servers
		// Touch envMap so the compiler does not complain on the unused-variable
		// path when buildMcpEntry already inlined it; keep the validator
		// visible at the call chain entry for fail-fast behavior.
		_ = envMap
		return true, nil
	})
}

// moaiMcpRemove removes a named entry via mutateClaudeJSONAtomic. Partial-delete
// safety is structural: only the named key is deleted, every other key in
// root["mcpServers"] (and every top-level key) is preserved.
func moaiMcpRemove(dir, name, scope string) error {
	if name == "" {
		return fmt.Errorf("remove requires a non-empty <name>")
	}
	configPath := configPathForScope(dir, scope)
	return mutateClaudeJSONAtomic(configPath, func(root map[string]any) (bool, error) {
		servers, _ := root["mcpServers"].(map[string]any)
		if servers == nil {
			return false, nil // nothing to remove — no-op, no backup
		}
		if _, ok := servers[name]; !ok {
			return false, nil // missing entry — no-op
		}
		delete(servers, name)
		root["mcpServers"] = servers
		return true, nil
	})
}

// mcpListEntry is the JSON-list element emitted by moaiMcpList when --json.
type mcpListEntry struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	EnvRefs []string       `json:"env_refs,omitempty"`
	Entry   map[string]any `json:"entry"`
}

// mcpListDocument is the JSON document emitted by moaiMcpList when --json.
type mcpListDocument struct {
	Scope   string         `json:"scope"`
	Count   int            `json:"count"`
	Entries []mcpListEntry `json:"entries"`
}

// moaiMcpList emits the active entries. JSON when asJSON; plain text otherwise.
// Reads the file directly (no RMW needed for a read).
func moaiMcpList(out io.Writer, dir, scope string, asJSON bool) error {
	configPath := configPathForScope(dir, scope)
	root, err := readClaudeJSON(configPath)
	if err != nil {
		return err
	}
	servers, _ := root["mcpServers"].(map[string]any)
	names := make([]string, 0, len(servers))
	for k := range servers {
		names = append(names, k)
	}
	sort.Strings(names)

	entries := make([]mcpListEntry, 0, len(names))
	for _, name := range names {
		raw, _ := servers[name].(map[string]any)
		entries = append(entries, mcpListEntry{
			Name:    name,
			Type:    classifyMcpType(raw),
			EnvRefs: collectEnvRefs(raw),
			Entry:   raw,
		})
	}

	if asJSON {
		doc := mcpListDocument{Scope: scope, Count: len(entries), Entries: entries}
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal list: %w", err)
		}
		_, _ = out.Write(data)
		_, _ = out.Write([]byte("\n"))
		return nil
	}
	// plain text — table-ish
	var b bytes.Buffer
	fmt.Fprintf(&b, "scope: %s\n", scope)
	fmt.Fprintf(&b, "count: %d\n", len(entries))
	for _, e := range entries {
		fmt.Fprintf(&b, "- %s\t(%s)\n", e.Name, e.Type)
		if len(e.EnvRefs) > 0 {
			fmt.Fprintf(&b, "  env_refs: %s\n", strings.Join(e.EnvRefs, ", "))
		}
	}
	_, _ = out.Write(b.Bytes())
	return nil
}

// classifyMcpType returns "http" or "stdio" based on the entry shape. An
// explicit "type" field wins; absence + a "url" field also implies http.
func classifyMcpType(entry map[string]any) string {
	if t, ok := entry["type"].(string); ok && t != "" {
		return t
	}
	if _, ok := entry["url"]; ok {
		return "http"
	}
	return "stdio"
}

// collectEnvRefs scans an entry's env block AND headers for ${VAR} literal
// references and returns the unique set (sorted). The Claude Code runtime
// expands these at load; surfacing them in `list` lets the user see at a glance
// which entries need runtime secret expansion.
func collectEnvRefs(entry map[string]any) []string {
	if entry == nil {
		return nil
	}
	seen := map[string]struct{}{}
	scan := func(v string) {
		// match ${VAR} where VAR is uppercase identifier
		re := regexp.MustCompile(`\$\{[A-Z_][A-Z0-9_]*\}`)
		for _, m := range re.FindAllString(v, -1) {
			seen[m] = struct{}{}
		}
	}
	if env, ok := entry["env"].(map[string]any); ok {
		for _, v := range env {
			if s, ok := v.(string); ok {
				scan(s)
			}
		}
	}
	if headers, ok := entry["headers"].(map[string]any); ok {
		for _, v := range headers {
			if s, ok := v.(string); ok {
				scan(s)
			}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// parseEnvFlags parses --env KEY=VAL pairs and validates VAL is a ${VAR} literal.
// Returns a map suitable for direct insertion into the entry's env block.
func parseEnvFlags(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		idx := strings.IndexByte(p, '=')
		if idx <= 0 {
			return nil, fmt.Errorf("invalid --env %q (expected KEY=${VAR})", p)
		}
		key := p[:idx]
		val := p[idx+1:]
		if !mcpSecretLiteralRe.MatchString(val) {
			return nil, fmt.Errorf("invalid --env %q: VALUE must be a single ${VAR} literal (e.g. API_KEY=${MY_API_KEY}); resolved secrets are rejected to avoid git-tracked credential leaks", p)
		}
		out[key] = val
	}
	return out, nil
}

// buildMcpEntry constructs the entry map for `add` based on --type.
func buildMcpEntry(a mcpAddArgs) (map[string]any, error) {
	envMap, err := parseEnvFlags(a.env)
	if err != nil {
		return nil, err
	}
	switch a.typeArg {
	case "", "stdio":
		if a.command == "" {
			return nil, fmt.Errorf("stdio entry requires --command (got empty)")
		}
		entry := map[string]any{
			"command": a.command,
			"args":    a.args,
		}
		if len(envMap) > 0 {
			entry["env"] = envMap
		}
		return entry, nil
	case "http":
		if a.url == "" {
			return nil, fmt.Errorf("http entry requires --url (got empty)")
		}
		entry := map[string]any{
			"type": "http",
			"url":  a.url,
		}
		if len(a.headers) > 0 {
			hdrs, err := parseHeaderFlags(a.headers)
			if err != nil {
				return nil, err
			}
			entry["headers"] = hdrs
		}
		return entry, nil
	default:
		return nil, fmt.Errorf("unknown --type %q (want stdio or http)", a.typeArg)
	}
}

// parseHeaderFlags parses --headers KEY=VAL pairs. Headers are NOT subject to
// the ${VAR}-literal validator (some headers are static like
// "Content-Type: application/json"), but a VALUE containing ${VAR} is left
// intact for the runtime to expand.
func parseHeaderFlags(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		idx := strings.IndexByte(p, '=')
		if idx <= 0 {
			return nil, fmt.Errorf("invalid --headers %q (expected KEY=VAL)", p)
		}
		out[p[:idx]] = p[idx+1:]
	}
	return out, nil
}
