//go:build ignore

// Command redprobe_t1378 is card t1378's committed plan-phase RED-now probe.
// It exercises only the exported API of internal/kanban and internal/homestate
// against a SYNTHETIC state root — a temp directory it creates and removes —
// never the live factory store. It is invisible to `./...` (the //go:build
// ignore tag plus its home under the dot-prefixed .moai tree) and is invoked
// directly:
//
//	go run .moai/specs/SPEC-CODEX-LANE-SLOTS-001/redprobe_t1378_main.go
//
// M2 green-side flip (card t1378): the probe now records the run's declared
// capacity FIRST — a count-less start records the derived-capacity marker
// (REQ-004), the leader-start step the pre-M2 tree had no way to express —
// so the second half documents the repair: the bounded codex-shape claim,
// which refused with `factory run tm3yoq has no free lane slots in 1..1` on
// the pre-implementation tree (acceptance.md § Evidence Ledger RED-1), now
// grows past the live lane-1 and draws lane-2 (AC-001 green).
package main

import (
	"context"
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

	// The leader start records the run's declared capacity: no count was
	// supplied, so the derived-capacity marker goes on the record (REQ-004).
	db, err := homestate.OpenFactory(root)
	if err != nil {
		fmt.Println("capacity record open error:", err)
		os.Exit(1)
	}
	err = db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: "tm3yoq", Backend: "codex", ManifestJSON: "{}",
		LeadPID:      os.Getpid(),
		LaneCapacity: homestate.LaneCapacityDerived,
	})
	_ = db.Close()
	if err != nil {
		fmt.Println("capacity record error:", err)
		os.Exit(1)
	}

	// AC-001/AC-009: a claude-shape (unbounded) join draws lane-1 for
	// run tm3yoq; the codex-shape bounded join (the launcher's 1..1 guess)
	// then grows on the capacity-open run (REQ-005) and draws lane-2 —
	// the pre-implementation refusal no longer occurs.
	c1, err := kanban.ClaimFactoryLane(root, "", true, os.Getpid(), "tm3yoq", alive)
	if err != nil {
		fmt.Println("RED-1 first claim error:", err)
		os.Exit(1)
	}
	fmt.Println("RED-1 claude-shape claim:", c1.Label)
	c2, err := kanban.ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "tm3yoq", 1, alive)
	fmt.Println("RED-1 codex-shape claim error:", err)
	if err == nil {
		fmt.Println("RED-1 codex-shape claim label:", c2.Label)
	}

	// RED-2 (AC-006): the runs record now carries the lane-capacity datum
	// (pre-M1 the column list ended at lead_process_start).
	db, err = homestate.OpenFactory(root)
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
