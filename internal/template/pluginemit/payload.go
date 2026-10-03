// payload.go — the plugin payload, derived from the template tree (REQ-004 to
// REQ-008).
//
// The payload is an allow-list: the skills and the commands of the tier view,
// and one .mcp.json. A new scaffold-only directory (rules, hooks, settings,
// output styles, workflows) cannot leak into it, because nothing but those two
// source directories is read. No component name is held here: the skills are
// whatever the tier view lists, the commands whatever its command directory
// holds, so a rename in the template tree changes the payload with no edit.
package pluginemit

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/modu-ai/moai-adk/internal/template"
)

const (
	// skillsSource and commandsSource are the template-tree directories the
	// payload is read from; skillsDest and commandsDest their plugin-root
	// counterparts. Commands are laid out flat: the runtime inventory of an
	// installed plugin counts none of a nested commands/<dir>/ (P-30).
	skillsSource   = ".claude/skills"
	commandsSource = ".claude/commands/moai"
	skillsDest     = PluginRoot + "/skills"
	commandsDest   = PluginRoot + "/commands"

	// MCPPayloadPath is the payload file that declares the plugin's MCP server.
	MCPPayloadPath = PluginRoot + "/.mcp.json"

	templateSuffix = ".tmpl"
	commandSuffix  = ".md"
)

// mcpDoc is the payload .mcp.json: only the moai server, with the command and
// args the template carries.
type mcpDoc struct {
	MCPServers map[string]MCPEntry `json:"mcpServers"`
}

// payload derives the payload files from the tier view: skills copied byte for
// byte, commands copied or rendered with the template default context (the
// English variant) and placed flat, and the .mcp.json. A source directory the
// view does not have contributes nothing.
func payload(view fs.FS, mcp MCPEntry) (map[string][]byte, error) {
	out := map[string][]byte{}
	renderer := template.NewRenderer(view)

	// read returns the payload bytes of one source file: rendered when the
	// source is a .tmpl file, copied unchanged otherwise.
	read := func(src string) ([]byte, error) {
		if strings.HasSuffix(src, templateSuffix) {
			data, err := renderer.Render(src, template.NewTemplateContext())
			if err != nil {
				return nil, fmt.Errorf("pluginemit: render %s: %w", src, err)
			}
			return data, nil
		}
		data, err := fs.ReadFile(view, src)
		if err != nil {
			return nil, fmt.Errorf("pluginemit: read %s: %w", src, err)
		}
		return data, nil
	}

	if _, err := fs.Stat(view, skillsSource); err == nil {
		err = fs.WalkDir(view, skillsSource, func(src string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := read(src)
			if err != nil {
				return err
			}
			rel := strings.TrimSuffix(strings.TrimPrefix(src, skillsSource+"/"), templateSuffix)
			out[path.Join(skillsDest, rel)] = data
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("pluginemit: walk %s: %w", skillsSource, err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("pluginemit: stat %s: %w", skillsSource, err)
	}

	entries, err := fs.ReadDir(view, commandsSource)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("pluginemit: read %s: %w", commandsSource, err)
	}
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), templateSuffix)
		if e.IsDir() || !strings.HasSuffix(name, commandSuffix) {
			continue
		}
		data, err := read(path.Join(commandsSource, e.Name()))
		if err != nil {
			return nil, err
		}
		out[path.Join(commandsDest, name)] = data
	}

	mcpData, err := marshal(mcpDoc{MCPServers: map[string]MCPEntry{PluginName: mcp}})
	if err != nil {
		return nil, fmt.Errorf("pluginemit: marshal %s: %w", MCPPayloadPath, err)
	}
	out[MCPPayloadPath] = mcpData
	return out, nil
}
