package graph

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type replacingMetaPublisher struct {
	osPrivateGTDPublisher
	winnerData, winnerMeta []byte
}

func (p replacingMetaPublisher) Install(source privateGTDPublication, target string) error {
	if filepath.Base(target) != privateGTDMetaName {
		return p.osPrivateGTDPublisher.Install(source, target)
	}
	if _, _, err := publishPrivateGTDProjection(filepath.Dir(target), p.winnerData, p.winnerMeta); err != nil {
		return err
	}
	return errors.New("injected metadata publication refusal after competing writer")
}

func TestPrivateGTDFailedPublicationPreservesConcurrentWinner(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "todo")
	winnerData, winnerMeta := []byte("winner data\n"), []byte("winner metadata\n")
	publisher := replacingMetaPublisher{winnerData: winnerData, winnerMeta: winnerMeta}
	if _, _, err := publishPrivateGTDProjectionWith(publisher, dir, []byte("loser secret\n"), []byte("loser metadata\n")); err == nil {
		t.Fatal("metadata refusal was accepted")
	}
	for name, want := range map[string][]byte{privateGTDDataName: winnerData, privateGTDMetaName: winnerMeta} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != string(want) {
			t.Errorf("competing writer's %s changed: got %q, err %v; want %q", name, got, err, want)
		}
	}
}
