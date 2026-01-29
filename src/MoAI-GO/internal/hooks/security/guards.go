package security

import (
	"path/filepath"
	"regexp"
	"strings"
)

// SecurityGuard validates file paths and bash commands for security
type SecurityGuard struct {
	blockedPaths           []string
	blockedCommandPatterns []*regexp.Regexp
	warnPaths              []string
}

// NewSecurityGuard creates a new security guard with default rules
func NewSecurityGuard() *SecurityGuard {
	return &SecurityGuard{
		blockedPaths: []string{
			".env",
			".env.*",
			"credentials.json",
			"credentials.*.json",
			"secrets/",
			"secrets/*",
			".ssh/",
			".ssh/*",
			".gnupg/",
			".gnupg/*",
			".pem",
			".key",
			".crt",
			".token",
			".aws/",
			".gcloud/",
			".azure/",
			".kube/",
		},
		blockedCommandPatterns: compilePatterns([]string{
			// Database deletion commands - Supabase
			`supabase\s+db\s+reset`,
			`supabase\s+projects?\s+delete`,
			`supabase\s+functions?\s+delete`,
			// Database deletion commands - Neon
			`neon\s+database\s+delete`,
			`neon\s+projects?\s+delete`,
			`neon\s+branch\s+delete`,
			// Database deletion commands - PlanetScale
			`pscale\s+database\s+delete`,
			`pscale\s+branch\s+delete`,
			// Database deletion commands - Railway
			`railway\s+delete`,
			// SQL dangerous commands
			`DROP\s+DATABASE`,
			`DROP\s+SCHEMA`,
			`TRUNCATE\s+TABLE`,
			// Unix dangerous file operations
			`rm\s+-rf\s+/`,
			`rm\s+-rf\s+~`,
			`rm\s+-rf\s+\*`,
			`rm\s+-rf\s+\.\*`,
			`rm\s+-rf\s+\.git\b`,
			`rm\s+-rf\s+node_modules\s*$`,
			`chmod\s+777`,
			// Pipe to shell patterns
			`curl\s+.*\|\s*bash`,
			`wget\s+.*\|\s*sh`,
			// Windows dangerous file operations (CMD)
			`rd\s+/s\s+/q\s+[A-Za-z]:\\`,
			`rmdir\s+/s\s+/q\s+[A-Za-z]:\\`,
			`del\s+/f\s+/q\s+[A-Za-z]:\\`,
			`rd\s+/s\s+/q\s+\\\\`,
			`rd\s+/s\s+/q\s+\.git\b`,
			`del\s+/s\s+/q\s+\*\.\*`,
			`format\s+[A-Za-z]:`,
			// Windows dangerous file operations (PowerShell)
			`Remove-Item\s+.*-Recurse\s+.*-Force\s+[A-Za-z]:\\`,
			`Remove-Item\s+.*-Recurse\s+.*-Force\s+~`,
			`Remove-Item\s+.*-Recurse\s+.*-Force\s+\$env:`,
			`Remove-Item\s+.*-Recurse\s+.*-Force\s+\.git\b`,
			`Clear-Content\s+.*-Force`,
			// Git dangerous commands
			`git\s+push\s+.*--force\s+origin\s+(main|master)`,
			`git\s+branch\s+-D\s+(main|master)`,
			// Cloud infrastructure deletion
			`terraform\s+destroy`,
			`pulumi\s+destroy`,
			`aws\s+.*\s+delete-`,
			`gcloud\s+.*\s+delete\b`,
			// Azure CLI dangerous commands
			`az\s+group\s+delete`,
			`az\s+storage\s+account\s+delete`,
			`az\s+sql\s+server\s+delete`,
			// Docker dangerous commands
			`docker\s+system\s+prune\s+(-a|--all)`,
			`docker\s+image\s+prune\s+(-a|--all)`,
			`docker\s+container\s+prune`,
			`docker\s+volume\s+prune`,
			`docker\s+network\s+prune`,
			`docker\s+builder\s+prune\s+(-a|--all)`,
		}),
		warnPaths: []string{
			".claude/settings.json",
			".claude/settings.local.json",
			"/etc/",
			"/usr/local/",
			"package-lock.json",
			"yarn.lock",
			"pnpm-lock.yaml",
			"Cargo.lock",
			"poetry.lock",
			"composer.lock",
			"Pipfile.lock",
		},
	}
}

// compilePatterns compiles regex patterns for efficient matching
func compilePatterns(patterns []string) []*regexp.Regexp {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		// Use case-insensitive matching for commands
		re := regexp.MustCompile(`(?i)` + pattern)
		compiled = append(compiled, re)
	}
	return compiled
}

// Decision represents the security decision
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionBlock Decision = "block"
	DecisionWarn  Decision = "warn"
)

// ValidatePath checks if a file path is allowed
func (g *SecurityGuard) ValidatePath(filePath string) (Decision, string) {
	// Normalize path for matching
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")

	// Check blocked patterns
	for _, pattern := range g.blockedPaths {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err != nil {
			// Invalid pattern, skip it
			continue
		}
		if matched {
			return DecisionBlock, "Protected file: access denied for security reasons"
		}

		// Also check if path contains blocked patterns
		if strings.Contains(normalizedPath, pattern) {
			return DecisionBlock, "Protected path: access denied for security reasons"
		}
	}

	// Check warn patterns
	for _, pattern := range g.warnPaths {
		if strings.Contains(normalizedPath, pattern) {
			return DecisionWarn, "Critical config file: " + filepath.Base(filePath)
		}

		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err != nil {
			// Invalid pattern, skip it
			continue
		}
		if matched {
			return DecisionWarn, "Critical config file: " + filepath.Base(filePath)
		}
	}

	return DecisionAllow, ""
}

// ValidateCommand checks if a bash command is dangerous
func (g *SecurityGuard) ValidateCommand(command string) (Decision, string) {
	// Check for dangerous commands
	for _, pattern := range g.blockedCommandPatterns {
		if pattern.MatchString(command) {
			return DecisionBlock, "Dangerous command blocked: " + command
		}
	}

	return DecisionAllow, ""
}
