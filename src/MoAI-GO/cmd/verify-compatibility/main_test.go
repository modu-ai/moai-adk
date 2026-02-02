package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// ---------------------------------------------------------------------------
// typesMatch() tests
// ---------------------------------------------------------------------------

func TestTypesMatch_BothNil(t *testing.T) {
	if !typesMatch(nil, nil) {
		t.Error("expected nil, nil to match")
	}
}

func TestTypesMatch_OneNil(t *testing.T) {
	if typesMatch(nil, "hello") {
		t.Error("expected nil, string to not match")
	}
	if typesMatch("hello", nil) {
		t.Error("expected string, nil to not match")
	}
}

func TestTypesMatch_Strings(t *testing.T) {
	if !typesMatch("hello", "world") {
		t.Error("expected string, string to match")
	}
	if typesMatch("hello", 42.0) {
		t.Error("expected string, float64 to not match")
	}
}

func TestTypesMatch_Float64(t *testing.T) {
	if !typesMatch(42.0, 3.14) {
		t.Error("expected float64, float64 to match")
	}
	if typesMatch(42.0, "string") {
		t.Error("expected float64, string to not match")
	}
}

func TestTypesMatch_Bool(t *testing.T) {
	if !typesMatch(true, false) {
		t.Error("expected bool, bool to match")
	}
	if typesMatch(true, "true") {
		t.Error("expected bool, string to not match")
	}
}

func TestTypesMatch_Map(t *testing.T) {
	a := map[string]any{"key": "value"}
	b := map[string]any{"other": 42}
	if !typesMatch(a, b) {
		t.Error("expected map, map to match")
	}
	if typesMatch(a, "string") {
		t.Error("expected map, string to not match")
	}
}

func TestTypesMatch_Slice(t *testing.T) {
	a := []any{1, 2, 3}
	b := []any{"a", "b"}
	if !typesMatch(a, b) {
		t.Error("expected slice, slice to match")
	}
	if typesMatch(a, "string") {
		t.Error("expected slice, string to not match")
	}
}

func TestTypesMatch_UnknownType(t *testing.T) {
	// A type not handled by the switch (e.g., int)
	if typesMatch(42, 43) {
		t.Error("expected unknown type (int) to not match")
	}
}

// ---------------------------------------------------------------------------
// stringSlicesEqual() tests
// ---------------------------------------------------------------------------

func TestStringSlicesEqual_Equal(t *testing.T) {
	a := []string{"one", "two", "three"}
	b := []string{"three", "one", "two"}
	if !stringSlicesEqual(a, b) {
		t.Error("expected equal slices (order-independent) to match")
	}
}

func TestStringSlicesEqual_DifferentLength(t *testing.T) {
	a := []string{"one", "two"}
	b := []string{"one", "two", "three"}
	if stringSlicesEqual(a, b) {
		t.Error("expected different-length slices to not match")
	}
}

func TestStringSlicesEqual_DifferentContent(t *testing.T) {
	a := []string{"one", "two", "three"}
	b := []string{"one", "two", "four"}
	if stringSlicesEqual(a, b) {
		t.Error("expected different-content slices to not match")
	}
}

func TestStringSlicesEqual_Empty(t *testing.T) {
	if !stringSlicesEqual([]string{}, []string{}) {
		t.Error("expected empty slices to match")
	}
}

func TestStringSlicesEqual_SingleElement(t *testing.T) {
	if !stringSlicesEqual([]string{"a"}, []string{"a"}) {
		t.Error("expected single equal elements to match")
	}
	if stringSlicesEqual([]string{"a"}, []string{"b"}) {
		t.Error("expected single different elements to not match")
	}
}

// ---------------------------------------------------------------------------
// getKeys() tests
// ---------------------------------------------------------------------------

func TestGetKeys_NonEmpty(t *testing.T) {
	m := map[string]any{
		"name":  "test",
		"value": 42,
		"flag":  true,
	}

	keys := getKeys(m)
	sort.Strings(keys)

	expected := []string{"flag", "name", "value"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("key %d: expected %q, got %q", i, k, keys[i])
		}
	}
}

func TestGetKeys_Empty(t *testing.T) {
	keys := getKeys(map[string]any{})
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

// ---------------------------------------------------------------------------
// extractErrorMessage() tests
// ---------------------------------------------------------------------------

func TestExtractErrorMessage_TopLevel(t *testing.T) {
	data := map[string]any{
		"error": "something went wrong",
	}
	msg := extractErrorMessage(data)
	if msg != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", msg)
	}
}

func TestExtractErrorMessage_Nested(t *testing.T) {
	data := map[string]any{
		"data": map[string]any{
			"error": "nested error message",
		},
	}
	msg := extractErrorMessage(data)
	if msg != "nested error message" {
		t.Errorf("expected 'nested error message', got %q", msg)
	}
}

func TestExtractErrorMessage_NoError(t *testing.T) {
	data := map[string]any{
		"success": true,
	}
	msg := extractErrorMessage(data)
	if msg != "" {
		t.Errorf("expected empty string, got %q", msg)
	}
}

func TestExtractErrorMessage_NonStringError(t *testing.T) {
	data := map[string]any{
		"error": 42,
	}
	msg := extractErrorMessage(data)
	if msg != "" {
		t.Errorf("expected empty string for non-string error, got %q", msg)
	}
}

func TestExtractErrorMessage_NestedDataNotMap(t *testing.T) {
	data := map[string]any{
		"data": "not a map",
	}
	msg := extractErrorMessage(data)
	if msg != "" {
		t.Errorf("expected empty string when data is not a map, got %q", msg)
	}
}

func TestExtractErrorMessage_TopLevelTakesPrecedence(t *testing.T) {
	data := map[string]any{
		"error": "top-level error",
		"data": map[string]any{
			"error": "nested error",
		},
	}
	msg := extractErrorMessage(data)
	if msg != "top-level error" {
		t.Errorf("expected top-level error to take precedence, got %q", msg)
	}
}

// ---------------------------------------------------------------------------
// truncateOutput() tests
// ---------------------------------------------------------------------------

func TestTruncateOutput_Short(t *testing.T) {
	result := truncateOutput("hello", 500)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateOutput_ExactLength(t *testing.T) {
	s := "12345"
	result := truncateOutput(s, 5)
	if result != s {
		t.Errorf("expected exact string at boundary, got %q", result)
	}
}

func TestTruncateOutput_TooLong(t *testing.T) {
	s := "abcdefghij"
	result := truncateOutput(s, 5)
	if result != "abcde... (truncated)" {
		t.Errorf("expected truncated output, got %q", result)
	}
}

func TestTruncateOutput_Empty(t *testing.T) {
	result := truncateOutput("", 100)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestTruncateOutput_ZeroMax(t *testing.T) {
	result := truncateOutput("hello", 0)
	if result != "... (truncated)" {
		t.Errorf("expected truncated at 0, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// formatBool() tests
// ---------------------------------------------------------------------------

func TestFormatBool_True(t *testing.T) {
	result := formatBool(true)
	if result != "\u2713" { // checkmark
		t.Errorf("expected checkmark, got %q", result)
	}
}

func TestFormatBool_False(t *testing.T) {
	result := formatBool(false)
	if result != "\u2717" { // X mark
		t.Errorf("expected X mark, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// compareJSONSchemas() tests
// ---------------------------------------------------------------------------

func TestCompareJSONSchemas_Matching(t *testing.T) {
	python := `{"name": "test", "count": 42, "active": true}`
	goOut := `{"name": "other", "count": 99, "active": false}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if !match {
		t.Errorf("expected schema match, got differences: %v", result.Differences)
	}
}

func TestCompareJSONSchemas_DifferentKeys(t *testing.T) {
	python := `{"name": "test", "count": 42}`
	goOut := `{"name": "test", "total": 42}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if match {
		t.Error("expected schema mismatch for different keys")
	}
	if len(result.Differences) == 0 {
		t.Error("expected differences to be recorded")
	}
}

func TestCompareJSONSchemas_DifferentTypes(t *testing.T) {
	python := `{"value": "string"}`
	goOut := `{"value": 42}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if match {
		t.Error("expected schema mismatch for different types")
	}
	if len(result.Differences) == 0 {
		t.Error("expected type mismatch difference")
	}
}

func TestCompareJSONSchemas_InvalidPythonJSON(t *testing.T) {
	python := `{not valid json}`
	goOut := `{"key": "value"}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if match {
		t.Error("expected mismatch for invalid Python JSON")
	}
	if len(result.Differences) == 0 {
		t.Error("expected parse error difference")
	}
}

func TestCompareJSONSchemas_InvalidGoJSON(t *testing.T) {
	python := `{"key": "value"}`
	goOut := `{not valid json}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if match {
		t.Error("expected mismatch for invalid Go JSON")
	}
	if len(result.Differences) == 0 {
		t.Error("expected parse error difference")
	}
}

func TestCompareJSONSchemas_BothInvalidJSON(t *testing.T) {
	python := `{bad}`
	goOut := `{also bad}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if match {
		t.Error("expected mismatch for both invalid JSON")
	}
}

func TestCompareJSONSchemas_EmptyObjects(t *testing.T) {
	python := `{}`
	goOut := `{}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if !match {
		t.Error("expected empty objects to match")
	}
}

func TestCompareJSONSchemas_NestedObjects(t *testing.T) {
	python := `{"data": {"inner": "value"}, "count": 1}`
	goOut := `{"data": {"inner": "other"}, "count": 2}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if !match {
		t.Errorf("expected nested objects with same types to match, got differences: %v", result.Differences)
	}
}

func TestCompareJSONSchemas_ArrayFields(t *testing.T) {
	python := `{"items": [1, 2, 3]}`
	goOut := `{"items": [4, 5]}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if !match {
		t.Errorf("expected array fields to match by type, got differences: %v", result.Differences)
	}
}

func TestCompareJSONSchemas_DifferentKeyCount(t *testing.T) {
	python := `{"a": 1, "b": 2}`
	goOut := `{"a": 1}`

	result := &VerificationResult{}
	match := compareJSONSchemas(python, goOut, result)
	if match {
		t.Error("expected mismatch for different key count")
	}
}

// ---------------------------------------------------------------------------
// compareErrorMessages() tests
// ---------------------------------------------------------------------------

func TestCompareErrorMessages_BothSuccess(t *testing.T) {
	result := &VerificationResult{}
	match := compareErrorMessages(`{"success": true}`, `{"success": true}`, 0, 0, result)
	if !match {
		t.Error("expected match when both succeed")
	}
}

func TestCompareErrorMessages_BothFailed_BothHaveErrors(t *testing.T) {
	python := `{"error": "python error"}`
	goOut := `{"error": "go error"}`
	result := &VerificationResult{}

	match := compareErrorMessages(python, goOut, 1, 1, result)
	if !match {
		t.Error("expected match when both failed with error fields")
	}
}

func TestCompareErrorMessages_BothFailed_MissingPythonError(t *testing.T) {
	python := `{"success": false}`
	goOut := `{"error": "go error"}`
	result := &VerificationResult{}

	match := compareErrorMessages(python, goOut, 1, 1, result)
	if match {
		t.Error("expected mismatch when Python has no error field")
	}
}

func TestCompareErrorMessages_BothFailed_MissingGoError(t *testing.T) {
	python := `{"error": "python error"}`
	goOut := `{"success": false}`
	result := &VerificationResult{}

	match := compareErrorMessages(python, goOut, 1, 1, result)
	if match {
		t.Error("expected mismatch when Go has no error field")
	}
}

func TestCompareErrorMessages_BothFailed_InvalidPythonJSON(t *testing.T) {
	result := &VerificationResult{}
	match := compareErrorMessages("{bad", `{"error": "go err"}`, 1, 1, result)
	if !match {
		t.Error("expected match (ignore) when Python JSON parse fails")
	}
}

func TestCompareErrorMessages_BothFailed_InvalidGoJSON(t *testing.T) {
	result := &VerificationResult{}
	match := compareErrorMessages(`{"error": "py err"}`, "{bad", 1, 1, result)
	if !match {
		t.Error("expected match (ignore) when Go JSON parse fails")
	}
}

func TestCompareErrorMessages_OneFailed(t *testing.T) {
	result := &VerificationResult{}
	match := compareErrorMessages(`{"success": true}`, `{"error": "failed"}`, 0, 1, result)
	if match {
		t.Error("expected mismatch when exit codes differ")
	}
	if len(result.Differences) == 0 {
		t.Error("expected difference to be recorded for mismatched exit codes")
	}
}

func TestCompareErrorMessages_OtherOneFailed(t *testing.T) {
	result := &VerificationResult{}
	match := compareErrorMessages(`{"error": "failed"}`, `{"success": true}`, 1, 0, result)
	if match {
		t.Error("expected mismatch when Python fails but Go succeeds")
	}
}

// ---------------------------------------------------------------------------
// VerificationResult struct tests
// ---------------------------------------------------------------------------

func TestVerificationResult_AllMatch(t *testing.T) {
	result := &VerificationResult{
		Hook:           "session_start",
		SchemaMatch:    true,
		ExitCodeMatch:  true,
		ErrorMatch:     true,
		PythonExitCode: 0,
		GoExitCode:     0,
	}

	if !(result.SchemaMatch && result.ExitCodeMatch && result.ErrorMatch) {
		t.Error("expected all matches to be true")
	}
}

func TestVerificationResult_WithDifferences(t *testing.T) {
	result := &VerificationResult{
		Hook:          "test_hook",
		SchemaMatch:   false,
		ExitCodeMatch: true,
		ErrorMatch:    true,
		Differences:   []string{"Key mismatch", "Type mismatch"},
	}

	if result.SchemaMatch && result.ExitCodeMatch && result.ErrorMatch {
		t.Error("expected at least one mismatch")
	}
	if len(result.Differences) != 2 {
		t.Errorf("expected 2 differences, got %d", len(result.Differences))
	}
}

// ---------------------------------------------------------------------------
// HookOutput struct tests
// ---------------------------------------------------------------------------

func TestHookOutput_StructLayout(t *testing.T) {
	output := HookOutput{
		Success: true,
		Data: map[string]any{
			"key": "value",
		},
	}
	if !output.Success {
		t.Error("expected Success=true")
	}
	if output.Data["key"] != "value" {
		t.Error("expected Data key=value")
	}
}

func TestHookOutput_WithError(t *testing.T) {
	output := HookOutput{
		Success: false,
		Error:   "something failed",
	}
	if output.Success {
		t.Error("expected Success=false")
	}
	if output.Error != "something failed" {
		t.Errorf("expected error message, got %q", output.Error)
	}
}

// ---------------------------------------------------------------------------
// Integration: compareJSONSchemas with compareErrorMessages
// ---------------------------------------------------------------------------

func TestFullComparison_IdenticalOutputs(t *testing.T) {
	output := `{"success": true, "data": {"users": [1, 2, 3], "count": 3}}`

	result := &VerificationResult{
		Hook:           "test_hook",
		PythonOutput:   output,
		GoOutput:       output,
		PythonExitCode: 0,
		GoExitCode:     0,
	}

	result.ExitCodeMatch = (result.PythonExitCode == result.GoExitCode)
	result.SchemaMatch = compareJSONSchemas(result.PythonOutput, result.GoOutput, result)
	result.ErrorMatch = compareErrorMessages(result.PythonOutput, result.GoOutput, result.PythonExitCode, result.GoExitCode, result)

	if !result.SchemaMatch {
		t.Errorf("expected schema match for identical outputs, differences: %v", result.Differences)
	}
	if !result.ExitCodeMatch {
		t.Error("expected exit code match for identical outputs")
	}
	if !result.ErrorMatch {
		t.Error("expected error match for identical outputs")
	}
}

func TestFullComparison_DifferentSchemas(t *testing.T) {
	pythonOut := `{"name": "test", "version": "1.0"}`
	goOut := `{"name": "test", "build": 42}`

	result := &VerificationResult{
		Hook:           "schema_test",
		PythonOutput:   pythonOut,
		GoOutput:       goOut,
		PythonExitCode: 0,
		GoExitCode:     0,
	}

	result.ExitCodeMatch = (result.PythonExitCode == result.GoExitCode)
	result.SchemaMatch = compareJSONSchemas(result.PythonOutput, result.GoOutput, result)
	result.ErrorMatch = compareErrorMessages(result.PythonOutput, result.GoOutput, result.PythonExitCode, result.GoExitCode, result)

	if result.SchemaMatch {
		t.Error("expected schema mismatch for different field names")
	}
	if !result.ExitCodeMatch {
		t.Error("expected exit code match")
	}
	if len(result.Differences) == 0 {
		t.Error("expected differences to be recorded")
	}
}

// ---------------------------------------------------------------------------
// printResults() tests
// ---------------------------------------------------------------------------

func TestPrintResults_DoesNotPanic(t *testing.T) {
	result := &VerificationResult{
		Hook:           "test_hook",
		PythonOutput:   `{"success": true}`,
		GoOutput:       `{"success": true}`,
		PythonExitCode: 0,
		GoExitCode:     0,
		SchemaMatch:    true,
		ExitCodeMatch:  true,
		ErrorMatch:     true,
	}

	// Redirect stdout to avoid test output noise
	oldStdout := os.Stdout
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = devNull
	defer func() {
		os.Stdout = oldStdout
		_ = devNull.Close()
	}()

	// Should not panic
	printResults(result)
}

func TestPrintResults_WithDifferences(t *testing.T) {
	result := &VerificationResult{
		Hook:           "test_hook",
		PythonOutput:   `{"name": "test"}`,
		GoOutput:       `{"value": 42}`,
		PythonExitCode: 0,
		GoExitCode:     1,
		SchemaMatch:    false,
		ExitCodeMatch:  false,
		ErrorMatch:     false,
		Differences:    []string{"Key mismatch", "Type mismatch"},
	}

	oldStdout := os.Stdout
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = devNull
	defer func() {
		os.Stdout = oldStdout
		_ = devNull.Close()
	}()

	printResults(result)
}

func TestPrintResults_LongOutput(t *testing.T) {
	// Create output longer than 500 chars to trigger truncation in printResults
	longOutput := ""
	for i := 0; i < 100; i++ {
		longOutput += "abcdefghij"
	}

	result := &VerificationResult{
		Hook:           "long_hook",
		PythonOutput:   longOutput,
		GoOutput:       longOutput,
		PythonExitCode: 0,
		GoExitCode:     0,
		SchemaMatch:    false,
		ExitCodeMatch:  true,
		ErrorMatch:     true,
	}

	oldStdout := os.Stdout
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = devNull
	defer func() {
		os.Stdout = oldStdout
		_ = devNull.Close()
	}()

	printResults(result)
}

// ---------------------------------------------------------------------------
// findGoBinary() tests
// ---------------------------------------------------------------------------

func TestFindGoBinary_NotFound(t *testing.T) {
	// Override HOME to a temp dir where no binary exists
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	_, err := findGoBinary()
	if err == nil {
		t.Error("expected error when binary is not found")
	}
}

func TestFindGoBinary_FoundInLocalBin(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	// Create a fake binary at ~/.local/bin/moai-adk
	localBin := filepath.Join(tmpHome, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(localBin, "moai-adk")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findGoBinary()
	if err != nil {
		t.Fatalf("expected to find binary: %v", err)
	}
	if path != fakeBinary {
		t.Errorf("expected path %s, got %s", fakeBinary, path)
	}
}

func TestFindGoBinary_FoundInGOPATH(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	gopath := filepath.Join(tmpHome, "gopath")
	t.Setenv("GOPATH", gopath)

	gopathBin := filepath.Join(gopath, "bin")
	if err := os.MkdirAll(gopathBin, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(gopathBin, "moai-adk")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findGoBinary()
	if err != nil {
		t.Fatalf("expected to find binary: %v", err)
	}
	if path != fakeBinary {
		t.Errorf("expected path %s, got %s", fakeBinary, path)
	}
}

// ---------------------------------------------------------------------------
// runPythonHook() tests
// ---------------------------------------------------------------------------

func TestRunPythonHook_HookNotFound(t *testing.T) {
	// Save and restore working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, _, err = runPythonHook("nonexistent_hook")
	if err == nil {
		t.Error("expected error when Python hook file does not exist")
	}
}

// ---------------------------------------------------------------------------
// runGoHook() tests
// ---------------------------------------------------------------------------

func TestRunGoHook_BinaryNotFound(t *testing.T) {
	// Override HOME so findGoBinary fails
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	_, _, err := runGoHook("test_hook")
	if err == nil {
		t.Error("expected error when Go binary is not found")
	}
}

// ---------------------------------------------------------------------------
// verifyHook() tests
// ---------------------------------------------------------------------------

func TestVerifyHook_PythonHookNotFound(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, err = verifyHook("nonexistent_hook")
	if err == nil {
		t.Error("expected error when Python hook is not found")
	}
}

// ---------------------------------------------------------------------------
// runPythonHook() with real hook file
// ---------------------------------------------------------------------------

func TestRunPythonHook_WithRealHookFile(t *testing.T) {
	// Save and restore working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create the hook directory structure
	hookDir := filepath.Join(tmpDir, ".claude", "hooks", "moai")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple Python hook that outputs JSON
	hookFile := filepath.Join(hookDir, "test_hook.py")
	hookContent := `import json
print(json.dumps({"success": True, "data": {"message": "hello from test"}}))
`
	if err := os.WriteFile(hookFile, []byte(hookContent), 0755); err != nil {
		t.Fatal(err)
	}

	output, exitCode, err := runPythonHook("test_hook")
	if err != nil {
		// uv may not be available - skip instead of fail
		t.Skipf("uv not available for Python hook test: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d, output: %s", exitCode, output)
	}
	if output == "" {
		t.Error("expected non-empty output from Python hook")
	}
}

func TestRunPythonHook_HookExitsWithError(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	hookDir := filepath.Join(tmpDir, ".claude", "hooks", "moai")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a Python hook that exits with error
	hookFile := filepath.Join(hookDir, "failing_hook.py")
	hookContent := `import sys
sys.exit(1)
`
	if err := os.WriteFile(hookFile, []byte(hookContent), 0755); err != nil {
		t.Fatal(err)
	}

	output, exitCode, err := runPythonHook("failing_hook")
	// err should be nil (exec errors are captured in exitCode)
	_ = output
	_ = err
	if exitCode == 0 {
		t.Error("expected non-zero exit code from failing hook")
	}
}

// ---------------------------------------------------------------------------
// runGoHook() with mock binary
// ---------------------------------------------------------------------------

func TestRunGoHook_WithMockBinary(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	// Create a fake moai-adk script that outputs JSON
	localBin := filepath.Join(tmpHome, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(localBin, "moai-adk")
	scriptContent := `#!/bin/sh
echo '{"success": true, "data": {"hook": "test"}}'
`
	if err := os.WriteFile(fakeBinary, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	output, exitCode, err := runGoHook("test_hook")
	if err != nil {
		t.Fatalf("runGoHook failed: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if output == "" {
		t.Error("expected non-empty output from mock Go binary")
	}
}

func TestRunGoHook_MockBinaryExitsWithError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	localBin := filepath.Join(tmpHome, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(localBin, "moai-adk")
	scriptContent := `#!/bin/sh
echo '{"error": "mock error"}'
exit 1
`
	if err := os.WriteFile(fakeBinary, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	_, exitCode, err := runGoHook("test_hook")
	// err should be nil (exit errors are captured via ExitError)
	_ = err
	if exitCode == 0 {
		t.Error("expected non-zero exit code from failing mock binary")
	}
}

// ---------------------------------------------------------------------------
// verifyHook() with mock binaries
// ---------------------------------------------------------------------------

func TestVerifyHook_WithMockBinaries(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Setup Python hook
	hookDir := filepath.Join(tmpDir, ".claude", "hooks", "moai")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		t.Fatal(err)
	}
	hookFile := filepath.Join(hookDir, "compat_test.py")
	pyContent := `import json
print(json.dumps({"success": True, "data": {"source": "python"}}))
`
	if err := os.WriteFile(hookFile, []byte(pyContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Setup mock Go binary
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	localBin := filepath.Join(tmpHome, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(localBin, "moai-adk")
	goScript := `#!/bin/sh
echo '{"success": true, "data": {"source": "go"}}'
`
	if err := os.WriteFile(fakeBinary, []byte(goScript), 0755); err != nil {
		t.Fatal(err)
	}

	result, err := verifyHook("compat_test")
	if err != nil {
		t.Skipf("verifyHook failed (uv may not be available): %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Hook != "compat_test" {
		t.Errorf("expected hook name 'compat_test', got %q", result.Hook)
	}
	if !result.ExitCodeMatch {
		t.Error("expected exit codes to match (both 0)")
	}
}

// ---------------------------------------------------------------------------
// findGoBinary() HOME fallback test
// ---------------------------------------------------------------------------

func TestFindGoBinary_USERPROFILEFallback(t *testing.T) {
	tmpHome := t.TempDir()
	// Set HOME empty and USERPROFILE to tmpHome
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("GOPATH", filepath.Join(tmpHome, "go"))

	// No binary exists - should return error
	_, err := findGoBinary()
	if err == nil {
		t.Error("expected error when no binary exists")
	}
}
