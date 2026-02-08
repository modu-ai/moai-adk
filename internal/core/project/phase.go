package project

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// PhaseExecutor orchestrates the full project initialization workflow
// by running detection, validation, and initialization phases in sequence.
type PhaseExecutor struct {
	detector            Detector
	methodologyDetector MethodologyDetector
	validator           ProjectValidator
	initializer         Initializer
	reporter            ProgressReporter // Optional progress reporter for UI
	logger              *slog.Logger
}

// NewPhaseExecutor creates a PhaseExecutor with all required dependencies.
func NewPhaseExecutor(
	detector Detector,
	methodologyDetector MethodologyDetector,
	validator ProjectValidator,
	initializer Initializer,
	logger *slog.Logger,
) *PhaseExecutor {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &PhaseExecutor{
		detector:            detector,
		methodologyDetector: methodologyDetector,
		validator:           validator,
		initializer:         initializer,
		reporter:            &NoOpReporter{}, // Default: no progress reporting
		logger:              logger,
	}
}

// SetReporter sets the progress reporter for UI updates.
func (pe *PhaseExecutor) SetReporter(reporter ProgressReporter) {
	pe.reporter = reporter
}

// Execute runs the full initialization workflow:
//  1. PhaseDetect: Detect languages, frameworks, and project type.
//  2. PhaseMethodology: Auto-detect recommended development methodology.
//  3. PhaseValidate: Validate project structure (check existing .moai/).
//  4. PhaseInit: Create directories, configs, templates.
//  5. PhaseComplete: Return results.
func (pe *PhaseExecutor) Execute(ctx context.Context, opts InitOptions) (*InitResult, error) {
	opts.ProjectRoot = filepath.Clean(opts.ProjectRoot)

	pe.logger.Info("starting project initialization", "root", opts.ProjectRoot)

	// Phase 1: Detect languages, frameworks, project type
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pe.reporter.StepStart("Detection", "Analyzing project structure")
	languages, frameworks, projectType, err := pe.phaseDetect(opts.ProjectRoot)
	if err != nil {
		pe.logger.Warn("detection phase had issues", "error", err)
		// Detection failures are non-fatal; proceed with defaults
	}
	pe.reporter.StepComplete("Detected project structure")

	// Apply detected values as defaults when not explicitly set
	opts = applyDetectedDefaults(opts, languages, frameworks, projectType)

	// Phase 2: Detect methodology (if not explicitly set)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opts.DevelopmentMode == "" {
		pe.reporter.StepStart("Methodology Detection", "Determining development mode")
		if rec, methErr := pe.phaseMethodology(opts.ProjectRoot, languages); methErr == nil && rec != nil {
			opts.DevelopmentMode = rec.Recommended
			pe.logger.Info("methodology auto-detected",
				"recommended", rec.Recommended,
				"confidence", rec.Confidence,
				"project_type", rec.ProjectType,
			)
			pe.reporter.StepUpdate(fmt.Sprintf("Recommended: %s", rec.Recommended))
		} else {
			opts.DevelopmentMode = "ddd" // fallback default
			pe.logger.Debug("methodology detection failed, using default", "error", methErr)
			pe.reporter.StepUpdate("Using default: DDD")
		}
		pe.reporter.StepComplete("Methodology determined")
	} else {
		// Validate explicitly provided development mode
		if !models.DevelopmentMode(opts.DevelopmentMode).IsValid() {
			return nil, fmt.Errorf("%w: %s", ErrInvalidDevelopmentMode, opts.DevelopmentMode)
		}
		pe.logger.Info("using explicitly set development mode", "mode", opts.DevelopmentMode)
		pe.reporter.StepStart("Validation", fmt.Sprintf("Using mode: %s", opts.DevelopmentMode))
		pe.reporter.StepComplete("Mode validated")
	}

	// Phase 3: Validate project structure
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pe.reporter.StepStart("Validation", "Validating project structure")
	if err := pe.phaseValidate(opts); err != nil {
		pe.reporter.StepError(err)
		return nil, err
	}
	pe.reporter.StepComplete("Validation passed")

	// Phase 4: Initialize
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pe.reporter.StepStart("Initialization", "Creating project structure")
	result, err := pe.initializer.Init(ctx, opts)
	if err != nil {
		pe.reporter.StepError(err)
		return nil, fmt.Errorf("initialization: %w", err)
	}
	pe.reporter.StepComplete(fmt.Sprintf("Created %d files", len(result.CreatedFiles)))

	// Phase 5: Complete
	pe.logger.Info("project initialization complete",
		"files", len(result.CreatedFiles),
		"dirs", len(result.CreatedDirs),
		"mode", result.DevelopmentMode,
	)

	return result, nil
}

// phaseDetect runs language, framework, and project type detection.
func (pe *PhaseExecutor) phaseDetect(root string) ([]Language, []Framework, models.ProjectType, error) {
	pe.logger.Debug("phase: detect")

	languages, err := pe.detector.DetectLanguages(root)
	if err != nil {
		return nil, nil, "", fmt.Errorf("detect languages: %w", err)
	}

	frameworks, err := pe.detector.DetectFrameworks(root)
	if err != nil {
		return languages, nil, "", fmt.Errorf("detect frameworks: %w", err)
	}

	projectType, err := pe.detector.DetectProjectType(root)
	if err != nil {
		return languages, frameworks, "", fmt.Errorf("detect project type: %w", err)
	}

	return languages, frameworks, projectType, nil
}

// phaseMethodology runs methodology auto-detection.
func (pe *PhaseExecutor) phaseMethodology(root string, languages []Language) (*MethodologyRecommendation, error) {
	pe.logger.Debug("phase: methodology detection")
	return pe.methodologyDetector.DetectMethodology(root, languages)
}

// phaseValidate validates the project state and handles --force backup.
func (pe *PhaseExecutor) phaseValidate(opts InitOptions) error {
	pe.logger.Debug("phase: validate")

	result, err := pe.validator.Validate(opts.ProjectRoot)
	if err != nil {
		return fmt.Errorf("validate project: %w", err)
	}

	// If project already exists
	if !result.Valid {
		if opts.Force {
			// Backup existing project
			backupPath, backupErr := BackupExistingProject(opts.ProjectRoot)
			if backupErr != nil {
				return fmt.Errorf("backup for force reinit: %w", backupErr)
			}
			pe.logger.Info("backed up existing project", "path", backupPath)
		} else {
			return fmt.Errorf("%w", ErrProjectExists)
		}
	}

	return nil
}

// applyDetectedDefaults fills in missing InitOptions fields from detection results.
func applyDetectedDefaults(opts InitOptions, languages []Language, frameworks []Framework, projectType models.ProjectType) InitOptions {
	// ProjectName defaults to directory basename
	if opts.ProjectName == "" {
		opts.ProjectName = filepath.Base(opts.ProjectRoot)
	}

	// Language defaults to primary detected language
	if opts.Language == "" {
		if len(languages) > 0 {
			opts.Language = languages[0].Name
		} else {
			opts.Language = "Go" // fallback
		}
	}

	// Framework defaults to first detected or "none"
	if opts.Framework == "" {
		if len(frameworks) > 0 {
			opts.Framework = frameworks[0].Name
		} else {
			opts.Framework = "none"
		}
	}

	// Features defaults to empty
	if opts.Features == nil {
		opts.Features = []string{}
	}

	// UserName defaults to OS user
	if opts.UserName == "" {
		opts.UserName = osUserName()
	}

	// ConvLang defaults to "en"
	if opts.ConvLang == "" {
		opts.ConvLang = "en"
	}

	// GitMode defaults to "manual"
	if opts.GitMode == "" {
		opts.GitMode = "manual"
	}

	// Output language defaults to "en"
	if opts.GitCommitLang == "" {
		opts.GitCommitLang = "en"
	}
	if opts.CodeCommentLang == "" {
		opts.CodeCommentLang = "en"
	}
	if opts.DocLang == "" {
		opts.DocLang = "en"
	}

	return opts
}

// osUserName returns the current OS username from environment variables.
func osUserName() string {
	if name := os.Getenv("USER"); name != "" {
		return name
	}
	if name := os.Getenv("USERNAME"); name != "" {
		return name
	}
	return "user"
}
