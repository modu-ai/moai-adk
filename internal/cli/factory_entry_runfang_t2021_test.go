package cli

import (
	"bytes"
	"context"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Card t2021 round 2 (operator decisions R437/R438): the launcher keeps two entry
// flags, -f (leader) and -l / --lane (lane). -a, -a=x, --agent and --agent=x are
// refused, pointing at -l / --lane; after a literal "--" everything is forwarded
// unchanged except --agent / --agent=x, which are refused pointing at --lane.
//
// Every row goes through the real runFang path (cobra -> fang -> moaiErrorHandler),
// with the launch seam installed so an accepted row records the backend argv and
// environment instead of starting a session. A refused row must record nothing.

type t2021EntryRow struct {
	name    string
	args    []string // everything after "moai"
	want    string   // refusal rows: text the printed error must contain
	notWant string   // text that must not appear (re-cased flag, wrong pointer)

	// launch rows (want == ""):
	worker  string   // MOAI_FACTORY_WORKER at launch ("" for a leader / no role)
	workers string   // MOAI_FACTORY_WORKERS at launch
	argsHas []string // tokens the backend argv must contain, in order of appearance
	argsNot []string // tokens the backend argv must not contain
}

func TestT2021_EntryFlagsThroughRunFang(t *testing.T) {
	const removedAgent = "flag was removed; use -l (or --lane) to join"
	const removedAgentAfterDashes = "flag was removed; use --lane"
	const oneEntry = "a launch carries at most one entry token; -f/--factory and -l/--lane are two entry tokens"
	const takesNoArg = "-l/--lane flag takes no argument"

	rows := []t2021EntryRow{
		// 1. -f unchanged; -l and --lane enter lane mode (what -a did before R437).
		{name: "cc -f", args: []string{"cc", "-f"}, worker: "", workers: "1", argsNot: []string{"-f"}},
		{name: "cc -l", args: []string{"cc", "-l"}, worker: "lane-1", workers: "0", argsHas: []string{"--name", "lane-1"}, argsNot: []string{"-l"}},
		{name: "cc --lane", args: []string{"cc", "--lane"}, worker: "lane-1", workers: "0", argsHas: []string{"--name", "lane-1"}, argsNot: []string{"--lane"}},
		{name: "glm -l", args: []string{"glm", "-l"}, worker: "lane-1", workers: "0", argsHas: []string{"lane-1"}, argsNot: []string{"-l"}},
		{name: "glm --lane", args: []string{"glm", "--lane"}, worker: "lane-1", workers: "0", argsHas: []string{"lane-1"}, argsNot: []string{"--lane"}},

		// 1b. -l / --lane / -l=x / --lane=x take no argument (the pre-R385 policy, kept).
		{name: "cc -l=x", args: []string{"cc", "-l=x"}, want: takesNoArg, notWant: "-L/--Lane"},
		{name: "cc --lane=x", args: []string{"cc", "--lane=x"}, want: takesNoArg, notWant: "-L/--Lane"},
		{name: "cc -l x", args: []string{"cc", "-l", "x"}, want: takesNoArg, notWant: "-L/--Lane"},
		{name: "glm --lane=x", args: []string{"glm", "--lane=x"}, want: takesNoArg, notWant: "-L/--Lane"},

		// 2. -a and --agent are removed and never start another role.
		{name: "cc -a", args: []string{"cc", "-a"}, want: "The -a " + removedAgent, notWant: "-A flag"},
		{name: "cc -a=x", args: []string{"cc", "-a=x"}, want: "The -a " + removedAgent, notWant: "-A flag"},
		{name: "cc --agent", args: []string{"cc", "--agent"}, want: "The --agent " + removedAgent, notWant: "--Agent"},
		{name: "cc --agent=x", args: []string{"cc", "--agent=x"}, want: "The --agent " + removedAgent, notWant: "--Agent"},
		{name: "cc --agent reviewer", args: []string{"cc", "--agent", "reviewer"}, want: "The --agent " + removedAgent, notWant: "--Agent"},
		{name: "glm -a", args: []string{"glm", "-a"}, want: "The -a " + removedAgent, notWant: "-A flag"},
		{name: "glm --agent=x", args: []string{"glm", "--agent=x"}, want: "The --agent " + removedAgent, notWant: "--Agent"},

		// 3. After "--": forwarded unchanged, except --agent / --agent=x.
		{name: "cc -- --agent x", args: []string{"cc", "--", "--agent", "x"}, want: "The --agent " + removedAgentAfterDashes, notWant: "--Agent"},
		{name: "cc -- --agent=x", args: []string{"cc", "--", "--agent=x"}, want: "The --agent " + removedAgentAfterDashes, notWant: "--Agent"},
		{name: "cc -l -- --agent x", args: []string{"cc", "-l", "--", "--agent", "x"}, want: "The --agent " + removedAgentAfterDashes, notWant: "--Agent"},
		{name: "glm -- --agent=x", args: []string{"glm", "--", "--agent=x"}, want: "The --agent " + removedAgentAfterDashes, notWant: "--Agent"},
		{name: "cc -- --model x (control)", args: []string{"cc", "--", "--model", "x"}, worker: "", workers: "", argsHas: []string{"--", "--model", "x"}},
		{name: "cc -- -a (control)", args: []string{"cc", "--", "-a"}, worker: "", workers: "", argsHas: []string{"--", "-a"}},
		{name: "cc -l -- --model x (control)", args: []string{"cc", "-l", "--", "--model", "x"}, worker: "lane-1", workers: "0", argsHas: []string{"lane-1", "--", "--model", "x"}},

		// 4. -f with -l is a conflict; uppercase -L / -A keep passing through.
		{name: "cc -f -l", args: []string{"cc", "-f", "-l"}, want: oneEntry, notWant: "-F/--Factory"},
		{name: "cc -l -f", args: []string{"cc", "-l", "-f"}, want: oneEntry, notWant: "-F/--Factory"},
		{name: "cc -f --lane", args: []string{"cc", "-f", "--lane"}, want: oneEntry, notWant: "-F/--Factory"},
		{name: "cc --factory -l", args: []string{"cc", "--factory", "-l"}, want: oneEntry, notWant: "-F/--Factory"},
		{name: "cc -L (control)", args: []string{"cc", "-L"}, worker: "", workers: "", argsHas: []string{"-L"}},
		{name: "cc -A (control)", args: []string{"cc", "-A"}, worker: "", workers: "", argsHas: []string{"-A"}},

		// 5. Other refusals stay (retargeted text only).
		{name: "cc -k", args: []string{"cc", "-k"}, want: "-k/--kanban flag is retired", notWant: "-K/--Kanban"},
		{name: "cc -f=x", args: []string{"cc", "-f=x"}, want: "a lane joins with -l", notWant: "-F/--Factory"},
		{name: "cc -l --clear-policy bogus", args: []string{"cc", "-l", "--clear-policy", "bogus"}, want: "--clear-policy flag takes clear-each", notWant: "--Clear-Policy"},
		{name: "cc --clear-policy clear-each", args: []string{"cc", "--clear-policy", "clear-each"}, want: "composes with -l or --lane only", notWant: "--Clear-Policy"},
		{name: "cc -f --name lane-2", args: []string{"cc", "-f", "--name", "lane-2"}, want: "value does not select a lane", notWant: "--Name"},
		{name: "cc -l --name foo", args: []string{"cc", "-l", "--name", "foo"}, want: "-l/--lane flag already names the role", notWant: "-L/--Lane"},
		{name: "cc -f --leader leader-2", args: []string{"cc", "-f", "--leader", "leader-2"}, want: "composes with -l or --lane only", notWant: "--Leader"},
	}

	const run = "run-t2021-r2-entry"
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Setenv("MOAI_HOME", t.TempDir())
			t.Chdir(root)
			seedGitWorkingTree(t, root)
			clearFactoryTestEnv(t)
			c := installFactoryLaunchSeam(t)
			if err := recordFactoryRunStart(root, run, factory.BackendClaude, "", homestate.LaneCapacityDerived); err != nil {
				t.Fatalf("record factory run: %v", err)
			}

			var errOut bytes.Buffer
			rootCmd.SetErr(&errOut)
			rootCmd.SetOut(io.Discard)
			rootCmd.SetArgs(tc.args)
			t.Cleanup(func() {
				rootCmd.SetErr(nil)
				rootCmd.SetOut(nil)
				rootCmd.SetArgs(nil)
			})
			err := runFang(context.Background(), rootCmd)
			printed := errOut.String()

			if tc.want != "" {
				if err == nil {
					t.Fatalf("%v: runFang returned nil, want a refusal containing %q", tc.args, tc.want)
				}
				if !strings.Contains(strings.ToLower(printed), strings.ToLower(tc.want)) {
					t.Errorf("printed error lacks %q:\n%s", tc.want, printed)
				}
				if tc.notWant != "" && strings.Contains(printed, tc.notWant) {
					t.Errorf("printed error carries %q:\n%s", tc.notWant, printed)
				}
				if c.args != nil || c.worker != "" {
					t.Errorf("a refused entry must launch nothing, got args %v worker %q", c.args, c.worker)
				}
				return
			}

			if err != nil {
				t.Fatalf("%v: runFang = %v, want an accepted launch", tc.args, err)
			}
			if c.worker != tc.worker {
				t.Errorf("MOAI_FACTORY_WORKER at launch = %q, want %q", c.worker, tc.worker)
			}
			if tc.workers != "" && c.workers != tc.workers {
				t.Errorf("MOAI_FACTORY_WORKERS at launch = %q, want %q", c.workers, tc.workers)
			}
			for _, tok := range tc.argsHas {
				if !slices.Contains(c.args, tok) {
					t.Errorf("backend argv %v lacks %q", c.args, tok)
				}
			}
			for _, tok := range tc.argsNot {
				if slices.Contains(c.args, tok) {
					t.Errorf("backend argv %v must not carry %q", c.args, tok)
				}
			}
		})
	}
}
