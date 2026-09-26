package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const factoryAppTokenEnv = "MOAI_FACTORY_APP_SERVER_TOKEN"

type factoryAppReply struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *codexRPCError  `json:"error"`
}

type factoryAppClient struct {
	conn      *websocket.Conn
	events    chan factoryAppReply
	nextID    int
	busy      bool
	completed map[string]string
}

func (c *factoryAppClient) read() {
	defer close(c.events)
	for {
		var event factoryAppReply
		if err := c.conn.ReadJSON(&event); err != nil {
			return
		}
		// Streaming deltas can be numerous. The controller needs lifecycle and
		// RPC results; the TUI receives its own display stream.
		if event.ID == 0 && event.Method != "turn/started" && event.Method != "turn/completed" {
			continue
		}
		c.events <- event
	}
}

func (c *factoryAppClient) observe(event factoryAppReply) {
	switch event.Method {
	case "turn/started":
		c.busy = true
	case "turn/completed":
		c.busy = false
		var payload struct {
			Turn struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"turn"`
		}
		if json.Unmarshal(event.Params, &payload) == nil && payload.Turn.ID != "" {
			if c.completed == nil {
				c.completed = make(map[string]string)
			}
			c.completed[payload.Turn.ID] = payload.Turn.Status
		}
	}
}

func (c *factoryAppClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.nextID++
	id := c.nextID
	if err := c.conn.WriteJSON(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case event, ok := <-c.events:
			if !ok {
				return nil, errors.New("Factory Codex App Server connection closed")
			}
			c.observe(event)
			if event.ID != id {
				continue
			}
			if event.Error != nil {
				return nil, fmt.Errorf("%s: %s (%d)", method, event.Error.Message, event.Error.Code)
			}
			return event.Result, nil
		}
	}
}

func (c *factoryAppClient) waitTurn(ctx context.Context, turnID string) error {
	for {
		if status, ok := c.completed[turnID]; ok {
			delete(c.completed, turnID)
			if status != "completed" {
				return fmt.Errorf("Factory Codex turn %s ended as %s", turnID, status)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-c.events:
			if !ok {
				return errors.New("Factory Codex App Server connection closed")
			}
			c.observe(event)
		}
	}
}

func (c *factoryAppClient) startTurn(ctx context.Context, threadID, prompt string) error {
	result, err := c.call(ctx, "turn/start", map[string]any{
		"threadId": threadID,
		"input":    []map[string]string{{"type": "text", "text": prompt}},
	})
	if err != nil {
		return err
	}
	var reply struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(result, &reply); err != nil || reply.Turn.ID == "" {
		return errors.New("Factory Codex turn/start returned no turn id")
	}
	c.busy = true
	if _, completed := c.completed[reply.Turn.ID]; completed {
		c.busy = false
	}
	return c.waitTurn(ctx, reply.Turn.ID)
}

func factoryAppToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func factoryAppURL() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	address := listener.Addr().(*net.TCPAddr)
	if err := listener.Close(); err != nil {
		return "", err
	}
	return "ws://127.0.0.1:" + strconv.Itoa(address.Port), nil
}

func factoryAppReady(ctx context.Context, url string) error {
	client := http.Client{Timeout: 300 * time.Millisecond}
	endpoint := strings.Replace(url, "ws://", "http://", 1) + "/readyz"
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func managedCodexOptions(args []string) (appArgs []string, model string, err error) {
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "-c" || arg == "--config":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("Factory Codex %s requires a value", arg)
			}
			i++
			appArgs = append(appArgs, "-c", args[i])
		case strings.HasPrefix(arg, "--config="):
			appArgs = append(appArgs, "-c", strings.TrimPrefix(arg, "--config="))
		case arg == "-m" || arg == "--model":
			if i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" {
				return nil, "", fmt.Errorf("Factory Codex %s requires a model", arg)
			}
			i++
			model = args[i]
		case strings.HasPrefix(arg, "--model="):
			model = strings.TrimPrefix(arg, "--model=")
			if model == "" {
				return nil, "", errors.New("Factory Codex --model requires a model")
			}
		default:
			return nil, "", fmt.Errorf("Factory Codex managed launch does not support %s", arg)
		}
	}
	return appArgs, model, nil
}

// These overrides belong to the two Codex processes owned by this Factory
// session. Project and user config retain their ordinary approval behavior.
func factoryMoAIMCPApprovalArgs() []string {
	return []string{
		"-c", `mcp_servers.moai.default_tools_approval_mode="approve"`,
		"-c", `mcp_servers.moai.tools.factory_msg_send.approval_mode="approve"`,
		"-c", `mcp_servers.moai.tools.factory_msg_receipt.approval_mode="approve"`,
	}
}

// runManagedFactoryCodex owns both the App Server control connection and the
// interactive TUI. A separate connection can start turns on the same thread;
// the broker remains the source of the message body and receipt.
func runManagedFactoryCodex(req codexLaunchRequest) (err error) {
	appOptions, model, err := managedCodexOptions(req.Args)
	if err != nil {
		return err
	}
	root := launchProjectRoot()
	runID := os.Getenv(config.EnvMoaiKanbanID)
	ownerPID := os.Getpid()
	ownerStart := homestate.CurrentProcessFingerprint()
	if ownerStart == "" {
		return errors.New("Factory Codex owner identity unavailable")
	}
	env := withSessionPID(codexChildEnv(), ownerPID)
	pending, err := registerFactoryLaunchPending(context.Background(), root, env, ownerPID, ownerStart)
	if err != nil {
		return err
	}
	bound := false
	defer func() {
		if !bound {
			err = errors.Join(err, rollbackFactoryLaunchPending(context.Background(), root, pending))
		}
	}()
	token, err := factoryAppToken()
	if err != nil {
		return err
	}
	secretDir, err := os.MkdirTemp("", "moai-factory-app-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(secretDir)
	tokenFile := filepath.Join(secretDir, "token")
	if err := os.WriteFile(tokenFile, []byte(token), 0o600); err != nil {
		return err
	}
	url, err := factoryAppURL()
	if err != nil {
		return err
	}
	args := []string{"app-server", "--listen", url, "--ws-auth", "capability-token", "--ws-token-file", tokenFile}
	args = append(args, appOptions...)
	args = append(args, factoryMoAIMCPApprovalArgs()...)
	server := exec.Command(req.Program, args...)
	server.Dir, server.Env, server.Stderr = req.Dir, env, os.Stderr
	if err := server.Start(); err != nil {
		return err
	}
	defer func() {
		_ = server.Process.Kill()
		_ = server.Wait()
	}()
	readyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := factoryAppReady(readyCtx, url); err != nil {
		return fmt.Errorf("Factory Codex App Server readiness: %w", err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Authorization": []string{"Bearer " + token}})
	if err != nil {
		return err
	}
	defer conn.Close()
	client := &factoryAppClient{conn: conn, events: make(chan factoryAppReply, 32)}
	go client.read()
	if _, err := client.call(readyCtx, "initialize", map[string]any{
		"clientInfo": map[string]string{"name": "moai_factory", "title": "MoAI Factory", "version": "1"},
	}); err != nil {
		return err
	}
	if err := conn.WriteJSON(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		return err
	}
	threadParams := map[string]any{"cwd": req.Dir}
	if model != "" {
		threadParams["model"] = model
	}
	threadResult, err := client.call(readyCtx, "thread/start", threadParams)
	if err != nil {
		return err
	}
	var thread struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(threadResult, &thread); err != nil || thread.Thread.ID == "" {
		return errors.New("Factory Codex thread/start returned no thread id")
	}
	threadID := thread.Thread.ID
	label := launchEnvValue(env, config.EnvMoaiFactoryWorker)
	if label == "" {
		label = "lead"
	}
	if _, err := client.call(readyCtx, "thread/name/set", map[string]any{"threadId": threadID, "name": label}); err != nil {
		return err
	}
	// The remote TUI requires a persisted rollout before it can resume a
	// thread. Inject a control record without a model turn; this avoids paying
	// the project's entire instruction prefix merely to open the session.
	_, err = client.call(readyCtx, "thread/inject_items", map[string]any{"threadId": threadID, "items": []map[string]any{{
		"type": "message", "role": "assistant", "content": []map[string]string{{"type": "output_text", "text": "MoAI Factory " + label + " session ready (launcher)."}},
	}}})
	if err != nil {
		return err
	}
	store, err := factorymsg.Open(root, runID)
	if err != nil {
		return err
	}
	defer closeFactoryToolStore("managed_codex_factory", store)
	// App Server turns need not run a TUI SessionStart hook. The launcher
	// already owns the thread ID and process identity, so it binds the
	// provisional endpoint itself before admitting messages.
	peerCtx, peerCancel := context.WithTimeout(context.Background(), 5*time.Second)
	activePeer := pending
	activePeer.SessionUUID = threadID
	_, err = store.RegisterPeer(peerCtx, activePeer)
	peerCancel()
	if err != nil {
		return fmt.Errorf("Factory Codex session registration: %w", err)
	}
	bound = true
	tuiArgs := append([]string{"resume", "--remote", url, "--remote-auth-token-env", factoryAppTokenEnv}, req.Args...)
	tuiArgs = append(tuiArgs, factoryMoAIMCPApprovalArgs()...)
	tui := exec.Command(req.Program, append(tuiArgs, threadID)...)
	tui.Dir = req.Dir
	tui.Env = append(env, factoryAppTokenEnv+"="+token)
	tui.Stdin, tui.Stdout, tui.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := tui.Start(); err != nil {
		return err
	}
	tuiDone := make(chan error, 1)
	go func() { tuiDone <- tui.Wait() }()
	tuiFinished := false
	defer func() {
		if !tuiFinished {
			_ = tui.Process.Kill()
			<-tuiDone
		}
	}()
	ticker := time.NewTicker(managedFactoryPollInterval)
	defer ticker.Stop()
	var queued []factorymsg.Claim
	for {
		select {
		case err := <-tuiDone:
			tuiFinished = true
			return err
		case event, ok := <-client.events:
			if !ok {
				return errors.New("Factory Codex App Server connection closed")
			}
			client.observe(event)
		case <-ticker.C:
		}
		if client.busy {
			continue
		}
		if len(queued) == 0 {
			claims, claimErr := claimManagedFactoryInbox(store, ownerPID, ownerStart)
			if claimErr != nil {
				fmt.Fprintln(os.Stderr, "Factory Codex inbox:", claimErr)
				continue
			}
			queued = claims
		}
		if len(queued) == 0 {
			continue
		}
		prompt := managedFactoryInboxPrompt(runID, queued)
		turnCtx, turnCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		turnErr := client.startTurn(turnCtx, threadID, prompt)
		turnCancel()
		if turnErr != nil {
			fmt.Fprintln(os.Stderr, "Factory Codex turn:", turnErr)
			continue
		}
		queued = nil // broker receipts, not turn completion, prove processing
	}
}
