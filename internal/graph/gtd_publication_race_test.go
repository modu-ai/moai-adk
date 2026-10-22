package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type replacingMetaPublisher struct {
	osPrivateGTDPublisher
	winnerData, winnerMeta []byte
	displacedData          string
}

var errCompetingGTDPublicationRefusal = errors.New("injected metadata publication refusal after competing writer")

func (p replacingMetaPublisher) Install(source privateGTDPublication, target string) error {
	if filepath.Base(target) != privateGTDMetaName {
		return p.osPrivateGTDPublisher.Install(source, target)
	}
	// Ordinary Windows replacement can refuse an open destination.
	// Move that object aside first, as a competing writer can, then publish
	// the successor pair while the losing writer still owns the old inode.
	if err := os.Rename(filepath.Join(filepath.Dir(target), privateGTDDataName), p.displacedData); err != nil {
		return fmt.Errorf("competing writer could not move losing data: %w", err)
	}
	if _, _, err := publishPrivateGTDProjection(filepath.Dir(target), p.winnerData, p.winnerMeta); err != nil {
		return fmt.Errorf("competing writer could not publish: %w", err)
	}
	return errCompetingGTDPublicationRefusal
}

func TestPrivateGTDFailedPublicationPreservesConcurrentWinner(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "todo")
	winnerData, winnerMeta := []byte("winner data\n"), []byte("winner metadata\n")
	displaced := filepath.Join(dir, "displaced-losing-data")
	publisher := replacingMetaPublisher{winnerData: winnerData, winnerMeta: winnerMeta, displacedData: displaced}
	if _, _, err := publishPrivateGTDProjectionWith(publisher, dir, []byte("loser secret\n"), []byte("loser metadata\n")); !errors.Is(err, errCompetingGTDPublicationRefusal) {
		t.Fatalf("did not reach metadata refusal after successful competing publication: %v", err)
	}
	for name, want := range map[string][]byte{privateGTDDataName: winnerData, privateGTDMetaName: winnerMeta} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != string(want) {
			t.Errorf("competing writer's %s changed: got %q, err %v; want %q", name, got, err, want)
		}
	}
	if got, err := os.ReadFile(displaced); err == nil {
		if len(got) != 0 {
			t.Errorf("losing object retained secret bytes after refusal: %q", got)
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("read displaced losing object: %v", err)
	}
}
