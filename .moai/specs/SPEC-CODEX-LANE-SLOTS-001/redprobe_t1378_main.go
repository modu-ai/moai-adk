//go:build ignore

// Command redprobe_t1378 is card t1378's committed plan-phase RED-now probe.
// It exercises only the exported API of internal/kanban and internal/homestate
// against a SYNTHETIC state root — a temp directory it creates and removes —
// never the live factory store. It is invisible to `./...` (the //go:build
// ignore tag plus its home under the dot-prefixed .moai tree) and is invoked
// directly:
//
//	go run .moai/specs/SPEC-CODEX-LANE-SLOTS-001/redprobe_t1378_main.go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

func main() {
	root, err := os.MkdirTemp("", "redprobe-t1378-*")
	if err != nil {
		fmt.Println("tempdir error:", err)
		os.Exit(1)
	}
	defer func() { _ = os.RemoveAll(root) }()
	alive := func(int) bool { return true }

	// RED-1 (AC-001/AC-009): a claude-shape (unbounded) join draws lane-1 for
	// run tm3yoq; the codex-shape bounded join (1..1) for the same run then
	// refuses instead of drawing lane-2.
	c1, err := kanban.ClaimFactoryLane(root, "", true, os.Getpid(), "tm3yoq", alive)
	if err != nil {
		fmt.Println("RED-1 first claim error:", err)
		os.Exit(1)
	}
	fmt.Println("RED-1 claude-shape claim:", c1.Label)
	_, err = kanban.ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "tm3yoq", 1, alive)
	fmt.Println("RED-1 codex-shape claim error:", err)

	// RED-2 (AC-006): the runs record carries no lane-capacity datum.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		fmt.Println("RED-2 open error:", err)
		os.Exit(1)
	}
	rows, err := db.DB.Query("PRAGMA table_info(runs)")
	if err != nil {
		fmt.Println("RED-2 pragma error:", err)
		_ = db.Close()
		os.Exit(1)
	}
	var cols []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			fmt.Println("RED-2 scan error:", err)
			_ = rows.Close()
			_ = db.Close()
			os.Exit(1)
		}
		cols = append(cols, name)
	}
	_ = rows.Close()
	_ = db.Close()
	fmt.Println("RED-2 runs table columns:", strings.Join(cols, ","))
}
