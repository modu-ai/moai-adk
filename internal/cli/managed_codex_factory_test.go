package cli

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestManagedCodexOptionsPreserveModelAndConfig(t *testing.T) {
	app, model, err := managedCodexOptions([]string{"-c", `model_reasoning_effort="high"`, "--model", "gpt-6-sol", "--config=web_search=\"live\""})
	if err != nil {
		t.Fatal(err)
	}
	if model != "gpt-6-sol" || !slices.Equal(app, []string{"-c", `model_reasoning_effort="high"`, "-c", `web_search="live"`}) {
		t.Fatalf("app=%v model=%q", app, model)
	}
	if _, _, err := managedCodexOptions([]string{"--sandbox", "danger-full-access"}); err == nil {
		t.Fatal("unsupported options must fail before launching either host")
	}
}

func TestFactoryAppClientWaitTurnTracksSpecificCompletion(t *testing.T) {
	client := &factoryAppClient{events: make(chan factoryAppReply, 2), busy: true}
	client.events <- factoryAppReply{Method: "turn/completed", Params: json.RawMessage(`{"turn":{"id":"other","status":"completed"}}`)}
	client.events <- factoryAppReply{Method: "turn/completed", Params: json.RawMessage(`{"turn":{"id":"target","status":"completed"}}`)}
	if err := client.waitTurn(context.Background(), "target"); err != nil {
		t.Fatal(err)
	}
	if _, ok := client.completed["other"]; !ok {
		t.Fatal("unrelated completion was lost")
	}
}

func TestFactoryAppClientWaitTurnRejectsFailedCompletion(t *testing.T) {
	client := &factoryAppClient{events: make(chan factoryAppReply, 1), busy: true}
	client.events <- factoryAppReply{Method: "turn/completed", Params: json.RawMessage(`{"turn":{"id":"target","status":"failed"}}`)}
	if err := client.waitTurn(context.Background(), "target"); err == nil || !strings.Contains(err.Error(), "failed") {
		t.Fatalf("waitTurn error = %v", err)
	}
}
