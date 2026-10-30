package cli

// AC-AFL-017 (SPEC-AGENTIC-FACTORY-LOCAL-001): the `moai af` command offers a verb for every step of the card
// flow: claim, stage move, evidence record, Enqueue, Release. The verb names are the ones the SPEC leaves open
// (OD-33); these are the names this lane chose, and the test pins them.

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func afVerb(af *cobra.Command, name string) *cobra.Command {
	for _, c := range af.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestAFCommand_VerbsRegistered(t *testing.T) {
	var af *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "af" {
			af = c
		}
	}
	if af == nil {
		t.Fatal("`moai af` is not registered on the root command")
	}
	if af.Hidden {
		t.Error("`moai af` must be visible: it is the card flow's command")
	}
	for verb, step := range map[string]string{
		"claim":     "claim",
		"move":      "stage move",
		"evidence":  "evidence record",
		"enqueue":   "Enqueue",
		"release":   "Release",
		"reap":      "reap",
		"land":      "land",
		"reconcile": "reconcile",
	} {
		c := afVerb(af, verb)
		if c == nil {
			t.Errorf("no `moai af %s` verb for the %s step", verb, step)
			continue
		}
		if c.RunE == nil && c.Run == nil {
			t.Errorf("`moai af %s` is registered but runs nothing", verb)
		}
		if strings.TrimSpace(c.Short) == "" {
			t.Errorf("`moai af %s` has no one-line description", verb)
		}
	}
}
