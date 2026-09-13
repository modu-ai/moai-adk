package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/profile"
)

type gatewayStartedChild struct {
	Address, OverlayPath string
	Stop                 func()
}
type gatewaySessionOptions struct {
	Mode       string
	Catalog    gateway.CatalogSnapshot
	GLM        config.GLMModels
	GLMKey     string
	SessionEnv []string
	Child      gateway.StartOptions
	Overlay    map[string]any
	// Family is optional until the caller has completed the private-state
	// bootstrap. If it is present, every launch gets an explicit descriptor.
	Family *conversation.Manager
	// Payload binds the private child configuration to the descriptor selected
	// before child startup. It is required for production native sessions.
	Payload func(conversation.Descriptor) (json.RawMessage, error)
	Start   func(context.Context, gateway.StartOptions) (gatewayStartedChild, error)
}

// newGatewaySessionBinding is called only by an approved transport binding.
// It does not select an authentication header or enable production by itself.
func newGatewaySessionBinding(options gatewaySessionOptions) *gatewayLaunchBinding {
	start := options.Start
	if start == nil {
		start = func(ctx context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
			child, err := gateway.StartChild(ctx, opts)
			if err != nil {
				return gatewayStartedChild{}, err
			}
			return gatewayStartedChild{Address: child.Address, OverlayPath: child.OverlayPath, Stop: func() {
				bounded, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = child.Stop(bounded)
			}}, nil
		}
	}
	return &gatewayLaunchBinding{Mode: options.Mode, Prepare: func(in gatewayLaunchRequest) (gateway.LaunchPlan, func(), error) {
		if in.Mode != options.Mode {
			return gateway.LaunchPlan{}, nil, errors.New("gateway launcher mode mismatch")
		}
		var descriptor conversation.Descriptor
		if options.Family != nil {
			var err error
			descriptor, err = prepareGatewayConversation(in, options.Family)
			if err != nil {
				return gateway.LaunchPlan{}, nil, err
			}
			if err := shareGatewayPeerRegistry(descriptor.ConfigDir, in); err != nil {
				return gateway.LaunchPlan{}, nil, err
			}
			if err := seedGatewayUIState(descriptor.ConfigDir, in); err != nil {
				return gateway.LaunchPlan{}, nil, err
			}
			if err := seedGatewayBypassAcceptance(descriptor.ConfigDir, in); err != nil {
				return gateway.LaunchPlan{}, nil, err
			}
			if in.ExplicitModel == "" && descriptor.Model != "" {
				in.ExplicitModel = descriptor.Model
			}
			if options.Payload != nil {
				payload, err := options.Payload(descriptor)
				if err != nil {
					return gateway.LaunchPlan{}, nil, err
				}
				options.Child.Config.Payload = payload
			}
		}
		if _, _, _, err := prepareGatewayOverlay(in.Args, in.Inherited, nil); err != nil {
			return gateway.LaunchPlan{}, nil, err
		}
		// Resolve the provider contract before allocating a child. The actual
		// bound address is validated again after startup.
		preview, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: in.Mode, ExplicitModel: in.ExplicitModel, ClaudeDefault: in.ClaudeDefault, GLM: options.GLM, GLMKey: options.GLMKey, Catalog: options.Catalog, Address: "127.0.0.1:1"})
		if err != nil {
			return gateway.LaunchPlan{}, nil, err
		}
		overlay := make(map[string]any, len(options.Overlay)+2)
		for key, value := range options.Overlay {
			overlay[key] = value
		}
		overlay["teammateMode"] = "in-process"
		overlay["model"] = preview.InitialModel
		pickerOptions := []any{}
		available := []string{}
		for _, model := range preview.Catalog.Entries() {
			label := model.RouteID
			switch model.RouteID {
			case "claude-opus-5":
				label = "Opus 5"
			case "claude-opus-5[1m]":
				label = "Opus 5 (1M context)"
			case "claude-sonnet-5":
				label = "Sonnet 5"
			case "claude-sonnet-5[1m]":
				label = "Sonnet 5 (1M context)"
			}
			pickerOptions = append(pickerOptions, map[string]any{"model": model.RouteID, "label": label})
			available = append(available, model.RouteID)
		}
		overlay["modelPicker"] = map[string]any{"replaceBuiltInOptions": true, "options": pickerOptions}
		overlay["availableModels"] = available

		// @MX:NOTE: [AUTO] REQ-MG-022 disables inherited client fallback for this session only.
		overlay["fallbackModel"] = []string{}
		data, inherited, _, err := prepareGatewayOverlay(in.Args, in.Inherited, overlay)
		if err != nil {
			return gateway.LaunchPlan{}, nil, err
		}
		opts := options.Child
		opts.Config.Overlay = data
		child, err := start(context.Background(), opts)
		if err != nil {
			return gateway.LaunchPlan{}, nil, err
		}
		plan, err := prepareGatewayLaunch(gatewayPrepareInput{Mode: in.Mode, ExplicitModel: in.ExplicitModel, ClaudeDefault: in.ClaudeDefault, GLM: options.GLM, GLMKey: options.GLMKey, Catalog: options.Catalog, Inherited: inherited, Address: child.Address, SessionEnv: options.SessionEnv})
		if err != nil {
			if child.Stop != nil {
				child.Stop()
			}
			return gateway.LaunchPlan{}, nil, err
		}
		if isNamedProfile(in.ProfileName) {
			filtered := plan.ChildEnv[:0]
			for _, item := range plan.ChildEnv {
				if !strings.HasPrefix(item, "CLAUDE_CONFIG_DIR=") {
					filtered = append(filtered, item)
				}
			}
			plan.ChildEnv = append(filtered, "CLAUDE_CONFIG_DIR="+profile.GetProfileDir(in.ProfileName))
		}
		if options.Family != nil {
			plan.Args = append(descriptor.Args, gatewayConversationPassthrough(in.Args)...)
			// The native transcript/config namespace is private to this family.
			// Secure storage remains the caller's explicitly selected namespace.
			filtered := plan.ChildEnv[:0]
			for _, item := range plan.ChildEnv {
				if strings.HasPrefix(item, "CLAUDE_CONFIG_DIR=") || strings.HasPrefix(item, "CLAUDE_SECURE_STORAGE_CONFIG_DIR=") {
					continue
				}
				filtered = append(filtered, item)
			}
			plan.ChildEnv = append(filtered, "CLAUDE_CONFIG_DIR="+descriptor.ConfigDir)
			secure, err := conversation.ResolveSecureNamespace(in.SecureStorage, in.SecureStorageSet, in.OriginalConfig, in.OriginalConfigSet)
			if err != nil {
				if child.Stop != nil {
					child.Stop()
				}
				return gateway.LaunchPlan{}, nil, err
			}
			if secure != "" {
				plan.ChildEnv = append(plan.ChildEnv, "CLAUDE_SECURE_STORAGE_CONFIG_DIR="+secure)
			}
		}
		plan.ChildSettings = child.OverlayPath
		if plan.ChildSettings == "" {
			if child.Stop != nil {
				child.Stop()
			}
			return gateway.LaunchPlan{}, nil, errors.New("gateway settings handoff missing")
		}
		return plan, child.Stop, nil
	}}
}

func prepareGatewayConversation(in gatewayLaunchRequest, families *conversation.Manager) (conversation.Descriptor, error) {
	if families == nil || in.CWD == "" || in.Project == "" {
		return conversation.Descriptor{}, errors.New("gateway conversation bootstrap unavailable")
	}
	if err := repairGatewayEnvelope(in, families); err != nil {
		return conversation.Descriptor{}, err
	}
	if in.Continue {
		d, err := families.Continue(context.Background(), in.Project)
		if err != nil {
			return conversation.Descriptor{}, fmt.Errorf("gateway continue: %w", err)
		}
		return d, nil
	}
	for i := 0; i < len(in.Args); i++ {
		if in.Args[i] != "--resume" {
			continue
		}
		if i+1 >= len(in.Args) || in.Args[i+1] == "" {
			return conversation.Descriptor{}, errors.New("gateway resume requires a UUID")
		}
		id := in.Args[i+1]
		for _, arg := range in.Args[i+2:] {
			if arg == "--fork-session" {
				d, err := families.Fork(context.Background(), id)
				if err != nil {
					return conversation.Descriptor{}, fmt.Errorf("gateway fork: %w", err)
				}
				return d, nil
			}
		}
		d, err := families.Resume(context.Background(), id)
		if err != nil {
			return conversation.Descriptor{}, fmt.Errorf("gateway resume: %w", err)
		}
		return d, nil
	}
	d, err := families.New(context.Background(), conversation.NewRequest{CWD: in.CWD, Project: in.Project, SecureStorage: in.SecureStorage, OriginalConfig: in.OriginalConfig, SecureStorageSet: in.SecureStorageSet, OriginalConfigSet: in.OriginalConfigSet})
	if err != nil {
		return conversation.Descriptor{}, fmt.Errorf("gateway new conversation: %w", err)
	}
	d.Args = []string{"--session-id", d.UUID}
	return d, nil
}

func gatewayConversationPassthrough(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--resume":
			i++
		case "--fork-session":
		case repairGatewayEnvelopeFlag:
		case "--session-id":
			i++
		default:
			out = append(out, args[i])
		}
	}
	return out
}

func replaceGatewaySettingsArgs(args []string, path string) ([]string, error) {
	result := make([]string, 0, len(args)+2)
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			result = append(result, "--settings", path)
			return append(result, args[i:]...), nil
		}
		if args[i] == "--settings" {
			if i+1 >= len(args) {
				return nil, errors.New("--settings requires a value")
			}
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--settings=") {
			continue
		}
		result = append(result, args[i])
	}
	return append(result, "--settings", path), nil
}
