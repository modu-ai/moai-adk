package translate

import (
	"errors"
	"reflect"
	"strings"
)

// PolicyProfile is an explicit, versioned policy subset, never a model capability.
// The adapter binds it to a provider, authentication method and fixed endpoint.
type PolicyProfile string

const (
	PolicyGPTNative       PolicyProfile = "claude-code-2.1.268-gpt"
	PolicyAnthropicNative PolicyProfile = "claude-code-2.1.268-anthropic"
)

// NativePolicy describes validated syntax. KeepAll and UserID still require the
// caller's receipt/UUID authorization; syntax alone never authorizes history.
type NativePolicy struct {
	High, Title, KeepAll bool
	// Effort is the explicitly validated GPT effort; an absent effort stays absent.
	Effort string
	// Display records the validated native thinking presentation. The GPT
	// Responses projection maps omitted to no reasoning summary and summarized
	// to the auto reasoning summary; other values fail closed.
	Display string
	UserID  string
}

// gptEffortAllowlist is the union of every effort the official OpenAI model
// pages list for a catalog GPT model; the per-model split lives in
// gptModelEffortAllowed. History: card t695 D2 (2026-09-13) verified
// low..max by real request against the subscription upstream and recorded
// gpt-6-astra rejecting max; card t841 (2026-09-14) re-aligned the set with
// the official model pages, which now govern: none joins the union, and the
// astra max clamp is withdrawn. The Anthropic profile keeps its previously
// validated {high} subset; no OpenAI evidence applies to it.
func gptEffortAllowlist(profile PolicyProfile, value string) bool {
	if profile != PolicyGPTNative {
		return value == "high"
	}
	switch value {
	case "none", "low", "medium", "high", "xhigh", "max":
		return true
	}
	return false
}

// gptModelEffortAllowed is the per-model effort set from the official model
// pages (developers.openai.com/api/docs/models/<model>, fetched 2026-09-14):
// gpt-5.6-sol, gpt-5.6-terra and gpt-5.6-luna list none, low, medium
// (default), high, xhigh and max; gpt-6-astra lists low, medium, high, xhigh
// and max, so none fails closed there. An accepted value is projected as-is.
func gptModelEffortAllowed(model, value string) bool {
	if model == "gpt-6-astra" && value == "none" {
		return false
	}
	return gptEffortAllowlist(PolicyGPTNative, value)
}

// ValidateNativePolicy validates only policy fields of a strict JSON request.
// It does not validate public messages or authenticate a native signature.
func ValidateNativePolicy(body []byte, profile PolicyProfile) (NativePolicy, error) {
	root, e := objectJSON(body)
	if e != nil {
		return NativePolicy{}, e
	}
	return nativePolicy(root, profile)
}
func nativePolicy(root map[string]any, profile PolicyProfile) (p NativePolicy, e error) {
	if profile != PolicyGPTNative && profile != PolicyAnthropicNative {
		return p, errors.New("unsupported native policy profile")
	}
	if v, ok := root["output_config"]; ok {
		m, ok := v.(map[string]any)
		if !ok || len(m) == 0 || keys(m, "effort", "format") != nil {
			return p, errors.New("invalid output_config")
		}
		if effort, ok := m["effort"]; ok {
			value, valid := effort.(string)
			if !valid || !gptEffortAllowlist(profile, value) {
				return p, errors.New("unsupported effort")
			}
			p.High = value == "high"
			p.Effort = value
		}
		if v, ok := m["format"]; ok {
			// GPT: any json_schema format is forwarded; the subscription upstream
			// validates the schema itself (card t695 D2). Response-side title
			// validation stays tied to the exact title-only schema, where the
			// output contract is verified. Anthropic keeps the exact-schema rule:
			// its native body is forwarded verbatim without D2 evidence.
			f, ok := v.(map[string]any)
			if !ok || len(f) != 2 || f["type"] != "json_schema" || keys(f, "type", "schema") != nil {
				return p, errors.New("invalid output format")
			}
			if _, ok := f["schema"].(map[string]any); !ok {
				return p, errors.New("invalid output format")
			}
			titleSchema, _ := objectJSON([]byte(`{"type":"object","properties":{"title":{"type":"string"}},"required":["title"],"additionalProperties":false}`))
			if profile != PolicyGPTNative {
				if !reflect.DeepEqual(f["schema"], titleSchema) {
					return p, errors.New("unsupported output schema")
				}
				p.Title = true
			} else if reflect.DeepEqual(f["schema"], titleSchema) {
				p.Title = true
			}
		}
	}
	if v, ok := root["thinking"]; ok {
		m, ok := v.(map[string]any)
		if !ok || len(m) < 1 || len(m) > 2 || keys(m, "type", "display") != nil {
			return p, errors.New("invalid thinking policy")
		}
		typ, ok := m["type"].(string)
		if !ok {
			return p, errors.New("invalid thinking policy")
		}
		if display, ok := m["display"]; ok {
			value, valid := display.(string)
			if !valid || value != "omitted" && value != "summarized" {
				return p, errors.New("invalid thinking display")
			}
			if typ == "disabled" {
				return p, errors.New("thinking display conflicts with disabled policy")
			}
			p.Display = value
		}
		switch typ {
		case "adaptive":
			if profile == PolicyGPTNative {
				// Adaptive reasoning is upstream-verified with every allowlisted
				// effort (card t695 D2); it only requires a validated one.
				if p.Effort == "" {
					return p, errors.New("adaptive requires a validated effort")
				}
			} else if !p.High {
				return p, errors.New("adaptive requires validated high profile")
			}
		case "disabled":
			if profile != PolicyAnthropicNative {
				return p, errors.New("unverified disabled mapping")
			}
		default:
			return p, errors.New("unsupported thinking policy")
		}
	}
	if v, ok := root["context_management"]; ok {
		expected, _ := objectJSON([]byte(`{"edits":[{"type":"clear_thinking_20251015","keep":"all"}]}`))
		if !reflect.DeepEqual(v, expected) {
			return p, errors.New("unsupported context policy")
		}
		p.KeepAll = true
	}
	if v, ok := root["metadata"]; ok {
		m, ok := v.(map[string]any)
		if !ok || len(m) != 1 || keys(m, "user_id") != nil {
			return p, errors.New("unsupported native metadata")
		}
		p.UserID, e = stringField(m, "user_id")
		if e != nil {
			return p, e
		}
		id, err := objectJSON([]byte(p.UserID))
		if err != nil || len(id) != 3 || keys(id, "account_uuid", "device_id", "session_id") != nil {
			return p, errors.New("invalid native user_id")
		}
		for _, key := range []string{"account_uuid", "device_id", "session_id"} {
			if key == "account_uuid" {
				_, err = stringFieldAllowEmpty(id, key)
			} else {
				_, err = stringField(id, key)
			}
			if err != nil {
				return p, errors.New("invalid native identity field")
			}
		}
	}
	return p, nil
}
func (p NativePolicy) applyGPT(out map[string]any, root map[string]any) {
	if p.Effort != "" {
		reasoning := map[string]any{"effort": p.Effort}
		if p.Display == "summarized" {
			reasoning["summary"] = "auto"
		}
		out["reasoning"] = reasoning
	}
	if v, ok := root["output_config"]; ok {
		if m, ok := v.(map[string]any); ok {
			if f, ok := m["format"].(map[string]any); ok {
				schema, _ := f["schema"].(map[string]any)
				out["text"] = map[string]any{"format": map[string]any{"type": "json_schema", "name": "moai_native_output", "strict": true, "schema": schema}}
			}
		}
	}
}
func validateTitle(blocks []any, stop string) error {
	if stop != "end_turn" {
		return errors.New("title did not complete")
	}
	var text strings.Builder
	for _, value := range blocks {
		b, ok := value.(map[string]any)
		if !ok || b["type"] != "text" {
			return errors.New("title contains nontext output")
		}
		s, ok := b["text"].(string)
		if !ok {
			return errors.New("invalid title text")
		}
		text.WriteString(s)
	}
	m, e := objectJSON([]byte(text.String()))
	if e != nil || len(m) != 1 || keys(m, "title") != nil {
		return errors.New("title schema mismatch")
	}
	if _, e = stringFieldAllowEmpty(m, "title"); e != nil {
		return errors.New("title must be string")
	}
	return nil
}
