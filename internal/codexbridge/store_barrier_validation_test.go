package codexbridge

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBarrierRejectsCorruptAndForeignDiskRecord(t *testing.T) {
	for _, kind := range []string{"malformed", "foreign-owner", "unsafe-mode", "directory"} {
		t.Run(kind, func(t *testing.T) {
			if kind == "unsafe-mode" && runtime.GOOS == "windows" {
				t.Skip("POSIX mode contract; Windows ACL validation requires platform CI")
			}
			_, _, store := fixture(t)
			owner := request("barrier-validation").Owner
			rec := record{Schema: storeSchema, Owner: owner, Phase: "idle"}
			if kind == "foreign-owner" {
				rec.Owner.AccountScope = "foreign"
			}
			raw, err := json.Marshal(rec)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "malformed" {
				raw = []byte(`{"Schema":`)
			}
			path := filepath.Join(store.dir, storeKey(owner)+".json")
			if kind == "directory" {
				err = os.Mkdir(path, 0700)
			} else {
				mode := os.FileMode(0600)
				if kind == "unsafe-mode" {
					mode = 0644
				}
				err = os.WriteFile(path, raw, mode)
			}
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = store.Barrier(owner)
			if err == nil {
				t.Fatal("unsafe barrier accepted")
			}
			if kind == "foreign-owner" && !errors.Is(err, ErrScope) {
				t.Fatalf("foreign ownership error=%v", err)
			}
		})
	}
}
