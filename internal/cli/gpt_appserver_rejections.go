package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

type managedGPTRejectionLogger struct {
	mu        sync.Mutex
	dir, name string
}

func newManagedGPTRejectionLogger(profile, conversationID string) (*managedGPTRejectionLogger, error) {
	if err := codexapp.ValidatePrivatePath(profile, true); err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(conversationID))
	return &managedGPTRejectionLogger{dir: profile, name: "moai-rejections-" + hex.EncodeToString(digest[:]) + ".jsonl"}, nil
}

// RecordGatewayRejection never records request text or provider exceptions.
// A full log retains its existing evidence instead of growing past one MiB.
func (l *managedGPTRejectionLogger) RecordGatewayRejection(fields map[string]string) {
	switch fields["cause"] {
	case "history_changed", "agent_summary_untrusted", "history_scope_mismatch", "receipt_manifest_mismatch", "resume_duplicate_input", "history_receipt_missing", "history_prefix_mismatch", "appserver_scope_mismatch", "appserver_recovery_required", "appserver_protocol_error", "appserver_limit_exceeded", "appserver_event_queue_bytes_exceeded", "appserver_authority_changed", "appserver_request_canceled", "appserver_request_timeout", "appserver_transport_error", "native_receipt_authorization_required":
	default:
		return
	}
	switch fields["route"] {
	case "gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna":
	default:
		return
	}
	if digest, err := hex.DecodeString(fields["digest"]); err != nil || len(digest) != 32 {
		return
	}
	row := map[string]string{"at": time.Now().UTC().Format(time.RFC3339Nano), "cause": fields["cause"], "route": fields["route"], "digest": fields["digest"]}
	if fields["cause"] == "appserver_transport_error" {
		if code, err := strconv.ParseInt(fields["rpc_code"], 10, 32); err == nil && strconv.FormatInt(code, 10) == fields["rpc_code"] {
			row["rpc_code"] = fields["rpc_code"]
		}
	}
	if fields["cause"] == "appserver_scope_mismatch" {
		for _, key := range []string{"summary_classified", "agent_header_present", "summary_prompt_match"} {
			switch fields[key] {
			case "true", "false":
				row[key] = fields[key]
			}
		}
		switch fields["last_message_role"] {
		case "user", "assistant", "system", "missing", "unknown":
			row["last_message_role"] = fields["last_message_role"]
		}
		switch fields["last_content_type"] {
		case "string", "text", "tool_result", "tool_use", "image", "thinking", "redacted_thinking", "missing", "unknown":
			row["last_content_type"] = fields["last_content_type"]
		}
		switch fields["reason"] {
		case "active_model_mismatch", "working_directory_mismatch", "history_prefix_mismatch", "input_while_tools_pending", "tool_result_count_mismatch", "unknown_or_duplicate_tool_result", "unsupported_tool_result_content", "empty_tool_result_image", "unexpected_tool_result", "empty_turn_input", "prepare_owner_binding", "prepare_owner_length", "prepare_active_model_or_cwd", "prepare_public_delta", "unclassified_scope":
			row["reason"] = fields["reason"]
		}
	}
	raw, err := json.Marshal(row)
	if err != nil {
		return
	}
	raw = append(raw, '\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	if codexapp.ValidatePrivatePath(l.dir, true) != nil {
		return
	}
	root, err := os.OpenRoot(l.dir)
	if err != nil {
		return
	}
	defer func() { _ = root.Close() }() // Logging is best-effort and never changes the public rejection.
	info, err := root.Lstat(l.name)
	var file *os.File
	if errors.Is(err, os.ErrNotExist) {
		// O_EXCL rejects a concurrent symlink or existing file before opening.
		file, err = root.OpenFile(l.name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	} else if err == nil && info.Mode().IsRegular() && info.Mode().Perm() == 0600 {
		file, err = root.OpenFile(l.name, os.O_WRONLY|os.O_APPEND, 0600)
	} else {
		return
	}
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || opened.Mode().Perm() != 0600 || (info != nil && !os.SameFile(info, opened)) || opened.Size()+int64(len(raw)) > 1<<20 {
		return
	}
	_, _ = file.Write(raw)
}
