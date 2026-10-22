package factory

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
)

type vacuumExecutorFunc func(context.Context, string, ...any) (sql.Result, error)

func (f vacuumExecutorFunc) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return f(ctx, q, args...)
}

func TestVacuumGTDOutputIsPrivateBeforeFirstSQLByte(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.db")
	entered := false
	exec := vacuumExecutorFunc(func(_ context.Context, _ string, _ ...any) (sql.Result, error) {
		entered = true
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() != 0 {
			return nil, errors.New("output is not an empty regular file before SQL")
		}
		if err := privatefile.Check(path); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return nil, err
		}
		_, writeErr := f.Write([]byte("synthetic full-snapshot payload"))
		return nil, errors.Join(writeErr, f.Close())
	})
	if err := vacuumGTDInto(t.Context(), exec, path); err != nil {
		t.Fatalf("SQL entered without a private output: %v", err)
	}
	if !entered {
		t.Fatal("VACUUM writer did not run")
	}
	if err := privatefile.Check(path); err != nil {
		t.Fatal(err)
	}
	// Function-owned guards must be closed, so normal cleanup remains possible.
	if err := os.Remove(path); err != nil {
		t.Fatalf("completed output remained pinned: %v", err)
	}
}

type gatedSnapshotReader struct {
	io.Reader
	entered chan struct{}
	release chan struct{}
	first   bool
}

func (r *gatedSnapshotReader) Read(p []byte) (int, error) {
	if !r.first {
		r.first = true
		close(r.entered)
		<-r.release
	}
	return r.Reader.Read(p)
}

func TestRestoreGTDRejectsDestinationCreatedDuringCopy(t *testing.T) {
	for _, occupied := range []string{"", "another writer's original data"} {
		name := "nonempty"
		if occupied == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			dir := privateBacklogFixtureDir(t)
			destination := filepath.Join(dir, "backlog.db")
			original := []byte("synthetic full database snapshot")
			r := &gatedSnapshotReader{Reader: bytes.NewReader(original), entered: make(chan struct{}), release: make(chan struct{})}
			done := make(chan error, 1)
			go func() { done <- restoreGTDSnapshot(t.Context(), r, destination) }()
			select {
			case <-r.entered:
			case err := <-done:
				t.Fatalf("restore failed before copy barrier: %v", err)
			}
			if err := os.WriteFile(destination, []byte(occupied), 0o600); err != nil {
				close(r.release)
				<-done
				t.Fatal(err)
			}
			close(r.release)
			if err := <-done; !errors.Is(err, os.ErrExist) {
				t.Fatalf("concurrent destination error=%v, want existing-file refusal", err)
			}
			got, err := os.ReadFile(destination)
			if err != nil || string(got) != occupied {
				t.Fatalf("concurrent destination was changed: %v", err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0].Name() != "backlog.db" {
				t.Fatal("failed restore left staging artifacts")
			}
		})
	}
}

func TestVacuumGTDRefusesExistingOutputBeforeSQL(t *testing.T) {
	for _, content := range []string{"", "existing backup"} {
		t.Run(fmt.Sprintf("size-%d", len(content)), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "backup.db")
			if err := privatefile.WriteFile(path, []byte(content)); err != nil {
				t.Fatal(err)
			}
			entered := false
			exec := vacuumExecutorFunc(func(context.Context, string, ...any) (sql.Result, error) { entered = true; return nil, nil })
			if err := vacuumGTDInto(t.Context(), exec, path); !errors.Is(err, os.ErrExist) {
				t.Fatalf("occupied error=%v", err)
			}
			if entered {
				t.Fatal("SQL entered for an occupied destination")
			}
			if got, err := os.ReadFile(path); err != nil || string(got) != content {
				t.Fatalf("existing backup changed: %v", err)
			}
			if err := privatefile.Check(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestVacuumGTDFailureKeepsPrivatePartialOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.db")
	sentinel := errors.New("synthetic SQL failure")
	exec := vacuumExecutorFunc(func(context.Context, string, ...any) (sql.Result, error) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return nil, err
		}
		_, writeErr := f.WriteString("partial synthetic snapshot")
		return nil, errors.Join(sentinel, writeErr, f.Close())
	})
	if err := vacuumGTDInto(t.Context(), exec, path); !errors.Is(err, sentinel) {
		t.Fatalf("SQL error identity lost: %v", err)
	}
	if err := privatefile.Check(path); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "partial synthetic snapshot" {
		t.Fatalf("partial output behavior changed: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("failed output stayed pinned: %v", err)
	}
}

type cancelSnapshotReader struct{ cancel context.CancelFunc }

func (r cancelSnapshotReader) Read(p []byte) (int, error) {
	r.cancel()
	return copy(p, "partial snapshot"), nil
}

func TestGTDSnapshotCancellationDoesNotPublish(t *testing.T) {
	t.Run("before-backup", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "absent.db")
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		exec := vacuumExecutorFunc(func(context.Context, string, ...any) (sql.Result, error) {
			t.Fatal("canceled SQL entered")
			return nil, nil
		})
		if err := vacuumGTDInto(ctx, exec, path); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel error=%v", err)
		}
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("canceled backup created output: %v", err)
		}
	})
	t.Run("during-restore", func(t *testing.T) {
		dir := privateBacklogFixtureDir(t)
		path := filepath.Join(dir, "absent.db")
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		if err := restoreGTDSnapshot(ctx, cancelSnapshotReader{cancel}, path); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel error=%v", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 0 {
			t.Fatalf("canceled restore published or leaked staging: %v", err)
		}
	})
}

func TestGTDWholeDatabaseBackupRestorePreservesAdditionalSchema(t *testing.T) {
	source := newGTDTestStore(t)
	ctx := t.Context()
	item, err := CaptureGTDItem(ctx, source, CaptureInput{Content: "snapshot fixture", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "whole-snapshot"})
	if err != nil {
		t.Fatal(err)
	}
	db, err := openGTDDB(source)
	if err != nil {
		t.Fatal(err)
	}
	// A logical export of the known GTD tables cannot satisfy this whole-DB contract.
	_, err = db.ExecContext(ctx, `CREATE TABLE snapshot_extra(k TEXT PRIMARY KEY, payload BLOB NOT NULL);
 CREATE INDEX snapshot_extra_payload ON snapshot_extra(payload);
 CREATE VIEW snapshot_extra_view AS SELECT k,payload FROM snapshot_extra;
 INSERT INTO snapshot_extra VALUES('outside-gtd',x'000102ff');`)
	closeErr := db.Close()
	if err := errors.Join(err, closeErr); err != nil {
		t.Fatal(err)
	}
	before, err := ExportGTD(ctx, source, true)
	if err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "snapshot'quoted.db")
	if err := BackupGTDStore(ctx, source, backup, true); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(backup); err != nil {
		t.Fatal(err)
	}
	digestBefore, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	restored := newGTDTestStore(t)
	if err := RestoreGTDStore(ctx, backup, restored, true); err != nil {
		t.Fatal(err)
	}
	if got, err := LoadGTDItem(ctx, restored, item.ItemID); err != nil || got.Content != item.Content {
		t.Fatalf("GTD item not preserved: %v", err)
	}
	restoredDB, err := openGTDDB(restored)
	if err != nil {
		t.Fatal(err)
	}
	var payload []byte
	queryErr := restoredDB.QueryRowContext(ctx, `SELECT payload FROM snapshot_extra_view WHERE k='outside-gtd'`).Scan(&payload)
	var indexCount int
	indexErr := restoredDB.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_schema WHERE type='index' AND name='snapshot_extra_payload'`).Scan(&indexCount)
	if err := errors.Join(queryErr, indexErr, restoredDB.Close()); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(payload, []byte{0, 1, 2, 255}) || indexCount != 1 {
		t.Fatal("whole DB snapshot omitted unrelated schema/data")
	}
	after, err := ExportGTD(ctx, source, true)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("source logical state changed: %v", err)
	}
	digestAfter, err := os.ReadFile(backup)
	if err != nil || !bytes.Equal(digestBefore, digestAfter) {
		t.Fatalf("restore changed rollback snapshot: %v", err)
	}
	// A second restore must preserve the already published destination and backup.
	if err := RestoreGTDStore(ctx, backup, restored, true); err == nil {
		t.Fatal("second restore overwrote existing database")
	}
}
