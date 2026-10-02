package kanban

import (
	"reflect"
	"strings"
	"testing"
)

func TestRelaunchProviderForBackend(t *testing.T) {
	for backend, want := range map[string]string{
		"glm":     RelaunchProviderGLM,
		"gpt":     RelaunchProviderCodex,
		"claude":  RelaunchProviderCC,
		"":        RelaunchProviderCC,
		"unknown": RelaunchProviderCC,
	} {
		if got := RelaunchProviderForBackend(backend); got != want {
			t.Errorf("RelaunchProviderForBackend(%q) = %q, want %q", backend, got, want)
		}
	}
}

func TestRelaunchCommandLineAndLaunchLine(t *testing.T) {
	cases := []struct {
		name       string
		cmd        RelaunchCommand
		wantLine   string
		wantLaunch string
	}{
		{"cc bare", RelaunchCommand{Provider: "cc"}, "moai factory relaunch --provider cc", "moai cc -f lane"},
		{"cc lane", RelaunchCommand{Provider: "cc", Lane: "lane-3"}, "moai factory relaunch --provider cc --lane lane-3", "moai cc -f lane-3"},
		{"glm lane run", RelaunchCommand{Provider: "glm", Lane: "lane-3", Run: "runA"}, "moai factory relaunch --provider glm --lane lane-3 --run runA", "moai glm -f lane-3 --factory-run runA"},
		{"cc run", RelaunchCommand{Provider: "cc", Run: "runA"}, "moai factory relaunch --provider cc --run runA", "moai cc -f lane --factory-run runA"},
		{"cc from-run", RelaunchCommand{Provider: "cc", FromRun: "runX"}, "moai factory relaunch --provider cc --from-run runX", "moai cc -f lane"},
		{"flag order", RelaunchCommand{Provider: "cc", Lane: "lane-3", Run: "runA", FromRun: "runX"}, "moai factory relaunch --provider cc --lane lane-3 --run runA --from-run runX", "moai cc -f lane-3 --factory-run runA"},
		// REQ-SRH-016: the codex entry takes no lane pin and no run flag.
		{"codex drops lane and run", RelaunchCommand{Provider: "codex", Lane: "lane-3", Run: "runA"}, "moai factory relaunch --provider codex", "moai codex -f lane"},
		{"codex keeps from-run", RelaunchCommand{Provider: "codex", FromRun: "runX"}, "moai factory relaunch --provider codex --from-run runX", "moai codex -f lane"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cmd.Line(); got != tc.wantLine {
				t.Errorf("Line() = %q, want %q", got, tc.wantLine)
			}
			if got := tc.cmd.LaunchLine(); got != tc.wantLaunch {
				t.Errorf("LaunchLine() = %q, want %q", got, tc.wantLaunch)
			}
			if got, want := "moai "+strings.Join(tc.cmd.LaunchArgs(), " "), tc.wantLaunch; got != want {
				t.Errorf("LaunchArgs() joined = %q, want %q", got, want)
			}
			if strings.ContainsAny(tc.cmd.Line(), "<>") {
				t.Errorf("Line() %q carries an angle-bracket placeholder", tc.cmd.Line())
			}
		})
	}
}

// TestRelaunchNoticeTable pins spec.md §D.7 row by row, for provider cc and
// provider codex.
func TestRelaunchNoticeTable(t *testing.T) {
	const verb = "moai factory relaunch --provider "
	cases := []struct {
		name  string
		state RelaunchNoticeState
		want  []string
		more  int
	}{
		{"R1 current active", RelaunchNoticeState{Provider: "cc", Lane: "lane-3", Run: "runX", RunActive: true}, nil, 0},
		{"R2 current rebound", RelaunchNoticeState{Provider: "cc", Lane: "lane-3", Run: "runX", ActiveRuns: []string{"runY"}}, nil, 0},
		{"R3 current slot held", RelaunchNoticeState{Provider: "cc", Lane: "lane-3", Run: "runX", ActiveRuns: []string{"runY"}, SlotHeldByLiveOther: true}, []string{verb + "cc --run runY"}, 0},
		{"R4 current unbound", RelaunchNoticeState{Provider: "cc", Lane: "lane-3", Run: "runX"}, nil, 0},
		{"R5 current ambiguous", RelaunchNoticeState{Provider: "cc", Lane: "lane-3", Run: "runX", ActiveRuns: []string{"runZ", "runY"}},
			[]string{verb + "cc --lane lane-3 --run runY", verb + "cc --lane lane-3 --run runZ"}, 0},
		{"R5 four candidates cap at three", RelaunchNoticeState{Provider: "cc", Lane: "lane-3", Run: "runX", ActiveRuns: []string{"d", "c", "b", "a"}},
			[]string{verb + "cc --lane lane-3 --run a", verb + "cc --lane lane-3 --run b", verb + "cc --lane lane-3 --run c"}, 1},
		{"R5 codex single line", RelaunchNoticeState{Provider: "codex", Lane: "lane-3", Run: "runX", ActiveRuns: []string{"runY", "runZ"}}, []string{verb + "codex"}, 0},
		{"R3 codex", RelaunchNoticeState{Provider: "codex", Lane: "lane-3", Run: "runX", ActiveRuns: []string{"runY"}, SlotHeldByLiveOther: true}, []string{verb + "codex"}, 0},
		{"R6 legacy active", RelaunchNoticeState{Provider: "cc", Legacy: true, Run: "runX", RunActive: true}, []string{verb + "cc --from-run runX"}, 0},
		{"R6 legacy active codex", RelaunchNoticeState{Provider: "codex", Legacy: true, Run: "runX", RunActive: true}, []string{verb + "codex --from-run runX"}, 0},
		{"R7 legacy none", RelaunchNoticeState{Provider: "cc", Legacy: true, Run: "runX"}, nil, 0},
		{"R8 legacy one", RelaunchNoticeState{Provider: "cc", Legacy: true, Run: "runX", ActiveRuns: []string{"runY"}}, []string{verb + "cc"}, 0},
		{"R9 legacy several", RelaunchNoticeState{Provider: "glm", Legacy: true, Run: "runX", ActiveRuns: []string{"runZ", "runY"}},
			[]string{verb + "glm --run runY", verb + "glm --run runZ"}, 0},
		{"R9 legacy codex", RelaunchNoticeState{Provider: "codex", Legacy: true, Run: "runX", ActiveRuns: []string{"runY", "runZ"}}, []string{verb + "codex"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]string(nil), tc.state.ActiveRuns...)
			got := RelaunchNoticeFor(tc.state)
			if !reflect.DeepEqual(got.Lines, tc.want) && (len(got.Lines) != 0 || len(tc.want) != 0) {
				t.Errorf("Lines = %q, want %q", got.Lines, tc.want)
			}
			if got.More != tc.more {
				t.Errorf("More = %d, want %d", got.More, tc.more)
			}
			if !reflect.DeepEqual(tc.state.ActiveRuns, before) {
				t.Errorf("RelaunchNoticeFor mutated its input ActiveRuns: %q -> %q", before, tc.state.ActiveRuns)
			}
			for _, line := range got.Lines {
				if strings.ContainsAny(line, "<>") {
					t.Errorf("line %q carries an angle-bracket placeholder", line)
				}
			}
		})
	}
}
