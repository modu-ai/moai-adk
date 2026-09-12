package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBrokerRegressionBrokerProcess(t *testing.T) {
	if !strings.Contains(strings.Join(os.Args, " "), "-- app-server") {
		return
	}
	home := os.Getenv("CODEX_HOME")
	modeRaw, _ := os.ReadFile(filepath.Join(home, "audit-mode"))
	mode := string(modeRaw)
	os.WriteFile(filepath.Join(home, "audit-pid"), []byte(strconv.Itoa(os.Getpid())), 0600)
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var message struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		json.Unmarshal(scan.Bytes(), &message)
		switch message.Method {
		case "initialize":
			fmt.Println(`{"id":1,"result":{}}`)
		case "account/login/start":
			id := "current"
			if mode == "blocked-cancel" {
				id = strings.Repeat("x", 256<<10)
			}
			json.NewEncoder(os.Stdout).Encode(map[string]any{"id": 2, "result": map[string]string{"type": "chatgpt", "authUrl": "https://auth.openai.com/authorize", "loginId": id}})
			if mode == "blocked-cancel" {
				time.Sleep(20 * time.Second)
				os.Exit(0)
			}
			tokenFixture(t, home, "audit", time.Now().Add(time.Hour))
			fmt.Println(`{"method":"account/login/completed","params":{"loginId":"current","success":true}}`)
		}
	}
	if mode == "forced-success" {
		time.Sleep(20 * time.Second)
	}
	os.Exit(0)
}

func TestBrokerRegressionBrokerCancellationCannotBlockOnPipe(t *testing.T) {
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	home, e := privateBrokerTestHome(t)
	if e != nil {
		t.Fatal(e)
	}
	os.Chmod(home, 0700)
	os.WriteFile(filepath.Join(home, "audit-mode"), []byte("blocked-cancel"), 0600)
	var process *os.Process
	t.Cleanup(func() {
		if process != nil {
			process.Kill()
		}
	})
	b := CodexBroker{Executable: executable, Timeout: 150 * time.Millisecond, OnLogin: func(LoginPrompt) error { return nil }, testEnv: []string{"GORACE=atexit_sleep_ms=0"}, testArgs: []string{"-test.run=^TestBrokerRegressionBrokerProcess$", "--"}}
	done := make(chan error, 1)
	go func() { done <- b.Run(context.Background(), home, false) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("timeout returned success")
		}
		t.Log("bounded cancellation returned an error")
	case <-time.After(1500 * time.Millisecond):
		raw, _ := os.ReadFile(filepath.Join(home, "audit-pid"))
		pid, _ := strconv.Atoi(string(raw))
		process, _ = os.FindProcess(pid)
		if process != nil {
			process.Kill()
		}
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("broker did not return even after cleanup kill")
		}
		t.Error("broker timeout elapsed but cancellation remained blocked writing loginId to unread stdin; explicit cleanup kill was required")
	}
}

func TestBrokerRegressionForcedBrokerExitCannotPublishSuccess(t *testing.T) {
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	b := CodexBroker{Executable: executable, Timeout: time.Second, OnLogin: func(LoginPrompt) error { return nil }, testEnv: []string{"GORACE=atexit_sleep_ms=0"}, testArgs: []string{"-test.run=^TestBrokerRegressionBrokerProcess$", "--"}}
	_, e = s.Login(context.Background(), brokerFunc(func(ctx context.Context, home string, refresh bool) error {
		os.WriteFile(filepath.Join(home, "audit-mode"), []byte("forced-success"), 0600)
		return b.Run(ctx, home, refresh)
	}))
	status, statusErr := s.Status(context.Background())
	if statusErr != nil {
		t.Fatal(statusErr)
	}
	if e == nil || status.LoggedIn {
		t.Error("broker killed after missing graceful exit was accepted and its credential published")
	}
}
