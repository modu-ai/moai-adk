// manifest.go — the four manifests (REQ-001, REQ-002, REQ-003).
//
// Field names, nesting and value shapes are copied from the moai-cowork
// marketplace precedent (Claude and Codex) and from mods/moai-board's plugin
// manifest; the generator invents no field. Struct field order is the emitted
// key order.
package pluginemit

import (
	"bytes"
	"encoding/json"
)

const (
	repoURL    = "https://github.com/modu-ai/moai-adk"
	authorName = "MoAI-ADK"
	license    = "Apache-2.0"

	marketplaceDescription = "MoAI-ADK marketplace: the moai core plugin, generated from the MoAI-ADK template tree. Register it with `claude plugin marketplace add modu-ai/moai-adk`."
	pluginDescription      = "MoAI-ADK core plugin: the core-tier skills, the /moai commands and the moai MCP server, generated from the MoAI-ADK template tree."
	shortDescription       = "MoAI-ADK core skills, commands and MCP server"
)

type nameOnly struct {
	Name string `json:"name"`
}

type claudeMarketplaceDoc struct {
	Name     string                 `json:"name"`
	Owner    nameOnly               `json:"owner"`
	Metadata claudeMarketplaceMeta  `json:"metadata"`
	Plugins  []claudeMarketplaceRef `json:"plugins"`
}

type claudeMarketplaceMeta struct {
	Description string `json:"description"`
	Version     string `json:"version"`
}

type claudeMarketplaceRef struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// codexMarketplaceDoc carries no version field at either level: the Codex
// marketplace shape has none (P-35); the Codex plugin version lives in the
// Codex plugin manifest.
type codexMarketplaceDoc struct {
	Name      string                `json:"name"`
	Interface displayNameOnly       `json:"interface"`
	Plugins   []codexMarketplaceRef `json:"plugins"`
}

type displayNameOnly struct {
	DisplayName string `json:"displayName"`
}

type codexMarketplaceRef struct {
	Name     string      `json:"name"`
	Source   codexSource `json:"source"`
	Policy   codexPolicy `json:"policy"`
	Category string      `json:"category"`
}

type codexSource struct {
	Source string `json:"source"`
	Path   string `json:"path"`
}

type codexPolicy struct {
	Installation   string `json:"installation"`
	Authentication string `json:"authentication"`
}

type claudePluginDoc struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      nameOnly `json:"author"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	License     string   `json:"license"`
}

type codexPluginDoc struct {
	claudePluginDoc
	Skills     string              `json:"skills"`
	MCPServers map[string]MCPEntry `json:"mcpServers"`
	Interface  codexInterface      `json:"interface"`
}

type codexInterface struct {
	DisplayName      string `json:"displayName"`
	ShortDescription string `json:"shortDescription"`
	LongDescription  string `json:"longDescription"`
	DeveloperName    string `json:"developerName"`
	Category         string `json:"category"`
	WebsiteURL       string `json:"websiteURL"`
}

func claudeMarketplace(ver string) claudeMarketplaceDoc {
	return claudeMarketplaceDoc{
		Name:     MarketplaceName,
		Owner:    nameOnly{Name: "modu-ai"},
		Metadata: claudeMarketplaceMeta{Description: marketplaceDescription, Version: ver},
		Plugins: []claudeMarketplaceRef{{
			Name:        PluginName,
			Source:      "./" + PluginRoot,
			Version:     ver,
			Description: pluginDescription,
			Category:    "development",
		}},
	}
}

func codexMarketplace() codexMarketplaceDoc {
	return codexMarketplaceDoc{
		Name:      MarketplaceName,
		Interface: displayNameOnly{DisplayName: authorName},
		Plugins: []codexMarketplaceRef{{
			Name:     PluginName,
			Source:   codexSource{Source: "local", Path: "./" + PluginRoot},
			Policy:   codexPolicy{Installation: "AVAILABLE", Authentication: "ON_INSTALL"},
			Category: "Coding",
		}},
	}
}

func claudePlugin(ver string) claudePluginDoc {
	return claudePluginDoc{
		Name:        PluginName,
		Version:     ver,
		Description: pluginDescription,
		Author:      nameOnly{Name: authorName},
		Homepage:    repoURL,
		Repository:  repoURL,
		License:     license,
	}
}

func codexPlugin(ver string, mcp MCPEntry) codexPluginDoc {
	return codexPluginDoc{
		claudePluginDoc: claudePlugin(ver),
		Skills:          "./skills/",
		MCPServers:      map[string]MCPEntry{PluginName: mcp},
		Interface: codexInterface{
			DisplayName:      authorName,
			ShortDescription: shortDescription,
			LongDescription:  pluginDescription,
			DeveloperName:    authorName,
			Category:         "Coding",
			WebsiteURL:       repoURL,
		},
	}
}

// marshal renders v as two-space-indented JSON with a trailing newline and no
// HTML escaping, so the committed bytes are readable and stable.
func marshal(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
