# SPEC-UI-003: Implementation Plan

## Task Decomposition

### Task 1: Add Dependencies (backend-dev)
- Add `github.com/charmbracelet/huh` v2.x to go.mod
- Add `github.com/charmbracelet/glamour` to go.mod
- Run `go mod tidy`
- Verify compilation

### Task 2: Create MoAI huh Theme (backend-dev)
- Create MoAI-branded huh.Theme in internal/ui/theme.go
- Map existing color palette to huh theme structure
- Support AdaptiveColor for dark/light terminals
- Support NoColor/accessibility mode

### Task 3: Replace Selector with huh.Select (backend-dev)
- Rewrite selectorImpl to use huh.NewSelect internally
- Maintain Select(label, items) interface
- Support fuzzy filtering via huh's built-in filter
- Handle headless mode via huh accessibility

### Task 4: Replace Checkbox with huh.MultiSelect (backend-dev)
- Rewrite checkboxImpl to use huh.NewMultiSelect internally
- Maintain MultiSelect(label, items) interface
- Handle headless mode via huh accessibility

### Task 5: Replace Prompt with huh.Input/Confirm (backend-dev)
- Rewrite promptImpl.Input to use huh.NewInput internally
- Rewrite promptImpl.Confirm to use huh.NewConfirm internally
- Maintain WithPlaceholder, WithValidation, WithDefault options
- Handle headless mode via huh accessibility

### Task 6: Upgrade Progress/Spinner (backend-dev)
- Replace ASCII progress bar with bubbles/progress model
- Replace text-only spinner with bubbles/spinner model
- Support both interactive and headless modes
- Animated rendering in interactive mode

### Task 7: Rewrite Wizard with huh.Form + Groups (backend-dev)
- Convert wizard questions to huh fields
- Group questions into huh.Group pages
- Implement conditional visibility using huh's dynamic forms
- Apply MoAI theme
- Maintain WizardResult output structure

### Task 8: Add Glamour Markdown Rendering (backend-dev)
- Create internal/ui/glamour.go with RenderMarkdown function
- Auto-detect light/dark terminal for styling
- Support word wrapping based on terminal width
- Expose via a new MarkdownRenderer interface

### Task 9: Update All Tests (tester)
- Rewrite internal/ui/*_test.go for huh-based components
- Rewrite internal/cli/wizard/*_test.go for huh Form
- Add new tests for glamour.go
- Ensure 85%+ coverage
- Run with -race flag

### Task 10: Quality Validation (quality)
- Run `go test -race ./...`
- Run `go vet ./...`
- Run `golangci-lint run`
- Verify `make build` succeeds
- Check interface compatibility

## Dependencies

```
Task 1 (deps) -> Task 2 (theme)
Task 2 (theme) -> Tasks 3, 4, 5, 6, 7 (parallel component rewrites)
Tasks 3, 4, 5, 6, 7 -> Task 8 (glamour, can be parallel with 3-7)
Tasks 3, 4, 5, 6, 7, 8 -> Task 9 (tests)
Task 9 (tests) -> Task 10 (quality)
```

## Risk Analysis

| Risk | Mitigation |
|------|-----------|
| huh v2 API breaking changes | Pin specific version, use v2 stable API |
| Wizard conditional logic complexity | Map existing Condition functions to huh TitleFunc/OptionsFunc |
| Test mock incompatibility | Create huh-compatible test utilities |
| Headless mode regression | Test huh accessibility mode early |
| go.mod version conflicts | Run go mod tidy and verify compilation first |
