package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type LoginPrompt struct {
	URL     string
	LoginID string
}

// CodexBroker reuses the installed official broker in a fresh private home.
// Its protocol success is not provider token acceptance: Store.Refresh requires
// an independent verifier before publication. Live installed-version compatibility
// is a separate gate; this wrapper never reads a user's existing Codex home.
type CodexBroker struct {
	Executable string
	Timeout    time.Duration
	OnLogin    func(LoginPrompt) error
	testArgs   []string
	testEnv    []string
}
type rpcMessage struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
	Params json.RawMessage `json:"params"`
	Error  json.RawMessage `json:"error"`
}

func (b CodexBroker) Run(ctx context.Context, home string, refresh bool) error {
	operation := "login"
	if refresh {
		operation = "refresh"
	}
	return b.run(ctx, home, operation)
}

// Logout performs the official account/logout RPC in the supplied owned scratch.
// Its empty success response does not disclose the remote revocation outcome.
func (b CodexBroker) Logout(ctx context.Context, home string) error {
	return b.run(ctx, home, "logout")
}

// @MX:WARN: [AUTO] Broker stdout scanning and process reaping are joined before return.
// @MX:REASON: A token file is not safe to publish while the broker can still write it.
func (b CodexBroker) run(ctx context.Context, home string, operation string) (resultErr error) {
	if !filepath.IsAbs(b.Executable) || !filepath.IsAbs(home) || b.Timeout <= 0 {
		return ErrBroker
	}
	if privateDirectory(home) != nil {
		return ErrBroker
	}
	if operation == "login" && b.OnLogin == nil {
		return ErrBroker
	}
	for _, name := range []string{"config", "data", "cache", "tmp"} {
		if makePrivateDirectory(filepath.Join(home, name)) != nil {
			return ErrBroker
		}
	}
	// A clean CWD prevents project-local config discovery. All auth/config/home
	// overrides are constructed here, rather than subtracting keys from Environ.
	env := []string{"HOME=" + home, "USERPROFILE=" + home, "CODEX_HOME=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, "config"), "XDG_DATA_HOME=" + filepath.Join(home, "data"), "XDG_CACHE_HOME=" + filepath.Join(home, "cache"), "TMPDIR=" + filepath.Join(home, "tmp"), "TMP=" + filepath.Join(home, "tmp"), "TEMP=" + filepath.Join(home, "tmp"), "PATH=/usr/bin:/bin", "RUST_LOG=off"}
	if runtime.GOOS == "windows" {
		if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" {
			env = append(env, "SystemRoot="+systemRoot)
		}
	}
	args := append(append([]string{}, b.testArgs...), "app-server", "--listen", "stdio://", "-c", `cli_auth_credentials_store="file"`)
	cmd := exec.Command(b.Executable, args...)
	cmd.Dir = home
	cmd.Env = env
	cmd.Env = append(cmd.Env, b.testEnv...)
	cmd.Stderr = io.Discard
	stdin, e := cmd.StdinPipe()
	if e != nil {
		return ErrBroker
	}
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		_ = stdin.Close() // process never started; discarding the pipe
		return ErrBroker
	}
	if e = cmd.Start(); e != nil {
		_ = stdin.Close()  // start failed; discarding the pipes
		_ = stdout.Close() // start failed; discarding the pipes
		return ErrBroker
	}
	bounded, cancel := context.WithTimeout(ctx, b.Timeout)
	defer cancel()
	// Cancellation must not depend on the protocol loop reaching receive: any
	// JSON write can block when the broker stops reading its pipe.
	stopWatch := make(chan struct{})
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		select {
		case <-bounded.Done():
			_ = stdin.Close()
			_ = cmd.Process.Kill()
		case <-stopWatch:
		}
	}()
	defer func() {
		close(stopWatch)
		<-watchDone
		if bounded.Err() != nil {
			resultErr = ErrBroker
		}
	}()
	messages := make(chan rpcMessage, 8)
	scanDone := make(chan struct{})
	readCtx, stopRead := context.WithCancel(context.Background())
	go func() {
		defer close(scanDone)
		defer close(messages)
		scan := bufio.NewScanner(stdout)
		scan.Buffer(make([]byte, 4096), 1<<20)
		for scan.Scan() {
			var m rpcMessage
			if json.Unmarshal(scan.Bytes(), &m) != nil {
				return
			}
			select {
			case messages <- m:
			case <-readCtx.Done():
				return
			}
		}
		if scan.Err() != nil {
			return
		} // Receive treats all terminal read errors as broker failure.
	}()
	waitDone := make(chan error, 1)
	// Wait starts only during teardown, after protocol messages have been read;
	// os/exec.Wait otherwise closes StdoutPipe before the scanner finishes.
	defer func() {
		_ = stdin.Close() // teardown; the encoder has finished or the protocol already failed
		go func() { waitDone <- cmd.Wait() }()
		select {
		case waitErr := <-waitDone:
			if waitErr != nil {
				resultErr = ErrBroker
			}
		case <-time.After(250 * time.Millisecond):
			// A forced exit is never a successfully completed credential write.
			resultErr = ErrBroker
			_ = cmd.Process.Kill()
			<-waitDone
		}
		stopRead()
		_ = stdout.Close() // teardown discard
		<-scanDone
	}()
	encoder := json.NewEncoder(stdin)
	send := func(id int, method string, params any) error {
		return encoder.Encode(map[string]any{"id": id, "method": method, "params": params})
	}
	receive := func() (rpcMessage, error) {
		select {
		case <-bounded.Done():
			return rpcMessage{}, ErrBroker
		case m, ok := <-messages:
			if !ok || len(m.Error) > 0 && string(m.Error) != "null" {
				return rpcMessage{}, ErrBroker
			}
			return m, nil
		}
	}
	if send(1, "initialize", map[string]any{"clientInfo": map[string]string{"name": "moai-gateway-auth", "version": "0.1.0"}}) != nil {
		return ErrBroker
	}
	for {
		m, e := receive()
		if e != nil {
			return e
		}
		if m.ID == 1 {
			if len(m.Result) == 0 {
				return ErrBroker
			}
			break
		}
		if m.ID != 0 {
			return ErrBroker
		}
	}
	if encoder.Encode(map[string]any{"method": "initialized"}) != nil {
		return ErrBroker
	}
	if operation == "logout" {
		if send(2, "account/logout", nil) != nil {
			return ErrBroker
		}
		for {
			m, e := receive()
			if e != nil {
				return e
			}
			if m.ID == 0 {
				continue
			}
			if m.ID != 2 {
				return ErrBroker
			}
			var result map[string]json.RawMessage
			if json.Unmarshal(m.Result, &result) != nil || result == nil || len(result) != 0 {
				return ErrBroker
			}
			return nil
		}
	}
	if operation == "refresh" {
		if send(2, "account/read", map[string]bool{"refreshToken": true}) != nil {
			return ErrBroker
		}
		for {
			m, e := receive()
			if e != nil {
				return e
			}
			if m.ID == 0 {
				continue
			}
			if m.ID != 2 {
				return ErrBroker
			}
			var result struct {
				Account struct {
					Type string `json:"type"`
				} `json:"account"`
			}
			if json.Unmarshal(m.Result, &result) != nil || result.Account.Type != "chatgpt" {
				return ErrBroker
			}
			return nil
		}
	}
	if send(2, "account/login/start", map[string]string{"type": "chatgpt"}) != nil {
		return ErrBroker
	}
	var loginID string
	for {
		m, e := receive()
		if e != nil {
			return e
		}
		if m.ID == 0 {
			continue
		}
		if m.ID != 2 {
			return ErrBroker
		}
		var result struct {
			Type string `json:"type"`
			URL  string `json:"authUrl"`
			ID   string `json:"loginId"`
		}
		if json.Unmarshal(m.Result, &result) != nil || result.Type != "chatgpt" || result.ID == "" {
			return ErrBroker
		}
		parsed, e := url.Parse(result.URL)
		if e != nil || parsed.Scheme != "https" || parsed.Host != "auth.openai.com" || parsed.User != nil {
			return ErrBroker
		}
		loginID = result.ID
		if b.OnLogin(LoginPrompt{URL: result.URL, LoginID: loginID}) != nil {
			return ErrBroker
		}
		break
	}
	for {
		m, e := receive()
		if e != nil {
			_ = send(3, "account/login/cancel", map[string]string{"loginId": loginID})
			return e
		}
		if m.Method != "account/login/completed" {
			continue
		}
		var completion struct {
			ID      string `json:"loginId"`
			Success bool   `json:"success"`
		}
		if json.Unmarshal(m.Params, &completion) != nil || completion.ID != loginID || !completion.Success {
			return ErrBroker
		}
		return nil
	}
}
