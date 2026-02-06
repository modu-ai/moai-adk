package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/tui"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// RunInitWizardTUI runs the TUI wizard for project initialization.
// This is a PoC implementation that demonstrates bubbletea TUI for the init command.
func RunInitWizardTUI(cmd *cobra.Command, rootFlag string) error {
	// Check if we're in a TTY
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("TUI requires a terminal (TTY)")
	}

	// Run the TUI wizard
	result, err := tui.RunWizardTUI()
	if err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.") //nolint:errcheck
			return nil
		}
		return fmt.Errorf("TUI wizard failed: %w", err)
	}

	if result == nil {
		_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.") //nolint:errcheck
		return nil
	}

	// Display success message
	PrintBanner(version.GetVersion())
	fmt.Printf("\n✓ Project '%s' initialized successfully!\n", result.ProjectName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. cd", result.ProjectName)
	fmt.Println("  2. Open Claude Code in the project directory")
	fmt.Println("  3. Start coding with MoAI-ADK!")

	return nil
}

// RunInitWithTUI runs the full initialization with TUI wizard and progress display.
// This integrates the TUI wizard with the actual initialization logic.
func RunInitWithTUI(cmd *cobra.Command, rootFlag string, opts project.InitOptions) (*project.InitResult, error) {
	// Check if we're in a TTY
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return nil, fmt.Errorf("TUI requires a terminal (TTY)")
	}

	// Step 1: Run TUI wizard to collect project info
	result, err := tui.RunWizardTUI()
	if err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return nil, fmt.Errorf("wizard cancelled")
		}
		return nil, fmt.Errorf("TUI wizard failed: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("wizard cancelled")
	}

	// Apply wizard results to opts
	if opts.ProjectName == "" {
		opts.ProjectName = result.ProjectName
	}

	// Step 2: Build dependencies
	registry := foundation.DefaultRegistry
	detector := project.NewDetector(registry, nil)
	methDetector := project.NewMethodologyDetector(nil)
	validator := project.NewValidator(nil)
	mgr := manifest.NewManager()

	// Wire embedded template deployer
	embeddedFS, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("load embedded templates: %w", err)
	}

	renderer := template.NewRenderer(embeddedFS)
	deployer := template.NewDeployerWithRenderer(embeddedFS, renderer)

	initializer := project.NewInitializer(deployer, mgr, nil)
	executor := project.NewPhaseExecutor(detector, methDetector, validator, initializer, nil)

	// Step 3: Define initialization steps
	stepNames := []string{
		"Detection",
		"Methodology Detection",
		"Validation",
		"Initialization",
	}

	// Step 4: Create initialization function
	initFn := func(ctx context.Context, reporter project.ProgressReporter) (*project.InitResult, error) {
		executor.SetReporter(reporter)
		return executor.Execute(ctx, opts)
	}

	// Step 5: Run progress TUI
	finalResult, err := tui.RunProgressTUI("MoAI Project Initialization", stepNames, initFn)
	if err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	return finalResult, nil
}
