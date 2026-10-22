package cli

// @MX:ANCHOR: [AUTO] moai paths — the one confirmed source of the project key and database paths.
// @MX:REASON: [AUTO] every instruction surface (agent definitions, skills, README, docs-site, help texts)
// points readers here instead of letting them guess a path; its JSON keys and --field names are the contract
// the foreman watch and the doctor ghost check consume (SPEC-MOAI-DB-PATHS-001).
// @MX:SPEC: SPEC-MOAI-DB-PATHS-001

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Resolver seams: the production values are the home-state resolvers the todo
// and factory verbs use once the project root is fixed. The root itself is
// resolved by this command (CLAUDE_PROJECT_DIR first, then the launch
// directory) — the same rule the todo verbs and the factory-opening path
// through factoryCardRoot() apply, so every surface reports one project.
// Tests replace the seams to simulate an unresolvable moai home; the command
// never computes a project key or joins path segments of its own.
var (
	pathsProjectDir    = homestate.ProjectDir
	pathsFactoryDBPath = homestate.FactoryDBPath
)

// pathsRunSentinel is a valid run id that stands in for the real one while the
// broker path is obtained from its resolver; it is rendered as <run-id>.
const pathsRunSentinel = "zz-run-id-zz"

// pathsFields are the closed set of --field names, in listing order.
var pathsFields = []string{"project-key", "project-dir", "todo-dir", "todo-db", "factory-db"}

// pathsCompanion describes a -wal or -shm file beside a database.
type pathsCompanion struct {
	Exists    bool  `json:"exists"`
	SizeBytes int64 `json:"size_bytes"`
}

// pathsDB is the stat report of one database file and its companions.
type pathsDB struct {
	Path      string         `json:"path"`
	Exists    bool           `json:"exists"`
	SizeBytes int64          `json:"size_bytes"`
	WAL       pathsCompanion `json:"wal"`
	SHM       pathsCompanion `json:"shm"`
}

// pathsTodo is the queue entry: a database report plus its directory and state.
type pathsTodo struct {
	Dir        string  `json:"dir"`
	QueueState string  `json:"queue_state"`
	LegacyDir  *string `json:"legacy_dir"`
	pathsDB
}

// pathsBrokerTemplate is the broker entry when no --run is given.
type pathsBrokerTemplate struct {
	PathTemplate string `json:"path_template"`
	ExistingRuns *int   `json:"existing_runs"`
}

// pathsReport is the whole report. A nil pointer is an unresolved field.
type pathsReport struct {
	ProjectKey  *string    `json:"project_key"`
	ProjectDir  *string    `json:"project_dir"`
	QueueLayout string     `json:"queue_layout"`
	Todo        *pathsTodo `json:"todo"`
	Factory     *pathsDB   `json:"factory"`
	Broker      any        `json:"broker"`
}

// newPathsCmd builds `moai paths`: a read-only report of the project key and
// the queue, factory, and broker database paths. Once the project root is
// fixed (CLAUDE_PROJECT_DIR first, then the launch directory) the paths come
// from the home-state resolvers the owning verbs use; moai factory runs and
// moai factory handoff recover-resume fix the root from the working directory
// instead (SPEC-MOAI-DB-PATHS-001, card t1827).
func newPathsCmd() *cobra.Command {
	var jsonOutput bool
	var field, run string
	cmd := &cobra.Command{
		Use:   "paths",
		Short: "Print the confirmed database paths of this project",
		Long: `Print the project key and the confirmed paths of the queue, factory, and
broker databases. The project root is CLAUDE_PROJECT_DIR first, then the
launch directory; moai todo, the factory card root, and the factory verbs
that open the factory all use the same rule.

Use this command instead of guessing a database path: opening a missing file
with sqlite3 creates an empty one. The command only reads: it creates no file,
opens no database, takes no lock, and moves nothing.

  moai paths                  human listing, one name: value line per field
  moai paths --json           one JSON object on stdout
  moai paths --field NAME     one bare value; NAME is project-key, project-dir,
                              todo-dir, todo-db, or factory-db
  moai paths --run RUN-ID     the broker database path of that factory run

The queue is reported at the canonical location the adopting moai todo verbs
leave it in; a queue still sitting in a legacy directory shows queue_state
pre-adoption and names that directory in legacy_dir. A field that cannot be
resolved is reported as null and the command exits non-zero.`,
		GroupID: "tools",
		Args:    cobra.NoArgs,
		// Usage text on stdout would break the exact-output contract of --json and --field.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fieldSet := cmd.Flags().Changed("field")
			if fieldSet && !pathsKnownField(field) {
				return fmt.Errorf("paths: unknown --field %q (valid: %s)", field, strings.Join(pathsFields, ", "))
			}
			rep, unresolved, runInvalid := collectPaths(resolveProjectDir(), cmd.Flags().Changed("run"), run)
			if runInvalid != nil {
				return runInvalid
			}
			out := cmd.OutOrStdout()
			switch {
			case fieldSet:
				value, ok := pathsFieldValue(rep, field)
				if !ok {
					return fmt.Errorf("paths: %s cannot be resolved", field)
				}
				_, _ = fmt.Fprintln(out, value)
				return nil // only the requested field has to resolve
			case jsonOutput:
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(rep); err != nil {
					return fmt.Errorf("paths: encode: %w", err)
				}
			default:
				renderPathsListing(out, rep)
			}
			if len(unresolved) > 0 {
				return fmt.Errorf("paths: cannot resolve %s", strings.Join(unresolved, ", "))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "write exactly one JSON object to stdout")
	cmd.Flags().StringVar(&field, "field", "", "write only this field's bare value ("+strings.Join(pathsFields, ", ")+")")
	cmd.Flags().StringVar(&run, "run", "", "report the broker database path of this factory run id")
	cmd.MarkFlagsMutuallyExclusive("json", "field")
	return cmd
}

func pathsKnownField(name string) bool {
	for _, f := range pathsFields {
		if f == name {
			return true
		}
	}
	return false
}

// collectPaths resolves every reported path from launch. A non-nil runInvalid
// is an invalid --run id: nothing may be reported for it. unresolved names the
// fields that could not be resolved.
func collectPaths(launch string, runSet bool, runID string) (rep pathsReport, unresolved []string, runInvalid error) {
	root := factory.ResolveTodoQueueRoot(launch)

	key := homestate.ProjectKey(root)
	rep.ProjectKey = &key

	if dir, err := pathsProjectDir(root); err == nil {
		rep.ProjectDir = &dir
	} else {
		unresolved = append(unresolved, "project_dir")
	}

	queue := factory.QueueLocationForRoot(root)
	rep.QueueLayout = queue.Layout
	rep.Todo = &pathsTodo{Dir: queue.Dir, QueueState: "canonical", pathsDB: statPathsDB(queue.DBPath)}
	if queue.LegacyDir != "" {
		legacy := queue.LegacyDir
		rep.Todo.QueueState = "pre-adoption"
		rep.Todo.LegacyDir = &legacy
	}

	// An invalid --run id reports nothing, even when the home cannot be
	// resolved; the id shape is checked before anything is resolved.
	if runSet {
		if !factorymsg.ValidRunID(runID) {
			return rep, nil, fmt.Errorf("paths: invalid factory run id %q", runID)
		}
	}

	factoryPath, err := pathsFactoryDBPath(root)
	if err != nil {
		unresolved = append(unresolved, "factory")
		return rep, append(unresolved, "broker"), nil
	}
	fdb := statPathsDB(factoryPath)
	rep.Factory = &fdb

	if runSet {
		brokerPath, err := factorymsg.BrokerPath(root, runID)
		if err != nil {
			return rep, nil, fmt.Errorf("paths: %w", err)
		}
		rep.Broker = statPathsDB(brokerPath)
		return rep, unresolved, nil
	}
	rep.Broker = brokerTemplate(root)
	if rep.Broker == nil {
		unresolved = append(unresolved, "broker")
	}
	return rep, unresolved, nil
}

// brokerTemplate renders the broker path with <run-id> in place of the run id
// and counts the run directories found beside it. The path comes from the
// broker's own resolver; nil means it could not be resolved.
func brokerTemplate(root string) any {
	sample, err := factorymsg.BrokerPath(root, pathsRunSentinel)
	if err != nil {
		return nil
	}
	tmpl := sample
	if i := strings.LastIndex(sample, pathsRunSentinel); i >= 0 {
		tmpl = sample[:i] + "<run-id>" + sample[i+len(pathsRunSentinel):]
	}
	return pathsBrokerTemplate{PathTemplate: tmpl, ExistingRuns: countRunDirs(sample)}
}

// countRunDirs counts the run directories beside sample's own: sample is
// <factory-dir>/messages/<id>/broker.db, so its grandparent holds one
// directory per run. nil means the directory could not be read.
func countRunDirs(sample string) *int {
	entries, err := os.ReadDir(filepath.Dir(filepath.Dir(sample)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			zero := 0
			return &zero
		}
		return nil
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			n++
		}
	}
	return &n
}

// statPathsDB reads existence and size of a database and its -wal/-shm
// companions by stat only; it never opens any of them.
func statPathsDB(path string) pathsDB {
	db := pathsDB{Path: path}
	db.Exists, db.SizeBytes = statRegular(path)
	db.WAL.Exists, db.WAL.SizeBytes = statRegular(path + "-wal")
	db.SHM.Exists, db.SHM.SizeBytes = statRegular(path + "-shm")
	return db
}

func statRegular(path string) (bool, int64) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false, 0
	}
	return true, info.Size()
}

// pathsFieldValue returns the bare value of a --field name; ok is false when
// that field is unresolved.
func pathsFieldValue(rep pathsReport, name string) (string, bool) {
	switch name {
	case "project-key":
		return derefOrEmpty(rep.ProjectKey)
	case "project-dir":
		return derefOrEmpty(rep.ProjectDir)
	case "todo-dir":
		return rep.Todo.Dir, true
	case "todo-db":
		return rep.Todo.Path, true
	case "factory-db":
		if rep.Factory == nil {
			return "", false
		}
		return rep.Factory.Path, true
	}
	return "", false
}

func derefOrEmpty(s *string) (string, bool) {
	if s == nil {
		return "", false
	}
	return *s, true
}

// renderPathsListing writes the human listing: one `name: value` line per
// field, the five --field names first and in that order.
func renderPathsListing(out io.Writer, rep pathsReport) {
	line := func(name, value string) { _, _ = fmt.Fprintf(out, "%s: %s\n", name, value) }
	orUnresolved := func(s *string) string {
		if s == nil {
			return "unresolved"
		}
		return *s
	}
	line("project-key", orUnresolved(rep.ProjectKey))
	line("project-dir", orUnresolved(rep.ProjectDir))
	line("todo-dir", rep.Todo.Dir)
	line("todo-db", rep.Todo.Path)
	if rep.Factory != nil {
		line("factory-db", rep.Factory.Path)
	} else {
		line("factory-db", "unresolved")
	}
	line("todo-db-stat", statSummary(rep.Todo.pathsDB))
	if rep.Factory != nil {
		line("factory-db-stat", statSummary(*rep.Factory))
	}
	line("queue-state", rep.Todo.QueueState)
	if rep.Todo.LegacyDir != nil {
		line("legacy-dir", *rep.Todo.LegacyDir)
	} else {
		line("legacy-dir", "none")
	}
	line("queue-layout", rep.QueueLayout)
	switch b := rep.Broker.(type) {
	case pathsBrokerTemplate:
		line("broker-template", b.PathTemplate)
		if b.ExistingRuns != nil {
			line("broker-runs", fmt.Sprint(*b.ExistingRuns))
		} else {
			line("broker-runs", "unreadable")
		}
	case pathsDB:
		line("broker-db", b.Path)
		line("broker-db-stat", statSummary(b))
	default:
		line("broker-template", "unresolved")
	}
}

func statSummary(db pathsDB) string {
	return fmt.Sprintf("exists=%t size_bytes=%d wal_exists=%t wal_size_bytes=%d shm_exists=%t shm_size_bytes=%d",
		db.Exists, db.SizeBytes, db.WAL.Exists, db.WAL.SizeBytes, db.SHM.Exists, db.SHM.SizeBytes)
}
