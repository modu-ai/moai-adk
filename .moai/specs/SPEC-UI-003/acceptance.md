# SPEC-UI-003: Acceptance Tests

## Scenario 1: Interactive Wizard Renders with huh

```gherkin
Given the user runs "moai init" in an interactive terminal
When the wizard displays the language selection question
Then the question is rendered inside a rounded border
And the MoAI orange (#DA7756) color is applied to the title
And arrow key navigation cycles through options
And Enter confirms the selection
```

## Scenario 2: Interface Compatibility

```gherkin
Given existing code calls ui.NewSelector(theme, hm).Select("Choose", items)
When the selector is invoked
Then it returns (selectedValue string, error) matching the original interface
And no caller code changes are required
```

## Scenario 3: Headless Mode Fallback

```gherkin
Given the environment variable MOAI_NO_COLOR=1 is set
And stdin is not a terminal (CI/CD pipeline)
When any UI component is invoked
Then huh accessibility mode is activated
And sequential text prompts replace TUI rendering
And default values are returned when available
```

## Scenario 4: Dark Terminal Theme

```gherkin
Given the terminal has a dark background
When a huh form is rendered
Then light foreground colors from the dark palette are used
And AdaptiveColor automatically selects appropriate values
```

## Scenario 5: Animated Progress

```gherkin
Given a template deployment is in progress
When Progress.Spinner("Deploying...") is called
Then an animated spinner character rotates in the terminal
And the title text is displayed next to the spinner
And Spinner.Stop() halts the animation
```

## Scenario 6: Glamour Markdown

```gherkin
Given a help command displays documentation
When RenderMarkdown(content) is called
Then the markdown is rendered with syntax highlighting
And headings are bold and colored
And code blocks have visual distinction
And output wraps to terminal width
```

## Scenario 7: Build and Test

```gherkin
Given all TUI modernization changes are complete
When "go test -race ./internal/ui/... ./internal/cli/wizard/..." is run
Then all tests pass
And coverage is 85% or higher
And "make build" completes without errors
```

## Edge Cases

- Terminal width < 40 characters: graceful degradation
- Unicode emoji in options: proper width calculation
- Very long option labels: text truncation
- 0 items passed to selector: ErrNoItems returned
- Ctrl+C during any form: ErrCancelled returned
- Wizard with all conditional questions hidden: skip to completion
