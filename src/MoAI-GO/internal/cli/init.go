package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/initializer"
	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize MoAI project",
		Long: `Initialize a new MoAI project with templates, configuration,
and project structure. This command sets up the necessary files and directories
for AI-powered development with Claude Code.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return runInit(path, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization in non-empty directory")

	return cmd
}

// runInit executes the init command logic
func runInit(path string, force bool) error {
	// Print banner with version info
	output.PrintBanner("1.12.5")

	// Print welcome message
	output.PrintWelcomeMessage()

	// Get project path
	projectPath := path
	if path == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}
		projectPath = cwd
	}

	// Detect existing project
	detector := initializer.NewProjectDetector()
	shouldProceed, err := detector.ShouldInit(projectPath, force)
	if err != nil {
		return err
	}

	if !shouldProceed && !force {
		fmt.Println(output.ErrorStyle.Render("\n✗ Initialization cancelled. Use --force to override."))
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
	fmt.Println(output.InfoStyle.Render("🚀 Starting installation..."))
	fmt.Println()

	// Create progress writer
	progressWriter, _ := output.CreateProgressCallback()
	progressWriter.Start()

	// Phase 1: Preparation and backup
	progressWriter.Update("Phase 1: Preparation and backup...")

	// Phase 2: Create directory structure
	progressWriter.Update("Phase 2: Creating directory structure...")

	// Create extractor and extract templates
	extractor := initializer.NewExtractor()
	if err := extractor.ExtractTemplates(projectPath); err != nil {
		return fmt.Errorf("error extracting templates: %w", err)
	}

	// Phase 3: Install resources
	progressWriter.Update("Phase 3: Installing resources...")

	// Generate settings.json with direct binary path
	generator, err := initializer.NewSettingsGenerator()
	if err != nil {
		return fmt.Errorf("error creating settings generator: %w", err)
	}

	if err := generator.WriteToFile(projectPath); err != nil {
		return fmt.Errorf("error writing settings.json: %w", err)
	}

	// Phase 4: Generate configurations
	progressWriter.Update("Phase 4: Generating configurations...")

	// Write configuration files
	configWriter := initializer.NewConfigWriter(projectPath)
	if err := configWriter.WriteLanguageConfig(language); err != nil {
		return fmt.Errorf("error writing language config: %w", err)
	}

	if err := configWriter.WriteUserConfig(userName); err != nil {
		return fmt.Errorf("error writing user config: %w", err)
	}

	// Phase 5: Validation and finalization
	progressWriter.Update("Phase 5: Validation and finalization...")

	// Mark progress as complete
	progressWriter.Complete()

	// Display success message with Python-style layout
	separator := output.MutedStyle.Render("────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println(output.SuccessStyle.Render("✅ Initialization Completed Successfully!"))
	fmt.Println(separator)
	fmt.Println()
	fmt.Println(output.InfoStyle.Render("📊 Summary:"))
	fmt.Printf("  %s  %s\n", output.MutedStyle.Render("📁 Location:"), projectPath)
	fmt.Printf("  %s  %s\n", output.MutedStyle.Render("🌐 Language:"), string(language))
	fmt.Printf("  %s  %s\n", output.MutedStyle.Render("👤 User:"), userName)
	fmt.Printf("  %s  %s\n", output.MutedStyle.Render("🔀 Git:"), "manual (github-flow, branch: manual)")
	fmt.Println()
	fmt.Println(separator)
	fmt.Println()
	fmt.Println(output.InfoStyle.Render("🚀 Next Steps:"))
	fmt.Printf("  %s %s\n", output.MutedStyle.Render("1."), "Review .claude/settings.json")
	fmt.Printf("  %s %s\n", output.MutedStyle.Render("2."), "Customize .moai/config/sections/")
	fmt.Printf("  %s %s\n", output.MutedStyle.Render("3."), "Start development with Claude Code")
	fmt.Println()

	return nil
}
