package cli

import (
	"math"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestCodexAuditWaitTimeoutForLoadBelowThreshold(t *testing.T) {
	if got := codexAuditWaitTimeoutForLoad(20*time.Minute, config.DefaultCodexAuditWaitLoadThreshold-0.1); got != 20*time.Minute {
		t.Errorf("load just below the threshold: got %v, want the base bound unchanged", got)
	}
}

func TestCodexAuditWaitTimeoutForLoadAtAndAboveThreshold(t *testing.T) {
	if got := codexAuditWaitTimeoutForLoad(20*time.Minute, config.DefaultCodexAuditWaitLoadThreshold); got != 40*time.Minute {
		t.Errorf("load at the threshold: got %v, want 40m (2x)", got)
	}
	if got := codexAuditWaitTimeoutForLoad(20*time.Minute, 357); got != 40*time.Minute {
		t.Errorf("load 357 (the t1781 incident load): got %v, want 40m (2x)", got)
	}
}

func TestCodexAuditWaitTimeoutForLoadZero(t *testing.T) {
	if got := codexAuditWaitTimeoutForLoad(time.Minute, 0); got != time.Minute {
		t.Errorf("load 0: got %v, want the base bound", got)
	}
}

// TestLoadAverage1NonNegative pins the reader seam where a platform provides
// one: a finite, non-negative load, or an honest error the caller fails open
// on. Skips where no reader exists (the build-tag fallback).
func TestLoadAverage1NonNegative(t *testing.T) {
	load, err := loadAverage1()
	if err != nil {
		t.Skipf("no load reader on this platform: %v", err)
	}
	if math.IsNaN(load) || math.IsInf(load, 0) || load < 0 {
		t.Errorf("loadAverage1 = %v, want a finite non-negative load", load)
	}
}
