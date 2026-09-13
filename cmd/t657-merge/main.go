// cmd/t657-merge — the ONE-OFF queue-merge tool for card t657
// (SPEC-TODO-QUEUE-HOME-MERGE-001, plan.md §F/§H).
//
// This is deliberately NOT a user-facing CLI verb: the SPEC constrains the
// merge to a one-off program reusing the internal/kanban store API. It
// addresses the real stores ONLY by explicit absolute path flags
// (REQ-TQM-018) and refuses relative paths.
//
// Modes:
//
//	-observe          read-only census of both stores (counts, id sets, last_seq)
//	-dry-run          backup + ordering evidence + merge decision, writes NO store bytes
//	-execute          the destructive merge write (M4) — gated, M4/M5 NOT authorized from
//	                  this delegation; the flag exists so the tool is complete, and the
//	                  rehearsal suite (todo_merge_procedure_test.go) is its proof
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

type storeCensus struct {
	Store         string         `json:"store"`
	Layout        string         `json:"layout"` // db / json / empty
	LiveByState   map[string]int `json:"live_by_state"`
	LiveTotal     int            `json:"live_total"`
	ArchivedTotal int            `json:"archived_total"`
	LastSeq       int            `json:"last_seq"`
	LiveIDs       []string       `json:"live_ids"`
	ArchivedIDs   []string       `json:"archived_ids"`
	ProjectUUID   string         `json:"project_uuid,omitempty"`
}

func census(dir string) (*storeCensus, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(abs) {
		return nil, fmt.Errorf("store dir must be absolute: %q", dir)
	}
	// backlog.json is the store's frozen compatibility name; the engine
	// derives its own artifact as the sibling backlog.db.
	queue := filepath.Join(abs, "backlog.json")
	store := kanban.NewBacklogStore(queue)
	rec, err := store.LoadPure()
	if err != nil {
		return nil, err
	}
	c := &storeCensus{
		Store:         abs,
		LiveByState:   map[string]int{},
		LiveIDs:       []string{},
		ArchivedIDs:   []string{},
		LiveTotal:     len(rec.Items),
		ArchivedTotal: len(rec.Archived),
		LastSeq:       rec.LastSeq,
	}
	if rec.ProjectUUID != nil {
		c.ProjectUUID = *rec.ProjectUUID
	}
	for _, it := range rec.Items {
		c.LiveByState[string(it.State)]++
		c.LiveIDs = append(c.LiveIDs, it.ID)
	}
	for _, e := range rec.Archived {
		c.ArchivedIDs = append(c.ArchivedIDs, e.Item.ID)
	}
	dbPath := strings.TrimSuffix(queue, ".json") + ".db"
	if _, err := os.Stat(dbPath); err == nil {
		c.Layout = "db"
	} else if _, err := os.Stat(queue); err == nil {
		c.Layout = "json"
	} else {
		c.Layout = "empty"
	}
	return c, nil
}

func writeTSV(path string, rows [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, row := range rows {
		b.WriteString(strings.Join(row, "\t"))
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func main() {
	home := flag.String("home", "", "home store DIRECTORY (absolute)")
	project := flag.String("project", "", "project store DIRECTORY (absolute)")
	backup := flag.String("backup", "", "backup output DIRECTORY (absolute)")
	dryRun := flag.Bool("dry-run", false, "backup + ordering evidence + merge decision; writes NO store bytes")
	observe := flag.Bool("observe", false, "read-only census of the addressed stores")
	execute := flag.Bool("execute", false, "DESTRUCTIVE: perform the merge write (M4 gate applies)")
	reportDir := flag.String("report", ".moai/reports/t657", "evidence report directory (authoring artifact)")
	flag.Parse()

	if *observe {
		dirs := []string{*home, *project}
		out := make([]*storeCensus, 0, len(dirs))
		for _, dir := range dirs {
			if dir == "" {
				continue
			}
			c, err := census(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "observe %s: %v\n", dir, err)
				os.Exit(1)
			}
			out = append(out, c)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "encode census: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *home == "" || *project == "" || *backup == "" {
		fmt.Fprintln(os.Stderr, "-home, -project, -backup are required")
		os.Exit(2)
	}
	paths := kanban.QueueMergePaths{HomeDir: *home, ProjectDir: *project, BackupDir: *backup}

	outcome, err := kanban.RunQueueMerge(paths, *dryRun || !*execute)
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge: %v\n", err)
		os.Exit(1)
	}

	// Evidence artifacts (authoring artifacts under the report dir).
	rep := *reportDir
	mappingRows := [][]string{{"old_id", "new_id"}}
	for _, row := range outcome.Report.Mapping() {
		mappingRows = append(mappingRows, []string{row.OldID, row.NewID})
	}
	if err := writeTSV(filepath.Join(rep, "id-mapping.tsv"), mappingRows); err != nil {
		fmt.Fprintf(os.Stderr, "write mapping: %v\n", err)
		os.Exit(1)
	}
	reconRows := [][]string{{"card_id", "kind", "detail"}}
	for _, r := range outcome.Report.Reconciliation {
		reconRows = append(reconRows, []string{r.CardID, r.Kind, r.Detail})
	}
	if err := writeTSV(filepath.Join(rep, "reconciliation.tsv"), reconRows); err != nil {
		fmt.Fprintf(os.Stderr, "write reconciliation: %v\n", err)
		os.Exit(1)
	}
	backupJSON, _ := json.MarshalIndent(outcome.Backup, "", "  ")
	_ = os.WriteFile(filepath.Join(rep, "backup-hashes.json"), backupJSON, 0o644)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"dry_run":      outcome.DryRun,
		"high_water":   outcome.Report.HighWater,
		"migrated":     len(outcome.Report.Migrated),
		"renumbered":   len(outcome.Report.Renumbered),
		"duplicates":   len(outcome.Report.Duplicates),
		"verification": outcome.Verification,
	})
}
