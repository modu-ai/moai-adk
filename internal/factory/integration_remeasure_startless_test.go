package factory

// integration_remeasure_startless_test.go — card t1582 sync-audit F2
// (REQ-MWQ2-003, the truncated-capture refusal of t1576 review round 4): the
// truncation check reads a SET difference — the started packages that have no
// terminal event — not a count difference. A start-less terminal event (its
// start line cut from the head of the stream) reports a package the stream
// never started; it must not cancel an unfinished started package.

import "testing"

func TestClassifyStartlessTerminalCannotCancelTruncation(t *testing.T) {
	// pkgA starts and never reports (the capture is cut off). pkgB's package-level
	// pass arrives without its start event. The count difference (1 started minus
	// 1 finished) read this stream as complete.
	truncated := "" +
		`{"Action":"start","Package":"pkgA"}` + "\n" +
		`{"Action":"run","Package":"pkgA","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"pkgA","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"pkgB"}` + "\n"
	if _, _, err := countGoTestJSONTests(truncated); err == nil {
		t.Fatalf("a started package that never reported must refuse the stream, even with a start-less terminal event present")
	}

	// Control: once pkgA reports its own terminal event, the same stream is a
	// finished sweep and must still classify with its per-test pass count.
	complete := truncated + `{"Action":"pass","Package":"pkgA"}` + "\n"
	count, _, err := countGoTestJSONTests(complete)
	if err != nil {
		t.Fatalf("the complete stream must still classify: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1 (the per-test pass of pkgA)", count)
	}
}
