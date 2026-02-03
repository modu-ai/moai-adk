package statusline

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestUpdateChecker_CheckUpdate(t *testing.T) {
	tests := []struct {
		name            string
		currentVersion  string
		latest          string
		fetchErr        error
		wantCurrent     string
		wantLatest      string
		wantUpdateAvail bool
		wantAvail       bool
	}{
		{
			name:            "update available",
			currentVersion:  "1.2.0",
			latest:          "1.3.0",
			wantCurrent:     "1.2.0",
			wantLatest:      "1.3.0",
			wantUpdateAvail: true,
			wantAvail:       true,
		},
		{
			name:            "no update - same version",
			currentVersion:  "1.3.0",
			latest:          "1.3.0",
			wantCurrent:     "1.3.0",
			wantLatest:      "1.3.0",
			wantUpdateAvail: false,
			wantAvail:       true,
		},
		{
			name:           "fetch error - no cache",
			currentVersion: "1.2.0",
			fetchErr:       errors.New("network error"),
			wantCurrent:    "1.2.0",
			wantAvail:      true,
		},
		{
			name:           "nil fetch function",
			currentVersion: "1.2.0",
			wantCurrent:    "1.2.0",
			wantAvail:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fetchFn LatestVersionFunc
			if tt.latest != "" || tt.fetchErr != nil {
				fetchFn = func(_ context.Context) (string, error) {
					return tt.latest, tt.fetchErr
				}
			}

			checker := NewUpdateChecker(tt.currentVersion, time.Hour, fetchFn)
			got, err := checker.CheckUpdate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Current != tt.wantCurrent {
				t.Errorf("Current = %q, want %q", got.Current, tt.wantCurrent)
			}
			if got.Latest != tt.wantLatest {
				t.Errorf("Latest = %q, want %q", got.Latest, tt.wantLatest)
			}
			if got.UpdateAvailable != tt.wantUpdateAvail {
				t.Errorf("UpdateAvailable = %v, want %v", got.UpdateAvailable, tt.wantUpdateAvail)
			}
			if got.Available != tt.wantAvail {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvail)
			}
		})
	}
}

func TestUpdateChecker_CacheHit(t *testing.T) {
	callCount := 0
	fetchFn := func(_ context.Context) (string, error) {
		callCount++
		return "2.0.0", nil
	}

	checker := NewUpdateChecker("1.0.0", time.Hour, fetchFn)

	// First call: fetches from function
	got1, err := checker.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if got1.Latest != "2.0.0" {
		t.Errorf("first call Latest = %q, want %q", got1.Latest, "2.0.0")
	}
	if callCount != 1 {
		t.Errorf("first call count = %d, want 1", callCount)
	}

	// Second call: should use cache
	got2, err := checker.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if got2.Latest != "2.0.0" {
		t.Errorf("second call Latest = %q, want %q", got2.Latest, "2.0.0")
	}
	if callCount != 1 {
		t.Errorf("after cache hit, call count = %d, want 1", callCount)
	}
}

func TestUpdateChecker_CacheExpiry(t *testing.T) {
	callCount := 0
	fetchFn := func(_ context.Context) (string, error) {
		callCount++
		return "2.0.0", nil
	}

	// Use very short TTL for testing
	checker := NewUpdateChecker("1.0.0", 1*time.Millisecond, fetchFn)

	// First call
	_, err := checker.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("first call count = %d, want 1", callCount)
	}

	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)

	// Second call: cache expired, should fetch again
	_, err = checker.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("after cache expiry, call count = %d, want 2", callCount)
	}
}

func TestUpdateChecker_FetchErrorWithCache(t *testing.T) {
	callCount := 0
	fetchFn := func(_ context.Context) (string, error) {
		callCount++
		if callCount == 1 {
			return "2.0.0", nil
		}
		return "", errors.New("network error")
	}

	// Use very short TTL
	checker := NewUpdateChecker("1.0.0", 1*time.Millisecond, fetchFn)

	// First call: success
	got1, err := checker.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if got1.Latest != "2.0.0" {
		t.Errorf("first call Latest = %q, want %q", got1.Latest, "2.0.0")
	}

	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)

	// Second call: fetch fails, should return cached data
	got2, err := checker.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if got2.Latest != "2.0.0" {
		t.Errorf("fallback Latest = %q, want %q (cached)", got2.Latest, "2.0.0")
	}
}
