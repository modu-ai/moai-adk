package convention

// builtinConventions holds the built-in convention configurations.
var builtinConventions = map[string]ConventionConfig{
	"conventional-commits": {
		Name:      "conventional-commits",
		Pattern:   `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\(.+\))?!?: .+`,
		Types:     []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"},
		MaxLength: 100,
		Examples: []string{
			"feat(auth): add JWT token validation",
			"fix: resolve null pointer in user service",
			"docs(readme): update installation guide",
		},
	},
	"angular": {
		Name:      "angular",
		Pattern:   `^(build|ci|docs|feat|fix|perf|refactor|test)(\([a-z-]+\))?: .+`,
		Types:     []string{"build", "ci", "docs", "feat", "fix", "perf", "refactor", "test"},
		MaxLength: 100,
		Examples: []string{
			"feat(router): add lazy loading support",
			"fix(compiler): handle edge case in template parser",
		},
	},
	"karma": {
		Name:      "karma",
		Pattern:   `^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+\))?: .+`,
		Types:     []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "chore"},
		MaxLength: 100,
		Examples: []string{
			"feat(service): add user notification endpoint",
			"test(api): add integration tests for auth module",
		},
	},
}

// BuiltinNames returns the list of available built-in convention names.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtinConventions))
	for k := range builtinConventions {
		names = append(names, k)
	}
	return names
}

// GetBuiltin returns a built-in convention config by name.
// Returns nil if not found.
func GetBuiltin(name string) *ConventionConfig {
	cfg, ok := builtinConventions[name]
	if !ok {
		return nil
	}
	return &cfg
}
