package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestLogoutBrokerHelperProcess(t *testing.T) {
	if os.Getenv("MOAI_LOGOUT_HELPER") == "" {
		return
	}
	home := os.Getenv("CODEX_HOME")
	if home != os.Getenv("HOME") || os.Getenv("OPENAI_API_KEY") != "" {
		os.Exit(43)
	}
	if _, e := os.Stat(filepath.Join(home, "auth.json")); e != nil {
		os.Exit(44)
	}
	_ = os.WriteFile(filepath.Join(home, "pid"), []byte(strconv.Itoa(os.Getpid())), 0600)
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var m struct {
			Method string `json:"method"`
		}
		if json.Unmarshal(scan.Bytes(), &m) != nil {
			os.Exit(45)
		}
		switch m.Method {
		case "initialize":
			fmt.Println(`{"id":1,"result":{}}`)
		case "initialized":
		case "account/logout":
			switch os.Getenv("MOAI_LOGOUT_HELPER") {
			case "timeout":
				time.Sleep(30 * time.Second)
			case "error":
				fmt.Println(`{"id":2,"error":{"message":"private upstream error"}}`)
			case "wrong":
				fmt.Println(`{"id":3,"result":{}}`)
			case "null":
				fmt.Println(`{"id":2,"result":null}`)
			default:
				_ = os.Remove(filepath.Join(home, "auth.json"))
				fmt.Println(`{"id":2,"result":{}}`)
			}
		default:
			os.Exit(46)
		}
	}
	os.Exit(0)
}

func TestCodexBrokerLogoutRPCAndReaping(t *testing.T) {
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("OPENAI_API_KEY", "unrelated-private")
	for _, mode := range []string{"completed", "error", "wrong", "null", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			home, homeErr := privateBrokerTestHome(t)
			if homeErr != nil {
				t.Fatal(homeErr)
			}
			if e := os.Chmod(home, 0700); e != nil {
				t.Fatal(e)
			}
			tokenFixture(t, home, "owned", time.Now().Add(time.Hour))
			b := CodexBroker{Executable: executable, Timeout: 200 * time.Millisecond, testArgs: []string{"-test.run=^TestLogoutBrokerHelperProcess$", "--"}, testEnv: []string{"MOAI_LOGOUT_HELPER=" + mode, "GORACE=atexit_sleep_ms=0"}}
			e := b.Logout(context.Background(), home)
			if (e == nil) != (mode == "completed") {
				t.Fatal(mode, e)
			}
			raw, e := os.ReadFile(filepath.Join(home, "pid"))
			if e != nil {
				t.Fatal(e)
			}
			pid, e := strconv.Atoi(string(raw))
			if e != nil {
				t.Fatal(e)
			}
			p, e := os.FindProcess(pid)
			if e == nil {
				defer p.Release()
				if p.Signal(syscall.Signal(0)) == nil {
					t.Fatal("broker process still alive")
				}
			}
		})
	}
}
