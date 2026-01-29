package cli

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/go/internal/initializer"
	"github.com/modu-ai/moai-adk/go/internal/output"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize MoAI project",
		Long: `Initialize a new MoAI project with templates, configuration,
and project structure. This command sets up the necessary files and directories
for AI-powered development with Claude Code.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization in non-empty directory")

	return cmd
}

// runInit executes the init command logic
func runInit(force bool) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Display header
	fmt.Println(output.HeaderStyle.Render("MoAI-ADK Project Initialization"))
	fmt.Println()

	// Detect existing project
	detector := initializer.NewProjectDetector()
	shouldProceed, err := detector.ShouldInit(cwd, force)
	if err != nil {
		return err
	}

	if !shouldProceed && !force {
		fmt.Println(output.ErrorStyle.Render("Initialization cancelled. Use --force to override."))
		return fmt.Errorf("directory not empty or existing project detected")
	}

	// Create prompter
	prompter := initializer.NewPrompter()

	// Prompt for language
	language, err := prompter.PromptLanguage()
	if err != nil {
		return fmt.Errorf("error prompting for language: %w", err)
	}

	// Prompt for user name
	userName, err := prompter.PromptUserName()
	if err != nil {
		return fmt.Errorf("error prompting for user name: %w", err)
	}

	fmt.Println()
	fmt.Println(output.InfoStyle.Render("Creating project structure..."))
	fmt.Println()

	// Create extractor and extract templates
	extractor := initializer.NewExtractor()
	if err := extractor.ExtractTemplates(cwd); err != nil {
		return fmt.Errorf("error extracting templates: %w", err)
	}

	// Generate settings.json with direct binary path
	generator, err := initializer.NewSettingsGenerator()
	if err != nil {
		return fmt.Errorf("error creating settings generator: %w", err)
	}

	if err := generator.WriteToFile(cwd); err != nil {
		return fmt.Errorf("error writing settings.json: %w", err)
	}

	// Write configuration files
	configWriter := initializer.NewConfigWriter(cwd)
	if err := configWriter.WriteLanguageConfig(language); err != nil {
		return fmt.Errorf("error writing language config: %w", err)
	}

	if err := configWriter.WriteUserConfig(userName); err != nil {
		return fmt.Errorf("error writing user config: %w", err)
	}

	// Display success message
	fmt.Println()
	fmt.Println(output.HeaderStyle.Render("✓ Project initialized successfully!"))
	fmt.Println()
	fmt.Println(output.InfoStyle.Render("Binary path:"))
	fmt.Println(output.MutedStyle.Render("  " + generator.GetBinaryPath()))
	fmt.Println()
	fmt.Println(output.InfoStyle.Render("Configuration:"))
	fmt.Println(output.MutedStyle.Render("  Language: " + string(language)))
	fmt.Println(output.MutedStyle.Render("  User: " + userName))
	fmt.Println()
	fmt.Println(output.MutedStyle.Render("Next steps:"))
	fmt.Println(output.MutedStyle.Render("  1. Review .claude/settings.json"))
	fmt.Println(output.MutedStyle.Render("  2. Customize .moai/config/sections/"))
	fmt.Println(output.MutedStyle.Render("  3. Start development with Claude Code"))
	fmt.Println()

	return nil
}
