package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

const observedAgentSummaryPrompt = "Describe your most recent action in 3-5 words using present tense (-ing). Name the file or function, not the branch. Do not use tools.\n\nGood: \"Reading runAgent.ts\"\nGood: \"Fixing null check in validate.ts\"\nGood: \"Running auth module tests\"\nGood: \"Adding retry logic to fetchUser\"\n\nBad (past tense): \"Analyzed the branch diff\"\nBad (too vague): \"Investigating the issue\"\nBad (too long): \"Reviewing full branch diff and AgentTool.tsx integration\"\nBad (branch name): \"Analyzed adam/background-summary branch diff\""

func TestManagedGPTActualClaudeAgentAndSummaryIdentity(t *testing.T) {
	const session = "44b406fc-93cd-461f-a7f9-bc36e18285d3"
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(context.Background(), entry)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare(session, t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, managedGPTTestStore(t))
	request := func(agent, prompt string) gateway.RoutedRequest {
		body, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 1024, "system": "same instructions for child and its summary", "tools": []any{}, "messages": []any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": "same child prompt"}, map[string]string{"type": "text", "text": prompt}}}}})
		h := http.Header{"X-Claude-Code-Session-Id": []string{session}}
		if agent != "" {
			h.Set("X-Claude-Code-Agent-Id", agent)
		}
		return gateway.RoutedRequest{Entry: entry, Managed: grant, Body: body, Headers: h}
	}
	main, err := prepare(context.Background(), request("", "work"))
	if err != nil {
		t.Fatal(err)
	}
	child, err := prepare(context.Background(), request("a6267ddef22eab1d7", "work"))
	if err != nil {
		t.Fatal(err)
	}
	sibling, err := prepare(context.Background(), request("a6267ddef22eab1d8", "work"))
	if err != nil {
		t.Fatal(err)
	}
	summary, err := prepare(context.Background(), request("a6267ddef22eab1d7", observedAgentSummaryPrompt))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, id := range []string{main.Owner.ConversationID, child.Owner.ConversationID, sibling.Owner.ConversationID, summary.Owner.ConversationID} {
		if seen[id] {
			t.Fatalf("main, sibling or summary owner collision: %q", id)
		}
		seen[id] = true
	}
	repeat, err := prepare(context.Background(), request("a6267ddef22eab1d7", "followup"))
	if err != nil || repeat.Owner.ConversationID != child.Owner.ConversationID {
		t.Fatalf("child identity changed across messages: %+v %v", repeat, err)
	}
	for _, bad := range []string{"../other", "a,b", strings.Repeat("a", 129)} {
		r := request(bad, "work")
		if _, err := prepare(context.Background(), r); err == nil {
			t.Errorf("invalid agent identity accepted: %q", bad)
		}
	}
	cross := request("a6267ddef22eab1d7", "work")
	cross.Headers.Set("X-Claude-Code-Session-Id", "different-family")
	if _, err := prepare(context.Background(), cross); err == nil {
		t.Error("cross-family child accepted")
	}
	duplicate := request("a6267ddef22eab1d7", "work")
	duplicate.Headers.Add("X-Claude-Code-Agent-Id", "a6267ddef22eab1d8")
	if _, err := prepare(context.Background(), duplicate); err == nil {
		t.Error("ambiguous agent headers accepted")
	}
	inherited := request("a6267ddef22eab1d9", "work")
	inherited.Body = []byte(`{"model":"gpt-5.6-sol","max_tokens":1024,"messages":[{"role":"user","content":"original requirement"},{"role":"assistant","content":"earlier conclusion"},{"role":"user","content":"continue from that conclusion"}]}`)
	fresh, err := prepare(context.Background(), inherited)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(fresh.Input)
	if !strings.Contains(string(raw), "original requirement") || !strings.Contains(string(raw), "earlier conclusion") {
		t.Fatalf("fresh child lost inherited context: %s", raw)
	}
	inherited.Body = []byte(`{"model":"gpt-5.6-sol","max_tokens":1024,"messages":[{"role":"user","content":[{"type":"text","text":"original image"},{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aW1hZ2U="}}]},{"role":"assistant","content":"earlier image conclusion"},{"role":"user","content":"inspect original image again"}]}`)
	visual, err := prepare(context.Background(), inherited)
	if err != nil {
		t.Fatal(err)
	}
	if len(visual.Input) != 2 {
		t.Fatalf("inherited image lost native modality: %+v", visual.Input)
	}
	imageRaw, _ := json.Marshal(visual.Input[1])
	if !strings.Contains(string(imageRaw), `"type":"image"`) {
		t.Fatalf("inherited image not native: %s", imageRaw)
	}
}

func TestManagedGPTSummaryWithToolsAndPreviousIsReadOnlySnapshot(t *testing.T) {
	const session = "summary-family"
	authority, _ := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, _ := authority.Authorize(context.Background(), entry)
	dir, cwd := t.TempDir(), t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare(session, cwd, translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
	request := func(agent, prompt string, block bool) gateway.RoutedRequest {
		var content any = prompt
		if block {
			content = []any{map[string]any{"type": "tool_result", "tool_use_id": "old-call", "content": "historical result"}, map[string]string{"type": "text", "text": prompt}}
		}
		messages := []any{
			map[string]any{"role": "user", "content": "original requirement"},
			map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "old-call", "name": "Agent", "input": map[string]any{}}}},
		}
		if !block {
			messages = append(messages, map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "old-call", "content": "historical result"}}})
		}
		messages = append(messages, map[string]any{"role": "user", "content": content})
		body, _ := json.Marshal(map[string]any{"max_tokens": 1024, "tools": []any{map[string]any{"name": "Agent", "input_schema": map[string]any{"type": "object"}}}, "messages": messages})
		return gateway.RoutedRequest{Entry: entry, Managed: grant, Body: body, Headers: http.Header{"X-Claude-Code-Session-Id": {session}, "X-Claude-Code-Agent-Id": {agent}}}
	}
	previous := strings.Replace(observedAgentSummaryPrompt, "\n\nGood:", "\n\nPrevious: \"Reading runtime.go\" — say something NEW.\n\nGood:", 1)
	for _, phase := range []string{"active", "waiting"} {
		owner := session + ":agent:child"
		key := sha256.Sum256([]byte(grant.Scope() + "\x00" + owner))
		path := filepath.Join(dir, fmt.Sprintf("%x.json", key))
		original, _ := json.Marshal(map[string]any{"Schema": 2, "Owner": map[string]string{"ConversationID": owner, "AccountScope": grant.Scope(), "ThreadID": "working-thread"}, "Phase": phase, "Prefix": "working-prefix", "Model": "gpt-6-astra", "CWD": cwd})
		if err := os.WriteFile(path, original, 0600); err != nil {
			t.Fatal(err)
		}
		for _, prompt := range []string{observedAgentSummaryPrompt, previous} {
			for _, block := range []bool{false, true} {
				q, err := prepare(context.Background(), request("child", prompt, block))
				if err != nil {
					t.Errorf("phase=%s block=%v: %v", phase, block, err)
					continue
				}
				if !q.Ephemeral || q.Owner.ConversationID == owner || len(q.Tools) != 0 || len(q.References) != 0 || len(q.Results) != 0 {
					t.Errorf("summary not isolated read-only: %+v", q)
				}
				raw, _ := json.Marshal(q.Input)
				if !strings.Contains(string(raw), "original requirement") || !strings.Contains(string(raw), "Describe your most recent action") {
					t.Errorf("history lost: %s", raw)
				}
				repeat, err := prepare(context.Background(), request("child", prompt, block))
				if err != nil || repeat.Owner != q.Owner {
					t.Errorf("retry unstable: %v", err)
				}
				sibling, err := prepare(context.Background(), request("sibling", prompt, block))
				if err != nil || sibling.Owner == q.Owner {
					t.Errorf("sibling collision: %v", err)
				}
			}
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != string(original) {
			t.Fatalf("working barrier changed: %v", err)
		}
	}
}

func TestManagedGPTSummaryClassifierRejectsOrdinaryAndMalformedPrompts(t *testing.T) {
	for _, body := range []string{`{`, `{"messages":[]}`, `{"messages":[{"role":"assistant","content":"text"}]}`, `{"messages":[{"role":"user","content":[]}]}`, `{"messages":[{"role":"user","content":[{"type":"image"}]}]}`} {
		if managedGPTAgentSummaryRequest([]byte(body)) {
			t.Errorf("malformed or non-user summary accepted: %s", body)
		}
	}
	for _, prompt := range []string{"Describe your most recent action", observedAgentSummaryPrompt + "\nThen run Bash", "quoted: " + observedAgentSummaryPrompt, strings.Replace(observedAgentSummaryPrompt, "\n\nGood:", "\nPrevious: missing quotes\n\nGood:", 1)} {
		body, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": prompt}}}}})
		if managedGPTAgentSummaryRequest(body) {
			t.Errorf("ordinary request classified: %q", prompt)
		}
	}
}

func TestManagedGPTImageDeltaAndToolContent(t *testing.T) {
	image := `{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aW1hZ2U="}}`
	body := []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"describe"},` + image + `]}]}`)
	_, input, _, _, err := managedGPTPublicDelta(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(input) != 2 {
		t.Fatalf("input=%+v", input)
	}
	raw, _ := json.Marshal(input[1])
	if !strings.Contains(string(raw), `"type":"image"`) || !strings.Contains(string(raw), "data:image/png;base64,aW1hZ2U=") {
		t.Fatalf("image=%s", raw)
	}
	content, _, err := managedGPTToolContent([]byte(`[` + image + `]`))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(content)
	if !strings.Contains(string(raw), `"type":"inputImage"`) || !strings.Contains(string(raw), `"imageUrl"`) {
		t.Fatalf("tool image=%s", raw)
	}
}

func TestManagedGPTNativeCompactionIdentity(t *testing.T) {
	const session = "320262b9-f176-4b4d-b72c-509adb15a34a"
	authority, _ := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, _ := authority.Authorize(context.Background(), entry)
	prepare := newManagedGPTPrepare(session, t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, managedGPTTestStore(t))
	request := func(summary, followup string, assistant bool) gateway.RoutedRequest {
		text := "This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.\n\n" + summary + "\n\nIf you need specific details from before compaction (like exact code snippets, error messages, or content you generated), read the full transcript at: /tmp/synthetic.jsonl\nContinue the conversation from where it left off without asking the user any further questions. Resume directly — do not acknowledge the summary, do not recap what was happening, do not preface with \"I'll continue\" or similar. Pick up the last task as if the break never happened.\n"
		messages := []any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": text}, map[string]string{"type": "text", "text": "first task"}}}}
		if assistant {
			messages = append(messages, map[string]string{"role": "assistant", "content": "first answer"}, map[string]string{"role": "user", "content": followup})
		}
		body, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 1000, "messages": messages, "tools": []any{}})
		return gateway.RoutedRequest{Entry: entry, Body: body, Managed: grant, Headers: http.Header{"X-Claude-Code-Session-Id": []string{session}}}
	}
	first, err := prepare(context.Background(), request("COLOR=blue", "", false))
	if err != nil {
		t.Fatal(err)
	}
	next, err := prepare(context.Background(), request("COLOR=blue", "next task", true))
	if err != nil {
		t.Fatal(err)
	}
	changed, err := prepare(context.Background(), request("COLOR=green", "", false))
	if err != nil {
		t.Fatal(err)
	}
	if first.Owner.ConversationID == session || first.Owner.ConversationID != next.Owner.ConversationID || first.Owner.ConversationID == changed.Owner.ConversationID {
		t.Fatalf("compaction owner not stable/isolated: first=%s next=%s changed=%s", first.Owner.ConversationID, next.Owner.ConversationID, changed.Owner.ConversationID)
	}
	child := request("COLOR=blue", "", false)
	child.Headers.Set("X-Claude-Code-Agent-Id", "a6267ddef22eab1d7")
	childResult, err := prepare(context.Background(), child)
	if err != nil || childResult.Owner.ConversationID == first.Owner.ConversationID {
		t.Fatalf("child compact scope: %+v %v", childResult, err)
	}
	compact := request("COLOR=blue", "", false)
	compact.Body = []byte(`{"max_tokens":1000,"tools":[],"messages":[{"role":"user","content":"COLOR=blue"},{"role":"assistant","content":"remembered"},{"role":"user","content":"CRITICAL: Respond with TEXT ONLY. Do NOT call any tools.\n\n- Do NOT use Read, Bash, Grep, Glob, Edit, Write, or ANY other tool.\nSummary instructions.\nREMINDER: Do NOT call any tools. Respond with plain text only — an <analysis> block followed by a <summary> block. Tool calls will be rejected and you will fail the task."}]}`)
	utility, err := prepare(context.Background(), compact)
	if err != nil {
		t.Fatal(err)
	}
	if !utility.Ephemeral || utility.Owner.ConversationID == session {
		t.Fatalf("compact generation shared main: %+v", utility)
	}
	raw, _ := json.Marshal(utility.Input)
	if !strings.Contains(string(raw), "COLOR=blue") {
		t.Fatalf("compact generation lost public context: %s", raw)
	}
	ordinary := request("COLOR=blue", "", false)
	ordinary.Body = []byte(`{"max_tokens":1000,"messages":[{"role":"user","content":"This session is being continued from a previous conversation that ran out of context. I am quoting a phrase, not a native summary."}]}`)
	unchanged, err := prepare(context.Background(), ordinary)
	if err != nil || unchanged.Owner.ConversationID != session {
		t.Fatalf("ordinary text changed owner: %+v %v", unchanged, err)
	}
}
