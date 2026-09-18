// Package codextools binds the hybrid native/late-dispatcher tool registry to a
// single authenticated Claude conversation. It does not execute tools or RPCs.
package codextools

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const DispatcherName = "moai_dispatch_tool"
const PlaceholderName = "DeferredToolPlaceholder"
const (
	maxTools         = 1024
	maxSchemaBytes   = 256 << 10
	maxSnapshotBytes = 8 << 20
	maxArgsBytes     = 1 << 20
	maxCalls         = 65536
)

var ErrInvalid = errors.New("invalid or unauthorized hybrid tool operation")

// ErrArguments is recoverable model input, but remains an ErrInvalid operation.
// It is returned only after owner, tool and current schema authority checks.
var ErrArguments = fmt.Errorf("tool arguments do not match the declared schema: %w", ErrInvalid)
var safeName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type Binding struct {
	ConversationID string
	ThreadID       string
	AccountScope   string
}
type Definition struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type,omitempty"`
	Deferred    bool            `json:"defer_loading,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}
type Reference struct {
	Type     string `json:"type"`
	ToolName string `json:"tool_name"`
}
type DynamicTool struct {
	Type         string          `json:"type"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"inputSchema"`
	DeferLoading bool            `json:"deferLoading"`
}
type Call struct {
	TurnID    string
	CallID    string
	Tool      string
	Arguments json.RawMessage
}
type Invocation struct {
	Name         string
	Arguments    json.RawMessage
	SchemaDigest string
	Epoch        uint64
}
type entry struct {
	definition Definition
	digest     string
	schema     *jsonschema.Schema
	native     bool
	alias      string
	epoch      uint64
}
type pending struct {
	turn   string
	name   string
	digest string
	done   bool
}
type Registry struct {
	mu      sync.Mutex
	owner   Binding
	tools   map[string]*entry
	aliases map[string]string
	native  []DynamicTool
	calls   map[string]pending
	epoch   uint64
}

type denyLoader struct{}

func (denyLoader) Load(string) (any, error) { return nil, ErrInvalid }

func prepare(def Definition) (*entry, error) {
	if def.Name == "" || len(def.Name) > 256 || len(def.Description) > 64<<10 || def.Type != "" && def.Type != "custom" {
		return nil, ErrInvalid
	}
	if def.Name == DispatcherName || strings.HasPrefix(def.Name, "moai_native_") {
		return nil, ErrInvalid
	}
	value, err := strictJSON(def.InputSchema, maxSchemaBytes)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, ErrInvalid
	}
	if def.Name == PlaceholderName {
		properties, valid := object["properties"].(map[string]any)
		if !def.Deferred || object["type"] != "object" || !valid || len(properties) != 0 || len(object) != 2 {
			return nil, ErrInvalid
		}
		return nil, nil
	}
	if dialect, exists := object["$schema"]; exists {
		switch dialect {
		case "http://json-schema.org/draft-07/schema#", "https://json-schema.org/draft/2019-09/schema", "https://json-schema.org/draft/2020-12/schema":
		default:
			return nil, ErrInvalid
		}
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseLoader(denyLoader{})
	const resource = "https://moai.invalid/tool-schema"
	if compiler.AddResource(resource, value) != nil {
		return nil, ErrInvalid
	}
	compiled, err := compiler.Compile(resource)
	if err != nil {
		return nil, ErrInvalid
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, ErrInvalid
	}
	def.InputSchema = canonical
	encoded, _ := json.Marshal(def)
	sum := sha256.Sum256(encoded)
	return &entry{definition: def, digest: hex.EncodeToString(sum[:]), schema: compiled}, nil
}
func prepareSnapshot(defs []Definition) (map[string]*entry, error) {
	if len(defs) > maxTools {
		return nil, ErrInvalid
	}
	result := map[string]*entry{}
	seen := map[string]bool{}
	size := 0
	for _, def := range defs {
		size += len(def.InputSchema) + len(def.Name) + len(def.Description)
		if size > maxSnapshotBytes || seen[def.Name] {
			return nil, ErrInvalid
		}
		seen[def.Name] = true
		e, err := prepare(def)
		if err != nil {
			return nil, err
		}
		if e != nil {
			result[def.Name] = e
		}
	}
	return result, nil
}

// New accepts definitions only after the caller authenticated the Claude request.
// Binding values must come from the owning session, never model-generated text.
// @MX:ANCHOR: [AUTO] Conversation-scoped hybrid registry construction.
// @MX:REASON: Native declarations, discovery and tool dispatch share the owner boundary.
func New(owner Binding, initial []Definition) (*Registry, error) {
	if owner.ConversationID == "" || owner.AccountScope == "" {
		return nil, ErrInvalid
	}
	tools, err := prepareSnapshot(initial)
	if err != nil {
		return nil, err
	}
	r := &Registry{owner: owner, tools: tools, aliases: map[string]string{}, calls: map[string]pending{}}
	for _, def := range initial {
		e := tools[def.Name]
		if e == nil {
			continue
		}
		alias := def.Name
		// App Server reserves its MCP namespace even for syntactically valid
		// names. Alias only the native declaration; Claude keeps its public ID.
		if !safeName.MatchString(alias) || alias == "mcp" || strings.HasPrefix(alias, "mcp__") {
			sum := sha256.Sum256([]byte(alias))
			alias = "moai_native_" + hex.EncodeToString(sum[:16])
		}
		if _, exists := r.aliases[alias]; exists {
			return nil, ErrInvalid
		}
		e.native = true
		e.alias = alias
		r.aliases[alias] = def.Name
		r.native = append(r.native, DynamicTool{Type: "function", Name: alias, Description: def.Description, InputSchema: append(json.RawMessage(nil), e.definition.InputSchema...)})
	}
	r.native = append(r.native, DynamicTool{Type: "function", Name: DispatcherName, Description: "Call a tool discovered by ToolSearch. Only newly discovered tools use this route; initial tools have their own native functions.", InputSchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"arguments":{"type":"object"}},"required":["name","arguments"],"additionalProperties":false}`)})
	return r, nil
}

// BindThread binds the server-allocated thread once, after NativeTools was used
// in thread/start. No discovery or invocation is authorized before this binding.
func (r *Registry) BindThread(threadID string) (Binding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.owner.ThreadID != "" || threadID == "" || len(threadID) > 256 {
		return Binding{}, ErrInvalid
	}
	r.owner.ThreadID = threadID
	return r.owner, nil
}

func (r *Registry) NativeTools() []DynamicTool {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := append([]DynamicTool(nil), r.native...)
	for i := range out {
		out[i].InputSchema = append(json.RawMessage(nil), out[i].InputSchema...)
	}
	return out
}

// Discover atomically joins typed references to the current authenticated request's
// complete definitions. It neither changes the fixed native list nor sends RPCs.
func (r *Registry) Discover(owner Binding, refs []Reference, current []Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.owner.ThreadID == "" || owner != r.owner || len(refs) == 0 || len(refs) > maxTools {
		return ErrInvalid
	}
	snapshot, err := prepareSnapshot(current)
	if err != nil {
		return err
	}
	for name, existing := range r.tools {
		if candidate := snapshot[name]; candidate != nil && candidate.digest != existing.digest {
			return ErrInvalid
		}
	}
	added := map[string]*entry{}
	seen := map[string]bool{}
	for _, ref := range refs {
		if ref.Type != "tool_reference" || seen[ref.ToolName] {
			return ErrInvalid
		}
		seen[ref.ToolName] = true
		candidate := snapshot[ref.ToolName]
		if candidate == nil {
			return ErrInvalid
		}
		if existing := r.tools[ref.ToolName]; existing != nil {
			if existing.digest != candidate.digest {
				return ErrInvalid
			}
			continue
		}
		added[ref.ToolName] = candidate
	}
	if len(r.tools)+len(added) > maxTools {
		return ErrInvalid
	}
	if len(added) == 0 {
		return nil
	}
	r.epoch++
	for name, e := range added {
		e.epoch = r.epoch
		r.tools[name] = e
	}
	return nil
}

// Begin validates the full schema and reserves a unique call before returning a
// Claude invocation. Returning an invocation does not itself execute the tool.
// @MX:ANCHOR: [AUTO] Native and dispatcher execution authorization boundary.
// @MX:REASON: Both routes must enforce identical schema and conversation checks.
func (r *Registry) Begin(owner Binding, call Call, current []Definition) (Invocation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.owner.ThreadID == "" || owner != r.owner || call.TurnID == "" || call.CallID == "" || len(call.TurnID) > 256 || len(call.CallID) > 256 || len(r.calls) >= maxCalls {
		return Invocation{}, ErrInvalid
	}
	if _, exists := r.calls[call.CallID]; exists {
		return Invocation{}, ErrInvalid
	}
	value, err := strictJSON(call.Arguments, maxArgsBytes)
	if err != nil {
		return Invocation{}, err
	}
	args, ok := value.(map[string]any)
	if !ok {
		return Invocation{}, ErrInvalid
	}
	var e *entry
	if call.Tool == DispatcherName {
		if len(args) != 2 {
			return Invocation{}, ErrInvalid
		}
		name, ok := args["name"].(string)
		if !ok {
			return Invocation{}, ErrInvalid
		}
		e = r.tools[name]
		if e == nil || e.native {
			return Invocation{}, ErrInvalid
		}
		args, ok = args["arguments"].(map[string]any)
		if !ok {
			return Invocation{}, ErrInvalid
		}
	} else {
		name := r.aliases[call.Tool]
		e = r.tools[name]
		if e == nil || !e.native {
			return Invocation{}, ErrInvalid
		}
	}
	snapshot, err := prepareSnapshot(current)
	if err != nil {
		return Invocation{}, err
	}
	active := snapshot[e.definition.Name]
	if active == nil || active.digest != e.digest {
		return Invocation{}, ErrInvalid
	}
	if e.schema.Validate(args) != nil {
		return Invocation{}, ErrArguments
	}
	canonical, _ := json.Marshal(args)
	r.calls[call.CallID] = pending{turn: call.TurnID, name: e.definition.Name, digest: e.digest}
	return Invocation{Name: e.definition.Name, Arguments: canonical, SchemaDigest: e.digest, Epoch: e.epoch}, nil
}

// Complete checks the current allowed schema and consumes the matching result
// once. It never consumes a foreign, mutated, unknown or previously consumed call.
func (r *Registry) Complete(owner Binding, turnID, callID string, current []Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.owner.ThreadID == "" || owner != r.owner {
		return ErrInvalid
	}
	p, ok := r.calls[callID]
	if !ok || p.done || p.turn != turnID {
		return ErrInvalid
	}
	snapshot, err := prepareSnapshot(current)
	if err != nil {
		return err
	}
	e := snapshot[p.name]
	if e == nil || e.digest != p.digest {
		return ErrInvalid
	}
	p.done = true
	r.calls[callID] = p
	return nil
}
