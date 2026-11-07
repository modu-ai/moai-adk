package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// TestDriftOriginSentence_UsesMeasuredBaseCount pins card t2001 round-5 G1:
// the origin sentence states the base's count from LayerReport.BaseCount
// (measured directly), never from Value - Contribution, which undercounts when
// the base and the change moved overlapping files.
func TestDriftOriginSentence_UsesMeasuredBaseCount(t *testing.T) {
	t.Parallel()
	intp := func(n int) *int { return &n }

	t.Run("inherited with overlap prints the measured base", func(t *testing.T) {
		t.Parallel()
		got := driftOriginSentence(graph.LayerReport{
			Value: 6, Contribution: intp(2), BaseCount: intp(6),
			Threshold: 5, DriftOrigin: graph.DriftInherited,
		})
		if !strings.Contains(got, "base alone carries 6 >= threshold 5") {
			t.Errorf("sentence must state the measured base 6, got: %q", got)
		}
		if strings.Contains(got, "carries 4") {
			t.Errorf("sentence must not print the subtracted figure 4, got: %q", got)
		}
	})

	t.Run("crossed-by-change stays truthful", func(t *testing.T) {
		t.Parallel()
		got := driftOriginSentence(graph.LayerReport{
			Value: 6, Contribution: intp(3), BaseCount: intp(4),
			Threshold: 5, DriftOrigin: graph.DriftCrossedByChange,
		})
		if !strings.Contains(got, "base carried 4 < threshold 5") {
			t.Errorf("sentence must state the measured base 4 < 5, got: %q", got)
		}
	})

	t.Run("nil BaseCount prints no base number", func(t *testing.T) {
		t.Parallel()
		for _, origin := range []string{graph.DriftInherited, graph.DriftCrossedByChange} {
			got := driftOriginSentence(graph.LayerReport{
				Value: 6, Contribution: intp(2), BaseCount: nil,
				Threshold: 5, DriftOrigin: origin,
			})
			if strings.Contains(got, "carries 4") || strings.Contains(got, "carried 4") || strings.Contains(got, "carries 6") {
				t.Errorf("%s: no base figure may be printed without a measurement, got: %q", origin, got)
			}
			if !strings.Contains(got, "this change added 2") {
				t.Errorf("%s: the known contribution must still be stated, got: %q", origin, got)
			}
		}
	})
}
