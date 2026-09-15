package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagedGPTRejectionLogIsPrivateBoundedAndRedacted(t *testing.T) {
	dir, resolveErr := filepath.EvalSymlinks(t.TempDir())
	if resolveErr != nil {
		t.Fatal(resolveErr)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	logger, err := newManagedGPTRejectionLogger(dir, "fixture-owner")
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{"cause": "appserver_scope_mismatch", "route": "gpt-5.6-sol", "digest": strings.Repeat("a", 64), "prompt": "CANARY-PRIVATE"}
	logger.RecordGatewayRejection(fields)
	raw, err := os.ReadFile(filepath.Join(dir, logger.name))
	if err != nil {
		t.Fatal(err)
	}
	var row map[string]string
	if err = json.Unmarshal(raw, &row); err != nil {
		t.Fatal(err)
	}
	if len(row) != 4 || row["cause"] != fields["cause"] || row["at"] == "" || strings.Contains(string(raw), "CANARY") {
		t.Fatal(string(raw))
	}
	info, err := os.Stat(filepath.Join(dir, logger.name))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal(info, err)
	}
	fields["cause"] = "CANARY-PRIVATE"
	logger.RecordGatewayRejection(fields)
	after, _ := os.ReadFile(filepath.Join(dir, logger.name))
	if string(after) != string(raw) {
		t.Fatal("unknown cause accepted")
	}
	if err = os.WriteFile(filepath.Join(dir, logger.name), make([]byte, 1<<20), 0600); err != nil {
		t.Fatal(err)
	}
	fields["cause"] = "appserver_scope_mismatch"
	logger.RecordGatewayRejection(fields)
	info, err = os.Stat(filepath.Join(dir, logger.name))
	if err != nil || info.Size() != 1<<20 {
		t.Fatal(info, err)
	}
}

func TestManagedGPTRejectionLoggerBoundsRPCCode(t *testing.T) {
	for _, code := range []string{"-32603", "2147483647", "-2147483648", "2147483648", "-2147483649", "01", "+1", "1.2", "CANARY"} {
		t.Run(code, func(t *testing.T) {
			dir, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			logger, err := newManagedGPTRejectionLogger(dir, "rpc-owner")
			if err != nil {
				t.Fatal(err)
			}
			logger.RecordGatewayRejection(map[string]string{"cause": "appserver_transport_error", "route": "gpt-5.6-sol", "digest": strings.Repeat("d", 64), "rpc_code": code, "message": "CANARY"})
			raw, err := os.ReadFile(filepath.Join(dir, logger.name))
			if err != nil {
				t.Fatal(err)
			}
			var row map[string]string
			if err = json.Unmarshal(raw, &row); err != nil {
				t.Fatal(err)
			}
			valid := code == "-32603" || code == "2147483647" || code == "-2147483648"
			if valid && row["rpc_code"] != code || !valid && row["rpc_code"] != "" {
				t.Fatalf("rpc code filter mismatch: %q -> %q", code, row["rpc_code"])
			}
			if strings.Contains(string(raw), "CANARY") {
				t.Fatal("untrusted RPC metadata leaked")
			}
		})
	}
}

func TestManagedGPTRejectionLoggerRejectsSymlink(t *testing.T) {
	dir, resolveErr := filepath.EvalSymlinks(t.TempDir())
	if resolveErr != nil {
		t.Fatal(resolveErr)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	logger, err := newManagedGPTRejectionLogger(dir, "owner")
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "target")
	if err = os.WriteFile(target, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(target, filepath.Join(dir, logger.name)); err != nil {
		t.Fatal(err)
	}
	logger.RecordGatewayRejection(map[string]string{"cause": "appserver_scope_mismatch", "route": "gpt-5.6-sol", "digest": strings.Repeat("b", 64)})
	raw, err := os.ReadFile(target)
	if err != nil || string(raw) != "preserve" {
		t.Fatal(string(raw), err)
	}
}

func TestManagedGPTRejectionLoggerPreservesQueueCauseOnly(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	logger, err := newManagedGPTRejectionLogger(dir, "queue-owner")
	if err != nil {
		t.Fatal(err)
	}
	logger.RecordGatewayRejection(map[string]string{"cause": "appserver_event_queue_bytes_exceeded", "route": "gpt-5.6-sol", "digest": strings.Repeat("c", 64), "detail": "CANARY PRIVATE"})
	raw, err := os.ReadFile(filepath.Join(dir, logger.name))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "appserver_event_queue_bytes_exceeded") || strings.Contains(string(raw), "CANARY") {
		t.Fatal(string(raw))
	}
}
