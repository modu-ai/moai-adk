package hook

import (
	"os"
	"path"
	"strings"
)

// Structural guard for file-removal commands, replacing the six `rm -rf <target>`
// regexes that used to live in DefaultSecurityPolicy.
//
// Those regexes were wrong in both directions at once (issue #1658). They pinned
// the flag cluster in literal order, so `rm -fr /` walked past a guard that
// stopped `rm -rf /`; and they ended at a bare slash, so every absolute path was
// in range and routine scratch cleanup was refused. Widening the regex fixes one
// direction and worsens the other, which is why the check is structural instead:
// the command is tokenised the way a shell would tokenise it, and the decision is
// made on the resolved TARGET rather than on the surrounding text.
//
// Flags are therefore not part of the decision at all. Any removal aimed at a
// protected target is refused whatever flags carry it, which is what removes the
// ordering bypass — there is no longer an ordering to reverse.

// shellSeparators are the unquoted runes and operators that end one command and
// begin another. Command substitution boundaries are included so the inner
// command is examined as a command rather than as text.
var shellSeparators = []string{"&&", "||", "$(", ";", "|", "&", "\n", "(", ")", "`"}

// splitCommandSegments splits a command line into individual command segments on
// unquoted separators. Quoted spans are carried through intact so a separator
// appearing inside an argument does not split it.
func splitCommandSegments(command string) []string {
	var (
		segments []string
		current  strings.Builder
		quote    rune
	)
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case quote != 0:
			current.WriteRune(c)
			if c == quote {
				quote = 0
			} else if c == '\\' && quote == '"' && i+1 < len(runes) {
				i++
				current.WriteRune(runes[i])
			}
		case c == '\'' || c == '"':
			quote = c
			current.WriteRune(c)
		case c == '\\' && i+1 < len(runes):
			current.WriteRune(c)
			i++
			current.WriteRune(runes[i])
		default:
			if width := separatorWidthAt(runes, i); width > 0 {
				segments = append(segments, current.String())
				current.Reset()
				i += width - 1
				continue
			}
			current.WriteRune(c)
		}
	}
	segments = append(segments, current.String())
	return segments
}

// separatorWidthAt reports the rune width of the separator starting at index i,
// or 0 when no separator starts there.
func separatorWidthAt(runes []rune, i int) int {
	for _, sep := range shellSeparators {
		s := []rune(sep)
		if i+len(s) > len(runes) {
			continue
		}
		if string(runes[i:i+len(s)]) == sep {
			return len(s)
		}
	}
	return 0
}

// tokenizeSegment splits one command segment into words, removing the quoting
// characters themselves. Quoting is a shell concern that disappears before the
// command sees its arguments, so `rm "-rf" "/"` yields the same three words as
// `rm -rf /` — quoting a target is not a way to hide it from this check.
func tokenizeSegment(segment string) []string {
	var (
		tokens  []string
		current strings.Builder
		open    bool
		quote   rune
	)
	flush := func() {
		if open {
			tokens = append(tokens, current.String())
			current.Reset()
			open = false
		}
	}
	runes := []rune(segment)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
				continue
			}
			if c == '\\' && quote == '"' && i+1 < len(runes) {
				i++
				c = runes[i]
			}
			current.WriteRune(c)
			open = true
		case c == '\'' || c == '"':
			quote = c
			open = true
		case c == '\\' && i+1 < len(runes):
			i++
			current.WriteRune(runes[i])
			open = true
		case c == ' ' || c == '\t':
			flush()
		default:
			current.WriteRune(c)
			open = true
		}
	}
	flush()
	return tokens
}

// removalTargets returns the non-flag operands of a segment when that segment
// invokes `rm`, and reports whether it does. Leading environment assignments and
// a `sudo` prefix are stepped over so they cannot hide the invocation.
func removalTargets(segment string) ([]string, bool) {
	tokens := tokenizeSegment(segment)
	for len(tokens) > 0 {
		head := tokens[0]
		if strings.Contains(head, "=") && !strings.HasPrefix(head, "=") {
			tokens = tokens[1:]
			continue
		}
		if path.Base(head) == "sudo" || path.Base(head) == "env" {
			tokens = tokens[1:]
			continue
		}
		break
	}
	if len(tokens) == 0 || path.Base(tokens[0]) != "rm" {
		return nil, false
	}
	// Every operand is a candidate target, flags included. Flags are never
	// protected paths, so separating them buys no decision and only reintroduces
	// the flag-shaped parsing this guard exists to get away from.
	return tokens[1:], true
}

// homeAliases are the textual forms that name the user's home directory before
// the shell expands them. They are checked before expansion because the guard
// sees the command as written, not as the shell will resolve it.
var homeAliases = map[string]bool{
	"~": true, "$HOME": true, "${HOME}": true, "~/": true,
}

// bareProtectedTargets are operands that are dangerous by themselves, independent
// of where the command happens to run.
var bareProtectedTargets = map[string]bool{
	"/": true, "*": true, ".*": true, "/*": true, ".": true, "..": true,
}

// protectedBasenames are directory names whose removal destroys work that no
// build step can reproduce.
var protectedBasenames = map[string]bool{
	".git": true, "node_modules": true,
}

// isProtectedRemovalTarget reports whether removing target would destroy the
// filesystem root, the user's home directory, a top-level system directory, or a
// directory holding irreproducible work.
//
// The old pattern asked whether the target STARTED with a slash, which made every
// absolute path protected — including the scratch directories that ordinary work
// creates and cleans up. This asks how DEEP the path is instead: `/usr` is a
// top-level directory and is protected, `/tmp/build-123` is two levels down and
// is not.
func isProtectedRemovalTarget(target, home string) bool {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return false
	}
	if homeAliases[trimmed] || bareProtectedTargets[trimmed] {
		return true
	}

	expanded := trimmed
	if home != "" {
		switch {
		case strings.HasPrefix(expanded, "~/"):
			expanded = path.Join(home, expanded[2:])
		case strings.HasPrefix(expanded, "$HOME/"):
			expanded = path.Join(home, expanded[len("$HOME/"):])
		case strings.HasPrefix(expanded, "${HOME}/"):
			expanded = path.Join(home, expanded[len("${HOME}/"):])
		}
	}

	// Collapse trailing slashes so `/usr/` and `/usr` decide the same way, while
	// keeping the root itself intact.
	for len(expanded) > 1 && strings.HasSuffix(expanded, "/") {
		expanded = expanded[:len(expanded)-1]
	}
	if expanded == "/" {
		return true
	}
	if home != "" && expanded == strings.TrimSuffix(home, "/") {
		return true
	}
	if protectedBasenames[path.Base(expanded)] {
		return true
	}
	// An absolute path with a single segment is a top-level directory.
	if strings.HasPrefix(expanded, "/") && strings.Count(expanded, "/") == 1 {
		return true
	}
	return false
}

// dangerousRemovalTarget reports the first protected target that the command aims
// a removal at, or "" when it aims none.
func dangerousRemovalTarget(command string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	for _, segment := range splitCommandSegments(command) {
		targets, isRemoval := removalTargets(segment)
		if !isRemoval {
			continue
		}
		for _, target := range targets {
			if isProtectedRemovalTarget(target, home) {
				return target
			}
		}
	}
	return ""
}
