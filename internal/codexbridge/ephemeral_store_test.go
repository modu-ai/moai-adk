package codexbridge

import (
	"github.com/modu-ai/moai-adk/internal/codextools"
	"testing"
)

func TestEphemeralRemovalRequiresExactCompletedOwner(t *testing.T) {
	_, _, store := fixture(t)
	owner := codextools.Binding{ConversationID: "fixture", AccountScope: "account", ThreadID: "thread"}
	r := record{Owner: owner, Phase: "waiting", Model: "gpt-5.6-luna", CWD: "/synthetic"}
	if err := store.save(r); err != nil {
		t.Fatal(err)
	}
	if err := store.removeCompleted(owner); err == nil {
		t.Fatal("pending barrier deleted")
	}
	r.Phase = "idle"
	if err := store.save(r); err != nil {
		t.Fatal(err)
	}
	foreign := owner
	foreign.ThreadID = "foreign"
	if err := store.removeCompleted(foreign); err == nil {
		t.Fatal("foreign thread removed barrier")
	}
	if err := store.removeCompleted(owner); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.load(owner); err != nil || found {
		t.Fatalf("completed barrier remains found=%v err=%v", found, err)
	}
}
