package permission

import (
	"encoding/json"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestBubbleDispatcher_ShouldBubble(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	tests := []struct {
		name            string
		mode            PermissionMode
		isFork          bool
		decision        Decision
		parentAvailable bool
		shouldBubble    bool
	}{
		{
			name:            "bubble mode with fork and ask decision bubbles",
			mode:            ModeBubble,
			isFork:          true,
			decision:        DecisionAsk,
			parentAvailable: true,
			shouldBubble:    true,
		},
		{
			name:            "bubble mode with allow decision does not bubble",
			mode:            ModeBubble,
			isFork:          true,
			decision:        DecisionAllow,
			parentAvailable: true,
			shouldBubble:    false,
		},
		{
			name:            "bubble mode with deny decision does not bubble",
			mode:            ModeBubble,
			isFork:          true,
			decision:        DecisionDeny,
			parentAvailable: true,
			shouldBubble:    false,
		},
		{
			name:            "bubble mode without parent does not bubble",
			mode:            ModeBubble,
			isFork:          true,
			decision:        DecisionAsk,
			parentAvailable: false,
			shouldBubble:    false,
		},
		{
			name:            "bubble mode on non-fork does not bubble",
			mode:            ModeBubble,
			isFork:          false,
			decision:        DecisionAsk,
			parentAvailable: true,
			shouldBubble:    false,
		},
		{
			name:            "default mode does not bubble",
			mode:            ModeDefault,
			isFork:          true,
			decision:        DecisionAsk,
			parentAvailable: true,
			shouldBubble:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dispatcher.ShouldBubble(tt.mode, tt.isFork, tt.decision, tt.parentAvailable)
			if got != tt.shouldBubble {
				t.Errorf("ShouldBubble() = %v, want %v", got, tt.shouldBubble)
			}
		})
	}
}

func TestBubbleDispatcher_CreateBubbleRequest(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	result := &ResolveResult{
		Decision:   DecisionAsk,
		ResolvedBy: config.SrcProject,
		Origin:     ".claude/settings.json",
		Trace: ResolutionTrace{
			Tool:  "Write",
			Input: "/tmp/test.txt",
		},
	}

	req := dispatcher.CreateBubbleRequest("Write", json.RawMessage("/tmp/test.txt"), result, "parent-session-123")

	if req.Tool != "Write" {
		t.Errorf("BubbleRequest.Tool = %v, want 'Write'", req.Tool)
	}
	if req.Origin != ".claude/settings.json" {
		t.Errorf("BubbleRequest.Origin = %v, want '.claude/settings.json'", req.Origin)
	}
	if req.ParentSessionID != "parent-session-123" {
		t.Errorf("BubbleRequest.ParentSessionID = %v, want 'parent-session-123'", req.ParentSessionID)
	}
}

func TestBubbleDispatcher_FormatBubblePrompt(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	req := &BubbleRequest{
		Tool:            "Write",
		Input:           json.RawMessage(`{"path": "/tmp/test.txt"}`),
		ForkDepth:       2,
		Origin:          ".claude/settings.json",
		ParentSessionID: "parent-session-123",
	}

	prompt := dispatcher.FormatBubblePrompt(req)

	// Check that prompt contains key information
	requiredStrings := []string{
		"Write",
		"/tmp/test.txt",
		".claude/settings.json",
		"Fork depth: 2",
		"Allow this operation?",
	}

	for _, s := range requiredStrings {
		if !contains(prompt, s) {
			t.Errorf("FormatBubblePrompt() output missing required string: %s\nGot: %s", s, prompt)
		}
	}
}

func TestBubbleDispatcher_HandleBubbleResponse(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	originalTrace := ResolutionTrace{
		Tool:  "Write",
		Input: "/tmp/test.txt",
		Tries: []TierTry{
			{
				Tier:    config.SrcProject,
				Matched: true,
				Reason:  "rule matched",
			},
		},
	}

	req := &BubbleRequest{
		Tool:            "Write",
		Input:           json.RawMessage("/tmp/test.txt"),
		ForkDepth:       1,
		Origin:          ".claude/settings.json",
		ParentSessionID: "parent-session-123",
	}

	resp := &BubbleResponse{
		Decision: DecisionAllow,
		Reason:   "User approved the operation",
	}

	result := dispatcher.HandleBubbleResponse(req, resp, originalTrace)

	if result.Decision != DecisionAllow {
		t.Errorf("HandleBubbleResponse() Decision = %v, want %v", result.Decision, DecisionAllow)
	}
	if result.Origin != ".claude/settings.json" {
		t.Errorf("HandleBubbleResponse() Origin = %v, want '.claude/settings.json'", result.Origin)
	}
	if result.SystemMessage == "" {
		t.Error("HandleBubbleResponse() SystemMessage should be set")
	}
	if !contains(result.SystemMessage, "user decided: allow") {
		t.Errorf("HandleBubbleResponse() SystemMessage = %v, should contain 'user decided: allow'", result.SystemMessage)
	}
}

func TestBubbleDispatcher_ValidateForkDepth(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	tests := []struct {
		name         string
		depth        int
		mode         PermissionMode
		wantWarn     bool
		warnContains string
	}{
		{
			name:     "depth within limit - no warning",
			depth:    2,
			mode:     ModeAcceptEdits,
			wantWarn: false,
		},
		{
			name:         "depth exceeds limit - warning",
			depth:        4,
			mode:         ModeAcceptEdits,
			wantWarn:     true,
			warnContains: "exceeds limit",
		},
		{
			name:     "depth exceeds limit but plan mode - no warning",
			depth:    4,
			mode:     ModePlan,
			wantWarn: false,
		},
		{
			name:     "depth exceeds limit but bubble mode - no warning",
			depth:    4,
			mode:     ModeBubble,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning, err := dispatcher.ValidateForkDepth(tt.depth, tt.mode)
			if err != nil {
				t.Errorf("ValidateForkDepth() unexpected error = %v", err)
			}
			if tt.wantWarn && warning == "" {
				t.Error("ValidateForkDepth() should return warning")
			}
			if !tt.wantWarn && warning != "" {
				t.Errorf("ValidateForkDepth() should not return warning, got %v", warning)
			}
			if tt.wantWarn && tt.warnContains != "" && !contains(warning, tt.warnContains) {
				t.Errorf("ValidateForkDepth() warning = %v, should contain %q", warning, tt.warnContains)
			}
		})
	}
}

func TestBubbleDispatcher_IsParentAvailable(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	tests := []struct {
		name              string
		parentSessionID   string
		parentAvailable   bool
	}{
		{
			name:            "valid session ID - available",
			parentSessionID: "parent-session-123",
			parentAvailable: true,
		},
		{
			name:            "empty session ID - not available",
			parentSessionID: "",
			parentAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dispatcher.IsParentAvailable(tt.parentSessionID)
			if got != tt.parentAvailable {
				t.Errorf("IsParentAvailable() = %v, want %v", got, tt.parentAvailable)
			}
		})
	}
}

func TestBubbleDispatcher_DispatchToParent(t *testing.T) {
	dispatcher := NewBubbleDispatcher()

	req := &BubbleRequest{
		Tool:            "Write",
		Input:           json.RawMessage("/tmp/test.txt"),
		ForkDepth:       1,
		Origin:          ".claude/settings.json",
		ParentSessionID: "parent-session-123",
	}

	_, err := dispatcher.DispatchToParent(req)
	if err == nil {
		t.Error("DispatchToParent() should return error (not implemented)")
	}
	if !contains(err.Error(), "not implemented") {
		t.Errorf("DispatchToParent() error = %v, should contain 'not implemented'", err)
	}
}
