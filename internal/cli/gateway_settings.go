package cli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

// cleanupGatewaySettings removes legacy routing state through the existing
// locked atomic mutation seam. Credentials and gateway URLs are never written.
// The sole restoration is the pre-existing backup required by REQ-MG-021.
func cleanupGatewaySettings(path string) error {
	return mutateSettingsLocal(path, func(doc map[string]any) {
		if env, ok := doc["env"].(map[string]any); ok {
			backup, _ := env["MOAI_BACKUP_AUTH_TOKEN"].(string)
			for _, key := range gatewayScrubKeys() {
				delete(env, key)
			}
			if backup != "" {
				env[config.EnvAnthropicAuthToken] = backup
			}
			if len(env) == 0 {
				delete(doc, "env")
			}
		}
	})
}

// prepareGatewayOverlay is a gated compatibility candidate: settings.env is
// applied in memory, while only non-env settings are materialized. Actual Claude
// source precedence must be verified before activating this launch binding.
// @MX:WARN: [AUTO] Settings input is bounded and environment extraction precedes materialization.
// @MX:REASON: Argument boundaries, duplicate JSON keys and secret-bearing env must remain distinct during merging.
func prepareGatewayOverlay(args, inherited []string, overlay map[string]any) ([]byte, []string, []string, error) {
	doc := map[string]any{}
	remaining := make([]string, 0, len(args))
	env := append([]string(nil), inherited...)
	seen := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			remaining = append(remaining, args[i:]...)
			break
		}
		if args[i] == "--fallback-model" || strings.HasPrefix(args[i], "--fallback-model=") {
			return nil, nil, nil, errors.New("--fallback-model is unsupported in gateway mode; select a model explicitly")
		}
		value := ""
		isSettings := false
		if args[i] == "--settings" {
			if i+1 >= len(args) {
				return nil, nil, nil, errors.New("--settings requires a value")
			}
			i++
			value = args[i]
			isSettings = true
		} else if strings.HasPrefix(args[i], "--settings=") {
			value = strings.TrimPrefix(args[i], "--settings=")
			isSettings = true
		}
		if !isSettings {
			remaining = append(remaining, args[i])
			continue
		}
		if seen {
			return nil, nil, nil, errors.New("gateway launch requires a single --settings source")
		}
		seen = true
		data := []byte(value)
		if !strings.HasPrefix(strings.TrimSpace(value), "{") {
			file, err := os.Open(value)
			if err != nil {
				return nil, nil, nil, errors.New("cannot read --settings source")
			}
			data, err = io.ReadAll(io.LimitReader(file, gateway.MaxChildConfigBytes+1))
			_ = file.Close()
			if err != nil {
				return nil, nil, nil, errors.New("cannot read --settings source")
			}
		}
		if len(data) > gateway.MaxChildConfigBytes || gateway.ValidateJSONObject(data) != nil {
			return nil, nil, nil, errors.New("--settings requires an unambiguous bounded JSON object")
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, nil, nil, errors.New("invalid --settings object")
		}
	}
	if raw, exists := doc["env"]; exists {
		values, ok := raw.(map[string]any)
		if !ok {
			return nil, nil, nil, errors.New("--settings env must be a string map")
		}
		keys := make([]string, 0, len(values))
		for key, value := range values {
			str, ok := value.(string)
			if !ok || key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(str, 0) {
				return nil, nil, nil, errors.New("--settings env must contain valid string entries")
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		filtered := env[:0]
		for _, item := range env {
			key, _, _ := strings.Cut(item, "=")
			if _, overridden := values[key]; !overridden {
				filtered = append(filtered, item)
			}
		}
		env = filtered
		for _, key := range keys {
			env = append(env, key+"="+values[key].(string))
		}
		delete(doc, "env")
	}
	for key, value := range overlay {
		if key == "env" {
			return nil, nil, nil, errors.New("gateway overlay cannot materialize environment")
		}
		doc[key] = value
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, nil, nil, errors.New("cannot encode gateway settings")
	}
	return data, env, remaining, nil
}
